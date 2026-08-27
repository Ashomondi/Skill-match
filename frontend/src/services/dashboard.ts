const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

export type ApplicationStatus = 'saved' | 'applied' | 'screening' | 'interview' | 'offer' | 'rejected' | 'withdrawn';
export interface RecentApplication { id: string; role: string; company: string; status: ApplicationStatus; updatedAt: string; }
export interface DashboardData { savedJobs: number; totalApplications: number; byStatus: Record<ApplicationStatus, number>; recentApplications: RecentApplication[]; }
const statuses: ApplicationStatus[] = ['saved', 'applied', 'screening', 'interview', 'offer', 'rejected', 'withdrawn'];
const emptyCounts = () => Object.fromEntries(statuses.map((status) => [status, 0])) as Record<ApplicationStatus, number>;
const request = async (path: string) => { const token = localStorage.getItem('token'); const response = await fetch(`${API_BASE_URL}${path}`, { headers: token ? { Authorization: `Bearer ${token}` } : {} }); const body = await response.json().catch(() => null); if (!response.ok) throw new Error(body?.error?.message || body?.error || body?.message || 'Unable to load dashboard data.'); return body; };
// The backend wraps success payloads in { data: ... }.
const unwrap = (body: any): any => (body && typeof body === 'object' && 'data' in body ? body.data : body);
const listFrom = (body: any): any[] => { const inner = unwrap(body); return Array.isArray(inner) ? inner : inner?.applications || inner?.data || []; };

export const dashboardService = {
  async getDashboard(): Promise<DashboardData> {
    const [applicationsBody, savedBody] = await Promise.all([request('/applications'), request('/saved-jobs')]);
    const applications = listFrom(applicationsBody);
    const savedList = listFrom(savedBody);
    const savedJobs = savedList.length;
    const byStatus = emptyCounts();
    const normalized = applications.map((application: any) => {
      const status = String(application.status || 'saved').toLowerCase() as ApplicationStatus;
      if (statuses.includes(status)) byStatus[status] += 1;
      return { id: String(application.id), role: application.role || application.job?.title || application.job_title || 'Untitled role', company: application.company || application.job?.company || application.company_name || 'Unknown company', status: statuses.includes(status) ? status : 'saved', updatedAt: application.updated_at || application.applied_at || application.created_at || '' } satisfies RecentApplication;
    });
    normalized.sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime());
    return { savedJobs, totalApplications: applications.length, byStatus, recentApplications: normalized.slice(0, 5) };
  },
};
