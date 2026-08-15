import { Job } from './jobs';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

export interface Recommendation extends Job {
  relevanceScore?: number;
  relevanceLabel?: string;
  reasons: string[];
}

const normalizeRecommendation = (item: any): Recommendation => {
  const job = item.job ?? item;
  const reasons = item.reasons ?? item.match_reasons ?? item.relevance_reasons ?? job.reasons ?? [];
  const score = Number(item.relevance_score ?? item.score ?? item.match_score ?? job.match_score) || undefined;
  return {
    id: String(job.id ?? job.job_id ?? item.job_id),
    title: job.title ?? job.job_title ?? 'Untitled role',
    company: job.company ?? job.company_name ?? 'Company not provided',
    location: job.location ?? 'Location not provided',
    workType: job.work_type ?? job.workType ?? job.employment_type,
    seniority: job.seniority ?? job.seniority_level ?? job.experience_level,
    salary: job.salary ?? job.salary_range,
    postedAt: job.posted_at ?? job.created_at,
    matchScore: score,
    relevanceScore: score,
    relevanceLabel: item.relevance ?? item.relevance_label ?? item.match_label,
    reasons: Array.isArray(reasons) ? reasons.map(String).filter(Boolean) : reasons ? [String(reasons)] : [],
  };
};

export const recommendationsService = {
  async list(): Promise<Recommendation[]> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_BASE_URL}/recommendations`, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    });
    const body = await response.json().catch(() => null);
    if (!response.ok) throw new Error(body?.error || body?.message || 'Recommendations could not be loaded.');
    const items = Array.isArray(body) ? body : body?.recommendations ?? body?.jobs ?? body?.data ?? [];
    return Array.isArray(items) ? items.map(normalizeRecommendation) : [];
  },
};
