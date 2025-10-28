package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jeffrey/intellinieuws/internal/models"
	"github.com/jeffrey/intellinieuws/internal/scraper"
	"github.com/jeffrey/intellinieuws/pkg/logger"
)

// ScraperHandler handles scraper-related HTTP requests
type ScraperHandler struct {
	scraperService *scraper.Service
	articleHandler *ArticleHandler
	logger         *logger.Logger
}

// NewScraperHandler creates a new scraper handler
func NewScraperHandler(scraperService *scraper.Service, articleHandler *ArticleHandler, log *logger.Logger) *ScraperHandler {
	return &ScraperHandler{
		scraperService: scraperService,
		articleHandler: articleHandler,
		logger:         log.WithComponent("scraper-handler"),
	}
}

// TriggerScrape handles POST /api/v1/scrape
func (h *ScraperHandler) TriggerScrape(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	var req struct {
		Source string `json:"source"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_REQUEST", "Failed to parse request body", err.Error(), requestID),
		)
	}

	// If source is provided, scrape single source
	if req.Source != "" {
		feedURL, exists := scraper.ScrapeSources[req.Source]
		if !exists {
			return c.Status(fiber.StatusBadRequest).JSON(
				models.NewErrorResponse("INVALID_SOURCE", "Source not found", "Available sources: nu.nl, ad.nl, nos.nl", requestID),
			)
		}

		h.logger.Infof("Triggering scrape for source: %s", req.Source)
		result, err := h.scraperService.ScrapeWithRetry(c.Context(), req.Source, feedURL)
		if err != nil {
			h.logger.WithError(err).Errorf("Scrape failed for source: %s", req.Source)
			return c.Status(fiber.StatusInternalServerError).JSON(
				models.NewErrorResponse("SCRAPING_FAILED", "Failed to scrape source", err.Error(), requestID),
			)
		}

		// Invalidate cache after successful scrape
		if result.ArticlesStored > 0 && h.articleHandler != nil {
			h.articleHandler.InvalidateCache(c.Context())
		}

		response := fiber.Map{
			"status":           "success",
			"source":           result.Source,
			"articles_found":   result.ArticlesFound,
			"articles_stored":  result.ArticlesStored,
			"articles_skipped": result.ArticlesSkipped,
			"duration_seconds": result.Duration.Seconds(),
		}

		return c.JSON(models.NewSuccessResponse(response, requestID))
	}

	// Scrape all sources
	h.logger.Info("Triggering scrape for all sources")
	results, err := h.scraperService.ScrapeAllSources(c.Context())
	if err != nil {
		h.logger.WithError(err).Error("Scrape failed for all sources")
		return c.Status(fiber.StatusInternalServerError).JSON(
			models.NewErrorResponse("SCRAPING_FAILED", "Failed to scrape all sources", err.Error(), requestID),
		)
	}

	// Format results
	formattedResults := make([]fiber.Map, 0, len(results))
	for _, result := range results {
		formattedResults = append(formattedResults, fiber.Map{
			"source":           result.Source,
			"status":           result.Status,
			"articles_found":   result.ArticlesFound,
			"articles_stored":  result.ArticlesStored,
			"articles_skipped": result.ArticlesSkipped,
			"duration_seconds": result.Duration.Seconds(),
			"error":            result.Error,
		})
	}

	// Invalidate cache after scraping (if any articles were stored)
	totalStored := 0
	for _, result := range results {
		totalStored += result.ArticlesStored
	}
	if totalStored > 0 && h.articleHandler != nil {
		h.articleHandler.InvalidateCache(c.Context())
	}

	response := fiber.Map{
		"total_sources": len(results),
		"total_stored":  totalStored,
		"results":       formattedResults,
	}

	return c.JSON(models.NewSuccessResponse(response, requestID))
}

// GetSources handles GET /api/v1/sources
func (h *ScraperHandler) GetSources(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	sources := make([]models.SourceInfo, 0, len(scraper.ScrapeSources))
	for source, feedURL := range scraper.ScrapeSources {
		sources = append(sources, models.SourceInfo{
			Name:     source,
			FeedURL:  feedURL,
			IsActive: true,
		})
	}

	return c.JSON(models.NewSuccessResponse(sources, requestID))
}

// GetScraperStats handles GET /api/v1/scraper/stats
func (h *ScraperHandler) GetScraperStats(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	stats, err := h.scraperService.GetStats(c.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get scraper stats")
		return c.Status(fiber.StatusInternalServerError).JSON(
			models.NewErrorResponse("DATABASE_ERROR", "Failed to retrieve scraper statistics", err.Error(), requestID),
		)
	}

	return c.JSON(models.NewSuccessResponse(stats, requestID))
}
