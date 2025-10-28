package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Service handles caching operations
type Service struct {
	client *redis.Client
	ttl    time.Duration
}

// NewService creates a new cache service
func NewService(client *redis.Client, ttl time.Duration) *Service {
	if client == nil {
		return nil // Cache disabled
	}
	return &Service{
		client: client,
		ttl:    ttl,
	}
}

// Get retrieves a value from cache
func (s *Service) Get(ctx context.Context, key string, dest interface{}) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("cache not available")
	}

	val, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("cache miss")
	}
	if err != nil {
		return fmt.Errorf("cache error: %w", err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	return nil
}

// Set stores a value in cache
func (s *Service) Set(ctx context.Context, key string, value interface{}) error {
	if s == nil || s.client == nil {
		return nil // Cache disabled, don't error
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return s.client.Set(ctx, key, data, s.ttl).Err()
}

// Delete removes a value from cache
func (s *Service) Delete(ctx context.Context, key string) error {
	if s == nil || s.client == nil {
		return nil
	}

	return s.client.Del(ctx, key).Err()
}

// DeletePattern removes all keys matching a pattern
func (s *Service) DeletePattern(ctx context.Context, pattern string) error {
	if s == nil || s.client == nil {
		return nil
	}

	iter := s.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := s.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

// IsAvailable checks if cache is available
func (s *Service) IsAvailable() bool {
	if s == nil || s.client == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	return s.client.Ping(ctx).Err() == nil
}

// GenerateKey creates a cache key from components
func GenerateKey(prefix string, parts ...string) string {
	key := prefix
	for _, part := range parts {
		if part != "" {
			key += ":" + part
		}
	}
	return key
}

// Cache key prefixes
const (
	PrefixArticle      = "article"
	PrefixArticles     = "articles"
	PrefixStats        = "stats"
	PrefixSources      = "sources"
	PrefixScraperInfo  = "scraper"
	PrefixAITrending   = "ai:trending"
	PrefixAISentiment  = "ai:sentiment"
	PrefixAIEntity     = "ai:entity"
	PrefixAIEnrichment = "ai:enrichment"
)
