import React from 'react';
import { ArrowUpRight, Building2, MapPin } from 'lucide-react';
import { Link } from 'react-router-dom';
import { Job } from '../services/jobs';
import { MatchRing } from './MatchRing';

const postedLabel = (date?: string) => {
  if (!date) return '';
  const days = Math.max(0, Math.floor((Date.now() - new Date(date).getTime()) / 86400000));
  if (days === 0) return 'Posted today';
  return `Posted ${days}d ago`;
};

export const JobCard: React.FC<{ job: Job }> = ({ job }) => (
  <article className="flex flex-col gap-5 rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-secondary)] p-5 transition-shadow hover:shadow-sm sm:flex-row sm:items-center">
    <div className="grid h-11 w-11 shrink-0 place-items-center rounded-full bg-[var(--bg-card)] font-serif text-lg font-bold text-[var(--text-heading)]">{job.company.charAt(0).toUpperCase()}</div>
    <div className="min-w-0 flex-1">
      <h2 className="text-lg font-semibold text-[var(--text-heading)]">{job.title}</h2>
      <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-sm text-[var(--text-muted)]">
        <span className="inline-flex items-center gap-1.5"><Building2 size={14} />{job.company}</span>
        <span className="inline-flex items-center gap-1.5"><MapPin size={14} />{job.location}</span>
      </div>
      <div className="mt-3 flex flex-wrap items-center gap-2 text-xs text-[var(--text-insight)]">
        {job.workType && <span className="rounded-full bg-[var(--bg-chip)] px-3 py-1">{job.workType}</span>}
        {job.seniority && <span className="rounded-full bg-[var(--bg-chip)] px-3 py-1">{job.seniority}</span>}
        {job.salary && <span className="rounded-full bg-[var(--bg-chip)] px-3 py-1">{job.salary}</span>}
        {job.postedAt && <span>{postedLabel(job.postedAt)}</span>}
      </div>
    </div>
    <div className="flex items-center justify-between gap-4 sm:justify-end">
      {job.matchScore !== undefined && <MatchRing value={job.matchScore} />}
      <Link to={`/discover/${job.id}`} aria-label={`Open ${job.title} details`} className="grid h-10 w-10 place-items-center rounded-md bg-[var(--btn-primary-bg)] text-[var(--btn-primary-text)]"><ArrowUpRight size={18} /></Link>
    </div>
  </article>
);
