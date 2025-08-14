package client

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"time"
)

// HTTPClient is a global HTTP client with timeout
var HTTPClient = &http.Client{
	Timeout: 10 * time.Second,
}

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
