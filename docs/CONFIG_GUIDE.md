# SVG Generation Service Configuration

本项目已重构为基于YAML的配置管理系统，提供更灵活的配置方式。

## 📁 配置文件

### 主配置文件：`config.yaml`
项目根目录下的主配置文件，包含所有服务配置。

### 环境变量配置：`.env`
包含敏感信息如API密钥等。

## 🔧 配置文件结构

### 服务器配置
```yaml
server:
  port: 8080                # 服务端口
  host: "0.0.0.0"          # 监听地址
  timeout: 60s             # 服务器超时
  read_timeout: 30s        # 读取超时
  write_timeout: 30s       # 写入超时
```

### Provider配置
```yaml
providers:
  svgio:
    enabled: true                          # 是否启用
    base_url: "https://api.svg.io"        # API基础URL
    endpoints:
      generate: "/v1/generate-image"       # 生成端点
      get_image: "/v1/get-image/"         # 获取图片端点
    timeout: 60s                          # 请求超时
    max_retries: 3                        # 最大重试次数

  recraft:
    enabled: true
    base_url: "https://external.api.recraft.ai"
    endpoints:
      generate: "/v1/images/generations"
      vectorize: "/v1/images/vectorize"
    timeout: 60s
    max_retries: 3
    default_model: "recraftv3"
    supported_models: ["recraftv3", "recraftv2"]

  claude:
    enabled: true
    base_url: "https://api.qnaigc.com/v1/"
    endpoints:
      chat: "/chat/completions"
    timeout: 60s
    max_retries: 3
    default_model: "claude-4.0-sonnet"
    max_tokens: 4000
    temperature: 0.7
```

### 翻译服务配置
```yaml
translation:
  enabled: true                                    # 是否启用翻译
  service_url: "https://api.siliconflow.cn/v1/chat/completions"
  default_model: "deepseek-ai/DeepSeek-R1-0528-Qwen3-8B"
  timeout: 45s
  max_retries: 2
  fallback_enabled: true
  fallback_models: ["gpt-3.5-turbo", "gpt-4"]    # 备用模型
```

### HTTP客户端配置
```yaml
http_client:
  timeout: 60s                    # 全局HTTP超时
  max_idle_conns: 100            # 最大空闲连接数
  max_idle_conns_per_host: 10    # 每个host最大空闲连接
  idle_conn_timeout: 90s         # 空闲连接超时
  dial_timeout: 30s              # 拨号超时
  tls_handshake_timeout: 10s     # TLS握手超时
```

### 功能特性开关
```yaml
features:
  enable_cors: true              # 启用CORS
  enable_metrics: false          # 启用指标采集
  enable_tracing: false          # 启用链路追踪
  enable_rate_limiting: false    # 启用限流
  enable_caching: false          # 启用缓存
```

### 安全配置
```yaml
security:
  enable_api_key_validation: false  # 启用API Key验证
  allowed_origins: ["*"]            # 允许的源
  max_request_size: "10MB"          # 最大请求大小
  enable_request_id: true           # 启用请求ID
```

## 🚀 使用方法

### 1. 基本启动
```bash
# 使用默认配置文件 config.yaml
go run main.go

# 指定配置文件路径
CONFIG_PATH=/path/to/custom/config.yaml go run main.go
```

### 2. 环境变量覆盖
```bash
# 通过环境变量指定配置文件
export CONFIG_PATH=/etc/svg-service/config.yaml
go run main.go
```

### 3. Docker部署
```dockerfile
# Dockerfile 示例
COPY config.yaml /app/config.yaml
ENV CONFIG_PATH=/app/config.yaml
```

## 🔧 配置验证

系统启动时会自动验证配置：

### 必要验证
- 服务器端口范围 (1-65535)
- 至少启用一个Provider
- 启用的Provider必须有有效的base_url

### 可选验证
- URL格式验证
- 超时时间合理性检查
- 模型名称有效性

## 🎯 配置最佳实践

### 1. 环境分离
```yaml
# 开发环境
server:
  port: 8080
logging:
  level: "debug"

# 生产环境  
server:
  port: 80
logging:
  level: "info"
features:
  enable_metrics: true
```

### 2. 安全配置
```yaml
# 生产环境安全配置
security:
  enable_api_key_validation: true
  allowed_origins: ["https://your-domain.com"]
  max_request_size: "5MB"

features:
  enable_rate_limiting: true
```

### 3. 性能优化
```yaml
# 高并发配置
http_client:
  max_idle_conns: 200
  max_idle_conns_per_host: 20

providers:
  svgio:
    timeout: 30s
    max_retries: 1
```

## 🔄 配置热更新

当前版本需要重启服务来应用配置更改。未来版本将支持配置热更新。

## 📋 迁移指南

### 从硬编码配置迁移

**之前 (config.go)**:
```go
const SVGIOBaseURL = "https://api.svg.io"
const SVGIOGeneratePath = "/v1/generate-image"
```

**现在 (config.yaml)**:
```yaml
providers:
  svgio:
    base_url: "https://api.svg.io"
    endpoints:
      generate: "/v1/generate-image"
```

**代码中使用**:
```go
// 之前
url := config.SVGIOBaseURL + config.SVGIOGeneratePath

// 现在  
url := config.AppConfig.Providers.SVGIO.BaseURL + config.AppConfig.Providers.SVGIO.Endpoints.Generate
```

## 🐛 故障排除

### 常见错误

1. **配置文件未找到**
   ```
   Error: configuration file not found: config.yaml
   ```
   解决：确保 config.yaml 在项目根目录或通过 CONFIG_PATH 指定路径

2. **Provider配置错误**
   ```
   Error: provider svgio is enabled but base_url is empty
   ```
   解决：检查 Provider 的 base_url 配置

3. **端口冲突**
   ```
   Error: invalid server port: 0
   ```
   解决：设置有效的端口范围 1-65535

### 调试技巧

1. **启用详细日志**
   ```yaml
   logging:
     level: "debug"
     enable_request_logging: true
   ```

2. **验证配置加载**
   ```bash
   # 检查配置是否正确加载
   go run main.go 2>&1 | grep "Configuration loaded"
   ```

3. **测试Provider连接**
   ```bash
   # 发送测试请求验证配置
   curl -X POST http://localhost:8080/health
   ```
