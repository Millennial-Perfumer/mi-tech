import { API_BASE } from './api';
import { useState, useEffect } from 'react';

interface Trigger {
  id: number;
  webhook_topic: string;
  template_id: number;
  enabled: boolean;
  created_at: string;
}

interface Template {
  id: number;
  template_name: string;
  status: string;
}

interface AutomationTriggersProps {
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
}

export function AutomationTriggers({ fetchWithAuth }: AutomationTriggersProps) {
  const [triggers, setTriggers] = useState<Trigger[]>([]);
  const [templates, setTemplates] = useState<Template[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [showForm, setShowForm] = useState(false);

  const [formData, setFormData] = useState({
    topic: 'orders/create',
    templateID: ''
  });

  const fetchData = async () => {
    setIsLoading(true);
    try {
      const [triggersResp, templatesResp] = await Promise.all([
        fetchWithAuth(`${API_BASE}/api/automation/whatsapp/triggers`),
        fetchWithAuth(`${API_BASE}/api/automation/whatsapp/templates`)
      ]);
      const triggersData = await triggersResp.json();
      const templatesData = await templatesResp.json();
      setTriggers(triggersData || []);
      // Filter only approved/pending (user said only approved, but pending is also okay for mapping usually)
      setTemplates((templatesData || []).filter((t: Template) => t.status === 'approved' || t.status === 'APPROVED'));
    } catch (err) {
      console.error('Failed to fetch data:', err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => { fetchData(); }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.templateID) return;
    
    setIsSaving(true);
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/automation/whatsapp/triggers`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          webhook_topic: formData.topic,
          template_id: parseInt(formData.templateID)
        })
      });
      if (resp.ok) {
        setShowForm(false);
        setFormData({ topic: 'orders/create', templateID: '' });
        fetchData();
      }
    } catch (err) {
      console.error('Failed to save trigger:', err);
    } finally {
      setIsSaving(false);
    }
  };

  const handleToggle = async (id: number, currentEnabled: boolean) => {
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/automation/whatsapp/triggers`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id, enabled: !currentEnabled })
      });
      if (resp.ok) fetchData();
    } catch (err) {
      console.error('Failed to toggle trigger:', err);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this trigger mapping?')) return;
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/automation/whatsapp/triggers?id=${id}`, { method: 'DELETE' });
      if (resp.ok) fetchData();
    } catch (err) {
      console.error('Failed to delete trigger:', err);
    }
  };

  const webhookOptions = [
    { value: 'orders/create', label: 'Order Created' },
    { value: 'orders/paid', label: 'Order Paid' },
    { value: 'orders/fulfilled', label: 'Order Fulfilled' },
    { value: 'orders/cancelled', label: 'Order Cancelled' },
    { value: 'orders/updated', label: 'Order Updated' },
  ];

  return (
    <div className="automation-page">
      <div style={{ 
        display: 'flex', 
        justifyContent: 'space-between', 
        alignItems: 'center', 
        marginBottom: '2rem',
        padding: '1.25rem 1.5rem',
        background: 'white',
        borderRadius: '16px',
        boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.05), 0 2px 4px -1px rgba(0, 0, 0, 0.03)',
        border: '1px solid #f1f5f9'
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '2rem' }}>
          <div>
            <h1 style={{ margin: 0, fontSize: '1.5rem', fontWeight: 800, color: '#0f172a', letterSpacing: '-0.025em' }}>Webhook Triggers</h1>
            <p style={{ margin: '4px 0 0 0', color: '#64748b', fontSize: '0.9rem', fontWeight: 500 }}>
              Map Shopify events to automated message templates
            </p>
          </div>
          
          <div style={{ width: '1px', height: '40px', backgroundColor: '#e2e8f0' }}></div>
        </div>

        <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
          <button 
            className="btn-primary" 
            onClick={() => setShowForm(!showForm)}
            style={{
              backgroundColor: showForm ? '#475569' : '#0ea5e9',
              color: 'white',
              border: 'none',
              padding: '0.65rem 1.25rem',
              borderRadius: '10px',
              fontSize: '0.875rem',
              fontWeight: 600,
              boxShadow: showForm ? 'none' : '0 4px 6px -1px rgba(14, 165, 233, 0.2)',
              cursor: 'pointer',
              transition: 'all 0.2s',
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem'
            }}
          >
            {showForm ? (
               <>
                 <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                 Cancel Registration
               </>
            ) : (
               <>
                 <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
                 Register New Trigger
               </>
            )}
          </button>
        </div>
      </div>

      {showForm && (
        <div className="card" style={{ marginBottom: '2rem', maxWidth: '500px', padding: '2rem' }}>
          <h3 style={{ marginBottom: '1.5rem' }}>Register New Trigger</h3>
          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <label>Select Shopify Event</label>
              <select value={formData.topic} onChange={e => setFormData({...formData, topic: e.target.value})}>
                {webhookOptions.map(opt => <option key={opt.value} value={opt.value}>{opt.label}</option>)}
              </select>
            </div>
            <div className="form-group">
              <label>Map to Template</label>
              <select 
                value={formData.templateID} 
                onChange={e => setFormData({...formData, templateID: e.target.value})}
                required
              >
                <option value="">Select an approved template...</option>
                {templates.map(t => (
                  <option key={t.id} value={t.id}>{t.template_name}</option>
                ))}
              </select>
            </div>
            <div style={{ marginTop: '1.5rem', display: 'flex', gap: '1rem' }}>
              <button type="submit" className="btn-primary" disabled={isSaving} style={{ flex: 1 }}>
                {isSaving ? 'Saving...' : 'Create Mapping'}
              </button>
              <button type="button" className="btn-secondary" onClick={() => setShowForm(false)} style={{ flex: 1 }}>
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="table-container" style={{ width: '100%', overflowX: 'auto' }}>
        <table style={{ width: '100%', borderCollapse: 'separate', borderSpacing: 0 }}>
          <thead>
            <tr style={{ backgroundColor: '#f8fafc' }}>
              <th style={{ padding: '1rem 1.5rem', fontSize: '0.75rem', fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', borderBottom: '1px solid #f1f5f9' }}>Webhook Event</th>
              <th style={{ padding: '1rem 1.5rem', fontSize: '0.75rem', fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', borderBottom: '1px solid #f1f5f9' }}>Template</th>
              <th style={{ padding: '1rem 1.5rem', fontSize: '0.75rem', fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', borderBottom: '1px solid #f1f5f9' }}>Status</th>
              <th style={{ padding: '1rem 1.5rem', fontSize: '0.75rem', fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', borderBottom: '1px solid #f1f5f9' }}>Created Time</th>
              <th style={{ padding: '1rem 1.5rem', fontSize: '0.75rem', fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', borderBottom: '1px solid #f1f5f9', textAlign: 'right' }}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr><td colSpan={5} style={{ textAlign: 'center', padding: '3rem', color: '#64748b' }}>Loading triggers...</td></tr>
            ) : triggers.length === 0 ? (
              <tr><td colSpan={5} style={{ textAlign: 'center', padding: '3rem', color: '#64748b' }}>No triggers found.</td></tr>
            ) : (
              triggers.map(tr => (
                <tr key={tr.id} style={{ transition: 'background-color 0.2s' }} className="hover-row">
                  <td style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid #f1f5f9' }}>
                    <code style={{ background: '#f1f5f9', padding: '0.2rem 0.5rem', borderRadius: '4px', fontSize: '0.85rem', color: '#334155', fontWeight: 600 }}>{tr.webhook_topic}</code>
                  </td>
                  <td style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid #f1f5f9', color: '#1e293b', fontWeight: 500 }}>
                    {templates.find(t => t.id === tr.template_id)?.template_name || `ID: ${tr.template_id}`}
                  </td>
                  <td style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid #f1f5f9' }}>
                    <span className={`badge-pill ${tr.enabled ? 'badge-pill-success' : 'badge-pill-warning'}`} style={{ fontSize: '0.75rem' }}>
                      <span className="dot" />
                      {tr.enabled ? 'ENABLED' : 'DISABLED'}
                    </span>
                  </td>
                  <td style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid #f1f5f9', color: '#64748b', fontSize: '0.85rem' }}>
                    {new Date(tr.created_at).toLocaleDateString('en-GB')}
                  </td>
                  <td style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid #f1f5f9', textAlign: 'right' }}>
                    <div style={{ display: 'flex', gap: '0.75rem', justifyContent: 'flex-end', alignItems: 'center' }}>
                      <button 
                        onClick={() => handleToggle(tr.id, tr.enabled)}
                        title={tr.enabled ? "Disable Trigger" : "Enable Trigger"}
                        style={{ background: 'none', border: 'none', cursor: 'pointer', padding: '4px', display: 'flex', alignItems: 'center', color: tr.enabled ? '#10b981' : '#94a3b8' }}
                      >
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                          <path d="M18.36 6.64a9 9 0 1 1-12.73 0"></path>
                          <line x1="12" y1="2" x2="12" y2="12"></line>
                        </svg>
                      </button>
                      <button 
                        onClick={() => handleDelete(tr.id)} 
                        title="Delete Trigger"
                        style={{ background: 'none', border: 'none', cursor: 'pointer', padding: '4px', display: 'flex', alignItems: 'center', color: '#ef4444' }}
                      >
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                          <polyline points="3 6 5 6 21 6"></polyline>
                          <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                          <line x1="10" y1="11" x2="10" y2="17"></line>
                          <line x1="14" y1="11" x2="14" y2="17"></line>
                        </svg>
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
