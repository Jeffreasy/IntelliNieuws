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

// StockNews represents a news article from FMP
type StockNews struct {
	Symbol        string    `json:"symbol"`
	PublishedDate time.Time `json:"publishedDate"`
	Title         string    `json:"title"`
	Image         string    `json:"image"`
	Site          string    `json:"site"`
	Text          string    `json:"text"`
	URL           string    `json:"url"`
}

// FMPNewsResponse represents the response from FMP news API
type FMPNewsResponse struct {
	Symbol        string `json:"symbol"`
	PublishedDate string `json:"publishedDate"`
	Title         string `json:"title"`
	Image         string `json:"image"`
	Site          string `json:"site"`
	Text          string `json:"text"`
	URL           string `json:"url"`
}

// HistoricalPrice represents historical price data
type HistoricalPrice struct {
	Date          time.Time `json:"date"`
	Open          float64   `json:"open"`
	High          float64   `json:"high"`
	Low           float64   `json:"low"`
	Close         float64   `json:"close"`
	AdjClose      float64   `json:"adjClose"`
	Volume        int64     `json:"volume"`
	Change        float64   `json:"change"`
	ChangePercent float64   `json:"changePercent"`
}

// FMPHistoricalResponse represents FMP historical price response
type FMPHistoricalResponse struct {
	Symbol     string                   `json:"symbol"`
	Historical []FMPHistoricalDataPoint `json:"historical"`
}

type FMPHistoricalDataPoint struct {
	Date          string  `json:"date"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
	AdjClose      float64 `json:"adjClose"`
	Volume        int64   `json:"volume"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"changePercent"`
}

// KeyMetrics represents key financial metrics
type KeyMetrics struct {
	Symbol            string  `json:"symbol"`
	MarketCap         int64   `json:"marketCap"`
	PE                float64 `json:"peRatio"`
	PEG               float64 `json:"pegRatio"`
	PB                float64 `json:"priceToBook"`
	PS                float64 `json:"priceToSales"`
	ROE               float64 `json:"roe"`
	ROA               float64 `json:"roa"`
	DebtToEquity      float64 `json:"debtToEquity"`
	CurrentRatio      float64 `json:"currentRatio"`
	DividendYield     float64 `json:"dividendYield"`
	EPS               float64 `json:"eps"`
	RevenuePerShare   float64 `json:"revenuePerShare"`
	FreeCashFlowYield float64 `json:"freeCashFlowYield"`
}

// FMPKeyMetricsResponse represents FMP key metrics response
type FMPKeyMetricsResponse struct {
	Symbol            string  `json:"symbol"`
	Date              string  `json:"date"`
	Period            string  `json:"period"`
	MarketCap         int64   `json:"marketCap"`
	PeRatio           float64 `json:"peRatio"`
	PegRatio          float64 `json:"pegRatio"`
	PriceToBookRatio  float64 `json:"priceToBookRatio"`
	PriceToSalesRatio float64 `json:"priceToSalesRatio"`
	Roe               float64 `json:"roe"`
	Roa               float64 `json:"roa"`
	DebtToEquity      float64 `json:"debtToEquity"`
	CurrentRatio      float64 `json:"currentRatio"`
	DividendYield     float64 `json:"dividendYield"`
	Eps               float64 `json:"eps"`
	RevenuePerShare   float64 `json:"revenuePerShare"`
	FreeCashFlowYield float64 `json:"freeCashFlowYield"`
}

// EarningsCalendar represents an earnings announcement
type EarningsCalendar struct {
	Symbol       string    `json:"symbol"`
	Date         time.Time `json:"date"`
	EPS          float64   `json:"eps"`
	EPSEstimated float64   `json:"epsEstimated"`
	Time         string    `json:"time"`
	Revenue      int64     `json:"revenue"`
	RevenueEst   int64     `json:"revenueEstimated"`
}

// FMPEarningsCalendarResponse represents FMP earnings calendar response
type FMPEarningsCalendarResponse struct {
	Symbol           string  `json:"symbol"`
	Date             string  `json:"date"`
	Eps              float64 `json:"eps"`
	EpsEstimated     float64 `json:"epsEstimated"`
	Time             string  `json:"time"`
	Revenue          int64   `json:"revenue"`
	RevenueEstimated int64   `json:"revenueEstimated"`
}

// MarketMover represents top gaining/losing stocks
type MarketMover struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"changePercent"`
	Price         float64 `json:"price"`
	Volume        int64   `json:"volume"`
}

// FMPMarketMoverResponse represents FMP gainers/losers response
type FMPMarketMoverResponse struct {
	Symbol            string  `json:"symbol"`
	Name              string  `json:"name"`
	Change            float64 `json:"change"`
	ChangesPercentage float64 `json:"changesPercentage"`
	Price             float64 `json:"price"`
	Volume            int64   `json:"volume"`
}

// SectorPerformance represents sector performance data
type SectorPerformance struct {
	Sector        string  `json:"sector"`
	ChangePercent float64 `json:"changePercent"`
}

// FMPSectorPerformanceResponse represents FMP sector performance response
type FMPSectorPerformanceResponse struct {
	Sector            string `json:"sector"`
	ChangesPercentage string `json:"changesPercentage"`
}

// AnalystRating represents analyst ratings
type AnalystRating struct {
	Symbol          string    `json:"symbol"`
	Date            time.Time `json:"date"`
	AnalystName     string    `json:"analystName"`
	AnalystCompany  string    `json:"analystCompany"`
	GradeNew        string    `json:"gradeNew"`
	GradePrevious   string    `json:"gradePrevious"`
	Action          string    `json:"action"`
	PriceTarget     float64   `json:"priceTarget"`
	PriceWhenPosted float64   `json:"priceWhenPosted"`
}

// FMPAnalystRatingResponse represents FMP analyst rating response
type FMPAnalystRatingResponse struct {
	Symbol          string  `json:"symbol"`
	Date            string  `json:"date"`
	AnalystName     string  `json:"analystName"`
	AnalystCompany  string  `json:"analystCompany"`
	GradeNew        string  `json:"newGrade"`
	GradePrevious   string  `json:"previousGrade"`
	Action          string  `json:"gradeAction"`
	PriceTarget     float64 `json:"priceTarget"`
	PriceWhenPosted float64 `json:"priceWhenPosted"`
}

// PriceTarget represents price target consensus
type PriceTarget struct {
	Symbol        string  `json:"symbol"`
	TargetHigh    float64 `json:"targetHigh"`
	TargetLow     float64 `json:"targetLow"`
	TargetConsens float64 `json:"targetConsensus"`
	TargetMedian  float64 `json:"targetMedian"`
}

// FMPPriceTargetResponse represents FMP price target response
type FMPPriceTargetResponse struct {
	Symbol          string  `json:"symbol"`
	TargetHigh      float64 `json:"targetHigh"`
	TargetLow       float64 `json:"targetLow"`
	TargetConsensus float64 `json:"targetConsensus"`
	TargetMedian    float64 `json:"targetMedian"`
}
