import { memo, useState, useMemo, useCallback, type ReactNode } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { ChevronDown, ChevronRight, CheckCircle2 } from 'lucide-react';
import type { ToolCall } from '@/types/chat';
import { getIconByName } from '@/utils/iconMapper';
import {
  getToolActionText,
  getToolCompletedText,
  getToolRepeatText,
  getToolIcon,
  getResultCount,
  parseSearchResults,
  buildCompletedSummary,
} from '@/utils/toolUtils';
import styles from './ToolStatusPill.module.css';

interface ToolStatusPillProps {
  toolCalls: ToolCall[];
  isStreaming: boolean;
  messageId: string;
  statusUpdate?: string;
  onToggleToolExpansion: (messageId: string, toolId: string) => void;
  onToggleAllTools: (messageId: string, toolCalls: ToolCall[], expand: boolean) => void;
  renderToolContent: (tool: ToolCall) => ReactNode;
  renderCompletedToolContent: (tool: ToolCall) => ReactNode;
}

const MAX_VISIBLE = 3;
const easeOut = [0.4, 0, 0.2, 1] as const;

interface Step {
  key: string;
  text: string;
  iconName: string;
}

function toolToStreamingStep(tool: ToolCall, occurrence: number): Step {
  // If this tool already completed while message is still streaming,
  // show past-tense text instead of present-tense
  if (tool.status === 'completed') {
    return toolToCompletedStep(tool);
  }

  const actionText = getToolActionText(tool.name);
  let text =
    actionText.startsWith('Using ') && tool.displayName ? `Using ${tool.displayName}` : actionText;

  // Vary text for repeated tool calls (e.g. multiple searches)
  if (occurrence >= 2) {
    text = getToolRepeatText(tool.name, occurrence) ?? text;
  }

  return {
    key: tool.id,
    text,
    iconName: tool.icon || getToolIcon(tool.name),
  };
}

function toolToCompletedStep(tool: ToolCall): Step {
  const completedText = getToolCompletedText(tool.name);
  const text =
    completedText.startsWith('Used ') && tool.displayName
      ? `Used ${tool.displayName}`
      : completedText;
  return {
    key: tool.id,
    text,
    iconName: tool.icon || getToolIcon(tool.name),
  };
}

export const ToolStatusPill = memo(function ToolStatusPill({
  toolCalls,
  isStreaming,
  messageId: _messageId,
  statusUpdate,
  onToggleToolExpansion: _onToggleToolExpansion,
  onToggleAllTools: _onToggleAllTools,
  renderToolContent: _renderToolContent,
  renderCompletedToolContent: _renderCompletedToolContent,
}: ToolStatusPillProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [expandedToolIds, setExpandedToolIds] = useState<Set<string>>(new Set());

  const toggleToolDetail = useCallback((toolId: string) => {
    setExpandedToolIds(prev => {
      const next = new Set(prev);
      if (next.has(toolId)) next.delete(toolId);
      else next.add(toolId);
      return next;
    });
  }, []);

  // ─── Streaming steps ───
  const streamingSteps = useMemo<Step[]>(() => {
    if (!isStreaming) return [];
    if (statusUpdate && toolCalls.length === 0) {
      return [{ key: 'status', text: statusUpdate, iconName: '__spinner__' }];
    }
    const counts: Record<string, number> = {};
    return toolCalls.map(tool => {
      counts[tool.name] = (counts[tool.name] || 0) + 1;
      return toolToStreamingStep(tool, counts[tool.name]);
    });
  }, [isStreaming, toolCalls, statusUpdate]);

  const visibleSteps = streamingSteps.slice(-MAX_VISIBLE);
  const hiddenCount = Math.max(0, streamingSteps.length - MAX_VISIBLE);

  // ─── Completed steps ───
  const completedSteps = useMemo(() => {
    if (isStreaming) return [];
    return toolCalls
      .filter(t => t.status === 'completed')
      .map(t => ({
        ...toolToCompletedStep(t),
        tool: t,
      }));
  }, [toolCalls, isStreaming]);

  // Search results data keyed by tool id
  const searchResultsMap = useMemo(() => {
    if (isStreaming) return new Map<string, ReturnType<typeof parseSearchResults>>();
    const map = new Map<string, ReturnType<typeof parseSearchResults>>();
    for (const t of toolCalls) {
      if (t.name === 'search_web' && t.status === 'completed' && t.result) {
        map.set(t.id, parseSearchResults(t.result, t.query));
      }
    }
    return map;
  }, [toolCalls, isStreaming]);

  const completedSummary = useMemo(() => {
    if (isStreaming) return '';
    return buildCompletedSummary(toolCalls);
  }, [toolCalls, isStreaming]);

  // Nothing to show
  if (!isStreaming && toolCalls.length === 0) return null;
  if (isStreaming && streamingSteps.length === 0) return null;

  return (
    <div className={styles.wrapper}>
      {isStreaming ? (
        /* ═══ STREAMING: vertical step list, latest 3 ═══ */
        <div className={styles.stepList}>
          <AnimatePresence>
            {hiddenCount > 0 && (
              <motion.div
                key="hidden"
                className={styles.collapsedRow}
                initial={{ opacity: 0, height: 0 }}
                animate={{ opacity: 1, height: 'auto' }}
                exit={{ opacity: 0, height: 0 }}
                transition={{ duration: 0.2 }}
              >
                <div className={styles.collapsedIconCol}>
                  <span className={styles.dot} />
                  <span className={styles.dot} />
                  <span className={styles.dot} />
                </div>
                <span className={styles.collapsedLabel}>
                  +{hiddenCount} {hiddenCount === 1 ? 'step' : 'steps'}
                </span>
              </motion.div>
            )}
          </AnimatePresence>

          <AnimatePresence mode="popLayout" initial={false}>
            {visibleSteps.map((step, idx) => {
              const isLast = idx === visibleSteps.length - 1;
              const isSpinner = step.iconName === '__spinner__';
              const StepIcon = isSpinner ? null : getIconByName(step.iconName);
              return (
                <motion.div
                  key={step.key}
                  className={styles.stepRow}
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -6 }}
                  transition={{ duration: 0.22, ease: easeOut }}
                  layout
                >
                  <div className={styles.stepIconCol}>
                    <div className={`${styles.stepIcon} ${isLast ? styles.stepIconActive : ''}`}>
                      {isSpinner ? (
                        <span className={styles.spinnerAsterisk}>✦</span>
                      ) : (
                        StepIcon && <StepIcon size={14} />
                      )}
                    </div>
                    {!isLast && <div className={styles.connector} />}
                  </div>
                  <span className={`${styles.stepText} ${isLast ? styles.stepTextActive : ''}`}>
                    {step.text}
                  </span>
                </motion.div>
              );
            })}
          </AnimatePresence>
        </div>
      ) : (
        /* ═══ COMPLETED: same vertical step list, collapsible ═══ */
        <>
          {/* Toggle header */}
          <div className={styles.completedHeader} onClick={() => setIsExpanded(prev => !prev)}>
            <span className={styles.completedHeaderText}>{completedSummary}</span>
            <motion.div
              animate={{ rotate: isExpanded ? 180 : 0 }}
              transition={{ duration: 0.2, ease: easeOut }}
              className={styles.completedChevron}
            >
              <ChevronDown size={14} />
            </motion.div>
          </div>

          {/* Expanded: full step list */}
          <AnimatePresence>
            {isExpanded && (
              <motion.div
                className={styles.stepList}
                initial={{ opacity: 0, height: 0 }}
                animate={{ opacity: 1, height: 'auto' }}
                exit={{ opacity: 0, height: 0 }}
                transition={{ duration: 0.25, ease: easeOut }}
                style={{ overflow: 'hidden' }}
              >
                {completedSteps.map((step, idx) => {
                  const isLast = idx === completedSteps.length - 1;
                  const StepIcon = getIconByName(step.iconName);
                  const parsed = searchResultsMap.get(step.tool.id);
                  const isSearch = step.tool.name === 'search_web' && parsed;
                  const detail =
                    !isSearch && step.tool.result
                      ? getResultCount(step.tool.result, step.tool.query)
                      : '';
                  const hasResult = !isSearch && !!step.tool.result;
                  const isToolExpanded = expandedToolIds.has(step.tool.id);

                  return (
                    <div key={step.key} className={styles.completedStepBlock}>
                      <div
                        className={`${styles.stepRow} ${hasResult ? styles.stepRowClickable : ''}`}
                        onClick={hasResult ? () => toggleToolDetail(step.tool.id) : undefined}
                      >
                        <div className={styles.stepIconCol}>
                          <div className={styles.stepIcon}>
                            <StepIcon size={14} />
                          </div>
                          {(!isLast || isSearch || isToolExpanded) && (
                            <div className={styles.connector} />
                          )}
                        </div>
                        <div className={styles.completedStepContent}>
                          <span className={styles.completedStepText}>{step.text}</span>
                          {step.tool.query && (
                            <span className={styles.completedStepQuery}>{step.tool.query}</span>
                          )}
                          {/* Right-aligned: result count or detail */}
                          {isSearch && parsed && (
                            <span className={styles.completedStepRight}>{parsed.resultCount}</span>
                          )}
                          {detail && <span className={styles.completedStepRight}>{detail}</span>}
                          {hasResult && (
                            <motion.div
                              animate={{ rotate: isToolExpanded ? 90 : 0 }}
                              transition={{ duration: 0.15 }}
                              className={styles.completedStepChevron}
                            >
                              <ChevronRight size={12} />
                            </motion.div>
                          )}
                        </div>
                      </div>

                      {/* Inline search results */}
                      {isSearch && parsed && parsed.items.length > 0 && (
                        <div className={styles.searchResultsInline}>
                          <div className={styles.searchResultsConnector} />
                          <div className={styles.searchResultsCard}>
                            {parsed.items.map((item, itemIdx) => (
                              <a
                                key={`${step.key}-${itemIdx}`}
                                className={styles.resultItem}
                                href={item.url}
                                target="_blank"
                                rel="noopener noreferrer"
                              >
                                <img
                                  className={styles.favicon}
                                  src={item.faviconUrl}
                                  alt=""
                                  loading="lazy"
                                  onError={e => {
                                    (e.target as HTMLImageElement).style.display = 'none';
                                  }}
                                />
                                <span className={styles.resultTitle}>{item.title}</span>
                                <span className={styles.resultDomain}>{item.domain}</span>
                              </a>
                            ))}
                          </div>
                        </div>
                      )}

                      {/* Expandable tool result detail */}
                      <AnimatePresence>
                        {isToolExpanded && hasResult && (
                          <motion.div
                            className={styles.toolDetailInline}
                            initial={{ opacity: 0, height: 0 }}
                            animate={{ opacity: 1, height: 'auto' }}
                            exit={{ opacity: 0, height: 0 }}
                            transition={{ duration: 0.2, ease: easeOut }}
                          >
                            <div className={styles.toolDetailConnector} />
                            <pre className={styles.toolDetailText}>{step.tool.result}</pre>
                          </motion.div>
                        )}
                      </AnimatePresence>
                    </div>
                  );
                })}

                {/* Done step */}
                <div className={styles.stepRow}>
                  <div className={styles.stepIconCol}>
                    <div className={styles.stepIconDone}>
                      <CheckCircle2 size={14} />
                    </div>
                  </div>
                  <span className={styles.completedStepText}>Done</span>
                </div>
              </motion.div>
            )}
          </AnimatePresence>
        </>
      )}
    </div>
  );
});
