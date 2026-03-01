package handlers

import (
	"claraverse/internal/services"
	"time"

	"github.com/gofiber/fiber/v2"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	connManager *services.ConnectionManager
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(connManager *services.ConnectionManager) *HealthHandler {
	return &HealthHandler{connManager: connManager}
}

// Handle responds with server health status
func (h *HealthHandler) Handle(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":      "healthy",
		"connections": h.connManager.Count(),
		"timestamp":   time.Now().Format(time.RFC3339),
	})
}
