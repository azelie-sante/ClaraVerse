package execution

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"claraverse/internal/models"
	"claraverse/internal/services"
	"claraverse/internal/tools"
)

// CrewExecAdapter lets the Crew worker reuse the battle-tested headless
// tool-loop (AgentBlockExecutor) WITHOUT the DAG: one synthetic llm_inference
// block per card run. Implements services.CrewExecutor.
type CrewExecAdapter struct {
	inner *AgentBlockExecutor
}

func NewCrewExecAdapter(
	chatService *services.ChatService,
	providerService *services.ProviderService,
	toolRegistry *tools.Registry,
	credentialService *services.CredentialService,
) *CrewExecAdapter {
	return &CrewExecAdapter{inner: NewAgentBlockExecutor(chatService, providerService, toolRegistry, credentialService)}
}

// RunCrewTask executes one agent pass with a hard tool allowlist. Also returns
// the ordered list of tools the run actually called (the card's audit trail).
func (a *CrewExecAdapter) RunCrewTask(ctx context.Context, userID, model, systemPrompt, userPrompt string, allowedTools []string) (string, []string, int64, int64, error) {
	cfg := map[string]any{
		"systemPrompt":     systemPrompt,
		"userPrompt":       userPrompt,
		"enabledTools":     toAnySlice(allowedTools),
		"maxToolCalls":     15,
		"requireToolUsage": false,
	}
	if model != "" {
		cfg["model"] = model
	}
	block := models.Block{
		ID:      "crew-task",
		Type:    "llm_inference",
		Name:    "crew-task",
		Config:  cfg,
		Timeout: 420, // seconds — long enough for research/tool-heavy passes
	}
	result, err := a.inner.Execute(ctx, block, map[string]any{"__user_id__": userID})
	if err != nil {
		return "", nil, 0, 0, err
	}
	var toolsUsed []string
	if calls, okC := result["toolCalls"].([]models.ToolCallRecord); okC {
		for _, tc := range calls {
			toolsUsed = append(toolsUsed, tc.Name)
		}
	}
	// IMPORTANT: use rawResponse — the model's ACTUAL final text. The executor's
	// "response"/"text" fields are DAG data-passing conveniences that surface
	// raw tool output (a search dump beat the model's synthesis on length and
	// shipped as the card deliverable — exactly what a reviewer must never see).
	out, _ := result["rawResponse"].(string)
	if out == "" {
		out, _ = result["response"].(string)
	}
	out = stripToolCallLeakage(out)
	var tin, tout int64
	switch tk := result["tokens"].(type) {
	case map[string]int:
		tin, tout = int64(tk["input"]), int64(tk["output"])
	case map[string]any:
		if v, ok := tk["input"].(int); ok {
			tin = int64(v)
		}
		if v, ok := tk["output"].(int); ok {
			tout = int64(v)
		}
	}
	if out == "" {
		return "", toolsUsed, tin, tout, fmt.Errorf("agent run produced no text output")
	}
	// The streaming path doesn't always report usage — estimate (chars/4) so
	// member budgets still bite. Real usage wins whenever it's present.
	if tin+tout == 0 {
		tin = int64(len(systemPrompt)+len(userPrompt)) / 4
		tout = int64(len(out)) / 4
	}
	return out, toolsUsed, tin, tout, nil
}

// stripToolCallLeakage removes raw tool-call token syntax (some models'
// <|tool_call...|> markers) that models sometimes emit as plain text —
// especially on tool-less passes. That markup must never reach a reviewer.
var toolLeakRe = regexp.MustCompile(`(?s)<\|tool_calls_section_begin\|>.*?(<\|tool_calls_section_end\|>|$)`)
var toolTokenRe = regexp.MustCompile(`<\|/?tool_[a-z_]*\|>`)

func stripToolCallLeakage(s string) string {
	s = toolLeakRe.ReplaceAllString(s, "")
	s = toolTokenRe.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

func toAnySlice(in []string) []any {
	out := make([]any, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}
