package services

import "strings"

// Tool grouping (Phase 2 of tool selection).
//
// There are ~192 tools (mostly third-party integrations). Asking the small tool
// predictor to pick the right handful out of 192 — every turn — is unreliable
// (it hallucinated, flip-flopped, and split tools that belong together). Instead
// the predictor chooses among a small set of capability GROUPS: a coarse choice
// a small model gets right, where related tools (e.g. the whole data pipeline)
// always travel together.
//
// groupOfTool maps EVERY tool to exactly one group; anything unrecognized falls
// to "integrations", so no tool is ever unreachable. The predictor is shown only
// the groups that actually contain a tool the user has access to.

type toolGroupInfo struct {
	Name string
	Desc string
}

// toolGroupCatalog is the predictor-facing list (name + description). "core" is
// intentionally absent — core tools are always included separately, never a
// choice. Keep descriptions concrete so the predictor routes correctly.
var toolGroupCatalog = []toolGroupInfo{
	{"data_analysis", "Analyze or visualize data: uploaded CSV/Excel files, run Python, compute statistics, build charts/dashboards/funnels, generate reports, PDFs or documents. Use for ANY data file, analysis, chart, or report request."},
	{"web_research", "Search the web, scrape web pages, search images on the web, or search the user's own knowledge base."},
	{"media", "Generate images, edit images, describe/understand an image, or transcribe audio."},
	{"memory", "Remember facts about the user across chats, or recall previously stored memories."},
	{"email", "Read or send email (Gmail, SendGrid, Brevo, Mailchimp, Netcore)."},
	{"messaging", "Send a message to Slack, Discord, Telegram, Teams, Google Chat, SMS/WhatsApp, or a webhook."},
	{"calendar", "Create, list, update calendar events or find free slots (Google Calendar, Calendly)."},
	{"files_sheets", "Google Drive files and folders, Google Sheets, or Airtable."},
	{"docs_notes", "Notion pages and databases."},
	{"project_mgmt", "Issues, tasks, repos and deploys (GitHub, GitLab, Jira, Linear, ClickUp, Trello, Netlify)."},
	{"crm_sales", "CRM and commerce data (HubSpot, LeadSquared, Shopify)."},
	{"analytics_events", "Marketing & product analytics: Google Analytics (GA4) traffic/users, Google Ads & Facebook Ads performance (spend/clicks/conversions), Google Search Console (SEO/organic search), Mixpanel, PostHog."},
	{"social_media", "Post or read on X/Twitter, LinkedIn, YouTube, or WhatsApp/LinkedIn via Unipile."},
	{"meetings", "Zoom meetings: create, list, update, manage registrants."},
	{"design", "Canva designs and brand assets."},
	{"databases", "Query or write MongoDB, Redis, S3, or call an arbitrary REST API."},
	{"blockchain", "0G chain data: balances, transactions, contracts, NFTs, network stats."},
	{"integrations", "Other third-party integrations not covered by the groups above."},
}

// groupOfTool returns the single group a tool belongs to. Unknown tools fall to
// "integrations" so nothing is ever unreachable.
func groupOfTool(name string) string {
	switch name {
	case "ask_user", "get_current_time", "calculate_math", "spawn_subagent":
		return "core"
	case "run_python", "analyze_data", "read_spreadsheet", "read_data_file", "read_document",
		"download_file", "train_model", "create_document", "html_to_pdf", "create_text_file", "create_presentation":
		return "data_analysis"
	case "search_web", "scrape_web", "search_images", "search_knowledge":
		return "web_research"
	case "generate_image", "edit_image", "describe_image", "transcribe_audio":
		return "media"
	case "add_memory", "search_memory":
		return "memory"
	case "produce_artifact", "read_artifact", "list_artifacts":
		return "orchestration"
	}

	hp := func(prefixes ...string) bool {
		for _, p := range prefixes {
			if strings.HasPrefix(name, p) {
				return true
			}
		}
		return false
	}

	switch {
	case hp("gmail_", "mailchimp_") || name == "send_email" || name == "send_brevo_email" || name == "netcore_send_email":
		return "email"
	case hp("twilio_") || name == "send_slack_message" || name == "send_discord_message" ||
		name == "send_telegram_message" || name == "send_teams_message" ||
		name == "send_google_chat_message" || name == "send_webhook" || name == "referralmonk_whatsapp":
		return "messaging"
	case hp("googlecalendar_", "calendly_"):
		return "calendar"
	case hp("googledrive_", "googlesheets_", "airtable_"):
		return "files_sheets"
	case hp("notion_"):
		return "docs_notes"
	case hp("github_", "gitlab_", "netlify_", "jira_", "linear_", "clickup_", "trello_"):
		return "project_mgmt"
	case hp("hubspot_", "leadsquared_", "shopify_") || name == "netcore_add_contact" || name == "netcore_track_event":
		return "crm_sales"
	case hp("mixpanel_", "posthog_", "facebook_ads_", "google_analytics", "google_ads", "google_search_console"):
		return "analytics_events"
	case hp("twitter_", "x_", "linkedin_", "youtube_", "unipile_"):
		return "social_media"
	case hp("zoom_"):
		return "meetings"
	case hp("canva_"):
		return "design"
	case hp("mongodb_", "redis_", "s3_") || name == "api_request" || name == "test_api":
		return "databases"
	case hp("get_0g_"):
		return "blockchain"
	}
	return "integrations"
}

// availableGroupDefs returns synthetic OpenAI-format "tool" definitions — one per
// group that has at least one tool in the user's catalog. These are passed to the
// existing PredictTools, which then returns the chosen GROUP names (reusing all
// its retry/health/cache machinery unchanged).
func availableGroupDefs(catalog []map[string]interface{}) []map[string]interface{} {
	present := map[string]bool{}
	for _, t := range catalog {
		present[groupOfTool(toolDefName(t))] = true
	}
	defs := make([]map[string]interface{}, 0, len(toolGroupCatalog))
	for _, g := range toolGroupCatalog {
		if !present[g.Name] {
			continue
		}
		defs = append(defs, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        g.Name,
				"description": g.Desc,
			},
		})
	}
	return defs
}

// expandGroupsToTools returns every catalog tool belonging to one of the selected
// groups. selectedGroupDefs are the synthetic group defs PredictTools returned
// (their names are group names).
func expandGroupsToTools(selectedGroupDefs, catalog []map[string]interface{}) []map[string]interface{} {
	want := map[string]bool{}
	for _, g := range selectedGroupDefs {
		if n := toolDefName(g); n != "" {
			want[n] = true
		}
	}
	out := make([]map[string]interface{}, 0, len(catalog))
	for _, t := range catalog {
		if want[groupOfTool(toolDefName(t))] {
			out = append(out, t)
		}
	}
	return out
}

// toolDefName extracts the function name from an OpenAI-format tool definition.
func toolDefName(t map[string]interface{}) string {
	if fn, ok := t["function"].(map[string]interface{}); ok {
		if n, ok := fn["name"].(string); ok {
			return n
		}
	}
	return ""
}

// unionToolsByName returns base plus any catalog tools whose name is in wanted
// and not already present. Additive — it never removes a selected tool, so the
// predictor's picks are preserved and the group closure is guaranteed on top.
func unionToolsByName(base, catalog []map[string]interface{}, wanted map[string]bool) []map[string]interface{} {
	have := make(map[string]bool, len(base))
	for _, t := range base {
		have[toolDefName(t)] = true
	}
	out := base
	for _, t := range catalog {
		n := toolDefName(t)
		if n != "" && wanted[n] && !have[n] {
			out = append(out, t)
			have[n] = true
		}
	}
	return out
}
