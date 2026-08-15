import React from 'react';
import { MessageSquare, Plus, X } from 'lucide-react';
import { Conversation } from '../services/chat';

interface SidebarProps {
  conversations: Conversation[];
  selectedId: string;
  open: boolean;
  onClose: () => void;
  onNew: () => void;
  onSelect: (id: string) => void;
}

const conversationDate = (date: string) => {
  const value = new Date(date);
  const today = new Date();
  if (value.toDateString() === today.toDateString()) return 'Today';
  return value.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
};

export const Sidebar: React.FC<SidebarProps> = ({ conversations, selectedId, open, onClose, onNew, onSelect }) => (
  <>
    {open && <button type="button" aria-label="Close conversation history" className="fixed inset-0 z-30 bg-black/30 lg:hidden" onClick={onClose} />}
    <aside className={`fixed inset-y-0 left-0 z-40 flex w-[min(19rem,85vw)] flex-col border-r border-[var(--border-hairline)] bg-[var(--bg-secondary)] p-4 shadow-xl transition-transform lg:static lg:z-auto lg:w-72 lg:shrink-0 lg:translate-x-0 lg:rounded-xl lg:border lg:shadow-sm ${open ? 'translate-x-0' : '-translate-x-full'}`} aria-label="Conversation history">
      <div className="flex items-center justify-between gap-3">
        <h2 className="font-serif text-xl font-semibold text-[var(--text-heading)]">Conversations</h2>
        <button type="button" className="rounded-md p-2 text-[var(--text-muted)] hover:bg-[var(--bg-card)] lg:hidden" onClick={onClose} aria-label="Close sidebar"><X size={19} /></button>
      </div>
      <button type="button" onClick={onNew} className="mt-4 flex w-full items-center justify-center gap-2 rounded-lg bg-[var(--btn-primary-bg)] px-4 py-3 text-sm font-semibold text-[var(--btn-primary-text)]"><Plus size={17} />New conversation</button>
      <div className="mt-5 flex-1 space-y-1 overflow-y-auto">
        {conversations.length === 0 && <p className="px-2 py-6 text-center text-sm leading-6 text-[var(--text-muted)]">Your previous conversations will appear here.</p>}
        {conversations.map((conversation) => (
          <button key={conversation.id} type="button" onClick={() => onSelect(conversation.id)} aria-current={conversation.id === selectedId ? 'true' : undefined} className={`flex w-full items-start gap-3 rounded-lg px-3 py-3 text-left transition-colors ${conversation.id === selectedId ? 'bg-[var(--bg-card)] text-[var(--text-heading)]' : 'text-[var(--text-body)] hover:bg-[var(--bg-primary)]'}`}>
            <MessageSquare className="mt-0.5 shrink-0" size={16} />
            <span className="min-w-0 flex-1"><span className="block truncate text-sm font-medium">{conversation.title}</span><span className="mt-1 block text-xs text-[var(--text-muted)]">{conversationDate(conversation.updatedAt)}</span></span>
          </button>
        ))}
      </div>
    </aside>
  </>
);
