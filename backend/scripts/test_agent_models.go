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

// Test configuration
const (
	testPrompt = "Write a simple 'Hello World' function in Python."
	numTests   = 3 // Run each test 3 times and average
)

// Provider and model structures
type ModelAlias struct {
	ActualModel                string `json:"actual_model"`
	DisplayName                string `json:"display_name"`
	Description                string `json:"description,omitempty"`
	SupportsVision             *bool  `json:"supports_vision,omitempty"`
	Agents                     *bool  `json:"agents,omitempty"`
	StructuredOutputSupport    string `json:"structured_output_support,omitempty"`
	StructuredOutputCompliance *int   `json:"structured_output_compliance,omitempty"`
	StructuredOutputWarning    string `json:"structured_output_warning,omitempty"`
	StructuredOutputSpeedMs    *int   `json:"structured_output_speed_ms,omitempty"`
	StructuredOutputBadge      string `json:"structured_output_badge,omitempty"`
}

type ProviderConfig struct {
	Name         string                 `json:"name"`
	BaseURL      string                 `json:"base_url"`
	APIKey       string                 `json:"api_key"`
	Enabled      bool                   `json:"enabled"`
	AudioOnly    bool                   `json:"audio_only,omitempty"`
	ImageOnly    bool                   `json:"image_only,omitempty"`
	Favicon      string                 `json:"favicon,omitempty"`
	ModelAliases map[string]ModelAlias  `json:"model_aliases,omitempty"`
}

type ProvidersConfig struct {
	Providers []ProviderConfig `json:"providers"`
}

// OpenAI-compatible chat request/response
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// TestResult holds the result of testing a model
type TestResult struct {
	ModelAlias  string
	ModelName   string
	Provider    string
	AvgSpeedMs  int
	Success     bool
	Error       string
}

func main() {
	fmt.Println("🧪 Agent Model Testing Script")
	fmt.Println("=============================\n")

	// Read providers.json
	data, err := os.ReadFile("providers.json")
	if err != nil {
		fmt.Printf("❌ Error reading providers.json: %v\n", err)
		os.Exit(1)
	}

	var config ProvidersConfig
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("❌ Error parsing providers.json: %v\n", err)
		os.Exit(1)
	}

	// Collect models to test
	var results []TestResult

	for _, provider := range config.Providers {
		// Skip disabled providers
		if !provider.Enabled {
			continue
		}

		// Skip audio/image only providers
		if provider.AudioOnly || provider.ImageOnly {
			continue
		}

		// Skip providers without model aliases
		if len(provider.ModelAliases) == 0 {
			continue
		}

		fmt.Printf("\n📦 Testing provider: %s\n", provider.Name)
		fmt.Printf("   Base URL: %s\n", provider.BaseURL)

		for aliasName, alias := range provider.ModelAliases {
			// Only test models with agents: true
			if alias.Agents == nil || !*alias.Agents {
				fmt.Printf("   ⏭️  Skipping %s (agents: false or not set)\n", aliasName)
				continue
			}

			fmt.Printf("\n   🔬 Testing model: %s (%s)\n", alias.DisplayName, aliasName)

			result := testModel(provider, aliasName, alias)
			results = append(results, result)

			if result.Success {
				fmt.Printf("   ✅ Success! Avg speed: %dms\n", result.AvgSpeedMs)
			} else {
				fmt.Printf("   ❌ Failed: %s\n", result.Error)
			}

			// Sleep between tests to avoid rate limiting
			time.Sleep(2 * time.Second)
		}
	}

	// Print summary
	fmt.Println("\n\n📊 Test Results Summary")
	fmt.Println("========================\n")

	fmt.Printf("%-30s %-40s %-15s %s\n", "Provider", "Model", "Speed (ms)", "Status")
	fmt.Println(strings.Repeat("-", 100))

	for _, result := range results {
		status := "✅ Pass"
		speedStr := fmt.Sprintf("%d", result.AvgSpeedMs)
		if !result.Success {
			status = "❌ Fail"
			speedStr = "N/A"
		}
		fmt.Printf("%-30s %-40s %-15s %s\n",
			truncate(result.Provider, 30),
			truncate(result.ModelName, 40),
			speedStr,
			status)
	}

	// Generate update suggestions
	fmt.Println("\n\n💡 Suggested Updates for providers.json")
	fmt.Println("=========================================\n")

	for _, result := range results {
		if result.Success {
			badge := ""
			support := "good"

			if result.AvgSpeedMs < 2000 {
				badge = "FASTEST"
				support = "excellent"
			} else if result.AvgSpeedMs < 5000 {
				support = "excellent"
			} else if result.AvgSpeedMs < 10000 {
				support = "good"
			} else {
				support = "fair"
			}

			fmt.Printf("Model: %s\n", result.ModelAlias)
			fmt.Printf("  \"structured_output_speed_ms\": %d,\n", result.AvgSpeedMs)
			fmt.Printf("  \"structured_output_support\": \"%s\",\n", support)
			if badge != "" {
				fmt.Printf("  \"structured_output_badge\": \"%s\",\n", badge)
			}
			fmt.Println()
		}
	}
}

func testModel(provider ProviderConfig, aliasName string, alias ModelAlias) TestResult {
	result := TestResult{
		ModelAlias: aliasName,
		ModelName:  alias.DisplayName,
		Provider:   provider.Name,
	}

	var totalDuration int64

	for i := 0; i < numTests; i++ {
		start := time.Now()

		err := callModel(provider.BaseURL, provider.APIKey, alias.ActualModel)

		duration := time.Since(start).Milliseconds()

		if err != nil {
			result.Success = false
			result.Error = err.Error()
			return result
		}

		totalDuration += duration
		fmt.Printf("      Test %d: %dms\n", i+1, duration)
	}

	result.Success = true
	result.AvgSpeedMs = int(totalDuration / int64(numTests))

	return result
}

func callModel(baseURL, apiKey, modelName string) error {
	// Prepare request
	reqBody := ChatRequest{
		Model: modelName,
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: testPrompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	// Make HTTP request
	url := baseURL + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	if chatResp.Error != nil {
		return fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return fmt.Errorf("no response from model")
	}

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
