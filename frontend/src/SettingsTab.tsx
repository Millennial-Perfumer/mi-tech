import { useState } from 'react';

interface SettingsTabProps {
  settings: Record<string, string>;
  onUpdateSetting: (key: string, value: string) => Promise<void>;
  isSyncing: boolean;
  isResetting: boolean;
}

export function SettingsTab({ settings, onUpdateSetting, isSyncing, isResetting }: SettingsTabProps) {
  const [localSettings, setLocalSettings] = useState(settings);
  const [isSaving, setIsSaving] = useState<string | null>(null);

  const handleToggle = async (key: string) => {
    const newValue = localSettings[key] === 'true' ? 'false' : 'true';
    setIsSaving(key);
    try {
      await onUpdateSetting(key, newValue);
      setLocalSettings(prev => ({ ...prev, [key]: newValue }));
    } finally {
      setIsSaving(null);
    }
  };

  const handleInputChange = (key: string, value: string) => {
    setLocalSettings(prev => ({ ...prev, [key]: value }));
  };

  const handleSaveProfile = async () => {
    const keysToSave = [
      'business_name', 
      'business_gstin', 
      'business_phone', 
      'business_address_line1', 
      'business_address_line2'
    ];
    
    setIsSaving('business_profile');
    try {
      for (const key of keysToSave) {
        if (localSettings[key] !== settings[key]) {
          await onUpdateSetting(key, localSettings[key] || '');
        }
      }
    } finally {
      setIsSaving(null);
    }
  };

  const hasProfileChanges = [
    'business_name', 
    'business_gstin', 
    'business_phone', 
    'business_address_line1', 
    'business_address_line2'
  ].some(key => (localSettings[key] || '') !== (settings[key] || ''));

  return (
    <div className="settings-container" style={{ padding: '0.5rem' }}>
      <div className="settings-grid" style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(400px, 1fr))', gap: '2rem' }}>
        
        {/* Feature Toggles */}
        <section className="card" style={{ padding: '2rem' }}>
          <h2 style={{ fontSize: '1.25rem', fontWeight: 700, marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="var(--accent-color)" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"></polyline></svg>
            Feature Management
          </h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', marginBottom: '2rem' }}>
            Enable or disable specific tools and buttons across the application interface.
          </p>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '1rem', background: '#f8fafc', borderRadius: '12px', border: '1px solid #e2e8f0' }}>
              <div>
                <div style={{ fontWeight: 600, color: '#0f172a' }}>Reset & Resync Button</div>
                <div style={{ fontSize: '0.8rem', color: '#64748b' }}>Show the Red "Reset" button in the header</div>
              </div>
              <button 
                onClick={() => handleToggle('show_reset_button')}
                disabled={isSaving === 'show_reset_button' || isResetting || isSyncing}
                style={{
                  position: 'relative',
                  width: '48px',
                  height: '24px',
                  borderRadius: '12px',
                  backgroundColor: localSettings.show_reset_button === 'true' ? 'var(--accent-color)' : '#cbd5e1',
                  border: 'none',
                  cursor: 'pointer',
                  transition: 'all 0.3s ease',
                  padding: '2px'
                }}
              >
                <div style={{
                  width: '20px',
                  height: '20px',
                  borderRadius: '50%',
                  backgroundColor: 'white',
                  transform: localSettings.show_reset_button === 'true' ? 'translateX(24px)' : 'translateX(0)',
                  transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                  boxShadow: '0 1px 3px rgba(0,0,0,0.1)'
                }} />
              </button>
            </div>

            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '1rem', background: '#f8fafc', borderRadius: '12px', border: '1px solid #e2e8f0' }}>
              <div>
                <div style={{ fontWeight: 600, color: '#0f172a' }}>Manual Sync Button</div>
                <div style={{ fontSize: '0.8rem', color: '#64748b' }}>Show the \"Manual Sync\" button in the header</div>
              </div>
              <button 
                onClick={() => handleToggle('show_sync_button')}
                disabled={isSaving === 'show_sync_button' || isSyncing || isResetting}
                style={{
                  position: 'relative',
                  width: '48px',
                  height: '24px',
                  borderRadius: '12px',
                  backgroundColor: localSettings.show_sync_button !== 'false' ? 'var(--accent-color)' : '#cbd5e1', // Default to true if not set
                  border: 'none',
                  cursor: 'pointer',
                  transition: 'all 0.3s ease',
                  padding: '2px'
                }}
              >
                <div style={{
                  width: '20px',
                  height: '20px',
                  borderRadius: '50%',
                  backgroundColor: 'white',
                  transform: localSettings.show_sync_button !== 'false' ? 'translateX(24px)' : 'translateX(0)',
                  transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                  boxShadow: '0 1px 3px rgba(0,0,0,0.1)'
                }} />
              </button>
            </div>

            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '1rem', background: '#f8fafc', borderRadius: '12px', border: '1px solid #e2e8f0' }}>
              <div>
                <div style={{ fontWeight: 600, color: '#0f172a' }}>Send Invoice with Order</div>
                <div style={{ fontSize: '0.8rem', color: '#64748b' }}>Attach PDF invoice to WhatsApp order messages</div>
              </div>
              <button 
                onClick={() => handleToggle('send_invoice')}
                disabled={isSaving === 'send_invoice'}
                style={{
                  position: 'relative',
                  width: '48px',
                  height: '24px',
                  borderRadius: '12px',
                  backgroundColor: localSettings.send_invoice !== 'false' ? 'var(--accent-color)' : '#cbd5e1', // Default to true if not set
                  border: 'none',
                  cursor: 'pointer',
                  transition: 'all 0.3s ease',
                  padding: '2px'
                }}
              >
                <div style={{
                  width: '20px',
                  height: '20px',
                  borderRadius: '50%',
                  backgroundColor: 'white',
                  transform: localSettings.send_invoice !== 'false' ? 'translateX(24px)' : 'translateX(0)',
                  transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                  boxShadow: '0 1px 3px rgba(0,0,0,0.1)'
                }} />
              </button>
            </div>
          </div>
        </section>

        {/* Business Information */}
        <section className="card" style={{ padding: '2rem' }}>
          <h2 style={{ fontSize: '1.25rem', fontWeight: 700, marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="var(--accent-color)" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path><polyline points="9 22 9 12 15 12 15 22"></polyline></svg>
            Business Profile
          </h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', marginBottom: '2rem' }}>
            These details are used for generating GST-compliant invoices and reports.
          </p>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
            <div className="input-group">
              <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: '#475569', marginBottom: '0.5rem' }}>Business Name</label>
              <input 
                type="text" 
                value={localSettings.business_name || ''} 
                onChange={(e) => handleInputChange('business_name', e.target.value)}
                style={{ width: '100%', padding: '0.75rem', borderRadius: '8px', border: '1px solid #e2e8f0', fontSize: '0.9rem' }}
                placeholder="Enter business name"
              />
            </div>
            
            <div className="input-group">
              <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: '#475569', marginBottom: '0.5rem' }}>GSTIN</label>
              <input 
                type="text" 
                value={localSettings.business_gstin || ''} 
                onChange={(e) => handleInputChange('business_gstin', e.target.value)}
                style={{ width: '100%', padding: '0.75rem', borderRadius: '8px', border: '1px solid #e2e8f0', fontSize: '0.9rem' }}
                placeholder="Enter GST registration number"
              />
            </div>

            <div className="input-group">
              <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: '#475569', marginBottom: '0.5rem' }}>Phone Number</label>
              <input 
                type="text" 
                value={localSettings.business_phone || ''} 
                onChange={(e) => handleInputChange('business_phone', e.target.value)}
                style={{ width: '100%', padding: '0.75rem', borderRadius: '8px', border: '1px solid #e2e8f0', fontSize: '0.9rem' }}
                placeholder="Enter business contact number"
              />
            </div>

            <div className="input-group">
              <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: '#475569', marginBottom: '0.5rem' }}>Address Line 1</label>
              <input 
                type="text" 
                value={localSettings.business_address_line1 || ''} 
                onChange={(e) => handleInputChange('business_address_line1', e.target.value)}
                style={{ width: '100%', padding: '0.75rem', borderRadius: '8px', border: '1px solid #e2e8f0', fontSize: '0.9rem' }}
                placeholder="Street address, P.O. box, company name, c/o"
              />
            </div>

            <div className="input-group">
              <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: '#475569', marginBottom: '0.5rem' }}>Address Line 2</label>
              <textarea 
                value={localSettings.business_address_line2 || ''} 
                onChange={(e) => handleInputChange('business_address_line2', e.target.value)}
                style={{ width: '100%', padding: '0.75rem', borderRadius: '8px', border: '1px solid #e2e8f0', fontSize: '0.9rem', minHeight: '80px', resize: 'vertical' }}
                placeholder="Apartment, suite, unit, building, floor, etc."
              />
            </div>

            <div style={{ marginTop: '1rem', display: 'flex', justifyContent: 'flex-end' }}>
              <button 
                className="btn-primary"
                onClick={handleSaveProfile}
                disabled={isSaving === 'business_profile' || !hasProfileChanges}
                style={{ 
                  padding: '0.75rem 2rem', 
                  borderRadius: '8px', 
                  fontWeight: 600,
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.5rem',
                  opacity: (isSaving === 'business_profile' || !hasProfileChanges) ? 0.6 : 1,
                  cursor: (isSaving === 'business_profile' || !hasProfileChanges) ? 'not-allowed' : 'pointer'
                }}
              >
                {isSaving === 'business_profile' ? (
                  <>
                    <svg className="spin" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="12" y1="2" x2="12" y2="6"></line><line x1="12" y1="18" x2="12" y2="22"></line><line x1="4.93" y1="4.93" x2="7.76" y2="7.76"></line><line x1="16.24" y1="16.24" x2="19.07" y2="19.07"></line><line x1="2" y1="12" x2="6" y2="12"></line><line x1="18" y1="12" x2="22" y2="12"></line><line x1="4.93" y1="19.07" x2="7.76" y2="16.24"></line><line x1="16.24" y1="7.76" x2="19.07" y2="4.93"></line></svg>
                    Saving...
                  </>
                ) : 'Save Profile Changes'}
              </button>
            </div>
          </div>
        </section>
      </div>
    </div>
  );
}
