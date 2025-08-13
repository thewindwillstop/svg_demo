// main.go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	upstreamBaseURL = "https://api.svg.io"
	generatePath    = "/v1/generate-image"
	getImagePath    = "/v1/get-image/" // + {imageId}
)

func downloadSVG(w http.ResponseWriter, r *http.Request) {
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

func readAllBytes(r interface{}) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.(interface{ Read([]byte) (int, error) }))
	return buf.Bytes()
}

type GenerateRequest struct {
	Prompt         string `json:"prompt"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
	Style          string `json:"style,omitempty"`
	// 可选：前端区分用途（例如 png 或 svg_inline），当前普通 /v1/images 忽略该字段
	Format string `json:"format,omitempty"`
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
}

type upstreamGenerateReq struct {
	Prompt           string `json:"prompt"`
	NegativePrompt   string `json:"negativePrompt"`
	Style            string `json:"style,omitempty"`
	InitialImage     any    `json:"initialImage"`     // 必须传 null，否则会被序列化为 ""
	InitialImageType any    `json:"initialImageType"` // 同上
}

type upstreamGenerateItem struct {
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

type upstreamGenerateResp struct {
	Success bool                   `json:"success"`
	Data    []upstreamGenerateItem `json:"data"`
}

type errorResp struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

func main() {
	log.Printf("Starting SVG image generation service...")

	_ = godotenv.Load(".env")
	apiKey := os.Getenv("SVGIO_API_KEY")
	if apiKey == "" {
		log.Fatal("missing SVGIO_API_KEY")
	}
	log.Printf("API key loaded successfully (length: %d)", len(apiKey))

	mux := http.NewServeMux()

	// 新增: 直接生成并返回 SVG 文件
	mux.HandleFunc("/v1/images/svg", func(w http.ResponseWriter, r *http.Request) {
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

		ctx, cancel := context.WithTimeout(r.Context(), 28*time.Second)
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
	})
	 mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
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
    })

	// 原 JSON 接口（返回元数据 & URL）
	mux.HandleFunc("/v1/images", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[IMG] Request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)

		if r.Method != http.MethodPost {
			log.Printf("[IMG] Method not allowed: %s", r.Method)
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only POST is allowed", nil)
			return
		}

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

		ctx, cancel := context.WithTimeout(r.Context(), 28*time.Second)
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
	})

	// 已有下载转发接口（外部 URL -> 文件）

	addr := "0.0.0.0:8080"
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, withCommonHeaders(mux)); err != nil {
		log.Fatal(err)
	}
}

func callSVGIOGenerate(ctx context.Context, apiKey string, req GenerateRequest) (*ImageResponse, error) {
	log.Printf("[UPSTREAM] Starting generation request...")

	upReq := upstreamGenerateReq{
		Prompt:           req.Prompt,
		NegativePrompt:   req.NegativePrompt,
		Style:            req.Style,
		InitialImage:     nil,
		InitialImageType: nil,
	}

	body, _ := json.Marshal(upReq)
	log.Printf("[UPSTREAM] Sending request to %s with payload size: %d bytes", upstreamBaseURL+generatePath, len(body))

	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, upstreamBaseURL+generatePath, bytesReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		log.Printf("[UPSTREAM] HTTP request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("[UPSTREAM] Received response with status: %s", resp.Status)

	if resp.StatusCode >= 300 {
		var raw any
		_ = json.NewDecoder(resp.Body).Decode(&raw)
		log.Printf("[UPSTREAM] Error response body: %+v", raw)
		return nil, errors.New("upstream status: " + resp.Status)
	}

	var upResp upstreamGenerateResp
	if err := json.NewDecoder(resp.Body).Decode(&upResp); err != nil {
		log.Printf("[UPSTREAM] Failed to decode response: %v", err)
		return nil, err
	}
	if !upResp.Success || len(upResp.Data) == 0 {
		log.Printf("[UPSTREAM] Invalid response: success=%v, data_count=%d", upResp.Success, len(upResp.Data))
		return nil, errors.New("upstream no data")
	}
	it := upResp.Data[0]
	log.Printf("[UPSTREAM] Successfully parsed response - ID: %s, SVG: %s, PNG: %s", it.ID, it.SVGURL, it.PNGURL)

	createdAt, _ := time.Parse(time.RFC3339, it.CreatedAt)
	return &ImageResponse{
		ID:             it.ID,
		Prompt:         it.Prompt,
		NegativePrompt: it.NegativePrompt,
		Style:          it.Style,
		SVGURL:         it.SVGURL,
		PNGURL:         it.PNGURL,
		Width:          it.Width,
		Height:         it.Height,
		CreatedAt:      createdAt,
	}, nil
}

// Helpers

func writeError(w http.ResponseWriter, status int, code, msg string, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResp{Code: code, Message: msg, Details: details})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// ...existing code...
func withCommonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[MIDDLEWARE] %s %s from %s - User-Agent: %s", r.Method, r.URL.Path, r.RemoteAddr, r.Header.Get("User-Agent"))

		// CORS
		w.Header().Set("Access-Control-Allow-Origin", "*") // 放宽全部域；生产可改为具体域
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

// ...existing code...

// bytesReader is a tiny helper to avoid importing bytes in the snippet header
func bytesReader(b []byte) *bytes.Reader { return bytes.NewReader(b) }

// 下载通用文件，返回字节
func downloadFile(ctx context.Context, fileURL string) ([]byte, error) {
	log.Printf("[DOWNLOAD] Starting download from: %s", fileURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		log.Printf("[DOWNLOAD] Failed to create request: %v", err)
		return nil, err
	}
	resp, err := httpClient.Do(req)
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

// 将 int 转为字符串（避免多次 fmt 引入）
func toString(v int) string { return strconv.Itoa(v) }
