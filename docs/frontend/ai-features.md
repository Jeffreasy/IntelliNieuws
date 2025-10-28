# Frontend AI API Documentation

Complete documentatie voor het gebruik van AI-verrijkte data in je frontend applicatie.

## 📋 Inhoudsopgave

1. [Overzicht](#overzicht)
2. [AI Endpoints](#ai-endpoints)
3. [TypeScript Types](#typescript-types)
4. [Frontend Implementatie](#frontend-implementatie)
5. [Use Cases & Voorbeelden](#use-cases--voorbeelden)
6. [UI Component Voorbeelden](#ui-component-voorbeelden)
7. [Best Practices](#best-practices)

## Overzicht

De API biedt nu AI-verrijkte data voor elk artikel:
- **Sentiment Analysis** - Emotionele toon detectie
- **Entity Extraction** - Personen, organisaties, locaties
- **Category Classification** - Automatische categorisering
- **Keyword Extraction** - Relevante keywords met scores
- **Trending Topics** - Real-time trending onderwerpen
- **AI-generated Summaries** - Korte samenvattingen (optioneel)

## AI Endpoints

### 1. Get Article AI Enrichment

Haal AI-verrijkte data op voor een specifiek artikel.

**Endpoint:** `GET /api/v1/articles/:id/enrichment`

**Authenticatie:** Niet vereist (public endpoint)

**Response:**
```json
{
  "success": true,
  "data": {
    "processed": true,
    "processed_at": "2025-01-28T14:30:00Z",
    "sentiment": {
      "score": 0.65,
      "label": "positive",
      "confidence": 0.89
    },
    "categories": {
      "Politics": 0.92,
      "Economy": 0.45
    },
    "entities": {
      "persons": ["Mark Rutte", "Geert Wilders"],
      "organizations": ["VVD", "PVV", "Tweede Kamer"],
      "locations": ["Den Haag", "Nederland"]
    },
    "keywords": [
      {"word": "verkiezingen", "score": 0.95},
      {"word": "coalitie", "score": 0.88},
      {"word": "formatie", "score": 0.82}
    ],
    "summary": "AI-gegenereerde samenvatting van het artikel..."
  },
  "request_id": "abc123",
  "timestamp": "2025-01-28T14:30:00Z"
}
```

**Voorbeeld:**
```javascript
const getArticleAI = async (articleId) => {
  const response = await fetch(
    `http://localhost:8080/api/v1/articles/${articleId}/enrichment`
  );
  return response.json();
};
```

---

### 2. Get Sentiment Statistics

Haal sentiment statistieken op, optioneel gefilterd op bron en datumrange.

**Endpoint:** `GET /api/v1/ai/sentiment/stats`

**Query Parameters:**
- `source` (string, optioneel) - Filter op nieuwsbron (bijv. "nu.nl")
- `start_date` (RFC3339, optioneel) - Start datum
- `end_date` (RFC3339, optioneel) - Eind datum

**Response:**
```json
{
  "success": true,
  "data": {
    "total_articles": 150,
    "positive_count": 45,
    "neutral_count": 80,
    "negative_count": 25,
    "average_sentiment": 0.15,
    "most_positive_title": "Economie groeit sneller dan verwacht",
    "most_negative_title": "Grote zorgen over klimaatverandering"
  },
  "request_id": "abc123",
  "timestamp": "2025-01-28T14:30:00Z"
}
```

**Voorbeelden:**
```javascript
// Algemene stats
const getSentimentStats = async () => {
  const response = await fetch(
    'http://localhost:8080/api/v1/ai/sentiment/stats'
  );
  return response.json();
};

// Gefilterd op bron
const getSentimentStatsBySource = async (source) => {
  const response = await fetch(
    `http://localhost:8080/api/v1/ai/sentiment/stats?source=${source}`
  );
  return response.json();
};

// Met datumrange
const getSentimentStatsByDateRange = async (startDate, endDate) => {
  const params = new URLSearchParams({
    start_date: startDate,
    end_date: endDate
  });
  const response = await fetch(
    `http://localhost:8080/api/v1/ai/sentiment/stats?${params}`
  );
  return response.json();
};
```

---

### 3. Get Trending Topics

Haal trending onderwerpen op gebaseerd op keyword frequency en recency.

**Endpoint:** `GET /api/v1/ai/trending`

**Query Parameters:**
- `hours` (integer, default: 24) - Kijk terug X uur
- `min_articles` (integer, default: 3) - Minimum aantal artikelen per topic

**Response:**
```json
{
  "success": true,
  "data": {
    "topics": [
      {
        "keyword": "verkiezingen",
        "article_count": 15,
        "average_sentiment": 0.2,
        "sources": ["nu.nl", "ad.nl", "nos.nl"]
      },
      {
        "keyword": "klimaat",
        "article_count": 12,
        "average_sentiment": -0.3,
        "sources": ["nu.nl", "nos.nl"]
      }
    ],
    "hours_back": 24,
    "min_articles": 3,
    "count": 2
  },
  "request_id": "abc123",
  "timestamp": "2025-01-28T14:30:00Z"
}
```

**Voorbeelden:**
```javascript
// Laatste 24 uur
const getTrendingTopics = async () => {
  const response = await fetch(
    'http://localhost:8080/api/v1/ai/trending?hours=24&min_articles=3'
  );
  return response.json();
};

// Laatste week
const getTrendingTopicsWeek = async () => {
  const response = await fetch(
    'http://localhost:8080/api/v1/ai/trending?hours=168&min_articles=5'
  );
  return response.json();
};
```

---

### 4. Get Articles by Entity

Haal artikelen op die een specifieke persoon, organisatie of locatie vermelden.

**Endpoint:** `GET /api/v1/ai/entity/:name`

**Query Parameters:**
- `type` (string, optioneel) - Entity type: "persons", "organizations", "locations"
- `limit` (integer, default: 50, max: 100) - Aantal resultaten

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 123,
      "title": "Article over Mark Rutte",
      "summary": "...",
      "url": "https://...",
      "published": "2025-01-28T12:00:00Z",
      "source": "nu.nl",
      // ... standaard article velden
    }
  ],
  "meta": {
    "pagination": {
      "total": 25,
      "limit": 50,
      "offset": 0
    },
    "filtering": {
      "search": "Mark Rutte"
    }
  },
  "request_id": "abc123",
  "timestamp": "2025-01-28T14:30:00Z"
}
```

**Voorbeelden:**
```javascript
// Zoek artikelen over persoon
const getArticlesByPerson = async (personName) => {
  const encoded = encodeURIComponent(personName);
  const response = await fetch(
    `http://localhost:8080/api/v1/ai/entity/${encoded}?type=persons&limit=20`
  );
  return response.json();
};

// Zoek artikelen over organisatie
const getArticlesByOrg = async (orgName) => {
  const encoded = encodeURIComponent(orgName);
  const response = await fetch(
    `http://localhost:8080/api/v1/ai/entity/${encoded}?type=organizations`
  );
  return response.json();
};

// Zoek artikelen over locatie
const getArticlesByLocation = async (location) => {
  const encoded = encodeURIComponent(location);
  const response = await fetch(
    `http://localhost:8080/api/v1/ai/entity/${encoded}?type=locations`
  );
  return response.json();
};
```

---

### 5. Get AI Processor Status

Haal de status op van de background AI processor.

**Endpoint:** `GET /api/v1/ai/processor/stats`

**Authenticatie:** Niet vereist

**Response:**
```json
{
  "success": true,
  "data": {
    "is_running": true,
    "process_count": 145,
    "last_run": "2025-01-28T14:35:00Z"
  },
  "request_id": "abc123",
  "timestamp": "2025-01-28T14:30:00Z"
}
```

---

### 6. Process Article (Protected)

Trigger AI processing voor een specifiek artikel.

**Endpoint:** `POST /api/v1/articles/:id/process`

**Authenticatie:** Vereist (API Key)

**Headers:**
```
X-API-Key: your-api-key
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Article processed successfully",
    "article_id": 123,
    "enrichment": {
      // AI enrichment data
    }
  },
  "request_id": "abc123",
  "timestamp": "2025-01-28T14:30:00Z"
}
```

---

### 7. Trigger Batch Processing (Protected)

Start handmatige batch processing van pending artikelen.

**Endpoint:** `POST /api/v1/ai/process/trigger`

**Authenticatie:** Vereist (API Key)

**Headers:**
```
X-API-Key: your-api-key
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Processing completed",
    "total_processed": 10,
    "success_count": 9,
    "failure_count": 1,
    "duration": "45.3s"
  },
  "request_id": "abc123",
  "timestamp": "2025-01-28T14:30:00Z"
}
```

---

## TypeScript Types

```typescript
// AI Enrichment Types
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
  score: number;        // -1.0 (zeer negatief) tot 1.0 (zeer positief)
  label: 'positive' | 'negative' | 'neutral';
  confidence?: number;  // 0.0 tot 1.0
}

interface EntityExtraction {
  persons?: string[];
  organizations?: string[];
  locations?: string[];
}

interface Keyword {
  word: string;
  score: number;  // 0.0 tot 1.0 (relevantie)
}

// Sentiment Statistics
interface SentimentStats {
  total_articles: number;
  positive_count: number;
  neutral_count: number;
  negative_count: number;
  average_sentiment: number;
  most_positive_title?: string;
  most_negative_title?: string;
}

// Trending Topics
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

// Processor Status
interface ProcessorStats {
  is_running: boolean;
  process_count: number;
  last_run: string;
}

// Extended Article met AI data
interface ArticleWithAI extends Article {
  ai_enrichment?: AIEnrichment;
}
```

---

## Frontend Implementatie

### React Hooks Voorbeelden

#### useArticleAI Hook
```typescript
import { useQuery } from '@tanstack/react-query';

export const useArticleAI = (articleId: number) => {
  return useQuery({
    queryKey: ['article-ai', articleId],
    queryFn: async () => {
      const response = await fetch(
        `http://localhost:8080/api/v1/articles/${articleId}/enrichment`
      );
      const data = await response.json();
      if (!data.success) throw new Error(data.error.message);
      return data.data as AIEnrichment;
    },
    staleTime: 5 * 60 * 1000, // 5 minuten
  });
};
```

#### useTrendingTopics Hook
```typescript
export const useTrendingTopics = (hours = 24, minArticles = 3) => {
  return useQuery({
    queryKey: ['trending', hours, minArticles],
    queryFn: async () => {
      const response = await fetch(
        `http://localhost:8080/api/v1/ai/trending?hours=${hours}&min_articles=${minArticles}`
      );
      const data = await response.json();
      if (!data.success) throw new Error(data.error.message);
      return data.data as TrendingTopicsResponse;
    },
    refetchInterval: 5 * 60 * 1000, // Refresh elke 5 minuten
  });
};
```

#### useSentimentStats Hook
```typescript
export const useSentimentStats = (
  source?: string,
  startDate?: string,
  endDate?: string
) => {
  return useQuery({
    queryKey: ['sentiment-stats', source, startDate, endDate],
    queryFn: async () => {
      const params = new URLSearchParams();
      if (source) params.append('source', source);
      if (startDate) params.append('start_date', startDate);
      if (endDate) params.append('end_date', endDate);
      
      const response = await fetch(
        `http://localhost:8080/api/v1/ai/sentiment/stats?${params}`
      );
      const data = await response.json();
      if (!data.success) throw new Error(data.error.message);
      return data.data as SentimentStats;
    },
    staleTime: 2 * 60 * 1000, // 2 minuten
  });
};
```

---

## Use Cases & Voorbeelden

### 1. Sentiment Dashboard

Toon sentiment verdeling over alle artikelen:

```typescript
const SentimentDashboard = () => {
  const { data: stats, isLoading } = useSentimentStats();
  
  if (isLoading) return <Spinner />;
  if (!stats) return null;
  
  const sentimentPercentages = {
    positive: (stats.positive_count / stats.total_articles) * 100,
    neutral: (stats.neutral_count / stats.total_articles) * 100,
    negative: (stats.negative_count / stats.total_articles) * 100,
  };
  
  return (
    <div className="sentiment-dashboard">
      <h2>Sentiment Analyse</h2>
      
      {/* Sentiment Distribution Chart */}
      <PieChart data={[
        { label: 'Positief', value: stats.positive_count, color: '#10b981' },
        { label: 'Neutraal', value: stats.neutral_count, color: '#6b7280' },
        { label: 'Negatief', value: stats.negative_count, color: '#ef4444' },
      ]} />
      
      {/* Average Sentiment Indicator */}
      <SentimentMeter score={stats.average_sentiment} />
      
      {/* Extremes */}
      <div className="sentiment-extremes">
        <div className="positive">
          <h3>Meest Positief</h3>
          <p>{stats.most_positive_title}</p>
        </div>
        <div className="negative">
          <h3>Meest Negatief</h3>
          <p>{stats.most_negative_title}</p>
        </div>
      </div>
    </div>
  );
};
```

### 2. Trending Topics Widget

Toon trending onderwerpen met real-time updates:

```typescript
const TrendingTopicsWidget = () => {
  const { data, isLoading } = useTrendingTopics(24, 3);
  
  if (isLoading) return <Spinner />;
  if (!data) return null;
  
  return (
    <div className="trending-widget">
      <h3>🔥 Trending Now</h3>
      <div className="topics-list">
        {data.topics.map((topic, index) => (
          <TrendingTopicCard 
            key={topic.keyword}
            rank={index + 1}
            topic={topic}
          />
        ))}
      </div>
      <p className="meta">
        Based on {data.topics.reduce((sum, t) => sum + t.article_count, 0)} articles
        in the last {data.hours_back} hours
      </p>
    </div>
  );
};

const TrendingTopicCard = ({ rank, topic }: { rank: number; topic: TrendingTopic }) => {
  const sentimentColor = topic.average_sentiment > 0.2 ? 'green' :
                        topic.average_sentiment < -0.2 ? 'red' : 'gray';
  
  return (
    <Link to={`/search?q=${encodeURIComponent(topic.keyword)}`}>
      <div className="trending-card">
        <span className="rank">#{rank}</span>
        <div className="content">
          <h4>{topic.keyword}</h4>
          <div className="meta">
            <span className="count">{topic.article_count} articles</span>
            <SentimentBadge 
              score={topic.average_sentiment}
              color={sentimentColor}
            />
          </div>
          <div className="sources">
            {topic.sources.map(source => (
              <SourceBadge key={source} name={source} />
            ))}
          </div>
        </div>
      </div>
    </Link>
  );
};
```

### 3. Article Card met AI Enrichment

Toon artikel met sentiment, entities en keywords:

```typescript
const ArticleCard = ({ article }: { article: Article }) => {
  const { data: ai } = useArticleAI(article.id);
  
  return (
    <div className="article-card">
      <img src={article.image_url} alt={article.title} />
      
      <div className="content">
        <div className="header">
          <SourceBadge name={article.source} />
          {ai?.sentiment && (
            <SentimentBadge score={ai.sentiment.score} />
          )}
        </div>
        
        <h3>{article.title}</h3>
        <p>{ai?.summary || article.summary}</p>
        
        {/* Categories */}
        {ai?.categories && (
          <div className="categories">
            {Object.entries(ai.categories)
              .sort(([,a], [,b]) => b - a)
              .slice(0, 2)
              .map(([category, confidence]) => (
                <CategoryBadge 
                  key={category}
                  name={category}
                  confidence={confidence}
                />
              ))}
          </div>
        )}
        
        {/* Entities */}
        {ai?.entities && (
          <div className="entities">
            {ai.entities.persons?.slice(0, 3).map(person => (
              <EntityChip 
                key={person}
                name={person}
                type="person"
                onClick={() => navigateToEntity(person, 'persons')}
              />
            ))}
          </div>
        )}
        
        {/* Keywords */}
        {ai?.keywords && (
          <div className="keywords">
            {ai.keywords.slice(0, 5).map(kw => (
              <KeywordTag 
                key={kw.word}
                keyword={kw.word}
                score={kw.score}
              />
            ))}
          </div>
        )}
        
        <div className="footer">
          <time>{formatDate(article.published)}</time>
          <Link to={`/articles/${article.id}`}>Lees meer →</Link>
        </div>
      </div>
    </div>
  );
};
```

### 4. Entity Explorer

Verken artikelen per persoon, organisatie of locatie:

```typescript
const EntityExplorer = ({ entityName, entityType }: { 
  entityName: string;
  entityType: 'persons' | 'organizations' | 'locations';
}) => {
  const [articles, setArticles] = useState<Article[]>([]);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    const fetchArticles = async () => {
      setLoading(true);
      const encoded = encodeURIComponent(entityName);
      const response = await fetch(
        `http://localhost:8080/api/v1/ai/entity/${encoded}?type=${entityType}&limit=20`
      );
      const data = await response.json();
      if (data.success) {
        setArticles(data.data);
      }
      setLoading(false);
    };
    
    fetchArticles();
  }, [entityName, entityType]);
  
  if (loading) return <Spinner />;
  
  return (
    <div className="entity-explorer">
      <div className="header">
        <EntityIcon type={entityType} />
        <h1>{entityName}</h1>
        <span className="count">{articles.length} artikelen</span>
      </div>
      
      <div className="articles-grid">
        {articles.map(article => (
          <ArticleCard key={article.id} article={article} />
        ))}
      </div>
    </div>
  );
};
```

### 5. Sentiment Timeline

Toon sentiment over tijd:

```typescript
const SentimentTimeline = ({ source }: { source?: string }) => {
  const [dateRange, setDateRange] = useState({
    start: subDays(new Date(), 30),
    end: new Date()
  });
  
  const { data: stats } = useSentimentStats(
    source,
    dateRange.start.toISOString(),
    dateRange.end.toISOString()
  );
  
  return (
    <div className="sentiment-timeline">
      <h2>Sentiment Over Tijd</h2>
      
      {/* Date Range Picker */}
      <DateRangePicker 
        value={dateRange}
        onChange={setDateRange}
      />
      
      {/* Sentiment Line Chart */}
      {stats && (
        <LineChart
          data={prepareTimelineData(stats)}
          yAxis={{ label: 'Sentiment Score', min: -1, max: 1 }}
          xAxis={{ label: 'Datum' }}
        />
      )}
      
      {/* Summary Stats */}
      <div className="summary">
        <StatCard 
          label="Gemiddeld Sentiment"
          value={stats?.average_sentiment.toFixed(2)}
          trend={calculateTrend(stats)}
        />
        <StatCard 
          label="Totaal Artikelen"
          value={stats?.total_articles}
        />
      </div>
    </div>
  );
};
```

---

## UI Component Voorbeelden

### Sentiment Badge
```typescript
const SentimentBadge = ({ score, label }: { 
  score: number;
  label?: string;
}) => {
  const getSentimentConfig = (score: number) => {
    if (score >= 0.2) return { color: 'green', icon: '😊', text: 'Positief' };
    if (score <= -0.2) return { color: 'red', icon: '😟', text: 'Negatief' };
    return { color: 'gray', icon: '😐', text: 'Neutraal' };
  };
  
  const config = getSentimentConfig(score);
  
  return (
    <span className={`sentiment-badge ${config.color}`}>
      <span className="icon">{config.icon}</span>
      <span className="label">{label || config.text}</span>
      <span className="score">({score.toFixed(2)})</span>
    </span>
  );
};
```

### Entity Chip
```typescript
const EntityChip = ({ 
  name, 
  type, 
  onClick 
}: { 
  name: string;
  type: 'person' | 'organization' | 'location';
  onClick?: () => void;
}) => {
  const icons = {
    person: '👤',
    organization: '🏢',
    location: '📍'
  };
  
  return (
    <button 
      className={`entity-chip ${type}`}
      onClick={onClick}
    >
      <span className="icon">{icons[type]}</span>
      <span className="name">{name}</span>
    </button>
  );
};
```

### Keyword Tag
```typescript
const KeywordTag = ({ 
  keyword, 
  score 
}: { 
  keyword: string;
  score: number;
}) => {
  // Size based on relevance score
  const fontSize = `${0.8 + (score * 0.5)}em`;
  const opacity = 0.5 + (score * 0.5);
  
  return (
    <span 
      className="keyword-tag"
      style={{ fontSize, opacity }}
      title={`Relevance: ${(score * 100).toFixed(0)}%`}
    >
      {keyword}
    </span>
  );
};
```

---

## Best Practices

### 1. Caching Strategy
```typescript
// Cache AI enrichment data longer
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minuten voor AI data
      cacheTime: 30 * 60 * 1000, // 30 minuten in cache
    },
  },
});
```

### 2. Error Handling
```typescript
const { data, error, isLoading } = useArticleAI(articleId);

if (error) {
  // Graceful degradation - toon artikel zonder AI data
  return <ArticleCardBasic article={article} />;
}

if (isLoading) {
  // Toon skeleton
  return <ArticleCardSkeleton />;
}

// Toon enriched article
return <ArticleCardEnriched article={article} ai={data} />;
```

### 3. Progressive Enhancement
```typescript
// Toon basis artikel, laad AI data async
const ArticleDetail = ({ article }: { article: Article }) => {
  const { data: ai } = useArticleAI(article.id);
  
  return (
    <div>
      {/* Basis artikel info - altijd beschikbaar */}
      <ArticleContent article={article} />
      
      {/* AI enrichments - progressively enhanced */}
      {ai?.sentiment && <SentimentSection sentiment={ai.sentiment} />}
      {ai?.entities && <EntitiesSection entities={ai.entities} />}
      {ai?.keywords && <KeywordsSection keywords={ai.keywords} />}
      {ai?.summary && <AISummarySection summary={ai.summary} />}
    </div>
  );
};
```

### 4. Real-time Updates
```typescript
// Polling voor trending topics
const { data } = useTrendingTopics(24, 3);

// Of gebruik WebSockets voor real-time
useEffect(() => {
  const ws = new WebSocket('ws://localhost:8080/ws/trending');
  
  ws.onmessage = (event) => {
    const topics = JSON.parse(event.data);
    updateTrendingTopics(topics);
  };
  
  return () => ws.close();
}, []);
```

### 5. Performance Optimizations
```typescript
// Lazy load AI data alleen wanneer nodig
const ArticleCard = ({ article }: { article: Article }) => {
  const [showAI, setShowAI] = useState(false);
  const { data: ai } = useArticleAI(article.id, { enabled: showAI });
  
  return (
    <div>
      <ArticleBasicInfo article={article} />
      
      <button onClick={() => setShowAI(true)}>
        Toon AI Insights
      </button>
      
      {showAI && ai && <AIInsights data={ai} />}
    </div>
  );
};
```

### 6. Accessibility
```typescript
// Sentiment met aria labels
<div 
  className="sentiment-indicator"
  role="img"
  aria-label={`Sentiment: ${sentiment.label} (score: ${sentiment.score})`}
>
  <SentimentIcon score={sentiment.score} />
</div>

// Keyboard navigation voor entities
<button
  className="entity-chip"
  onClick={() => navigateToEntity(entity)}
  onKeyPress={(e) => e.key === 'Enter' && navigateToEntity(entity)}
  tabIndex={0}
>
  {entity}
</button>
```

---

## Checklist voor Frontend Developer

- [ ] Implementeer TypeScript types voor AI data
- [ ] Maak herbruikbare AI UI components (SentimentBadge, EntityChip, etc.)
- [ ] Implementeer sentiment dashboard
- [ ] Implementeer trending topics widget
- [ ] Implementeer entity explorer
- [ ] Voeg AI data toe aan article cards
- [ ] Implementeer progressive enhancement
- [ ] Test graceful degradation (zonder AI data)
- [ ] Implementeer proper error handling
- [ ] Optimaliseer caching strategy
- [ ] Test performance met grote datasets
- [ ] Implementeer accessibility features
- [ ] Test op verschillende schermformaten
- [ ] Documenteer AI components in Storybook

---

## Support & Resources

- **API Documentatie**: [`FRONTEND_API.md`](FRONTEND_API.md)
- **Backend Documentatie**: [`AI_PROCESSING.md`](AI_PROCESSING.md)
- **Quick Start**: [`AI_QUICKSTART.md`](AI_QUICKSTART.md)
- **Health Check**: `GET /health`
- **Processor Status**: `GET /api/v1/ai/processor/stats`

Voor vragen of issues:
1. Check processor status
2. Check article enrichment endpoint
3. Verify AI is enabled in backend (.env)
4. Check server logs met request_id

---

**Status**: ✅ AI API volledig gedocumenteerd en klaar voor frontend integratie  
**Versie**: 1.0.0  
**Laatste Update**: 2025-01-28