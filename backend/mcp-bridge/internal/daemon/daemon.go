package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/claraverse/mcp-client/internal/bridge"
	"github.com/claraverse/mcp-client/internal/config"
	"github.com/claraverse/mcp-client/internal/coretools"
	"github.com/claraverse/mcp-client/internal/registry"
)

// Daemon manages the background MCP bridge service.
// It connects to the cloud backend via WebSocket, manages MCP servers,
// and routes tool calls through the UnifiedRegistry (core + MCP tools).
type Daemon struct {
	bridge   *bridge.Bridge
	registry *registry.Registry
	config   *config.Config
	listener net.Listener
	clients  map[net.Conn]bool
	clientMu sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	clientID string

	// Unified tool registry (core + MCP)
	unifiedRegistry *coretools.UnifiedRegistry

	// State
	status      string
	backendURL  string
	lastError   error
	activities  []Activity
	activityMu  sync.RWMutex
	maxActivity int
}

// Activity represents a tool execution event
type Activity struct {
	Timestamp  time.Time `json:"timestamp"`
	ToolName   string    `json:"tool_name"`
	Arguments  string    `json:"arguments"`
	Success    bool      `json:"success"`
	Error      string    `json:"error,omitempty"`
	Latency    int64     `json:"latency_ms"`
	ServerName string    `json:"server_name"`
}

// Message types for IPC
type IPCMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Message types
const (
	MsgTypeStatus       = "status"
	MsgTypeActivity     = "activity"
	MsgTypeServerList   = "servers"
	MsgTypeAddServer    = "add_server"
	MsgTypeRemoveServer = "remove_server"
	MsgTypeToggleServer = "toggle_server"
	MsgTypeShutdown     = "shutdown"
	MsgTypePing         = "ping"
	MsgTypePong         = "pong"
	MsgTypeError        = "error"
	MsgTypeOK           = "ok"

	// Agent brain message types (kept for client.go compatibility, daemon no longer handles these)
	MsgTypeChatMessage   = "chat_message"
	MsgTypeChatResponse  = "chat_response"
	MsgTypeAgentEvent    = "agent_event"
	MsgTypeSessionList   = "session_list"
	MsgTypeSessionLoad   = "session_load"
	MsgTypeSessionNew    = "session_new"
	MsgTypeSessionCreated = "session_created"
	MsgTypeModelList     = "model_list"

	// Task message types (kept for client.go compatibility, daemon no longer handles these)
	MsgTypeTaskSubmit    = "task_submit"
	MsgTypeTaskSubmitted = "task_submitted"
	MsgTypeTaskList      = "task_list"
	MsgTypeTaskDetail    = "task_detail"
	MsgTypeTaskCancel    = "task_cancel"
	MsgTypeTaskEvent     = "task_event"
)

// StatusPayload contains current daemon status
type StatusPayload struct {
	Status     string       `json:"status"`
	BackendURL string       `json:"backend_url"`
	Error      string       `json:"error,omitempty"`
	Servers    []ServerInfo `json:"servers"`
	ToolCount  int          `json:"tool_count"`
	UserEmail  string       `json:"user_email,omitempty"`
	UserID     string       `json:"user_id,omitempty"`
	DeviceID   string       `json:"device_id,omitempty"`
	TokenExp   int64        `json:"token_exp,omitempty"`
}

// ServerInfo represents an MCP server
type ServerInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Command     string   `json:"command"`
	Args        []string `json:"args"`
	Enabled     bool     `json:"enabled"`
	Connected   bool     `json:"connected"`
	ToolCount   int      `json:"tool_count"`
	Error       string   `json:"error,omitempty"`
}

// AddServerPayload for adding a new server
type AddServerPayload struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Command     string   `json:"command"`
	Args        []string `json:"args"`
}

// ServerNamePayload for remove/toggle operations
type ServerNamePayload struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled,omitempty"`
}

// GetSocketPath returns the IPC socket path
func GetSocketPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.TempDir(), "clara_companion.port")
	}
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		runtimeDir = os.TempDir()
	}
	return filepath.Join(runtimeDir, "clara_companion.sock")
}

// GetPIDPath returns the PID file path
func GetPIDPath() string {
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		runtimeDir = os.TempDir()
	}
	return filepath.Join(runtimeDir, "clara_companion.pid")
}

// IsRunning checks if a daemon is already running
func IsRunning() bool {
	socketPath := GetSocketPath()

	if runtime.GOOS == "windows" {
		data, err := os.ReadFile(socketPath)
		if err != nil {
			return false
		}
		conn, err := net.DialTimeout("tcp", string(data), time.Second)
		if err != nil {
			return false
		}
		conn.Close()
		return true
	}

	conn, err := net.DialTimeout("unix", socketPath, time.Second)
	if err != nil {
		os.Remove(socketPath)
		return false
	}
	conn.Close()
	return true
}

// NewDaemon creates a new daemon instance
func NewDaemon(cfg *config.Config, verbose bool) (*Daemon, error) {
	ctx, cancel := context.WithCancel(context.Background())

	reg := registry.NewRegistry(verbose)
	b := bridge.NewBridge(cfg.BackendURL, cfg.AuthToken, cfg.RefreshToken, cfg.TokenExpiry, verbose)

	if cfg.Device != nil && cfg.Device.DeviceID != "" && cfg.Device.RefreshToken != "" {
		b.SetDeviceInfo(cfg.Device.DeviceID, cfg.Device.RefreshToken)
	}

	return &Daemon{
		bridge:      b,
		registry:    reg,
		config:      cfg,
		clients:     make(map[net.Conn]bool),
		ctx:         ctx,
		cancel:      cancel,
		clientID:    fmt.Sprintf("daemon-%d", os.Getpid()),
		status:      "starting",
		backendURL:  cfg.BackendURL,
		maxActivity: 100,
	}, nil
}

// Start starts the daemon
func (d *Daemon) Start() error {
	if err := d.setupListener(); err != nil {
		return fmt.Errorf("failed to set up IPC listener: %w", err)
	}

	if err := os.WriteFile(GetPIDPath(), []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		log.Printf("Warning: failed to write PID file: %v", err)
	}

	// Clean up stale browser state from previous sessions
	coretools.CleanupBrowserOnStartup()

	// Auto-launch Chrome with remote debugging — use config if available
	var chromeReady bool
	if d.config.Browser != nil && d.config.Browser.AutoLaunch {
		chromeReady = coretools.EnsureBrowserReadyWithConfig(
			d.config.Browser.ProfilePath, d.config.Browser.Port)
	} else if d.config.Browser == nil {
		chromeReady = coretools.EnsureBrowserReady() // legacy
	} else {
		chromeReady = coretools.IsPortOpen("127.0.0.1", d.config.Browser.Port)
	}
	if !chromeReady {
		d.stripBrowserUrlFromConfig()
	}

	// Start MCP servers
	d.startServers()

	// Auto-configure browser MCP server if Chrome is ready but no browser server configured
	if chromeReady && !d.hasBrowserServer() {
		d.autoStartBrowserMCP()
	}

	// Initialize unified tool registry (core + MCP)
	d.unifiedRegistry = coretools.NewUnifiedRegistry(d.registry)
	log.Printf("[DAEMON] Initialized %d core tools + MCP tools", d.unifiedRegistry.CoreToolCount())

	// Set up bridge handlers
	d.setupBridgeHandlers()

	// Connect to backend
	go d.connectBackend()

	// Accept IPC connections
	go d.acceptConnections()

	// Handle shutdown signals
	d.handleSignals()

	return nil
}

func (d *Daemon) setupListener() error {
	socketPath := GetSocketPath()

	if runtime.GOOS == "windows" {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return err
		}
		d.listener = listener
		addr := listener.Addr().String()
		if err := os.WriteFile(socketPath, []byte(addr), 0644); err != nil {
			listener.Close()
			return err
		}
		return nil
	}

	os.Remove(socketPath)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}
	d.listener = listener
	os.Chmod(socketPath, 0600)
	return nil
}

// stripBrowserUrlFromConfig removes --browserUrl args from chrome-devtools config
// when Chrome isn't available on port 9222.
func (d *Daemon) stripBrowserUrlFromConfig() {
	for i, srv := range d.config.MCPServers {
		if !strings.Contains(strings.ToLower(srv.Name), "chrome") &&
			!strings.Contains(strings.Join(srv.Args, " "), "chrome-devtools-mcp") {
			continue
		}
		var filteredArgs []string
		skipNext := false
		for _, arg := range srv.Args {
			if skipNext {
				skipNext = false
				continue
			}
			if arg == "--browserUrl" {
				skipNext = true
				continue
			}
			filteredArgs = append(filteredArgs, arg)
		}
		if len(filteredArgs) != len(srv.Args) {
			log.Printf("[BROWSER] Removed --browserUrl from %s config (falling back to isolated Chrome)", srv.Name)
			d.config.MCPServers[i].Args = filteredArgs
		}
	}
}

func (d *Daemon) startServers() {
	browserPort := 9222
	if d.config.Browser != nil && d.config.Browser.Port > 0 {
		browserPort = d.config.Browser.Port
	}

	for _, srv := range d.config.MCPServers {
		if srv.Enabled {
			// For browser MCP servers: inject --wsEndpoint at runtime
			if coretools.IsBrowserMCPServer(srv.Name, srv.Args) {
				srv.Args = coretools.InjectBrowserEndpoint(srv.Args, browserPort)
			}
			if err := d.registry.StartServer(srv); err != nil {
				log.Printf("Failed to start server %s: %v", srv.Name, err)
			}
		}
	}
}

func (d *Daemon) setupBridgeHandlers() {
	d.bridge.SetToolCallHandler(func(tc bridge.ToolCall) {
		start := time.Now()

		// Route through UnifiedRegistry: core tools first, then MCP tools
		var result string
		var err error
		if d.unifiedRegistry != nil {
			result, err = d.unifiedRegistry.Execute(context.Background(), tc.ToolName, tc.Arguments)
		} else {
			result, err = d.registry.ExecuteTool(tc.ToolName, tc.Arguments)
		}
		latency := time.Since(start)

		serverName := ""
		if d.unifiedRegistry != nil && d.unifiedRegistry.IsCoreTool(tc.ToolName) {
			serverName = "core"
		} else {
			serverName = d.registry.GetServerForTool(tc.ToolName)
		}

		activity := Activity{
			Timestamp:  time.Now(),
			ToolName:   tc.ToolName,
			Arguments:  truncateArgs(tc.Arguments),
			Latency:    latency.Milliseconds(),
			ServerName: serverName,
		}

		if err != nil {
			activity.Success = false
			activity.Error = err.Error()
			d.bridge.SendToolResult(tc.CallID, false, "", err.Error())
		} else {
			activity.Success = true
			d.bridge.SendToolResult(tc.CallID, true, result, "")
		}

		d.addActivity(activity)
		d.broadcastActivity(activity)
	})

	d.bridge.SetTokenRefreshHandler(func(accessToken, refreshToken string, expiry int64) {
		d.config.AuthToken = accessToken
		d.config.RefreshToken = refreshToken
		d.config.TokenExpiry = expiry
		if d.config.Device != nil {
			d.config.Device.RefreshToken = refreshToken
		}
		config.Save(d.config)
		d.broadcastStatus()
	})

	d.bridge.SetDisconnectHandler(func() {
		d.status = "reconnecting"
		d.broadcastStatus()
	})

	d.bridge.SetReconnectHandler(func() {
		d.registerTools()
		d.status = "connected"
		d.broadcastStatus()
	})

	// Handle persona sync pushed from the cloud Nexus UI (no-op without local brain)
	d.bridge.SetPersonaSyncHandler(func(payload map[string]interface{}) {
		log.Printf("[SYNC] Received persona sync from cloud (no local brain to apply)")
	})

	// Handle backend requesting a full state re-sync
	d.bridge.SetRequestSyncHandler(func() {
		// No local state to sync without brain — just log
		log.Printf("[SYNC] Backend requested state sync (no local state)")
	})

	// Handle server management commands from the web UI
	d.bridge.SetServerCommandHandler(func(cmd bridge.ServerCommand) error {
		switch cmd.Action {
		case "add_server":
			name, _ := cmd.Payload["name"].(string)
			description, _ := cmd.Payload["description"].(string)
			command, _ := cmd.Payload["command"].(string)

			var args []string
			if rawArgs, ok := cmd.Payload["args"].([]interface{}); ok {
				for _, a := range rawArgs {
					if s, ok := a.(string); ok {
						args = append(args, s)
					}
				}
			}

			if name == "" || command == "" {
				return fmt.Errorf("name and command are required")
			}

			err := d.addServer(AddServerPayload{
				Name: name, Description: description,
				Command: command, Args: args,
			})
			if err != nil {
				return err
			}
			d.broadcastStatus()
			return nil

		case "remove_server":
			name, _ := cmd.Payload["name"].(string)
			if name == "" {
				return fmt.Errorf("server name is required")
			}
			if err := d.removeServer(name); err != nil {
				return err
			}
			d.broadcastStatus()
			return nil

		case "toggle_server":
			name, _ := cmd.Payload["name"].(string)
			enabled, _ := cmd.Payload["enabled"].(bool)
			if name == "" {
				return fmt.Errorf("server name is required")
			}
			if err := d.toggleServer(name, enabled); err != nil {
				return err
			}
			d.broadcastStatus()
			return nil

		default:
			return fmt.Errorf("unknown server command: %s", cmd.Action)
		}
	})
}

func (d *Daemon) connectBackend() {
	d.status = "connecting"
	d.broadcastStatus()

	if err := d.bridge.Connect(); err != nil {
		log.Printf("Initial connection failed: %v, retrying with backoff...", err)
		d.status = "reconnecting"
		d.broadcastStatus()
		// Use ConnectWithRetry which handles exponential backoff and token refresh
		d.bridge.ConnectWithRetry()
		return
	}

	d.registerTools()
	d.status = "connected"
	d.broadcastStatus()
}

func (d *Daemon) registerTools() {
	var allTools []interface{}

	// Core tools from unified registry
	if d.unifiedRegistry != nil {
		for _, def := range d.unifiedRegistry.GetCoreToolDefsAsMap() {
			allTools = append(allTools, def)
		}
	}

	// MCP tools from server registry
	mcpTools := d.registry.GetAllTools()
	for _, t := range mcpTools {
		name, _ := t["name"].(string)
		if d.unifiedRegistry != nil && d.unifiedRegistry.IsCoreTool(name) {
			continue
		}
		allTools = append(allTools, t)
	}

	servers := d.buildServerPayload()
	if err := d.bridge.RegisterTools(d.clientID, "1.0.0", runtime.GOOS, allTools, servers); err != nil {
		log.Printf("Warning: failed to register tools: %v", err)
	}
}

func (d *Daemon) buildServerPayload() []interface{} {
	var servers []interface{}
	for _, srv := range d.config.MCPServers {
		servers = append(servers, map[string]interface{}{
			"name":        srv.Name,
			"description": srv.Description,
			"command":     srv.Command,
			"args":        srv.Args,
			"type":        srv.Type,
			"enabled":     srv.Enabled,
		})
	}
	return servers
}

func (d *Daemon) acceptConnections() {
	for {
		conn, err := d.listener.Accept()
		if err != nil {
			select {
			case <-d.ctx.Done():
				return
			default:
				log.Printf("Accept error: %v", err)
				continue
			}
		}

		d.clientMu.Lock()
		d.clients[conn] = true
		d.clientMu.Unlock()

		go d.handleClient(conn)
	}
}

func (d *Daemon) handleClient(conn net.Conn) {
	defer func() {
		d.clientMu.Lock()
		delete(d.clients, conn)
		d.clientMu.Unlock()
		conn.Close()
	}()

	// Send initial status
	d.sendStatus(conn)

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var msg IPCMessage
		if err := decoder.Decode(&msg); err != nil {
			return
		}

		switch msg.Type {
		case MsgTypePing:
			encoder.Encode(IPCMessage{Type: MsgTypePong})

		case MsgTypeStatus:
			d.sendStatus(conn)

		case MsgTypeAddServer:
			var payload AddServerPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				d.sendError(conn, "invalid payload")
				continue
			}
			if err := d.addServer(payload); err != nil {
				d.sendError(conn, err.Error())
			} else {
				d.sendOK(conn)
				d.broadcastStatus()
			}

		case MsgTypeRemoveServer:
			var payload ServerNamePayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				d.sendError(conn, "invalid payload")
				continue
			}
			if err := d.removeServer(payload.Name); err != nil {
				d.sendError(conn, err.Error())
			} else {
				d.sendOK(conn)
				d.broadcastStatus()
			}

		case MsgTypeToggleServer:
			var payload ServerNamePayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				d.sendError(conn, "invalid payload")
				continue
			}
			if err := d.toggleServer(payload.Name, payload.Enabled); err != nil {
				d.sendError(conn, err.Error())
			} else {
				d.sendOK(conn)
				d.broadcastStatus()
			}

		case MsgTypeShutdown:
			d.sendOK(conn)
			d.Shutdown()
			return
		}
	}
}

func (d *Daemon) sendStatus(conn net.Conn) {
	status := d.getStatusPayload()
	payload, _ := json.Marshal(status)
	json.NewEncoder(conn).Encode(IPCMessage{
		Type:    MsgTypeStatus,
		Payload: payload,
	})
}

func (d *Daemon) sendError(conn net.Conn, errMsg string) {
	payload, _ := json.Marshal(map[string]string{"error": errMsg})
	json.NewEncoder(conn).Encode(IPCMessage{
		Type:    MsgTypeError,
		Payload: payload,
	})
}

func (d *Daemon) sendOK(conn net.Conn) {
	json.NewEncoder(conn).Encode(IPCMessage{Type: MsgTypeOK})
}

func (d *Daemon) getStatusPayload() StatusPayload {
	servers := d.getServerList()
	errStr := ""
	if d.lastError != nil {
		errStr = d.lastError.Error()
	}

	toolCount := d.registry.GetToolCount()
	if d.unifiedRegistry != nil {
		toolCount += d.unifiedRegistry.CoreToolCount()
	}

	status := StatusPayload{
		Status:     d.status,
		BackendURL: d.backendURL,
		Error:      errStr,
		Servers:    servers,
		ToolCount:  toolCount,
	}

	if d.config.Device != nil {
		status.UserEmail = d.config.Device.UserEmail
		status.UserID = d.config.Device.UserID
		status.DeviceID = d.config.Device.DeviceID
		status.TokenExp = d.config.TokenExpiry
	}

	return status
}

func (d *Daemon) getServerList() []ServerInfo {
	var servers []ServerInfo
	for _, srv := range d.config.MCPServers {
		info := ServerInfo{
			Name:        srv.Name,
			Description: srv.Description,
			Command:     srv.Command,
			Args:        srv.Args,
			Enabled:     srv.Enabled,
		}
		if srv.Enabled {
			if _, err := d.registry.GetServer(srv.Name); err == nil {
				info.Connected = true
				info.ToolCount = d.registry.GetServerToolCount(srv.Name)
			}
		}
		servers = append(servers, info)
	}
	return servers
}

func (d *Daemon) broadcastStatus() {
	status := d.getStatusPayload()
	payload, _ := json.Marshal(status)
	msg := IPCMessage{Type: MsgTypeStatus, Payload: payload}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	data = append(data, '\n')

	d.clientMu.RLock()
	defer d.clientMu.RUnlock()

	for conn := range d.clients {
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		conn.Write(data)
	}
}

func (d *Daemon) broadcastActivity(activity Activity) {
	payload, _ := json.Marshal(activity)
	msg := IPCMessage{Type: MsgTypeActivity, Payload: payload}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	data = append(data, '\n')

	d.clientMu.RLock()
	defer d.clientMu.RUnlock()

	for conn := range d.clients {
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		conn.Write(data)
	}
}

func (d *Daemon) addActivity(activity Activity) {
	d.activityMu.Lock()
	defer d.activityMu.Unlock()

	d.activities = append(d.activities, activity)
	if len(d.activities) > d.maxActivity {
		d.activities = d.activities[1:]
	}
}

func (d *Daemon) addServer(payload AddServerPayload) error {
	srv := config.MCPServer{
		Name:        payload.Name,
		Description: payload.Description,
		Command:     payload.Command,
		Args:        payload.Args,
		Type:        "stdio",
		Enabled:     true,
	}

	d.config.MCPServers = append(d.config.MCPServers, srv)
	if err := config.Save(d.config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if err := d.registry.StartServer(srv); err != nil {
		return fmt.Errorf("saved but failed to start: %w", err)
	}

	d.registerTools()
	return nil
}

func (d *Daemon) removeServer(name string) error {
	d.registry.StopServer(name)

	var newServers []config.MCPServer
	for _, srv := range d.config.MCPServers {
		if srv.Name != name {
			newServers = append(newServers, srv)
		}
	}
	d.config.MCPServers = newServers

	if err := config.Save(d.config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	d.registerTools()
	return nil
}

func (d *Daemon) toggleServer(name string, enabled bool) error {
	var srv *config.MCPServer
	for i := range d.config.MCPServers {
		if d.config.MCPServers[i].Name == name {
			d.config.MCPServers[i].Enabled = enabled
			srv = &d.config.MCPServers[i]
			break
		}
	}

	if srv == nil {
		return fmt.Errorf("server not found: %s", name)
	}

	if err := config.Save(d.config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if enabled {
		if err := d.registry.StartServer(*srv); err != nil {
			return fmt.Errorf("failed to start server: %w", err)
		}
	} else {
		d.registry.StopServer(name)
	}

	d.registerTools()
	return nil
}

// hasBrowserServer checks if any configured MCP server is a browser server.
func (d *Daemon) hasBrowserServer() bool {
	knownNames := []string{"chrome", "playwright", "puppeteer", "browser", "browsermcp", "devtools"}
	for _, srv := range d.config.MCPServers {
		lower := strings.ToLower(srv.Name + " " + srv.Command + " " + strings.Join(srv.Args, " "))
		for _, known := range knownNames {
			if strings.Contains(lower, known) {
				return true
			}
		}
	}
	// Also check running servers in registry
	for _, name := range d.registry.GetServerNames() {
		lower := strings.ToLower(name)
		for _, known := range knownNames {
			if strings.Contains(lower, known) {
				return true
			}
		}
	}
	return false
}

// autoStartBrowserMCP starts chrome-devtools-mcp as a managed MCP server
// so the browser core tool can delegate to it.
// Uses --wsEndpoint (direct WebSocket) instead of --browserUrl to avoid
// Chrome profile lock conflicts with other chrome-devtools-mcp instances
// (e.g. from Claude Code / VSCode extensions).
func (d *Daemon) autoStartBrowserMCP() {
	// Check if npx is available
	npxPath, err := exec.LookPath("npx")
	if err != nil {
		log.Printf("[BROWSER] npx not found, cannot auto-start chrome-devtools-mcp")
		return
	}

	port := 9222
	if d.config.Browser != nil && d.config.Browser.Port > 0 {
		port = d.config.Browser.Port
	}

	// Build args with runtime endpoint injection (--wsEndpoint preferred, --browserUrl fallback)
	baseArgs := []string{"chrome-devtools-mcp@latest"}
	args := coretools.InjectBrowserEndpoint(baseArgs, port)

	srv := config.MCPServer{
		Name:        "chrome-devtools-mcp",
		Description: "Chrome DevTools MCP (auto-configured)",
		Command:     npxPath,
		Args:        args,
		Type:        "stdio",
		Enabled:     true,
	}

	log.Printf("[BROWSER] Auto-starting chrome-devtools-mcp (Chrome ready on port %d)", port)
	if err := d.registry.StartServer(srv); err != nil {
		log.Printf("[BROWSER] Failed to auto-start chrome-devtools-mcp: %v", err)
		return
	}

	log.Printf("[BROWSER] chrome-devtools-mcp started — browser tools now available")

	// Re-register tools with backend so it knows about the new browser tools
	go func() {
		time.Sleep(time.Second) // Give bridge time to connect
		d.registerTools()
	}()
}

func (d *Daemon) handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	d.Shutdown()
}

// Shutdown gracefully shuts down the daemon
func (d *Daemon) Shutdown() {
	log.Println("Shutting down daemon...")

	d.cancel()

	d.clientMu.Lock()
	for conn := range d.clients {
		conn.Close()
	}
	d.clientMu.Unlock()

	if d.listener != nil {
		d.listener.Close()
	}

	// Stop tracked background processes (dev servers, watchers, etc.)
	if d.unifiedRegistry != nil {
		d.unifiedRegistry.CleanupBackground()
	}

	d.registry.StopAll()
	d.bridge.Close()

	os.Remove(GetSocketPath())
	os.Remove(GetPIDPath())

	os.Exit(0)
}

func truncateArgs(args map[string]interface{}) string {
	if len(args) == 0 {
		return "-"
	}
	for k, v := range args {
		s := fmt.Sprintf("%s=%v", k, v)
		if len(s) > 30 {
			return s[:27] + "..."
		}
		return s
	}
	return "-"
}

func convertTools(tools []map[string]interface{}) []interface{} {
	result := make([]interface{}, len(tools))
	for i, tool := range tools {
		result[i] = tool
	}
	return result
}
