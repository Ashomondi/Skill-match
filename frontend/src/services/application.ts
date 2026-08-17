const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

export type ApplicationStatus = 'applied' | 'screening' | 'interview' | 'offer' | 'rejected' | 'withdrawn';
export interface Application { id: string; jobId?: string; role: string; company: string; status: ApplicationStatus; appliedAt: string; resumeUrl?: string; }

const statuses: ApplicationStatus[] = ['applied', 'screening', 'interview', 'offer', 'rejected', 'withdrawn'];
const normalize = (item: any): Application => {
  const job = item.job ?? item;
  const rawStatus = String(item.status ?? 'applied').toLowerCase() as ApplicationStatus;
  return {
    id: String(item.id ?? item.application_id),
    jobId: item.job_id ? String(item.job_id) : job.id ? String(job.id) : undefined,
    role: job.title ?? item.role ?? item.job_title ?? 'Untitled role',
    company: job.company ?? item.company ?? item.company_name ?? 'Unknown company',
    status: statuses.includes(rawStatus) ? rawStatus : 'applied',
    appliedAt: item.applied_at ?? item.created_at ?? item.updated_at ?? '',
    resumeUrl: item.resume_url ?? item.tailored_resume_url,
  };
};

const headers = (): Record<string, string> => {
  const token = localStorage.getItem('token');
  return token ? { Authorization: `Bearer ${token}` } : {};
};

const errorMessage = async (response: Response, fallback: string) => {
  const body = await response.json().catch(() => ({}));
  return body.error || body.message || fallback;
};

export const applicationService = {
  async list(): Promise<Application[]> {
    const response = await fetch(`${API_BASE_URL}/applications`, { headers: headers() });
    if (!response.ok) throw new Error(await errorMessage(response, 'Applications could not be loaded.'));
    const body = await response.json().catch(() => null);
    const items = Array.isArray(body) ? body : body?.applications ?? body?.data ?? [];
    return Array.isArray(items) ? items.map(normalize) : [];
  },

  async apply(jobId: string): Promise<Application> {
    const response = await fetch(`${API_BASE_URL}/applications`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...headers() },
      body: JSON.stringify({ job_id: jobId }),
    });
    if (!response.ok) throw new Error(await errorMessage(response, 'Your application could not be submitted.'));
    const body = await response.json().catch(() => ({}));
    return normalize(body.application ?? body);
  },
};
