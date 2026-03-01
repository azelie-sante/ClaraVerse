package handlers

import (
	"claraverse/internal/models"
	"claraverse/internal/services"
	"log"

	"github.com/gofiber/fiber/v2"
)

// SystemModelsHandler handles system model assignment endpoints
type SystemModelsHandler struct {
	settingsService *services.SettingsService
}

// NewSystemModelsHandler creates a new system models handler
func NewSystemModelsHandler(settingsService *services.SettingsService) *SystemModelsHandler {
	return &SystemModelsHandler{
		settingsService: settingsService,
	}
}

// GetSystemModelAssignments returns the current system model assignments
// GET /api/admin/system-models
func (h *SystemModelsHandler) GetSystemModelAssignments(c *fiber.Ctx) error {
	assignments, err := h.settingsService.GetSystemModelAssignments(c.Context())
	if err != nil {
		log.Printf("❌ [ADMIN] Failed to get system model assignments: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve system model assignments",
		})
	}

	return c.JSON(assignments)
}

// UpdateSystemModelAssignments updates the system model assignments
// PUT /api/admin/system-models
func (h *SystemModelsHandler) UpdateSystemModelAssignments(c *fiber.Ctx) error {
	var assignments models.SystemModelAssignments
	if err := c.BodyParser(&assignments); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.settingsService.SetSystemModelAssignments(c.Context(), &assignments); err != nil {
		log.Printf("❌ [ADMIN] Failed to update system model assignments: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update system model assignments",
		})
	}

	log.Printf("✅ [ADMIN] Updated system model assignments")
	return c.JSON(fiber.Map{
		"success": true,
		"message": "System model assignments updated successfully",
	})
}
