import { useState } from 'react';
import { Terminal, Copy, Check, Info } from 'lucide-react';
import styles from './ClaraCliSection.module.css';

// ============================================================================
// ClaraCliSection - setup instructions for Clara Agent, the local CLI agent.
//
// Clara Agent runs on the user's own machine and talks to their configured
// model provider directly. It authenticates with the same device-authorization
// flow the Devices tab lists, so once `/login claraverse` completes, the
// machine shows up under Settings > Devices and can be revoked from there.
// ============================================================================

interface CommandStep {
  title: string;
  text: string;
  command: string;
}

const STEPS: CommandStep[] = [
  {
    title: 'Install',
    text: 'Run once from the ClaraVerse repo. Builds the agent and puts the claracli command on your PATH.',
    command: 'cd clara-agent && npm run setup',
  },
  {
    title: 'Start it',
    text: 'Opens the agent in your terminal, in whatever project directory you are in.',
    command: 'claracli',
  },
  {
    title: 'Connect it to this account',
    text: 'Run this inside the agent. You will get a short code to confirm in your browser, then it picks up the model configured above automatically.',
    command: '/login claraverse',
  },
];

export const ClaraCliSection = () => {
  const [copied, setCopied] = useState<string | null>(null);

  const handleCopy = async (command: string) => {
    try {
      await navigator.clipboard.writeText(command);
      setCopied(command);
      setTimeout(() => setCopied(current => (current === command ? null : current)), 1500);
    } catch {
      // Clipboard can be unavailable over plain http or without permission.
      // The command is visible and selectable either way, so fail quietly.
    }
  };

  return (
    <div className={styles.container}>
      <div>
        <h2 className={styles.title}>
          <Terminal className={styles.titleIcon} />
          Clara Agent (CLI)
        </h2>
        <p className={styles.description}>
          A coding agent that runs in your terminal on this machine, using the same model and
          account as the web app. It can read, write, and edit files and run commands locally.
        </p>
      </div>

      <div className={styles.steps}>
        {STEPS.map((step, index) => (
          <div key={step.title} className={styles.step}>
            <div className={styles.stepNumber}>{index + 1}</div>
            <div className={styles.stepBody}>
              <h3 className={styles.stepTitle}>{step.title}</h3>
              <p className={styles.stepText}>{step.text}</p>
              <div className={styles.codeRow}>
                <div className={styles.code}>{step.command}</div>
                <button
                  type="button"
                  className={styles.copyButton}
                  onClick={() => handleCopy(step.command)}
                  aria-label={`Copy: ${step.command}`}
                >
                  {copied === step.command ? <Check size={14} /> : <Copy size={14} />}
                  {copied === step.command ? 'Copied' : 'Copy'}
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className={styles.note}>
        <Info className={styles.noteIcon} />
        <p className={styles.noteText}>
          Once connected, this machine appears under <strong>Devices</strong>, where you can rename
          or revoke its access at any time. To point the agent at a different ClaraVerse instance,
          set <code className={styles.inlineCode}>CLARAVERSE_URL</code> before starting it.
        </p>
      </div>

      <div className={styles.note}>
        <Info className={styles.noteIcon} />
        <p className={styles.noteText}>
          Clara Agent runs with your user account&apos;s permissions and has no sandbox. It can
          modify files and run commands on this machine, so review what you ask it to do.
        </p>
      </div>
    </div>
  );
};
