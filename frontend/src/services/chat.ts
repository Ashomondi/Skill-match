import { API_BASE_URL, getErrorMessage } from './api';
const LEGACY_STORAGE_KEY = 'skillmatch-conversations';

export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  createdAt?: string;
}

export interface Conversation {
  id: string;
  title: string;
  messages: ChatMessage[];
  createdAt: string;
  updatedAt: string;
}

export interface ChatSendOptions {
  resumeId?: string;
}

export interface ChatHistoryOptions {
  limit?: number;
}

/**
 * MemoryContext describes which persistent-memory sources the assistant is
 * able to draw on for the current turn. It is intentionally coarse-grained:
 * the UI shows a label per source, never any of the underlying data.
 */
export interface MemoryContext {
  profile: boolean;
  resume: boolean;
  conversation: boolean;
}

const token = () => localStorage.getItem('token');

/** Gets the authenticated user's ID for user-scoped local caching. */
export const getCurrentUserId = (): string => {
  try {
    const userStr = localStorage.getItem('user');
    if (!userStr) return 'anonymous';
    const user = JSON.parse(userStr);
    return user.id || user.email || 'anonymous';
  } catch {
    return 'anonymous';
  }
};

/** Scopes the conversation storage key per authenticated user to prevent data mixing on shared devices. */
export const getStorageKey = (): string => {
  const userId = getCurrentUserId();
  return `skillmatch-conversations_${userId}`;
};

const readConversations = (): Conversation[] => {
  try {
    const key = getStorageKey();
    let raw = localStorage.getItem(key);
    // Backward compatibility: migrate legacy global key to current user key if not set
    if (!raw && localStorage.getItem(LEGACY_STORAGE_KEY)) {
      raw = localStorage.getItem(LEGACY_STORAGE_KEY);
      if (raw) {
        try {
          localStorage.setItem(key, raw);
        } catch {}
      }
    }
    const stored = JSON.parse(raw || '[]');
    return Array.isArray(stored) ? stored : [];
  } catch {
    return [];
  }
};

const writeConversations = (conversations: Conversation[]) => {
  try {
    localStorage.setItem(getStorageKey(), JSON.stringify(conversations));
  } catch {}
};

const titleFrom = (content: string) => {
  const title = content.replace(/\s+/g, ' ').trim();
  return title.length > 42 ? `${title.slice(0, 42)}…` : title || 'New conversation';
};

/** Records whether the user has at least one resume on file (best-effort). */
export const setResumePresent = (present: boolean) => {
  localStorage.setItem('skillmatch-resume-present', present ? '1' : '0');
};

/** Builds a memory-context snapshot from what is locally knowable. */
export const getMemoryContext = (historyLength: number): MemoryContext => ({
  profile: !!localStorage.getItem('user'),
  resume: localStorage.getItem('skillmatch-resume-present') === '1',
  conversation: historyLength > 0,
});

export const chatService = {
  list(): Conversation[] {
    return readConversations().sort((a, b) => b.updatedAt.localeCompare(a.updatedAt));
  },

  get(id: string): Conversation | undefined {
    return readConversations().find((conversation) => conversation.id === id);
  },

  create(): Conversation {
    const now = new Date().toISOString();
    return { id: crypto.randomUUID(), title: 'New conversation', messages: [], createdAt: now, updatedAt: now };
  },

  save(conversation: Conversation): Conversation {
    const conversations = readConversations();
    const next = {
      ...conversation,
      title: conversation.messages[0]?.content ? titleFrom(conversation.messages[0].content) : conversation.title,
      updatedAt: new Date().toISOString(),
    };
    const index = conversations.findIndex((item) => item.id === next.id);
    if (index >= 0) conversations[index] = next;
    else conversations.push(next);
    writeConversations(conversations);
    return next;
  },

  /**
   * Retrieves conversation history from the backend (the single source of truth),
   * updating local cache upon success and falling back to cache if offline.
   */
  async getHistory(options: ChatHistoryOptions = {}): Promise<ChatMessage[]> {
    const userToken = token();
    if (!userToken) {
      const cached = readConversations();
      return cached[0]?.messages || [];
    }

    try {
      const query = options.limit ? `?limit=${options.limit}` : '';
      const response = await fetch(`${API_BASE_URL}/chat${query}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${userToken}`,
        },
      });

      if (!response.ok) {
        // Fallback to cache on non-200 responses
        const cached = readConversations();
        return cached[0]?.messages || [];
      }

      const data = await response.json().catch(() => ({}));
      const rawMessages = Array.isArray(data)
        ? data
        : Array.isArray(data.messages)
        ? data.messages
        : Array.isArray(data.data)
        ? data.data
        : [];

      const normalized: ChatMessage[] = rawMessages.map((m: any) => ({
        id: String(m.id || crypto.randomUUID()),
        role: m.role === 'assistant' ? 'assistant' : 'user',
        content: String(m.content || ''),
        createdAt: m.createdAt || m.created_at || new Date().toISOString(),
      }));

      // Cache server history locally for fast offline access
      if (normalized.length > 0) {
        const conversations = readConversations();
        const active = conversations[0] || this.create();
        active.messages = normalized;
        active.updatedAt = new Date().toISOString();
        if (!active.title || active.title === 'New conversation') {
          active.title = active.messages[0]?.content ? titleFrom(active.messages[0].content) : active.title;
        }
        this.save(active);
      }

      return normalized;
    } catch {
      // Network failure / offline: return local cached messages
      const cached = readConversations();
      return cached[0]?.messages || [];
    }
  },

  /**
   * Sends a message to the AI assistant.
   * Notice: The redundant client-side history array is no longer transmitted
   * to POST /api/chat because the backend reconstructs memory context server-side.
   */
  async send(
    message: string,
    optionsOrHistory?: ChatSendOptions | ChatMessage[],
    maybeOptions?: ChatSendOptions,
  ): Promise<string> {
    const options: ChatSendOptions =
      optionsOrHistory && !Array.isArray(optionsOrHistory)
        ? optionsOrHistory
        : maybeOptions || {};

    const userToken = token();
    const response = await fetch(`${API_BASE_URL}/chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(userToken ? { Authorization: `Bearer ${userToken}` } : {}),
      },
      body: JSON.stringify({
        message,
        resumeId: options.resumeId,
      }),
    });

    if (!response.ok) {
      throw new Error(await getErrorMessage(response, 'The assistant could not respond.'));
    }

    const data = await response.json().catch(() => ({}));
    return data.reply || data.response || data.message || data.content || 'I could not generate a response.';
  },
};
