-- ============================================================================
-- Rollback Script: V003__create_analytics_views.sql
-- Description: Rollback analytics materialized views
-- Version: 1.0.0
-- Author: NieuwsScraper Team
-- Date: 2025-10-30
-- WARNING: This will delete all pre-calculated analytics data
-- ============================================================================

-- Confirm before execution
DO $$
BEGIN
    RAISE NOTICE '⚠️  WARNING: This rollback will DROP all analytics views and materialized views!';
    RAISE NOTICE 'Pre-calculated analytics data will be lost';
    RAISE NOTICE 'Press Ctrl+C within 5 seconds to cancel...';
    PERFORM pg_sleep(5);
END $$;

-- ============================================================================
-- DROP HELPER VIEWS
-- ============================================================================

DROP VIEW IF EXISTS v_hot_entities_7d CASCADE;
DROP VIEW IF EXISTS v_sentiment_trends_7d CASCADE;
DROP VIEW IF EXISTS v_trending_keywords_24h CASCADE;

RAISE NOTICE 'Dropped 3 helper views';

-- ============================================================================
-- DROP FUNCTIONS
-- ============================================================================

DROP FUNCTION IF EXISTS refresh_analytics_views(BOOLEAN) CASCADE;
DROP FUNCTION IF EXISTS get_entity_sentiment_analysis(TEXT, INTEGER) CASCADE;
DROP FUNCTION IF EXISTS get_trending_topics(INTEGER, INTEGER, INTEGER) CASCADE;

RAISE NOTICE 'Dropped 3 analytics functions';

-- ============================================================================
-- DROP MATERIALIZED VIEWS
-- ============================================================================

DROP MATERIALIZED VIEW IF EXISTS mv_entity_mentions CASCADE;
RAISE NOTICE 'Dropped materialized view: mv_entity_mentions';

DROP MATERIALIZED VIEW IF EXISTS mv_sentiment_timeline CASCADE;
RAISE NOTICE 'Dropped materialized view: mv_sentiment_timeline';

DROP MATERIALIZED VIEW IF EXISTS mv_trending_keywords CASCADE;
RAISE NOTICE 'Dropped materialized view: mv_trending_keywords';

-- ============================================================================
-- REMOVE MIGRATION RECORD
-- ============================================================================

DELETE FROM schema_migrations WHERE version = 'V003';
RAISE NOTICE 'Removed migration record: V003';

-- ============================================================================
-- FINALIZE ROLLBACK
-- ============================================================================

DO $$ 
BEGIN 
    RAISE NOTICE '✅ Rollback V003 completed successfully';
    RAISE NOTICE 'All analytics views and materialized views have been removed';
    RAISE NOTICE 'Database is now in post-V002 state';
    RAISE NOTICE 'Note: Source data in articles table is preserved';
END $$;