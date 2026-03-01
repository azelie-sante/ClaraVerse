package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"claraverse/internal/models"
	"claraverse/internal/services"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MCPWebSocketHandler handles MCP client WebSocket connections
type MCPWebSocketHandler struct {
	mcpService    *services.MCPBridgeService
	engramService *services.EngramService
	personaStore  *services.PersonaService
	eventBus      *services.NexusEventBus
	sessionStore  *services.NexusSessionStore
}

// NewMCPWebSocketHandler creates a new MCP WebSocket handler
func NewMCPWebSocketHandler(mcpService *services.MCPBridgeService) *MCPWebSocketHandler {
	return &MCPWebSocketHandler{
		mcpService: mcpService,
	}
}

// SetSyncServices injects the Nexus services needed for TUI ↔ cloud sync.
// Called after Nexus services are initialized (requires MongoDB).
func (h *MCPWebSocketHandler) SetSyncServices(
	engramService *services.EngramService,
	personaStore *services.PersonaService,
	eventBus *services.NexusEventBus,
	sessionStore *services.NexusSessionStore,
) {
	h.engramService = engramService
	h.personaStore = personaStore
	h.eventBus = eventBus
	h.sessionStore = sessionStore
}

// HandleConnection handles incoming MCP client WebSocket connections
func (h *MCPWebSocketHandler) HandleConnection(c *websocket.Conn) {
	// Get user from fiber context (set by auth middleware)
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		userID = ""
	}

	if userID == "" || userID == "anonymous" {
		log.Printf("❌ MCP connection rejected: no authenticated user")
		c.WriteJSON(fiber.Map{
			"type": "error",
			"payload": map[string]interface{}{
				"message": "Authentication required",
			},
		})
		c.Close()
		return
	}

	authType := "unknown"
	if _, ok := c.Locals("auth_type").(string); ok {
		authType = c.Locals("auth_type").(string)
	}
	log.Printf("[MCP-WS] Connection: user=%s auth_type=%s", userID, authType)

	// Set read deadline so hung connections are detected.
	// Reset on every successful read or pong.
	const readTimeout = 90 * time.Second
	c.SetReadDeadline(time.Now().Add(readTimeout))
	c.SetPongHandler(func(string) error {
		c.SetReadDeadline(time.Now().Add(readTimeout))
		return nil
	})

	var mcpConn *models.MCPConnection
	var clientID string

	// Read loop
	for {
		var msg models.MCPClientMessage
		err := c.ReadJSON(&msg)
		if err != nil {
			if mcpConn != nil {
				log.Printf("[MCP-WS] Disconnected: user=%s client=%s err=%v", userID, clientID, err)
				h.mcpService.DisconnectClient(clientID)
			}
			break
		}

		// Reset read deadline on each successful message
		c.SetReadDeadline(time.Now().Add(readTimeout))

		switch msg.Type {
		case "register_tools":
			// Parse registration payload
			regData, err := json.Marshal(msg.Payload)
			if err != nil {
				log.Printf("[MCP-WS] Failed to marshal registration payload: %v", err)
				continue
			}

			var registration models.MCPToolRegistration
			err = json.Unmarshal(regData, &registration)
			if err != nil {
				log.Printf("[MCP-WS] Failed to unmarshal registration: %v", err)
				c.WriteJSON(models.MCPServerMessage{
					Type: "error",
					Payload: map[string]interface{}{
						"message": "Invalid registration format",
					},
				})
				continue
			}

			log.Printf("[MCP-WS] register_tools: user=%s tools=%d", userID, len(registration.Tools))

			// Register client
			conn, err := h.mcpService.RegisterClient(userID, &registration)
			log.Printf("[MCP-WS] RegisterClient: user=%s success=%v err=%v", userID, err == nil, err)
			if err != nil {
				c.WriteJSON(models.MCPServerMessage{
					Type: "error",
					Payload: map[string]interface{}{
						"message": fmt.Sprintf("Registration failed: %v", err),
					},
				})
				continue
			}

			mcpConn = conn
			clientID = registration.ClientID

			// Start write loop
			go h.writeLoop(c, conn)

			log.Printf("[MCP-WS] Registered: user=%s client=%s tools=%d", userID, clientID, len(registration.Tools))

		case "tool_result":
			// Handle tool execution result
			resultData, err := json.Marshal(msg.Payload)
			if err != nil {
				log.Printf("Failed to marshal tool result: %v", err)
				continue
			}

			var result models.MCPToolResult
			err = json.Unmarshal(resultData, &result)
			if err != nil {
				log.Printf("Failed to unmarshal tool result: %v", err)
				continue
			}

			// Log execution for audit
			h.mcpService.LogToolExecution(userID, "", "", result.Success, result.Error)

			log.Printf("Tool result received: call_id=%s, success=%v", result.CallID, result.Success)

			// Forward result to pending result channel
			if conn, exists := h.mcpService.GetConnection(clientID); exists {
				if resultChan, pending := conn.PendingResults[result.CallID]; pending {
					// Non-blocking send to result channel
					select {
					case resultChan <- result:
						log.Printf("✅ Tool result forwarded to waiting channel: %s", result.CallID)
					default:
						log.Printf("⚠️  Result channel full or closed for call_id: %s", result.CallID)
					}
				} else {
					log.Printf("⚠️  No pending result channel for call_id: %s", result.CallID)
				}
			}

		case "server_command_ack":
			// Handle server management command acknowledgment
			resultData, err := json.Marshal(msg.Payload)
			if err != nil {
				log.Printf("Failed to marshal server command ack: %v", err)
				continue
			}

			var result models.MCPServerCommandResult
			err = json.Unmarshal(resultData, &result)
			if err != nil {
				log.Printf("Failed to unmarshal server command ack: %v", err)
				continue
			}

			log.Printf("Server command ack: request_id=%s, success=%v", result.RequestID, result.Success)

			// Forward result to pending command channel
			if conn, exists := h.mcpService.GetConnection(clientID); exists {
				if cmdChan, pending := conn.PendingCommands[result.RequestID]; pending {
					select {
					case cmdChan <- result:
						log.Printf("✅ Server command ack forwarded: %s", result.RequestID)
					default:
						log.Printf("⚠️  Command channel full or closed for request_id: %s", result.RequestID)
					}
				}
			}

		case "sync_state":
			// TUI is sending its full local state (memories, persona, skills) after connecting
			log.Printf("[MCP-SYNC] Received sync_state from TUI: userID=%s, engramSvc=%v, payloadKeys=%d",
				userID, h.engramService != nil, len(msg.Payload))
			if userID != "" {
				h.handleSyncState(userID, msg.Payload)
			}

		case "memory_update":
			// TUI is pushing newly extracted memories
			if userID != "" {
				h.handleMemoryUpdate(userID, msg.Payload)
			}

		case "heartbeat":
			// Update heartbeat
			if clientID != "" {
				err := h.mcpService.UpdateHeartbeat(clientID)
				if err != nil {
					log.Printf("Failed to update heartbeat: %v", err)
				}
			}

		case "disconnect":
			// Client is gracefully disconnecting
			if clientID != "" {
				h.mcpService.DisconnectClient(clientID)
			}
			c.Close()
			return

		default:
			log.Printf("Unknown message type from MCP client: %s", msg.Type)
			c.WriteJSON(models.MCPServerMessage{
				Type: "error",
				Payload: map[string]interface{}{
					"message": "Unknown message type",
				},
			})
		}
	}
}

// writeLoop handles outgoing messages to the MCP client
func (h *MCPWebSocketHandler) writeLoop(c *websocket.Conn, conn *models.MCPConnection) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-conn.WriteChan:
			if !ok {
				// Channel closed
				return
			}

			err := c.WriteJSON(msg)
			if err != nil {
				log.Printf("Failed to write message to MCP client: %v", err)
				return
			}

		case <-conn.StopChan:
			// Stop signal received
			return

		case <-ticker.C:
			// Send ping to keep connection alive
			err := c.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				log.Printf("Failed to send ping to MCP client: %v", err)
				return
			}
		}
	}
}

// resolveSessionID looks up the user's current Nexus session ID.
// Returns a zero ObjectID if the session store is not set or lookup fails.
func (h *MCPWebSocketHandler) resolveSessionID(ctx context.Context, userID string) primitive.ObjectID {
	if h.sessionStore == nil {
		return primitive.NilObjectID
	}
	session, err := h.sessionStore.GetByUser(ctx, userID)
	if err != nil || session == nil {
		return primitive.NilObjectID
	}
	return session.ID
}

// memoryTypeMapping maps local TUI memory types to cloud engram types.
var memoryTypeMapping = map[string]string{
	"fact":       "user_fact",
	"preference": "user_fact",
	"context":    "status_log",
	"skill_note": "daemon_output",
}

// handleSyncState processes the full local state from the TUI daemon.
// It upserts memories as engrams, persona facts, and skills.
func (h *MCPWebSocketHandler) handleSyncState(userID string, payload map[string]interface{}) {
	ctx := context.Background()

	// Look up user's session to attach SessionID to engrams
	sessionID := h.resolveSessionID(ctx, userID)

	log.Printf("[MCP-SYNC] handleSyncState: memories=%v, persona=%v, skills=%v",
		payload["memories"] != nil, payload["persona"] != nil, payload["skills"] != nil)

	// Process memories → engrams
	if memoriesRaw, ok := payload["memories"]; ok {
		h.upsertMemoriesAsEngrams(ctx, userID, sessionID, memoriesRaw)
	}

	// Process persona → persona facts
	if personaRaw, ok := payload["persona"]; ok {
		h.upsertPersonaFromTUI(ctx, userID, personaRaw)
	}

	// Process skills → engrams
	if skillsRaw, ok := payload["skills"]; ok {
		h.upsertSkillsAsEngrams(ctx, userID, sessionID, skillsRaw)
	}

	log.Printf("📥 Processed sync_state from TUI for user %s", userID)

	// Notify Nexus frontends that bridge state has updated
	if h.eventBus != nil {
		h.eventBus.Publish(userID, services.NexusEvent{
			Type: "bridge_state_updated",
			Data: map[string]interface{}{"bridge_connected": true},
		})
	}
}

// handleMemoryUpdate processes incremental memory updates from the TUI daemon.
func (h *MCPWebSocketHandler) handleMemoryUpdate(userID string, payload map[string]interface{}) {
	ctx := context.Background()

	sessionID := h.resolveSessionID(ctx, userID)

	if memoriesRaw, ok := payload["memories"]; ok {
		h.upsertMemoriesAsEngrams(ctx, userID, sessionID, memoriesRaw)
	}

	// Notify Nexus frontends
	if h.eventBus != nil {
		h.eventBus.Publish(userID, services.NexusEvent{
			Type: "bridge_state_updated",
			Data: map[string]interface{}{"memory_updated": true},
		})
	}
}

// upsertMemoriesAsEngrams converts TUI memories to engram entries, deduplicating by key.
func (h *MCPWebSocketHandler) upsertMemoriesAsEngrams(ctx context.Context, userID string, sessionID primitive.ObjectID, memoriesRaw interface{}) {
	if h.engramService == nil {
		return
	}

	data, err := json.Marshal(memoriesRaw)
	if err != nil {
		return
	}

	var memories []struct {
		ID      int64   `json:"id"`
		Type    string  `json:"type"`
		Key     string  `json:"key"`
		Content string  `json:"content"`
		Summary string  `json:"summary"`
		Source  string  `json:"source"`
		Confidence float64 `json:"confidence"`
	}
	if err := json.Unmarshal(data, &memories); err != nil {
		log.Printf("[MCP-SYNC] Failed to parse memories: %v (raw: %.200s)", err, string(data))
		return
	}

	log.Printf("[MCP-SYNC] Parsed %d memories from TUI for user %s", len(memories), userID)

	upserted := 0
	for _, m := range memories {
		engramKey := fmt.Sprintf("tui_memory_%d", m.ID)
		engramType := memoryTypeMapping[m.Type]
		if engramType == "" {
			engramType = "user_fact"
		}

		// Check if engram with this key already exists for this user
		existing, _ := h.engramService.FindByKey(ctx, userID, engramKey)
		if existing != nil {
			// Already synced — skip
			continue
		}

		entry := &models.EngramEntry{
			SessionID: sessionID,
			UserID:    userID,
			Type:      engramType,
			Key:       engramKey,
			Value:     m.Content,
			Summary:   m.Summary,
			Source:    "tui_sync",
		}
		if err := h.engramService.Write(ctx, entry); err != nil {
			log.Printf("[MCP-SYNC] Failed to write engram for memory %d: %v", m.ID, err)
			continue
		}
		upserted++
	}
	if upserted > 0 {
		log.Printf("[MCP-SYNC] Upserted %d memories as engrams for user %s", upserted, userID)
	}
}

// upsertPersonaFromTUI converts TUI persona (traits, user_facts, preferences) to PersonaFact entries.
func (h *MCPWebSocketHandler) upsertPersonaFromTUI(ctx context.Context, userID string, personaRaw interface{}) {
	if h.personaStore == nil {
		return
	}

	data, err := json.Marshal(personaRaw)
	if err != nil {
		return
	}

	var p struct {
		Traits      []string          `json:"traits"`
		UserFacts   map[string]string `json:"user_facts"`
		Preferences map[string]string `json:"preferences"`
	}
	if err := json.Unmarshal(data, &p); err != nil {
		log.Printf("[MCP-SYNC] Failed to parse persona: %v", err)
		return
	}

	// Get existing facts to deduplicate
	existing, _ := h.personaStore.GetAll(ctx, userID)
	existingSet := make(map[string]bool)
	for _, f := range existing {
		existingSet[strings.ToLower(f.Content)] = true
	}

	created := 0
	// Traits → personality category
	for _, trait := range p.Traits {
		if existingSet[strings.ToLower(trait)] {
			continue
		}
		_ = h.personaStore.Create(ctx, &models.PersonaFact{
			UserID:     userID,
			Category:   "personality",
			Content:    trait,
			Confidence: 0.8,
			Source:     "tui_sync",
		})
		created++
	}

	// UserFacts → expertise category
	for key, val := range p.UserFacts {
		content := fmt.Sprintf("%s: %s", key, val)
		if existingSet[strings.ToLower(content)] {
			continue
		}
		_ = h.personaStore.Create(ctx, &models.PersonaFact{
			UserID:     userID,
			Category:   "expertise",
			Content:    content,
			Confidence: 0.8,
			Source:     "tui_sync",
		})
		created++
	}

	// Preferences → communication category
	for key, val := range p.Preferences {
		content := fmt.Sprintf("%s: %s", key, val)
		if existingSet[strings.ToLower(content)] {
			continue
		}
		_ = h.personaStore.Create(ctx, &models.PersonaFact{
			UserID:     userID,
			Category:   "communication",
			Content:    content,
			Confidence: 0.8,
			Source:     "tui_sync",
		})
		created++
	}

	if created > 0 {
		log.Printf("[MCP-SYNC] Created %d persona facts from TUI for user %s", created, userID)
	}
}

// upsertSkillsAsEngrams converts TUI skills to engram entries.
func (h *MCPWebSocketHandler) upsertSkillsAsEngrams(ctx context.Context, userID string, sessionID primitive.ObjectID, skillsRaw interface{}) {
	if h.engramService == nil {
		return
	}

	data, err := json.Marshal(skillsRaw)
	if err != nil {
		return
	}

	var skills []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Content     string `json:"content"`
	}
	if err := json.Unmarshal(data, &skills); err != nil {
		log.Printf("[MCP-SYNC] Failed to parse skills: %v", err)
		return
	}

	upserted := 0
	for _, s := range skills {
		engramKey := fmt.Sprintf("tui_skill_%s", s.Name)

		existing, _ := h.engramService.FindByKey(ctx, userID, engramKey)
		if existing != nil {
			continue
		}

		entry := &models.EngramEntry{
			SessionID: sessionID,
			UserID:    userID,
			Type:      "daemon_output",
			Key:       engramKey,
			Value:     s.Content,
			Summary:   s.Description,
			Source:    "tui_skill",
		}
		if err := h.engramService.Write(ctx, entry); err != nil {
			log.Printf("[MCP-SYNC] Failed to write engram for skill %s: %v", s.Name, err)
			continue
		}
		upserted++
	}
	if upserted > 0 {
		log.Printf("[MCP-SYNC] Upserted %d skills as engrams for user %s", upserted, userID)
	}
}
