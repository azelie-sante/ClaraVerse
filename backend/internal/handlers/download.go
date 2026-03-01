package handlers

import (
	"claraverse/internal/document"
	"log"

	"github.com/gofiber/fiber/v2"
)

// DownloadHandler handles file download requests
type DownloadHandler struct {
	documentService *document.Service
}

// NewDownloadHandler creates a new download handler
func NewDownloadHandler() *DownloadHandler {
	return &DownloadHandler{
		documentService: document.GetService(),
	}
}

// Download serves a generated document and marks it for deletion
func (h *DownloadHandler) Download(c *fiber.Ctx) error {
	documentID := c.Params("id")

	// Get user ID from auth middleware
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" || userID == "anonymous" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required to download documents",
		})
	}

	// Get document
	doc, exists := h.documentService.GetDocument(documentID)
	if !exists {
		log.Printf("⚠️  [DOWNLOAD] Document not found: %s (user: %s)", documentID, userID)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Document not found or already deleted",
		})
	}

	// Verify ownership
	if doc.UserID != userID {
		log.Printf("🚫 [SECURITY] User %s denied access to document %s (owned by %s)",
			userID, documentID, doc.UserID)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied to this document",
		})
	}

	log.Printf("📥 [DOWNLOAD] Serving document: %s (user: %s, size: %d bytes)",
		doc.Filename, doc.UserID, doc.Size)

	// Determine content type
	contentType := doc.ContentType
	if contentType == "" {
		contentType = "application/octet-stream" // Fallback for unknown types
	}

	// Set headers for download
	c.Set("Content-Disposition", "attachment; filename=\""+doc.Filename+"\"")
	c.Set("Content-Type", contentType)

	// Send file
	err := c.SendFile(doc.FilePath)
	if err != nil {
		log.Printf("❌ [DOWNLOAD] Failed to send file: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to download file",
		})
	}

	// Mark as downloaded (will be deleted in 5 minutes by cleanup job)
	h.documentService.MarkDownloaded(documentID)

	log.Printf("✅ [DOWNLOAD] Document downloaded: %s (user: %s)", doc.Filename, userID)

	return nil
}
