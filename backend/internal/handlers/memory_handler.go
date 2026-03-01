package handlers

import (
	"context"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"claraverse/internal/models"
	"claraverse/internal/services"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MemoryHandler handles memory-related API endpoints
type MemoryHandler struct {
	memoryStorageService   *services.MemoryStorageService
	memoryExtractionService *services.MemoryExtractionService
	chatService            *services.ChatService
}

// NewMemoryHandler creates a new memory handler
func NewMemoryHandler(
	memoryStorageService *services.MemoryStorageService,
	memoryExtractionService *services.MemoryExtractionService,
	chatService *services.ChatService,
) *MemoryHandler {
	return &MemoryHandler{
		memoryStorageService:   memoryStorageService,
		memoryExtractionService: memoryExtractionService,
		chatService:            chatService,
	}
}

// ListMemories returns paginated list of memories with optional filters
// GET /api/v1/memories?category=preferences&tags=ui,theme&includeArchived=false&page=1&pageSize=20
func (h *MemoryHandler) ListMemories(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Parse query parameters
	category := c.Query("category", "")
	tagsParam := c.Query("tags", "")
	includeArchived := c.Query("includeArchived", "false") == "true"
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Parse and sanitize tags (SECURITY: prevent NoSQL injection)
	var tags []string
	if tagsParam != "" {
		// Split by comma
		rawTags := strings.Split(tagsParam, ",")
		for _, rawTag := range rawTags {
			sanitizedTag := sanitizeTag(strings.TrimSpace(rawTag))
			if sanitizedTag != "" && len(sanitizedTag) <= 50 {
				tags = append(tags, sanitizedTag)
			}
		}

		// Limit number of tags to prevent abuse
		if len(tags) > 20 {
			tags = tags[:20]
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get memories
	memories, total, err := h.memoryStorageService.ListMemories(
		ctx,
		userID,
		category,
		tags,
		includeArchived,
		page,
		pageSize,
	)
	if err != nil {
		log.Printf("❌ [MEMORY-API] Failed to list memories: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve memories",
		})
	}

	// Build response
	memoryResponses := make([]fiber.Map, len(memories))
	for i, mem := range memories {
		memoryResponses[i] = buildMemoryResponse(mem)
	}

	return c.JSON(fiber.Map{
		"memories": memoryResponses,
		"pagination": fiber.Map{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetMemory returns a single memory by ID
// GET /api/v1/memories/:id
func (h *MemoryHandler) GetMemory(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}
	memoryIDParam := c.Params("id")

	memoryID, err := primitive.ObjectIDFromHex(memoryIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid memory ID",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	memory, err := h.memoryStorageService.GetMemory(ctx, userID, memoryID)
	if err != nil {
		log.Printf("❌ [MEMORY-API] Failed to get memory: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Memory not found",
		})
	}

	return c.JSON(buildMemoryResponse(*memory))
}

// CreateMemory creates a new manual memory
// POST /api/v1/memories
type CreateMemoryRequest struct {
	Content  string   `json:"content"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
}

func (h *MemoryHandler) CreateMemory(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	var req CreateMemoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request (SECURITY: input validation)
	const MaxMemoryContentLength = 10000 // 10KB per memory
	const MaxTagCount = 20
	const MaxTagLength = 50

	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Content is required",
		})
	}

	if len(req.Content) > MaxMemoryContentLength {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Content must be less than 10,000 characters",
		})
	}

	if req.Category == "" {
		req.Category = "context"
	}

	// Sanitize and validate tags
	if len(req.Tags) > MaxTagCount {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Maximum 20 tags allowed",
		})
	}

	sanitizedTags := make([]string, 0, len(req.Tags))
	for _, tag := range req.Tags {
		sanitized := sanitizeTag(strings.TrimSpace(tag))
		if sanitized != "" && len(sanitized) <= MaxTagLength {
			sanitizedTags = append(sanitizedTags, sanitized)
		}
	}
	req.Tags = sanitizedTags

	// Validate category
	validCategories := map[string]bool{
		"personal_info": true,
		"preferences":   true,
		"context":       true,
		"fact":          true,
		"instruction":   true,
	}
	if !validCategories[req.Category] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid category. Must be one of: personal_info, preferences, context, fact, instruction",
		})
	}

	if req.Tags == nil {
		req.Tags = []string{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create memory with default engagement score of 0.5 (manually created)
	memory, err := h.memoryStorageService.CreateMemory(
		ctx,
		userID,
		req.Content,
		req.Category,
		req.Tags,
		0.5, // Default engagement for manual memories
		"",  // No conversation ID for manual memories
	)
	if err != nil {
		log.Printf("❌ [MEMORY-API] Failed to create memory: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create memory",
		})
	}

	// Decrypt for response
	decryptedMemory, err := h.memoryStorageService.GetMemory(ctx, userID, memory.ID)
	if err != nil {
		log.Printf("❌ [MEMORY-API] Failed to decrypt memory: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Memory created but failed to retrieve",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(buildMemoryResponse(*decryptedMemory))
}

// UpdateMemory updates an existing memory
// PUT /api/v1/memories/:id
type UpdateMemoryRequest struct {
	Content  *string   `json:"content,omitempty"`
	Category *string   `json:"category,omitempty"`
	Tags     *[]string `json:"tags,omitempty"`
}

func (h *MemoryHandler) UpdateMemory(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}
	memoryIDParam := c.Params("id")

	memoryID, err := primitive.ObjectIDFromHex(memoryIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid memory ID",
		})
	}

	var req UpdateMemoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get existing memory to retrieve current values
	existingMemory, err := h.memoryStorageService.GetMemory(ctx, userID, memoryID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Memory not found",
		})
	}

	// Validate and sanitize input
	const MaxMemoryContentLength = 10000
	const MaxTagCount = 20
	const MaxTagLength = 50

	// Update fields
	content := existingMemory.DecryptedContent
	if req.Content != nil {
		if len(*req.Content) > MaxMemoryContentLength {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Content must be less than 10,000 characters",
			})
		}
		content = *req.Content
	}

	category := existingMemory.Category
	if req.Category != nil {
		category = *req.Category
		// Validate category
		validCategories := map[string]bool{
			"personal_info": true,
			"preferences":   true,
			"context":       true,
			"fact":          true,
			"instruction":   true,
		}
		if !validCategories[category] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid category",
			})
		}
	}

	tags := existingMemory.Tags
	if req.Tags != nil {
		if len(*req.Tags) > MaxTagCount {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Maximum 20 tags allowed",
			})
		}

		// Sanitize tags
		sanitizedTags := make([]string, 0, len(*req.Tags))
		for _, tag := range *req.Tags {
			sanitized := sanitizeTag(strings.TrimSpace(tag))
			if sanitized != "" && len(sanitized) <= MaxTagLength {
				sanitizedTags = append(sanitizedTags, sanitized)
			}
		}
		tags = sanitizedTags
	}

	// SECURITY FIX: Use atomic update instead of delete-create to prevent race conditions
	updatedMemory, err := h.memoryStorageService.UpdateMemoryInPlace(
		ctx,
		userID,
		memoryID,
		content,
		category,
		tags,
		existingMemory.SourceEngagement,
		existingMemory.ConversationID,
	)
	if err != nil {
		log.Printf("❌ [MEMORY-API] Failed to update memory: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update memory",
		})
	}

	// Get decrypted version for response
	decryptedMemory, err := h.memoryStorageService.GetMemory(ctx, userID, updatedMemory.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Memory updated but failed to retrieve",
		})
	}

	return c.JSON(buildMemoryResponse(*decryptedMemory))
}

// DeleteMemory permanently deletes a memory
// DELETE /api/v1/memories/:id
func (h *MemoryHandler) DeleteMemory(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}
	memoryIDParam := c.Params("id")

	memoryID, err := primitive.ObjectIDFromHex(memoryIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid memory ID",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = h.memoryStorageService.DeleteMemory(ctx, userID, memoryID)
	if err != nil {
		log.Printf("❌ [MEMORY-API] Failed to delete memory: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Memory not found or already deleted",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Memory deleted successfully",
	})
}

// ArchiveMemory archives a memory
// POST /api/v1/memories/:id/archive
func (h *MemoryHandler) ArchiveMemory(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}
	memoryIDParam := c.Params("id")

	memoryID, err := primitive.ObjectIDFromHex(memoryIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid memory ID",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = h.memoryStorageService.ArchiveMemory(ctx, userID, memoryID)
	if err != nil {
		log.Printf("❌ [MEMORY-API] Failed to archive memory: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Memory not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Memory archived successfully",
	})
}

// UnarchiveMemory restores an archived memory
// POST /api/v1/memories/:id/unarchive
func (h *MemoryHandler) UnarchiveMemory(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}
	memoryIDParam := c.Params("id")

	memoryID, err := primitive.ObjectIDFromHex(memoryIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid memory ID",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = h.memoryStorageService.UnarchiveMemory(ctx, userID, memoryID)
	if err != nil {
		log.Printf("❌ [MEMORY-API] Failed to unarchive memory: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Memory not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Memory restored successfully",
	})
}

// GetMemoryStats returns statistics about user's memories
// GET /api/v1/memories/stats
func (h *MemoryHandler) GetMemoryStats(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stats, err := h.memoryStorageService.GetMemoryStats(ctx, userID)
	if err != nil {
		log.Printf("❌ [MEMORY-API] Failed to get memory stats: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve memory statistics",
		})
	}

	return c.JSON(stats)
}

// TriggerMemoryExtraction manually triggers memory extraction for a conversation
// POST /api/v1/conversations/:id/extract-memories
func (h *MemoryHandler) TriggerMemoryExtraction(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}
	conversationID := c.Params("id")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get conversation messages
	messages := h.chatService.GetConversationMessages(conversationID)
	if len(messages) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Conversation not found or has no messages",
		})
	}

	// Enqueue extraction job
	err := h.memoryExtractionService.EnqueueExtraction(ctx, userID, conversationID, messages)
	if err != nil {
		log.Printf("❌ [MEMORY-API] Failed to trigger extraction: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to trigger memory extraction",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Memory extraction queued successfully",
	})
}

// buildMemoryResponse creates a response object from a decrypted memory
func buildMemoryResponse(mem models.DecryptedMemory) fiber.Map {
	return fiber.Map{
		"id":                 mem.ID.Hex(),
		"content":            mem.DecryptedContent,
		"category":           mem.Category,
		"tags":               mem.Tags,
		"score":              mem.Score,
		"access_count":       mem.AccessCount,
		"last_accessed_at":   mem.LastAccessedAt,
		"is_archived":        mem.IsArchived,
		"archived_at":        mem.ArchivedAt,
		"source_engagement":  mem.SourceEngagement,
		"conversation_id":    mem.ConversationID,
		"created_at":         mem.CreatedAt,
		"updated_at":         mem.UpdatedAt,
		"version":            mem.Version,
	}
}

// sanitizeTag removes potentially dangerous characters from tags
// SECURITY: Prevents NoSQL injection via tag parameters
func sanitizeTag(tag string) string {
	// Only allow alphanumeric characters, hyphens, and underscores
	// This prevents MongoDB operators like $where, $regex, etc.
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	sanitized := reg.ReplaceAllString(tag, "")
	return sanitized
}
