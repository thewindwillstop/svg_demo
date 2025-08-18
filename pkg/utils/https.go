package utils

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"
	"svg-generator/internal/types"
)


// HTTPClient is a global HTTP client with timeout
var HTTPClient = &http.Client{
	Timeout: 60 * time.Second,
}

// ========== HTTP 相关功能 ==========

// DownloadFile downloads a file from the given URL
func DownloadFile(ctx context.Context, fileURL string) ([]byte, error) {
	log.Printf("[DOWNLOAD] Starting download from: %s", fileURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		log.Printf("[DOWNLOAD] Failed to create request: %v", err)
		return nil, err
	}
	resp, err := HTTPClient.Do(req)
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

// WriteError writes an error response in JSON format
func WriteError(w http.ResponseWriter, status int, code, msg string, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(types.ErrorResp{Code: code, Message: msg, Details: details})
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
