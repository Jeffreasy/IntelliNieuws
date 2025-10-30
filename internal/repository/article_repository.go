package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jeffrey/intellinieuws/internal/models"
)

// ArticleRepository handles database operations for articles
type ArticleRepository struct {
	db *pgxpool.Pool
}

// NewArticleRepository creates a new article repository
func NewArticleRepository(db *pgxpool.Pool) *ArticleRepository {
	return &ArticleRepository{db: db}
}

// Create inserts a new article into the database
func (r *ArticleRepository) Create(ctx context.Context, article *models.ArticleCreate) (*models.Article, error) {
	// Generate content hash for duplicate detection
	article.ContentHash = generateContentHash(article.Title, article.URL)

	// Sanitize text fields to prevent UTF-8 encoding errors
	article.Title = sanitizeUTF8(article.Title)
	article.Summary = sanitizeUTF8(article.Summary)
	article.Author = sanitizeUTF8(article.Author)
	article.Category = sanitizeUTF8(article.Category)

	query := `
		INSERT INTO articles (title, summary, url, published, source, keywords, image_url, author, category, content_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`

	var result models.Article
	err := r.db.QueryRow(ctx, query,
		article.Title,
		article.Summary,
		article.URL,
		article.Published,
		article.Source,
		article.Keywords,
		article.ImageURL,
		article.Author,
		article.Category,
		article.ContentHash,
	).Scan(&result.ID, &result.CreatedAt, &result.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create article: %w", err)
	}

	// Populate result with input data
	result.Title = article.Title
	result.Summary = article.Summary
	result.URL = article.URL
	result.Published = article.Published
	result.Source = article.Source
	result.Keywords = article.Keywords
	result.ImageURL = article.ImageURL
	result.Author = article.Author
	result.Category = article.Category
	result.ContentHash = article.ContentHash

	return &result, nil
}

// CreateBatch inserts multiple articles in a single transaction for better performance
func (r *ArticleRepository) CreateBatch(ctx context.Context, articles []*models.ArticleCreate) (int, error) {
	if len(articles) == 0 {
		return 0, nil
	}

	// Use batch without explicit transaction for better concurrency
	batch := &pgx.Batch{}
	query := `
		INSERT INTO articles (title, summary, url, published, source, keywords, image_url, author, category, content_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (url) DO NOTHING
		RETURNING id
	`

	for _, article := range articles {
		// Generate content hash
		article.ContentHash = generateContentHash(article.Title, article.URL)

		batch.Queue(query,
			article.Title,
			article.Summary,
			article.URL,
			article.Published,
			article.Source,
			article.Keywords,
			article.ImageURL,
			article.Author,
			article.Category,
			article.ContentHash,
		)
	}

	// Execute batch with pool (automatically handles concurrency)
	results := r.db.SendBatch(ctx, batch)
	defer results.Close()

	// Count successful inserts (those that return an ID)
	inserted := 0
	for i := 0; i < len(articles); i++ {
		var id int64
		err := results.QueryRow().Scan(&id)
		if err == nil {
			inserted++
		}
		// ON CONFLICT returns no rows, that's OK
	}

	return inserted, nil
}

// GetByID retrieves an article by ID (includes content if extracted)
func (r *ArticleRepository) GetByID(ctx context.Context, id int64) (*models.Article, error) {
	query := `
		SELECT id, title, summary, url, published, source, keywords, image_url,
		       author, category, content_hash, created_at, updated_at,
		       COALESCE(content, '') as content,
		       COALESCE(content_extracted, FALSE) as content_extracted,
		       content_extracted_at
		FROM articles
		WHERE id = $1
	`

	var article models.Article
	var content string
	var contentExtracted bool
	var contentExtractedAt *time.Time

	err := r.db.QueryRow(ctx, query, id).Scan(
		&article.ID,
		&article.Title,
		&article.Summary,
		&article.URL,
		&article.Published,
		&article.Source,
		&article.Keywords,
		&article.ImageURL,
		&article.Author,
		&article.Category,
		&article.ContentHash,
		&article.CreatedAt,
		&article.UpdatedAt,
		&content,
		&contentExtracted,
		&contentExtractedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("article not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	// Set content fields
	article.Content = content
	article.ContentExtracted = contentExtracted
	article.ContentExtractedAt = contentExtractedAt

	return &article, nil
}

// ListLight retrieves articles WITHOUT full content (optimized for list views)
// Use this for API list endpoints to avoid transferring large content fields
func (r *ArticleRepository) ListLight(ctx context.Context, filter models.ArticleFilter) ([]models.Article, int, error) {
	// Lightweight query - exclude content field for performance
	query := `
		SELECT id, title, summary, url, published, source, keywords, image_url,
		       author, category, content_hash, created_at, updated_at,
		       COALESCE(content_extracted, FALSE) as content_extracted,
		       content_extracted_at
		FROM articles
		WHERE 1=1
	`
	countQuery := "SELECT COUNT(*) FROM articles WHERE 1=1"
	args := []interface{}{}
	argPos := 1

	// Apply filters
	if filter.Source != "" {
		query += fmt.Sprintf(" AND source = $%d", argPos)
		countQuery += fmt.Sprintf(" AND source = $%d", argPos)
		args = append(args, filter.Source)
		argPos++
	}

	if filter.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argPos)
		countQuery += fmt.Sprintf(" AND category = $%d", argPos)
		args = append(args, filter.Category)
		argPos++
	}

	if filter.Keyword != "" {
		query += fmt.Sprintf(" AND $%d = ANY(keywords)", argPos)
		countQuery += fmt.Sprintf(" AND $%d = ANY(keywords)", argPos)
		args = append(args, filter.Keyword)
		argPos++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND published >= $%d", argPos)
		countQuery += fmt.Sprintf(" AND published >= $%d", argPos)
		args = append(args, filter.StartDate)
		argPos++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND published <= $%d", argPos)
		countQuery += fmt.Sprintf(" AND published <= $%d", argPos)
		args = append(args, filter.EndDate)
		argPos++
	}

	// Get total count
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count articles: %w", err)
	}

	// Apply ordering
	orderBy := "published"
	if filter.SortBy != "" {
		orderBy = filter.SortBy
	}
	orderDir := "DESC"
	if filter.SortOrder == "asc" {
		orderDir = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", orderBy, orderDir)

	// Apply pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, filter.Limit)
		argPos++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, filter.Offset)
		argPos++
	}

	// Execute query
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list articles: %w", err)
	}
	defer rows.Close()

	// Parse results
	articles := []models.Article{}
	for rows.Next() {
		var article models.Article
		var contentExtracted bool
		var contentExtractedAt *time.Time

		err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Summary,
			&article.URL,
			&article.Published,
			&article.Source,
			&article.Keywords,
			&article.ImageURL,
			&article.Author,
			&article.Category,
			&article.ContentHash,
			&article.CreatedAt,
			&article.UpdatedAt,
			&contentExtracted,
			&contentExtractedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan article: %w", err)
		}

		// Set content fields (no content in lightweight query)
		article.ContentExtracted = contentExtracted
		article.ContentExtractedAt = contentExtractedAt

		articles = append(articles, article)
	}

	return articles, total, nil
}

// List retrieves articles with filters and sorting (includes full content)
// For list views, prefer ListLight() for better performance
func (r *ArticleRepository) List(ctx context.Context, filter models.ArticleFilter) ([]models.Article, int, error) {
	// Build dynamic query (include content fields)
	query := `
		SELECT id, title, summary, url, published, source, keywords, image_url,
		       author, category, content_hash, created_at, updated_at,
		       COALESCE(content, '') as content,
		       COALESCE(content_extracted, FALSE) as content_extracted,
		       content_extracted_at
		FROM articles
		WHERE 1=1
	`
	countQuery := "SELECT COUNT(*) FROM articles WHERE 1=1"
	args := []interface{}{}
	argPos := 1

	// Apply filters
	if filter.Source != "" {
		query += fmt.Sprintf(" AND source = $%d", argPos)
		countQuery += fmt.Sprintf(" AND source = $%d", argPos)
		args = append(args, filter.Source)
		argPos++
	}

	if filter.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argPos)
		countQuery += fmt.Sprintf(" AND category = $%d", argPos)
		args = append(args, filter.Category)
		argPos++
	}

	if filter.Keyword != "" {
		query += fmt.Sprintf(" AND $%d = ANY(keywords)", argPos)
		countQuery += fmt.Sprintf(" AND $%d = ANY(keywords)", argPos)
		args = append(args, filter.Keyword)
		argPos++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND published >= $%d", argPos)
		countQuery += fmt.Sprintf(" AND published >= $%d", argPos)
		args = append(args, filter.StartDate)
		argPos++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND published <= $%d", argPos)
		countQuery += fmt.Sprintf(" AND published <= $%d", argPos)
		args = append(args, filter.EndDate)
		argPos++
	}

	// Get total count
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count articles: %w", err)
	}

	// Apply ordering
	orderBy := "published"
	if filter.SortBy != "" {
		orderBy = filter.SortBy
	}
	orderDir := "DESC"
	if filter.SortOrder == "asc" {
		orderDir = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", orderBy, orderDir)

	// Apply pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, filter.Limit)
		argPos++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, filter.Offset)
		argPos++
	}

	// Execute query
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list articles: %w", err)
	}
	defer rows.Close()

	// Parse results
	articles := []models.Article{}
	for rows.Next() {
		var article models.Article
		var content string
		var contentExtracted bool
		var contentExtractedAt *time.Time

		err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Summary,
			&article.URL,
			&article.Published,
			&article.Source,
			&article.Keywords,
			&article.ImageURL,
			&article.Author,
			&article.Category,
			&article.ContentHash,
			&article.CreatedAt,
			&article.UpdatedAt,
			&content,
			&contentExtracted,
			&contentExtractedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan article: %w", err)
		}

		// Set content fields
		article.Content = content
		article.ContentExtracted = contentExtracted
		article.ContentExtractedAt = contentExtractedAt

		articles = append(articles, article)
	}

	return articles, total, nil
}

// SearchLight performs full-text search WITHOUT full content (optimized for list views)
func (r *ArticleRepository) SearchLight(ctx context.Context, filter models.ArticleFilter) ([]models.Article, int, error) {
	// Lightweight search query - exclude content field for performance
	query := `
		SELECT id, title, summary, url, published, source, keywords, image_url,
		       author, category, content_hash, created_at, updated_at,
		       COALESCE(content_extracted, FALSE) as content_extracted,
		       content_extracted_at
		FROM articles
		WHERE (
			to_tsvector('english', title || ' ' || COALESCE(summary, ''))
			@@ plainto_tsquery('english', $1)
			OR title ILIKE $2
			OR summary ILIKE $2
		)
	`
	countQuery := `
		SELECT COUNT(*)
		FROM articles
		WHERE (
			to_tsvector('english', title || ' ' || COALESCE(summary, ''))
			@@ plainto_tsquery('english', $1)
			OR title ILIKE $2
			OR summary ILIKE $2
		)
	`

	searchPattern := "%" + filter.Search + "%"
	args := []interface{}{filter.Search, searchPattern}
	argPos := 3

	// Apply additional filters
	if filter.Source != "" {
		query += fmt.Sprintf(" AND source = $%d", argPos)
		countQuery += fmt.Sprintf(" AND source = $%d", argPos)
		args = append(args, filter.Source)
		argPos++
	}

	if filter.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argPos)
		countQuery += fmt.Sprintf(" AND category = $%d", argPos)
		args = append(args, filter.Category)
		argPos++
	}

	// Get total count
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Apply ordering
	orderBy := "published"
	if filter.SortBy != "" {
		orderBy = filter.SortBy
	}
	orderDir := "DESC"
	if filter.SortOrder == "asc" {
		orderDir = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", orderBy, orderDir)

	// Apply pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, filter.Limit)
		argPos++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, filter.Offset)
	}

	// Execute query
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search articles: %w", err)
	}
	defer rows.Close()

	// Parse results
	articles := []models.Article{}
	for rows.Next() {
		var article models.Article
		var contentExtracted bool
		var contentExtractedAt *time.Time

		err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Summary,
			&article.URL,
			&article.Published,
			&article.Source,
			&article.Keywords,
			&article.ImageURL,
			&article.Author,
			&article.Category,
			&article.ContentHash,
			&article.CreatedAt,
			&article.UpdatedAt,
			&contentExtracted,
			&contentExtractedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan article: %w", err)
		}

		// Set content fields (no content in lightweight query)
		article.ContentExtracted = contentExtracted
		article.ContentExtractedAt = contentExtractedAt

		articles = append(articles, article)
	}

	return articles, total, nil
}

// Search performs full-text search on articles
func (r *ArticleRepository) Search(ctx context.Context, filter models.ArticleFilter) ([]models.Article, int, error) {
	// Build search query using PostgreSQL full-text search (include content fields)
	query := `
		SELECT id, title, summary, url, published, source, keywords, image_url,
		       author, category, content_hash, created_at, updated_at,
		       COALESCE(content, '') as content,
		       COALESCE(content_extracted, FALSE) as content_extracted,
		       content_extracted_at
		FROM articles
		WHERE (
			to_tsvector('english', title || ' ' || COALESCE(summary, ''))
			@@ plainto_tsquery('english', $1)
			OR title ILIKE $2
			OR summary ILIKE $2
		)
	`
	countQuery := `
		SELECT COUNT(*)
		FROM articles
		WHERE (
			to_tsvector('english', title || ' ' || COALESCE(summary, ''))
			@@ plainto_tsquery('english', $1)
			OR title ILIKE $2
			OR summary ILIKE $2
		)
	`

	searchPattern := "%" + filter.Search + "%"
	args := []interface{}{filter.Search, searchPattern}
	argPos := 3

	// Apply additional filters
	if filter.Source != "" {
		query += fmt.Sprintf(" AND source = $%d", argPos)
		countQuery += fmt.Sprintf(" AND source = $%d", argPos)
		args = append(args, filter.Source)
		argPos++
	}

	if filter.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argPos)
		countQuery += fmt.Sprintf(" AND category = $%d", argPos)
		args = append(args, filter.Category)
		argPos++
	}

	// Get total count
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Apply ordering
	orderBy := "published"
	if filter.SortBy != "" {
		orderBy = filter.SortBy
	}
	orderDir := "DESC"
	if filter.SortOrder == "asc" {
		orderDir = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", orderBy, orderDir)

	// Apply pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, filter.Limit)
		argPos++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, filter.Offset)
	}

	// Execute query
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search articles: %w", err)
	}
	defer rows.Close()

	// Parse results
	articles := []models.Article{}
	for rows.Next() {
		var article models.Article
		var content string
		var contentExtracted bool
		var contentExtractedAt *time.Time

		err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Summary,
			&article.URL,
			&article.Published,
			&article.Source,
			&article.Keywords,
			&article.ImageURL,
			&article.Author,
			&article.Category,
			&article.ContentHash,
			&article.CreatedAt,
			&article.UpdatedAt,
			&content,
			&contentExtracted,
			&contentExtractedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan article: %w", err)
		}

		// Set content fields
		article.Content = content
		article.ContentExtracted = contentExtracted
		article.ContentExtractedAt = contentExtractedAt

		articles = append(articles, article)
	}

	return articles, total, nil
}

// ExistsByURL checks if an article with the given URL already exists
func (r *ArticleRepository) ExistsByURL(ctx context.Context, url string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM articles WHERE url = $1)"
	var exists bool
	err := r.db.QueryRow(ctx, query, url).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check article existence: %w", err)
	}
	return exists, nil
}

// ExistsByURLBatch checks which URLs from the given list already exist (batch operation)
// Returns a map of URL -> exists for efficient lookup
func (r *ArticleRepository) ExistsByURLBatch(ctx context.Context, urls []string) (map[string]bool, error) {
	if len(urls) == 0 {
		return make(map[string]bool), nil
	}

	query := `SELECT url FROM articles WHERE url = ANY($1)`
	rows, err := r.db.Query(ctx, query, urls)
	if err != nil {
		return nil, fmt.Errorf("failed to check batch URL existence: %w", err)
	}
	defer rows.Close()

	// Create result map with all URLs set to false initially
	existsMap := make(map[string]bool, len(urls))
	for _, url := range urls {
		existsMap[url] = false
	}

	// Mark existing URLs as true
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			continue
		}
		existsMap[url] = true
	}

	return existsMap, nil
}

// ExistsByContentHash checks if an article with the given content hash already exists
func (r *ArticleRepository) ExistsByContentHash(ctx context.Context, contentHash string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM articles WHERE content_hash = $1)"
	var exists bool
	err := r.db.QueryRow(ctx, query, contentHash).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check article existence by hash: %w", err)
	}
	return exists, nil
}

// DeleteOlderThan deletes articles older than the specified date
func (r *ArticleRepository) DeleteOlderThan(ctx context.Context, date time.Time) (int64, error) {
	query := "DELETE FROM articles WHERE published < $1"
	result, err := r.db.Exec(ctx, query, date)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old articles: %w", err)
	}
	return result.RowsAffected(), nil
}

// GetStatsBySource retrieves article statistics by source
func (r *ArticleRepository) GetStatsBySource(ctx context.Context) (map[string]int, error) {
	query := "SELECT source, COUNT(*) as count FROM articles GROUP BY source ORDER BY count DESC"
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var source string
		var count int
		if err := rows.Scan(&source, &count); err != nil {
			return nil, fmt.Errorf("failed to scan stats: %w", err)
		}
		stats[source] = count
	}

	return stats, nil
}

// GetComprehensiveStats retrieves comprehensive statistics about articles
func (r *ArticleRepository) GetComprehensiveStats(ctx context.Context) (*models.StatsResponse, error) {
	stats := &models.StatsResponse{
		ArticlesBySource: make(map[string]int),
		Categories:       make(map[string]models.CategoryInfo),
	}

	// Get total articles
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM articles").Scan(&stats.TotalArticles)
	if err != nil {
		return nil, fmt.Errorf("failed to get total articles: %w", err)
	}

	// Get articles by source
	sourceQuery := "SELECT source, COUNT(*) as count FROM articles GROUP BY source ORDER BY count DESC"
	rows, err := r.db.Query(ctx, sourceQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get source stats: %w", err)
	}
	for rows.Next() {
		var source string
		var count int
		if err := rows.Scan(&source, &count); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan source stats: %w", err)
		}
		stats.ArticlesBySource[source] = count
	}
	rows.Close()

	// Get recent articles count (last 24 hours)
	recentQuery := "SELECT COUNT(*) FROM articles WHERE published >= NOW() - INTERVAL '24 hours'"
	err = r.db.QueryRow(ctx, recentQuery).Scan(&stats.RecentArticles)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent articles: %w", err)
	}

	// Get oldest and newest article dates
	dateQuery := "SELECT MIN(published), MAX(published) FROM articles WHERE published IS NOT NULL"
	var oldest, newest *time.Time
	err = r.db.QueryRow(ctx, dateQuery).Scan(&oldest, &newest)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get article date range: %w", err)
	}
	stats.OldestArticle = oldest
	stats.NewestArticle = newest

	// Get category stats
	categoryQuery := `
		SELECT category, COUNT(*) as count
		FROM articles
		WHERE category IS NOT NULL AND category != ''
		GROUP BY category
		ORDER BY count DESC
	`
	rows, err = r.db.Query(ctx, categoryQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get category stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			return nil, fmt.Errorf("failed to scan category stats: %w", err)
		}
		stats.Categories[category] = models.CategoryInfo{
			Name:         category,
			ArticleCount: count,
		}
	}

	return stats, nil
}

// GetCategories retrieves all distinct categories with article counts
func (r *ArticleRepository) GetCategories(ctx context.Context) ([]models.CategoryInfo, error) {
	query := `
		SELECT category, COUNT(*) as count
		FROM articles
		WHERE category IS NOT NULL AND category != ''
		GROUP BY category
		ORDER BY count DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	categories := []models.CategoryInfo{}
	for rows.Next() {
		var category models.CategoryInfo
		if err := rows.Scan(&category.Name, &category.ArticleCount); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// generateContentHash creates a SHA256 hash of the article content for duplicate detection
func generateContentHash(title, url string) string {
	h := sha256.New()
	h.Write([]byte(title + url))
	return hex.EncodeToString(h.Sum(nil))
}

// sanitizeUTF8 removes invalid UTF-8 byte sequences from strings
// This prevents PostgreSQL "invalid byte sequence for encoding UTF8" errors
func sanitizeUTF8(s string) string {
	// strings.ToValidUTF8 replaces invalid UTF-8 sequences with the replacement character
	// We use empty string as replacement to simply remove invalid bytes
	return strings.ToValidUTF8(s, "")
}

// UpdateContent updates the full content of an article after HTML extraction
func (r *ArticleRepository) UpdateContent(ctx context.Context, id int64, content string) error {
	// Sanitize content: remove invalid UTF-8 sequences to prevent database errors
	content = sanitizeUTF8(content)

	query := `
		UPDATE articles
		SET content = $2,
		    content_extracted = TRUE,
		    content_extracted_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id, content)
	if err != nil {
		return fmt.Errorf("failed to update content: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("article not found")
	}

	return nil
}

// GetArticlesNeedingContent returns articles that need content extraction
func (r *ArticleRepository) GetArticlesNeedingContent(ctx context.Context, limit int) ([]int64, error) {
	query := `
		SELECT id
		FROM articles
		WHERE (content_extracted = FALSE OR content_extracted IS NULL)
		  AND url IS NOT NULL
		  AND url != ''
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles needing content: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// GetContentExtractionStats returns statistics about content extraction
func (r *ArticleRepository) GetContentExtractionStats(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE content_extracted = TRUE) as extracted,
			COUNT(*) FILTER (WHERE content_extracted = FALSE OR content_extracted IS NULL) as pending
		FROM articles
	`

	var total, extracted, pending int
	err := r.db.QueryRow(ctx, query).Scan(&total, &extracted, &pending)
	if err != nil {
		return nil, fmt.Errorf("failed to get content extraction stats: %w", err)
	}

	return map[string]int{
		"total":     total,
		"extracted": extracted,
		"pending":   pending,
	}, nil
}

// GetArticleWithContent retrieves an article including its full content
func (r *ArticleRepository) GetArticleWithContent(ctx context.Context, id int64) (*models.Article, error) {
	query := `
		SELECT id, title, summary, url, published, source, keywords, image_url,
		       author, category, content_hash, created_at, updated_at,
		       content, content_extracted, content_extracted_at
		FROM articles
		WHERE id = $1
	`

	var article models.Article
	var content *string
	var contentExtracted *bool
	var contentExtractedAt *time.Time

	err := r.db.QueryRow(ctx, query, id).Scan(
		&article.ID,
		&article.Title,
		&article.Summary,
		&article.URL,
		&article.Published,
		&article.Source,
		&article.Keywords,
		&article.ImageURL,
		&article.Author,
		&article.Category,
		&article.ContentHash,
		&article.CreatedAt,
		&article.UpdatedAt,
		&content,
		&contentExtracted,
		&contentExtractedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("article not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get article with content: %w", err)
	}

	// Set optional fields
	if content != nil {
		article.Content = *content
	}
	if contentExtracted != nil {
		article.ContentExtracted = *contentExtracted
	}
	if contentExtractedAt != nil {
		article.ContentExtractedAt = contentExtractedAt
	}

	return &article, nil
}
