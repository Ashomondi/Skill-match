import { useCallback, useEffect, useRef, useState } from 'react';
import {
  ProfileUpdate,
  UserProfile,
  createEmptyProfile,
  formatWorkExperience,
  profileService,
} from '../services/profile';

export type ProfileMutationStatus = 'idle' | 'pending' | 'success' | 'error';

const messageFrom = (error: unknown, fallback: string): string =>
  error instanceof Error && error.message ? error.message : fallback;

const withOptimisticData = (current: UserProfile | null, data: ProfileUpdate): UserProfile => {
  const next: UserProfile = { ...(current ?? createEmptyProfile()), ...data };

  if (data.workExperience !== undefined) {
    next.experience = data.workExperience.map(formatWorkExperience).filter(Boolean);
  }

  return next;
};

export function useProfile() {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [mutationStatus, setMutationStatus] = useState<ProfileMutationStatus>('idle');
  const [mutationError, setMutationError] = useState<string | null>(null);

  const mountedRef = useRef(true);
  const profileRef = useRef<UserProfile | null>(null);
  const requestRef = useRef(0);

  const commitProfile = useCallback((next: UserProfile | null) => {
    profileRef.current = next;
    if (mountedRef.current) setProfile(next);
  }, []);

  const load = useCallback(async () => {
    const requestId = ++requestRef.current;

    setLoading(true);
    setError(null);

    try {
      const next = await profileService.getProfile();
      if (requestId === requestRef.current) commitProfile(next);
    } catch (loadError) {
      if (requestId === requestRef.current) {
        setError(messageFrom(loadError, 'Profile could not be loaded.'));
      }
    } finally {
      if (requestId === requestRef.current) setLoading(false);
    }
  }, [commitProfile]);

  useEffect(() => {
    mountedRef.current = true;
    void load();
    return () => {
      mountedRef.current = false;
    };
  }, [load]);

  /**
   * Applies `data` optimistically, then persists it. The previous profile is
   * restored if the request fails, and the returned promise rejects so callers
   * can surface the failure.
   */
  const update = useCallback(async (data: ProfileUpdate): Promise<UserProfile> => {
    const requestId = ++requestRef.current;
    const previous = profileRef.current;

    commitProfile(withOptimisticData(previous, data));
    setMutationStatus('pending');
    setMutationError(null);

    try {
      const next = await profileService.updateProfile(data);
      if (requestId === requestRef.current) {
        commitProfile(next);
        setMutationStatus('success');
      }
      return next;
    } catch (updateError) {
      if (requestId === requestRef.current) {
        commitProfile(previous);
        setMutationError(messageFrom(updateError, 'Profile could not be saved.'));
        setMutationStatus('error');
      }
      throw updateError;
    }
  }, [commitProfile]);

  const reload = useCallback(async () => {
    await load();
  }, [load]);

  const resetMutation = useCallback(() => {
    setMutationStatus('idle');
    setMutationError(null);
  }, []);

  const clearErrors = useCallback(() => {
    setError(null);
    setMutationError(null);
  }, []);

  return {
    profile,
    loading,
    error,
    update,
    reload,
    retry: reload,
    isUpdating: mutationStatus === 'pending',
    mutationStatus,
    mutationError,
    resetMutation,
    clearErrors,
  };
}
