package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/claraverse/mcp-client/internal/tui"
)

// Activity represents the activity log view
type Activity struct {
	theme    *tui.Theme
	state    *tui.AppState
	viewport viewport.Model
	ready    bool
	filter   string
}

// NewActivity creates a new activity view
func NewActivity(state *tui.AppState) *Activity {
	return &Activity{
		theme: tui.DefaultTheme,
		state: state,
	}
}

// Init initializes the activity view
func (a *Activity) Init(width, height int) {
	a.viewport = viewport.New(width-4, height-6)
	a.viewport.Style = lipgloss.NewStyle()
	a.ready = true
}

// Update handles input for the activity view
func (a *Activity) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
			a.state.ClearActivities()
			return nil
		}
	}

	if a.ready {
		a.viewport, cmd = a.viewport.Update(msg)
	}

	return cmd
}

// View renders the activity view
func (a *Activity) View(width, height int) string {
	activities := a.state.GetActivities()

	if len(activities) == 0 {
		emptyMsg := a.theme.ValueMuted.Render("No activity recorded yet. Execute some MCP tools to see them here.")
		help := "\n" + a.theme.Help.Render("[c] clear")
		return lipgloss.NewStyle().Padding(1, 2).Render(emptyMsg + help)
	}

	var lines []string
	for _, act := range activities {
		// Skip if filter is set and doesn't match
		if a.filter != "" && !strings.Contains(strings.ToLower(act.ToolName), strings.ToLower(a.filter)) {
			continue
		}

		timeStr := act.Timestamp.Format("15:04:05")
		latencyStr := fmt.Sprintf("%6dms", act.Latency.Milliseconds())

		var statusIcon string
		if act.Success {
			statusIcon = a.theme.StatusSuccess.Render("✓")
		} else {
			statusIcon = a.theme.StatusError.Render("✗")
		}

		line := fmt.Sprintf("%s  %s  %-25s  %-30s  %s  %s",
			a.theme.ActivityTime.Render(timeStr),
			statusIcon,
			a.theme.ActivityTool.Render(truncate(act.ToolName, 25)),
			a.theme.ActivityArgs.Render(truncate(act.Arguments, 30)),
			a.theme.ActivityLatency.Render(latencyStr),
			a.theme.ValueMuted.Render(act.ServerName),
		)

		if !act.Success && act.Error != "" {
			line += "\n" + a.theme.StatusError.Render("    └─ "+truncate(act.Error, 60))
		}

		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")

	// Update viewport content
	if a.ready {
		a.viewport.SetContent(content)
	}

	help := "\n" + a.theme.Help.Render("[↑↓] scroll  [/] search  [f] filter  [c] clear")

	return lipgloss.NewStyle().Padding(1, 2).Render(content + help)
}

// SetFilter sets the activity filter
func (a *Activity) SetFilter(filter string) {
	a.filter = filter
}

// ClearFilter clears the activity filter
func (a *Activity) ClearFilter() {
	a.filter = ""
}
