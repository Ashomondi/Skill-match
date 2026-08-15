import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { AlertCircle, Briefcase, RefreshCw, Search } from 'lucide-react';
import { AppShell } from '../components/AppShell';
import { ApplicationCard } from '../components/ApplicationCard';
import { Application, ApplicationStatus, applicationService } from '../services/application';

const filters: Array<'all' | ApplicationStatus> = ['all', 'applied', 'screening', 'interview', 'offer', 'rejected', 'withdrawn'];

export const Applications: React.FC = () => {
  const [applications, setApplications] = useState<Application[]>([]);
  const [active, setActive] = useState<(typeof filters)[number]>('all');
  const [query, setQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const load = useCallback(async () => { setLoading(true); setError(null); try { setApplications(await applicationService.list()); } catch (loadError) { setError(loadError instanceof Error ? loadError.message : 'Applications could not be loaded.'); } finally { setLoading(false); } }, []);
  useEffect(() => { void load(); }, [load]);
  const visible = useMemo(() => applications.filter((item) => (active === 'all' || item.status === active) && `${item.role} ${item.company}`.toLowerCase().includes(query.trim().toLowerCase())), [applications, active, query]);
  const count = (status: (typeof filters)[number]) => status === 'all' ? applications.length : applications.filter((item) => item.status === status).length;

  return <AppShell><div className="mx-auto w-full max-w-6xl"><header><p className="text-xs font-semibold uppercase tracking-[0.16em] text-[var(--accent-gold)]">Application tracker</p><h1 className="mt-2 font-serif text-4xl font-bold text-[var(--text-heading)]">Your applications</h1><p className="mt-2 text-sm text-[var(--text-muted)]">Track progress and return to every role in your pipeline.</p></header>
    <div className="mt-7 flex gap-2 overflow-x-auto pb-2" aria-label="Filter applications">{filters.map((filter) => <button type="button" key={filter} onClick={() => setActive(filter)} aria-pressed={active === filter} className={`shrink-0 rounded-full px-4 py-2 text-sm capitalize ${active === filter ? 'bg-[var(--tab-active-bg)] text-[var(--tab-active-text)]' : 'bg-[var(--bg-chip)] text-[var(--text-heading)]'}`}>{filter} <span className="opacity-70">{count(filter)}</span></button>)}</div>
    <label className="mt-4 flex items-center gap-2 rounded-md border border-[var(--border-hairline)] bg-[var(--bg-input)] px-3 focus-within:border-[var(--accent-gold)]"><Search size={18} /><input value={query} onChange={(event) => setQuery(event.target.value)} className="h-11 min-w-0 flex-1 bg-transparent text-sm outline-none" placeholder="Search roles or companies" aria-label="Search applications" /></label>
    {error && <div role="alert" className="mt-5 flex flex-col gap-3 rounded-md border border-[var(--status-rejected)] bg-[var(--status-rejected)]/10 p-4 text-sm sm:flex-row sm:items-center sm:justify-between"><span className="flex items-center gap-2"><AlertCircle size={18} />{error}</span><button type="button" onClick={() => void load()} className="inline-flex items-center justify-center gap-2 rounded-md border border-[var(--text-button-fill)] px-3 py-2 font-semibold"><RefreshCw size={15} />Try again</button></div>}
    {loading ? <div className="mt-5 grid gap-4 sm:grid-cols-2 xl:grid-cols-3" aria-label="Loading applications">{[1,2,3].map((item) => <div key={item} className="h-40 animate-pulse rounded-lg bg-[var(--bg-card)]" />)}</div> : !error && visible.length === 0 ? <div className="mt-5 border-y border-[var(--border-hairline)] bg-[var(--bg-secondary)] px-6 py-16 text-center sm:rounded-lg sm:border"><Briefcase className="mx-auto text-[var(--accent-gold)]" /><h2 className="mt-3 font-serif text-2xl font-semibold text-[var(--text-heading)]">{applications.length ? 'No matching applications' : 'No applications yet'}</h2><p className="mt-2 text-sm text-[var(--text-muted)]">{applications.length ? 'Try another filter or search term.' : 'Applications you submit will appear here.'}</p></div> : <div className="mt-5 grid overflow-hidden border-y border-[var(--border-hairline)] sm:grid-cols-2 sm:gap-4 sm:overflow-visible sm:border-0 xl:grid-cols-3">{visible.map((application) => <ApplicationCard key={application.id} application={application} />)}</div>}
  </div></AppShell>;
};
