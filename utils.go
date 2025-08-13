package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// HTTP 响应辅助函数
func writeError(w http.ResponseWriter, status int, code, msg string, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResp{Code: code, Message: msg, Details: details})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// CORS 中间件
func withCommonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[MIDDLEWARE] %s %s from %s - User-Agent: %s", r.Method, r.URL.Path, r.RemoteAddr, r.Header.Get("User-Agent"))

		// CORS
		w.Header().Set("Access-Control-Allow-Origin", "*") // 放宽全部域；生产可改为具体域
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Expose-Headers", "X-Image-Id, X-Image-Width, X-Image-Height, Content-Disposition")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// 其他安全/缓存
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")

		// 预检请求直接返回
		if r.Method == http.MethodOptions {
			log.Printf("[MIDDLEWARE] CORS preflight request handled")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// 工具函数
func bytesReader(b []byte) *bytes.Reader {
	return bytes.NewReader(b)
}

func toString(v int) string {
	return strconv.Itoa(v)
}

func readAllBytes(r interface{}) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.(interface{ Read([]byte) (int, error) }))
	return buf.Bytes()
}
