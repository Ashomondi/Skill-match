// frontend/src/services/resume.ts

import { setResumePresent } from './chat';
import { API_BASE_URL, authHeaders, getErrorMessage } from './api';

export type ResumeStatus = 'uploaded' | 'parsing' | 'parsed' | 'failed';

export interface Resume {
  id: string;
  name: string;
  uploadedAt: string;
  status: ResumeStatus;
  size?: number;
  failureReason?: string;
  filename?: string;
  url?: string;
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

/** Maps backend status values to canonical frontend states. */
export const normalizeStatus = (status?: string): ResumeStatus => {
  switch (String(status ?? '').trim().toLowerCase()) {
    case 'parsed':
    case 'ready':
    case 'active':
      return 'parsed';
    case 'failed':
    case 'error':
      return 'failed';
    case 'parsing':
    case 'processing':
      return 'parsing';
    case 'uploaded':
    default:
      return 'uploaded';
  }
};

const normalizeResume = (item: any): Resume => ({
  id: String(item?.id ?? item?.resume_id ?? ''),
  name: item?.name || item?.filename || item?.original_filename || 'Untitled resume',
  filename: item?.filename || item?.original_filename,
  uploadedAt: item?.uploadedAt || item?.created_at || item?.createdAt || new Date().toISOString(),
  status: normalizeStatus(item?.status),
  size: Number(item?.size ?? item?.file_size ?? item?.file_size_bytes) || undefined,
  failureReason: item?.failure_reason ?? item?.failureReason ?? item?.error,
  url: item?.url,
});

export const resumeService = {
  async list(): Promise<Resume[]> {
    const response = await fetch(`${API_BASE_URL}/resumes`, { headers: authHeaders() });
    if (!response.ok) throw new Error(await getErrorMessage(response, 'Unable to load your resumes.'));
    const body = await response.json().catch(() => []);
    const items = Array.isArray(body) ? body : body?.resumes ?? body?.data ?? [];
    const resumes = Array.isArray(items) ? items.map(normalizeResume) : [];
    setResumePresent(resumes.length > 0);
    return resumes;
  },

  async get(id: string): Promise<Resume> {
    if (!id) throw new Error('A resume ID is required.');
    const response = await fetch(`${API_BASE_URL}/resumes/${encodeURIComponent(id)}`, { headers: authHeaders() });
    if (!response.ok) throw new Error(await getErrorMessage(response, 'The resume could not be loaded.'));
    const body = await response.json().catch(() => ({}));
    return normalizeResume(body.resume ?? body.data ?? body);
  },

  /** Uploads (or replaces) a resume via POST /api/resumes. */
  async upload(
    file: File,
    optionsOrProgress?: UploadResumeOptions | ((progress: number) => void),
    maybeReplaceId?: string
  ): Promise<Resume> {
    const options: UploadResumeOptions =
      typeof optionsOrProgress === 'function'
        ? { onProgress: optionsOrProgress, replaceId: maybeReplaceId }
        : optionsOrProgress || {};

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
        let data: any = null;
        try {
          data = JSON.parse(xhr.responseText);
        } catch {
          /* non-JSON payload */
        }

        if (xhr.status >= 200 && xhr.status < 300) {
          setResumePresent(true);
          resolve(normalizeResume(data?.resume ?? data?.data ?? data));
        } else {
          const message =
            data?.error?.message ||
            data?.error ||
            data?.message ||
            `Upload failed with status ${xhr.status}`;
          reject(new Error(message));
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
    onProgress?: (progress: number) => void
  ): Promise<UploadResponse> {
    const resume = await this.upload(file, { onProgress });
    return {
      message: 'Upload successful',
      fileUrl: resume.url || '',
      resumeId: resume.id,
      status: resume.status,
    };
  },

  async deleteResume(id: string): Promise<void> {
    if (!id) throw new Error('A resume ID is required.');
    const response = await fetch(`${API_BASE_URL}/resumes/${encodeURIComponent(id)}`, {
      method: 'DELETE',
      headers: authHeaders(),
    });
    if (!response.ok) throw new Error(await getErrorMessage(response, 'Your resume could not be deleted.'));
    setResumePresent((await this.list()).length > 0);
  },

  async remove(id: string): Promise<void> {
    return this.deleteResume(id);
  },
};
