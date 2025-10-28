package stock

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jeffrey/intellinieuws/pkg/logger"
	"github.com/redis/go-redis/v9"
)

const (
	fmpAPIURL    = "https://financialmodelingprep.com/api/v3"
	alphaAPIURL  = "https://www.alphavantage.co/query"
	cachePrefix  = "stock:"
	cacheQuote   = "quote:"
	cacheProfile = "profile:"
)

// Service handles stock data fetching and caching
type Service struct {
	config      *Config
	httpClient  *http.Client
	redis       *redis.Client
	logger      *logger.Logger
	rateLimiter *time.Ticker
}

// NewService creates a new stock service
func NewService(cfg *Config, redisClient *redis.Client, log *logger.Logger) *Service {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 5 * time.Minute // Default cache for 5 minutes
	}
	if cfg.RateLimitPerMin == 0 {
		cfg.RateLimitPerMin = 30 // Conservative default
	}

	var rateLimiter *time.Ticker
	if cfg.RateLimitPerMin > 0 {
		interval := time.Minute / time.Duration(cfg.RateLimitPerMin)
		rateLimiter = time.NewTicker(interval)
	}

	return &Service{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		redis:       redisClient,
		logger:      log.WithComponent("stock-service"),
		rateLimiter: rateLimiter,
	}
}

// GetQuote fetches stock quote with caching
func (s *Service) GetQuote(ctx context.Context, symbol string) (*StockQuote, error) {
	symbol = strings.ToUpper(symbol)

	// Check cache first
	if s.config.EnableCache && s.redis != nil {
		cacheKey := cachePrefix + cacheQuote + symbol
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var quote StockQuote
			if err := json.Unmarshal([]byte(cached), &quote); err == nil {
				s.logger.Debugf("Cache HIT for stock quote: %s", symbol)
				return &quote, nil
			}
		}
	}

	// Apply rate limiting
	if s.rateLimiter != nil {
		<-s.rateLimiter.C
	}

	// Fetch from API
	var quote *StockQuote
	var err error

	switch s.config.APIProvider {
	case "alphavantage":
		quote, err = s.fetchQuoteAlphaVantage(ctx, symbol)
	default: // "fmp" or anything else
		quote, err = s.fetchQuoteFMP(ctx, symbol)
	}

	if err != nil {
		return nil, err
	}

	// Cache the result
	if s.config.EnableCache && s.redis != nil && quote != nil {
		cacheKey := cachePrefix + cacheQuote + symbol
		data, _ := json.Marshal(quote)
		s.redis.Set(ctx, cacheKey, data, s.config.CacheTTL)
		s.logger.Debugf("Cached stock quote: %s (TTL: %v)", symbol, s.config.CacheTTL)
	}

	return quote, nil
}

// GetProfile fetches company profile with caching
func (s *Service) GetProfile(ctx context.Context, symbol string) (*StockProfile, error) {
	symbol = strings.ToUpper(symbol)

	// Check cache first
	if s.config.EnableCache && s.redis != nil {
		cacheKey := cachePrefix + cacheProfile + symbol
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var profile StockProfile
			if err := json.Unmarshal([]byte(cached), &profile); err == nil {
				s.logger.Debugf("Cache HIT for stock profile: %s", symbol)
				return &profile, nil
			}
		}
	}

	// Apply rate limiting
	if s.rateLimiter != nil {
		<-s.rateLimiter.C
	}

	// Fetch from API (FMP only for now)
	profile, err := s.fetchProfileFMP(ctx, symbol)
	if err != nil {
		return nil, err
	}

	// Cache the result (longer TTL for profiles as they change less frequently)
	if s.config.EnableCache && s.redis != nil && profile != nil {
		cacheKey := cachePrefix + cacheProfile + symbol
		data, _ := json.Marshal(profile)
		cacheTTL := 24 * time.Hour // Profiles change rarely
		s.redis.Set(ctx, cacheKey, data, cacheTTL)
		s.logger.Debugf("Cached stock profile: %s (TTL: %v)", symbol, cacheTTL)
	}

	return profile, nil
}

// GetMultipleQuotes fetches multiple stock quotes efficiently
func (s *Service) GetMultipleQuotes(ctx context.Context, symbols []string) (map[string]*StockQuote, error) {
	results := make(map[string]*StockQuote)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Limit concurrency to avoid overwhelming the API
	semaphore := make(chan struct{}, 5)

	for _, symbol := range symbols {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			quote, err := s.GetQuote(ctx, sym)
			if err != nil {
				s.logger.WithError(err).Warnf("Failed to fetch quote for %s", sym)
				return
			}

			mu.Lock()
			results[strings.ToUpper(sym)] = quote
			mu.Unlock()
		}(symbol)
	}

	wg.Wait()
	return results, nil
}

// fetchQuoteFMP fetches quote from Financial Modeling Prep
func (s *Service) fetchQuoteFMP(ctx context.Context, symbol string) (*StockQuote, error) {
	url := fmt.Sprintf("%s/quote/%s?apikey=%s", fmpAPIURL, symbol, s.config.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch quote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var fmpQuotes []FMPQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&fmpQuotes); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(fmpQuotes) == 0 {
		return nil, fmt.Errorf("no data found for symbol: %s", symbol)
	}

	fmpQuote := fmpQuotes[0]
	quote := &StockQuote{
		Symbol:            fmpQuote.Symbol,
		Name:              fmpQuote.Name,
		Price:             fmpQuote.Price,
		Change:            fmpQuote.Change,
		ChangePercent:     fmpQuote.ChangesPercentage,
		Volume:            fmpQuote.Volume,
		MarketCap:         fmpQuote.MarketCap,
		Exchange:          fmpQuote.Exchange,
		Currency:          "USD", // FMP uses USD by default
		LastUpdated:       time.Unix(fmpQuote.Timestamp, 0),
		PreviousClose:     fmpQuote.PreviousClose,
		Open:              fmpQuote.Open,
		DayHigh:           fmpQuote.DayHigh,
		DayLow:            fmpQuote.DayLow,
		YearHigh:          fmpQuote.YearHigh,
		YearLow:           fmpQuote.YearLow,
		PriceAvg50:        fmpQuote.PriceAvg50,
		PriceAvg200:       fmpQuote.PriceAvg200,
		EPS:               fmpQuote.EPS,
		PE:                fmpQuote.PE,
		SharesOutstanding: fmpQuote.SharesOutstanding,
	}

	return quote, nil
}

// fetchProfileFMP fetches company profile from FMP
func (s *Service) fetchProfileFMP(ctx context.Context, symbol string) (*StockProfile, error) {
	url := fmt.Sprintf("%s/profile/%s?apikey=%s", fmpAPIURL, symbol, s.config.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var fmpProfiles []FMPProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&fmpProfiles); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(fmpProfiles) == 0 {
		return nil, fmt.Errorf("no profile found for symbol: %s", symbol)
	}

	fmpProfile := fmpProfiles[0]
	profile := &StockProfile{
		Symbol:      fmpProfile.Symbol,
		CompanyName: fmpProfile.CompanyName,
		Currency:    fmpProfile.Currency,
		Exchange:    fmpProfile.Exchange,
		Industry:    fmpProfile.Industry,
		Sector:      fmpProfile.Sector,
		Website:     fmpProfile.Website,
		Description: fmpProfile.Description,
		CEO:         fmpProfile.CEO,
		Country:     fmpProfile.Country,
		IPODate:     fmpProfile.IPODate,
	}

	return profile, nil
}

// fetchQuoteAlphaVantage fetches quote from Alpha Vantage (fallback)
func (s *Service) fetchQuoteAlphaVantage(ctx context.Context, symbol string) (*StockQuote, error) {
	url := fmt.Sprintf("%s?function=GLOBAL_QUOTE&symbol=%s&apikey=%s", alphaAPIURL, symbol, s.config.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch quote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var avResp AlphaVantageQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&avResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Parse Alpha Vantage response (strings to floats)
	quote := &StockQuote{
		Symbol:   symbol,
		Exchange: "Unknown",
		Currency: "USD",
	}

	// Parse numeric values (simplified, needs proper error handling)
	fmt.Sscanf(avResp.GlobalQuote.Price, "%f", &quote.Price)
	fmt.Sscanf(avResp.GlobalQuote.Change, "%f", &quote.Change)
	fmt.Sscanf(avResp.GlobalQuote.ChangePercent, "%f%%", &quote.ChangePercent)
	fmt.Sscanf(avResp.GlobalQuote.PreviousClose, "%f", &quote.PreviousClose)
	fmt.Sscanf(avResp.GlobalQuote.Open, "%f", &quote.Open)
	fmt.Sscanf(avResp.GlobalQuote.High, "%f", &quote.DayHigh)
	fmt.Sscanf(avResp.GlobalQuote.Low, "%f", &quote.DayLow)

	quote.LastUpdated = time.Now()

	return quote, nil
}

// Close stops the rate limiter
func (s *Service) Close() {
	if s.rateLimiter != nil {
		s.rateLimiter.Stop()
	}
}

// GetCacheStats returns cache statistics
func (s *Service) GetCacheStats(ctx context.Context) map[string]interface{} {
	stats := map[string]interface{}{
		"enabled": s.config.EnableCache,
		"ttl":     s.config.CacheTTL.String(),
	}

	if s.redis != nil {
		// Count cached items (quotes and profiles)
		quoteKeys, _ := s.redis.Keys(ctx, cachePrefix+cacheQuote+"*").Result()
		profileKeys, _ := s.redis.Keys(ctx, cachePrefix+cacheProfile+"*").Result()
		stats["cached_quotes"] = len(quoteKeys)
		stats["cached_profiles"] = len(profileKeys)
	}

	return stats
}
