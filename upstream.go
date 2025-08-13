package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

// SVG.IO API 客户端
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
