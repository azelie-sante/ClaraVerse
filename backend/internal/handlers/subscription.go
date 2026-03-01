package handlers

import (
	"claraverse/internal/services"
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
)

// SubscriptionHandler handles subscription-related endpoints
type SubscriptionHandler struct {
	paymentService *services.PaymentService
	userService    *services.UserService
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(paymentService *services.PaymentService, userService *services.UserService) *SubscriptionHandler {
	return &SubscriptionHandler{
		paymentService: paymentService,
		userService:    userService,
	}
}

// ListPlans returns all available subscription plans
// GET /api/subscriptions/plans
func (h *SubscriptionHandler) ListPlans(c *fiber.Ctx) error {
	plans := h.paymentService.GetAvailablePlans()
	return c.JSON(fiber.Map{
		"plans": plans,
	})
}

// GetCurrent returns the user's current subscription
// GET /api/subscriptions/current
func (h *SubscriptionHandler) GetCurrent(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Get email from auth middleware
	email, _ := c.Locals("user_email").(string)

	ctx := context.Background()

	// Sync user from Supabase (creates user if not exists, applies promo)
	user, err := h.userService.SyncUserFromSupabase(ctx, userID, email)
	if err != nil {
		log.Printf("⚠️  Failed to sync user %s: %v", userID, err)
	}

	sub, err := h.paymentService.GetCurrentSubscription(ctx, userID)
	if err != nil {
		log.Printf("⚠️  Failed to get subscription for user %s: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get subscription",
		})
	}

	// Get user data for promo detection and welcome popup status
	isPromoUser := false
	hasSeenWelcomePopup := false

	if user != nil {
		hasSeenWelcomePopup = user.HasSeenWelcomePopup
		// Promo user = PRO tier + has expiration + no Dodo subscription
		isPromoUser = user.SubscriptionTier == "pro" &&
			user.SubscriptionExpiresAt != nil &&
			user.DodoSubscriptionID == ""
	}

	// Build complete subscription response
	response := fiber.Map{
		"id":                   sub.ID.Hex(),
		"user_id":              sub.UserID,
		"tier":                 sub.Tier,
		"status":               sub.Status,
		"current_period_start": sub.CurrentPeriodStart,
		"current_period_end":   sub.CurrentPeriodEnd,
		"cancel_at_period_end": sub.CancelAtPeriodEnd,
		"is_promo_user":        isPromoUser,
		"has_seen_welcome_popup": hasSeenWelcomePopup,
		"created_at":           sub.CreatedAt,
		"updated_at":           sub.UpdatedAt,
	}

	// Add optional Dodo fields if present
	if sub.DodoSubscriptionID != "" {
		response["dodo_subscription_id"] = sub.DodoSubscriptionID
	}
	if sub.DodoCustomerID != "" {
		response["dodo_customer_id"] = sub.DodoCustomerID
	}

	// Add subscription expiration (for promo users)
	if user != nil && user.SubscriptionExpiresAt != nil {
		response["subscription_expires_at"] = user.SubscriptionExpiresAt.Format("2006-01-02T15:04:05Z07:00")
	}

	// Add scheduled change info if exists
	if sub.HasScheduledChange() {
		response["scheduled_tier"] = sub.ScheduledTier
		response["scheduled_change_at"] = sub.ScheduledChangeAt
	}

	// Add cancelled_at if present
	if sub.CancelledAt != nil && !sub.CancelledAt.IsZero() {
		response["cancelled_at"] = sub.CancelledAt
	}

	return c.JSON(response)
}

// CreateCheckoutRequest represents a checkout creation request
type CreateCheckoutRequest struct {
	PlanID string `json:"plan_id" validate:"required"`
}

// CreateCheckout creates a checkout session for a subscription
// POST /api/subscriptions/checkout
func (h *SubscriptionHandler) CreateCheckout(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Get email from auth context for syncing new users
	userEmail, _ := c.Locals("user_email").(string)

	var req CreateCheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.PlanID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "plan_id is required",
		})
	}

	ctx := context.Background()
	checkout, err := h.paymentService.CreateCheckoutSession(ctx, userID, userEmail, req.PlanID)
	if err != nil {
		log.Printf("⚠️  Failed to create checkout for user %s: %v", userID, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(checkout)
}

// ChangePlanRequest represents a plan change request
type ChangePlanRequest struct {
	PlanID string `json:"plan_id" validate:"required"`
}

// ChangePlan changes the user's subscription plan
// POST /api/subscriptions/change-plan
func (h *SubscriptionHandler) ChangePlan(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	var req ChangePlanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.PlanID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "plan_id is required",
		})
	}

	ctx := context.Background()
	result, err := h.paymentService.ChangePlan(ctx, userID, req.PlanID)
	if err != nil {
		log.Printf("⚠️  Failed to change plan for user %s: %v", userID, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}

// PreviewPlanChange previews a plan change
// GET /api/subscriptions/change-plan/preview?plan_id=pro
func (h *SubscriptionHandler) PreviewPlanChange(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	planID := c.Query("plan_id")
	if planID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "plan_id query parameter is required",
		})
	}

	ctx := context.Background()
	preview, err := h.paymentService.PreviewPlanChange(ctx, userID, planID)
	if err != nil {
		log.Printf("⚠️  Failed to preview plan change for user %s: %v", userID, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(preview)
}

// Cancel cancels the user's subscription
// POST /api/subscriptions/cancel
func (h *SubscriptionHandler) Cancel(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	ctx := context.Background()
	err := h.paymentService.CancelSubscription(ctx, userID)
	if err != nil {
		log.Printf("⚠️  Failed to cancel subscription for user %s: %v", userID, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Get updated subscription to return cancel date
	sub, _ := h.paymentService.GetCurrentSubscription(ctx, userID)
	return c.JSON(fiber.Map{
		"status":    "pending_cancel",
		"cancel_at": sub.CurrentPeriodEnd,
		"message":   "Your subscription will be cancelled at the end of the billing period. You'll retain access until then.",
	})
}

// Reactivate reactivates a cancelled subscription
// POST /api/subscriptions/reactivate
func (h *SubscriptionHandler) Reactivate(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	ctx := context.Background()
	err := h.paymentService.ReactivateSubscription(ctx, userID)
	if err != nil {
		log.Printf("⚠️  Failed to reactivate subscription for user %s: %v", userID, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "active",
		"message": "Subscription reactivated successfully",
	})
}

// GetPortalURL returns the DodoPayments customer portal URL
// GET /api/subscriptions/portal
func (h *SubscriptionHandler) GetPortalURL(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	ctx := context.Background()
	url, err := h.paymentService.GetCustomerPortalURL(ctx, userID)
	if err != nil {
		log.Printf("⚠️  Failed to get portal URL for user %s: %v", userID, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"portal_url": url,
	})
}

// ListInvoices returns invoice history (placeholder - requires DodoPayments API)
// GET /api/subscriptions/invoices
func (h *SubscriptionHandler) ListInvoices(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// TODO: Implement invoice listing when DodoPayments API is available
	return c.JSON(fiber.Map{
		"invoices": []interface{}{},
		"message":  "Invoice history coming soon",
	})
}

// SyncSubscription manually syncs subscription data from DodoPayments
// POST /api/subscriptions/sync
func (h *SubscriptionHandler) SyncSubscription(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	result, err := h.paymentService.SyncSubscriptionFromDodo(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}

// GetUsageStats returns current usage statistics for the user
// GET /api/subscriptions/usage
func (h *SubscriptionHandler) GetUsageStats(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	ctx := c.Context()
	stats, err := h.paymentService.GetUsageStats(ctx, userID)
	if err != nil {
		log.Printf("⚠️  Failed to get usage stats for user %s: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get usage statistics",
		})
	}

	return c.JSON(stats)
}
