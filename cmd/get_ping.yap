// Health check.
//
// Request:
//   GET /ping
//
// Response:
//   200 {"status":"ok","time":"..."}

import ()

ctx := &Context
json map[string]any{
    "status": "ok",
    "time":   timeNowRFC3339(),
}
