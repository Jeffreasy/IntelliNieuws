package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jeffrey/intellinieuws/internal/models"
	"github.com/jeffrey/intellinieuws/pkg/logger"
)

// AnalyticsHandler handles analytics-related requests
type AnalyticsHandler struct {
	db     *pgxpool.Pool
	logger *logger.Logger
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(db *pgxpool.Pool, log *logger.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		db:     db,
		logger: log.WithComponent("analytics-handler"),
	}
}

// TrendingKeyword represents a trending keyword with stats
type TrendingKeyword struct {
	Keyword       string   `json:"keyword"`
	ArticleCount  int64    `json:"article_count"`
	SourceCount   int64    `json:"source_count"`
	Sources       []string `json:"sources"`
	AvgSentiment  float64  `json:"avg_sentiment"`
	AvgRelevance  float64  `json:"avg_relevance"`
	MostRecent    string   `json:"most_recent"`
	TrendingScore float64  `json:"trending_score"`
}

// SentimentTrend represents sentiment data over time
type SentimentTrend struct {
	Day                string  `json:"day"`
	Source             string  `json:"source"`
	TotalArticles      int     `json:"total_articles"`
	PositiveCount      int     `json:"positive_count"`
	NeutralCount       int     `json:"neutral_count"`
	NegativeCount      int     `json:"negative_count"`
	AvgSentiment       float64 `json:"avg_sentiment"`
	PositivePercentage float64 `json:"positive_percentage"`
	NegativePercentage float64 `json:"negative_percentage"`
}

// HotEntity represents a frequently mentioned entity
type HotEntity struct {
	Entity           string   `json:"entity"`
	EntityType       string   `json:"entity_type"`
	TotalMentions    int64    `json:"total_mentions"`
	DaysMentioned    int      `json:"days_mentioned"`
	Sources          []string `json:"sources"`
	OverallSentiment float64  `json:"overall_sentiment"`
	MostRecent       string   `json:"most_recent_mention"`
}

// GetTrendingKeywords returns trending keywords from the last 24 hours
// GET /api/v1/analytics/trending?hours=24&min_articles=3&limit=20
func (h *AnalyticsHandler) GetTrendingKeywords(c *fiber.Ctx) error {
	// Parse query parameters
	hours := c.QueryInt("hours", models.DefaultTrendingHoursBack)
	minArticles := c.QueryInt("min_articles", models.DefaultTrendingMinArticles)
	limit := c.QueryInt("limit", models.DefaultTrendingLimit)

	// Validate parameters
	if hours < 1 {
		hours = models.DefaultTrendingHoursBack
	}
	if minArticles < 1 {
		minArticles = models.DefaultTrendingMinArticles
	}
	if limit < 1 || limit > 100 {
		limit = models.DefaultTrendingLimit
	}

	h.logger.Debugf("Fetching trending keywords: hours=%d, min_articles=%d, limit=%d", hours, minArticles, limit)

	// Query trending topics using database function
	query := `SELECT * FROM get_trending_topics($1, $2, $3)`
	rows, err := h.db.Query(context.Background(), query, hours, minArticles, limit)
	if err != nil {
		h.logger.Errorf("Failed to get trending keywords: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch trending keywords",
			Code:    fiber.StatusInternalServerError,
		})
	}
	defer rows.Close()

	// Parse results
	trending := make([]TrendingKeyword, 0)
	for rows.Next() {
		var kw TrendingKeyword
		var mostRecent interface{}

		err := rows.Scan(
			&kw.Keyword,
			&kw.ArticleCount,
			&kw.SourceCount,
			&kw.Sources,
			&kw.AvgSentiment,
			&kw.AvgRelevance,
			&kw.TrendingScore,
			&mostRecent,
		)
		if err != nil {
			h.logger.Warnf("Failed to scan trending keyword: %v", err)
			continue
		}

		if mostRecent != nil {
			if t, ok := mostRecent.(string); ok {
				kw.MostRecent = t
			}
		}

		trending = append(trending, kw)
	}

	h.logger.Infof("Returning %d trending keywords", len(trending))

	return c.JSON(fiber.Map{
		"trending": trending,
		"meta": fiber.Map{
			"hours":        hours,
			"min_articles": minArticles,
			"limit":        limit,
			"count":        len(trending),
		},
	})
}

// GetSentimentTrends returns sentiment trends over the last 7 days
// GET /api/v1/analytics/sentiment-trends?source=nu.nl
func (h *AnalyticsHandler) GetSentimentTrends(c *fiber.Ctx) error {
	source := c.Query("source")

	h.logger.Debugf("Fetching sentiment trends for source: %s", source)

	// Build query
	query := `SELECT * FROM v_sentiment_trends_7d`
	args := make([]interface{}, 0)

	if source != "" {
		query += ` WHERE source = $1`
		args = append(args, source)
	}

	query += ` ORDER BY day DESC, source`

	rows, err := h.db.Query(context.Background(), query, args...)
	if err != nil {
		h.logger.Errorf("Failed to get sentiment trends: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch sentiment trends",
			Code:    fiber.StatusInternalServerError,
		})
	}
	defer rows.Close()

	// Parse results
	trends := make([]SentimentTrend, 0)
	for rows.Next() {
		var trend SentimentTrend
		err := rows.Scan(
			&trend.Day,
			&trend.Source,
			&trend.TotalArticles,
			&trend.PositiveCount,
			&trend.NeutralCount,
			&trend.NegativeCount,
			&trend.AvgSentiment,
			&trend.PositivePercentage,
			&trend.NegativePercentage,
		)
		if err != nil {
			h.logger.Warnf("Failed to scan sentiment trend: %v", err)
			continue
		}

		trends = append(trends, trend)
	}

	h.logger.Infof("Returning %d sentiment trends", len(trends))

	return c.JSON(fiber.Map{
		"trends": trends,
		"meta": fiber.Map{
			"source": source,
			"count":  len(trends),
		},
	})
}

// GetHotEntities returns most mentioned entities from the last 7 days
// GET /api/v1/analytics/hot-entities?entity_type=person&limit=50
func (h *AnalyticsHandler) GetHotEntities(c *fiber.Ctx) error {
	entityType := c.Query("entity_type")
	limit := c.QueryInt("limit", 50)

	if limit < 1 || limit > 100 {
		limit = 50
	}

	h.logger.Debugf("Fetching hot entities: type=%s, limit=%d", entityType, limit)

	// Build query
	query := `SELECT * FROM v_hot_entities_7d`
	args := make([]interface{}, 0)

	if entityType != "" {
		query += ` WHERE entity_type = $1`
		args = append(args, entityType)
		query += ` LIMIT $2`
		args = append(args, limit)
	} else {
		query += ` LIMIT $1`
		args = append(args, limit)
	}

	rows, err := h.db.Query(context.Background(), query, args...)
	if err != nil {
		h.logger.Errorf("Failed to get hot entities: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch hot entities",
			Code:    fiber.StatusInternalServerError,
		})
	}
	defer rows.Close()

	// Parse results
	entities := make([]HotEntity, 0)
	for rows.Next() {
		var entity HotEntity
		var mostRecent interface{}

		err := rows.Scan(
			&entity.Entity,
			&entity.EntityType,
			&entity.TotalMentions,
			&entity.DaysMentioned,
			&entity.Sources,
			&entity.OverallSentiment,
			&mostRecent,
		)
		if err != nil {
			h.logger.Warnf("Failed to scan entity: %v", err)
			continue
		}

		if mostRecent != nil {
			if t, ok := mostRecent.(string); ok {
				entity.MostRecent = t
			}
		}

		entities = append(entities, entity)
	}

	h.logger.Infof("Returning %d hot entities", len(entities))

	return c.JSON(fiber.Map{
		"entities": entities,
		"meta": fiber.Map{
			"entity_type": entityType,
			"limit":       limit,
			"count":       len(entities),
		},
	})
}

// GetEntitySentiment returns sentiment analysis for a specific entity
// GET /api/v1/analytics/entity-sentiment?entity=Elon+Musk&days=30
func (h *AnalyticsHandler) GetEntitySentiment(c *fiber.Ctx) error {
	entity := c.Query("entity")
	if entity == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "missing_parameter",
			Message: "entity parameter is required",
			Code:    fiber.StatusBadRequest,
		})
	}

	days := c.QueryInt("days", 30)
	if days < 1 || days > 365 {
		days = 30
	}

	h.logger.Debugf("Fetching entity sentiment: entity=%s, days=%d", entity, days)

	// Query entity sentiment analysis
	query := `SELECT * FROM get_entity_sentiment_analysis($1, $2)`
	rows, err := h.db.Query(context.Background(), query, entity, days)
	if err != nil {
		h.logger.Errorf("Failed to get entity sentiment: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch entity sentiment analysis",
			Code:    fiber.StatusInternalServerError,
		})
	}
	defer rows.Close()

	// Parse results
	type EntitySentimentDay struct {
		Day          string   `json:"day"`
		MentionCount int64    `json:"mention_count"`
		AvgSentiment float64  `json:"avg_sentiment"`
		Sources      []string `json:"sources"`
		Categories   []string `json:"categories"`
	}

	timeline := make([]EntitySentimentDay, 0)
	for rows.Next() {
		var day EntitySentimentDay
		err := rows.Scan(
			&day.Day,
			&day.MentionCount,
			&day.AvgSentiment,
			&day.Sources,
			&day.Categories,
		)
		if err != nil {
			h.logger.Warnf("Failed to scan entity sentiment: %v", err)
			continue
		}

		timeline = append(timeline, day)
	}

	h.logger.Infof("Returning entity sentiment for '%s': %d days", entity, len(timeline))

	return c.JSON(fiber.Map{
		"entity":   entity,
		"timeline": timeline,
		"meta": fiber.Map{
			"days":  days,
			"count": len(timeline),
		},
	})
}

// RefreshAnalytics refreshes all materialized views
// POST /api/v1/analytics/refresh
func (h *AnalyticsHandler) RefreshAnalytics(c *fiber.Ctx) error {
	concurrent := c.Query("concurrent", "true") != "false"

	h.logger.Infof("Refreshing analytics views (concurrent=%v)", concurrent)

	// Call database refresh function
	query := `SELECT * FROM refresh_analytics_views($1)`
	rows, err := h.db.Query(context.Background(), query, concurrent)
	if err != nil {
		h.logger.Errorf("Failed to refresh analytics: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "refresh_failed",
			Message: err.Error(),
			Code:    fiber.StatusInternalServerError,
		})
	}
	defer rows.Close()

	// Collect refresh results
	type RefreshResult struct {
		ViewName      string `json:"view_name"`
		RefreshTimeMs int    `json:"refresh_time_ms"`
		RowsAffected  int64  `json:"rows_affected"`
	}

	results := make([]RefreshResult, 0)
	totalTime := 0
	totalRows := int64(0)

	for rows.Next() {
		var result RefreshResult
		err := rows.Scan(&result.ViewName, &result.RefreshTimeMs, &result.RowsAffected)
		if err != nil {
			h.logger.Warnf("Failed to scan refresh result: %v", err)
			continue
		}

		totalTime += result.RefreshTimeMs
		totalRows += result.RowsAffected
		results = append(results, result)
	}

	h.logger.Infof("Analytics refresh completed: %d views, %d total rows, %dms",
		len(results), totalRows, totalTime)

	return c.JSON(fiber.Map{
		"message": "Analytics refreshed successfully",
		"results": results,
		"summary": fiber.Map{
			"total_views":     len(results),
			"total_rows":      totalRows,
			"total_time_ms":   totalTime,
			"concurrent_mode": concurrent,
		},
	})
}

// GetAnalyticsOverview returns a comprehensive analytics overview
// GET /api/v1/analytics/overview
func (h *AnalyticsHandler) GetAnalyticsOverview(c *fiber.Ctx) error {
	h.logger.Debug("Fetching analytics overview")

	ctx := context.Background()

	// Get trending keywords (top 10)
	trendingQuery := `SELECT * FROM v_trending_keywords_24h LIMIT 10`
	trendingRows, err := h.db.Query(ctx, trendingQuery)
	if err != nil {
		h.logger.Errorf("Failed to get trending keywords: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch analytics overview",
			Code:    fiber.StatusInternalServerError,
		})
	}

	trending := make([]TrendingKeyword, 0)
	for trendingRows.Next() {
		var kw TrendingKeyword
		var mostRecent *string

		err := trendingRows.Scan(
			&kw.Keyword,
			&kw.ArticleCount,
			&kw.Sources,
			&kw.AvgSentiment,
			&mostRecent,
			&kw.TrendingScore,
		)
		if err == nil {
			if mostRecent != nil {
				kw.MostRecent = *mostRecent
			}
			trending = append(trending, kw)
		}
	}
	trendingRows.Close()

	// Get hot entities (top 10)
	entitiesQuery := `SELECT * FROM v_hot_entities_7d LIMIT 10`
	entityRows, err := h.db.Query(ctx, entitiesQuery)
	if err != nil {
		h.logger.Warnf("Failed to get hot entities: %v", err)
	}

	entities := make([]HotEntity, 0)
	if entityRows != nil {
		for entityRows.Next() {
			var entity HotEntity
			var mostRecent *string

			err := entityRows.Scan(
				&entity.Entity,
				&entity.EntityType,
				&entity.TotalMentions,
				&entity.DaysMentioned,
				&entity.Sources,
				&entity.OverallSentiment,
				&mostRecent,
			)
			if err == nil {
				if mostRecent != nil {
					entity.MostRecent = *mostRecent
				}
				entities = append(entities, entity)
			}
		}
		entityRows.Close()
	}

	// Get materialized view status
	mvQuery := `
		SELECT matviewname, pg_size_pretty(pg_total_relation_size('public.'||matviewname)) as size
		FROM pg_matviews 
		WHERE schemaname = 'public'
	`
	mvRows, err := h.db.Query(ctx, mvQuery)

	mvStatus := make([]fiber.Map, 0)
	if err == nil && mvRows != nil {
		for mvRows.Next() {
			var name, size string
			if err := mvRows.Scan(&name, &size); err == nil {
				mvStatus = append(mvStatus, fiber.Map{
					"name": name,
					"size": size,
				})
			}
		}
		mvRows.Close()
	}

	h.logger.Info("Analytics overview fetched successfully")

	return c.JSON(fiber.Map{
		"trending_keywords":  trending,
		"hot_entities":       entities,
		"materialized_views": mvStatus,
		"meta": fiber.Map{
			"trending_count": len(trending),
			"entities_count": len(entities),
			"views_count":    len(mvStatus),
		},
	})
}

// GetArticleStats returns comprehensive article statistics
// GET /api/v1/analytics/article-stats
func (h *AnalyticsHandler) GetArticleStats(c *fiber.Ctx) error {
	h.logger.Debug("Fetching article statistics")

	query := `SELECT * FROM v_article_stats ORDER BY total_articles DESC`
	rows, err := h.db.Query(context.Background(), query)
	if err != nil {
		h.logger.Errorf("Failed to get article stats: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch article statistics",
			Code:    fiber.StatusInternalServerError,
		})
	}
	defer rows.Close()

	type ArticleSourceStats struct {
		Source                string   `json:"source"`
		SourceName            string   `json:"source_name"`
		TotalArticles         int      `json:"total_articles"`
		ArticlesToday         int      `json:"articles_today"`
		ArticlesWeek          int      `json:"articles_week"`
		AIProcessedCount      int      `json:"ai_processed_count"`
		ContentExtractedCount int      `json:"content_extracted_count"`
		LatestArticleDate     string   `json:"latest_article_date"`
		OldestArticleDate     string   `json:"oldest_article_date"`
		AvgSentiment          *float64 `json:"avg_sentiment"`
	}

	stats := make([]ArticleSourceStats, 0)
	for rows.Next() {
		var stat ArticleSourceStats
		var latestDate, oldestDate *string

		err := rows.Scan(
			&stat.Source,
			&stat.SourceName,
			&stat.TotalArticles,
			&stat.ArticlesToday,
			&stat.ArticlesWeek,
			&stat.AIProcessedCount,
			&stat.ContentExtractedCount,
			&latestDate,
			&oldestDate,
			&stat.AvgSentiment,
		)
		if err != nil {
			h.logger.Warnf("Failed to scan article stat: %v", err)
			continue
		}

		if latestDate != nil {
			stat.LatestArticleDate = *latestDate
		}
		if oldestDate != nil {
			stat.OldestArticleDate = *oldestDate
		}

		stats = append(stats, stat)
	}

	h.logger.Infof("Returning stats for %d sources", len(stats))

	return c.JSON(fiber.Map{
		"sources": stats,
		"meta": fiber.Map{
			"count": len(stats),
		},
	})
}

// GetMaintenanceSchedule returns recommended maintenance schedule
// GET /api/v1/analytics/maintenance-schedule
func (h *AnalyticsHandler) GetMaintenanceSchedule(c *fiber.Ctx) error {
	h.logger.Debug("Fetching maintenance schedule")

	query := `SELECT * FROM get_maintenance_schedule()`
	rows, err := h.db.Query(context.Background(), query)
	if err != nil {
		h.logger.Errorf("Failed to get maintenance schedule: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch maintenance schedule",
			Code:    fiber.StatusInternalServerError,
		})
	}
	defer rows.Close()

	type MaintenanceTask struct {
		Task            string  `json:"task"`
		Frequency       string  `json:"frequency"`
		LastRun         *string `json:"last_run"`
		NextRecommended *string `json:"next_recommended"`
		Status          string  `json:"status"`
	}

	tasks := make([]MaintenanceTask, 0)
	for rows.Next() {
		var task MaintenanceTask
		err := rows.Scan(
			&task.Task,
			&task.Frequency,
			&task.LastRun,
			&task.NextRecommended,
			&task.Status,
		)
		if err != nil {
			h.logger.Warnf("Failed to scan maintenance task: %v", err)
			continue
		}

		tasks = append(tasks, task)
	}

	h.logger.Infof("Returning %d maintenance tasks", len(tasks))

	return c.JSON(fiber.Map{
		"tasks": tasks,
		"meta": fiber.Map{
			"count": len(tasks),
		},
	})
}

// GetDatabaseHealth returns database health metrics
// GET /api/v1/analytics/database-health
func (h *AnalyticsHandler) GetDatabaseHealth(c *fiber.Ctx) error {
	h.logger.Debug("Fetching database health metrics")

	ctx := context.Background()

	// Get table sizes
	sizeQuery := `
		SELECT 
			tablename,
			pg_size_pretty(pg_total_relation_size('public.'||tablename)) as size,
			pg_total_relation_size('public.'||tablename) as bytes
		FROM pg_tables 
		WHERE schemaname = 'public'
		ORDER BY bytes DESC
	`
	sizeRows, err := h.db.Query(ctx, sizeQuery)

	type TableSize struct {
		Table string `json:"table"`
		Size  string `json:"size"`
		Bytes int64  `json:"bytes"`
	}

	tableSizes := make([]TableSize, 0)
	if err == nil {
		for sizeRows.Next() {
			var ts TableSize
			if err := sizeRows.Scan(&ts.Table, &ts.Size, &ts.Bytes); err == nil {
				tableSizes = append(tableSizes, ts)
			}
		}
		sizeRows.Close()
	}

	// Get cache hit ratio
	cacheQuery := `
		SELECT 
			ROUND(100.0 * SUM(heap_blks_hit) / NULLIF(SUM(heap_blks_hit) + SUM(heap_blks_read), 0), 2) as hit_ratio
		FROM pg_statio_user_tables
	`
	var cacheHitRatio *float64
	h.db.QueryRow(ctx, cacheQuery).Scan(&cacheHitRatio)

	// Get connection count
	connQuery := `
		SELECT COUNT(*) 
		FROM pg_stat_activity 
		WHERE datname = current_database()
	`
	var connectionCount int
	h.db.QueryRow(ctx, connQuery).Scan(&connectionCount)

	h.logger.Info("Database health metrics fetched")

	return c.JSON(fiber.Map{
		"table_sizes":      tableSizes,
		"cache_hit_ratio":  cacheHitRatio,
		"connection_count": connectionCount,
		"status":           "healthy",
	})
}
