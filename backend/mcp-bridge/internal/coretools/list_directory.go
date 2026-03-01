package coretools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ListDirTool lists directory contents with optional recursion.
type ListDirTool struct{}

func NewListDirTool() *ListDirTool {
	return &ListDirTool{}
}

func (t *ListDirTool) Name() string { return "list_directory" }

func (t *ListDirTool) Description() string {
	return "List the contents of a directory on the user's local filesystem.\n\n" +
		"Flat mode (default): shows directories first (alphabetically), then files with sizes. " +
		"Gives a quick overview of a single directory level.\n\n" +
		"Recursive mode: shows a tree-like structure with indentation, configurable depth " +
		"(default 3, max 10). Skips common non-source directories (node_modules, .git, etc.). " +
		"Capped at 500 entries to prevent excessive output.\n\n" +
		"Hidden files (starting with '.') are excluded by default; set show_hidden=true to include them.\n\n" +
		"Use this to explore project structure, understand directory layout, " +
		"or get an overview before diving into specific files with read_file."
}

func (t *ListDirTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Directory path to list. Defaults to current working directory.",
			},
			"recursive": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to list contents recursively. Defaults to false.",
			},
			"max_depth": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum directory depth for recursive listing. Defaults to 3.",
			},
			"show_hidden": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to show hidden files/directories (starting with '.'). Defaults to false.",
			},
		},
		"required": []string{},
	}
}

type dirEntry struct {
	path  string
	isDir bool
	size  int64
}

func (t *ListDirTool) Execute(_ context.Context, args map[string]interface{}) (string, error) {
	dirPath := "."
	if p, ok := args["path"].(string); ok && p != "" {
		dirPath = p
	}

	recursive := false
	if r, ok := args["recursive"].(bool); ok {
		recursive = r
	}

	maxDepth := 3
	if md, ok := args["max_depth"].(float64); ok && md > 0 {
		maxDepth = int(md)
		if maxDepth > 10 {
			maxDepth = 10
		}
	}

	showHidden := false
	if sh, ok := args["show_hidden"].(bool); ok {
		showHidden = sh
	}

	// Verify path exists and is a directory
	info, err := os.Stat(dirPath)
	if err != nil {
		return "", fmt.Errorf("cannot access path: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", dirPath)
	}

	absPath, _ := filepath.Abs(dirPath)

	if recursive {
		return t.listRecursive(absPath, maxDepth, showHidden)
	}
	return t.listFlat(absPath, showHidden)
}

func (t *ListDirTool) listFlat(dirPath string, showHidden bool) (string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	var dirs, files []dirEntry

	for _, e := range entries {
		name := e.Name()
		if !showHidden && strings.HasPrefix(name, ".") {
			continue
		}

		fi, err := e.Info()
		if err != nil {
			continue
		}

		entry := dirEntry{
			path:  name,
			isDir: e.IsDir(),
			size:  fi.Size(),
		}

		if e.IsDir() {
			dirs = append(dirs, entry)
		} else {
			files = append(files, entry)
		}
	}

	// Sort alphabetically within each group
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].path < dirs[j].path })
	sort.Slice(files, func(i, j int) bool { return files[i].path < files[j].path })

	var sb strings.Builder
	fmt.Fprintf(&sb, "Directory: %s (%d directories, %d files)\n\n", dirPath, len(dirs), len(files))

	for _, d := range dirs {
		fmt.Fprintf(&sb, "  %s/\n", d.path)
	}
	for _, f := range files {
		fmt.Fprintf(&sb, "  %s (%s)\n", f.path, formatSize(f.size))
	}

	return sb.String(), nil
}

func (t *ListDirTool) listRecursive(dirPath string, maxDepth int, showHidden bool) (string, error) {
	var sb strings.Builder
	totalDirs := 0
	totalFiles := 0
	maxEntries := 500 // cap to prevent huge output

	fmt.Fprintf(&sb, "Directory tree: %s (max depth %d)\n\n", dirPath, maxDepth)

	err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible
		}

		// Calculate depth relative to root
		relPath, _ := filepath.Rel(dirPath, path)
		if relPath == "." {
			return nil
		}

		depth := strings.Count(relPath, string(filepath.Separator))
		if depth > maxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		name := d.Name()

		// Skip hidden files/dirs
		if !showHidden && strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip ignored directories
		if d.IsDir() && ShouldIgnoreDir(name) {
			return filepath.SkipDir
		}

		totalDirs++
		totalFiles++
		if totalDirs+totalFiles > maxEntries {
			return filepath.SkipAll
		}

		// Format with tree-like indentation
		indent := strings.Repeat("  ", depth)
		if d.IsDir() {
			fmt.Fprintf(&sb, "%s%s/\n", indent, name)
		} else {
			fi, err := d.Info()
			if err != nil {
				fmt.Fprintf(&sb, "%s%s\n", indent, name)
			} else {
				fmt.Fprintf(&sb, "%s%s (%s)\n", indent, name, formatSize(fi.Size()))
			}
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to walk directory: %w", err)
	}

	if totalDirs+totalFiles >= maxEntries {
		sb.WriteString("\n... (output capped at 500 entries)")
	}

	return sb.String(), nil
}
