package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/claraverse/mcp-client/internal/config"
	"github.com/claraverse/mcp-client/internal/tui"
)

// SettingItem represents a configurable setting
type SettingItem struct {
	Key         string
	Label       string
	Value       string
	Description string
	Editable    bool
}

// Settings represents the settings view
type Settings struct {
	theme  *tui.Theme
	state  *tui.AppState
	cursor int
	items  []SettingItem
}

// NewSettings creates a new settings view
func NewSettings(state *tui.AppState) *Settings {
	versionValue := state.CurrentVersion
	if versionValue == "" {
		versionValue = "unknown"
	}
	versionDesc := "Currently installed version"
	if state.LatestVersion != "" {
		versionValue += "  ->  v" + state.LatestVersion + " available"
		versionDesc = "Run: curl -fsSL https://github.com/claraverse-space/clara-companion-releases | bash"
	}

	return &Settings{
		theme: tui.DefaultTheme,
		state: state,
		items: []SettingItem{
			{
				Key:         "version",
				Label:       "Version",
				Value:       versionValue,
				Description: versionDesc,
				Editable:    false,
			},
			{
				Key:         "backend_url",
				Label:       "Backend URL",
				Value:       config.DefaultBackendURL,
				Description: "WebSocket URL for MCP backend",
				Editable:    true,
			},
			{
				Key:         "reconnect_delay",
				Label:       "Reconnect Delay",
				Value:       "5s",
				Description: "Delay before reconnecting on disconnect",
				Editable:    true,
			},
			{
				Key:         "auto_start",
				Label:       "Auto-start",
				Value:       "Enabled",
				Description: "Automatically start bridge on login",
				Editable:    true,
			},
			{
				Key:         "config_path",
				Label:       "Config Path",
				Value:       "~/.claraverse/mcp-config.yaml",
				Description: "Location of configuration file",
				Editable:    false,
			},
			{
				Key:         "log_path",
				Label:       "Log Path",
				Value:       "~/.claraverse/logs/clara_companion.log",
				Description: "Location of log file",
				Editable:    false,
			},
			{
				Key:         "rerun_setup",
				Label:       "Re-run Setup",
				Value:       "",
				Description: "Re-run the first-time Chrome profile setup wizard",
				Editable:    true,
			},
		},
	}
}

// Update handles input for the settings view
func (s *Settings) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if s.cursor > 0 {
				s.cursor--
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if s.cursor < len(s.items)-1 {
				s.cursor++
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if s.cursor >= 0 && s.cursor < len(s.items) && s.items[s.cursor].Key == "rerun_setup" {
				return func() tea.Msg { return tui.RerunSetupMsg{} }
			}
		}
	}

	return nil
}

// View renders the settings view
func (s *Settings) View(width, height int) string {
	_, _, deviceID, _ := s.state.GetUser()

	// Device ID header
	header := s.theme.Title.Render("Settings") + "\n"
	header += s.theme.Label.Render("Device ID: ") + s.theme.ValueMuted.Render(deviceID) + "\n\n"

	// Settings list
	var rows []string
	for i, item := range s.items {
		isSelected := i == s.cursor

		var labelStyle, valueStyle lipgloss.Style
		if isSelected {
			labelStyle = s.theme.Label.Foreground(tui.ColorAccent)
			valueStyle = s.theme.Value.Bold(true)
		} else {
			labelStyle = s.theme.Label
			valueStyle = s.theme.Value
		}

		cursor := "  "
		if isSelected {
			cursor = s.theme.ListCursor.Render("> ")
		}

		editHint := ""
		if item.Editable && isSelected {
			editHint = s.theme.ValueMuted.Render(" (press Enter to edit)")
		} else if !item.Editable && isSelected {
			editHint = s.theme.ValueMuted.Render(" (read-only)")
		}

		row := fmt.Sprintf("%s%s: %s%s\n    %s",
			cursor,
			labelStyle.Render(item.Label),
			valueStyle.Render(item.Value),
			editHint,
			s.theme.ValueMuted.Render(item.Description),
		)
		rows = append(rows, row)
	}

	content := header + strings.Join(rows, "\n\n")

	help := "\n\n" + s.theme.Help.Render("[↑↓] navigate  [Enter] edit  [r] reset to defaults")

	return lipgloss.NewStyle().Padding(1, 2).Render(content + help)
}

// GetSelectedSetting returns the currently selected setting
func (s *Settings) GetSelectedSetting() *SettingItem {
	if s.cursor >= 0 && s.cursor < len(s.items) {
		item := s.items[s.cursor]
		return &item
	}
	return nil
}

// UpdateSetting updates a setting value
func (s *Settings) UpdateSetting(key, value string) {
	for i := range s.items {
		if s.items[i].Key == key {
			s.items[i].Value = value
			break
		}
	}
}

// LoadFromState loads settings from the application state
func (s *Settings) LoadFromState() {
	_, backendURL, _ := s.state.GetStatus()
	if backendURL != "" {
		s.UpdateSetting("backend_url", backendURL)
	}
}
