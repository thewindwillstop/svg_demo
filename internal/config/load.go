package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// 全局配置实例
var AppConfig *Config

// LoadFromEnv 从环境变量加载配置
func LoadFromEnv() (*Config, error) {
	config := &Config{}

	// 服务器配置
	config.Server = ServerConfig{
		Port:         getEnvAsInt("SERVER_PORT", 8080),
		Host:         getEnvAsString("SERVER_HOST", "0.0.0.0"),
		Timeout:      getEnvAsDuration("SERVER_TIMEOUT", "60s"),
		ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", "30s"),
		WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", "30s"),
	}

	// SVG.IO配置
	config.Providers.SVGIO = SVGIOConfig{
		BaseURL:    getEnvAsString("SVGIO_BASE_URL", "https://api.svg.io"),
		Timeout:    getEnvAsDuration("SVGIO_TIMEOUT", "60s"),
		MaxRetries: getEnvAsInt("SVGIO_MAX_RETRIES", 3),
		Enabled:    getEnvAsBool("SVGIO_ENABLED", true),
		Endpoints: SVGIOEndpoints{
			Generate: getEnvAsString("SVGIO_GENERATE_ENDPOINT", "/v1/generate"),
		},
	}

	// Recraft配置
	config.Providers.Recraft = RecraftConfig{
		BaseURL:         getEnvAsString("RECRAFT_BASE_URL", "https://external.api.recraft.ai"),
		Timeout:         getEnvAsDuration("RECRAFT_TIMEOUT", "60s"),
		MaxRetries:      getEnvAsInt("RECRAFT_MAX_RETRIES", 3),
		Enabled:         getEnvAsBool("RECRAFT_ENABLED", true),
		DefaultModel:    getEnvAsString("RECRAFT_DEFAULT_MODEL", "recraftv3"),
		SupportedModels: getEnvAsStringSlice("RECRAFT_SUPPORTED_MODELS", "recraftv3,recraftv2"),
		Endpoints: RecraftEndpoints{
			Generate:  getEnvAsString("RECRAFT_GENERATE_ENDPOINT", "/v1/images/generations"),
			Vectorize: getEnvAsString("RECRAFT_VECTORIZE_ENDPOINT", "/v1/images/vectorize"),
		},
	}

	// OpenAI配置
	config.Providers.OpenAI = OpenAIConfig{
		BaseURL:      getEnvAsString("OPENAI_BASE_URL", "https://api.qnaigc.com/v1/"),
		Timeout:      getEnvAsDuration("OPENAI_TIMEOUT", "60s"),
		MaxRetries:   getEnvAsInt("OPENAI_MAX_RETRIES", 3),
		Enabled:      getEnvAsBool("OPENAI_ENABLED", true),
		DefaultModel: getEnvAsString("OPENAI_DEFAULT_MODEL", "claude-4.0-sonnet"),
		MaxTokens:    getEnvAsInt("OPENAI_MAX_TOKENS", 4000),
		Temperature:  getEnvAsFloat("OPENAI_TEMPERATURE", 0.7),
	}

	// 翻译服务配置
	config.Translation = TranslationConfig{
		Enabled:      getEnvAsBool("TRANSLATION_ENABLED", true),
		ServiceURL:   getEnvAsString("TRANSLATION_SERVICE_URL", "https://api.openai.com/v1/chat/completions"),
		DefaultModel: getEnvAsString("TRANSLATION_DEFAULT_MODEL", "gpt-3.5-turbo"),
		Timeout:      getEnvAsDuration("TRANSLATION_TIMEOUT", "45s"),
		MaxRetries:   getEnvAsInt("TRANSLATION_MAX_RETRIES", 2),
	}

	return config, nil
}

// InitConfig 初始化配置
func InitConfig() error {
	config, err := LoadFromEnv()
	if err != nil {
		return fmt.Errorf("failed to load config from environment: %w", err)
	}
	AppConfig = config
	return nil
}

// 辅助函数：从环境变量获取字符串值
func getEnvAsString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 辅助函数：从环境变量获取整数值
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// 辅助函数：从环境变量获取浮点数值
func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// 辅助函数：从环境变量获取布尔值
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// 辅助函数：从环境变量获取时间间隔值
func getEnvAsDuration(key, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return 30 * time.Second // 默认值
}

// 辅助函数：从环境变量获取字符串切片
func getEnvAsStringSlice(key, defaultValue string) []string {
	value := getEnvAsString(key, defaultValue)
	if value == "" {
		return []string{}
	}
	return strings.Split(value, ",")
}