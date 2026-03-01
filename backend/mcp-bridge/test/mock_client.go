package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

// Message types matching backend protocol
type Message struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

func main() {
	// Command line flags
	backendURL := flag.String("url", "ws://localhost:3001/mcp/connect", "Backend WebSocket URL")
	token := flag.String("token", "", "JWT authentication token (required)")
	flag.Parse()

	if *token == "" {
		log.Fatal("❌ Error: --token is required\n\nUsage: go run mock_client.go --token YOUR_JWT_TOKEN\n")
	}

	// Build connection URL with token
	url := fmt.Sprintf("%s?token=%s", *backendURL, *token)

	log.Printf("🔌 Connecting to backend: %s", *backendURL)

	// Connect to WebSocket
	conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		if resp != nil {
			log.Fatalf("❌ Connection failed: %v (HTTP %d)", err, resp.StatusCode)
		}
		log.Fatalf("❌ Connection failed: %v", err)
	}
	defer conn.Close()

	log.Println("✅ Connected to backend!")

	// Set up signal handling for graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan struct{})

	// Start read loop
	go func() {
		defer close(done)
		for {
			var msg Message
			err := conn.ReadJSON(&msg)
			if err != nil {
				log.Printf("❌ Read error: %v", err)
				return
			}

			handleMessage(msg)
		}
	}()

	// Register test tools
	log.Println("📦 Registering test tools...")
	err = registerTools(conn)
	if err != nil {
		log.Fatalf("❌ Registration failed: %v", err)
	}

	// Send heartbeat every 15 seconds
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	log.Println("✅ Mock client running. Press Ctrl+C to exit.")
	log.Println("💡 Now try chatting in the web browser and check if tools are visible!")

	for {
		select {
		case <-done:
			return

		case <-ticker.C:
			// Send heartbeat
			err := sendHeartbeat(conn)
			if err != nil {
				log.Printf("❌ Heartbeat failed: %v", err)
				return
			}
			log.Println("💓 Heartbeat sent")

		case <-interrupt:
			log.Println("\n🛑 Interrupt received, closing connection...")

			// Send disconnect message
			sendDisconnect(conn)

			// Close connection gracefully
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Printf("Write close error: %v", err)
			}

			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func registerTools(conn *websocket.Conn) error {
	// Create test tools
	tools := []Tool{
		{
			Name:        "mock_echo",
			Description: "A test tool that echoes back the input message",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Message to echo back",
					},
				},
				"required": []string{"message"},
			},
		},
		{
			Name:        "mock_timestamp",
			Description: "A test tool that returns current timestamp",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	// Create registration message
	msg := Message{
		Type: "register_tools",
		Payload: map[string]interface{}{
			"client_id":      "mock-client-" + fmt.Sprintf("%d", time.Now().Unix()),
			"client_version": "1.0.0-test",
			"platform":       "test",
			"tools":          tools,
		},
	}

	log.Printf("📤 Sending registration: %d tools", len(tools))
	err := conn.WriteJSON(msg)
	if err != nil {
		return fmt.Errorf("failed to send registration: %w", err)
	}

	return nil
}

func sendHeartbeat(conn *websocket.Conn) error {
	msg := Message{
		Type: "heartbeat",
		Payload: map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	return conn.WriteJSON(msg)
}

func sendDisconnect(conn *websocket.Conn) error {
	msg := Message{
		Type:    "disconnect",
		Payload: map[string]interface{}{},
	}

	return conn.WriteJSON(msg)
}

func handleMessage(msg Message) {
	switch msg.Type {
	case "ack":
		log.Printf("✅ ACK received: %+v", msg.Payload)
		if status, ok := msg.Payload["status"].(string); ok {
			log.Printf("   Status: %s", status)
		}
		if toolsReg, ok := msg.Payload["tools_registered"].(float64); ok {
			log.Printf("   Tools registered: %.0f", toolsReg)
		}

	case "tool_call":
		log.Printf("🔧 TOOL CALL received: %+v", msg.Payload)
		callID := msg.Payload["call_id"].(string)
		toolName := msg.Payload["tool_name"].(string)
		args := msg.Payload["arguments"]

		log.Printf("   Call ID: %s", callID)
		log.Printf("   Tool: %s", toolName)
		log.Printf("   Arguments: %+v", args)

		// TODO: In a real client, this would execute the MCP tool
		// For now, we just log it
		log.Println("   ⚠️  Note: This mock client cannot execute tools yet")
		log.Println("   💡 To test tool execution, build the full MCP client")

	case "error":
		log.Printf("❌ ERROR received: %+v", msg.Payload)
		if errMsg, ok := msg.Payload["message"].(string); ok {
			log.Printf("   Message: %s", errMsg)
		}

	default:
		log.Printf("📨 Unknown message type: %s", msg.Type)
		log.Printf("   Payload: %+v", msg.Payload)
	}
}
