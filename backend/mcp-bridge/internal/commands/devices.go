package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/claraverse/mcp-client/internal/config"
	"github.com/spf13/cobra"
)

var DevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "Manage connected devices",
	Long:  `List, rename, or revoke devices connected to your ClaraVerse account.`,
}

var devicesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all connected devices",
	RunE:  runDevicesList,
}

var devicesRevokeCmd = &cobra.Command{
	Use:   "revoke <device-id>",
	Short: "Revoke a device's access",
	Args:  cobra.ExactArgs(1),
	RunE:  runDevicesRevoke,
}

var devicesRenameCmd = &cobra.Command{
	Use:   "rename <device-id> <new-name>",
	Short: "Rename a device",
	Args:  cobra.ExactArgs(2),
	RunE:  runDevicesRename,
}

func init() {
	DevicesCmd.AddCommand(devicesListCmd)
	DevicesCmd.AddCommand(devicesRevokeCmd)
	DevicesCmd.AddCommand(devicesRenameCmd)
}

// DeviceInfo represents a device from the API
type DeviceInfo struct {
	DeviceID     string    `json:"device_id"`
	Name         string    `json:"name"`
	Platform     string    `json:"platform"`
	Version      string    `json:"client_version"`
	IsActive     bool      `json:"is_active"`
	IsCurrent    bool      `json:"is_current"`
	LastActiveAt time.Time `json:"last_active_at"`
	LastIP       string    `json:"last_ip"`
	LastLocation string    `json:"last_location"`
	CreatedAt    time.Time `json:"created_at"`
}

// DeviceListResponse from GET /api/devices
type DeviceListResponse struct {
	Devices []DeviceInfo `json:"devices"`
}

func getAPIURL(cfg *config.Config) string {
	backendURL := cfg.BackendURL
	if backendURL == "" {
		backendURL = config.DefaultBackendURL
	}
	apiURL := strings.Replace(backendURL, "ws://", "http://", 1)
	apiURL = strings.Replace(apiURL, "wss://", "https://", 1)
	return strings.TrimSuffix(apiURL, "/mcp/connect")
}

func runDevicesList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.AuthToken == "" {
		return fmt.Errorf("not logged in. Run: clara_companion login")
	}

	apiURL := getAPIURL(cfg)

	req, err := http.NewRequest("GET", apiURL+"/api/devices", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.AuthToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		// Try refreshing token
		if cfg.Device != nil && cfg.Device.RefreshToken != "" {
			if err := RefreshDeviceToken(cfg); err != nil {
				return fmt.Errorf("session expired. Please run: clara_companion login")
			}
			// Retry with new token
			return runDevicesList(cmd, args)
		}
		return fmt.Errorf("session expired. Please run: clara_companion login")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to list devices: %s (status %d)", string(body), resp.StatusCode)
	}

	var devicesResp DeviceListResponse
	if err := json.Unmarshal(body, &devicesResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if len(devicesResp.Devices) == 0 {
		fmt.Println("No devices connected.")
		return nil
	}

	fmt.Println("📱 Connected Devices")
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Println()

	for _, device := range devicesResp.Devices {
		// Status indicator
		status := "🟢"
		if !device.IsActive {
			status = "🔴"
		}

		// Current device indicator
		current := ""
		if device.IsCurrent {
			current = " \033[1;33m(THIS DEVICE)\033[0m"
		}

		// Platform emoji
		platformEmoji := "💻"
		switch device.Platform {
		case "darwin":
			platformEmoji = "🍎"
		case "linux":
			platformEmoji = "🐧"
		case "windows":
			platformEmoji = "🪟"
		}

		fmt.Printf("%s %s %s%s\n", status, platformEmoji, device.Name, current)
		fmt.Printf("   ID: %s\n", device.DeviceID)
		fmt.Printf("   Platform: %s • Version: %s\n", device.Platform, device.Version)

		// Last active
		if !device.LastActiveAt.IsZero() {
			ago := time.Since(device.LastActiveAt)
			agoStr := formatDuration(ago)
			fmt.Printf("   Last active: %s ago", agoStr)
			if device.LastLocation != "" {
				fmt.Printf(" • %s", device.LastLocation)
			}
			fmt.Println()
		}

		fmt.Println()
	}

	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Println("Commands: clara_companion devices revoke <id> | clara_companion devices rename <id> <name>")

	return nil
}

func runDevicesRevoke(cmd *cobra.Command, args []string) error {
	deviceID := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.AuthToken == "" {
		return fmt.Errorf("not logged in. Run: clara_companion login")
	}

	// Check if trying to revoke current device
	if cfg.Device != nil && cfg.Device.DeviceID == deviceID {
		fmt.Println("⚠️  You are about to revoke THIS device.")
		fmt.Print("This will log you out. Continue? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	apiURL := getAPIURL(cfg)

	req, err := http.NewRequest("DELETE", apiURL+"/api/devices/"+deviceID, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.AuthToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("session expired. Please run: clara_companion login")
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("device not found: %s", deviceID)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to revoke device: %s (status %d)", string(body), resp.StatusCode)
	}

	fmt.Printf("✅ Device %s has been revoked.\n", deviceID)

	// If we revoked our own device, clear local config
	if cfg.Device != nil && cfg.Device.DeviceID == deviceID {
		cfg.AuthToken = ""
		cfg.RefreshToken = ""
		cfg.Device = nil
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to clear local credentials: %w", err)
		}
		fmt.Println("📤 Logged out locally.")
	}

	return nil
}

func runDevicesRename(cmd *cobra.Command, args []string) error {
	deviceID := args[0]
	newName := args[1]

	if len(newName) > 50 {
		newName = newName[:50]
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.AuthToken == "" {
		return fmt.Errorf("not logged in. Run: clara_companion login")
	}

	apiURL := getAPIURL(cfg)

	reqBody := map[string]string{"name": newName}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req, err := http.NewRequest("PUT", apiURL+"/api/devices/"+deviceID, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.AuthToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("session expired. Please run: clara_companion login")
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("device not found: %s", deviceID)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to rename device: %s (status %d)", string(body), resp.StatusCode)
	}

	fmt.Printf("✅ Device renamed to: %s\n", newName)
	return nil
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	} else if d < time.Hour {
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", mins)
	} else if d < 24*time.Hour {
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	} else {
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	}
}
