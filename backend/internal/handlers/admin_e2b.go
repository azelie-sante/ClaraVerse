package handlers

import (
	"claraverse/internal/services"
	"context"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const settingKeyE2BAPIKey = "e2b.api_key"

// E2BHandler handles E2B code execution settings
type E2BHandler struct {
	settingsService *services.SettingsService
}

// NewE2BHandler creates a new E2B settings handler
func NewE2BHandler(settingsService *services.SettingsService) *E2BHandler {
	return &E2BHandler{
		settingsService: settingsService,
	}
}

// GetE2BSettings returns the current E2B configuration
// GET /api/admin/e2b-settings
func (h *E2BHandler) GetE2BSettings(c *fiber.Ctx) error {
	apiKey, err := h.settingsService.Get(c.Context(), settingKeyE2BAPIKey)
	if err != nil {
		log.Printf("❌ [ADMIN] Failed to get E2B settings: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve E2B settings",
		})
	}

	// Mask the API key for display
	maskedKey := ""
	if apiKey != "" {
		if len(apiKey) > 8 {
			maskedKey = apiKey[:4] + strings.Repeat("*", len(apiKey)-8) + apiKey[len(apiKey)-4:]
		} else {
			maskedKey = "****"
		}
	}

	return c.JSON(fiber.Map{
		"api_key_set":    apiKey != "",
		"api_key_masked": maskedKey,
	})
}

// UpdateE2BSettings updates the E2B configuration
// PUT /api/admin/e2b-settings
func (h *E2BHandler) UpdateE2BSettings(c *fiber.Ctx) error {
	var body struct {
		APIKey string `json:"api_key"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	apiKey := strings.TrimSpace(body.APIKey)

	if err := h.settingsService.Set(c.Context(), settingKeyE2BAPIKey, apiKey); err != nil {
		log.Printf("❌ [ADMIN] Failed to update E2B settings: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update E2B settings",
		})
	}

	action := "cleared"
	if apiKey != "" {
		action = "updated"
	}

	log.Printf("✅ [ADMIN] E2B API key %s", action)
	return c.JSON(fiber.Map{
		"success": true,
		"message": "E2B API key " + action + " successfully",
	})
}

// GetE2BAPIKey returns the raw E2B API key (for internal use by the executor)
func GetE2BAPIKey(settingsService *services.SettingsService) string {
	key, err := settingsService.Get(context.Background(), settingKeyE2BAPIKey)
	if err != nil {
		return ""
	}
	return key
}
