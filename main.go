// main.go
package main

import (
	"log"
	"net/http"
	"svg-generator/internal/config"
	"svg-generator/internal/handlers"
	"svg-generator/internal/service"
	"svg-generator/pkg/utils"

	"github.com/joho/godotenv"
)

func main() {
	log.Printf("Starting multi-provider SVG image generation service...")

	// 加载环境变量
	_ = godotenv.Load(".env")

	// 从环境变量初始化配置
	if err := config.InitConfig(); err != nil {
		log.Fatalf("Failed to load configuration from environment: %v", err)
	}

	log.Printf("Configuration loaded successfully from environment variables")

	// 加载 API 密钥
	svgioAPIKey := utils.GetEnv("SVGIO_API_KEY", "")
	recraftAPIKey := utils.GetEnv("RECRAFT_API_KEY", "")
	openaiAPIKey := utils.GetEnv("OPENAI_API_KEY", "")

	// 验证至少有一个Provider可用
	enabledProviders := 0
	if svgioAPIKey != "" && config.AppConfig.IsProviderEnabled("svgio") {
		log.Printf("SVG.IO API key loaded successfully (length: %d)", len(svgioAPIKey))
		enabledProviders++
	}
	if recraftAPIKey != "" && config.AppConfig.IsProviderEnabled("recraft") {
		log.Printf("Recraft API key loaded successfully (length: %d)", len(recraftAPIKey))
		enabledProviders++
	}
	if openaiAPIKey != "" && config.AppConfig.IsProviderEnabled("openai") {
		log.Printf("OpenAI API key loaded successfully (length: %d)", len(openaiAPIKey))
		enabledProviders++
	}

	if enabledProviders == 0 {
		log.Fatal("No providers available: either API keys are missing or all providers are disabled in config")
	}

	// 初始化服务管理器
	serviceManager := service.NewServiceManager(svgioAPIKey, recraftAPIKey, openaiAPIKey)
	log.Printf("Service manager initialized with available providers")

	// 初始化翻译服务
	var translateService utils.TranslateService
	if openaiAPIKey != "" && config.AppConfig.Translation.Enabled {
		translateService = utils.NewOpenAITranslateService(openaiAPIKey)
		log.Printf("Translation service initialized with OpenAI")
	} else {
		log.Printf("Warning: Translation service disabled or OPENAI_API_KEY not found")
	}

	mux := http.NewServeMux()

	// 统一路由 - 生成图像原数据 (JSON metadata)
	mux.HandleFunc("/v1/images", handlers.UnifiedImageHandler(serviceManager, translateService))

	// 统一路由 - 生成SVG (direct SVG download)
	mux.HandleFunc("/v1/images/svg", handlers.UnifiedSVGHandler(serviceManager, translateService))

	// 通用路由
	mux.HandleFunc("/health", handlers.HealthHandler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			handlers.CORSPreflight()(w, r)
			return
		}
		http.NotFound(w, r)
	})

	addr := config.AppConfig.GetServerAddr()
	log.Printf("listening on %s", addr)
	log.Printf("Available endpoints:")
	log.Printf("  - POST /v1/images     (Generate image metadata - supports provider parameter)")
	log.Printf("  - POST /v1/images/svg (Generate SVG - supports provider parameter)")
	log.Printf("  - GET  /health        (Health check)")
	log.Printf("Supported providers: svgio, recraft, openai")

	if err := http.ListenAndServe(addr, utils.WithCommonHeaders(mux)); err != nil {
		log.Fatal(err)
	}
}
