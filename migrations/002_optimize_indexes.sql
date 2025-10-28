-- Migration: 002_optimize_indexes.sql
-- Description: Add composite indexes for common query patterns

-- Composite index for filtering by source and published date (most common query)
CREATE INDEX IF NOT EXISTS idx_articles_source_published ON articles(source, published DESC);

-- Composite index for filtering by category and published date
CREATE INDEX IF NOT EXISTS idx_articles_category_published ON articles(category, published DESC);

-- Composite index for source and created_at (API queries)
CREATE INDEX IF NOT EXISTS idx_articles_source_created ON articles(source, created_at DESC);

-- GIN index for keywords array searching (for keyword filtering)
CREATE INDEX IF NOT EXISTS idx_articles_keywords_gin ON articles USING GIN(keywords);

-- Full-text search index on title and summary
CREATE INDEX IF NOT EXISTS idx_articles_title_search ON articles USING GIN(to_tsvector('english', title));
CREATE INDEX IF NOT EXISTS idx_articles_summary_search ON articles USING GIN(to_tsvector('english', COALESCE(summary, '')));

-- Composite full-text search index
CREATE INDEX IF NOT EXISTS idx_articles_fulltext_search ON articles USING GIN(
    to_tsvector('english', title || ' ' || COALESCE(summary, ''))
);

-- Partial index for active sources only (if we add is_deleted column later)
CREATE INDEX IF NOT EXISTS idx_articles_active_published ON articles(published DESC) 
WHERE source IS NOT NULL;

-- Index for date range queries
CREATE INDEX IF NOT EXISTS idx_articles_date_range ON articles(published, source);

-- Performance: Update statistics
ANALYZE articles;
ANALYZE sources;
ANALYZE scraping_jobs;

-- Add comments for documentation
COMMENT ON INDEX idx_articles_source_published IS 'Optimizes queries filtering by source and ordering by published date';
COMMENT ON INDEX idx_articles_keywords_gin IS 'Enables fast keyword array searches using GIN index';
COMMENT ON INDEX idx_articles_fulltext_search IS 'Enables full-text search across title and summary';