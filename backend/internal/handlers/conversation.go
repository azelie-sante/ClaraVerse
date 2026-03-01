package handlers

import (
	"claraverse/internal/models"
	"claraverse/internal/services"
	"log"

	"github.com/gofiber/fiber/v2"
)

// ConversationHandler handles conversation-related HTTP requests
type ConversationHandler struct {
	chatService    *services.ChatService
	builderService *services.BuilderConversationService
}

// NewConversationHandler creates a new conversation handler
func NewConversationHandler(chatService *services.ChatService, builderService *services.BuilderConversationService) *ConversationHandler {
	return &ConversationHandler{
		chatService:    chatService,
		builderService: builderService,
	}
}

// GetStatus returns the status of a conversation (exists, has files, time until expiration)
// GET /api/conversations/:id/status
func (h *ConversationHandler) GetStatus(c *fiber.Ctx) error {
	conversationID := c.Params("id")

	if conversationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Conversation ID is required",
		})
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	log.Printf("📊 [CONVERSATION] Status check for conversation %s (user: %s)", conversationID, userID)

	// Verify conversation ownership
	if !h.chatService.IsConversationOwner(conversationID, userID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied to this conversation",
		})
	}

	status := h.chatService.GetConversationStatus(conversationID)

	return c.JSON(status)
}

// ListBuilderConversations returns all builder conversations for an agent
// GET /api/agents/:id/conversations
func (h *ConversationHandler) ListBuilderConversations(c *fiber.Ctx) error {
	agentID := c.Params("id")
	if agentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Agent ID is required",
		})
	}

	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	if h.builderService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Builder conversation service not available",
		})
	}

	conversations, err := h.builderService.GetConversationsByAgent(c.Context(), agentID, userID)
	if err != nil {
		log.Printf("❌ Failed to list builder conversations: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list conversations",
		})
	}

	return c.JSON(conversations)
}

// GetBuilderConversation returns a specific builder conversation
// GET /api/agents/:id/conversations/:convId
func (h *ConversationHandler) GetBuilderConversation(c *fiber.Ctx) error {
	conversationID := c.Params("convId")
	if conversationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Conversation ID is required",
		})
	}

	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	if h.builderService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Builder conversation service not available",
		})
	}

	conversation, err := h.builderService.GetConversation(c.Context(), conversationID, userID)
	if err != nil {
		log.Printf("❌ Failed to get builder conversation: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Conversation not found",
		})
	}

	return c.JSON(conversation)
}

// CreateBuilderConversation creates a new builder conversation for an agent
// POST /api/agents/:id/conversations
func (h *ConversationHandler) CreateBuilderConversation(c *fiber.Ctx) error {
	agentID := c.Params("id")
	if agentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Agent ID is required",
		})
	}

	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	if h.builderService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Builder conversation service not available",
		})
	}

	var req struct {
		ModelID string `json:"model_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	conversation, err := h.builderService.CreateConversation(c.Context(), agentID, userID, req.ModelID)
	if err != nil {
		log.Printf("❌ Failed to create builder conversation: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create conversation",
		})
	}

	log.Printf("✅ Created builder conversation %s for agent %s", conversation.ID, agentID)
	return c.Status(fiber.StatusCreated).JSON(conversation)
}

// AddBuilderMessage adds a message to a builder conversation
// POST /api/agents/:id/conversations/:convId/messages
func (h *ConversationHandler) AddBuilderMessage(c *fiber.Ctx) error {
	conversationID := c.Params("convId")
	if conversationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Conversation ID is required",
		})
	}

	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	if h.builderService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Builder conversation service not available",
		})
	}

	var req models.AddMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Role == "" || req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Role and content are required",
		})
	}

	message, err := h.builderService.AddMessage(c.Context(), conversationID, userID, &req)
	if err != nil {
		log.Printf("❌ Failed to add message to conversation: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add message",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(message)
}

// DeleteBuilderConversation deletes a builder conversation
// DELETE /api/agents/:id/conversations/:convId
func (h *ConversationHandler) DeleteBuilderConversation(c *fiber.Ctx) error {
	conversationID := c.Params("convId")
	if conversationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Conversation ID is required",
		})
	}

	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	if h.builderService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Builder conversation service not available",
		})
	}

	if err := h.builderService.DeleteConversation(c.Context(), conversationID, userID); err != nil {
		log.Printf("❌ Failed to delete builder conversation: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete conversation",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Conversation deleted successfully",
	})
}

// GetOrCreateBuilderConversation gets the most recent conversation or creates a new one
// GET /api/agents/:id/conversations/current
func (h *ConversationHandler) GetOrCreateBuilderConversation(c *fiber.Ctx) error {
	agentID := c.Params("id")
	if agentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Agent ID is required",
		})
	}

	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	if h.builderService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Builder conversation service not available",
		})
	}

	modelID := c.Query("model_id", "")

	conversation, err := h.builderService.GetOrCreateConversation(c.Context(), agentID, userID, modelID)
	if err != nil {
		log.Printf("❌ Failed to get/create builder conversation: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get conversation",
		})
	}

	return c.JSON(conversation)
}
