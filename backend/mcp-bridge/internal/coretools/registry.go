package coretools

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/claraverse/mcp-client/internal/registry"
)

// UnifiedRegistry provides a single interface for executing both built-in
// core tools and MCP server tools. Core tools are checked first.
type UnifiedRegistry struct {
	builtinTools map[string]Tool
	mcpRegistry  *registry.Registry
	mu           sync.RWMutex
}

// NewUnifiedRegistry creates a registry with all core tools + MCP tools.
func NewUnifiedRegistry(mcpRegistry *registry.Registry) *UnifiedRegistry {
	r := &UnifiedRegistry{
		builtinTools: make(map[string]Tool),
		mcpRegistry:  mcpRegistry,
	}

	// Register all core tools
	r.Register(NewExecTool())                // execute_bash
	r.Register(NewReadTool())                // read_file
	r.Register(NewWriteTool())               // write_file
	r.Register(NewEditTool())                // string_replace
	r.Register(NewBrowserTool(mcpRegistry))  // browser
	r.Register(NewFindFilesTool())           // find_files
	r.Register(NewGrepTool())                // grep
	r.Register(NewListDirTool())             // list_directory
	r.Register(NewDeviceInfoTool())          // get_device_info
	r.Register(NewRunBackgroundTool())       // run_background
	r.Register(NewHTTPRequestTool())         // http_request

	return r
}

// Register adds a built-in tool.
func (r *UnifiedRegistry) Register(tool Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.builtinTools[tool.Name()] = tool
}

// GetAllToolDefs returns tool definitions for ALL available tools (core + MCP).
// These are passed to the LLM so it knows what tools are available.
func (r *UnifiedRegistry) GetAllToolDefs() []ToolDef {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var defs []ToolDef

	// Core tools first
	for _, tool := range r.builtinTools {
		defs = append(defs, ToToolDef(tool))
	}

	// MCP tools (from all running servers)
	if r.mcpRegistry != nil {
		mcpTools := r.mcpRegistry.GetAllTools()
		for _, t := range mcpTools {
			name, _ := t["name"].(string)
			desc, _ := t["description"].(string)
			params, _ := t["parameters"].(map[string]interface{})

			if name == "" {
				continue
			}

			// Skip MCP tools that shadow core tools
			if _, isCore := r.builtinTools[name]; isCore {
				continue
			}

			defs = append(defs, ToolDef{
				Type: "function",
				Function: ToolFuncDef{
					Name:        name,
					Description: desc,
					Parameters:  params,
				},
			})
		}
	}

	return defs
}

// Execute runs a tool by name. Checks core tools first, then MCP tools.
func (r *UnifiedRegistry) Execute(ctx context.Context, toolName string, args map[string]interface{}) (string, error) {
	r.mu.RLock()
	coreTool, isCore := r.builtinTools[toolName]
	r.mu.RUnlock()

	if isCore {
		log.Printf("[TOOLS] Executing core tool: %s", toolName)
		return coreTool.Execute(ctx, args)
	}

	// Fall through to MCP registry
	if r.mcpRegistry != nil {
		log.Printf("[TOOLS] Executing MCP tool: %s", toolName)
		return r.mcpRegistry.ExecuteTool(toolName, args)
	}

	return "", fmt.Errorf("tool %q not found in core tools or MCP servers", toolName)
}

// IsCoreTool returns true if the tool is a built-in core tool.
func (r *UnifiedRegistry) IsCoreTool(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.builtinTools[name]
	return ok
}

// CoreToolCount returns the number of registered core tools.
func (r *UnifiedRegistry) CoreToolCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.builtinTools)
}

// GetCoreToolDefs returns only core tool definitions (no MCP tools).
func (r *UnifiedRegistry) GetCoreToolDefs() []ToolDef {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var defs []ToolDef
	for _, tool := range r.builtinTools {
		defs = append(defs, ToToolDef(tool))
	}
	return defs
}

// GetCoreAndBrowserToolDefs returns core tools + browser MCP tools.
func (r *UnifiedRegistry) GetCoreAndBrowserToolDefs() []ToolDef {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var defs []ToolDef
	for _, tool := range r.builtinTools {
		defs = append(defs, ToToolDef(tool))
	}

	// Add browser MCP tools (chrome-devtools)
	if r.mcpRegistry != nil {
		browserTools := []string{
			"navigate_page", "take_snapshot", "take_screenshot",
			"click", "fill", "press_key", "hover", "wait_for",
			"new_page", "list_pages", "select_page", "close_page",
			"evaluate_script",
		}
		browserSet := make(map[string]bool, len(browserTools))
		for _, t := range browserTools {
			browserSet[t] = true
		}

		mcpTools := r.mcpRegistry.GetAllTools()
		for _, t := range mcpTools {
			name, _ := t["name"].(string)
			if name == "" || !browserSet[name] {
				continue
			}
			if _, isCore := r.builtinTools[name]; isCore {
				continue
			}
			desc, _ := t["description"].(string)
			params, _ := t["parameters"].(map[string]interface{})
			defs = append(defs, ToolDef{
				Type: "function",
				Function: ToolFuncDef{
					Name:        name,
					Description: desc,
					Parameters:  params,
				},
			})
		}
	}

	return defs
}

// CleanupBackground stops all tracked background processes.
// Should be called during daemon shutdown.
func (r *UnifiedRegistry) CleanupBackground() {
	r.mu.RLock()
	tool, ok := r.builtinTools["run_background"]
	r.mu.RUnlock()

	if ok {
		if bgTool, isBg := tool.(*RunBackgroundTool); isBg {
			bgTool.Cleanup()
		}
	}
}

// GetCoreToolDefsAsMap returns core tool definitions as generic maps.
// Used by the bridge to register tools with the cloud backend.
func (r *UnifiedRegistry) GetCoreToolDefsAsMap() []map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var defs []map[string]interface{}
	for _, tool := range r.builtinTools {
		defs = append(defs, ToToolDefMap(tool))
	}
	return defs
}
