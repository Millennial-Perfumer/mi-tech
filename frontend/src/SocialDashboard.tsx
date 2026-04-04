import React, { useState, useEffect } from 'react';
import { API_BASE } from './api';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { SocialComposerModal } from './SocialComposerModal';
import { CustomDatePicker } from './CustomDatePicker';
import { PostInsightsModal } from './PostInsightsModal';

interface SocialDashboardProps {
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
  userRole: string;
  startDate: string;
  endDate: string;
  onUpdateDateRange: (start: string, end: string) => void;
}

interface SocialMetrics {
  account?: {
    follower_count: number;
    total_reach: number;
    total_views: number;
  };
  total_reach: number;
  total_views: number;
  total_engagement: number;
  posts: Array<{
    id: string;
    content: string;
    media_url: string;
    thumbnail_url: string;
    permalink: string;
    published_at: string;
    reach: number;
    views: number;
    engagement: number;
    restricted?: boolean;
  }>;
}

export const SocialDashboard: React.FC<SocialDashboardProps> = ({ 
  fetchWithAuth, 
  userRole, 
  startDate, 
  endDate, 
  onUpdateDateRange 
}) => {
  const [activePlatform, setActivePlatform] = useState<'instagram' | 'facebook' | 'threads'>('instagram');
  const [metrics, setMetrics] = useState<SocialMetrics | null>(null);
  const [historyData, setHistoryData] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isComposerOpen, setIsComposerOpen] = useState(false);
  const [selectedPostForInsights, setSelectedPostForInsights] = useState<any>(null);
  const [isInsightsModalOpen, setIsInsightsModalOpen] = useState(false);
  const [fetchError, setFetchError] = useState<string | null>(null);

  const fetchMetrics = async () => {
    setIsLoading(true);
    try {
      const response = await fetchWithAuth(
        `${API_BASE}/api/marketing/smm/overview?platform=${activePlatform}&start_date=${startDate}&end_date=${endDate}`
      );
      const data = await response.json();
      if (data.success) {
        setMetrics(data.overview);
        setFetchError(null);
        // Map the posts into history data for the chart (if available)
        if (data.overview.posts && data.overview.posts.length > 0) {
          const sorted = [...data.overview.posts].sort((a, b) => 
            new Date(a.published_at).getTime() - new Date(b.published_at).getTime()
          );
          setHistoryData(sorted.map(p => ({
            date: new Date(p.published_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
            engagement: p.engagement,
            views: p.views,
          })));
        } else {
          setHistoryData(generateMockHistory(startDate, endDate));
        }
      } else {
        setFetchError(data.error || 'Failed to fetch metrics from Meta');
        setMetrics(null);
        setHistoryData(generateMockHistory(startDate, endDate));
      }
    } catch (err) {
      console.error('Failed to fetch social metrics:', err);
    } finally {
      setIsLoading(false);
    }
  };

  const syncPulse = async () => {
    try {
      // In No-DB mode, sync simply triggers a refresh of the live fetch
      await fetchMetrics();
    } catch (err) {
      console.error('Refresh failed:', err);
    }
  };

  useEffect(() => {
    // Clear metrics when switching platforms to avoid ghosting
    setMetrics(null);
    setFetchError(null);
    fetchMetrics();
  }, [activePlatform, startDate, endDate]);

  const generateMockHistory = (start: string, end: string) => {
    const data = [];
    const s = new Date(start);
    const e = new Date(end);
    // Limit to 30 days if range is too large to avoid chart clutter
    const diff = Math.min(30, Math.ceil((e.getTime() - s.getTime()) / (1000 * 3600 * 24)));
    for (let i = 0; i <= diff; i++) {
      const d = new Date(s);
      d.setDate(s.getDate() + i);
      data.push({
        date: d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
        engagement: 0,
        views: 0,
      });
    }
    return data;
  };

  const getPlatformColor = (p: string) => {
    switch (p) {
      case 'instagram': return '#E1306C';
      case 'facebook': return '#1877F2';
      case 'threads': return '#000000';
      default: return '#10b981';
    }
  };

  return (
    <div className="social-dashboard-container" style={{ 
      animation: 'fadeIn 0.6s ease-out',
      fontFamily: "'Montserrat', sans-serif",
      color: 'white'
    }}>
      {/* Header Island with Platform Switcher and Date Picker */}
      <div className="glass-island" style={{
        background: 'rgba(255, 255, 255, 0.03)',
        backdropFilter: 'blur(40px)',
        WebkitBackdropFilter: 'blur(40px)',
        borderRadius: '32px',
        border: '1px solid rgba(255, 255, 255, 0.12)',
        padding: '1.25rem 2rem',
        marginBottom: '2.5rem',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.5)',
        transition: 'all 0.4s cubic-bezier(0.4, 0, 0.2, 1)'
      }}>
        <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
          {(['instagram', 'facebook', 'threads'] as const).map(p => (
            <button
              key={p}
              onClick={() => setActivePlatform(p)}
              style={{
                background: activePlatform === p ? getPlatformColor(p) : 'rgba(255,255,255,0.03)',
                color: activePlatform === p ? 'white' : 'rgba(255,255,255,0.4)',
                border: activePlatform === p ? `1px solid ${getPlatformColor(p)}` : '1px solid rgba(255,255,255,0.05)',
                padding: '0.6rem 1.4rem',
                borderRadius: '14px',
                fontWeight: 700,
                cursor: 'pointer',
                transition: 'all 0.3s cubic-bezier(0.16, 1, 0.3, 1)',
                textTransform: 'capitalize',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                fontSize: '0.9rem'
              }}
            >
              {p}
            </button>
          ))}
          
          <div style={{ width: '1px', height: '24px', background: 'rgba(255,255,255,0.1)', margin: '0 0.5rem' }} />
          
          {/* Standard Date Picker */}
          <CustomDatePicker 
            startDate={startDate} 
            endDate={endDate} 
            onDateChange={onUpdateDateRange} 
          />
        </div>

        <div style={{ display: 'flex', gap: '1rem' }}>
          <button
            onClick={syncPulse}
            disabled={isLoading}
            className="glass-btn"
            style={{
              background: 'rgba(255, 255, 255, 0.05)',
              border: '1px solid rgba(255, 255, 255, 0.1)',
              color: 'white',
              padding: '0.6rem 1.25rem',
              borderRadius: '12px',
              fontWeight: 600,
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: '0.6rem',
              transition: 'all 0.3s'
            }}
          >
            <svg 
              width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"
              style={{ animation: isLoading ? 'spin 1.5s linear infinite' : 'none' }}
            >
              <path d="M21 2v6h-6M3 22v-6h6M2 12c0-4.418 3.582-8 8-8 3.328 0 6.223 2.038 7.425 4.928M22 12c0 4.418-3.582 8-8 8-3.328 0-6.223-2.038-7.425-4.928" />
            </svg>
            Refresh Pulse
          </button>

          {userRole === 'admin' && (
            <button 
              className="glass-btn primary"
              onClick={() => setIsComposerOpen(true)}
              style={{
                background: 'linear-gradient(135deg, #10b981, #059669)',
                border: 'none',
                color: 'white',
                padding: '0.6rem 1.25rem',
                borderRadius: '12px',
                fontWeight: 800,
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                boxShadow: '0 10px 20px rgba(16, 185, 129, 0.2)'
              }}
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
              Compose
            </button>
          )}
        </div>
      </div>

      {/* Error Alert Island */}
      {fetchError && (
        <div className="glass-island" style={{
          background: 'rgba(239, 68, 68, 0.08)',
          backdropFilter: 'blur(40px)',
          WebkitBackdropFilter: 'blur(40px)',
          borderRadius: '24px',
          border: '1px solid rgba(239, 68, 68, 0.25)',
          padding: '1.25rem 2rem',
          marginBottom: '2rem',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          animation: 'slideDown 0.5s ease-out',
          boxShadow: '0 15px 30px -10px rgba(239, 68, 68, 0.15)'
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '1.25rem' }}>
            <div style={{ 
              width: '44px', height: '44px', borderRadius: '14px', 
              background: 'rgba(239, 68, 68, 0.15)', display: 'flex', alignItems: 'center', 
              justifyContent: 'center', fontSize: '1.2rem',
              border: '1px solid rgba(239, 68, 68, 0.3)'
            }}>
              ⚠️
            </div>
            <div>
              <div style={{ fontSize: '0.7rem', color: '#fca5a5', fontWeight: 800, textTransform: 'uppercase', letterSpacing: '0.15em', marginBottom: '0.2rem' }}>
                Platform Access Restricted
              </div>
              <div style={{ fontSize: '0.95rem', fontWeight: 600, color: 'rgba(255,255,255,0.9)', marginBottom: '0.2rem' }}>
                {fetchError}
              </div>
              <div style={{ fontSize: '0.75rem', opacity: 0.6 }}>
                Meta requires an upgrade of your permissions. <a href="https://developers.facebook.com/tools/debug/accesstoken/" target="_blank" rel="noopener noreferrer" style={{ color: '#fecaca', textDecoration: 'underline' }}>Check Token Scopes</a>
              </div>
            </div>
          </div>
          <div style={{ display: 'flex', gap: '1rem' }}>
            <button
              onClick={() => fetchMetrics()}
              style={{
                background: 'rgba(239, 68, 68, 0.15)',
                border: '1px solid rgba(239, 68, 68, 0.3)',
                color: '#fecaca',
                padding: '0.6rem 1.25rem',
                borderRadius: '12px',
                fontWeight: 700,
                cursor: 'pointer',
                fontSize: '0.8rem',
                transition: 'all 0.3s'
              }}
            >
              Retry Connection
            </button>
          </div>
        </div>
      )}

      {/* Account Overall KPIs (Lifetime/Total) */}
      {activePlatform === 'instagram' && metrics?.account && (
        <div style={{ 
          display: 'grid', 
          gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', 
          gap: '1rem', 
          marginBottom: '2rem',
          padding: '1.5rem',
          background: 'rgba(255, 255, 255, 0.02)',
          borderRadius: '24px',
          border: '1px solid rgba(255, 255, 255, 0.05)'
        }}>
          {[
            { label: 'Followers', value: metrics.account.follower_count, icon: '👥', color: '#ff006e' },
            { label: 'Growth Reach', value: metrics.account.total_reach, icon: '📈', color: '#8338ec' },
            { label: 'Direct Impressions', value: metrics.account.total_views, icon: '🌊', color: '#3a86ff' }
          ].map(m => (
            <div key={m.label} style={{ 
              display: 'flex', 
              alignItems: 'center', 
              gap: '1.25rem',
              padding: '0.5rem 1rem',
              borderRadius: '16px',
              transition: 'background 0.3s ease'
            }}>
              <div style={{ 
                width: '44px', height: '44px', borderRadius: '14px', 
                background: `${m.color}15`, display: 'flex', alignItems: 'center', 
                justifyContent: 'center', fontSize: '1.3rem',
                border: `1px solid ${m.color}30`,
                boxShadow: `0 8px 16px ${m.color}10`
              }}>
                {m.icon}
              </div>
              <div>
                <div style={{ fontSize: '0.65rem', color: 'rgba(255,255,255,0.4)', fontWeight: 800, textTransform: 'uppercase', letterSpacing: '0.1em', marginBottom: '0.2rem' }}>{m.label}</div>
                <div style={{ fontSize: '1.4rem', fontWeight: 900, color: 'white', letterSpacing: '-0.02em' }}>{m.value.toLocaleString()}</div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Main KPI Grid (Period-specific) */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '1.5rem', marginBottom: '2.5rem' }}>
        {[
          { label: 'Total Engagement', value: metrics?.total_engagement || 0, icon: '⚡', color: '#fbbf24', trend: 'Stable' },
          { label: 'Unique Reach', value: metrics?.total_reach || 0, icon: '🎯', color: '#10b981', trend: 'Syncing' },
          { label: 'Total Visibility', value: metrics?.total_views || 0, icon: '👁️', color: '#3b82f6', trend: 'Live' },
          { label: 'Period Content', value: metrics?.posts?.length || 0, icon: '💎', color: '#a855f7', trend: 'Active' }
        ].map(m => (
          <div key={m.label} className="metric-card-glass" style={{
            background: 'rgba(255, 255, 255, 0.02)',
            backdropFilter: 'blur(30px)',
            WebkitBackdropFilter: 'blur(30px)',
            borderRadius: '28px',
            padding: '1.75rem 2rem',
            border: '1px solid rgba(255, 255, 255, 0.08)',
            position: 'relative',
            overflow: 'hidden',
            boxShadow: '0 20px 40px -15px rgba(0, 0, 0, 0.3)',
            transition: 'transform 0.3s ease'
          }}>
            <div style={{ position: 'absolute', top: '15px', right: '15px', padding: '0.3rem 0.6rem', borderRadius: '8px', background: 'rgba(255,255,255,0.05)', fontSize: '0.65rem', fontWeight: 800, color: 'rgba(255,255,255,0.4)', textTransform: 'uppercase' }}>
              {m.trend}
            </div>
            <div style={{ position: 'absolute', bottom: '-20px', right: '-10px', fontSize: '6rem', opacity: 0.03 }}>{m.icon}</div>
            <div style={{ color: 'rgba(255,255,255,0.4)', fontSize: '0.75rem', fontWeight: 800, letterSpacing: '0.08em', textTransform: 'uppercase', marginBottom: '0.5rem' }}>{m.label}</div>
            <div style={{ fontSize: '2.8rem', fontWeight: 900, letterSpacing: '-0.04em', color: 'white', textShadow: `0 10px 20px ${m.color}20` }}>
              {isLoading ? (
                <span style={{ opacity: 0.3 }}>--</span>
              ) : (
                m.value.toLocaleString()
              )}
            </div>
            <div style={{ width: '30px', height: '4px', background: m.color, borderRadius: '2px', marginTop: '1rem', opacity: 0.8 }} />
          </div>
        ))}
      </div>

      {/* Chart and Post Table */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr', gap: '2rem', marginBottom: '2rem' }}>
        <div className="glass-card chart-card" style={{
          background: 'rgba(25, 25, 25, 0.4)',
          backdropFilter: 'blur(40px)',
          borderRadius: '32px',
          padding: '2.5rem',
          border: '1px solid rgba(255, 255, 255, 0.1)',
          height: '450px',
          boxShadow: '0 30px 60px rgba(0,0,0,0.3)'
        }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '2rem' }}>
            <h3 style={{ fontSize: '1.5rem', fontWeight: 800 }}>Engagement Pulse</h3>
            <div style={{ fontSize: '0.8rem', opacity: 0.5 }}>Real-time Meta Snapshot</div>
          </div>
          <ResponsiveContainer width="100%" height="80%">
            <AreaChart data={historyData}>
              <defs>
                <linearGradient id="colorEng" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor={getPlatformColor(activePlatform)} stopOpacity={0.4}/>
                  <stop offset="60%" stopColor={getPlatformColor(activePlatform)} stopOpacity={0.1}/>
                  <stop offset="95%" stopColor={getPlatformColor(activePlatform)} stopOpacity={0}/>
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="0" vertical={false} stroke="rgba(255,255,255,0.02)" />
              <XAxis 
                dataKey="date" 
                stroke="rgba(255,255,255,0.2)" 
                fontSize={10} 
                tickLine={false} 
                axisLine={false} 
                dy={15} 
                interval="preserveStartEnd"
              />
              <YAxis 
                stroke="rgba(255,255,255,0.2)" 
                fontSize={10} 
                tickLine={false} 
                axisLine={false} 
                dx={-15} 
              />
              <Tooltip 
                cursor={{ stroke: 'rgba(255,255,255,0.1)', strokeWidth: 1 }}
                contentStyle={{ 
                  background: 'rgba(20, 20, 20, 0.8)', 
                  border: '1px solid rgba(255,255,255,0.1)', 
                  borderRadius: '16px', 
                  backdropFilter: 'blur(20px)', 
                  WebkitBackdropFilter: 'blur(20px)',
                  color: 'white',
                  boxShadow: '0 10px 30px rgba(0,0,0,0.4)',
                  padding: '12px'
                }}
                itemStyle={{ color: 'white', fontWeight: 800, fontSize: '0.9rem' }}
                labelStyle={{ color: 'rgba(255,255,255,0.5)', fontWeight: 600, fontSize: '0.75rem', marginBottom: '4px' }}
              />
              <Area 
                type="monotone" 
                dataKey="engagement" 
                stroke={getPlatformColor(activePlatform)} 
                strokeWidth={3}
                fillOpacity={1} 
                fill="url(#colorEng)" 
                animationDuration={1500}
                activeDot={{ r: 6, strokeWidth: 0, fill: 'white' }}
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>

        <div className="glass-card" style={{
          background: 'rgba(255, 255, 255, 0.02)',
          borderRadius: '32px',
          padding: '2.5rem',
          border: '1px solid rgba(255, 255, 255, 0.05)'
        }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '2rem', alignItems: 'center' }}>
            <h3 style={{ fontSize: '1.5rem', fontWeight: 800 }}>Timeframe Post Analysis</h3>
            <span style={{ fontSize: '0.8rem', color: '#10b981', fontWeight: 700, padding: '0.4rem 0.8rem', background: 'rgba(16, 185, 129, 0.1)', borderRadius: '12px' }}>
              {metrics?.posts?.length || 0} Posts Found
            </span>
          </div>
          
          <div style={{ overflowX: 'auto' }}>
            {isLoading ? (
              <div style={{ padding: '4rem', textAlign: 'center', opacity: 0.5 }}>Fetching Meta Data...</div>
            ) : metrics?.posts?.length === 0 ? (
              <div style={{ padding: '4rem', textAlign: 'center', opacity: 0.5 }}>No content published in this period.</div>
            ) : (
              <table style={{ width: '100%', borderCollapse: 'separate', borderSpacing: '0 12px' }}>
                <thead>
                  <tr style={{ textAlign: 'left', color: 'rgba(255,255,255,0.3)', fontSize: '0.75rem', textTransform: 'uppercase', letterSpacing: '0.1em' }}>
                    <th style={{ padding: '1rem 1.5rem' }}>Media</th>
                    <th style={{ padding: '1rem 1.5rem', minWidth: '120px' }}>Date</th>
                    <th style={{ padding: '1rem 1.5rem', minWidth: '140px' }}>Interactions</th>
                    <th style={{ padding: '1rem 1.5rem', minWidth: '160px' }}>Reach / Plays</th>
                    <th style={{ padding: '1rem 1.5rem' }}>Status</th>
                    <th style={{ padding: '1rem 1.5rem', textAlign: 'right' }}>Action</th>
                  </tr>
                </thead>
                <tbody>
                  {[...(metrics?.posts || [])].sort((a, b) => new Date(b.published_at).getTime() - new Date(a.published_at).getTime()).map(post => (
                    <tr 
                      key={post.id} 
                      onClick={() => {
                        setSelectedPostForInsights(post);
                        setIsInsightsModalOpen(true);
                      }}
                      style={{ 
                        background: 'rgba(255,255,255,0.02)', 
                        transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                        cursor: 'pointer'
                      }}
                      onMouseEnter={(e) => {
                        e.currentTarget.style.background = 'rgba(255,255,255,0.05)';
                        e.currentTarget.style.transform = 'translateY(-2px)';
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.background = 'rgba(255,255,255,0.02)';
                        e.currentTarget.style.transform = 'translateY(0)';
                      }}
                    >
                      <td style={{ padding: '1rem 1.5rem', borderRadius: '16px 0 0 16px' }}>
                        <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
                          <img 
                            src={post.thumbnail_url || post.media_url} 
                            alt="thumb" 
                            style={{ width: '48px', height: '48px', borderRadius: '12px', objectFit: 'cover', border: '1px solid rgba(255,255,255,0.1)' }} 
                          />
                          <div style={{ fontSize: '0.85rem', fontWeight: 600, maxWidth: '200px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                            {post.content || 'No Caption'}
                          </div>
                        </div>
                      </td>
                      <td style={{ padding: '1rem 1.5rem', fontSize: '0.75rem', color: 'rgba(255,255,255,0.4)', fontWeight: 600 }}>
                        {new Date(post.published_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
                      </td>
                      <td style={{ padding: '1rem 1.5rem', fontWeight: 800, fontSize: '1.1rem', color: post.restricted ? 'rgba(255,255,255,0.2)' : '#f59e0b' }}>
                        {post.restricted ? '--' : post.engagement.toLocaleString()}
                      </td>
                      <td style={{ padding: '1rem 1.5rem', fontWeight: 600, color: post.restricted ? 'rgba(255,255,255,0.2)' : 'white' }}>
                        {post.restricted ? '--' : post.views.toLocaleString()}
                      </td>
                      <td style={{ padding: '1rem 1.5rem' }}>
                        <span style={{ 
                          padding: '0.4rem 0.8rem', 
                          borderRadius: '12px', 
                          background: post.restricted 
                            ? 'rgba(239, 68, 68, 0.1)' 
                            : post.engagement > 100 
                              ? 'rgba(16, 185, 129, 0.1)' 
                              : post.engagement > 0 ? 'rgba(59, 130, 246, 0.1)' : 'rgba(255,255,255,0.05)', 
                          color: post.restricted 
                            ? '#ef4444' 
                            : post.engagement > 100 
                              ? '#10b981' 
                              : post.engagement > 0 ? '#3b82f6' : 'rgba(255,255,255,0.3)', 
                          fontSize: '0.65rem', 
                          fontWeight: 800,
                          textTransform: 'uppercase',
                          letterSpacing: '0.05em',
                          border: `1px solid ${post.restricted ? 'rgba(239, 68, 68, 0.2)' : post.engagement > 100 ? 'rgba(16, 185, 129, 0.2)' : 'rgba(255,255,255,0.05)'}`
                        }}>
                          {post.restricted ? 'Access Restricted' : post.engagement > 100 ? 'High Performance' : post.engagement > 0 ? 'Developing' : 'Steady'}
                        </span>
                      </td>
                      <td style={{ padding: '1rem 1.5rem', borderRadius: '0 16px 16px 0', textAlign: 'right' }}>
                        <div style={{ display: 'flex', gap: '1rem', alignItems: 'center', justifyContent: 'flex-end' }}>
                          <span style={{ color: 'rgba(255,255,255,0.3)', fontSize: '0.75rem', fontWeight: 700 }}>Details →</span>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>
      </div>

      <SocialComposerModal 
        isOpen={isComposerOpen}
        onClose={() => setIsComposerOpen(false)}
        fetchWithAuth={fetchWithAuth}
      />

      <PostInsightsModal 
        isOpen={isInsightsModalOpen}
        post={selectedPostForInsights}
        onClose={() => setIsInsightsModalOpen(false)}
        fetchWithAuth={fetchWithAuth}
      />
    </div>
  );
};
