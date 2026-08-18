import React from 'react';

interface SocialQueuePostItem {
  id?: number | string;
  folder_name: string;
  post_type: string;
  caption: string;
  hashtags: string;
  status: string;
  created_at?: string;
  gdrive_folder_id?: string;
}

interface SocialQueueDashboardProps {
  queuePosts: SocialQueuePostItem[];
  gdriveFolderUrl?: string;
  onOpenUploader: () => void;
  onRefresh: () => void;
  onNavigate?: (tab: string) => void;
}

export const SocialQueueDashboard: React.FC<SocialQueueDashboardProps> = ({
  queuePosts,
  gdriveFolderUrl,
  onOpenUploader,
  onRefresh,
  onNavigate,
}) => {
  const queuedCount = queuePosts.filter(p => p.status === 'QUEUED').length;
  const publishingCount = queuePosts.filter(p => p.status === 'PUBLISHING').length;
  const publishedCount = queuePosts.filter(p => p.status === 'PUBLISHED').length;

  const carouselCount = queuePosts.filter(p => p.post_type === 'CAROUSEL').length;
  const videoCount = queuePosts.filter(p => p.post_type === 'VIDEO').length;
  const singlePhotoCount = queuePosts.filter(p => p.post_type === 'SINGLE_PHOTO').length;

  return (
    <div className="social-queue-dashboard" style={{ animation: 'fadeIn 0.4s ease-out' }}>
      {/* Top Controls & Overview Island */}
      <div className="glass-island-premium" style={{ padding: '2.25rem', marginBottom: '2rem', borderRadius: '24px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: '1.5rem' }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.4rem' }}>
              <h2 style={{ margin: 0, fontSize: '1.35rem', fontWeight: 800, color: 'var(--text-primary)', letterSpacing: '-0.02em' }}>
                Queue Control Center
              </h2>
            </div>
            <p style={{ margin: 0, color: 'var(--text-secondary)', fontSize: '0.875rem', maxWidth: '620px', lineHeight: '1.5' }}>
              Upload single photos, multi-image carousels, or video reels directly into your Google Drive automation folder. n8n processes subfolders on schedule and publishes across Facebook, Instagram, Threads, and X.
            </p>
          </div>

          {/* Action Controls Group */}
          <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap', alignItems: 'center' }}>
            {gdriveFolderUrl && (
              <a
                href={gdriveFolderUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="btn-secondary"
                style={{
                  padding: '0.65rem 1.25rem',
                  borderRadius: '12px',
                  fontWeight: 700,
                  fontSize: '0.85rem',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.5rem',
                  textDecoration: 'none',
                  color: 'var(--text-primary)',
                  border: '1px solid var(--border-color)',
                  height: '42px'
                }}
              >
                Open Google Drive ↗
              </a>
            )}

            <button
              onClick={onRefresh}
              className="btn-secondary"
              style={{
                padding: '0.65rem 1.25rem',
                borderRadius: '12px',
                fontWeight: 600,
                fontSize: '0.85rem',
                height: '42px',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem'
              }}
            >
              Refresh
            </button>

            {onNavigate && (
              <button
                onClick={() => onNavigate('settings')}
                className="btn-secondary"
                style={{
                  padding: '0.65rem 1.25rem',
                  borderRadius: '12px',
                  fontWeight: 600,
                  fontSize: '0.85rem',
                  height: '42px'
                }}
                title="Configure Drive URL in Settings"
              >
                Settings
              </button>
            )}

            <button
              onClick={onOpenUploader}
              className="btn-primary hover-lift"
              style={{
                background: 'linear-gradient(135deg, var(--accent-color), #059669)',
                border: 'none',
                color: 'white',
                padding: '0.65rem 1.4rem',
                borderRadius: '12px',
                fontWeight: 800,
                fontSize: '0.875rem',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                height: '42px',
                boxShadow: '0 8px 20px rgba(16, 185, 129, 0.25)'
              }}
            >
              + New Queue Post
            </button>
          </div>
        </div>

        {/* Metric Cards Grid */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1.25rem', marginTop: '2rem' }}>
          <div className="hover-lift" style={{ padding: '1.25rem 1.5rem', borderRadius: '18px', background: 'var(--bg-card)', border: '1px solid var(--border-color)', transition: 'all 0.2s ease' }}>
            <div style={{ fontSize: '0.725rem', color: 'var(--text-tertiary)', textTransform: 'uppercase', fontWeight: 800, letterSpacing: '0.05em' }}>Total Queued</div>
            <div style={{ fontSize: '1.85rem', fontWeight: 800, color: '#60a5fa', marginTop: '0.3rem' }}>{queuedCount}</div>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', marginTop: '0.2rem' }}>Pending processing</div>
          </div>

          <div className="hover-lift" style={{ padding: '1.25rem 1.5rem', borderRadius: '18px', background: 'var(--bg-card)', border: '1px solid var(--border-color)', transition: 'all 0.2s ease' }}>
            <div style={{ fontSize: '0.725rem', color: 'var(--text-tertiary)', textTransform: 'uppercase', fontWeight: 800, letterSpacing: '0.05em' }}>Publishing Now</div>
            <div style={{ fontSize: '1.85rem', fontWeight: 800, color: '#fbbf24', marginTop: '0.3rem' }}>{publishingCount}</div>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', marginTop: '0.2rem' }}>Active n8n workflow</div>
          </div>

          <div className="hover-lift" style={{ padding: '1.25rem 1.5rem', borderRadius: '18px', background: 'var(--bg-card)', border: '1px solid var(--border-color)', transition: 'all 0.2s ease' }}>
            <div style={{ fontSize: '0.725rem', color: 'var(--text-tertiary)', textTransform: 'uppercase', fontWeight: 800, letterSpacing: '0.05em' }}>Published Total</div>
            <div style={{ fontSize: '1.85rem', fontWeight: 800, color: '#34d399', marginTop: '0.3rem' }}>{publishedCount}</div>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', marginTop: '0.2rem' }}>Successfully posted</div>
          </div>

          <div className="hover-lift" style={{ padding: '1.25rem 1.5rem', borderRadius: '18px', background: 'var(--bg-card)', border: '1px solid var(--border-color)', transition: 'all 0.2s ease' }}>
            <div style={{ fontSize: '0.725rem', color: 'var(--text-tertiary)', textTransform: 'uppercase', fontWeight: 800, letterSpacing: '0.05em' }}>Format Distribution</div>
            <div style={{ fontSize: '0.9rem', fontWeight: 800, color: 'var(--text-primary)', marginTop: '0.5rem' }}>
              {singlePhotoCount} Photo · {carouselCount} Carousel · {videoCount} Video
            </div>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', marginTop: '0.2rem' }}>Media breakdown</div>
          </div>
        </div>
      </div>

      {/* Main Table View */}
      <div className="glass-island-premium" style={{ padding: '2rem', borderRadius: '24px', marginTop: '1.5rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
          <h3 style={{ margin: 0, fontSize: '1.15rem', fontWeight: 800, color: 'var(--text-primary)' }}>
            Active Queue Items ({queuePosts.length})
          </h3>
        </div>

        {queuePosts.length === 0 ? (
          <div style={{ padding: '4rem 2rem', textAlign: 'center', color: 'var(--text-tertiary)' }}>
            <div style={{ fontSize: '2.5rem', marginBottom: '1rem', opacity: 0.5 }}>📁</div>
            <h4 style={{ margin: '0 0 0.5rem 0', color: 'var(--text-secondary)' }}>No Queue Posts Found</h4>
            <p style={{ margin: 0, fontSize: '0.875rem' }}>Click "+ New Queue Post" above to upload content directly to Google Drive.</p>
          </div>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'separate', borderSpacing: '0 0.6rem' }}>
              <thead>
                <tr style={{ color: 'var(--text-tertiary)', fontSize: '0.75rem', textTransform: 'uppercase', textAlign: 'left', letterSpacing: '0.05em' }}>
                  <th style={{ padding: '0.75rem 1.25rem' }}>Folder Name</th>
                  <th style={{ padding: '0.75rem 1.25rem' }}>Format</th>
                  <th style={{ padding: '0.75rem 1.25rem' }}>Caption Preview</th>
                  <th style={{ padding: '0.75rem 1.25rem' }}>Hashtags</th>
                  <th style={{ padding: '0.75rem 1.25rem' }}>Status</th>
                  <th style={{ padding: '0.75rem 1.25rem' }}>Created At</th>
                  <th style={{ padding: '0.75rem 1.25rem', textAlign: 'right' }}>Automation</th>
                </tr>
              </thead>
              <tbody>
                {queuePosts.map((item) => (
                  <tr key={item.id || item.folder_name} style={{ background: 'var(--bg-card)', borderRadius: '14px', transition: 'all 0.2s ease' }}>
                    <td style={{ padding: '1rem 1.25rem', borderRadius: '14px 0 0 14px', fontFamily: 'monospace', fontWeight: 700, color: 'var(--accent-color)', fontSize: '0.9rem' }}>
                      {item.folder_name}
                    </td>
                    <td style={{ padding: '1rem 1.25rem' }}>
                      <span style={{
                        padding: '0.35rem 0.75rem',
                        borderRadius: '12px',
                        fontSize: '0.75rem',
                        fontWeight: 800,
                        letterSpacing: '0.03em',
                        background: item.post_type === 'CAROUSEL' ? 'rgba(168, 85, 247, 0.15)' : item.post_type === 'VIDEO' ? 'rgba(239, 68, 68, 0.15)' : 'rgba(16, 185, 129, 0.15)',
                        color: item.post_type === 'CAROUSEL' ? '#c084fc' : item.post_type === 'VIDEO' ? '#f87171' : '#34d399',
                        border: `1px solid ${item.post_type === 'CAROUSEL' ? 'rgba(168, 85, 247, 0.3)' : item.post_type === 'VIDEO' ? 'rgba(239, 68, 68, 0.3)' : 'rgba(16, 185, 129, 0.3)'}`
                      }}>
                        {item.post_type === 'CAROUSEL' ? 'CAROUSEL' : item.post_type === 'VIDEO' ? 'VIDEO' : 'PHOTO'}
                      </span>
                    </td>
                    <td style={{ padding: '1rem 1.25rem', fontSize: '0.875rem', maxWidth: '240px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', color: 'var(--text-primary)' }}>
                      {item.caption || 'No caption'}
                    </td>
                    <td style={{ padding: '1rem 1.25rem', fontSize: '0.8rem', color: 'var(--text-secondary)', maxWidth: '140px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {item.hashtags || '--'}
                    </td>
                    <td style={{ padding: '1rem 1.25rem' }}>
                      <span style={{
                        padding: '0.35rem 0.85rem',
                        borderRadius: '20px',
                        fontSize: '0.75rem',
                        fontWeight: 800,
                        background: item.status === 'PUBLISHED' ? 'var(--status-active-bg)' : item.status === 'PUBLISHING' ? 'var(--status-warning-bg)' : 'rgba(59, 130, 246, 0.15)',
                        color: item.status === 'PUBLISHED' ? 'var(--status-active)' : item.status === 'PUBLISHING' ? 'var(--status-warning)' : '#60a5fa',
                        border: `1px solid ${item.status === 'PUBLISHED' ? 'var(--status-active)' : 'rgba(59, 130, 246, 0.3)'}`
                      }}>
                        {item.status || 'QUEUED'}
                      </span>
                    </td>
                    <td style={{ padding: '1rem 1.25rem', fontSize: '0.8rem', color: 'var(--text-tertiary)' }}>
                      {item.created_at ? new Date(item.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', month: 'short', day: 'numeric' }) : 'Recently'}
                    </td>
                    <td style={{ padding: '1rem 1.25rem', borderRadius: '0 14px 14px 0', textAlign: 'right' }}>
                      <span style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--accent-color)' }}>
                        Synced to GDrive ✓
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};
