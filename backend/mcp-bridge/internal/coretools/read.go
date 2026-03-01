package coretools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

// ReadTool reads files from the local filesystem with progressive disclosure.
// Small files (<= MetadataThresholdLines) return content directly.
// Large files return metadata first with instructions to use start_line/end_line.
type ReadTool struct {
	MaxFileSize int64
}

func NewReadTool() *ReadTool {
	return &ReadTool{
		MaxFileSize: 10 * 1024 * 1024, // 10 MB
	}
}

func (t *ReadTool) Name() string { return "read_file" }

func (t *ReadTool) Description() string {
	return "Read a file from the user's local filesystem and return its content with line numbers.\n\n" +
		"Behavior by file size:\n" +
		"- Small files (up to 300 lines): returns the full content with 6-digit line numbers.\n" +
		"- Large files (over 300 lines): returns metadata (type, size, line count, token estimate) " +
		"plus a 50-line preview. Use start_line and end_line to read specific sections.\n" +
		"- Minified files (lines over 10K chars): warns and suggests alternatives.\n\n" +
		"Set metadata_only=true to get file info (type, size, lines, tokens) without reading content. " +
		"Max file size is 10MB. Lines longer than 2000 characters are truncated.\n\n" +
		"Use this tool to examine source code, configuration files, logs, and any text file. " +
		"For binary files or very large files, use execute_bash with head/tail instead."
}

func (t *ReadTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Absolute path to the file to read.",
			},
			"start_line": map[string]interface{}{
				"type":        "integer",
				"description": "Line number to start reading from (1-based). Defaults to 1.",
			},
			"end_line": map[string]interface{}{
				"type":        "integer",
				"description": "Line number to stop reading at (inclusive). Defaults to start_line + 2000.",
			},
			"metadata_only": map[string]interface{}{
				"type":        "boolean",
				"description": "If true, return only file metadata (line count, size, type) without content.",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ReadTool) Execute(_ context.Context, args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return "", fmt.Errorf("path is required")
	}

	// Check file exists and size
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("cannot access file: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s is a directory, not a file. Use list_directory instead", path)
	}
	if info.Size() > t.MaxFileSize {
		return "", fmt.Errorf("file too large (%d bytes, max %d). Use execute_bash with head/tail to sample it", info.Size(), t.MaxFileSize)
	}

	// Count total lines first (needed for progressive disclosure)
	totalLines, err := countLines(path)
	if err != nil {
		return "", fmt.Errorf("failed to count lines: %w", err)
	}

	fileType := DetectFileType(path)
	estimatedTokens := EstimateTokens(fmt.Sprintf("%d", info.Size())) // rough estimate from file size
	estimatedTokens = int(info.Size()) / 4                           // more accurate

	metadataOnly := false
	if mo, ok := args["metadata_only"].(bool); ok {
		metadataOnly = mo
	}

	// Metadata response
	metadata := fmt.Sprintf("File: %s\nType: %s\nSize: %d bytes\nLines: %d\nEstimated tokens: ~%d",
		path, fileType, info.Size(), totalLines, estimatedTokens)

	if metadataOnly {
		return metadata, nil
	}

	// Check for minified content
	if isFileMinified(path) {
		return metadata + "\n\nWARNING: This file appears to be minified (extremely long lines). " +
			"Reading it would consume excessive tokens. Consider:\n" +
			"- Using a source map or unminified version instead\n" +
			"- Using grep to search for specific content\n" +
			"- Using execute_bash with head -c to sample a portion", nil
	}

	startLine := 1
	if sl, ok := args["start_line"].(float64); ok && sl > 0 {
		startLine = int(sl)
	}

	endLine := startLine + 2000 - 1
	if el, ok := args["end_line"].(float64); ok && el > 0 {
		endLine = int(el)
	}

	// Progressive disclosure: large files get metadata + instructions
	if totalLines > MetadataThresholdLines && startLine == 1 && endLine >= startLine+1999 {
		// User didn't specify a range and file is large — give metadata first
		return metadata + "\n\n" +
			"This file is large. To read specific sections, use start_line and end_line parameters.\n" +
			"Example: read_file with start_line=1, end_line=100 for the first 100 lines.\n\n" +
			"Showing first 50 lines as preview:\n\n" +
			readLineRange(path, 1, 50), nil
	}

	// Read the requested range
	content := readLineRange(path, startLine, endLine)
	if content == "" {
		if totalLines == 0 {
			return "(empty file)", nil
		}
		return fmt.Sprintf("(no content at line %d, file has %d lines)", startLine, totalLines), nil
	}

	// Add range context
	actualEnd := startLine + strings.Count(content, "\n") - 1
	header := fmt.Sprintf("File: %s (lines %d-%d of %d)\n\n", path, startLine, actualEnd, totalLines)

	return header + content, nil
}

// countLines counts the total number of lines in a file.
func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	count := 0
	for scanner.Scan() {
		count++
	}
	return count, scanner.Err()
}

// isFileMinified checks if the first few lines of a file are extremely long.
func isFileMinified(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for i := 0; i < 3 && scanner.Scan(); i++ {
		if len(scanner.Text()) > MinifiedLineThreshold {
			return true
		}
	}
	return false
}

// readLineRange reads lines from startLine to endLine (inclusive, 1-based) and
// formats them with line numbers.
func readLineRange(path string, startLine, endLine int) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var result strings.Builder
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		if lineNum < startLine {
			continue
		}
		if lineNum > endLine {
			break
		}

		line := scanner.Text()
		if len(line) > MaxLineTruncation {
			line = line[:MaxLineTruncation] + "..."
		}

		fmt.Fprintf(&result, "%6d\t%s\n", lineNum, line)
	}

	return result.String()
}
