package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/claraverse/mcp-client/internal/config"
	"github.com/claraverse/mcp-client/internal/daemon"
	"github.com/spf13/cobra"
)

// DaemonCmd runs the MCP bridge as a background daemon
var DaemonCmd = &cobra.Command{
	Use:    "daemon",
	Short:  "Run as background daemon",
	Long:   `Run the MCP bridge as a background daemon without the TUI.`,
	Hidden: true, // Internal command, not shown in help
	RunE:   runDaemon,
}

// DaemonStopCmd stops the running daemon
var DaemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the background daemon",
	RunE:  stopDaemon,
}

// DaemonStatusCmd shows daemon status
var DaemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show daemon status",
	RunE:  daemonStatus,
}

func init() {
	DaemonCmd.AddCommand(DaemonStopCmd)
	DaemonCmd.AddCommand(DaemonStatusCmd)
}

func runDaemon(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.AuthToken == "" {
		return fmt.Errorf("not authenticated. Please run 'clara_companion login' first")
	}

	if daemon.IsRunning() {
		return fmt.Errorf("daemon is already running")
	}

	verbose, _ := cmd.Flags().GetBool("verbose")
	d, err := daemon.NewDaemon(cfg, verbose)
	if err != nil {
		return fmt.Errorf("failed to create daemon: %w", err)
	}

	fmt.Println("Starting daemon...")
	return d.Start()
}

func stopDaemon(cmd *cobra.Command, args []string) error {
	if !daemon.IsRunning() {
		fmt.Println("Daemon is not running")
		return nil
	}

	client, err := daemon.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer client.Close()

	if err := client.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown daemon: %w", err)
	}

	fmt.Println("Daemon stopped")
	return nil
}

func daemonStatus(cmd *cobra.Command, args []string) error {
	if !daemon.IsRunning() {
		fmt.Println("Daemon is not running")
		return nil
	}

	client, err := daemon.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer client.Close()

	fmt.Println("Daemon is running")
	fmt.Printf("Socket: %s\n", daemon.GetSocketPath())
	return nil
}

// StartDaemonBackground starts the daemon as a background process
func StartDaemonBackground() error {
	if daemon.IsRunning() {
		return nil // Already running
	}

	// Get the current executable path
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Start daemon as a detached process
	cmd := exec.Command(exe, "daemon")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	// Detach from parent process group
	setSysProcAttr(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	// Release the process so it continues after parent exits
	if err := cmd.Process.Release(); err != nil {
		return fmt.Errorf("failed to release daemon process: %w", err)
	}

	return nil
}
