# Quick Reference - Database Migrations

## 🚀 One-Line Commands

```bash
# Apply all migrations (new installation)
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper -f /migrations/V001__create_base_schema.sql && \
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper -f /migrations/V002__create_emails_table.sql && \
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper -f /migrations/V003__create_analytics_views.sql

# Health check
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper -f /migrations/utilities/02_health_check.sql

# Maintenance
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper -f /migrations/utilities/03_maintenance.sql

# Backup
docker exec nieuws-scraper-db pg_dump -U postgres nieuws_scraper > backup_$(date +%Y%m%d_%H%M%S).sql

# Restore
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper < backup_20250130_120000.sql
```

## 📊 Useful SQL Queries

### Migration Status
```sql
SELECT * FROM schema_migrations ORDER BY applied_at DESC;
```

### Table Sizes
```sql
SELECT 
    tablename,
    pg_size_pretty(pg_total_relation_size('public.'||tablename)) AS size
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size('public.'||tablename) DESC;
```

### Row Counts
```sql
SELECT 
    'articles' AS table_name, COUNT(*) AS rows FROM articles
UNION ALL SELECT 'sources', COUNT(*) FROM sources
UNION ALL SELECT 'scraping_jobs', COUNT(*) FROM scraping_jobs
UNION ALL SELECT 'emails', COUNT(*) FROM emails;
```

### Processing Status
```sql
SELECT 
    COUNT(*) FILTER (WHERE content_extracted = TRUE) AS with_content,
    COUNT(*) FILTER (WHERE ai_processed = TRUE) AS ai_processed,
    COUNT(*) AS total
FROM articles;
```

### Recent Activity
```sql
SELECT 
    DATE(created_at) AS date,
    COUNT(*) AS articles
FROM articles
WHERE created_at >= CURRENT_DATE - INTERVAL '7 days'
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

### Trending Keywords (Last 24h)
```sql
SELECT * FROM v_trending_keywords_24h LIMIT 10;
```

### Failed Jobs
```sql
SELECT source, status, error, created_at 
FROM scraping_jobs 
WHERE status = 'failed' 
ORDER BY created_at DESC 
LIMIT 10;
```

### Index Usage
```sql
SELECT 
    schemaname, tablename, indexname, 
    idx_scan AS scans,
    pg_size_pretty(pg_relation_size(indexrelid)) AS size
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan DESC
LIMIT 20;
```

### Cache Hit Ratio
```sql
SELECT 
    ROUND(100.0 * SUM(heap_blks_hit) / NULLIF(SUM(heap_blks_hit) + SUM(heap_blks_read), 0), 2) AS cache_hit_ratio
FROM pg_statio_user_tables;
```

## 🔄 Maintenance Tasks

### Refresh Analytics (Run every 5-15 minutes)
```sql
SELECT * FROM refresh_analytics_views(TRUE);
```

### Clean Old Emails (Run weekly)
```sql
SELECT * FROM cleanup_old_emails(90, TRUE);
```

### Vacuum & Analyze (Run daily)
```sql
VACUUM ANALYZE articles;
VACUUM ANALYZE emails;
```

### Reindex (Only if bloated)
```sql
REINDEX TABLE CONCURRENTLY articles;
```

### Update Statistics
```sql
ANALYZE articles;
ANALYZE sources;
ANALYZE scraping_jobs;
ANALYZE emails;
```

## 🔍 Troubleshooting

### Check Connection
```bash
docker exec nieuws-scraper-db psql -U postgres -c "SELECT version();"
```

### Check Active Connections
```sql
SELECT count(*) FROM pg_stat_activity WHERE datname = 'nieuws_scraper';
```

### Kill Long Running Query
```sql
SELECT pg_terminate_backend(pid) 
FROM pg_stat_activity 
WHERE query_start < CURRENT_TIMESTAMP - INTERVAL '5 minutes'
  AND state = 'active';
```

### Check Locks
```sql
SELECT * FROM pg_locks WHERE NOT granted;
```

### Find Bloated Tables
```sql
SELECT 
    schemaname, tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
    ROUND(100 * (pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename))::NUMERIC / 
        NULLIF(pg_total_relation_size(schemaname||'.'||tablename), 0), 2) AS bloat_pct
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

## 📈 Performance Monitoring

### Slow Queries
```sql
SELECT 
    query,
    calls,
    ROUND(mean_exec_time::NUMERIC, 2) AS avg_ms,
    ROUND(total_exec_time::NUMERIC, 2) AS total_ms
FROM pg_stat_statements
WHERE query NOT LIKE '%pg_stat%'
ORDER BY mean_exec_time DESC
LIMIT 10;
```

### Most Called Queries
```sql
SELECT 
    LEFT(query, 100) AS query_preview,
    calls,
    ROUND(mean_exec_time::NUMERIC, 2) AS avg_ms
FROM pg_stat_statements
WHERE query NOT LIKE '%pg_stat%'
ORDER BY calls DESC
LIMIT 10;
```

### Table I/O Stats
```sql
SELECT 
    schemaname, tablename,
    heap_blks_read AS disk_reads,
    heap_blks_hit AS cache_hits,
    ROUND(100.0 * heap_blks_hit / NULLIF(heap_blks_hit + heap_blks_read, 0), 2) AS cache_hit_ratio
FROM pg_statio_user_tables
WHERE schemaname = 'public'
ORDER BY heap_blks_read DESC;
```

## 🛠️ Common Operations

### Add New Source
```sql
INSERT INTO sources (name, domain, rss_feed_url, is_active)
VALUES ('Example', 'example.com', 'https://example.com/rss', TRUE);
```

### Disable Source
```sql
UPDATE sources SET is_active = FALSE WHERE domain = 'example.com';
```

### Reset Failed Job Counter
```sql
UPDATE sources SET consecutive_failures = 0 WHERE domain = 'example.com';
```

### Mark Article for Reprocessing
```sql
UPDATE articles SET ai_processed = FALSE, content_extracted = FALSE WHERE id = 123;
```

### Get Articles by Stock Ticker
```sql
SELECT id, title, published
FROM articles
WHERE ai_stock_tickers @> '[{"ticker": "AAPL"}]'::jsonb
ORDER BY published DESC;
```

### Get Articles by Entity
```sql
SELECT id, title, published
FROM articles
WHERE ai_entities->'persons' ? 'Elon Musk'
ORDER BY published DESC;
```

### Search Articles
```sql
SELECT id, title, ts_rank(to_tsvector('english', title || ' ' || summary), query) AS rank
FROM articles, to_tsquery('english', 'bitcoin & price') AS query
WHERE to_tsvector('english', title || ' ' || summary) @@ query
ORDER BY rank DESC
LIMIT 20;
```

## 📝 Configuration Values

### Optimal PostgreSQL Settings
```ini
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 16MB
maintenance_work_mem = 128MB
random_page_cost = 1.1
effective_io_concurrency = 200
max_parallel_workers_per_gather = 4
```

### Autovacuum Settings
```ini
autovacuum = on
autovacuum_naptime = 20s
autovacuum_max_workers = 3
```

## 🔐 Security Commands

### Create Read-Only User
```sql
CREATE USER readonly WITH PASSWORD 'secure_password';
GRANT CONNECT ON DATABASE nieuws_scraper TO readonly;
GRANT USAGE ON SCHEMA public TO readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO readonly;
```

### Create Application User
```sql
CREATE USER app_user WITH PASSWORD 'secure_password';
GRANT CONNECT ON DATABASE nieuws_scraper TO app_user;
GRANT USAGE ON SCHEMA public TO app_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_user;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_user;
```

### Revoke Permissions
```sql
REVOKE ALL ON DATABASE nieuws_scraper FROM public;
```

## 📞 Emergency Procedures

### Immediate Rollback
```bash
# V003 only
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper < migrations/rollback/V003__rollback.sql

# V002 and V003
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper < migrations/rollback/V003__rollback.sql
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper < migrations/rollback/V002__rollback.sql

# Full rollback (nuclear option)
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper < migrations/rollback/V003__rollback.sql
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper < migrations/rollback/V002__rollback.sql
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper < migrations/rollback/V001__rollback.sql
```

### Emergency Backup
```bash
docker exec nieuws-scraper-db pg_dump -U postgres -Fc nieuws_scraper > emergency_backup.dump
```

### Emergency Restore
```bash
docker exec -i nieuws-scraper-db pg_restore -U postgres -d nieuws_scraper -c emergency_backup.dump
```

## 🎯 Daily Checklist

- [ ] Run health check
- [ ] Review failed scraping jobs
- [ ] Check disk space
- [ ] Verify materialized views are fresh
- [ ] Monitor query performance
- [ ] Check error logs

## 📅 Weekly Tasks

- [ ] Run maintenance script
- [ ] Clean old emails
- [ ] Review and optimize slow queries
- [ ] Verify backups
- [ ] Update statistics
- [ ] Review index usage

## 📆 Monthly Tasks

- [ ] Full VACUUM ANALYZE
- [ ] Review table partitioning needs
- [ ] Optimize indexes
- [ ] Test backup restore
- [ ] Review and update documentation
- [ ] Performance audit

## 🔗 Quick Links

- **Full Documentation:** [`README.md`](README.md)
- **Migration Guide:** [`MIGRATION-GUIDE.md`](MIGRATION-GUIDE.md)
- **Health Check:** [`utilities/02_health_check.sql`](utilities/02_health_check.sql)
- **Maintenance:** [`utilities/03_maintenance.sql`](utilities/03_maintenance.sql)