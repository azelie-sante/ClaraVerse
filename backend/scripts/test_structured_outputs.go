package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Test schema for news articles (realistic production use case)
var newsSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"articles": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title":         map[string]interface{}{"type": "string"},
					"source":        map[string]interface{}{"type": "string"},
					"url":           map[string]interface{}{"type": "string"},
					"summary":       map[string]interface{}{"type": "string"},
					"publishedDate": map[string]interface{}{"type": "string"},
				},
				"required":             []string{"title", "source", "url", "summary", "publishedDate"},
				"additionalProperties": false,
			},
		},
		"totalResults": map[string]interface{}{"type": "number"},
		"fetchedAt":    map[string]interface{}{"type": "string"},
	},
	"required":             []string{"articles", "totalResults", "fetchedAt"},
	"additionalProperties": false,
}

// Provider configuration
type Provider struct {
	Name    string
	BaseURL string
	APIKey  string
	Models  []string
}

// Test result
type TestResult struct {
	Provider         string
	Model            string
	SupportsStrict   bool
	SupportsJSON     bool
	SchemaCompliance float64 // 0-100 score
	ResponseTime     time.Duration
	Error            string
	RawResponse      string
	ParsedCorrectly  bool
	MissingFields    []string
	ExtraFields      []string
}

func main() {
	fmt.Println("🧪 STRUCTURED OUTPUT TESTING SUITE")
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Println()

	fmt.Println("📋 Test Specification:")
	fmt.Println("   - Task: Generate 3 AI news articles with structured data")
	fmt.Println("   - Schema: News articles with title, source, url, summary, publishedDate")
	fmt.Println("   - Test 1: Strict JSON schema mode (OpenAI-style)")
	fmt.Println("   - Test 2: Fallback JSON object mode + schema in prompt")
	fmt.Println("   - Validation: Check for required fields, extra fields, compliance %")
	fmt.Println()

	fmt.Println("📂 Loading provider configuration from providers.json...")
	providers := loadProviders()

	if len(providers) == 0 {
		fmt.Println("❌ No providers found or all providers are disabled")
		return
	}

	fmt.Printf("✅ Found %d enabled providers\n", len(providers))

	// Count total models
	totalModels := 0
	for _, p := range providers {
		totalModels += len(p.Models)
	}
	fmt.Printf("📊 Will test %d models across %d providers (up to 2 tests per model)\n", totalModels, len(providers))
	fmt.Println()

	results := []TestResult{}
	currentTest := 0

	for _, provider := range providers {
		fmt.Printf("📦 Testing Provider: %s (%s)\n", provider.Name, provider.BaseURL)
		fmt.Printf("   Models to test: %v\n", provider.Models)
		fmt.Println(strings.Repeat("-", 80))

		for _, model := range provider.Models {
			currentTest++
			fmt.Printf("\n[%d/%d] Testing: %s / %s\n", currentTest, totalModels, provider.Name, model)
			fmt.Print("   ⏳ Attempting strict JSON schema mode... ")

			result := testModel(provider, model)
			results = append(results, result)
			printResult(result)
		}
		fmt.Println()
	}

	// Generate report
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📝 GENERATING FINAL REPORT")
	fmt.Println(strings.Repeat("=", 80) + "\n")
	generateReport(results)
}

func loadProviders() []Provider {
	// Load from providers.json file
	providersFile := "../providers.json"
	data, err := os.ReadFile(providersFile)
	if err != nil {
		fmt.Printf("⚠️  Could not read providers.json: %v\n", err)
		fmt.Println("Using empty provider list")
		return []Provider{}
	}

	var config struct {
		Providers []struct {
			Name          string                            `json:"name"`
			BaseURL       string                            `json:"base_url"`
			APIKey        string                            `json:"api_key"`
			Enabled       bool                              `json:"enabled"`
			AudioOnly     bool                              `json:"audio_only"`
			ImageOnly     bool                              `json:"image_only"`
			ImageEditOnly bool                              `json:"image_edit_only"`
			ModelAliases  map[string]map[string]interface{} `json:"model_aliases"`
		} `json:"providers"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("⚠️  Could not parse providers.json: %v\n", err)
		return []Provider{}
	}

	var providers []Provider
	for _, p := range config.Providers {
		// Skip disabled providers
		if !p.Enabled {
			continue
		}

		// Skip image-only and audio-only providers (they don't support chat completions)
		if p.AudioOnly || p.ImageOnly || p.ImageEditOnly {
			continue
		}

		// Extract actual model IDs from model_aliases (use actual_model value, not the alias key)
		var models []string
		for aliasKey, aliasData := range p.ModelAliases {
			// Try to get actual_model, fall back to alias key if not found
			if actualModel, ok := aliasData["actual_model"].(string); ok && actualModel != "" {
				models = append(models, actualModel)
			} else {
				models = append(models, aliasKey)
			}
		}

		// Skip if no models configured
		if len(models) == 0 {
			continue
		}

		// Limit to first 3 models per provider for faster testing
		if len(models) > 3 {
			models = models[:3]
		}

		providers = append(providers, Provider{
			Name:    p.Name,
			BaseURL: p.BaseURL,
			APIKey:  p.APIKey,
			Models:  models,
		})
	}

	return providers
}

func testModel(provider Provider, model string) TestResult {
	result := TestResult{
		Provider: provider.Name,
		Model:    model,
	}

	// Skip if no API key
	if provider.APIKey == "" {
		fmt.Println("❌ (No API key)")
		result.Error = "API key not configured"
		return result
	}

	fmt.Println()

	// Test 1: Strict JSON Schema mode
	fmt.Println("   📝 Test 1/2: Strict JSON Schema Mode")
	fmt.Print("      Sending request with response_format={type: json_schema, strict: true}... ")
	strictResult := testStrictMode(provider, model)
	result.SupportsStrict = strictResult.Success
	result.ResponseTime = strictResult.Duration

	if strictResult.Success {
		fmt.Printf("✅ Success (%v)\n", strictResult.Duration)
		fmt.Print("      Validating schema compliance... ")

		// Validate schema compliance
		compliance := validateSchemaCompliance(strictResult.Response)
		result.SchemaCompliance = compliance.Score
		result.ParsedCorrectly = compliance.Valid
		result.MissingFields = compliance.MissingFields
		result.ExtraFields = compliance.ExtraFields
		result.RawResponse = truncate(strictResult.Response, 200)

		fmt.Printf("%.1f%%\n", compliance.Score)
	} else {
		fmt.Printf("❌ Failed\n")
		fmt.Printf("      Error: %s\n", truncate(strictResult.Error, 80))
		result.Error = strictResult.Error

		// Test 2: Fallback to basic JSON mode
		fmt.Println()
		fmt.Println("   📝 Test 2/2: Fallback JSON Object Mode")
		fmt.Print("      Sending request with response_format={type: json_object} + schema in prompt... ")
		jsonResult := testJSONMode(provider, model)
		result.SupportsJSON = jsonResult.Success

		if jsonResult.Success {
			fmt.Printf("✅ Success (%v)\n", jsonResult.Duration)
			fmt.Print("      Validating schema compliance... ")

			compliance := validateSchemaCompliance(jsonResult.Response)
			result.SchemaCompliance = compliance.Score
			result.ParsedCorrectly = compliance.Valid
			result.MissingFields = compliance.MissingFields
			result.ExtraFields = compliance.ExtraFields
			result.RawResponse = truncate(jsonResult.Response, 200)

			fmt.Printf("%.1f%%\n", compliance.Score)
		} else {
			fmt.Printf("❌ Failed\n")
			fmt.Printf("      Error: %s\n", truncate(jsonResult.Error, 80))
		}
	}

	fmt.Println()

	return result
}

type APITestResult struct {
	Success  bool
	Response string
	Error    string
	Duration time.Duration
}

func testStrictMode(provider Provider, model string) APITestResult {
	start := time.Now()

	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a news aggregator. Return news in the exact schema format."},
			{"role": "user", "content": "Get 3 AI news articles from today"},
		},
		"temperature": 0.3,
		"response_format": map[string]interface{}{
			"type": "json_schema",
			"json_schema": map[string]interface{}{
				"name":   "news_output",
				"strict": true,
				"schema": newsSchema,
			},
		},
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return APITestResult{Success: false, Error: fmt.Sprintf("Marshal error: %v", err)}
	}

	req, err := http.NewRequest("POST", provider.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return APITestResult{Success: false, Error: fmt.Sprintf("Request error: %v", err)}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+provider.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return APITestResult{Success: false, Error: fmt.Sprintf("HTTP error: %v", err)}
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		errorMsg := string(bodyBytes)
		// Try to parse error response for better readability
		var errorResp map[string]interface{}
		if json.Unmarshal(bodyBytes, &errorResp) == nil {
			if errObj, ok := errorResp["error"].(map[string]interface{}); ok {
				if message, ok := errObj["message"].(string); ok {
					errorMsg = message
				}
			}
		}
		return APITestResult{
			Success:  false,
			Error:    fmt.Sprintf("Status %d: %s", resp.StatusCode, errorMsg),
			Duration: duration,
		}
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return APITestResult{Success: false, Error: fmt.Sprintf("Decode error: %v", err)}
	}

	// Extract content
	content := extractContent(result)
	return APITestResult{
		Success:  true,
		Response: content,
		Duration: duration,
	}
}

func testJSONMode(provider Provider, model string) APITestResult {
	start := time.Now()

	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a news aggregator. Return news as valid JSON with articles array, totalResults, and fetchedAt fields."},
			{"role": "user", "content": "Get 3 AI news articles from today. Return as JSON."},
		},
		"temperature": 0.3,
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return APITestResult{Success: false, Error: fmt.Sprintf("Marshal error: %v", err)}
	}

	req, err := http.NewRequest("POST", provider.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return APITestResult{Success: false, Error: fmt.Sprintf("Request error: %v", err)}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+provider.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return APITestResult{Success: false, Error: fmt.Sprintf("HTTP error: %v", err)}
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		errorMsg := string(bodyBytes)
		// Try to parse error response for better readability
		var errorResp map[string]interface{}
		if json.Unmarshal(bodyBytes, &errorResp) == nil {
			if errObj, ok := errorResp["error"].(map[string]interface{}); ok {
				if message, ok := errObj["message"].(string); ok {
					errorMsg = message
				}
			}
		}
		return APITestResult{
			Success:  false,
			Error:    fmt.Sprintf("Status %d: %s", resp.StatusCode, errorMsg),
			Duration: duration,
		}
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return APITestResult{Success: false, Error: fmt.Sprintf("Decode error: %v", err)}
	}

	content := extractContent(result)
	return APITestResult{
		Success:  true,
		Response: content,
		Duration: duration,
	}
}

func extractContent(response map[string]interface{}) string {
	if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					return content
				}
			}
		}
	}
	return ""
}

type ComplianceResult struct {
	Valid         bool
	Score         float64
	MissingFields []string
	ExtraFields   []string
}

func validateSchemaCompliance(jsonStr string) ComplianceResult {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return ComplianceResult{Valid: false, Score: 0}
	}

	missing := []string{}
	extra := []string{}
	score := 100.0

	// Check required top-level fields
	requiredFields := []string{"articles", "totalResults", "fetchedAt"}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			missing = append(missing, field)
			score -= 33.33
		}
	}

	// Check for extra fields (should have additionalProperties: false)
	allowedFields := map[string]bool{"articles": true, "totalResults": true, "fetchedAt": true}
	for field := range data {
		if !allowedFields[field] {
			extra = append(extra, field)
			score -= 10
		}
	}

	// Check articles array structure
	if articles, ok := data["articles"].([]interface{}); ok {
		for i, article := range articles {
			if articleMap, ok := article.(map[string]interface{}); ok {
				articleRequired := []string{"title", "source", "url", "summary"}
				for _, field := range articleRequired {
					if _, ok := articleMap[field]; !ok {
						missing = append(missing, fmt.Sprintf("articles[%d].%s", i, field))
						score -= 5
					}
				}
			}
		}
	}

	if score < 0 {
		score = 0
	}

	return ComplianceResult{
		Valid:         len(missing) == 0 && len(extra) == 0,
		Score:         score,
		MissingFields: missing,
		ExtraFields:   extra,
	}
}

func printResult(result TestResult) {
	// Summary line
	status := "❌ NO SUPPORT"
	if result.SupportsStrict {
		status = "✅ FULL SUPPORT"
	} else if result.SupportsJSON {
		status = "⚠️  PARTIAL SUPPORT"
	}

	fmt.Printf("   Result: %s (Compliance: %.1f%%)\n", status, result.SchemaCompliance)

	if len(result.MissingFields) > 0 {
		fmt.Printf("   ⚠️  Missing Fields: %v\n", result.MissingFields)
	}

	if len(result.ExtraFields) > 0 {
		fmt.Printf("   ⚠️  Extra Fields: %v\n", result.ExtraFields)
	}

	if result.Error != "" && !result.SupportsStrict && !result.SupportsJSON {
		fmt.Printf("   Error: %s\n", truncate(result.Error, 100))
	}
}

func generateReport(results []TestResult) {
	fmt.Println()
	fmt.Println("📊 SUMMARY REPORT")
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Println()

	// Group by support level
	strictSupport := []TestResult{}
	jsonSupport := []TestResult{}
	noSupport := []TestResult{}

	for _, r := range results {
		if r.SupportsStrict {
			strictSupport = append(strictSupport, r)
		} else if r.SupportsJSON {
			jsonSupport = append(jsonSupport, r)
		} else {
			noSupport = append(noSupport, r)
		}
	}

	fmt.Printf("✅ FULL SUPPORT (Strict JSON Schema): %d models\n", len(strictSupport))
	for _, r := range strictSupport {
		fmt.Printf("   • %s / %s (%.1f%% compliance, %v)\n",
			r.Provider, r.Model, r.SchemaCompliance, r.ResponseTime)
	}
	fmt.Println()

	fmt.Printf("⚠️  PARTIAL SUPPORT (JSON Object only): %d models\n", len(jsonSupport))
	for _, r := range jsonSupport {
		fmt.Printf("   • %s / %s (%.1f%% compliance)\n",
			r.Provider, r.Model, r.SchemaCompliance)
	}
	fmt.Println()

	fmt.Printf("❌ NO SUPPORT: %d models\n", len(noSupport))
	for _, r := range noSupport {
		fmt.Printf("   • %s / %s - %s\n", r.Provider, r.Model, r.Error)
	}
	fmt.Println()

	// Recommendations
	fmt.Println("💡 RECOMMENDATIONS FOR PRODUCTION")
	fmt.Println(strings.Repeat("-", 80))

	if len(strictSupport) > 0 {
		best := strictSupport[0]
		for _, r := range strictSupport {
			if r.SchemaCompliance > best.SchemaCompliance {
				best = r
			}
		}
		fmt.Printf("🏆 BEST: %s / %s (%.1f%% compliance, %v response time)\n",
			best.Provider, best.Model, best.SchemaCompliance, best.ResponseTime)
		fmt.Println("   → Use for production workflows requiring guaranteed structure")
	}

	if len(jsonSupport) > 0 {
		fmt.Println()
		fmt.Println("⚠️  FALLBACK OPTIONS (require prompt engineering):")
		for _, r := range jsonSupport {
			fmt.Printf("   • %s / %s - Add schema to system prompt\n", r.Provider, r.Model)
		}
	}

	fmt.Println()
	fmt.Println("🔧 IMPLEMENTATION STRATEGY:")
	fmt.Println("   1. Detect provider capability at runtime")
	fmt.Println("   2. Use strict mode for OpenAI and compatible providers")
	fmt.Println("   3. Fallback to JSON mode + prompt schema for others")
	fmt.Println("   4. Validate output and retry with stronger prompts if needed")
}

func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}
