import { useState } from 'react';
import { AutomationDashboard } from './AutomationDashboard';
import { AutomationTemplates } from './AutomationTemplates';
import { AutomationTriggers } from './AutomationTriggers';
import { AutomationMessages } from './AutomationMessages';

export function WhatsAppAutomation() {
  const [activeSubTab, setActiveSubTab] = useState('dashboard');

  const tabs = [
    { id: 'dashboard', label: 'Dashboard' },
    { id: 'templates', label: 'Templates' },
    { id: 'triggers', label: 'Triggers' },
    { id: 'messages', label: 'Message Logs' },
  ];

  return (
    <div className="whatsapp-automation-container">
      <div className="sub-tabs" style={{ 
        display: 'flex', 
        gap: '2rem', 
        borderBottom: '1px solid var(--border-color)', 
        marginBottom: '2rem',
        paddingBottom: '0.5rem'
      }}>
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

      <div className="automation-content">
        {activeSubTab === 'dashboard' && <AutomationDashboard />}
        {activeSubTab === 'templates' && <AutomationTemplates />}
        {activeSubTab === 'triggers' && <AutomationTriggers />}
        {activeSubTab === 'messages' && <AutomationMessages />}
      </div>
    </div>
  );
}
