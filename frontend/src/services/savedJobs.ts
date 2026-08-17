const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

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

const headers = (): Record<string, string> => {
  const token = localStorage.getItem('token');
  return token ? { Authorization: `Bearer ${token}` } : {};
};

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

const errorMessage = async (response: Response, fallback: string) => {
  const body = await response.json().catch(() => ({}));
  return body.error || body.message || fallback;
};

export const savedJobsService = {
  async list(): Promise<SavedJob[]> {
    const response = await fetch(`${API_BASE_URL}/saved-jobs`, { headers: headers() });
    if (!response.ok) throw new Error(await errorMessage(response, 'Saved jobs could not be loaded.'));
    const body = await response.json().catch(() => []);
    const items = Array.isArray(body) ? body : body.saved_jobs ?? body.savedJobs ?? body.data ?? [];
    return Array.isArray(items) ? items.map(normalizeSavedJob) : [];
  },

  async remove(id: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/saved-jobs/${id}`, { method: 'DELETE', headers: headers() });
    if (!response.ok) throw new Error(await errorMessage(response, 'The saved job could not be removed.'));
  },

  async save(jobId: string): Promise<SavedJob> {
    const response = await fetch(`${API_BASE_URL}/saved-jobs`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...headers() },
      body: JSON.stringify({ job_id: jobId }),
    });
    if (!response.ok) throw new Error(await errorMessage(response, 'The job could not be saved.'));
    const body = await response.json().catch(() => ({}));
    return normalizeSavedJob(body.saved_job ?? body);
  },

  async isSaved(jobId: string): Promise<boolean> {
    const jobs = await this.list();
    return jobs.some((job) => job.jobId === jobId);
  },
};
