// main.go
package main

import (
	"log"
	"net/http"
	"os" // 创建服务管理器
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

	// 加载配置文件
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	if err := config.InitConfig(configPath); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded successfully from: %s", configPath)

	// 加载 API 密钥
	svgioAPIKey := os.Getenv("SVGIO_API_KEY")
	recraftAPIKey := os.Getenv("RECRAFT_API_KEY")
	claudeAPIKey := os.Getenv("CLAUDE_API_KEY")
	claudeBaseURL := os.Getenv("CLAUDE_BASE_URL")

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
	if claudeAPIKey != "" && config.AppConfig.IsProviderEnabled("claude") {
		log.Printf("Claude API key loaded successfully (length: %d)", len(claudeAPIKey))
		enabledProviders++
	}

	if enabledProviders == 0 {
		log.Fatal("No providers available: either API keys are missing or all providers are disabled in config")
	}

	// 初始化服务管理器
	serviceManager := service.NewServiceManager(svgioAPIKey, recraftAPIKey, claudeAPIKey, claudeBaseURL)
	log.Printf("Service manager initialized with available providers")

	// 初始化翻译服务
	var translateService utils.TranslateService
	translateAPIKey := os.Getenv("OPENAI_API_KEY")
	if translateAPIKey != "" && config.AppConfig.Translation.Enabled {
		translateService = utils.NewOpenAITranslateService(translateAPIKey)
		log.Printf("Translation service initialized with OpenAI")
	} else {
		log.Printf("Warning: Translation service disabled or OPENAI_API_KEY not found")
	}

	mux := http.NewServeMux()

	// 注册路由处理器 - SVG.IO 提供商
	if svgioAPIKey != "" && config.AppConfig.IsProviderEnabled("svgio") {
		mux.HandleFunc("/v1/images/svgio/svg", handlers.SVGHandler(serviceManager, translateService))
		mux.HandleFunc("/v1/images/svgio", handlers.ImageHandler(serviceManager, translateService))
		log.Printf("SVG.IO routes registered")
	}

	// 注册路由处理器 - Recraft 提供商
	if recraftAPIKey != "" && config.AppConfig.IsProviderEnabled("recraft") {
		mux.HandleFunc("/v1/images/recraft/svg", handlers.RecraftSVGHandler(serviceManager, translateService))
		mux.HandleFunc("/v1/images/recraft", handlers.RecraftImageHandler(serviceManager, translateService))
		log.Printf("Recraft routes registered")
	}

	// 注册路由处理器 - Claude 提供商
	if claudeAPIKey != "" && config.AppConfig.IsProviderEnabled("claude") {
		mux.HandleFunc("/v1/images/claude/svg", handlers.ClaudeSVGHandler(serviceManager, translateService))
		mux.HandleFunc("/v1/images/claude", handlers.ClaudeImageHandler(serviceManager, translateService))
		log.Printf("Claude routes registered")
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

	addr := config.AppConfig.GetServerAddr()
	log.Printf("listening on %s", addr)
	log.Printf("Available endpoints:")
	if svgioAPIKey != "" && config.AppConfig.IsProviderEnabled("svgio") {
		log.Printf("  - POST /v1/images/svgio/svg   (SVG.IO - direct SVG download)")
		log.Printf("  - POST /v1/images/svgio       (SVG.IO - JSON metadata)")
	}
	if recraftAPIKey != "" && config.AppConfig.IsProviderEnabled("recraft") {
		log.Printf("  - POST /v1/images/recraft/svg (Recraft - direct SVG download)")
		log.Printf("  - POST /v1/images/recraft     (Recraft - JSON metadata)")
	}
	if claudeAPIKey != "" && config.AppConfig.IsProviderEnabled("claude") {
		log.Printf("  - POST /v1/images/claude/svg  (Claude - direct SVG download)")
		log.Printf("  - POST /v1/images/claude      (Claude - JSON metadata)")
	}
	log.Printf("  - GET  /health                 (Health check)")

	if err := http.ListenAndServe(addr, utils.WithCommonHeaders(mux)); err != nil {
		log.Fatal(err)
	}
}
