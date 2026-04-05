import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface User {
  username: string;
  role: 'admin' | 'read';
}

interface AuthState {
  token: string | null;
  user: User | null;
  setToken: (token: string | null) => void;
  logout: () => void;
  getUserRole: () => 'admin' | 'read';
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: localStorage.getItem('token'),
      user: null, // User info usually extracted from JWT or fetched
      setToken: (token) => {
        if (token) {
          localStorage.setItem('token', token);
          // Simple JWT decode for role (mimicking web app logic)
          try {
            const payload = JSON.parse(atob(token.split('.')[1]));
            set({ token, user: { username: payload.username, role: payload.role || 'read' } });
          } catch (e) {
            set({ token, user: null });
          }
        } else {
          get().logout();
        }
      },
      logout: () => {
        localStorage.removeItem('token');
        set({ token: null, user: null });
      },
      getUserRole: () => get().user?.role || 'read',
    }),
    {
      name: 'auth-storage',
    }
  )
);
