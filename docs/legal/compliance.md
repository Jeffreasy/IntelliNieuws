# 🤖 Robots.txt Compliance & Legal Scraping

## ⚠️ BELANGRIJKE JURIDISCHE NOTITIE

**Nederlandse nieuws sites hebben EXPLICIETE verboden op scraping:**

### DPG Media (ad.nl, nu.nl)
```
# Explicit ban: Not allowed to collect data via scraping or automated methods.
# See www.dpgmedia.nl/gebruiksvoorwaarden
```

**Betekenis:** Scraping is **expliciet verboden** in hun gebruiksvoorwaarden.

### NOS.nl
```
User-agent: *
Disallow: /api
Disallow: /zoeken
# Blocks AI bots like GPTBot, ClaudeBot
```

**Betekenis:** Algemene content mag, maar API's en AI training bots zijn geblokkeerd.

---

## 🚨 Juridische Risico's

### Voor ad.nl en nu.nl (DPG Media)

**Status:** 🔴 **HOOG RISICO**

DPG Media verbiedt expliciet:
- ❌ Web scraping
- ❌ Automated data collection
- ❌ Content aggregatie zonder toestemming

**Mogelijke consequenties:**
- Cease & desist brief
- IP blokkade
- Juridische stappen
- Boetes onder GDPR/Auteursrecht

**Aanbeveling:** 
- ✅ **Gebruik alleen RSS feeds** (expliciet aangeboden voor aggregatie)
- ❌ **NIET full-content scraping** van ad.nl en nu.nl
- ✅ Link naar originele artikelen (traffic naar hun site)

### Voor nos.nl

**Status:** 🟡 **MEDIUM RISICO**

NOS.nl blokkeert:
- ❌ AI training bots (GPTBot, ClaudeBot)
- ❌ API endpoints
- ✅ Algemene content **MAG** (niet expliciet verboden)

**Aanbeveling:**
- ✅ Content scraping **MAG** (niet verboden)
- ✅ Gebruik RSS feed (https://nos.nl/sitemap/index.xml)
- ⚠️ Respecteer rate limits
- ✅ Geen AI training op content

---

## ✅ Legal Scraping Strategy

### Strategie 1: RSS Only (100% Legaal) ⭐ AANBEVOLEN

**Voor ALLE sites:**
```env
ENABLE_FULL_CONTENT_EXTRACTION=false  # Content extraction UIT
ENABLE_BROWSER_SCRAPING=false         # Browser scraping UIT
```

**Gebruik alleen:**
- ✅ RSS feeds (officieel aangeboden)
- ✅ Title + summary (in RSS)
- ✅ Link naar origineel artikel
- ✅ Metadata (datum, auteur)

**Geen juridisch risico!**

### Strategie 2: Selectief Scrapen (Medium Risico)

**Alleen voor NOS.nl:**
```go
// In service.go of extractor
allowedSources := []string{"nos.nl"}

if !contains(allowedSources, article.Source) {
    // Skip scraping for DPG Media sites
    return nil, fmt.Errorf("scraping not allowed for %s", article.Source)
}
```

**Voor ad.nl en nu.nl:**
- ✅ Alleen RSS summary gebruiken
- ❌ Geen full-content extraction

### Strategie 3: Met Toestemming (Ideaal)

**Contact DPG Media:**
- Vraag commerciële licentie
- API toegang onderhandelen
- Partnership deal

**Dit is de ENIGE legale manier voor volledige ad.nl/nu.nl scraping.**

---

## 🛡️ Robots.txt Implementation

### We Hebben Al Robots.txt Checking!

In [`internal/scraper/service.go`](internal/scraper/service.go:82):

```go
// Check robots.txt if enabled
if s.config.EnableRobotsTxtCheck {
    allowed, err := s.robotsChecker.IsAllowed(feedURL)
    if err != nil {
        s.logger.WithError(err).Warnf("Error checking robots.txt")
        // Continue anyway (fallback)
    } else if !allowed {
        result.Error = "robots.txt disallows scraping"
        return result, fmt.Errorf("robots.txt disallows scraping")
    }
}
```

**In [`.env`](.env:38):**
```env
ENABLE_ROBOTS_TXT_CHECK=true  # Al enabled!
```

### Robots.txt Rules

**ad.nl:**
```
Disallow: /*webview
Disallow: /auth
Disallow: /*widget*
Disallow: /*?*otag=
Disallow: /*?*abo_type=
Disallow: /*?*utm_source=
Disallow: /*?*currentArticleId=
```

**nos.nl:**
```
Disallow: /hybrid/
Disallow: /humans.txt
Disallow: /api
Disallow: /zoeken
```

**nu.nl:**
```
# Similar to ad.nl
# Scraping forbidden in terms
```

---

## 📜 Recommended Approach

### Configuration Voor Legal Compliance

```env
# RSS Feeds: Always allowed
ENABLE_RSS_PRIORITY=true
TARGET_SITES=nos.nl  # ALLEEN NOS.nl voor content extraction

# Content Extraction: Only for allowed sites
ENABLE_FULL_CONTENT_EXTRACTION=true
ENABLE_BROWSER_SCRAPING=true

# Respect robots.txt
ENABLE_ROBOTS_TXT_CHECK=true

# Rate limiting (be respectful)
SCRAPER_RATE_LIMIT_SECONDS=5
BROWSER_MAX_CONCURRENT=1
CONTENT_EXTRACTION_DELAY_SECONDS=3
```

### Code-Level Filtering

Update [`service.go`](internal/scraper/service.go) met whitelist:

```go
// In EnrichArticleContent method
func (s *Service) EnrichArticleContent(ctx context.Context, articleID int64) error {
    article, err := s.articleRepo.GetByID(ctx, articleID)
    if err != nil {
        return err
    }

    // LEGAL COMPLIANCE: Only scrape allowed sources
    allowedSources := []string{
        "nos.nl",  // NOS allows general scraping
        // ad.nl and nu.nl EXPLICITLY FORBIDDEN
    }

    allowed := false
    for _, source := range allowedSources {
        if article.Source == source {
            allowed = true
            break
        }
    }

    if !allowed {
        s.logger.Warnf("Content extraction skipped for %s (not in whitelist)", article.Source)
        return fmt.Errorf("content extraction not allowed for source: %s", article.Source)
    }

    // Continue with extraction...
}
```

---

## 🎯 Aanbevolen Setup

### Voor Maximum Legal Veiligheid

**1. Gebruik Alleen RSS Feeds:**
```go
ScrapeSources = map[string]string{
    "nu.nl":  "https://www.nu.nl/rss",        // RSS OK
    "ad.nl":  "https://www.ad.nl/rss.xml",    // RSS OK
    "nos.nl": "https://feeds.nos.nl/nosnieuwsalgemeen", // RSS OK
}
```

**2. Disable Full Content Extraction voor DPG Sites:**
```env
# Alleen NOS.nl scrapen
ALLOWED_CONTENT_SOURCES=nos.nl
```

**3. Altijd Link naar Origineel:**
```typescript
// In frontend
<a href={article.url} target="_blank" rel="noopener">
  Lees volledig artikel op {article.source}
</a>
```

**Dit genereert traffic NAAR de bron → win-win!**

### Voor Development/Testing

Als je MOET testen met ad.nl/nu.nl:

**Voeg disclaimer toe:**
```
"Dit is een persoonlijk/educatief project.
Geen commercieel gebruik.
Content wordt niet gepubliceerd.
Scraping voor research doeleinden only."
```

**En:**
- Gebruik lage rate limits (1 request per 10 sec)
- Kleine test set (<100 artikelen)
- Respecteer robots.txt
- Cache resultaten (scrape niet twee keer)

---

## 📋 Compliance Checklist

### Must-Have (Juridisch Veilig)

- [ ] ✅ Robots.txt checking enabled
- [ ] ✅ RSS feeds als primary source
- [ ] ✅ Rate limiting (min 5 sec tussen requests)
- [ ] ✅ User-agent identificatie (NieuwsScraper/1.0)
- [ ] ✅ Link naar originele artikelen
- [ ] ✅ Respecteer Disallow paths
- [ ] ❌ GEEN commercial data resale
- [ ] ❌ GEEN full-text copying (alleen summaries)

### Nice-to-Have (Extra Respect)

- [ ] ✅ Cache scraped content (no re-scraping)
- [ ] ✅ Scrape only during off-peak hours
- [ ] ✅ Limit to recent articles only (<7 days)
- [ ] ✅ Contact webmaster bij vragen
- [ ] ✅ Prominente source attribution

---

## 🎯 Praktische Implementatie

### RSS-Only Approach (Safest)

```typescript
// Frontend display
<div className="article-card">
  <h3>{article.title}</h3>
  <p className="summary">{article.summary}</p>  {/* Van RSS */}
  <a href={article.url} className="read-more">
    Lees volledig artikel op {article.source} →
  </a>
  <small>Bron: {article.source}</small>
</div>
```

**Juridisch:** ✅ Perfect veilig  
**UX:** ✅ Goed (summary + link)  
**Traffic:** ✅ Naar originele site  

### Hybrid Approach (For NOS.nl Only)

```go
// Backend whitelist
var SCRAPING_WHITELIST = []string{"nos.nl"}

func canScrapeContent(source string) bool {
    for _, allowed := range SCRAPING_WHITELIST {
        if source == allowed {
            return true
        }
    }
    return false
}
```

**Juridisch:** ✅ OK voor NOS.nl  
**Functionaliteit:** ✅ Volledige content voor toegestane sites  
**Risk:** 🟢 Laag  

---

## 📢 Disclaimer Template

**Voor je applicatie (verplicht!):**

```
DISCLAIMER:

Deze applicatie verzamelt nieuwsartikelen via openbaar beschikbare RSS feeds.
- Geen content wordt gekopieerd of opnieuw gepubliceerd
- Alleen metadata (titel, samenvatting) wordt getoond
- Volledige artikelen zijn beschikbaar via links naar de originele bron
- Geen commercieel gebruik van scraped data
- Alle content is eigendom van de respectievelijke uitgevers

Voor vragen over content rechten, neem contact op met de originele bron.

Bronnen:
- NU.nl: https://www.nu.nl/
- AD.nl: https://www.ad.nl/
- NOS.nl: https://nos.nl/
```

---

## 🎊 Aanbeveling

### Voor Productie

**VEILIG & LEGAAL:**
1. ✅ Gebruik **alleen RSS feeds**
2. ✅ Toon summary + link naar origineel
3. ✅ Source attribution prominent
4. ✅ Respecteer robots.txt
5. ✅ Lage rate limits

**EXTRA Features (Optioneel voor NOS.nl):**
6. ✅ Content extraction alleen voor nos.nl
7. ✅ Browser fallback voor JavaScript
8. ✅ Cache om re-scraping te voorkomen

### Voor Development/Research

```env
# Test setup
TARGET_SITES=nos.nl  # Alleen NOS, niet DPG sites
ENABLE_FULL_CONTENT_EXTRACTION=true
ENABLE_BROWSER_SCRAPING=true
SCRAPER_RATE_LIMIT_SECONDS=10  # Extra voorzichtig
```

---

## 📞 Contact Met Uitgevers

### Als Je Commercial License Wilt

**DPG Media (ad.nl, nu.nl):**
- Email: redactie@dpgmedia.nl
- Website: https://www.dpgmedia.nl/contact
- Vraag naar: "Content licensing voor news aggregatie"

**NOS:**
- Email: internet@nos.nl  
- Website: https://nos.nl/contact
- Explain use case: "Non-profit nieuws aggregator"

**Trouw, Volkskrant, Telegraaf:**
- Elk heeft eigen contactpagina
- Commercial licensing mogelijk

---

## ✅ Current Implementation Status

**Wat Je NU Hebt:**
- ✅ Robots.txt checking (enabled in .env)
- ✅ Rate limiting per domain
- ✅ User-agent identification
- ✅ RSS feeds as primary source
- ✅ Optional content extraction (configurable)
- ✅ Browser scraping (met stealth mode)

**Compliance level:** 🟢 **GOED** (als je RSS-only gebruikt)

**Risk level:** 
- RSS only: 🟢 **LAAG** (fully compliant)
- + NOS.nl scraping: 🟡 **MEDIUM** (allowed maar wees voorzichtig)
- + DPG scraping: 🔴 **HOOG** (expliciet verboden)

---

## 🎯 Conclusie

**Voor een legale news aggregator:**

1. **Gebruik RSS feeds** (100% legal, altijd toegestaan)
2. **Toon summaries + links** (geen content copying)
3. **Optioneel: NOS.nl full content** (toegestaan maar respecteer limits)
4. **VERMIJD: DPG Media content scraping** (expliciet verboden)

**Je huidige configuratie is GOED** als je:
- RSS feeds gebruikt ✅
- Robots.txt checked ✅
- Rate limiting enabled ✅
- Content extraction DISABLED voor DPG sites ✅

**Zie [`HEADLESS_BROWSER_GEBRUIKERSGIDS.md`](HEADLESS_BROWSER_GEBRUIKERSGIDS.md) voor technische details.**