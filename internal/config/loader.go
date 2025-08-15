package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)
// LoadConfig 从YAML文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	// 如果没有指定配置文件路径，使用默认路径
	if configPath == "" {
		configPath = "config.yaml"
	}

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 尝试在项目根目录查找
		if _, err := os.Stat(filepath.Join(".", "config.yaml")); err == nil {
			configPath = filepath.Join(".", "config.yaml")
		} else {
			return nil, fmt.Errorf("configuration file not found: %s", configPath)
		}
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// InitConfig 初始化全局配置
func InitConfig(configPath string) error {
	config, err := LoadConfig(configPath)
	if err != nil {
		return err
	}

	AppConfig = config
	return nil
}

// validateConfig 验证配置有效性
func validateConfig(config *Config) error {
	// 验证服务器配置
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	// 验证至少启用一个Provider
	if !config.Providers.SVGIO.Enabled && !config.Providers.Recraft.Enabled && !config.Providers.Claude.Enabled {
		return fmt.Errorf("at least one provider must be enabled")
	}

	// 验证Provider URL
	providers := []struct {
		name    string
		enabled bool
		baseURL string
	}{
		{"svgio", config.Providers.SVGIO.Enabled, config.Providers.SVGIO.BaseURL},
		{"recraft", config.Providers.Recraft.Enabled, config.Providers.Recraft.BaseURL},
		{"claude", config.Providers.Claude.Enabled, config.Providers.Claude.BaseURL},
	}

	for _, p := range providers {
		if p.enabled && p.baseURL == "" {
			return fmt.Errorf("provider %s is enabled but base_url is empty", p.name)
		}
	}

	return nil
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
	case "claude":
		return c.Providers.Claude.Enabled
	default:
		return false
	}
}
