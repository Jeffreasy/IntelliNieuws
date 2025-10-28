-- Migration: Create materialized view for trending keywords
-- This optimizes the expensive trending topics query by pre-aggregating data
-- Expected performance improvement: 90% faster (5s → 0.5s)

-- Create materialized view for trending keywords
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_trending_keywords AS
SELECT 
    kw->>'word' as keyword,
    DATE_TRUNC('hour', a.published) as hour_bucket,
    COUNT(DISTINCT a.id) as article_count,
    AVG(a.ai_sentiment) as avg_sentiment,
    ARRAY_AGG(DISTINCT a.source) as sources,
    MAX(a.published) as latest_article
FROM articles a,
     LATERAL jsonb_array_elements(a.ai_keywords) as kw
WHERE a.ai_processed = TRUE
  AND a.ai_keywords IS NOT NULL
  AND a.ai_keywords != 'null'::jsonb
  AND jsonb_typeof(a.ai_keywords) = 'array'
GROUP BY kw->>'word', DATE_TRUNC('hour', a.published);

-- Create indexes on materialized view for fast queries
CREATE INDEX IF NOT EXISTS idx_mv_trending_hour 
    ON mv_trending_keywords(hour_bucket DESC);

CREATE INDEX IF NOT EXISTS idx_mv_trending_keyword 
    ON mv_trending_keywords(keyword);

CREATE INDEX IF NOT EXISTS idx_mv_trending_count 
    ON mv_trending_keywords(article_count DESC);

-- Create index for efficient refresh
CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_trending_unique
    ON mv_trending_keywords(keyword, hour_bucket);

-- Grant permissions (adjust as needed)
-- GRANT SELECT ON mv_trending_keywords TO your_app_user;

-- Refresh the view initially
REFRESH MATERIALIZED VIEW CONCURRENTLY mv_trending_keywords;

-- Notes:
-- 1. Refresh this view periodically (every 5-15 minutes) for fresh data
-- 2. Use REFRESH MATERIALIZED VIEW CONCURRENTLY to avoid blocking queries
-- 3. The UNIQUE index enables CONCURRENTLY refresh
-- 4. Old data (>7 days) can be cleaned up periodically if needed

-- Example refresh command (run periodically):
-- REFRESH MATERIALIZED VIEW CONCURRENTLY mv_trending_keywords;

-- Example cleanup command (optional, for old data):
-- DELETE FROM mv_trending_keywords WHERE hour_bucket < NOW() - INTERVAL '7 days';
-- REFRESH MATERIALIZED VIEW CONCURRENTLY mv_trending_keywords;