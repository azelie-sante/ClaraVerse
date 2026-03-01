package main

import (
	"fmt"
	"os"

	"github.com/claraverse/mcp-client/internal/commands"
	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via -ldflags "-X main.Version=X.Y.Z"
	Version = "0.0.0-dev"
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "clara_companion",
	Short: "Clara Companion - Connect local tools to cloud",
	Long: `Clara Companion connects local MCP servers to your ClaraVerse cloud chat.

Quick Start:
  clara_companion                 Launch interactive dashboard (default)
  clara_companion login           Authenticate (first time)
  clara_companion start           Start headless (for scripts/automation)

Commands:
  login                      Authenticate with ClaraVerse
  logout                     Log out and revoke this device
  start                      Start headless mode (no UI)
  status                     Show connection and auth status
  list                       List configured MCP servers
  add <name>                 Add a new MCP server
  remove <name>              Remove an MCP server
  service install/uninstall  Manage background service
  devices list/revoke        Manage connected devices

Examples:
  clara_companion                                 # Launch dashboard
  clara_companion login                           # First-time setup
  clara_companion add browser --command npx --args @browsermcp/mcp@latest
  clara_companion service install                 # Auto-start on login

Config: ~/.claraverse/mcp-config.yaml
Logs:   ~/.claraverse/logs/clara_companion.log`,
	Version: Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		// When no subcommand is specified, try to launch TUI
		forceSetup, _ := cmd.Flags().GetBool("setup")
		return commands.RunTUIDefaultWithSetup(forceSetup)
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.Flags().Bool("setup", false, "Re-run the first-time setup wizard")

	// Add all commands
	rootCmd.AddCommand(commands.TUICmd)
	rootCmd.AddCommand(commands.LoginCmd)
	rootCmd.AddCommand(commands.StartCmd)
	rootCmd.AddCommand(commands.AddCmd)
	rootCmd.AddCommand(commands.ListCmd)
	rootCmd.AddCommand(commands.RemoveCmd)
	rootCmd.AddCommand(commands.StatusCmd)
	rootCmd.AddCommand(commands.ServiceCmd)
	rootCmd.AddCommand(commands.DevicesCmd)
	rootCmd.AddCommand(commands.LogoutCmd)
	rootCmd.AddCommand(commands.DaemonCmd)
}

func main() {
	commands.AppVersion = Version
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
