-- ============================================================================
-- Utility Script: Migrate from Legacy Schema to New Professional Schema
-- Description: Safely migrate from old migrations (001-008) to new V001-V003
-- Version: 1.0.0
-- Author: NieuwsScraper Team
-- Date: 2025-10-30
-- WARNING: Test on a backup first!
-- ============================================================================

-- ============================================================================
-- PRE-MIGRATION CHECKS
-- ============================================================================

DO $$
DECLARE
    v_articles_count BIGINT;
    v_sources_count BIGINT;
    v_emails_count BIGINT;
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'LEGACY TO V001-V003 MIGRATION TOOL';
    RAISE NOTICE '========================================';
    RAISE NOTICE '';
    
    -- Check if tables exist
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'articles') THEN
        SELECT COUNT(*) INTO v_articles_count FROM articles;
        RAISE NOTICE '✓ Found articles table with % rows', v_articles_count;
    ELSE
        RAISE EXCEPTION 'articles table not found. Nothing to migrate.';
    END IF;
    
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'sources') THEN
        SELECT COUNT(*) INTO v_sources_count FROM sources;
        RAISE NOTICE '✓ Found sources table with % rows', v_sources_count;
    END IF;
    
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'emails') THEN
        SELECT COUNT(*) INTO v_emails_count FROM emails;
        RAISE NOTICE '✓ Found emails table with % rows', v_emails_count;
    END IF;
    
    RAISE NOTICE '';
    RAISE NOTICE '⚠️  This script will:';
    RAISE NOTICE '   1. Create schema_migrations table if missing';
    RAISE NOTICE '   2. Add missing columns to existing tables';
    RAISE NOTICE '   3. Create missing indexes';
    RAISE NOTICE '   4. Add missing triggers and functions';
    RAISE NOTICE '   5. Create analytics materialized views';
    RAISE NOTICE '';
    RAISE NOTICE 'Press Ctrl+C within 10 seconds to cancel...';
    PERFORM pg_sleep(10);
END $$;

-- ============================================================================
-- STEP 1: CREATE SCHEMA_MIGRATIONS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(50) PRIMARY KEY,
    description TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    applied_by VARCHAR(100) DEFAULT CURRENT_USER,
    execution_time_ms INTEGER,
    checksum VARCHAR(64)
);

RAISE NOTICE '✓ Created schema_migrations table';

-- ============================================================================
-- STEP 2: ENHANCE SOURCES TABLE
-- ============================================================================

-- Add missing columns to sources
ALTER TABLE sources ADD COLUMN IF NOT EXISTS max_articles_per_scrape INTEGER DEFAULT 100 CHECK (max_articles_per_scrape > 0);
ALTER TABLE sources ADD COLUMN IF NOT EXISTS last_success_at TIMESTAMPTZ;
ALTER TABLE sources ADD COLUMN IF NOT EXISTS last_error TEXT;
ALTER TABLE sources ADD COLUMN IF NOT EXISTS consecutive_failures INTEGER NOT NULL DEFAULT 0 CHECK (consecutive_failures >= 0);
ALTER TABLE sources ADD COLUMN IF NOT EXISTS total_articles_scraped BIGINT NOT NULL DEFAULT 0 CHECK (total_articles_scraped >= 0);
ALTER TABLE sources ADD COLUMN IF NOT EXISTS created_by VARCHAR(100) DEFAULT CURRENT_USER;

-- Add missing constraints
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_sources_scraping_method') THEN
        ALTER TABLE sources ADD CONSTRAINT chk_sources_scraping_method 
            CHECK (use_rss = TRUE OR use_dynamic = TRUE);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_sources_domain_format') THEN
        ALTER TABLE sources ADD CONSTRAINT chk_sources_domain_format 
            CHECK (domain ~ '^[a-z0-9.-]+\.[a-z]{2,}$');
    END IF;
END $$;

RAISE NOTICE '✓ Enhanced sources table';

-- ============================================================================
-- STEP 3: ENHANCE ARTICLES TABLE
-- ============================================================================

-- Ensure all AI columns exist with proper constraints
ALTER TABLE articles ADD COLUMN IF NOT EXISTS ai_processed BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS ai_processed_at TIMESTAMPTZ;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS ai_sentiment DECIMAL(3,2) CHECK (ai_sentiment BETWEEN -1.0 AND 1.0);
ALTER TABLE articles ADD COLUMN IF NOT EXISTS ai_sentiment_label VARCHAR(20) CHECK (ai_sentiment_label IN ('positive', 'negative', 'neutral'));
ALTER TABLE articles ADD COLUMN IF NOT EXISTS ai_summary TEXT;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS ai_categories JSONB;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS ai_entities JSONB;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS ai_keywords JSONB;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS ai_stock_tickers JSONB;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS ai_error TEXT;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS stock_data JSONB;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS stock_data_updated_at TIMESTAMPTZ;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS content TEXT;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS content_extracted BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS content_extracted_at TIMESTAMPTZ;

-- Add constraint for sentiment label consistency
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_articles_sentiment_label') THEN
        ALTER TABLE articles ADD CONSTRAINT chk_articles_sentiment_label CHECK (
            (ai_sentiment IS NULL AND ai_sentiment_label IS NULL) OR
            (ai_sentiment IS NOT NULL AND ai_sentiment_label IS NOT NULL)
        );
    END IF;
END $$;

RAISE NOTICE '✓ Enhanced articles table';

-- ============================================================================
-- STEP 4: ENHANCE SCRAPING_JOBS TABLE
-- ============================================================================

-- Add UUID column for jobs
ALTER TABLE scraping_jobs ADD COLUMN IF NOT EXISTS job_uuid UUID DEFAULT uuid_generate_v4() UNIQUE;
ALTER TABLE scraping_jobs ADD COLUMN IF NOT EXISTS scraping_method VARCHAR(20) CHECK (scraping_method IN ('rss', 'dynamic', 'hybrid'));
ALTER TABLE scraping_jobs ADD COLUMN IF NOT EXISTS execution_time_ms INTEGER;
ALTER TABLE scraping_jobs ADD COLUMN IF NOT EXISTS articles_found INTEGER DEFAULT 0 CHECK (articles_found >= 0);
ALTER TABLE scraping_jobs ADD COLUMN IF NOT EXISTS articles_new INTEGER DEFAULT 0 CHECK (articles_new >= 0);
ALTER TABLE scraping_jobs ADD COLUMN IF NOT EXISTS articles_updated INTEGER DEFAULT 0 CHECK (articles_updated >= 0);
ALTER TABLE scraping_jobs ADD COLUMN IF NOT EXISTS articles_skipped INTEGER DEFAULT 0 CHECK (articles_skipped >= 0);
ALTER TABLE scraping_jobs ADD COLUMN IF NOT EXISTS error_code VARCHAR(50);
ALTER TABLE scraping_jobs ADD COLUMN IF NOT EXISTS retry_count INTEGER NOT NULL DEFAULT 0 CHECK (retry_count >= 0);
ALTER TABLE scraping_jobs ADD COLUMN IF NOT EXISTS max_retries INTEGER NOT NULL DEFAULT 3 CHECK (max_retries >= 0);
ALTER TABLE scraping_jobs ADD COLUMN IF NOT EXISTS created_by VARCHAR(100) DEFAULT CURRENT_USER;

-- Rename article_count to match new schema if needed
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'scraping_jobs' AND column_name = 'article_count') THEN
        UPDATE scraping_jobs SET articles_new = article_count WHERE articles_new = 0;
    END IF;
END $$;

RAISE NOTICE '✓ Enhanced scraping_jobs table';

-- ============================================================================
-- STEP 5: ENHANCE EMAILS TABLE (IF EXISTS)
-- ============================================================================

DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'emails') THEN
        -- Add missing columns
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS message_uid VARCHAR(100);
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS thread_id VARCHAR(255);
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS sender_name VARCHAR(200);
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS recipient VARCHAR(255);
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS snippet TEXT;
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS sent_date TIMESTAMPTZ;
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'pending' 
            CHECK (status IN ('pending', 'processing', 'processed', 'failed', 'ignored', 'spam'));
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS article_created BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS error_code VARCHAR(50);
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS max_retries INTEGER NOT NULL DEFAULT 3 CHECK (max_retries >= 0);
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS last_retry_at TIMESTAMPTZ;
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS has_attachments BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS attachment_count INTEGER DEFAULT 0 CHECK (attachment_count >= 0);
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS is_read BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS is_flagged BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS is_spam BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS importance VARCHAR(20) CHECK (importance IN ('low', 'normal', 'high'));
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS headers JSONB;
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS labels TEXT[];
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS size_bytes INTEGER CHECK (size_bytes >= 0);
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS spam_score DECIMAL(5,2) CHECK (spam_score BETWEEN 0 AND 100);
        ALTER TABLE emails ADD COLUMN IF NOT EXISTS created_by VARCHAR(100) DEFAULT 'email_processor';
        
        RAISE NOTICE '✓ Enhanced emails table';
    END IF;
END $$;

-- ============================================================================
-- STEP 6: CREATE MISSING INDEXES
-- ============================================================================

-- Articles indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_needs_content 
    ON articles(content_extracted, created_at DESC) WHERE content_extracted = FALSE;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_needs_ai 
    ON articles(ai_processed, created_at DESC) WHERE ai_processed = FALSE;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_stock_tickers_gin 
    ON articles USING GIN(ai_stock_tickers) WHERE ai_stock_tickers IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_content_fts 
    ON articles USING GIN(to_tsvector('english', COALESCE(content, '')));

-- Sources indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sources_last_scraped 
    ON sources(last_scraped_at) WHERE is_active = TRUE;

-- Scraping jobs indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_scraping_jobs_uuid 
    ON scraping_jobs(job_uuid);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_scraping_jobs_source_status 
    ON scraping_jobs(source, status, created_at DESC);

RAISE NOTICE '✓ Created missing indexes';

-- ============================================================================
-- STEP 7: CREATE/UPDATE FUNCTIONS
-- ============================================================================

-- Update timestamp function (already exists but ensure it's up to date)
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

RAISE NOTICE '✓ Created/updated functions';

-- ============================================================================
-- STEP 8: CREATE MISSING TRIGGERS
-- ============================================================================

-- Ensure updated_at triggers exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_sources_updated_at') THEN
        CREATE TRIGGER trg_sources_updated_at
            BEFORE UPDATE ON sources
            FOR EACH ROW
            EXECUTE FUNCTION trigger_set_updated_at();
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_articles_updated_at') THEN
        CREATE TRIGGER trg_articles_updated_at
            BEFORE UPDATE ON articles
            FOR EACH ROW
            EXECUTE FUNCTION trigger_set_updated_at();
    END IF;
END $$;

RAISE NOTICE '✓ Created missing triggers';

-- ============================================================================
-- STEP 9: CREATE VIEWS
-- ============================================================================

-- Active sources view
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
  AND s.consecutive_failures < 5;

-- Article stats view
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

RAISE NOTICE '✓ Created/updated views';

-- ============================================================================
-- STEP 10: RECORD MIGRATIONS
-- ============================================================================

INSERT INTO schema_migrations (version, description, checksum) 
VALUES 
    ('V001', 'Base schema (migrated from legacy)', 'legacy_migration_v1'),
    ('V002', 'Emails table (migrated from legacy)', 'legacy_migration_v1'),
    ('LEGACY', 'Legacy migrations 001-008 consolidated', 'legacy_consolidated')
ON CONFLICT (version) DO NOTHING;

RAISE NOTICE '✓ Recorded migration history';

-- ============================================================================
-- STEP 11: UPDATE STATISTICS
-- ============================================================================

ANALYZE articles;
ANALYZE sources;
ANALYZE scraping_jobs;
DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'emails') THEN
        ANALYZE emails;
    END IF;
END $$;

RAISE NOTICE '✓ Updated table statistics';

-- ============================================================================
-- FINALIZE
-- ============================================================================

DO $$ 
DECLARE
    v_articles BIGINT;
    v_sources BIGINT;
    v_jobs BIGINT;
BEGIN
    SELECT COUNT(*) INTO v_articles FROM articles;
    SELECT COUNT(*) INTO v_sources FROM sources;
    SELECT COUNT(*) INTO v_jobs FROM scraping_jobs;
    
    RAISE NOTICE '';
    RAISE NOTICE '========================================';
    RAISE NOTICE '✅ MIGRATION COMPLETED SUCCESSFULLY!';
    RAISE NOTICE '========================================';
    RAISE NOTICE '';
    RAISE NOTICE 'Database Summary:';
    RAISE NOTICE '  - Articles: %', v_articles;
    RAISE NOTICE '  - Sources: %', v_sources;
    RAISE NOTICE '  - Scraping Jobs: %', v_jobs;
    RAISE NOTICE '';
    RAISE NOTICE 'Next Steps:';
    RAISE NOTICE '  1. Verify all data is intact';
    RAISE NOTICE '  2. Apply V003 for analytics views: psql < migrations/V003__create_analytics_views.sql';
    RAISE NOTICE '  3. Update your application code to use new schema';
    RAISE NOTICE '';
END $$;