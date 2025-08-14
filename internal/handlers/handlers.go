package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"miniSvg/internal/client"
	"miniSvg/internal/translate"
	"miniSvg/internal/types"
	"miniSvg/internal/upstream"
	"miniSvg/pkg/utils"
	"net/http"
	"strconv"
	"time"
)

// SVGHandler SVG生成和下载处理器
func SVGHandler(apiKey string, translateService translate.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[SVG] Request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)

		if r.Method != http.MethodPost {
			log.Printf("[SVG] Method not allowed: %s", r.Method)
			utils.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only POST is allowed", nil)
			return
		}
		var req types.GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[SVG] JSON decode error: %v", err)
			utils.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid request body", err.Error())
			return
		}
		log.Printf("[SVG] Request parsed - prompt: %q, style: %q", req.Prompt, req.Style)

		if len(req.Prompt) < 3 {
			log.Printf("[SVG] Prompt too short: %d chars", len(req.Prompt))
			utils.WriteError(w, http.StatusBadRequest, "invalid_argument", "prompt must be at least 3 characters", nil)
			return
		}

		// 翻译处理
		originalPrompt := req.Prompt
		translatedPrompt := req.Prompt
		wasTranslated := false

		if !req.SkipTranslate && translateService != nil {
			translateCtx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
			defer cancel()

			translated, err := translateService.Translate(translateCtx, req.Prompt)
			if err != nil {
				log.Printf("[SVG] Translation failed: %v", err)
				// 翻译失败时使用原文继续处理，不中断流程
			} else if translated != req.Prompt {
				translatedPrompt = translated
				wasTranslated = true
				log.Printf("[SVG] Prompt translated: %q -> %q", originalPrompt, translatedPrompt)
			}
		}

		// 使用翻译后的提示词
		req.Prompt = translatedPrompt

		ctx, cancel := context.WithTimeout(r.Context(), 28*time.Second)
		defer cancel()

		log.Printf("[SVG] Calling upstream API...")
		img, err := upstream.CallSVGIOGenerate(ctx, apiKey, req)
		if err != nil {
			log.Printf("[SVG] Upstream generation failed: %v", err)
			status := http.StatusBadGateway
			if errors.Is(err, context.DeadlineExceeded) {
				status = http.StatusGatewayTimeout
			}
			utils.WriteError(w, status, "upstream_error", "failed to generate image", err.Error())
			return
		}
		log.Printf("[SVG] Generation successful - ID: %s, SVG URL: %s", img.ID, img.SVGURL)

		// 下载 SVG 内容
		log.Printf("[SVG] Downloading SVG content from: %s", img.SVGURL)
		svgBytes, err := client.DownloadFile(ctx, img.SVGURL)
		if err != nil {
			log.Printf("[SVG] Download failed: %v", err)
			utils.WriteError(w, http.StatusBadGateway, "download_error", "failed to download generated svg", err.Error())
			return
		}
		log.Printf("[SVG] Download successful - size: %d bytes", len(svgBytes))

		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+img.ID+".svg\"")
		// 可以附带元信息 header
		w.Header().Set("X-Image-Id", img.ID)
		w.Header().Set("X-Image-Width", strconv.Itoa(img.Width))
		w.Header().Set("X-Image-Height", strconv.Itoa(img.Height))
		// 添加翻译信息到响应头
		if wasTranslated {
			w.Header().Set("X-Original-Prompt", originalPrompt)
			w.Header().Set("X-Translated-Prompt", translatedPrompt)
			w.Header().Set("X-Was-Translated", "true")
		}
		utils.SetCORSHeaders(w)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(svgBytes); err != nil {
			log.Printf("[SVG] Write response error: %v", err)
		} else {
			log.Printf("[SVG] Response sent successfully")
		}
	}
}

// ImageHandler JSON 元数据接口处理器
func ImageHandler(apiKey string, translateService translate.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[IMG] Request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)

		var req types.GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[IMG] JSON decode error: %v", err)
			utils.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid request body", err.Error())
			return
		}
		log.Printf("[IMG] Request parsed - prompt: %q, style: %q", req.Prompt, req.Style)

		if len(req.Prompt) < 3 {
			log.Printf("[IMG] Prompt too short: %d chars", len(req.Prompt))
			utils.WriteError(w, http.StatusBadRequest, "invalid_argument", "prompt must be at least 3 characters", nil)
			return
		}

		// 翻译处理
		originalPrompt := req.Prompt
		translatedPrompt := req.Prompt
		wasTranslated := false

		if !req.SkipTranslate && translateService != nil {
			translateCtx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
			defer cancel()

			translated, err := translateService.Translate(translateCtx, req.Prompt)
			if err != nil {
				log.Printf("[IMG] Translation failed: %v", err)
				// 翻译失败时使用原文继续处理，不中断流程
			} else if translated != req.Prompt {
				translatedPrompt = translated
				wasTranslated = true
				log.Printf("[IMG] Prompt translated: %q -> %q", originalPrompt, translatedPrompt)
			}
		}

		// 使用翻译后的提示词
		req.Prompt = translatedPrompt

		//接口超时时间
		ctx, cancel := context.WithTimeout(r.Context(), 28*time.Second)
		defer cancel()

		log.Printf("[IMG] Calling upstream API...")
		img, err := upstream.CallSVGIOGenerate(ctx, apiKey, req)
		if err != nil {
			log.Printf("[IMG] Upstream generation failed: %v", err)
			status := http.StatusBadGateway
			if errors.Is(err, context.DeadlineExceeded) {
				status = http.StatusGatewayTimeout
			}
			utils.WriteError(w, status, "upstream_error", "failed to generate image", err.Error())
			return
		}
		log.Printf("[IMG] Generation successful - ID: %s, PNG URL: %s", img.ID, img.PNGURL)

		// 在响应中包含翻译信息
		img.OriginalPrompt = originalPrompt
		img.TranslatedPrompt = translatedPrompt
		img.WasTranslated = wasTranslated

		utils.WriteJSON(w, http.StatusOK, img)
		log.Printf("[IMG] Response sent successfully")
	}
}

// PingHandler Ping 健康检查处理器
func PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			utils.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only GET/HEAD allowed", nil)
			return
		}
		resp := map[string]any{
			"status": "ok",
			"time":   time.Now().UTC().Format(time.RFC3339Nano),
			"uptime": time.Since(start).String(),
		}
		// HEAD 不写 body
		if r.Method == http.MethodHead {
			utils.SetCORSHeaders(w)
			w.WriteHeader(http.StatusOK)
			return
		}
		utils.WriteJSON(w, http.StatusOK, resp)
	}
}

// DownloadSVGHandler SVG URL 下载处理器
func DownloadSVGHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only GET is allowed", nil)
			return
		}

		svgURL := r.URL.Query().Get("url")
		if svgURL == "" {
			utils.WriteError(w, http.StatusBadRequest, "missing_parameter", "url parameter is required", nil)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, svgURL, nil)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, "invalid_url", "invalid SVG URL", err.Error())
			return
		}

		resp, err := client.HTTPClient.Do(req)
		if err != nil {
			utils.WriteError(w, http.StatusBadGateway, "download_error", "failed to download SVG", err.Error())
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 300 {
			utils.WriteError(w, http.StatusBadGateway, "upstream_error", "failed to fetch SVG from upstream", nil)
			return
		}

		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Content-Disposition", "attachment; filename=\"image.svg\"")
		utils.SetCORSHeaders(w)
		w.WriteHeader(http.StatusOK)

		svgBytes, err := client.DownloadFile(ctx, svgURL)
		if err != nil {
			log.Printf("failed to download SVG: %v", err)
		} else {
			if _, err := w.Write(svgBytes); err != nil {
				log.Printf("failed to write SVG response: %v", err)
			}
		}
	}
}
