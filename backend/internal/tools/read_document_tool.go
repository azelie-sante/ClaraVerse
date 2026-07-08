package tools

import (
	"claraverse/internal/filecache"
	"encoding/json"
	"fmt"
	"log"
)

// isDocumentMime reports whether a MIME type is a readable document (PDF/Word/
// PowerPoint). Used by the newest-in-conversation fallback so a follow-up turn
// resolves to the document the user attached earlier.
func isDocumentMime(mime string) bool {
	switch mime {
	case "application/pdf",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",     // .docx
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",   // .pptx
		"application/msword",           // .doc
		"application/vnd.ms-powerpoint": // .ppt
		return true
	}
	return false
}

// NewReadDocumentTool creates the read_document tool for reading PDF/DOCX/PPTX files
func NewReadDocumentTool() *Tool {
	return &Tool{
		Name:        "read_document",
		DisplayName: "Read Document",
		Description: "Extracts and returns the text content from an uploaded PDF, DOCX, or PPTX document. Use this to read document contents that were uploaded by the user.",
		Icon:        "FileText",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_id": map[string]interface{}{
					"type":        "string",
					"description": "The file ID of the uploaded document (from the upload response)",
				},
			},
			"required": []string{"file_id"},
		},
		Execute:  executeReadDocument,
		Source:   ToolSourceBuiltin,
		Category: "data_sources",
		Keywords: []string{"document", "pdf", "docx", "pptx", "read", "extract", "text", "file", "content", "word", "powerpoint"},
	}
}

func executeReadDocument(args map[string]interface{}) (string, error) {
	// Extract file_id parameter
	fileID, ok := args["file_id"].(string)
	if !ok || fileID == "" {
		return "", fmt.Errorf("file_id parameter is required and must be a string")
	}

	// Extract user context (injected by tool executor)
	userID, _ := args["__user_id__"].(string)
	conversationID, _ := args["__conversation_id__"].(string)

	// Clean up internal parameters
	delete(args, "__user_id__")
	delete(args, "__conversation_id__")

	log.Printf("📄 [READ-DOCUMENT] Reading document file_id=%s (user=%s)", fileID, userID)

	// Get file cache service
	fileCacheService := filecache.GetService()

	// Try to get file with ownership validation first
	var file *filecache.CachedFile
	var err error

	if userID != "" && conversationID != "" {
		file, err = fileCacheService.GetByUserAndConversation(fileID, userID, conversationID)
		if err != nil {
			// Try just by user (conversation might not match in workflow context)
			file, _ = fileCacheService.Get(fileID)
			if file != nil && file.UserID != userID {
				log.Printf("🚫 [READ-DOCUMENT] Access denied: file %s belongs to different user", fileID)
				return "", fmt.Errorf("access denied: you don't have permission to read this file")
			}
		}
	} else {
		// Fallback: get file without strict ownership check (for workflow context)
		file, _ = fileCacheService.Get(fileID)
	}

	// Models routinely pass the human FILENAME instead of the UUID file_id.
	if file == nil {
		if f := fileCacheService.ResolveByName(userID, conversationID, fileID); f != nil {
			log.Printf("📄 [READ-DOCUMENT] Resolved by filename %q → %s", fileID, f.FileID)
			file = f
		}
	}
	// Follow-up turn: the frontend doesn't re-send the attachment, and the model
	// often references it by a wrong/forgotten name. Fall back to the most recent
	// document the user attached in this conversation.
	if file == nil {
		if f := fileCacheService.NewestInConversation(userID, conversationID, isDocumentMime); f != nil {
			log.Printf("📄 [READ-DOCUMENT] %q unresolved → conversation's most recent document %s (%s)", fileID, f.FileID, f.Filename)
			file = f
		}
	}

	if file == nil {
		log.Printf("❌ [READ-DOCUMENT] File not found: %s", fileID)
		return "", fmt.Errorf("file not found — please upload it again")
	}

	// Validate file type is a document
	supportedTypes := map[string]bool{
		"application/pdf":                                                              true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document":      true, // .docx
		"application/vnd.openxmlformats-officedocument.presentationml.presentation":    true, // .pptx
		"application/msword":                                                           true, // .doc
		"application/vnd.ms-powerpoint":                                                true, // .ppt
	}

	if !supportedTypes[file.MimeType] {
		log.Printf("⚠️ [READ-DOCUMENT] Unsupported file type: %s", file.MimeType)
		return "", fmt.Errorf("unsupported file type: %s. Use read_data_file for CSV/JSON/text files", file.MimeType)
	}

	// Get extracted text
	var textContent string
	if file.ExtractedText != nil {
		textContent = file.ExtractedText.String()
	}

	if textContent == "" {
		log.Printf("⚠️ [READ-DOCUMENT] No text content extracted from file: %s", fileID)
		return "", fmt.Errorf("no text content could be extracted from this document")
	}

	// Build response
	response := map[string]interface{}{
		"success":    true,
		"file_id":    file.FileID,
		"filename":   file.Filename,
		"mime_type":  file.MimeType,
		"size":       file.Size,
		"page_count": file.PageCount,
		"word_count": file.WordCount,
		"content":    textContent,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	log.Printf("✅ [READ-DOCUMENT] Successfully read document %s: %d pages, %d words",
		file.Filename, file.PageCount, file.WordCount)

	return string(responseJSON), nil
}
