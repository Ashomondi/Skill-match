import React from 'react';
import { SavedJob } from '../services/savedJobs';
import { SavedJobCard } from './saved-jobs/SavedJobCard';

interface SavedJobsProps {
  jobs: SavedJob[];
  removingId: string | null;
  onRemove: (savedJob: SavedJob) => void;
}

export const SavedJobs: React.FC<SavedJobsProps> = ({ jobs, removingId, onRemove }) => (
  <div className="grid overflow-hidden border-y border-[var(--border-hairline)] sm:grid-cols-2 sm:gap-4 sm:overflow-visible sm:border-0 xl:grid-cols-3">
    {jobs.map((savedJob) => <SavedJobCard key={savedJob.id} savedJob={savedJob} removing={removingId === savedJob.id} onRemove={onRemove} />)}
  </div>
);
