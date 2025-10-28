package stock

import "time"

// StockQuote represents a stock quote from the API
type StockQuote struct {
	Symbol            string    `json:"symbol"`
	Name              string    `json:"name"`
	Price             float64   `json:"price"`
	Change            float64   `json:"change"`
	ChangePercent     float64   `json:"change_percent"`
	Volume            int64     `json:"volume"`
	MarketCap         int64     `json:"market_cap,omitempty"`
	Exchange          string    `json:"exchange"`
	Currency          string    `json:"currency"`
	LastUpdated       time.Time `json:"last_updated"`
	PreviousClose     float64   `json:"previous_close,omitempty"`
	Open              float64   `json:"open,omitempty"`
	DayHigh           float64   `json:"day_high,omitempty"`
	DayLow            float64   `json:"day_low,omitempty"`
	YearHigh          float64   `json:"year_high,omitempty"`
	YearLow           float64   `json:"year_low,omitempty"`
	PriceAvg50        float64   `json:"price_avg_50,omitempty"`
	PriceAvg200       float64   `json:"price_avg_200,omitempty"`
	EPS               float64   `json:"eps,omitempty"`
	PE                float64   `json:"pe,omitempty"`
	SharesOutstanding int64     `json:"shares_outstanding,omitempty"`
}

// StockProfile represents company profile information
type StockProfile struct {
	Symbol      string `json:"symbol"`
	CompanyName string `json:"company_name"`
	Currency    string `json:"currency"`
	Exchange    string `json:"exchange"`
	Industry    string `json:"industry,omitempty"`
	Sector      string `json:"sector,omitempty"`
	Website     string `json:"website,omitempty"`
	Description string `json:"description,omitempty"`
	CEO         string `json:"ceo,omitempty"`
	Country     string `json:"country,omitempty"`
	IPODate     string `json:"ipo_date,omitempty"`
}

// Config holds stock service configuration
type Config struct {
	APIKey          string
	APIProvider     string // "fmp" or "alphavantage"
	CacheTTL        time.Duration
	RateLimitPerMin int
	Timeout         time.Duration
	EnableCache     bool
}

// FMPQuoteResponse represents the response from Financial Modeling Prep API
type FMPQuoteResponse struct {
	Symbol            string  `json:"symbol"`
	Name              string  `json:"name"`
	Price             float64 `json:"price"`
	ChangesPercentage float64 `json:"changesPercentage"`
	Change            float64 `json:"change"`
	DayLow            float64 `json:"dayLow"`
	DayHigh           float64 `json:"dayHigh"`
	YearHigh          float64 `json:"yearHigh"`
	YearLow           float64 `json:"yearLow"`
	MarketCap         int64   `json:"marketCap"`
	PriceAvg50        float64 `json:"priceAvg50"`
	PriceAvg200       float64 `json:"priceAvg200"`
	Exchange          string  `json:"exchange"`
	Volume            int64   `json:"volume"`
	AvgVolume         int64   `json:"avgVolume"`
	Open              float64 `json:"open"`
	PreviousClose     float64 `json:"previousClose"`
	EPS               float64 `json:"eps"`
	PE                float64 `json:"pe"`
	SharesOutstanding int64   `json:"sharesOutstanding"`
	Timestamp         int64   `json:"timestamp"`
}

// FMPProfileResponse represents company profile from FMP
type FMPProfileResponse struct {
	Symbol      string  `json:"symbol"`
	CompanyName string  `json:"companyName"`
	Currency    string  `json:"currency"`
	Exchange    string  `json:"exchangeShortName"`
	Industry    string  `json:"industry"`
	Sector      string  `json:"sector"`
	Website     string  `json:"website"`
	Description string  `json:"description"`
	CEO         string  `json:"ceo"`
	Country     string  `json:"country"`
	IPODate     string  `json:"ipoDate"`
	Price       float64 `json:"price"`
	Beta        float64 `json:"beta"`
	MarketCap   int64   `json:"mktCap"`
}

// AlphaVantageQuoteResponse represents response from Alpha Vantage
type AlphaVantageQuoteResponse struct {
	GlobalQuote struct {
		Symbol           string `json:"01. symbol"`
		Open             string `json:"02. open"`
		High             string `json:"03. high"`
		Low              string `json:"04. low"`
		Price            string `json:"05. price"`
		Volume           string `json:"06. volume"`
		LatestTradingDay string `json:"07. latest trading day"`
		PreviousClose    string `json:"08. previous close"`
		Change           string `json:"09. change"`
		ChangePercent    string `json:"10. change percent"`
	} `json:"Global Quote"`
}
