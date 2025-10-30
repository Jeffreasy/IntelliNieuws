package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jeffrey/intellinieuws/internal/scraper"
	"github.com/jeffrey/intellinieuws/pkg/logger"
)

// Scheduler manages periodic scraping tasks and analytics refresh
type Scheduler struct {
	scraperService         *scraper.Service
	db                     *pgxpool.Pool
	logger                 *logger.Logger
	interval               time.Duration
	analyticsRefreshTicker *time.Ticker
	ticker                 *time.Ticker
	stopChan               chan struct{}
	wg                     sync.WaitGroup
	running                bool
	mu                     sync.Mutex
}

// NewScheduler creates a new scheduler
func NewScheduler(
	scraperService *scraper.Service,
	db *pgxpool.Pool,
	interval time.Duration,
	log *logger.Logger,
) *Scheduler {
	return &Scheduler{
		scraperService: scraperService,
		db:             db,
		logger:         log.WithComponent("scheduler"),
		interval:       interval,
		stopChan:       make(chan struct{}),
	}
}

// Start begins the scheduled scraping
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		s.logger.Warn("Scheduler already running")
		return
	}
	s.running = true
	s.ticker = time.NewTicker(s.interval)
	s.mu.Unlock()

	s.logger.Infof("Starting scheduler with interval: %v", s.interval)

	// Start analytics refresh ticker (every 15 minutes)
	if s.db != nil {
		s.analyticsRefreshTicker = time.NewTicker(15 * time.Minute)
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			// Run initial refresh
			s.refreshAnalytics(ctx)

			for {
				select {
				case <-s.analyticsRefreshTicker.C:
					s.refreshAnalytics(ctx)
				case <-s.stopChan:
					return
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Run initial scrape
	go s.runScrape(ctx)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.ticker.C:
				s.runScrape(ctx)
			case <-s.stopChan:
				s.logger.Info("Scheduler stopped")
				return
			case <-ctx.Done():
				s.logger.Info("Scheduler context cancelled")
				return
			}
		}
	}()
}

// runScrape executes a scraping operation
func (s *Scheduler) runScrape(ctx context.Context) {
	s.logger.Info("Running scheduled scrape")
	startTime := time.Now()

	results, err := s.scraperService.ScrapeAllSources(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Scheduled scrape failed")
		return
	}

	// Log results
	totalStored := 0
	totalSkipped := 0
	for source, result := range results {
		totalStored += result.ArticlesStored
		totalSkipped += result.ArticlesSkipped

		if result.Error != "" {
			s.logger.Warnf("Source %s had errors: %s", source, result.Error)
		}
	}

	duration := time.Since(startTime)
	s.logger.Infof("Scheduled scrape completed: stored=%d, skipped=%d, duration=%v",
		totalStored, totalSkipped, duration)
}

// refreshAnalytics refreshes materialized views
func (s *Scheduler) refreshAnalytics(ctx context.Context) {
	if s.db == nil {
		return
	}

	s.logger.Info("Refreshing analytics materialized views...")
	startTime := time.Now()

	query := `SELECT * FROM refresh_analytics_views(TRUE)`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		s.logger.WithError(err).Error("Failed to refresh analytics views")
		return
	}
	defer rows.Close()

	totalRows := int64(0)
	viewCount := 0
	for rows.Next() {
		var viewName string
		var refreshTimeMs int
		var rowsAffected int64

		if err := rows.Scan(&viewName, &refreshTimeMs, &rowsAffected); err != nil {
			s.logger.WithError(err).Warn("Failed to scan refresh result")
			continue
		}

		totalRows += rowsAffected
		viewCount++
		s.logger.Debugf("Refreshed %s: %d rows in %dms", viewName, rowsAffected, refreshTimeMs)
	}

	duration := time.Since(startTime)
	s.logger.Infof("Analytics refresh completed: %d views, %d rows, duration=%v",
		viewCount, totalRows, duration)
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.logger.Info("Stopping scheduler...")
	close(s.stopChan)

	if s.ticker != nil {
		s.ticker.Stop()
	}

	if s.analyticsRefreshTicker != nil {
		s.analyticsRefreshTicker.Stop()
	}

	s.wg.Wait()
	s.running = false
	s.logger.Info("Scheduler stopped successfully")
}

// IsRunning returns whether the scheduler is currently running
func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// UpdateInterval updates the scraping interval (requires restart)
func (s *Scheduler) UpdateInterval(interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.interval = interval
	s.logger.Infof("Scheduler interval updated to: %v", interval)
}
