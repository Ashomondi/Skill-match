const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

export interface Job {
  id: string;
  title: string;
  company: string;
  location: string;
  description?: string;
  skills?: string[];
  workType?: string;
  seniority?: string;
  salary?: string;
  sourceUrl?: string;
  postedAt?: string;
  matchScore?: number;
}

export interface JobSearchParams {
  query?: string;
  location?: string;
  seniority?: string;
  workType?: string;
  page?: number;
  pageSize?: number;
}

export interface JobSearchResult {
  jobs: Job[];
  total: number;
  page: number;
  totalPages: number;
}

const authHeaders = (): Record<string, string> => {
  const token = localStorage.getItem('token');
  return token ? { Authorization: `Bearer ${token}` } : {};
};

const errorMessage = async (response: Response, fallback: string) => {
  const body = await response.json().catch(() => ({}));
  return body.error || body.message || fallback;
};

const normalizeJob = (item: any): Job => ({
  id: String(item.id ?? item.job_id),
  title: item.title ?? item.job_title ?? 'Untitled role',
  company: item.company ?? item.company_name ?? 'Company not provided',
  location: item.location ?? 'Location not provided',
  description: item.description ?? item.summary,
  skills: Array.isArray(item.skills) ? item.skills.map(String) : undefined,
  workType: item.work_type ?? item.workType ?? item.employment_type,
  seniority: item.seniority ?? item.seniority_level ?? item.experience_level,
  salary: item.salary ?? item.salary_range,
  sourceUrl: item.source_url ?? item.sourceUrl ?? item.apply_url,
  postedAt: item.posted_at ?? item.created_at ?? item.postedAt,
  matchScore: Number(item.match_score ?? item.matchScore) || undefined,
});

export const jobsService = {
  async search(params: JobSearchParams = {}): Promise<JobSearchResult> {
    const searchParams = new URLSearchParams();
    if (params.query?.trim()) searchParams.set('q', params.query.trim());
    if (params.location) searchParams.set('location', params.location);
    if (params.seniority) searchParams.set('seniority', params.seniority);
    if (params.workType) searchParams.set('work_type', params.workType);
    const page = Math.max(1, params.page ?? 1);
    const pageSize = Math.min(50, Math.max(1, params.pageSize ?? 10));
    searchParams.set('page', String(page));
    searchParams.set('page_size', String(pageSize));

    const response = await fetch(`${API_BASE_URL}/jobs?${searchParams.toString()}`, { headers: authHeaders() });
    if (!response.ok) throw new Error(await errorMessage(response, 'Jobs could not be loaded.'));

    const body = await response.json().catch(() => null);
    const items = Array.isArray(body) ? body : body?.jobs ?? body?.results ?? body?.data ?? [];
    const jobs = Array.isArray(items) ? items.map(normalizeJob) : [];
    const total = Number(body?.total ?? body?.count ?? jobs.length);
    return {
      jobs,
      total,
      page,
      totalPages: Math.max(1, Math.ceil(total / pageSize)),
    };
  },

  async get(id: string): Promise<Job> {
    if (!id) throw new Error('A job ID is required.');
    const response = await fetch(`${API_BASE_URL}/jobs/${encodeURIComponent(id)}`, { headers: authHeaders() });
    if (!response.ok) throw new Error(await errorMessage(response, 'The job details could not be loaded.'));
    const body = await response.json().catch(() => ({}));
    return normalizeJob(body.job ?? body);
  },
};
