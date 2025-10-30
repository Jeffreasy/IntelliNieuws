package browser

import (
	"context"
	"fmt"
	"html"
	"math/rand"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/jeffrey/intellinieuws/pkg/logger"
	"github.com/jeffrey/intellinieuws/pkg/utils"
)

// Extractor extracts content from JavaScript-rendered pages using headless Chrome
type Extractor struct {
	pool             *BrowserPool
	logger           *logger.Logger
	timeout          time.Duration
	waitAfterLoad    time.Duration
	fallbackOnly     bool
	maxConcurrent    int
	semaphore        chan struct{}
	userAgentRotator *utils.UserAgentRotator
}

// ExtractorConfig holds browser extractor configuration
type ExtractorConfig struct {
	Timeout       time.Duration
	WaitAfterLoad time.Duration
	FallbackOnly  bool
	MaxConcurrent int
}

// NewExtractor creates a new browser-based content extractor
func NewExtractor(pool *BrowserPool, config ExtractorConfig, log *logger.Logger) *Extractor {
	return &Extractor{
		pool:             pool,
		logger:           log.WithComponent("browser-extractor"),
		timeout:          config.Timeout,
		waitAfterLoad:    config.WaitAfterLoad,
		fallbackOnly:     config.FallbackOnly,
		maxConcurrent:    config.MaxConcurrent,
		semaphore:        make(chan struct{}, config.MaxConcurrent),
		userAgentRotator: utils.NewUserAgentRotator(true), // v3.0: Enable rotation
	}
}

// ExtractContent extracts article content using headless browser
func (e *Extractor) ExtractContent(ctx context.Context, url string, source string) (string, error) {
	// Acquire semaphore to limit concurrent browser operations
	select {
	case e.semaphore <- struct{}{}:
		defer func() { <-e.semaphore }()
	case <-ctx.Done():
		return "", ctx.Err()
	}

	e.logger.Infof("Browser extracting from %s (source: %s)", url, source)
	startTime := time.Now()

	// Acquire browser from pool with timeout
	acquireCtx, acquireCancel := context.WithTimeout(ctx, 5*time.Second)
	defer acquireCancel()

	browser, err := e.pool.Acquire(acquireCtx)
	if err != nil {
		return "", fmt.Errorf("failed to acquire browser: %w", err)
	}
	defer e.pool.Release(browser)

	// Create page with timeout and stealth
	page, err := browser.Timeout(e.timeout).Page(proto.TargetCreateTarget{URL: ""})
	if err != nil {
		return "", fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()

	// Apply stealth mode to evade detection
	_, err = page.Eval(`() => {
		// Override navigator.webdriver
		Object.defineProperty(navigator, 'webdriver', {get: () => false});
		
		// Override chrome detection
		window.chrome = {runtime: {}};
		
		// Override permissions
		const originalQuery = window.navigator.permissions.query;
		window.navigator.permissions.query = (parameters) => (
			parameters.name === 'notifications' ?
				Promise.resolve({state: Notification.permission}) :
				originalQuery(parameters)
		);
	}`)
	if err != nil {
		e.logger.WithError(err).Warn("Failed to apply stealth mode")
	}

	// v3.0: Use rotated realistic user agent for stealth
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	if e.userAgentRotator != nil {
		userAgent = e.userAgentRotator.GetUserAgent()
		e.logger.Debugf("Using rotated user-agent for stealth")
	}

	if err := page.SetUserAgent(&proto.NetworkSetUserAgentOverride{UserAgent: userAgent}); err != nil {
		e.logger.WithError(err).Warn("Failed to set user agent")
	}

	// Set realistic viewport
	if err := page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:             1920,
		Height:            1080,
		DeviceScaleFactor: 1,
		Mobile:            false,
	}); err != nil {
		e.logger.WithError(err).Warn("Failed to set viewport")
	}

	// Navigate to URL
	if err := page.Navigate(url); err != nil {
		return "", fmt.Errorf("failed to navigate: %w", err)
	}

	// Wait for page to load
	if err := page.WaitLoad(); err != nil {
		return "", fmt.Errorf("page load timeout: %w", err)
	}

	// Additional wait for JavaScript rendering with random variation (mimic human)
	randomDelay := e.waitAfterLoad + time.Duration(rand.Intn(1000))*time.Millisecond
	e.logger.Debugf("Waiting %v for JavaScript to render (human-like)", randomDelay)
	time.Sleep(randomDelay)

	// Handle cookie consent popups (common on Dutch news sites)
	e.handleCookieConsent(page)

	// Optional: Random scroll to trigger lazy-loaded content
	_, _ = page.Eval(`window.scrollTo(0, document.body.scrollHeight / 2)`)
	time.Sleep(500 * time.Millisecond)

	// Try site-specific extraction first (using goquery for better parsing)
	content, err := e.extractBySource(page, source)
	if err == nil && len(content) > 200 {
		duration := time.Since(startTime)
		e.logger.Infof("Browser extracted %d characters from %s in %v (site-specific)", len(content), url, duration)
		return content, nil
	}

	// Fallback to generic extraction
	e.logger.Debugf("Site-specific extraction failed, trying generic for %s", source)
	content, err = e.extractGeneric(page)
	if err != nil {
		duration := time.Since(startTime)
		e.logger.WithError(err).Warnf("Browser extraction failed for %s after %v", url, duration)
		return "", err
	}

	if len(content) < 200 {
		return "", fmt.Errorf("extracted content too short (%d chars)", len(content))
	}

	duration := time.Since(startTime)
	e.logger.Infof("Browser extracted %d characters from %s in %v (generic)", len(content), url, duration)
	return content, nil
}

// extractBySource tries site-specific selectors using goquery for better parsing
func (e *Extractor) extractBySource(page *rod.Page, source string) (string, error) {
	// Get HTML and parse with goquery for better extraction
	html, err := page.HTML()
	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}

	selectors := getSiteSelectors(source)

	for _, selector := range selectors {
		text := doc.Find(selector).Text()
		text = cleanText(text)

		if len(text) > 200 {
			e.logger.Debugf("Found content using selector '%s' for %s (%d chars)", selector, source, len(text))
			return text, nil
		}
	}

	return "", fmt.Errorf("no site-specific selector matched")
}

// extractGeneric uses common article selectors with goquery
func (e *Extractor) extractGeneric(page *rod.Page) (string, error) {
	html, err := page.HTML()
	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}

	// Remove noise elements before extraction
	doc.Find("script, style, nav, header, footer, aside, .advertisement, .ad, .menu, .cookie-banner").Remove()

	// Common article selectors (ordered by specificity)
	genericSelectors := []string{
		"article",
		"[role='main'] article",
		"main article",
		".article-content",
		".article-body",
		".post-content",
		"[itemprop='articleBody']",
		"main",
		"[role='main']",
	}

	for _, selector := range genericSelectors {
		text := doc.Find(selector).Text()
		text = cleanText(text)

		if len(text) > 200 {
			e.logger.Debugf("Found content using generic selector '%s' (%d chars)", selector, len(text))
			return text, nil
		}
	}

	// Last resort: get all paragraphs
	return e.extractParagraphs(page)
}

// extractParagraphs extracts all paragraph text as last resort
func (e *Extractor) extractParagraphs(page *rod.Page) (string, error) {
	elements, err := page.Timeout(2 * time.Second).Elements("p")
	if err != nil {
		return "", err
	}

	var paragraphs []string
	for _, element := range elements {
		text, err := element.Text()
		if err != nil {
			continue
		}

		text = strings.TrimSpace(text)
		// Filter short snippets and navigation
		if len(text) > 50 && !isNavigationText(text) {
			paragraphs = append(paragraphs, text)
		}
	}

	if len(paragraphs) < 3 {
		return "", fmt.Errorf("insufficient content found (%d paragraphs)", len(paragraphs))
	}

	e.logger.Debugf("Extracted %d paragraphs using fallback", len(paragraphs))
	return strings.Join(paragraphs, "\n\n"), nil
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
			"article",
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

	return []string{"article", "main"}
}

// cleanText removes extra whitespace and normalizes text
func cleanText(text string) string {
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
		"menu",
		"zoeken",
	}

	for _, phrase := range navigationPhrases {
		if strings.Contains(lowerText, phrase) && len(text) < 100 {
			return true
		}
	}

	return false
}

// handleCookieConsent tries to accept cookie consent popups
func (e *Extractor) handleCookieConsent(page *rod.Page) {
	// Common cookie consent button selectors for Dutch sites
	consentSelectors := []string{
		"button[class*='accept']",
		"button[class*='cookie']",
		"button[class*='akkoord']",
		"button[class*='toestemming']",
		".cookie-consent-accept",
		"#accept-cookies",
		"[data-testid='accept-cookies']",
		"button:contains('Accepteren')",
		"button:contains('Akkoord')",
	}

	for _, selector := range consentSelectors {
		element, err := page.Timeout(1 * time.Second).Element(selector)
		if err == nil {
			// Try to click, ignore errors
			if err := element.Click(proto.InputMouseButtonLeft, 1); err == nil {
				e.logger.Debug("Clicked cookie consent button")
				time.Sleep(500 * time.Millisecond)
				return
			}
		}
	}
}

// CanExtract checks if browser extraction is likely to work for a URL
func (e *Extractor) CanExtract(url string) bool {
	// Could add URL pattern matching here
	// For now, assume all URLs can be tried
	return true
}
