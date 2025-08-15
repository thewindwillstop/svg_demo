package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"svg-generator/internal/config"
	"svg-generator/internal/types"
	"svg-generator/pkg/utils"
	"time"
)

// RecraftService 实现 Recraft API 调用
type RecraftService struct {
	apiKey  string
	baseURL string
}

// NewRecraftService 创建 Recraft 服务实例
func NewRecraftService(apiKey string) *RecraftService {
	return &RecraftService{
		apiKey:  apiKey,
		baseURL: config.AppConfig.Providers.Recraft.BaseURL,
	}
}

// GenerateImage 使用 Recraft API 生成图像
func (s *RecraftService) GenerateImage(ctx context.Context, req types.GenerateRequest) (*types.ImageResponse, error) {
	log.Printf("[RECRAFT] Starting generation request...")

	// 构建优化后的提示词（添加无背景要求）
	enhancedPrompt, enhancedNegativePrompt := s.buildRecraftPrompt(req.Prompt, req.Style, req.NegativePrompt)

	log.Printf("[RECRAFT] Original prompt: %s", req.Prompt)
	log.Printf("[RECRAFT] Enhanced prompt: %s", enhancedPrompt)
	log.Printf("[RECRAFT] Enhanced negative prompt: %s", enhancedNegativePrompt)

	// 构建 Recraft API 请求
	recraftReq := types.RecraftGenerateReq{
		Prompt:         enhancedPrompt,
		NegativePrompt: enhancedNegativePrompt,
		Style:          req.Style,
		Substyle:       req.Substyle,
		Model:          req.Model,
		Size:           req.Size,
		N:              req.NumImages,
		ResponseFormat: "url", // 固定使用 URL 格式
	}

	// 设置默认值
	if recraftReq.Model == "" {
		recraftReq.Model = "recraftv2"
	}
	if recraftReq.Size == "" {
		recraftReq.Size = "1024x1024"
	}
	if recraftReq.Style == "" {
		recraftReq.Style = "vector_illustration"
	}
	if recraftReq.N == 0 {
		recraftReq.N = 1
	}

	body, err := json.Marshal(recraftReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := s.baseURL + config.AppConfig.Providers.Recraft.Endpoints.Generate
	log.Printf("[RECRAFT] Sending request to %s with payload size: %d bytes", url, len(body))

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := utils.HTTPClient.Do(httpReq)
	if err != nil {
		log.Printf("[RECRAFT] HTTP request failed: %v", err)
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[RECRAFT] Received response with status: %s", resp.Status)

	if resp.StatusCode >= 300 {
		var errResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		log.Printf("[RECRAFT] Error response body: %+v", errResp)
		return nil, fmt.Errorf("recraft API error: %s", resp.Status)
	}

	var recraftResp types.RecraftGenerateResp
	if err := json.NewDecoder(resp.Body).Decode(&recraftResp); err != nil {
		log.Printf("[RECRAFT] Failed to decode response: %v", err)
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(recraftResp.Data) == 0 {
		log.Printf("[RECRAFT] No images in response")
		return nil, errors.New("no images generated")
	}

	imageData := recraftResp.Data[0] // 取第一张图片
	log.Printf("[RECRAFT] Successfully parsed response - URL: %s", imageData.URL)

	// 解析图片尺寸
	width, height := parseSizeFromString(recraftReq.Size)

	// 生成一个简单的 ID（Recraft 不提供）
	imageID := generateImageID()

	// 对于 Recraft，我们需要将图片转换为 SVG
	// 这里我们可以使用 Recraft 的 vectorize API
	svgURL := imageData.URL // 默认使用原图

	// 如果需要 SVG，调用 vectorize API
	if req.Format == "svg" || strings.Contains(req.Style, "vector") {
		vectorizedURL, err := s.vectorizeImage(ctx, imageData.URL)
		if err != nil {
			log.Printf("[RECRAFT] Vectorization failed: %v", err)
			// 失败时继续使用原图
		} else {
			svgURL = vectorizedURL
		}
	}

	return &types.ImageResponse{
		ID:             imageID,
		Prompt:         req.Prompt,
		NegativePrompt: req.NegativePrompt,
		Style:          recraftReq.Style,
		SVGURL:         svgURL,
		PNGURL:         imageData.URL,
		Width:          width,
		Height:         height,
		CreatedAt:      time.Unix(int64(recraftResp.Created), 0),
		Provider:       types.ProviderRecraft,
	}, nil
}

// vectorizeImage 使用 Recraft 的向量化 API 将图片转换为 SVG
func (s *RecraftService) vectorizeImage(ctx context.Context, imageURL string) (string, error) {
	log.Printf("[RECRAFT] Vectorizing image: %s", imageURL)

	// 下载图片
	imageBytes, err := utils.DownloadFile(ctx, imageURL)
	if err != nil {
		return "", fmt.Errorf("download image: %w", err)
	}

	// 创建 multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加图片文件
	part, err := writer.CreateFormFile("file", "image.png")
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}
	_, err = part.Write(imageBytes)
	if err != nil {
		return "", fmt.Errorf("write image data: %w", err)
	}

	// 添加 response_format 参数
	writer.WriteField("response_format", "url")
	writer.Close()

	url := s.baseURL + config.AppConfig.Providers.Recraft.Endpoints.Vectorize
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return "", fmt.Errorf("create vectorize request: %w", err)
	}

	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := utils.HTTPClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("vectorize request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("vectorize API error: %s", resp.Status)
	}

	var vectorizeResp types.RecraftVectorizeResp
	if err := json.NewDecoder(resp.Body).Decode(&vectorizeResp); err != nil {
		return "", fmt.Errorf("decode vectorize response: %w", err)
	}

	log.Printf("[RECRAFT] Vectorization successful - SVG URL: %s", vectorizeResp.Image.URL)
	return vectorizeResp.Image.URL, nil
}

// parseSizeFromString 解析尺寸字符串（如 "1024x1024"）
func parseSizeFromString(size string) (width, height int) {
	parts := strings.Split(size, "x")
	if len(parts) != 2 {
		return 1024, 1024 // 默认值
	}

	width, _ = strconv.Atoi(parts[0])
	height, _ = strconv.Atoi(parts[1])

	if width <= 0 {
		width = 1024
	}
	if height <= 0 {
		height = 1024
	}

	return width, height
}

// generateImageID 生成简单的图片 ID
func generateImageID() string {
	return fmt.Sprintf("recraft_%d", time.Now().UnixNano())
}

// buildRecraftPrompt 构建Recraft的提示词，自动添加无背景要求
func (s *RecraftService) buildRecraftPrompt(prompt, style, negativePrompt string) (string, string) {
	var promptBuilder strings.Builder
	var negativeBuilder strings.Builder

	// 构建主提示词
	promptBuilder.WriteString(prompt)

	promptBuilder.WriteString(", 背景色为透明色，不要背景边框")

	// 构建负面提示词
	if negativePrompt != "" {
		negativeBuilder.WriteString(negativePrompt)
		negativeBuilder.WriteString(", ")
	}

	// 添加默认的背景相关负面提示词

	return promptBuilder.String(), negativeBuilder.String()
}
