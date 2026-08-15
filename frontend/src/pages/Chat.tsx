import React from 'react';
import { AppShell } from '../components/AppShell';
import { ChatBox } from '../components/ChatBox';

export const Chat: React.FC = () => (
  <AppShell>
    <div className="mx-auto flex min-h-[calc(100vh-190px)] w-full max-w-4xl flex-col">
      <div className="mb-6">
        <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[var(--accent-gold)]">SkillMatch assistant</p>
        <h1 className="mt-2 font-serif text-4xl font-bold text-[var(--text-heading)]">Your career, with a memory.</h1>
        <p className="mt-2 max-w-2xl text-sm leading-6 text-[var(--text-muted)]">Ask about roles, tailor your experience, or plan your next application.</p>
      </div>
      <ChatBox />
    </div>
  </AppShell>
);
