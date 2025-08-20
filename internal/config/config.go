package config

import (
	"fmt"
	"time"
)

// Config 应用程序配置结构
type Config struct {
	Server      ServerConfig
	Providers   ProvidersConfig
	Translation TranslationConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int
	Host         string
	Timeout      time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// ProvidersConfig 提供商配置
type ProvidersConfig struct {
	SVGIO   SVGIOConfig
	Recraft RecraftConfig
	OpenAI  OpenAIConfig
}

// SVGIOConfig SVG.IO提供商配置
type SVGIOConfig struct {
	BaseURL    string
	Timeout    time.Duration
	MaxRetries int
	Enabled    bool
	Endpoints  SVGIOEndpoints
}

// SVGIOEndpoints SVG.IO API端点配置
type SVGIOEndpoints struct {
	Generate string
}

// RecraftConfig Recraft提供商配置
type RecraftConfig struct {
	BaseURL         string
	Timeout         time.Duration
	MaxRetries      int
	Enabled         bool
	DefaultModel    string
	SupportedModels []string
	Endpoints       RecraftEndpoints
}

// RecraftEndpoints Recraft API端点配置
type RecraftEndpoints struct {
	Generate   string
	Vectorize  string
}

// OpenAIConfig OpenAI提供商配置 (支持所有OpenAI兼容的模型)
type OpenAIConfig struct {
	BaseURL      string
	Timeout      time.Duration
	MaxRetries   int
	Enabled      bool
	DefaultModel string
	MaxTokens    int
	Temperature  float64
}

// TranslationConfig 翻译服务配置
type TranslationConfig struct {
	Enabled      bool
	ServiceURL   string
	DefaultModel string
	Timeout      time.Duration
	MaxRetries   int
}


// GetServerAddr 获取服务器监听地址
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// IsProviderEnabled 检查Provider是否启用
func (c *Config) IsProviderEnabled(provider string) bool {
	switch provider {
	case "svgio":
		return c.Providers.SVGIO.Enabled
	case "recraft":
		return c.Providers.Recraft.Enabled
	case "openai":
		return c.Providers.OpenAI.Enabled
	default:
		return false
	}
}
