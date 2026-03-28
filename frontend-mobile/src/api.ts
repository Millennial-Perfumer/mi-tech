/* eslint-disable @typescript-eslint/no-explicit-any */
export const API_BASE = (import.meta as any).env.VITE_API_URL || 'http://localhost:8080';

export const handleLogout = () => {
  localStorage.removeItem('mobileToken');
  window.location.reload();
};

export const fetchWithAuth = async (url: string, options: RequestInit = {}) => {
  const token = localStorage.getItem('mobileToken');
  const headers: Record<string, string> = {
    ...options.headers as Record<string, string>,
    'Authorization': `Bearer ${token}`
  };

  try {
    const response = await fetch(url, { ...options, headers });
    if (response.status === 401) {
      handleLogout();
    }
    return response;
  } catch (error) {
    console.error('API Error:', error);
    throw error;
  }
};

export const getTodayIST = () => {
  return new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Kolkata' });
};
