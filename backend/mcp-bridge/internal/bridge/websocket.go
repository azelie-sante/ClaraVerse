package bridge

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/claraverse/mcp-client/internal/auth"
	"github.com/gorilla/websocket"
)

const (
	// preemptiveRefreshBuffer is how long before token expiry to proactively refresh.
	// Checked every heartbeat tick (30s).
	preemptiveRefreshBuffer = 5 * time.Minute
)

// ErrAuthenticationFailed indicates the connection was rejected due to auth issues
var ErrAuthenticationFailed = errors.New("authentication failed")

// Message represents a WebSocket message
type Message struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// ToolCall represents a tool execution request from backend
type ToolCall struct {
	CallID    string                 `json:"call_id"`
	ToolName  string                 `json:"tool_name"`
	Arguments map[string]interface{} `json:"arguments"`
	Timeout   int                    `json:"timeout"`
}

// DeviceInfo holds device credentials for token refresh
type DeviceInfo struct {
	DeviceID     string
	RefreshToken string
}

// ServerCommand represents a server management command from the backend
type ServerCommand struct {
	RequestID string
	Action    string // "add_server", "remove_server", "toggle_server"
	Payload   map[string]interface{}
}

// Bridge manages the WebSocket connection to the backend
type Bridge struct {
	backendURL      string
	authToken       string
	refreshToken    string
	tokenExpiry     int64
	deviceInfo      *DeviceInfo // Device credentials for device token refresh
	conn            *websocket.Conn
	writeChan       chan Message
	stopChan        chan struct{}
	connDone        chan struct{}   // closed when current connection is being torn down
	writeWg         sync.WaitGroup // tracks active writeLoop goroutine
	reconnectDelay  time.Duration
	maxReconnect    time.Duration
	connected       bool
	authFailed      bool // Set to true when auth fails to prevent infinite retries
	mutex           sync.RWMutex
	ackChan         chan error // Signals RegisterTools that the backend has ACK'd (nil) or errored
	onToolCall      func(ToolCall)
	onTokenRefresh  func(accessToken, refreshToken string, expiry int64) // Callback to save refreshed tokens
	onReconnect     func() // Callback to re-register tools after reconnection
	onServerCommand func(ServerCommand) error // Callback for server management commands from web UI
	onDisconnect    func()                              // Callback when connection is lost (before reconnect attempt)
	onPersonaSync   func(payload map[string]interface{}) // Callback for persona sync from cloud
	onRequestSync   func()                              // Callback when backend requests a full state re-sync
	verbose         bool
}

// NewBridge creates a new WebSocket bridge
func NewBridge(backendURL, authToken, refreshToken string, tokenExpiry int64, verbose bool) *Bridge {
	return &Bridge{
		backendURL:     backendURL,
		authToken:      authToken,
		refreshToken:   refreshToken,
		tokenExpiry:    tokenExpiry,
		writeChan:      make(chan Message, 100),
		stopChan:       make(chan struct{}),
		reconnectDelay: 1 * time.Second,
		maxReconnect:   60 * time.Second,
		verbose:        verbose,
	}
}

// SetDeviceInfo sets device credentials for device token refresh
func (b *Bridge) SetDeviceInfo(deviceID, refreshToken string) {
	b.deviceInfo = &DeviceInfo{
		DeviceID:     deviceID,
		RefreshToken: refreshToken,
	}
}

// SetTokenRefreshHandler sets the callback for when tokens are refreshed
func (b *Bridge) SetTokenRefreshHandler(handler func(accessToken, refreshToken string, expiry int64)) {
	b.onTokenRefresh = handler
}

// SetToolCallHandler sets the callback for tool call events
func (b *Bridge) SetToolCallHandler(handler func(ToolCall)) {
	b.onToolCall = handler
}

// SetReconnectHandler sets the callback for when the bridge reconnects
// This should be used to re-register tools after reconnection
func (b *Bridge) SetReconnectHandler(handler func()) {
	b.onReconnect = handler
}

// SetServerCommandHandler sets the callback for server management commands from the backend
func (b *Bridge) SetServerCommandHandler(handler func(ServerCommand) error) {
	b.onServerCommand = handler
}

// SetDisconnectHandler sets the callback for when the connection is lost (called before reconnect attempt)
func (b *Bridge) SetDisconnectHandler(handler func()) {
	b.onDisconnect = handler
}

// SetPersonaSyncHandler sets the callback for persona sync pushed from the cloud backend
func (b *Bridge) SetPersonaSyncHandler(handler func(payload map[string]interface{})) {
	b.onPersonaSync = handler
}

// SetRequestSyncHandler sets the callback for when the backend requests a full state re-sync
func (b *Bridge) SetRequestSyncHandler(handler func()) {
	b.onRequestSync = handler
}

// SendServerCommandAck sends an acknowledgment for a server management command
func (b *Bridge) SendServerCommandAck(requestID string, success bool, errMsg string) {
	msg := Message{
		Type: "server_command_ack",
		Payload: map[string]interface{}{
			"request_id": requestID,
			"success":    success,
			"error":      errMsg,
		},
	}
	select {
	case b.writeChan <- msg:
	default:
		log.Printf("[Bridge] writeChan full, dropping server_command_ack message")
	}
}

// SendSyncState sends local memories, persona, and skills to the backend after registration.
func (b *Bridge) SendSyncState(memories []map[string]interface{}, persona map[string]interface{}, skills []map[string]interface{}) {
	select {
	case b.writeChan <- Message{
		Type: "sync_state",
		Payload: map[string]interface{}{
			"memories": memories,
			"persona":  persona,
			"skills":   skills,
		},
	}:
	default:
		log.Printf("[Bridge] writeChan full, dropping sync_state message")
	}
}

// SendMemoryUpdate sends newly extracted memories to the backend.
func (b *Bridge) SendMemoryUpdate(memories []map[string]interface{}) {
	select {
	case b.writeChan <- Message{
		Type: "memory_update",
		Payload: map[string]interface{}{
			"memories": memories,
		},
	}:
	default:
		log.Printf("[Bridge] writeChan full, dropping memory_update message")
	}
}

// Connect establishes the WebSocket connection
func (b *Bridge) Connect() error {
	b.mutex.RLock()
	token := b.authToken
	b.mutex.RUnlock()

	url := fmt.Sprintf("%s?token=%s", b.backendURL, token)

	if b.verbose {
		log.Printf("[Bridge] Connecting to %s", b.backendURL)
	}

	conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		// Check if this is an authentication error (401 or 403)
		if resp != nil && (resp.StatusCode == 401 || resp.StatusCode == 403) {
			return fmt.Errorf("%w: %v", ErrAuthenticationFailed, err)
		}
		// Check for auth-related error messages
		errStr := err.Error()
		if strings.Contains(strings.ToLower(errStr), "401") ||
			strings.Contains(strings.ToLower(errStr), "403") ||
			strings.Contains(strings.ToLower(errStr), "unauthorized") ||
			strings.Contains(strings.ToLower(errStr), "authentication") {
			return fmt.Errorf("%w: %v", ErrAuthenticationFailed, err)
		}
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Drain stale messages from writeChan before starting new loops
	drainLoop:
	for {
		select {
		case <-b.writeChan:
		default:
			break drainLoop
		}
	}

	b.mutex.Lock()
	b.conn = conn
	b.connected = true
	b.authFailed = false // Reset auth failed on successful connection
	b.reconnectDelay = 1 * time.Second // Reset reconnect delay on successful connection
	b.connDone = make(chan struct{})
	b.mutex.Unlock()

	log.Println("✅ Connected to backend")

	// Start read and write loops
	b.writeWg.Add(1)
	go b.readLoop()
	go b.writeLoop()

	return nil
}

// ConnectWithRetry connects with automatic retry and exponential backoff
func (b *Bridge) ConnectWithRetry() {
	attempt := 0
	for {
		select {
		case <-b.stopChan:
			return
		default:
		}

		// Check if auth has failed - don't retry if so
		b.mutex.RLock()
		authFailed := b.authFailed
		b.mutex.RUnlock()
		if authFailed {
			log.Println("❌ Authentication failed. Please run: clara_companion login")
			return
		}

		err := b.Connect()
		if err == nil {
			// Successfully reconnected, trigger re-registration
			if b.onReconnect != nil {
				log.Println("📦 Re-registering tools after reconnection...")
				b.onReconnect()
			}
			return
		}

		// Check if this is an auth error
		if errors.Is(err, ErrAuthenticationFailed) {
			// Try to refresh token
			if b.tryRefreshToken() {
				log.Println("🔄 Token refreshed, retrying connection...")
				continue
			}
			// Refresh failed, stop retrying
			b.mutex.Lock()
			b.authFailed = true
			b.mutex.Unlock()
			log.Println("❌ Token expired and refresh failed. Please run: clara_companion login")
			return
		}

		attempt++
		log.Printf("❌ Connection failed (attempt %d): %v", attempt, err)
		log.Printf("🔄 Retrying in %v...", b.reconnectDelay)

		time.Sleep(b.reconnectDelay)

		// Exponential backoff
		b.reconnectDelay = time.Duration(math.Min(
			float64(b.reconnectDelay*2),
			float64(b.maxReconnect),
		))
	}
}

// tryRefreshToken attempts to refresh the auth token
func (b *Bridge) tryRefreshToken() bool {
	log.Println("🔄 Attempting to refresh token...")

	var newToken *auth.TokenResponse
	var err error

	// Try device token refresh first if we have device credentials
	if b.deviceInfo != nil && b.deviceInfo.DeviceID != "" && b.deviceInfo.RefreshToken != "" {
		newToken, err = auth.RefreshDeviceToken(b.backendURL, b.deviceInfo.RefreshToken, b.deviceInfo.DeviceID)
		if err == nil {
			// Update device refresh token
			b.deviceInfo.RefreshToken = newToken.RefreshToken
		}
	} else if b.refreshToken != "" {
		// Fall back to Supabase refresh for legacy tokens
		newToken, err = auth.RefreshToken(b.refreshToken)
	} else {
		log.Println("❌ No refresh credentials available")
		return false
	}

	if err != nil {
		log.Printf("❌ Token refresh failed: %v", err)
		return false
	}

	// Update tokens
	b.mutex.Lock()
	b.authToken = newToken.AccessToken
	b.refreshToken = newToken.RefreshToken
	b.tokenExpiry = time.Now().Unix() + int64(newToken.ExpiresIn)
	b.mutex.Unlock()

	log.Println("✅ Token refreshed successfully")

	// Notify callback to save new tokens
	if b.onTokenRefresh != nil {
		b.onTokenRefresh(newToken.AccessToken, newToken.RefreshToken, b.tokenExpiry)
	}

	return true
}

// readLoop handles incoming messages
func (b *Bridge) readLoop() {
	defer func() {
		b.handleDisconnect()
	}()

	for {
		var msg Message
		err := b.conn.ReadJSON(&msg)
		if err != nil {
			if b.verbose {
				log.Printf("[Bridge] Read error: %v", err)
			}
			return
		}

		b.handleMessage(msg)
	}
}

// writeLoop handles outgoing messages and periodic pings
func (b *Bridge) writeLoop() {
	defer b.writeWg.Done()

	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	// Send WebSocket-level pings every 45s so the backend's 90s read deadline doesn't expire
	pingTicker := time.NewTicker(45 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case msg := <-b.writeChan:
			b.mutex.RLock()
			conn := b.conn
			b.mutex.RUnlock()
			if conn == nil {
				return
			}
			err := conn.WriteJSON(msg)
			if err != nil {
				if b.verbose {
					log.Printf("[Bridge] Write error: %v", err)
				}
				return
			}

		case <-heartbeatTicker.C:
			// Send application-level heartbeat
			if err := b.SendHeartbeat(); err != nil {
				return
			}

			// Preemptive token refresh: check if token is expiring soon
			b.mutex.RLock()
			expiry := b.tokenExpiry
			b.mutex.RUnlock()
			if auth.IsTokenExpiringSoon(expiry, preemptiveRefreshBuffer) {
				log.Println("🔄 Token expiring soon, refreshing preemptively...")
				b.tryRefreshToken()
			}

		case <-pingTicker.C:
			// Send WebSocket-level ping to keep backend read deadline alive
			b.mutex.RLock()
			conn := b.conn
			b.mutex.RUnlock()
			if conn == nil {
				return
			}
			if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second)); err != nil {
				if b.verbose {
					log.Printf("[Bridge] Ping write error: %v", err)
				}
				return
			}

		case <-b.connDone:
			return

		case <-b.stopChan:
			return
		}
	}
}

// handleMessage processes incoming messages
func (b *Bridge) handleMessage(msg Message) {
	if b.verbose {
		log.Printf("[Bridge] Received: %s", msg.Type)
	}

	switch msg.Type {
	case "ack":
		log.Printf("✅ Registration acknowledged")
		if status, ok := msg.Payload["status"].(string); ok {
			log.Printf("   Status: %s", status)
		}
		if toolsReg, ok := msg.Payload["tools_registered"].(float64); ok {
			log.Printf("   Tools registered: %.0f", toolsReg)
		}
		// Signal RegisterTools that registration succeeded
		b.mutex.Lock()
		ch := b.ackChan
		b.ackChan = nil
		b.mutex.Unlock()
		if ch != nil {
			ch <- nil
		}

	case "tool_call":
		// Parse tool call (safe assertions to avoid panics on malformed messages)
		callID, _ := msg.Payload["call_id"].(string)
		toolName, _ := msg.Payload["tool_name"].(string)
		if callID == "" || toolName == "" {
			log.Printf("[Bridge] tool_call missing call_id or tool_name, ignoring")
			return
		}
		args, _ := msg.Payload["arguments"].(map[string]interface{})
		timeout, _ := msg.Payload["timeout"].(float64)

		toolCall := ToolCall{
			CallID:    callID,
			ToolName:  toolName,
			Arguments: args,
			Timeout:   int(timeout),
		}

		log.Printf("🔧 Tool call: %s (call_id: %s)", toolName, callID)

		// Call handler if set
		if b.onToolCall != nil {
			b.onToolCall(toolCall)
		}

	case "add_server", "remove_server", "toggle_server":
		// Server management command from web UI via backend
		requestID, _ := msg.Payload["request_id"].(string)
		log.Printf("🔧 Server command: %s (request_id: %s)", msg.Type, requestID)

		cmd := ServerCommand{
			RequestID: requestID,
			Action:    msg.Type,
			Payload:   msg.Payload,
		}

		if b.onServerCommand != nil {
			go func() {
				err := b.onServerCommand(cmd)
				if err != nil {
					log.Printf("❌ Server command failed: %v", err)
					b.SendServerCommandAck(requestID, false, err.Error())
				} else {
					log.Printf("✅ Server command completed: %s", msg.Type)
					b.SendServerCommandAck(requestID, true, "")
				}
			}()
		} else {
			b.SendServerCommandAck(requestID, false, "server command handler not configured")
		}

	case "persona_sync":
		log.Printf("🔄 Persona sync from cloud backend")
		if b.onPersonaSync != nil {
			b.onPersonaSync(msg.Payload)
		}

	case "request_sync":
		log.Printf("🔄 Backend requested full state re-sync")
		if b.onRequestSync != nil {
			go b.onRequestSync()
		}

	case "error":
		errMsg, _ := msg.Payload["message"].(string)
		log.Printf("❌ Error from backend: %s", errMsg)

		// Signal RegisterTools if a registration ACK is pending
		b.mutex.Lock()
		ch := b.ackChan
		b.ackChan = nil
		b.mutex.Unlock()
		if ch != nil {
			ch <- fmt.Errorf("backend error: %s", errMsg)
		}

		// Check if this is an auth error
		if strings.Contains(strings.ToLower(errMsg), "authentication") ||
			strings.Contains(strings.ToLower(errMsg), "unauthorized") ||
			strings.Contains(strings.ToLower(errMsg), "token") {
			b.mutex.Lock()
			b.authFailed = true
			b.mutex.Unlock()
		}

	default:
		if b.verbose {
			log.Printf("[Bridge] Unknown message type: %s", msg.Type)
		}
	}
}

// handleDisconnect handles disconnection and reconnection
func (b *Bridge) handleDisconnect() {
	b.mutex.Lock()
	b.connected = false
	authFailed := b.authFailed
	// Signal any pending RegisterTools call so it doesn't hang for 10s
	ch := b.ackChan
	b.ackChan = nil
	// Signal writeLoop to stop
	if b.connDone != nil {
		close(b.connDone)
	}
	if b.conn != nil {
		b.conn.Close()
	}
	b.mutex.Unlock()

	// Unblock RegisterTools with a disconnect error (outside lock)
	if ch != nil {
		ch <- fmt.Errorf("connection lost during registration")
	}

	// Wait for writeLoop to exit before reconnecting to avoid concurrent writes
	b.writeWg.Wait()

	log.Println("🔌 Disconnected from backend")

	// Notify listeners of disconnect
	if b.onDisconnect != nil {
		b.onDisconnect()
	}

	// Don't attempt reconnect if auth has failed
	if authFailed {
		log.Println("❌ Authentication failed. Please run: clara_companion login")
		return
	}

	log.Println("🔄 Attempting to reconnect...")

	// Reconnect with exponential backoff
	b.ConnectWithRetry()
}

// RegisterTools sends tool registration message with optional server info.
// It blocks until the backend sends an ACK, an error, or a 10s timeout elapses.
func (b *Bridge) RegisterTools(clientID, clientVersion, platform string, tools []interface{}, servers ...[]interface{}) error {
	payload := map[string]interface{}{
		"client_id":      clientID,
		"client_version": clientVersion,
		"platform":       platform,
		"tools":          tools,
	}
	if len(servers) > 0 && servers[0] != nil {
		payload["servers"] = servers[0]
	}

	msg := Message{
		Type:    "register_tools",
		Payload: payload,
	}

	// Create ack channel before sending so handleMessage can signal it
	ackCh := make(chan error, 1)
	b.mutex.Lock()
	b.ackChan = ackCh
	b.mutex.Unlock()

	b.writeChan <- msg

	// Wait for ACK, error, or timeout
	select {
	case err := <-ackCh:
		return err
	case <-time.After(10 * time.Second):
		b.mutex.Lock()
		b.ackChan = nil
		b.mutex.Unlock()
		return fmt.Errorf("registration ACK timeout after 10s")
	}
}

// SendToolResult sends tool execution result back to backend
func (b *Bridge) SendToolResult(callID string, success bool, result, errorMsg string) error {
	msg := Message{
		Type: "tool_result",
		Payload: map[string]interface{}{
			"call_id": callID,
			"success": success,
			"result":  result,
			"error":   errorMsg,
		},
	}

	b.writeChan <- msg
	return nil
}

// SendHeartbeat sends a heartbeat message
func (b *Bridge) SendHeartbeat() error {
	msg := Message{
		Type: "heartbeat",
		Payload: map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	b.writeChan <- msg
	return nil
}

// Close gracefully closes the bridge
func (b *Bridge) Close() error {
	// Send disconnect message
	msg := Message{
		Type:    "disconnect",
		Payload: map[string]interface{}{},
	}
	b.writeChan <- msg

	// Wait a bit for message to send
	time.Sleep(100 * time.Millisecond)

	close(b.stopChan)

	b.mutex.Lock()
	// Unblock any pending RegisterTools call
	ch := b.ackChan
	b.ackChan = nil
	if b.conn != nil {
		defer b.conn.Close()
	}
	b.mutex.Unlock()

	if ch != nil {
		ch <- fmt.Errorf("bridge closed")
	}

	return nil
}

// IsConnected returns whether the bridge is currently connected
func (b *Bridge) IsConnected() bool {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.connected
}
