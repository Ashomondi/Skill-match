import { useCallback, useEffect, useState } from 'react';
import { Job, JobSearchParams, jobsService } from '../services/jobs';

export function useJobs(params: JobSearchParams) {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(1);
  const [page, setPage] = useState(Math.max(1, params.page ?? 1));
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { query = '', location = '', seniority = '', workType = '', pageSize = 10 } = params;

  const load = useCallback(async (targetPage: number = page) => {
    setLoading(true);
    setError(null);
    try {
      const result = await jobsService.search({ query, location, seniority, workType, page: targetPage, pageSize });
      setJobs(result.jobs);
      setTotal(result.total);
      setTotalPages(result.totalPages);
      setPage(result.page);
    } catch (loadError) {
      setJobs([]);
      setTotal(0);
      setTotalPages(1);
      setError(loadError instanceof Error ? loadError.message : 'Jobs could not be loaded.');
    } finally {
      setLoading(false);
    }
  }, [query, location, seniority, workType, page, pageSize]);

  useEffect(() => {
    setPage(Math.max(1, params.page ?? 1));
  }, [params.page]);

  useEffect(() => {
    const timeout = window.setTimeout(() => { void load(); }, query ? 350 : 0);
    return () => window.clearTimeout(timeout);
  }, [load, query]);

  const goToPage = useCallback((next: number) => {
    const clamped = Math.min(Math.max(1, next), Math.max(1, totalPages));
    if (clamped !== page) void load(clamped);
  }, [load, page, totalPages]);

  return { jobs, total, totalPages, page, goToPage, loading, error, retry: load };
}
