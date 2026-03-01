export interface User {
  id: string;
  name: string;
  email: string;
  createdAt: string;
}

export interface ApiResponse<T> {
  data: T;
  message?: string;
  status: number;
}

export type LoadingState = 'idle' | 'loading' | 'success' | 'error';

// Re-export credential types
export * from './credential';

// Re-export subscription types
export * from './subscription';
