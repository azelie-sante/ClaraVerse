import { memo, useState } from 'react';
import {
  ChevronDown,
  ChevronRight,
  Wrench,
  CheckCircle2,
  AlertCircle,
  Loader2,
  X,
} from 'lucide-react';
import { MarkdownRenderer } from '@/components/design-system/content/MarkdownRenderer';
import type { NexusConversationItem } from '@/types/nexus';
import styles from './Nexus.module.css';

interface DaemonActivityBlockProps {
  daemonId: string;
  daemonRole?: string;
  items: NexusConversationItem[];
  onCancel?: (daemonId: string) => void;
}

export const DaemonActivityBlock = memo(function DaemonActivityBlock({
  daemonId,
  daemonRole,
  items,
  onCancel,
}: DaemonActivityBlockProps) {
  const [expanded, setExpanded] = useState(true);

  const hasToolCalls = items.some(i => i.toolName);
  const lastItem = items[items.length - 1];
  const isComplete =
    lastItem && !lastItem.toolName && items.length > 1 && lastItem.content.length > 100;
  const isError = lastItem?.type === 'error';
  const isActive = !isComplete && !isError;

  const toolItems = items.filter(i => i.toolName);
  const textItems = items.filter(i => !i.toolName);
  const resultItem =
    textItems.length > 0 && textItems[textItems.length - 1].content.length > 100
      ? textItems[textItems.length - 1]
      : null;
  const statusItems = resultItem ? textItems.slice(0, -1) : textItems;

  return (
    <div
      className={[
        styles.daemonBlock,
        isComplete && styles.daemonBlockComplete,
        isError && styles.daemonBlockError,
        isActive && styles.daemonBlockActive,
      ]
        .filter(Boolean)
        .join(' ')}
    >
      <button className={styles.daemonBlockHeader} onClick={() => setExpanded(!expanded)}>
        {expanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
        <span className={styles.daemonBlockRole}>{daemonRole ?? 'Daemon'}</span>
        {isActive && <Loader2 size={14} className={styles.spin} />}
        {isComplete && <CheckCircle2 size={14} className={styles.successIcon} />}
        {isError && <AlertCircle size={14} className={styles.errorIcon} />}
        {isActive && onCancel && (
          <button
            className={styles.daemonCancelBtn}
            onClick={e => {
              e.stopPropagation();
              onCancel(daemonId);
            }}
            title="Stop daemon"
          >
            <X size={14} />
          </button>
        )}
      </button>

      {expanded && (
        <div className={styles.daemonBlockItems}>
          {/* Status/thinking items */}
          {statusItems.length > 0 && (
            <div className={styles.statusList}>
              {statusItems.map(item => (
                <div key={item.id} className={styles.statusItem}>
                  {item.content}
                </div>
              ))}
            </div>
          )}

          {/* Tool timeline */}
          {hasToolCalls && (
            <div className={styles.toolTimeline}>
              {toolItems.map(item => (
                <div key={item.id} className={styles.toolItem}>
                  <div
                    className={[styles.toolDot, item.toolResult && styles.toolDotSuccess]
                      .filter(Boolean)
                      .join(' ')}
                  />
                  <Wrench
                    size={12}
                    style={{ flexShrink: 0, color: 'var(--color-text-tertiary)' }}
                  />
                  <div className={styles.toolInfo}>
                    <span className={styles.toolName}>{item.toolName}</span>
                    {item.toolResult && (
                      <span className={styles.toolResult}>{item.toolResult.slice(0, 200)}</span>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* Final result */}
          {resultItem && (
            <div className={styles.daemonResult}>
              <MarkdownRenderer content={resultItem.content} />
            </div>
          )}

          {/* Errors */}
          {items
            .filter(i => i.type === 'error')
            .map(item => (
              <div key={item.id} className={styles.daemonErrorMsg}>
                {item.content}
              </div>
            ))}
        </div>
      )}
    </div>
  );
});
