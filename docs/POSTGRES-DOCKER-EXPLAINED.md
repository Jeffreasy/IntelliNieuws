# PostgreSQL in Docker - Volledige Uitleg

Een grondige uitleg van hoe PostgreSQL werkt binnen uw Docker setup voor het NieuwsScraper project.

---

## 📋 Inhoudsopgave

1. [Overzicht](#overzicht)
2. [Docker Container Lifecycle](#docker-container-lifecycle)
3. [Data Persistence](#data-persistence)
4. [Database Initialisatie](#database-initialisatie)
5. [Connecties vanuit de App](#connecties-vanuit-de-app)
6. [Backup Mechanisme](#backup-mechanisme)
7. [Praktische Voorbeelden](#praktische-voorbeelden)

---

## 🎯 Overzicht

### Wat gebeurt er precies?

Wanneer u `docker-compose up -d` uitvoert, gebeurt het volgende:

```
1. Docker trekt PostgreSQL 15 Alpine image (als nog niet aanwezig)
2. Docker creëert een isolated container voor PostgreSQL
3. Docker mount een named volume voor data persistence
4. Docker mount uw migrations folder als read-only
5. PostgreSQL start en voert migraties automatisch uit
6. Health check controleert of database klaar is
7. App container kan nu verbinden via 'postgres' hostname
```

### De Stack Visualisatie

```
┌─────────────────────────────────────────────────────────────┐
│                    DOCKER HOST (Windows)                     │
│                                                               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │         Docker Network: nieuws-scraper-network        │  │
│  │              (172.20.0.0/16)                          │  │
│  │                                                        │  │
│  │  ┌──────────────┐      ┌──────────────┐             │  │
│  │  │   App        │◄─────┤  PostgreSQL  │             │  │
│  │  │  Container   │      │   Container   │             │  │
│  │  │              │      │               │             │  │
│  │  │ Hostname:app │      │ Hostname:     │             │  │
│  │  │ Port: 8080   │      │   postgres    │             │  │
│  │  └──────────────┘      │ Port: 5432    │             │  │
│  │                        └───────┬───────┘             │  │
│  │                                │                      │  │
│  │                                ▼                      │  │
│  │                   ┌────────────────────┐             │  │
│  │                   │  Named Volume      │             │  │
│  │                   │  postgres_data     │             │  │
│  │                   │                    │             │  │
│  │                   │  Persistent Data   │             │  │
│  │                   └────────────────────┘             │  │
│  └────────────────────────────────────────────────────────┘
│                                                               │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  Host Filesystem                                       │ │
│  │  ./migrations/  ──mount──▶  /docker-entrypoint-initdb.d│ │
│  │  ./backups/     ◄──mount──  /backups                  │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔄 Docker Container Lifecycle

### Stap 1: Container Creatie

Wanneer u `docker-compose up -d` uitvoert:

```yaml
# docker-compose.yml
postgres:
  image: postgres:15-alpine          # ← Docker trekt deze image
  container_name: nieuws-scraper-postgres
  restart: unless-stopped
  environment:
    POSTGRES_USER: scraper           # ← Database gebruiker
    POSTGRES_PASSWORD: scraper_password  # ← Database password
    POSTGRES_DB: nieuws_scraper     # ← Database naam
```

**Wat gebeurt er:**

1. **Image Pull:**
   ```bash
   # Docker checkt lokale cache
   # Als niet aanwezig: download van Docker Hub (~80MB)
   docker pull postgres:15-alpine
   ```

2. **Container Aanmaken:**
   ```bash
   # Docker creëert een geïsoleerde container
   docker create \
     --name nieuws-scraper-postgres \
     --env POSTGRES_USER=scraper \
     --env POSTGRES_PASSWORD=scraper_password \
     --env POSTGRES_DB=nieuws_scraper \
     postgres:15-alpine
   ```

3. **Volume Mounting:**
   ```bash
   # Named volume voor data
   -v postgres_data:/var/lib/postgresql/data
   
   # Migrations folder (read-only)
   -v ./migrations:/docker-entrypoint-initdb.d:ro
   ```

### Stap 2: PostgreSQL Initialisatie

**Bij eerste keer opstarten:**

```
PostgreSQL Container Start
    ↓
Controleer: Is /var/lib/postgresql/data leeg?
    ↓ [JA]
Initialiseer database cluster
    ↓
Maak database 'nieuws_scraper'
    ↓
Maak user 'scraper' met password
    ↓
Voer scripts uit in /docker-entrypoint-initdb.d/ (alfabetisch)
    ↓
001_create_tables.sql        → Maak tables
002_optimize_indexes.sql     → Maak indexes
003_add_ai_columns.sql       → Add AI columns
... (alle .sql files)
    ↓
Database klaar! ✅
```

**Bij herstart:**

```
PostgreSQL Container Start
    ↓
Controleer: Is /var/lib/postgresql/data leeg?
    ↓ [NEE - data bestaat al]
Skip initialisatie scripts
    ↓
Start PostgreSQL met bestaande data
    ↓
Database klaar! ✅
```

### Stap 3: Health Check

Docker controleert continu of PostgreSQL healthy is:

```yaml
healthcheck:
  test: ["CMD-SHELL", "pg_isready -U scraper -d nieuws_scraper"]
  interval: 10s        # Check elke 10 seconden
  timeout: 5s          # Max 5 sec wachten
  retries: 5           # 5 pogingen voordat unhealthy
  start_period: 10s    # Wacht 10 sec na start voor eerste check
```

**Health Check Flow:**

```
Container Start
    ↓
Start Period: 10 seconden
    ↓
┌────────────────────────────────┐
│ Health Check Loop (elke 10s)  │
│                                │
│ pg_isready -U scraper -d ...  │
│         ↓                      │
│   [SUCCESS] → Status: healthy  │
│   [FAIL]    → retry (max 5x)  │
│   [5x FAIL] → Status: unhealthy│
└────────────────────────────────┘
```

**Statusen:**

- `starting` - Container net gestart, nog in start_period
- `healthy` - Health check succesvol
- `unhealthy` - 5 mislukte checks op rij

---

## 💾 Data Persistence

### Named Volume: postgres_data

**Wat is een Named Volume?**

Een named volume is een door Docker beheerde storage locatie die **BUITEN** de container bestaat.

```bash
# Waar staat de data?
# Windows: \\wsl$\docker-desktop-data\data\docker\volumes\
# Linux: /var/lib/docker/volumes/

# Volume naam in uw setup:
nieuws-scraper_postgres_data
```

### Hoe werkt het?

```
┌─────────────────────────────────────────────────────────────┐
│  HOST FILESYSTEM (Windows)                                  │
│                                                              │
│  C:\Users\jeffrey\...\docker\volumes\                       │
│  └── nieuws-scraper_postgres_data\                          │
│      └── _data\                                              │
│          ├── base/           ← Database files               │
│          ├── global/         ← System catalogs              │
│          ├── pg_wal/         ← Write-Ahead Log              │
│          ├── pg_stat/        ← Statistics                   │
│          └── pg_xact/        ← Transaction files            │
│                                                              │
│              ↕ MOUNTED TO ↕                                  │
│                                                              │
│  ┌──────────────────────────────────────────────────┐       │
│  │  POSTGRES CONTAINER                              │       │
│  │                                                   │       │
│  │  /var/lib/postgresql/data/                       │       │
│  │  ├── base/       ← Reads/Writes here             │       │
│  │  ├── global/                                     │       │
│  │  ├── pg_wal/                                     │       │
│  │  └── ...                                         │       │
│  └──────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────┘
```

### Waarom is dit belangrijk?

**✅ Container verwijderen = Data blijft bestaan**

```bash
# Stop en verwijder container
docker-compose down

# Container is weg, maar data blijft!
# Volume postgres_data bestaat nog steeds

# Start opnieuw
docker-compose up -d

# PostgreSQL gebruikt dezelfde data! 🎉
# Geen data loss!
```

**❌ Volume verwijderen = Data is weg**

```bash
# DIT WIST ALLES!
docker-compose down -v

# Volume is verwijderd, data is GONE
# Volgende start = Nieuwe lege database
```

---

## 🗂️ Database Initialisatie

### Migrations Folder

```
./migrations/
├── 001_create_tables.sql         ← Eerst
├── 002_optimize_indexes.sql      ← Tweede
├── 003_add_ai_columns.sql        ← Derde
├── 004_create_trending_view.sql  ← Vierde
├── 005_add_content_column.sql    ← Vijfde
├── 006_add_stock_tickers.sql     ← Zesde
└── 007_create_emails_table.sql   ← Laatste
```

### Hoe worden deze uitgevoerd?

**Docker-entrypoint mechanisme:**

```yaml
volumes:
  # Deze mount zorgt voor auto-executie
  - ./migrations:/docker-entrypoint-initdb.d:ro
```

**Wat doet PostgreSQL Docker image?**

1. **Bij EERSTE start** (lege database):

```bash
#!/bin/bash
# Dit draait automatisch in de container

# Check of database leeg is
if [ ! -f /var/lib/postgresql/data/PG_VERSION ]; then
    echo "Initializing database..."
    
    # Initialiseer PostgreSQL
    initdb -D /var/lib/postgresql/data
    
    # Start PostgreSQL tijdelijk
    pg_ctl start
    
    # Voer alle scripts uit in /docker-entrypoint-initdb.d/
    for file in /docker-entrypoint-initdb.d/*.sql; do
        echo "Running $file..."
        psql -U scraper -d nieuws_scraper < "$file"
    done
    
    # Stop tijdelijke instantie
    pg_ctl stop
fi

# Start PostgreSQL normaal
postgres
```

2. **Bij HERSTART** (data bestaat):

```bash
# Skip alle initialisatie
# Start direct met bestaande data
postgres
```

### Volgorde van Uitvoering

**Alfabetisch!** Daarom de nummering:

```
001_create_tables.sql         → Eerst uitgevoerd
002_optimize_indexes.sql      → Daarna
003_add_ai_columns.sql        → Daarna
...
007_create_emails_table.sql   → Laatst
```

**Praktisch voorbeeld:**

```sql
-- migrations/001_create_tables.sql
CREATE TABLE IF NOT EXISTS articles (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    url TEXT UNIQUE NOT NULL,
    -- ...
);
```

```sql
-- migrations/002_optimize_indexes.sql
CREATE INDEX IF NOT EXISTS idx_articles_published 
ON articles(published_at DESC);

CREATE INDEX IF NOT EXISTS idx_articles_source 
ON articles(source);
```

**Waarom werkt dit?**

- Table moet bestaan VOOR je index maakt
- Volgorde 001 → 002 garandeert dit
- `IF NOT EXISTS` voorkomt errors bij herstart

---

## 🔌 Connecties vanuit de App

### Docker Networking

Wanneer containers in hetzelfde netwerk zitten, kunnen ze elkaar vinden via **hostname**:

```yaml
# docker-compose.yml
networks:
  nieuws-scraper-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

### Hoe verbindt de App?

**In de App container (`cmd/api/main.go`):**

```go
// Environment variables
POSTGRES_HOST=postgres       // ← Hostname van de PostgreSQL container!
POSTGRES_PORT=5432
POSTGRES_USER=scraper
POSTGRES_PASSWORD=scraper_password
POSTGRES_DB=nieuws_scraper

// Connection string
dsn := fmt.Sprintf(
    "postgres://%s:%s@%s:%s/%s?sslmode=disable",
    "scraper",                    // user
    "scraper_password",           // password
    "postgres",                   // ← hostname (niet "localhost"!)
    "5432",                       // port
    "nieuws_scraper",            // database
)

// Maak connection pool
dbPool, err := pgxpool.New(context.Background(), dsn)
```

### DNS Resolutie binnen Docker

```
App Container vraagt: "Wat is het IP van 'postgres'?"
    ↓
Docker Internal DNS
    ↓
Antwoord: "172.20.0.2" (dynamisch toegewezen)
    ↓
App maakt connectie naar 172.20.0.2:5432
    ↓
PostgreSQL Container accepteert connectie
    ↓
Database verbinding actief! ✅
```

### Externe Connectie (vanaf Host)

Als u vanaf uw Windows machine wilt verbinden:

```bash
# Port mapping in docker-compose.yml
ports:
  - "5432:5432"  # Host:Container
```

**Dit betekent:**

```
Windows Host (localhost:5432)
    ↓
Docker port forward
    ↓
Container (postgres:5432)
```

**Verbinden vanaf host:**

```bash
# Via psql
psql -h localhost -p 5432 -U scraper -d nieuws_scraper

# Via connection string
postgres://scraper:scraper_password@localhost:5432/nieuws_scraper

# Via tool (DBeaver, pgAdmin)
Host: localhost
Port: 5432
Database: nieuws_scraper
User: scraper
Password: scraper_password
```

---

## 💾 Backup Mechanisme

### Backup Service

```yaml
backup:
  image: postgres:15-alpine
  container_name: nieuws-scraper-backup
  depends_on:
    postgres:
      condition: service_healthy
  volumes:
    - postgres_data:/var/lib/postgresql/data:ro  # Read-only!
    - ./backups:/backups                          # Output folder
```

### Hoe werkt het?

**1. Backup Script:**

```bash
#!/bin/sh
# Dit draait in de backup container

while true; do
    echo "Starting backup at $(date)"
    
    # Maak backup via pg_dump over netwerk
    pg_dump \
        -h postgres \                    # ← Verbind naar postgres container
        -U scraper \                     # ← Username
        -d nieuws_scraper \             # ← Database
        > /backups/backup_$(date +%Y%m%d_%H%M%S).sql
    
    echo "Backup completed"
    
    # Verwijder oude backups (>7 dagen)
    find /backups -name "backup_*.sql" -type f -mtime +7 -delete
    
    # Wacht 24 uur
    sleep 86400
done
```

**2. Backup Flow:**

```
┌──────────────────────────────────────────────────────────┐
│  BACKUP CONTAINER                                        │
│                                                           │
│  ┌────────────────────────────────────────────┐         │
│  │ pg_dump command                            │         │
│  │                                            │         │
│  │ SELECT * FROM articles;                   │         │
│  │ SELECT * FROM ai_enrichments;             │         │
│  │ SELECT * FROM ...                         │         │
│  └──────────────────┬─────────────────────────┘         │
│                     │                                    │
│                     ↓ SQL queries over network           │
│              ┌──────────────┐                           │
│              │  PostgreSQL  │                           │
│              │  Container   │                           │
│              └──────────────┘                           │
│                     │                                    │
│                     ↓ Returns data                       │
│              ┌──────────────┐                           │
│              │  Writes to   │                           │
│              │  /backups/   │                           │
│              │              │                           │
│              │  backup_     │                           │
│              │  20251029_   │                           │
│              │  120000.sql  │                           │
│              └──────────────┘                           │
│                     │                                    │
│                     ↓ Volume mount                       │
└─────────────────────┼────────────────────────────────────┘
                      │
                      ↓
            ┌──────────────────┐
            │  HOST FILESYSTEM │
            │  ./backups/      │
            │                  │
            │  backup_         │
            │  20251029_       │
            │  120000.sql      │
            └──────────────────┘
```

**3. Backup Format:**

```sql
-- backup_20251029_120000.sql

-- Database schema
CREATE TABLE articles (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    -- ...
);

-- Indexes
CREATE INDEX idx_articles_published ...;

-- Data
INSERT INTO articles (id, title, url, ...) VALUES 
(1, 'Artikel 1', 'https://...', ...),
(2, 'Artikel 2', 'https://...', ...),
-- ... alle rijen
```

### Manual Backup

```bash
# Instant backup vanuit app container
docker-compose exec postgres pg_dump -U scraper nieuws_scraper > backup_manual.sql

# Of vanuit backup container
docker-compose exec backup \
    pg_dump -h postgres -U scraper nieuws_scraper > /backups/backup_manual.sql
```

### Restore

```bash
# Restore vanuit backup
docker-compose exec -T postgres \
    psql -U scraper -d nieuws_scraper < backups/backup_20251029_120000.sql
```

---

## 🛠️ Praktische Voorbeelden

### 1. Database Shell Openen

```bash
# Methode 1: Via docker-compose
docker-compose exec postgres psql -U scraper -d nieuws_scraper

# Methode 2: Via docker
docker exec -it nieuws-scraper-postgres psql -U scraper -d nieuws_scraper

# Je bent nu in de PostgreSQL shell:
nieuws_scraper=#
```

### 2. Queries Uitvoeren

```bash
# Toon alle tables
\dt

# Toon table structuur
\d articles

# Voer query uit
SELECT COUNT(*) FROM articles;

# Voer query uit vanuit host
docker-compose exec postgres \
    psql -U scraper -d nieuws_scraper -c "SELECT COUNT(*) FROM articles;"
```

### 3. Database Resetten

```bash
# Stop alle services
docker-compose down

# Verwijder volume (WIST ALLE DATA!)
docker volume rm nieuws-scraper_postgres_data

# Start opnieuw (verse database)
docker-compose up -d

# Migraties worden automatisch uitgevoerd
```

### 4. Nieuwe Migratie Toevoegen

```bash
# 1. Maak nieuw bestand
# migrations/008_add_new_table.sql

# 2. Voeg SQL toe
echo "CREATE TABLE IF NOT EXISTS new_table (
    id SERIAL PRIMARY KEY,
    name TEXT
);" > migrations/008_add_new_table.sql

# 3. Voer uit (2 opties)

# Optie A: Herstart PostgreSQL (alleen eerste keer)
docker-compose restart postgres

# Optie B: Manual execute (altijd)
docker-compose exec postgres \
    psql -U scraper -d nieuws_scraper < migrations/008_add_new_table.sql
```

### 5. Connection Monitoring

```bash
# Toon actieve connecties
docker-compose exec postgres \
    psql -U scraper -d nieuws_scraper -c "
SELECT 
    pid,
    application_name,
    client_addr,
    state,
    query_start
FROM pg_stat_activity
WHERE datname = 'nieuws_scraper';
"

# Output:
#  pid  | application_name  | client_addr | state  | query_start
# ------+-------------------+-------------+--------+-------------
#  123  | nieuws-scraper-api| 172.20.0.3 | idle   | 2025-10-29...
#  124  | nieuws-scraper-api| 172.20.0.3 | active | 2025-10-29...
```

### 6. Performance Stats

```bash
# Database grootte
docker-compose exec postgres \
    psql -U scraper -d nieuws_scraper -c "
SELECT 
    pg_size_pretty(pg_database_size('nieuws_scraper')) AS db_size;
"

# Table groottes
docker-compose exec postgres \
    psql -U scraper -d nieuws_scraper -c "
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
"
```

### 7. Logs Bekijken

```bash
# Alle PostgreSQL logs
docker-compose logs postgres

# Laatste 100 regels
docker-compose logs --tail=100 postgres

# Follow mode (real-time)
docker-compose logs -f postgres

# Zoeken in logs
docker-compose logs postgres | grep "ERROR"
```

---

## 🔍 Troubleshooting

### Probleem: "Connection refused"

**Symptoom:**
```
Error: connection refused to postgres:5432
```

**Oorzaak & Oplossing:**

```bash
# 1. Check of container draait
docker-compose ps postgres
# STATUS moet "Up (healthy)" zijn

# 2. Check health
docker inspect nieuws-scraper-postgres | grep -A 10 Health

# 3. Check logs
docker-compose logs postgres | tail -50

# 4. Wait for health
# PostgreSQL kan 30-60 sec nodig hebben bij eerste start
```

### Probleem: "Database does not exist"

**Symptoom:**
```
Error: database "nieuws_scraper" does not exist
```

**Oorzaak:**
Database werd niet aangemaakt tijdens initialisatie.

**Oplossing:**

```bash
# Optie 1: Manual create
docker-compose exec postgres \
    psql -U scraper -c "CREATE DATABASE nieuws_scraper;"

# Optie 2: Complete reset
docker-compose down -v
docker-compose up -d
```

### Probleem: "Permission denied"

**Symptoom:**
```
Error: permission denied for table articles
```

**Oorzaak:**
User heeft geen rechten.

**Oplossing:**

```bash
# Grant rechten
docker-compose exec postgres \
    psql -U postgres -d nieuws_scraper -c "
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO scraper;
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO scraper;
"
```

---

## 📊 Best Practices

### DO ✅

1. **Gebruik Named Volumes voor data**
   ```yaml
   volumes:
     - postgres_data:/var/lib/postgresql/data
   ```

2. **Voeg health checks toe**
   ```yaml
   healthcheck:
     test: ["CMD-SHELL", "pg_isready"]
   ```

3. **Gebruik environment variables**
   ```yaml
   environment:
     POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
   ```

4. **Maak regelmatig backups**
   - Automatisch dagelijks
   - Manual voor belangrijke momenten

5. **Monitor connecties**
   ```sql
   SELECT * FROM pg_stat_activity;
   ```

### DON'T ❌

1. **NOOIT docker-compose down -v zonder backup!**
   - Dit wist alle data!

2. **Geen hardcoded passwords**
   ```yaml
   # FOUT
   POSTGRES_PASSWORD: mysecretpassword
   
   # GOED
   POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
   ```

3. **Geen poort expose in production**
   ```yaml
   # Development: OK
   ports:
     - "5432:5432"
   
   # Production: Remove ports!
   ```

4. **Niet handmatig in volume folders werken**
   - Gebruik altijd docker commands

---

## 🎓 Conclusie

**PostgreSQL in Docker = Geïsoleerde, Portable Database**

### Key Takeaways:

1. **Container** = Tijdelijk (kan worden verwijderd)
2. **Volume** = Permanent (bevat alle data)
3. **Network** = Communicatie tussen containers
4. **Migrations** = Auto-execute bij eerste start
5. **Backups** = Dagelijks naar ./backups/

### Lifecycle Samenvatting:

```
docker-compose up
    ↓
1. Pull image (postgres:15-alpine)
2. Create container (nieuws-scraper-postgres)
3. Mount volume (postgres_data)
4. Initialize database (eerste keer)
5. Run migrations (eerste keer)
6. Start PostgreSQL
7. Health check (repeat)
8. App connects via 'postgres' hostname
    ↓
[DATABASE READY] ✅
```

---

**Vragen? Check:**
- [Docker Setup Guide](docker-setup.md)
- [Troubleshooting](operations/troubleshooting.md)
- [Quick Reference](operations/quick-reference.md)

---

*Made with ❤️ for the NieuwsScraper project*