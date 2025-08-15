package config

import (

	"time"

)

// Config 应用程序配置结构
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Providers   ProvidersConfig   `yaml:"providers"`
	Translation TranslationConfig `yaml:"translation"`
	HTTPClient  HTTPClientConfig  `yaml:"http_client"`
	Logging     LoggingConfig     `yaml:"logging"`
	Features    FeaturesConfig    `yaml:"features"`
	Security    SecurityConfig    `yaml:"security"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int           `yaml:"port"`
	Host         string        `yaml:"host"`
	Timeout      time.Duration `yaml:"timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// ProvidersConfig 提供商配置
type ProvidersConfig struct {
	SVGIO   SVGIOConfig   `yaml:"svgio"`
	Recraft RecraftConfig `yaml:"recraft"`
	Claude  ClaudeConfig  `yaml:"claude"`
}

// SVGIOConfig SVG.IO提供商配置
type SVGIOConfig struct {
	BaseURL    string         `yaml:"base_url"`
	Endpoints  SVGIOEndpoints `yaml:"endpoints"`
	Timeout    time.Duration  `yaml:"timeout"`
	MaxRetries int            `yaml:"max_retries"`
	Enabled    bool           `yaml:"enabled"`
}

// SVGIOEndpoints SVG.IO端点配置
type SVGIOEndpoints struct {
	Generate string `yaml:"generate"`
	GetImage string `yaml:"get_image"`
}

// RecraftConfig Recraft提供商配置
type RecraftConfig struct {
	BaseURL         string           `yaml:"base_url"`
	Endpoints       RecraftEndpoints `yaml:"endpoints"`
	Timeout         time.Duration    `yaml:"timeout"`
	MaxRetries      int              `yaml:"max_retries"`
	Enabled         bool             `yaml:"enabled"`
	DefaultModel    string           `yaml:"default_model"`
	SupportedModels []string         `yaml:"supported_models"`
}

// RecraftEndpoints Recraft端点配置
type RecraftEndpoints struct {
	Generate  string `yaml:"generate"`
	Vectorize string `yaml:"vectorize"`
}

// ClaudeConfig Claude提供商配置
type ClaudeConfig struct {
	BaseURL      string          `yaml:"base_url"`
	Endpoints    ClaudeEndpoints `yaml:"endpoints"`
	Timeout      time.Duration   `yaml:"timeout"`
	MaxRetries   int             `yaml:"max_retries"`
	Enabled      bool            `yaml:"enabled"`
	DefaultModel string          `yaml:"default_model"`
	MaxTokens    int             `yaml:"max_tokens"`
	Temperature  float64         `yaml:"temperature"`
}

// ClaudeEndpoints Claude端点配置
type ClaudeEndpoints struct {
	Chat string `yaml:"chat"`
}

// TranslationConfig 翻译服务配置
type TranslationConfig struct {
	Enabled         bool          `yaml:"enabled"`
	ServiceURL      string        `yaml:"service_url"`
	DefaultModel    string        `yaml:"default_model"`
	Timeout         time.Duration `yaml:"timeout"`
	MaxRetries      int           `yaml:"max_retries"`
	FallbackEnabled bool          `yaml:"fallback_enabled"`
	FallbackModels  []string      `yaml:"fallback_models"`
}

// HTTPClientConfig HTTP客户端配置
type HTTPClientConfig struct {
	Timeout             time.Duration `yaml:"timeout"`
	MaxIdleConns        int           `yaml:"max_idle_conns"`
	MaxIdleConnsPerHost int           `yaml:"max_idle_conns_per_host"`
	IdleConnTimeout     time.Duration `yaml:"idle_conn_timeout"`
	DialTimeout         time.Duration `yaml:"dial_timeout"`
	TLSHandshakeTimeout time.Duration `yaml:"tls_handshake_timeout"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level                string `yaml:"level"`
	Format               string `yaml:"format"`
	Output               string `yaml:"output"`
	EnableRequestLogging bool   `yaml:"enable_request_logging"`
	EnableErrorStack     bool   `yaml:"enable_error_stack"`
}

// FeaturesConfig 功能特性配置
type FeaturesConfig struct {
	EnableCORS         bool `yaml:"enable_cors"`
	EnableMetrics      bool `yaml:"enable_metrics"`
	EnableTracing      bool `yaml:"enable_tracing"`
	EnableRateLimiting bool `yaml:"enable_rate_limiting"`
	EnableCaching      bool `yaml:"enable_caching"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	EnableAPIKeyValidation bool     `yaml:"enable_api_key_validation"`
	AllowedOrigins         []string `yaml:"allowed_origins"`
	MaxRequestSize         string   `yaml:"max_request_size"`
	EnableRequestID        bool     `yaml:"enable_request_id"`
}

// 全局配置实例
var AppConfig *Config
