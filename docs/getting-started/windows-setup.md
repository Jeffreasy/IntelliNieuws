# Windows Setup Guide voor Nieuws Scraper

Deze guide helpt je om de Nieuws Scraper backend op Windows te draaien **zonder Docker**.

## 📋 Prerequisites

1. **Go 1.22+** - [Download hier](https://go.dev/dl/)
2. **PostgreSQL 16+** - [Download hier](https://www.postgresql.org/download/windows/) of via [winget](https://github.com/microsoft/winget-cli): `winget install PostgreSQL.PostgreSQL`
3. **Redis** (optioneel - voor rate limiting) - [Download Memurai](https://www.memurai.com/) of [Redis voor Windows](https://github.com/microsoftarchive/redis/releases)
4. **Git** - [Download hier](https://git-scm.com/download/win)

## 🚀 Quick Start (10 minuten)

### Stap 1: Clone het project
```powershell
git clone https://github.com/jeffrey/nieuws-scraper.git
cd nieuws-scraper
```

### Stap 2: PostgreSQL installeren en starten

**Optie A: Via winget (aanbevolen)**
```powershell
winget install PostgreSQL.PostgreSQL
```

**Optie B: Handmatige installatie**
1. Download van https://www.postgresql.org/download/windows/
2. Installeer met standaard instellingen
3. Onthoud het wachtwoord voor de `postgres` gebruiker

**PostgreSQL controleren:**
```powershell
# Check of PostgreSQL service draait
Get-Service -Name postgresql*

# Als niet actief, start het
Start-Service -Name postgresql-x64-16
```

### Stap 3: Redis installeren (optioneel)

Redis is alleen nodig voor API rate limiting. Als je het niet installeert, werkt de API nog steeds.

**Memurai (aanbevolen voor Windows):**
```powershell
winget install Memurai.Memurai-Developer
```

Of download van: https://www.memurai.com/

### Stap 4: Project setup

```powershell
# Run setup script
.\scripts\setup.ps1
```

Dit script doet het volgende:
- Maakt `.env` bestand aan
- Downloadt Go dependencies
- Controleert of PostgreSQL en Redis draaien

### Stap 5: Database aanmaken

```powershell
# Run database setup script
.\scripts\create-db.ps1
```

Of handmatig:
```powershell
# Open PostgreSQL command line
psql -U postgres

# In psql:
CREATE DATABASE nieuws_scraper;
\c nieuws_scraper
\i migrations/001_create_tables.sql
\q
```

### Stap 6: Configuratie aanpassen

Bewerk `.env` met Notepad of je favoriete editor:
```powershell
notepad .env
```

Belangrijke instellingen:
```env
# Database - pas aan als je andere credentials gebruikt
POSTGRES_USER=postgres
POSTGRES_PASSWORD=jouw-wachtwoord-hier
POSTGRES_DB=nieuws_scraper

# API Security - kies een sterke API key
API_KEY=jouw-geheime-key-hier
```

### Stap 7: Start de API

```powershell
.\scripts\start.ps1
```

De API is nu beschikbaar op `http://localhost:8080`

### Stap 8: Controleer of het werkt

**In PowerShell:**
```powershell
# Health check
Invoke-WebRequest http://localhost:8080/health

# Artikelen ophalen
Invoke-WebRequest http://localhost:8080/api/v1/articles
```

**In je browser:**
- Health check: http://localhost:8080/health
- Artikelen: http://localhost:8080/api/v1/articles

## 💻 Development Commands

### Build de API
```powershell
go build -o bin/api.exe ./cmd/api
```

### Run lokaal
```powershell
# Met script (controleert database verbinding)
.\scripts\start.ps1

# Of direct
go run ./cmd/api/main.go
```

### Tests draaien
```powershell
go test ./...

# Met coverage
go test -cover ./...
```

### Code formatteren
```powershell
go fmt ./...
```

### Dependencies opschonen
```powershell
go mod tidy
```

## 📊 API Testen

### Met PowerShell (Invoke-WebRequest)
```powershell
# Health check
Invoke-WebRequest http://localhost:8080/health

# Artikelen ophalen
Invoke-WebRequest http://localhost:8080/api/v1/articles

# Specifieke bron
Invoke-WebRequest "http://localhost:8080/api/v1/articles?source=nu.nl&limit=5"

# Scraping triggeren (vervang API key)
Invoke-WebRequest -Method POST `
  -Uri http://localhost:8080/api/v1/scrape `
  -Headers @{"X-API-Key"="your-secret-key-here"}
```

### Met curl (als geïnstalleerd)
```powershell
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/articles
curl -X POST http://localhost:8080/api/v1/scrape -H "X-API-Key: your-key"
```

## 🔧 Troubleshooting

### "Cannot connect to database"

**Controleer of PostgreSQL draait:**
```powershell
Get-Service -Name postgresql*
```

**Als service niet draait:**
```powershell
Start-Service -Name postgresql-x64-16
```

**Test database verbinding:**
```powershell
psql -U postgres -d nieuws_scraper -c "SELECT version();"
```

### "Redis connection failed"

Dit is een waarschuwing, geen error. De API werkt zonder Redis, maar zonder rate limiting.

Om Redis te starten (als geïnstalleerd):
```powershell
# Memurai
Start-Service Memurai

# Of Redis
redis-server
```

### "Port already in use"

```powershell
# Check welke process poort 8080 gebruikt
netstat -ano | findstr :8080

# Wijzig poort in .env
# API_PORT=8081
```

### "go: command not found"

1. Installeer Go van https://go.dev/dl/
2. Herstart PowerShell
3. Test met: `go version`

### "psql: command not found"

PostgreSQL `bin` folder toevoegen aan PATH:
```powershell
# Voeg toe aan PATH (pas pad aan)
$env:Path += ";C:\Program Files\PostgreSQL\16\bin"
```

Of permanent in System Environment Variables.

### Database wachtwoord issues

Als je het postgres wachtwoord bent vergeten:

1. Open `pg_hba.conf` (meestal in `C:\Program Files\PostgreSQL\16\data\`)
2. Wijzig `md5` naar `trust` voor localhost
3. Herstart PostgreSQL service
4. Wijzig wachtwoord met `psql`
5. Wijzig `trust` terug naar `md5`
6. Herstart PostgreSQL weer

## 📝 Database Beheer

### Migrations draaien
```powershell
# Met script
.\scripts\create-db.ps1

# Of handmatig
psql -U postgres -d nieuws_scraper -f migrations/001_create_tables.sql
```

### Database resetten
```powershell
# Drop en hermaak database
psql -U postgres -c "DROP DATABASE IF EXISTS nieuws_scraper;"
psql -U postgres -c "CREATE DATABASE nieuws_scraper;"
psql -U postgres -d nieuws_scraper -f migrations/001_create_tables.sql
```

### Database backup
```powershell
# Backup maken
pg_dump -U postgres nieuws_scraper > backup.sql

# Backup terugzetten
psql -U postgres -d nieuws_scraper < backup.sql
```

## 🎯 Development Workflow

1. **Code wijzigen**
2. **Test lokaal:**
   ```powershell
   go run ./cmd/api/main.go
   ```
3. **Of build en run:**
   ```powershell
   go build -o bin/api.exe ./cmd/api
   .\bin\api.exe
   ```

## 🌐 Productie Deployment

Voor Windows Server productie:

### Optie 1: Als Windows Service (met NSSM)

```powershell
# Build binary
go build -o api.exe ./cmd/api

# Download NSSM van https://nssm.cc/

# Installeer als service
nssm install NieuwsScraperAPI C:\path\to\api.exe

# Start service
nssm start NieuwsScraperAPI
```

### Optie 2: Met Task Scheduler

1. Build: `go build -o api.exe ./cmd/api`
2. Open Task Scheduler
3. Create Task: "Nieuws Scraper API"
4. Trigger: At startup
5. Action: Start program `C:\path\to\api.exe`
6. Settings: "Run whether user is logged on or not"

### Optie 3: IIS Reverse Proxy

Als je IIS gebruikt, configureer een reverse proxy naar `localhost:8080`.

## 💡 Tips

- **Windows Terminal** voor betere PowerShell ervaring
- **VSCode met Go extension** is ideaal voor development
- **pgAdmin** voor grafische PostgreSQL beheer
- **Redis Insight** voor Redis monitoring (als je Redis gebruikt)
- Wijzig poorten in `.env` als er conflicten zijn
- Log level aanpassen met `LOG_LEVEL` in `.env`

## 🔐 Security Tips

- Kies een sterke `API_KEY` in productie
- Gebruik SSL/TLS voor database connecties in productie
- Run API niet als Administrator/SYSTEM user
- Configureer Windows Firewall voor API poort
- Gebruik environment variables voor gevoelige data (niet `.env` in productie)

## 📊 Performance Tips

- **Redis**: Installeer Redis voor betere API performance
- **Connection Pooling**: Standaard al geconfigureerd in de code
- **Indexes**: Database indexes zijn al aangemaakt in migrations
- **Rate Limiting**: Configureer in `.env` voor je use case

## 🆘 Support

Als je problemen hebt:
1. Check de logs in de console
2. Controleer `.env` configuratie
3. Test database en Redis verbindingen
4. Open een GitHub Issue met error details

## 📚 Meer Info

- Volledige API documentatie: zie [README.md](README.md)
- Go documentatie: https://go.dev/doc/
- PostgreSQL docs: https://www.postgresql.org/docs/
- Redis docs: https://redis.io/docs/

---

**Happy Coding! 🚀**