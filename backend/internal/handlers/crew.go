package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"claraverse/internal/models"
	"claraverse/internal/services"
	"claraverse/internal/services/rag"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// CrewHandler is the REST surface for Nexus v2 (projects → members → cards).
type CrewHandler struct {
	svc     *services.CrewService
	planner CrewPlanner
}

func NewCrewHandler(svc *services.CrewService, planner CrewPlanner) *CrewHandler {
	return &CrewHandler{svc: svc, planner: planner}
}

func crewUserID(c *fiber.Ctx) (string, bool) {
	id, _ := c.Locals("user_id").(string)
	return id, id != ""
}

func crewErr(c *fiber.Ctx, err error) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
}

// Roles — GET /api/crew/roles
func (h *CrewHandler) Roles(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"roles": h.svc.Roles()})
}

// ─── Projects ────────────────────────────────────────────────────────────────

func (h *CrewHandler) CreateProject(c *fiber.Ctx) error {
	uid, ok := crewUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "auth required"})
	}
	var req struct{ Name, Brief string }
	if err := c.BodyParser(&req); err != nil {
		return crewErr(c, err)
	}
	p, err := h.svc.CreateProject(c.Context(), uid, req.Name, req.Brief)
	if err != nil {
		return crewErr(c, err)
	}
	return c.JSON(p)
}

func (h *CrewHandler) ListProjects(c *fiber.Ctx) error {
	uid, _ := crewUserID(c)
	out, err := h.svc.ListProjects(c.Context(), uid)
	if err != nil {
		log.Printf("❌ [CREW] list projects: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load projects"})
	}
	return c.JSON(fiber.Map{"projects": out})
}

// GetProject returns the project plus its members and cards (the board).
func (h *CrewHandler) GetProject(c *fiber.Ctx) error {
	uid, _ := crewUserID(c)
	id := c.Params("id")
	p, err := h.svc.GetProject(c.Context(), id, uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	members, _ := h.svc.ListMembers(c.Context(), uid, id)
	cards, _ := h.svc.ListCards(c.Context(), uid, id)
	return c.JSON(fiber.Map{"project": p, "members": members, "cards": cards})
}

func (h *CrewHandler) UpdateProject(c *fiber.Ctx) error {
	uid, _ := crewUserID(c)
	var req struct {
		Name   string `json:"name"`
		Brief  string `json:"brief"`
		Status string `json:"status"`
		// Pointers: present-but-empty clears, absent leaves unchanged.
		Goal       *string                 `json:"goal"`
		Objectives *[]models.CrewObjective `json:"objectives"`
	}
	if err := c.BodyParser(&req); err != nil {
		return crewErr(c, err)
	}
	set := bson.M{}
	if req.Name != "" {
		set["name"] = req.Name
	}
	if req.Brief != "" {
		set["brief"] = req.Brief
	}
	if req.Status == "active" || req.Status == "archived" {
		set["status"] = req.Status
	}
	if req.Goal != nil {
		set["goal"] = strings.TrimSpace(*req.Goal)
	}
	if req.Objectives != nil {
		objs := make([]models.CrewObjective, 0, len(*req.Objectives))
		for i, o := range *req.Objectives {
			o.Title = strings.TrimSpace(o.Title)
			if o.Title == "" {
				continue
			}
			if o.ID == "" {
				o.ID = fmt.Sprintf("obj-%d-%d", time.Now().UnixNano(), i)
			}
			objs = append(objs, o)
		}
		if len(objs) > 24 {
			objs = objs[:24]
		}
		set["objectives"] = objs
	}
	if err := h.svc.UpdateProject(c.Context(), c.Params("id"), uid, set); err != nil {
		return crewErr(c, err)
	}
	return c.JSON(fiber.Map{"success": true})
}

// ─── Members ─────────────────────────────────────────────────────────────────

func (h *CrewHandler) HireMember(c *fiber.Ctx) error {
	uid, _ := crewUserID(c)
	var req struct {
		RoleKey string   `json:"role_key"`
		Name    string   `json:"name"`
		Tools   []string `json:"tools"`
		Model   string   `json:"model"`
	}
	if err := c.BodyParser(&req); err != nil {
		return crewErr(c, err)
	}
	m, err := h.svc.HireMember(c.Context(), uid, c.Params("id"), req.RoleKey, req.Name, req.Tools, req.Model)
	if err != nil {
		return crewErr(c, err)
	}
	return c.JSON(m)
}

// CrewPlanner is the seam to the worker's PlanProject (goal → draft cards).
type CrewPlanner interface {
	PlanProject(ctx context.Context, userID, projectID string) ([]models.CrewCard, error)
	PlanCard(ctx context.Context, userID, cardID string) ([]models.CrewCard, error)
}

func (h *CrewHandler) Templates(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"templates": h.svc.ListTemplates()})
}

func (h *CrewHandler) CreateProjectFromTemplate(c *fiber.Ctx) error {
	uid, _ := crewUserID(c)
	var req struct {
		Template string `json:"template"`
		Name     string `json:"name"`
		Brief    string `json:"brief"`
	}
	if err := c.BodyParser(&req); err != nil {
		return crewErr(c, err)
	}
	p, err := h.svc.CreateProjectFromTemplate(c.Context(), uid, req.Template, strings.TrimSpace(req.Name), strings.TrimSpace(req.Brief))
	if err != nil {
		return crewErr(c, err)
	}
	return c.JSON(p)
}

func (h *CrewHandler) PlanProject(c *fiber.Ctx) error {
	uid, _ := crewUserID(c)
	if h.planner == nil {
		return crewErr(c, fmt.Errorf("planner is not available"))
	}
	cards, err := h.planner.PlanProject(c.Context(), uid, c.Params("id"))
	if err != nil {
		return crewErr(c, err)
	}
	return c.JSON(fiber.Map{"cards": cards})
}

func (h *CrewHandler) PlanCard(c *fiber.Ctx) error {
	uid, _ := crewUserID(c)
	if h.planner == nil {
		return crewErr(c, fmt.Errorf("planner is not available"))
	}
	cards, err := h.planner.PlanCard(c.Context(), uid, c.Params("cardId"))
	if err != nil {
		return crewErr(c, err)
	}
	return c.JSON(fiber.Map{"cards": cards})
}

func (h *CrewHandler) SetMemberBudget(c *fiber.Ctx) error {
	uid, _ := crewUserID(c)
	var req struct {
		MonthlyBudget int64 `json:"monthly_budget"`
	}
	if err := c.BodyParser(&req); err != nil {
		return crewErr(c, err)
	}
	if err := h.svc.SetMemberBudget(c.Context(), c.Params("memberId"), uid, req.MonthlyBudget); err != nil {
		return crewErr(c, err)
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *CrewHandler) SetMemberCharter(c *fiber.Ctx) error {
	uid, _ := crewUserID(c)
	var req struct {
		Charter string `json:"charter"`
	}
	if err := c.BodyParser(&req); err != nil {
		return crewErr(c, err)
	}
	if err := h.svc.SetMemberCharter(c.Context(), c.Params("memberId"), uid, req.Charter); err != nil {
		return crewErr(c, err)
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *CrewHandler) SetMemberStatus(c *fiber.Ctx) error {
	uid, _ := crewUserID(c)
	var req struct{ Status string }
	if err := c.BodyParser(&req); err != nil {
		return crewErr(c, err)
	}
	if err := h.svc.SetMemberStatus(c.Context(), c.Params("memberId"), uid, req.Status); err != nil {
		return crewErr(c, err)
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *CrewHandler) FireMember(c *fiber.Ctx) error {
	uid, _ := crewUserID(c)
	if err := h.svc.FireMember(c.Context(), c.Params("memberId"), uid); err != nil {
		return crewErr(c, err)
	}
	return c.JSON(fiber.Map{"success": true})
}

// Skills — GET /api/crew/skills (builtin + the user's own, no prompts)
func (h *CrewHandler) Skills(c *fiber.Ctx) error {
	uid, _ := crewUserID(c)
	skills, err := h.svc.ListAvailableSkills(c.Context(), uid)
	if err != nil {
		log.Printf("❌ [CREW] list skills: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load skills"})
	}
	return c.JSON(fiber.Map{"skills": skills})
}

// SetMemberSkills — PUT /api/crew/members/:memberId/skills {skill_ids}
func (h *CrewHandler) SetMemberSkills(c *fiber.Ctx) error {
	uid, _ := crewUserID(c)
	var req struct {
		SkillIDs []string `json:"skill_ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return crewErr(c, err)
	}
	if err := h.svc.SetMemberSkills(c.Context(), c.Params("memberId"), uid, req.SkillIDs); err != nil {
		return crewErr(c, err)
	}
	return c.JSON(fiber.Map{"success": true})
}

// UploadMemberDoc — POST /api/crew/members/:memberId/docs (multipart "file")
// Attaches a reference document to a member: parsed to text server-side and
// injected into that member's future runs ("RAG in hand").
func (h *CrewHandler) UploadMemberDoc(c *fiber.Ctx) error {
	uid, _ := crewUserID(c)
	fh, err := c.FormFile("file")
	if err != nil {
		return crewErr(c, fmt.Errorf("attach a file"))
	}
	if fh.Size > 15*1024*1024 {
		return crewErr(c, fmt.Errorf("file too large (max 15MB)"))
	}
	f, err := fh.Open()
	if err != nil {
		return crewErr(c, err)
	}
	defer func() { _ = f.Close() }()
	data, err := io.ReadAll(f)
	if err != nil {
		return crewErr(c, err)
	}
	segs, err := rag.ParseFile(fh.Filename, fh.Header.Get("Content-Type"), data)
	if err != nil {
		return crewErr(c, fmt.Errorf("could not read that file: %v", err))
	}
	var b strings.Builder
	for _, seg := range segs {
		b.WriteString(seg.Text)
		b.WriteString("\n")
	}
	if err := h.svc.AddMemberDoc(c.Context(), c.Params("memberId"), uid, fh.Filename, b.String()); err != nil {
		return crewErr(c, err)
	}
	return c.JSON(fiber.Map{"success": true})
}

// DeleteMemberDoc — DELETE /api/crew/members/:memberId/docs/:index
func (h *CrewHandler) DeleteMemberDoc(c *fiber.Ctx) error {
	uid, _ := crewUserID(c)
	idx, _ := strconv.Atoi(c.Params("index"))
	if err := h.svc.RemoveMemberDoc(c.Context(), c.Params("memberId"), uid, idx); err != nil {
		return crewErr(c, err)
	}
	return c.JSON(fiber.Map{"success": true})
}

// UnqueueCard — POST /api/crew/cards/:cardId/unqueue (queued → draft, drag-back)
func (h *CrewHandler) UnqueueCard(c *fiber.Ctx) error {
	uid, _ := crewUserID(c)
	if err := h.svc.UnqueueCard(c.Context(), c.Params("cardId"), uid); err != nil {
		return crewErr(c, err)
	}
	return c.JSON(fiber.Map{"success": true})
}

// ─── Cards ───────────────────────────────────────────────────────────────────

func (h *CrewHandler) CreateCard(c *fiber.Ctx) error {
	uid, _ := crewUserID(c)
	var req struct {
		Title       string   `json:"title"`
		Detail      string   `json:"detail"`
		ObjectiveID string   `json:"objective_id"` // which project objective this serves
		Repeat      string   `json:"repeat"`       // "" | daily | weekly | monthly
		AssigneeID  string   `json:"assignee_id"`  // legacy single
		AssigneeIDs []string `json:"assignee_ids"` // preferred
		DependsOn   []string `json:"depends_on"`   // card ids that must be done first
	}
	if err := c.BodyParser(&req); err != nil {
		return crewErr(c, err)
	}
	ids := req.AssigneeIDs
	if len(ids) == 0 && req.AssigneeID != "" {
		ids = []string{req.AssigneeID}
	}
	card, err := h.svc.CreateCard(c.Context(), uid, c.Params("id"), req.Title, req.Detail, req.ObjectiveID, req.Repeat, ids, req.DependsOn)
	if err != nil {
		return crewErr(c, err)
	}
	return c.JSON(card)
}

func (h *CrewHandler) UpdateCard(c *fiber.Ctx) error {
	uid, _ := crewUserID(c)
	var req struct {
		Title       *string   `json:"title"`
		Detail      *string   `json:"detail"`
		AssigneeIDs *[]string `json:"assignee_ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return crewErr(c, err)
	}
	set := bson.M{}
	if req.Title != nil {
		set["title"] = *req.Title
	}
	if req.Detail != nil {
		set["detail"] = *req.Detail
	}
	if req.AssigneeIDs != nil {
		set["assigneeIds"] = *req.AssigneeIDs
		set["assigneeId"] = "" // clear legacy field
	}
	if err := h.svc.UpdateDraft(c.Context(), c.Params("cardId"), uid, set); err != nil {
		return crewErr(c, err)
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *CrewHandler) DeleteCard(c *fiber.Ctx) error {
	uid, _ := crewUserID(c)
	if err := h.svc.DeleteCard(c.Context(), c.Params("cardId"), uid); err != nil {
		return crewErr(c, err)
	}
	return c.JSON(fiber.Map{"success": true})
}

// PromoteCard — POST /api/crew/cards/:cardId/queue (draft → queued)
func (h *CrewHandler) PromoteCard(c *fiber.Ctx) error {
	uid, _ := crewUserID(c)
	if err := h.svc.PromoteCard(c.Context(), c.Params("cardId"), uid); err != nil {
		return crewErr(c, err)
	}
	return c.JSON(fiber.Map{"success": true})
}

// ReviewCard — POST /api/crew/cards/:cardId/review {approve, feedback}
// The human gate: approve → done; feedback → auto re-queue with context.
func (h *CrewHandler) ReviewCard(c *fiber.Ctx) error {
	uid, _ := crewUserID(c)
	var req struct {
		Approve  bool   `json:"approve"`
		Feedback string `json:"feedback"`
	}
	if err := c.BodyParser(&req); err != nil {
		return crewErr(c, err)
	}
	if err := h.svc.ReviewCard(c.Context(), c.Params("cardId"), uid, req.Approve, req.Feedback); err != nil {
		return crewErr(c, err)
	}
	return c.JSON(fiber.Map{"success": true})
}
