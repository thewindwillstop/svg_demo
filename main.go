// main.go
package main

import (
	"log"
	"miniSvg/internal/handlers"
	"miniSvg/internal/translate"
	"miniSvg/pkg/utils"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	log.Printf("Starting SVG image generation service...")

	_ = godotenv.Load(".env")
	apiKey := os.Getenv("SVGIO_API_KEY")
	if apiKey == "" {
		log.Fatal("missing SVGIO_API_KEY")
	}
	log.Printf("API key loaded successfully (length: %d)", len(apiKey))

	// 初始化翻译服务
	var translateService translate.Service
	translateAPIKey := os.Getenv("OPENAI_API_KEY")
	if translateAPIKey != "" {
		translateService = translate.NewOpenAIService(translateAPIKey)
		log.Printf("Translation service initialized with OpenAI")
	} else {
		log.Printf("Warning: OPENAI_API_KEY not found, translation will be skipped")
	}

	mux := http.NewServeMux()

	// 注册路由处理器
	mux.HandleFunc("/v1/images/svg", handlers.SVGHandler(apiKey, translateService))
	mux.HandleFunc("/ping", handlers.PingHandler())
	mux.HandleFunc("/v1/images", handlers.ImageHandler(apiKey, translateService))
	mux.HandleFunc("/download", handlers.DownloadSVGHandler())

	addr := "0.0.0.0:8080"
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, utils.WithCommonHeaders(mux)); err != nil {
		log.Fatal(err)
	}
}
