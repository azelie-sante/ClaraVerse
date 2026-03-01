import { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { ChevronLeft, ChevronRight, X, Pencil, ArrowRight, Check } from 'lucide-react';
import type { ActivePrompt, PromptAnswer } from '@/types/interactivePrompt';
import type { InteractiveQuestion } from '@/types/websocket';
import styles from './InteractivePromptMessage.module.css';

interface InteractivePromptMessageProps {
  prompt: ActivePrompt;
  onSubmit: (answers: Record<string, PromptAnswer>) => void;
  onSkip: () => void;
  isSubmitting?: boolean;
}

const easeOut: [number, number, number, number] = [0.4, 0, 0.2, 1];

const slideVariants = {
  enter: (dir: number) => ({
    x: dir > 0 ? 60 : -60,
    opacity: 0,
  }),
  center: {
    x: 0,
    opacity: 1,
  },
  exit: (dir: number) => ({
    x: dir > 0 ? -60 : 60,
    opacity: 0,
  }),
};

// ─── Validation ───

function validateQuestion(
  question: InteractiveQuestion,
  answer: PromptAnswer | undefined
): string | null {
  if (question.required) {
    if (!answer || answer.value === '' || answer.value === null || answer.value === undefined) {
      return 'This field is required';
    }
    if (Array.isArray(answer.value) && answer.value.length === 0) {
      return 'Please select at least one option';
    }
  }

  if (!answer || !answer.value) return null;

  const { validation } = question;
  if (!validation) return null;

  if (question.type === 'text') {
    const v = String(answer.value);
    if (validation.min_length && v.length < validation.min_length) {
      return `Minimum ${validation.min_length} characters`;
    }
    if (validation.max_length && v.length > validation.max_length) {
      return `Maximum ${validation.max_length} characters`;
    }
    if (validation.pattern && !new RegExp(validation.pattern).test(v)) {
      return 'Invalid format';
    }
  }

  if (question.type === 'number') {
    const n = Number(answer.value);
    if (validation.min !== undefined && n < validation.min)
      return `Minimum value is ${validation.min}`;
    if (validation.max !== undefined && n > validation.max)
      return `Maximum value is ${validation.max}`;
  }

  return null;
}

export function InteractivePromptMessage({
  prompt,
  onSubmit,
  onSkip,
  isSubmitting = false,
}: InteractivePromptMessageProps) {
  const [currentIndex, setCurrentIndex] = useState(0);
  const [direction, setDirection] = useState<1 | -1>(1);
  const [answers, setAnswers] = useState<Record<string, PromptAnswer>>({});
  const [focusedOptionIndex, setFocusedOptionIndex] = useState(0);
  const [otherInputActive, setOtherInputActive] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  const cardRef = useRef<HTMLDivElement>(null);
  const otherInputRef = useRef<HTMLInputElement>(null);
  const textInputRef = useRef<HTMLInputElement>(null);
  const autoAdvanceTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const totalQuestions = prompt.questions.length;
  const currentQuestion = prompt.questions[currentIndex];
  const isFirst = currentIndex === 0;
  const isLast = currentIndex === totalQuestions - 1;
  const currentAnswer = answers[currentQuestion?.id];

  // Options count including "other"
  const totalOptions = useMemo(() => {
    if (!currentQuestion) return 0;
    const base = currentQuestion.options?.length ?? 0;
    return base + (currentQuestion.allow_other ? 1 : 0);
  }, [currentQuestion]);

  // Reset state when prompt changes
  useEffect(() => {
    const initial: Record<string, PromptAnswer> = {};
    prompt.questions.forEach(q => {
      if (q.default_value !== undefined) {
        initial[q.id] = { questionId: q.id, value: q.default_value, isOther: false };
      }
    });
    setAnswers(initial);
    setCurrentIndex(0);
    setDirection(1);
    setFocusedOptionIndex(0);
    setOtherInputActive(false);
    setErrors({});

    setTimeout(() => {
      cardRef.current?.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    }, 100);
  }, [prompt.promptId, prompt.questions]);

  // Auto-focus text input when a text/number question appears
  useEffect(() => {
    if (currentQuestion?.type === 'text' || currentQuestion?.type === 'number') {
      // Use requestAnimationFrame to ensure the DOM has rendered
      requestAnimationFrame(() => {
        textInputRef.current?.focus();
      });
    }
  }, [currentIndex, currentQuestion?.type]);

  // Focus other input when activated
  useEffect(() => {
    if (otherInputActive) {
      setTimeout(() => otherInputRef.current?.focus(), 50);
    }
  }, [otherInputActive]);

  // Cleanup auto-advance timer
  useEffect(() => {
    return () => {
      if (autoAdvanceTimer.current) clearTimeout(autoAdvanceTimer.current);
    };
  }, []);

  // ─── Answer helpers ───

  const setAnswer = useCallback(
    (questionId: string, value: PromptAnswer['value'], isOther = false, otherText?: string) => {
      setAnswers(prev => ({
        ...prev,
        [questionId]: { questionId, value, isOther, otherText },
      }));
      // Clear error on change
      setErrors(prev => {
        if (!prev[questionId]) return prev;
        const next = { ...prev };
        delete next[questionId];
        return next;
      });
    },
    []
  );

  // ─── Navigation ───

  const goNext = useCallback(
    (overrideAnswers?: Record<string, PromptAnswer>) => {
      if (!currentQuestion) return;

      const resolvedAnswers = overrideAnswers ?? answers;
      const error = validateQuestion(currentQuestion, resolvedAnswers[currentQuestion.id]);
      if (error) {
        setErrors(prev => ({ ...prev, [currentQuestion.id]: error }));
        return;
      }

      if (isLast) {
        onSubmit(resolvedAnswers);
      } else {
        setDirection(1);
        setCurrentIndex(prev => prev + 1);
        setFocusedOptionIndex(0);
        setOtherInputActive(false);
      }
    },
    [currentQuestion, answers, isLast, onSubmit]
  );

  const goPrev = useCallback(() => {
    if (!isFirst) {
      setDirection(-1);
      setCurrentIndex(prev => prev - 1);
      setFocusedOptionIndex(0);
      setOtherInputActive(false);
    }
  }, [isFirst]);

  // ─── Option selection ───

  const handleSelectOption = useCallback(
    (questionId: string, option: string) => {
      const updatedAnswers = {
        ...answers,
        [questionId]: { questionId, value: option, isOther: false } as PromptAnswer,
      };
      setAnswers(updatedAnswers);
      setOtherInputActive(false);
      // Clear error
      setErrors(prev => {
        if (!prev[questionId]) return prev;
        const next = { ...prev };
        delete next[questionId];
        return next;
      });

      // Auto-advance for single-select, passing fresh answers
      if (currentQuestion?.type === 'select') {
        if (autoAdvanceTimer.current) clearTimeout(autoAdvanceTimer.current);
        autoAdvanceTimer.current = setTimeout(() => {
          goNext(updatedAnswers);
        }, 300);
      }
    },
    [currentQuestion?.type, answers, goNext]
  );

  const toggleMultiSelectOption = useCallback(
    (questionId: string, option: string) => {
      const current = (answers[questionId]?.value as string[]) || [];
      const next = current.includes(option)
        ? current.filter(v => v !== option)
        : [...current, option];
      setAnswer(questionId, next, false);
    },
    [answers, setAnswer]
  );

  // ─── Keyboard handler ───

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (isSubmitting || !currentQuestion) return;

      if (e.key === 'Escape') {
        if (prompt.allowSkip) onSkip();
        return;
      }

      // ⌘+Enter or Ctrl+Enter to submit/advance
      if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
        e.preventDefault();
        goNext();
        return;
      }

      const q = currentQuestion;

      // ArrowLeft / ArrowRight for prev/next (only when not in a text input)
      const inTextInput = otherInputActive || q.type === 'text' || q.type === 'number';
      if (e.key === 'ArrowLeft' && !inTextInput) {
        e.preventDefault();
        goPrev();
        return;
      }
      if (e.key === 'ArrowRight' && !inTextInput) {
        e.preventDefault();
        goNext();
        return;
      }

      // Option-based questions (select / multi-select)
      if ((q.type === 'select' || q.type === 'multi-select') && q.options && !otherInputActive) {
        if (e.key === 'ArrowUp') {
          e.preventDefault();
          setFocusedOptionIndex(prev => (prev - 1 + totalOptions) % totalOptions);
        } else if (e.key === 'ArrowDown') {
          e.preventDefault();
          setFocusedOptionIndex(prev => (prev + 1) % totalOptions);
        } else if (e.key === 'Enter') {
          e.preventDefault();
          const optLen = q.options.length;
          if (focusedOptionIndex < optLen) {
            if (q.type === 'select') {
              handleSelectOption(q.id, q.options[focusedOptionIndex]);
            } else {
              toggleMultiSelectOption(q.id, q.options[focusedOptionIndex]);
            }
          } else if (q.allow_other) {
            // Focus "Something else" input
            setOtherInputActive(true);
          }
        }

        // Number keys 1-9 quick select
        const num = parseInt(e.key);
        if (num >= 1 && num <= totalOptions) {
          e.preventDefault();
          const optLen = q.options.length;
          if (num <= optLen) {
            if (q.type === 'select') {
              handleSelectOption(q.id, q.options[num - 1]);
            } else {
              toggleMultiSelectOption(q.id, q.options[num - 1]);
            }
          } else if (q.allow_other) {
            setOtherInputActive(true);
          }
        }
      }

      // Text/number: Enter to advance
      if ((q.type === 'text' || q.type === 'number') && e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        goNext();
      }

      // Checkbox: Enter or Space to toggle
      if (q.type === 'checkbox' && (e.key === 'Enter' || e.key === ' ')) {
        e.preventDefault();
        const current = (answers[q.id]?.value as boolean) || false;
        setAnswer(q.id, !current);
      }
    };

    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [
    isSubmitting,
    currentQuestion,
    prompt.allowSkip,
    onSkip,
    goNext,
    goPrev,
    otherInputActive,
    totalOptions,
    focusedOptionIndex,
    handleSelectOption,
    toggleMultiSelectOption,
    answers,
    setAnswer,
  ]);

  // ─── Other input Enter handler ───

  const handleOtherKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter') {
        e.preventDefault();
        goNext();
      }
      if (e.key === 'Escape') {
        setOtherInputActive(false);
      }
    },
    [goNext]
  );

  // ─── Question renderers ───

  const renderSelectQuestion = (q: InteractiveQuestion) => {
    const selected = currentAnswer?.value as string | undefined;

    return (
      <div className={styles.optionsList}>
        {q.options?.map((option, idx) => {
          const isFocused = focusedOptionIndex === idx;
          const isSelected = selected === option && !currentAnswer?.isOther;
          return (
            <div
              key={option}
              className={`${styles.optionRow} ${isFocused ? styles.optionRowFocused : ''} ${isSelected ? styles.optionRowSelected : ''}`}
              onClick={() => handleSelectOption(q.id, option)}
            >
              <div className={styles.optionBadge}>{idx + 1}</div>
              <span className={styles.optionLabel}>{option}</span>
              <ArrowRight size={14} className={styles.optionArrow} />
            </div>
          );
        })}
        {q.allow_other && (
          <div
            className={`${styles.otherRow} ${focusedOptionIndex === (q.options?.length ?? 0) ? styles.otherRowFocused : ''}`}
            onClick={() => {
              setOtherInputActive(true);
              setAnswer(q.id, currentAnswer?.otherText || '', true, currentAnswer?.otherText);
            }}
          >
            <div className={styles.otherIcon}>
              <Pencil size={14} />
            </div>
            {otherInputActive ? (
              <input
                ref={otherInputRef}
                className={styles.otherInput}
                placeholder="Something else"
                value={currentAnswer?.otherText || ''}
                onChange={e => setAnswer(q.id, e.target.value, true, e.target.value)}
                onKeyDown={handleOtherKeyDown}
                disabled={isSubmitting}
              />
            ) : (
              <span className={styles.optionLabel} style={{ opacity: 0.5 }}>
                Something else
              </span>
            )}
          </div>
        )}
      </div>
    );
  };

  const renderMultiSelectQuestion = (q: InteractiveQuestion) => {
    const selected = (currentAnswer?.value as string[]) || [];

    return (
      <div className={styles.optionsList}>
        {q.options?.map((option, idx) => {
          const isFocused = focusedOptionIndex === idx;
          const isChecked = selected.includes(option);
          return (
            <div
              key={option}
              className={`${styles.optionRow} ${isFocused ? styles.optionRowFocused : ''} ${isChecked ? styles.optionRowSelected : ''}`}
              onClick={() => toggleMultiSelectOption(q.id, option)}
            >
              <div
                className={`${styles.checkIndicator} ${isChecked ? styles.checkIndicatorChecked : ''}`}
              >
                {isChecked && <Check size={12} color="white" />}
              </div>
              <span className={styles.optionLabel}>{option}</span>
            </div>
          );
        })}
        {q.allow_other && (
          <div
            className={`${styles.otherRow} ${focusedOptionIndex === (q.options?.length ?? 0) ? styles.otherRowFocused : ''}`}
            onClick={() => {
              setOtherInputActive(true);
              if (!currentAnswer?.isOther) {
                setAnswer(q.id, selected, true, currentAnswer?.otherText || '');
              }
            }}
          >
            <div
              className={`${styles.checkIndicator} ${currentAnswer?.isOther ? styles.checkIndicatorChecked : ''}`}
            >
              {currentAnswer?.isOther && <Check size={12} color="white" />}
            </div>
            {otherInputActive ? (
              <input
                ref={otherInputRef}
                className={styles.otherInput}
                placeholder="Something else"
                value={currentAnswer?.otherText || ''}
                onChange={e => {
                  const text = e.target.value;
                  const vals = selected.filter(v => v !== currentAnswer?.otherText);
                  if (text) vals.push(text);
                  setAnswers(prev => ({
                    ...prev,
                    [q.id]: { questionId: q.id, value: vals, isOther: true, otherText: text },
                  }));
                }}
                onKeyDown={handleOtherKeyDown}
                disabled={isSubmitting}
              />
            ) : (
              <span className={styles.optionLabel} style={{ opacity: 0.5 }}>
                Something else
              </span>
            )}
          </div>
        )}
      </div>
    );
  };

  const renderTextQuestion = (q: InteractiveQuestion) => {
    const val = (currentAnswer?.value as string) ?? '';
    const hasError = !!errors[q.id];

    return (
      <div>
        <input
          ref={textInputRef}
          type={q.type === 'number' ? 'number' : 'text'}
          className={`${styles.textInput} ${hasError ? styles.textInputError : ''}`}
          placeholder={q.placeholder || 'Type your answer...'}
          value={val}
          onChange={e => {
            const v = q.type === 'number' ? e.target.valueAsNumber || 0 : e.target.value;
            setAnswer(q.id, v);
          }}
          autoFocus
          disabled={isSubmitting}
          min={q.validation?.min}
          max={q.validation?.max}
          maxLength={q.validation?.max_length}
        />
        {q.validation?.max_length && (
          <div className={styles.inputHint}>
            {val.length} / {q.validation.max_length}
          </div>
        )}
      </div>
    );
  };

  const renderCheckboxQuestion = (q: InteractiveQuestion) => {
    const checked = (currentAnswer?.value as boolean) || false;

    return (
      <div
        className={styles.toggleRow}
        onClick={() => {
          setAnswer(q.id, !checked);
          // Auto-advance after brief delay
          setTimeout(() => goNext(), 300);
        }}
      >
        <div className={`${styles.toggleTrack} ${checked ? styles.toggleTrackActive : ''}`}>
          <div className={styles.toggleKnob} />
        </div>
        <span className={styles.toggleLabel}>{checked ? 'Yes' : 'No'}</span>
      </div>
    );
  };

  const renderCurrentQuestion = () => {
    if (!currentQuestion) return null;
    switch (currentQuestion.type) {
      case 'select':
        return renderSelectQuestion(currentQuestion);
      case 'multi-select':
        return renderMultiSelectQuestion(currentQuestion);
      case 'text':
      case 'number':
        return renderTextQuestion(currentQuestion);
      case 'checkbox':
        return renderCheckboxQuestion(currentQuestion);
      default:
        return null;
    }
  };

  // ─── Keyboard hint text ───

  const keyboardHint = useMemo(() => {
    if (!currentQuestion) return null;
    const q = currentQuestion;
    if (q.type === 'select') {
      return (
        <>
          <span>↑↓ navigate</span>
          <span className={styles.hintSep}>·</span>
          <span>Enter select</span>
          <span className={styles.hintSep}>·</span>
          <span>←→ prev/next</span>
          {prompt.allowSkip && (
            <>
              <span className={styles.hintSep}>·</span>
              <span>Esc skip</span>
            </>
          )}
        </>
      );
    }
    if (q.type === 'multi-select') {
      return (
        <>
          <span>↑↓ navigate</span>
          <span className={styles.hintSep}>·</span>
          <span>Enter toggle</span>
          <span className={styles.hintSep}>·</span>
          <span>←→ prev/next</span>
          <span className={styles.hintSep}>·</span>
          <span>⌘ Enter submit</span>
          {prompt.allowSkip && (
            <>
              <span className={styles.hintSep}>·</span>
              <span>Esc skip</span>
            </>
          )}
        </>
      );
    }
    if (q.type === 'text' || q.type === 'number') {
      return (
        <>
          <span>Enter continue</span>
          {prompt.allowSkip && (
            <>
              <span className={styles.hintSep}>·</span>
              <span>Esc skip</span>
            </>
          )}
        </>
      );
    }
    if (q.type === 'checkbox') {
      return (
        <>
          <span>Enter or Space toggle</span>
          <span className={styles.hintSep}>·</span>
          <span>←→ prev/next</span>
          {prompt.allowSkip && (
            <>
              <span className={styles.hintSep}>·</span>
              <span>Esc skip</span>
            </>
          )}
        </>
      );
    }
    return null;
  }, [currentQuestion, prompt.allowSkip]);

  if (!currentQuestion) return null;

  const multiSelectCount =
    currentQuestion.type === 'multi-select' ? ((currentAnswer?.value as string[]) || []).length : 0;

  return (
    <motion.div
      ref={cardRef}
      className={styles.card}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -10 }}
      transition={{ duration: 0.3 }}
    >
      {/* Header */}
      <div className={styles.header}>
        <span className={styles.questionText}>
          {currentQuestion.label}
          {currentQuestion.required && (
            <span style={{ color: 'var(--color-error, #ff453a)', marginLeft: 4 }}>*</span>
          )}
        </span>
        <div className={styles.headerRight}>
          <button
            className={styles.navButton}
            onClick={goPrev}
            disabled={isFirst || isSubmitting}
            aria-label="Previous question"
          >
            <ChevronLeft size={16} />
          </button>
          <span className={styles.counter}>
            {currentIndex + 1} of {totalQuestions}
          </span>
          <button
            className={`${styles.navButton} ${isLast ? styles.navButtonSubmit : ''}`}
            onClick={goNext}
            disabled={isSubmitting}
            aria-label={isLast ? 'Submit answers' : 'Next question'}
          >
            {isLast ? <ArrowRight size={16} /> : <ChevronRight size={16} />}
          </button>
          {prompt.allowSkip && (
            <button
              className={styles.closeButton}
              onClick={onSkip}
              disabled={isSubmitting}
              aria-label="Skip prompt"
            >
              <X size={16} />
            </button>
          )}
        </div>
      </div>

      {/* Description (first question only) */}
      {prompt.description && currentIndex === 0 && (
        <p className={styles.description}>{prompt.description}</p>
      )}

      {/* Animated question body */}
      <AnimatePresence mode="wait" custom={direction}>
        <motion.div
          key={currentQuestion.id}
          custom={direction}
          variants={slideVariants}
          initial="enter"
          animate="center"
          exit="exit"
          transition={{ duration: 0.25, ease: easeOut }}
          className={styles.questionSlide}
        >
          {renderCurrentQuestion()}
        </motion.div>
      </AnimatePresence>

      {/* Footer */}
      <div className={styles.footer}>
        <div className={styles.footerLeft}>
          {currentQuestion.type === 'multi-select' && multiSelectCount > 0 && (
            <span className={styles.selectedCount}>{multiSelectCount} selected</span>
          )}
          {errors[currentQuestion.id] && (
            <span className={styles.errorText}>{errors[currentQuestion.id]}</span>
          )}
        </div>
        <div className={styles.footerRight}>
          {prompt.allowSkip && (
            <button className={styles.skipButton} onClick={onSkip} disabled={isSubmitting}>
              Skip
            </button>
          )}
          {currentQuestion.type === 'multi-select' && (
            <button
              className={styles.footerNextButton}
              onClick={goNext}
              disabled={isSubmitting}
              aria-label={isLast ? 'Submit answers' : 'Next question'}
            >
              <ArrowRight size={14} />
            </button>
          )}
        </div>
      </div>

      {/* Keyboard hints */}
      <div className={styles.keyboardHints}>{keyboardHint}</div>
    </motion.div>
  );
}

InteractivePromptMessage.displayName = 'InteractivePromptMessage';
