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
	emailHandler *handlers.EmailHandler,
	cacheHandler *handlers.CacheHandler,
	configHandler *handlers.ConfigHandler,
	rateLimiter *middleware.RateLimiter,
	auth *middleware.APIKeyAuth,
	log *logger.Logger,
	db *pgxpool.Pool,
	redis *redis.Client,
	cacheService *cache.Service,
	scraperService *scraper.Service,
	aiProcessor *ai.Processor,
) {
	// Initialize analytics handler
	analyticsHandler := handlers.NewAnalyticsHandler(db, log)

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

	// Analytics routes (public, no auth) - Must be before auth middleware
	analytics := api.Group("/analytics")
	analytics.Get("/trending", analyticsHandler.GetTrendingKeywords)
	analytics.Get("/sentiment-trends", analyticsHandler.GetSentimentTrends)
	analytics.Get("/hot-entities", analyticsHandler.GetHotEntities)
	analytics.Get("/entity-sentiment", analyticsHandler.GetEntitySentiment)
	analytics.Get("/overview", analyticsHandler.GetAnalyticsOverview)
	analytics.Get("/article-stats", analyticsHandler.GetArticleStats)
	analytics.Get("/maintenance-schedule", analyticsHandler.GetMaintenanceSchedule)
	analytics.Get("/database-health", analyticsHandler.GetDatabaseHealth)
	analytics.Post("/refresh", analyticsHandler.RefreshAnalytics)

	// Configuration routes (public read, protected write) - Must be before auth middleware
	if configHandler != nil {
		config := api.Group("/config")

		// Public read endpoints
		config.Get("/profiles", configHandler.GetProfiles)                // Get all profiles
		config.Get("/current", configHandler.GetCurrentConfig)            // Get current config
		config.Get("/scheduler/status", configHandler.GetSchedulerStatus) // Scheduler status
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

	// Stock ticker routes (public) - FMP Free Tier Only
	// Note: Many advanced features require FMP premium subscription ($14/month)
	if stockHandler != nil {
		stocks := api.Group("/stocks")

		// ✅ FREE TIER - Core endpoints (US stocks only)
		stocks.Get("/quote/:symbol", stockHandler.GetQuote)       // Single quote (US stocks)
		stocks.Get("/profile/:symbol", stockHandler.GetProfile)   // Basic profile
		stocks.Get("/earnings", stockHandler.GetEarningsCalendar) // Earnings calendar
		stocks.Get("/search", stockHandler.SearchSymbol)          // Symbol search
		stocks.Get("/stats", stockHandler.GetStats)               // Cache stats

		// ⚠️ PREMIUM FEATURES - Disabled for free tier
		// Uncomment these if you upgrade to FMP Starter plan ($14/month)
		// stocks.Post("/quotes", stockHandler.GetMultipleQuotes)        // Batch quotes
		// stocks.Get("/news/:symbol", stockHandler.GetStockNews)        // Stock news
		// stocks.Get("/historical/:symbol", stockHandler.GetHistoricalPrices)  // Historical
		// stocks.Get("/metrics/:symbol", stockHandler.GetKeyMetrics)    // Metrics
		// market := stocks.Group("/market")
		// market.Get("/gainers", stockHandler.GetMarketGainers)         // Gainers
		// market.Get("/losers", stockHandler.GetMarketLosers)           // Losers
		// market.Get("/actives", stockHandler.GetMostActives)           // Actives
		// stocks.Get("/sectors", stockHandler.GetSectorPerformance)     // Sectors
		// stocks.Get("/ratings/:symbol", stockHandler.GetAnalystRatings) // Ratings
		// stocks.Get("/target/:symbol", stockHandler.GetPriceTarget)    // Targets
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

	// Email integration routes (protected)
	if emailHandler != nil {
		emails := protected.Group("/email")
		emails.Post("/fetch-existing", emailHandler.FetchExistingEmails) // Manually fetch existing emails
		emails.Get("/stats", emailHandler.GetStats)                      // Email processing stats
	}

	// Scraper routes (protected)
	protected.Post("/scrape", scraperHandler.TriggerScrape)
	protected.Get("/scraper/stats", scraperHandler.GetScraperStats)

	// AI processing routes (protected)
	if aiHandler != nil {
		protected.Post("/articles/:id/process", aiHandler.ProcessArticle)
		protected.Post("/ai/process/trigger", aiHandler.TriggerProcessing)
	}

	// Cache management routes (protected)
	if cacheHandler != nil {
		cacheRoutes := protected.Group("/cache")
		cacheRoutes.Get("/stats", cacheHandler.GetStatistics)         // Cache statistics
		cacheRoutes.Get("/keys", cacheHandler.GetCacheKeys)           // List cache keys
		cacheRoutes.Get("/size", cacheHandler.GetCacheSize)           // Total cache size
		cacheRoutes.Get("/memory", cacheHandler.GetMemoryUsage)       // Memory usage
		cacheRoutes.Post("/invalidate", cacheHandler.InvalidateCache) // Invalidate cache
		cacheRoutes.Post("/warm", cacheHandler.WarmCache)             // Warm cache
	}

	// Configuration write routes (protected)
	if configHandler != nil {
		configProtected := protected.Group("/config")
		configProtected.Post("/profile/:name", configHandler.SwitchProfile) // Switch profile
		configProtected.Patch("/setting", configHandler.UpdateSetting)      // Update setting
		configProtected.Post("/reset", configHandler.ResetToDefaults)       // Reset to defaults
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
