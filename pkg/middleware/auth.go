package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// APIKeyAuth middleware for API key authentication
type APIKeyAuth struct {
	apiKey       string
	headerName   string
	errorMessage string
}

// NewAPIKeyAuth creates a new API key authentication middleware
func NewAPIKeyAuth(apiKey, headerName string) *APIKeyAuth {
	if headerName == "" {
		headerName = "X-API-Key"
	}

	return &APIKeyAuth{
		apiKey:       apiKey,
		headerName:   headerName,
		errorMessage: "Invalid or missing API key",
	}
}

// Handler returns the Fiber middleware handler
func (a *APIKeyAuth) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip if no API key is configured
		if a.apiKey == "" {
			return c.Next()
		}

		// Get API key from header
		providedKey := c.Get(a.headerName)

		// Check if API key matches
		if providedKey != a.apiKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": a.errorMessage,
			})
		}

		return c.Next()
	}
}

// Optional returns a middleware that allows requests without API key
func (a *APIKeyAuth) Optional() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get API key from header
		providedKey := c.Get(a.headerName)

		// If key is provided, validate it
		if providedKey != "" && providedKey != a.apiKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid API key",
			})
		}

		// Set authenticated flag in context
		if providedKey == a.apiKey {
			c.Locals("authenticated", true)
		}

		return c.Next()
	}
}
