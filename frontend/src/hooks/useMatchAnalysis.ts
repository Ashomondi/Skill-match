import { useCallback, useEffect, useRef, useState } from 'react';
import {
  AtsReadinessReport,
  MatchAnalysisPayload,
  MatchReport,
  matchingService,
} from '../services/matching';

export type MatchAnalysisStatus = 'idle' | 'loading' | 'success' | 'error';

interface NormalizedPayload {
  jobId?: string;
  jobDescription?: string;
}

const messageFrom = (error: unknown, fallback: string): string =>
  error instanceof Error && error.message ? error.message : fallback;

const normalizePayload = (payload: MatchAnalysisPayload): NormalizedPayload => ({
  jobId: (payload.jobId ?? payload.job_id ?? '').trim() || undefined,
  jobDescription: (payload.jobDescription ?? payload.job_description ?? '').trim() || undefined,
});

// Small stable hash so long job descriptions do not become cache keys.
const hashText = (value: string): string => {
  let hash = 2166136261;
  for (let index = 0; index < value.length; index += 1) {
    hash ^= value.charCodeAt(index);
    hash = Math.imul(hash, 16777619);
  }
  return (hash >>> 0).toString(36);
};

const matchCacheKey = (cvId: string, payload: NormalizedPayload): string =>
  payload.jobId
    ? `job:${cvId}:${payload.jobId}`
    : `desc:${cvId}:${payload.jobDescription?.length ?? 0}:${hashText(payload.jobDescription ?? '')}`;

const atsCacheKey = (cvId: string): string => `cv:${cvId}`;

export interface UseMatchAnalysisResult {
  matchReport: MatchReport | null;
  report: MatchReport | null;
  match: MatchReport | null;
  atsReport: AtsReadinessReport | null;
  ats: AtsReadinessReport | null;
  readiness: AtsReadinessReport | null;
  loading: boolean;
  isAnalyzing: boolean;
  isCheckingAts: boolean;
  atsLoading: boolean;
  error: string | null;
  matchError: string | null;
  atsError: string | null;
  matchStatus: MatchAnalysisStatus;
  atsStatus: MatchAnalysisStatus;
  analyzeMatch: (
    cvIdOrPayload: string | MatchAnalysisPayload,
    payload?: MatchAnalysisPayload,
  ) => Promise<MatchReport>;
  analyze: (
    cvIdOrPayload: string | MatchAnalysisPayload,
    payload?: MatchAnalysisPayload,
  ) => Promise<MatchReport>;
  checkAtsReadiness: (cvId?: string) => Promise<AtsReadinessReport>;
  checkAts: (cvId?: string) => Promise<AtsReadinessReport>;
  reset: () => void;
  clearCache: () => void;
}

/**
 * Runs match and ATS analysis with per-hook caching.
 *
 * A default CV id can be supplied once, or passed to each call. Identical
 * requests are served from the in-memory cache; later requests win over
 * earlier in-flight ones so stale responses never overwrite newer state.
 */
export function useMatchAnalysis(defaultCvId?: string): UseMatchAnalysisResult {
  const [matchReport, setMatchReport] = useState<MatchReport | null>(null);
  const [atsReport, setAtsReport] = useState<AtsReadinessReport | null>(null);
  const [matchStatus, setMatchStatus] = useState<MatchAnalysisStatus>('idle');
  const [atsStatus, setAtsStatus] = useState<MatchAnalysisStatus>('idle');
  const [matchError, setMatchError] = useState<string | null>(null);
  const [atsError, setAtsError] = useState<string | null>(null);

  const mountedRef = useRef(true);
  const matchRequestRef = useRef(0);
  const atsRequestRef = useRef(0);
  const matchCacheRef = useRef(new Map<string, MatchReport>());
  const atsCacheRef = useRef(new Map<string, AtsReadinessReport>());

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
      // Invalidate any responses that arrive after unmount.
      matchRequestRef.current += 1;
      atsRequestRef.current += 1;
    };
  }, []);

  const analyzeMatch = useCallback(
    async (
      cvIdOrPayload: string | MatchAnalysisPayload,
      maybePayload?: MatchAnalysisPayload,
    ): Promise<MatchReport> => {
      const cvId = (typeof cvIdOrPayload === 'string' ? cvIdOrPayload : defaultCvId ?? '').trim();
      const payload = normalizePayload(
        typeof cvIdOrPayload === 'string' ? maybePayload ?? {} : cvIdOrPayload,
      );
      const requestId = ++matchRequestRef.current;

      if (!cvId) {
        const message = 'A CV id is required to analyze a match.';
        if (mountedRef.current) {
          setMatchError(message);
          setMatchStatus('error');
        }
        throw new Error(message);
      }

      if (!payload.jobId && !payload.jobDescription) {
        const message = 'Provide a job id or a job description to analyze.';
        if (mountedRef.current) {
          setMatchError(message);
          setMatchStatus('error');
        }
        throw new Error(message);
      }

      const cacheKey = matchCacheKey(cvId, payload);
      const cached = matchCacheRef.current.get(cacheKey);
      if (cached) {
        if (mountedRef.current) {
          setMatchReport(cached);
          setMatchError(null);
          setMatchStatus('success');
        }
        return cached;
      }

      const requestPayload: MatchAnalysisPayload = {};
      if (payload.jobId) requestPayload.jobId = payload.jobId;
      if (payload.jobDescription) requestPayload.jobDescription = payload.jobDescription;

      if (mountedRef.current) {
        setMatchReport(null);
        setMatchStatus('loading');
        setMatchError(null);
      }

      try {
        const report = await matchingService.analyzeMatch(cvId, requestPayload);

        if (mountedRef.current && requestId === matchRequestRef.current) {
          matchCacheRef.current.set(cacheKey, report);
          setMatchReport(report);
          setMatchStatus('success');
        }

        return report;
      } catch (error) {
        if (mountedRef.current && requestId === matchRequestRef.current) {
          setMatchError(messageFrom(error, 'The match analysis could not be completed.'));
          setMatchStatus('error');
        }
        throw error;
      }
    },
    [defaultCvId],
  );

  const checkAtsReadiness = useCallback(
    async (cvIdArg?: string): Promise<AtsReadinessReport> => {
      const cvId = (cvIdArg ?? defaultCvId ?? '').trim();
      const requestId = ++atsRequestRef.current;

      if (!cvId) {
        const message = 'A CV id is required for the ATS readiness check.';
        if (mountedRef.current) {
          setAtsError(message);
          setAtsStatus('error');
        }
        throw new Error(message);
      }

      const cacheKey = atsCacheKey(cvId);
      const cached = atsCacheRef.current.get(cacheKey);
      if (cached) {
        if (mountedRef.current) {
          setAtsReport(cached);
          setAtsError(null);
          setAtsStatus('success');
        }
        return cached;
      }

      if (mountedRef.current) {
        setAtsReport(null);
        setAtsStatus('loading');
        setAtsError(null);
      }

      try {
        const report = await matchingService.checkAtsReadiness(cvId);

        if (mountedRef.current && requestId === atsRequestRef.current) {
          atsCacheRef.current.set(cacheKey, report);
          setAtsReport(report);
          setAtsStatus('success');
        }

        return report;
      } catch (error) {
        if (mountedRef.current && requestId === atsRequestRef.current) {
          setAtsError(messageFrom(error, 'The ATS readiness check could not be completed.'));
          setAtsStatus('error');
        }
        throw error;
      }
    },
    [defaultCvId],
  );

  const reset = useCallback(() => {
    matchRequestRef.current += 1;
    atsRequestRef.current += 1;
    setMatchReport(null);
    setAtsReport(null);
    setMatchStatus('idle');
    setAtsStatus('idle');
    setMatchError(null);
    setAtsError(null);
  }, []);

  const clearCache = useCallback(() => {
    matchCacheRef.current.clear();
    atsCacheRef.current.clear();
  }, []);

  const isAnalyzing = matchStatus === 'loading';
  const isCheckingAts = atsStatus === 'loading';

  return {
    matchReport,
    report: matchReport,
    match: matchReport,
    atsReport,
    ats: atsReport,
    readiness: atsReport,
    loading: isAnalyzing || isCheckingAts,
    isAnalyzing,
    isCheckingAts,
    atsLoading: isCheckingAts,
    error: matchError ?? atsError,
    matchError,
    atsError,
    matchStatus,
    atsStatus,
    analyzeMatch,
    analyze: analyzeMatch,
    checkAtsReadiness,
    checkAts: checkAtsReadiness,
    reset,
    clearCache,
  };
}
