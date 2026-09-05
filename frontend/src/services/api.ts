// frontend/src/services/api.ts

export const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

/**
 * Returns authorization headers containing the stored JWT Bearer token, if available.
 */
export const authHeaders = (): Record<string, string> => {
  const token = localStorage.getItem('token');
  return token ? { Authorization: `Bearer ${token}` } : {};
};

/**
 * Consistently extracts user-facing error messages from API responses,
 * falling back to a provided default message.
 */
export const getErrorMessage = async (
  response: Response,
  fallback: string
): Promise<string> => {
  const body = await response.json().catch(() => ({}));
  return (
    body?.error?.message ||
    body?.error ||
    body?.message ||
    body?.detail ||
    fallback
  );
};
