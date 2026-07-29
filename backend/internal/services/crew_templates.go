package services

import (
	"context"
	"fmt"
	"time"

	"claraverse/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Team templates (Paperclip's "companies" pattern, first-party curated only —
// third-party template packs are a prompt-injection vector, so these live in
// code and ship with the server). Each template seeds a full working team:
// goal scaffold, objectives, members with charters, and starter draft cards
// already linked to objectives and assignees. Nothing runs until the user
// queues a card — templates are safe to explore.

type TplMember struct {
	RoleKey string `json:"role_key"`
	Name    string `json:"name"`
	Charter string `json:"charter"`
	// Tools overrides the role's defaults when set (must be within AllowedTools).
	Tools []string `json:"tools,omitempty"`
}

type TplCard struct {
	Title     string `json:"title"`
	Detail    string `json:"detail"`
	ObjIdx    int    `json:"obj_idx"`    // index into Objectives (-1 = none)
	MemberIdx []int  `json:"member_idx"` // indexes into Members
	Repeat    string `json:"repeat,omitempty"`
}

type CrewTemplate struct {
	Key        string      `json:"key"`
	Name       string      `json:"name"`
	Blurb      string      `json:"blurb"`
	Goal       string      `json:"goal"` // scaffold — user should edit
	Objectives []string    `json:"objectives"`
	Members    []TplMember `json:"members"`
	Cards      []TplCard   `json:"cards"`
}

var crewTemplates = []CrewTemplate{
	{
		Key:   "content-team",
		Name:  "Content Team",
		Blurb: "Researcher → writer → SEO: a weekly content engine with everything stopping at your review.",
		Goal:  "Publish consistent, well-researched content that grows organic traffic. (Edit this: what topic, for whom, and what number makes it a win?)",
		Objectives: []string{
			"Build the research foundation",
			"Draft and refine content",
			"Optimize and distribute",
		},
		Members: []TplMember{
			{RoleKey: "researcher", Name: "Researcher", Charter: "Owns all fact-finding and source-backed research briefs. Does NOT write final copy — produces material the writer builds on."},
			{RoleKey: "content-writer", Name: "Content Writer", Charter: "Owns drafting articles, newsletters and posts in the project voice, building on approved research. Hands SEO tuning to the SEO Specialist."},
			{RoleKey: "seo", Name: "SEO Specialist", Charter: "Owns keyword research, on-page recommendations and SERP analysis. Does NOT rewrite content wholesale — proposes concrete optimizations."},
		},
		Cards: []TplCard{
			{Title: "Research our audience and the 5 topics they care about most", Detail: "Deliver a cited brief: who reads us, what they search for, and the 5 highest-value topics with evidence.", ObjIdx: 0, MemberIdx: []int{0}},
			{Title: "Keyword landscape for our niche", Detail: "Map the keywords worth targeting: volume/difficulty reasoning, our realistic winners, and the content each needs.", ObjIdx: 0, MemberIdx: []int{2}},
			{Title: "Draft article #1 from the approved research", Detail: "Once the research brief is approved, write the first full article (1,500–2,500 words) with title options and a meta description.", ObjIdx: 1, MemberIdx: []int{1}},
			{Title: "Weekly content ideas digest", Detail: "Every week: 5 fresh content ideas grounded in what's trending in our space, each with a one-line angle and target keyword.", ObjIdx: 0, MemberIdx: []int{0}, Repeat: "weekly"},
		},
	},
	{
		Key:   "research-lab",
		Name:  "Research Lab",
		Blurb: "Two researchers and an analyst producing cited reports and data-backed findings.",
		Goal:  "Deliver decision-ready research reports. (Edit this: what decisions, for whom, by when?)",
		Objectives: []string{
			"Market and competitor intelligence",
			"Data analysis and synthesis",
		},
		Members: []TplMember{
			{RoleKey: "researcher", Name: "Lead Researcher", Charter: "Owns primary web research and competitive intelligence. Every claim cited."},
			{RoleKey: "researcher", Name: "Second Researcher", Charter: "Owns deep-dives and verification: takes a teammate's findings and independently pressure-tests the load-bearing claims."},
			{RoleKey: "data-analyst", Name: "Data Analyst", Charter: "Owns anything numeric: datasets, calculations, trend analysis. Findings lead with the numbers."},
		},
		Cards: []TplCard{
			{Title: "Full competitive teardown", Detail: "Pick our top 3 competitors: positioning, pricing, recent moves, weaknesses we can exploit. Cited report.", ObjIdx: 0, MemberIdx: []int{0}},
			{Title: "Verify and stress-test the teardown", Detail: "Independently verify the approved teardown's key claims; flag anything that doesn't hold.", ObjIdx: 0, MemberIdx: []int{1}},
			{Title: "Weekly market watch", Detail: "Every week: what changed in our market — launches, pricing moves, notable content. Short, cited, opinionated.", ObjIdx: 0, MemberIdx: []int{0}, Repeat: "weekly"},
		},
	},
	{
		Key:   "seo-squad",
		Name:  "SEO Squad",
		Blurb: "Audit, keywords, and content fixes for a site you name in the brief.",
		Goal:  "Grow organic search traffic to the site named in the brief. (Edit this: which site, which pages matter, what growth number?)",
		Objectives: []string{
			"Audit and fix the foundations",
			"Win the target keywords",
		},
		Members: []TplMember{
			{RoleKey: "seo", Name: "SEO Lead", Charter: "Owns audits, keyword strategy and prioritized recommendations. Every recommendation actionable and ranked by impact."},
			{RoleKey: "researcher", Name: "Researcher", Charter: "Owns competitor content research: what ranks above us and why, page by page."},
			{RoleKey: "content-writer", Name: "Content Writer", Charter: "Owns rewrites and new copy implementing the SEO Lead's approved recommendations."},
		},
		Cards: []TplCard{
			{Title: "Technical & on-page SEO audit", Detail: "Audit the site in the project brief: what's holding rankings back, ranked by impact, with concrete fixes.", ObjIdx: 0, MemberIdx: []int{0}},
			{Title: "Who outranks us and why", Detail: "For our 5 most important queries: analyze the pages that outrank us and what they do that we don't.", ObjIdx: 1, MemberIdx: []int{1}},
			{Title: "Monthly rank & content review", Detail: "Every month: re-check our target queries, what moved, and the next 3 highest-impact actions.", ObjIdx: 1, MemberIdx: []int{0}, Repeat: "monthly"},
		},
	},
	{
		Key:   "ops-assistant",
		Name:  "Ops Assistant",
		Blurb: "A generalist wired into your integrations for recurring operational work — inbox digests, CRM checks, reports.",
		Goal:  "Keep operational busywork handled and reviewed without anyone chasing it. (Edit this: which systems, what cadence, what matters?)",
		Objectives: []string{
			"Recurring operations",
			"Ad-hoc requests",
		},
		Members: []TplMember{
			{RoleKey: "generalist", Name: "Ops Generalist", Charter: "Owns recurring operational tasks across every connected integration. Summarizes and drafts; never sends or commits anything externally without it being reviewed first.", Tools: []string{
				"search_web", "scrape_web", "get_current_time", "read_document", "read_data_file",
				"gmail_fetch_emails", "gmail_get_message", "gmail_list_labels", "gmail_create_draft", "gmail_list_drafts",
				"googlecalendar_list_events", "googlesheets_read",
			}},
		},
		Cards: []TplCard{
			{Title: "Weekly inbox digest with draft replies", Detail: "Every week: summarize the important email threads and draft a reply for each one that needs a response. Nothing is sent — replies come to review.", ObjIdx: 0, MemberIdx: []int{0}, Repeat: "weekly"},
			{Title: "Daily priorities snapshot", Detail: "Every day: a short morning brief — what's new across connected systems and the 3 things worth attention today.", ObjIdx: 0, MemberIdx: []int{0}, Repeat: "daily"},
		},
	},
}

// ListTemplates returns the curated template catalog.
func (s *CrewService) ListTemplates() []CrewTemplate {
	return crewTemplates
}

// CreateProjectFromTemplate spins up a full team in one call: project with
// goal + objectives, members with role tools and charters, and starter cards
// as DRAFTS (nothing runs until the user queues them).
func (s *CrewService) CreateProjectFromTemplate(ctx context.Context, userID, templateKey, name, brief string) (*models.CrewProject, error) {
	var tpl *CrewTemplate
	for i := range crewTemplates {
		if crewTemplates[i].Key == templateKey {
			tpl = &crewTemplates[i]
			break
		}
	}
	if tpl == nil {
		return nil, fmt.Errorf("unknown template %q", templateKey)
	}
	if name == "" {
		name = tpl.Name
	}
	p, err := s.CreateProject(ctx, userID, name, brief)
	if err != nil {
		return nil, err
	}
	pid := p.ID.Hex()

	objs := make([]models.CrewObjective, len(tpl.Objectives))
	for i, t := range tpl.Objectives {
		objs[i] = models.CrewObjective{ID: fmt.Sprintf("obj-%d-%d", time.Now().UnixNano(), i), Title: t}
	}
	oid, _ := primitive.ObjectIDFromHex(pid)
	if _, err := s.projects.UpdateOne(ctx, bson.M{"_id": oid},
		bson.M{"$set": bson.M{"goal": tpl.Goal, "objectives": objs, "updatedAt": time.Now()}}); err != nil {
		return nil, err
	}

	memberIDs := make([]string, 0, len(tpl.Members))
	for _, tm := range tpl.Members {
		role := s.RoleByKey(tm.RoleKey)
		if role == nil {
			continue
		}
		hireTools := role.DefaultTools
		if len(tm.Tools) > 0 {
			hireTools = tm.Tools
		}
		m, err := s.HireMember(ctx, userID, pid, tm.RoleKey, tm.Name, hireTools, "")
		if err != nil {
			return nil, fmt.Errorf("hiring %s: %w", tm.Name, err)
		}
		if tm.Charter != "" {
			_ = s.SetMemberCharter(ctx, m.ID.Hex(), userID, tm.Charter)
		}
		memberIDs = append(memberIDs, m.ID.Hex())
	}

	for _, tc := range tpl.Cards {
		objectiveID := ""
		if tc.ObjIdx >= 0 && tc.ObjIdx < len(objs) {
			objectiveID = objs[tc.ObjIdx].ID
		}
		var assignees []string
		for _, mi := range tc.MemberIdx {
			if mi >= 0 && mi < len(memberIDs) {
				assignees = append(assignees, memberIDs[mi])
			}
		}
		if _, err := s.CreateCard(ctx, userID, pid, tc.Title, tc.Detail, objectiveID, tc.Repeat, assignees, nil); err != nil {
			return nil, fmt.Errorf("creating card %q: %w", tc.Title, err)
		}
	}
	return s.GetProject(ctx, pid, userID)
}
