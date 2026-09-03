import { useCallback, useEffect, useState } from 'react';
import { Profile, ProfileInput, profileService } from '../services/profile';

export function useProfile() {
  const [profile, setProfile] = useState<Profile | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [saveError, setSaveError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setProfile(await profileService.get());
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : 'Profile could not be loaded.');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { void load(); }, [load]);

  const save = useCallback(async (input: ProfileInput): Promise<Profile> => {
    setSaving(true);
    setSaveError(null);
    try {
      const updated = await profileService.update(input);
      setProfile(updated);
      return updated;
    } catch (saveErr) {
      setSaveError(saveErr instanceof Error ? saveErr.message : 'Profile could not be saved.');
      throw saveErr;
    } finally {
      setSaving(false);
    }
  }, []);

  return {
    profile,
    loading,
    saving,
    error,
    saveError,
    load,
    save,
    clearSaveError: () => setSaveError(null),
  };
}
