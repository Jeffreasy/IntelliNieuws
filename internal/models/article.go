package models

import (
	"time"
)

// Article represents a news article
type Article struct {
	ID          int64     `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Summary     string    `json:"summary" db:"summary"`
	URL         string    `json:"url" db:"url"`
	Published   time.Time `json:"published" db:"published"`
	Source      string    `json:"source" db:"source"`
	Keywords    []string  `json:"keywords" db:"keywords"`
	ImageURL    string    `json:"image_url" db:"image_url"`
	Author      string    `json:"author" db:"author"`
	Category    string    `json:"category" db:"category"`
	ContentHash string    `json:"-" db:"content_hash"` // For duplicate detection
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	// Full content extraction fields
	Content            string     `json:"content,omitempty" db:"content"`
	ContentExtracted   bool       `json:"content_extracted" db:"content_extracted"`
	ContentExtractedAt *time.Time `json:"content_extracted_at,omitempty" db:"content_extracted_at"`
}

// ArticleFilter represents filters for querying articles
type ArticleFilter struct {
	Source    string
	Category  string
	Keyword   string
	Search    string
	StartDate *time.Time
	EndDate   *time.Time
	SortBy    string
	SortOrder string
	Limit     int
	Offset    int
}

// ScrapingJob represents a scraping job
type ScrapingJob struct {
	ID           int64     `json:"id" db:"id"`
	Source       string    `json:"source" db:"source"`
	Status       string    `json:"status" db:"status"` // pending, running, completed, failed
	StartedAt    time.Time `json:"started_at" db:"started_at"`
	CompletedAt  time.Time `json:"completed_at" db:"completed_at"`
	Error        string    `json:"error" db:"error"`
	ArticleCount int       `json:"article_count" db:"article_count"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// ScrapingJobStatus constants
const (
	JobStatusPending   = "pending"
	JobStatusRunning   = "running"
	JobStatusCompleted = "completed"
	JobStatusFailed    = "failed"
)

// Source represents a news source configuration
type Source struct {
	ID            int64     `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	Domain        string    `json:"domain" db:"domain"`
	RSSFeedURL    string    `json:"rss_feed_url" db:"rss_feed_url"`
	UseRSS        bool      `json:"use_rss" db:"use_rss"`
	UseDynamic    bool      `json:"use_dynamic" db:"use_dynamic"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	RateLimitSec  int       `json:"rate_limit_sec" db:"rate_limit_sec"`
	LastScrapedAt time.Time `json:"last_scraped_at" db:"last_scraped_at"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// ArticleCreate represents the data needed to create an article
type ArticleCreate struct {
	Title       string    `json:"title" validate:"required,min=3,max=500"`
	Summary     string    `json:"summary" validate:"max=2000"`
	URL         string    `json:"url" validate:"required,url"`
	Published   time.Time `json:"published" validate:"required"`
	Source      string    `json:"source" validate:"required,min=2,max=100"`
	Keywords    []string  `json:"keywords"`
	ImageURL    string    `json:"image_url" validate:"omitempty,url"`
	Author      string    `json:"author" validate:"max=200"`
	Category    string    `json:"category" validate:"max=100"`
	ContentHash string    `json:"-"`
}

// ArticleResponse represents the API response for an article
type ArticleResponse struct {
	Article Article `json:"article"`
}

// ArticleListResponse represents the API response for a list of articles
type ArticleListResponse struct {
	Articles   []Article          `json:"articles"`
	Pagination PaginationResponse `json:"pagination"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}
