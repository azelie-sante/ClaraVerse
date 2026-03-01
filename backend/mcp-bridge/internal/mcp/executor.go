package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      int                    `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      int                    `json:"id"`
	Result  map[string]interface{} `json:"result,omitempty"`
	Error   *JSONRPCError          `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Tool represents an MCP tool definition
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// Executor manages communication with an MCP server
type Executor struct {
	serverPath string
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	stderr     io.ReadCloser
	reader     *bufio.Reader
	writer     *bufio.Writer
	requestID  int
	mutex      sync.Mutex
	verbose    bool
}

// NewExecutor creates a new MCP executor for a stdio server (path-based)
func NewExecutor(serverPath string, verbose bool) (*Executor, error) {
	return NewExecutorWithCommand("", serverPath, nil, verbose)
}

// NewExecutorWithCommand creates a new MCP executor with command and args support
func NewExecutorWithCommand(name, command string, args []string, verbose bool) (*Executor, error) {
	return NewExecutorWithContext(context.Background(), name, command, args, verbose)
}

// NewExecutorWithContext creates a new MCP executor with context for cancellation support
func NewExecutorWithContext(ctx context.Context, name, command string, args []string, verbose bool) (*Executor, error) {
	// Check if already cancelled
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var cmd *exec.Cmd

	if len(args) > 0 {
		// Command with arguments (e.g., npx @browsermcp/mcp@latest)
		cmd = exec.CommandContext(ctx, command, args...)
		if verbose {
			log.Printf("[MCP] Starting server with command: %s %v", command, args)
		}
	} else {
		// Single executable path
		cmd = exec.CommandContext(ctx, command)
		if verbose {
			log.Printf("[MCP] Starting server: %s", command)
		}
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the server
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	executor := &Executor{
		serverPath: command,
		cmd:        cmd,
		stdin:      stdin,
		stdout:     stdout,
		stderr:     stderr,
		reader:     bufio.NewReader(stdout),
		writer:     bufio.NewWriter(stdin),
		requestID:  0,
		verbose:    verbose,
	}

	// Start stderr reader
	go executor.readStderr()

	// Initialize the server
	if err := executor.initialize(); err != nil {
		executor.Close()
		return nil, fmt.Errorf("failed to initialize server: %w", err)
	}

	return executor, nil
}

// readStderr logs stderr output
func (e *Executor) readStderr() {
	scanner := bufio.NewScanner(e.stderr)
	for scanner.Scan() {
		if e.verbose {
			log.Printf("[MCP stderr] %s", scanner.Text())
		}
	}
}

// initialize sends the initialize request to the MCP server
func (e *Executor) initialize() error {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      e.nextID(),
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"roots": map[string]interface{}{
					"listChanged": true,
				},
			},
			"clientInfo": map[string]interface{}{
				"name":    "clara-companion",
				"version": "1.0.0",
			},
		},
	}

	resp, err := e.sendRequest(req)
	if err != nil {
		return fmt.Errorf("initialize failed: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("initialize error: %s", resp.Error.Message)
	}

	if e.verbose {
		log.Printf("[MCP] Server initialized")
	}

	return nil
}

// ListTools retrieves all available tools from the MCP server
func (e *Executor) ListTools() ([]Tool, error) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      e.nextID(),
		Method:  "tools/list",
	}

	resp, err := e.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("tools/list failed: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("tools/list error: %s", resp.Error.Message)
	}

	// Parse tools from result
	toolsData, ok := resp.Result["tools"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid tools response format")
	}

	var tools []Tool
	for _, t := range toolsData {
		toolMap, ok := t.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := toolMap["name"].(string)
		description, _ := toolMap["description"].(string)
		if name == "" {
			continue
		}
		tool := Tool{
			Name:        name,
			Description: description,
		}

		if schema, ok := toolMap["inputSchema"].(map[string]interface{}); ok {
			tool.InputSchema = schema
		}

		tools = append(tools, tool)
	}

	return tools, nil
}

// CallTool executes a tool on the MCP server
func (e *Executor) CallTool(toolName string, arguments map[string]interface{}) (string, error) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      e.nextID(),
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": arguments,
		},
	}

	resp, err := e.sendRequest(req)
	if err != nil {
		return "", fmt.Errorf("tools/call failed: %w", err)
	}

	if resp.Error != nil {
		return "", fmt.Errorf("tool error: %s", resp.Error.Message)
	}

	// Extract result content
	content, ok := resp.Result["content"].([]interface{})
	if !ok || len(content) == 0 {
		// Some tools (like browser_navigate) return empty content on success
		// Check if there's any result data to return
		if len(resp.Result) > 0 {
			// Try to return a JSON representation of the result
			resultJSON, err := json.Marshal(resp.Result)
			if err == nil && string(resultJSON) != "{}" {
				return string(resultJSON), nil
			}
		}
		// No content but no error = successful side-effect operation
		return "Tool executed successfully", nil
	}

	// Get text from first content item
	firstContent, ok := content[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid content format")
	}

	text, ok := firstContent["text"].(string)
	if !ok {
		// Try to get other content types
		if data, hasData := firstContent["data"]; hasData {
			if str, isStr := data.(string); isStr {
				return str, nil
			}
		}
		return "", fmt.Errorf("no text in content")
	}

	return text, nil
}

// sendRequest sends a JSON-RPC request and waits for response
func (e *Executor) sendRequest(req JSONRPCRequest) (*JSONRPCResponse, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Marshal request
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	if e.verbose {
		log.Printf("[MCP →] %s", string(data))
	}

	// Write request
	if _, err := e.writer.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}
	if _, err := e.writer.WriteString("\n"); err != nil {
		return nil, fmt.Errorf("failed to write newline: %w", err)
	}
	if err := e.writer.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush: %w", err)
	}

	// Read response - skip non-JSON lines (logs, etc.)
	var resp JSONRPCResponse
	maxAttempts := 100 // Prevent infinite loop
	for attempt := 0; attempt < maxAttempts; attempt++ {
		line, err := e.reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue // Skip empty lines
		}

		if e.verbose {
			log.Printf("[MCP ←] %s", line)
		}

		// Try to parse as JSON-RPC response
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			// Not valid JSON-RPC, might be a log line - skip it
			if e.verbose {
				log.Printf("[MCP] Skipping non-JSON line: %s", line)
			}
			continue
		}

		// Successfully parsed JSON-RPC response
		return &resp, nil
	}

	return nil, fmt.Errorf("no valid JSON-RPC response found after %d lines", maxAttempts)
}

// nextID returns the next request ID
func (e *Executor) nextID() int {
	e.requestID++
	return e.requestID
}

// Close terminates the MCP server
func (e *Executor) Close() error {
	if e.stdin != nil {
		e.stdin.Close()
	}
	if e.stdout != nil {
		e.stdout.Close()
	}
	if e.stderr != nil {
		e.stderr.Close()
	}
	if e.cmd != nil && e.cmd.Process != nil {
		return e.cmd.Process.Kill()
	}
	return nil
}
