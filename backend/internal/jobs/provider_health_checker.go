package jobs

import (
	"claraverse/internal/health"
	"context"
	"log"
	"time"
)

// ProviderHealthChecker performs periodic health checks on all registered providers
type ProviderHealthChecker struct {
	healthService *health.Service
	interval      time.Duration
	lastRun       time.Time
}

// NewProviderHealthChecker creates a new provider health checker job
func NewProviderHealthChecker(healthService *health.Service, interval time.Duration) *ProviderHealthChecker {
	return &ProviderHealthChecker{
		healthService: healthService,
		interval:      interval,
	}
}

// Run executes health checks on all registered providers across all capabilities
func (p *ProviderHealthChecker) Run(ctx context.Context) error {
	log.Println("[HEALTH-JOB] Starting provider health checks...")
	p.lastRun = time.Now()

	capabilities := []health.CapabilityType{
		health.CapabilityChat,
		health.CapabilityVision,
		health.CapabilityImageGen,
		health.CapabilityImageEdit,
		health.CapabilityAudio,
	}

	totalChecked := 0
	totalHealthy := 0
	totalFailed := 0

	for _, cap := range capabilities {
		providers := p.healthService.GetAllProviders(cap)
		if len(providers) == 0 {
			continue
		}

		log.Printf("[HEALTH-JOB] Checking %d %s provider(s)...", len(providers), cap)

		for _, provider := range providers {
			select {
			case <-ctx.Done():
				log.Println("[HEALTH-JOB] Cancelled")
				return ctx.Err()
			default:
			}

			err := p.healthService.CheckProviderHealth(cap, provider.ProviderID, provider.ModelName)
			totalChecked++

			if err != nil {
				totalFailed++
				log.Printf("[HEALTH-JOB] %s %s/%s: FAILED (%v)",
					cap, provider.ProviderName, provider.ModelName, err)
			} else {
				totalHealthy++
			}

			// Small delay between checks to avoid rate limiting
			time.Sleep(2 * time.Second)
		}
	}

	log.Printf("[HEALTH-JOB] Health checks complete: %d checked, %d healthy, %d failed",
		totalChecked, totalHealthy, totalFailed)
	return nil
}

// GetNextRunTime returns when the next health check should run
func (p *ProviderHealthChecker) GetNextRunTime() time.Time {
	if p.lastRun.IsZero() {
		// First run: wait 2 minutes after startup to let providers register
		return time.Now().Add(2 * time.Minute)
	}
	return p.lastRun.Add(p.interval)
}
