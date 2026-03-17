import React, { useState, useEffect } from 'react';

interface Template {
  id: number;
  template_name: string;
  language: string;
  category: string;
  body: string;
  status: string;
}

interface ManualWhatsAppModalProps {
  isOpen: boolean;
  onClose: () => void;
  orderId: string | number;
  orderNumber: string;
  customerName: string;
  token: string | null;
}

export const ManualWhatsAppModal: React.FC<ManualWhatsAppModalProps> = ({
  isOpen,
  onClose,
  orderId,
  orderNumber,
  customerName,
  token
}) => {
  const [templates, setTemplates] = useState<Template[]>([]);
  const [loading, setLoading] = useState(false);
  const [sending, setSending] = useState(false);
  const [selectedTemplateId, setSelectedTemplateId] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen && token) {
      fetchTemplates();
    }
  }, [isOpen, token]);

  const fetchTemplates = async () => {
    setLoading(true);
    try {
      const response = await fetch('http://localhost:8080/api/automation/whatsapp/templates', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      const data = await response.json();
      setTemplates(data || []);
    } catch (err) {
      setError('Failed to load templates');
    } finally {
      setLoading(false);
    }
  };

  const handleSend = async () => {
    if (!selectedTemplateId) return;

    setSending(true);
    setError(null);
    try {
      const response = await fetch('http://localhost:8080/api/automation/whatsapp/send-manual', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          order_id: orderId,
          template_id: selectedTemplateId
        })
      });

      const data = await response.json();
      if (data.success) {
        onClose();
        alert('Message sent successfully!');
      } else {
        setError(data.message || 'Failed to send message');
      }
    } catch (err) {
      setError('Network error sending message');
    } finally {
      setSending(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={e => e.stopPropagation()} style={{ maxWidth: '600px', width: '90%' }}>
        <div className="modal-header">
          <h3>Send WhatsApp Message</h3>
          <button className="close-btn" onClick={onClose}>&times;</button>
        </div>
        
        <div className="modal-body">
          <div style={{ marginBottom: '1.5rem', padding: '1rem', background: '#f8fafc', borderRadius: '8px', border: '1px solid #e2e8f0' }}>
            <div style={{ fontSize: '0.9rem', color: '#64748b', marginBottom: '0.25rem' }}>Sending to</div>
            <div style={{ fontWeight: 600, fontSize: '1.1rem' }}>{customerName}</div>
            <div style={{ fontSize: '0.85rem', color: '#94a3b8' }}>Order #{orderNumber}</div>
          </div>

          <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Select Template</label>
          {loading ? (
            <div style={{ textAlign: 'center', padding: '1rem' }}>Loading templates...</div>
          ) : (
            <div style={{ maxHeight: '300px', overflowY: 'auto', border: '1px solid #e2e8f0', borderRadius: '8px' }}>
              {templates.length === 0 ? (
                <div style={{ padding: '1rem', textAlign: 'center', color: '#64748b' }}>No templates found.</div>
              ) : (
                templates.map(t => (
                  <div 
                    key={t.id}
                    className={`template-option ${selectedTemplateId === t.id ? 'selected' : ''}`}
                    onClick={() => setSelectedTemplateId(t.id)}
                    style={{
                      padding: '1rem',
                      borderBottom: '1px solid #f1f5f9',
                      cursor: 'pointer',
                      background: selectedTemplateId === t.id ? '#f0f9ff' : 'transparent',
                      borderColor: selectedTemplateId === t.id ? '#0ea5e9' : '#f1f5f9'
                    }}
                  >
                    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.25rem' }}>
                      <span style={{ fontWeight: 600 }}>{t.template_name}</span>
                      <span className={`status-badge status-${t.status.toLowerCase()}`} style={{ fontSize: '0.7rem' }}>
                        {t.status}
                      </span>
                    </div>
                    <div style={{ fontSize: '0.8rem', color: '#64748b', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                      {t.body}
                    </div>
                  </div>
                ))
              )}
            </div>
          )}

          {error && <div style={{ color: '#ef4444', marginTop: '1rem', fontSize: '0.9rem' }}>{error}</div>}
        </div>

        <div className="modal-footer" style={{ display: 'flex', gap: '1rem', justifyContent: 'flex-end', marginTop: '1.5rem' }}>
          <button className="btn btn-secondary" onClick={onClose} disabled={sending}>Cancel</button>
          <button 
            className="btn btn-primary" 
            onClick={handleSend} 
            disabled={sending || !selectedTemplateId}
            style={{ minWidth: '100px' }}
          >
            {sending ? 'Sending...' : 'Send Message'}
          </button>
        </div>
      </div>
    </div>
  );
};
