import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';

// Helper function to get user-specific storage name
function getUserStorageName(baseName: string): string {
  try {
    const authStorage = localStorage.getItem('auth-storage');
    if (authStorage) {
      const { state } = JSON.parse(authStorage);
      if (state?.user?.id) {
        return `${baseName}-${state.user.id}`;
      }
    }
  } catch (error) {
    console.warn('Failed to get user ID for storage name:', error);
  }
  return baseName;
}

interface AppState {
  // State
  theme: 'light' | 'dark';
  sidebarOpen: boolean;

  // Actions
  setTheme: (theme: 'light' | 'dark') => void;
  toggleSidebar: () => void;
}

export const useAppStore = create<AppState>()(
  devtools(
    persist(
      set => ({
        // Initial state
        theme: 'light',
        sidebarOpen: true,

        // Actions
        setTheme: theme => set({ theme }),
        toggleSidebar: () => set(state => ({ sidebarOpen: !state.sidebarOpen })),
      }),
      {
        name: getUserStorageName('app-storage'),
      }
    )
  )
);
