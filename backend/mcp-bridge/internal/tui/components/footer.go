package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/claraverse/mcp-client/internal/tui"
)

// Footer represents the bottom status bar component
type Footer struct {
	theme *tui.Theme
}

// NewFooter creates a new footer component
func NewFooter() *Footer {
	return &Footer{
		theme: tui.DefaultTheme,
	}
}

// View renders the footer
func (f *Footer) View(width int, helpText string, status tui.ConnectionStatus) string {
	// Status indicator
	var statusStr string
	switch status {
	case tui.StatusConnected:
		statusStr = f.theme.StatusSuccess.Render("●") + " Connected"
	case tui.StatusConnecting, tui.StatusReconnecting:
		statusStr = f.theme.StatusWarning.Render("●") + " " + status.String()
	default:
		statusStr = f.theme.StatusError.Render("●") + " " + status.String()
	}

	padding := width - lipgloss.Width(helpText) - lipgloss.Width(statusStr) - 4
	if padding < 0 {
		padding = 0
	}

	footer := helpText + strings.Repeat(" ", padding) + statusStr

	return f.theme.FooterContainer.Width(width).Render(footer)
}

// HelpText contains help text builders for each view
type HelpText struct {
	theme *tui.Theme
}

// NewHelpText creates a new help text builder
func NewHelpText() *HelpText {
	return &HelpText{
		theme: tui.DefaultTheme,
	}
}

// KeyHelp renders a single key binding hint
func (h *HelpText) KeyHelp(key, label string) string {
	return h.theme.HelpKey.Render(key) + " " + h.theme.Help.Render(label)
}

// Dashboard returns help text for the dashboard view
func (h *HelpText) Dashboard() string {
	return h.KeyHelp("1-3", "tabs") + "  " +
		h.KeyHelp("r", "refresh") + "  " +
		h.KeyHelp("q", "quit")
}

// Servers returns help text for the servers view
func (h *HelpText) Servers() string {
	return h.KeyHelp("1-3", "tabs") + "  " +
		h.KeyHelp("a", "add") + "  " +
		h.KeyHelp("d", "remove") + "  " +
		h.KeyHelp("e", "toggle") + "  " +
		h.KeyHelp("q", "quit")
}

// Settings returns help text for the settings view
func (h *HelpText) Settings() string {
	return h.KeyHelp("1-3", "tabs") + "  " +
		h.KeyHelp("enter", "edit") + "  " +
		h.KeyHelp("q", "quit")
}

// ForTab returns help text for a specific tab
func (h *HelpText) ForTab(tab tui.TabIndex) string {
	switch tab {
	case tui.TabDashboard:
		return h.Dashboard()
	case tui.TabServers:
		return h.Servers()
	case tui.TabSettings:
		return h.Settings()
	default:
		return h.KeyHelp("q", "quit")
	}
}
