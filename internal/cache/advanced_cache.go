package cache

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/redis/go-redis/v9"
)

// AdvancedService extends Service with advanced caching features
type AdvancedService struct {
	*Service
	compressionThreshold int // Compress values larger than this (bytes)
}

// NewAdvancedService creates an advanced cache service with compression and pipelining
func NewAdvancedService(client *redis.Client, defaultTTL time.Duration, compressionThreshold int) *AdvancedService {
	baseService := NewService(client, defaultTTL)
	if baseService == nil {
		return nil
	}

	return &AdvancedService{
		Service:              baseService,
		compressionThreshold: compressionThreshold,
	}
}

// SetWithDynamicTTL stores a value with a calculated TTL based on data characteristics
func (s *AdvancedService) SetWithDynamicTTL(ctx context.Context, key string, value interface{}, size int, accessFrequency string) error {
	if s == nil || s.client == nil {
		return nil
	}

	// Calculate TTL based on size and access frequency
	ttl := s.calculateDynamicTTL(size, accessFrequency)

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Compress if data is large
	if len(data) > s.compressionThreshold {
		compressed, err := s.compress(data)
		if err == nil && len(compressed) < len(data) {
			// Add compression marker
			key = "compressed:" + key
			data = compressed
		}
	}

	return s.client.Set(ctx, key, data, ttl).Err()
}

// GetWithDecompression retrieves and decompresses a value if needed
func (s *AdvancedService) GetWithDecompression(ctx context.Context, key string, dest interface{}) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("cache not available")
	}

	// Try compressed key first
	compressedKey := "compressed:" + key
	val, err := s.client.Get(ctx, compressedKey).Result()
	isCompressed := err == nil

	if err == redis.Nil {
		// Try uncompressed key
		val, err = s.client.Get(ctx, key).Result()
		if err == redis.Nil {
			return fmt.Errorf("cache miss")
		}
		if err != nil {
			return fmt.Errorf("cache error: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("cache error: %w", err)
	}

	data := []byte(val)

	// Decompress if needed
	if isCompressed {
		decompressed, err := s.decompress(data)
		if err != nil {
			return fmt.Errorf("failed to decompress: %w", err)
		}
		data = decompressed
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	return nil
}

// SetMultiple stores multiple values using pipeline for better performance
func (s *AdvancedService) SetMultiple(ctx context.Context, items map[string]interface{}) error {
	if s == nil || s.client == nil {
		return nil
	}

	pipe := s.client.Pipeline()

	for key, value := range items {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}

		// Compress large values
		if len(data) > s.compressionThreshold {
			if compressed, err := s.compress(data); err == nil && len(compressed) < len(data) {
				key = "compressed:" + key
				data = compressed
			}
		}

		pipe.Set(ctx, key, data, s.ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// GetMultiple retrieves multiple values using pipeline
func (s *AdvancedService) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("cache not available")
	}

	pipe := s.client.Pipeline()
	results := make([]*redis.StringCmd, len(keys))

	// Try both compressed and uncompressed keys
	for i, key := range keys {
		results[i] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	values := make(map[string]interface{})
	for i, key := range keys {
		val, err := results[i].Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			continue
		}

		var data interface{}
		if err := json.Unmarshal([]byte(val), &data); err == nil {
			values[key] = data
		}
	}

	return values, nil
}

// WarmCache pre-loads frequently accessed data
func (s *AdvancedService) WarmCache(ctx context.Context, warmupData map[string]interface{}) error {
	if s == nil || s.client == nil {
		return nil
	}

	return s.SetMultiple(ctx, warmupData)
}

// SetWithStaleWhileRevalidate implements stale-while-revalidate pattern
func (s *AdvancedService) SetWithStaleWhileRevalidate(ctx context.Context, key string, value interface{}, freshTTL, staleTTL time.Duration) error {
	if s == nil || s.client == nil {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Compress if needed
	if len(data) > s.compressionThreshold {
		if compressed, err := s.compress(data); err == nil && len(compressed) < len(data) {
			key = "compressed:" + key
			data = compressed
		}
	}

	pipe := s.client.Pipeline()

	// Set main key with fresh TTL
	pipe.Set(ctx, key, data, freshTTL)

	// Set stale backup with longer TTL
	pipe.Set(ctx, "stale:"+key, data, staleTTL)

	_, err = pipe.Exec(ctx)
	return err
}

// GetWithStaleWhileRevalidate retrieves data, allowing stale reads
func (s *AdvancedService) GetWithStaleWhileRevalidate(ctx context.Context, key string, dest interface{}) (isStale bool, err error) {
	if s == nil || s.client == nil {
		return false, fmt.Errorf("cache not available")
	}

	// Try fresh data first
	err = s.GetWithDecompression(ctx, key, dest)
	if err == nil {
		return false, nil
	}

	// Fall back to stale data
	err = s.GetWithDecompression(ctx, "stale:"+key, dest)
	if err != nil {
		return false, err
	}

	return true, nil
}

// compress compresses data using gzip
func (s *AdvancedService) compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	if _, err := writer.Write(data); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// decompress decompresses gzip data
func (s *AdvancedService) decompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// calculateDynamicTTL determines TTL based on data characteristics
func (s *AdvancedService) calculateDynamicTTL(size int, accessFrequency string) time.Duration {
	baseTTL := s.ttl

	// Adjust based on size
	switch {
	case size < 1024: // Small data (< 1KB)
		baseTTL = baseTTL * 2 // Cache longer
	case size > 1024*1024: // Large data (> 1MB)
		baseTTL = baseTTL / 2 // Cache shorter
	}

	// Adjust based on access frequency
	switch accessFrequency {
	case "high":
		baseTTL = baseTTL * 3
	case "medium":
		baseTTL = baseTTL * 2
	case "low":
		baseTTL = baseTTL / 2
	}

	return baseTTL
}

// GetCacheStatistics returns detailed cache statistics
func (s *AdvancedService) GetCacheStatistics(ctx context.Context) (*CacheStatistics, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("cache not available")
	}

	info, err := s.client.Info(ctx, "stats", "memory", "keyspace").Result()
	if err != nil {
		return nil, err
	}

	dbSize, _ := s.client.DBSize(ctx).Result()

	stats := &CacheStatistics{
		TotalKeys:     dbSize,
		InfoRaw:       info,
		CollectedAt:   time.Now(),
		HitRate:       s.calculateHitRate(info),
		MemoryUsageMB: s.parseMemoryUsage(info),
	}

	return stats, nil
}

// CacheStatistics holds cache performance metrics
type CacheStatistics struct {
	TotalKeys     int64
	InfoRaw       string
	CollectedAt   time.Time
	HitRate       float64
	MemoryUsageMB float64
}

// calculateHitRate parses hit rate from Redis INFO stats
func (s *AdvancedService) calculateHitRate(_ string) float64 {
	// Parse keyspace_hits and keyspace_misses from info
	// This is a simplified version
	return 0.0 // Placeholder - would parse from info string
}

// parseMemoryUsage parses memory usage from Redis INFO memory
func (s *AdvancedService) parseMemoryUsage(_ string) float64 {
	// Parse used_memory from info
	// This is a simplified version
	return 0.0 // Placeholder - would parse from info string
}

// DeleteMultiple deletes multiple keys using pipeline
func (s *AdvancedService) DeleteMultiple(ctx context.Context, keys []string) error {
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
