package coretools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/claraverse/mcp-client/internal/config"
	"github.com/claraverse/mcp-client/internal/registry"
)

// launchedChromeCmd holds the Chrome process we started so we can kill it on shutdown.
var launchedChromeCmd *exec.Cmd

// CleanupLaunchedChrome kills the Chrome process that was auto-launched, if any.
func CleanupLaunchedChrome() {
	cmd := launchedChromeCmd
	if cmd == nil || cmd.Process == nil {
		return
	}
	log.Printf("[BROWSER] Cleaning up auto-launched Chrome (PID %d)", cmd.Process.Pid)
	// Kill the entire process group
	syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	done := make(chan struct{})
	go func() { cmd.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	launchedChromeCmd = nil
}

// BrowserTool delegates browser operations to a Playwright MCP server if available.
// Includes auto-recovery: detects broken pipes, stale Chrome locks, and orphaned
// processes, then restarts the browser MCP server and retries the operation.
type BrowserTool struct {
	mcpRegistry *registry.Registry
	mu          sync.Mutex
	lastRestart time.Time // rate-limit restarts to at most once per 30s
}

func NewBrowserTool(mcpRegistry *registry.Registry) *BrowserTool {
	return &BrowserTool{
		mcpRegistry: mcpRegistry,
	}
}

func (t *BrowserTool) Name() string { return "browser" }

func (t *BrowserTool) Description() string {
	return "Control a web browser to navigate pages, interact with elements, and extract content. " +
		"Delegates to a browser MCP server (Chrome DevTools, Playwright, or Puppeteer).\n\n" +
		"Actions:\n" +
		"- navigate: go to a URL (requires 'url' parameter).\n" +
		"- click: click an element (requires 'selector' — CSS selector or element ref).\n" +
		"- fill: type into an input field (requires 'selector' and 'value').\n" +
		"- screenshot: capture a screenshot of the current page.\n" +
		"- content: get the page's text content.\n" +
		"- snapshot: get an accessibility tree snapshot of the page.\n\n" +
		"Includes automatic recovery: if Chrome crashes or the connection breaks " +
		"(broken pipe, EOF, connection reset), the browser is automatically restarted " +
		"and the operation is retried. Chrome is auto-launched with remote debugging " +
		"if not already running."
}

func (t *BrowserTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "The browser action: 'navigate', 'click', 'fill', 'screenshot', 'content', 'snapshot'",
				"enum":        []string{"navigate", "click", "fill", "screenshot", "content", "snapshot"},
			},
			"url": map[string]interface{}{
				"type":        "string",
				"description": "URL to navigate to (for 'navigate' action).",
			},
			"selector": map[string]interface{}{
				"type":        "string",
				"description": "CSS selector or element ref for click/fill actions.",
			},
			"value": map[string]interface{}{
				"type":        "string",
				"description": "Value to fill (for 'fill' action).",
			},
		},
		"required": []string{"action"},
	}
}

// browserToolCandidates maps our action names to possible MCP tool names.
// Different browser MCP servers use different conventions (browser_navigate vs navigate_page).
// We try each candidate in order.
var browserToolCandidates = map[string][]string{
	"navigate":   {"navigate_page", "browser_navigate"},
	"click":      {"click", "browser_click"},
	"fill":       {"fill", "browser_fill"},
	"screenshot": {"take_screenshot", "browser_take_screenshot"},
	"content":    {"browser_content", "get_content"},
	"snapshot":   {"take_snapshot", "browser_take_snapshot"},
}

// browserToolNames is a flat set of all known browser tool names (for findBrowserServer fallback).
var browserToolNames = map[string]string{
	"navigate":   "navigate_page",
	"click":      "click",
	"fill":       "fill",
	"screenshot": "take_screenshot",
	"content":    "browser_content",
	"snapshot":   "take_snapshot",
}

// knownBrowserServerNames are server names that provide browser MCP tools.
var knownBrowserServerNames = []string{
	"chrome",
	"playwright",
	"puppeteer",
	"browser",
	"browsermcp",
}

func (t *BrowserTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	action, ok := args["action"].(string)
	if !ok || action == "" {
		return "", fmt.Errorf("action is required")
	}

	// Find candidate MCP tool names for this action
	candidates, ok := browserToolCandidates[action]
	if !ok {
		return "", fmt.Errorf("unknown browser action: %s", action)
	}

	// Build MCP tool arguments
	mcpArgs := make(map[string]interface{})
	switch action {
	case "navigate":
		url, _ := args["url"].(string)
		if url == "" {
			return "", fmt.Errorf("url is required for navigate action")
		}
		mcpArgs["url"] = url
	case "click":
		selector, _ := args["selector"].(string)
		if selector == "" {
			return "", fmt.Errorf("selector is required for click action")
		}
		mcpArgs["selector"] = selector
	case "fill":
		selector, _ := args["selector"].(string)
		value, _ := args["value"].(string)
		if selector == "" {
			return "", fmt.Errorf("selector is required for fill action")
		}
		mcpArgs["selector"] = selector
		mcpArgs["value"] = value
	}

	// Try each candidate tool name (different MCP servers use different names)
	var lastErr error
	for _, mcpToolName := range candidates {
		result, err := t.mcpRegistry.ExecuteTool(mcpToolName, mcpArgs)
		if err == nil {
			return result, nil
		}
		lastErr = err
		// If the error is NOT "not found", don't try other names — it's a real error
		if !strings.Contains(err.Error(), "not found") {
			break
		}
	}

	// Check if this is a recoverable error (broken pipe, EOF, connection reset)
	if !isBrowserRecoverableError(lastErr) {
		return "", fmt.Errorf("browser action failed: %w", lastErr)
	}

	// Attempt auto-recovery: clean up Chrome state and restart browser server
	log.Printf("[BROWSER] Recoverable error detected: %v — attempting auto-recovery", lastErr)

	if restartErr := t.recoverBrowserServer(); restartErr != nil {
		return "", fmt.Errorf("browser action failed and auto-recovery also failed: original=%v, recovery=%v", lastErr, restartErr)
	}

	// Retry once after recovery — try all candidates again
	for _, mcpToolName := range candidates {
		log.Printf("[BROWSER] Retrying %s after auto-recovery", mcpToolName)
		result, retryErr := t.mcpRegistry.ExecuteTool(mcpToolName, mcpArgs)
		if retryErr == nil {
			return result, nil
		}
		if !strings.Contains(retryErr.Error(), "not found") {
			return "", fmt.Errorf("browser action failed after auto-recovery: %w", retryErr)
		}
	}

	return "", fmt.Errorf("browser action failed after auto-recovery: no working browser tool found")
}

// isBrowserRecoverableError checks if the error indicates a broken browser
// connection that might be fixed by restarting the browser MCP server.
func isBrowserRecoverableError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	recoverable := []string{
		"broken pipe",
		"connection reset",
		"eof",
		"write: broken pipe",
		"failed to flush",
		"use of closed network connection",
		"connection refused",
		"not found in any running server",
	}
	for _, pattern := range recoverable {
		if strings.Contains(msg, pattern) {
			return true
		}
	}
	return false
}

// recoverBrowserServer cleans up stale Chrome state and restarts the browser
// MCP server. Rate-limited to at most once per 30 seconds.
// After Chrome recovery, fetches a fresh WebSocket endpoint to avoid stale
// --wsEndpoint URLs (the browser UUID changes on each Chrome restart).
func (t *BrowserTool) recoverBrowserServer() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Rate limit: don't restart more than once per 30s
	if time.Since(t.lastRestart) < 30*time.Second {
		return fmt.Errorf("browser restart rate-limited (last restart %v ago)", time.Since(t.lastRestart))
	}

	log.Printf("[BROWSER] Starting auto-recovery...")

	// Step 1: Clean up stale Chrome locks
	cleanupChromeLocks()

	// Step 2: Kill orphaned Chrome processes that might hold the profile lock
	killOrphanedChrome()

	// Step 3: Try to ensure Chrome is available on port 9222
	EnsureBrowserReady()

	// Step 4: Find the browser MCP server in the registry
	serverName := t.findBrowserServer()
	if serverName == "" {
		return fmt.Errorf("no browser MCP server found in registry to restart")
	}

	// Step 5: Fetch fresh WebSocket endpoint (browser UUID changes on restart)
	// and restart with updated args to avoid stale --wsEndpoint URLs
	wsEndpoint := fetchChromeWSEndpoint()
	if wsEndpoint != "" {
		log.Printf("[BROWSER] Got fresh WebSocket endpoint: %s", wsEndpoint)
		// Stop old server, start fresh with new wsEndpoint
		t.mcpRegistry.StopServer(serverName)
		time.Sleep(500 * time.Millisecond)

		npxPath, err := exec.LookPath("npx")
		if err == nil {
			srv := config.MCPServer{
				Name:        serverName,
				Description: "Chrome DevTools MCP (auto-recovered)",
				Command:     npxPath,
				Args:        []string{"chrome-devtools-mcp@latest", "--wsEndpoint", wsEndpoint},
				Type:        "stdio",
				Enabled:     true,
			}
			if err := t.mcpRegistry.StartServer(srv); err != nil {
				return fmt.Errorf("failed to restart browser server with fresh wsEndpoint: %w", err)
			}
			t.lastRestart = time.Now()
			log.Printf("[BROWSER] Auto-recovery complete — %s restarted with fresh WebSocket endpoint", serverName)
			return nil
		}
	}

	// Fallback: simple restart with existing config
	log.Printf("[BROWSER] Restarting browser MCP server: %s (existing config)", serverName)
	if err := t.mcpRegistry.RestartServer(serverName); err != nil {
		return fmt.Errorf("failed to restart browser server %s: %w", serverName, err)
	}

	t.lastRestart = time.Now()
	log.Printf("[BROWSER] Auto-recovery complete — browser server %s restarted", serverName)
	return nil
}

// fetchChromeWSEndpoint gets the CDP WebSocket debugger URL from Chrome's
// /json/version endpoint. Returns empty string on failure.
func fetchChromeWSEndpoint() string {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://127.0.0.1:9222/json/version")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var info struct {
		WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return ""
	}
	return info.WebSocketDebuggerURL
}

// findBrowserServer looks through the registry for a running browser MCP server.
func (t *BrowserTool) findBrowserServer() string {
	serverNames := t.mcpRegistry.GetServerNames()
	for _, name := range serverNames {
		lower := strings.ToLower(name)
		for _, known := range knownBrowserServerNames {
			if strings.Contains(lower, known) {
				return name
			}
		}
	}

	// Fallback: check if any server provides known browser tool names
	for _, mcpName := range browserToolNames {
		serverName := t.mcpRegistry.GetServerForTool(mcpName)
		if serverName != "" {
			return serverName
		}
	}
	return ""
}

// cleanupChromeLocks removes stale Chrome profile lock files that prevent
// new Chrome instances from starting.
func cleanupChromeLocks() {
	// Common Chrome profile directories used by browser MCP servers
	homeDir, _ := os.UserHomeDir()
	lockPaths := []string{
		filepath.Join(homeDir, ".cache", "chrome-devtools-mcp", "chrome-profile", "SingletonLock"),
		filepath.Join(homeDir, ".cache", "chrome-devtools-mcp", "chrome-profile", "SingletonSocket"),
		filepath.Join(homeDir, ".cache", "chrome-devtools-mcp", "chrome-profile", "SingletonCookie"),
		filepath.Join(homeDir, ".cache", "puppeteer", "chrome-profile", "SingletonLock"),
		filepath.Join(homeDir, ".cache", "playwright", "chrome-profile", "SingletonLock"),
	}

	for _, lockPath := range lockPaths {
		if _, err := os.Lstat(lockPath); err == nil {
			if err := os.Remove(lockPath); err != nil {
				log.Printf("[BROWSER] Warning: failed to remove lock %s: %v", lockPath, err)
			} else {
				log.Printf("[BROWSER] Cleaned up stale lock: %s", lockPath)
			}
		}
	}
}

// killOrphanedChrome kills Chrome/Chromium processes that might be holding
// locks on the profile directory.
func killOrphanedChrome() {
	// Use pkill to find Chrome processes with the MCP profile directory
	patterns := []string{
		"chrome-devtools-mcp/chrome-profile",
	}

	for _, pattern := range patterns {
		cmd := exec.Command("pkill", "-f", pattern)
		if err := cmd.Run(); err == nil {
			log.Printf("[BROWSER] Killed orphaned Chrome processes matching: %s", pattern)
			// Give processes time to die
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// CleanupOnStartup should be called during daemon initialization to
// clean up stale Chrome state from previous sessions.
func CleanupBrowserOnStartup() {
	cleanupChromeLocks()
	killOrphanedChrome()
	log.Printf("[BROWSER] Startup cleanup complete")
}

// EnsureBrowserReady checks if Chrome is available with remote debugging on port 9222
// using the default profile path. Delegates to EnsureBrowserReadyWithConfig.
func EnsureBrowserReady() bool {
	homeDir, _ := os.UserHomeDir()
	defaultProfile := filepath.Join(homeDir, ".cache", "chrome-devtools-mcp", "chrome-profile")
	return EnsureBrowserReadyWithConfig(defaultProfile, 9222)
}

// EnsureBrowserReadyWithConfig checks if Chrome is available with remote debugging
// on the given port. If Chrome is not running, it auto-launches Chrome with the
// specified profile path and port. Launched WITHOUT --enable-automation to avoid
// CAPTCHA triggers. Returns true if the port is accepting connections.
func EnsureBrowserReadyWithConfig(profilePath string, port int) bool {
	if IsPortOpen("127.0.0.1", port) {
		log.Printf("[BROWSER] Chrome already available with remote debugging on port %d", port)
		return true
	}

	if IsChromeRunning() {
		log.Printf("[BROWSER] Chrome is running but without remote debugging port %d", port)
		return false
	}

	chromePath := FindChromeBinary()
	if chromePath == "" {
		log.Printf("[BROWSER] Chrome binary not found, cannot auto-launch")
		return false
	}

	os.MkdirAll(profilePath, 0755)

	log.Printf("[BROWSER] Auto-launching Chrome with remote debugging on port %d...", port)
	cmd := exec.Command(chromePath,
		fmt.Sprintf("--remote-debugging-port=%d", port),
		"--user-data-dir="+profilePath,
		"--no-first-run",
		"--no-default-browser-check",
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		log.Printf("[BROWSER] Failed to launch Chrome: %v", err)
		return false
	}

	launchedChromeCmd = cmd
	go cmd.Wait()

	for i := 0; i < 40; i++ {
		time.Sleep(250 * time.Millisecond)
		if IsPortOpen("127.0.0.1", port) {
			log.Printf("[BROWSER] Chrome launched with remote debugging on port %d (PID %d)", port, cmd.Process.Pid)
			return true
		}
	}

	log.Printf("[BROWSER] Chrome launched but port %d not responding after 10s", port)
	return false
}

// IsPortOpen checks if a TCP port is accepting connections.
func IsPortOpen(host string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// IsChromeRunning checks if any Chrome process is already running.
func IsChromeRunning() bool {
	out, _ := exec.Command("pgrep", "-x", "chrome").Output()
	return len(strings.TrimSpace(string(out))) > 0
}

// FindChromeBinary locates the Chrome/Chromium executable on the system.
func FindChromeBinary() string {
	// PATH-based lookup (works on both Linux and macOS if Chrome is on PATH)
	pathCandidates := []string{
		"google-chrome-stable",
		"google-chrome",
		"chromium-browser",
		"chromium",
	}
	for _, name := range pathCandidates {
		path, err := exec.LookPath(name)
		if err == nil {
			return path
		}
	}

	// macOS: Chrome is an .app bundle, not typically on PATH
	if runtime.GOOS == "darwin" {
		macPaths := []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
		}
		for _, p := range macPaths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}

	return ""
}

// GetChromeWSEndpoint fetches the CDP WebSocket debugger URL from Chrome's
// /json/version endpoint. Returns empty string on failure.
func GetChromeWSEndpoint(port int) string {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/json/version", port))
	if err != nil {
		log.Printf("[BROWSER] Failed to fetch /json/version: %v", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[BROWSER] Failed to read /json/version response: %v", err)
		return ""
	}

	var info struct {
		WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		log.Printf("[BROWSER] Failed to parse /json/version: %v", err)
		return ""
	}

	if info.WebSocketDebuggerURL == "" {
		log.Printf("[BROWSER] /json/version returned empty webSocketDebuggerUrl")
		return ""
	}

	return info.WebSocketDebuggerURL
}

// WaitForChromeReady polls Chrome's debug port until it responds or timeout is reached.
// Returns the WebSocket endpoint if Chrome becomes ready, empty string on timeout.
func WaitForChromeReady(port int, timeout time.Duration) string {
	deadline := time.Now().Add(timeout)
	attempt := 0
	for time.Now().Before(deadline) {
		attempt++
		if ws := GetChromeWSEndpoint(port); ws != "" {
			log.Printf("[BROWSER] Chrome ready after %d attempts on port %d", attempt, port)
			return ws
		}
		time.Sleep(500 * time.Millisecond)
	}
	log.Printf("[BROWSER] Chrome not ready after %v on port %d (%d attempts)", timeout, port, attempt)
	return ""
}

// IsBrowserMCPServer checks if a server config looks like a browser/chrome-devtools MCP server.
func IsBrowserMCPServer(name string, args []string) bool {
	if strings.Contains(strings.ToLower(name), "chrome") || strings.Contains(strings.ToLower(name), "browser") {
		return true
	}
	for _, arg := range args {
		if strings.Contains(arg, "chrome-devtools-mcp") {
			return true
		}
	}
	return false
}

// InjectBrowserEndpoint returns a copy of the args with --wsEndpoint injected if Chrome is available.
// If the args already contain --wsEndpoint or --browserUrl, returns them unchanged.
// Falls back to --browserUrl if WebSocket endpoint can't be fetched.
func InjectBrowserEndpoint(args []string, port int) []string {
	for _, arg := range args {
		if strings.Contains(arg, "--wsEndpoint") || strings.Contains(arg, "--browserUrl") {
			return args // already configured
		}
	}

	wsEndpoint := GetChromeWSEndpoint(port)
	if wsEndpoint != "" {
		log.Printf("[BROWSER] Injecting --wsEndpoint (avoids profile lock): %s", wsEndpoint)
		return append(args, "--wsEndpoint", wsEndpoint)
	}

	// Fallback to --browserUrl (stable, works even if /json/version fails)
	browserURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	log.Printf("[BROWSER] Injecting --browserUrl fallback: %s", browserURL)
	return append(args, "--browserUrl", browserURL)
}
