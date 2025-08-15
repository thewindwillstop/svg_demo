package service

import (
	"context"
	"errors"

	"svg-generator/internal/config"
	"svg-generator/internal/types"
)

// Provider 定义上游服务提供商接口
type Provider interface {
	GenerateImage(ctx context.Context, req types.GenerateRequest) (*types.ImageResponse, error)
}

// ServiceManager 管理多个上游服务
type ServiceManager struct {
	svgioService   Provider
	recraftService Provider
	claudeService  Provider
}

// NewServiceManager 创建服务管理器
func NewServiceManager(svgioAPIKey, recraftAPIKey, claudeAPIKey, claudeBaseURL string) *ServiceManager {
	var svgioService *SVGIOService
	var recraftService *RecraftService
	var claudeService *ClaudeService

	if svgioAPIKey != "" && config.AppConfig.Providers.SVGIO.Enabled {
		svgioService = NewSVGIOService(svgioAPIKey)
	}

	if recraftAPIKey != "" && config.AppConfig.Providers.Recraft.Enabled {
		recraftService = NewRecraftService(recraftAPIKey)
	}

	if claudeAPIKey != "" && config.AppConfig.Providers.Claude.Enabled {
		claudeService = NewClaudeService(claudeAPIKey, claudeBaseURL)
	}

	return &ServiceManager{
		svgioService:   svgioService,
		recraftService: recraftService,
		claudeService:  claudeService,
	}
}

// RegisterProvider 注册新的Provider
func (sm *ServiceManager) RegisterProvider(providerType types.Provider, provider Provider) {
	switch providerType {
	case types.ProviderSVGIO:
		sm.svgioService = provider
	case types.ProviderRecraft:
		sm.recraftService = provider
	case types.ProviderClaude:
		sm.claudeService = provider
	}
}

// GetProvider 获取指定的Provider
func (sm *ServiceManager) GetProvider(providerType types.Provider) Provider {
	switch providerType {
	case types.ProviderSVGIO:
		return sm.svgioService
	case types.ProviderRecraft:
		return sm.recraftService
	case types.ProviderClaude:
		return sm.claudeService
	default:
		return sm.svgioService // 默认返回SVGIO
	}
}

// GenerateImage 根据提供商生成图像
func (sm *ServiceManager) GenerateImage(ctx context.Context, req types.GenerateRequest) (*types.ImageResponse, error) {
	provider := sm.GetProvider(req.Provider)
	if provider == nil {
		return nil, errors.New("provider not configured: " + string(req.Provider))
	}
	return provider.GenerateImage(ctx, req)
}
