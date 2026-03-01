/**
 * Skill Service
 * API layer for the Skills system — browse, enable, create custom skills
 */

import { api } from './api';

// ============================================================================
// Types
// ============================================================================

export interface Skill {
  id: string;
  name: string;
  description: string;
  icon: string;
  category: string;
  system_prompt: string;
  required_tools: string[];
  preferred_servers: string[];
  keywords: string[];
  trigger_patterns: string[];
  mode: 'auto' | 'manual';
  is_builtin: boolean;
  author_id?: string;
  version: string;
  created_at: string;
  updated_at: string;
}

export interface UserSkillWithDetails {
  id: string;
  user_id: string;
  skill_id: string;
  enabled: boolean;
  created_at: string;
  skill: Skill;
}

export interface CreateSkillRequest {
  name: string;
  description: string;
  icon: string;
  category: string;
  system_prompt: string;
  required_tools: string[];
  preferred_servers: string[];
  keywords: string[];
  trigger_patterns: string[];
  mode: 'auto' | 'manual';
}

interface ListSkillsResponse {
  skills: Skill[];
  categories: Record<string, Skill[]>;
  total: number;
}

interface UserSkillsResponse {
  skills: UserSkillWithDetails[];
  total: number;
}

// ============================================================================
// API Calls
// ============================================================================

export async function listSkills(category?: string): Promise<ListSkillsResponse> {
  const query = category ? `?category=${encodeURIComponent(category)}` : '';
  return api.get<ListSkillsResponse>(`/api/skills${query}`);
}

export async function getSkill(id: string): Promise<Skill> {
  return api.get<Skill>(`/api/skills/${id}`);
}

export async function createSkill(data: CreateSkillRequest): Promise<Skill> {
  return api.post<Skill>('/api/skills', data);
}

export async function updateSkill(id: string, data: CreateSkillRequest): Promise<void> {
  await api.put('/api/skills/' + id, data);
}

export async function deleteSkill(id: string): Promise<void> {
  await api.delete('/api/skills/' + id);
}

export async function getMySkills(): Promise<UserSkillsResponse> {
  return api.get<UserSkillsResponse>('/api/skills/mine');
}

export async function enableSkill(id: string): Promise<void> {
  await api.post('/api/skills/' + id + '/enable', {});
}

export async function disableSkill(id: string): Promise<void> {
  await api.post('/api/skills/' + id + '/disable', {});
}

export async function bulkEnableSkills(skillIds: string[]): Promise<void> {
  await api.post('/api/skills/bulk-enable', { skill_ids: skillIds });
}

// ============================================================================
// Import / Export / Community
// ============================================================================

export interface CommunitySkillEntry {
  name: string;
  description: string;
  repo_url: string;
  raw_url: string;
  author: string;
  license: string;
  category: string;
  icon: string;
}

interface CommunitySkillsResponse {
  skills: CommunitySkillEntry[];
  total: number;
}

interface ExportSkillResponse {
  content: string;
}

export async function importFromSkillMD(content: string): Promise<Skill> {
  return api.post<Skill>('/api/skills/import/skillmd', { content });
}

export async function importFromGitHub(url: string): Promise<Skill> {
  return api.post<Skill>('/api/skills/import/github', { url });
}

export async function exportSkillMD(id: string): Promise<string> {
  const resp = await api.get<ExportSkillResponse>(`/api/skills/${id}/export`);
  return resp.content;
}

export async function listCommunitySkills(): Promise<CommunitySkillsResponse> {
  return api.get<CommunitySkillsResponse>('/api/skills/community');
}
