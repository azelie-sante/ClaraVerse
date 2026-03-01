package handlers

import (
	"claraverse/internal/models"
	"claraverse/internal/services"
	"log"

	"github.com/gofiber/fiber/v2"
)

// ScheduleHandler handles schedule-related HTTP requests
type ScheduleHandler struct {
	schedulerService *services.SchedulerService
	agentService     *services.AgentService
}

// NewScheduleHandler creates a new schedule handler
func NewScheduleHandler(schedulerService *services.SchedulerService, agentService *services.AgentService) *ScheduleHandler {
	return &ScheduleHandler{
		schedulerService: schedulerService,
		agentService:     agentService,
	}
}

// Create creates a new schedule for an agent
// POST /api/agents/:id/schedule
func (h *ScheduleHandler) Create(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	agentID := c.Params("id")
	if agentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Agent ID is required",
		})
	}

	// Verify agent exists and belongs to user
	agent, err := h.agentService.GetAgent(agentID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Agent not found",
		})
	}

	// Check if workflow has file inputs (which expire in 30 minutes)
	if hasFileInputs(agent.Workflow) {
		log.Printf("🚫 [SCHEDULE] Cannot schedule agent %s: workflow has file inputs", agentID)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":      "Cannot schedule workflows with file inputs",
			"reason":     "Uploaded files expire after 30 minutes and won't be available at scheduled execution time",
			"suggestion": "Use the API trigger endpoint instead: POST /api/trigger/" + agentID,
		})
	}

	var req models.CreateScheduleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.CronExpression == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cronExpression is required",
		})
	}
	if req.Timezone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "timezone is required",
		})
	}

	log.Printf("📅 [SCHEDULE] Creating schedule for agent %s (user: %s, cron: %s)", agentID, userID, req.CronExpression)

	schedule, err := h.schedulerService.CreateSchedule(c.Context(), agentID, userID, &req)
	if err != nil {
		log.Printf("❌ [SCHEDULE] Failed to create schedule: %v", err)

		// Check for specific errors
		if err.Error() == "agent already has a schedule" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if err.Error()[:14] == "schedule limit" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	log.Printf("✅ [SCHEDULE] Created schedule %s for agent %s", schedule.ID.Hex(), agentID)
	return c.Status(fiber.StatusCreated).JSON(schedule.ToResponse())
}

// Get retrieves the schedule for an agent
// GET /api/agents/:id/schedule
func (h *ScheduleHandler) Get(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	agentID := c.Params("id")
	if agentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Agent ID is required",
		})
	}

	schedule, err := h.schedulerService.GetScheduleByAgentID(c.Context(), agentID, userID)
	if err != nil {
		if err.Error() == "schedule not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "No schedule found for this agent",
			})
		}
		log.Printf("❌ [SCHEDULE] Failed to get schedule: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get schedule",
		})
	}

	return c.JSON(schedule.ToResponse())
}

// Update updates the schedule for an agent
// PUT /api/agents/:id/schedule
func (h *ScheduleHandler) Update(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	agentID := c.Params("id")
	if agentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Agent ID is required",
		})
	}

	// Get existing schedule
	existingSchedule, err := h.schedulerService.GetScheduleByAgentID(c.Context(), agentID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No schedule found for this agent",
		})
	}

	var req models.UpdateScheduleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	log.Printf("📝 [SCHEDULE] Updating schedule %s for agent %s", existingSchedule.ID.Hex(), agentID)

	schedule, err := h.schedulerService.UpdateSchedule(c.Context(), existingSchedule.ID.Hex(), userID, &req)
	if err != nil {
		log.Printf("❌ [SCHEDULE] Failed to update schedule: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	log.Printf("✅ [SCHEDULE] Updated schedule %s", schedule.ID.Hex())
	return c.JSON(schedule.ToResponse())
}

// Delete deletes the schedule for an agent
// DELETE /api/agents/:id/schedule
func (h *ScheduleHandler) Delete(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	agentID := c.Params("id")
	if agentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Agent ID is required",
		})
	}

	// Get existing schedule
	existingSchedule, err := h.schedulerService.GetScheduleByAgentID(c.Context(), agentID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No schedule found for this agent",
		})
	}

	log.Printf("🗑️ [SCHEDULE] Deleting schedule %s for agent %s", existingSchedule.ID.Hex(), agentID)

	if err := h.schedulerService.DeleteSchedule(c.Context(), existingSchedule.ID.Hex(), userID); err != nil {
		log.Printf("❌ [SCHEDULE] Failed to delete schedule: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete schedule",
		})
	}

	log.Printf("✅ [SCHEDULE] Deleted schedule for agent %s", agentID)
	return c.Status(fiber.StatusNoContent).Send(nil)
}

// TriggerNow triggers an immediate execution of the schedule
// POST /api/agents/:id/schedule/run
func (h *ScheduleHandler) TriggerNow(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	agentID := c.Params("id")
	if agentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Agent ID is required",
		})
	}

	// Get existing schedule
	existingSchedule, err := h.schedulerService.GetScheduleByAgentID(c.Context(), agentID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No schedule found for this agent",
		})
	}

	log.Printf("▶️ [SCHEDULE] Triggering immediate run for schedule %s (agent: %s)", existingSchedule.ID.Hex(), agentID)

	if err := h.schedulerService.TriggerNow(c.Context(), existingSchedule.ID.Hex(), userID); err != nil {
		log.Printf("❌ [SCHEDULE] Failed to trigger schedule: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to trigger schedule",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Schedule triggered successfully",
	})
}

// GetUsage returns the user's schedule usage stats
// GET /api/schedules/usage
func (h *ScheduleHandler) GetUsage(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	usage, err := h.schedulerService.GetScheduleUsage(c.Context(), userID)
	if err != nil {
		log.Printf("❌ [SCHEDULE] Failed to get schedule usage: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get schedule usage",
		})
	}

	return c.JSON(usage)
}

// hasFileInputs checks if a workflow has any variable blocks with file input type
// File inputs cannot be scheduled because uploaded files expire after 30 minutes
func hasFileInputs(workflow *models.Workflow) bool {
	if workflow == nil {
		return false
	}
	for _, block := range workflow.Blocks {
		if block.Type == "variable" {
			// Config is map[string]any
			if inputType, exists := block.Config["inputType"]; exists && inputType == "file" {
				return true
			}
		}
	}
	return false
}
