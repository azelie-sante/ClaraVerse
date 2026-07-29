package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"claraverse/internal/database"
	"claraverse/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CrewService persists Crew projects, members and cards, and enforces the
// pipeline: draft → queued → working → review → done (review can re-queue).
type CrewService struct {
	projects *mongo.Collection
	members  *mongo.Collection
	cards    *mongo.Collection
	skills   *mongo.Collection // read-only: attach skills to members
}

func NewCrewService(db *database.MongoDB) *CrewService {
	return &CrewService{
		projects: db.Collection("crew_projects"),
		members:  db.Collection("crew_members"),
		cards:    db.Collection("crew_cards"),
		skills:   db.Collection(database.CollectionSkills),
	}
}

// EnsureIndexes creates the indexes the pipeline relies on.
func (s *CrewService) EnsureIndexes(ctx context.Context) error {
	if _, err := s.projects.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "userId", Value: 1}, {Key: "updatedAt", Value: -1}},
	}); err != nil {
		return err
	}
	if _, err := s.members.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "projectId", Value: 1}},
	}); err != nil {
		return err
	}
	_, err := s.cards.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "projectId", Value: 1}, {Key: "status", Value: 1}, {Key: "updatedAt", Value: -1}},
	})
	return err
}

// ─── Role catalog ─────────────────────────────────────────────────────────────
// The ONLY roles users can hire, each with a hard tool superset. A member never
// gets a tool outside its role's AllowedTools, and never outside what the user
// checked at hire time.

var crewRoles = []models.CrewRole{
	{
		Key: "researcher", Label: "Researcher", Blurb: "Web research, fact-finding, source-backed summaries.",
		Prompt: "You are a meticulous research analyst. Research thoroughly using web search and page reading, cross-check claims across sources, and produce a structured, source-cited write-up. Never fabricate; say what you could not verify.",
		AllowedTools: []string{"search_web", "scrape_web", "search_images"},
		DefaultTools: []string{"search_web", "scrape_web"},
	},
	{
		Key: "content-writer", Label: "Content Writer", Blurb: "Articles, posts, copy — drafted to the project's voice.",
		Prompt: "You are a senior content writer. Write in the project's voice and for its audience. Structure clearly, open strong, no filler, no clichés. When facts are needed, research them first and cite sources.",
		AllowedTools: []string{"search_web", "scrape_web", "generate_image"},
		DefaultTools: []string{"search_web"},
	},
	{
		Key: "data-analyst", Label: "Data Analyst", Blurb: "Datasets, calculations, findings with the numbers up front.",
		Prompt: "You are a careful data analyst. Read the provided data precisely, compute with Python when arithmetic matters, and present findings with the key numbers up front. Show your method briefly.",
		AllowedTools: []string{"read_data_file", "read_spreadsheet", "run_python", "analyze_data", "calculate_math", "search_web"},
		DefaultTools: []string{"read_data_file", "run_python"},
	},
	{
		Key: "designer", Label: "Visual Designer", Blurb: "Image generation and editing for the project's brand.",
		Prompt: "You are an art director. Generate and refine imagery that fits the project's brand and mood — deliberate palettes, strong composition. Describe what you produced and why it fits.",
		AllowedTools: []string{"generate_image", "edit_image", "search_images", "search_web"},
		DefaultTools: []string{"generate_image", "search_images"},
	},
	{
		Key: "outreach", Label: "Outreach & Comms", Blurb: "Drafts emails and messages via your connected accounts.",
		Prompt: "You draft professional outreach and internal communications. Match tone to the recipient, keep it tight, and always produce a DRAFT for human review — you never send anything yourself.",
		AllowedTools: []string{"gmail_create_draft", "gmail_fetch_emails", "gmail_list_drafts", "search_web"},
		DefaultTools: []string{"search_web", "gmail_create_draft"},
	},
	{
		Key: "developer", Label: "Developer", Blurb: "Writes code, scripts and technical docs as deliverables.",
		Prompt: "You are a senior software engineer. Deliver complete, runnable code with brief setup notes — full files in fenced code blocks, no placeholders or TODOs. Verify logic with Python when it helps. State assumptions explicitly.",
		AllowedTools: []string{"run_python", "search_web", "scrape_web", "read_data_file"},
		DefaultTools: []string{"run_python", "search_web"},
	},
	{
		Key: "seo", Label: "SEO Specialist", Blurb: "Keyword research, content optimization, SERP analysis.",
		Prompt: "You are an SEO specialist. Research keywords and competing pages, then deliver concrete recommendations: target keywords with intent, title/meta suggestions, content structure, and internal-linking ideas. Ground every claim in what you actually found.",
		AllowedTools: []string{"search_web", "scrape_web"},
		DefaultTools: []string{"search_web", "scrape_web"},
	},
	{
		Key: "social", Label: "Social Media Manager", Blurb: "Platform-native posts, calendars and hooks.",
		Prompt: "You are a social media manager. Produce platform-native content: hooks that stop the scroll, correct lengths per platform, hashtag strategy, and a posting rationale. Draft only — a human approves before anything is posted.",
		AllowedTools: []string{"search_web", "x_search_posts", "x_get_user_posts", "youtube_search_videos", "generate_image"},
		DefaultTools: []string{"search_web", "generate_image"},
	},
	{
		Key: "document-analyst", Label: "Document Analyst", Blurb: "Reads reports, PDFs and files; extracts what matters.",
		Prompt: "You are a document analyst. Read the provided documents fully, extract the decisions, numbers, obligations and risks, and deliver a faithful structured summary with page/section references. Quote exactly where wording matters.",
		AllowedTools: []string{"read_document", "read_data_file", "read_spreadsheet", "transcribe_audio", "search_web"},
		DefaultTools: []string{"read_document", "read_data_file"},
	},
	{
		Key: "editor", Label: "Editor / Translator", Blurb: "Polishes, restructures, translates — keeps the meaning.",
		Prompt: "You are a demanding editor and translator. Improve clarity, rhythm and correctness without changing meaning. For translations, be idiomatic, not literal. Return the full edited text plus a short list of the significant changes you made.",
		AllowedTools: []string{"search_web"},
		DefaultTools: []string{"search_web"},
	},
	{
		Key: "planner", Label: "Planner / Coordinator", Blurb: "Breaks goals into plans, timelines and task lists.",
		Prompt: "You are a pragmatic project planner. Turn goals into concrete plans: phases, tasks with owners-by-role, dependencies, timeline estimates and risks. Suggest which crew roles should take which task. Be realistic, not optimistic.",
		AllowedTools: []string{"search_web", "calculate_math"},
		DefaultTools: []string{"search_web"},
	},
	{
		Key: "generalist", Label: "Generalist", Blurb: "The all-rounder — can use every integration you've connected.",
		Prompt: "You are a dependable generalist operator with broad access to the user's connected tools and integrations. Understand the task in the project's context, pick the RIGHT tools for it (don't spray), and deliver complete, review-ready work. For any outward-facing action (sending messages, posting, modifying external systems), prefer producing a draft/plan in your deliverable unless the task explicitly instructs you to execute it.",
		AllowedTools: []string{
			// Core research & files
			"search_web", "scrape_web", "search_images", "generate_image", "edit_image", "describe_image",
			"transcribe_audio", "read_document", "read_data_file", "read_spreadsheet", "run_python",
			"analyze_data", "calculate_math", "create_document", "create_presentation", "create_text_file",
			"html_to_pdf", "download_file", "get_current_time", "api_request",
			// Gmail
			"gmail_fetch_emails", "gmail_get_message", "gmail_list_labels", "gmail_add_label",
			"gmail_create_draft", "gmail_list_drafts", "gmail_reply_to_thread", "gmail_send_draft", "gmail_send_email",
			// Google Calendar
			"googlecalendar_list_calendars", "googlecalendar_list_events", "googlecalendar_get_event",
			"googlecalendar_find_free_slots", "googlecalendar_create_event", "googlecalendar_quick_add_event",
			"googlecalendar_update_event",
			// Google Drive
			"googledrive_list_files", "googledrive_search_files", "googledrive_get_file",
			"googledrive_download_file", "googledrive_create_folder", "googledrive_copy_file",
			// Google Sheets
			"googlesheets_create", "googlesheets_get_info", "googlesheets_list_sheets", "googlesheets_read",
			"googlesheets_search", "googlesheets_append", "googlesheets_write", "googlesheets_upsert_rows",
			"googlesheets_find_replace", "googlesheets_add_sheet",
			// Messaging
			"send_slack_message", "send_discord_message", "send_teams_message", "send_google_chat_message",
			"send_telegram_message", "twilio_send_sms", "twilio_send_whatsapp",
			// Social
			"x_search_posts", "x_get_user", "x_get_user_posts", "x_post_tweet",
			"linkedin_get_my_info", "linkedin_get_company_info", "linkedin_create_post",
			"youtube_search_videos", "youtube_get_video", "youtube_get_channel", "youtube_get_comments",
			// Project management
			"notion_search", "notion_query_database", "notion_create_page", "notion_update_page",
			"trello_boards", "trello_lists", "trello_cards", "trello_create_card",
			"clickup_tasks", "clickup_create_task", "clickup_update_task",
			"jira_issues", "jira_create_issue", "jira_update_issue",
			"linear_issues", "linear_create_issue", "linear_update_issue",
			"airtable_list", "airtable_read", "airtable_create", "airtable_update",
			"github_get_repo", "github_list_issues", "github_create_issue", "github_add_comment",
			// CRM, commerce & analytics
			"hubspot_contacts", "hubspot_companies", "hubspot_deals",
			"leadsquared_leads", "leadsquared_activities", "leadsquared_lists",
			"shopify_products", "shopify_orders", "shopify_customers",
			"mailchimp_lists", "mailchimp_add_subscriber",
			// Meetings & design
			"zoom_list_meetings", "zoom_get_meeting", "zoom_create_meeting",
			"calendly_events", "calendly_event_types", "calendly_invitees",
			"canva_list_designs", "canva_get_design", "canva_create_design", "canva_export_design",
		},
		DefaultTools: []string{"search_web", "scrape_web", "run_python", "read_document", "read_data_file", "gmail_fetch_emails", "googlesheets_read", "googlecalendar_list_events"},
	},
}

func (s *CrewService) Roles() []models.CrewRole { return crewRoles }

// RoleByKey returns the catalog entry (nil if unknown).
func (s *CrewService) RoleByKey(key string) *models.CrewRole {
	for i := range crewRoles {
		if crewRoles[i].Key == key {
			return &crewRoles[i]
		}
	}
	return nil
}

// ─── Projects ────────────────────────────────────────────────────────────────

func (s *CrewService) CreateProject(ctx context.Context, userID, name, brief string) (*models.CrewProject, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("project name is required")
	}
	now := time.Now()
	p := &models.CrewProject{UserID: userID, Name: name, Brief: strings.TrimSpace(brief), Status: "active", CreatedAt: now, UpdatedAt: now}
	res, err := s.projects.InsertOne(ctx, p)
	if err != nil {
		return nil, err
	}
	p.ID = res.InsertedID.(primitive.ObjectID)
	return p, nil
}

func (s *CrewService) ListProjects(ctx context.Context, userID string) ([]models.CrewProject, error) {
	cur, err := s.projects.Find(ctx,
		bson.M{"userId": userID, "status": bson.M{"$ne": "archived"}},
		options.Find().SetSort(bson.M{"updatedAt": -1}))
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	out := []models.CrewProject{}
	return out, cur.All(ctx, &out)
}

func (s *CrewService) GetProject(ctx context.Context, id, userID string) (*models.CrewProject, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("project not found")
	}
	var p models.CrewProject
	if err := s.projects.FindOne(ctx, bson.M{"_id": oid, "userId": userID}).Decode(&p); err != nil {
		return nil, fmt.Errorf("project not found")
	}
	return &p, nil
}

func (s *CrewService) UpdateProject(ctx context.Context, id, userID string, set bson.M) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("project not found")
	}
	set["updatedAt"] = time.Now()
	res, err := s.projects.UpdateOne(ctx, bson.M{"_id": oid, "userId": userID}, bson.M{"$set": set})
	if err == nil && res.MatchedCount == 0 {
		return fmt.Errorf("project not found")
	}
	return err
}

// ─── Members ─────────────────────────────────────────────────────────────────

// HireMember adds an agent to a project. Tools are intersected with the role's
// AllowedTools — the hard cap on what this agent can ever do.
func (s *CrewService) HireMember(ctx context.Context, userID, projectID, roleKey, name string, tools []string, model string) (*models.CrewMember, error) {
	if _, err := s.GetProject(ctx, projectID, userID); err != nil {
		return nil, err
	}
	role := s.RoleByKey(roleKey)
	if role == nil {
		return nil, fmt.Errorf("unknown role %q", roleKey)
	}
	allowed := map[string]bool{}
	for _, t := range role.AllowedTools {
		allowed[t] = true
	}
	granted := []string{}
	for _, t := range tools {
		if allowed[t] {
			granted = append(granted, t)
		}
	}
	if len(granted) == 0 {
		granted = append(granted, role.DefaultTools...)
	}
	if strings.TrimSpace(name) == "" {
		name = role.Label
	}
	m := &models.CrewMember{
		ProjectID: projectID, UserID: userID, RoleKey: roleKey, Name: strings.TrimSpace(name),
		Tools: granted, Model: strings.TrimSpace(model), Status: "active", CreatedAt: time.Now(),
	}
	res, err := s.members.InsertOne(ctx, m)
	if err != nil {
		return nil, err
	}
	m.ID = res.InsertedID.(primitive.ObjectID)
	return m, nil
}

func (s *CrewService) ListMembers(ctx context.Context, userID, projectID string) ([]models.CrewMember, error) {
	cur, err := s.members.Find(ctx, bson.M{"projectId": projectID, "userId": userID})
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	out := []models.CrewMember{}
	return out, cur.All(ctx, &out)
}

func (s *CrewService) GetMember(ctx context.Context, id, userID string) (*models.CrewMember, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("member not found")
	}
	var m models.CrewMember
	if err := s.members.FindOne(ctx, bson.M{"_id": oid, "userId": userID}).Decode(&m); err != nil {
		return nil, fmt.Errorf("member not found")
	}
	return &m, nil
}

func (s *CrewService) SetMemberStatus(ctx context.Context, id, userID, status string) error {
	if status != "active" && status != "paused" {
		return fmt.Errorf("invalid status")
	}
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("member not found")
	}
	res, err := s.members.UpdateOne(ctx, bson.M{"_id": oid, "userId": userID}, bson.M{"$set": bson.M{"status": status}})
	if err == nil && res.MatchedCount == 0 {
		return fmt.Errorf("member not found")
	}
	return err
}

func (s *CrewService) FireMember(ctx context.Context, id, userID string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("member not found")
	}
	// Remove them from every card's assignment state (surfaced, not hidden):
	// cards left with no assignees fall back to draft.
	now := time.Now()
	_, _ = s.cards.UpdateMany(ctx,
		bson.M{"userId": userID, "$or": []bson.M{{"assigneeId": id}, {"assigneeIds": id}}},
		bson.M{
			"$pull":  bson.M{"assigneeIds": id, "awaiting": id, "running": id},
			"$set":   bson.M{"updatedAt": now},
			"$unset": bson.M{},
		})
	_, _ = s.cards.UpdateMany(ctx,
		bson.M{"userId": userID, "assigneeId": id},
		bson.M{"$set": bson.M{"assigneeId": ""}})
	_, _ = s.cards.UpdateMany(ctx,
		bson.M{"userId": userID, "status": bson.M{"$in": []string{models.CardQueued, models.CardWorking}},
			"assigneeIds": bson.M{"$size": 0}, "assigneeId": ""},
		bson.M{"$set": bson.M{"status": models.CardDraft, "awaiting": []string{}, "running": []string{}, "updatedAt": now}})
	res, err := s.members.DeleteOne(ctx, bson.M{"_id": oid, "userId": userID})
	if err == nil && res.DeletedCount == 0 {
		return fmt.Errorf("member not found")
	}
	return err
}

// Member reference documents ("RAG in hand") — extracted text stored on the
// member, injected into its runs. Hard caps keep prompts and docs sane.
const (
	maxDocsPerMember = 8
	maxDocChars      = 30_000
)

func (s *CrewService) AddMemberDoc(ctx context.Context, memberID, userID, name, text string) error {
	m, err := s.GetMember(ctx, memberID, userID)
	if err != nil {
		return err
	}
	if len(m.Documents) >= maxDocsPerMember {
		return fmt.Errorf("this member already has %d documents — remove one first", maxDocsPerMember)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return fmt.Errorf("no readable text found in that file")
	}
	if len(text) > maxDocChars {
		text = text[:maxDocChars]
	}
	oid, _ := primitive.ObjectIDFromHex(memberID)
	doc := models.CrewDoc{Name: strings.TrimSpace(name), Text: text, Size: len(text), At: time.Now()}
	_, err = s.members.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$push": bson.M{"documents": doc}})
	return err
}

func (s *CrewService) RemoveMemberDoc(ctx context.Context, memberID, userID string, index int) error {
	m, err := s.GetMember(ctx, memberID, userID)
	if err != nil {
		return err
	}
	if index < 0 || index >= len(m.Documents) {
		return fmt.Errorf("document not found")
	}
	docs := append(append([]models.CrewDoc{}, m.Documents[:index]...), m.Documents[index+1:]...)
	oid, _ := primitive.ObjectIDFromHex(memberID)
	_, err = s.members.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{"documents": docs}})
	return err
}

// UnqueueCard: queued → draft (user drag-back). Only before any member claims it.
func (s *CrewService) UnqueueCard(ctx context.Context, id, userID string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("card not found")
	}
	res, err := s.cards.UpdateOne(ctx,
		bson.M{"_id": oid, "userId": userID, "status": models.CardQueued, "running": bson.M{"$size": 0}},
		bson.M{"$set": bson.M{"status": models.CardDraft, "awaiting": []string{}, "updatedAt": time.Now()}})
	if err == nil && res.MatchedCount == 0 {
		return fmt.Errorf("only queued (not yet claimed) cards can be moved back to drafts")
	}
	return err
}

// ListAvailableSkills returns skills a user may attach (builtin + their own).
func (s *CrewService) ListAvailableSkills(ctx context.Context, userID string) ([]models.Skill, error) {
	cur, err := s.skills.Find(ctx,
		bson.M{"$or": []bson.M{{"is_builtin": true}, {"author_id": userID}}},
		options.Find().SetProjection(bson.M{"system_prompt": 0}).SetSort(bson.M{"name": 1}))
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	out := []models.Skill{}
	return out, cur.All(ctx, &out)
}

// GetSkillsByIDs fetches full skills (with prompts) for a member's run.
func (s *CrewService) GetSkillsByIDs(ctx context.Context, ids []string) []models.Skill {
	if len(ids) == 0 {
		return nil
	}
	oids := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		if oid, err := primitive.ObjectIDFromHex(id); err == nil {
			oids = append(oids, oid)
		}
	}
	cur, err := s.skills.Find(ctx, bson.M{"_id": bson.M{"$in": oids}})
	if err != nil {
		return nil
	}
	defer func() { _ = cur.Close(ctx) }()
	out := []models.Skill{}
	_ = cur.All(ctx, &out)
	return out
}

// SetMemberSkills replaces a member's attached skill list.
func (s *CrewService) SetMemberSkills(ctx context.Context, memberID, userID string, skillIDs []string) error {
	if _, err := s.GetMember(ctx, memberID, userID); err != nil {
		return err
	}
	if len(skillIDs) > 6 {
		return fmt.Errorf("a member can have at most 6 skills")
	}
	oid, _ := primitive.ObjectIDFromHex(memberID)
	_, err := s.members.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{"skillIds": skillIDs}})
	return err
}

// SetMemberCharter sets this member's lane: what they own on the project.
func (s *CrewService) SetMemberCharter(ctx context.Context, memberID, userID, charter string) error {
	if _, err := s.GetMember(ctx, memberID, userID); err != nil {
		return err
	}
	charter = strings.TrimSpace(charter)
	if len(charter) > 2000 {
		charter = charter[:2000]
	}
	oid, _ := primitive.ObjectIDFromHex(memberID)
	_, err := s.members.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{"charter": charter}})
	return err
}

// AddMemberTokens accumulates lifetime usage on a member.
func (s *CrewService) AddMemberTokens(ctx context.Context, memberID string, in, out int64) {
	oid, err := primitive.ObjectIDFromHex(memberID)
	if err != nil {
		return
	}
	mk := time.Now().Format("2006-01")
	// New month → reset the window before incrementing.
	_, _ = s.members.UpdateOne(ctx, bson.M{"_id": oid, "monthKey": bson.M{"$ne": mk}},
		bson.M{"$set": bson.M{"monthKey": mk, "monthTokens": 0}})
	_, _ = s.members.UpdateOne(ctx, bson.M{"_id": oid},
		bson.M{"$inc": bson.M{"tokensIn": in, "tokensOut": out, "monthTokens": in + out}})
}

// SetMemberBudget caps a member's tokens per calendar month (0 = unlimited).
func (s *CrewService) SetMemberBudget(ctx context.Context, memberID, userID string, budget int64) error {
	if _, err := s.GetMember(ctx, memberID, userID); err != nil {
		return err
	}
	if budget < 0 {
		budget = 0
	}
	oid, _ := primitive.ObjectIDFromHex(memberID)
	_, err := s.members.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{"monthlyBudget": budget}})
	return err
}

// MemberOverBudget reports whether a member has exhausted this month's tokens.
func MemberOverBudget(m *models.CrewMember) bool {
	if m.MonthlyBudget <= 0 {
		return false
	}
	if m.MonthKey != time.Now().Format("2006-01") {
		return false // month rolled over — window resets on next run
	}
	return m.MonthTokens >= m.MonthlyBudget
}

// nextRecurrence maps a repeat cadence to the next run time (zero = none).
func nextRecurrence(repeat string, from time.Time) time.Time {
	switch repeat {
	case "daily":
		return from.Add(24 * time.Hour)
	case "weekly":
		return from.Add(7 * 24 * time.Hour)
	case "monthly":
		return from.AddDate(0, 1, 0)
	}
	return time.Time{}
}

// RequeueDueRecurring re-queues approved recurring cards whose time has come.
// Called from the worker tick; each re-queue is atomic (surface, don't retry).
func (s *CrewService) RequeueDueRecurring(ctx context.Context) {
	now := time.Now()
	cur, err := s.cards.Find(ctx, bson.M{
		"status": models.CardDone, "repeat": bson.M{"$in": []string{"daily", "weekly", "monthly"}},
		"nextRunAt": bson.M{"$gt": time.Time{}, "$lte": now},
	}, options.Find().SetLimit(20))
	if err != nil {
		return
	}
	defer func() { _ = cur.Close(ctx) }()
	var due []models.CrewCard
	if cur.All(ctx, &due) != nil {
		return
	}
	for _, c := range due {
		// Archived projects are retired: their recurring cards must never
		// re-queue (weekly cards would silently burn tokens forever).
		if p, err := s.GetProject(ctx, c.ProjectID, c.UserID); err != nil || p.Status == "archived" {
			_, _ = s.cards.UpdateOne(ctx, bson.M{"_id": c.ID}, bson.M{"$set": bson.M{"nextRunAt": time.Time{}, "repeat": ""}})
			continue
		}
		awaiting := []string{}
		for _, aid := range cardAssignees(&c) {
			if m, err := s.GetMember(ctx, aid, c.UserID); err == nil && m.Status == "active" {
				awaiting = append(awaiting, aid)
			}
		}
		if len(awaiting) == 0 {
			// No one to run it — push a day so we don't spin every tick.
			_, _ = s.cards.UpdateOne(ctx, bson.M{"_id": c.ID}, bson.M{"$set": bson.M{"nextRunAt": now.Add(24 * time.Hour)}})
			continue
		}
		_, _ = s.cards.UpdateOne(ctx,
			bson.M{"_id": c.ID, "status": models.CardDone},
			bson.M{
				"$set": bson.M{
					"status": models.CardQueued, "awaiting": awaiting, "running": []string{},
					"nextRunAt": time.Time{}, "error": "", "updatedAt": now,
				},
				"$unset": bson.M{"combineLock": ""},
			})
		log.Printf("🔁 [CREW] recurring card %q re-queued (%s)", c.Title, c.Repeat)
	}
}

// ─── Cards & the pipeline ────────────────────────────────────────────────────

func (s *CrewService) CreateCard(ctx context.Context, userID, projectID, title, detail, objectiveID, repeat string, assigneeIDs, dependsOn []string) (*models.CrewCard, error) {
	if repeat != "daily" && repeat != "weekly" && repeat != "monthly" {
		repeat = ""
	}
	// Dependencies must be real cards on the SAME project. Cycles are
	// impossible by construction: a new card can only depend on cards that
	// already exist. Cap keeps prompts and claim checks sane.
	var deps []string
	for _, depID := range dependsOn {
		if len(deps) >= 5 {
			break
		}
		dep, err := s.GetCard(ctx, depID, userID)
		if err != nil || dep.ProjectID != projectID {
			continue
		}
		deps = append(deps, depID)
	}
	project, err := s.GetProject(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if objectiveID != "" {
		found := false
		for _, o := range project.Objectives {
			if o.ID == objectiveID {
				found = true
				break
			}
		}
		if !found {
			objectiveID = "" // stale objective — drop the link rather than fail
		}
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("card title is required")
	}
	for _, aid := range assigneeIDs {
		if _, err := s.GetMember(ctx, aid, userID); err != nil {
			return nil, fmt.Errorf("assignee not found")
		}
	}
	now := time.Now()
	c := &models.CrewCard{
		ProjectID: projectID, UserID: userID, Title: title, Detail: strings.TrimSpace(detail),
		AssigneeIDs: assigneeIDs, ObjectiveID: objectiveID, Repeat: repeat, DependsOn: deps, Status: models.CardDraft, CreatedAt: now, UpdatedAt: now,
	}
	res, err := s.cards.InsertOne(ctx, c)
	if err != nil {
		return nil, err
	}
	c.ID = res.InsertedID.(primitive.ObjectID)
	return c, nil
}

func (s *CrewService) ListCards(ctx context.Context, userID, projectID string) ([]models.CrewCard, error) {
	cur, err := s.cards.Find(ctx, bson.M{"projectId": projectID, "userId": userID},
		options.Find().SetSort(bson.M{"updatedAt": -1}))
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	out := []models.CrewCard{}
	return out, cur.All(ctx, &out)
}

func (s *CrewService) GetCard(ctx context.Context, id, userID string) (*models.CrewCard, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("card not found")
	}
	var c models.CrewCard
	if err := s.cards.FindOne(ctx, bson.M{"_id": oid, "userId": userID}).Decode(&c); err != nil {
		return nil, fmt.Errorf("card not found")
	}
	return &c, nil
}

// UpdateDraft edits a card while it's still a draft (title/detail/assignee).
func (s *CrewService) UpdateDraft(ctx context.Context, id, userID string, set bson.M) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("card not found")
	}
	set["updatedAt"] = time.Now()
	res, err := s.cards.UpdateOne(ctx,
		bson.M{"_id": oid, "userId": userID, "status": models.CardDraft},
		bson.M{"$set": set})
	if err == nil && res.MatchedCount == 0 {
		return fmt.Errorf("only draft cards can be edited")
	}
	return err
}

func (s *CrewService) DeleteCard(ctx context.Context, id, userID string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("card not found")
	}
	res, err := s.cards.DeleteOne(ctx, bson.M{"_id": oid, "userId": userID,
		"status": bson.M{"$in": []string{models.CardDraft, models.CardDone}}})
	if err == nil && res.DeletedCount == 0 {
		return fmt.Errorf("only draft or done cards can be deleted")
	}
	return err
}

// cardAssignees resolves a card's assignee list (with legacy single-field fallback).
func cardAssignees(c *models.CrewCard) []string {
	if len(c.AssigneeIDs) > 0 {
		return c.AssigneeIDs
	}
	if c.AssigneeID != "" {
		return []string{c.AssigneeID}
	}
	return nil
}

// PromoteCard: draft → queued (user action). Every ACTIVE assignee gets a
// pass this round — multiple members can work the same card in parallel.
func (s *CrewService) PromoteCard(ctx context.Context, id, userID string) error {
	c, err := s.GetCard(ctx, id, userID)
	if err != nil {
		return err
	}
	if c.Status != models.CardDraft {
		return fmt.Errorf("only draft cards can be queued")
	}
	ids := cardAssignees(c)
	if len(ids) == 0 {
		return fmt.Errorf("assign at least one team member before queueing")
	}
	awaiting := []string{}
	for _, aid := range ids {
		m, err := s.GetMember(ctx, aid, userID)
		if err != nil {
			continue
		}
		if m.Status == "active" {
			awaiting = append(awaiting, aid)
		}
	}
	if len(awaiting) == 0 {
		return fmt.Errorf("all assignees are paused or missing")
	}
	oid, _ := primitive.ObjectIDFromHex(id)
	_, err = s.cards.UpdateOne(ctx, bson.M{"_id": oid},
		bson.M{"$set": bson.M{
			"status": models.CardQueued, "error": "", "awaiting": awaiting,
			"running": []string{}, "assigneeIds": ids, "updatedAt": time.Now(),
		}})
	return err
}

// ClaimNextCard atomically claims ONE pending pass for a member: pulls the
// member from the card's awaiting set and marks it running. The atomic
// findOneAndUpdate IS the lock — the same member can never double-claim, and
// different members claim their own passes on the same card independently.
func (s *CrewService) ClaimNextCard(ctx context.Context, memberID string) (*models.CrewCard, error) {
	// Oldest-first candidates; claim the first whose dependencies are all
	// done. Dependency-blocked cards simply wait — they surface as queued.
	cur, err := s.cards.Find(ctx,
		bson.M{"awaiting": memberID, "status": bson.M{"$in": []string{models.CardQueued, models.CardWorking}}},
		options.Find().SetSort(bson.M{"updatedAt": 1}).SetLimit(10))
	if err != nil {
		return nil, err
	}
	var candidates []models.CrewCard
	if err := cur.All(ctx, &candidates); err != nil {
		return nil, err
	}
	now := time.Now()
	for i := range candidates {
		if !s.depsSatisfied(ctx, &candidates[i]) {
			continue
		}
		// Coordinated cards wait for the team-lead split before anyone runs.
		if len(candidates[i].AssigneeIDs) > 1 && len(candidates[i].Parts) == 0 {
			continue
		}
		var c models.CrewCard
		err := s.cards.FindOneAndUpdate(ctx,
			bson.M{"_id": candidates[i].ID, "awaiting": memberID, "status": bson.M{"$in": []string{models.CardQueued, models.CardWorking}}},
			bson.M{
				"$pull":     bson.M{"awaiting": memberID},
				"$addToSet": bson.M{"running": memberID},
				"$set":      bson.M{"status": models.CardWorking, "claimedAt": now, "updatedAt": now},
			},
			options.FindOneAndUpdate().SetReturnDocument(options.After),
		).Decode(&c)
		if err == mongo.ErrNoDocuments {
			continue // raced away — try the next candidate
		}
		if err != nil {
			return nil, err
		}
		return &c, nil
	}
	return nil, nil
}

// depsSatisfied reports whether every dependency card is DONE. A missing or
// deleted dependency never blocks forever — it's treated as satisfied.
func (s *CrewService) depsSatisfied(ctx context.Context, c *models.CrewCard) bool {
	for _, depID := range c.DependsOn {
		oid, err := primitive.ObjectIDFromHex(depID)
		if err != nil {
			continue
		}
		var dep models.CrewCard
		if err := s.cards.FindOne(ctx, bson.M{"_id": oid}).Decode(&dep); err != nil {
			continue
		}
		if dep.Status != models.CardDone {
			return false
		}
	}
	return true
}

// SubmitOutput records one member's pass. The card moves to review only when
// EVERY assignee's pass has landed (awaiting and running both empty). ALWAYS
// human review — no agent self-approval.
func (s *CrewService) SubmitOutput(ctx context.Context, cardID, memberID, memberName, output string, toolsUsed []string, tokensUsed int64, runErr string) error {
	oid, err := primitive.ObjectIDFromHex(cardID)
	if err != nil {
		return fmt.Errorf("card not found")
	}
	now := time.Now()
	if runErr != "" {
		// Failed run: surface the error; the whole card returns to draft so the
		// human decides (surface, don't self-heal). Clears in-flight state.
		_, err = s.cards.UpdateOne(ctx, bson.M{"_id": oid, "status": models.CardWorking},
			bson.M{"$set": bson.M{
				"status": models.CardDraft, "error": memberName + ": " + runErr,
				"awaiting": []string{}, "running": []string{}, "updatedAt": now,
			}})
		return err
	}
	if len(toolsUsed) > 40 {
		toolsUsed = toolsUsed[:40]
	}
	rev := models.CardRevision{MemberID: memberID, MemberName: memberName, Output: output, ToolsUsed: toolsUsed, TokensUsed: tokensUsed, At: now}
	if _, err = s.cards.UpdateOne(ctx, bson.M{"_id": oid, "status": models.CardWorking},
		bson.M{
			"$pull": bson.M{"running": memberID},
			"$set":  bson.M{"latestOutput": output, "error": "", "updatedAt": now},
			"$push": bson.M{"revisions": rev},
		}); err != nil {
		return err
	}
	// All passes in? → review — EXCEPT coordinated (split) cards, which stay
	// working until the combine pass merges the parts into one deliverable.
	_, err = s.cards.UpdateOne(ctx,
		bson.M{"_id": oid, "status": models.CardWorking, "awaiting": bson.M{"$size": 0}, "running": bson.M{"$size": 0},
			"parts": bson.M{"$exists": false}},
		bson.M{"$set": bson.M{"status": models.CardReview, "updatedAt": time.Now()}})
	return err
}

// ReviewCard is the human gate: approve → done; changes → the feedback is
// recorded on the latest revision and the card AUTOMATICALLY re-queues.
func (s *CrewService) ReviewCard(ctx context.Context, id, userID string, approve bool, feedback string) error {
	c, err := s.GetCard(ctx, id, userID)
	if err != nil {
		return err
	}
	if c.Status != models.CardReview {
		return fmt.Errorf("card is not waiting for review")
	}
	if len(c.Revisions) == 0 {
		return fmt.Errorf("card has no output to review")
	}
	oid, _ := primitive.ObjectIDFromHex(id)
	last := len(c.Revisions) - 1
	now := time.Now()
	set := bson.M{
		fmt.Sprintf("revisions.%d.review", last):   strings.TrimSpace(feedback),
		fmt.Sprintf("revisions.%d.approved", last): approve,
		"updatedAt": now,
	}
	if approve {
		set["status"] = models.CardDone
		// Recurring card: schedule the next automatic run.
		if next := nextRecurrence(c.Repeat, now); !next.IsZero() {
			set["nextRunAt"] = next
		}
	} else {
		if strings.TrimSpace(feedback) == "" {
			return fmt.Errorf("add feedback so the agent knows what to change")
		}
		// Auto re-queue: every ACTIVE assignee gets another pass with the
		// review attached.
		awaiting := []string{}
		for _, aid := range cardAssignees(c) {
			if m, err := s.GetMember(ctx, aid, userID); err == nil && m.Status == "active" {
				awaiting = append(awaiting, aid)
			}
		}
		if len(awaiting) == 0 {
			return fmt.Errorf("no active assignees to re-queue to — resume or add a member first")
		}
		set["status"] = models.CardQueued
		set["awaiting"] = awaiting
		set["running"] = []string{}
	}
	_, err = s.cards.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": set, "$unset": bson.M{"combineLock": ""}})
	return err
}

// MembersWithQueuedCards returns ACTIVE members that have at least one queued
// card — the worker's scan set. One aggregation, cheap on the status index.
func (s *CrewService) MembersWithQueuedCards(ctx context.Context) ([]models.CrewMember, error) {
	ids, err := s.cards.Distinct(ctx, "awaiting", bson.M{"status": bson.M{"$in": []string{models.CardQueued, models.CardWorking}}})
	if err != nil {
		return nil, err
	}
	oids := make([]primitive.ObjectID, 0, len(ids))
	for _, v := range ids {
		if str, ok := v.(string); ok && str != "" {
			if oid, err := primitive.ObjectIDFromHex(str); err == nil {
				oids = append(oids, oid)
			}
		}
	}
	if len(oids) == 0 {
		return nil, nil
	}
	cur, err := s.members.Find(ctx, bson.M{"_id": bson.M{"$in": oids}, "status": "active"})
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	out := []models.CrewMember{}
	return out, cur.All(ctx, &out)
}

// ─── Context building (fixes "no one knows the project") ────────────────────

// ProjectMemory returns the team's APPROVED work so far (org memory): the last
// few done cards with a snippet of their approved output. Context, not
// constraint — the prompt frames it as "what the team already produced".
func (s *CrewService) ProjectMemory(ctx context.Context, userID, projectID string, limit int) (string, error) {
	if limit <= 0 {
		limit = 6
	}
	cur, err := s.cards.Find(ctx,
		bson.M{"projectId": projectID, "userId": userID, "status": models.CardDone},
		options.Find().SetSort(bson.M{"updatedAt": -1}).SetLimit(int64(limit)))
	if err != nil {
		return "", err
	}
	defer func() { _ = cur.Close(ctx) }()
	var cards []models.CrewCard
	if err := cur.All(ctx, &cards); err != nil {
		return "", err
	}
	// Newest first from Mongo; write oldest→newest so the narrative reads in
	// order. Budgeted so a few long reports can't blow up every prompt.
	budget := 9000
	var b strings.Builder
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		out := strings.TrimSpace(c.LatestOutput)
		snip := 1200
		if snip > budget {
			snip = budget
		}
		if len(out) > snip {
			out = out[:snip] + "…"
		}
		fmt.Fprintf(&b, "### %s (approved)\n%s\n\n", c.Title, out)
		budget -= len(out)
		if budget <= 0 {
			break
		}
	}
	return strings.TrimSpace(b.String()), nil
}

// BoardSummary is what every agent run sees about the rest of the board.
func (s *CrewService) BoardSummary(ctx context.Context, userID, projectID string) (string, error) {
	cards, err := s.ListCards(ctx, userID, projectID)
	if err != nil {
		return "", err
	}
	members, err := s.ListMembers(ctx, userID, projectID)
	if err != nil {
		return "", err
	}
	names := map[string]string{}
	for _, m := range members {
		names[m.ID.Hex()] = m.Name
	}
	var b strings.Builder
	for _, c := range cards {
		var assigned []string
		for _, id := range c.AssigneeIDs {
			if n := names[id]; n != "" {
				assigned = append(assigned, n)
			}
		}
		if len(assigned) == 0 {
			if n := names[c.AssigneeID]; n != "" { // legacy single-assignee cards
				assigned = append(assigned, n)
			}
		}
		who := "unassigned"
		if len(assigned) > 0 {
			who = strings.Join(assigned, ", ")
		}
		fmt.Fprintf(&b, "- [%s] %s (%s)\n", c.Status, c.Title, who)
	}
	return b.String(), nil
}
