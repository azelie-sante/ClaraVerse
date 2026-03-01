/**
 * Claw Service — API layer for routines, tools, and status
 */

import { api } from './api';

// ============================================================================
// Types
// ============================================================================

export interface ClawStatus {
  telegram: {
    connected: boolean;
    botUsername?: string;
    botName?: string;
  };
  mcp: {
    connected: boolean;
    platform?: string;
    toolCount: number;
  };
  routines: {
    total: number;
    active: number;
    nextRun?: string;
  };
  setupComplete: boolean;
}

export interface RoutineTemplate {
  id: string;
  name: string;
  description: string;
  prompt: string;
  cronExpression: string;
  requiredTools: string[];
  icon: string;
}

export interface Routine {
  id: string;
  userId: string;
  name: string;
  prompt: string;
  cronExpression: string;
  timezone: string;
  enabled: boolean;
  deliveryMethod: 'telegram' | 'store';
  modelId?: string;
  enabledTools: string[];
  template: string;

  // Stats
  totalRuns: number;
  successfulRuns: number;
  failedRuns: number;
  lastRunAt?: string;
  nextRunAt?: string;
  lastResult?: string;

  createdAt: string;
  updatedAt: string;
}

export interface CreateRoutineRequest {
  name: string;
  prompt: string;
  cronExpression: string;
  timezone: string;
  deliveryMethod: 'telegram' | 'store';
  modelId?: string;
  enabledTools?: string[];
  template?: string;
}

export interface UpdateRoutineRequest {
  name?: string;
  prompt?: string;
  cronExpression?: string;
  timezone?: string;
  enabled?: boolean;
  deliveryMethod?: 'telegram' | 'store';
  modelId?: string;
  enabledTools?: string[];
}

export interface RoutineRun {
  id: string;
  status: string;
  mode: string;
  goal: string;
  summary?: string;
  error?: string;
  created_at: string;
  completed_at?: string;
}

export interface TestRoutineRequest {
  name?: string;
  prompt: string;
  modelId?: string;
  enabledTools?: string[];
}

export interface AvailableTool {
  name: string;
  display_name: string;
  description: string;
  icon: string;
  category: string;
  keywords: string[];
  source: string;
}

export interface ToolCategory {
  name: string;
  count: number;
  tools: AvailableTool[];
}

export interface ToolsResponse {
  categories: ToolCategory[];
  total: number;
}

// ============================================================================
// Status
// ============================================================================

export async function getClawStatus(): Promise<ClawStatus> {
  return api.get<ClawStatus>('/api/claras-claw/status');
}

// ============================================================================
// Routines
// ============================================================================

export async function listRoutines(): Promise<Routine[]> {
  const response = await api.get<{ routines: Routine[] }>('/api/routines');
  return response.routines || [];
}

export async function createRoutine(data: CreateRoutineRequest): Promise<Routine> {
  return api.post<Routine>('/api/routines', data);
}

export async function updateRoutine(id: string, data: UpdateRoutineRequest): Promise<Routine> {
  return api.put<Routine>(`/api/routines/${id}`, data);
}

export async function deleteRoutine(id: string): Promise<void> {
  await api.delete(`/api/routines/${id}`);
}

export async function triggerRoutine(id: string): Promise<{ message: string }> {
  return api.post<{ message: string }>(`/api/routines/${id}/trigger`, {});
}

export async function fetchRoutineRuns(routineId: string, limit = 10): Promise<RoutineRun[]> {
  return api.get<RoutineRun[]>(`/api/routines/${routineId}/runs?limit=${limit}`);
}

export async function testRoutine(data: TestRoutineRequest): Promise<string> {
  const response = await api.post<{ result: string }>('/api/routines/test', data);
  return response.result;
}

// ============================================================================
// Tools
// ============================================================================

export async function getAvailableTools(): Promise<ToolsResponse> {
  return api.get<ToolsResponse>('/api/tools/available');
}

// ============================================================================
// Routine Templates
// ============================================================================

export const ROUTINE_TEMPLATES: RoutineTemplate[] = [
  {
    id: 'morning-briefing',
    name: 'Morning Briefing',
    description: 'Get a daily summary every morning with weather, news, and your schedule',
    prompt:
      'Give me a concise morning briefing. Include: 1) A motivational greeting, 2) Any important reminders or tasks I should know about, 3) A brief summary of trending news topics. Keep it friendly and under 500 words.',
    cronExpression: '0 8 * * *',
    requiredTools: [],
    icon: 'sunrise',
  },
  {
    id: 'daily-digest',
    name: 'Daily Digest',
    description: 'End-of-day recap summarizing your activity and preparing for tomorrow',
    prompt:
      'Create an end-of-day digest. Summarize what happened today and suggest priorities for tomorrow. Keep it brief and actionable.',
    cronExpression: '0 18 * * *',
    requiredTools: [],
    icon: 'book-open',
  },
  {
    id: 'reminder',
    name: 'Recurring Reminder',
    description: 'Set up a recurring reminder message',
    prompt: 'Send me a reminder: [Your reminder text here]',
    cronExpression: '0 9 * * 1-5',
    requiredTools: [],
    icon: 'bell',
  },
  {
    id: 'file-watcher',
    name: 'File Watcher Summary',
    description: 'Monitor a directory for changes and get summaries (requires MCP)',
    prompt:
      'Check the filesystem for any recent changes in the project directory. List new or modified files and summarize what changed. Use the filesystem tools to read directory contents.',
    cronExpression: '0 */4 * * *',
    requiredTools: ['list_directory', 'read_file'],
    icon: 'folder-search',
  },
];
