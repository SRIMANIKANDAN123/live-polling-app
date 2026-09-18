package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"live-polling-app/backend/internal/utils"
)

// CORS allows the configured frontend origin to call the API with
// credentials/headers needed for JWT auth.
func CORS(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// SecurityHeaders adds a minimal, sane set of production HTTP headers.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// rateLimiter is a minimal fixed-window limiter keyed by client IP. It is
// process-local (fine for a single instance / assignment scope); a
// production multi-instance deployment would back this with Redis INCR +
// EXPIRE instead.
type rateLimiter struct {
	mu      sync.Mutex
	hits    map[string]int
	window  time.Duration
	limit   int
	resetAt time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{hits: make(map[string]int), window: window, limit: limit, resetAt: time.Now().Add(window)}
}

func (r *rateLimiter) allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if time.Now().After(r.resetAt) {
		r.hits = make(map[string]int)
		r.resetAt = time.Now().Add(r.window)
	}
	r.hits[key]++
	return r.hits[key] <= r.limit
}

// RateLimit returns middleware allowing `limit` requests per `window` per
// client IP. Used on voting and auth endpoints per the spec.
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	rl := newRateLimiter(limit, window)
	return func(c *gin.Context) {
		if !rl.allow(c.ClientIP()) {
			utils.Fail(c, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests, please slow down")
			return
		}
		c.Next()
	}
}
