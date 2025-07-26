package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
)

func ToStr(c interface{}) string {
	vvv, okkk := c.(string)
	if okkk {
		return vvv
	}
	v, ok := c.(int)
	if ok {
		return strconv.Itoa(v)
	}
	vv, okk := c.(int64)
	if okk {
		return strconv.FormatInt(vv, 10)
	}

	return ""
}

// GenerateRandomCode 生成随机文件码
func GenerateRandomCode(length int) string {
	const digits = "0123456789abcdefghigklmnopqrstuvwxyz"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	for i := 0; i < length; i++ {
		b[i] = digits[int(b[i])%len(digits)]
	}

	return string(b)
}

// CheckAndCreateParentDir 检查并创建父目录, 传入的是具体文件地址
func CheckAndCreateParentDir(filePath string) error {
	parentDir := filepath.Dir(filePath)
	_, errf := os.Stat(parentDir)
	if errf == nil {
		// 父目录存在
		return nil
	} else if os.IsNotExist(errf) {
		// 父目录不存在
		err := os.MkdirAll(parentDir, 0755)
		if err != nil {
			return fmt.Errorf("创建文件夹失败: %v", err)
		}
	}
	return nil
}
func CheckAndCreateDir(filePath string) error {
	_, errf := os.Stat(filePath)
	if errf == nil {
		// 父目录存在
		return nil
	} else if os.IsNotExist(errf) {
		// 父目录不存在
		err := os.MkdirAll(filePath, 0755)
		if err != nil {
			return fmt.Errorf("创建文件夹失败: %v", err)
		}
	}
	return nil
}
func IsZip(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	return data[0] == 0x50 && data[1] == 0x4B &&
		data[2] == 0x03 && data[3] == 0x04
}

// CompressSingleFileToZip 压缩zip
func CompressSingleFileToZip(filename string, fileData []byte) ([]byte, error) {
	// 创建一个内存缓冲区来保存ZIP文件
	buf := new(bytes.Buffer)

	// 创建一个新的ZIP写入器
	zipWriter := zip.NewWriter(buf)

	// 在ZIP文件中创建一个新文件
	fileWriter, err := zipWriter.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("无法在ZIP中创建文件: %w", err)
	}

	// 将文件数据写入ZIP
	_, err = fileWriter.Write(fileData)
	if err != nil {
		return nil, fmt.Errorf("无法写入ZIP文件: %w", err)
	}

	// 关闭ZIP写入器以确保所有数据都写入缓冲区
	err = zipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("无法关闭ZIP写入器: %w", err)
	}

	// 返回ZIP文件的字节数据
	return buf.Bytes(), nil
}

// ExtractSingleFileZip 内存中解压zip包，返回文件名到内容的映射
func ExtractSingleFileZip(zipData []byte) ([]byte, error) {
	// 创建一个内存中的zip读取器
	reader := bytes.NewReader(zipData)
	zipReader, err := zip.NewReader(reader, int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("创建zip读取器失败: %w", err)
	}

	// 遍历zip文件中的所有文件
	for _, file := range zipReader.File {
		// 打开压缩文件
		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("打开文件 %s 失败: %w", file.Name, err)
		}
		defer rc.Close()

		// 读取文件内容到内存
		content, err := io.ReadAll(rc)
		if err != nil {
			return nil, fmt.Errorf("读取文件 %s 内容失败: %w", file.Name, err)
		}
		return content, nil
	}

	return nil, fmt.Errorf("没有文件")
}

func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("没有找到有效的IP地址")
}
