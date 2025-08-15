# 图像生成服务接口文档 (Image Generation Service API) - 支持翻译

Base URL: `http://localhost:8080`

## 🆕 翻译功能
本服务现在支持自动翻译中文提示词为英文！当检测到中文字符时，会先调用 OpenAI API 进行翻译，然后使用翻译后的英文提示词生成图像。

提供能力:
1. 生成图片并返回元数据 (含 SVG/PNG 外链 + 翻译信息)：`POST /v1/images`
2. 直接生成并返回 SVG 文件（二进制响应，自动下载）：`POST /v1/images/svg`
3. 代理下载已有 SVG：`GET /v1/download?url=...`

环境变量:
- `SVGIO_API_KEY` (必需) 上游 svg.io 的 Bearer Token
- `OPENAI_API_KEY` (可选) OpenAI API Key，用于翻译功能

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
  "prompt": "一只可爱的狐狸头像",
  "negative_prompt": "文字, 水印",
  "style": "flat",
  "skip_translate": false
}
```
字段说明:
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| prompt | string | 是 | 最少 3 字符，支持中文自动翻译 |
| negative_prompt | string | 否 | 反向提示词，也支持中文翻译 |
| style | string | 否 | 图像风格 |
| skip_translate | bool | 否 | 是否跳过翻译（默认 false） |
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
  "created_at": "2025-08-13T09:11:22Z",
  "original_prompt": "一只可爱的狐狸头像",
  "translated_prompt": "A minimalist fox head vector logo",
  "was_translated": true
}
```

新增字段说明:
| 字段 | 类型 | 说明 |
|------|------|------|
| original_prompt | string | 用户输入的原始提示词 |
| translated_prompt | string | 翻译后的英文提示词 |
| was_translated | bool | 是否进行了翻译处理 |

cURL 示例:
```bash
# 中文输入示例
curl -X POST http://localhost:8080/v1/images \
  -H 'Content-Type: application/json' \
  -d '{
    "prompt": "一只可爱的狐狸头像",
    "style": "卡通"
  }'

# 英文输入示例（跳过翻译）
curl -X POST http://localhost:8080/v1/images \
  -H 'Content-Type: application/json' \
  -d '{
    "prompt": "A cute fox avatar",
    "style": "cartoon",
    "skip_translate": true
  }'
---
## 2. 直接生成并返回 SVG
`POST /v1/images/svg`

请求 Body 同上，也支持中文翻译。

响应:
- Headers:
  - `Content-Type: image/svg+xml`
  - `Content-Disposition: attachment; filename="<id>.svg"`
  - `X-Image-Id`, `X-Image-Width`, `X-Image-Height`
  - 🆕 `X-Original-Prompt` (如果进行了翻译)
  - 🆕 `X-Translated-Prompt` (如果进行了翻译)
  - 🆕 `X-Was-Translated` (如果进行了翻译，值为 "true")
- Body: SVG 文本

cURL:
```bash
# 中文输入示例
curl -X POST http://localhost:8080/v1/images/svg \
  -H 'Content-Type: application/json' \
  -d '{
    "prompt": "几何猫头鹰图标",
    "style": "线条风格"
  }' \
  -o owl.svg

# 英文输入示例
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

---
**版本**: v0.1  
**最后更新**: 2025-08-13
