# 📚 Frontend Documentatie - Complete Gids

**🆕 Database Optimized (2025-10-30):** Analytics endpoints zijn 90% sneller dankzij materialized views! Zie [Database Fixes Complete](../DATABASE-FIXES-COMPLETE.md) voor details.

Welkom bij de complete frontend documentatie voor de IntelliNieuws API. Deze gids helpt je om snel een moderne frontend applicatie te bouwen die volledig geïntegreerd is met onze news aggregation API.

## 🎯 Wat kan de API?

De IntelliNieuws API is een krachtige backend voor nieuws aggregatie met:

### Core Features
- ✅ **Multi-source scraping** - Aggregeert nieuws van Nederlandse bronnen (RSS + HTML + Browser)
- ✅ **AI-powered enrichment** - Sentiment analyse, entity extraction, en categorisatie
- ✅ **Real-time trending** - Detecteert trending onderwerpen op basis van recente artikelen
- ✅ **Full-text search** - Krachtige zoekfunctionaliteit in artikelen
- ✅ **Smart caching** - Redis-backed caching voor optimale performance
- ✅ **Comprehensive health monitoring** - Kubernetes-compatible health checks
- ✅ **Rate limiting** - Bescherming tegen overbelasting
- ✅ **RESTful design** - Standaard HTTP methods en response formats

### NEW: Stock Market Integration (v2.1) ✨
- ✅ **Real-time Stock Quotes** - FMP API voor US aandelen (AAPL, MSFT, GOOGL, etc.)
- ✅ **Company Profiles** - Bedrijfsinformatie en financial data
- ✅ **Earnings Calendar** - Upcoming earnings announcements
- ✅ **Symbol Search** - Zoek bedrijven en stock symbols
- ✅ **Auto-Enrichment** - Automatische stock data bij artikelen
- ✅ **Free Tier** - Werkt met FMP gratis tier (250 calls/dag)

### NEW: Email Integration (v2.1) ✨
- ✅ **Outlook IMAP** - Ontvang emails als nieuws items
- ✅ **Sender Filtering** - Whitelist-based (bijv. noreply@x.ai)
- ✅ **Auto-Processing** - Email → Article conversion
- ✅ **Scheduled Polling** - Configurable interval (5 min default)
- ✅ **AI Enrichment** - Automatic sentiment/entity extraction
- ✅ **Database Tracking** - Complete email metadata

## 📖 Documentatie Structuur

### 🚨 Troubleshooting
- 🔧 **[Troubleshooting: No Data](TROUBLESHOOTING-NO-DATA.md)** - Fix "geen data" problemen in Trending & Sentiment
  - Database health checks
  - Materialized views verification
  - Complete diagnostic script
  - Browser console tests

### NEW: Stock & Email Features (v2.1) ✨

**Stock Market Integration:**
- 💹 **[Stock API Reference](../api/stock-api-reference.md)** - Complete FMP API docs (432 lines)
- 📊 **[Stock Tickers Integration](stock-tickers-integration.md)** - Frontend integration guide
- 💰 **[FMP Free Tier Guide](../FMP-FREE-TIER-FINAL.md)** - Gratis tier setup & limitations
- ⚡ **[FMP Quick Start](../quick-start-fmp.md)** - 5-minute setup

**Email Integration:**
- 📧 **[Email Integration Guide](../features/email-integration.md)** - Complete IMAP setup (471 lines)
- ⚡ **[Email Quick Start](../features/email-quickstart.md)** - 5-minute email setup
- 📝 **[Email Summary](../features/EMAIL-INTEGRATION-SUMMARY.md)** - Implementation details

### Core Documentation

### 1. **[COMPLETE-API-REFERENCE.md](COMPLETE-API-REFERENCE.md)** - Complete API Reference ⭐ NEW
**🆕 Bijgewerkt met laatste database improvements!**

Complete API documentatie met alle endpoints en TypeScript types:
- 📊 **Database Performance:** Materialized views info (90% sneller)
- 🗄️ **Sources Metadata:** Complete source tracking endpoints
- 📈 **Analytics Endpoints:** Trending, sentiment, entities
- 🔄 **Auto-Refresh:** Materialized views management
- 💻 **TypeScript:** Volledige type definities
- 🎣 **React Hooks:** TanStack Query v5 examples
- 📝 **Components:** Complete implementatie voorbeelden

### 2. **[FRONTEND_API.md](FRONTEND_API.md)** - Core API Reference (Legacy)
**Start hier als je nieuw bent!**

Complete API documentatie met alle endpoints, parameters, en response formats:

- 📋 Standaard response formats
- 🔐 Authenticatie & rate limiting
- 📄 Articles endpoints (list, get, search, stats)
- 🔍 Search functionaliteit
- 📊 Statistics & analytics
- 🗂️ Sources & categories
- 🤖 AI enrichment basics
- 💻 TypeScript type definitions
- 🎨 Frontend implementatie voorbeelden

**Perfect voor:**
- Beginnen met de API
- Opzoeken van endpoint details
- Implementeren van basis functionaliteit

### 3. **[FRONTEND_AI_API.md](FRONTEND_AI_API.md)** - AI Features Guide
**Voor geavanceerde AI features.**

Complete gids voor het gebruik van AI-verrijkte data:

- 🎭 **Sentiment Analysis** - Emotionele toon detectie
- 🏷️ **Entity Extraction** - Personen, organisaties, locaties
- 📂 **Category Classification** - Automatische categorisering
- 🔑 **Keyword Extraction** - Relevante keywords met scores
- 📈 **Trending Topics** - Real-time trending onderwerpen
- 📝 **AI-generated Summaries** - Korte samenvattingen

**Inclusief:**
- Complete TypeScript types voor AI data
- React hooks voorbeelden
- UI component voorbeelden
- Use cases & implementaties
- Best practices voor AI features

**Perfect voor:**
- Implementeren van sentiment dashboards
- Trending topics widgets
- Entity-based filtering
- AI-powered article cards

### 4. **[FRONTEND_ADVANCED.md](FRONTEND_ADVANCED.md)** - Advanced Patterns
**Voor production-ready applicaties.**

Geavanceerde patronen en best practices:

- 🏥 **Health Monitoring** - Alle health check endpoints
- 🔧 **Advanced TypeScript** - Complete type definitions
- 🏗️ **Production Patterns** - Retry logic, circuit breakers, deduplication
- ⚡ **Performance** - Caching strategies, virtualization
- 🔄 **Real-time Updates** - Smart polling, WebSocket support
- 🛡️ **Error Recovery** - Graceful degradation, offline support
- 🧪 **Testing** - Mock API, integration tests
- 🚀 **Deployment** - Complete checklist

**Perfect voor:**
- Production deployments
- Performance optimalisatie
- Error handling strategieën
- Testing & monitoring

### 5. **[FRONTEND_OPTIMIZATIONS.md](FRONTEND_OPTIMIZATIONS.md)** - Optimization Summary
Overzicht van alle optimalisaties die zijn doorgevoerd in de backend voor frontend integratie.

## 🚀 Quick Start

### Stap 1: Basis Setup

```typescript
// 1. Installeer dependencies
npm install axios @tanstack/react-query

// 2. Setup API client
import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  timeout: 10000,
});

// 3. Test de connectie
const health = await api.get('/health');
console.log(health.data);
```

### Stap 2: Haal Artikelen Op

```typescript
// Basic article list
const { data } = await api.get('/articles', {
  params: {
    limit: 20,
    source: 'nu.nl',
    sort_by: 'published',
    sort_order: 'desc'
  }
});

console.log(data.data); // Array van articles
console.log(data.meta.pagination); // Pagination info
```

### Stap 3: Implementeer met React Query

```typescript
import { useQuery } from '@tanstack/react-query';

function ArticleList() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['articles'],
    queryFn: async () => {
      const response = await api.get('/articles');
      return response.data;
    }
  });

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <div>
      {data.data.map(article => (
        <ArticleCard key={article.id} article={article} />
      ))}
    </div>
  );
}
```

## 🎨 Voorbeeld Implementaties

### Dashboard met AI Features

```typescript
import { useQuery } from '@tanstack/react-query';

function NewsDashboard() {
  // Articles
  const { data: articles } = useQuery({
    queryKey: ['articles', { limit: 10 }],
    queryFn: () => api.get('/articles', { 
      params: { limit: 10, sort_by: 'published' } 
    })
  });

  // Trending Topics
  const { data: trending } = useQuery({
    queryKey: ['trending'],
    queryFn: () => api.get('/ai/trending', {
      params: { hours: 24, min_articles: 3 }
    }),
    refetchInterval: 60000 // Refresh elke minuut
  });

  // Sentiment Stats
  const { data: sentiment } = useQuery({
    queryKey: ['sentiment'],
    queryFn: () => api.get('/ai/sentiment/stats')
  });

  return (
    <div className="dashboard">
      <SentimentOverview stats={sentiment?.data} />
      <TrendingTopics topics={trending?.data?.topics} />
      <ArticleGrid articles={articles?.data} />
    </div>
  );
}
```

### Article Card met AI Enrichment

```typescript
function ArticleCard({ article }) {
  const { data: ai } = useQuery({
    queryKey: ['enrichment', article.id],
    queryFn: () => api.get(`/articles/${article.id}/enrichment`)
  });

  return (
    <div className="article-card">
      <h3>{article.title}</h3>
      <p>{article.summary}</p>
      
      {/* AI Features */}
      {ai?.data?.sentiment && (
        <SentimentBadge 
          score={ai.data.sentiment.score}
          label={ai.data.sentiment.label}
        />
      )}
      
      {ai?.data?.keywords && (
        <KeywordTags keywords={ai.data.keywords.slice(0, 5)} />
      )}
      
      {ai?.data?.entities?.persons && (
        <EntityChips entities={ai.data.entities.persons} />
      )}
    </div>
  );
}
```

## 📊 Belangrijkste Endpoints

### Basis Endpoints

| Endpoint | Method | Beschrijving | Auth |
|----------|--------|--------------|------|
| `/health` | GET | Comprehensive health check | ❌ |
| `/health/live` | GET | Liveness probe | ❌ |
| `/health/ready` | GET | Readiness probe | ❌ |
| `/health/metrics` | GET | Detailed metrics | ❌ |
| `/articles` | GET | List articles met filters | ❌ |
| `/articles/:id` | GET | Single article | ❌ |
| `/articles/search` | GET | Full-text search | ❌ |
| `/articles/stats` | GET | Statistics | ❌ |
| `/sources` | GET | Available sources | ❌ |
| `/categories` | GET | Available categories | ❌ |

### AI Endpoints

| Endpoint | Method | Beschrijving | Auth |
|----------|--------|--------------|------|
| `/articles/:id/enrichment` | GET | AI enrichment voor artikel | ❌ |
| `/ai/sentiment/stats` | GET | Sentiment statistieken | ❌ |
| `/ai/trending` | GET | Trending topics | ❌ |
| `/ai/entity/:name` | GET | Articles by entity | ❌ |
| `/ai/processor/stats` | GET | Processor status | ❌ |

### Protected Endpoints

| Endpoint | Method | Beschrijving | Auth |
|----------|--------|--------------|------|
| `/scrape` | POST | Trigger scraping | ✅ |
| `/scraper/stats` | GET | Scraper statistics | ✅ |
| `/articles/:id/process` | POST | Process article | ✅ |
| `/ai/process/trigger` | POST | Trigger batch processing | ✅ |

## 🔑 Belangrijke Concepten

### Response Format

Alle endpoints gebruiken een gestandaardiseerd response format:

```typescript
interface APIResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: string;
  };
  meta?: {
    pagination?: PaginationMeta;
    sorting?: SortingMeta;
    filtering?: FilteringMeta;
  };
  request_id: string;
  timestamp: string;
}
```

### Error Codes

| Code | Beschrijving | HTTP Status |
|------|--------------|-------------|
| `INVALID_ID` | Ongeldig artikel ID | 400 |
| `INVALID_DATE` | Ongeldige datum formaat | 400 |
| `MISSING_QUERY` | Verplichte query parameter ontbreekt | 400 |
| `NOT_FOUND` | Resource niet gevonden | 404 |
| `DATABASE_ERROR` | Database fout | 500 |
| `SEARCH_ERROR` | Zoek fout | 500 |
| `PROCESSING_ERROR` | AI processing fout | 500 |

### Rate Limiting

De API heeft rate limiting per IP/API key:
- **Limiet**: 100 requests per 60 seconden (configureerbaar)
- **Headers**: 
  - `X-RateLimit-Limit`: Maximum aantal requests
  - `X-RateLimit-Remaining`: Resterende requests
  - `X-RateLimit-Reset`: Reset tijd in seconden

## 🎓 Learning Path

### Beginner
1. Start met [FRONTEND_API.md](FRONTEND_API.md)
2. Implementeer basis article list
3. Voeg paginatie toe
4. Implementeer search
5. Voeg filters toe (source, category, date)

### Intermediate
1. Lees [FRONTEND_AI_API.md](FRONTEND_AI_API.md)
2. Implementeer sentiment dashboard
3. Voeg trending topics widget toe
4. Implementeer entity-based filtering
5. Voeg AI-enriched article cards toe

### Advanced
1. Lees [FRONTEND_ADVANCED.md](FRONTEND_ADVANCED.md)
2. Implementeer retry logic en circuit breaker
3. Optimaliseer caching strategy
4. Implementeer offline support
5. Setup monitoring en error tracking
6. Write integration tests
7. Deploy naar productie

## 🛠️ Development Tools

### Recommended Stack

**Frontend Framework:**
- React 18+ met TypeScript
- Next.js (voor SSR/SSG)
- Vue 3 met TypeScript
- Angular 15+

**State Management & Data Fetching:**
- TanStack Query (React Query) - **Aanbevolen**
- SWR
- Redux Toolkit met RTK Query
- Zustand + Axios

**UI Components:**
- Tailwind CSS
- shadcn/ui
- Material-UI
- Chakra UI

**Testing:**
- Vitest / Jest
- React Testing Library
- MSW (Mock Service Worker)
- Playwright / Cypress

**Monitoring:**
- Sentry (Error tracking)
- LogRocket (Session replay)
- Google Analytics
- PostHog

## 📈 Performance Tips

1. **Use Pagination** - Nooit alle data tegelijk laden
2. **Implement Virtualization** - Voor lange lijsten (react-window)
3. **Smart Caching** - Verschillende TTLs per data type
4. **Prefetch Next Page** - Betere UX
5. **Debounce Search** - Minder API calls
6. **Optimize Images** - Lazy loading, responsive images
7. **Code Splitting** - Kleinere initial bundle
8. **Monitor Performance** - Gebruik Web Vitals

## 🔒 Security Best Practices

1. **Never expose API keys** in client code
2. **Validate all inputs** voor injection attacks
3. **Use HTTPS only** in productie
4. **Implement CSP headers** tegen XSS
5. **Sanitize user content** voor display
6. **Rate limit client-side** ook
7. **Handle errors safely** zonder sensitive info te lekken
8. **Keep dependencies updated** voor security patches

## 🐛 Debugging Tips

### Request Tracking
Elke response bevat een `request_id`:
```typescript
console.log('Request ID:', data.request_id);
```
Gebruik dit ID voor:
- Log lookup in backend
- Support tickets
- Performance monitoring

### Health Monitoring
Monitor de `/health` endpoint:
```typescript
setInterval(async () => {
  const health = await api.get('/health');
  if (health.data.status !== 'healthy') {
    console.warn('API is degraded:', health.data);
  }
}, 30000);
```

### Rate Limit Tracking
Check rate limit headers:
```typescript
api.interceptors.response.use(response => {
  const remaining = response.headers['x-ratelimit-remaining'];
  const reset = response.headers['x-ratelimit-reset'];
  
  if (remaining < 10) {
    console.warn(`Low rate limit: ${remaining} requests remaining`);
  }
  
  return response;
});
```

## 📞 Support & Resources

### Documentatie
- **API Reference**: [FRONTEND_API.md](FRONTEND_API.md)
- **AI Features**: [FRONTEND_AI_API.md](FRONTEND_AI_API.md)
- **Advanced Patterns**: [FRONTEND_ADVANCED.md](FRONTEND_ADVANCED.md)
- **Backend Docs**: [AI_PROCESSING.md](AI_PROCESSING.md)
- **Quick Start**: [AI_QUICKSTART.md](AI_QUICKSTART.md)

### Backend Setup
- **Start Guide**: [START_BACKEND.md](START_BACKEND.md)
- **Deployment**: [DEPLOYMENT_GUIDE.md](DEPLOYMENT_GUIDE.md)
- **Windows Setup**: [WINDOWS.md](WINDOWS.md)

### Health Checks
- **Health**: `GET http://localhost:8080/health`
- **Liveness**: `GET http://localhost:8080/health/live`
- **Readiness**: `GET http://localhost:8080/health/ready`
- **Metrics**: `GET http://localhost:8080/health/metrics`

## ✅ Pre-Production Checklist

- [ ] Alle API endpoints getest
- [ ] Error handling geïmplementeerd
- [ ] Loading states toegevoegd
- [ ] Rate limiting gerespecteerd
- [ ] Caching geoptimaliseerd
- [ ] Health monitoring setup
- [ ] Error tracking geconfigureerd
- [ ] Performance monitoring actief
- [ ] Security best practices toegepast
- [ ] Offline ondersteuning (indien nodig)
- [ ] Tests geschreven
- [ ] Documentation bijgewerkt
- [ ] Environment variables geconfigureerd
- [ ] CORS instellingen geverifieerd
- [ ] Production build getest

## 🎉 Klaar om te Starten!

1. **Lees** [FRONTEND_API.md](FRONTEND_API.md) voor basis API kennis
2. **Implementeer** een simpele article list
3. **Voeg toe** AI features uit [FRONTEND_AI_API.md](FRONTEND_AI_API.md)
4. **Optimaliseer** met patronen uit [FRONTEND_ADVANCED.md](FRONTEND_ADVANCED.md)
5. **Deploy** naar productie met de checklist

## 🆕 Recent Updates

### v3.1 (2025-10-30) - Database Optimization
- ✅ **Materialized Views Fixed** - All 3 views operational
- ✅ **90% Faster Analytics** - Trending queries: 5s → 0.5s
- ✅ **Sources Metadata** - Automatic tracking implemented
- ✅ **Triggers Optimized** - Duplicate triggers removed
- ✅ **Troubleshooting Guide** - Complete diagnostic toolkit
- ✅ **API Reference Updated** - New Source & Database endpoints

### v2.1 (2025-01-28)
- ✅ Stock Market Integration (FMP API)
- ✅ Email Integration (Outlook IMAP)
- ✅ Complete API documentation
- ✅ AI features fully documented

### v1.0.0 (2025-01-28)
- ✅ Complete API documentatie
- ✅ AI features volledig gedocumenteerd
- ✅ Advanced patterns toegevoegd
- ✅ Health monitoring endpoints
- ✅ TypeScript type definitions
- ✅ React hooks voorbeelden
- ✅ Testing strategieën
- ✅ Deployment guides

---

**Gemaakt met ❤️ voor frontend developers**

Voor vragen of verbeteringen, open een issue of pull request!