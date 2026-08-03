package handlers

import (
	"claraverse/internal/models"
	"claraverse/internal/services"

	"github.com/gofiber/fiber/v2"
)

// AgentProviderConfigHandler serves Clara Agent (the local CLI coding agent)
// the real connection details for the account's configured model provider,
// plus the catalog of models available on it.
//
// Unlike GET /api/providers (public, credentials stripped), this endpoint
// requires an authenticated caller and returns the actual base URL and API
// key so Clara Agent can call the provider directly, the way its BYOK
// design expects.
type AgentProviderConfigHandler struct {
	providerService *services.ProviderService
	modelService    *services.ModelService
	userService     *services.UserService
}

func NewAgentProviderConfigHandler(
	providerService *services.ProviderService,
	modelService *services.ModelService,
	userService *services.UserService,
) *AgentProviderConfigHandler {
	return &AgentProviderConfigHandler{
		providerService: providerService,
		modelService:    modelService,
		userService:     userService,
	}
}

// AgentModelInfo describes one model in a way Clara Agent can turn directly
// into a catalog entry (context window, tool/vision support) rather than
// having to guess defaults.
type AgentModelInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	DisplayName   string `json:"display_name,omitempty"`
	ContextLength int    `json:"context_length,omitempty"`
	SupportsTools bool   `json:"supports_tools"`
	SupportsVision bool  `json:"supports_vision"`
}

// AgentProviderConfigResponse mirrors what Clara Agent's claraverse
// provider extension expects (see clara-agent/packages/coding-agent/src/
// extensions/claraverse/provider.ts).
type AgentProviderConfigResponse struct {
	BaseURL      string           `json:"base_url"`
	APIKey       string           `json:"api_key,omitempty"`
	DefaultModel string           `json:"default_model,omitempty"`
	Models       []AgentModelInfo `json:"models,omitempty"`
}

// GetProviderConfig handles GET /api/agent/provider-config
func (h *AgentProviderConfigHandler) GetProviderConfig(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" || userID == "anonymous" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	// Resolution order for which provider to hand over: the user's own
	// default-model preference, else the first enabled provider. Both are
	// existing concepts (models.UserPreferences.DefaultModelID and the
	// providers table); this endpoint picks between them rather than
	// introducing a third source of truth.
	var provider *models.Provider
	var defaultModelName string

	if h.userService != nil {
		if prefs, err := h.userService.GetPreferences(c.Context(), userID); err == nil && prefs != nil && prefs.DefaultModelID != "" {
			if p, modelName, err := h.providerService.GetByModelIDWithName(prefs.DefaultModelID); err == nil {
				provider = p
				defaultModelName = modelName
			}
			// A preference pointing at a deleted model falls through to the
			// provider-level default rather than failing the whole request.
		}
	}

	if provider == nil {
		providers, err := h.providerService.GetAll()
		if err != nil || len(providers) == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "no configured model provider found for this account",
			})
		}
		provider = &providers[0]
		defaultModelName = provider.DefaultModel
	}

	// Ship the whole visible catalog for this provider so Clara Agent's
	// /model picker has real options with real context windows, instead of
	// a single entry with guessed defaults.
	var modelInfos []AgentModelInfo
	if h.modelService != nil {
		if list, err := h.modelService.GetByProvider(provider.ID, true); err == nil {
			for _, m := range list {
				modelInfos = append(modelInfos, AgentModelInfo{
					ID:             m.Name,
					Name:           m.Name,
					DisplayName:    m.DisplayName,
					ContextLength:  m.ContextLength,
					SupportsTools:  m.SupportsTools,
					SupportsVision: m.SupportsVision,
				})
			}
		}
	}

	// Fall back to the provider's own default when no catalog row is visible,
	// so a minimally-configured instance still yields something usable.
	if defaultModelName == "" && len(modelInfos) > 0 {
		defaultModelName = modelInfos[0].Name
	}
	if defaultModelName == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "configured provider has no models available",
		})
	}

	return c.JSON(AgentProviderConfigResponse{
		BaseURL:      provider.BaseURL,
		APIKey:       provider.APIKey,
		DefaultModel: defaultModelName,
		Models:       modelInfos,
	})
}
