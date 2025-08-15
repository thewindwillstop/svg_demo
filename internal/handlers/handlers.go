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

	"miniSvg/internal/client"
	"miniSvg/internal/translate"
	"miniSvg/internal/types"
	"miniSvg/internal/upstream"
	"miniSvg/pkg/utils"
)

// SVGHandler SVG生成和下载处理器 (使用 SVG.IO)
func SVGHandler(serviceManager *upstream.ServiceManager, translateService translate.Service) http.HandlerFunc {
	return generateHandler(serviceManager, translateService, types.ProviderSVGIO, true)
}

// RecraftSVGHandler Recraft SVG生成和下载处理器
func RecraftSVGHandler(serviceManager *upstream.ServiceManager, translateService translate.Service) http.HandlerFunc {
	return generateHandler(serviceManager, translateService, types.ProviderRecraft, true)
}

// ImageHandler JSON 元数据接口处理器 (使用 SVG.IO)
func ImageHandler(serviceManager *upstream.ServiceManager, translateService translate.Service) http.HandlerFunc {
	return generateHandler(serviceManager, translateService, types.ProviderSVGIO, false)
}

// RecraftImageHandler Recraft JSON 元数据接口处理器
func RecraftImageHandler(serviceManager *upstream.ServiceManager, translateService translate.Service) http.HandlerFunc {
	return generateHandler(serviceManager, translateService, types.ProviderRecraft, false)
}

// ClaudeSVGHandler Claude SVG生成和下载处理器
func ClaudeSVGHandler(serviceManager *upstream.ServiceManager, translateService translate.Service) http.HandlerFunc {
	return generateHandler(serviceManager, translateService, types.ProviderClaude, true)
}

// ClaudeImageHandler Claude JSON 元数据接口处理器
func ClaudeImageHandler(serviceManager *upstream.ServiceManager, translateService translate.Service) http.HandlerFunc {
	return generateHandler(serviceManager, translateService, types.ProviderClaude, false)
}

// generateHandler 通用图像生成处理器
func generateHandler(serviceManager *upstream.ServiceManager, translateService translate.Service, provider types.Provider, directSVG bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		providerName := string(provider)
		log.Printf("[%s] Request from %s: %s %s", providerName, r.RemoteAddr, r.Method, r.URL.Path)

		if r.Method != http.MethodPost {
			log.Printf("[%s] Method not allowed: %s", providerName, r.Method)
			utils.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only POST is allowed", nil)
			return
		}

		var req types.GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[%s] JSON decode error: %v", providerName, err)
			utils.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid request body", err.Error())
			return
		}

		// 强制设置提供商
		req.Provider = provider

		log.Printf("[%s] Request parsed - prompt: %q, style: %q, provider: %s", providerName, req.Prompt, req.Style, req.Provider)

		if len(req.Prompt) < 3 {
			log.Printf("[%s] Prompt too short: %d chars", providerName, len(req.Prompt))
			utils.WriteError(w, http.StatusBadRequest, "invalid_argument", "prompt must be at least 3 characters", nil)
			return
		}

		// 翻译处理 (仅对 SVG.IO 提供商进行翻译，Recraft 和 Claude 支持中文)
		originalPrompt := req.Prompt
		translatedPrompt := req.Prompt
		wasTranslated := false

		if !req.SkipTranslate && translateService != nil && provider == types.ProviderSVGIO {
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
				svgBytes, err = client.DownloadFile(ctx, img.SVGURL)
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
			w.Header().Set("X-Provider", string(provider))
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
				Provider: provider,
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
