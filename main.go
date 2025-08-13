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
	_ = godotenv.Load(".env")
	apiKey := os.Getenv("SVGIO_API_KEY")
	if apiKey == "" {
		log.Fatal("missing SVGIO_API_KEY")
	}

	mux := http.NewServeMux()

	// 新增: 直接生成并返回 SVG 文件
	mux.HandleFunc("/v1/images/svg", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only POST is allowed", nil)
			return
		}
		var req GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body", err.Error())
			return
		}
		if len(req.Prompt) < 3 {
			writeError(w, http.StatusBadRequest, "invalid_argument", "prompt must be at least 3 characters", nil)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 28*time.Second)
		defer cancel()

		img, err := callSVGIOGenerate(ctx, apiKey, req)
		if err != nil {
			status := http.StatusBadGateway
			if errors.Is(err, context.DeadlineExceeded) {
				status = http.StatusGatewayTimeout
			}
			writeError(w, status, "upstream_error", "failed to generate image", err.Error())
			return
		}

		// 下载 SVG 内容
		svgBytes, err := downloadFile(ctx, img.SVGURL)
		if err != nil {
			writeError(w, http.StatusBadGateway, "download_error", "failed to download generated svg", err.Error())
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+img.ID+".svg\"")
		// 可以附带元信息 header
		w.Header().Set("X-Image-Id", img.ID)
		w.Header().Set("X-Image-Width", toString(img.Width))
		w.Header().Set("X-Image-Height", toString(img.Height))
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(svgBytes); err != nil {
			log.Printf("write svg error: %v", err)
		}
	})

	// 原 JSON 接口（返回元数据 & URL）
	mux.HandleFunc("/v1/images", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only POST is allowed", nil)
			return
		}

		var req GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body", err.Error())
			return
		}
		if len(req.Prompt) < 3 {
			writeError(w, http.StatusBadRequest, "invalid_argument", "prompt must be at least 3 characters", nil)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 28*time.Second)
		defer cancel()

		img, err := callSVGIOGenerate(ctx, apiKey, req)
		if err != nil {
			status := http.StatusBadGateway
			if errors.Is(err, context.DeadlineExceeded) {
				status = http.StatusGatewayTimeout
			}
			writeError(w, status, "upstream_error", "failed to generate image", err.Error())
			return
		}

		writeJSON(w, http.StatusOK, img)
	})

	// 已有下载转发接口（外部 URL -> 文件）

	addr := "0.0.0.0:8080"
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, withCommonHeaders(mux)); err != nil {
		log.Fatal(err)
	}
}

func callSVGIOGenerate(ctx context.Context, apiKey string, req GenerateRequest) (*ImageResponse, error) {
	upReq := upstreamGenerateReq{
		Prompt:           req.Prompt,
		NegativePrompt:   req.NegativePrompt,
		Style:            req.Style,
		InitialImage:     nil,
		InitialImageType: nil,
	}

	body, _ := json.Marshal(upReq)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, upstreamBaseURL+generatePath, bytesReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		var raw any
		_ = json.NewDecoder(resp.Body).Decode(&raw)
		return nil, errors.New("upstream status: " + resp.Status)
	}

	var upResp upstreamGenerateResp
	if err := json.NewDecoder(resp.Body).Decode(&upResp); err != nil {
		return nil, err
	}
	if !upResp.Success || len(upResp.Data) == 0 {
		return nil, errors.New("upstream no data")
	}
	it := upResp.Data[0]

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

func withCommonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// bytesReader is a tiny helper to avoid importing bytes in the snippet header
func bytesReader(b []byte) *bytes.Reader { return bytes.NewReader(b) }

// 下载通用文件，返回字节
func downloadFile(ctx context.Context, fileURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, errors.New("fetch status: " + resp.Status)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// 将 int 转为字符串（避免多次 fmt 引入）
func toString(v int) string { return strconv.Itoa(v) }
