package utils

import (
	"bytes"
	"os"
	"strconv"
)

// ========== 辅助函数 ==========

// GetEnv gets environment variable with default value
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// BytesReader creates a bytes reader from byte slice
func BytesReader(b []byte) *bytes.Reader {
	return bytes.NewReader(b)
}

// ToString converts integer to string
func ToString(v int) string {
	return strconv.Itoa(v)
}

// ReadAllBytes reads all bytes from a reader
func ReadAllBytes(r interface{}) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.(interface{ Read([]byte) (int, error) }))
	return buf.Bytes()
}
