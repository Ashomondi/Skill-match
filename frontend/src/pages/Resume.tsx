import React, { useCallback, useEffect, useState } from 'react';
import { AlertCircle, CheckCircle2 } from 'lucide-react';
import { AppShell } from '../components/AppShell';
import { Resume, ResumeList } from '../components/ResumeList';
import { ResumeUploader } from '../components/ResumeUploader';
import { resumeService } from '../services/resume';
import { useUpload } from '../hooks/useUpload';

export const ResumePage: React.FC = () => {
  const { upload, isUploading } = useUpload();
  const [resumes, setResumes] = useState<Resume[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [replaceTarget, setReplaceTarget] = useState<Resume | null>(null);

  const loadResumes = useCallback(async () => {
    setIsLoading(true);
    try {
      setResumes(await resumeService.list());
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : 'Unable to load your resumes.');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => { void loadResumes(); }, [loadResumes]);

  const handleUpload = useCallback(async (file: File) => {
    setError(null);
    setMessage(null);
    const created = await upload(file, replaceTarget?.id);
    if (!created) return;

    setResumes((current) => (
      replaceTarget
        ? [created, ...current.filter((resume) => resume.id !== replaceTarget.id)]
        : [created, ...current]
    ));
    setMessage(replaceTarget ? 'Your resume was replaced successfully.' : 'Your resume was uploaded successfully.');
    setReplaceTarget(null);
  }, [upload, replaceTarget]);

  const handleDelete = useCallback(async (resume: Resume) => {
    if (!window.confirm(`Delete ${resume.name}? This cannot be undone.`)) return;
    setDeletingId(resume.id);
    setError(null);
    try {
      await resumeService.remove(resume.id);
      setResumes((current) => current.filter((item) => item.id !== resume.id));
      setMessage('Resume deleted.');
    } catch (deleteError) {
      setError(deleteError instanceof Error ? deleteError.message : 'Your resume could not be deleted.');
    } finally {
      setDeletingId(null);
    }
  }, []);

  const handleReplace = useCallback((resume: Resume) => {
    setReplaceTarget(resume);
    setMessage(`Choose a new file to replace ${resume.name}.`);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }, []);

  return (
    <AppShell>
      <div className="mx-auto w-full max-w-5xl">
        <header className="mb-8 max-w-2xl">
          <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[var(--accent-gold)]">CV tailor</p>
          <h1 className="mt-2 font-serif text-[38px] font-semibold tracking-[-0.03em] text-[var(--text-heading)]">Your master profile starts here.</h1>
          <p className="mt-3 text-[15px] leading-6 text-[var(--text-muted)]">Upload the resume that best represents your experience. You can replace it whenever your story evolves.</p>
        </header>

        {message ? <div className="mb-5 flex items-center gap-2 rounded border border-[var(--status-offer)] bg-[var(--status-offer)]/10 p-3 text-sm text-[var(--text-heading)]"><CheckCircle2 className="h-4 w-4 text-[var(--status-offer)]" />{message}</div> : null}
        {error ? <div role="alert" className="mb-5 flex items-center gap-2 rounded border border-[var(--status-rejected)] bg-[var(--status-rejected)]/10 p-3 text-sm text-[var(--text-heading)]"><AlertCircle className="h-4 w-4 text-[var(--status-rejected)]" />{error}</div> : null}

        <ResumeUploader onUpload={handleUpload} isUploading={isUploading} />

        {isLoading ? <div className="mt-8 rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-secondary)] p-6 text-sm text-[var(--text-muted)]">Loading your resumes…</div> : <ResumeList resumes={resumes} isDeletingId={deletingId} onDelete={handleDelete} onReplace={handleReplace} />}
      </div>
    </AppShell>
  );
};

export default ResumePage;
