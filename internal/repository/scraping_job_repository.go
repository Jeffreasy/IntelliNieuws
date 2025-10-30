package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jeffrey/intellinieuws/internal/models"
	"github.com/jeffrey/intellinieuws/pkg/logger"
)

// ScrapingJobRepository handles database operations for scraping jobs
type ScrapingJobRepository struct {
	db     *pgxpool.Pool
	logger *logger.Logger
}

// NewScrapingJobRepository creates a new scraping job repository
func NewScrapingJobRepository(db *pgxpool.Pool, log *logger.Logger) *ScrapingJobRepository {
	return &ScrapingJobRepository{
		db:     db,
		logger: log.WithComponent("scraping-job-repo"),
	}
}

// CreateJob creates a new scraping job
func (r *ScrapingJobRepository) CreateJob(ctx context.Context, source string) (int64, error) {
	query := `
		INSERT INTO scraping_jobs (source, status, created_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(ctx, query, source, models.JobStatusPending, time.Now()).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create scraping job: %w", err)
	}

	r.logger.Debugf("Created scraping job %d for source %s", id, source)
	return id, nil
}

// CreateJobWithDetails creates a new scraping job with UUID and method
func (r *ScrapingJobRepository) CreateJobWithDetails(ctx context.Context, source, jobUUID, scrapingMethod string) (int64, error) {
	query := `
		INSERT INTO scraping_jobs (source, job_uuid, scraping_method, status, max_retries, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(ctx, query,
		source,
		jobUUID,
		scrapingMethod,
		models.JobStatusPending,
		3, // max_retries
		time.Now(),
		"scraper-service",
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create scraping job with details: %w", err)
	}

	r.logger.Debugf("Created scraping job %d (UUID: %s) for source %s with method %s", id, jobUUID, source, scrapingMethod)
	return id, nil
}

// StartJob marks a job as running
func (r *ScrapingJobRepository) StartJob(ctx context.Context, jobID int64) error {
	query := `
		UPDATE scraping_jobs
		SET status = $1, started_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, models.JobStatusRunning, time.Now(), jobID)
	if err != nil {
		return fmt.Errorf("failed to start job: %w", err)
	}

	r.logger.Debugf("Started scraping job %d", jobID)
	return nil
}

// CompleteJob marks a job as completed
func (r *ScrapingJobRepository) CompleteJob(ctx context.Context, jobID int64, articleCount int) error {
	query := `
		UPDATE scraping_jobs
		SET status = $1, completed_at = $2, articles_new = $3
		WHERE id = $4
	`

	_, err := r.db.Exec(ctx, query, models.JobStatusCompleted, time.Now(), articleCount, jobID)
	if err != nil {
		return fmt.Errorf("failed to complete job: %w", err)
	}

	r.logger.Debugf("Completed scraping job %d with %d articles", jobID, articleCount)
	return nil
}

// FailJob marks a job as failed with an error message
func (r *ScrapingJobRepository) FailJob(ctx context.Context, jobID int64, errorMsg string) error {
	query := `
		UPDATE scraping_jobs
		SET status = $1, completed_at = $2, error = $3
		WHERE id = $4
	`

	_, err := r.db.Exec(ctx, query, models.JobStatusFailed, time.Now(), errorMsg, jobID)
	if err != nil {
		return fmt.Errorf("failed to mark job as failed: %w", err)
	}

	r.logger.Debugf("Failed scraping job %d: %s", jobID, errorMsg)
	return nil
}

// FailJobWithDetails marks a job as failed with error message, code and execution time
func (r *ScrapingJobRepository) FailJobWithDetails(ctx context.Context, jobID int64, errorMsg, errorCode string, executionTimeMs int) error {
	query := `
		UPDATE scraping_jobs
		SET status = $1, completed_at = $2, error = $3, error_code = $4,
		    execution_time_ms = $5, retry_count = retry_count + 1
		WHERE id = $6
	`

	_, err := r.db.Exec(ctx, query,
		models.JobStatusFailed,
		time.Now(),
		errorMsg,
		errorCode,
		executionTimeMs,
		jobID,
	)
	if err != nil {
		return fmt.Errorf("failed to mark job as failed with details: %w", err)
	}

	r.logger.Debugf("Failed scraping job %d with code %s: %s", jobID, errorCode, errorMsg)
	return nil
}

// CompleteJobWithDetails marks a job as completed with detailed statistics
func (r *ScrapingJobRepository) CompleteJobWithDetails(ctx context.Context, jobID int64, articlesFound, articlesNew, articlesUpdated, articlesSkipped, executionTimeMs int) error {
	query := `
		UPDATE scraping_jobs
		SET status = $1, completed_at = $2,
		    articles_found = $3, articles_new = $4, articles_updated = $5, articles_skipped = $6,
		    execution_time_ms = $7
		WHERE id = $8
	`

	_, err := r.db.Exec(ctx, query,
		models.JobStatusCompleted,
		time.Now(),
		articlesFound,
		articlesNew,
		articlesUpdated,
		articlesSkipped,
		executionTimeMs,
		jobID,
	)
	if err != nil {
		return fmt.Errorf("failed to complete job with details: %w", err)
	}

	r.logger.Debugf("Completed scraping job %d: found=%d, new=%d, updated=%d, skipped=%d, time=%dms",
		jobID, articlesFound, articlesNew, articlesUpdated, articlesSkipped, executionTimeMs)
	return nil
}

// GetRecentJobs returns recent scraping jobs
func (r *ScrapingJobRepository) GetRecentJobs(ctx context.Context, limit int) ([]*models.ScrapingJob, error) {
	query := `
		SELECT id, source, status, started_at, completed_at, error, articles_new, created_at
		FROM scraping_jobs
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*models.ScrapingJob
	for rows.Next() {
		job := &models.ScrapingJob{}
		var startedAt, completedAt sql.NullTime
		var errorMsg sql.NullString

		err := rows.Scan(
			&job.ID,
			&job.Source,
			&job.Status,
			&startedAt,
			&completedAt,
			&errorMsg,
			&job.ArticlesNew,
			&job.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}

		if startedAt.Valid {
			t := startedAt.Time
			job.StartedAt = &t
		}
		if completedAt.Valid {
			t := completedAt.Time
			job.CompletedAt = &t
		}
		if errorMsg.Valid {
			job.Error = errorMsg.String
		}

		// Backwards compatibility
		job.ArticleCount = job.ArticlesNew

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// GetJobsBySource returns scraping jobs for a specific source
func (r *ScrapingJobRepository) GetJobsBySource(ctx context.Context, source string, limit int) ([]*models.ScrapingJob, error) {
	query := `
		SELECT id, source, status, started_at, completed_at, error, articles_new, created_at
		FROM scraping_jobs
		WHERE source = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, source, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by source: %w", err)
	}
	defer rows.Close()

	var jobs []*models.ScrapingJob
	for rows.Next() {
		job := &models.ScrapingJob{}
		var startedAt, completedAt sql.NullTime
		var errorMsg sql.NullString

		err := rows.Scan(
			&job.ID,
			&job.Source,
			&job.Status,
			&startedAt,
			&completedAt,
			&errorMsg,
			&job.ArticlesNew,
			&job.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}

		if startedAt.Valid {
			t := startedAt.Time
			job.StartedAt = &t
		}
		if completedAt.Valid {
			t := completedAt.Time
			job.CompletedAt = &t
		}
		if errorMsg.Valid {
			job.Error = errorMsg.String
		}

		// Backwards compatibility
		job.ArticleCount = job.ArticlesNew

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// GetJobStats returns statistics about scraping jobs
func (r *ScrapingJobRepository) GetJobStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'completed') as completed,
			COUNT(*) FILTER (WHERE status = 'failed') as failed,
			COUNT(*) FILTER (WHERE status = 'running') as running,
			COUNT(*) FILTER (WHERE status = 'pending') as pending,
			COALESCE(SUM(articles_new), 0) as total_articles,
			COALESCE(AVG(EXTRACT(EPOCH FROM (completed_at - started_at))), 0) as avg_duration_seconds
		FROM scraping_jobs
		WHERE created_at > NOW() - INTERVAL '24 hours'
	`

	var stats struct {
		Total              int64
		Completed          int64
		Failed             int64
		Running            int64
		Pending            int64
		TotalArticles      int64
		AvgDurationSeconds float64
	}

	err := r.db.QueryRow(ctx, query).Scan(
		&stats.Total,
		&stats.Completed,
		&stats.Failed,
		&stats.Running,
		&stats.Pending,
		&stats.TotalArticles,
		&stats.AvgDurationSeconds,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get job stats: %w", err)
	}

	return map[string]interface{}{
		"last_24h": map[string]interface{}{
			"total":                stats.Total,
			"completed":            stats.Completed,
			"failed":               stats.Failed,
			"running":              stats.Running,
			"pending":              stats.Pending,
			"total_articles":       stats.TotalArticles,
			"avg_duration_seconds": stats.AvgDurationSeconds,
		},
	}, nil
}

// UpdateSourceMetadata updates the source's last_scraped_at and total_articles_scraped
func (r *ScrapingJobRepository) UpdateSourceMetadata(ctx context.Context, source string, articlesScraped int, success bool) error {
	var query string

	if success {
		// Update last_success_at and reset consecutive failures on success
		query = `
			UPDATE sources
			SET last_scraped_at = $1,
			    last_success_at = $1,
			    total_articles_scraped = total_articles_scraped + $2,
			    consecutive_failures = 0,
			    last_error = NULL
			WHERE domain = $3
		`
	} else {
		// Only update last_scraped_at and increment consecutive failures on failure
		query = `
			UPDATE sources
			SET last_scraped_at = $1,
			    consecutive_failures = consecutive_failures + 1
			WHERE domain = $2
		`
	}

	var err error
	if success {
		_, err = r.db.Exec(ctx, query, time.Now(), articlesScraped, source)
	} else {
		_, err = r.db.Exec(ctx, query, time.Now(), source)
	}

	if err != nil {
		return fmt.Errorf("failed to update source metadata: %w", err)
	}

	if success {
		r.logger.Debugf("Updated source %s: scraped %d articles, reset failures", source, articlesScraped)
	} else {
		r.logger.Debugf("Updated source %s: incremented consecutive failures", source)
	}
	return nil
}

// UpdateSourceError updates the source with an error message
func (r *ScrapingJobRepository) UpdateSourceError(ctx context.Context, source string, errorMsg string) error {
	query := `
		UPDATE sources
		SET last_scraped_at = $1,
		    last_error = $2,
		    consecutive_failures = consecutive_failures + 1
		WHERE domain = $3
	`

	_, err := r.db.Exec(ctx, query, time.Now(), errorMsg, source)
	if err != nil {
		return fmt.Errorf("failed to update source error: %w", err)
	}

	r.logger.Debugf("Updated source %s with error: %s", source, errorMsg)
	return nil
}
