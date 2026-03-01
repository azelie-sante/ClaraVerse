package commands

import (
	"fmt"

	"github.com/claraverse/mcp-client/internal/config"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured MCP servers",
	Long:  `Display all MCP servers in your configuration, including their status and paths.`,
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.MCPServers) == 0 {
		fmt.Println("📋 No MCP servers configured")
		fmt.Println()
		fmt.Println("Add a server with: clara_companion add <name> --path <server-path>")
		return nil
	}

	fmt.Println("📋 Configured MCP Servers:")
	fmt.Println()

	for i, server := range cfg.MCPServers {
		status := "🟢 Enabled"
		if !server.Enabled {
			status = "🔴 Disabled"
		}

		fmt.Printf("%d. %s %s\n", i+1, server.Name, status)
		fmt.Printf("   Type: %s\n", server.Type)
		if server.Path != "" {
			fmt.Printf("   Path: %s\n", server.Path)
		}
		if server.URL != "" {
			fmt.Printf("   URL: %s\n", server.URL)
		}
		if server.Description != "" {
			fmt.Printf("   Description: %s\n", server.Description)
		}
		fmt.Println()
	}

	enabledCount := len(cfg.GetEnabledServers())
	fmt.Printf("Total: %d servers (%d enabled, %d disabled)\n", len(cfg.MCPServers), enabledCount, len(cfg.MCPServers)-enabledCount)

	return nil
}
