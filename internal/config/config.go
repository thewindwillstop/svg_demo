package config

// SVG.IO API 配置常量
const (
	SVGIOBaseURL      = "https://api.svg.io"
	SVGIOGeneratePath = "/v1/generate-image"
	SVGIOGetImagePath = "/v1/get-image/" // + {imageId}
)

// Recraft API 配置常量
const (
	RecraftBaseURL       = "https://external.api.recraft.ai"
	RecraftGeneratePath  = "/v1/images/generations"
	RecraftVectorizePath = "/v1/images/vectorize"
)

// 翻译服务配置
const (
	TranslateServiceURL = "https://api.siliconflow.cn/v1/chat/completions" // OpenAI翻译
	DefaultModel        = "deepseek-ai/DeepSeek-R1-0528-Qwen3-8B"          // 默认翻译模型
)
