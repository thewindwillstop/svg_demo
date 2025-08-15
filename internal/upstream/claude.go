package upstream

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"miniSvg/internal/client"
	"miniSvg/internal/types"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// ClaudeService 实现 Claude API 调用
type ClaudeService struct {
	apiKey  string
	baseURL string
}

// NewClaudeService 创建 Claude 服务实例
func NewClaudeService(apiKey, baseURL string) *ClaudeService {
	if baseURL == "" {
		baseURL = "https://api.qnaigc.com/v1/"
	}
	return &ClaudeService{
		apiKey:  apiKey,
		baseURL: strings.TrimSuffix(baseURL, "/"),
	}
}

// GenerateImage 使用 Claude 生成 SVG 代码
func (s *ClaudeService) GenerateImage(ctx context.Context, req types.GenerateRequest) (*types.ImageResponse, error) {
	log.Printf("[CLAUDE] Starting SVG generation request...")

	// 构建 Claude 提示词
	prompt := s.buildSVGPrompt(req.Prompt, req.Style, req.NegativePrompt)

	// 构建 Claude API 请求
	claudeReq := types.ClaudeGenerateReq{
		Model:       "claude-4.0-sonnet",
		MaxTokens:   4000,
		Temperature: 0.7,
		System: `You are a world-class SVG graphics designer and vector artist with expertise in creating stunning, precise, and semantically meaningful SVG illustrations. Your specialties include:

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
		Messages: []types.ClaudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	body, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := s.baseURL + "/chat/completions"
	log.Printf("[CLAUDE] Sending request to %s", url)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.HTTPClient.Do(httpReq)
	if err != nil {
		log.Printf("[CLAUDE] HTTP request failed: %v", err)
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[CLAUDE] Received response with status: %s", resp.Status)

	if resp.StatusCode >= 300 {
		var errResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		log.Printf("[CLAUDE] Error response body: %+v", errResp)
		return nil, fmt.Errorf("claude API error: %s", resp.Status)
	}

	// 读取原始响应内容进行调试
	bodyBytes := make([]byte, 0)
	if body, err := io.ReadAll(resp.Body); err == nil {
		bodyBytes = body
		log.Printf("[CLAUDE] Raw response body: %s", string(body))
	}

	// 重新创建Reader用于JSON解码
	bodyReader := bytes.NewReader(bodyBytes)

	var claudeResp types.ClaudeGenerateResp
	if err := json.NewDecoder(bodyReader).Decode(&claudeResp); err != nil {
		log.Printf("[CLAUDE] Failed to decode response: %v", err)

		// 尝试解析为通用格式
		bodyReader.Seek(0, 0)
		return s.tryParseGenericFormat(bodyReader, &req)
	}

	// 添加调试信息
	log.Printf("[CLAUDE] Response structure: ID=%s, Type=%s, Role=%s, Content length=%d",
		claudeResp.ID, claudeResp.Type, claudeResp.Role, len(claudeResp.Content))

	// 如果有内容，打印第一个内容的类型和前100个字符
	if len(claudeResp.Content) > 0 {
		firstContent := claudeResp.Content[0]
		log.Printf("[CLAUDE] First content: Type=%s, Text prefix=%s",
			firstContent.Type, truncateString(firstContent.Text, 100))
	}

	if len(claudeResp.Content) == 0 {
		log.Printf("[CLAUDE] Content array is empty, response: %+v", claudeResp)
		// 重置Reader并尝试通用格式解析
		bodyReader.Seek(0, 0)
		return s.tryParseGenericFormat(bodyReader, &req)
	}

	// 提取SVG代码
	svgContent := claudeResp.Content[0].Text
	svgCode := s.extractSVGCode(svgContent)

	if svgCode == "" {
		log.Printf("[CLAUDE] No valid SVG found in response")
		return nil, fmt.Errorf("no valid SVG generated")
	}

	// 生成临时SVG文件URL (实际应用中可能需要保存到文件服务)
	imageID := generateClaudeImageID()
	svgURL := s.createSVGDataURL(svgCode)

	log.Printf("[CLAUDE] Successfully generated SVG - ID: %s", imageID)

	return &types.ImageResponse{
		ID:             imageID,
		Prompt:         req.Prompt,
		NegativePrompt: req.NegativePrompt,
		Style:          req.Style,
		SVGURL:         svgURL,
		PNGURL:         svgURL, // Claude生成的是SVG代码，两个URL相同
		Width:          1024,   // 默认尺寸
		Height:         1024,
		CreatedAt:      time.Now(),
		Provider:       types.ProviderClaude,
	}, nil
}

// buildSVGPrompt 构建用于生成SVG的提示词
func (s *ClaudeService) buildSVGPrompt(prompt, style, negativePrompt string) string {
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
Return ONLY the complete SVG code starting with <svg> and ending with </svg>. No explanations, no code blocks, no additional text.`)

	return promptBuilder.String()
}

// extractSVGCode 从响应中提取SVG代码
func (s *ClaudeService) extractSVGCode(content string) string {
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
func (s *ClaudeService) createSVGDataURL(svgCode string) string {
	// 对于演示，我们返回一个data URL
	// 实际生产环境中，你可能想要保存到文件服务并返回真实URL
	return fmt.Sprintf("data:image/svg+xml;base64,%s",
		encodeSVGToBase64(svgCode))
}

// generateClaudeImageID 生成Claude图片ID
func generateClaudeImageID() string {
	return fmt.Sprintf("claude_%d", time.Now().UnixNano())
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

// tryParseGenericFormat 尝试解析通用格式响应
func (s *ClaudeService) tryParseGenericFormat(bodyReader io.Reader, req *types.GenerateRequest) (*types.ImageResponse, error) {
	// 尝试解析为通用的map格式
	var genericResp map[string]interface{}
	if err := json.NewDecoder(bodyReader).Decode(&genericResp); err != nil {
		return nil, fmt.Errorf("failed to decode generic response: %w", err)
	}

	log.Printf("[CLAUDE] Generic response structure: %+v", genericResp)

	// 尝试从不同的字段提取文本内容
	var textContent string

	// 检查常见的文本字段
	if choices, ok := genericResp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					textContent = content
				}
			}
		}
	}

	// 如果没有找到choices，尝试其他字段
	if textContent == "" {
		if content, ok := genericResp["content"].(string); ok {
			textContent = content
		} else if text, ok := genericResp["text"].(string); ok {
			textContent = text
		}
	}

	if textContent == "" {
		return nil, fmt.Errorf("no text content found in response")
	}

	log.Printf("[CLAUDE] Extracted text content: %s", truncateString(textContent, 200))

	// 从文本中提取SVG代码
	svgCode := s.extractSVGCode(textContent)
	if svgCode == "" {
		return nil, fmt.Errorf("no SVG code found in response")
	}

	return &types.ImageResponse{
		ID:        generateClaudeImageID(),
		Prompt:    req.Prompt,
		SVGURL:    fmt.Sprintf("data:image/svg+xml;base64,%s", encodeSVGToBase64(svgCode)),
		Width:     512,
		Height:    512,
		CreatedAt: time.Now(),
	}, nil
}
