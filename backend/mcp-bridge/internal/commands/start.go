package commands

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/claraverse/mcp-client/internal/auth"
	"github.com/claraverse/mcp-client/internal/bridge"
	"github.com/claraverse/mcp-client/internal/config"
	"github.com/claraverse/mcp-client/internal/registry"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the MCP client and connect to backend",
	Long: `Starts the MCP client daemon, connects to the ClaraVerse backend,
and registers all enabled MCP servers. The client will run in the foreground
and handle tool execution requests from the backend.`,
	RunE: runStart,
}

func runStart(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if authenticated
	if cfg.AuthToken == "" {
		return fmt.Errorf("not authenticated. Please run 'clara_companion login' first")
	}

	verbose, _ := cmd.Flags().GetBool("verbose")

	log.Println("🚀 Starting Clara Companion")
	log.Printf("📍 Config: %s", config.GetConfigPath())
	log.Printf("🌐 Backend: %s", cfg.BackendURL)

	// Check if token is expired or expiring soon, and try to refresh
	if auth.IsTokenExpiringSoon(cfg.TokenExpiry, auth.TokenExpiryBuffer) {
		log.Println("⚠️  Token expired or expiring soon, attempting refresh...")

		var refreshErr error

		// Use device token refresh if we have device credentials (from clara_companion login)
		if cfg.Device != nil && cfg.Device.DeviceID != "" && cfg.Device.RefreshToken != "" {
			refreshErr = RefreshDeviceToken(cfg)
		} else if cfg.RefreshToken != "" {
			// Fall back to Supabase refresh for legacy tokens
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
					log.Printf("⚠️  Warning: failed to save refreshed token: %v", err)
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

		log.Println("✅ Token refreshed successfully")
	}

	// Create server registry
	reg := registry.NewRegistry(verbose)

	// Start all enabled MCP servers
	enabledServers := cfg.GetEnabledServers()
	if len(enabledServers) == 0 {
		log.Println("⚠️  No MCP servers configured. Add servers with 'clara_companion add'")
	}

	for _, server := range enabledServers {
		if err := reg.StartServer(server); err != nil {
			log.Printf("❌ Failed to start %s: %v", server.Name, err)
			continue
		}
	}

	if reg.GetServerCount() == 0 {
		return fmt.Errorf("no MCP servers started successfully")
	}

	log.Printf("✅ Started %d MCP servers with %d total tools", reg.GetServerCount(), reg.GetToolCount())

	// Create WebSocket bridge
	b := bridge.NewBridge(cfg.BackendURL, cfg.AuthToken, cfg.RefreshToken, cfg.TokenExpiry, verbose)

	// Set device info for device token refresh if available
	if cfg.Device != nil && cfg.Device.DeviceID != "" && cfg.Device.RefreshToken != "" {
		b.SetDeviceInfo(cfg.Device.DeviceID, cfg.Device.RefreshToken)
	}

	// Generate client ID (persists across reconnections)
	clientID := uuid.New().String()
	tools := reg.GetAllTools()

	// Build server configs for registration
	servers := convertServers(enabledServers)

	// Helper function to register tools
	registerTools := func() {
		log.Printf("📦 Registering %d tools from %d servers...", len(tools), len(servers))
		if err := b.RegisterTools(clientID, "1.0.0", runtime.GOOS, convertTools(tools), servers); err != nil {
			log.Printf("⚠️  Warning: failed to register tools: %v", err)
		}
	}

	// Set tool call handler
	b.SetToolCallHandler(func(tc bridge.ToolCall) {
		handleToolCall(reg, b, tc)
	})

	// Set token refresh handler to save new tokens
	b.SetTokenRefreshHandler(func(accessToken, refreshToken string, expiry int64) {
		cfg.AuthToken = accessToken
		cfg.RefreshToken = refreshToken
		cfg.TokenExpiry = expiry
		// Also update device refresh token if using device auth
		if cfg.Device != nil {
			cfg.Device.RefreshToken = refreshToken
		}
		if err := config.Save(cfg); err != nil {
			log.Printf("⚠️  Warning: failed to save refreshed token: %v", err)
		}
	})

	// Set reconnect handler to re-register tools after reconnection
	b.SetReconnectHandler(func() {
		registerTools()
	})

	// Connect to backend
	log.Println("🔌 Connecting to backend...")
	if err := b.Connect(); err != nil {
		return fmt.Errorf("failed to connect to backend: %w", err)
	}

	// Register tools for initial connection
	registerTools()

	log.Println("✅ MCP client running. Press Ctrl+C to exit.")
	log.Println("💡 Tools are now available in your web chat!")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	log.Println("\n🛑 Shutting down...")
	b.Close()
	reg.StopAll()
	log.Println("✅ Goodbye!")

	return nil
}

func handleToolCall(reg *registry.Registry, b *bridge.Bridge, tc bridge.ToolCall) {
	log.Printf("🔧 Executing tool: %s (call_id: %s)", tc.ToolName, tc.CallID)

	// Execute the tool
	result, err := reg.ExecuteTool(tc.ToolName, tc.Arguments)

	if err != nil {
		log.Printf("❌ Tool execution failed: %v", err)
		b.SendToolResult(tc.CallID, false, "", err.Error())
		return
	}

	log.Printf("✅ Tool executed successfully: %s", tc.ToolName)
	b.SendToolResult(tc.CallID, true, result, "")
}

func convertTools(tools []map[string]interface{}) []interface{} {
	result := make([]interface{}, len(tools))
	for i, tool := range tools {
		result[i] = tool
	}
	return result
}

func convertServers(servers []config.MCPServer) []interface{} {
	result := make([]interface{}, len(servers))
	for i, s := range servers {
		result[i] = map[string]interface{}{
			"name":        s.Name,
			"description": s.Description,
			"command":     s.Command,
			"args":        s.Args,
			"type":        s.Type,
			"enabled":     s.Enabled,
		}
	}
	return result
}
