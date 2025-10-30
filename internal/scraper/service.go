package scraper

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

	"github.com/jeffrey/intellinieuws/internal/models"
	"github.com/jeffrey/intellinieuws/internal/repository"
	"github.com/jeffrey/intellinieuws/internal/scraper/browser"
	"github.com/jeffrey/intellinieuws/internal/scraper/html"
	"github.com/jeffrey/intellinieuws/internal/scraper/rss"
	"github.com/jeffrey/intellinieuws/pkg/config"
	"github.com/jeffrey/intellinieuws/pkg/logger"
	"github.com/jeffrey/intellinieuws/pkg/utils"
)

// Service manages all scraping operations
type Service struct {
	rssScrap         *rss.Scraper
	contentExtractor *html.ContentExtractor
	browserPool      *browser.BrowserPool
	browserExtractor *browser.Extractor
	articleRepo      *repository.ArticleRepository
	jobRepo          *repository.ScrapingJobRepository
	rateLimiter      *utils.ScraperRateLimiter
	robotsChecker    *utils.RobotsChecker
	logger           *logger.Logger
	config           *config.ScraperConfig
	circuitBreaker   *utils.CircuitBreakerManager // PHASE 4: Resilience
}

// NewService creates a new scraper service
func NewService(
	cfg *config.ScraperConfig,
	articleRepo *repository.ArticleRepository,
	jobRepo *repository.ScrapingJobRepository,
	log *logger.Logger,
) *Service {
	// Initialize browser scraping if enabled
	var browserPool *browser.BrowserPool
	var browserExtractor *browser.Extractor

	if cfg.EnableBrowserScraping {
		log.Info("Initializing headless browser pool...")
		pool, err := browser.NewBrowserPool(cfg.BrowserPoolSize, log)
		if err != nil {
			log.WithError(err).Warn("Failed to initialize browser pool, browser scraping disabled")
		} else {
			browserPool = pool
			browserExtractor = browser.NewExtractor(pool, browser.ExtractorConfig{
				Timeout:       cfg.BrowserTimeout,
				WaitAfterLoad: cfg.BrowserWaitAfterLoad,
				MaxConcurrent: cfg.BrowserMaxConcurrent,
			}, log)
			log.Infof("Browser pool initialized: %d instances, fallback_only=%v",
				cfg.BrowserPoolSize, cfg.BrowserFallbackOnly)
		}
	}

	// Initialize content extractor
	contentExtractor := html.NewContentExtractor(cfg.UserAgent, log)

	// Enable browser fallback if configured
	if cfg.EnableBrowserScraping && browserExtractor != nil && cfg.BrowserFallbackOnly {
		contentExtractor.SetBrowserExtractor(browserExtractor, true)
	}

	return &Service{
		rssScrap:         rss.NewScraper(cfg.UserAgent, log),
		contentExtractor: contentExtractor,
		browserPool:      browserPool,
		browserExtractor: browserExtractor,
		articleRepo:      articleRepo,
		jobRepo:          jobRepo,
		rateLimiter:      utils.NewScraperRateLimiter(cfg.RateLimitSeconds),
		robotsChecker:    utils.NewRobotsChecker(cfg.UserAgent),
		logger:           log.WithComponent("scraper-service"),
		config:           cfg,
		circuitBreaker:   utils.NewCircuitBreakerManager(),
	}
}

// ScrapeSources defined RSS feeds
var ScrapeSources = map[string]string{
	"nu.nl":  "https://www.nu.nl/rss",
	"ad.nl":  "https://www.ad.nl/rss.xml",
	"nos.nl": "https://feeds.nos.nl/nosnieuwsalgemeen",
}

// ScrapeSource scrapes a single news source with comprehensive error handling
func (s *Service) ScrapeSource(ctx context.Context, source string, feedURL string) (*ScrapingResult, error) {
	s.logger.Infof("Starting scrape for source: %s", source)
	startTime := time.Now()

	result := &ScrapingResult{
		Source:    source,
		StartTime: startTime,
		Status:    models.JobStatusRunning,
	}

	// Create job record
	jobID, err := s.jobRepo.CreateJob(ctx, source)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to create job record, continuing anyway")
		jobID = 0 // Continue without job tracking
	}

	// Mark job as started
	if jobID > 0 {
		if err := s.jobRepo.StartJob(ctx, jobID); err != nil {
			s.logger.WithError(err).Warn("Failed to start job record")
		}
	}

	// Defer panic recovery and job completion
	defer func() {
		if r := recover(); r != nil {
			s.logger.Errorf("Panic recovered in scrape for %s: %v", source, r)
			result.Error = fmt.Sprintf("panic: %v", r)
			result.Status = models.JobStatusFailed
			result.EndTime = time.Now()

			// Mark job as failed
			if jobID > 0 {
				if err := s.jobRepo.FailJob(ctx, jobID, result.Error); err != nil {
					s.logger.WithError(err).Warn("Failed to mark job as failed")
				}
			}
		}
	}()

	// Check context cancellation
	if ctx.Err() != nil {
		result.Error = "context cancelled"
		result.Status = models.JobStatusFailed
		result.EndTime = time.Now()

		// Mark job as failed
		if jobID > 0 {
			if err := s.jobRepo.FailJob(ctx, jobID, result.Error); err != nil {
				s.logger.WithError(err).Warn("Failed to mark job as failed")
			}
		}
		return result, ctx.Err()
	}

	// Check robots.txt if enabled
	if s.config.EnableRobotsTxtCheck {
		allowed, err := s.robotsChecker.IsAllowed(feedURL)
		if err != nil {
			s.logger.WithError(err).Warnf("Error checking robots.txt for %s, continuing anyway", source)
			// Continue scraping even if robots.txt check fails
		} else if !allowed {
			result.Error = "robots.txt disallows scraping"
			result.Status = models.JobStatusFailed
			result.EndTime = time.Now()
			s.logger.Warnf("Robots.txt disallows scraping of %s", source)

			// Mark job as failed
			if jobID > 0 {
				if err := s.jobRepo.FailJob(ctx, jobID, result.Error); err != nil {
					s.logger.WithError(err).Warn("Failed to mark job as failed")
				}
			}
			return result, fmt.Errorf("robots.txt disallows scraping of %s", feedURL)
		}
	}

	// Apply rate limiting with timeout
	domain, err := utils.GetDomain(feedURL)
	if err != nil {
		result.Error = fmt.Sprintf("invalid URL: %v", err)
		result.Status = models.JobStatusFailed
		result.EndTime = time.Now()
		return result, fmt.Errorf("invalid URL for %s: %w", source, err)
	}

	rateLimitCtx, rateLimitCancel := context.WithTimeout(ctx, 30*time.Second)
	defer rateLimitCancel()

	if err := s.rateLimiter.Wait(rateLimitCtx, domain); err != nil {
		result.Error = fmt.Sprintf("rate limit error: %v", err)
		result.Status = models.JobStatusFailed
		result.EndTime = time.Now()

		// Mark job as failed
		if jobID > 0 {
			if err := s.jobRepo.FailJob(ctx, jobID, result.Error); err != nil {
				s.logger.WithError(err).Warn("Failed to mark job as failed")
			}
		}
		return result, fmt.Errorf("rate limit error for %s: %w", source, err)
	}

	// Scrape RSS feed with timeout and circuit breaker (PHASE 4: Resilience)
	scrapeCtx, scrapeCancel := context.WithTimeout(ctx, s.config.GetTimeout())
	defer scrapeCancel()

	// Use circuit breaker to prevent cascading failures
	cb := s.circuitBreaker.GetOrCreate(source, 5, 5*time.Minute)

	var articles []*models.ArticleCreate
	err = cb.Call(func() error {
		var scrapeErr error
		articles, scrapeErr = s.rssScrap.ScrapeFeed(scrapeCtx, feedURL, source)
		return scrapeErr
	})

	if err != nil {
		if cb.IsOpen() {
			result.Error = fmt.Sprintf("circuit breaker open (too many failures)")
			s.logger.Warnf("Circuit breaker OPEN for %s - blocking requests", source)
		} else {
			result.Error = fmt.Sprintf("scraping failed: %v", err)
		}
		result.Status = models.JobStatusFailed
		result.EndTime = time.Now()

		// Mark job as failed
		if jobID > 0 {
			if err := s.jobRepo.FailJob(ctx, jobID, result.Error); err != nil {
				s.logger.WithError(err).Warn("Failed to mark job as failed")
			}
		}
		return result, fmt.Errorf("scraping failed for %s: %w", source, err)
	}

	s.logger.Infof("Found %d articles from %s", len(articles), source)

	if len(articles) == 0 {
		s.logger.Warnf("No articles found for %s", source)
		result.Status = models.JobStatusCompleted
		result.EndTime = time.Now()
		result.Duration = time.Since(startTime)

		// Mark job as completed with 0 articles
		if jobID > 0 {
			if err := s.jobRepo.CompleteJob(ctx, jobID, 0); err != nil {
				s.logger.WithError(err).Warn("Failed to complete job record")
			}
		}
		return result, nil
	}

	// Filter and validate articles before batch insert
	validArticles := make([]*models.ArticleCreate, 0, len(articles))
	skipped := 0

	// Batch duplicate check if enabled (OPTIMIZED: 50 queries → 1 query)
	var existsMap map[string]bool
	if s.config.EnableDuplicateDetection {
		// Collect all URLs for batch checking
		urls := make([]string, 0, len(articles))
		for _, article := range articles {
			if article.URL != "" {
				urls = append(urls, article.URL)
			}
		}

		// Single batch query to check all URLs at once
		var err error
		existsMap, err = s.articleRepo.ExistsByURLBatch(ctx, urls)
		if err != nil {
			s.logger.WithError(err).Warn("Batch duplicate check failed, continuing with all articles...")
			// Create empty map so we don't skip any articles on error
			existsMap = make(map[string]bool)
		}
		s.logger.Debugf("Batch duplicate check completed for %d URLs", len(urls))
	}

	// Filter articles based on batch duplicate check results
	for _, article := range articles {
		// Check context
		if ctx.Err() != nil {
			s.logger.Warn("Context cancelled during article filtering")
			break
		}

		// Validate article data
		if article.URL == "" {
			s.logger.Warnf("Skipping article with empty URL: %s", article.Title)
			skipped++
			continue
		}

		// Check if URL exists using batch result (O(1) lookup)
		if s.config.EnableDuplicateDetection && existsMap[article.URL] {
			skipped++
			continue
		}

		validArticles = append(validArticles, article)
	}

	// Batch insert all valid articles
	stored := 0
	var storageErrors []string

	if len(validArticles) > 0 {
		storeCtx, storeCancel := context.WithTimeout(ctx, 30*time.Second)
		defer storeCancel()

		inserted, err := s.articleRepo.CreateBatch(storeCtx, validArticles)
		stored = inserted

		if err != nil {
			errMsg := fmt.Sprintf("Batch insert error: %v", err)
			s.logger.Error(errMsg)
			storageErrors = append(storageErrors, errMsg)
		}

		// Update skipped count for duplicates caught by database
		if inserted < len(validArticles) {
			skipped += (len(validArticles) - inserted)
		}
	}

	result.ArticlesFound = len(articles)
	result.ArticlesStored = stored
	result.ArticlesSkipped = skipped
	result.Status = models.JobStatusCompleted
	result.EndTime = time.Now()
	result.Duration = time.Since(startTime)

	if len(storageErrors) > 0 {
		result.Error = fmt.Sprintf("%d storage errors occurred", len(storageErrors))
		result.Status = StatusPartialSuccess
	}

	// Mark job as completed
	if jobID > 0 {
		if result.Status == models.JobStatusCompleted {
			if err := s.jobRepo.CompleteJob(ctx, jobID, stored); err != nil {
				s.logger.WithError(err).Warn("Failed to complete job record")
			}
		} else {
			if err := s.jobRepo.FailJob(ctx, jobID, result.Error); err != nil {
				s.logger.WithError(err).Warn("Failed to mark job as failed")
			}
		}
	}

	s.logger.Infof("Completed scrape for %s: stored=%d, skipped=%d, errors=%d, duration=%v",
		source, stored, skipped, len(storageErrors), result.Duration)

	return result, nil
}

// ScrapeAllSources scrapes all configured sources in parallel with controlled concurrency
func (s *Service) ScrapeAllSources(ctx context.Context) (map[string]*ScrapingResult, error) {
	s.logger.Info("Starting parallel scrape for all sources")
	startTime := time.Now()

	// Filter sources based on config
	sourcesToScrape := make(map[string]string)
	for _, targetSite := range s.config.TargetSites {
		if feedURL, exists := ScrapeSources[targetSite]; exists {
			sourcesToScrape[targetSite] = feedURL
		} else {
			s.logger.Warnf("Source %s not found in ScrapeSources map", targetSite)
		}
	}

	// Use channels to collect results
	type scrapeJob struct {
		source string
		result *ScrapingResult
		err    error
	}

	// Limit concurrent scraping with semaphore (OPTIMIZED: prevents overwhelming)
	maxConcurrent := 3 // Reasonable limit for stability
	if len(sourcesToScrape) < maxConcurrent {
		maxConcurrent = len(sourcesToScrape)
	}
	semaphore := make(chan struct{}, maxConcurrent)

	resultChan := make(chan scrapeJob, len(sourcesToScrape))
	var wg sync.WaitGroup

	s.logger.Infof("Scraping %d sources with max concurrency: %d", len(sourcesToScrape), maxConcurrent)

	// Launch goroutines for parallel scraping with semaphore control
	for source, feedURL := range sourcesToScrape {
		wg.Add(1)
		go func(src, url string) {
			defer wg.Done()

			// Acquire semaphore (blocks if limit reached)
			semaphore <- struct{}{}
			defer func() { <-semaphore }() // Release semaphore when done

			result, err := s.ScrapeSource(ctx, src, url)
			resultChan <- scrapeJob{
				source: src,
				result: result,
				err:    err,
			}

			if err != nil {
				s.logger.WithError(err).Errorf("Failed to scrape source: %s", src)
			}
		}(source, feedURL)
	}

	// Close channel when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	results := make(map[string]*ScrapingResult)
	for job := range resultChan {
		results[job.source] = job.result
	}

	totalDuration := time.Since(startTime)
	s.logger.Infof("Completed parallel scrape for all sources in %v", totalDuration)

	return results, nil
}

// ScrapeWithRetry scrapes with enhanced retry logic and exponential backoff
func (s *Service) ScrapeWithRetry(ctx context.Context, source string, feedURL string) (*ScrapingResult, error) {
	var lastErr error
	var result *ScrapingResult

	for attempt := 1; attempt <= s.config.RetryAttempts; attempt++ {
		result, lastErr = s.ScrapeSource(ctx, source, feedURL)

		if lastErr == nil {
			return result, nil
		}

		// Check if error is rate limit (429) or timeout
		isRateLimit := isRateLimitError(lastErr)
		isTimeout := isTimeoutError(lastErr)

		if attempt < s.config.RetryAttempts {
			// Calculate exponential backoff with jitter (5s, 10s, 20s)
			baseDelay := time.Duration(1<<uint(attempt-1)) * 5 * time.Second

			// Add jitter (±20%) to prevent thundering herd
			jitter := time.Duration(float64(baseDelay) * 0.2 * (2.0*rand.Float64() - 1.0))
			backoff := baseDelay + jitter

			// Special handling for rate limits (longer backoff)
			if isRateLimit {
				backoff = backoff * 3 // 15s, 30s, 60s for 429 errors
				s.logger.Warnf("Rate limit detected for %s (attempt %d/%d), extended backoff: %v",
					source, attempt, s.config.RetryAttempts, backoff)
			} else if isTimeout {
				s.logger.Warnf("Timeout for %s (attempt %d/%d), retrying in %v",
					source, attempt, s.config.RetryAttempts, backoff)
			} else {
				s.logger.Warnf("Scrape attempt %d/%d failed for %s: %v. Retrying in %v...",
					attempt, s.config.RetryAttempts, source, lastErr, backoff)
			}

			// Wait before retry with context cancellation support
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return result, ctx.Err()
			}
		}
	}

	return result, fmt.Errorf("all %d attempts failed: %w", s.config.RetryAttempts, lastErr)
}

// ScrapingResult contains the result of a scraping operation
type ScrapingResult struct {
	Source          string
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
	Status          string
	ArticlesFound   int
	ArticlesStored  int
	ArticlesSkipped int
	Error           string
}

// Success statuses
const (
	StatusSuccess        = "success"
	StatusPartialSuccess = "partial_success"
	StatusFailed         = "failed"
)

// GetStats returns scraping statistics including circuit breaker status
func (s *Service) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats, err := s.articleRepo.GetStatsBySource(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return map[string]interface{}{
		"articles_by_source": stats,
		"rate_limit_delay":   s.rateLimiter.GetDelay().Seconds(),
		"sources_configured": s.config.TargetSites,
		"circuit_breakers":   s.circuitBreaker.GetAllStats(), // PHASE 4: Circuit breaker stats
	}, nil
}

// GetHealth returns health status of the scraper service (PHASE 4: Health monitoring)
func (s *Service) GetHealth(ctx context.Context) map[string]interface{} {
	health := map[string]interface{}{
		"status":            "healthy",
		"circuit_breakers":  s.circuitBreaker.GetAllStats(),
		"rate_limiter":      s.rateLimiter.GetDelay().Seconds(),
		"sources_available": len(s.config.TargetSites),
	}

	// Check if any circuit breakers are open
	for _, cb := range s.circuitBreaker.GetAllStats() {
		if state, ok := cb["state"].(string); ok && state == "open" {
			health["status"] = "degraded"
			health["warning"] = fmt.Sprintf("Circuit breaker '%s' is open", cb["name"])
			break
		}
	}

	return health
}

// EnrichArticleContent downloads full text for an article (Hybrid approach)
func (s *Service) EnrichArticleContent(ctx context.Context, articleID int64) error {
	// Get article from database
	article, err := s.articleRepo.GetByID(ctx, articleID)
	if err != nil {
		return fmt.Errorf("failed to get article: %w", err)
	}

	// Skip if already has content
	article, err = s.articleRepo.GetArticleWithContent(ctx, articleID)
	if err != nil {
		return fmt.Errorf("failed to get article with content: %w", err)
	}

	if article.ContentExtracted {
		s.logger.Debugf("Article %d already has content, skipping", articleID)
		return nil
	}

	s.logger.Infof("Enriching article %d: %s", articleID, article.Title)

	// Extract full content with rate limiting
	domain, err := utils.GetDomain(article.URL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Apply rate limiting
	if err := s.rateLimiter.Wait(ctx, domain); err != nil {
		return fmt.Errorf("rate limit error: %w", err)
	}

	// Extract content
	content, err := s.contentExtractor.ExtractContent(ctx, article.URL, article.Source)
	if err != nil {
		s.logger.WithError(err).Warnf("Failed to extract content for article %d", articleID)
		return err
	}

	// Update article with full content
	if err := s.articleRepo.UpdateContent(ctx, articleID, content); err != nil {
		return fmt.Errorf("failed to update content: %w", err)
	}

	s.logger.Infof("Successfully enriched article %d with %d characters", articleID, len(content))
	return nil
}

// EnrichArticlesBatch enriches multiple articles with content extraction
func (s *Service) EnrichArticlesBatch(ctx context.Context, articleIDs []int64) (int, error) {
	var wg sync.WaitGroup
	successChan := make(chan bool, len(articleIDs))

	// Limit concurrency to avoid overwhelming servers
	semaphore := make(chan struct{}, 3)

	for _, id := range articleIDs {
		wg.Add(1)
		go func(articleID int64) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			err := s.EnrichArticleContent(ctx, articleID)
			successChan <- (err == nil)
		}(id)
	}

	wg.Wait()
	close(successChan)

	// Count successes
	successCount := 0
	for success := range successChan {
		if success {
			successCount++
		}
	}

	s.logger.Infof("Enriched %d/%d articles with content", successCount, len(articleIDs))
	return successCount, nil
}

// GetContentExtractionStats returns statistics about content extraction
func (s *Service) GetContentExtractionStats(ctx context.Context) (map[string]interface{}, error) {
	stats, err := s.articleRepo.GetContentExtractionStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get content extraction stats: %w", err)
	}

	return map[string]interface{}{
		"content_extraction": stats,
		"browser_pool":       s.getBrowserPoolStats(),
	}, nil
}

// isRateLimitError checks if error is a rate limit (429) error
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "too many requests")
}

// isTimeoutError checks if error is a timeout error
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline exceeded") ||
		strings.Contains(errStr, "context canceled")
}

// getBrowserPoolStats returns browser pool statistics
func (s *Service) getBrowserPoolStats() map[string]interface{} {
	if s.browserPool == nil {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	stats := s.browserPool.GetStats()
	stats["enabled"] = true
	return stats
}

// Cleanup closes browser pool and other resources
func (s *Service) Cleanup() {
	if s.browserPool != nil {
		s.logger.Info("Closing browser pool...")
		s.browserPool.Close()
	}
}
