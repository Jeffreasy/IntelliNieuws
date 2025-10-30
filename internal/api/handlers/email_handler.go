package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jeffrey/intellinieuws/internal/email"
	"github.com/jeffrey/intellinieuws/pkg/logger"
)

// EmailHandler handles email-related HTTP requests
type EmailHandler struct {
	emailProcessor *email.Processor
	logger         *logger.Logger
}

// NewEmailHandler creates a new email handler
func NewEmailHandler(emailProcessor *email.Processor, log *logger.Logger) *EmailHandler {
	return &EmailHandler{
		emailProcessor: emailProcessor,
		logger:         log.WithComponent("email-handler"),
	}
}

// FetchExistingEmails handles POST /api/v1/email/fetch-existing
func (h *EmailHandler) FetchExistingEmails(c *fiber.Ctx) error {
	h.logger.Info("Received request to fetch existing emails")

	articlesCreated, err := h.emailProcessor.FetchExistingEmails(c.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch existing emails")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch existing emails",
		})
	}

	h.logger.Infof("Successfully fetched %d existing emails", articlesCreated)

	return c.JSON(fiber.Map{
		"message":          "Existing emails fetched successfully",
		"articles_created": articlesCreated,
		"status":           "completed",
	})
}

// GetStats handles GET /api/v1/email/stats
func (h *EmailHandler) GetStats(c *fiber.Ctx) error {
	stats, err := h.emailProcessor.GetStats(c.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get email stats")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch email statistics",
		})
	}

	return c.JSON(fiber.Map{
		"stats": stats,
	})
}
