package preflight

import (
	"claraverse/internal/database"
	"fmt"
	"log"
	"os"
)

// CheckResult represents the result of a preflight check
type CheckResult struct {
	Name    string
	Status  string // "pass", "fail", "warning"
	Message string
	Error   error
}

// Checker performs pre-flight checks before server starts
type Checker struct {
	db             *database.DB
	requiredEnvars []string
}

// NewChecker creates a new preflight checker
func NewChecker(db *database.DB) *Checker {
	return &Checker{
		db: db,
		requiredEnvars: []string{
			// Optional: Add required environment variables here
		},
	}
}

// RunAll runs all preflight checks and returns results
func (c *Checker) RunAll() []CheckResult {
	log.Println("🔍 Running pre-flight checks...")

	results := []CheckResult{
		c.checkDatabaseConnection(),
		c.checkDatabaseSchema(),
		c.checkEnvironmentVariables(),
	}

	// Print summary
	passed := 0
	failed := 0
	warnings := 0

	for _, result := range results {
		switch result.Status {
		case "pass":
			log.Printf("   ✅ %s: %s", result.Name, result.Message)
			passed++
		case "fail":
			log.Printf("   ❌ %s: %s", result.Name, result.Message)
			if result.Error != nil {
				log.Printf("      Error: %v", result.Error)
			}
			failed++
		case "warning":
			log.Printf("   ⚠️  %s: %s", result.Name, result.Message)
			warnings++
		}
	}

	log.Printf("\n📊 Pre-flight summary: %d passed, %d failed, %d warnings\n", passed, failed, warnings)

	return results
}

// HasFailures returns true if any check failed
func HasFailures(results []CheckResult) bool {
	for _, result := range results {
		if result.Status == "fail" {
			return true
		}
	}
	return false
}

// checkDatabaseConnection verifies database connectivity
func (c *Checker) checkDatabaseConnection() CheckResult {
	if err := c.db.Ping(); err != nil {
		return CheckResult{
			Name:    "Database Connection",
			Status:  "fail",
			Message: "Cannot connect to database",
			Error:   err,
		}
	}

	return CheckResult{
		Name:    "Database Connection",
		Status:  "pass",
		Message: "Database connection successful",
	}
}

// checkDatabaseSchema verifies all required tables exist
func (c *Checker) checkDatabaseSchema() CheckResult {
	requiredTables := []string{
		"providers",
		"models",
		"provider_model_filters",
		"model_capabilities",
		"model_refresh_log",
	}

	for _, table := range requiredTables {
		var count int
		// MySQL-compatible query using INFORMATION_SCHEMA
		query := "SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?"
		err := c.db.QueryRow(query, table).Scan(&count)
		if err != nil || count == 0 {
			return CheckResult{
				Name:    "Database Schema",
				Status:  "fail",
				Message: fmt.Sprintf("Required table '%s' not found", table),
				Error:   err,
			}
		}
	}

	return CheckResult{
		Name:    "Database Schema",
		Status:  "pass",
		Message: fmt.Sprintf("All %d required tables exist", len(requiredTables)),
	}
}


// checkEnvironmentVariables verifies required environment variables are set
func (c *Checker) checkEnvironmentVariables() CheckResult {
	missing := []string{}

	for _, envar := range c.requiredEnvars {
		if os.Getenv(envar) == "" {
			missing = append(missing, envar)
		}
	}

	if len(missing) > 0 {
		return CheckResult{
			Name:    "Environment Variables",
			Status:  "warning",
			Message: fmt.Sprintf("Missing environment variables: %v", missing),
		}
	}

	// Check optional but recommended variables
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		return CheckResult{
			Name:    "Environment Variables",
			Status:  "warning",
			Message: "Supabase authentication not configured (running in development mode)",
		}
	}

	return CheckResult{
		Name:    "Environment Variables",
		Status:  "pass",
		Message: "All environment variables configured",
	}
}

// checkProviderConnectivity tests if we can reach provider APIs (optional, can be slow)
func (c *Checker) checkProviderConnectivity() CheckResult {
	// This is an optional check that could be added
	// It would test actual connectivity to provider APIs
	// For now, we'll skip it to keep startup fast

	return CheckResult{
		Name:    "Provider Connectivity",
		Status:  "pass",
		Message: "Skipped (optional check)",
	}
}

// QuickCheck runs minimal checks for fast startup
func (c *Checker) QuickCheck() []CheckResult {
	log.Println("⚡ Running quick pre-flight checks...")

	results := []CheckResult{
		c.checkDatabaseConnection(),
	}

	passed := 0
	failed := 0

	for _, result := range results {
		if result.Status == "pass" {
			log.Printf("   ✅ %s", result.Name)
			passed++
		} else if result.Status == "fail" {
			log.Printf("   ❌ %s: %s", result.Name, result.Message)
			failed++
		}
	}

	return results
}
