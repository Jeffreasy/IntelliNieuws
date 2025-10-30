package html

import (
	"compress/gzip"
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jeffrey/intellinieuws/pkg/logger"
	"github.com/jeffrey/intellinieuws/pkg/utils"
	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html/charset"
)

// BrowserExtractor interface for fallback
type BrowserExtractor interface {
	ExtractContent(ctx context.Context, url string, source string) (string, error)
}

// ContentExtractor extracts main content from HTML pages with optional browser fallback
type ContentExtractor struct {
	client           *http.Client
	sanitizer        *bluemonday.Policy
	logger           *logger.Logger
	userAgent        string
	userAgentRotator *utils.UserAgentRotator
	browserExtractor BrowserExtractor
	useBrowser       bool
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
		sanitizer:        bluemonday.StrictPolicy(), // Only text, no HTML
		logger:           log.WithComponent("html-extractor"),
		userAgent:        userAgent,
		userAgentRotator: utils.NewUserAgentRotator(true), // v3.0: Enable rotation
	}
}

// SetBrowserExtractor sets the browser extractor for fallback
func (e *ContentExtractor) SetBrowserExtractor(browserExtractor BrowserExtractor, enabled bool) {
	e.browserExtractor = browserExtractor
	e.useBrowser = enabled
	if enabled {
		e.logger.Info("Browser fallback enabled for content extraction")
	}
}

// ExtractContent downloads and extracts main content from URL with browser fallback
func (e *ContentExtractor) ExtractContent(ctx context.Context, url string, source string) (string, error) {
	e.logger.Debugf("Extracting content from %s (source: %s)", url, source)

	// Try HTML extraction first (fast)
	content, htmlErr := e.extractHTML(ctx, url, source)
	if htmlErr == nil && len(content) > 200 {
		e.logger.Infof("HTML extraction successful: %d characters from %s", len(content), url)
		return content, nil
	}

	// Log HTML extraction failure
	if htmlErr != nil {
		e.logger.WithError(htmlErr).Debugf("HTML extraction failed for %s", url)
	}

	// Try browser extraction if enabled and HTML failed
	if e.useBrowser && e.browserExtractor != nil {
		e.logger.Infof("HTML extraction failed, trying browser for %s", url)

		browserContent, browserErr := e.browserExtractor.ExtractContent(ctx, url, source)
		if browserErr == nil && len(browserContent) > 200 {
			e.logger.Infof("Browser extraction successful: %d characters from %s", len(browserContent), url)
			return browserContent, nil
		}

		// Log browser failure
		if browserErr != nil {
			e.logger.WithError(browserErr).Warnf("Browser extraction also failed for %s", url)
		}
	}

	// Both methods failed
	if htmlErr != nil {
		return "", fmt.Errorf("all extraction methods failed: HTML error: %w", htmlErr)
	}

	return "", fmt.Errorf("no content found after trying all extraction methods")
}

// extractHTML performs HTML-based extraction (original logic)
func (e *ContentExtractor) extractHTML(ctx context.Context, url string, source string) (string, error) {
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
		e.logger.Debugf("Source-specific extraction failed for %s, using generic", source)
		content = e.extractGeneric(doc)
	}

	if content == "" {
		// Last resort: try to extract ANY text from body
		e.logger.Debugf("Generic extraction failed, trying body text extraction for %s", url)
		content = e.extractBodyText(doc)
	}

	if content == "" {
		return "", fmt.Errorf("no content found in HTML")
	}

	// Clean and sanitize
	content = e.sanitizer.Sanitize(content)
	content = e.cleanText(content)

	return content, nil
}

// fetchHTML downloads HTML from URL with stealth headers and proper encoding handling
func (e *ContentExtractor) fetchHTML(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	// v3.0: Use rotated user-agent if enabled, otherwise use default
	userAgent := e.userAgent
	if e.userAgentRotator != nil {
		userAgent = e.userAgentRotator.GetUserAgent()
	}

	// v3.0: Set realistic headers with rotation
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")

	// Rotate Accept-Language for realism
	acceptLang := "nl-NL,nl;q=0.9,en;q=0.8"
	if e.userAgentRotator != nil {
		acceptLang = e.userAgentRotator.GetAcceptLanguage()
	}
	req.Header.Set("Accept-Language", acceptLang)

	// Add referer if rotator provides one (realistic browsing)
	if e.userAgentRotator != nil {
		if referer := e.userAgentRotator.GetReferer(); referer != "" {
			req.Header.Set("Referer", referer)
		}
	}

	// Additional realistic headers
	// NOTE: Do NOT set Accept-Encoding manually - Go's http.Client handles gzip automatically
	// Setting it manually disables automatic decompression!
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := e.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// CRITICAL FIX: Handle gzip/deflate/br compression manually if needed
	var reader io.Reader = resp.Body

	// Check if response is gzip compressed
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
		e.logger.Debug("Decompressing gzip response")
	}

	// CRITICAL FIX: Auto-detect character encoding (handles ISO-8859-1, Windows-1252, UTF-8, etc.)
	// This converts non-UTF-8 content to UTF-8 automatically
	contentType := resp.Header.Get("Content-Type")
	utf8Reader, err := charset.NewReader(reader, contentType)
	if err != nil {
		// If charset detection fails, use original reader
		e.logger.WithError(err).Warn("Charset detection failed, using raw content")
		utf8Reader = reader
	}

	// Read the properly encoded content
	body, err := io.ReadAll(utf8Reader)
	if err != nil {
		return "", err
	}

	// Final safety: ensure valid UTF-8
	text := string(body)
	text = strings.ToValidUTF8(text, "")

	return text, nil
}

// extractBySource uses site-specific selectors
func (e *ContentExtractor) extractBySource(doc *goquery.Document, source string) string {
	selectors := getSiteSelectors(source)

	for _, selector := range selectors {
		content := doc.Find(selector).Text()
		if content != "" && len(strings.TrimSpace(content)) > 200 {
			e.logger.Debugf("Found content using selector '%s' for %s", selector, source)
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
		".content",
		"main",
	}

	for _, selector := range genericSelectors {
		content := doc.Find(selector).Text()
		if content != "" && len(strings.TrimSpace(content)) > 200 {
			e.logger.Debugf("Found content using generic selector '%s'", selector)
			return content
		}
	}

	// Last resort: get all paragraphs
	var paragraphs []string
	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		// Filter out short navigation text and common UI elements
		if len(text) > 50 && !isNavigationText(text) {
			paragraphs = append(paragraphs, text)
		}
	})

	if len(paragraphs) > 0 {
		e.logger.Debugf("Found %d paragraphs using fallback extraction", len(paragraphs))
		return strings.Join(paragraphs, "\n\n")
	}

	return ""
}

// extractBodyText extracts all text from body as last resort
func (e *ContentExtractor) extractBodyText(doc *goquery.Document) string {
	// Remove script, style, nav, header, footer elements
	doc.Find("script, style, nav, header, footer, aside, .advertisement, .ad, .menu").Remove()

	// Get all text from body
	bodyText := doc.Find("body").Text()

	// Split into paragraphs and filter
	lines := strings.Split(bodyText, "\n")
	var validLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Keep lines with substantial content
		if len(trimmed) > 100 && !isNavigationText(trimmed) {
			validLines = append(validLines, trimmed)
		}
	}

	if len(validLines) >= 3 { // At least 3 substantial paragraphs
		e.logger.Debugf("Extracted %d lines using body text fallback", len(validLines))
		return strings.Join(validLines, "\n\n")
	}

	return ""
}

// cleanText removes extra whitespace and normalizes text
func (e *ContentExtractor) cleanText(text string) string {
	// Decode HTML entities first (e.g., &amp;, &quot;, &#8220;, etc.)
	text = html.UnescapeString(text)

	// Remove multiple spaces
	text = strings.Join(strings.Fields(text), " ")

	// Remove multiple newlines
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}

	return strings.TrimSpace(text)
}

// isNavigationText checks if text is likely navigation/UI text
func isNavigationText(text string) bool {
	lowerText := strings.ToLower(text)
	navigationPhrases := []string{
		"lees meer",
		"lees ook",
		"delen",
		"reageer",
		"reacties",
		"advertentie",
		"cookie",
		"privacy",
		"volg ons",
		"nieuwsbrief",
	}

	for _, phrase := range navigationPhrases {
		if strings.Contains(lowerText, phrase) && len(text) < 100 {
			return true
		}
	}

	return false
}

// getSiteSelectors returns CSS selectors for specific news sites
func getSiteSelectors(source string) []string {
	selectors := map[string][]string{
		"nu.nl": {
			".article__body",
			".block-text",
			"article .text",
		},
		"ad.nl": {
			".article__body",
			".article-detail__body",
			"article .body",
		},
		"nos.nl": {
			".article-content",
			".content-area",
			"article .text",
		},
		"trouw.nl": {
			".article__body",
			".article-body",
		},
		"volkskrant.nl": {
			".article__body",
			".article-content",
		},
		"telegraaf.nl": {
			".ArticleBodyBlocks__body",
			"article .body",
		},
		"rtlnieuws.nl": {
			".article-body",
			".content-block",
		},
	}

	if sels, exists := selectors[source]; exists {
		return sels
	}

	return []string{} // Return empty, will trigger generic extraction
}

// ExtractMetadata extracts additional metadata from HTML
func (e *ContentExtractor) ExtractMetadata(ctx context.Context, url string) (map[string]string, error) {
	html, err := e.fetchHTML(ctx, url)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	metadata := make(map[string]string)

	// Extract Open Graph metadata
	doc.Find("meta[property^='og:']").Each(func(i int, s *goquery.Selection) {
		property, _ := s.Attr("property")
		content, _ := s.Attr("content")
		if property != "" && content != "" {
			metadata[property] = content
		}
	})

	// Extract Twitter Card metadata
	doc.Find("meta[name^='twitter:']").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		content, _ := s.Attr("content")
		if name != "" && content != "" {
			metadata[name] = content
		}
	})

	// Extract description
	if desc, exists := doc.Find("meta[name='description']").Attr("content"); exists {
		metadata["description"] = desc
	}

	return metadata, nil
}
