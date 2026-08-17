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

  useEffect(() => {
    setMessages(conversation.messages);
    setError(null);
    setContextSources(sourceList(conversation.messages.length));
  }, [conversation.id, conversation.messages]);

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
    const history = [...messages, userMessage];
    persist(history);
    try {
      const reply = await chatService.send(content, history);
      persist([...history, { id: `${Date.now()}-assistant`, role: 'assistant', content: reply, createdAt: new Date().toISOString() }]);
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
    contextSources,
    contextLabel: contextSources.map((source) => SOURCE_LABELS[source]),
  };
}
