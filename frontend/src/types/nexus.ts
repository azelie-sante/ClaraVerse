// Nexus Multi-Agent System Types

// ── Task ────────────────────────────────────────────────────────────────

export type NexusTaskStatus =
  | 'draft'
  | 'pending'
  | 'executing'
  | 'waiting_input'
  | 'completed'
  | 'failed'
  | 'cancelled';

export interface NexusTask {
  id: string;
  session_id: string;
  user_id: string;
  parent_task_id?: string;
  project_id?: string;
  prompt: string;
  goal: string;
  priority: number;
  source: string;
  mode: 'quick' | 'daemon' | 'multi_daemon';
  daemon_id?: string;
  status: NexusTaskStatus;
  result?: NexusTaskResult;
  error?: string;
  sub_task_ids?: string[];
  model_id?: string;
  retry_of_task_id?: string;
  manual_retry_count?: number;
  created_at: string;
  started_at?: string;
  completed_at?: string;
  updated_at: string;
}

export interface NexusTaskResult {
  summary: string;
  data?: Record<string, unknown>;
  artifacts?: NexusArtifact[];
}

export interface NexusArtifact {
  type: 'file' | 'image' | 'code' | 'link';
  name: string;
  content: string;
  mime_type?: string;
}

// ── Daemon ──────────────────────────────────────────────────────────────

export type DaemonStatus =
  | 'idle'
  | 'executing'
  | 'waiting_input'
  | 'completed'
  | 'failed'
  | 'cancelled';

export interface Daemon {
  id: string;
  session_id: string;
  user_id: string;
  task_id: string;
  role: string;
  role_label: string;
  task_summary?: string;
  persona: string;
  assigned_tools: string[];
  plan_index: number;
  depends_on?: number[];
  dependency_results?: Record<string, string>;
  status: DaemonStatus;
  current_action?: string;
  progress: number;
  working_memory?: DaemonMemoryEntry[];
  messages?: DaemonMessage[];
  iterations: number;
  max_iterations: number;
  retry_count: number;
  max_retries: number;
  model_id?: string;
  created_at: string;
  started_at?: string;
  completed_at?: string;
}

export interface DaemonMemoryEntry {
  key: string;
  value: string;
  summary: string;
  timestamp: string;
}

export interface DaemonMessage {
  role: 'system' | 'user' | 'assistant' | 'tool';
  content: string;
  tool_call?: DaemonToolCall;
  tool_result?: DaemonToolResult;
  timestamp: string;
}

export interface DaemonToolCall {
  id: string;
  name: string;
  arguments: string;
}

export interface DaemonToolResult {
  tool_call_id: string;
  content: string;
  is_error?: boolean;
}

// ── Daemon Template ─────────────────────────────────────────────────────

export interface DaemonTemplate {
  id: string;
  user_id: string;
  name: string;
  slug: string;
  description: string;
  role: string;
  role_label: string;
  persona: string;
  instructions: string;
  constraints: string;
  output_format: string;
  default_tools: string[];
  icon: string;
  color: string;
  max_iterations: number;
  max_retries: number;
  is_default: boolean;
  is_active: boolean;
  learnings?: TemplateLearning[];
  stats: TemplateStats;
  created_at: string;
  updated_at: string;
}

export interface TemplateLearning {
  key: string;
  content: string;
  category: string;
  confidence: number;
  reinforced_count: number;
  created_at: string;
  last_seen_at: string;
}

export interface TemplateStats {
  total_runs: number;
  successful_runs: number;
  failed_runs: number;
  avg_iterations: number;
}

// ── Persona ─────────────────────────────────────────────────────────────

export interface PersonaFact {
  id: string;
  user_id: string;
  category: 'personality' | 'communication' | 'expertise' | 'boundaries';
  content: string;
  confidence: number;
  source: 'user_explicit' | 'inferred' | 'default';
  reinforced_count: number;
  created_at: string;
  updated_at: string;
}

// ── Session ─────────────────────────────────────────────────────────────

export interface NexusSession {
  id: string;
  user_id: string;
  recent_task_ids?: string[];
  context_summary?: string;
  active_daemon_ids?: string[];
  active_task_ids?: string[];
  pinned_skill_ids?: string[];
  model_id?: string;
  total_tasks: number;
  completed_tasks: number;
  failed_tasks: number;
  created_at: string;
  updated_at: string;
  last_activity_at: string;
}

// ── Engram ───────────────────────────────────────────────────────────────

export interface EngramEntry {
  id: string;
  session_id: string;
  user_id: string;
  type: 'task_result' | 'daemon_output' | 'user_fact' | 'status_log';
  key: string;
  value: string;
  summary: string;
  source: string;
  expires_at?: string;
  created_at: string;
}

// ── WebSocket Messages ──────────────────────────────────────────────────

// Client → Server
export type NexusClientMessageType =
  | 'send_message'
  | 'create_task'
  | 'update_task_status'
  | 'answer_daemon'
  | 'cancel_daemon'
  | 'cancel_all'
  | 'retry_task'
  | 'explain_task'
  | 'pin_skill'
  | 'unpin_skill'
  | 'get_session'
  | 'get_daemon_detail'
  | 'update_persona'
  | 'ping';

export interface NexusClientMessage {
  type: NexusClientMessageType;
  content?: string;
  daemon_id?: string;
  task_id?: string;
  skill_id?: string;
  answer?: string;
  model_id?: string;
  facts?: PersonaFactUpdate[];
  // Task creation fields (for create_task / send_message with config)
  tools?: string[];
  skill_ids?: string[];
  daemon_mode?: 'auto' | 'quick' | 'daemon' | 'multi_daemon';
  template_id?: string;
  priority?: number;
  scheduled_at?: string;
  project_id?: string;
  save_ids?: string[];
  // Status change (for update_task_status)
  status?: NexusTaskStatus;
}

export interface PersonaFactUpdate {
  id?: string;
  category: string;
  content: string;
  action: 'create' | 'update' | 'delete';
  confidence?: number;
}

// Server → Client
export type NexusServerMessageType =
  | 'connected'
  | 'pong'
  | 'session_state'
  | 'error'
  // Cortex
  | 'cortex_thinking'
  | 'cortex_response'
  | 'cortex_classified'
  // Daemons
  | 'daemon_deployed'
  | 'daemon_status'
  | 'daemon_thinking'
  | 'daemon_tool_call'
  | 'daemon_tool_result'
  | 'daemon_completed'
  | 'daemon_failed'
  | 'daemon_question'
  | 'daemon_cancelled'
  | 'daemon_detail'
  // Tasks
  | 'task_created'
  | 'task_updated'
  | 'task_completed'
  | 'task_failed'
  | 'task_status_changed'
  | 'retry_started'
  // Memory
  | 'engram_updated'
  | 'persona_updated'
  // Bridge
  | 'bridge_state_updated'
  // Missed updates (reconnect)
  | 'missed_updates'
  // Skills
  | 'skill_pinned'
  | 'skill_unpinned'
  | 'all_cancelled';

export interface NexusServerMessage {
  type: NexusServerMessageType;
  data?: unknown;
}

// ── Daemon Update (from backend DaemonUpdate struct) ────────────────────

export interface DaemonUpdate {
  daemon_id: string;
  index: number;
  role: string;
  type: string;
  status?: string;
  current_action?: string;
  progress?: number;
  content?: string;
  result?: DaemonResult;
  tool_name?: string;
  tool_result?: string;
  error?: string;
  can_retry?: boolean;
}

export interface DaemonResult {
  summary: string;
  data?: Record<string, unknown>;
  artifacts?: NexusArtifact[];
}

// ── Classification ──────────────────────────────────────────────────────

export interface ClassificationResult {
  mode: 'quick' | 'daemon' | 'multi_daemon';
  daemons_planned?: DaemonPlan[];
}

export interface DaemonPlan {
  index: number;
  role: string;
  role_label: string;
  persona: string;
  task_summary: string;
  tools_needed: string[];
  depends_on: number[];
}

// ── Session State (from get_session response) ───────────────────────────

export interface NexusSessionState {
  session: NexusSession;
  active_daemons: Daemon[];
  recent_tasks: NexusTask[];
  persona: PersonaFact[];
  recent_engrams: EngramEntry[];
}

// ── Conversation Item (for display) ─────────────────────────────────────

export interface NexusConversationItem {
  id: string;
  type:
    | 'user_message'
    | 'cortex_response'
    | 'cortex_thinking'
    | 'daemon_activity'
    | 'task_result'
    | 'error';
  content: string;
  timestamp: Date;
  daemonId?: string;
  daemonRole?: string;
  taskId?: string;
  toolName?: string;
  toolResult?: string;
}

// ── Project ─────────────────────────────────────────────────────────

export interface NexusProject {
  id: string;
  user_id: string;
  name: string;
  description?: string;
  system_instruction?: string;
  icon: string;
  color: string;
  is_archived: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

// ── Save ──────────────────────────────────────────────────────────────

export interface NexusSave {
  id: string;
  user_id: string;
  title: string;
  content: string;
  tags?: string[];
  source_task_id?: string;
  source_project_id?: string;
  created_at: string;
  updated_at: string;
}

// ── Missed Updates (while you were away) ─────────────────────────────

export interface MissedUpdate {
  id: string;
  event_type: string;
  task_id?: string;
  goal?: string;
  summary: string;
  content: string;
  timestamp: string;
}
