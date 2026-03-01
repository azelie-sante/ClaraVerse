package coretools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ExecTool executes shell commands with structured output and dangerous command blocking.
type ExecTool struct {
	DefaultTimeout time.Duration
}

func NewExecTool() *ExecTool {
	return &ExecTool{
		DefaultTimeout: 2 * time.Minute,
	}
}

func (t *ExecTool) Name() string { return "execute_bash" }

func (t *ExecTool) Description() string {
	return "Execute a terminal command on the user's local machine via bash. " +
		"Runs the command through `bash -c`, so full bash syntax is supported: " +
		"pipes (|), redirects (>, >>), command chaining (&&, ||, ;), brace expansion, " +
		"subshells, environment variables, and globs.\n\n" +
		"Returns structured output with EXIT_CODE, STDOUT, and STDERR sections. " +
		"Exit code 0 means success; non-zero indicates failure; -1 means timeout.\n\n" +
		"Use cases: running programs, git operations, package management (apt, brew, npm), " +
		"compiling/building code, starting/stopping services, file operations, " +
		"and any other terminal task.\n\n" +
		"Safety: dangerous patterns (rm -rf /, fork bombs, disk wipes) are blocked. " +
		"Output is capped at 30K characters with smart truncation (head + tail preserved). " +
		"Default timeout is 2 minutes, configurable up to 10 minutes.\n\n" +
		"If the companion is running with elevated privileges (root/sudo), " +
		"sudo prefixes in commands are automatically stripped since they are unnecessary."
}

func (t *ExecTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The terminal command to execute. Runs via `bash -c`, so full bash syntax works: pipes (cmd1 | cmd2), chaining (cmd1 && cmd2), redirects (> file), globs (*.txt), environment variables ($HOME), and subshells ($(cmd)). Do not wrap in quotes — pass the raw command string.",
			},
			"working_dir": map[string]interface{}{
				"type":        "string",
				"description": "Working directory for the command. Defaults to the user's home directory. Use absolute paths.",
			},
			"timeout_ms": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in milliseconds. Default: 120000 (2 min). Max: 600000 (10 min). Use longer timeouts for builds, installs, or large downloads.",
			},
		},
		"required": []string{"command"},
	}
}

func (t *ExecTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	command, ok := args["command"].(string)
	if !ok || command == "" {
		return "", fmt.Errorf("command is required")
	}

	// Auto-strip sudo when already running as root (UID 0).
	// When the companion runs with elevated privileges (e.g. via sudo),
	// sudo is unnecessary and can cause issues (e.g. prompting for password
	// in a non-interactive shell). Strip it transparently.
	if os.Getuid() == 0 {
		command = stripSudo(command)
	}

	// Block dangerous commands
	if reason := isDangerousCommand(command); reason != "" {
		return "", fmt.Errorf("command blocked: %s", reason)
	}

	timeout := t.DefaultTimeout
	if timeoutMs, ok := args["timeout_ms"].(float64); ok && timeoutMs > 0 {
		timeout = time.Duration(timeoutMs) * time.Millisecond
		if timeout > 10*time.Minute {
			timeout = 10 * time.Minute
		}
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", command)

	if workingDir, ok := args["working_dir"].(string); ok && workingDir != "" {
		cmd.Dir = workingDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Build structured output
	var result strings.Builder
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else if ctx.Err() == context.DeadlineExceeded {
			exitCode = -1
		} else {
			exitCode = -1
		}
	}

	fmt.Fprintf(&result, "EXIT_CODE: %d\n", exitCode)

	if ctx.Err() == context.DeadlineExceeded {
		fmt.Fprintf(&result, "TIMEOUT: Command timed out after %v\n", timeout)
	}

	if stdout.Len() > 0 {
		result.WriteString("\nSTDOUT:\n")
		stdoutStr := stdout.String()
		if len(stdoutStr) > OutputLimitChars {
			stdoutStr = TruncateSmartly(stdoutStr, OutputLimitChars)
		}
		result.WriteString(stdoutStr)
	}

	if stderr.Len() > 0 {
		result.WriteString("\nSTDERR:\n")
		stderrStr := stderr.String()
		if len(stderrStr) > OutputLimitChars {
			stderrStr = TruncateSmartly(stderrStr, OutputLimitChars)
		}
		result.WriteString(stderrStr)
	}

	if stdout.Len() == 0 && stderr.Len() == 0 {
		result.WriteString("\n(no output)")
	}

	return result.String(), nil
}

// stripSudo removes the sudo prefix from a command string.
// Handles common forms: "sudo cmd", "sudo -E cmd", "sudo -u user cmd", etc.
// For compound commands (pipes, &&, ;) it strips sudo from each segment.
func stripSudo(command string) string {
	command = strings.TrimSpace(command)
	if !strings.Contains(command, "sudo") {
		return command
	}

	// For compound commands, handle each segment independently
	// Split on common shell operators while preserving them
	segments := splitShellSegments(command)
	for i, seg := range segments {
		seg = strings.TrimSpace(seg)
		if strings.HasPrefix(seg, "sudo ") || seg == "sudo" {
			segments[i] = stripSingleSudo(seg)
		}
	}
	return strings.Join(segments, "")
}

// stripSingleSudo strips sudo and its flags from a single command (no pipes/&&).
func stripSingleSudo(cmd string) string {
	cmd = strings.TrimSpace(cmd)
	if !strings.HasPrefix(cmd, "sudo") {
		return cmd
	}

	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return cmd
	}
	if parts[0] != "sudo" {
		return cmd
	}

	// Skip "sudo" and its flags to find the actual command
	i := 1
	for i < len(parts) {
		p := parts[i]
		if p == "--" {
			// End of sudo flags
			i++
			break
		}
		if !strings.HasPrefix(p, "-") {
			// Not a flag — this is the command (unless it's a -u argument)
			break
		}
		// Flags that take an argument: -u, -g, -C, -D, -T
		if p == "-u" || p == "-g" || p == "-C" || p == "-D" || p == "-T" {
			i += 2 // skip flag + its argument
			continue
		}
		// Combined flags like -Eu or standalone like -E, -i, -s, -n, -H, -k
		i++
	}

	if i >= len(parts) {
		// "sudo" with only flags and no command (e.g. "sudo -i") — leave as-is
		// but we're root so just return a shell invocation
		return cmd
	}

	return strings.Join(parts[i:], " ")
}

// splitShellSegments splits a command on shell operators (&&, ||, ;, |)
// while preserving the operators as separate segments.
func splitShellSegments(cmd string) []string {
	var segments []string
	var current strings.Builder
	runes := []rune(cmd)

	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		switch ch {
		case '&':
			if i+1 < len(runes) && runes[i+1] == '&' {
				segments = append(segments, current.String())
				segments = append(segments, " && ")
				current.Reset()
				i++ // skip second &
				continue
			}
			current.WriteRune(ch)
		case '|':
			if i+1 < len(runes) && runes[i+1] == '|' {
				segments = append(segments, current.String())
				segments = append(segments, " || ")
				current.Reset()
				i++
				continue
			}
			segments = append(segments, current.String())
			segments = append(segments, " | ")
			current.Reset()
		case ';':
			segments = append(segments, current.String())
			segments = append(segments, "; ")
			current.Reset()
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		segments = append(segments, current.String())
	}

	return segments
}

// isDangerousCommand checks if a command matches known dangerous patterns.
// Returns the reason string if dangerous, empty string if safe.
// Normalizes whitespace so "rm  -rf  /" also matches.
func isDangerousCommand(command string) string {
	lower := strings.ToLower(strings.TrimSpace(command))
	// Collapse multiple spaces/tabs to single space to defeat whitespace injection
	normalized := strings.Join(strings.Fields(lower), " ")

	for _, pattern := range DangerousCommands {
		normalizedPattern := strings.Join(strings.Fields(strings.ToLower(pattern)), " ")
		if strings.Contains(normalized, normalizedPattern) {
			return fmt.Sprintf("contains dangerous pattern: %s", pattern)
		}
	}

	return ""
}
