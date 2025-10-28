package utils

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/temoto/robotstxt"
)

// RobotsChecker manages robots.txt compliance checking
type RobotsChecker struct {
	cache     map[string]*robotsCacheEntry
	mu        sync.RWMutex
	userAgent string
}

type robotsCacheEntry struct {
	data      *robotstxt.RobotsData
	expiresAt time.Time
}

// NewRobotsChecker creates a new robots.txt checker
func NewRobotsChecker(userAgent string) *RobotsChecker {
	return &RobotsChecker{
		cache:     make(map[string]*robotsCacheEntry),
		userAgent: userAgent,
	}
}

// IsAllowed checks if the given URL is allowed to be scraped according to robots.txt
func (rc *RobotsChecker) IsAllowed(targetURL string) (bool, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return false, fmt.Errorf("invalid URL: %w", err)
	}

	// Build robots.txt URL
	robotsURL := fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, parsedURL.Host)

	// Check cache
	rc.mu.RLock()
	cached, exists := rc.cache[robotsURL]
	rc.mu.RUnlock()

	if exists && time.Now().Before(cached.expiresAt) {
		// Use cached robots.txt
		return cached.data.TestAgent(parsedURL.Path, rc.userAgent), nil
	}

	// Fetch robots.txt
	robotsData, err := rc.fetchRobotsTxt(robotsURL)
	if err != nil {
		// If robots.txt doesn't exist or error, allow by default
		return true, nil
	}

	// Cache the result (24 hours)
	rc.mu.Lock()
	rc.cache[robotsURL] = &robotsCacheEntry{
		data:      robotsData,
		expiresAt: time.Now().Add(24 * time.Hour),
	}
	rc.mu.Unlock()

	// Check if path is allowed
	return robotsData.TestAgent(parsedURL.Path, rc.userAgent), nil
}

// fetchRobotsTxt downloads and parses robots.txt
func (rc *RobotsChecker) fetchRobotsTxt(robotsURL string) (*robotstxt.RobotsData, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(robotsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch robots.txt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("robots.txt returned status %d", resp.StatusCode)
	}

	robotsData, err := robotstxt.FromResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse robots.txt: %w", err)
	}

	return robotsData, nil
}

// GetCrawlDelay returns the crawl delay for a target URL
// Note: This is a placeholder since robotstxt library doesn't provide CrawlDelay method
// We use our own rate limiting configuration instead
func (rc *RobotsChecker) GetCrawlDelay(targetURL string) time.Duration {
	// Return 0, we use our own configurable rate limiting
	// which is more reliable than parsing robots.txt Crawl-delay directive
	return 0
}

// ClearCache clears the robots.txt cache
func (rc *RobotsChecker) ClearCache() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.cache = make(map[string]*robotsCacheEntry)
}

// GetDomain extracts the domain from a URL
func GetDomain(targetURL string) (string, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	return parsedURL.Host, nil
}

// NormalizeURL normalizes a URL for consistency
func NormalizeURL(rawURL string) (string, error) {
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Remove trailing slash
	path := parsedURL.Path
	if strings.HasSuffix(path, "/") && len(path) > 1 {
		path = path[:len(path)-1]
	}

	normalized := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, path)
	if parsedURL.RawQuery != "" {
		normalized += "?" + parsedURL.RawQuery
	}

	return normalized, nil
}
