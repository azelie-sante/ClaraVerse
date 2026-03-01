package handlers

import (
	"bytes"
	"claraverse/internal/services"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestWebhookHandler_InvalidSignature(t *testing.T) {
	app := fiber.New()
	paymentService := services.NewPaymentService("", "secret123", "", nil, nil, nil)
	handler := NewWebhookHandler(paymentService)

	app.Post("/api/webhooks/dodo", handler.HandleDodoWebhook)

	payload := []byte(`{"type":"subscription.active"}`)
	req := httptest.NewRequest("POST", "/api/webhooks/dodo", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Webhook-Signature", "invalid_signature")

	resp, _ := app.Test(req)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", resp.StatusCode)
	}
}

func TestWebhookHandler_ValidSignature(t *testing.T) {
	secret := "webhook_secret_123"
	app := fiber.New()
	paymentService := services.NewPaymentService("", secret, "", nil, nil, nil)
	handler := NewWebhookHandler(paymentService)

	app.Post("/api/webhooks/dodo", handler.HandleDodoWebhook)

	payload := []byte(`{"type":"subscription.active","data":{"subscription_id":"sub_123"}}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest("POST", "/api/webhooks/dodo", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Webhook-Signature", signature)

	resp, _ := app.Test(req)

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
}

func TestWebhookHandler_MissingSignature(t *testing.T) {
	app := fiber.New()
	paymentService := services.NewPaymentService("", "secret", "", nil, nil, nil)
	handler := NewWebhookHandler(paymentService)

	app.Post("/api/webhooks/dodo", handler.HandleDodoWebhook)

	payload := []byte(`{"type":"subscription.active"}`)
	req := httptest.NewRequest("POST", "/api/webhooks/dodo", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	// No signature header

	resp, _ := app.Test(req)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", resp.StatusCode)
	}
}

func TestWebhookHandler_InvalidJSON(t *testing.T) {
	secret := "webhook_secret"
	app := fiber.New()
	paymentService := services.NewPaymentService("", secret, "", nil, nil, nil)
	handler := NewWebhookHandler(paymentService)

	app.Post("/api/webhooks/dodo", handler.HandleDodoWebhook)

	payload := []byte(`{invalid json}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest("POST", "/api/webhooks/dodo", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Webhook-Signature", signature)

	resp, _ := app.Test(req)

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.StatusCode)
	}
}

func TestWebhookHandler_AllEventTypes(t *testing.T) {
	secret := "webhook_secret"
	app := fiber.New()
	paymentService := services.NewPaymentService("", secret, "", nil, nil, nil)
	handler := NewWebhookHandler(paymentService)

	app.Post("/api/webhooks/dodo", handler.HandleDodoWebhook)

	eventTypes := []string{
		"subscription.active",
		"subscription.updated",
		"subscription.on_hold",
		"subscription.renewed",
		"subscription.cancelled",
		"payment.succeeded",
		"payment.failed",
	}

	for _, eventType := range eventTypes {
		t.Run(eventType, func(t *testing.T) {
			payload := []byte(`{"type":"` + eventType + `","data":{},"id":"evt_123"}`)

			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write(payload)
			signature := hex.EncodeToString(mac.Sum(nil))

			req := httptest.NewRequest("POST", "/api/webhooks/dodo", bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Webhook-Signature", signature)

			resp, _ := app.Test(req)

			if resp.StatusCode != fiber.StatusOK {
				t.Errorf("Event %s: expected 200, got %d", eventType, resp.StatusCode)
			}
		})
	}
}

func TestWebhookHandler_AlternativeSignatureHeader(t *testing.T) {
	secret := "webhook_secret"
	app := fiber.New()
	paymentService := services.NewPaymentService("", secret, "", nil, nil, nil)
	handler := NewWebhookHandler(paymentService)

	app.Post("/api/webhooks/dodo", handler.HandleDodoWebhook)

	payload := []byte(`{"type":"subscription.active","data":{},"id":"evt_123"}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest("POST", "/api/webhooks/dodo", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Dodo-Signature", signature) // Alternative header name

	resp, _ := app.Test(req)

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200 with Dodo-Signature header, got %d", resp.StatusCode)
	}
}

