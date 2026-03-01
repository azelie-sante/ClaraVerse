import { create } from 'zustand';
import type {
  Plan,
  Subscription,
  PlanChangePreview,
  SubscriptionTierType,
  UsageStats,
} from '@/types/subscription';
import { SubscriptionTier, SubscriptionStatus } from '@/types/subscription';

// Limit exceeded data structure
export interface LimitExceededData {
  type: 'messages' | 'file_uploads' | 'image_generations';
  limit: number;
  used: number;
  resetAt: string;
  suggestedTier: string;
}

interface SubscriptionState {
  // Data
  subscription: Subscription | null;
  plans: Plan[];
  usageStats: UsageStats | null;
  limitExceeded: LimitExceededData | null;

  // Loading states
  isLoadingSubscription: boolean;
  isLoadingPlans: boolean;
  isLoadingUsage: boolean;
  isChangingPlan: boolean;

  // Error states
  subscriptionError: string | null;
  plansError: string | null;
  usageError: string | null;

  // Actions
  fetchSubscription: () => Promise<void>;
  fetchPlans: () => Promise<void>;
  fetchUsageStats: () => Promise<void>;
  refreshAll: () => Promise<void>;
  syncSubscription: () => Promise<void>;
  previewPlanChange: (_planId: string) => Promise<PlanChangePreview | null>;
  createCheckout: (_planId: string) => Promise<string | null>;
  changePlan: (_planId: string) => Promise<boolean>;
  cancelSubscription: () => Promise<boolean>;
  reactivateSubscription: () => Promise<boolean>;
  getPortalURL: () => Promise<string | null>;
  setLimitExceeded: (data: LimitExceededData | null) => void;
  clearLimitExceeded: () => void;
  clearErrors: () => void;
  reset: () => void;
}

// OSS edition: all users get unlimited access (no subscription tiers)
const ossUnlimitedSubscription: Subscription = {
  id: 'oss',
  user_id: '',
  tier: SubscriptionTier.FREE,
  status: SubscriptionStatus.ACTIVE,
  cancel_at_period_end: false,
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
};

export const useSubscriptionStore = create<SubscriptionState>(set => ({
  // Initial state — OSS users always have an active subscription
  subscription: ossUnlimitedSubscription,
  plans: [],
  usageStats: null,
  limitExceeded: null,
  isLoadingSubscription: false,
  isLoadingPlans: false,
  isLoadingUsage: false,
  isChangingPlan: false,
  subscriptionError: null,
  plansError: null,
  usageError: null,

  // All actions are no-ops in OSS — no subscription backend exists
  fetchSubscription: async () => {},
  fetchPlans: async () => {},
  fetchUsageStats: async () => {},
  refreshAll: async () => {},
  syncSubscription: async () => {},
  previewPlanChange: async () => null,
  createCheckout: async () => null,
  changePlan: async () => true,
  cancelSubscription: async () => true,
  reactivateSubscription: async () => true,
  getPortalURL: async () => null,
  setLimitExceeded: (data: LimitExceededData | null) => {
    set({ limitExceeded: data });
  },
  clearLimitExceeded: () => {
    set({ limitExceeded: null });
  },
  clearErrors: () => {
    set({ subscriptionError: null, plansError: null, usageError: null });
  },
  reset: () => {
    set({
      subscription: ossUnlimitedSubscription,
      plans: [],
      usageStats: null,
      limitExceeded: null,
      isLoadingSubscription: false,
      isLoadingPlans: false,
      isLoadingUsage: false,
      isChangingPlan: false,
      subscriptionError: null,
      plansError: null,
      usageError: null,
    });
  },
}));

// Selectors
export const selectCurrentTier = (state: SubscriptionState): SubscriptionTierType => {
  return state.subscription?.tier || SubscriptionTier.FREE;
};

export const selectIsFreeTier = (state: SubscriptionState): boolean => {
  return selectCurrentTier(state) === SubscriptionTier.FREE;
};

export const selectIsPaidTier = (state: SubscriptionState): boolean => {
  const tier = selectCurrentTier(state);
  return tier !== SubscriptionTier.FREE;
};

export const selectCanUpgrade = (_state: SubscriptionState): boolean => {
  return false; // OSS: no upgrades
};

export const selectCanDowngrade = (_state: SubscriptionState): boolean => {
  return false; // OSS: no downgrades
};

export const selectIsCancelling = (state: SubscriptionState): boolean => {
  return state.subscription?.cancel_at_period_end === true;
};

export const selectPlanById = (state: SubscriptionState, planId: string): Plan | undefined => {
  return state.plans.find(p => p.id === planId);
};

export const selectCurrentPlan = (state: SubscriptionState): Plan | undefined => {
  const tier = selectCurrentTier(state);
  return state.plans.find(p => p.tier === tier);
};

export default useSubscriptionStore;
