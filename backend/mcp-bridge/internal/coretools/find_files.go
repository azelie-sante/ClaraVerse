package coretools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FindFilesTool searches for files matching a glob pattern.
type FindFilesTool struct{}

func NewFindFilesTool() *FindFilesTool {
	return &FindFilesTool{}
}

func (t *FindFilesTool) Name() string { return "find_files" }

func (t *FindFilesTool) Description() string {
	return "Search for files by name using glob patterns (e.g., '*.go', 'Dockerfile*', '*.test.ts').\n\n" +
		"The pattern matches against the filename only (not the full path). " +
		"Searches recursively from the given directory, automatically skipping common " +
		"non-source directories: node_modules, .git, dist, build, __pycache__, vendor, .venv, " +
		"and others.\n\n" +
		"Results include file paths with human-readable sizes, sorted by modification time " +
		"(newest first). Default limit is 50 results, configurable up to 200.\n\n" +
		"Use this to locate files before reading or editing them, to discover project structure, " +
		"or to find files of a specific type across the codebase."
}

func (t *FindFilesTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Glob pattern to match filenames (e.g., '*.go', '*.test.ts', 'Dockerfile*'). Matches against the filename only, not the full path.",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Directory to search in. Defaults to current working directory.",
			},
			"max_results": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of results. Defaults to 50, max 200.",
			},
		},
		"required": []string{"pattern"},
	}
}

type fileResult struct {
	path    string
	size    int64
	modTime int64
}

func (t *FindFilesTool) Execute(_ context.Context, args map[string]interface{}) (string, error) {
	pattern, ok := args["pattern"].(string)
	if !ok || pattern == "" {
		return "", fmt.Errorf("pattern is required")
	}

	searchPath := "."
	if p, ok := args["path"].(string); ok && p != "" {
		searchPath = p
	}

	maxResults := 50
	if mr, ok := args["max_results"].(float64); ok && mr > 0 {
		maxResults = int(mr)
		if maxResults > 200 {
			maxResults = 200
		}
	}

	// Verify search path exists
	info, err := os.Stat(searchPath)
	if err != nil {
		return "", fmt.Errorf("cannot access path: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", searchPath)
	}

	var results []fileResult
	totalScanned := 0

	err = filepath.WalkDir(searchPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible paths
		}

		// Skip ignored directories
		if d.IsDir() {
			if ShouldIgnoreDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		totalScanned++

		// Match against filename
		matched, matchErr := filepath.Match(pattern, d.Name())
		if matchErr != nil {
			return fmt.Errorf("invalid glob pattern: %w", matchErr)
		}

		if matched {
			fi, err := d.Info()
			if err != nil {
				return nil
			}
			results = append(results, fileResult{
				path:    path,
				size:    fi.Size(),
				modTime: fi.ModTime().Unix(),
			})
		}

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		return fmt.Sprintf("No files matching '%s' found in %s (%d files scanned)", pattern, searchPath, totalScanned), nil
	}

	// Sort by modification time (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].modTime > results[j].modTime
	})

	// Truncate to max results
	truncated := false
	totalMatches := len(results)
	if len(results) > maxResults {
		results = results[:maxResults]
		truncated = true
	}

	// Format output
	var sb strings.Builder
	fmt.Fprintf(&sb, "Found %d files matching '%s'", totalMatches, pattern)
	if truncated {
		fmt.Fprintf(&sb, " (showing first %d)", maxResults)
	}
	sb.WriteString(":\n\n")

	for _, r := range results {
		fmt.Fprintf(&sb, "%s (%s)\n", r.path, formatSize(r.size))
	}

	return sb.String(), nil
}

// formatSize returns a human-readable file size.
func formatSize(bytes int64) string {
	switch {
	case bytes >= 1024*1024:
		return fmt.Sprintf("%.1fMB", float64(bytes)/(1024*1024))
	case bytes >= 1024:
		return fmt.Sprintf("%.1fKB", float64(bytes)/1024)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
