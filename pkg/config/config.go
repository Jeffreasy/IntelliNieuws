package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	NATS     NATSConfig
	Scraper  ScraperConfig
	API      APIConfig
	Logging  LoggingConfig
	AI       AIConfig
	Stock    StockConfig
	Email    EmailConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	APIPort     int
	ScraperPort int
	Environment string
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host                 string
	Port                 int
	Password             string
	DB                   int
	PoolSize             int
	MinIdleConns         int
	DefaultTTLMinutes    int
	CompressionThreshold int
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL string
}

// ScraperConfig holds scraper-specific configuration
type ScraperConfig struct {
	UserAgent                string
	RateLimitSeconds         int
	MaxConcurrent            int
	TimeoutSeconds           int
	RetryAttempts            int
	TargetSites              []string
	EnableRSSPriority        bool
	EnableDynamicScraping    bool
	EnableRobotsTxtCheck     bool
	EnableDuplicateDetection bool
	ScheduleEnabled          bool
	ScheduleIntervalMinutes  int
	// Content extraction settings (Hybrid scraping)
	EnableFullContentExtraction bool
	ContentExtractionInterval   time.Duration
	ContentExtractionBatchSize  int
	ContentExtractionAsync      bool
	// Browser scraping settings (for JavaScript-rendered content)
	EnableBrowserScraping bool
	BrowserPoolSize       int
	BrowserTimeout        time.Duration
	BrowserWaitAfterLoad  time.Duration
	BrowserFallbackOnly   bool
	BrowserMaxConcurrent  int
	// Stealth features (v3.0)
	EnableUserAgentRotation bool
	EnableProxyRotation     bool
	ProxyProvider           string
	ScraperAPIKey           string
	ScrapeDoToken           string
	ProxyRotationStrategy   string
	ProxyUseOnErrorRate     float64
}

// APIConfig holds API-specific configuration
type APIConfig struct {
	RateLimitRequests      int
	RateLimitWindowSeconds int
	TimeoutSeconds         int
	APIKeyHeader           string
	APIKey                 string
	EnableMetrics          bool
	MetricsPort            int
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// AIConfig holds AI processing configuration
type AIConfig struct {
	// OpenAI settings
	OpenAIAPIKey    string
	OpenAIModel     string
	OpenAIMaxTokens int

	// Processing settings
	Enabled         bool
	AsyncProcessing bool
	BatchSize       int
	ProcessInterval time.Duration
	RetryFailed     bool
	MaxRetries      int

	// Feature toggles
	EnableSentiment  bool
	EnableEntities   bool
	EnableCategories bool
	EnableKeywords   bool
	EnableSummary    bool
	EnableSimilarity bool

	// Cost control
	MaxDailyCost       float64
	RateLimitPerMinute int
	Timeout            time.Duration
}

// StockConfig holds stock API configuration
type StockConfig struct {
	APIKey          string
	APIProvider     string // "fmp" or "alphavantage"
	CacheTTL        time.Duration
	RateLimitPerMin int
	Timeout         time.Duration
	EnableCache     bool
}

// EmailConfig holds email integration configuration
type EmailConfig struct {
	Enabled         bool
	Host            string
	Port            int
	Username        string
	Password        string
	UseTLS          bool
	AllowedSenders  []string
	PollInterval    time.Duration
	MaxRetries      int
	RetryDelay      time.Duration
	MarkAsRead      bool
	DeleteAfterRead bool
	FetchExisting   bool // Fetch existing emails on first run
	MaxDaysBack     int  // How many days back to fetch (default: 30)
}

// Load reads configuration from environment variables and .env file
func Load() (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Read from .env file
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()

	// Attempt to read config file (don't error if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	cfg := &Config{
		Server: ServerConfig{
			APIPort:     v.GetInt("API_PORT"),
			ScraperPort: v.GetInt("SCRAPER_PORT"),
			Environment: v.GetString("ENV"),
		},
		Database: DatabaseConfig{
			Host:     v.GetString("POSTGRES_HOST"),
			Port:     v.GetInt("POSTGRES_PORT"),
			User:     v.GetString("POSTGRES_USER"),
			Password: v.GetString("POSTGRES_PASSWORD"),
			Database: v.GetString("POSTGRES_DB"),
			SSLMode:  v.GetString("POSTGRES_SSL_MODE"),
		},
		Redis: RedisConfig{
			Host:                 v.GetString("REDIS_HOST"),
			Port:                 v.GetInt("REDIS_PORT"),
			Password:             v.GetString("REDIS_PASSWORD"),
			DB:                   v.GetInt("REDIS_DB"),
			PoolSize:             v.GetInt("REDIS_POOL_SIZE"),
			MinIdleConns:         v.GetInt("REDIS_MIN_IDLE_CONNS"),
			DefaultTTLMinutes:    v.GetInt("CACHE_DEFAULT_TTL_MINUTES"),
			CompressionThreshold: v.GetInt("CACHE_COMPRESSION_THRESHOLD"),
		},
		NATS: NATSConfig{
			URL: v.GetString("NATS_URL"),
		},
		Scraper: ScraperConfig{
			UserAgent:                   v.GetString("SCRAPER_USER_AGENT"),
			RateLimitSeconds:            v.GetInt("SCRAPER_RATE_LIMIT_SECONDS"),
			MaxConcurrent:               v.GetInt("SCRAPER_MAX_CONCURRENT"),
			TimeoutSeconds:              v.GetInt("SCRAPER_TIMEOUT_SECONDS"),
			RetryAttempts:               v.GetInt("SCRAPER_RETRY_ATTEMPTS"),
			TargetSites:                 strings.Split(v.GetString("TARGET_SITES"), ","),
			EnableRSSPriority:           v.GetBool("ENABLE_RSS_PRIORITY"),
			EnableDynamicScraping:       v.GetBool("ENABLE_DYNAMIC_SCRAPING"),
			EnableRobotsTxtCheck:        v.GetBool("ENABLE_ROBOTS_TXT_CHECK"),
			EnableDuplicateDetection:    v.GetBool("ENABLE_DUPLICATE_DETECTION"),
			ScheduleEnabled:             v.GetBool("SCRAPER_SCHEDULE_ENABLED"),
			ScheduleIntervalMinutes:     v.GetInt("SCRAPER_SCHEDULE_INTERVAL_MINUTES"),
			EnableFullContentExtraction: v.GetBool("ENABLE_FULL_CONTENT_EXTRACTION"),
			ContentExtractionInterval:   time.Duration(v.GetInt("CONTENT_EXTRACTION_INTERVAL_MINUTES")) * time.Minute,
			ContentExtractionBatchSize:  v.GetInt("CONTENT_EXTRACTION_BATCH_SIZE"),
			ContentExtractionAsync:      v.GetBool("CONTENT_EXTRACTION_ASYNC"),
			EnableBrowserScraping:       v.GetBool("ENABLE_BROWSER_SCRAPING"),
			BrowserPoolSize:             v.GetInt("BROWSER_POOL_SIZE"),
			BrowserTimeout:              time.Duration(v.GetInt("BROWSER_TIMEOUT_SECONDS")) * time.Second,
			BrowserWaitAfterLoad:        time.Duration(v.GetInt("BROWSER_WAIT_AFTER_LOAD_MS")) * time.Millisecond,
			BrowserFallbackOnly:         v.GetBool("BROWSER_FALLBACK_ONLY"),
			BrowserMaxConcurrent:        v.GetInt("BROWSER_MAX_CONCURRENT"),
			EnableUserAgentRotation:     v.GetBool("ENABLE_USER_AGENT_ROTATION"),
			EnableProxyRotation:         v.GetBool("ENABLE_PROXY_ROTATION"),
			ProxyProvider:               v.GetString("PROXY_PROVIDER"),
			ScraperAPIKey:               v.GetString("SCRAPERAPI_KEY"),
			ScrapeDoToken:               v.GetString("SCRAPEDO_TOKEN"),
			ProxyRotationStrategy:       v.GetString("PROXY_ROTATION_STRATEGY"),
			ProxyUseOnErrorRate:         v.GetFloat64("PROXY_USE_ON_ERROR_RATE"),
		},
		API: APIConfig{
			RateLimitRequests:      v.GetInt("API_RATE_LIMIT_REQUESTS"),
			RateLimitWindowSeconds: v.GetInt("API_RATE_LIMIT_WINDOW_SECONDS"),
			TimeoutSeconds:         v.GetInt("API_TIMEOUT_SECONDS"),
			APIKeyHeader:           v.GetString("API_KEY_HEADER"),
			APIKey:                 v.GetString("API_KEY"),
			EnableMetrics:          v.GetBool("ENABLE_METRICS"),
			MetricsPort:            v.GetInt("METRICS_PORT"),
		},
		Logging: LoggingConfig{
			Level:  v.GetString("LOG_LEVEL"),
			Format: v.GetString("LOG_FORMAT"),
		},
		AI: AIConfig{
			OpenAIAPIKey:       v.GetString("OPENAI_API_KEY"),
			OpenAIModel:        v.GetString("OPENAI_MODEL"),
			OpenAIMaxTokens:    v.GetInt("OPENAI_MAX_TOKENS"),
			Enabled:            v.GetBool("AI_ENABLED"),
			AsyncProcessing:    v.GetBool("AI_ASYNC_PROCESSING"),
			BatchSize:          v.GetInt("AI_BATCH_SIZE"),
			ProcessInterval:    time.Duration(v.GetInt("AI_PROCESS_INTERVAL_MINUTES")) * time.Minute,
			RetryFailed:        v.GetBool("AI_RETRY_FAILED"),
			MaxRetries:         v.GetInt("AI_MAX_RETRIES"),
			EnableSentiment:    v.GetBool("AI_ENABLE_SENTIMENT"),
			EnableEntities:     v.GetBool("AI_ENABLE_ENTITIES"),
			EnableCategories:   v.GetBool("AI_ENABLE_CATEGORIES"),
			EnableKeywords:     v.GetBool("AI_ENABLE_KEYWORDS"),
			EnableSummary:      v.GetBool("AI_ENABLE_SUMMARY"),
			EnableSimilarity:   v.GetBool("AI_ENABLE_SIMILARITY"),
			MaxDailyCost:       v.GetFloat64("AI_MAX_DAILY_COST"),
			RateLimitPerMinute: v.GetInt("AI_RATE_LIMIT_PER_MINUTE"),
			Timeout:            time.Duration(v.GetInt("AI_TIMEOUT_SECONDS")) * time.Second,
		},
		Stock: StockConfig{
			APIKey:          v.GetString("STOCK_API_KEY"),
			APIProvider:     v.GetString("STOCK_API_PROVIDER"),
			CacheTTL:        time.Duration(v.GetInt("STOCK_API_CACHE_TTL_MINUTES")) * time.Minute,
			RateLimitPerMin: v.GetInt("STOCK_API_RATE_LIMIT_PER_MINUTE"),
			Timeout:         time.Duration(v.GetInt("STOCK_API_TIMEOUT_SECONDS")) * time.Second,
			EnableCache:     v.GetBool("STOCK_API_ENABLE_CACHE"),
		},
		Email: EmailConfig{
			Enabled:         v.GetBool("EMAIL_ENABLED"),
			Host:            v.GetString("EMAIL_HOST"),
			Port:            v.GetInt("EMAIL_PORT"),
			Username:        v.GetString("EMAIL_USERNAME"),
			Password:        v.GetString("EMAIL_PASSWORD"),
			UseTLS:          v.GetBool("EMAIL_USE_TLS"),
			AllowedSenders:  strings.Split(v.GetString("EMAIL_ALLOWED_SENDERS"), ","),
			PollInterval:    time.Duration(v.GetInt("EMAIL_POLL_INTERVAL_MINUTES")) * time.Minute,
			MaxRetries:      v.GetInt("EMAIL_MAX_RETRIES"),
			RetryDelay:      time.Duration(v.GetInt("EMAIL_RETRY_DELAY_SECONDS")) * time.Second,
			MarkAsRead:      v.GetBool("EMAIL_MARK_AS_READ"),
			DeleteAfterRead: v.GetBool("EMAIL_DELETE_AFTER_READ"),
			FetchExisting:   v.GetBool("EMAIL_FETCH_EXISTING"),
			MaxDaysBack:     v.GetInt("EMAIL_MAX_DAYS_BACK"),
		},
	}

	return cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("API_PORT", 8080)
	v.SetDefault("SCRAPER_PORT", 8081)
	v.SetDefault("ENV", "development")

	// Database defaults
	v.SetDefault("POSTGRES_HOST", "localhost")
	v.SetDefault("POSTGRES_PORT", 5432)
	v.SetDefault("POSTGRES_USER", "scraper")
	v.SetDefault("POSTGRES_PASSWORD", "scraper_password")
	v.SetDefault("POSTGRES_DB", "nieuws_scraper")
	v.SetDefault("POSTGRES_SSL_MODE", "disable")

	// Redis defaults
	v.SetDefault("REDIS_HOST", "localhost")
	v.SetDefault("REDIS_PORT", 6379)
	v.SetDefault("REDIS_PASSWORD", "")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("REDIS_POOL_SIZE", 20)
	v.SetDefault("REDIS_MIN_IDLE_CONNS", 5)
	v.SetDefault("CACHE_DEFAULT_TTL_MINUTES", 5)
	v.SetDefault("CACHE_COMPRESSION_THRESHOLD", 1024)

	// NATS defaults
	v.SetDefault("NATS_URL", "nats://localhost:4222")

	// Scraper defaults
	v.SetDefault("SCRAPER_USER_AGENT", "NieuwsScraper/1.0")
	v.SetDefault("SCRAPER_RATE_LIMIT_SECONDS", 5)
	v.SetDefault("SCRAPER_MAX_CONCURRENT", 3)
	v.SetDefault("SCRAPER_TIMEOUT_SECONDS", 30)
	v.SetDefault("SCRAPER_RETRY_ATTEMPTS", 3)
	v.SetDefault("TARGET_SITES", "nu.nl,ad.nl,nos.nl")
	v.SetDefault("ENABLE_RSS_PRIORITY", true)
	v.SetDefault("ENABLE_DYNAMIC_SCRAPING", false)
	v.SetDefault("ENABLE_ROBOTS_TXT_CHECK", true)
	v.SetDefault("ENABLE_DUPLICATE_DETECTION", true)
	v.SetDefault("SCRAPER_SCHEDULE_ENABLED", false)
	v.SetDefault("SCRAPER_SCHEDULE_INTERVAL_MINUTES", 15)

	// API defaults
	v.SetDefault("API_RATE_LIMIT_REQUESTS", 100)
	v.SetDefault("API_RATE_LIMIT_WINDOW_SECONDS", 60)
	v.SetDefault("API_TIMEOUT_SECONDS", 30)
	v.SetDefault("API_KEY_HEADER", "X-API-Key")
	v.SetDefault("ENABLE_METRICS", true)
	v.SetDefault("METRICS_PORT", 9090)

	// Logging defaults
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_FORMAT", "json")

	// AI defaults
	v.SetDefault("OPENAI_MODEL", "gpt-3.5-turbo")
	v.SetDefault("OPENAI_MAX_TOKENS", 1000)
	v.SetDefault("AI_ENABLED", false)
	v.SetDefault("AI_ASYNC_PROCESSING", true)
	v.SetDefault("AI_BATCH_SIZE", 10)
	v.SetDefault("AI_PROCESS_INTERVAL_MINUTES", 5)
	v.SetDefault("AI_RETRY_FAILED", true)
	v.SetDefault("AI_MAX_RETRIES", 3)
	v.SetDefault("AI_ENABLE_SENTIMENT", true)
	v.SetDefault("AI_ENABLE_ENTITIES", true)
	v.SetDefault("AI_ENABLE_CATEGORIES", true)
	v.SetDefault("AI_ENABLE_KEYWORDS", true)
	v.SetDefault("AI_ENABLE_SUMMARY", false)
	v.SetDefault("AI_ENABLE_SIMILARITY", false)
	v.SetDefault("AI_MAX_DAILY_COST", 10.0)
	v.SetDefault("AI_RATE_LIMIT_PER_MINUTE", 60)
	v.SetDefault("AI_TIMEOUT_SECONDS", 30)

	// Content extraction defaults
	v.SetDefault("ENABLE_FULL_CONTENT_EXTRACTION", false)
	v.SetDefault("CONTENT_EXTRACTION_INTERVAL_MINUTES", 10)
	v.SetDefault("CONTENT_EXTRACTION_BATCH_SIZE", 10)
	v.SetDefault("CONTENT_EXTRACTION_ASYNC", true)

	// Browser scraping defaults
	v.SetDefault("ENABLE_BROWSER_SCRAPING", false)
	v.SetDefault("BROWSER_POOL_SIZE", 3)
	v.SetDefault("BROWSER_TIMEOUT_SECONDS", 15)
	v.SetDefault("BROWSER_WAIT_AFTER_LOAD_MS", 2000)
	v.SetDefault("BROWSER_FALLBACK_ONLY", true)
	v.SetDefault("BROWSER_MAX_CONCURRENT", 2)

	// Stealth defaults (v3.0)
	v.SetDefault("ENABLE_USER_AGENT_ROTATION", false)
	v.SetDefault("ENABLE_PROXY_ROTATION", false)
	v.SetDefault("PROXY_PROVIDER", "scraperapi")
	v.SetDefault("SCRAPERAPI_KEY", "")
	v.SetDefault("SCRAPEDO_TOKEN", "")
	v.SetDefault("PROXY_ROTATION_STRATEGY", "failover")
	v.SetDefault("PROXY_USE_ON_ERROR_RATE", 0.10)

	// Stock API defaults
	v.SetDefault("STOCK_API_PROVIDER", "fmp")
	v.SetDefault("STOCK_API_CACHE_TTL_MINUTES", 5)
	v.SetDefault("STOCK_API_RATE_LIMIT_PER_MINUTE", 30)
	v.SetDefault("STOCK_API_TIMEOUT_SECONDS", 10)
	v.SetDefault("STOCK_API_ENABLE_CACHE", true)

	// Email defaults
	v.SetDefault("EMAIL_ENABLED", false)
	v.SetDefault("EMAIL_HOST", "outlook.office365.com")
	v.SetDefault("EMAIL_PORT", 993)
	v.SetDefault("EMAIL_USE_TLS", true)
	v.SetDefault("EMAIL_ALLOWED_SENDERS", "noreply@x.ai")
	v.SetDefault("EMAIL_POLL_INTERVAL_MINUTES", 5)
	v.SetDefault("EMAIL_MAX_RETRIES", 3)
	v.SetDefault("EMAIL_RETRY_DELAY_SECONDS", 5)
	v.SetDefault("EMAIL_MARK_AS_READ", true)
	v.SetDefault("EMAIL_DELETE_AFTER_READ", false)
	v.SetDefault("EMAIL_FETCH_EXISTING", true)
	v.SetDefault("EMAIL_MAX_DAYS_BACK", 30)
}

// GetDSN returns PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

// GetRedisAddr returns Redis address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetRateLimitDuration returns rate limit duration
func (c *ScraperConfig) GetRateLimitDuration() time.Duration {
	return time.Duration(c.RateLimitSeconds) * time.Second
}

// GetTimeout returns timeout duration
func (c *ScraperConfig) GetTimeout() time.Duration {
	return time.Duration(c.TimeoutSeconds) * time.Second
}

// GetScheduleInterval returns schedule interval duration
func (c *ScraperConfig) GetScheduleInterval() time.Duration {
	return time.Duration(c.ScheduleIntervalMinutes) * time.Minute
}

// GetAPITimeout returns API timeout duration
func (c *APIConfig) GetAPITimeout() time.Duration {
	return time.Duration(c.TimeoutSeconds) * time.Second
}

// GetRateLimitWindow returns rate limit window duration
func (c *APIConfig) GetRateLimitWindow() time.Duration {
	return time.Duration(c.RateLimitWindowSeconds) * time.Second
}

// IsDevelopment checks if running in development mode
func (c *ServerConfig) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction checks if running in production mode
func (c *ServerConfig) IsProduction() bool {
	return c.Environment == "production"
}
