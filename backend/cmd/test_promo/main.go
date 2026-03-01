package main

import (
	"claraverse/internal/config"
	"claraverse/internal/database"
	"claraverse/internal/services"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("==============================================")
	fmt.Println("🎁 Promotional Campaign Verification Test")
	fmt.Println("==============================================")
	fmt.Println()

	// Load .env file
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("⚠️  No .env file found: %v", err)
	}

	// Load configuration
	cfg := config.Load()

	fmt.Println("📋 Current Promo Configuration:")
	fmt.Printf("  Enabled: %v\n", cfg.PromoEnabled)
	fmt.Printf("  Start Date: %s\n", cfg.PromoStartDate.Format(time.RFC3339))
	fmt.Printf("  End Date: %s\n", cfg.PromoEndDate.Format(time.RFC3339))
	fmt.Printf("  Duration: %d days\n", cfg.PromoDuration)
	fmt.Println()

	// Check if we're in the promo window
	now := time.Now()
	inPromoWindow := !now.Before(cfg.PromoStartDate) && now.Before(cfg.PromoEndDate)

	fmt.Println("⏰ Current Time Check:")
	fmt.Printf("  Current Time (UTC): %s\n", now.UTC().Format(time.RFC3339))
	fmt.Printf("  In Promo Window: %v\n", inPromoWindow)
	if inPromoWindow {
		fmt.Println("  ✅ New users should receive Pro tier!")
	} else {
		fmt.Println("  ❌ New users should receive Free tier (outside promo window)")
	}
	fmt.Println()

	// Connect to MongoDB
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("❌ MONGODB_URI not set")
	}

	fmt.Println("🔗 Connecting to MongoDB...")
	db, err := database.NewMongoDB(mongoURI)
	if err != nil {
		log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
	}
	defer db.Close(context.Background())
	fmt.Println("✅ Connected to MongoDB")
	fmt.Println()

	// Create UserService (nil usageLimiter for test - not needed here)
	userService := services.NewUserService(db, cfg, nil)

	// Create a test user
	ctx := context.Background()
	testUserID := fmt.Sprintf("test-promo-verify-%d", time.Now().Unix())
	testEmail := fmt.Sprintf("test-promo-%d@example.com", time.Now().Unix())

	fmt.Println("🔨 Creating test user...")
	fmt.Printf("  User ID: %s\n", testUserID)
	fmt.Printf("  Email: %s\n", testEmail)

	user, err := userService.SyncUserFromSupabase(ctx, testUserID, testEmail)
	if err != nil {
		log.Fatalf("❌ Failed to create user: %v", err)
	}

	fmt.Println("✅ User created successfully!")
	fmt.Println()

	// Verify subscription details
	fmt.Println("📊 User Subscription Details:")
	fmt.Printf("  Tier: %s\n", user.SubscriptionTier)
	fmt.Printf("  Status: %s\n", user.SubscriptionStatus)
	if user.SubscriptionExpiresAt != nil {
		fmt.Printf("  Expires At: %s\n", user.SubscriptionExpiresAt.Format(time.RFC3339))
		daysUntilExpiry := int(user.SubscriptionExpiresAt.Sub(user.CreatedAt).Hours() / 24)
		fmt.Printf("  Duration: ~%d days\n", daysUntilExpiry)
	} else {
		fmt.Println("  Expires At: Never (no expiration)")
	}
	fmt.Println()

	// Verify results
	fmt.Println("🧪 Verification Results:")

	expectedTier := "free"
	if inPromoWindow {
		expectedTier = "pro"
	}

	if user.SubscriptionTier == expectedTier {
		fmt.Printf("  ✅ Tier is correct: %s\n", user.SubscriptionTier)
	} else {
		fmt.Printf("  ❌ Tier is INCORRECT! Expected: %s, Got: %s\n", expectedTier, user.SubscriptionTier)
	}

	if inPromoWindow {
		if user.SubscriptionExpiresAt != nil {
			expectedExpiration := user.CreatedAt.Add(time.Duration(cfg.PromoDuration) * 24 * time.Hour)
			delta := user.SubscriptionExpiresAt.Sub(expectedExpiration)
			if delta < 2*time.Second && delta > -2*time.Second {
				fmt.Printf("  ✅ Expiration date is correct: %s\n", user.SubscriptionExpiresAt.Format(time.RFC3339))
			} else {
				fmt.Printf("  ⚠️  Expiration date might be incorrect. Expected: %s, Got: %s\n",
					expectedExpiration.Format(time.RFC3339),
					user.SubscriptionExpiresAt.Format(time.RFC3339))
			}
		} else {
			fmt.Println("  ❌ Expiration date is missing (should be set for promo users)")
		}
	}

	fmt.Println()

	// Cleanup
	fmt.Println("🧹 Cleaning up test user...")
	if err := userService.DeleteUser(ctx, testUserID); err != nil {
		log.Printf("⚠️  Failed to delete test user: %v", err)
	} else {
		fmt.Println("✅ Test user deleted")
	}
	fmt.Println()

	// Summary
	fmt.Println("==============================================")
	fmt.Println("📝 Summary")
	fmt.Println("==============================================")
	if inPromoWindow && user.SubscriptionTier == "pro" && user.SubscriptionExpiresAt != nil {
		fmt.Println("✅ PROMO IS WORKING!")
		fmt.Println("   New users are receiving Pro tier with expiration.")
		fmt.Printf("   They will have Pro access for %d days.\n", cfg.PromoDuration)
	} else if !inPromoWindow && user.SubscriptionTier == "free" {
		fmt.Println("✅ PROMO LOGIC IS CORRECT!")
		fmt.Println("   New users are receiving Free tier (outside promo window).")
	} else {
		fmt.Println("❌ PROMO MAY NOT BE WORKING CORRECTLY!")
		fmt.Println("   Please check the configuration and code.")
	}
	fmt.Println("==============================================")
}
