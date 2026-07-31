# Lite Model Mode Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a per-model `lite_mode` flag that strips four pieces of per-turn orchestration small models cannot exploit, cutting a chat turn from up to two inferences and 20k tokens down to one inference and roughly 200 tokens plus a small tool set.

**Architecture:** A pure resolver turns a `*models.Model` into a `TurnPolicy` value object with four fields. `StreamChatCompletion` resolves the policy once and reads its fields at four existing decision points. The flag persists as a boolean column on `models` and `model_aliases`, following the pattern already used by `smart_tool_router`.

**Tech Stack:** Go 1.x, MySQL, Fiber v2, React + TypeScript (admin UI).

## Global Constraints

- The flag defaults to `false`. No existing model may change behaviour.
- `resolveTurnPolicy` must stay pure — no database, network, or clock access — so it is testable without infrastructure.
- Tests use the Go standard `testing` package and table-driven subtests. No testify; the repo does not use it.
- The zero value of `TurnPolicy` must equal today's behaviour, so a failed model lookup degrades to normal operation, never to lite.
- Skill routing is **kept** in lite mode. Only the skill's system prompt injection is skipped.
- Commit messages describe the code as it stands. No AI attribution trailers, and no comments narrating what changed.

---

### Task 1: TurnPolicy resolver

The pure core, built first so everything downstream has a tested contract.

**Files:**
- Create: `backend/internal/services/turn_policy.go`
- Test: `backend/internal/services/turn_policy_test.go`

**Interfaces:**
- Consumes: `claraverse/internal/models` (the `Model` struct)
- Produces: `ToolStrategy` (`ToolsFull`, `ToolsEssentialsOnly`), `TurnPolicy{SkipMemorySelection, SkipSkillPrompt, ToolStrategy, LiteSystemPrompt bool/ToolStrategy}`, and `resolveTurnPolicy(m *models.Model) TurnPolicy`

- [ ] **Step 1: Write the failing test**

Create `backend/internal/services/turn_policy_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/services/ -run TestResolveTurnPolicy -v`
Expected: FAIL — compile error, `undefined: TurnPolicy` and `models.Model` has no field `LiteMode`.

- [ ] **Step 3: Add the LiteMode field**

In `backend/internal/models/model.go`, add to the `Model` struct after `AgentsEnabled` (line 19):

```go
	LiteMode          bool      `json:"lite_mode"`                 // If true, chat turns skip memory, skill prompts and the tool predictor
```

- [ ] **Step 4: Write minimal implementation**

Create `backend/internal/services/turn_policy.go`:

```go
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
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd backend && go test ./internal/services/ -run 'TestResolveTurnPolicy|TestTurnPolicyZeroValue|TestLitePolicyNeverUsesFullTools' -v`
Expected: PASS, all three tests.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/services/turn_policy.go backend/internal/services/turn_policy_test.go backend/internal/models/model.go
git commit -m "Add TurnPolicy resolver for per-model orchestration decisions"
```

---

### Task 2: Persist the flag

**Files:**
- Modify: `backend/migrations/001_initial_schema.sql` (models table ~line 56, model_aliases ~line 85)
- Modify: `backend/internal/database/database.go:265` (add a migration block after the `agents_enabled` one)

**Interfaces:**
- Consumes: `models.Model.LiteMode` from Task 1
- Produces: a `lite_mode` column on both tables, defaulting to `FALSE`

- [ ] **Step 1: Add the column to the fresh-install schema**

In `backend/migrations/001_initial_schema.sql`, in the `models` table beside `agents_enabled`:

```sql
    lite_mode BOOLEAN DEFAULT FALSE COMMENT 'Skip memory, skill prompts and tool prediction for speed',
```

Add the identical line to the `model_aliases` table beside its own `agents_enabled` column.

- [ ] **Step 2: Add the self-migration for existing databases**

In `backend/internal/database/database.go`, immediately after the `agents_enabled` block that ends at line 274, add:

```go
	// Migration: Add lite_mode column to models table (if missing)
	if exists, _ := tableExists("models"); exists {
		if colExists, _ := columnExists("models", "lite_mode"); !colExists {
			log.Println("📦 Running migration: Adding lite_mode to models table")
			if _, err := db.Exec("ALTER TABLE models ADD COLUMN lite_mode BOOLEAN DEFAULT FALSE COMMENT 'Skip memory, skill prompts and tool prediction for speed'"); err != nil {
				return fmt.Errorf("failed to add lite_mode to models: %w", err)
			}
			log.Println("✅ Migration completed: models.lite_mode added")
		}
	}

	// Migration: Add lite_mode column to model_aliases table (if missing)
	if exists, _ := tableExists("model_aliases"); exists {
		if colExists, _ := columnExists("model_aliases", "lite_mode"); !colExists {
			log.Println("📦 Running migration: Adding lite_mode to model_aliases table")
			if _, err := db.Exec("ALTER TABLE model_aliases ADD COLUMN lite_mode BOOLEAN DEFAULT FALSE COMMENT 'Skip memory, skill prompts and tool prediction for speed'"); err != nil {
				return fmt.Errorf("failed to add lite_mode to model_aliases: %w", err)
			}
			log.Println("✅ Migration completed: model_aliases.lite_mode added")
		}
	}
```

- [ ] **Step 3: Verify it compiles**

Run: `cd backend && go build ./internal/...`
Expected: no output (success).

- [ ] **Step 4: Commit**

```bash
git add backend/migrations/001_initial_schema.sql backend/internal/database/database.go
git commit -m "Add lite_mode column to models and model_aliases"
```

---

### Task 3: Admin read and write path

Without this the toggle cannot be set or displayed.

**Files:**
- Modify: `backend/internal/services/model_management_service.go:1110-1121` (`UpdateModelRequest`), and the `UpdateModel` builder at `:73`
- Modify: `backend/internal/handlers/model_management.go:129` (raw-body boolean parsing)

**Interfaces:**
- Consumes: the `lite_mode` column from Task 2
- Produces: `UpdateModelRequest.LiteMode *bool`, persisted by `UpdateModel`

- [ ] **Step 1: Add the request field**

In `backend/internal/services/model_management_service.go`, add to `UpdateModelRequest` (after `SmartToolRouter *bool`, line 1119):

```go
	LiteMode          *bool
```

- [ ] **Step 2: Persist it in the update builder**

In the same file, in `UpdateModel`, immediately after the `SmartToolRouter` block (around line 115):

```go
	if req.LiteMode != nil {
		updateParts = append(updateParts, "lite_mode = ?")
		args = append(args, *req.LiteMode)
	}
```

- [ ] **Step 3: Parse it from the request body**

In `backend/internal/handlers/model_management.go`, after the `smart_tool_router` block (line 129-134):

```go
		if val, exists := rawBody["lite_mode"]; exists {
			if boolVal, ok := val.(bool); ok {
				req.LiteMode = &boolVal
				log.Printf("[DEBUG] Manually parsed lite_mode: %v", boolVal)
			}
		}
```

This mirrors the existing workaround for Fiber's `BodyParser` mishandling `*bool` when the value is `false` — without it, turning the flag **off** would silently do nothing.

- [ ] **Step 4: Return it to the admin UI**

Find every query that reads model rows for the admin surface and add `lite_mode` to both the column list and the `Scan` target:

Run: `cd backend && grep -rn "smart_tool_router" internal/services/model_service.go internal/services/model_management_service.go internal/handlers/admin.go`

For each `SELECT` listing `smart_tool_router`, add `lite_mode` immediately after it, and add `&m.LiteMode` at the matching position in the corresponding `rows.Scan(...)`. The column order in the SELECT and the Scan argument order must match exactly.

- [ ] **Step 5: Verify it compiles and tests still pass**

Run: `cd backend && go build ./internal/... && go test ./internal/services/ -run TestResolveTurnPolicy`
Expected: build succeeds, test passes.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/services/model_management_service.go backend/internal/handlers/model_management.go backend/internal/services/model_service.go
git commit -m "Allow lite_mode to be read and toggled through model management"
```

---

### Task 4: The lite system prompt

**Files:**
- Modify: `backend/internal/services/chat_service.go:3642` (beside `getDefaultSystemPrompt`)
- Test: `backend/internal/services/turn_policy_test.go`

**Interfaces:**
- Produces: `getLiteSystemPrompt() string`

- [ ] **Step 1: Write the failing test**

Append to `backend/internal/services/turn_policy_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/services/ -run TestLiteSystemPromptIsShort -v`
Expected: FAIL — `undefined: getLiteSystemPrompt`.

- [ ] **Step 3: Write the implementation**

In `backend/internal/services/chat_service.go`, immediately before `getDefaultSystemPrompt()` at line 3642:

```go
// getLiteSystemPrompt returns the compact prompt used for models flagged
// lite_mode. It states identity and the tool contract and stops there: the
// artifact conventions, formatting rules and extended tool guidance in the
// default prompt are routinely ignored or echoed back verbatim by small
// models, so including them costs tokens and buys nothing.
func getLiteSystemPrompt() string {
	return `You are Clara, a helpful AI assistant.

Answer directly and concisely. Prefer a short, correct answer over a long one.

You may be given tools. Call a tool only when you cannot answer without it, and use the exact tool name and argument names supplied. If no tool fits, answer from your own knowledge.

If a request is ambiguous, ask one short clarifying question instead of guessing.`
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/services/ -run TestLiteSystemPromptIsShort -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/services/chat_service.go backend/internal/services/turn_policy_test.go
git commit -m "Add compact system prompt for lite models"
```

---

### Task 5: Wire the policy into the turn

The behavioural change. Four edits at four existing decision points.

**Files:**
- Modify: `backend/internal/services/chat_service.go` — new helper near `:3572`; edits at `:1891`, `:1920`, `:2006`, `:3538`, `:3572`

**Interfaces:**
- Consumes: `resolveTurnPolicy` (Task 1), `getLiteSystemPrompt` (Task 4), `models.Model.LiteMode` (Task 1)
- Produces: `(s *ChatService) turnPolicyFor(modelID string) TurnPolicy`

- [ ] **Step 1: Add the database-backed wrapper**

In `backend/internal/services/chat_service.go`, immediately before `GetSystemPrompt` at line 3572:

```go
// turnPolicyFor loads the lite_mode flag for a model and resolves the policy.
// Any lookup failure yields the zero policy, which is the full-featured path.
func (s *ChatService) turnPolicyFor(modelID string) TurnPolicy {
	if modelID == "" || s.db == nil {
		return TurnPolicy{}
	}
	var lite bool
	if err := s.db.QueryRow(`SELECT lite_mode FROM models WHERE id = ?`, modelID).Scan(&lite); err != nil {
		return TurnPolicy{}
	}
	return resolveTurnPolicy(&models.Model{ID: modelID, LiteMode: lite})
}
```

- [ ] **Step 2: Resolve the policy once per turn**

In `StreamChatCompletion`, immediately after the `config, err := s.GetEffectiveConfig(...)` call at line 1696 and its error handling, add:

```go
	policy := s.turnPolicyFor(userConn.ModelID)
	if policy.ToolStrategy == ToolsEssentialsOnly {
		log.Printf("⚡ [LITE] Lite mode active for model %s — skipping memory, skill prompt and tool prediction", userConn.ModelID)
	}
```

- [ ] **Step 3: Skip the tool predictor**

Two edits.

First, at line 1881, add the policy check to the predictor branch condition so a lite turn never enters it. Replace:

```go
			} else if s.toolPredictorService != nil && len(credentialFilteredTools) > 0 {
```

with:

```go
			} else if s.toolPredictorService != nil && len(credentialFilteredTools) > 0 && policy.ToolStrategy != ToolsEssentialsOnly {
```

Second, replace the final `else` block at lines 1918-1922, which currently reads:

```go
			} else {
				// No skill, no predictor — use all credential-filtered tools
				log.Printf("📦 [REQUEST] No skill matched, no predictor, using all %d filtered tools", len(credentialFilteredTools))
				tools = credentialFilteredTools
			}
```

with:

```go
			} else if policy.ToolStrategy == ToolsEssentialsOnly {
				essentialToolNames := map[string]bool{
					"ask_user": true, "search_web": true, "search_images": true,
					"get_current_time": true, "calculate_math": true, "scrape_web": true,
					"download_file": true, "describe_image": true,
				}
				filtered := make([]map[string]interface{}, 0, len(essentialToolNames))
				for _, toolDef := range credentialFilteredTools {
					if fn, ok := toolDef["function"].(map[string]interface{}); ok {
						if name, ok := fn["name"].(string); ok && essentialToolNames[name] {
							filtered = append(filtered, toolDef)
						}
					}
				}
				tools = filtered
				log.Printf("⚡ [LITE] Using %d essential tools (from %d available)", len(tools), len(credentialFilteredTools))
			} else {
				// No skill, no predictor — use all credential-filtered tools
				log.Printf("📦 [REQUEST] No skill matched, no predictor, using all %d filtered tools", len(credentialFilteredTools))
				tools = credentialFilteredTools
			}
```

Leave the data-file dependency-closure block that follows at line 1924 untouched — a lite turn still gets the full data-analysis group when a data file is present, which is correct: those tools are useless individually.

- [ ] **Step 4: Skip skill prompt injection**

At line 2006 (`// SKILL PROMPT INJECTION`), wrap the block that prepends the skill's system prompt:

```go
	if !policy.SkipSkillPrompt {
		// existing skill prompt injection block, unchanged
	}
```

The skill's tool narrowing is untouched — only the prompt text is skipped.

- [ ] **Step 5: Skip memory selection**

The call at line 3538 lives in `buildMemoryContext(userConn *models.UserConnection) string`, declared at line 3505. Add a guard directly after its existing nil check, so the block reads:

```go
func (s *ChatService) buildMemoryContext(userConn *models.UserConnection) string {
	// Check if memory selection service is available
	if s.memorySelectionService == nil {
		return ""
	}

	// Lite models skip memory entirely — the selection call blocks the turn for
	// up to 3 seconds and small models make poor use of injected memories.
	if s.turnPolicyFor(userConn.ModelID).SkipMemorySelection {
		return ""
	}
```

This removes both the blocking call and the ~1–2k tokens it would have injected.

- [ ] **Step 6: Use the lite system prompt**

`GetSystemPrompt` builds `appendix` at lines 3573-3583 as `getAskUserInstructions() + getMarkdownFormattingGuidelines()` when `includeAskUser` is true, and just the formatting guidelines otherwise. Measured sizes: formatting guidelines 1,315 chars (~329 tokens), ask_user instructions 2,109 chars (~527 tokens).

A lite turn drops the **formatting guidelines** — markdown conventions are what small models ignore or echo back verbatim — but keeps the **ask_user instructions** when tools are offered, because `ask_user` is in the lite essential tool set and is useless without them.

In `GetSystemPrompt`, after `temporalContext` is built (line 3586) and before the Priority 1 `SystemInstructions` check at line 3591, add:

```go
	if s.turnPolicyFor(userConn.ModelID).LiteSystemPrompt {
		liteAppendix := ""
		if includeAskUser {
			liteAppendix = getAskUserInstructions()
		}
		log.Printf("⚡ [LITE] Using compact system prompt for %s", userConn.ModelID)
		return temporalContext + getLiteSystemPrompt() + liteAppendix
	}
```

This deliberately omits `memoryContext`, which is empty in lite mode anyway, and it must sit **before** the Priority 1 check so a lite model is not handed the full prompt stack via another path.

- [ ] **Step 7: Verify the build and the full service test suite**

Run: `cd backend && go build ./internal/... && go test ./internal/services/ 2>&1 | tail -20`
Expected: build succeeds. The services suite has 21 pre-existing failures unrelated to this work (they need a live database) — confirm the count is still 21 and that no `TurnPolicy` or `LiteSystemPrompt` test is among the failures.

- [ ] **Step 8: Commit**

```bash
git add backend/internal/services/chat_service.go
git commit -m "Apply turn policy to tool selection, skill prompts, memory and system prompt"
```

---

### Task 6: Admin UI toggle

**Files:**
- Modify: `frontend/src/pages/admin/ModelManagement.tsx`
- Modify: `frontend/src/types/` — whichever file declares the `Model` interface (find with the grep below)

**Interfaces:**
- Consumes: the `lite_mode` field returned by the admin API from Task 3

- [ ] **Step 1: Add the field to the TypeScript type**

Run: `cd frontend && grep -rn "smart_tool_router" src/types/`

In the interface that declares `smart_tool_router: boolean`, add alongside it:

```ts
  lite_mode: boolean;
```

- [ ] **Step 2: Add the toggle to the admin table**

Run: `cd frontend && grep -n "smart_tool_router" src/pages/admin/ModelManagement.tsx`

For each place `smart_tool_router` appears — the column header, the cell rendering its toggle, and the update payload — add an equivalent `lite_mode` entry. Label the column **Lite** with the tooltip: `Skips memory, skill prompts and tool prediction. Best for small local models.`

- [ ] **Step 3: Verify types and lint on the touched files**

Run: `cd frontend && npx tsc --noEmit -p tsconfig.app.json 2>&1 | grep -E "ModelManagement|types/" | head`
Expected: no new errors mentioning these files. The project has ~106 pre-existing errors elsewhere; ignore those.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/admin/ModelManagement.tsx frontend/src/types/
git commit -m "Add lite mode toggle to admin model management"
```

---

## Verification

After Task 6, confirm end to end:

1. `cd backend && go build ./cmd/server/` succeeds.
2. Start the stack, open admin model management, toggle **Lite** on a small model, reload — the toggle persists (this exercises the `*bool` false-value workaround from Task 3).
3. Send a chat message on that model and confirm the backend logs show `⚡ [LITE] Lite mode active`, `⚡ [LITE] Using N essential tools`, and `⚡ [LITE] Using compact system prompt`, and that **no** `[TOOL-PREDICTOR]` line appears for that turn.
4. Send a message on a non-lite model and confirm none of the `[LITE]` lines appear and behaviour is unchanged.
