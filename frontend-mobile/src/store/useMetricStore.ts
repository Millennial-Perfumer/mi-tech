import { create } from 'zustand';
import { apiFetch } from '../api';

interface Metrics {
  total_revenue: number;
  total_invoices: number;
  total_gst_collected: number;
  total_orders: number;
  fulfilled_orders: number;
  unfulfilled_orders: number;
}

interface MetricsState {
  metrics: Metrics | null;
  isLoading: boolean;
  error: string | null;
  fetchMetrics: (startDate: string, endDate: string) => Promise<void>;
}

export const useMetricStore = create<MetricsState>((set) => ({
  metrics: null,
  isLoading: false,
  error: null,
  fetchMetrics: async (startDate, endDate) => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiFetch(`/api/dashboard/metrics?start_date=${startDate}&end_date=${endDate}`);
      const data = await response.json();
      if (data.success) {
        set({ metrics: data.metrics, isLoading: false });
      } else {
        set({ error: data.message || 'Failed to fetch metrics', isLoading: false });
      }
    } catch (err: any) {
      if (err.message !== 'Unauthorized') {
        set({ error: 'Network error', isLoading: false });
      } else {
        set({ isLoading: false });
      }
    }
  },
}));
