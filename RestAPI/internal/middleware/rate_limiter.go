package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	cleanup  time.Duration
}

// NewRateLimiter creates a new rate limiter
// rate: tokens per second
// burst: maximum burst size
func NewRateLimiter(r float64, b int) *RateLimiter {
	limiter := &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(r),
		burst:    b,
		cleanup:  5 * time.Minute,
	}

	// Start cleanup goroutine
	go limiter.cleanupStaleEntries()

	return limiter
}

// getLimiter retrieves or creates a limiter for the given key
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[key] = limiter
	}

	return limiter
}

// cleanupStaleEntries removes inactive limiters periodically
func (rl *RateLimiter) cleanupStaleEntries() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for key, limiter := range rl.limiters {
			// Remove limiter if it has full tokens (indicating no recent activity)
			if limiter.Tokens() == float64(rl.burst) {
				delete(rl.limiters, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware returns a Gin middleware function
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use client IP as the key
		key := c.ClientIP()

		limiter := rl.getLimiter(key)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success":   false,
				"error":     "Rate limit exceeded",
				"code":      "TOO_MANY_REQUESTS",
				"timestamp": time.Now().Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByUser creates a rate limiter that uses user ID from context
func RateLimitByUser(r float64, b int) gin.HandlerFunc {
	limiter := NewRateLimiter(r, b)

	return func(c *gin.Context) {
		// Try to get user ID from context (set by auth middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			// Fallback to IP-based rate limiting
			userID = c.ClientIP()
		}

		key := userID.(string)
		rateLimiter := limiter.getLimiter(key)

		if !rateLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success":   false,
				"error":     "Rate limit exceeded",
				"code":      "TOO_MANY_REQUESTS",
				"timestamp": time.Now().Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByIP creates a simple IP-based rate limiter
func RateLimitByIP(requestsPerSecond float64, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(requestsPerSecond, burst)
	return limiter.Middleware()
}
