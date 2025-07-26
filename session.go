package main

import (
	"net/http"
	"share_file/utils"
	"time"
)

type Session struct {
	username  string
	token     string
	timestamp int64 //过期的时间戳
}

var sessionTime int64 = 1800

// session
var session = map[string]*Session{}

func generateSessionToken(username string) string {
	code := utils.GenerateRandomCode(10)
	session[code] = &Session{username: username, token: code, timestamp: time.Now().Unix() + sessionTime}
	return code
}

// isLoggedIn 检查用户是否已登录
func isLoggedIn(r *http.Request) (bool, *Session) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return false, nil
	}

	// 解析Cookie值
	username, token := parseSessionCookie(cookie.Value)
	if username == "" || token == "" {
		return false, nil
	}
	s, ok := session[token]
	if !ok {
		return false, nil
	}
	now := time.Now().Unix()
	// 验证令牌
	if s.username == username {
		if s.timestamp > now {
			s.timestamp = time.Now().Unix() + sessionTime
			return true, s
		} else {
			delete(session, token)
		}

	}
	return false, nil
}

// parseSessionCookie 解析会话Cookie 用户:token
func parseSessionCookie(cookieValue string) (string, string) {
	for i := 0; i < len(cookieValue); i++ {
		if cookieValue[i] == '|' {
			return cookieValue[:i], cookieValue[i+1:]
		}
	}
	return "", ""
}
