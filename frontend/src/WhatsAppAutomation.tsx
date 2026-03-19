import { useState, useEffect } from 'react';
import { AutomationDashboard } from './AutomationDashboard';
import { AutomationTemplates } from './AutomationTemplates';
import { AutomationTriggers } from './AutomationTriggers';
import { AutomationMessages } from './AutomationMessages';

interface WhatsAppAutomationProps {
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
  startDate: string;
  endDate: string;
  onDateChange: (start: string, end: string) => void;
  refreshTrigger?: number;
}

export function WhatsAppAutomation({ fetchWithAuth, startDate, endDate, onDateChange, refreshTrigger }: WhatsAppAutomationProps) {
  const [activeSubTab, setActiveSubTab] = useState<string>(() => {
    return localStorage.getItem('gstAppActiveSubTab') || 'dashboard';
  });

  useEffect(() => {
    localStorage.setItem('gstAppActiveSubTab', activeSubTab);
  }, [activeSubTab]);

  const tabs = [
    { id: 'dashboard', label: 'Dashboard' },
    { id: 'templates', label: 'Templates' },
    { id: 'triggers', label: 'Triggers' },
    { id: 'messages', label: 'Message Logs' },
  ];

  return (
    <div className="whatsapp-automation-container">
      <div className="sub-tabs-header" style={{ 
        borderBottom: '1px solid var(--border-color)', 
        marginBottom: '1rem',
        paddingBottom: '0.5rem'
      }}>
        <div className="sub-tabs" style={{ display: 'flex', gap: '2rem' }}>
          {tabs.map(tab => (
            <button 
              key={tab.id}
              className={`sub-tab ${activeSubTab === tab.id ? 'active' : ''}`} 
              onClick={() => setActiveSubTab(tab.id)}
              style={{
                background: 'none',
                border: 'none',
                padding: '0.5rem 0',
                cursor: 'pointer',
                fontSize: '0.9rem',
                fontWeight: 600,
                color: activeSubTab === tab.id ? 'var(--accent-color)' : 'var(--text-secondary)',
                borderBottom: activeSubTab === tab.id ? '2px solid var(--accent-color)' : 'none',
                transition: 'all 0.2s ease'
              }}
            >
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      <div className="automation-content">
        {activeSubTab === 'dashboard' && <AutomationDashboard fetchWithAuth={fetchWithAuth} startDate={startDate} endDate={endDate} onDateChange={onDateChange} refreshTrigger={refreshTrigger} />}
        {activeSubTab === 'templates' && <AutomationTemplates fetchWithAuth={fetchWithAuth} startDate={startDate} endDate={endDate} onDateChange={onDateChange} />}
        {activeSubTab === 'triggers' && <AutomationTriggers fetchWithAuth={fetchWithAuth} />}
        {activeSubTab === 'messages' && <AutomationMessages fetchWithAuth={fetchWithAuth} startDate={startDate} endDate={endDate} onDateChange={onDateChange} refreshTrigger={refreshTrigger} />}
      </div>
    </div>
  );
}
