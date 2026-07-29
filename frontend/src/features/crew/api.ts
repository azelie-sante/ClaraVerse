/**
 * Nexus v2 ("Crew") API client — projects → team members → the card pipeline.
 * New path, fully separate from the legacy Nexus.
 */
import { getApiBaseUrl } from '@/lib/config';
import { authClient } from '@/lib/auth';

export interface CrewRole {
  key: string;
  label: string;
  blurb: string;
  allowed_tools: string[];
  default_tools: string[];
}
export interface CrewTemplate {
  key: string;
  name: string;
  blurb: string;
  goal: string;
  objectives: string[];
  members: { role_key: string; name: string; charter: string }[];
  cards: { title: string; detail: string; obj_idx: number; member_idx: number[]; repeat?: string }[];
}
export interface CrewObjective {
  id: string;
  title: string;
  done: boolean;
}
export interface CrewProject {
  id: string;
  name: string;
  brief: string;
  goal?: string;
  objectives?: CrewObjective[];
  status: string;
  created_at: string;
  updated_at: string;
}
export interface CrewMember {
  id: string;
  project_id: string;
  role_key: string;
  name: string;
  charter?: string;
  monthly_budget?: number;
  month_tokens?: number;
  tools: string[];
  model?: string;
  documents?: { name: string; size: number; at: string }[];
  skill_ids?: string[];
  status: 'active' | 'paused';
  tokens_in: number;
  tokens_out: number;
}
export interface CardRevision {
  member_id?: string;
  member_name?: string;
  output: string;
  tools_used?: string[];
  review?: string;
  approved: boolean;
  tokens_used?: number;
  at: string;
}
export interface CrewCard {
  id: string;
  project_id: string;
  title: string;
  detail?: string;
  assignee_id?: string;
  assignee_ids?: string[];
  objective_id?: string;
  repeat?: string;
  next_run_at?: string;
  depends_on?: string[];
  parts?: Record<string, string>;
  awaiting?: string[];
  running?: string[];
  status: 'draft' | 'queued' | 'working' | 'review' | 'done';
  latest_output?: string;
  revisions?: CardRevision[];
  error?: string;
  created_at: string;
  updated_at: string;
}

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const token = authClient.getAccessToken();
  const res = await fetch(`${getApiBaseUrl()}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(init?.headers ?? {}),
    },
  });
  if (!res.ok) {
    const body = (await res.json().catch(() => ({}))) as { error?: string };
    throw new Error(body.error || `Request failed (${res.status})`);
  }
  return res.json() as Promise<T>;
}

export interface CrewSkill {
  id: string;
  name: string;
  description: string;
  icon?: string;
  category?: string;
  required_tools?: string[];
}

export const crewApi = {
  roles: () => req<{ roles: CrewRole[] }>('/api/crew/roles').then((r) => r.roles),
  skills: () => req<{ skills: CrewSkill[] }>('/api/crew/skills').then((r) => r.skills),
  setMemberSkills: (memberId: string, skillIds: string[]) =>
    req(`/api/crew/members/${memberId}/skills`, { method: 'PUT', body: JSON.stringify({ skill_ids: skillIds }) }),
  listProjects: () => req<{ projects: CrewProject[] }>('/api/crew/projects').then((r) => r.projects),
  createProject: (name: string, brief: string) =>
    req<CrewProject>('/api/crew/projects', { method: 'POST', body: JSON.stringify({ name, brief }) }),
  getBoard: (id: string) =>
    req<{ project: CrewProject; members: CrewMember[]; cards: CrewCard[] }>(`/api/crew/projects/${id}`),
  hire: (projectId: string, data: { role_key: string; name?: string; tools: string[]; model?: string }) =>
    req<CrewMember>(`/api/crew/projects/${projectId}/members`, { method: 'POST', body: JSON.stringify(data) }),
  setMemberStatus: (memberId: string, status: 'active' | 'paused') =>
    req(`/api/crew/members/${memberId}/status`, { method: 'PUT', body: JSON.stringify({ status }) }),
  fire: (memberId: string) => req(`/api/crew/members/${memberId}`, { method: 'DELETE' }),
  updateProject: (projectId: string, data: { name?: string; brief?: string; goal?: string; objectives?: CrewObjective[] }) =>
    req<{ success: boolean }>(`/api/crew/projects/${projectId}`, { method: 'PUT', body: JSON.stringify(data) }),
  archiveProject: (projectId: string) =>
    req<{ success: boolean }>(`/api/crew/projects/${projectId}`, { method: 'PUT', body: JSON.stringify({ status: 'archived' }) }),
  setMemberCharter: (memberId: string, charter: string) =>
    req<{ success: boolean }>(`/api/crew/members/${memberId}/charter`, { method: 'PUT', body: JSON.stringify({ charter }) }),
  templates: () => req<{ templates: CrewTemplate[] }>('/api/crew/templates').then((r) => r.templates),
  createFromTemplate: (template: string, name?: string, brief?: string) =>
    req<CrewProject>('/api/crew/projects/from-template', { method: 'POST', body: JSON.stringify({ template, name, brief }) }),
  planCard: (cardId: string) =>
    req<{ cards: CrewCard[] }>(`/api/crew/cards/${cardId}/plan`, { method: 'POST' }),
  planProject: (projectId: string) =>
    req<{ cards: CrewCard[] }>(`/api/crew/projects/${projectId}/plan`, { method: 'POST' }),
  setMemberBudget: (memberId: string, monthlyBudget: number) =>
    req<{ success: boolean }>(`/api/crew/members/${memberId}/budget`, { method: 'PUT', body: JSON.stringify({ monthly_budget: monthlyBudget }) }),
  createCard: (projectId: string, data: { title: string; detail?: string; objective_id?: string; repeat?: string; assignee_ids?: string[]; depends_on?: string[] }) =>
    req<CrewCard>(`/api/crew/projects/${projectId}/cards`, { method: 'POST', body: JSON.stringify(data) }),
  updateCard: (cardId: string, data: { title?: string; detail?: string; assignee_ids?: string[] }) =>
    req(`/api/crew/cards/${cardId}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteCard: (cardId: string) => req(`/api/crew/cards/${cardId}`, { method: 'DELETE' }),
  queueCard: (cardId: string) => req(`/api/crew/cards/${cardId}/queue`, { method: 'POST' }),
  unqueueCard: (cardId: string) => req(`/api/crew/cards/${cardId}/unqueue`, { method: 'POST' }),

  /** Attach a reference document to a member (parsed server-side → RAG-in-hand). */
  async uploadMemberDoc(memberId: string, file: File): Promise<void> {
    const token = authClient.getAccessToken();
    const fd = new FormData();
    fd.append('file', file);
    const res = await fetch(`${getApiBaseUrl()}/api/crew/members/${memberId}/docs`, {
      method: 'POST',
      headers: token ? { Authorization: `Bearer ${token}` } : {},
      body: fd,
    });
    if (!res.ok) {
      const body = (await res.json().catch(() => ({}))) as { error?: string };
      throw new Error(body.error || 'Upload failed');
    }
  },
  deleteMemberDoc: (memberId: string, index: number) =>
    req(`/api/crew/members/${memberId}/docs/${index}`, { method: 'DELETE' }),
  reviewCard: (cardId: string, approve: boolean, feedback: string) =>
    req(`/api/crew/cards/${cardId}/review`, { method: 'POST', body: JSON.stringify({ approve, feedback }) }),

  /** Download a card's output as a styled PDF report. */
  async downloadPdf(cardId: string, title: string, rev?: number): Promise<void> {
    const token = authClient.getAccessToken();
    const res = await fetch(`${getApiBaseUrl()}/api/crew/cards/${cardId}/pdf${rev ? `?rev=${rev}` : ''}`, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    });
    if (!res.ok) throw new Error('PDF export failed');
    const blob = await res.blob();
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${title.replace(/[^a-zA-Z0-9 _-]+/g, '').slice(0, 60) || 'crew-report'}.pdf`;
    document.body.appendChild(a);
    a.click();
    a.remove();
    setTimeout(() => URL.revokeObjectURL(url), 5000);
  },
};
