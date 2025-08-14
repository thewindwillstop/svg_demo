package types

import "time"

// API 请求和响应类型定义

type GenerateRequest struct {
	Prompt         string `json:"prompt"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
	Style          string `json:"style,omitempty"`
	// 可选：前端区分用途（例如 png 或 svg_inline），当前普通 /v1/images 忽略该字段
	Format string `json:"format,omitempty"`
	
	// 新增：是否跳过翻译（当用户确定输入的是英文时）
	SkipTranslate bool `json:"skip_translate,omitempty"`
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

// 上游 API 相关类型

type UpstreamGenerateReq struct {
	Prompt           string `json:"prompt"`
	NegativePrompt   string `json:"negativePrompt"`
	Style            string `json:"style,omitempty"`
	InitialImage     any    `json:"initialImage"`     // 必须传 null，否则会被序列化为 ""
	InitialImageType any    `json:"initialImageType"` // 同上
}

type UpstreamGenerateItem struct {
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

type UpstreamGenerateResp struct {
	Success bool                   `json:"success"`
	Data    []UpstreamGenerateItem `json:"data"`
}
