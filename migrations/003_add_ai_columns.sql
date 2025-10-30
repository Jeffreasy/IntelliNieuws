-- Migration: Add AI processing columns to articles table
-- Description: Extends articles table with AI-enriched data fields

-- Add AI processing columns
ALTER TABLE articles ADD COLUMN IF NOT EXISTS ai_processed BOOLEAN DEFAULT FALSE;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS ai_sentiment FLOAT;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS ai_sentiment_label VARCHAR(20);
ALTER TABLE articles ADD COLUMN IF NOT EXISTS ai_categories JSONB;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS ai_entities JSONB;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS ai_summary TEXT;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS ai_keywords JSONB;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS ai_processed_at TIMESTAMP;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS ai_error TEXT;

-- Add comments for documentation
COMMENT ON COLUMN articles.ai_processed IS 'Whether article has been processed by AI';
COMMENT ON COLUMN articles.ai_sentiment IS 'Sentiment score from -1.0 (negative) to 1.0 (positive)';
COMMENT ON COLUMN articles.ai_sentiment_label IS 'Sentiment label: positive, negative, neutral';
COMMENT ON COLUMN articles.ai_categories IS 'AI-detected categories with confidence scores';
COMMENT ON COLUMN articles.ai_entities IS 'Extracted entities: persons, organizations, locations';
COMMENT ON COLUMN articles.ai_summary IS 'AI-generated summary';
COMMENT ON COLUMN articles.ai_keywords IS 'Extracted keywords with relevance scores';
COMMENT ON COLUMN articles.ai_processed_at IS 'Timestamp when AI processing completed';
COMMENT ON COLUMN articles.ai_error IS 'Error message if AI processing failed';

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_articles_ai_processed ON articles(ai_processed) WHERE ai_processed = FALSE;
CREATE INDEX IF NOT EXISTS idx_articles_ai_sentiment ON articles(ai_sentiment) WHERE ai_sentiment IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_articles_ai_categories ON articles USING GIN(ai_categories) WHERE ai_categories IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_articles_ai_entities ON articles USING GIN(ai_entities) WHERE ai_entities IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_articles_ai_keywords ON articles USING GIN(ai_keywords) WHERE ai_keywords IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_articles_ai_processed_at ON articles(ai_processed_at) WHERE ai_processed_at IS NOT NULL;

-- Create view for AI-enriched articles
CREATE OR REPLACE VIEW v_ai_enriched_articles AS
SELECT 
    id,
    title,
    summary,
    url,
    published,
    source,
    keywords,
    category,
    ai_processed,
    ai_sentiment,
    ai_sentiment_label,
    ai_categories,
    ai_entities,
    ai_summary,
    ai_keywords,
    ai_processed_at,
    created_at,
    updated_at
FROM articles
WHERE ai_processed = TRUE
  AND ai_error IS NULL;

-- Create view for pending AI processing
CREATE OR REPLACE VIEW v_pending_ai_processing AS
SELECT 
    id,
    title,
    summary,
    url,
    published,
    source,
    created_at
FROM articles
WHERE ai_processed = FALSE
   OR (ai_processed = TRUE AND ai_error IS NOT NULL)
ORDER BY created_at DESC;

-- Create function to get articles by entity
CREATE OR REPLACE FUNCTION get_articles_by_entity(entity_name TEXT, entity_type TEXT DEFAULT NULL)
RETURNS TABLE (
    article_id BIGINT,
    title TEXT,
    summary TEXT,
    url TEXT,
    published TIMESTAMP,
    source TEXT,
    sentiment FLOAT,
    entities JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        a.id,
        a.title,
        a.summary,
        a.url,
        a.published,
        a.source,
        a.ai_sentiment,
        a.ai_entities
    FROM articles a
    WHERE a.ai_processed = TRUE
      AND a.ai_entities IS NOT NULL
      AND (
          (entity_type IS NULL AND a.ai_entities::text ILIKE '%' || entity_name || '%')
          OR (entity_type = 'persons' AND a.ai_entities->'persons' ? entity_name)
          OR (entity_type = 'organizations' AND a.ai_entities->'organizations' ? entity_name)
          OR (entity_type = 'locations' AND a.ai_entities->'locations' ? entity_name)
      )
    ORDER BY a.published DESC;
END;
$$ LANGUAGE plpgsql;

-- Create function to get sentiment statistics
CREATE OR REPLACE FUNCTION get_sentiment_stats(
    source_filter TEXT DEFAULT NULL,
    start_date TIMESTAMP DEFAULT NULL,
    end_date TIMESTAMP DEFAULT NULL
)
RETURNS TABLE (
    total_articles BIGINT,
    positive_count BIGINT,
    neutral_count BIGINT,
    negative_count BIGINT,
    avg_sentiment FLOAT,
    most_positive_title TEXT,
    most_negative_title TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COUNT(*)::BIGINT as total,
        COUNT(CASE WHEN LOWER(ai_sentiment_label) = 'positive' THEN 1 END)::BIGINT as positive,
        COUNT(CASE WHEN LOWER(ai_sentiment_label) = 'neutral' THEN 1 END)::BIGINT as neutral,
        COUNT(CASE WHEN LOWER(ai_sentiment_label) = 'negative' THEN 1 END)::BIGINT as negative,
        AVG(ai_sentiment) as avg_sent,
        (SELECT title FROM articles WHERE ai_sentiment IS NOT NULL 
         AND (source_filter IS NULL OR source = source_filter)
         AND (start_date IS NULL OR published >= start_date)
         AND (end_date IS NULL OR published <= end_date)
         ORDER BY ai_sentiment DESC LIMIT 1) as most_pos,
        (SELECT title FROM articles WHERE ai_sentiment IS NOT NULL 
         AND (source_filter IS NULL OR source = source_filter)
         AND (start_date IS NULL OR published >= start_date)
         AND (end_date IS NULL OR published <= end_date)
         ORDER BY ai_sentiment ASC LIMIT 1) as most_neg
    FROM articles
    WHERE ai_processed = TRUE
      AND ai_sentiment IS NOT NULL
      AND (source_filter IS NULL OR source = source_filter)
      AND (start_date IS NULL OR published >= start_date)
      AND (end_date IS NULL OR published <= end_date);
END;
$$ LANGUAGE plpgsql;

-- Create function to get trending topics
CREATE OR REPLACE FUNCTION get_trending_topics(
    hours_back INTEGER DEFAULT 24,
    min_articles INTEGER DEFAULT 3
)
RETURNS TABLE (
    keyword TEXT,
    article_count BIGINT,
    avg_sentiment FLOAT,
    sources TEXT[]
) AS $$
BEGIN
    RETURN QUERY
    WITH keywords_expanded AS (
        SELECT 
            a.id,
            a.source,
            a.ai_sentiment,
            jsonb_array_elements(a.ai_keywords) as kw
        FROM articles a
        WHERE a.ai_processed = TRUE
          AND a.ai_keywords IS NOT NULL
          AND a.published >= NOW() - (hours_back || ' hours')::INTERVAL
    ),
    keyword_stats AS (
        SELECT 
            kw->>'word' as word,
            COUNT(DISTINCT id)::BIGINT as cnt,
            AVG(ai_sentiment) as avg_sent,
            ARRAY_AGG(DISTINCT source) as srcs
        FROM keywords_expanded
        GROUP BY kw->>'word'
        HAVING COUNT(DISTINCT id) >= min_articles
    )
    SELECT * FROM keyword_stats
    ORDER BY cnt DESC, avg_sent DESC
    LIMIT 20;
END;
$$ LANGUAGE plpgsql;

-- Grant permissions (adjust as needed)
GRANT SELECT ON v_ai_enriched_articles TO PUBLIC;
GRANT SELECT ON v_pending_ai_processing TO PUBLIC;