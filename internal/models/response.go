package models

import (
	"time"
)

// APIResponse is a standardized wrapper for all API responses
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// APIError represents error details in API responses
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Meta contains metadata for paginated responses
type Meta struct {
	Pagination *PaginationMeta `json:"pagination,omitempty"`
	Sorting    *SortingMeta    `json:"sorting,omitempty"`
	Filtering  *FilteringMeta  `json:"filtering,omitempty"`
}

// PaginationMeta contains enhanced pagination metadata
type PaginationMeta struct {
	Total       int  `json:"total"`
	Limit       int  `json:"limit"`
	Offset      int  `json:"offset"`
	CurrentPage int  `json:"current_page"`
	TotalPages  int  `json:"total_pages"`
	HasNext     bool `json:"has_next"`
	HasPrev     bool `json:"has_prev"`
}

// SortingMeta contains sorting information
type SortingMeta struct {
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
}

// FilteringMeta contains active filter information
type FilteringMeta struct {
	Source    string `json:"source,omitempty"`
	Category  string `json:"category,omitempty"`
	Keyword   string `json:"keyword,omitempty"`
	Search    string `json:"search,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Service   string                 `json:"service"`
	Version   string                 `json:"version"`
	Timestamp string                 `json:"timestamp"`
	Checks    map[string]HealthCheck `json:"checks"`
}

// HealthCheck represents a single health check result
type HealthCheck struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// SourceInfo represents information about a news source
type SourceInfo struct {
	Name         string `json:"name"`
	Domain       string `json:"domain"`
	FeedURL      string `json:"feed_url"`
	ArticleCount int    `json:"article_count"`
	IsActive     bool   `json:"is_active"`
}

// CategoryInfo represents information about a category
type CategoryInfo struct {
	Name         string `json:"name"`
	ArticleCount int    `json:"article_count"`
}

// StatsResponse represents statistics response
type StatsResponse struct {
	TotalArticles    int                     `json:"total_articles"`
	ArticlesBySource map[string]int          `json:"articles_by_source"`
	RecentArticles   int                     `json:"recent_articles_24h"`
	OldestArticle    *time.Time              `json:"oldest_article,omitempty"`
	NewestArticle    *time.Time              `json:"newest_article,omitempty"`
	Categories       map[string]CategoryInfo `json:"categories,omitempty"`
}

// NewSuccessResponse creates a successful API response
func NewSuccessResponse(data interface{}, requestID string) APIResponse {
	return APIResponse{
		Success:   true,
		Data:      data,
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// NewSuccessResponseWithMeta creates a successful API response with metadata
func NewSuccessResponseWithMeta(data interface{}, meta *Meta, requestID string) APIResponse {
	return APIResponse{
		Success:   true,
		Data:      data,
		Meta:      meta,
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// NewErrorResponse creates an error API response
func NewErrorResponse(code, message, details, requestID string) APIResponse {
	return APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// CalculatePaginationMeta calculates pagination metadata
func CalculatePaginationMeta(total, limit, offset int) *PaginationMeta {
	if limit == 0 {
		limit = 50
	}

	currentPage := (offset / limit) + 1
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	return &PaginationMeta{
		Total:       total,
		Limit:       limit,
		Offset:      offset,
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		HasNext:     offset+limit < total,
		HasPrev:     offset > 0,
	}
}
