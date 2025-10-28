# NieuwsScraper - Agents Mapping & Architectuur

## Overzicht

Dit document brengt alle agents/services in het NieuwsScraper project in kaart, hun verantwoordelijkheden, interacties en onderlinge afhankelijkheden.

## Agent Classificatie

### 🤖 Autonome Agents
Agents die zelfstandig taken uitvoeren met minimale externe tussenkomst.

### 🔄 Processing Agents  
Agents die data transformeren en verrijken.

### 🌐 Interface Agents
Agents die externe interacties afhandelen (API, gebruikers, externe services).

### ⏱️ Scheduling Agents
Agents die periodieke taken beheren.

---

## 1. AI Processing Agents

### 1.1 AI Service (`internal/ai/service.go`)
**Type:** 🔄 Processing Agent  
**Status:** Actief  
**Verantwoordelijkheid:** Centrale coördinatie van alle AI-verwerkingsoperaties

**Kernfunctionaliteiten:**
- **Artikel Processing:** Verwerkt individuele artikelen met AI
- **Batch Processing:** Efficiënte verwerking van meerdere artikelen
- **Pending Articles:** Identificeert en verwerkt onverwerkte artikelen
- **Analytics:** Sentiment statistieken en trending topics
- **Entity Queries:** Zoekt artikelen op basis van entities

**Key Methods:**
```go
ProcessArticle(ctx, articleID) -> *AIEnrichment
ProcessBatch(ctx, articleIDs) -> *BatchProcessingResult
ProcessPendingArticles(ctx, limit) -> *BatchProcessingResult
GetSentimentStats(ctx, source, startDate, endDate) -> *SentimentStats
GetTrendingTopics(ctx, hoursBack, minArticles) -> []TrendingTopic
GetArticlesByEntity(ctx, entityName, entityType, limit) -> []Article
GetEnrichment(ctx, articleID) -> *AIEnrichment
```

**Dependencies:**
- `*pgxpool.Pool` - Database connectie
- `*OpenAIClient` - AI processing via OpenAI
- `*Config` - Configuratie instellingen
- `*logger.Logger` - Logging

**Configuration:**
```go
type Config struct {
    // OpenAI settings
    OpenAIAPIKey    string
    OpenAIModel     string (default: "gpt-3.5-turbo")
    OpenAIMaxTokens int (default: 1000)
    
    // Processing
    Enabled         bool
    AsyncProcessing bool
    BatchSize       int (default: 10)
    ProcessInterval time.Duration (default: 5min)
    RetryFailed     bool
    MaxRetries      int
    
    // Features
    EnableSentiment  bool
    EnableEntities   bool
    EnableCategories bool
    EnableKeywords   bool
    EnableSummary    bool
    EnableSimilarity bool
    
    // Cost control
    MaxDailyCost       float64
    RateLimitPerMinute int
    Timeout            time.Duration
}
```

---

### 1.2 AI Processor (`internal/ai/processor.go`)
**Type:** 🤖 Autonome Agent + ⏱️ Scheduling Agent  
**Status:** Actief (Background Worker)  
**Verantwoordelijkheid:** Automatische background processing van artikelen

**Kernfunctionaliteiten:**
- **Background Loop:** Continu draaien met configurable interval
- **Automatic Processing:** Automatisch verwerken van nieuwe artikelen
- **Manual Triggers:** Ondersteuning voor handmatige triggers
- **Retry Logic:** Opnieuw verwerken van gefaalde artikelen
- **Statistics:** Tracking van verwerkingsstatistieken

**Key Methods:**
```go
Start(ctx) error                              // Start background processing
Stop()                                        // Stop gracefully
IsRunning() bool                              // Status check
GetStats() ProcessorStats                     // Get statistics
ManualTrigger(ctx) -> *BatchProcessingResult  // Manual processing trigger
RetryFailed(ctx, maxRetries) -> *BatchProcessingResult
```

**Processing Loop:**
```
1. Start → Ticker interval (default: 5 minuten)
2. Query pending articles (ai_processed = FALSE)
3. Process batch (default: 10 artikelen)
4. Apply rate limiting
5. Update database met enrichment data
6. Log results en statistics
7. Wait voor next interval
```

**Lifecycle Management:**
```go
// Start processor
processor := ai.NewProcessor(service, config, logger)
processor.Start(ctx)

// Stop gracefully
processor.Stop()

// Manual trigger
result, err := processor.ManualTrigger(ctx)
```

**Statistics Tracking:**
```go
type ProcessorStats struct {
    IsRunning    bool
    ProcessCount int       // Total articles processed
    LastRun      time.Time // Last processing time
}
```

---

### 1.3 OpenAI Client (`internal/ai/openai_client.go`)
**Type:** 🌐 Interface Agent (External API)  
**Status:** Actief  
**Verantwoordelijkheid:** Directe communicatie met OpenAI API

**Kernfunctionaliteiten:**
- **Sentiment Analysis:** Detecteert positief/negatief/neutraal sentiment
- **Entity Extraction:** Extraheert personen, organisaties, locaties
- **Categorization:** Automatische categorie toewijzing
- **Keyword Extraction:** Intelligente keyword extractie
- **Summary Generation:** Genereert korte samenvattingen
- **Comprehensive Processing:** Alle features in één API call

**Key Methods:**
```go
Complete(ctx, messages, temperature) -> *OpenAIResponse
AnalyzeSentiment(ctx, title, content) -> *SentimentAnalysis
ExtractEntities(ctx, title, content) -> *EntityExtraction
CategorizeArticle(ctx, title, content) -> map[string]float64
ExtractKeywords(ctx, title, content) -> []Keyword
GenerateSummary(ctx, title, content) -> string
ProcessArticle(ctx, title, content, opts) -> *AIEnrichment
```

**API Communication:**
```go
// Request format
type OpenAIRequest struct {
    Model       string        "gpt-3.5-turbo" / "gpt-4"
    Messages    []ChatMessage
    Temperature float64       (0.0 - 1.0)
    MaxTokens   int
}

// Response handling
type OpenAIResponse struct {
    ID      string
    Choices []Choice
    Usage   Usage  // Token usage tracking
}
```

**Processing Optimizations:**
- **Single Call Processing:** Alle analyses in één API call (kosten-efficiënt)
- **Text Truncation:** Automatische truncation tot 4000 chars
- **Error Handling:** Robuuste error handling en fallbacks
- **Token Tracking:** Monitoring van API usage

**Cost Management:**
```go
// Token usage tracking
type Usage struct {
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
}

// Geschatte kosten per artikel
GPT-3.5-turbo: ~$0.002 - $0.005
GPT-4:         ~$0.02  - $0.03
```

---

## 2. Scraping Agents

### 2.1 Scraper Service (`internal/scraper/service.go`)
**Type:** 🤖 Autonome Agent  
**Status:** Actief  
**Verantwoordelijkheid:** Coördinatie van alle scraping operaties

**Kernfunctionaliteiten:**
- **Source Management:** Beheer van nieuws bronnen
- **Parallel Scraping:** Gelijktijdig scrapen van meerdere bronnen
- **Rate Limiting:** Respecteert rate limits per domain
- **Robots.txt Checking:** Optionele robots.txt validatie
- **Duplicate Detection:** Voorkomt duplicate artikelen
- **Batch Storage:** Efficiënte database operaties
- **Retry Logic:** Automatische retry met exponential backoff

**Key Methods:**
```go
ScrapeSource(ctx, source, feedURL) -> *ScrapingResult
ScrapeAllSources(ctx) -> map[string]*ScrapingResult
ScrapeWithRetry(ctx, source, feedURL) -> *ScrapingResult
GetStats(ctx) -> map[string]interface{}
```

**Configured Sources:**
```go
var ScrapeSources = map[string]string{
    "nu.nl":  "https://www.nu.nl/rss",
    "ad.nl":  "https://www.ad.nl/rss.xml",
    "nos.nl": "https://feeds.nos.nl/nosnieuwsalgemeen",
}
```

**Scraping Pipeline:**
```
1. Context Validation
   ↓
2. Robots.txt Check (optioneel)
   ↓
3. Rate Limiting (per domain)
   ↓
4. RSS Feed Parsing
   ↓
5. Article Validation & Filtering
   ↓
6. Duplicate Detection
   ↓
7. Batch Database Insert
   ↓
8. Result Logging & Metrics
```

**Result Tracking:**
```go
type ScrapingResult struct {
    Source          string
    StartTime       time.Time
    EndTime         time.Time
    Duration        time.Duration
    Status          string  // running/completed/failed/partial_success
    ArticlesFound   int
    ArticlesStored  int
    ArticlesSkipped int
    Error           string
}
```

**Dependencies:**
- `*rss.Scraper` - RSS parsing
- `*ArticleRepository` - Database operaties
- `*ScraperRateLimiter` - Rate limiting
- `*RobotsChecker` - Robots.txt validatie

---

### 2.2 RSS Scraper (`internal/scraper/rss/rss_scraper.go`)
**Type:** 🔄 Processing Agent  
**Status:** Actief  
**Verantwoordelijkheid:** RSS feed parsing en artikel extractie

**Kernfunctionaliteiten:**
- **RSS/Atom Parsing:** Ondersteunt meerdere feed formats
- **Article Conversion:** Transformeert feed items naar artikelen
- **Content Cleaning:** HTML tag removal en text normalisatie
- **Image Extraction:** Extraheert artikel afbeeldingen
- **Metadata Extraction:** Author, categories, keywords
- **Concurrent Scraping:** Parallel verwerking van feeds

**Key Methods:**
```go
ScrapeFeed(ctx, feedURL, source) -> []*ArticleCreate
ScrapeMultipleFeeds(ctx, feeds) -> (results, errors)
convertFeedItem(item, source) -> *ArticleCreate
```

**Article Extraction:**
```go
type ArticleCreate struct {
    Title     string    // Cleaned title
    Summary   string    // HTML-stripped, truncated to 2000 chars
    URL       string    // Article link
    Published time.Time // Published/updated date
    Source    string    // Source identifier
    Keywords  []string  // Extracted from categories
    ImageURL  string    // First image found
    Author    string    // Author name
    Category  string    // Primary category
}
```

**Content Processing:**
- **HTML Cleaning:** Verwijdert tags, behoudt structuur
- **Text Normalization:** Multiple spaces/newlines cleanup
- **Truncation:** Max 2000 chars voor summary
- **Date Parsing:** Flexible date handling

**Dependencies:**
- `github.com/mmcdole/gofeed` - RSS/Atom parser
- `*RobotsChecker` - URL validation

---

## 3. Scheduling Agent

### 3.1 Scheduler (`internal/scheduler/scheduler.go`)
**Type:** ⏱️ Scheduling Agent + 🤖 Autonome Agent  
**Status:** Actief (Background Worker)  
**Verantwoordelijkheid:** Periodieke scraping orchestration

**Kernfunctionaliteiten:**
- **Periodic Scraping:** Configureerbare scraping interval
- **Lifecycle Management:** Start/stop control
- **Initial Run:** Onmiddellijke uitvoering bij start
- **Context Awareness:** Graceful shutdown op context cancellation
- **Result Aggregation:** Verzamelt en logt scraping resultaten

**Key Methods:**
```go
Start(ctx)                           // Start periodic scraping
Stop()                               // Graceful shutdown
IsRunning() bool                     // Status check
UpdateInterval(interval)             // Update scraping interval
```

**Scheduling Logic:**
```go
// Configuration
interval := 30 * time.Minute  // Default: elke 30 minuten

// Lifecycle
1. Start → Run initial scrape
2. Create ticker with interval
3. Loop:
   - Wait for ticker OR stop signal OR context done
   - Execute scrape for all sources
   - Log aggregated results
4. Stop → Close channels, wait for completion
```

**Operation Flow:**
```
Start
  ↓
Initial Scrape (immediate)
  ↓
Start Ticker (30min interval)
  ↓
┌─────────────────────┐
│  Wait for:          │
│  - Ticker tick      │
│  - Stop signal      │
│  - Context cancel   │
└─────────────────────┘
  ↓
Execute ScrapeAllSources()
  ↓
Log Results:
  - Total stored
  - Total skipped
  - Per-source errors
  - Duration
  ↓
Continue or Stop
```

**Thread Safety:**
```go
type Scheduler struct {
    ticker   *time.Ticker
    stopChan chan struct{}
    wg       sync.WaitGroup
    running  bool
    mu       sync.Mutex
}
```

---

## 4. API Interface Agents

### 4.1 AI Handler (`internal/api/handlers/ai_handler.go`)
**Type:** 🌐 Interface Agent (HTTP API)  
**Status:** Actief  
**Verantwoordelijkheid:** AI functionaliteit via REST API

**Endpoints:**
```go
GET  /api/v1/articles/:id/enrichment    // Get AI enrichment
POST /api/v1/articles/:id/process       // Trigger processing
GET  /api/v1/ai/sentiment/stats         // Sentiment statistics
GET  /api/v1/ai/trending                // Trending topics
GET  /api/v1/ai/entity/:name            // Articles by entity
POST /api/v1/ai/process/trigger         // Manual processing trigger
GET  /api/v1/ai/processor/stats         // Processor statistics
```

**Request/Response Patterns:**
```go
// Get Enrichment
GET /api/v1/articles/123/enrichment
Response: {
    "processed": true,
    "sentiment": {
        "score": 0.65,
        "label": "positive",
        "confidence": 0.85
    },
    "categories": {
        "Politics": 0.89,
        "Economy": 0.45
    },
    "entities": {
        "persons": ["Mark Rutte"],
        "organizations": ["EU"],
        "locations": ["Amsterdam"]
    },
    "keywords": [
        {"word": "verkiezingen", "score": 0.92}
    ]
}

// Trending Topics
GET /api/v1/ai/trending?hours=24&min_articles=3
Response: {
    "topics": [
        {
            "keyword": "klimaat",
            "article_count": 15,
            "average_sentiment": 0.3,
            "sources": ["nu.nl", "nos.nl"]
        }
    ],
    "hours_back": 24,
    "min_articles": 3,
    "count": 10
}
```

**Features:**
- Request validation
- Error handling met gestandaardiseerde responses
- Query parameter parsing
- Cache invalidatie triggers
- Metrics logging

---

### 4.2 Scraper Handler (`internal/api/handlers/scraper_handler.go`)
**Type:** 🌐 Interface Agent (HTTP API)  
**Status:** Actief  
**Verantwoordelijkheid:** Scraping operaties via REST API

**Endpoints:**
```go
POST /api/v1/scrape           // Trigger scraping
GET  /api/v1/sources          // List available sources
GET  /api/v1/scraper/stats    // Scraper statistics
```

**Request/Response Patterns:**
```go
// Trigger Single Source
POST /api/v1/scrape
Body: {"source": "nu.nl"}
Response: {
    "status": "success",
    "source": "nu.nl",
    "articles_found": 50,
    "articles_stored": 45,
    "articles_skipped": 5,
    "duration_seconds": 2.5
}

// Trigger All Sources
POST /api/v1/scrape
Body: {}  // empty = all sources
Response: {
    "total_sources": 3,
    "total_stored": 120,
    "results": [...]
}

// Get Sources
GET /api/v1/sources
Response: [
    {
        "name": "nu.nl",
        "feed_url": "https://www.nu.nl/rss",
        "is_active": true
    }
]
```

**Features:**
- Single/multiple source scraping
- Retry logic
- Cache invalidation na successful scrape
- Real-time progress logging

---

### 4.3 Article Handler (vermeld, niet gedetailleerd bekeken)
**Type:** 🌐 Interface Agent (HTTP API)  
**Status:** Actief  
**Verantwoordelijkheid:** Artikel CRUD operaties

**Vermoedelijke functionaliteit:**
- List articles met filtering/pagination
- Get single article
- Cache management
- Search functionaliteit

---

## 5. Support Services

### 5.1 Article Repository (`internal/repository/article_repository.go`)
**Type:** 🔄 Data Access Agent  
**Verantwoordelijkheid:** Database operaties voor artikelen

**Core Operations:**
- CRUD operaties
- Batch inserts
- Duplicate detection (`ExistsByURL`)
- Statistics queries (`GetStatsBySource`)
- Filtering en pagination

---

### 5.2 Cache Service (`internal/cache/cache_service.go`)
**Type:** 🔄 Performance Agent  
**Verantwoordelijkheid:** Caching voor API responses

**Functionaliteit:**
- Response caching
- TTL management
- Cache invalidatie
- Memory/Redis backing

---

## Agent Interactie Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     USER / CLIENT                            │
└───────────────────────────┬─────────────────────────────────┘
                            │
                    HTTP Requests
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   API HANDLERS (Interface Agents)            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ AI Handler   │  │Scraper Handler│ │Article Handler│      │
│  └──────┬───────┘  └──────┬────────┘  └──────┬───────┘      │
└─────────┼──────────────────┼──────────────────┼──────────────┘
          │                  │                  │
          │                  │                  │
          ▼                  ▼                  ▼
┌─────────────────────────────────────────────────────────────┐
│                 CORE SERVICES (Processing Agents)            │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │           AI Service                                  │   │
│  │  - ProcessArticle()                                   │   │
│  │  - ProcessBatch()                                     │   │
│  │  - GetSentimentStats()                               │   │
│  │  - GetTrendingTopics()                               │   │
│  └────────┬────────────────┬────────────────────────────┘   │
│           │                │                                 │
│           ▼                ▼                                 │
│  ┌────────────────┐  ┌────────────────┐                    │
│  │ AI Processor   │  │ OpenAI Client  │                    │
│  │ (Background)   │  │ (External API) │                    │
│  └────────────────┘  └────────────────┘                    │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │           Scraper Service                             │   │
│  │  - ScrapeSource()                                     │   │
│  │  - ScrapeAllSources()                                 │   │
│  └────────┬───────────────────────────────────────────────  │
│           │                                                   │
│           ▼                                                   │
│  ┌────────────────┐                                          │
│  │  RSS Scraper   │                                          │
│  │  (Feed Parser) │                                          │
│  └────────────────┘                                          │
└───────────────────────────┬───────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│              DATA LAYER (Repository/Cache)                   │
│  ┌──────────────────┐         ┌──────────────────┐          │
│  │Article Repository│         │  Cache Service   │          │
│  └────────┬─────────┘         └──────────────────┘          │
└───────────┼───────────────────────────────────────────────────┘
            │
            ▼
┌─────────────────────────────────────────────────────────────┐
│                    DATABASE (PostgreSQL)                     │
│  - articles table                                            │
│  - AI enrichment columns                                     │
│  - Indexes & optimizations                                   │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│              AUTONOMOUS AGENTS (Background)                  │
│                                                               │
│  ┌────────────────┐         ┌──────────────────┐            │
│  │   Scheduler    │────────▶│ Scraper Service  │            │
│  │ (30min ticker) │         │                  │            │
│  └────────────────┘         └──────────────────┘            │
│                                                               │
│  ┌────────────────┐         ┌──────────────────┐            │
│  │ AI Processor   │────────▶│   AI Service     │            │
│  │ (5min ticker)  │         │                  │            │
│  └────────────────┘         └──────────────────┘            │
└─────────────────────────────────────────────────────────────┘
```

---

## Agent Communicatie Flows

### Flow 1: Scheduled Scraping
```
Scheduler (30min) 
  → Scraper Service.ScrapeAllSources()
    → RSS Scraper.ScrapeFeed() (parallel voor elke source)
      → Article Repository.CreateBatch()
        → Database
  → Article Handler: InvalidateCache()
```

### Flow 2: AI Processing (Background)
```
AI Processor (5min)
  → AI Service.ProcessPendingArticles()
    → Query unprocessed articles
    → AI Service.ProcessArticle() (per artikel)
      → OpenAI Client.ProcessArticle()
        → OpenAI API (external)
      → Parse & validate response
    → Article Repository.Update() (met AI data)
      → Database
```

### Flow 3: Manual Scrape Trigger (API)
```
Client
  → POST /api/v1/scrape
    → Scraper Handler.TriggerScrape()
      → Scraper Service.ScrapeWithRetry()
        → RSS Scraper.ScrapeFeed()
          → Article Repository.CreateBatch()
            → Database
        → Retry logic (bij failures)
      → Article Handler.InvalidateCache()
  ← JSON Response met resultaten
```

### Flow 4: AI Enrichment Query
```
Client
  → GET /api/v1/articles/:id/enrichment
    → AI Handler.GetEnrichment()
      → AI Service.GetEnrichment()
        → Database query
      ← AIEnrichment data
    ← JSON Response
```

### Flow 5: Trending Topics
```
Client
  → GET /api/v1/ai/trending?hours=24
    → AI Handler.GetTrendingTopics()
      → AI Service.GetTrendingTopics()
        → Complex database query:
          - Extract keywords from articles
          - Count occurrences
          - Calculate avg sentiment
          - Group by keyword
        ← []TrendingTopic
      ← JSON Response
```

---

## Agent Configuratie & Dependencies

### Environment Variables per Agent

**AI Service & Processor:**
```env
OPENAI_API_KEY=sk-...
OPENAI_MODEL=gpt-3.5-turbo
OPENAI_MAX_TOKENS=1000

AI_ENABLED=true
AI_ASYNC_PROCESSING=true
AI_BATCH_SIZE=10
AI_PROCESS_INTERVAL_MINUTES=5

AI_ENABLE_SENTIMENT=true
AI_ENABLE_ENTITIES=true
AI_ENABLE_CATEGORIES=true
AI_ENABLE_KEYWORDS=true
AI_ENABLE_SUMMARY=false

AI_MAX_DAILY_COST=10.00
AI_RATE_LIMIT_PER_MINUTE=60
```

**Scraper Service:**
```env
SCRAPER_USER_AGENT=NieuwsScraper/1.0
SCRAPER_RATE_LIMIT_SECONDS=2
SCRAPER_ENABLE_ROBOTS_TXT=true
SCRAPER_ENABLE_DUPLICATE_DETECTION=true
SCRAPER_RETRY_ATTEMPTS=3
SCRAPER_TIMEOUT_SECONDS=30
SCRAPER_TARGET_SITES=nu.nl,ad.nl,nos.nl
```

**Scheduler:**
```env
SCHEDULER_ENABLED=true
SCHEDULER_INTERVAL_MINUTES=30
```

---

## Performance & Monitoring

### Metrics per Agent

**AI Processor:**
- `process_count` - Total processed articles
- `success_count` - Successful processings
- `failure_count` - Failed processings
- `avg_processing_time` - Average time per article
- `last_run` - Last processing timestamp

**Scraper Service:**
- `articles_found` - Total articles discovered
- `articles_stored` - Successfully stored
- `articles_skipped` - Duplicates/invalid
- `scrape_duration` - Time per source
- `error_rate` - Failure percentage

**Scheduler:**
- `total_runs` - Number of scheduled runs
- `last_run_duration` - Duration of last scrape
- `is_running` - Current status

### Logging Strategy

**Per Agent:**
- Component tagging: `logger.WithComponent("agent-name")`
- Error tracking: `logger.WithError(err)`
- Contextual logging: Request IDs, article IDs
- Performance metrics: Durations, counts

---

## Failure Modes & Recovery

### AI Processing Failures
**Failure:** OpenAI API unavailable/rate limited
**Recovery:** 
- Exponential backoff
- Mark as failed in database
- Retry queue voor later
- Graceful degradation (app werkt zonder AI)

### Scraping Failures
**Failure:** Feed unavailable/parse error
**Recovery:**
- Per-source retry logic (3 attempts)
- Continue met andere sources
- Log errors voor monitoring
- Robots.txt respect

### Database Failures
**Failure:** Connection loss/query timeout
**Recovery:**
- Connection pooling
- Automatic reconnection
- Transaction rollbacks
- Error logging

---

## Schaalbaarheid

### Horizontal Scaling Options

**AI Processing:**
- Multiple AI Processor instances
- Work queue distributie (NATS/RabbitMQ)
- Cached results sharing (Redis)

**Scraping:**
- Dedicated scraper instances per source
- Load balancing voor parallel scraping
- Distributed rate limiting

**API:**
- Multiple handler instances
- Load balancer (Nginx/HAProxy)
- Shared cache layer

### Vertical Scaling
- Batch size aanpassingen
- Connection pool tuning
- Memory limits per agent
- CPU allocation

---

## Best Practices

### Agent Development
1. **Lifecycle Management:** Proper Start/Stop methods
2. **Context Awareness:** Respect context cancellation
3. **Thread Safety:** Mutex protection voor shared state
4. **Error Handling:** Comprehensive error recovery
5. **Logging:** Structured logging met context
6. **Configuration:** Environment-based config
7. **Testing:** Unit tests per agent
8. **Metrics:** Prometheus-compatible metrics

### Agent Coordination
1. **Loose Coupling:** Agents via interfaces
2. **Async Communication:** Channels/queues
3. **Rate Limiting:** Respect external APIs
4. **Graceful Shutdown:** Clean resource cleanup
5. **Health Checks:** Status endpoints
6. **Circuit Breakers:** Prevent cascade failures

---

## Toekomstige Agent Uitbreidingen

### Geplande Agents

**1. Content Analyzer Agent**
- Full-text artikel scraping
- Image analysis
- Fact checking integratie

**2. Recommendation Agent**
- Personalized recommendations
- User preference learning
- Collaborative filtering

**3. Notification Agent**
- Real-time alerts
- Custom triggers
- Multi-channel (email/push/SMS)

**4. Archive Agent**
- Historical data management
- Cold storage
- Data retention policies

**5. Analytics Agent**
- Advanced metrics
- Trend prediction
- Anomaly detection

---

## Conclusie

Het NieuwsScraper systeem bestaat uit **8 primaire agents** die samenwerken:

1. **AI Service** - Centrale AI coördinatie
2. **AI Processor** - Background AI processing
3. **OpenAI Client** - External AI API interface
4. **Scraper Service** - Scraping coördinatie
5. **RSS Scraper** - Feed parsing
6. **Scheduler** - Periodieke orchestration
7. **AI Handler** - AI API interface
8. **Scraper Handler** - Scraping API interface

Plus **2 support services**:
- Article Repository - Data access
- Cache Service - Performance optimization

Deze agents vormen een robuust, schaalbaar en onderhoudsbaar systeem voor geautomatiseerde nieuwsverwerking met AI-verrijking.