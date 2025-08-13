// main.go
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// HTTP 客户端配置
var httpClient = &http.Client{
	Timeout: 100 * time.Second,
}

func main() {
	log.Printf("Starting SVG image generation service...")

	_ = godotenv.Load(".env")
	apiKey := os.Getenv("SVGIO_API_KEY")
	if apiKey == "" {
		log.Fatal("missing SVGIO_API_KEY")
	}
	log.Printf("API key loaded successfully (length: %d)", len(apiKey))

	mux := http.NewServeMux()

	// 注册路由处理器
	mux.HandleFunc("/v1/images/svg", handleSVGGeneration(apiKey))
	mux.HandleFunc("/ping", handlePing())
	mux.HandleFunc("/v1/images", handleImageGeneration(apiKey))
	mux.HandleFunc("/download", handleDownloadSVG())

	addr := "0.0.0.0:8080"
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, withCommonHeaders(mux)); err != nil {
		log.Fatal(err)
	}
}
