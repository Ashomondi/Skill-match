const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

export type ResumeStatus = 'processing' | 'ready' | 'failed';

export interface Resume {
  id: string;
  name: string;
  filename?: string;
  size?: number;
  status: ResumeStatus;
  uploadedAt: string;
  url?: string;
}

const authHeaders = (): Record<string, string> => {
  const token = localStorage.getItem('token');
  return token ? { Authorization: `Bearer ${token}` } : {};
};

const mapStatus = (status: string): ResumeStatus => {
  const value = String(status || 'uploaded').toLowerCase();
  if (value === 'parsed') return 'ready';
  if (value === 'failed') return 'failed';
  return 'processing';
};

const normalize = (item: any): Resume => ({
  id: String(item.id),
  name: item.name || item.filename || 'Untitled resume',
  filename: item.filename,
  size: item.size,
  status: mapStatus(item.status),
  uploadedAt: item.uploadedAt || item.created_at || new Date().toISOString(),
  url: item.url,
});

export const resumeService = {
  async list(): Promise<Resume[]> {
    const response = await fetch(`${API_BASE_URL}/resumes`, { headers: authHeaders() });
    const body = await response.json().catch(() => null);
    if (!response.ok) throw new Error(body?.error || body?.message || 'Resumes could not be loaded.');

    const items = Array.isArray(body) ? body : body?.resumes ?? [];
    return Array.isArray(items) ? items.map(normalize) : [];
  },

  async upload(file: File, onProgress?: (progress: number) => void, replaceId?: string): Promise<Resume> {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest();
      const formData = new FormData();
      formData.append('resume', file);
      if (replaceId) formData.append('replaceId', replaceId);

      xhr.open('POST', `${API_BASE_URL}/resumes`, true);

      const token = localStorage.getItem('token');
      if (token) xhr.setRequestHeader('Authorization', `Bearer ${token}`);

      if (xhr.upload && onProgress) {
        xhr.upload.onprogress = (event) => {
          if (event.lengthComputable) {
            onProgress(Math.round((event.loaded * 100) / event.total));
          }
        };
      }

      xhr.onload = () => {
        let data: any = null;
        try {
          data = JSON.parse(xhr.responseText);
        } catch {
          /* non-JSON error body */
        }
        if (xhr.status >= 200 && xhr.status < 300) {
          resolve(normalize(data));
        } else {
          reject(new Error(data?.error || data?.message || `Upload failed with status ${xhr.status}`));
        }
      };

      xhr.onerror = () => reject(new Error('Network error occurred during file upload'));
      xhr.send(formData);
    });
  },

  async remove(id: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/resumes/${id}`, { method: 'DELETE', headers: authHeaders() });
    if (!response.ok) {
      const body = await response.json().catch(() => null);
      throw new Error(body?.error || body?.message || 'The resume could not be deleted.');
    }
  },
};
