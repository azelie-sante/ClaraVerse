import { api, ApiError } from '@/services/api';

export type ChatPrivacyMode = 'local' | 'cloud' | null;

export interface UserPreferences {
  store_builder_chat_history: boolean;
  default_model_id?: string;
  tool_predictor_model_id?: string;
  chat_privacy_mode?: ChatPrivacyMode;
  theme?: string;
  font_size?: string;

  // Memory system preferences
  memory_enabled?: boolean;
  memory_extraction_threshold?: number;
  memory_max_injection?: number;
  memory_extractor_model_id?: string;
  memory_selector_model_id?: string;
}

export interface UpdatePreferencesRequest {
  store_builder_chat_history?: boolean;
  default_model_id?: string;
  tool_predictor_model_id?: string;
  chat_privacy_mode?: ChatPrivacyMode;
  theme?: string;
  font_size?: string;

  // Memory system preferences
  memory_enabled?: boolean;
  memory_extraction_threshold?: number;
  memory_max_injection?: number;
  memory_extractor_model_id?: string;
  memory_selector_model_id?: string;
}

/**
 * Get user preferences from backend
 */
export async function getPreferences(): Promise<UserPreferences> {
  try {
    return await api.get<UserPreferences>('/api/preferences');
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      console.log('User not authenticated, using local preferences');
      throw error;
    }
    console.error('Failed to get preferences:', error);
    throw error;
  }
}

/**
 * Update user preferences on backend
 */
export async function updatePreferences(
  updates: UpdatePreferencesRequest
): Promise<UserPreferences> {
  return api.put<UserPreferences>('/api/preferences', updates);
}

/**
 * Set chat privacy mode on backend
 */
export async function setChatPrivacyMode(mode: ChatPrivacyMode): Promise<UserPreferences> {
  return updatePreferences({ chat_privacy_mode: mode });
}

/**
 * Check if user is authenticated
 */
export function isAuthenticated(): boolean {
  try {
    const authStorage = localStorage.getItem('auth-storage');
    if (authStorage) {
      const { state } = JSON.parse(authStorage);
      return !!state?.accessToken;
    }
  } catch {
    // Invalid JSON
  }
  return false;
}
