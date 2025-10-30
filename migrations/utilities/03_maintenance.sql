-- ============================================================================
-- Utility Script: Database Maintenance
-- Description: Routine maintenance tasks for optimal database performance
-- Version: 1.0.0
-- Author: NieuwsScraper Team
-- Date: 2025-10-30
-- ============================================================================

\echo '========================================'
\echo 'DATABASE MAINTENANCE UTILITY'
\echo '========================================'
\echo ''

-- ============================================================================
-- MAINTENANCE MENU
-- ============================================================================

\echo 'Available maintenance tasks:'
\echo '  1. Refresh materialized views'
\echo '  2. Clean up old processed emails'
\echo '  3. Clean up old scraping jobs'
\echo '  4. Vacuum and analyze tables'
\echo '  5. Reindex tables'
\echo '  6. Update table statistics'
\echo '  7. Full maintenance (all tasks)'
\echo ''

-- ============================================================================
-- TASK 1: REFRESH MATERIALIZED VIEWS
-- ============================================================================

\echo '1. REFRESHING MATERIALIZED VIEWS'
\echo '---------------------------------'

DO $$
DECLARE
    v_start TIMESTAMPTZ;
    v_duration INTERVAL;
    v_row_count BIGINT;
BEGIN
    IF EXISTS (SELECT FROM pg_matviews WHERE matviewname = 'mv_trending_keywords') THEN
        v_start := CLOCK_TIMESTAMP();
        REFRESH MATERIALIZED VIEW CONCURRENTLY mv_trending_keywords;
        v_duration := CLOCK_TIMESTAMP() - v_start;
        SELECT COUNT(*) INTO v_row_count FROM mv_trending_keywords;
        RAISE NOTICE '✓ Refreshed mv_trending_keywords: % rows in %', v_row_count, v_duration;
    ELSE
        RAISE NOTICE '⚠️  mv_trending_keywords does not exist - run V003 migration';
    END IF;
    
    IF EXISTS (SELECT FROM pg_matviews WHERE matviewname = 'mv_sentiment_timeline') THEN
        v_start := CLOCK_TIMESTAMP();
        REFRESH MATERIALIZED VIEW CONCURRENTLY mv_sentiment_timeline;
        v_duration := CLOCK_TIMESTAMP() - v_start;
        SELECT COUNT(*) INTO v_row_count FROM mv_sentiment_timeline;
        RAISE NOTICE '✓ Refreshed mv_sentiment_timeline: % rows in %', v_row_count, v_duration;
    END IF;
    
    IF EXISTS (SELECT FROM pg_matviews WHERE matviewname = 'mv_entity_mentions') THEN
        v_start := CLOCK_TIMESTAMP();
        REFRESH MATERIALIZED VIEW CONCURRENTLY mv_entity_mentions;
        v_duration := CLOCK_TIMESTAMP() - v_start;
        SELECT COUNT(*) INTO v_row_count FROM mv_entity_mentions;
        RAISE NOTICE '✓ Refreshed mv_entity_mentions: % rows in %', v_row_count, v_duration;
    END IF;
END $$;

\echo ''

-- ============================================================================
-- TASK 2: CLEAN UP OLD EMAILS
-- ============================================================================

\echo '2. CLEANING UP OLD EMAILS'
\echo '-------------------------'

DO $$
DECLARE
    v_deleted BIGINT;
    v_space_freed BIGINT;
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'emails') THEN
        IF EXISTS (SELECT FROM information_schema.routines WHERE routine_name = 'cleanup_old_emails') THEN
            SELECT emails_deleted, space_freed_bytes INTO v_deleted, v_space_freed
            FROM cleanup_old_emails(90, TRUE);
            
            RAISE NOTICE '✓ Deleted % old emails, freed %', v_deleted, pg_size_pretty(v_space_freed);
        ELSE
            -- Fallback if function doesn't exist
            WITH deleted AS (
                DELETE FROM emails
                WHERE received_date < CURRENT_DATE - INTERVAL '90 days'
                  AND status IN ('processed', 'ignored', 'spam')
                  AND article_id IS NULL
                RETURNING id
            )
            SELECT COUNT(*) INTO v_deleted FROM deleted;
            
            RAISE NOTICE '✓ Deleted % old emails (function not available, using fallback)', v_deleted;
        END IF;
    ELSE
        RAISE NOTICE '⚠️  Emails table does not exist';
    END IF;
END $$;

\echo ''

-- ============================================================================
-- TASK 3: CLEAN UP OLD SCRAPING JOBS
-- ============================================================================

\echo '3. CLEANING UP OLD SCRAPING JOBS'
\echo '---------------------------------'

DO $$
DECLARE
    v_deleted BIGINT;
BEGIN
    WITH deleted AS (
        DELETE FROM scraping_jobs
        WHERE created_at < CURRENT_DATE - INTERVAL '30 days'
          AND status IN ('completed', 'failed')
        RETURNING id
    )
    SELECT COUNT(*) INTO v_deleted FROM deleted;
    
    RAISE NOTICE '✓ Deleted % old scraping jobs (>30 days)', v_deleted;
END $$;

\echo ''

-- ============================================================================
-- TASK 4: VACUUM AND ANALYZE
-- ============================================================================

\echo '4. RUNNING VACUUM ANALYZE'
\echo '-------------------------'

-- Articles table
VACUUM ANALYZE articles;
\echo '✓ Vacuumed and analyzed: articles'

-- Sources table
VACUUM ANALYZE sources;
\echo '✓ Vacuumed and analyzed: sources'

-- Scraping jobs table
VACUUM ANALYZE scraping_jobs;
\echo '✓ Vacuumed and analyzed: scraping_jobs'

-- Emails table (if exists)
DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'emails') THEN
        EXECUTE 'VACUUM ANALYZE emails';
        RAISE NOTICE '✓ Vacuumed and analyzed: emails';
    END IF;
END $$;

-- Materialized views
DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_matviews WHERE matviewname = 'mv_trending_keywords') THEN
        EXECUTE 'VACUUM ANALYZE mv_trending_keywords';
        RAISE NOTICE '✓ Vacuumed and analyzed: mv_trending_keywords';
    END IF;
    
    IF EXISTS (SELECT FROM pg_matviews WHERE matviewname = 'mv_sentiment_timeline') THEN
        EXECUTE 'VACUUM ANALYZE mv_sentiment_timeline';
        RAISE NOTICE '✓ Vacuumed and analyzed: mv_sentiment_timeline';
    END IF;
    
    IF EXISTS (SELECT FROM pg_matviews WHERE matviewname = 'mv_entity_mentions') THEN
        EXECUTE 'VACUUM ANALYZE mv_entity_mentions';
        RAISE NOTICE '✓ Vacuumed and analyzed: mv_entity_mentions';
    END IF;
END $$;

\echo ''

-- ============================================================================
-- TASK 5: REINDEX TABLES
-- ============================================================================

\echo '5. REINDEXING TABLES (if needed)'
\echo '--------------------------------'

DO $$
DECLARE
    v_bloat_pct NUMERIC;
BEGIN
    -- Check articles table bloat
    SELECT ROUND(100 * (pg_total_relation_size('articles') - pg_relation_size('articles'))::NUMERIC / 
        NULLIF(pg_total_relation_size('articles'), 0), 2)
    INTO v_bloat_pct;
    
    IF v_bloat_pct > 50 THEN
        RAISE NOTICE '⚠️  Articles table has %.2f%% index bloat - reindexing...', v_bloat_pct;
        EXECUTE 'REINDEX TABLE CONCURRENTLY articles';
        RAISE NOTICE '✓ Reindexed: articles';
    ELSE
        RAISE NOTICE '✓ Articles table bloat is acceptable (%.2f%%)', v_bloat_pct;
    END IF;
    
    -- Check emails table bloat (if exists)
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'emails') THEN
        SELECT ROUND(100 * (pg_total_relation_size('emails') - pg_relation_size('emails'))::NUMERIC / 
            NULLIF(pg_total_relation_size('emails'), 0), 2)
        INTO v_bloat_pct;
        
        IF v_bloat_pct > 50 THEN
            RAISE NOTICE '⚠️  Emails table has %.2f%% index bloat - reindexing...', v_bloat_pct;
            EXECUTE 'REINDEX TABLE CONCURRENTLY emails';
            RAISE NOTICE '✓ Reindexed: emails';
        ELSE
            RAISE NOTICE '✓ Emails table bloat is acceptable (%.2f%%)', v_bloat_pct;
        END IF;
    END IF;
END $$;

\echo ''

-- ============================================================================
-- TASK 6: UPDATE STATISTICS
-- ============================================================================

\echo '6. UPDATING TABLE STATISTICS'
\echo '----------------------------'

ANALYZE articles;
\echo '✓ Updated statistics: articles'

ANALYZE sources;
\echo '✓ Updated statistics: sources'

ANALYZE scraping_jobs;
\echo '✓ Updated statistics: scraping_jobs'

DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'emails') THEN
        EXECUTE 'ANALYZE emails';
        RAISE NOTICE '✓ Updated statistics: emails';
    END IF;
END $$;

\echo ''

-- ============================================================================
-- MAINTENANCE SUMMARY
-- ============================================================================

\echo '7. MAINTENANCE SUMMARY'
\echo '----------------------'

DO $$
DECLARE
    v_total_size TEXT;
    v_articles_count BIGINT;
    v_mv_count INTEGER;
    v_index_count INTEGER;
BEGIN
    -- Total database size
    SELECT pg_size_pretty(pg_database_size(current_database())) INTO v_total_size;
    
    -- Article count
    SELECT COUNT(*) INTO v_articles_count FROM articles;
    
    -- Materialized view count
    SELECT COUNT(*) INTO v_mv_count FROM pg_matviews WHERE schemaname = 'public';
    
    -- Index count
    SELECT COUNT(*) INTO v_index_count FROM pg_indexes WHERE schemaname = 'public';
    
    RAISE NOTICE '';
    RAISE NOTICE 'Database Statistics:';
    RAISE NOTICE '  - Total Size: %', v_total_size;
    RAISE NOTICE '  - Articles: %', v_articles_count;
    RAISE NOTICE '  - Materialized Views: %', v_mv_count;
    RAISE NOTICE '  - Indexes: %', v_index_count;
    RAISE NOTICE '';
    RAISE NOTICE 'Maintenance Recommendations:';
    RAISE NOTICE '  - Run this script weekly for optimal performance';
    RAISE NOTICE '  - Monitor table bloat with health check script';
    RAISE NOTICE '  - Refresh materialized views every 5-15 minutes';
    RAISE NOTICE '  - Consider pg_cron for automated maintenance';
END $$;

\echo ''
\echo '========================================'
\echo '✅ MAINTENANCE COMPLETE'
\echo '========================================'

-- ============================================================================
-- ADDITIONAL MAINTENANCE FUNCTIONS
-- ============================================================================

-- Function to get maintenance schedule recommendations
CREATE OR REPLACE FUNCTION get_maintenance_schedule()
RETURNS TABLE (
    task TEXT,
    frequency TEXT,
    last_run TIMESTAMPTZ,
    next_recommended TIMESTAMPTZ,
    status TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        'Refresh Materialized Views'::TEXT,
        'Every 5-15 minutes'::TEXT,
        (SELECT MAX(last_refresh) FROM pg_matviews WHERE schemaname = 'public'),
        (SELECT MAX(last_refresh) + INTERVAL '15 minutes' FROM pg_matviews WHERE schemaname = 'public'),
        CASE 
            WHEN (SELECT MAX(last_refresh) FROM pg_matviews WHERE schemaname = 'public') < CURRENT_TIMESTAMP - INTERVAL '15 minutes' THEN '⚠️  OVERDUE'
            ELSE '✓ ON SCHEDULE'
        END
    UNION ALL
    SELECT 
        'Clean Old Emails'::TEXT,
        'Weekly'::TEXT,
        NULL::TIMESTAMPTZ,
        NULL::TIMESTAMPTZ,
        '⚠️  RUN MANUALLY'::TEXT
    UNION ALL
    SELECT 
        'Vacuum Analyze'::TEXT,
        'Daily'::TEXT,
        (SELECT MAX(last_autovacuum) FROM pg_stat_user_tables WHERE schemaname = 'public'),
        (SELECT MAX(last_autovacuum) + INTERVAL '1 day' FROM pg_stat_user_tables WHERE schemaname = 'public'),
        CASE 
            WHEN (SELECT MAX(last_autovacuum) FROM pg_stat_user_tables WHERE schemaname = 'public') < CURRENT_TIMESTAMP - INTERVAL '1 day' THEN '⚠️  OVERDUE'
            ELSE '✓ ON SCHEDULE'
        END;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_maintenance_schedule IS 'Returns recommended maintenance schedule and status';

\echo ''
\echo 'Created maintenance helper function: get_maintenance_schedule()'
\echo 'Usage: SELECT * FROM get_maintenance_schedule();'