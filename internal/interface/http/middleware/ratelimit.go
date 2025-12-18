package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RateLimiter interface for rate limiting
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Duration, error)
}

// RateLimitConfig defines rate limit configuration per tier
type RateLimitConfig struct {
	FreeLimit      int
	PremiumLimit   int
	ProLimit       int
	WindowDuration time.Duration
}

// DefaultAPIRateLimitConfig returns default API rate limit config
func DefaultAPIRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		FreeLimit:      100,  // 100 requests per minute for free users
		PremiumLimit:   500,  // 500 requests per minute for premium
		ProLimit:       1000, // 1000 requests per minute for pro
		WindowDuration: time.Minute,
	}
}

// DefaultGenerationRateLimitConfig returns default generation rate limit config
func DefaultGenerationRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		FreeLimit:      -1, // Unlimited for free users (for testing)
		PremiumLimit:   -1, // Unlimited for premium
		ProLimit:       -1, // Unlimited for pro
		WindowDuration: 24 * time.Hour,
	}
}

// RateLimitMiddleware handles rate limiting
type RateLimitMiddleware struct {
	limiter RateLimiter
}

// NewRateLimitMiddleware creates a new RateLimitMiddleware
func NewRateLimitMiddleware(limiter RateLimiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: limiter,
	}
}

// Limit middleware applies rate limiting based on IP
func (m *RateLimitMiddleware) Limit(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("ratelimit:ip:%s", c.ClientIP())

		allowed, remaining, retryAfter, err := m.limiter.Allow(c.Request.Context(), key, limit, window)
		if err != nil {
			// On error, allow the request but log
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if !allowed {
			c.Header("Retry-After", fmt.Sprintf("%.0f", retryAfter.Seconds()))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"code":        "RATE_LIMIT_EXCEEDED",
				"retry_after": retryAfter.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// LimitByUser middleware applies rate limiting based on user ID and tier
func (m *RateLimitMiddleware) LimitByUser(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get(UserIDKey)
		if !exists {
			// Fall back to IP-based limiting for unauthenticated users
			if config.FreeLimit < 0 {
				// Unlimited
				c.Next()
				return
			}
			key := fmt.Sprintf("ratelimit:ip:%s", c.ClientIP())
			m.applyLimit(c, key, config.FreeLimit, config.WindowDuration)
			return
		}

		// Get user tier
		tier, _ := c.Get(UserTierKey)
		userTier, ok := tier.(entity.UserTier)
		if !ok {
			userTier = entity.UserTierFree
		}

		// Determine limit based on tier
		limit := config.FreeLimit
		switch userTier {
		case entity.UserTierPremium:
			limit = config.PremiumLimit
		case entity.UserTierPro:
			limit = config.ProLimit
		}

		// Check if unlimited (-1)
		if limit < 0 {
			// Unlimited - skip rate limiting
			c.Next()
			return
		}

		key := fmt.Sprintf("ratelimit:user:%s", userID.(uuid.UUID).String())
		m.applyLimit(c, key, limit, config.WindowDuration)
	}
}

// LimitGeneration applies rate limiting for video generation
func (m *RateLimitMiddleware) LimitGeneration() gin.HandlerFunc {
	config := DefaultGenerationRateLimitConfig()
	return m.LimitByUser(config)
}

// applyLimit applies the rate limit and sets headers
func (m *RateLimitMiddleware) applyLimit(c *gin.Context, key string, limit int, window time.Duration) {
	allowed, remaining, retryAfter, err := m.limiter.Allow(c.Request.Context(), key, limit, window)
	if err != nil {
		c.Next()
		return
	}

	c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
	c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

	if !allowed {
		c.Header("Retry-After", fmt.Sprintf("%.0f", retryAfter.Seconds()))
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":       "Rate limit exceeded",
			"code":        "RATE_LIMIT_EXCEEDED",
			"retry_after": retryAfter.Seconds(),
		})
		c.Abort()
		return
	}

	c.Next()
}
