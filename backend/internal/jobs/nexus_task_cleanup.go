package jobs

import (
	"claraverse/internal/database"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// NexusTaskCleanupJob finds Nexus tasks and daemons stuck in "executing" status
// (e.g., from a daemon goroutine hang or context leak) and marks them as failed.
// This runs periodically while the server is up, complementing the boot-time
// CleanupStaleDaemons/ClearAllActive calls.
type NexusTaskCleanupJob struct {
	mongoDB        *database.MongoDB
	taskCollection *mongo.Collection
	daemonColl     *mongo.Collection
	interval       time.Duration
	maxAge         time.Duration // tasks executing longer than this are considered orphaned
	lastRun        time.Time
}

// NewNexusTaskCleanupJob creates a new Nexus task cleanup job.
// interval: how often to run (e.g., 5 minutes)
// maxAge: tasks executing longer than this are orphaned (e.g., 15 minutes)
func NewNexusTaskCleanupJob(mongoDB *database.MongoDB, interval, maxAge time.Duration) *NexusTaskCleanupJob {
	var taskColl, daemonColl *mongo.Collection
	if mongoDB != nil {
		taskColl = mongoDB.Database().Collection(database.CollectionNexusTasks)
		daemonColl = mongoDB.Database().Collection(database.CollectionNexusDaemons)
	}
	return &NexusTaskCleanupJob{
		mongoDB:        mongoDB,
		taskCollection: taskColl,
		daemonColl:     daemonColl,
		interval:       interval,
		maxAge:         maxAge,
	}
}

// Run finds and cleans up orphaned Nexus tasks and daemons.
func (j *NexusTaskCleanupJob) Run(ctx context.Context) error {
	j.lastRun = time.Now()

	if j.taskCollection == nil {
		return nil
	}

	cutoff := time.Now().Add(-j.maxAge)

	// Clean up stuck tasks (executing/waiting_input that started before cutoff)
	taskFilter := bson.M{
		"status":    bson.M{"$in": bson.A{"executing", "waiting_input"}},
		"startedAt": bson.M{"$lt": cutoff},
	}
	taskUpdate := bson.M{
		"$set": bson.M{
			"status":      "failed",
			"error":       "Task timed out — stuck in executing state",
			"completedAt": time.Now(),
			"updatedAt":   time.Now(),
		},
	}

	taskResult, err := j.taskCollection.UpdateMany(ctx, taskFilter, taskUpdate)
	if err != nil {
		log.Printf("[NEXUS-CLEANUP] Failed to update orphaned tasks: %v", err)
		return err
	}

	// Clean up stuck daemons
	daemonFilter := bson.M{
		"status":    bson.M{"$in": bson.A{"executing", "waiting_input", "idle"}},
		"startedAt": bson.M{"$lt": cutoff},
	}
	daemonUpdate := bson.M{
		"$set": bson.M{
			"status":        "failed",
			"currentAction": "orphan cleanup: stuck daemon",
			"completedAt":   time.Now(),
		},
	}

	daemonResult, err := j.daemonColl.UpdateMany(ctx, daemonFilter, daemonUpdate)
	if err != nil {
		log.Printf("[NEXUS-CLEANUP] Failed to update orphaned daemons: %v", err)
		return err
	}

	if taskResult.ModifiedCount > 0 || daemonResult.ModifiedCount > 0 {
		log.Printf("[NEXUS-CLEANUP] Cleaned up %d orphaned tasks, %d orphaned daemons (older than %s)",
			taskResult.ModifiedCount, daemonResult.ModifiedCount, cutoff.Format(time.RFC3339))
	}

	return nil
}

// GetNextRunTime returns when this job should next execute.
func (j *NexusTaskCleanupJob) GetNextRunTime() time.Time {
	if j.lastRun.IsZero() {
		return time.Now().Add(2 * time.Minute)
	}
	return j.lastRun.Add(j.interval)
}
