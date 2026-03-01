package commands

import (
	"fmt"

	"github.com/claraverse/mcp-client/internal/config"
	"github.com/spf13/cobra"
)

var (
	serverPath string
	serverType string
	serverDesc string
)

var AddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new MCP server",
	Long: `Add a new MCP server to your configuration. The server will be
enabled by default and started when you run 'clara_companion start'.

Examples:
  clara_companion add filesystem --path /usr/local/bin/mcp-server-filesystem
  clara_companion add database --path ./mcp-server-sqlite --type stdio`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

func init() {
	AddCmd.Flags().StringVar(&serverPath, "path", "", "Path to MCP server executable (required)")
	AddCmd.Flags().StringVar(&serverType, "type", "stdio", "Server type: stdio or sse")
	AddCmd.Flags().StringVar(&serverDesc, "description", "", "Server description")
	AddCmd.MarkFlagRequired("path")
}

func runAdd(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create server config
	server := config.MCPServer{
		Name:        name,
		Path:        serverPath,
		Type:        serverType,
		Description: serverDesc,
		Enabled:     true,
	}

	// Add server
	if err := cfg.AddServer(server); err != nil {
		return fmt.Errorf("failed to add server: %w", err)
	}

	// Save config
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Added MCP server: %s\n", name)
	fmt.Printf("📁 Path: %s\n", serverPath)
	fmt.Printf("📝 Type: %s\n", serverType)
	if serverDesc != "" {
		fmt.Printf("💬 Description: %s\n", serverDesc)
	}
	fmt.Println()
	fmt.Println("Server will be started automatically when you run 'clara_companion start'")

	return nil
}
