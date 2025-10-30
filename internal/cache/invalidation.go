package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// InvalidationService handles cache invalidation strategies
type InvalidationService struct {
	client *redis.Client
}

// NewInvalidationService creates a new cache invalidation service
func NewInvalidationService(client *redis.Client) *InvalidationService {
	if client == nil {
		return nil
	}
	return &InvalidationService{
		client: client,
	}
}

// InvalidateArticle invalidates all cache entries related to an article
func (s *InvalidationService) InvalidateArticle(ctx context.Context, articleID string) error {
	if s == nil || s.client == nil {
		return nil // Cache disabled
	}

	patterns := []string{
		// Individual article
		fmt.Sprintf("%s:%s:*", PrefixArticle, articleID),
		// Article lists that might contain this article
		fmt.Sprintf("%s:*", PrefixArticles),
		// Stats that might be affected
		fmt.Sprintf("%s:*", PrefixStats),
	}

	return s.deletePatterns(ctx, patterns)
}

// InvalidateArticleList invalidates article list caches
func (s *InvalidationService) InvalidateArticleList(ctx context.Context) error {
	if s == nil || s.client == nil {
		return nil
	}

	patterns := []string{
		fmt.Sprintf("%s:*", PrefixArticles),
		fmt.Sprintf("%s:*", PrefixStats),
	}

	return s.deletePatterns(ctx, patterns)
}

// InvalidateAIData invalidates AI-related cache entries
func (s *InvalidationService) InvalidateAIData(ctx context.Context, articleID string) error {
	if s == nil || s.client == nil {
		return nil
	}

	patterns := []string{
		// AI enrichment for this article
		fmt.Sprintf("%s:%s", PrefixAIEnrichment, articleID),
		// Trending topics (might be affected)
		fmt.Sprintf("%s:*", PrefixAITrending),
		// Sentiment stats
		fmt.Sprintf("%s:*", PrefixAISentiment),
	}

	return s.deletePatterns(ctx, patterns)
}

// InvalidateStockData invalidates stock-related cache entries
func (s *InvalidationService) InvalidateStockData(ctx context.Context, symbol string) error {
	if s == nil || s.client == nil {
		return nil
	}

	patterns := []string{
		fmt.Sprintf("stock:quote:%s", symbol),
		fmt.Sprintf("stock:profile:%s", symbol),
		fmt.Sprintf("stock:news:%s", symbol),
		fmt.Sprintf("stock:metrics:%s", symbol),
		fmt.Sprintf("stock:ratings:%s", symbol),
		fmt.Sprintf("stock:target:%s", symbol),
	}

	return s.deletePatterns(ctx, patterns)
}

// InvalidateAllStockData invalidates all stock-related caches
func (s *InvalidationService) InvalidateAllStockData(ctx context.Context) error {
	if s == nil || s.client == nil {
		return nil
	}

	patterns := []string{
		"stock:*",
	}

	return s.deletePatterns(ctx, patterns)
}

// InvalidateBySource invalidates all articles from a specific source
func (s *InvalidationService) InvalidateBySource(ctx context.Context, source string) error {
	if s == nil || s.client == nil {
		return nil
	}

	patterns := []string{
		fmt.Sprintf("%s:source:%s:*", PrefixArticles, source),
		fmt.Sprintf("%s:*", PrefixStats),
	}

	return s.deletePatterns(ctx, patterns)
}

// InvalidateByEntity invalidates cache for a specific entity
func (s *InvalidationService) InvalidateByEntity(ctx context.Context, entityName string) error {
	if s == nil || s.client == nil {
		return nil
	}

	patterns := []string{
		fmt.Sprintf("%s:%s:*", PrefixAIEntity, entityName),
	}

	return s.deletePatterns(ctx, patterns)
}

// InvalidateAll clears all cache entries (use with caution!)
func (s *InvalidationService) InvalidateAll(ctx context.Context) error {
	if s == nil || s.client == nil {
		return nil
	}

	return s.client.FlushDB(ctx).Err()
}

// deletePatterns deletes all keys matching the given patterns
func (s *InvalidationService) deletePatterns(ctx context.Context, patterns []string) error {
	for _, pattern := range patterns {
		iter := s.client.Scan(ctx, 0, pattern, 0).Iterator()
		for iter.Next(ctx) {
			if err := s.client.Del(ctx, iter.Val()).Err(); err != nil {
				return fmt.Errorf("failed to delete key %s: %w", iter.Val(), err)
			}
		}
		if err := iter.Err(); err != nil {
			return fmt.Errorf("scan error for pattern %s: %w", pattern, err)
		}
	}
	return nil
}

// GetCacheKeys returns all keys matching a pattern (for debugging)
func (s *InvalidationService) GetCacheKeys(ctx context.Context, pattern string) ([]string, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("cache not available")
	}

	var keys []string
	iter := s.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return keys, nil
}

// GetCacheSize returns the number of keys in the cache
func (s *InvalidationService) GetCacheSize(ctx context.Context) (int64, error) {
	if s == nil || s.client == nil {
		return 0, fmt.Errorf("cache not available")
	}

	return s.client.DBSize(ctx).Result()
}

// GetCacheMemoryUsage returns memory usage information
func (s *InvalidationService) GetCacheMemoryUsage(ctx context.Context) (map[string]string, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("cache not available")
	}

	info, err := s.client.Info(ctx, "memory").Result()
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"memory_info": info,
	}, nil
}

// DeleteMultiple deletes multiple keys and their variants (compressed, stale)
func (s *InvalidationService) DeleteMultiple(ctx context.Context, keys []string) error {
	if s == nil || s.client == nil {
		return nil
	}

	pipe := s.client.Pipeline()
	for _, key := range keys {
		pipe.Del(ctx, key)
		pipe.Del(ctx, "compressed:"+key)
		pipe.Del(ctx, "stale:"+key)
	}

	_, err := pipe.Exec(ctx)
	return err
}
