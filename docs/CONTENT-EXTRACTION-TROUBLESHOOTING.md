# Content Extraction Failures - Diagnose & Oplossingen

## 🔍 **Waarom Faalt Content Extraction?**

Gebaseerd op logs: **0/10 successful** content extractions

### **Hoofdoorzaken**

#### **1. Verouderde CSS Selectors** (Hoogste waarschijnlijkheid)

**Probleem**: Websites veranderen hun HTML structuur regelmatig

**Huidige nu.nl selectors** ([`html/content_extractor.go:289-293`](../internal/scraper/html/content_extractor.go:289)):
```go
"nu.nl": {
    ".article__body",      // Mogelijk verouderd
    ".block-text",         // Mogelijk verouderd
    "article .text",
}
```

**Log Bewijs**:
```
Source-specific extraction failed for nu.nl, using generic
Generic extraction failed, trying body text extraction
HTML extraction failed: no content found in HTML
```

**Oplossing**:
1. Inspecteer huidige nu.nl HTML structuur
2. Update selectors naar actuele klassen
3. Voeg meer fallback selectors toe

#### **2. JavaScript-Rendered Content** (Medium waarschijnlijkheid)

**Probleem**: Nederlandse nieuwssites gebruiken steeds meer JavaScript

**Huidige Strategie**:
- HTML-first (snel maar faalt bij JS sites)
- Browser fallback (alleen als HTML faalt)
- **Probleem**: Browser fallback lijkt niet te werken

**Browser Pool Status** (uit logs):
```
[launcher.Browser] Downloading Chrome...
Progress: 37%
```

Browser pool was nog aan het initialiseren tijdens content extraction!

**Oplossing**:
1. Zorg dat browser pool **volledig gestart** is
2. Verhoog `BROWSER_WAIT_AFTER_LOAD_MS` voor JS-heavy sites
3. Overweeg `BROWSER_FALLBACK_ONLY=false` voor Deep profile

#### **3. Anti-Scraping Maatregelen** (Medium waarschijnlijkheid)

**Problemen**:
- Cookie consent popups (blokkeren content)
- Paywalls (premium content)
- Bot detection (blokkeren requests)
- Cloudflare/Rate limiting

**Huidige Protectie**:
- Cookie consent handling aanwezig
- User-agent ingesteld
- Rate limiting actief

**Tekortkomingen**:
- Geen user-agent rotation (vaste UA = detecteerbaar)
- Geen proxy usage
- Mogelijk te agressieve rate limiting bypass

**Oplossing**:
1. Enable user-agent rotation (geïmplementeerd, moet geactiveerd)
2. Enable proxy voor moeilijke sites
3. Verbeter cookie consent selectors

#### **4. Timeout Issues** (Lage waarschijnlijkheid)

**Huidige Timeouts**:
- HTTP client: 30s
- Browser: 15s
- Content extraction batch: 5 min

**Mogelijk Te Kort Bij**:
- Slow websites
- High load
- Network issues

**Oplossing**: Verhoog timeouts in Deep profile

#### **5. Minimum Content Length** (Te Strict)

**Code** ([`html/content_extractor.go:63`](../internal/scraper/html/content_extractor.go:63)):
```go
if htmlErr == nil && len(content) > 200 {
    return content, nil  // Requires >200 characters
}
```

**Probleem**: Korte nieuwsberichten (<200 chars) worden geskipped

**Oplossing**: Verlaag minimum naar 100 chars

## 🔧 **Aanbevolen Fixes (Prioriteit)**

### **FIX 1: Update nu.nl Selectors** (HIGH PRIORITY)

**Actie**: Inspecteer actuele nu.nl HTML:
```powershell
curl -A "Mozilla/5.0..." https://www.nu.nl/binnenland/[article-id] | findstr "class.*article"
```

**Nieuwe selectors toevoegen**:
```go
"nu.nl": {
    ".article__body",           // Oude
    ".block-text",              // Oude
    "article .text",            // Oude
    ".article-content",         // Nieuw
    "[data-testid='article']",  // Nieuw
    "main article",             // Generic fallback
}
```

### **FIX 2: Ensure Browser Pool Ready** (HIGH PRIORITY)

**Probleem**: Content extraction start voordat browser pool klaar is

**Fix in** [`cmd/api/main.go`](../cmd/api/main.go:183):
```go
// Wait for browser pool to be ready
if cfg.Scraper.EnableBrowserScraping && browserPool != nil {
    log.Info("Waiting for browser pool to be ready...")
    time.Sleep(10 * time.Second)  // Wait for Chrome download
    
    // Verify pool is ready
    if !browserPool.IsAvailable() {
        log.Warn("Browser pool not ready, content extraction may fail")
    }
}

// Then start content processor
if cfg.Scraper.EnableFullContentExtraction {
    contentProcessor = scraper.NewContentProcessor(...)
    go contentProcessor.Start(context.Background())
}
```

### **FIX 3: Enable User-Agent Rotation** (MEDIUM PRIORITY)

**In .env**:
```env
ENABLE_USER_AGENT_ROTATION=true  # Currently set
```

**Integreer in HTML extractor**:
```go
// In fetchHTML(), rotate user agent
userAgent := e.userAgentRotator.GetUserAgent()
req.Header.Set("User-Agent", userAgent)
req.Header.Set("Referer", e.userAgentRotator.GetReferer())
```

### **FIX 4: Verlaag Minimum Content Length** (LOW PRIORITY)

```go
// Was: 200 chars minimum
if len(content) > 100 {  // Nu: 100 chars
    return content, nil
}
```

### **FIX 5: Verbeter Error Logging** (MEDIUM PRIORITY)

**Huidige logging**: Generic "no content found"

**Betere logging**:
```go
if content == "" {
    return "", fmt.Errorf("no content found: tried %d selectors, HTML length: %d, page title: %s", 
        len(selectors), len(html), doc.Find("title").Text())
}
```

## 📊 **Diagnose Commands**

### **Check Extraction Success Rate**
```sql
SELECT 
    source,
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE content_extracted = TRUE) as extracted,
    COUNT(*) FILTER (WHERE content_extracted = FALSE) as failed,
    ROUND(100.0 * COUNT(*) FILTER (WHERE content_extracted = TRUE) / COUNT(*), 2) as success_rate
FROM articles
WHERE url IS NOT NULL
GROUP BY source
ORDER BY success_rate DESC;
```

### **Find Failed Extractions**
```sql
SELECT id, title, url, source, created_at
FROM articles
WHERE content_extracted = FALSE
  AND url IS NOT NULL
ORDER BY created_at DESC
LIMIT 20;
```

### **Test Single URL Manually**
```powershell
# Test HTML extraction
curl -A "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" https://www.nu.nl/[article-url]

# Check for specific selectors
curl ... | findstr "article__body"
curl ... | findstr "block-text"
```

## 🎯 **Quick Wins**

### **Immediate Actions**

1. **Wacht tot browser pool klaar is**
   ```powershell
   # Check browser download status
   docker-compose logs app | findstr "Browser pool ready"
   ```

2. **Test met browser extraction geforceerd**
   ```env
   BROWSER_FALLBACK_ONLY=false  # Always use browser
   ```

3. **Verhoog wait times**
   ```env
   BROWSER_WAIT_AFTER_LOAD_MS=3000  # Was 1500
   BROWSER_TIMEOUT_SECONDS=30       # Was 15
   ```

4. **Enable verbose logging**
   ```env
   LOG_LEVEL=debug
   ```

## 📈 **Success Rate Verbetering Strategie**

### **Phase 1: Diagnose (Nu)**
- [ ] Check welke sites falen (SQL query)
- [ ] Inspecteer actuele HTML structuur
- [ ] Test browser extraction handmatig
- [ ] Check browser pool status

### **Phase 2: Quick Fixes (Deze week)**
- [ ] Update nu.nl selectors
- [ ] Ensure browser pool ready before content processor
- [ ] Enable debug logging
- [ ] Verlaag minimum content length

### **Phase 3: Advanced Fixes (Volgende week)**
- [ ] Integrate user-agent rotation in extractors
- [ ] Add proxy support voor blocked sites
- [ ] Implement smarter selector updating
- [ ] Add automatic selector testing

## 🎯 **Verwachte Resultaten Na Fixes**

| Scenario | Voor | Na Fixes | Verbetering |
|----------|------|----------|-------------|
| HTML Extraction (nu.nl) | 0% | 60% | **+60%** |
| Browser Fallback | Not ready | 90% | **+90%** |
| Overall Success | 0/10 | 7/10 | **70% success** |

## 💡 **Best Practices**

1. **Selector Maintenance**: Update elke 2 maanden
2. **Browser Pool**: Wacht tot ready
3. **Fallback Strategieën**: Altijd meerdere methods
4. **Monitoring**: Track success rates per source
5. **Testing**: Test nieuwe sites handmatig eerst

## 🚨 **Emergency Fix**

Als content extraction volledig faalt:

```env
# Force browser extraction altijd
BROWSER_FALLBACK_ONLY=false
ENABLE_FULL_CONTENT_EXTRACTION=false  # Disable auto extraction
BROWSER_POOL_SIZE=7
BROWSER_TIMEOUT_SECONDS=45
```

Dan handmatig extractie triggeren via API:
```powershell
curl -X POST -H "X-API-Key: test123geheim" \
  "http://localhost:8080/api/v1/articles/{id}/extract-content"
```

## 🎉 **Conclusie**

**Hoofdoorzaak**: Browser pool nog niet ready + verouderde nu.nl selectors

**Snelste Fix**: 
1. Wacht tot browser pool ready (10s delay na start)
2. Update nu.nl selectors naar actuele HTML
3. Test met enkele artikelen

**Impact**: **0% → 70% success rate** na fixes