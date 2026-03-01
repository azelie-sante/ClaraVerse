/**
 * Interactive Prompt Test Helper
 *
 * This utility allows you to manually trigger interactive prompts from the browser console
 * to test the feature before backend integration.
 *
 * Usage:
 * 1. Open browser console (F12)
 * 2. Type: window.testInteractivePrompt()
 * 3. Or try: window.testComplexPrompt()
 */

import { useChatStore } from '@/store/useChatStore';
import type { ActivePrompt } from '@/types/interactivePrompt';
import type { InteractiveQuestion } from '@/types/websocket';

/**
 * Simple test prompt with basic question types
 */
export function triggerSimplePrompt(): void {
  const state = useChatStore.getState();

  const questions: InteractiveQuestion[] = [
    {
      id: 'q1',
      type: 'text',
      label: 'What is your name?',
      placeholder: 'Enter your name...',
      required: true,
    },
    {
      id: 'q2',
      type: 'number',
      label: 'How old are you?',
      required: false,
      validation: {
        min: 0,
        max: 150,
      },
    },
  ];

  const prompt: ActivePrompt = {
    promptId: `test-${Date.now()}`,
    conversationId: state.conversationId || 'test-conversation',
    title: 'Simple Test Prompt',
    description: 'This is a test prompt to verify the interactive prompt feature works.',
    questions,
    allowSkip: true,
    timestamp: Date.now(),
  };

  state.setActivePrompt(prompt);
  console.log('✅ Simple prompt triggered!');
}

/**
 * Complex test prompt with all question types
 */
export function triggerComplexPrompt(): void {
  const state = useChatStore.getState();

  const questions: InteractiveQuestion[] = [
    {
      id: 'language',
      type: 'select',
      label: 'What programming language do you want to use?',
      required: true,
      options: ['Python', 'JavaScript', 'TypeScript', 'Java', 'Go'],
      allow_other: true,
    },
    {
      id: 'features',
      type: 'multi-select',
      label: 'Which features do you need?',
      required: true,
      options: ['Authentication', 'Database', 'API', 'Testing'],
      allow_other: true,
    },
    {
      id: 'complexity',
      type: 'number',
      label: 'Complexity level (1-10)',
      required: true,
      validation: {
        min: 1,
        max: 10,
      },
    },
    {
      id: 'async',
      type: 'checkbox',
      label: 'Use async/await?',
      required: false,
    },
    {
      id: 'description',
      type: 'text',
      label: 'Project description',
      placeholder: 'Describe your project...',
      required: false,
      validation: {
        min_length: 10,
        max_length: 200,
      },
    },
  ];

  const prompt: ActivePrompt = {
    promptId: `test-complex-${Date.now()}`,
    conversationId: state.conversationId || 'test-conversation',
    title: 'Create a New Project',
    description: 'To create your project, I need some more information about your requirements.',
    questions,
    allowSkip: false,
    timestamp: Date.now(),
  };

  state.setActivePrompt(prompt);
  console.log('✅ Complex prompt triggered!');
}

/**
 * Test prompt that doesn't allow skipping
 */
export function triggerRequiredPrompt(): void {
  const state = useChatStore.getState();

  const questions: InteractiveQuestion[] = [
    {
      id: 'email',
      type: 'text',
      label: 'Email address',
      placeholder: 'your@email.com',
      required: true,
      validation: {
        pattern: '^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$',
      },
    },
    {
      id: 'agree',
      type: 'checkbox',
      label: 'I agree to the terms and conditions',
      required: true,
    },
  ];

  const prompt: ActivePrompt = {
    promptId: `test-required-${Date.now()}`,
    conversationId: state.conversationId || 'test-conversation',
    title: 'Required Information',
    description: 'Please provide the following required information to continue.',
    questions,
    allowSkip: false,
    timestamp: Date.now(),
  };

  state.setActivePrompt(prompt);
  console.log('✅ Required prompt triggered (cannot skip)!');
}

/**
 * Test multiple prompts in sequence (queue test)
 */
export function triggerMultiplePrompts(): void {
  console.log('📋 Triggering 3 prompts in sequence...');

  setTimeout(() => {
    console.log('1️⃣ First prompt');
    triggerSimplePrompt();
  }, 100);

  setTimeout(() => {
    console.log('2️⃣ Second prompt (will be queued)');
    const state = useChatStore.getState();
    const prompt: ActivePrompt = {
      promptId: `test-second-${Date.now()}`,
      conversationId: state.conversationId || 'test-conversation',
      title: 'Second Prompt',
      description: 'This should appear after you answer the first one.',
      questions: [
        {
          id: 'color',
          type: 'select',
          label: 'What is your favorite color?',
          required: true,
          options: ['Red', 'Blue', 'Green', 'Yellow'],
        },
      ],
      allowSkip: true,
      timestamp: Date.now(),
    };
    state.setActivePrompt(prompt);
  }, 200);

  setTimeout(() => {
    console.log('3️⃣ Third prompt (will be queued)');
    const state = useChatStore.getState();
    const prompt: ActivePrompt = {
      promptId: `test-third-${Date.now()}`,
      conversationId: state.conversationId || 'test-conversation',
      title: 'Third Prompt',
      description: 'This should appear after you answer the second one.',
      questions: [
        {
          id: 'framework',
          type: 'select',
          label: 'Preferred framework?',
          required: true,
          options: ['React', 'Vue', 'Angular', 'Svelte'],
          allow_other: true,
        },
      ],
      allowSkip: true,
      timestamp: Date.now(),
    };
    state.setActivePrompt(prompt);
  }, 300);

  console.log('✅ All prompts queued! Answer them one by one.');
}

// Expose to window for browser console access
declare global {
  interface Window {
    testInteractivePrompt: typeof triggerSimplePrompt;
    testComplexPrompt: typeof triggerComplexPrompt;
    testRequiredPrompt: typeof triggerRequiredPrompt;
    testMultiplePrompts: typeof triggerMultiplePrompts;
  }
}

// Only expose in development
if (import.meta.env.DEV) {
  window.testInteractivePrompt = triggerSimplePrompt;
  window.testComplexPrompt = triggerComplexPrompt;
  window.testRequiredPrompt = triggerRequiredPrompt;
  window.testMultiplePrompts = triggerMultiplePrompts;

  console.log(`
🧪 Interactive Prompt Test Helpers Loaded!

Available test functions:
• window.testInteractivePrompt()    - Simple 2-question prompt
• window.testComplexPrompt()        - Complex prompt with all question types
• window.testRequiredPrompt()       - Prompt that can't be skipped
• window.testMultiplePrompts()      - Test prompt queueing (3 prompts)

Try running any of these in the console to test the feature!
  `);
}
