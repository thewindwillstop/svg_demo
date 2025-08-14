# SVG 生成服务 - 重构版本 + 翻译功能

## 🆕 新增功能：自动翻译

支持中文提示词自动翻译为英文！当检测到中文字符时，会先调用 OpenAI API 进行翻译，然后使用翻译后的英文提示词生成图像。

### 翻译功能特点：
- 🧠 智能检测：自动识别中文字符
- 🔄 实时翻译：使用 OpenAI API 将中文翻译为英文
- 🚀 无缝集成：翻译失败不影响图像生成流程
- 📊 透明信息：响应中包含原文、译文和翻译状态
- ⏭️ 可跳过：支持 `skip_translate` 参数强制跳过翻译 

## 项目结构

```
Svg_demo/
├── main.go        # 主入口，服务启动和路由注册
├── types.go       # 数据类型定义（增加翻译字段）
├── config.go      # 配置常量
├── handlers.go    # HTTP 请求处理器（集成翻译逻辑）
├── upstream.go    # 上游 API 客户端
├── client.go      # HTTP 客户端工具
├── utils.go       # 工具函数和中间件
├── translate.go   # 🆕 翻译服务模块
├── test_translation.sh # 🆕 翻译功能测试脚本
├── .env.example   # 环境变量示例
└── API_DOC.md     # API 文档（更新翻译功能）
```

## 模块说明

### main.go
- 服务启动入口
- 环境变量加载
- 路由注册
- HTTP 服务器启动

### types.go
- API 请求响应类型定义
- 上游 API 类型定义
- 错误响应类型

### config.go
- API 端点配置
- 服务常量定义

### handlers.go
- `/v1/images/svg` - SVG 文件生成和下载
- `/v1/images` - 图像元数据生成
- `/ping` - 健康检查
- `/download` - URL 代理下载

### upstream.go
- SVG.IO API 客户端
- 上游请求处理和响应解析

### client.go
- 通用文件下载客户端
- HTTP 请求工具

### translate.go
- OpenAI API 翻译服务实现
- 中文字符检测算法
- 翻译错误处理和降级策略

### test_translation.sh
- 翻译功能自动化测试脚本
- 验证中文翻译和英文跳过逻辑
- 测试 SVG 和 JSON 两种响应格式

## 优势

1. **模块化**: 代码按功能拆分，便于维护
2. **单一职责**: 每个文件负责特定功能
3. **易扩展**: 新功能可以轻松添加到对应模块
4. **易测试**: 模块化便于单元测试
5. **可读性**: 代码结构清晰，易于理解

## 运行方式

```bash
# 编译
go build .

# 运行
./Svg_demo

# 或直接运行
go run .
```

## 功能保持不变

重构后所有 API 功能和行为保持完全一致：
- `/v1/images/svg` - 直接返回 SVG 文件
- `/v1/images` - 返回图像元数据和 URL
- `/ping` - 健康检查
- `/download?url=` - URL 代理下载

## 环境要求

- Go 1.19+
- `.env` 文件包含 `SVGIO_API_KEY`
- 🆕 `.env` 文件包含 `OPENAI_API_KEY`（可选，用于翻译功能）

## 使用示例

### 中文输入自动翻译
```bash
# JSON 响应
curl -X POST http://localhost:8080/v1/images \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "一只可爱的卡通狐狸", "style": "卡通"}'

# 直接下载 SVG
curl -X POST http://localhost:8080/v1/images/svg \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "简约的猫头鹰图标"}' \
  -o owl.svg
```

### 跳过翻译
```bash
curl -X POST http://localhost:8080/v1/images \
  -H 'Content-Type: application/json' \
  -d '{
    "prompt": "A cute cartoon fox",
    "style": "cartoon",
    "skip_translate": true
  }'
```

### 运行测试脚本
```bash
# 确保服务运行中
go run . &

# 运行翻译功能测试
./test_translation.sh
```

Base URL: `http://localhost:8080`

提供能力:
1. 生成图片并返回元数据 (含 SVG/PNG 外链)：`POST /v1/images`
2. 直接生成并返回 SVG 文件（二进制响应，自动下载）：`POST /v1/images/svg`
3. 代理下载已有 SVG：`GET /v1/download?url=...`

环境变量:
- `SVGIO_API_KEY` (必需) 上游 svg.io 的 Bearer Token

所有 JSON 响应使用 UTF-8 编码；除 `/v1/images/svg` 与 `/v1/download`（可能直接返回 `image/svg+xml`）。

---
## 错误响应统一格式
```json
{
  "code": "upstream_error",
  "message": "failed to generate image",
  "details": "可选，额外调试信息"
}
```
常见 code:
| code | 含义 |
|------|------|
| invalid_json | 请求体 JSON 解析失败 |
| invalid_argument | 参数非法（如 prompt 过短） |
| method_not_allowed | HTTP 方法不支持 |
| upstream_error | 上游生成失败/状态码>=300 |
| download_error | 下载失败 |
| missing_parameter | 缺少必要查询参数 |
| invalid_url | URL 不合法 |

---
## 1. 生成图片 (返回元数据)
`POST /v1/images`

请求 Body:
```json
{
  "prompt": "A minimalist fox head vector logo",
  "negative_prompt": "text, watermark",
  "style": "FLAT_VECTOR"
}
```
字段说明:
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| prompt | string | 是 | 最少 3 字符 |
| negative_prompt | string | 是 | 反向提示词 |
| style | string | 否 | 样式标签 |
| format | string | 否 | 预留，当前忽略 |

成功响应 200:
```json
{
  "id": "img_xxx",
  "prompt": "A minimalist fox head vector logo",
  "negative_prompt": "text, watermark",
  "style": "flat",
  "svg_url": "https://cdn.svg.io/generated/abc123.svg",
  "png_url": "https://cdn.svg.io/generated/abc123.png",
  "width": 512,
  "height": 512,
  "created_at": "2025-08-13T09:11:22Z"
}
```

cURL 示例:
```bash
curl -X POST http://localhost:8080/v1/images \
  -H 'Content-Type: application/json' \
  -d '{"prompt":"A minimalist fox logo","style":"flat"}'
```

---
## 2. 直接生成并返回 SVG
`POST /v1/images/svg`

请求 Body 同上。

响应:
- Headers:
  - `Content-Type: image/svg+xml`
  - `Content-Disposition: attachment; filename="<id>.svg"`
  - `X-Image-Id`, `X-Image-Width`, `X-Image-Height`
- Body: SVG 文本

cURL:
```bash
curl -X POST http://localhost:8080/v1/images/svg \
  -H 'Content-Type: application/json' \
  -d '{"prompt":"Geometric owl emblem","style":"line"}' \
  -o owl.svg
```


---
## 前端交互建议
| 按钮 | 调用 | 返回 | 说明 |
|------|------|------|------|
| 生成 PNG | POST /v1/images | JSON | 使用 `png_url` 展示或下载 |
| 生成 SVG | POST /v1/images/svg | SVG | 浏览器自动下载 |
| 重新下载 SVG | GET /v1/download?url= | SVG | 统一代理避免跨域/失效 |


## 未来可扩展
- 异步任务队列 (返回 task_id 轮询状态)
- PNG 直接流式返回接口 `/v1/images/png`
- 生成参数增加 size / seed / color palette
- 缓存与速率限制

