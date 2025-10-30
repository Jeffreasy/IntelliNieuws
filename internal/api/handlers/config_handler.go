package handlers

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jeffrey/intellinieuws/internal/models"
	"github.com/jeffrey/intellinieuws/pkg/config"
	"github.com/jeffrey/intellinieuws/pkg/logger"
)

// ConfigHandler handles configuration and runtime settings
type ConfigHandler struct {
	config    *config.Config
	scheduler interface {
		UpdateInterval(interval time.Duration)
		IsRunning() bool
	}
	logger         *logger.Logger
	mu             sync.RWMutex
	activeProfile  string
	scraperConfigs map[string]*config.ScraperConfig
}

// NewConfigHandler creates a new configuration handler
func NewConfigHandler(cfg *config.Config, log *logger.Logger) *ConfigHandler {
	handler := &ConfigHandler{
		config:         cfg,
		logger:         log.WithComponent("config-handler"),
		activeProfile:  "balanced",
		scraperConfigs: make(map[string]*config.ScraperConfig),
	}

	// Initialize predefined profiles
	handler.initializeProfiles()

	return handler
}

// SetScheduler sets the scheduler reference for runtime updates
func (h *ConfigHandler) SetScheduler(scheduler interface {
	UpdateInterval(interval time.Duration)
	IsRunning() bool
}) {
	h.scheduler = scheduler
}

// initializeProfiles creates the predefined scraper profiles
func (h *ConfigHandler) initializeProfiles() {
	// Fast profile - Maximum throughput
	h.scraperConfigs["fast"] = &config.ScraperConfig{
		UserAgent:                   h.config.Scraper.UserAgent,
		RateLimitSeconds:            2,
		MaxConcurrent:               10,
		TimeoutSeconds:              15,
		RetryAttempts:               3,
		TargetSites:                 []string{"nu.nl", "ad.nl", "nos.nl"},
		EnableRSSPriority:           true,
		EnableDynamicScraping:       false,
		EnableRobotsTxtCheck:        false,
		EnableDuplicateDetection:    true,
		ScheduleEnabled:             true,
		ScheduleIntervalMinutes:     5,
		EnableFullContentExtraction: false,
		EnableBrowserScraping:       true,
		BrowserPoolSize:             10,
		BrowserTimeout:              15 * time.Second,
		BrowserWaitAfterLoad:        1000 * time.Millisecond,
		BrowserFallbackOnly:         true,
		BrowserMaxConcurrent:        5,
	}

	// Balanced profile - Default (current settings)
	h.scraperConfigs["balanced"] = &h.config.Scraper

	// Deep profile - Maximum quality
	h.scraperConfigs["deep"] = &config.ScraperConfig{
		UserAgent:                   h.config.Scraper.UserAgent,
		RateLimitSeconds:            5,
		MaxConcurrent:               3,
		TimeoutSeconds:              30,
		RetryAttempts:               3,
		TargetSites:                 []string{"nu.nl", "ad.nl", "nos.nl", "trouw.nl"},
		EnableRSSPriority:           true,
		EnableDynamicScraping:       true,
		EnableRobotsTxtCheck:        true,
		EnableDuplicateDetection:    true,
		ScheduleEnabled:             true,
		ScheduleIntervalMinutes:     60,
		EnableFullContentExtraction: true,
		ContentExtractionInterval:   10 * time.Minute,
		ContentExtractionBatchSize:  20,
		ContentExtractionAsync:      true,
		EnableBrowserScraping:       true,
		BrowserPoolSize:             7,
		BrowserTimeout:              30 * time.Second,
		BrowserWaitAfterLoad:        3000 * time.Millisecond,
		BrowserFallbackOnly:         false,
		BrowserMaxConcurrent:        4,
	}

	// Conservative profile - Minimal load
	h.scraperConfigs["conservative"] = &config.ScraperConfig{
		UserAgent:                   h.config.Scraper.UserAgent,
		RateLimitSeconds:            10,
		MaxConcurrent:               2,
		TimeoutSeconds:              60,
		RetryAttempts:               3,
		TargetSites:                 []string{"nu.nl", "ad.nl", "nos.nl"},
		EnableRSSPriority:           true,
		EnableDynamicScraping:       false,
		EnableRobotsTxtCheck:        true,
		EnableDuplicateDetection:    true,
		ScheduleEnabled:             true,
		ScheduleIntervalMinutes:     30,
		EnableFullContentExtraction: false,
		EnableBrowserScraping:       true,
		BrowserPoolSize:             2,
		BrowserTimeout:              15 * time.Second,
		BrowserWaitAfterLoad:        2000 * time.Millisecond,
		BrowserFallbackOnly:         true,
		BrowserMaxConcurrent:        1,
	}
}

// GetProfiles returns all available scraper profiles
// GET /api/v1/config/profiles
func (h *ConfigHandler) GetProfiles(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	h.mu.RLock()
	defer h.mu.RUnlock()

	profiles := make(map[string]interface{})
	for name, cfg := range h.scraperConfigs {
		profiles[name] = fiber.Map{
			"name":                    name,
			"rate_limit_seconds":      cfg.RateLimitSeconds,
			"max_concurrent":          cfg.MaxConcurrent,
			"timeout_seconds":         cfg.TimeoutSeconds,
			"schedule_interval_min":   cfg.ScheduleIntervalMinutes,
			"browser_pool_size":       cfg.BrowserPoolSize,
			"browser_max_concurrent":  cfg.BrowserMaxConcurrent,
			"target_sites":            cfg.TargetSites,
			"enable_browser_scraping": cfg.EnableBrowserScraping,
			"enable_full_content":     cfg.EnableFullContentExtraction,
			"active":                  name == h.activeProfile,
		}
	}

	response := fiber.Map{
		"profiles":       profiles,
		"active_profile": h.activeProfile,
		"total_profiles": len(profiles),
	}

	return c.JSON(models.NewSuccessResponse(response, requestID))
}

// GetCurrentConfig returns the current active configuration
// GET /api/v1/config/current
func (h *ConfigHandler) GetCurrentConfig(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	h.mu.RLock()
	cfg := h.scraperConfigs[h.activeProfile]
	profile := h.activeProfile
	h.mu.RUnlock()

	response := fiber.Map{
		"active_profile":             profile,
		"rate_limit_seconds":         cfg.RateLimitSeconds,
		"max_concurrent":             cfg.MaxConcurrent,
		"timeout_seconds":            cfg.TimeoutSeconds,
		"retry_attempts":             cfg.RetryAttempts,
		"schedule_interval_min":      cfg.ScheduleIntervalMinutes,
		"target_sites":               cfg.TargetSites,
		"enable_browser_scraping":    cfg.EnableBrowserScraping,
		"browser_pool_size":          cfg.BrowserPoolSize,
		"browser_timeout_seconds":    int(cfg.BrowserTimeout.Seconds()),
		"browser_wait_after_load_ms": int(cfg.BrowserWaitAfterLoad.Milliseconds()),
		"browser_fallback_only":      cfg.BrowserFallbackOnly,
		"browser_max_concurrent":     cfg.BrowserMaxConcurrent,
		"enable_full_content":        cfg.EnableFullContentExtraction,
		"content_batch_size":         cfg.ContentExtractionBatchSize,
		"enable_robots_check":        cfg.EnableRobotsTxtCheck,
		"enable_duplicate_detection": cfg.EnableDuplicateDetection,
	}

	return c.JSON(models.NewSuccessResponse(response, requestID))
}

// SwitchProfile switches to a different scraper profile
// POST /api/v1/config/profile/:name
func (h *ConfigHandler) SwitchProfile(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)
	profileName := c.Params("name")

	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if profile exists
	if _, exists := h.scraperConfigs[profileName]; !exists {
		return c.Status(fiber.StatusNotFound).JSON(
			models.NewErrorResponse("PROFILE_NOT_FOUND",
				fmt.Sprintf("Profile '%s' not found", profileName),
				"Available profiles: fast, balanced, deep, conservative",
				requestID),
		)
	}

	// Update active profile
	oldProfile := h.activeProfile
	h.activeProfile = profileName
	newCfg := h.scraperConfigs[profileName]

	// Update scheduler interval if scheduler is available
	if h.scheduler != nil {
		interval := time.Duration(newCfg.ScheduleIntervalMinutes) * time.Minute
		h.scheduler.UpdateInterval(interval)
		h.logger.Infof("Updated scheduler interval to %v", interval)
	}

	// Update the global config
	h.config.Scraper = *newCfg

	h.logger.Infof("Switched scraper profile from '%s' to '%s'", oldProfile, profileName)

	response := fiber.Map{
		"success":        true,
		"message":        fmt.Sprintf("Switched to profile '%s'", profileName),
		"old_profile":    oldProfile,
		"new_profile":    profileName,
		"new_interval":   newCfg.ScheduleIntervalMinutes,
		"new_rate_limit": newCfg.RateLimitSeconds,
	}

	return c.JSON(models.NewSuccessResponse(response, requestID))
}

// UpdateSetting updates a specific configuration setting
// PATCH /api/v1/config/setting
func (h *ConfigHandler) UpdateSetting(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	// Parse request body
	var req struct {
		Setting string      `json:"setting"`
		Value   interface{} `json:"value"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_REQUEST",
				"Failed to parse request body",
				err.Error(),
				requestID),
		)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	cfg := h.scraperConfigs[h.activeProfile]

	// Update the specific setting
	updated := false
	var newValue interface{}

	switch req.Setting {
	case "rate_limit_seconds":
		if val, ok := req.Value.(float64); ok && val >= 1 && val <= 60 {
			cfg.RateLimitSeconds = int(val)
			newValue = int(val)
			updated = true
		}
	case "max_concurrent":
		if val, ok := req.Value.(float64); ok && val >= 1 && val <= 20 {
			cfg.MaxConcurrent = int(val)
			newValue = int(val)
			updated = true
		}
	case "timeout_seconds":
		if val, ok := req.Value.(float64); ok && val >= 10 && val <= 120 {
			cfg.TimeoutSeconds = int(val)
			newValue = int(val)
			updated = true
		}
	case "schedule_interval_minutes":
		if val, ok := req.Value.(float64); ok && val >= 1 && val <= 1440 {
			cfg.ScheduleIntervalMinutes = int(val)
			newValue = int(val)
			updated = true
			// Update scheduler
			if h.scheduler != nil {
				h.scheduler.UpdateInterval(time.Duration(int(val)) * time.Minute)
			}
		}
	case "browser_pool_size":
		if val, ok := req.Value.(float64); ok && val >= 1 && val <= 20 {
			cfg.BrowserPoolSize = int(val)
			newValue = int(val)
			updated = true
		}
	case "browser_max_concurrent":
		if val, ok := req.Value.(float64); ok && val >= 1 && val <= 10 {
			cfg.BrowserMaxConcurrent = int(val)
			newValue = int(val)
			updated = true
		}
	case "content_batch_size":
		if val, ok := req.Value.(float64); ok && val >= 5 && val <= 50 {
			cfg.ContentExtractionBatchSize = int(val)
			newValue = int(val)
			updated = true
		}
	case "enable_browser_scraping":
		if val, ok := req.Value.(bool); ok {
			cfg.EnableBrowserScraping = val
			newValue = val
			updated = true
		}
	case "enable_full_content":
		if val, ok := req.Value.(bool); ok {
			cfg.EnableFullContentExtraction = val
			newValue = val
			updated = true
		}
	case "enable_robots_check":
		if val, ok := req.Value.(bool); ok {
			cfg.EnableRobotsTxtCheck = val
			newValue = val
			updated = true
		}
	case "browser_fallback_only":
		if val, ok := req.Value.(bool); ok {
			cfg.BrowserFallbackOnly = val
			newValue = val
			updated = true
		}
	default:
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_SETTING",
				fmt.Sprintf("Setting '%s' is not configurable", req.Setting),
				"Check API documentation for valid settings",
				requestID),
		)
	}

	if !updated {
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewErrorResponse("INVALID_VALUE",
				fmt.Sprintf("Invalid value for setting '%s'", req.Setting),
				"Value must be within allowed range",
				requestID),
		)
	}

	// Update global config
	h.config.Scraper = *cfg

	h.logger.Infof("Updated setting '%s' to %v in profile '%s'", req.Setting, newValue, h.activeProfile)

	response := fiber.Map{
		"success":   true,
		"message":   fmt.Sprintf("Setting '%s' updated successfully", req.Setting),
		"setting":   req.Setting,
		"old_value": req.Value,
		"new_value": newValue,
		"profile":   h.activeProfile,
	}

	return c.JSON(models.NewSuccessResponse(response, requestID))
}

// GetSchedulerStatus returns the current scheduler status
// GET /api/v1/config/scheduler/status
func (h *ConfigHandler) GetSchedulerStatus(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	var isRunning bool
	if h.scheduler != nil {
		isRunning = h.scheduler.IsRunning()
	}

	h.mu.RLock()
	cfg := h.scraperConfigs[h.activeProfile]
	profile := h.activeProfile
	h.mu.RUnlock()

	response := fiber.Map{
		"running":          isRunning,
		"active_profile":   profile,
		"interval_minutes": cfg.ScheduleIntervalMinutes,
		"next_run":         calculateNextRun(cfg.ScheduleIntervalMinutes),
		"enabled":          cfg.ScheduleEnabled,
	}

	return c.JSON(models.NewSuccessResponse(response, requestID))
}

// ResetToDefaults resets the active profile to its default values
// POST /api/v1/config/reset
func (h *ConfigHandler) ResetToDefaults(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	h.mu.Lock()
	defer h.mu.Unlock()

	// Reinitialize profiles to restore defaults
	oldProfile := h.activeProfile
	h.initializeProfiles()

	h.logger.Infof("Reset profile '%s' to default values", oldProfile)

	response := fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Profile '%s' reset to default values", oldProfile),
		"profile": oldProfile,
	}

	return c.JSON(models.NewSuccessResponse(response, requestID))
}

// calculateNextRun estimates the next scheduled run time
func calculateNextRun(intervalMinutes int) time.Time {
	return time.Now().Add(time.Duration(intervalMinutes) * time.Minute)
}

// RestartServer initiates a graceful server restart
// POST /api/v1/config/restart
func (h *ConfigHandler) RestartServer(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	// Parse optional delay parameter (seconds)
	delay := c.QueryInt("delay", 3) // Default 3 seconds
	if delay < 0 || delay > 60 {
		delay = 3
	}

	h.logger.Warnf("Server restart requested via API (delay: %ds, request_id: %s)", delay, requestID)

	// Send immediate response before starting shutdown
	response := fiber.Map{
		"success":            true,
		"message":            "Server restart initiated",
		"delay_seconds":      delay,
		"estimated_downtime": "30-45 seconds",
		"steps": []string{
			"Graceful shutdown in progress",
			"Stopping schedulers and processors",
			"Closing database connections",
			"Restarting container",
			"Server will be back online shortly",
		},
	}

	// Send response first
	if err := c.JSON(models.NewSuccessResponse(response, requestID)); err != nil {
		h.logger.WithError(err).Error("Failed to send restart response")
	}

	// Trigger graceful shutdown after delay (in goroutine)
	go func() {
		time.Sleep(time.Duration(delay) * time.Second)
		h.logger.Warn("Initiating graceful shutdown for restart...")

		// Trigger Fiber app shutdown (will be caught by signal handler in main.go)
		if err := c.App().Shutdown(); err != nil {
			h.logger.WithError(err).Error("Failed to shutdown server")
		}
	}()

	return nil
}

// GetRestartStatus checks if server is ready after restart
// GET /api/v1/config/restart/status
func (h *ConfigHandler) GetRestartStatus(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	// If this endpoint responds, server is running
	response := fiber.Map{
		"status":      "running",
		"ready":       true,
		"message":     "Server is operational",
		"uptime_info": "Check /health/metrics for detailed uptime",
	}

	return c.JSON(models.NewSuccessResponse(response, requestID))
}
