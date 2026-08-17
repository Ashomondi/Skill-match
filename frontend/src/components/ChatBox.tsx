import React, { useEffect, useRef, useState } from 'react';
import { ArrowUp, Brain, Loader2, Sparkles } from 'lucide-react';
import { useChat, MemorySource } from '../hooks/useChat';
import { Conversation } from '../services/chat';

interface ChatBoxProps {
  conversation: Conversation;
  onChange: (conversation: Conversation) => void;
}

const sourceStyles: Record<MemorySource, string> = {
  profile: 'bg-[var(--bg-chip)] text-[var(--text-insight)]',
  resume: 'bg-[var(--bg-chip)] text-[var(--text-insight)]',
  conversation: 'bg-[var(--bg-chip)] text-[var(--text-insight)]',
};

export const ChatBox: React.FC<ChatBoxProps> = ({ conversation, onChange }) => {
  const { messages, loading, error, sendMessage, contextSources, contextLabel } = useChat(conversation, onChange);
  const [draft, setDraft] = useState('');
  const endRef = useRef<HTMLDivElement>(null);

  useEffect(() => { endRef.current?.scrollIntoView({ behavior: 'smooth' }); }, [messages, loading]);

  const submit = async (event: React.FormEvent) => {
    event.preventDefault();
    const content = draft.trim();
    if (!content || loading) return;
    setDraft('');
    await sendMessage(content);
  };

  return <section className="flex min-h-[520px] flex-1 flex-col overflow-hidden rounded-xl border border-[var(--border-hairline)] bg-[var(--bg-secondary)] shadow-sm">
    <div className="border-b border-[var(--border-hairline)] px-5 py-4">
      <div className="flex items-center gap-3">
        <span className="grid h-9 w-9 place-items-center rounded-full bg-[var(--bg-card)] text-[var(--text-heading)]"><Sparkles size={17} /></span>
        <div className="min-w-0"><h2 className="truncate font-semibold text-[var(--text-heading)]">{conversation.title}</h2><p className="text-xs text-[var(--text-muted)]">Career assistant</p></div>
        {loading && <span className="ml-auto inline-flex items-center gap-1.5 rounded-full bg-[var(--bg-chip)] px-3 py-1 text-xs text-[var(--text-muted)]"><Loader2 className="animate-spin" size={12} />Thinking…</span>}
      </div>
      {contextSources.length > 0 && <div className="mt-3 flex flex-wrap items-center gap-2" aria-label="Memory context being used">
        <span className="inline-flex items-center gap-1.5 text-xs font-semibold text-[var(--accent-gold)]"><Brain size={13} />Drawing on</span>
        {contextSources.map((source, index) => <span key={source} className={`rounded-full px-2.5 py-1 text-xs ${sourceStyles[source]}`}>{contextLabel[index]}</span>)}
      </div>}
    </div>
    <div className="flex-1 space-y-4 overflow-y-auto p-4 sm:p-6" aria-live="polite">
      {messages.length === 0 && <div className="mx-auto mt-12 max-w-sm text-center"><Sparkles className="mx-auto text-[var(--accent-gold)]" size={24} /><p className="mt-3 font-serif text-xl font-semibold text-[var(--text-heading)]">What would you like to work on?</p><p className="mt-2 text-sm text-[var(--text-muted)]">Try asking for role recommendations or feedback on your CV.</p></div>}
      {messages.map((message) => <div key={message.id} className={`flex ${message.role === 'user' ? 'justify-end' : 'justify-start'}`}><div className={`max-w-[85%] whitespace-pre-wrap rounded-2xl px-4 py-3 text-sm leading-6 sm:max-w-[72%] ${message.role === 'user' ? 'rounded-br-sm bg-[var(--btn-primary-bg)] text-[var(--btn-primary-text)]' : 'rounded-bl-sm border border-[var(--border-hairline)] bg-[var(--bg-primary)] text-[var(--text-insight)]'}`}>{message.content}</div></div>)}
      {loading && <div className="flex justify-start"><div className="flex items-center gap-2 rounded-2xl rounded-bl-sm border border-[var(--border-hairline)] bg-[var(--bg-primary)] px-4 py-3 text-sm text-[var(--text-muted)]"><Loader2 className="animate-spin" size={15} />Thinking...</div></div>}
      {error && <p className="text-center text-sm text-[var(--status-rejected)]">{error}</p>}
      <div ref={endRef} />
    </div>
    <form onSubmit={submit} className="border-t border-[var(--border-hairline)] p-3 sm:p-4"><div className="flex items-end gap-2 rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-input)] p-2"><textarea value={draft} onChange={(event) => setDraft(event.target.value)} onKeyDown={(event) => { if (event.key === 'Enter' && !event.shiftKey) { event.preventDefault(); void submit(event); } }} rows={1} placeholder="Message SkillMatch..." aria-label="Message SkillMatch" className="max-h-32 min-h-10 flex-1 resize-none bg-transparent px-2 py-2 text-sm text-[var(--text-heading)] outline-none" disabled={loading} /><button type="submit" aria-label="Send message" title="Send message" disabled={!draft.trim() || loading} className="grid h-10 w-10 shrink-0 place-items-center rounded-md bg-[var(--btn-primary-bg)] text-[var(--btn-primary-text)] transition-opacity disabled:cursor-not-allowed disabled:opacity-40"><ArrowUp size={18} /></button></div></form>
  </section>;
};
