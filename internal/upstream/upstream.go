package upstream

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"miniSvg/internal/client"
	"miniSvg/internal/config"
	"miniSvg/internal/types"
	"net/http"
	"time"
)

// upstreamGenerateReq 上游API生成请求
type upstreamGenerateReq struct {
	Prompt           string  `json:"prompt"`
	NegativePrompt   *string `json:"negativePrompt,omitempty"`
	Style            *string `json:"style,omitempty"`
	InitialImage     *string `json:"initialImage,omitempty"`
	InitialImageType *string `json:"initialImageType,omitempty"`
}

// upstreamGenerateResp 上游API生成响应
type upstreamGenerateResp struct {
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

// CallSVGIOGenerate 调用SVG.IO生成API
func CallSVGIOGenerate(ctx context.Context, apiKey string, req types.GenerateRequest) (*types.ImageResponse, error) {
	log.Printf("[UPSTREAM] Starting generation request...")

	upReq := upstreamGenerateReq{
		Prompt:           req.Prompt,
		NegativePrompt:   &req.NegativePrompt,
		Style:            &req.Style,
		InitialImage:     nil,
		InitialImageType: nil,
	}

	body, _ := json.Marshal(upReq)
	log.Printf("[UPSTREAM] Sending request to %s with payload size: %d bytes", config.UpstreamBaseURL+config.GeneratePath, len(body))

	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, config.UpstreamBaseURL+config.GeneratePath, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.HTTPClient.Do(httpReq)
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
	}, nil
}
