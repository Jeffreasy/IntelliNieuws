package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Email represents an email message
type Email struct {
	ID              int64         `json:"id" db:"id"`
	MessageID       string        `json:"message_id" db:"message_id"`
	MessageUID      string        `json:"message_uid,omitempty" db:"message_uid"`
	ThreadID        string        `json:"thread_id,omitempty" db:"thread_id"`
	Sender          string        `json:"sender" db:"sender"`
	SenderName      string        `json:"sender_name,omitempty" db:"sender_name"`
	Recipient       string        `json:"recipient,omitempty" db:"recipient"`
	Subject         string        `json:"subject" db:"subject"`
	BodyText        string        `json:"body_text" db:"body_text"`
	BodyHTML        string        `json:"body_html" db:"body_html"`
	Snippet         string        `json:"snippet,omitempty" db:"snippet"`
	ReceivedDate    time.Time     `json:"received_date" db:"received_date"`
	SentDate        *time.Time    `json:"sent_date,omitempty" db:"sent_date"`
	Status          string        `json:"status" db:"status"` // pending, processing, processed, failed, ignored, spam
	ProcessedAt     *time.Time    `json:"processed_at,omitempty" db:"processed_at"`
	ArticleID       *int64        `json:"article_id,omitempty" db:"article_id"`
	ArticleCreated  bool          `json:"article_created" db:"article_created"`
	Error           string        `json:"error,omitempty" db:"error"`
	ErrorCode       string        `json:"error_code,omitempty" db:"error_code"`
	RetryCount      int           `json:"retry_count" db:"retry_count"`
	MaxRetries      int           `json:"max_retries" db:"max_retries"`
	LastRetryAt     *time.Time    `json:"last_retry_at,omitempty" db:"last_retry_at"`
	HasAttachments  bool          `json:"has_attachments" db:"has_attachments"`
	AttachmentCount int           `json:"attachment_count" db:"attachment_count"`
	IsRead          bool          `json:"is_read" db:"is_read"`
	IsFlagged       bool          `json:"is_flagged" db:"is_flagged"`
	IsSpam          bool          `json:"is_spam" db:"is_spam"`
	Importance      string        `json:"importance,omitempty" db:"importance"`
	Metadata        EmailMetadata `json:"metadata" db:"metadata"`
	Headers         EmailMetadata `json:"headers,omitempty" db:"headers"`
	Labels          []string      `json:"labels,omitempty" db:"labels"`
	SizeBytes       *int          `json:"size_bytes,omitempty" db:"size_bytes"`
	SpamScore       *float64      `json:"spam_score,omitempty" db:"spam_score"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
	CreatedBy       string        `json:"created_by,omitempty" db:"created_by"`
}

// EmailMetadata contains additional email metadata
type EmailMetadata map[string]interface{}

// Value implements the driver.Valuer interface for database storage
func (m EmailMetadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for database retrieval
func (m *EmailMetadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(EmailMetadata)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, m)
}

// EmailCreate represents the data needed to create an email
type EmailCreate struct {
	MessageID    string        `json:"message_id" validate:"required"`
	Sender       string        `json:"sender" validate:"required,email"`
	Subject      string        `json:"subject" validate:"required"`
	BodyText     string        `json:"body_text"`
	BodyHTML     string        `json:"body_html"`
	ReceivedDate time.Time     `json:"received_date" validate:"required"`
	Metadata     EmailMetadata `json:"metadata"`
}

// EmailFilter represents filters for querying emails
type EmailFilter struct {
	Sender    string
	Processed *bool
	StartDate *time.Time
	EndDate   *time.Time
	HasError  bool
	SortBy    string
	SortOrder string
	Limit     int
	Offset    int
}

// EmailStats represents email processing statistics
type EmailStats struct {
	TotalEmails     int `json:"total_emails"`
	ProcessedEmails int `json:"processed_emails"`
	PendingEmails   int `json:"pending_emails"`
	FailedEmails    int `json:"failed_emails"`
	ArticlesCreated int `json:"articles_created"`
}

// EmailResponse represents the API response for an email
type EmailResponse struct {
	Email Email `json:"email"`
}

// EmailListResponse represents the API response for a list of emails
type EmailListResponse struct {
	Emails     []Email            `json:"emails"`
	Pagination PaginationResponse `json:"pagination"`
}
