/**
 * Crew project board — shared app sidebar, only the main area changes.
 * No modals: everything opens as a right-side slide-over (80% width for card
 * detail/review, 50% for creation/hire/member). Cards drag between Drafts and
 * Queued; the pipeline (working → review → done) is driven by agents + review.
 */
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Check, ChevronRight, FileDown, FileText, Loader2, Paperclip, Pause, Play, Plus, RotateCcw, Send, Sparkles, Target, Trash2, UserPlus, Users, X,
} from 'lucide-react';
import { MarkdownRenderer } from '@/components/design-system/content/MarkdownRenderer/MarkdownRenderer';
import { CrewSidebar } from '@/features/crew/ui/CrewSidebar';
import { CrewHelp } from '@/features/crew/ui/CrewHelp';
import { crewApi, type CrewCard, type CrewMember, type CrewObjective, type CrewProject as Project, type CrewRole, type CrewSkill } from '@/features/crew/api';

const COLUMNS: { key: CrewCard['status']; label: string; dot: string; pill: string; empty: string }[] = [
  { key: 'draft', label: 'Drafts', dot: 'bg-zinc-500', pill: 'bg-zinc-500/10 text-zinc-400', empty: 'Create a card to plan work' },
  { key: 'queued', label: 'Queued', dot: 'bg-[var(--color-accent)]', pill: 'bg-[var(--color-accent)]/12 text-[var(--color-accent)]', empty: 'Drag drafts here to start' },
  { key: 'working', label: 'Working', dot: 'bg-amber-400 animate-pulse', pill: 'bg-amber-500/10 text-amber-400', empty: 'Nothing running' },
  { key: 'review', label: 'Review', dot: 'bg-orange-400', pill: 'bg-orange-500/10 text-orange-400', empty: 'Finished work lands here' },
  { key: 'done', label: 'Done', dot: 'bg-emerald-400', pill: 'bg-emerald-500/10 text-emerald-400', empty: 'Approve reviews to fill this' },
];

/** Deterministic member avatar hue from its id. */
function avatarHue(id: string): number {
  let h = 0;
  for (let i = 0; i < id.length; i++) h = (h * 31 + id.charCodeAt(i)) % 360;
  return h;
}
function initials(name: string): string {
  const parts = name.trim().split(/\s+/);
  return ((parts[0]?.[0] ?? '') + (parts[1]?.[0] ?? '')).toUpperCase() || '?';
}
function relTime(iso?: string): string {
  if (!iso) return '';
  const s = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
  if (s < 60) return 'just now';
  if (s < 3600) return `${Math.floor(s / 60)}m`;
  if (s < 86400) return `${Math.floor(s / 3600)}h`;
  return `${Math.floor(s / 86400)}d`;
}

// Group tools by integration for the hire picker; flag outward-acting ones.
const TOOL_GROUPS: { label: string; match: (t: string) => boolean }[] = [
  { label: 'Research & web', match: (t) => ['search_web', 'scrape_web', 'search_images', 'api_request', 'download_file', 'get_current_time'].includes(t) },
  { label: 'Files & analysis', match: (t) => ['read_document', 'read_data_file', 'read_spreadsheet', 'run_python', 'analyze_data', 'calculate_math', 'transcribe_audio', 'describe_image'].includes(t) },
  { label: 'Create & export', match: (t) => ['create_document', 'create_presentation', 'create_text_file', 'html_to_pdf', 'generate_image', 'edit_image'].includes(t) },
  { label: 'Gmail', match: (t) => t.startsWith('gmail_') },
  { label: 'Calendar', match: (t) => t.startsWith('googlecalendar_') || t.startsWith('calendly_') },
  { label: 'Drive & Sheets', match: (t) => t.startsWith('googledrive_') || t.startsWith('googlesheets_') },
  { label: 'Messaging', match: (t) => t.startsWith('send_') || t.startsWith('twilio_') },
  { label: 'Social', match: (t) => t.startsWith('x_') || t.startsWith('linkedin_') || t.startsWith('youtube_') },
  { label: 'Project tools', match: (t) => ['notion_', 'trello_', 'clickup_', 'jira_', 'linear_', 'airtable_', 'github_'].some((p) => t.startsWith(p)) },
  { label: 'CRM & analytics', match: (t) => ['hubspot_', 'leadsquared_', 'shopify_', 'mailchimp_', 'google_analytics', 'google_search_console', 'google_ads', 'facebook_ads_'].some((p) => t.startsWith(p)) },
  { label: 'Meetings & design', match: (t) => t.startsWith('zoom_') || t.startsWith('canva_') },
];
const LOCAL_SAFE = new Set(['create_document', 'create_presentation', 'create_text_file', 'generate_image', 'edit_image', 'html_to_pdf']);
/** Tools that ACT on external systems (send/post/modify) — badge + caution. */
function actsExternally(t: string): boolean {
  if (LOCAL_SAFE.has(t)) return false;
  return /(send|post_tweet|_create|create_|_update|update_|_delete|delete_|reply|trash|subscribe|upsert|append|write|clear|add_label|add_subscriber|find_replace|add_sheet)/.test(t);
}
function groupTools(tools: string[]): { label: string; tools: string[] }[] {
  const used = new Set<string>();
  const groups = TOOL_GROUPS.map((g) => {
    const list = tools.filter((t) => !used.has(t) && g.match(t));
    list.forEach((t) => used.add(t));
    return { label: g.label, tools: list };
  }).filter((g) => g.tools.length > 0);
  const rest = tools.filter((t) => !used.has(t));
  if (rest.length) groups.push({ label: 'Other', tools: rest });
  return groups;
}

type Panel =
  | { kind: 'card'; cardId: string }
  | { kind: 'new-card' }
  | { kind: 'hire' }
  | { kind: 'member'; memberId: string }
  | { kind: 'goal' }
  | { kind: 'help' }
  | null;

export default function CrewProjectPage() {
  const { projectId = '' } = useParams();
  const navigate = useNavigate();
  const [project, setProject] = useState<Project | null>(null);
  const [members, setMembers] = useState<CrewMember[]>([]);
  const [cards, setCards] = useState<CrewCard[]>([]);
  const [roles, setRoles] = useState<CrewRole[]>([]);
  const [recents, setRecents] = useState<Project[]>([]);
  const [error, setError] = useState('');
  const [panel, setPanel] = useState<Panel>(null);
  const [dragOver, setDragOver] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      const b = await crewApi.getBoard(projectId);
      setProject(b.project);
      setMembers(b.members);
      setCards(b.cards);
      setError('');
    } catch (e) {
      setError((e as Error).message);
    }
  }, [projectId]);

  useEffect(() => {
    void load();
    void crewApi.roles().then(setRoles).catch(() => {});
    void crewApi.listProjects().then((p) => setRecents(p.slice(0, 5))).catch(() => {});
  }, [load]);

  // Live-poll while the pipeline is moving (or the tab is visible).
  const activeRef = useRef(false);
  activeRef.current = cards.some((c) => c.status === 'queued' || c.status === 'working');
  useEffect(() => {
    const t = setInterval(() => {
      if (activeRef.current || document.visibilityState === 'visible') void load();
    }, 8000);
    // Agents run SERVER-SIDE regardless of this tab — refresh instantly when
    // the user returns so completed work appears immediately, not after a poll.
    const onFocus = () => void load();
    window.addEventListener('focus', onFocus);
    document.addEventListener('visibilitychange', onFocus);
    return () => {
      clearInterval(t);
      window.removeEventListener('focus', onFocus);
      document.removeEventListener('visibilitychange', onFocus);
    };
  }, [load]);

  const memberName = useCallback((id?: string) => members.find((m) => m.id === id)?.name ?? 'Unassigned', [members]);
  const cardAssignees = useCallback(
    (c: CrewCard): string[] => (c.assignee_ids?.length ? c.assignee_ids : c.assignee_id ? [c.assignee_id] : []),
    []
  );
  const byStatus = useMemo(() => {
    const m: Record<string, CrewCard[]> = {};
    for (const c of COLUMNS) m[c.key] = [];
    for (const c of cards) (m[c.status] ??= []).push(c);
    return m;
  }, [cards]);
  const reviewCount = byStatus['review']?.length ?? 0;
  const openCard = panel?.kind === 'card' ? (cards.find((c) => c.id === panel.cardId) ?? null) : null;
  const openMember = panel?.kind === 'member' ? (members.find((m) => m.id === panel.memberId) ?? null) : null;

  // Drag & drop transitions: draft→queued (start), queued→draft (pull back),
  // review→done (approve). Everything else is agent/review-driven.
  const draggingRef = useRef(false);
  const onDropTo = useCallback(
    async (col: string, cardId: string) => {
      setDragOver(null);
      const card = cards.find((c) => c.id === cardId);
      if (!card || card.status === col) return;
      try {
        if (card.status === 'draft' && col === 'queued') await crewApi.queueCard(cardId);
        else if (card.status === 'queued' && col === 'draft') await crewApi.unqueueCard(cardId);
        else if (card.status === 'review' && col === 'done') await crewApi.reviewCard(cardId, true, '');
        else {
          setError(
            card.status === 'draft' || card.status === 'queued' || card.status === 'review'
              ? `Cards can move: Drafts ⇄ Queued, and Review → Done (approve)`
              : `${card.status} cards are moved by the agents and your reviews`
          );
          setTimeout(() => setError(''), 3500);
          return;
        }
        await load();
      } catch (e) {
        setError((e as Error).message);
        setTimeout(() => setError(''), 4500);
      }
    },
    [cards, load]
  );

  if (error && !project) {
    return (
      <div className="h-screen flex bg-[var(--color-background)] text-[var(--color-text-primary)]">
        <CrewSidebar onNewProject={() => navigate('/crew')} onProjects={() => navigate('/crew')} onHelp={() => setPanel({ kind: 'help' })} />
        <div className="flex-1 flex items-center justify-center">
          <div className="text-center space-y-3">
            <p className="text-[var(--color-text-secondary)]">{error}</p>
            <button onClick={() => navigate('/crew')} className="px-4 py-2 rounded-lg border border-[var(--color-border)] hover:bg-[var(--color-surface-hover)]">
              Back to projects
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="h-screen flex bg-[var(--color-background)] text-[var(--color-text-primary)] overflow-hidden">
      {/* Shared app sidebar — same component as Chat / Builder / Settings. */}
      <CrewSidebar
        onNewProject={() => navigate('/crew')}
        onProjects={() => navigate('/crew')}
        onHelp={() => setPanel({ kind: 'help' })}
        recents={recents.map((p) => ({ id: p.id, title: p.name, onClick: () => navigate(`/crew/${p.id}`) }))}
      />

      {/* Main area */}
      <div className="flex-1 flex flex-col min-w-0">
        <header className="flex items-center gap-3 px-5 py-3 bg-[var(--color-surface)]/25 backdrop-blur-xl shrink-0">
          <div className="min-w-0">
            <h1 className="text-[15px] font-semibold tracking-tight truncate">{project?.name ?? '…'}</h1>
            {project?.goal ? (
              <button onClick={() => setPanel({ kind: 'goal' })} className="flex items-center gap-1.5 text-[11px] text-[var(--color-text-tertiary)] hover:text-[var(--color-text-secondary)] truncate max-w-2xl">
                <Target size={11} className="shrink-0 text-[var(--color-accent)]" />
                <span className="truncate">{project.goal}</span>
              </button>
            ) : (
              <p className="text-[11px] text-[var(--color-text-tertiary)] truncate max-w-2xl">{project?.brief}</p>
            )}
          </div>
          <div className="ml-auto flex items-center gap-2 shrink-0">
            {error && <span className="text-xs text-red-400 max-w-xs truncate">{error}</span>}
            {reviewCount > 0 && (
              <span className="text-xs px-2.5 py-1 rounded-full bg-amber-500/15 text-amber-400">{reviewCount} awaiting review</span>
            )}
            <button
              onClick={() => setPanel({ kind: 'goal' })}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-[var(--color-border)] text-sm hover:bg-[var(--color-surface-hover)]"
            >
              <Target size={14} /> Goal
            </button>
            <button
              onClick={() => setPanel({ kind: 'hire' })}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-[var(--color-border)] text-sm hover:bg-[var(--color-surface-hover)]"
            >
              <UserPlus size={14} /> Hire
            </button>
            <button
              onClick={() => setPanel({ kind: 'new-card' })}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-[var(--color-accent)] text-white text-sm font-medium hover:opacity-90"
            >
              <Plus size={15} /> New card
            </button>
          </div>
        </header>

        <div className="flex-1 flex min-h-0">
          {/* Team rail */}
          <aside className="w-56 shrink-0 bg-[var(--color-surface)]/30 backdrop-blur-xl p-3.5 overflow-auto">
            <h2 className="text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--color-text-secondary)] mb-2.5 flex items-center gap-1.5">
              <Users size={12} /> Team <span className="text-[var(--color-text-tertiary)] font-normal">· {members.length}</span>
            </h2>
            <div className="space-y-1.5">
              {members.map((m) => {
                const roleLabel = roles.find((r) => r.key === m.role_key)?.label ?? m.role_key;
                const busyNow = cards.some((c) => c.status === 'working' && (c.running ?? []).includes(m.id));
                return (
                  <button
                    key={m.id}
                    onClick={() => setPanel({ kind: 'member', memberId: m.id })}
                    className="w-full text-left rounded-lg p-2 hover:bg-[var(--color-surface-hover)] transition flex items-center gap-2.5"
                  >
                    <span
                      className="w-8 h-8 rounded-full flex items-center justify-center text-[11px] font-semibold shrink-0"
                      style={{ background: `hsl(${avatarHue(m.id)} 45% 22%)`, color: `hsl(${avatarHue(m.id)} 70% 75%)` }}
                    >
                      {initials(m.name)}
                    </span>
                    <span className="min-w-0 flex-1">
                      <span className="block text-[13px] font-medium truncate">{m.name}</span>
                      <span className="block text-[11px] text-[var(--color-text-tertiary)] truncate">
                        {busyNow ? 'working…' : m.name === roleLabel ? (m.status === 'active' ? 'ready' : 'paused') : roleLabel}
                        {(m.documents?.length ?? 0) > 0 && ` · ${m.documents!.length} docs`}
                      </span>
                    </span>
                    <span className={`w-2 h-2 rounded-full shrink-0 ${busyNow ? 'bg-amber-400 animate-pulse' : m.status === 'active' ? 'bg-emerald-400' : 'bg-zinc-500'}`} />
                  </button>
                );
              })}
              {members.length === 0 && (
                <p className="text-xs text-[var(--color-text-tertiary)]">No team yet — hit Hire to add your first member.</p>
              )}
            </div>
          </aside>

          {/* Board */}
          <main className="flex-1 overflow-x-auto p-4">
            <div className="flex gap-3 h-full pr-2">
              {COLUMNS.map((col) => {
                const list = byStatus[col.key] ?? [];
                return (
                  <section
                    key={col.key}
                    onDragOver={(e) => {
                      e.preventDefault();
                      e.dataTransfer.dropEffect = 'move';
                      setDragOver(col.key);
                    }}
                    onDragLeave={() => setDragOver((d) => (d === col.key ? null : d))}
                    onDrop={(e) => {
                      e.preventDefault();
                      const id = e.dataTransfer.getData('text/crew-card') || e.dataTransfer.getData('text/plain');
                      if (id) void onDropTo(col.key, id);
                    }}
                    className={`flex-1 flex flex-col min-w-[190px] max-w-[320px] rounded-xl transition-colors ${
                      dragOver === col.key ? 'bg-[var(--color-accent)]/8 ring-1 ring-[var(--color-accent)]/40' : ''
                    }`}
                  >
                    <div className="px-3 pt-3 pb-2">
                      <div className="flex items-center gap-2">
                        <span className={`w-1.5 h-1.5 rounded-full ${col.dot}`} />
                        <span className="text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--color-text-secondary)]">{col.label}</span>
                        <span className={`ml-auto text-[10px] px-1.5 py-0.5 rounded-full tabular-nums ${col.pill}`}>{list.length}</span>
                      </div>
                    </div>
                    <div className="flex-1 overflow-auto px-2 pb-2 space-y-2">
                      {list.map((c) => {
                        const assignees = cardAssignees(c);
                        const draggable = c.status === 'draft' || c.status === 'queued' || c.status === 'review';
                        return (
                          <div
                            key={c.id}
                            draggable={draggable}
                            onDragStart={(e) => {
                              draggingRef.current = true;
                              e.dataTransfer.effectAllowed = 'move';
                              e.dataTransfer.setData('text/crew-card', c.id);
                              e.dataTransfer.setData('text/plain', c.id); // some engines require a plain payload
                            }}
                            onDragEnd={() => {
                              setDragOver(null);
                              setTimeout(() => (draggingRef.current = false), 50);
                            }}
                            onClick={() => {
                              if (draggingRef.current) return; // a drag just ended — don't open the panel
                              setPanel({ kind: 'card', cardId: c.id });
                            }}
                            className={`group rounded-lg bg-[var(--color-surface)] shadow-sm shadow-black/20 ring-1 ring-white/5 p-3 transition hover:ring-[var(--color-accent)]/50 hover:shadow-lg hover:shadow-black/25 hover:-translate-y-px ${
                              draggable ? 'cursor-grab active:cursor-grabbing' : 'cursor-pointer'
                            }`}
                          >
                            <div className="text-[13px] font-medium leading-snug line-clamp-3 text-[var(--color-text-primary)]">{c.title}</div>
                            {(c.depends_on ?? []).length > 0 && (() => {
                              const blocked = (c.depends_on ?? []).some((d) => cards.find((x) => x.id === d && x.status !== 'done'));
                              return (
                                <span className={`mt-1.5 mr-1.5 inline-flex items-center gap-1 text-[9.5px] px-1.5 py-0.5 rounded-full ${
                                  blocked && (c.status === 'queued' || c.status === 'draft')
                                    ? 'bg-amber-500/10 text-amber-400'
                                    : 'bg-[var(--color-surface-hover)] text-[var(--color-text-tertiary)]'
                                }`} title={(c.depends_on ?? []).map((d) => cards.find((x) => x.id === d)?.title ?? '?').join(', ')}>
                                  <ChevronRight size={8} /> {blocked && c.status === 'queued' ? 'waiting on dependency' : `after ${(c.depends_on ?? []).length}`}
                                </span>
                              );
                            })()}
                            {c.repeat && (
                              <span className="mt-1.5 mr-1.5 inline-flex items-center gap-1 text-[9.5px] px-1.5 py-0.5 rounded-full bg-[var(--color-surface-hover)] text-[var(--color-text-tertiary)]">
                                <RotateCcw size={8} /> {c.repeat}
                              </span>
                            )}
                            {c.objective_id && project?.objectives?.some((o) => o.id === c.objective_id) && (
                              <div className="mt-1.5 inline-flex items-center gap-1 text-[9.5px] px-1.5 py-0.5 rounded-full bg-[var(--color-accent)]/8 text-[var(--color-accent)]/80 max-w-full">
                                <Target size={8} className="shrink-0" />
                                <span className="truncate">{project.objectives.find((o) => o.id === c.objective_id)!.title}</span>
                              </div>
                            )}
                            {c.error && <div className="mt-1.5 text-[11px] text-red-400 line-clamp-2">{c.error}</div>}
                            <div className="flex items-center gap-1.5 mt-2.5">
                              <div className="flex -space-x-1.5">
                                {assignees.slice(0, 3).map((id) => (
                                  <span
                                    key={id}
                                    title={memberName(id)}
                                    className="w-5 h-5 rounded-full border border-[var(--color-background)] flex items-center justify-center text-[8px] font-semibold"
                                    style={{ background: `hsl(${avatarHue(id)} 45% 22%)`, color: `hsl(${avatarHue(id)} 70% 75%)` }}
                                  >
                                    {initials(memberName(id))}
                                  </span>
                                ))}
                                {assignees.length === 0 && (
                                  <span className="text-[10px] text-[var(--color-text-tertiary)]">unassigned</span>
                                )}
                              </div>
                              <span className="ml-auto flex items-center gap-1.5 text-[10px] text-[var(--color-text-tertiary)] tabular-nums">
                                {c.status === 'working' && <Loader2 size={10} className="animate-spin text-amber-400" />}
                                {(c.revisions?.length ?? 0) > 0 && (
                                  <span className="px-1 py-px rounded bg-[var(--color-surface-hover)]">r{c.revisions!.length}</span>
                                )}
                                {relTime(c.updated_at)}
                              </span>
                            </div>
                          </div>
                        );
                      })}
                      {list.length === 0 &&
                        (dragOver === col.key ? (
                          <div className="mx-1 mt-1 rounded-lg border border-dashed border-[var(--color-accent)]/70 px-3 py-5 text-center text-[10.5px] text-[var(--color-accent)]">
                            {col.empty}
                          </div>
                        ) : (
                          <p className="px-2 pt-1 text-[10.5px] text-[var(--color-text-tertiary)]/50">{col.empty}</p>
                        ))}
                    </div>
                  </section>
                );
              })}
            </div>
          </main>
        </div>
      </div>

      {/* Slide-overs (no modals): 80% for card detail, 50% for creation. */}
      {openCard && (
        <SlideOver widthPct={80} onClose={() => setPanel(null)}>
          <CardPanel card={openCard} members={members} onClose={() => setPanel(null)} onChanged={load} />
        </SlideOver>
      )}
      {panel?.kind === 'new-card' && (
        <SlideOver widthPct={50} onClose={() => setPanel(null)}>
          <NewCardPanel projectId={projectId} project={project} members={members} cards={cards} onClose={() => setPanel(null)} onCreated={load} />
        </SlideOver>
      )}
      {panel?.kind === 'hire' && (
        <SlideOver widthPct={50} onClose={() => setPanel(null)}>
          <HirePanel projectId={projectId} roles={roles} onClose={() => setPanel(null)} onHired={load} />
        </SlideOver>
      )}
      {openMember && (
        <SlideOver widthPct={50} onClose={() => setPanel(null)}>
          <MemberPanel member={openMember} roles={roles} onClose={() => setPanel(null)} onChanged={load} />
        </SlideOver>
      )}
      {panel?.kind === 'goal' && project && (
        <SlideOver widthPct={50} onClose={() => setPanel(null)}>
          <GoalPanel project={project} onClose={() => setPanel(null)} onChanged={load} />
        </SlideOver>
      )}
      {panel?.kind === 'help' && <CrewHelp onClose={() => setPanel(null)} />}
    </div>
  );
}

// ─── SlideOver shell ─────────────────────────────────────────────────────────

function SlideOver({ widthPct, onClose, children }: { widthPct: number; onClose: () => void; children: React.ReactNode }) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onClose]);
  return (
    <div className="fixed inset-0 z-40" onClick={onClose}>
      <div className="absolute inset-0 bg-black/50 backdrop-blur-[2px]" />
      <div
        className="absolute inset-y-0 right-0 frost-panel !border-y-0 !border-r-0 shadow-2xl flex flex-col crew-slide-in"
        style={{ width: `${widthPct}%`, minWidth: 380 }}
        onClick={(e) => e.stopPropagation()}
      >
        {children}
      </div>
    </div>
  );
}

// ─── Card panel (80%) — detail, revisions, the review gate ───────────────────

function CardPanel({ card, members, onClose, onChanged }: { card: CrewCard; members: CrewMember[]; onClose: () => void; onChanged: () => Promise<void> }) {
  const [feedback, setFeedback] = useState('');
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState('');
  // Member parts collapse by default — the combined "Team deliverable" (or the
  // latest revision when there is no combine) is what the reviewer reads.
  const [openRevs, setOpenRevs] = useState<Record<number, boolean>>({});
  const revDefaultOpen = (i: number): boolean => {
    const revs = card.revisions ?? [];
    if (revs[i]?.member_name === 'Team deliverable') return i === revs.length - 1 || revs.slice(i + 1).every((x) => x.member_name !== 'Team deliverable');
    if (revs.some((x) => x.member_name === 'Team deliverable')) return false;
    return i === revs.length - 1;
  };
  const isRevOpen = (i: number) => openRevs[i] ?? revDefaultOpen(i);
  const names = (card.assignee_ids?.length ? card.assignee_ids : card.assignee_id ? [card.assignee_id] : [])
    .map((id) => members.find((m) => m.id === id)?.name ?? '?')
    .join(', ');

  const act = async (fn: () => Promise<unknown>) => {
    setBusy(true);
    setErr('');
    try {
      await fn();
      await onChanged();
    } catch (e) {
      setErr((e as Error).message);
    } finally {
      setBusy(false);
    }
  };

  return (
    <>
      <div className="flex items-start justify-between gap-3 px-6 py-4 border-b border-[var(--color-border)] shrink-0">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <span className={`text-[10px] uppercase tracking-wide px-2 py-0.5 rounded-full ${
              card.status === 'review' ? 'bg-amber-500/15 text-amber-400'
              : card.status === 'done' ? 'bg-emerald-500/15 text-emerald-400'
              : 'bg-[var(--color-surface-hover)] text-[var(--color-text-tertiary)]'}`}>
              {card.status}
            </span>
            <span className="text-xs text-[var(--color-text-tertiary)]">{names || 'Unassigned'}</span>
          </div>
          <h2 className="font-semibold mt-1.5 text-lg">{card.title}</h2>
          {card.detail && <p className="text-sm text-[var(--color-text-secondary)] mt-1 whitespace-pre-wrap">{card.detail}</p>}
        </div>
        <button onClick={onClose} className="text-[var(--color-text-tertiary)] hover:text-[var(--color-text-primary)] shrink-0">
          <X size={18} />
        </button>
      </div>

      <div className="flex-1 overflow-auto px-6 py-4 space-y-4">
        {card.error && <div className="rounded-lg bg-red-500/10 border border-red-500/20 px-3 py-2 text-sm text-red-400">{card.error}</div>}
        {card.parts && Object.keys(card.parts).length > 0 && (
          <div className="rounded-lg bg-[var(--color-background)]/50 p-3">
            <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--color-text-secondary)] mb-1.5">How the team lead split this task</div>
            <div className="space-y-1">
              {Object.entries(card.parts).map(([mid, brief]) => (
                <div key={mid} className="text-[11.5px] text-[var(--color-text-tertiary)] leading-relaxed">
                  <span className="text-[var(--color-text-secondary)] font-medium">{members.find((m) => m.id === mid)?.name ?? '?'}:</span> {brief}
                </div>
              ))}
            </div>
            <p className="text-[10px] text-[var(--color-text-tertiary)]/70 mt-1.5">Each member does only their part; the parts are then combined into the single "Team deliverable" you review.</p>
          </div>
        )}
        {!card.parts && (card.assignee_ids?.length ?? 0) > 1 && (card.revisions?.length ?? 0) > 1 && (
          <p className="text-[11px] text-[var(--color-text-tertiary)]">
            {card.assignee_ids!.length} teammates were assigned, so each submitted their own independent take below — compare them, approve the card, or send it back with feedback (everyone revises).
          </p>
        )}
        {(card.revisions ?? []).map((r, i) => (
          <div key={i} className="rounded-lg border border-[var(--color-border)] overflow-hidden">
            <button
              onClick={() => setOpenRevs((p) => ({ ...p, [i]: !isRevOpen(i) }))}
              className="w-full px-4 py-2 bg-[var(--color-background)] text-xs text-[var(--color-text-tertiary)] flex items-center justify-between hover:bg-[var(--color-surface-hover)] transition-colors"
            >
              <span className="flex items-center gap-1.5">
                <ChevronRight size={12} className={`transition-transform ${isRevOpen(i) ? 'rotate-90' : ''}`} />
                {(() => {
                  if (r.member_name === 'Team deliverable') return '🧵 Team deliverable — the combined result';
                  const priorSame = (card.revisions ?? []).slice(0, i).filter((x) => x.member_name === r.member_name).length;
                  const who = r.member_name ?? 'Agent';
                  const part = card.parts ? "'s part" : "'s take";
                  return priorSame === 0 ? `${who}${part}` : `${who} · revision ${priorSame + 1} (after your feedback)`;
                })()}
                {r.tokens_used ? ` · ${(r.tokens_used / 1000).toFixed(1)}k tokens` : ''}
              </span>
              <span className="flex items-center gap-3">
                {r.approved && <span className="text-emerald-400 flex items-center gap-1"><Check size={12} /> approved</span>}
                <span
                  role="button"
                  tabIndex={0}
                  onClick={(e) => {
                    e.stopPropagation();
                    void crewApi.downloadPdf(card.id, card.title, i + 1).catch(() => {});
                  }}
                  title="Download this output as a PDF report"
                  className="flex items-center gap-1 text-[var(--color-text-tertiary)] hover:text-[var(--color-accent)]"
                >
                  <FileDown size={13} /> PDF
                </span>
              </span>
            </button>
            {isRevOpen(i) && (r.tools_used ?? []).length > 0 && (
              <div className="px-4 py-1.5 bg-[var(--color-background)]/50 text-[10px] text-[var(--color-text-tertiary)] truncate" title={(r.tools_used ?? []).join(' → ')}>
                ran: {(r.tools_used ?? []).join(' → ')}
              </div>
            )}
            {isRevOpen(i) && (
              <div className="p-4 text-sm crew-md">
                <MarkdownRenderer content={r.output} />
              </div>
            )}
            {isRevOpen(i) && r.review && (
              <div className="px-4 py-2 border-t border-[var(--color-border)] text-[13px]">
                <span className="text-[var(--color-text-tertiary)]">Your feedback: </span>
                <span className="text-amber-400">{r.review}</span>
              </div>
            )}
          </div>
        ))}
        {(card.revisions?.length ?? 0) === 0 && card.status !== 'working' && (
          <p className="text-sm text-[var(--color-text-tertiary)]">No output yet.</p>
        )}
        {card.status === 'working' && (
          <div className="flex items-center gap-2 text-sm text-[var(--color-text-tertiary)]">
            <Loader2 size={14} className="animate-spin text-[var(--color-accent)]" /> Agents are working on this card…
          </div>
        )}
      </div>

      <div className="px-6 py-4 border-t border-[var(--color-border)] space-y-2 shrink-0">
        {err && <div className="text-xs text-red-400">{err}</div>}
        {card.status === 'draft' && (
          <div className="flex items-center gap-2">
            <button
              disabled={busy || (!card.assignee_id && !card.assignee_ids?.length)}
              onClick={() => void act(() => crewApi.queueCard(card.id))}
              className="flex items-center gap-1.5 px-3.5 py-2 rounded-lg bg-[var(--color-accent)] text-white text-sm font-medium hover:opacity-90 disabled:opacity-40"
            >
              <ChevronRight size={15} /> Queue for the team
            </button>
            <button
              disabled={busy}
              title="A planning agent breaks THIS card into smaller sub-cards — assigned, dependency-chained, and created as drafts for you to review."
              onClick={() => void act(() => crewApi.planCard(card.id)).then(onClose)}
              className="flex items-center gap-1.5 px-3.5 py-2 rounded-lg border border-[var(--color-border)] text-sm hover:bg-[var(--color-surface-hover)] disabled:opacity-40"
            >
              {busy ? <Loader2 size={14} className="animate-spin" /> : <Sparkles size={14} />} Break into cards
            </button>
            {!card.assignee_id && !card.assignee_ids?.length && (
              <span className="text-xs text-[var(--color-text-tertiary)]">Assign at least one member first</span>
            )}
            <button
              disabled={busy}
              onClick={() => {
                if (window.confirm('Delete this card?')) void act(() => crewApi.deleteCard(card.id)).then(onClose);
              }}
              className="ml-auto p-2 rounded-lg text-[var(--color-text-tertiary)] hover:text-red-400"
            >
              <Trash2 size={15} />
            </button>
          </div>
        )}
        {card.status === 'review' && (
          <div className="space-y-2">
            <textarea
              value={feedback}
              onChange={(e) => setFeedback(e.target.value)}
              placeholder="Feedback for changes (required to send back)…"
              rows={2}
              className="w-full rounded-lg bg-[var(--color-background)] border border-[var(--color-border)] px-3 py-2 text-sm placeholder:text-[var(--color-text-tertiary)] focus:outline-none focus:border-[var(--color-accent)]"
            />
            <div className="flex items-center gap-2">
              <button
                disabled={busy}
                onClick={() => void act(() => crewApi.reviewCard(card.id, true, feedback))}
                className="flex items-center gap-1.5 px-3.5 py-2 rounded-lg bg-emerald-600 text-white text-sm font-medium hover:opacity-90 disabled:opacity-40"
              >
                <Check size={15} /> Approve → Done
              </button>
              <button
                disabled={busy || !feedback.trim()}
                onClick={() => void act(() => crewApi.reviewCard(card.id, false, feedback)).then(() => setFeedback(''))}
                className="flex items-center gap-1.5 px-3.5 py-2 rounded-lg border border-amber-500/40 text-amber-400 text-sm font-medium hover:bg-amber-500/10 disabled:opacity-40"
              >
                <RotateCcw size={15} /> Request changes → re-queue
              </button>
            </div>
          </div>
        )}
      </div>
    </>
  );
}

// ─── New card panel (50%) ────────────────────────────────────────────────────

function NewCardPanel({ projectId, project, members, cards, onClose, onCreated }: { projectId: string; project: Project | null; members: CrewMember[]; cards: CrewCard[]; onClose: () => void; onCreated: () => Promise<void> }) {
  const objectives = project?.objectives ?? [];
  const [title, setTitle] = useState('');
  const [detail, setDetail] = useState('');
  const [assignees, setAssignees] = useState<string[]>([]);
  const [objectiveId, setObjectiveId] = useState('');
  const [repeat, setRepeat] = useState('');
  const [dependsOn, setDependsOn] = useState<string[]>([]);
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState('');

  const create = async () => {
    setBusy(true);
    setErr('');
    try {
      await crewApi.createCard(projectId, { title, detail: detail || undefined, objective_id: objectiveId || undefined, repeat: repeat || undefined, assignee_ids: assignees.length ? assignees : undefined, depends_on: dependsOn.length ? dependsOn : undefined });
      await onCreated();
      onClose();
    } catch (e) {
      setErr((e as Error).message);
      setBusy(false);
    }
  };

  return (
    <>
      <div className="flex items-center justify-between px-6 py-4 border-b border-[var(--color-border)] shrink-0">
        <h2 className="font-semibold">New card <span className="text-xs font-normal text-[var(--color-text-tertiary)]">— starts as a draft</span></h2>
        <button onClick={onClose} className="text-[var(--color-text-tertiary)] hover:text-[var(--color-text-primary)]"><X size={18} /></button>
      </div>
      <div className="flex-1 overflow-auto px-6 py-4 space-y-4">
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="What needs to be done?"
          autoFocus
          className="w-full rounded-lg bg-[var(--color-background)] border border-[var(--color-border)] px-3 py-2.5 text-sm focus:outline-none focus:border-[var(--color-accent)]"
        />
        <textarea
          value={detail}
          onChange={(e) => setDetail(e.target.value)}
          placeholder="Details, requirements, links, acceptance criteria… the more specific, the better the output."
          rows={8}
          className="w-full rounded-lg bg-[var(--color-background)] border border-[var(--color-border)] px-3 py-2.5 text-sm focus:outline-none focus:border-[var(--color-accent)]"
        />
        {cards.length > 0 && (
          <div>
            <div className="text-xs text-[var(--color-text-tertiary)] mb-2">
              Depends on — this card won't run until these are done, and it receives their approved output as input.
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-1.5 max-h-40 overflow-auto">
              {cards.map((c) => (
                <label
                  key={c.id}
                  className={`flex items-center gap-2 text-xs rounded-lg border px-3 py-2 cursor-pointer transition ${
                    dependsOn.includes(c.id) ? 'border-[var(--color-accent)] bg-[var(--color-accent)]/10' : 'border-[var(--color-border)] hover:bg-[var(--color-surface-hover)]'
                  }`}
                >
                  <input
                    type="checkbox"
                    checked={dependsOn.includes(c.id)}
                    disabled={!dependsOn.includes(c.id) && dependsOn.length >= 5}
                    onChange={(e) => setDependsOn((prev) => (e.target.checked ? [...prev, c.id] : prev.filter((x) => x !== c.id)))}
                    className="accent-[var(--color-accent)]"
                  />
                  <span className="truncate flex-1">{c.title}</span>
                  <span className={`shrink-0 text-[9px] px-1 py-px rounded ${c.status === 'done' ? 'bg-emerald-500/10 text-emerald-400' : 'bg-[var(--color-surface-hover)] text-[var(--color-text-tertiary)]'}`}>{c.status}</span>
                </label>
              ))}
            </div>
          </div>
        )}
        <div>
          <div className="text-xs text-[var(--color-text-tertiary)] mb-2">Repeat — after each approval the card automatically re-queues on this cadence.</div>
          <div className="flex gap-1.5">
            {[
              ['', 'One-off'],
              ['daily', 'Daily'],
              ['weekly', 'Weekly'],
              ['monthly', 'Monthly'],
            ].map(([v, l]) => (
              <button
                key={v}
                onClick={() => setRepeat(v)}
                className={`text-xs px-2.5 py-1.5 rounded-lg border transition ${
                  repeat === v ? 'border-[var(--color-accent)] bg-[var(--color-accent)]/10 text-[var(--color-accent)]' : 'border-[var(--color-border)] hover:bg-[var(--color-surface-hover)]'
                }`}
              >
                {l}
              </button>
            ))}
          </div>
        </div>
        {objectives.length > 0 && (
          <div>
            <div className="text-xs text-[var(--color-text-tertiary)] mb-2">Which objective does this serve?</div>
            <div className="flex flex-wrap gap-1.5">
              {objectives.filter((o) => !o.done).map((o) => (
                <button
                  key={o.id}
                  onClick={() => setObjectiveId((cur) => (cur === o.id ? '' : o.id))}
                  className={`text-xs px-2.5 py-1.5 rounded-lg border transition ${
                    objectiveId === o.id ? 'border-[var(--color-accent)] bg-[var(--color-accent)]/10 text-[var(--color-accent)]' : 'border-[var(--color-border)] hover:bg-[var(--color-surface-hover)]'
                  }`}
                >
                  {o.title}
                </button>
              ))}
            </div>
          </div>
        )}
        <div>
          <div className="text-xs text-[var(--color-text-tertiary)] mb-2">
            Assign to — one or several. Each produces their own take; you review them together.
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-1.5">
            {members.length === 0 && <p className="text-xs text-[var(--color-text-tertiary)]">No team members yet — hire someone first.</p>}
            {members.map((m) => (
              <label
                key={m.id}
                className={`flex items-center gap-2 text-sm rounded-lg border px-3 py-2 cursor-pointer transition ${
                  assignees.includes(m.id) ? 'border-[var(--color-accent)] bg-[var(--color-accent)]/10' : 'border-[var(--color-border)] hover:bg-[var(--color-surface-hover)]'
                }`}
              >
                <input
                  type="checkbox"
                  checked={assignees.includes(m.id)}
                  onChange={(e) => setAssignees((prev) => (e.target.checked ? [...prev, m.id] : prev.filter((x) => x !== m.id)))}
                  className="accent-[var(--color-accent)]"
                />
                <span className="truncate">{m.name}</span>
              </label>
            ))}
          </div>
        </div>
      </div>
      <div className="px-6 py-4 border-t border-[var(--color-border)] shrink-0 space-y-2">
        {err && <div className="text-xs text-red-400">{err}</div>}
        <button
          disabled={busy || !title.trim()}
          onClick={() => void create()}
          className="w-full flex items-center justify-center gap-1.5 py-2.5 rounded-lg bg-[var(--color-accent)] text-white text-sm font-medium hover:opacity-90 disabled:opacity-40"
        >
          <Send size={14} /> {busy ? 'Creating…' : 'Create draft'}
        </button>
      </div>
    </>
  );
}

// ─── Hire panel (50%) ────────────────────────────────────────────────────────

function HirePanel({ projectId, roles, onClose, onHired }: { projectId: string; roles: CrewRole[]; onClose: () => void; onHired: () => Promise<void> }) {
  const [roleKey, setRoleKey] = useState('');
  const [name, setName] = useState('');
  const [tools, setTools] = useState<string[]>([]);
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState('');
  const role = roles.find((r) => r.key === roleKey);

  useEffect(() => {
    if (role) setTools(role.default_tools);
  }, [role]);

  const hire = async () => {
    if (!role) return;
    setBusy(true);
    setErr('');
    try {
      await crewApi.hire(projectId, { role_key: role.key, name: name.trim() || undefined, tools });
      await onHired();
      onClose();
    } catch (e) {
      setErr((e as Error).message);
      setBusy(false);
    }
  };

  return (
    <>
      <div className="flex items-center justify-between px-6 py-4 border-b border-[var(--color-border)] shrink-0">
        <h2 className="font-semibold flex items-center gap-2"><UserPlus size={17} className="text-[var(--color-accent)]" /> Hire a team member</h2>
        <button onClick={onClose} className="text-[var(--color-text-tertiary)] hover:text-[var(--color-text-primary)]"><X size={18} /></button>
      </div>
      <div className="flex-1 overflow-auto px-6 py-4">
        {!role ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
            {roles.map((r) => (
              <button key={r.key} onClick={() => setRoleKey(r.key)} className="text-left rounded-lg border border-[var(--color-border)] bg-[var(--color-background)] p-3.5 hover:border-[var(--color-accent)] transition">
                <div className="text-sm font-medium">{r.label}</div>
                <div className="text-xs text-[var(--color-text-tertiary)] mt-0.5">{r.blurb}</div>
              </button>
            ))}
          </div>
        ) : (
          <div className="space-y-4">
            <div className="text-sm">
              <span className="text-[var(--color-text-tertiary)]">Role:</span> {role.label}{' '}
              <button onClick={() => setRoleKey('')} className="text-xs text-[var(--color-accent)] ml-1">change</button>
            </div>
            <input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder={`Name (optional, default "${role.label}")`}
              className="w-full rounded-lg bg-[var(--color-background)] border border-[var(--color-border)] px-3 py-2.5 text-sm focus:outline-none focus:border-[var(--color-accent)]"
            />
            <div>
              <div className="flex items-baseline justify-between mb-2">
                <div className="text-xs text-[var(--color-text-tertiary)]">Tools this member may use — nothing else will ever be available to it:</div>
                <span className="text-[11px] text-[var(--color-accent)] tabular-nums">{tools.length} selected</span>
              </div>
              <div className="space-y-3">
                {groupTools(role.allowed_tools).map((g) => {
                  const allOn = g.tools.every((t) => tools.includes(t));
                  return (
                    <div key={g.label} className="rounded-lg border border-[var(--color-border)] overflow-hidden">
                      <label className="flex items-center gap-2 px-3 py-2 bg-[var(--color-background)] cursor-pointer">
                        <input
                          type="checkbox"
                          checked={allOn}
                          onChange={(e) =>
                            setTools((prev) =>
                              e.target.checked
                                ? [...new Set([...prev, ...g.tools])]
                                : prev.filter((x) => !g.tools.includes(x))
                            )
                          }
                          className="accent-[var(--color-accent)]"
                        />
                        <span className="text-[12px] font-semibold">{g.label}</span>
                        <span className="ml-auto text-[10px] text-[var(--color-text-tertiary)] tabular-nums">
                          {g.tools.filter((t) => tools.includes(t)).length}/{g.tools.length}
                        </span>
                      </label>
                      <div className="px-3 py-2 grid grid-cols-1 sm:grid-cols-2 gap-x-3 gap-y-1">
                        {g.tools.map((t) => (
                          <label key={t} className="flex items-center gap-1.5 text-sm min-w-0">
                            <input
                              type="checkbox"
                              checked={tools.includes(t)}
                              onChange={(e) => setTools((prev) => (e.target.checked ? [...prev, t] : prev.filter((x) => x !== t)))}
                              className="accent-[var(--color-accent)] shrink-0"
                            />
                            <code className="text-[11px] truncate">{t.replace(/^(gmail|googlecalendar|googledrive|googlesheets|calendly|zoom|canva|notion|trello|clickup|jira|linear|airtable|github|hubspot|leadsquared|shopify|mailchimp|x|linkedin|youtube|twilio|facebook_ads)_/, '')}</code>
                            {actsExternally(t) && (
                              <span title="Acts on external systems (sends/posts/modifies) during the run" className="text-amber-400 text-[10px] shrink-0">⚡</span>
                            )}
                          </label>
                        ))}
                      </div>
                    </div>
                  );
                })}
              </div>
              <p className="text-[11px] text-amber-400/80 mt-2">⚡ = acts on external systems during the run (sends, posts, modifies). Grant deliberately — reviews happen after the run.</p>
            </div>
            <p className="text-xs text-[var(--color-text-tertiary)]">
              After hiring, open the member to attach reference documents — they'll ground every task in them.
            </p>
          </div>
        )}
      </div>
      {role && (
        <div className="px-6 py-4 border-t border-[var(--color-border)] shrink-0 space-y-2">
          {err && <div className="text-xs text-red-400">{err}</div>}
          <button
            disabled={busy || tools.length === 0}
            onClick={() => void hire()}
            className="w-full py-2.5 rounded-lg bg-[var(--color-accent)] text-white text-sm font-medium hover:opacity-90 disabled:opacity-40"
          >
            {busy ? 'Hiring…' : 'Hire'}
          </button>
        </div>
      )}
    </>
  );
}

// ─── Member panel (50%) — details, documents (RAG), pause/fire ───────────────

function GoalPanel({ project, onClose, onChanged }: { project: Project; onClose: () => void; onChanged: () => Promise<void> }) {
  const [goal, setGoal] = useState(project.goal ?? '');
  const [objectives, setObjectives] = useState<CrewObjective[]>(project.objectives ?? []);
  const [newObj, setNewObj] = useState('');
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState('');

  const save = async () => {
    setBusy(true);
    setErr('');
    try {
      await crewApi.updateProject(project.id, { goal, objectives });
      await onChanged();
      onClose();
    } catch (e) {
      setErr((e as Error).message);
      setBusy(false);
    }
  };
  const addObjective = () => {
    const t = newObj.trim();
    if (!t) return;
    setObjectives((prev) => [...prev, { id: `obj-${Date.now()}-${prev.length}`, title: t, done: false }]);
    setNewObj('');
  };

  return (
    <>
      <div className="flex items-center justify-between px-6 py-4 border-b border-[var(--color-border)] shrink-0">
        <h2 className="font-semibold flex items-center gap-2"><Target size={17} className="text-[var(--color-accent)]" /> Goal & objectives</h2>
        <button onClick={onClose} className="text-[var(--color-text-tertiary)] hover:text-[var(--color-text-primary)]"><X size={18} /></button>
      </div>
      <div className="flex-1 overflow-auto px-6 py-4 space-y-5">
        <div>
          <div className="text-xs text-[var(--color-text-tertiary)] mb-2">
            The north star. Every member sees this on every task and must serve it — make it one concrete, measurable outcome.
          </div>
          <textarea
            value={goal}
            onChange={(e) => setGoal(e.target.value)}
            placeholder={'e.g. "Publish 8 SEO articles that rank for our target keywords and drive 10k organic visits by September."'}
            rows={3}
            className="w-full rounded-lg bg-[var(--color-background)] border border-[var(--color-border)] px-3 py-2.5 text-sm focus:outline-none focus:border-[var(--color-accent)]"
          />
        </div>
        <div>
          <div className="text-xs text-[var(--color-text-tertiary)] mb-2">
            Objectives — the workstreams that map to the goal. Cards link to one, so every agent knows which part of the plan its task serves.
          </div>
          <div className="space-y-1.5">
            {objectives.map((o, i) => (
              <div key={o.id} className="flex items-center gap-2 rounded-lg bg-[var(--color-background)]/60 px-3 py-2">
                <input
                  type="checkbox"
                  checked={o.done}
                  onChange={(e) => setObjectives((prev) => prev.map((x, j) => (j === i ? { ...x, done: e.target.checked } : x)))}
                  className="accent-[var(--color-accent)]"
                  title="Mark objective done"
                />
                <span className={`text-sm flex-1 ${o.done ? 'line-through text-[var(--color-text-tertiary)]' : ''}`}>{o.title}</span>
                <button onClick={() => setObjectives((prev) => prev.filter((_, j) => j !== i))} className="text-[var(--color-text-tertiary)] hover:text-red-400" aria-label="Remove objective">
                  <Trash2 size={13} />
                </button>
              </div>
            ))}
          </div>
          <div className="flex gap-2 mt-2">
            <input
              value={newObj}
              onChange={(e) => setNewObj(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && addObjective()}
              placeholder="Add an objective…"
              className="flex-1 rounded-lg bg-[var(--color-background)] border border-[var(--color-border)] px-3 py-2 text-sm focus:outline-none focus:border-[var(--color-accent)]"
            />
            <button onClick={addObjective} className="px-3 rounded-lg border border-[var(--color-border)] hover:bg-[var(--color-surface-hover)]"><Plus size={15} /></button>
          </div>
        </div>
      </div>
      <div className="px-6 py-4 border-t border-[var(--color-border)] shrink-0 space-y-2">
        {err && <div className="text-xs text-red-400">{err}</div>}
        <div className="flex gap-2">
          <button
            disabled={busy}
            onClick={() => void save()}
            className="flex-1 py-2.5 rounded-lg bg-[var(--color-accent)] text-white text-sm font-medium hover:opacity-90 disabled:opacity-40"
          >
            {busy ? 'Saving…' : 'Save goal & objectives'}
          </button>
          <button
            disabled={busy}
            title="Archive — removes the project from your list. Nothing is deleted."
            onClick={() => {
              if (!window.confirm('Archive this project? It will disappear from your list and its cards stop running. Nothing is deleted from the database.')) return;
              setBusy(true);
              void crewApi
                .archiveProject(project.id)
                .then(() => {
                  window.location.href = '/crew';
                })
                .catch((e) => {
                  setErr((e as Error).message);
                  setBusy(false);
                });
            }}
            className="px-3.5 py-2.5 rounded-lg border border-red-500/30 text-red-400 text-sm hover:bg-red-500/10 disabled:opacity-40"
          >
            Archive
          </button>
        </div>
      </div>
    </>
  );
}

function MemberPanel({ member, roles, onClose, onChanged }: { member: CrewMember; roles: CrewRole[]; onClose: () => void; onChanged: () => Promise<void> }) {
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState('');
  const [allSkills, setAllSkills] = useState<CrewSkill[]>([]);
  const fileRef = useRef<HTMLInputElement>(null);
  const role = roles.find((r) => r.key === member.role_key);
  useEffect(() => {
    void crewApi.skills().then(setAllSkills).catch(() => {});
  }, []);
  const toggleSkill = (id: string) => {
    const cur = member.skill_ids ?? [];
    const next = cur.includes(id) ? cur.filter((x) => x !== id) : [...cur, id];
    void (async () => {
      setBusy(true);
      setErr('');
      try {
        await crewApi.setMemberSkills(member.id, next);
        await onChanged();
      } catch (e) {
        setErr((e as Error).message);
      } finally {
        setBusy(false);
      }
    })();
  };

  const act = async (fn: () => Promise<unknown>) => {
    setBusy(true);
    setErr('');
    try {
      await fn();
      await onChanged();
    } catch (e) {
      setErr((e as Error).message);
    } finally {
      setBusy(false);
    }
  };

  return (
    <>
      <div className="flex items-center justify-between px-6 py-4 border-b border-[var(--color-border)] shrink-0">
        <div>
          <h2 className="font-semibold">{member.name}</h2>
          <p className="text-xs text-[var(--color-text-tertiary)]">
            {role?.label ?? member.role_key} · {member.status} · {((member.tokens_in + member.tokens_out) / 1000).toFixed(1)}k tokens lifetime
          </p>
        </div>
        <button onClick={onClose} className="text-[var(--color-text-tertiary)] hover:text-[var(--color-text-primary)]"><X size={18} /></button>
      </div>

      <div className="flex-1 overflow-auto px-6 py-4 space-y-5">
        <div>
          <h3 className="text-xs font-semibold uppercase tracking-wider text-[var(--color-text-tertiary)] mb-2">Lane — what {member.name} owns</h3>
          <textarea
            defaultValue={member.charter ?? ''}
            onBlur={(e) => {
              if ((e.target.value ?? '') !== (member.charter ?? '')) void act(() => crewApi.setMemberCharter(member.id, e.target.value));
            }}
            placeholder={'e.g. "Owns all keyword research and on-page SEO recommendations. Does NOT write final copy — hands drafts to the Content Writer."'}
            rows={3}
            className="w-full rounded-lg bg-[var(--color-background)] border border-[var(--color-border)] px-3 py-2 text-xs focus:outline-none focus:border-[var(--color-accent)]"
          />
          <p className="text-[10.5px] text-[var(--color-text-tertiary)] mt-1">Injected into every task this member runs — it stays in its lane and flags hand-offs. Saves when you click away.</p>
        </div>
        <div>
          <h3 className="text-xs font-semibold uppercase tracking-wider text-[var(--color-text-tertiary)] mb-2">Monthly token budget</h3>
          <div className="flex items-center gap-2">
            <input
              type="number"
              min={0}
              step={100000}
              defaultValue={member.monthly_budget ?? 0}
              onBlur={(e) => {
                const v = Math.max(0, Number(e.target.value) || 0);
                if (v !== (member.monthly_budget ?? 0)) void act(() => crewApi.setMemberBudget(member.id, v));
              }}
              className="w-40 rounded-lg bg-[var(--color-background)] border border-[var(--color-border)] px-3 py-2 text-sm focus:outline-none focus:border-[var(--color-accent)]"
            />
            <span className="text-[11px] text-[var(--color-text-tertiary)]">
              tokens/month · 0 = unlimited{(member.monthly_budget ?? 0) > 0 ? ` · used ${(((member.month_tokens ?? 0) / 1000) | 0)}k of ${(((member.monthly_budget ?? 0) / 1000) | 0)}k this month` : ''}
            </span>
          </div>
          {(member.monthly_budget ?? 0) > 0 && (member.month_tokens ?? 0) >= (member.monthly_budget ?? 0) && (
            <p className="text-[11px] text-amber-400 mt-1">Budget exhausted — this member pauses until next month (or raise the cap).</p>
          )}
        </div>
        <div>
          <h3 className="text-xs font-semibold uppercase tracking-wider text-[var(--color-text-tertiary)] mb-2">Tools</h3>
          <div className="flex flex-wrap gap-1.5">
            {member.tools.map((t) => (
              <code key={t} className="text-xs px-2 py-1 rounded-md bg-[var(--color-background)] border border-[var(--color-border)]">{t}</code>
            ))}
          </div>
        </div>

        <div>
          <h3 className="text-xs font-semibold uppercase tracking-wider text-[var(--color-text-tertiary)] mb-2 flex items-center gap-1.5">
            <Sparkles size={12} /> Skills <span className="normal-case tracking-normal">· {(member.skill_ids ?? []).length}/6</span>
          </h3>
          <p className="text-[11px] text-[var(--color-text-tertiary)] mb-2">
            Attach skills — each adds its practiced approach to this member and grants the tools it needs.
          </p>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-1.5 mb-1">
            {allSkills.map((sk) => {
              const on = (member.skill_ids ?? []).includes(sk.id);
              return (
                <button
                  key={sk.id}
                  disabled={busy || (!on && (member.skill_ids ?? []).length >= 6)}
                  onClick={() => toggleSkill(sk.id)}
                  title={sk.description}
                  className={`text-left rounded-lg border px-3 py-2 transition disabled:opacity-40 ${
                    on ? 'border-[var(--color-accent)] bg-[var(--color-accent)]/10' : 'border-[var(--color-border)] hover:bg-[var(--color-surface-hover)]'
                  }`}
                >
                  <div className="text-[13px] font-medium truncate">{sk.name}</div>
                  <div className="text-[10px] text-[var(--color-text-tertiary)] truncate">{sk.description}</div>
                </button>
              );
            })}
            {allSkills.length === 0 && <p className="text-xs text-[var(--color-text-tertiary)]">No skills available yet.</p>}
          </div>
        </div>

        <div>
          <div className="flex items-center justify-between mb-2">
            <h3 className="text-xs font-semibold uppercase tracking-wider text-[var(--color-text-tertiary)] flex items-center gap-1.5">
              <FileText size={12} /> Reference documents
            </h3>
            <button
              disabled={busy}
              onClick={() => fileRef.current?.click()}
              className="flex items-center gap-1 text-xs px-2.5 py-1 rounded-md border border-[var(--color-border)] text-[var(--color-accent)] hover:bg-[var(--color-surface-hover)] disabled:opacity-40"
            >
              <Paperclip size={11} /> {busy ? 'Uploading…' : 'Add document'}
            </button>
            <input
              ref={fileRef}
              type="file"
              hidden
              accept=".pdf,.txt,.md,.csv,.docx,.html"
              onChange={(e) => {
                const f = e.target.files?.[0];
                if (f) void act(() => crewApi.uploadMemberDoc(member.id, f));
                e.target.value = '';
              }}
            />
          </div>
          <p className="text-[11px] text-[var(--color-text-tertiary)] mb-2">
            The member grounds every task in these before using tools (brand guides, product specs, style guides, data sheets…).
          </p>
          <div className="space-y-1.5">
            {(member.documents ?? []).map((d, i) => (
              <div key={i} className="flex items-center gap-2 rounded-lg border border-[var(--color-border)] bg-[var(--color-background)] px-3 py-2">
                <FileText size={13} className="text-[var(--color-text-tertiary)] shrink-0" />
                <span className="text-sm truncate flex-1">{d.name}</span>
                <span className="text-[11px] text-[var(--color-text-tertiary)] shrink-0">{(d.size / 1000).toFixed(1)}k chars</span>
                <button
                  disabled={busy}
                  onClick={() => void act(() => crewApi.deleteMemberDoc(member.id, i))}
                  className="text-[var(--color-text-tertiary)] hover:text-red-400 shrink-0"
                >
                  <Trash2 size={13} />
                </button>
              </div>
            ))}
            {(member.documents?.length ?? 0) === 0 && (
              <p className="text-xs text-[var(--color-text-tertiary)]">No documents yet.</p>
            )}
          </div>
        </div>
      </div>

      <div className="px-6 py-4 border-t border-[var(--color-border)] shrink-0 space-y-2">
        {err && <div className="text-xs text-red-400">{err}</div>}
        <div className="flex items-center gap-2">
          <button
            disabled={busy}
            onClick={() => void act(() => crewApi.setMemberStatus(member.id, member.status === 'active' ? 'paused' : 'active'))}
            className="flex items-center gap-1.5 px-3.5 py-2 rounded-lg border border-[var(--color-border)] text-sm hover:bg-[var(--color-surface-hover)] disabled:opacity-40"
          >
            {member.status === 'active' ? <><Pause size={14} /> Pause</> : <><Play size={14} /> Resume</>}
          </button>
          <button
            disabled={busy}
            onClick={() => {
              if (window.confirm(`Remove ${member.name}? Their queued cards return to drafts.`))
                void act(() => crewApi.fire(member.id)).then(onClose);
            }}
            className="ml-auto flex items-center gap-1.5 px-3.5 py-2 rounded-lg border border-red-500/30 text-red-400 text-sm hover:bg-red-500/10 disabled:opacity-40"
          >
            <Trash2 size={14} /> Remove
          </button>
        </div>
      </div>
    </>
  );
}
