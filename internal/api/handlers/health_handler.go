package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jeffrey/intellinieuws/internal/ai"
	"github.com/jeffrey/intellinieuws/internal/cache"
	"github.com/jeffrey/intellinieuws/internal/models"
	"github.com/jeffrey/intellinieuws/internal/scraper"
	"github.com/jeffrey/intellinieuws/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db             *pgxpool.Pool
	redis          *redis.Client
	cacheService   *cache.Service
	scraperService *scraper.Service
	aiProcessor    *ai.Processor
	logger         *logger.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(
	db *pgxpool.Pool,
	redis *redis.Client,
	cacheService *cache.Service,
	scraperService *scraper.Service,
	aiProcessor *ai.Processor,
	log *logger.Logger,
) *HealthHandler {
	return &HealthHandler{
		db:             db,
		redis:          redis,
		cacheService:   cacheService,
		scraperService: scraperService,
		aiProcessor:    aiProcessor,
		logger:         log.WithComponent("health-handler"),
	}
}

// HealthStatus represents the overall health status
type HealthStatus struct {
	Status     string                     `json:"status"` // healthy, degraded, unhealthy
	Timestamp  time.Time                  `json:"timestamp"`
	Version    string                     `json:"version"`
	Uptime     float64                    `json:"uptime_seconds"`
	Components map[string]ComponentHealth `json:"components"`
	Metrics    map[string]interface{}     `json:"metrics,omitempty"`
}

// ComponentHealth represents health of a single component
type ComponentHealth struct {
	Status  string                 `json:"status"` // healthy, degraded, unhealthy
	Message string                 `json:"message,omitempty"`
	Latency float64                `json:"latency_ms,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

var startTime = time.Now()

// GetHealth returns comprehensive health status
// GET /health
func (h *HealthHandler) GetHealth(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)
	health := &HealthStatus{
		Status:     "healthy",
		Timestamp:  time.Now(),
		Version:    "1.0.0",
		Uptime:     time.Since(startTime).Seconds(),
		Components: make(map[string]ComponentHealth),
		Metrics:    make(map[string]interface{}),
	}

	// Check database health
	dbHealth := h.checkDatabase(c.Context())
	health.Components["database"] = dbHealth
	if dbHealth.Status != "healthy" {
		health.Status = "degraded"
	}

	// Check Redis health
	redisHealth := h.checkRedis(c.Context())
	health.Components["redis"] = redisHealth
	if redisHealth.Status == "unhealthy" && h.cacheService != nil {
		health.Status = "degraded"
	}

	// Check scraper health
	scraperHealth := h.checkScraper(c.Context())
	health.Components["scraper"] = scraperHealth
	if scraperHealth.Status != "healthy" {
		health.Status = "degraded"
	}

	// Check AI processor health
	if h.aiProcessor != nil {
		aiHealth := h.checkAIProcessor()
		health.Components["ai_processor"] = aiHealth
		if aiHealth.Status != "healthy" {
			health.Status = "degraded"
		}
	}

	// Add system metrics
	h.addSystemMetrics(health)

	// Determine HTTP status code
	statusCode := fiber.StatusOK
	if health.Status == "unhealthy" {
		statusCode = fiber.StatusServiceUnavailable
	} else if health.Status == "degraded" {
		statusCode = fiber.StatusOK // Still operational, just degraded
	}

	return c.Status(statusCode).JSON(models.NewSuccessResponse(health, requestID))
}

// GetLiveness returns simple liveness check (Kubernetes-style)
// GET /health/live
func (h *HealthHandler) GetLiveness(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "alive",
		"time":   time.Now(),
	})
}

// GetReadiness returns readiness check (Kubernetes-style)
// GET /health/ready
func (h *HealthHandler) GetReadiness(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
	defer cancel()

	// Check critical dependencies
	ready := true
	components := make(map[string]bool)

	// Check database
	if err := h.db.Ping(ctx); err != nil {
		ready = false
		components["database"] = false
	} else {
		components["database"] = true
	}

	// Check Redis (optional)
	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			components["redis"] = false
			// Redis is optional, so don't mark as not ready
		} else {
			components["redis"] = true
		}
	}

	status := "ready"
	statusCode := fiber.StatusOK
	if !ready {
		status = "not_ready"
		statusCode = fiber.StatusServiceUnavailable
	}

	return c.Status(statusCode).JSON(fiber.Map{
		"status":     status,
		"components": components,
		"time":       time.Now(),
	})
}

// checkDatabase checks database health
func (h *HealthHandler) checkDatabase(ctx context.Context) ComponentHealth {
	start := time.Now()

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := h.db.Ping(pingCtx); err != nil {
		return ComponentHealth{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Ping failed: %v", err),
			Latency: float64(time.Since(start).Milliseconds()),
		}
	}

	// Get pool stats
	stats := h.db.Stat()
	details := map[string]interface{}{
		"total_conns":      stats.TotalConns(),
		"idle_conns":       stats.IdleConns(),
		"acquired_conns":   stats.AcquiredConns(),
		"max_conns":        stats.MaxConns(),
		"acquire_count":    stats.AcquireCount(),
		"acquire_duration": stats.AcquireDuration().Milliseconds(),
	}

	return ComponentHealth{
		Status:  "healthy",
		Message: "Database connection healthy",
		Latency: float64(time.Since(start).Milliseconds()),
		Details: details,
	}
}

// checkRedis checks Redis health
func (h *HealthHandler) checkRedis(ctx context.Context) ComponentHealth {
	start := time.Now()

	if h.redis == nil {
		return ComponentHealth{
			Status:  "disabled",
			Message: "Redis not configured",
		}
	}

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := h.redis.Ping(pingCtx).Err(); err != nil {
		return ComponentHealth{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Ping failed: %v", err),
			Latency: float64(time.Since(start).Milliseconds()),
		}
	}

	return ComponentHealth{
		Status:  "healthy",
		Message: "Redis connection healthy",
		Latency: float64(time.Since(start).Milliseconds()),
		Details: map[string]interface{}{
			"cache_available": h.cacheService != nil && h.cacheService.IsAvailable(),
		},
	}
}

// checkScraper checks scraper service health
func (h *HealthHandler) checkScraper(ctx context.Context) ComponentHealth {
	if h.scraperService == nil {
		return ComponentHealth{
			Status:  "disabled",
			Message: "Scraper service not configured",
		}
	}

	health := h.scraperService.GetHealth(ctx)

	status := "healthy"
	if healthStatus, ok := health["status"].(string); ok && healthStatus != "healthy" {
		status = healthStatus
	}

	return ComponentHealth{
		Status:  status,
		Message: "Scraper service operational",
		Details: health,
	}
}

// checkAIProcessor checks AI processor health
func (h *HealthHandler) checkAIProcessor() ComponentHealth {
	if h.aiProcessor == nil {
		return ComponentHealth{
			Status:  "disabled",
			Message: "AI processor not configured",
		}
	}

	stats := h.aiProcessor.GetStats()

	status := "healthy"
	message := "AI processor operational"

	if !stats.IsRunning {
		status = "degraded"
		message = "AI processor not running"
	} else if time.Since(stats.LastRun) > 30*time.Minute {
		status = "degraded"
		message = "AI processor hasn't run recently"
	}

	return ComponentHealth{
		Status:  status,
		Message: message,
		Details: map[string]interface{}{
			"is_running":       stats.IsRunning,
			"process_count":    stats.ProcessCount,
			"last_run":         stats.LastRun,
			"current_interval": stats.CurrentInterval.String(),
		},
	}
}

// addSystemMetrics adds system-level metrics to health status
func (h *HealthHandler) addSystemMetrics(health *HealthStatus) {
	health.Metrics["uptime_seconds"] = time.Since(startTime).Seconds()
	health.Metrics["timestamp"] = time.Now().Unix()

	// Add database pool metrics
	if dbComp, ok := health.Components["database"]; ok {
		if dbComp.Details != nil {
			health.Metrics["db_total_conns"] = dbComp.Details["total_conns"]
			health.Metrics["db_idle_conns"] = dbComp.Details["idle_conns"]
			health.Metrics["db_acquired_conns"] = dbComp.Details["acquired_conns"]
		}
	}

	// Add AI processor metrics
	if aiComp, ok := health.Components["ai_processor"]; ok {
		if aiComp.Details != nil {
			health.Metrics["ai_process_count"] = aiComp.Details["process_count"]
			health.Metrics["ai_is_running"] = aiComp.Details["is_running"]
		}
	}
}

// GetMetrics returns detailed metrics (Prometheus-compatible format)
// GET /health/metrics
func (h *HealthHandler) GetMetrics(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	metrics := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"uptime":    time.Since(startTime).Seconds(),
	}

	// Database metrics
	if h.db != nil {
		stats := h.db.Stat()
		metrics["db_total_conns"] = stats.TotalConns()
		metrics["db_idle_conns"] = stats.IdleConns()
		metrics["db_acquired_conns"] = stats.AcquiredConns()
		metrics["db_max_conns"] = stats.MaxConns()
		metrics["db_acquire_count"] = stats.AcquireCount()
		metrics["db_acquire_duration_ms"] = stats.AcquireDuration().Milliseconds()
	}

	// AI processor metrics
	if h.aiProcessor != nil {
		stats := h.aiProcessor.GetStats()
		metrics["ai_is_running"] = stats.IsRunning
		metrics["ai_process_count"] = stats.ProcessCount
		metrics["ai_last_run"] = stats.LastRun.Unix()
		metrics["ai_current_interval_seconds"] = stats.CurrentInterval.Seconds()
	}

	// Scraper metrics
	if h.scraperService != nil {
		scraperStats, err := h.scraperService.GetStats(context.Background())
		if err == nil {
			metrics["scraper"] = scraperStats
		}
	}

	return c.JSON(models.NewSuccessResponse(metrics, requestID))
}
