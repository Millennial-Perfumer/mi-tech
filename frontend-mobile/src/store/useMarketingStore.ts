import { create } from 'zustand';
import { apiFetch } from '../api';

interface Insight {
  id: string;
  name: string;
  status: string;
  spend: number;
  impressions: number;
  clicks: number;
  ctr: number;
  cpc: number;
  reach: number;
}

interface MarketingState {
  campaigns: Insight[];
  adsets: Insight[];
  ads: Insight[];
  summary: any[];
  isLoading: boolean;
  error: string | null;
  fetchOverview: () => Promise<void>;
  fetchAdSets: (campaignId: string) => Promise<void>;
  fetchAds: (adsetId: string) => Promise<void>;
}

export const useMarketingStore = create<MarketingState>((set) => ({
  campaigns: [],
  adsets: [],
  ads: [],
  summary: [],
  isLoading: false,
  error: null,
  fetchOverview: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiFetch(`/api/marketing/meta/overview`);
      const data = await response.json();
      if (data.success) {
        // Overview returns campaigns via insights at account level,
        // but we might want the actual campaigns list too.
        // The backend GetMetaOverview returns 'insights' (campaign level) and 'summary' (account level).
        set({ 
          campaigns: data.insights || [], 
          summary: data.summary || [],
          isLoading: false 
        });
      } else {
        set({ error: data.message || 'Failed to fetch overview', isLoading: false });
      }
    } catch (err: any) {
      if (err.message !== 'Unauthorized') {
        set({ error: 'Network error', isLoading: false });
      } else {
        set({ isLoading: false });
      }
    }
  },
  fetchAdSets: async (campaignId) => {
    set({ isLoading: true, error: null, adsets: [], ads: [] });
    try {
      const response = await apiFetch(`/api/marketing/meta/adsets?campaign_id=${campaignId}`);
      const data = await response.json();
      if (data.success) {
        set({ adsets: data.adsets || [], isLoading: false });
      } else {
        set({ error: data.message || 'Failed to fetch adsets', isLoading: false });
      }
    } catch (err: any) {
      if (err.message !== 'Unauthorized') {
        set({ error: 'Network error', isLoading: false });
      } else {
        set({ isLoading: false });
      }
    }
  },
  fetchAds: async (adsetId) => {
    set({ isLoading: true, error: null, ads: [] });
    try {
      const response = await apiFetch(`/api/marketing/meta/ads?adset_id=${adsetId}`);
      const data = await response.json();
      if (data.success) {
        set({ ads: data.ads || [], isLoading: false });
      } else {
        set({ error: data.message || 'Failed to fetch ads', isLoading: false });
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
