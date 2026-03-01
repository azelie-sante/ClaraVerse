package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// This script validates that workflow generator creates compliant schemas
// Run after making test workflows to ensure they pass our own standards

func main() {
	fmt.Println("🔍 Workflow Generator Compliance Validator")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// Example workflow block from workflow generator
	exampleBlock := map[string]interface{}{
		"type": "llm_inference",
		"name": "News Fetcher",
		"config": map[string]interface{}{
			"outputFormat": "json",
			"outputSchema": map[string]interface{}{
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
							"required":             []interface{}{"title", "source", "url", "summary", "publishedDate"},
							"additionalProperties": false,
						},
					},
					"totalResults": map[string]interface{}{"type": "number"},
					"fetchedAt":    map[string]interface{}{"type": "string"},
				},
				"required":             []interface{}{"articles", "totalResults", "fetchedAt"},
				"additionalProperties": false,
			},
		},
	}

	config := exampleBlock["config"].(map[string]interface{})
	schema := config["outputSchema"].(map[string]interface{})

	issues := validateSchema(schema, "root")

	if len(issues) == 0 {
		fmt.Println("✅ Schema is VALID and compliant!")
		fmt.Println()
		fmt.Println("✓ All properties are in required arrays")
		fmt.Println("✓ additionalProperties: false is set")
		fmt.Println("✓ Nested objects have complete required arrays")
		fmt.Println()
		fmt.Println("This schema will pass OpenAI strict mode validation.")
	} else {
		fmt.Println("❌ Schema has ISSUES:")
		fmt.Println()
		for _, issue := range issues {
			fmt.Printf("  • %s\n", issue)
		}
		fmt.Println()
		fmt.Println("⚠️  This schema will FAIL OpenAI strict mode validation!")
	}

	// Pretty print schema
	fmt.Println()
	fmt.Println("📋 Schema Structure:")
	fmt.Println(strings.Repeat("-", 60))
	schemaJSON, _ := json.MarshalIndent(schema, "", "  ")
	fmt.Println(string(schemaJSON))
}

func validateSchema(schema map[string]interface{}, path string) []string {
	var issues []string

	// Check if type is object
	schemaType, hasType := schema["type"].(string)
	if !hasType {
		issues = append(issues, fmt.Sprintf("%s: Missing 'type' field", path))
		return issues
	}

	if schemaType == "object" {
		// Get properties
		properties, hasProps := schema["properties"].(map[string]interface{})
		if !hasProps {
			issues = append(issues, fmt.Sprintf("%s: Object type missing 'properties'", path))
			return issues
		}

		// Get required array
		requiredRaw, hasRequired := schema["required"]
		if !hasRequired {
			issues = append(issues, fmt.Sprintf("%s: Missing 'required' array (OpenAI strict mode requires it)", path))
		} else {
			required, ok := requiredRaw.([]interface{})
			if !ok {
				issues = append(issues, fmt.Sprintf("%s: 'required' is not an array", path))
			} else {
				// Check if all properties are in required
				requiredMap := make(map[string]bool)
				for _, r := range required {
					if rStr, ok := r.(string); ok {
						requiredMap[rStr] = true
					}
				}

				// Check each property
				for propName := range properties {
					if !requiredMap[propName] {
						issues = append(issues, fmt.Sprintf("%s.%s: Property defined but NOT in required array (OpenAI strict mode rejects this)", path, propName))
					}
				}

				// Check if all required fields exist in properties
				for _, r := range required {
					if rStr, ok := r.(string); ok {
						if _, exists := properties[rStr]; !exists {
							issues = append(issues, fmt.Sprintf("%s: Required field '%s' not in properties", path, rStr))
						}
					}
				}
			}
		}

		// Check additionalProperties
		if additionalProps, hasAdditional := schema["additionalProperties"]; hasAdditional {
			if additionalBool, ok := additionalProps.(bool); ok && additionalBool {
				issues = append(issues, fmt.Sprintf("%s: additionalProperties should be false for strict schemas", path))
			}
		} else {
			issues = append(issues, fmt.Sprintf("%s: Missing 'additionalProperties: false' (recommended for strict schemas)", path))
		}

		// Recursively validate nested objects
		for propName, propValue := range properties {
			if propSchema, ok := propValue.(map[string]interface{}); ok {
				propType, _ := propSchema["type"].(string)

				if propType == "object" {
					nestedIssues := validateSchema(propSchema, fmt.Sprintf("%s.%s", path, propName))
					issues = append(issues, nestedIssues...)
				} else if propType == "array" {
					// Check array items
					if items, hasItems := propSchema["items"].(map[string]interface{}); hasItems {
						itemType, _ := items["type"].(string)
						if itemType == "object" {
							nestedIssues := validateSchema(items, fmt.Sprintf("%s.%s[items]", path, propName))
							issues = append(issues, nestedIssues...)
						}
					}
				}
			}
		}
	}

	return issues
}
