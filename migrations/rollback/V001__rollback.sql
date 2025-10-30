-- ============================================================================
-- Rollback Script: V001__create_base_schema.sql
-- Description: Rollback base schema creation
-- Version: 1.0.0
-- Author: NieuwsScraper Team
-- Date: 2025-10-30
-- WARNING: This will delete all data in articles, sources, and scraping_jobs tables
-- ============================================================================

-- Confirm before execution
DO $$
BEGIN
    RAISE NOTICE '⚠️  WARNING: This rollback will DROP tables and DELETE ALL DATA!';
    RAISE NOTICE 'Tables to be dropped: articles, sources, scraping_jobs, schema_migrations';
    RAISE NOTICE 'Press Ctrl+C within 5 seconds to cancel...';
    PERFORM pg_sleep(5);
END $$;

-- ============================================================================
-- DROP VIEWS
-- ============================================================================

DROP VIEW IF EXISTS v_recent_scraping_activity CASCADE;
DROP VIEW IF EXISTS v_article_stats CASCADE;
DROP VIEW IF EXISTS v_active_sources CASCADE;

RAISE NOTICE 'Dropped 3 views';

-- ============================================================================
-- DROP TRIGGERS
-- ============================================================================

DROP TRIGGER IF EXISTS trg_articles_updated_at ON articles;
DROP TRIGGER IF EXISTS trg_sources_updated_at ON sources;

RAISE NOTICE 'Dropped 2 triggers';

-- ============================================================================
-- DROP FUNCTIONS
-- ============================================================================

DROP FUNCTION IF EXISTS trigger_set_updated_at() CASCADE;

RAISE NOTICE 'Dropped 1 function';

-- ============================================================================
-- DROP TABLES (in correct order due to foreign keys)
-- ============================================================================

-- Drop scraping_jobs first (no dependencies)
DROP TABLE IF EXISTS scraping_jobs CASCADE;
RAISE NOTICE 'Dropped table: scraping_jobs';

-- Drop articles (no dependencies on other tables)
DROP TABLE IF EXISTS articles CASCADE;
RAISE NOTICE 'Dropped table: articles';

-- Drop sources
DROP TABLE IF EXISTS sources CASCADE;
RAISE NOTICE 'Dropped table: sources';

-- ============================================================================
-- REMOVE MIGRATION RECORD
-- ============================================================================

DELETE FROM schema_migrations WHERE version = 'V001';
RAISE NOTICE 'Removed migration record: V001';

-- ============================================================================
-- DROP EXTENSIONS (optional - only if not used by other schemas)
-- ============================================================================

-- Uncomment these if you want to remove extensions
-- DROP EXTENSION IF EXISTS "pg_trgm";
-- DROP EXTENSION IF EXISTS "uuid-ossp";

-- ============================================================================
-- FINALIZE ROLLBACK
-- ============================================================================

DO $$ 
BEGIN 
    RAISE NOTICE '✅ Rollback V001 completed successfully';
    RAISE NOTICE 'All base schema objects have been removed';
    RAISE NOTICE 'Database is now in pre-V001 state';
END $$;