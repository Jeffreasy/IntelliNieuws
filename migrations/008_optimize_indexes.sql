-- Migration: Optimize indexes for scraper performance
-- Version: 3.0
-- Description: Add composite indexes for frequently used queries

-- Index for content extraction queries (used by background processor)
-- This speeds up GetArticlesNeedingContent() significantly
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_content_extraction 
ON articles(content_extracted, created_at DESC) 
WHERE (content_extracted = FALSE OR content_extracted IS NULL);

-- Index for published date sorting (most frequent query pattern)
-- Speeds up List() queries ordered by published DESC
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_published_desc 
ON articles(published DESC) 
WHERE published IS NOT NULL;

-- Composite index for source + published filtering
-- Optimizes queries filtering by source with date sorting
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_source_published 
ON articles(source, published DESC);

-- Index for category filtering with published date
-- Speeds up category-based queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_category_published 
ON articles(category, published DESC)
WHERE category IS NOT NULL AND category != '';

-- Index for URL existence checks (used by ExistsByURLBatch)
-- This is likely already covered by the UNIQUE constraint, but explicit is better
-- Note: Partial index for non-null URLs only
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_url_lookup
ON articles(url)
WHERE url IS NOT NULL AND url != '';

-- Index for full-text search performance
-- Speeds up Search() queries using to_tsvector
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_fulltext_search
ON articles USING gin(to_tsvector('english', title || ' ' || COALESCE(summary, '')));

-- Update table statistics for query planner
ANALYZE articles;

-- Log completion
DO $$ 
BEGIN 
    RAISE NOTICE 'Scraper index optimizations completed successfully';
END $$;