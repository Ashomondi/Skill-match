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

const requestError = async (response: Response, fallback: string) => { const body = await response.json().catch(() => null); return body?.error?.message || body?.error || body?.message || fallback; };

export const applicationService = {
  async create(jobId: string): Promise<Application> {
    const token = localStorage.getItem("token");
    const response = await fetch(API_BASE_URL + "/applications", { method: "POST", headers: { "Content-Type": "application/json", ...(token ? { Authorization: "Bearer " + token } : {}) }, body: JSON.stringify({ job_id: jobId }) });
    if (!response.ok) throw new Error(await requestError(response, "Application could not be submitted."));
    return normalize(await response.json());
  },

  async updateStatus(id: string, status: ApplicationStatus): Promise<Application> {
    const token = localStorage.getItem("token");
    const response = await fetch(API_BASE_URL + "/applications/" + encodeURIComponent(id) + "/status", { method: "PATCH", headers: { "Content-Type": "application/json", ...(token ? { Authorization: "Bearer " + token } : {}) }, body: JSON.stringify({ status }) });
    if (!response.ok) throw new Error(await requestError(response, "Application status could not be updated."));
    return normalize(await response.json());
  },
  async list(): Promise<Application[]> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_BASE_URL}/applications`, { headers: token ? { Authorization: `Bearer ${token}` } : {} });
    const body = await response.json().catch(() => null);
    if (!response.ok) throw new Error(body?.error || body?.message || 'Applications could not be loaded.');
    const items = Array.isArray(body) ? body : body?.applications ?? body?.data ?? [];
    return Array.isArray(items) ? items.map(normalize) : [];
  },
};
