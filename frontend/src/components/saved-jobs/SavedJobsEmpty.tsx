import React from 'react';
import { Bookmark } from 'lucide-react';
import { Link } from 'react-router-dom';

export const SavedJobsEmpty: React.FC = () => (
  <div className="border-y border-[var(--border-hairline)] bg-[var(--bg-secondary)] px-6 py-16 text-center sm:rounded-lg sm:border">
    <Bookmark className="mx-auto text-[var(--accent-gold)]" size={28} />
    <h2 className="mt-4 font-serif text-2xl font-semibold text-[var(--text-heading)]">No saved jobs yet</h2>
    <p className="mx-auto mt-2 max-w-md text-sm leading-6 text-[var(--text-muted)]">Save promising roles while exploring matches and they will appear here.</p>
    <Link to="/discover" className="mt-6 inline-flex rounded-md bg-[var(--btn-primary-bg)] px-5 py-3 text-sm font-semibold text-[var(--btn-primary-text)]">Discover jobs</Link>
  </div>
);
