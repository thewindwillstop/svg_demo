# SVG Generation Service with Multiple Providers

A Go HTTP service that generates SVG images using multiple providers:
- **SVG.IO**: English-only prompts with translation support
- **Recraft**: Supports Chinese prompts directly, optimized for no-background illustrations
- **Claude**: AI-powered SVG code generation with advanced prompt understanding

## Features

- 🌐 Three image generation providers with different strengths
- 🔄 Automatic Chinese-to-English translation for SVG.IO
- 📁 Direct SVG file download or JSON metadata response
- 🚀 High-performance HTTP service with extended timeouts
- 🎨 Various style options and intelligent prompt enhancement
- 🤖 AI-generated SVG code with Claude for precise vector graphics

## API Endpoints

### SVG.IO Provider (with translation)
- `POST /v1/images/svg` - Direct SVG download (SVG.IO)
- `POST /v1/images/svgio` - Direct SVG download (SVG.IO)  
- `POST /v1/images` - JSON metadata (SVG.IO)

### Recraft Provider (Chinese support, no background)
- `POST /v1/images/recraft/svg` - Direct SVG download (Recraft)
- `POST /v1/images/recraft` - JSON metadata (Recraft)

### Claude Provider (AI-generated SVG code)
- `POST /v1/images/claude/svg` - Direct SVG download (Claude)
- `POST /v1/images/claude` - JSON metadata (Claude)

### Health Check
- `GET /health` - Service health status

## Request Format

```json
{
  "prompt": "一只可爱的小猫",
  "style": "cartoon",
  "negative_prompt": "background, complex details",
  "skip_translate": false
}
```

### Request Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `prompt` | string | Yes | Description of the image to generate |
| `style` | string | No | Style of the image (e.g., "cartoon", "realistic", "vector_illustration") |
| `negative_prompt` | string | No | Elements to avoid in the generated image |
| `skip_translate` | boolean | No | Skip translation for SVG.IO (default: false) |

## Response Formats

### Direct SVG Download
Returns the SVG file directly with headers:
- `Content-Type: image/svg+xml`
- `X-Image-Id`: Generated image ID
- `X-Provider`: Provider used (svgio/recraft/claude)
- `X-Was-Translated`: Whether translation was applied (SVG.IO only)
- `X-Original-Prompt`: Original prompt before translation
- `X-Translated-Prompt`: Translated prompt (if applicable)

### JSON Metadata
```json
{
  "id": "img_123",
  "prompt": "一只可爱的小猫",
  "negative_prompt": "background",
  "style": "cartoon",
  "svg_url": "https://...",
  "png_url": "https://...",
  "width": 1024,
  "height": 1024,
  "created_at": "2025-08-15T14:30:00Z",
  "provider": "claude",
  "original_prompt": "一只可爱的小猫",
  "translated_prompt": "a cute cat",
  "was_translated": true
}
```

## Setup

1. **Clone and install dependencies:**
   ```bash
   git clone <repo>
   cd Svg_demo
   go mod download
   ```

2. **Configure environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your API keys
   ```

3. **Required API Keys:**
   - `SVGIO_API_KEY` - For SVG.IO provider
   - `RECRAFT_API_KEY` - For Recraft provider  
   - `CLAUDE_API_KEY` - For Claude provider
   - `CLAUDE_BASE_URL` - Claude API base URL (default: https://api.qnaigc.com/v1/)
   - `RECRAFT_API_URL` - Recraft API base URL (default: https://external.api.recraft.ai)
   - `OPENAI_API_KEY` - For translation (optional)

   At least one provider API key is required.

4. **Run the service:**
   ```bash
   go run main.go
   ```

## Usage Examples

### Using SVG.IO with Chinese prompt (auto-translation)
```bash
curl -X POST http://localhost:8080/v1/images/svg \
  -H "Content-Type: application/json" \
  -d '{"prompt": "一只可爱的小猫", "style": "cartoon"}' \
  -o cat.svg
```

### Using Recraft with Chinese prompt (direct, no background)
```bash
curl -X POST http://localhost:8080/v1/images/recraft/svg \
  -H "Content-Type: application/json" \
  -d '{"prompt": "一只可爱的小猫", "style": "vector_illustration"}' \
  -o cat.svg
```

### Using Claude for AI-generated SVG code
```bash
curl -X POST http://localhost:8080/v1/images/claude/svg \
  -H "Content-Type: application/json" \
  -d '{"prompt": "a smiling sun with rays", "style": "modern flat design"}' \
  -o sun.svg
```

### Get JSON metadata (Claude example)
```bash
curl -X POST http://localhost:8080/v1/images/claude \
  -H "Content-Type: application/json" \
  -d '{"prompt": "geometric mountain landscape", "style": "minimalist", "negative_prompt": "complex details, realistic textures"}'
```

## Provider Selection

- **SVG.IO**: 
  - Best for: English prompts, traditional SVG generation
  - Features: Requires translation for Chinese, stable API
  - Use when: You need reliable SVG generation with English descriptions

- **Recraft**: 
  - Best for: Chinese prompts, vector illustrations without backgrounds
  - Features: Native Chinese support, automatic background removal for vector styles
  - Use when: You want clean illustrations without backgrounds, Chinese text support

- **Claude**: 
  - Best for: Custom SVG code generation, complex vector graphics, precise control
  - Features: AI-generated SVG code, understands detailed requirements, creates semantic markup
  - Use when: You need precise control over SVG structure, complex graphics, or custom vector art

The service automatically handles provider-specific requirements and optimizations.

## Architecture

```
main.go
├── internal/
│   ├── handlers/     # HTTP request handlers
│   ├── upstream/     # Provider integrations
│   ├── translate/    # Translation service
│   ├── types/        # Type definitions
│   ├── config/       # Configuration
│   └── client/       # HTTP client utilities
└── pkg/
    └── utils/        # Utility functions
```

## Development

```bash
# Run tests
go test ./...

# Build binary
go build -o svg-service .

# Run with hot reload
go run main.go
```

## Testing Different Providers

```bash
# Test SVG.IO with translation
curl -X POST http://localhost:8080/v1/images/svg \
  -H "Content-Type: application/json" \
  -d '{"prompt": "红色的龙", "style": "fantasy"}'

# Test Recraft with no background
curl -X POST http://localhost:8080/v1/images/recraft \
  -H "Content-Type: application/json" \
  -d '{"prompt": "简约的树叶图标", "style": "vector_illustration"}'

# Test Claude AI-generated SVG
curl -X POST http://localhost:8080/v1/images/claude \
  -H "Content-Type: application/json" \
  -d '{"prompt": "geometric hexagon pattern", "style": "modern minimalist", "negative_prompt": "3D effects, shadows"}'
```

## Performance Notes

- Translation timeout: 45 seconds
- Image generation timeout: 60 seconds  
- HTTP client timeout: 60 seconds
- All providers support concurrent requests
- Claude generates SVG code as base64 data URLs for immediate use
