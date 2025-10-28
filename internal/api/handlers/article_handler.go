package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jeffrey/intellinieuws/internal/cache"
	"github.com/jeffrey/intellinieuws/internal/models"
	"github.com/jeffrey/intellinieuws/internal/repository"
	"github.com/jeffrey/intellinieuws/pkg/logger"
)

// ArticleHandler handles article-related HTTP requests
type ArticleHandler struct {
	repo           *repository.ArticleRepository
	cache          *cache.Service
	scraperService interface {
		EnrichArticleContent(ctx context.Context, articleID int64) error
	}
	logger *logger.Logger
}

// NewArticleHandler creates a new article handler
func NewArticleHandler(repo *repository.ArticleRepository, cacheService *cache.Service, log *logger.Logger) *ArticleHandler {
	return &ArticleHandler{
		repo:   repo,
		cache:  cacheService,
		logger: log.WithComponent("article-handler"),
	}
}

// SetScraperService sets the scraper service for content extraction
func (h *ArticleHandler) SetScraperService(scraperService interface {
	EnrichArticleContent(ctx context.Context, articleID int64) error
}) {
	h.scraperService = scraperService
}

// GetArticle handles GET /api/v1/articles/:id
func (h *ArticleHandler) GetArticle(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	// Parse ID
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_ID", "Article ID must be a valid integer", err.Error(), requestID),
		)
	}

	// Try cache first
	cacheKey := cache.GenerateKey(cache.PrefixArticle, c.Params("id"))
	var article models.Article

	if h.cache != nil {
		if err := h.cache.Get(c.Context(), cacheKey, &article); err == nil {
			h.logger.Debug("Cache hit for article")
			return c.JSON(models.NewSuccessResponse(article, requestID))
		}
	}

	// Cache miss - get from database
	articlePtr, err := h.repo.GetByID(c.Context(), id)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to get article: %d", id)
		return c.Status(fiber.StatusNotFound).JSON(
			models.NewErrorResponse("NOT_FOUND", "Article not found", fmt.Sprintf("No article with ID %d", id), requestID),
		)
	}

	// Store in cache
	if h.cache != nil {
		if err := h.cache.Set(c.Context(), cacheKey, articlePtr); err != nil {
			h.logger.WithError(err).Warn("Failed to cache article")
		}
	}

	return c.JSON(models.NewSuccessResponse(*articlePtr, requestID))
}

// ListArticles handles GET /api/v1/articles
func (h *ArticleHandler) ListArticles(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	// Parse query parameters
	filter := models.ArticleFilter{
		Source:    c.Query("source"),
		Category:  c.Query("category"),
		Keyword:   c.Query("keyword"),
		SortBy:    c.Query("sort_by", "published"),
		SortOrder: c.Query("sort_order", "desc"),
		Limit:     c.QueryInt("limit", 50),
		Offset:    c.QueryInt("offset", 0),
	}

	// Validate limit
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Limit < 1 {
		filter.Limit = 50
	}

	// Validate sort parameters
	validSortFields := map[string]bool{"published": true, "created_at": true, "title": true}
	if !validSortFields[filter.SortBy] {
		filter.SortBy = "published"
	}
	filter.SortOrder = strings.ToLower(filter.SortOrder)
	if filter.SortOrder != "asc" && filter.SortOrder != "desc" {
		filter.SortOrder = "desc"
	}

	// Parse date filters
	var startDateStr, endDateStr string
	if startDateStr = c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(
				models.NewErrorResponse("INVALID_DATE", "start_date must be in RFC3339 format", err.Error(), requestID),
			)
		}
		filter.StartDate = &startDate
	}

	if endDateStr = c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(
				models.NewErrorResponse("INVALID_DATE", "end_date must be in RFC3339 format", err.Error(), requestID),
			)
		}
		filter.EndDate = &endDate
	}

	// Generate cache key from query parameters
	cacheKey := cache.GenerateKey(cache.PrefixArticles,
		filter.Source,
		filter.Category,
		filter.Keyword,
		filter.SortBy,
		filter.SortOrder,
		fmt.Sprintf("limit:%d:offset:%d", filter.Limit, filter.Offset),
	)

	// Try cache first (only for simple queries without date filters)
	if h.cache != nil && filter.StartDate == nil && filter.EndDate == nil {
		var cachedArticles []models.Article
		var cachedTotal int
		cacheData := struct {
			Articles []models.Article
			Total    int
		}{}
		if err := h.cache.Get(c.Context(), cacheKey, &cacheData); err == nil {
			h.logger.Debug("Cache hit for articles list")
			cachedArticles = cacheData.Articles
			cachedTotal = cacheData.Total

			meta := &models.Meta{
				Pagination: models.CalculatePaginationMeta(cachedTotal, filter.Limit, filter.Offset),
				Sorting: &models.SortingMeta{
					SortBy:    filter.SortBy,
					SortOrder: filter.SortOrder,
				},
				Filtering: buildFilteringMeta(filter, startDateStr, endDateStr),
			}

			return c.JSON(models.NewSuccessResponseWithMeta(cachedArticles, meta, requestID))
		}
	}

	// Cache miss - get from database
	articles, total, err := h.repo.List(c.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list articles")
		return c.Status(fiber.StatusInternalServerError).JSON(
			models.NewErrorResponse("DATABASE_ERROR", "Failed to retrieve articles", err.Error(), requestID),
		)
	}

	// Store in cache (only for simple queries)
	if h.cache != nil && filter.StartDate == nil && filter.EndDate == nil {
		cacheData := struct {
			Articles []models.Article
			Total    int
		}{
			Articles: articles,
			Total:    total,
		}
		if err := h.cache.Set(c.Context(), cacheKey, cacheData); err != nil {
			h.logger.WithError(err).Warn("Failed to cache articles list")
		}
	}

	meta := &models.Meta{
		Pagination: models.CalculatePaginationMeta(total, filter.Limit, filter.Offset),
		Sorting: &models.SortingMeta{
			SortBy:    filter.SortBy,
			SortOrder: filter.SortOrder,
		},
		Filtering: buildFilteringMeta(filter, startDateStr, endDateStr),
	}

	return c.JSON(models.NewSuccessResponseWithMeta(articles, meta, requestID))
}

// GetStats handles GET /api/v1/articles/stats
func (h *ArticleHandler) GetStats(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)
	cacheKey := cache.GenerateKey(cache.PrefixStats, "comprehensive")

	// Try cache first
	var cachedStats models.StatsResponse
	if h.cache != nil {
		if err := h.cache.Get(c.Context(), cacheKey, &cachedStats); err == nil {
			h.logger.Debug("Cache hit for stats")
			return c.JSON(models.NewSuccessResponse(cachedStats, requestID))
		}
	}

	// Cache miss - get from database
	stats, err := h.repo.GetComprehensiveStats(c.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get stats")
		return c.Status(fiber.StatusInternalServerError).JSON(
			models.NewErrorResponse("DATABASE_ERROR", "Failed to retrieve statistics", err.Error(), requestID),
		)
	}

	// Store in cache
	if h.cache != nil {
		if err := h.cache.Set(c.Context(), cacheKey, stats); err != nil {
			h.logger.WithError(err).Warn("Failed to cache stats")
		}
	}

	return c.JSON(models.NewSuccessResponse(stats, requestID))
}

// SearchArticles handles GET /api/v1/articles/search with full-text search
func (h *ArticleHandler) SearchArticles(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	searchQuery := c.Query("q")
	if searchQuery == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewErrorResponse("MISSING_QUERY", "Search query parameter 'q' is required", "", requestID),
		)
	}

	filter := models.ArticleFilter{
		Search:    searchQuery,
		Source:    c.Query("source"),
		Category:  c.Query("category"),
		SortBy:    c.Query("sort_by", "published"),
		SortOrder: c.Query("sort_order", "desc"),
		Limit:     c.QueryInt("limit", 50),
		Offset:    c.QueryInt("offset", 0),
	}

	// Validate limit
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Limit < 1 {
		filter.Limit = 50
	}

	// Search articles
	articles, total, err := h.repo.Search(c.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to search articles")
		return c.Status(fiber.StatusInternalServerError).JSON(
			models.NewErrorResponse("SEARCH_ERROR", "Failed to search articles", err.Error(), requestID),
		)
	}

	meta := &models.Meta{
		Pagination: models.CalculatePaginationMeta(total, filter.Limit, filter.Offset),
		Sorting: &models.SortingMeta{
			SortBy:    filter.SortBy,
			SortOrder: filter.SortOrder,
		},
		Filtering: &models.FilteringMeta{
			Search:   searchQuery,
			Source:   filter.Source,
			Category: filter.Category,
		},
	}

	return c.JSON(models.NewSuccessResponseWithMeta(articles, meta, requestID))
}

// GetCategories handles GET /api/v1/categories
func (h *ArticleHandler) GetCategories(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)
	cacheKey := cache.GenerateKey(cache.PrefixStats, "categories")

	// Try cache first
	var cachedCategories []models.CategoryInfo
	if h.cache != nil {
		if err := h.cache.Get(c.Context(), cacheKey, &cachedCategories); err == nil {
			h.logger.Debug("Cache hit for categories")
			return c.JSON(models.NewSuccessResponse(cachedCategories, requestID))
		}
	}

	// Cache miss - get from database
	categories, err := h.repo.GetCategories(c.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get categories")
		return c.Status(fiber.StatusInternalServerError).JSON(
			models.NewErrorResponse("DATABASE_ERROR", "Failed to retrieve categories", err.Error(), requestID),
		)
	}

	// Store in cache
	if h.cache != nil {
		if err := h.cache.Set(c.Context(), cacheKey, categories); err != nil {
			h.logger.WithError(err).Warn("Failed to cache categories")
		}
	}

	return c.JSON(models.NewSuccessResponse(categories, requestID))
}

// buildFilteringMeta creates filtering metadata for response
func buildFilteringMeta(filter models.ArticleFilter, startDate, endDate string) *models.FilteringMeta {
	meta := &models.FilteringMeta{
		Source:    filter.Source,
		Category:  filter.Category,
		Keyword:   filter.Keyword,
		Search:    filter.Search,
		StartDate: startDate,
		EndDate:   endDate,
	}
	return meta
}

// InvalidateCache clears all article-related cache
func (h *ArticleHandler) InvalidateCache(ctx context.Context) {
	if h.cache == nil {
		return
	}

	patterns := []string{
		cache.PrefixArticles + ":*",
		cache.PrefixStats + ":*",
		cache.PrefixArticle + ":*",
	}

	for _, pattern := range patterns {
		if err := h.cache.DeletePattern(ctx, pattern); err != nil {
			h.logger.WithError(err).Warnf("Failed to invalidate cache pattern: %s", pattern)
		}
	}

	h.logger.Info("Cache invalidated")
}

// ExtractContent handles POST /api/v1/articles/:id/extract-content
func (h *ArticleHandler) ExtractContent(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	// Parse ID
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_ID", "Article ID must be a valid integer", err.Error(), requestID),
		)
	}

	// Check if scraper service is available
	if h.scraperService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(
			models.NewErrorResponse("SERVICE_UNAVAILABLE", "Content extraction service not available", "", requestID),
		)
	}

	h.logger.Infof("Extracting content for article %d", id)

	// Extract content
	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	if err := h.scraperService.EnrichArticleContent(ctx, id); err != nil {
		h.logger.WithError(err).Errorf("Failed to extract content for article %d", id)
		return c.Status(fiber.StatusInternalServerError).JSON(
			models.NewErrorResponse("EXTRACTION_FAILED", "Failed to extract content", err.Error(), requestID),
		)
	}

	// Get updated article
	article, err := h.repo.GetArticleWithContent(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			models.NewErrorResponse("DATABASE_ERROR", "Failed to retrieve article after extraction", err.Error(), requestID),
		)
	}

	// Invalidate cache
	cacheKey := cache.GenerateKey(cache.PrefixArticle, c.Params("id"))
	if h.cache != nil {
		h.cache.Delete(c.Context(), cacheKey)
	}

	response := fiber.Map{
		"success":    true,
		"message":    "Content extracted successfully",
		"characters": len(article.Content),
		"article":    article,
	}

	return c.JSON(models.NewSuccessResponse(response, requestID))
}
