package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// DefaultBackendURL is the default WebSocket URL for the ClaraVerse backend.
// Override at build time with: go build -ldflags "-X github.com/claraverse/mcp-client/internal/config.DefaultBackendURL=ws://localhost:3001/mcp/connect"
var DefaultBackendURL = "wss://claraverse.app/mcp/connect"

// DefaultAPIBaseURL returns the HTTP(S) base URL derived from the default backend URL.
// e.g. "wss://claraverse.app/mcp/connect" → "https://claraverse.app"
//      "ws://localhost:3001/mcp/connect"   → "http://localhost:3001"
func DefaultAPIBaseURL() string {
	u := DefaultBackendURL
	// Strip WebSocket path (try longest suffix first, stop after first match)
	for _, suffix := range []string{"/mcp/connect", "/mcp"} {
		if strings.HasSuffix(u, suffix) {
			u = u[:len(u)-len(suffix)]
			break
		}
	}
	// Convert ws(s):// to http(s)://
	if strings.HasPrefix(u, "wss://") {
		return "https://" + u[6:]
	}
	if strings.HasPrefix(u, "ws://") {
		return "http://" + u[5:]
	}
	return u
}

// Config represents the application configuration
type Config struct {
	BackendURL   string      `yaml:"backend_url" mapstructure:"backend_url"`
	AuthToken    string      `yaml:"auth_token" mapstructure:"auth_token"`
	RefreshToken string      `yaml:"refresh_token,omitempty" mapstructure:"refresh_token"`
	TokenExpiry  int64       `yaml:"token_expiry,omitempty" mapstructure:"token_expiry"` // Unix timestamp
	UserID       string      `yaml:"user_id" mapstructure:"user_id"`
	MCPServers   []MCPServer `yaml:"mcp_servers" mapstructure:"mcp_servers"`

	// Device authorization (OAuth 2.0 Device Grant)
	Device *DeviceConfig `yaml:"device,omitempty" mapstructure:"device"`

	// LLM providers (user's own API keys — optional)
	Providers []ProviderConfig `yaml:"providers,omitempty" mapstructure:"providers"`

	// Agent brain settings
	Agent *AgentConfig `yaml:"agent,omitempty" mapstructure:"agent"`

	// Browser configuration (set by first-run wizard)
	Browser *BrowserConfig `yaml:"browser,omitempty" mapstructure:"browser"`

	// Whether the first-run setup wizard has been completed
	FirstRunCompleted bool `yaml:"first_run_completed" mapstructure:"first_run_completed"`
}

// BrowserConfig holds Chrome browser settings for remote debugging.
type BrowserConfig struct {
	ProfilePath string `yaml:"profile_path" mapstructure:"profile_path"`
	Port        int    `yaml:"port" mapstructure:"port"`
	AutoLaunch  bool   `yaml:"auto_launch" mapstructure:"auto_launch"`
}

// ChromeProfile represents a discovered Chrome profile on disk.
type ChromeProfile struct {
	Name string // Display name (e.g. "Default", "Work")
	Path string // Full path to the profile directory
}

// DeviceConfig holds device-specific authentication data
type DeviceConfig struct {
	DeviceID     string `yaml:"device_id" mapstructure:"device_id"`
	RefreshToken string `yaml:"refresh_token" mapstructure:"refresh_token"`
	UserID       string `yaml:"user_id" mapstructure:"user_id"`
	UserEmail    string `yaml:"user_email" mapstructure:"user_email"`
}

// MCPServer represents a configured MCP server
type MCPServer struct {
	Name        string                 `yaml:"name" mapstructure:"name"`
	Path        string                 `yaml:"path,omitempty" mapstructure:"path"`       // For executable path
	Command     string                 `yaml:"command,omitempty" mapstructure:"command"` // For command-based (e.g., "npx")
	Args        []string               `yaml:"args,omitempty" mapstructure:"args"`       // Command arguments
	URL         string                 `yaml:"url,omitempty" mapstructure:"url"`
	Type        string                 `yaml:"type" mapstructure:"type"` // "stdio" or "sse"
	Config      map[string]interface{} `yaml:"config,omitempty" mapstructure:"config"`
	Enabled     bool                   `yaml:"enabled" mapstructure:"enabled"`
	Description string                 `yaml:"description,omitempty" mapstructure:"description"`
}

// ProviderConfig holds configuration for a direct LLM provider (user's own API keys).
type ProviderConfig struct {
	Name         string `yaml:"name" mapstructure:"name"`
	BaseURL      string `yaml:"base_url" mapstructure:"base_url"`
	APIKey       string `yaml:"api_key" mapstructure:"api_key"`
	DefaultModel string `yaml:"default_model" mapstructure:"default_model"`
	Enabled      bool   `yaml:"enabled" mapstructure:"enabled"`
}

// AgentConfig holds settings for the local agent brain.
type AgentConfig struct {
	DefaultProvider     string  `yaml:"default_provider" mapstructure:"default_provider"`
	MaxIterations       int     `yaml:"max_iterations" mapstructure:"max_iterations"`
	MaxParallelDaemons  int     `yaml:"max_parallel_daemons" mapstructure:"max_parallel_daemons"`
	MaxConcurrentTasks  int     `yaml:"max_concurrent_tasks" mapstructure:"max_concurrent_tasks"`
	ContextWindow       int     `yaml:"context_window" mapstructure:"context_window"`
	CompactionThreshold float64 `yaml:"compaction_threshold" mapstructure:"compaction_threshold"`
}

var (
	configPath string
	configDir  string
)

func init() {
	// When running under sudo, os.UserHomeDir() returns /root.
	// Check SUDO_USER to resolve the real user's home directory.
	var home string
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		if u, err := user.Lookup(sudoUser); err == nil {
			home = u.HomeDir
		}
	}
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			panic(fmt.Sprintf("failed to get home directory: %v", err))
		}
	}

	configDir = filepath.Join(home, ".claraverse")
	configPath = filepath.Join(configDir, "mcp-config.yaml")
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	return configPath
}

// GetConfigDir returns the config directory
func GetConfigDir() string {
	return configDir
}

// Load loads the configuration from file
func Load() (*Config, error) {
	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		defaultConfig := &Config{
			BackendURL: DefaultBackendURL,
			MCPServers: []MCPServer{},
		}
		if err := Save(defaultConfig); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return defaultConfig, nil
	}

	// Read config file
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// Delete removes the config file so the next Load() creates a fresh one
// with the build-time DefaultBackendURL.
func Delete() {
	os.Remove(configPath)
}

// Save saves the configuration to file
func Save(cfg *Config) error {
	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file with secure permissions
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// AddServer adds a new MCP server to the configuration
func (c *Config) AddServer(server MCPServer) error {
	// Check if server already exists
	for i, s := range c.MCPServers {
		if s.Name == server.Name {
			// Update existing server
			c.MCPServers[i] = server
			return nil
		}
	}

	// Add new server
	c.MCPServers = append(c.MCPServers, server)
	return nil
}

// RemoveServer removes an MCP server by name
func (c *Config) RemoveServer(name string) error {
	for i, s := range c.MCPServers {
		if s.Name == name {
			c.MCPServers = append(c.MCPServers[:i], c.MCPServers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("server %s not found", name)
}

// GetServer retrieves a server by name
func (c *Config) GetServer(name string) (*MCPServer, error) {
	for _, s := range c.MCPServers {
		if s.Name == name {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("server %s not found", name)
}

// GetEnabledServers returns only enabled servers
func (c *Config) GetEnabledServers() []MCPServer {
	var enabled []MCPServer
	for _, s := range c.MCPServers {
		if s.Enabled {
			enabled = append(enabled, s)
		}
	}
	return enabled
}

// ListChromeProfiles scans the system for existing Chrome profile directories.
// Returns profiles sorted by name. The caller should prepend the "New isolated profile"
// option separately.
func ListChromeProfiles() []ChromeProfile {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var chromeDir string
	switch runtime.GOOS {
	case "linux":
		chromeDir = filepath.Join(home, ".config", "google-chrome")
	case "darwin":
		chromeDir = filepath.Join(home, "Library", "Application Support", "Google", "Chrome")
	default:
		return nil
	}

	// Check if Chrome data directory exists
	if _, err := os.Stat(chromeDir); os.IsNotExist(err) {
		return nil
	}

	// Scan for profile directories: "Default", "Profile 1", "Profile 2", etc.
	var profiles []ChromeProfile

	entries, err := os.ReadDir(chromeDir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name != "Default" && !isProfileDir(name) {
			continue
		}

		profilePath := filepath.Join(chromeDir, name)
		displayName := readChromeProfileName(profilePath)
		if displayName == "" {
			displayName = name
		}

		profiles = append(profiles, ChromeProfile{
			Name: displayName,
			Path: profilePath,
		})
	}

	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})

	return profiles
}

// isProfileDir checks if a directory name matches Chrome's profile naming pattern.
func isProfileDir(name string) bool {
	// Chrome uses "Profile 1", "Profile 2", etc.
	return len(name) > 8 && name[:8] == "Profile "
}

// readChromeProfileName reads the display name from a Chrome profile's Preferences file.
func readChromeProfileName(profilePath string) string {
	prefsPath := filepath.Join(profilePath, "Preferences")
	data, err := os.ReadFile(prefsPath)
	if err != nil {
		return ""
	}

	var prefs struct {
		Profile struct {
			Name string `json:"name"`
		} `json:"profile"`
	}
	if err := json.Unmarshal(data, &prefs); err != nil {
		return ""
	}
	return prefs.Profile.Name
}

// DefaultIsolatedProfilePath returns the path for the isolated Clara Companion Chrome profile.
func DefaultIsolatedProfilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "clara-companion", "chrome-profile")
}
