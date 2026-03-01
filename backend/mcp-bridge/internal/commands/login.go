package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/claraverse/mcp-client/internal/config"
	"github.com/spf13/cobra"
)

var (
	noBrowser      bool
	skipServers    bool
	skipService    bool
	noStart        bool
	nonInteractive bool
)

// Styles for the TUI
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("63")).
			MarginBottom(1)

	stepStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 2)
)

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with ClaraVerse",
	Long: `Authenticate with your ClaraVerse account using device authorization.

After authentication, the interactive dashboard will launch automatically.

Works with any login method (email, Google, GitHub, etc.)`,
	RunE: runLogin,
}

func init() {
	LoginCmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Don't automatically open browser")
	LoginCmd.Flags().BoolVar(&skipServers, "skip-servers", false, "Skip MCP server selection")
	LoginCmd.Flags().BoolVar(&skipService, "skip-service", false, "Skip service installation prompt")
	LoginCmd.Flags().BoolVar(&noStart, "no-start", false, "Don't auto-start the client")
	LoginCmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Use defaults for all prompts")
}

// DeviceCodeResponse from POST /api/device/code
type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// DeviceTokenResponse from GET /api/device/token
type DeviceTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	DeviceID     string `json:"device_id"`
	User         struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

// DeviceTokenErrorResponse for polling errors
type DeviceTokenErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Clear screen and show header
	fmt.Print("\033[H\033[2J")
	printHeader()

	// Delete existing config and start fresh so the build-time
	// DefaultBackendURL (dev vs prod) is always used.
	config.Delete()
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Prompt for server URL
	backendURL, err := promptServerURL()
	if err != nil {
		return err
	}
	cfg.BackendURL = backendURL
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Run device auth
	if err := runDeviceAuth(cfg); err != nil {
		return err
	}

	// Launch TUI
	fmt.Println(successStyle.Render("Launching Clara Companion..."))
	fmt.Println()
	return RunTUIDefault()
}

func promptServerURL() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(stepStyle.Render("  Server Connection"))
	fmt.Println()

	// Determine default from build-time or fallback
	defaultWS := config.DefaultBackendURL
	defaultHTTP := config.DefaultAPIBaseURL()
	if defaultHTTP == "" {
		defaultHTTP = "http://localhost:3000"
		defaultWS = "ws://localhost:3000/mcp/connect"
	}

	fmt.Printf("  [1] Default (%s)\n", defaultHTTP)
	fmt.Println("  [2] Custom URL")
	fmt.Println()
	fmt.Print("  Choose (1/2) [1]: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if choice == "" || choice == "1" {
		fmt.Printf("  → Using %s\n\n", defaultHTTP)
		return defaultWS, nil
	}

	fmt.Print("  Enter your ClaraVerse URL (e.g. https://my-server.com): ")
	customURL, _ := reader.ReadString('\n')
	customURL = strings.TrimSpace(customURL)

	if customURL == "" {
		return defaultWS, nil
	}

	// Normalize: strip trailing slash
	customURL = strings.TrimRight(customURL, "/")

	// Convert HTTP URL to WS URL for the backend connection
	wsURL := customURL
	if strings.HasPrefix(wsURL, "https://") {
		wsURL = "wss://" + strings.TrimPrefix(wsURL, "https://")
	} else if strings.HasPrefix(wsURL, "http://") {
		wsURL = "ws://" + strings.TrimPrefix(wsURL, "http://")
	} else {
		// Assume plain hostname, use ws://
		wsURL = "ws://" + wsURL
	}
	wsURL += "/mcp/connect"

	fmt.Printf("  → Connecting to %s\n\n", customURL)
	return wsURL, nil
}

func printHeader() {
	header := boxStyle.Render("Clara Companion Login")
	fmt.Println(titleStyle.Render(header))
	fmt.Println()
}

func printStepHeader(step, total int, title string) {
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println(stepStyle.Render(fmt.Sprintf(" STEP %d/%d: %s", step, total, title)))
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()
}

func runDeviceAuth(cfg *config.Config) error {
	printStepHeader(1, 1, "Authentication")

	// Get backend URL (strip ws:// and use https://)
	backendURL := cfg.BackendURL
	if backendURL == "" {
		backendURL = config.DefaultBackendURL
	}

	// Convert WS URL to HTTP API URL
	apiURL := strings.Replace(backendURL, "ws://", "http://", 1)
	apiURL = strings.Replace(apiURL, "wss://", "https://", 1)
	apiURL = strings.TrimSuffix(apiURL, "/mcp/connect")

	fmt.Println("🔐 Starting device authorization...")
	fmt.Println()

	// Step 1: Request device code
	deviceCode, err := requestDeviceCode(apiURL)
	if err != nil {
		return fmt.Errorf("failed to get device code: %w", err)
	}

	// Display the code prominently
	fmt.Println("To authenticate, open this URL in your browser:")
	fmt.Printf("  \033[1;36m%s\033[0m\n", deviceCode.VerificationURI)
	fmt.Println()
	fmt.Println("Then enter this code:")
	fmt.Println()
	displayCode(deviceCode.UserCode)
	fmt.Println()

	// Auto-open browser unless disabled
	if !noBrowser {
		fmt.Println("Opening browser...")
		if err := openBrowser(deviceCode.VerificationURIComplete); err != nil {
			fmt.Printf("⚠️  Could not open browser: %v\n", err)
			fmt.Printf("Please open the URL manually: %s\n", deviceCode.VerificationURIComplete)
		}
	}

	// Calculate expiry time
	expiresAt := time.Now().Add(time.Duration(deviceCode.ExpiresIn) * time.Second)

	fmt.Printf("Waiting for authorization... (expires in %d:%02d)\n",
		deviceCode.ExpiresIn/60, deviceCode.ExpiresIn%60)
	fmt.Println()

	// Step 2: Poll for token
	interval := deviceCode.Interval
	if interval < 5 {
		interval = 5 // Minimum 5 seconds per RFC 8628
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if expired
			if time.Now().After(expiresAt) {
				return fmt.Errorf("authorization code expired. Please try again")
			}

			// Update countdown
			remaining := time.Until(expiresAt)
			fmt.Printf("\r⏳ Waiting... (%d:%02d remaining)   ",
				int(remaining.Minutes()), int(remaining.Seconds())%60)

			// Poll for token
			tokenResp, errResp, err := pollForToken(apiURL, deviceCode.DeviceCode)
			if err != nil {
				return fmt.Errorf("polling error: %w", err)
			}

			if errResp != nil {
				switch errResp.Error {
				case "authorization_pending":
					// Continue polling
					continue
				case "slow_down":
					// Increase interval
					interval += 5
					ticker.Reset(time.Duration(interval) * time.Second)
					continue
				case "expired_token":
					return fmt.Errorf("authorization code expired. Please try again")
				case "access_denied":
					return fmt.Errorf("authorization was denied")
				default:
					return fmt.Errorf("authorization failed: %s", errResp.ErrorDescription)
				}
			}

			// Success!
			fmt.Println("\r                                        ") // Clear line
			fmt.Println()
			fmt.Println(successStyle.Render("✅ Authorization successful!"))
			fmt.Printf("📧 Logged in as: %s\n", tokenResp.User.Email)
			fmt.Printf("📱 Device ID: %s\n", tokenResp.DeviceID)
			fmt.Println()

			// Save to config
			cfg.AuthToken = tokenResp.AccessToken
			cfg.RefreshToken = tokenResp.RefreshToken
			cfg.TokenExpiry = time.Now().Unix() + int64(tokenResp.ExpiresIn)
			cfg.UserID = tokenResp.User.ID
			cfg.Device = &config.DeviceConfig{
				DeviceID:     tokenResp.DeviceID,
				RefreshToken: tokenResp.RefreshToken,
				UserID:       tokenResp.User.ID,
				UserEmail:    tokenResp.User.Email,
			}

			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			waitForEnter()
			return nil
		}
	}
}

func requestDeviceCode(apiURL string) (*DeviceCodeResponse, error) {
	reqBody := map[string]string{
		"client_id":      "clara_companion",
		"client_version": "1.0.0",
		"platform":       runtime.GOOS,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(apiURL+"/api/device/code", "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server error: %s (status %d)", string(body), resp.StatusCode)
	}

	var deviceCode DeviceCodeResponse
	if err := json.Unmarshal(body, &deviceCode); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &deviceCode, nil
}

func pollForToken(apiURL, deviceCode string) (*DeviceTokenResponse, *DeviceTokenErrorResponse, error) {
	url := fmt.Sprintf("%s/api/device/token?device_code=%s&client_id=clara_companion", apiURL, deviceCode)

	resp, err := http.Get(url)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for error response
	var errResp DeviceTokenErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
		return nil, &errResp, nil
	}

	// Check for success response
	var tokenResp DeviceTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		// Might be a pending response without explicit error
		return nil, &DeviceTokenErrorResponse{Error: "authorization_pending"}, nil
	}

	return &tokenResp, nil, nil
}

func displayCode(code string) {
	// Format code with dash in middle for readability (ABCD-1234)
	formatted := code
	if len(code) == 8 {
		formatted = code[:4] + "-" + code[4:]
	}

	// Display in a box
	fmt.Println("   ┌────────────────┐")
	fmt.Printf("   │   \033[1;32m%s\033[0m    │\n", formatted)
	fmt.Println("   └────────────────┘")
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		// Try xdg-open first, then fallback to common browsers
		if _, err := exec.LookPath("xdg-open"); err == nil {
			cmd = "xdg-open"
			args = []string{url}
		} else if _, err := exec.LookPath("google-chrome"); err == nil {
			cmd = "google-chrome"
			args = []string{url}
		} else if _, err := exec.LookPath("firefox"); err == nil {
			cmd = "firefox"
			args = []string{url}
		} else {
			return fmt.Errorf("no browser found")
		}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	// Don't wait for browser to close
	process := exec.Command(cmd, args...)
	process.Stdin = nil
	process.Stdout = nil
	process.Stderr = nil

	// Start in background
	if err := process.Start(); err != nil {
		return err
	}

	// Don't wait for it
	go func() {
		_ = process.Wait()
	}()

	return nil
}


func waitForEnter() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(dimStyle.Render("Press Enter to continue..."))
	reader.ReadString('\n')
	fmt.Println()
}

// RefreshDeviceToken refreshes the device token using the refresh token
func RefreshDeviceToken(cfg *config.Config) error {
	if cfg.Device == nil || cfg.Device.RefreshToken == "" || cfg.Device.DeviceID == "" {
		return fmt.Errorf("no device credentials found")
	}

	// Get API URL from config
	apiURL := cfg.BackendURL
	if apiURL == "" {
		apiURL = config.DefaultBackendURL
	}
	apiURL = strings.Replace(apiURL, "ws://", "http://", 1)
	apiURL = strings.Replace(apiURL, "wss://", "https://", 1)
	apiURL = strings.TrimSuffix(apiURL, "/mcp/connect")

	reqBody := map[string]string{
		"refresh_token": cfg.Device.RefreshToken,
		"device_id":     cfg.Device.DeviceID,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(apiURL+"/api/device/refresh", "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Check if device was revoked
		var errResp map[string]string
		if json.Unmarshal(body, &errResp) == nil {
			if errResp["error"] == "invalid_grant" {
				return fmt.Errorf("device has been revoked. Please run: clara_companion login")
			}
		}
		return fmt.Errorf("refresh failed: %s (status %d)", string(body), resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Update config with new tokens
	cfg.AuthToken = tokenResp.AccessToken
	cfg.RefreshToken = tokenResp.RefreshToken
	cfg.TokenExpiry = time.Now().Unix() + int64(tokenResp.ExpiresIn)
	cfg.Device.RefreshToken = tokenResp.RefreshToken

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

