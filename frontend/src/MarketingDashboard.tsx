import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { API_BASE } from './api';
import { CustomDatePicker } from './CustomDatePicker';
import { subDays, format } from 'date-fns';
import './App.css'; 

interface MarketingDashboardProps {
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
  userRole: string;
  onNavigate?: (tab: string) => void;
}

interface Insight {
  spend: string;
  reach: string;
  impressions: string;
  inline_link_clicks: string;
  ctr: string;
  cpp: string;
  purchase_roas: string;
  purchase_roas_val: string;
  frequency: string;
  purchase_value: string;
  conversions: string;
}

export const MarketingDashboard: React.FC<MarketingDashboardProps> = ({ fetchWithAuth, userRole, onNavigate }) => {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  // Drill-down State
  const [viewLevel, setViewLevel] = useState<'CAMPAIGNS' | 'AD_SETS' | 'ADS'>('CAMPAIGNS');
  const [activeCampaign, setActiveCampaign] = useState<{id: string, name: string} | null>(null);
  const [activeAdSet, setActiveAdSet] = useState<{id: string, name: string} | null>(null);
  
  // Modal State
  const [selectedAd, setSelectedAd] = useState<any | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  
  const [metaData, setMetaData] = useState<{
    metrics: Insight;
    campaigns: any[];
    adsets: any[];
    ads: any[];
    insights: any[];
    summary: any[];
    activeId: string;
    accounts: any[];
  }>({
    metrics: { 
      spend: '0.00', reach: '0', impressions: '0', inline_link_clicks: '0', 
      ctr: '0.00', cpp: '0.00', purchase_roas: '0.00', purchase_roas_val: '0.00',
      frequency: '0.0', purchase_value: '0.00', conversions: '0'
    },
    campaigns: [],
    adsets: [],
    ads: [],
    insights: [],
    summary: [],
    activeId: '',
    accounts: []
  });

  // Date State (Persist across refreshes)
  const [startDate, setStartDate] = useState(localStorage.getItem('meta_start_date') || format(subDays(new Date(), 29), 'yyyy-MM-dd'));
  const [endDate, setEndDate] = useState(localStorage.getItem('meta_end_date') || format(new Date(), 'yyyy-MM-dd'));

  const fetchMetaOverview = useCallback(async (start?: string, end?: string) => {
    setIsLoading(true);
    setError(null);
    try {
      const s = start || startDate;
      const e = end || endDate;
      
      const resp = await fetchWithAuth(`${API_BASE}/api/marketing/meta/overview?start_date=${s}&end_date=${e}`);
      const data = await resp.json();
      
      if (data.success) {
        // Aggregate metrics from the individual campaign insights
        const aggregated = aggregateMetrics(data.insights || []);
        
        let campaignList = [];
        if (data.active_id) {
          const campResp = await fetchWithAuth(`${API_BASE}/api/marketing/meta/campaigns?ad_account_id=${data.active_id}`);
          const campData = await campResp.json();
          if (campData.success) {
            campaignList = campData.campaigns || [];
          }
        }

        setMetaData(prev => ({
          ...prev,
          metrics: aggregated,
          campaigns: campaignList,
          insights: data.insights || [],
          summary: data.summary || [],
          activeId: data.active_id || prev.activeId,
          accounts: data.accounts || prev.accounts
        }));
        } else {
          setError(data.message || 'Failed to fetch Meta data');
        }
      } catch (err) {
        console.error('Meta fetch error:', err);
        setError('Connection error. Please check your API settings.');
      } finally {
        setIsLoading(false);
      }
    }, [fetchWithAuth, startDate, endDate]);
  
    const fetchAdSets = async (campaignId: string, campaignName: string, start?: string, end?: string) => {
      setIsLoading(true);
      setError(null);
      const s = start || startDate;
      const e = end || endDate;
      try {
        const resp = await fetchWithAuth(`${API_BASE}/api/marketing/meta/adsets?campaign_id=${campaignId}&start_date=${s}&end_date=${e}`);
        const data = await resp.json();
        if (data.success) {
          setActiveCampaign({ id: campaignId, name: campaignName });
          const aggregated = aggregateMetrics(data.insights || []);
          setMetaData(prev => ({ 
            ...prev, 
            adsets: data.adsets || [],
            insights: data.insights || [],
            metrics: aggregated
          }));
          setViewLevel('AD_SETS');
        } else {
          setError(data.message || 'Failed to load Ad Sets');
        }
      } catch (err) {
        console.error('Fetch AdSets error:', err);
        setError('API Error: Could not retrieve Ad Sets.');
      } finally {
        setIsLoading(false);
      }
    };
  
    const fetchAds = async (adsetId: string, adsetName: string, start?: string, end?: string) => {
      setIsLoading(true);
      setError(null);
      const s = start || startDate;
      const e = end || endDate;
      try {
        const resp = await fetchWithAuth(`${API_BASE}/api/marketing/meta/ads?adset_id=${adsetId}&start_date=${s}&end_date=${e}`);
        const data = await resp.json();
        if (data.success) {
          setActiveAdSet({ id: adsetId, name: adsetName });
          const aggregated = aggregateMetrics(data.insights || []);
          setMetaData(prev => ({ 
            ...prev, 
            ads: data.ads || [],
            insights: data.insights || [],
            metrics: aggregated
          }));
          setViewLevel('ADS');
        } else {
          setError(data.message || 'Failed to load Ads');
        }
      } catch (err) {
        console.error('Fetch Ads error:', err);
        setError('API Error: Could not retrieve Ads.');
      } finally {
        setIsLoading(false);
      }
    };

  const goBackToCampaigns = () => {
    setViewLevel('CAMPAIGNS');
    setActiveCampaign(null);
    setActiveAdSet(null);
    fetchMetaOverview(); // Reset top metrics to account level
  };

  const goBackToAdSets = () => {
    setViewLevel('AD_SETS');
    setActiveAdSet(null);
    if (activeCampaign) fetchAdSets(activeCampaign.id, activeCampaign.name);
  };

  useEffect(() => {
    fetchMetaOverview();
  }, [fetchMetaOverview]);

  const handleDateChange = (start: string, end: string) => {
    setStartDate(start);
    setEndDate(end);
    localStorage.setItem('meta_start_date', start);
    localStorage.setItem('meta_end_date', end);
    if (viewLevel === 'CAMPAIGNS') fetchMetaOverview(start, end);
    else if (viewLevel === 'AD_SETS' && activeCampaign) fetchAdSets(activeCampaign.id, activeCampaign.name, start, end);
    else if (viewLevel === 'ADS' && activeAdSet) fetchAds(activeAdSet.id, activeAdSet.name, start, end);
  };

  const formatCurrency = (val: string | number) => {
    const num = typeof val === 'string' ? parseFloat(val) : val;
    return new Intl.NumberFormat('en-IN', { style: 'currency', currency: 'INR' }).format(num || 0);
  };

  const formatBudget = (val: string | number) => {
    const num = typeof val === 'string' ? parseFloat(val) : val;
    // Meta budgets are in currency sub-units (cents/paisa), so we divide by 100
    return formatCurrency(num / 100);
  };

  const formatNumber = (val: string | number) => {
    const num = typeof val === 'string' ? parseFloat(val) : val;
    if (isNaN(num)) return '0';
    return new Intl.NumberFormat('en-IN').format(num || 0);
  };

  // Exception Logic
  const getExceptions = () => {
    const list = [];
    const roas = parseFloat(metaData.metrics.purchase_roas_val || '0');
    const freq = parseFloat(metaData.metrics.frequency || '0');
    
    if (roas < 1.5 && roas > 0) list.push({ type: 'warning', msg: 'Low ROAS Alert: Check performance.', color: '#F59E0B' });
    if (freq > 3.0) list.push({ type: 'danger', msg: 'High Frequency detected: Fatigue alert.', color: '#EF4444' });
    
    return list;
  };

  const exceptions = getExceptions();

  const aggregateMetrics = (insights: any[]) => {
    const agg = {
      spend: '0.00', reach: '0', impressions: '0', inline_link_clicks: '0', 
      ctr: '0.00', cpp: '0.00', purchase_roas: '0.00', frequency: '0.0',
      purchase_value: '0.00', conversions: '0', purchase_roas_val: '0.00'
    };
    if (!insights || insights.length === 0) return agg;
    
    let totalSpend = 0, totalVal = 0, totalConv = 0, totalImpressions = 0, totalReach = 0, totalFreq = 0;
    
    insights.forEach(ins => {
      totalSpend += parseFloat(ins.spend || '0');
      totalVal += parseFloat(ins.purchase_value || '0');
      totalConv += parseInt(ins.conversions || '0');
      totalImpressions += parseInt(ins.impressions || '0');
      totalReach += parseInt(ins.reach || '0');
      // For frequency, we'll use average or specific
      totalFreq = Math.max(totalFreq, parseFloat(ins.frequency || '0')); 
    });
    
    agg.spend = totalSpend.toFixed(2);
    agg.purchase_value = totalVal.toFixed(2);
    agg.conversions = totalConv.toString();
    agg.purchase_roas = totalSpend > 0 ? (totalVal / totalSpend).toFixed(2) : '0.00';
    agg.purchase_roas_val = agg.purchase_roas;
    agg.frequency = totalFreq.toFixed(1);
    
    return agg;
  };

  const currentMetrics = useMemo(() => {
    // If we are at the account level and have a ground-truth summary from Meta, use it.
    if (viewLevel === 'CAMPAIGNS' && metaData.summary && metaData.summary.length > 0) {
      return {
        ...metaData.summary[0],
        purchase_roas_val: metaData.summary[0].purchase_roas_val || metaData.summary[0].purchase_roas // Ensure fallback
      };
    }
    // Otherwise fallback to manual aggregation of the current visible items
    return aggregateMetrics(metaData.insights || []);
  }, [metaData.insights, metaData.summary, viewLevel]);

  // Helper to get insight for a specific object ID
  const getInsightForId = (id: string) => {
    return metaData.insights?.find((ins: any) => ins.ad_id === id || ins.adset_id === id || ins.campaign_id === id);
  };

  return (
    <div className="tab-content-fade" style={{ padding: '0 1rem' }}>
      {/* Header with Date Picker */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem', flexWrap: 'wrap', gap: '1rem', background: 'var(--bg-secondary)', padding: '1.25rem 1.5rem', borderRadius: '24px', border: '1px solid var(--border-color)', backdropFilter: 'blur(10px)' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          <div style={{ background: 'rgba(24, 119, 242, 0.1)', padding: '12px', borderRadius: '16px' }}>
            <svg width="24" height="24" viewBox="0 0 24 24" fill="#1877F2"><path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"/></svg>
          </div>
          <div>
            <h2 style={{ margin: 0, fontSize: '1.25rem', fontWeight: 800, color: 'var(--text-primary)', letterSpacing: '-0.5px' }}>
              Meta Intelligence {userRole === 'admin' && <span style={{ fontSize: '0.65rem', background: 'rgba(59, 130, 246, 0.1)', color: '#60A5FA', padding: '2px 8px', borderRadius: '6px', verticalAlign: 'middle', marginLeft: '8px' }}>Admin</span>}
            </h2>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginTop: '2px' }}>
              <span style={{ fontSize: '0.8rem', color: 'var(--text-tertiary)', fontWeight: 600 }}>{metaData.accounts.find(a => a.id === metaData.activeId)?.name || 'Ad Account'}</span>
              <span style={{ width: '4px', height: '4px', borderRadius: '50%', background: 'var(--text-tertiary)', opacity: 0.3 }}></span>
              <span style={{ fontSize: '0.8rem', color: '#10B981', fontWeight: 700 }}>Real-time</span>
            </div>
          </div>
        </div>
        
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          <CustomDatePicker 
            startDate={startDate} 
            endDate={endDate} 
            onDateChange={handleDateChange} 
          />
          <button 
            className="toolbar-btn" 
            onClick={() => {
              if (viewLevel === 'CAMPAIGNS') fetchMetaOverview();
              else if (viewLevel === 'AD_SETS' && activeCampaign) fetchAdSets(activeCampaign.id, activeCampaign.name);
              else if (viewLevel === 'ADS' && activeAdSet) fetchAds(activeAdSet.id, activeAdSet.name);
            }}
            disabled={isLoading}
            style={{ padding: '10px', borderRadius: '12px' }}
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" className={isLoading ? 'animate-spin' : ''}><path d="M21 2v6h-6"></path><path d="M3 12a9 9 0 0 1 15-6.7L21 8"></path><path d="M3 22v-6h6"></path><path d="M21 12a9 9 0 0 1-15 6.7L3 16"></path></svg>
          </button>
          <button className="btn-secondary" onClick={() => onNavigate?.('settings')} style={{ background: 'rgba(255,255,255,0.05)', padding: '0.65rem 1.25rem', borderRadius: '14px', fontSize: '0.85rem' }}>
            API Settings
          </button>
        </div>
      </div>

      {isLoading && !metaData.metrics.spend ? (
          <div style={{ padding: '8rem', textAlign: 'center' }}>
            <div className="spinner" style={{ margin: '0 auto 2rem', width: '50px', height: '50px', borderTopColor: '#1877F2' }}></div>
            <p style={{ fontWeight: 700, fontSize: '1.1rem', color: 'var(--text-secondary)' }}>Aggregating Meta Performance...</p>
          </div>
      ) : error ? (
        <div className="glass-card" style={{ padding: '4rem', textAlign: 'center', background: 'rgba(239, 68, 68, 0.03)', border: '1px solid rgba(239, 68, 68, 0.15)' }}>
          <div style={{ color: '#ef4444', marginBottom: '1.5rem' }}>
            <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
          </div>
          <h3 style={{ fontSize: '1.5rem', fontWeight: 800, marginBottom: '0.75rem' }}>API Sync Failed</h3>
          <p style={{ color: 'var(--text-tertiary)', marginBottom: '2rem', maxWidth: '500px', margin: '0 auto 2rem' }}>{error}</p>
          <button className="btn-primary" onClick={() => fetchMetaOverview()} style={{ padding: '0.85rem 2rem' }}>Reconnect Meta Marketing</button>
        </div>
      ) : (
        <>
          {/* Period suggestion alert if spend is 0 */}
          {parseFloat(metaData.metrics.spend || '0') === 0 && !isLoading && (
            <div style={{ padding: '1rem 1.5rem', borderRadius: '16px', background: 'rgba(59, 130, 246, 0.05)', border: '1px solid rgba(59, 130, 246, 0.15)', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#60A5FA" strokeWidth="2.5"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
              <span style={{ fontSize: '0.85rem', color: '#93C5FD', fontWeight: 600 }}>No spend detected for this period. Try searching the last 30 days.</span>
            </div>
          )}
          {/* Breadcrumbs & Exceptions */}
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', padding: '0 0.5rem' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
              <span 
                onClick={goBackToCampaigns} 
                style={{ cursor: 'pointer', color: viewLevel === 'CAMPAIGNS' ? 'var(--text-primary)' : 'var(--text-tertiary)', fontWeight: viewLevel === 'CAMPAIGNS' ? 800 : 700, fontSize: '0.85rem' }}
              >
                All Campaigns
              </span>
              {activeCampaign && (
                <>
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="var(--text-tertiary)" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round" style={{ opacity: 0.3 }}><polyline points="9 18 15 12 9 6"></polyline></svg>
                  <span 
                    onClick={goBackToAdSets}
                    style={{ cursor: 'pointer', color: viewLevel === 'AD_SETS' ? 'var(--text-primary)' : 'var(--text-tertiary)', fontWeight: viewLevel === 'AD_SETS' ? 800 : 700, fontSize: '0.85rem' }}
                  >
                    {activeCampaign.name}
                  </span>
                </>
              )}
              {activeAdSet && (
                <>
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="var(--text-tertiary)" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round" style={{ opacity: 0.3 }}><polyline points="9 18 15 12 9 6"></polyline></svg>
                  <span style={{ color: 'var(--text-primary)', fontWeight: 800, fontSize: '0.85rem' }}>{activeAdSet.name}</span>
                </>
              )}
            </div>

            {/* Exception Indicators */}
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              {exceptions.map((ex, i) => (
                <div key={i} title={ex.msg} style={{ width: '10px', height: '10px', borderRadius: '50%', background: ex.color, boxShadow: `0 0 6px ${ex.color}` }}></div>
              ))}
            </div>
          </div>

          {/* Metrics Grid (Dynamic based on level) */}
          <div className="metrics-grid" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: '1rem', marginBottom: '2rem' }}>
            <MetricCard label="Spend" value={formatCurrency(currentMetrics.spend)} sub="Current Period" color="#1877F2" icon={<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><circle cx="12" cy="12" r="10"/><path d="M16 8h-6a2 2 0 1 0 0 4h4a2 2 0 1 1 0 4H8"/><path d="M12 18V6"/></svg>} />
            <MetricCard label="ROAS" value={`${parseFloat(currentMetrics.purchase_roas_val || '0').toFixed(2)}x`} sub="Return on spend" color="#6366F1" icon={<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>} />
            <MetricCard label="Conv." value={formatNumber(currentMetrics.conversions || 0)} sub="Total results" color="#10B981" icon={<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>} />
            <MetricCard label="Freq." value={parseFloat(currentMetrics.frequency || '0').toFixed(1)} sub="Ad Saturation" color="#F59E0B" icon={<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>} />
          </div>

          {/* Drill-down Intelligence Container */}
          <div className="glass-card" style={{ borderRadius: '24px', overflow: 'hidden', border: '1px solid rgba(255, 255, 255, 0.08)', background: 'rgba(17, 24, 39, 0.4)', backdropFilter: 'blur(20px)', boxShadow: '0 20px 50px rgba(0,0,0,0.3)' }}>
            <div style={{ padding: '1.5rem 1.8rem', borderBottom: '1px solid rgba(255, 255, 255, 0.05)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                <div style={{ width: '8px', height: '8px', borderRadius: '50%', background: '#1877F2', boxShadow: '0 0 10px #1877F2' }}></div>
                <h3 style={{ margin: 0, fontSize: '0.85rem', fontWeight: 800, textTransform: 'uppercase', letterSpacing: '1.5px', color: 'var(--text-secondary)' }}>
                  {viewLevel === 'CAMPAIGNS' ? 'Live Campaigns' : viewLevel === 'AD_SETS' ? 'Ad Set Strategy' : 'Ad Creative Performance'}
                </h3>
              </div>
              <div className="badge admin" style={{ fontSize: '0.7rem', fontWeight: 800, padding: '0.4rem 1rem', borderRadius: '12px' }}>
                {viewLevel === 'CAMPAIGNS' ? `${metaData.campaigns.length} Active` : viewLevel === 'AD_SETS' ? `${metaData.adsets.length} Ad Sets` : `${metaData.ads.length} Ads`}
              </div>
            </div>

            <div style={{ overflowX: 'auto' }}>
              <table className="premium-table" style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ background: 'rgba(255,255,255,0.01)' }}>
                    <th style={{ textAlign: 'left', padding: '1.2rem 1.8rem', color: 'var(--text-tertiary)', fontSize: '0.7rem', textTransform: 'uppercase', letterSpacing: '1px', borderBottom: '1px solid rgba(255,255,255,0.03)' }}>
                      {viewLevel === 'CAMPAIGNS' ? 'Campaign Details' : viewLevel === 'AD_SETS' ? 'Ad Set Name' : 'Creative & Preview'}
                    </th>
                    <th style={{ textAlign: 'left', padding: '1.2rem 1.8rem', color: 'var(--text-tertiary)', fontSize: '0.7rem', textTransform: 'uppercase', letterSpacing: '1px', borderBottom: '1px solid rgba(255,255,255,0.03)' }}>Status</th>
                    <th style={{ textAlign: 'left', padding: '1.2rem 1.8rem', color: 'var(--text-tertiary)', fontSize: '0.7rem', textTransform: 'uppercase', letterSpacing: '1px', borderBottom: '1px solid rgba(255,255,255,0.03)' }}>
                      {viewLevel === 'CAMPAIGNS' ? 'Objective' : viewLevel === 'AD_SETS' ? 'Budget' : 'Performance'}
                    </th>
                    {viewLevel === 'ADS' && (
                      <>
                        <th style={{ textAlign: 'left', padding: '1.2rem 1.8rem', color: 'var(--text-tertiary)', fontSize: '0.7rem', textTransform: 'uppercase', letterSpacing: '1px', borderBottom: '1px solid rgba(255,255,255,0.03)' }}>Reach</th>
                        <th style={{ textAlign: 'left', padding: '1.2rem 1.8rem', color: 'var(--text-tertiary)', fontSize: '0.7rem', textTransform: 'uppercase', letterSpacing: '1px', borderBottom: '1px solid rgba(255,255,255,0.03)' }}>Freq.</th>
                        <th style={{ textAlign: 'left', padding: '1.2rem 1.8rem', color: 'var(--text-tertiary)', fontSize: '0.7rem', textTransform: 'uppercase', letterSpacing: '1px', borderBottom: '1px solid rgba(255,255,255,0.03)' }}>Clicks</th>
                      </>
                    )}
                    <th style={{ textAlign: 'right', padding: '1.2rem 1.8rem', color: 'var(--text-tertiary)', fontSize: '0.7rem', textTransform: 'uppercase', letterSpacing: '1px', borderBottom: '1px solid rgba(255,255,255,0.03)' }}>Action</th>
                  </tr>
                </thead>
                <tbody>
                  {viewLevel === 'CAMPAIGNS' && metaData.campaigns.length > 0 ? metaData.campaigns.map(camp => {
                    const insight = getInsightForId(camp.id);
                    return (
                      <tr key={camp.id} className="hover-row" style={{ cursor: 'pointer' }}>
                        <td onClick={() => fetchAdSets(camp.id, camp.name)} style={{ padding: '1.4rem 1.8rem', borderBottom: '1px solid rgba(255,255,255,0.02)' }}>
                          <div style={{ fontWeight: 700, color: '#F9FAFB', fontSize: '1rem', marginBottom: '2px' }}>{camp.name}</div>
                          <div style={{ fontSize: '0.7rem', color: '#9CA3AF', fontFamily: 'var(--font-mono)' }}>ID: {camp.id}</div>
                        </td>
                        <td onClick={() => fetchAdSets(camp.id, camp.name)} style={{ padding: '1.4rem 1.8rem', borderBottom: '1px solid rgba(255,255,255,0.02)' }}>
                          <span style={{ fontSize: '0.7rem', fontWeight: 800, color: camp.effective_status === 'ACTIVE' ? '#10B981' : '#F59E0B' }}>{camp.effective_status}</span>
                        </td>
                        <td onClick={() => fetchAdSets(camp.id, camp.name)} style={{ padding: '1.4rem 1.8rem', borderBottom: '1px solid rgba(255,255,255,0.02)' }}>
                          <div style={{ fontWeight: 600, fontSize: '0.85rem' }}>{camp.objective.replace(/_/g, ' ')}</div>
                        </td>
                        <td onClick={() => fetchAdSets(camp.id, camp.name)} style={{ padding: '1.4rem 1.8rem', borderBottom: '1px solid rgba(255,255,255,0.02)', textAlign: 'right' }}>
                          <span style={{ fontWeight: 700, color: '#F9FAFB' }}>{formatCurrency(insight?.spend || 0)}</span>
                        </td>
                        <td style={{ padding: '1.4rem 1.8rem', textAlign: 'right', borderBottom: '1px solid rgba(255,255,255,0.02)' }}>
                           <a href={`https://adsmanager.facebook.com/adsmanager/manage/campaigns?act=${metaData.activeId}&selected_campaign_ids=${camp.id}`} target="_blank" rel="noreferrer" onClick={(e) => e.stopPropagation()}>
                             <button className="toolbar-btn" style={{ padding: '8px' }} title="Change in Meta Dashboard">
                               <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#9CA3AF" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path><polyline points="15 3 21 3 21 9"></polyline><line x1="10" y1="14" x2="21" y2="3"></line></svg>
                             </button>
                           </a>
                        </td>
                      </tr>
                    );
                  }) : viewLevel === 'CAMPAIGNS' && (
                    <tr>
                      <td colSpan={4} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-tertiary)' }}>
                        <p style={{ fontWeight: 700 }}>No Campaigns Found</p>
                        <p style={{ fontSize: '0.8rem' }}>Check your Meta Ad Account permissions or selected account.</p>
                      </td>
                    </tr>
                  )}

                  {viewLevel === 'AD_SETS' && metaData.adsets.map(adset => {
                    const insight = getInsightForId(adset.id);
                    
                    // Budget Fallback: If adset budget is 0, it might be a CBO campaign
                    let displayBudget = adset.daily_budget || adset.lifetime_budget || '0';
                    let isCBO = false;
                    if (parseFloat(displayBudget) === 0 && activeCampaign) {
                      const parentCamp = metaData.campaigns.find(c => c.id === activeCampaign.id);
                      if (parentCamp && (parentCamp.daily_budget || parentCamp.lifetime_budget)) {
                        displayBudget = parentCamp.daily_budget || parentCamp.lifetime_budget;
                        isCBO = true;
                      }
                    }

                    return (
                      <tr key={adset.id} className="hover-row" style={{ cursor: 'pointer' }}>
                        <td onClick={() => fetchAds(adset.id, adset.name)} style={{ padding: '1.4rem 1.8rem', borderBottom: '1px solid rgba(255,255,255,0.02)' }}>
                          <div style={{ fontWeight: 700, color: '#F9FAFB', fontSize: '1rem', marginBottom: '2px' }}>{adset.name}</div>
                          <div style={{ fontSize: '0.7rem', color: '#9CA3AF' }}>ID: {adset.id}</div>
                        </td>
                        <td onClick={() => fetchAds(adset.id, adset.name)} style={{ padding: '1.4rem 1.8rem', borderBottom: '1px solid rgba(255,255,255,0.02)' }}>
                          <span style={{ fontSize: '0.7rem', fontWeight: 800, color: adset.effective_status === 'ACTIVE' ? '#10B981' : '#F59E0B' }}>{adset.effective_status}</span>
                        </td>
                        <td onClick={() => fetchAds(adset.id, adset.name)} style={{ padding: '1.4rem 1.8rem', borderBottom: '1px solid rgba(255,255,255,0.02)' }}>
                          <div style={{ fontWeight: 600, fontSize: '0.9rem', color: '#60A5FA' }}>{formatBudget(displayBudget)}</div>
                          <div style={{ fontSize: '0.65rem', color: '#6B7280' }}>{isCBO ? 'Advantage Campaign Budget' : `Budget ${adset.bid_strategy?.replace(/_/g, ' ') || 'Lowest Cost'}`}</div>
                        </td>
                        <td onClick={() => fetchAds(adset.id, adset.name)} style={{ padding: '1.4rem 1.8rem', borderBottom: '1px solid rgba(255,255,255,0.02)', textAlign: 'right' }}>
                          <span style={{ fontWeight: 700, color: '#F9FAFB' }}>{formatCurrency(insight?.spend || 0)}</span>
                        </td>
                        <td onClick={() => fetchAds(adset.id, adset.name)} style={{ padding: '1.4rem 1.8rem', textAlign: 'right', borderBottom: '1px solid rgba(255,255,255,0.02)' }}>
                          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#6B7280" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"><polyline points="9 18 15 12 9 6"></polyline></svg>
                        </td>
                      </tr>
                    );
                  })}

                  {viewLevel === 'ADS' && metaData.ads.map(ad => {
                    const insight = getInsightForId(ad.id);
                    return (
                      <tr key={ad.id} className="hover-row" onClick={() => { setSelectedAd(ad); setIsModalOpen(true); }} style={{ cursor: 'pointer' }}>
                        <td style={{ padding: '1.2rem 1.8rem', borderBottom: '1px solid rgba(255,255,255,0.02)' }}>
                          <div style={{ display: 'flex', alignItems: 'center', gap: '1.25rem' }}>
                            {ad.creative?.thumbnail_url ? (
                              <img 
                                src={ad.creative.thumbnail_url} 
                                alt="Creative" 
                                style={{ width: '80px', height: '80px', borderRadius: '12px', objectFit: 'cover', border: '2px solid rgba(255,255,255,0.1)', boxShadow: '0 4px 12px rgba(0,0,0,0.2)' }} 
                              />
                            ) : (
                              <div style={{ width: '80px', height: '80px', borderRadius: '12px', background: 'rgba(255,255,255,0.05)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                                 <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,0.1)" strokeWidth="2"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>
                              </div>
                            )}
                            <div>
                              <div style={{ fontWeight: 700, color: '#F9FAFB', fontSize: '0.95rem', marginBottom: '4px' }}>{ad.name}</div>
                              <div style={{ fontSize: '0.7rem', color: '#9CA3AF' }}>Ad ID: {ad.id}</div>
                            </div>
                          </div>
                        </td>
                        <td style={{ padding: '1.2rem 1.8rem', borderBottom: '1px solid rgba(255,255,255,0.02)' }}>
                           <span style={{ fontSize: '0.7rem', fontWeight: 800, color: ad.effective_status === 'ACTIVE' ? '#10B981' : '#F59E0B' }}>{ad.effective_status}</span>
                        </td>
                        <td style={{ padding: '1.2rem 1.8rem', borderBottom: '1px solid rgba(255,255,255,0.02)' }}>
                          {insight ? (
                            <>
                              <div style={{ fontWeight: 800, fontSize: '1.1rem', color: parseFloat(insight.purchase_roas_val) > 2 ? '#10B981' : '#F9FAFB' }}>
                                {parseFloat(insight.purchase_roas_val || '0').toFixed(2)}x ROAS
                              </div>
                              <div style={{ fontSize: '0.7rem', color: '#9CA3AF' }}>Spend: {formatCurrency(insight.spend)}</div>
                            </>
                          ) : (
                            <div style={{ fontSize: '0.75rem', color: '#4B5563', fontStyle: 'italic' }}>No Insights...</div>
                          )}
                        </td>
                        {viewLevel === 'ADS' && (
                          <>
                            <td style={{ padding: '1.2rem 1.8rem', borderBottom: '1px solid rgba(255,255,255,0.02)' }}>
                              <div style={{ fontWeight: 700, color: '#F9FAFB' }}>{formatNumber(insight?.reach || 0)}</div>
                              <div style={{ fontSize: '0.6rem', color: '#9CA3AF' }}>{formatNumber(insight?.impressions || 0)} Imp.</div>
                            </td>
                            <td style={{ padding: '1.2rem 1.8rem', borderBottom: '1px solid rgba(255,255,255,0.02)' }}>
                              <div style={{ fontWeight: 700, color: '#F9FAFB' }}>{parseFloat(insight?.frequency || '0').toFixed(2)}</div>
                            </td>
                            <td style={{ padding: '1.2rem 1.8rem', borderBottom: '1px solid rgba(255,255,255,0.02)' }}>
                              <div style={{ fontWeight: 700, color: '#F9FAFB' }}>{formatNumber(insight?.inline_link_clicks || 0)}</div>
                              <div style={{ fontSize: '0.6rem', color: '#9CA3AF' }}>{insight?.ctr ? `${parseFloat(insight.ctr).toFixed(2)}% CTR` : '0% CTR'}</div>
                            </td>
                          </>
                        )}
                        <td style={{ padding: '1.2rem 1.8rem', textAlign: 'right', borderBottom: '1px solid rgba(255,255,255,0.02)' }}>
                           <a href={`https://adsmanager.facebook.com/adsmanager/manage/ads?act=${metaData.activeId}&selected_ad_ids=${ad.id}`} target="_blank" rel="noreferrer">
                             <button className="toolbar-btn" style={{ padding: '8px' }} title="Edit in Meta Dashboard">
                               <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#9CA3AF" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
                             </button>
                           </a>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}

      {/* Ad Detail Modal */}
      {isModalOpen && selectedAd && (
        <div 
          onClick={() => setIsModalOpen(false)}
          style={{ 
            position: 'fixed', top: 0, left: 0, width: '100vw', height: '100vh', 
            background: 'rgba(0,0,0,0.85)', backdropFilter: 'blur(10px)', 
            zIndex: 1000, display: 'flex', alignItems: 'center', justifyContent: 'center' 
          }}
        >
          <div 
            onClick={(e) => e.stopPropagation()}
            style={{ 
              width: '900px', maxHeight: '90vh', background: '#111827', 
              borderRadius: '24px', border: '1px solid rgba(255,255,255,0.08)',
              padding: '2.5rem', overflowY: 'auto'
            }}
          >
            {/* Modal Header */}
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start', marginBottom: '2.5rem' }}>
              <div style={{ display: 'flex', gap: '1.5rem', alignItems: 'center' }}>
                <img 
                  src={selectedAd.creative?.thumbnail_url || 'https://via.placeholder.com/100'} 
                  alt="Ad" 
                  style={{ width: '100px', height: '100px', borderRadius: '16px', objectFit: 'cover' }}
                />
                <div>
                  <h2 style={{ fontSize: '1.5rem', fontWeight: 800, color: '#F9FAFB', marginBottom: '0.5rem' }}>{selectedAd.name}</h2>
                  <div style={{ display: 'flex', gap: '0.8rem', alignItems: 'center' }}>
                    <span style={{ background: 'rgba(96, 165, 250, 0.1)', color: '#60A5FA', fontSize: '0.7rem', fontWeight: 800, padding: '4px 10px', borderRadius: '8px' }}>
                      {selectedAd.status}
                    </span>
                    <span style={{ color: '#6B7280', fontSize: '0.75rem' }}>ID: {selectedAd.id}</span>
                  </div>
                </div>
              </div>
              <button 
                onClick={() => setIsModalOpen(false)}
                style={{ background: 'rgba(255,255,255,0.05)', border: 'none', borderRadius: '50%', width: '40px', height: '40px', color: '#fff', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
              >
                ✕
              </button>
            </div>

            {/* Metrics Content */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '2rem' }}>
              
              {/* Performance Section */}
              <div className="modal-section" style={{ background: 'rgba(255,255,255,0.02)', padding: '1.5rem', borderRadius: '20px', border: '1px solid rgba(255,255,255,0.05)' }}>
                <h3 style={{ fontSize: '0.8rem', color: '#6B7280', textTransform: 'uppercase', marginBottom: '1.5rem', letterSpacing: '1px' }}>Core Financials</h3>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.5rem' }}>
                  <div>
                    <div style={{ fontSize: '0.65rem', color: '#9CA3AF', marginBottom: '4px' }}>SPEND</div>
                    <div style={{ fontSize: '1.1rem', fontWeight: 800, color: '#F9FAFB' }}>{formatCurrency(getInsightForId(selectedAd.id)?.spend || 0)}</div>
                  </div>
                  <div>
                    <div style={{ fontSize: '0.65rem', color: '#9CA3AF', marginBottom: '4px' }}>ROAS</div>
                    <div style={{ fontSize: '1.1rem', fontWeight: 800, color: '#10B981' }}>{parseFloat(getInsightForId(selectedAd.id)?.purchase_roas_val || '0').toFixed(2)}x</div>
                  </div>
                  <div>
                    <div style={{ fontSize: '0.65rem', color: '#9CA3AF', marginBottom: '4px' }}>CPC</div>
                    <div style={{ fontSize: '1.1rem', fontWeight: 800, color: '#F9FAFB' }}>{formatCurrency(getInsightForId(selectedAd.id)?.cpc || 0)}</div>
                  </div>
                  <div>
                    <div style={{ fontSize: '0.65rem', color: '#9CA3AF', marginBottom: '4px' }}>CPM</div>
                    <div style={{ fontSize: '1.1rem', fontWeight: 800, color: '#F9FAFB' }}>{formatCurrency(getInsightForId(selectedAd.id)?.cpm || 0)}</div>
                  </div>
                </div>
              </div>

              {/* Quality & Visibility */}
              <div className="modal-section" style={{ background: 'rgba(255,255,255,0.02)', padding: '1.5rem', borderRadius: '20px', border: '1px solid rgba(255,255,255,0.05)' }}>
                <h3 style={{ fontSize: '0.8rem', color: '#6B7280', textTransform: 'uppercase', marginBottom: '1.5rem', letterSpacing: '1px' }}>Quality & Delivery</h3>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.5rem' }}>
                  <div>
                    <div style={{ fontSize: '0.65rem', color: '#9CA3AF', marginBottom: '4px' }}>FREQUENCY</div>
                    <div style={{ fontSize: '1.1rem', fontWeight: 800, color: '#F9FAFB' }}>{parseFloat(getInsightForId(selectedAd.id)?.frequency || '0').toFixed(2)}</div>
                  </div>
                  <div>
                    <div style={{ fontSize: '0.65rem', color: '#9CA3AF', marginBottom: '4px' }}>REACH</div>
                    <div style={{ fontSize: '1.1rem', fontWeight: 800, color: '#F9FAFB' }}>{formatNumber(getInsightForId(selectedAd.id)?.reach || 0)}</div>
                  </div>
                  <div>
                    <div style={{ fontSize: '0.65rem', color: '#9CA3AF', marginBottom: '4px' }}>QUALITY RANK</div>
                    <div style={{ fontSize: '0.8rem', fontWeight: 800, color: '#60A5FA' }}>{getInsightForId(selectedAd.id)?.quality_ranking?.replace(/_/g, ' ') || 'AVERAGE'}</div>
                  </div>
                  <div>
                    <div style={{ fontSize: '0.65rem', color: '#9CA3AF', marginBottom: '4px' }}>CONV. RANK</div>
                    <div style={{ fontSize: '0.8rem', fontWeight: 800, color: '#60A5FA' }}>{getInsightForId(selectedAd.id)?.conversion_rate_ranking?.replace(/_/g, ' ') || 'AVERAGE'}</div>
                  </div>
                </div>
              </div>

              {/* Social Engagement */}
              <div className="modal-section" style={{ background: 'rgba(255,255,255,0.02)', padding: '1.5rem', borderRadius: '20px', border: '1px solid rgba(255,255,255,0.05)' }}>
                <h3 style={{ fontSize: '0.8rem', color: '#6B7280', textTransform: 'uppercase', marginBottom: '1.5rem', letterSpacing: '1px' }}>Social Engagement</h3>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.5rem' }}>
                  <div>
                    <div style={{ fontSize: '0.65rem', color: '#9CA3AF', marginBottom: '4px' }}>CLICKS</div>
                    <div style={{ fontSize: '1.1rem', fontWeight: 800, color: '#F9FAFB' }}>{formatNumber(getInsightForId(selectedAd.id)?.inline_link_clicks || 0)}</div>
                  </div>
                  <div>
                    <div style={{ fontSize: '0.65rem', color: '#9CA3AF', marginBottom: '4px' }}>CTR</div>
                    <div style={{ fontSize: '1.1rem', fontWeight: 800, color: '#F9FAFB' }}>{parseFloat(getInsightForId(selectedAd.id)?.ctr || '0').toFixed(2)}%</div>
                  </div>
                  <div>
                    <div style={{ fontSize: '0.65rem', color: '#9CA3AF', marginBottom: '4px' }}>REACTIONS</div>
                    <div style={{ fontSize: '1.1rem', fontWeight: 800, color: '#F9FAFB' }}>{formatNumber(getInsightForId(selectedAd.id)?.post_reaction_val || 0)}</div>
                  </div>
                  <div>
                    <div style={{ fontSize: '0.65rem', color: '#9CA3AF', marginBottom: '4px' }}>COMMENTS</div>
                    <div style={{ fontSize: '1.1rem', fontWeight: 800, color: '#F9FAFB' }}>{formatNumber(getInsightForId(selectedAd.id)?.post_comment_val || 0)}</div>
                  </div>
                </div>
              </div>

              {/* Video Funnel */}
              <div className="modal-section" style={{ background: 'rgba(255,255,255,0.02)', padding: '1.5rem', borderRadius: '20px', border: '1px solid rgba(255,255,255,0.05)' }}>
                <h3 style={{ fontSize: '0.8rem', color: '#6B7280', textTransform: 'uppercase', marginBottom: '1.5rem', letterSpacing: '1px' }}>Video Funnel</h3>
                <div style={{ borderLeft: '2px solid rgba(255,255,255,0.05)', paddingLeft: '1.5rem' }}>
                  <div style={{ marginBottom: '1rem' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.7rem', color: '#9CA3AF', marginBottom: '5px' }}>
                      <span>25% Watched</span>
                      <span style={{ fontWeight: 800, color: '#F9FAFB' }}>{formatNumber(getInsightForId(selectedAd.id)?.video_p25_val || 0)}</span>
                    </div>
                    <div style={{ height: '4px', background: '#374151', borderRadius: '2px' }}>
                      <div style={{ height: '100%', width: '100%', background: '#60A5FA', borderRadius: '2px' }}></div>
                    </div>
                  </div>
                  <div style={{ marginBottom: '1rem' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.7rem', color: '#9CA3AF', marginBottom: '5px' }}>
                      <span>50% Watched</span>
                      <span style={{ fontWeight: 800, color: '#F9FAFB' }}>{formatNumber(getInsightForId(selectedAd.id)?.video_p50_val || 0)}</span>
                    </div>
                    <div style={{ height: '4px', background: '#374151', borderRadius: '2px' }}>
                      <div style={{ height: '100%', width: '50%', background: '#60A5FA', borderRadius: '2px' }}></div>
                    </div>
                  </div>
                  <div>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.7rem', color: '#9CA3AF', marginBottom: '5px' }}>
                      <span>100% Watched</span>
                      <span style={{ fontWeight: 800, color: '#F9FAFB' }}>{formatNumber(getInsightForId(selectedAd.id)?.video_p100_val || 0)}</span>
                    </div>
                    <div style={{ height: '4px', background: '#374151', borderRadius: '2px' }}>
                      <div style={{ height: '100%', width: '20%', background: '#60A5FA', borderRadius: '2px' }}></div>
                    </div>
                  </div>
                </div>
              </div>

            </div>
          </div>
        </div>
      )}
    </div>
  );
};

const MetricCard = ({ label, value, sub, icon, color }: any) => (
  <div className="glass-card" style={{ padding: '1.5rem', borderTop: `4px solid ${color}`, display: 'flex', flexDirection: 'column', gap: '0.75rem', background: 'var(--bg-secondary)' }}>
    <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
      <div style={{ color: color }}>{icon}</div>
      <span style={{ fontSize: '0.75rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.5px' }}>{label}</span>
    </div>
    <div style={{ fontSize: '1.5rem', fontWeight: 900, color: 'var(--text-primary)' }}>{value}</div>
    <div style={{ fontSize: '0.65rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>{sub}</div>
  </div>
);
