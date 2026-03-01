// Subscription status constants
export const SubscriptionStatus = {
  ACTIVE: 'active',
  ON_HOLD: 'on_hold',
  PENDING_CANCEL: 'pending_cancel',
  CANCELLED: 'cancelled',
  PAUSED: 'paused',
} as const;

export type SubscriptionStatusType = (typeof SubscriptionStatus)[keyof typeof SubscriptionStatus];

// Subscription tiers
export const SubscriptionTier = {
  FREE: 'free',
  PRO: 'pro',
  MAX: 'max',
  ENTERPRISE: 'enterprise',
  LEGACY_UNLIMITED: 'legacy_unlimited',
} as const;

export type SubscriptionTierType = (typeof SubscriptionTier)[keyof typeof SubscriptionTier];

// Tier order for comparison (lower = lower tier)
export const TierOrder: Record<SubscriptionTierType, number> = {
  [SubscriptionTier.FREE]: 0,
  [SubscriptionTier.PRO]: 1,
  [SubscriptionTier.MAX]: 2,
  [SubscriptionTier.ENTERPRISE]: 3,
  [SubscriptionTier.LEGACY_UNLIMITED]: 4, // Highest tier
};

// TierLimits defines rate limits and quotas per subscription tier
export interface TierLimits {
  maxSchedules: number;
  maxApiKeys: number;
  requestsPerMinute: number;
  requestsPerHour: number;
  retentionDays: number;
  maxExecutionsPerDay: number;
  maxMessagesPerMonth: number;
  maxFileUploadsPerDay: number;
  maxImageGensPerDay: number;
}

// Plan represents a subscription plan with pricing
export interface Plan {
  id: string;
  name: string;
  tier: SubscriptionTierType;
  price_monthly: number; // cents
  dodo_product_id: string;
  features: string[];
  limits: TierLimits;
  contact_sales: boolean;
}

// Subscription tracks a user's subscription state
export interface Subscription {
  id: string;
  user_id: string;
  dodo_subscription_id?: string;
  dodo_customer_id?: string;

  // Current state
  tier: SubscriptionTierType;
  status: SubscriptionStatusType;

  // Billing info
  current_period_start?: string;
  current_period_end?: string;

  // Scheduled changes (for downgrades/cancellations)
  scheduled_tier?: string;
  scheduled_change_at?: string;
  cancel_at_period_end: boolean;

  // Promo user fields
  subscription_expires_at?: string;
  is_promo_user?: boolean;
  has_seen_welcome_popup?: boolean;

  // Timestamps
  created_at: string;
  updated_at: string;
  cancelled_at?: string;
}

// PlanChangePreview shows what will happen before confirming a plan change
export interface PlanChangePreview {
  change_type: 'upgrade' | 'downgrade';
  immediate: boolean;
  current_tier: SubscriptionTierType;
  new_tier: SubscriptionTierType;
  prorated_amount?: number; // cents, only for upgrades
  effective_at?: string; // ISO date string
}

// PlanChangeResult is the result after a plan change
export interface PlanChangeResult {
  type: 'upgrade' | 'downgrade';
  immediate: boolean;
  new_tier: SubscriptionTierType;
  effective_at?: string;
}

// CheckoutResponse from creating a checkout session
export interface CheckoutResponse {
  checkout_url: string;
  session_id: string;
}

// PortalResponse from getting customer portal URL
export interface PortalResponse {
  portal_url: string;
}

// UsageStat represents a single usage metric with current value and max
export interface UsageStat {
  current: number;
  max: number;
}

// UsageStatWithTime represents a usage metric with current value, limit, and reset time
export interface UsageStatWithTime {
  current: number;
  max: number; // Backend returns 'max' not 'limit'
  reset_at: string; // ISO 8601 datetime
}

// UsageStats represents all current usage statistics for a user
export interface UsageStats {
  schedules: UsageStat;
  api_keys: UsageStat;
  executions_today: UsageStat;
  requests_per_min: UsageStat;
  messages: UsageStatWithTime;
  file_uploads: UsageStatWithTime;
  image_generations: UsageStatWithTime;
}

// Helper functions

/**
 * Compare two tiers
 * @returns -1 if fromTier < toTier (upgrade), 0 if equal, 1 if fromTier > toTier (downgrade)
 */
export function compareTiers(fromTier: SubscriptionTierType, toTier: SubscriptionTierType): number {
  const fromOrder = TierOrder[fromTier] ?? 0;
  const toOrder = TierOrder[toTier] ?? 0;

  if (fromOrder < toOrder) return -1;
  if (fromOrder > toOrder) return 1;
  return 0;
}

/**
 * Check if a tier change is an upgrade
 */
export function isUpgrade(fromTier: SubscriptionTierType, toTier: SubscriptionTierType): boolean {
  return compareTiers(fromTier, toTier) < 0;
}

/**
 * Check if a tier change is a downgrade
 */
export function isDowngrade(fromTier: SubscriptionTierType, toTier: SubscriptionTierType): boolean {
  return compareTiers(fromTier, toTier) > 0;
}

/**
 * Check if subscription is currently active (user has access)
 */
export function isSubscriptionActive(subscription: Subscription): boolean {
  return [
    SubscriptionStatus.ACTIVE,
    SubscriptionStatus.ON_HOLD,
    SubscriptionStatus.PENDING_CANCEL,
  ].includes(subscription.status as SubscriptionStatusType);
}

/**
 * Format price from cents to display string
 */
export function formatPrice(cents: number): string {
  if (cents === 0) return 'Free';
  return `$${(cents / 100).toFixed(2)}`;
}

/**
 * Format price for monthly display
 */
export function formatMonthlyPrice(cents: number): string {
  if (cents === 0) return 'Free';
  return `$${(cents / 100).toFixed(2)}/mo`;
}

/**
 * Get tier display name
 */
export function getTierDisplayName(tier: SubscriptionTierType): string {
  const names: Record<SubscriptionTierType, string> = {
    [SubscriptionTier.FREE]: 'Free',
    [SubscriptionTier.PRO]: 'Pro',
    [SubscriptionTier.MAX]: 'Max',
    [SubscriptionTier.ENTERPRISE]: 'Enterprise',
    [SubscriptionTier.LEGACY_UNLIMITED]: 'Legacy Unlimited',
  };
  return names[tier] || tier;
}

/**
 * Get status display name
 */
export function getStatusDisplayName(status: SubscriptionStatusType): string {
  const names: Record<SubscriptionStatusType, string> = {
    [SubscriptionStatus.ACTIVE]: 'Active',
    [SubscriptionStatus.ON_HOLD]: 'On Hold',
    [SubscriptionStatus.PENDING_CANCEL]: 'Cancelling',
    [SubscriptionStatus.CANCELLED]: 'Cancelled',
    [SubscriptionStatus.PAUSED]: 'Paused',
  };
  return names[status] || status;
}

/**
 * Get status color for badges
 */
export function getStatusColor(
  status: SubscriptionStatusType
): 'success' | 'warning' | 'error' | 'default' {
  switch (status) {
    case SubscriptionStatus.ACTIVE:
      return 'success';
    case SubscriptionStatus.ON_HOLD:
    case SubscriptionStatus.PENDING_CANCEL:
      return 'warning';
    case SubscriptionStatus.CANCELLED:
      return 'error';
    default:
      return 'default';
  }
}

/**
 * Get tier color for badges
 */
export function getTierColor(
  tier: SubscriptionTierType
): 'default' | 'primary' | 'success' | 'warning' {
  switch (tier) {
    case SubscriptionTier.FREE:
      return 'default';
    case SubscriptionTier.PRO:
      return 'primary';
    case SubscriptionTier.MAX:
      return 'success';
    case SubscriptionTier.ENTERPRISE:
      return 'warning';
    case SubscriptionTier.LEGACY_UNLIMITED:
      return 'success'; // Green for legacy users
    default:
      return 'default';
  }
}

/**
 * Format limit value (-1 means unlimited)
 */
export function formatLimit(value: number): string {
  return value === -1 ? 'Unlimited' : value.toLocaleString();
}
