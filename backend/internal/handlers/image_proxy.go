package handlers

import (
	"claraverse/internal/security"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ImageProxyHandler handles image proxy requests
type ImageProxyHandler struct {
	client *http.Client
	cache  *imageCache
}

// imageCache provides in-memory caching for proxied images
type imageCache struct {
	mu       sync.RWMutex
	cache    map[string]*cachedImage
	maxSize  int64 // Max total cache size in bytes
	currSize int64 // Current cache size
}

type cachedImage struct {
	data        []byte
	contentType string
	timestamp   time.Time
	size        int64
}

const (
	maxImageSize     = 10 * 1024 * 1024  // 10MB max per image
	maxCacheSize     = 50 * 1024 * 1024  // 50MB total cache
	cacheTTL         = 10 * time.Minute
	requestTimeout   = 15 * time.Second
)

// NewImageProxyHandler creates a new image proxy handler
func NewImageProxyHandler() *ImageProxyHandler {
	return &ImageProxyHandler{
		client: &http.Client{
			Timeout: requestTimeout,
			// Don't follow redirects automatically - we'll handle them
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 3 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
		cache: &imageCache{
			cache:   make(map[string]*cachedImage),
			maxSize: maxCacheSize,
		},
	}
}

// ProxyImage handles GET /api/proxy/image?url={encoded_url}
func (h *ImageProxyHandler) ProxyImage(c *fiber.Ctx) error {
	// Get URL parameter
	imageURL := c.Query("url")
	if imageURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "url parameter is required",
		})
	}

	// Validate URL
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		log.Printf("⚠️ [IMAGE-PROXY] Invalid URL: %s - %v", imageURL, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid URL format",
		})
	}

	// Only allow http/https schemes
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		log.Printf("⚠️ [IMAGE-PROXY] Blocked non-http URL: %s", imageURL)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "only http and https URLs are allowed",
		})
	}

	// SSRF protection: block requests to internal/private networks
	if err := security.ValidateURLForSSRF(imageURL); err != nil {
		log.Printf("🚫 [IMAGE-PROXY] SSRF blocked: %s - %v", truncateURL(imageURL), err)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access to internal resources is not allowed",
		})
	}

	// Check cache first
	if cached := h.cache.get(imageURL); cached != nil {
		log.Printf("✅ [IMAGE-PROXY] Cache hit for: %s", truncateURL(imageURL))
		c.Set("Content-Type", cached.contentType)
		c.Set("Cache-Control", "public, max-age=3600")
		c.Set("X-Cache", "HIT")
		return c.Send(cached.data)
	}

	// Fetch the image
	log.Printf("🖼️ [IMAGE-PROXY] Fetching: %s", truncateURL(imageURL))

	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		log.Printf("❌ [IMAGE-PROXY] Failed to create request: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create request",
		})
	}

	// Set headers to look like a browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	// Set appropriate referer based on host (some sites like Bing require specific referers)
	referer := parsedURL.Scheme + "://" + parsedURL.Host + "/"
	host := strings.ToLower(parsedURL.Host)
	if strings.Contains(host, "bing.net") || strings.Contains(host, "bing.com") {
		referer = "https://www.bing.com/"
	} else if strings.Contains(host, "google") {
		referer = "https://www.google.com/"
	} else if strings.Contains(host, "duckduckgo") {
		referer = "https://duckduckgo.com/"
	}
	req.Header.Set("Referer", referer)

	resp, err := h.client.Do(req)
	if err != nil {
		log.Printf("❌ [IMAGE-PROXY] Fetch failed: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "failed to fetch image",
		})
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("⚠️ [IMAGE-PROXY] Upstream returned %d for: %s", resp.StatusCode, truncateURL(imageURL))
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "upstream server returned error",
		})
	}

	// Validate content type
	contentType := resp.Header.Get("Content-Type")
	if !isValidImageContentType(contentType) {
		log.Printf("⚠️ [IMAGE-PROXY] Invalid content type: %s for: %s", contentType, truncateURL(imageURL))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "URL does not point to a valid image",
		})
	}

	// Read body with size limit
	limitedReader := io.LimitReader(resp.Body, maxImageSize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		log.Printf("❌ [IMAGE-PROXY] Failed to read response: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to read image data",
		})
	}

	// Check if image exceeds size limit
	if int64(len(data)) > maxImageSize {
		log.Printf("⚠️ [IMAGE-PROXY] Image too large: %d bytes for: %s", len(data), truncateURL(imageURL))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "image exceeds maximum allowed size (10MB)",
		})
	}

	// Cache the image
	h.cache.set(imageURL, data, contentType)

	log.Printf("✅ [IMAGE-PROXY] Served: %s (%d bytes)", truncateURL(imageURL), len(data))

	// Send response
	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", "public, max-age=3600")
	c.Set("X-Cache", "MISS")
	return c.Send(data)
}

// isValidImageContentType checks if the content type is a valid image type
func isValidImageContentType(contentType string) bool {
	contentType = strings.ToLower(contentType)
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/svg+xml",
		"image/avif",
		"image/bmp",
		"image/tiff",
	}

	for _, vt := range validTypes {
		if strings.HasPrefix(contentType, vt) {
			return true
		}
	}
	return false
}

// truncateURL truncates URL for logging
func truncateURL(u string) string {
	if len(u) > 80 {
		return u[:77] + "..."
	}
	return u
}

// Cache methods

func (c *imageCache) get(key string) *cachedImage {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, exists := c.cache[key]
	if !exists {
		return nil
	}

	// Check TTL
	if time.Since(cached.timestamp) > cacheTTL {
		return nil
	}

	return cached
}

func (c *imageCache) set(key string, data []byte, contentType string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	size := int64(len(data))

	// If this single image is too large, don't cache it
	if size > c.maxSize/2 {
		return
	}

	// Evict old entries if needed
	for c.currSize+size > c.maxSize && len(c.cache) > 0 {
		c.evictOldest()
	}

	// Store in cache
	c.cache[key] = &cachedImage{
		data:        data,
		contentType: contentType,
		timestamp:   time.Now(),
		size:        size,
	}
	c.currSize += size
}

func (c *imageCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for k, v := range c.cache {
		if oldestKey == "" || v.timestamp.Before(oldestTime) {
			oldestKey = k
			oldestTime = v.timestamp
		}
	}

	if oldestKey != "" {
		c.currSize -= c.cache[oldestKey].size
		delete(c.cache, oldestKey)
	}
}
