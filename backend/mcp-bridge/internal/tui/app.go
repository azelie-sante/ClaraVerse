package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/claraverse/mcp-client/internal/updater"
)

// PopularServer represents a pre-configured MCP server
type PopularServer struct {
	Name        string
	Description string
	Command     string
	Args        []string
	Category    string
	Stars       string // Rating indicator
}

// PopularServers is the list of popular MCP servers (curated from mcp-awesome.com and official repos)
var PopularServers = []PopularServer{
	// Essential (Official Reference)
	{"filesystem", "Secure file operations with access controls", "npx", []string{"-y", "@modelcontextprotocol/server-filesystem", "."}, "Essential", "★"},
	{"git", "Read, search, and manipulate Git repos", "npx", []string{"-y", "@modelcontextprotocol/server-git"}, "Essential", "★"},
	{"memory", "Knowledge graph persistent memory", "npx", []string{"-y", "@modelcontextprotocol/server-memory"}, "Essential", "★"},
	{"fetch", "Web content fetching & conversion", "npx", []string{"-y", "@modelcontextprotocol/server-fetch"}, "Essential", "★"},
	{"time", "Time and timezone conversion", "npx", []string{"-y", "@modelcontextprotocol/server-time"}, "Essential", "★"},
	// Browser Automation
	{"playwright", "Cross-browser automation (Chromium/FF/WebKit)", "npx", []string{"-y", "@playwright/mcp@latest"}, "Browser", "★★★"},
	{"puppeteer", "Headless Chrome automation", "npx", []string{"-y", "@modelcontextprotocol/server-puppeteer"}, "Browser", "★★"},
	{"chrome-devtools", "Chrome DevTools integration", "npx", []string{"-y", "chrome-devtools-mcp@latest"}, "Browser", "★★"},
	{"browserbase", "Cloud browser automation", "npx", []string{"-y", "@browserbasehq/mcp-server-browserbase"}, "Browser", "★★"},
	// Database
	{"postgres", "PostgreSQL query & schema operations", "npx", []string{"-y", "@modelcontextprotocol/server-postgres"}, "Database", "★★★"},
	{"sqlite", "SQLite database operations", "npx", []string{"-y", "@modelcontextprotocol/server-sqlite"}, "Database", "★★"},
	{"supabase", "Supabase PostgreSQL, auth & storage", "npx", []string{"-y", "supabase-mcp-server"}, "Database", "★★"},
	{"mongodb", "MongoDB & Atlas integration", "npx", []string{"-y", "mongodb-mcp-server"}, "Database", "★"},
	// Development
	{"github", "GitHub repos, issues, PRs & code search", "npx", []string{"-y", "@modelcontextprotocol/server-github"}, "Development", "★★★"},
	{"gitlab", "GitLab API operations", "npx", []string{"-y", "@modelcontextprotocol/server-gitlab"}, "Development", "★★"},
	{"docker", "Container lifecycle & image management", "npx", []string{"-y", "mcp-server-docker"}, "Development", "★★"},
	{"kubernetes", "K8s cluster management", "npx", []string{"-y", "mcp-server-kubernetes"}, "Development", "★"},
	// Productivity
	{"slack", "Slack messaging & channel operations", "npx", []string{"-y", "@modelcontextprotocol/server-slack"}, "Productivity", "★★"},
	{"notion", "Notion pages & database queries", "npx", []string{"-y", "@modelcontextprotocol/server-notion"}, "Productivity", "★★"},
	{"linear", "Linear issue tracking integration", "npx", []string{"-y", "mcp-linear"}, "Productivity", "★"},
	// Search & Data
	{"brave-search", "Web search via Brave API", "npx", []string{"-y", "@modelcontextprotocol/server-brave-search"}, "Search", "★★"},
	{"exa", "AI-powered web search", "npx", []string{"-y", "exa-mcp-server"}, "Search", "★"},
	// Cloud & Infra
	{"aws", "AWS services integration", "npx", []string{"-y", "@aws/mcp"}, "Cloud", "★★"},
	{"cloudflare", "Cloudflare Workers, KV, R2, D1", "npx", []string{"-y", "@cloudflare/mcp-server-cloudflare"}, "Cloud", "★★"},
}

// ViewMode represents what view is currently shown
type ViewMode int

const (
	ViewBooting ViewMode = iota // Boot animation (first view)
	ViewMain
	ViewAddServer
	ViewCustomServer
	ViewInstalling
	ViewSetupWizard // First-run setup wizard
)

// InstallResult is sent when server installation completes
type InstallResult struct {
	Name    string
	Success bool
	Error   string
}

// ServerCallbacks allows the TUI to communicate server changes to the bridge
type ServerCallbacks struct {
	OnAddServer    func(ctx context.Context, name, desc, cmd string, args []string) error
	OnRemoveServer func(name string) error
	OnToggleServer func(name string, enabled bool) error
}

// App is the main TUI application model
type App struct {
	state *AppState
	theme *Theme

	activeTab    TabIndex
	ready        bool
	quitting     bool
	serverCursor int

	spinner spinner.Model

	// View mode
	viewMode       ViewMode
	addServerIdx   int
	customInputs   []textinput.Model
	customFocusIdx int

	// Modal
	showModal    bool
	modalTitle   string
	modalMessage string
	modalAction  func()

	// Toast
	toast       string
	toastExpiry time.Time

	// Server operation callbacks
	callbacks *ServerCallbacks

	// Installation progress and cancellation
	installingName string
	installError   string
	installCtx     context.Context
	installCancel  context.CancelFunc
	installState   int // 0=idle, 1=running, 2=cancelling

	// Boot animation state
	bootState *BootState

	// Help overlay
	showHelp bool

	// Scrolling state for add server view
	addServerScroll int

	// Search filter for add server view
	addServerSearch   textinput.Model
	addServerSearchOn bool

	// Quit dialog
	showQuitDialog bool
	quitSelection  int      // 0=background, 1=quit
	finalQuitMode  QuitMode // The final quit mode chosen by user

	// Daemon mode
	daemonMode   bool // true if connected to daemon
	daemonClient interface {
		Shutdown() error
		Close() error
	}

	// Setup wizard
	wizard             *WizardModel
	wizardDoneCallback func(WizardResult)

	// Version & update
	currentVersion string
	latestVersion  string // set when update is available
}

// NewApp creates a new TUI application
func NewApp(state *AppState) *App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = DefaultTheme.Spinner

	// Create inputs for custom server
	inputs := make([]textinput.Model, 3)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "my-server"
	inputs[0].CharLimit = 32
	inputs[0].Width = 30
	inputs[0].Prompt = ""

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "npx"
	inputs[1].CharLimit = 64
	inputs[1].Width = 30
	inputs[1].Prompt = ""

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "-y @package/name"
	inputs[2].CharLimit = 128
	inputs[2].Width = 40
	inputs[2].Prompt = ""

	// Search input for add server
	searchInput := textinput.New()
	searchInput.Placeholder = "Filter servers..."
	searchInput.CharLimit = 32
	searchInput.Width = 25
	searchInput.Prompt = "/ "

	return &App{
		state:           state,
		theme:           DefaultTheme,
		activeTab:       TabDashboard,
		spinner:         s,
		customInputs:    inputs,
		viewMode:        ViewBooting,
		bootState:       NewBootState(),
		addServerSearch: searchInput,
	}
}

// SetCallbacks sets the server operation callbacks
func (a *App) SetCallbacks(cb *ServerCallbacks) {
	a.callbacks = cb
}

// DaemonClient interface for daemon communication
type DaemonClient interface {
	Shutdown() error
	Close() error
}

// SetVersion sets the current app version for update checking.
func (a *App) SetVersion(v string) {
	a.currentVersion = v
	a.state.CurrentVersion = v
}

// SetDaemonClient sets the daemon client for background mode support
func (a *App) SetDaemonClient(client DaemonClient) {
	a.daemonClient = client
	a.daemonMode = true
	// Use reconnect boot animation
	a.bootState = NewReconnectBootState()
}

// SetWizard sets the wizard model for first-run setup.
func (a *App) SetWizard(w *WizardModel) {
	a.wizard = w
}

// SetViewMode changes the current view mode.
func (a *App) SetViewMode(mode ViewMode) {
	a.viewMode = mode
}

// SetWizardDoneCallback sets a callback invoked when the wizard completes.
func (a *App) SetWizardDoneCallback(cb func(WizardResult)) {
	a.wizardDoneCallback = cb
}

// GetQuitMode returns how the user chose to quit (background or full)
func (a *App) GetQuitMode() QuitMode {
	return a.finalQuitMode
}

// IsDaemonMode returns whether the app is connected to a daemon
func (a *App) IsDaemonMode() bool {
	return a.daemonMode
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(a.spinner.Tick, a.tick(), a.bootTick(), a.checkForUpdate())
}

func (a *App) checkForUpdate() tea.Cmd {
	currentVer := a.currentVersion
	return func() tea.Msg {
		latest, err := updater.CheckLatestVersion()
		if err != nil {
			return nil
		}
		if updater.IsNewer(currentVer, latest) {
			return UpdateAvailableMsg{Version: latest}
		}
		return nil
	}
}

func (a *App) bootTick() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return BootTickMsg{Time: t}
	})
}

func (a *App) tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg{Time: t}
	})
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return a.handleKey(msg)

	case tea.WindowSizeMsg:
		a.state.SetSize(msg.Width, msg.Height)
		a.ready = true
		return a, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd

	case TickMsg:
		return a, a.tick()

	case StatusUpdateMsg:
		a.state.SetStatus(msg.Status, msg.BackendURL, msg.Error)

	case ServerUpdateMsg:
		a.state.SetServers(msg.Servers)

	case ActivityMsg:
		a.state.AddActivity(msg.Entry)

	case TokenUpdateMsg:
		a.state.SetUser(msg.UserEmail, msg.UserID, "", msg.ExpiresAt)

	case UpdateAvailableMsg:
		a.latestVersion = msg.Version
		a.state.LatestVersion = msg.Version

	case InstallResult:
		if msg.Success {
			a.toast = fmt.Sprintf("✓ %s installed successfully", msg.Name)
			a.toastExpiry = time.Now().Add(3 * time.Second)
		} else {
			a.toast = fmt.Sprintf("✗ Failed: %s", msg.Error)
			a.toastExpiry = time.Now().Add(5 * time.Second)
			a.installError = msg.Error
		}
		a.viewMode = ViewMain
		a.installingName = ""
		a.installState = 0

	case InstallCancelledMsg:
		a.toast = fmt.Sprintf("Installation of %s cancelled", msg.Name)
		a.toastExpiry = time.Now().Add(3 * time.Second)
		a.viewMode = ViewAddServer
		a.installingName = ""
		a.installState = 0

	case BootTickMsg:
		if a.viewMode == ViewBooting && a.bootState != nil {
			if AdvanceBootState(a.bootState) {
				// Boot complete, transition to main view
				a.viewMode = ViewMain
			}
			return a, a.bootTick()
		}

	case RerunSetupMsg:
		wizard := NewWizardModel()
		a.wizard = wizard
		a.viewMode = ViewSetupWizard
		return a, nil

	case WizardDoneMsg:
		a.wizard = nil
		if msg.Result.Completed && a.wizardDoneCallback != nil {
			a.wizardDoneCallback(msg.Result)
		}
		if !msg.Result.Completed {
			// User cancelled — quit
			a.quitting = true
			a.finalQuitMode = QuitModeFull
			return a, tea.Quit
		}
		// Wizard completed — transition to boot animation
		a.viewMode = ViewBooting
		a.bootState = NewBootState()
		return a, a.bootTick()

	}

	return a, nil
}

func (a *App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()

	// Handle setup wizard
	if a.viewMode == ViewSetupWizard && a.wizard != nil {
		a.wizard.SetSize(a.state.GetSize())
		if _, cmd := a.wizard.HandleKey(msg); cmd != nil {
			return a, cmd
		}
		return a, nil
	}

	// Handle boot animation - any key skips to end (except during logo phase)
	if a.viewMode == ViewBooting && a.bootState != nil {
		// Allow skip after logo animation completes (phase 1+)
		if a.bootState.Phase >= 1 {
			SkipBoot(a.bootState)
			a.viewMode = ViewMain
		}
		return a, nil
	}

	// Handle installation in progress - only allow Esc to cancel
	if a.viewMode == ViewInstalling {
		if k == "esc" && a.installCancel != nil && a.installState == 1 {
			a.installState = 2 // cancelling
			a.installCancel()
		}
		return a, nil // Block all other keys during install
	}

	// Handle custom server input mode
	if a.viewMode == ViewCustomServer {
		return a.handleCustomServerKey(msg)
	}

	// Handle add server selection mode
	if a.viewMode == ViewAddServer {
		return a.handleAddServerKey(msg)
	}

	// Handle help overlay
	if a.showHelp {
		switch k {
		case "?", "esc", "q":
			a.showHelp = false
		}
		return a, nil
	}

	// Handle quit dialog
	if a.showQuitDialog {
		switch k {
		case "up", "k", "left", "h":
			a.quitSelection = 0
		case "down", "j", "right", "l":
			a.quitSelection = 1
		case "b", "1":
			// Background mode
			a.showQuitDialog = false
			a.quitting = true
			a.finalQuitMode = QuitModeBackground
			// Just close TUI, daemon keeps running
			if a.daemonClient != nil {
				a.daemonClient.Close()
			}
			return a, tea.Quit
		case "q", "2":
			// Full quit
			a.showQuitDialog = false
			a.quitting = true
			a.finalQuitMode = QuitModeFull
			// Shutdown daemon if connected
			if a.daemonClient != nil {
				a.daemonClient.Shutdown()
			}
			return a, tea.Quit
		case "enter":
			a.showQuitDialog = false
			a.quitting = true
			if a.quitSelection == 0 {
				// Background mode
				a.finalQuitMode = QuitModeBackground
				if a.daemonClient != nil {
					a.daemonClient.Close()
				}
			} else {
				// Full quit
				a.finalQuitMode = QuitModeFull
				if a.daemonClient != nil {
					a.daemonClient.Shutdown()
				}
			}
			return a, tea.Quit
		case "esc":
			a.showQuitDialog = false
		}
		return a, nil
	}

	// Handle modal
	if a.showModal {
		switch k {
		case "y", "enter":
			a.showModal = false
			if a.modalAction != nil {
				a.modalAction()
			}
		case "n", "esc":
			a.showModal = false
		}
		return a, nil
	}

	// Global keys
	switch k {
	case "?":
		a.showHelp = true
		return a, nil
	case "q", "ctrl+c":
		// Show quit dialog instead of quitting immediately
		a.showQuitDialog = true
		a.quitSelection = 0 // Default to background
		return a, nil
	case "1":
		a.activeTab = TabDashboard
	case "2":
		a.activeTab = TabServers
	case "3":
		a.activeTab = TabSettings
	case "tab":
		a.activeTab = (a.activeTab + 1) % TabCount
	case "shift+tab":
		a.activeTab = (a.activeTab + TabCount - 1) % TabCount
	}

	// View-specific keys
	switch a.activeTab {
	case TabServers:
		return a.handleServersKey(k)
	}

	return a, nil
}

func (a *App) handleServersKey(k string) (tea.Model, tea.Cmd) {
	servers := a.state.GetServers()
	n := len(servers)

	switch k {
	case "up", "k":
		if a.serverCursor > 0 {
			a.serverCursor--
		}
	case "down", "j":
		if a.serverCursor < n-1 {
			a.serverCursor++
		}
	case "a":
		a.viewMode = ViewAddServer
		a.addServerIdx = 0
	case "d":
		if n > 0 && a.serverCursor < n {
			name := servers[a.serverCursor].Name
			a.showModal = true
			a.modalTitle = "Remove Server"
			a.modalMessage = fmt.Sprintf("Remove '%s'?", name)
			a.modalAction = func() {
				a.removeServer(name)
			}
		}
	case "e":
		if n > 0 && a.serverCursor < n {
			a.toggleServer(a.serverCursor)
		}
	case "c":
		a.state.ClearActivities()
		a.toast = "Activity cleared"
		a.toastExpiry = time.Now().Add(2 * time.Second)
	}
	return a, nil
}

func (a *App) handleAddServerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()

	// Handle search mode
	if a.addServerSearchOn {
		switch k {
		case "esc":
			a.addServerSearchOn = false
			a.addServerSearch.Blur()
			a.addServerSearch.SetValue("")
			a.addServerIdx = 0
			a.addServerScroll = 0
			return a, nil
		case "enter":
			a.addServerSearchOn = false
			a.addServerSearch.Blur()
			// Keep the filter active but stop editing
			return a, nil
		case "up", "down":
			// Allow navigation while searching
			a.addServerSearchOn = false
			a.addServerSearch.Blur()
		default:
			// Update search input with the original key message
			var cmd tea.Cmd
			a.addServerSearch, cmd = a.addServerSearch.Update(msg)
			// Reset selection when filter changes
			a.addServerIdx = 0
			a.addServerScroll = 0
			return a, cmd
		}
	}

	// Get filtered servers
	filtered := a.getFilteredServers()
	total := len(filtered) + 1 // +1 for custom option
	_, h := a.state.GetSize()
	visibleRows := h - 14 // Account for header, search bar, footer, etc.
	if visibleRows < 5 {
		visibleRows = 5
	}

	switch k {
	case "/":
		a.addServerSearchOn = true
		a.addServerSearch.Focus()
		return a, nil
	case "up", "k":
		if a.addServerIdx > 0 {
			a.addServerIdx--
		}
		a.ensureAddServerVisible(filtered, visibleRows)
	case "down", "j":
		if a.addServerIdx < total-1 {
			a.addServerIdx++
		}
		a.ensureAddServerVisible(filtered, visibleRows)
	case "pgup", "ctrl+u":
		fit := a.selectableItemsFit(filtered, a.addServerScroll, visibleRows)
		if fit < 1 {
			fit = 1
		}
		a.addServerIdx -= fit
		if a.addServerIdx < 0 {
			a.addServerIdx = 0
		}
		a.ensureAddServerVisible(filtered, visibleRows)
	case "pgdown", "ctrl+d":
		fit := a.selectableItemsFit(filtered, a.addServerScroll, visibleRows)
		if fit < 1 {
			fit = 1
		}
		a.addServerIdx += fit
		if a.addServerIdx >= total {
			a.addServerIdx = total - 1
		}
		a.ensureAddServerVisible(filtered, visibleRows)
	case "home", "g":
		a.addServerIdx = 0
		a.addServerScroll = 0
	case "end", "G":
		a.addServerIdx = total - 1
		a.ensureAddServerVisible(filtered, visibleRows)
	case "enter":
		if a.addServerIdx == total-1 {
			// Custom server
			a.viewMode = ViewCustomServer
			a.customFocusIdx = 0
			a.customInputs[0].Focus()
			a.addServerSearch.SetValue("") // Clear search
		} else if a.addServerIdx < len(filtered) {
			// Popular server - start async installation
			srv := filtered[a.addServerIdx]

			// Check if already exists
			for _, s := range a.state.GetServers() {
				if s.Name == srv.Name {
					a.toast = "Server already added"
					a.toastExpiry = time.Now().Add(2 * time.Second)
					return a, nil
				}
			}

			a.viewMode = ViewInstalling
			a.installingName = srv.Name
			a.installError = ""
			a.addServerSearch.SetValue("") // Clear search

			// Return a command that runs the installation
			return a, a.installServerCmd(srv.Name, srv.Description, srv.Command, srv.Args)
		}
	case "esc", "q":
		if a.addServerSearch.Value() != "" {
			// Clear search first
			a.addServerSearch.SetValue("")
			a.addServerIdx = 0
			a.addServerScroll = 0
		} else {
			a.viewMode = ViewMain
		}
	}
	return a, nil
}

// getFilteredServers returns servers matching the current search filter
func (a *App) getFilteredServers() []PopularServer {
	filter := strings.ToLower(a.addServerSearch.Value())
	if filter == "" {
		return PopularServers
	}

	var filtered []PopularServer
	for _, srv := range PopularServers {
		if strings.Contains(strings.ToLower(srv.Name), filter) ||
			strings.Contains(strings.ToLower(srv.Description), filter) ||
			strings.Contains(strings.ToLower(srv.Category), filter) {
			filtered = append(filtered, srv)
		}
	}
	return filtered
}

// selectableItemsFit counts how many selectable items (servers + custom option)
// fit within the given visual row budget, starting from scrollStart.
// Category headers and the separator before the custom option consume visual rows.
func (a *App) selectableItemsFit(filtered []PopularServer, scrollStart, visibleRows int) int {
	rowsUsed := 0
	items := 0
	lastCat := ""
	if scrollStart > 0 && scrollStart <= len(filtered) {
		lastCat = filtered[scrollStart-1].Category
	}
	for i := scrollStart; i < len(filtered) && rowsUsed < visibleRows; i++ {
		if filtered[i].Category != lastCat {
			lastCat = filtered[i].Category
			rowsUsed++ // category header
			if rowsUsed >= visibleRows {
				break
			}
		}
		rowsUsed++
		items++
	}
	// Check if custom option fits (separator + custom = 2 rows)
	if rowsUsed+2 <= visibleRows {
		items++
	}
	return items
}

// ensureAddServerVisible adjusts addServerScroll so that addServerIdx is visible.
func (a *App) ensureAddServerVisible(filtered []PopularServer, visibleRows int) {
	// Scroll up if cursor is above the visible area
	if a.addServerIdx < a.addServerScroll {
		a.addServerScroll = a.addServerIdx
	}
	// Scroll down if cursor is below the visible area
	for a.addServerScroll <= a.addServerIdx {
		fit := a.selectableItemsFit(filtered, a.addServerScroll, visibleRows)
		if fit <= 0 {
			a.addServerScroll++
			continue
		}
		if a.addServerIdx < a.addServerScroll+fit {
			break // selected item is visible
		}
		a.addServerScroll++
	}
}

// installServerCmd returns a tea.Cmd that installs a server asynchronously
func (a *App) installServerCmd(name, desc, cmd string, args []string) tea.Cmd {
	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	a.installCtx = ctx
	a.installCancel = cancel
	a.installState = 1 // running

	return func() tea.Msg {
		defer func() {
			a.installState = 0 // idle
			a.installCancel = nil
			a.installCtx = nil
		}()

		var err error
		if a.callbacks != nil && a.callbacks.OnAddServer != nil {
			err = a.callbacks.OnAddServer(ctx, name, desc, cmd, args)
		}

		// Check if cancelled
		if ctx.Err() != nil {
			return InstallCancelledMsg{Name: name}
		}

		if err != nil {
			return InstallResult{Name: name, Success: false, Error: err.Error()}
		}

		// Update local state
		servers := a.state.GetServers()
		found := false
		for _, s := range servers {
			if s.Name == name {
				found = true
				break
			}
		}
		if !found {
			servers = append(servers, ServerInfo{
				Name:        name,
				Description: desc,
				Command:     cmd,
				Args:        args,
				Enabled:     true,
				Connected:   true,
			})
			a.state.SetServers(servers)
		}

		return InstallResult{Name: name, Success: true}
	}
}

func (a *App) handleCustomServerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()

	switch k {
	case "esc":
		a.viewMode = ViewAddServer
		a.resetCustomInputs()
		return a, nil
	case "tab", "down":
		a.customInputs[a.customFocusIdx].Blur()
		a.customFocusIdx = (a.customFocusIdx + 1) % 3
		a.customInputs[a.customFocusIdx].Focus()
		return a, nil
	case "shift+tab", "up":
		a.customInputs[a.customFocusIdx].Blur()
		a.customFocusIdx = (a.customFocusIdx + 2) % 3
		a.customInputs[a.customFocusIdx].Focus()
		return a, nil
	case "enter":
		name := a.customInputs[0].Value()
		cmd := a.customInputs[1].Value()
		argsStr := a.customInputs[2].Value()

		if name == "" || cmd == "" {
			a.toast = "Name and command required"
			a.toastExpiry = time.Now().Add(2 * time.Second)
			return a, nil
		}

		// Check if already exists
		for _, s := range a.state.GetServers() {
			if s.Name == name {
				a.toast = "Server already exists"
				a.toastExpiry = time.Now().Add(2 * time.Second)
				return a, nil
			}
		}

		args := strings.Fields(argsStr)
		a.viewMode = ViewInstalling
		a.installingName = name
		a.installError = ""
		a.resetCustomInputs()

		return a, a.installServerCmd(name, "Custom server", cmd, args)
	}

	// Update the focused input
	var teaCmd tea.Cmd
	a.customInputs[a.customFocusIdx], teaCmd = a.customInputs[a.customFocusIdx].Update(msg)
	return a, teaCmd
}

func (a *App) resetCustomInputs() {
	for i := range a.customInputs {
		a.customInputs[i].SetValue("")
		a.customInputs[i].Blur()
	}
	a.customFocusIdx = 0
}


func (a *App) addServer(name, desc, cmd string, args []string) {
	servers := a.state.GetServers()

	// Check if exists
	for _, s := range servers {
		if s.Name == name {
			a.toast = "Server already exists"
			a.toastExpiry = time.Now().Add(2 * time.Second)
			return
		}
	}

	// Call the callback if set (persists to config and starts server)
	if a.callbacks != nil && a.callbacks.OnAddServer != nil {
		if err := a.callbacks.OnAddServer(context.Background(), name, desc, cmd, args); err != nil {
			a.toast = fmt.Sprintf("Error: %s", err.Error())
			a.toastExpiry = time.Now().Add(3 * time.Second)
			return
		}
	}

	// Update local state (note: callback may also update state with connected info)
	// Re-fetch servers in case callback updated them
	servers = a.state.GetServers()
	found := false
	for _, s := range servers {
		if s.Name == name {
			found = true
			break
		}
	}
	if !found {
		// Callback didn't add to state, add manually
		servers = append(servers, ServerInfo{
			Name:        name,
			Description: desc,
			Command:     cmd,
			Args:        args,
			Enabled:     true,
			Connected:   false,
		})
		a.state.SetServers(servers)
	}
	a.toast = fmt.Sprintf("Added %s", name)
	a.toastExpiry = time.Now().Add(2 * time.Second)
}

func (a *App) toggleServer(idx int) {
	servers := a.state.GetServers()
	if idx >= 0 && idx < len(servers) {
		newEnabled := !servers[idx].Enabled
		name := servers[idx].Name

		// Call the callback if set (persists to config and starts/stops server)
		if a.callbacks != nil && a.callbacks.OnToggleServer != nil {
			if err := a.callbacks.OnToggleServer(name, newEnabled); err != nil {
				a.toast = fmt.Sprintf("Error: %s", err.Error())
				a.toastExpiry = time.Now().Add(3 * time.Second)
				return
			}
		}

		// Update local state
		servers[idx].Enabled = newEnabled
		if !newEnabled {
			servers[idx].Connected = false
			servers[idx].ToolCount = 0
		}
		a.state.SetServers(servers)

		status := "enabled"
		if !newEnabled {
			status = "disabled"
		}
		a.toast = fmt.Sprintf("Server %s", status)
		a.toastExpiry = time.Now().Add(2 * time.Second)
	}
}

func (a *App) removeServer(name string) {
	// Call the callback if set (persists to config and stops server)
	if a.callbacks != nil && a.callbacks.OnRemoveServer != nil {
		if err := a.callbacks.OnRemoveServer(name); err != nil {
			a.toast = fmt.Sprintf("Error: %s", err.Error())
			a.toastExpiry = time.Now().Add(3 * time.Second)
			return
		}
	}

	servers := a.state.GetServers()
	var newList []ServerInfo
	for _, s := range servers {
		if s.Name != name {
			newList = append(newList, s)
		}
	}
	a.state.SetServers(newList)
	if a.serverCursor >= len(newList) && a.serverCursor > 0 {
		a.serverCursor--
	}
	a.toast = "Server removed"
	a.toastExpiry = time.Now().Add(2 * time.Second)
}

func (a *App) View() (output string) {
	// Recover from any panics to prevent TUI crash
	defer func() {
		if r := recover(); r != nil {
			output = fmt.Sprintf("\n  Error rendering view: %v\n\n  Press 'q' to quit.", r)
		}
	}()

	if a.quitting {
		return ""
	}
	if !a.ready {
		return "\n  " + a.spinner.View() + " Starting..."
	}

	w, h := a.state.GetSize()

	// Ensure minimum dimensions
	if w < 40 {
		w = 40
	}
	if h < 10 {
		h = 10
	}

	// Show setup wizard
	if a.viewMode == ViewSetupWizard && a.wizard != nil {
		a.wizard.SetSize(w, h)
		return a.wizard.View()
	}

	// Show boot animation
	if a.viewMode == ViewBooting && a.bootState != nil {
		return RenderBootScreen(a.bootState, w, h)
	}

	// Show installing view
	if a.viewMode == ViewInstalling {
		return a.viewInstalling(w, h)
	}

	// Show add server view
	if a.viewMode == ViewAddServer || a.viewMode == ViewCustomServer {
		return a.viewAddServer(w, h)
	}

	var b strings.Builder
	b.WriteString(a.viewHeader(w))
	b.WriteString("\n")
	b.WriteString(a.viewTabs(w))
	b.WriteString("\n")
	b.WriteString(a.viewContent(w, h-4))

	lines := strings.Count(b.String(), "\n")
	for lines < h-1 {
		b.WriteString("\n")
		lines++
	}
	b.WriteString(a.viewFooter(w))

	out := b.String()

	if a.showModal {
		out = a.overlayModal(out, w, h)
	}
	if a.showHelp {
		out = a.overlayHelp(out, w, h)
	}
	if a.showQuitDialog {
		out = a.overlayQuitDialog(out, w, h)
	}
	if time.Now().Before(a.toastExpiry) && a.toast != "" {
		out = a.overlayToast(out, w)
	}

	return out
}

func (a *App) viewAddServer(w, h int) string {
	var b strings.Builder

	title := a.theme.Title.Render("Add MCP Server")
	subtitle := a.theme.ValueMuted.Render("Select a popular server or add custom")

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(title))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(subtitle))
	b.WriteString("\n\n")

	if a.viewMode == ViewCustomServer {
		// Custom server form
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(a.theme.Label.Render("Custom Server")))
		b.WriteString("\n\n")

		labels := []string{"Name:", "Command:", "Args:"}
		for i, label := range labels {
			style := a.theme.Label
			if i == a.customFocusIdx {
				style = a.theme.Value.Bold(true)
			}
			b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(style.Render(label)))
			b.WriteString("\n")
			inputStyle := a.theme.Input
			if i == a.customFocusIdx {
				inputStyle = a.theme.InputFocus
			}
			b.WriteString(lipgloss.NewStyle().Padding(0, 4).Render(inputStyle.Render(a.customInputs[i].View())))
			b.WriteString("\n\n")
		}

		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(
			a.theme.Help.Render("[Tab] next field  [Enter] add  [Esc] back"),
		))
	} else {
		// Popular servers list with scrolling and search
		filtered := a.getFilteredServers()
		total := len(filtered) + 1 // +1 for custom option
		visibleRows := h - 14      // Account for search bar
		if visibleRows < 5 {
			visibleRows = 5
		}

		// Show search bar
		searchStyle := a.theme.Input
		if a.addServerSearchOn {
			searchStyle = a.theme.InputFocus
		}
		searchBar := searchStyle.Render(a.addServerSearch.View())
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(searchBar))
		if a.addServerSearch.Value() != "" {
			b.WriteString("  " + a.theme.ValueMuted.Render(fmt.Sprintf("(%d matches)", len(filtered))))
		}
		b.WriteString("\n\n")

		// Show scroll indicator at top
		if a.addServerScroll > 0 {
			scrollUp := a.theme.ValueMuted.Render(fmt.Sprintf("  ▲ %d more above", a.addServerScroll))
			b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(scrollUp))
			b.WriteString("\n")
		}

		// Build flat list including custom option
		type listItem struct {
			isCategory  bool
			isSeparator bool
			isServer    bool
			isCustom    bool
			category    string
			server      *PopularServer
			idx         int
		}

		// Determine which categories have at least one visible server
		visibleCategories := make(map[string]bool)
		for i, srv := range filtered {
			if i >= a.addServerScroll {
				visibleCategories[srv.Category] = true
			}
		}

		var items []listItem
		currentCat := ""
		idx := 0
		for i := range filtered {
			srv := &filtered[i]
			if srv.Category != currentCat {
				currentCat = srv.Category
				items = append(items, listItem{isCategory: true, category: currentCat})
			}
			items = append(items, listItem{isServer: true, server: srv, idx: idx})
			idx++
		}
		items = append(items, listItem{isSeparator: true})
		items = append(items, listItem{isCustom: true, idx: len(filtered)})

		// Render visible items
		rendered := 0
		lastRenderedIdx := -1
		for _, item := range items {
			// Skip servers and custom option before scroll position
			if item.isServer && item.idx < a.addServerScroll {
				continue
			}
			if item.isCustom && item.idx < a.addServerScroll {
				continue
			}
			// Skip categories that have no visible servers
			if item.isCategory && !visibleCategories[item.category] {
				continue
			}

			// Stop if we've rendered enough
			if rendered >= visibleRows {
				break
			}

			if item.isCategory {
				b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(
					a.theme.Label.Render("─── " + item.category + " ───"),
				))
				b.WriteString("\n")
				rendered++
			} else if item.isSeparator {
				b.WriteString("\n")
				rendered++
			} else if item.isServer {
				cursor := "  "
				style := lipgloss.NewStyle()
				if item.idx == a.addServerIdx {
					cursor = a.theme.ListCursor.Render("> ")
					style = a.theme.ListItemActive
				}

				// Check if already added
				added := ""
				for _, s := range a.state.GetServers() {
					if s.Name == item.server.Name {
						added = a.theme.StatusSuccess.Render(" ✓")
						break
					}
				}

				line := fmt.Sprintf("%s%-16s %s %s%s",
					cursor,
					item.server.Name,
					a.theme.ValueMuted.Render(truncate(item.server.Description, 32)),
					a.theme.StatusWarning.Render(item.server.Stars),
					added,
				)
				b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(style.Render(line)))
				b.WriteString("\n")
				rendered++
				lastRenderedIdx = item.idx
			} else if item.isCustom {
				cursor := "  "
				style := lipgloss.NewStyle()
				if item.idx == a.addServerIdx {
					cursor = a.theme.ListCursor.Render("> ")
					style = a.theme.ListItemActive
				}
				line := cursor + a.theme.Value.Render("+ Custom Server...")
				b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(style.Render(line)))
				b.WriteString("\n")
				rendered++
				lastRenderedIdx = item.idx
			}
		}

		// Show scroll indicator at bottom
		remaining := total - 1 - lastRenderedIdx
		if remaining > 0 {
			scrollDown := a.theme.ValueMuted.Render(fmt.Sprintf("  ▼ %d more below", remaining))
			b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(scrollDown))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(
			a.theme.Help.Render("[/] search  [↑↓] navigate  [Enter] select  [Esc] back"),
		))
	}

	// Pad to fill screen
	lines := strings.Count(b.String(), "\n")
	for lines < h-1 {
		b.WriteString("\n")
		lines++
	}

	// Footer
	b.WriteString(a.theme.FooterContainer.Width(w).Render(
		a.theme.Help.Render("Adding servers...") + strings.Repeat(" ", w-40) + a.theme.StatusSuccess.Render("● Connected"),
	))

	out := b.String()
	if time.Now().Before(a.toastExpiry) && a.toast != "" {
		out = a.overlayToast(out, w)
	}
	return out
}

func (a *App) viewInstalling(w, h int) string {
	var b strings.Builder

	title := a.theme.Title.Render("Installing MCP Server")
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(title))
	b.WriteString("\n\n")

	// Progress box - show different content based on state
	var progressContent string
	if a.installState == 2 { // cancelling
		progressContent = a.theme.StatusWarning.Render("⏳ Cancelling installation...") + "\n\n"
		progressContent += a.theme.ValueMuted.Render("• Stopping process\n")
		progressContent += a.theme.ValueMuted.Render("• Cleaning up\n")
	} else {
		progressContent = a.spinner.View() + " Installing " + a.theme.Value.Render(a.installingName) + "...\n\n"
		progressContent += a.theme.ValueMuted.Render("• Downloading package from npm\n")
		progressContent += a.theme.ValueMuted.Render("• Starting MCP server process\n")
		progressContent += a.theme.ValueMuted.Render("• Discovering available tools\n")
		progressContent += a.theme.ValueMuted.Render("• Registering with backend\n")
	}

	box := a.theme.Card.Width(w - 8).Render(progressContent)
	b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(box))
	b.WriteString("\n\n")

	// Info
	if a.installState == 2 {
		info := a.theme.ValueMuted.Render("Please wait while the installation is cancelled...")
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(info))
	} else {
		info := a.theme.ValueMuted.Render("This may take a moment on first run as npm downloads the package...")
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(info))
	}

	// Pad to fill screen
	lines := strings.Count(b.String(), "\n")
	for lines < h-1 {
		b.WriteString("\n")
		lines++
	}

	// Footer with cancel hint
	var footerContent string
	if a.installState == 2 {
		footerContent = a.theme.Help.Render("Cancelling...") + strings.Repeat(" ", w-40) + a.theme.StatusWarning.Render("⏳ Stopping")
	} else {
		footerContent = a.theme.HelpKey.Render("[Esc]") + " " + a.theme.Help.Render("cancel") + strings.Repeat(" ", w-50) + a.spinner.View() + " Installing"
	}
	b.WriteString(a.theme.FooterContainer.Width(w).Render(footerContent))

	return b.String()
}

func (a *App) viewHeader(w int) string {
	email, _, _, _ := a.state.GetUser()
	status, _, _ := a.state.GetStatus()

	logo := a.theme.LogoDot.Render("◉") + a.theme.Logo.Render(" Clara Companion")

	var statusStr string
	switch status {
	case StatusConnected:
		uptime := time.Since(a.state.GetConnectedAt())
		statusStr = a.theme.StatusSuccess.Render("● Live") + " " + a.theme.ValueMuted.Render(formatUptime(uptime))
	case StatusConnecting, StatusReconnecting:
		statusStr = a.spinner.View() + " Connecting"
	default:
		statusStr = a.theme.StatusError.Render("● Disconnected")
	}

	right := statusStr
	if a.latestVersion != "" {
		right += "  " + a.theme.StatusWarning.Render("↑ v"+a.latestVersion)
	}
	if email != "" {
		right += "  " + a.theme.UserEmail.Render(email)
	}

	gap := w - lipgloss.Width(logo) - lipgloss.Width(right) - 4
	if gap < 1 {
		gap = 1
	}

	return a.theme.HeaderContainer.Width(w).Render(logo + strings.Repeat(" ", gap) + right)
}

func (a *App) viewTabs(w int) string {
	names := TabNames()
	var tabs []string

	for i, name := range names {
		label := fmt.Sprintf(" %d %s ", i+1, name)
		if TabIndex(i) == a.activeTab {
			tabs = append(tabs, a.theme.TabActive.Render(label))
		} else {
			tabs = append(tabs, a.theme.TabInactive.Render(label))
		}
	}

	return a.theme.TabContainer.Width(w).Render(lipgloss.JoinHorizontal(lipgloss.Top, tabs...))
}

func (a *App) viewContent(w, h int) string {
	switch a.activeTab {
	case TabDashboard:
		return a.viewDashboard(w)
	case TabServers:
		return a.viewServers(w)
	case TabSettings:
		return a.viewSettings(w)
	}
	return ""
}

func (a *App) viewDashboard(w int) string {
	status, backend, _ := a.state.GetStatus()
	_, _, _, tokenExp := a.state.GetUser()
	summary := a.state.Summary()

	halfW := w/2 - 3
	if halfW < 20 {
		halfW = 20
	}

	// ── Row 1: Connection ──
	var connStatus string
	if status == StatusConnected {
		uptime := time.Since(a.state.GetConnectedAt())
		uptimeStr := formatUptime(uptime)
		connStatus = a.theme.StatusSuccess.Render("● Live") + "  " + a.theme.ValueMuted.Render("uptime "+uptimeStr)
	} else {
		connStatus = a.theme.StatusError.Render("● " + status.String())
	}

	exp := time.Until(tokenExp)
	expStr := fmt.Sprintf("%dm left", int(exp.Minutes()))
	if exp > time.Hour {
		expStr = fmt.Sprintf("%dh %dm left", int(exp.Hours()), int(exp.Minutes())%60)
	}

	connBox := a.theme.Card.Width(w - 4).Render(
		a.theme.Title.Render("Connection") + "\n" +
			connStatus + "\n" +
			a.theme.Label.Render("Backend  ") + a.theme.Value.Render(truncate(backend, w-16)) + "\n" +
			a.theme.Label.Render("Token    ") + a.theme.Value.Render(expStr),
	)

	// ── Row 2: Servers + Recent Activity ──
	servers := a.state.GetServers()
	var srvLines []string
	connected, totalTools := 0, 0
	for _, srv := range servers {
		if !srv.Enabled {
			continue
		}
		if srv.Connected {
			connected++
			totalTools += srv.ToolCount
			srvLines = append(srvLines, fmt.Sprintf("%s %-14s %s",
				a.theme.StatusSuccess.Render("●"),
				truncate(srv.Name, 14),
				a.theme.ValueMuted.Render(fmt.Sprintf("%dt", srv.ToolCount)),
			))
		} else {
			srvLines = append(srvLines, fmt.Sprintf("%s %s",
				a.theme.StatusError.Render("●"),
				a.theme.ValueMuted.Render(srv.Name),
			))
		}
	}
	if len(srvLines) == 0 {
		srvLines = []string{a.theme.ValueMuted.Render("No servers")}
	}
	srvLines = append(srvLines, a.theme.ValueMuted.Render(fmt.Sprintf("  %d connected · %d tools", connected, totalTools)))
	srvBox := a.theme.Card.Width(halfW).Render(
		a.theme.Title.Render("MCP Servers") + "\n" + strings.Join(srvLines, "\n"),
	)

	acts := a.state.GetRecentActivities(4)
	var actLines []string
	for _, act := range acts {
		icon := a.theme.StatusSuccess.Render("✓")
		if !act.Success {
			icon = a.theme.StatusError.Render("✗")
		}
		actLines = append(actLines, fmt.Sprintf("%s %s %s",
			a.theme.ValueMuted.Render(act.Timestamp.Format("15:04")),
			truncate(act.ToolName, 18),
			icon,
		))
	}
	if len(actLines) == 0 {
		actLines = []string{a.theme.ValueMuted.Render("No activity")}
	}
	actBox := a.theme.Card.Width(halfW).Render(
		a.theme.Title.Render("Recent Activity") + "\n" + strings.Join(actLines, "\n"),
	)

	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, srvBox, "  ", actBox)

	stats := a.theme.ValueMuted.Render(fmt.Sprintf("%d servers • %d tools • %d calls",
		summary.ConnectedCount, summary.TotalTools, summary.ActivityCount))

	return lipgloss.NewStyle().Padding(1, 2).Render(
		connBox + "\n\n" + bottomRow + "\n\n" + stats,
	)
}

func (a *App) viewServers(w int) string {
	servers := a.state.GetServers()

	if len(servers) == 0 {
		box := a.theme.Card.Width(w - 4).Render(
			a.theme.ValueMuted.Render("No servers configured.\n\nPress [a] to add a server."),
		)
		return lipgloss.NewStyle().Padding(1, 2).Render(box + "\n\n" + a.helpServers())
	}

	hdr := fmt.Sprintf("   %-18s %-12s %-6s %s",
		a.theme.Label.Render("NAME"),
		a.theme.Label.Render("STATUS"),
		a.theme.Label.Render("TOOLS"),
		a.theme.Label.Render("DESCRIPTION"),
	)

	var rows []string
	for i, srv := range servers {
		cursor := "   "
		if i == a.serverCursor {
			cursor = a.theme.ListCursor.Render(" > ")
		}

		var status string
		if !srv.Enabled {
			status = a.theme.ValueMuted.Render("disabled")
		} else if srv.Connected {
			status = a.theme.StatusSuccess.Render("connected")
		} else if srv.Error != "" {
			status = a.theme.StatusError.Render("error")
		} else {
			status = a.theme.StatusWarning.Render("pending")
		}

		tools := ""
		if srv.Connected {
			tools = fmt.Sprintf("%d", srv.ToolCount)
		}

		row := fmt.Sprintf("%s%-18s %-12s %-6s %s",
			cursor, truncate(srv.Name, 18), status, tools, truncate(srv.Description, 28),
		)
		if i == a.serverCursor {
			row = a.theme.ListItemActive.Render(row)
		}
		rows = append(rows, row)
	}

	content := hdr + "\n" + strings.Repeat("─", w-6) + "\n" + strings.Join(rows, "\n")
	return lipgloss.NewStyle().Padding(1, 2).Render(content + "\n\n" + a.helpServers())
}


func (a *App) viewSettings(w int) string {
	email, userID, deviceID, tokenExp := a.state.GetUser()
	_, backend, _ := a.state.GetStatus()

	rows := []string{
		a.theme.Label.Render("Backend URL:") + "\n  " + a.theme.Value.Render(backend),
		a.theme.Label.Render("Config:") + "\n  " + a.theme.Value.Render("~/.claraverse/mcp-config.yaml"),
		a.theme.Label.Render("Token Expires:") + "\n  " + a.theme.Value.Render(tokenExp.Format("2006-01-02 15:04")),
	}

	// Device section (absorbed from Devices tab)
	deviceContent := a.theme.Title.Render("Device") + " " + a.theme.StatusSuccess.Render("●") + "\n" +
		a.theme.Label.Render("  Email: ") + a.theme.Value.Render(email) + "\n" +
		a.theme.Label.Render("  User: ") + a.theme.ValueMuted.Render(truncate(userID, 24)) + "\n" +
		a.theme.Label.Render("  Device: ") + a.theme.ValueMuted.Render(truncate(deviceID, 24))
	rows = append(rows, deviceContent)

	return lipgloss.NewStyle().Padding(1, 2).Render(strings.Join(rows, "\n\n"))
}

func (a *App) viewFooter(w int) string {
	help := a.getHelp()
	status, _, _ := a.state.GetStatus()

	var statusStr string
	if status == StatusConnected {
		statusStr = a.theme.StatusSuccess.Render("● Live")
	} else {
		statusStr = a.theme.StatusError.Render("● " + status.String())
	}

	versionStr := ""
	if a.latestVersion != "" {
		versionStr = a.theme.StatusWarning.Render("Update v"+a.latestVersion+" available") + "  "
	} else if a.currentVersion != "" {
		versionStr = a.theme.ValueMuted.Render("v"+a.currentVersion) + "  "
	}

	rightSide := versionStr + statusStr
	gap := w - lipgloss.Width(help) - lipgloss.Width(rightSide) - 4
	if gap < 1 {
		gap = 1
	}

	return a.theme.FooterContainer.Width(w).Render(help + strings.Repeat(" ", gap) + rightSide)
}

func (a *App) getHelp() string {
	h := func(k, d string) string {
		return a.theme.HelpKey.Render("["+k+"]") + " " + a.theme.Help.Render(d)
	}
	base := h("1-3", "tabs") + "  "
	switch a.activeTab {
	case TabServers:
		return base + h("↑↓", "nav") + "  " + h("a", "add") + "  " + h("d", "del") + "  " + h("e", "toggle") + "  " + h("q", "quit")
	default:
		return base + h("q", "quit")
	}
}

func (a *App) helpServers() string {
	return a.theme.Help.Render("[↑↓] navigate  [a] add  [d] remove  [e] toggle  [q] quit")
}


func (a *App) overlayModal(base string, w, h int) string {
	content := a.theme.ModalTitle.Render(a.modalTitle) + "\n\n" +
		a.theme.ModalContent.Render(a.modalMessage) + "\n\n" +
		a.theme.Help.Render("[y] yes  [n] no")

	modal := a.theme.ModalContainer.Render(content)

	// Use lipgloss.Place for proper ANSI-safe centering
	centered := lipgloss.Place(w, h,
		lipgloss.Center, lipgloss.Center,
		modal,
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#000000")),
	)
	return centered
}

func (a *App) overlayToast(base string, w int) string {
	// Truncate toast message if too long for window
	maxLen := w - 8
	if maxLen < 10 {
		maxLen = 10
	}
	toastMsg := a.toast
	if len(toastMsg) > maxLen {
		toastMsg = toastMsg[:maxLen-3] + "..."
	}

	toast := a.theme.StatusSuccess.
		Background(ColorSurface).
		Padding(0, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorSuccess).
		Render(toastMsg)

	x := (w - lipgloss.Width(toast)) / 2
	if x < 0 {
		x = 0
	}
	return overlay(x, 2, toast, base)
}

func (a *App) overlayHelp(base string, w, h int) string {
	// Build help content with all keybindings
	var content strings.Builder

	title := a.theme.ModalTitle.Render("Keyboard Shortcuts")
	content.WriteString(title)
	content.WriteString("\n\n")

	// Navigation section
	content.WriteString(a.theme.Label.Render("Navigation"))
	content.WriteString("\n")
	content.WriteString(a.helpLine("1-3", "Switch tabs"))
	content.WriteString(a.helpLine("Tab", "Next tab"))
	content.WriteString(a.helpLine("Shift+Tab", "Previous tab"))
	content.WriteString(a.helpLine("j/k or ↑/↓", "Navigate lists"))
	content.WriteString("\n")

	// Server management section
	content.WriteString(a.theme.Label.Render("Server Management"))
	content.WriteString("\n")
	content.WriteString(a.helpLine("a", "Add new server"))
	content.WriteString(a.helpLine("d", "Remove server"))
	content.WriteString(a.helpLine("e", "Enable/disable server"))
	content.WriteString(a.helpLine("Enter", "Select/confirm"))
	content.WriteString(a.helpLine("Esc", "Cancel/back"))
	content.WriteString("\n")

	// General section
	content.WriteString(a.theme.Label.Render("General"))
	content.WriteString("\n")
	content.WriteString(a.helpLine("?", "Show/hide help"))
	content.WriteString(a.helpLine("q", "Quit"))
	content.WriteString("\n")

	content.WriteString(a.theme.ValueMuted.Render("Press ? or Esc to close"))

	// Create modal box
	helpBox := lipgloss.NewStyle().
		Background(ColorSurface).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent).
		Padding(1, 2).
		Render(content.String())

	// Use lipgloss.Place to center the modal properly (handles ANSI correctly)
	centered := lipgloss.Place(w, h,
		lipgloss.Center, lipgloss.Center,
		helpBox,
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#000000")),
	)

	return centered
}

func (a *App) helpLine(key, desc string) string {
	return fmt.Sprintf("  %s %s\n",
		a.theme.HelpKey.Render(fmt.Sprintf("%-12s", key)),
		a.theme.Help.Render(desc))
}

func (a *App) overlayQuitDialog(base string, w, h int) string {
	var content strings.Builder

	title := a.theme.ModalTitle.Render("Exit Clara Companion")
	content.WriteString(title)
	content.WriteString("\n\n")

	content.WriteString(a.theme.ModalContent.Render("How would you like to exit?"))
	content.WriteString("\n\n")

	// Option 1: Run in background
	opt1Cursor := "  "
	opt1Style := a.theme.Value
	if a.quitSelection == 0 {
		opt1Cursor = a.theme.ListCursor.Render("> ")
		opt1Style = a.theme.ListItemActive
	}
	content.WriteString(opt1Style.Render(opt1Cursor + "[b] Run in background"))
	content.WriteString("\n")
	content.WriteString(a.theme.ValueMuted.Render("      Keep servers running, close UI"))
	content.WriteString("\n\n")

	// Option 2: Full quit
	opt2Cursor := "  "
	opt2Style := a.theme.Value
	if a.quitSelection == 1 {
		opt2Cursor = a.theme.ListCursor.Render("> ")
		opt2Style = a.theme.ListItemActive
	}
	content.WriteString(opt2Style.Render(opt2Cursor + "[q] Quit completely"))
	content.WriteString("\n")
	content.WriteString(a.theme.ValueMuted.Render("      Stop all servers and exit"))
	content.WriteString("\n\n")

	content.WriteString(a.theme.Help.Render("[↑↓] select  [Enter] confirm  [Esc] cancel"))

	// Create modal box
	quitBox := lipgloss.NewStyle().
		Background(ColorSurface).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent).
		Padding(1, 2).
		Render(content.String())

	// Use lipgloss.Place to center the modal properly
	centered := lipgloss.Place(w, h,
		lipgloss.Center, lipgloss.Center,
		quitBox,
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#000000")),
	)

	return centered
}

func formatUptime(d time.Duration) string {
	s := int(d.Seconds()) % 60
	m := int(d.Minutes()) % 60
	h := int(d.Hours())
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func truncate(s string, max int) string {
	if len(s) <= max || max < 3 {
		return s
	}
	return s[:max-2] + ".."
}

func overlay(x, y int, top, base string) string {
	// Handle edge cases
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	baseLines := strings.Split(base, "\n")
	topLines := strings.Split(top, "\n")
	topW := lipgloss.Width(top)

	for i, line := range topLines {
		row := y + i
		if row < 0 || row >= len(baseLines) {
			continue
		}

		baseLine := baseLines[row]
		baseW := lipgloss.Width(baseLine)

		// Build the new line: left part + overlay + right part
		var newLine strings.Builder

		// Add left padding/content before overlay
		if x > 0 {
			if baseW >= x {
				// Use ANSI-safe truncation for the left part
				newLine.WriteString(truncateAnsi(baseLine, x))
			} else {
				// Base line is shorter than x, pad with spaces
				newLine.WriteString(baseLine)
				newLine.WriteString(strings.Repeat(" ", x-baseW))
			}
		}

		// Add the overlay line
		newLine.WriteString(line)

		// Add right part of base if it extends beyond overlay
		lineW := lipgloss.Width(line)
		rightStart := x + lineW
		if baseW > rightStart {
			// Skip the portion covered by overlay and add the rest
			// For simplicity, just pad to maintain width
			remaining := baseW - rightStart
			if remaining > 0 {
				newLine.WriteString(strings.Repeat(" ", remaining))
			}
		}

		baseLines[row] = newLine.String()
		_ = topW // suppress unused warning
	}
	return strings.Join(baseLines, "\n")
}

// truncateAnsi truncates a string to a visual width, preserving ANSI sequences
func truncateAnsi(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}

	// Use lipgloss to handle ANSI-aware truncation
	style := lipgloss.NewStyle().MaxWidth(width)
	return style.Render(s)
}

// Run starts the TUI
func Run(state *AppState) error {
	app := NewApp(state)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
