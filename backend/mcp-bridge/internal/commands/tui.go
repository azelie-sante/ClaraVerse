package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/claraverse/mcp-client/internal/auth"
	"github.com/claraverse/mcp-client/internal/bridge"
	"github.com/claraverse/mcp-client/internal/config"
	"github.com/claraverse/mcp-client/internal/coretools"
	"github.com/claraverse/mcp-client/internal/registry"
	"github.com/claraverse/mcp-client/internal/tui"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// TUICmd launches the interactive TUI application
var TUICmd = &cobra.Command{
	Use:    "tui",
	Short:  "Launch interactive TUI dashboard",
	Hidden: true, // Hidden because running `clara_companion` without args launches TUI by default
	Long: `Launch the interactive terminal user interface for managing MCP servers.

The TUI provides:
  - Real-time connection status
  - Server management (add/remove/enable/disable)
  - Activity log with tool execution history
  - Settings configuration

Navigation:
  - Use 1-6 or Tab to switch tabs
  - Use arrow keys or j/k to navigate lists
  - Press 'q' to quit
  - Mouse support enabled for clicking tabs and buttons`,
	RunE: runTUI,
}

func init() {
	TUICmd.Flags().Bool("setup", false, "Re-run the first-time setup wizard")
}

// TUIBridge wraps the bridge and registry for the TUI
type TUIBridge struct {
	bridge          *bridge.Bridge
	registry        *registry.Registry
	unifiedRegistry *coretools.UnifiedRegistry
	state           *tui.AppState
	program         *tea.Program
	clientID        string
	cancel          context.CancelFunc
}

func runTUI(cmd *cobra.Command, args []string) error {
	// Redirect log output to file to avoid corrupting TUI display
	originalOutput := log.Writer()
	logDir := filepath.Join(config.GetConfigDir(), "logs")
	os.MkdirAll(logDir, 0755)
	logFile, logErr := os.OpenFile(filepath.Join(logDir, "tui.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if logErr == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	} else {
		log.SetOutput(io.Discard)
	}
	defer log.SetOutput(originalOutput)

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if authenticated
	if cfg.AuthToken == "" {
		fmt.Println("Not authenticated. Please run 'clara_companion login' first.")
		return nil
	}

	// Check and refresh token if needed
	if auth.IsTokenExpiringSoon(cfg.TokenExpiry, auth.TokenExpiryBuffer) {
		var refreshErr error

		if cfg.Device != nil && cfg.Device.DeviceID != "" && cfg.Device.RefreshToken != "" {
			refreshErr = RefreshDeviceToken(cfg)
		} else if cfg.RefreshToken != "" {
			var newAccessToken, newRefreshToken string
			var newExpiry int64
			newAccessToken, newRefreshToken, newExpiry, refreshErr = auth.ValidateAndRefreshIfNeeded(
				cfg.AuthToken, cfg.RefreshToken, cfg.TokenExpiry,
			)
			if refreshErr == nil {
				cfg.AuthToken = newAccessToken
				cfg.RefreshToken = newRefreshToken
				cfg.TokenExpiry = newExpiry
				if err := config.Save(cfg); err != nil {
					log.Printf("Warning: failed to save refreshed token: %v", err)
				}
			}
		} else {
			refreshErr = auth.ErrTokenExpired
		}

		if refreshErr != nil {
			if errors.Is(refreshErr, auth.ErrTokenExpired) {
				return fmt.Errorf("token expired and no refresh token available. Please run 'clara_companion login'")
			}
			return fmt.Errorf("token refresh failed: %w. Please run 'clara_companion login'", refreshErr)
		}
	}

	// Initialize application state
	state := tui.NewAppState()

	// Set user info from config
	tokenExp := time.Unix(cfg.TokenExpiry, 0)
	state.SetUser(cfg.Device.UserEmail, cfg.Device.UserID, cfg.Device.DeviceID, tokenExp)

	// Set backend URL and initial status
	state.SetStatus(tui.StatusConnecting, cfg.BackendURL, nil)

	// Create server registry
	verbose := false
	if cmd != nil {
		verbose, _ = cmd.Flags().GetBool("verbose")
	}
	reg := registry.NewRegistry(verbose)

	// Check if first-run setup is needed (or --setup flag)
	forceSetup := forceSetupFlag
	if cmd != nil {
		if fs, _ := cmd.Flags().GetBool("setup"); fs {
			forceSetup = true
		}
	}
	forceSetupFlag = false // Reset for next invocation
	needsWizard := !cfg.FirstRunCompleted || forceSetup

	// Only do browser/server startup if wizard is not needed.
	// If wizard is needed, these will be done after the wizard completes.
	if !needsWizard {
		// Clean up stale browser state and auto-launch Chrome with config
		coretools.CleanupBrowserOnStartup()
		if cfg.Browser != nil && cfg.Browser.AutoLaunch {
			if !coretools.EnsureBrowserReadyWithConfig(cfg.Browser.ProfilePath, cfg.Browser.Port) {
				stripBrowserUrlFromConfig(cfg)
			}
		} else if cfg.Browser == nil {
			// Legacy path: no browser config yet
			if !coretools.EnsureBrowserReady() {
				stripBrowserUrlFromConfig(cfg)
			}
		}
	}

	// Create UnifiedRegistry (core tools + MCP tools)
	unifiedReg := coretools.NewUnifiedRegistry(reg)

	// Start all enabled MCP servers and update state (skip if wizard needed)
	if !needsWizard {
		startMCPServersWithBrowser(cfg, reg, state)
	}

	// Create WebSocket bridge
	b := bridge.NewBridge(cfg.BackendURL, cfg.AuthToken, cfg.RefreshToken, cfg.TokenExpiry, verbose)

	// Set device info for device token refresh if available
	if cfg.Device != nil && cfg.Device.DeviceID != "" && cfg.Device.RefreshToken != "" {
		b.SetDeviceInfo(cfg.Device.DeviceID, cfg.Device.RefreshToken)
	}

	// Create TUI bridge wrapper
	ctx, cancel := context.WithCancel(context.Background())
	tb := &TUIBridge{
		bridge:          b,
		registry:        reg,
		unifiedRegistry: unifiedReg,
		state:           state,
		clientID:        uuid.New().String(),
		cancel:          cancel,
	}

	// Set up bridge handlers (tool calls, token refresh, reconnect)
	tb.setupHandlers(cfg)

	// Create the TUI app
	app := tui.NewApp(state)
	app.SetVersion(AppVersion)

	// If wizard is needed, attach it to the app
	if needsWizard {
		wizard := tui.NewWizardModel()
		app.SetWizard(wizard)
		app.SetViewMode(tui.ViewSetupWizard)

		// Callback: when wizard completes, save config + add chrome-devtools-mcp + start servers
		app.SetWizardDoneCallback(func(result tui.WizardResult) {
			cfg.Browser = &result.BrowserConfig
			cfg.FirstRunCompleted = true

			// Auto-add chrome-devtools-mcp if not already present.
			// Save with --browserUrl only (stable across Chrome restarts).
			// The runtime --wsEndpoint injection happens in startMCPServersWithBrowser().
			hasChromeDevtools := false
			for _, srv := range cfg.MCPServers {
				if srv.Name == "chrome-devtools" {
					hasChromeDevtools = true
					break
				}
			}
			if !hasChromeDevtools {
				port := result.BrowserConfig.Port
				cfg.MCPServers = append(cfg.MCPServers, config.MCPServer{
					Name:        "chrome-devtools",
					Description: "Chrome DevTools integration",
					Command:     "npx",
					Args:        []string{"-y", "chrome-devtools-mcp@latest", "--browserUrl", fmt.Sprintf("http://127.0.0.1:%d", port)},
					Type:        "stdio",
					Enabled:     true,
				})
			}

			config.Save(cfg)

			// Launch Chrome and wait until it's ready before starting MCP servers
			coretools.CleanupBrowserOnStartup()
			coretools.EnsureBrowserReadyWithConfig(cfg.Browser.ProfilePath, cfg.Browser.Port)
			coretools.WaitForChromeReady(cfg.Browser.Port, 10*time.Second)
			startMCPServersWithBrowser(cfg, reg, state)
		})
	}

	// Set up server operation callbacks
	app.SetCallbacks(&tui.ServerCallbacks{
		OnAddServer: func(ctx context.Context, name, desc, cmd string, args []string) error {
			return tb.addServer(ctx, cfg, name, desc, cmd, args)
		},
		OnRemoveServer: func(name string) error {
			return tb.removeServer(cfg, name)
		},
		OnToggleServer: func(name string, enabled bool) error {
			return tb.toggleServer(cfg, name, enabled)
		},
	})

	// Create Bubble Tea program
	program := tea.NewProgram(app, tea.WithAltScreen())
	tb.program = program

	// Connect to backend in background
	go tb.connectAndRun(ctx)

	// Run the TUI
	if _, err := program.Run(); err != nil {
		cancel()
		b.Close()
		reg.StopAll()
		return fmt.Errorf("TUI error: %w", err)
	}

	// Check how user chose to quit
	quitMode := app.GetQuitMode()

	// Cleanup on exit
	cancel()
	b.Close()

	if quitMode == tui.QuitModeBackground {
		// User wants to run in background - spawn daemon
		reg.StopAll()
		if err := StartDaemonBackground(); err != nil {
			fmt.Printf("Warning: failed to start background service: %v\n", err)
		}
	} else {
		// Full quit - stop all servers
		reg.StopAll()
	}

	return nil
}

func (tb *TUIBridge) setupHandlers(cfg *config.Config) {
	// Set tool call handler — routes through UnifiedRegistry (core + MCP)
	tb.bridge.SetToolCallHandler(func(tc bridge.ToolCall) {
		start := time.Now()

		// Execute via UnifiedRegistry (checks core tools first, then MCP)
		result, err := tb.unifiedRegistry.Execute(context.Background(), tc.ToolName, tc.Arguments)
		latency := time.Since(start)

		// Create activity entry
		entry := tui.ActivityEntry{
			Timestamp:  time.Now(),
			ToolName:   tc.ToolName,
			Arguments:  truncateArgs(tc.Arguments),
			Latency:    latency,
			ServerName: tb.registry.GetServerForTool(tc.ToolName),
		}

		if err != nil {
			entry.Success = false
			entry.Error = err.Error()
			tb.bridge.SendToolResult(tc.CallID, false, "", err.Error())
		} else {
			entry.Success = true
			tb.bridge.SendToolResult(tc.CallID, true, result, "")
		}

		// Update activity log
		tb.state.AddActivity(entry)

		// Send update to TUI
		if tb.program != nil {
			tb.program.Send(tui.ActivityMsg{Entry: entry})
		}
	})

	// Set token refresh handler
	tb.bridge.SetTokenRefreshHandler(func(accessToken, refreshToken string, expiry int64) {
		cfg.AuthToken = accessToken
		cfg.RefreshToken = refreshToken
		cfg.TokenExpiry = expiry
		if cfg.Device != nil {
			cfg.Device.RefreshToken = refreshToken
		}
		if err := config.Save(cfg); err != nil {
			log.Printf("Warning: failed to save refreshed token: %v", err)
		}

		// Update state
		tb.state.SetUser(cfg.Device.UserEmail, cfg.Device.UserID, cfg.Device.DeviceID, time.Unix(expiry, 0))

		if tb.program != nil {
			tb.program.Send(tui.TokenUpdateMsg{
				ExpiresAt: time.Unix(expiry, 0),
				UserEmail: cfg.Device.UserEmail,
				UserID:    cfg.Device.UserID,
			})
		}
	})

	// Set disconnect handler — update TUI status immediately on connection loss
	tb.bridge.SetDisconnectHandler(func() {
		tb.state.SetStatus(tui.StatusReconnecting, cfg.BackendURL, nil)
		if tb.program != nil {
			tb.program.Send(tui.StatusUpdateMsg{
				Status:     tui.StatusReconnecting,
				BackendURL: cfg.BackendURL,
			})
		}
	})

	// Set reconnect handler
	tb.bridge.SetReconnectHandler(func() {
		if err := tb.registerTools(); err != nil {
			log.Printf("❌ Tool registration failed after reconnect: %v", err)
			tb.state.SetStatus(tui.StatusError, cfg.BackendURL, err)
			if tb.program != nil {
				tb.program.Send(tui.StatusUpdateMsg{
					Status:     tui.StatusError,
					BackendURL: cfg.BackendURL,
					Error:      err,
				})
			}
			return
		}

		tb.state.SetStatus(tui.StatusConnected, cfg.BackendURL, nil)
		if tb.program != nil {
			tb.program.Send(tui.StatusUpdateMsg{
				Status:     tui.StatusConnected,
				BackendURL: cfg.BackendURL,
			})
		}
	})
}

func (tb *TUIBridge) connectAndRun(ctx context.Context) {
	backendURL := tb.state.BackendURL

	// Update status to connecting
	tb.state.SetStatus(tui.StatusConnecting, backendURL, nil)
	if tb.program != nil {
		tb.program.Send(tui.StatusUpdateMsg{
			Status:     tui.StatusConnecting,
			BackendURL: backendURL,
		})
	}

	// Try initial connection
	if err := tb.bridge.Connect(); err != nil {
		log.Printf("Initial connection failed: %v, retrying with backoff...", err)
		tb.state.SetStatus(tui.StatusReconnecting, backendURL, err)
		if tb.program != nil {
			tb.program.Send(tui.StatusUpdateMsg{
				Status:     tui.StatusReconnecting,
				BackendURL: backendURL,
				Error:      err,
			})
		}
		// Use ConnectWithRetry which handles exponential backoff and token refresh.
		// The reconnect handler will set status to connected and register tools.
		tb.bridge.ConnectWithRetry()
		// Wait for context cancellation
		<-ctx.Done()
		return
	}

	// Register tools (core + MCP) — wait for ACK before showing Connected
	if err := tb.registerTools(); err != nil {
		log.Printf("❌ Tool registration failed on initial connect: %v", err)
		tb.state.SetStatus(tui.StatusError, backendURL, err)
		if tb.program != nil {
			tb.program.Send(tui.StatusUpdateMsg{
				Status:     tui.StatusError,
				BackendURL: backendURL,
				Error:      err,
			})
		}
	} else {
		tb.state.SetStatus(tui.StatusConnected, backendURL, nil)
		if tb.program != nil {
			tb.program.Send(tui.StatusUpdateMsg{
				Status:     tui.StatusConnected,
				BackendURL: backendURL,
			})
		}
	}

	// Wait for context cancellation
	<-ctx.Done()
}

func (tb *TUIBridge) registerTools() error {
	// Collect core tool defs as generic maps
	coreDefs := tb.unifiedRegistry.GetCoreToolDefsAsMap()

	// Collect MCP tool defs (excluding those shadowed by core tools)
	mcpTools := tb.registry.GetAllTools()
	var allTools []interface{}

	for _, def := range coreDefs {
		allTools = append(allTools, def)
	}
	for _, t := range mcpTools {
		name, _ := t["name"].(string)
		if name == "" {
			continue
		}
		// Skip MCP tools that shadow core tools
		if tb.unifiedRegistry.IsCoreTool(name) {
			continue
		}
		allTools = append(allTools, t)
	}

	if err := tb.bridge.RegisterTools(tb.clientID, "1.0.0", runtime.GOOS, allTools); err != nil {
		log.Printf("Warning: failed to register tools: %v", err)
		return err
	}
	return nil
}

func truncateArgs(args map[string]interface{}) string {
	if len(args) == 0 {
		return "-"
	}

	// Get first key-value pair as summary
	for k, v := range args {
		s := fmt.Sprintf("%s=%v", k, v)
		if len(s) > 30 {
			return s[:27] + "..."
		}
		return s
	}
	return "-"
}

// addServer adds a new MCP server, persists to config, and starts it
func (tb *TUIBridge) addServer(ctx context.Context, cfg *config.Config, name, desc, cmd string, args []string) error {
	// Check if already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create the server config
	srv := config.MCPServer{
		Name:        name,
		Description: desc,
		Command:     cmd,
		Args:        args,
		Type:        "stdio",
		Enabled:     true,
	}

	// Add to config
	cfg.MCPServers = append(cfg.MCPServers, srv)

	// Save config
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Check context before starting
	select {
	case <-ctx.Done():
		// Remove from config since we didn't start it
		tb.removeServerFromConfig(cfg, name)
		return ctx.Err()
	default:
	}

	// Start the server with context for cancellation
	if err := tb.registry.StartServerWithContext(ctx, srv); err != nil {
		// Remove from config if cancelled or failed
		if ctx.Err() != nil {
			tb.removeServerFromConfig(cfg, name)
			return ctx.Err()
		}
		// Still saved to config, just failed to start
		tb.state.SetServers(tb.buildServerList(cfg))
		return fmt.Errorf("saved but failed to start: %w", err)
	}

	// Update state with connected server
	serverInfo := tui.ServerInfo{
		Name:        name,
		Description: desc,
		Command:     cmd,
		Args:        args,
		Enabled:     true,
		Connected:   true,
		ToolCount:   tb.registry.GetServerToolCount(name),
	}

	servers := tb.state.GetServers()
	servers = append(servers, serverInfo)
	tb.state.SetServers(servers)

	// Re-register all tools with the backend (best-effort)
	_ = tb.registerTools()

	return nil
}

// removeServer removes an MCP server, persists to config, and stops it
func (tb *TUIBridge) removeServer(cfg *config.Config, name string) error {
	// Stop the server first
	tb.registry.StopServer(name)

	// Remove from config
	var newServers []config.MCPServer
	for _, srv := range cfg.MCPServers {
		if srv.Name != name {
			newServers = append(newServers, srv)
		}
	}
	cfg.MCPServers = newServers

	// Save config
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Re-register tools with the backend (some tools removed, best-effort)
	_ = tb.registerTools()

	return nil
}

// toggleServer enables/disables an MCP server, persists to config, and starts/stops it
func (tb *TUIBridge) toggleServer(cfg *config.Config, name string, enabled bool) error {
	// Find and update the server in config
	var srv *config.MCPServer
	for i := range cfg.MCPServers {
		if cfg.MCPServers[i].Name == name {
			cfg.MCPServers[i].Enabled = enabled
			srv = &cfg.MCPServers[i]
			break
		}
	}

	if srv == nil {
		return fmt.Errorf("server not found: %s", name)
	}

	// Save config
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Start or stop the server
	if enabled {
		if err := tb.registry.StartServer(*srv); err != nil {
			return fmt.Errorf("failed to start server: %w", err)
		}
		// Update state
		servers := tb.state.GetServers()
		for i := range servers {
			if servers[i].Name == name {
				servers[i].Enabled = true
				servers[i].Connected = true
				servers[i].ToolCount = tb.registry.GetServerToolCount(name)
				break
			}
		}
		tb.state.SetServers(servers)
	} else {
		tb.registry.StopServer(name)
	}

	// Re-register tools with the backend (best-effort)
	_ = tb.registerTools()

	return nil
}

// removeServerFromConfig removes a server from config and saves
func (tb *TUIBridge) removeServerFromConfig(cfg *config.Config, name string) {
	var newServers []config.MCPServer
	for _, srv := range cfg.MCPServers {
		if srv.Name != name {
			newServers = append(newServers, srv)
		}
	}
	cfg.MCPServers = newServers
	config.Save(cfg)
}

// buildServerList creates a ServerInfo list from config
func (tb *TUIBridge) buildServerList(cfg *config.Config) []tui.ServerInfo {
	servers := make([]tui.ServerInfo, 0, len(cfg.MCPServers))
	for _, srv := range cfg.MCPServers {
		serverInfo := tui.ServerInfo{
			Name:        srv.Name,
			Description: srv.Description,
			Command:     srv.Command,
			Args:        srv.Args,
			Enabled:     srv.Enabled,
			Connected:   false,
		}
		if srv.Enabled {
			// Check if server is running by trying to get it
			if _, err := tb.registry.GetServer(srv.Name); err == nil {
				serverInfo.Connected = true
				serverInfo.ToolCount = tb.registry.GetServerToolCount(srv.Name)
			}
		}
		servers = append(servers, serverInfo)
	}
	return servers
}

// startMCPServersWithBrowser starts all enabled MCP servers and updates TUI state.
// For browser MCP servers, it dynamically injects --wsEndpoint at runtime so
// chrome-devtools-mcp connects to the Chrome we launched (not its own).
// The --wsEndpoint is NOT persisted to config (it changes on every Chrome restart).
func startMCPServersWithBrowser(cfg *config.Config, reg *registry.Registry, state *tui.AppState) {
	// Determine browser port for endpoint injection
	browserPort := 9222
	if cfg.Browser != nil && cfg.Browser.Port > 0 {
		browserPort = cfg.Browser.Port
	}

	servers := make([]tui.ServerInfo, 0, len(cfg.MCPServers))
	for _, srv := range cfg.MCPServers {
		serverInfo := tui.ServerInfo{
			Name:        srv.Name,
			Description: srv.Description,
			Command:     srv.Command,
			Args:        srv.Args,
			Enabled:     srv.Enabled,
			Connected:   false,
		}

		if srv.Enabled {
			// For browser MCP servers: inject --wsEndpoint at runtime (ephemeral, not saved)
			startSrv := srv
			if coretools.IsBrowserMCPServer(srv.Name, srv.Args) {
				startSrv.Args = coretools.InjectBrowserEndpoint(srv.Args, browserPort)
			}

			if err := reg.StartServer(startSrv); err != nil {
				serverInfo.Error = err.Error()
			} else {
				serverInfo.Connected = true
				serverInfo.ToolCount = reg.GetServerToolCount(srv.Name)
			}
		}

		servers = append(servers, serverInfo)
	}
	state.SetServers(servers)
}

// stripBrowserUrlFromConfig removes --browserUrl from chrome-devtools-mcp server args
// when Chrome isn't available on port 9222, so the MCP server falls back to isolated Chrome.
func stripBrowserUrlFromConfig(cfg *config.Config) {
	for i, srv := range cfg.MCPServers {
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
			cfg.MCPServers[i].Args = filteredArgs
		}
	}
}

// RunTUIDefault runs the TUI as the default command when no subcommand is specified
func RunTUIDefault() error {
	return RunTUIDefaultWithSetup(false)
}

// RunTUIDefaultWithSetup runs the TUI, optionally forcing the setup wizard.
func RunTUIDefaultWithSetup(forceSetup bool) error {
	// Check if we have a TTY
	fi, _ := os.Stdout.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		// Not a terminal, don't run TUI
		return fmt.Errorf("not a terminal, use specific commands instead")
	}

	forceSetupFlag = forceSetup
	return runTUI(nil, nil)
}

// forceSetupFlag is set by RunTUIDefaultWithSetup when called from the root command.
var forceSetupFlag bool

// AppVersion is set by main.go before command execution. The TUI reads it.
var AppVersion = "0.0.0-dev"
