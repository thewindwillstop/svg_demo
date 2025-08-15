// main.go
package main

import (
	"log"
	"miniSvg/internal/handlers"
	"miniSvg/internal/translate"
	"miniSvg/internal/upstream"
	"miniSvg/pkg/utils"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	log.Printf("Starting multi-provider SVG image generation service...")

	_ = godotenv.Load(".env")

	// 加载 API 密钥
	svgioAPIKey := os.Getenv("SVGIO_API_KEY")
	recraftAPIKey := os.Getenv("RECRAFT_API_KEY")

	if svgioAPIKey == "" && recraftAPIKey == "" {
		log.Fatal("At least one of SVGIO_API_KEY or RECRAFT_API_KEY must be provided")
	}

	if svgioAPIKey != "" {
		log.Printf("SVG.IO API key loaded successfully (length: %d)", len(svgioAPIKey))
	}
	if recraftAPIKey != "" {
		log.Printf("Recraft API key loaded successfully (length: %d)", len(recraftAPIKey))
	}

	// 初始化服务管理器
	serviceManager := upstream.NewServiceManager(svgioAPIKey, recraftAPIKey)
	log.Printf("Service manager initialized with available providers")

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

	// 注册路由处理器 - SVG.IO 提供商
	if svgioAPIKey != "" {
		mux.HandleFunc("/v1/images/svg", handlers.SVGHandler(serviceManager, translateService))
		mux.HandleFunc("/v1/images/svgio", handlers.SVGHandler(serviceManager, translateService))
		mux.HandleFunc("/v1/images", handlers.ImageHandler(serviceManager, translateService))
		log.Printf("SVG.IO routes registered")
	}

	// 注册路由处理器 - Recraft 提供商
	if recraftAPIKey != "" {
		mux.HandleFunc("/v1/images/recraft/svg", handlers.RecraftSVGHandler(serviceManager, translateService))
		mux.HandleFunc("/v1/images/recraft", handlers.RecraftImageHandler(serviceManager, translateService))
		log.Printf("Recraft routes registered")
	}

	// 通用路由
	mux.HandleFunc("/health", handlers.HealthHandler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			handlers.CORSPreflight()(w, r)
			return
		}
		http.NotFound(w, r)
	})

	addr := "0.0.0.0:8080"
	log.Printf("listening on %s", addr)
	log.Printf("Available endpoints:")
	if svgioAPIKey != "" {
		log.Printf("  - POST /v1/images/svg      (SVG.IO - direct SVG download)")
		log.Printf("  - POST /v1/images/svgio    (SVG.IO - direct SVG download)")
		log.Printf("  - POST /v1/images          (SVG.IO - JSON metadata)")
	}
	if recraftAPIKey != "" {
		log.Printf("  - POST /v1/images/recraft/svg (Recraft - direct SVG download)")
		log.Printf("  - POST /v1/images/recraft     (Recraft - JSON metadata)")
	}
	log.Printf("  - GET  /health             (Health check)")

	if err := http.ListenAndServe(addr, utils.WithCommonHeaders(mux)); err != nil {
		log.Fatal(err)
	}
}
