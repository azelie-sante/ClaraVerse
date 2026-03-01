package tui

import (
	"sync"
	"time"
)

// AppState holds the shared application state
type AppState struct {
	mu sync.RWMutex

	// Connection info
	Status      ConnectionStatus
	BackendURL  string
	LastError   error
	ConnectedAt time.Time // When the current connection was established

	// User info
	UserEmail string
	UserID    string
	DeviceID  string
	TokenExp  time.Time

	// Servers
	Servers []ServerInfo

	// Activity log
	Activities []ActivityEntry
	maxActivities int

	// Devices
	Devices []DeviceInfo

	// Settings
	Settings map[string]interface{}

	// Version
	CurrentVersion string
	LatestVersion  string // non-empty when update available

	// UI state
	Width      int
	Height     int
	ActiveTab  TabIndex
	ShowHelp   bool
	ShowModal  bool
	ModalType  string
}

// NewAppState creates a new application state
func NewAppState() *AppState {
	return &AppState{
		Status:        StatusDisconnected,
		Servers:       make([]ServerInfo, 0),
		Activities:    make([]ActivityEntry, 0),
		maxActivities: 100,
		Devices:       make([]DeviceInfo, 0),
		Settings:      make(map[string]interface{}),
		Width:         80,
		Height:        24,
		ActiveTab:     TabDashboard,
	}
}

// SetStatus updates the connection status
func (s *AppState) SetStatus(status ConnectionStatus, backendURL string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if status == StatusConnected && s.Status != StatusConnected {
		s.ConnectedAt = time.Now()
	}
	s.Status = status
	s.BackendURL = backendURL
	s.LastError = err
}

// GetStatus returns the current connection status
func (s *AppState) GetStatus() (ConnectionStatus, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Status, s.BackendURL, s.LastError
}

// GetConnectedAt returns when the current connection was established
func (s *AppState) GetConnectedAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ConnectedAt
}

// SetUser updates user information
func (s *AppState) SetUser(email, userID, deviceID string, tokenExp time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.UserEmail = email
	s.UserID = userID
	s.DeviceID = deviceID
	s.TokenExp = tokenExp
}

// GetUser returns user information
func (s *AppState) GetUser() (string, string, string, time.Time) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.UserEmail, s.UserID, s.DeviceID, s.TokenExp
}

// TokenExpiresIn returns time until token expires
func (s *AppState) TokenExpiresIn() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Until(s.TokenExp)
}

// IsTokenExpiringSoon returns true if token expires within the given duration
func (s *AppState) IsTokenExpiringSoon(threshold time.Duration) bool {
	return s.TokenExpiresIn() < threshold
}

// SetServers updates the server list
func (s *AppState) SetServers(servers []ServerInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Servers = servers
}

// GetServers returns the server list
func (s *AppState) GetServers() []ServerInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]ServerInfo, len(s.Servers))
	copy(result, s.Servers)
	return result
}

// GetEnabledServers returns only enabled servers
func (s *AppState) GetEnabledServers() []ServerInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []ServerInfo
	for _, srv := range s.Servers {
		if srv.Enabled {
			result = append(result, srv)
		}
	}
	return result
}

// GetTotalToolCount returns the total number of tools across all servers
func (s *AppState) GetTotalToolCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	total := 0
	for _, srv := range s.Servers {
		if srv.Enabled && srv.Connected {
			total += srv.ToolCount
		}
	}
	return total
}

// AddActivity adds a new activity entry
func (s *AppState) AddActivity(entry ActivityEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Prepend the new entry
	s.Activities = append([]ActivityEntry{entry}, s.Activities...)

	// Trim if over limit
	if len(s.Activities) > s.maxActivities {
		s.Activities = s.Activities[:s.maxActivities]
	}
}

// GetActivities returns the activity log
func (s *AppState) GetActivities() []ActivityEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]ActivityEntry, len(s.Activities))
	copy(result, s.Activities)
	return result
}

// GetRecentActivities returns the last n activities
func (s *AppState) GetRecentActivities(n int) []ActivityEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if n > len(s.Activities) {
		n = len(s.Activities)
	}
	result := make([]ActivityEntry, n)
	copy(result, s.Activities[:n])
	return result
}

// ClearActivities clears the activity log
func (s *AppState) ClearActivities() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Activities = make([]ActivityEntry, 0)
}

// SetDevices updates the device list
func (s *AppState) SetDevices(devices []DeviceInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Devices = devices
}

// GetDevices returns the device list
func (s *AppState) GetDevices() []DeviceInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]DeviceInfo, len(s.Devices))
	copy(result, s.Devices)
	return result
}

// SetSetting updates a setting
func (s *AppState) SetSetting(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Settings[key] = value
}

// GetSetting returns a setting value
func (s *AppState) GetSetting(key string) interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Settings[key]
}

// GetSettingString returns a setting as string
func (s *AppState) GetSettingString(key string, defaultVal string) string {
	val := s.GetSetting(key)
	if str, ok := val.(string); ok {
		return str
	}
	return defaultVal
}

// GetSettingBool returns a setting as bool
func (s *AppState) GetSettingBool(key string, defaultVal bool) bool {
	val := s.GetSetting(key)
	if b, ok := val.(bool); ok {
		return b
	}
	return defaultVal
}

// SetSize updates the terminal size
func (s *AppState) SetSize(width, height int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Width = width
	s.Height = height
}

// GetSize returns the terminal size
func (s *AppState) GetSize() (int, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Width, s.Height
}

// ContentWidth returns the available content width
func (s *AppState) ContentWidth() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Width - 4 // Account for padding/borders
}

// ContentHeight returns the available content height
func (s *AppState) ContentHeight() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Height - 6 // Account for header, tabs, footer
}

// SetActiveTab sets the currently active tab
func (s *AppState) SetActiveTab(tab TabIndex) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ActiveTab = tab
}

// GetActiveTab returns the currently active tab
func (s *AppState) GetActiveTab() TabIndex {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ActiveTab
}

// Summary returns a summary of the application state
func (s *AppState) Summary() StateSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	enabledCount := 0
	connectedCount := 0
	totalTools := 0
	for _, srv := range s.Servers {
		if srv.Enabled {
			enabledCount++
			if srv.Connected {
				connectedCount++
				totalTools += srv.ToolCount
			}
		}
	}

	return StateSummary{
		Connected:      s.Status == StatusConnected,
		ServerCount:    len(s.Servers),
		EnabledCount:   enabledCount,
		ConnectedCount: connectedCount,
		TotalTools:     totalTools,
		ActivityCount:  len(s.Activities),
		DeviceCount:    len(s.Devices),
		TokenExpiresIn: time.Until(s.TokenExp),
	}
}

// StateSummary provides a quick overview of application state
type StateSummary struct {
	Connected      bool
	ServerCount    int
	EnabledCount   int
	ConnectedCount int
	TotalTools     int
	ActivityCount  int
	DeviceCount    int
	TokenExpiresIn time.Duration
}

