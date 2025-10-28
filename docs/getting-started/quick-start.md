# Quick Start Guide

Start met IntelliNieuws in 5 minuten.

## ⚡ Snelle Installatie

### Stap 1: Clone Repository
```bash
git clone https://github.com/yourusername/intellinieuws.git
cd intellinieuws
```

### Stap 2: Database Setup
```powershell
# Windows (PowerShell)
.\scripts\create-db.ps1

# Of handmatig:
psql -U postgres -c "CREATE DATABASE nieuws_scraper;"
psql -U postgres -d nieuws_scraper -f migrations/001_create_tables.sql
psql -U postgres -d nieuws_scraper -f migrations/002_optimize_indexes.sql
psql -U postgres -d nieuws_scraper -f migrations/003_add_ai_columns_simple.sql
psql -U postgres -d nieuws_scraper -f migrations/004_create_trending_materialized_view.sql
```

### Stap 3: Configuratie
```powershell
# Copy environment template
cp .env.example .env

# Edit .env met je settings
notepad .env
```

**Minimale configuratie:**
```env
# Database
DATABASE_PASSWORD=jouw_postgres_wachtwoord

# OpenAI (voor AI features)
OPENAI_API_KEY=sk-jouw-openai-key

# API Security
API_KEY=jouw-geheime-api-key
```

### Stap 4: Build & Run
```powershell
# Build
go build -o api.exe .\cmd\api

# Start
.\api.exe
```

### Stap 5: Verificatie
```powershell
# Health check
curl http://localhost:8080/health

# Expected: {"status": "healthy"}
```

## ✅ Success!

Je backend draait nu op `http://localhost:8080`

## 🎯 Volgende Stappen

1. **Test de API**: `curl http://localhost:8080/api/v1/articles`
2. **Trigger scraping**: Zie [API Examples](../api/examples.md)
3. **Bouw een frontend**: Zie [Frontend Guide](../frontend/README.md)

## 📊 Belangrijke Endpoints

```bash
# Health check
GET http://localhost:8080/health

# List articles
GET http://localhost:8080/api/v1/articles

# AI trending topics
GET http://localhost:8080/api/v1/ai/trending

# Trigger scraping (requires API key)
POST http://localhost:8080/api/v1/scrape
Header: X-API-Key: your-api-key
```

## 🐛 Problemen?

Zie [Troubleshooting](../operations/troubleshooting.md) of [Complete Installation](installation.md)

---

**Je bent klaar! 🎉** Start met de [API Documentation](../api/README.md)