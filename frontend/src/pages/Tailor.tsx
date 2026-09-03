import React, { useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { AlertCircle, CheckCircle2, ChevronLeft, Loader2, RefreshCw, Scissors, Send } from 'lucide-react';
import { AppShell } from '../components/AppShell';
import { jobsService, Job } from '../services/jobs';
import { resumeService, Resume } from '../services/resume';
import { tailoringService, TailoringUnavailableError } from '../services/tailoring';
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

const draftKey = (jobId: string) => `tailor:draft:${jobId}`;

export const Tailor: React.FC = () => {
  const { jobId = '' } = useParams();
  const [job, setJob] = useState<Job | null>(null);
  const [resume, setResume] = useState<Resume | null>(null);
  const [content, setContent] = useState('');
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [aiUnavailable, setAiUnavailable] = useState(false);
  const [submitted, setSubmitted] = useState(false);

  useEffect(() => {
    if (!jobId) { setLoading(false); return; }
    setContent(sessionStorage.getItem(draftKey(jobId)) || '');
    void (async () => {
      try {
        const [loadedJob, resumes] = await Promise.all([jobsService.get(jobId), resumeService.list()]);
        setJob(loadedJob); setResume(resumes[0] || null);
      } catch (err) { setError(err instanceof Error ? err.message : 'Could not load tailoring data.'); }
      finally { setLoading(false); }
    })();
  }, [jobId]);

  useEffect(() => {
    if (!jobId || !content) return;
    sessionStorage.setItem(draftKey(jobId), content);
  }, [content, jobId]);

  if (!jobId) return <JobPicker />;

  const generate = async () => {
    if (!job || !resume) return;
    setGenerating(true); setError(''); setAiUnavailable(false);
    try { setContent(await tailoringService.generate({ resumeId: resume.id, jobTitle: job.title, company: job.company, jobDescription: job.description, currentContent: content })); }
    catch (err) {
      if (err instanceof TailoringUnavailableError) setAiUnavailable(true);
      else setError(err instanceof Error ? err.message : 'CV tailoring failed.');
    }
    finally { setGenerating(false); }
  };

  const submit = async () => {
    if (!job) return;
    setSubmitting(true); setError('');
    try {
      await applicationService.create(job.id);
      setSubmitted(true);
      sessionStorage.removeItem(draftKey(jobId));
    }
    catch (err) { setError(err instanceof Error ? err.message : 'Application could not be submitted.'); }
    finally { setSubmitting(false); }
  };

  return <AppShell>
    <Link to={`/discover/${jobId}`} className="inline-flex items-center gap-1 text-sm text-[var(--text-button-fill)]"><ChevronLeft size={16}/>Back to job details</Link>
    <h1 className="mt-5 font-serif text-4xl font-bold text-[var(--text-heading)]">Tailor your CV</h1>
    {loading ? <div className="mt-8 flex items-center gap-2 text-sm"><Loader2 className="animate-spin" size={18}/>Loading job and resume...</div>
      : error && !job ? <p role="alert" className="mt-8 text-sm text-[var(--status-rejected)]">{error}</p>
      : <>
        <p className="mt-2 text-sm text-[var(--text-muted)]">{job?.title} at {job?.company}</p>
        {!resume && <div className="mt-6 flex items-start gap-3 rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-secondary)] p-4 text-sm"><AlertCircle size={18} className="mt-0.5 shrink-0 text-[var(--accent-gold)]" /><span>Upload a resume before tailoring. <Link className="font-semibold text-[var(--text-button-fill)]" to="/resume">Open resume manager</Link></span></div>}
        {resume && <section className="mt-6 max-w-4xl">
          <textarea value={content} onChange={event => setContent(event.target.value)} placeholder="Generate a tailored CV, then review and edit it here." className="min-h-[30rem] w-full rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-input)] p-5 text-sm leading-6 outline-none focus:border-[var(--accent-gold)]" aria-label="Tailored CV content" />
          {aiUnavailable && <div role="alert" className="mt-3 flex items-start gap-2 rounded border border-[var(--accent-gold)] bg-[var(--bg-secondary)] p-3 text-sm"><AlertCircle size={16} className="mt-0.5 shrink-0 text-[var(--accent-gold)]" /><span>The AI service is temporarily unavailable. Your draft is safe below — try regenerating in a moment.</span></div>}
          {error && <p role="alert" className="mt-3 text-sm text-[var(--status-rejected)]">{error}</p>}
          {submitted && <div role="status" className="mt-4 flex items-center gap-2 rounded border border-[var(--status-offer)] bg-[var(--status-offer)]/10 p-3 text-sm"><CheckCircle2 size={16} className="text-[var(--status-offer)]" /><span>Application submitted. <Link className="font-semibold text-[var(--text-button-fill)]" to="/applications">Track it here</Link>.</span></div>}
          <div className="mt-4 flex flex-wrap gap-3">
            <button type="button" onClick={() => void generate()} disabled={generating || submitting} className="inline-flex items-center gap-2 rounded border border-[var(--text-button-fill)] px-4 py-2 text-sm disabled:opacity-50">{generating ? <Loader2 className="animate-spin" size={16}/> : <RefreshCw size={16}/>}{generating ? 'Generating…' : content ? 'Regenerate' : 'Generate tailored CV'}</button>
            <button type="button" onClick={() => void submit()} disabled={generating || submitting || !content || submitted} className="inline-flex items-center gap-2 rounded bg-[var(--btn-primary-bg)] px-4 py-2 text-sm text-[var(--btn-primary-text)] disabled:opacity-50">{submitting ? <Loader2 className="animate-spin" size={16}/> : <Send size={16}/>}{submitting ? 'Submitting…' : submitted ? 'Application submitted' : 'Submit application'}</button>
          </div>
        </section>}
      </>}
  </AppShell>;
};
