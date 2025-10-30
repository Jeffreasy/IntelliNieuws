package browser

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/jeffrey/intellinieuws/pkg/logger"
)

// BrowserPool manages reusable browser instances for efficient scraping
type BrowserPool struct {
	available chan *rod.Browser // Channel for available browsers (optimized)
	size      int
	launcher  *launcher.Launcher
	logger    *logger.Logger
	launchURL string
	closed    bool
	mu        sync.Mutex
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
		available: make(chan *rod.Browser, size), // Buffered channel for instant signaling
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

		// Put browser in available channel immediately
		pool.available <- browser
		poolLogger.Debugf("Browser instance %d/%d created (stealth enabled)", i+1, size)
	}

	poolLogger.Infof("Browser pool ready: %d instances available", size)
	return pool, nil
}

// Acquire gets a browser from the pool (blocks if none available)
// Optimized: Uses channel-based signaling instead of polling for instant acquisition
func (p *BrowserPool) Acquire(ctx context.Context) (*rod.Browser, error) {
	// Check if pool is closed first
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, fmt.Errorf("browser pool is closed")
	}
	p.mu.Unlock()

	// Wait for available browser via channel (instant, no polling delay)
	select {
	case browser := <-p.available:
		p.logger.Debugf("Browser acquired (%d remaining in pool)", len(p.available))
		return browser, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Release returns a browser to the pool
// Optimized: Non-blocking channel send for instant availability signaling
func (p *BrowserPool) Release(browser *rod.Browser) {
	if browser == nil {
		return
	}

	p.mu.Lock()
	closed := p.closed
	p.mu.Unlock()

	if closed {
		// Pool is closed, close this browser
		browser.MustClose()
		p.logger.Debug("Browser closed (pool is closed)")
		return
	}

	// Non-blocking send to available channel
	select {
	case p.available <- browser:
		p.logger.Debugf("Browser released (%d available in pool)", len(p.available))
	default:
		// Channel full (should never happen with correct pool size)
		p.logger.Warn("Browser pool channel full, closing browser")
		browser.MustClose()
	}
}

// Close shuts down all browsers and cleans up resources
func (p *BrowserPool) Close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	p.mu.Unlock()

	p.logger.Info("Closing browser pool...")

	// Close channel and drain all browsers
	close(p.available)
	count := 0
	for browser := range p.available {
		browser.MustClose()
		count++
		p.logger.Debugf("Closed browser instance %d/%d", count, p.size)
	}

	// Cleanup launcher
	p.launcher.Cleanup()
	p.logger.Info("Browser pool closed successfully")
}

// GetStats returns pool statistics
func (p *BrowserPool) GetStats() map[string]interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	available := len(p.available)
	return map[string]interface{}{
		"pool_size":  p.size,
		"available":  available,
		"in_use":     p.size - available,
		"closed":     p.closed,
		"launch_url": p.launchURL,
	}
}

// IsAvailable checks if pool has available browsers
func (p *BrowserPool) IsAvailable() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.available) > 0 && !p.closed
}
