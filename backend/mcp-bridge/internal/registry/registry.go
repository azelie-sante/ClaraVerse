package registry

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/claraverse/mcp-client/internal/config"
	"github.com/claraverse/mcp-client/internal/mcp"
)

// ServerInstance represents a running MCP server
type ServerInstance struct {
	Config   config.MCPServer
	Executor *mcp.Executor
	Tools    []mcp.Tool
}

// Registry manages all MCP server instances
type Registry struct {
	servers map[string]*ServerInstance
	mutex   sync.RWMutex
	verbose bool
}

// NewRegistry creates a new server registry
func NewRegistry(verbose bool) *Registry {
	return &Registry{
		servers: make(map[string]*ServerInstance),
		verbose: verbose,
	}
}

// StartServer starts an MCP server
func (r *Registry) StartServer(cfg config.MCPServer) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if already running
	if _, exists := r.servers[cfg.Name]; exists {
		return fmt.Errorf("server %s is already running", cfg.Name)
	}

	// Only support stdio for now
	if cfg.Type != "stdio" {
		return fmt.Errorf("only stdio servers are supported (server %s uses %s)", cfg.Name, cfg.Type)
	}

	log.Printf("🚀 Starting MCP server: %s", cfg.Name)

	// Create executor - check if command-based or path-based
	var executor *mcp.Executor
	var err error

	if cfg.Command != "" {
		// Command-based server (e.g., npx @browsermcp/mcp@latest)
		executor, err = mcp.NewExecutorWithCommand(cfg.Name, cfg.Command, cfg.Args, r.verbose)
	} else if cfg.Path != "" {
		// Path-based server (e.g., /path/to/server.exe)
		executor, err = mcp.NewExecutor(cfg.Path, r.verbose)
	} else {
		return fmt.Errorf("server %s must have either 'path' or 'command' configured", cfg.Name)
	}

	if err != nil {
		return fmt.Errorf("failed to start server %s: %w", cfg.Name, err)
	}

	// List tools
	tools, err := executor.ListTools()
	if err != nil {
		executor.Close()
		return fmt.Errorf("failed to list tools from %s: %w", cfg.Name, err)
	}

	instance := &ServerInstance{
		Config:   cfg,
		Executor: executor,
		Tools:    tools,
	}

	r.servers[cfg.Name] = instance

	log.Printf("✅ Server %s started with %d tools", cfg.Name, len(tools))
	for _, tool := range tools {
		log.Printf("   - %s: %s", tool.Name, tool.Description)
	}

	return nil
}

// StartServerWithContext starts an MCP server with context for cancellation
func (r *Registry) StartServerWithContext(ctx context.Context, cfg config.MCPServer) error {
	// Check if already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if already running
	if _, exists := r.servers[cfg.Name]; exists {
		return fmt.Errorf("server %s is already running", cfg.Name)
	}

	// Only support stdio for now
	if cfg.Type != "stdio" {
		return fmt.Errorf("only stdio servers are supported (server %s uses %s)", cfg.Name, cfg.Type)
	}

	log.Printf("🚀 Starting MCP server: %s", cfg.Name)

	// Check context again before starting
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create executor with context
	var executor *mcp.Executor
	var err error

	if cfg.Command != "" {
		executor, err = mcp.NewExecutorWithContext(ctx, cfg.Name, cfg.Command, cfg.Args, r.verbose)
	} else if cfg.Path != "" {
		// Path-based uses background context (legacy behavior)
		executor, err = mcp.NewExecutor(cfg.Path, r.verbose)
	} else {
		return fmt.Errorf("server %s must have either 'path' or 'command' configured", cfg.Name)
	}

	if err != nil {
		return fmt.Errorf("failed to start server %s: %w", cfg.Name, err)
	}

	// Check context before listing tools
	select {
	case <-ctx.Done():
		executor.Close()
		return ctx.Err()
	default:
	}

	// List tools
	tools, err := executor.ListTools()
	if err != nil {
		executor.Close()
		return fmt.Errorf("failed to list tools from %s: %w", cfg.Name, err)
	}

	instance := &ServerInstance{
		Config:   cfg,
		Executor: executor,
		Tools:    tools,
	}

	r.servers[cfg.Name] = instance

	log.Printf("✅ Server %s started with %d tools", cfg.Name, len(tools))
	for _, tool := range tools {
		log.Printf("   - %s: %s", tool.Name, tool.Description)
	}

	return nil
}

// StopServer stops an MCP server
func (r *Registry) StopServer(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	instance, exists := r.servers[name]
	if !exists {
		return fmt.Errorf("server %s is not running", name)
	}

	log.Printf("🛑 Stopping MCP server: %s", name)

	if err := instance.Executor.Close(); err != nil {
		log.Printf("Warning: error closing executor for %s: %v", name, err)
	}

	delete(r.servers, name)

	log.Printf("✅ Server %s stopped", name)
	return nil
}

// RestartServer stops a server and starts it again with its original config.
// Returns error if the server was not running or failed to restart.
func (r *Registry) RestartServer(name string) error {
	r.mutex.Lock()

	instance, exists := r.servers[name]
	if !exists {
		r.mutex.Unlock()
		return fmt.Errorf("server %s is not running", name)
	}

	cfg := instance.Config
	log.Printf("🔄 Restarting MCP server: %s", name)

	// Stop
	if err := instance.Executor.Close(); err != nil {
		log.Printf("Warning: error closing executor for %s during restart: %v", name, err)
	}
	delete(r.servers, name)
	r.mutex.Unlock()

	// Brief pause for process cleanup
	time.Sleep(500 * time.Millisecond)

	// Start fresh (StartServer acquires its own lock)
	return r.StartServer(cfg)
}

// StopAll stops all running servers
func (r *Registry) StopAll() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for name, instance := range r.servers {
		log.Printf("🛑 Stopping server: %s", name)
		instance.Executor.Close()
	}

	r.servers = make(map[string]*ServerInstance)
}

// GetAllTools returns all tools from all running servers
func (r *Registry) GetAllTools() []map[string]interface{} {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var allTools []map[string]interface{}

	for serverName, instance := range r.servers {
		for _, tool := range instance.Tools {
			// Convert MCP tool to OpenAI format
			toolDef := map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.InputSchema,
				"server_name": serverName,
			}
			allTools = append(allTools, toolDef)
		}
	}

	return allTools
}

// ExecuteTool executes a tool by finding which server provides it
func (r *Registry) ExecuteTool(toolName string, arguments map[string]interface{}) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Find which server has this tool
	for serverName, instance := range r.servers {
		for _, tool := range instance.Tools {
			if tool.Name == toolName {
				log.Printf("🔧 Executing %s on server %s", toolName, serverName)
				return instance.Executor.CallTool(toolName, arguments)
			}
		}
	}

	return "", fmt.Errorf("tool %s not found in any running server", toolName)
}

// GetServerCount returns the number of running servers
func (r *Registry) GetServerCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.servers)
}

// GetToolCount returns the total number of tools across all servers
func (r *Registry) GetToolCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	count := 0
	for _, instance := range r.servers {
		count += len(instance.Tools)
	}
	return count
}

// GetServerNames returns names of all running servers
func (r *Registry) GetServerNames() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	names := make([]string, 0, len(r.servers))
	for name := range r.servers {
		names = append(names, name)
	}
	return names
}

// GetServer returns a server instance by name
func (r *Registry) GetServer(name string) (*ServerInstance, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	instance, exists := r.servers[name]
	if !exists {
		return nil, fmt.Errorf("server %s not found", name)
	}
	return instance, nil
}

// GetServerToolCount returns the number of tools for a specific server
func (r *Registry) GetServerToolCount(name string) int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	instance, exists := r.servers[name]
	if !exists {
		return 0
	}
	return len(instance.Tools)
}

// GetServerForTool returns the server name that provides a specific tool
func (r *Registry) GetServerForTool(toolName string) string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for serverName, instance := range r.servers {
		for _, tool := range instance.Tools {
			if tool.Name == toolName {
				return serverName
			}
		}
	}
	return ""
}
