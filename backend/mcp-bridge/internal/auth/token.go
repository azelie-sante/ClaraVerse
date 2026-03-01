package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	// TokenExpiryBuffer is how long before expiry we should refresh
	TokenExpiryBuffer = 5 * time.Minute
)

// Common errors
var (
	ErrTokenExpired         = errors.New("token has expired")
	ErrRefreshFailed        = errors.New("failed to refresh token")
	ErrInvalidToken         = errors.New("invalid token format")
	ErrAuthenticationFailed = errors.New("authentication failed")
)

// TokenResponse represents the response from token refresh
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	User         struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

// JWTClaims represents the relevant claims in a JWT
type JWTClaims struct {
	Exp int64  `json:"exp"` // Expiration time (Unix timestamp)
	Sub string `json:"sub"` // Subject (user ID)
	Iat int64  `json:"iat"` // Issued at
}

// GetBackendURL returns the backend URL from env or default
func GetBackendURL() string {
	url := os.Getenv("BACKEND_URL")
	if url == "" {
		url = "http://localhost:3001"
	}
	return strings.TrimSuffix(url, "/")
}

// IsTokenExpired checks if a token has expired based on the stored expiry time
func IsTokenExpired(tokenExpiry int64) bool {
	if tokenExpiry == 0 {
		return true // No expiry stored, assume expired
	}
	return time.Now().Unix() >= tokenExpiry
}

// IsTokenExpiringSoon checks if token will expire within the buffer period
func IsTokenExpiringSoon(tokenExpiry int64, buffer time.Duration) bool {
	if tokenExpiry == 0 {
		return true // No expiry stored, assume expiring
	}
	expiresAt := time.Unix(tokenExpiry, 0)
	return time.Until(expiresAt) <= buffer
}

// ParseJWTExpiry extracts the expiration timestamp from a JWT without verifying signature
// This is safe for checking expiry since the backend will verify the full token
func ParseJWTExpiry(token string) (int64, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, ErrInvalidToken
	}

	// Decode the payload (second part)
	payload := parts[1]
	// Add padding if needed
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		// Try standard encoding
		decoded, err = base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return 0, fmt.Errorf("failed to decode JWT payload: %w", err)
		}
	}

	var claims JWTClaims
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return 0, fmt.Errorf("failed to parse JWT claims: %w", err)
	}

	return claims.Exp, nil
}

// RefreshToken refreshes access token via the local backend's /api/auth/refresh endpoint
func RefreshToken(refreshToken string) (*TokenResponse, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("no refresh token available: %w", ErrRefreshFailed)
	}

	backendURL := GetBackendURL()

	reqBody := map[string]string{
		"refresh_token": refreshToken,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req, err := http.NewRequest("POST", backendURL+"/api/auth/refresh", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s (status: %d)", ErrRefreshFailed, string(body), resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("no access token in response: %w", ErrRefreshFailed)
	}

	return &tokenResp, nil
}

// RefreshDeviceToken refreshes a device token via the ClaraVerse backend
// In OSS mode, this is the same as RefreshToken since there's no separate device auth
func RefreshDeviceToken(backendURL, refreshToken, _ string) (*TokenResponse, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("no credentials available: %w", ErrRefreshFailed)
	}

	// Convert WebSocket URL to HTTP API URL
	apiURL := strings.Replace(backendURL, "ws://", "http://", 1)
	apiURL = strings.Replace(apiURL, "wss://", "https://", 1)
	apiURL = strings.TrimSuffix(apiURL, "/mcp/connect")

	reqBody := map[string]string{
		"refresh_token": refreshToken,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL+"/api/auth/refresh", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s (status: %d)", ErrRefreshFailed, string(body), resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("no access token in response: %w", ErrRefreshFailed)
	}

	return &tokenResp, nil
}

// ValidateAndRefreshIfNeeded checks if the token needs refreshing and does so if possible
// Returns: newAccessToken, newRefreshToken, newExpiry, error
func ValidateAndRefreshIfNeeded(accessToken, refreshToken string, tokenExpiry int64) (string, string, int64, error) {
	// Check if token is expired or expiring soon
	if !IsTokenExpiringSoon(tokenExpiry, TokenExpiryBuffer) {
		// Token is still valid
		return accessToken, refreshToken, tokenExpiry, nil
	}

	// Token needs refresh
	if refreshToken == "" {
		return "", "", 0, ErrTokenExpired
	}

	// Attempt to refresh
	newToken, err := RefreshToken(refreshToken)
	if err != nil {
		return "", "", 0, fmt.Errorf("token refresh failed: %w", err)
	}

	newExpiry := time.Now().Unix() + int64(newToken.ExpiresIn)
	return newToken.AccessToken, newToken.RefreshToken, newExpiry, nil
}
