package services

import (
	"strings"
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
	if len(lite) >= len(full) {
		t.Errorf("lite prompt (%d chars) must be shorter than the default (%d chars)", len(lite), len(full))
	}
}

func TestLiteAppendixContract(t *testing.T) {
	// The lite path deliberately drops the markdown formatting guidelines but
	// keeps the ask_user instructions, because ask_user is in the lite
	// essential tool set and is useless without them.
	assembled := getLiteSystemPrompt() + getAskUserInstructions()

	if strings.Contains(assembled, getMarkdownFormattingGuidelines()) {
		t.Error("lite path must not carry the markdown formatting guidelines")
	}
	if !strings.Contains(assembled, "ask_user") {
		t.Error("lite path must keep the ask_user instructions")
	}
}

func TestEssentialToolNamesAreRealTools(t *testing.T) {
	// A lite turn's entire toolbox is these exact-match strings. If a tool is
	// renamed in the registry, a lite turn silently ships without it.
	for name := range essentialToolNames {
		if groupOfTool(name) == "integrations" {
			t.Errorf("essential tool %q is not a recognised tool name — a lite turn would ship without it", name)
		}
	}
}
