package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// RateLimiter middleware for API rate limiting using Redis
type RateLimiter struct {
	redis         *redis.Client
	maxRequests   int
	windowSeconds int
}

// NewRateLimiter creates a new rate limiter middleware
func NewRateLimiter(redisClient *redis.Client, maxRequests, windowSeconds int) *RateLimiter {
	return &RateLimiter{
		redis:         redisClient,
		maxRequests:   maxRequests,
		windowSeconds: windowSeconds,
	}
}

// Handler returns the Fiber middleware handler
func (rl *RateLimiter) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get client identifier (IP address or API key)
		identifier := c.IP()
		if apiKey := c.Get("X-API-Key"); apiKey != "" {
			identifier = apiKey
		}

		// Create Redis key
		key := fmt.Sprintf("rate_limit:%s", identifier)

		// Get current count
		ctx := context.Background()
		count, err := rl.redis.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			// Log error but don't block the request
			return c.Next()
		}

		// Check if rate limit exceeded
		if count >= rl.maxRequests {
			c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.maxRequests))
			c.Set("X-RateLimit-Remaining", "0")
			c.Set("Retry-After", fmt.Sprintf("%d", rl.windowSeconds))

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Rate limit exceeded",
				"message": fmt.Sprintf("Maximum %d requests per %d seconds", rl.maxRequests, rl.windowSeconds),
			})
		}

		// Increment counter
		pipe := rl.redis.Pipeline()
		pipe.Incr(ctx, key)
		if count == 0 {
			// Set expiry only on first request
			pipe.Expire(ctx, key, time.Duration(rl.windowSeconds)*time.Second)
		}
		_, err = pipe.Exec(ctx)
		if err != nil {
			// Log error but don't block the request
			return c.Next()
		}

		// Set rate limit headers
		remaining := rl.maxRequests - count - 1
		if remaining < 0 {
			remaining = 0
		}
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.maxRequests))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		return c.Next()
	}
}
