package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"claraverse/internal/models"

	"go.mongodb.org/mongo-driver/bson"
)

// The coordinator turns a multi-assignee card into a real team effort instead
// of N independent duplicates:
//
//	split   — a team-lead pass divides the task into per-member sub-briefs
//	work    — each member runs ONLY their part (existing claim machinery)
//	combine — when every part is in, a synthesis pass merges them into ONE
//	          deliverable, which is what reaches human review
//
// Both coordinator passes are guarded by at-most-once locks on the card
// (stale after 10 min), and every failure surfaces on the card — never a
// silent retry loop.

const coordLockStale = 10 * time.Minute

// depContext gathers the approved outputs of a card's dependencies so the
// coordinator passes work from the same input data the members will get.
func (w *CrewWorker) depContext(ctx context.Context, card *models.CrewCard, perDep int) string {
	if len(card.DependsOn) == 0 {
		return ""
	}
	var b strings.Builder
	for _, depID := range card.DependsOn {
		dep, err := w.svc.GetCard(ctx, depID, card.UserID)
		if err != nil || dep.Status != models.CardDone || strings.TrimSpace(dep.LatestOutput) == "" {
			continue
		}
		fmt.Fprintf(&b, "\n### From %q:\n%s\n", dep.Title, truncateCrew(dep.LatestOutput, perDep))
	}
	if b.Len() == 0 {
		return ""
	}
	return "\n## INPUT FROM COMPLETED DEPENDENCY TASKS (this task builds on it)\n" + b.String()
}

// handleSplits finds queued multi-assignee cards that still need a work split
// and runs the team-lead pass for each (bounded by the worker's slot pool).
func (w *CrewWorker) handleSplits(ctx context.Context) {
	cur, err := w.svc.cards.Find(ctx, bson.M{
		"status":        models.CardQueued,
		"assigneeIds.1": bson.M{"$exists": true},
		"parts":         bson.M{"$exists": false},
	})
	if err != nil {
		return
	}
	var due []models.CrewCard
	if cur.All(ctx, &due) != nil {
		return
	}
	for i := range due {
		card := due[i]
		// At-most-once claim (or steal a stale lock from a crashed run).
		res, err := w.svc.cards.UpdateOne(ctx,
			bson.M{"_id": card.ID, "$or": bson.A{
				bson.M{"splitLock": bson.M{"$exists": false}},
				bson.M{"splitLock": bson.M{"$lt": time.Now().Add(-coordLockStale)}},
			}},
			bson.M{"$set": bson.M{"splitLock": time.Now()}})
		if err != nil || res.ModifiedCount == 0 {
			continue
		}
		go func() {
			select {
			case w.slots <- struct{}{}:
				defer func() { <-w.slots }()
			case <-time.After(5 * time.Second):
				_, _ = w.svc.cards.UpdateOne(ctx, bson.M{"_id": card.ID}, bson.M{"$unset": bson.M{"splitLock": ""}})
				return
			}
			w.runSplit(ctx, &card)
		}()
	}
}

func (w *CrewWorker) runSplit(ctx context.Context, card *models.CrewCard) {
	runCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	fail := func(msg string) {
		// Surface, don't self-heal: back to draft with the error visible.
		_, _ = w.svc.cards.UpdateOne(ctx, bson.M{"_id": card.ID}, bson.M{
			"$set":   bson.M{"status": models.CardDraft, "error": msg, "awaiting": []string{}, "running": []string{}, "updatedAt": time.Now()},
			"$unset": bson.M{"splitLock": ""},
		})
	}
	members, err := w.svc.ListMembers(runCtx, card.UserID, card.ProjectID)
	if err != nil {
		fail("work split failed: " + err.Error())
		return
	}
	byID := map[string]*models.CrewMember{}
	var tb strings.Builder
	for i := range members {
		m := &members[i]
		byID[m.ID.Hex()] = m
	}
	assigned := []*models.CrewMember{}
	for _, aid := range card.AssigneeIDs {
		if m, ok := byID[aid]; ok {
			assigned = append(assigned, m)
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
	}
	if len(assigned) < 2 {
		// Degenerate — let the normal single flow handle it.
		_, _ = w.svc.cards.UpdateOne(ctx, bson.M{"_id": card.ID}, bson.M{"$unset": bson.M{"splitLock": ""}})
		return
	}

	sys := `You have NO tools in this pass — respond with plain text only; never emit tool-call syntax.

You are the team lead dividing ONE task among your team so their combined work fully covers it with no duplication.

Rules:
- Give EVERY listed member exactly one part, matched to their role and lane.
- Parts must be complementary — together they cover the whole task; no two members doing the same thing.
- CRITICAL: all parts run IN PARALLEL. A part must NEVER depend on another member's not-yet-produced output ("based on X's findings" is forbidden). Divide by AREA of the task, not by stage of a pipeline. If the task is inherently sequential, give each member a self-contained slice they can complete from the task input alone.
- Each part is a complete working brief: what to produce, what to focus on, what NOT to cover (it's a teammate's part).
- Respond with ONLY a JSON array, no prose, no code fences:
[{"member":"Exact Member Name","brief":"..."}]`
	user := fmt.Sprintf("## THE TASK\n%s\n\n%s\n%s\n## THE TEAM (every member must get a part)\n%s",
		card.Title, strings.TrimSpace(card.Detail), w.depContext(runCtx, card, 2500), tb.String())

	out, _, tin, tout, err := w.exec.RunCrewTask(runCtx, card.UserID, CrewModel, sys, user, nil)
	if err != nil {
		fail("work split failed: " + err.Error())
		return
	}
	var proposed []struct {
		Member string `json:"member"`
		Brief  string `json:"brief"`
	}
	if json.Unmarshal([]byte(extractJSONArray(out)), &proposed) != nil {
		fail("work split failed: the team lead returned an unreadable plan — re-queue to retry")
		return
	}
	nameToID := map[string]string{}
	for _, m := range assigned {
		nameToID[strings.ToLower(strings.TrimSpace(m.Name))] = m.ID.Hex()
	}
	parts := map[string]string{}
	for _, p := range proposed {
		if id, ok := nameToID[strings.ToLower(strings.TrimSpace(p.Member))]; ok && strings.TrimSpace(p.Brief) != "" {
			parts[id] = strings.TrimSpace(p.Brief)
		}
	}
	// Every assignee works — anyone the plan missed contributes their role's view.
	for _, m := range assigned {
		if _, ok := parts[m.ID.Hex()]; !ok {
			parts[m.ID.Hex()] = "Contribute your role's specific expertise to the task. Focus on what only you cover; teammates handle the rest."
		}
	}
	_, err = w.svc.cards.UpdateOne(ctx, bson.M{"_id": card.ID}, bson.M{
		"$set":   bson.M{"parts": parts, "error": "", "updatedAt": time.Now()},
		"$unset": bson.M{"splitLock": ""},
	})
	if err == nil {
		log.Printf("🧩 [CREW] split %q into %d parts (%d tokens)", card.Title, len(parts), tin+tout)
	}
}

// handleCombines finds multi-part cards whose every member pass has landed and
// runs the synthesis pass that produces the single team deliverable.
func (w *CrewWorker) handleCombines(ctx context.Context) {
	cur, err := w.svc.cards.Find(ctx, bson.M{
		"status":   models.CardWorking,
		"parts":    bson.M{"$exists": true},
		"awaiting": bson.M{"$size": 0},
		"running":  bson.M{"$size": 0},
	})
	if err != nil {
		return
	}
	var due []models.CrewCard
	if cur.All(ctx, &due) != nil {
		return
	}
	for i := range due {
		card := due[i]
		res, err := w.svc.cards.UpdateOne(ctx,
			bson.M{"_id": card.ID, "$or": bson.A{
				bson.M{"combineLock": bson.M{"$exists": false}},
				bson.M{"combineLock": bson.M{"$lt": time.Now().Add(-coordLockStale)}},
			}},
			bson.M{"$set": bson.M{"combineLock": time.Now()}})
		if err != nil || res.ModifiedCount == 0 {
			continue
		}
		go func() {
			select {
			case w.slots <- struct{}{}:
				defer func() { <-w.slots }()
			case <-time.After(5 * time.Second):
				_, _ = w.svc.cards.UpdateOne(ctx, bson.M{"_id": card.ID}, bson.M{"$unset": bson.M{"combineLock": ""}})
				return
			}
			w.runCombine(ctx, &card)
		}()
	}
}

func (w *CrewWorker) runCombine(ctx context.Context, card *models.CrewCard) {
	runCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	// The parts from THIS cycle: everything after the last combined revision.
	start := 0
	for i, r := range card.Revisions {
		if r.MemberName == "Team deliverable" {
			start = i + 1
		}
	}
	cycle := card.Revisions[start:]
	if len(cycle) == 0 {
		_, _ = w.svc.cards.UpdateOne(ctx, bson.M{"_id": card.ID}, bson.M{"$unset": bson.M{"combineLock": ""}})
		return
	}
	toReview := func(errNote string) {
		// Even if synthesis fails, the human sees the raw parts in review.
		set := bson.M{"status": models.CardReview, "updatedAt": time.Now()}
		if errNote != "" {
			set["error"] = errNote
		}
		_, _ = w.svc.cards.UpdateOne(ctx, bson.M{"_id": card.ID}, bson.M{"$set": set})
	}
	project, err := w.svc.GetProject(runCtx, card.ProjectID, card.UserID)
	if err != nil {
		toReview("combine skipped: " + err.Error())
		return
	}
	var pb strings.Builder
	for _, r := range cycle {
		fmt.Fprintf(&pb, "\n### Part by %s:\n%s\n", r.MemberName, truncateCrew(r.Output, 9000))
	}
	feedback := ""
	for _, r := range card.Revisions[:start] {
		if strings.TrimSpace(r.Review) != "" {
			feedback += "- " + strings.TrimSpace(r.Review) + "\n"
		}
	}
	fbSection := ""
	if feedback != "" {
		fbSection = "\n## EARLIER REVIEWER FEEDBACK (the combined result MUST address it)\n" + feedback
	}

	sys := `You have NO tools in this pass — respond with plain text only; never emit tool-call syntax. Work ONLY from the parts given below; do not research anything new.

You are the team lead assembling your teammates' parts into ONE final deliverable.

Rules:
- Merge, don't concatenate: one coherent piece with a single voice, structure and flow — as if one excellent author wrote it.
- Use everything of value from every part; resolve overlaps and contradictions (say which source you kept if it matters).
- GUARD AGAINST FABRICATION: if a part invents specifics that cannot be verified from sources — names of real people, salary/statistics, review counts — do NOT present them as fact. Replace with [PLACEHOLDER — needs real data] and keep them in a "Data needed" list. Placeholders are professional; invented facts are a firing offense.
- One H1 title, an executive summary up top, clean sections, key numbers bold. Keep citations/sources from the parts.
- Do not mention the splitting process or the teammates — the reader sees only the finished work.`
	user := fmt.Sprintf("## PROJECT BRIEF\n%s\n\n## THE TASK\n%s\n\n%s%s%s\n## THE PARTS TO COMBINE\n%s",
		strings.TrimSpace(project.Brief), card.Title, strings.TrimSpace(card.Detail), w.depContext(runCtx, card, 1500), fbSection, pb.String())

	out, _, tin, tout, err := w.exec.RunCrewTask(runCtx, card.UserID, CrewModel, sys, user, nil)
	if err != nil || strings.TrimSpace(out) == "" {
		msg := "combine failed — showing the individual parts instead"
		if err != nil {
			msg += ": " + err.Error()
		}
		toReview(msg)
		return
	}
	rev := models.CardRevision{MemberName: "Team deliverable", Output: out, TokensUsed: tin + tout, At: time.Now()}
	_, err = w.svc.cards.UpdateOne(ctx, bson.M{"_id": card.ID, "status": models.CardWorking}, bson.M{
		"$push": bson.M{"revisions": rev},
		"$set":  bson.M{"latestOutput": out, "status": models.CardReview, "error": "", "updatedAt": time.Now()},
	})
	if err != nil {
		return
	}
	log.Printf("🧵 [CREW] combined %d parts of %q → review (%d tokens)", len(cycle), card.Title, tin+tout)
	if w.Notify != nil {
		w.Notify(card.UserID, project.Name, card.Title)
	}
}
