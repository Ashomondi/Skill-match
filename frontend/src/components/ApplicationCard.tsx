import React from 'react';
import { ExternalLink, FileText } from 'lucide-react';
import { Link } from 'react-router-dom';
import { Application } from '../services/application';

export const ApplicationCard = React.memo(({ application, onStatusChange, updating }: { application: Application; onStatusChange: (status: Application["status"]) => void; updating: boolean }) => (
  <article className="flex flex-col gap-4 border-b border-[var(--border-hairline)] bg-[var(--bg-secondary)] p-5 last:border-b-0 sm:rounded-lg sm:border">
    <div className="flex items-start gap-3">
      <span className="grid h-10 w-10 shrink-0 place-items-center rounded-full bg-[var(--bg-card)] font-semibold text-[var(--text-heading)]">{application.company.charAt(0).toUpperCase()}</span>
      <div className="min-w-0 flex-1"><h2 className="font-semibold text-[var(--text-heading)]">{application.role}</h2><p className="mt-1 truncate text-sm text-[var(--text-muted)]">{application.company}</p></div>
      <select aria-label="Update application status" value={application.status} disabled={updating} onChange={(event) => onStatusChange(event.target.value as Application["status"])} className={`rounded-full px-3 py-1 text-xs font-semibold capitalize `}><option value="applied">applied</option><option value="screening">screening</option><option value="interview">interview</option><option value="offer">offer</option><option value="rejected">rejected</option><option value="withdrawn">withdrawn</option></select>
    </div>
    <div className="flex flex-wrap items-center justify-between gap-3 border-t border-[var(--border-hairline)] pt-4 text-xs text-[var(--text-muted)]">
      <span>{application.appliedAt ? `Applied ${new Date(application.appliedAt).toLocaleDateString()}` : 'Application date unavailable'}</span>
      <div className="flex items-center gap-3">{application.resumeUrl && <a href={application.resumeUrl} target="_blank" rel="noreferrer" className="inline-flex items-center gap-1 font-semibold text-[var(--text-button-fill)]"><FileText size={15} />CV</a>}{application.jobId && <Link to={`/discover/${application.jobId}`} className="inline-flex items-center gap-1 font-semibold text-[var(--text-button-fill)]">Job <ExternalLink size={14} /></Link>}</div>
    </div>
  </article>
));
