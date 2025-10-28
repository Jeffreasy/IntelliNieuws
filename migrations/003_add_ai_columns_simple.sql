-- Simple AI Migration (without stored procedures)
-- This adds only the essential AI columns to get started

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

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_articles_ai_processed ON articles(ai_processed) WHERE ai_processed = FALSE;
CREATE INDEX IF NOT EXISTS idx_articles_ai_sentiment ON articles(ai_sentiment) WHERE ai_sentiment IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_articles_ai_categories ON articles USING GIN(ai_categories) WHERE ai_categories IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_articles_ai_entities ON articles USING GIN(ai_entities) WHERE ai_entities IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_articles_ai_keywords ON articles USING GIN(ai_keywords) WHERE ai_keywords IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_articles_ai_processed_at ON articles(ai_processed_at) WHERE ai_processed_at IS NOT NULL;

-- Success message
DO $$
BEGIN
    RAISE NOTICE 'AI columns and indexes created successfully!';
END $$;