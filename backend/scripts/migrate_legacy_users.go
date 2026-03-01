package main

import (
	"claraverse/internal/database"
	"claraverse/internal/models"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
)

// SupabaseUser represents a user from Supabase Auth
type SupabaseUser struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// SupabaseUsersResponse represents the response from Supabase Admin API
type SupabaseUsersResponse struct {
	Users []SupabaseUser `json:"users"`
}

// MigrationStats tracks migration statistics
type MigrationStats struct {
	TotalSupabaseUsers int
	AlreadyMigrated    int
	UpdatedToLegacy    int
	CreatedAsLegacy    int
	Errors             int
}

func main() {
	// Parse command line flags
	dryRun := flag.Bool("dry-run", false, "Run in dry-run mode (no changes made)")
	flag.Parse()

	fmt.Println("==============================================")
	fmt.Println("🔄 Legacy User Migration Script v2")
	fmt.Println("==============================================")
	fmt.Println()
	if *dryRun {
		fmt.Println("🔍 DRY RUN MODE - No changes will be made")
		fmt.Println()
	}
	fmt.Println("This script will:")
	fmt.Println("  1. Fetch all users from Supabase Auth")
	fmt.Println("  2. For each user:")
	fmt.Println("     - If exists in MongoDB → update to legacy_unlimited")
	fmt.Println("     - If not exists → create new doc with legacy_unlimited")
	fmt.Println()

	// Load environment
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("⚠️  No .env file found: %v (using environment variables)", err)
	}

	// Get required environment variables
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("❌ MONGODB_URI environment variable is required")
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		log.Fatal("❌ SUPABASE_URL environment variable is required")
	}

	// Need service role key for admin API
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if supabaseServiceKey == "" {
		// Fall back to regular key (might work if it's actually the service key)
		supabaseServiceKey = os.Getenv("SUPABASE_KEY")
		if supabaseServiceKey == "" {
			log.Fatal("❌ SUPABASE_SERVICE_ROLE_KEY environment variable is required")
		}
		fmt.Println("⚠️  Using SUPABASE_KEY (ensure it's the service role key for admin access)")
	}

	fmt.Println("📋 Configuration:")
	fmt.Printf("  MongoDB URI: %s\n", maskURI(mongoURI))
	fmt.Printf("  Supabase URL: %s\n", supabaseURL)
	fmt.Println()

	// Connect to MongoDB
	fmt.Println("🔗 Connecting to MongoDB...")
	db, err := database.NewMongoDB(mongoURI)
	if err != nil {
		log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
	}
	defer db.Close(context.Background())
	fmt.Println("✅ Connected to MongoDB")
	fmt.Println()

	ctx := context.Background()

	// Fetch all users from Supabase
	fmt.Println("🔗 Fetching users from Supabase Auth...")
	supabaseUsers, err := fetchSupabaseUsers(supabaseURL, supabaseServiceKey)
	if err != nil {
		log.Fatalf("❌ Failed to fetch Supabase users: %v", err)
	}
	fmt.Printf("✅ Found %d users in Supabase Auth\n", len(supabaseUsers))
	fmt.Println()

	if len(supabaseUsers) == 0 {
		fmt.Println("✅ No users to migrate!")
		return
	}

	// Show sample of users
	fmt.Println("📋 Sample of users (first 5):")
	for i, user := range supabaseUsers {
		if i >= 5 {
			break
		}
		fmt.Printf("  - %s (Created: %s)\n", user.Email, user.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Println()

	// Process migration
	stats := MigrationStats{TotalSupabaseUsers: len(supabaseUsers)}
	usersCollection := db.Database().Collection("users")
	now := time.Now()

	fmt.Println("🔨 Processing users...")
	fmt.Println()

	for _, supabaseUser := range supabaseUsers {
		// Check if user exists in MongoDB
		var existingUser models.User
		err := usersCollection.FindOne(ctx, bson.M{"supabaseUserId": supabaseUser.ID}).Decode(&existingUser)

		if err == nil {
			// User exists in MongoDB
			if existingUser.SubscriptionTier == models.TierLegacyUnlimited {
				// Already migrated
				stats.AlreadyMigrated++
				continue
			}

			// Update existing user to legacy_unlimited
			if *dryRun {
				fmt.Printf("  [DRY RUN] Would UPDATE: %s → legacy_unlimited (current: %s)\n",
					supabaseUser.Email, existingUser.SubscriptionTier)
			} else {
				update := bson.M{
					"$set": bson.M{
						"subscriptionTier":   models.TierLegacyUnlimited,
						"subscriptionStatus": models.SubStatusActive,
						"migratedToLegacyAt": now,
					},
				}
				_, err := usersCollection.UpdateOne(ctx, bson.M{"supabaseUserId": supabaseUser.ID}, update)
				if err != nil {
					fmt.Printf("  ❌ Failed to update %s: %v\n", supabaseUser.Email, err)
					stats.Errors++
					continue
				}
				fmt.Printf("  ✅ UPDATED: %s → legacy_unlimited\n", supabaseUser.Email)
			}
			stats.UpdatedToLegacy++

		} else {
			// User doesn't exist in MongoDB - create new document
			if *dryRun {
				fmt.Printf("  [DRY RUN] Would CREATE: %s as legacy_unlimited\n", supabaseUser.Email)
			} else {
				newUser := bson.M{
					"supabaseUserId":     supabaseUser.ID,
					"email":              supabaseUser.Email,
					"createdAt":          supabaseUser.CreatedAt, // Preserve original Supabase created_at
					"lastLoginAt":        now,
					"subscriptionTier":   models.TierLegacyUnlimited,
					"subscriptionStatus": models.SubStatusActive,
					"migratedToLegacyAt": now,
					"preferences": bson.M{
						"storeBuilderChatHistory": true,
					},
				}

				_, err := usersCollection.InsertOne(ctx, newUser)
				if err != nil {
					fmt.Printf("  ❌ Failed to create %s: %v\n", supabaseUser.Email, err)
					stats.Errors++
					continue
				}
				fmt.Printf("  ✅ CREATED: %s as legacy_unlimited\n", supabaseUser.Email)
			}
			stats.CreatedAsLegacy++
		}
	}

	// Print summary
	fmt.Println()
	fmt.Println("==============================================")
	if *dryRun {
		fmt.Println("📊 DRY RUN Summary (no changes made)")
	} else {
		fmt.Println("📊 Migration Summary")
	}
	fmt.Println("==============================================")
	fmt.Printf("  Total Supabase users:     %d\n", stats.TotalSupabaseUsers)
	fmt.Printf("  Already legacy_unlimited: %d\n", stats.AlreadyMigrated)
	fmt.Printf("  Updated to legacy:        %d\n", stats.UpdatedToLegacy)
	fmt.Printf("  Created as legacy:        %d\n", stats.CreatedAsLegacy)
	fmt.Printf("  Errors:                   %d\n", stats.Errors)
	fmt.Println()

	if *dryRun {
		fmt.Println("💡 To apply changes, run without --dry-run flag:")
		fmt.Println("   go run migrate_legacy_users.go")
	} else {
		// Verification
		fmt.Println("🔍 Verification:")
		legacyCount, err := usersCollection.CountDocuments(ctx, bson.M{
			"subscriptionTier": models.TierLegacyUnlimited,
		})
		if err != nil {
			log.Printf("⚠️  Could not verify: %v", err)
		} else {
			fmt.Printf("  Total legacy_unlimited users in MongoDB: %d\n", legacyCount)
		}
	}

	fmt.Println()
	fmt.Println("==============================================")
	if *dryRun {
		fmt.Println("✅ Dry Run Complete!")
	} else {
		fmt.Println("✅ Migration Complete!")
	}
	fmt.Println("==============================================")
}

// fetchSupabaseUsers fetches all users from Supabase Auth Admin API
func fetchSupabaseUsers(supabaseURL, serviceKey string) ([]SupabaseUser, error) {
	var allUsers []SupabaseUser
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("%s/auth/v1/admin/users?page=%d&per_page=%d", supabaseURL, page, perPage)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+serviceKey)
		req.Header.Set("apikey", serviceKey)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch users: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		var response SupabaseUsersResponse
		if err := json.Unmarshal(body, &response); err != nil {
			// Try parsing as array directly (some Supabase versions return array)
			var users []SupabaseUser
			if err2 := json.Unmarshal(body, &users); err2 != nil {
				return nil, fmt.Errorf("failed to parse response: %w (original: %v)", err2, err)
			}
			allUsers = append(allUsers, users...)
			if len(users) < perPage {
				break
			}
		} else {
			allUsers = append(allUsers, response.Users...)
			if len(response.Users) < perPage {
				break
			}
		}

		page++
	}

	return allUsers, nil
}

// maskURI masks sensitive parts of a connection URI for logging
func maskURI(uri string) string {
	if len(uri) > 20 {
		return uri[:15] + "..." + uri[len(uri)-10:]
	}
	return "***"
}

// ensureIndex creates an index on supabaseUserId if it doesn't exist
func ensureIndex(ctx context.Context, collection interface{}) {
	// This is a helper that could be called to ensure fast lookups
	// For now, we assume the index exists or MongoDB handles it
}
