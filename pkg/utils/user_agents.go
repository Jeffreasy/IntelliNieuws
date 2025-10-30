package utils

import (
	"math/rand"
	"sync"
)

// UserAgentRotator handles rotation of user agents and headers
type UserAgentRotator struct {
	userAgents      []string
	referers        []string
	acceptLanguages []string
	currentIndex    int
	mu              sync.Mutex
	enabled         bool
}

// RealBrowserUserAgents contains a list of real browser user agents
var RealBrowserUserAgents = []string{
	// Chrome on Windows
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",

	// Chrome on macOS
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",

	// Firefox on Windows
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:122.0) Gecko/20100101 Firefox/122.0",

	// Firefox on macOS
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:120.0) Gecko/20100101 Firefox/120.0",

	// Safari on macOS
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Safari/605.1.15",

	// Edge on Windows
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 Edg/119.0.0.0",

	// Chrome on Linux
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",

	// Mobile Chrome (Android)
	"Mozilla/5.0 (Linux; Android 13; SM-S918B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
	"Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",

	// Mobile Safari (iOS)
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
}

// CommonReferers contains realistic referer URLs
var CommonReferers = []string{
	"https://www.google.com/",
	"https://www.google.nl/",
	"https://www.bing.com/",
	"https://duckduckgo.com/",
	"https://www.nu.nl/",
	"https://www.ad.nl/",
	"https://nos.nl/",
	"https://www.reddit.com/",
	"https://twitter.com/",
	"https://www.facebook.com/",
	"", // Direct navigation (no referer)
	"",
	"",
}

// AcceptLanguages contains realistic language preferences
var AcceptLanguages = []string{
	"nl-NL,nl;q=0.9,en;q=0.8",
	"nl,en-US;q=0.9,en;q=0.8",
	"nl-NL,nl;q=0.9,en-US;q=0.8,en;q=0.7",
	"nl-BE,nl;q=0.9,en;q=0.8",
	"en-US,en;q=0.9,nl;q=0.8",
}

// NewUserAgentRotator creates a new user agent rotator
func NewUserAgentRotator(enabled bool) *UserAgentRotator {
	return &UserAgentRotator{
		userAgents:      RealBrowserUserAgents,
		referers:        CommonReferers,
		acceptLanguages: AcceptLanguages,
		enabled:         enabled,
		currentIndex:    rand.Intn(len(RealBrowserUserAgents)),
	}
}

// GetUserAgent returns next user agent (rotating or random)
func (r *UserAgentRotator) GetUserAgent() string {
	if !r.enabled || len(r.userAgents) == 0 {
		return RealBrowserUserAgents[0] // Default fallback
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Random selection instead of sequential rotation (more realistic)
	ua := r.userAgents[rand.Intn(len(r.userAgents))]
	return ua
}

// GetReferer returns a random referer
func (r *UserAgentRotator) GetReferer() string {
	if !r.enabled || len(r.referers) == 0 {
		return ""
	}

	return r.referers[rand.Intn(len(r.referers))]
}

// GetAcceptLanguage returns a random Accept-Language header
func (r *UserAgentRotator) GetAcceptLanguage() string {
	if !r.enabled || len(r.acceptLanguages) == 0 {
		return "nl-NL,nl;q=0.9,en;q=0.8"
	}

	return r.acceptLanguages[rand.Intn(len(r.acceptLanguages))]
}

// GetRandomHeaders returns a complete set of randomized headers
func (r *UserAgentRotator) GetRandomHeaders() map[string]string {
	headers := map[string]string{
		"User-Agent":                r.GetUserAgent(),
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
		"Accept-Language":           r.GetAcceptLanguage(),
		"Accept-Encoding":           "gzip, deflate, br",
		"DNT":                       "1",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
	}

	// Add referer sometimes (not always - more realistic)
	if rand.Float64() > 0.3 { // 70% chance of having referer
		referer := r.GetReferer()
		if referer != "" {
			headers["Referer"] = referer
		}
	}

	// Add cache headers sometimes
	if rand.Float64() > 0.5 {
		headers["Cache-Control"] = "max-age=0"
	}

	return headers
}
