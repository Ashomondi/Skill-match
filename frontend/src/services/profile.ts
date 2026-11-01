import { authService } from './auth';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

export interface WorkExperience {
  id?: string;
  title: string;
  company?: string;
  location?: string;
  startDate?: string;
  endDate?: string;
  current?: boolean;
  description?: string;
}

export interface Education {
  id?: string;
  institution: string;
  degree?: string;
  fieldOfStudy?: string;
  startDate?: string;
  endDate?: string;
  description?: string;
}

/**
 * User profile returned by `GET /profile`.
 *
 * `experience` is the flat, single-line representation persisted by the API.
 * `workExperience` is the structured projection used by the profile UI.
 * `education` is only populated when the API returns it; the current profile
 * endpoint does not persist education yet.
 */
export interface UserProfile {
  userId: string;
  bio: string;
  skills: string[];
  experience: string[];
  workExperience: WorkExperience[];
  education: Education[];
  resumeUrl: string;
  updatedAt?: string;
}

/** Fields that `PUT /profile` can persist. */
export type ProfileUpdate = Partial<
  Pick<UserProfile, 'bio' | 'skills' | 'experience' | 'workExperience' | 'resumeUrl'>
>;

type JsonRecord = Record<string, unknown>;

const isRecord = (value: unknown): value is JsonRecord =>
  typeof value === 'object' && value !== null && !Array.isArray(value);

const asArray = (value: unknown): unknown[] => (Array.isArray(value) ? value : []);

const asString = (value: unknown): string | undefined => {
  if (typeof value === 'string') return value;
  if (typeof value === 'number') return String(value);
  return undefined;
};

const readString = (source: JsonRecord, ...keys: string[]): string | undefined => {
  for (const key of keys) {
    const value = asString(source[key]);
    if (value !== undefined) return value;
  }
  return undefined;
};

const asStringArray = (value: unknown): string[] =>
  asArray(value)
    .map((item) => (asString(item) ?? '').trim())
    .filter(Boolean);

const authHeaders = (): Record<string, string> => {
  const token = authService.getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
};

const errorMessage = (body: any, fallback: string): string =>
  body?.error?.message || (typeof body?.error === 'string' ? body.error : '') || body?.message || fallback;

const normalizeWorkExperience = (value: unknown): WorkExperience | null => {
  if (typeof value === 'string') {
    const title = value.trim();
    return title ? { title } : null;
  }
  if (!isRecord(value)) return null;

  const title = readString(value, 'title', 'role', 'position', 'job_title')?.trim();
  const company = readString(value, 'company', 'organization', 'employer')?.trim();
  if (!title && !company) return null;

  const endDate = readString(value, 'endDate', 'end_date', 'end')?.trim();
  const current =
    value.current === true ||
    value.is_current === true ||
    (typeof endDate === 'string' && endDate.toLowerCase() === 'present');

  return {
    id: readString(value, 'id'),
    title: title || company || '',
    company: company || undefined,
    location: readString(value, 'location')?.trim() || undefined,
    startDate: readString(value, 'startDate', 'start_date', 'start')?.trim() || undefined,
    endDate: endDate || undefined,
    current,
    description: readString(value, 'description', 'summary')?.trim() || undefined,
  };
};

const normalizeEducation = (value: unknown): Education | null => {
  if (typeof value === 'string') {
    const institution = value.trim();
    return institution ? { institution } : null;
  }
  if (!isRecord(value)) return null;

  const institution = readString(
    value,
    'institution',
    'school',
    'university',
    'college',
  )?.trim();
  const degree = readString(value, 'degree', 'qualification')?.trim();
  if (!institution && !degree) return null;

  return {
    id: readString(value, 'id'),
    institution: institution || degree || '',
    degree: degree || undefined,
    fieldOfStudy: readString(value, 'fieldOfStudy', 'field_of_study', 'major')?.trim() || undefined,
    startDate: readString(value, 'startDate', 'start_date', 'start')?.trim() || undefined,
    endDate: readString(value, 'endDate', 'end_date', 'end')?.trim() || undefined,
    description: readString(value, 'description', 'summary')?.trim() || undefined,
  };
};

/**
 * Flattens a structured experience entry into the single-line format the
 * profile API accepts (max 500 characters per entry).
 */
export const formatWorkExperience = (item: WorkExperience): string => {
  const headline = [item.title, item.company].filter(Boolean).join(item.title && item.company ? ' at ' : '');
  const range = [item.startDate, item.current ? 'Present' : item.endDate].filter(Boolean).join(' - ');
  const details = [range, item.location, item.description].filter(Boolean).join(' · ');
  const line = [headline, details].filter(Boolean).join(' — ');
  return line.replace(/\s+/g, ' ').trim().slice(0, 500);
};

export const createEmptyProfile = (): UserProfile => ({
  userId: '',
  bio: '',
  skills: [],
  experience: [],
  workExperience: [],
  education: [],
  resumeUrl: '',
});

const normalizeProfile = (value: unknown): UserProfile => {
  if (!isRecord(value)) return createEmptyProfile();

  const rawExperience = value.work_experience ?? value.workExperience ?? value.experience;
  const workExperience = asArray(rawExperience)
    .map(normalizeWorkExperience)
    .filter((item): item is WorkExperience => item !== null);

  const experience = Array.isArray(value.experience)
    ? asStringArray(value.experience)
    : workExperience.map(formatWorkExperience).filter(Boolean);

  return {
    userId: readString(value, 'user_id', 'userId', 'id') ?? '',
    bio: readString(value, 'bio') ?? '',
    skills: asStringArray(value.skills),
    experience,
    workExperience,
    education: asArray(value.education ?? value.educations)
      .map(normalizeEducation)
      .filter((item): item is Education => item !== null),
    resumeUrl: readString(value, 'resume_url', 'resumeUrl') ?? '',
    updatedAt: readString(value, 'updated_at', 'updatedAt'),
  };
};

const buildUpdatePayload = (data: ProfileUpdate): Record<string, unknown> => {
  const payload: Record<string, unknown> = {};

  if (data.bio !== undefined) payload.bio = data.bio;
  if (data.skills !== undefined) payload.skills = data.skills;
  if (data.resumeUrl !== undefined) payload.resume_url = data.resumeUrl;

  if (data.workExperience !== undefined) {
    payload.experience = data.workExperience.map(formatWorkExperience).filter(Boolean);
  } else if (data.experience !== undefined) {
    payload.experience = data.experience.map((entry) => entry.trim()).filter(Boolean);
  }

  return payload;
};

export async function getProfile(): Promise<UserProfile> {
  const response = await fetch(`${API_BASE_URL}/profile`, {
    headers: { Accept: 'application/json', ...authHeaders() },
  });

  // The endpoint returns a default empty profile for new users, but treat a
  // missing profile as empty too so the UI can always render.
  if (response.status === 404) return createEmptyProfile();

  const body = await response.json().catch(() => null);
  if (!response.ok) {
    throw new Error(errorMessage(body, 'Profile could not be loaded.'));
  }

  return normalizeProfile(isRecord(body) ? body.data ?? body : body);
}

export async function updateProfile(data: ProfileUpdate): Promise<UserProfile> {
  const payload = buildUpdatePayload(data);
  if (Object.keys(payload).length === 0) {
    throw new Error('No profile changes were provided.');
  }

  const response = await fetch(`${API_BASE_URL}/profile`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', Accept: 'application/json', ...authHeaders() },
    body: JSON.stringify(payload),
  });

  const body = await response.json().catch(() => null);
  if (!response.ok) {
    throw new Error(errorMessage(body, 'Profile could not be saved.'));
  }

  return normalizeProfile(body?.data ?? body);
}

export const profileService = {
  getProfile,
  updateProfile,
  // Short aliases keep call sites readable.
  get: getProfile,
  update: updateProfile,
};
