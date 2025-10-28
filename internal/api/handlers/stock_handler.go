package handlers

import (
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

	if len(request.Symbols) > 20 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Maximum 20 symbols allowed per request",
		})
	}

	quotes, err := h.stockService.GetMultipleQuotes(c.Context(), request.Symbols)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get multiple quotes")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch stock quotes",
		})
	}

	return c.JSON(quotes)
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

// GetArticlesByTicker handles GET /api/v1/articles/by-ticker/:symbol
// This is implemented in ai_handler.go as it uses the AI service
