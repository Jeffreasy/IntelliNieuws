# 🚀 FMP Integration - Quick Start Guide

## Binnen 5 Minuten aan de Slag

Deze guide helpt je binnen 5 minuten alle nieuwe FMP features te gebruiken in je IntelliNieuws applicatie.

---

## ⚙️ Setup (2 minuten)

### Stap 1: FMP API Key Verkrijgen

1. Ga naar https://site.financialmodelingprep.com/developer/docs/
2. Maak gratis account aan
3. Kopieer je API key

### Stap 2: Configuratie

Voeg toe aan je `.env`:

```bash
# Stock API Configuration
STOCK_API_PROVIDER=fmp
STOCK_API_KEY=your_fmp_api_key_here
STOCK_API_CACHE_TTL_MINUTES=5
STOCK_API_RATE_LIMIT_PER_MINUTE=30
STOCK_API_ENABLE_CACHE=true
```

### Stap 3: Start Backend

```bash
# Rebuild met nieuwe features
go build -o api.exe ./cmd/api

# Start
./api.exe
```

Klaar! Backend draait op `http://localhost:8080` ✅

---

## 🧪 Test Features (3 minuten)

### Test 1: Batch Quotes (Cost Optimization)

```bash
curl -X POST http://localhost:8080/api/v1/stocks/quotes \
  -H "Content-Type: application/json" \
  -d '{"symbols": ["ASML", "SHELL", "ING", "AAPL", "MSFT"]}'
```

**Verwacht resultaat:**
```json
{
  "quotes": {
    "ASML": { "price": 745.30, "change_percent": 1.71 },
    "SHELL": { "price": 28.45, "change_percent": -0.52 }
  },
  "meta": {
    "total": 5,
    "using_batch": true,
    "cost_saving": "80%"  // 🎉
  }
}
```

✅ **Success:** Je ziet `"using_batch": true` en `cost_saving`!

---

### Test 2: Market Overview

```bash
# Top gainers
curl http://localhost:8080/api/v1/stocks/market/gainers

# Top losers  
curl http://localhost:8080/api/v1/stocks/market/losers

# Sector performance
curl http://localhost:8080/api/v1/stocks/sectors
```

**Verwacht:** Top 10 gaining/losing stocks en sector percentages

---

### Test 3: Historical Chart Data

```bash
curl "http://localhost:8080/api/v1/stocks/historical/AAPL?from=2024-01-01&to=2024-01-31"
```

**Verwacht:** Array met OHLC data voor elke dag

---

### Test 4: Auto-Enrichment

```bash
# 1. Check artikel met stock tickers
curl http://localhost:8080/api/v1/articles/1

# 2. Check of ai_stock_tickers is gevuld
# 3. Check of stock_data automatic is toegevoegd
```

**Verwacht:** Artikel heeft `stock_data` met real-time quotes!

---

## 💻 Frontend Integration (Direct te gebruiken!)

### Optie 1: Kopieer-en-Plak Widgets

**Market Overview Widget:**

```typescript
// components/MarketOverview.tsx
import { useEffect, useState } from 'react';

export function MarketOverview() {
  const [gainers, setGainers] = useState([]);
  const [losers, setLosers] = useState([]);

  useEffect(() => {
    // Fetch market data
    Promise.all([
      fetch('http://localhost:8080/api/v1/stocks/market/gainers').then(r => r.json()),
      fetch('http://localhost:8080/api/v1/stocks/market/losers').then(r => r.json())
    ]).then(([gainersData, losersData]) => {
      setGainers(gainersData.gainers || []);
      setLosers(losersData.losers || []);
    });
  }, []);

  return (
    <div style={{ display: 'flex', gap: '20px' }}>
      <div>
        <h3>📈 Top Gainers</h3>
        {gainers.slice(0, 5).map(stock => (
          <div key={stock.symbol} style={{ color: 'green' }}>
            {stock.symbol}: +{stock.changePercent.toFixed(2)}%
          </div>
        ))}
      </div>
      <div>
        <h3>📉 Top Losers</h3>
        {losers.slice(0, 5).map(stock => (
          <div key={stock.symbol} style={{ color: 'red' }}>
            {stock.symbol}: {stock.changePercent.toFixed(2)}%
          </div>
        ))}
      </div>
    </div>
  );
}
```

**Stock Price Card:**

```typescript
// components/StockCard.tsx
import { useEffect, useState } from 'react';

export function StockCard({ symbol }: { symbol: string }) {
  const [quote, setQuote] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch(`http://localhost:8080/api/v1/stocks/quote/${symbol}`)
      .then(r => r.json())
      .then(data => {
        setQuote(data);
        setLoading(false);
      });
  }, [symbol]);

  if (loading) return <div>Loading...</div>;
  if (!quote) return null;

  return (
    <div className="stock-card">
      <h3>{quote.symbol}</h3>
      <div className="price">${quote.price.toFixed(2)}</div>
      <div className={quote.change >= 0 ? 'positive' : 'negative'}>
        {quote.change >= 0 ? '▲' : '▼'} 
        {Math.abs(quote.changePercent).toFixed(2)}%
      </div>
      <div className="details">
        <span>Vol: {(quote.volume / 1000000).toFixed(2)}M</span>
        <span>Cap: ${(quote.market_cap / 1e9).toFixed(2)}B</span>
      </div>
    </div>
  );
}
```

**Earnings Calendar:**

```typescript
// components/EarningsCalendar.tsx
import { useEffect, useState } from 'react';

export function EarningsCalendar() {
  const [earnings, setEarnings] = useState([]);

  useEffect(() => {
    const from = new Date().toISOString().split('T')[0];
    const to = new Date(Date.now() + 7*24*60*60*1000).toISOString().split('T')[0];
    
    fetch(`http://localhost:8080/api/v1/stocks/earnings?from=${from}&to=${to}`)
      .then(r => r.json())
      .then(data => setEarnings(data.earnings || []));
  }, []);

  return (
    <div>
      <h3>📅 Upcoming Earnings (Next 7 Days)</h3>
      {earnings.map(e => (
        <div key={`${e.symbol}-${e.date}`} style={{ padding: '10px', borderBottom: '1px solid #eee' }}>
          <strong>{e.symbol}</strong> - {new Date(e.date).toLocaleDateString()}
          <div style={{ fontSize: '0.9em', color: '#666' }}>
            EPS Est: ${e.epsEstimated} | Rev Est: ${(e.revenueEstimated / 1e9).toFixed(2)}B
          </div>
        </div>
      ))}
    </div>
  );
}
```

---

## 🎯 Common Use Cases

### Use Case 1: Real-time Stock Dashboard

**Goal:** Toon live koersen voor watchlist

**Code:**
```typescript
const watchlist = ['ASML', 'SHELL', 'ING', 'AAPL', 'MSFT'];

// Single batch call!
const response = await fetch('http://localhost:8080/api/v1/stocks/quotes', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ symbols: watchlist })
});

const data = await response.json();
console.log('Fetched', Object.keys(data.quotes).length, 'quotes');
console.log('Cost saving:', data.meta.cost_saving); // "80%"
```

**Result:** 1 API call in plaats van 5 → **80% goedkoper**

---

### Use Case 2: Article Enrichment

**Goal:** Voeg stock context toe aan artikelen

**Code:**
```typescript
async function enrichArticleWithStocks(article) {
  // Extract tickers from AI enrichment
  const tickers = article.ai_enrichment?.entities?.stock_tickers || [];
  
  if (tickers.length === 0) return article;
  
  // Batch fetch quotes
  const symbols = tickers.map(t => t.symbol);
  const response = await fetch('http://localhost:8080/api/v1/stocks/quotes', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ symbols })
  });
  
  const { quotes } = await response.json();
  
  // Merge data
  return {
    ...article,
    stockContext: tickers.map(ticker => ({
      ...ticker,
      currentPrice: quotes[ticker.symbol]?.price,
      change: quotes[ticker.symbol]?.change_percent,
      marketCap: quotes[ticker.symbol]?.market_cap
    }))
  };
}
```

---

### Use Case 3: Market Trending Stocks

**Goal:** Toon trending stocks in sidebar

**Code:**
```typescript
function TrendingStocks() {
  const [trending, setTrending] = useState([]);

  useEffect(() => {
    // Mix gainers and most actives
    Promise.all([
      fetch('http://localhost:8080/api/v1/stocks/market/gainers').then(r => r.json()),
      fetch('http://localhost:8080/api/v1/stocks/market/actives').then(r => r.json())
    ]).then(([gainers, actives]) => {
      // Combine and deduplicate
      const combined = [
        ...gainers.gainers.slice(0, 3),
        ...actives.actives.slice(0, 3)
      ];
      
      const unique = Array.from(
        new Map(combined.map(s => [s.symbol, s])).values()
      );
      
      setTrending(unique);
    });
  }, []);

  return (
    <div className="trending">
      <h4>🔥 Trending Now</h4>
      {trending.map(stock => (
        <div key={stock.symbol}>
          {stock.symbol}: {stock.changePercent.toFixed(2)}%
        </div>
      ))}
    </div>
  );
}
```

---

## 📊 Testing Scenarios

### Scenario 1: Cost Optimization Check

**Test:**
```bash
# Request 10 symbols
curl -X POST http://localhost:8080/api/v1/stocks/quotes \
  -H "Content-Type: application/json" \
  -d '{"symbols": ["ASML","SHELL","ING","AAPL","MSFT","GOOGL","AMZN","META","TSLA","NVDA"]}'
```

**Check in logs:**
```
INFO: ✅ Fetched 10 quotes in single batch API call (cost: 1 call, saved: 9 calls)
INFO: Batch quotes: 10 symbols fetched in 245ms (40.82 symbols/sec)
```

✅ **Expected:** 1 API call voor 10 symbols = 90% cost saving

---

### Scenario 2: Cache Performance

**Test:**
```bash
# First call
time curl http://localhost:8080/api/v1/stocks/quote/ASML

# Second call (within 5 min)
time curl http://localhost:8080/api/v1/stocks/quote/ASML
```

**Check in logs:**
```
# First call
INFO: Fetching quote for ASML from FMP API

# Second call  
DEBUG: Cache HIT for stock quote: ASML
```

✅ **Expected:** Second call is instant (<10ms)

---

### Scenario 3: Auto-Enrichment

**Test:**
```bash
# 1. Scrape artikel (creëert artikel met tickers via AI)
curl -X POST http://localhost:8080/api/v1/scrape \
  -H "X-API-Key: your-key"

# 2. Check logs voor auto-enrichment
tail -f logs/app.log | grep "Auto-enriching"
```

**Expected in logs:**
```
INFO: AI processor started
INFO: Found 5 pending articles, processing with worker pool
INFO: Successfully processed article 123
INFO: 🔄 Auto-enriching 5 articles with stock data...
INFO: 🚀 Fetching stock data for 8 unique symbols across 5 articles using BATCH API
INFO: ✅ Fetched 8 quotes in single batch API call (saved: 7 calls)
INFO: ✅ Enriched 5 articles with stock data (1 batch API call for 8 symbols)
```

✅ **Expected:** Automatic enrichment na AI processing!

---

## 💡 Pro Tips

### Tip 1: Warm Cache on Startup

```typescript
// app/lib/cacheWarmer.ts
export async function warmStockCache() {
  const popularSymbols = [
    'ASML', 'SHELL', 'ING',           // Dutch
    'AAPL', 'MSFT', 'GOOGL', 'NVDA'  // US Tech
  ];

  await fetch('http://localhost:8080/api/v1/stocks/quotes', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ symbols: popularSymbols })
  });

  console.log('✅ Cache warmed for', popularSymbols.length, 'popular stocks');
}

// Call in _app.tsx or layout.tsx
useEffect(() => {
  warmStockCache();
}, []);
```

**Benefit:** Eerste page load is instant voor deze symbols!

---

### Tip 2: Efficient Article Loading

```typescript
// Fetch articles with stock context
async function loadArticlesWithStocks() {
  // 1. Get articles
  const articles = await fetch('http://localhost:8080/api/v1/articles')
    .then(r => r.json());

  // 2. Extract ALL stock symbols from ALL articles
  const allSymbols = new Set();
  articles.forEach(article => {
    article.ai_enrichment?.entities?.stock_tickers?.forEach(ticker => {
      allSymbols.add(ticker.symbol);
    });
  });

  // 3. Single batch call voor ALLE symbols
  const quotes = await fetch('http://localhost:8080/api/v1/stocks/quotes', {
    method: 'POST',
    body: JSON.stringify({ symbols: Array.from(allSymbols) })
  }).then(r => r.json());

  // 4. Enrich articles
  return articles.map(article => ({
    ...article,
    liveStockData: article.ai_enrichment?.entities?.stock_tickers?.map(t => ({
      ...t,
      quote: quotes.quotes[t.symbol]
    }))
  }));
}
```

**Benefit:** 1 batch call in plaats van N individuele calls!

---

### Tip 3: Real-time Updates

```typescript
// Auto-refresh every minute (cached = no API cost!)
function StockTicker({ symbol }: { symbol: string }) {
  const [quote, setQuote] = useState(null);

  useEffect(() => {
    const fetchQuote = async () => {
      const data = await fetch(`http://localhost:8080/api/v1/stocks/quote/${symbol}`)
        .then(r => r.json());
      setQuote(data);
    };

    fetchQuote();
    
    // Refresh every minute
    // Within 5 min = cached = GRATIS! 🎉
    const interval = setInterval(fetchQuote, 60 * 1000);
    return () => clearInterval(interval);
  }, [symbol]);

  if (!quote) return null;

  return (
    <div className={`ticker ${quote.change >= 0 ? 'up' : 'down'}`}>
      {quote.symbol}: ${quote.price.toFixed(2)} 
      ({quote.change >= 0 ? '+' : ''}{quote.changePercent.toFixed(2)}%)
    </div>
  );
}
```

---

## 🎨 Complete Frontend Example

### Full-Featured Stock Dashboard

```typescript
// pages/stocks.tsx
import { useEffect, useState } from 'react';

export default function StocksDashboard() {
  const [gainers, setGainers] = useState([]);
  const [losers, setLosers] = useState([]);
  const [sectors, setSectors] = useState([]);
  const [earnings, setEarnings] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function loadDashboard() {
      const [gainersRes, losersRes, sectorsRes, earningsRes] = await Promise.all([
        fetch('http://localhost:8080/api/v1/stocks/market/gainers'),
        fetch('http://localhost:8080/api/v1/stocks/market/losers'),
        fetch('http://localhost:8080/api/v1/stocks/sectors'),
        fetch('http://localhost:8080/api/v1/stocks/earnings')
      ]);

      const [gainersData, losersData, sectorsData, earningsData] = await Promise.all([
        gainersRes.json(),
        losersRes.json(),
        sectorsRes.json(),
        earningsRes.json()
      ]);

      setGainers(gainersData.gainers || []);
      setLosers(losersData.losers || []);
      setSectors(sectorsData.sectors || []);
      setEarnings(earningsData.earnings || []);
      setLoading(false);
    }

    loadDashboard();
    
    // Refresh every 5 minutes
    const interval = setInterval(loadDashboard, 5 * 60 * 1000);
    return () => clearInterval(interval);
  }, []);

  if (loading) return <div>Loading market data...</div>;

  return (
    <div className="dashboard">
      <h1>📊 Market Dashboard</h1>
      
      <div className="grid">
        {/* Gainers */}
        <section className="card">
          <h2>📈 Top Gainers</h2>
          {gainers.slice(0, 5).map(stock => (
            <div key={stock.symbol} className="stock-row gain">
              <span className="symbol">{stock.symbol}</span>
              <span className="name">{stock.name}</span>
              <span className="price">${stock.price.toFixed(2)}</span>
              <span className="change">+{stock.changePercent.toFixed(2)}%</span>
            </div>
          ))}
        </section>

        {/* Losers */}
        <section className="card">
          <h2>📉 Top Losers</h2>
          {losers.slice(0, 5).map(stock => (
            <div key={stock.symbol} className="stock-row loss">
              <span className="symbol">{stock.symbol}</span>
              <span className="name">{stock.name}</span>
              <span className="price">${stock.price.toFixed(2)}</span>
              <span className="change">{stock.changePercent.toFixed(2)}%</span>
            </div>
          ))}
        </section>

        {/* Sectors */}
        <section className="card">
          <h2>🏭 Sector Performance</h2>
          {sectors.map(sector => (
            <div key={sector.sector} className="sector-row">
              <span className="sector-name">{sector.sector}</span>
              <div className="sector-bar">
                <div 
                  className={`bar ${sector.changePercent >= 0 ? 'positive' : 'negative'}`}
                  style={{ width: `${Math.abs(sector.changePercent) * 20}%` }}
                />
              </div>
              <span className={sector.changePercent >= 0 ? 'positive' : 'negative'}>
                {sector.changePercent >= 0 ? '+' : ''}{sector.changePercent.toFixed(2)}%
              </span>
            </div>
          ))}
        </section>

        {/* Earnings */}
        <section className="card">
          <h2>📅 Upcoming Earnings</h2>
          {earnings.slice(0, 5).map(e => (
            <div key={`${e.symbol}-${e.date}`} className="earnings-row">
              <span className="symbol">{e.symbol}</span>
              <span className="date">{new Date(e.date).toLocaleDateString()}</span>
              <span className="estimate">EPS: ${e.epsEstimated}</span>
            </div>
          ))}
        </section>
      </div>

      <style jsx>{`
        .dashboard { padding: 20px; }
        .grid { 
          display: grid; 
          grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
          gap: 20px;
        }
        .card { 
          background: white;
          padding: 20px;
          border-radius: 8px;
          box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .stock-row, .sector-row, .earnings-row {
          display: flex;
          justify-content: space-between;
          padding: 8px 0;
          border-bottom: 1px solid #eee;
        }
        .gain { color: #22c55e; }
        .loss { color: #ef4444; }
        .positive { color: #22c55e; }
        .negative { color: #ef4444; }
        .symbol { font-weight: bold; }
      `}</style>
    </div>
  );
}
```

**Deploy:**
```bash
cd frontend
npm install
npm run dev
```

Open http://localhost:3000/stocks → **Instant market dashboard!** 🎉

---

## 🔍 Debugging & Troubleshooting

### Probleem 1: "No data found for symbol"

**Oplossing:**
```bash
# Check of symbol correct is
curl http://localhost:8080/api/v1/stocks/search?q=ASML

# Gebruik correcte symbol uit search results
```

---

### Probleem 2: "API error (status 401)"

**Oplossing:**
```bash
# Check API key in .env
echo $STOCK_API_KEY

# Test API key direct
curl "https://financialmodelingprep.com/api/v3/quote/AAPL?apikey=YOUR_KEY"
```

---

### Probleem 3: Slow Performance

**Check:**
```bash
# Is Redis running?
redis-cli ping

# Check cache stats
curl http://localhost:8080/api/v1/stocks/stats
```

**Expected:**
```json
{
  "cache": {
    "enabled": true,
    "cached_quotes": 45,
    "cached_profiles": 12
  }
}
```

If `enabled: false`, check Redis connection!

---

### Probleem 4: Rate Limit Errors

**Check logs:**
```bash
tail -f logs/app.log | grep "rate limit"
```

**Oplossing:**
```bash
# Verhoog cache TTL in .env
STOCK_API_CACHE_TTL_MINUTES=10  # Was 5

# Of verlaag rate limit
STOCK_API_RATE_LIMIT_PER_MINUTE=20  # Was 30
```

---

## 📈 Next Steps

### Immediate (< 1 dag)

1. ✅ Test alle endpoints
2. ✅ Deploy naar development
3. ✅ Integreer in frontend
4. ✅ Monitor API usage

### Short-term (< 1 week)

1. 📊 Build dashboard met market widgets
2. 📰 Combineer FMP news met je Nederlandse bronnen
3. 📈 Add historical price charts
4. 🎯 Implement earnings alerts

### Long-term (< 1 maand)

1. 🔔 Real-time price alerts
2. 💼 Portfolio tracking features
3. 📊 Advanced analytics dashboard
4. 🤖 AI-powered trading signals

---

## 📚 Complete Feature List

### Implemented (15 endpoints) ✅

- [x] Real-time stock quotes (single + batch)
- [x] Company profiles
- [x] Stock-specific news
- [x] Historical price data (OHLC)
- [x] Key financial metrics & ratios
- [x] Earnings calendar
- [x] Company search
- [x] Market gainers/losers/actives
- [x] Sector performance
- [x] Analyst ratings
- [x] Price target consensus
- [x] Auto-enrichment workflow
- [x] Multi-layer caching
- [x] Batch API optimization
- [x] Comprehensive documentation

### Available in FMP (Not Yet Implemented)

**Easy to add (< 1 hour each):**
- [ ] Insider trading data
- [ ] Institutional ownership (13F)
- [ ] Stock screener
- [ ] Technical indicators (RSI, MACD, SMA)
- [ ] Dividend data
- [ ] Stock splits calendar
- [ ] ETF holdings
- [ ] Forex rates
- [ ] Crypto prices
- [ ] Economic indicators

**Want één van deze features? Laat het weten!** 🚀

---

## 🎉 Success Checklist

Verify your implementation:

- [x] ✅ API key configured in `.env`
- [x] ✅ Backend compiles zonder errors
- [x] ✅ Redis running (optional maar sterk aanbevolen)
- [x] ✅ Database migration 006 applied
- [x] ✅ Test endpoints met curl
- [x] ✅ Check logs voor batch API messages
- [x] ✅ Verify cache hit/miss in logs
- [x] ✅ Test auto-enrichment workflow

**All checked?** You're ready for production! 🚀

---

## 📞 Support

**Documentatie:**
- [Complete API Reference](../api/stock-api-reference.md)
- [Cost Optimization Report](./cost-optimization-report.md)
- [Implementation Details](./fmp-integration-complete.md)

**Vragen?**
- Check logs: `tail -f logs/app.log`
- Test endpoints: See testing scenarios above
- Monitor cache: `GET /api/v1/stocks/stats`

---

**Happy Trading! 📈💰**