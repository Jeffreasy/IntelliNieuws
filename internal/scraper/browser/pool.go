package browser

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/jeffrey/intellinieuws/pkg/logger"
)

// BrowserPool manages reusable browser instances for efficient scraping
type BrowserPool struct {
	browsers  []*rod.Browser
	mu        sync.Mutex
	size      int
	launcher  *launcher.Launcher
	logger    *logger.Logger
	launchURL string
	closed    bool
}

// NewBrowserPool creates a new browser pool with specified size
// For Windows, it will auto-detect Chrome or download if not found
func NewBrowserPool(size int, log *logger.Logger) (*BrowserPool, error) {
	poolLogger := log.WithComponent("browser-pool")
	poolLogger.Infof("Initializing browser pool with %d instances", size)

	// Configure launcher for Windows with stealth mode
	l := launcher.New().
		Headless(true).
		Leakless(true).               // Prevent Chrome detection leaks
		NoSandbox(true).              // Required for Windows without admin
		Set("disable-dev-shm-usage"). // Prevent memory issues
		Set("disable-gpu").           // Better compatibility
		Set("disable-software-rasterizer").
		Set("disable-extensions").
		Set("disable-default-apps").
		Set("disable-blink-features", "AutomationControlled"). // Hide automation
		Set("window-size", "1920,1080")                        // Realistic window size

	// Launch Chrome (will auto-download if not found on Windows)
	poolLogger.Info("Launching Chrome browser...")
	url, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("failed to launch Chrome: %w (Chrome will be auto-downloaded on first run)", err)
	}

	poolLogger.Infof("Chrome launched successfully at %s", url)

	pool := &BrowserPool{
		browsers:  make([]*rod.Browser, 0, size),
		size:      size,
		launcher:  l,
		launchURL: url,
		logger:    poolLogger,
	}

	// Pre-create browser instances with stealth mode
	poolLogger.Info("Creating browser instances with stealth mode...")
	for i := 0; i < size; i++ {
		browser := rod.New().
			ControlURL(url).
			MustConnect().
			NoDefaultDevice(). // Remove automation markers
			MustIncognito()    // Use incognito mode
		pool.browsers = append(pool.browsers, browser)
		poolLogger.Debugf("Browser instance %d/%d created (stealth enabled)", i+1, size)
	}

	poolLogger.Infof("Browser pool ready: %d instances available", size)
	return pool, nil
}

// Acquire gets a browser from the pool (blocks if none available)
func (p *BrowserPool) Acquire(ctx context.Context) (*rod.Browser, error) {
	// Wait for available browser with timeout
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			p.mu.Lock()
			if p.closed {
				p.mu.Unlock()
				return nil, fmt.Errorf("browser pool is closed")
			}

			if len(p.browsers) > 0 {
				browser := p.browsers[0]
				p.browsers = p.browsers[1:]
				p.mu.Unlock()
				p.logger.Debugf("Browser acquired (%d remaining in pool)", len(p.browsers))
				return browser, nil
			}
			p.mu.Unlock()
		}
	}
}

// Release returns a browser to the pool
func (p *BrowserPool) Release(browser *rod.Browser) {
	if browser == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		// Pool is closed, close this browser
		browser.MustClose()
		return
	}

	p.browsers = append(p.browsers, browser)
	p.logger.Debugf("Browser released (%d available in pool)", len(p.browsers))
}

// Close shuts down all browsers and cleans up resources
func (p *BrowserPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	p.closed = true
	p.logger.Info("Closing browser pool...")

	// Close all browsers
	for i, browser := range p.browsers {
		browser.MustClose()
		p.logger.Debugf("Closed browser instance %d/%d", i+1, len(p.browsers))
	}

	// Cleanup launcher
	p.launcher.Cleanup()
	p.logger.Info("Browser pool closed successfully")
}

// GetStats returns pool statistics
func (p *BrowserPool) GetStats() map[string]interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	return map[string]interface{}{
		"pool_size":  p.size,
		"available":  len(p.browsers),
		"in_use":     p.size - len(p.browsers),
		"closed":     p.closed,
		"launch_url": p.launchURL,
	}
}

// IsAvailable checks if pool has available browsers
func (p *BrowserPool) IsAvailable() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.browsers) > 0 && !p.closed
}
