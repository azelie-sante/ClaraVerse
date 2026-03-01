package handlers

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"claraverse/internal/models"
	"claraverse/internal/services"

	"github.com/gofiber/fiber/v2"
)

// DeviceAuthHandler handles device authentication endpoints
type DeviceAuthHandler struct {
	deviceService *services.DeviceService
}

// NewDeviceAuthHandler creates a new device auth handler
func NewDeviceAuthHandler(deviceService *services.DeviceService) *DeviceAuthHandler {
	return &DeviceAuthHandler{
		deviceService: deviceService,
	}
}

// GenerateDeviceCode handles POST /api/device/code
// This is a public endpoint - no authentication required
func (h *DeviceAuthHandler) GenerateDeviceCode(c *fiber.Ctx) error {
	var req struct {
		ClientID string `json:"client_id"`
		Version  string `json:"client_version"`
		Platform string `json:"platform"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.ClientID == "" {
		req.ClientID = "mcp-client"
	}
	if req.Platform == "" {
		req.Platform = "unknown"
	}

	clientInfo := models.DeviceClientInfo{
		ClientID: req.ClientID,
		Version:  req.Version,
		Platform: req.Platform,
	}

	// Get client IP
	ipAddress := c.IP()
	forwardedFor := c.Get("X-Forwarded-For")
	if forwardedFor != "" {
		// Take the first IP in the chain
		ipAddress = strings.Split(forwardedFor, ",")[0]
		ipAddress = strings.TrimSpace(ipAddress)
	}

	userAgent := c.Get("User-Agent")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.deviceService.GenerateDeviceCode(ctx, clientInfo, ipAddress, userAgent)
	if err != nil {
		log.Printf("Error generating device code: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate device code",
		})
	}

	log.Printf("📱 Device code generated: %s (platform: %s)", response.UserCode, req.Platform)

	return c.JSON(response)
}

// PollForToken handles GET /api/device/token
// This is a public endpoint - CLI polls this to get tokens
func (h *DeviceAuthHandler) PollForToken(c *fiber.Ctx) error {
	deviceCode := c.Query("device_code")
	clientID := c.Query("client_id")

	if deviceCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.DeviceTokenErrorResponse{
			Error:            "invalid_request",
			ErrorDescription: "Missing device_code parameter.",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tokenResponse, errorResponse := h.deviceService.PollForToken(ctx, deviceCode, clientID)

	if errorResponse != nil {
		// Return appropriate HTTP status based on error type
		status := fiber.StatusBadRequest
		switch errorResponse.Error {
		case models.DeviceTokenErrorAuthorizationPending:
			status = fiber.StatusOK // RFC 8628: 200 OK with error in body
		case models.DeviceTokenErrorSlowDown:
			status = fiber.StatusOK // RFC 8628: 200 OK with error in body
		case models.DeviceTokenErrorExpiredToken:
			status = fiber.StatusBadRequest
		case models.DeviceTokenErrorAccessDenied:
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(errorResponse)
	}

	log.Printf("✅ Device authorized: user=%s, device=%s", tokenResponse.User.ID, tokenResponse.DeviceID)

	return c.JSON(tokenResponse)
}

// AuthorizeDevice handles POST /api/device/authorize
// This requires authentication - user enters code in browser
func (h *DeviceAuthHandler) AuthorizeDevice(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		userID = ""
	}
	userEmail, _ := c.Locals("user_email").(string)

	if userID == "" || userID == "anonymous" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "authentication_required",
			"message": "Please log in to authorize a device.",
		})
	}

	var req models.DeviceAuthorizeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Invalid request body.",
		})
	}

	if req.UserCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_code",
			"message": "Please enter a device code.",
		})
	}

	// Normalize user code
	userCode := strings.ToUpper(strings.ReplaceAll(req.UserCode, " ", ""))
	userCode = strings.ReplaceAll(userCode, "-", "")

	if len(userCode) != 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_code",
			"message": "Invalid code format. Code should be 8 characters.",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.deviceService.AuthorizeDevice(ctx, req.UserCode, userID, userEmail)
	if err != nil {
		log.Printf("Device authorization failed: %v", err)

		if errors.Is(err, services.ErrInvalidUserCode) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "invalid_code",
				"message": "The code you entered is invalid or has expired.",
			})
		}
		if errors.Is(err, services.ErrDeviceCodeExpired) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "expired_code",
				"message": "This code has expired. Please request a new code from your CLI.",
			})
		}
		if errors.Is(err, services.ErrMaxAttemptsExceeded) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "too_many_attempts",
				"message": "Too many failed attempts. Please request a new code.",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "authorization_failed",
			"message": "Failed to authorize device. Please try again.",
		})
	}

	log.Printf("✅ User %s authorized device (platform: %s)", userID, response.DeviceInfo.Platform)

	return c.JSON(response)
}

// GetPendingAuth handles GET /api/device/pending
func (h *DeviceAuthHandler) GetPendingAuth(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"pending": []interface{}{},
	})
}

// ListDevices handles GET /api/devices
func (h *DeviceAuthHandler) ListDevices(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		userID = ""
	}

	if userID == "" || userID == "anonymous" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Get current device ID from token if available
	currentDeviceID := ""
	if claims, ok := c.Locals("device_claims").(*services.DeviceTokenClaims); ok {
		currentDeviceID = claims.DeviceID
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	devices, err := h.deviceService.ListUserDevices(ctx, userID, currentDeviceID)
	if err != nil {
		log.Printf("Error listing devices: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list devices",
		})
	}

	return c.JSON(models.DeviceListResponse{
		Devices: devices,
	})
}

// UpdateDevice handles PUT /api/devices/:id
func (h *DeviceAuthHandler) UpdateDevice(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		userID = ""
	}
	deviceID := c.Params("id")

	if userID == "" || userID == "anonymous" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	var req models.DeviceUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}

	// Limit name length
	if len(req.Name) > 50 {
		req.Name = req.Name[:50]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := h.deviceService.UpdateDevice(ctx, userID, deviceID, req.Name)
	if err != nil {
		if errors.Is(err, services.ErrDeviceNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Device not found",
			})
		}
		log.Printf("Error updating device: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update device",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Device updated successfully",
	})
}

// RevokeDevice handles DELETE /api/devices/:id
func (h *DeviceAuthHandler) RevokeDevice(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		userID = ""
	}
	deviceID := c.Params("id")

	if userID == "" || userID == "anonymous" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := h.deviceService.RevokeDevice(ctx, userID, deviceID)
	if err != nil {
		if errors.Is(err, services.ErrDeviceNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Device not found",
			})
		}
		log.Printf("Error revoking device: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to revoke device",
		})
	}

	log.Printf("🔒 Device %s revoked by user %s", deviceID, userID)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Device has been revoked.",
	})
}

// RefreshToken handles POST /api/devices/:id/refresh-token
func (h *DeviceAuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req models.DeviceRefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_request",
			"error_description": "Invalid request body.",
		})
	}

	if req.RefreshToken == "" || req.DeviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_request",
			"error_description": "Missing refresh_token or device_id.",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.deviceService.RefreshAccessToken(ctx, req.RefreshToken, req.DeviceID)
	if err != nil {
		log.Printf("Token refresh failed: %v", err)

		if errors.Is(err, services.ErrInvalidRefreshToken) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":             "invalid_grant",
				"error_description": "Invalid refresh token.",
			})
		}
		if errors.Is(err, services.ErrDeviceRevoked) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":             "invalid_grant",
				"error_description": "Device has been revoked.",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             "server_error",
			"error_description": "Failed to refresh token.",
		})
	}

	return c.JSON(response)
}
