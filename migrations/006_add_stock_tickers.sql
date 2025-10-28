-- Migration: 006_add_stock_tickers.sql
-- Description: Add stock tickers column for entity extraction and stock data integration

-- Add stock_tickers column to articles table
ALTER TABLE articles ADD COLUMN IF NOT EXISTS ai_stock_tickers JSONB;

-- Add stock_data column for cached stock information
ALTER TABLE articles ADD COLUMN IF NOT EXISTS stock_data JSONB;

-- Create GIN index for efficient stock ticker queries
CREATE INDEX IF NOT EXISTS idx_articles_stock_tickers ON articles USING GIN(ai_stock_tickers) WHERE ai_stock_tickers IS NOT NULL;

-- Create index for articles with stock data
CREATE INDEX IF NOT EXISTS idx_articles_stock_data ON articles(id) WHERE stock_data IS NOT NULL;

-- Add stock_data_updated_at for cache invalidation
ALTER TABLE articles ADD COLUMN IF NOT EXISTS stock_data_updated_at TIMESTAMP;

-- Success message
DO $$
BEGIN
    RAISE NOTICE '✅ Stock tickers columns and indexes created successfully!';
END $$;