package jobs

import (
	"context"
	"log"
	"sync"
	"time"
)

// Job interface that all scheduled jobs must implement
type Job interface {
	Run(ctx context.Context) error
	GetNextRunTime() time.Time
}

// JobScheduler manages and runs scheduled jobs
type JobScheduler struct {
	jobs    map[string]Job
	timers  map[string]*time.Timer
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	mu      sync.Mutex
	running bool
}

// NewJobScheduler creates a new job scheduler
func NewJobScheduler() *JobScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &JobScheduler{
		jobs:   make(map[string]Job),
		timers: make(map[string]*time.Timer),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Register adds a job to the scheduler
func (s *JobScheduler) Register(name string, job Job) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs[name] = job
	log.Printf("✅ [SCHEDULER] Registered job: %s", name)
}

// Start begins running all registered jobs
func (s *JobScheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	s.running = true
	log.Printf("🚀 [SCHEDULER] Starting job scheduler with %d jobs", len(s.jobs))

	// Schedule all jobs
	for name, job := range s.jobs {
		s.scheduleJob(name, job)
	}

	return nil
}

// scheduleJob schedules a single job
func (s *JobScheduler) scheduleJob(name string, job Job) {
	nextRun := job.GetNextRunTime()
	duration := time.Until(nextRun)

	log.Printf("⏰ [SCHEDULER] Job '%s' scheduled to run at %s (in %v)",
		name, nextRun.Format(time.RFC3339), duration)

	timer := time.AfterFunc(duration, func() {
		s.runJob(name, job)
	})

	s.timers[name] = timer
}

// runJob executes a job and reschedules it
func (s *JobScheduler) runJob(name string, job Job) {
	s.wg.Add(1)
	defer s.wg.Done()

	log.Printf("▶️  [SCHEDULER] Running job: %s", name)
	startTime := time.Now()

	// Run the job
	if err := job.Run(s.ctx); err != nil {
		log.Printf("❌ [SCHEDULER] Job '%s' failed: %v", name, err)
	}

	duration := time.Since(startTime)
	log.Printf("✅ [SCHEDULER] Job '%s' completed in %v", name, duration)

	// Reschedule the job
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		s.scheduleJob(name, job)
	}
}

// Stop gracefully stops all jobs
func (s *JobScheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}

	log.Println("🛑 [SCHEDULER] Stopping job scheduler...")
	s.running = false

	// Stop all timers
	for name, timer := range s.timers {
		timer.Stop()
		log.Printf("⏹️  [SCHEDULER] Stopped job: %s", name)
	}
	s.timers = make(map[string]*time.Timer)

	s.mu.Unlock()

	// Cancel context and wait for running jobs
	s.cancel()
	s.wg.Wait()

	log.Println("✅ [SCHEDULER] Job scheduler stopped")
}

// RunNow immediately runs a specific job (useful for testing)
func (s *JobScheduler) RunNow(name string) error {
	s.mu.Lock()
	job, exists := s.jobs[name]
	s.mu.Unlock()

	if !exists {
		log.Printf("⚠️  [SCHEDULER] Job '%s' not found", name)
		return nil
	}

	log.Printf("🚀 [SCHEDULER] Running job '%s' immediately", name)
	return job.Run(s.ctx)
}

// GetStatus returns the status of all jobs
func (s *JobScheduler) GetStatus() map[string]JobStatus {
	s.mu.Lock()
	defer s.mu.Unlock()

	status := make(map[string]JobStatus)
	for name, job := range s.jobs {
		status[name] = JobStatus{
			Name:        name,
			NextRunTime: job.GetNextRunTime(),
			Registered:  true,
		}
	}

	return status
}

// JobStatus represents the status of a job
type JobStatus struct {
	Name        string    `json:"name"`
	NextRunTime time.Time `json:"next_run_time"`
	Registered  bool      `json:"registered"`
}
