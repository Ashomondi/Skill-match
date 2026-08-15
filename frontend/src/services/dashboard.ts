const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

export type ApplicationStatus = 'saved' | 'applied' | 'screening' | 'interview' | 'offer' | 'rejected' | 'withdrawn';
export interface RecentApplication { id: string; role: string; company: string; status: ApplicationStatus; updatedAt: string; }
export interface DashboardData { savedJobs: number; totalApplications: number; byStatus: Record<ApplicationStatus, number>; recentApplications: RecentApplication[]; }
const statuses: ApplicationStatus[] = ['saved', 'applied', 'screening', 'interview', 'offer', 'rejected', 'withdrawn'];
const emptyCounts = () => Object.fromEntries(statuses.map((status) => [status, 0])) as Record<ApplicationStatus, number>;
const request = async (path: string) => { const token = localStorage.getItem('token'); const response = await fetch(`${API_BASE_URL}${path}`, { headers: token ? { Authorization: `Bearer ${token}` } : {} }); const body = await response.json().catch(() => null); if (!response.ok) throw new Error(body?.error || body?.message || 'Unable to load dashboard data.'); return body; };
const listFrom = (body: any): any[] => Array.isArray(body) ? body : body?.applications || body?.data || [];

export const dashboardService = {
  async getDashboard(): Promise<DashboardData> {
    const [applicationsResult, savedResult] = await Promise.allSettled([request('/applications'), request('/saved-jobs')]);
    if (applicationsResult.status === 'rejected' && savedResult.status === 'rejected') throw applicationsResult.reason;
    const applications = applicationsResult.status === 'fulfilled' ? listFrom(applicationsResult.value) : [];
    const savedBody = savedResult.status === 'fulfilled' ? savedResult.value : [];
    const savedJobs = Array.isArray(savedBody) ? savedBody.length : Number(savedBody?.total ?? savedBody?.count ?? savedBody?.saved_jobs?.length ?? 0);
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
