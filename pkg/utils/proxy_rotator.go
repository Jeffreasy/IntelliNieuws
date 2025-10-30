package utils

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// ProxyProvider represents a proxy service provider
type ProxyProvider string

const (
	ProxyProviderScraperAPI ProxyProvider = "scraperapi"
	ProxyProviderScrapeDo   ProxyProvider = "scrapedo"
	ProxyProviderDirect     ProxyProvider = "direct"
)

// ProxyConfig holds proxy configuration
type ProxyConfig struct {
	Provider         ProxyProvider
	ScraperAPIKey    string
	ScrapeDoToken    string
	EnableRotation   bool
	RotationStrategy string // "random", "round-robin", "failover"
}

// ProxyRotator manages proxy rotation for web scraping
type ProxyRotator struct {
	config           *ProxyConfig
	currentIndex     int
	mu               sync.Mutex
	failedProxies    map[string]int // Track failures per proxy
	lastUsed         map[string]time.Time
	enabled          bool
	scraperAPIClient *ScraperAPIClient
	scrapeDoClient   *ScrapeDoClient
}

// ScraperAPIClient handles ScraperAPI integration
type ScraperAPIClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// ScrapeDoClient handles Scrape.do integration
type ScrapeDoClient struct {
	token      string
	baseURL    string
	httpClient *http.Client
}

// NewProxyRotator creates a new proxy rotator
func NewProxyRotator(config *ProxyConfig) *ProxyRotator {
	rotator := &ProxyRotator{
		config:        config,
		enabled:       config.EnableRotation,
		failedProxies: make(map[string]int),
		lastUsed:      make(map[string]time.Time),
	}

	// Initialize ScraperAPI client if configured
	if config.ScraperAPIKey != "" {
		rotator.scraperAPIClient = &ScraperAPIClient{
			apiKey:  config.ScraperAPIKey,
			baseURL: "http://api.scraperapi.com",
			httpClient: &http.Client{
				Timeout: 30 * time.Second,
			},
		}
	}

	// Initialize Scrape.do client if configured
	if config.ScrapeDoToken != "" {
		rotator.scrapeDoClient = &ScrapeDoClient{
			token:   config.ScrapeDoToken,
			baseURL: "https://api.scrape.do",
			httpClient: &http.Client{
				Timeout: 30 * time.Second,
			},
		}
	}

	return rotator
}

// GetProxyURL returns a proxy URL based on provider and strategy
func (r *ProxyRotator) GetProxyURL(targetURL string) (string, error) {
	if !r.enabled {
		return targetURL, nil // Direct access, no proxy
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	switch r.config.Provider {
	case ProxyProviderScraperAPI:
		return r.getScraperAPIURL(targetURL)
	case ProxyProviderScrapeDo:
		return r.getScrapeDoURL(targetURL)
	default:
		return targetURL, nil // Direct
	}
}

// getScraperAPIURL builds ScraperAPI proxy URL
// Format: http://api.scraperapi.com?api_key=KEY&url=TARGET&render=true
func (r *ProxyRotator) getScraperAPIURL(targetURL string) (string, error) {
	if r.scraperAPIClient == nil {
		return "", fmt.Errorf("ScraperAPI client not configured")
	}

	params := url.Values{}
	params.Add("api_key", r.scraperAPIClient.apiKey)
	params.Add("url", targetURL)
	params.Add("render", "false")      // false for RSS, true for JS-heavy sites
	params.Add("country_code", "nl")   // Netherlands proxy
	params.Add("keep_headers", "true") // Preserve custom headers
	params.Add("autoparse", "false")   // We handle parsing
	params.Add("premium", "false")     // Use standard proxies (cheaper)

	proxyURL := fmt.Sprintf("%s?%s", r.scraperAPIClient.baseURL, params.Encode())
	return proxyURL, nil
}

// getScrapeDoURL builds Scrape.do proxy URL
// Format: https://api.scrape.do?token=TOKEN&url=TARGET&render=false
func (r *ProxyRotator) getScrapeDoURL(targetURL string) (string, error) {
	if r.scrapeDoClient == nil {
		return "", fmt.Errorf("Scrape.do client not configured")
	}

	params := url.Values{}
	params.Add("token", r.scrapeDoClient.token)
	params.Add("url", targetURL)
	params.Add("render", "false")       // false for RSS, true for JS
	params.Add("geoCode", "nl")         // Netherlands
	params.Add("customHeaders", "true") // Allow custom headers

	proxyURL := fmt.Sprintf("%s?%s", r.scrapeDoClient.baseURL, params.Encode())
	return proxyURL, nil
}

// GetProxyClient returns an HTTP client configured with current proxy
func (r *ProxyRotator) GetProxyClient() (*http.Client, error) {
	if !r.enabled {
		return &http.Client{
			Timeout: 30 * time.Second,
		}, nil
	}

	// ScraperAPI and Scrape.do use their own proxy infrastructure
	// We just need to make requests to their API endpoints
	return &http.Client{
		Timeout: 60 * time.Second, // Longer timeout for proxy services
	}, nil
}

// ShouldUseProxy determines if proxy should be used for this request
func (r *ProxyRotator) ShouldUseProxy(source string, errorRate float64) bool {
	if !r.enabled {
		return false
	}

	// Always use proxy in Fast profile (aggressive)
	if r.config.RotationStrategy == "always" {
		return true
	}

	// Use proxy if error rate is high (adaptive)
	if errorRate > 0.1 { // >10% errors
		return true
	}

	// Use proxy for specific sources known to have bot detection
	knownBotDetection := map[string]bool{
		"telegraaf.nl":  true,
		"volkskrant.nl": true,
		"trouw.nl":      true,
	}

	return knownBotDetection[source]
}

// RecordSuccess records a successful request
func (r *ProxyRotator) RecordSuccess(proxyURL string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Reset failure count on success
	delete(r.failedProxies, proxyURL)
	r.lastUsed[proxyURL] = time.Now()
}

// RecordFailure records a failed request
func (r *ProxyRotator) RecordFailure(proxyURL string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.failedProxies[proxyURL]++

	// Log if proxy is consistently failing
	if r.failedProxies[proxyURL] >= 3 {
		// Could implement automatic failover here
	}
}

// GetStats returns proxy rotation statistics
func (r *ProxyRotator) GetStats() map[string]interface{} {
	r.mu.Lock()
	defer r.mu.Unlock()

	return map[string]interface{}{
		"enabled":           r.enabled,
		"provider":          r.config.Provider,
		"has_scraperapi":    r.scraperAPIClient != nil,
		"has_scrapedo":      r.scrapeDoClient != nil,
		"failed_proxies":    len(r.failedProxies),
		"rotation_strategy": r.config.RotationStrategy,
	}
}

// IsEnabled returns whether proxy rotation is enabled
func (r *ProxyRotator) IsEnabled() bool {
	return r.enabled
}

// GetProvider returns the current proxy provider
func (r *ProxyRotator) GetProvider() ProxyProvider {
	return r.config.Provider
}
