package services

import "claraverse/internal/models"

// ToolStrategy selects how the per-turn tool set is assembled.
type ToolStrategy int

const (
	// ToolsFull uses the matched skill's tools, otherwise the tool predictor,
	// otherwise every credential-filtered tool.
	ToolsFull ToolStrategy = iota
	// ToolsEssentialsOnly uses the matched skill's tools, otherwise the
	// essential set — never the predictor, so a turn costs one inference.
	ToolsEssentialsOnly
)

// TurnPolicy holds the per-turn orchestration decisions derived from the
// selected model. Its zero value is the full-featured path, so any caller
// that cannot resolve a model gets today's behaviour.
type TurnPolicy struct {
	SkipMemorySelection bool
	SkipSkillPrompt     bool
	ToolStrategy        ToolStrategy
	LiteSystemPrompt    bool
}

// resolveTurnPolicy derives the policy for a turn. Pure: no I/O.
func resolveTurnPolicy(m *models.Model) TurnPolicy {
	if m == nil || !m.LiteMode {
		return TurnPolicy{}
	}
	return TurnPolicy{
		SkipMemorySelection: true,
		SkipSkillPrompt:     true,
		ToolStrategy:        ToolsEssentialsOnly,
		LiteSystemPrompt:    true,
	}
}
