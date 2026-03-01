/**
 * Chat Usage Examples
 * Demonstrates how to use the refactored chat system
 */

/* eslint-disable react-refresh/only-export-components */

import { useChatStore } from '@/store/useChatStore';
import type { Message, Chat } from '@/types/chat';
import { sendMessage, validateMessage } from '@/services/chatService';
import { getChatStats } from '@/utils/chatUtils';

// ============================================
// Example 1: Basic Usage in a Component
// ============================================

export function ChatExample1() {
  const { selectedChat, createChat, addMessage } = useChatStore();

  const handleNewMessage = async (content: string) => {
    // Validate
    const validation = validateMessage(content);
    if (!validation.isValid) {
      alert(validation.error);
      return;
    }

    // Create message
    const userMessage: Message = {
      id: `msg-${Date.now()}`,
      role: 'user',
      content,
      timestamp: new Date(),
    };

    // Create or update chat
    const chat = selectedChat();
    const chatId = chat?.id || createChat('New Chat', userMessage);

    // Send to AI
    const response = await sendMessage(content);
    addMessage(chatId, {
      id: response.id,
      role: 'assistant',
      content: response.content,
      timestamp: response.timestamp,
    });
  };

  // Example usage: return a button that sends a message
  return <button onClick={() => handleNewMessage('Hello')}>Send Message</button>;
}

// ============================================
// Example 2: Accessing Chat State
// ============================================

export function ChatStatsComponent() {
  const { selectedChat } = useChatStore();
  const chat = selectedChat();

  if (!chat) return <div>No chat selected</div>;

  const stats = getChatStats(chat.messages);

  return (
    <div>
      <h3>{chat.title}</h3>
      <p>Total messages: {stats.totalMessages}</p>
      <p>User messages: {stats.userMessages}</p>
      <p>Assistant messages: {stats.assistantMessages}</p>
      <p>Total words: {stats.totalWords}</p>
    </div>
  );
}

// ============================================
// Example 3: Filtered Chats List
// ============================================

export function ChatListComponent() {
  const { filteredChats, searchQuery, setSearchQuery, selectChat } = useChatStore();
  const chats = filteredChats();

  return (
    <div>
      <input
        type="search"
        value={searchQuery}
        onChange={e => setSearchQuery(e.target.value)}
        placeholder="Search chats..."
      />
      {chats.map(chat => (
        <button key={chat.id} onClick={() => selectChat(chat.id)}>
          {chat.title}
        </button>
      ))}
    </div>
  );
}

// ============================================
// Example 4: Error Handling
// ============================================

export function ChatWithErrorHandling() {
  const { error, clearError, setError, setLoading } = useChatStore();

  const handleAction = async () => {
    try {
      setLoading(true);
      // Your async operation
      await someAsyncOperation();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      {error && (
        <div className="error-banner">
          {error}
          <button onClick={clearError}>Dismiss</button>
        </div>
      )}
      <button onClick={handleAction}>Do Something</button>
    </div>
  );
}

// ============================================
// Example 5: Outside React Components
// ============================================

// You can use the store outside of React components
export function exportChats() {
  const { chats } = useChatStore.getState();
  const data = JSON.stringify(chats, null, 2);
  // Download or send data
  return data;
}

export function importChats(data: Chat[]) {
  const store = useChatStore.getState();
  data.forEach(chat => {
    store.createChat(chat.title, chat.messages[0]);
    chat.messages.slice(1).forEach((msg: Message) => {
      store.addMessage(chat.id, msg);
    });
  });
}

// ============================================
// Example 6: Subscribe to Store Changes
// ============================================

// Listen to specific state changes
export const unsubscribeFromStore = useChatStore.subscribe(state => {
  const selectedChat = state.selectedChat();
  console.log('Selected chat changed:', selectedChat);
});

// Clean up when done
// unsubscribeFromStore();

// ============================================
// Example 7: Testing
// ============================================

// Mock the store for tests
// Note: Uncomment when vitest is installed and configured
/*
import { beforeEach, test, expect } from 'vitest';

beforeEach(() => {
  // Reset store state before each test
  useChatStore.setState({
    chats: [],
    selectedChatId: null,
    isNewChat: false,
    searchQuery: '',
    activeNav: 'chats',
  });
});

test('creates a new chat', () => {
  const { createChat } = useChatStore.getState();

  const message: Message = {
    id: '1',
    role: 'user',
    content: 'Hello',
    timestamp: new Date(),
  };

  const chatId = createChat('Test Chat', message);

  const { chats } = useChatStore.getState();
  expect(chats).toHaveLength(1);
  expect(chats[0].id).toBe(chatId);
  expect(chats[0].title).toBe('Test Chat');
});
*/

// ============================================
// Example 8: Custom Hooks
// ============================================

// Create custom hooks for common patterns
export function useCurrentChat() {
  const selectedChat = useChatStore(state => state.selectedChat());
  return selectedChat;
}

export function useChatActions() {
  return {
    createChat: useChatStore(state => state.createChat),
    addMessage: useChatStore(state => state.addMessage),
    updateChatTitle: useChatStore(state => state.updateChatTitle),
    deleteChat: useChatStore(state => state.deleteChat),
  };
}

// Use in component
export function MyComponent() {
  const chat = useCurrentChat();
  const { createChat, addMessage } = useChatActions();

  // Your component logic
  return (
    <div>
      <p>Current chat: {chat?.title || 'None'}</p>
      <button
        onClick={() =>
          createChat('New', { id: '1', role: 'user', content: 'Hi', timestamp: new Date() })
        }
      >
        Create
      </button>
      <button
        onClick={() =>
          chat &&
          addMessage(chat.id, {
            id: '2',
            role: 'assistant',
            content: 'Hello',
            timestamp: new Date(),
          })
        }
      >
        Add Message
      </button>
    </div>
  );
}

// ============================================
// Example 9: Integrating with Real API
// ============================================

// Replace the mock in chatService.ts with real API
/*
export async function sendMessage(
  message: string,
  chatHistory: Message[] = []
): Promise<ChatResponse> {
  try {
    const response = await fetch('https://api.anthropic.com/v1/messages', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-api-key': process.env.ANTHROPIC_API_KEY,
        'anthropic-version': '2023-06-01',
      },
      body: JSON.stringify({
        model: 'claude-3-sonnet-20240229',
        messages: [
          ...chatHistory.map((msg) => ({
            role: msg.role,
            content: msg.content,
          })),
          { role: 'user', content: message },
        ],
        max_tokens: 1024,
      }),
    });

    if (!response.ok) {
      throw new ChatServiceError(
        'Failed to get response from AI',
        'API_ERROR',
        response.status
      );
    }

    const data = await response.json();
    
    return {
      id: data.id,
      content: data.content[0].text,
      timestamp: new Date(),
    };
  } catch (error) {
    if (error instanceof ChatServiceError) {
      throw error;
    }
    throw new ChatServiceError('Network error occurred', 'NETWORK_ERROR');
  }
}
*/

// ============================================
// Example 10: Persisting to Backend
// ============================================

// Save chats to your backend
export async function syncChatsToBackend() {
  const { chats } = useChatStore.getState();

  try {
    await fetch('/api/chats/sync', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ chats }),
    });
  } catch (error) {
    console.error('Failed to sync chats:', error);
  }
}

// Load chats from backend
export async function loadChatsFromBackend() {
  try {
    const response = await fetch('/api/chats');
    const chats: Chat[] = await response.json();

    // Update store
    useChatStore.setState({ chats });
  } catch (error) {
    console.error('Failed to load chats:', error);
  }
}

// Placeholder for someAsyncOperation
async function someAsyncOperation() {
  return Promise.resolve();
}
