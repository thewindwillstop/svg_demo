# SVG Generation Service API Documentation

## 服务概述

SVG Generation Service 是一个支持多提供商的图像生成服务，支持中文和英文提示词，可以生成高质量的SVG矢量图像。

### 支持的提供商
- **SVG.IO**: 英文优化，支持中文自动翻译
- **Recraft**: 原生支持中文，直接生成矢量图

### 服务地址
- 默认端口: `8080`
- 基础URL: `http://localhost:8080`

---

## API 端点

### 1. 健康检查

#### `GET /health`

检查服务运行状态。

**响应示例:**
```json
{
  "status": "ok",
  "time": "2025-08-15T10:30:00Z"
}
```

---

### 2. SVG.IO 提供商接口

#### `POST /v1/images/svg`
#### `POST /v1/images/svgio`

使用 SVG.IO 提供商生成并直接下载SVG文件。

**请求格式:**
```json
{
  "prompt": "一只可爱的小猫",
  "style": "cartoon",
  "negative_prompt": "ugly, blurry",
  "skip_translate": false
}
```

**请求参数:**

| 参数 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `prompt` | string | 是 | 图像描述（支持中文，自动翻译为英文） |
| `style` | string | 否 | 图像风格，默认: "FLAT_VECTOR" |
| `negative_prompt` | string | 否 | 负面提示词（要避免的内容） |
| `skip_translate` | boolean | 否 | 跳过翻译，默认: false |

**响应:**
- 返回SVG文件流
- Content-Type: `image/svg+xml`
- 包含以下响应头：
  - `X-Image-Id`: 图像ID
  - `X-Image-Width`: 图像宽度
  - `X-Image-Height`: 图像高度
  - `X-Provider`: 使用的提供商 (svgio)
  - `X-Was-Translated`: 是否进行了翻译 (true/false)
  - `X-Original-Prompt`: 原始提示词（如果进行了翻译）
  - `X-Translated-Prompt`: 翻译后的提示词（如果进行了翻译）

#### `POST /v1/images`

使用 SVG.IO 提供商生成图像，返回JSON元数据。

**请求格式:** 同上

**响应示例:**
```json
{
  "id": "img_abc123",
  "prompt": "a cute cat",
  "negative_prompt": "ugly, blurry",
  "style": "cartoon",
  "svg_url": "https://api.svg.io/download/abc123.svg",
  "png_url": "https://api.svg.io/download/abc123.png",
  "width": 1024,
  "height": 1024,
  "created_at": "2025-08-15T10:30:00Z",
  "provider": "svgio",
  "original_prompt": "一只可爱的小猫",
  "translated_prompt": "a cute cat",
  "was_translated": true
}
```

---

### 3. Recraft 提供商接口

#### `POST /v1/images/recraft/svg`

使用 Recraft 提供商生成并直接下载SVG文件。

**请求格式:**
```json
{
  "prompt": "一只可爱的小猫",
  "style": "vector_illustration",
  "negative_prompt": "ugly, blurry"
}
```

**请求参数:**

| 参数 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `prompt` | string | 是 | 图像描述（原生支持中文） |
| `style` | string | 否 | 图像风格，默认: "vector_illustration" |
| `negative_prompt` | string | 否 | 负面提示词 |

**Recraft 支持的风格:**
- `vector_illustration` - 矢量插画
- `digital_illustration` - 数字插画
- `realistic_image` - 写实图像
- `icon` - 图标风格

**响应:**
- 返回SVG文件流
- Content-Type: `image/svg+xml`
- 包含以下响应头：
  - `X-Image-Id`: 图像ID
  - `X-Provider`: recraft

#### `POST /v1/images/recraft`

使用 Recraft 提供商生成图像，返回JSON元数据。

**请求格式:** 同上

**响应示例:**
```json
{
  "id": "recraft_1692123456789",
  "prompt": "一只可爱的小猫",
  "negative_prompt": "ugly, blurry",
  "style": "vector_illustration",
  "svg_url": "https://api.recraft.ai/images/xyz789.svg",
  "png_url": "https://api.recraft.ai/images/xyz789.svg",
  "width": 1024,
  "height": 1024,
  "created_at": "2025-08-15T10:30:00Z",
  "provider": "recraft",
  "was_translated": false
}
```

---

## 错误响应

所有接口在出错时返回统一的错误格式：

```json
{
  "code": "error_code",
  "message": "Human readable error message",
  "details": "Additional error details"
}
```

**常见错误码:**

| 错误码 | HTTP状态码 | 描述 |
|--------|------------|------|
| `method_not_allowed` | 405 | 请求方法不允许 |
| `invalid_json` | 400 | 请求JSON格式错误 |
| `invalid_argument` | 400 | 参数错误（如prompt太短） |
| `upstream_error` | 502 | 上游API调用失败 |
| `download_error` | 502 | 文件下载失败 |

---

## 使用示例

### cURL 示例

**1. 使用SVG.IO生成SVG（中文提示词）:**
```bash
curl -X POST http://localhost:8080/v1/images/svg \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "一只坐在月亮上的小猫",
    "style": "cartoon"
  }' \
  -o cat_on_moon.svg
```

**2. 使用Recraft生成SVG（中文提示词）:**
```bash
curl -X POST http://localhost:8080/v1/images/recraft/svg \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "一只坐在月亮上的小猫",
    "style": "vector_illustration"
  }' \
  -o cat_on_moon_recraft.svg
```

**3. 获取JSON元数据:**
```bash
curl -X POST http://localhost:8080/v1/images/recraft \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "一只坐在月亮上的小猫",
    "style": "vector_illustration"
  }'
```

### JavaScript 示例

```javascript
// 使用 fetch API 获取SVG文件
async function generateSVG(prompt, provider = 'svgio') {
  const endpoint = provider === 'recraft' 
    ? '/v1/images/recraft/svg' 
    : '/v1/images/svg';
    
  const response = await fetch(`http://localhost:8080${endpoint}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      prompt: prompt,
      style: provider === 'recraft' ? 'vector_illustration' : 'cartoon'
    })
  });

  if (response.ok) {
    const svgBlob = await response.blob();
    const imageId = response.headers.get('X-Image-Id');
    const wasTranslated = response.headers.get('X-Was-Translated');
    
    return {
      svgBlob,
      imageId,
      wasTranslated: wasTranslated === 'true'
    };
  } else {
    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
  }
}

// 使用示例
generateSVG('一只可爱的小猫', 'recraft')
  .then(result => {
    console.log('生成成功:', result.imageId);
    // 创建下载链接
    const url = URL.createObjectURL(result.svgBlob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${result.imageId}.svg`;
    a.click();
  })
  .catch(console.error);
```

### Python 示例

```python
import requests
import json

def generate_svg_image(prompt, provider='svgio', style=None):
    """
    生成SVG图像
    
    Args:
        prompt (str): 图像描述
        provider (str): 'svgio' 或 'recraft'
        style (str): 图像风格
    
    Returns:
        bytes: SVG文件内容
    """
    if provider == 'recraft':
        url = 'http://localhost:8080/v1/images/recraft/svg'
        default_style = 'vector_illustration'
    else:
        url = 'http://localhost:8080/v1/images/svg'
        default_style = 'cartoon'
    
    payload = {
        'prompt': prompt,
        'style': style or default_style
    }
    
    response = requests.post(url, json=payload)
    
    if response.status_code == 200:
        return response.content, response.headers
    else:
        error = response.json()
        raise Exception(f"API Error: {error['message']}")

# 使用示例
try:
    svg_content, headers = generate_svg_image('一只可爱的小猫', 'recraft')
    
    # 保存SVG文件
    image_id = headers.get('X-Image-Id')
    with open(f'{image_id}.svg', 'wb') as f:
        f.write(svg_content)
    
    print(f"SVG生成成功: {image_id}.svg")
    print(f"提供商: {headers.get('X-Provider')}")
    
except Exception as e:
    print(f"生成失败: {e}")
```

---

## 配置说明

### 环境变量

| 变量名 | 必需 | 描述 |
|--------|------|------|
| `SVGIO_API_KEY` | 可选* | SVG.IO API密钥 |
| `RECRAFT_API_KEY` | 可选* | Recraft API密钥 |
| `OPENAI_API_KEY` | 可选 | OpenAI API密钥（用于翻译） |

*至少需要配置一个提供商的API密钥

### 启动服务

```bash
# 设置环境变量
export SVGIO_API_KEY="your_svgio_api_key"
export RECRAFT_API_KEY="your_recraft_api_key"
export OPENAI_API_KEY="your_openai_api_key"

# 启动服务
go run main.go
```

---

## 性能说明

### 超时设置
- 翻译超时: 45秒
- 图像生成超时: 60秒
- HTTP客户端超时: 60秒

### 提供商选择建议
- **中文提示词**: 推荐使用Recraft，原生支持，响应更快
- **英文提示词**: 两个提供商都可以，SVG.IO可能在某些风格上表现更好
- **复杂场景**: Recraft的vectorize API更适合复杂的矢量图生成

### 限制说明
- 提示词最少3个字符
- 单次请求生成1张图片
- 默认图片尺寸: 1024x1024
- 支持的响应格式: SVG, PNG预览

---

## 版本历史

### v1.0.0
- 支持SVG.IO和Recraft双提供商
- 中文提示词自动翻译
- 直接SVG下载和JSON元数据接口
- 健康检查接口
- CORS支持
