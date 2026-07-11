package middleware

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware allows: any localhost/127.0.0.1 origin (Flutter web dev
// server, arbitrary port), any ngrok tunnel domain (the dev-over-internet
// flow already relied on elsewhere in this codebase), and whatever is
// explicitly listed in the CORS_ALLOWED_ORIGINS env var (comma-separated) —
// e.g. your production web app's real domain. AllowCredentials stays false
// since auth here is Bearer-JWT, never cookies, so there's no CSRF exposure
// from this being permissive in dev.
func CORSMiddleware(extraAllowedOrigins string) gin.HandlerFunc {
	extras := map[string]bool{}
	for _, o := range strings.Split(extraAllowedOrigins, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			extras[o] = true
		}
	}

	return cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			if extras[origin] {
				return true
			}
			return strings.Contains(origin, "://localhost") ||
				strings.Contains(origin, "://127.0.0.1") ||
				strings.Contains(origin, ".ngrok-free.app") ||
				strings.Contains(origin, ".ngrok-free.dev") ||
				strings.Contains(origin, ".ngrok.io") ||
				strings.Contains(origin, ".ngrok.app")
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "ngrok-skip-browser-warning"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	})
}
