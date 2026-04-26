import React, { useState } from 'react';
import { API_BASE } from './api';
import { useToast } from './ToastContext';

interface SocialComposerModalProps {
  isOpen: boolean;
  onClose: () => void;
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
}

export const SocialComposerModal: React.FC<SocialComposerModalProps> = ({ isOpen, onClose, fetchWithAuth }) => {
  const { success: toastSuccess, error: toastError } = useToast();
  const [content, setContent] = useState('');
  const [isPosting, setIsPosting] = useState(false);
  const [selectedPlatforms, setSelectedPlatforms] = useState<string[]>(['instagram']);

  if (!isOpen) return null;

  const handlePost = async () => {
    setIsPosting(true);
    try {
      for (const platform of selectedPlatforms) {
        const response = await fetchWithAuth(`${API_BASE}/api/marketing/smm/post`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            platform,
            message: content, // Facebook
            caption: content, // Instagram
            text: content,    // Threads
            image_url: 'https://placeholder.com/image.jpg' // Example required for IG
          })
        });
        if (!response.ok) throw new Error(`Failed to post to ${platform}`);
      }
      toastSuccess('Content published successfully to selected platforms!');
      onClose();
      setContent('');
    } catch (err) {
      console.error('Posting error:', err);
      toastError('Failed to publish content.');
    } finally {
      setIsPosting(false);
    }
  };

  return (
    <div className="modal-overlay" style={{ zIndex: 2000 }} onClick={onClose}>
      <div className="premium-modal glass-card-premium" style={{ maxWidth: '600px', position: 'relative' }} onClick={e => e.stopPropagation()}>
        <button
          onClick={onClose}
          aria-label="Close modal"
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
            justifyContent: 'center',
            transition: 'all 0.2s'
          }}
          onMouseEnter={(e) => (e.currentTarget.style.background = 'var(--bg-hover)')}
          onMouseLeave={(e) => (e.currentTarget.style.background = 'var(--bg-input)')}
        >
          ×
        </button>
        <h2>Multi-Platform Composer</h2>
        <p style={{ color: 'var(--text-secondary)', marginBottom: '1.5rem' }}>Draft once, publish across Meta.</p>
        
        <div style={{ marginBottom: '1.5rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
            <label style={{ fontWeight: 600 }}>Post Content</label>
            <span
              aria-live="polite"
              style={{
                fontSize: '0.8rem',
                color: content.length > 500 ? 'var(--status-danger)' : 'var(--text-tertiary)',
                fontWeight: content.length > 500 ? 700 : 400
              }}
            >
              {content.length} / 500
            </span>
          </div>
          <textarea 
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="What's on your mind? (Use #hashtags for IG/Threads)"
            style={{
              width: '100%',
              height: '150px',
              padding: '1rem',
              borderRadius: '12px',
              background: 'var(--bg-input)',
              border: '1px solid var(--border-color)',
              color: 'var(--text-primary)',
              fontFamily: 'inherit',
              resize: 'none',
              outline: 'none'
            }}
          />
        </div>

        <div style={{ marginBottom: '2rem' }}>
          <label style={{ display: 'block', marginBottom: '0.75rem', fontWeight: 600 }}>Destination Platforms</label>
          <div style={{ display: 'flex', gap: '1rem' }}>
            {['instagram', 'facebook', 'threads'].map(p => (
              <label key={p} style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', cursor: 'pointer' }}>
                <input 
                  type="checkbox" 
                  checked={selectedPlatforms.includes(p)}
                  onChange={(e) => {
                    if (e.target.checked) setSelectedPlatforms([...selectedPlatforms, p]);
                    else setSelectedPlatforms(selectedPlatforms.filter(x => x !== p));
                  }}
                />
                <span style={{ textTransform: 'capitalize' }}>{p}</span>
              </label>
            ))}
          </div>
        </div>

        <div className="modal-actions">
          <button className="btn-secondary" onClick={onClose}>Cancel</button>
          <button 
            className="btn-primary" 
            onClick={handlePost}
            disabled={isPosting || !content.trim() || selectedPlatforms.length === 0}
            style={{ background: 'linear-gradient(135deg, var(--status-active), var(--status-active-bg))', color: 'var(--status-active)' }}
          >
            {isPosting ? 'Publishing...' : 'Publish Now'}
          </button>
        </div>
      </div>
    </div>
  );
};
