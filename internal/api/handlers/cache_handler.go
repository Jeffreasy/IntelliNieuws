package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jeffrey/intellinieuws/internal/cache"
	"github.com/jeffrey/intellinieuws/pkg/logger"
)

// CacheHandler handles cache-related endpoints
type CacheHandler struct {
	cacheService         *cache.Service
	advancedCacheService *cache.AdvancedService
	invalidationService  *cache.InvalidationService
	log                  *logger.Logger
}

// NewCacheHandler creates a new cache handler
func NewCacheHandler(
	cacheService *cache.Service,
	advancedCacheService *cache.AdvancedService,
	invalidationService *cache.InvalidationService,
	log *logger.Logger,
) *CacheHandler {
	return &CacheHandler{
		cacheService:         cacheService,
		advancedCacheService: advancedCacheService,
		invalidationService:  invalidationService,
		log:                  log,
	}
}

// GetStatistics returns cache statistics
func (h *CacheHandler) GetStatistics(c *fiber.Ctx) error {
	if h.advancedCacheService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Cache service not available",
		})
	}

	stats, err := h.advancedCacheService.GetCacheStatistics(c.Context())
	if err != nil {
		h.log.WithError(err).Error("Failed to get cache statistics")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve cache statistics",
		})
	}

	return c.JSON(fiber.Map{
		"status":          "ok",
		"total_keys":      stats.TotalKeys,
		"hit_rate":        stats.HitRate,
		"memory_usage_mb": stats.MemoryUsageMB,
		"collected_at":    stats.CollectedAt,
	})
}

// InvalidateCache invalidates cache entries based on pattern
func (h *CacheHandler) InvalidateCache(c *fiber.Ctx) error {
	if h.invalidationService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Cache invalidation service not available",
		})
	}

	type InvalidateRequest struct {
		Pattern       string `json:"pattern"`
		ArticleID     string `json:"article_id,omitempty"`
		Source        string `json:"source,omitempty"`
		StockSymbol   string `json:"stock_symbol,omitempty"`
		InvalidateAll bool   `json:"invalidate_all,omitempty"`
	}

	var req InvalidateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var err error
	ctx := c.Context()

	switch {
	case req.InvalidateAll:
		err = h.invalidationService.InvalidateAll(ctx)
	case req.ArticleID != "":
		err = h.invalidationService.InvalidateArticle(ctx, req.ArticleID)
	case req.Source != "":
		err = h.invalidationService.InvalidateBySource(ctx, req.Source)
	case req.StockSymbol != "":
		err = h.invalidationService.InvalidateStockData(ctx, req.StockSymbol)
	case req.Pattern != "":
		keys, _ := h.invalidationService.GetCacheKeys(ctx, req.Pattern)
		if len(keys) > 0 {
			err = h.invalidationService.DeleteMultiple(ctx, keys)
		}
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Must provide pattern, article_id, source, stock_symbol, or invalidate_all",
		})
	}

	if err != nil {
		h.log.WithError(err).Error("Failed to invalidate cache")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to invalidate cache",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "ok",
		"message": "Cache invalidated successfully",
	})
}

// GetCacheKeys returns all cache keys matching a pattern
func (h *CacheHandler) GetCacheKeys(c *fiber.Ctx) error {
	if h.invalidationService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Cache service not available",
		})
	}

	pattern := c.Query("pattern", "*")

	keys, err := h.invalidationService.GetCacheKeys(c.Context(), pattern)
	if err != nil {
		h.log.WithError(err).Error("Failed to get cache keys")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve cache keys",
		})
	}

	return c.JSON(fiber.Map{
		"status": "ok",
		"count":  len(keys),
		"keys":   keys,
	})
}

// GetCacheSize returns the total number of keys in cache
func (h *CacheHandler) GetCacheSize(c *fiber.Ctx) error {
	if h.invalidationService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Cache service not available",
		})
	}

	size, err := h.invalidationService.GetCacheSize(c.Context())
	if err != nil {
		h.log.WithError(err).Error("Failed to get cache size")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve cache size",
		})
	}

	return c.JSON(fiber.Map{
		"status":     "ok",
		"total_keys": size,
	})
}

// GetMemoryUsage returns Redis memory usage information
func (h *CacheHandler) GetMemoryUsage(c *fiber.Ctx) error {
	if h.invalidationService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Cache service not available",
		})
	}

	memInfo, err := h.invalidationService.GetCacheMemoryUsage(c.Context())
	if err != nil {
		h.log.WithError(err).Error("Failed to get memory usage")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve memory usage",
		})
	}

	return c.JSON(fiber.Map{
		"status":      "ok",
		"memory_info": memInfo,
	})
}

// WarmCache pre-loads frequently accessed data
func (h *CacheHandler) WarmCache(c *fiber.Ctx) error {
	if h.advancedCacheService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Cache service not available",
		})
	}

	type WarmupRequest struct {
		Data map[string]interface{} `json:"data"`
	}

	var req WarmupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.Data) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No data provided for cache warming",
		})
	}

	err := h.advancedCacheService.WarmCache(c.Context(), req.Data)
	if err != nil {
		h.log.WithError(err).Error("Failed to warm cache")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to warm cache",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "ok",
		"message": "Cache warmed successfully",
		"count":   len(req.Data),
	})
}
