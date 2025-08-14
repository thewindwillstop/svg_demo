package config

// API 配置常量
const (
	UpstreamBaseURL = "https://api.svg.io"
	GeneratePath    = "/v1/generate-image"
	GetImagePath    = "/v1/get-image/" // + {imageId}
)

// 翻译服务配置
const (
	TranslateServiceURL = "https://api.siliconflow.cn/v1/chat/completions" // OpenAI翻译
	DefaultModel        = "deepseek-ai/DeepSeek-R1-0528-Qwen3-8B"                              // 默认翻译模型
)
