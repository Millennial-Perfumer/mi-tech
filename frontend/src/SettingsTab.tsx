import { useState, useEffect } from 'react';

// Animation for collapsible sections
const SLIDE_IN_ANIMATION = `
  @keyframes slideIn {
    from { opacity: 0; transform: translateY(-8px); }
    to { opacity: 1; transform: translateY(0); }
  }
`;

const API_BASE = 'http://localhost:8080';

interface AppConfig {
  key: string;
  value: string;
  is_secret: boolean;
  label: string;
  category: string;
  sort_order: number;
}

interface SettingsTabProps {
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
}

const CATEGORY_META: Record<string, { title: string; icon: React.ReactNode; color: string }> = {
  shopify: {
    title: 'Shopify',
    icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><circle cx="9" cy="21" r="1"></circle><circle cx="20" cy="21" r="1"></circle><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"></path></svg>,
    color: '#10b981'
  },
  whatsapp: {
    title: 'WhatsApp',
    icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"></path></svg>,
    color: '#22c55e'
  },
  system: {
    title: 'System',
    icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>,
    color: '#6366f1'
  },
  business: {
    title: 'Business',
    icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path><polyline points="9 22 9 12 15 12 15 22"></polyline></svg>,
    color: '#f59e0b'
  }
};

export function SettingsTab({ fetchWithAuth }: SettingsTabProps) {
  // Configs state
  const [configs, setConfigs] = useState<AppConfig[]>([]);
  const [isLoadingConfigs, setIsLoadingConfigs] = useState(true);
  const [isRevealed, setIsRevealed] = useState(false);
  const [showPasswordModal, setShowPasswordModal] = useState(false);
  const [password, setPassword] = useState('');
  const [passwordError, setPasswordError] = useState('');
  const [editingKey, setEditingKey] = useState<string | null>(null);
  const [editValue, setEditValue] = useState('');
  const [isSavingConfig, setIsSavingConfig] = useState(false);
  const [expandedCategories, setExpandedCategories] = useState<Record<string, boolean>>({});
  const [notification, setNotification] = useState<{ message: string; type: 'success' | 'error' } | null>(null);

  const showNotification = (message: string, type: 'success' | 'error' = 'success') => {
    setNotification({ message, type });
    setTimeout(() => setNotification(null), 5000);
  };

  // Fetch configs on mount
  useEffect(() => {
    fetchConfigs();
  }, []);

  const fetchConfigs = async () => {
    setIsLoadingConfigs(true);
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/configs`);
      const data = await resp.json();
      if (data.success) {
        setConfigs(data.configs || []);
      } else {
        showNotification(data.message || 'Failed to load configurations', 'error');
      }
    } catch (err) {
      console.error('Failed to fetch configs:', err);
      showNotification('Failed to load configurations. Please check your connection.', 'error');
    } finally {
      setIsLoadingConfigs(false);
    }
  };

  const handleReveal = async () => {
    setPasswordError('');
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/configs/reveal`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ password })
      });
      const data = await resp.json();
      if (data.success) {
        setConfigs(data.configs || []);
        setIsRevealed(true);
        setShowPasswordModal(false);
        setPassword('');
      } else {
        setPasswordError(data.message || 'Incorrect password');
      }
    } catch (err) {
      setPasswordError('Network error');
    }
  };

  const handleHide = async () => {
    setIsRevealed(false);
    await fetchConfigs();
  };

  const handleStartEdit = (config: AppConfig) => {
    setEditingKey(config.key);
    setEditValue(config.value);
  };

  const handleCancelEdit = () => {
    setEditingKey(null);
    setEditValue('');
  };

  const handleSaveConfig = async (key: string, value: string) => {
    setIsSavingConfig(true);
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/configs`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ key, value })
      });
      const data = await resp.json();
      if (data.success) {
        setConfigs(prev => prev.map(c => c.key === key ? { ...c, value } : c));
        setEditingKey(null);
        setEditValue('');
        showNotification('Configuration updated successfully');
      } else {
        showNotification(data.message || 'Failed to save configuration', 'error');
      }
    } catch (err) {
      console.error('Failed to save config:', err);
      showNotification('Failed to save configuration. Network error.', 'error');
    } finally {
      setIsSavingConfig(false);
    }
  };

  const handleToggleConfig = async (cfg: AppConfig) => {
    const newValue = cfg.value === 'true' ? 'false' : 'true';
    await handleSaveConfig(cfg.key, newValue);
  };

  const toggleCategory = (cat: string) => {
    setExpandedCategories(prev => ({
      ...prev,
      [cat]: !prev[cat]
    }));
  };

  // Group configs by category
  const groupedConfigs: Record<string, AppConfig[]> = {};
  configs.forEach(cfg => {
    if (!groupedConfigs[cfg.category]) groupedConfigs[cfg.category] = [];
    groupedConfigs[cfg.category].push(cfg);
  });

  // Sort categories: Business > Shopify > WhatsApp > System
  const categoryOrder = ['business', 'shopify', 'whatsapp', 'system'];
  const sortedCategories = Object.keys(groupedConfigs).sort((a, b) => {
    const idxA = categoryOrder.indexOf(a);
    const idxB = categoryOrder.indexOf(b);
    return (idxA === -1 ? 99 : idxA) - (idxB === -1 ? 99 : idxB);
  });

  // Sort items within each category by sort_order
  sortedCategories.forEach(cat => {
    groupedConfigs[cat].sort((a, b) => a.sort_order - b.sort_order);
  });

  return (
    <div className="settings-container" style={{ padding: '0.5rem' }}>
      <style>{SLIDE_IN_ANIMATION}</style>
      <section className="card" style={{ padding: '2rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
          <div>
            <h2 style={{ fontSize: '1.25rem', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '0.5rem' }}>
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="var(--accent-color)" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"></path></svg>
              API Keys & Configuration
            </h2>
            <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem' }}>
              Manage Business details, API keys, and Feature configurations. Secret values are masked by default.
            </p>
          </div>
          <button
            className={isRevealed ? 'btn-secondary' : 'btn-primary'}
            onClick={() => {
              if (isRevealed) {
                handleHide();
              } else {
                setShowPasswordModal(true);
                setPassword('');
                setPasswordError('');
              }
            }}
            style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', padding: '0.6rem 1.25rem', fontSize: '0.85rem', fontWeight: 600, borderRadius: '10px', whiteSpace: 'nowrap' }}
          >
            {isRevealed ? (
              <>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path><line x1="1" y1="1" x2="23" y2="23"></line></svg>
                Hide Secrets
              </>
            ) : (
              <>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path><circle cx="12" cy="12" r="3"></circle></svg>
                Reveal Secrets
              </>
            )}
          </button>
        </div>

        {isLoadingConfigs ? (
          <div style={{ textAlign: 'center', padding: '2rem', color: '#64748b' }}>Loading configurations...</div>
        ) : configs.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '2rem', color: '#94a3b8' }}>No configurations found. Run the migration to initialize.</div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
            {sortedCategories.map(category => {
              const items = groupedConfigs[category];
              const meta = CATEGORY_META[category] || { title: category, icon: null, color: '#64748b' };
              const isExpanded = !!expandedCategories[category];

              return (
                <div key={category} style={{
                  border: '1px solid #f1f5f9',
                  borderRadius: '12px',
                  overflow: 'hidden',
                  background: '#ffffff'
                }}>
                  <div
                    onClick={() => toggleCategory(category)}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                      padding: '1rem 1.25rem',
                      background: !isExpanded ? '#ffffff' : '#f8fafc',
                      cursor: 'pointer',
                      transition: 'all 0.2s ease',
                      userSelect: 'none'
                    }}
                    onMouseEnter={e => e.currentTarget.style.background = '#f1f5f9'}
                    onMouseLeave={e => e.currentTarget.style.background = !isExpanded ? '#ffffff' : '#f8fafc'}
                  >
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                      <span style={{ color: meta.color, display: 'flex', alignItems: 'center' }}>{meta.icon}</span>
                      <span style={{ fontSize: '0.85rem', fontWeight: 700, color: '#1e293b', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                        {meta.title}
                      </span>
                      <span style={{ fontSize: '0.7rem', color: '#94a3b8', background: '#f1f5f9', padding: '1px 6px', borderRadius: '4px' }}>
                        {items.length}
                      </span>
                    </div>
                    <svg
                      width="18"
                      height="18"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="#94a3b8"
                      strokeWidth="2.5"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      style={{
                        transform: isExpanded ? 'rotate(180deg)' : 'rotate(0deg)',
                        transition: 'transform 0.3s cubic-bezier(0.4, 0, 0.2, 1)'
                      }}
                    >
                      <polyline points="6 9 12 15 18 9"></polyline>
                    </svg>
                  </div>

                  {isExpanded && (
                    <div style={{
                      display: 'flex',
                      flexDirection: 'column',
                      gap: '0.5rem',
                      padding: '1rem',
                      borderTop: '1px solid #f1f5f9',
                      animation: 'slideIn 0.2s ease-out'
                    }}>
                    {items.map(cfg => (
                      <div
                        key={cfg.key}
                        className="config-row"
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: '1rem',
                          padding: '0.75rem 1rem',
                          background: editingKey === cfg.key ? '#f0f9ff' : '#f8fafc',
                          borderRadius: '10px',
                          border: editingKey === cfg.key ? '1px solid var(--accent-color)' : '1px solid #e2e8f0',
                          transition: 'all 0.2s'
                        }}
                      >
                        <div style={{ minWidth: '160px', flexShrink: 0 }}>
                          <div style={{ fontSize: '0.85rem', fontWeight: 600, color: '#0f172a' }}>{cfg.label}</div>
                          <div style={{ fontSize: '0.7rem', color: '#94a3b8', fontFamily: 'monospace' }}>{cfg.key}</div>
                        </div>
                        
                        <div style={{ flex: 1, minWidth: 0 }}>
                          {(cfg.value === 'true' || cfg.value === 'false') && editingKey !== cfg.key ? (
                            <button 
                              onClick={() => handleToggleConfig(cfg)}
                              disabled={isSavingConfig}
                              style={{
                                position: 'relative',
                                width: '42px',
                                height: '22px',
                                borderRadius: '11px',
                                backgroundColor: cfg.value === 'true' ? 'var(--accent-color)' : '#cbd5e1',
                                border: 'none',
                                cursor: 'pointer',
                                transition: 'all 0.3s ease',
                                padding: '2px'
                              }}
                            >
                              <div style={{
                                width: '18px',
                                height: '18px',
                                borderRadius: '50%',
                                backgroundColor: 'white',
                                transform: cfg.value === 'true' ? 'translateX(20px)' : 'translateX(0)',
                                transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                                boxShadow: '0 1px 3px rgba(0,0,0,0.1)'
                              }} />
                            </button>
                          ) : editingKey === cfg.key ? (
                            <input
                              type="text"
                              value={editValue}
                              onChange={e => setEditValue(e.target.value)}
                              autoFocus
                              style={{
                                width: '100%',
                                padding: '0.5rem 0.75rem',
                                borderRadius: '6px',
                                border: '1px solid var(--accent-color)',
                                fontSize: '0.85rem',
                                fontFamily: 'monospace',
                                outline: 'none',
                                boxShadow: '0 0 0 3px rgba(14, 165, 233, 0.1)'
                              }}
                              onKeyDown={e => {
                                if (e.key === 'Enter') handleSaveConfig(cfg.key, editValue);
                                if (e.key === 'Escape') handleCancelEdit();
                              }}
                            />
                          ) : (
                            <div style={{
                              fontSize: '0.85rem',
                              fontFamily: 'monospace',
                              color: cfg.is_secret && !isRevealed ? '#94a3b8' : '#334155',
                              overflow: 'hidden',
                              textOverflow: 'ellipsis',
                              whiteSpace: 'nowrap',
                              letterSpacing: cfg.is_secret && !isRevealed ? '0.1em' : 'normal'
                            }}>
                              {cfg.value || <span style={{ color: '#cbd5e1', fontStyle: 'italic' }}>Not set</span>}
                            </div>
                          )}
                        </div>

                        <div style={{ display: 'flex', gap: '0.35rem', flexShrink: 0 }}>
                          {cfg.is_secret && (
                            <span style={{
                              fontSize: '0.65rem',
                              fontWeight: 700,
                              color: '#ef4444',
                              background: '#fef2f2',
                              padding: '2px 6px',
                              borderRadius: '4px',
                              border: '1px solid #fee2e2',
                              textTransform: 'uppercase',
                              letterSpacing: '0.05em'
                            }}>
                              Secret
                            </span>
                          )}
                          
                          {editingKey === cfg.key ? (
                            <>
                              <button
                                className="toolbar-btn"
                                title="Save"
                                disabled={isSavingConfig}
                                onClick={() => handleSaveConfig(cfg.key, editValue)}
                                style={{ color: '#10b981' }}
                              >
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>
                              </button>
                              <button
                                className="toolbar-btn"
                                title="Cancel"
                                onClick={handleCancelEdit}
                                style={{ color: '#ef4444' }}
                              >
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                              </button>
                            </>
                          ) : (
                            <button
                              className="toolbar-btn"
                              title="Edit"
                              onClick={() => {
                                if (cfg.is_secret && !isRevealed) {
                                  setShowPasswordModal(true);
                                  setPassword('');
                                  setPasswordError('');
                                } else {
                                  handleStartEdit(cfg);
                                }
                              }}
                            >
                              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
                            </button>
                          )}
                        </div>
                      </div>
                    ))}
                    {category === 'shopify' && (
                      <div
                        className="config-row"
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'space-between',
                          gap: '1rem',
                          padding: '1rem',
                          background: 'linear-gradient(to right, #f0f9ff, #e0f2fe)',
                          borderRadius: '10px',
                          border: '1px dashed #0ea5e9',
                          marginTop: '0.5rem'
                        }}
                      >
                        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                          <div style={{ 
                            width: '40px', 
                            height: '40px', 
                            borderRadius: '10px', 
                            background: '#0ea5e9', 
                            display: 'flex', 
                            alignItems: 'center', 
                            justifyContent: 'center',
                            color: 'white'
                          }}>
                            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="23 4 23 10 17 10"></polyline><polyline points="1 20 1 14 7 14"></polyline><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path></svg>
                          </div>
                          <div>
                            <div style={{ fontSize: '0.875rem', fontWeight: 700, color: '#0369a1' }}>Manual Order Synchronization</div>
                            <div style={{ fontSize: '0.75rem', color: '#64748b' }}>Fetch missing orders or update existing ones directly from Shopify.</div>
                          </div>
                        </div>
                        <button 
                          className="btn-primary" 
                          onClick={() => {
                            // Trigger the global manual sync button logic
                            const syncBtn = document.querySelector('button[title*="Manually fetch orders"]') as HTMLButtonElement;
                            if (syncBtn) syncBtn.click();
                          }}
                          style={{ padding: '0.5rem 1rem', fontSize: '0.85rem' }}
                        >
                          Launch Sync
                        </button>
                      </div>
                    )}
                  </div>
                )}
                </div>
              );
            })}
          </div>
        )}
      </section>

      {/* Password Modal */}
      {showPasswordModal && (
        <div className="modal-overlay" onClick={() => setShowPasswordModal(false)}>
          <div className="premium-modal" style={{ maxWidth: '380px' }} onClick={e => e.stopPropagation()}>
            <div className="modal-header-icon" style={{ background: 'linear-gradient(135deg, #ef4444, #f97316)' }}>
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>
            </div>
            <h2 style={{ fontSize: '1.25rem' }}>Enter Password</h2>
            <p style={{ fontSize: '0.85rem' }}>Enter your admin password to reveal secret values.</p>
            
            <input
              type="password"
              value={password}
              onChange={e => { setPassword(e.target.value); setPasswordError(''); }}
              placeholder="Admin password"
              autoFocus
              style={{ width: '100%', padding: '0.75rem 1rem', borderRadius: '10px', border: passwordError ? '2px solid #ef4444' : '2px solid #e2e8f0', fontSize: '0.95rem', outline: 'none', transition: 'border-color 0.2s' }}
              onKeyDown={e => { if (e.key === 'Enter') handleReveal(); }}
            />
            {passwordError && (
              <div style={{ color: '#ef4444', fontSize: '0.8rem', marginTop: '0.5rem', fontWeight: 600 }}>
                {passwordError}
              </div>
            )}
            
            <div className="modal-actions" style={{ marginTop: '1.5rem' }}>
              <button className="btn-secondary" onClick={() => setShowPasswordModal(false)}>Cancel</button>
              <button className="btn-primary" onClick={handleReveal} disabled={!password}>Unlock</button>
            </div>
          </div>
        </div>
      )}
      {/* Notifications */}
      {notification && (
        <div style={{
          position: 'fixed',
          bottom: '2rem',
          right: '2rem',
          padding: '1rem 1.5rem',
          background: notification.type === 'error' ? '#fef2f2' : '#f0fdf4',
          border: notification.type === 'error' ? '1px solid #fee2e2' : '1px solid #dcfce7',
          borderRadius: '12px',
          boxShadow: '0 10px 15px -3px rgba(0, 0, 0, 0.1)',
          display: 'flex',
          alignItems: 'center',
          gap: '0.75rem',
          color: notification.type === 'error' ? '#991b1b' : '#166534',
          zIndex: 3000,
          animation: 'slideIn 0.3s ease-out',
          fontWeight: 600,
          fontSize: '0.9rem'
        }}>
          {notification.type === 'error' ? (
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
          ) : (
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>
          )}
          {notification.message}
        </div>
      )}
    </div>
  );
}
