package main

import (
	"embed"
	"flag"
	"log"
	"os"
	"path/filepath"
	"share_file/utils"
	"strings"
	"time"
)

var (
	uploadDir       = "./uploads"
	base            = "/"
	codeLength      = 6
	cleanupInterval = 1 * time.Hour
	defaultExpiry   = 24 * time.Hour
	port            = 8080
	enableZipFlag   = false
	loginFlag       = 0
)

var database *FileStorage
var userMap = map[string]string{}

//go:embed static/locales/*
var localesFiles embed.FS

func main() {

	//解析参数
	var initFlag bool
	flag.BoolVar(&initFlag, "init", false, "Initialize service script")
	workDir := flag.String("work_dir", ".", "Directory to workspace")
	codeLen := flag.Int("code_len", 6, "Length of the generated access codes")
	cleanup := flag.Duration("cleanup_interval", 24*time.Hour, "Interval for cleaning up expired files")
	expiry := flag.Duration("default_expiry", 24*time.Hour, "Default expiry time for uploads")
	portFlag := flag.Int("port", 8080, "Port to run the server on")
	enableZip := flag.Bool("enable_zip", false, "Enable zip compression")
	logFlag := flag.Bool("log", false, "log record to workspace file")
	login := flag.Int("login_flag", 0, "0:allow all handle,1:upload need login,2: download need login,3:all need login")
	users := flag.String("users", "", "User info example admin:admin,test:test")
	baseFlag := flag.String("base", "/", "html static resource base url example nginx need /share/")

	// Parse the flags
	flag.Parse()

	if initFlag {
		if err := installServiceScript(); err != nil {
			log.Fatal(err)
		} else {
			log.Println("service.sh generator successfully")
		}
		os.Exit(0)
		return
	}

	base = *baseFlag
	codeLength = *codeLen
	cleanupInterval = *cleanup
	defaultExpiry = *expiry
	port = *portFlag
	enableZipFlag = *enableZip
	loginFlag = *login
	if *users != "" {
		// 首先按逗号分割字符串
		pairs := strings.Split(*users, ",")

		// 遍历每个键值对
		for _, pair := range pairs {
			// 按冒号分割键和值
			parts := strings.Split(pair, ":")
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				userMap[key] = value
			}
		}
	}
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
			log.SetOutput(f)
		}
	}

	//给utils i8n 传递参数
	utils.Use(&localesFiles)

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
