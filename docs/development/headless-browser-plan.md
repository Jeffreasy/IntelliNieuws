# Headless Browser Scraping Implementatie Plan

## 🎯 Doelstelling

**JavaScript-rendered content scrapen** door een echte browser te gebruiken.

**Use case:** Artikelen die falen met normale HTML scraping (zoals artikel 182)

---

## 🔧 Technologie Keuze: Rod vs Chromedp

### Optie 1: Rod ⭐ AANBEVOLEN

**Library:** `github.com/go-rod/rod`

**Voordelen:**
- ✅ **Makkelijker API** - High-level, intuïtief
- ✅ **Betere error handling** - Automatic retries
- ✅ **Built-in pool management** - Browser instances hergebruiken
- ✅ **Actieve development** - Regelmatige updates
- ✅ **Goede documentatie** - Veel voorbeelden
- ✅ **DevTools protocol** - Debugging support

**Nadelen:**
- Iets meer overhead dan Chromedp
- Grotere dependency tree

### Optie 2: Chromedp

**Library:** `github.com/chromedp/chromedp`

**Voordelen:**
- Lagere footprint
- Meer controle over details

**Nadelen:**
- ❌ Lower-level API (complexer)
- ❌ Meer boilerplate code
- ❌ Moeilijker resource management

**AANBEVELING: Gebruik Rod** - veel makkelijker en beter getest.

---

## 📦 Architectuur Overview

### Triple-Layer Fallback Strategie

```
1. HTML Scraping (Snel - 1-2 sec)
   └─> Success? → Return content ✅
        └─> Fail? ↓

2. Browser Scraping (Langzaam - 5-10 sec)  
   └─> Success? → Return content ✅
        └─> Fail? ↓

3. RSS Summary (Altijd beschikbaar)
   └─> Return summary ✅ (Always works)
```

### Browser Pool Architecture

```
┌─────────────────────────────────────────┐
│       Browser Pool Manager              │
│  - 3-5 browser instances (reusable)     │
│  - Connection pooling                   │
│  - Automatic cleanup                    │
│  - Resource limits                      │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│     Content Extraction Request          │
│  1. Try HTML scraping first             │
│  2. If fail → Try browser scraping      │
│  3. If fail → Use RSS summary           │
└─────────────────────────────────────────┘
```

---

## 🏗️ Implementatie Plan

### Fase 1: Dependencies & Setup

**Dependencies installeren:**
```bash
go get github.com/go-rod/rod@latest
```

**Benodigdheden:**
- Chrome/Chromium binary (automatisch gedownload door Rod)
- Of Docker image met Chrome

### Fase 2: Browser Pool Manager

**Nieuwe file:** `internal/scraper/browser/pool.go`

```go
package browser

import (
    "context"
    "sync"
    "time"

    "github.com/go-rod/rod"
    "github.com/go-rod/rod/lib/launcher"
    "github.com/go-rod/rod/lib/proto"
    "github.com/jeffrey/nieuws-scraper/pkg/logger"
)

// BrowserPool manages reusable browser instances
type BrowserPool struct {
    browsers []*rod.Browser
    mu       sync.Mutex
    size     int
    launcher *launcher.Launcher
    logger   *logger.Logger
}

// NewBrowserPool creates a browser pool
func NewBrowserPool(size int, log *logger.Logger) (*BrowserPool, error) {
    // Launch Chrome
    l := launcher.New().
        Headless(true).
        NoSandbox(true).
        Set("disable-web-security", "true")

    url := l.MustLaunch()

    pool := &BrowserPool{
        browsers: make([]*rod.Browser, 0, size),
        size:     size,
        launcher: l,
        logger:   log.WithComponent("browser-pool"),
    }

    // Pre-create browser instances
    for i := 0; i < size; i++ {
        browser := rod.New().ControlURL(url).MustConnect()
        pool.browsers = append(pool.browsers, browser)
    }

    log.Infof("Browser pool initialized with %d instances", size)
    return pool, nil
}

// Acquire gets a browser from pool
func (p *BrowserPool) Acquire(ctx context.Context) (*rod.Browser, error) {
    p.mu.Lock()
    defer p.mu.Unlock()

    if len(p.browsers) == 0 {
        return nil, fmt.Errorf("no browsers available")
    }

    browser := p.browsers[0]
    p.browsers = p.browsers[1:]
    
    return browser, nil
}

// Release returns browser to pool
func (p *BrowserPool) Release(browser *rod.Browser) {
    p.mu.Lock()
    defer p.mu.Unlock()

    p.browsers = append(p.browsers, browser)
}

// Close closes all browsers
func (p *BrowserPool) Close() {
    p.mu.Lock()
    defer p.mu.Unlock()

    for _, browser := range p.browsers {
        browser.MustClose()
    }
    p.launcher.Cleanup()
}
```

### Fase 3: Browser Content Extractor

**Nieuwe file:** `internal/scraper/browser/browser_extractor.go`

```go
package browser

import (
    "context"
    "fmt"
    "strings"
    "time"

    "github.com/go-rod/rod"
    "github.com/go-rod/rod/lib/proto"
    "github.com/jeffrey/nieuws-scraper/pkg/logger"
)

// BrowserExtractor extracts content using headless Chrome
type BrowserExtractor struct {
    pool    *BrowserPool
    logger  *logger.Logger
    timeout time.Duration
}

// NewBrowserExtractor creates extractor
func NewBrowserExtractor(pool *BrowserPool, timeout time.Duration, log *logger.Logger) *BrowserExtractor {
    return &BrowserExtractor{
        pool:    pool,
        logger:  log.WithComponent("browser-extractor"),
        timeout: timeout,
    }
}

// ExtractContent extracts content using browser
func (e *BrowserExtractor) ExtractContent(ctx context.Context, url string, source string) (string, error) {
    e.logger.Infof("Extracting content with browser from %s", url)
    startTime := time.Now()

    // Acquire browser from pool
    browser, err := e.pool.Acquire(ctx)
    if err != nil {
        return "", fmt.Errorf("failed to acquire browser: %w", err)
    }
    defer e.pool.Release(browser)

    // Create page with timeout
    page, err := browser.Timeout(e.timeout).Page(proto.TargetCreateTarget{URL: url})
    if err != nil {
        return "", fmt.Errorf("failed to create page: %w", err)
    }
    defer page.Close()

    // Wait for page to load
    if err := page.WaitLoad(); err != nil {
        return "", fmt.Errorf("page load timeout: %w", err)
    }

    // Additional wait for JavaScript rendering (1-2 seconds)
    time.Sleep(2 * time.Second)

    // Try site-specific selectors first
    content, err := e.extractBySelector(page, source)
    if err == nil && content != "" {
        duration := time.Since(startTime)
        e.logger.Infof("Browser extracted %d characters in %v", len(content), duration)
        return content, nil
    }

    // Fallback: get all article text
    content, err = e.extractGeneric(page)
    if err != nil {
        return "", err
    }

    duration := time.Since(startTime)
    e.logger.Infof("Browser extracted %d characters (generic) in %v", len(content), duration)
    return content, nil
}

// extractBySelector tries site-specific selectors
func (e *BrowserExtractor) extractBySelector(page *rod.Page, source string) (string, error) {
    selectors := map[string][]string{
        "nu.nl":  {".article__body", "article"},
        "ad.nl":  {".article__body", "article"},
        "nos.nl": {".article-content", "article"},
    }

    if sels, ok := selectors[source]; ok {
        for _, sel := range sels {
            element, err := page.Element(sel)
            if err == nil {
                text, _ := element.Text()
                if len(text) > 200 {
                    return cleanText(text), nil
                }
            }
        }
    }

    return "", fmt.Errorf("no selector matched")
}

// extractGeneric extracts using common patterns
func (e *BrowserExtractor) extractGeneric(page *rod.Page) (string, error) {
    // Try to find article element
    selectors := []string{"article", "main", "[role='main']"}
    
    for _, sel := range selectors {
        element, err := page.Element(sel)
        if err == nil {
            text, _ := element.Text()
            if len(text) > 200 {
                return cleanText(text), nil
            }
        }
    }

    // Last resort: get all paragraph text
    elements, err := page.Elements("p")
    if err != nil {
        return "", err
    }

    var paragraphs []string
    for _, el := range elements {
        text, _ := el.Text()
        if len(text) > 50 {
            paragraphs = append(paragraphs, text)
        }
    }

    if len(paragraphs) > 0 {
        return strings.Join(paragraphs, "\n\n"), nil
    }

    return "", fmt.Errorf("no content found")
}

func cleanText(text string) string {
    text = strings.Join(strings.Fields(text), " ")
    return strings.TrimSpace(text)
}
```

### Fase 4: Intelligent Fallback in Content Extractor

**Update bestaande:** `internal/scraper/html/content_extractor.go`

```go
// In struct, add browser extractor
type ContentExtractor struct {
    client          *http.Client
    sanitizer       *bluemonday.Policy
    browserExtractor *browser.BrowserExtractor  // NEW
    useBrowser      bool
    logger          *logger.Logger
    userAgent       string
}

// ExtractContent with fallback to browser
func (e *ContentExtractor) ExtractContent(ctx context.Context, url string, source string) (string, error) {
    // Try HTML scraping first (FAST)
    content, err := e.extractHTML(ctx, url, source)
    if err == nil && content != "" {
        e.logger.Infof("HTML extraction successful (%d chars)", len(content))
        return content, nil
    }

    // Fallback to browser if enabled and HTML failed
    if e.useBrowser && e.browserExtractor != nil {
        e.logger.Warnf("HTML extraction failed, trying browser for %s", url)
        content, err = e.browserExtractor.ExtractContent(ctx, url, source)
        if err == nil {
            e.logger.Infof("Browser extraction successful (%d chars)", len(content))
            return content, nil
        }
        e.logger.WithError(err).Warnf("Browser extraction also failed for %s", url)
    }

    return "", fmt.Errorf("all extraction methods failed")
}
```

---

## ⚙️ Configuration

### .env Settings

```env
# Headless Browser Scraping
ENABLE_BROWSER_SCRAPING=false        # Enable headless browser
BROWSER_POOL_SIZE=3                  # How many browser instances
BROWSER_TIMEOUT_SECONDS=15           # Max time per page
BROWSER_WAIT_AFTER_LOAD_MS=2000      # Wait for JS rendering
BROWSER_FALLBACK_ONLY=true           # Only use if HTML fails (RECOMMENDED)
BROWSER_MAX_CONCURRENT=2             # Max concurrent browser extractions
```

### Config Struct Update

```go
type ScraperConfig struct {
    // ... existing fields
    
    // Browser scraping
    EnableBrowserScraping    bool
    BrowserPoolSize          int
    BrowserTimeout           time.Duration
    BrowserWaitAfterLoad     time.Duration
    BrowserFallbackOnly      bool
    BrowserMaxConcurrent     int
}
```

---

## 💾 Resource Management

### Memory Implications

**Per browser instance:**
- RAM: ~50-100 MB
- CPU: 10-20% tijdens scraping
- Disk: ~500 MB (Chrome binary)

**Voor 3 browser instances:**
- RAM: ~200-300 MB
- CPU: Spikes tijdens gebruik
- Disk: ~500 MB (één Chrome binary gedeeld)

### Optimization Strategies

1. **Lazy Initialization** - Start browsers alleen als nodig
2. **Pool Reuse** - Hergebruik browser instances
3. **Selective Use** - Alleen voor JavaScript sites
4. **Timeout Management** - Force kill na timeout
5. **Memory Cleanup** - Periodic browser restart

---

## 🎭 Use Cases

### Scenario 1: HTML First, Browser Fallback ⭐ AANBEVOLEN

```
Request → HTML Scraping (1-2 sec)
            ├─ Success → Return (70-80% van gevallen)
            └─ Fail → Browser Scraping (5-10 sec)
                        ├─ Success → Return (15-20% van gevallen)
                        └─ Fail → RSS Summary (5-10% van gevallen)
```

**Voordeel:** Meeste artikelen blijven snel, langzame browser alleen als nodig

### Scenario 2: Browser Only for Specific Sites

```go
javascriptSites := []string{"dynamic-news.nl", "spa-site.nl"}

if contains(javascriptSites, article.Source) {
    // Skip HTML, go direct to browser
    content = browserExtractor.Extract(url)
} else {
    // Normal HTML scraping
    content = htmlExtractor.Extract(url)
}
```

### Scenario 3: User-Triggered Browser Scraping

```
User clicks "Haal volledige tekst op"
  └─> Try HTML first
       └─> If fails, show: "Probeer met browser? (kan 10 sec duren)"
            └─> User confirms
                 └─> Try browser scraping
```

---

## 📊 Performance Impact

### Speed Comparison

| Method | Avg Time | Success Rate | Use Case |
|--------|----------|--------------|----------|
| RSS | 100-200ms | 100% | Metadata always |
| HTML Scraping | 1-2 sec | 70-80% | Static sites |
| Browser Scraping | 5-10 sec | 90-95% | JavaScript sites |
| Combined (HTML → Browser) | 2-3 sec avg | 95%+ | Best of both! |

### Cost Analysis

**Throughput:**
- **HTML Only:** 30-50 articles/minute
- **Browser Only:** 6-10 articles/minute  
- **Hybrid (HTML first):** 25-40 articles/minute

**Resources:**
- **HTML:** 10MB RAM, 5% CPU
- **Browser:** 300MB RAM, 50% CPU (during scraping)
- **Hybrid:** 150MB RAM average, 15-20% CPU average

---

## 🛠️ Implementation Steps

### Step 1: Install Rod

```bash
go get github.com/go-rod/rod@latest
```

### Step 2: Create Browser Pool

Create `internal/scraper/browser/pool.go` (zie code voorbeeld hierboven)

### Step 3: Create Browser Extractor

Create `internal/scraper/browser/browser_extractor.go`

### Step 4: Integrate with Content Extractor

Update `internal/scraper/html/content_extractor.go` to use browser als fallback

### Step 5: Update Service

```go
// In service.go
type Service struct {
    // ... existing
    browserPool *browser.BrowserPool
}

func NewService(...) *Service {
    var browserPool *browser.BrowserPool
    if cfg.EnableBrowserScraping {
        browserPool, _ = browser.NewBrowserPool(cfg.BrowserPoolSize, log)
    }

    contentExtractor := html.NewContentExtractor(cfg.UserAgent, browserPool, log)
    
    return &Service{
        // ... existing
        browserPool: browserPool,
    }
}
```

### Step 6: Graceful Shutdown

```go
// In main.go shutdown
if browserPool != nil {
    log.Info("Closing browser pool...")
    browserPool.Close()
}
```

---

## ⚠️ Challenges & Solutions

### Challenge 1: Chrome Binary

**Problem:** Chrome moet geïnstalleerd zijn

**Solutions:**
- **Auto-download:** Rod download automatisch Chrome
- **Docker:** Gebruik image met Chrome: `ghcr.io/go-rod/rod`
- **System Chrome:** Gebruik installed Chrome/Chromium

### Challenge 2: Resource Limits

**Problem:** Teveel browser instances = crash

**Solution:**
```go
// Limit concurrent browser usage
semaphore := make(chan struct{}, 2)  // Max 2 concurrent

func extractWithBrowser(url string) {
    semaphore <- struct{}{}
    defer func() { <-semaphore }()
    
    // Extract...
}
```

### Challenge 3: Memory Leaks

**Problem:** Browsers niet proper afgesloten

**Solution:**
```go
// Always defer close
page := browser.MustPage(url)
defer page.Close()

// Periodic browser restart
if usageCount > 100 {
    browser.MustClose()
    browser = pool.GetFresh()
}
```

### Challenge 4: Timeouts

**Problem:** Sommige pagina's laden nooit

**Solution:**
```go
page := browser.Timeout(15 * time.Second).MustPage(url)

// With context
ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
defer cancel()
```

---

## 🎯 Aanbevolen Configuratie

### Development (Lokaal Testen)

```env
ENABLE_BROWSER_SCRAPING=true
BROWSER_POOL_SIZE=2              # Klein voor development
BROWSER_TIMEOUT_SECONDS=10
BROWSER_WAIT_AFTER_LOAD_MS=2000
BROWSER_FALLBACK_ONLY=true       # Alleen als HTML faalt
BROWSER_MAX_CONCURRENT=1         # Voorzichtig met resources
```

### Production

```env
ENABLE_BROWSER_SCRAPING=true
BROWSER_POOL_SIZE=5              # Meer voor throughput
BROWSER_TIMEOUT_SECONDS=15
BROWSER_WAIT_AFTER_LOAD_MS=3000  # Meer tijd voor complex JS
BROWSER_FALLBACK_ONLY=true       # Blijf HTML eerst proberen!
BROWSER_MAX_CONCURRENT=3
```

### Production (Hoge Volume)

```env
ENABLE_BROWSER_SCRAPING=true
BROWSER_POOL_SIZE=10
BROWSER_TIMEOUT_SECONDS=20
BROWSER_WAIT_AFTER_LOAD_MS=2000
BROWSER_FALLBACK_ONLY=true
BROWSER_MAX_CONCURRENT=5
```

---

## 📋 Implementation Checklist

- [ ] Rod dependency installeren
- [ ] Browser pool manager maken
- [ ] Browser extractor implementeren
- [ ] Fallback logica in content extractor
- [ ] Configuration toevoegen
- [ ] Service integratie
- [ ] Main.go browser pool init/shutdown
- [ ] Error handling & retries
- [ ] Resource limits & timeouts
- [ ] Logging & monitoring
- [ ] Testing met problematische artikelen
- [ ] Performance metrics
- [ ] Documentation

---

## 🎊 Verwachte Resultaten

**VOOR (Alleen HTML):**
- 70-80% success rate
- Snel (1-2 sec per artikel)
- Laag resource gebruik

**NA (HTML + Browser Fallback):**
- **90-95% success rate** ⭐
- Gemiddeld 2-3 sec per artikel (meeste blijven snel!)
- Matig resource gebruik (200-300MB RAM)

**JavaScript-rendered sites die nu falen → Zullen werken!**

---

## 💰 Kosten vs Baten

### Kosten
- 💾 **Extra RAM:** ~200-300 MB
- ⚡ **CPU:** +10-15% gemiddeld
- 💿 **Disk:** +500 MB (Chrome binary)
- 🐌 **Langzamer:** JavaScript artikelen 5-10 sec vs 1-2 sec

### Baten
- ✅ **+20-25% success rate** (van 70% naar 90-95%)
- ✅ **JavaScript sites werken** (voorheen 0% → nu 90%+)
- ✅ **Betere gebruikerservaring** (meer volledige content)
- ✅ **Betere AI analyse** (meer data)

**Voor nieuws aggregator: WAARD DE INVESTERING!** ✨

---

## 🚀 Quick Start (Als Je Dit Wilt)

Zeg gewoon "ga verder" en ik implementeer:
1. Rod installatie
2. Browser pool (met pool management)
3. Browser extractor (met site-specific logic)
4. Fallback integratie (HTML → Browser → Summary)
5. Configuration (volledig configureerbaar)
6. Testing (met de artikel die nu faalt)

**Schatting:** ~2-3 uur implementatie voor volledige headless browser support! 🎯

---

**Wil je dat ik begin met de implementatie?**