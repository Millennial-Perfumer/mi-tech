import { API_BASE } from './api';
import { useState, useEffect } from 'react';
import { useToast } from './ToastContext';
import { useConfirm } from './ConfirmContext';

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

interface AutomationEvent {
  id: number;
  name: string;
  topic: string;
  description: string;
}

interface AutomationTriggersProps {
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
  userRole?: string;
}

export function AutomationTriggers({ fetchWithAuth, userRole = 'read' }: AutomationTriggersProps) {
  const { success: toastSuccess, error: toastError } = useToast();
  const { confirm: customConfirm } = useConfirm();
  const [triggers, setTriggers] = useState<Trigger[]>([]);
  const [templates, setTemplates] = useState<Template[]>([]);
  const [events, setEvents] = useState<AutomationEvent[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isEventModalOpen, setIsEventModalOpen] = useState(false);

  const [formData, setFormData] = useState({
    topic: '',
    templateID: ''
  });

  const [eventFormData, setEventFormData] = useState({
    name: '',
    topic: '',
    description: ''
  });

  const fetchData = async () => {
    setIsLoading(true);
    try {
      const [triggersResp, templatesResp, eventsResp] = await Promise.all([
        fetchWithAuth(`${API_BASE}/api/automation/whatsapp/triggers`),
        fetchWithAuth(`${API_BASE}/api/automation/whatsapp/templates`),
        fetchWithAuth(`${API_BASE}/api/automation/whatsapp/events`)
      ]);
      const triggersData = await triggersResp.json();
      const templatesData = await templatesResp.json();
      const eventsData = await eventsResp.json();
      
      setTriggers(triggersData || []);
      setTemplates((templatesData || []).filter((t: Template) => t.status === 'approved' || t.status === 'APPROVED'));
      setEvents(eventsData || []);
      
      if (eventsData && eventsData.length > 0 && !formData.topic) {
        setFormData(prev => ({ ...prev, topic: eventsData[0].topic }));
      }
    } catch (err) {
      console.error('Failed to fetch data:', err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => { fetchData(); }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.templateID || !formData.topic) return;
    
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
        toastSuccess('Trigger mapping registered successfully');
        setIsModalOpen(false);
        setFormData({ topic: events[0]?.topic || '', templateID: '' });
        fetchData();
      } else {
        const errText = await resp.text();
        toastError(`Failed to save trigger: ${errText}`);
      }
    } catch (err) {
      console.error('Failed to save trigger:', err);
      toastError('Network error while saving trigger mapping.');
    } finally {
      setIsSaving(false);
    }
  };

  const handleCreateEvent = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!eventFormData.name || !eventFormData.topic) return;

    setIsSaving(true);
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/automation/whatsapp/events`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(eventFormData)
      });
      if (resp.ok) {
        toastSuccess('Event created successfully');
        setIsEventModalOpen(false);
        setEventFormData({ name: '', topic: '', description: '' });
        fetchData();
      } else {
        const errText = await resp.text();
        toastError(`Failed to create event: ${errText}`);
      }
    } catch (err) {
      console.error('Failed to create event:', err);
      toastError('Network error while creating event.');
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
      if (resp.ok) {
        toastSuccess(`Trigger ${!currentEnabled ? 'enabled' : 'disabled'} successfully`);
        fetchData();
      } else {
        toastError('Failed to toggle trigger status.');
      }
    } catch (err) {
      console.error('Failed to toggle trigger:', err);
      toastError('Network error while toggling trigger.');
    }
  };

  const handleDelete = async (id: number) => {
    const confirmed = await customConfirm({
        title: 'Delete Trigger Mapping',
        message: 'Are you sure you want to delete this trigger mapping? This will stop automated messages for this event.',
        variant: 'danger',
        confirmLabel: 'Delete'
    });

    if (!confirmed) return;

    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/automation/whatsapp/triggers?id=${id}`, { method: 'DELETE' });
      if (resp.ok) {
        toastSuccess('Trigger mapping deleted successfully');
        fetchData();
      } else {
        toastError('Failed to delete trigger mapping.');
      }
    } catch (err) {
      console.error('Failed to delete trigger:', err);
      toastError('Network error while deleting trigger mapping.');
    }
  };

  return (
    <div className="automation-page">
      <div style={{ 
        display: 'flex', 
        justifyContent: 'space-between', 
        alignItems: 'center', 
        marginBottom: '2rem',
        padding: '1.25rem 1.5rem',
        background: 'var(--surface-color)',
        borderRadius: '16px',
        boxShadow: 'var(--shadow-sm)',
        border: '1px solid var(--border-color)'
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '2rem' }}>
          <div>
            <h1 style={{ margin: 0, fontSize: '1.5rem', fontWeight: 800, color: 'var(--text-primary)', letterSpacing: '-0.025em' }}>Webhook Triggers</h1>
            <p style={{ margin: '4px 0 0 0', color: 'var(--text-secondary)', fontSize: '0.9rem', fontWeight: 500 }}>
              Map Shopify events to automated message templates
            </p>
          </div>
          <div style={{ width: '1px', height: '40px', backgroundColor: 'var(--border-color)' }}></div>
        </div>

        {userRole === 'admin' && (
          <div style={{ display: 'flex', gap: '1rem' }}>
            <button 
              className="btn-secondary" 
              onClick={() => setIsEventModalOpen(true)}
              style={{
                backgroundColor: 'rgba(255, 255, 255, 0.05)',
                color: 'var(--text-primary)',
                border: '1px solid var(--border-color)',
                padding: '0.65rem 1.25rem',
                borderRadius: '10px',
                fontSize: '0.875rem',
                fontWeight: 600,
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem'
              }}
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="16"></line><line x1="8" y1="12" x2="16" y2="12"></line></svg>
              Create Event
            </button>
            <button 
              className="btn-primary" 
              onClick={() => setIsModalOpen(true)}
              style={{
                backgroundColor: 'var(--status-active)',
                color: 'white',
                border: 'none',
                padding: '0.65rem 1.25rem',
                borderRadius: '10px',
                fontSize: '0.875rem',
                fontWeight: 600,
                boxShadow: 'var(--shadow-md)',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem'
              }}
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
              Register Trigger
            </button>
          </div>
        )}
      </div>

      {/* Trigger Registration Modal */}
      {isModalOpen && (
        <div className="modal-overlay" style={{
          position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
          backgroundColor: 'rgba(0, 0, 0, 0.7)', backdropFilter: 'blur(8px)',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          zIndex: 1000, padding: '20px'
        }} onClick={() => setIsModalOpen(false)}>
          <div className="modal-content" style={{
            background: 'var(--surface-color)', border: '1px solid var(--border-color)',
            borderRadius: '24px', padding: '2.5rem', width: '100%', maxWidth: '500px',
            boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.5)', position: 'relative'
          }} onClick={e => e.stopPropagation()}>
            <button onClick={() => setIsModalOpen(false)} style={{ position: 'absolute', top: '1.5rem', right: '1.5rem', background: 'none', border: 'none', color: 'var(--text-tertiary)', cursor: 'pointer' }}>
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
            </button>
            <h2 style={{ fontSize: '1.5rem', fontWeight: 800, marginBottom: '0.5rem' }}>Register Trigger</h2>
            <p style={{ color: 'var(--text-tertiary)', fontSize: '0.9rem', marginBottom: '2rem' }}>Map an event to an approved automated template.</p>
            <form onSubmit={handleSubmit}>
              <div className="form-group" style={{ marginBottom: '1.5rem' }}>
                <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 700, textTransform: 'uppercase', color: 'var(--text-tertiary)', marginBottom: '0.5rem' }}>Event</label>
                <select className="input-field" style={{ width: '100%', padding: '0.75rem', borderRadius: '12px', border: '1px solid var(--border-color)', background: 'var(--bg-input)', color: 'var(--text-primary)' }}
                  value={formData.topic} onChange={e => setFormData({...formData, topic: e.target.value})} required>
                  <option value="">Select an event...</option>
                  {events.map(ev => <option key={ev.topic} value={ev.topic}>{ev.name} ({ev.topic})</option>)}
                </select>
              </div>
              <div className="form-group" style={{ marginBottom: '2rem' }}>
                <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 700, textTransform: 'uppercase', color: 'var(--text-tertiary)', marginBottom: '0.5rem' }}>Template</label>
                <select className="input-field" style={{ width: '100%', padding: '0.75rem', borderRadius: '12px', border: '1px solid var(--border-color)', background: 'var(--bg-input)', color: 'var(--text-primary)' }}
                  value={formData.templateID} onChange={e => setFormData({...formData, templateID: e.target.value})} required>
                  <option value="">Select an approved template...</option>
                  {templates.map(t => <option key={t.id} value={t.id}>{t.template_name}</option>)}
                </select>
              </div>
              <div style={{ display: 'flex', gap: '1rem' }}>
                <button type="submit" className="btn-primary" disabled={isSaving} style={{ flex: 2, backgroundColor: 'var(--status-active)', color: 'white', padding: '1rem', borderRadius: '14px', fontWeight: 700, border: 'none', cursor: 'pointer' }}>
                  {isSaving ? 'Registering...' : 'Register Mapping'}
                </button>
                <button type="button" className="btn-secondary" onClick={() => setIsModalOpen(false)} style={{ flex: 1, background: 'var(--bg-input)', border: '1px solid var(--border-color)', padding: '1rem', borderRadius: '14px', fontWeight: 600, cursor: 'pointer' }}>Cancel</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Event Creation Modal */}
      {isEventModalOpen && (
        <div className="modal-overlay" style={{
          position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
          backgroundColor: 'rgba(0, 0, 0, 0.7)', backdropFilter: 'blur(8px)',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          zIndex: 1000, padding: '20px'
        }} onClick={() => setIsEventModalOpen(false)}>
          <div className="modal-content" style={{
            background: 'var(--surface-color)', border: '1px solid var(--border-color)',
            borderRadius: '24px', padding: '2.5rem', width: '100%', maxWidth: '500px',
            boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.5)', position: 'relative'
          }} onClick={e => e.stopPropagation()}>
            <button onClick={() => setIsEventModalOpen(false)} style={{ position: 'absolute', top: '1.5rem', right: '1.5rem', background: 'none', border: 'none', color: 'var(--text-tertiary)', cursor: 'pointer' }}>
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
            </button>
            <h2 style={{ fontSize: '1.5rem', fontWeight: 800, marginBottom: '0.5rem' }}>Create Event</h2>
            <p style={{ color: 'var(--text-tertiary)', fontSize: '0.9rem', marginBottom: '2rem' }}>Define a new trigger event for your store.</p>
            <form onSubmit={handleCreateEvent}>
              <div className="form-group" style={{ marginBottom: '1.5rem' }}>
                <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 700, textTransform: 'uppercase', color: 'var(--text-tertiary)', marginBottom: '0.5rem' }}>Event Name</label>
                <input type="text" className="input-field" placeholder="e.g. Flash Sale" style={{ width: '100%', padding: '0.75rem', borderRadius: '12px', border: '1px solid var(--border-color)', background: 'var(--bg-input)', color: 'var(--text-primary)' }}
                  value={eventFormData.name} onChange={e => setEventFormData({...eventFormData, name: e.target.value})} required />
              </div>
              <div className="form-group" style={{ marginBottom: '1.5rem' }}>
                <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 700, textTransform: 'uppercase', color: 'var(--text-tertiary)', marginBottom: '0.5rem' }}>Topic / Tag</label>
                <input type="text" className="input-field" placeholder="e.g. special/flash_sale" style={{ width: '100%', padding: '0.75rem', borderRadius: '12px', border: '1px solid var(--border-color)', background: 'var(--bg-input)', color: 'var(--text-primary)' }}
                  value={eventFormData.topic} onChange={e => setEventFormData({...eventFormData, topic: e.target.value})} required />
              </div>
              <div className="form-group" style={{ marginBottom: '2rem' }}>
                <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 700, textTransform: 'uppercase', color: 'var(--text-tertiary)', marginBottom: '0.5rem' }}>Description</label>
                <textarea className="input-field" rows={3} placeholder="Optional description..." style={{ width: '100%', padding: '0.75rem', borderRadius: '12px', border: '1px solid var(--border-color)', background: 'var(--bg-input)', color: 'var(--text-primary)', resize: 'none' }}
                  value={eventFormData.description} onChange={e => setEventFormData({...eventFormData, description: e.target.value})} />
              </div>
              <div style={{ display: 'flex', gap: '1rem' }}>
                <button type="submit" className="btn-primary" disabled={isSaving} style={{ flex: 2, backgroundColor: 'var(--status-active)', color: 'white', padding: '1rem', borderRadius: '14px', fontWeight: 700, border: 'none', cursor: 'pointer' }}>
                  {isSaving ? 'Creating...' : 'Create Event'}
                </button>
                <button type="button" className="btn-secondary" onClick={() => setIsEventModalOpen(false)} style={{ flex: 1, background: 'var(--bg-input)', border: '1px solid var(--border-color)', padding: '1rem', borderRadius: '14px', fontWeight: 600, cursor: 'pointer' }}>Cancel</button>
              </div>
            </form>
          </div>
        </div>
      )}

      <div className="table-container" style={{ width: '100%', overflowX: 'auto', background: 'var(--surface-color)', borderRadius: '20px', border: '1px solid var(--border-color)', padding: '0.5rem' }}>
        <table style={{ width: '100%', borderCollapse: 'separate', borderSpacing: 0 }}>
          <thead>
            <tr>
              <th style={{ padding: '1.25rem 1.5rem', fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', borderBottom: '1px solid var(--border-color)' }}>Event</th>
              <th style={{ padding: '1.25rem 1.5rem', fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', borderBottom: '1px solid var(--border-color)' }}>Template</th>
              <th style={{ padding: '1.25rem 1.5rem', fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', borderBottom: '1px solid var(--border-color)' }}>Status</th>
              <th style={{ padding: '1.25rem 1.5rem', fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', borderBottom: '1px solid var(--border-color)' }}>Created</th>
              <th style={{ padding: '1.25rem 1.5rem', fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', borderBottom: '1px solid var(--border-color)', textAlign: 'right' }}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr><td colSpan={5} style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-secondary)' }}>Loading triggers...</td></tr>
            ) : triggers.length === 0 ? (
              <tr><td colSpan={5} style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-secondary)' }}>No triggers found.</td></tr>
            ) : (
              triggers.map(tr => (
                <tr key={tr.id} style={{ transition: 'background-color 0.2s' }} className="hover-row">
                  <td style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid var(--border-color-light)' }}>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                      <span style={{ fontWeight: 600, color: 'var(--text-primary)', fontSize: '0.9rem' }}>
                        {events.find(ev => ev.topic === tr.webhook_topic)?.name || tr.webhook_topic}
                      </span>
                      <code style={{ fontSize: '0.7rem', color: 'var(--text-tertiary)', opacity: 0.7 }}>{tr.webhook_topic}</code>
                    </div>
                  </td>
                  <td style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid var(--border-color-light)', color: 'var(--text-primary)', fontWeight: 500 }}>
                    {templates.find(t => t.id === tr.template_id)?.template_name || `ID: ${tr.template_id}`}
                  </td>
                  <td style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid var(--border-color-light)' }}>
                    <span className={`badge-pill ${tr.enabled ? 'badge-pill-success' : 'badge-pill-warning'}`} style={{ fontSize: '0.7rem', padding: '4px 10px' }}>
                      <span className="dot" />
                      {tr.enabled ? 'ACTIVE' : 'INACTIVE'}
                    </span>
                  </td>
                  <td style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid var(--border-color-light)', color: 'var(--text-secondary)', fontSize: '0.85rem' }}>
                    {new Date(tr.created_at).toLocaleDateString('en-GB')}
                  </td>
                  <td style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid var(--border-color-light)', textAlign: 'right' }}>
                    {userRole === 'admin' ? (
                      <div style={{ display: 'flex', gap: '1rem', justifyContent: 'flex-end', alignItems: 'center' }}>
                        <button onClick={() => handleToggle(tr.id, tr.enabled)} title={tr.enabled ? "Pause" : "Resume"} style={{ background: 'none', border: 'none', cursor: 'pointer', color: tr.enabled ? 'var(--status-active)' : 'var(--text-tertiary)' }} className="action-btn">
                          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round"><path d="M18.36 6.64a9 9 0 1 1-12.73 0"></path><line x1="12" y1="2" x2="12" y2="12"></line></svg>
                        </button>
                        <button onClick={() => handleDelete(tr.id)} title="Remove" style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--status-error)' }} className="action-btn">
                          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>
                        </button>
                      </div>
                    ) : <span style={{ color: 'var(--text-tertiary)', fontSize: '0.8rem' }}>View Only</span>}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      <style>{`
        .modal-overlay { animation: fadeIn 0.2s ease-out; }
        @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
        .hover-row:hover { background-color: rgba(255, 255, 255, 0.02); }
        .action-btn:hover { transform: scale(1.1); }
      `}</style>
    </div>
  );
}
