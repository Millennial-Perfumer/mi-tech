import { create } from 'zustand';
import { apiFetch } from '../api';

interface Template {
  id: string;
  name: string;
  content: string;
  trigger_event: string;
  is_active: boolean;
}

interface AutomationState {
  templates: Template[];
  isLoading: boolean;
  error: string | null;
  fetchTemplates: () => Promise<void>;
  toggleTemplate: (id: string, active: boolean) => Promise<void>;
}

export const useAutomationStore = create<AutomationState>((set, get) => ({
  templates: [],
  isLoading: false,
  error: null,
  fetchTemplates: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiFetch(`/api/automation/templates`);
      const data = await response.json();
      if (data.success) {
        set({ templates: data.templates, isLoading: false });
      } else {
        set({ error: data.message || 'Failed to fetch templates', isLoading: false });
      }
    } catch (err: any) {
      if (err.message !== 'Unauthorized') {
        set({ error: 'Network error', isLoading: false });
      } else {
        set({ isLoading: false });
      }
    }
  },
  toggleTemplate: async (id, active) => {
    try {
      const response = await apiFetch(`/api/automation/templates/${id}/toggle`, {
        method: 'POST',
        body: JSON.stringify({ is_active: active })
      });
      const data = await response.json();
      if (data.success) {
        set({
          templates: get().templates.map(t => t.id === id ? { ...t, is_active: active } : t)
        });
      }
    } catch (err: any) {
      console.error('Failed to toggle template', err);
    }
  }
}));
