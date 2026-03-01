package jobs

import (
	"claraverse/internal/database"
	"claraverse/internal/services"
	"context"
	"log"
	"time"
)

// PromoExpirationChecker handles expiration of promotional pro subscriptions
// NOTE: This job is now a no-op since Pro tier is permanent for all users
type PromoExpirationChecker struct {
	mongoDB     *database.MongoDB
	userService *services.UserService
	tierService *services.TierService
}

// NewPromoExpirationChecker creates a new promo expiration checker
func NewPromoExpirationChecker(
	mongoDB *database.MongoDB,
	userService *services.UserService,
	tierService *services.TierService,
) *PromoExpirationChecker {
	return &PromoExpirationChecker{
		mongoDB:     mongoDB,
		userService: userService,
		tierService: tierService,
	}
}

// Run is now a no-op since Pro tier is permanent for all users
// Kept for backward compatibility in case this job is still registered
func (p *PromoExpirationChecker) Run(ctx context.Context) error {
	// Pro tier is now permanent for all users - no expiration needed
	log.Println("ℹ️  [PROMO-EXPIRATION] Skipped - Pro tier is now permanent for all users")
	return nil
}

// GetNextRunTime returns when the job should run next (hourly)
func (p *PromoExpirationChecker) GetNextRunTime() time.Time {
	return time.Now().UTC().Add(1 * time.Hour)
}
