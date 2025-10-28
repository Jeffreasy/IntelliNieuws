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
		SET status = $1, completed_at = $2, article_count = $3
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

// GetRecentJobs returns recent scraping jobs
func (r *ScrapingJobRepository) GetRecentJobs(ctx context.Context, limit int) ([]*models.ScrapingJob, error) {
	query := `
		SELECT id, source, status, started_at, completed_at, error, article_count, created_at
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
			&job.ArticleCount,
			&job.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}

		if startedAt.Valid {
			job.StartedAt = startedAt.Time
		}
		if completedAt.Valid {
			job.CompletedAt = completedAt.Time
		}
		if errorMsg.Valid {
			job.Error = errorMsg.String
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// GetJobsBySource returns scraping jobs for a specific source
func (r *ScrapingJobRepository) GetJobsBySource(ctx context.Context, source string, limit int) ([]*models.ScrapingJob, error) {
	query := `
		SELECT id, source, status, started_at, completed_at, error, article_count, created_at
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
			&job.ArticleCount,
			&job.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}

		if startedAt.Valid {
			job.StartedAt = startedAt.Time
		}
		if completedAt.Valid {
			job.CompletedAt = completedAt.Time
		}
		if errorMsg.Valid {
			job.Error = errorMsg.String
		}

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
			COALESCE(SUM(article_count), 0) as total_articles,
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
