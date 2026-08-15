import { useCallback, useEffect, useState } from 'react';
import { chatService, ChatMessage, Conversation } from '../services/chat';

export function useChat(conversation: Conversation, onChange: (conversation: Conversation) => void) {
  const [messages, setMessages] = useState<ChatMessage[]>(conversation.messages);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setMessages(conversation.messages);
    setError(null);
  }, [conversation.id, conversation.messages]);

  const persist = useCallback((nextMessages: ChatMessage[]) => {
    const saved = chatService.save({ ...conversation, messages: nextMessages });
    setMessages(saved.messages);
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

  return { messages, loading, error, sendMessage };
}
