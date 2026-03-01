package handlers

import (
	"claraverse/internal/models"
	"claraverse/internal/services"

	"github.com/gofiber/fiber/v2"
)

// SkillHandler handles skill-related HTTP requests
type SkillHandler struct {
	skillService *services.SkillService
}

// NewSkillHandler creates a new skill handler
func NewSkillHandler(skillService *services.SkillService) *SkillHandler {
	return &SkillHandler{skillService: skillService}
}

// ListSkills returns all skills, optionally filtered by category
func (h *SkillHandler) ListSkills(c *fiber.Ctx) error {
	category := c.Query("category", "")

	skills, err := h.skillService.ListSkills(c.Context(), category)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list skills",
		})
	}

	// Group by category for the frontend
	categoryMap := make(map[string][]models.Skill)
	for _, skill := range skills {
		categoryMap[skill.Category] = append(categoryMap[skill.Category], skill)
	}

	return c.JSON(fiber.Map{
		"skills":     skills,
		"categories": categoryMap,
		"total":      len(skills),
	})
}

// GetSkill returns a single skill by ID
func (h *SkillHandler) GetSkill(c *fiber.Ctx) error {
	id := c.Params("id")

	skill, err := h.skillService.GetSkill(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Skill not found",
		})
	}

	return c.JSON(skill)
}

// CreateSkill creates a new custom skill
func (h *SkillHandler) CreateSkill(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var req models.CreateSkillRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" || req.SystemPrompt == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name and system_prompt are required",
		})
	}

	skill := &models.Skill{
		Name:             req.Name,
		Description:      req.Description,
		Icon:             req.Icon,
		Category:         req.Category,
		SystemPrompt:     req.SystemPrompt,
		RequiredTools:    req.RequiredTools,
		PreferredServers: req.PreferredServers,
		Keywords:         req.Keywords,
		TriggerPatterns:  req.TriggerPatterns,
		Mode:             req.Mode,
		AuthorID:         userID,
	}

	if err := h.skillService.CreateSkill(c.Context(), skill); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create skill",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(skill)
}

// UpdateSkill updates an existing skill
func (h *SkillHandler) UpdateSkill(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	id := c.Params("id")

	// Verify ownership for non-builtin skills
	existing, err := h.skillService.GetSkill(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Skill not found",
		})
	}
	if !existing.IsBuiltin && existing.AuthorID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You can only edit your own skills",
		})
	}

	var req models.CreateSkillRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	skill := &models.Skill{
		Name:             req.Name,
		Description:      req.Description,
		Icon:             req.Icon,
		Category:         req.Category,
		SystemPrompt:     req.SystemPrompt,
		RequiredTools:    req.RequiredTools,
		PreferredServers: req.PreferredServers,
		Keywords:         req.Keywords,
		TriggerPatterns:  req.TriggerPatterns,
		Mode:             req.Mode,
	}

	if err := h.skillService.UpdateSkill(c.Context(), id, skill); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update skill",
		})
	}

	return c.JSON(fiber.Map{"success": true})
}

// DeleteSkill deletes a custom skill
func (h *SkillHandler) DeleteSkill(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	id := c.Params("id")

	// Verify ownership
	existing, err := h.skillService.GetSkill(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Skill not found",
		})
	}
	if existing.IsBuiltin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Cannot delete built-in skills",
		})
	}
	if existing.AuthorID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You can only delete your own skills",
		})
	}

	if err := h.skillService.DeleteSkill(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete skill",
		})
	}

	return c.JSON(fiber.Map{"success": true})
}

// GetMySkills returns the current user's enabled skills with details
func (h *SkillHandler) GetMySkills(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userSkills, err := h.skillService.GetUserSkills(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user skills",
		})
	}

	return c.JSON(fiber.Map{
		"skills": userSkills,
		"total":  len(userSkills),
	})
}

// EnableSkill enables a skill for the current user
func (h *SkillHandler) EnableSkill(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	id := c.Params("id")

	if err := h.skillService.EnableSkill(c.Context(), userID, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to enable skill: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{"success": true})
}

// DisableSkill disables a skill for the current user
func (h *SkillHandler) DisableSkill(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	id := c.Params("id")

	if err := h.skillService.DisableSkill(c.Context(), userID, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to disable skill",
		})
	}

	return c.JSON(fiber.Map{"success": true})
}

// BulkEnable enables multiple skills at once
func (h *SkillHandler) BulkEnable(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var req models.BulkEnableRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.skillService.BulkEnableSkills(c.Context(), userID, req.SkillIDs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to bulk enable skills",
		})
	}

	return c.JSON(fiber.Map{"success": true, "enabled": len(req.SkillIDs)})
}

// ═══════════════════════════════════════════════════════════════════════════
// SKILL.MD IMPORT / EXPORT / COMMUNITY
// ═══════════════════════════════════════════════════════════════════════════

// ImportSkillMD parses pasted SKILL.md content and creates a skill
func (h *SkillHandler) ImportSkillMD(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var req models.ImportSkillMDRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "SKILL.md content is required",
		})
	}

	skill, err := h.skillService.ImportFromSkillMD(c.Context(), req.Content, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to import: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(skill)
}

// ImportFromGitHub fetches a SKILL.md from a GitHub URL and creates a skill
func (h *SkillHandler) ImportFromGitHub(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var req models.ImportGitHubURLRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "GitHub URL is required",
		})
	}

	skill, err := h.skillService.ImportFromGitHubURL(c.Context(), req.URL, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to import from GitHub: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(skill)
}

// ExportSkillMD exports a skill as SKILL.md format
func (h *SkillHandler) ExportSkillMD(c *fiber.Ctx) error {
	id := c.Params("id")

	content, err := h.skillService.ExportAsSkillMD(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Skill not found",
		})
	}

	return c.JSON(fiber.Map{"content": content})
}

// ListCommunitySkills returns community skills from GitHub
func (h *SkillHandler) ListCommunitySkills(c *fiber.Ctx) error {
	skills, err := h.skillService.FetchCommunitySkills(c.Context())
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":  "Failed to fetch community skills: " + err.Error(),
			"skills": []interface{}{},
			"total":  0,
		})
	}

	return c.JSON(fiber.Map{
		"skills": skills,
		"total":  len(skills),
	})
}
