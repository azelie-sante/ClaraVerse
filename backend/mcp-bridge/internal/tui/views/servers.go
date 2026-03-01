package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/claraverse/mcp-client/internal/tui"
)

// Servers represents the servers management view
type Servers struct {
	theme    *tui.Theme
	state    *tui.AppState
	cursor   int
	selected map[int]bool
}

// NewServers creates a new servers view
func NewServers(state *tui.AppState) *Servers {
	return &Servers{
		theme:    tui.DefaultTheme,
		state:    state,
		selected: make(map[int]bool),
	}
}

// Update handles input for the servers view
func (s *Servers) Update(msg tea.Msg) tea.Cmd {
	servers := s.state.GetServers()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if s.cursor > 0 {
				s.cursor--
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if s.cursor < len(servers)-1 {
				s.cursor++
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("home", "g"))):
			s.cursor = 0
		case key.Matches(msg, key.NewBinding(key.WithKeys("end", "G"))):
			if len(servers) > 0 {
				s.cursor = len(servers) - 1
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys(" ", "x"))):
			s.selected[s.cursor] = !s.selected[s.cursor]
		}
	}

	return nil
}

// View renders the servers view
func (s *Servers) View(width, height int) string {
	servers := s.state.GetServers()

	// Count connected
	connected := 0
	totalTools := 0
	for _, srv := range servers {
		if srv.Connected {
			connected++
			totalTools += srv.ToolCount
		}
	}

	var b strings.Builder

	// Header with summary
	header := s.theme.Title.Render("MCP Servers")
	summary := s.theme.ValueMuted.Render(fmt.Sprintf("%d connected · %d tools", connected, totalTools))
	gap := width - lipgloss.Width(header) - lipgloss.Width(summary) - 8
	if gap < 2 {
		gap = 2
	}
	b.WriteString(lipgloss.NewStyle().Padding(1, 2).Render(header + strings.Repeat(" ", gap) + summary))
	b.WriteString("\n")

	cardWidth := width - 6
	if cardWidth < 40 {
		cardWidth = 40
	}

	// Server list card
	if len(servers) == 0 {
		emptyBox := s.theme.Card.Width(cardWidth).Render(
			s.theme.ValueMuted.Render("No servers configured.\n\nPress [a] to add a server."),
		)
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(emptyBox))
	} else {
		var rows []string
		for i, srv := range servers {
			isSelected := i == s.cursor

			cursor := "  "
			if isSelected {
				cursor = s.theme.ListCursor.Render("> ")
			}

			// Status dot
			var status string
			if !srv.Enabled {
				status = s.theme.ValueMuted.Render("○")
			} else if srv.Connected {
				status = s.theme.StatusSuccess.Render("●")
			} else if srv.Error != "" {
				status = s.theme.StatusError.Render("●")
			} else {
				status = s.theme.StatusWarning.Render("●")
			}

			// Tool count
			toolStr := ""
			if srv.Connected {
				toolStr = s.theme.ValueMuted.Render(fmt.Sprintf("%d tools", srv.ToolCount))
			} else if !srv.Enabled {
				toolStr = s.theme.ValueMuted.Render("(disabled)")
			} else if srv.Error != "" {
				toolStr = s.theme.StatusError.Render("error")
			}

			// Description
			desc := truncate(srv.Description, 30)

			row := fmt.Sprintf("%s%s %-18s %s  %s",
				cursor, status,
				s.theme.Value.Render(truncate(srv.Name, 18)),
				s.theme.ValueMuted.Render(desc),
				toolStr,
			)

			if isSelected {
				row = s.theme.ListItemActive.Render(row)
			}
			rows = append(rows, row)
		}

		serverCard := s.theme.Card.Width(cardWidth).Render(strings.Join(rows, "\n"))
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(serverCard))
	}
	b.WriteString("\n\n")

	// Recent Tool Calls sub-panel
	b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(s.renderActivityTable(cardWidth)))

	b.WriteString("\n\n")
	help := s.theme.Help.Render("[↑↓] navigate  [a] add  [d] remove  [e] enable/disable  [c] clear log")
	b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(help))

	return b.String()
}

func (s *Servers) renderActivityTable(width int) string {
	activities := s.state.GetRecentActivities(8)

	header := s.theme.Label.Render("Recent Tool Calls")
	if len(activities) == 0 {
		return header + "\n\n" + s.theme.ValueMuted.Render("  No recent tool calls.")
	}

	var lines []string
	lines = append(lines, header)
	lines = append(lines, "")

	// Table header
	tableHdr := fmt.Sprintf("  %-10s %-20s %-14s %-8s %s",
		s.theme.Label.Render("TIME"),
		s.theme.Label.Render("TOOL"),
		s.theme.Label.Render("SERVER"),
		s.theme.Label.Render("LATENCY"),
		s.theme.Label.Render("STATUS"),
	)
	lines = append(lines, tableHdr)

	for _, act := range activities {
		timeStr := act.Timestamp.Format("15:04:05")
		latencyStr := fmt.Sprintf("%dms", act.Latency.Milliseconds())

		var statusIcon string
		if act.Success {
			statusIcon = s.theme.StatusSuccess.Render("✓")
		} else {
			statusIcon = s.theme.StatusError.Render("✗")
		}

		errInfo := ""
		if !act.Success && act.Error != "" {
			errInfo = " " + s.theme.StatusError.Render(truncate(act.Error, 20))
		}

		row := fmt.Sprintf("  %-10s %-20s %-14s %8s %s%s",
			s.theme.ValueMuted.Render(timeStr),
			s.theme.Value.Render(truncate(act.ToolName, 20)),
			s.theme.ValueMuted.Render(truncate(act.ServerName, 14)),
			s.theme.ValueMuted.Render(latencyStr),
			statusIcon,
			errInfo,
		)
		lines = append(lines, row)
	}

	return strings.Join(lines, "\n")
}

// GetSelectedServer returns the currently selected server
func (s *Servers) GetSelectedServer() *tui.ServerInfo {
	servers := s.state.GetServers()
	if s.cursor >= 0 && s.cursor < len(servers) {
		srv := servers[s.cursor]
		return &srv
	}
	return nil
}

// GetCursor returns the current cursor position
func (s *Servers) GetCursor() int {
	return s.cursor
}
