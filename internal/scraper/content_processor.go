package scraper

import (
	"context"
	"sync"
	"time"

	"github.com/jeffrey/intellinieuws/pkg/logger"
)

// ContentProcessor handles background content extraction
type ContentProcessor struct {
	service   *Service
	logger    *logger.Logger
	interval  time.Duration
	enabled   bool
	stopChan  chan struct{}
	wg        sync.WaitGroup
	isRunning bool
	mu        sync.Mutex
}

// NewContentProcessor creates a new content processor
func NewContentProcessor(service *Service, interval time.Duration, enabled bool, log *logger.Logger) *ContentProcessor {
	return &ContentProcessor{
		service:  service,
		logger:   log.WithComponent("content-processor"),
		interval: interval,
		enabled:  enabled,
		stopChan: make(chan struct{}),
	}
}

// Start begins background processing
func (p *ContentProcessor) Start(ctx context.Context) error {
	p.mu.Lock()
	if p.isRunning {
		p.mu.Unlock()
		return nil
	}
	p.isRunning = true
	p.mu.Unlock()

	if !p.enabled {
		p.logger.Info("Content extraction is disabled, processor not started")
		return nil
	}

	p.logger.Infof("Starting content processor (interval: %v)", p.interval)

	p.wg.Add(1)
	go p.run(ctx)

	return nil
}

// Stop stops the background processor
func (p *ContentProcessor) Stop() {
	p.mu.Lock()
	if !p.isRunning {
		p.mu.Unlock()
		return
	}
	p.mu.Unlock()

	p.logger.Info("Stopping content processor...")
	close(p.stopChan)
	p.wg.Wait()

	p.mu.Lock()
	p.isRunning = false
	p.mu.Unlock()

	p.logger.Info("Content processor stopped")
}

// IsRunning returns whether the processor is running
func (p *ContentProcessor) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.isRunning
}

// run is the main processing loop
func (p *ContentProcessor) run(ctx context.Context) {
	defer p.wg.Done()

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	// Process immediately on start
	p.processArticles(ctx)

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Context cancelled, stopping content processor")
			return
		case <-p.stopChan:
			p.logger.Info("Stop signal received")
			return
		case <-ticker.C:
			p.processArticles(ctx)
		}
	}
}

// processArticles processes articles needing content extraction
func (p *ContentProcessor) processArticles(ctx context.Context) {
	p.logger.Debug("Processing articles needing content extraction...")
	startTime := time.Now()

	// Create a timeout context for this batch
	batchCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Get articles needing content
	articleIDs, err := p.service.articleRepo.GetArticlesNeedingContent(batchCtx, 10)
	if err != nil {
		p.logger.WithError(err).Error("Failed to get articles needing content")
		return
	}

	if len(articleIDs) == 0 {
		p.logger.Debug("No articles need content extraction")
		return
	}

	p.logger.Infof("Found %d articles needing content extraction", len(articleIDs))

	// Enrich articles
	successCount, err := p.service.EnrichArticlesBatch(batchCtx, articleIDs)
	if err != nil {
		p.logger.WithError(err).Warn("Batch enrichment completed with errors")
	}

	duration := time.Since(startTime)
	p.logger.Infof("Content extraction batch completed: %d/%d successful, duration: %v",
		successCount, len(articleIDs), duration)
}

// GetStats returns processor statistics
func (p *ContentProcessor) GetStats() map[string]interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	return map[string]interface{}{
		"is_running": p.isRunning,
		"enabled":    p.enabled,
		"interval":   p.interval.String(),
	}
}
