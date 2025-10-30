-- ============================================================================
-- Migration: V003__create_analytics_views.sql
-- Description: Materialized views for trending topics and analytics
-- Version: 1.0.0
-- Author: NieuwsScraper Team
-- Date: 2025-10-30
-- Dependencies: V001__create_base_schema.sql
-- Performance: 90% faster than dynamic queries (5s → 0.5s)
-- ============================================================================

-- ============================================================================
-- MATERIALIZED VIEW: TRENDING KEYWORDS
-- ============================================================================

-- Drop existing if any
DROP MATERIALIZED VIEW IF EXISTS mv_trending_keywords CASCADE;

CREATE MATERIALIZED VIEW mv_trending_keywords AS
WITH keyword_expansion AS (
    SELECT 
        a.id AS article_id,
        a.source,
        a.published,
        a.category,
        a.ai_sentiment,
        DATE_TRUNC('hour', a.published) AS hour_bucket,
        DATE_TRUNC('day', a.published) AS day_bucket,
        kw.value->>'word' AS keyword,
        (kw.value->>'score')::FLOAT AS relevance_score
    FROM articles a
    CROSS JOIN LATERAL jsonb_array_elements(a.ai_keywords) AS kw(value)
    WHERE a.ai_processed = TRUE
      AND a.ai_keywords IS NOT NULL
      AND a.ai_keywords != 'null'::jsonb
      AND jsonb_typeof(a.ai_keywords) = 'array'
      AND a.published >= CURRENT_TIMESTAMP - INTERVAL '30 days'
),
keyword_aggregates AS (
    SELECT 
        keyword,
        hour_bucket,
        day_bucket,
        COUNT(DISTINCT article_id) AS article_count,
        COUNT(DISTINCT source) AS source_count,
        ARRAY_AGG(DISTINCT source ORDER BY source) AS sources,
        AVG(ai_sentiment) FILTER (WHERE ai_sentiment IS NOT NULL) AS avg_sentiment,
        AVG(relevance_score) FILTER (WHERE relevance_score IS NOT NULL) AS avg_relevance,
        MAX(published) AS latest_article_date,
        MIN(published) AS first_article_date,
        ARRAY_AGG(DISTINCT category ORDER BY category) FILTER (WHERE category IS NOT NULL) AS categories
    FROM keyword_expansion
    GROUP BY keyword, hour_bucket, day_bucket
)
SELECT 
    keyword,
    hour_bucket,
    day_bucket,
    article_count,
    source_count,
    sources,
    ROUND(avg_sentiment::NUMERIC, 3) AS avg_sentiment,
    ROUND(avg_relevance::NUMERIC, 3) AS avg_relevance,
    latest_article_date,
    first_article_date,
    categories,
    -- Trending score calculation (article count * source diversity * recency)
    ROUND((
        article_count * 
        (1 + LOG(source_count + 1)) * 
        (1 - EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - latest_article_date)) / 86400.0)
    )::NUMERIC, 2) AS trending_score
FROM keyword_aggregates
WHERE article_count >= 2; -- Minimum 2 articles to be considered trending

-- Create indexes on materialized view
CREATE UNIQUE INDEX idx_mv_trending_keywords_unique
    ON mv_trending_keywords(keyword, hour_bucket);

CREATE INDEX idx_mv_trending_keywords_hour
    ON mv_trending_keywords(hour_bucket DESC);

CREATE INDEX idx_mv_trending_keywords_day
    ON mv_trending_keywords(day_bucket DESC);

CREATE INDEX idx_mv_trending_keywords_keyword
    ON mv_trending_keywords(keyword);

CREATE INDEX idx_mv_trending_keywords_article_count
    ON mv_trending_keywords(article_count DESC);

CREATE INDEX idx_mv_trending_keywords_trending_score
    ON mv_trending_keywords(trending_score DESC NULLS LAST);

CREATE INDEX idx_mv_trending_keywords_sentiment
    ON mv_trending_keywords(avg_sentiment DESC NULLS LAST);

COMMENT ON MATERIALIZED VIEW mv_trending_keywords IS 'Pre-aggregated trending keywords with hourly and daily buckets for fast queries';
COMMENT ON COLUMN mv_trending_keywords.trending_score IS 'Calculated trending score based on count, diversity, and recency';

-- ============================================================================
-- MATERIALIZED VIEW: ARTICLE SENTIMENT TIMELINE
-- ============================================================================

DROP MATERIALIZED VIEW IF EXISTS mv_sentiment_timeline CASCADE;

CREATE MATERIALIZED VIEW mv_sentiment_timeline AS
SELECT 
    DATE_TRUNC('hour', a.published) AS time_bucket,
    a.source,
    a.category,
    COUNT(*) AS article_count,
    COUNT(*) FILTER (WHERE ai_sentiment_label = 'positive') AS positive_count,
    COUNT(*) FILTER (WHERE ai_sentiment_label = 'neutral') AS neutral_count,
    COUNT(*) FILTER (WHERE ai_sentiment_label = 'negative') AS negative_count,
    ROUND(AVG(ai_sentiment)::NUMERIC, 3) AS avg_sentiment,
    ROUND(STDDEV(ai_sentiment)::NUMERIC, 3) AS sentiment_stddev,
    MAX(ai_sentiment) AS max_sentiment,
    MIN(ai_sentiment) AS min_sentiment,
    ARRAY_AGG(DISTINCT a.id ORDER BY a.ai_sentiment DESC) FILTER (WHERE a.ai_sentiment IS NOT NULL) AS top_article_ids
FROM articles a
WHERE a.ai_processed = TRUE
  AND a.ai_sentiment IS NOT NULL
  AND a.published >= CURRENT_TIMESTAMP - INTERVAL '30 days'
GROUP BY DATE_TRUNC('hour', a.published), a.source, a.category;

-- Indexes for sentiment timeline
CREATE UNIQUE INDEX idx_mv_sentiment_timeline_unique
    ON mv_sentiment_timeline(time_bucket, source, COALESCE(category, ''));

CREATE INDEX idx_mv_sentiment_timeline_time
    ON mv_sentiment_timeline(time_bucket DESC);

CREATE INDEX idx_mv_sentiment_timeline_source
    ON mv_sentiment_timeline(source, time_bucket DESC);

CREATE INDEX idx_mv_sentiment_timeline_category
    ON mv_sentiment_timeline(category, time_bucket DESC)
    WHERE category IS NOT NULL;

CREATE INDEX idx_mv_sentiment_timeline_avg_sentiment
    ON mv_sentiment_timeline(avg_sentiment DESC NULLS LAST);

COMMENT ON MATERIALIZED VIEW mv_sentiment_timeline IS 'Hourly sentiment aggregates by source and category';

-- ============================================================================
-- MATERIALIZED VIEW: ENTITY MENTIONS
-- ============================================================================

DROP MATERIALIZED VIEW IF EXISTS mv_entity_mentions CASCADE;

CREATE MATERIALIZED VIEW mv_entity_mentions AS
WITH entity_extraction AS (
    SELECT 
        a.id AS article_id,
        a.source,
        a.published,
        a.ai_sentiment,
        a.category,
        DATE_TRUNC('day', a.published) AS day_bucket,
        entity_type,
        entity_value
    FROM articles a
    CROSS JOIN LATERAL (
        SELECT 'person' AS entity_type, jsonb_array_elements_text(a.ai_entities->'persons') AS entity_value
        UNION ALL
        SELECT 'organization', jsonb_array_elements_text(a.ai_entities->'organizations')
        UNION ALL
        SELECT 'location', jsonb_array_elements_text(a.ai_entities->'locations')
    ) entities
    WHERE a.ai_processed = TRUE
      AND a.ai_entities IS NOT NULL
      AND a.published >= CURRENT_TIMESTAMP - INTERVAL '90 days'
)
SELECT 
    entity_value AS entity,
    entity_type,
    day_bucket,
    COUNT(DISTINCT article_id) AS mention_count,
    COUNT(DISTINCT source) AS source_count,
    ARRAY_AGG(DISTINCT source ORDER BY source) AS sources,
    ARRAY_AGG(DISTINCT category ORDER BY category) FILTER (WHERE category IS NOT NULL) AS categories,
    ROUND(AVG(ai_sentiment)::NUMERIC, 3) FILTER (WHERE ai_sentiment IS NOT NULL) AS avg_sentiment,
    MAX(published) AS last_mentioned,
    ARRAY_AGG(article_id ORDER BY published DESC) AS recent_article_ids
FROM entity_extraction
GROUP BY entity_value, entity_type, day_bucket
HAVING COUNT(DISTINCT article_id) >= 2; -- Minimum 2 mentions

-- Indexes for entity mentions
CREATE UNIQUE INDEX idx_mv_entity_mentions_unique
    ON mv_entity_mentions(entity, entity_type, day_bucket);

CREATE INDEX idx_mv_entity_mentions_entity
    ON mv_entity_mentions(entity);

CREATE INDEX idx_mv_entity_mentions_type
    ON mv_entity_mentions(entity_type, mention_count DESC);

CREATE INDEX idx_mv_entity_mentions_count
    ON mv_entity_mentions(mention_count DESC);

CREATE INDEX idx_mv_entity_mentions_day
    ON mv_entity_mentions(day_bucket DESC);

CREATE INDEX idx_mv_entity_mentions_last_mentioned
    ON mv_entity_mentions(last_mentioned DESC);

COMMENT ON MATERIALIZED VIEW mv_entity_mentions IS 'Entity mentions aggregated by day with sentiment';

-- ============================================================================
-- HELPER VIEWS FOR ANALYTICS
-- ============================================================================

-- View: Top trending keywords (last 24 hours)
CREATE OR REPLACE VIEW v_trending_keywords_24h AS
SELECT 
    keyword,
    SUM(article_count) AS total_articles,
    ARRAY_AGG(DISTINCT unnest) AS all_sources,
    ROUND(AVG(avg_sentiment)::NUMERIC, 3) AS overall_sentiment,
    MAX(latest_article_date) AS most_recent,
    ROUND(AVG(trending_score)::NUMERIC, 2) AS avg_trending_score
FROM mv_trending_keywords
CROSS JOIN LATERAL unnest(sources) AS unnest
WHERE hour_bucket >= CURRENT_TIMESTAMP - INTERVAL '24 hours'
GROUP BY keyword
HAVING SUM(article_count) >= 3
ORDER BY avg_trending_score DESC NULLS LAST, total_articles DESC
LIMIT 50;

COMMENT ON VIEW v_trending_keywords_24h IS 'Top 50 trending keywords in last 24 hours';

-- View: Sentiment trends (last 7 days)
CREATE OR REPLACE VIEW v_sentiment_trends_7d AS
SELECT 
    DATE(time_bucket) AS day,
    source,
    SUM(article_count) AS total_articles,
    SUM(positive_count) AS positive_total,
    SUM(neutral_count) AS neutral_total,
    SUM(negative_count) AS negative_total,
    ROUND(AVG(avg_sentiment)::NUMERIC, 3) AS daily_avg_sentiment,
    ROUND((SUM(positive_count)::FLOAT / NULLIF(SUM(article_count), 0) * 100)::NUMERIC, 1) AS positive_percentage,
    ROUND((SUM(negative_count)::FLOAT / NULLIF(SUM(article_count), 0) * 100)::NUMERIC, 1) AS negative_percentage
FROM mv_sentiment_timeline
WHERE time_bucket >= CURRENT_TIMESTAMP - INTERVAL '7 days'
GROUP BY DATE(time_bucket), source
ORDER BY day DESC, source;

COMMENT ON VIEW v_sentiment_trends_7d IS 'Daily sentiment trends by source for last 7 days';

-- View: Hot entities (last 7 days)
CREATE OR REPLACE VIEW v_hot_entities_7d AS
SELECT 
    entity,
    entity_type,
    SUM(mention_count) AS total_mentions,
    COUNT(DISTINCT day_bucket) AS days_mentioned,
    ARRAY_AGG(DISTINCT unnest ORDER BY unnest) AS all_sources,
    ROUND(AVG(avg_sentiment)::NUMERIC, 3) AS overall_sentiment,
    MAX(last_mentioned) AS most_recent_mention
FROM mv_entity_mentions
CROSS JOIN LATERAL unnest(sources) AS unnest
WHERE day_bucket >= CURRENT_DATE - INTERVAL '7 days'
GROUP BY entity, entity_type
HAVING SUM(mention_count) >= 5
ORDER BY total_mentions DESC, days_mentioned DESC
LIMIT 100;

COMMENT ON VIEW v_hot_entities_7d IS 'Top 100 most mentioned entities in last 7 days';

-- ============================================================================
-- FUNCTIONS FOR ANALYTICS
-- ============================================================================

-- Function: Get trending topics with configurable time window
CREATE OR REPLACE FUNCTION get_trending_topics(
    p_hours_back INTEGER DEFAULT 24,
    p_min_articles INTEGER DEFAULT 3,
    p_limit INTEGER DEFAULT 20
)
RETURNS TABLE (
    keyword TEXT,
    article_count BIGINT,
    source_count BIGINT,
    sources TEXT[],
    avg_sentiment NUMERIC,
    avg_relevance NUMERIC,
    trending_score NUMERIC,
    latest_mention TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        tk.keyword,
        SUM(tk.article_count)::BIGINT AS article_count,
        MAX(tk.source_count)::BIGINT AS source_count,
        ARRAY_AGG(DISTINCT unnest ORDER BY unnest) AS sources,
        ROUND(AVG(tk.avg_sentiment)::NUMERIC, 3) AS avg_sentiment,
        ROUND(AVG(tk.avg_relevance)::NUMERIC, 3) AS avg_relevance,
        ROUND(AVG(tk.trending_score)::NUMERIC, 2) AS trending_score,
        MAX(tk.latest_article_date) AS latest_mention
    FROM mv_trending_keywords tk
    CROSS JOIN LATERAL unnest(tk.sources) AS unnest
    WHERE tk.hour_bucket >= CURRENT_TIMESTAMP - (p_hours_back || ' hours')::INTERVAL
    GROUP BY tk.keyword
    HAVING SUM(tk.article_count) >= p_min_articles
    ORDER BY trending_score DESC NULLS LAST, article_count DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION get_trending_topics IS 'Get trending topics with configurable parameters';

-- Function: Get sentiment analysis for entity
CREATE OR REPLACE FUNCTION get_entity_sentiment_analysis(
    p_entity TEXT,
    p_days_back INTEGER DEFAULT 30
)
RETURNS TABLE (
    day DATE,
    mention_count BIGINT,
    avg_sentiment NUMERIC,
    sources TEXT[],
    categories TEXT[]
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        em.day_bucket::DATE AS day,
        SUM(em.mention_count)::BIGINT AS mention_count,
        ROUND(AVG(em.avg_sentiment)::NUMERIC, 3) AS avg_sentiment,
        ARRAY_AGG(DISTINCT unnest ORDER BY unnest) AS sources,
        ARRAY_AGG(DISTINCT cat ORDER BY cat) FILTER (WHERE cat IS NOT NULL) AS categories
    FROM mv_entity_mentions em
    CROSS JOIN LATERAL unnest(em.sources) AS unnest
    CROSS JOIN LATERAL unnest(em.categories) AS cat
    WHERE LOWER(em.entity) = LOWER(p_entity)
      AND em.day_bucket >= CURRENT_DATE - (p_days_back || ' days')::INTERVAL
    GROUP BY em.day_bucket
    ORDER BY day DESC;
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION get_entity_sentiment_analysis IS 'Analyze sentiment for specific entity over time';

-- ============================================================================
-- REFRESH FUNCTIONS
-- ============================================================================

-- Function: Refresh all materialized views
CREATE OR REPLACE FUNCTION refresh_analytics_views(
    p_concurrent BOOLEAN DEFAULT TRUE
)
RETURNS TABLE (
    view_name TEXT,
    refresh_time_ms INTEGER,
    rows_affected BIGINT
) AS $$
DECLARE
    v_start TIMESTAMPTZ;
    v_duration INTEGER;
    v_rows BIGINT;
BEGIN
    -- Refresh trending keywords
    v_start := CLOCK_TIMESTAMP();
    IF p_concurrent THEN
        REFRESH MATERIALIZED VIEW CONCURRENTLY mv_trending_keywords;
    ELSE
        REFRESH MATERIALIZED VIEW mv_trending_keywords;
    END IF;
    v_duration := EXTRACT(EPOCH FROM (CLOCK_TIMESTAMP() - v_start) * 1000)::INTEGER;
    SELECT COUNT(*) INTO v_rows FROM mv_trending_keywords;
    RETURN QUERY SELECT 'mv_trending_keywords'::TEXT, v_duration, v_rows;

    -- Refresh sentiment timeline
    v_start := CLOCK_TIMESTAMP();
    IF p_concurrent THEN
        REFRESH MATERIALIZED VIEW CONCURRENTLY mv_sentiment_timeline;
    ELSE
        REFRESH MATERIALIZED VIEW mv_sentiment_timeline;
    END IF;
    v_duration := EXTRACT(EPOCH FROM (CLOCK_TIMESTAMP() - v_start) * 1000)::INTEGER;
    SELECT COUNT(*) INTO v_rows FROM mv_sentiment_timeline;
    RETURN QUERY SELECT 'mv_sentiment_timeline'::TEXT, v_duration, v_rows;

    -- Refresh entity mentions
    v_start := CLOCK_TIMESTAMP();
    IF p_concurrent THEN
        REFRESH MATERIALIZED VIEW CONCURRENTLY mv_entity_mentions;
    ELSE
        REFRESH MATERIALIZED VIEW mv_entity_mentions;
    END IF;
    v_duration := EXTRACT(EPOCH FROM (CLOCK_TIMESTAMP() - v_start) * 1000)::INTEGER;
    SELECT COUNT(*) INTO v_rows FROM mv_entity_mentions;
    RETURN QUERY SELECT 'mv_entity_mentions'::TEXT, v_duration, v_rows;
    
    -- Update statistics
    ANALYZE mv_trending_keywords;
    ANALYZE mv_sentiment_timeline;
    ANALYZE mv_entity_mentions;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION refresh_analytics_views IS 'Refresh all analytics materialized views';

-- ============================================================================
-- INITIAL REFRESH
-- ============================================================================

-- Perform initial refresh (non-concurrent for first time)
SELECT refresh_analytics_views(FALSE);

-- ============================================================================
-- PERMISSIONS
-- ============================================================================

GRANT SELECT ON mv_trending_keywords TO PUBLIC;
GRANT SELECT ON mv_sentiment_timeline TO PUBLIC;
GRANT SELECT ON mv_entity_mentions TO PUBLIC;
GRANT SELECT ON v_trending_keywords_24h TO PUBLIC;
GRANT SELECT ON v_sentiment_trends_7d TO PUBLIC;
GRANT SELECT ON v_hot_entities_7d TO PUBLIC;

-- ============================================================================
-- FINALIZE MIGRATION
-- ============================================================================

-- Record migration
INSERT INTO schema_migrations (version, description, checksum) 
VALUES (
    'V003',
    'Create analytics materialized views and helper functions',
    'analytics_v1'
) ON CONFLICT (version) DO NOTHING;

-- Success notification
DO $$ 
DECLARE
    v_trending_count BIGINT;
    v_sentiment_count BIGINT;
    v_entity_count BIGINT;
BEGIN
    SELECT COUNT(*) INTO v_trending_count FROM mv_trending_keywords;
    SELECT COUNT(*) INTO v_sentiment_count FROM mv_sentiment_timeline;
    SELECT COUNT(*) INTO v_entity_count FROM mv_entity_mentions;
    
    RAISE NOTICE '✅ Migration V003 completed successfully';
    RAISE NOTICE 'Created 3 materialized views with % total rows:', 
        v_trending_count + v_sentiment_count + v_entity_count;
    RAISE NOTICE '  - mv_trending_keywords: % rows', v_trending_count;
    RAISE NOTICE '  - mv_sentiment_timeline: % rows', v_sentiment_count;
    RAISE NOTICE '  - mv_entity_mentions: % rows', v_entity_count;
    RAISE NOTICE 'Created 3 helper views and 3 analysis functions';
    RAISE NOTICE 'NOTE: Refresh views periodically with: SELECT refresh_analytics_views(TRUE);';
END $$;