// frontend/src/services/resume.ts

import { setResumePresent } from './chat';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

export type ResumeStatus = 'processing' | 'active' | 'failed';

export interface Resume {
  id: string;
  name: string;
  uploadedAt: string;
  status: ResumeStatus;
  size?: number;
  failureReason?: string;
}

export interface UploadResponse {
  message: string;
  fileUrl: string;
  resumeId: string;
  status?: string;
}

export interface UploadResumeOptions {
  replaceId?: string;
  onProgress?: (progress: number) => void;
}

/** Maps backend status values to the three UI-facing states. */
export const normalizeStatus = (status?: string): ResumeStatus => {
  switch (String(status ?? '').toLowerCase()) {
    case 'parsed':
    case 'active':
    case 'ready':
      return 'active';
    case 'failed':
      return 'failed';
    case 'uploaded':
    case 'parsing':
    case 'processing':
    default:
      return 'processing';
  }
};

const authHeaders = (): Record<string, string> => {
  const token = localStorage.getItem('token');
  return token ? { Authorization: `Bearer ${token}` } : {};
};

const errorMessage = async (response: Response, fallback: string) => {
  const body = await response.json().catch(() => ({}));
  return body.error || body.message || fallback;
};

const normalizeResume = (item: any): Resume => ({
  id: String(item.id ?? item.resume_id),
  name: item.name || item.filename || item.original_filename || 'Untitled resume',
  uploadedAt: item.uploadedAt || item.created_at || item.createdAt || new Date().toISOString(),
  status: normalizeStatus(item.status),
  size: Number(item.size ?? item.file_size ?? item.file_size_bytes) || undefined,
  failureReason: item.failure_reason ?? item.error,
});

export const resumeService = {
  async list(): Promise<Resume[]> {
    const response = await fetch(`${API_BASE_URL}/resumes`, { headers: authHeaders() });
    if (!response.ok) throw new Error(await errorMessage(response, 'Unable to load your resumes.'));
    const body = await response.json().catch(() => []);
    const items = Array.isArray(body) ? body : body.resumes ?? body.data ?? [];
    const resumes = Array.isArray(items) ? items.map(normalizeResume) : [];
    setResumePresent(resumes.length > 0);
    return resumes;
  },

  async get(id: string): Promise<Resume> {
    if (!id) throw new Error('A resume ID is required.');
    const response = await fetch(`${API_BASE_URL}/resumes/${encodeURIComponent(id)}`, { headers: authHeaders() });
    if (!response.ok) throw new Error(await errorMessage(response, 'The resume could not be loaded.'));
    const body = await response.json().catch(() => ({}));
    return normalizeResume(body.resume ?? body);
  },

  /** Uploads (or replaces) a resume via POST /api/resumes. */
  async upload(file: File, options: UploadResumeOptions = {}): Promise<Resume> {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest();
      const formData = new FormData();
      formData.append('resume', file);
      if (options.replaceId) formData.append('replaceId', options.replaceId);

      const token = localStorage.getItem('token');

      xhr.open('POST', `${API_BASE_URL}/resumes`, true);
      if (token) xhr.setRequestHeader('Authorization', `Bearer ${token}`);

      if (xhr.upload && options.onProgress) {
        xhr.upload.onprogress = (event) => {
          if (event.lengthComputable) options.onProgress?.(Math.round((event.loaded * 100) / event.total));
        };
      }

      xhr.onload = () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            const data = JSON.parse(xhr.responseText);
            setResumePresent(true);
            resolve(normalizeResume(data.resume ?? data));
          } catch (e) {
            reject(new Error('Invalid JSON response from server'));
          }
        } else {
          try {
            const errData = JSON.parse(xhr.responseText);
            reject(new Error(errData.message || 'Failed to upload resume'));
          } catch (e) {
            reject(new Error(`Upload failed with status ${xhr.status}`));
          }
        }
      };

      xhr.onerror = () => {
        reject(new Error('Network error occurred during file upload'));
      };

      xhr.send(formData);
    });
  },

  async uploadCV(
    file: File,
    onProgress?: (progress: number) => void,
  ): Promise<UploadResponse> {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest();
      const formData = new FormData();
      formData.append('cv', file);

      const token = localStorage.getItem('token');

      xhr.open('POST', `${API_BASE_URL}/resumes/upload`, true);

      if (token) {
        xhr.setRequestHeader('Authorization', `Bearer ${token}`);
      }

      if (xhr.upload && onProgress) {
        xhr.upload.onprogress = (event) => {
          if (event.lengthComputable) {
            onProgress(Math.round((event.loaded * 100) / event.total));
          }
        };
      }

      xhr.onload = () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            const data = JSON.parse(xhr.responseText);
            setResumePresent(true);
            resolve(data);
          } catch (e) {
            reject(new Error('Invalid JSON response from server'));
          }
        } else {
          try {
            const errData = JSON.parse(xhr.responseText);
            reject(new Error(errData.message || 'Failed to upload CV'));
          } catch (e) {
            reject(new Error(`Upload failed with status ${xhr.status}`));
          }
        }
      };

      xhr.onerror = () => {
        reject(new Error('Network error occurred during file upload'));
      };

      xhr.send(formData);
    });
  },

  async deleteResume(id: string): Promise<void> {
    if (!id) throw new Error('A resume ID is required.');
    const response = await fetch(`${API_BASE_URL}/resumes/${encodeURIComponent(id)}`, {
      method: 'DELETE',
      headers: authHeaders(),
    });
    if (!response.ok) throw new Error(await errorMessage(response, 'Your resume could not be deleted.'));
    setResumePresent((await this.list()).length > 0);
  },
};
