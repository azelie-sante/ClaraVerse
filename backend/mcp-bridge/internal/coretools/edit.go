package coretools

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// EditTool performs string replacement edits in files and returns the updated
// file content with line numbers for reliable subsequent edits.
type EditTool struct{}

func NewEditTool() *EditTool {
	return &EditTool{}
}

func (t *EditTool) Name() string { return "string_replace" }

func (t *EditTool) Description() string {
	return "Edit an existing file by finding and replacing a specific string.\n\n" +
		"The old_string must appear exactly once in the file to ensure the correct location is edited. " +
		"If it appears zero times, the file may have changed — re-read it first. " +
		"If it appears multiple times, include more surrounding context to make it unique.\n\n" +
		"To delete code, set new_string to an empty string. " +
		"Returns the full updated file with line numbers showing the change location.\n\n" +
		"This is the preferred way to make targeted changes to existing files — " +
		"it is safer than rewriting the entire file with write_file because it only " +
		"touches the specific section you want to change, reducing the risk of " +
		"accidentally losing other content."
}

func (t *EditTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Absolute path to the file to edit.",
			},
			"old_string": map[string]interface{}{
				"type":        "string",
				"description": "The exact string to find and replace. Must appear exactly once in the file.",
			},
			"new_string": map[string]interface{}{
				"type":        "string",
				"description": "The replacement string.",
			},
		},
		"required": []string{"path", "old_string", "new_string"},
	}
}

func (t *EditTool) Execute(_ context.Context, args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return "", fmt.Errorf("path is required")
	}

	oldString, ok := args["old_string"].(string)
	if !ok || oldString == "" {
		return "", fmt.Errorf("old_string is required")
	}

	newString, _ := args["new_string"].(string) // Can be empty (deletion)

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w. The file may have been moved or deleted", err)
	}

	content := string(data)

	// Check uniqueness
	count := strings.Count(content, oldString)
	if count == 0 {
		return "", fmt.Errorf("old_string not found in file. " +
			"The file may have changed since you last read it. " +
			"Use read_file to see the current content before retrying")
	}
	if count > 1 {
		return "", fmt.Errorf("old_string appears %d times in file (must be unique). "+
			"Provide more surrounding context to make it unique", count)
	}

	// Find the line range of the replacement for context
	beforeReplace := content[:strings.Index(content, oldString)]
	oldStartLine := strings.Count(beforeReplace, "\n") + 1
	oldEndLine := oldStartLine + strings.Count(oldString, "\n")
	newEndLine := oldStartLine + strings.Count(newString, "\n")

	// Perform replacement
	newContent := strings.Replace(content, oldString, newString, 1)

	// Write back
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Build response with context about the change
	lines := strings.Split(newContent, "\n")
	totalLines := len(lines)

	var result strings.Builder
	fmt.Fprintf(&result, "Replaced content at lines %d-%d (now lines %d-%d) in %s (%d total lines)\n\n",
		oldStartLine, oldEndLine, oldStartLine, newEndLine, path, totalLines)

	// Return full file content (truncated for large files)
	if totalLines > LargeFileReadbackThreshold {
		result.WriteString(FormatFileWithLineNumbers(newContent, LargeFileReadbackHead+LargeFileReadbackTail))
	} else {
		result.WriteString(formatLines(lines, 1))
	}

	return result.String(), nil
}
