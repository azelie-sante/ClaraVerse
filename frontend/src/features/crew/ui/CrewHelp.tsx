/**
 * Nexus playbook — the "how to get the most out of this" panel, opened from
 * the sidebar. Scenarios first (what to actually use it for), then the
 * pipeline mechanics, then the levers that make outputs good.
 */
import React from 'react';
import { X, Search, PenLine, Mail, GitCompareArrows, Repeat2, FileText } from 'lucide-react';

const SCENARIOS = [
  {
    icon: Search,
    title: 'Competitor & market research',
    body: 'Hire a Researcher and an SEO Specialist. Queue cards like "Full competitive teardown of X" or "What changed in this market this quarter". Each becomes a cited report you review — approve or send back with feedback.',
  },
  {
    icon: PenLine,
    title: 'Content pipeline',
    body: 'Researcher gathers the material → you approve it → Content Writer drafts on top of it. Approved research is in team memory, so the writer builds on real facts instead of guessing.',
  },
  {
    icon: Mail,
    title: 'Ops on your integrations',
    body: 'A Generalist gets every integration you\'ve connected (mail, CRM, …). Cards like "Summarize this week\'s inbox and draft replies" run server-side while you do something else.',
  },
  {
    icon: GitCompareArrows,
    title: 'Multiple takes on one problem',
    body: 'Assign the same card to two or three members. Each submits their own take for the same review — compare, approve the best, or merge them with your feedback.',
  },
  {
    icon: Repeat2,
    title: 'Iterate until it\'s right',
    body: 'Rejecting with feedback re-queues the card automatically. The agent sees every previous attempt and every review note, so revision 3 is genuinely better than revision 1.',
  },
  {
    icon: FileText,
    title: 'Grounded deliverables',
    body: 'Upload brand guidelines, product docs, or data files to a member (open the member → Documents). They ground every task in those files before reaching for tools.',
  },
];

const PIPELINE = [
  ['Draft', 'plan freely — nothing runs yet'],
  ['Queued', 'drag a draft here; an agent claims it within seconds'],
  ['Working', 'runs on the server — close the tab if you want'],
  ['Review', 'every output stops here for YOU. Approve, or reject with feedback'],
  ['Done', 'approved work joins team memory for all future tasks'],
];

const LEVERS = [
  ['Set a goal and objectives', 'Open Goal in the project header. The goal is a non-negotiable in every task run; objectives map the work, and each card links to the objective it serves — so nothing drifts off-plan.'],
  ['Break big cards down', 'Open a draft card and hit "Break into cards" — a planning agent splits it into smaller sub-cards, assigned and dependency-chained so outputs flow in order. Always drafts you approve first.'],
  ['Write a real brief', 'Every member sees the project brief on every task. "Research project" gets vague output; "Weekly competitor intel for our growth team, focus on pricing and product launches" gets sharp output.'],
  ['Cards are tasks, not chats', 'One card = one deliverable with a clear "done". Split big goals into cards — they run in parallel.'],
  ['Your review is the steering wheel', 'Feedback on rejection is the highest-leverage text you write — it\'s injected verbatim into the retry.'],
  ['Skills & documents compound', 'Attach skills for practiced methods and docs for ground truth. Both apply to every task that member touches.'],
];

export const CrewHelp: React.FC<{ onClose: () => void }> = ({ onClose }) => (
  <div className="fixed inset-0 z-40" role="dialog" aria-modal>
    <div className="absolute inset-0 bg-black/50" onClick={onClose} />
    <div
      className="absolute inset-y-0 right-0 frost-panel !border-y-0 !border-r-0 shadow-2xl flex flex-col crew-slide-in overflow-hidden"
      style={{ width: 'min(50%, 640px)', minWidth: 380 }}
    >
      <div className="flex items-center justify-between px-6 py-4 shrink-0">
        <div>
          <h2 className="text-[15px] font-semibold tracking-tight">Getting the most out of Crew</h2>
          <p className="text-[11px] text-[var(--color-text-tertiary)] mt-0.5">A team of agents you manage — not another chat</p>
        </div>
        <button onClick={onClose} className="text-[var(--color-text-tertiary)] hover:text-[var(--color-text-primary)]" aria-label="Close">
          <X size={18} />
        </button>
      </div>

      <div className="flex-1 overflow-auto px-6 pb-8 space-y-8">
        <section>
          <h3 className="text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--color-text-secondary)] mb-3">What it's for</h3>
          <div className="space-y-2">
            {SCENARIOS.map((s) => (
              <div key={s.title} className="flex gap-3 rounded-lg bg-[var(--color-background)]/50 p-3">
                <s.icon size={15} className="shrink-0 mt-0.5 text-[var(--color-accent)]" />
                <div>
                  <div className="text-[13px] font-medium">{s.title}</div>
                  <div className="text-[12px] text-[var(--color-text-tertiary)] mt-0.5 leading-relaxed">{s.body}</div>
                </div>
              </div>
            ))}
          </div>
        </section>

        <section>
          <h3 className="text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--color-text-secondary)] mb-3">How the pipeline works</h3>
          <ol className="space-y-1.5">
            {PIPELINE.map(([stage, desc], i) => (
              <li key={stage} className="flex items-baseline gap-2.5 text-[12.5px]">
                <span className="text-[10px] tabular-nums text-[var(--color-text-tertiary)] w-3">{i + 1}</span>
                <span className="font-medium min-w-[62px]">{stage}</span>
                <span className="text-[var(--color-text-tertiary)]">{desc}</span>
              </li>
            ))}
          </ol>
        </section>

        <section>
          <h3 className="text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--color-text-secondary)] mb-3">Team memory</h3>
          <p className="text-[12.5px] text-[var(--color-text-tertiary)] leading-relaxed">
            Every task run sees the project brief, the live board (what everyone else is doing), and the team's recent
            approved work — so terminology, facts, and tone stay consistent across the whole project. Unapproved work is
            deliberately invisible: one member's rejected draft never contaminates another's task.
          </p>
        </section>

        <section>
          <h3 className="text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--color-text-secondary)] mb-3">Getting better output</h3>
          <div className="space-y-3">
            {LEVERS.map(([t, d]) => (
              <div key={t}>
                <div className="text-[13px] font-medium">{t}</div>
                <div className="text-[12px] text-[var(--color-text-tertiary)] mt-0.5 leading-relaxed">{d}</div>
              </div>
            ))}
          </div>
        </section>

        <section>
          <h3 className="text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--color-text-secondary)] mb-3">What's possible today — and what's not yet</h3>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div className="rounded-lg bg-emerald-500/8 p-3.5">
              <div className="text-[12px] font-medium text-emerald-400 mb-1.5">Agents can</div>
              <ul className="space-y-1 text-[11.5px] text-[var(--color-text-tertiary)] leading-relaxed">
                <li>Research the web deeply and write cited reports</li>
                <li>Draft blog posts, emails, and marketing copy</li>
                <li>Read &amp; act on your connected integrations (mail, CRM…)</li>
                <li>Analyze spreadsheets, data files, and documents</li>
                <li>Work from your uploaded reference documents</li>
                <li>Run on a schedule — daily, weekly or monthly recurring cards</li>
                <li>Break a big card into a dependency-chained pipeline of sub-cards</li>
                <li>Revise until your review approves</li>
              </ul>
            </div>
            <div className="rounded-lg bg-[var(--color-background)]/50 p-3.5">
              <div className="text-[12px] font-medium text-[var(--color-text-secondary)] mb-1.5">Not yet (on the roadmap)</div>
              <ul className="space-y-1 text-[11.5px] text-[var(--color-text-tertiary)] leading-relaxed">
                <li>Publishing directly to a blog/CMS — agents produce the ready-to-publish piece + exact publishing steps</li>
                <li>Posting to social media accounts</li>
                <li>Agents talking to each other mid-task — hand-offs go through your review</li>
              </ul>
              <p className="text-[10.5px] text-[var(--color-text-tertiary)]/70 mt-2">
                Agents are told exactly what they can't do — they'll never claim to have published or sent something without the tool to do it.
              </p>
            </div>
          </div>
        </section>

        <section className="rounded-lg bg-[var(--color-accent)]/8 p-4">
          <div className="text-[13px] font-medium text-[var(--color-accent)]">When to use chat instead</div>
          <p className="text-[12px] text-[var(--color-text-tertiary)] mt-1 leading-relaxed">
            Quick questions and back-and-forth exploration belong in Chat. Nexus wins when the work is a{' '}
            <span className="text-[var(--color-text-secondary)]">deliverable</span>: it runs in the background, several
            tasks run in parallel, everything stops for your review, and the project remembers what was approved.
          </p>
        </section>
      </div>
    </div>
  </div>
);
