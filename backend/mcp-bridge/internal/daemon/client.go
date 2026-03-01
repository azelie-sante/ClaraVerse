package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"
	"sync"
	"time"
)

// Client connects to the daemon via IPC
type Client struct {
	conn     net.Conn
	encoder  *json.Encoder
	decoder  *json.Decoder
	handlers ClientHandlers
	mu       sync.Mutex
	closed   bool
}

// ClientHandlers contains callbacks for daemon events
type ClientHandlers struct {
	OnStatus   func(StatusPayload)
	OnActivity func(Activity)
	OnError    func(string)

	// Agent brain
	OnChatResponse   func(json.RawMessage)
	OnAgentEvent     func(json.RawMessage)
	OnSessionList    func(json.RawMessage)
	OnSessionCreated func(json.RawMessage)
	OnModelList      func(json.RawMessage)

	// Tasks
	OnTaskSubmitted func(json.RawMessage)
	OnTaskList      func(json.RawMessage)
	OnTaskDetail    func(json.RawMessage)
	OnTaskEvent     func(json.RawMessage)
}

// Connect connects to the running daemon
func Connect() (*Client, error) {
	socketPath := GetSocketPath()
	var conn net.Conn
	var err error

	if runtime.GOOS == "windows" {
		// Read port from file
		data, err := os.ReadFile(socketPath)
		if err != nil {
			return nil, fmt.Errorf("daemon not running (no port file)")
		}
		conn, err = net.DialTimeout("tcp", string(data), 5*time.Second)
		if err != nil {
			return nil, fmt.Errorf("daemon not responding: %w", err)
		}
	} else {
		conn, err = net.DialTimeout("unix", socketPath, 5*time.Second)
		if err != nil {
			return nil, fmt.Errorf("daemon not running: %w", err)
		}
	}

	client := &Client{
		conn:    conn,
		encoder: json.NewEncoder(conn),
		decoder: json.NewDecoder(conn),
	}

	return client, nil
}

// SetHandlers sets the event handlers
func (c *Client) SetHandlers(h ClientHandlers) {
	c.handlers = h
}

// Listen starts listening for daemon messages (blocking)
func (c *Client) Listen() {
	for {
		var msg IPCMessage
		if err := c.decoder.Decode(&msg); err != nil {
			if !c.closed {
				if c.handlers.OnError != nil {
					c.handlers.OnError("connection lost")
				}
			}
			return
		}

		switch msg.Type {
		case MsgTypeStatus:
			if c.handlers.OnStatus != nil {
				var status StatusPayload
				if err := json.Unmarshal(msg.Payload, &status); err == nil {
					c.handlers.OnStatus(status)
				}
			}

		case MsgTypeActivity:
			if c.handlers.OnActivity != nil {
				var activity Activity
				if err := json.Unmarshal(msg.Payload, &activity); err == nil {
					c.handlers.OnActivity(activity)
				}
			}

		case MsgTypeError:
			if c.handlers.OnError != nil {
				var errPayload map[string]string
				if err := json.Unmarshal(msg.Payload, &errPayload); err == nil {
					c.handlers.OnError(errPayload["error"])
				}
			}

		case MsgTypePong:
			// Ping response, ignore

		case MsgTypeOK:
			// Operation succeeded, ignore

		// Agent brain messages
		case MsgTypeChatResponse:
			if c.handlers.OnChatResponse != nil {
				c.handlers.OnChatResponse(msg.Payload)
			}
		case MsgTypeAgentEvent:
			if c.handlers.OnAgentEvent != nil {
				c.handlers.OnAgentEvent(msg.Payload)
			}
		case MsgTypeSessionList:
			if c.handlers.OnSessionList != nil {
				c.handlers.OnSessionList(msg.Payload)
			}
		case MsgTypeSessionCreated:
			if c.handlers.OnSessionCreated != nil {
				c.handlers.OnSessionCreated(msg.Payload)
			}
		case MsgTypeModelList:
			if c.handlers.OnModelList != nil {
				c.handlers.OnModelList(msg.Payload)
			}

		// Task messages
		case MsgTypeTaskSubmitted:
			if c.handlers.OnTaskSubmitted != nil {
				c.handlers.OnTaskSubmitted(msg.Payload)
			}
		case MsgTypeTaskList:
			if c.handlers.OnTaskList != nil {
				c.handlers.OnTaskList(msg.Payload)
			}
		case MsgTypeTaskDetail:
			if c.handlers.OnTaskDetail != nil {
				c.handlers.OnTaskDetail(msg.Payload)
			}
		case MsgTypeTaskEvent:
			if c.handlers.OnTaskEvent != nil {
				c.handlers.OnTaskEvent(msg.Payload)
			}
		}
	}
}

// RequestStatus requests current status from daemon
func (c *Client) RequestStatus() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.encoder.Encode(IPCMessage{Type: MsgTypeStatus})
}

// Ping sends a ping to check connection
func (c *Client) Ping() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.encoder.Encode(IPCMessage{Type: MsgTypePing})
}

// AddServer requests adding a new server
func (c *Client) AddServer(name, description, command string, args []string) error {
	payload := AddServerPayload{
		Name:        name,
		Description: description,
		Command:     command,
		Args:        args,
	}
	data, _ := json.Marshal(payload)

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.encoder.Encode(IPCMessage{
		Type:    MsgTypeAddServer,
		Payload: data,
	})
}

// RemoveServer requests removing a server
func (c *Client) RemoveServer(name string) error {
	payload := ServerNamePayload{Name: name}
	data, _ := json.Marshal(payload)

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.encoder.Encode(IPCMessage{
		Type:    MsgTypeRemoveServer,
		Payload: data,
	})
}

// ToggleServer requests enabling/disabling a server
func (c *Client) ToggleServer(name string, enabled bool) error {
	payload := ServerNamePayload{Name: name, Enabled: enabled}
	data, _ := json.Marshal(payload)

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.encoder.Encode(IPCMessage{
		Type:    MsgTypeToggleServer,
		Payload: data,
	})
}

// Shutdown requests daemon shutdown
func (c *Client) Shutdown() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.encoder.Encode(IPCMessage{Type: MsgTypeShutdown})
}

// ── Agent brain methods ────────────────────────────────────────────

// SendChatMessage sends a chat message to the daemon
func (c *Client) SendChatMessage(sessionID, content, modelID string) error {
	payload := map[string]string{
		"session_id": sessionID,
		"content":    content,
		"model_id":   modelID,
	}
	data, _ := json.Marshal(payload)
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.encoder.Encode(IPCMessage{Type: MsgTypeChatMessage, Payload: data})
}

// RequestSessionList requests the list of sessions
func (c *Client) RequestSessionList() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.encoder.Encode(IPCMessage{Type: MsgTypeSessionList})
}

// RequestSessionLoad requests loading a session's messages
func (c *Client) RequestSessionLoad(sessionID string) error {
	payload := map[string]string{"session_id": sessionID}
	data, _ := json.Marshal(payload)
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.encoder.Encode(IPCMessage{Type: MsgTypeSessionLoad, Payload: data})
}

// NewSession requests creating a new session
func (c *Client) NewSession(title string) error {
	payload := map[string]string{"title": title}
	data, _ := json.Marshal(payload)
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.encoder.Encode(IPCMessage{Type: MsgTypeSessionNew, Payload: data})
}

// RequestModelList requests the list of available models
func (c *Client) RequestModelList() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.encoder.Encode(IPCMessage{Type: MsgTypeModelList})
}

// ── Task methods ───────────────────────────────────────────────────

// SubmitTask submits a background task
func (c *Client) SubmitTask(content, modelID string) error {
	payload := map[string]string{
		"content":  content,
		"model_id": modelID,
	}
	data, _ := json.Marshal(payload)
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.encoder.Encode(IPCMessage{Type: MsgTypeTaskSubmit, Payload: data})
}

// RequestTaskList requests the list of tasks
func (c *Client) RequestTaskList() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.encoder.Encode(IPCMessage{Type: MsgTypeTaskList})
}

// RequestTaskDetail requests details for a specific task
func (c *Client) RequestTaskDetail(taskID string) error {
	payload := map[string]string{"task_id": taskID}
	data, _ := json.Marshal(payload)
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.encoder.Encode(IPCMessage{Type: MsgTypeTaskDetail, Payload: data})
}

// CancelTask requests cancelling a task
func (c *Client) CancelTask(taskID string) error {
	payload := map[string]string{"task_id": taskID}
	data, _ := json.Marshal(payload)
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.encoder.Encode(IPCMessage{Type: MsgTypeTaskCancel, Payload: data})
}

// Close closes the connection
func (c *Client) Close() error {
	c.closed = true
	return c.conn.Close()
}
