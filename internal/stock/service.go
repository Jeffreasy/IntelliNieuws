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
	fmpAPIURL    = "https://financialmodelingprep.com/stable"
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

	// Profile API only available for FMP, not Alpha Vantage
	if s.config.APIProvider != "fmp" {
		// Return basic profile from quote data instead
		return s.getProfileFromQuote(ctx, symbol)
	}

	// Apply rate limiting
	if s.rateLimiter != nil {
		<-s.rateLimiter.C
	}

	// Fetch from API (FMP only)
	profile, err := s.fetchProfileFMP(ctx, symbol)
	if err != nil {
		// Fallback to quote-based profile
		s.logger.WithError(err).Warn("Failed to fetch FMP profile, falling back to quote-based profile")
		return s.getProfileFromQuote(ctx, symbol)
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

// getProfileFromQuote creates a basic profile from quote data (fallback for Alpha Vantage)
func (s *Service) getProfileFromQuote(ctx context.Context, symbol string) (*StockProfile, error) {
	quote, err := s.GetQuote(ctx, symbol)
	if err != nil {
		return nil, err
	}

	profile := &StockProfile{
		Symbol:      quote.Symbol,
		CompanyName: quote.Name,
		Currency:    quote.Currency,
		Exchange:    quote.Exchange,
	}

	return profile, nil
}

// GetMultipleQuotes fetches multiple stock quotes efficiently using batch API
func (s *Service) GetMultipleQuotes(ctx context.Context, symbols []string) (map[string]*StockQuote, error) {
	if len(symbols) == 0 {
		return make(map[string]*StockQuote), nil
	}

	// Normalize and deduplicate symbols
	uniqueSymbols := make(map[string]bool)
	for _, sym := range symbols {
		uniqueSymbols[strings.ToUpper(sym)] = true
	}

	// Check cache first for all symbols
	results := make(map[string]*StockQuote)
	uncachedSymbols := make([]string, 0)

	if s.config.EnableCache && s.redis != nil {
		for symbol := range uniqueSymbols {
			cacheKey := cachePrefix + cacheQuote + symbol
			cached, err := s.redis.Get(ctx, cacheKey).Result()
			if err == nil {
				var quote StockQuote
				if err := json.Unmarshal([]byte(cached), &quote); err == nil {
					results[symbol] = &quote
					continue
				}
			}
			uncachedSymbols = append(uncachedSymbols, symbol)
		}

		if len(uncachedSymbols) == 0 {
			s.logger.Debugf("Batch cache HIT for all %d symbols", len(symbols))
			return results, nil
		}
		s.logger.Debugf("Batch cache: %d hits, %d misses", len(results), len(uncachedSymbols))
	} else {
		for symbol := range uniqueSymbols {
			uncachedSymbols = append(uncachedSymbols, symbol)
		}
	}

	// Apply rate limiting once for the batch
	if s.rateLimiter != nil {
		<-s.rateLimiter.C
	}

	// Fetch uncached quotes using batch API
	var batchQuotes map[string]*StockQuote
	var err error

	switch s.config.APIProvider {
	case "alphavantage":
		// Alpha Vantage doesn't support batch, fallback to individual calls
		batchQuotes, err = s.fetchQuotesIndividual(ctx, uncachedSymbols)
	default: // "fmp" - try batch first, fallback to individual if premium required
		batchQuotes, err = s.fetchQuotesBatchFMP(ctx, uncachedSymbols)

		// If batch fails due to premium requirement, fallback to individual calls
		if err != nil && strings.Contains(err.Error(), "Premium") {
			s.logger.Warn("Batch API requires premium subscription, falling back to individual calls")
			batchQuotes, err = s.fetchQuotesIndividual(ctx, uncachedSymbols)
		}
	}

	if err != nil {
		return results, err
	}

	// Cache the fetched quotes and merge with cached results
	if s.config.EnableCache && s.redis != nil {
		for symbol, quote := range batchQuotes {
			cacheKey := cachePrefix + cacheQuote + symbol
			data, _ := json.Marshal(quote)
			s.redis.Set(ctx, cacheKey, data, s.config.CacheTTL)
			results[symbol] = quote
		}
		s.logger.Debugf("Cached %d new stock quotes (TTL: %v)", len(batchQuotes), s.config.CacheTTL)
	} else {
		for symbol, quote := range batchQuotes {
			results[symbol] = quote
		}
	}

	return results, nil
}

// fetchQuotesBatchFMP fetches multiple quotes in a single API call (FMP batch endpoint)
func (s *Service) fetchQuotesBatchFMP(ctx context.Context, symbols []string) (map[string]*StockQuote, error) {
	if len(symbols) == 0 {
		return make(map[string]*StockQuote), nil
	}

	// FMP batch quote endpoint uses query parameter: ?symbol=SYM1,SYM2,...
	symbolsParam := strings.Join(symbols, ",")
	url := fmt.Sprintf("%s/quote?symbol=%s&apikey=%s", fmpAPIURL, symbolsParam, s.config.APIKey)

	s.logger.Debugf("Fetching batch quotes for %d symbols via FMP batch API", len(symbols))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch batch quotes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("batch API error (status %d): %s", resp.StatusCode, string(body))
	}

	var fmpQuotes []FMPQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&fmpQuotes); err != nil {
		return nil, fmt.Errorf("failed to decode batch response: %w", err)
	}

	// Convert to map
	results := make(map[string]*StockQuote)
	for _, fmpQuote := range fmpQuotes {
		quote := &StockQuote{
			Symbol:            fmpQuote.Symbol,
			Name:              fmpQuote.Name,
			Price:             fmpQuote.Price,
			Change:            fmpQuote.Change,
			ChangePercent:     fmpQuote.ChangesPercentage,
			Volume:            fmpQuote.Volume,
			MarketCap:         fmpQuote.MarketCap,
			Exchange:          fmpQuote.Exchange,
			Currency:          "USD",
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
		results[strings.ToUpper(fmpQuote.Symbol)] = quote
	}

	s.logger.Infof("✅ Fetched %d quotes in single batch API call (cost: 1 call, saved: %d calls)",
		len(results), len(symbols)-1)

	return results, nil
}

// fetchQuotesIndividual fetches quotes individually (fallback for providers without batch API)
func (s *Service) fetchQuotesIndividual(ctx context.Context, symbols []string) (map[string]*StockQuote, error) {
	results := make(map[string]*StockQuote)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Limit concurrency
	semaphore := make(chan struct{}, 5)

	for _, symbol := range symbols {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Apply rate limiting per call
			if s.rateLimiter != nil {
				<-s.rateLimiter.C
			}

			var quote *StockQuote
			var err error

			switch s.config.APIProvider {
			case "alphavantage":
				quote, err = s.fetchQuoteAlphaVantage(ctx, sym)
			default:
				quote, err = s.fetchQuoteFMP(ctx, sym)
			}

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
	s.logger.Warnf("⚠️  Used individual API calls for %d symbols (consider using FMP batch API)", len(symbols))
	return results, nil
}

// fetchQuoteFMP fetches quote from Financial Modeling Prep
func (s *Service) fetchQuoteFMP(ctx context.Context, symbol string) (*StockQuote, error) {
	url := fmt.Sprintf("%s/quote?symbol=%s&apikey=%s", fmpAPIURL, symbol, s.config.APIKey)

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
	url := fmt.Sprintf("%s/profile?symbol=%s&apikey=%s", fmpAPIURL, symbol, s.config.APIKey)

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

// GetStockNews fetches stock news for a specific symbol with caching
func (s *Service) GetStockNews(ctx context.Context, symbol string, limit int) ([]StockNews, error) {
	symbol = strings.ToUpper(symbol)

	if limit <= 0 || limit > 50 {
		limit = 10
	}

	// Check cache first
	cacheKey := cachePrefix + "news:" + symbol
	if s.config.EnableCache && s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var news []StockNews
			if err := json.Unmarshal([]byte(cached), &news); err == nil {
				s.logger.Debugf("Cache HIT for stock news: %s", symbol)
				return news, nil
			}
		}
	}

	// Apply rate limiting
	if s.rateLimiter != nil {
		<-s.rateLimiter.C
	}

	// Fetch from FMP API
	url := fmt.Sprintf("%s/stock_news?tickers=%s&limit=%d&apikey=%s", fmpAPIURL, symbol, limit, s.config.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch news: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var fmpNews []FMPNewsResponse
	if err := json.NewDecoder(resp.Body).Decode(&fmpNews); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to StockNews
	news := make([]StockNews, 0, len(fmpNews))
	for _, item := range fmpNews {
		publishedDate, _ := time.Parse("2006-01-02 15:04:05", item.PublishedDate)
		news = append(news, StockNews{
			Symbol:        item.Symbol,
			PublishedDate: publishedDate,
			Title:         item.Title,
			Image:         item.Image,
			Site:          item.Site,
			Text:          item.Text,
			URL:           item.URL,
		})
	}

	// Cache for 15 minutes (news changes frequently)
	if s.config.EnableCache && s.redis != nil && len(news) > 0 {
		data, _ := json.Marshal(news)
		s.redis.Set(ctx, cacheKey, data, 15*time.Minute)
		s.logger.Debugf("Cached stock news: %s (TTL: 15m)", symbol)
	}

	return news, nil
}

// GetHistoricalPrices fetches historical price data with caching
func (s *Service) GetHistoricalPrices(ctx context.Context, symbol string, from, to time.Time) ([]HistoricalPrice, error) {
	symbol = strings.ToUpper(symbol)

	// Format dates
	fromStr := from.Format("2006-01-02")
	toStr := to.Format("2006-01-02")

	// Check cache
	cacheKey := fmt.Sprintf("%shistorical:%s:%s:%s", cachePrefix, symbol, fromStr, toStr)
	if s.config.EnableCache && s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var prices []HistoricalPrice
			if err := json.Unmarshal([]byte(cached), &prices); err == nil {
				s.logger.Debugf("Cache HIT for historical prices: %s", symbol)
				return prices, nil
			}
		}
	}

	// Apply rate limiting
	if s.rateLimiter != nil {
		<-s.rateLimiter.C
	}

	// Fetch from FMP API
	url := fmt.Sprintf("%s/historical-price-full/%s?from=%s&to=%s&apikey=%s",
		fmpAPIURL, symbol, fromStr, toStr, s.config.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch historical prices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var fmpHist FMPHistoricalResponse
	if err := json.NewDecoder(resp.Body).Decode(&fmpHist); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to HistoricalPrice
	prices := make([]HistoricalPrice, 0, len(fmpHist.Historical))
	for _, item := range fmpHist.Historical {
		date, _ := time.Parse("2006-01-02", item.Date)
		prices = append(prices, HistoricalPrice{
			Date:          date,
			Open:          item.Open,
			High:          item.High,
			Low:           item.Low,
			Close:         item.Close,
			AdjClose:      item.AdjClose,
			Volume:        item.Volume,
			Change:        item.Change,
			ChangePercent: item.ChangePercent,
		})
	}

	// Cache for 24 hours (historical data doesn't change)
	if s.config.EnableCache && s.redis != nil && len(prices) > 0 {
		data, _ := json.Marshal(prices)
		s.redis.Set(ctx, cacheKey, data, 24*time.Hour)
		s.logger.Debugf("Cached historical prices: %s (TTL: 24h)", symbol)
	}

	return prices, nil
}

// GetKeyMetrics fetches key financial metrics with caching
func (s *Service) GetKeyMetrics(ctx context.Context, symbol string) (*KeyMetrics, error) {
	symbol = strings.ToUpper(symbol)

	// Check cache (longer TTL for metrics)
	cacheKey := cachePrefix + "metrics:" + symbol
	if s.config.EnableCache && s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var metrics KeyMetrics
			if err := json.Unmarshal([]byte(cached), &metrics); err == nil {
				s.logger.Debugf("Cache HIT for key metrics: %s", symbol)
				return &metrics, nil
			}
		}
	}

	// Apply rate limiting
	if s.rateLimiter != nil {
		<-s.rateLimiter.C
	}

	// Fetch from FMP API (TTM = trailing twelve months)
	url := fmt.Sprintf("%s/key-metrics-ttm/%s?apikey=%s", fmpAPIURL, symbol, s.config.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch key metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var fmpMetrics []FMPKeyMetricsResponse
	if err := json.NewDecoder(resp.Body).Decode(&fmpMetrics); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(fmpMetrics) == 0 {
		return nil, fmt.Errorf("no metrics found for symbol: %s", symbol)
	}

	// Convert first result
	item := fmpMetrics[0]
	metrics := &KeyMetrics{
		Symbol:            item.Symbol,
		MarketCap:         item.MarketCap,
		PE:                item.PeRatio,
		PEG:               item.PegRatio,
		PB:                item.PriceToBookRatio,
		PS:                item.PriceToSalesRatio,
		ROE:               item.Roe,
		ROA:               item.Roa,
		DebtToEquity:      item.DebtToEquity,
		CurrentRatio:      item.CurrentRatio,
		DividendYield:     item.DividendYield,
		EPS:               item.Eps,
		RevenuePerShare:   item.RevenuePerShare,
		FreeCashFlowYield: item.FreeCashFlowYield,
	}

	// Cache for 1 hour (metrics update quarterly)
	if s.config.EnableCache && s.redis != nil {
		data, _ := json.Marshal(metrics)
		s.redis.Set(ctx, cacheKey, data, 1*time.Hour)
		s.logger.Debugf("Cached key metrics: %s (TTL: 1h)", symbol)
	}

	return metrics, nil
}

// GetEarningsCalendar fetches upcoming earnings announcements
func (s *Service) GetEarningsCalendar(ctx context.Context, from, to time.Time) ([]EarningsCalendar, error) {
	fromStr := from.Format("2006-01-02")
	toStr := to.Format("2006-01-02")

	// Check cache
	cacheKey := fmt.Sprintf("%searnings:%s:%s", cachePrefix, fromStr, toStr)
	if s.config.EnableCache && s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var calendar []EarningsCalendar
			if err := json.Unmarshal([]byte(cached), &calendar); err == nil {
				s.logger.Debugf("Cache HIT for earnings calendar")
				return calendar, nil
			}
		}
	}

	// Apply rate limiting
	if s.rateLimiter != nil {
		<-s.rateLimiter.C
	}

	// Fetch from FMP API
	url := fmt.Sprintf("%s/earnings-calendar?from=%s&to=%s&apikey=%s",
		fmpAPIURL, fromStr, toStr, s.config.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch earnings calendar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var fmpEarnings []FMPEarningsCalendarResponse
	if err := json.NewDecoder(resp.Body).Decode(&fmpEarnings); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to EarningsCalendar
	calendar := make([]EarningsCalendar, 0, len(fmpEarnings))
	for _, item := range fmpEarnings {
		date, _ := time.Parse("2006-01-02", item.Date)
		calendar = append(calendar, EarningsCalendar{
			Symbol:       item.Symbol,
			Date:         date,
			EPS:          item.Eps,
			EPSEstimated: item.EpsEstimated,
			Time:         item.Time,
			Revenue:      item.Revenue,
			RevenueEst:   item.RevenueEstimated,
		})
	}

	// Cache for 6 hours
	if s.config.EnableCache && s.redis != nil && len(calendar) > 0 {
		data, _ := json.Marshal(calendar)
		s.redis.Set(ctx, cacheKey, data, 6*time.Hour)
		s.logger.Debug("Cached earnings calendar (TTL: 6h)")
	}

	return calendar, nil
}

// SearchSymbol searches for company symbols by query
func (s *Service) SearchSymbol(ctx context.Context, query string, limit int) ([]StockProfile, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	// Apply rate limiting
	if s.rateLimiter != nil {
		<-s.rateLimiter.C
	}

	// Fetch from FMP API
	url := fmt.Sprintf("%s/search?query=%s&limit=%d&apikey=%s",
		fmpAPIURL, query, limit, s.config.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search symbols: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var results []struct {
		Symbol            string `json:"symbol"`
		Name              string `json:"name"`
		Currency          string `json:"currency"`
		StockExchange     string `json:"stockExchange"`
		ExchangeShortName string `json:"exchangeShortName"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to StockProfile
	profiles := make([]StockProfile, 0, len(results))
	for _, item := range results {
		profiles = append(profiles, StockProfile{
			Symbol:      item.Symbol,
			CompanyName: item.Name,
			Currency:    item.Currency,
			Exchange:    item.ExchangeShortName,
		})
	}

	return profiles, nil
}

// GetMarketGainers fetches top gaining stocks with caching
func (s *Service) GetMarketGainers(ctx context.Context) ([]MarketMover, error) {
	// Check cache (5 min TTL for market data)
	cacheKey := cachePrefix + "gainers"
	if s.config.EnableCache && s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var gainers []MarketMover
			if err := json.Unmarshal([]byte(cached), &gainers); err == nil {
				s.logger.Debug("Cache HIT for market gainers")
				return gainers, nil
			}
		}
	}

	// Apply rate limiting
	if s.rateLimiter != nil {
		<-s.rateLimiter.C
	}

	url := fmt.Sprintf("%s/stock_market/gainers?apikey=%s", fmpAPIURL, s.config.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch gainers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var fmpMovers []FMPMarketMoverResponse
	if err := json.NewDecoder(resp.Body).Decode(&fmpMovers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to MarketMover (take top 10)
	limit := 10
	if len(fmpMovers) < limit {
		limit = len(fmpMovers)
	}

	gainers := make([]MarketMover, 0, limit)
	for i := 0; i < limit; i++ {
		item := fmpMovers[i]
		gainers = append(gainers, MarketMover{
			Symbol:        item.Symbol,
			Name:          item.Name,
			Change:        item.Change,
			ChangePercent: item.ChangesPercentage,
			Price:         item.Price,
			Volume:        item.Volume,
		})
	}

	// Cache for 5 minutes
	if s.config.EnableCache && s.redis != nil && len(gainers) > 0 {
		data, _ := json.Marshal(gainers)
		s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
		s.logger.Debug("Cached market gainers (TTL: 5m)")
	}

	return gainers, nil
}

// GetMarketLosers fetches top losing stocks with caching
func (s *Service) GetMarketLosers(ctx context.Context) ([]MarketMover, error) {
	cacheKey := cachePrefix + "losers"
	if s.config.EnableCache && s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var losers []MarketMover
			if err := json.Unmarshal([]byte(cached), &losers); err == nil {
				s.logger.Debug("Cache HIT for market losers")
				return losers, nil
			}
		}
	}

	if s.rateLimiter != nil {
		<-s.rateLimiter.C
	}

	url := fmt.Sprintf("%s/stock_market/losers?apikey=%s", fmpAPIURL, s.config.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch losers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var fmpMovers []FMPMarketMoverResponse
	if err := json.NewDecoder(resp.Body).Decode(&fmpMovers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	limit := 10
	if len(fmpMovers) < limit {
		limit = len(fmpMovers)
	}

	losers := make([]MarketMover, 0, limit)
	for i := 0; i < limit; i++ {
		item := fmpMovers[i]
		losers = append(losers, MarketMover{
			Symbol:        item.Symbol,
			Name:          item.Name,
			Change:        item.Change,
			ChangePercent: item.ChangesPercentage,
			Price:         item.Price,
			Volume:        item.Volume,
		})
	}

	if s.config.EnableCache && s.redis != nil && len(losers) > 0 {
		data, _ := json.Marshal(losers)
		s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
		s.logger.Debug("Cached market losers (TTL: 5m)")
	}

	return losers, nil
}

// GetMostActives fetches most actively traded stocks
func (s *Service) GetMostActives(ctx context.Context) ([]MarketMover, error) {
	cacheKey := cachePrefix + "actives"
	if s.config.EnableCache && s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var actives []MarketMover
			if err := json.Unmarshal([]byte(cached), &actives); err == nil {
				s.logger.Debug("Cache HIT for most actives")
				return actives, nil
			}
		}
	}

	if s.rateLimiter != nil {
		<-s.rateLimiter.C
	}

	url := fmt.Sprintf("%s/stock_market/actives?apikey=%s", fmpAPIURL, s.config.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch actives: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var fmpMovers []FMPMarketMoverResponse
	if err := json.NewDecoder(resp.Body).Decode(&fmpMovers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	limit := 10
	if len(fmpMovers) < limit {
		limit = len(fmpMovers)
	}

	actives := make([]MarketMover, 0, limit)
	for i := 0; i < limit; i++ {
		item := fmpMovers[i]
		actives = append(actives, MarketMover{
			Symbol:        item.Symbol,
			Name:          item.Name,
			Change:        item.Change,
			ChangePercent: item.ChangesPercentage,
			Price:         item.Price,
			Volume:        item.Volume,
		})
	}

	if s.config.EnableCache && s.redis != nil && len(actives) > 0 {
		data, _ := json.Marshal(actives)
		s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
		s.logger.Debug("Cached most actives (TTL: 5m)")
	}

	return actives, nil
}

// GetSectorPerformance fetches sector performance data
func (s *Service) GetSectorPerformance(ctx context.Context) ([]SectorPerformance, error) {
	cacheKey := cachePrefix + "sectors"
	if s.config.EnableCache && s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var sectors []SectorPerformance
			if err := json.Unmarshal([]byte(cached), &sectors); err == nil {
				s.logger.Debug("Cache HIT for sector performance")
				return sectors, nil
			}
		}
	}

	if s.rateLimiter != nil {
		<-s.rateLimiter.C
	}

	url := fmt.Sprintf("%s/sector-performance?apikey=%s", fmpAPIURL, s.config.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sector performance: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var fmpSectors []FMPSectorPerformanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&fmpSectors); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	sectors := make([]SectorPerformance, 0, len(fmpSectors))
	for _, item := range fmpSectors {
		// Parse percentage string (e.g., "1.23%" -> 1.23)
		var changePercent float64
		fmt.Sscanf(item.ChangesPercentage, "%f%%", &changePercent)

		sectors = append(sectors, SectorPerformance{
			Sector:        item.Sector,
			ChangePercent: changePercent,
		})
	}

	if s.config.EnableCache && s.redis != nil && len(sectors) > 0 {
		data, _ := json.Marshal(sectors)
		s.redis.Set(ctx, cacheKey, data, 15*time.Minute)
		s.logger.Debug("Cached sector performance (TTL: 15m)")
	}

	return sectors, nil
}

// GetAnalystRatings fetches analyst ratings for a symbol
func (s *Service) GetAnalystRatings(ctx context.Context, symbol string, limit int) ([]AnalystRating, error) {
	symbol = strings.ToUpper(symbol)

	if limit <= 0 || limit > 50 {
		limit = 20
	}

	cacheKey := fmt.Sprintf("%sratings:%s", cachePrefix, symbol)
	if s.config.EnableCache && s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var ratings []AnalystRating
			if err := json.Unmarshal([]byte(cached), &ratings); err == nil {
				s.logger.Debugf("Cache HIT for analyst ratings: %s", symbol)
				return ratings, nil
			}
		}
	}

	if s.rateLimiter != nil {
		<-s.rateLimiter.C
	}

	url := fmt.Sprintf("%s/grade/%s?limit=%d&apikey=%s", fmpAPIURL, symbol, limit, s.config.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ratings: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var fmpRatings []FMPAnalystRatingResponse
	if err := json.NewDecoder(resp.Body).Decode(&fmpRatings); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	ratings := make([]AnalystRating, 0, len(fmpRatings))
	for _, item := range fmpRatings {
		date, _ := time.Parse("2006-01-02 15:04:05", item.Date)
		ratings = append(ratings, AnalystRating{
			Symbol:          item.Symbol,
			Date:            date,
			AnalystName:     item.AnalystName,
			AnalystCompany:  item.AnalystCompany,
			GradeNew:        item.GradeNew,
			GradePrevious:   item.GradePrevious,
			Action:          item.Action,
			PriceTarget:     item.PriceTarget,
			PriceWhenPosted: item.PriceWhenPosted,
		})
	}

	if s.config.EnableCache && s.redis != nil && len(ratings) > 0 {
		data, _ := json.Marshal(ratings)
		s.redis.Set(ctx, cacheKey, data, 1*time.Hour)
		s.logger.Debugf("Cached analyst ratings: %s (TTL: 1h)", symbol)
	}

	return ratings, nil
}

// GetPriceTarget fetches analyst price target consensus
func (s *Service) GetPriceTarget(ctx context.Context, symbol string) (*PriceTarget, error) {
	symbol = strings.ToUpper(symbol)

	cacheKey := fmt.Sprintf("%starget:%s", cachePrefix, symbol)
	if s.config.EnableCache && s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var target PriceTarget
			if err := json.Unmarshal([]byte(cached), &target); err == nil {
				s.logger.Debugf("Cache HIT for price target: %s", symbol)
				return &target, nil
			}
		}
	}

	if s.rateLimiter != nil {
		<-s.rateLimiter.C
	}

	url := fmt.Sprintf("%s/price-target-consensus/%s?apikey=%s", fmpAPIURL, symbol, s.config.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch price target: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var fmpTarget FMPPriceTargetResponse
	if err := json.NewDecoder(resp.Body).Decode(&fmpTarget); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	target := &PriceTarget{
		Symbol:        fmpTarget.Symbol,
		TargetHigh:    fmpTarget.TargetHigh,
		TargetLow:     fmpTarget.TargetLow,
		TargetConsens: fmpTarget.TargetConsensus,
		TargetMedian:  fmpTarget.TargetMedian,
	}

	if s.config.EnableCache && s.redis != nil {
		data, _ := json.Marshal(target)
		s.redis.Set(ctx, cacheKey, data, 1*time.Hour)
		s.logger.Debugf("Cached price target: %s (TTL: 1h)", symbol)
	}

	return target, nil
}
