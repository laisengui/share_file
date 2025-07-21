package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"share_file/utils"
	"strconv"
	"time"
)

// startHttp 启动http服务
func startHttp() {
	// 设置路由
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/download/", downloadHandler)
	// 设置根路径处理
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "templates/index.html")
			return
		}
		http.NotFound(w, r)
	})
	// 设置静态文件服务
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 获取本机IP地址
	ip, err := utils.GetLocalIP()
	if err != nil {
		log.Printf("无法获取本地IP: %v, 使用localhost", err)
		ip = "localhost"
	}

	log.Printf("服务启动成功! 访问地址: http://%s:%d", ip, port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}

// uploadHandler 上传文件
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	lang := utils.DetectLanguage(r)
	if r.Method != http.MethodPost {
		http.Error(w, utils.T(lang, "supportPost"), http.StatusMethodNotAllowed)
		return
	}

	// 解析表单数据，限制上传大小为100MB
	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		http.Error(w, utils.T(lang, "parseFormError"), http.StatusBadRequest)
		return
	}

	// 获取文件
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, utils.T(lang, "fileNotFound"), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 获取有效期
	expiryHours, err := strconv.Atoi(r.FormValue("expiry"))
	if err != nil || expiryHours <= 0 {
		expiryHours = int(defaultExpiry.Hours())
	}
	// 获取下载次数
	times, err := strconv.Atoi(r.FormValue("times"))
	if err != nil || times <= 0 {
		times = -99
	}

	// 生成随机码
	code := utils.GenerateRandomCode(codeLength)

	// 创建目标文件
	id := uuid.New()
	filename := filepath.Join(uploadDir, id.String())
	dst, err := os.Create(filename)
	if err != nil {
		http.Error(w, utils.T(lang, "createFileError"), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// 复制文件内容
	size, err := io.Copy(dst, file)
	if err != nil {
		http.Error(w, utils.T(lang, "saveFileError"), http.StatusInternalServerError)
		return
	}

	// 记录文件信息
	fileInfo := FileInfo{
		Uuid:        id.String(),
		Filename:    handler.Filename,
		Code:        code,
		Size:        size,
		Times:       times,
		UploadTime:  time.Now(),
		ExpiryTime:  time.Now().Add(time.Duration(expiryHours) * time.Hour),
		ContentType: handler.Header.Get("Content-Type"),
	}

	err = database.SaveFile(code, fileInfo)
	if err != nil {
		http.Error(w, utils.T(lang, "saveFileError"), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	response := map[string]interface{}{
		"code":    code,
		"message": utils.T(lang, "fileUploadSuccess"),
		"expiry":  expiryHours,
		"times":   times,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// downloadHandler 下载文件
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	lang := utils.DetectLanguage(r)

	code := r.URL.Path[len("/download/"):]
	if len(code) != codeLength {
		http.Error(w, utils.T(lang, "fileUploadSuccess"), http.StatusBadRequest)
		return
	}

	fileInfo, err := database.GetFile(code)
	if err != nil {
		http.Error(w, utils.T(lang, "fileNotExistsOrExpire"), http.StatusNotFound)
		return
	}
	if time.Now().After(fileInfo.ExpiryTime) {
		http.Error(w, utils.T(lang, "fileExpire"), http.StatusGone)
		return
	}
	if fileInfo.Times < 1 && fileInfo.Times > -50 {
		http.Error(w, utils.T(lang, "fileNotTimes"), http.StatusGone)
		return
	}

	// 构建实际存储的文件名
	filename := filepath.Join(uploadDir, fileInfo.Uuid)

	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		http.Error(w, utils.T(lang, "fileNotFound"), http.StatusNotFound)
		return
	}

	//如果有设置次数 次数需要更新
	if fileInfo.Times > 0 {
		fileInfo.Times = fileInfo.Times - 1
		err = database.SaveFile(code, fileInfo)
		if err != nil {
			log.Printf("update fileinfo error %s", code)
		}
	}

	// 设置响应头
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileInfo.Filename))
	w.Header().Set("Content-Type", fileInfo.ContentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size))

	// 提供文件下载
	http.ServeFile(w, r, filename)
}
