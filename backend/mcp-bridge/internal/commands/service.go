package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/spf13/cobra"
)

var ServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage MCP client as a background service",
	Long: `Manage the MCP client as a background service that runs automatically on login.

Subcommands:
  install   - Install the service
  uninstall - Remove the service
  status    - Check service status
  start     - Start the service
  stop      - Stop the service`,
}

var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install MCP client as a background service",
	RunE:  runServiceInstall,
}

var serviceUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove MCP client background service",
	RunE:  runServiceUninstall,
}

var serviceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check MCP client service status",
	RunE:  runServiceStatus,
}

var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the MCP client service",
	RunE:  runServiceStart,
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the MCP client service",
	RunE:  runServiceStop,
}

func init() {
	ServiceCmd.AddCommand(serviceInstallCmd)
	ServiceCmd.AddCommand(serviceUninstallCmd)
	ServiceCmd.AddCommand(serviceStatusCmd)
	ServiceCmd.AddCommand(serviceStartCmd)
	ServiceCmd.AddCommand(serviceStopCmd)
}

// macOS launchd plist template
const launchdPlistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.claraverse.clara-companion</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.ExecutablePath}}</string>
        <string>start</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>{{.LogPath}}/clara_companion.log</string>
    <key>StandardErrorPath</key>
    <string>{{.LogPath}}/clara_companion.error.log</string>
    <key>WorkingDirectory</key>
    <string>{{.WorkingDir}}</string>
</dict>
</plist>
`

// Linux systemd service template
const systemdServiceTemplate = `[Unit]
Description=Clara Companion
After=network.target

[Service]
Type=simple
ExecStart={{.ExecutablePath}} start
Restart=always
RestartSec=10
WorkingDirectory={{.WorkingDir}}
StandardOutput=append:{{.LogPath}}/clara_companion.log
StandardError=append:{{.LogPath}}/clara_companion.error.log

[Install]
WantedBy=default.target
`

type serviceConfig struct {
	ExecutablePath string
	LogPath        string
	WorkingDir     string
}

func getServiceConfig() (*serviceConfig, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve executable path: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	logPath := filepath.Join(home, ".claraverse", "logs")
	if err := os.MkdirAll(logPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	return &serviceConfig{
		ExecutablePath: execPath,
		LogPath:        logPath,
		WorkingDir:     home,
	}, nil
}

func getLaunchdPlistPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, "Library", "LaunchAgents", "com.claraverse.clara-companion.plist"), nil
}

func getSystemdServicePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	serviceDir := filepath.Join(home, ".config", "systemd", "user")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create systemd directory: %w", err)
	}

	return filepath.Join(serviceDir, "claraverse-mcp.service"), nil
}

func runServiceInstall(cmd *cobra.Command, args []string) error {
	cfg, err := getServiceConfig()
	if err != nil {
		return err
	}

	switch runtime.GOOS {
	case "darwin":
		return installLaunchdService(cfg)
	case "linux":
		return installSystemdService(cfg)
	case "windows":
		fmt.Println("ℹ️  Background service is not yet supported on Windows.")
		fmt.Println()
		fmt.Println("Alternatives:")
		fmt.Println("  1. Add clara_companion.exe to your Startup folder:")
		fmt.Printf("     %s\\AppData\\Roaming\\Microsoft\\Windows\\Start Menu\\Programs\\Startup\n", os.Getenv("USERPROFILE"))
		fmt.Println()
		fmt.Println("  2. Create a scheduled task with Task Scheduler:")
		fmt.Printf("     schtasks /create /tn \"ClaraVerse MCP\" /tr \"%s start\" /sc onlogon\n", cfg.ExecutablePath)
		fmt.Println()
		fmt.Println("  3. Run manually when needed:")
		fmt.Println("     clara_companion start")
		return nil
	default:
		return fmt.Errorf("service installation not supported on %s", runtime.GOOS)
	}
}

func installLaunchdService(cfg *serviceConfig) error {
	plistPath, err := getLaunchdPlistPath()
	if err != nil {
		return err
	}

	// Ensure LaunchAgents directory exists
	launchAgentsDir := filepath.Dir(plistPath)
	if err := os.MkdirAll(launchAgentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	// Create plist file
	tmpl, err := template.New("plist").Parse(launchdPlistTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(plistPath)
	if err != nil {
		return fmt.Errorf("failed to create plist file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, cfg); err != nil {
		return fmt.Errorf("failed to write plist: %w", err)
	}

	// Load the service
	loadCmd := exec.Command("launchctl", "load", plistPath)
	if output, err := loadCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to load service: %s", string(output))
	}

	fmt.Println("✅ Service installed successfully!")
	fmt.Printf("   Location: %s\n", plistPath)
	fmt.Println()
	fmt.Println("   Commands:")
	fmt.Println("   • clara_companion service status    - Check status")
	fmt.Println("   • clara_companion service stop      - Stop service")
	fmt.Println("   • clara_companion service start     - Start service")
	fmt.Println("   • clara_companion service uninstall - Remove service")
	fmt.Println()
	fmt.Printf("   Logs: %s/clara_companion.log\n", cfg.LogPath)

	return nil
}

func installSystemdService(cfg *serviceConfig) error {
	servicePath, err := getSystemdServicePath()
	if err != nil {
		return err
	}

	// Create service file
	tmpl, err := template.New("service").Parse(systemdServiceTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(servicePath)
	if err != nil {
		return fmt.Errorf("failed to create service file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, cfg); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	// Reload systemd user daemon
	reloadCmd := exec.Command("systemctl", "--user", "daemon-reload")
	if output, err := reloadCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to reload systemd: %s", string(output))
	}

	// Enable the service
	enableCmd := exec.Command("systemctl", "--user", "enable", "claraverse-mcp.service")
	if output, err := enableCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to enable service: %s", string(output))
	}

	// Start the service
	startCmd := exec.Command("systemctl", "--user", "start", "claraverse-mcp.service")
	if output, err := startCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start service: %s", string(output))
	}

	fmt.Println("✅ Service installed and started successfully!")
	fmt.Printf("   Location: %s\n", servicePath)
	fmt.Println()
	fmt.Println("   Commands:")
	fmt.Println("   • clara_companion service status    - Check status")
	fmt.Println("   • clara_companion service stop      - Stop service")
	fmt.Println("   • clara_companion service start     - Start service")
	fmt.Println("   • clara_companion service uninstall - Remove service")
	fmt.Println()
	fmt.Printf("   Logs: %s/clara_companion.log\n", cfg.LogPath)

	return nil
}

func runServiceUninstall(cmd *cobra.Command, args []string) error {
	switch runtime.GOOS {
	case "darwin":
		return uninstallLaunchdService()
	case "linux":
		return uninstallSystemdService()
	case "windows":
		fmt.Println("ℹ️  Background service is not installed on Windows.")
		fmt.Println("If you created a scheduled task, remove it with:")
		fmt.Println("  schtasks /delete /tn \"ClaraVerse MCP\" /f")
		return nil
	default:
		return fmt.Errorf("service uninstall not supported on %s", runtime.GOOS)
	}
}

func uninstallLaunchdService() error {
	plistPath, err := getLaunchdPlistPath()
	if err != nil {
		return err
	}

	// Check if service exists
	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		fmt.Println("ℹ️  Service is not installed.")
		return nil
	}

	// Unload the service
	unloadCmd := exec.Command("launchctl", "unload", plistPath)
	unloadCmd.CombinedOutput() // Ignore errors if not loaded

	// Remove plist file
	if err := os.Remove(plistPath); err != nil {
		return fmt.Errorf("failed to remove plist file: %w", err)
	}

	fmt.Println("✅ Service uninstalled successfully!")
	return nil
}

func uninstallSystemdService() error {
	servicePath, err := getSystemdServicePath()
	if err != nil {
		return err
	}

	// Check if service exists
	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		fmt.Println("ℹ️  Service is not installed.")
		return nil
	}

	// Stop the service
	stopCmd := exec.Command("systemctl", "--user", "stop", "claraverse-mcp.service")
	stopCmd.CombinedOutput() // Ignore errors if not running

	// Disable the service
	disableCmd := exec.Command("systemctl", "--user", "disable", "claraverse-mcp.service")
	disableCmd.CombinedOutput() // Ignore errors

	// Remove service file
	if err := os.Remove(servicePath); err != nil {
		return fmt.Errorf("failed to remove service file: %w", err)
	}

	// Reload systemd
	reloadCmd := exec.Command("systemctl", "--user", "daemon-reload")
	reloadCmd.CombinedOutput()

	fmt.Println("✅ Service uninstalled successfully!")
	return nil
}

func runServiceStatus(cmd *cobra.Command, args []string) error {
	switch runtime.GOOS {
	case "darwin":
		return statusLaunchdService()
	case "linux":
		return statusSystemdService()
	case "windows":
		fmt.Println("📋 Service Status: Not applicable on Windows")
		fmt.Println()
		fmt.Println("To check if clara_companion is running:")
		fmt.Println("  tasklist | findstr clara_companion")
		return nil
	default:
		return fmt.Errorf("service status not supported on %s", runtime.GOOS)
	}
}

func statusLaunchdService() error {
	plistPath, err := getLaunchdPlistPath()
	if err != nil {
		return err
	}

	// Check if plist exists
	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		fmt.Println("📋 Service Status: Not installed")
		fmt.Println()
		fmt.Println("   Install with: clara_companion service install")
		return nil
	}

	// Check if running
	listCmd := exec.Command("launchctl", "list", "com.claraverse.clara-companion")
	output, err := listCmd.CombinedOutput()

	if err != nil {
		fmt.Println("📋 Service Status: Installed but not running")
		fmt.Printf("   Location: %s\n", plistPath)
		fmt.Println()
		fmt.Println("   Start with: clara_companion service start")
	} else {
		fmt.Println("📋 Service Status: ✅ Running")
		fmt.Printf("   Location: %s\n", plistPath)
		fmt.Println()
		fmt.Println("   Details:")
		fmt.Printf("   %s\n", string(output))
	}

	return nil
}

func statusSystemdService() error {
	servicePath, err := getSystemdServicePath()
	if err != nil {
		return err
	}

	// Check if service file exists
	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		fmt.Println("📋 Service Status: Not installed")
		fmt.Println()
		fmt.Println("   Install with: clara_companion service install")
		return nil
	}

	// Check status
	statusCmd := exec.Command("systemctl", "--user", "status", "claraverse-mcp.service")
	output, _ := statusCmd.CombinedOutput()

	fmt.Println("📋 Service Status:")
	fmt.Printf("   Location: %s\n", servicePath)
	fmt.Println()
	fmt.Println(string(output))

	return nil
}

func runServiceStart(cmd *cobra.Command, args []string) error {
	switch runtime.GOOS {
	case "darwin":
		plistPath, err := getLaunchdPlistPath()
		if err != nil {
			return err
		}
		startCmd := exec.Command("launchctl", "load", plistPath)
		if output, err := startCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to start service: %s", string(output))
		}
		fmt.Println("✅ Service started!")
		return nil

	case "linux":
		startCmd := exec.Command("systemctl", "--user", "start", "claraverse-mcp.service")
		if output, err := startCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to start service: %s", string(output))
		}
		fmt.Println("✅ Service started!")
		return nil

	case "windows":
		fmt.Println("ℹ️  Background service is not available on Windows.")
		fmt.Println("Use 'clara_companion start' to run manually.")
		return nil

	default:
		return fmt.Errorf("service start not supported on %s", runtime.GOOS)
	}
}

func runServiceStop(cmd *cobra.Command, args []string) error {
	switch runtime.GOOS {
	case "darwin":
		plistPath, err := getLaunchdPlistPath()
		if err != nil {
			return err
		}
		stopCmd := exec.Command("launchctl", "unload", plistPath)
		if output, err := stopCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to stop service: %s", string(output))
		}
		fmt.Println("✅ Service stopped!")
		return nil

	case "linux":
		stopCmd := exec.Command("systemctl", "--user", "stop", "claraverse-mcp.service")
		if output, err := stopCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to stop service: %s", string(output))
		}
		fmt.Println("✅ Service stopped!")
		return nil

	case "windows":
		fmt.Println("ℹ️  Background service is not available on Windows.")
		fmt.Println("To stop a running clara_companion:")
		fmt.Println("  taskkill /IM clara_companion.exe /F")
		return nil

	default:
		return fmt.Errorf("service stop not supported on %s", runtime.GOOS)
	}
}
