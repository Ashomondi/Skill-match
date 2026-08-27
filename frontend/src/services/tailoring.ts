const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

export const tailoringService = {
  async generate(input: { resumeId: string; jobTitle: string; company: string; jobDescription?: string; currentContent?: string }): Promise<string> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_BASE_URL}/tailor`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...(token ? { Authorization: `Bearer ${token}` } : {}) },
      body: JSON.stringify({ resume_id: input.resumeId, job_title: input.jobTitle, company: input.company, job_description: input.jobDescription || '', current_content: input.currentContent || '' }),
    });
    const body = await response.json().catch(() => null);
    if (!response.ok) throw new Error(body?.error?.message || body?.error || 'CV tailoring failed.');
    return String(body?.content || '');
  },
};
