
# Hybrid Scraping Implementatie Plan

## Concept: RSS + Optionele Full-Text Extraction

**Strategie:**
1. ✅ **RSS scraping** (snel, metadata) - HEBBEN WE AL
2. ➕ **Optionele HTML scraping** (volledige tekst) - TE IMPLEMENTEREN
3. 🤖 **AI processing** op volledige tekst (betere analyse) - HEBBEN WE AL

## Architectuur Overview

```
RSS Feed → Basic Article Data (title, summary, URL)
    ↓
    ├─→ Opslaan in database ✅
    ↓
    └─→ [OPTIONEEL] Fetch full article HTML
             ↓
             ├─→ Extract main content
             ├─→ Clean & sanitize
             └─→ Update article.content in database
                      ↓
                      └─→ AI processes full text (better analysis!)
```

## Fase 1: Dependencies Toevoegen

### Benodigde Go Libraries

**Voor HTML parsing:**
```go
// go.mod
require (
    github.com/PuerkitoBio/goquery v1.8.1  // jQuery-like HTML parsing
    github.com/microcosm-cc/bluemonday v1.0.26  // HTML sanitization
    golang.org/x/net v0.17.0  // Charset detection
)
```

**Installeren:**
```powershell
go get github.com/PuerkitoBio/goquery@latest
go get github.com/microcosm-cc/bluemonday@latest
go get golang.org/x/net/html/charset@latest
```

## Fase 2: HTML Content Extractor Maken

### Nieuwe file: `internal/scraper/html/content_extractor.go`

```go
package html

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "strings"
    "time"

    "github.com/PuerkitoBio/goquery"
    "github.com/microcosm-cc/bluemonday"
    "github.com/jeffrey/nieuws-scraper/pkg/logger"
)

// ContentExtractor extracts main content from HTML pages
type ContentExtractor struct {
    client    *http.Client
    sanitizer *bluemonday.Policy
    logger    *logger.Logger
}

// NewContentExtractor creates a new content extractor
func NewContentExtractor(userAgent string, log *logger.Logger) *ContentExtractor {
    return &ContentExtractor{
        client: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                MaxIdleConns:        100,
                MaxIdleConnsPerHost: 10,
                IdleConnTimeout:     90 * time.Second,
            },
        },
        sanitizer: bluemonday.StrictPolicy(), // Only text, no HTML
        logger:    log.WithComponent("html-extractor"),
    }
}

// ExtractContent downloads and extracts main content from URL
func (e *ContentExtractor) ExtractContent(ctx context.Context, url string, source string) (string, error) {
    // Download HTML
    html, err := e.fetchHTML(ctx, url)
    if err != nil {
        return "", fmt.Errorf("failed to fetch HTML: %w", err)
    }

    // Parse with goquery
    doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
    if err != nil {
        return "", fmt.Errorf("failed to parse HTML: %w", err)
    }

    // Extract content based on source
    content := e.extractBySource(doc, source)
    if content == "" {
        // Fallback to generic extraction
        content = e.extractGeneric(doc)
    }

    // Clean and sanitize
    content = e.sanitizer.Sanitize(content)
    content = e.cleanText(content)

    return content, nil
}

// fetchHTML downloads HTML from URL
func (e *ContentExtractor) fetchHTML(ctx context.Context, url string) (string, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return "", err
    }

    req.Header.Set("User-Agent", "NieuwsScraper/1.0")
    req.Header.Set("Accept", "text/html")

    resp, err := e.client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("HTTP %d", resp.StatusCode)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(body), nil
}

// extractBySource uses site-specific selectors
func (e *ContentExtractor) extractBySource(doc *goquery.Document, source string) string {
    selectors := getSiteSelectors(source)
    
    for _, selector := range selectors {
        if content := doc.Find(selector).Text(); content != "" {
            return content
        }
    }
    
    return ""
}

// extractGeneric uses common article selectors as fallback
func (e *ContentExtractor) extractGeneric(doc *goquery.Document) string {
    // Try common article selectors
    genericSelectors := []string{
        "article",
        ".article-content",
        ".article-body",
        ".post-content",
        "main article",
        "[itemprop='articleBody']",
    }

    for _, selector := range genericSelectors {
        if content := doc.Find(selector).Text(); content != "" {
            return content
        }
    }

    // Last resort: get all paragraphs
    var paragraphs []string
    doc.Find("p").Each(func(i int, s *goquery.Selection) {
        text := strings.TrimSpace(s.Text())
        if len(text) > 50 { // Filter out short navigation text
            paragraphs = append(paragraphs, text)
        }
    })

    return strings.Join(paragraphs, "\n\n")
}

// cleanText removes extra whitespace and normalizes text
func (e *ContentExtractor) cleanText(text string) string {
    // Remove multiple spaces
    text = strings.Join(strings.Fields(text), " ")
    
    // Remove multiple newlines
    for strings.Contains(text, "\n\n\n") {
        text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
    }
    
    return strings.TrimSpace(text)
}

// getSiteSelectors returns CSS selectors for specific news sites
func getSiteSelectors(source string) []string {
    selectors := map[string][]string{
        "nu.nl": {
            ".article__body",
            ".block-text",
        },
        "ad.nl": {
            ".article__body",
            ".article-detail__body",
        },
        "nos.nl": {
            ".article-content",
            ".content-area",
        },
        "trouw.nl": {
            ".article__body",
        },
        "volkskrant.nl": {
            ".article__body",
        },
    }

    if sels, exists := selectors[source]; exists {
        return sels
    }

    return []string{} // Return empty, will trigger generic extraction
}
```

## Fase 3: Integratie met Bestaande Scraper

### Update: `internal/scraper/service.go`

```go
// Add to Service struct
type Service struct {
    rssScrap         *rss.Scraper
    contentExtractor *html.ContentExtractor  // NEW
    articleRepo      *repository.ArticleRepository
    // ... existing fields
}

// Update NewService
func NewService(
    cfg *config.ScraperConfig,
    articleRepo *repository.ArticleRepository,
    log *logger.Logger,
) *Service {
    return &Service{
        rssScrap:         rss.NewScraper(cfg.UserAgent, log),
        contentExtractor: html.NewContentExtractor(cfg.UserAgent, log), // NEW
        articleRepo:      articleRepo,
        // ... rest
    }
}

// NEW: EnrichArticleContent downloads full text for an article
func (s *Service) EnrichArticleContent(ctx context.Context, articleID int64) error {
    // Get article from database
    article, err := s.articleRepo.GetByID(ctx, articleID)
    if err != nil {
        return fmt.Errorf("failed to get article: %w", err)
    }

    // Skip if already has content
    if article.Content != "" {
        s.logger.Debugf("Article %d already has content, skipping", articleID)
        return nil
    }

    s.logger.Infof("Enriching article %d: %s", articleID, article.Title)

    // Extract full content
    content, err := s.contentExtractor.ExtractContent(ctx, article.URL, article.Source)
    if err != nil {
        s.logger.WithError(err).Warnf("Failed to extract content for article %d", articleID)
        return err
    }

    // Update article with full content
    if err := s.articleRepo.UpdateContent(ctx, articleID, content); err != nil {
        return fmt.Errorf("failed to update content: %w", err)
    }

    s.logger.Infof("Successfully enriched article %d with %d characters", articleID, len(content))
    return nil
}

// NEW: EnrichArticlesBatch enriches multiple articles
func (s *Service) EnrichArticlesBatch(ctx context.Context, articleIDs []int64) error {
    var wg sync.WaitGroup
    errors := make(chan error, len(articleIDs))
    
    // Limit concurrency to avoid overwhelming servers
    semaphore := make(chan struct{}, 3)
    
    for _, id := range articleIDs {
        wg.Add(1)
        go func(articleID int64) {
            defer wg.Done()
            
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            if err := s.EnrichArticleContent(ctx, articleID); err != nil {
                errors <- err
            }
        }(id)
    }
    
    wg.Wait()
    close(errors)
    
    // Collect errors
    var enrichErrors []error
    for err := range errors {
        enrichErrors = append(enrichErrors, err)
    }
    
    if len(enrichErrors) > 0 {
        return fmt.Errorf("%d enrichment errors occurred", len(enrichErrors))
    }
    
    return nil
}
```

## Fase 4: Database Schema Update

### Migration: `migrations/005_add_content_column.sql`

```sql
-- Add content column for full article text
ALTER TABLE articles 
ADD COLUMN IF NOT EXISTS content TEXT,
ADD COLUMN IF NOT EXISTS content_extracted BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS content_extracted_at TIMESTAMPTZ;

-- Index for finding articles without content
CREATE INDEX IF NOT EXISTS idx_articles_needs_content 
ON articles(content_extracted, created_at) 
WHERE content_extracted = FALSE OR content_extracted IS NULL;

-- Update existing articles
UPDATE articles 
SET content_extracted = FALSE 
WHERE content IS NULL OR content = '';

COMMENT ON COLUMN articles.content IS 'Full article text extracted from HTML';
COMMENT ON COLUMN articles.content_extracted IS 'Whether full content has been extracted';
COMMENT ON COLUMN articles.content_extracted_at IS 'When content was extracted';
```

## Fase 5: Repository Methods

### Update: `internal/repository/article_repository.go`

```go
// UpdateContent updates the full content of an article
func (r *ArticleRepository) UpdateContent(ctx context.Context, id int64, content string) error {
    query := `
        UPDATE articles
        SET content = $2,
            content_extracted = TRUE,
            content_extracted_at = NOW(),
            updated_at = NOW()
        WHERE id = $1
    `
    
    _, err := r.db.Exec(ctx, query, id, content)
    return err
}

// GetArticlesNeedingContent returns articles that need content extraction
func (r *ArticleRepository) GetArticlesNeedingContent(ctx context.Context, limit int) ([]int64, error) {
    query := `
        SELECT id
        FROM articles
        WHERE (content_extracted = FALSE OR content_extracted IS NULL)
          AND url IS NOT NULL
          AND url != ''
        ORDER BY created_at DESC
        LIMIT $1
    `
    
    rows, err := r.db.Query(ctx, query, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var ids []int64
    for rows.Next() {
        var id int64
        if err := rows.Scan(&id); err != nil {
            continue
        }
        ids = append(ids, id)
    }
    
    return ids, nil
}
```

## Fase 6: Configuration

### Update: `.env`

```env
# Full Content Extraction
ENABLE_FULL_CONTENT_EXTRACTION=true  # Toggle feature on/off
CONTENT_EXTRACTION_BATCH_SIZE=10     # How many articles to process at once
CONTENT_EXTRACTION_DELAY_SECONDS=2   # Delay between requests (be nice!)
CONTENT_EXTRACTION_ASYNC=true        # Process in background
```

## Fase 7: Background Processor

### Nieuwe file: `internal/scraper/content_processor.go`

```go
package scraper

import (
    "context"
    "time"
)

// ContentProcessor handles background content extraction
type ContentProcessor struct {
    service  *Service
    interval time.Duration
    enabled  bool
}

// Start begins background processing
func (p *ContentProcessor) Start(ctx context.Context) {
    if !p.enabled {
        return
    }
    
    ticker := time.NewTicker(p.interval)
    defer ticker.Stop()
    
    // Process immediately on start
    p.processArticles(ctx)
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            p.processArticles(ctx)
        }
    }
}

func (p *ContentProcessor) processArticles(ctx context.Context) {
    // Get articles needing content
    ids, err := p.service.articleRepo.GetArticlesNeedingContent(ctx, 10)
    if err != nil || len(ids) == 0 {
        return
    }
    
    // Enrich content
    p.service.EnrichArticlesBatch(ctx, ids)
}
```

## Implementatie Volgorde

### Week 1: Basis Setup
1. ✅ Dependencies installeren
2. ✅ HTML extractor maken (`content_extractor.go`)
3. ✅ Database migratie uitvoeren
4. ✅ Repository methods toevoegen

### Week 2: Integratie
5. ✅ Service methods toevoegen
6. ✅ Config toevoegen
7. ✅ Testen met enkele artikelen
8. ✅ CSS selectors verfijnen per site

### Week 3: Background Processing
9. ✅ Background processor maken
10. ✅ Integreren met main
11. ✅ Monitoring toevoegen
12. ✅ Error handling verbeteren

## Voordelen van Deze Aanpak

✅ **Optioneel** - Kan uitgeschakeld worden
✅ **Async** - Blokkeert RSS scraping niet
✅ **Gradueel** - Verwerkt artikelen op achtergrond
✅ **Betrouwbaar** - Heeft RSS fallback
✅ **Configureerbaar** - Per site aan te passen
✅ **Schaalbaar** - Batch processing
✅ **Respectvol** - Rate limiting ingebouwd

## AI Processing Verbetering

Met volledige tekst wordt AI analyse **veel beter**:

**Voor (alleen RSS summary):**
```
Title: "Marco Borsato terecht"
Summary: "OM eist 5 maanden cel"
AI: Sentiment = -0.3, Keywords = ["Borsato", "rechtszaak"]
```

**Na (volledige tekst):**
```
Title: "Marco Borsato terecht"
Content: "Marco Borsato staat terecht... [2000 woorden]"
AI: Sentiment = -0.7, Keywords = ["Borsato", "ontucht", "slachtoffer", "verdediging", "OM"], 
    Entities = ["Marco Borsato", "Openbaar Ministerie", "Rechtbank Amsterdam"],
    Categories = {"Crime": 0.9, "Entertainment": 0.6}
```

**Veel meer context = betere AI analyse!**

## Kosten Impact

**Extra kosten:**
- Bandwidth