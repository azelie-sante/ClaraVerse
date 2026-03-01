package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/claraverse/mcp-client/internal/tui"
)

// Devices represents the devices management view
type Devices struct {
	theme  *tui.Theme
	state  *tui.AppState
	cursor int
}

// NewDevices creates a new devices view
func NewDevices(state *tui.AppState) *Devices {
	return &Devices{
		theme: tui.DefaultTheme,
		state: state,
	}
}

// Update handles input for the devices view
func (d *Devices) Update(msg tea.Msg) tea.Cmd {
	devices := d.state.GetDevices()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if d.cursor > 0 {
				d.cursor--
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if d.cursor < len(devices)-1 {
				d.cursor++
			}
		}
	}

	return nil
}

// View renders the devices view
func (d *Devices) View(width, height int) string {
	devices := d.state.GetDevices()

	if len(devices) == 0 {
		emptyMsg := d.theme.ValueMuted.Render("No devices found. This device will appear after authentication.")
		return lipgloss.NewStyle().Padding(1, 2).Render(emptyMsg)
	}

	var cards []string
	for i, dev := range devices {
		isSelected := i == d.cursor

		// Device name with current indicator
		name := dev.Name
		if name == "" {
			name = "Unnamed Device"
		}

		var nameStyle lipgloss.Style
		if isSelected {
			nameStyle = d.theme.Value.Bold(true).Foreground(tui.ColorAccent)
		} else {
			nameStyle = d.theme.Value
		}

		nameLine := nameStyle.Render(name)
		if dev.Current {
			nameLine += d.theme.StatusSuccess.Render(" (current)")
		}

		// Device ID (truncated)
		idLine := d.theme.Label.Render("ID: ") + d.theme.ValueMuted.Render(truncateID(dev.ID))

		// Email
		emailLine := d.theme.Label.Render("Email: ") + d.theme.ValueMuted.Render(dev.Email)

		// Last active
		lastActive := "Never"
		if !dev.LastActive.IsZero() {
			lastActive = dev.LastActive.Format("2006-01-02 15:04")
		}
		activeLine := d.theme.Label.Render("Last active: ") + d.theme.ValueMuted.Render(lastActive)

		// Cursor indicator
		cursor := "  "
		if isSelected {
			cursor = d.theme.ListCursor.Render("> ")
		}

		card := fmt.Sprintf("%s%s\n  %s\n  %s\n  %s",
			cursor, nameLine, idLine, emailLine, activeLine)
		cards = append(cards, card)
	}

	content := strings.Join(cards, "\n\n")

	help := "\n\n" + d.theme.Help.Render("[↑↓] navigate  [d] revoke  [r] rename")

	return lipgloss.NewStyle().Padding(1, 2).Render(content + help)
}

// GetSelectedDevice returns the currently selected device
func (d *Devices) GetSelectedDevice() *tui.DeviceInfo {
	devices := d.state.GetDevices()
	if d.cursor >= 0 && d.cursor < len(devices) {
		dev := devices[d.cursor]
		return &dev
	}
	return nil
}

func truncateID(id string) string {
	if len(id) > 12 {
		return id[:8] + "..."
	}
	return id
}
