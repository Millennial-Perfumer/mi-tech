import { useState, useEffect, useCallback } from 'react';
import { API_BASE, fetchWithAuth } from '../api';
import { MobileCard } from '../components/MobileCard';
import './Settings.tsx.css';

interface User {
  id: number;
  username: string;
  role: string;
}

export const Settings: React.FC = () => {
  const [appSettings, setAppSettings] = useState<Record<string, string>>({});
  const [users, setUsers] = useState<User[]>([]);
  const [saveStatus, setSaveStatus] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    try {
      const settingsRes = await fetchWithAuth(`${API_BASE}/api/settings`);
      const settingsData = await settingsRes.json();
      if (settingsData.success) {
        setAppSettings(settingsData.settings);
      }

      const usersRes = await fetchWithAuth(`${API_BASE}/api/users`);
      const usersData = await usersRes.json();
      if (usersData.success) {
        setUsers(usersData.users);
      }
    } catch (err) {
      console.error('Error fetching settings/users:', err);
    }
  }, []);

  useEffect(() => {
    /* eslint-disable react-hooks/set-state-in-effect */
    void fetchData();
  }, [fetchData]);

  const toggleSetting = async (key: string) => {
    const newValue = appSettings[key] === 'true' ? 'false' : 'true';
    setAppSettings(prev => ({ ...prev, [key]: newValue }));

    try {
      const res = await fetchWithAuth(`${API_BASE}/api/settings`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ [key]: newValue })
      });
      const data = await res.json();
      if (data.success) {
        setSaveStatus('Changes saved!');
        setTimeout(() => setSaveStatus(null), 2000);
      }
    } catch (err) {
      console.error('Failed to save setting:', err);
      setSaveStatus('Save failed!');
      setTimeout(() => setSaveStatus(null), 2000);
    }
  };

  return (
    <div className="settings-container">
      <h2 className="section-title">Settings</h2>

      {saveStatus && <div className={`save-banner ${saveStatus.includes('failed') ? 'error' : 'success'}`}>{saveStatus}</div>}

      <div className="settings-group">
        <h3 className="group-title">System Preferences</h3>
        <MobileCard className="setting-row">
          <div className="setting-info">
            <span className="setting-label">Show Sync Button</span>
            <span className="setting-desc">Allow manual Shopify synchronization</span>
          </div>
          <label className="toggle-switch">
            <input
              type="checkbox"
              checked={appSettings.show_sync_button === 'true'}
              onChange={() => toggleSetting('show_sync_button')}
            />
            <span className="toggle-slider"></span>
          </label>
        </MobileCard>

        <MobileCard className="setting-row">
          <div className="setting-info">
            <span className="setting-label">Auto-Refresh Dashboard</span>
            <span className="setting-desc">Keep metrics updated in real-time</span>
          </div>
          <label className="toggle-switch">
            <input
              type="checkbox"
              checked={appSettings.auto_refresh === 'true'}
              onChange={() => toggleSetting('auto_refresh')}
            />
            <span className="toggle-slider"></span>
          </label>
        </MobileCard>
      </div>

      <div className="settings-group">
        <h3 className="group-title">User Management</h3>
        <div className="user-list">
          {users.map(user => (
            <MobileCard key={user.id} className="user-card-small">
              <div className="user-avatar-sm">{user.username.charAt(0).toUpperCase()}</div>
              <div className="user-details">
                <span className="u-name">{user.username}</span>
                <span className="u-role">{user.role}</span>
              </div>
            </MobileCard>
          ))}
        </div>
      </div>

      <button className="danger-btn full-width" style={{marginTop: '2rem'}} onClick={() => { localStorage.removeItem('mobileToken'); window.location.reload(); }}>
        🚪 Sign Out
      </button>
    </div>
  );
};
