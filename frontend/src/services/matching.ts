import { authService } from './auth';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

export interface MatchAnalysisPayload {
  jobId?: string;
  jobDescription?: string;
  /** Snake-case aliases accepted for convenience. */
  job_id?: string;
  job_description?: string;
}

export interface CategoryScores {
  skills: number;
  experience: number;
  education: number;
  keywords: number;
}

/** Mirrors `models.MatchingReport` from the backend (0–100 scores). */
export interface MatchReport {
  overallScore: number;
  matchedSkills: string[];
  missingMandatorySkills: string[];
  missingPreferredSkills: string[];
  keywordSuggestions: string[];
  categoryScores: CategoryScores;
  jobId?: string;
  cvId?: string;
  jobTitle?: string;
  company?: string;
  generatedAt?: string;
  disclaimer?: string;
}

export type AtsCheckStatus = 'pass' | 'warning' | 'fail';

export interface AtsCheck {
  id: string;
  label: string;
  status: AtsCheckStatus;
  message?: string;
  tip?: string;
}

export interface AtsSummary {
  passed: number;
  warnings: number;
  failed: number;
}

/** Structural/readability report from the ATS analyzer (not job suitability). */
export interface AtsReadinessReport {
  score: number;
  checks: AtsCheck[];
  tips: string[];
  disclaimer: string;
  summary: AtsSummary;
}

export type MatchCategoryScores = CategoryScores;
export type AtsReport = AtsReadinessReport;

type JsonRecord = Record<string, unknown>;

const isRecord = (value: unknown): value is JsonRecord =>
  typeof value === 'object' && value !== null && !Array.isArray(value);

const readString = (source: JsonRecord, ...keys: string[]): string | undefined => {
  for (const key of keys) {
    const value = source[key];
    if (typeof value === 'string' && value.trim()) return value.trim();
    if (typeof value === 'number') return String(value);
  }
  return undefined;
};

const toNumber = (value: unknown, fallback = 0): number => {
  const number = Number(value);
  return Number.isFinite(number) ? number : fallback;
};

const clampScore = (value: unknown): number => Math.max(0, Math.min(100, toNumber(value)));

const uniqueStrings = (value: unknown): string[] => {
  if (!Array.isArray(value)) return [];

  const seen = new Set<string>();
  const result: string[] = [];

  for (const item of value) {
    const text = typeof item === 'string'
      ? item
      : isRecord(item)
        ? readString(item, 'name', 'label', 'value', 'text', 'keyword', 'skill')
        : undefined;

    const trimmed = text?.trim();
    if (!trimmed) continue;

    const key = trimmed.toLowerCase();
    if (seen.has(key)) continue;

    seen.add(key);
    result.push(trimmed);
  }

  return result;
};

const authHeaders = (): Record<string, string> => {
  const token = authService.getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
};

const errorMessage = (body: any, fallback: string): string =>
  body?.error?.message || (typeof body?.error === 'string' ? body.error : '') || body?.message || fallback;

const unwrap = (body: unknown): unknown => {
  if (!isRecord(body)) return body;
  const data = body.data ?? body.report ?? body.result;
  return data ?? body;
};

const normalizeAtsStatus = (value: unknown): AtsCheckStatus => {
  const raw = String(value ?? '').trim().toLowerCase();
  if (['pass', 'passed', 'ok', 'success', 'good'].includes(raw)) return 'pass';
  if (['fail', 'failed', 'error', 'missing', 'critical'].includes(raw)) return 'fail';
  return 'warning';
};

const humanize = (value: string): string =>
  value
    .replace(/[_-]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
    .replace(/\b\w/g, (letter) => letter.toUpperCase());

const normalizeAtsCheck = (value: unknown, index: number): AtsCheck | null => {
  if (!isRecord(value)) return null;

  const id = readString(value, 'id', 'check_id', 'key', 'code', 'rule') ?? `check-${index + 1}`;
  const label = readString(value, 'label', 'title', 'name', 'check', 'rule') ?? humanize(id);

  return {
    id,
    label,
    status: normalizeAtsStatus(value.status ?? value.result ?? value.state),
    message: readString(value, 'message', 'detail', 'description', 'summary'),
    tip: readString(value, 'tip', 'advice', 'suggestion', 'recommendation', 'fix'),
  };
};

const normalizeAtsChecks = (value: unknown): AtsCheck[] => {
  if (Array.isArray(value)) {
    return value
      .map((item, index) => normalizeAtsCheck(item, index))
      .filter((check): check is AtsCheck => check !== null);
  }

  if (isRecord(value)) {
    return Object.entries(value)
      .map(([key, item], index) => {
        if (isRecord(item)) {
          return normalizeAtsCheck({ id: key, label: humanize(key), ...item }, index);
        }
        return {
          id: key,
          label: humanize(key),
          status: normalizeAtsStatus(item),
        };
      })
      .filter((check): check is AtsCheck => check !== null);
  }

  return [];
};

const buildAtsSummary = (checks: AtsCheck[]): AtsSummary => ({
  passed: checks.filter((check) => check.status === 'pass').length,
  warnings: checks.filter((check) => check.status === 'warning').length,
  failed: checks.filter((check) => check.status === 'fail').length,
});

export const normalizeMatchReport = (value: unknown): MatchReport => {
  const root = isRecord(value) ? (isRecord(value.report) ? value.report : value) : {};
  const rawCategories = root.category_scores ?? root.categoryScores ?? root.breakdown ?? root.scores;
  const categories = isRecord(rawCategories) ? rawCategories : {};

  return {
    overallScore: clampScore(root.overall_score ?? root.overallScore ?? root.score ?? root.match_score),
    matchedSkills: uniqueStrings(root.matched_skills ?? root.matchedSkills ?? root.matched),
    missingMandatorySkills: uniqueStrings(
      root.missing_mandatory_skills ?? root.missingMandatorySkills ?? root.missing_required_skills,
    ),
    missingPreferredSkills: uniqueStrings(
      root.missing_preferred_skills ?? root.missingPreferredSkills ?? root.missing_optional_skills,
    ),
    keywordSuggestions: uniqueStrings(
      root.keyword_suggestions ?? root.keywordSuggestions ?? root.missing_keywords ?? root.missingKeywords,
    ),
    categoryScores: {
      skills: clampScore(categories.skills ?? categories.skill),
      experience: clampScore(categories.experience ?? categories.experience_score),
      education: clampScore(categories.education ?? categories.education_score),
      keywords: clampScore(categories.keywords ?? categories.keyword),
    },
    jobId: readString(root, 'job_id', 'jobId'),
    cvId: readString(root, 'cv_id', 'cvId'),
    jobTitle: readString(root, 'job_title', 'jobTitle'),
    company: readString(root, 'company', 'company_name'),
    generatedAt: readString(root, 'generated_at', 'generatedAt', 'analyzed_at'),
    disclaimer: readString(root, 'disclaimer', 'note'),
  };
};

export const normalizeAtsReadinessReport = (value: unknown): AtsReadinessReport => {
  const root = isRecord(value) ? (isRecord(value.report) ? value.report : value) : {};
  const checks = normalizeAtsChecks(root.checks ?? root.checklist ?? root.items ?? root.results);

  return {
    score: clampScore(root.score ?? root.ats_score ?? root.atsScore ?? root.overall_score),
    checks,
    tips: uniqueStrings(root.tips ?? root.actionable_tips ?? root.recommendations ?? root.suggestions),
    disclaimer:
      readString(root, 'disclaimer', 'note') ??
      'ATS scores are estimates based on standard parsing conventions.',
    summary: buildAtsSummary(checks),
  };
};

const cleanCvId = (cvId: string): string => {
  const trimmed = cvId?.trim();
  if (!trimmed) throw new Error('A CV id is required.');
  return trimmed;
};

/**
 * Requests a structured CV-to-job match report.
 * `POST /api/cvs/{id}/match`
 */
export async function analyzeMatch(cvId: string, payload: MatchAnalysisPayload): Promise<MatchReport> {
  const id = cleanCvId(cvId);
  const jobId = (payload.jobId ?? payload.job_id ?? '').trim();
  const jobDescription = (payload.jobDescription ?? payload.job_description ?? '').trim();

  if (!jobId && !jobDescription) {
    throw new Error('Provide a job id or a job description to analyze.');
  }

  const body: Record<string, string> = {};
  if (jobId) body.job_id = jobId;
  if (jobDescription) body.job_description = jobDescription;

  const response = await fetch(`${API_BASE_URL}/cvs/${encodeURIComponent(id)}/match`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Accept: 'application/json', ...authHeaders() },
    body: JSON.stringify(body),
  });

  const responseBody = await response.json().catch(() => null);
  if (!response.ok) {
    throw new Error(errorMessage(responseBody, 'The match analysis could not be completed.'));
  }

  return normalizeMatchReport(unwrap(responseBody));
}

/**
 * Requests the ATS readability/structure audit for a CV.
 * `GET /api/cvs/{id}/ats-check`
 */
export async function checkAtsReadiness(cvId: string): Promise<AtsReadinessReport> {
  const id = cleanCvId(cvId);
  const response = await fetch(`${API_BASE_URL}/cvs/${encodeURIComponent(id)}/ats-check`, {
    headers: { Accept: 'application/json', ...authHeaders() },
  });

  const body = await response.json().catch(() => null);
  if (!response.ok) {
    throw new Error(errorMessage(body, 'The ATS readiness check could not be completed.'));
  }

  return normalizeAtsReadinessReport(unwrap(body));
}

export const matchingService = {
  analyzeMatch,
  checkAtsReadiness,
  checkAts: checkAtsReadiness,
};
