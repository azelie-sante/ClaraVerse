package jobs

import (
	"claraverse/internal/database"
	"claraverse/internal/services"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// RetentionCleanupJob handles deletion of old execution data based on tier retention limits
type RetentionCleanupJob struct {
	mongoDB     *database.MongoDB
	tierService *services.TierService
	executions  *mongo.Collection
}

// NewRetentionCleanupJob creates a new retention cleanup job
func NewRetentionCleanupJob(mongoDB *database.MongoDB, tierService *services.TierService) *RetentionCleanupJob {
	var executions *mongo.Collection
	if mongoDB != nil {
		executions = mongoDB.Database().Collection("executions")
	}

	return &RetentionCleanupJob{
		mongoDB:     mongoDB,
		tierService: tierService,
		executions:  executions,
	}
}

// Run executes the retention cleanup for all users
func (j *RetentionCleanupJob) Run(ctx context.Context) error {
	if j.mongoDB == nil || j.tierService == nil {
		log.Println("[RETENTION] Retention cleanup disabled (requires MongoDB and TierService)")
		return nil
	}

	log.Println("[RETENTION] Starting execution retention cleanup...")
	startTime := time.Now()

	// Get all unique user IDs from executions collection
	userIDs, err := j.getUniqueUserIDs(ctx)
	if err != nil {
		log.Printf("[RETENTION] Failed to get user IDs: %v", err)
		return err
	}

	log.Printf("[RETENTION] Found %d users with executions", len(userIDs))

	totalDeleted := 0
	for _, userID := range userIDs {
		deleted, err := j.cleanupUserExecutions(ctx, userID)
		if err != nil {
			log.Printf("[RETENTION] Failed to cleanup executions for user %s: %v", userID, err)
			continue
		}
		if deleted > 0 {
			totalDeleted += deleted
			log.Printf("[RETENTION] Deleted %d old executions for user %s", deleted, userID)
		}
	}

	duration := time.Since(startTime)
	log.Printf("[RETENTION] Cleanup complete: deleted %d executions in %v", totalDeleted, duration)

	return nil
}

// getUniqueUserIDs returns all unique user IDs that have executions
func (j *RetentionCleanupJob) getUniqueUserIDs(ctx context.Context) ([]string, error) {
	results, err := j.executions.Distinct(ctx, "userId", bson.M{})
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(results))
	for _, result := range results {
		if userID, ok := result.(string); ok {
			userIDs = append(userIDs, userID)
		}
	}

	return userIDs, nil
}

// cleanupUserExecutions deletes old executions for a specific user based on their tier retention
func (j *RetentionCleanupJob) cleanupUserExecutions(ctx context.Context, userID string) (int, error) {
	// Get user's tier limits
	limits := j.tierService.GetLimits(ctx, userID)

	// Calculate cutoff date
	cutoffDate := time.Now().UTC().AddDate(0, 0, -limits.RetentionDays)

	filter := bson.M{
		"userId": userID,
		"createdAt": bson.M{
			"$lt": cutoffDate,
		},
	}

	result, err := j.executions.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return int(result.DeletedCount), nil
}

// GetNextRunTime returns when the job should run next (daily at 2 AM UTC)
func (j *RetentionCleanupJob) GetNextRunTime() time.Time {
	now := time.Now().UTC()

	// Schedule for 2 AM UTC
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, time.UTC)

	// If we've passed 2 AM today, schedule for tomorrow
	if now.After(nextRun) {
		nextRun = nextRun.Add(24 * time.Hour)
	}

	return nextRun
}

// GetStats returns statistics about execution retention
func (j *RetentionCleanupJob) GetStats(ctx context.Context, userID string) (*RetentionStats, error) {
	if j.mongoDB == nil || j.tierService == nil {
		return nil, nil
	}

	limits := j.tierService.GetLimits(ctx, userID)
	cutoffDate := time.Now().UTC().AddDate(0, 0, -limits.RetentionDays)

	// Count total executions
	total, err := j.executions.CountDocuments(ctx, bson.M{"userId": userID})
	if err != nil {
		return nil, err
	}

	// Count old executions (will be deleted)
	old, err := j.executions.CountDocuments(ctx, bson.M{
		"userId": userID,
		"createdAt": bson.M{
			"$lt": cutoffDate,
		},
	})
	if err != nil {
		return nil, err
	}

	return &RetentionStats{
		TotalExecutions:     int(total),
		RetainedExecutions:  int(total - old),
		DeletableExecutions: int(old),
		RetentionDays:       limits.RetentionDays,
		CutoffDate:          cutoffDate,
	}, nil
}

// RetentionStats provides statistics about execution retention
type RetentionStats struct {
	TotalExecutions     int       `json:"total_executions"`
	RetainedExecutions  int       `json:"retained_executions"`
	DeletableExecutions int       `json:"deletable_executions"`
	RetentionDays       int       `json:"retention_days"`
	CutoffDate          time.Time `json:"cutoff_date"`
}

// ExecutionRetention model for tracking deletion events (audit log)
type ExecutionRetention struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID        string             `bson:"userId" json:"user_id"`
	DeletedAt     time.Time          `bson:"deletedAt" json:"deleted_at"`
	Count         int                `bson:"count" json:"count"`
	RetentionDays int                `bson:"retentionDays" json:"retention_days"`
	CutoffDate    time.Time          `bson:"cutoffDate" json:"cutoff_date"`
}
