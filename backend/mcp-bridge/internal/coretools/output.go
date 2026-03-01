package coretools

import (
	"fmt"
	"path/filepath"
	"strings"
)

// TruncateSmartly keeps the beginning (60%) and end (40%) of output, inserting
// an omission notice in the middle. This preserves file headers/structure at the
// top and the most recent output at the bottom.
func TruncateSmartly(output string, limit int) string {
	if len(output) <= limit {
		return output
	}
	headSize := limit * 60 / 100
	tailSize := limit - headSize - 80 // leave room for notice
	if tailSize < 0 {
		tailSize = 0
	}

	omitted := len(output) - headSize - tailSize
	notice := fmt.Sprintf("\n\n... [%d characters omitted] ...\n\n", omitted)
	return output[:headSize] + notice + output[len(output)-tailSize:]
}

// FormatFileWithLineNumbers formats file content with line numbers.
// If the content exceeds maxLines, it shows head + tail with an omission notice.
func FormatFileWithLineNumbers(content string, maxLines int) string {
	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	if totalLines <= maxLines || maxLines <= 0 {
		return formatLines(lines, 1)
	}

	headCount := maxLines * 60 / 100
	tailCount := maxLines - headCount
	if tailCount < 1 {
		tailCount = 1
	}

	var sb strings.Builder
	sb.WriteString(formatLines(lines[:headCount], 1))
	omitted := totalLines - headCount - tailCount
	sb.WriteString(fmt.Sprintf("\n... [%d lines omitted (lines %d-%d)] ...\n\n",
		omitted, headCount+1, totalLines-tailCount))
	tailStart := totalLines - tailCount
	sb.WriteString(formatLines(lines[tailStart:], tailStart+1))
	return sb.String()
}

// formatLines formats a slice of lines with line numbers starting at startLine.
func formatLines(lines []string, startLine int) string {
	var sb strings.Builder
	for i, line := range lines {
		// Truncate very long lines
		if len(line) > MaxLineTruncation {
			line = line[:MaxLineTruncation] + "..."
		}
		fmt.Fprintf(&sb, "%6d\t%s\n", startLine+i, line)
	}
	return sb.String()
}

// EstimateTokens returns an approximate token count using the ~4 chars/token heuristic.
func EstimateTokens(content string) int {
	if len(content) == 0 {
		return 0
	}
	return (len(content) + 3) / 4
}

// DetectFileType returns a human-readable file type based on extension.
func DetectFileType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	types := map[string]string{
		".go":     "Go",
		".py":     "Python",
		".js":     "JavaScript",
		".ts":     "TypeScript",
		".tsx":    "TypeScript React",
		".jsx":    "JavaScript React",
		".rs":     "Rust",
		".java":   "Java",
		".c":      "C",
		".cpp":    "C++",
		".h":      "C/C++ Header",
		".cs":     "C#",
		".rb":     "Ruby",
		".php":    "PHP",
		".swift":  "Swift",
		".kt":     "Kotlin",
		".scala":  "Scala",
		".html":   "HTML",
		".css":    "CSS",
		".scss":   "SCSS",
		".less":   "Less",
		".json":   "JSON",
		".yaml":   "YAML",
		".yml":    "YAML",
		".toml":   "TOML",
		".xml":    "XML",
		".sql":    "SQL",
		".sh":     "Shell",
		".bash":   "Bash",
		".zsh":    "Zsh",
		".md":     "Markdown",
		".txt":    "Text",
		".csv":    "CSV",
		".log":    "Log",
		".env":    "Environment",
		".ini":    "INI",
		".cfg":    "Config",
		".conf":   "Config",
		".mod":    "Go Module",
		".sum":    "Go Checksum",
		".lock":   "Lock File",
		".dockerfile": "Dockerfile",
		".proto":  "Protocol Buffers",
		".graphql": "GraphQL",
		".vue":    "Vue",
		".svelte": "Svelte",
	}

	if t, ok := types[ext]; ok {
		return t
	}

	// Check filename-based types
	base := strings.ToLower(filepath.Base(path))
	switch {
	case base == "dockerfile":
		return "Dockerfile"
	case base == "makefile":
		return "Makefile"
	case base == ".gitignore":
		return "Gitignore"
	case base == "go.mod":
		return "Go Module"
	case base == "go.sum":
		return "Go Checksum"
	case base == "package.json":
		return "npm Package"
	case base == "cargo.toml":
		return "Rust Cargo"
	}

	if ext != "" {
		return ext[1:] // strip the dot
	}
	return "unknown"
}

// IsMinified returns true if the content appears to be minified
// (any line exceeds the minification threshold).
func IsMinified(content string) bool {
	lines := strings.SplitN(content, "\n", 5) // check first few lines
	for _, line := range lines {
		if len(line) > MinifiedLineThreshold {
			return true
		}
	}
	return false
}

// ShouldIgnoreDir checks if a directory name should be ignored in searches.
func ShouldIgnoreDir(name string) bool {
	for _, ignored := range DefaultIgnoreDirs {
		if name == ignored {
			return true
		}
	}
	return false
}
