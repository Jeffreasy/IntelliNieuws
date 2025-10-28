# Stock Ticker Extraction en Integratie

## Overzicht

IntelliNieuws kan automatisch aandelen-tickers uit nieuwsartikelen extraheren via AI en real-time koersdata ophalen via externe APIs. Deze feature is nuttig voor financiële nieuwsanalyse en het volgen van aandelen-gerelateerde artikelen.

## Features

### 1. **Automatische Ticker Extractie**
- AI (OpenAI) detecteert automatisch aandelen-tickers in artikelen
- Ondersteunt Nederlandse aandelen (ASML, Shell, ING, Philips, etc.)
- Ondersteunt Amerikaanse aandelen (AAPL, MSFT, GOOGL, TSLA, NVDA, etc.)
- Extraheert ticker symbol, company name, en exchange

### 2. **Real-time Koersdata**
- Koppeling met Financial Modeling Prep (FMP) API
- Fallback naar Alpha Vantage API
- Cached koersdata voor performance (5 min TTL)
- Support voor multiple tickers tegelijk

### 3. **Database Integratie**
- Stock tickers worden opgeslagen in `ai_stock_tickers` kolom (JSONB)
- GIN index voor snelle queries
- Opslag van stock data in `stock_data` kolom met TTL

## Configuratie

### Environment Variabelen

Voeg toe aan `.env`:

```bash
# Stock API Configuration
STOCK_API_PROVIDER=fmp                    # "fmp" of "alphavantage"
STOCK_API_KEY=your_api_key_here          # Required
STOCK_API_CACHE_TTL_MINUTES=5            # Cache duration
STOCK_API_RATE_LIMIT_PER_MINUTE=30       # API rate limit
STOCK_API_TIMEOUT_SECONDS=10             # Request timeout
STOCK_API_ENABLE_CACHE=true              # Enable Redis caching
```

### API Keys

#### Financial Modeling Prep (Aanbevolen)
- **Gratis tier**: 250 calls/dag
- **Signup**: https://site.financialmodelingprep.com/developer/docs/
- **Coverage**: Wereldwijd, inclusief AEX
- **Features**: Real-time quotes, company profiles, historische data

#### Alpha Vantage (Fallback)
- **Gratis tier**: 5 calls/min, 500 calls/dag
- **Signup**: https://www.alphavantage.co/support/#api-key
- **Coverage**: Wereldwijd
- **Features**: Real-time quotes, technische indicators

## Database Schema

### Migration: 006_add_stock_tickers.sql

```sql
-- Stock tickers extracted by AI
ALTER TABLE articles ADD COLUMN ai_stock_tickers JSONB;

-- Cached stock data from API
ALTER TABLE articles ADD COLUMN stock_data JSONB;
ALTER TABLE articles ADD COLUMN stock_data_updated_at TIMESTAMP;

-- Indexes
CREATE INDEX idx_articles_stock_tickers ON articles USING GIN(ai_stock_tickers);
CREATE INDEX idx_articles_stock_data ON articles(id) WHERE stock_data IS NOT NULL;
```

### Data Structuur

**ai_stock_tickers** (Array of objects):
```json
[
  {
    "symbol": "ASML",
    "name": "ASML Holding",
    "exchange": "AEX",
    "mentions": 3,
    "context": "ASML rapporteerde sterke kwartaalcijfers..."
  }
]
```

**stock_data** (Object mapping symbol to data):
```json
{
  "ASML": {
    "price": 745.30,
    "change": 12.50,
    "change_percent": 1.71,
    "volume": 1250000,
    "market_cap": 295000000000,
    "last_updated": "2024-01-15T14:30:00Z"
  }
}
```

## Gebruik

### Backend - AI Processing

Ticker extraction gebeurt automatisch tijdens AI processing:

```go
// Automatisch tijdens AI processing
enrichment, err := aiService.ProcessArticle(ctx, articleID)
// enrichment.Entities.StockTickers bevat geëxtraheerde tickers

// Artikelen ophalen per ticker
articles, err := aiService.GetArticlesByStockTicker(ctx, "ASML", 10)
```

### Backend - Stock Data Ophalen

```go
// Single quote
quote, err := stockService.GetQuote(ctx, "ASML")

// Multiple quotes (efficient)
symbols := []string{"ASML", "SHELL", "ING"}
quotes, err := stockService.GetMultipleQuotes(ctx, symbols)

// Company profile
profile, err := stockService.GetProfile(ctx, "ASML")
```

### API Endpoints

```bash
# Artikelen met specifieke ticker
GET /api/v1/articles/by-ticker/:ticker?limit=10

# Stock quote data
GET /api/v1/stocks/quote/:symbol

# Multiple quotes
POST /api/v1/stocks/quotes
Body: {"symbols": ["ASML", "SHELL", "ING"]}

# Company profile
GET /api/v1/stocks/profile/:symbol
```

### Response Voorbeelden

**Article met stock tickers:**
```json
{
  "id": 123,
  "title": "ASML boekt recordomzet",
  "ai_enrichment": {
    "entities": {
      "stock_tickers": [
        {
          "symbol": "ASML",
          "name": "ASML Holding",
          "exchange": "AEX",
          "mentions": 2
        }
      ]
    }
  },
  "stock_data": {
    "ASML": {
      "price": 745.30,
      "change": 12.50,
      "change_percent": 1.71,
      "market_cap": 295000000000
    }
  }
}
```

## Performance & Caching

### Cache Strategie

1. **Redis Cache** (5 min TTL)
   - Quotes worden gecached
   - Profiles worden gecached (24 uur)
   - Vermindert API calls met ~80%

2. **Database Cache** (`stock_data` kolom)
   - Secondary cache layer
   - Langere TTL mogelijk
   - Fallback bij Redis failure

3. **Rate Limiting**
   - Configurable per provider
   - Voorkomt API quota overschrijding
   - Queue-based voor multiple requests

### Cost Optimization

**Gratis tier (FMP):**
- 250 calls/dag
- Met caching: ~10,000 artikelen/dag mogelijk
- Cost: $0

**Paid tier (indien nodig):**
- Starter: $15/maand (unlimited calls)
- Professional: $50/maand (extra features)

## Ondersteunde Aandelen

### Nederlandse Aandelen (AEX)
- ASML, Shell (SHELL), ING, Philips (PHIA)
- Unilever (UNA), ASMI, IMCD, DSM
- ABN AMRO (ABN), NN Group (NN)

### Amerikaanse Aandelen
- Tech: AAPL, MSFT, GOOGL, META, AMZN, NVDA
- Auto: TSLA, GM, F
- Finance: JPM, BAC, GS

### Internationale Aandelen
- Europese indices: DAX, CAC40, FTSE
- Aziatische markten: Nikkei, Hang Seng

## Best Practices

### 1. Efficient API Usage
```go
// ❌ Slecht: Individuele calls
for _, symbol := range symbols {
    quote, _ := service.GetQuote(ctx, symbol)
}

// ✅ Goed: Batch requests
quotes, _ := service.GetMultipleQuotes(ctx, symbols)
```

### 2. Cache Warming
```go
// Warm cache voor populaire aandelen
popularStocks := []string{"ASML", "SHELL", "ING", "AAPL", "TSLA"}
go stockService.GetMultipleQuotes(context.Background(), popularStocks)
```

### 3. Error Handling
```go
quote, err := stockService.GetQuote(ctx, symbol)
if err != nil {
    // Fallback: gebruik gecachte data uit database
    cachedData := getFromDatabase(symbol)
    if cachedData != nil && time.Since(cachedData.UpdatedAt) < 1*time.Hour {
        return cachedData
    }
    return nil, err
}
```

## Monitoring

### Metrics

```bash
# Cache statistics
GET /api/v1/stocks/stats

Response:
{
  "cache": {
    "enabled": true,
    "cached_quotes": 45,
    "cached_profiles": 12,
    "hit_rate": 0.82
  },
  "api": {
    "provider": "fmp",
    "rate_limit": 30,
    "calls_today": 156
  }
}
```

### Logs

```json
{
  "level": "info",
  "component": "stock-service",
  "message": "Cache HIT for stock quote: ASML",
  "cache_age": "2m15s"
}
```

## Troubleshooting

### Probleem: Geen tickers gedetecteerd

**Oplossing:**
1. Check of AI processing enabled is (`AI_ENABLED=true`)
2. Check of entities extraction aan staat (`AI_ENABLE_ENTITIES=true`)
3. Verificeer OpenAI API key

### Probleem: API rate limit errors

**Oplossing:**
1. Verhoog `STOCK_API_CACHE_TTL_MINUTES`
2. Verlaag `STOCK_API_RATE_LIMIT_PER_MINUTE`
3. Upgrade naar paid tier
4. Enable Redis caching

### Probleem: Foute ticker symbolen

**Oplossing:**
1. Verbeter AI prompt specificiteit
2. Post-process met validation against known symbols
3. Use company name matching als fallback

## Toekomstige Uitbreidingen

1. **Historical Data**
   - Prijsgrafieken
   - Technische indicators
   - Volume analysis

2. **Alerts**
   - Price threshold alerts
   - News sentiment + price correlation
   - Unusual volume detection

3. **Portfolio Tracking**
   - User watchlists
   - Performance tracking
   - Real-time updates via WebSocket

4. **Advanced Analytics**
   - Sentiment impact on price
   - News-driven trading signals
   - Correlation analysis

## Links

- [FMP API Docs](https://site.financialmodelingprep.com/developer/docs/)
- [Alpha Vantage Docs](https://www.alphavantage.co/documentation/)
- [AI Processing Guide](./ai-processing.md)
- [API Reference](../api/README.md)