# ClaraVerse — Systems Reference

This document covers the **agentic + automation** subsystems that distinguish ClaraVerse from a generic chat UI: the durable workflow engine, the embedding-based memory system, the tool registry + MCP routing layer, and the OTel-backed observability story.

For repo layout, build/run instructions, and the broader chat/auth/admin systems, see [ARCHITECTURE.md](ARCHITECTURE.md) and [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md).

The level of detail here is "engineer who needs to extend or debug this subsystem" — every claim is backed by a file:line citation so you can jump into the code.

---

## 1. Workflow Engine — Durable Block DAG

The workflow engine runs user-defined block DAGs (LLM blocks, tool blocks, webhooks, conditionals, variables) with these durability properties: per-block checkpoint, heartbeat, orphan-resume on boot.

### Entry points

`internal/execution/engine.go` — the main runner.
`internal/execution/agent_block_executor.go` — the LLM block executor.
`internal/execution/llm_executor.go` — provider-aware LLM call.

### Durability

`internal/execution/state_store.go` — `MongoStateStore`. One document per execution in `workflow_execution_state`:
- `Init(execution_id, workflow_id, user_id, snapshot, input)` — upsert preserving `block_outputs`.
- `CheckpointBlock(execution_id, block_id, ckpt)` — atomic per-block write using dot-notation `block_outputs.<id>` so concurrent block goroutines don't race.
- `GetBlockOutput(execution_id, block_id)` — the **idempotency cache**. The engine calls this before each block; a hit means the block already completed in this execution and the output is reused. This is what makes side-effecting blocks safe across crash + resume.
- `Heartbeat` every 10s, `FindOrphaned(90s)` on boot, `MarkCompleted` at end.

Block IDs containing `.` or `$` get sanitized (`sanitizeBSONKey`) — Mongo would otherwise treat them as path separators / operators and silently corrupt the nested document. The sanitizer is symmetric: writes + reads both route through it.

### Concurrency limit

`internal/execution/concurrency.go` — `WorkflowLimiter`. Per-workflow concurrency cap (default 5) to stop a misconfigured webhook trigger from self-DoS-ing the system. `Acquire` returns a `*LimiterError` when at cap; the webhook handler maps this to HTTP 429 + `Retry-After` (`internal/handlers/webhook_trigger.go`).

### Observability

`internal/execution/tracing.go` bootstraps OpenTelemetry with three exporters wired in parallel: stdout (dev), OTLP HTTP (for Tempo/Jaeger if the operator wants it), and a custom **Mongo span exporter** (`internal/execution/mongo_span_exporter.go`).

The Mongo exporter is what powers the **in-product trace viewer** at `/admin/traces` — no external Tempo install needed. Every executor (`agent_block_executor.go`, `llm_executor.go`, `block_checker.go`, `engine.go`) wraps work in an OTel span; spans land in Mongo, the admin endpoint (`internal/handlers/traces_admin.go`) returns them, the React page (`frontend/src/pages/admin/Traces.tsx`) renders a waterfall.

### Per-execution cost

`internal/execution/cost.go` rolls up token + dollar cost across all LLM blocks in an execution. Surfaces in the execution detail view.

---

## 2. Memory — Embedding-Backed + Agentic Tools

Memory has two related but separate jobs:

### Storage + decay

`internal/services/memory_extraction_service.go` — after each conversation turn, an LLM extraction pass pulls candidate memories (facts the user told us, preferences, ongoing goals) and stores them with an embedding.

`internal/services/memory_decay_service.go` — periodic job (every 6h) that recomputes a freshness score and archives memories that drop below threshold. Stops the corpus from growing without bound.

### Retrieval

`internal/services/memory_selection_service.go` — for each chat turn, embed the incoming message (Bedrock Titan v2 via `internal/services/embedding_service.go`), cosine-similarity rank against the user's active memories, return top-k. Drops per-turn memory cost from "all memories in system prompt" (~500ms + ~1-2k tokens) to "top-k by similarity" (~50ms + ~0 tokens for the unused ones).

### Agentic tools

`internal/tools/memory_tools.go` exposes two tools the model can call mid-conversation:
- `add_memory(content, tags?)` — write a new memory immediately, not waiting for the post-turn extraction.
- `search_memory(query, k?)` — explicit recall when the model knows it needs older context.

The user-facing UI is `frontend/src/components/settings/MemoryList.tsx` (Settings → Memory). Lets the user view, edit, and delete what the system has stored about them.

### Bedrock embeddings note

The Bedrock OpenAI-compatible shim does NOT expose `/openai/v1/embeddings`. The embedding service detects Titan model IDs and routes those calls through the native invoke endpoint (`/model/amazon.titan-embed-text-v2:0/invoke`) using the same Bearer ABSK token. See `embedding_service.go` for the URL switch.

---

## 3. Tools + MCP Routing

### The shared registry

`internal/tools/registry.go:46` — one global singleton holding two layers:
- **Built-in tools** — backend functions registered in `registerBuiltInTools`. ~80 tools: search, math, GitHub, Slack, MongoDB, image gen, the memory tools, code execution via E2B.
- **MCP tools** — registered per-user when a Dobby's Claw client connects (`internal/services/mcp_bridge_service.go:113`). Marked `Source = MCP_LOCAL`.

Both layers serialize to OpenAI Chat Completions tool format via `Registry.List`. Chat and agent LLM calls share the same format.

### Tool selection asymmetry

| | Chat | Crew agent |
|---|---|---|
| Max iterations | 10 | 25 |
| Skill-first filtering | yes (skill's `RequiredTools` only) | no (skills add tools, never restrict) |
| Tool count cap | none enforced | 100, MCP-prioritized |
| Result truncation | post-stream cleanup | adaptive per-call |
| Overflow recovery | basic | 3-tier |
| Cross-step handoff | none | card pipeline + team memory |

**Chat tool selection** (`chat_service.go:1680-2000`): if a skill matched via `SkillService.RouteMessage`, only its `RequiredTools` + a small always-on set (`ask_user`, `search_web`, `describe_image`); else the `ToolPredictorService` narrows credential-filtered tools by LLM prediction or heuristic.

### MCP routing

The user's MCP client (Dobby's Claw) WebSocket-connects to the backend and sends its tool catalog (`MCPBridgeService.RegisterClient`). When the LLM calls a tool, the executor checks `tool.Source`:
- `MCP_LOCAL` → `MCPBridge.ExecuteToolOnClient` sends the call over the WS to the user's machine, waits 30-60s for the result on a channel.
- otherwise → in-process via `registry.Execute`.

This is why MCP tools work without any backend-side credentials — they execute on the user's machine.

```mermaid
sequenceDiagram
  participant L as LLM
  participant E as Executor
  participant R as Registry
  participant B as MCP Bridge
  participant C as Client (user's machine)
  L->>E: tool_call: foo
  E->>R: lookup foo
  R-->>E: tool{Source: MCP_LOCAL}
  E->>B: ExecuteToolOnClient(foo, args)
  B->>C: WebSocket: tool_call
  C->>C: run foo locally
  C->>B: result
  B-->>E: result
  E-->>L: tool result
```

### Skill resolution (3-tier)

For each user message:
1. **Explicit** from frontend
2. **Pinned** to the session
3. **Auto-routed** — `SkillService.RouteMessage` (`skill_service.go:354`) scores skills against the message: `TriggerPatterns` +20 prefix / +10 substring; `Keywords` +10 exact / +5 partial. Threshold 15.

Resolved skills attach two ways: `SystemPrompt` pasted into the agent's system prompt under `## Active Skills`, and `RequiredTools` merged into the agent's assigned tools.

---

## 4. The Bedrock OpenAI Shim

We hit Bedrock as if it were OpenAI — `bedrock-runtime.<region>.amazonaws.com/openai/v1/chat/completions`, Bearer ABSK key, standard tool-calling format. No SigV4, no proxy.

`internal/services/provider_service.go` auto-discovers Bedrock models via `/openai/v1/models` and stores them in MySQL. The admin UI lists discovered models with toggle visibility per tier.

Model identifier safety: the system used to pass composite `id:slug` strings to providers, which Bedrock rejected. The `GetByModelID` method was deleted entirely — all callers now use `GetByModelIDWithName` (returns the provider's actual model name) or `GetTextProviderWithModel`. Compiler-enforced contract; broken at compile time if a regression sneaks in.

Embeddings detour: `/openai/v1/embeddings` doesn't exist on Bedrock's shim. `embedding_service.go` routes Titan calls to the native invoke endpoint (see Memory section).

---

## 5. Where to start when extending the system

| Goal | Entry point |
|---|---|
| Add a new built-in tool | `internal/tools/registry.go` `registerBuiltInTools`, then write your tool implementation |
| Add a new built-in skill | `internal/services/skill_seeds.go` `getBuiltinSkills` |
| Add a new workflow block type | `internal/execution/` add an executor; register in `engine.go` |
| Add a new LLM provider | `internal/services/llm_*_provider.go`; register in `provider_service.go` |
| Add a Crew endpoint | `internal/handlers/crew.go` + register in `cmd/server/main.go` |

---

## 6. Integration test surface

Run with `go test -tags=integration -timeout 60s ./...` against a live Mongo (default `mongodb://localhost:27017`). Each test gets a throw-away database that's dropped on cleanup.

| Test file | Covers |
|---|---|
| `internal/execution/state_store_integration_test.go` | Workflow durability lifecycle, orphan detection, block ID sanitization, concurrency limiter |

Unit tests (no Mongo) run with plain `go test ./...`. Several pre-existing tests fail because the project migrated from SQLite to MySQL but their setup wasn't updated — unrelated to the new systems documented here.

---

## 7. Anti-goals (deliberately not in this system)

- **Multi-tenant cross-user agent communication.** Projects and their cards are user-scoped on purpose.
- **Agent-to-agent network calls.** All cross-agent communication goes through the card pipeline — keeps the execution model simple.
- **Backend-side persistent compute beyond the agent loop.** E2B sandboxes are per-chat (15-min idle eviction), not long-running services.
- **Forking workflows mid-execution.** A workflow execution is a DAG; if you need branching, the right model is multiple workflows triggered by the same event.
