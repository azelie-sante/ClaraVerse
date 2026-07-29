package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ─── Crew — the from-scratch agents rebuild ──────────────────────────────────
//
// Design (user-specified, 2026-07-29):
//   - A PROJECT is the unit of work: a brief every member and card can see —
//     fixing the old system's "no one knows what project they're on" problem.
//   - TEAM MEMBERS are chosen from a role catalog at setup (content writer,
//     developer, researcher, …). Each member gets ONLY the tools selected for
//     it — nothing else is ever available to that agent.
//   - Work is CARDS on a pipeline:  draft → queued → working → review → done.
//     Users create any number of drafts. Promoting queues a card; an agent
//     claims it atomically, works it, and submits output — which ALWAYS goes
//     to human review. Approve → done. Request changes → the review feedback
//     is attached and the card automatically re-queues for another pass.
//   - Agents run on the server-side harness (no local dependency); every run receives
//     the project brief + the board summary + the card's full revision history.

// CrewProject is a workspace: a goal/brief shared by every member and card.
type CrewProject struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID string             `bson:"userId" json:"user_id"`
	Name   string             `bson:"name" json:"name"`
	Brief  string             `bson:"brief" json:"brief"` // the shared context — "the why"
	// Goal is the project's north star: one measurable outcome every card must
	// serve. Injected into every agent run as a non-negotiable.
	Goal string `bson:"goal,omitempty" json:"goal,omitempty"`
	// Objectives map the goal into workstreams; cards link to one via
	// ObjectiveID so every task knows which part of the goal it serves.
	Objectives []CrewObjective `bson:"objectives,omitempty" json:"objectives,omitempty"`
	Status     string          `bson:"status" json:"status"` // active | archived
	CreatedAt  time.Time       `bson:"createdAt" json:"created_at"`
	UpdatedAt  time.Time       `bson:"updatedAt" json:"updated_at"`
}

// CrewObjective is one workstream under the project goal.
type CrewObjective struct {
	ID    string `bson:"id" json:"id"`
	Title string `bson:"title" json:"title"`
	Done  bool   `bson:"done" json:"done"`
}

// CrewMember is one hired agent on a project: a role + a fixed tool allowlist.
type CrewMember struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID string             `bson:"projectId" json:"project_id"`
	UserID    string             `bson:"userId" json:"user_id"`
	RoleKey   string             `bson:"roleKey" json:"role_key"` // from the role catalog
	// Charter is this member's lane: what they own on THIS project and what
	// they must hand off. Editable per member; injected into every run.
	Charter string `bson:"charter,omitempty" json:"charter,omitempty"`
	Name      string             `bson:"name" json:"name"`        // display name, e.g. "Riley (Researcher)"
	// Tools this member may use — chosen at hire time from the role's allowed
	// set. NOTHING outside this list is ever exposed to the agent.
	Tools     []string  `bson:"tools" json:"tools"`
	Model     string    `bson:"model,omitempty" json:"model,omitempty"` // empty = platform default
	Status    string    `bson:"status" json:"status"`                   // active | paused
	Documents []CrewDoc `bson:"documents,omitempty" json:"documents,omitempty"`
	// Attached skills (skill ids): each contributes its system prompt and
	// grants its required tools for this member's runs.
	SkillIDs []string `bson:"skillIds,omitempty" json:"skill_ids,omitempty"`
	TokensIn  int64     `bson:"tokensIn" json:"tokens_in"`              // lifetime usage
	TokensOut int64     `bson:"tokensOut" json:"tokens_out"`
	// MonthlyBudget caps tokens per calendar month (0 = unlimited). When
	// exhausted the member simply stops claiming cards until the month rolls.
	MonthlyBudget int64  `bson:"monthlyBudget,omitempty" json:"monthly_budget,omitempty"`
	MonthKey      string `bson:"monthKey,omitempty" json:"month_key,omitempty"`
	MonthTokens   int64  `bson:"monthTokens,omitempty" json:"month_tokens,omitempty"`
	CreatedAt time.Time `bson:"createdAt" json:"created_at"`
}

// CrewDoc is a reference document attached to a member — its extracted text is
// injected into that member's runs ("RAG in hand": the member grounds its work
// in these before using tools).
type CrewDoc struct {
	Name string    `bson:"name" json:"name"`
	Text string    `bson:"text" json:"-"` // extracted text (never shipped to the UI wholesale)
	Size int       `bson:"size" json:"size"` // extracted text length
	At   time.Time `bson:"at" json:"at"`
}

// Card pipeline statuses.
const (
	CardDraft   = "draft"   // created by the user; not visible to agents
	CardQueued  = "queued"  // promoted by the user; claimable by its assignee
	CardWorking = "working" // atomically claimed; an agent run is in flight
	CardReview  = "review"  // output submitted; waiting for HUMAN review (always)
	CardDone    = "done"    // approved by the human
)

// CardRevision is one completed agent pass and (optionally) the human review
// that followed it. The full list is fed back to the agent on re-queues.
type CardRevision struct {
	MemberID   string    `bson:"memberId,omitempty" json:"member_id,omitempty"`
	MemberName string    `bson:"memberName,omitempty" json:"member_name,omitempty"`
	Output     string    `bson:"output" json:"output"`
	Review     string    `bson:"review,omitempty" json:"review,omitempty"` // human feedback (empty until reviewed)
	Approved   bool      `bson:"approved" json:"approved"`
	TokensUsed int64     `bson:"tokensUsed,omitempty" json:"tokens_used,omitempty"`
	// ToolsUsed is the audit trail: which tools this run actually called, in order.
	ToolsUsed []string `bson:"toolsUsed,omitempty" json:"tools_used,omitempty"`
	At         time.Time `bson:"at" json:"at"`
}

// CrewCard is one unit of work moving through the pipeline.
type CrewCard struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID  string             `bson:"projectId" json:"project_id"`
	UserID     string             `bson:"userId" json:"user_id"`
	Title      string             `bson:"title" json:"title"`
	Detail     string             `bson:"detail,omitempty" json:"detail,omitempty"` // what exactly to do
	AssigneeID string             `bson:"assigneeId,omitempty" json:"assignee_id,omitempty"` // legacy single assignee (pre multi)
	// Multi-assignee: several members can work the SAME card — each produces
	// their own attempt. Awaiting = still to submit this round; Running = mid-run.
	AssigneeIDs []string `bson:"assigneeIds,omitempty" json:"assignee_ids,omitempty"`
	// ObjectiveID links this card to one project objective (the goal mapping).
	ObjectiveID string `bson:"objectiveId,omitempty" json:"objective_id,omitempty"`
	// Parts is the team-lead's split of a multi-assignee card: memberID → that
	// member's specific sub-brief. Present ⇒ coordinated flow (split → parallel
	// parts → combine); absent on single-assignee cards.
	Parts map[string]string `bson:"parts,omitempty" json:"parts,omitempty"`
	// SplitLock / CombineLock: at-most-once guards for the coordinator passes
	// (stale after 10 min so a crashed run never wedges the card).
	SplitLock   time.Time `bson:"splitLock,omitempty" json:"-"`
	CombineLock time.Time `bson:"combineLock,omitempty" json:"-"`
	// DependsOn lists card IDs that must be DONE before this card runs. The
	// worker skips it until then, and their approved outputs are injected into
	// this card's prompt.
	DependsOn []string `bson:"dependsOn,omitempty" json:"depends_on,omitempty"`
	// Repeat makes this a recurring card: "daily" | "weekly" | "monthly".
	// After approval the card is automatically re-queued at NextRunAt.
	Repeat    string    `bson:"repeat,omitempty" json:"repeat,omitempty"`
	NextRunAt time.Time `bson:"nextRunAt,omitempty" json:"next_run_at,omitempty"`
	Awaiting    []string `bson:"awaiting,omitempty" json:"awaiting,omitempty"`
	Running     []string `bson:"running,omitempty" json:"running,omitempty"`
	Status     string             `bson:"status" json:"status"`
	// Latest output lives in Revisions[len-1].Output; kept denormalized for lists.
	LatestOutput string         `bson:"latestOutput,omitempty" json:"latest_output,omitempty"`
	Revisions    []CardRevision `bson:"revisions,omitempty" json:"revisions,omitempty"`
	ClaimedAt    *time.Time     `bson:"claimedAt,omitempty" json:"claimed_at,omitempty"`
	Error        string         `bson:"error,omitempty" json:"error,omitempty"` // last run failure (surfaced, not hidden)
	CreatedAt    time.Time      `bson:"createdAt" json:"created_at"`
	UpdatedAt    time.Time      `bson:"updatedAt" json:"updated_at"`
}

// CrewRole is a catalog entry users hire from. AllowedTools is the superset a
// user may grant this role; DefaultTools is the pre-checked subset.
type CrewRole struct {
	Key          string   `json:"key"`
	Label        string   `json:"label"`
	Blurb        string   `json:"blurb"`
	Prompt       string   `json:"-"` // role system prompt (server-side only)
	AllowedTools []string `json:"allowed_tools"`
	DefaultTools []string `json:"default_tools"`
}
