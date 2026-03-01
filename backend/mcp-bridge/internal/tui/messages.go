package tui

import (
	"time"
)

// TabIndex represents the currently active tab
type TabIndex int

const (
	TabDashboard TabIndex = iota
	TabServers
	TabSettings
)

// String returns the tab name
func (t TabIndex) String() string {
	switch t {
	case TabDashboard:
		return "Dashboard"
	case TabServers:
		return "Servers"
	case TabSettings:
		return "Settings"
	default:
		return "Unknown"
	}
}

// TabNames returns all tab names
func TabNames() []string {
	return []string{"Dashboard", "Servers", "Settings"}
}

const TabCount = 3

// ConnectionStatus represents the connection state
type ConnectionStatus int

const (
	StatusDisconnected ConnectionStatus = iota
	StatusConnecting
	StatusConnected
	StatusReconnecting
	StatusError
)

// String returns a human-readable status
func (s ConnectionStatus) String() string {
	switch s {
	case StatusDisconnected:
		return "Disconnected"
	case StatusConnecting:
		return "Connecting..."
	case StatusConnected:
		return "Connected"
	case StatusReconnecting:
		return "Reconnecting..."
	case StatusError:
		return "Error"
	default:
		return "Unknown"
	}
}

// IsConnected returns true if status indicates active connection
func (s ConnectionStatus) IsConnected() bool {
	return s == StatusConnected
}

// Custom messages for Bubble Tea

// StatusUpdateMsg is sent when connection status changes
type StatusUpdateMsg struct {
	Status    ConnectionStatus
	BackendURL string
	Error     error
}

// ServerUpdateMsg is sent when server list changes
type ServerUpdateMsg struct {
	Servers []ServerInfo
}

// ServerInfo represents an MCP server
type ServerInfo struct {
	Name        string
	Description string
	Command     string
	Args        []string
	Enabled     bool
	Connected   bool
	ToolCount   int
	Error       string
}

// ActivityMsg is sent when a tool is executed
type ActivityMsg struct {
	Entry ActivityEntry
}

// ActivityEntry represents a single activity log entry
type ActivityEntry struct {
	Timestamp time.Time
	ToolName  string
	Arguments string
	Success   bool
	Latency   time.Duration
	Error     string
	ServerName string
}

// DeviceUpdateMsg is sent when device list changes
type DeviceUpdateMsg struct {
	Devices []DeviceInfo
}

// DeviceInfo represents a connected device
type DeviceInfo struct {
	ID         string
	Name       string
	Email      string
	LastActive time.Time
	Current    bool
}

// TokenUpdateMsg is sent when token info changes
type TokenUpdateMsg struct {
	ExpiresAt time.Time
	UserEmail string
	UserID    string
}

// TickMsg is sent periodically for updates
type TickMsg struct {
	Time time.Time
}

// InstallCancelledMsg is sent when installation is cancelled
type InstallCancelledMsg struct {
	Name string
}

// BootTickMsg is sent during boot animation
type BootTickMsg struct {
	Time time.Time
}

// ErrorMsg indicates an error occurred
type ErrorMsg struct {
	Error   error
	Context string
}

// ConfirmMsg is sent when user confirms an action
type ConfirmMsg struct {
	Action   string
	Payload  interface{}
	Confirmed bool
}

// ToastMsg displays a notification
type ToastMsg struct {
	Message  string
	Duration time.Duration
	Type     ToastType
}

// ToastType defines the toast notification style
type ToastType int

const (
	ToastInfo ToastType = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// ModalAction represents an action from a modal
type ModalAction int

const (
	ModalCancel ModalAction = iota
	ModalConfirm
)

// ModalResultMsg is sent when modal is dismissed
type ModalResultMsg struct {
	Action  ModalAction
	Payload interface{}
}

// WindowSizeMsg is sent when terminal size changes
type WindowSizeMsg struct {
	Width  int
	Height int
}

// RefreshMsg triggers a data refresh
type RefreshMsg struct{}

// QuitMsg signals the app should quit
type QuitMsg struct{}

// FocusMsg changes input focus
type FocusMsg struct {
	Component string
}

// InputSubmitMsg is sent when text input is submitted
type InputSubmitMsg struct {
	Field string
	Value string
}

// ServerActionMsg represents a server management action
type ServerActionMsg struct {
	Action     ServerAction
	ServerName string
	Success    bool
	Error      error
}

// ServerAction defines server management actions
type ServerAction int

const (
	ServerActionAdd ServerAction = iota
	ServerActionRemove
	ServerActionEnable
	ServerActionDisable
	ServerActionRestart
)

// String returns the action name
func (a ServerAction) String() string {
	switch a {
	case ServerActionAdd:
		return "add"
	case ServerActionRemove:
		return "remove"
	case ServerActionEnable:
		return "enable"
	case ServerActionDisable:
		return "disable"
	case ServerActionRestart:
		return "restart"
	default:
		return "unknown"
	}
}

// SettingUpdateMsg is sent when a setting changes
type SettingUpdateMsg struct {
	Key   string
	Value interface{}
}

// InitCompleteMsg is sent when initialization is complete
type InitCompleteMsg struct {
	Config    interface{}
	Connected bool
}

// QuitMode represents how the app should exit
type QuitMode int

const (
	QuitModeAsk        QuitMode = iota // Show quit dialog
	QuitModeBackground                 // Keep running in background
	QuitModeFull                       // Full shutdown
)

// QuitResultMsg is sent when quit dialog is resolved
type QuitResultMsg struct {
	Mode QuitMode
}

// RerunSetupMsg signals that the user wants to re-run the setup wizard.
type RerunSetupMsg struct{}

// UpdateAvailableMsg is sent when a newer version is found on GitHub.
type UpdateAvailableMsg struct {
	Version string
}

// DaemonStatusMsg is sent when connected to daemon
type DaemonStatusMsg struct {
	Connected  bool
	Status     string
	BackendURL string
	ToolCount  int
}

