import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { adminService } from '@/services/adminService';
import type {
  OverviewStats,
  ProviderAnalytics,
  ModelAnalytics,
  ChatAnalytics,
  AgentAnalytics,
  HealthDashboardResponse,
  InsightsOverview,
  DailyMetrics,
  HealthDistribution,
  ActivationFunnel,
  FeedbackStreamResponse,
  AutoPilotContext,
  CollectionStats,
  AutoPilotAnalysisResponse,
} from '@/types/admin';

interface AdminState {
  // Analytics data
  overviewStats: OverviewStats | null;
  providerAnalytics: ProviderAnalytics[];
  modelAnalytics: ModelAnalytics[];
  chatAnalytics: ChatAnalytics | null;
  agentAnalytics: AgentAnalytics | null;

  // Health dashboard
  healthDashboard: HealthDashboardResponse | null;
  isLoadingHealth: boolean;
  healthError: string | null;

  // Insights dashboard
  insightsOverview: InsightsOverview | null;
  insightsMetrics: DailyMetrics[];
  healthDistribution: HealthDistribution | null;
  activationFunnel: ActivationFunnel | null;
  feedbackStream: FeedbackStreamResponse | null;
  isLoadingInsights: boolean;
  insightsError: string | null;

  // Collection stats (raw DB data)
  collectionStats: CollectionStats | null;
  isLoadingCollectionStats: boolean;

  // AutoPilot dashboard
  autoPilotContext: AutoPilotContext | null;
  isLoadingAutoPilot: boolean;
  autoPilotError: string | null;

  // AI Analysis
  aiAnalysis: AutoPilotAnalysisResponse | null;
  isAnalyzing: boolean;
  analysisError: string | null;

  // Loading states
  isLoadingStats: boolean;
  isLoadingProviders: boolean;
  isLoadingModels: boolean;
  isLoadingChats: boolean;
  isLoadingAgents: boolean;

  // Errors
  statsError: string | null;
  providersError: string | null;
  modelsError: string | null;
  chatsError: string | null;
  agentsError: string | null;

  // Actions
  fetchOverviewStats: () => Promise<void>;
  fetchProviderAnalytics: () => Promise<void>;
  fetchModelAnalytics: () => Promise<void>;
  fetchChatAnalytics: () => Promise<void>;
  fetchAgentAnalytics: () => Promise<void>;
  fetchHealthDashboard: () => Promise<void>;
  fetchInsightsDashboard: () => Promise<void>;
  fetchCollectionStats: () => Promise<void>;
  backfillMetrics: (days?: number) => Promise<number>;
  fetchAutoPilotContext: () => Promise<void>;
  analyzeWithAI: (modelId: string) => Promise<void>;
  refreshAllAnalytics: () => Promise<void>;
}

export const useAdminStore = create<AdminState>()(
  devtools(
    set => ({
      // Initial state
      overviewStats: null,
      providerAnalytics: [],
      modelAnalytics: [],
      chatAnalytics: null,
      agentAnalytics: null,

      healthDashboard: null,
      isLoadingHealth: false,
      healthError: null,

      // Insights
      insightsOverview: null,
      insightsMetrics: [],
      healthDistribution: null,
      activationFunnel: null,
      feedbackStream: null,
      isLoadingInsights: false,
      insightsError: null,

      // Collection stats
      collectionStats: null,
      isLoadingCollectionStats: false,

      // AutoPilot
      autoPilotContext: null,
      isLoadingAutoPilot: false,
      autoPilotError: null,

      // AI Analysis
      aiAnalysis: null,
      isAnalyzing: false,
      analysisError: null,

      isLoadingStats: false,
      isLoadingProviders: false,
      isLoadingModels: false,
      isLoadingChats: false,
      isLoadingAgents: false,

      statsError: null,
      providersError: null,
      modelsError: null,
      chatsError: null,
      agentsError: null,

      // Actions
      fetchOverviewStats: async () => {
        set({ isLoadingStats: true, statsError: null });
        try {
          const stats = await adminService.getOverviewStats();
          set({ overviewStats: stats, isLoadingStats: false });
        } catch (error) {
          console.error('Failed to fetch overview stats:', error);
          set({
            statsError: error instanceof Error ? error.message : 'Failed to fetch stats',
            isLoadingStats: false,
          });
        }
      },

      fetchProviderAnalytics: async () => {
        set({ isLoadingProviders: true, providersError: null });
        try {
          const analytics = await adminService.getProviderAnalytics();
          set({ providerAnalytics: analytics, isLoadingProviders: false });
        } catch (error) {
          console.error('Failed to fetch provider analytics:', error);
          set({
            providersError:
              error instanceof Error ? error.message : 'Failed to fetch provider analytics',
            isLoadingProviders: false,
          });
        }
      },

      fetchModelAnalytics: async () => {
        set({ isLoadingModels: true, modelsError: null });
        try {
          const analytics = await adminService.getModelAnalytics();
          set({ modelAnalytics: analytics, isLoadingModels: false });
        } catch (error) {
          console.error('Failed to fetch model analytics:', error);
          set({
            modelsError: error instanceof Error ? error.message : 'Failed to fetch model analytics',
            isLoadingModels: false,
          });
        }
      },

      fetchChatAnalytics: async () => {
        set({ isLoadingChats: true, chatsError: null });
        try {
          const analytics = await adminService.getChatAnalytics();
          set({ chatAnalytics: analytics, isLoadingChats: false });
        } catch (error) {
          console.error('Failed to fetch chat analytics:', error);
          set({
            chatsError: error instanceof Error ? error.message : 'Failed to fetch chat analytics',
            isLoadingChats: false,
          });
        }
      },

      fetchAgentAnalytics: async () => {
        set({ isLoadingAgents: true, agentsError: null });
        try {
          const analytics = await adminService.getAgentAnalytics();
          set({ agentAnalytics: analytics, isLoadingAgents: false });
        } catch (error) {
          console.error('Failed to fetch agent analytics:', error);
          set({
            agentsError: error instanceof Error ? error.message : 'Failed to fetch agent analytics',
            isLoadingAgents: false,
          });
        }
      },

      fetchHealthDashboard: async () => {
        set({ isLoadingHealth: true, healthError: null });
        try {
          const data = await adminService.getHealthDashboard();
          set({ healthDashboard: data, isLoadingHealth: false });
        } catch (error) {
          console.error('Failed to fetch health dashboard:', error);
          set({
            healthError: error instanceof Error ? error.message : 'Failed to fetch health data',
            isLoadingHealth: false,
          });
        }
      },

      fetchInsightsDashboard: async () => {
        set({ isLoadingInsights: true, insightsError: null });
        try {
          // Fetch all insights data in parallel
          const [overview, metricsResponse, healthDist, funnel, feedback] = await Promise.all([
            adminService.getInsightsOverview(),
            adminService.getInsightsMetrics(30),
            adminService.getHealthDistribution(),
            adminService.getActivationFunnel(),
            adminService.getFeedbackStream(20, 'all'),
          ]);
          set({
            insightsOverview: overview,
            insightsMetrics: metricsResponse.metrics,
            healthDistribution: healthDist,
            activationFunnel: funnel,
            feedbackStream: feedback,
            isLoadingInsights: false,
          });
        } catch (error) {
          console.error('Failed to fetch insights dashboard:', error);
          set({
            insightsError: error instanceof Error ? error.message : 'Failed to fetch insights data',
            isLoadingInsights: false,
          });
        }
      },

      fetchCollectionStats: async () => {
        set({ isLoadingCollectionStats: true });
        try {
          const stats = await adminService.getCollectionStats();
          set({ collectionStats: stats, isLoadingCollectionStats: false });
        } catch (error) {
          console.error('Failed to fetch collection stats:', error);
          set({ isLoadingCollectionStats: false });
        }
      },

      backfillMetrics: async (days = 90) => {
        try {
          const result = await adminService.backfillMetrics(days);
          return result.processed;
        } catch (error) {
          console.error('Failed to backfill metrics:', error);
          return 0;
        }
      },

      fetchAutoPilotContext: async () => {
        set({ isLoadingAutoPilot: true, autoPilotError: null });
        try {
          const context = await adminService.getAutoPilotContext();
          set({ autoPilotContext: context, isLoadingAutoPilot: false });
        } catch (error) {
          console.error('Failed to fetch autopilot context:', error);
          set({
            autoPilotError:
              error instanceof Error ? error.message : 'Failed to fetch autopilot context',
            isLoadingAutoPilot: false,
          });
        }
      },

      analyzeWithAI: async (modelId: string) => {
        set({ isAnalyzing: true, analysisError: null });
        try {
          const result = await adminService.analyzeWithAI(modelId);
          set({ aiAnalysis: result, isAnalyzing: false });
        } catch (error) {
          console.error('Failed to analyze with AI:', error);
          const msg = error instanceof Error ? error.message : 'AI analysis failed';
          set({ analysisError: msg, isAnalyzing: false });
        }
      },

      refreshAllAnalytics: async () => {
        // Fetch all analytics in parallel
        await Promise.allSettled([
          adminService.getOverviewStats().then(stats => set({ overviewStats: stats })),
          adminService
            .getProviderAnalytics()
            .then(analytics => set({ providerAnalytics: analytics })),
          adminService.getModelAnalytics().then(analytics => set({ modelAnalytics: analytics })),
          adminService.getChatAnalytics().then(analytics => set({ chatAnalytics: analytics })),
          adminService.getAgentAnalytics().then(analytics => set({ agentAnalytics: analytics })),
        ]);
      },
    }),
    {
      name: 'admin-store',
    }
  )
);
