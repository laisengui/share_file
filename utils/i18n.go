package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
)

var (
	translations = make(map[string]map[string]interface{})
	localeMutex  sync.RWMutex
	defaultLang  = "zh"
)

func init() {
	loadTranslations()
}

func loadTranslations() {
	localeMutex.Lock()
	defer localeMutex.Unlock()

	locales := []string{"zh", "en"}
	for _, lang := range locales {
		filePath := filepath.Join("static", "locales", lang+".json")
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Printf("Error loading %s translation: %v", lang, err)
			continue
		}

		var trans map[string]interface{}
		if err := json.Unmarshal(data, &trans); err != nil {
			log.Printf("Error parsing %s translation: %v", lang, err)
			continue
		}

		translations[lang] = trans
	}
}

func T(lang string, code string) string {
	return ToStr(GetTranslations(lang)[code])
}

func GetTranslations(lang string) map[string]interface{} {
	localeMutex.RLock()
	defer localeMutex.RUnlock()

	if trans, ok := translations[lang]; ok {
		return trans
	}
	return translations[defaultLang]
}

func DetectLanguage(r *http.Request) string {
	// 1. Check URL query parameter
	if lang := r.URL.Query().Get("lang"); lang != "" {
		if _, ok := translations[lang]; ok {
			return lang
		}
	}

	// 2. Check cookie
	if cookie, err := r.Cookie("lang"); err == nil {
		if _, ok := translations[cookie.Value]; ok {
			return cookie.Value
		}
	}

	// 3. Check Accept-Language header
	acceptLang := r.Header.Get("Accept-Language")
	if acceptLang != "" {
		lang := strings.Split(acceptLang, ",")[0]
		lang = strings.Split(lang, ";")[0]
		lang = strings.Split(lang, "-")[0]
		if _, ok := translations[lang]; ok {
			return lang
		}
	}

	// 4. Return default
	return defaultLang
}
