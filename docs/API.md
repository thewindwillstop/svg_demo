# SVG Generation Service - API 文档

## 📖 目录
- [API概述](#api概述)
- [认证说明](#认证说明)
- [通用规范](#通用规范)
- [错误处理](#错误处理)
- [Provider端点](#provider端点)
- [请求示例](#请求示例)
- [响应示例](#响应示例)
- [SDK和工具](#sdk和工具)

---

## 🎯 API概述

### 基本信息
- **服务名称**: SVG Generation Service
- **API版本**: v1
- **基础URL**: `http://localhost:8080`
- **协议**: HTTP/HTTPS
- **数据格式**: JSON
- **字符编码**: UTF-8

### 核心功能
- 🎨 多Provider AI图像生成 (SVG.IO, Recraft, Claude)
- 🌍 智能中英文翻译
- 📄 SVG矢量图生成
- 🔄 JSON元数据和直接文件下载两种模式

---

## 🔐 认证说明

### API密钥配置
服务需要配置相应的Provider API密钥：

```bash
# 环境变量配置
SVGIO_API_KEY=your_svgio_api_key_here
RECRAFT_API_KEY=your_recraft_api_key_here  
CLAUDE_API_KEY=your_claude_api_key_here
OPENAI_API_KEY=your_openai_api_key_here  # 翻译服务
```

### 安全说明
- 客户端无需传递API密钥
- 服务端统一管理所有Provider认证
- 支持CORS跨域请求

---

## 📐 通用规范

### HTTP方法
- `POST`: 创建图像生成任务
- `GET`: 获取服务状态和健康检查
- `OPTIONS`: CORS预检请求

### 请求头
```http
Content-Type: application/json
Accept: application/json, image/svg+xml
```

### 响应头
```http
Content-Type: application/json | image/svg+xml
X-Provider: svgio | recraft | claude
X-Request-ID: uuid
Access-Control-Allow-Origin: *
```

### 数据类型规范
- 所有时间使用 ISO 8601 格式 (`2025-08-15T10:30:00Z`)
- 图像尺寸使用像素值 (整数)
- Provider枚举值: `svgio`, `recraft`, `claude`

---

## ⚠️ 错误处理

### 错误响应格式
所有错误响应使用统一JSON格式：

```json
{
  "code": "error_type",
  "message": "用户友好的错误描述",
  "details": "可选的调试信息"
}
```

### 错误码参考

| HTTP状态码 | 错误码 | 含义 | 解决方案 |
|-----------|--------|------|----------|
| `400` | `invalid_json` | JSON解析失败 | 检查请求体格式 |
| `400` | `invalid_argument` | 参数非法 | 检查prompt长度等参数 |
| `405` | `method_not_allowed` | HTTP方法不支持 | 使用POST方法 |
| `500` | `parse_error` | 响应解析失败 | 联系技术支持 |
| `502` | `upstream_error` | Provider API失败 | 稍后重试或更换Provider |
| `504` | `timeout` | 请求超时 | 简化prompt或稍后重试 |

### 错误示例
```json
{
  "code": "invalid_argument",
  "message": "prompt must be at least 3 characters",
  "details": "Current prompt length: 2"
}
```

---

## 🔌 Provider端点

### 端点概览

| Provider | JSON元数据端点 | 直接SVG下载端点 | 特色功能 |
|----------|---------------|----------------|----------|
| **SVG.IO** | `POST /v1/images/svgio` | `POST /v1/images/svgio/svg` | 自动翻译 |
| **Recraft** | `POST /v1/images/recraft` | `POST /v1/images/recraft/svg` | 中文原生 |
| **Claude** | `POST /v1/images/claude` | `POST /v1/images/claude/svg` | AI代码生成 |

### 通用请求体格式

```json
{
  "prompt": "图像描述文本",
  "negative_prompt": "不想要的元素",
  "style": "风格标签",
  "skip_translate": false
}
```

### 字段说明

| 字段 | 类型 | 必填 | 长度限制 | 说明 |
|------|------|------|----------|------|
| `prompt` | string | ✅ | 3-500字符 | 图像描述，支持中英文 |
| `negative_prompt` | string | ❌ | 0-200字符 | 反向提示词，描述不想要的元素 |
| `style` | string | ❌ | 0-50字符 | 艺术风格标签 |
| `skip_translate` | boolean | ❌ | - | 仅SVG.IO有效，跳过翻译 |

### Provider特定参数

#### Recraft 额外参数
```json
{
  "model": "recraftv3",           // 模型版本: recraftv3, recraftv2
  "size": "1024x1024",           // 图像尺寸
  "substyle": "minimalism",      // 子风格
  "n": 1                         // 生成数量 (1-6)
}
```

#### Claude 额外参数
```json
{
  "temperature": 0.7,            // 创造性 (0.0-1.0)
  "max_tokens": 4000             // 最大token数
}
```

---

## 📝 请求示例

### 1. SVG.IO Provider

#### 获取JSON元数据
```bash
curl -X POST http://localhost:8080/v1/images/svgio \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "一只可爱的卡通狐狸",
    "style": "FLAT_VECTOR",
    "negative_prompt": "background, text"
  }'
```

#### 直接下载SVG
```bash
curl -X POST http://localhost:8080/v1/images/svgio/svg \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "minimalist cat icon",
    "skip_translate": true
  }' \
  -o cat_icon.svg
```

### 2. Recraft Provider

#### 中文创作
```bash
curl -X POST http://localhost:8080/v1/images/recraft \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "科技感的机器人头像",
    "model": "recraftv3",
    "style": "digital_art",
    "size": "512x512"
  }'
```

#### 无背景图标
```bash
curl -X POST http://localhost:8080/v1/images/recraft/svg \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "极简主义山峰图标",
    "style": "minimalism"
  }' \
  -o mountain.svg
```

### 3. Claude Provider

#### AI代码生成
```bash
curl -X POST http://localhost:8080/v1/images/claude \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Create a responsive SVG logo with geometric patterns",
    "temperature": 0.8
  }'
```

#### 复杂图形
```bash
curl -X POST http://localhost:8080/v1/images/claude/svg \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Design a data visualization chart showing growth trends"
  }' \
  -o chart.svg
```

### 4. 健康检查

```bash
curl -X GET http://localhost:8080/health
```

---

## 📄 响应示例

### JSON元数据响应

```json
{
  "id": "svgio_abc123def456",
  "prompt": "A cute cartoon fox",
  "negative_prompt": "background, text",
  "style": "FLAT_VECTOR",
  "svg_url": "https://cdn.svg.io/generated/abc123.svg",
  "png_url": "https://cdn.svg.io/generated/abc123.png",
  "width": 512,
  "height": 512,
  "created_at": "2025-08-15T10:30:15Z",
  "provider": "svgio",
  "original_prompt": "一只可爱的卡通狐狸",
  "translated_prompt": "A cute cartoon fox",
  "was_translated": true
}
```

### 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 唯一图像标识符 |
| `prompt` | string | 实际使用的提示词 |
| `negative_prompt` | string | 反向提示词 |
| `style` | string | 应用的风格 |
| `svg_url` | string | SVG文件下载链接 |
| `png_url` | string | PNG文件下载链接 (如果可用) |
| `width` | integer | 图像宽度 (像素) |
| `height` | integer | 图像高度 (像素) |
| `created_at` | string | 创建时间 (ISO 8601) |
| `provider` | string | 使用的Provider |
| `original_prompt` | string | 原始提示词 (翻译前) |
| `translated_prompt` | string | 翻译后提示词 |
| `was_translated` | boolean | 是否进行了翻译 |

### 直接SVG文件响应

```http
HTTP/1.1 200 OK
Content-Type: image/svg+xml
Content-Disposition: attachment; filename="claude_xyz789.svg"
Content-Length: 2048
X-Image-Id: claude_xyz789
X-Image-Width: 400
X-Image-Height: 300
X-Provider: claude
X-Was-Translated: false

<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 300">
  <!-- SVG内容 -->
</svg>
```

### 健康检查响应

```json
{
  "status": "ok",
  "time": "2025-08-15T10:30:00Z",
  "version": "1.0.0",
  "providers": {
    "svgio": "available",
    "recraft": "available", 
    "claude": "available"
  }
}
```

---

## 🛠️ SDK和工具

### JavaScript/TypeScript

```typescript
interface GenerateRequest {
  prompt: string;
  negative_prompt?: string;
  style?: string;
  skip_translate?: boolean;
}

interface ImageResponse {
  id: string;
  prompt: string;
  svg_url: string;
  png_url?: string;
  width: number;
  height: number;
  provider: 'svgio' | 'recraft' | 'claude';
  was_translated: boolean;
}

class SVGClient {
  constructor(private baseURL: string = 'http://localhost:8080') {}

  async generateSVGIO(request: GenerateRequest): Promise<ImageResponse> {
    const response = await fetch(`${this.baseURL}/v1/images/svgio`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request)
    });
    return response.json();
  }

  async downloadSVG(provider: string, request: GenerateRequest): Promise<Blob> {
    const response = await fetch(`${this.baseURL}/v1/images/${provider}/svg`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request)
    });
    return response.blob();
  }
}
```

### Python

```python
import requests
from typing import Optional, Dict, Any

class SVGClient:
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url
        
    def generate_svgio(self, prompt: str, style: Optional[str] = None, 
                      skip_translate: bool = False) -> Dict[str, Any]:
        """生成SVG.IO图像元数据"""
        data = {
            "prompt": prompt,
            "skip_translate": skip_translate
        }
        if style:
            data["style"] = style
            
        response = requests.post(
            f"{self.base_url}/v1/images/svgio",
            json=data
        )
        response.raise_for_status()
        return response.json()
    
    def download_svg(self, provider: str, prompt: str) -> bytes:
        """直接下载SVG文件"""
        response = requests.post(
            f"{self.base_url}/v1/images/{provider}/svg",
            json={"prompt": prompt}
        )
        response.raise_for_status()
        return response.content

# 使用示例
client = SVGClient()
result = client.generate_svgio("cute cat icon", style="minimalist")
svg_data = client.download_svg("claude", "geometric logo")
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

type GenerateRequest struct {
    Prompt        string `json:"prompt"`
    Style         string `json:"style,omitempty"`
    SkipTranslate bool   `json:"skip_translate,omitempty"`
}

type ImageResponse struct {
    ID        string `json:"id"`
    Prompt    string `json:"prompt"`
    SVGURL    string `json:"svg_url"`
    Provider  string `json:"provider"`
    Width     int    `json:"width"`
    Height    int    `json:"height"`
}

type SVGClient struct {
    BaseURL string
    Client  *http.Client
}

func NewSVGClient(baseURL string) *SVGClient {
    return &SVGClient{
        BaseURL: baseURL,
        Client:  &http.Client{},
    }
}

func (c *SVGClient) GenerateSVGIO(req GenerateRequest) (*ImageResponse, error) {
    data, _ := json.Marshal(req)
    resp, err := c.Client.Post(
        c.BaseURL+"/v1/images/svgio",
        "application/json",
        bytes.NewReader(data),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result ImageResponse
    err = json.NewDecoder(resp.Body).Decode(&result)
    return &result, err
}
```

---

## 📋 最佳实践

### 1. 错误处理
```javascript
try {
  const result = await client.generateSVGIO({
    prompt: "cute cat",
    style: "minimalist"
  });
  console.log('Generated:', result.svg_url);
} catch (error) {
  if (error.response?.status === 502) {
    console.log('Provider error, try again later');
  } else if (error.response?.status === 400) {
    console.log('Invalid parameters:', error.response.data.message);
  }
}
```

### 2. 超时处理
```javascript
const controller = new AbortController();
setTimeout(() => controller.abort(), 30000); // 30秒超时

fetch('/v1/images/claude', {
  method: 'POST',
  signal: controller.signal,
  body: JSON.stringify({prompt: "complex diagram"})
});
```

### 3. 批量处理
```javascript
const prompts = ["cat", "dog", "bird"];
const results = await Promise.allSettled(
  prompts.map(prompt => client.generateSVGIO({prompt}))
);
```

### 4. 缓存策略
```javascript
// 基于prompt内容的缓存键
const cacheKey = btoa(JSON.stringify({prompt, style, provider}));
const cached = localStorage.getItem(cacheKey);
if (cached) {
  return JSON.parse(cached);
}
```

---

## 🔄 版本历史

### v1.0.0 (2025-08-15)
- ✅ 初始API发布
- ✅ 支持SVG.IO, Recraft, Claude三个Provider
- ✅ 智能翻译功能
- ✅ JSON和SVG两种响应格式

### v1.1.0 (计划中)
- 🔄 批量生成API
- 🔄 异步任务队列
- 🔄 缓存机制
- 🔄 限流保护

---

## 📞 技术支持

### 联系方式
- **Issue反馈**: [GitHub Issues](https://github.com/your-org/svg-demo/issues)
- **技术文档**: [项目Wiki](https://github.com/your-org/svg-demo/wiki)
- **更新日志**: [CHANGELOG.md](../CHANGELOG.md)

### 常见问题
1. **Q: 支持哪些图像格式？**  
   A: 主要支持SVG格式，部分Provider提供PNG格式

2. **Q: 有请求频率限制吗？**  
   A: 目前无限制，未来版本会添加合理限流

3. **Q: 翻译功能支持哪些语言？**  
   A: 目前主要支持中文到英文的翻译

4. **Q: 如何选择最适合的Provider？**  
   A: SVG.IO适合英文创作，Recraft适合中文，Claude适合复杂图形

---

*最后更新: 2025-08-15*  
*API版本: v1.0.0*
