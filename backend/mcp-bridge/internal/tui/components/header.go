package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/claraverse/mcp-client/internal/tui"
)

// Header represents the top header bar component
type Header struct {
	theme   *tui.Theme
	spinner spinner.Model
}

// NewHeader creates a new header component
func NewHeader() *Header {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = tui.DefaultTheme.Spinner

	return &Header{
		theme:   tui.DefaultTheme,
		spinner: s,
	}
}

// View renders the header
func (h *Header) View(width int, status tui.ConnectionStatus, email string, spinnerView string) string {
	// Logo
	dot := h.theme.LogoDot.Render("◉")
	logo := h.theme.Logo.Render(" ClaraVerse MCP")

	// Status indicator
	var statusStr string
	switch status {
	case tui.StatusConnected:
		statusStr = h.theme.StatusSuccess.Render("●") + " Connected"
	case tui.StatusConnecting, tui.StatusReconnecting:
		statusStr = spinnerView + " " + status.String()
	default:
		statusStr = h.theme.StatusError.Render("●") + " " + status.String()
	}

	// User email
	userStr := ""
	if email != "" {
		userStr = h.theme.UserEmail.Render(email)
	}

	// Build header
	leftSide := dot + logo
	rightSide := statusStr + "  " + userStr

	padding := width - lipgloss.Width(leftSide) - lipgloss.Width(rightSide) - 4
	if padding < 0 {
		padding = 0
	}

	header := leftSide + strings.Repeat(" ", padding) + rightSide

	return h.theme.HeaderContainer.Width(width).Render(header)
}

// TabBar represents the tab navigation component
type TabBar struct {
	theme     *tui.Theme
	activeTab tui.TabIndex
}

// NewTabBar creates a new tab bar component
func NewTabBar() *TabBar {
	return &TabBar{
		theme: tui.DefaultTheme,
	}
}

// SetActive sets the active tab
func (t *TabBar) SetActive(tab tui.TabIndex) {
	t.activeTab = tab
}

// View renders the tab bar
func (t *TabBar) View(width int) string {
	tabs := tui.TabNames()
	var rendered []string

	for i, tab := range tabs {
		num := string(rune('1' + i))

		if tui.TabIndex(i) == t.activeTab {
			label := " " + num + " " + tab + " "
			rendered = append(rendered, t.theme.TabActive.Render(label))
		} else {
			numStyled := t.theme.TabNumberMuted.Render(num)
			label := " " + numStyled + " " + tab + " "
			rendered = append(rendered, t.theme.TabInactive.Render(label))
		}
	}

	tabRow := lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
	return t.theme.TabContainer.Width(width).Render(tabRow)
}

// GetTabFromClick returns the tab index from a mouse click position
func (t *TabBar) GetTabFromClick(x, y int) tui.TabIndex {
	// Tabs are on row 0 of the tab bar
	if y != 0 {
		return -1
	}

	// Calculate tab positions
	tabs := tui.TabNames()
	pos := 2 // Starting position after padding
	for i, tab := range tabs {
		tabWidth := len(tab) + 4 // Account for padding and number
		if x >= pos && x < pos+tabWidth {
			return tui.TabIndex(i)
		}
		pos += tabWidth + 1 // +1 for margin
	}
	return -1
}
