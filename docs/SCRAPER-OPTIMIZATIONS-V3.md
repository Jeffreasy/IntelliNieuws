# Scraper Optimalisaties v3.0

## 📊 Analyse Resultaten

### Huidige Status
- ✅ Batch operations geïmplementeerd (ExistsByURLBatch, CreateBatch)
- ✅ Circuit breakers aanwezig
- ✅ Redis caching actief
- ✅ Connection pooling werkend
- ⚠️ Enkele inefficiënties geïdentificeerd

### Geïdentificeerde Optimalisaties

## 🎯 Optimalisatie Plan

### 1. Database Query Optimalisaties (HIGH IMPACT)

**Probleem**: `List()` en `Search()` selecteren altijd volledige content
```go
// Huidige query haalt ALTIJD full content op (kan MB's zijn)
SELECT ..., content, content_extracted, content_extracted_at FROM articles
```

**Impact**: 
- Grote responses (MB's data)
- Trage API calls
- Hoog geheugenverbruik

**Oplossing**: Lightweight queries voor lijst views
- Nieuwe method `ListLight()` - zonder content field
- Gebruik alleen `GetByID()` voor volledige content
- Besparing: ~90% bij 50 articles

### 2. Index Optimalisaties (HIGH IMPACT)

**Probleem**: Mogelijk ontbrekende indexes voor veelgebruikte queries

**Oplossing**: Composite indexes toevoegen
```sql
-- Voor content extraction queries
CREATE INDEX CONCURRENTLY idx_articles_content_extraction 
ON articles(content_extracted, created_at DESC) 
WHERE content_extracted = FALSE;

-- Voor published date sorting (meest frequent)
CREATE INDEX CONCURRENTLY idx_articles_published_desc 
ON articles(published DESC) 
WHERE published IS NOT NULL;

-- Voor source filtering
CREATE INDEX CONCURRENTLY idx_articles_source_published 
ON articles(source, published DESC);
```

**Impact**: 10-100x sneller queries

### 3. Browser Pool Optimalisaties (MEDIUM IMPACT)

**Probleem**: Polling-based acquire (100ms ticker)
```go
// Inefficiënt: polling every 100ms
for {
    select {
    case <-ticker.C:
        // check if browser available
    }
}
```

**Oplossing**: Channel-based signaling
```go
// Efficiënt: immediate notification
select {
case browser := <-p.available:
    return browser
case <-ctx.Done():
    return nil, ctx.Err()
}
```

**Impact**: 
- Geen 100ms delay meer
- Lagere CPU load
- Snellere response times

### 4. Content Extraction Cache (MEDIUM IMPACT)

**Probleem**: HTML wordt meerdere keren geparsed voor verschillende selectors

**Oplossing**: Parse één keer, cache goquery.Document
- Single parse per URL
- Reuse voor alle selectors
- Memory efficient met TTL

**Impact**: 3-5x sneller extraction

### 5. Rate Limiter Optimalisaties (LOW-MEDIUM IMPACT)

**Probleem**: Ook ticker-based (inefficiënt)

**Oplossing**: Token bucket met channels
- Immediate availability check
- Better burst handling
- Lower latency

### 6. Response Caching (HIGH IMPACT)

**Probleem**: Geen caching van lijst responses

**Oplossing**: Redis cache voor:
- Article lists (TTL: 2 min)
- Stats/categories (TTL: 5 min)
- Search results (TTL: 1 min)

**Impact**: 
- 95% cache hit ratio mogelijk
- Sub-ms response times
- Database load -80%

### 7. Concurrency Tuning (MEDIUM IMPACT)

**Huidige Settings**:
- Browser max concurrent: 2
- Scraper max concurrent: 3
- Pool size: 3

**Aanbevelingen**:
```env
# Voor betere throughput
BROWSER_MAX_CONCURRENT=3        # Was: 2
SCRAPER_MAX_CONCURRENT=5        # Was: 3
BROWSER_POOL_SIZE=5             # Was: 3
REDIS_POOL_SIZE=30              # Was: 20
```

### 8. Prepared Statements (LOW IMPACT)

**Oplossing**: Gebruik pgx prepared statements voor frequent queries
- `ExistsByURLBatch`
- `CreateBatch`
- `List` queries

## 📈 Verwachte Performance Verbetering

| Operatie | Voor | Na | Verbetering |
|----------|------|-----|-------------|
| Article List (50 items) | 250ms | 25ms | **10x** |
| Duplicate Check (50 URLs) | 50ms | 5ms | **10x** |
| Browser Acquire | 100-200ms | <10ms | **10-20x** |
| Content Extraction | 3-5s | 1-2s | **2-3x** |
| Cached Responses | 100ms | 2ms | **50x** |

### Totaal Verwacht:
- **70% sneller** voor scraping operations
- **90% sneller** voor API responses (met cache)
- **50% minder** database load
- **60% minder** CPU gebruik

## 🚀 Implementatie Volgorde

1. ✅ Database index optimalisaties (geen code changes)
2. ⬜ Lightweight query methods (backward compatible)
3. ⬜ Response caching layer
4. ⬜ Browser pool channel-based acquire
5. ⬜ Rate limiter improvements
6. ⬜ Content extraction cache
7. ⬜ Concurrency tuning (config only)

## ⚠️ Risico's & Mitigatie

| Risico | Impact | Mitigatie |
|--------|--------|-----------|
| Index creation locks | Medium | Use CONCURRENTLY |
| Cache invalidation bugs | High | Conservative TTLs |
| Memory usage increase | Low | Monitor, add limits |
| Breaking changes | Low | Keep old methods |

## 🧪 Testing Plan

1. Load testing met 1000 articles
2. Concurrent scraping stress test
3. Cache hit ratio monitoring
4. Memory profiling
5. Database slow query log

## 📊 Monitoring

Nieuwe metrics toevoegen:
- Cache hit/miss ratio
- Average query times
- Browser pool utilization
- Content extraction success rate
- Rate limit wait times

## 🔄 Rollback Plan

Alle changes zijn backward compatible:
1. Nieuwe methods, oude blijven bestaan
2. Indexes kunnen gedropped worden
3. Cache kan disabled worden via config
4. Config changes zijn reversible

## 💡 Future Optimalisaties (v3.1+)

1. **Distributed caching** met Redis Cluster
2. **Query result streaming** voor grote datasets
3. **Async content extraction** met message queue
4. **Smart rate limiting** based on response times
5. **Predictive browser pool scaling**
6. **GraphQL API** voor flexible queries
7. **Content CDN** voor images
8. **Full-text search** met Elasticsearch

## 📝 Notes

- Alle optimalisaties zijn **NON-BREAKING**
- Focus op **80/20 rule** - grootste impact eerst
- **Monitoring** is cruciaal voor validatie
- **Incremental rollout** per optimalisatie