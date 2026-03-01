package tui

import (
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ClaraVerse ASCII logo - compact version for terminal
var claraverseLogo = `
 ██████╗██╗      █████╗ ██████╗  █████╗ ██╗   ██╗███████╗██████╗ ███████╗███████╗
██╔════╝██║     ██╔══██╗██╔══██╗██╔══██╗██║   ██║██╔════╝██╔══██╗██╔════╝██╔════╝
██║     ██║     ███████║██████╔╝███████║██║   ██║█████╗  ██████╔╝███████╗█████╗
██║     ██║     ██╔══██║██╔══██╗██╔══██║╚██╗ ██╔╝██╔══╝  ██╔══██╗╚════██║██╔══╝
╚██████╗███████╗██║  ██║██║  ██║██║  ██║ ╚████╔╝ ███████╗██║  ██║███████║███████╗
 ╚═════╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝  ╚═══╝  ╚══════╝╚═╝  ╚═╝╚══════╝╚══════╝`

// Smaller logo for narrow terminals
var claraverseLogoSmall = `
 ██████╗██╗   ██╗
██╔════╝██║   ██║
██║     ██║   ██║
██║     ╚██╗ ██╔╝
╚██████╗ ╚████╔╝
 ╚═════╝  ╚═══╝  `

// Boot steps shown during initialization
var bootSteps = []string{
	"Initializing core systems",
	"Loading MCP configuration",
	"Starting server registry",
	"Connecting to backend",
	"Registering tools",
	"System ready",
}

// Boot steps for reconnecting to existing daemon
var reconnectSteps = []string{
	"Detecting background service",
	"Reconnecting to daemon",
	"Syncing state",
	"Service online",
}

// Glitch characters for cyberpunk effect
var glitchChars = []rune{'░', '▒', '▓', '█', '╬', '╫', '╪', '┼', '╳', '▀', '▄', '▌', '▐'}

// BootState holds the boot animation state
type BootState struct {
	Phase       int       // 0=logo animating, 1=steps, 2=ready, 3=complete
	Step        int       // Current boot step (0-5)
	Tick        int       // Animation frame counter
	GlitchSeed  int64     // Random seed for consistent glitch
	ReadyAt     time.Time // When "ready" state started
	StartedAt   time.Time // Boot start time
	Reconnecting bool     // True if reconnecting to existing daemon
}

// NewBootState creates a new boot animation state
func NewBootState() *BootState {
	return &BootState{
		Phase:      0,
		Step:       0,
		Tick:       0,
		GlitchSeed: time.Now().UnixNano(),
		StartedAt:  time.Now(),
	}
}

// NewReconnectBootState creates a boot state for reconnecting to daemon
func NewReconnectBootState() *BootState {
	return &BootState{
		Phase:       0,
		Step:        0,
		Tick:        0,
		GlitchSeed:  time.Now().UnixNano(),
		StartedAt:   time.Now(),
		Reconnecting: true,
	}
}

// BootStyles defines colors for boot animation
type BootStyles struct {
	LogoPrimary   lipgloss.Style // Rose pink
	LogoSecondary lipgloss.Style // Cyan accent
	LogoGlitch    lipgloss.Style // Glitch color
	StepComplete  lipgloss.Style // Completed step
	StepCurrent   lipgloss.Style // Current step (animated)
	StepPending   lipgloss.Style // Pending step
	Ready         lipgloss.Style // Ready message
	Subtitle      lipgloss.Style // Subtitle text
}

// DefaultBootStyles returns cyberpunk-themed boot styles
func DefaultBootStyles() *BootStyles {
	return &BootStyles{
		LogoPrimary:   lipgloss.NewStyle().Foreground(ColorAccent),      // Rose pink
		LogoSecondary: lipgloss.NewStyle().Foreground(ColorInfo),        // Cyan
		LogoGlitch:    lipgloss.NewStyle().Foreground(ColorTextMuted),   // Glitch artifacts
		StepComplete:  lipgloss.NewStyle().Foreground(ColorSuccess),     // Green
		StepCurrent:   lipgloss.NewStyle().Foreground(ColorAccent).Bold(true),
		StepPending:   lipgloss.NewStyle().Foreground(ColorTextMuted),
		Ready:         lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true),
		Subtitle:      lipgloss.NewStyle().Foreground(ColorTextMuted).Italic(true),
	}
}

// RenderBootScreen renders the boot animation screen
func RenderBootScreen(state *BootState, w, h int) string {
	styles := DefaultBootStyles()
	var b strings.Builder

	// Choose logo based on terminal width
	logo := claraverseLogo
	if w < 85 {
		logo = claraverseLogoSmall
	}

	// Apply glitch effect throughout boot (cyberpunk aesthetic)
	// Intensity varies with tick for dynamic effect
	if state.Phase < 3 {
		logo = applyGlitch(logo, state.Tick, state.GlitchSeed)
	}

	// Color the logo with gradient effect
	logoLines := strings.Split(logo, "\n")
	coloredLogo := colorLogo(logoLines, styles, state.Tick)

	// Choose steps based on mode
	steps := bootSteps
	if state.Reconnecting {
		steps = reconnectSteps
	}

	// Calculate vertical centering
	totalHeight := len(logoLines) + 2 + len(steps) + 4
	topPadding := (h - totalHeight) / 2
	if topPadding < 1 {
		topPadding = 1
	}

	// Add top padding
	for i := 0; i < topPadding; i++ {
		b.WriteString("\n")
	}

	// Render centered logo - center each line individually for proper alignment
	for _, line := range strings.Split(coloredLogo, "\n") {
		b.WriteString(centerText(line, w))
		b.WriteString("\n")
	}

	// Subtitle - different for reconnect mode
	subtitle := "MCP Bridge Terminal"
	if state.Reconnecting {
		subtitle = "Reconnecting to Background Service"
	}
	b.WriteString(centerText(styles.Subtitle.Render(subtitle), w))
	b.WriteString("\n\n")

	// Boot steps
	if state.Phase >= 1 {
		for i, step := range steps {
			var line string
			if i < state.Step {
				// Completed
				line = styles.StepComplete.Render("  [✓] " + step)
			} else if i == state.Step {
				// Current (animated spinner)
				spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
				spinner := spinners[state.Tick%len(spinners)]
				line = styles.StepCurrent.Render("  [" + spinner + "] " + step)
			} else {
				// Pending
				line = styles.StepPending.Render("  [ ] " + step)
			}
			b.WriteString(centerText(line, w))
			b.WriteString("\n")
		}
	}

	// Ready message
	if state.Phase == 2 {
		b.WriteString("\n")
		readyText := styles.Ready.Render(">>> SYSTEM ONLINE <<<")
		b.WriteString(centerText(readyText, w))
		b.WriteString("\n")

		// Blinking "Press any key" (toggles every ~500ms / 10 ticks)
		if (state.Tick/10)%2 == 0 {
			pressKey := styles.Subtitle.Render("Press any key to continue...")
			b.WriteString(centerText(pressKey, w))
		}
	}

	// Pad to fill screen
	lines := strings.Count(b.String(), "\n")
	for lines < h-1 {
		b.WriteString("\n")
		lines++
	}

	return b.String()
}

// applyGlitch applies cyberpunk glitch effect to text
func applyGlitch(text string, tick int, seed int64) string {
	rng := rand.New(rand.NewSource(seed + int64(tick)))
	runes := []rune(text)

	// Pulsing glitch intensity - cycles between low and high
	// Creates a continuous cyberpunk aesthetic
	cycle := tick % 20
	var glitchCount int
	if cycle < 5 {
		// Intense burst
		glitchCount = 15 + rng.Intn(10)
	} else if cycle < 10 {
		// Medium
		glitchCount = 5 + rng.Intn(5)
	} else {
		// Low/subtle
		glitchCount = 2 + rng.Intn(3)
	}

	for i := 0; i < glitchCount; i++ {
		idx := rng.Intn(len(runes))
		// Don't glitch newlines or spaces
		if runes[idx] != '\n' && runes[idx] != ' ' {
			runes[idx] = glitchChars[rng.Intn(len(glitchChars))]
		}
	}

	// Occasionally shift entire lines horizontally (scanline effect)
	if rng.Float32() < 0.15 {
		lines := strings.Split(string(runes), "\n")
		lineIdx := rng.Intn(len(lines))
		if len(lines[lineIdx]) > 2 {
			shift := rng.Intn(3) - 1 // -1, 0, or 1
			if shift > 0 {
				lines[lineIdx] = " " + lines[lineIdx][:len(lines[lineIdx])-1]
			} else if shift < 0 && len(lines[lineIdx]) > 0 {
				lines[lineIdx] = lines[lineIdx][1:] + " "
			}
		}
		return strings.Join(lines, "\n")
	}

	return string(runes)
}

// colorLogo applies gradient coloring to logo
func colorLogo(lines []string, styles *BootStyles, tick int) string {
	var result strings.Builder

	for i, line := range lines {
		if line == "" {
			result.WriteString("\n")
			continue
		}

		// Alternate colors based on line position and animation tick
		// Creates a wave/scan effect
		wavePos := (tick / 2) % (len(lines) * 2)
		distFromWave := abs(i - wavePos)

		if distFromWave < 2 {
			// Wave highlight in cyan
			result.WriteString(styles.LogoSecondary.Render(line))
		} else {
			// Base color rose pink
			result.WriteString(styles.LogoPrimary.Render(line))
		}
		result.WriteString("\n")
	}

	return strings.TrimSuffix(result.String(), "\n")
}

// Helper functions

func centerText(text string, width int) string {
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}
	padding := (width - textWidth) / 2
	return strings.Repeat(" ", padding) + text
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// AdvanceBootState advances the boot animation state
// Returns true when boot is complete
func AdvanceBootState(state *BootState) bool {
	state.Tick++

	// Choose steps based on mode
	steps := bootSteps
	if state.Reconnecting {
		steps = reconnectSteps
	}

	switch state.Phase {
	case 0: // Logo animation
		// Logo glitch effect runs for ~30 ticks (1.5s at 50ms/tick)
		if state.Tick >= 30 {
			state.Phase = 1
			state.Tick = 0
		}
	case 1: // Boot steps
		// Each step takes ~8 ticks (400ms) for faster feel
		if state.Tick%8 == 0 && state.Tick > 0 {
			state.Step++
			if state.Step >= len(steps) {
				state.Phase = 2
				state.ReadyAt = time.Now()
				state.Tick = 0
			}
		}
	case 2: // Ready (hold for 2 seconds or key press)
		if time.Since(state.ReadyAt) >= 2*time.Second {
			state.Phase = 3
			return true
		}
	case 3: // Complete
		return true
	}

	return false
}

// SkipBoot immediately completes the boot animation
func SkipBoot(state *BootState) {
	state.Phase = 3
}
