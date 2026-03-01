import { useAuthStore } from '@/store/useAuthStore';

/**
 * API client utility with automatic authentication
 */

interface FetchOptions extends RequestInit {
  requiresAuth?: boolean;
  skipAuthRetry?: boolean;
}

// Track if we're currently refreshing to prevent multiple refresh attempts
let isRefreshing = false;
let refreshPromise: Promise<boolean> | null = null;

export const apiClient = {
  /**
   * Make an authenticated API request
   */
  async fetch(url: string, options: FetchOptions = {}) {
    const { requiresAuth = true, skipAuthRetry = false, headers = {}, ...restOptions } = options;

    // Get access token from auth store
    const accessToken = useAuthStore.getState().getAccessToken();

    // Add authorization header if token exists and auth is required
    const finalHeaders: HeadersInit = {
      'Content-Type': 'application/json',
      ...headers,
    };

    if (requiresAuth && accessToken) {
      (finalHeaders as Record<string, string>)['Authorization'] = `Bearer ${accessToken}`;
    }

    try {
      const response = await fetch(url, {
        ...restOptions,
        headers: finalHeaders,
      });

      // Handle 401 Unauthorized - token might be expired
      if (response.status === 401 && requiresAuth && !skipAuthRetry) {
        // Try to refresh session (only once)
        const refreshSucceeded = await attemptTokenRefresh();

        if (refreshSucceeded) {
          // Retry request with new token
          const newToken = useAuthStore.getState().getAccessToken();
          if (newToken) {
            (finalHeaders as Record<string, string>)['Authorization'] = `Bearer ${newToken}`;
            // Use skipAuthRetry to prevent infinite loop
            return fetch(url, {
              ...restOptions,
              headers: finalHeaders,
            });
          }
        }

        // Refresh failed - force logout
        useAuthStore.getState().forceLogout('Your session has expired. Please sign in again.');

        // Return the original 401 response
        return response;
      }

      return response;
    } catch (error) {
      console.error('API request failed:', error);
      throw error;
    }
  },

  /**
   * GET request
   */
  async get(url: string, options: FetchOptions = {}) {
    return this.fetch(url, { ...options, method: 'GET' });
  },

  /**
   * POST request
   */
  async post(url: string, data?: unknown, options: FetchOptions = {}) {
    return this.fetch(url, {
      ...options,
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    });
  },

  /**
   * PUT request
   */
  async put(url: string, data?: unknown, options: FetchOptions = {}) {
    return this.fetch(url, {
      ...options,
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    });
  },

  /**
   * PATCH request
   */
  async patch(url: string, data?: unknown, options: FetchOptions = {}) {
    return this.fetch(url, {
      ...options,
      method: 'PATCH',
      body: data ? JSON.stringify(data) : undefined,
    });
  },

  /**
   * DELETE request
   */
  async delete(url: string, options: FetchOptions = {}) {
    return this.fetch(url, { ...options, method: 'DELETE' });
  },

  /**
   * Get the current access token
   */
  getToken() {
    return useAuthStore.getState().getAccessToken();
  },
};

/**
 * Attempt to refresh the authentication token
 * Uses a singleton pattern to prevent multiple simultaneous refresh attempts
 */
async function attemptTokenRefresh(): Promise<boolean> {
  // If already refreshing, wait for the existing refresh to complete
  if (isRefreshing && refreshPromise) {
    return refreshPromise;
  }

  isRefreshing = true;
  refreshPromise = (async () => {
    try {
      const { checkSession } = useAuthStore.getState();
      const success = await checkSession();
      return success;
    } catch (error) {
      console.error('Token refresh failed:', error);
      return false;
    } finally {
      isRefreshing = false;
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}

/**
 * Hook to use authenticated API client
 */
export function useApiClient() {
  const { accessToken, session } = useAuthStore();

  return {
    apiClient,
    accessToken,
    session,
    isAuthenticated: !!accessToken,
  };
}
