import { useCallback, useEffect, useMemo, useState } from 'react';
import { resumeService, Resume } from '../services/resume';

const ACTIVE_CV_KEY = 'skillmatch-active-cv';

const readActiveId = (): string => {
  try {
    return localStorage.getItem(ACTIVE_CV_KEY) || '';
  } catch {
    return '';
  }
};

const persistActiveId = (id: string) => {
  try {
    localStorage.setItem(ACTIVE_CV_KEY, id);
  } catch {
    // storage unavailable; selection just won't persist across reloads
  }
};

export interface ResumeContextValue {
  resumes: Resume[];
  loading: boolean;
  isUploading: boolean;
  error: string | null;
  active: Resume | null;
  refresh: () => Promise<Resume[]>;
  select: (id: string) => void;
  upload: (file: File, replaceId?: string) => Promise<Resume | null>;
}

export function useResumeContext(): ResumeContextValue {
  const [resumes, setResumes] = useState<Resume[]>([]);
  const [loading, setLoading] = useState(true);
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [activeId, setActiveId] = useState<string>(readActiveId);

  const refresh = useCallback(async (): Promise<Resume[]> => {
    setError(null);
    try {
      const list = await resumeService.list();
      setResumes(list);
      return list;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Your resumes could not be loaded.');
      return [];
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { void refresh(); }, [refresh]);

  const select = useCallback((id: string) => {
    setActiveId(id);
    persistActiveId(id);
  }, []);

  const upload = useCallback(async (file: File, replaceId?: string): Promise<Resume | null> => {
    setError(null);
    setIsUploading(true);
    try {
      const created = await resumeService.upload(file, undefined, replaceId);
      if (created) {
        await refresh();
        select(created.id);
        return created;
      }
      return null;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Your resume could not be uploaded.');
      return null;
    } finally {
      setIsUploading(false);
    }
  }, [refresh, select]);

  const active = useMemo<Resume | null>(() => {
    const byId = resumes.find((resume) => resume.id === activeId);
    if (byId) return byId;
    const firstReady = resumes.find((resume) => resume.status === 'ready');
    if (firstReady) return firstReady;
    return resumes[0] ?? null;
  }, [resumes, activeId]);

  return { resumes, loading, isUploading, error, active, refresh, select, upload };
}
