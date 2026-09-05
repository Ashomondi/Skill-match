import { useCallback, useEffect, useState } from 'react';
import { chatService, ChatMessage, getMemoryContext, Conversation } from '../services/chat';

export type MemorySource = 'profile' | 'resume' | 'conversation';

const SOURCE_LABELS: Record<MemorySource, string> = {
  profile: 'Your profile',
  resume: 'Your resume',
  conversation: 'Past conversations',
};

const sourceList = (historyLength: number): MemorySource[] => {
  const context = getMemoryContext(historyLength);
  const sources: MemorySource[] = [];
  if (context.profile) sources.push('profile');
  if (context.resume) sources.push('resume');
  if (context.conversation) sources.push('conversation');
  return sources;
};

export function useChat(conversation: Conversation, onChange: (conversation: Conversation) => void) {
  const [messages, setMessages] = useState<ChatMessage[]>(conversation.messages);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [contextSources, setContextSources] = useState<MemorySource[]>(() => sourceList(conversation.messages.length));

  // Sync state with incoming conversation prop (e.g. switching conversations)
  useEffect(() => {
    setMessages(conversation.messages);
    setError(null);
    setContextSources(sourceList(conversation.messages.length));
  }, [conversation.id, conversation.messages]);

  // Load latest persistent history from the backend on mount/switch
  const loadServerHistory = useCallback(async () => {
    try {
      const serverMessages = await chatService.getHistory();
      if (serverMessages.length > 0) {
        setMessages(serverMessages);
        setContextSources(sourceList(serverMessages.length));
        onChange({ ...conversation, messages: serverMessages });
      }
    } catch {
      // Offline or network error: stay on cached messages
    }
  }, [conversation, onChange]);

  useEffect(() => {
    let mounted = true;
    chatService.getHistory().then((serverMessages) => {
      if (mounted && serverMessages.length > 0) {
        setMessages(serverMessages);
        setContextSources(sourceList(serverMessages.length));
        onChange({ ...conversation, messages: serverMessages });
      }
    }).catch(() => {});

    return () => {
      mounted = false;
    };
  }, [conversation.id]);

  const persist = useCallback((nextMessages: ChatMessage[]) => {
    const saved = chatService.save({ ...conversation, messages: nextMessages });
    setMessages(saved.messages);
    setContextSources(sourceList(saved.messages.length));
    onChange(saved);
  }, [conversation, onChange]);

  const sendMessage = useCallback(async (content: string) => {
    setError(null);
    setLoading(true);
    const now = new Date().toISOString();
    const userMessage: ChatMessage = { id: `${Date.now()}-user`, role: 'user', content, createdAt: now };
    const currentList = [...messages, userMessage];
    persist(currentList);

    try {
      // Redundant client-side history array is no longer passed;
      // server manages context from database persistence.
      const reply = await chatService.send(content);
      const assistantMessage: ChatMessage = {
        id: `${Date.now()}-assistant`,
        role: 'assistant',
        content: reply,
        createdAt: new Date().toISOString(),
      };
      persist([...currentList, assistantMessage]);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to reach the assistant.');
    } finally {
      setLoading(false);
    }
  }, [messages, persist]);

  return {
    messages,
    loading,
    error,
    sendMessage,
    loadServerHistory,
    contextSources,
    contextLabel: contextSources.map((source) => SOURCE_LABELS[source]),
  };
}
