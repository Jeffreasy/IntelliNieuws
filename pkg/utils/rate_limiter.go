package utils

import (
	"context"
	"sync"
	"time"
)

// ScraperRateLimiter manages rate limiting for web scraping
type ScraperRateLimiter struct {
	delay      time.Duration
	lastAccess map[string]time.Time
	mu         sync.Mutex
}

// NewScraperRateLimiter creates a new rate limiter for scraping
func NewScraperRateLimiter(delaySeconds int) *ScraperRateLimiter {
	return &ScraperRateLimiter{
		delay:      time.Duration(delaySeconds) * time.Second,
		lastAccess: make(map[string]time.Time),
	}
}

// Wait blocks until it's safe to make a request to the given domain
func (rl *ScraperRateLimiter) Wait(ctx context.Context, domain string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Check if we need to wait
	if lastTime, exists := rl.lastAccess[domain]; exists {
		elapsed := time.Since(lastTime)
		if elapsed < rl.delay {
			waitTime := rl.delay - elapsed

			// Release lock while waiting
			rl.mu.Unlock()

			// Wait with context support
			select {
			case <-time.After(waitTime):
				// Continue
			case <-ctx.Done():
				// Re-acquire lock before returning
				rl.mu.Lock()
				return ctx.Err()
			}

			// Re-acquire lock after waiting
			rl.mu.Lock()
		}
	}

	// Update last access time
	rl.lastAccess[domain] = time.Now()
	return nil
}

// SetDelay updates the delay duration
func (rl *ScraperRateLimiter) SetDelay(delaySeconds int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.delay = time.Duration(delaySeconds) * time.Second
}

// Clear clears all rate limit records
func (rl *ScraperRateLimiter) Clear() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.lastAccess = make(map[string]time.Time)
}

// GetDelay returns the current delay duration
func (rl *ScraperRateLimiter) GetDelay() time.Duration {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.delay
}
