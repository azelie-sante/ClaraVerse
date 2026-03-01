package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/claraverse/mcp-client/internal/tui"
)

// ModalType defines the type of modal
type ModalType int

const (
	ModalConfirmation ModalType = iota
	ModalInput
	ModalInfo
	ModalError
)

// Modal represents a modal dialog component
type Modal struct {
	theme     *tui.Theme
	visible   bool
	modalType ModalType
	title     string
	message   string
	payload   interface{}

	// For input modals
	textInput textinput.Model
	inputLabel string

	// Button labels
	confirmLabel string
	cancelLabel  string

	// Focus state
	focusConfirm bool
}

// NewModal creates a new modal component
func NewModal() *Modal {
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.CharLimit = 256
	ti.Width = 40

	return &Modal{
		theme:        tui.DefaultTheme,
		textInput:    ti,
		confirmLabel: "Confirm",
		cancelLabel:  "Cancel",
		focusConfirm: true,
	}
}

// Show displays a confirmation modal
func (m *Modal) Show(title, message string, payload interface{}) {
	m.visible = true
	m.modalType = ModalConfirmation
	m.title = title
	m.message = message
	m.payload = payload
	m.focusConfirm = true
}

// ShowInput displays an input modal
func (m *Modal) ShowInput(title, label, placeholder string, payload interface{}) {
	m.visible = true
	m.modalType = ModalInput
	m.title = title
	m.inputLabel = label
	m.payload = payload
	m.textInput.Placeholder = placeholder
	m.textInput.SetValue("")
	m.textInput.Focus()
}

// ShowInfo displays an info modal
func (m *Modal) ShowInfo(title, message string) {
	m.visible = true
	m.modalType = ModalInfo
	m.title = title
	m.message = message
}

// ShowError displays an error modal
func (m *Modal) ShowError(title, message string) {
	m.visible = true
	m.modalType = ModalError
	m.title = title
	m.message = message
}

// Hide hides the modal
func (m *Modal) Hide() {
	m.visible = false
	m.textInput.Blur()
}

// IsVisible returns whether the modal is visible
func (m *Modal) IsVisible() bool {
	return m.visible
}

// SetButtonLabels sets custom button labels
func (m *Modal) SetButtonLabels(confirm, cancel string) {
	m.confirmLabel = confirm
	m.cancelLabel = cancel
}

// GetPayload returns the modal payload
func (m *Modal) GetPayload() interface{} {
	return m.payload
}

// GetInputValue returns the text input value
func (m *Modal) GetInputValue() string {
	return m.textInput.Value()
}

// Update handles modal input
func (m *Modal) Update(msg tea.Msg) (tea.Cmd, *tui.ModalResultMsg) {
	if !m.visible {
		return nil, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			m.Hide()
			return nil, &tui.ModalResultMsg{Action: tui.ModalCancel}

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if m.modalType == ModalInput || m.focusConfirm {
				m.Hide()
				result := &tui.ModalResultMsg{
					Action:  tui.ModalConfirm,
					Payload: m.payload,
				}
				if m.modalType == ModalInput {
					result.Payload = m.textInput.Value()
				}
				return nil, result
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("y"))):
			if m.modalType == ModalConfirmation {
				m.Hide()
				return nil, &tui.ModalResultMsg{Action: tui.ModalConfirm, Payload: m.payload}
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("n"))):
			if m.modalType == ModalConfirmation {
				m.Hide()
				return nil, &tui.ModalResultMsg{Action: tui.ModalCancel}
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			if m.modalType == ModalConfirmation {
				m.focusConfirm = !m.focusConfirm
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("left", "right"))):
			if m.modalType == ModalConfirmation {
				m.focusConfirm = !m.focusConfirm
			}
		}

		// Update text input for input modals
		if m.modalType == ModalInput {
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return cmd, nil
		}
	}

	return nil, nil
}

// View renders the modal
func (m *Modal) View(width, height int) string {
	if !m.visible {
		return ""
	}

	var content string

	switch m.modalType {
	case ModalConfirmation:
		content = m.renderConfirmation()
	case ModalInput:
		content = m.renderInput()
	case ModalInfo, ModalError:
		content = m.renderInfo()
	}

	// Style the modal container
	borderColor := tui.ColorAccent
	if m.modalType == ModalError {
		borderColor = tui.ColorError
	}

	modal := lipgloss.NewStyle().
		Background(tui.ColorSurface).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Width(50).
		Render(content)

	return modal
}

// ViewOverlay renders the modal as an overlay on base content
func (m *Modal) ViewOverlay(base string, width, height int) string {
	if !m.visible {
		return base
	}

	modal := m.View(width, height)

	// Center the modal
	modalWidth := lipgloss.Width(modal)
	modalHeight := lipgloss.Height(modal)
	x := (width - modalWidth) / 2
	y := (height - modalHeight) / 2

	return placeOverlay(x, y, modal, base)
}

func (m *Modal) renderConfirmation() string {
	title := m.theme.ModalTitle.Render(m.title)
	message := m.theme.ModalContent.Render(m.message)

	// Buttons
	var confirmBtn, cancelBtn string
	if m.focusConfirm {
		confirmBtn = m.theme.ButtonPrimary.Render(" " + m.confirmLabel + " [y] ")
		cancelBtn = m.theme.ButtonSecondary.Render(" " + m.cancelLabel + " [n] ")
	} else {
		confirmBtn = m.theme.ButtonSecondary.Render(" " + m.confirmLabel + " [y] ")
		cancelBtn = m.theme.ButtonPrimary.Render(" " + m.cancelLabel + " [n] ")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, "  ", confirmBtn)
	buttonRow := m.theme.ModalButtonRow.Render(buttons)

	return lipgloss.JoinVertical(lipgloss.Left, title, "", message, "", buttonRow)
}

func (m *Modal) renderInput() string {
	title := m.theme.ModalTitle.Render(m.title)
	label := m.theme.InputLabel.Render(m.inputLabel)

	inputStyle := m.theme.InputFocus
	input := inputStyle.Render(m.textInput.View())

	hint := m.theme.Help.Render("Press Enter to confirm, Esc to cancel")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", label, input, "", hint)
}

func (m *Modal) renderInfo() string {
	var titleStyle lipgloss.Style
	if m.modalType == ModalError {
		titleStyle = m.theme.StatusError.Bold(true)
	} else {
		titleStyle = m.theme.ModalTitle
	}

	title := titleStyle.Render(m.title)
	message := m.theme.ModalContent.Render(m.message)

	okBtn := m.theme.ButtonPrimary.Render(" OK [Enter] ")
	buttonRow := m.theme.ModalButtonRow.Render(okBtn)

	return lipgloss.JoinVertical(lipgloss.Left, title, "", message, "", buttonRow)
}

// placeOverlay places an overlay on top of base content at the specified position
func placeOverlay(x, y int, overlay, base string) string {
	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")

	for i, line := range overlayLines {
		row := y + i
		if row < 0 || row >= len(baseLines) {
			continue
		}

		baseLine := baseLines[row]
		baseRunes := []rune(baseLine)

		// Pad base line if needed
		for len(baseRunes) < x {
			baseRunes = append(baseRunes, ' ')
		}

		// Insert overlay
		overlayRunes := []rune(line)
		endX := x + len(overlayRunes)

		if x < len(baseRunes) {
			if endX > len(baseRunes) {
				baseRunes = append(baseRunes[:x], overlayRunes...)
			} else {
				copy(baseRunes[x:], overlayRunes)
			}
		}

		baseLines[row] = string(baseRunes)
	}

	return strings.Join(baseLines, "\n")
}
