import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import * as clawService from '@/services/clawService';
import type {
  ClawStatus,
  Routine,
  CreateRoutineRequest,
  UpdateRoutineRequest,
  ToolCategory,
} from '@/services/clawService';

interface ClawState {
  // Connection status
  telegramConnected: boolean;
  mcpConnected: boolean;

  // Routines
  routines: Routine[];

  // Tools
  toolCategories: ToolCategory[];
  toolsLoading: boolean;

  // Actions
  fetchStatus: () => Promise<void>;
  fetchRoutines: () => Promise<void>;
  createRoutine: (data: CreateRoutineRequest) => Promise<Routine>;
  updateRoutine: (id: string, data: UpdateRoutineRequest) => Promise<void>;
  deleteRoutine: (id: string) => Promise<void>;
  triggerRoutine: (id: string) => Promise<void>;
  fetchTools: () => Promise<void>;
}

export const useClawStore = create<ClawState>()(
  devtools(
    (set, get) => ({
      // Initial state
      telegramConnected: false,
      mcpConnected: false,
      routines: [],
      toolCategories: [],
      toolsLoading: false,

      fetchStatus: async () => {
        try {
          const status: ClawStatus = await clawService.getClawStatus();
          set({
            telegramConnected: status.telegram.connected,
            mcpConnected: status.mcp.connected,
          });
        } catch {
          // Status fetch is best-effort — badges just stay disconnected
        }
      },

      fetchRoutines: async () => {
        try {
          const routines = await clawService.listRoutines();
          set({ routines });
        } catch {
          // Silently fail — empty list is fine
        }
      },

      createRoutine: async (data: CreateRoutineRequest) => {
        const routine = await clawService.createRoutine(data);
        set({ routines: [...get().routines, routine] });
        return routine;
      },

      updateRoutine: async (id: string, data: UpdateRoutineRequest) => {
        const updated = await clawService.updateRoutine(id, data);
        set({
          routines: get().routines.map(r => (r.id === id ? updated : r)),
        });
      },

      deleteRoutine: async (id: string) => {
        await clawService.deleteRoutine(id);
        set({
          routines: get().routines.filter(r => r.id !== id),
        });
      },

      triggerRoutine: async (id: string) => {
        await clawService.triggerRoutine(id);
      },

      fetchTools: async () => {
        set({ toolsLoading: true });
        try {
          const response = await clawService.getAvailableTools();
          set({ toolCategories: response.categories, toolsLoading: false });
        } catch {
          set({ toolsLoading: false });
        }
      },
    }),
    { name: 'claw-store' }
  )
);
