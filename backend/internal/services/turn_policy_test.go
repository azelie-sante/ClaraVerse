package services

import (
	"testing"

	"claraverse/internal/models"
)

func TestResolveTurnPolicy(t *testing.T) {
	tests := []struct {
		name  string
		model *models.Model
		want  TurnPolicy
	}{
		{
			name:  "nil model falls back to full behaviour",
			model: nil,
			want:  TurnPolicy{},
		},
		{
			name:  "normal model keeps full behaviour",
			model: &models.Model{ID: "gpt-4", LiteMode: false},
			want:  TurnPolicy{},
		},
		{
			name:  "lite model strips orchestration",
			model: &models.Model{ID: "qwen-1.5b", LiteMode: true},
			want: TurnPolicy{
				SkipMemorySelection: true,
				SkipSkillPrompt:     true,
				ToolStrategy:        ToolsEssentialsOnly,
				LiteSystemPrompt:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveTurnPolicy(tt.model)
			if got != tt.want {
				t.Errorf("resolveTurnPolicy() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// The zero value must mean "today's behaviour" so a failed lookup can never
// silently downgrade a large model.
func TestTurnPolicyZeroValueIsFullBehaviour(t *testing.T) {
	var p TurnPolicy
	if p.ToolStrategy != ToolsFull {
		t.Errorf("zero TurnPolicy.ToolStrategy = %v, want ToolsFull", p.ToolStrategy)
	}
	if p.SkipMemorySelection || p.SkipSkillPrompt || p.LiteSystemPrompt {
		t.Errorf("zero TurnPolicy should skip nothing, got %+v", p)
	}
}

func TestLitePolicyNeverUsesFullTools(t *testing.T) {
	p := resolveTurnPolicy(&models.Model{LiteMode: true})
	if p.ToolStrategy == ToolsFull {
		t.Error("lite policy must not use ToolsFull — this is the regression that looks like nothing but unexplained slowness")
	}
}

func TestLiteSystemPromptIsShort(t *testing.T) {
	lite := getLiteSystemPrompt()
	full := getDefaultSystemPrompt()

	if lite == "" {
		t.Fatal("lite system prompt must not be empty")
	}
	// The whole point is a small prompt. ~4 chars per token, so 1200 chars is
	// roughly 300 tokens — a generous ceiling for a ~200 token target.
	if len(lite) > 1200 {
		t.Errorf("lite prompt is %d chars, want <= 1200", len(lite))
	}
	if len(lite) >= len(full) {
		t.Errorf("lite prompt (%d chars) must be shorter than the default (%d chars)", len(lite), len(full))
	}
}
