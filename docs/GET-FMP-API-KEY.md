# 🔑 FMP API Key Verkrijgen - Stap voor Stap

## Waarom Nieuwe Key Nodig?

De API key in het DocumentatieAPI document (`ePj53WDsqerUu3HEAWB1dMetoLuOmZ8v`) is een **voorbeeld/documentatie key** die niet werkt voor echte API calls.

Je hebt je **eigen gratis FMP API key** nodig.

---

## ✅ Stap 1: Account Aanmaken (2 minuten)

1. **Ga naar:** https://site.financialmodelingprep.com/developer/docs/

2. **Klik op "Get Free API Key"** of "Sign Up"

3. **Vul formulier in:**
   - Email
   - Password
   - (Optioneel: bedrijfsnaam)

4. **Bevestig je email**

5. **Log in op dashboard**

---

## ✅ Stap 2: API Key Kopiëren

1. **Ga naar Dashboard:** https://site.financialmodelingprep.com/developer/docs/

2. **Zoek je API Key:**
   - Staat direct zichtbaar op de dashboard pagina
   - Format: Lange alphanumeric string (bijv: `abc123def456...`)

3. **Kopieer de key**

---

## ✅ Stap 3: Configureren in .env

Open `.env` en update:

```bash
# Stock API Configuration
STOCK_API_PROVIDER=fmp
STOCK_API_KEY=JE_NIEUWE_FMP_API_KEY_HIER  # ← Plak hier je key
STOCK_API_CACHE_TTL_MINUTES=5
STOCK_API_RATE_LIMIT_PER_MINUTE=30
STOCK_API_ENABLE_CACHE=true
```

---

## ✅ Stap 4: Backend Herstarten

```powershell
# Stop huidige backend (Ctrl+C in terminal)

# Rebuild en start
go build -o api.exe ./cmd/api
.\api.exe
```

---

## ✅ Stap 5: Testen

```powershell
# Test single quote
curl http://localhost:8080/api/v1/stocks/quote/AAPL

# Test batch quotes
curl -X POST http://localhost:8080/api/v1/stocks/quotes `
  -H "Content-Type: application/json" `
  -d '{\"symbols\": [\"AAPL\", \"MSFT\"]}'

# Run volledige test suite
.\scripts\test-fmp-integration.ps1
```

**Verwacht resultaat:**
```json
{
  "symbol": "AAPL",
  "name": "Apple Inc.",
  "price": 178.50,
  "change": 2.30,
  "change_percent": 1.31
}
```

---

## 🎁 Gratis Tier Details

**FMP Free Tier:**
- ✅ 250 API calls per dag
- ✅ Real-time stock quotes
- ✅ Historical data
- ✅ Company profiles
- ✅ Financial statements
- ✅ News & earnings
- ✅ Market data
- ❌ Geen credit card vereist

**Met onze batch optimization:**
- 250 calls/dag = **tot 25,000 stock quotes/dag mogelijk!**
- Cache hit rate 80% = effectief **125,000 quotes/dag**
- Meer dan genoeg voor 100-1000 artikelen/dag

---

## ⚠️ Alternatief: Alpha Vantage (Werkende Key)

Als je FMP problemen hebt, kan je tijdelijk terug naar Alpha Vantage:

**In .env:**
```bash
STOCK_API_PROVIDER=alphavantage
STOCK_API_KEY=demo  # Gebruik 'demo' key voor testing
STOCK_API_RATE_LIMIT_PER_MINUTE=5
```

**Let op:**
- ⚠️ Alpha Vantage heeft GEEN batch endpoint
- ⚠️ Veel langzamer (5 calls/min vs 30 voor FMP)
- ⚠️ Geen market performance endpoints
- ⚠️ Geen analyst data
- ✅ Maar werkt wel direct

---

## 🔍 Verificatie

### Check 1: Test FMP API Key Direct

```bash
# Test je nieuwe key direct bij FMP
curl "https://financialmodelingprep.com/api/v3/quote/AAPL?apikey=JE_KEY_HIER"
```

**Succes:**
```json
[{
  "symbol": "AAPL",
  "name": "Apple Inc.",
  "price": 178.50
}]
```

**Fout:**
```json
{
  "Error Message": "Invalid API key"
}
```

### Check 2: Backend Logs

Na herstart, check logs:

```bash
# Zou moeten zien:
{"level":"info","message":"Stock service initialized successfully"}
{"level":"info","message":"Initializing stock service"}

# Test API call:
curl http://localhost:8080/api/v1/stocks/quote/AAPL

# Check logs voor:
{"level":"info","component":"stock-service","message":"Fetching quote..."}
```

---

## 💡 Pro Tips

### Tip 1: Test Key Voordat Je Configureert

```powershell
# Test in browser of Postman eerst:
https://financialmodelingprep.com/api/v3/quote/AAPL?apikey=YOUR_KEY

# Als dat werkt, plak key in .env
```

### Tip 2: Monitor API Usage

FMP dashboard toont:
- Daily API calls used
- Remaining calls
- Usage statistics

Check dit regelmatig!

### Tip 3: Enable Redis voor Beste Performance

```bash
# Install Redis (Windows)
# Download from: https://github.com/microsoftarchive/redis/releases

# Start Redis
redis-server

# In .env:
REDIS_HOST=localhost
REDIS_PORT=6379
STOCK_API_ENABLE_CACHE=true
```

Met Redis: **80-90% cache hit rate = bijna gratis!**

---

## 🎯 Verwachte Resultaten

### Met Werkende FMP Key

**Test 1: Single Quote**
```json
{
  "symbol": "AAPL",
  "name": "Apple Inc.",
  "price": 178.50,
  "change": 2.30,
  "change_percent": 1.31,
  "volume": 45000000,
  "market_cap": 2750000000000
}
```

**Test 2: Batch Quotes**
```json
{
  "quotes": {
    "AAPL": { "price": 178.50 },
    "MSFT": { "price": 375.20 }
  },
  "meta": {
    "total": 2,
    "using_batch": true,
    "cost_saving": "50%"
  }
}
```

**Backend Logs:**
```
INFO: ✅ Fetched 2 quotes in single batch API call (saved: 1 calls)
INFO: Stock service connected to AI service for automatic enrichment
```

---

## 🆘 Hulp Nodig?

### FMP Support
- Website: https://site.financialmodelingprep.com
- Docs: https://site.financialmodelingprep.com/developer/docs/
- Status: https://status.financialmodelingprep.com/

### Veelvoorkomende Problemen

**1. "Invalid API key"**
- Check key in FMP dashboard
- Kopieer opnieuw (geen spaties!)
- Test direct bij FMP API

**2. "Rate limit exceeded"**
- Verhoog cache TTL
- Verlaag rate limit in .env
- Enable Redis

**3. "Failed to fetch"**
- Check internet connectie
- Test FMP API status
- Check firewall

---

## 🎉 Zodra Key Werkt

Run test script:
```powershell
.\scripts\test-fmp-integration.ps1
```

**Verwacht:**
```
✅ Quote for ASML: 745.30
✅ Fetched 3 quotes in batch
   Cost Saving: 67%
   Using Batch: True
✅ 15 FMP endpoints working correctly
```

**Dan ben je klaar!** Alle features werken en je bespaart $180/jaar! 🎊

---

**Volgende stap:** Verkrijg je FMP API key en update `.env` 🔑