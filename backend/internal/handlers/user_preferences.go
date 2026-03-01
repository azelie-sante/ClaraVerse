package handlers

import (
	"claraverse/internal/models"
	"claraverse/internal/services"

	"github.com/gofiber/fiber/v2"
)

// UserPreferencesHandler handles user preferences HTTP requests
type UserPreferencesHandler struct {
	userService *services.UserService
}

// NewUserPreferencesHandler creates a new UserPreferencesHandler
func NewUserPreferencesHandler(userService *services.UserService) *UserPreferencesHandler {
	return &UserPreferencesHandler{userService: userService}
}

// Get retrieves user preferences
// GET /api/preferences
func (h *UserPreferencesHandler) Get(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Get email from context (set by auth middleware)
	email, _ := c.Locals("user_email").(string)

	// Try to sync user first (creates if not exists)
	_, err := h.userService.SyncUserFromSupabase(c.Context(), userID, email)
	if err != nil {
		// Log but don't fail - try to get preferences anyway
		println("Warning: Failed to sync user:", err.Error())
	}

	prefs, err := h.userService.GetPreferences(c.Context(), userID)
	if err != nil {
		// Return default preferences if user not found
		return c.JSON(models.UserPreferences{
			StoreBuilderChatHistory: true,
			ChatPrivacyMode:         "",
			Theme:                   "dark",
			FontSize:                "medium",
		})
	}

	return c.JSON(prefs)
}

// Update updates user preferences
// PUT /api/preferences
func (h *UserPreferencesHandler) Update(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	var req models.UpdateUserPreferencesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate chat privacy mode if provided
	if req.ChatPrivacyMode != nil {
		mode := *req.ChatPrivacyMode
		if mode != models.ChatPrivacyModeLocal && mode != models.ChatPrivacyModeCloud && mode != "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid chat_privacy_mode. Must be 'local' or 'cloud'",
			})
		}
	}

	prefs, err := h.userService.UpdatePreferences(c.Context(), userID, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update preferences",
		})
	}

	return c.JSON(prefs)
}
