# Frontend API Documentation

Deze documentatie beschrijft de API endpoints voor gebruik in de frontend applicatie.

## Base URL
```
http://localhost:8080/api/v1
```

## Authenticatie

De API ondersteunt optionele API key authenticatie via de header:
```
X-API-Key: your-api-key-here
```

Voor publieke endpoints is authenticatie optioneel. Voor admin endpoints (scraping) is authenticatie verplicht.

## Standaard Response Formaat

Alle API responses volgen dit standaard formaat:

### Succes Response
```json
{
  "success": true,
  "data": { /* response data */ },
  "meta": { /* metadata zoals paginatie */ },
  "request_id": "unique-request-id",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Error Response
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": "Additional error details"
  },
  "request_id": "unique-request-id",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## CORS & Headers

De API is geconfigureerd met volledige CORS ondersteuning:
- **Allowed Origins**: `*` (alle origins)
- **Allowed Methods**: `GET, POST, PUT, DELETE, PATCH, OPTIONS`
- **Exposed Headers**: `X-Request-ID, X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset`

## Rate Limiting

De API heeft rate limiting per IP/API key:
- **Limiet**: 100 requests per 60 seconden (configureerbaar)
- **Headers**: 
  - `X-RateLimit-Limit`: Maximum aantal requests
  - `X-RateLimit-Remaining`: Resterende requests
  - `X-RateLimit-Reset`: Reset tijd in seconden

## Endpoints

### Health Monitoring

De API biedt verschillende health check endpoints voor monitoring en observability:

#### Comprehensive Health Check

**GET** `/health`

Geeft gedetailleerde status van alle componenten met metrics.

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "timestamp": "2024-01-01T12:00:00Z",
    "version": "1.0.0",
    "uptime_seconds": 3600.5,
    "components": {
      "database": {
        "status": "healthy",
        "message": "Database connection healthy",
        "latency_ms": 2.5,
        "details": {
          "total_conns": 10,
          "idle_conns": 7,
          "acquired_conns": 3,
          "max_conns": 25
        }
      },
      "redis": {
        "status": "healthy",
        "message": "Redis connection healthy",
        "latency_ms": 1.2,
        "details": {
          "cache_available": true
        }
      },
      "scraper": {
        "status": "healthy",
        "message": "Scraper service operational"
      },
      "ai_processor": {
        "status": "healthy",
        "message": "AI processor operational",
        "details": {
          "is_running": true,
          "process_count": 150,
          "last_run": "2024-01-01T11:55:00Z",
          "current_interval": "5m0s"
        }
      }
    },
    "metrics": {
      "uptime_seconds": 3600.5,
      "timestamp": 1704117000,
      "db_total_conns": 10,
      "db_idle_conns": 7,
      "db_acquired_conns": 3,
      "ai_process_count": 150,
      "ai_is_running": true
    }
  },
  "request_id": "abc123",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

**Status Codes:**
- `200 OK` - System is healthy or degraded but operational
- `503 Service Unavailable` - System is unhealthy

**Component Status Values:**
- `healthy` - Component is functioning normally
- `degraded` - Component is operational but with issues
- `unhealthy` - Component has failed
- `disabled` - Component is not configured

#### Liveness Probe

**GET** `/health/live`

Simpele check of de applicatie draait (Kubernetes-compatible).

**Response:**
```json
{
  "status": "alive",
  "time": "2024-01-01T12:00:00Z"
}
```

**Status:** Always returns `200 OK` if application is running.

#### Readiness Probe

**GET** `/health/ready`

Check of de applicatie klaar is om traffic te ontvangen (Kubernetes-compatible).

**Response:**
```json
{
  "status": "ready",
  "components": {
    "database": true,
    "redis": true
  },
  "time": "2024-01-01T12:00:00Z"
}
```

**Status Codes:**
- `200 OK` - Application is ready
- `503 Service Unavailable` - Application is not ready

#### Detailed Metrics

**GET** `/health/metrics`

Prometheus-compatible metrics endpoint met gedetailleerde statistieken.

**Response:**
```json
{
  "success": true,
  "data": {
    "timestamp": 1704117000,
    "uptime": 3600.5,
    "db_total_conns": 10,
    "db_idle_conns": 7,
    "db_acquired_conns": 3,
    "db_max_conns": 25,
    "db_acquire_count": 1500,
    "db_acquire_duration_ms": 5,
    "ai_is_running": true,
    "ai_process_count": 150,
    "ai_last_run": 1704116700,
    "ai_current_interval_seconds": 300,
    "scraper": {
      "total_scrapes": 50,
      "successful_scrapes": 48,
      "failed_scrapes": 2
    }
  },
  "request_id": "abc123",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Articles

#### List Articles

**GET** `/api/v1/articles`

Haal een lijst van artikelen op met filtering, sorting en paginatie.

**Query Parameters:**
- `limit` (integer, default: 50, max: 100) - Aantal resultaten per pagina
- `offset` (integer, default: 0) - Offset voor paginatie
- `source` (string) - Filter op bron (bijv. "nu.nl", "ad.nl", "nos.nl")
- `category` (string) - Filter op categorie
- `keyword` (string) - Filter op keyword in article keywords
- `start_date` (RFC3339 string) - Start datum filter (bijv. "2024-01-01T00:00:00Z")
- `end_date` (RFC3339 string) - Eind datum filter
- `sort_by` (string, default: "published") - Sorteer op: "published", "created_at", "title"
- `sort_order` (string, default: "desc") - Sorteer richting: "asc", "desc"

**Voorbeeld Request:**
```
GET /api/v1/articles?limit=20&offset=0&source=nu.nl&sort_by=published&sort_order=desc
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 123,
      "title": "Article Title",
      "summary": "Article summary...",
      "url": "https://example.com/article",
      "published": "2024-01-01T12:00:00Z",
      "source": "nu.nl",
      "keywords": ["sport", "voetbal"],
      "image_url": "https://example.com/image.jpg",
      "author": "John Doe",
      "category": "Sport",
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "meta": {
    "pagination": {
      "total": 150,
      "limit": 20,
      "offset": 0,
      "current_page": 1,
      "total_pages": 8,
      "has_next": true,
      "has_prev": false
    },
    "sorting": {
      "sort_by": "published",
      "sort_order": "desc"
    },
    "filtering": {
      "source": "nu.nl"
    }
  },
  "request_id": "abc123",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

#### Get Single Article

**GET** `/api/v1/articles/:id`

Haal een specifiek artikel op via ID.

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 123,
    "title": "Article Title",
    "summary": "Article summary...",
    "url": "https://example.com/article",
    "published": "2024-01-01T12:00:00Z",
    "source": "nu.nl",
    "keywords": ["sport", "voetbal"],
    "image_url": "https://example.com/image.jpg",
    "author": "John Doe",
    "category": "Sport",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  },
  "request_id": "abc123",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

#### Search Articles

**GET** `/api/v1/articles/search`

Zoek artikelen met full-text search.

**Query Parameters:**
- `q` (string, required) - Zoekterm
- `limit` (integer, default: 50, max: 100)
- `offset` (integer, default: 0)
- `source` (string) - Filter op bron
- `category` (string) - Filter op categorie
- `sort_by` (string, default: "published")
- `sort_order` (string, default: "desc")

**Voorbeeld Request:**
```
GET /api/v1/articles/search?q=voetbal&limit=20
```

**Response:** Zelfde structuur als List Articles

#### Get Article Statistics

**GET** `/api/v1/articles/stats`

Haal uitgebreide statistieken op over alle artikelen.

**Response:**
```json
{
  "success": true,
  "data": {
    "total_articles": 1500,
    "articles_by_source": {
      "nu.nl": 800,
      "ad.nl": 450,
      "nos.nl": 250
    },
    "recent_articles_24h": 45,
    "oldest_article": "2024-01-01T00:00:00Z",
    "newest_article": "2024-01-15T18:30:00Z",
    "categories": {
      "Sport": {
        "name": "Sport",
        "article_count": 450
      },
      "Nieuws": {
        "name": "Nieuws",
        "article_count": 600
      }
    }
  },
  "request_id": "abc123",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Sources

#### Get All Sources

**GET** `/api/v1/sources`

Haal alle beschikbare nieuwsbronnen op.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "name": "nu.nl",
      "feed_url": "https://www.nu.nl/rss/Algemeen",
      "is_active": true
    },
    {
      "name": "ad.nl",
      "feed_url": "https://www.ad.nl/rss.xml",
      "is_active": true
    }
  ],
  "request_id": "abc123",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Categories

#### Get All Categories

**GET** `/api/v1/categories`

Haal alle beschikbare categorieën op met artikel aantallen.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "name": "Sport",
      "article_count": 450
    },
    {
      "name": "Nieuws",
      "article_count": 600
    }
  ],
  "request_id": "abc123",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Scraper (Protected)

Deze endpoints vereisen authenticatie via API key.

#### Trigger Scraping

**POST** `/api/v1/scrape`

Start het scrapen van nieuwsbronnen.

**Headers:**
```
X-API-Key: your-api-key
```

**Request Body:**
```json
{
  "source": "nu.nl"  // Optioneel: specifieke bron, leeg voor alle bronnen
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "success",
    "source": "nu.nl",
    "articles_found": 25,
    "articles_stored": 20,
    "articles_skipped": 5,
    "duration_seconds": 3.45
  },
  "request_id": "abc123",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

#### Get Scraper Statistics

**GET** `/api/v1/scraper/stats`

Haal scraper statistieken op (vereist authenticatie).

**Headers:**
```
X-API-Key: your-api-key
```

## Error Codes

| Code | Beschrijving |
|------|--------------|
| `INVALID_ID` | Ongeldig artikel ID |
| `INVALID_DATE` | Ongeldige datum formaat |
| `INVALID_REQUEST` | Ongeldige request body |
| `INVALID_SOURCE` | Onbekende nieuwsbron |
| `NOT_FOUND` | Resource niet gevonden |
| `MISSING_QUERY` | Verplichte query parameter ontbreekt |
| `DATABASE_ERROR` | Database fout |
| `SEARCH_ERROR` | Zoek fout |
| `SCRAPING_FAILED` | Scraping mislukt |

## Frontend Implementatie Tips

### 1. Paginatie Implementatie

```javascript
const fetchArticles = async (page = 1, limit = 20) => {
  const offset = (page - 1) * limit;
  const response = await fetch(
    `http://localhost:8080/api/v1/articles?limit=${limit}&offset=${offset}`
  );
  const data = await response.json();
  
  if (data.success) {
    return {
      articles: data.data,
      pagination: data.meta.pagination
    };
  }
  throw new Error(data.error.message);
};
```

### 2. Zoeken met Debounce

```javascript
const searchArticles = async (query) => {
  const response = await fetch(
    `http://localhost:8080/api/v1/articles/search?q=${encodeURIComponent(query)}&limit=20`
  );
  const data = await response.json();
  
  if (data.success) {
    return data.data;
  }
  throw new Error(data.error.message);
};

// Gebruik met debounce
import { debounce } from 'lodash';
const debouncedSearch = debounce(searchArticles, 300);
```

### 3. Filtering & Sorting

```javascript
const fetchFilteredArticles = async (filters) => {
  const params = new URLSearchParams({
    limit: filters.limit || 50,
    offset: filters.offset || 0,
    sort_by: filters.sortBy || 'published',
    sort_order: filters.sortOrder || 'desc'
  });
  
  if (filters.source) params.append('source', filters.source);
  if (filters.category) params.append('category', filters.category);
  if (filters.startDate) params.append('start_date', filters.startDate);
  if (filters.endDate) params.append('end_date', filters.endDate);
  
  const response = await fetch(
    `http://localhost:8080/api/v1/articles?${params}`
  );
  const data = await response.json();
  
  return data;
};
```

### 4. Error Handling

```javascript
const handleApiError = (errorData) => {
  const { error } = errorData;
  
  switch (error.code) {
    case 'NOT_FOUND':
      // Toon 404 pagina
      break;
    case 'DATABASE_ERROR':
      // Toon server error message
      break;
    case 'INVALID_DATE':
      // Toon validatie error
      break;
    default:
      // Toon generieke error
      console.error(error.message);
  }
};
```

### 5. Request ID Tracking

Elke response bevat een `request_id` die gebruikt kan worden voor debugging:

```javascript
const fetchWithTracking = async (url) => {
  const response = await fetch(url);
  const data = await response.json();
  
  // Log request ID voor debugging
  console.log('Request ID:', data.request_id);
  
  return data;
};
```

## TypeScript Types

```typescript
interface APIResponse<T> {
  success: boolean;
  data?: T;
  error?: APIError;
  meta?: Meta;
  request_id: string;
  timestamp: string;
}

interface APIError {
  code: string;
  message: string;
  details?: string;
}

interface Meta {
  pagination?: PaginationMeta;
  sorting?: SortingMeta;
  filtering?: FilteringMeta;
}

interface PaginationMeta {
  total: number;
  limit: number;
  offset: number;
  current_page: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

interface SortingMeta {
  sort_by: string;
  sort_order: 'asc' | 'desc';
}

interface FilteringMeta {
  source?: string;
  category?: string;
  keyword?: string;
  search?: string;
  start_date?: string;
  end_date?: string;
}

interface Article {
  id: number;
  title: string;
  summary: string;
  url: string;
  published: string;
  source: string;
  keywords: string[];
  image_url: string;
  author: string;
  category: string;
  created_at: string;
  updated_at: string;
}

interface SourceInfo {
  name: string;
  feed_url: string;
  is_active: boolean;
}

interface CategoryInfo {
  name: string;
  article_count: number;
}

interface StatsResponse {
  total_articles: number;
  articles_by_source: Record<string, number>;
  recent_articles_24h: number;
  oldest_article?: string;
  newest_article?: string;
  categories: Record<string, CategoryInfo>;
}
```

## Cache Strategie

De API gebruikt caching met Redis (5 minuten TTL):
- Article lists worden gecached
- Individual articles worden gecached
- Stats worden gecached
- Search results worden NIET gecached (real-time)

Bij nieuwe scraping runs wordt de cache automatisch geïnvalideerd.

## Best Practices

1. **Request IDs**: Gebruik de `request_id` uit responses voor debugging
2. **Rate Limiting**: Check de `X-RateLimit-*` headers en pas je request rate aan
3. **Error Handling**: Implementeer proper error handling voor alle error codes
4. **Paginatie**: Gebruik altijd paginatie voor grote datasets
5. **Caching**: Implementeer client-side caching voor performance
6. **Timestamps**: Alle timestamps zijn in UTC (RFC3339 format)
7. **Search**: Gebruik debouncing voor search queries
8. **Loading States**: Toon loading indicators tijdens API calls