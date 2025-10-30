# Migration Guide - NieuwsScraper Professional Schema

## 🎯 Quick Start

### Voor Nieuwe Installaties

```bash
# 1. Apply all migrations in order
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper < migrations/V001__create_base_schema.sql
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper < migrations/V002__create_emails_table.sql
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper < migrations/V003__create_analytics_views.sql

# 2. Verify installation
docker exec nieuws-scraper-db psql -U postgres -d nieuws_scraper -c "SELECT * FROM schema_migrations;"
```

### Voor Bestaande Installaties (Legacy Schema)

```bash
# 1. Backup your database first!
docker exec nieuws-scraper-db pg_dump -U postgres nieuws_scraper > backup_$(date +%Y%m%d).sql

# 2. Run legacy migration utility
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper < migrations/utilities/01_migrate_from_legacy.sql

# 3. Apply V003 for analytics
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper < migrations/V003__create_analytics_views.sql

# 4. Run health check
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper < migrations/utilities/02_health_check.sql
```

### Gebruik PowerShell Script (Aanbevolen)

```powershell
# Voor nieuwe installaties
.\scripts\migrations\apply-new-migrations.ps1

# Voor bestaande installaties
.\scripts\migrations\apply-new-migrations.ps1 -SkipLegacyMigration:$false

# Dry run (test zonder wijzigingen)
.\scripts\migrations\apply-new-migrations.ps1 -DryRun
```

## 📋 Pre-Migration Checklist

- [ ] **Backup maken** van de database
- [ ] **Test migrations** op een kopie van productie data
- [ ] **Review changes** in alle migration files
- [ ] **Plan downtime** (indien nodig)
- [ ] **Notify users** (indien applicable)
- [ ] **Check disk space** (voor nieuwe indexes)
- [ ] **Verify dependencies** (PostgreSQL versie, extensions)

## 🔄 Migration Scenarios

### Scenario 1: Nieuwe Installatie (Clean Database)

**Situatie:** Je start met een lege database.

**Steps:**
1. Run V001 → Base schema
2. Run V002 → Emails table
3. Run V003 → Analytics views
4. Verify with health check
5. Start application

**Verwachte tijd:** 2-5 minuten

---

### Scenario 2: Upgrade van Legacy (001-008)

**Situatie:** Je hebt oude migrations (001-008) al toegepast.

**Steps:**
1. **Backup database**
2. Run `01_migrate_from_legacy.sql` → Adds missing columns and features
3. Run V003 → Analytics views
4. Run health check
5. Test application thoroughly
6. Update application code (if needed)

**Verwachte tijd:** 5-10 minuten

**Downtime:** Minimal (2-3 minutes for large datasets)

---

### Scenario 3: Partial Migration (V001/V002 Only)

**Situatie:** Je wil alleen basis schema zonder analytics.

**Steps:**
1. Run V001 → Base schema
2. Run V002 → Emails table
3. Skip V003 (can be added later)
4. Verify with health check

**Verwachte tijd:** 2-3 minuten

---

### Scenario 4: Adding Analytics to Existing Installation

**Situatie:** Je hebt V001/V002 maar wil analytics toevoegen.

**Steps:**
1. Verify V001/V002 are applied: `SELECT * FROM schema_migrations;`
2. Run V003 → Analytics views
3. Wait for initial refresh (can take 1-5 minutes)
4. Set up periodic refresh schedule

**Verwachte tijd:** 3-7 minuten

## 🚨 Rollback Procedures

### Full Rollback (Emergency)

```bash
# Rollback in reverse order
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper < migrations/rollback/V003__rollback.sql
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper < migrations/rollback/V002__rollback.sql
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper < migrations/rollback/V001__rollback.sql

# Restore from backup if needed
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper < backup_YYYYMMDD.sql
```

### Partial Rollback

```bash
# Only rollback V003 (keeps data)
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper < migrations/rollback/V003__rollback.sql
```

## ⚙️ Configuration Changes

### PostgreSQL Settings (Recommended)

Add to `postgresql.conf` or Docker environment:

```ini
# Memory
shared_buffers = 256MB              # 25% of RAM
effective_cache_size = 1GB          # 50-75% of RAM
work_mem = 16MB                     # Per operation
maintenance_work_mem = 128MB        # For VACUUM, CREATE INDEX

# Query Planner
random_page_cost = 1.1              # For SSD storage
effective_io_concurrency = 200      # For SSD storage

# Parallelism
max_parallel_workers_per_gather = 4
max_parallel_workers = 8

# WAL
wal_buffers = 16MB
checkpoint_completion_target = 0.9

# Autovacuum (Important!)
autovacuum = on
autovacuum_max_workers = 3
autovacuum_naptime = 20s
```

### Docker Compose Updates

```yaml
services:
  db:
    environment:
      - POSTGRES_SHARED_PRELOAD_LIBRARIES=pg_stat_statements
      - POSTGRES_SHARED_BUFFERS=256MB
      - POSTGRES_EFFECTIVE_CACHE_SIZE=1GB
      - POSTGRES_WORK_MEM=16MB
      - POSTGRES_MAINTENANCE_WORK_MEM=128MB
    shm_size: 256mb  # For shared memory
```

## 📊 Post-Migration Verification

### 1. Check Migration Status

```sql
SELECT * FROM schema_migrations ORDER BY version;
```

**Expected Output:**
- V001: Base schema
- V002: Emails table
- V003: Analytics views (optional)

### 2. Verify Table Counts

```sql
SELECT 
    'articles' AS table_name, COUNT(*) AS rows FROM articles
UNION ALL
SELECT 'sources', COUNT(*) FROM sources
UNION ALL
SELECT 'scraping_jobs', COUNT(*) FROM scraping_jobs
UNION ALL
SELECT 'emails', COUNT(*) FROM emails;
```

### 3. Test Key Features

```sql
-- Test full-text search
SELECT id, title FROM articles 
WHERE to_tsvector('english', title) @@ to_tsquery('bitcoin')
LIMIT 5;

-- Test AI columns
SELECT COUNT(*) FROM articles WHERE ai_processed = TRUE;

-- Test materialized views (if V003 applied)
SELECT * FROM v_trending_keywords_24h LIMIT 10;

-- Test email processing
SELECT status, COUNT(*) FROM emails GROUP BY status;
```

### 4. Check Index Usage

```sql
SELECT 
    schemaname, tablename, indexname, idx_scan
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan DESC
LIMIT 10;
```

### 5. Run Health Check

```bash
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper < migrations/utilities/02_health_check.sql
```

## 🔧 Troubleshooting

### Problem: Migration Fails with "relation already exists"

**Solution:**
```sql
-- Check what exists
SELECT tablename FROM pg_tables WHERE schemaname = 'public';

-- Option 1: Drop conflicting table (if safe)
DROP TABLE IF EXISTS table_name CASCADE;

-- Option 2: Use legacy migration script
-- This will handle existing tables gracefully
```

### Problem: "Permission denied" errors

**Solution:**
```sql
-- Grant necessary permissions
GRANT ALL ON SCHEMA public TO your_user;
GRANT ALL ON ALL TABLES IN SCHEMA public TO your_user;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO your_user;
```

### Problem: Materialized views too slow to refresh

**Solution:**
```sql
-- Reduce date range in views
-- Or refresh less frequently

-- Check current size
SELECT 
    matviewname, 
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||matviewname)) 
FROM pg_matviews;

-- Consider partitioning for very large datasets
```

### Problem: Index creation takes too long

**Solution:**
```sql
-- Use CONCURRENTLY to allow concurrent operations
CREATE INDEX CONCURRENTLY idx_name ON table_name(column);

-- Or increase maintenance_work_mem temporarily
SET maintenance_work_mem = '512MB';
CREATE INDEX idx_name ON table_name(column);
RESET maintenance_work_mem;
```

### Problem: Application errors after migration

**Checklist:**
1. Verify all migrations applied: `SELECT * FROM schema_migrations;`
2. Check for missing columns: Run health check
3. Verify triggers are active: `SELECT * FROM pg_trigger WHERE tgname LIKE 'trg_%';`
4. Test queries manually in psql
5. Check application logs for specific SQL errors
6. Verify connection string and credentials

## 📈 Performance Optimization Tips

### 1. Index Maintenance

```sql
-- Find unused indexes
SELECT schemaname, tablename, indexname, idx_scan
FROM pg_stat_user_indexes
WHERE idx_scan = 0 AND schemaname = 'public';

-- Reindex if bloated
REINDEX INDEX CONCURRENTLY index_name;
REINDEX TABLE CONCURRENTLY table_name;
```

### 2. Query Optimization

```sql
-- Use EXPLAIN ANALYZE to optimize queries
EXPLAIN ANALYZE
SELECT * FROM articles WHERE source = 'nu.nl' ORDER BY published DESC LIMIT 10;

-- Check slow queries
SELECT query, mean_exec_time, calls
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;
```

### 3. Materialized View Refresh Strategy

```sql
-- Refresh during low-traffic periods
-- Use CONCURRENTLY to avoid blocking

-- Schedule with pg_cron (if available)
SELECT cron.schedule('refresh-analytics', '*/15 * * * *', 
    'SELECT refresh_analytics_views(TRUE);');
```

### 4. Table Partitioning (Future Enhancement)

For very large article tables (>10M rows), consider partitioning:

```sql
-- Example: Partition articles by month
CREATE TABLE articles_2024_01 PARTITION OF articles
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

## 🔄 Maintenance Schedule

### Daily
- ✅ Run health check script
- ✅ Monitor disk space
- ✅ Check for failed scraping jobs

### Weekly
- ✅ Run maintenance script (`03_maintenance.sql`)
- ✅ Review slow queries
- ✅ Clean up old data (emails, jobs)

### Monthly
- ✅ Full VACUUM ANALYZE
- ✅ Review and optimize indexes
- ✅ Update table statistics
- ✅ Backup validation

## 📞 Support

### Getting Help

1. **Documentation**: Check [`README.md`](README.md) first
2. **Health Check**: Run `02_health_check.sql` for diagnostics
3. **Logs**: Check PostgreSQL logs for errors
4. **Community**: Search for similar issues

### Useful Commands

```bash
# Check PostgreSQL logs
docker logs nieuws-scraper-db --tail 100

# Connect to database
docker exec -it nieuws-scraper-db psql -U postgres -d nieuws_scraper

# Backup database
docker exec nieuws-scraper-db pg_dump -U postgres nieuws_scraper > backup.sql

# Restore database
docker exec -i nieuws-scraper-db psql -U postgres -d nieuws_scraper < backup.sql
```

## 📚 Additional Resources

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [PostgreSQL Performance Tips](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [Index Usage Guide](https://www.postgresql.org/docs/current/indexes.html)
- [Migration Best Practices](https://www.postgresql.org/docs/current/ddl-alter.html)

## 🎓 Migration History

| Version | Date | Description | Breaking Changes |
|---------|------|-------------|------------------|
| V003 | 2025-10-30 | Analytics materialized views | None |
| V002 | 2025-10-30 | Email integration table | None |
| V001 | 2025-10-30 | Base schema consolidation | Supersedes 001-008 |
| Legacy | Pre-2025 | Original migrations (deprecated) | See legacy docs |

## ✅ Success Criteria

Migration is successful when:

- [x] All schema_migrations records present
- [x] Health check shows no errors
- [x] Application connects successfully
- [x] All tables have expected row counts
- [x] Indexes are being used (check pg_stat_user_indexes)
- [x] Materialized views refresh successfully (if V003 applied)
- [x] No performance degradation
- [x] Backup can be restored successfully

## 🚀 Next Steps After Migration

1. **Update Application Code**
   - Verify all queries work with new schema
   - Test all features thoroughly
   - Update any hardcoded table/column names

2. **Set Up Monitoring**
   - Configure alerts for failed scraping jobs
   - Monitor materialized view freshness
   - Track query performance

3. **Optimize Performance**
   - Review query execution plans
   - Adjust PostgreSQL settings
   - Set up periodic maintenance

4. **Document Changes**
   - Update API documentation
   - Note any behavioral changes
   - Train team on new features

5. **Plan Future Enhancements**
   - Table partitioning for large datasets
   - Additional indexes based on query patterns
   - New analytics views based on usage