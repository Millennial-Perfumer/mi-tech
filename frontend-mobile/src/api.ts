import { useAuthStore } from './store/useAuthStore';

export const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export async function apiFetch(endpoint: string, options: RequestInit = {}) {
  const token = useAuthStore.getState().token;
  
  const headers = {
    'Content-Type': 'application/json',
    ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
    ...options.headers,
  };

  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers,
  });

  if (response.status === 401) {
    // Token is invalid or expired - force logout
    console.warn('Unauthorized access (401) detected - logging out.');
    useAuthStore.getState().logout();
    throw new Error('Unauthorized');
  }

  return response;
}
