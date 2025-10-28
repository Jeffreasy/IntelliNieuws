package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jeffrey/intellinieuws/internal/stock"
	"github.com/jeffrey/intellinieuws/pkg/logger"
)

// StockHandler handles stock-related HTTP requests
type StockHandler struct {
	stockService *stock.Service
	logger       *logger.Logger
}

// NewStockHandler creates a new stock handler
func NewStockHandler(stockService *stock.Service, log *logger.Logger) *StockHandler {
	return &StockHandler{
		stockService: stockService,
		logger:       log.WithComponent("stock-handler"),
	}
}

// GetQuote handles GET /api/v1/stocks/quote/:symbol
func (h *StockHandler) GetQuote(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Symbol parameter is required",
		})
	}

	quote, err := h.stockService.GetQuote(c.Context(), symbol)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to get quote for %s", symbol)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch stock quote",
		})
	}

	return c.JSON(quote)
}

// GetMultipleQuotes handles POST /api/v1/stocks/quotes
// Now supports up to 100 symbols per request thanks to FMP batch API optimization
func (h *StockHandler) GetMultipleQuotes(c *fiber.Ctx) error {
	var request struct {
		Symbols []string `json:"symbols"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(request.Symbols) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Symbols array is required and cannot be empty",
		})
	}

	// Increased from 20 to 100 thanks to batch API optimization (1 API call instead of N)
	if len(request.Symbols) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Maximum 100 symbols allowed per request",
		})
	}

	startTime := c.Context().Time()
	quotes, err := h.stockService.GetMultipleQuotes(c.Context(), request.Symbols)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get multiple quotes")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch stock quotes",
		})
	}

	// Log performance metrics
	duration := c.Context().Time().Sub(startTime)
	h.logger.Infof("Batch quotes: %d symbols fetched in %v (%.2f symbols/sec)",
		len(quotes), duration, float64(len(quotes))/duration.Seconds())

	return c.JSON(fiber.Map{
		"quotes": quotes,
		"meta": fiber.Map{
			"total":       len(quotes),
			"requested":   len(request.Symbols),
			"duration_ms": duration.Milliseconds(),
			"using_batch": true,
			"cost_saving": fmt.Sprintf("%.0f%%", float64(len(request.Symbols)-1)/float64(len(request.Symbols))*100),
		},
	})
}

// GetProfile handles GET /api/v1/stocks/profile/:symbol
func (h *StockHandler) GetProfile(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Symbol parameter is required",
		})
	}

	profile, err := h.stockService.GetProfile(c.Context(), symbol)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to get profile for %s", symbol)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch company profile",
		})
	}

	return c.JSON(profile)
}

// GetStats handles GET /api/v1/stocks/stats
func (h *StockHandler) GetStats(c *fiber.Ctx) error {
	stats := h.stockService.GetCacheStats(c.Context())
	return c.JSON(fiber.Map{
		"cache": stats,
	})
}

// GetStockNews handles GET /api/v1/stocks/news/:symbol
func (h *StockHandler) GetStockNews(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Symbol parameter is required",
		})
	}

	limit := c.QueryInt("limit", 10)
	if limit > 50 {
		limit = 50
	}

	news, err := h.stockService.GetStockNews(c.Context(), symbol, limit)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to get stock news for %s", symbol)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch stock news",
		})
	}

	return c.JSON(fiber.Map{
		"symbol": symbol,
		"news":   news,
		"total":  len(news),
	})
}

// GetHistoricalPrices handles GET /api/v1/stocks/historical/:symbol
func (h *StockHandler) GetHistoricalPrices(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Symbol parameter is required",
		})
	}

	// Parse date parameters (default: last 30 days)
	toDate := time.Now()
	fromDate := toDate.AddDate(0, 0, -30)

	if fromStr := c.Query("from"); fromStr != "" {
		if parsed, err := time.Parse("2006-01-02", fromStr); err == nil {
			fromDate = parsed
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		if parsed, err := time.Parse("2006-01-02", toStr); err == nil {
			toDate = parsed
		}
	}

	prices, err := h.stockService.GetHistoricalPrices(c.Context(), symbol, fromDate, toDate)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to get historical prices for %s", symbol)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch historical prices",
		})
	}

	return c.JSON(fiber.Map{
		"symbol":     symbol,
		"from":       fromDate.Format("2006-01-02"),
		"to":         toDate.Format("2006-01-02"),
		"prices":     prices,
		"dataPoints": len(prices),
	})
}

// GetKeyMetrics handles GET /api/v1/stocks/metrics/:symbol
func (h *StockHandler) GetKeyMetrics(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Symbol parameter is required",
		})
	}

	metrics, err := h.stockService.GetKeyMetrics(c.Context(), symbol)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to get key metrics for %s", symbol)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch key metrics",
		})
	}

	return c.JSON(metrics)
}

// GetEarningsCalendar handles GET /api/v1/stocks/earnings
func (h *StockHandler) GetEarningsCalendar(c *fiber.Ctx) error {
	// Parse date parameters (default: next 7 days)
	fromDate := time.Now()
	toDate := fromDate.AddDate(0, 0, 7)

	if fromStr := c.Query("from"); fromStr != "" {
		if parsed, err := time.Parse("2006-01-02", fromStr); err == nil {
			fromDate = parsed
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		if parsed, err := time.Parse("2006-01-02", toStr); err == nil {
			toDate = parsed
		}
	}

	calendar, err := h.stockService.GetEarningsCalendar(c.Context(), fromDate, toDate)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get earnings calendar")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch earnings calendar",
		})
	}

	return c.JSON(fiber.Map{
		"from":     fromDate.Format("2006-01-02"),
		"to":       toDate.Format("2006-01-02"),
		"earnings": calendar,
		"total":    len(calendar),
	})
}

// SearchSymbol handles GET /api/v1/stocks/search
func (h *StockHandler) SearchSymbol(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Query parameter 'q' is required",
		})
	}

	limit := c.QueryInt("limit", 10)
	if limit > 50 {
		limit = 50
	}

	results, err := h.stockService.SearchSymbol(c.Context(), query, limit)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to search symbols with query: %s", query)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search symbols",
		})
	}

	return c.JSON(fiber.Map{
		"query":   query,
		"results": results,
		"total":   len(results),
	})
}

// GetArticlesByTicker handles GET /api/v1/articles/by-ticker/:symbol
// This is implemented in ai_handler.go as it uses the AI service

// GetMarketGainers handles GET /api/v1/stocks/market/gainers
func (h *StockHandler) GetMarketGainers(c *fiber.Ctx) error {
	gainers, err := h.stockService.GetMarketGainers(c.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get market gainers")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch market gainers",
		})
	}

	return c.JSON(fiber.Map{
		"gainers": gainers,
		"total":   len(gainers),
	})
}

// GetMarketLosers handles GET /api/v1/stocks/market/losers
func (h *StockHandler) GetMarketLosers(c *fiber.Ctx) error {
	losers, err := h.stockService.GetMarketLosers(c.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get market losers")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch market losers",
		})
	}

	return c.JSON(fiber.Map{
		"losers": losers,
		"total":  len(losers),
	})
}

// GetMostActives handles GET /api/v1/stocks/market/actives
func (h *StockHandler) GetMostActives(c *fiber.Ctx) error {
	actives, err := h.stockService.GetMostActives(c.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get most actives")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch most actives",
		})
	}

	return c.JSON(fiber.Map{
		"actives": actives,
		"total":   len(actives),
	})
}

// GetSectorPerformance handles GET /api/v1/stocks/sectors
func (h *StockHandler) GetSectorPerformance(c *fiber.Ctx) error {
	sectors, err := h.stockService.GetSectorPerformance(c.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get sector performance")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch sector performance",
		})
	}

	return c.JSON(fiber.Map{
		"sectors": sectors,
		"total":   len(sectors),
	})
}

// GetAnalystRatings handles GET /api/v1/stocks/ratings/:symbol
func (h *StockHandler) GetAnalystRatings(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Symbol parameter is required",
		})
	}

	limit := c.QueryInt("limit", 20)
	if limit > 50 {
		limit = 50
	}

	ratings, err := h.stockService.GetAnalystRatings(c.Context(), symbol, limit)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to get analyst ratings for %s", symbol)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch analyst ratings",
		})
	}

	return c.JSON(fiber.Map{
		"symbol":  symbol,
		"ratings": ratings,
		"total":   len(ratings),
	})
}

// GetPriceTarget handles GET /api/v1/stocks/target/:symbol
func (h *StockHandler) GetPriceTarget(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Symbol parameter is required",
		})
	}

	target, err := h.stockService.GetPriceTarget(c.Context(), symbol)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to get price target for %s", symbol)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch price target",
		})
	}

	return c.JSON(target)
}
