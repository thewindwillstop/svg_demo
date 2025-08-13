package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

// SVG 生成和下载处理器
func handleSVGGeneration(apiKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[SVG] Request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)

		if r.Method != http.MethodPost {
			log.Printf("[SVG] Method not allowed: %s", r.Method)
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only POST is allowed", nil)
			return
		}
		var req GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[SVG] JSON decode error: %v", err)
			writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body", err.Error())
			return
		}
		log.Printf("[SVG] Request parsed - prompt: %q, style: %q", req.Prompt, req.Style)

		if len(req.Prompt) < 3 {
			log.Printf("[SVG] Prompt too short: %d chars", len(req.Prompt))
			writeError(w, http.StatusBadRequest, "invalid_argument", "prompt must be at least 3 characters", nil)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 50*time.Second)
		defer cancel()

		log.Printf("[SVG] Calling upstream API...")
		img, err := callSVGIOGenerate(ctx, apiKey, req)
		if err != nil {
			log.Printf("[SVG] Upstream generation failed: %v", err)
			status := http.StatusBadGateway
			if errors.Is(err, context.DeadlineExceeded) {
				status = http.StatusGatewayTimeout
			}
			writeError(w, status, "upstream_error", "failed to generate image", err.Error())
			return
		}
		log.Printf("[SVG] Generation successful - ID: %s, SVG URL: %s", img.ID, img.SVGURL)

		// 下载 SVG 内容
		log.Printf("[SVG] Downloading SVG content from: %s", img.SVGURL)
		svgBytes, err := downloadFile(ctx, img.SVGURL)
		if err != nil {
			log.Printf("[SVG] Download failed: %v", err)
			writeError(w, http.StatusBadGateway, "download_error", "failed to download generated svg", err.Error())
			return
		}
		log.Printf("[SVG] Download successful - size: %d bytes", len(svgBytes))

		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+img.ID+".svg\"")
		// 可以附带元信息 header
		w.Header().Set("X-Image-Id", img.ID)
		w.Header().Set("X-Image-Width", toString(img.Width))
		w.Header().Set("X-Image-Height", toString(img.Height))
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(svgBytes); err != nil {
			log.Printf("[SVG] Write response error: %v", err)
		} else {
			log.Printf("[SVG] Response sent successfully")
		}
	}
}

// JSON 元数据接口处理器
func handleImageGeneration(apiKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[IMG] Request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)

		var req GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[IMG] JSON decode error: %v", err)
			writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body", err.Error())
			return
		}
		log.Printf("[IMG] Request parsed - prompt: %q, style: %q", req.Prompt, req.Style)

		if len(req.Prompt) < 3 {
			log.Printf("[IMG] Prompt too short: %d chars", len(req.Prompt))
			writeError(w, http.StatusBadRequest, "invalid_argument", "prompt must be at least 3 characters", nil)
			return
		}

		//接口超时时间
		ctx, cancel := context.WithTimeout(r.Context(), 50*time.Second)
		defer cancel()

		log.Printf("[IMG] Calling upstream API...")
		img, err := callSVGIOGenerate(ctx, apiKey, req)
		if err != nil {
			log.Printf("[IMG] Upstream generation failed: %v", err)
			status := http.StatusBadGateway
			if errors.Is(err, context.DeadlineExceeded) {
				status = http.StatusGatewayTimeout
			}
			writeError(w, status, "upstream_error", "failed to generate image", err.Error())
			return
		}
		log.Printf("[IMG] Generation successful - ID: %s, PNG URL: %s", img.ID, img.PNGURL)

		writeJSON(w, http.StatusOK, img)
		log.Printf("[IMG] Response sent successfully")
	}
}

// Ping 健康检查处理器
func handlePing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only GET/HEAD allowed", nil)
			return
		}
		resp := map[string]any{
			"status": "ok",
			"time":   time.Now().UTC().Format(time.RFC3339Nano),
			"uptime": time.Since(start).String(),
		}
		// HEAD 不写 body
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

// SVG URL 下载处理器
func handleDownloadSVG() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only GET is allowed", nil)
			return
		}

		svgURL := r.URL.Query().Get("url")
		if svgURL == "" {
			writeError(w, http.StatusBadRequest, "missing_parameter", "url parameter is required", nil)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, svgURL, nil)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_url", "invalid SVG URL", err.Error())
			return
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			writeError(w, http.StatusBadGateway, "download_error", "failed to download SVG", err.Error())
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 300 {
			writeError(w, http.StatusBadGateway, "upstream_error", "failed to fetch SVG from upstream", nil)
			return
		}

		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Content-Disposition", "attachment; filename=\"image.svg\"")
		w.WriteHeader(http.StatusOK)

		_, err = w.Write(readAllBytes(resp.Body))
		if err != nil {
			log.Printf("failed to write SVG response: %v", err)
		}
	}
}
