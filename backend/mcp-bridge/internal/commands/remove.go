package commands

import (
	"fmt"

	"github.com/claraverse/mcp-client/internal/config"
	"github.com/spf13/cobra"
)

var RemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove an MCP server",
	Long:  `Remove an MCP server from your configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

func runRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Remove server
	if err := cfg.RemoveServer(name); err != nil {
		return fmt.Errorf("failed to remove server: %w", err)
	}

	// Save config
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Removed MCP server: %s\n", name)
	return nil
}
