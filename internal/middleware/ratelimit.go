package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	mu      sync.Mutex
	limits  map[string]*userLimit
	global  int
	window  time.Duration
}

type userLimit struct {
	count    int
	resetAt  time.Time
}

func NewRateLimiter(global int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		limits: make(map[string]*userLimit),
		global: global,
		window: window,
	}
}

var globalLimiter = NewRateLimiter(100, 1*time.Minute)
var authLimiter = NewRateLimiter(5, 1*time.Minute)

func rateLimit(limiter *rateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetString("userId")
		if key == "" {
			key = c.ClientIP()
		}
		if key == "" {
			c.Next()
			return
		}

		limiter.mu.Lock()
		limit, exists := limiter.limits[key]
		now := time.Now()

		if !exists || now.After(limit.resetAt) {
			limiter.limits[key] = &userLimit{count: 1, resetAt: now.Add(limiter.window)}
			limiter.mu.Unlock()
			c.Next()
			return
		}

		if limit.count >= limiter.global {
			retryAfter := int(time.Until(limit.resetAt).Seconds())
			c.Header("X-RateLimit-Limit", strconv.Itoa(limiter.global))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", limit.resetAt.Format(time.RFC3339))
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			limiter.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limited",
				"message": "Too many requests. Please try again later.",
			})
			return
		}

		limit.count++
		remaining := limiter.global - limit.count
		c.Header("X-RateLimit-Limit", strconv.Itoa(limiter.global))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", limit.resetAt.Format(time.RFC3339))
		limiter.mu.Unlock()
		c.Next()
	}
}

func RateLimitGlobal() gin.HandlerFunc {
	return rateLimit(globalLimiter)
}

func RateLimitAuth() gin.HandlerFunc {
	return rateLimit(authLimiter)
}
