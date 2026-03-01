import { authClient } from '@/lib/auth';

// Local user type (compatible with existing code)
export interface User {
  id: string;
  email: string;
  email_verified: boolean;
  role?: string;
  user_metadata?: Record<string, unknown>;
  created_at?: string;
}

// Local session type (compatible with existing code)
export interface Session {
  access_token: string;
  refresh_token: string;
  expires_at: number;
  expires_in: number;
  user: User;
}

export interface AuthError {
  message: string;
  status?: number;
}

export interface AuthResponse {
  user: User | null;
  session: Session | null;
  error: AuthError | null;
}

export const authService = {
  // Sign up with email and password
  async signUp(
    email: string,
    password: string,
    _username?: string,
    _captchaToken?: string
  ): Promise<AuthResponse> {
    try {
      const user = await authClient.signUp(email, password);

      // Get session after signup
      const { session, error } = await this.getSession();

      return {
        user,
        session,
        error,
      };
    } catch (error) {
      return {
        user: null,
        session: null,
        error: {
          message: error instanceof Error ? error.message : 'Registration failed',
        },
      };
    }
  },

  // Sign in with email and password
  async signIn(email: string, password: string, _captchaToken?: string): Promise<AuthResponse> {
    try {
      const user = await authClient.signIn(email, password);

      // Get session after login
      const { session, error } = await this.getSession();

      return {
        user,
        session,
        error,
      };
    } catch (error) {
      return {
        user: null,
        session: null,
        error: {
          message: error instanceof Error ? error.message : 'Login failed',
        },
      };
    }
  },

  // Sign in with GitHub (not supported in local mode)
  async signInWithGithub(): Promise<{ error: AuthError | null }> {
    return {
      error: {
        message: 'OAuth login not available in local mode. Use email/password authentication.',
      },
    };
  },

  // Sign in with Google (not supported in local mode)
  async signInWithGoogle(): Promise<{ error: AuthError | null }> {
    return {
      error: {
        message: 'OAuth login not available in local mode. Use email/password authentication.',
      },
    };
  },

  // Sign out
  async signOut(): Promise<{ error: AuthError | null }> {
    try {
      await authClient.signOut();
      return { error: null };
    } catch (err) {
      console.error('SignOut error:', err);
      return { error: null }; // Still clear local state
    }
  },

  // Get current session
  async getSession(): Promise<{ session: Session | null; error: AuthError | null }> {
    try {
      const user = authClient.getUser();
      const accessToken = authClient.getAccessToken();

      if (!user || !accessToken) {
        return { session: null, error: null };
      }

      // Decode JWT to get expiry (simple parsing, not cryptographic verification)
      const tokenParts = accessToken.split('.');
      if (tokenParts.length !== 3) {
        return { session: null, error: null };
      }

      const payload = JSON.parse(atob(tokenParts[1]));
      const expiresAt = payload.exp || 0;
      const expiresIn = Math.max(0, expiresAt - Math.floor(Date.now() / 1000));

      const session: Session = {
        access_token: accessToken,
        refresh_token: '', // Stored in httpOnly cookie
        expires_at: expiresAt,
        expires_in: expiresIn,
        user,
      };

      return { session, error: null };
    } catch (error) {
      return {
        session: null,
        error: {
          message: error instanceof Error ? error.message : 'Failed to get session',
        },
      };
    }
  },

  // Get current user
  async getCurrentUser(): Promise<{ user: User | null; error: AuthError | null }> {
    try {
      const user = authClient.getUser();
      return { user, error: null };
    } catch (error) {
      return {
        user: null,
        error: {
          message: error instanceof Error ? error.message : 'Failed to get user',
        },
      };
    }
  },

  // Listen to auth state changes
  // Note: Local JWT auth doesn't have real-time auth state changes like Supabase
  // This is a compatibility shim - returns a no-op unsubscribe function
  onAuthStateChange(callback: (user: User | null) => void): () => void {
    // Call once with current user
    const user = authClient.getUser();
    callback(user);

    // Return no-op unsubscribe function
    return () => {
      // No subscription to clean up
    };
  },

  // Reset password (requires SMTP to be configured on backend)
  async resetPassword(_email: string): Promise<{ error: AuthError | null }> {
    return {
      error: {
        message: 'Password reset not yet implemented. Contact your administrator to reset your password.',
      },
    };
  },

  // Update password
  async updatePassword(_newPassword: string): Promise<{ error: AuthError | null }> {
    return {
      error: {
        message: 'Password update not yet implemented. Use password reset functionality instead.',
      },
    };
  },

  // Refresh session
  async refreshSession(): Promise<{ session: Session | null; error: AuthError | null }> {
    try {
      await authClient.refreshToken();

      // Get updated session
      return await this.getSession();
    } catch (error) {
      return {
        session: null,
        error: {
          message: error instanceof Error ? error.message : 'Failed to refresh session',
        },
      };
    }
  },

  // Check if session is valid
  async isSessionValid(): Promise<boolean> {
    const { session } = await this.getSession();
    if (!session) return false;

    // Check if session is expired
    const now = Math.floor(Date.now() / 1000);
    return now < session.expires_at;
  },

  // Update user metadata (not supported in local mode)
  async updateUserMetadata(
    _metadata: Record<string, unknown>
  ): Promise<{ error: AuthError | null }> {
    return {
      error: {
        message: 'User metadata update not supported in local mode',
      },
    };
  },
};
