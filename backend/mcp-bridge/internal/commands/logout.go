package commands

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/claraverse/mcp-client/internal/config"
	"github.com/spf13/cobra"
)

var LogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out from ClaraVerse",
	Long:  `Log out from ClaraVerse and revoke this device's access.`,
	RunE:  runLogout,
}

func runLogout(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Device == nil || cfg.Device.DeviceID == "" {
		// Just clear local auth
		cfg.AuthToken = ""
		cfg.RefreshToken = ""
		cfg.Device = nil
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to clear credentials: %w", err)
		}
		fmt.Println("✅ Logged out (no device was registered).")
		return nil
	}

	// Revoke device on server
	apiURL := getAPIURL(cfg)

	req, err := http.NewRequest("DELETE", apiURL+"/api/devices/"+cfg.Device.DeviceID, nil)
	if err != nil {
		// Just clear locally if we can't create the request
		clearLocalAuth(cfg)
		fmt.Println("✅ Logged out locally.")
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+cfg.AuthToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// Server unreachable, just clear locally
		clearLocalAuth(cfg)
		fmt.Println("✅ Logged out locally (could not reach server to revoke device).")
		return nil
	}
	defer resp.Body.Close()

	// Read response but don't fail if it errors
	body, _ := io.ReadAll(resp.Body)

	// Clear local auth regardless of server response
	clearLocalAuth(cfg)

	if resp.StatusCode == http.StatusOK {
		fmt.Println("✅ Logged out and device revoked.")
	} else if resp.StatusCode == http.StatusUnauthorized {
		fmt.Println("✅ Logged out (session was already expired).")
	} else {
		fmt.Printf("✅ Logged out locally (server response: %s).\n", strings.TrimSpace(string(body)))
	}

	return nil
}

func clearLocalAuth(cfg *config.Config) {
	cfg.AuthToken = ""
	cfg.RefreshToken = ""
	cfg.Device = nil
	_ = config.Save(cfg)
}
