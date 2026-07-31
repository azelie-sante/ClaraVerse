# Lite model mode — design

## Problem

Every chat turn pays a fixed orchestration cost that is sized for large models.
Before the first token streams, the backend can spend:

| Item | Cost |
| --- | --- |
| Base system prompt | ~187 tokens, plus an always-appended appendix: markdown formatting guidelines ~329 tokens, and ask_user instructions ~527 tokens when tools are offered |
| Tool payload (no skill matched, no predictor) | 15,000–20,000 tokens |
| Tool prediction | an extra LLM round-trip, which itself ships the full tool list plus 6 messages of history |
| Memory selection | a blocking call with a 3-second timeout, plus ~1–2k tokens |

Title generation and context summarisation already run off the critical path
and are not part of this problem.

On a small model this is not merely slow, it is fatal. A 15–20k-token tool
payload does not fit alongside the conversation in an 8k context window, so the
request either overflows or the model produces a degraded or truncated reply.
Small models also gain the least from the machinery: they follow long system
prompts poorly, use subtle injected memories badly, and are unreliable at tool
selection.

## Goal

Let an administrator mark a model as "lite". Turns on that model skip the
orchestration that small models cannot exploit, so that a turn costs one
inference and a small bounded prompt.

Non-goal: changing behaviour for any existing model. The flag defaults to
false and the feature is inert until it is deliberately enabled.

## Behaviour

A lite turn differs from a normal turn in exactly four ways.

| Stage | Normal | Lite |
| --- | --- | --- |
| Skill routing | runs | **runs, unchanged** |
| Skill system prompt injection | injected when a skill matches | **skipped** |
| Tool set when no skill matches | tool-predictor LLM call, or the full payload when no predictor is configured | **the 8 essential tools, no extra call** |
| Memory selection | blocking, ≤3s, ~1–2k tokens | **skipped** |
| Base system prompt | ~187 tokens, plus an always-appended appendix: markdown formatting guidelines ~329 tokens, and ask_user instructions ~527 tokens when tools are offered | **~200-token variant** |

Streaming, the tool loop, tool execution, title generation and post-stream
context summarisation are untouched.

Skill routing is deliberately **kept**. It is keyword and pattern scoring with
no LLM call, and when a skill matches it *narrows* the tool set to that skill's
`RequiredTools` plus the essentials. Disabling it would enlarge the payload,
which is the opposite of the goal. What a skill adds beyond tools is its system
prompt, and that is the part a lite turn drops.

The 8 essential tools are the set already defined in `chat_service.go`:
`ask_user`, `search_web`, `search_images`, `get_current_time`, `calculate_math`,
`scrape_web`, `download_file`, `describe_image`.

Expected result per turn: one inference instead of two, and a prompt of roughly
96 tokens (623 with the ask_user instructions) plus a small tool set, instead of
a variable payload up to 20k on top of a ~1,043-token assembled prompt.

## Storage

`lite_mode BOOLEAN DEFAULT FALSE` on both the `models` and `model_aliases`
tables, following the pattern already used by `smart_tool_router` and
`agents_enabled`:

- Declared in `migrations/001_initial_schema.sql` for fresh installs.
- Added to existing databases by an idempotent self-migration in
  `internal/database/database.go`, guarded by the existing `tableExists` and
  `columnExists` helpers.

A `LiteMode` boolean field, tagged `json:"lite_mode"`, is added to the `Model`
struct in `internal/models/model.go` and to the alias representation used by
the admin endpoints.

`internal/handlers/model_management.go` accepts `lite_mode` as an optional
pointer field so partial updates leave it untouched when absent, matching how
`smart_tool_router` is handled today.

The admin model management table gains a toggle beside the existing capability
toggles.

## The policy object

All four decisions derive from one flag, so they are resolved once into a value
object rather than re-derived at each site:

```go
type ToolStrategy int

const (
    ToolsFull ToolStrategy = iota // current behaviour: skill, else predictor, else full payload
    ToolsEssentialsOnly           // skill's tools when matched, otherwise the 8 essentials
)

type TurnPolicy struct {
    SkipMemorySelection bool
    SkipSkillPrompt     bool
    ToolStrategy        ToolStrategy
    LiteSystemPrompt    bool
}

func resolveTurnPolicy(m *models.Model) TurnPolicy
```

`resolveTurnPolicy` is pure: no database, no network, no clock. A nil model, or
a model with `LiteMode` false, yields the zero value, which is exactly today's
behaviour.

`StreamChatCompletion` calls it once after resolving the model and reads its
fields at the four decision points.

## Integration points

All in `internal/services/chat_service.go`, inside `StreamChatCompletion`:

1. **Tool selection.** When `ToolStrategy == ToolsEssentialsOnly` and no skill
   matched, use the essential tool set directly instead of calling
   `toolPredictorService.PredictTools`.
2. **Skill prompt injection.** When `SkipSkillPrompt`, do not prepend the
   matched skill's system prompt. Tool narrowing from the skill still applies.
3. **Memory selection.** When `SkipMemorySelection`, skip the
   `memorySelectionService.SelectRelevantMemories` call and its 3-second
   context entirely.
4. **System prompt.** When `LiteSystemPrompt`, `GetSystemPrompt` returns the
   short variant plus the ask_user instructions when tools are offered. The
   markdown formatting guidelines are dropped: they are the conventions small
   models ignore or echo back verbatim. The ask_user instructions are kept
   because `ask_user` is in the lite essential tool set and is useless without
   them.

The lite system prompt is a new constant beside `getDefaultSystemPrompt()`. It
states the assistant's identity, that it may call the tools it is given, and
that it should answer directly and concisely. It deliberately omits the
artifact conventions, formatting rules and tool-usage guidance that small
models tend to ignore or echo back verbatim.

## Failure behaviour

If the model cannot be resolved, `resolveTurnPolicy` receives nil and returns
the zero value, which is `ToolsFull` — today's behaviour. Lite mode is only
ever entered when a model is successfully resolved and explicitly flagged. No
lookup failure can silently degrade a large model.

## Testing

`resolveTurnPolicy` is table-tested with no infrastructure:

- nil model yields the zero policy
- a model with `LiteMode` false yields the zero policy
- a model with `LiteMode` true yields all four lite decisions
- a lite policy never yields `ToolsFull`

The last case is the regression that would otherwise be invisible: a lite turn
quietly doing full-fat work looks like nothing at all, just unexplained
slowness.

The database migration is covered by the existing pattern — it is idempotent
and asserts the column exists after running.

## Out of scope

- Per-request overrides. The flag is a property of the model.
- A user-facing toggle in chat. Administrators decide which models are lite.
- Auto-detection from parameter count or context length. If that is wanted
  later it becomes an additional input to `resolveTurnPolicy`, not a redesign.
- Any change to non-lite turns.
