import { API_BASE_URL, authHeaders } from './api';

export interface WorkExperience {
  id: string;
  title: string;
  company: string;
  duration: string;
  bullets: string[];
}

export interface Education {
  id: string;
  degree: string;
  institution: string;
  years: string;
}

export interface UserProfile {
  fullName: string;
  email: string;
  phone: string;
  location: string;
  summary: string;
  skills: string[];
  workHistory: WorkExperience[];
  education: Education[];
}

const STORAGE_KEY = 'skillmatch_user_profile';

const DEFAULT_PROFILE: UserProfile = {
  fullName: 'Jane Doe',
  email: 'jane.doe@example.com',
  phone: '(555) 123-4567',
  location: 'San Francisco, CA',
  summary:
    'Senior Product Designer and Full-Stack Contributor with 8+ years of experience leading design systems and enterprise UX. Passionate about building accessible, scalable, high-performance digital environments.',
  skills: ['Figma', 'Prototyping', 'Design Systems', 'TypeScript', 'React', 'HTML/CSS', 'User Research'],
  workHistory: [
    {
      id: 'work-1',
      title: 'Senior Product Designer',
      company: 'Acme Corp',
      duration: 'Oct 2020 – Present',
      bullets: [
        'Spearheaded the redesign of the core enterprise dashboard, improving user retention by 24%.',
        'Established a comprehensive design system used by 4 cross-functional teams.',
      ],
    },
    {
      id: 'work-2',
      title: 'Product Designer',
      company: 'TechNova Solutions',
      duration: 'Jun 2017 – Sep 2020',
      bullets: [
        'Led end-to-end product design for B2B SaaS workflows and customer reporting dashboards.',
      ],
    },
  ],
  education: [
    {
      id: 'edu-1',
      degree: 'B.F.A. in Interaction Design',
      institution: 'California College of the Arts',
      years: '2013 – 2017',
    },
  ],
};

export const userService = {
  async getProfile(): Promise<UserProfile> {
    // Try reading cached / persisted profile from localStorage first
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        return JSON.parse(stored);
      }
    } catch {
      // Fall through
    }

    // Try fetching from backend API if available
    try {
      const response = await fetch(`${API_BASE_URL}/users/profile`, {
        headers: authHeaders(),
      });
      if (response.ok) {
        const data = await response.json();
        const profile = data.profile ?? data;
        localStorage.setItem(STORAGE_KEY, JSON.stringify(profile));
        return profile;
      }
    } catch {
      // Fall through to default
    }

    return DEFAULT_PROFILE;
  },

  async updateProfile(profile: UserProfile): Promise<UserProfile> {
    // Save to local storage for immediate persistence
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(profile));
    } catch (err) {
      console.error('Failed caching profile to localStorage:', err);
    }

    // Attempt backend sync
    try {
      const response = await fetch(`${API_BASE_URL}/users/profile`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          ...authHeaders(),
        },
        body: JSON.stringify(profile),
      });
      if (response.ok) {
        const data = await response.json();
        return data.profile ?? data;
      }
    } catch {
      // Backend not running / offline — local persistence is sufficient
    }

    return profile;
  },
};
