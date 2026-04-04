import React, { useState } from 'react';
import { API_BASE } from './api';

interface SocialComposerModalProps {
  isOpen: boolean;
  onClose: () => void;
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
}

export const SocialComposerModal: React.FC<SocialComposerModalProps> = ({ isOpen, onClose, fetchWithAuth }) => {
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
      alert('Content published successfully to selected platforms!');
      onClose();
    } catch (err) {
      console.error('Posting error:', err);
      alert('Failed to publish content.');
    } finally {
      setIsPosting(false);
    }
  };

  return (
    <div className="modal-overlay" style={{ zIndex: 2000 }}>
      <div className="premium-modal glass-card" style={{ 
        maxWidth: '600px', 
        background: 'rgba(20, 20, 20, 0.95)', 
        backdropFilter: 'blur(30px)',
        border: '1px solid rgba(255, 24, 119, 0.1)'
      }}>
        <h2>Multi-Platform Composer</h2>
        <p style={{ color: 'var(--text-tertiary)', marginBottom: '1.5rem' }}>Draft once, publish across Meta.</p>
        
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
              background: 'rgba(255, 255, 255, 0.05)',
              border: '1px solid rgba(255, 255, 255, 0.1)',
              color: 'white',
              fontFamily: 'inherit',
              resize: 'none'
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
            style={{ background: 'linear-gradient(135deg, #10b981, #059669)' }}
          >
            {isPosting ? 'Publishing...' : 'Publish Now'}
          </button>
        </div>
      </div>
    </div>
  );
};
