package handlers

import (
	"claraverse/internal/models"
	"claraverse/internal/services"
	"log"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// APIKeyHandler handles API key management endpoints
type APIKeyHandler struct {
	apiKeyService *services.APIKeyService
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(apiKeyService *services.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyService: apiKeyService,
	}
}

// Create creates a new API key
// POST /api/keys
func (h *APIKeyHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req models.CreateAPIKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}

	if len(req.Scopes) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one scope is required",
		})
	}

	result, err := h.apiKeyService.Create(c.Context(), userID, &req)
	if err != nil {
		log.Printf("❌ [APIKEY] Failed to create API key: %v", err)
		// Check for limit error
		if err.Error()[:14] == "API key limit" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// List lists all API keys for the user
// GET /api/keys
func (h *APIKeyHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	keys, err := h.apiKeyService.ListByUser(c.Context(), userID)
	if err != nil {
		log.Printf("❌ [APIKEY] Failed to list API keys: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list API keys",
		})
	}

	if keys == nil {
		keys = []*models.APIKeyListItem{}
	}

	return c.JSON(fiber.Map{
		"keys": keys,
	})
}

// Get retrieves a specific API key
// GET /api/keys/:id
func (h *APIKeyHandler) Get(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	keyIDStr := c.Params("id")

	keyID, err := primitive.ObjectIDFromHex(keyIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid key ID",
		})
	}

	key, err := h.apiKeyService.GetByIDAndUser(c.Context(), keyID, userID)
	if err != nil {
		if err.Error() == "API key not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "API key not found",
			})
		}
		log.Printf("❌ [APIKEY] Failed to get API key: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get API key",
		})
	}

	return c.JSON(key.ToListItem())
}

// Revoke revokes an API key (soft delete)
// POST /api/keys/:id/revoke
func (h *APIKeyHandler) Revoke(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	keyIDStr := c.Params("id")

	keyID, err := primitive.ObjectIDFromHex(keyIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid key ID",
		})
	}

	if err := h.apiKeyService.Revoke(c.Context(), keyID, userID); err != nil {
		if err.Error() == "API key not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "API key not found",
			})
		}
		log.Printf("❌ [APIKEY] Failed to revoke API key: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to revoke API key",
		})
	}

	return c.JSON(fiber.Map{
		"message": "API key revoked successfully",
	})
}

// Delete permanently deletes an API key
// DELETE /api/keys/:id
func (h *APIKeyHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	keyIDStr := c.Params("id")

	keyID, err := primitive.ObjectIDFromHex(keyIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid key ID",
		})
	}

	if err := h.apiKeyService.Delete(c.Context(), keyID, userID); err != nil {
		if err.Error() == "API key not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "API key not found",
			})
		}
		log.Printf("❌ [APIKEY] Failed to delete API key: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete API key",
		})
	}

	return c.JSON(fiber.Map{
		"message": "API key deleted successfully",
	})
}
