import { apiClient } from '@/lib/apiClient';
import { getApiBaseUrl } from '@/lib/config';

const API_BASE_URL = getApiBaseUrl();

export interface E2BSettings {
  api_key_set: boolean;
  api_key_masked: string;
}

/**
 * Fetch current E2B settings
 */
export async function fetchE2BSettings(): Promise<E2BSettings> {
  const response = await apiClient.get(`${API_BASE_URL}/api/admin/e2b-settings`, {
    requiresAuth: true,
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch E2B settings: ${response.statusText}`);
  }

  return await response.json();
}

/**
 * Update E2B API key
 */
export async function updateE2BApiKey(apiKey: string): Promise<void> {
  const response = await apiClient.put(
    `${API_BASE_URL}/api/admin/e2b-settings`,
    { api_key: apiKey },
    { requiresAuth: true }
  );

  if (!response.ok) {
    throw new Error(`Failed to update E2B settings: ${response.statusText}`);
  }
}
