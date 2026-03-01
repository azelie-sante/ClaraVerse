package handlers

import (
	"bytes"
	"claraverse/internal/models"
	"claraverse/internal/services"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func setupSubscriptionTestApp(t *testing.T) (*fiber.App, *services.PaymentService) {
	app := fiber.New()
	paymentService := services.NewPaymentService("test", "secret", "biz", nil, nil, nil)
	return app, paymentService
}

func mockSubscriptionAuthMiddleware(userID string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		c.Locals("user_email", "test@example.com")
		return c.Next()
	}
}

func TestSubscriptionHandler_ListPlans(t *testing.T) {
	app, paymentService := setupSubscriptionTestApp(t)
	handler := NewSubscriptionHandler(paymentService)

	app.Get("/api/subscriptions/plans", handler.ListPlans)

	req := httptest.NewRequest("GET", "/api/subscriptions/plans", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Plans []models.Plan `json:"plans"`
	}
	json.Unmarshal(body, &result)

	if len(result.Plans) < 4 {
		t.Errorf("Expected at least 4 plans, got %d", len(result.Plans))
	}
}

func TestSubscriptionHandler_GetCurrent_Unauthenticated(t *testing.T) {
	app, paymentService := setupSubscriptionTestApp(t)
	handler := NewSubscriptionHandler(paymentService)

	app.Get("/api/subscriptions/current", handler.GetCurrent)

	req := httptest.NewRequest("GET", "/api/subscriptions/current", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", resp.StatusCode)
	}
}

func TestSubscriptionHandler_GetCurrent_Authenticated(t *testing.T) {
	app, paymentService := setupSubscriptionTestApp(t)
	handler := NewSubscriptionHandler(paymentService)

	app.Use(mockAuthMiddleware("user-123"))
	app.Get("/api/subscriptions/current", handler.GetCurrent)

	req := httptest.NewRequest("GET", "/api/subscriptions/current", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	// Without MongoDB, expect default free tier response
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Tier   string `json:"tier"`
		Status string `json:"status"`
	}
	json.Unmarshal(body, &result)

	if result.Tier != models.TierFree {
		t.Errorf("Expected free tier, got %s", result.Tier)
	}
}

func TestSubscriptionHandler_CreateCheckout_InvalidPlan(t *testing.T) {
	app, paymentService := setupSubscriptionTestApp(t)
	handler := NewSubscriptionHandler(paymentService)

	app.Use(mockAuthMiddleware("user-123"))
	app.Post("/api/subscriptions/checkout", handler.CreateCheckout)

	reqBody := bytes.NewBuffer([]byte(`{"plan_id": "invalid_plan"}`))
	req := httptest.NewRequest("POST", "/api/subscriptions/checkout", reqBody)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("Expected 400 for invalid plan, got %d", resp.StatusCode)
	}
}

func TestSubscriptionHandler_CreateCheckout_FreePlan(t *testing.T) {
	app, paymentService := setupSubscriptionTestApp(t)
	handler := NewSubscriptionHandler(paymentService)

	app.Use(mockAuthMiddleware("user-123"))
	app.Post("/api/subscriptions/checkout", handler.CreateCheckout)

	reqBody := bytes.NewBuffer([]byte(`{"plan_id": "free"}`))
	req := httptest.NewRequest("POST", "/api/subscriptions/checkout", reqBody)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	// Should reject - can't checkout for free plan
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("Expected 400 for free plan checkout, got %d", resp.StatusCode)
	}
}

func TestSubscriptionHandler_CreateCheckout_EnterprisePlan(t *testing.T) {
	app, paymentService := setupSubscriptionTestApp(t)
	handler := NewSubscriptionHandler(paymentService)

	app.Use(mockAuthMiddleware("user-123"))
	app.Post("/api/subscriptions/checkout", handler.CreateCheckout)

	reqBody := bytes.NewBuffer([]byte(`{"plan_id": "enterprise"}`))
	req := httptest.NewRequest("POST", "/api/subscriptions/checkout", reqBody)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	// Should reject - enterprise requires contact sales
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("Expected 400 for enterprise checkout, got %d", resp.StatusCode)
	}
}

func TestSubscriptionHandler_CreateCheckout_MissingPlanID(t *testing.T) {
	app, paymentService := setupSubscriptionTestApp(t)
	handler := NewSubscriptionHandler(paymentService)

	app.Use(mockAuthMiddleware("user-123"))
	app.Post("/api/subscriptions/checkout", handler.CreateCheckout)

	reqBody := bytes.NewBuffer([]byte(`{}`))
	req := httptest.NewRequest("POST", "/api/subscriptions/checkout", reqBody)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("Expected 400 for missing plan_id, got %d", resp.StatusCode)
	}
}

func TestSubscriptionHandler_PreviewPlanChange(t *testing.T) {
	app, paymentService := setupSubscriptionTestApp(t)
	handler := NewSubscriptionHandler(paymentService)

	app.Use(mockAuthMiddleware("user-123"))
	app.Get("/api/subscriptions/change-plan/preview", handler.PreviewPlanChange)

	tests := []struct {
		name       string
		planID     string
		expectCode int
	}{
		{"valid upgrade", "pro", fiber.StatusOK},
		{"invalid plan", "invalid", fiber.StatusBadRequest},
		{"missing plan", "", fiber.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/subscriptions/change-plan/preview"
			if tt.planID != "" {
				url += "?plan_id=" + tt.planID
			}

			req := httptest.NewRequest("GET", url, nil)
			resp, _ := app.Test(req)

			if resp.StatusCode != tt.expectCode {
				t.Errorf("Expected %d, got %d", tt.expectCode, resp.StatusCode)
			}
		})
	}
}

func TestSubscriptionHandler_Cancel(t *testing.T) {
	app, paymentService := setupSubscriptionTestApp(t)
	handler := NewSubscriptionHandler(paymentService)

	app.Use(mockAuthMiddleware("user-123"))
	app.Post("/api/subscriptions/cancel", handler.Cancel)

	req := httptest.NewRequest("POST", "/api/subscriptions/cancel", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	// Without active subscription, should handle gracefully
	// Could be 200 (already free) or 400 (nothing to cancel)
	if resp.StatusCode != fiber.StatusOK && resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("Unexpected status: %d", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Reactivate_NoActiveCancellation(t *testing.T) {
	app, paymentService := setupSubscriptionTestApp(t)
	handler := NewSubscriptionHandler(paymentService)

	app.Use(mockAuthMiddleware("user-123"))
	app.Post("/api/subscriptions/reactivate", handler.Reactivate)

	req := httptest.NewRequest("POST", "/api/subscriptions/reactivate", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	// Without pending cancellation, should fail
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.StatusCode)
	}
}

func TestSubscriptionHandler_GetPortalURL_Unauthenticated(t *testing.T) {
	app, paymentService := setupSubscriptionTestApp(t)
	handler := NewSubscriptionHandler(paymentService)

	app.Get("/api/subscriptions/portal", handler.GetPortalURL)

	req := httptest.NewRequest("GET", "/api/subscriptions/portal", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", resp.StatusCode)
	}
}

func TestSubscriptionHandler_ListInvoices(t *testing.T) {
	app, paymentService := setupSubscriptionTestApp(t)
	handler := NewSubscriptionHandler(paymentService)

	app.Use(mockAuthMiddleware("user-123"))
	app.Get("/api/subscriptions/invoices", handler.ListInvoices)

	req := httptest.NewRequest("GET", "/api/subscriptions/invoices", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
}

