package types

import "time"

// Provider 定义不同的图像生成提供商
type Provider string

const (
	ProviderSVGIO   Provider = "svgio"
	ProviderRecraft Provider = "recraft"
	ProviderClaude  Provider = "claude"
)

// API 请求和响应类型定义

type GenerateRequest struct {
	Prompt         string   `json:"prompt"`
	NegativePrompt string   `json:"negative_prompt,omitempty"`
	Style          string   `json:"style,omitempty"`
	Provider       Provider `json:"provider,omitempty"` // 新增：指定使用的提供商
	// 可选：前端区分用途（例如 png 或 svg_inline），当前普通 /v1/images 忽略该字段
	Format string `json:"format,omitempty"`

	// 新增：是否跳过翻译（当用户确定输入的是英文时）
	SkipTranslate bool `json:"skip_translate,omitempty"`

	// Recraft 特有参数
	Model     string `json:"model,omitempty"`    // recraftv3 或 recraftv2
	Size      string `json:"size,omitempty"`     // 图像尺寸，如 "1024x1024"
	Substyle  string `json:"substyle,omitempty"` // 子风格
	NumImages int    `json:"n,omitempty"`        // 生成图像数量 (1-6)
}

type ImageResponse struct {
	ID             string    `json:"id"`
	Prompt         string    `json:"prompt"`
	NegativePrompt string    `json:"negative_prompt"`
	Style          string    `json:"style"`
	SVGURL         string    `json:"svg_url"`
	PNGURL         string    `json:"png_url"`
	Width          int       `json:"width"`
	Height         int       `json:"height"`
	CreatedAt      time.Time `json:"created_at"`
	Provider       Provider  `json:"provider"` // 新增：使用的提供商
	// 新增：翻译相关信息
	OriginalPrompt   string `json:"original_prompt,omitempty"`   // 原始提示词
	TranslatedPrompt string `json:"translated_prompt,omitempty"` // 翻译后的提示词
	WasTranslated    bool   `json:"was_translated"`              // 是否进行了翻译
}

type ErrorResp struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// SVG.IO 上游 API 相关类型

type SVGIOGenerateReq struct {
	Prompt           string `json:"prompt"`
	NegativePrompt   string `json:"negativePrompt"`
	Style            string `json:"style,omitempty"`
	InitialImage     any    `json:"initialImage"`     // 必须传 null，否则会被序列化为 ""
	InitialImageType any    `json:"initialImageType"` // 同上
}

type SVGIOGenerateItem struct {
	ID                  string `json:"id"`
	Description         string `json:"description"`
	Height              int    `json:"height"`
	HasInitialImage     bool   `json:"hasInitialImage"`
	IsPrivate           bool   `json:"isPrivate"`
	NSFWTextDetected    bool   `json:"nsfwTextDetected"`
	NSFWContentDetected bool   `json:"nsfwContentDetected"`
	PNGURL              string `json:"pngUrl"`
	SVGURL              string `json:"svgUrl"`
	Style               string `json:"style"`
	Prompt              string `json:"prompt"`
	NegativePrompt      string `json:"negativePrompt"`
	Width               int    `json:"width"`
	UpdatedAt           string `json:"updatedAt"`
	CreatedAt           string `json:"createdAt"`
}

type SVGIOGenerateResp struct {
	Success bool                `json:"success"`
	Data    []SVGIOGenerateItem `json:"data"`
}

// Recraft API 相关类型

type RecraftGenerateReq struct {
	Prompt         string `json:"prompt"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
	Style          string `json:"style,omitempty"`
	Substyle       string `json:"substyle,omitempty"`
	Model          string `json:"model,omitempty"`
	Size           string `json:"size,omitempty"`
	N              int    `json:"n,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
}

type RecraftImageData struct {
	URL           string `json:"url"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

type RecraftGenerateResp struct {
	Created int                `json:"created"`
	Data    []RecraftImageData `json:"data"`
}

type RecraftVectorizeReq struct {
	ResponseFormat string `json:"response_format,omitempty"`
}

type RecraftVectorizeResp struct {
	Image RecraftImageData `json:"image"`
}

// Claude API 相关类型

type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClaudeGenerateReq struct {
	Model       string          `json:"model"`
	Messages    []ClaudeMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature,omitempty"`
	System      string          `json:"system,omitempty"`
}

type ClaudeGenerateResp struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}
