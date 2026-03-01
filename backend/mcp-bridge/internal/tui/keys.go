package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines all keyboard shortcuts
type KeyMap struct {
	// Navigation
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding

	// Tabs
	Tab1    key.Binding
	Tab2    key.Binding
	Tab3    key.Binding
	NextTab key.Binding
	PrevTab key.Binding

	// Actions
	Enter   key.Binding
	Select  key.Binding
	Back    key.Binding
	Cancel  key.Binding
	Refresh key.Binding
	Help    key.Binding
	Quit    key.Binding

	// Server actions
	Add     key.Binding
	Remove  key.Binding
	Enable  key.Binding
	Disable key.Binding
	Edit    key.Binding

	// Modal actions
	Confirm key.Binding
	Escape  key.Binding

	// Search/Filter
	Search key.Binding
	Filter key.Binding
	Clear  key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Navigation
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("pgdn", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("home/g", "go to top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("end/G", "go to bottom"),
		),

		// Tabs (1-3 for direct access)
		Tab1: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "dashboard"),
		),
		Tab2: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "servers"),
		),
		Tab3: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "settings"),
		),
		NextTab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next tab"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev tab"),
		),

		// Actions
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Select: key.NewBinding(
			key.WithKeys(" ", "x"),
			key.WithHelp("space", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("backspace", "b"),
			key.WithHelp("backspace", "back"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),

		// Server actions
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add"),
		),
		Remove: key.NewBinding(
			key.WithKeys("d", "delete"),
			key.WithHelp("d", "remove"),
		),
		Enable: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "enable"),
		),
		Disable: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "disable"),
		),
		Edit: key.NewBinding(
			key.WithKeys("E"),
			key.WithHelp("E", "edit"),
		),

		// Modal actions
		Confirm: key.NewBinding(
			key.WithKeys("y", "enter"),
			key.WithHelp("y/enter", "confirm"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc", "n"),
			key.WithHelp("esc/n", "cancel"),
		),

		// Search/Filter
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter"),
		),
		Clear: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "clear"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab1, k.Tab2, k.Tab3, k.Refresh, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},     // Navigation
		{k.Tab1, k.Tab2, k.Tab3},             // Tabs
		{k.Add, k.Remove, k.Enable, k.Disable}, // Server actions
		{k.Search, k.Filter, k.Clear},        // Search
		{k.Refresh, k.Help, k.Quit},          // General
	}
}

// TabHelpText returns help text for the current tab
func TabHelpText(tabIndex int, keys KeyMap) string {
	base := "[1-3] tabs  "

	switch tabIndex {
	case 0: // Dashboard
		return base + RenderKeyHelp("r", "refresh") + "  " + RenderKeyHelp("q", "quit")
	case 1: // Servers
		return base + RenderKeyHelp("a", "add") + "  " + RenderKeyHelp("d", "remove") + "  " + RenderKeyHelp("e", "toggle") + "  " + RenderKeyHelp("q", "quit")
	case 2: // Settings
		return base + RenderKeyHelp("enter", "edit") + "  " + RenderKeyHelp("q", "quit")
	default:
		return base + RenderKeyHelp("q", "quit")
	}
}
