package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/jeffrey/intellinieuws/internal/ai"
	"github.com/jeffrey/intellinieuws/internal/api/handlers"
	"github.com/jeffrey/intellinieuws/internal/cache"
	"github.com/jeffrey/intellinieuws/internal/models"
	"github.com/jeffrey/intellinieuws/internal/scraper"
	"github.com/jeffrey/intellinieuws/pkg/logger"
	"github.com/jeffrey/intellinieuws/pkg/middleware"
)

// SetupRoutes configures all API routes
func SetupRoutes(
	app *fiber.App,
	articleHandler *handlers.ArticleHandler,
	scraperHandler *handlers.ScraperHandler,
	aiHandler *handlers.AIHandler,
	stockHandler *handlers.StockHandler,
	rateLimiter *middleware.RateLimiter,
	auth *middleware.APIKeyAuth,
	log *logger.Logger,
	db *pgxpool.Pool,
	redis *redis.Client,
	cacheService *cache.Service,
	scraperService *scraper.Service,
	aiProcessor *ai.Processor,
) {
	// Global middleware
	app.Use(recover.New())
	app.Use(requestid.New())

	// Enhanced CORS configuration for frontend
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-API-Key, X-Request-ID",
		AllowCredentials: false,
		ExposeHeaders:    "X-Request-ID, X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset",
		MaxAge:           300,
	}))

	// Custom logger middleware
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		requestID := c.Locals("requestid").(string)

		err := c.Next()

		duration := time.Since(start)
		log.Infof("[%s] %s %s - %d - %v",
			requestID,
			c.Method(),
			c.Path(),
			c.Response().StatusCode(),
			duration,
		)

		return err
	})

	// PHASE 4: Comprehensive health monitoring (no auth required)
	healthHandler := handlers.NewHealthHandler(db, redis, cacheService, scraperService, aiProcessor, log)

	app.Get("/health", healthHandler.GetHealth)          // Comprehensive health
	app.Get("/health/live", healthHandler.GetLiveness)   // Liveness probe
	app.Get("/health/ready", healthHandler.GetReadiness) // Readiness probe
	app.Get("/health/metrics", healthHandler.GetMetrics) // Detailed metrics

	// API v1 routes
	api := app.Group("/api/v1")

	// Apply rate limiting to all API routes
	if rateLimiter != nil {
		api.Use(rateLimiter.Handler())
	}

	// Public routes (optional auth)
	if auth != nil {
		api.Use(auth.Optional())
	}

	// Article routes
	articles := api.Group("/articles")
	articles.Get("/", articleHandler.ListArticles)
	articles.Get("/stats", articleHandler.GetStats)
	articles.Get("/search", articleHandler.SearchArticles)
	articles.Get("/:id", articleHandler.GetArticle)

	// Content extraction route (protected)
	if auth != nil {
		articles.Post("/:id/extract-content", auth.Handler(), articleHandler.ExtractContent)
	} else {
		articles.Post("/:id/extract-content", articleHandler.ExtractContent)
	}

	// AI enrichment routes (public)
	if aiHandler != nil {
		articles.Get("/:id/enrichment", aiHandler.GetEnrichment)
	}

	// Source routes
	api.Get("/sources", scraperHandler.GetSources)
	api.Get("/categories", articleHandler.GetCategories)

	// AI analytics routes (public)
	if aiHandler != nil {
		ai := api.Group("/ai")
		ai.Get("/sentiment/stats", aiHandler.GetSentimentStats)
		ai.Get("/trending", aiHandler.GetTrendingTopics)
		ai.Get("/entity/:name", aiHandler.GetArticlesByEntity)
		ai.Get("/processor/stats", aiHandler.GetProcessorStats)

		// Conversational AI chat endpoint (public)
		ai.Post("/chat", aiHandler.Chat)
	}

	// Stock ticker routes (public)
	if stockHandler != nil {
		stocks := api.Group("/stocks")
		stocks.Get("/quote/:symbol", stockHandler.GetQuote)
		stocks.Post("/quotes", stockHandler.GetMultipleQuotes)
		stocks.Get("/profile/:symbol", stockHandler.GetProfile)
		stocks.Get("/stats", stockHandler.GetStats)
	}

	// Stock ticker article routes (public)
	if aiHandler != nil {
		articles.Get("/by-ticker/:symbol", aiHandler.GetArticlesByTicker)
	}

	// Protected routes (requires auth)
	protected := api.Group("")
	if auth != nil {
		protected.Use(auth.Handler())
	}

	// Scraper routes (protected)
	protected.Post("/scrape", scraperHandler.TriggerScrape)
	protected.Get("/scraper/stats", scraperHandler.GetScraperStats)

	// AI processing routes (protected)
	if aiHandler != nil {
		protected.Post("/articles/:id/process", aiHandler.ProcessArticle)
		protected.Post("/ai/process/trigger", aiHandler.TriggerProcessing)
	}

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		requestID := c.Locals("requestid").(string)
		return c.Status(fiber.StatusNotFound).JSON(
			models.NewErrorResponse(
				"NOT_FOUND",
				"The requested resource was not found",
				c.Path(),
				requestID,
			),
		)
	})
}
