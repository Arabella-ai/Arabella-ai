package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter implements rate limiting using Redis
type RateLimiter struct {
	client *redis.Client
}

// NewRateLimiter creates a new RateLimiter
func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

// Allow checks if a request is allowed under the rate limit
// Returns: allowed, remaining, retryAfter, error
func (r *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Duration, error) {
	now := time.Now().UnixNano()
	windowStart := now - int64(window)

	pipe := r.client.Pipeline()

	// Remove old entries
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// Count current entries
	countCmd := pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now),
		Member: fmt.Sprintf("%d", now),
	})

	// Set expiration
	pipe.Expire(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, 0, err
	}

	count := countCmd.Val()
	remaining := int(int64(limit) - count - 1)
	if remaining < 0 {
		remaining = 0
	}

	if count >= int64(limit) {
		// Get the oldest entry to calculate retry after
		oldest, err := r.client.ZRange(ctx, key, 0, 0).Result()
		if err != nil || len(oldest) == 0 {
			return false, remaining, window, nil
		}

		var oldestTime int64
		fmt.Sscanf(oldest[0], "%d", &oldestTime)
		retryAfter := time.Duration(oldestTime + int64(window) - now)
		if retryAfter < 0 {
			retryAfter = 0
		}

		return false, remaining, retryAfter, nil
	}

	return true, remaining, 0, nil
}

// Reset resets the rate limit for a key
func (r *RateLimiter) Reset(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// GetCount returns the current count for a key
func (r *RateLimiter) GetCount(ctx context.Context, key string, window time.Duration) (int64, error) {
	now := time.Now().UnixNano()
	windowStart := now - int64(window)

	// Remove old entries first
	r.client.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

	return r.client.ZCard(ctx, key).Result()
}

// FixedWindowAllow implements fixed window rate limiting (simpler, less accurate)
func (r *RateLimiter) FixedWindowAllow(ctx context.Context, key string, limit int, windowSeconds int) (bool, int, time.Duration, error) {
	// Create a key with the current window
	windowKey := fmt.Sprintf("%s:%d", key, time.Now().Unix()/int64(windowSeconds))

	// Increment the counter
	count, err := r.client.Incr(ctx, windowKey).Result()
	if err != nil {
		return false, 0, 0, err
	}

	// Set expiration on first request
	if count == 1 {
		r.client.Expire(ctx, windowKey, time.Duration(windowSeconds)*time.Second)
	}

	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	if count > int64(limit) {
		// Calculate retry after
		ttl, _ := r.client.TTL(ctx, windowKey).Result()
		if ttl < 0 {
			ttl = time.Duration(windowSeconds) * time.Second
		}
		return false, remaining, ttl, nil
	}

	return true, remaining, 0, nil
}

