package middleware

import (
	"net/http"

	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// RequireRole gates a route group to a single JWT role (e.g. "client" for the
// read-only client-portal endpoints). Must run after JWTAuthMiddleware,
// which is what populates the "role" context key.
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get("role")
		if !exists || val.(string) != role {
			response.Error(c, http.StatusForbidden, "Forbidden")
			c.Abort()
			return
		}
		c.Next()
	}
}
