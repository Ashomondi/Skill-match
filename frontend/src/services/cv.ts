import { authService } from './auth';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

export const CV_ACCENT_COLORS = [
  { id: 'slate', label: 'Slate', value: '#475569' },
  { id: 'navy', label: 'Navy', value: '#1E3A8A' },
  { id: 'emerald', label: 'Emerald', value: '#047857' },
  { id: 'burgundy', label: 'Burgundy', value: '#7F1D1D' },
  { id: 'charcoal', label: 'Charcoal', value: '#1F2937' },
] as const;

export const CV_TYPOGRAPHY = [
  { id: 'sans', label: 'Sans', family: 'Inter', fontFamily: 'Inter, ui-sans-serif, system-ui, -apple-system, sans-serif' },
  { id: 'serif', label: 'Serif', family: 'Merriweather', fontFamily: 'Merriweather, Georgia, "Times New Roman", serif' },
  { id: 'clean', label: 'Clean', family: 'Roboto', fontFamily: 'Roboto, "Helvetica Neue", Arial, sans-serif' },
] as const;

export const CV_SPACING = [
  { id: 'compact', label: 'Compact', value: '0.75rem' },
  { id: 'normal', label: 'Normal', value: '1.25rem' },
  { id: 'spacious', label: 'Spacious', value: '1.75rem' },
] as const;

export type CvAccentColor = (typeof CV_ACCENT_COLORS)[number]['id'];
export type CvTypography = (typeof CV_TYPOGRAPHY)[number]['id'];
export type CvSpacing = (typeof CV_SPACING)[number]['id'];

export interface CvThemeConfig {
  accentColor: CvAccentColor;
  typography: CvTypography;
  spacing: CvSpacing;
}

export const DEFAULT_CV_THEME: CvThemeConfig = {
  accentColor: 'navy',
  typography: 'sans',
  spacing: 'normal',
};

export type CvSectionType =
  | 'summary'
  | 'experience'
  | 'education'
  | 'skills'
  | 'projects'
  | 'custom';

export interface CvSection {
  id: string;
  cvId: string;
  sectionType: CvSectionType;
  title: string;
  sortOrder: number;
  content: Record<string, unknown>;
  createdAt?: string;
  updatedAt?: string;
}

export interface CvDocument {
  id: string;
  userId: string;
  title: string;
  targetRole: string;
  templateId: string;
  themeConfig: CvThemeConfig;
  isPrimary: boolean;
  createdAt?: string;
  updatedAt?: string;
  sections: CvSection[];
}

export interface UpdateCvInput {
  title?: string;
  targetRole?: string;
  templateId?: string;
  themeConfig?: CvThemeConfig;
  isPrimary?: boolean;
}

export interface CreateCvInput {
  title: string;
  targetRole?: string;
  templateId?: string;
  themeConfig?: CvThemeConfig;
  isPrimary?: boolean;
  resumeId?: string;
  sections?: Array<{
    id?: string;
    sectionType: CvSectionType;
    title: string;
    content?: Record<string, unknown>;
  }>;
}

type JsonRecord = Record<string, unknown>;

const isRecord = (value: unknown): value is JsonRecord =>
  typeof value === 'object' && value !== null && !Array.isArray(value);

const readString = (source: JsonRecord, ...keys: string[]): string | undefined => {
  for (const key of keys) {
    const value = source[key];
    if (typeof value === 'string') return value;
    if (typeof value === 'number') return String(value);
  }
  return undefined;
};

const authHeaders = (): Record<string, string> => {
  const token = authService.getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
};

const errorMessage = (body: any, fallback: string): string =>
  body?.error?.message || (typeof body?.error === 'string' ? body.error : '') || body?.message || fallback;

const normalizeAccent = (value: unknown): CvAccentColor => {
  const raw = String(value ?? '').trim().toLowerCase();
  const match = CV_ACCENT_COLORS.find(
    (option) => option.id === raw || option.label.toLowerCase() === raw || option.value.toLowerCase() === raw,
  );
  if (match) return match.id;
  if (raw.includes('emerald') || raw === 'green') return 'emerald';
  if (raw.includes('burgundy') || raw === 'red') return 'burgundy';
  if (raw.includes('charcoal') || ['black', 'gray', 'grey'].includes(raw)) return 'charcoal';
  if (raw.includes('navy') || raw === 'blue') return 'navy';
  return DEFAULT_CV_THEME.accentColor;
};

const normalizeTypography = (value: unknown): CvTypography => {
  const raw = String(value ?? '').trim().toLowerCase();
  const match = CV_TYPOGRAPHY.find(
    (option) =>
      option.id === raw ||
      option.label.toLowerCase() === raw ||
      option.family.toLowerCase() === raw ||
      option.fontFamily.toLowerCase() === raw,
  );
  if (match) return match.id;
  if (raw.includes('merriweather') || raw.includes('serif')) return 'serif';
  if (raw.includes('roboto') || raw.includes('clean')) return 'clean';
  return DEFAULT_CV_THEME.typography;
};

const normalizeSpacing = (value: unknown): CvSpacing => {
  const raw = String(value ?? '').trim().toLowerCase();
  const match = CV_SPACING.find(
    (option) => option.id === raw || option.label.toLowerCase() === raw || option.value === raw,
  );
  if (match) return match.id;
  if (raw.includes('compact') || raw.includes('tight') || raw === 'small') return 'compact';
  if (raw.includes('spacious') || raw.includes('relaxed') || raw === 'large') return 'spacious';
  return DEFAULT_CV_THEME.spacing;
};

/** Normalises the JSONB `theme_config` returned by the API. */
export const normalizeThemeConfig = (value: unknown): CvThemeConfig => {
  if (!isRecord(value)) return { ...DEFAULT_CV_THEME };

  return {
    accentColor: normalizeAccent(value.accentColor ?? value.accent_color ?? value.accent ?? value.color),
    typography: normalizeTypography(
      value.typography ?? value.font ?? value.fontFamily ?? value.font_family ?? value.type,
    ),
    spacing: normalizeSpacing(value.spacing ?? value.density ?? value.layout),
  };
};

const normalizeSection = (value: unknown): CvSection | null => {
  if (!isRecord(value)) return null;
  const id = readString(value, 'id');
  if (!id) return null;

  return {
    id,
    cvId: readString(value, 'cv_id', 'cvId') ?? '',
    sectionType: (readString(value, 'section_type', 'sectionType') ?? 'custom') as CvSectionType,
    title: readString(value, 'title') ?? '',
    sortOrder: Number(value.sort_order ?? value.sortOrder ?? 0),
    content: isRecord(value.content) ? value.content : {},
    createdAt: readString(value, 'created_at', 'createdAt'),
    updatedAt: readString(value, 'updated_at', 'updatedAt'),
  };
};

const normalizeDocument = (value: unknown): CvDocument => {
  const source = isRecord(value) ? value : {};

  return {
    id: readString(source, 'id') ?? '',
    userId: readString(source, 'user_id', 'userId') ?? '',
    title: readString(source, 'title') ?? '',
    targetRole: readString(source, 'target_role', 'targetRole') ?? '',
    templateId: readString(source, 'template_id', 'templateId') ?? '',
    themeConfig: normalizeThemeConfig(source.theme_config ?? source.themeConfig),
    isPrimary: Boolean(source.is_primary ?? source.isPrimary),
    createdAt: readString(source, 'created_at', 'createdAt'),
    updatedAt: readString(source, 'updated_at', 'updatedAt'),
    sections: Array.isArray(source.sections)
      ? source.sections.map(normalizeSection).filter((section): section is CvSection => section !== null)
      : [],
  };
};

const unwrap = (body: unknown): unknown => (isRecord(body) ? body.data ?? body : body);

const buildUpdatePayload = (input: UpdateCvInput): Record<string, unknown> => {
  const payload: Record<string, unknown> = {};
  if (input.title !== undefined) payload.title = input.title;
  if (input.targetRole !== undefined) payload.target_role = input.targetRole;
  if (input.templateId !== undefined) payload.template_id = input.templateId;
  if (input.themeConfig !== undefined) payload.theme_config = normalizeThemeConfig(input.themeConfig);
  if (input.isPrimary !== undefined) payload.is_primary = input.isPrimary;
  return payload;
};

const buildCreatePayload = (input: CreateCvInput): Record<string, unknown> => {
  const payload: Record<string, unknown> = {
    title: input.title,
    target_role: input.targetRole ?? '',
    template_id: input.templateId ?? '',
    theme_config: normalizeThemeConfig(input.themeConfig ?? DEFAULT_CV_THEME),
    is_primary: input.isPrimary ?? false,
  };
  if (input.resumeId !== undefined) payload.resume_id = input.resumeId;
  if (input.sections !== undefined) {
    payload.sections = input.sections.map((section) => ({
      ...(section.id ? { id: section.id } : {}),
      section_type: section.sectionType,
      title: section.title,
      content: section.content ?? {},
    }));
  }
  return payload;
};

export async function listCvs(limit?: number): Promise<CvDocument[]> {
  const query = limit && limit > 0 ? `?limit=${encodeURIComponent(String(limit))}` : '';
  const response = await fetch(`${API_BASE_URL}/cvs${query}`, {
    headers: { Accept: 'application/json', ...authHeaders() },
  });
  const body = await response.json().catch(() => null);
  if (!response.ok) throw new Error(errorMessage(body, 'CVs could not be loaded.'));

  const data = unwrap(body);
  const items = Array.isArray(data) ? data : isRecord(data) && Array.isArray(data.cvs) ? data.cvs : [];
  return items.map(normalizeDocument);
}

export async function getCv(id: string): Promise<CvDocument> {
  const response = await fetch(`${API_BASE_URL}/cvs/${encodeURIComponent(id)}`, {
    headers: { Accept: 'application/json', ...authHeaders() },
  });
  const body = await response.json().catch(() => null);
  if (!response.ok) throw new Error(errorMessage(body, 'CV could not be loaded.'));
  return normalizeDocument(unwrap(body));
}

export async function updateCv(id: string, input: UpdateCvInput): Promise<CvDocument> {
  const payload = buildUpdatePayload(input);
  if (Object.keys(payload).length === 0) throw new Error('No CV changes were provided.');

  const response = await fetch(`${API_BASE_URL}/cvs/${encodeURIComponent(id)}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', Accept: 'application/json', ...authHeaders() },
    body: JSON.stringify(payload),
  });
  const body = await response.json().catch(() => null);
  if (!response.ok) throw new Error(errorMessage(body, 'CV could not be saved.'));
  return normalizeDocument(unwrap(body));
}

/** Persists just the theme settings for a CV document. */
export async function updateCvTheme(id: string, themeConfig: CvThemeConfig): Promise<CvDocument> {
  return updateCv(id, { themeConfig });
}

export async function createCv(input: CreateCvInput): Promise<CvDocument> {
  const response = await fetch(`${API_BASE_URL}/cvs`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Accept: 'application/json', ...authHeaders() },
    body: JSON.stringify(buildCreatePayload(input)),
  });
  const body = await response.json().catch(() => null);
  if (!response.ok) throw new Error(errorMessage(body, 'CV could not be created.'));
  return normalizeDocument(unwrap(body));
}

export async function deleteCv(id: string): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/cvs/${encodeURIComponent(id)}`, {
    method: 'DELETE',
    headers: { ...authHeaders() },
  });
  if (!response.ok) {
    const body = await response.json().catch(() => null);
    throw new Error(errorMessage(body, 'CV could not be deleted.'));
  }
}

export const cvService = {
  list: listCvs,
  get: getCv,
  create: createCv,
  update: updateCv,
  updateTheme: updateCvTheme,
  remove: deleteCv,
};
