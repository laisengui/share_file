package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"share_file/utils"
	"time"
)

var (
	uploadDir       = "./uploads"
	codeLength      = 6
	cleanupInterval = 1 * time.Hour
	defaultExpiry   = 24 * time.Hour
	port            = 8080
)

var database *FileStorage

func main() {

	//解析参数
	workDir := flag.String("work_dir", ".", "Directory to workspace")
	codeLen := flag.Int("codeLength", 6, "Length of the generated access codes")
	cleanup := flag.Duration("cleanup_interval", 24*time.Hour, "Interval for cleaning up expired files")
	expiry := flag.Duration("default_expiry", 24*time.Hour, "Default expiry time for uploads")
	portFlag := flag.Int("port", 8080, "Port to run the server on")
	logFlag := flag.Bool("log", false, "log record to workspace file")

	// Parse the flags
	flag.Parse()

	codeLength = *codeLen
	cleanupInterval = *cleanup
	defaultExpiry = *expiry
	port = *portFlag
	//初始化目录
	//存放上传文件的目录
	uploadDir = filepath.Join(*workDir, "uploads")
	err := utils.CheckAndCreateDir(uploadDir)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	//存放日志的目录
	if *logFlag {
		logDir := filepath.Join(*workDir, "logs")
		err = utils.CheckAndCreateDir(logDir)
		if err != nil {
			log.Fatal(err)
		} else {
			logFile := filepath.Join(logDir, "console.log")
			f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
			os.Stdout = f
			os.Stderr = f
		}
	}

	//初始化数据库
	// 打开数据库
	databaseDir := filepath.Join(*workDir, "database")

	database, err = NewFileStorage(databaseDir)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer database.Close()

	// Print the configuration for verification
	log.Printf("Starting server with configuration:")
	log.Printf("work directory: %s", *workDir)
	log.Printf("Code length: %d", *codeLen)
	log.Printf("Cleanup interval: %v", *cleanup)
	log.Printf("Default expiry: %v", *expiry)
	log.Printf("Server port: %d", *portFlag)

	//启动http服务
	startHttp()

}
