const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

export interface Job {
  id: string;
  title: string;
  company: string;
  location: string;
  workType?: string;
  seniority?: string;
  salary?: string;
  postedAt?: string;
  matchScore?: number;
}

export interface JobSearchParams {
  query?: string;
  location?: string;
  seniority?: string;
  workType?: string;
}

export interface JobSearchResult {
  jobs: Job[];
  total: number;
}

const normalizeJob = (item: any): Job => ({
  id: String(item.id ?? item.job_id),
  title: item.title ?? item.job_title ?? 'Untitled role',
  company: item.company ?? item.company_name ?? 'Company not provided',
  location: item.location ?? 'Location not provided',
  workType: item.work_type ?? item.workType ?? item.employment_type,
  seniority: item.seniority ?? item.seniority_level ?? item.experience_level,
  salary: item.salary ?? item.salary_range,
  postedAt: item.posted_at ?? item.created_at ?? item.postedAt,
  matchScore: Number(item.match_score ?? item.matchScore) || undefined,
});

export const jobsService = {
  async get(id: string): Promise<Job> {
    const token = localStorage.getItem("token");
    const response = await fetch(API_BASE_URL + "/jobs/" + encodeURIComponent(id), { headers: token ? { Authorization: "Bearer " + token } : {} });
    const body = await response.json().catch(() => null);
    if (!response.ok) throw new Error(body?.error?.message || body?.error || body?.message || "Job could not be loaded.");
    return normalizeJob(body?.data ?? body);
  },

  async search(params: JobSearchParams = {}): Promise<JobSearchResult> {
    const searchParams = new URLSearchParams();
    if (params.query?.trim()) searchParams.set('q', params.query.trim());
    if (params.location) searchParams.set('location', params.location);
    if (params.seniority) searchParams.set('seniority', params.seniority);
    if (params.workType) searchParams.set('work_type', params.workType);

    const token = localStorage.getItem('token');
    const response = await fetch(`${API_BASE_URL}/jobs/search${searchParams.size ? `?${searchParams}` : ''}`, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    });
    const body = await response.json().catch(() => null);
    if (!response.ok) throw new Error(body?.error || body?.message || 'Jobs could not be loaded.');

    const items = Array.isArray(body) ? body : body?.jobs ?? body?.results ?? body?.data ?? [];
    const jobs = Array.isArray(items) ? items.map(normalizeJob) : [];
    return { jobs, total: Number(body?.total ?? body?.count ?? jobs.length) };
  },
};
