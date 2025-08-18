package utils
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	
	"svg-generator/internal/config")

// ========== 翻译服务 ==========

// OpenAI API 请求结构
type openaiTranslateRequest struct {
	Model       string                   `json:"model"`
	Messages    []openaiTranslateMessage `json:"messages"`
	MaxTokens   int                      `json:"max_tokens"`
	Temperature float64                  `json:"temperature"`
}

type openaiTranslateMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAI API 响应结构
type openaiTranslateResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// TranslateService 翻译服务接口
type TranslateService interface {
	Translate(ctx context.Context, text string) (string, error)
}

// OpenAITranslateService OpenAI 翻译服务实现
type OpenAITranslateService struct {
	apiKey string
	model  string
}

// NewOpenAITranslateService 创建OpenAI翻译服务实例
func NewOpenAITranslateService(apiKey string) *OpenAITranslateService {
	return &OpenAITranslateService{
		apiKey: apiKey,
		model:  config.AppConfig.Translation.DefaultModel,
	}
}

// Translate 翻译文本
func (s *OpenAITranslateService) Translate(ctx context.Context, text string) (string, error) {
	// 检测是否包含中文字符
	if !ContainsChinese(text) {
		log.Printf("[TRANSLATE] Text appears to be English already, skipping translation: %q", text)
		return text, nil
	}

	log.Printf("[TRANSLATE] Translating text: %q", text)

	prompt := fmt.Sprintf(`请将以下文本翻译成英文，保持原意，适合用作AI图像生成的提示词。只返回翻译结果，不要其他解释：

%s`, text)

	reqBody := openaiTranslateRequest{
		Model: s.model,
		Messages: []openaiTranslateMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   150,
		Temperature: 0.3,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.AppConfig.Translation.ServiceURL, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		log.Printf("[TRANSLATE] HTTP request failed: %v", err)
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	var translateResp openaiTranslateResponse
	if err := json.NewDecoder(resp.Body).Decode(&translateResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if translateResp.Error != nil {
		log.Printf("[TRANSLATE] API error: %s", translateResp.Error.Message)
		return "", fmt.Errorf("translate API error: %s", translateResp.Error.Message)
	}

	if len(translateResp.Choices) == 0 {
		return "", errors.New("no translation choices returned")
	}

	translated := strings.TrimSpace(translateResp.Choices[0].Message.Content)
	log.Printf("[TRANSLATE] Translation result: %q -> %q", text, translated)

	return translated, nil
}

// ContainsChinese 检测文本是否包含中文字符
func ContainsChinese(text string) bool {
	for _, char := range text {
		if char >= 0x4e00 && char <= 0x9fff {
			return true
		}
	}
	return false
}
