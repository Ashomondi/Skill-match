import { API_BASE_URL, getErrorMessage } from './api';

const DEMO_AUTH_ENABLED = import.meta.env.DEV && import.meta.env.VITE_DEMO_AUTH_ENABLED !== 'false';
const DEMO_EMAIL = import.meta.env.VITE_DEMO_EMAIL || 'demo@skill-match.test';
const DEMO_PASSWORD = import.meta.env.VITE_DEMO_PASSWORD || 'password123';
const DEMO_NAME = import.meta.env.VITE_DEMO_NAME || 'Demo User';

// The backend wraps success payloads in { data: ... }.
const unwrap = (body: any): AuthResponse => (body?.data ?? body) as AuthResponse;

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterCredentials extends LoginCredentials {
  fullName: string;
}

export interface AuthResponse {
  token: string;
  user: {
    id: string;
    email: string;
    fullName: string;
  };
}

export const authService = {
  async login(credentials: LoginCredentials): Promise<AuthResponse> {
    if (
      DEMO_AUTH_ENABLED &&
      DEMO_EMAIL &&
      DEMO_PASSWORD &&
      credentials.email === DEMO_EMAIL &&
      credentials.password === DEMO_PASSWORD
    ) {
      const demoData: AuthResponse = {
        token: 'demo-jwt-token',
        user: { id: 'demo-user-1', email: DEMO_EMAIL, fullName: DEMO_NAME },
      };
      localStorage.setItem('token', demoData.token);
      localStorage.setItem('user', JSON.stringify(demoData.user));
      return demoData;
    }

    const response = await fetch(`${API_BASE_URL}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(credentials),
    });

    if (!response.ok) {
      throw new Error(await getErrorMessage(response, 'Failed to log in'));
    }

    const data = unwrap(await response.json());
    localStorage.setItem('token', data.token);
    localStorage.setItem('user', JSON.stringify(data.user));
    return data;
  },

  async register(credentials: RegisterCredentials): Promise<AuthResponse> {
    const response = await fetch(`${API_BASE_URL}/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(credentials),
    });

    if (!response.ok) {
      throw new Error(await getErrorMessage(response, 'Failed to register account'));
    }

    const data = unwrap(await response.json());
    localStorage.setItem('token', data.token);
    localStorage.setItem('user', JSON.stringify(data.user));
    return data;
  },

  logout(): void {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
  },

  getCurrentUser() {
    const userStr = localStorage.getItem('user');
    return userStr ? JSON.parse(userStr) : null;
  },

  getToken(): string | null {
    return localStorage.getItem('token');
  }
};
