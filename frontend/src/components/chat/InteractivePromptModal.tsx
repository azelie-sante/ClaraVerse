import { useState, useEffect, useCallback } from 'react';
import { createPortal } from 'react-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { X, AlertCircle, CheckCircle } from 'lucide-react';
import type { ActivePrompt, PromptAnswer, ValidationResult } from '@/types/interactivePrompt';
import type { InteractiveQuestion } from '@/types/websocket';
import './InteractivePromptModal.css';

interface InteractivePromptModalProps {
  isOpen: boolean;
  prompt: ActivePrompt | null;
  onSubmit: (answers: Record<string, PromptAnswer>) => void;
  onSkip: () => void;
  isSubmitting?: boolean;
}

export function InteractivePromptModal({
  isOpen,
  prompt,
  onSubmit,
  onSkip,
  isSubmitting = false,
}: InteractivePromptModalProps) {
  const [answers, setAnswers] = useState<Record<string, PromptAnswer>>({});
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [touchedFields, setTouchedFields] = useState<Set<string>>(new Set());

  // Reset state when modal opens
  useEffect(() => {
    if (isOpen && prompt) {
      const initialAnswers: Record<string, PromptAnswer> = {};
      prompt.questions.forEach(question => {
        if (question.default_value !== undefined) {
          initialAnswers[question.id] = {
            questionId: question.id,
            value: question.default_value,
            isOther: false,
          };
        }
      });
      setAnswers(initialAnswers);
      setErrors({});
      setTouchedFields(new Set());
    }
  }, [isOpen, prompt]);

  // Validate a single answer
  const validateAnswer = useCallback(
    (question: InteractiveQuestion, answer: PromptAnswer | undefined): string | null => {
      // Required validation
      if (question.required) {
        if (!answer || answer.value === '' || answer.value === null || answer.value === undefined) {
          return 'This field is required';
        }

        // Check for empty arrays in multi-select
        if (Array.isArray(answer.value) && answer.value.length === 0) {
          return 'Please select at least one option';
        }
      }

      if (!answer || !answer.value) {
        return null; // Optional field, no validation needed
      }

      const { validation } = question;
      if (!validation) return null;

      // Type-specific validation
      if (question.type === 'text') {
        const textValue = String(answer.value);

        if (validation.min_length && textValue.length < validation.min_length) {
          return `Minimum length is ${validation.min_length} characters`;
        }

        if (validation.max_length && textValue.length > validation.max_length) {
          return `Maximum length is ${validation.max_length} characters`;
        }

        if (validation.pattern) {
          const regex = new RegExp(validation.pattern);
          if (!regex.test(textValue)) {
            return 'Invalid format';
          }
        }
      }

      if (question.type === 'number') {
        const numValue = Number(answer.value);

        if (validation.min !== undefined && numValue < validation.min) {
          return `Minimum value is ${validation.min}`;
        }

        if (validation.max !== undefined && numValue > validation.max) {
          return `Maximum value is ${validation.max}`;
        }
      }

      return null;
    },
    []
  );

  // Validate all answers
  const validateAll = useCallback((): ValidationResult => {
    if (!prompt) {
      return { isValid: true, errors: {} };
    }

    const newErrors: Record<string, string> = {};

    prompt.questions.forEach(question => {
      const answer = answers[question.id];
      const error = validateAnswer(question, answer);
      if (error) {
        newErrors[question.id] = error;
      }
    });

    return {
      isValid: Object.keys(newErrors).length === 0,
      errors: newErrors,
    };
  }, [prompt, answers, validateAnswer]);

  // Handle answer change
  const handleAnswerChange = (
    questionId: string,
    value: PromptAnswer['value'],
    isOther: boolean = false
  ) => {
    setAnswers(prev => ({
      ...prev,
      [questionId]: {
        questionId,
        value,
        isOther,
      },
    }));

    // Clear error when user starts typing
    if (errors[questionId]) {
      setErrors(prev => {
        const newErrors = { ...prev };
        delete newErrors[questionId];
        return newErrors;
      });
    }
  };

  // Handle "Other" text input change
  const handleOtherTextChange = (questionId: string, otherText: string) => {
    setAnswers(prev => ({
      ...prev,
      [questionId]: {
        ...prev[questionId],
        questionId,
        value: otherText,
        isOther: true,
        otherText,
      },
    }));
  };

  // Handle field blur
  const handleBlur = (questionId: string) => {
    setTouchedFields(prev => new Set(prev).add(questionId));

    const question = prompt?.questions.find(q => q.id === questionId);
    if (question) {
      const error = validateAnswer(question, answers[questionId]);
      if (error) {
        setErrors(prev => ({ ...prev, [questionId]: error }));
      }
    }
  };

  // Handle submit
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    // Mark all fields as touched
    const allQuestionIds = new Set(prompt?.questions.map(q => q.id) || []);
    setTouchedFields(allQuestionIds);

    // Validate all
    const validation = validateAll();

    if (!validation.isValid) {
      setErrors(validation.errors);
      // Scroll to first error
      const firstErrorId = Object.keys(validation.errors)[0];
      const element = document.getElementById(`question-${firstErrorId}`);
      element?.scrollIntoView({ behavior: 'smooth', block: 'center' });
      return;
    }

    onSubmit(answers);
  };

  // Handle skip/close
  const handleClose = () => {
    if (prompt?.allowSkip) {
      onSkip();
    }
  };

  // Handle keyboard shortcuts
  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && prompt?.allowSkip) {
        handleClose();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, prompt?.allowSkip]);

  if (!isOpen || !prompt) return null;

  // Render different question types
  const renderQuestion = (question: InteractiveQuestion) => {
    const answer = answers[question.id];
    const error = touchedFields.has(question.id) ? errors[question.id] : null;
    const hasError = Boolean(error);

    switch (question.type) {
      case 'text':
        return (
          <div key={question.id} id={`question-${question.id}`} className="prompt-question">
            <label htmlFor={`input-${question.id}`} className="prompt-label">
              {question.label}
              {question.required && <span className="required-star">*</span>}
            </label>
            <input
              id={`input-${question.id}`}
              type="text"
              className={`prompt-input ${hasError ? 'error' : ''}`}
              placeholder={question.placeholder}
              value={(answer?.value as string) || ''}
              onChange={e => handleAnswerChange(question.id, e.target.value)}
              onBlur={() => handleBlur(question.id)}
              disabled={isSubmitting}
              maxLength={question.validation?.max_length}
            />
            {hasError && (
              <div className="prompt-error">
                <AlertCircle size={14} />
                <span>{error}</span>
              </div>
            )}
            {question.validation?.max_length && (
              <div className="prompt-hint">
                {((answer?.value as string) || '').length} / {question.validation.max_length}
              </div>
            )}
          </div>
        );

      case 'number':
        return (
          <div key={question.id} id={`question-${question.id}`} className="prompt-question">
            <label htmlFor={`input-${question.id}`} className="prompt-label">
              {question.label}
              {question.required && <span className="required-star">*</span>}
            </label>
            <input
              id={`input-${question.id}`}
              type="number"
              className={`prompt-input ${hasError ? 'error' : ''}`}
              placeholder={question.placeholder}
              value={(answer?.value as number) ?? ''}
              onChange={e => handleAnswerChange(question.id, e.target.valueAsNumber || 0)}
              onBlur={() => handleBlur(question.id)}
              disabled={isSubmitting}
              min={question.validation?.min}
              max={question.validation?.max}
            />
            {hasError && (
              <div className="prompt-error">
                <AlertCircle size={14} />
                <span>{error}</span>
              </div>
            )}
          </div>
        );

      case 'checkbox':
        return (
          <div key={question.id} id={`question-${question.id}`} className="prompt-question">
            <label className="prompt-checkbox-label">
              <input
                type="checkbox"
                className="prompt-checkbox"
                checked={(answer?.value as boolean) || false}
                onChange={e => handleAnswerChange(question.id, e.target.checked)}
                onBlur={() => handleBlur(question.id)}
                disabled={isSubmitting}
              />
              <span>
                {question.label}
                {question.required && <span className="required-star">*</span>}
              </span>
            </label>
            {hasError && (
              <div className="prompt-error">
                <AlertCircle size={14} />
                <span>{error}</span>
              </div>
            )}
          </div>
        );

      case 'select':
        const selectedValue = answer?.value as string;
        const isOtherSelected = answer?.isOther || false;

        return (
          <div key={question.id} id={`question-${question.id}`} className="prompt-question">
            <label className="prompt-label">
              {question.label}
              {question.required && <span className="required-star">*</span>}
            </label>
            <div className="prompt-radio-group">
              {question.options?.map(option => (
                <label key={option} className="prompt-radio-label">
                  <input
                    type="radio"
                    name={`radio-${question.id}`}
                    className="prompt-radio"
                    value={option}
                    checked={selectedValue === option && !isOtherSelected}
                    onChange={() => handleAnswerChange(question.id, option, false)}
                    onBlur={() => handleBlur(question.id)}
                    disabled={isSubmitting}
                  />
                  <span>{option}</span>
                </label>
              ))}
              {question.allow_other && (
                <div className="prompt-other-wrapper">
                  <label className="prompt-radio-label">
                    <input
                      type="radio"
                      name={`radio-${question.id}`}
                      className="prompt-radio"
                      checked={isOtherSelected}
                      onChange={() => {
                        handleAnswerChange(question.id, answer?.otherText || '', true);
                      }}
                      disabled={isSubmitting}
                    />
                    <span>Other</span>
                  </label>
                  {isOtherSelected && (
                    <input
                      type="text"
                      className={`prompt-input prompt-other-input ${hasError ? 'error' : ''}`}
                      placeholder="Please specify..."
                      value={answer?.otherText || ''}
                      onChange={e => handleOtherTextChange(question.id, e.target.value)}
                      onBlur={() => handleBlur(question.id)}
                      disabled={isSubmitting}
                      autoFocus
                    />
                  )}
                </div>
              )}
            </div>
            {hasError && (
              <div className="prompt-error">
                <AlertCircle size={14} />
                <span>{error}</span>
              </div>
            )}
          </div>
        );

      case 'multi-select':
        const selectedValues = (answer?.value as string[]) || [];
        const isOtherChecked = answer?.isOther || false;

        return (
          <div key={question.id} id={`question-${question.id}`} className="prompt-question">
            <label className="prompt-label">
              {question.label}
              {question.required && <span className="required-star">*</span>}
            </label>
            <div className="prompt-checkbox-group">
              {question.options?.map(option => (
                <label key={option} className="prompt-checkbox-label">
                  <input
                    type="checkbox"
                    className="prompt-checkbox"
                    value={option}
                    checked={selectedValues.includes(option)}
                    onChange={e => {
                      const newValues = e.target.checked
                        ? [...selectedValues, option]
                        : selectedValues.filter(v => v !== option);
                      handleAnswerChange(question.id, newValues, false);
                    }}
                    onBlur={() => handleBlur(question.id)}
                    disabled={isSubmitting}
                  />
                  <span>{option}</span>
                </label>
              ))}
              {question.allow_other && (
                <div className="prompt-other-wrapper">
                  <label className="prompt-checkbox-label">
                    <input
                      type="checkbox"
                      className="prompt-checkbox"
                      checked={isOtherChecked}
                      onChange={e => {
                        if (e.target.checked) {
                          handleAnswerChange(question.id, [...selectedValues], true);
                        } else {
                          const newValues = selectedValues.filter(v => v !== answer?.otherText);
                          handleAnswerChange(question.id, newValues, false);
                        }
                      }}
                      disabled={isSubmitting}
                    />
                    <span>Other</span>
                  </label>
                  {isOtherChecked && (
                    <input
                      type="text"
                      className={`prompt-input prompt-other-input ${hasError ? 'error' : ''}`}
                      placeholder="Please specify..."
                      value={answer?.otherText || ''}
                      onChange={e => {
                        const otherText = e.target.value;
                        const newValues = selectedValues.filter(v => v !== answer?.otherText);
                        if (otherText) {
                          newValues.push(otherText);
                        }
                        setAnswers(prev => ({
                          ...prev,
                          [question.id]: {
                            questionId: question.id,
                            value: newValues,
                            isOther: true,
                            otherText,
                          },
                        }));
                      }}
                      onBlur={() => handleBlur(question.id)}
                      disabled={isSubmitting}
                      autoFocus
                    />
                  )}
                </div>
              )}
            </div>
            {hasError && (
              <div className="prompt-error">
                <AlertCircle size={14} />
                <span>{error}</span>
              </div>
            )}
          </div>
        );

      default:
        return null;
    }
  };

  const validation = validateAll();
  const canSubmit = validation.isValid && !isSubmitting;

  return createPortal(
    <AnimatePresence>
      {isOpen && (
        <motion.div
          className="interactive-prompt-overlay"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          onClick={handleClose}
        >
          <motion.div
            className="interactive-prompt-modal"
            initial={{ scale: 0.9, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            exit={{ scale: 0.9, opacity: 0 }}
            onClick={e => e.stopPropagation()}
          >
            {prompt.allowSkip && (
              <button
                className="prompt-close-button"
                onClick={handleClose}
                disabled={isSubmitting}
                aria-label="Close"
              >
                <X size={20} />
              </button>
            )}

            <div className="prompt-header">
              <AlertCircle size={24} className="prompt-icon" />
              <h2 className="prompt-title">{prompt.title}</h2>
            </div>

            {prompt.description && <p className="prompt-description">{prompt.description}</p>}

            <form onSubmit={handleSubmit} className="prompt-form">
              <div className="prompt-questions">{prompt.questions.map(renderQuestion)}</div>

              <div className="prompt-actions">
                {prompt.allowSkip && (
                  <button
                    type="button"
                    className="prompt-button prompt-button-secondary"
                    onClick={handleClose}
                    disabled={isSubmitting}
                  >
                    Skip
                  </button>
                )}
                <button
                  type="submit"
                  className="prompt-button prompt-button-primary"
                  disabled={!canSubmit}
                >
                  {isSubmitting ? (
                    <>
                      <span className="spinner" />
                      Submitting...
                    </>
                  ) : (
                    <>
                      <CheckCircle size={18} />
                      Submit
                    </>
                  )}
                </button>
              </div>
            </form>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>,
    document.body
  );
}
