import { useCallback, useEffect, useState } from 'react';
import { Recommendation, recommendationsService } from '../services/recommendations';

export function useRecommendations() {
  const [recommendations, setRecommendations] = useState<Recommendation[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setRecommendations(await recommendationsService.list());
    } catch (loadError) {
      setRecommendations([]);
      setError(loadError instanceof Error ? loadError.message : 'Recommendations could not be loaded.');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { void load(); }, [load]);
  return { recommendations, loading, error, retry: load };
}
