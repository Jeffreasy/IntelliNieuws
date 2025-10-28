# Frontend Advanced Guide

Complete gids voor geavanceerde frontend integratie met de NieuwsScraper API.

## 📋 Inhoudsopgave

1. [Health Monitoring & Observability](#health-monitoring--observability)
2. [Advanced TypeScript Types](#advanced-typescript-types)
3. [Production-Ready Patterns](#production-ready-patterns)
4. [Performance Optimalisatie](#performance-optimalisatie)
5. [Real-time Updates & Polling](#real-time-updates--polling)
6. [Error Recovery Strategieën](#error-recovery-strategieën)
7. [Testing Strategieën](#testing-strategieën)
8. [Deployment Checklist](#deployment-checklist)

---

## Health Monitoring & Observability

### Health Check Endpoints

De API biedt meerdere health check endpoints voor verschillende use cases:

#### 1. Comprehensive Health Check
**Endpoint:** `GET /health`

Geeft gedetailleerde status van alle componenten:

```typescript
interface HealthResponse {
  status: 'healthy' | 'degraded' | 'unhealthy';
  timestamp: string;
  version: string;
  uptime_seconds: number;
  components: {
    database: ComponentHealth;
    redis: ComponentHealth;
    scraper: ComponentHealth;
    ai_processor?: ComponentHealth;
  };
  metrics: {
    uptime_seconds: number;
    timestamp: number;
    db_total_conns?: number;
    db_idle_conns?: number;
    db_acquired_conns?: number;
    ai_process_count?: number;
    ai_is_running?: boolean;
  };
}

interface ComponentHealth {
  status: 'healthy' | 'degraded' | 'unhealthy' | 'disabled';
  message?: string;
  latency_ms?: number;
  details?: Record<string, any>;
}
```

**Response Voorbeeld:**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "timestamp": "2025-01-28T16:30:00Z",
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
          "last_run": "2025-01-28T16:25:00Z",
          "current_interval": "5m0s"
        }
      }
    },
    "metrics": {
      "uptime_seconds": 3600.5,
      "timestamp": 1706461800,
      "db_total_conns": 10,
      "db_idle_conns": 7,
      "db_acquired_conns": 3,
      "ai_process_count": 150,
      "ai_is_running": true
    }
  },
  "request_id": "abc123",
  "timestamp": "2025-01-28T16:30:00Z"
}
```

**Gebruik:**
```typescript
const checkHealth = async (): Promise<HealthResponse> => {
  const response = await fetch('http://localhost:8080/health');
  const data = await response.json();
  return data.data;
};

// Implementeer status dashboard
const HealthDashboard = () => {
  const { data: health, isLoading } = useQuery({
    queryKey: ['health'],
    queryFn: checkHealth,
    refetchInterval: 30000, // Check elke 30 seconden
  });

  if (isLoading) return <Spinner />;

  return (
    <div className="health-dashboard">
      <StatusIndicator status={health.status} />
      <Uptime seconds={health.uptime_seconds} />
      
      <div className="components">
        {Object.entries(health.components).map(([name, component]) => (
          <ComponentCard 
            key={name}
            name={name}
            health={component}
          />
        ))}
      </div>
    </div>
  );
};
```

#### 2. Liveness Probe
**Endpoint:** `GET /health/live`

Simpele check of de applicatie draait (voor Kubernetes/Docker):

```json
{
  "status": "alive",
  "time": "2025-01-28T16:30:00Z"
}
```

**Gebruik:**
```typescript
const checkLiveness = async (): Promise<boolean> => {
  try {
    const response = await fetch('http://localhost:8080/health/live');
    return response.ok;
  } catch {
    return false;
  }
};
```

#### 3. Readiness Probe
**Endpoint:** `GET /health/ready`

Check of de applicatie klaar is om traffic te ontvangen:

```json
{
  "status": "ready",
  "components": {
    "database": true,
    "redis": true
  },
  "time": "2025-01-28T16:30:00Z"
}
```

**HTTP Status:**
- `200 OK` - Ready
- `503 Service Unavailable` - Not ready

#### 4. Detailed Metrics
**Endpoint:** `GET /health/metrics`

Prometheus-compatibele metrics:

```json
{
  "success": true,
  "data": {
    "timestamp": 1706461800,
    "uptime": 3600.5,
    "db_total_conns": 10,
    "db_idle_conns": 7,
    "db_acquired_conns": 3,
    "db_max_conns": 25,
    "db_acquire_count": 1500,
    "db_acquire_duration_ms": 5,
    "ai_is_running": true,
    "ai_process_count": 150,
    "ai_last_run": 1706461500,
    "ai_current_interval_seconds": 300,
    "scraper": {
      "total_scrapes": 50,
      "successful_scrapes": 48,
      "failed_scrapes": 2
    }
  },
  "request_id": "abc123",
  "timestamp": "2025-01-28T16:30:00Z"
}
```

---

## Advanced TypeScript Types

### Complete Type Definitions

```typescript
// ============================================
// API Response Types
// ============================================

interface APIResponse<T = any> {
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

// ============================================
// Article Types
// ============================================

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

interface ArticleFilter {
  source?: string;
  category?: string;
  keyword?: string;
  search?: string;
  start_date?: string;
  end_date?: string;
  sort_by?: 'published' | 'created_at' | 'title';
  sort_order?: 'asc' | 'desc';
  limit?: number;
  offset?: number;
}

// ============================================
// Source & Category Types
// ============================================

interface SourceInfo {
  name: string;
  domain: string;
  feed_url: string;
  article_count: number;
  is_active: boolean;
}

interface CategoryInfo {
  name: string;
  article_count: number;
}

// ============================================
// Statistics Types
// ============================================

interface StatsResponse {
  total_articles: number;
  articles_by_source: Record<string, number>;
  recent_articles_24h: number;
  oldest_article?: string;
  newest_article?: string;
  categories: Record<string, CategoryInfo>;
}

// ============================================
// Health Types
// ============================================

interface HealthResponse {
  status: 'healthy' | 'degraded' | 'unhealthy';
  timestamp: string;
  version: string;
  uptime_seconds: number;
  components: Record<string, ComponentHealth>;
  metrics: Record<string, any>;
}

interface ComponentHealth {
  status: 'healthy' | 'degraded' | 'unhealthy' | 'disabled';
  message?: string;
  latency_ms?: number;
  details?: Record<string, any>;
}

interface LivenessResponse {
  status: 'alive';
  time: string;
}

interface ReadinessResponse {
  status: 'ready' | 'not_ready';
  components: Record<string, boolean>;
  time: string;
}

interface MetricsResponse {
  timestamp: number;
  uptime: number;
  db_total_conns?: number;
  db_idle_conns?: number;
  db_acquired_conns?: number;
  db_max_conns?: number;
  db_acquire_count?: number;
  db_acquire_duration_ms?: number;
  ai_is_running?: boolean;
  ai_process_count?: number;
  ai_last_run?: number;
  ai_current_interval_seconds?: number;
  scraper?: {
    total_scrapes: number;
    successful_scrapes: number;
    failed_scrapes: number;
  };
}

// ============================================
// AI Types (from FRONTEND_AI_API.md)
// ============================================

interface AIEnrichment {
  processed: boolean;
  processed_at?: string;
  sentiment?: SentimentAnalysis;
  categories?: Record<string, number>;
  entities?: EntityExtraction;
  keywords?: Keyword[];
  summary?: string;
  error?: string;
}

interface SentimentAnalysis {
  score: number;
  label: 'positive' | 'negative' | 'neutral';
  confidence?: number;
}

interface EntityExtraction {
  persons?: string[];
  organizations?: string[];
  locations?: string[];
}

interface Keyword {
  word: string;
  score: number;
}

interface SentimentStats {
  total_articles: number;
  positive_count: number;
  neutral_count: number;
  negative_count: number;
  average_sentiment: number;
  most_positive_title?: string;
  most_negative_title?: string;
}

interface TrendingTopic {
  keyword: string;
  article_count: number;
  average_sentiment: number;
  sources: string[];
}

interface TrendingTopicsResponse {
  topics: TrendingTopic[];
  hours_back: number;
  min_articles: number;
  count: number;
}

interface ProcessorStats {
  is_running: boolean;
  process_count: number;
  last_run: string;
  current_interval?: string;
}

// ============================================
// Scraper Types
// ============================================

interface ScrapeRequest {
  source?: string;
}

interface ScrapeResponse {
  status: string;
  source?: string;
  articles_found: number;
  articles_stored: number;
  articles_skipped: number;
  duration_seconds: number;
}
```

---

## Production-Ready Patterns

### 1. API Client met Retry Logic

```typescript
import axios, { AxiosInstance, AxiosError, AxiosRequestConfig } from 'axios';

class APIClient {
  private client: AxiosInstance;
  private readonly maxRetries = 3;
  private readonly retryDelay = 1000;

  constructor(baseURL: string, apiKey?: string) {
    this.client = axios.create({
      baseURL,
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
        ...(apiKey && { 'X-API-Key': apiKey }),
      },
    });

    this.setupInterceptors();
  }

  private setupInterceptors() {
    // Request interceptor
    this.client.interceptors.request.use(
      (config) => {
        // Add request ID voor tracking
        config.headers['X-Request-ID'] = this.generateRequestID();
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Response interceptor
    this.client.interceptors.response.use(
      (response) => {
        // Log successful requests
        console.log(`[API] ${response.config.method?.toUpperCase()} ${response.config.url} - ${response.status}`);
        return response;
      },
      async (error: AxiosError) => {
        return this.handleError(error);
      }
    );
  }

  private async handleError(error: AxiosError): Promise<any> {
    const config = error.config as AxiosRequestConfig & { _retryCount?: number };
    
    if (!config) {
      return Promise.reject(error);
    }

    // Initialize retry count
    config._retryCount = config._retryCount || 0;

    // Check if we should retry
    const shouldRetry = 
      config._retryCount < this.maxRetries &&
      this.isRetryableError(error);

    if (!shouldRetry) {
      return Promise.reject(error);
    }

    // Increment retry count
    config._retryCount++;

    // Calculate delay with exponential backoff
    const delay = this.retryDelay * Math.pow(2, config._retryCount - 1);

    console.log(`[API] Retrying request (${config._retryCount}/${this.maxRetries}) after ${delay}ms`);

    // Wait before retry
    await this.sleep(delay);

    // Retry request
    return this.client.request(config);
  }

  private isRetryableError(error: AxiosError): boolean {
    // Retry on network errors
    if (!error.response) {
      return true;
    }

    // Retry on specific status codes
    const retryableStatusCodes = [408, 429, 500, 502, 503, 504];
    return retryableStatusCodes.includes(error.response.status);
  }

  private sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  private generateRequestID(): string {
    return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  // Public methods
  public async get<T>(url: string, config?: AxiosRequestConfig): Promise<APIResponse<T>> {
    const response = await this.client.get<APIResponse<T>>(url, config);
    return response.data;
  }

  public async post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<APIResponse<T>> {
    const response = await this.client.post<APIResponse<T>>(url, data, config);
    return response.data;
  }

  public async put<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<APIResponse<T>> {
    const response = await this.client.put<APIResponse<T>>(url, data, config);
    return response.data;
  }

  public async delete<T>(url: string, config?: AxiosRequestConfig): Promise<APIResponse<T>> {
    const response = await this.client.delete<APIResponse<T>>(url, config);
    return response.data;
  }
}

// Gebruik
const api = new APIClient('http://localhost:8080/api/v1', process.env.REACT_APP_API_KEY);

// Voorbeelden
const articles = await api.get<Article[]>('/articles', {
  params: { limit: 20, source: 'nu.nl' }
});

const scrapeResult = await api.post<ScrapeResponse>('/scrape', {
  source: 'nu.nl'
});
```

### 2. Circuit Breaker Pattern

```typescript
class CircuitBreaker {
  private failures = 0;
  private successCount = 0;
  private lastFailureTime?: Date;
  private state: 'CLOSED' | 'OPEN' | 'HALF_OPEN' = 'CLOSED';

  constructor(
    private readonly threshold: number = 5,
    private readonly timeout: number = 60000,
    private readonly monitoringPeriod: number = 10000
  ) {}

  async execute<T>(fn: () => Promise<T>): Promise<T> {
    if (this.state === 'OPEN') {
      if (this.shouldAttemptReset()) {
        this.state = 'HALF_OPEN';
      } else {
        throw new Error('Circuit breaker is OPEN');
      }
    }

    try {
      const result = await fn();
      this.onSuccess();
      return result;
    } catch (error) {
      this.onFailure();
      throw error;
    }
  }

  private onSuccess() {
    this.failures = 0;
    
    if (this.state === 'HALF_OPEN') {
      this.successCount++;
      if (this.successCount >= 2) {
        this.state = 'CLOSED';
        this.successCount = 0;
      }
    }
  }

  private onFailure() {
    this.failures++;
    this.lastFailureTime = new Date();
    this.successCount = 0;

    if (this.failures >= this.threshold) {
      this.state = 'OPEN';
      console.error(`Circuit breaker opened after ${this.failures} failures`);
    }
  }

  private shouldAttemptReset(): boolean {
    return (
      this.lastFailureTime !== undefined &&
      Date.now() - this.lastFailureTime.getTime() >= this.timeout
    );
  }

  public getState() {
    return {
      state: this.state,
      failures: this.failures,
      successCount: this.successCount,
      lastFailureTime: this.lastFailureTime,
    };
  }
}

// Gebruik met API client
class ResilientAPIClient extends APIClient {
  private circuitBreaker = new CircuitBreaker(5, 60000);

  public async get<T>(url: string, config?: AxiosRequestConfig): Promise<APIResponse<T>> {
    return this.circuitBreaker.execute(() => super.get<T>(url, config));
  }

  public getCircuitBreakerState() {
    return this.circuitBreaker.getState();
  }
}
```

### 3. Request Deduplication

```typescript
class RequestDeduplicator {
  private pendingRequests = new Map<string, Promise<any>>();

  async execute<T>(key: string, fn: () => Promise<T>): Promise<T> {
    // Check if request is already pending
    if (this.pendingRequests.has(key)) {
      console.log(`[Dedup] Reusing pending request: ${key}`);
      return this.pendingRequests.get(key) as Promise<T>;
    }

    // Execute new request
    const promise = fn().finally(() => {
      // Clean up after request completes
      this.pendingRequests.delete(key);
    });

    this.pendingRequests.set(key, promise);
    return promise;
  }

  clear() {
    this.pendingRequests.clear();
  }
}

// Gebruik in API client
const deduplicator = new RequestDeduplicator();

async function fetchArticles(params: ArticleFilter) {
  const key = `articles:${JSON.stringify(params)}`;
  return deduplicator.execute(key, () => 
    api.get<Article[]>('/articles', { params })
  );
}
```

---

## Performance Optimalisatie

### 1. Intelligent Caching Strategy

```typescript
import { QueryClient, QueryFunction } from '@tanstack/react-query';

// Custom cache configuratie per endpoint type
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minuten default
      cacheTime: 30 * 60 * 1000, // 30 minuten in cache
      retry: 3,
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    },
  },
});

// Custom stale times per data type
const STALE_TIMES = {
  articles: 5 * 60 * 1000,      // 5 min
  stats: 2 * 60 * 1000,          // 2 min
  trending: 1 * 60 * 1000,       // 1 min
  sentiment: 5 * 60 * 1000,      // 5 min
  health: 30 * 1000,             // 30 sec
  sources: 60 * 60 * 1000,       // 1 uur (changes rarely)
  categories: 10 * 60 * 1000,    // 10 min
};

// Hook voorbeelden met optimale cache settings
export const useArticles = (filter: ArticleFilter) => {
  return useQuery({
    queryKey: ['articles', filter],
    queryFn: () => api.get<Article[]>('/articles', { params: filter }),
    staleTime: STALE_TIMES.articles,
    // Prefetch volgende pagina
    onSuccess: (data) => {
      if (data.meta?.pagination?.has_next) {
        const nextFilter = {
          ...filter,
          offset: (filter.offset || 0) + (filter.limit || 50),
        };
        queryClient.prefetchQuery({
          queryKey: ['articles', nextFilter],
          queryFn: () => api.get<Article[]>('/articles', { params: nextFilter }),
        });
      }
    },
  });
};

export const useTrendingTopics = (hours = 24, minArticles = 3) => {
  return useQuery({
    queryKey: ['trending', hours, minArticles],
    queryFn: () => api.get<TrendingTopicsResponse>('/ai/trending', {
      params: { hours, min_articles: minArticles }
    }),
    staleTime: STALE_TIMES.trending,
    refetchInterval: STALE_TIMES.trending, // Auto-refresh
  });
};
```

### 2. Optimistic Updates

```typescript
import { useMutation, useQueryClient } from '@tanstack/react-query';

// Voorbeeld: Trigger scraping met optimistic update
export const useTriggerScrape = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (source?: string) =>
      api.post<ScrapeResponse>('/scrape', { source }),
    
    // Optimistic update
    onMutate: async (source) => {
      // Cancel outgoing queries
      await queryClient.cancelQueries({ queryKey: ['articles'] });

      // Snapshot previous value
      const previousArticles = queryClient.getQueryData(['articles']);

      // Optimistically update UI
      queryClient.setQueryData(['scrapeStatus'], {
        status: 'running',
        source,
        started_at: new Date().toISOString(),
      });

      return { previousArticles };
    },

    // On success, invalidate and refetch
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['articles'] });
      queryClient.invalidateQueries({ queryKey: ['stats'] });
      queryClient.setQueryData(['scrapeStatus'], data.data);
    },

    // On error, roll back
    onError: (err, variables, context) => {
      if (context?.previousArticles) {
        queryClient.setQueryData(['articles'], context.previousArticles);
      }
      queryClient.setQueryData(['scrapeStatus'], {
        status: 'failed',
        error: err.message,
      });
    },
  });
};
```

### 3. Virtualized Lists

```typescript
import { useVirtualizer } from '@tanstack/react-virtual';
import { useRef } from 'react';

interface VirtualizedArticleListProps {
  articles: Article[];
  onLoadMore: () => void;
  hasMore: boolean;
}

export const VirtualizedArticleList = ({ 
  articles, 
  onLoadMore, 
  hasMore 
}: VirtualizedArticleListProps) => {
  const parentRef = useRef<HTMLDivElement>(null);

  const virtualizer = useVirtualizer({
    count: hasMore ? articles.length + 1 : articles.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 200, // Estimated height per row
    overscan: 5,
  });

  const items = virtualizer.getVirtualItems();

  // Load more when near bottom
  useEffect(() => {
    const lastItem = items[items.length - 1];
    if (!lastItem) return;

    if (
      lastItem.index >= articles.length - 1 &&
      hasMore &&
      !isLoading
    ) {
      onLoadMore();
    }
  }, [hasMore, onLoadMore, articles.length, items]);

  return (
    <div
      ref={parentRef}
      style={{
        height: '600px',
        overflow: 'auto',
      }}
    >
      <div
        style={{
          height: `${virtualizer.getTotalSize()}px`,
          width: '100%',
          position: 'relative',
        }}
      >
        {items.map((virtualRow) => {
          const isLoaderRow = virtualRow.index > articles.length - 1;
          const article = articles[virtualRow.index];

          return (
            <div
              key={virtualRow.index}
              style={{
                position: 'absolute',
                top: 0,
                left: 0,
                width: '100%',
                height: `${virtualRow.size}px`,
                transform: `translateY(${virtualRow.start}px)`,
              }}
            >
              {isLoaderRow ? (
                <LoadingIndicator />
              ) : (
                <ArticleCard article={article} />
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
};
```

---

## Real-time Updates & Polling

### 1. Smart Polling Strategy

```typescript
import { useQuery } from '@tanstack/react-query';
import { useEffect, useState } from 'react';

// Adaptive polling - sneller bij activiteit, langzamer bij inactiviteit
export const useAdaptivePolling = <T>(
  queryKey: any[],
  queryFn: QueryFunction<T>,
  options?: {
    minInterval?: number;
    maxInterval?: number;
    activityThreshold?: number;
  }
) => {
  const {
    minInterval = 5000,
    maxInterval = 60000,
    activityThreshold = 300000, // 5 min
  } = options || {};

  const [interval, setInterval] = useState(minInterval);
  const [lastActivity, setLastActivity] = useState(Date.now());

  // Update activity timestamp on user interaction
  useEffect(() => {
    const handleActivity = () => setLastActivity(Date.now());
    
    window.addEventListener('mousemove', handleActivity);
    window.addEventListener('keydown', handleActivity);
    window.addEventListener('click', handleActivity);

    return () => {
      window.removeEventListener('mousemove', handleActivity);
      window.removeEventListener('keydown', handleActivity);
      window.removeEventListener('click', handleActivity);
    };
  }, []);

  // Adjust interval based on activity
  useEffect(() => {
    const timeSinceActivity = Date.now() - lastActivity;
    
    if (timeSinceActivity < activityThreshold) {
      setInterval(minInterval);
    } else {
      setInterval(maxInterval);
    }
  }, [lastActivity, minInterval, maxInterval, activityThreshold]);

  return useQuery({
    queryKey,
    queryFn,
    refetchInterval: interval,
    refetchIntervalInBackground: false,
  });
};

// Gebruik
const TrendingWidget = () => {
  const { data } = useAdaptivePolling(
    ['trending'],
    () => api.get<TrendingTopicsResponse>('/ai/trending'),
    {
      minInterval: 30000,  // 30 sec when active
      maxInterval: 300000, // 5 min when inactive
    }
  );

  return <TrendingTopicsList topics={data?.data?.topics || []} />;
};
```

### 2. WebSocket Support (Future Enhancement)

```typescript
// WebSocket client voor real-time updates
class WebSocketClient {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private listeners = new Map<string, Set<(data: any) => void>>();

  constructor(private url: string) {
    this.connect();
  }

  private connect() {
    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => {
      console.log('[WS] Connected');
      this.reconnectAttempts = 0;
    };

    this.ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        this.handleMessage(message);
      } catch (error) {
        console.error('[WS] Failed to parse message:', error);
      }
    };

    this.ws.onerror = (error) => {
      console.error('[WS] Error:', error);
    };

    this.ws.onclose = () => {
      console.log('[WS] Disconnected');
      this.attemptReconnect();
    };
  }

  private handleMessage(message: { type: string; data: any }) {
    const listeners = this.listeners.get(message.type);
    if (listeners) {
      listeners.forEach(listener => listener(message.data));
    }
  }

  private attemptReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('[WS] Max reconnect attempts reached');
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

    console.log(`[WS] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
    
    setTimeout(() => this.connect(), delay);
  }

  public subscribe(type: string, callback: (data: any) => void) {
    if (!this.listeners.has(type)) {
      this.listeners.set(type, new Set());
    }
    this.listeners.get(type)!.add(callback);

    // Return unsubscribe function
    return () => {
      const listeners = this.listeners.get(type);
      if (listeners) {
        listeners.delete(callback);
      }
    };
  }

  public send(type: string, data: any) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, data }));
    }
  }

  public disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}

// React hook voor WebSocket
export const useWebSocket = (url: string) => {
  const [client] = useState(() => new WebSocketClient(url));

  useEffect(() => {
    return () => client.disconnect();
  }, [client]);

  return client;
};

// Gebruik
const RealtimeTrending = () => {
  const ws = useWebSocket('ws://localhost:8080/ws/trending');
  const [topics, setTopics] = useState<TrendingTopic[]>([]);

  useEffect(() => {
    return ws.subscribe('trending_update', (data: TrendingTopic[]) => {
      setTopics(data);
    });
  }, [ws]);

  return <TrendingTopicsList topics={topics} />;
};
```

---

## Error Recovery Strategieën

### 1. Graceful Degradation

```typescript
// Component met graceful degradation
const ArticleDetailWithAI = ({ articleId }: { articleId: number }) => {
  const { data: article, error: articleError } = useQuery({
    queryKey: ['article', articleId],
    queryFn: () => api.get<Article>(`/articles/${articleId}`),
  });

  const { data: aiEnrichment, error: aiError } = useQuery({
    queryKey: ['enrichment', articleId],
    queryFn: () => api.get<AIEnrichment>(`/articles/${articleId}/enrichment`),
    // Don't fail the whole component if AI is unavailable
    retry: 1,
    enabled: !!article,
  });

  if (articleError) {
    return <ErrorState error={articleError} />;
  }

  if (!article) {
    return <ArticleSkeleton />;
  }

  return (
    <div className="article-detail">
      {/* Core content always renders */}
      <ArticleContent article={article.data} />

      {/* AI features gracefully degrade */}
      {aiEnrichment?.data ? (
        <>
          {aiEnrichment.data.sentiment && (
            <SentimentBadge sentiment={aiEnrichment.data.sentiment} />
          )}
          {aiEnrichment.data.entities && (
            <EntitiesSection entities={aiEnrichment.data.entities} />
          )}
          {aiEnrichment.data.keywords && (
            <KeywordsSection keywords={aiEnrichment.data.keywords} />
          )}
        </>
      ) : aiError ? (
        <div className="ai-unavailable">
          <InfoIcon />
          <span>AI insights tijdelijk niet beschikbaar</span>
        </div>
      ) : (
        <AILoadingSkeleton />
      )}
    </div>
  );
};
```

### 2. Offline Support

```typescript
import { useEffect, useState } from 'react';

export const useOnlineStatus = () => {
  const [isOnline, setIsOnline] = useState(navigator.onLine);

  useEffect(() => {
    const handleOnline = () => setIsOnline(true);
    const handleOffline = () => setIsOnline(false);

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);

  return isOnline;
};

// Offline indicator component
export const OfflineIndicator = () => {
  const isOnline = useOnlineStatus();

  if (isOnline) return null;

  return (
    <div className="offline-banner">
      <WifiOffIcon />
      <span>Je bent offline. Sommige functies zijn niet beschikbaar.</span>
    </div>
  );
};

// Automatic retry when coming back online
export const useAutoRetry = () => {
  const queryClient = useQueryClient();
  const isOnline = useOnlineStatus();
  const wasOffline = useRef(!isOnline);

  useEffect(() => {
    if (wasOffline.current && isOnline) {
      // Just came back online - retry failed queries
      queryClient.refetchQueries({
        predicate: (query) => query.state.status === 'error',
      });
    }
    wasOffline.current = !isOnline;
  }, [isOnline, queryClient]);
};
```

---

## Testing Strategieën

### 1. Mock API voor Development

```typescript
import { rest } from 'msw';
import { setupServer } from 'msw/node';

// Mock data
const mockArticles: Article[] = [
  {
    id: 1,
    title: 'Test Article 1',
    summary: 'This is a test article',
    url: 'https://example.com/1',
    published: '2025-01-28T10:00:00Z',
    source: 'nu.nl',
    keywords: ['test', 'mock'],
    image_url: 'https://example.com/image1.jpg',
    author: 'Test Author',
    category: 'News',
    created_at: '2025-01-28T10:00:00Z',
    updated_at: '2025-01-28T10:00:00Z',
  },
  // ... more mock articles
];

// Mock handlers
export const handlers = [
  // Articles list
  rest.get('/api/v1/articles', (req, res, ctx) => {
    const limit = Number(req.url.searchParams.get('limit')) || 50;
    const offset = Number(req.url.searchParams.get('offset')) || 0;
    const source = req.url.searchParams.get('source');

    let filtered = mockArticles;
    if (source) {
      filtered = filtered.filter(a => a.source === source);
    }

    const paginated = filtered.slice(offset, offset + limit);

    return res(
      ctx.status(200),
      ctx.json({
        success: true,
        data: paginated,
        meta: {
          pagination: {
            total: filtered.length,
            limit,
            offset,
            current_page: Math.floor(offset / limit) + 1,
            total_pages: Math.ceil(filtered.length / limit),
            has_next: offset + limit < filtered.length,
            has_prev: offset > 0,
          },
        },
        request_id: 'mock-request-id',
        timestamp: new Date().toISOString(),
      })
    );
  }),

  // Health check
  rest.get('/health', (req, res, ctx) => {
    return res(
      ctx.status(200),
      ctx.json({
        success: true,
        data: {
          status: 'healthy',
          timestamp: new Date().toISOString(),
          version: '1.0.0',
          uptime_seconds: 3600,
          components: {
            database: { status: 'healthy' },
            redis: { status: 'healthy' },
            scraper: { status: 'healthy' },
          },
          metrics: {},
        },
        request_id: 'mock-request-id',
        timestamp: new Date().toISOString(),
      })
    );
  }),

  // Simulate slow response
  rest.get('/api/v1/articles/search', async (req, res, ctx) => {
    await new Promise(resolve => setTimeout(resolve, 1000));
    return res(
      ctx.status(200),
      ctx.json({
        success: true,
        data: mockArticles.slice(0, 5),
        request_id: 'mock-request-id',
        timestamp: new Date().toISOString(),
      })
    );
  }),

  // Simulate error
  rest.post('/api/v1/scrape', (req, res, ctx) => {
    return res(
      ctx.status(500),
      ctx.json({
        success: false,
        error: {
          code: 'SCRAPING_FAILED',
          message: 'Failed to scrape source',
          details: 'Connection timeout',
        },
        request_id: 'mock-request-id',
        timestamp: new Date().toISOString(),
      })
    );
  }),
];

// Setup server
export const server = setupServer(...handlers);
```

### 2. Integration Tests

```typescript
import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { server } from './mocks/server';

// Enable API mocking before tests
beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('ArticleList', () => {
  it('loads and displays articles', async () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    });

    render(
      <QueryClientProvider client={queryClient}>
        <ArticleList />
      </QueryClientProvider>
    );

    // Wait for loading to finish
    await waitFor(() => {
      expect(screen.queryByTestId('loading')).not.toBeInTheDocument();
    });

    // Check articles are rendered
    expect(screen.getByText('Test Article 1')).toBeInTheDocument();
  });

  it('handles errors gracefully', async () => {
    // Override handler to return error
    server.use(
      rest.get('/api/v1/articles', (req, res, ctx) => {
        return res(
          ctx.status(500),
          ctx.json({
            success: false,
            error: {
              code: 'DATABASE_ERROR',
              message: 'Database connection failed',
            },
          })
        );
      })
    );

    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    });

    render(
      <QueryClientProvider client={queryClient}>
        <ArticleList />
      </QueryClientProvider>
    );

    await waitFor(() => {
      expect(screen.getByText(/Database connection failed/i)).toBeInTheDocument();
    });
  });
});
```

---

## Deployment Checklist

### Pre-Deployment

- [ ] **Environment Variables**
  ```bash
  REACT_APP_API_URL=https://api.example.com
  REACT_APP_API_KEY=your-production-key
  REACT_APP_ENABLE_ANALYTICS=true
  REACT_APP_SENTRY_DSN=your-sentry-dsn
  ```

- [ ] **API Configuration**
  - [ ] Update base URL naar productie API
  - [ ] Configureer API keys
  - [ ] Test alle endpoints
  - [ ] Verificeer CORS settings

- [ ] **Performance**
  - [ ] Enable production build
  - [ ] Configure code splitting
  - [ ] Enable compression (gzip/brotli)
  - [ ] Optimize bundle size
  - [ ] Configure CDN voor static assets

- [ ] **Monitoring**
  - [ ] Setup error tracking (Sentry)
  - [ ] Configure analytics
  - [ ] Setup performance monitoring
  - [ ] Configure health check monitoring

- [ ] **Security**
  - [ ] Review API key handling
  - [ ] Configure CSP headers
  - [ ] Enable HTTPS only
  - [ ] Review CORS settings
  - [ ] Sanitize user inputs

### Post-Deployment

- [ ] **Verification**
  - [ ] Test health endpoints
  - [ ] Verify API connectivity
  - [ ] Test core functionality
  - [ ] Check error tracking
  - [ ] Verify analytics

- [ ] **Monitoring**
  - [ ] Setup alerts voor errors
  - [ ] Monitor API response times
  - [ ] Track user engagement
  - [ ] Monitor cache hit rates

---

## Best Practices Samenvatting

### DO ✅

1. **Use TypeScript** voor type safety
2. **Implement retry logic** voor netwerk fouten
3. **Cache intelligently** met verschillende TTLs
4. **Handle errors gracefully** met fallbacks
5. **Monitor health** van de API
6. **Use pagination** voor grote datasets
7. **Implement loading states** voor alle async operations
8. **Track request IDs** voor debugging
9. **Respect rate limits** via headers
10. **Test with mock data** in development

### DON'T ❌

1. **Don't ignore error responses** - handle alle error codes
2. **Don't poll too aggressively** - respect server resources
3. **Don't cache everything** - sommige data moet fresh zijn
4. **Don't expose API keys** in client code
5. **Don't ignore health check failures** - implement fallbacks
6. **Don't retry non-retryable errors** (4xx errors)
7. **Don't block UI** tijdens API calls
8. **Don't forget to cleanup** subscriptions/timers
9. **Don't skip validation** van API responses
10. **Don't assume network availability** - handle offline

---

## Resources

- **API Documentation**: [`FRONTEND_API.md`](FRONTEND_API.md)
- **AI API Documentation**: [`FRONTEND_AI_API.md`](FRONTEND_AI_API.md)
- **Backend Documentation**: [`AI_PROCESSING.md`](AI_PROCESSING.md)
- **Quick Start**: [`AI_QUICKSTART.md`](AI_QUICKSTART.md)

---

**Status**: ✅ Complete geavanceerde frontend guide  
**Versie**: 1.0.0  
**Laatste Update**: 2025-01-28