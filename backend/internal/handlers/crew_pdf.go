package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gofiber/fiber/v2"
	"github.com/yuin/goldmark"
	gmext "github.com/yuin/goldmark/extension"
)

// Shared headless-Chrome allocator for PDF rendering (started once, lazily).
var (
	crewAllocOnce sync.Once
	crewAllocCtx  context.Context
)

func browserAllocator() context.Context {
	crewAllocOnce.Do(func() {
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
			chromedp.Flag("disable-web-security", true),
			chromedp.Flag("hide-scrollbars", true),
		)
		if p := os.Getenv("CHROME_BIN"); p != "" {
			opts = append(opts, chromedp.ExecPath(p))
		}
		crewAllocCtx, _ = chromedp.NewExecAllocator(context.Background(), opts...)
	})
	return crewAllocCtx
}

// CardPDF — GET /api/crew/cards/:cardId/pdf?rev=N
// Renders a card's output (markdown) into a clean, print-styled PDF report via
// goldmark + the shared headless Chrome. rev is 1-based; default = latest.
func (h *CrewHandler) CardPDF(c *fiber.Ctx) error {
	uid, ok := crewUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "auth required"})
	}
	card, err := h.svc.GetCard(c.Context(), c.Params("cardId"), uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	if len(card.Revisions) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "no output to export yet"})
	}
	idx := len(card.Revisions) - 1
	if q := c.QueryInt("rev", 0); q >= 1 && q <= len(card.Revisions) {
		idx = q - 1
	}
	rev := card.Revisions[idx]

	project, _ := h.svc.GetProject(c.Context(), card.ProjectID, uid)
	projectName := ""
	if project != nil {
		projectName = project.Name
	}

	// Markdown → HTML (GFM tables/strikethrough etc.).
	md := goldmark.New(goldmark.WithExtensions(gmext.GFM))
	var body bytes.Buffer
	if err := md.Convert([]byte(rev.Output), &body); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not render markdown"})
	}

	author := rev.MemberName
	if author == "" {
		author = "Crew agent"
	}
	html := fmt.Sprintf(`<!doctype html><html><head><meta charset="utf-8"><style>
  @page { margin: 18mm 16mm; }
  * { print-color-adjust: exact; -webkit-print-color-adjust: exact; }
  body { font: 11pt/1.55 -apple-system, 'Segoe UI', Helvetica, Arial, sans-serif; color: #1a1c16; margin: 0; }
  .head { border-bottom: 2px solid #4a7c2a; padding-bottom: 10px; margin-bottom: 22px; }
  .eyebrow { font-size: 8pt; letter-spacing: .18em; text-transform: uppercase; color: #4a7c2a; font-weight: 600; }
  h1.title { font-size: 19pt; margin: 6px 0 2px; letter-spacing: -0.01em; }
  .meta { font-size: 8.5pt; color: #777; }
  h1 { font-size: 15pt; margin: 20px 0 6px; } h2 { font-size: 13pt; margin: 18px 0 5px; } h3 { font-size: 11.5pt; margin: 14px 0 4px; }
  p { margin: 6px 0; } ul, ol { margin: 6px 0; padding-left: 22px; } li { margin: 3px 0; }
  code { font: 9pt ui-monospace, Menlo, monospace; background: #f2f3ee; padding: 1px 4px; border-radius: 3px; }
  pre { background: #f2f3ee; border-radius: 6px; padding: 10px 12px; overflow-x: auto; }
  pre code { background: none; padding: 0; }
  table { border-collapse: collapse; width: 100%%; font-size: 9.5pt; margin: 10px 0; }
  th, td { border: 1px solid #ddd; padding: 5px 8px; text-align: left; }
  th { background: #f2f3ee; }
  blockquote { border-left: 3px solid #4a7c2a; margin: 8px 0; padding: 2px 12px; color: #555; }
  a { color: #3a6621; }
  img { max-width: 100%%; }
</style></head><body>
  <div class="head">
    <div class="eyebrow">ClaraVerse · Crew Report</div>
    <h1 class="title">%s</h1>
    <div class="meta">%s · by %s · %s</div>
  </div>
  %s
</body></html>`,
		escapeHTML(card.Title), escapeHTML(projectName), escapeHTML(author),
		rev.At.Format("Jan 2, 2006 15:04"), body.String())

	pdf, err := printHTMLToPDF(html)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "PDF rendering failed"})
	}
	fname := safeFileName(card.Title) + ".pdf"
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", `attachment; filename="`+fname+`"`)
	return c.Send(pdf)
}

// printHTMLToPDF serves the HTML locally and prints it with the shared Chrome.
func printHTMLToPDF(html string) ([]byte, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	taskCtx, tc := chromedp.NewContext(browserAllocator())
	defer tc()
	taskCtx, cancel := context.WithTimeout(taskCtx, 30*time.Second)
	defer cancel()

	var pdf []byte
	err := chromedp.Run(taskCtx,
		chromedp.Navigate(srv.URL),
		chromedp.WaitReady("body"),
		chromedp.Sleep(400*time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var e error
			pdf, _, e = page.PrintToPDF().WithPrintBackground(true).Do(ctx)
			return e
		}),
	)
	return pdf, err
}

var htmlEscaper = map[rune]string{'<': "&lt;", '>': "&gt;", '&': "&amp;", '"': "&quot;"}

func escapeHTML(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if esc, ok := htmlEscaper[r]; ok {
			out = append(out, []rune(esc)...)
		} else {
			out = append(out, r)
		}
	}
	return string(out)
}

var fnameRe = regexp.MustCompile(`[^a-zA-Z0-9 _-]+`)

func safeFileName(s string) string {
	s = fnameRe.ReplaceAllString(s, "")
	if len(s) > 60 {
		s = s[:60]
	}
	if s == "" {
		s = "crew-report"
	}
	return s
}
