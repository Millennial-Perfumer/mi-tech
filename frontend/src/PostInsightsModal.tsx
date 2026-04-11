import React, { useEffect, useState } from 'react';
import { API_BASE } from './api';

interface PostInsightsModalProps {
  post: any;
  isOpen: boolean;
  onClose: () => void;
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
}

export const PostInsightsModal: React.FC<PostInsightsModalProps> = ({ post, isOpen, onClose, fetchWithAuth }) => {
  const [insights, setInsights] = useState<any>(null);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (isOpen && post) {
      fetchDetailedInsights();
    }
  }, [isOpen, post]);

  const fetchDetailedInsights = async () => {
    setIsLoading(true);
    setInsights(null);
    try {
      const response = await fetchWithAuth(`${API_BASE}/api/marketing/smm/post/insights?id=${post.id}&media_type=${post.mediaType || post.media_type || ''}`);
      const data = await response.json();
      if (data.success) {
        setInsights(data.insights);
      }
    } catch (err) {
      console.error('Failed to fetch detailed insights:', err);
    } finally {
      setIsLoading(false);
    }
  };

  if (!isOpen) return null;

  const metrics = insights?.metrics || {};
  const followBreakdown = insights?.breakdowns?.follow_type || {};
  
  // Calculate follower/non-follower stats for Reach
  const reachTotal = metrics.reach || 0;
  // Support both cases just in case normalization failed
  const reachFollower = followBreakdown.FOLLOWER || followBreakdown.follower || 0;
  const reachNonFollower = followBreakdown.NON_FOLLOWER || followBreakdown.non_follower || 0;
  
  // Use the larger of reachTotal or the sum of followers/non-followers as denominator
  // (In case backend used Impressions as fallback, sum might exceed reach unique count)
  const calcBase = Math.max(reachTotal, reachFollower + reachNonFollower);
  
  const hasReachData = (reachFollower > 0 || reachNonFollower > 0);
  const reachFollowerPct = calcBase > 0 ? ((reachFollower / calcBase) * 100).toFixed(1) : '0';
  const reachNonFollowerPct = calcBase > 0 ? ((reachNonFollower / calcBase) * 100).toFixed(1) : '0';

  // Logic for "Processing" vs "Unavailable"
  const publishedDate = post?.published_at ? new Date(post.published_at) : new Date();
  const hoursSincePost = (new Date().getTime() - publishedDate.getTime()) / (1000 * 60 * 60);
  const isRecent = hoursSincePost < 48;
  const isCarousel = (post?.mediaType || post?.media_type) === 'CAROUSEL_ALBUM';

  // Calculate Interactions stats (using likes+comments+shares+saved as total if total_interactions is missing)
  const interactionsTotal = metrics.total_interactions || (metrics.likes + metrics.comments + metrics.shares + metrics.saved) || 0;

  return (
    <div style={{
      position: 'fixed',
      top: 0,
      left: 0,
      width: '100%',
      height: '100%',
      background: 'rgba(0,0,0,0.4)',
      backdropFilter: 'blur(8px)',
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      zIndex: 2000,
      padding: '2rem'
    }} onClick={onClose}>
      <div className="glass-card-premium" style={{
        background: 'var(--surface-color)',
        width: '100%',
        maxWidth: '500px',
        maxHeight: '90vh',
        overflowY: 'auto',
        color: 'var(--text-primary)',
        padding: '2.5rem',
        position: 'relative'
      }} onClick={e => e.stopPropagation()}>
        <button 
          onClick={onClose}
          aria-label="Close insights"
          style={{
            position: 'absolute',
            top: '20px',
            right: '20px',
            background: 'var(--bg-input)',
            border: 'none',
            color: 'var(--text-primary)',
            width: '32px',
            height: '32px',
            borderRadius: '50%',
            cursor: 'pointer',
            fontSize: '1.2rem',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center'
          }}
        >
          ×
        </button>

        <h2 style={{ fontSize: '1.8rem', fontWeight: 800, marginBottom: '2rem', display: 'flex', alignItems: 'center', gap: '0.8rem' }}>
          Post Insights
        </h2>

        {isLoading ? (
          <div style={{ padding: '4rem', textAlign: 'center', opacity: 0.5 }}>Syncing with Meta...</div>
        ) : !insights ? (
          <div style={{ padding: '4rem', textAlign: 'center', opacity: 0.5 }}>Insights unavailable for this media type.</div>
        ) : (
          <div>
            {/* Views Section */}
            <section style={{ marginBottom: '2.5rem' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
                <h3 style={{ fontSize: '1.2rem', fontWeight: 700, margin: 0 }}>Views</h3>
                <span style={{ fontSize: '1.2rem', fontWeight: 800 }}>{metrics.views?.toLocaleString() || 0}</span>
              </div>

              {/* Follower breakdown bars — only show if data exists */}
              {hasReachData ? (
                <>
                  <div style={{ marginBottom: '1rem' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.85rem', marginBottom: '0.4rem', color: 'var(--text-secondary)' }}>
                      <span>Followers</span>
                      <span>{reachFollower.toLocaleString()} ({reachFollowerPct}%)</span>
                    </div>
                    <div style={{ height: '6px', background: 'var(--bg-input)', borderRadius: '3px', overflow: 'hidden' }}>
                      <div style={{ height: '100%', width: `${reachFollowerPct}%`, background: 'var(--accent-color)', borderRadius: '3px', transition: 'width 0.6s ease' }} />
                    </div>
                  </div>
                  <div style={{ marginBottom: '1.5rem' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.85rem', marginBottom: '0.4rem', color: 'var(--text-secondary)' }}>
                      <span>Non-followers</span>
                      <span>{reachNonFollower.toLocaleString()} ({reachNonFollowerPct}%)</span>
                    </div>
                    <div style={{ height: '6px', background: 'var(--bg-input)', borderRadius: '3px', overflow: 'hidden' }}>
                      <div style={{ height: '100%', width: `${reachNonFollowerPct}%`, background: '#3b82f6', borderRadius: '3px', transition: 'width 0.6s ease' }} />
                    </div>
                  </div>
                </>
              ) : (
                <div style={{
                  padding: '1.25rem',
                  background: isCarousel ? 'var(--bg-input)' : isRecent ? 'var(--status-warning-bg)' : 'var(--bg-input)',
                  border: isCarousel ? '1px solid var(--border-color)' : isRecent ? '1px solid var(--status-warning)' : '1px solid var(--border-color)',
                  borderRadius: '16px',
                  marginBottom: '2rem',
                  fontSize: '0.85rem'
                }}>
                  <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'flex-start' }}>
                    <span style={{ fontSize: '1.2rem' }}>{isCarousel ? '💡' : isRecent ? '⏳' : 'ℹ️'}</span>
                    <div>
                      <div style={{ fontWeight: 700, color: isRecent && !isCarousel ? 'var(--status-warning)' : 'var(--text-primary)', marginBottom: '0.25rem' }}>
                        {isCarousel ? 'Follower Breakdown Unavailable' : isRecent ? 'Insight Processing' : 'Data Unavailable'}
                      </div>
                      <div style={{ color: 'var(--text-secondary)', lineHeight: 1.5 }}>
                        {isCarousel 
                          ? 'Instagram does not provide follower vs non-follower breakdowns for Carousel albums via the API.'
                          : isRecent
                            ? 'Meta unique reach data can take 24–48 hours to process. Engagement (likes/shares) is available immediately.'
                            : 'This breakdown is unavailable for this post. This can happen if the post has low engagement or the account has < 100 followers.'}
                      </div>
                    </div>
                  </div>
                </div>
              )}

              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '1.5rem', fontWeight: 700 }}>
                <span>Accounts reached</span>
                <span>{reachTotal.toLocaleString()}</span>
              </div>
            </section>

            {/* Interactions Section */}
            <section style={{ borderTop: '1px solid var(--border-color)', paddingTop: '2.5rem' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
                <h3 style={{ fontSize: '1.2rem', fontWeight: 700, margin: 0 }}>Interactions</h3>
                <span style={{ fontSize: '1.2rem', fontWeight: 800 }}>{interactionsTotal.toLocaleString()}</span>
              </div>

              <div style={{ marginBottom: '2rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span style={{ color: 'var(--text-secondary)' }}>Post interactions</span>
                  <span style={{ fontWeight: 600 }}>{interactionsTotal}</span>
                </div>
                {[
                  { label: 'Likes', value: metrics.likes, icon: '❤️' },
                  { label: 'Saves', value: metrics.saved, icon: '🔖' },
                  { label: 'Comments', value: metrics.comments, icon: '💬' },
                  { label: 'Shares', value: metrics.shares, icon: '📤' }
                ].map(i => (
                  <div key={i.label} style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.95rem' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.7rem' }}>
                      <span style={{ opacity: 0.6 }}>{i.icon}</span>
                      <span style={{ color: 'var(--text-secondary)' }}>{i.label}</span>
                    </div>
                    <span style={{ fontWeight: 700 }}>{i.value?.toLocaleString() || 0}</span>
                  </div>
                ))}
              </div>
            </section>
          </div>
        )}
      </div>
    </div>
  );
};
