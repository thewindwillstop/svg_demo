# SVG 生成服务 - 多Provider架构

## 🎯 项目概述

基于Go语言的高性能SVG图像生成服务，支持多个AI图像生成Provider，采用模块化架构设计。

## ✨ 核心特性

### 🔄 多Provider支持
- **SVG.IO**: 专业SVG生成 + 自动翻译功能 
- **Recraft**: 中文原生支持 + 无背景优化
- **Claude**: AI代码生成 + 智能SVG创作

### 🌍 智能翻译
- 🧠 **智能检测**: 自动识别中文字符
- 🔄 **实时翻译**: OpenAI API驱动的中英翻译
- 🚀 **无缝集成**: 翻译失败不影响生成流程
- 📊 **透明信息**: 完整的翻译状态反馈
- ⏭️ **可选跳过**: 支持`skip_translate`参数

### 🏗️ 架构优势
- **策略模式**: 动态Provider切换
- **适配器模式**: 统一不同API接口
- **高并发**: Goroutine池 + 连接复用
- **容错设计**: 优雅降级 + 错误隔离 

## 📁 项目结构

```
Svg_demo/
├── cmd/                    # 应用程序入口
├── internal/              # 内部模块
│   ├── client/           # HTTP客户端工具
│   ├── config/           # 配置管理
│   ├── handlers/         # HTTP请求处理器
│   ├── translate/        # 翻译服务模块
│   ├── types/           # 数据类型定义
│   └── upstream/        # Provider适配器
├── pkg/                  # 公共工具包
│   └── utils/           # 工具函数
├── scripts/             # 脚本文件
├── docs/               # 文档目录
├── main.go             # 服务启动入口
├── go.mod              # Go模块定义
└── .env.example        # 环境变量示例
```

## 🧩 模块说明

### 核心模块

#### `main.go`
- 服务启动入口点
- 环境变量加载和验证
- 多Provider服务管理器初始化
- HTTP路由注册和服务启动

#### `internal/handlers/`
- **统一处理器**: 模板方法模式实现
- **Provider路由**: 支持SVG.IO、Recraft、Claude
- **CORS支持**: 完整的跨域处理
- **错误处理**: 统一的错误响应格式

#### `internal/upstream/`
- **ServiceManager**: Provider策略管理器
- **SVGIOService**: SVG.IO API适配器
- **RecraftService**: Recraft API适配器 + 背景优化
- **ClaudeService**: Claude AI适配器 + 智能提示

#### `internal/translate/`
- **OpenAI集成**: GPT模型翻译服务
- **中文检测**: Unicode字符识别算法
- **容错机制**: 翻译失败时优雅降级

#### `internal/types/`
- **统一数据模型**: 跨Provider标准化
- **API契约**: 请求/响应结构定义
- **Provider枚举**: 类型安全的Provider选择

### 工具模块

#### `pkg/utils/`
- **HTTP工具**: CORS、错误响应、公共头设置
- **通用函数**: 文件处理、字符串操作

#### `internal/client/`
- **HTTP客户端**: 连接池、超时控制
- **文件下载**: 流式处理、内存优化

#### `internal/config/`
- **配置管理**: API端点、常量定义
- **环境适配**: 开发/生产环境支持

## 🚀 快速开始

### 环境要求
- **Go 1.24+**
- **至少一个Provider API Key**
- **OpenAI API Key** (可选，用于翻译功能)

### 安装配置

1. **克隆项目**
```bash
git clone <repository-url>
cd Svg_demo
```

2. **配置环境变量**
```bash
cp .env.example .env
# 编辑 .env 文件，配置API密钥
```

3. **安装依赖**
```bash
go mod download
```

4. **启动服务**
```bash
go run main.go
```

### 环境变量说明

```bash
# SVG.IO Provider (支持翻译)
SVGIO_API_KEY=your_svgio_api_key_here

# Recraft Provider (中文原生)
RECRAFT_API_KEY=your_recraft_api_key_here
RECRAFT_API_URL=https://external.api.recraft.ai

# Claude Provider (AI代码生成)
CLAUDE_API_KEY=your_claude_api_key_here  
CLAUDE_BASE_URL=https://api.qnaigc.com/v1/

# 翻译服务 (可选)
OPENAI_API_KEY=your_openai_api_key_here
```

## 🎨 Provider特性对比

| Provider | 语言支持 | 特色功能 | 适用场景 |
|----------|----------|----------|----------|
| **SVG.IO** | 英文 + 自动翻译 | 专业SVG生成 | 高质量矢量图标 |
| **Recraft** | 中文原生支持 | 无背景优化 | 中文创作、透明背景 |
| **Claude** | 多语言 | AI代码生成 | 复杂SVG、编程创作 |

## 💡 使用示例

### 🔸 SVG.IO Provider (自动翻译)

```bash
# 中文输入 - JSON响应
curl -X POST http://localhost:8080/v1/images \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "一只可爱的卡通狐狸", "style": "FLAT_VECTOR"}'

# 中文输入 - 直接下载SVG
curl -X POST http://localhost:8080/v1/images/svg \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "简约的猫头鹰图标"}' \
  -o owl.svg

# 英文输入 - 跳过翻译
curl -X POST http://localhost:8080/v1/images \
  -H 'Content-Type: application/json' \
  -d '{
    "prompt": "A cute cartoon fox",
    "style": "FLAT_VECTOR",
    "skip_translate": true
  }'
```

### 🔸 Recraft Provider (中文原生)

```bash
# 中文创作 - 自动无背景
curl -X POST http://localhost:8080/v1/images/recraft/svg \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "极简主义的山峰图标", "style": "minimalism"}' \
  -o mountain.svg

# JSON元数据
curl -X POST http://localhost:8080/v1/images/recraft \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "科技感的机器人头像", "model": "recraftv3"}'
```

### 🔸 Claude Provider (AI代码生成)

```bash
# AI智能SVG生成
curl -X POST http://localhost:8080/v1/images/claude/svg \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "Create a responsive SVG logo with geometric patterns"}' \
  -o logo.svg

# 复杂图形创作
curl -X POST http://localhost:8080/v1/images/claude \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "Design a data visualization chart in SVG format"}'
```

### 🔸 健康检查

```bash
curl http://localhost:8080/health
```

## 🔗 API 文档

**Base URL**: `http://localhost:8080`

### 📋 可用端点

| Provider | SVG下载端点 | JSON元数据端点 | 特色 |
|----------|-------------|----------------|------|
| **SVG.IO** | `POST /v1/images/svg`<br>`POST /v1/images/svgio` | `POST /v1/images` | 自动翻译 |
| **Recraft** | `POST /v1/images/recraft/svg` | `POST /v1/images/recraft` | 中文原生 |
| **Claude** | `POST /v1/images/claude/svg` | `POST /v1/images/claude` | AI代码生成 |
| **通用** | - | `GET /health` | 健康检查 |

### 📝 请求格式

**通用请求体**:
```json
{
  "prompt": "图像描述文本",
  "negative_prompt": "不想要的元素",
  "style": "风格标签",
  "skip_translate": false,
  "provider": "auto"
}
```

**字段说明**:
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `prompt` | string | ✅ | 图像描述，最少3字符 |
| `negative_prompt` | string | ⬜ | 反向提示词 |
| `style` | string | ⬜ | 风格标签 |
| `skip_translate` | boolean | ⬜ | 跳过翻译(仅SVG.IO) |
| `model` | string | ⬜ | 模型选择(Recraft) |
| `size` | string | ⬜ | 图像尺寸(Recraft) |

### 📤 响应格式

**JSON元数据响应**:
```json
{
  "id": "img_xxx",
  "prompt": "A minimalist fox head vector logo",
  "negative_prompt": "text, watermark",
  "style": "FLAT_VECTOR",
  "svg_url": "https://cdn.provider.com/abc123.svg",
  "png_url": "https://cdn.provider.com/abc123.png",
  "width": 512,
  "height": 512,
  "created_at": "2025-08-15T09:11:22Z",
  "provider": "svgio",
  "original_prompt": "简约的狐狸头标志",
  "translated_prompt": "A minimalist fox head logo", 
  "was_translated": true
}
```

**SVG直接下载响应**:
- **Content-Type**: `image/svg+xml`
- **Content-Disposition**: `attachment; filename="<id>.svg"`
- **Headers**: `X-Image-Id`, `X-Image-Width`, `X-Image-Height`, `X-Provider`

### ⚠️ 错误处理

**统一错误响应格式**:
```json
{
  "code": "error_type",
  "message": "用户友好的错误描述", 
  "details": "可选的调试信息"
}
```

**常见错误码**:
| Code | HTTP状态 | 含义 |
|------|----------|------|
| `invalid_json` | 400 | 请求体JSON解析失败 |
| `invalid_argument` | 400 | 参数非法(如prompt过短) |
| `method_not_allowed` | 405 | HTTP方法不支持 |
| `upstream_error` | 502 | Provider API调用失败 |
| `parse_error` | 500 | 响应解析失败 |
| `timeout` | 504 | 请求超时 |

## 🏗️ 架构特点

### 设计模式应用
- **🎯 策略模式**: 多Provider动态切换
- **🔧 适配器模式**: 统一不同API接口
- **📋 模板方法**: 标准化请求处理流程
- **🏭 工厂模式**: Provider实例创建管理

### 性能优化
- **⚡ 连接复用**: HTTP连接池管理
- **🔄 并发处理**: Goroutine异步处理
- **⏱️ 超时控制**: 分层超时保护机制
- **🛡️ 容错设计**: 优雅降级策略

### 扩展性设计
- **🔌 插件化**: 新Provider易于接入
- **⚙️ 配置驱动**: 环境变量灵活配置
- **🧪 测试友好**: 接口抽象便于Mock
- **📊 监控就绪**: 预留指标采集点

## 🔮 路线图

### v1.1 (计划中)
- [ ] 批量生成API
- [ ] 图像缓存机制
- [ ] 限流和防护
- [ ] 监控指标采集

### v1.2 (规划中)  
- [ ] 微服务架构拆分
- [ ] 分布式任务队列
- [ ] 图像风格迁移
- [ ] WebSocket实时推送

## 🤝 贡献指南

欢迎提交Issue和Pull Request！

### 开发环境setup
```bash
# 克隆仓库
git clone <repo-url>
cd Svg_demo

# 安装依赖
go mod download

# 运行测试
go test ./...

# 启动开发服务
go run main.go
```

### 新Provider接入
1. 在`internal/upstream/`中实现Provider Service
2. 实现`UpstreamService`接口
3. 在`ServiceManager`中注册
4. 添加对应的Handler路由
5. 更新文档和测试

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

