package handlers

import (
	"claraverse/internal/securefile"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
)

// SecureDownloadHandler handles secure file downloads with access codes
type SecureDownloadHandler struct {
	secureFileService *securefile.Service
}

// NewSecureDownloadHandler creates a new secure download handler
func NewSecureDownloadHandler() *SecureDownloadHandler {
	return &SecureDownloadHandler{
		secureFileService: securefile.GetService(),
	}
}

// Download handles file downloads with access code validation
// GET /api/files/:id?code=ACCESS_CODE
func (h *SecureDownloadHandler) Download(c *fiber.Ctx) error {
	fileID := c.Params("id")
	accessCode := c.Query("code")

	if fileID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "file_id is required",
		})
	}

	if accessCode == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "access code is required",
		})
	}

	log.Printf("📥 [SECURE-DOWNLOAD] Download request for file %s", fileID)

	// Get file with access code verification
	file, content, err := h.secureFileService.GetFile(fileID, accessCode)
	if err != nil {
		log.Printf("❌ [SECURE-DOWNLOAD] Failed to get file %s: %v", fileID, err)

		if err.Error() == "invalid access code" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid access code",
			})
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "file not found or expired",
		})
	}

	// Set headers for download
	c.Set("Content-Type", file.MimeType)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.Filename))
	c.Set("Content-Length", fmt.Sprintf("%d", file.Size))
	c.Set("X-File-ID", file.ID)
	c.Set("X-Expires-At", file.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"))

	log.Printf("✅ [SECURE-DOWNLOAD] Serving file %s (%s, %d bytes)", file.ID, file.Filename, file.Size)

	return c.Send(content)
}

// GetInfo returns file metadata without downloading
// GET /api/files/:id/info?code=ACCESS_CODE
func (h *SecureDownloadHandler) GetInfo(c *fiber.Ctx) error {
	fileID := c.Params("id")
	accessCode := c.Query("code")

	if fileID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "file_id is required",
		})
	}

	if accessCode == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "access code is required",
		})
	}

	file, err := h.secureFileService.GetFileInfo(fileID, accessCode)
	if err != nil {
		if err.Error() == "invalid access code" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid access code",
			})
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "file not found or expired",
		})
	}

	return c.JSON(fiber.Map{
		"id":         file.ID,
		"filename":   file.Filename,
		"mime_type":  file.MimeType,
		"size":       file.Size,
		"created_at": file.CreatedAt,
		"expires_at": file.ExpiresAt,
	})
}

// Delete removes a file (requires authentication and ownership)
// DELETE /api/files/:id
func (h *SecureDownloadHandler) Delete(c *fiber.Ctx) error {
	fileID := c.Params("id")
	userID := c.Locals("user_id")

	if fileID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "file_id is required",
		})
	}

	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	err := h.secureFileService.DeleteFile(fileID, userID.(string))
	if err != nil {
		if err.Error() == "access denied" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "you don't have permission to delete this file",
			})
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "file not found",
		})
	}

	log.Printf("✅ [SECURE-DOWNLOAD] File %s deleted by user %s", fileID, userID)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "file deleted",
	})
}

// ListUserFiles returns all files for the authenticated user
// GET /api/files
func (h *SecureDownloadHandler) ListUserFiles(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	files := h.secureFileService.ListUserFiles(userID.(string))

	// Convert to response format (without sensitive data)
	response := make([]fiber.Map, 0, len(files))
	for _, file := range files {
		response = append(response, fiber.Map{
			"id":         file.ID,
			"filename":   file.Filename,
			"mime_type":  file.MimeType,
			"size":       file.Size,
			"created_at": file.CreatedAt,
			"expires_at": file.ExpiresAt,
		})
	}

	return c.JSON(fiber.Map{
		"files": response,
		"count": len(response),
	})
}
