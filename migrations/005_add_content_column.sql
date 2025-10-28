-- Migration: Add content extraction columns
-- This enables storing full article text extracted from HTML pages
-- Part of the hybrid RSS + HTML scraping approach

-- Add content column for full article text
ALTER TABLE articles 
ADD COLUMN IF NOT EXISTS content TEXT,
ADD COLUMN IF NOT EXISTS content_extracted BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS content_extracted_at TIMESTAMPTZ;

-- Index for finding articles without content (for background processing)
CREATE INDEX IF NOT EXISTS idx_articles_needs_content 
ON articles(content_extracted, created_at) 
WHERE content_extracted = FALSE OR content_extracted IS NULL;

-- Index for content search (if needed later)
CREATE INDEX IF NOT EXISTS idx_articles_content_search 
ON articles USING gin(to_tsvector('dutch', content))
WHERE content IS NOT NULL;

-- Update existing articles to mark them as needing content extraction
UPDATE articles 
SET content_extracted = FALSE 
WHERE content IS NULL OR content = '';

-- Add column comments for documentation
COMMENT ON COLUMN articles.content IS 'Full article text extracted from HTML page';
COMMENT ON COLUMN articles.content_extracted IS 'Whether full content has been extracted from the article URL';
COMMENT ON COLUMN articles.content_extracted_at IS 'Timestamp when content was successfully extracted';

-- Performance statistics
DO $$
DECLARE
    total_articles INT;
    needs_content INT;
BEGIN
    SELECT COUNT(*) INTO total_articles FROM articles;
    SELECT COUNT(*) INTO needs_content FROM articles WHERE content_extracted = FALSE OR content_extracted IS NULL;
    
    RAISE NOTICE 'Migration complete: % total articles, % need content extraction', total_articles, needs_content;
END $$;