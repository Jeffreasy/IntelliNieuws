-- Migration: 001_create_tables.sql
-- Description: Create initial database schema for news scraper

-- Create articles table
CREATE TABLE IF NOT EXISTS articles (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    summary TEXT,
    url VARCHAR(1000) NOT NULL UNIQUE,
    published TIMESTAMP NOT NULL,
    source VARCHAR(100) NOT NULL,
    keywords TEXT[], -- Array of keywords
    image_url VARCHAR(1000),
    author VARCHAR(200),
    category VARCHAR(100),
    content_hash VARCHAR(64) UNIQUE, -- SHA256 hash for duplicate detection
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for articles
CREATE INDEX idx_articles_source ON articles(source);
CREATE INDEX idx_articles_published ON articles(published DESC);
CREATE INDEX idx_articles_category ON articles(category);
CREATE INDEX idx_articles_created_at ON articles(created_at DESC);
CREATE INDEX idx_articles_content_hash ON articles(content_hash);

-- Create sources table
CREATE TABLE IF NOT EXISTS sources (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    domain VARCHAR(255) NOT NULL UNIQUE,
    rss_feed_url VARCHAR(1000),
    use_rss BOOLEAN NOT NULL DEFAULT true,
    use_dynamic BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    rate_limit_sec INTEGER NOT NULL DEFAULT 5,
    last_scraped_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for sources
CREATE INDEX idx_sources_is_active ON sources(is_active);
CREATE INDEX idx_sources_domain ON sources(domain);

-- Create scraping_jobs table
CREATE TABLE IF NOT EXISTS scraping_jobs (
    id BIGSERIAL PRIMARY KEY,
    source VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, running, completed, failed
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    error TEXT,
    article_count INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for scraping_jobs
CREATE INDEX idx_scraping_jobs_source ON scraping_jobs(source);
CREATE INDEX idx_scraping_jobs_status ON scraping_jobs(status);
CREATE INDEX idx_scraping_jobs_created_at ON scraping_jobs(created_at DESC);

-- Insert default sources
INSERT INTO sources (name, domain, rss_feed_url, use_rss, use_dynamic, is_active, rate_limit_sec) 
VALUES 
    ('NU.nl', 'nu.nl', 'https://www.nu.nl/rss', true, false, true, 5),
    ('AD.nl', 'ad.nl', 'https://www.ad.nl/rss.xml', true, false, true, 5),
    ('NOS.nl', 'nos.nl', 'https://feeds.nos.nl/nosnieuwsalgemeen', true, false, true, 5)
ON CONFLICT (domain) DO NOTHING;

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for updated_at
CREATE TRIGGER update_articles_updated_at
    BEFORE UPDATE ON articles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sources_updated_at
    BEFORE UPDATE ON sources
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create view for article statistics
CREATE OR REPLACE VIEW article_stats AS
SELECT 
    source,
    COUNT(*) as total_articles,
    MAX(published) as latest_article,
    MIN(published) as oldest_article,
    COUNT(DISTINCT DATE(published)) as unique_days
FROM articles
GROUP BY source;

-- Create view for recent scraping activity
CREATE OR REPLACE VIEW recent_scraping_activity AS
SELECT 
    sj.source,
    sj.status,
    sj.article_count,
    sj.started_at,
    sj.completed_at,
    sj.error,
    s.name as source_name,
    s.is_active as source_active
FROM scraping_jobs sj
LEFT JOIN sources s ON sj.source = s.domain
ORDER BY sj.created_at DESC
LIMIT 100;