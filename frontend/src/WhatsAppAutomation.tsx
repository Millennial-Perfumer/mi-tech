import { useState, useEffect } from 'react';
import { AutomationDashboard } from './AutomationDashboard';
import { AutomationTemplates } from './AutomationTemplates';
import { AutomationTriggers } from './AutomationTriggers';
import { AutomationMessages } from './AutomationMessages';
import { CustomDatePicker } from './CustomDatePicker';

interface WhatsAppAutomationProps {
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
}

export function WhatsAppAutomation({ fetchWithAuth }: WhatsAppAutomationProps) {
  const [activeSubTab, setActiveSubTab] = useState<string>(() => {
    return localStorage.getItem('gstAppActiveSubTab') || 'dashboard';
  });

  const [startDate, setStartDate] = useState<string>(() => {
    const d = new Date();
    d.setMonth(d.getMonth() - 1);
    return d.toISOString().split('T')[0];
  });
  const [endDate, setEndDate] = useState<string>(new Date().toISOString().split('T')[0]);

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

      {activeSubTab !== 'triggers' && (
        <div style={{ display: 'flex', justifyContent: 'flex-start', marginBottom: '2rem' }}>
          <CustomDatePicker 
            startDate={startDate}
            endDate={endDate}
            onDateChange={(start, end) => {
              setStartDate(start);
              setEndDate(end);
            }}
          />
        </div>
      )}

      <div className="automation-content">
        {activeSubTab === 'dashboard' && <AutomationDashboard fetchWithAuth={fetchWithAuth} startDate={startDate} endDate={endDate} />}
        {activeSubTab === 'templates' && <AutomationTemplates fetchWithAuth={fetchWithAuth} startDate={startDate} endDate={endDate} />}
        {activeSubTab === 'triggers' && <AutomationTriggers fetchWithAuth={fetchWithAuth} />}
        {activeSubTab === 'messages' && <AutomationMessages fetchWithAuth={fetchWithAuth} startDate={startDate} endDate={endDate} />}
      </div>
    </div>
  );
}
