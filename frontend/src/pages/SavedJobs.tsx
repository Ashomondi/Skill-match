import React, { useCallback, useEffect, useState } from 'react';
import { AlertCircle, Bookmark, Loader2, RefreshCw } from 'lucide-react';
import { AppShell } from '../components/AppShell';
import { SavedJobs as SavedJobsList } from '../components/SavedJobs';
import { SavedJobsEmpty } from '../components/saved-jobs/SavedJobsEmpty';
import { SavedJob, savedJobsService } from '../services/savedJobs';

export const SavedJobs: React.FC = () => {
  const [jobs, setJobs] = useState<SavedJob[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [removingId, setRemovingId] = useState<string | null>(null);

  const loadJobs = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setJobs(await savedJobsService.list());
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : 'Saved jobs could not be loaded.');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { void loadJobs(); }, [loadJobs]);

  const removeJob = async (savedJob: SavedJob) => {
    if (!window.confirm(`Remove ${savedJob.title} from saved jobs?`)) return;
    setRemovingId(savedJob.id);
    setError(null);
    try {
      await savedJobsService.remove(savedJob.jobId);
      setJobs((current) => current.filter((job) => job.id !== savedJob.id));
    } catch (removeError) {
      setError(removeError instanceof Error ? removeError.message : 'The saved job could not be removed.');
    } finally {
      setRemovingId(null);
    }
  };

  return (
    <AppShell>
      <div className="mx-auto w-full max-w-6xl">
        <header className="mb-8 flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[var(--accent-gold)]">Your shortlist</p>
            <h1 className="mt-2 font-serif text-4xl font-bold text-[var(--text-heading)]">Saved jobs</h1>
            <p className="mt-2 max-w-xl text-sm leading-6 text-[var(--text-muted)]">Keep promising roles together and return when you are ready to apply.</p>
          </div>
          {!loading && !error && jobs.length > 0 && <span className="inline-flex w-fit items-center gap-2 text-sm text-[var(--text-muted)]"><Bookmark size={16} />{jobs.length} saved {jobs.length === 1 ? 'role' : 'roles'}</span>}
        </header>

        {error && <div role="alert" className="mb-5 flex flex-col gap-3 rounded-md border border-[var(--status-rejected)] bg-[var(--status-rejected)]/10 p-4 text-sm text-[var(--text-heading)] sm:flex-row sm:items-center sm:justify-between"><span className="flex items-center gap-2"><AlertCircle className="shrink-0 text-[var(--status-rejected)]" size={18} />{error}</span><button type="button" onClick={() => void loadJobs()} className="inline-flex items-center justify-center gap-2 rounded-md border border-[var(--text-button-fill)] px-3 py-2 font-semibold text-[var(--text-button-fill)]"><RefreshCw size={15} />Try again</button></div>}
        {loading ? <div className="flex min-h-64 items-center justify-center gap-3 text-sm text-[var(--text-muted)]"><Loader2 className="animate-spin" size={20} />Loading saved jobs...</div> : jobs.length === 0 && !error ? <SavedJobsEmpty /> : <SavedJobsList jobs={jobs} removingId={removingId} onRemove={removeJob} />}
      </div>
    </AppShell>
  );
};
