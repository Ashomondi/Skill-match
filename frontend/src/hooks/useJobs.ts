import { useCallback, useEffect, useState } from 'react';
import { Job, JobSearchParams, jobsService } from '../services/jobs';

export function useJobs(params: JobSearchParams) {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { query = '', location = '', seniority = '', workType = '' } = params;

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await jobsService.search({ query, location, seniority, workType });
      setJobs(result.jobs);
      setTotal(result.total);
    } catch (loadError) {
      setJobs([]);
      setTotal(0);
      setError(loadError instanceof Error ? loadError.message : 'Jobs could not be loaded.');
    } finally {
      setLoading(false);
    }
  }, [query, location, seniority, workType]);

  useEffect(() => {
    const timeout = window.setTimeout(() => { void load(); }, query ? 350 : 0);
    return () => window.clearTimeout(timeout);
  }, [load, query]);

  return { jobs, total, loading, error, retry: load };
}
