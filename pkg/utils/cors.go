package utils	
import (
	"log"
	"net/http"
)

// ========== 中间件和 CORS ==========
// WithCommonHeaders CORS middleware and common headers
func WithCommonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[MIDDLEWARE] %s %s from %s - User-Agent: %s", r.Method, r.URL.Path, r.RemoteAddr, r.Header.Get("User-Agent"))

		// CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
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

// SetCORSHeaders 设置CORS头
func SetCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	w.Header().Set("Access-Control-Expose-Headers", "X-Image-Id, X-Image-Width, X-Image-Height, Content-Disposition")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")
}
