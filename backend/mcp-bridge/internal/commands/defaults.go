package commands

// DefaultMCPServer represents a pre-configured MCP server that users can enable
type DefaultMCPServer struct {
	Name        string   // Unique identifier
	Description string   // Human-readable description
	Command     string   // Command to run (e.g., "npx")
	Args        []string // Command arguments
	Category    string   // Category for grouping (e.g., "Essential", "Developer Tools")
}

// DefaultMCPServers contains all recommended MCP servers
var DefaultMCPServers = []DefaultMCPServer{
	// Essential
	{
		Name:        "filesystem",
		Description: "File operations (read, write, search)",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-filesystem", "."},
		Category:    "Essential",
	},
	{
		Name:        "git",
		Description: "Git operations (status, diff, commit)",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-git"},
		Category:    "Essential",
	},
	{
		Name:        "memory",
		Description: "Persistent memory across sessions",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-memory"},
		Category:    "Essential",
	},
	// Browser
	{
		Name:        "chrome-devtools",
		Description: "Chrome browser automation & DevTools",
		Command:     "npx",
		Args:        []string{"-y", "chrome-devtools-mcp@latest"},
		Category:    "Browser",
	},
	{
		Name:        "puppeteer",
		Description: "Headless Chrome automation",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-puppeteer"},
		Category:    "Browser",
	},
	{
		Name:        "playwright",
		Description: "Cross-browser automation",
		Command:     "npx",
		Args:        []string{"-y", "@anthropic/mcp-server-playwright"},
		Category:    "Browser",
	},
	// Data & Search
	{
		Name:        "fetch",
		Description: "Fetch URLs and convert to markdown",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-fetch"},
		Category:    "Data",
	},
	{
		Name:        "brave-search",
		Description: "Web search via Brave Search API",
		Command:     "npx",
		Args:        []string{"-y", "@anthropic/mcp-server-brave-search"},
		Category:    "Data",
	},
	{
		Name:        "sqlite",
		Description: "SQLite database operations",
		Command:     "npx",
		Args:        []string{"-y", "@anthropic/mcp-server-sqlite"},
		Category:    "Data",
	},
	// Development
	{
		Name:        "github",
		Description: "GitHub API operations",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-github"},
		Category:    "Development",
	},
	{
		Name:        "gitlab",
		Description: "GitLab API operations",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-gitlab"},
		Category:    "Development",
	},
	{
		Name:        "slack",
		Description: "Slack messaging integration",
		Command:     "npx",
		Args:        []string{"-y", "@anthropic/mcp-server-slack"},
		Category:    "Communication",
	},
}

// GetDefaultServerByName returns a default server by name
func GetDefaultServerByName(name string) *DefaultMCPServer {
	for _, server := range DefaultMCPServers {
		if server.Name == name {
			return &server
		}
	}
	return nil
}

// GetDefaultServersByCategory returns all servers in a category
func GetDefaultServersByCategory(category string) []DefaultMCPServer {
	var servers []DefaultMCPServer
	for _, server := range DefaultMCPServers {
		if server.Category == category {
			servers = append(servers, server)
		}
	}
	return servers
}

// GetCategories returns all unique categories
func GetCategories() []string {
	categoryMap := make(map[string]bool)
	for _, server := range DefaultMCPServers {
		categoryMap[server.Category] = true
	}

	var categories []string
	for category := range categoryMap {
		categories = append(categories, category)
	}
	return categories
}
