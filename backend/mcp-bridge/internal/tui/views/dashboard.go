package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/claraverse/mcp-client/internal/tui"
)

// Dashboard represents the dashboard view
type Dashboard struct {
	theme *tui.Theme
	state *tui.AppState
}

// NewDashboard creates a new dashboard view
func NewDashboard(state *tui.AppState) *Dashboard {
	return &Dashboard{
		theme: tui.DefaultTheme,
		state: state,
	}
}

// View renders the dashboard
func (d *Dashboard) View(width, height int) string {
	halfW := width/2 - 3
	if halfW < 20 {
		halfW = 20
	}

	// Row 1: Connection
	connPanel := d.renderConnectionPanel(width - 4)

	// Row 2: Servers + Recent Activity
	srvPanel := d.renderServersPanel(halfW)
	actPanel := d.renderActivityPanel(halfW)
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, srvPanel, "  ", actPanel)

	// Stats line
	summary := d.state.Summary()
	statsLine := d.theme.ValueMuted.Render(fmt.Sprintf("%d servers · %d tools · %d calls",
		summary.ConnectedCount, summary.TotalTools, summary.ActivityCount))

	content := lipgloss.JoinVertical(lipgloss.Left,
		"",
		connPanel,
		"",
		bottomRow,
		"",
		statsLine,
	)

	return lipgloss.NewStyle().Padding(1, 2).Render(content)
}

func (d *Dashboard) renderConnectionPanel(width int) string {
	status, backendURL, _ := d.state.GetStatus()
	_, _, _, tokenExp := d.state.GetUser()

	var statusLine string
	if status == tui.StatusConnected {
		statusLine = d.theme.StatusSuccess.Render("● Connected")
	} else {
		statusLine = d.theme.StatusError.Render("● " + status.String())
	}

	expiresIn := time.Until(tokenExp)
	var tokenLine string
	if expiresIn > 0 {
		if expiresIn > time.Hour {
			tokenLine = fmt.Sprintf("%dh %dm left", int(expiresIn.Hours()), int(expiresIn.Minutes())%60)
		} else {
			tokenLine = fmt.Sprintf("%dm left", int(expiresIn.Minutes()))
		}
		if expiresIn < 10*time.Minute {
			tokenLine = d.theme.StatusWarning.Render(tokenLine)
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		statusLine,
		d.theme.Label.Render("Backend  ")+d.theme.Value.Render(truncate(backendURL, width-12)),
		d.theme.Label.Render("Token    ")+d.theme.Value.Render(tokenLine),
	)

	return d.theme.Card.Width(width).Render(
		d.theme.Title.Render("Connection") + "\n" + content,
	)
}

func (d *Dashboard) renderServersPanel(width int) string {
	servers := d.state.GetServers()

	var lines []string
	connected, totalTools := 0, 0
	for _, srv := range servers {
		if !srv.Enabled {
			continue
		}

		if srv.Connected {
			connected++
			totalTools += srv.ToolCount
			status := d.theme.StatusSuccess.Render("●")
			lines = append(lines, fmt.Sprintf("%s %-14s %s",
				status,
				d.theme.Value.Render(truncate(srv.Name, 14)),
				d.theme.ValueMuted.Render(fmt.Sprintf("%dt", srv.ToolCount)),
			))
		} else if srv.Error != "" {
			status := d.theme.StatusError.Render("●")
			lines = append(lines, fmt.Sprintf("%s %-14s %s",
				status,
				d.theme.Value.Render(truncate(srv.Name, 14)),
				d.theme.StatusError.Render("error"),
			))
		} else {
			status := d.theme.ValueMuted.Render("○")
			lines = append(lines, fmt.Sprintf("%s %s",
				status, d.theme.ValueMuted.Render(srv.Name),
			))
		}
	}

	if len(lines) == 0 {
		lines = append(lines, d.theme.ValueMuted.Render("No servers"))
	}

	// Footer line
	lines = append(lines, d.theme.ValueMuted.Render(fmt.Sprintf("  %d connected · %d tools", connected, totalTools)))

	content := strings.Join(lines, "\n")
	return d.theme.Card.Width(width).Render(
		d.theme.Title.Render("MCP Servers") + "\n" + content,
	)
}

func (d *Dashboard) renderActivityPanel(width int) string {
	activities := d.state.GetRecentActivities(4)

	if len(activities) == 0 {
		return d.theme.Card.Width(width).Render(
			d.theme.Title.Render("Recent Activity") + "\n" +
				d.theme.ValueMuted.Render("No recent activity"),
		)
	}

	var lines []string
	for _, act := range activities {
		timeStr := act.Timestamp.Format("15:04")

		var statusIcon string
		if act.Success {
			statusIcon = d.theme.StatusSuccess.Render("✓")
		} else {
			statusIcon = d.theme.StatusError.Render("✗")
		}

		line := fmt.Sprintf("%s %s %s",
			d.theme.ValueMuted.Render(timeStr),
			d.theme.Value.Render(truncate(act.ToolName, 18)),
			statusIcon,
		)
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	return d.theme.Card.Width(width).Render(
		d.theme.Title.Render("Recent Activity") + "\n" + content,
	)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
