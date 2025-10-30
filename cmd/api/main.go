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
	"github.com/jeffrey/intellinieuws/internal/email"
	"github.com/jeffrey/intellinieuws/internal/repository"
	"github.com/jeffrey/intellinieuws/internal/scheduler"
	"github.com/jeffrey/intellinieuws/internal/scraper"
	"github.com/jeffrey/intellinieuws/internal/stock"
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

	// Initialize Redis client with connection pooling (from config)
	redisClient := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.GetRedisAddr(),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,     // From config (default: 20)
		MinIdleConns: cfg.Redis.MinIdleConns, // From config (default: 5)
		MaxRetries:   3,                      // Maximum number of retries before giving up
		DialTimeout:  5 * time.Second,        // Dial timeout
		ReadTimeout:  3 * time.Second,        // Timeout for socket reads
		WriteTimeout: 3 * time.Second,        // Timeout for socket writes
		PoolTimeout:  4 * time.Second,        // Amount of time client waits for connection
		// Connection age configuration
		ConnMaxLifetime: 30 * time.Minute, // Maximum connection age
		ConnMaxIdleTime: 5 * time.Minute,  // Close idle connections after this duration
	})
	defer redisClient.Close()

	// Test Redis connection
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.WithError(err).Warn("Failed to connect to Redis, continuing without cache")
		redisClient = nil
	} else {
		log.Infof("Successfully connected to Redis with connection pool (size: %d, min_idle: %d)",
			cfg.Redis.PoolSize, cfg.Redis.MinIdleConns)
	}

	// Initialize advanced cache service with compression and dynamic TTL
	var cacheService *cache.Service
	var advancedCacheService *cache.AdvancedService
	if redisClient != nil {
		defaultTTL := time.Duration(cfg.Redis.DefaultTTLMinutes) * time.Minute
		cacheService = cache.NewService(redisClient, defaultTTL)
		advancedCacheService = cache.NewAdvancedService(redisClient, defaultTTL, cfg.Redis.CompressionThreshold)

		if cacheService.IsAvailable() {
			log.Infof("Cache service initialized: TTL=%dm, compression_threshold=%dB",
				cfg.Redis.DefaultTTLMinutes, cfg.Redis.CompressionThreshold)

			// Pre-warm cache with frequently accessed data
			go func() {
				time.Sleep(5 * time.Second) // Wait for app to fully start
				warmupCtx, warmupCancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer warmupCancel()

				warmupData := map[string]interface{}{
					"system:status": map[string]string{"status": "ready"},
				}

				if err := advancedCacheService.WarmCache(warmupCtx, warmupData); err != nil {
					log.WithError(err).Warn("Failed to warm cache")
				} else {
					log.Info("Cache warmed successfully")
				}
			}()
		}
	} else {
		log.Info("Cache service disabled (Redis not available)")
	}

	// Initialize repositories
	articleRepo := repository.NewArticleRepository(dbPool)
	jobRepo := repository.NewScrapingJobRepository(dbPool, log)

	// Initialize services
	scraperService := scraper.NewService(&cfg.Scraper, articleRepo, jobRepo, log)

	// Initialize scheduler if enabled (with database for analytics refresh)
	var scraperScheduler *scheduler.Scheduler
	if cfg.Scraper.ScheduleEnabled {
		interval := cfg.Scraper.GetScheduleInterval()
		scraperScheduler = scheduler.NewScheduler(scraperService, dbPool, interval, log)

		// Start scheduler in background
		go scraperScheduler.Start(context.Background())
		log.Infof("Scheduled scraping enabled with interval: %v (analytics refresh: every 15min)", interval)
	} else {
		log.Info("Scheduled scraping disabled")
	}

	// Initialize content processor for hybrid scraping
	var contentProcessor *scraper.ContentProcessor
	if cfg.Scraper.EnableFullContentExtraction && cfg.Scraper.ContentExtractionAsync {
		// CRITICAL FIX: Wait for browser pool to be ready before starting content processor
		if cfg.Scraper.EnableBrowserScraping && scraperService != nil {
			log.Info("Waiting for browser pool initialization (Chrome download may take 20-30s)...")
			time.Sleep(20 * time.Second)

			// Verify browser pool is available
			browserStats := scraperService.GetHealth(context.Background())
			if pool, ok := browserStats["browser_pool"].(map[string]interface{}); ok {
				if enabled, ok := pool["enabled"].(bool); ok && enabled {
					log.Info("✅ Browser pool ready for content extraction")
				}
			}
		}

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

	// Initialize stock service (if configured)
	var stockService *stock.Service
	var stockHandler *handlers.StockHandler

	if cfg.Stock.APIKey != "" {
		log.Info("Initializing stock service")

		stockConfig := &stock.Config{
			APIKey:          cfg.Stock.APIKey,
			APIProvider:     cfg.Stock.APIProvider,
			CacheTTL:        cfg.Stock.CacheTTL,
			RateLimitPerMin: cfg.Stock.RateLimitPerMin,
			Timeout:         cfg.Stock.Timeout,
			EnableCache:     cfg.Stock.EnableCache && redisClient != nil,
		}

		stockService = stock.NewService(stockConfig, redisClient, log)
		stockHandler = handlers.NewStockHandler(stockService, log)

		// Connect stock service to AI service for automatic enrichment via adapter
		if aiService != nil {
			stockAdapter := &StockServiceAdapter{service: stockService}
			aiService.SetStockService(stockAdapter)
			log.Info("Stock service connected to AI service for automatic enrichment")
		}

		log.Info("Stock service initialized successfully")
	} else {
		log.Info("Stock service disabled (no API key configured)")
	}

	// Initialize email service and processor (if configured)
	var emailProcessor *email.Processor

	if cfg.Email.Enabled {
		log.Info("Initializing email service")

		// Create email repository
		emailRepo := repository.NewEmailRepository(dbPool)

		// Configure email service
		emailServiceConfig := &email.Config{
			Host:            cfg.Email.Host,
			Port:            cfg.Email.Port,
			Username:        cfg.Email.Username,
			Password:        cfg.Email.Password,
			UseTLS:          cfg.Email.UseTLS,
			AllowedSenders:  cfg.Email.AllowedSenders,
			PollInterval:    cfg.Email.PollInterval,
			MaxRetries:      cfg.Email.MaxRetries,
			RetryDelay:      cfg.Email.RetryDelay,
			MarkAsRead:      cfg.Email.MarkAsRead,
			DeleteAfterRead: cfg.Email.DeleteAfterRead,
			FetchExisting:   cfg.Email.FetchExisting,
			MaxDaysBack:     cfg.Email.MaxDaysBack,
		}

		emailService := email.NewService(emailServiceConfig, log)

		// Test connection
		testCtx, testCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer testCancel()
		if err := emailService.TestConnection(testCtx); err != nil {
			log.WithError(err).Warn("Email connection test failed, continuing anyway...")
		} else {
			log.Info("Email connection test successful")
		}

		// Configure email processor
		processorConfig := &email.ProcessorConfig{
			PollInterval:    cfg.Email.PollInterval,
			MaxRetries:      cfg.Email.MaxRetries,
			ProcessArticles: true, // Automatically convert emails to articles
			UseAI:           cfg.AI.Enabled && aiService != nil,
		}

		emailProcessor = email.NewProcessor(
			emailService,
			emailRepo,
			articleRepo,
			aiService,
			processorConfig,
			log,
		)

		// Start email processor in background
		go emailProcessor.Start(context.Background())
		log.Infof("Email processor started with interval: %v", cfg.Email.PollInterval)
	} else {
		log.Info("Email integration disabled")
	}

	// Initialize handlers
	articleHandler := handlers.NewArticleHandler(articleRepo, cacheService, log)
	articleHandler.SetScraperService(scraperService) // Enable content extraction endpoint
	scraperHandler := handlers.NewScraperHandler(scraperService, articleHandler, log)

	// Initialize configuration handler for runtime settings management
	configHandler := handlers.NewConfigHandler(cfg, log)
	if scraperScheduler != nil {
		configHandler.SetScheduler(scraperScheduler)
	}
	log.Info("Configuration handler initialized with 4 profiles (fast, balanced, deep, conservative)")

	// Initialize cache handler
	var cacheHandler *handlers.CacheHandler
	if redisClient != nil {
		invalidationService := cache.NewInvalidationService(redisClient)
		cacheHandler = handlers.NewCacheHandler(cacheService, advancedCacheService, invalidationService, log)
		log.Info("Cache handler initialized with advanced features")
	}

	// Initialize email handler (if email processor exists)
	var emailHandler *handlers.EmailHandler
	if emailProcessor != nil {
		emailHandler = handlers.NewEmailHandler(emailProcessor, log)
		log.Info("Email handler initialized")
	}

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

	// Setup routes with comprehensive health monitoring and configuration API
	api.SetupRoutes(app, articleHandler, scraperHandler, aiHandler, stockHandler, emailHandler, cacheHandler, configHandler, rateLimiter, auth, log, dbPool, redisClient, cacheService, scraperService, aiProcessor)

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

	// Stop email processor if running
	if emailProcessor != nil && emailProcessor.IsRunning() {
		log.Info("Stopping email processor...")
		emailProcessor.Stop()
	}

	// Cleanup scraper service (closes browser pool if active)
	log.Info("Cleaning up scraper resources...")
	scraperService.Cleanup()

	// Cleanup stock service (closes rate limiter)
	if stockService != nil {
		log.Info("Cleaning up stock service...")
		stockService.Close()
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.WithError(err).Error("Server forced to shutdown")
	}

	log.Info("Server exited")
}

// StockServiceAdapter adapts stock.Service to ai.StockService interface
type StockServiceAdapter struct {
	service *stock.Service
}

func (a *StockServiceAdapter) GetMultipleQuotes(ctx context.Context, symbols []string) (map[string]*ai.StockQuote, error) {
	quotes, err := a.service.GetMultipleQuotes(ctx, symbols)
	if err != nil {
		return nil, err
	}

	// Convert stock.StockQuote to ai.StockQuote
	result := make(map[string]*ai.StockQuote)
	for symbol, quote := range quotes {
		result[symbol] = &ai.StockQuote{
			Symbol:        quote.Symbol,
			Name:          quote.Name,
			Price:         quote.Price,
			Change:        quote.Change,
			ChangePercent: quote.ChangePercent,
			Volume:        quote.Volume,
			MarketCap:     quote.MarketCap,
			Exchange:      quote.Exchange,
		}
	}

	return result, nil
}
