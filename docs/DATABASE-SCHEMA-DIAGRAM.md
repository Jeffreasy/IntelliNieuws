
# Database Schema Diagram

## 📊 Complete Schema Overzicht

```
┌─────────────────────────────────────────────────────────────────────┐
│                        NIEUWSSCRAPER DATABASE                       │
│                          Professional Schema V2                      │
└─────────────────────────────────────────────────────────────────────┘

┌──────────────────────────┐
│   schema_migrations      │
├──────────────────────────┤
│ • version (PK)           │
│ • description            │
│ • applied_at             │
│ • applied_by             │
│ • execution_time_ms      │
│ • checksum               │
└──────────────────────────┘
         │
         │ Tracks
         ▼
┌────────────────────────────────────────────────────────────────────┐
│                            CORE TABLES                              │
└────────────────────────────────────────────────────────────────────┘

┌──────────────────────────┐      ┌──────────────────────────┐
│       sources            │      │      scraping_jobs       │
├──────────────────────────┤      ├──────────────────────────┤
│ • id (PK)                │      │ • id (PK)                │
│ • name (UNIQUE)          │      │ • job_uuid (UNIQUE)      │
│ • domain (UNIQUE)        │      │ • source                 │─┐
│ • rss_feed_url           │      │ • scraping_method        │ │
│ • use_rss                │      │ • status                 │ │
│ • use_dynamic            │      │ • started_at             │ │
│ • is_active              │      │ • completed_at           │ │
│ • rate_limit_seconds     │      │ • execution_time_ms      │ │
│ • max_articles_per_scrape│      │ • articles_found         │ │
│ • last_scraped_at        │      │ • articles_new           │ │
│ • last_success_at        │      │ • articles_updated       │ │
│ • last_error             │      │ • articles_skipped       │ │
│ • consecutive_failures   │      │ • error                  │ │
│ • total_articles_scraped │      │ • error_code             │ │
│ • created_at             │      │ • retry_count            │ │
│ • updated_at             │      │ • max_retries            │ │
│ • created_by             │      │ • created_at             │ │
└──────────────────────────┘      │ • created_by             │ │
         │                         └──────────────────────────┘ │
         │ Referenced by                      │                 │
         ▼                                     │                 │
┌──────────────────────────────────────────────┴────────────────┴───┐
│                           articles                                 │
├────────────────────────────────────────────────────────────────────┤
│ • id (PK)                                                          │
│ • title                    • ai_processed                          │
│ • summary                  • ai_processed_at                       │
│ • content                  • ai_sentiment (-1.0 to 1.0)           │
│ • url (UNIQUE)             • ai_sentiment_label                    │
│ • published                • ai_summary                            │
│ • source (FK→sources)      • ai_categories (JSONB)                │
│ • author                   • ai_entities (JSONB)                   │
│ • category                 • ai_keywords (JSONB)                   │
│ • keywords (TEXT[])        • ai_stock_tickers (JSONB)             │
│ • image_url                • ai_error                              │
│ • content_hash (UNIQUE)    • stock_data (JSONB)                   │
│ • content_extracted        • stock_data_updated_at                │
│ • content_extracted_at     • created_at                            │
│ • created_by               • updated_at                            │
└────────────────────────────────────────────────────────────────────┘
         ▲
         │ Referenced by (FK)
         │
┌──────────────────────────┐
│        emails            │
├──────────────────────────┤
│ • id (PK)                │
│ • message_id (UNIQUE)    │
│ • message_uid            │
│ • thread_id              │
│ • sender                 │
│ • sender_name            │
│ • recipient              │
│ • subject                │
│ • body_text              │
│ • body_html              │
│ • snippet (auto)         │
│ • received_date          │
│ • sent_date              │
│ • status                 │
│ • processed_at           │
│ • article_id (FK)        │───────────────┘
│ • article_created        │
│ • error                  │
│ • error_code             │
│ • retry_count            │
│ • max_retries            │
│ • last_retry_at          │
│ • has_attachments        │
│ • attachment_count       │
│ • is_read                │
│ • is_flagged             │
│ • is_spam                │
│ • importance             │
│ • metadata (JSONB)       │
│ • headers (JSONB)        │
│ • labels (TEXT[])        │
│ • size_bytes             │
│ • spam_score             │
│ • created_at             │
│ • updated_at             │
│ • created_by             │
└──────────────────────────┘

┌────────────────────────────────────────────────────────────────────┐
│                      MATERIALIZED VIEWS                             │
└────────────────────────────────────────────────────────────────────┘

┌──────────────────────────┐      ┌──────────────────────────┐
│  mv_trending_keywords    │      │ mv_sentiment_timeline    │
├──────────────────────────┤      ├──────────────────────────┤
│ • keyword                │      │ • time_bucket (hour)     │
│ • hour_bucket            │      │ • source                 │
│ • day_bucket             │      │ • category               │
│ • article_count          │      │ • article_count          │
│ • source_count           │      │ • positive_count         │
│ • sources (TEXT[])       │      │ • neutral_count          │
│ • avg_sentiment          │      │ • negative_count         │
│ • avg_relevance          │      │ • avg_sentiment          │
│ • latest_article_date    │      │ • sentiment_stddev       │
│ • first_article_date     │      │ • max_sentiment          │
│ • categories (TEXT[])    │      │ • min_sentiment          │
│ • trending_score         │      │ • top_article_ids        │
└──────────────────────────┘      └──────────────────────────┘
    ↓ Feeds                           ↓ Feeds
v_trending_keywords_24h           v_sentiment_trends_7d

┌──────────────────────────┐
│  mv_entity_mentions      │
├──────────────────────────┤
│ • entity                 │
│ • entity_type            │
│ • day_bucket             │
│ • mention_count          │
│ • source_count           │
│ • sources (TEXT[])       │
│ • categories (TEXT[])    │
│ • avg_sentiment          │
│ • last_mentioned         │
│ • recent_article_ids     │
└──────────────────────────┘
    ↓ Feeds
v_hot_entities_7d

┌────────────────────────────────────────────────────────────────────┐
│                         MONITORING VIEWS                            │
└────────────────────────────────────────────────────────────────────┘

v_active_sources              - Sources ready to scrape
v_article_stats               - Statistics per source
v_recent_scraping_activity    - Last 100 scraping jobs
v_emails_pending_processing   - Emails awaiting processing
v_email_stats                 - Email processing statistics
v_email_sender_stats          - Statistics per sender
v_recent_email_activity       - Last 100 emails
v_trending_keywords_24h       - Top 50 trending (24h)
v_sentiment_trends_7d         - Sentiment trends (7d)
v_hot_entities_7d             - Top 100 entities (7d)

┌────────────────────────────────────────────────────────────────────┐
│                        HELPER FUNCTIONS                             │
└────────────────────────────────────────────────────────────────────┘

Email Management:
  • get_emails_for_retry(max_age_hours, batch_size)
  • mark_email_processed(email_id, article_id)
  • mark_email_failed(email_id, error, error_code)
  • cleanup_old_emails(days_to_keep, keep_with_articles)

Analytics:
  • get_trending_topics(hours_back, min_articles, limit)
  • get_entity_sentiment_analysis(entity, days_back)
  • refresh_analytics_views(concurrent)
  • get_maintenance_schedule()

┌────────────────────────────────────────────────────────────────────┐
│                       INDEX STRATEGY                                │
└────────────────────────────────────────────────────────────────────┘

articles (20+ indexes):
  ├── B-tree: id, source, published, category, created_at
  ├── GIN: keywords, ai_categories, ai_entities, ai_keywords, ai_stock_tickers
  ├── Full-text: title, summary, content, combined
  ├── Composite: (source, published), (category, published)
  └── Partial: content_extracted = FALSE, ai_processed = FALSE

emails (15+ indexes):
  ├── B-tree: message_id, sender, received_date, article_id
  ├── GIN: subject (FTS), metadata, headers, labels
  ├── Composite: (sender, status, received_date)
  └── Partial: status = 'pending', status = 'failed'

sources (3 indexes):
  ├── B-tree: domain
  ├── Partial: is_active = TRUE
  └── Composite: domain lookup

scraping_jobs (5 indexes):
  ├── B-tree: id, job_uuid, source, status, created_at
  └── Composite: (source, status, created_at)

┌────────────────────────────────────────────────────────────────────┐
│                        DATA FLOW                                    │
└────────────────────────────────────────────────────────────────────┘

RSS Feed ──┐
           ├─► Scraper ──► scraping_jobs ──► articles ──┐
Dynamic ───┘                                             │
                                                         ├─► AI Processor
Email ─────► emails ─────────────────────────────────► articles
                                                         │
                                                         ├─► Content Extractor
                                                         │
                                                         └─► Stock Enrichment
                                                              │
                                                              ▼
                                                        Materialized Views
                                                              │
                                                              ▼
                                                        Analytics API
                                                              │
                                                              ▼
                                                          Frontend

┌────────────────────────────────────────────────────────────────────┐
│                    CONSTRAINT ENFORCEMENT                           │
└────────────────────────────────────────────────────────────────────┘

articles:
  ✓ UNIQUE (url)                    - No duplicate URLs
  ✓ UNIQUE (content_hash)           - No duplicate content
  ✓ CHECK (url ~ '^https?://')      - Valid URL format
  ✓ CHECK (ai_sentiment -1.0..1.0)  - Valid sentiment range
  ✓ CHECK sentiment label sync      - Label matches score

sources:
  ✓ UNIQUE (name)                   - Unique source names
  ✓ UNIQUE (domain)                 - Unique domains
  ✓ CHECK (use_rss OR use_dynamic)  - Must use one method
  ✓ CHECK domain format             - Valid domain regex

scraping_jobs:
  ✓ UNIQUE (job_uuid)               - Unique job identifiers
  ✓ CHECK status values             - Valid status enum
  ✓ CHECK timing logic              - Started <= Completed
  ✓ CHECK article counts >= 0       - Non-negative counts

emails:
  ✓ UNIQUE (message_id)             - No duplicate emails
  ✓ CHECK status values             - Valid status enum
  ✓ CHECK article linkage           - ID matches created flag
  ✓ CHECK retry logic               - Count <= max
  ✓ FK (article_id) CASCADE         - Referential integrity

┌────────────────────────────────────────────────────────────────────┐
│                    TRIGGER AUTOMATION                               │
└────────────────────────────────────────────────────────────────────┘

Automatic Timestamps:
  articles.updated_at    ───► trigger_set_updated_at()
  sources.updated_at     ───► trigger_set_updated_at()
  emails.updated_at      ───► trigger_set_updated_at()

Email Processing:
  emails.snippet         ───► trigger_update_email_snippet()
  emails.article_id      ───► trigger_validate_email_article()

┌────────────────────────────────────────────────────────────────────┐
│                   PERFORMANCE OPTIMIZATION                          │
└────────────────────────────────────────────────────────────────────┘

Query Patterns → Index Strategy:
  
  1. Article List (ORDER BY published DESC)
     → idx_articles_published_desc
     → Response: 50ms

  2. Source Filter (WHERE source = 'nu.nl')
     → idx_articles_source_published (composite)
     → Response: 30ms

  3. Full-Text Search (title/summary/content)
     → idx_articles_combined_fts (GIN)
     → Response: 100ms

  4. Trending Keywords (24h aggregation)
     → mv_trending_keywords (materialized)
     → Response: 0.5s (was 5s)

  5. Entity Lookup (JSONB ? 'entity')
     → idx_articles_ai_entities_gin (GIN)
     → Response: 75ms

  6. Stock Ticker Lookup (JSONB @> ticker)
     → idx_articles_stock_tickers_gin (GIN)
     → Response: 50ms

┌────────────────────────────────────────────────────────────────────┐
│                      ANALYTICS PIPELINE                             │
└────────────────────────────────────────────────────────────────────┘

Raw Articles
    │
    ├─► AI Processing
    │     ├─► Sentiment Analysis    ──► ai_sentiment, ai_sentiment_label
    │     ├─► Entity Extraction     ──► ai_entities (persons, orgs, locations)
    │     ├─► Keyword Extraction    ──► ai_keywords
    │     ├─► Category Detection    ──► ai_categories
    │     ├─► Stock Ticker Detection ──► ai_stock_tickers
    │     └─► Summary Generation    ──► ai_summary
    │
    ├─► Content Extraction
    │     └─► HTML Processing       ──► content (full text)
    │
    ├─► Stock Enrichment
    │     └─► Market Data           ──► stock_data (prices, metrics)
    │
    └─► Analytics Aggregation
          ├─► Trending Keywords     ──► mv_trending_keywords
          ├─► Sentiment Timeline    ──► mv_sentiment_timeline
          └─► Entity Mentions       ──► mv_entity_mentions
                │
                └─► Analytics API Endpoints

┌────────────────────────────────────────────────────────────────────┐
│                    MAINTENANCE WORKFLOW                             │
└────────────────────────────────────────────────────────────────────┘

Periodic Tasks:

  Every 5-15 minutes:
    → Refresh materialized views
    → SELECT refresh_analytics_views(TRUE);

  Daily:
    → Auto-vacuum (automatic)
    → Update statistics (automatic)
    → Health check monitoring

  Weekly:
    → Clean old emails
    → SELECT cleanup_old_emails(90, TRUE);
    → Clean old scraping jobs
    → Manual maintenance review

  Monthly:
    → Full VACUUM ANALYZE
    → Index optimization
    → Performance audit
    → Backup validation

┌────────────────────────────────────────────────────────────────────┐
│                   TABLE SIZE ESTIMATES                              │
└────────────────────────────────────────────────────────────────────┘

Current (185 articles):
  articles              ~2.5 MB
  sources               ~8 KB
  scraping_jobs         ~64 KB
  emails                ~512 KB
  mv_trending_keywords  ~136 KB
  Total                 ~3.2 MB

Projected (100K articles):
  articles              ~1.2 GB (with indexes: ~3 GB)
  mv_trending_keywords  ~50 MB
  mv_sentiment_timeline ~20 MB
  mv_entity_mentions    ~30 MB
  
Recommended: Partition articles by month at 1M+ rows

┌────────────────────────────────────────────────────────────────────┐
│                      API ENDPOINTS                                  │
└────────────────────────────────────────────────────────────────────┘

Analytics (Public):
  GET  /api/v1/analytics/trending
  GET  /api/v1/analytics/sentiment-trends
  GET  /api/v1/analytics/hot-entities
  GET  /api/v1/analytics/entity-sentiment
  GET  /api/v1/analytics/overview
  GET  /api/v1/analytics/article-stats
  GET  /api/v1/analytics/maintenance-schedule
  GET  /api/v1/analytics/database-health
  POST /api/v1/analytics/refresh

Articles (Public with optional auth):
  GET  /api/v1/articles
  GET  /api/v1/articles/:id
  GET  /api/v1/articles/search
  GET  /api/v1/articles/stats
  GET  /api/v1/articles/by-ticker/:symbol
  POST /api/v1/articles/:id/extract-content (protected)

Health (Public):
  GET  /health
  GET  /health/live
  GET  /health/ready
  GET  /health/metrics

┌────────────────────────────────────────────────────────────────────┐
│                   SECURITY MODEL                                    │
└────────────────────────────────────────────────────────────────────┘

Public Access:
  ✓ Analytics endpoints (no sensitive data)
  ✓ Article reading (public news)
  ✓ Health checks
  ✓ Stock data (public market data)

Protected Access (API Key Required):
  ✓ Scraping triggers
  ✓ AI processing triggers
  ✓ Cache management
  ✓ Email processing
  ✓ Content extraction

Database Level:
  ✓ Connection pooling (max 25)
  ✓ Prepared statements (SQL injection prevention)
  ✓ Foreign key constraints
  ✓ Check constraints
  ✓ Audit trails (created_by, timestamps)

┌────────────────────────────────────────────────────────────────────┐
│                   LEGEND                                            │
└────────────────────────────────────────────────────────────────────┘

PK   - Primary Key
FK   - Foreign Key
GIN  - Generalized Inverted Index (for JSONB, arrays, full-text)
JSONB - JSON Binary format (indexed, queryable)
TEXT[] - Array of text values
auto - Automatically generated via trigger
```

## 🔍 Relationships

### Foreign Keys
```
emails.article_id ──FK──► articles.id (ON DELETE SET NULL, ON UPDATE CASCADE)
```

### Logical Relationships
```
articles.source ──references──► sources.domain (not enforced FK for flexibility)
scraping_jobs.source ──references──► sources.domain (tracked relationship)
```

## 📊 Capacity Planning

### Current Capacity
- **185 articles** - Optimaal
- **3 sources** - Room for 50+ sources
- **93 jobs** - 30-day retention
- **Indexes** - Efficient up to 1M articles

### Growth Scenarios

**10K Articles:**
- Total size: ~30 MB
- Query performance: Excellent
- Action needed: None

**100K Articles:**
- Total size: ~3 GB
- Query performance: Good
- Action needed: Consider partitioning

**1M+ Articles:**
- Total size: ~30 GB
- Query performance: Requires optimization
- Action needed: Implement partitioning by month

## 🛠️ Maintenance Windows

| Task | Downtime | Impact | Frequency |
|------|----------|--------|-----------|
| Refresh views (CONCURRENT) | 0s | None | 5-15 min |
| VACUUM ANALYZE | 0s | None (auto) | Daily |
| Reindex (CONCURRENT) | 0s | CPU | Monthly |
| Full VACUUM | ~5min | Read-only | Quarterly |
| Schema upgrade | 0-5s | Minimal | As needed |

## 🎯 Query Optimization Tips

1. **Use Materialized Views** for analytics (90% faster)
2. **Leverage Partial Indexes** (WHERE clauses include