package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"svg-generator/internal/config"
	"svg-generator/internal/types"
)

// HTTPClient is a global HTTP client with timeout
var HTTPClient = &http.Client{
	Timeout: 60 * time.Second,
}

// ========== HTTP 相关功能 ==========

// DownloadFile downloads a file from the given URL
func DownloadFile(ctx context.Context, fileURL string) ([]byte, error) {
	log.Printf("[DOWNLOAD] Starting download from: %s", fileURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		log.Printf("[DOWNLOAD] Failed to create request: %v", err)
		return nil, err
	}
	resp, err := HTTPClient.Do(req)
	if err != nil {
		log.Printf("[DOWNLOAD] HTTP request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("[DOWNLOAD] Received response with status: %s", resp.Status)

	if resp.StatusCode >= 300 {
		log.Printf("[DOWNLOAD] Bad status code: %d", resp.StatusCode)
		return nil, errors.New("fetch status: " + resp.Status)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[DOWNLOAD] Failed to read response body: %v", err)
		return nil, err
	}
	log.Printf("[DOWNLOAD] Successfully downloaded %d bytes", len(b))
	return b, nil
}

// WriteError writes an error response in JSON format
func WriteError(w http.ResponseWriter, status int, code, msg string, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(types.ErrorResp{Code: code, Message: msg, Details: details})
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// ========== 中间件和 CORS ==========

// WithCommonHeaders CORS middleware and common headers
func WithCommonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[MIDDLEWARE] %s %s from %s - User-Agent: %s", r.Method, r.URL.Path, r.RemoteAddr, r.Header.Get("User-Agent"))

		// CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Expose-Headers", "X-Image-Id, X-Image-Width, X-Image-Height, Content-Disposition")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// 其他安全/缓存
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")

		// 预检请求直接返回
		if r.Method == http.MethodOptions {
			log.Printf("[MIDDLEWARE] CORS preflight request handled")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SetCORSHeaders 设置CORS头
func SetCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	w.Header().Set("Access-Control-Expose-Headers", "X-Image-Id, X-Image-Width, X-Image-Height, Content-Disposition")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")
}

// ========== 辅助函数 ==========

// BytesReader creates a bytes reader from byte slice
func BytesReader(b []byte) *bytes.Reader {
	return bytes.NewReader(b)
}

// ToString converts integer to string
func ToString(v int) string {
	return strconv.Itoa(v)
}

// ReadAllBytes reads all bytes from a reader
func ReadAllBytes(r interface{}) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.(interface{ Read([]byte) (int, error) }))
	return buf.Bytes()
}

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
