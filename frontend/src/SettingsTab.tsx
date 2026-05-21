import { API_BASE } from './api';
import { useState, useEffect, useRef } from 'react';
import { createPortal } from 'react-dom';
import { useToast } from './ToastContext';
import { useConfirm } from './ConfirmContext';

// Animation for collapsible sections
const SLIDE_IN_ANIMATION = `
  @keyframes slideIn {
    from { opacity: 0; transform: translateY(-8px); }
    to { opacity: 1; transform: translateY(0); }
  }
`;


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
  meta_shared: {
    title: 'Meta Shared Details',
    icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path></svg>,
    color: '#0ea5e9'
  },
  shopify: {
    title: 'Shopify',
    icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><circle cx="9" cy="21" r="1"></circle><circle cx="20" cy="21" r="1"></circle><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"></path></svg>,
    color: 'var(--status-active)'
  },
  whatsapp: {
    title: 'WhatsApp',
    icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"></path></svg>,
    color: 'var(--status-active)'
  },
  system: {
    title: 'System',
    icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>,
    color: 'var(--accent-color)'
  },
  business: {
    title: 'Business',
    icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path><polyline points="9 22 9 12 15 12 15 22"></polyline></svg>,
    color: 'var(--status-warning)'
  },
  marketing: {
    title: 'Marketing',
    icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M12 1v22M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>,
    color: '#0ea5e9'
  },
  social_media: {
    title: 'Social Media',
    icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M18 2h-3a5 5 0 0 0-5 5v3H7v4h3v8h4v-8h3l1-4h-4V7a1 1 0 0 1 1-1h3z"></path></svg>,
    color: '#E4405F'
  },
  feedback: {
    title: 'Feedback',
    icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path><path d="M9 10L11 12L15 8"></path></svg>,
    color: '#f59e0b'
  },
  amazon: {
    title: 'Amazon SP-API',
    icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path><polyline points="3.27 6.96 12 12.01 20.73 6.96"></polyline><line x1="12" y1="22.08" x2="12" y2="12"></line></svg>,
    color: '#FF9900'
  },
  inventory: {
    title: 'Inventory',
    icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><rect x="2" y="7" width="20" height="14" rx="2" ry="2"></rect><path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"></path></svg>,
    color: '#10b981'
  }
};

const formatTime12h = (timeStr: string) => {
  if (!timeStr) return '';
  const parts = timeStr.split(':');
  if (parts.length < 2) return timeStr;
  const hrs = parseInt(parts[0], 10);
  const mins = parseInt(parts[1], 10);
  if (isNaN(hrs) || isNaN(mins)) return timeStr;
  const ampm = hrs >= 12 ? 'PM' : 'AM';
  const displayHrs = hrs % 12 === 0 ? 12 : hrs % 12;
  const displayMins = mins < 10 ? `0${mins}` : mins;
  return `${displayHrs}:${displayMins} ${ampm}`;
};

const formatDateNice = (dateStr: string) => {
  if (!dateStr) return '';
  try {
    const date = new Date(dateStr);
    if (isNaN(date.getTime())) return dateStr;
    return date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
  } catch (e) {
    return dateStr;
  }
};

interface CustomTimePickerProps {
  value: string;
  onChange: (newValue: string) => void;
  triggerRef: React.RefObject<HTMLDivElement | null>;
}

function CustomTimePicker({ value, onChange, triggerRef }: CustomTimePickerProps) {
  const [hours24, minutesStr] = (value || "10:00").split(':');
  let initialHour = parseInt(hours24 || "10", 10);
  const initialMinute = parseInt(minutesStr || "00", 10);
  
  const initialPeriod = initialHour >= 12 ? 'PM' : 'AM';
  let initialHour12 = initialHour % 12;
  if (initialHour12 === 0) initialHour12 = 12;

  const [selectedHour, setSelectedHour] = useState(initialHour12);
  const [selectedMinute, setSelectedMinute] = useState(initialMinute);
  const [selectedPeriod, setSelectedPeriod] = useState(initialPeriod);

  const hourRef = useRef<HTMLDivElement>(null);
  const minuteRef = useRef<HTMLDivElement>(null);

  const [coords, setCoords] = useState({ top: 0, left: 0 });

  // Sync state changes to HH:MM format
  useEffect(() => {
    let hour24 = selectedHour;
    if (selectedPeriod === 'PM') {
      if (hour24 !== 12) hour24 += 12;
    } else {
      if (hour24 === 12) hour24 = 0;
    }
    const hh = String(hour24).padStart(2, '0');
    const mm = String(selectedMinute).padStart(2, '0');
    onChange(`${hh}:${mm}`);
  }, [selectedHour, selectedMinute, selectedPeriod]);

  // Center the selected elements on mount
  useEffect(() => {
    const timer = setTimeout(() => {
      if (hourRef.current) {
        const activeEl = hourRef.current.querySelector('[data-selected="true"]');
        if (activeEl) {
          hourRef.current.scrollTop = (activeEl as HTMLElement).offsetTop - 55;
        }
      }
      if (minuteRef.current) {
        const activeEl = minuteRef.current.querySelector('[data-selected="true"]');
        if (activeEl) {
          minuteRef.current.scrollTop = (activeEl as HTMLElement).offsetTop - 55;
        }
      }
    }, 80);
    return () => clearTimeout(timer);
  }, []);

  // Update popup position dynamically to overlay correctly on the body
  useEffect(() => {
    const updatePosition = () => {
      if (triggerRef.current) {
        const rect = triggerRef.current.getBoundingClientRect();
        setCoords({
          top: rect.bottom + window.scrollY,
          left: rect.left + window.scrollX
        });
      }
    };

    updatePosition();

    // Listen on capture scroll to catch scrolling inside any parent element
    window.addEventListener('scroll', updatePosition, true);
    window.addEventListener('resize', updatePosition);

    return () => {
      window.removeEventListener('scroll', updatePosition, true);
      window.removeEventListener('resize', updatePosition);
    };
  }, [triggerRef]);

  const hoursList = Array.from({ length: 12 }, (_, i) => i + 1);
  const minutesList = Array.from({ length: 60 }, (_, i) => i);

  return createPortal(
    <div className="custom-time-picker-dropdown" style={{
      display: 'flex',
      flexDirection: 'column',
      background: 'var(--surface-color)',
      border: '1px solid var(--border-color)',
      borderRadius: '14px',
      boxShadow: 'var(--shadow-lg)',
      padding: '0.875rem',
      width: '260px',
      userSelect: 'none',
      zIndex: 9999,
      position: 'absolute',
      top: `${coords.top + 4}px`,
      left: `${coords.left}px`,
      animation: 'slideIn 0.2s ease-out'
    }}>
      <style>{`
        .time-picker-column::-webkit-scrollbar {
          width: 0px;
          background: transparent;
        }
        .time-picker-column {
          scrollbar-width: none;
          -ms-overflow-style: none;
        }
      `}</style>
      
      {/* Header display */}
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        gap: '0.35rem',
        fontSize: '1.35rem',
        fontWeight: 700,
        color: 'var(--text-primary)',
        paddingBottom: '0.5rem',
        borderBottom: '1px solid var(--border-color)',
        marginBottom: '0.5rem'
      }}>
        <span>{String(selectedHour).padStart(2, '0')}</span>
        <span style={{ color: 'var(--text-tertiary)' }}>:</span>
        <span>{String(selectedMinute).padStart(2, '0')}</span>
        <span style={{ fontSize: '0.8rem', color: 'var(--accent-color)', marginLeft: '0.25rem', fontWeight: 800 }}>
          {selectedPeriod}
        </span>
      </div>

      {/* Selectors */}
      <div style={{ display: 'flex', gap: '0.5rem', height: '140px' }}>
        {/* Hours */}
        <div ref={hourRef} className="time-picker-column" style={{
          flex: 1,
          overflowY: 'auto',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          gap: '2px',
          paddingBottom: '60px'
        }}>
          <div style={{ fontSize: '0.6rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase', marginBottom: '4px', position: 'sticky', top: 0, background: 'var(--surface-color)', width: '100%', textAlign: 'center', padding: '2px 0' }}>HR</div>
          {hoursList.map(h => {
            const isSelected = selectedHour === h;
            return (
              <button
                key={h}
                data-selected={isSelected}
                onClick={() => setSelectedHour(h)}
                style={{
                  width: '100%',
                  padding: '4px 0',
                  borderRadius: '6px',
                  fontSize: '0.8rem',
                  fontWeight: isSelected ? 700 : 500,
                  color: isSelected ? 'white' : 'var(--text-secondary)',
                  background: isSelected ? 'var(--accent-color)' : 'transparent',
                  transition: 'all 0.15s ease',
                  textAlign: 'center',
                  minHeight: '28px'
                }}
              >
                {String(h).padStart(2, '0')}
              </button>
            );
          })}
        </div>

        {/* Minutes */}
        <div ref={minuteRef} className="time-picker-column" style={{
          flex: 1,
          overflowY: 'auto',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          gap: '2px',
          paddingBottom: '60px'
        }}>
          <div style={{ fontSize: '0.6rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase', marginBottom: '4px', position: 'sticky', top: 0, background: 'var(--surface-color)', width: '100%', textAlign: 'center', padding: '2px 0' }}>MIN</div>
          {minutesList.map(m => {
            const isSelected = selectedMinute === m;
            return (
              <button
                key={m}
                data-selected={isSelected}
                onClick={() => setSelectedMinute(m)}
                style={{
                  width: '100%',
                  padding: '4px 0',
                  borderRadius: '6px',
                  fontSize: '0.8rem',
                  fontWeight: isSelected ? 700 : 500,
                  color: isSelected ? 'white' : 'var(--text-secondary)',
                  background: isSelected ? 'var(--accent-color)' : 'transparent',
                  transition: 'all 0.15s ease',
                  textAlign: 'center',
                  minHeight: '28px'
                }}
              >
                {String(m).padStart(2, '0')}
              </button>
            );
          })}
        </div>

        {/* Period */}
        <div style={{
          flex: 1,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          gap: '4px'
        }}>
          <div style={{ fontSize: '0.6rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase', marginBottom: '4px', width: '100%', textAlign: 'center', padding: '2px 0' }}>AM/PM</div>
          {['AM', 'PM'].map(p => {
            const isSelected = selectedPeriod === p;
            return (
              <button
                key={p}
                onClick={() => setSelectedPeriod(p)}
                style={{
                  width: '100%',
                  padding: '6px 0',
                  borderRadius: '6px',
                  fontSize: '0.8rem',
                  fontWeight: isSelected ? 700 : 500,
                  color: isSelected ? 'white' : 'var(--text-secondary)',
                  background: isSelected ? 'var(--accent-color)' : 'transparent',
                  transition: 'all 0.15s ease',
                  textAlign: 'center',
                  minHeight: '30px'
                }}
              >
                {p}
              </button>
            );
          })}
        </div>
      </div>
    </div>,
    document.body
  );
}

export function SettingsTab({ fetchWithAuth }: SettingsTabProps) {
  const { success: toastSuccess, error: toastError } = useToast();
  const { confirm } = useConfirm();
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
  const timeTriggerRef = useRef<HTMLDivElement>(null);

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
        toastError(data.message || 'Failed to load configurations');
      }
    } catch (err) {
      console.error('Failed to fetch configs:', err);
      toastError('Failed to load configurations. Please check your connection.');
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
        toastSuccess('Configuration updated successfully');
      } else {
        toastError(data.message || 'Failed to save configuration');
      }
    } catch (err) {
      console.error('Failed to save config:', err);
      toastError('Failed to save configuration. Network error.');
    } finally {
      setIsSavingConfig(false);
    }
  };

  const handleToggleConfig = async (cfg: AppConfig) => {
    const newValue = cfg.value === 'true' ? 'false' : 'true';
    await handleSaveConfig(cfg.key, newValue);
  };

  const handleResetInventory = async () => {
    const confirmed = await confirm({
      title: 'Reset Physical Inventory',
      message: 'Are you sure you want to wipe the entire physical inventory and all mappings? This cannot be undone.',
      variant: 'danger',
      confirmLabel: 'Wipe Warehouse'
    });
    if (!confirmed) return;
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/inventory`, { method: 'DELETE' });
      if (resp.ok) {
        toastSuccess('Warehouse reset complete');
      } else {
        toastError('Failed to reset warehouse');
      }
    } catch (err) {
      toastError('Error resetting warehouse');
    }
  };

  const handleResetOrders = async () => {
    const confirmed = await confirm({
      title: 'Wipe Historical Data',
      message: 'Are you sure you want to delete all historical synced data? This will clear all orders and customers and cannot be undone. You will need to manually sync to recover data.',
      variant: 'danger',
      confirmLabel: 'Wipe Shopify Orders'
    });
    if (!confirmed) return;
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/shopify/reset`, { method: 'POST' });
      const data = await resp.json();
      if (data.success) {
        toastSuccess('Database wiped and re-sync triggered successfully.');
      } else {
        toastError('Failed to reset orders.');
      }
    } catch (err) {
      toastError('Error occurred while resetting.');
    }
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

  // Sort categories: Business > Shopify > Amazon > Inventory > Meta Shared > Marketing > Social Media > WhatsApp > System
  const categoryOrder = ['business', 'shopify', 'amazon', 'inventory', 'meta_shared', 'marketing', 'social_media', 'whatsapp', 'feedback', 'system'];
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
          <div style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-tertiary)' }}>Loading configurations...</div>
        ) : configs.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-tertiary)' }}>No configurations found. Run the migration to initialize.</div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
            {sortedCategories.map(category => {
              const items = groupedConfigs[category];
              const meta = CATEGORY_META[category] || { title: category, icon: null, color: 'var(--text-secondary)' };
              const isExpanded = !!expandedCategories[category];

              return (
                <div key={category} style={{
                  border: '1px solid var(--border-color)',
                  borderRadius: '12px',
                  overflow: 'hidden',
                  background: 'var(--surface-color)'
                }}>
                  <div
                    onClick={() => toggleCategory(category)}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                      padding: '1rem 1.25rem',
                      background: !isExpanded ? 'var(--surface-color)' : 'var(--bg-hover)',
                      cursor: 'pointer',
                      transition: 'all 0.2s ease',
                      userSelect: 'none'
                    }}
                    onMouseEnter={e => e.currentTarget.style.background = 'var(--bg-hover)'}
                    onMouseLeave={e => e.currentTarget.style.background = !isExpanded ? 'var(--surface-color)' : 'var(--bg-hover)'}
                  >
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                      <span style={{ color: meta.color, display: 'flex', alignItems: 'center' }}>{meta.icon}</span>
                      <span style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-primary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                        {meta.title}
                      </span>
                      <span style={{ fontSize: '0.7rem', color: 'var(--text-tertiary)', background: 'var(--bg-hover)', padding: '1px 6px', borderRadius: '4px' }}>
                        {items.length}
                      </span>
                    </div>
                    <svg
                      width="18"
                      height="18"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="var(--text-tertiary)"
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
                      borderTop: '1px solid var(--border-color)',
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
                          background: editingKey === cfg.key ? 'var(--accent-subtle)' : 'var(--bg-input)',
                          borderRadius: '10px',
                          border: editingKey === cfg.key ? '1px solid var(--accent-color)' : '1px solid var(--border-color)',
                          transition: 'all 0.2s'
                        }}
                      >
                        <div style={{ minWidth: '160px', flexShrink: 0 }}>
                          <div style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-primary)' }}>{cfg.label}</div>
                          <div style={{ fontSize: '0.7rem', color: 'var(--text-tertiary)', fontFamily: 'monospace' }}>{cfg.key}</div>
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
                                backgroundColor: cfg.value === 'true' ? 'var(--accent-color)' : 'var(--border-color)',
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
                                backgroundColor: 'var(--text-primary)',
                                transform: cfg.value === 'true' ? 'translateX(20px)' : 'translateX(0)',
                                transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                                boxShadow: 'var(--shadow-sm)'
                              }} />
                            </button>
                          ) : editingKey === cfg.key ? (
                            cfg.key.endsWith('_time') ? (
                              <div style={{ position: 'relative' }}>
                                <div
                                  ref={timeTriggerRef}
                                  style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: '0.5rem',
                                    padding: '0.5rem 0.75rem',
                                    borderRadius: '6px',
                                    border: '1px solid var(--accent-color)',
                                    fontSize: '0.85rem',
                                    fontFamily: 'monospace',
                                    background: 'var(--bg-input)',
                                    color: 'var(--text-primary)',
                                    boxShadow: '0 0 0 3px rgba(16, 185, 129, 0.15)',
                                    width: 'fit-content'
                                  }}
                                >
                                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="var(--accent-color)" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" style={{ flexShrink: 0 }}><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
                                  <span style={{ fontWeight: 600 }}>{formatTime12h(editValue) || "Select Time"}</span>
                                </div>
                                <CustomTimePicker
                                  value={editValue}
                                  onChange={setEditValue}
                                  triggerRef={timeTriggerRef}
                                />
                              </div>
                            ) : cfg.key.endsWith('_date') ? (
                              <input
                                type="date"
                                value={editValue}
                                onChange={e => setEditValue(e.target.value)}
                                autoFocus
                                style={{
                                  padding: '0.5rem 0.75rem',
                                  borderRadius: '6px',
                                  border: '1px solid var(--accent-color)',
                                  fontSize: '0.85rem',
                                  fontFamily: 'monospace',
                                  outline: 'none',
                                  boxShadow: '0 0 0 3px rgba(14, 165, 233, 0.1)',
                                  color: 'var(--text-primary)',
                                  backgroundColor: 'var(--bg-input)'
                                }}
                                onKeyDown={e => {
                                  if (e.key === 'Enter') handleSaveConfig(cfg.key, editValue);
                                  if (e.key === 'Escape') handleCancelEdit();
                                }}
                              />
                            ) : (
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
                            )
                          ) : (
                            <div style={{
                              fontSize: '0.85rem',
                              fontFamily: 'monospace',
                              color: cfg.is_secret && !isRevealed ? 'var(--text-tertiary)' : 'var(--text-secondary)',
                              overflow: 'hidden',
                              textOverflow: 'ellipsis',
                              whiteSpace: 'nowrap',
                              letterSpacing: cfg.is_secret && !isRevealed ? '0.1em' : 'normal'
                            }}>
                              {cfg.key.endsWith('_time') && !cfg.is_secret ? (
                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-secondary)' }}>
                                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="var(--text-tertiary)" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" style={{ flexShrink: 0 }}><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
                                  <span>{formatTime12h(cfg.value) || <span style={{ color: 'var(--text-tertiary)', fontStyle: 'italic' }}>Not set</span>}</span>
                                </div>
                              ) : cfg.key.endsWith('_date') && !cfg.is_secret ? (
                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-secondary)' }}>
                                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="var(--text-tertiary)" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" style={{ flexShrink: 0 }}><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
                                  <span>{formatDateNice(cfg.value) || <span style={{ color: 'var(--text-tertiary)', fontStyle: 'italic' }}>Not set</span>}</span>
                                </div>
                              ) : (
                                cfg.value || <span style={{ color: 'var(--text-tertiary)', fontStyle: 'italic' }}>Not set</span>
                              )}
                            </div>
                          )}
                        </div>

                        <div style={{ display: 'flex', gap: '0.35rem', flexShrink: 0 }}>
                          {cfg.is_secret && (
                            <span style={{
                              fontSize: '0.65rem',
                              fontWeight: 700,
                              color: 'var(--status-error)',
                              background: 'var(--status-error-bg)',
                              padding: '2px 6px',
                              borderRadius: '4px',
                              border: '1px solid var(--status-error)',
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
                                style={{ color: 'var(--status-active)' }}
                              >
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>
                              </button>
                              <button
                                className="toolbar-btn"
                                title="Cancel"
                                onClick={handleCancelEdit}
                                style={{ color: 'var(--status-error)' }}
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
                          background: 'var(--status-active-bg)',
                          borderRadius: '10px',
                          border: '1px dashed var(--status-active)',
                          marginTop: '0.5rem'
                        }}
                      >
                        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                          <div style={{ 
                            width: '40px', 
                            height: '40px', 
                            borderRadius: '10px', 
                            background: 'var(--status-active)', 
                            display: 'flex', 
                            alignItems: 'center', 
                            justifyContent: 'center',
                            color: 'var(--text-primary)'
                          }}>
                            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="23 4 23 10 17 10"></polyline><polyline points="1 20 1 14 7 14"></polyline><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path></svg>
                          </div>
                          <div>
                            <div style={{ fontSize: '0.875rem', fontWeight: 700, color: 'var(--status-active)' }}>Manual Order Synchronization</div>
                            <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>Fetch missing orders or update existing ones directly from Shopify.</div>
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
                    {category === 'system' && configs.some(c => c.key === 'enable_danger_zone' && c.value === 'true') && (
                      <div
                        className="config-row"
                        style={{
                          display: 'flex',
                          flexDirection: 'column',
                          gap: '1.5rem',
                          padding: '1.5rem',
                          background: 'rgba(239, 68, 68, 0.05)',
                          borderRadius: '16px',
                          border: '1px solid rgba(239, 68, 68, 0.2)',
                          marginTop: '2rem'
                        }}
                      >
                        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                          <div style={{ 
                            width: '40px', 
                            height: '40px', 
                            borderRadius: '10px', 
                            background: '#ef4444', 
                            display: 'flex', 
                            alignItems: 'center', 
                            justifyContent: 'center',
                            color: 'white',
                            boxShadow: '0 4px 12px rgba(239, 68, 68, 0.3)'
                          }}>
                            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>
                          </div>
                          <div>
                            <div style={{ fontSize: '1rem', fontWeight: 700, color: '#b91c1c' }}>Danger Zone</div>
                            <div style={{ fontSize: '0.8rem', color: '#dc2626', opacity: 0.8 }}>Consolidated destructive actions. Use with extreme caution.</div>
                          </div>
                        </div>

                        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '1.5rem' }}>
                          <div style={{ padding: '1.25rem', background: 'white', borderRadius: '12px', border: '1px solid #fee2e2' }}>
                            <h5 style={{ margin: '0 0 0.5rem 0', fontSize: '0.9rem', fontWeight: 600, color: '#991b1b' }}>Inventory Wipe</h5>
                            <p style={{ margin: '0 0 1.25rem 0', fontSize: '0.75rem', color: '#b91c1c', lineHeight: 1.5 }}>Wipe all physical products and SKUs. This will break all current mapping authority.</p>
                            <button 
                              className="btn-secondary" 
                              onClick={handleResetInventory}
                              style={{ width: '100%', padding: '0.6rem', fontSize: '0.85rem', borderColor: '#ef4444', color: '#ef4444', background: 'white' }}
                            >
                              Wipe Warehouse
                            </button>
                          </div>

                          <div style={{ padding: '1.25rem', background: 'white', borderRadius: '12px', border: '1px solid #fee2e2' }}>
                            <h5 style={{ margin: '0 0 0.5rem 0', fontSize: '0.9rem', fontWeight: 600, color: '#991b1b' }}>Orders Wipe</h5>
                            <p style={{ margin: '0 0 1.25rem 0', fontSize: '0.75rem', color: '#b91c1c', lineHeight: 1.5 }}>Wipe all synced orders and customers. Data must be re-synced manually.</p>
                            <button 
                              className="btn-secondary" 
                              onClick={handleResetOrders}
                              style={{ width: '100%', padding: '0.6rem', fontSize: '0.85rem', borderColor: '#ef4444', color: '#ef4444', background: 'white' }}
                            >
                              Wipe Shopify Orders
                            </button>
                          </div>
                        </div>
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
            <div className="modal-header-icon" style={{ background: 'linear-gradient(135deg, var(--status-error), var(--status-warning))' }}>
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
             style={{ width: '100%', padding: '0.75rem 1rem', borderRadius: '10px', border: passwordError ? '2px solid var(--status-error)' : '2px solid var(--border-color)', fontSize: '0.95rem', outline: 'none', transition: 'border-color 0.2s', background: 'var(--bg-input)', color: 'var(--text-primary)' }}
              onKeyDown={e => { if (e.key === 'Enter') handleReveal(); }}
            />
            {passwordError && (
              <div style={{ color: 'var(--status-error)', fontSize: '0.8rem', marginTop: '0.5rem', fontWeight: 600 }}>
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
    </div>
  );
}
