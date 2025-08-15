# SVG Generator Service - 多Provider架构

## 🎯 项目概述

基于Go语言的高性能SVG图像生成服务，支持多个AI图像生成Provider，采用模块化架构设计。现已完成Docker化部署，支持容器化运行。

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

### 🚀 Docker部署
- **容器化**: 多阶段构建，优化镜像大小
- **安全性**: 非root用户，安全扫描通过
- **编排支持**: Docker Compose + 管理脚本
- **健康检查**: 自动恢复 + 状态监控

### 🏗️ 架构优势
- **策略模式**: 动态Provider切换
- **适配器模式**: 统一不同API接口
- **高并发**: Goroutine池 + 连接复用
- **容错设计**: 优雅降级 + 错误隔离 

## 📁 项目结构

```
svg-generator/
├── internal/              # 内部模块
│   ├── config/           # 配置管理和加载
│   ├── handlers/         # HTTP请求处理器
│   ├── service/          # Provider服务实现
│   └── types/           # 数据类型定义
├── pkg/                  # 公共工具包
│   └── utils/           # 工具函数（HTTP、翻译、中间件）
├── scripts/             # 部署和管理脚本
├── docs/               # 文档目录
├── bin/                # 编译输出目录
├── main.go             # 服务启动入口
├── config.yaml         # 主配置文件
├── config.dev.yaml     # 开发环境配置
├── Dockerfile          # Docker镜像构建
├── docker-compose.yml  # 容器编排配置
├── go.mod              # Go模块定义
└── .env.example        # 环境变量示例
```

## 🧩 模块说明

### 核心模块

#### `main.go`
- 服务启动入口点
- 环境变量加载和验证（使用godotenv）
- 配置文件加载（支持自定义CONFIG_PATH）
- 多Provider服务管理器初始化
- HTTP路由注册和服务启动

#### `internal/handlers/`
- **统一处理器**: 模板方法模式实现
- **Provider路由**: 支持SVG.IO、Recraft、Claude
- **CORS支持**: 完整的跨域处理
- **错误处理**: 统一的错误响应格式

#### `internal/service/`
- **ServiceManager**: Provider策略管理器和接口抽象
- **SVGIOService**: SVG.IO API适配器
- **RecraftService**: Recraft API适配器 + 背景优化
- **ClaudeService**: Claude AI适配器 + 智能提示
- **Provider接口**: 统一的GenerateImage方法定义

#### `internal/types/`
- **统一数据模型**: 跨Provider标准化
- **API契约**: GenerateRequest/ImageResponse结构定义
- **Provider枚举**: 类型安全的Provider选择
- **上游API类型**: SVG.IO、Recraft、Claude的原生API结构

#### `internal/config/`
- **配置加载器**: YAML配置文件解析
- **环境适配**: 开发/生产环境支持
- **Provider配置**: 各Provider的启用状态和参数

### 工具模块

#### `pkg/utils/`
- **HTTP工具**: CORS中间件、错误响应、公共头设置
- **翻译服务**: OpenAI API集成的中英翻译功能
- **中文检测**: Unicode字符识别算法
- **文件处理**: HTTP下载、字节流处理
- **通用函数**: 类型转换、辅助工具

## 🚀 快速开始

### 部署方式选择

#### 🐳 Docker部署（推荐）

**快速启动**:
```bash
# 克隆项目
git clone <repository-url>
cd svg-generator

# 配置环境变量
cp .env.example .env
# 编辑 .env 文件，填入你的 API 密钥

# 一键部署
./scripts/docker.sh deploy
```

**完整部署指南**: 详见 [DEPLOYMENT.md](DEPLOYMENT.md)

#### 🔧 本地开发部署

**环境要求**:
- **Go 1.24+**
- **至少一个Provider API Key**
- **OpenAI API Key** (可选，用于翻译功能)

**安装配置**:

1. **克隆项目**
```bash
git clone <repository-url>
cd svg-generator
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
# 使用默认配置
go run main.go

# 使用自定义配置
CONFIG_PATH=config.dev.yaml go run main.go
```

### Docker使用

**基本命令**:
```bash
# 构建和运行
./scripts/docker.sh build
./scripts/docker.sh run

# 查看状态和日志
./scripts/docker.sh status
./scripts/docker.sh logs

# 使用Docker Compose
docker-compose up -d
docker-compose logs -f
```

配置文件支持：
- **config.yaml**: 生产环境配置
- **config.dev.yaml**: 开发环境配置
- **CONFIG_PATH环境变量**: 自定义配置文件路径

## 🎨 Provider特性对比

| Provider | 语言支持 | 特色功能 | 适用场景 |
|----------|----------|----------|----------|
| **SVG.IO** | 英文 + 自动翻译 | 专业SVG生成 | 高质量矢量图标 |
| **Recraft** | 中文原生支持 | 无背景优化 | 中文创作、透明背景 |
| **Claude** | 多语言 | AI代码生成 | 复杂SVG、编程创作 |

## 🏗️ 技术架构

### 核心技术栈
- **Go 1.24+**: 高性能后端服务语言
- **标准库HTTP**: 原生HTTP服务器，无额外框架依赖
- **YAML配置**: 灵活的配置管理方案
- **Docker**: 容器化部署和编排

### 架构设计模式
- **🎯 策略模式**: 多Provider动态切换，通过Provider接口统一管理
- **🔧 适配器模式**: 统一不同API接口，标准化输入输出格式
- **📋 模板方法**: 标准化请求处理流程，复用通用逻辑
- **🏭 工厂模式**: Provider实例创建管理，集中化服务初始化

### 性能优化
- **⚡ 连接复用**: HTTP客户端连接池管理，减少建连开销
- **🔄 并发处理**: Goroutine原生并发支持，高效处理并发请求
- **⏱️ 超时控制**: 分层超时保护机制，防止请求阻塞
- **🛡️ 容错设计**: 优雅降级策略，服务韧性保障

### 扩展性设计
- **🔌 插件化**: 新Provider易于接入，只需实现Provider接口
- **⚙️ 配置驱动**: 环境变量和YAML配置灵活管理
- **🧪 测试友好**: 接口抽象便于Mock测试
- **📊 监控就绪**: 预留指标采集点，支持后续监控集成

### 服务流程图

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Request  │───▶│  CORS & Auth    │───▶│   Route Match   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
                                                        ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Provider Call  │◀───│  Translate (*)  │◀───│ Request Parse   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │
         ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│    SVG.IO       │    │     Recraft     │    │     Claude      │
│   API Call      │    │    API Call     │    │   API Call      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 ▼
                    ┌─────────────────┐
                    │ Response Format │
                    │  JSON / SVG     │
                    └─────────────────┘
```

## 🔮 路线图

### v1.1 (计划中)
- [ ] API性能监控和指标收集
- [ ] 请求限流和防护机制
- [ ] 图像缓存和CDN集成
- [ ] 批量生成API端点
- [ ] Webhook回调支持

### v1.2 (规划中)  
- [ ] 图像风格迁移功能
- [ ] 自定义模型集成
- [ ] WebSocket实时推送
- [ ] 多语言翻译支持（不仅限于中英）
- [ ] 图像质量评估和优化


## 🤝 贡献指南

欢迎提交Issue和Pull Request！

### 开发环境setup
```bash
# 克隆仓库
git clone <repo-url>
cd svg-generator

# 安装依赖
go mod download

# 运行测试
go test ./...

# 启动开发服务
go run main.go
```

