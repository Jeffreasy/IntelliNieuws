# Scripts Directory

This directory contains various **backend administrative utility scripts** for the Nieuws Scraper project.

> **⚠️ Note voor Frontend Developers**: Deze scripts zijn backend tools die **niet direct** door de frontend gebruikt worden. De frontend communiceert via de REST API endpoints beschreven in [`FRONTEND_API.md`](../FRONTEND_API.md).

## Go Scripts

Each Go script is organized in its own subdirectory to avoid package conflicts:

### list-tables
Location: `scripts/list-tables/list-tables.go`

Lists all database tables, materialized views, and column details.

**Usage:**
```bash
go run scripts/list-tables/list-tables.go
# or build first:
go build -o scripts/list-tables/list-tables.exe scripts/list-tables/list-tables.go
scripts/list-tables/list-tables.exe
```

### migrate-ai
Location: `scripts/migrate-ai/migrate-ai.go`

Applies AI-related database migrations and checks processing status.

**Usage:**
```bash
go run scripts/migrate-ai/migrate-ai.go
# or build first:
go build -o scripts/migrate-ai/migrate-ai.exe scripts/migrate-ai/migrate-ai.go
scripts/migrate-ai/migrate-ai.exe
```

### test-job-tracking
Location: `scripts/test-job-tracking/test-job-tracking.go`

Tests the scraping job tracking functionality.

**Usage:**
```bash
go run scripts/test-job-tracking/test-job-tracking.go
# or build first:
go build -o scripts/test-job-tracking/test-job-tracking.exe scripts/test-job-tracking/test-job-tracking.go
scripts/test-job-tracking/test-job-tracking.exe
```

## PowerShell Scripts

- `apply-ai-migration.ps1` - Apply AI migrations
- `apply-content-migration.ps1` - Apply content migrations
- `apply-optimizations.ps1` - Apply database optimizations
- `create-db.ps1` - Create database
- `list-tables.ps1` - PowerShell version of list-tables
- `refresh-materialized-views.ps1` - Refresh materialized views
- `setup.ps1` - Setup script for Windows
- `setup.sh` - Setup script for Unix/Linux
- `start.ps1` - Start the application
- `test-performance.ps1` - Performance testing
- `test-scraper.ps1` - Scraper testing

## Note on Go Script Organization

Each Go script with a `main` function must be in its own directory. This is because Go treats all `.go` files in the same directory as part of the same package. Having multiple `main` functions in the same directory would cause duplicate declaration errors.

---

## Voor Frontend Developers

De frontend applicatie gebruikt **NIET** deze Go scripts direct. In plaats daarvan communiceert de frontend met de backend via de REST API.

### Hoe de Frontend de Backend Gebruikt

1. **Start de API Server**:
   ```bash
   go run cmd/api/main.go
   # Server draait op http://localhost:8080
   ```

2. **Frontend API Endpoints**:
   De frontend maakt HTTP requests naar endpoints zoals:
   ```javascript
   // Artikelen ophalen
   fetch('http://localhost:8080/api/v1/articles?limit=20')
   
   // Zoeken
   fetch('http://localhost:8080/api/v1/articles/search?q=voetbal')
   
   // Statistieken
   fetch('http://localhost:8080/api/v1/articles/stats')
   
   // AI Sentiment
   fetch('http://localhost:8080/api/v1/ai/sentiment/stats')
   
   // Trending Topics
   fetch('http://localhost:8080/api/v1/ai/trending')
   ```

3. **Volledige API Documentatie**:
   Zie [`FRONTEND_API.md`](../FRONTEND_API.md) voor:
   - Alle beschikbare endpoints
   - Request/response voorbeelden
   - Authentication
   - Error handling
   - TypeScript types
   - Implementatie voorbeelden

### Backend Service Architectuur

```
Frontend (React/Vue/etc)
    ↓ HTTP Requests
API Server (Go) - http://localhost:8080
    ↓
├─ Article Handler (/api/v1/articles)
├─ AI Handler (/api/v1/ai)
├─ Scraper Handler (/api/v1/scrape) [Protected]
└─ Health Handler (/health)
    ↓
PostgreSQL Database
Redis Cache
```

### Admin Scripts (Deze Directory)

De scripts in deze directory zijn **alleen voor backend administratie**:
- **Database inspectie**: Bekijk tabellen en data
- **Migraties**: Update database schema
- **Testing**: Test backend functionaliteit

Deze worden **door developers/admins** uitgevoerd, niet door de frontend.

---