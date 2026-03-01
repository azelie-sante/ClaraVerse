package handlers

import (
	"claraverse/internal/services"

	"github.com/gofiber/fiber/v2"
)

// ConfigHandler handles configuration API requests
type ConfigHandler struct {
	configService *services.ConfigService
}

// NewConfigHandler creates a new config handler
func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{
		configService: services.GetConfigService(),
	}
}

// GetRecommendedModels returns recommended models for all providers
func (h *ConfigHandler) GetRecommendedModels(c *fiber.Ctx) error {
	recommended := h.configService.GetAllRecommendedModels()

	// Convert to a frontend-friendly format
	response := make(map[string]interface{})
	for providerID, models := range recommended {
		response[string(rune(providerID+'0'))] = fiber.Map{
			"top":     models.Top,
			"medium":  models.Medium,
			"fastest": models.Fastest,
		}
	}

	return c.JSON(fiber.Map{
		"recommended_models": recommended,
	})
}
