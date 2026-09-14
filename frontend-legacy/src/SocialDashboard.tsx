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
    <div className="social-dashboard-container tab-content-fade" style={{ 
      animation: 'fadeIn 0.6s ease-out'
    }}>
      {/* Header Island with Platform Switcher and Date Picker */}
      <div className="glass-island-premium" style={{
        padding: '1.25rem 2rem',
        marginBottom: '2.5rem',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
      }}>
        <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
          {(['instagram', 'facebook', 'threads'] as const).map(p => (
            <button
              key={p}
              onClick={() => setActivePlatform(p)}
              style={{
                background: activePlatform === p ? 'var(--accent-color)' : 'var(--bg-input)',
                color: activePlatform === p ? 'white' : 'var(--text-secondary)',
                border: '1px solid var(--border-color)',
                padding: '0.6rem 1.4rem',
                borderRadius: '14px',
                fontWeight: 700,
                cursor: 'pointer',
                transition: 'all 0.3s cubic-bezier(0.16, 1, 0.3, 1)',
                textTransform: 'capitalize',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                fontSize: '0.9rem',
                boxShadow: activePlatform === p ? 'var(--shadow-md)' : 'none'
              }}
            >
              {p}
            </button>
          ))}
          
          <div style={{ width: '1px', height: '24px', background: 'var(--border-color)', margin: '0 0.5rem' }} />
          
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
            className="btn-secondary"
            style={{
              padding: '0.6rem 1.25rem',
              borderRadius: '12px',
              fontWeight: 600,
              display: 'flex',
              alignItems: 'center',
              gap: '0.6rem',
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
                background: 'linear-gradient(135deg, var(--status-active), var(--accent-hover))',
                border: 'none',
                color: 'white',
                padding: '0.6rem 1.25rem',
                borderRadius: '12px',
                fontWeight: 800,
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                boxShadow: 'var(--shadow-md)'
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
            <div className="glass-island-premium" style={{
              background: 'var(--status-danger-bg)',
              borderColor: 'var(--status-danger)',
              padding: '1.25rem 2rem',
              marginBottom: '2rem',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between'
            }}>
              <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
                <div style={{
                  width: '44px', height: '44px', borderRadius: '14px', 
                  background: 'var(--status-danger-bg)', display: 'flex', alignItems: 'center', 
                  justifyContent: 'center', fontSize: '1.2rem',
                  border: '1px solid var(--status-danger)'
                }}>
                  ⚠️
                </div>
                <div>
                  <div style={{ fontSize: '0.7rem', color: 'var(--status-danger)', fontWeight: 800, textTransform: 'uppercase', letterSpacing: '0.15em', marginBottom: '0.2rem' }}>
                    Platform Access Restricted
                  </div>
                  <div style={{ fontSize: '0.95rem', fontWeight: 600, color: 'var(--text-primary)', marginBottom: '0.2rem' }}>
                    {fetchError}
                  </div>
                  <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>
                    Meta requires an upgrade of your permissions. <a href="https://developers.facebook.com/tools/debug/accesstoken/" target="_blank" rel="noopener noreferrer" style={{ color: 'var(--status-danger)', textDecoration: 'underline' }}>Check Token Scopes</a>
                  </div>
                </div>
              </div>
              <div style={{ display: 'flex', gap: '1rem' }}>
                <button
                  onClick={() => fetchMetrics()}
                  style={{
                    background: 'var(--status-danger-bg)',
                    border: '1px solid var(--status-danger)',
                    color: 'var(--status-danger)',
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
          {(activePlatform === 'instagram' || activePlatform === 'facebook') && metrics?.account && (
            <div style={{ 
              display: 'grid', 
              gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', 
              gap: '1rem', 
              marginBottom: '2rem',
              padding: '1.5rem',
              background: 'var(--bg-input)',
              borderRadius: '24px',
              border: '1px solid var(--border-color)'
            }}>
              {[
                { label: 'Followers', value: metrics.account.follower_count, icon: '👥', color: 'var(--accent-color)' },
                { label: 'Growth Reach', value: metrics.account.total_reach, icon: '📈', color: 'var(--status-active)' },
                { label: 'Direct Impressions', value: metrics.account.total_views, icon: '🌊', color: 'var(--accent-color)' }
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
                    background: 'var(--nav-active-bg)', display: 'flex', alignItems: 'center', 
                    justifyContent: 'center', fontSize: '1.3rem',
                    border: `1px solid var(--border-color)`,
                    boxShadow: `var(--shadow-sm)`
                  }}>
                    {m.icon}
                  </div>
                  <div>
                    <div style={{ fontSize: '0.7rem', color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontWeight: 700 }}>
                      {m.label}
                    </div>
                    <div style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--text-primary)', marginTop: '0.1rem' }}>
                      {m.value.toLocaleString()}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* Metric Overview Cards */}
          <div style={{ 
            display: 'grid', 
            gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', 
            gap: '1.5rem', 
            marginBottom: '2.5rem' 
          }}>
            {[
              { label: 'Total Reach', value: metrics ? metrics.total_reach : 0, change: '+14.2%', color: 'var(--accent-color)', icon: '👁️' },
              { label: 'Total Views / Impressions', value: metrics ? metrics.total_views : 0, change: '+8.1%', color: 'var(--status-active)', icon: '🎬' },
              { label: 'Total Engagement', value: metrics ? metrics.total_engagement : 0, change: '+22.4%', color: 'var(--status-warning)', icon: '❤️' },
            ].map((stat, idx) => (
              <div 
                key={idx} 
                className="glass-card-premium hover-lift" 
                style={{ 
                  padding: '1.75rem',
                  display: 'flex',
                  flexDirection: 'column',
                  justifyContent: 'space-between',
                  minHeight: '160px',
                  background: 'var(--bg-card)',
                  borderRadius: '24px',
                  border: '1px solid var(--border-color)'
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                  <span style={{ fontSize: '1.5rem' }}>{stat.icon}</span>
                  <span style={{ 
                    fontSize: '0.75rem', 
                    fontWeight: 700, 
                    color: 'var(--status-active)', 
                    background: 'var(--status-active-bg)', 
                    padding: '0.25rem 0.6rem', 
                    borderRadius: '20px',
                    border: '1px solid var(--status-active)'
                  }}>
                    {stat.change}
                  </span>
                </div>
                <div>
                  <div style={{ fontSize: '0.8rem', color: 'var(--text-tertiary)', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                    {stat.label}
                  </div>
                  <div style={{ fontSize: '2rem', fontWeight: 800, color: 'var(--text-primary)', marginTop: '0.25rem' }}>
                    {isLoading ? '...' : stat.value.toLocaleString()}
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* Performance Chart Island */}
          <div className="glass-island-premium" style={{ padding: '2rem', marginBottom: '2.5rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
              <div>
                <h3 style={{ margin: 0, fontSize: '1.1rem', fontWeight: 700 }}>Engagement & Reach Trend</h3>
                <p style={{ margin: '0.25rem 0 0 0', fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                  Aggregated performance metrics over selected timeframe
                </p>
              </div>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <span style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', fontSize: '0.75rem', color: 'var(--text-secondary)', fontWeight: 600 }}>
                  <span style={{ width: '8px', height: '8px', borderRadius: '50%', background: getPlatformColor(activePlatform) }}></span> Engagement
                </span>
              </div>
            </div>

            <div style={{ width: '100%', height: 300 }}>
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={historyData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                  <defs>
                    <linearGradient id="colorEngagement" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor={getPlatformColor(activePlatform)} stopOpacity={0.4}/>
                      <stop offset="95%" stopColor={getPlatformColor(activePlatform)} stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border-color)" vertical={false} />
                  <XAxis dataKey="date" stroke="var(--text-tertiary)" fontSize={12} tickLine={false} axisLine={false} />
                  <YAxis stroke="var(--text-tertiary)" fontSize={12} tickLine={false} axisLine={false} />
                  <Tooltip 
                    contentStyle={{ 
                      background: 'var(--bg-card)', 
                      borderColor: 'var(--border-color)', 
                      borderRadius: '12px',
                      color: 'var(--text-primary)',
                      boxShadow: 'var(--shadow-md)'
                    }} 
                  />
                  <Area 
                    type="monotone" 
                    dataKey="engagement" 
                    stroke={getPlatformColor(activePlatform)} 
                    strokeWidth={3}
                    fillOpacity={1} 
                    fill="url(#colorEngagement)" 
                  />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </div>

          {/* Posts Grid / Table */}
          <div className="glass-island-premium" style={{ padding: '2rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
              <div>
                <h3 style={{ margin: 0, fontSize: '1.1rem', fontWeight: 700 }}>Published Media Feed</h3>
                <p style={{ margin: '0.25rem 0 0 0', fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                  Recent posts published on {activePlatform.toUpperCase()}
                </p>
              </div>
            </div>

            {isLoading ? (
              <div style={{ padding: '3rem', textAlign: 'center', color: 'var(--text-tertiary)' }}>
                Loading platform feed...
              </div>
            ) : !metrics?.posts || metrics.posts.length === 0 ? (
              <div style={{ padding: '3rem', textAlign: 'center', color: 'var(--text-tertiary)' }}>
                No posts found for this platform in the selected date range.
              </div>
            ) : (
              <div style={{ overflowX: 'auto' }}>
                <table style={{ width: '100%', borderCollapse: 'separate', borderSpacing: '0 0.75rem' }}>
                  <thead>
                    <tr style={{ color: 'var(--text-tertiary)', fontSize: '0.75rem', textTransform: 'uppercase', letterSpacing: '0.05em', textAlign: 'left' }}>
                      <th style={{ padding: '0.5rem 1.5rem' }}>Content</th>
                      <th style={{ padding: '0.5rem 1.5rem' }}>Published</th>
                      <th style={{ padding: '0.5rem 1.5rem' }}>Engagement</th>
                      <th style={{ padding: '0.5rem 1.5rem' }}>Views / Reach</th>
                      <th style={{ padding: '0.5rem 1.5rem' }}>Status</th>
                      <th style={{ padding: '0.5rem 1.5rem', textAlign: 'right' }}>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {metrics.posts.map(post => (
                      <tr 
                        key={post.id} 
                        style={{ 
                          background: 'var(--bg-card)', 
                          transition: 'all 0.3s ease',
                          cursor: 'pointer' 
                        }}
                        onClick={() => {
                          setSelectedPostForInsights(post);
                          setIsInsightsModalOpen(true);
                        }}
                        onMouseEnter={(e) => {
                          e.currentTarget.style.background = 'var(--bg-input)';
                          e.currentTarget.style.transform = 'translateY(-2px)';
                        }}
                        onMouseLeave={(e) => {
                          e.currentTarget.style.background = 'var(--bg-card)';
                          e.currentTarget.style.transform = 'translateY(0)';
                        }}
                      >
                        <td style={{ padding: '1rem 1.5rem', borderRadius: '16px 0 0 16px' }}>
                          <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
                            <img 
                              src={post.thumbnail_url || post.media_url} 
                              alt="thumb" 
                              style={{ width: '48px', height: '48px', borderRadius: '12px', objectFit: 'cover', border: '1px solid var(--border-color)' }} 
                            />
                            <div style={{ fontSize: '0.85rem', fontWeight: 600, maxWidth: '200px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                              {post.content || 'No Caption'}
                            </div>
                          </div>
                        </td>
                        <td style={{ padding: '1rem 1.5rem', fontSize: '0.75rem', color: 'var(--text-tertiary)', fontWeight: 600 }}>
                          {new Date(post.published_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
                        </td>
                        <td style={{ padding: '1rem 1.5rem', fontWeight: 800, fontSize: '1.1rem', color: post.restricted ? 'var(--text-tertiary)' : 'var(--status-warning)' }}>
                          {post.restricted ? '--' : post.engagement.toLocaleString()}
                        </td>
                        <td style={{ padding: '1rem 1.5rem', fontWeight: 600, color: post.restricted ? 'var(--text-tertiary)' : 'var(--text-primary)' }}>
                          {post.restricted ? '--' : post.views.toLocaleString()}
                        </td>
                        <td style={{ padding: '1rem 1.5rem' }}>
                          <span style={{ 
                            padding: '0.4rem 0.8rem', 
                            borderRadius: '12px', 
                            background: post.restricted 
                              ? 'var(--status-danger-bg)' 
                              : post.engagement > 100 
                                ? 'var(--status-active-bg)' 
                                : post.engagement > 0 ? 'var(--bg-input)' : 'var(--bg-input)', 
                            color: post.restricted 
                              ? 'var(--status-danger)' 
                              : post.engagement > 100 
                                ? 'var(--status-active)' 
                                : post.engagement > 0 ? 'var(--accent-color)' : 'var(--text-tertiary)', 
                            fontSize: '0.65rem', 
                            fontWeight: 800,
                            textTransform: 'uppercase',
                            letterSpacing: '0.05em',
                            border: `1px solid ${post.restricted ? 'var(--status-danger)' : post.engagement > 100 ? 'var(--status-active)' : 'var(--border-color)'}`
                          }}>
                            {post.restricted ? 'Access Restricted' : post.engagement > 100 ? 'High Performance' : post.engagement > 0 ? 'Developing' : 'Steady'}
                          </span>
                        </td>
                        <td style={{ padding: '1rem 1.5rem', borderRadius: '0 16px 16px 0', textAlign: 'right' }}>
                          <div style={{ display: 'flex', gap: '1rem', alignItems: 'center', justifyContent: 'flex-end' }}>
                            <span style={{ color: 'var(--text-tertiary)', fontSize: '0.75rem', fontWeight: 700 }}>Details →</span>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
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
