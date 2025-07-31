package main

import (
	"crypto/subtle"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"share_file/utils"
	"strconv"
	"strings"
	"time"
)

//go:embed static/*
var staticFiles embed.FS

// startHttp 启动http服务
func startHttp() {
	// 设置路由

	http.HandleFunc("/login/status", loginStatusHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/download/", downloadHandler)

	// 设置静态文件服务
	// 创建子文件系统，去掉 static/ 前缀
	subFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	dir := http.FileServer(http.FS(subFS))
	http.Handle("/static/", http.StripPrefix("/static/", dir))

	// 设置根路径处理
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Redirect(w, r, "/", http.StatusFound) // 302 临时重定向
			return
		}
		data, err := fs.ReadFile(subFS, "index.html")
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		if base != "/" {
			html := string(data)
			s := strings.ReplaceAll(html, "<base href=\"/\">", "<base href=\""+base+"\">")
			data = []byte(s)
		}

		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
	})

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

	log.Println("上传文件")
	lang := utils.DetectLanguage(r)

	if loginFlag == 1 || loginFlag == 3 {
		if login, _ := isLoggedIn(r); !login {
			log.Println("上传文件需要登录")
			http.Error(w, utils.T(lang, "needLogin"), http.StatusUnauthorized)
			return
		}
	}

	if r.Method != http.MethodPost {
		http.Error(w, utils.T(lang, "supportPost"), http.StatusMethodNotAllowed)
		return
	}

	// 解析表单数据，限制上传大小为1000MB
	r.Body = http.MaxBytesReader(w, r.Body, maxMbFlag<<20)
	err := r.ParseMultipartForm(maxMbFlag << 20)
	if err != nil {
		log.Printf("上传异常 %v", err)
		http.Error(w, utils.T(lang, "parseFormError"), http.StatusBadRequest)
		return
	}

	// 获取文件
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Printf("上传异常 %v", err)
		http.Error(w, utils.T(lang, "fileNotFound"), http.StatusBadRequest)
		return
	}
	defer file.Close()

	filedata, err := io.ReadAll(file)
	if err != nil {
		log.Printf("上传异常 %v", err)
		http.Error(w, utils.T(lang, "fileReadError"), http.StatusBadRequest)
		return
	}
	hasCompress := false
	if enableZipFlag {
		if !utils.IsZip(filedata) {
			filedata, err = utils.CompressSingleFileToZip("file", filedata)
			if err != nil {
				log.Printf("上传异常 %v", err)
				http.Error(w, utils.T(lang, "fileCompressError"), http.StatusBadRequest)
				return
			}
			hasCompress = true
		}
	}

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
	var code string
	for {
		code = utils.GenerateRandomCode(codeLength)
		_, e := database.GetFile(code)
		if e != nil {
			//检测冲突, 报错代表不冲突
			break
		}
	}

	// 创建目标文件
	id := uuid.New()
	filename := filepath.Join(uploadDir, id.String())
	dst, err := os.Create(filename)
	if err != nil {
		log.Printf("上传异常 %v", err)
		http.Error(w, utils.T(lang, "createFileError"), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// 复制文件内容
	size, err := dst.Write(filedata)
	if err != nil {
		log.Printf("上传异常 %v", err)
		http.Error(w, utils.T(lang, "saveFileError"), http.StatusInternalServerError)
		return
	}

	// 记录文件信息
	fileInfo := FileInfo{
		Uuid:        id.String(),
		Filename:    handler.Filename,
		Code:        code,
		Size:        int64(size),
		Times:       times,
		Compression: hasCompress,
		UploadTime:  time.Now(),
		ExpiryTime:  time.Now().Add(time.Duration(expiryHours) * time.Hour),
		ContentType: handler.Header.Get("Content-Type"),
	}

	err = database.SaveFile(code, fileInfo)
	if err != nil {
		log.Printf("上传异常 %v", err)
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

	log.Println("开始下载")
	if loginFlag == 2 || loginFlag == 3 {
		if login, _ := isLoggedIn(r); !login {
			log.Println("权限不足")
			http.Error(w, utils.T(lang, "needLogin"), http.StatusUnauthorized)
			return
		}
	}

	code := r.URL.Path[len("/download/"):]

	log.Printf("下载代码 %s /n", code)
	if len(code) != codeLength {
		http.Error(w, utils.T(lang, "fileUploadSuccess"), http.StatusBadRequest)
		return
	}

	fileInfo, err := database.GetFile(code)
	if err != nil {
		log.Printf("下载代码 %s 文件不存在/n", code)
		http.Error(w, utils.T(lang, "fileNotExistsOrExpire"), http.StatusNotFound)
		return
	}
	if time.Now().After(fileInfo.ExpiryTime) {
		log.Printf("下载代码 %s 文件过期/n", code)
		http.Error(w, utils.T(lang, "fileExpire"), http.StatusGone)
		return
	}
	if fileInfo.Times < 1 && fileInfo.Times > -50 {
		log.Printf("下载代码 %s 文件次数不足/n", code)
		http.Error(w, utils.T(lang, "fileNotTimes"), http.StatusGone)
		return
	}

	// 构建实际存储的文件名
	filename := filepath.Join(uploadDir, fileInfo.Uuid)

	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Printf("下载代码 %s 文件没找到/n", code)
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
	encoded := url.PathEscape(fileInfo.Filename)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", encoded))
	//w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", encoded))
	w.Header().Set("Content-Type", fileInfo.ContentType)
	//w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size))

	// 提供文件下载
	http.ServeFile(w, r, filename)
	log.Printf("下载代码 %s 下载成功/n", code)
}

// loginStatusHandler 检查登录状态
func loginStatusHandler(w http.ResponseWriter, r *http.Request) {

	// 验证令牌
	login, s := isLoggedIn(r)
	username := ""
	if login {
		username = s.username
	}
	// 返回成功响应
	response := map[string]interface{}{
		"loggedIn": login,
		"username": username,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// loginHandler 登录处理器
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// 处理登录请求
		username := r.FormValue("username")
		password := r.FormValue("password")

		log.Printf("用户登录 %s \n", username)

		// 验证用户名和密码
		storedPassword, ok := userMap[username]
		if !ok || subtle.ConstantTimeCompare([]byte(password), []byte(storedPassword)) != 1 {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		// 创建Cookie
		expiration := time.Now().Add(24 * time.Hour)
		cookie := http.Cookie{
			Name:     "session",
			Value:    username + "|" + generateSessionToken(username),
			Expires:  expiration,
			HttpOnly: true,
			Path:     "/",
		}
		http.SetCookie(w, &cookie)

		// 返回成功响应
		response := map[string]interface{}{
			"username": username,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		log.Printf("用户登录成功 %s \n", username)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// logoutHandler 登出处理器
func logoutHandler(w http.ResponseWriter, r *http.Request) {

	login, s := isLoggedIn(r)
	if login {
		log.Printf("用户登出 %s \n", s.username)
		delete(userMap, s.token)
	}
	// 删除Cookie
	cookie := http.Cookie{
		Name:     "session",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, &cookie)
	// 返回成功响应
	response := map[string]interface{}{
		"username": "",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
