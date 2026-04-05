import React, { useEffect } from 'react';
import { useAutomationStore } from '../store/useAutomationStore';
import { Zap, Power, RefreshCcw, Copy, Edit3 } from 'lucide-react';

export const Automation: React.FC = () => {
  const { templates, isLoading, fetchTemplates, toggleTemplate } = useAutomationStore();

  useEffect(() => {
    fetchTemplates();
  }, []);

  return (
    <div>
      <header style={{ marginBottom: '2rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <p style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--accent-color)', textTransform: 'uppercase', letterSpacing: '0.15em' }}>Workflows</p>
          <h1>Automation</h1>
        </div>
        <button onClick={fetchTemplates} style={{ color: 'var(--text-tertiary)', background: 'transparent', border: 'none' }}>
          <RefreshCcw size={20} className={isLoading ? 'animate-spin' : ''} />
        </button>
      </header>

      <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
        {isLoading ? (
          <p style={{ textAlign: 'center', color: 'var(--text-tertiary)' }}>Loading templates...</p>
        ) : templates.length === 0 ? (
          <p style={{ textAlign: 'center', color: 'var(--text-tertiary)' }}>No automation templates found.</p>
        ) : templates.map((template) => (
          <div key={template.id} className="glass-card" style={{ padding: '1.25rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1rem' }}>
              <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
                <div style={{ 
                  color: template.is_active ? 'var(--accent-color)' : 'var(--text-tertiary)', 
                  background: 'var(--bg-input)', 
                  padding: '10px', 
                  borderRadius: '12px' 
                }}>
                  <Zap size={20} fill={template.is_active ? 'var(--accent-color)' : 'none'} />
                </div>
                <div>
                  <h3 style={{ fontSize: '1rem', color: '#fff' }}>{template.name}</h3>
                  <p style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>{template.trigger_event}</p>
                </div>
              </div>
              
              <button 
                onClick={() => toggleTemplate(template.id, !template.is_active)}
                style={{
                  background: template.is_active ? 'var(--accent-subtle)' : 'rgba(255,255,255,0.05)',
                  border: `1px solid ${template.is_active ? 'var(--accent-color)' : 'var(--glass-border)'}`,
                  color: template.is_active ? 'var(--accent-color)' : 'var(--text-tertiary)',
                  width: '48px',
                  height: '48px',
                  borderRadius: '12px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center'
                }}
              >
                <Power size={20} />
              </button>
            </div>

            <div style={{ 
              background: 'rgba(0,0,0,0.2)', 
              borderRadius: '10px', 
              padding: '1rem', 
              fontSize: '0.85rem', 
              color: 'var(--text-secondary)',
              lineHeight: 1.5,
              border: '1px solid rgba(255,255,255,0.03)'
            }}>
              {template.content}
            </div>

            <div style={{ display: 'flex', gap: '0.75rem', marginTop: '1.25rem' }}>
              <button className="glass-panel" style={{ flex: 1, gap: '0.5rem', fontSize: '0.85rem', display: 'flex', alignItems: 'center', justifyContent: 'center', borderRadius: '12px', background: 'transparent' }}>
                <Copy size={16} />
                Duplicate
              </button>
              <button className="glass-panel" style={{ flex: 1, gap: '0.5rem', fontSize: '0.85rem', display: 'flex', alignItems: 'center', justifyContent: 'center', borderRadius: '12px', background: 'transparent' }}>
                <Edit3 size={16} />
                Edit
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
