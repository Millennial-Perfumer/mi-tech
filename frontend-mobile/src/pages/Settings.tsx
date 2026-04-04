import React, { useState } from 'react';
import { useAuthStore } from '../store/useAuthStore';
import { useMetricStore } from '../store/useMetricStore';
import { apiFetch } from '../api';
import { 
  RefreshCw, 
  Trash2, 
  LogOut, 
  Shield
} from 'lucide-react';

export const Settings: React.FC = () => {
  const logout = useAuthStore(state => state.logout);
  const { fetchMetrics } = useMetricStore();
  const [isSyncing, setIsSyncing] = useState(false);
  const [syncStatus, setSyncStatus] = useState<string | null>(null);

  const handleSync = async () => {
    setIsSyncing(true);
    setSyncStatus('Iniciating Shopify Sync...');
    
    try {
      const response = await apiFetch(`/api/shopify/sync`, {
        method: 'POST'
      });
      const data = await response.json();
      if (data.success) {
        setSyncStatus('Sync compete. Updating metrics...');
        const end = new Date();
        const start = new Date();
        start.setDate(end.getDate() - 30);
        await fetchMetrics(start.toISOString().split('T')[0], end.toISOString().split('T')[0]);
        setSyncStatus('Database is now up-to-date.');
      } else {
        setSyncStatus(`Sync failed: ${data.message}`);
      }
    } catch (err: any) {
      if (err.message !== 'Unauthorized') {
        setSyncStatus('Network error during sync.');
      }
    } finally {
      setIsSyncing(false);
      setTimeout(() => setSyncStatus(null), 3000);
    }
  };

  const handleReset = async () => {
    if (!window.confirm('CRITICAL: This will wipe all local order and customer data. Are you absolutely sure?')) return;
    
    try {
      const response = await apiFetch(`/api/shopify/reset-data`, {
        method: 'POST'
      });
      const data = await response.json();
      if (data.success) {
        alert('Data reset successfully. Please re-sync.');
        window.location.reload();
      }
    } catch (err: any) {
      if (err.message !== 'Unauthorized') {
        alert('Failed to reset data.');
      }
    }
  };

  return (
    <div>
      <header style={{ marginBottom: '2.5rem' }}>
        <p style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--accent-color)', textTransform: 'uppercase', letterSpacing: '0.15em' }}>Management</p>
        <h1>Settings</h1>
      </header>

      <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
        <div className="glass-card" style={{ padding: '1.5rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1.25rem' }}>
            <div style={{ background: 'var(--accent-subtle)', color: 'var(--accent-color)', padding: '10px', borderRadius: '12px' }}>
              <RefreshCw size={20} />
            </div>
            <h3>Cloud Sync</h3>
          </div>
          <p style={{ fontSize: '0.85rem', marginBottom: '1.5rem' }}>Pull latest orders and customers from Shopify into the local Mi Tech database.</p>
          
          <button 
            disabled={isSyncing}
            onClick={handleSync}
            style={{ 
              width: '100%', 
              background: 'var(--accent-color)', 
              color: '#fff', 
              border: 'none', 
              padding: '1rem', 
              fontSize: '1rem',
              fontWeight: 700,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: '0.5rem',
              borderRadius: '12px'
            }}
          >
            <RefreshCw size={20} className={isSyncing ? 'animate-spin' : ''} />
            {isSyncing ? 'Syncing...' : 'Start Shopify Sync'}
          </button>
          
          {syncStatus && (
            <p style={{ fontSize: '0.8rem', textAlign: 'center', marginTop: '1rem', color: 'var(--accent-color)' }}>
              {syncStatus}
            </p>
          )}
        </div>

        <div className="glass-card" style={{ padding: '1.5rem', border: '1px solid rgba(239, 68, 68, 0.2)' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1.25rem' }}>
            <div style={{ background: 'rgba(239, 68, 68, 0.1)', color: '#ef4444', padding: '10px', borderRadius: '12px' }}>
              <Trash2 size={20} />
            </div>
            <h3>Danger Zone</h3>
          </div>
          <p style={{ fontSize: '0.85rem', marginBottom: '1.5rem' }}>Wipe all local Shopify data cache. This does not affect your Shopify store directly.</p>
          
          <button 
            onClick={handleReset}
            className="glass-panel"
            style={{ 
              width: '100%', 
              color: '#ef4444',
              borderColor: 'rgba(239, 68, 68, 0.3)',
              padding: '1rem', 
              fontSize: '0.95rem',
              fontWeight: 600,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: '0.5rem',
              borderRadius: '12px',
              background: 'transparent'
            }}
          >
            Reset Local Database
          </button>
        </div>

        <div className="glass-card" style={{ padding: '1.5rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1.5rem' }}>
            <div style={{ background: 'var(--bg-input)', color: 'var(--text-secondary)', padding: '10px', borderRadius: '12px' }}>
              <Shield size={20} />
            </div>
            <h3>Account</h3>
          </div>
          
          <button 
            onClick={logout}
            className="glass-panel"
            style={{ 
              width: '100%', 
              padding: '1rem', 
              fontSize: '0.95rem',
              fontWeight: 600,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: '0.5rem',
              borderRadius: '12px',
              background: 'transparent'
            }}
          >
            <LogOut size={20} />
            Sign Out
          </button>
        </div>

        <div style={{ textAlign: 'center', marginTop: '1rem', opacity: 0.5 }}>
          <p style={{ fontSize: '0.7rem' }}>Mi Tech Mobile v1.0.0 (BETA)</p>
        </div>
      </div>
    </div>
  );
};
