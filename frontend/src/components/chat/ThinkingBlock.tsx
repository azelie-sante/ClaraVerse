import { ChevronDown } from 'lucide-react';
import styles from './ThinkingBlock.module.css';

interface ThinkingBlockProps {
  reasoning: string;
  isStreaming: boolean;
  isExpanded: boolean;
  thinkingVerb: string;
  onToggle: () => void;
  messageId: string;
}

export function ThinkingBlock({
  reasoning,
  isStreaming,
  isExpanded,
  thinkingVerb,
  onToggle,
  messageId,
}: ThinkingBlockProps) {
  return (
    <div className={styles.wrapper} data-thinking-id={messageId}>
      {/* Header row — clickable */}
      <div className={styles.headerRow} onClick={onToggle}>
        <div className={styles.iconCol}>
          <div className={`${styles.icon} ${isStreaming ? styles.iconActive : ''}`}>
            {isStreaming ? <span className={styles.rippleRing}>○</span> : <span>○</span>}
          </div>
        </div>
        <span className={`${styles.titleText} ${isStreaming ? styles.titleTextActive : ''}`}>
          {isStreaming ? thinkingVerb : 'Thought process'}
        </span>
        <ChevronDown
          size={14}
          className={styles.chevron}
          style={{
            transform: isExpanded ? 'rotate(180deg)' : 'rotate(0deg)',
          }}
        />
      </div>

      {/* Expandable content with connector line */}
      {isExpanded && (
        <div className={styles.contentWrapper}>
          <div className={styles.connectorCol}>
            <div className={styles.connector} />
          </div>
          <div className={styles.content} data-thinking-content={messageId}>
            {reasoning}
          </div>
        </div>
      )}
    </div>
  );
}

ThinkingBlock.displayName = 'ThinkingBlock';
