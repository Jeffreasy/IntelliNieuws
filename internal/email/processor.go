package email

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jeffrey/intellinieuws/internal/ai"
	"github.com/jeffrey/intellinieuws/internal/models"
	"github.com/jeffrey/intellinieuws/internal/repository"
	"github.com/jeffrey/intellinieuws/pkg/logger"
)

// Processor handles email fetching and processing
type Processor struct {
	emailService *Service
	emailRepo    *repository.EmailRepository
	articleRepo  *repository.ArticleRepository
	aiService    *ai.Service
	config       *ProcessorConfig
	logger       *logger.Logger
	ticker       *time.Ticker
	stopChan     chan struct{}
	wg           sync.WaitGroup
	running      bool
	mu           sync.Mutex
}

// ProcessorConfig holds processor configuration
type ProcessorConfig struct {
	PollInterval    time.Duration
	MaxRetries      int
	ProcessArticles bool // Whether to automatically process emails into articles
	UseAI           bool // Whether to use AI for processing
}

// NewProcessor creates a new email processor
func NewProcessor(
	emailService *Service,
	emailRepo *repository.EmailRepository,
	articleRepo *repository.ArticleRepository,
	aiService *ai.Service,
	config *ProcessorConfig,
	log *logger.Logger,
) *Processor {
	return &Processor{
		emailService: emailService,
		emailRepo:    emailRepo,
		articleRepo:  articleRepo,
		aiService:    aiService,
		config:       config,
		logger:       log.WithComponent("email-processor"),
		stopChan:     make(chan struct{}),
	}
}

// Start begins email processing
func (p *Processor) Start(ctx context.Context) {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		p.logger.Warn("Email processor already running")
		return
	}
	p.running = true
	p.ticker = time.NewTicker(p.config.PollInterval)
	p.mu.Unlock()

	p.logger.Infof("Starting email processor with interval: %v", p.config.PollInterval)

	// Fetch existing emails on startup if configured
	if p.emailService.config.FetchExisting {
		p.logger.Info("Fetching existing emails on startup...")
		go func() {
			fetchCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
			defer cancel()

			if count, err := p.FetchExistingEmails(fetchCtx); err != nil {
				p.logger.WithError(err).Error("Failed to fetch existing emails on startup")
			} else {
				p.logger.Infof("Successfully fetched %d existing emails on startup", count)
			}
		}()
	}

	// Run initial fetch of new emails
	go p.processCycle(ctx)

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for {
			select {
			case <-p.ticker.C:
				p.processCycle(ctx)
			case <-p.stopChan:
				p.logger.Info("Email processor stopped")
				return
			case <-ctx.Done():
				p.logger.Info("Email processor context cancelled")
				return
			}
		}
	}()
}

// processCycle executes one cycle of email fetching and processing
func (p *Processor) processCycle(ctx context.Context) {
	p.logger.Info("Starting email processing cycle")
	startTime := time.Now()

	// Fetch new emails
	emails, err := p.emailService.FetchNewEmails(ctx)
	if err != nil {
		p.logger.WithError(err).Error("Failed to fetch emails")
		return
	}

	if len(emails) == 0 {
		p.logger.Info("No new emails to process")
		return
	}

	p.logger.Infof("Fetched %d new emails", len(emails))

	// Store emails in database
	stored := 0
	skipped := 0
	for _, email := range emails {
		// Check if email already exists
		exists, err := p.emailRepo.Exists(ctx, email.MessageID)
		if err != nil {
			p.logger.WithError(err).Warn("Failed to check email existence")
			continue
		}

		if exists {
			p.logger.Debugf("Skipping duplicate email: %s", email.MessageID)
			skipped++
			continue
		}

		// Store email
		storedEmail, err := p.emailRepo.Create(ctx, email)
		if err != nil {
			p.logger.WithError(err).Error("Failed to store email")
			continue
		}

		stored++
		p.logger.Infof("Stored email: %s from %s", storedEmail.Subject, storedEmail.Sender)

		// Process into article if configured
		if p.config.ProcessArticles {
			if err := p.processEmailToArticle(ctx, storedEmail); err != nil {
				p.logger.WithError(err).Errorf("Failed to process email %d into article", storedEmail.ID)
				// Mark as failed
				p.emailRepo.MarkAsFailed(ctx, storedEmail.ID, err.Error())
			}
		}
	}

	// Retry failed emails
	if p.config.ProcessArticles {
		p.retryFailedEmails(ctx)
	}

	duration := time.Since(startTime)
	p.logger.Infof("Email processing cycle completed: stored=%d, skipped=%d, duration=%v",
		stored, skipped, duration)
}

// processEmailToArticle converts an email into an article
func (p *Processor) processEmailToArticle(ctx context.Context, email *models.Email) error {
	p.logger.Infof("Processing email %d into article: %s", email.ID, email.Subject)

	// Mark email as processing
	if err := p.updateEmailStatus(ctx, email.ID, models.EmailStatusProcessing); err != nil {
		p.logger.WithError(err).Warn("Failed to mark email as processing")
	}

	// Extract content (prefer text over HTML for now)
	content := email.BodyText
	if content == "" {
		content = email.BodyHTML
	}
	if content == "" {
		content = email.Subject
	}

	// Clean and prepare content
	content = strings.TrimSpace(content)
	if len(content) > 5000 {
		content = content[:5000] // Limit content length
	}

	// Create article
	article := &models.ArticleCreate{
		Title:       email.Subject,
		Summary:     p.generateSummary(content, 200),
		URL:         fmt.Sprintf("email://%s", email.MessageID),
		Published:   email.ReceivedDate,
		Source:      fmt.Sprintf("Email from %s", email.Sender),
		Author:      email.Sender,
		Category:    "Email",
		ContentHash: p.generateHash(email.MessageID),
	}

	// Store article
	storedArticle, err := p.articleRepo.Create(ctx, article)
	if err != nil {
		// Mark as failed with error code
		p.markEmailFailedWithCode(ctx, email.ID, err.Error(), "ARTICLE_CREATE_FAILED")
		return fmt.Errorf("failed to create article: %w", err)
	}

	p.logger.Infof("Created article %d from email %d", storedArticle.ID, email.ID)

	// Mark email as processed
	articleID := storedArticle.ID
	if err := p.emailRepo.MarkAsProcessed(ctx, email.ID, &articleID); err != nil {
		p.logger.WithError(err).Warn("Failed to mark email as processed")
	}

	// Process with AI if enabled
	if p.config.UseAI && p.aiService != nil {
		go func() {
			aiCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			_, err := p.aiService.ProcessArticle(aiCtx, storedArticle.ID)
			if err != nil {
				p.logger.WithError(err).Warnf("Failed to AI process article %d", storedArticle.ID)
			} else {
				p.logger.Infof("Successfully AI processed article %d", storedArticle.ID)
			}
		}()
	}

	return nil
}

// retryFailedEmails retries processing of failed emails
func (p *Processor) retryFailedEmails(ctx context.Context) {
	failedEmails, err := p.emailRepo.GetUnprocessed(ctx, p.config.MaxRetries, 10)
	if err != nil {
		p.logger.WithError(err).Error("Failed to get unprocessed emails")
		return
	}

	if len(failedEmails) == 0 {
		return
	}

	p.logger.Infof("Retrying %d failed emails", len(failedEmails))

	for _, email := range failedEmails {
		// Update last_retry_at before processing
		if err := p.updateLastRetryAt(ctx, email.ID); err != nil {
			p.logger.WithError(err).Warn("Failed to update last_retry_at")
		}

		if err := p.processEmailToArticle(ctx, &email); err != nil {
			p.logger.WithError(err).Errorf("Retry failed for email %d", email.ID)
			p.markEmailFailedWithCode(ctx, email.ID, err.Error(), "RETRY_FAILED")
		}
	}
}

// generateSummary generates a summary from content
func (p *Processor) generateSummary(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}

	// Find last sentence boundary
	summary := content[:maxLength]
	lastPeriod := strings.LastIndex(summary, ".")
	if lastPeriod > 0 && lastPeriod > maxLength-50 {
		return summary[:lastPeriod+1]
	}

	return summary + "..."
}

// generateHash generates a simple hash for content
func (p *Processor) generateHash(input string) string {
	return fmt.Sprintf("email-%s", input)
}

// Stop stops the email processor
func (p *Processor) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return
	}

	p.logger.Info("Stopping email processor...")
	close(p.stopChan)

	if p.ticker != nil {
		p.ticker.Stop()
	}

	p.wg.Wait()
	p.running = false
	p.logger.Info("Email processor stopped successfully")
}

// IsRunning returns whether the processor is currently running
func (p *Processor) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}

// GetStats returns email processing statistics
func (p *Processor) GetStats(ctx context.Context) (*models.EmailStats, error) {
	return p.emailRepo.GetStats(ctx)
}

// FetchExistingEmails fetches existing emails from the inbox based on configuration
func (p *Processor) FetchExistingEmails(ctx context.Context) (int, error) {
	p.logger.Info("Fetching existing emails from inbox...")

	emails, err := p.emailService.FetchExistingEmails(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch existing emails: %w", err)
	}

	processedCount := 0

	for _, email := range emails {
		// Check if email already exists
		exists, err := p.emailRepo.Exists(ctx, email.MessageID)
		if err != nil {
			p.logger.WithError(err).Warn("Failed to check email existence")
			continue
		}

		if exists {
			p.logger.Debugf("Skipping duplicate email: %s", email.MessageID)
			continue
		}

		// Store email
		storedEmail, err := p.emailRepo.Create(ctx, email)
		if err != nil {
			p.logger.WithError(err).Error("Failed to store email")
			continue
		}

		p.logger.Infof("Stored existing email: %s from %s", storedEmail.Subject, storedEmail.Sender)

		// Process into article if configured
		if p.config.ProcessArticles {
			if err := p.processEmailToArticle(ctx, storedEmail); err != nil {
				p.logger.WithError(err).Errorf("Failed to process existing email %d into article", storedEmail.ID)
				// Mark as failed
				p.emailRepo.MarkAsFailed(ctx, storedEmail.ID, err.Error())
			} else {
				processedCount++
			}
		}
	}

	p.logger.Infof("Successfully processed %d existing emails into articles", processedCount)
	return processedCount, nil
}

// updateEmailStatus updates the status of an email
func (p *Processor) updateEmailStatus(ctx context.Context, emailID int64, status string) error {
	return p.emailRepo.UpdateStatus(ctx, emailID, status)
}

// markEmailFailedWithCode marks an email as failed with error message and code
func (p *Processor) markEmailFailedWithCode(ctx context.Context, emailID int64, errorMsg, errorCode string) {
	if err := p.emailRepo.MarkAsFailedWithCode(ctx, emailID, errorMsg, errorCode); err != nil {
		p.logger.WithError(err).Error("Failed to mark email as failed with code")
	}
}

// updateLastRetryAt updates the last_retry_at timestamp
func (p *Processor) updateLastRetryAt(ctx context.Context, emailID int64) error {
	return p.emailRepo.UpdateLastRetryAt(ctx, emailID)
}
