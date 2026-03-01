package handlers

import (
	"claraverse/internal/models"
	"claraverse/internal/services"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NexusHandler handles REST endpoints for the Nexus multi-agent system
type NexusHandler struct {
	taskStore     *services.NexusTaskStore
	sessionStore  *services.NexusSessionStore
	daemonPool    *services.DaemonPool
	personaStore  *services.PersonaService
	engramService *services.EngramService
	templateStore *services.DaemonTemplateStore
	projectStore  *services.NexusProjectStore
	saveStore     *services.NexusSaveStore
}

// NewNexusHandler creates a new Nexus REST handler
func NewNexusHandler(
	taskStore *services.NexusTaskStore,
	sessionStore *services.NexusSessionStore,
	daemonPool *services.DaemonPool,
	personaStore *services.PersonaService,
	engramService *services.EngramService,
	templateStore *services.DaemonTemplateStore,
	projectStore *services.NexusProjectStore,
	saveStore *services.NexusSaveStore,
) *NexusHandler {
	return &NexusHandler{
		taskStore:     taskStore,
		sessionStore:  sessionStore,
		daemonPool:    daemonPool,
		personaStore:  personaStore,
		engramService: engramService,
		templateStore: templateStore,
		projectStore:  projectStore,
		saveStore:     saveStore,
	}
}

// GetSession returns the current user's Nexus session
func (h *NexusHandler) GetSession(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	session, err := h.sessionStore.GetOrCreate(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(session)
}

// ListTasks returns tasks for the current user
func (h *NexusHandler) ListTasks(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	status := c.Query("status")
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)
	projectID := c.Query("project_id")

	filters := services.TaskFilters{
		Status: models.NexusTaskStatus(status),
		Limit:  int64(limit),
		Offset: int64(offset),
	}
	if projectID != "" {
		filters.ProjectID = &projectID
	}

	tasks, err := h.taskStore.List(c.Context(), userID, filters)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(tasks)
}

// GetTask returns a specific task
func (h *NexusHandler) GetTask(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	taskID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid task ID"})
	}

	task, err := h.taskStore.GetByID(c.Context(), userID, taskID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(task)
}

// CreateTask creates a new task (currently used for drafts — tasks with status="draft")
func (h *NexusHandler) CreateTask(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var body struct {
		Prompt    string `json:"prompt"`
		Goal      string `json:"goal"`
		Priority  int    `json:"priority"`
		Mode      string `json:"mode"`
		Status    string `json:"status"`
		ModelID   string `json:"model_id"`
		ProjectID string `json:"project_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if body.Prompt == "" && body.Goal == "" {
		return c.Status(400).JSON(fiber.Map{"error": "prompt or goal required"})
	}

	// Only allow creating draft tasks via REST — pending+ tasks go through WebSocket/Cortex
	if body.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{"error": "only draft tasks can be created via REST"})
	}

	session, err := h.sessionStore.GetOrCreate(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	task := models.NexusTask{
		SessionID: session.ID,
		UserID:    userID,
		Prompt:    body.Prompt,
		Goal:      body.Goal,
		Priority:  body.Priority,
		Source:    "user",
		Mode:      body.Mode,
		Status:    models.NexusTaskStatusDraft,
		ModelID:   body.ModelID,
	}
	if task.Goal == "" {
		task.Goal = task.Prompt
	}
	if task.Mode == "" {
		task.Mode = "quick"
	}

	if body.ProjectID != "" {
		oid, err := primitive.ObjectIDFromHex(body.ProjectID)
		if err == nil {
			task.ProjectID = &oid
		}
	}

	if err := h.taskStore.Create(c.Context(), &task); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Add to session's recent tasks so it shows up on reconnect
	_ = h.sessionStore.AddRecentTask(c.Context(), userID, task.ID)

	return c.Status(201).JSON(task)
}

// UpdateTask updates a draft task's content (prompt/goal)
func (h *NexusHandler) UpdateTask(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	taskID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid task ID"})
	}

	var body struct {
		Prompt string `json:"prompt"`
		Goal   string `json:"goal"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.taskStore.UpdateContent(c.Context(), userID, taskID, body.Prompt, body.Goal); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "updated"})
}

// DeleteTask removes a completed/failed/cancelled/draft task
func (h *NexusHandler) DeleteTask(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	taskID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid task ID"})
	}

	if err := h.taskStore.Delete(c.Context(), userID, taskID); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	// Also remove from session's recent task list so it doesn't reappear on reconnect
	_ = h.sessionStore.RemoveRecentTask(c.Context(), userID, taskID)

	return c.JSON(fiber.Map{"status": "deleted"})
}

// ListDaemons returns active daemons for the current user
func (h *NexusHandler) ListDaemons(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	daemons, err := h.daemonPool.GetActiveDaemons(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(daemons)
}

// GetDaemon returns a specific daemon's details
func (h *NexusHandler) GetDaemon(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	daemonID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid daemon ID"})
	}

	daemon, err := h.daemonPool.GetByID(c.Context(), userID, daemonID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(daemon)
}

// CancelDaemon cancels a specific daemon (ownership-checked)
func (h *NexusHandler) CancelDaemon(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	daemonID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid daemon ID"})
	}

	// Verify ownership before cancelling
	daemon, err := h.daemonPool.GetByID(c.Context(), userID, daemonID)
	if err != nil || daemon == nil {
		return c.Status(404).JSON(fiber.Map{"error": "daemon not found"})
	}

	if cancelErr := h.daemonPool.Cancel(daemonID.Hex()); cancelErr != nil {
		return c.Status(404).JSON(fiber.Map{"error": cancelErr.Error()})
	}

	return c.JSON(fiber.Map{"status": "cancelled"})
}

// GetPersona returns the user's persona facts
func (h *NexusHandler) GetPersona(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	facts, err := h.personaStore.GetOrCreateDefaults(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(facts)
}

// GetEngrams returns recent engram entries
func (h *NexusHandler) GetEngrams(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	limit := c.QueryInt("limit", 20)

	engrams, err := h.engramService.GetRecent(c.Context(), userID, int64(limit))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(engrams)
}

// --- Daemon Templates ---

// ListDaemonTemplates returns all templates visible to the user (system + user-owned)
func (h *NexusHandler) ListDaemonTemplates(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	templates, err := h.templateStore.GetAllForUser(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(templates)
}

// CreateDaemonTemplate creates a new user-owned daemon template
func (h *NexusHandler) CreateDaemonTemplate(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var template models.DaemonTemplate
	if err := c.BodyParser(&template); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	template.UserID = userID
	template.IsDefault = false
	template.IsActive = true

	if err := h.templateStore.Create(c.Context(), &template); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(template)
}

// UpdateDaemonTemplate updates a user-owned daemon template
func (h *NexusHandler) UpdateDaemonTemplate(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	templateID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid template ID"})
	}

	var updates models.DaemonTemplate
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.templateStore.Update(c.Context(), userID, templateID, &updates); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "updated"})
}

// DeleteDaemonTemplate deletes a user-owned daemon template
func (h *NexusHandler) DeleteDaemonTemplate(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	templateID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid template ID"})
	}

	if err := h.templateStore.Delete(c.Context(), userID, templateID); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}

// --- Projects ---

// ListProjects returns non-archived projects for the current user
func (h *NexusHandler) ListProjects(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	projects, err := h.projectStore.List(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if projects == nil {
		projects = []models.NexusProject{}
	}

	return c.JSON(projects)
}

// CreateProject creates a new project
func (h *NexusHandler) CreateProject(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var project models.NexusProject
	if err := c.BodyParser(&project); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if project.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name is required"})
	}

	project.UserID = userID
	if err := h.projectStore.Create(c.Context(), &project); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(project)
}

// UpdateProject updates a project
func (h *NexusHandler) UpdateProject(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	projectID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid project ID"})
	}

	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Remap JSON snake_case keys to BSON camelCase field names
	if v, ok := body["system_instruction"]; ok {
		body["systemInstruction"] = v
		delete(body, "system_instruction")
	}

	// Only allow updating safe fields (using BSON field names)
	allowed := map[string]bool{"name": true, "description": true, "systemInstruction": true, "icon": true, "color": true, "isArchived": true, "sortOrder": true}
	updates := make(map[string]interface{})
	for k, v := range body {
		if allowed[k] {
			updates[k] = v
		}
	}

	if err := h.projectStore.Update(c.Context(), userID, projectID, updates); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "updated"})
}

// DeleteProject deletes a project and reassigns its tasks to another project
func (h *NexusHandler) DeleteProject(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	projectID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid project ID"})
	}

	// Reassign tasks to another project so they aren't orphaned.
	// Accept optional reassign_to in request body; otherwise pick the first remaining project.
	var body struct {
		ReassignTo string `json:"reassign_to"`
	}
	_ = c.BodyParser(&body) // OK if no body

	if body.ReassignTo != "" {
		if targetID, err := primitive.ObjectIDFromHex(body.ReassignTo); err == nil {
			_ = h.taskStore.ReassignProjectTasks(c.Context(), userID, projectID, targetID)
		}
	} else {
		// Find the first non-deleted project that isn't this one
		projects, err := h.projectStore.List(c.Context(), userID)
		if err != nil {
			// Can't determine other projects — unlink tasks so they aren't orphaned
			_ = h.taskStore.UnsetProjectForAll(c.Context(), userID, projectID)
		} else {
			reassigned := false
			for _, p := range projects {
				if p.ID != projectID {
					_ = h.taskStore.ReassignProjectTasks(c.Context(), userID, projectID, p.ID)
					reassigned = true
					break
				}
			}
			if !reassigned {
				_ = h.taskStore.UnsetProjectForAll(c.Context(), userID, projectID)
			}
		}
	}

	if err := h.projectStore.Delete(c.Context(), userID, projectID); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}

// MoveTaskToProject assigns or removes a task from a project
func (h *NexusHandler) MoveTaskToProject(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	taskID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid task ID"})
	}

	var body struct {
		ProjectID *string `json:"project_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if body.ProjectID == nil || *body.ProjectID == "" {
		// Remove from project
		if err := h.taskStore.SetProjectID(c.Context(), userID, taskID, nil); err != nil {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
	} else {
		oid, err := primitive.ObjectIDFromHex(*body.ProjectID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid project_id"})
		}
		if err := h.taskStore.SetProjectID(c.Context(), userID, taskID, &oid); err != nil {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.JSON(fiber.Map{"status": "moved"})
}

// --- Saves ---

// ListSaves returns saved outputs for the current user
func (h *NexusHandler) ListSaves(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	filters := services.SaveFilters{
		Tag:    c.Query("tag"),
		Limit:  int64(c.QueryInt("limit", 50)),
		Offset: int64(c.QueryInt("offset", 0)),
	}

	saves, err := h.saveStore.List(c.Context(), userID, filters)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if saves == nil {
		saves = []models.NexusSave{}
	}

	return c.JSON(saves)
}

// GetSave returns a specific save
func (h *NexusHandler) GetSave(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	saveID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid save ID"})
	}

	save, err := h.saveStore.GetByID(c.Context(), userID, saveID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(save)
}

// CreateSave saves a task output or user-created document
func (h *NexusHandler) CreateSave(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var body struct {
		Title           string   `json:"title"`
		Content         string   `json:"content"`
		Tags            []string `json:"tags"`
		SourceTaskID    string   `json:"source_task_id"`
		SourceProjectID string   `json:"source_project_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if body.Title == "" || body.Content == "" {
		return c.Status(400).JSON(fiber.Map{"error": "title and content are required"})
	}

	save := models.NexusSave{
		UserID:  userID,
		Title:   body.Title,
		Content: body.Content,
		Tags:    body.Tags,
	}

	if body.SourceTaskID != "" {
		oid, err := primitive.ObjectIDFromHex(body.SourceTaskID)
		if err == nil {
			save.SourceTaskID = &oid
		}
	}
	if body.SourceProjectID != "" {
		oid, err := primitive.ObjectIDFromHex(body.SourceProjectID)
		if err == nil {
			save.SourceProjectID = &oid
		}
	}

	if err := h.saveStore.Create(c.Context(), &save); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(save)
}

// UpdateSave updates a save's title, content, or tags
func (h *NexusHandler) UpdateSave(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	saveID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid save ID"})
	}

	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Only allow updating safe fields
	allowed := map[string]bool{"title": true, "content": true, "tags": true}
	updates := make(map[string]interface{})
	for k, v := range body {
		if allowed[k] {
			updates[k] = v
		}
	}

	if err := h.saveStore.Update(c.Context(), userID, saveID, updates); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "updated"})
}

// DeleteSave removes a save
func (h *NexusHandler) DeleteSave(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	saveID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid save ID"})
	}

	if err := h.saveStore.Delete(c.Context(), userID, saveID); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}
