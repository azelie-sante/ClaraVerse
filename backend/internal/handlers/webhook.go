package handlers

import (
	"claraverse/internal/services"
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// WebhookHandler handles DodoPayments webhooks
type WebhookHandler struct {
	paymentService *services.PaymentService
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(paymentService *services.PaymentService) *WebhookHandler {
	return &WebhookHandler{
		paymentService: paymentService,
	}
}

// HandleDodoWebhook handles incoming webhooks from DodoPayments
// POST /api/webhooks/dodo
// DodoPayments uses Standard Webhooks format with headers:
// - webhook-id: unique message ID
// - webhook-signature: v1,<base64_signature>
// - webhook-timestamp: unix timestamp
func (h *WebhookHandler) HandleDodoWebhook(c *fiber.Ctx) error {
	// Get payload
	payload := c.Body()
	if len(payload) == 0 {
		log.Printf("❌ Webhook missing payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing payload",
		})
	}

	// Convert Fiber headers to http.Header for SDK
	headers := make(http.Header)
	c.Request().Header.VisitAll(func(key, value []byte) {
		headers.Add(string(key), string(value))
	})

	// Verify and parse webhook using SDK
	event, err := h.paymentService.VerifyAndParseWebhook(payload, headers)
	if err != nil {
		log.Printf("❌ Webhook verification failed: %v", err)

		// Distinguish between parse errors (400) and auth errors (401)
		if strings.Contains(err.Error(), "parse") || strings.Contains(err.Error(), "invalid character") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid payload format: " + err.Error(),
			})
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid webhook: " + err.Error(),
		})
	}

	// Handle event
	ctx := context.Background()
	if err := h.paymentService.HandleWebhookEvent(ctx, event); err != nil {
		log.Printf("❌ Webhook processing error: %v", err)

		// Return 200 for idempotency errors (duplicate events) - no retry needed
		if strings.Contains(err.Error(), "already processed") ||
		   strings.Contains(err.Error(), "duplicate") {
			log.Printf("⚠️  Duplicate webhook event %s (ID: %s) - already processed", event.Type, event.ID)
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"received": true,
				"message":  "Event already processed (idempotent)",
			})
		}

		// Return 500 for actual processing failures to allow DodoPayments to retry
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Failed to process webhook",
			"event_id": event.ID,
			"type":     event.Type,
		})
	}

	log.Printf("✅ Webhook event processed: %s (ID: %s)", event.Type, event.ID)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"received": true,
		"event_id": event.ID,
		"type":     event.Type,
	})
}

