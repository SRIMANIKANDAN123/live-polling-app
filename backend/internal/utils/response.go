package utils

import "github.com/gin-gonic/gin"

// APIError is the consistent error shape returned to clients. Internal
// error details are never included here.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Success writes a consistent success envelope: {"success": true, "data": ...}
func Success(c *gin.Context, status int, data interface{}) {
	c.JSON(status, gin.H{
		"success": true,
		"data":    data,
	})
}

// Fail writes a consistent error envelope: {"success": false, "error": {...}}
func Fail(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"success": false,
		"error": APIError{
			Code:    code,
			Message: message,
		},
	})
}
