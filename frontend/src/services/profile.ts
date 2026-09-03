const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

export interface Profile {
  id: string;
  bio: string;
  skills: string[];
  experience: string[];
  resumeUrl: string;
  updatedAt?: string;
}

export interface ProfileInput {
  bio: string;
  skills: string[];
  experience: string[];
  resumeUrl: string;
}

const authHeaders = (): Record<string, string> => {
  const token = localStorage.getItem('token');
  return token ? { Authorization: `Bearer ${token}` } : {};
};

const errorMessage = (body: any, fallback: string): string =>
  body?.error?.message || body?.error || body?.message || fallback;

const asStringArray = (value: unknown): string[] =>
  Array.isArray(value) ? value.map(String).map((item) => item.trim()).filter(Boolean) : [];

const normalize = (item: any): Profile => ({
  id: String(item.user_id ?? item.id ?? ''),
  bio: typeof item.bio === 'string' ? item.bio : '',
  skills: asStringArray(item.skills),
  experience: asStringArray(item.experience),
  resumeUrl: item.resume_url ?? item.resumeUrl ?? '',
  updatedAt: item.updated_at ?? item.updatedAt,
});

export const profileService = {
  async get(): Promise<Profile | null> {
    const response = await fetch(`${API_BASE_URL}/profile`, { headers: authHeaders() });
    if (response.status === 404) return null;
    const body = await response.json().catch(() => null);
    if (!response.ok) throw new Error(errorMessage(body, 'Profile could not be loaded.'));
    const inner = body?.data ?? body;
    if (!inner || typeof inner !== 'object') return null;
    return normalize(inner);
  },

  async update(input: ProfileInput): Promise<Profile> {
    const response = await fetch(`${API_BASE_URL}/profile`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', ...authHeaders() },
      body: JSON.stringify({ bio: input.bio, skills: input.skills, experience: input.experience, resume_url: input.resumeUrl }),
    });
    const body = await response.json().catch(() => null);
    if (!response.ok) throw new Error(errorMessage(body, 'Profile could not be saved.'));
    return normalize(body?.data ?? body);
  },
};
