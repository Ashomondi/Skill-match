import React, { useEffect, useState } from 'react';
import { AlertCircle, Briefcase, ChevronLeft, RefreshCw, Search } from 'lucide-react';
import { Link, useParams } from 'react-router-dom';
import { AppShell } from '../components/AppShell';
import { JobCard } from '../components/JobCard';
import { MatchRing } from '../components/MatchRing';
import { RecommendationsSection } from '../components/jobs/RecommendationsSection';
import { useJobs } from '../hooks/useJobs';
import { savedJobsService } from '../services/savedJobs';
import { formatJobDescription, Job, jobsService } from '../services/jobs';

const filterClassName = 'h-11 rounded-md border border-[var(--border-hairline)] bg-[var(--bg-input)] px-3 text-sm text-[var(--text-heading)] outline-none focus:border-[var(--accent-gold)]';

export const Jobs: React.FC = () => {
  const [query, setQuery] = useState('');
  const [location, setLocation] = useState('');
  const [seniority, setSeniority] = useState('');
  const [workType, setWorkType] = useState('');
  const { jobs, total, loading, error, retry } = useJobs({ query, location, seniority, workType });

  return (
    <AppShell>
      <div className="mx-auto w-full max-w-6xl">
        <header className="max-w-2xl">
          <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[var(--accent-gold)]">Job discovery</p>
          <h1 className="mt-2 font-serif text-4xl font-bold text-[var(--text-heading)]">Find your next role</h1>
          <p className="mt-2 text-sm leading-6 text-[var(--text-muted)]">Search available roles and explore opportunities aligned with your experience.</p>
        </header>

        <RecommendationsSection />

        <div className="mt-10 border-t border-[var(--border-hairline)] pt-8">
          <h2 className="font-serif text-2xl font-bold text-[var(--text-heading)]">Explore all jobs</h2>
          <p className="mt-1 text-sm text-[var(--text-muted)]">Search beyond your personalized recommendations.</p>
        </div>
        <div className="mt-5 grid gap-3 lg:grid-cols-[minmax(260px,1fr)_180px_180px_180px]">
          <label className="flex min-w-0 items-center gap-2 rounded-md border border-[var(--border-hairline)] bg-[var(--bg-input)] px-3 focus-within:border-[var(--accent-gold)]">
            <Search className="shrink-0 text-[var(--text-muted)]" size={18} />
            <input value={query} onChange={(event) => setQuery(event.target.value)} className="h-11 min-w-0 flex-1 bg-transparent text-sm text-[var(--text-heading)] outline-none" placeholder="Search title, company, or keyword" aria-label="Search jobs" />
          </label>
          <select value={location} onChange={(event) => setLocation(event.target.value)} className={filterClassName} aria-label="Filter by location"><option value="">All locations</option><option value="remote">Remote</option><option value="hybrid">Hybrid</option><option value="onsite">On-site</option></select>
          <select value={seniority} onChange={(event) => setSeniority(event.target.value)} className={filterClassName} aria-label="Filter by seniority"><option value="">All levels</option><option value="entry">Entry level</option><option value="mid">Mid level</option><option value="senior">Senior</option><option value="lead">Lead</option></select>
          <select value={workType} onChange={(event) => setWorkType(event.target.value)} className={filterClassName} aria-label="Filter by work type"><option value="">All work types</option><option value="full-time">Full-time</option><option value="part-time">Part-time</option><option value="contract">Contract</option><option value="internship">Internship</option></select>
        </div>

        <div className="mt-8 flex items-center justify-between gap-3 text-sm text-[var(--text-muted)]">
          <p>{loading ? 'Searching jobs...' : `${total} ${total === 1 ? 'role' : 'roles'} found`}</p>
          {(query || location || seniority || workType) && <button type="button" onClick={() => { setQuery(''); setLocation(''); setSeniority(''); setWorkType(''); }} className="font-semibold text-[var(--text-button-fill)]">Clear filters</button>}
        </div>

        {error && <div role="alert" className="mt-4 flex flex-col gap-3 rounded-md border border-[var(--status-rejected)] bg-[var(--status-rejected)]/10 p-4 text-sm text-[var(--text-heading)] sm:flex-row sm:items-center sm:justify-between"><span className="flex items-center gap-2"><AlertCircle className="shrink-0 text-[var(--status-rejected)]" size={18} />{error}</span><button type="button" onClick={() => void retry()} className="inline-flex items-center justify-center gap-2 rounded-md border border-[var(--text-button-fill)] px-3 py-2 font-semibold text-[var(--text-button-fill)]"><RefreshCw size={15} />Try again</button></div>}

        {loading ? <div className="mt-4 space-y-3" aria-label="Loading jobs">{[1, 2, 3].map((item) => <div key={item} className="h-36 animate-pulse rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-card)] sm:h-28" />)}</div> : !error && jobs.length === 0 ? <div className="mt-4 border-y border-[var(--border-hairline)] bg-[var(--bg-secondary)] px-6 py-16 text-center sm:rounded-lg sm:border"><Briefcase className="mx-auto text-[var(--accent-gold)]" size={28} /><h2 className="mt-4 font-serif text-2xl font-semibold text-[var(--text-heading)]">No matching jobs</h2><p className="mx-auto mt-2 max-w-md text-sm leading-6 text-[var(--text-muted)]">Try a broader keyword or clear one of your filters.</p></div> : <div className="mt-4 space-y-3">{jobs.map((job) => <JobCard key={job.id} job={job} />)}</div>}
      </div>
    </AppShell>
  );
};

export const JobDetail: React.FC = () => {
  const { jobId } = useParams();
  const [job, setJob] = useState<Job | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  useEffect(() => { if (!jobId) return; void jobsService.get(jobId).then(setJob).catch((err) => setError(err instanceof Error ? err.message : "Job could not be loaded.")).finally(() => setLoading(false)); }, [jobId]);
  return (
    <AppShell>
      <Link to="/discover" className="inline-flex items-center gap-1 text-sm text-[var(--text-button-fill)]"><ChevronLeft size={16} />Back to results</Link>
      <div className="mt-5 grid gap-8 lg:grid-cols-[1fr_310px]">
        <article>
          {loading && <p className="mt-6 text-sm text-[var(--text-muted)]">Loading job...</p>}{error && <p role="alert" className="mt-6 text-sm text-[var(--status-rejected)]">{error}</p>}{job && <><h1 className="font-serif text-4xl font-bold text-[var(--text-heading)]">{job.title}</h1><p className="mt-2">{job.company} • {job.location}</p></>}
          {job && <div className="mt-4 flex flex-wrap gap-2">{[job.remote ? 'Remote' : 'On-site', job.workType, job.salary].filter(Boolean).map((item) => <span className="rounded-full bg-[var(--bg-chip)] px-3 py-1 text-sm" key={item}>{item}</span>)}</div>}
          {job && <section className="mt-8"><h2 className="font-serif text-2xl font-bold text-[var(--text-heading)]">About the Role</h2><p className="mt-3 whitespace-pre-wrap leading-7">{formatJobDescription(job.description) || 'No description provided.'}</p></section>}
          <Link to={`/discover/${jobId}/tailor`} className="mt-8 inline-block rounded bg-[var(--btn-primary-bg)] px-5 py-3 text-sm font-semibold text-[var(--btn-primary-text)]">Tailor my CV for this role</Link>
        </article>
        <aside className="h-fit rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-secondary)] p-5">
          <h2 className="font-serif text-2xl font-bold text-[var(--text-heading)]">Why this matches you</h2>
          <div className="my-5 flex justify-center"><MatchRing value={85} size={110} /></div>
          <ul className="space-y-3 text-sm leading-6 text-[var(--text-insight)]"><li>Your 6 years of experience aligns with the “5+ years” requirement.</li><li>Strong overlap in B2B platform design.</li><li>Missing explicit mention of Figma prototyping, consider highlighting this.</li></ul>
          <Link to={`/discover/${jobId}/tailor`} className="mt-6 block rounded bg-[var(--btn-primary-bg)] py-3 text-center text-sm font-semibold text-[var(--btn-primary-text)]">Tailor my CV for this role</Link>
          <button type="button" disabled={!job || saving || saved} onClick={() => { if (!jobId) return; setSaving(true); setError(null); void savedJobsService.save(jobId).then(() => setSaved(true)).catch((err) => setError(err instanceof Error ? err.message : "The job could not be saved.")).finally(() => setSaving(false)); }} className="mt-3 w-full rounded border border-[var(--text-button-fill)] py-3 text-sm text-[var(--text-button-fill)] disabled:opacity-50">{saving ? "Saving..." : saved ? "Saved" : "Save for later"}</button>
        </aside>
      </div>
    </AppShell>
  );
};
