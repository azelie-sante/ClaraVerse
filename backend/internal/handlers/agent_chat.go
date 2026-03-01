package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"claraverse/internal/services"

	"github.com/gofiber/fiber/v2"
)

// AgentChatHandler proxies LLM calls from the desktop daemon.
// The daemon sends OpenAI-compatible requests; this handler resolves the
// model → provider, forwards the request, and returns the response as-is.
type AgentChatHandler struct {
	chatService *services.ChatService
	httpClient  *http.Client
}

// NewAgentChatHandler creates a new agent chat proxy handler.
func NewAgentChatHandler(chatService *services.ChatService) *AgentChatHandler {
	return &AgentChatHandler{
		chatService: chatService,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// Chat proxies an OpenAI-compatible chat completion request to the resolved provider.
// POST /api/agent/chat
// Body: {"model": "...", "messages": [...], "tools": [...]}
func (h *AgentChatHandler) Chat(c *fiber.Ctx) error {
	body := c.Body()
	if len(body) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "empty request body"})
	}

	// Parse the full request so we can rewrite the model field
	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON: " + err.Error()})
	}

	modelID, _ := req["model"].(string)

	// Resolve model → provider config (handles empty model via default fallback)
	cfg, err := h.chatService.GetModelConfig(modelID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to resolve model: " + err.Error()})
	}

	// Rewrite the model field to the resolved model (critical when daemon sends empty model)
	req["model"] = cfg.Model
	rewrittenBody, err := json.Marshal(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to rewrite request"})
	}

	log.Printf("[agent-chat] Resolved model %q → %q (provider: %s)", modelID, cfg.Model, cfg.BaseURL)

	// Forward to provider
	url := cfg.BaseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(c.Context(), "POST", url, bytes.NewReader(rewrittenBody))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create request"})
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "provider request failed: " + err.Error()})
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to read provider response"})
	}

	c.Set("Content-Type", "application/json")
	return c.Status(resp.StatusCode).Send(respBody)
}
