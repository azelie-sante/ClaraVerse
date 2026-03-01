package coretools

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

// BackgroundProcess tracks a process started in the background.
type BackgroundProcess struct {
	PID       int
	Command   string
	StartedAt time.Time
	Port      int    // 0 if no port was specified
	Dir       string // working directory
	cmd       *exec.Cmd
	done      chan struct{} // closed when process exits
	exitCode  int
	exited    bool
}

// RunBackgroundTool starts long-running processes in the background (like dev servers),
// lists tracked processes, and stops them by PID.
type RunBackgroundTool struct {
	mu        sync.Mutex
	processes map[int]*BackgroundProcess
}

func NewRunBackgroundTool() *RunBackgroundTool {
	return &RunBackgroundTool{
		processes: make(map[int]*BackgroundProcess),
	}
}

func (t *RunBackgroundTool) Name() string { return "run_background" }

func (t *RunBackgroundTool) Description() string {
	return "Manage long-running background processes like dev servers, watchers, or build daemons.\n\n" +
		"Actions:\n" +
		"- start: Launch a command in the background and return immediately with its PID. " +
		"Optionally waits for a TCP port to become available before returning, so you know the " +
		"server is ready. The process runs independently and survives tool call completion.\n" +
		"- stop: Terminate a background process by PID. Sends SIGTERM first, then SIGKILL " +
		"after 5 seconds if the process hasn't exited.\n" +
		"- list: Show all tracked background processes with their PID, command, port, " +
		"uptime, and whether they're still running.\n\n" +
		"Common use cases:\n" +
		"- Start a React dev server: action=start, command='npm run dev', wait_port=5173\n" +
		"- Start a backend API: action=start, command='python manage.py runserver', wait_port=8000\n" +
		"- Check what's running: action=list\n" +
		"- Stop a server when done: action=stop, pid=<PID from start>\n\n" +
		"Processes are tracked in memory. If the companion restarts, tracking is lost " +
		"but the processes may still be running (check with execute_bash 'ps aux')."
}

func (t *RunBackgroundTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "The action to perform: 'start' to launch a background process, 'stop' to terminate one, 'list' to show all tracked processes.",
				"enum":        []string{"start", "stop", "list"},
			},
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The shell command to run in the background (for 'start' action). Runs via bash -c. Example: 'npm run dev', 'python -m http.server 8080'.",
			},
			"working_dir": map[string]interface{}{
				"type":        "string",
				"description": "Working directory for the command (for 'start' action). Defaults to the user's home directory.",
			},
			"wait_port": map[string]interface{}{
				"type":        "integer",
				"description": "TCP port to wait for after starting the process (for 'start' action). The tool polls until this port accepts connections, then returns. Timeout: 60 seconds. Omit if the process doesn't listen on a port.",
			},
			"pid": map[string]interface{}{
				"type":        "integer",
				"description": "Process ID to stop (for 'stop' action). Get this from the 'start' or 'list' action.",
			},
		},
		"required": []string{"action"},
	}
}

func (t *RunBackgroundTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	action, ok := args["action"].(string)
	if !ok || action == "" {
		return "", fmt.Errorf("action is required (start, stop, or list)")
	}

	switch action {
	case "start":
		return t.start(args)
	case "stop":
		return t.stop(args)
	case "list":
		return t.list(), nil
	default:
		return "", fmt.Errorf("unknown action: %s (use start, stop, or list)", action)
	}
}

func (t *RunBackgroundTool) start(args map[string]interface{}) (string, error) {
	command, ok := args["command"].(string)
	if !ok || command == "" {
		return "", fmt.Errorf("command is required for start action")
	}

	// Check for dangerous commands
	if reason := isDangerousCommand(command); reason != "" {
		return "", fmt.Errorf("command blocked: %s", reason)
	}

	// Auto-strip sudo when running as root
	if os.Getuid() == 0 {
		command = stripSudo(command)
	}

	workingDir := ""
	if wd, ok := args["working_dir"].(string); ok && wd != "" {
		workingDir = wd
	}

	waitPort := 0
	if wp, ok := args["wait_port"].(float64); ok && wp > 0 {
		waitPort = int(wp)
	}

	// Start the process
	cmd := exec.Command("bash", "-c", command)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	// Detach from process group so it survives if the companion exits
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Discard stdout/stderr to prevent pipe blocking
	// The process runs independently — logs go to its own stdout
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start process: %w", err)
	}

	pid := cmd.Process.Pid

	// Track the process
	bp := &BackgroundProcess{
		PID:       pid,
		Command:   command,
		StartedAt: time.Now(),
		Port:      waitPort,
		Dir:       workingDir,
		cmd:       cmd,
		done:      make(chan struct{}),
	}

	// Monitor process exit in background
	go func() {
		err := cmd.Wait()
		t.mu.Lock()
		bp.exited = true
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				bp.exitCode = exitErr.ExitCode()
			} else {
				bp.exitCode = -1
			}
		}
		t.mu.Unlock()
		close(bp.done)
	}()

	t.mu.Lock()
	t.processes[pid] = bp
	t.mu.Unlock()

	var result strings.Builder
	fmt.Fprintf(&result, "Process started successfully.\n")
	fmt.Fprintf(&result, "PID: %d\n", pid)
	fmt.Fprintf(&result, "Command: %s\n", command)

	// Wait for port if specified
	if waitPort > 0 {
		fmt.Fprintf(&result, "Waiting for port %d...\n", waitPort)

		ready := t.waitForPort(waitPort, 60*time.Second, bp.done)
		if ready {
			fmt.Fprintf(&result, "Port %d is ready — server is accepting connections.\n", waitPort)
		} else {
			// Check if process died
			if bp.exited {
				fmt.Fprintf(&result, "WARNING: Process exited (code %d) before port %d became available.\n", bp.exitCode, waitPort)
				fmt.Fprintf(&result, "The server may have failed to start. Check logs or run the command manually.\n")
			} else {
				fmt.Fprintf(&result, "WARNING: Port %d did not become available within 60 seconds.\n", waitPort)
				fmt.Fprintf(&result, "The process is still running (PID %d). It may need more time to start, ", pid)
				fmt.Fprintf(&result, "or it may be listening on a different port.\n")
			}
		}
	}

	return result.String(), nil
}

func (t *RunBackgroundTool) stop(args map[string]interface{}) (string, error) {
	pidFloat, ok := args["pid"].(float64)
	if !ok || pidFloat <= 0 {
		return "", fmt.Errorf("pid is required for stop action")
	}
	pid := int(pidFloat)

	t.mu.Lock()
	bp, tracked := t.processes[pid]
	t.mu.Unlock()

	if tracked && bp.exited {
		t.mu.Lock()
		delete(t.processes, pid)
		t.mu.Unlock()
		return fmt.Sprintf("Process %d has already exited (code %d).", pid, bp.exitCode), nil
	}

	// Try SIGTERM first (graceful shutdown)
	proc, err := os.FindProcess(pid)
	if err != nil {
		return "", fmt.Errorf("cannot find process %d: %w", pid, err)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		// Process may already be gone
		if tracked {
			t.mu.Lock()
			delete(t.processes, pid)
			t.mu.Unlock()
		}
		return fmt.Sprintf("Process %d is not running (may have already exited).", pid), nil
	}

	// Wait up to 5 seconds for graceful exit
	exited := false
	if tracked {
		select {
		case <-bp.done:
			exited = true
		case <-time.After(5 * time.Second):
		}
	} else {
		time.Sleep(1 * time.Second)
		if err := proc.Signal(syscall.Signal(0)); err != nil {
			exited = true
		}
	}

	if !exited {
		// Force kill
		proc.Signal(syscall.SIGKILL)
		if tracked {
			select {
			case <-bp.done:
			case <-time.After(3 * time.Second):
			}
		}
	}

	// Also kill the entire process group (catches child processes like node)
	syscall.Kill(-pid, syscall.SIGKILL)

	if tracked {
		t.mu.Lock()
		delete(t.processes, pid)
		t.mu.Unlock()
	}

	return fmt.Sprintf("Process %d stopped.", pid), nil
}

func (t *RunBackgroundTool) list() string {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.processes) == 0 {
		return "No tracked background processes."
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Tracked background processes (%d):\n\n", len(t.processes))

	for _, bp := range t.processes {
		status := "running"
		if bp.exited {
			status = fmt.Sprintf("exited (code %d)", bp.exitCode)
		}

		uptime := time.Since(bp.StartedAt).Truncate(time.Second)
		fmt.Fprintf(&sb, "  PID: %d\n", bp.PID)
		fmt.Fprintf(&sb, "  Command: %s\n", bp.Command)
		fmt.Fprintf(&sb, "  Status: %s\n", status)
		fmt.Fprintf(&sb, "  Uptime: %s\n", uptime)
		if bp.Port > 0 {
			fmt.Fprintf(&sb, "  Port: %d\n", bp.Port)
		}
		if bp.Dir != "" {
			fmt.Fprintf(&sb, "  Dir: %s\n", bp.Dir)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// waitForPort polls a TCP port until it accepts connections or timeout/process death.
func (t *RunBackgroundTool) waitForPort(port int, timeout time.Duration, processDone <-chan struct{}) bool {
	deadline := time.After(timeout)
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	for {
		select {
		case <-deadline:
			return false
		case <-processDone:
			// Process exited before port was ready
			return false
		default:
			conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
			if err == nil {
				conn.Close()
				return true
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// Cleanup stops all tracked background processes. Called during daemon shutdown.
func (t *RunBackgroundTool) Cleanup() {
	t.mu.Lock()
	pids := make([]int, 0, len(t.processes))
	for pid := range t.processes {
		pids = append(pids, pid)
	}
	t.mu.Unlock()

	for _, pid := range pids {
		t.stop(map[string]interface{}{
			"pid": float64(pid),
		})
	}
}
