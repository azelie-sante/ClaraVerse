/**
 * Nexus v2 home (/crew): your agent projects. Create a project with a brief —
 * the shared context every team member and card will see.
 */
import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Loader2, Plus, Users, X } from 'lucide-react';
import { CrewSidebar } from '@/features/crew/ui/CrewSidebar';
import { CrewHelp } from '@/features/crew/ui/CrewHelp';
import { crewApi, type CrewProject, type CrewTemplate } from '@/features/crew/api';

export default function CrewHome() {
  const navigate = useNavigate();
  const [projects, setProjects] = useState<CrewProject[]>([]);
  const [loading, setLoading] = useState(true);
  const [showNew, setShowNew] = useState(false);
  const [showHelp, setShowHelp] = useState(false);
  const [name, setName] = useState('');
  const [brief, setBrief] = useState('');
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState('');
  const [templates, setTemplates] = useState<CrewTemplate[]>([]);
  const [tpl, setTpl] = useState<string>('');

  const load = useCallback(async () => {
    try {
      setProjects(await crewApi.listProjects());
    } catch (e) {
      setErr((e as Error).message);
    } finally {
      setLoading(false);
    }
  }, []);
  useEffect(() => {
    void load();
    void crewApi.templates().then(setTemplates).catch(() => {});
  }, [load]);

  const create = async () => {
    setBusy(true);
    setErr('');
    try {
      const p = tpl
        ? await crewApi.createFromTemplate(tpl, name.trim() || undefined, brief)
        : await crewApi.createProject(name, brief);
      navigate(`/crew/${p.id}`);
    } catch (e) {
      setErr((e as Error).message);
      setBusy(false);
    }
  };

  return (
    <div className="h-screen flex bg-[var(--color-background)] text-[var(--color-text-primary)]">
      <CrewSidebar
        onNewProject={() => setShowNew(true)}
        onProjects={() => {}}
        onHelp={() => setShowHelp(true)}
        projectsActive
        recents={projects.slice(0, 5).map((p) => ({ id: p.id, title: p.name, onClick: () => navigate(`/crew/${p.id}`) }))}
      />
      <div className="flex-1 overflow-auto">
      <div className="max-w-4xl mx-auto px-6 py-10">
        <div className="flex items-end justify-between mb-2">
          <div>
            <p className="text-xs uppercase tracking-[0.2em] text-[var(--color-accent)]">Agents · Crew</p>
            <h1 className="text-3xl font-bold mt-1.5">Your projects</h1>
          </div>
          <button
            onClick={() => setShowNew(true)}
            className="flex items-center gap-1.5 px-3.5 py-2 rounded-lg bg-[var(--color-accent)] text-white text-sm font-medium hover:opacity-90"
          >
            <Plus size={15} /> New project
          </button>
        </div>
        <p className="text-sm text-[var(--color-text-tertiary)] mb-8 max-w-xl">
          Create a project, hire a team of agents, queue work as cards — everything they produce comes to you for review.
        </p>

        {err && <div className="rounded-lg bg-red-500/10 border border-red-500/20 px-4 py-3 text-sm text-red-400 mb-4">{err}</div>}

        {loading ? (
          <div className="flex justify-center py-16 text-[var(--color-text-tertiary)]"><Loader2 className="animate-spin" /></div>
        ) : projects.length === 0 ? (
          <div className="text-center py-16 rounded-xl border border-dashed border-[var(--color-border)]">
            <Users size={26} className="mx-auto text-[var(--color-text-tertiary)] mb-3" />
            <p className="font-medium">No projects yet</p>
            <p className="text-sm text-[var(--color-text-tertiary)] mt-1">Start one — a project brief is all your team needs.</p>
          </div>
        ) : (
          <div className="grid sm:grid-cols-2 gap-3">
            {projects.map((p) => (
              <button
                key={p.id}
                onClick={() => navigate(`/crew/${p.id}`)}
                className="text-left rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] p-4 hover:border-[var(--color-accent)] transition"
              >
                <div className="font-medium">{p.name}</div>
                <div className="text-sm text-[var(--color-text-tertiary)] mt-1 line-clamp-2">{p.brief || 'No brief yet'}</div>
              </button>
            ))}
          </div>
        )}
      </div>

      </div>
      {showHelp && <CrewHelp onClose={() => setShowHelp(false)} />}
      {showNew && (
        <div className="fixed inset-0 z-40 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4" onClick={() => setShowNew(false)}>
          <div className="w-full max-w-2xl rounded-xl frost-panel shadow-2xl p-5 max-h-[85vh] overflow-auto" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between mb-4">
              <h2 className="font-semibold">New project</h2>
              <button onClick={() => setShowNew(false)} className="text-[var(--color-text-tertiary)] hover:text-[var(--color-text-primary)]"><X size={18} /></button>
            </div>
            <div className="space-y-3">
              {templates.length > 0 && (
                <div>
                  <div className="text-xs text-[var(--color-text-tertiary)] mb-2">Start from a template — a full team, objectives and starter cards, ready to review before anything runs. Or go blank.</div>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                    <button
                      onClick={() => setTpl('')}
                      className={`text-left rounded-lg border p-3 transition ${tpl === '' ? 'border-[var(--color-accent)] bg-[var(--color-accent)]/10' : 'border-[var(--color-border)] hover:bg-[var(--color-surface-hover)]'}`}
                    >
                      <div className="text-sm font-medium">Blank project</div>
                      <div className="text-[11px] text-[var(--color-text-tertiary)] mt-0.5">Build your own team from scratch.</div>
                    </button>
                    {templates.map((t) => (
                      <button
                        key={t.key}
                        onClick={() => setTpl(t.key)}
                        className={`text-left rounded-lg border p-3 transition ${tpl === t.key ? 'border-[var(--color-accent)] bg-[var(--color-accent)]/10' : 'border-[var(--color-border)] hover:bg-[var(--color-surface-hover)]'}`}
                      >
                        <div className="text-sm font-medium">{t.name}</div>
                        <div className="text-[11px] text-[var(--color-text-tertiary)] mt-0.5">{t.blurb}</div>
                        <div className="text-[10px] text-[var(--color-text-tertiary)]/70 mt-1.5">
                          {t.members.length} members · {t.cards.length} starter cards{t.cards.some((c) => c.repeat) ? ' · recurring' : ''}
                        </div>
                      </button>
                    ))}
                  </div>
                </div>
              )}
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder={tpl ? `Project name (default: ${templates.find((t) => t.key === tpl)?.name ?? ''})` : 'Project name'}
                autoFocus
                className="w-full rounded-lg bg-[var(--color-background)] border border-[var(--color-border)] px-3 py-2 text-sm focus:outline-none focus:border-[var(--color-accent)]"
              />
              <textarea
                value={brief}
                onChange={(e) => setBrief(e.target.value)}
                placeholder="The brief — what is this project about, for whom, and what does good look like? Every agent on the team sees this on every task."
                rows={5}
                className="w-full rounded-lg bg-[var(--color-background)] border border-[var(--color-border)] px-3 py-2 text-sm focus:outline-none focus:border-[var(--color-accent)]"
              />
              {err && <div className="text-xs text-red-400">{err}</div>}
              <button
                disabled={busy || (!tpl && !name.trim())}
                onClick={() => void create()}
                className="w-full py-2 rounded-lg bg-[var(--color-accent)] text-white text-sm font-medium hover:opacity-90 disabled:opacity-40"
              >
                {busy ? 'Creating…' : tpl ? 'Create team from template' : 'Create project'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
