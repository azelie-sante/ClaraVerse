package coretools

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPRequestTool makes HTTP requests and returns the response.
type HTTPRequestTool struct {
	MaxBodySize int64 // max response body to read (default 1MB)
}

func NewHTTPRequestTool() *HTTPRequestTool {
	return &HTTPRequestTool{
		MaxBodySize: 1 * 1024 * 1024, // 1 MB
	}
}

func (t *HTTPRequestTool) Name() string { return "http_request" }

func (t *HTTPRequestTool) Description() string {
	return "Make an HTTP request to a URL and return the response.\n\n" +
		"Supported methods: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS.\n\n" +
		"Returns the HTTP status code, selected response headers (Content-Type, Location, " +
		"Set-Cookie, etc.), and the response body. Large response bodies are truncated " +
		"to 1MB with smart truncation preserving the beginning and end.\n\n" +
		"Common use cases:\n" +
		"- Check if a dev server is running: method=GET, url=http://localhost:5173\n" +
		"- Test an API endpoint: method=POST, url=http://localhost:3000/api/users, " +
		"body='{\"name\":\"test\"}', headers={'Content-Type': 'application/json'}\n" +
		"- Download a file's content: method=GET, url=https://example.com/data.json\n" +
		"- Check redirects: method=HEAD, url=https://example.com (follows redirects by default)\n\n" +
		"Timeout defaults to 30 seconds. TLS certificate verification is enabled by default; " +
		"set insecure=true for self-signed certificates (e.g., local development)."
}

func (t *HTTPRequestTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"method": map[string]interface{}{
				"type":        "string",
				"description": "HTTP method: GET, POST, PUT, DELETE, PATCH, HEAD, or OPTIONS. Defaults to GET.",
				"enum":        []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
			},
			"url": map[string]interface{}{
				"type":        "string",
				"description": "The full URL to request. Must include the scheme (http:// or https://). Example: http://localhost:5173, https://api.example.com/v1/users",
			},
			"headers": map[string]interface{}{
				"type":        "object",
				"description": "Request headers as key-value pairs. Example: {\"Content-Type\": \"application/json\", \"Authorization\": \"Bearer token123\"}",
			},
			"body": map[string]interface{}{
				"type":        "string",
				"description": "Request body (for POST, PUT, PATCH). Send JSON by setting the Content-Type header to application/json.",
			},
			"timeout_ms": map[string]interface{}{
				"type":        "integer",
				"description": "Request timeout in milliseconds. Default: 30000 (30 seconds). Max: 120000 (2 minutes).",
			},
			"insecure": map[string]interface{}{
				"type":        "boolean",
				"description": "Skip TLS certificate verification. Use for self-signed certs in local development. Default: false.",
			},
		},
		"required": []string{"url"},
	}
}

func (t *HTTPRequestTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	urlStr, ok := args["url"].(string)
	if !ok || urlStr == "" {
		return "", fmt.Errorf("url is required")
	}

	// Validate URL scheme
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return "", fmt.Errorf("url must start with http:// or https://")
	}

	method := "GET"
	if m, ok := args["method"].(string); ok && m != "" {
		method = strings.ToUpper(m)
	}

	timeout := 30 * time.Second
	if tm, ok := args["timeout_ms"].(float64); ok && tm > 0 {
		timeout = time.Duration(tm) * time.Millisecond
		if timeout > 2*time.Minute {
			timeout = 2 * time.Minute
		}
	}

	insecure := false
	if ins, ok := args["insecure"].(bool); ok {
		insecure = ins
	}

	// Build request body
	var bodyReader io.Reader
	if body, ok := args["body"].(string); ok && body != "" {
		bodyReader = strings.NewReader(body)
	}

	// Create request
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if headers, ok := args["headers"].(map[string]interface{}); ok {
		for key, val := range headers {
			if strVal, ok := val.(string); ok {
				req.Header.Set(key, strVal)
			}
		}
	}

	// Set default User-Agent if not provided
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "Clara-Companion/1.0")
	}

	// Build HTTP client
	transport := &http.Transport{}
	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("request timed out after %v", timeout)
		}
		// Provide helpful error messages for common failures
		errMsg := err.Error()
		if strings.Contains(errMsg, "connection refused") {
			return fmt.Sprintf("Connection refused at %s — the server is not running or not listening on this port.", urlStr), nil
		}
		if strings.Contains(errMsg, "no such host") {
			return fmt.Sprintf("DNS lookup failed for %s — the hostname could not be resolved.", urlStr), nil
		}
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Build response
	var result strings.Builder

	// Status line
	fmt.Fprintf(&result, "HTTP %d %s\n", resp.StatusCode, resp.Status)

	// Selected response headers
	interestingHeaders := []string{
		"Content-Type", "Content-Length", "Location",
		"Set-Cookie", "X-Request-Id", "X-Powered-By",
		"Server", "Cache-Control", "Access-Control-Allow-Origin",
	}

	headerCount := 0
	for _, h := range interestingHeaders {
		if v := resp.Header.Get(h); v != "" {
			if headerCount == 0 {
				result.WriteString("\nHeaders:\n")
			}
			fmt.Fprintf(&result, "  %s: %s\n", h, v)
			headerCount++
		}
	}

	// Read body (HEAD requests have no body)
	if method != "HEAD" {
		body, err := io.ReadAll(io.LimitReader(resp.Body, t.MaxBodySize+1))
		if err != nil {
			fmt.Fprintf(&result, "\nBody: (error reading: %v)\n", err)
		} else if len(body) > 0 {
			bodyStr := string(body)
			truncated := false

			if int64(len(body)) > t.MaxBodySize {
				bodyStr = TruncateSmartly(bodyStr, int(t.MaxBodySize))
				truncated = true
			}

			// For very large bodies, also apply the standard output limit
			if len(bodyStr) > OutputLimitChars {
				bodyStr = TruncateSmartly(bodyStr, OutputLimitChars)
				truncated = true
			}

			result.WriteString("\nBody:\n")
			result.WriteString(bodyStr)

			if truncated {
				fmt.Fprintf(&result, "\n\n(body truncated — original size: %d bytes)", len(body))
			}
		} else {
			result.WriteString("\nBody: (empty)")
		}
	}

	return result.String(), nil
}
