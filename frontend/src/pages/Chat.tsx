import React, { useCallback, useState } from 'react';
import { Menu } from 'lucide-react';
import { AppShell } from '../components/AppShell';
import { ChatBox } from '../components/ChatBox';
import { Sidebar } from '../components/Sidebar';
import { chatService, Conversation } from '../services/chat';

const initialState = () => {
  const conversations = chatService.list();
  const active = conversations[0] || chatService.create();
  return { conversations, active };
};

export const Chat: React.FC = () => {
  const [{ conversations: initialConversations, active: initialActive }] = useState(initialState);
  const [conversations, setConversations] = useState(initialConversations);
  const [active, setActive] = useState(initialActive);
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const updateConversation = useCallback((updated: Conversation) => {
    setActive(updated);
    setConversations(chatService.list());
  }, []);

  const selectConversation = (id: string) => {
    const selected = chatService.get(id);
    if (selected) setActive(selected);
    setSidebarOpen(false);
  };

  const newConversation = () => {
    setActive(chatService.create());
    setSidebarOpen(false);
  };

  return (
    <AppShell>
      <div className="mx-auto flex min-h-[calc(100vh-190px)] w-full max-w-6xl gap-5">
        <Sidebar conversations={conversations} selectedId={active.id} open={sidebarOpen} onClose={() => setSidebarOpen(false)} onNew={newConversation} onSelect={selectConversation} />
        <div className="flex min-w-0 flex-1 flex-col">
          <div className="mb-6 flex items-start gap-3">
            <button type="button" onClick={() => setSidebarOpen(true)} aria-label="Open conversation history" title="Conversation history" className="mt-1 grid h-10 w-10 shrink-0 place-items-center rounded-md border border-[var(--border-hairline)] bg-[var(--bg-secondary)] text-[var(--text-heading)] lg:hidden"><Menu size={20} /></button>
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[var(--accent-gold)]">SkillMatch assistant</p>
              <h1 className="mt-2 font-serif text-3xl font-bold text-[var(--text-heading)] sm:text-4xl">Your career, with a memory.</h1>
              <p className="mt-2 max-w-2xl text-sm leading-6 text-[var(--text-muted)]">Ask about roles, tailor your experience, or plan your next application.</p>
            </div>
          </div>
          <ChatBox conversation={active} onChange={updateConversation} />
        </div>
      </div>
    </AppShell>
  );
};
