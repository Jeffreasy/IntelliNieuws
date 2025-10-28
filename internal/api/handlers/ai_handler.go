package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jeffrey/intellinieuws/internal/ai"
	"github.com/jeffrey/intellinieuws/internal/cache"
	"github.com/jeffrey/intellinieuws/internal/models"
	"github.com/jeffrey/intellinieuws/pkg/logger"
)

// AIHandler handles AI-related API requests
type AIHandler struct {
	aiService   *ai.Service
	processor   *ai.Processor
	chatService *ai.ChatService
	cache       *cache.Service
	logger      *logger.Logger
}

// NewAIHandler creates a new AI handler
func NewAIHandler(aiService *ai.Service, processor *ai.Processor, chatService *ai.ChatService, cacheService *cache.Service, log *logger.Logger) *AIHandler {
	return &AIHandler{
		aiService:   aiService,
		processor:   processor,
		chatService: chatService,
		cache:       cacheService,
		logger:      log.WithComponent("ai-handler"),
	}
}

// GetEnrichment returns AI enrichment for a specific article
// GET /api/v1/articles/:id/enrichment
func (h *AIHandler) GetEnrichment(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	articleID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_ID", "Invalid article ID", err.Error(), requestID),
		)
	}

	// Try cache first (OPTIMIZED)
	cacheKey := cache.GenerateKey(cache.PrefixAIEnrichment, c.Params("id"))
	var enrichment *ai.AIEnrichment

	if h.cache != nil && h.cache.IsAvailable() {
		if err := h.cache.Get(c.Context(), cacheKey, &enrichment); err == nil {
			h.logger.Debugf("Cache HIT for enrichment %d", articleID)
			return c.JSON(models.NewSuccessResponse(enrichment, requestID))
		}
	}

	enrichment, err = h.aiService.GetEnrichment(c.Context(), articleID)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to get enrichment for article %d", articleID)
		return c.Status(fiber.StatusInternalServerError).JSON(
			models.NewErrorResponse("DATABASE_ERROR", "Failed to retrieve enrichment", err.Error(), requestID),
		)
	}

	// Cache the result
	if h.cache != nil && h.cache.IsAvailable() {
		if err := h.cache.Set(c.Context(), cacheKey, enrichment); err != nil {
			h.logger.WithError(err).Warn("Failed to cache enrichment")
		}
	}

	return c.JSON(models.NewSuccessResponse(enrichment, requestID))
}

// ProcessArticle triggers AI processing for a specific article
// POST /api/v1/articles/:id/process
func (h *AIHandler) ProcessArticle(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	articleID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_ID", "Invalid article ID", err.Error(), requestID),
		)
	}

	enrichment, err := h.aiService.ProcessArticle(c.Context(), articleID)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to process article %d", articleID)
		return c.Status(fiber.StatusInternalServerError).JSON(
			models.NewErrorResponse("PROCESSING_ERROR", "Failed to process article", err.Error(), requestID),
		)
	}

	// Invalidate cache after processing
	if h.cache != nil && h.cache.IsAvailable() {
		cacheKey := cache.GenerateKey(cache.PrefixAIEnrichment, c.Params("id"))
		h.cache.Delete(c.Context(), cacheKey)
	}

	response := map[string]interface{}{
		"message":    "Article processed successfully",
		"article_id": articleID,
		"enrichment": enrichment,
	}

	return c.JSON(models.NewSuccessResponse(response, requestID))
}

// GetSentimentStats returns sentiment statistics
// GET /api/v1/ai/sentiment/stats
func (h *AIHandler) GetSentimentStats(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)
	source := c.Query("source")

	var startDate, endDate *time.Time
	if startStr := c.Query("start_date"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startDate = &t
		}
	}
	if endStr := c.Query("end_date"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endDate = &t
		}
	}

	// Try cache first (OPTIMIZED: 60-80% load reduction)
	cacheKey := cache.GenerateKey(cache.PrefixAISentiment, source, c.Query("start_date"), c.Query("end_date"))
	var stats *ai.SentimentStats

	if h.cache != nil && h.cache.IsAvailable() {
		if err := h.cache.Get(c.Context(), cacheKey, &stats); err == nil {
			h.logger.Debugf("Cache HIT for sentiment stats")
			return c.JSON(models.NewSuccessResponse(stats, requestID))
		}
	}

	stats, err := h.aiService.GetSentimentStats(c.Context(), source, startDate, endDate)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get sentiment stats")
		return c.Status(fiber.StatusInternalServerError).JSON(
			models.NewErrorResponse("DATABASE_ERROR", "Failed to retrieve sentiment statistics", err.Error(), requestID),
		)
	}

	// Cache the result (5 minutes TTL for stats)
	if h.cache != nil && h.cache.IsAvailable() {
		if err := h.cache.Set(c.Context(), cacheKey, stats); err != nil {
			h.logger.WithError(err).Warn("Failed to cache sentiment stats")
		}
	}

	return c.JSON(models.NewSuccessResponse(stats, requestID))
}

// GetTrendingTopics returns trending topics
// GET /api/v1/ai/trending
func (h *AIHandler) GetTrendingTopics(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	hoursBack := c.QueryInt("hours", 24)
	if hoursBack <= 0 {
		hoursBack = 24
	}

	minArticles := c.QueryInt("min_articles", 3)
	if minArticles <= 0 {
		minArticles = 3
	}

	// Try cache first (OPTIMIZED: This is an expensive query)
	cacheKey := cache.GenerateKey(cache.PrefixAITrending,
		fmt.Sprintf("h%d", hoursBack),
		fmt.Sprintf("m%d", minArticles))

	type cachedResponse struct {
		Topics      []ai.TrendingTopic `json:"topics"`
		HoursBack   int                `json:"hours_back"`
		MinArticles int                `json:"min_articles"`
		Count       int                `json:"count"`
	}

	var cached cachedResponse
	if h.cache != nil && h.cache.IsAvailable() {
		if err := h.cache.Get(c.Context(), cacheKey, &cached); err == nil {
			h.logger.Debugf("Cache HIT for trending topics")
			return c.JSON(models.NewSuccessResponse(cached, requestID))
		}
	}

	topics, err := h.aiService.GetTrendingTopics(c.Context(), hoursBack, minArticles)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get trending topics")
		return c.Status(fiber.StatusInternalServerError).JSON(
			models.NewErrorResponse("DATABASE_ERROR", "Failed to retrieve trending topics", err.Error(), requestID),
		)
	}

	response := cachedResponse{
		Topics:      topics,
		HoursBack:   hoursBack,
		MinArticles: minArticles,
		Count:       len(topics),
	}

	// Cache the result (2 minutes TTL for trending topics)
	if h.cache != nil && h.cache.IsAvailable() {
		if err := h.cache.Set(c.Context(), cacheKey, response); err != nil {
			h.logger.WithError(err).Warn("Failed to cache trending topics")
		}
	}

	return c.JSON(models.NewSuccessResponse(response, requestID))
}

// GetArticlesByEntity returns articles mentioning a specific entity
// GET /api/v1/ai/entity/:name
func (h *AIHandler) GetArticlesByEntity(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)
	entityName := c.Params("name")

	if entityName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewErrorResponse("MISSING_PARAMETER", "Entity name is required", "", requestID),
		)
	}

	entityType := c.Query("type") // persons, organizations, locations
	limit := c.QueryInt("limit", 50)
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 50
	}

	// Try cache first (OPTIMIZED)
	cacheKey := cache.GenerateKey(cache.PrefixAIEntity, entityName, entityType, fmt.Sprintf("l%d", limit))
	var articles []models.Article

	if h.cache != nil && h.cache.IsAvailable() {
		if err := h.cache.Get(c.Context(), cacheKey, &articles); err == nil {
			h.logger.Debugf("Cache HIT for entity %s", entityName)

			meta := &models.Meta{
				Pagination: models.CalculatePaginationMeta(len(articles), limit, 0),
				Filtering: &models.FilteringMeta{
					Search: entityName,
				},
			}
			return c.JSON(models.NewSuccessResponseWithMeta(articles, meta, requestID))
		}
	}

	articles, err := h.aiService.GetArticlesByEntity(c.Context(), entityName, entityType, limit)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to get articles for entity %s", entityName)
		return c.Status(fiber.StatusInternalServerError).JSON(
			models.NewErrorResponse("DATABASE_ERROR", "Failed to retrieve articles", err.Error(), requestID),
		)
	}

	// Cache the result
	if h.cache != nil && h.cache.IsAvailable() {
		if err := h.cache.Set(c.Context(), cacheKey, articles); err != nil {
			h.logger.WithError(err).Warn("Failed to cache entity articles")
		}
	}

	meta := &models.Meta{
		Pagination: models.CalculatePaginationMeta(len(articles), limit, 0),
		Filtering: &models.FilteringMeta{
			Search: entityName,
		},
	}

	return c.JSON(models.NewSuccessResponseWithMeta(articles, meta, requestID))
}

// GetArticlesByTicker returns articles mentioning a specific stock ticker
// GET /api/v1/articles/by-ticker/:symbol
func (h *AIHandler) GetArticlesByTicker(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)
	symbol := c.Params("symbol")

	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewErrorResponse("MISSING_PARAMETER", "Stock symbol is required", "", requestID),
		)
	}

	limit := c.QueryInt("limit", 50)
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 50
	}

	// Try cache first (OPTIMIZED)
	cacheKey := cache.GenerateKey(cache.PrefixAIEntity, "ticker", symbol, fmt.Sprintf("l%d", limit))
	var articles []models.Article

	if h.cache != nil && h.cache.IsAvailable() {
		if err := h.cache.Get(c.Context(), cacheKey, &articles); err == nil {
			h.logger.Debugf("Cache HIT for stock ticker %s", symbol)

			meta := &models.Meta{
				Pagination: models.CalculatePaginationMeta(len(articles), limit, 0),
				Filtering: &models.FilteringMeta{
					Search: symbol,
				},
			}
			return c.JSON(models.NewSuccessResponseWithMeta(articles, meta, requestID))
		}
	}

	articles, err := h.aiService.GetArticlesByStockTicker(c.Context(), symbol, limit)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to get articles for stock ticker %s", symbol)
		return c.Status(fiber.StatusInternalServerError).JSON(
			models.NewErrorResponse("DATABASE_ERROR", "Failed to retrieve articles", err.Error(), requestID),
		)
	}

	// Cache the result
	if h.cache != nil && h.cache.IsAvailable() {
		if err := h.cache.Set(c.Context(), cacheKey, articles); err != nil {
			h.logger.WithError(err).Warn("Failed to cache ticker articles")
		}
	}

	meta := &models.Meta{
		Pagination: models.CalculatePaginationMeta(len(articles), limit, 0),
		Filtering: &models.FilteringMeta{
			Search: symbol,
		},
	}

	return c.JSON(models.NewSuccessResponseWithMeta(articles, meta, requestID))
}

// TriggerProcessing manually triggers AI processing
// POST /api/v1/ai/process/trigger
func (h *AIHandler) TriggerProcessing(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	if h.processor == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(
			models.NewErrorResponse("SERVICE_UNAVAILABLE", "AI processor not available", "", requestID),
		)
	}

	result, err := h.processor.ManualTrigger(c.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to trigger processing")
		return c.Status(fiber.StatusInternalServerError).JSON(
			models.NewErrorResponse("PROCESSING_ERROR", "Failed to trigger processing", err.Error(), requestID),
		)
	}

	// Invalidate all AI caches after batch processing
	if h.cache != nil && h.cache.IsAvailable() {
		h.cache.DeletePattern(c.Context(), cache.PrefixAITrending+"*")
		h.cache.DeletePattern(c.Context(), cache.PrefixAISentiment+"*")
	}

	response := map[string]interface{}{
		"message":         "Processing completed",
		"total_processed": result.TotalProcessed,
		"success_count":   result.SuccessCount,
		"failure_count":   result.FailureCount,
		"duration":        result.Duration.String(),
	}

	return c.JSON(models.NewSuccessResponse(response, requestID))
}

// GetProcessorStats returns processor statistics
// GET /api/v1/ai/processor/stats
func (h *AIHandler) GetProcessorStats(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	if h.processor == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(
			models.NewErrorResponse("SERVICE_UNAVAILABLE", "AI processor not available", "", requestID),
		)
	}

	stats := h.processor.GetStats()
	return c.JSON(models.NewSuccessResponse(stats, requestID))
}

// Chat handles conversational AI requests
// POST /api/v1/ai/chat
func (h *AIHandler) Chat(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	if h.chatService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(
			models.NewErrorResponse("SERVICE_UNAVAILABLE", "Chat service not available", "", requestID),
		)
	}

	// Parse request
	var req ai.ChatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_REQUEST", "Invalid request body", err.Error(), requestID),
		)
	}

	// Validate message
	if req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewErrorResponse("MISSING_MESSAGE", "Message is required", "", requestID),
		)
	}

	if len(req.Message) > 1000 {
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewErrorResponse("MESSAGE_TOO_LONG", "Message must be less than 1000 characters", "", requestID),
		)
	}

	// Try cache first (for same questions)
	cacheKey := cache.GenerateKey(cache.PrefixAIEnrichment, "chat", req.Message)
	var response *ai.ChatResponse

	if h.cache != nil && h.cache.IsAvailable() {
		if err := h.cache.Get(c.Context(), cacheKey, &response); err == nil {
			h.logger.Debugf("Cache HIT for chat message")
			return c.JSON(models.NewSuccessResponse(response, requestID))
		}
	}

	// Process chat message with optional article context
	response, err := h.chatService.ProcessChatMessageWithContext(c.Context(), req.Message, req.Context, req.ArticleContent, req.ArticleID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to process chat message")
		return c.Status(fiber.StatusInternalServerError).JSON(
			models.NewErrorResponse("PROCESSING_ERROR", "Failed to process message", err.Error(), requestID),
		)
	}

	// Cache the response
	if h.cache != nil && h.cache.IsAvailable() {
		if err := h.cache.Set(c.Context(), cacheKey, response); err != nil {
			h.logger.WithError(err).Warn("Failed to cache chat response")
		}
	}

	return c.JSON(models.NewSuccessResponse(response, requestID))
}
