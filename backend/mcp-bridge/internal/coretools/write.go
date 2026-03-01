package coretools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteTool writes content to a file with read-back verification.
// After writing, it returns the file content with line numbers so the LLM
// has the current state for subsequent edits.
type WriteTool struct {
	MaxFileSize int64
}

func NewWriteTool() *WriteTool {
	return &WriteTool{
		MaxFileSize: 10 * 1024 * 1024, // 10 MB
	}
}

func (t *WriteTool) Name() string { return "write_file" }

func (t *WriteTool) Description() string {
	return "Create or overwrite a file on the user's local filesystem with the given content.\n\n" +
		"Automatically creates any missing parent directories. " +
		"If the file already exists, its content is completely replaced. " +
		"After writing, returns a read-back of the file with line numbers " +
		"so you can verify the content is correct.\n\n" +
		"For large files (500+ lines), the read-back shows the first 100 and last 50 lines. " +
		"Max content size is 10MB.\n\n" +
		"Use this to create new files (source code, configs, scripts, etc.) or to completely " +
		"rewrite existing files. For partial edits to existing files, prefer string_replace " +
		"as it is safer and preserves the rest of the file."
}

func (t *WriteTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Absolute path to the file to write.",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The content to write to the file.",
			},
		},
		"required": []string{"path", "content"},
	}
}

func (t *WriteTool) Execute(_ context.Context, args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return "", fmt.Errorf("path is required")
	}

	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("content is required")
	}

	if int64(len(content)) > t.MaxFileSize {
		return "", fmt.Errorf("content too large (%d bytes, max %d)", len(content), t.MaxFileSize)
	}

	// Create parent directories
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Build response with read-back
	lines := strings.Split(content, "\n")
	totalLines := len(lines)
	tokens := EstimateTokens(content)

	var result strings.Builder
	fmt.Fprintf(&result, "Successfully wrote %d bytes to %s (%d lines, ~%d tokens)\n\n",
		len(content), path, totalLines, tokens)

	// For large files, show head + tail
	if totalLines > LargeFileReadbackThreshold {
		result.WriteString(FormatFileWithLineNumbers(content, LargeFileReadbackHead+LargeFileReadbackTail))
	} else {
		result.WriteString(formatLines(lines, 1))
	}

	return result.String(), nil
}
