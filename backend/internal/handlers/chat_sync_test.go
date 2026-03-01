package handlers

import (
	"bytes"
	"claraverse/internal/models"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// Mock user middleware for testing
func mockAuthMiddleware(userID string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return c.Next()
	}
}

func TestChatSyncHandler_CreateOrUpdate_Validation(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		body           interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing user ID",
			userID:         "",
			body:           models.CreateChatRequest{ID: "chat-1", Title: "Test"},
			expectedStatus: fiber.StatusUnauthorized,
			expectedError:  "Authentication required",
		},
		{
			name:           "empty chat ID",
			userID:         "user-123",
			body:           models.CreateChatRequest{ID: "", Title: "Test"},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "Chat ID is required",
		},
		{
			name:           "invalid JSON body",
			userID:         "user-123",
			body:           "not json",
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			// Add mock auth middleware
			app.Use(mockAuthMiddleware(tt.userID))

			// Create handler with nil service (will fail on actual operations but validation should happen first)
			handler := &ChatSyncHandler{service: nil}
			app.Post("/chats", handler.CreateOrUpdate)

			var body []byte
			var err error
			if str, ok := tt.body.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("Failed to marshal body: %v", err)
				}
			}

			req := httptest.NewRequest("POST", "/chats", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Check error message
			respBody, _ := io.ReadAll(resp.Body)
			var result map[string]string
			json.Unmarshal(respBody, &result)

			if result["error"] != tt.expectedError {
				t.Errorf("Expected error %q, got %q", tt.expectedError, result["error"])
			}
		})
	}
}

func TestChatSyncHandler_Get_Validation(t *testing.T) {
	// Test only auth validation - service calls will panic with nil service
	app := fiber.New()
	app.Use(mockAuthMiddleware(""))

	handler := &ChatSyncHandler{service: nil}
	app.Get("/chats/:id", handler.Get)

	req := httptest.NewRequest("GET", "/chats/chat-123", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(respBody, &result)

	if result["error"] != "Authentication required" {
		t.Errorf("Expected error %q, got %q", "Authentication required", result["error"])
	}
}

func TestChatSyncHandler_List_Validation(t *testing.T) {
	// Test only auth validation - service calls will panic with nil service
	app := fiber.New()
	app.Use(mockAuthMiddleware(""))

	handler := &ChatSyncHandler{service: nil}
	app.Get("/chats", handler.List)

	req := httptest.NewRequest("GET", "/chats", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestChatSyncHandler_Update_Validation(t *testing.T) {
	app := fiber.New()
	app.Use(mockAuthMiddleware(""))

	handler := &ChatSyncHandler{service: nil}
	app.Put("/chats/:id", handler.Update)

	req := httptest.NewRequest("PUT", "/chats/chat-123", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req, -1)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestChatSyncHandler_Delete_Validation(t *testing.T) {
	app := fiber.New()
	app.Use(mockAuthMiddleware(""))

	handler := &ChatSyncHandler{service: nil}
	app.Delete("/chats/:id", handler.Delete)

	req := httptest.NewRequest("DELETE", "/chats/chat-123", nil)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestChatSyncHandler_BulkSync_Validation(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		body           interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing user ID",
			userID:         "",
			body:           models.BulkSyncRequest{Chats: []models.CreateChatRequest{}},
			expectedStatus: fiber.StatusUnauthorized,
			expectedError:  "Authentication required",
		},
		{
			name:           "empty chats array",
			userID:         "user-123",
			body:           models.BulkSyncRequest{Chats: []models.CreateChatRequest{}},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "No chats provided",
		},
		{
			name:   "too many chats",
			userID: "user-123",
			body: func() models.BulkSyncRequest {
				chats := make([]models.CreateChatRequest, 101)
				for i := range chats {
					chats[i] = models.CreateChatRequest{ID: "chat-" + string(rune(i))}
				}
				return models.BulkSyncRequest{Chats: chats}
			}(),
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "Maximum 100 chats per bulk sync",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(mockAuthMiddleware(tt.userID))

			handler := &ChatSyncHandler{service: nil}
			app.Post("/chats/sync", handler.BulkSync)

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/chats/sync", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedError != "" {
				respBody, _ := io.ReadAll(resp.Body)
				var result map[string]string
				json.Unmarshal(respBody, &result)
				if result["error"] != tt.expectedError {
					t.Errorf("Expected error %q, got %q", tt.expectedError, result["error"])
				}
			}
		})
	}
}

func TestChatSyncHandler_SyncAll_Validation(t *testing.T) {
	app := fiber.New()
	app.Use(mockAuthMiddleware(""))

	handler := &ChatSyncHandler{service: nil}
	app.Get("/chats/sync", handler.SyncAll)

	req := httptest.NewRequest("GET", "/chats/sync", nil)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestChatSyncHandler_AddMessage_Validation(t *testing.T) {
	// Test only auth validation
	app := fiber.New()
	app.Use(mockAuthMiddleware(""))

	handler := &ChatSyncHandler{service: nil}
	app.Post("/chats/:id/messages", handler.AddMessage)

	body, _ := json.Marshal(models.ChatAddMessageRequest{})
	req := httptest.NewRequest("POST", "/chats/chat-123/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestChatSyncHandler_DeleteAll_Validation(t *testing.T) {
	app := fiber.New()
	app.Use(mockAuthMiddleware(""))

	handler := &ChatSyncHandler{service: nil}
	app.Delete("/chats", handler.DeleteAll)

	req := httptest.NewRequest("DELETE", "/chats", nil)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

// Test request body parsing
func TestRequestBodyParsing(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldParse bool
	}{
		{
			name:        "valid create chat request",
			input:       `{"id":"chat-1","title":"Test","messages":[{"id":"msg-1","role":"user","content":"Hello","timestamp":1700000000000}]}`,
			shouldParse: true,
		},
		{
			name:        "valid update request",
			input:       `{"title":"New Title","version":5}`,
			shouldParse: true,
		},
		{
			name:        "valid bulk sync request",
			input:       `{"chats":[{"id":"chat-1","title":"Test","messages":[]}]}`,
			shouldParse: true,
		},
		{
			name:        "request with attachments",
			input:       `{"id":"chat-1","title":"Test","messages":[{"id":"msg-1","role":"user","content":"Hello","timestamp":1700000000000,"attachments":[{"id":"att-1","name":"file.pdf","type":"application/pdf","size":1024}]}]}`,
			shouldParse: true,
		},
		{
			name:        "request with starred",
			input:       `{"id":"chat-1","title":"Test","messages":[],"is_starred":true}`,
			shouldParse: true,
		},
		{
			name:        "malformed JSON",
			input:       `{"id":"chat-1"`,
			shouldParse: false,
		},
		{
			name:        "empty object",
			input:       `{}`,
			shouldParse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req models.CreateChatRequest
			err := json.Unmarshal([]byte(tt.input), &req)

			if tt.shouldParse && err != nil {
				t.Errorf("Expected to parse successfully, got error: %v", err)
			}
			if !tt.shouldParse && err == nil {
				t.Error("Expected parse error, got nil")
			}
		})
	}
}

// Test response format
func TestResponseFormat(t *testing.T) {
	// Test ChatResponse JSON format
	response := models.ChatResponse{
		ID:        "chat-123",
		Title:     "Test Chat",
		Messages:  []models.ChatMessage{},
		IsStarred: true,
		Version:   1,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(jsonData, &parsed)

	// Check snake_case field names
	if _, ok := parsed["is_starred"]; !ok {
		t.Error("Expected is_starred field in JSON")
	}
	if _, ok := parsed["created_at"]; !ok {
		t.Error("Expected created_at field in JSON")
	}

	// Test ChatListResponse JSON format
	listResponse := models.ChatListResponse{
		Chats:      []models.ChatListItem{},
		TotalCount: 10,
		Page:       1,
		PageSize:   20,
		HasMore:    false,
	}

	jsonData, _ = json.Marshal(listResponse)
	json.Unmarshal(jsonData, &parsed)

	if _, ok := parsed["total_count"]; !ok {
		t.Error("Expected total_count field in JSON")
	}
	if _, ok := parsed["page_size"]; !ok {
		t.Error("Expected page_size field in JSON")
	}
	if _, ok := parsed["has_more"]; !ok {
		t.Error("Expected has_more field in JSON")
	}
}
