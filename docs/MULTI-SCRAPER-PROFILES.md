# Multi-Scraper Profiles - Design Document

## 🎯 Concept

In plaats van één scraper met vaste settings, kunnen we **4 scraper profiles** hebben met verschillende configuraties:

## 📊 Voorgestelde Profiles

### Profile 1: **Fast & Aggressive** 
**Doel**: Maximale throughput, verse nieuws
```env
# Profile: FAST
SCRAPER_RATE_LIMIT_SECONDS=2          # Zeer agressief
SCRAPER_MAX_CONCURRENT=10             # Maximum parallel
BROWSER_POOL_SIZE=10                  # Grote pool
BROWSER_MAX_CONCURRENT=5              # Veel concurrent
SCRAPER_TIMEOUT_SECONDS=15            # Korte timeout
ENABLE_ROBOTS_TXT_CHECK=false         # Skip voor snelheid
ENABLE_DUPLICATE_DETECTION=true       # Wel deduplication
```
**Use Case**: Breaking news, real-time updates
**Scrape Interval**: 5 minuten

### Profile 2: **Balanced & Efficient** (CURRENT)
**Doel**: Goede balans tussen snelheid en respect
```env
# Profile: BALANCED
SCRAPER_RATE_LIMIT_SECONDS=3
SCRAPER_MAX_CONCURRENT=5
BROWSER_POOL_SIZE=5
BROWSER_MAX_CONCURRENT=3
SCRAPER_TIMEOUT_SECONDS=30
ENABLE_ROBOTS_TXT_CHECK=true
ENABLE_DUPLICATE_DETECTION=true
```
**Use Case**: Normal operations, production default
**Scrape Interval**: 15 minuten

### Profile 3: **Conservative & Respectful**
**Doel**: Minimale server belasting, maximaal respect
```env
# Profile: CONSERVATIVE
SCRAPER_RATE_LIMIT_SECONDS=10         # Zeer respectvol
SCRAPER_MAX_CONCURRENT=2              # Weinig parallel
BROWSER_POOL_SIZE=2                   # Kleine pool
BROWSER_MAX_CONCURRENT=1              # Serieel
SCRAPER_TIMEOUT_SECONDS=60            # Lange timeout
ENABLE_ROBOTS_TXT_CHECK=true
ENABLE_DUPLICATE_DETECTION=true
```
**Use Case**: Bij rate limiting warnings, beperkte resources
**Scrape Interval**: 30 minuten

### Profile 4: **Deep & Thorough**
**Doel**: Maximale content kwaliteit, volledige extractie
```env
# Profile: DEEP
SCRAPER_RATE_LIMIT_SECONDS=5
SCRAPER_MAX_CONCURRENT=3
BROWSER_POOL_SIZE=7                   # Grotere pool voor JS sites
BROWSER_MAX_CONCURRENT=4
BROWSER_TIMEOUT_SECONDS=30            # Meer tijd
BROWSER_WAIT_AFTER_LOAD_MS=3000       # Meer render tijd
ENABLE_FULL_CONTENT_EXTRACTION=true
CONTENT_EXTRACTION_BATCH_SIZE=20
BROWSER_FALLBACK_ONLY=false           # Browser altijd gebruiken
```
**Use Case**: Achtergrond enrichment, kwaliteitsartikelen
**Scrape Interval**: 60 minuten

## 🏗️ Implementatie Opties

### Optie A: **Environment-Based Profiles** (Simpel)

**Deployment**:
```yaml
# docker-compose.yml
services:
  scraper-fast:
    image: nieuws-scraper
    env_file: .env.fast
    
  scraper-balanced:
    image: nieuws-scraper
    env_file: .env.balanced
    
  scraper-conservative:
    image: nieuws-scraper
    env_file: .env.conservative
    
  scraper-deep:
    image: nieuws-scraper
    env_file: .env.deep
```

**Pros**: 
- ✅ Simpel te implementeren
- ✅ Complete isolatie
- ✅ Geen code changes

**Cons**:
- ❌ 4x resources
- ❌ Database conflicts mogelijk
- ❌ Dubbele browser pools

### Optie B: **Single Service, Multiple Schedulers** (Recommended)

**Implementatie**:
```go
// cmd/api/main.go
type ScraperProfile struct {
    Name     string
    Config   *config.ScraperConfig
    Interval time.Duration
}

profiles := []ScraperProfile{
    {
        Name: "fast",
        Config: &config.ScraperConfig{
            RateLimitSeconds: 2,
            MaxConcurrent: 10,
            // ...
        },
        Interval: 5 * time.Minute,
    },
    {
        Name: "balanced",
        Config: &config.ScraperConfig{
            RateLimitSeconds: 3,
            MaxConcurrent: 5,
            // ...
        },
        Interval: 15 * time.Minute,
    },
    // ...
}

// Start scheduler voor elk profile
for _, profile := range profiles {
    scraperService := scraper.NewService(profile.Config, ...)
    scheduler := scheduler.NewScheduler(scraperService, profile.Interval, log)
    go scheduler.Start(ctx)
}
```

**Pros**:
- ✅ Single deployment
- ✅ Shared resources (database, redis)
- ✅ Configureerbaar per profile
- ✅ Dynamisch aan/uit zetten

**Cons**:
- ⚠️ Meer memory usage
- ⚠️ Shared browser pool kan conflict
- ⚠️ Rate limiting per domain nodig

### Optie C: **Dynamic Profile Switching** (Advanced)

**Implementatie**:
```go
type AdaptiveScraper struct {
    profiles map[string]*ScraperProfile
    current  string
}

func (s *AdaptiveScraper) SelectProfile(ctx context.Context) string {
    // Check system load
    cpuUsage := getSystemCPU()
    queueSize := getQueueSize()
    errorRate := getErrorRate()
    
    switch {
    case errorRate > 0.5:
        return "conservative" // Te veel errors, rustig aan
    case queueSize > 100:
        return "fast" // Grote achterstand, gas geven
    case cpuUsage < 30:
        return "deep" // Resources over, kwaliteit verhogen
    default:
        return "balanced" // Normal operations
    }
}
```

**Pros**:
- ✅ Fully automatic
- ✅ Adapts to conditions
- ✅ Optimal resource usage

**Cons**:
- ❌ Complex implementation
- ❌ Harder to debug
- ❌ Needs extensive testing

## 🚀 **Aanbevolen Aanpak: Hybrid**

**Fase 1 (Nu)**: Environment-based profiles
- 2 instances: Fast (5 min) + Balanced (15 min)
- Minimale code changes
- Production-ready

**Fase 2 (Later)**: Multiple schedulers
- Single deployment, 4 schedulers
- Shared resources
- Configureerbaar

**Fase 3 (Future)**: Dynamic switching
- Auto-adapt aan omstandigheden
- ML-based optimization
- Fully autonomous

## 📝 **Implementatie Stappen (Optie B)**

### 1. Nieuwe Config Structure
```go
// pkg/config/scraper_profiles.go
type ScraperProfile struct {
    Name              string
    RateLimitSeconds  int
    MaxConcurrent     int
    BrowserPoolSize   int
    Interval          time.Duration
    TargetSites       []string
}

var Profiles = map[string]ScraperProfile{
    "fast": {
        Name: "fast",
        RateLimitSeconds: 2,
        MaxConcurrent: 10,
        Interval: 5 * time.Minute,
        TargetSites: []string{"nu.nl"},  // Breaking news only
    },
    "balanced": {
        Name: "balanced",
        RateLimitSeconds: 3,
        MaxConcurrent: 5,
        Interval: 15 * time.Minute,
        TargetSites: []string{"nu.nl", "ad.nl", "nos.nl"},
    },
    "deep": {
        Name: "deep",
        RateLimitSeconds: 5,
        MaxConcurrent: 3,
        BrowserPoolSize: 7,
        Interval: 60 * time.Minute,
        TargetSites: []string{"nu.nl", "ad.nl", "nos.nl", "trouw.nl"},
    },
}
```

### 2. Modified Main.go
```go
// Initialize multiple scrapers from profiles
enabledProfiles := strings.Split(cfg.ScraperProfiles, ",")
for _, profileName := range enabledProfiles {
    profile := config.Profiles[profileName]
    
    // Create scraper with profile config
    scraperService := scraper.NewService(&profile, articleRepo, jobRepo, log)
    
    // Start scheduler
    scheduler := scheduler.NewScheduler(scraperService, profile.Interval, log)
    go scheduler.Start(ctx)
    
    log.Infof("Started scraper profile '%s' with interval %v", profileName, profile.Interval)
}
```

### 3. Environment Configuration
```env
# Enable multiple scraper profiles
SCRAPER_PROFILES=fast,balanced,deep
```

## 🎯 **Use Cases**

### Use Case 1: **24/7 News Coverage**
```
Fast (5 min):     Breaking news van nu.nl
Balanced (15 min): Algemeen nieuws van alle bronnen
Deep (60 min):     Uitgebreide artikelen met volledige content
```

### Use Case 2: **Resource Optimization**
```
Dag (8-22u):    Fast + Balanced + Deep (max coverage)
Nacht (22-8u):  Balanced only (resource saving)
Weekend:        Conservative (respect for servers)
```

### Use Case 3: **Source-Specific**
```
Profile A: nu.nl  (5 min, agressief)
Profile B: ad.nl  (15 min, balanced)
Profile C: nos.nl (30 min, deep analysis)
Profile D: All    (60 min, archival)
```

## 💡 **Voordelen Multi-Profile Approach**

1. **Flexibility**: Verschillende strategieën voor verschillende behoeften
2. **Optimization**: Elke profile geoptimaliseerd voor specifiek doel
3. **Scalability**: Makkelijk nieuwe profiles toevoegen
4. **Resource Management**: Betere load balancing
5. **Cost Control**: Deep profiles alleen wanneer nodig

## ⚠️ **Considerations**

1. **Database Load**: 4 scrapers = 4x writes
   - Oplossing: Shared duplicate detection
   - Oplossing: Database connection pooling

2. **Rate Limiting**: Meerdere scrapers naar zelfde domain
   - Oplossing: Shared rate limiter per domain
   - Oplossing: Profile-aware delays

3. **Memory Usage**: Meerdere browser pools
   - Oplossing: Shared browser pool
   - Oplossing: Dynamic pool sizing

4. **Complexity**: Meerdere schedulers beheren
   - Oplossing: Central orchestrator
   - Oplossing: Profile management API

## 🚀 **Quick Start (Optie A - Docker Compose)**

```yaml
# docker-compose.profiles.yml
version: '3.8'
services:
  scraper-fast:
    extends:
      file: docker-compose.yml
      service: app
    environment:
      - SCRAPER_PROFILE=fast
      - SCRAPER_RATE_LIMIT_SECONDS=2
      - SCRAPER_MAX_CONCURRENT=10
      - SCRAPER_SCHEDULE_INTERVAL_MINUTES=5
    container_name: scraper-fast
    
  scraper-balanced:
    extends:
      file: docker-compose.yml
      service: app
    environment:
      - SCRAPER_PROFILE=balanced
      - SCRAPER_RATE_LIMIT_SECONDS=3
      - SCRAPER_MAX_CONCURRENT=5
      - SCRAPER_SCHEDULE_INTERVAL_MINUTES=15
    container_name: scraper-balanced
```

**Deploy**:
```bash
docker-compose -f docker-compose.yml -f docker-compose.profiles.yml up -d
```

## 📊 **Expected Results**

Met 4 profiles (Fast, Balanced, Deep, Conservative):

**Coverage**:
- Fast: 12 scrapes/uur × 30 artikelen = 360 artikelen/uur
- Balanced: 4 scrapes/uur × 80 artikelen = 320 artikelen/uur
- Deep: 1 scrape/uur × 100 artikelen = 100 artikelen/uur
- **Totaal**: ~780 artikelen/uur, ~18,000/dag

**Resource Usage**:
- Database: Shared pool (25 connections)
- Redis: Shared pool (30 connections)
- Browser: Shared or per-profile
- CPU: 2-4 cores recommended

## 🎉 Conclusie

Multi-profile scraping is **zeer effectief** voor:
- ✅ Verschillende prioriteiten per bron
- ✅ Optimale resource allocatie
- ✅ Betere coverage
- ✅ Flexible deployment

**Aanbeveling**: Start met Optie A (Docker Compose) voor snelle implementatie, migreer later naar Optie B voor betere control.