package jobs

import (
	"claraverse/internal/database"
	"claraverse/internal/models"
	"claraverse/internal/services"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GracePeriodChecker handles expiration of grace periods for ON_HOLD subscriptions
type GracePeriodChecker struct {
	mongoDB         *database.MongoDB
	userService     *services.UserService
	tierService     *services.TierService
	gracePeriodDays int
	subscriptions   *mongo.Collection
}

// NewGracePeriodChecker creates a new grace period checker
func NewGracePeriodChecker(
	mongoDB *database.MongoDB,
	userService *services.UserService,
	tierService *services.TierService,
	gracePeriodDays int,
) *GracePeriodChecker {
	if gracePeriodDays == 0 {
		gracePeriodDays = 7 // Default: 7 day grace period
	}

	var subscriptions *mongo.Collection
	if mongoDB != nil {
		subscriptions = mongoDB.Database().Collection("subscriptions")
	}

	return &GracePeriodChecker{
		mongoDB:         mongoDB,
		userService:     userService,
		tierService:     tierService,
		gracePeriodDays: gracePeriodDays,
		subscriptions:   subscriptions,
	}
}

// Run checks for expired grace periods and downgrades subscriptions
func (g *GracePeriodChecker) Run(ctx context.Context) error {
	if g.mongoDB == nil || g.userService == nil || g.tierService == nil {
		log.Println("⚠️  [GRACE-PERIOD] Grace period checker disabled (requires MongoDB, UserService, TierService)")
		return nil
	}

	log.Println("⏰ [GRACE-PERIOD] Checking for expired grace periods...")
	startTime := time.Now()

	// Calculate cutoff date
	cutoffDate := time.Now().UTC().AddDate(0, 0, -g.gracePeriodDays)

	// Find subscriptions that are ON_HOLD and past grace period
	filter := bson.M{
		"status": models.SubStatusOnHold,
		"updatedAt": bson.M{
			"$lt": cutoffDate,
		},
	}

	cursor, err := g.subscriptions.Find(ctx, filter)
	if err != nil {
		log.Printf("❌ [GRACE-PERIOD] Failed to query subscriptions: %v", err)
		return err
	}
	defer cursor.Close(ctx)

	expiredCount := 0
	for cursor.Next(ctx) {
		var sub models.Subscription
		if err := cursor.Decode(&sub); err != nil {
			log.Printf("⚠️  [GRACE-PERIOD] Failed to decode subscription: %v", err)
			continue
		}

		if err := g.expireSubscription(ctx, &sub); err != nil {
			log.Printf("⚠️  [GRACE-PERIOD] Failed to expire subscription %s: %v", sub.ID.Hex(), err)
			continue
		}

		expiredCount++
		log.Printf("✅ [GRACE-PERIOD] Expired subscription %s for user %s (on hold for %d days)",
			sub.ID.Hex(), sub.UserID, g.gracePeriodDays)
	}

	duration := time.Since(startTime)
	log.Printf("✅ [GRACE-PERIOD] Check complete: expired %d subscriptions in %v", expiredCount, duration)

	return nil
}

// expireSubscription downgrades a subscription after grace period expires
func (g *GracePeriodChecker) expireSubscription(ctx context.Context, sub *models.Subscription) error {
	// Update subscription to cancelled
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"tier":        models.TierFree,
			"status":      models.SubStatusCancelled,
			"cancelledAt": now,
			"updatedAt":   now,
		},
	}

	_, err := g.subscriptions.UpdateOne(ctx, bson.M{"_id": sub.ID}, update)
	if err != nil {
		return err
	}

	// Update user tier to free
	if g.userService != nil {
		err = g.userService.UpdateSubscriptionWithStatus(
			ctx,
			sub.UserID,
			models.TierFree,
			models.SubStatusCancelled,
			nil,
		)
		if err != nil {
			log.Printf("⚠️  [GRACE-PERIOD] Failed to update user tier: %v", err)
			// Don't fail the job if user update fails
		}
	}

	// Invalidate tier cache
	if g.tierService != nil {
		g.tierService.InvalidateCache(sub.UserID)
	}

	return nil
}

// GetNextRunTime returns when the job should run next (hourly)
func (g *GracePeriodChecker) GetNextRunTime() time.Time {
	return time.Now().UTC().Add(1 * time.Hour)
}

// GetExpiredSubscriptions returns subscriptions past grace period (for monitoring)
func (g *GracePeriodChecker) GetExpiredSubscriptions(ctx context.Context) ([]models.Subscription, error) {
	if g.mongoDB == nil {
		return nil, nil
	}

	cutoffDate := time.Now().UTC().AddDate(0, 0, -g.gracePeriodDays)

	filter := bson.M{
		"status": models.SubStatusOnHold,
		"updatedAt": bson.M{
			"$lt": cutoffDate,
		},
	}

	cursor, err := g.subscriptions.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var subscriptions []models.Subscription
	if err := cursor.All(ctx, &subscriptions); err != nil {
		return nil, err
	}

	return subscriptions, nil
}
