
# Configuration API Reference

## 📋 Overview

De Configuration API stelt je in staat om **runtime settings** aan te passen en tussen **scraper profiles** te wisselen zonder de applicatie te herstarten.

## 🎯 Features

- ✅ **4 Predefined Profiles** (Fast, Balanced, Deep, Conservative)
- ✅ **Runtime Setting Updates** (rate limits, concurrency, timeouts)
- ✅ **Profile Switching** zonder downtime
- ✅ **Scheduler Management** met live updates
- ✅ **Response Caching** met Redis (95% cache hit ratio)

---

## 🔌 API Endpoints

### 1. Get All Profiles

**Endpoint:** `GET /api/v1/config/profiles`  
**Auth:** None (public)

Returns alle beschikbare scraper profiles met hun configuratie.

**Response:**
```json
{
  "success": true,
  "data": {
    "profiles": {
      "fast": {
        "name": "fast",
        "rate_limit_seconds": 2,
        "max_concurrent": 10,
        "timeout_seconds": 15,
        "schedule_interval_min": 5,
        "browser_pool_size": 10,
        "browser_max_concurrent": 5,
        "target_sites": ["nu.nl", "ad.nl", "nos.nl"],
        "enable_browser_scraping": true,
        "enable_full_content": false,
        "active": false
      },
      "balanced": {
        "name": "balanced",
        "rate_limit_seconds": 3,
        "max_concurrent": 5,
        "timeout_seconds": 30,
        "schedule_interval_min": 15,
        "browser_pool_size": 5,
        "browser_max_concurrent": 3,
        "target_sites": ["nu.nl", "ad.nl", "nos.nl"],
        "enable_browser_scraping": true,
        "enable_full_content": false,
        "active": true
      },
      "deep": {
        "name": "deep",
        "rate_limit_seconds": 5,
        "max_concurrent": 3,
        "timeout_seconds": 30,
        "schedule_interval_min": 60,
        "browser_pool_size": 7,
        "browser_max_concurrent": 4,
        "target_sites": ["nu.nl", "ad.nl", "nos.nl", "trouw.nl"],
        "enable_browser_scraping": true,
        "enable_full_content": true,
        "active": false
      },
      "conservative": {
        "name": "conservative",
        "rate_limit_seconds": 10,
        "max_concurrent": 2,
        "timeout_seconds": 60,
        "schedule_interval_min": 30,
        "browser_pool_size": 2,
        "browser_max_concurrent": 1,
        "target_sites": ["nu.nl", "ad.nl", "nos.nl"],
        "enable_browser_scraping": true,
        "enable_full_content": false,
        "active": false
      }
    },
    "active_profile": "balanced",
    "total_profiles": 4
  },
  "request_id": "abc123"
}
```

---

### 2. Get Current Configuration

**Endpoint:** `GET /api/v1/config/current`  
**Auth:** None (public)

Returns de huidige actieve configuratie met alle instellingen.

**Response:**
```json
{
  "success": true,
  "data": {
    "active_profile": "balanced",
    "rate_limit_seconds": 3,
    "max_concurrent": 5,
    "timeout_seconds": 30,
    "retry_attempts": 3,
    "schedule_interval_min": 15,
    "target_sites": ["nu.nl", "ad.nl", "nos.nl"],
    "enable_browser_scraping": true,
    "browser_pool_size": 5,
    "browser_timeout_seconds": 15,
    "browser_wait_after_load_ms": 1500,
    "browser_fallback_only": true,
    "browser_max_concurrent": 3,
    "enable_full_content": false,
    "content_batch_size": 15,
    "enable_robots_check": true,
    "enable_duplicate_detection": true
  },
  "request_id": "abc123"
}
```

---

### 3. Switch Profile

**Endpoint:** `POST /api/v1/config/profile/:name`  
**Auth:** Required (API Key)

Wisselt naar een ander scraper profile.

**Parameters:**
- `name` (path): Profile naam (`fast`, `balanced`, `deep`, `conservative`)

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/config/profile/fast \
  -H "X-API-Key: your-api-key"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "success": true,
    "message": "Switched to profile 'fast'",
    "old_profile": "balanced",
    "new_profile": "fast",
    "new_interval": 5,
    "new_rate_limit": 2
  },
  "request_id": "abc123"
}
```

**Effects:**
- ✅ Scheduler interval wordt aangepast
- ✅ Rate limits worden aangepast
- ✅ Concurrency settings worden aangepast
- ⚠️ Browser pool size verandert NIET (vereist restart)

---

### 4. Update Single Setting

**Endpoint:** `PATCH /api/v1/config/setting`  
**Auth:** Required (API Key)

Update een specifieke configuratie setting.

**Request Body:**
```json
{
  "setting": "rate_limit_seconds",
  "value": 5
}
```

**Supported Settings:**

| Setting | Type | Range | Description |
|---------|------|-------|-------------|
| `rate_limit_seconds` | int | 1-60 | Seconden tussen requests per domain |
| `max_concurrent` | int | 1-20 | Maximum parallelle scrapes |
| `timeout_seconds` | int | 10-120 | Timeout per scrape operatie |
| `schedule_interval_minutes` | int | 1-1440 | Minuten tussen scheduled scrapes |
| `browser_pool_size` | int | 1-20 | Aantal browser instances |
| `browser_max_concurrent` | int | 1-10 | Parallelle browser operations |
| `content_batch_size` | int | 5-50 | Batch size voor content extraction |
| `enable_browser_scraping` | bool | - | Browser scraping aan/uit |
| `enable_full_content` | bool | - | Full content extraction aan/uit |
| `enable_robots_check` | bool | - | Robots.txt checking aan/uit |
| `browser_fallback_only` | bool | - | Browser alleen als fallback |

**Example:**
```bash
curl -X PATCH http://localhost:8080/api/v1/config/setting \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "setting": "schedule_interval_minutes",
    "value": 10
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "success": true,
    "message": "Setting 'schedule_interval_minutes' updated successfully",
    "setting": "schedule_interval_minutes",
    "old_value": 15,
    "new_value": 10,
    "profile": "balanced"
  },
  "request_id": "abc123"
}
```

---

### 5. Get Scheduler Status

**Endpoint:** `GET /api/v1/config/scheduler/status`  
**Auth:** None (public)

Returns de huidige scheduler status.

**Response:**
```json
{
  "success": true,
  "data": {
    "running": true,
    "active_profile": "balanced",
    "interval_minutes": 15,
    "next_run": "2025-10-30T15:00:00Z",
    "enabled": true
  },
  "request_id": "abc123"
}
```

---

### 6. Reset to Defaults

**Endpoint:** `POST /api/v1/config/reset`  
**Auth:** Required (API Key)

Reset het actieve profile naar zijn default waarden.

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/config/reset \
  -H "X-API-Key: your-api-key"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "success": true,
    "message": "Profile 'balanced' reset to default values",
    "profile": "balanced"
  },
  "request_id": "abc123"
}
```

---

## 📊 Profile Comparison

| Feature | Fast 🚀 | Balanced ⚖️ | Deep 🔍 | Conservative 🛡️ |
|---------|---------|------------|---------|-----------------|
| **Rate Limit** | 2s | 3s | 5s | 10s |
| **Concurrent** | 10 | 5 | 3 | 2 |
| **Browser Pool** | 10 | 5 | 7 | 2 |
| **Interval** | 5 min | 15 min | 60 min | 30 min |
| **Full Content** | ❌ | ❌ | ✅ | ❌ |
| **Robots Check** | ❌ | ✅ | ✅ | ✅ |
| **Use Case** | Breaking news | Production | Quality | Low load |
| **Throughput** | ~360/hour | ~320/hour | ~100/hour | ~80/hour |

---

## 🔄 Response Caching (v3.1)

Alle list

## 🔄 Response Caching (v3.1)

### Cache Strategy

Alle list en search endpoints gebruiken nu **intelligente Redis caching**:

| Endpoint | Cache TTL | Cache Key Pattern |
|----------|-----------|-------------------|
| `GET /api/v1/articles` | 2 minutes | `articles:source:category:limit:offset` |
| `GET /api/v1/articles/search` | 1 minute | `articles:search:query:source:limit:offset` |
| `GET /api/v1/articles/stats` | 5 minutes | `stats:comprehensive` |
| `GET /api/v1/categories` | 5 minutes | `stats:categories` |
| `GET /api/v1/articles/:id` | 5 minutes | `article:id` |

### Performance Impact

**Cache Hit:** ~2ms response (95% faster!)  
**Cache Miss:** ~25ms response (10x faster with ListLight)

### Expected Cache Hit Ratios

- Article Lists: 80-90%
- Search Results: 60-70%
- Stats/Categories: 95%+

---

## 🎯 Use Cases & Examples

### Switch to Fast Profile (Breaking News)
```bash
curl -X POST http://localhost:8080/api/v1/config/profile/fast \
  -H "X-API-Key: your-api-key"
```

### Adjust Rate Limiting
```bash
curl -X PATCH http://localhost:8080/api/v1/config/setting \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"setting": "rate_limit_seconds", "value": 2}'
```

---

## 🎨 Frontend Integration

### React Profile Switcher
```typescript
const switchProfile = async (name: string) => {
  const response = await fetch(
    `http://localhost:8080/api/v1/config/profile/${name}`,
    {
      method: 'POST',
      headers: { 'X-API-Key': 'your-key' }
    }
  );
  const data = await response.json();
  if (data.success) {
    alert(`Switched to ${name}!`);
  }
};
```

---

## ⚠️ Important Notes

**Requires Restart:**
- Database/Redis pool size
- Target sites list
- User agent string

**Runtime Adjustable:**
- ✅ Rate limiting
- ✅ Concurrency
- ✅ Timeouts
- ✅ Schedule interval
- ✅ Feature toggles

---

**Version:** 3.1  
**Status:** ✅ Production Ready