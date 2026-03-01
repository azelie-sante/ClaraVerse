package jobs

import (
	"claraverse/internal/database"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// OrphanExecutionCleanupJob finds executions stuck in "running" status
// (e.g., from a server crash) and marks them as "failed".
type OrphanExecutionCleanupJob struct {
	mongoDB    *database.MongoDB
	collection *mongo.Collection
	interval   time.Duration
	maxAge     time.Duration // executions running longer than this are considered orphaned
	lastRun    time.Time
}

// NewOrphanExecutionCleanupJob creates a new orphan execution cleanup job.
// interval: how often to run (e.g., 5 minutes)
// maxAge: executions running longer than this are orphaned (e.g., 15 minutes)
func NewOrphanExecutionCleanupJob(mongoDB *database.MongoDB, interval, maxAge time.Duration) *OrphanExecutionCleanupJob {
	var collection *mongo.Collection
	if mongoDB != nil {
		collection = mongoDB.Database().Collection("executions")
	}
	return &OrphanExecutionCleanupJob{
		mongoDB:    mongoDB,
		collection: collection,
		interval:   interval,
		maxAge:     maxAge,
	}
}

// Run finds and cleans up orphaned executions.
func (j *OrphanExecutionCleanupJob) Run(ctx context.Context) error {
	j.lastRun = time.Now()

	if j.collection == nil {
		log.Println("⚠️ [ORPHAN-CLEANUP] Skipped: MongoDB not available")
		return nil
	}

	cutoff := time.Now().Add(-j.maxAge)

	// Find executions that are still "running" but started before the cutoff
	filter := bson.M{
		"status": bson.M{"$in": bson.A{"running", "pending"}},
		"startedAt": bson.M{"$lt": cutoff},
	}

	update := bson.M{
		"$set": bson.M{
			"status":      "failed",
			"error":       "Execution interrupted: server restart or timeout",
			"completedAt": time.Now(),
		},
	}

	result, err := j.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Printf("❌ [ORPHAN-CLEANUP] Failed to update orphaned executions: %v", err)
		return err
	}

	if result.ModifiedCount > 0 {
		log.Printf("🧹 [ORPHAN-CLEANUP] Cleaned up %d orphaned executions (started before %s)",
			result.ModifiedCount, cutoff.Format(time.RFC3339))
	}

	return nil
}

// GetNextRunTime returns when this job should next execute.
func (j *OrphanExecutionCleanupJob) GetNextRunTime() time.Time {
	if j.lastRun.IsZero() {
		// First run: 1 minute after startup (give time for server to init)
		return time.Now().Add(1 * time.Minute)
	}
	return j.lastRun.Add(j.interval)
}
