import React, { useState } from 'react';
import { API_BASE } from './api';
import { useToast } from './ToastContext';

interface SocialComposerModalProps {
  isOpen: boolean;
  onClose: () => void;
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
}

export const SocialComposerModal: React.FC<SocialComposerModalProps> = ({ isOpen, onClose, fetchWithAuth }) => {
  const { success, error } = useToast();
  const [content, setContent] = useState('');
  const [isPosting, setIsPosting] = useState(false);
  const [selectedPlatforms, setSelectedPlatforms] = useState<string[]>(['instagram']);

  if (!isOpen) return null;

  const handlePost = async () => {
    setIsPosting(true);
    try {
      for (const platform of selectedPlatforms) {
        await fetchWithAuth(`${API_BASE}/api/marketing/smm/post`, {
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
      }
      success('Content published successfully to selected platforms!');
      onClose();
    } catch (err) {
      console.error('Posting error:', err);
      error('Failed to publish content.');
    } finally {
      setIsPosting(false);
    }
  };

  return (
    <div className="modal-overlay" style={{ zIndex: 2000 }}>
      <div className="premium-modal glass-card-premium" style={{ maxWidth: '600px', position: 'relative' }}>
        <button
          className="close-modal-btn"
          onClick={onClose}
          aria-label="Close modal"
          title="Close"
        >
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
        <h2>Multi-Platform Composer</h2>
        <p style={{ color: 'var(--text-secondary)', marginBottom: '1.5rem' }}>Draft once, publish across Meta.</p>
        
        <div style={{ marginBottom: '1.5rem' }}>
          <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 600 }}>Post Content</label>
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
          <div
            className="character-count"
            aria-live="polite"
          >
            {content.length} {content.length === 1 ? 'character' : 'characters'}
          </div>
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
      <style>{`
        .close-modal-btn {
          position: absolute;
          top: 1.5rem;
          right: 1.5rem;
          background: transparent;
          border: none;
          color: var(--text-tertiary);
          cursor: pointer;
          padding: 8px;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          transition: all 0.2s;
          z-index: 10;
        }
        .close-modal-btn:hover {
          color: var(--text-primary);
          background: var(--bg-hover);
        }
        .character-count {
          margin-top: 0.5rem;
          font-size: 0.75rem;
          color: var(--text-tertiary);
          text-align: right;
          font-weight: 500;
        }
      `}</style>
    </div>
  );
};
