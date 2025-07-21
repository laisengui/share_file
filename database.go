package main

import (
	"encoding/json"
	"github.com/dgraph-io/badger"
	"log"
	"os"
	"path/filepath"
	"time"
)

type FileInfo struct {
	Uuid        string    `json:"uuid"`
	Filename    string    `json:"filename"`
	Code        string    `json:"code"`
	Size        int64     `json:"size"`
	Times       int       `json:"times"`
	UploadTime  time.Time `json:"upload_time"`
	ExpiryTime  time.Time `json:"expiry_time"`
	ContentType string    `json:"content_type"`
}

type FileStorage struct {
	db *badger.DB
}

func NewFileStorage(dbPath string) (*FileStorage, error) {
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil // 禁用Badger的日志，可选

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	fs := &FileStorage{db: db}

	// 启动定期GC
	go fs.startGC(5 * time.Minute)

	// 启动定时清理过期文件
	go fs.cleanupExpiredFiles()
	return fs, nil
}

func (fs *FileStorage) startGC(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
	again:
		err := fs.db.RunValueLogGC(0.7)
		if err == nil {
			goto again
		}
	}
}

func (fs *FileStorage) Close() error {
	return fs.db.Close()
}

// SaveFile 保存文件信息
func (fs *FileStorage) SaveFile(code string, info FileInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	return fs.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(code), data)
	})
}

// GetFile 获取文件信息
func (fs *FileStorage) GetFile(code string) (FileInfo, error) {
	var info FileInfo
	err := fs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(code))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &info)
		})
	})

	return info, err
}

// DeleteFile 删除文件信息
func (fs *FileStorage) DeleteFile(code string) error {
	return fs.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(code))
	})
}

// GetAllFiles 获取所有文件信息 (谨慎使用，数据量大时可能有问题)
func (fs *FileStorage) GetAllFiles() (map[string]FileInfo, error) {
	result := make(map[string]FileInfo)

	err := fs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			var info FileInfo

			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &info)
			})
			if err != nil {
				return err
			}

			result[string(item.Key())] = info
		}
		return nil
	})

	return result, err
}

// cleanupExpiredFiles 清除过期文件
func (fs *FileStorage) cleanupExpiredFiles() {
	for {
		time.Sleep(cleanupInterval)

		now := time.Now()
		all, err := fs.GetAllFiles()
		if err != nil {
			log.Printf("Error getting all files: %v", err)
			continue
		}
		for code, fileInfo := range all {
			if now.After(fileInfo.ExpiryTime) {
				// 删除文件
				filename := filepath.Join(uploadDir, fileInfo.Uuid)
				if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
					log.Printf("无法删除过期文件 %s: %v", filename, err)
				}
				// 从map中删除
				fs.DeleteFile(code)
			}
		}
	}
}
