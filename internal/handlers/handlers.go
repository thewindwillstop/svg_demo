package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"svg-generator/internal/service"
	"svg-generator/internal/types"
	"svg-generator/pkg/utils"
)

// UnifiedImageHandler 统一的图像元数据处理器 - 根据请求中的provider参数选择模型
func UnifiedImageHandler(serviceManager *service.ServiceManager, translateService utils.TranslateService) http.HandlerFunc {
	return unifiedGenerateHandler(serviceManager, translateService, false)
}

// UnifiedSVGHandler 统一的SVG处理器 - 根据请求中的provider参数选择模型
func UnifiedSVGHandler(serviceManager *service.ServiceManager, translateService utils.TranslateService) http.HandlerFunc {
	return unifiedGenerateHandler(serviceManager, translateService, true)
}

// unifiedGenerateHandler 统一的图像生成处理器
func unifiedGenerateHandler(serviceManager *service.ServiceManager, translateService utils.TranslateService, directSVG bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Unified handler request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)

		if r.Method != http.MethodPost {
			log.Printf("Method not allowed: %s", r.Method)
			utils.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only POST is allowed", nil)
			return
		}

		var req types.GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("JSON decode error: %v", err)
			utils.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid request body", err.Error())
			return
		}

		// 验证provider参数
		if req.Provider == "" {
			req.Provider = types.ProviderSVGIO // 默认使用SVG.IO
		}

		// 验证provider是否有效
		validProviders := []types.Provider{types.ProviderSVGIO, types.ProviderRecraft, types.ProviderOpenAI}
		isValidProvider := false
		for _, p := range validProviders {
			if req.Provider == p {
				isValidProvider = true
				break
			}
		}

		if !isValidProvider {
			log.Printf("Invalid provider: %s", req.Provider)
			utils.WriteError(w, http.StatusBadRequest, "invalid_provider", "provider must be one of: svgio, recraft, openai", nil)
			return
		}

		providerName := string(req.Provider)
		log.Printf("[%s] Request parsed - prompt: %q, style: %q, provider: %s", providerName, req.Prompt, req.Style, req.Provider)

		if len(req.Prompt) < 3 {
			log.Printf("[%s] Prompt too short: %d chars", providerName, len(req.Prompt))
			utils.WriteError(w, http.StatusBadRequest, "invalid_argument", "prompt must be at least 3 characters", nil)
			return
		}

		// 翻译处理 (仅对 SVG.IO 提供商进行翻译，Recraft 和 OpenAI 支持中文)
		originalPrompt := req.Prompt
		translatedPrompt := req.Prompt
		wasTranslated := false

		if !req.SkipTranslate && translateService != nil && req.Provider == types.ProviderSVGIO {
			translateCtx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
			defer cancel()
			translated, err := translateService.Translate(translateCtx, req.Prompt)
			if err != nil {
				log.Printf("[%s] Translation failed: %v", providerName, err)
				// 翻译失败时使用原文继续处理，不中断流程
			} else if translated != req.Prompt {
				translatedPrompt = translated
				wasTranslated = true
				log.Printf("[%s] Prompt translated: %q -> %q", providerName, originalPrompt, translatedPrompt)
			}
		}

		// 使用翻译后的提示词
		req.Prompt = translatedPrompt

		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()

		log.Printf("[%s] Calling upstream API...", providerName)
		img, err := serviceManager.GenerateImage(ctx, req)
		if err != nil {
			log.Printf("[%s] Upstream generation failed: %v", providerName, err)
			status := http.StatusBadGateway
			if errors.Is(err, context.DeadlineExceeded) {
				status = http.StatusGatewayTimeout
			}
			utils.WriteError(w, status, "upstream_error", "failed to generate image", err.Error())
			return
		}

		log.Printf("[%s] Generation successful - ID: %s, SVG URL: %s", providerName, img.ID, img.SVGURL)

		if directSVG {
			// 直接返回 SVG 文件
			log.Printf("[%s] Processing SVG content from: %s", providerName, img.SVGURL)

			var svgBytes []byte
			var err error

			// 检查是否是data URL
			if strings.HasPrefix(img.SVGURL, "data:") {
				// 处理data URL
				svgBytes, err = parseDataURL(img.SVGURL)
				if err != nil {
					log.Printf("[%s] Failed to parse data URL: %v", providerName, err)
					utils.WriteError(w, http.StatusInternalServerError, "parse_error", "failed to parse data URL", err.Error())
					return
				}
				log.Printf("[%s] Parsed data URL - size: %d bytes", providerName, len(svgBytes))
			} else {
				// 处理HTTP/HTTPS URL
				svgBytes, err = utils.DownloadFile(ctx, img.SVGURL)
				if err != nil {
					log.Printf("[%s] Download failed: %v", providerName, err)
					utils.WriteError(w, http.StatusBadGateway, "download_error", "failed to download generated svg", err.Error())
					return
				}
				log.Printf("[%s] Download successful - size: %d bytes", providerName, len(svgBytes))
			}

			w.Header().Set("Content-Type", "image/svg+xml")
			w.Header().Set("Content-Disposition", "attachment; filename=\""+img.ID+".svg\"")
			// 可以附带元信息 header
			w.Header().Set("X-Image-Id", img.ID)
			w.Header().Set("X-Image-Width", strconv.Itoa(img.Width))
			w.Header().Set("X-Image-Height", strconv.Itoa(img.Height))
			w.Header().Set("X-Provider", string(req.Provider))
			// 添加翻译信息到响应头
			if wasTranslated {
				w.Header().Set("X-Original-Prompt", originalPrompt)
				w.Header().Set("X-Translated-Prompt", translatedPrompt)
				w.Header().Set("X-Was-Translated", "true")
			}
			utils.SetCORSHeaders(w)
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write(svgBytes); err != nil {
				log.Printf("[%s] Write response error: %v", providerName, err)
			} else {
				log.Printf("[%s] Response sent successfully", providerName)
			}
		} else {
			// 返回 JSON 元数据
			response := types.ImageResponse{
				ID:       img.ID,
				SVGURL:   img.SVGURL,
				Width:    img.Width,
				Height:   img.Height,
				Provider: req.Provider,
			}

			// 添加翻译信息
			if wasTranslated {
				response.OriginalPrompt = originalPrompt
				response.TranslatedPrompt = translatedPrompt
				response.WasTranslated = wasTranslated
			}

			w.Header().Set("Content-Type", "application/json")
			utils.SetCORSHeaders(w)
			w.WriteHeader(http.StatusOK)

			if err := json.NewEncoder(w).Encode(response); err != nil {
				log.Printf("[%s] JSON encode error: %v", providerName, err)
			} else {
				log.Printf("[%s] JSON response sent successfully", providerName)
			}
		}
	}
}

// HealthHandler 健康检查处理器
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.SetCORSHeaders(w)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	}
}

// parseDataURL 解析data URL并返回解码后的数据
func parseDataURL(dataURL string) ([]byte, error) {
	// data URL格式: data:[<mediatype>][;base64],<data>
	// 例如: data:image/svg+xml;base64,<base64-encoded-data>

	if !strings.HasPrefix(dataURL, "data:") {
		return nil, errors.New("invalid data URL: missing data: prefix")
	}

	// 去掉"data:"前缀
	dataURL = dataURL[5:]

	// 查找逗号分隔符
	commaIndex := strings.Index(dataURL, ",")
	if commaIndex == -1 {
		return nil, errors.New("invalid data URL: missing comma separator")
	}

	// 获取媒体类型和编码信息
	mediaType := dataURL[:commaIndex]
	data := dataURL[commaIndex+1:]

	// 检查是否是base64编码
	if strings.Contains(mediaType, "base64") {
		// base64解码
		decoded, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			return nil, errors.New("failed to decode base64 data: " + err.Error())
		}
		return decoded, nil
	} else {
		// 非base64编码，直接返回字符串的字节
		return []byte(data), nil
	}
}

// CORSPreflight 处理 CORS 预检请求
func CORSPreflight() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.SetCORSHeaders(w)
		w.WriteHeader(http.StatusOK)
	}
}
