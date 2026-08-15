import React from 'react';
import { ArrowUpRight, BookmarkX, Building2, Loader2, MapPin } from 'lucide-react';
import { Link } from 'react-router-dom';
import { SavedJob } from '../../services/savedJobs';

interface SavedJobCardProps {
  savedJob: SavedJob;
  removing: boolean;
  onRemove: (savedJob: SavedJob) => void;
}

export const SavedJobCard: React.FC<SavedJobCardProps> = ({ savedJob, removing, onRemove }) => (
  <article className="flex h-full flex-col border-b border-[var(--border-hairline)] bg-[var(--bg-secondary)] p-5 last:border-b-0 sm:rounded-lg sm:border">
    <div className="flex items-start justify-between gap-4">
      <div className="min-w-0">
        <h2 className="text-lg font-semibold text-[var(--text-heading)]">{savedJob.title}</h2>
        <p className="mt-2 flex items-center gap-2 text-sm text-[var(--text-body)]"><Building2 size={15} />{savedJob.company}</p>
        <p className="mt-1 flex items-center gap-2 text-sm text-[var(--text-muted)]"><MapPin size={15} />{savedJob.location}</p>
      </div>
      {savedJob.matchScore !== undefined && <span className="shrink-0 rounded-full bg-[var(--bg-chip)] px-3 py-1 text-xs font-semibold text-[var(--text-heading)]">{savedJob.matchScore}% match</span>}
    </div>
    <div className="mt-4 flex flex-wrap gap-2 text-xs text-[var(--text-insight)]">
      {savedJob.workType && <span className="rounded-full bg-[var(--bg-chip)] px-3 py-1.5">{savedJob.workType}</span>}
      {savedJob.salary && <span className="rounded-full bg-[var(--bg-chip)] px-3 py-1.5">{savedJob.salary}</span>}
    </div>
    <div className="mt-auto flex flex-col gap-2 pt-6 sm:flex-row">
      <Link to={`/discover/${savedJob.jobId}`} className="inline-flex min-h-10 flex-1 items-center justify-center gap-2 rounded-md bg-[var(--btn-primary-bg)] px-4 py-2 text-sm font-semibold text-[var(--btn-primary-text)]">Open details <ArrowUpRight size={16} /></Link>
      <button type="button" disabled={removing} onClick={() => onRemove(savedJob)} className="inline-flex min-h-10 items-center justify-center gap-2 rounded-md border border-[var(--border-hairline)] px-4 py-2 text-sm font-semibold text-[var(--text-button-fill)] disabled:cursor-wait disabled:opacity-60">
        {removing ? <Loader2 className="animate-spin" size={16} /> : <BookmarkX size={16} />}Remove
      </button>
    </div>
  </article>
);
