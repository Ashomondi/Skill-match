import React, { useCallback, useEffect, useState } from 'react';
import { AlertCircle, Bookmark, Briefcase, Building2, CheckCircle2, ChevronLeft, ChevronLeft as PrevPage, ChevronRight as NextPage, MapPin, RefreshCw, Search, Send } from 'lucide-react';
import { Link, useParams } from 'react-router-dom';
import { AppShell } from '../components/AppShell';
import { JobCard } from '../components/JobCard';
import { MatchRing } from '../components/MatchRing';
import { RecommendationsSection } from '../components/jobs/RecommendationsSection';
import { useJobs } from '../hooks/useJobs';
import { Job, jobsService } from '../services/jobs';
import { applicationService } from '../services/application';
import { savedJobsService } from '../services/savedJobs';

const filterClassName = 'h-11 rounded-md border border-[var(--border-hairline)] bg-[var(--bg-input)] px-3 text-sm text-[var(--text-heading)] outline-none focus:border-[var(--accent-gold)]';

export const Jobs: React.FC = () => {
  const [query, setQuery] = useState('');
  const [location, setLocation] = useState('');
  const [seniority, setSeniority] = useState('');
  const [workType, setWorkType] = useState('');
  const { jobs, total, totalPages, page, goToPage, loading, error, retry } = useJobs({ query, location, seniority, workType });

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

        {!loading && !error && total > 0 && totalPages > 1 && <nav className="mt-6 flex items-center justify-center gap-4" aria-label="Pagination">
          <button type="button" disabled={page <= 1} onClick={() => goToPage(page - 1)} aria-label="Previous page" className="inline-flex h-10 items-center gap-1 rounded-md border border-[var(--border-hairline)] bg-[var(--bg-secondary)] px-3 text-sm font-semibold text-[var(--text-heading)] disabled:cursor-not-allowed disabled:opacity-40"><PrevPage size={15} />Prev</button>
          <span className="text-sm text-[var(--text-muted)]">Page <b className="text-[var(--text-heading)]">{page}</b> of {totalPages}</span>
          <button type="button" disabled={page >= totalPages} onClick={() => goToPage(page + 1)} aria-label="Next page" className="inline-flex h-10 items-center gap-1 rounded-md border border-[var(--border-hairline)] bg-[var(--bg-secondary)] px-3 text-sm font-semibold text-[var(--text-heading)] disabled:cursor-not-allowed disabled:opacity-40">Next<NextPage size={15} /></button>
        </nav>}
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
  const [applying, setApplying] = useState(false);
  const [applied, setApplied] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);
  const [actionMessage, setActionMessage] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!jobId) return;
    setLoading(true);
    setError(null);
    try {
      const details = await jobsService.get(jobId);
      setJob(details);
      setSaved(await savedJobsService.isSaved(jobId).catch(() => false));
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : 'The job could not be loaded.');
    } finally {
      setLoading(false);
    }
  }, [jobId]);

  useEffect(() => { void load(); }, [load]);

  const saveJob = async () => {
    if (!job) return;
    setSaving(true);
    setActionError(null);
    setActionMessage(null);
    try {
      await savedJobsService.save(job.id);
      setSaved(true);
      setActionMessage('Job saved to your shortlist.');
    } catch (saveError) {
      setActionError(saveError instanceof Error ? saveError.message : 'The job could not be saved.');
    } finally {
      setSaving(false);
    }
  };

  const apply = async () => {
    if (!job) return;
    setApplying(true);
    setActionError(null);
    setActionMessage(null);
    try {
      await applicationService.apply(job.id);
      setApplied(true);
      setActionMessage('Application submitted. Good luck!');
    } catch (applyError) {
      setActionError(applyError instanceof Error ? applyError.message : 'Your application could not be submitted.');
    } finally {
      setApplying(false);
    }
  };

  return (
    <AppShell>
      <Link to="/discover" className="inline-flex items-center gap-1 text-sm text-[var(--text-button-fill)]"><ChevronLeft size={16} />Back to results</Link>
      {loading ? <div className="mt-6 space-y-3" aria-label="Loading job details">{[1, 2, 3].map((item) => <div key={item} className="h-24 animate-pulse rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-card)]" />)}</div> : error ? <div role="alert" className="mt-6 flex flex-col gap-3 rounded-md border border-[var(--status-rejected)] bg-[var(--status-rejected)]/10 p-4 text-sm text-[var(--text-heading)] sm:flex-row sm:items-center sm:justify-between"><span className="flex items-center gap-2"><AlertCircle className="shrink-0 text-[var(--status-rejected)]" size={18} />{error}</span><button type="button" onClick={() => void load()} className="inline-flex items-center justify-center gap-2 rounded-md border border-[var(--text-button-fill)] px-3 py-2 font-semibold text-[var(--text-button-fill)]"><RefreshCw size={15} />Try again</button></div> : job && <div className="mt-5 grid gap-8 lg:grid-cols-[1fr_310px]">
        <article>
          <div className="flex items-center gap-3">
            <span className="grid h-12 w-12 place-items-center rounded-full bg-[var(--bg-card)] font-serif text-lg font-bold text-[var(--text-heading)]">{job.company.charAt(0).toUpperCase()}</span>
            <div>
              <h1 className="font-serif text-3xl font-bold text-[var(--text-heading)] sm:text-4xl">{job.title}</h1>
              <p className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-[var(--text-muted)]"><span className="inline-flex items-center gap-1.5"><Building2 size={14} />{job.company}</span><span className="inline-flex items-center gap-1.5"><MapPin size={14} />{job.location}</span></p>
            </div>
          </div>
          <div className="mt-4 flex flex-wrap gap-2">
            {job.workType && <span className="rounded-full bg-[var(--bg-chip)] px-3 py-1 text-sm text-[var(--text-insight)]">{job.workType}</span>}
            {job.seniority && <span className="rounded-full bg-[var(--bg-chip)] px-3 py-1 text-sm text-[var(--text-insight)]">{job.seniority}</span>}
            {job.salary && <span className="rounded-full bg-[var(--bg-chip)] px-3 py-1 text-sm text-[var(--text-insight)]">{job.salary}</span>}
            {job.postedAt && <span className="rounded-full bg-[var(--bg-chip)] px-3 py-1 text-sm text-[var(--text-insight)]">Posted {new Date(job.postedAt).toLocaleDateString()}</span>}
          </div>
          <section className="mt-8"><h2 className="font-serif text-2xl font-bold text-[var(--text-heading)]">About this role</h2><p className="mt-3 whitespace-pre-wrap leading-7 text-[var(--text-body)]">{job.description || 'No description provided for this role.'}</p></section>
          {job.skills && job.skills.length > 0 && <section className="mt-8"><h2 className="font-serif text-2xl font-bold text-[var(--text-heading)]">Skills</h2><div className="mt-3 flex flex-wrap gap-2">{job.skills.map((skill) => <span key={skill} className="rounded-full bg-[var(--bg-chip)] px-3 py-1 text-sm text-[var(--text-insight)]">{skill}</span>)}</div></section>}
          {job.sourceUrl && <section className="mt-8"><a href={job.sourceUrl} target="_blank" rel="noreferrer" className="text-sm font-semibold text-[var(--text-button-fill)] underline underline-offset-2">View original listing</a></section>}
          <Link to={`/discover/${job.id}/tailor`} className="mt-8 inline-block rounded bg-[var(--btn-primary-bg)] px-5 py-3 text-sm font-semibold text-[var(--btn-primary-text)]">Tailor my CV for this role</Link>
        </article>
        <aside className="h-fit rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-secondary)] p-5">
          {actionMessage && <div className="mb-4 flex items-start gap-2 rounded-md border border-[#7A8B6F] bg-[#7A8B6F]/10 p-3 text-sm text-[var(--text-heading)]"><CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-[#7A8B6F]" />{actionMessage}</div>}
          {actionError && <div role="alert" className="mb-4 flex items-start gap-2 rounded-md border border-[var(--status-rejected)] bg-[var(--status-rejected)]/10 p-3 text-sm text-[var(--text-heading)]"><AlertCircle className="mt-0.5 h-4 w-4 shrink-0 text-[var(--status-rejected)]" />{actionError}</div>}
          <h2 className="font-serif text-xl font-bold text-[var(--text-heading)]">Quick actions</h2>
          {job.matchScore !== undefined && <div className="my-5 flex justify-center"><MatchRing value={job.matchScore} size={110} /></div>}
          <div className="space-y-3">
            <button type="button" onClick={() => void apply()} disabled={applying || applied} className="flex w-full items-center justify-center gap-2 rounded bg-[var(--btn-primary-bg)] py-3 text-sm font-semibold text-[var(--btn-primary-text)] disabled:cursor-not-allowed disabled:opacity-60">{applying ? <><RefreshCw className="animate-spin" size={15} />Submitting…</> : applied ? <><CheckCircle2 size={16} />Applied</> : <><Send size={15} />Apply for this job</>}</button>
            <button type="button" onClick={() => void saveJob()} disabled={saving || saved} className="flex w-full items-center justify-center gap-2 rounded border border-[var(--text-button-fill)] py-3 text-sm font-semibold text-[var(--text-button-fill)] disabled:cursor-not-allowed disabled:opacity-60">{saving ? <><RefreshCw className="animate-spin" size={15} />Saving…</> : saved ? <><CheckCircle2 size={16} />Saved for later</> : <><Bookmark size={16} />Save for later</>}</button>
            <Link to={`/discover/${job.id}/tailor`} className="block rounded border border-[var(--border-hairline)] bg-[var(--bg-primary)] py-3 text-center text-sm font-semibold text-[var(--text-button-fill)]">Tailor my CV for this role</Link>
          </div>
        </aside>
      </div>}
    </AppShell>
  );
};
