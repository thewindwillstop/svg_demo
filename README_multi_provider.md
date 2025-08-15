# SVG Generation Service with Multiple Providers

A Go HTTP service that generates SVG images using multiple providers:
- **SVG.IO**: English-only prompts with translation support
- **Recraft**: Supports Chinese prompts directly

## Features

- 🌐 Multiple image generation providers
- 🔄 Automatic Chinese-to-English translation for SVG.IO
- 📁 Direct SVG file download or JSON metadata response
- 🚀 High-performance HTTP service
- 🎨 Various style options

## API Endpoints

### SVG.IO Provider (with translation)
- `POST /v1/images/svg` - Direct SVG download (SVG.IO)
- `POST /v1/images/svgio` - Direct SVG download (SVG.IO)  
- `POST /v1/images` - JSON metadata (SVG.IO)

### Recraft Provider (Chinese support)
- `POST /v1/images/recraft/svg` - Direct SVG download (Recraft)
- `POST /v1/images/recraft` - JSON metadata (Recraft)

### Health Check
- `GET /health` - Service health status

## Request Format

```json
{
  "prompt": "一只可爱的小猫",
  "style": "cartoon",
  "skip_translate": false
}
```

### Request Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `prompt` | string | Yes | Description of the image to generate |
| `style` | string | No | Style of the image (e.g., "cartoon", "realistic") |
| `skip_translate` | boolean | No | Skip translation for SVG.IO (default: false) |

## Response Formats

### Direct SVG Download
Returns the SVG file directly with headers:
- `Content-Type: image/svg+xml`
- `X-Image-Id`: Generated image ID
- `X-Provider`: Provider used (svgio/recraft)
- `X-Was-Translated`: Whether translation was applied

### JSON Metadata
```json
{
  "id": "img_123",
  "svg_url": "https://...",
  "width": 1024,
  "height": 1024,
  "provider": "recraft",
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

### Using Recraft with Chinese prompt (direct)
```bash
curl -X POST http://localhost:8080/v1/images/recraft/svg \
  -H "Content-Type: application/json" \
  -d '{"prompt": "一只可爱的小猫", "style": "cartoon"}' \
  -o cat.svg
```

### Get JSON metadata
```bash
curl -X POST http://localhost:8080/v1/images/recraft \
  -H "Content-Type: application/json" \
  -d '{"prompt": "一只可爱的小猫", "style": "cartoon"}'
```

## Provider Selection

- **SVG.IO**: Better for English prompts, requires translation for Chinese
- **Recraft**: Native Chinese support, newer model capabilities
- The service automatically handles provider-specific requirements

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
