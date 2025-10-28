package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/jeffrey/intellinieuws/internal/scraper"
	"github.com/jeffrey/intellinieuws/pkg/logger"
)

// Scheduler manages periodic scraping tasks
type Scheduler struct {
	scraperService *scraper.Service
	logger         *logger.Logger
	interval       time.Duration
	ticker         *time.Ticker
	stopChan       chan struct{}
	wg             sync.WaitGroup
	running        bool
	mu             sync.Mutex
}

// NewScheduler creates a new scheduler
func NewScheduler(
	scraperService *scraper.Service,
	interval time.Duration,
	log *logger.Logger,
) *Scheduler {
	return &Scheduler{
		scraperService: scraperService,
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
