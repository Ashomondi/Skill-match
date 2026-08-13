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
  },
};
