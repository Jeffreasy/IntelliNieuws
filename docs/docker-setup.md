# Docker Setup Guide - Production Ready ✨

Deze gids legt uit hoe u de Nieuws Scraper applicatie uitvoert met Docker, inclusief **production-ready** configuraties, security best practices, en automatische backups.

## 📋 Inhoudsopgave

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Configuratie](#configuratie)
- [Production Deployment](#production-deployment)
- [Security](#security)
- [Monitoring & Backups](#monitoring--backups)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop) geïnstalleerd en actief
- Voor Windows: PowerShell
- Minimaal 4GB RAM beschikbaar
- 10GB vrije disk ruimte (voor logs en backups)

---

## Quick Start

### 🚀 Development Setup (Aanbevolen voor lokale ontwikkeling)

1. **Clone het project en navigeer naar de directory:**
   ```powershell
   cd c:\Users\jeffrey\Desktop\Githubmains\NieuwsScraper
   ```

2. **Kopieer en configureer environment variables:**
   ```powershell
   copy .env.example .env
   ```
   
   **⚠️ BELANGRIJK:** Bewerk `.env` en wijzig:
   - `POSTGRES_PASSWORD` - Sterke database password
   - `REDIS_PASSWORD` - Sterke Redis password
   - `EMAIL_USERNAME` en `EMAIL_PASSWORD` - Jouw credentials (niet de voorbeelden!)
   - `OPENAI_API_KEY` - Voor AI features
   - `STOCK_API_KEY` - Voor stock data (FMP)

3. **Start alle services:**
   ```powershell
   # Optie 1: Met PowerShell script
   .\scripts\docker-run.ps1

   # Optie 2: Direct met Docker Compose
   docker-compose up -d

   # Optie 3: Met Make (als geïnstalleerd)
   make docker-run
   ```

4. **Verifieer dat alles draait:**
   ```powershell
   docker-compose ps
   ```

5. **Bekijk logs:**
   ```powershell
   docker-compose logs -f app
   ```

6. **Test de API:**
   ```powershell
   curl http://localhost:8080/health
   ```

✅ **Klaar!** De applicatie is beschikbaar op `http://localhost:8080`

---

## Configuratie

### 🔧 Services Overzicht

De Docker setup bevat **4 services**:

| Service | Port | Beschrijving | Resource Limits |
|---------|------|--------------|-----------------|
| **postgres** | 5432 | PostgreSQL 15 database | 1 CPU, 1GB RAM |
| **redis** | 6379 | Redis 7 cache met persistence | 0.5 CPU, 256MB RAM |
| **app** | 8080 | Go API applicatie | 2 CPU, 2GB RAM |
| **backup** | - | Automatische database backups | 0.5 CPU, 512MB RAM |

### 🔐 Environment Configuratie

De applicatie gebruikt environment variables uit twee bronnen:

1. **`.env` file** (lokaal, niet in git)
2. **`docker-compose.yml`** (defaults)

**Belangrijke variabelen:**

```env
# Database (VERPLICHT: Wijzig passwords!)
POSTGRES_USER=scraper
POSTGRES_PASSWORD=CHANGE_ME_STRONG_PASSWORD_HERE
POSTGRES_DB=nieuws_scraper

# Redis (VERPLICHT: Wijzig password!)
REDIS_PASSWORD=CHANGE_ME_STRONG_REDIS_PASSWORD

# Email (Optioneel)
EMAIL_ENABLED=true
EMAIL_USERNAME=your-email@outlook.com
EMAIL_PASSWORD=YOUR_EMAIL_PASSWORD_HERE

# AI Features (Optioneel)
AI_ENABLED=true
OPENAI_API_KEY=sk-your-key-here

# Stock Data (Optioneel)
STOCK_API_PROVIDER=fmp
STOCK_API_KEY=your-fmp-key-here
```

### 🎯 Development vs Production

**Development** (automatisch actief met `docker-compose up`):
- Gebruikt `docker-compose.override.yml`
- Meer resources (4 CPU, 4GB RAM voor app)
- Verbose logging (DEBUG level)
- Hot reload (volume mount)
- Geen backup service
- Poorten direct exposed

**Production** (gebruik `docker-compose.prod.yml`):
- Strikte resource limits
- Productie logging (INFO level)
- PostgreSQL query optimization
- Redis persistence optimized
- Automatische backups
- Poorten NIET exposed (gebruik reverse proxy)

---

## Production Deployment

### 🏭 Production Setup

1. **Maak een production `.env` file:**
   ```bash
   cp .env.example .env.production
   ```

2. **Configureer STERKE passwords:**
   ```env
   POSTGRES_PASSWORD=$(openssl rand -base64 32)
   REDIS_PASSWORD=$(openssl rand -base64 32)
   EMAIL_PASSWORD=your-secure-password
   ```

3. **Start met production compose:**
   ```bash
   docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
   ```

4. **Verifieer health status:**
   ```bash
   docker-compose ps
   docker-compose logs app | grep "Successfully connected"
   ```

### 🔄 Production Updates

**Zero-downtime deployment:**

```bash
# 1. Pull nieuwe changes
git pull origin main

# 2. Rebuild zonder downtime
docker-compose -f docker-compose.yml -f docker-compose.prod.yml build app

# 3. Rolling update (start nieuwe container voor oude stopt)
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d --no-deps app
```

### 🌐 Reverse Proxy (Aanbevolen voor Production)

Voeg Nginx toe voor HTTPS en load balancing:

```yaml
# In docker-compose.prod.yml (uncomment nginx service)
nginx:
  image: nginx:alpine
  ports:
    - "80:80"
    - "443:443"
  volumes:
    - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
    - ./nginx/certs:/etc/nginx/certs:ro
```

---

## Security

### 🔒 Security Features (✨ NIEUW!)

1. **✅ Credentials niet in Git**
   - `.env` is uitgesloten via `.gitignore`
   - Gebruik environment variables voor alle secrets

2. **✅ Redis Password Protection**
   ```yaml
   # Redis draait nu met authentication
   redis:
     command: redis-server --requirepass ${REDIS_PASSWORD}
   ```

3. **✅ PostgreSQL SSL (Optioneel)**
   ```env
   POSTGRES_SSL_MODE=require  # Voor externe connecties
   ```

4. **✅ Non-root User in Container**
   - App draait als `appuser` (niet root)
   - Defined in `Dockerfile`

5. **✅ Resource Limits**
   - Voorkomt resource exhaustion attacks
   - CPU en memory caps per service

6. **✅ Network Isolation**
   - Custom bridge network (`172.20.0.0/16`)
   - Services kunnen alleen intern communiceren

### 🛡️ Security Checklist

- [ ] Sterke passwords ingesteld in `.env`
- [ ] `.env` NIET in git (check met `git status`)
- [ ] Redis password configured
- [ ] API key authentication enabled
- [ ] Poorten niet direct exposed in production
- [ ] Regular updates: `docker-compose pull`
- [ ] Backup encryptie overwegen voor productie

---

## Monitoring & Backups

### 📊 Health Checks

Alle services hebben health checks:

```bash
# Check service health
docker-compose ps

# Individual service health
docker inspect nieuws-scraper-postgres | grep Health
docker inspect nieuws-scraper-redis | grep Health
docker inspect nieuws-scraper-app | grep Health
```

**Health Endpoints:**

```bash
# Application health
curl http://localhost:8080/health

# Detailed metrics
curl http://localhost:8080/health/metrics

# Liveness probe
curl http://localhost:8080/health/live

# Readiness probe
curl http://localhost:8080/health/ready
```

### 💾 Automatische Backups (✨ NIEUW!)

**Backup Service** draait dagelijks en:
- Maakt PostgreSQL dumps om middernacht
- Bewaart backups in `./backups/` directory
- Verwijdert automatisch backups ouder dan 7 dagen
- Formaat: `backup_YYYYMMDD_HHMMSS.sql`

**Manual backup:**

```bash
# Maak instant backup
docker-compose exec postgres pg_dump -U scraper nieuws_scraper > backup_manual.sql

# Restore from backup
docker-compose exec -T postgres psql -U scraper nieuws_scraper < backup_manual.sql
```

**Backup naar remote storage:**

```bash
# AWS S3 voorbeeld
aws s3 sync ./backups/ s3://your-bucket/backups/ --delete

# Azure Blob voorbeeld
az storage blob upload-batch -d backups -s ./backups/
```

### 📈 Performance Monitoring

**Redis Cache Stats:**

```bash
# Cache info via API
curl http://localhost:8080/api/v1/stocks/stats

# Direct Redis stats
docker-compose exec redis redis-cli --pass ${REDIS_PASSWORD} INFO stats
```

**PostgreSQL Performance:**

```bash
# Database connections
docker-compose exec postgres psql -U scraper -d nieuws_scraper -c "SELECT * FROM pg_stat_activity;"

# Table sizes
docker-compose exec postgres psql -U scraper -d nieuws_scraper -c "SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size FROM pg_tables WHERE schemaname = 'public' ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;"
```

### 📋 Logs Management

```bash
# Tail logs (all services)
docker-compose logs -f

# Specific service
docker-compose logs -f app

# Last 100 lines
docker-compose logs --tail=100 app

# Logs with timestamps
docker-compose logs -f --timestamps app

# Save logs to file
docker-compose logs app > app_logs.txt
```

**Log Rotation** is automatisch geconfigureerd:
- Max size: 10MB per file
- Max files: 3 (laatste 30MB logs)
- Defined in `docker-compose.yml` via logging driver

---

## Troubleshooting

### 🔧 Veelvoorkomende Problemen

#### 1. Services starten niet

**Symptoom:** `docker-compose up` geeft errors

**Oplossingen:**

```bash
# Check Docker Desktop draait
docker info

# Check poorten niet in gebruik
netstat -ano | findstr "5432"  # PostgreSQL
netstat -ano | findstr "6379"  # Redis
netstat -ano | findstr "8080"  # App

# Reset alles en probeer opnieuw
docker-compose down -v
docker-compose up -d
```

#### 2. Database verbinding mislukt

**Symptoom:** App kan niet connecten naar PostgreSQL

**Oplossingen:**

```bash
# Check database health
docker-compose ps postgres

# Check logs
docker-compose logs postgres

# Wacht op health check (kan 30-60 sec duren)
docker-compose logs -f app | grep "Successfully connected"

# Test connectie direct
docker-compose exec postgres psql -U scraper -d nieuws_scraper -c "SELECT 1;"
```

#### 3. Redis verbinding mislukt

**Symptoom:** `Failed to connect to Redis`

**Oplossingen:**

```bash
# Check Redis draait
docker-compose ps redis

# Test Redis met password
docker-compose exec redis redis-cli --pass ${REDIS_PASSWORD} PING
# Verwacht: PONG

# Check Redis logs
docker-compose logs redis
```

#### 4. Out of Memory errors

**Symptoom:** Services crashen, `OOMKilled` in logs

**Oplossingen:**

```bash
# Verhoog Docker memory limit
# Docker Desktop > Settings > Resources > Memory: 8GB

# Check resource gebruik
docker stats

# Restart services
docker-compose restart
```

#### 5. Poort conflicts

**Symptoom:** `port is already allocated`

**Oplossingen:**

```bash
# Wijzig poorten in docker-compose.yml of .env
ports:
  - "5433:5432"  # PostgreSQL op andere poort
  - "6380:6379"  # Redis op andere poort
  - "8081:8080"  # App op andere poort

# Of stop conflicterende services
# Bijvoorbeeld lokale PostgreSQL/Redis
```

#### 6. Volume permission errors

**Symptoom:** Permission denied errors

**Oplossingen:**

```bash
# Windows: Run Docker Desktop as Administrator

# Linux/Mac: Fix permissions
sudo chown -R $USER:$USER ./backups
sudo chown -R $USER:$USER ./migrations
```

### 🧹 Cleanup Commands

```bash
# Stop alle services
docker-compose down

# Stop en verwijder volumes (WIST DATABASE!)
docker-compose down -v

# Verwijder alle containers, networks, images
docker-compose down -v --rmi all --remove-orphans

# Clean Docker system (vrijmaken schijfruimte)
docker system prune -a --volumes

# Rebuild from scratch
docker-compose build --no-cache
docker-compose up -d
```

### 📞 Support & Debugging

**Debug mode inschakelen:**

```env
# In .env
LOG_LEVEL=debug
LOG_FORMAT=text  # Makkelijker leesbaar dan json
```

**Common debugging commands:**

```bash
# Shell in container
docker-compose exec app sh

# Database shell
docker-compose exec postgres psql -U scraper -d nieuws_scraper

# Redis shell
docker-compose exec redis redis-cli --pass ${REDIS_PASSWORD}

# Check environment variables
docker-compose exec app env | grep POSTGRES
```

---

## 🆕 Nieuwe Features in deze Versie

### ✨ Security Verbeteringen
- ✅ Redis password authentication
- ✅ Geen hardcoded credentials
- ✅ Environment variable based configuration
- ✅ Non-root container user

### ✨ Redis Optimalisaties
- ✅ Connection pooling (20 connections, 5 min idle)
- ✅ Persistence configuratie (AOF + RDB)
- ✅ Memory limits (256MB) met LRU eviction
- ✅ Cache invalidation service

### ✨ Docker Optimalisaties
- ✅ Resource limits per service
- ✅ Log rotation (10MB max per file)
- ✅ Health check optimalisaties
- ✅ .dockerignore voor kleinere images
- ✅ Multi-stage build voor efficient images

### ✨ Production Features
- ✅ Automatische database backups (dagelijks)
- ✅ Backup retention (7 dagen)
- ✅ Development vs Production compose files
- ✅ PostgreSQL query optimalisaties
- ✅ Network isolation

---

## 📚 Gerelateerde Documentatie

- [Quick Start Guide](../getting-started/quick-start.md)
- [Installation Guide](../getting-started/installation.md)
- [API Reference](../api/README.md)
- [Troubleshooting Guide](../operations/troubleshooting.md)

---

## ⚙️ Advanced Configuration

### Horizontal Scaling (Docker Swarm)

Voor high-availability deployment:

```bash
# Initialize swarm
docker swarm init

# Deploy stack
docker stack deploy -c docker-compose.yml -c docker-compose.prod.yml nieuws-scraper

# Scale app service
docker service scale nieuws-scraper_app=3
```

### Custom PostgreSQL Tuning

Voeg custom config toe via `docker-compose.yml`:

```yaml
postgres:
  command: >
    postgres
    -c shared_buffers=512MB
    -c effective_cache_size=2GB
    -c maintenance_work_mem=128MB
```

### Redis Cluster Mode

Voor distributie setup:

```yaml
redis:
  image: redis:7-alpine
  command: >
    redis-server
    --cluster-enabled yes
    --cluster-config-file nodes.conf
    --cluster-node-timeout 5000
```

---

**🎉 Klaar voor productie! Veel succes met uw IntelliNieuws deployment.**

Vragen? Check de [Troubleshooting Guide](../operations/troubleshooting.md) of open een [GitHub Issue](https://github.com/yourusername/intellinieuws/issues).