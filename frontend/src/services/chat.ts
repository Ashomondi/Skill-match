const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';
const STORAGE_KEY = 'skillmatch-conversations';

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

const token = () => localStorage.getItem('token');

const readConversations = (): Conversation[] => {
  try {
    const stored = JSON.parse(localStorage.getItem(STORAGE_KEY) || '[]');
    return Array.isArray(stored) ? stored : [];
  } catch {
    return [];
  }
};

const writeConversations = (conversations: Conversation[]) => {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(conversations));
};

const titleFrom = (content: string) => {
  const title = content.replace(/\s+/g, ' ').trim();
  return title.length > 42 ? `${title.slice(0, 42)}…` : title || 'New conversation';
};

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

  async send(message: string, history: ChatMessage[]): Promise<string> {
    const response = await fetch(`${API_BASE_URL}/chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token() ? { Authorization: `Bearer ${token()}` } : {}),
      },
      body: JSON.stringify({
        message,
        history: history.map(({ role, content }) => ({ role, content })),
      }),
    });
    const data = await response.json().catch(() => ({}));
    if (!response.ok) throw new Error(data.error || data.message || 'The assistant could not respond.');
    return data.reply || data.response || data.message || data.content || 'I could not generate a response.';
  },
};
