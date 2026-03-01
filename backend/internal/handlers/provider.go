package handlers

import (
	"claraverse/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// ProviderHandler handles provider-related requests
type ProviderHandler struct {
	providerService *services.ProviderService
}

// NewProviderHandler creates a new provider handler
func NewProviderHandler(providerService *services.ProviderService) *ProviderHandler {
	return &ProviderHandler{providerService: providerService}
}

// List returns all enabled providers (names only, no credentials)
func (h *ProviderHandler) List(c *fiber.Ctx) error {
	providers, err := h.providerService.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch providers",
		})
	}

	// Hide sensitive information (API keys, base URLs)
	type PublicProvider struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	publicProviders := make([]PublicProvider, len(providers))
	for i, p := range providers {
		publicProviders[i] = PublicProvider{
			ID:   p.ID,
			Name: p.Name,
		}
	}

	return c.JSON(fiber.Map{
		"providers": publicProviders,
		"count":     len(publicProviders),
	})
}

// GetModels returns models for a specific provider
func (h *ProviderHandler) GetModels(c *fiber.Ctx) error {
	providerID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid provider ID",
		})
	}

	// Get provider to verify it exists
	_, err = h.providerService.GetByID(providerID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Provider not found",
		})
	}

	return c.Next()
}
