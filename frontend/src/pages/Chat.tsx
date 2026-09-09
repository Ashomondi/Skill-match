import React, { useCallback, useState } from 'react';
import { FileText, Menu } from 'lucide-react';
import { AppShell } from '../components/AppShell';
import { ChatBox } from '../components/ChatBox';
import { CvPicker } from '../components/CvPicker';
import { Sidebar } from '../components/Sidebar';
import { chatService, Conversation } from '../services/chat';
import { useResumeContext } from '../hooks/useResumeContext';
import { isResumeReady } from '../services/resume';

const INCLUDE_CV_KEY = 'skillmatch-include-cv';
const initialIncludeCv = () => { try { return localStorage.getItem(INCLUDE_CV_KEY) !== '0'; } catch { return true; } };

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
  const [includeCv, setIncludeCv] = useState(initialIncludeCv);
  const rc = useResumeContext();

  const usableCv = rc.active && isResumeReady(rc.active) ? rc.active : null;
  const resumeId = includeCv && usableCv ? usableCv.id : undefined;

  const toggleIncludeCv = () => {
    setIncludeCv((value) => {
      const next = !value;
      try { localStorage.setItem(INCLUDE_CV_KEY, next ? '1' : '0'); } catch { /* ignore */ }
      return next;
    });
  };

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

          <div className="mb-4 space-y-3">
            <CvPicker
              resumes={rc.resumes}
              active={rc.active}
              loading={rc.loading}
              uploading={rc.isUploading}
              error={rc.error}
              onUpload={rc.upload}
              onSelect={rc.select}
              title="Assistant CV context"
              description="Upload your CV and the assistant can reference it when you ask for tailored advice, rewrites, or feedback."
            />
            <label className="flex cursor-pointer items-center justify-between gap-3 rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-secondary)] px-4 py-3 shadow-sm">
              <span className="flex items-center gap-2 text-sm font-medium text-[var(--text-heading)]">
                <FileText size={16} className="text-[var(--text-muted)]" />
                Include my CV as context for messages
              </span>
              <button
                type="button"
                role="switch"
                aria-checked={includeCv}
                aria-label="Include my CV as context"
                disabled={!usableCv}
                onClick={toggleIncludeCv}
                className={`relative h-6 w-11 shrink-0 rounded-full transition-colors disabled:cursor-not-allowed disabled:opacity-40 ${includeCv && usableCv ? 'bg-[var(--btn-primary-bg)]' : 'bg-[var(--border-hairline)]'}`}
              >
                <span className={`absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition-transform ${includeCv && usableCv ? 'translate-x-[1.4rem]' : 'translate-x-0.5'}`} />
              </button>
            </label>
            {!rc.loading && !usableCv ? (
              <p className="text-xs leading-5 text-[var(--text-muted)]">
                {rc.active?.status === 'failed'
                  ? `"${rc.active.name}" couldn't be parsed — replace it or upload a PDF/DOCX/TXT to enable CV context.`
                  : 'Upload a CV above to let the assistant answer questions about your experience.'}
              </p>
            ) : null}
          </div>

          <ChatBox conversation={active} onChange={updateConversation} resumeId={resumeId} resumeName={usableCv?.name ?? null} />
        </div>
      </div>
    </AppShell>
  );
};
