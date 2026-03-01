/**
 * Chat Service
 * Handles API communication for chat functionality
 */

import type { Message } from '@/types/chat';

export interface ChatResponse {
  id: string;
  content: string;
  timestamp: Date;
}

export class ChatServiceError extends Error {
  code?: string;
  statusCode?: number;

  constructor(message: string, code?: string, statusCode?: number) {
    super(message);
    this.name = 'ChatServiceError';
    this.code = code;
    this.statusCode = statusCode;
  }
}

/**
 * Send a message to the AI and get a response
 * @param message - The user's message
 * @param chatHistory - Previous messages for context
 * @returns Promise with the AI response
 */
export async function sendMessage(
  _message: string,
  _chatHistory: Message[] = []
): Promise<ChatResponse> {
  try {
    // TODO: Replace with actual API call
    // Example: const response = await fetch('/api/chat', {
    //   method: 'POST',
    //   body: JSON.stringify({ message, history: chatHistory })
    // });

    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 500 + Math.random() * 1000));

    // Simulate occasional errors for testing
    if (Math.random() < 0.05) {
      throw new ChatServiceError('Failed to connect to AI service', 'CONNECTION_ERROR', 503);
    }

    // Simulated response
    const responses = [
      'This is a simulated response. Real AI integration coming soon!',
      'I understand your question. In a production environment, this would connect to Claude or another AI model.',
      "I'm currently in demo mode. The actual implementation will use the Anthropic Claude API.",
      "Great question! Once integrated with a real backend, I'll be able to provide much more helpful responses.",
      'This is a placeholder response. The chat system is ready for backend integration.',
    ];

    const randomResponse = responses[Math.floor(Math.random() * responses.length)];

    return {
      id: `msg-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      content: randomResponse,
      timestamp: new Date(),
    };
  } catch (error) {
    if (error instanceof ChatServiceError) {
      throw error;
    }

    // Handle unexpected errors
    throw new ChatServiceError(
      'An unexpected error occurred while sending your message',
      'UNKNOWN_ERROR'
    );
  }
}

/**
 * Generate a title for a chat based on the first message
 * @param firstMessage - The first message in the chat
 * @returns A suggested title
 */
export function generateChatTitle(firstMessage: string): string {
  try {
    // Remove extra whitespace and newlines
    const cleaned = firstMessage.trim().replace(/\s+/g, ' ');

    // Take first 50 characters or until first sentence end
    const sentenceEnd = cleaned.search(/[.!?]\s/);
    const cutoff = sentenceEnd > 0 ? Math.min(sentenceEnd, 50) : 50;

    let title = cleaned.substring(0, cutoff);

    // Add ellipsis if truncated
    if (cleaned.length > cutoff) {
      title += '...';
    }

    return title || 'New Conversation';
  } catch (error) {
    console.error('Error generating chat title:', error);
    return 'New Conversation';
  }
}

/**
 * Validate a message before sending
 * @param message - The message to validate
 * @returns Validation result
 */
export function validateMessage(message: string): {
  isValid: boolean;
  error?: string;
} {
  const trimmed = message.trim();

  if (!trimmed) {
    return { isValid: false, error: 'Message cannot be empty' };
  }

  if (trimmed.length > 50000) {
    return { isValid: false, error: 'Message is too long (max 50,000 characters)' };
  }

  return { isValid: true };
}
