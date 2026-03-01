package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/claraverse/mcp-client/internal/config"
	"github.com/claraverse/mcp-client/internal/coretools"
)

// WizardDoneMsg is sent when the setup wizard completes or is cancelled.
type WizardDoneMsg struct {
	Result WizardResult
}

// WizardResult holds the output of the setup wizard.
type WizardResult struct {
	BrowserConfig config.BrowserConfig
	Completed     bool // false if user cancelled
}

// wizardProfileItem is a selectable item in the profile list.
type wizardProfileItem struct {
	Name string // Display name
	Path string // Profile directory path
}

// WizardModel holds state for the first-run setup wizard.
type WizardModel struct {
	chromeBinary string // Empty if Chrome not found
	profiles     []wizardProfileItem
	cursor       int
	width        int
	height       int
}

// NewWizardModel creates a new setup wizard model.
// It detects Chrome and scans for available profiles.
func NewWizardModel() *WizardModel {
	w := &WizardModel{
		chromeBinary: coretools.FindChromeBinary(),
	}

	// Only populate profiles if Chrome is available.
	if w.chromeBinary != "" {
		// First option: isolated profile (recommended)
		w.profiles = append(w.profiles, wizardProfileItem{
			Name: "New isolated profile (recommended)",
			Path: config.DefaultIsolatedProfilePath(),
		})

		// Existing Chrome profiles
		for _, p := range config.ListChromeProfiles() {
			w.profiles = append(w.profiles, wizardProfileItem{
				Name: p.Name,
				Path: p.Path,
			})
		}
	}

	return w
}

// SetSize updates the wizard's known terminal dimensions.
func (w *WizardModel) SetSize(width, height int) {
	w.width = width
	w.height = height
}

// HandleKey processes a key press and returns an optional tea.Cmd.
// Returns (handled, cmd). handled is true if the wizard consumed the key.
func (w *WizardModel) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	k := msg.String()

	// Chrome not installed — only allow quit keys
	if w.chromeBinary == "" {
		switch k {
		case "q", "esc", "ctrl+c":
			return true, tea.Quit
		}
		return true, nil
	}

	// Chrome is installed — profile picker navigation
	switch k {
	case "up", "k":
		if w.cursor > 0 {
			w.cursor--
		}
		return true, nil

	case "down", "j":
		if w.cursor < len(w.profiles)-1 {
			w.cursor++
		}
		return true, nil

	case "enter":
		selected := w.profiles[w.cursor]
		return true, func() tea.Msg {
			return WizardDoneMsg{
				Result: WizardResult{
					BrowserConfig: config.BrowserConfig{
						ProfilePath: selected.Path,
						Port:        9222,
						AutoLaunch:  true,
					},
					Completed: true,
				},
			}
		}

	case "esc", "q":
		return true, func() tea.Msg {
			return WizardDoneMsg{
				Result: WizardResult{Completed: false},
			}
		}
	}

	return true, nil
}

// View renders the wizard screen.
func (w *WizardModel) View() string {
	width := w.width
	if width < 40 {
		width = 40
	}
	height := w.height
	if height < 10 {
		height = 10
	}

	styles := DefaultBootStyles()
	theme := DefaultTheme

	var b strings.Builder

	// ── Title ──
	titleStyle := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true)
	subtitleStyle := lipgloss.NewStyle().
		Foreground(ColorTextMuted).
		Italic(true)

	// Small logo
	for _, line := range strings.Split(claraverseLogoSmall, "\n") {
		if line == "" {
			continue
		}
		b.WriteString(centerText(styles.LogoPrimary.Render(line), width))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(centerText(titleStyle.Render("Clara Companion Setup"), width))
	b.WriteString("\n")
	b.WriteString(centerText(subtitleStyle.Render("First-time configuration"), width))
	b.WriteString("\n\n")

	// ── Chrome detection ──
	if w.chromeBinary == "" {
		// Chrome not found — blocking message
		box := w.renderChromeNotFound(theme, width)
		b.WriteString(box)
	} else {
		// Chrome found — show status + profile picker
		chromeInfo := theme.StatusSuccess.Render("  Chrome detected") +
			"  " + theme.ValueMuted.Render(w.chromeBinary)
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(chromeInfo))
		b.WriteString("\n\n")

		// Profile picker
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(
			theme.Title.Render("Select Chrome Profile"),
		))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(
			theme.ValueMuted.Render("Choose which Chrome profile Clara Companion should use for browsing:"),
		))
		b.WriteString("\n\n")

		for i, profile := range w.profiles {
			cursor := "  "
			style := theme.ListItem
			if i == w.cursor {
				cursor = theme.ListCursor.Render("> ")
				style = theme.ListItemActive
			}

			name := profile.Name
			// Highlight the recommended option
			if i == 0 {
				name = theme.Value.Bold(true).Render(name)
			}

			line := fmt.Sprintf("%s%s", cursor, name)
			b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(style.Render(line)))
			b.WriteString("\n")

			// Show path under selected item
			if i == w.cursor {
				pathLine := "    " + theme.ValueMuted.Render(profile.Path)
				b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(pathLine))
				b.WriteString("\n")
			}
		}

		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(
			theme.Help.Render("[↑↓/jk] navigate  [Enter] confirm  [Esc] cancel"),
		))
	}

	// Pad to fill screen
	lines := strings.Count(b.String(), "\n")
	for lines < height-1 {
		b.WriteString("\n")
		lines++
	}

	return b.String()
}

// renderChromeNotFound renders the "Chrome required" blocking message.
func (w *WizardModel) renderChromeNotFound(theme *Theme, width int) string {
	var content strings.Builder

	content.WriteString(theme.StatusError.Render("  Chrome Not Found"))
	content.WriteString("\n\n")
	content.WriteString(theme.Value.Render("  Google Chrome is required for Clara Companion's browser tools."))
	content.WriteString("\n\n")
	content.WriteString(theme.Label.Render("  To install Chrome:"))
	content.WriteString("\n")
	content.WriteString(theme.ValueMuted.Render("  Linux:  "))
	content.WriteString(theme.Value.Render("https://google.com/chrome"))
	content.WriteString("\n")
	content.WriteString(theme.ValueMuted.Render("  macOS:  "))
	content.WriteString(theme.Value.Render("brew install --cask google-chrome"))
	content.WriteString("\n\n")
	content.WriteString(theme.Help.Render("  Install Chrome and re-run Clara Companion."))
	content.WriteString("\n\n")
	content.WriteString(theme.Help.Render("  [q/Esc] quit"))

	box := theme.Card.Width(width - 8).Render(content.String())
	return lipgloss.NewStyle().Padding(0, 2).Render(box)
}
