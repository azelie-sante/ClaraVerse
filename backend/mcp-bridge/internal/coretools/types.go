package coretools

import "context"

// Tool is the interface all core tools implement.
type Tool interface {
	// Name returns the tool name (e.g. "execute_bash", "read_file").
	Name() string

	// Description returns a human-readable description for the LLM.
	Description() string

	// InputSchema returns the JSON Schema for the tool's parameters.
	InputSchema() map[string]interface{}

	// Execute runs the tool with the given arguments.
	Execute(ctx context.Context, args map[string]interface{}) (string, error)
}

// ToolDef represents a tool definition for the LLM (OpenAI format).
type ToolDef struct {
	Type     string      `json:"type"` // "function"
	Function ToolFuncDef `json:"function"`
}

// ToolFuncDef represents a function definition within a ToolDef.
type ToolFuncDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToToolDef converts a Tool to a ToolDef for passing to the LLM.
func ToToolDef(t Tool) ToolDef {
	return ToolDef{
		Type: "function",
		Function: ToolFuncDef{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  t.InputSchema(),
		},
	}
}

// ToToolDefMap converts a Tool to a generic map (for bridge tool registration).
func ToToolDefMap(t Tool) map[string]interface{} {
	return map[string]interface{}{
		"name":        t.Name(),
		"description": t.Description(),
		"parameters":  t.InputSchema(),
	}
}
