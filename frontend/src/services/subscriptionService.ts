import { api } from './api';
import type {
  Plan,
  Subscription,
  PlanChangePreview,
  PlanChangeResult,
  CheckoutResponse,
  PortalResponse,
  UsageStats,
} from '@/types/subscription';

/**
 * Subscription Service
 * Handles all subscription-related API calls
 */
export const subscriptionService = {
  /**
   * Get all available subscription plans
   */
  getPlans: async (): Promise<Plan[]> => {
    const response = await api.get<{ plans: Plan[] }>('/api/subscriptions/plans');
    return response.plans || [];
  },

  /**
   * Get the current user's subscription
   */
  getCurrentSubscription: async (): Promise<Subscription> => {
    return api.get<Subscription>('/api/subscriptions/current');
  },

  /**
   * Create a checkout session for a plan
   * @param planId - The plan ID to subscribe to (e.g., 'pro', 'max')
   * @returns Checkout URL to redirect user to
   */
  createCheckout: async (planId: string): Promise<CheckoutResponse> => {
    return api.post<CheckoutResponse>('/api/subscriptions/checkout', { plan_id: planId });
  },

  /**
   * Preview a plan change before confirming
   * Shows proration for upgrades, effective date for downgrades
   * @param planId - The new plan ID
   */
  previewPlanChange: async (planId: string): Promise<PlanChangePreview> => {
    return api.get<PlanChangePreview>(`/api/subscriptions/change-plan/preview?plan_id=${planId}`);
  },

  /**
   * Execute a plan change
   * - Upgrades: Immediate with proration
   * - Downgrades: Scheduled for end of billing period
   * @param planId - The new plan ID
   */
  changePlan: async (planId: string): Promise<PlanChangeResult> => {
    return api.post<PlanChangeResult>('/api/subscriptions/change-plan', { plan_id: planId });
  },

  /**
   * Cancel the current subscription
   * Cancellation takes effect at the end of the billing period
   */
  cancelSubscription: async (): Promise<void> => {
    await api.post('/api/subscriptions/cancel', {});
  },

  /**
   * Reactivate a cancelled subscription
   * Only works if the subscription is in pending_cancel status
   */
  reactivateSubscription: async (): Promise<void> => {
    await api.post('/api/subscriptions/reactivate', {});
  },

  /**
   * Get the DodoPayments customer portal URL
   * Allows users to manage payment methods, view invoices, etc.
   */
  getPortalURL: async (): Promise<string> => {
    const response = await api.get<PortalResponse>('/api/subscriptions/portal');
    return response.portal_url;
  },

  /**
   * Get invoices for the current user
   */
  getInvoices: async (): Promise<Invoice[]> => {
    const response = await api.get<{ invoices: Invoice[] }>('/api/subscriptions/invoices');
    return response.invoices || [];
  },

  /**
   * Get current usage statistics for the user
   * Returns usage data for schedules, API keys, executions, and requests
   */
  getUsageStats: async (): Promise<UsageStats> => {
    return api.get<UsageStats>('/api/subscriptions/usage');
  },

  /**
   * Sync subscription from payment provider
   * Useful after checkout to ensure local state is updated
   */
  syncSubscription: async (): Promise<Subscription> => {
    return api.post<Subscription>('/api/subscriptions/sync', {});
  },
};

// Invoice type (simplified, extend as needed)
export interface Invoice {
  id: string;
  amount: number;
  currency: string;
  status: 'paid' | 'pending' | 'failed';
  created_at: string;
  pdf_url?: string;
}

export default subscriptionService;
