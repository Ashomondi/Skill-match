import React, { useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { ChevronLeft, Loader2, RefreshCw, Scissors, Send } from 'lucide-react';
import { AppShell } from '../components/AppShell';
import { jobsService, Job } from '../services/jobs';
import { resumeService, Resume } from '../services/resume';
import { tailoringService } from '../services/tailoring';
import { applicationService } from '../services/application';

const JobPicker: React.FC = () => {
  const navigate = useNavigate();
  const [jobs, setJobs] = useState<Job[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => { void (async () => {
    try { setJobs((await jobsService.search({})).jobs); }
    catch (err) { setError(err instanceof Error ? err.message : 'Jobs could not be loaded.'); }
    finally { setLoading(false); }
  })(); }, []);

  return (
    <AppShell>
      <Link to="/discover" className="inline-flex items-center gap-1 text-sm text-[var(--text-button-fill)]"><ChevronLeft size={16}/>Back to discover</Link>
      <h1 className="mt-5 font-serif text-4xl font-bold text-[var(--text-heading)]">Tailor your CV</h1>
      <p className="mt-2 max-w-2xl text-sm leading-6 text-[var(--text-muted)]">Pick a role and we will tailor your CV to match it. Choose a job below to get started.</p>
      {loading ? <div className="mt-8 flex items-center gap-2 text-sm"><Loader2 className="animate-spin" size={18}/>Loading jobs...</div>
        : error ? <p role="alert" className="mt-8 text-sm text-[var(--status-rejected)]">{error}</p>
        : jobs.length === 0 ? <div className="mt-8 rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-secondary)] p-8 text-center text-sm text-[var(--text-muted)]">No jobs are available right now. <Link to="/discover" className="text-[var(--text-button-fill)]">Browse discover</Link>.</div>
        : <ul className="mt-8 divide-y divide-[var(--border-hairline)] rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-secondary)]">
            {jobs.map((job) => <li key={job.id} className="flex items-center gap-3 px-5 py-4">
              <span className="grid h-10 w-10 shrink-0 place-items-center rounded-full bg-[var(--bg-card)] font-semibold text-[var(--text-heading)]">{job.company.charAt(0).toUpperCase()}</span>
              <div className="min-w-0 flex-1"><p className="truncate font-semibold text-[var(--text-heading)]">{job.title}</p><p className="truncate text-xs text-[var(--text-muted)]">{job.company} • {job.location}</p></div>
              <button type="button" onClick={() => navigate(`/discover/${job.id}/tailor`)} className="inline-flex shrink-0 items-center gap-2 rounded bg-[var(--btn-primary-bg)] px-4 py-2 text-sm font-semibold text-[var(--btn-primary-text)]"><Scissors size={15}/>Tailor</button>
            </li>)}
          </ul>}
    </AppShell>
  );
};

export const Tailor: React.FC = () => {
  const { jobId = '' } = useParams();
  const [job, setJob] = useState<Job | null>(null);
  const [resume, setResume] = useState<Resume | null>(null);
  const [content, setContent] = useState('');
  const [loading, setLoading] = useState(true);
  const [working, setWorking] = useState(false);
  const [error, setError] = useState('');
  const [submitted, setSubmitted] = useState(false);

  useEffect(() => { if (!jobId) { setLoading(false); return; } void (async () => {
    try {
      const [loadedJob, resumes] = await Promise.all([jobsService.get(jobId), resumeService.list()]);
      setJob(loadedJob); setResume(resumes[0] || null);
    } catch (err) { setError(err instanceof Error ? err.message : 'Could not load tailoring data.'); }
    finally { setLoading(false); }
  })(); }, [jobId]);

  if (!jobId) return <JobPicker />;

  const generate = async () => {
    if (!job || !resume) return;
    setWorking(true); setError('');
    try { setContent(await tailoringService.generate({ resumeId: resume.id, jobTitle: job.title, company: job.company, jobDescription: job.description, currentContent: content })); }
    catch (err) { setError(err instanceof Error ? err.message : 'CV tailoring failed.'); }
    finally { setWorking(false); }
  };

  const submit = async () => {
    if (!job) return;
    setWorking(true); setError('');
    try { await applicationService.create(job.id); setSubmitted(true); }
    catch (err) { setError(err instanceof Error ? err.message : 'Application could not be submitted.'); }
    finally { setWorking(false); }
  };

  return <AppShell><Link to={`/discover/${jobId}`} className="inline-flex items-center gap-1 text-sm text-[var(--text-button-fill)]"><ChevronLeft size={16}/>Back to job details</Link><h1 className="mt-5 font-serif text-4xl font-bold text-[var(--text-heading)]">Tailor your CV</h1>{loading ? <div className="mt-8 flex items-center gap-2 text-sm"><Loader2 className="animate-spin" size={18}/>Loading job and resume...</div> : error && !job ? <p className="mt-8 text-sm text-[var(--status-rejected)]">{error}</p> : <><p className="mt-2 text-sm text-[var(--text-muted)]">{job?.title} at {job?.company}</p>{!resume && <p className="mt-6 text-sm">Upload a resume before tailoring. <Link className="text-[var(--text-button-fill)]" to="/resume">Open resume manager</Link></p>} {resume && <section className="mt-6 max-w-4xl"><textarea value={content} onChange={event => setContent(event.target.value)} placeholder="Generate a tailored CV, then review and edit it here." className="min-h-[30rem] w-full rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-input)] p-5 text-sm leading-6 outline-none focus:border-[var(--accent-gold)]" aria-label="Tailored CV content" />{error && <p className="mt-3 text-sm text-[var(--status-rejected)]">{error}</p>}<div className="mt-4 flex flex-wrap gap-3"><button type="button" onClick={() => void generate()} disabled={working} className="inline-flex items-center gap-2 rounded border border-[var(--text-button-fill)] px-4 py-2 text-sm disabled:opacity-50"><RefreshCw size={16}/>{content ? 'Regenerate' : 'Generate tailored CV'}</button><button type="button" onClick={() => void submit()} disabled={working || !content || submitted} className="inline-flex items-center gap-2 rounded bg-[var(--btn-primary-bg)] px-4 py-2 text-sm text-[var(--btn-primary-text)] disabled:opacity-50"><Send size={16}/>{submitted ? 'Application submitted' : 'Submit application'}</button></div></section>}</>}</AppShell>;
};
