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
        fetchWithAuth('http://localhost:8080/api/automation/whatsapp/triggers'),
        fetchWithAuth('http://localhost:8080/api/automation/whatsapp/templates')
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
      const resp = await fetchWithAuth('http://localhost:8080/api/automation/whatsapp/triggers', {
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
      const resp = await fetchWithAuth(`http://localhost:8080/api/automation/whatsapp/triggers`, {
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
      const resp = await fetchWithAuth(`http://localhost:8080/api/automation/whatsapp/triggers?id=${id}`, { method: 'DELETE' });
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
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <h2 style={{ fontSize: '1.25rem', fontWeight: 600 }}>Webhook Triggers</h2>
        <button className="btn-primary" onClick={() => setShowForm(!showForm)}>
          {showForm ? 'Cancel' : 'Register Trigger'}
        </button>
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

      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>Webhook Event</th>
              <th>Template</th>
              <th>Status</th>
              <th>Created Time</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr><td colSpan={5} style={{ textAlign: 'center', padding: '2rem' }}>Loading triggers...</td></tr>
            ) : triggers.length === 0 ? (
              <tr><td colSpan={5} style={{ textAlign: 'center', padding: '2rem' }}>No triggers found.</td></tr>
            ) : (
              triggers.map(tr => (
                <tr key={tr.id}>
                  <td><code>{tr.webhook_topic}</code></td>
                  <td>{templates.find(t => t.id === tr.template_id)?.template_name || `ID: ${tr.template_id}`}</td>
                  <td>
                    <span className={`badge ${tr.enabled ? 'badge-success' : 'badge-warning'}`}>
                      {tr.enabled ? 'ENABLED' : 'DISABLED'}
                    </span>
                  </td>
                  <td style={{ fontSize: '0.85rem' }}>{new Date(tr.created_at).toLocaleDateString()}</td>
                  <td>
                    <div style={{ display: 'flex', gap: '0.5rem' }}>
                      <button className="btn-small" onClick={() => handleToggle(tr.id, tr.enabled)}>
                        {tr.enabled ? 'Disable' : 'Enable'}
                      </button>
                      <button className="btn-small" onClick={() => handleDelete(tr.id)} style={{ color: '#ef4444' }}>Delete</button>
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
