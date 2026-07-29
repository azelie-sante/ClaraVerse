package services

import (
	"os"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"claraverse/internal/models"
)

// CrewExecutor is the seam to the headless tool-loop engine (implemented by
// execution.AgentBlockExecutor via a tiny adapter in main.go). It runs one
// agent pass and returns the final text + token usage.
type CrewExecutor interface {
	RunCrewTask(ctx context.Context, userID, model, systemPrompt, userPrompt string, tools []string) (output string, toolsUsed []string, tokensIn, tokensOut int64, err error)
}

// CrewModel resolves the model every crew agent runs on: the CREW_MODEL env
// var when set, otherwise empty — the executor falls back to the instance's
// default provider/model (BYO-model friendly for self-hosters).
var CrewModel = os.Getenv("CREW_MODEL")

// CrewWorker turns queued cards into review-ready output: a ticker scans
// active members for queued cards, claims one atomically, runs the member's
// agent through the harness, and submits the result to HUMAN review. All
// failures are surfaced on the card (never silently retried).
type CrewWorker struct {
	svc  *CrewService
	exec CrewExecutor
	// Global concurrency cap — agent runs are model-heavy; keep the box safe.
	slots chan struct{}
	// Per-member in-flight guard so one member never runs two cards at once.
	busy sync.Map
	// Notify (optional) tells the user work reached review — e.g. via Telegram.
	Notify func(userID, projectName, cardTitle string)
}

func NewCrewWorker(svc *CrewService, exec CrewExecutor, concurrency int) *CrewWorker {
	if concurrency <= 0 {
		concurrency = 2
	}
	return &CrewWorker{svc: svc, exec: exec, slots: make(chan struct{}, concurrency)}
}

// Start launches the scan loop. Cheap when idle: one indexed query per tick.
func (w *CrewWorker) Start(ctx context.Context) {
	go func() {
		t := time.NewTicker(15 * time.Second)
		defer t.Stop()
		log.Println("✅ Crew worker started (Nexus v2 pipeline)")
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				w.tick(ctx)
			}
		}
	}()
}

func (w *CrewWorker) tick(ctx context.Context) {
	w.svc.RequeueDueRecurring(ctx)
	w.handleSplits(ctx)   // multi-assignee: team-lead divides the work first
	w.handleCombines(ctx) // ...and merges the parts into one deliverable after
	members, err := w.svc.MembersWithQueuedCards(ctx)
	if err != nil {
		log.Printf("⚠️ [CREW] scan: %v", err)
		return
	}
	for _, m := range members {
		if MemberOverBudget(&m) {
			continue // monthly token budget exhausted — resumes next month
		}
		mid := m.ID.Hex()
		if _, running := w.busy.LoadOrStore(mid, true); running {
			continue // this member is already working a card
		}
		member := m
		go func() {
			defer w.busy.Delete(mid)
			select {
			case w.slots <- struct{}{}:
				defer func() { <-w.slots }()
			case <-time.After(5 * time.Second):
				return // saturated — the card stays queued for the next tick
			}
			w.runOne(ctx, &member)
		}()
	}
}

// runOne claims and executes a single card for a member.
func (w *CrewWorker) runOne(ctx context.Context, m *models.CrewMember) {
	card, err := w.svc.ClaimNextCard(ctx, m.ID.Hex())
	if err != nil || card == nil {
		return
	}
	log.Printf("🤖 [CREW] %s working card %q (%s)", m.Name, card.Title, card.ID.Hex())

	runCtx, cancel := context.WithTimeout(ctx, 8*time.Minute)
	defer cancel()

	// Attached skills grant their required tools for this run (union with the
	// member's own allowlist). Computed first so the prompt can state the
	// member's EXACT capabilities.
	runTools := append([]string{}, m.Tools...)
	seenT := map[string]bool{}
	for _, t := range runTools {
		seenT[t] = true
	}
	for _, sk := range w.svc.GetSkillsByIDs(runCtx, m.SkillIDs) {
		for _, t := range sk.RequiredTools {
			if !seenT[t] {
				seenT[t] = true
				runTools = append(runTools, t)
			}
		}
	}

	sys, user, err := w.buildPrompts(runCtx, m, card, runTools)
	if err != nil {
		_ = w.svc.SubmitOutput(ctx, card.ID.Hex(), m.ID.Hex(), m.Name, "", nil, 0, "could not build task context: "+err.Error())
		return
	}
	out, toolsUsed, tin, tout, err := w.exec.RunCrewTask(runCtx, m.UserID, CrewModel, sys, user, runTools)
	w.svc.AddMemberTokens(ctx, m.ID.Hex(), tin, tout)
	if err != nil {
		log.Printf("⚠️ [CREW] %s failed card %s: %v", m.Name, card.ID.Hex(), err)
		_ = w.svc.SubmitOutput(ctx, card.ID.Hex(), m.ID.Hex(), m.Name, "", toolsUsed, tin+tout, err.Error())
		return
	}
	if strings.TrimSpace(out) == "" {
		_ = w.svc.SubmitOutput(ctx, card.ID.Hex(), m.ID.Hex(), m.Name, "", toolsUsed, tin+tout, "the agent produced no output")
		return
	}
	if err := w.svc.SubmitOutput(ctx, card.ID.Hex(), m.ID.Hex(), m.Name, out, toolsUsed, tin+tout, ""); err != nil {
		log.Printf("⚠️ [CREW] submit output %s: %v", card.ID.Hex(), err)
		return
	}
	log.Printf("✅ [CREW] %s finished card %q → review (%d tokens)", m.Name, card.Title, tin+tout)
	// Tell the user only when the card actually reached review (multi-assignee
	// cards wait for every pass before the inbox moment).
	if w.Notify != nil {
		if fresh, err := w.svc.GetCard(ctx, card.ID.Hex(), m.UserID); err == nil && fresh.Status == models.CardReview {
			if p, err := w.svc.GetProject(ctx, card.ProjectID, m.UserID); err == nil {
				w.Notify(m.UserID, p.Name, card.Title)
			}
		}
	}
}

// buildPrompts assembles the full context every run gets: role, project brief,
// board summary, the task, and the complete revision/review history.
func (w *CrewWorker) buildPrompts(ctx context.Context, m *models.CrewMember, card *models.CrewCard, runTools []string) (string, string, error) {
	project, err := w.svc.GetProject(ctx, card.ProjectID, m.UserID)
	if err != nil {
		return "", "", err
	}
	role := w.svc.RoleByKey(m.RoleKey)
	rolePrompt := "You are a capable agent."
	if role != nil {
		rolePrompt = role.Prompt
	}
	board, _ := w.svc.BoardSummary(ctx, m.UserID, card.ProjectID)
	// Goal + objective mapping: the north star and which workstream this card serves.
	goalSection := ""
	if strings.TrimSpace(project.Goal) != "" {
		goalSection = "\n## PROJECT GOAL (north star — NON-NEGOTIABLE)\nEvery deliverable must visibly serve this goal. If this task seems to conflict with it, say so at the top of your output and explain the tension.\n" + strings.TrimSpace(project.Goal) + "\n"
	}
	if len(project.Objectives) > 0 {
		var ob strings.Builder
		ob.WriteString("\n## OBJECTIVES (how the goal maps into workstreams)\n")
		for _, o := range project.Objectives {
			mark := " "
			if o.Done {
				mark = "x"
			}
			tag := ""
			if o.ID == card.ObjectiveID {
				tag = "  ← THIS TASK serves this objective"
			}
			fmt.Fprintf(&ob, "- [%s] %s%s\n", mark, o.Title, tag)
		}
		goalSection += ob.String()
	}
	// The member's lane on this project.
	charterSection := ""
	if strings.TrimSpace(m.Charter) != "" {
		charterSection = "\n## YOUR LANE ON THIS TEAM\n" + strings.TrimSpace(m.Charter) + "\nStay in this lane. If the task needs work outside it, do your part fully and explicitly note what should be handed to a teammate.\n"
	}
	// Capability honesty: the member knows exactly what it can and cannot do.
	capSection := "\n## WHAT YOU CAN AND CANNOT DO (be honest about this)\nYou have EXACTLY these tools and nothing else: " + strings.Join(runTools, ", ") + ".\nAnything outside this list — publishing to a CMS or blog, posting to social media, deploying, or any external action without a matching tool — is IMPOSSIBLE for you in this run.\nNEVER claim to have performed an action unless the tool call actually succeeded. If the task asks for an action you cannot perform, produce the complete deliverable (the post, the reply, the report) plus exact step-by-step instructions for the reviewer to execute it — and say plainly that you could not do it directly.\n"
	memory, _ := w.svc.ProjectMemory(ctx, m.UserID, card.ProjectID, 8)
	// Member reference documents ("RAG in hand"): ground the member's work in
	// its uploaded docs before it reaches for tools. Capped to keep prompts sane.
	docsSection := ""
	if len(m.Documents) > 0 {
		var db strings.Builder
		budget := 12_000
		for _, d := range m.Documents {
			t := d.Text
			if len(t) > budget {
				t = t[:budget]
			}
			fmt.Fprintf(&db, "### %s\n%s\n\n", d.Name, t)
			budget -= len(t)
			if budget <= 0 {
				break
			}
		}
		docsSection = "\n## YOUR REFERENCE DOCUMENTS — ground your work in these FIRST, before tools\n" + db.String()
	}
	skillSection := ""
	if len(m.SkillIDs) > 0 {
		var sb strings.Builder
		for _, sk := range w.svc.GetSkillsByIDs(ctx, m.SkillIDs) {
			if strings.TrimSpace(sk.SystemPrompt) == "" {
				continue
			}
			fmt.Fprintf(&sb, "### Skill: %s\n%s\n\n", sk.Name, strings.TrimSpace(sk.SystemPrompt))
		}
		if sb.Len() > 0 {
			skillSection = "\n## YOUR SKILLS — apply these practiced approaches where relevant\n" + sb.String()
		}
	}
	memorySection := ""
	if memory != "" {
		memorySection = "\n## TEAM MEMORY — approved work on this project so far\nUse this for continuity (terminology, facts already established, tone) — but it must NOT constrain or bias YOUR task's conclusions. Your deliverable stands on its own.\n" + memory + "\n"
	}

	sys := fmt.Sprintf(`%s

You are %q, a team member on the project %q.

## PROJECT BRIEF (the shared context — everything you produce serves this)
%s
%s%s
## THE BOARD (what the rest of the team is doing — align, don't duplicate)
%s%s%s%s%s
## HOW TO WORK (non-negotiable)
1. UNDERSTAND first. If the task is vague, do NOT produce something vague: choose the most useful concrete interpretation for this project, STATE your assumptions in one short line at the top, and deliver something specific and complete. You may add a short "Questions for the reviewer" list at the very END — after a full best-effort deliverable, never instead of one.
2. GATHER thoroughly. One tool call is almost never enough — search multiple angles, open the actual pages, cross-check. Keep going until you have real material.
3. THEN COMPOSE. The deliverable is YOUR writing: structured, titled, with a 2-3 sentence executive summary up top, clear sections, and your synthesis of what it means for THIS project. NEVER paste raw tool output (search listings, scraped text, JSON) as the deliverable — quote selectively, cite sources at the end.
4. PRESENT like a professional. Clean markdown: one H1 title, section headings, short paragraphs, tables where they help, bold the key numbers/findings. It should read like a report a colleague is proud to sign — a human reviews every word.
5. Never fabricate tool results or sources. Never invent names of real people, statistics, salaries, or metrics — write [PLACEHOLDER — needs real data] instead and list what's needed. If something could not be verified or a tool failed, say so explicitly.`,
		rolePrompt, m.Name, project.Name, strings.TrimSpace(project.Brief), goalSection, charterSection, board, skillSection, memorySection, docsSection, capSection)

	var u strings.Builder
	fmt.Fprintf(&u, "## YOUR TASK\n%s\n", card.Title)
	if len(card.DependsOn) > 0 {
		wrote := false
		for _, depID := range card.DependsOn {
			dep, err := w.svc.GetCard(ctx, depID, m.UserID)
			if err != nil || dep.Status != models.CardDone || strings.TrimSpace(dep.LatestOutput) == "" {
				continue
			}
			if !wrote {
				u.WriteString("\n## INPUT FROM COMPLETED DEPENDENCY TASKS — your task builds directly on this approved work\n")
				wrote = true
			}
			fmt.Fprintf(&u, "\n### From %q:\n%s\n", dep.Title, truncateCrew(dep.LatestOutput, 3500))
		}
	}
	if brief, ok := card.Parts[m.ID.Hex()]; ok && len(card.AssigneeIDs) > 1 {
		u.WriteString("\n## YOUR PART (the team lead split this task — do ONLY your part, teammates cover the rest)\n" + brief + "\n")
		var others strings.Builder
		for aid, b := range card.Parts {
			if aid == m.ID.Hex() {
				continue
			}
			om, err := w.svc.GetMember(ctx, aid, m.UserID)
			if err != nil {
				continue
			}
			fmt.Fprintf(&others, "- %s: %s\n", om.Name, truncateCrew(b, 200))
		}
		if others.Len() > 0 {
			u.WriteString("\nTeammates' parts (do NOT duplicate these):\n" + others.String())
		}
	} else if len(card.AssigneeIDs) > 1 {
		u.WriteString("\n(Note: other teammates are also submitting their own take on this task — produce YOUR best deliverable for YOUR role; the reviewer will combine them.)\n")
	}
	if strings.TrimSpace(card.Detail) != "" {
		fmt.Fprintf(&u, "\n%s\n", card.Detail)
	}
	if len(card.Revisions) > 0 {
		u.WriteString("\n## PREVIOUS ATTEMPTS ON THIS TASK — address every point of feedback\n")
		for i, r := range card.Revisions {
			fmt.Fprintf(&u, "\n### Attempt %d output:\n%s\n", i+1, truncateCrew(r.Output, 4000))
			if strings.TrimSpace(r.Review) != "" {
				fmt.Fprintf(&u, "\n### Reviewer feedback on attempt %d (MUST be addressed):\n%s\n", i+1, r.Review)
			}
		}
		u.WriteString("\nProduce a REVISED deliverable that fully incorporates the feedback. Do not repeat rejected work.\n")
	}
	return sys, u.String(), nil
}

func truncateCrew(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "\n…[truncated]"
}

// ─── Planner: goal → proposed draft cards (Paperclip's CEO pattern, but the
// plan ALWAYS lands as drafts for human approval — never auto-queued) ────────

// PlanProject runs one tool-less agent pass that reads the goal, brief,
// objectives and team, and proposes cards. Returns the created DRAFT cards.
func (w *CrewWorker) PlanProject(ctx context.Context, userID, projectID string) ([]models.CrewCard, error) {
	p, err := w.svc.GetProject(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	members, err := w.svc.ListMembers(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return nil, fmt.Errorf("hire at least one team member before planning")
	}
	if strings.TrimSpace(p.Goal) == "" && strings.TrimSpace(p.Brief) == "" {
		return nil, fmt.Errorf("set a goal or brief first — the planner needs something to plan toward")
	}
	board, _ := w.svc.BoardSummary(ctx, userID, projectID)

	var tb strings.Builder
	for _, m := range members {
		role := w.svc.RoleByKey(m.RoleKey)
		label := m.RoleKey
		if role != nil {
			label = role.Label
		}
		charter := m.Charter
		if charter == "" {
			charter = "(no charter set)"
		}
		fmt.Fprintf(&tb, "- %q — %s. Lane: %s\n", m.Name, label, charter)
	}
	var ob strings.Builder
	for _, o := range p.Objectives {
		if !o.Done {
			fmt.Fprintf(&ob, "- id=%q %s\n", o.ID, o.Title)
		}
	}

	sys := `You have NO tools in this pass — respond with plain text only; never emit tool-call syntax.

You are the planning lead for a team of AI agents. Break the project goal into concrete task cards.

Rules:
- Each card is ONE deliverable with a clear "done": specific, self-contained, reviewable.
- 3 to 8 cards. Order them so early cards feed later ones.
- Assign each card to the best-suited member(s) BY EXACT NAME from the team list. Respect their lanes.
- Link each card to the objective id it serves when one fits.
- A card the team should redo on a cadence may set "repeat": "daily" | "weekly" | "monthly" — use sparingly.
- Do NOT duplicate work already on the board.
- Respond with ONLY a JSON array, no prose, no code fences:
[{"title":"...","detail":"...","objective_id":"..." ,"assignees":["Member Name"],"repeat":""}]
"detail" is the full working instruction for the agent: context, requirements, acceptance criteria.`

	user := fmt.Sprintf("## GOAL\n%s\n\n## BRIEF\n%s\n\n## OBJECTIVES (use these ids)\n%s\n## TEAM\n%s\n## BOARD TODAY (don't duplicate)\n%s",
		strings.TrimSpace(p.Goal), strings.TrimSpace(p.Brief), ob.String(), tb.String(), board)

	runCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	out, _, _, _, err := w.exec.RunCrewTask(runCtx, userID, CrewModel, sys, user, nil)
	if err != nil {
		return nil, fmt.Errorf("planner run failed: %w", err)
	}

	var proposed []struct {
		Title       string   `json:"title"`
		Detail      string   `json:"detail"`
		ObjectiveID string   `json:"objective_id"`
		Assignees   []string `json:"assignees"`
		Repeat      string   `json:"repeat"`
	}
	if err := json.Unmarshal([]byte(extractJSONArray(out)), &proposed); err != nil {
		return nil, fmt.Errorf("planner returned an unreadable plan — try again")
	}
	nameToID := map[string]string{}
	for _, m := range members {
		nameToID[strings.ToLower(strings.TrimSpace(m.Name))] = m.ID.Hex()
	}
	created := []models.CrewCard{}
	for i, pc := range proposed {
		if i >= 8 || strings.TrimSpace(pc.Title) == "" {
			break
		}
		var assignees []string
		for _, n := range pc.Assignees {
			if id, okA := nameToID[strings.ToLower(strings.TrimSpace(n))]; okA {
				assignees = append(assignees, id)
			}
		}
		card, err := w.svc.CreateCard(ctx, userID, projectID, pc.Title, pc.Detail, pc.ObjectiveID, pc.Repeat, assignees, nil)
		if err != nil {
			continue // one bad card must not sink the plan
		}
		created = append(created, *card)
	}
	if len(created) == 0 {
		return nil, fmt.Errorf("the planner produced no usable cards — try again")
	}
	log.Printf("🧭 [CREW] planner proposed %d draft cards for project %s", len(created), projectID)
	return created, nil
}

// extractJSONArray pulls the outermost JSON array out of model output that may
// be wrapped in prose or code fences.
func extractJSONArray(s string) string {
	start := strings.Index(s, "[")
	end := strings.LastIndex(s, "]")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}

// PlanCard decomposes ONE card into smaller sub-cards (drafts), wiring
// dependencies between them so outputs flow in order. The parent card stays —
// the user decides whether to keep or delete it after reviewing the breakdown.
func (w *CrewWorker) PlanCard(ctx context.Context, userID, cardID string) ([]models.CrewCard, error) {
	card, err := w.svc.GetCard(ctx, cardID, userID)
	if err != nil {
		return nil, err
	}
	p, err := w.svc.GetProject(ctx, card.ProjectID, userID)
	if err != nil {
		return nil, err
	}
	members, err := w.svc.ListMembers(ctx, userID, card.ProjectID)
	if err != nil || len(members) == 0 {
		return nil, fmt.Errorf("hire at least one team member before breaking a card down")
	}
	var tb strings.Builder
	for _, m := range members {
		role := w.svc.RoleByKey(m.RoleKey)
		label := m.RoleKey
		if role != nil {
			label = role.Label
		}
		charter := m.Charter
		if charter == "" {
			charter = "(no charter set)"
		}
		fmt.Fprintf(&tb, "- %q — %s. Lane: %s\n", m.Name, label, charter)
	}

	sys := `You have NO tools in this pass — respond with plain text only; never emit tool-call syntax.

You are the planning lead for a team of AI agents. Break ONE task card into a short pipeline of smaller cards.

Rules:
- 2 to 6 sub-cards. Each is ONE deliverable with a clear "done", small enough for a single focused agent pass.
- Together the sub-cards must fully cover the original card — nothing dropped.
- Assign each sub-card to the best-suited member(s) BY EXACT NAME from the team list. Respect their lanes.
- When a sub-card builds on an earlier sub-card's output, declare it in "after": a list of earlier sub-card INDEXES (0-based). Only real data-dependencies — parallel work should have no "after".
- Respond with ONLY a JSON array, no prose, no code fences:
[{"title":"...","detail":"...","assignees":["Member Name"],"after":[0]}]
"detail" is the full working instruction: context, requirements, acceptance criteria.`

	user := fmt.Sprintf("## PROJECT GOAL\n%s\n\n## PROJECT BRIEF\n%s\n\n## TEAM\n%s\n## THE CARD TO BREAK DOWN\nTitle: %s\n\n%s",
		strings.TrimSpace(p.Goal), strings.TrimSpace(p.Brief), tb.String(), card.Title, strings.TrimSpace(card.Detail))

	runCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	out, _, _, _, err := w.exec.RunCrewTask(runCtx, userID, CrewModel, sys, user, nil)
	if err != nil {
		return nil, fmt.Errorf("planner run failed: %w", err)
	}
	var proposed []struct {
		Title     string   `json:"title"`
		Detail    string   `json:"detail"`
		Assignees []string `json:"assignees"`
		After     []int    `json:"after"`
	}
	if err := json.Unmarshal([]byte(extractJSONArray(out)), &proposed); err != nil {
		return nil, fmt.Errorf("planner returned an unreadable plan — try again")
	}
	nameToID := map[string]string{}
	for _, m := range members {
		nameToID[strings.ToLower(strings.TrimSpace(m.Name))] = m.ID.Hex()
	}
	created := []models.CrewCard{}
	createdIDs := []string{}
	for i, pc := range proposed {
		if i >= 6 || strings.TrimSpace(pc.Title) == "" {
			break
		}
		var assignees []string
		for _, n := range pc.Assignees {
			if id, okA := nameToID[strings.ToLower(strings.TrimSpace(n))]; okA {
				assignees = append(assignees, id)
			}
		}
		// Resolve "after" indexes to the ids of already-created sub-cards.
		var deps []string
		for _, ai := range pc.After {
			if ai >= 0 && ai < len(createdIDs) {
				deps = append(deps, createdIDs[ai])
			}
		}
		// Sub-cards inherit the parent's objective (same slice of the goal).
		sub, err := w.svc.CreateCard(ctx, userID, card.ProjectID, pc.Title, pc.Detail, card.ObjectiveID, "", assignees, deps)
		if err != nil {
			createdIDs = append(createdIDs, "") // keep indexes aligned
			continue
		}
		created = append(created, *sub)
		createdIDs = append(createdIDs, sub.ID.Hex())
	}
	if len(created) == 0 {
		return nil, fmt.Errorf("the planner produced no usable sub-cards — try again")
	}
	log.Printf("🧭 [CREW] card %q broken into %d sub-cards", card.Title, len(created))
	return created, nil
}
