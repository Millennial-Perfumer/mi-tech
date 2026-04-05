import { create } from 'zustand';
import { apiFetch } from '../api';

interface Customer {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  total_spent: string;
  orders_count: number;
}

interface CustomerState {
  customers: Customer[];
  isLoading: boolean;
  error: string | null;
  fetchCustomers: (search?: string) => Promise<void>;
}

export const useCustomerStore = create<CustomerState>((set) => ({
  customers: [],
  isLoading: false,
  error: null,
  fetchCustomers: async (search = '') => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiFetch(`/api/shopify/customers?search=${search}`);
      const data = await response.json();
      if (data.success) {
        set({ customers: data.customers, isLoading: false });
      } else {
        set({ error: data.message || 'Failed to fetch customers', isLoading: false });
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
