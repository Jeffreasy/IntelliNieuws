package models

// ============================================================================
// EMAIL STATUS CONSTANTS
// ============================================================================

// Email status values (matches database CHECK constraint)
const (
	EmailStatusPending    = "pending"
	EmailStatusProcessing = "processing"
	EmailStatusProcessed  = "processed"
	EmailStatusFailed     = "failed"
	EmailStatusIgnored    = "ignored"
	EmailStatusSpam       = "spam"
)

// Email importance levels (matches database CHECK constraint)
const (
	EmailImportanceLow    = "low"
	EmailImportanceNormal = "normal"
	EmailImportanceHigh   = "high"
)

// ============================================================================
// SCRAPING METHOD CONSTANTS
// ============================================================================

// Scraping method values (matches database CHECK constraint)
const (
	ScrapingMethodRSS     = "rss"
	ScrapingMethodDynamic = "dynamic"
	ScrapingMethodHybrid  = "hybrid"
)

// ============================================================================
// SENTIMENT CONSTANTS
// ============================================================================

// Sentiment labels (matches database CHECK constraint)
const (
	SentimentPositive = "positive"
	SentimentNegative = "negative"
	SentimentNeutral  = "neutral"
)

// ============================================================================
// DEFAULT VALUES
// ============================================================================

const (
	// Email defaults
	DefaultEmailMaxRetries = 3
	DefaultEmailRetryHours = 24

	// Scraping defaults
	DefaultScrapingMaxRetries   = 3
	DefaultRateLimitSeconds     = 5
	DefaultMaxArticlesPerScrape = 100

	// Analytics defaults
	DefaultTrendingHoursBack   = 24
	DefaultTrendingMinArticles = 3
	DefaultTrendingLimit       = 20

	// Pagination defaults
	DefaultPageLimit  = 50
	DefaultPageOffset = 0
	MaxPageLimit      = 1000
)

// ============================================================================
// VALIDATION CONSTANTS
// ============================================================================

const (
	// URL validation
	MinURLLength = 10
	MaxURLLength = 1000

	// Title validation
	MinTitleLength = 3
	MaxTitleLength = 500

	// Content validation
	MaxSummaryLength = 2000
	MaxContentLength = 1000000 // 1MB

	// Email validation
	MaxSubjectLength = 500
	MaxSenderLength  = 255
)

// ============================================================================
// CACHE KEY PREFIXES
// ============================================================================

const (
	CacheKeyPrefixArticles  = "articles:"
	CacheKeyPrefixTrending  = "trending:"
	CacheKeyPrefixSentiment = "sentiment:"
	CacheKeyPrefixEmails    = "emails:"
	CacheKeyPrefixSources   = "sources:"
	CacheKeyPrefixStats     = "stats:"
)

// ============================================================================
// TIME WINDOWS
// ============================================================================

const (
	// Data retention
	EmailRetentionDays       = 90
	ScrapingJobRetentionDays = 30

	// Analytics windows
	AnalyticsWindow24h = 24
	AnalyticsWindow7d  = 168 // 7 * 24 hours
	AnalyticsWindow30d = 720 // 30 * 24 hours

	// Refresh intervals (minutes)
	MaterializedViewRefreshInterval = 15
	CacheRefreshInterval            = 5
)
