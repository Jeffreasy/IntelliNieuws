# Scripts Directory

Professional utility scripts for the **Nieuws Scraper** backend project.

> **⚠️ Note for Frontend Developers**: These scripts are backend administration tools that are **not directly used** by the frontend. The frontend communicates via REST API endpoints documented in [`docs/frontend/api-reference.md`](../docs/frontend/api-reference.md).

---

## 📁 Directory Structure

```
scripts/
├── setup/          # Initial project setup and configuration
├── migrations/     # Database migration and optimization scripts
├── docker/         # Docker container management
├── operations/     # Runtime operations and service management
├── testing/        # Testing and validation scripts
├── tools/          # Utility tools (Go programs for database inspection)
└── README.md       # This file
```

---

## 🚀 Quick Start

### First Time Setup
```powershell
# Windows
.\scripts\setup\setup.ps1

# Unix/Linux
./scripts/setup/setup.sh
```

### Start the Application
```powershell
# Local development (without Docker)
.\scripts\operations\start.ps1

# With Docker
.\scripts\docker\docker-run.ps1
```

### Run Database Migrations
```powershell
# Apply all migrations
.\scripts\migrations\apply-ai-migration.ps1
.\scripts\migrations\apply-content-migration.ps1
.\scripts\migrations\apply-email-migration.ps1
```

---

## 📂 Directory Reference

### 🔧 setup/
Initial project setup and database creation.

| Script | Purpose | Usage |
|--------|---------|-------|
| [`setup.ps1`](setup/setup.ps1) | Windows setup wizard | `.\scripts\setup\setup.ps1` |
| [`setup.sh`](setup/setup.sh) | Unix/Linux setup script | `./scripts/setup/setup.sh` |
| [`create-db.ps1`](setup/create-db.ps1) | Create PostgreSQL database | `.\scripts\setup\create-db.ps1` |

**When to use:**
- First time project setup
- After cloning the repository
- When setting up a new development environment

---

### 🗄️ migrations/
Database schema changes and optimizations.

| Script | Purpose | Usage |
|--------|---------|-------|
| [`apply-ai-migration.ps1`](migrations/apply-ai-migration.ps1) | Add AI processing columns | `.\scripts\migrations\apply-ai-migration.ps1` |
| [`apply-content-migration.ps1`](migrations/apply-content-migration.ps1) | Add content extraction fields | `.\scripts\migrations\apply-content-migration.ps1` |
| [`apply-email-migration.ps1`](migrations/apply-email-migration.ps1) | Create emails table | `.\scripts\migrations\apply-email-migration.ps1` |
| [`apply-optimizations.ps1`](migrations/apply-optimizations.ps1) | Apply performance indexes | `.\scripts\migrations\apply-optimizations.ps1` |
| [`refresh-materialized-views.ps1`](migrations/refresh-materialized-views.ps1) | Refresh trending views | `.\scripts\migrations\refresh-materialized-views.ps1` |

**Migration Order:**
1. `apply-ai-migration.ps1` - AI functionality
2. `apply-content-migration.ps1` - Content extraction
3. `apply-email-migration.ps1` - Email integration
4. `apply-optimizations.ps1` - Performance tuning

---

### 🐳 docker/
Docker container lifecycle management.

| Script | Purpose | Usage |
|--------|---------|-------|
| [`docker-run.ps1`](docker/docker-run.ps1) | Interactive Docker menu | `.\scripts\docker\docker-run.ps1` |
| [`docker-cleanup-and-restart.ps1`](docker/docker-cleanup-and-restart.ps1) | Clean rebuild containers | `.\scripts\docker\docker-cleanup-and-restart.ps1` |

**Docker Menu Options:**
1. Start all services (PostgreSQL + Redis + App)
2. Stop all services
3. View logs
4. Rebuild and restart
5. Clean up (remove containers and volumes)
6. Check service status

---

### ⚙️ operations/
Runtime service management and operations.

| Script | Purpose | Usage |
|--------|---------|-------|
| [`start.ps1`](operations/start.ps1) | Start API server locally | `.\scripts\operations\start.ps1` |
| [`restart-with-fmp.ps1`](operations/restart-with-fmp.ps1) | Restart with FMP API | `.\scripts\operations\restart-with-fmp.ps1` |
| [`fix-sentiment-and-restart.ps1`](operations/fix-sentiment-and-restart.ps1) | Fix sentiment + restart | `.\scripts\operations\fix-sentiment-and-restart.ps1` |

**Common Operations:**
- **Start server:** `.\scripts\operations\start.ps1`
- **Restart with config changes:** `.\scripts\operations\restart-with-fmp.ps1`
- **Troubleshoot sentiment:** `.\scripts\operations\fix-sentiment-and-restart.ps1`

---

### 🧪 testing/
Testing and validation scripts for various features.

| Script | Purpose | Usage |
|--------|---------|-------|
| [`test-scraper.ps1`](testing/test-scraper.ps1) | Test scraping functionality | `.\scripts\testing\test-scraper.ps1` |
| [`test-performance.ps1`](testing/test-performance.ps1) | Performance benchmarks | `.\scripts\testing\test-performance.ps1` |
| [`test-sentiment-analysis.ps1`](testing/test-sentiment-analysis.ps1) | Test AI sentiment analysis | `.\scripts\testing\test-sentiment-analysis.ps1` |
| [`test-email-integration.ps1`](testing/test-email-integration.ps1) | Test email processing | `.\scripts\testing\test-email-integration.ps1` |
| [`test-fmp-integration.ps1`](testing/test-fmp-integration.ps1) | Test FMP stock API (full) | `.\scripts\testing\test-fmp-integration.ps1` |
| [`test-fmp-free-tier.ps1`](testing/test-fmp-free-tier.ps1) | Test FMP free tier only | `.\scripts\testing\test-fmp-free-tier.ps1` |

**Testing Workflow:**
1. Start the backend: `.\scripts\operations\start.ps1`
2. Run tests: `.\scripts\testing\test-scraper.ps1`
3. Check performance: `.\scripts\testing\test-performance.ps1`

---

### 🛠️ tools/
Go-based utility programs for database inspection and management.

| Tool | Purpose | Build Command | Run Command |
|------|---------|--------------|-------------|
| [`list-tables/`](tools/list-tables/) | List all DB tables + stats | `go build -o tools/list-tables/list-tables.exe tools/list-tables/list-tables.go` | `.\tools\list-tables\list-tables.exe` |
| [`migrate-ai/`](tools/migrate-ai/) | AI migration checker | `go build -o tools/migrate-ai/migrate-ai.exe tools/migrate-ai/migrate-ai.go` | `.\tools\migrate-ai\migrate-ai.exe` |
| [`test-job-tracking/`](tools/test-job-tracking/) | Test scraping jobs | `go build -o tools/test-job-tracking/test-job-tracking.exe tools/test-job-tracking/test-job-tracking.go` | `.\tools\test-job-tracking\test-job-tracking.exe` |
| [`list-tables.ps1`](tools/list-tables.ps1) | PowerShell table lister | N/A | `.\scripts\tools\list-tables.ps1` |

**Note on Go Tools:**
Each Go script is in its own directory to avoid package conflicts. You can run them directly with `go run` or build first for faster execution.

**Quick Run (no build required):**
```powershell
# List all database tables
go run .\scripts\tools\list-tables\list-tables.go

# Check AI migration status
go run .\scripts\tools\migrate-ai\migrate-ai.go

# Test job tracking
go run .\scripts\tools\test-job-tracking\test-job-tracking.go
```

---

## 🎯 Common Workflows

### New Project Setup
```powershell
# 1. Initial setup
.\scripts\setup\setup.ps1

# 2. Create database
.\scripts\setup\create-db.ps1

# 3. Apply migrations
.\scripts\migrations\apply-ai-migration.ps1
.\scripts\migrations\apply-content-migration.ps1
.\scripts\migrations\apply-email-migration.ps1
.\scripts\migrations\apply-optimizations.ps1

# 4. Start application
.\scripts\operations\start.ps1
```

### Daily Development
```powershell
# Start server
.\scripts\operations\start.ps1

# In another terminal - test scraper
.\scripts\testing\test-scraper.ps1

# Check performance
.\scripts\testing\test-performance.ps1
```

### Database Inspection
```powershell
# Quick view
.\scripts\tools\list-tables.ps1

# Detailed view (Go tool)
go run .\scripts\tools\list-tables\list-tables.go
```

### Troubleshooting
```powershell
# Restart with clean state
.\scripts\docker\docker-cleanup-and-restart.ps1

# Fix sentiment issues
.\scripts\operations\fix-sentiment-and-restart.ps1

# Check AI migration status
go run .\scripts\tools\migrate-ai\migrate-ai.go
```

---

## 📝 Environment Configuration

All scripts read from the `.env` file in the project root. Key variables:

```env
# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=nieuws_scraper

# Redis (optional)
REDIS_HOST=localhost
REDIS_PORT=6379

# API
API_KEY=test123geheim
PORT=8080

# Features
ENABLE_AI_PROCESSING=true
ENABLE_FULL_CONTENT_EXTRACTION=true
EMAIL_ENABLED=true
```

---

## 🔒 Security Notes

1. **API Keys**: Never commit real API keys to version control
2. **Compiled Files**: `.exe` files are excluded from git (in `.gitignore`)
3. **Protected Endpoints**: Some scripts use `X-API-Key` header for authentication
4. **Database Credentials**: Keep `.env` file secure and never commit it

---

## 🌐 Frontend Integration

**Frontend developers should use the REST API, not these scripts:**

```javascript
// Example: Fetch articles
const response = await fetch('http://localhost:8080/api/v1/articles?limit=20');
const data = await response.json();

// Example: Trigger scraping (requires API key)
await fetch('http://localhost:8080/api/v1/scrape', {
  method: 'POST',
  headers: { 'X-API-Key': 'test123geheim' }
});
```

**API Documentation:**
- Full API Reference: [`docs/api/stock-api-reference.md`](../docs/api/stock-api-reference.md)
- Frontend Guide: [`docs/frontend/quickstart.md`](../docs/frontend/quickstart.md)

---

## 🆘 Getting Help

### Backend Not Starting?
1. Check PostgreSQL is running: `Test-NetConnection localhost -Port 5432`
2. Verify `.env` configuration
3. Check logs: `docker logs nieuws-scraper-app`

### Database Issues?
1. Inspect tables: `go run .\scripts\tools\list-tables\list-tables.go`
2. Check migration status: `go run .\scripts\tools\migrate-ai\migrate-ai.go`
3. Re-run migrations if needed

### Performance Issues?
1. Run diagnostics: `.\scripts\testing\test-performance.ps1`
2. Apply optimizations: `.\scripts\migrations\apply-optimizations.ps1`
3. Refresh views: `.\scripts\migrations\refresh-materialized-views.ps1`

---

## 📚 Additional Documentation

- **Project Overview**: [`README.md`](../README.md)
- **API Reference**: [`docs/api/`](../docs/api/)
- **Feature Docs**: [`docs/features/`](../docs/features/)
- **Deployment**: [`docs/deployment/`](../docs/deployment/)
- **Operations**: [`docs/operations/`](../docs/operations/)

---

## 🔄 Version History

- **v2.0** - Reorganized script structure into logical categories
- **v1.5** - Added FMP stock API integration scripts
- **v1.4** - Added email integration testing
- **v1.3** - Added AI sentiment analysis scripts
- **v1.2** - Added Docker management scripts
- **v1.1** - Added database optimization scripts
- **v1.0** - Initial script collection

---

**Last Updated**: 2025-01-29  
**Maintained by**: Backend Team  
**Questions?** Check [`docs/operations/troubleshooting.md`](../docs/operations/troubleshooting.md)