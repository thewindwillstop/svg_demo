package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"svg-generator/internal/config"
	"svg-generator/internal/types"
	"svg-generator/pkg/utils"
)

// SVGIOService 实现 SVG.IO API 调用
type SVGIOService struct {
	apiKey  string
	baseURL string
}

// NewSVGIOService 创建 SVG.IO 服务实例
func NewSVGIOService(apiKey string) *SVGIOService {
	return &SVGIOService{
		apiKey:  apiKey,
		baseURL: config.AppConfig.Providers.SVGIO.BaseURL,
	}
}

// svgioGenerateReq SVG.IO API生成请求
type svgioGenerateReq struct {
	Prompt           string  `json:"prompt"`
	NegativePrompt   *string `json:"negativePrompt,omitempty"`
	Style            *string `json:"style,omitempty"`
}

// svgioGenerateResp SVG.IO API生成响应
type svgioGenerateResp struct {
	Success bool `json:"success"`
	Data    []struct {
		ID             string `json:"id"`
		Prompt         string `json:"prompt"`
		NegativePrompt string `json:"negativePrompt"`
		Style          string `json:"style"`
		SVGURL         string `json:"svgUrl"`
		PNGURL         string `json:"pngUrl"`
		Width          int    `json:"width"`
		Height         int    `json:"height"`
		CreatedAt      string `json:"createdAt"`
	} `json:"data"`
}

// GenerateImage 使用 SVG.IO API 生成图像
func (s *SVGIOService) GenerateImage(ctx context.Context, req types.GenerateRequest) (*types.ImageResponse, error) {
	log.Printf("[SVGIO] Starting generation request...")

	upReq := svgioGenerateReq{
		Prompt:           req.Prompt,
		NegativePrompt:   &req.NegativePrompt,
		Style:            &req.Style,
	}
	if req.NegativePrompt == "" {
		defaultNegativePrompt := "NULL"
		upReq.NegativePrompt = &defaultNegativePrompt
	}
	if req.Style == "" {
		defaultStyle := "FLAT_VECTOR"
		upReq.Style = &defaultStyle
	}

	body, _ := json.Marshal(upReq)
	apiURL := s.baseURL + config.AppConfig.Providers.SVGIO.Endpoints.Generate
	log.Printf("[SVGIO] Sending request to %s with payload size: %d bytes", apiURL, len(body))

	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := utils.HTTPClient.Do(httpReq)
	if err != nil {
		log.Printf("[SVGIO] HTTP request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("[SVGIO] Received response with status: %s", resp.Status)

	if resp.StatusCode >= 300 {
		var raw any
		_ = json.NewDecoder(resp.Body).Decode(&raw)
		log.Printf("[SVGIO] Error response body: %+v", raw)
		return nil, errors.New("upstream status: " + resp.Status)
	}

	var upResp svgioGenerateResp
	if err := json.NewDecoder(resp.Body).Decode(&upResp); err != nil {
		log.Printf("[SVGIO] Failed to decode response: %v", err)
		return nil, err
	}
	if !upResp.Success || len(upResp.Data) == 0 {
		log.Printf("[SVGIO] Invalid response: success=%v, data_count=%d", upResp.Success, len(upResp.Data))
		return nil, errors.New("upstream no data")
	}
	it := upResp.Data[0]
	log.Printf("[SVGIO] Successfully parsed response - ID: %s, SVG: %s, PNG: %s", it.ID, it.SVGURL, it.PNGURL)

	createdAt, _ := time.Parse(time.RFC3339, it.CreatedAt)
	return &types.ImageResponse{
		ID:             it.ID,
		Prompt:         it.Prompt,
		NegativePrompt: it.NegativePrompt,
		Style:          it.Style,
		SVGURL:         it.SVGURL,
		PNGURL:         it.PNGURL,
		Width:          it.Width,
		Height:         it.Height,
		CreatedAt:      createdAt,
		Provider:       types.ProviderSVGIO,
	}, nil
}

// CallSVGIOGenerate 调用SVG.IO生成API (保持向后兼容)
func CallSVGIOGenerate(ctx context.Context, apiKey string, req types.GenerateRequest) (*types.ImageResponse, error) {
	service := NewSVGIOService(apiKey)
	return service.GenerateImage(ctx, req)
}
