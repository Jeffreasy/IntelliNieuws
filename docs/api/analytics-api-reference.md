# Analytics API Reference

## 📊 Overview

De Analytics API biedt toegang tot real-time analytics, trending topics, sentiment analysis, en entity tracking. Deze endpoints gebruiken materialized views voor optimale performance (90% sneller dan dynamische queries).

**Base URL:** `http://localhost:8080/api/v1/analytics`

**Authentication:** Public (no API key required)

**Rate Limiting:** Ja (standaard limits van toepassing)

## 🔥 Trending Keywords

### Get Trending Keywords

Retourneert trending keywords op basis van article mentions, source diversity en recency.

**Endpoint:** `GET /analytics/trending`

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `hours` | integer | 24 | Time window in hours (1-168) |
| `min_articles` | integer | 3 | Minimum articles required |
| `limit` | integer | 20 | Max results (1-100) |

**Example Request:**
```bash
curl "http://localhost:8080/api/v1/analytics/trending?hours=24&min_articles=3&limit=10"
```

**Example Response:**
```json
{
  "trending": [
    {
      "keyword": "bitcoin",
      "article_count": 15,
      "source_count": 3,
      "sources": ["nu.nl", "ad.nl", "nos.nl"],
      "avg_sentiment": 0.45,
      "avg_relevance": 0.87,
      "most_recent": "2025-10-30T02:55:00Z",
      "trending_score": 42.5
    }
  ],
  "meta": {
    "hours": 24,
    "min_articles": 3,
    "limit": 10,
    "count": 15
  }
}
```

**Response Fields:**
- `keyword` - The trending keyword
- `article_count` - Number of articles mentioning it
- `source_count` - Number of unique sources
- `sources` - Array of source domains
- `avg_sentiment` - Average sentiment score (-1.0 to 1.0)
- `avg_relevance` - Average relevance score (0.0 to 1.0)
- `most_recent` - Timestamp of latest mention
- `trending_score` - Calculated trending score (higher = more trending)

**Use Cases:**
- Display trending topics on homepage
- Content recommendation
- Topic monitoring
- Market sentiment tracking

---

## 📈 Sentiment Trends

### Get Sentiment Trends

Retourneert daily sentiment trends over de laatste 7 dagen, gegroepeerd per source.

**Endpoint:** `GET /analytics/sentiment-trends`

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `source` | string | (all) | Filter by source domain |

**Example Request:**
```bash
curl "http://localhost:8080/api/v1/analytics/sentiment-trends?source=nu.nl"
```

**Example Response:**
```json
{
  "trends": [
    {
      "day": "2025-10-30",
      "source": "nu.nl",
      "total_articles": 45,
      "positive_count": 20,
      "neutral_count": 15,
      "negative_count": 10,
      "avg_sentiment": 0.25,
      "positive_percentage": 44.4,
      "negative_percentage": 22.2
    }
  ],
  "meta": {
    "source": "nu.nl",
    "count": 7
  }
}
```

**Response Fields:**
- `day` - Date (YYYY-MM-DD)
- `source` - Source domain
- `total_articles` - Total articles that day
- `positive_count` - Number of positive articles
- `neutral_count` - Number of neutral articles
- `negative_count` - Number of negative articles
- `avg_sentiment` - Average sentiment score
- `positive_percentage` - Percentage positive (0-100)
- `negative_percentage` - Percentage negative (0-100)

**Use Cases:**
- Sentiment timeline charts
- Source comparison
- Market mood tracking
- News bias analysis

---

## 👥 Hot Entities

### Get Most Mentioned Entities

Retourneert de meest genoemde entities (persons, organizations, locations) over de laatste 7 dagen.

**Endpoint:** `GET /analytics/hot-entities`

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `entity_type` | string | (all) | Filter by type: "person", "organization", "location" |
| `limit` | integer | 50 | Max results (1-100) |

**Example Request:**
```bash
curl "http://localhost:8080/api/v1/analytics/hot-entities?entity_type=person&limit=10"
```

**Example Response:**
```json
{
  "entities": [
    {
      "entity": "Elon Musk",
      "entity_type": "person",
      "total_mentions": 45,
      "days_mentioned": 6,
      "sources": ["nu.nl", "ad.nl", "nos.nl"],
      "overall_sentiment": 0.32,
      "most_recent_mention": "2025-10-30T02:50:00Z"
    }
  ],
  "meta": {
    "entity_type": "person",
    "limit": 10,
    "count": 25
  }
}
```

**Response Fields:**
- `entity` - Entity name
- `entity_type` - Type: person, organization, location
- `total_mentions` - Total mentions across all articles
- `days_mentioned` - Number of days mentioned
- `sources` - Array of sources mentioning entity
- `overall_sentiment` - Average sentiment for this entity
- `most_recent_mention` - Latest mention timestamp

**Use Cases:**
- Who's in the news
- Entity tracking
- Person/company monitoring
- Influence analysis

---

## 🎭 Entity Sentiment Analysis

### Get Entity Sentiment Timeline

Retourneert sentiment analysis voor een specifiek entity over tijd.

**Endpoint:** `GET /analytics/entity-sentiment`

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `entity` | string | ✅ Yes | Entity name to analyze |
| `days` | integer | No (30) | Days to look back (1-365) |

**Example Request:**
```bash
curl "http://localhost:8080/api/v1/analytics/entity-sentiment?entity=Elon%20Musk&days=30"
```

**Example Response:**
```json
{
  "entity": "Elon Musk",
  "timeline": [
    {
      "day": "2025-10-30",
      "mention_count": 5,
      "avg_sentiment": 0.45,
      "sources": ["nu.nl", "ad.nl"],
      "categories": ["tech", "business"]
    },
    {
      "day": "2025-10-29",
      "mention_count": 3,
      "avg_sentiment": -0.12,
      "sources": ["nos.nl"],
      "categories": ["tech"]
    }
  ],
  "meta": {
    "days": 30,
    "count": 15
  }
}
```

**Response Fields:**
- `day` - Date (YYYY-MM-DD)
- `mention_count` - Number of mentions that day
- `avg_sentiment` - Average sentiment score
- `sources` - Sources mentioning entity
- `categories` - Article categories

**Use Cases:**
- Person/company reputation tracking
- Sentiment timeline visualization
- PR monitoring
- Brand analysis

---

## 📋 Analytics Overview

### Get Comprehensive Overview

Retourneert een complete analytics overview met trending keywords, hot entities en materialized view status.

**Endpoint:** `GET /analytics/overview`

**Example Request:**
```bash
curl "http://localhost:8080/api/v1/analytics/overview"
```

**Example Response:**
```json
{
  "trending_keywords": [
    {
      "keyword": "bitcoin",
      "article_count": 15,
      "sources": ["nu.nl", "ad.nl"],
      "avg_sentiment": 0.45,
      "most_recent": "2025-10-30T02:55:00Z",
      "trending_score": 42.5
    }
  ],
  "hot_entities": [
    {
      "entity": "Elon Musk",
      "entity_type": "person",
      "total_mentions": 45,
      "days_mentioned": 6,
      "sources": ["nu.nl", "ad.nl", "nos.nl"],
      "overall_sentiment": 0.32
    }
  ],
  "materialized_views": [
    {
      "name": "mv_trending_keywords",
      "size": "136 kB"
    }
  ],
  "meta": {
    "trending_count": 10,
    "entities_count": 10,
    "views_count": 1
  }
}
```

**Use Cases:**
- Dashboard homepage
- Executive summary
- Quick analytics snapshot
- System health check

---

## 📊 Article Statistics

### Get Article Statistics by Source

Retourneert gedetailleerde statistieken per news source.

**Endpoint:** `GET /analytics/article-stats`

**Example Request:**
```bash
curl "http://localhost:8080/api/v1/analytics/article-stats"
```

**Example Response:**
```json
{
  "sources": [
    {
      "source": "nu.nl",
      "source_name": "NU.nl",
      "total_articles": 75,
      "articles_today": 12,
      "articles_week": 65,
      "ai_processed_count": 75,
      "content_extracted_count": 30,
      "latest_article_date": "2025-10-30T02:50:00Z",
      "oldest_article_date": "2025-10-15T08:30:00Z",
      "avg_sentiment": 0.15
    }
  ],
  "meta": {
    "count": 3
  }
}
```

**Response Fields:**
- `source` - Source domain
- `source_name` - Human-readable name
- `total_articles` - All-time article count
- `articles_today` - Last 24 hours
- `articles_week` - Last 7 days
- `ai_processed_count` - Articles with AI processing
- `content_extracted_count` - Articles with full content
- `latest_article_date` - Most recent article
- `oldest_article_date` - First article
- `avg_sentiment` - Average sentiment for this source

**Use Cases:**
- Source performance comparison
- Data quality monitoring
- Coverage analysis
- Processing status tracking

---

## 🔄 Refresh Analytics

### Refresh Materialized Views

Trigger een refresh van alle analytics materialized views. Gebruikt CONCURRENT mode voor zero downtime.

**Endpoint:** `POST /analytics/refresh`

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `concurrent` | boolean | true | Use CONCURRENT mode (recommended) |

**Example Request:**
```bash
curl -X POST "http://localhost:8080/api/v1/analytics/refresh?concurrent=true"
```

**Example Response:**
```json
{
  "message": "Analytics refreshed successfully",
  "results": [
    {
      "view_name": "mv_trending_keywords",
      "refresh_time_ms": 450,
      "rows_affected": 62
    },
    {
      "view_name": "mv_sentiment_timeline",
      "refresh_time_ms": 320,
      "rows_affected": 105
    }
  ],
  "summary": {
    "total_views": 3,
    "total_rows": 250,
    "total_time_ms": 1250,
    "concurrent_mode": true
  }
}
```

**Recommended Schedule:**
- Production: Every 5-15 minutes
- Development: Every 30-60 minutes
- Low traffic: Hourly

**Use Cases:**
- Scheduled refresh tasks
- Manual data update
- After bulk data import
- Testing updated analytics

---

## 🛠️ Maintenance Schedule

### Get Maintenance Recommendations

Retourneert aanbevolen maintenance taken en hun status.

**Endpoint:** `GET /analytics/maintenance-schedule`

**Example Request:**
```bash
curl "http://localhost:8080/api/v1/analytics/maintenance-schedule"
```

**Example Response:**
```json
{
  "tasks": [
    {
      "task": "Refresh Materialized Views",
      "frequency": "Every 5-15 minutes",
      "last_run": "2025-10-30T02:50:00Z",
      "next_recommended": "2025-10-30T03:05:00Z",
      "status": "✓ ON SCHEDULE"
    },
    {
      "task": "Clean Old Emails",
      "frequency": "Weekly",
      "last_run": null,
      "next_recommended": null,
      "status": "⚠️  RUN MANUALLY"
    }
  ],
  "meta": {
    "count": 3
  }
}
```

**Task Status Values:**
- `✓ ON SCHEDULE` - Task is up to date
- `⚠️  OVERDUE` - Task needs to run
- `⚠️  RUN MANUALLY` - Manual task (no automation)

**Use Cases:**
- Operations dashboard
- Maintenance planning
- System health monitoring
- Automation scheduling

---

## 💊 Database Health

### Get Database Health Metrics

Retourneert database health indicators zoals table sizes, cache hit ratio, en connection count.

**Endpoint:** `GET /analytics/database-health`

**Example Request:**
```bash
curl "http://localhost:8080/api/v1/analytics/database-health"
```

**Example Response:**
```json
{
  "table_sizes": [
    {
      "table": "articles",
      "size": "2.5 MB",
      "bytes": 2621440
    },
    {
      "table": "emails",
      "size": "512 kB",
      "bytes": 524288
    }
  ],
  "cache_hit_ratio": 99.2,
  "connection_count": 5,
  "status": "healthy"
}
```

**Health Indicators:**
- **Cache Hit Ratio:**
  - > 99% = Excellent
  - > 95% = Good
  - > 90% = Fair
  - < 90% = Poor (needs tuning)

- **Connection Count:**
  - < 50 = Normal
  - 50-80 = Busy
  - > 80 = High load

**Use Cases:**
- Performance monitoring
- Capacity planning
- Database tuning
- Operations dashboard

---

## 📋 Complete Endpoint List

### Public Endpoints

| Method | Endpoint | Description | Performance |
|--------|----------|-------------|-------------|
| GET | `/analytics/trending` | Trending keywords | ~50ms |
| GET | `/analytics/sentiment-trends` | Sentiment over time | ~100ms |
| GET | `/analytics/hot-entities` | Most mentioned entities | ~75ms |
| GET | `/analytics/entity-sentiment` | Entity sentiment timeline | ~150ms |
| GET | `/analytics/overview` | Complete overview | ~200ms |
| GET | `/analytics/article-stats` | Stats by source | ~50ms |
| GET | `/analytics/maintenance-schedule` | Maintenance tasks | ~25ms |
| GET | `/analytics/database-health` | Database metrics | ~100ms |
| POST | `/analytics/refresh` | Refresh views | ~1-2s |

### Performance Notes

- **Materialized views** worden elke 5-15 minuten refreshed
- **Queries** zijn 90% sneller dan dynamische aggregaties
- **Cache hit ratio** van 99%+ verwacht
- **Response times** < 200ms voor meeste endpoints

## 🔍 Query Examples

### Get Trending Topics (JavaScript)

```javascript
async function getTrendingTopics(hours = 24) {
  const response = await fetch(
    `http://localhost:8080/api/v1/analytics/trending?hours=${hours}&limit=10`
  );
  const data = await response.json();
  return data.trending;
}
```

### Get Sentiment Trends (JavaScript)

```javascript
async function getSentimentTrends(source = '') {
  const url = source 
    ? `http://localhost:8080/api/v1/analytics/sentiment-trends?source=${source}`
    : 'http://localhost:8080/api/v1/analytics/sentiment-trends';
  const response = await fetch(url);
  const data = await response.json();
  return data.trends;
}
```

### Track Entity Sentiment (Python)

```python
import requests

def track_entity_sentiment(entity: str, days: int = 30):
    url = f"http://localhost:8080/api/v1/analytics/entity-sentiment"
    params = {"entity": entity, "days": days}
    response = requests.get(url, params=params)
    return response.json()

# Example
timeline = track_entity_sentiment("Elon Musk", 30)
for day in timeline["timeline"]:
    print(f"{day['day']}: {day['mention_count']} mentions, sentiment: {day['avg_sentiment']}")
```

### Refresh Analytics (cURL)

```bash
# Refresh all views (concurrent mode)
curl -X POST "http://localhost:8080/api/v1/analytics/refresh"

# Refresh with blocking mode (faster but blocks queries)
curl -X POST "http://localhost:8080/api/v1/analytics/refresh?concurrent=false"
```

## 🎯 Integration Examples

### React Dashboard Component

```jsx
import { useEffect, useState } from 'react';

function AnalyticsDashboard() {
  const [trending, setTrending] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchAnalytics() {
      try {
        const response = await fetch('http://localhost:8080/api/v1/analytics/overview');
        const data = await response.json();
        setTrending(data.trending_keywords);
      } catch (error) {
        console.error('Failed to fetch analytics:', error);
      } finally {
        setLoading(false);
      }
    }

    fetchAnalytics();
    // Refresh every 5 minutes
    const interval = setInterval(fetchAnalytics, 5 * 60 * 1000);
    return () => clearInterval(interval);
  }, []);

  if (loading) return <div>Loading...</div>;

  return (
    <div>
      <h2>Trending Topics</h2>
      {trending.map(kw => (
        <div key={kw.keyword}>
          <strong>{kw.keyword}</strong>: {kw.article_count} articles
          (Score: {kw.trending_score})
        </div>
      ))}
    </div>
  );
}
```

### Automated Refresh Script

```bash
#!/bin/bash
# refresh-analytics.sh
# Add to crontab: */15 * * * * /path/to/refresh-analytics.sh

curl -X POST "http://localhost:8080/api/v1/analytics/refresh" \
  -H "Content-Type: application/json" \
  >> /var/log/analytics-refresh.log 2>&1

echo "Analytics refreshed at $(date)" >> /var/log/analytics-refresh.log
```

## ⚡ Performance Optimization

### Caching Strategy

```javascript
// Cache analytics data client-side
const ANALYTICS_CACHE_TIME = 5 * 60 * 1000; // 5 minutes

class AnalyticsCache {
  constructor() {
    this.cache = new Map();
  }

  async getTrending(hours = 24) {
    const key = `trending_${hours}`;
    const cached = this.cache.get(key);
    
    if (cached && Date.now() - cached.timestamp < ANALYTICS_CACHE_TIME) {
      return cached.data;
    }

    const response = await fetch(
      `http://localhost:8080/api/v1/analytics/trending?hours=${hours}`
    );
    const data = await response.json();

    this.cache.set(key, {
      data: data.trending,
      timestamp: Date.now()
    });

    return data.trending;
  }
}
```

### Batch Queries

```javascript
// Fetch multiple analytics in parallel
async function getAnalyticsDashboard() {
  const [trending, sentiment, entities] = await Promise.all([
    fetch('http://localhost:8080/api/v1/analytics/trending').then(r => r.json()),
    fetch('http://localhost:8080/api/v1/analytics/sentiment-trends').then(r => r.json()),
    fetch('http://localhost:8080/api/v1/analytics/hot-entities?limit=10').then(r => r.json())
  ]);

  return { trending, sentiment, entities };
}
```

## 🐛 Error Handling

### Common Errors

**500 Internal Server Error**
```json
{
  "error": "database_error",
  "message": "Failed to fetch trending keywords",
  "code": 500
}
```
**Cause:** Database query failed  
**Solution:** Check database connection, verify materialized views exist

**400 Bad Request**
```json
{
  "error": "missing_parameter",
  "message": "entity parameter is required",
  "code": 400
}
```
**Cause:** Required parameter missing  
**Solution:** Provide all required parameters

### Error Handling Example

```javascript
async function fetchTrendingWithErrorHandling() {
  try {
    const response = await fetch('http://localhost:8080/api/v1/analytics/trending');
    
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to fetch trending');
    }
    
    return await response.json();
  } catch (error) {
    console.error('Analytics error:', error);
    // Fallback to cached data or show error message
    return { trending: [], meta: { count: 0 } };
  }
}
```

## 📝 Best Practices

1. **Cache Results:** Client-side caching for 5-15 minutes
2. **Batch Requests:** Use Promise.all() voor multiple endpoints
3. **Error Handling:** Always handle errors gracefully
4. **Refresh Schedule:** POST /refresh every 5-15 minutes
5. **Parameter Validation:** Validate limits en time windows
6. **Monitor Performance:** Track response times
7. **Use Overview:** For dashboard homepage
8. **Specific Queries:** For detailed analysis

## 🔗 Related Endpoints

- **Articles API:** `/api/v1/articles` - Full article data
- **AI API:** `/api/v1/ai` - AI processing endpoints
- **Stocks API:** `/api/v1/stocks` - Stock data
- **Health API:** `/health/metrics` - System health

## 📞 Support

Voor vragen:
- Check [API Documentation](README.md)
- Review [Database Schema](../DATABASE-SCHEMA-V2-MIGRATION.md)
- See [Migration Guide](../../migrations/MIGRATION-GUIDE.md)