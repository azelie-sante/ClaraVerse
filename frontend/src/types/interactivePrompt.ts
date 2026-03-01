import type { InteractiveQuestion } from './websocket';

// Frontend state for active prompt
export interface ActivePrompt {
  promptId: string;
  conversationId: string;
  title: string;
  description?: string;
  questions: InteractiveQuestion[];
  allowSkip: boolean;
  timestamp: number;
}

// Frontend form state
export interface PromptFormState {
  answers: Record<string, PromptAnswer>;
  errors: Record<string, string>;
  isValid: boolean;
}

export interface PromptAnswer {
  questionId: string;
  value: string | number | boolean | string[];
  isOther?: boolean;
  otherText?: string; // For "Other" option with custom input
}

// Validation result
export interface ValidationResult {
  isValid: boolean;
  errors: Record<string, string>;
}
