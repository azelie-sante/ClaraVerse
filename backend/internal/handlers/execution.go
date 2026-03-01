package handlers

import (
	"claraverse/internal/services"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ExecutionHandler handles execution-related HTTP requests
type ExecutionHandler struct {
	executionService *services.ExecutionService
}

// NewExecutionHandler creates a new execution handler
func NewExecutionHandler(executionService *services.ExecutionService) *ExecutionHandler {
	return &ExecutionHandler{
		executionService: executionService,
	}
}

// ListByAgent returns paginated executions for a specific agent
// GET /api/agents/:id/executions
func (h *ExecutionHandler) ListByAgent(c *fiber.Ctx) error {
	agentID := c.Params("id")
	userID := c.Locals("user_id").(string)

	opts := h.parseListOptions(c)

	result, err := h.executionService.ListByAgent(c.Context(), agentID, userID, opts)
	if err != nil {
		log.Printf("❌ [EXECUTION] Failed to list agent executions: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list executions",
		})
	}

	return c.JSON(result)
}

// ListAll returns paginated executions for the current user
// GET /api/executions
func (h *ExecutionHandler) ListAll(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	opts := h.parseListOptions(c)

	result, err := h.executionService.ListByUser(c.Context(), userID, opts)
	if err != nil {
		log.Printf("❌ [EXECUTION] Failed to list user executions: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list executions",
		})
	}

	return c.JSON(result)
}

// GetByID returns a specific execution
// GET /api/executions/:id
func (h *ExecutionHandler) GetByID(c *fiber.Ctx) error {
	executionIDStr := c.Params("id")
	userID := c.Locals("user_id").(string)

	executionID, err := primitive.ObjectIDFromHex(executionIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid execution ID",
		})
	}

	execution, err := h.executionService.GetByIDAndUser(c.Context(), executionID, userID)
	if err != nil {
		if err.Error() == "execution not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Execution not found",
			})
		}
		log.Printf("❌ [EXECUTION] Failed to get execution: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get execution",
		})
	}

	return c.JSON(execution)
}

// GetStats returns execution statistics for an agent
// GET /api/agents/:id/executions/stats
func (h *ExecutionHandler) GetStats(c *fiber.Ctx) error {
	agentID := c.Params("id")
	userID := c.Locals("user_id").(string)

	stats, err := h.executionService.GetStats(c.Context(), agentID, userID)
	if err != nil {
		log.Printf("❌ [EXECUTION] Failed to get stats: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get execution stats",
		})
	}

	return c.JSON(stats)
}

// parseListOptions extracts pagination and filter options from query params
func (h *ExecutionHandler) parseListOptions(c *fiber.Ctx) *services.ListExecutionsOptions {
	opts := &services.ListExecutionsOptions{
		Page:        1,
		Limit:       20,
		Status:      c.Query("status"),
		TriggerType: c.Query("trigger_type"),
		AgentID:     c.Query("agent_id"),
	}

	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
		opts.Page = page
	}

	if limit, err := strconv.Atoi(c.Query("limit")); err == nil && limit > 0 && limit <= 100 {
		opts.Limit = limit
	}

	return opts
}
