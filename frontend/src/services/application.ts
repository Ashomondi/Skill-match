<<<<<<< HEAD
import api from './api';

export interface Application {
  id: string;
  jobId: string;
  jobTitle: string;
  company: string;
  status: 'Applied' | 'Phone Screening' | 'Interviewing' | 'Offered' | 'Rejected' | 'Withdrawn';
  appliedDate: string;
  notes?: string;
  location?: string;
}

export interface UpdateApplicationDTO {
  status?: Application['status'];
  notes?: string;
}

export const applicationService = {
  async getApplications(): Promise<Application[]> {
    const response = await api.get('/applications');
    return response.data;
  },

  async updateApplication(id: string, data: UpdateApplicationDTO): Promise<Application> {
    const response = await api.put(`/applications/${id}`, data);
    return response.data;
  },

  async deleteApplication(id: string): Promise<void> {
    await api.delete(`/applications/${id}`);
=======
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

export const applicationService = {
  async list(): Promise<Application[]> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_BASE_URL}/applications`, { headers: token ? { Authorization: `Bearer ${token}` } : {} });
    const body = await response.json().catch(() => null);
    if (!response.ok) throw new Error(body?.error || body?.message || 'Applications could not be loaded.');
    const items = Array.isArray(body) ? body : body?.applications ?? body?.data ?? [];
    return Array.isArray(items) ? items.map(normalize) : [];
>>>>>>> main
  },
};
