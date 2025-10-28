# Stock API Reference - FMP Integration

Complete guide voor alle stock-gerelateerde API endpoints met voorbeelden en use cases.

## 📋 Inhoudsopgave

1. [Quote Endpoints](#quote-endpoints)
2. [Historical Data](#historical-data)
3. [Financial Metrics](#financial-metrics)
4. [News & Calendar](#news--calendar)
5. [Company Search](#company-search)
6. [Batch Operations](#batch-operations)
7. [Frontend Examples](#frontend-examples)

---

## Quote Endpoints

### GET /api/v1/stocks/quote/:symbol

Ophalen van real-time stock quote voor één symbol.

**Parameters:**
- `symbol` (path): Stock symbol (bijv. ASML, AAPL, MSFT)

**Response:**
```json
{
  "symbol": "ASML",
  "name": "ASML Holding NV",
  "price": 745.30,
  "change": 12.50,
  "change_percent": 1.71,
  "volume": 1250000,
  "market_cap": 295000000000,
  "exchange": "NASDAQ",
  "currency": "USD",
  "last_updated": "2024-01-15T14:30:00Z",
  "previous_close": 732.80,
  "open": 735.00,
  "day_high": 748.20,
  "day_low": 733.50,
  "year_high": 850.00,
  "year_low": 550.00,
  "price_avg_50": 720.45,
  "price_avg_200": 680.30,
  "eps": 28.50,
  "pe": 26.15,
  "shares_outstanding": 395000000
}
```

**Frontend Example:**
```typescript
async function getStockQuote(symbol: string) {
  const response = await fetch(`/api/v1/stocks/quote/${symbol}`);
  const quote = await response.json();
  return quote;
}
```

---

### POST /api/v1/stocks/quotes

**⚡ BATCH ENDPOINT** - Ophalen van meerdere quotes in één API call.

**Maximum:** 100 symbols per request
**Cost:** 1 API call (was N calls voor N symbols)
**Savings:** 90-99% afhankelijk van aantal symbols

**Request Body:**
```json
{
  "symbols": ["ASML", "SHELL", "ING", "AAPL", "MSFT", "GOOGL"]
}
```

**Response:**
```json
{
  "quotes": {
    "ASML": {
      "symbol": "ASML",
      "price": 745.30,
      "change": 12.50,
      "change_percent": 1.71
    },
    "SHELL": {
      "symbol": "SHELL",
      "price": 28.45,
      "change": -0.15,
      "change_percent": -0.52
    }
  },
  "meta": {
    "total": 6,
    "requested": 6,
    "duration_ms": 245,
    "using_batch": true,
    "cost_saving": "83%"
  }
}
```

**Frontend Example:**
```typescript
async function getBatchQuotes(symbols: string[]) {
  const response = await fetch('/api/v1/stocks/quotes', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ symbols })
  });
  const data = await response.json();
  return data.quotes;
}

// Usage
const quotes = await getBatchQuotes(['ASML', 'SHELL', 'ING']);
```

---

## Historical Data

### GET /api/v1/stocks/historical/:symbol

Ophalen van historische prijsdata (OHLC - Open, High, Low, Close).

**Parameters:**
- `symbol` (path): Stock symbol
- `from` (query, optional): Start date (format: YYYY-MM-DD, default: 30 dagen geleden)
- `to` (query, optional): End date (format: YYYY-MM-DD, default: vandaag)

**Example Request:**
```
GET /api/v1/stocks/historical/ASML?from=2024-01-01&to=2024-01-31
```

**Response:**
```json
{
  "symbol": "ASML",
  "from": "2024-01-01",
  "to": "2024-01-31",
  "dataPoints": 21,
  "prices": [
    {
      "date": "2024-01-31T00:00:00Z",
      "open": 740.00,
      "high": 748.50,
      "low": 738.20,
      "close": 745.30,
      "adjClose": 745.30,
      "volume": 1250000,
      "change": 5.30,
      "changePercent": 0.72
    },
    {
      "date": "2024-01-30T00:00:00Z",
      "open": 735.00,
      "high": 742.00,
      "low": 733.50,
      "close": 740.00,
      "adjClose": 740.00,
      "volume": 980000,
      "change": 5.00,
      "changePercent": 0.68
    }
  ]
}
```

**Frontend Example - Chart.js Integration:**
```typescript
async function getHistoricalChart(symbol: string, days: number = 30) {
  const to = new Date();
  const from = new Date(to.getTime() - days * 24 * 60 * 60 * 1000);
  
  const response = await fetch(
    `/api/v1/stocks/historical/${symbol}?` +
    `from=${from.toISOString().split('T')[0]}&` +
    `to=${to.toISOString().split('T')[0]}`
  );
  
  const data = await response.json();
  
  return {
    labels: data.prices.map(p => new Date(p.date).toLocaleDateString()),
    datasets: [{
      label: symbol,
      data: data.prices.map(p => p.close),
      borderColor: 'rgb(75, 192, 192)',
      tension: 0.1
    }]
  };
}
```

---

## Financial Metrics

### GET /api/v1/stocks/metrics/:symbol

Ophalen van key financial metrics (P/E ratio, ROE, debt/equity, etc.).

**Parameters:**
- `symbol` (path): Stock symbol

**Response:**
```json
{
  "symbol": "ASML",
  "marketCap": 295000000000,
  "peRatio": 26.15,
  "pegRatio": 1.85,
  "priceToBook": 12.40,
  "priceToSales": 8.90,
  "roe": 0.48,
  "roa": 0.22,
  "debtToEquity": 0.35,
  "currentRatio": 2.15,
  "dividendYield": 0.012,
  "eps": 28.50,
  "revenuePerShare": 83.75,
  "freeCashFlowYield": 0.045
}
```

**Frontend Example - Metrics Dashboard:**
```typescript
interface FinancialMetrics {
  symbol: string;
  peRatio: number;
  roe: number;
  debtToEquity: number;
  dividendYield: number;
}

async function getFinancialMetrics(symbol: string): Promise<FinancialMetrics> {
  const response = await fetch(`/api/v1/stocks/metrics/${symbol}`);
  return await response.json();
}

// Display in dashboard
function MetricsCard({ symbol }: { symbol: string }) {
  const [metrics, setMetrics] = useState<FinancialMetrics>();
  
  useEffect(() => {
    getFinancialMetrics(symbol).then(setMetrics);
  }, [symbol]);
  
  return (
    <div className="metrics-grid">
      <div>P/E Ratio: {metrics?.peRatio.toFixed(2)}</div>
      <div>ROE: {(metrics?.roe * 100).toFixed(2)}%</div>
      <div>Debt/Equity: {metrics?.debtToEquity.toFixed(2)}</div>
      <div>Dividend Yield: {(metrics?.dividendYield * 100).toFixed(2)}%</div>
    </div>
  );
}
```

---

## News & Calendar

### GET /api/v1/stocks/news/:symbol

Ophalen van FMP stock news voor een specifiek symbol.

**Parameters:**
- `symbol` (path): Stock symbol
- `limit` (query, optional): Aantal artikelen (max 50, default 10)

**Example Request:**
```
GET /api/v1/stocks/news/AAPL?limit=5
```

**Response:**
```json
{
  "symbol": "AAPL",
  "total": 5,
  "news": [
    {
      "symbol": "AAPL",
      "publishedDate": "2024-01-15T14:30:00Z",
      "title": "Apple announces new AI features",
      "image": "https://...",
      "site": "Reuters",
      "text": "Apple Inc. announced today...",
      "url": "https://..."
    }
  ]
}
```

**Frontend Example:**
```typescript
async function getStockNews(symbol: string, limit: number = 10) {
  const response = await fetch(`/api/v1/stocks/news/${symbol}?limit=${limit}`);
  const data = await response.json();
  return data.news;
}

// Combine with your own articles
async function getAllNews(symbol: string) {
  const [fmpNews, localArticles] = await Promise.all([
    getStockNews(symbol, 5),
    fetch(`/api/v1/articles/by-ticker/${symbol}`).then(r => r.json())
  ]);
  
  return {
    international: fmpNews,
    local: localArticles
  };
}
```

---

### GET /api/v1/stocks/earnings

Ophalen van earnings calendar (aankomende earnings announcements).

**Parameters:**
- `from` (query, optional): Start date (YYYY-MM-DD, default: vandaag)
- `to` (query, optional): End date (YYYY-MM-DD, default: +7 dagen)

**Example Request:**
```
GET /api/v1/stocks/earnings?from=2024-01-15&to=2024-01-22
```

**Response:**
```json
{
  "from": "2024-01-15",
  "to": "2024-01-22",
  "total": 12,
  "earnings": [
    {
      "symbol": "AAPL",
      "date": "2024-01-18T00:00:00Z",
      "eps": 2.18,
      "epsEstimated": 2.10,
      "time": "amc",
      "revenue": 119575000000,
      "revenueEstimated": 118000000000
    }
  ]
}
```

**Frontend Example - Calendar Widget:**
```typescript
async function getEarningsCalendar(daysAhead: number = 7) {
  const from = new Date().toISOString().split('T')[0];
  const to = new Date(Date.now() + daysAhead * 24*60*60*1000)
    .toISOString().split('T')[0];
  
  const response = await fetch(
    `/api/v1/stocks/earnings?from=${from}&to=${to}`
  );
  const data = await response.json();
  return data.earnings;
}

// Display calendar
function EarningsCalendar() {
  const [earnings, setEarnings] = useState([]);
  
  useEffect(() => {
    getEarningsCalendar(14).then(setEarnings);
  }, []);
  
  return (
    <div className="earnings-calendar">
      {earnings.map(e => (
        <div key={e.symbol} className="earning-item">
          <span>{e.symbol}</span>
          <span>{new Date(e.date).toLocaleDateString()}</span>
          <span>EPS: ${e.eps} (est. ${e.epsEstimated})</span>
          <span className={e.eps > e.epsEstimated ? 'beat' : 'miss'}>
            {e.eps > e.epsEstimated ? '✅ Beat' : '⚠️ Miss'}
          </span>
        </div>
      ))}
    </div>
  );
}
```

---

## Company Search

### GET /api/v1/stocks/search

Zoeken naar bedrijven en stock symbols.

**Parameters:**
- `q` (query, required): Search query (bedrijfsnaam of symbol)
- `limit` (query, optional): Aantal resultaten (max 50, default 10)

**Example Request:**
```
GET /api/v1/stocks/search?q=apple&limit=5
```

**Response:**
```json
{
  "query": "apple",
  "total": 2,
  "results": [
    {
      "symbol": "AAPL",
      "company_name": "Apple Inc.",
      "currency": "USD",
      "exchange": "NASDAQ"
    },
    {
      "symbol": "AAPL.MX",
      "company_name": "Apple Inc.",
      "currency": "MXN",
      "exchange": "MEX"
    }
  ]
}
```

**Frontend Example - Autocomplete:**
```typescript
async function searchSymbols(query: string) {
  if (query.length < 2) return [];
  
  const response = await fetch(
    `/api/v1/stocks/search?q=${encodeURIComponent(query)}&limit=10`
  );
  const data = await response.json();
  return data.results;
}

// React autocomplete component
function SymbolSearch() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [loading, setLoading] = useState(false);

  const handleSearch = useCallback(
    debounce(async (q: string) => {
      setLoading(true);
      const results = await searchSymbols(q);
      setResults(results);
      setLoading(false);
    }, 300),
    []
  );

  return (
    <div>
      <input 
        type="text"
        value={query}
        onChange={(e) => {
          setQuery(e.target.value);
          handleSearch(e.target.value);
        }}
        placeholder="Search stocks..."
      />
      {loading && <Spinner />}
      <ul>
        {results.map(r => (
          <li key={r.symbol} onClick={() => selectSymbol(r.symbol)}>
            <strong>{r.symbol}</strong> - {r.company_name}
            <span className="exchange">{r.exchange}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}
```

---

## Batch Operations

### Best Practices voor Batch Requests

**❌ SLECHT - Individuele calls:**
```typescript
// Dit kost 10 API calls!
async function getMultipleQuotes(symbols: string[]) {
  const quotes = {};
  for (const symbol of symbols) {
    const quote = await fetch(`/api/v1/stocks/quote/${symbol}`);
    quotes[symbol] = await quote.json();
  }
  return quotes;
}
```

**✅ GOED - Batch endpoint:**
```typescript
// Dit kost 1 API call!
async function getMultipleQuotes(symbols: string[]) {
  const response = await fetch('/api/v1/stocks/quotes', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ symbols })
  });
  const data = await response.json();
  return data.quotes;
}
```

**Performance Comparison:**
```typescript
// Test met 20 symbols
console.time('Individual calls');
for (let i = 0; i < 20; i++) {
  await getQuote(symbols[i]); // 3.6 seconds total
}
console.timeEnd('Individual calls');

console.time('Batch call');
await getBatchQuotes(symbols); // 0.25 seconds total
console.timeEnd('Batch call');

// Result: 93% sneller + 95% goedkoper!
```

---

## Frontend Examples

### Complete Stock Widget

```typescript
import React, { useEffect, useState } from 'react';

interface StockWidgetProps {
  symbol: string;
}

export function StockWidget({ symbol }: StockWidgetProps) {
  const [quote, setQuote] = useState(null);
  const [metrics, setMetrics] = useState(null);
  const [news, setNews] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function loadStockData() {
      setLoading(true);
      
      // Parallel fetch (efficient)
      const [quoteData, metricsData, newsData] = await Promise.all([
        fetch(`/api/v1/stocks/quote/${symbol}`).then(r => r.json()),
        fetch(`/api/v1/stocks/metrics/${symbol}`).then(r => r.json()),
        fetch(`/api/v1/stocks/news/${symbol}?limit=5`).then(r => r.json())
      ]);
      
      setQuote(quoteData);
      setMetrics(metricsData);
      setNews(newsData.news);
      setLoading(false);
    }
    
    loadStockData();
    
    // Auto-refresh every 5 minutes
    const interval = setInterval(loadStockData, 5 * 60 * 1000);
    return () => clearInterval(interval);
  }, [symbol]);

  if (loading) return <Skeleton />;

  return (
    <div className="stock-widget">
      {/* Price Card */}
      <div className="price-card">
        <h2>{quote.symbol}</h2>
        <div className="price">${quote.price.toFixed(2)}</div>
        <div className={quote.change >= 0 ? 'positive' : 'negative'}>
          {quote.change >= 0 ? '+' : ''}{quote.change.toFixed(2)} 
          ({quote.change_percent.toFixed(2)}%)
        </div>
        <div className="volume">Vol: {(quote.volume / 1000000).toFixed(2)}M</div>
      </div>

      {/* Metrics Grid */}
      <div className="metrics-grid">
        <MetricItem label="P/E" value={metrics.peRatio.toFixed(2)} />
        <MetricItem label="Market Cap" value={formatMarketCap(metrics.marketCap)} />
        <MetricItem label="ROE" value={`${(metrics.roe * 100).toFixed(2)}%`} />
        <MetricItem label="Dividend" value={`${(metrics.dividendYield * 100).toFixed(2)}%`} />
      </div>

      {/* Recent News */}
      <div className="news-section">
        <h3>Recent News</h3>
        {news.map(item => (
          <a key={item.url} href={item.url} target="_blank" className="news-item">
            <div className="news-title">{item.title}</div>
            <div className="news-meta">
              {item.site} • {new Date(item.publishedDate).toLocaleDateString()}
            </div>
          </a>
        ))}
      </div>
    </div>
  );
}
```

### Portfolio Tracker

```typescript
interface Portfolio {
  symbols: string[];
}

export function PortfolioTracker({ symbols }: Portfolio) {
  const [quotes, setQuotes] = useState({});
  const [totalValue, setTotalValue] = useState(0);
  const [totalChange, setTotalChange] = useState(0);

  useEffect(() => {
    async function loadPortfolio() {
      // Single batch call voor alle symbols!
      const response = await fetch('/api/v1/stocks/quotes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ symbols })
      });
      
      const data = await response.json();
      setQuotes(data.quotes);
      
      // Calculate totals
      const value = Object.values(data.quotes).reduce(
        (sum, q) => sum + q.market_cap, 0
      );
      const change = Object.values(data.quotes).reduce(
        (sum, q) => sum + q.change_percent, 0
      ) / symbols.length;
      
      setTotalValue(value);
      setTotalChange(change);
    }
    
    loadPortfolio();
    
    // Refresh every 1 minute (cached, zo geen extra API calls)
    const interval = setInterval(loadPortfolio, 60 * 1000);
    return () => clearInterval(interval);
  }, [symbols]);

  return (
    <div className="portfolio">
      <div className="portfolio-header">
        <h2>Portfolio Overview</h2>
        <div className="total-change" className={totalChange >= 0 ? 'up' : 'down'}>
          {totalChange >= 0 ? '▲' : '▼'} {Math.abs(totalChange).toFixed(2)}%
        </div>
      </div>
      
      <div className="holdings">
        {symbols.map(symbol => {
          const quote = quotes[symbol];
          if (!quote) return null;
          
          return (
            <div key={symbol} className="holding">
              <div className="symbol">{symbol}</div>
              <div className="price">${quote.price.toFixed(2)}</div>
              <div className={quote.change >= 0 ? 'positive' : 'negative'}>
                {quote.change >= 0 ? '+' : ''}{quote.change_percent.toFixed(2)}%
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
```

### Articles met Stock Context

```typescript
async function getArticleWithStockContext(articleId: number) {
  // Get article with AI enrichment
  const article = await fetch(`/api/v1/articles/${articleId}`).then(r => r.json());
  
  // Extract stock tickers from AI enrichment
  const tickers = article.ai_enrichment?.entities?.stock_tickers || [];
  
  if (tickers.length === 0) {
    return { article, stocks: null };
  }
  
  // Fetch real-time quotes for all tickers in article (1 batch call!)
  const symbols = tickers.map(t => t.symbol);
  const quotes = await fetch('/api/v1/stocks/quotes', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ symbols })
  }).then(r => r.json());
  
  return {
    article,
    stocks: quotes.quotes,
    tickers: tickers
  };
}

// Display in article page
function ArticleWithStocks({ articleId }: { articleId: number }) {
  const [data, setData] = useState(null);
  
  useEffect(() => {
    getArticleWithStockContext(articleId).then(setData);
  }, [articleId]);
  
  if (!data) return <Loader />;
  
  return (
    <article>
      <h1>{data.article.title}</h1>
      <p>{data.article.summary}</p>
      
      {data.stocks && (
        <div className="mentioned-stocks">
          <h3>📈 Mentioned Stocks</h3>
          {data.tickers.map(ticker => {
            const quote = data.stocks[ticker.symbol];
            return (
              <div key={ticker.symbol} className="stock-mention">
                <strong>{ticker.symbol}</strong> - {ticker.name}
                {quote && (
                  <span className={quote.change >= 0 ? 'up' : 'down'}>
                    ${quote.price.toFixed(2)} 
                    ({quote.change >= 0 ? '+' : ''}{quote.change_percent.toFixed(2)}%)
                  </span>
                )}
                <div className="context">{ticker.context}</div>
              </div>
            );
          })}
        </div>
      )}
    </article>
  );
}
```

---

## ⚡ Performance Tips

### 1. Gebruik Batch Endpoints

```typescript
// ❌ SLECHT
const quotes = await Promise.all(
  symbols.map(s => fetch(`/api/v1/stocks/quote/${s}`))
);

// ✅ GOED
const { quotes } = await fetch('/api/v1/stocks/quotes', {
  method: 'POST',
  body: JSON.stringify({ symbols })
}).then(r => r.json());
```

### 2. Leverage Caching

```typescript
// Data is cached for 5 minutes
// Multiple calls binnen 5 min = geen extra API costs
async function getQuoteWithRetry(symbol: string) {
  const response = await fetch(`/api/v1/stocks/quote/${symbol}`);
  if (!response.ok) {
    // Retry is gratis dankzij cache
    await new Promise(r => setTimeout(r, 1000));
    return fetch(`/api/v1/stocks/quote/${symbol}`);
  }
  return response;
}
```

### 3. Parallel Fetching

```typescript
// Fetch verschillende data types parallel
async function getCompleteStockData(symbol: string) {
  const [quote, profile, metrics, news] = await Promise.all([
    fetch(`/api/v1/stocks/quote/${symbol}`).then(r => r.json()),
    fetch(`/api/v1/stocks/profile/${symbol}`).then(r => r.json()),
    fetch(`/api/v1/stocks/metrics/${symbol}`).then(r => r.json()),
    fetch(`/api/v1/stocks/news/${symbol}?limit=5`).then(r => r.json())
  ]);
  
  return { quote, profile, metrics, news: news.news };
}
```

---

## 💰 Cost Optimization Strategies

### Strategy 1: Smart Batching
```typescript
// Collect symbols from multiple articles
function extractAllSymbols(articles: Article[]): string[] {
  const symbols = new Set<string>();
  
  articles.forEach(article => {
    article.ai_enrichment?.entities?.stock_tickers?.forEach(ticker => {
      symbols.add(ticker.symbol);
    });
  });
  
  return Array.from(symbols);
}

// Single batch call
const symbols = extractAllSymbols(articles);
const quotes = await getBatchQuotes(symbols);
```

### Strategy 2: Cache-aware Updates
```typescript
// Only fetch if cache might be stale
async function getQuoteIfNeeded(symbol: string, lastUpdate: Date) {
  const cacheAge = Date.now() - lastUpdate.getTime();
  const CACHE_TTL = 5 * 60 * 1000; // 5 minutes
  
  if (cacheAge < CACHE_TTL) {
    // Use cached data from article.stock_data
    return article.stock_data[symbol];
  }
  
  // Fetch fresh data
  return await getStockQuote(symbol);
}
```

### Strategy 3: Bulk Pre-loading
```typescript
// Preload popular stocks on app init
async function warmCache() {
  const popularSymbols = ['ASML', 'SHELL', 'ING', 'AAPL', 'MSFT', 'GOOGL'];
  
  // Single batch call warms cache voor 5 minutes
  await getBatchQuotes(popularSymbols);
  
  console.log('✅ Cache warmed for', popularSymbols.length, 'symbols');
}

// Call on app mount
useEffect(() => {
  warmCache();
}, []);
```

---

## 🔍 Error Handling

```typescript
async function getStockData(symbol: string) {
  try {
    const response = await fetch(`/api/v1/stocks/quote/${symbol}`);
    
    if (!response.ok) {
      if (response.status === 404) {
        throw new Error(`Symbol '${symbol}' not found`);
      }
      if (response.status === 429) {
        throw new Error('Rate limit exceeded. Please try again later.');
      }
      throw new Error('Failed to fetch stock data');
    }
    
    return await response.json();
  } catch (error) {
    console.error('Stock data error:', error);
    
    // Fallback to cached data if available
    if (cachedData) {
      console.warn('Using cached data due to API error');
      return cachedData;
    }
    
    throw error;
  }
}
```

---

## 📊 Rate Limiting

**Current Limits:**
- **Rate**: 30 calls/minute (configurable)
- **Batch size**: 100 symbols max
- **Cache TTL**: 5 minutes (quotes), 1 hour (metrics), 24 hours (profiles)

**Headers:**
```
X-RateLimit-Limit: 30
X-RateLimit-Remaining: 25
X-RateLimit-Reset: 1705330800
```

**Frontend Handling:**
```typescript
async function fetchWithRateLimit(url: string, options?: RequestInit) {
  const response = await fetch(url, options);
  
  const limit = response.headers.get('X-RateLimit-Limit');
  const remaining = response.headers.get('X-RateLimit-Remaining');
  const reset = response.headers.get('X-RateLimit-Reset');
  
  if (remaining && parseInt(remaining) < 5) {
    console.warn(`Low rate limit: ${remaining}/${limit} remaining`);
    // Implement backoff or queue
  }
  
  if (response.status === 429) {
    const retryAfter = parseInt(reset) - Math.floor(Date.now() / 1000);
    throw new Error(`Rate limited. Retry after ${retryAfter}s`);
  }
  
  return response;
}
```

---

## 🎯 Use Cases

### Use Case 1: Real-time Stock Dashboard
```typescript
function StockDashboard() {
  const watchlist = ['ASML', 'SHELL', 'ING', 'AAPL', 'MSFT'];
  const [quotes, setQuotes] = useState({});
  
  useEffect(() => {
    const updateQuotes = async () => {
      const data = await getBatchQuotes(watchlist);
      setQuotes(data);
    };
    
    updateQuotes();
    const interval = setInterval(updateQuotes, 60000); // Every minute
    return () => clearInterval(interval);
  }, []);
  
  return <QuotesGrid quotes={quotes} />;
}
```

### Use Case 2: Article Enrichment
```typescript
async function enrichArticles(articles: Article[]) {
  // Extract all unique stock symbols
  const symbols = extractAllSymbols(articles);
  
  if (symbols.length === 0) return articles;
  
  // Fetch all quotes in one batch call
  const quotes = await getBatchQuotes(symbols);
  
  // Enrich articles with stock data
  return articles.map(article => ({
    ...article,
    stockContext: article.ai_enrichment?.entities?.stock_tickers?.map(ticker => ({
      ...ticker,
      currentPrice: quotes[ticker.symbol]?.price,
      change: quotes[ticker.symbol]?.change_percent
    }))
  }));
}
```

### Use Case 3: Earnings Alerts
```typescript
async function getUpcomingEarnings() {
  const from = new Date().toISOString().split('T')[0];
  const to = new Date(Date.now() + 7*24*60*60*1000).toISOString().split('T')[0];
  
  const response = await fetch(`/api/v1/stocks/earnings?from=${from}&to=${to}`);
  const data = await response.json();
  
  // Filter for high-impact earnings
  return data.earnings.filter(e => 
    Math.abs(e.eps - e.epsEstimated) / e.epsEstimated > 0.05 // 5% surprise
  );
}
```

---

## 📖 Related Documentation

- [Cost Optimization Report](./cost-optimization-report.md)
- [Stock Tickers Feature](./stock-tickers.md)
- [Frontend Integration Guide](../frontend/stock-tickers-integration.md)
- [API Reference](../api/README.md)

---

**Last Updated:** 2024-01-15
**API Version:** v1.0
**FMP API Version:** v3