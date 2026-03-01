import { api } from './api';

export interface Memory {
  id: string;
  content: string;
  category: 'personal_info' | 'preferences' | 'context' | 'fact' | 'instruction';
  tags: string[];
  score: number;
  access_count: number;
  last_accessed_at: string | null;
  is_archived: boolean;
  archived_at: string | null;
  source_engagement: number;
  conversation_id: string;
  created_at: string;
  updated_at: string;
  version: number;
}

export interface MemoryListResponse {
  memories: Memory[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

export interface MemoryStats {
  total_memories: number;
  active_memories: number;
  archived_memories: number;
  avg_score: number;
}

export interface CreateMemoryRequest {
  content: string;
  category: 'personal_info' | 'preferences' | 'context' | 'fact' | 'instruction';
  tags: string[];
}

export interface UpdateMemoryRequest {
  content?: string;
  category?: 'personal_info' | 'preferences' | 'context' | 'fact' | 'instruction';
  tags?: string[];
}

export interface ListMemoriesParams {
  category?: string;
  tags?: string;
  includeArchived?: boolean;
  page?: number;
  pageSize?: number;
}

class MemoryService {
  /**
   * List memories with optional filters
   */
  async listMemories(params: ListMemoriesParams = {}): Promise<MemoryListResponse> {
    const queryParams = new URLSearchParams();

    if (params.category) queryParams.append('category', params.category);
    if (params.tags) queryParams.append('tags', params.tags);
    if (params.includeArchived !== undefined)
      queryParams.append('includeArchived', String(params.includeArchived));
    if (params.page) queryParams.append('page', String(params.page));
    if (params.pageSize) queryParams.append('pageSize', String(params.pageSize));

    const url = `/api/memories${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
    return api.get<MemoryListResponse>(url);
  }

  /**
   * Get a single memory by ID
   */
  async getMemory(id: string): Promise<Memory> {
    return api.get<Memory>(`/api/memories/${id}`);
  }

  /**
   * Create a new memory
   */
  async createMemory(request: CreateMemoryRequest): Promise<Memory> {
    return api.post<Memory>('/api/memories', request);
  }

  /**
   * Update an existing memory
   */
  async updateMemory(id: string, request: UpdateMemoryRequest): Promise<Memory> {
    return api.put<Memory>(`/api/memories/${id}`, request);
  }

  /**
   * Delete a memory permanently
   */
  async deleteMemory(id: string): Promise<{ success: boolean; message: string }> {
    return api.delete<{ success: boolean; message: string }>(`/api/memories/${id}`);
  }

  /**
   * Archive a memory
   */
  async archiveMemory(id: string): Promise<{ success: boolean; message: string }> {
    return api.post<{ success: boolean; message: string }>(`/api/memories/${id}/archive`, {});
  }

  /**
   * Unarchive a memory
   */
  async unarchiveMemory(id: string): Promise<{ success: boolean; message: string }> {
    return api.post<{ success: boolean; message: string }>(`/api/memories/${id}/unarchive`, {});
  }

  /**
   * Get memory statistics
   */
  async getMemoryStats(): Promise<MemoryStats> {
    return api.get<MemoryStats>('/api/memories/stats');
  }

  /**
   * Manually trigger memory extraction for a conversation
   */
  async triggerMemoryExtraction(
    conversationId: string
  ): Promise<{ success: boolean; message: string }> {
    return api.post<{ success: boolean; message: string }>(
      `/conversations/${conversationId}/extract-memories`,
      {}
    );
  }
}

export default new MemoryService();
