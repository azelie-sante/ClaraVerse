package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// ClaraVerse Design System Colors
var (
	// Core colors
	ColorBackground   = lipgloss.Color("#0f0f0f")
	ColorSurface      = lipgloss.Color("#161616")
	ColorSurfaceLight = lipgloss.Color("#1a1a1a")
	ColorBorder       = lipgloss.Color("#2a2a2a")
	ColorBorderLight  = lipgloss.Color("#3a3a3a")

	// Accent (Rose Pink)
	ColorAccent      = lipgloss.Color("#e91e63")
	ColorAccentHover = lipgloss.Color("#f06292")
	ColorAccentDim   = lipgloss.Color("#880e4f")

	// Semantic colors
	ColorSuccess = lipgloss.Color("#30d158")
	ColorWarning = lipgloss.Color("#ffd60a")
	ColorError   = lipgloss.Color("#ff453a")
	ColorInfo    = lipgloss.Color("#64d2ff")

	// Text colors
	ColorTextPrimary   = lipgloss.Color("#ffffff")
	ColorTextSecondary = lipgloss.Color("#d0d0d0")
	ColorTextMuted     = lipgloss.Color("#808080")
	ColorTextDim       = lipgloss.Color("#505050")
)

// Theme contains all styled components
type Theme struct {
	// Base styles
	App    lipgloss.Style
	Panel  lipgloss.Style
	Card   lipgloss.Style
	Border lipgloss.Style

	// Header styles
	HeaderContainer lipgloss.Style
	Logo            lipgloss.Style
	LogoDot         lipgloss.Style
	UserEmail       lipgloss.Style

	// Tab styles
	TabContainer   lipgloss.Style
	TabActive      lipgloss.Style
	TabInactive    lipgloss.Style
	TabNumber      lipgloss.Style
	TabNumberMuted lipgloss.Style

	// Footer styles
	FooterContainer lipgloss.Style
	FooterKey       lipgloss.Style
	FooterLabel     lipgloss.Style
	StatusDot       lipgloss.Style

	// Content styles
	Title         lipgloss.Style
	Subtitle      lipgloss.Style
	Label         lipgloss.Style
	Value         lipgloss.Style
	ValueMuted    lipgloss.Style
	StatusSuccess lipgloss.Style
	StatusError   lipgloss.Style
	StatusWarning lipgloss.Style
	StatusInfo    lipgloss.Style

	// Table styles
	TableHeader lipgloss.Style
	TableRow    lipgloss.Style
	TableRowAlt lipgloss.Style
	TableCell   lipgloss.Style

	// Button styles
	ButtonPrimary   lipgloss.Style
	ButtonSecondary lipgloss.Style
	ButtonDanger    lipgloss.Style
	ButtonDisabled  lipgloss.Style

	// Modal styles
	ModalOverlay    lipgloss.Style
	ModalContainer  lipgloss.Style
	ModalTitle      lipgloss.Style
	ModalContent    lipgloss.Style
	ModalButtonRow  lipgloss.Style

	// Input styles
	Input       lipgloss.Style
	InputFocus  lipgloss.Style
	InputLabel  lipgloss.Style
	InputError  lipgloss.Style
	Placeholder lipgloss.Style

	// List styles
	ListItem       lipgloss.Style
	ListItemActive lipgloss.Style
	ListCursor     lipgloss.Style

	// Activity log styles
	ActivityTime    lipgloss.Style
	ActivityTool    lipgloss.Style
	ActivityArgs    lipgloss.Style
	ActivitySuccess lipgloss.Style
	ActivityError   lipgloss.Style
	ActivityLatency lipgloss.Style

	// Card styles
	CardActive    lipgloss.Style
	BadgeSuccess  lipgloss.Style
	BadgeWarning  lipgloss.Style
	BadgeError    lipgloss.Style
	BadgeInfo     lipgloss.Style
	BadgeMuted    lipgloss.Style
	ProgressFill  lipgloss.Style
	ProgressEmpty lipgloss.Style
	CodeBlock     lipgloss.Style

	// Misc
	Divider  lipgloss.Style
	Help     lipgloss.Style
	HelpKey  lipgloss.Style
	Spinner  lipgloss.Style
	Progress lipgloss.Style
}

// NewTheme creates the ClaraVerse themed styles
func NewTheme() *Theme {
	t := &Theme{}

	// Base styles
	t.App = lipgloss.NewStyle().
		Background(ColorBackground).
		Foreground(ColorTextPrimary)

	t.Panel = lipgloss.NewStyle().
		Background(ColorSurface).
		Padding(1, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder)

	t.Card = lipgloss.NewStyle().
		Background(ColorSurface).
		Padding(1, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder)

	t.Border = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder)

	// Header styles
	t.HeaderContainer = lipgloss.NewStyle().
		Background(ColorSurface).
		Padding(0, 2).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(ColorBorder)

	t.Logo = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorTextPrimary)

	t.LogoDot = lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true)

	t.UserEmail = lipgloss.NewStyle().
		Foreground(ColorTextSecondary)

	// Tab styles
	t.TabContainer = lipgloss.NewStyle().
		Background(ColorSurface).
		Padding(0, 2)

	t.TabActive = lipgloss.NewStyle().
		Background(ColorAccent).
		Foreground(ColorTextPrimary).
		Bold(true).
		Padding(0, 2).
		MarginRight(1)

	t.TabInactive = lipgloss.NewStyle().
		Background(ColorSurfaceLight).
		Foreground(ColorTextSecondary).
		Padding(0, 2).
		MarginRight(1)

	t.TabNumber = lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true)

	t.TabNumberMuted = lipgloss.NewStyle().
		Foreground(ColorTextMuted)

	// Footer styles
	t.FooterContainer = lipgloss.NewStyle().
		Background(ColorSurface).
		Padding(0, 2).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(ColorBorder)

	t.FooterKey = lipgloss.NewStyle().
		Background(ColorSurfaceLight).
		Foreground(ColorAccent).
		Padding(0, 1).
		Bold(true)

	t.FooterLabel = lipgloss.NewStyle().
		Foreground(ColorTextSecondary).
		MarginRight(2)

	t.StatusDot = lipgloss.NewStyle().
		Foreground(ColorSuccess)

	// Content styles
	t.Title = lipgloss.NewStyle().
		Foreground(ColorTextPrimary).
		Bold(true).
		MarginBottom(1)

	t.Subtitle = lipgloss.NewStyle().
		Foreground(ColorTextSecondary).
		Italic(true)

	t.Label = lipgloss.NewStyle().
		Foreground(ColorTextSecondary)

	t.Value = lipgloss.NewStyle().
		Foreground(ColorTextPrimary)

	t.ValueMuted = lipgloss.NewStyle().
		Foreground(ColorTextMuted)

	t.StatusSuccess = lipgloss.NewStyle().
		Foreground(ColorSuccess)

	t.StatusError = lipgloss.NewStyle().
		Foreground(ColorError)

	t.StatusWarning = lipgloss.NewStyle().
		Foreground(ColorWarning)

	t.StatusInfo = lipgloss.NewStyle().
		Foreground(ColorInfo)

	// Table styles
	t.TableHeader = lipgloss.NewStyle().
		Foreground(ColorTextSecondary).
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(ColorBorder).
		Padding(0, 1)

	t.TableRow = lipgloss.NewStyle().
		Padding(0, 1)

	t.TableRowAlt = lipgloss.NewStyle().
		Background(ColorSurfaceLight).
		Padding(0, 1)

	t.TableCell = lipgloss.NewStyle().
		Padding(0, 1)

	// Button styles
	t.ButtonPrimary = lipgloss.NewStyle().
		Background(ColorAccent).
		Foreground(ColorTextPrimary).
		Padding(0, 2).
		Bold(true)

	t.ButtonSecondary = lipgloss.NewStyle().
		Background(ColorSurfaceLight).
		Foreground(ColorTextPrimary).
		Padding(0, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder)

	t.ButtonDanger = lipgloss.NewStyle().
		Background(ColorError).
		Foreground(ColorTextPrimary).
		Padding(0, 2).
		Bold(true)

	t.ButtonDisabled = lipgloss.NewStyle().
		Background(ColorSurfaceLight).
		Foreground(ColorTextMuted).
		Padding(0, 2)

	// Modal styles
	t.ModalOverlay = lipgloss.NewStyle().
		Background(lipgloss.Color("#000000"))

	t.ModalContainer = lipgloss.NewStyle().
		Background(ColorSurface).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent).
		Padding(1, 2).
		Width(50)

	t.ModalTitle = lipgloss.NewStyle().
		Foreground(ColorTextPrimary).
		Bold(true).
		MarginBottom(1)

	t.ModalContent = lipgloss.NewStyle().
		Foreground(ColorTextSecondary).
		MarginBottom(1)

	t.ModalButtonRow = lipgloss.NewStyle().
		MarginTop(1)

	// Input styles
	t.Input = lipgloss.NewStyle().
		Background(ColorSurfaceLight).
		Foreground(ColorTextPrimary).
		Padding(0, 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder)

	t.InputFocus = lipgloss.NewStyle().
		Background(ColorSurfaceLight).
		Foreground(ColorTextPrimary).
		Padding(0, 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent)

	t.InputLabel = lipgloss.NewStyle().
		Foreground(ColorTextSecondary).
		MarginBottom(1)

	t.InputError = lipgloss.NewStyle().
		Foreground(ColorError).
		Italic(true)

	t.Placeholder = lipgloss.NewStyle().
		Foreground(ColorTextMuted)

	// List styles
	t.ListItem = lipgloss.NewStyle().
		Foreground(ColorTextPrimary).
		Padding(0, 1)

	t.ListItemActive = lipgloss.NewStyle().
		Background(ColorAccentDim).
		Foreground(ColorTextPrimary).
		Padding(0, 1)

	t.ListCursor = lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true)

	// Activity log styles
	t.ActivityTime = lipgloss.NewStyle().
		Foreground(ColorTextMuted).
		Width(8)

	t.ActivityTool = lipgloss.NewStyle().
		Foreground(ColorInfo).
		Bold(true).
		Width(20)

	t.ActivityArgs = lipgloss.NewStyle().
		Foreground(ColorTextSecondary).
		Width(25)

	t.ActivitySuccess = lipgloss.NewStyle().
		Foreground(ColorSuccess)

	t.ActivityError = lipgloss.NewStyle().
		Foreground(ColorError)

	t.ActivityLatency = lipgloss.NewStyle().
		Foreground(ColorTextMuted).
		Align(lipgloss.Right).
		Width(8)

	// Card styles
	t.CardActive = lipgloss.NewStyle().
		Background(ColorSurfaceLight).
		Padding(0, 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent)

	t.BadgeSuccess = lipgloss.NewStyle().
		Background(ColorSuccess).
		Foreground(lipgloss.Color("#000000")).
		Padding(0, 1).
		Bold(true)

	t.BadgeWarning = lipgloss.NewStyle().
		Background(ColorWarning).
		Foreground(lipgloss.Color("#000000")).
		Padding(0, 1).
		Bold(true)

	t.BadgeError = lipgloss.NewStyle().
		Background(ColorError).
		Foreground(lipgloss.Color("#ffffff")).
		Padding(0, 1).
		Bold(true)

	t.BadgeInfo = lipgloss.NewStyle().
		Background(ColorInfo).
		Foreground(lipgloss.Color("#000000")).
		Padding(0, 1).
		Bold(true)

	t.BadgeMuted = lipgloss.NewStyle().
		Background(ColorSurfaceLight).
		Foreground(ColorTextMuted).
		Padding(0, 1)

	t.ProgressFill = lipgloss.NewStyle().
		Foreground(ColorAccent)

	t.ProgressEmpty = lipgloss.NewStyle().
		Foreground(ColorBorder)

	t.CodeBlock = lipgloss.NewStyle().
		Background(ColorSurfaceLight).
		Foreground(ColorTextSecondary).
		Padding(0, 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderForeground(ColorAccentDim)

	// Misc styles
	t.Divider = lipgloss.NewStyle().
		Foreground(ColorBorder)

	t.Help = lipgloss.NewStyle().
		Foreground(ColorTextMuted)

	t.HelpKey = lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true)

	t.Spinner = lipgloss.NewStyle().
		Foreground(ColorAccent)

	t.Progress = lipgloss.NewStyle().
		Foreground(ColorAccent)

	return t
}

// DefaultTheme is the global theme instance
var DefaultTheme = NewTheme()

// Helper functions for common styled elements

// StatusDot returns a colored status indicator dot
func StatusDot(connected bool) string {
	if connected {
		return DefaultTheme.StatusSuccess.Render("●")
	}
	return DefaultTheme.StatusError.Render("●")
}

// RenderKeyHelp renders a key binding hint
func RenderKeyHelp(key, label string) string {
	return DefaultTheme.HelpKey.Render(key) + " " + DefaultTheme.Help.Render(label)
}

// HorizontalLine creates a horizontal divider
func HorizontalLine(width int) string {
	line := ""
	for i := 0; i < width; i++ {
		line += "─"
	}
	return DefaultTheme.Divider.Render(line)
}
