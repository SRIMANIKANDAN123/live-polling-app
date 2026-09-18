package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"live-polling-app/backend/internal/utils"
)

// RequireAuth validates the Authorization: Bearer <token> header and
// stores the userID/email in the request context for handlers to use.
func RequireAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			utils.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or malformed Authorization header")
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := utils.ParseToken(jwtSecret, tokenStr)
		if err != nil {
			utils.Fail(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token")
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Next()
	}
}

// OptionalAuth attaches user info if a valid token is present, but never
// blocks the request — used on public poll endpoints where voting works
// both logged-in and anonymous.
func OptionalAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if strings.HasPrefix(header, "Bearer ") {
			tokenStr := strings.TrimPrefix(header, "Bearer ")
			if claims, err := utils.ParseToken(jwtSecret, tokenStr); err == nil {
				c.Set("userID", claims.UserID)
				c.Set("userEmail", claims.Email)
			}
		}
		c.Next()
	}
}
