import { useCallback, useState } from 'react';
import { chatService, ChatMessage } from '../services/chat';

export function useChat() {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const sendMessage = useCallback(async (content: string) => {
    setError(null); setLoading(true);
    const userMessage: ChatMessage = { id: `${Date.now()}-user`, role: 'user', content };
    setMessages((current) => [...current, userMessage]);
    try { const reply = await chatService.send(content, [...messages, userMessage]); setMessages((current) => [...current, { id: `${Date.now()}-assistant`, role: 'assistant', content: reply }]); }
    catch (err) { setError(err instanceof Error ? err.message : 'Unable to reach the assistant.'); }
    finally { setLoading(false); }
  }, [messages]);
  return { messages, loading, error, sendMessage };
}
