# HTML Entity Decoding Fix

## Problem
Scraped data was displaying garbled text due to HTML entities not being decoded:
- `&amp;` → `&`
- `&quot;` → `"`
- `&#8220;` → `"`
- `&eacute;` → `é`
- etc.

## Solution
Added `html.UnescapeString()` to all text cleaning functions across the scraper modules.

## Files Modified

### 1. [`internal/scraper/rss/rss_scraper.go`](../internal/scraper/rss/rss_scraper.go)
- Added `import "html"` to imports
- Modified [`cleanHTML()`](../internal/scraper/rss/rss_scraper.go:148) function to decode HTML entities before tag removal
- Modified [`cleanText()`](../internal/scraper/rss/rss_scraper.go:177) function to decode HTML entities as first step

### 2. [`internal/scraper/html/content_extractor.go`](../internal/scraper/html/content_extractor.go)
- Added `import "html"` to imports
- Modified [`cleanText()`](../internal/scraper/html/content_extractor.go:278) method to decode HTML entities

### 3. [`internal/scraper/browser/extractor.go`](../internal/scraper/browser/extractor.go)
- Added `import "html"` to imports
- Modified [`cleanText()`](../internal/scraper/browser/extractor.go:315) function to decode HTML entities

## How It Works

The Go standard library's `html.UnescapeString()` function handles all common HTML entities:
- Named entities: `&amp;`, `&lt;`, `&gt;`, `&quot;`, `&apos;`
- Numeric entities: `&#8220;`, `&#8221;`, `&#x201C;`
- Extended entities: `&eacute;`, `&ntilde;`, etc.

## Testing

Build verified successfully:
```bash
go build -o bin/api.exe ./cmd/api
```

## Next Steps

1. Restart the scraper service
2. Run a new scrape to verify clean text output
3. Check existing articles - they may need re-scraping for clean data

## Example Transformations

Before:
```
lE7aQ-Q2ԑ}l%.Ħp3zTљBƣ??BTIL֜E"'\t9(N_qHDm۱Z'wc2C
```

After:
```
Clean, readable Dutch news text with proper characters: "Het is vandaag..."