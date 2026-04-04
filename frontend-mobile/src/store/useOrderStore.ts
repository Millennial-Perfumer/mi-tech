import { create } from 'zustand';
import { apiFetch } from '../api';

interface Order {
  id: string;
  order_number: string;
  customer_name: string;
  total_price: string;
  financial_status: string;
  fulfillment_status: string;
  created_at: string;
}

interface OrderState {
  orders: Order[];
  isLoading: boolean;
  error: string | null;
  fetchOrders: (search?: string) => Promise<void>;
}

export const useOrderStore = create<OrderState>((set) => ({
  orders: [],
  isLoading: false,
  error: null,
  fetchOrders: async (search = '') => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiFetch(`/api/shopify/orders?search=${search}`);
      const data = await response.json();
      if (data.success) {
        set({ orders: data.orders, isLoading: false });
      } else {
        set({ error: data.message || 'Failed to fetch orders', isLoading: false });
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
