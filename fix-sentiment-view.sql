-- Create v_sentiment_trends_7d view for sentiment analysis
-- Aggregate across all categories per day and source
CREATE OR REPLACE VIEW v_sentiment_trends_7d AS
SELECT
    DATE(st.time_bucket)::TEXT AS day,
    st.source,
    SUM(st.article_count)::INT AS total_articles,
    SUM(st.positive_count)::INT AS positive_count,
    SUM(st.neutral_count)::INT AS neutral_count,
    SUM(st.negative_count)::INT AS negative_count,
    ROUND(AVG(st.avg_sentiment)::NUMERIC, 3) AS avg_sentiment,
    ROUND((SUM(st.positive_count)::NUMERIC / NULLIF(SUM(st.article_count), 0) * 100), 1) AS positive_percentage,
    ROUND((SUM(st.negative_count)::NUMERIC / NULLIF(SUM(st.article_count), 0) * 100), 1) AS negative_percentage
FROM mv_sentiment_timeline st
WHERE st.time_bucket >= CURRENT_DATE - INTERVAL '7 days'
GROUP BY DATE(st.time_bucket), st.source
ORDER BY day DESC, st.source;

-- Also create v_hot_entities_7d if missing
CREATE OR REPLACE VIEW v_hot_entities_7d AS
SELECT
    em.entity,
    em.entity_type,
    SUM(em.mention_count)::BIGINT AS total_mentions,
    COUNT(DISTINCT DATE(em.day_bucket))::INT AS days_mentioned,
    em.sources::TEXT[] AS sources,
    ROUND(AVG(em.avg_sentiment)::NUMERIC, 3) AS overall_sentiment,
    MAX(em.last_mentioned) AS most_recent_mention
FROM mv_entity_mentions em
WHERE em.day_bucket >= CURRENT_DATE - INTERVAL '7 days'
  AND em.entity IS NOT NULL
  AND em.entity != ''
GROUP BY em.entity, em.entity_type, em.sources
HAVING SUM(em.mention_count) >= 2
ORDER BY total_mentions DESC, days_mentioned DESC;