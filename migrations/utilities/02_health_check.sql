-- ============================================================================
-- Utility Script: Database Health Check
-- Description: Comprehensive database health and integrity checks
-- Version: 1.0.0
-- Author: NieuwsScraper Team
-- Date: 2025-10-30
-- ============================================================================

\echo '========================================'
\echo 'DATABASE HEALTH CHECK'
\echo '========================================'
\echo ''

-- ============================================================================
-- 1. MIGRATION STATUS
-- ============================================================================

\echo '1. MIGRATION STATUS'
\echo '-------------------'

SELECT 
    version,
    description,
    applied_at,
    applied_by
FROM schema_migrations 
ORDER BY version;

\echo ''

-- ============================================================================
-- 2. TABLE SIZES
-- ============================================================================

\echo '2. TABLE SIZES'
\echo '--------------'

SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS total_size,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS table_size,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) AS index_size,
    pg_total_relation_size(schemaname||'.'||tablename) AS bytes
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY bytes DESC;

\echo ''

-- ============================================================================
-- 3. ROW COUNTS
-- ============================================================================

\echo '3. TABLE ROW COUNTS'
\echo '-------------------'

SELECT 
    'articles' AS table_name,
    COUNT(*) AS row_count,
    COUNT(*) FILTER (WHERE created_at >= CURRENT_DATE - INTERVAL '24 hours') AS rows_today,
    COUNT(*) FILTER (WHERE created_at >= CURRENT_DATE - INTERVAL '7 days') AS rows_week
FROM articles
UNION ALL
SELECT 
    'sources',
    COUNT(*),
    NULL,
    NULL
FROM sources
UNION ALL
SELECT 
    'scraping_jobs',
    COUNT(*),
    COUNT(*) FILTER (WHERE created_at >= CURRENT_DATE - INTERVAL '24 hours'),
    COUNT(*) FILTER (WHERE created_at >= CURRENT_DATE - INTERVAL '7 days')
FROM scraping_jobs
UNION ALL
SELECT 
    'emails',
    COUNT(*),
    COUNT(*) FILTER (WHERE created_at >= CURRENT_DATE - INTERVAL '24 hours'),
    COUNT(*) FILTER (WHERE created_at >= CURRENT_DATE - INTERVAL '7 days')
FROM emails
WHERE EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'emails');

\echo ''

-- ============================================================================
-- 4. INDEX HEALTH
-- ============================================================================

\echo '4. INDEX USAGE STATISTICS'
\echo '-------------------------'

SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan AS scans,
    idx_tup_read AS tuples_read,
    idx_tup_fetch AS tuples_fetched,
    pg_size_pretty(pg_relation_size(indexrelid)) AS index_size,
    CASE 
        WHEN idx_scan = 0 THEN '⚠️  UNUSED'
        WHEN idx_scan < 100 THEN '⚠️  LOW USAGE'
        ELSE '✓ ACTIVE'
    END AS status
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan DESC, pg_relation_size(indexrelid) DESC
LIMIT 20;

\echo ''

-- ============================================================================
-- 5. BLOAT CHECK
-- ============================================================================

\echo '5. TABLE BLOAT ESTIMATE'
\echo '-----------------------'

SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS total_size,
    ROUND(100 * (pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename))::NUMERIC / 
        NULLIF(pg_total_relation_size(schemaname||'.'||tablename), 0), 2) AS index_bloat_pct,
    CASE 
        WHEN ROUND(100 * (pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename))::NUMERIC / 
            NULLIF(pg_total_relation_size(schemaname||'.'||tablename), 0), 2) > 50 THEN '⚠️  HIGH'
        WHEN ROUND(100 * (pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename))::NUMERIC / 
            NULLIF(pg_total_relation_size(schemaname||'.'||tablename), 0), 2) > 30 THEN '⚠️  MODERATE'
        ELSE '✓ GOOD'
    END AS status
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

\echo ''

-- ============================================================================
-- 6. PROCESSING STATUS
-- ============================================================================

\echo '6. ARTICLE PROCESSING STATUS'
\echo '----------------------------'

SELECT 
    COUNT(*) AS total_articles,
    COUNT(*) FILTER (WHERE content_extracted = TRUE) AS content_extracted,
    COUNT(*) FILTER (WHERE content_extracted = FALSE OR content_extracted IS NULL) AS needs_content,
    COUNT(*) FILTER (WHERE ai_processed = TRUE) AS ai_processed,
    COUNT(*) FILTER (WHERE ai_processed = FALSE) AS needs_ai_processing,
    ROUND(100.0 * COUNT(*) FILTER (WHERE content_extracted = TRUE) / NULLIF(COUNT(*), 0), 2) AS content_completion_pct,
    ROUND(100.0 * COUNT(*) FILTER (WHERE ai_processed = TRUE) / NULLIF(COUNT(*), 0), 2) AS ai_completion_pct
FROM articles;

\echo ''

-- ============================================================================
-- 7. SENTIMENT DISTRIBUTION
-- ============================================================================

\echo '7. SENTIMENT ANALYSIS DISTRIBUTION'
\echo '-----------------------------------'

SELECT 
    ai_sentiment_label,
    COUNT(*) AS count,
    ROUND(AVG(ai_sentiment)::NUMERIC, 3) AS avg_score,
    ROUND(100.0 * COUNT(*) / SUM(COUNT(*)) OVER (), 2) AS percentage
FROM articles
WHERE ai_sentiment_label IS NOT NULL
GROUP BY ai_sentiment_label
ORDER BY count DESC;

\echo ''

-- ============================================================================
-- 8. SOURCE ACTIVITY
-- ============================================================================

\echo '8. SOURCE ACTIVITY (LAST 7 DAYS)'
\echo '---------------------------------'

SELECT 
    s.name,
    s.domain,
    s.is_active,
    COUNT(a.id) AS articles_count,
    COUNT(a.id) FILTER (WHERE a.created_at >= CURRENT_DATE - INTERVAL '24 hours') AS articles_today,
    MAX(a.created_at) AS last_article,
    s.last_scraped_at,
    s.consecutive_failures
FROM sources s
LEFT JOIN articles a ON a.source = s.domain AND a.created_at >= CURRENT_DATE - INTERVAL '7 days'
GROUP BY s.id, s.name, s.domain, s.is_active, s.last_scraped_at, s.consecutive_failures
ORDER BY articles_count DESC;

\echo ''

-- ============================================================================
-- 9. SCRAPING JOB STATUS
-- ============================================================================

\echo '9. SCRAPING JOB STATUS (LAST 24 HOURS)'
\echo '---------------------------------------'

SELECT 
    status,
    COUNT(*) AS count,
    ROUND(AVG(execution_time_ms)::NUMERIC, 0) AS avg_execution_ms,
    SUM(articles_new) AS total_new_articles,
    SUM(articles_updated) AS total_updated_articles
FROM scraping_jobs
WHERE created_at >= CURRENT_DATE - INTERVAL '24 hours'
GROUP BY status
ORDER BY 
    CASE status
        WHEN 'completed' THEN 1
        WHEN 'running' THEN 2
        WHEN 'pending' THEN 3
        WHEN 'failed' THEN 4
        ELSE 5
    END;

\echo ''

-- ============================================================================
-- 10. EMAIL PROCESSING STATUS
-- ============================================================================

\echo '10. EMAIL PROCESSING STATUS'
\echo '---------------------------'

DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'emails') THEN
        PERFORM 1;
    ELSE
        RAISE NOTICE 'Emails table does not exist';
    END IF;
END $$;

SELECT 
    status,
    COUNT(*) AS count,
    COUNT(*) FILTER (WHERE article_created = TRUE) AS articles_created,
    ROUND(100.0 * COUNT(*) FILTER (WHERE article_created = TRUE) / NULLIF(COUNT(*), 0), 2) AS conversion_rate_pct
FROM emails
WHERE EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'emails')
GROUP BY status
ORDER BY count DESC;

\echo ''

-- ============================================================================
-- 11. MATERIALIZED VIEW STATUS
-- ============================================================================

\echo '11. MATERIALIZED VIEW STATUS'
\echo '----------------------------'

SELECT 
    schemaname,
    matviewname,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||matviewname)) AS size,
    CASE 
        WHEN last_refresh IS NULL THEN '⚠️  NEVER REFRESHED'
        WHEN last_refresh < CURRENT_TIMESTAMP - INTERVAL '1 hour' THEN '⚠️  STALE (>1h)'
        ELSE '✓ FRESH'
    END AS status,
    last_refresh
FROM pg_matviews
WHERE schemaname = 'public'
ORDER BY matviewname;

\echo ''

-- ============================================================================
-- 12. CONNECTION STATS
-- ============================================================================

\echo '12. DATABASE CONNECTIONS'
\echo '------------------------'

SELECT 
    datname,
    COUNT(*) AS connections,
    COUNT(*) FILTER (WHERE state = 'active') AS active,
    COUNT(*) FILTER (WHERE state = 'idle') AS idle,
    COUNT(*) FILTER (WHERE state = 'idle in transaction') AS idle_in_transaction
FROM pg_stat_activity
WHERE datname = current_database()
GROUP BY datname;

\echo ''

-- ============================================================================
-- 13. LONG RUNNING QUERIES
-- ============================================================================

\echo '13. LONG RUNNING QUERIES (>30s)'
\echo '--------------------------------'

SELECT 
    pid,
    now() - pg_stat_activity.query_start AS duration,
    state,
    LEFT(query, 100) AS query_preview
FROM pg_stat_activity
WHERE (now() - pg_stat_activity.query_start) > interval '30 seconds'
  AND state != 'idle'
  AND datname = current_database()
ORDER BY duration DESC;

\echo ''

-- ============================================================================
-- 14. CACHE HIT RATIO
-- ============================================================================

\echo '14. CACHE HIT RATIO'
\echo '-------------------'

SELECT 
    'Table Cache' AS cache_type,
    SUM(heap_blks_read) AS disk_reads,
    SUM(heap_blks_hit) AS cache_hits,
    ROUND(100.0 * SUM(heap_blks_hit) / NULLIF(SUM(heap_blks_hit) + SUM(heap_blks_read), 0), 2) AS hit_ratio_pct,
    CASE 
        WHEN ROUND(100.0 * SUM(heap_blks_hit) / NULLIF(SUM(heap_blks_hit) + SUM(heap_blks_read), 0), 2) > 99 THEN '✓ EXCELLENT'
        WHEN ROUND(100.0 * SUM(heap_blks_hit) / NULLIF(SUM(heap_blks_hit) + SUM(heap_blks_read), 0), 2) > 95 THEN '✓ GOOD'
        WHEN ROUND(100.0 * SUM(heap_blks_hit) / NULLIF(SUM(heap_blks_hit) + SUM(heap_blks_read), 0), 2) > 90 THEN '⚠️  FAIR'
        ELSE '⚠️  POOR'
    END AS status
FROM pg_statio_user_tables
UNION ALL
SELECT 
    'Index Cache',
    SUM(idx_blks_read),
    SUM(idx_blks_hit),
    ROUND(100.0 * SUM(idx_blks_hit) / NULLIF(SUM(idx_blks_hit) + SUM(idx_blks_read), 0), 2),
    CASE 
        WHEN ROUND(100.0 * SUM(idx_blks_hit) / NULLIF(SUM(idx_blks_hit) + SUM(idx_blks_read), 0), 2) > 99 THEN '✓ EXCELLENT'
        WHEN ROUND(100.0 * SUM(idx_blks_hit) / NULLIF(SUM(idx_blks_hit) + SUM(idx_blks_read), 0), 2) > 95 THEN '✓ GOOD'
        WHEN ROUND(100.0 * SUM(idx_blks_hit) / NULLIF(SUM(idx_blks_hit) + SUM(idx_blks_read), 0), 2) > 90 THEN '⚠️  FAIR'
        ELSE '⚠️  POOR'
    END
FROM pg_statio_user_indexes;

\echo ''

-- ============================================================================
-- 15. RECOMMENDATIONS
-- ============================================================================

\echo '15. HEALTH CHECK RECOMMENDATIONS'
\echo '---------------------------------'

DO $$
DECLARE
    v_unused_indexes INTEGER;
    v_stale_views INTEGER;
    v_pending_articles INTEGER;
    v_failed_jobs INTEGER;
BEGIN
    -- Check for unused indexes
    SELECT COUNT(*) INTO v_unused_indexes
    FROM pg_stat_user_indexes
    WHERE schemaname = 'public' AND idx_scan = 0;
    
    IF v_unused_indexes > 0 THEN
        RAISE NOTICE '⚠️  Found % unused indexes - consider dropping them', v_unused_indexes;
    END IF;
    
    -- Check for stale materialized views
    SELECT COUNT(*) INTO v_stale_views
    FROM pg_matviews
    WHERE schemaname = 'public' 
      AND (last_refresh IS NULL OR last_refresh < CURRENT_TIMESTAMP - INTERVAL '1 hour');
    
    IF v_stale_views > 0 THEN
        RAISE NOTICE '⚠️  Found % stale materialized views - run: SELECT refresh_analytics_views(TRUE);', v_stale_views;
    END IF;
    
    -- Check for articles pending processing
    SELECT COUNT(*) INTO v_pending_articles
    FROM articles
    WHERE ai_processed = FALSE OR content_extracted = FALSE;
    
    IF v_pending_articles > 100 THEN
        RAISE NOTICE '⚠️  Found % articles pending processing - check background workers', v_pending_articles;
    END IF;
    
    -- Check for recent failed jobs
    SELECT COUNT(*) INTO v_failed_jobs
    FROM scraping_jobs
    WHERE status = 'failed' 
      AND created_at >= CURRENT_TIMESTAMP - INTERVAL '24 hours';
    
    IF v_failed_jobs > 0 THEN
        RAISE NOTICE '⚠️  Found % failed scraping jobs in last 24h - check error logs', v_failed_jobs;
    END IF;
    
    -- General recommendations
    RAISE NOTICE '';
    RAISE NOTICE '✓ Run VACUUM ANALYZE regularly for optimal performance';
    RAISE NOTICE '✓ Monitor table bloat and run VACUUM FULL if needed';
    RAISE NOTICE '✓ Refresh materialized views every 5-15 minutes';
    RAISE NOTICE '✓ Clean up old processed emails periodically';
END $$;

\echo ''
\echo '========================================'
\echo 'HEALTH CHECK COMPLETE'
\echo '========================================'