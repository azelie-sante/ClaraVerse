package tests

import (
	"claraverse/internal/handlers"
	"claraverse/internal/models"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// ============================================================================
// Test: Production User (Main Branch) Before Migration
// ============================================================================
// This test verifies what happens when a user from production (main branch)
// without any subscription fields logs in. They should get "free" tier.

func TestE2E_ProductionUserLoginBeforeMigration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup with promo disabled to avoid interference
	ts := SetupE2ETestWithMongoDB(t, &PromoTestConfig{
		PromoEnabled: false,
	})
	if ts == nil {
		return
	}
	defer ts.Cleanup()

	ctx := context.Background()
	userID := "test-e2e-prod-before-" + time.Now().Format("20060102150405")
	email := "test-e2e-prod-before@example.com"

	// Create user with ONLY main branch fields (simulating production user)
	// Main branch users have: _id, supabaseUserId, email, createdAt, lastLoginAt, preferences
	// NO subscriptionTier, subscriptionStatus, or other new fields
	collection := ts.MongoDB.Database().Collection("users")
	_, err := collection.InsertOne(ctx, bson.M{
		"supabaseUserId": userID,
		"email":          email,
		"createdAt":      time.Now().Add(-90 * 24 * time.Hour), // Created 90 days ago
		"lastLoginAt":    time.Now().Add(-24 * time.Hour),
		"preferences": bson.M{
			"storeBuilderChatHistory": true,
			"chatPrivacyMode":         "cloud",
		},
		// Explicitly NOT setting: subscriptionTier, subscriptionStatus, etc.
	})
	if err != nil {
		t.Fatalf("Failed to create production-like user: %v", err)
	}

	// Verify user was created without subscription fields
	var createdUser bson.M
	err = collection.FindOne(ctx, bson.M{"supabaseUserId": userID}).Decode(&createdUser)
	if err != nil {
		t.Fatalf("Failed to fetch created user: %v", err)
	}
	if _, exists := createdUser["subscriptionTier"]; exists {
		t.Fatal("User should NOT have subscriptionTier field (simulating main branch)")
	}

	// Create app with auth middleware for this user
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(testAuthMiddleware(userID, email))

	subHandler := handlers.NewSubscriptionHandler(ts.PaymentService, ts.UserService)
	app.Get("/api/subscriptions/current", subHandler.GetCurrent)

	// Call GET /api/subscriptions/current
	req := httptest.NewRequest("GET", "/api/subscriptions/current", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Without migration, TierService defaults to "free" for users without tier
	tier := ts.TierService.GetUserTier(ctx, userID)
	if tier != models.TierFree {
		t.Errorf("Expected TierService to return 'free' for unmigrated user, got '%s'", tier)
	}

	t.Logf("Response tier: %v", result["tier"])
	t.Logf("TierService tier: %s", tier)
	t.Log("Confirmed: Production users without subscription fields get 'free' tier until migrated")
}

// ============================================================================
// Test: Production User Migration to Legacy Unlimited
// ============================================================================
// This test simulates the migration script logic and verifies it works correctly.

func TestE2E_ProductionUserMigration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ts := SetupE2ETestWithMongoDB(t, &PromoTestConfig{
		PromoEnabled: false,
	})
	if ts == nil {
		return
	}
	defer ts.Cleanup()

	ctx := context.Background()
	userID := "test-e2e-migrate-" + time.Now().Format("20060102150405")
	email := "test-e2e-migrate@example.com"

	collection := ts.MongoDB.Database().Collection("users")

	// Step 1: Create production-like user (no subscription fields)
	_, err := collection.InsertOne(ctx, bson.M{
		"supabaseUserId": userID,
		"email":          email,
		"createdAt":      time.Now().Add(-90 * 24 * time.Hour),
		"lastLoginAt":    time.Now().Add(-24 * time.Hour),
		"preferences": bson.M{
			"storeBuilderChatHistory": true,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create production-like user: %v", err)
	}

	// Step 2: Apply migration logic (same as migrate_legacy_users.go)
	migrationFilter := bson.M{
		"$or": []bson.M{
			{"subscriptionTier": bson.M{"$exists": false}},
			{"subscriptionTier": ""},
		},
	}
	migrationTime := time.Now()
	migrationUpdate := bson.M{
		"$set": bson.M{
			"subscriptionTier":   models.TierLegacyUnlimited,
			"subscriptionStatus": models.SubStatusActive,
			"migratedToLegacyAt": migrationTime,
		},
	}

	result, err := collection.UpdateMany(ctx, migrationFilter, migrationUpdate)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}
	if result.ModifiedCount == 0 {
		t.Fatal("Expected at least 1 user to be migrated")
	}
	t.Logf("Migrated %d user(s)", result.ModifiedCount)

	// Step 3: Verify user now has legacy_unlimited tier
	var migratedUser models.User
	err = collection.FindOne(ctx, bson.M{"supabaseUserId": userID}).Decode(&migratedUser)
	if err != nil {
		t.Fatalf("Failed to fetch migrated user: %v", err)
	}

	if migratedUser.SubscriptionTier != models.TierLegacyUnlimited {
		t.Errorf("Expected tier '%s', got '%s'", models.TierLegacyUnlimited, migratedUser.SubscriptionTier)
	}
	if migratedUser.SubscriptionStatus != models.SubStatusActive {
		t.Errorf("Expected status '%s', got '%s'", models.SubStatusActive, migratedUser.SubscriptionStatus)
	}
	if migratedUser.MigratedToLegacyAt == nil {
		t.Error("Expected migratedToLegacyAt to be set")
	}

	// Step 4: Verify TierService returns correct tier
	// Invalidate cache first since we updated directly in DB
	ts.TierService.InvalidateCache(userID)
	tier := ts.TierService.GetUserTier(ctx, userID)
	if tier != models.TierLegacyUnlimited {
		t.Errorf("Expected TierService to return '%s', got '%s'", models.TierLegacyUnlimited, tier)
	}

	// Step 5: Verify limits are unlimited
	limits := models.GetTierLimits(tier)
	if limits.MaxSchedules != -1 {
		t.Errorf("Expected unlimited schedules (-1), got %d", limits.MaxSchedules)
	}
	if limits.MaxAPIKeys != -1 {
		t.Errorf("Expected unlimited API keys (-1), got %d", limits.MaxAPIKeys)
	}

	t.Log("Migration test passed: Production user successfully migrated to legacy_unlimited")
}

// ============================================================================
// Test: Production User Login After Migration
// ============================================================================
// This tests the full flow: production user already migrated, then logs in.

func TestE2E_ProductionUserLoginAfterMigration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup with ACTIVE promo (to verify legacy users are NOT affected by promo)
	now := time.Now()
	ts := SetupE2ETestWithMongoDB(t, &PromoTestConfig{
		PromoEnabled:   true,
		PromoStartDate: now.Add(-1 * time.Hour),
		PromoEndDate:   now.Add(24 * time.Hour),
		PromoDuration:  30,
	})
	if ts == nil {
		return
	}
	defer ts.Cleanup()

	ctx := context.Background()
	userID := "test-e2e-post-migrate-" + time.Now().Format("20060102150405")
	email := "test-e2e-post-migrate@example.com"

	collection := ts.MongoDB.Database().Collection("users")

	// Create user that has ALREADY been migrated (simulating post-migration state)
	migratedAt := time.Now().Add(-7 * 24 * time.Hour) // Migrated 7 days ago
	_, err := collection.InsertOne(ctx, bson.M{
		"supabaseUserId":     userID,
		"email":              email,
		"createdAt":          time.Now().Add(-120 * 24 * time.Hour), // Created 120 days ago
		"lastLoginAt":        time.Now().Add(-24 * time.Hour),
		"subscriptionTier":   models.TierLegacyUnlimited,
		"subscriptionStatus": models.SubStatusActive,
		"migratedToLegacyAt": migratedAt,
		"preferences": bson.M{
			"storeBuilderChatHistory": true,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create migrated user: %v", err)
	}

	// Create app with auth middleware
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(testAuthMiddleware(userID, email))

	subHandler := handlers.NewSubscriptionHandler(ts.PaymentService, ts.UserService)
	app.Get("/api/subscriptions/current", subHandler.GetCurrent)

	// Call GET /api/subscriptions/current
	req := httptest.NewRequest("GET", "/api/subscriptions/current", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify tier is legacy_unlimited (NOT converted to promo even though promo is active)
	if result["tier"] != models.TierLegacyUnlimited {
		t.Errorf("Expected tier '%s', got '%v'", models.TierLegacyUnlimited, result["tier"])
	}

	// Verify is_promo_user is false
	if result["is_promo_user"] != false {
		t.Errorf("Expected is_promo_user false for legacy user, got '%v'", result["is_promo_user"])
	}

	// Verify via TierService
	tier := ts.TierService.GetUserTier(ctx, userID)
	if tier != models.TierLegacyUnlimited {
		t.Errorf("TierService should return '%s', got '%s'", models.TierLegacyUnlimited, tier)
	}

	t.Log("Post-migration login test passed: Legacy user retains legacy_unlimited tier")
}

// ============================================================================
// Test: Migration Idempotency
// ============================================================================
// Running migration twice should not break anything or update migratedToLegacyAt.

func TestE2E_MigrationIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ts := SetupE2ETestWithMongoDB(t, &PromoTestConfig{
		PromoEnabled: false,
	})
	if ts == nil {
		return
	}
	defer ts.Cleanup()

	ctx := context.Background()
	userID := "test-e2e-idempotent-" + time.Now().Format("20060102150405")
	email := "test-e2e-idempotent@example.com"

	collection := ts.MongoDB.Database().Collection("users")

	// Create production-like user
	_, err := collection.InsertOne(ctx, bson.M{
		"supabaseUserId": userID,
		"email":          email,
		"createdAt":      time.Now().Add(-60 * 24 * time.Hour),
		"lastLoginAt":    time.Now().Add(-24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// First migration
	migrationFilter := bson.M{
		"$or": []bson.M{
			{"subscriptionTier": bson.M{"$exists": false}},
			{"subscriptionTier": ""},
		},
	}
	firstMigrationTime := time.Now()
	migrationUpdate := bson.M{
		"$set": bson.M{
			"subscriptionTier":   models.TierLegacyUnlimited,
			"subscriptionStatus": models.SubStatusActive,
			"migratedToLegacyAt": firstMigrationTime,
		},
	}

	result1, err := collection.UpdateMany(ctx, migrationFilter, migrationUpdate)
	if err != nil {
		t.Fatalf("First migration failed: %v", err)
	}
	if result1.ModifiedCount != 1 {
		t.Fatalf("Expected 1 user migrated on first run, got %d", result1.ModifiedCount)
	}

	// Get migratedToLegacyAt after first migration
	var userAfterFirst models.User
	err = collection.FindOne(ctx, bson.M{"supabaseUserId": userID}).Decode(&userAfterFirst)
	if err != nil {
		t.Fatalf("Failed to fetch user after first migration: %v", err)
	}
	firstMigratedAt := userAfterFirst.MigratedToLegacyAt

	// Wait a moment to ensure time difference would be detectable
	time.Sleep(10 * time.Millisecond)

	// Second migration (should not match the user since they already have a tier)
	result2, err := collection.UpdateMany(ctx, migrationFilter, bson.M{
		"$set": bson.M{
			"subscriptionTier":   models.TierLegacyUnlimited,
			"subscriptionStatus": models.SubStatusActive,
			"migratedToLegacyAt": time.Now(), // Different time
		},
	})
	if err != nil {
		t.Fatalf("Second migration failed: %v", err)
	}
	if result2.ModifiedCount != 0 {
		t.Errorf("Expected 0 users migrated on second run (already migrated), got %d", result2.ModifiedCount)
	}

	// Verify migratedToLegacyAt unchanged
	var userAfterSecond models.User
	err = collection.FindOne(ctx, bson.M{"supabaseUserId": userID}).Decode(&userAfterSecond)
	if err != nil {
		t.Fatalf("Failed to fetch user after second migration: %v", err)
	}

	if userAfterSecond.MigratedToLegacyAt == nil || firstMigratedAt == nil {
		t.Error("migratedToLegacyAt should be set")
	} else if !userAfterSecond.MigratedToLegacyAt.Equal(*firstMigratedAt) {
		t.Errorf("migratedToLegacyAt should not change: first=%v, second=%v",
			firstMigratedAt, userAfterSecond.MigratedToLegacyAt)
	}

	// Verify user still has correct tier
	if userAfterSecond.SubscriptionTier != models.TierLegacyUnlimited {
		t.Errorf("Expected tier '%s', got '%s'", models.TierLegacyUnlimited, userAfterSecond.SubscriptionTier)
	}

	t.Log("Idempotency test passed: Running migration twice does not affect already-migrated users")
}

// ============================================================================
// Test: Mixed Users Migration (Only Unmigrated Users Affected)
// ============================================================================
// Verifies that users with existing tiers are NOT affected by migration.

func TestE2E_MixedUsersMigration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ts := SetupE2ETestWithMongoDB(t, &PromoTestConfig{
		PromoEnabled: false,
	})
	if ts == nil {
		return
	}
	defer ts.Cleanup()

	ctx := context.Background()
	collection := ts.MongoDB.Database().Collection("users")
	timestamp := time.Now().Format("20060102150405")

	// Create 4 different user types:

	// 1. Production user (no tier) - SHOULD be migrated
	prodUserID := "test-e2e-mixed-prod-" + timestamp
	_, err := collection.InsertOne(ctx, bson.M{
		"supabaseUserId": prodUserID,
		"email":          "prod@example.com",
		"createdAt":      time.Now().Add(-90 * 24 * time.Hour),
		"lastLoginAt":    time.Now().Add(-24 * time.Hour),
		// No subscriptionTier
	})
	if err != nil {
		t.Fatalf("Failed to create prod user: %v", err)
	}

	// 2. Free tier user (has tier set) - should NOT be migrated
	freeUserID := "test-e2e-mixed-free-" + timestamp
	_, err = collection.InsertOne(ctx, bson.M{
		"supabaseUserId":     freeUserID,
		"email":              "free@example.com",
		"createdAt":          time.Now().Add(-30 * 24 * time.Hour),
		"lastLoginAt":        time.Now().Add(-1 * time.Hour),
		"subscriptionTier":   models.TierFree,
		"subscriptionStatus": models.SubStatusActive,
	})
	if err != nil {
		t.Fatalf("Failed to create free user: %v", err)
	}

	// 3. Pro user (paid) - should NOT be migrated
	proUserID := "test-e2e-mixed-pro-" + timestamp
	_, err = collection.InsertOne(ctx, bson.M{
		"supabaseUserId":     proUserID,
		"email":              "pro@example.com",
		"createdAt":          time.Now().Add(-60 * 24 * time.Hour),
		"lastLoginAt":        time.Now().Add(-2 * time.Hour),
		"subscriptionTier":   models.TierPro,
		"subscriptionStatus": models.SubStatusActive,
		"dodoCustomerId":     "cust_12345",
	})
	if err != nil {
		t.Fatalf("Failed to create pro user: %v", err)
	}

	// 4. User with empty tier (edge case) - SHOULD be migrated
	emptyTierUserID := "test-e2e-mixed-empty-" + timestamp
	_, err = collection.InsertOne(ctx, bson.M{
		"supabaseUserId":   emptyTierUserID,
		"email":            "empty@example.com",
		"createdAt":        time.Now().Add(-45 * 24 * time.Hour),
		"lastLoginAt":      time.Now().Add(-12 * time.Hour),
		"subscriptionTier": "", // Empty string
	})
	if err != nil {
		t.Fatalf("Failed to create empty tier user: %v", err)
	}

	// Run migration
	migrationFilter := bson.M{
		"$or": []bson.M{
			{"subscriptionTier": bson.M{"$exists": false}},
			{"subscriptionTier": ""},
		},
	}
	migrationUpdate := bson.M{
		"$set": bson.M{
			"subscriptionTier":   models.TierLegacyUnlimited,
			"subscriptionStatus": models.SubStatusActive,
			"migratedToLegacyAt": time.Now(),
		},
	}

	result, err := collection.UpdateMany(ctx, migrationFilter, migrationUpdate)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Should have migrated exactly 2 users (prodUserID and emptyTierUserID)
	if result.ModifiedCount != 2 {
		t.Errorf("Expected 2 users migrated, got %d", result.ModifiedCount)
	}

	// Verify prod user is migrated
	var prodUser models.User
	collection.FindOne(ctx, bson.M{"supabaseUserId": prodUserID}).Decode(&prodUser)
	if prodUser.SubscriptionTier != models.TierLegacyUnlimited {
		t.Errorf("Prod user should have tier '%s', got '%s'", models.TierLegacyUnlimited, prodUser.SubscriptionTier)
	}

	// Verify empty tier user is migrated
	var emptyUser models.User
	collection.FindOne(ctx, bson.M{"supabaseUserId": emptyTierUserID}).Decode(&emptyUser)
	if emptyUser.SubscriptionTier != models.TierLegacyUnlimited {
		t.Errorf("Empty tier user should have tier '%s', got '%s'", models.TierLegacyUnlimited, emptyUser.SubscriptionTier)
	}

	// Verify free user is NOT migrated
	var freeUser models.User
	collection.FindOne(ctx, bson.M{"supabaseUserId": freeUserID}).Decode(&freeUser)
	if freeUser.SubscriptionTier != models.TierFree {
		t.Errorf("Free user should still have tier '%s', got '%s'", models.TierFree, freeUser.SubscriptionTier)
	}

	// Verify pro user is NOT migrated
	var proUser models.User
	collection.FindOne(ctx, bson.M{"supabaseUserId": proUserID}).Decode(&proUser)
	if proUser.SubscriptionTier != models.TierPro {
		t.Errorf("Pro user should still have tier '%s', got '%s'", models.TierPro, proUser.SubscriptionTier)
	}

	t.Log("Mixed users test passed: Only production users (no tier) are migrated")
}
