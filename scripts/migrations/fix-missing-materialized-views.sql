-- ============================================================================
-- Fix Missing Materialized Views from V003 Migration
-- Created: 2025-10-30
-- ============================================================================

-- Drop existing views if any
DROP MATERIALIZED VIEW IF EXISTS mv_sentiment_timeline CASCADE;
DROP MATERIALIZED VIEW IF EXISTS mv_entity_mentions CASCADE;

-- ============================================================================
-- MATERIALIZED VIEW: ARTICLE SENTIMENT TIMELINE
-- ============================================================================

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
    ARRAY_AGG(a.id) FILTER (WHERE a.ai_sentiment IS NOT NULL) AS top_article_ids
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
    ROUND(AVG(ai_sentiment) FILTER (WHERE ai_sentiment IS NOT NULL)::NUMERIC, 3) AS avg_sentiment,
    MAX(published) AS last_mentioned,
    ARRAY_AGG(article_id) AS recent_article_ids
FROM entity_extraction
GROUP BY entity_value, entity_type, day_bucket
HAVING COUNT(DISTINCT article_id) >= 2;

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
-- REFRESH AND ANALYZE
-- ============================================================================

-- Initial data population
ANALYZE mv_sentiment_timeline;
ANALYZE mv_entity_mentions;

-- Success notification
DO $$ 
DECLARE
    v_sentiment_count BIGINT;
    v_entity_count BIGINT;
BEGIN
    SELECT COUNT(*) INTO v_sentiment_count FROM mv_sentiment_timeline;
    SELECT COUNT(*) INTO v_entity_count FROM mv_entity_mentions;
    
    RAISE NOTICE '✅ Missing materialized views created successfully';
    RAISE NOTICE '  - mv_sentiment_timeline: % rows', v_sentiment_count;
    RAISE NOTICE '  - mv_entity_mentions: % rows', v_entity_count;
END $$;