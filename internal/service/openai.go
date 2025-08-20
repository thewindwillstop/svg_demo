package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"svg-generator/internal/config"
	"svg-generator/internal/types"
	"svg-generator/pkg/utils"
	"time"
)

// OpenAIService 实现 OpenAI 兼容的 API 调用
type OpenAIService struct {
	apiKey  string
	baseURL string
}

// NewOpenAIService 创建 OpenAI 服务实例
func NewOpenAIService(apiKey, baseURL string) *OpenAIService {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAIService{
		apiKey:  apiKey,
		baseURL: strings.TrimSuffix(baseURL, "/"),
	}
}

// GenerateImage 使用 OpenAI 兼容的模型生成 SVG 代码
func (s *OpenAIService) GenerateImage(ctx context.Context, req types.GenerateRequest) (*types.ImageResponse, error) {
	log.Printf("[OPENAI] Starting SVG generation request...")

	// 构建 OpenAI 提示词
	prompt := s.buildSVGPrompt(req.Prompt, req.Style, req.NegativePrompt)

	// 构建 OpenAI API 请求
	openaiReq := types.OpenAIGenerateReq{
		Model:       getModelFromConfig(req.Model),
		MaxTokens:   getMaxTokensFromConfig(),
		Temperature: getTemperatureFromConfig(),
		Messages: []types.OpenAIMessage{
			{
				Role: "system",
				Content: `You are a world-class SVG graphics designer and vector artist with expertise in creating stunning, precise, and semantically meaningful SVG illustrations. Your specialties include:

1. **Technical Excellence**: You create perfectly valid, optimized SVG code that renders flawlessly across all browsers and devices
2. **Visual Design**: You have an exceptional eye for composition, color theory, typography, and visual hierarchy
3. **Style Adaptation**: You can seamlessly adapt to any artistic style - from minimalist line art to detailed illustrations, from cartoon to realistic, from modern flat design to vintage aesthetics
4. **Semantic Structure**: You use meaningful element IDs, proper grouping, and clean hierarchical structure in your SVG code
5. **Optimization**: Your SVG code is clean, efficient, and follows best practices for file size and performance

When creating SVG graphics, you:
- Pay careful attention to the exact subject, style, and mood requested
- Use appropriate colors, gradients, and visual effects to match the desired aesthetic
- Ensure proper proportions, perspective, and composition
- Add fine details that enhance the overall quality and realism
- Create scalable graphics that look crisp at any size
- Follow accessibility best practices when relevant

You respond ONLY with clean, valid SVG code - no explanations, no code blocks, just the pure SVG markup ready to render.`,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	body, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := s.baseURL + "/chat/completions"
	log.Printf("[OPENAI] Sending request to %s", url)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := utils.HTTPClient.Do(httpReq)
	if err != nil {
		log.Printf("[OPENAI] HTTP request failed: %v", err)
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[OPENAI] Received response with status: %s", resp.Status)

	if resp.StatusCode >= 300 {
		var errResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		log.Printf("[OPENAI] Error response body: %+v", errResp)
		return nil, fmt.Errorf("openai API error: %s", resp.Status)
	}

	// 读取原始响应内容进行调试
	bodyBytes := make([]byte, 0)
	if body, err := io.ReadAll(resp.Body); err == nil {
		bodyBytes = body
		log.Printf("[OPENAI] Raw response body: %s", string(body))
	}

	// 重新创建Reader用于JSON解码
	bodyReader := bytes.NewReader(bodyBytes)

	var openaiResp types.OpenAIGenerateResp
	if err := json.NewDecoder(bodyReader).Decode(&openaiResp); err != nil {
		log.Printf("[OPENAI] Failed to decode response: %v", err)
		return nil, fmt.Errorf("decode openai response: %w", err)
	}

	// 添加调试信息
	log.Printf("[OPENAI] Response structure: ID=%s, Object=%s, Model=%s, Choices length=%d",
		openaiResp.ID, openaiResp.Object, openaiResp.Model, len(openaiResp.Choices))

	// 如果有内容，打印第一个choice的内容
	if len(openaiResp.Choices) > 0 {
		firstChoice := openaiResp.Choices[0]
		log.Printf("[OPENAI] First choice: Role=%s, Content prefix=%s",
			firstChoice.Message.Role, truncateString(firstChoice.Message.Content, 100))
	}

	if len(openaiResp.Choices) == 0 {
		log.Printf("[OPENAI] Choices array is empty, response: %+v", openaiResp)
		return nil, fmt.Errorf("no choices in openai response")
	}

	// 提取SVG代码
	svgContent := openaiResp.Choices[0].Message.Content
	svgCode := s.extractSVGCode(svgContent)

	if svgCode == "" {
		log.Printf("[OPENAI] No valid SVG found in response")
		return nil, fmt.Errorf("no valid SVG generated")
	}

	// 生成临时SVG文件URL (实际应用中可能需要保存到文件服务)
	imageID := generateOpenAIImageID()
	svgURL := s.createSVGDataURL(svgCode)

	log.Printf("[OPENAI] Successfully generated SVG - ID: %s", imageID)

	return &types.ImageResponse{
		ID:             imageID,
		Prompt:         req.Prompt,
		NegativePrompt: req.NegativePrompt,
		Style:          req.Style,
		SVGURL:         svgURL,
		PNGURL:         svgURL, // OpenAI生成的是SVG代码，两个URL相同
		Width:          1024,   // 默认尺寸
		Height:         1024,
		CreatedAt:      time.Now(),
		Provider:       types.ProviderOpenAI,
	}, nil
}

// buildSVGPrompt 构建用于生成SVG的提示词
func (s *OpenAIService) buildSVGPrompt(prompt, style, negativePrompt string) string {
	var promptBuilder strings.Builder

	promptBuilder.WriteString("Create a high-quality SVG illustration of: ")
	promptBuilder.WriteString(prompt)

	if style != "" {
		promptBuilder.WriteString(fmt.Sprintf("\n\nArtistic style and visual requirements: %s", style))
	}

	if negativePrompt != "" {
		promptBuilder.WriteString(fmt.Sprintf("\n\nIMPORTANT - Do NOT include these elements: %s", negativePrompt))
	}

	promptBuilder.WriteString(`

Technical Requirements:
• Use viewBox="0 0 1024 1024" for consistent sizing
• Ensure the SVG is completely self-contained and valid
• Use semantic and descriptive element IDs (e.g., id="main-character", id="background-sky")
• Organize elements in logical groups using <g> tags
• Use appropriate colors that match the subject and style
• Include gradients, shadows, or other effects when they enhance the design
• Ensure the illustration is centered and well-composed within the viewBox
• Make it scalable and crisp at any resolution

Visual Quality Standards:
• Pay attention to proper proportions and anatomy
• Use appropriate line weights and stroke styles
• Include relevant details that make the illustration engaging
• Ensure good contrast and readability
• Follow the specified artistic style consistently
• Create depth and dimension through layering and visual effects

Output Format:
Return ONLY the complete SVG code starting with <svg> and ending with </svg>. No explanations, no code blocks, no additional text.
In SVG, do not use elements like rect; all drawing should be implemented via the path element.

`)

	return promptBuilder.String()
}

// extractSVGCode 从响应中提取SVG代码
func (s *OpenAIService) extractSVGCode(content string) string {
	// 查找SVG标签
	svgRegex := regexp.MustCompile(`(?s)<svg[^>]*>.*?</svg>`)
	matches := svgRegex.FindString(content)

	if matches != "" {
		return matches
	}

	// 如果没找到完整的SVG，尝试查找SVG代码块
	codeBlockRegex := regexp.MustCompile("(?s)```(?:svg|xml)?\n?(.*?)\n?```")
	codeMatches := codeBlockRegex.FindStringSubmatch(content)

	if len(codeMatches) > 1 {
		svgCode := strings.TrimSpace(codeMatches[1])
		if strings.Contains(svgCode, "<svg") {
			return svgCode
		}
	}

	// 最后尝试直接查找是否整个内容就是SVG
	if strings.Contains(content, "<svg") && strings.Contains(content, "</svg>") {
		return strings.TrimSpace(content)
	}

	return ""
}

// createSVGDataURL 创建SVG的Data URL
func (s *OpenAIService) createSVGDataURL(svgCode string) string {
	// 对于演示，我们返回一个data URL
	// 实际生产环境中，你可能想要保存到文件服务并返回真实URL
	return fmt.Sprintf("data:image/svg+xml;base64,%s",
		encodeSVGToBase64(svgCode))
}

// generateOpenAIImageID 生成OpenAI图片ID
func generateOpenAIImageID() string {
	return fmt.Sprintf("openai_%d", time.Now().UnixNano())
}

// encodeSVGToBase64 将SVG编码为Base64
func encodeSVGToBase64(svgCode string) string {
	return base64.StdEncoding.EncodeToString([]byte(svgCode))
}

// truncateString 截断字符串用于日志
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// getModelFromConfig 从配置或请求中获取模型名称
func getModelFromConfig(requestModel string) string {
	if requestModel != "" {
		return requestModel
	}
	if config.AppConfig != nil {
		return config.AppConfig.Providers.OpenAI.DefaultModel
	}
	return "gpt-4" // fallback
}

// getMaxTokensFromConfig 从配置中获取最大token数
func getMaxTokensFromConfig() int {
	if config.AppConfig != nil {
		return config.AppConfig.Providers.OpenAI.MaxTokens
	}
	return 4000 // fallback
}

// getTemperatureFromConfig 从配置中获取温度参数
func getTemperatureFromConfig() float64 {
	if config.AppConfig != nil {
		return config.AppConfig.Providers.OpenAI.Temperature
	}
	return 0.7 // fallback
}
