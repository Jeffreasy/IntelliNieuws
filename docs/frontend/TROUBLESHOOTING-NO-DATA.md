# 🔧 Troubleshooting: Geen Data in Trending & Sentiment

**Last Updated:** 2025-10-30 14:30 CET  
**Status:** ✅ Database Fixed & Optimized

---

## 🆕 Recent Database Updates (2025-10-30)

**BELANGRIJK:** De database is recent geoptimaliseerd met:
- ✅ **3 Materialized Views** hersteld en werkend
- ✅ **Sources Metadata Tracking** geïmplementeerd
- ✅ **Dubbele Triggers** verwijderd
- ✅ **Analytics 90% sneller** (5s → 0.5s)

Deze updates zouden de meeste problemen moeten oplossen!

---

## Probleem
"Trending Now" en "Sentiment Analyse" tonen geen data in de frontend.

---

## ✅ STAP 1: Verify Database Health

### Check Materialized Views
```bash
# Via Docker
docker exec nieuws-scraper-postgres psql -U scraper -d nieuws_scraper \
  -c "SELECT matviewname, pg_size_pretty(pg_total_relation_size(schemaname||'.'||matviewname)) as size 
      FROM pg_matviews WHERE schemaname = 'public';"
```

**Verwacht Resultaat:**
```
matviewname           | size
----------------------+-------
mv_entity_mentions    | 168 kB
mv_sentiment_timeline | 112 kB
mv_trending_keywords  | 176 kB
```

**Als NIET alle 3 views aanwezig zijn:**
```bash
# Run de fix script
docker exec -i nieuws-scraper-postgres psql -U scraper -d nieuws_scraper \
  < scripts/migrations/fix-missing-materialized-views.sql
```

### Check Sources Metadata
```bash
docker exec nieuws-scraper-postgres psql -U scraper -d nieuws_scraper \
  -c "SELECT name, last_scraped_at, total_articles_scraped, consecutive_failures 
      FROM sources;"
```

**Verwacht:** `last_scraped_at` moet gevuld zijn (niet NULL)

**Als NULL:** Sources worden niet bijgewerkt → rebuild app:
```bash
docker-compose up -d --build app
```

---

## ✅ STAP 2: Test Backend API

### 2.1 Health Check
```bash
curl http://localhost:8080/health
```

**Verwacht:**
```json
{
  "status": "healthy",
  "timestamp": "2025-10-30T13:30:00Z",
  "services": {
    "database": "up",
    "redis": "up"
  }
}
```

**Als API niet bereikbaar:**
```bash
# Check Docker status
docker ps -a --filter "name=nieuws-scraper"

# Restart als nodig
docker-compose up -d
```

### 2.2 Test Trending Endpoint
```bash
curl "http://localhost:8080/api/v1/analytics/trending?limit=5"
```

**Verwacht:**
```json
{
  "trending": [
    {
      "keyword": "example",
      "article_count": 10,
      "trending_score": 85.5
    }
  ],
  "meta": {
    "count": 5
  }
}
```

### 2.3 Test Sentiment Endpoint
```bash
curl "http://localhost:8080/api/v1/analytics/sentiment-trends"
```

**Verwacht:**
```json
{
  "trends": [
    {
      "day": "2025-10-30",
      "source": "nu.nl",
      "total_articles": 50,
      "positive_count": 20,
      "negative_count": 10,
      "neutral_count": 20
    }
  ]
}
```

---

## ✅ STAP 3: Check Data Availability

### 3.1 Check Article Count
```bash
curl http://localhost:8080/api/v1/articles/stats
```

**Verwacht (na recent database fixes):**
```json
{
  "nu.nl": 143,
  "ad.nl": 125,
  "nos.nl": 51
}
```

**Als alles 0 is:**
```bash
# Trigger scraping (requires API key if authentication enabled)
curl -X POST http://localhost:8080/api/v1/scrape

# OR via Docker exec:
docker exec nieuws-scraper-app sh -c "echo 'Trigger scrape via internal endpoint'"
```

### 3.2 Check AI Processing Status
```bash
docker exec nieuws-scraper-postgres psql -U scraper -d nieuws_scraper \
  -c "SELECT COUNT(*) as total, 
      COUNT(*) FILTER (WHERE ai_processed = true) as ai_processed 
      FROM articles;"
```

**Verwacht:**
```
total | ai_processed
------+-------------
  319 | 319
```

**Als ai_processed is 0:**
```bash
# Check AI configuration
docker exec nieuws-scraper-app printenv | grep AI_ENABLED
# Should show: AI_ENABLED=true

docker exec nieuws-scraper-app printenv | grep OPENAI_API_KEY
# Should have a value (not empty)
```

---

## ✅ STAP 4: Refresh Analytics Views

### Method 1: Via API (Preferred)
```bash
# Refresh all materialized views
curl -X POST http://localhost:8080/api/v1/analytics/refresh
```

### Method 2: Direct Database
```bash
docker exec nieuws-scraper-postgres psql -U scraper -d nieuws_scraper \
  -c "SELECT * FROM refresh_analytics_views(FALSE);"
```

**Verwacht Output:**
```
view_name             | refresh_time_ms | rows_affected
----------------------+-----------------+--------------
mv_trending_keywords  | 119            | 88
mv_sentiment_timeline | 90             | 133
mv_entity_mentions    | 85             | 182
```

**Als errors:** Check [`docs/DATABASE-FIXES-COMPLETE.md`](../DATABASE-FIXES-COMPLETE.md:1) voor fixes

---

## ✅ STAP 5: Frontend Configuration

### 5.1 Create/Update `.env.local`

**In je frontend directory:**
```env
# .env.local
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_API_KEY=your-api-key-here
```

### 5.2 Restart Next.js Dev Server
```bash
# Stop current dev server (Ctrl+C)
npm run dev
# or
pnpm dev
```

### 5.3 Verify in Browser Console
```javascript
// Open DevTools (F12) → Console tab
console.log('API URL:', process.env.NEXT_PUBLIC_API_URL)
// Expected: http://localhost:8080

// Test connection
fetch('http://localhost:8080/health')
  .then(r => r.json())
  .then(data => console.log('✅ API Health:', data))
  .catch(err => console.error('❌ API Error:', err));
```

---

## 🧪 Manual Testing in Browser

### Test 1: Trending API
```javascript
// Open Browser DevTools (F12) → Console tab
fetch('http://localhost:8080/api/v1/analytics/trending?limit=5')
  .then(r => r.json())
  .then(data => {
    console.log('📊 Trending Data:', data);
    if (data.trending && data.trending.length > 0) {
      console.log(`✅ Working! Found ${data.trending.length} keywords`);
      console.table(data.trending);
    } else {
      console.log('⚠️ No trending data available');
      console.log('Check: 1) AI processed articles, 2) Views refreshed, 3) Time range');
    }
  })
  .catch(err => {
    console.error('❌ API Error:', err);
    console.log('Check: 1) Backend running, 2) CORS enabled, 3) Correct URL');
  });
```

### Test 2: Sentiment API
```javascript
fetch('http://localhost:8080/api/v1/analytics/sentiment-trends')
  .then(r => r.json())
  .then(data => {
    console.log('😊 Sentiment Data:', data);
    if (data.trends && data.trends.length > 0) {
      console.log(`✅ Working! Found ${data.trends.length} trend days`);
      console.table(data.trends.slice(0, 5));
    } else {
      console.log('⚠️ No sentiment data available');
      console.log('Reason: No AI-processed articles in last 30 days');
    }
  })
  .catch(err => console.error('❌ API Error:', err));
```

### Test 3: Check React Query State
```javascript
// If using @tanstack/react-query-devtools
// Look at bottom-right corner for DevTools button

// Or check cache manually:
// (Only works if you have access to queryClient)
const analytics = queryClient.getQueryData(['analytics', 'trending']);
console.log('📦 Cached Analytics:', analytics);
```

---

## 📋 Expected Data Structures

### Trending Response
```typescript
{
  "trending": [
    {
      "keyword": "bitcoin",
      "article_count": 10,
      "source_count": 3,
      "sources": ["nu.nl", "ad.nl", "nos.nl"],
      "avg_sentiment": 0.25,
      "avg_relevance": 0.85,
      "trending_score": 87.5,
      "most_recent": "2025-10-30T13:00:00Z"
    }
  ],
  "meta": {
    "hours": 24,
    "min_articles": 3,
    "limit": 20,
    "count": 10
  }
}
```

### Sentiment Response
```typescript
{
  "trends": [
    {
      "day": "2025-10-30",
      "source": "nu.nl",
      "total_articles": 50,
      "positive_count": 20,
      "neutral_count": 20,
      "negative_count": 10,
      "avg_sentiment": 0.15,
      "positive_percentage": 40.0,
      "negative_percentage": 20.0
    }
  ],
  "meta": {
    "source": null,
    "count": 7
  }
}
```

---

## 🚨 Common Issues & Solutions

### Issue 1: CORS Errors

**Symptom:**
```
Access to fetch at 'http://localhost:8080' from origin 'http://localhost:3000' 
has been blocked by CORS policy
```

**Solution:**
Backend moet CORS enabled hebben. Check [`internal/api/routes.go`](../../internal/api/routes.go:1)

### Issue 2: Empty Arrays

**Symptom:**
```json
{
  "trending": [],
  "meta": { "count": 0 }
}
```

**Possible Causes:**
1. Geen AI-processed artikelen
2. Materialized views leeg
3. Time range te klein (probeer `hours=168` voor 7 dagen)
4. Min articles threshold te hoog

**Quick Fix:**
```bash
# Verlaag threshold en vergroot time range
curl "http://localhost:8080/api/v1/analytics/trending?hours=168&min_articles=1&limit=50"
```

### Issue 3: Old/Stale Data

**Symptom:** Data is oud of outdated

**Solution:**
```bash
# Refresh materialized views
curl -X POST http://localhost:8080/api/v1/analytics/refresh

# In frontend: invalidate cache
// In React component:
const queryClient = useQueryClient();
queryClient.invalidateQueries({ queryKey: ['analytics'] });
```

### Issue 4: Network Tab Shows 401/403

**Symptom:** Authentication errors

**Solution:**
```typescript
// Check API key in .env.local
NEXT_PUBLIC_API_KEY=your-api-key

// Verify in code:
const headers = {
  'X-API-Key': process.env.NEXT_PUBLIC_API_KEY
};
```

---

## 🚀 All-in-One Fix Script

**Save as `fix-frontend-data.sh`:**
```bash
#!/bin/bash
set -e

echo "🔍 NieuwsScraper Frontend Data Fix Script"
echo "==========================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Step 1: Check Docker
echo -e "${BLUE}1️⃣ Checking Docker containers...${NC}"
if ! docker ps | grep -q nieuws-scraper-app; then
    echo -e "${RED}❌ App container not running${NC}"
    echo "   Starting containers..."
    docker-compose up -d
    sleep 15
else
    echo -e "${GREEN}✅ App container running${NC}"
fi

# Step 2: Check Database Views
echo ""
echo -e "${BLUE}2️⃣ Checking materialized views...${NC}"
VIEW_COUNT=$(docker exec nieuws-scraper-postgres psql -U scraper -d nieuws_scraper -t -c "SELECT COUNT(*) FROM pg_matviews WHERE schemaname = 'public';")
VIEW_COUNT=$(echo $VIEW_COUNT | xargs) # Trim whitespace

if [ "$VIEW_COUNT" -lt 3 ]; then
    echo -e "${RED}❌ Missing materialized views (found $VIEW_COUNT, need 3)${NC}"
    echo "   Running fix script..."
    docker exec -i nieuws-scraper-postgres psql -U scraper -d nieuws_scraper < scripts/migrations/fix-missing-materialized-views.sql
    echo -e "${GREEN}✅ Views fixed${NC}"
else
    echo -e "${GREEN}✅ All 3 materialized views present${NC}"
fi

# Step 3: Check Articles
echo ""
echo -e "${BLUE}3️⃣ Checking articles in database...${NC}"
ARTICLE_COUNT=$(docker exec nieuws-scraper-postgres psql -U scraper -d nieuws_scraper -t -c "SELECT COUNT(*) FROM articles;")
ARTICLE_COUNT=$(echo $ARTICLE_COUNT | xargs)
echo "   Found $ARTICLE_COUNT articles"

if [ "$ARTICLE_COUNT" -lt 10 ]; then
    echo -e "${YELLOW}⚠️  Low article count. Consider running scraper.${NC}"
fi

# Step 4: Check AI Processing
echo ""
echo -e "${BLUE}4️⃣ Checking AI processing...${NC}"
AI_COUNT=$(docker exec nieuws-scraper-postgres psql -U scraper -d nieuws_scraper -t -c "SELECT COUNT(*) FROM articles WHERE ai_processed = true;")
AI_COUNT=$(echo $AI_COUNT | xargs)
echo "   Found $AI_COUNT AI-processed articles"

if [ "$AI_COUNT" -eq 0 ]; then
    echo -e "${YELLOW}⚠️  No AI-processed articles. Check if AI_ENABLED=true and OPENAI_API_KEY is set.${NC}"
fi

# Step 5: Refresh Analytics
echo ""
echo -e "${BLUE}5️⃣ Refreshing materialized views...${NC}"
docker exec nieuws-scraper-postgres psql -U scraper -d nieuws_scraper \
  -c "SELECT view_name, refresh_time_ms, rows_affected FROM refresh_analytics_views(FALSE);"
echo -e "${GREEN}✅ Views refreshed${NC}"

# Step 6: Test Endpoints
echo ""
echo -e "${BLUE}6️⃣ Testing API endpoints...${NC}"

# Test health
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Health endpoint working${NC}"
else
    echo -e "${RED}❌ Health endpoint failed${NC}"
fi

# Test trending
TRENDING_RESPONSE=$(curl -s "http://localhost:8080/api/v1/analytics/trending?limit=5")
TRENDING_COUNT=$(echo $TRENDING_RESPONSE | jq -r '.trending | length' 2>/dev/null || echo "0")
echo "   Trending keywords available: $TRENDING_COUNT"

# Test sentiment
SENTIMENT_RESPONSE=$(curl -s "http://localhost:8080/api/v1/analytics/sentiment-trends")
SENTIMENT_COUNT=$(echo $SENTIMENT_RESPONSE | jq -r '.trends | length' 2>/dev/null || echo "0")
echo "   Sentiment trends available: $SENTIMENT_COUNT"

# Summary
echo ""
echo "=========================================="
echo -e "${BLUE}📊 Summary:${NC}"
echo "   Articles: $ARTICLE_COUNT"
echo "   AI Processed: $AI_COUNT"
echo "   Materialized Views: $VIEW_COUNT"
echo "   Trending Keywords: $TRENDING_COUNT"
echo "   Sentiment Trends: $SENTIMENT_COUNT"
echo ""

if [ "$TRENDING_COUNT" -gt 0 ] && [ "$SENTIMENT_COUNT" -gt 0 ]; then
    echo -e "${GREEN}🎉 Everything looks good! Refresh your frontend browser.${NC}"
    echo ""
    echo "Sample Trending Data:"
    echo $TRENDING_RESPONSE | jq -r '.trending[0:3] | .[] | "  - \(.keyword) (\(.article_count) articles, score: \(.trending_score))"'
else
    echo -e "${YELLOW}⚠️  Some endpoints have no data.${NC}"
    echo ""
    echo "Troubleshooting:"
    if [ "$ARTICLE_COUNT" -eq 0 ]; then
        echo "  1. No articles in database - run scraper"
    fi
    if [ "$AI_COUNT" -eq 0 ]; then
        echo "  2. No AI processing - check AI_ENABLED and OPENAI_API_KEY"
    fi
    if [ "$VIEW_COUNT" -lt 3 ]; then
        echo "  3. Missing views - run fix script again"
    fi
fi

echo ""
echo "🔄 Manual checks:"
echo "   curl http://localhost:8080/api/v1/analytics/trending?limit=3"
echo "   curl http://localhost:8080/api/v1/analytics/sentiment-trends"
echo ""
```

**Make executable and run:**
```bash
chmod +x fix-frontend-data.sh
./fix-frontend-data.sh
```

---

## 🧪 Browser Console Testing

### Complete Test Suite
```javascript
// Open Browser DevTools (F12) → Console tab
// Paste this complete test:

console.log('🧪 Starting API Test Suite...\n');

const API_URL = 'http://localhost:8080';
const tests = [];

// Test 1: Health
tests.push(
  fetch(`${API_URL}/health`)
    .then(r => r.json())
    .then(data => {
      console.log('✅ Health:', data.status);
      return { test: 'health', passed: data.status === 'healthy' };
    })
    .catch(err => {
      console.error('❌ Health failed:', err.message);
      return { test: 'health', passed: false, error: err.message };
    })
);

// Test 2: Articles Stats
tests.push(
  fetch(`${API_URL}/api/v1/articles/stats`)
    .then(r => r.json())
    .then(data => {
      const total = Object.values(data).reduce((a, b) => a + b, 0);
      console.log(`✅ Articles: ${total} total`);
      return { test: 'articles', passed: total > 0, count: total };
    })
    .catch(err => {
      console.error('❌ Articles failed:', err.message);
      return { test: 'articles', passed: false, error: err.message };
    })
);

// Test 3: Trending
tests.push(
  fetch(`${API_URL}/api/v1/analytics/trending?limit=5`)
    .then(r => r.json())
    .then(data => {
      const count = data.trending?.length || 0;
      console.log(`${count > 0 ? '✅' : '⚠️'} Trending: ${count} keywords`);
      if (count > 0) {
        console.table(data.trending.slice(0, 3));
      }
      return { test: 'trending', passed: count > 0, count };
    })
    .catch(err => {
      console.error('❌ Trending failed:', err.message);
      return { test: 'trending', passed: false, error: err.message };
    })
);

// Test 4: Sentiment
tests.push(
  fetch(`${API_URL}/api/v1/analytics/sentiment-trends`)
    .then(r => r.json())
    .then(data => {
      const count = data.trends?.length || 0;
      console.log(`${count > 0 ? '✅' : '⚠️'} Sentiment: ${count} days`);
      if (count > 0) {
        console.table(data.trends.slice(0, 3));
      }
      return { test: 'sentiment', passed: count > 0, count };
    })
    .catch(err => {
      console.error('❌ Sentiment failed:', err.message);
      return { test: 'sentiment', passed: false, error: err.message };
    })
);

// Wait for all tests
Promise.all(tests).then(results => {
  console.log('\n📊 Test Results:');
  console.table(results);
  
  const allPassed = results.every(r => r.passed);
  if (allPassed) {
    console.log('\n🎉 All tests passed! Frontend should show data.');
  } else {
    console.log('\n⚠️ Some tests failed. Check errors above.');
  }
});
```

---

## 🔍 Diagnostics Checklist

### Backend
- [ ] Docker containers running (`docker ps`)
- [ ] App container healthy (`docker ps` shows "healthy")
- [ ] Health endpoint returns 200 (`curl http://localhost:8080/health`)
- [ ] Logs show no errors (`docker logs nieuws-scraper-app`)

### Database
- [ ] 3 materialized views exist
- [ ] Articles table has data (> 100 rows)
- [ ] AI processed articles exist (ai_processed = true)
- [ ] Sources metadata is populated (last_scraped_at not NULL)
- [ ] No constraint violations

### Frontend
- [ ] `.env.local` exists with `NEXT_PUBLIC_API_URL`
- [ ] Dev server restarted after .env change
- [ ] No CORS errors in browser console
- [ ] Network tab shows 200 responses
- [ ] React Query not showing errors

---

## 🎯 Quick Wins

### 1. Restart Everything
```bash
# Stop all
docker-compose down

# Start fresh
docker-compose up -d

# Wait for healthy
docker ps

# Restart frontend
cd frontend && npm run dev
```

### 2. Clear All Caches
```bash
# Browser cache: Ctrl+Shift+Del
# React Query cache: Refresh page with Ctrl+Shift+R
# Redis cache:
docker exec nieuws-scraper-redis redis-cli -a redis_password FLUSHDB
```

### 3. Force Refresh Data
```bash
# 1. Refresh materialized views
curl -X POST http://localhost:8080/api/v1/analytics/refresh

# 2. Clear browser cache (Ctrl+Shift+Del)

# 3. Hard refresh page (Ctrl+Shift+R)
```

---

## 🆘 Still Not Working?

### Deep Diagnostics

#### 1. Check Backend Logs
```bash
# Follow logs
docker logs -f nieuws-scraper-app

# Look for:
# - Database connection errors
# - AI processing errors
# - Analytics query errors
# - Rate limiting warnings
```

#### 2. Verify Database State
```sql
-- Connect to database
docker exec nieuws-scraper-postgres psql -U scraper -d nieuws_scraper

-- Check article distribution
SELECT 
    source,
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE ai_processed = true) as ai_processed,
    MAX(published) as latest
FROM articles 
GROUP BY source;

-- Check materialized view content
SELECT COUNT(*) as keywords FROM mv_trending_keywords;
SELECT COUNT(*) as sentiments FROM mv_sentiment_timeline;
SELECT COUNT(*) as entities FROM mv_entity_mentions;

-- If counts are 0, refresh:
SELECT * FROM refresh_analytics_views(FALSE);
```

#### 3. Test Direct Database Queries
```sql
-- Get trending keywords directly
SELECT keyword, article_count, trending_score 
FROM mv_trending_keywords 
WHERE hour_bucket >= CURRENT_TIMESTAMP - INTERVAL '24 hours'
ORDER BY trending_score DESC 
LIMIT 5;

-- Get sentiment trends directly
SELECT day, source, total_articles, avg_sentiment
FROM v_sentiment_trends_7d
ORDER BY day DESC
LIMIT 5;
```

**Als deze queries data tonen:** Backend API heeft een probleem  
**Als deze queries GEEN data tonen:** Database refresh nodig

---

## 🔧 Advanced Troubleshooting

### Enable Debug Logging

#### Backend (Go)
```bash
# Update environment variable
docker-compose down
docker-compose up -d -e LOG_LEVEL=debug

# Watch logs
docker logs -f nieuws-scraper-app
```

#### Frontend (React Query)
```typescript
// lib/queryClient.ts
import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60 * 1000,
      onError: (error) => {
        console.error('React Query Error:', error);
      },
      onSuccess: (data) => {
        console.log('React Query Success:', data);
      },
    },
  },
  logger: {
    log: console.log,
    warn: console.warn,
    error: console.error,
  },
});
```

### Network Inspection

```javascript
// In browser console, intercept fetch:
const originalFetch = window.fetch;
window.fetch = async (...args) => {
  console.log('🌐 Fetch:', args[0]);
  const response = await originalFetch(...args);
  console.log('📥 Response:', response.status, response.statusText);
  return response;
};

// Now reload page and watch console
```

---

## ✅ Success Indicators

You'll know it's working when:

### Backend
- ✅ `curl http://localhost:8080/health` returns `"status": "healthy"`
- ✅ `docker logs nieuws-scraper-app` shows no errors
- ✅ Database has 3 materialized views
- ✅ Sources table shows `last_scraped_at` timestamps

### Frontend
- ✅ "Trending Now" shows list of keywords
- ✅ "Sentiment Analyse" shows charts with data
- ✅ Browser console has NO errors
- ✅ Network tab shows 200 OK responses
- ✅ React Query DevTools shows successful queries

### Data Flow
```
Scraper → Articles → AI Processing → Materialized Views → API → Frontend
   ✅        ✅           ✅                ✅              ✅       ✅
```

---

## 📚 Related Documentation

- [`docs/DATABASE-FIXES-COMPLETE.md`](../DATABASE-FIXES-COMPLETE.md) - Recent database fixes
- [`docs/DATABASE-DOCKER-ANALYSIS.md`](../DATABASE-DOCKER-ANALYSIS.md) - Complete database analysis
- [`docs/frontend/COMPLETE-API-REFERENCE.md`](COMPLETE-API-REFERENCE.md) - Updated API docs
- [`scripts/migrations/fix-missing-materialized-views.sql`](../../scripts/migrations/fix-missing-materialized-views.sql) - Fix script

---

## 🆘 Emergency Reset

**Als ALLES faalt, complete reset:**

```bash
#!/bin/bash
echo "⚠️  EMERGENCY RESET - This will DELETE all data!"
read -p "Are you sure? (yes/no): " confirm

if [ "$confirm" = "yes" ]; then
    # Stop everything
    docker-compose down -v
    
    # Remove volumes (DELETES DATA!)
    docker volume rm nieuws-scraper_postgres_data
    docker volume rm nieuws-scraper_redis_data
    
    # Start fresh
    docker-compose up -d
    
    # Wait for database
    sleep 20
    
    # Migrations run automatically via docker-entrypoint-initdb.d
    
    # Fix missing views
    docker exec -i nieuws-scraper-postgres psql -U scraper -d nieuws_scraper \
      < scripts/migrations/fix-missing-materialized-views.sql
    
    echo "✅ Reset complete. Run scraper to populate data."
fi
```

**Then populate with data:**
```bash
# Trigger scraping (wait 2-3 minutes)
curl -X POST http://localhost:8080/api/v1/scrape

# Refresh analytics
curl -X POST http://localhost:8080/api/v1/analytics/refresh

# Test
curl "http://localhost:8080/api/v1/analytics/trending?limit=3"
```

---

## 💡 Pro Tips

### 1. Use React Query DevTools
```bash
npm install @tanstack/react-query-devtools
```

```typescript
// app/providers.tsx
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';

export function Providers({ children }) {
  return (
    <QueryClientProvider client={queryClient}>
      {children}
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
}
```

### 2. Add Loading States
```typescript
const { data, isLoading, isError, error } = useTrendingKeywords();

if (isLoading) return <div>Loading trending keywords...</div>;
if (isError) return <div>Error: {error.message}</div>;
if (!data || data.trending.length === 0) {
  return <div>No trending data available. Try refreshing later.</div>;
}
```

### 3. Automatic Refresh
```typescript
export function useTrendingKeywords(hours = 24) {
  return useQuery({
    queryKey: ['analytics', 'trending', hours],
    queryFn: () => apiClient.getTrendingKeywords(hours),
    refetchInterval: 5 * 60 * 1000, // Auto-refresh every 5 minutes
    staleTime: 2 * 60 * 1000,       // Consider stale after 2 minutes
  });
}
```

---

## 📞 Still Need Help?

### Check These Files
1. Backend logs: `docker logs nieuws-scraper-app`
2. Database logs: `docker logs nieuws-scraper-postgres`
3. Redis logs: `docker logs nieuws-scraper-redis`

### Run Diagnostics
```bash
# Complete diagnostic output
./fix-frontend-data.sh > diagnostic-report.txt 2>&1

# Check report
cat diagnostic-report.txt
```

### Verify Each Layer
```
Frontend (Next.js) → API (Go) → Redis Cache → PostgreSQL
     ↓                  ↓            ↓            ↓
  .env.local      routes.go    6379 port     5432 port
  localhost:3000  localhost:8080              
```

**Test each layer individually:**
```bash
# Layer 1: PostgreSQL
docker exec nieuws-scraper-postgres psql -U scraper -d nieuws_scraper -c "SELECT COUNT(*) FROM articles;"

# Layer 2: Redis
docker exec nieuws-scraper-redis redis-cli -a redis_password PING

# Layer 3: API
curl http://localhost:8080/health

# Layer 4: Frontend
# Open http://localhost:3000 in browser
```

---

## 🎉 Success Checklist

Before considering it "fixed", verify:

- [ ] Backend health endpoint returns healthy
- [ ] Database has 3 materialized views with data
- [ ] Articles table has > 100 rows
- [ ] At least 50% articles are AI-processed
- [ ] Sources metadata shows last_scraped_at timestamps
- [ ] Trending API returns > 0 keywords
- [ ] Sentiment API returns > 0 trends
- [ ] Browser console shows NO errors
- [ ] Network tab shows all 200 OK
- [ ] React Query shows successful queries
- [ ] Frontend displays data correctly

---

**Last Verified:** 2025-10-30 14:30 CET  
**Database Version:** V003 (with fixes)  
**App Version:** Latest (with sources metadata tracking)