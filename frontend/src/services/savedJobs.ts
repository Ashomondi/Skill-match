import { API_BASE_URL, authHeaders, getErrorMessage } from './api';

export interface SavedJob {
  id: string;
  jobId: string;
  title: string;
  company: string;
  location: string;
  workType?: string;
  salary?: string;
  matchScore?: number;
  savedAt?: string;
}

const normalizeSavedJob = (item: any): SavedJob => {
  const job = item.job || item.job_details || item;
  return {
    id: String(item.id ?? item.saved_job_id ?? job.id ?? job.job_id),
    jobId: String(item.job_id ?? item.jobId ?? job.id ?? job.job_id ?? item.id),
    title: job.title ?? job.job_title ?? 'Untitled role',
    company: job.company ?? job.company_name ?? 'Company not provided',
    location: job.location ?? 'Location not provided',
    workType: job.work_type ?? job.workType ?? job.employment_type,
    salary: job.salary ?? job.salary_range,
    matchScore: Number(job.match_score ?? job.matchScore) || undefined,
    savedAt: item.saved_at ?? item.created_at ?? item.savedAt,
  };
};

export const savedJobsService = {
  async list(): Promise<SavedJob[]> {
    const response = await fetch(`${API_BASE_URL}/saved-jobs`, { headers: authHeaders() });
    if (!response.ok) throw new Error(await getErrorMessage(response, 'Saved jobs could not be loaded.'));
    const body = await response.json().catch(() => []);
    const items = Array.isArray(body) ? body : body.saved_jobs ?? body.savedJobs ?? body.data ?? [];
    return Array.isArray(items) ? items.map(normalizeSavedJob) : [];
  },

  async remove(idOrJobId: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/saved-jobs/${encodeURIComponent(idOrJobId)}`, {
      method: 'DELETE',
      headers: authHeaders(),
    });
    if (!response.ok) throw new Error(await getErrorMessage(response, 'The saved job could not be removed.'));
  },

  async save(jobId: string): Promise<SavedJob> {
    const response = await fetch(`${API_BASE_URL}/saved-jobs`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeaders() },
      body: JSON.stringify({ job_id: jobId }),
    });
    if (!response.ok) throw new Error(await getErrorMessage(response, 'The job could not be saved.'));
    const body = await response.json().catch(() => ({}));
    return normalizeSavedJob(body.saved_job ?? body.data ?? body);
  },

  async isSaved(jobId: string): Promise<boolean> {
    const jobs = await this.list();
    return jobs.some((job) => job.jobId === jobId);
  },
};
