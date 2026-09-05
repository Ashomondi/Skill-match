import { API_BASE_URL, authHeaders, getErrorMessage } from './api';

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
  remote?: boolean;
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

// External job sources may return HTML markup; decode to clean text preserving breaks.
export const formatJobDescription = (value?: string): string => {
  if (!value) return '';
  const withBreaks = value
    .replace(/<\s*br\s*\/?>/gi, '\n')
    .replace(/<\s*\/(p|div|li|h[1-6])\s*>/gi, '\n');
  const doc = new DOMParser().parseFromString(withBreaks, 'text/html');
  return (doc.body.textContent || '').replace(/\n\s*\n\s*\n+/g, '\n\n').trim();
};

const normalizeJob = (item: any): Job => ({
  id: String(item?.id ?? item?.job_id ?? ''),
  title: String(item?.title ?? item?.job_title ?? 'Untitled role'),
  company: String(item?.company ?? item?.company_name ?? 'Company not provided'),
  location: String(item?.location ?? 'Location not provided'),
  description: item?.description ?? item?.summary ?? '',
  salary: item?.salary ?? item?.salary_range ?? undefined,
  remote: Boolean(item?.remote),
  sourceUrl: item?.source_url ?? item?.sourceUrl ?? item?.apply_url ?? item?.url ?? undefined,
  postedAt: item?.created_at ?? item?.posted_at ?? item?.postedAt ?? undefined,
  matchScore:
    typeof item?.match_score === 'number'
      ? item.match_score
      : typeof item?.matchScore === 'number'
      ? item.matchScore
      : undefined,
  skills: Array.isArray(item?.skills) ? item.skills.map(String) : undefined,
  workType: item?.work_type ?? item?.workType ?? item?.employment_type ?? undefined,
  seniority: item?.seniority ?? item?.seniority_level ?? item?.experience_level ?? undefined,
});

export const jobsService = {
  async get(id: string): Promise<Job> {
    if (!id) throw new Error('A job ID is required.');
    const response = await fetch(`${API_BASE_URL}/jobs/${encodeURIComponent(id)}`, {
      headers: authHeaders(),
    });
    if (!response.ok) throw new Error(await getErrorMessage(response, 'The job details could not be loaded.'));
    const body = await response.json().catch(() => ({}));
    return normalizeJob(body?.data ?? body?.job ?? body);
  },

  async search(params: JobSearchParams = {}): Promise<JobSearchResult> {
    const searchParams = new URLSearchParams();
    if (params.query?.trim()) searchParams.set('q', params.query.trim());
    if (params.location?.trim()) searchParams.set('location', params.location.trim());

    // Note: seniority and work_type are not modeled by the backend search API.
    // They are omitted from query params to avoid sending unhandled parameters.

    const page = Math.max(1, params.page ?? 1);
    const pageSize = Math.min(50, Math.max(1, params.pageSize ?? 10));
    const offset = (page - 1) * pageSize;

    searchParams.set('limit', String(pageSize));
    searchParams.set('offset', String(offset));

    const response = await fetch(`${API_BASE_URL}/jobs/search?${searchParams.toString()}`, {
      headers: authHeaders(),
    });
    if (!response.ok) throw new Error(await getErrorMessage(response, 'Jobs could not be loaded.'));

    const body = await response.json().catch(() => null);
    const items = Array.isArray(body)
      ? body
      : body?.jobs ?? body?.results ?? body?.data?.jobs ?? body?.data ?? [];
    const jobs = Array.isArray(items) ? items.map(normalizeJob) : [];
    const total = Number(
      body?.pagination?.total ??
      body?.total ??
      body?.data?.pagination?.total ??
      body?.count ??
      jobs.length
    );

    return {
      jobs,
      total,
      page,
      totalPages: Math.max(1, Math.ceil(total / pageSize)),
    };
  },
};
