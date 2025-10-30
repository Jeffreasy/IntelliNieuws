-- ============================================================================
-- Migration: V001__create_base_schema.sql
-- Description: Create base schema with articles, sources, and scraping_jobs
-- Version: 1.0.0
-- Author: NieuwsScraper Team
-- Date: 2025-10-30
-- Dependencies: None
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- For similarity searches

-- ============================================================================
-- SCHEMA VERSION TRACKING
-- ============================================================================
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(50) PRIMARY KEY,
    description TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    applied_by VARCHAR(100) DEFAULT CURRENT_USER,
    execution_time_ms INTEGER,
    checksum VARCHAR(64)
);

COMMENT ON TABLE schema_migrations IS 'Tracks applied database migrations for version control';

-- ============================================================================
-- CORE TABLES
-- ============================================================================

-- Sources table: Configuration for news sources
CREATE TABLE IF NOT EXISTS sources (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    domain VARCHAR(255) NOT NULL UNIQUE,
    rss_feed_url VARCHAR(1000),
    
    -- Scraping configuration
    use_rss BOOLEAN NOT NULL DEFAULT TRUE,
    use_dynamic BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    rate_limit_seconds INTEGER NOT NULL DEFAULT 5 CHECK (rate_limit_seconds >= 0),
    max_articles_per_scrape INTEGER DEFAULT 100 CHECK (max_articles_per_scrape > 0),
    
    -- Metadata
    last_scraped_at TIMESTAMPTZ,
    last_success_at TIMESTAMPTZ,
    last_error TEXT,
    consecutive_failures INTEGER NOT NULL DEFAULT 0 CHECK (consecutive_failures >= 0),
    total_articles_scraped BIGINT NOT NULL DEFAULT 0 CHECK (total_articles_scraped >= 0),
    
    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) DEFAULT CURRENT_USER,
    
    -- Constraints
    CONSTRAINT chk_sources_scraping_method CHECK (use_rss = TRUE OR use_dynamic = TRUE),
    CONSTRAINT chk_sources_domain_format CHECK (domain ~ '^[a-z0-9.-]+\.[a-z]{2,}$')
);

-- Indexes for sources
CREATE INDEX idx_sources_is_active ON sources(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_sources_domain_lookup ON sources(domain) WHERE is_active = TRUE;
CREATE INDEX idx_sources_last_scraped ON sources(last_scraped_at) WHERE is_active = TRUE;

COMMENT ON TABLE sources IS 'Configuration and metadata for news sources';
COMMENT ON COLUMN sources.rate_limit_seconds IS 'Minimum seconds between scraping requests';
COMMENT ON COLUMN sources.consecutive_failures IS 'Counter for consecutive scraping failures (reset on success)';

-- Articles table: Stores news articles
CREATE TABLE IF NOT EXISTS articles (
    id BIGSERIAL PRIMARY KEY,
    
    -- Core article data
    title VARCHAR(500) NOT NULL,
    summary TEXT,
    content TEXT,
    url VARCHAR(1000) NOT NULL,
    
    -- Publishing information
    published TIMESTAMPTZ NOT NULL,
    source VARCHAR(100) NOT NULL,
    author VARCHAR(200),
    
    -- Categorization
    category VARCHAR(100),
    keywords TEXT[],
    
    -- Media
    image_url VARCHAR(1000),
    
    -- Content management
    content_hash VARCHAR(64), -- SHA256 hash for deduplication
    content_extracted BOOLEAN NOT NULL DEFAULT FALSE,
    content_extracted_at TIMESTAMPTZ,
    
    -- AI processing flags
    ai_processed BOOLEAN NOT NULL DEFAULT FALSE,
    ai_processed_at TIMESTAMPTZ,
    ai_sentiment DECIMAL(3,2) CHECK (ai_sentiment BETWEEN -1.0 AND 1.0),
    ai_sentiment_label VARCHAR(20) CHECK (ai_sentiment_label IN ('positive', 'negative', 'neutral')),
    ai_summary TEXT,
    ai_categories JSONB,
    ai_entities JSONB,
    ai_keywords JSONB,
    ai_stock_tickers JSONB,
    ai_error TEXT,
    
    -- Stock data (cached)
    stock_data JSONB,
    stock_data_updated_at TIMESTAMPTZ,
    
    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT uq_articles_url UNIQUE (url),
    CONSTRAINT uq_articles_content_hash UNIQUE (content_hash),
    CONSTRAINT chk_articles_url_format CHECK (url ~ '^https?://'),
    CONSTRAINT chk_articles_sentiment_label CHECK (
        (ai_sentiment IS NULL AND ai_sentiment_label IS NULL) OR
        (ai_sentiment IS NOT NULL AND ai_sentiment_label IS NOT NULL)
    )
);

-- Performance indexes for articles
CREATE INDEX idx_articles_source ON articles(source);
CREATE INDEX idx_articles_published_desc ON articles(published DESC);
CREATE INDEX idx_articles_category ON articles(category) WHERE category IS NOT NULL;
CREATE INDEX idx_articles_created_at_desc ON articles(created_at DESC);
CREATE INDEX idx_articles_content_hash ON articles(content_hash) WHERE content_hash IS NOT NULL;

-- Content processing indexes
CREATE INDEX idx_articles_needs_content ON articles(content_extracted, created_at DESC) 
    WHERE content_extracted = FALSE;
CREATE INDEX idx_articles_needs_ai ON articles(ai_processed, created_at DESC) 
    WHERE ai_processed = FALSE;

-- Composite indexes for common queries
CREATE INDEX idx_articles_source_published ON articles(source, published DESC);
CREATE INDEX idx_articles_category_published ON articles(category, published DESC) 
    WHERE category IS NOT NULL;

-- GIN indexes for array and JSONB columns
CREATE INDEX idx_articles_keywords_gin ON articles USING GIN(keywords) 
    WHERE keywords IS NOT NULL;
CREATE INDEX idx_articles_ai_categories_gin ON articles USING GIN(ai_categories) 
    WHERE ai_categories IS NOT NULL;
CREATE INDEX idx_articles_ai_entities_gin ON articles USING GIN(ai_entities) 
    WHERE ai_entities IS NOT NULL;
CREATE INDEX idx_articles_ai_keywords_gin ON articles USING GIN(ai_keywords) 
    WHERE ai_keywords IS NOT NULL;
CREATE INDEX idx_articles_stock_tickers_gin ON articles USING GIN(ai_stock_tickers) 
    WHERE ai_stock_tickers IS NOT NULL;

-- Full-text search indexes
CREATE INDEX idx_articles_title_fts ON articles USING GIN(to_tsvector('english', title));
CREATE INDEX idx_articles_summary_fts ON articles USING GIN(to_tsvector('english', COALESCE(summary, '')));
CREATE INDEX idx_articles_content_fts ON articles USING GIN(to_tsvector('english', COALESCE(content, '')));
CREATE INDEX idx_articles_combined_fts ON articles USING GIN(
    to_tsvector('english', title || ' ' || COALESCE(summary, '') || ' ' || COALESCE(content, ''))
);

COMMENT ON TABLE articles IS 'News articles with AI enrichment and stock data';
COMMENT ON COLUMN articles.content_hash IS 'SHA256 hash for duplicate detection';
COMMENT ON COLUMN articles.ai_sentiment IS 'Sentiment score from -1.0 (negative) to 1.0 (positive)';

-- Scraping jobs table: Tracks scraping execution
CREATE TABLE IF NOT EXISTS scraping_jobs (
    id BIGSERIAL PRIMARY KEY,
    job_uuid UUID NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    
    -- Job configuration
    source VARCHAR(100) NOT NULL,
    scraping_method VARCHAR(20) NOT NULL CHECK (scraping_method IN ('rss', 'dynamic', 'hybrid')),
    
    -- Job status
    status VARCHAR(20) NOT NULL DEFAULT 'pending' 
        CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
    
    -- Execution details
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    execution_time_ms INTEGER,
    
    -- Results
    articles_found INTEGER DEFAULT 0 CHECK (articles_found >= 0),
    articles_new INTEGER DEFAULT 0 CHECK (articles_new >= 0),
    articles_updated INTEGER DEFAULT 0 CHECK (articles_updated >= 0),
    articles_skipped INTEGER DEFAULT 0 CHECK (articles_skipped >= 0),
    
    -- Error handling
    error TEXT,
    error_code VARCHAR(50),
    retry_count INTEGER NOT NULL DEFAULT 0 CHECK (retry_count >= 0),
    max_retries INTEGER NOT NULL DEFAULT 3 CHECK (max_retries >= 0),
    
    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) DEFAULT CURRENT_USER,
    
    -- Constraints
    CONSTRAINT chk_scraping_jobs_completion CHECK (
        (status IN ('pending', 'running') AND completed_at IS NULL) OR
        (status IN ('completed', 'failed', 'cancelled') AND completed_at IS NOT NULL)
    ),
    CONSTRAINT chk_scraping_jobs_timing CHECK (
        started_at IS NULL OR completed_at IS NULL OR started_at <= completed_at
    )
);

-- Indexes for scraping_jobs
CREATE INDEX idx_scraping_jobs_source ON scraping_jobs(source);
CREATE INDEX idx_scraping_jobs_status ON scraping_jobs(status);
CREATE INDEX idx_scraping_jobs_created_at_desc ON scraping_jobs(created_at DESC);
CREATE INDEX idx_scraping_jobs_uuid ON scraping_jobs(job_uuid);
CREATE INDEX idx_scraping_jobs_source_status ON scraping_jobs(source, status, created_at DESC);

COMMENT ON TABLE scraping_jobs IS 'Tracks execution and results of scraping jobs';

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at trigger to relevant tables
CREATE TRIGGER trg_sources_updated_at
    BEFORE UPDATE ON sources
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER trg_articles_updated_at
    BEFORE UPDATE ON articles
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================================
-- VIEWS
-- ============================================================================

-- View: Active sources ready for scraping
CREATE OR REPLACE VIEW v_active_sources AS
SELECT 
    s.id,
    s.name,
    s.domain,
    s.rss_feed_url,
    s.use_rss,
    s.use_dynamic,
    s.rate_limit_seconds,
    s.last_scraped_at,
    s.consecutive_failures,
    s.total_articles_scraped,
    CASE 
        WHEN s.last_scraped_at IS NULL THEN TRUE
        WHEN CURRENT_TIMESTAMP - s.last_scraped_at > (s.rate_limit_seconds || ' seconds')::INTERVAL THEN TRUE
        ELSE FALSE
    END AS ready_to_scrape
FROM sources s
WHERE s.is_active = TRUE
  AND s.consecutive_failures < 5; -- Disable after 5 consecutive failures

COMMENT ON VIEW v_active_sources IS 'Active sources ready for scraping based on rate limits';

-- View: Article statistics by source
CREATE OR REPLACE VIEW v_article_stats AS
SELECT 
    a.source,
    s.name AS source_name,
    COUNT(*) AS total_articles,
    COUNT(*) FILTER (WHERE a.created_at >= CURRENT_DATE - INTERVAL '24 hours') AS articles_today,
    COUNT(*) FILTER (WHERE a.created_at >= CURRENT_DATE - INTERVAL '7 days') AS articles_week,
    COUNT(*) FILTER (WHERE a.ai_processed = TRUE) AS ai_processed_count,
    COUNT(*) FILTER (WHERE a.content_extracted = TRUE) AS content_extracted_count,
    MAX(a.published) AS latest_article_date,
    MIN(a.published) AS oldest_article_date,
    AVG(a.ai_sentiment) FILTER (WHERE a.ai_sentiment IS NOT NULL) AS avg_sentiment
FROM articles a
LEFT JOIN sources s ON a.source = s.domain
GROUP BY a.source, s.name;

COMMENT ON VIEW v_article_stats IS 'Aggregated statistics per news source';

-- View: Recent scraping activity
CREATE OR REPLACE VIEW v_recent_scraping_activity AS
SELECT 
    sj.job_uuid,
    sj.source,
    s.name AS source_name,
    sj.scraping_method,
    sj.status,
    sj.articles_new,
    sj.articles_updated,
    sj.started_at,
    sj.completed_at,
    sj.execution_time_ms,
    sj.error,
    sj.retry_count
FROM scraping_jobs sj
LEFT JOIN sources s ON sj.source = s.domain
ORDER BY sj.created_at DESC
LIMIT 100;

COMMENT ON VIEW v_recent_scraping_activity IS 'Last 100 scraping jobs with results';

-- ============================================================================
-- INITIAL DATA
-- ============================================================================

-- Insert default Dutch news sources
INSERT INTO sources (name, domain, rss_feed_url, use_rss, use_dynamic, is_active, rate_limit_seconds, max_articles_per_scrape) 
VALUES 
    ('NU.nl', 'nu.nl', 'https://www.nu.nl/rss', TRUE, FALSE, TRUE, 5, 100),
    ('AD.nl', 'ad.nl', 'https://www.ad.nl/rss.xml', TRUE, FALSE, TRUE, 5, 100),
    ('NOS.nl', 'nos.nl', 'https://feeds.nos.nl/nosnieuwsalgemeen', TRUE, FALSE, TRUE, 5, 100),
    ('Telegraaf', 'telegraaf.nl', 'https://www.telegraaf.nl/rss', TRUE, FALSE, TRUE, 5, 100),
    ('RTL Nieuws', 'rtlnieuws.nl', 'https://www.rtlnieuws.nl/service/rss/algemeen.xml', TRUE, FALSE, TRUE, 5, 100)
ON CONFLICT (domain) DO NOTHING;

-- ============================================================================
-- PERMISSIONS (Adjust based on your security model)
-- ============================================================================

-- Grant read access to views
GRANT SELECT ON v_active_sources TO PUBLIC;
GRANT SELECT ON v_article_stats TO PUBLIC;
GRANT SELECT ON v_recent_scraping_activity TO PUBLIC;

-- ============================================================================
-- FINALIZE MIGRATION
-- ============================================================================

-- Update statistics for query planner
ANALYZE sources;
ANALYZE articles;
ANALYZE scraping_jobs;

-- Record this migration
INSERT INTO schema_migrations (version, description, checksum) 
VALUES (
    'V001',
    'Create base schema with articles, sources, and scraping_jobs',
    'base_schema_v1'
) ON CONFLICT (version) DO NOTHING;

-- Success notification
DO $$ 
BEGIN 
    RAISE NOTICE '✅ Migration V001 completed successfully';
    RAISE NOTICE 'Created tables: sources, articles, scraping_jobs, schema_migrations';
    RAISE NOTICE 'Created views: v_active_sources, v_article_stats, v_recent_scraping_activity';
    RAISE NOTICE 'Created % indexes on articles table', (
        SELECT COUNT(*) 
        FROM pg_indexes 
        WHERE tablename = 'articles'
    );
END $$;