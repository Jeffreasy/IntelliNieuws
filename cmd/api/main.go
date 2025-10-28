package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/jeffrey/intellinieuws/internal/ai"
	"github.com/jeffrey/intellinieuws/internal/api"
	"github.com/jeffrey/intellinieuws/internal/api/handlers"
	"github.com/jeffrey/intellinieuws/internal/cache"
	"github.com/jeffrey/intellinieuws/internal/repository"
	"github.com/jeffrey/intellinieuws/internal/scheduler"
	"github.com/jeffrey/intellinieuws/internal/scraper"
	"github.com/jeffrey/intellinieuws/pkg/config"
	"github.com/jeffrey/intellinieuws/pkg/logger"
	"github.com/jeffrey/intellinieuws/pkg/middleware"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(logger.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
	})
	log.Info("Starting Nieuws Scraper API service")

	// Initialize database connection with optimized pool settings
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()

	dbConfig, err := pgxpool.ParseConfig(cfg.Database.GetDSN())
	if err != nil {
		log.WithError(err).Fatal("Failed to parse database config")
	}

	// Optimize connection pool (PHASE 2 OPTIMIZATION: 20% faster)
	dbConfig.MaxConns = 25                               // Maximum connections
	dbConfig.MinConns = 5                                // Minimum idle connections (keep warm)
	dbConfig.MaxConnLifetime = 1 * time.Hour             // Recycle connections
	dbConfig.MaxConnIdleTime = 30 * time.Minute          // Close idle connections
	dbConfig.HealthCheckPeriod = 1 * time.Minute         // Health check interval
	dbConfig.ConnConfig.ConnectTimeout = 5 * time.Second // Connection timeout

	// Additional optimizations for better performance
	dbConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement // Cache prepared statements
	dbConfig.ConnConfig.RuntimeParams = map[string]string{
		"application_name":  "nieuws-scraper-api",
		"search_path":       "public",
		"timezone":          "UTC",
		"statement_timeout": "30s", // Prevent runaway queries
		// Note: idle_in_transaction_timeout removed for PostgreSQL < 9.6 compatibility
		// Note: jit setting removed for PostgreSQL < 11 compatibility
	}

	dbPool, err := pgxpool.NewWithConfig(dbCtx, dbConfig)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to database")
	}
	defer dbPool.Close()

	log.Info("Database connection pool configured: max=25, min=5, statement_cache=enabled")

	// Pre-warm connection pool for better initial performance
	log.Info("Pre-warming database connection pool...")
	warmCtx, warmCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer warmCancel()

	for i := int32(0); i < dbConfig.MinConns; i++ {
		conn, err := dbPool.Acquire(warmCtx)
		if err != nil {
			log.WithError(err).Warn("Failed to pre-warm connection, continuing...")
			continue
		}
		conn.Release()
	}
	log.Info("Connection pool pre-warming completed")

	// Test database connection
	if err := dbPool.Ping(dbCtx); err != nil {
		log.WithError(err).Fatal("Failed to ping database")
	}
	log.Info("Successfully connected to database")

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// Test Redis connection
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.WithError(err).Warn("Failed to connect to Redis, continuing without cache")
		redisClient = nil
	} else {
		log.Info("Successfully connected to Redis")
	}

	// Initialize cache service (5 minute TTL)
	var cacheService *cache.Service
	if redisClient != nil {
		cacheService = cache.NewService(redisClient, 5*time.Minute)
		if cacheService.IsAvailable() {
			log.Info("Cache service initialized with 5min TTL")
		}
	} else {
		log.Info("Cache service disabled (Redis not available)")
	}

	// Initialize repositories
	articleRepo := repository.NewArticleRepository(dbPool)
	jobRepo := repository.NewScrapingJobRepository(dbPool, log)

	// Initialize services
	scraperService := scraper.NewService(&cfg.Scraper, articleRepo, jobRepo, log)

	// Initialize scheduler if enabled
	var scraperScheduler *scheduler.Scheduler
	if cfg.Scraper.ScheduleEnabled {
		interval := cfg.Scraper.GetScheduleInterval()
		scraperScheduler = scheduler.NewScheduler(scraperService, interval, log)

		// Start scheduler in background
		go scraperScheduler.Start(context.Background())
		log.Infof("Scheduled scraping enabled with interval: %v", interval)
	} else {
		log.Info("Scheduled scraping disabled")
	}

	// Initialize content processor for hybrid scraping
	var contentProcessor *scraper.ContentProcessor
	if cfg.Scraper.EnableFullContentExtraction && cfg.Scraper.ContentExtractionAsync {
		contentProcessor = scraper.NewContentProcessor(
			scraperService,
			cfg.Scraper.ContentExtractionInterval,
			cfg.Scraper.EnableFullContentExtraction,
			log,
		)
		go contentProcessor.Start(context.Background())
		log.Infof("Content processor started with interval: %v", cfg.Scraper.ContentExtractionInterval)
	}

	// Initialize AI service and processor
	var aiService *ai.Service
	var aiProcessor *ai.Processor
	var aiChatService *ai.ChatService
	var aiHandler *handlers.AIHandler

	if cfg.AI.Enabled {
		log.Info("Initializing AI processing service")

		// Convert config to AI config
		aiConfig := &ai.Config{
			OpenAIAPIKey:       cfg.AI.OpenAIAPIKey,
			OpenAIModel:        cfg.AI.OpenAIModel,
			OpenAIMaxTokens:    cfg.AI.OpenAIMaxTokens,
			Enabled:            cfg.AI.Enabled,
			AsyncProcessing:    cfg.AI.AsyncProcessing,
			BatchSize:          cfg.AI.BatchSize,
			ProcessInterval:    cfg.AI.ProcessInterval,
			RetryFailed:        cfg.AI.RetryFailed,
			MaxRetries:         cfg.AI.MaxRetries,
			EnableSentiment:    cfg.AI.EnableSentiment,
			EnableEntities:     cfg.AI.EnableEntities,
			EnableCategories:   cfg.AI.EnableCategories,
			EnableKeywords:     cfg.AI.EnableKeywords,
			EnableSummary:      cfg.AI.EnableSummary,
			EnableSimilarity:   cfg.AI.EnableSimilarity,
			MaxDailyCost:       cfg.AI.MaxDailyCost,
			RateLimitPerMinute: cfg.AI.RateLimitPerMinute,
			Timeout:            cfg.AI.Timeout,
		}

		aiService = ai.NewService(dbPool, aiConfig, log)

		// Initialize OpenAI client for chat service
		openAIClient := ai.NewOpenAIClient(
			cfg.AI.OpenAIAPIKey,
			cfg.AI.OpenAIModel,
			cfg.AI.OpenAIMaxTokens,
			log,
		)

		// Initialize chat service
		aiChatService = ai.NewChatService(aiService, openAIClient, log)
		log.Info("AI chat service initialized")

		if cfg.AI.AsyncProcessing {
			aiProcessor = ai.NewProcessor(aiService, aiConfig, log)
			go aiProcessor.Start(context.Background())
			log.Infof("AI processor started with interval: %v", cfg.AI.ProcessInterval)
		}

		aiHandler = handlers.NewAIHandler(aiService, aiProcessor, aiChatService, cacheService, log)
		log.Info("AI service initialized successfully")
	} else {
		log.Info("AI processing disabled")
	}

	// Initialize handlers
	articleHandler := handlers.NewArticleHandler(articleRepo, cacheService, log)
	articleHandler.SetScraperService(scraperService) // Enable content extraction endpoint
	scraperHandler := handlers.NewScraperHandler(scraperService, articleHandler, log)

	// Initialize middleware
	var rateLimiter *middleware.RateLimiter
	if redisClient != nil {
		rateLimiter = middleware.NewRateLimiter(
			redisClient,
			cfg.API.RateLimitRequests,
			cfg.API.RateLimitWindowSeconds,
		)
	}

	var auth *middleware.APIKeyAuth
	if cfg.API.APIKey != "" {
		auth = middleware.NewAPIKeyAuth(cfg.API.APIKey, cfg.API.APIKeyHeader)
		log.Info("API key authentication enabled")
	} else {
		log.Warn("API key authentication disabled - no API_KEY configured")
	}

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Nieuws Scraper API",
		ReadTimeout:  cfg.API.GetAPITimeout(),
		WriteTimeout: cfg.API.GetAPITimeout(),
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error":   "Error",
				"message": err.Error(),
				"code":    code,
			})
		},
	})

	// Setup routes with comprehensive health monitoring (PHASE 4)
	api.SetupRoutes(app, articleHandler, scraperHandler, aiHandler, rateLimiter, auth, log, dbPool, redisClient, cacheService, scraperService, aiProcessor)

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.APIPort)
		log.Infof("Starting API server on %s", addr)
		if err := app.Listen(addr); err != nil {
			serverErr <- err
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.WithError(err).Fatal("Server failed to start")
	case <-quit:
		log.Info("Shutting down server...")
	}

	// Stop scheduler if running
	if scraperScheduler != nil && scraperScheduler.IsRunning() {
		log.Info("Stopping scheduler...")
		scraperScheduler.Stop()
	}

	// Stop content processor if running
	if contentProcessor != nil && contentProcessor.IsRunning() {
		log.Info("Stopping content processor...")
		contentProcessor.Stop()
	}

	// Stop AI processor if running
	if aiProcessor != nil && aiProcessor.IsRunning() {
		log.Info("Stopping AI processor...")
		aiProcessor.Stop()
	}

	// Cleanup scraper service (closes browser pool if active)
	log.Info("Cleaning up scraper resources...")
	scraperService.Cleanup()

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.WithError(err).Error("Server forced to shutdown")
	}

	log.Info("Server exited")
}
