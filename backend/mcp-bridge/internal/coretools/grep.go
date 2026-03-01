package coretools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GrepTool searches file contents using grep with regex support.
// Uses exec.Command with array args (no shell injection risk).
type GrepTool struct{}

func NewGrepTool() *GrepTool {
	return &GrepTool{}
}

func (t *GrepTool) Name() string { return "grep" }

func (t *GrepTool) Description() string {
	return "Search the contents of files for a regular expression pattern (POSIX extended regex).\n\n" +
		"Searches recursively from the given directory, returning matching lines in " +
		"file:line:content format. Automatically skips non-source directories " +
		"(node_modules, .git, dist, vendor, etc.).\n\n" +
		"Options:\n" +
		"- case_sensitive: defaults to true; set false for case-insensitive search.\n" +
		"- include: file glob filter (e.g., '*.ts') to limit which files are searched.\n" +
		"- max_results: cap on matching lines (default 30, max 100).\n\n" +
		"Long matching lines are truncated to 250 characters to save tokens. " +
		"Uses safe argument passing (no shell injection risk).\n\n" +
		"Use this to find function definitions, variable usage, imports, error messages, " +
		"configuration values, TODO comments, or any text pattern across the codebase."
}

func (t *GrepTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Regular expression pattern to search for (POSIX extended regex).",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Directory or file to search in. Defaults to current working directory.",
			},
			"max_results": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of matching lines to return. Defaults to 30, max 100.",
			},
			"case_sensitive": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the search should be case-sensitive. Defaults to true.",
			},
			"include": map[string]interface{}{
				"type":        "string",
				"description": "File glob pattern to include (e.g., '*.go', '*.ts'). Only searches matching files.",
			},
		},
		"required": []string{"query"},
	}
}

func (t *GrepTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("query is required")
	}

	searchPath := "."
	if p, ok := args["path"].(string); ok && p != "" {
		searchPath = p
	}

	maxResults := 30
	if mr, ok := args["max_results"].(float64); ok && mr > 0 {
		maxResults = int(mr)
		if maxResults > 100 {
			maxResults = 100
		}
	}

	caseSensitive := true
	if cs, ok := args["case_sensitive"].(bool); ok {
		caseSensitive = cs
	}

	// Build grep arguments (safe array, no shell injection)
	grepArgs := []string{"-r", "-n", "-E"}

	if !caseSensitive {
		grepArgs = append(grepArgs, "-i")
	}

	// Add exclude-dir flags
	for _, dir := range DefaultIgnoreDirs {
		grepArgs = append(grepArgs, fmt.Sprintf("--exclude-dir=%s", dir))
	}

	// Add include filter
	if include, ok := args["include"].(string); ok && include != "" {
		grepArgs = append(grepArgs, fmt.Sprintf("--include=%s", include))
	}

	// Limit output (extra buffer for post-processing)
	grepArgs = append(grepArgs, fmt.Sprintf("--max-count=%d", maxResults))

	grepArgs = append(grepArgs, query, searchPath)

	cmd := exec.CommandContext(ctx, "grep", grepArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// grep returns exit code 1 when no matches found — not an error
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return fmt.Sprintf("No matches found for '%s' in %s", query, searchPath), nil
		}
		if stderr.Len() > 0 {
			return "", fmt.Errorf("grep error: %s", strings.TrimSpace(stderr.String()))
		}
		return "", fmt.Errorf("grep failed: %w", err)
	}

	output := stdout.String()
	if output == "" {
		return fmt.Sprintf("No matches found for '%s' in %s", query, searchPath), nil
	}

	// Process output: truncate long matching lines
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var processed []string
	for i, line := range lines {
		if i >= maxResults {
			break
		}
		if len(line) > 300 {
			// Keep path:linenum prefix, truncate the content
			colonIdx := strings.Index(line, ":")
			if colonIdx > 0 {
				secondColon := strings.Index(line[colonIdx+1:], ":")
				if secondColon > 0 {
					prefix := line[:colonIdx+1+secondColon+1]
					content := line[colonIdx+1+secondColon+1:]
					if len(content) > 250 {
						content = content[:250] + "..."
					}
					line = prefix + content
				}
			}
		}
		processed = append(processed, line)
	}

	var sb strings.Builder
	matchCount := len(processed)
	fmt.Fprintf(&sb, "Found %d matches for '%s'", matchCount, query)
	if matchCount >= maxResults {
		sb.WriteString(" (results capped)")
	}
	sb.WriteString(":\n\n")
	sb.WriteString(strings.Join(processed, "\n"))

	return sb.String(), nil
}
