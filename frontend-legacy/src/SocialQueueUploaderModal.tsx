import React, { useState, useEffect, useRef } from 'react';
import { API_BASE } from './api';
import { useToast } from './ToastContext';

interface SocialQueueUploaderModalProps {
  isOpen: boolean;
  gdriveFolderUrl?: string;
  onClose: () => void;
  onSuccess: () => void;
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
}

export const SocialQueueUploaderModal: React.FC<SocialQueueUploaderModalProps> = ({
  isOpen,
  onClose,
  onSuccess,
  fetchWithAuth,
}) => {
  const { success: toastSuccess, error: toastError } = useToast();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [caption, setCaption] = useState('');
  const [hashtags, setHashtags] = useState('#automation #growth #tech');
  const [selectedPlatforms, setSelectedPlatforms] = useState<string[]>(['instagram', 'facebook', 'threads', 'x']);
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [folderPreview, setFolderPreview] = useState('');
  const [isDragging, setIsDragging] = useState(false);

  useEffect(() => {
    if (isOpen) {
      const now = new Date();
      const pad = (n: number) => n.toString().padStart(2, '0');
      const timeStr = `${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}_${pad(now.getHours())}${pad(now.getMinutes())}${pad(now.getSeconds())}`;
      setFolderPreview(`Post_${timeStr}`);
    }
  }, [isOpen]);

  if (!isOpen) return null;

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const filesArray = Array.from(e.target.files);
      setSelectedFiles(prev => [...prev, ...filesArray]);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      const filesArray = Array.from(e.dataTransfer.files);
      setSelectedFiles(prev => [...prev, ...filesArray]);
    }
  };

  const removeFile = (index: number) => {
    setSelectedFiles(prev => prev.filter((_, i) => i !== index));
  };

  const getDetectedPostType = () => {
    if (selectedFiles.length === 0) return 'SINGLE_PHOTO';
    const isVideo = selectedFiles.some(f => f.type.startsWith('video/'));
    if (isVideo) return 'VIDEO';
    if (selectedFiles.length > 1) return 'CAROUSEL';
    return 'SINGLE_PHOTO';
  };

  const togglePlatform = (platformId: string) => {
    if (selectedPlatforms.includes(platformId)) {
      setSelectedPlatforms(selectedPlatforms.filter(p => p !== platformId));
    } else {
      setSelectedPlatforms([...selectedPlatforms, platformId]);
    }
  };

  const handleSubmit = async () => {
    if (!caption.trim()) {
      toastError('Please enter a caption description.');
      return;
    }
    if (selectedPlatforms.length === 0) {
      toastError('Please select at least one platform.');
      return;
    }

    setIsSubmitting(true);
    try {
      const postType = getDetectedPostType();
      const formData = new FormData();
      formData.append('caption', caption);
      formData.append('hashtags', hashtags);
      formData.append('post_type', postType);
      formData.append('target_platforms', JSON.stringify(selectedPlatforms));

      selectedFiles.forEach(file => {
        formData.append('files', file);
      });

      const response = await fetchWithAuth(`${API_BASE}/api/marketing/smm/queue`, {
        method: 'POST',
        body: formData,
      });

      const data = await response.json();
      if (data.success) {
        const createdFolderName = data.post?.folder_name || folderPreview;
        toastSuccess(`Successfully queued post ${createdFolderName}!`);
        onSuccess();
        onClose();
        setCaption('');
        setSelectedFiles([]);
      } else {
        toastError(data.error || 'Failed to queue post to Google Drive.');
      }
    } catch (err) {
      console.error('Queue error:', err);
      toastError('Failed to upload post to Google Drive Queue.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const postTypeLabel = getDetectedPostType();

  return (
    <div className="modal-overlay" style={{ zIndex: 2000, background: 'rgba(0, 0, 0, 0.65)', backdropFilter: 'blur(8px)' }} onClick={onClose}>
      <div
        className="premium-modal glass-card-premium"
        style={{
          maxWidth: '640px',
          width: '90%',
          position: 'relative',
          padding: '2.25rem',
          borderRadius: '24px',
          background: 'var(--surface-color)',
          border: '1px solid var(--border-color)',
          boxShadow: 'var(--shadow-lg)',
          animation: 'fadeInUp 0.3s cubic-bezier(0.16, 1, 0.3, 1)',
          maxHeight: '90vh',
          overflowY: 'auto'
        }}
        onClick={e => e.stopPropagation()}
      >
        {/* Close Button */}
        <button
          onClick={onClose}
          aria-label="Close modal"
          style={{
            position: 'absolute',
            top: '1.5rem',
            right: '1.5rem',
            background: 'var(--bg-input)',
            border: '1px solid var(--border-color)',
            borderRadius: '50%',
            width: '36px',
            height: '36px',
            color: 'var(--text-secondary)',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            transition: 'all 0.2s ease',
          }}
          onMouseEnter={(e) => {
            e.currentTarget.style.color = 'var(--text-primary)';
            e.currentTarget.style.background = 'var(--bg-hover)';
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.color = 'var(--text-secondary)';
            e.currentTarget.style.background = 'var(--bg-input)';
          }}
        >
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>

        {/* Modal Header */}
        <div style={{ marginBottom: '1.5rem' }}>
          <h2 style={{ margin: '0 0 0.4rem 0', fontSize: '1.5rem', fontWeight: 800, color: 'var(--text-primary)', letterSpacing: '-0.02em' }}>
            Google Drive Auto-Queue Uploader
          </h2>
          <p style={{ margin: 0, color: 'var(--text-secondary)', fontSize: '0.875rem', lineHeight: '1.5' }}>
            Upload media and captions directly into your Google Drive queue. n8n auto-publishes to your social accounts on schedule.
          </p>
        </div>

        {/* Auto-Generated Folder Banner */}
        <div style={{
          padding: '1rem 1.25rem',
          borderRadius: '16px',
          background: 'rgba(16, 185, 129, 0.04)',
          border: '1px solid rgba(16, 185, 129, 0.2)',
          borderLeft: '4px solid var(--accent-color)',
          marginBottom: '1.5rem',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center'
        }}>
          <div>
            <span style={{ fontSize: '0.7rem', color: 'var(--text-tertiary)', textTransform: 'uppercase', fontWeight: 700, letterSpacing: '0.05em' }}>
              Auto-Generated GDrive Folder
            </span>
            <div style={{ fontSize: '1rem', fontWeight: 800, color: 'var(--accent-color)', fontFamily: 'monospace', marginTop: '0.2rem' }}>
              {folderPreview}
            </div>
          </div>
          <span style={{
            padding: '0.35rem 0.8rem',
            borderRadius: '20px',
            fontSize: '0.75rem',
            fontWeight: 800,
            letterSpacing: '0.03em',
            background: postTypeLabel === 'CAROUSEL' ? 'rgba(168, 85, 247, 0.15)' : postTypeLabel === 'VIDEO' ? 'rgba(239, 68, 68, 0.15)' : 'rgba(16, 185, 129, 0.15)',
            color: postTypeLabel === 'CAROUSEL' ? '#c084fc' : postTypeLabel === 'VIDEO' ? '#f87171' : '#34d399',
            border: `1px solid ${postTypeLabel === 'CAROUSEL' ? 'rgba(168, 85, 247, 0.3)' : postTypeLabel === 'VIDEO' ? 'rgba(239, 68, 68, 0.3)' : 'rgba(16, 185, 129, 0.3)'}`
          }}>
            {postTypeLabel}
          </span>
        </div>

        {/* Styled Drag and Drop File Upload Dropzone */}
        <div style={{ marginBottom: '1.5rem' }}>
          <label style={{ display: 'block', fontWeight: 700, marginBottom: '0.6rem', fontSize: '0.875rem', color: 'var(--text-primary)' }}>
            Media Files (Photos / Carousel / Video)
          </label>

          <input
            ref={fileInputRef}
            type="file"
            multiple
            accept="image/*,video/*"
            onChange={handleFileChange}
            style={{ display: 'none' }}
          />

          <div
            onDragOver={(e) => { e.preventDefault(); setIsDragging(true); }}
            onDragLeave={() => setIsDragging(false)}
            onDrop={handleDrop}
            onClick={() => fileInputRef.current?.click()}
            style={{
              padding: '1.75rem 1.25rem',
              borderRadius: '16px',
              border: `2px dashed ${isDragging ? 'var(--accent-color)' : 'var(--border-color)'}`,
              background: isDragging ? 'rgba(16, 185, 129, 0.05)' : 'var(--bg-input)',
              textAlign: 'center',
              cursor: 'pointer',
              transition: 'all 0.2s ease',
              boxShadow: isDragging ? '0 0 0 4px rgba(16, 185, 129, 0.1)' : 'none'
            }}
          >
            <div style={{
              width: '44px',
              height: '44px',
              borderRadius: '12px',
              background: 'var(--bg-card)',
              border: '1px solid var(--border-color)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              margin: '0 auto 0.75rem auto',
              color: 'var(--accent-color)'
            }}>
              <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                <polyline points="17 8 12 3 7 8"></polyline>
                <line x1="12" y1="3" x2="12" y2="15"></line>
              </svg>
            </div>
            <div style={{ fontSize: '0.9rem', fontWeight: 700, color: 'var(--text-primary)', marginBottom: '0.25rem' }}>
              Click or drag media files here to upload
            </div>
            <div style={{ fontSize: '0.78rem', color: 'var(--text-tertiary)' }}>
              Supports JPG, PNG, WEBP, MP4, MOV (Multiple photos automatically trigger Carousel mode)
            </div>
          </div>

          {/* Selected File Chips List */}
          {selectedFiles.length > 0 && (
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem', marginTop: '0.75rem' }}>
              {selectedFiles.map((file, idx) => (
                <div
                  key={idx}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.5rem',
                    padding: '0.4rem 0.8rem',
                    borderRadius: '10px',
                    background: 'var(--bg-card)',
                    border: '1px solid var(--border-color)',
                    fontSize: '0.8rem',
                    fontWeight: 600,
                    color: 'var(--text-primary)'
                  }}
                >
                  <span style={{ maxWidth: '180px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {file.name}
                  </span>
                  <span style={{ color: 'var(--text-tertiary)', fontSize: '0.7rem' }}>
                    ({(file.size / (1024 * 1024)).toFixed(1)}MB)
                  </span>
                  <button
                    onClick={(e) => { e.stopPropagation(); removeFile(idx); }}
                    style={{
                      background: 'none',
                      border: 'none',
                      color: '#ef4444',
                      cursor: 'pointer',
                      padding: '2px',
                      display: 'flex',
                      alignItems: 'center'
                    }}
                    title="Remove file"
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                      <line x1="18" y1="6" x2="6" y2="18"></line>
                      <line x1="6" y1="6" x2="18" y2="18"></line>
                    </svg>
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Caption Field */}
        <div style={{ marginBottom: '1.25rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
            <label htmlFor="queue-caption" style={{ fontWeight: 700, fontSize: '0.875rem', color: 'var(--text-primary)' }}>
              Caption Description
            </label>
            <span style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>
              {caption.length} chars
            </span>
          </div>
          <textarea
            id="queue-caption"
            value={caption}
            onChange={(e) => setCaption(e.target.value)}
            placeholder="Write your post caption here... (This will be saved as caption.txt inside GDrive folder)"
            rows={4}
            style={{
              width: '100%',
              padding: '0.875rem 1rem',
              borderRadius: '14px',
              background: 'var(--bg-input)',
              border: '1px solid var(--border-color)',
              color: 'var(--text-primary)',
              resize: 'none',
              fontFamily: 'inherit',
              fontSize: '0.875rem',
              lineHeight: '1.5',
              outline: 'none',
              transition: 'border-color 0.2s ease, box-shadow 0.2s ease'
            }}
            onFocus={(e) => {
              e.currentTarget.style.borderColor = 'var(--accent-color)';
              e.currentTarget.style.boxShadow = '0 0 0 3px rgba(16, 185, 129, 0.15)';
            }}
            onBlur={(e) => {
              e.currentTarget.style.borderColor = 'var(--border-color)';
              e.currentTarget.style.boxShadow = 'none';
            }}
          />
        </div>

        {/* Hashtags Field */}
        <div style={{ marginBottom: '1.5rem' }}>
          <label htmlFor="queue-hashtags" style={{ display: 'block', fontWeight: 700, marginBottom: '0.5rem', fontSize: '0.875rem', color: 'var(--text-primary)' }}>
            Hashtags
          </label>
          <input
            id="queue-hashtags"
            type="text"
            value={hashtags}
            onChange={(e) => setHashtags(e.target.value)}
            placeholder="#hashtags #separated #by #spaces"
            style={{
              width: '100%',
              padding: '0.75rem 1rem',
              borderRadius: '12px',
              background: 'var(--bg-input)',
              border: '1px solid var(--border-color)',
              color: 'var(--text-primary)',
              fontSize: '0.875rem',
              outline: 'none',
              transition: 'border-color 0.2s ease, box-shadow 0.2s ease'
            }}
            onFocus={(e) => {
              e.currentTarget.style.borderColor = 'var(--accent-color)';
              e.currentTarget.style.boxShadow = '0 0 0 3px rgba(16, 185, 129, 0.15)';
            }}
            onBlur={(e) => {
              e.currentTarget.style.borderColor = 'var(--border-color)';
              e.currentTarget.style.boxShadow = 'none';
            }}
          />
        </div>

        {/* Modern Target Platforms Toggle Cards */}
        <div style={{ marginBottom: '2rem' }}>
          <label style={{ display: 'block', fontWeight: 700, marginBottom: '0.75rem', fontSize: '0.875rem', color: 'var(--text-primary)' }}>
            Target Platforms
          </label>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(130px, 1fr))', gap: '0.75rem' }}>
            {[
              { id: 'instagram', label: 'Instagram' },
              { id: 'facebook', label: 'Facebook' },
              { id: 'threads', label: 'Threads' },
              { id: 'x', label: 'X (Twitter)' }
            ].map((p) => {
              const isSelected = selectedPlatforms.includes(p.id);
              return (
                <div
                  key={p.id}
                  onClick={() => togglePlatform(p.id)}
                  style={{
                    padding: '0.75rem 1rem',
                    borderRadius: '14px',
                    border: `1.5px solid ${isSelected ? 'var(--accent-color)' : 'var(--border-color)'}`,
                    background: isSelected ? 'rgba(16, 185, 129, 0.08)' : 'var(--bg-input)',
                    color: isSelected ? 'var(--accent-color)' : 'var(--text-secondary)',
                    fontWeight: 700,
                    fontSize: '0.85rem',
                    cursor: 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    transition: 'all 0.2s ease',
                    boxShadow: isSelected ? '0 4px 12px rgba(16, 185, 129, 0.15)' : 'none'
                  }}
                >
                  <span>{p.label}</span>
                  <div style={{
                    width: '18px',
                    height: '18px',
                    borderRadius: '50%',
                    background: isSelected ? 'var(--accent-color)' : 'transparent',
                    border: `1.5px solid ${isSelected ? 'var(--accent-color)' : 'var(--border-color)'}`,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    color: 'white',
                    fontSize: '10px'
                  }}>
                    {isSelected && (
                      <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="4" strokeLinecap="round" strokeLinejoin="round">
                        <polyline points="20 6 9 17 4 12"></polyline>
                      </svg>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Modal Actions */}
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', borderTop: '1px solid var(--border-color)', paddingTop: '1.5rem' }}>
          <button
            className="btn-secondary"
            onClick={onClose}
            disabled={isSubmitting}
            style={{
              padding: '0.75rem 1.5rem',
              borderRadius: '12px',
              fontWeight: 700,
              fontSize: '0.9rem'
            }}
          >
            Cancel
          </button>
          <button
            className="btn-primary hover-lift"
            onClick={handleSubmit}
            disabled={isSubmitting || !caption.trim()}
            style={{
              background: 'linear-gradient(135deg, var(--accent-color), #059669)',
              color: '#ffffff',
              padding: '0.75rem 1.75rem',
              borderRadius: '12px',
              fontWeight: 800,
              fontSize: '0.9rem',
              border: 'none',
              cursor: isSubmitting || !caption.trim() ? 'not-allowed' : 'pointer',
              opacity: isSubmitting || !caption.trim() ? 0.6 : 1,
              boxShadow: '0 8px 20px rgba(16, 185, 129, 0.3)',
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem'
            }}
          >
            {isSubmitting ? 'Uploading to GDrive...' : 'Queue to Google Drive'}
          </button>
        </div>
      </div>
    </div>
  );
};
