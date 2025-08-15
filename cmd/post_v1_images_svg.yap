// Generate image (direct SVG).
//
// Request:
//   POST /v1/images/svg
//
// Response:
//   200 (image/svg+xml)

import (
    "github.com/goplus/builder/spx-backend/internal/controller"
)

ctx := &Context

var req struct {
    Prompt         string `json:"prompt"`
    NegativePrompt string `json:"negative_prompt"`
    Style          string `json:"style"`
    SkipTranslate  bool   `json:"skip_translate"`
}

if err := bindJSON(ctx, &req); err != nil {
    replyWithCodeMsg(ctx, errorInvalidArgs, "invalid json")
    return
}
if len(req.Prompt) < 3 {
    replyWithCodeMsg(ctx, errorInvalidArgs, "prompt too short")
    return
}

params := controller.NewGenerateImageParams()
params.Prompt = req.Prompt
params.NegativePrompt = req.NegativePrompt
params.Style = req.Style
params.SkipTranslate = req.SkipTranslate
params.DirectSVG = true

svgBytes, meta, err := ctrl.GenerateImageSVG(ctx.Context(), params)
if err != nil {
    replyWithInnerError(ctx, err)
    return
}

setHeader(ctx, "Content-Type", "image/svg+xml")
setHeader(ctx, "Content-Disposition", "attachment; filename=\""+meta.ID+".svg\"")
setHeader(ctx, "X-Image-Id", meta.ID)
setHeader(ctx, "X-Image-Width", intToString(meta.Width))
setHeader(ctx, "X-Image-Height", intToString(meta.Height))
writeBytes(ctx, svgBytes)
