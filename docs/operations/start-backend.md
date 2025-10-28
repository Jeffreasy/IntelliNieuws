# 🚀 Start Backend - Praktische Stappen

**Quick start guide voor het draaien van de geoptimaliseerde backend**

---

## ⚡ SNELLE START (3 Stappen)

### Stap 1: Rebuild de Applicatie
```powershell
# In project directory
cd C:\Users\jeffrey\Desktop\Githubmains\NieuwsScraper

# Build nieuwe versie (met de fix)
go build -o api.exe .\cmd\api
```

### Stap 2: Controleer Database Setup
```powershell
# Check of migraties zijn gedraaid
psql -U postgres -d nieuws_scraper -c "\dt"

# Als je artikelen table NIET ziet, run migraties:
psql -U postgres -d nieuws_scraper -f migrations\001_create_tables.sql
psql -U postgres -d nieuws_scraper -f migrations\002_optimize_indexes.sql
psql -U postgres -d nieuws_scraper -f migrations\003_add_ai_columns_simple.sql

# BELANGRIJK: Nieuwe materialized view (v2.0 feature!)
psql -U postgres -d nieuws_scraper -f migrations\004_create_trending_materialized_view.sql
```

### Stap 3: Start Backend
```powershell
# Start met script (aanbevolen)
.\scripts\start.ps1

# Of direct:
$env:DATABASE_PASSWORD="postgres"; $env:OPENAI_API_KEY="sk-jouw-key-hier"; .\api.exe
```

---

## ✅ VERWACHTE OUTPUT

Als alles goed gaat, zie je:
```
{"level":"info","message":"Starting Nieuws Scraper API service"}
{"level":"info","message":"Successfully connected to database"}
{"level":"info","message":"Database connection pool configured: max=25, min=5"}
{"level":"info","message":"Connection pool pre-warming completed"}
{"level":"info","message":"Cache service initialized with 5min TTL"}
{"level":"info","message":"AI service initialized successfully"}
{"level":"info","message":"AI processor started with interval: 5m0s"}
{"level":"info","message":"Starting API server on :8080"}
```

**Redis waarschuwing is OK:**
```
WAARSCHUWING Redis is niet bereikbaar (optioneel - rate limiting werkt niet)
```
Dit betekent alleen dat API response caching niet werkt, maar de rest werkt wel!

---

## 🔧 TROUBLESHOOTING

### Error: "unrecognized configuration parameter"

**JE HEBT DIT AL GEZIEN** - Ik heb dit zojuist gefixt!

**Oplossing:**
```powershell
# 1. Rebuild met nieuwe code
go build -o api.exe .\cmd\api

# 2. Start opnieuw
.\scripts\start.ps1
```

### Error: "failed to connect to database"

**Oplossing A: Check PostgreSQL Service**
```powershell
# Check of PostgreSQL draait
Get-Service postgresql*

# Als niet running:
Start-Service postgresql-x64-12  # Of jouw versie
```

**Oplossing B: Check Database Exists**
```powershell
# Probeer te connecten
psql -U postgres -l

# Zie je nieuws_scraper in de lijst?
# Zo niet:
psql -U postgres -c "CREATE DATABASE nieuws_scraper;"
```

**Oplossing C: Check Wachtwoord**
```powershell
# Test met correct wachtwoord
psql -U postgres -d nieuws_scraper -c "SELECT 1;"

# Als dit werkt, update .env file met correct wachtwoord
```

### Warning: "Failed to pre-warm connection"

**Dit is NORMAAL** voor oudere PostgreSQL versies. De applicatie blijft gewoon werken!

---

## 📊 VERIFICATIE NA START

### Test 1: Health Check
```powershell
# In nieuwe PowerShell terminal
curl http://localhost:8080/health

# Expected:
# {"status":"success","data":{"status":"healthy",...}}
```

### Test 2: List Articles
```powershell
curl http://localhost:8080/api/v1/articles

# Should return JSON (mogelijk lege array als geen data)
```

### Test 3: Check Optimizations
```powershell
# Check processor stats (nieuwe feature!)
curl http://localhost:8080/api/v1/ai/processor/stats

# Check metrics (nieuwe feature!)
curl http://localhost:8080/health/metrics
```

---

## 🎯 EERSTE KEER SETUP (Complete)

### Als je dit voor het EERST doet:

```powershell
# === DATABASE SETUP ===
# 1. Create database
psql -U postgres -c "CREATE DATABASE nieuws_scraper;"

# 2. Run alle migraties
psql -U postgres -d nieuws_scraper -f migrations\001_create_tables.sql
psql -U postgres -d nieuws_scraper -f migrations\002_optimize_indexes.sql
psql -U postgres -d nieuws_scraper -f migrations\003_add_ai_columns_simple.sql
psql -U postgres -d nieuws_scraper -f migrations\004_create_trending_materialized_view.sql

# === ENVIRONMENT ===
# 3. Edit .env file met jouw settings
notepad .env

# Minimaal nodig:
# DATABASE_PASSWORD=jouw_postgres_wachtwoord
# OPENAI_API_KEY=sk-jouw-openai-key

# === BUILD & RUN ===
# 4. Build
go build -o api.exe .\cmd\api

# 5. Run
.\scripts\start.ps1

# === VERIFY ===
# 6. Test (in nieuwe terminal)
curl http://localhost:8080/health
```

---

## 🔄 DAGELIJKS GEBRUIK

### Backend Starten
```powershell
cd C:\Users\jeffrey\Desktop\Githubmains\NieuwsScraper
.\scripts\start.ps1
```

### Backend Stoppen
```powershell
# Druk op Ctrl+C in de terminal waar api.exe draait
# Of:
taskkill /F /IM api.exe
```

### Check of Backend Draait
```powershell
curl http://localhost:8080/health
```

---

## 📈 FEATURES TESTEN

### Test 1: Scrapen
```powershell
# Trigger scraping (vereist API key)
curl -X POST http://localhost:8080/api/v1/scrape `
  -H "X-API-Key: jouw-api-key-uit-.env"

# Check articles
curl http://localhost:8080/api/v1/articles
```

### Test 2: AI Processing
```powershell
# Trigger AI processing
curl -X POST http://localhost:8080/api/v1/ai/process/trigger `
  -H "X-API-Key: jouw-api-key"

# Check trending topics (NIEUW & SNEL in v2.0!)
curl http://localhost:8080/api/v1/ai/trending
```

### Test 3: Nieuwe Health Endpoints (v2.0)
```powershell
# Comprehensive health
curl http://localhost:8080/health

# Liveness probe
curl http://localhost:8080/health/live

# Readiness probe
curl http://localhost:8080/health/ready

# Detailed metrics
curl http://localhost:8080/health/metrics
```

---

## 🎯 OPTIMALISATIES CHECKLIST

### Check of Optimalisaties Werken

**In de logs (console) let op:**

✅ **Caching:**
```
"Cache HIT for content (key: xxxxx, hits: 2)"
```

✅ **Batch Duplicate Detection:**
```
"Batch duplicate check completed for X URLs"
```

✅ **Worker Pool:**
```
"Parallel batch processing completed: 4 workers"
```

✅ **Dynamic Intervals:**
```
"Adjusted processing interval to 2m0s (queue: 25 articles)"
```

✅ **Circuit Breakers:**
```
Check via: curl http://localhost:8080/api/v1/scraper/stats
Look for "circuit_breakers" in response
```

---

## 💡 PRO TIPS

### Snellere Startup
```powershell
# Pre-warm database (run migrations vooraf)
# Dan start applicatie direct

# Keep binary gebuild
# Dan hoef je niet elke keer opnieuw te builden
```

### Betere Logs
```powershell
# Run in debug mode (meer output)
$env:LOG_LEVEL="debug"; .\api.exe
```

### Monitoring in Real-Time
```powershell
# Terminal 1: Run backend
.\api.exe

# Terminal 2: Watch health
while ($true) { 
    curl http://localhost:8080/health/metrics | ConvertFrom-Json
    Start-Sleep 30 
}
```

---

## 🆘 HULP NODIG?

### PostgreSQL Issues
```powershell
# Check versie
psql -U postgres -c "SELECT version();"

# Check of database bestaat
psql -U postgres -l | findstr nieuws_scraper

# Test connectie
psql -U postgres -d nieuws_scraper -c "SELECT 1;"
```

### Build Issues
```powershell
# Clean build
go clean
go build -o api.exe .\cmd\api

# Check Go version (moet 1.21+)
go version
```

### Port Issues
```powershell
# Check wat draait op port 8080
netstat -ano | findstr :8080

# Change port in .env als nodig:
# API_PORT=8081
```

---

## ✅ SUCCESS CHECKLIST

Na het starten, check:
- [ ] Applicatie start zonder FATAL errors
- [ ] `curl http://localhost:8080/health` returnt "healthy"
- [ ] Logs tonen "Successfully connected to database"
- [ ] Logs tonen "AI processor started"
- [ ] Geen continue errors in output

**Als alle boxes checked: Je backend draait perfect! 🎉**

---

## 🚀 VOLGENDE STAPPEN

### Na Succesvolle Start
1. **Test scraping:** Voeg nieuwsbronnen toe
2. **Test AI:** Process enkele articles
3. **Monitor performance:** Check metrics elke dag
4. **Setup automation:** Scheduled scraping & materialized view refresh

### Voor Productie
1. Setup Windows Service voor auto-start
2. Configure monitoring/alerting
3. Setup database backups
4. Document je specifieke setup

---

**Nu probeer opnieuw:**
```powershell
# 1. Rebuild
go build -o api.exe .\cmd\api

# 2. Start
.\scripts\start.ps1

# 3. Test
curl http://localhost:8080/health
```

**Succes! 🚀**