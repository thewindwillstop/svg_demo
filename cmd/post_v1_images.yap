// Generate image (metadata & URLs).
//
// Request:
//   POST /v1/images
// Body(JSON):
//   { "prompt": "...", "negative_prompt": "...", "style": "..." }
//
// Response:
//   200 Image metadata JSON

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

resp, err := ctrl.GenerateImage(ctx.Context(), params)
if err != nil {
    replyWithInnerError(ctx, err)
    return
}
json resp
