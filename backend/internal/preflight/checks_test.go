package preflight

import (
	"claraverse/internal/database"
	"os"
	"testing"
)

func setupPreflightTest(t *testing.T) (*database.DB, string, func()) {
	tmpDB := "test_preflight.db"
	tmpProviders := "test_providers.json"

	db, err := database.New(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	if err := db.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(tmpDB)
		os.Remove(tmpProviders)
	}

	return db, tmpProviders, cleanup
}

func TestNewChecker(t *testing.T) {
	db, providersPath, cleanup := setupPreflightTest(t)
	defer cleanup()

	checker := NewChecker(db, providersPath)
	if checker == nil {
		t.Fatal("Expected non-nil checker")
	}

	if checker.db != db {
		t.Error("Checker database not set correctly")
	}

	if checker.providersPath != providersPath {
		t.Error("Checker providers path not set correctly")
	}
}

func TestCheckDatabaseConnection_Success(t *testing.T) {
	db, providersPath, cleanup := setupPreflightTest(t)
	defer cleanup()

	checker := NewChecker(db, providersPath)
	result := checker.checkDatabaseConnection()

	if result.Status != "pass" {
		t.Errorf("Expected status 'pass', got '%s'", result.Status)
	}

	if result.Name != "Database Connection" {
		t.Errorf("Expected name 'Database Connection', got '%s'", result.Name)
	}
}

func TestCheckDatabaseConnection_Failure(t *testing.T) {
	db, providersPath, cleanup := setupPreflightTest(t)
	cleanup() // Close database immediately to simulate failure

	checker := NewChecker(db, providersPath)
	result := checker.checkDatabaseConnection()

	if result.Status != "fail" {
		t.Errorf("Expected status 'fail', got '%s'", result.Status)
	}

	if result.Error == nil {
		t.Error("Expected error to be set")
	}
}

func TestCheckDatabaseSchema_Success(t *testing.T) {
	db, providersPath, cleanup := setupPreflightTest(t)
	defer cleanup()

	checker := NewChecker(db, providersPath)
	result := checker.checkDatabaseSchema()

	if result.Status != "pass" {
		t.Errorf("Expected status 'pass', got '%s': %s", result.Status, result.Message)
	}
}

func TestCheckDatabaseSchema_MissingTable(t *testing.T) {
	tmpDB := "test_preflight_incomplete.db"
	providersPath := "test_providers.json"

	db, err := database.New(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(tmpDB)
	}()

	// Don't initialize - tables won't exist

	checker := NewChecker(db, providersPath)
	result := checker.checkDatabaseSchema()

	if result.Status != "fail" {
		t.Errorf("Expected status 'fail', got '%s'", result.Status)
	}

	if result.Error == nil {
		t.Error("Expected error to be set")
	}
}

func TestCheckProvidersFile_Exists(t *testing.T) {
	db, providersPath, cleanup := setupPreflightTest(t)
	defer cleanup()

	// Create providers file
	content := `{"providers": []}`
	if err := os.WriteFile(providersPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test providers file: %v", err)
	}

	checker := NewChecker(db, providersPath)
	result := checker.checkProvidersFile()

	if result.Status != "pass" {
		t.Errorf("Expected status 'pass', got '%s'", result.Status)
	}
}

func TestCheckProvidersFile_Missing(t *testing.T) {
	db, providersPath, cleanup := setupPreflightTest(t)
	defer cleanup()

	// Don't create the file

	checker := NewChecker(db, providersPath)
	result := checker.checkProvidersFile()

	if result.Status != "warning" {
		t.Errorf("Expected status 'warning', got '%s'", result.Status)
	}
}

func TestCheckProvidersJSON_Valid(t *testing.T) {
	db, providersPath, cleanup := setupPreflightTest(t)
	defer cleanup()

	// Create valid providers file
	content := `{
		"providers": [
			{
				"name": "OpenAI",
				"base_url": "https://api.openai.com/v1",
				"api_key": "test-key",
				"enabled": true
			}
		]
	}`
	if err := os.WriteFile(providersPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test providers file: %v", err)
	}

	checker := NewChecker(db, providersPath)
	result := checker.checkProvidersJSON()

	if result.Status != "pass" {
		t.Errorf("Expected status 'pass', got '%s': %s", result.Status, result.Message)
	}
}

func TestCheckProvidersJSON_InvalidJSON(t *testing.T) {
	db, providersPath, cleanup := setupPreflightTest(t)
	defer cleanup()

	// Create invalid JSON
	content := `{invalid json}`
	if err := os.WriteFile(providersPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test providers file: %v", err)
	}

	checker := NewChecker(db, providersPath)
	result := checker.checkProvidersJSON()

	if result.Status != "fail" {
		t.Errorf("Expected status 'fail', got '%s'", result.Status)
	}

	if result.Error == nil {
		t.Error("Expected error to be set")
	}
}

func TestCheckProvidersJSON_EmptyProviders(t *testing.T) {
	db, providersPath, cleanup := setupPreflightTest(t)
	defer cleanup()

	// Create file with no providers
	content := `{"providers": []}`
	if err := os.WriteFile(providersPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test providers file: %v", err)
	}

	checker := NewChecker(db, providersPath)
	result := checker.checkProvidersJSON()

	if result.Status != "warning" {
		t.Errorf("Expected status 'warning', got '%s'", result.Status)
	}
}

func TestCheckProvidersJSON_MissingName(t *testing.T) {
	db, providersPath, cleanup := setupPreflightTest(t)
	defer cleanup()

	// Create provider without name
	content := `{
		"providers": [
			{
				"base_url": "https://api.test.com/v1",
				"api_key": "test-key",
				"enabled": true
			}
		]
	}`
	if err := os.WriteFile(providersPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test providers file: %v", err)
	}

	checker := NewChecker(db, providersPath)
	result := checker.checkProvidersJSON()

	if result.Status != "fail" {
		t.Errorf("Expected status 'fail', got '%s'", result.Status)
	}
}

func TestCheckProvidersJSON_MissingBaseURL(t *testing.T) {
	db, providersPath, cleanup := setupPreflightTest(t)
	defer cleanup()

	// Create provider without base_url
	content := `{
		"providers": [
			{
				"name": "Test Provider",
				"api_key": "test-key",
				"enabled": true
			}
		]
	}`
	if err := os.WriteFile(providersPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test providers file: %v", err)
	}

	checker := NewChecker(db, providersPath)
	result := checker.checkProvidersJSON()

	if result.Status != "fail" {
		t.Errorf("Expected status 'fail', got '%s'", result.Status)
	}
}

func TestCheckProvidersJSON_MissingAPIKey(t *testing.T) {
	db, providersPath, cleanup := setupPreflightTest(t)
	defer cleanup()

	// Create provider without api_key
	content := `{
		"providers": [
			{
				"name": "Test Provider",
				"base_url": "https://api.test.com/v1",
				"enabled": true
			}
		]
	}`
	if err := os.WriteFile(providersPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test providers file: %v", err)
	}

	checker := NewChecker(db, providersPath)
	result := checker.checkProvidersJSON()

	// Missing API key should be a warning, not a failure
	if result.Status != "warning" {
		t.Errorf("Expected status 'warning', got '%s'", result.Status)
	}
}

func TestCheckEnvironmentVariables(t *testing.T) {
	db, providersPath, cleanup := setupPreflightTest(t)
	defer cleanup()

	checker := NewChecker(db, providersPath)
	result := checker.checkEnvironmentVariables()

	// Should pass or warn, but not fail
	if result.Status == "fail" {
		t.Errorf("Expected status 'pass' or 'warning', got 'fail': %s", result.Message)
	}
}

func TestRunAll(t *testing.T) {
	db, providersPath, cleanup := setupPreflightTest(t)
	defer cleanup()

	// Create valid providers file
	content := `{
		"providers": [
			{
				"name": "OpenAI",
				"base_url": "https://api.openai.com/v1",
				"api_key": "test-key",
				"enabled": true
			}
		]
	}`
	if err := os.WriteFile(providersPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test providers file: %v", err)
	}

	checker := NewChecker(db, providersPath)
	results := checker.RunAll()

	if len(results) == 0 {
		t.Error("Expected results, got empty slice")
	}

	// Verify all expected checks ran
	expectedChecks := map[string]bool{
		"Database Connection":     false,
		"Database Schema":         false,
		"Providers File":          false,
		"Providers JSON":          false,
		"Environment Variables":   false,
	}

	for _, result := range results {
		if _, exists := expectedChecks[result.Name]; exists {
			expectedChecks[result.Name] = true
		}
	}

	for checkName, ran := range expectedChecks {
		if !ran {
			t.Errorf("Expected check '%s' to run", checkName)
		}
	}
}

func TestHasFailures(t *testing.T) {
	// Test with no failures
	results := []CheckResult{
		{Status: "pass"},
		{Status: "pass"},
		{Status: "warning"},
	}

	if HasFailures(results) {
		t.Error("Expected no failures")
	}

	// Test with failures
	results = append(results, CheckResult{Status: "fail"})

	if !HasFailures(results) {
		t.Error("Expected failures to be detected")
	}
}

func TestQuickCheck(t *testing.T) {
	db, providersPath, cleanup := setupPreflightTest(t)
	defer cleanup()

	// Create providers file
	content := `{"providers": []}`
	if err := os.WriteFile(providersPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test providers file: %v", err)
	}

	checker := NewChecker(db, providersPath)
	results := checker.QuickCheck()

	if len(results) == 0 {
		t.Error("Expected results from quick check")
	}

	// Quick check should run fewer checks than full check
	fullResults := checker.RunAll()
	if len(results) >= len(fullResults) {
		t.Error("Expected quick check to run fewer checks than full check")
	}
}
