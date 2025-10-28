package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jeffrey/intellinieuws/internal/models"
)

// EmailRepository handles email data operations
type EmailRepository struct {
	db *pgxpool.Pool
}

// NewEmailRepository creates a new email repository
func NewEmailRepository(db *pgxpool.Pool) *EmailRepository {
	return &EmailRepository{db: db}
}

// Create creates a new email record
func (r *EmailRepository) Create(ctx context.Context, email *models.EmailCreate) (*models.Email, error) {
	query := `
		INSERT INTO emails (message_id, sender, subject, body_text, body_html, received_date, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, message_id, sender, subject, body_text, body_html, received_date, 
		          processed, processed_at, article_id, error, retry_count, metadata, created_at, updated_at
	`

	var result models.Email
	err := r.db.QueryRow(
		ctx, query,
		email.MessageID,
		email.Sender,
		email.Subject,
		email.BodyText,
		email.BodyHTML,
		email.ReceivedDate,
		email.Metadata,
	).Scan(
		&result.ID,
		&result.MessageID,
		&result.Sender,
		&result.Subject,
		&result.BodyText,
		&result.BodyHTML,
		&result.ReceivedDate,
		&result.Processed,
		&result.ProcessedAt,
		&result.ArticleID,
		&result.Error,
		&result.RetryCount,
		&result.Metadata,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create email: %w", err)
	}

	return &result, nil
}

// GetByID retrieves an email by ID
func (r *EmailRepository) GetByID(ctx context.Context, id int64) (*models.Email, error) {
	query := `
		SELECT id, message_id, sender, subject, body_text, body_html, received_date,
		       processed, processed_at, article_id, error, retry_count, metadata, created_at, updated_at
		FROM emails
		WHERE id = $1
	`

	var email models.Email
	err := r.db.QueryRow(ctx, query, id).Scan(
		&email.ID,
		&email.MessageID,
		&email.Sender,
		&email.Subject,
		&email.BodyText,
		&email.BodyHTML,
		&email.ReceivedDate,
		&email.Processed,
		&email.ProcessedAt,
		&email.ArticleID,
		&email.Error,
		&email.RetryCount,
		&email.Metadata,
		&email.CreatedAt,
		&email.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get email: %w", err)
	}

	return &email, nil
}

// GetByMessageID retrieves an email by message ID (for deduplication)
func (r *EmailRepository) GetByMessageID(ctx context.Context, messageID string) (*models.Email, error) {
	query := `
		SELECT id, message_id, sender, subject, body_text, body_html, received_date,
		       processed, processed_at, article_id, error, retry_count, metadata, created_at, updated_at
		FROM emails
		WHERE message_id = $1
	`

	var email models.Email
	err := r.db.QueryRow(ctx, query, messageID).Scan(
		&email.ID,
		&email.MessageID,
		&email.Sender,
		&email.Subject,
		&email.BodyText,
		&email.BodyHTML,
		&email.ReceivedDate,
		&email.Processed,
		&email.ProcessedAt,
		&email.ArticleID,
		&email.Error,
		&email.RetryCount,
		&email.Metadata,
		&email.CreatedAt,
		&email.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get email by message ID: %w", err)
	}

	return &email, nil
}

// List retrieves emails with filtering
func (r *EmailRepository) List(ctx context.Context, filter *models.EmailFilter) ([]models.Email, int, error) {
	// Build query
	query := `
		SELECT id, message_id, sender, subject, body_text, body_html, received_date,
		       processed, processed_at, article_id, error, retry_count, metadata, created_at, updated_at
		FROM emails
		WHERE 1=1
	`
	countQuery := "SELECT COUNT(*) FROM emails WHERE 1=1"

	args := make([]interface{}, 0)
	argIndex := 1

	// Apply filters
	if filter.Sender != "" {
		query += fmt.Sprintf(" AND sender = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND sender = $%d", argIndex)
		args = append(args, filter.Sender)
		argIndex++
	}

	if filter.Processed != nil {
		query += fmt.Sprintf(" AND processed = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND processed = $%d", argIndex)
		args = append(args, *filter.Processed)
		argIndex++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND received_date >= $%d", argIndex)
		countQuery += fmt.Sprintf(" AND received_date >= $%d", argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND received_date <= $%d", argIndex)
		countQuery += fmt.Sprintf(" AND received_date <= $%d", argIndex)
		args = append(args, *filter.EndDate)
		argIndex++
	}

	if filter.HasError {
		query += " AND error IS NOT NULL AND error != ''"
		countQuery += " AND error IS NOT NULL AND error != ''"
	}

	// Get total count
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count emails: %w", err)
	}

	// Apply sorting
	sortBy := "received_date"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := "DESC"
	if filter.SortOrder != "" {
		sortOrder = strings.ToUpper(filter.SortOrder)
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Apply pagination
	limit := 50
	if filter.Limit > 0 {
		limit = filter.Limit
	}
	offset := 0
	if filter.Offset > 0 {
		offset = filter.Offset
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list emails: %w", err)
	}
	defer rows.Close()

	emails := make([]models.Email, 0)
	for rows.Next() {
		var email models.Email
		err := rows.Scan(
			&email.ID,
			&email.MessageID,
			&email.Sender,
			&email.Subject,
			&email.BodyText,
			&email.BodyHTML,
			&email.ReceivedDate,
			&email.Processed,
			&email.ProcessedAt,
			&email.ArticleID,
			&email.Error,
			&email.RetryCount,
			&email.Metadata,
			&email.CreatedAt,
			&email.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan email: %w", err)
		}
		emails = append(emails, email)
	}

	return emails, total, nil
}

// MarkAsProcessed marks an email as processed
func (r *EmailRepository) MarkAsProcessed(ctx context.Context, emailID int64, articleID *int64) error {
	query := `
		UPDATE emails
		SET processed = true, processed_at = $1, article_id = $2, error = NULL
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, time.Now(), articleID, emailID)
	if err != nil {
		return fmt.Errorf("failed to mark email as processed: %w", err)
	}

	return nil
}

// MarkAsFailed marks an email as failed with error message
func (r *EmailRepository) MarkAsFailed(ctx context.Context, emailID int64, errorMsg string) error {
	query := `
		UPDATE emails
		SET error = $1, retry_count = retry_count + 1
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, errorMsg, emailID)
	if err != nil {
		return fmt.Errorf("failed to mark email as failed: %w", err)
	}

	return nil
}

// GetUnprocessed retrieves unprocessed emails for retry
func (r *EmailRepository) GetUnprocessed(ctx context.Context, maxRetries int, limit int) ([]models.Email, error) {
	query := `
		SELECT id, message_id, sender, subject, body_text, body_html, received_date,
		       processed, processed_at, article_id, error, retry_count, metadata, created_at, updated_at
		FROM emails
		WHERE processed = false AND retry_count < $1
		ORDER BY received_date DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, maxRetries, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get unprocessed emails: %w", err)
	}
	defer rows.Close()

	emails := make([]models.Email, 0)
	for rows.Next() {
		var email models.Email
		err := rows.Scan(
			&email.ID,
			&email.MessageID,
			&email.Sender,
			&email.Subject,
			&email.BodyText,
			&email.BodyHTML,
			&email.ReceivedDate,
			&email.Processed,
			&email.ProcessedAt,
			&email.ArticleID,
			&email.Error,
			&email.RetryCount,
			&email.Metadata,
			&email.CreatedAt,
			&email.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan email: %w", err)
		}
		emails = append(emails, email)
	}

	return emails, nil
}

// GetStats retrieves email processing statistics
func (r *EmailRepository) GetStats(ctx context.Context) (*models.EmailStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_emails,
			COUNT(*) FILTER (WHERE processed = true) as processed_emails,
			COUNT(*) FILTER (WHERE processed = false) as pending_emails,
			COUNT(*) FILTER (WHERE error IS NOT NULL AND error != '') as failed_emails,
			COUNT(*) FILTER (WHERE article_id IS NOT NULL) as articles_created
		FROM emails
	`

	var stats models.EmailStats
	err := r.db.QueryRow(ctx, query).Scan(
		&stats.TotalEmails,
		&stats.ProcessedEmails,
		&stats.PendingEmails,
		&stats.FailedEmails,
		&stats.ArticlesCreated,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get email stats: %w", err)
	}

	return &stats, nil
}

// Delete deletes an email by ID
func (r *EmailRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM emails WHERE id = $1"

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete email: %w", err)
	}

	return nil
}

// Exists checks if an email with the given message ID already exists
func (r *EmailRepository) Exists(ctx context.Context, messageID string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM emails WHERE message_id = $1)"

	var exists bool
	err := r.db.QueryRow(ctx, query, messageID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}
