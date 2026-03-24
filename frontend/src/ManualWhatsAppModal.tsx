import { API_BASE } from './api';
import React, { useState, useEffect } from 'react';
import { useToast } from './ToastContext';

interface Trigger {
  id: number;
  webhook_topic: string;
  template_id: number;
  template_name: string;
  template_body: string;
  template_status: string;
}

const webhookLabels: Record<string, string> = {
  'orders/create': 'Order Placed',
  'orders/assigned': 'Order Assigned',
  'orders/fulfilled': 'Order Dispatched',
  'orders/out_for_delivery': 'Order Out for Delivery',
  'orders/delivered': 'Order Delivered',
  'orders/updated': 'Order Updated',
  'orders/cancelled': 'Order Cancelled',
  'orders/paid': 'Order Paid',
};

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
  const { success: toastSuccess, error: toastError } = useToast();
  const [triggers, setTriggers] = useState<Trigger[]>([]);
  const [loading, setLoading] = useState(false);
  const [sending, setSending] = useState(false);
  const [selectedTemplateId, setSelectedTemplateId] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen && token) {
      fetchTriggers();
    }
  }, [isOpen, token]);

  const fetchTriggers = async () => {
    setLoading(true);
    try {
      const response = await fetch(`${API_BASE}/api/automation/whatsapp/triggers`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      const data = await response.json();
      // Only show triggers that have a template mapped and are not archived
      setTriggers((data || []).filter((t: Trigger) => t.template_id > 0 && t.template_status !== 'ARCHIVED'));
    } catch (err) {
      setError('Failed to load triggers');
    } finally {
      setLoading(false);
    }
  };

  const handleSend = async () => {
    if (!selectedTemplateId) return;

    setSending(true);
    setError(null);
    try {
      const response = await fetch(`${API_BASE}/api/automation/whatsapp/send-manual`, {
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
        toastSuccess('Message sent successfully!');
        onClose();
      } else {
        const errMsg = data.message || 'Failed to send message';
        setError(errMsg);
        toastError(errMsg);
      }
    } catch (err) {
      setError('Network error sending message');
      toastError('Network error while sending WhatsApp message');
    } finally {
      setSending(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={e => e.stopPropagation()} style={{ maxWidth: '600px', width: '90%' }}>
        <div className="modal-header">
          <h3>Send Webhook Notification</h3>
          <button className="close-btn" onClick={onClose}>&times;</button>
        </div>
        
        <div className="modal-body">
          <div style={{ marginBottom: '1.5rem', padding: '1rem', background: '#f8fafc', borderRadius: '8px', border: '1px solid #e2e8f0' }}>
            <div style={{ fontSize: '0.9rem', color: '#64748b', marginBottom: '0.25rem' }}>Sending to</div>
            <div style={{ fontWeight: 600, fontSize: '1.1rem' }}>{customerName}</div>
            <div style={{ fontSize: '0.85rem', color: '#94a3b8' }}>Order #{orderNumber}</div>
          </div>

          <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Select Event to Trigger</label>
          {loading ? (
            <div style={{ textAlign: 'center', padding: '1rem' }}>Loading events...</div>
          ) : (
            <div style={{ maxHeight: '300px', overflowY: 'auto', border: '1px solid #e2e8f0', borderRadius: '8px' }}>
              {triggers.length === 0 ? (
                <div style={{ padding: '1rem', textAlign: 'center', color: '#64748b' }}>No automated triggers configured.</div>
              ) : (
                triggers.map(t => (
                  <div 
                    key={t.id}
                    className={`template-option ${selectedTemplateId === t.template_id ? 'selected' : ''}`}
                    onClick={() => setSelectedTemplateId(t.template_id)}
                    style={{
                      padding: '1rem',
                      borderBottom: '1px solid #f1f5f9',
                      cursor: 'pointer',
                      background: selectedTemplateId === t.template_id ? '#f0f9ff' : 'transparent',
                      borderColor: selectedTemplateId === t.template_id ? '#0ea5e9' : '#f1f5f9'
                    }}
                  >
                    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.25rem' }}>
                      <span style={{ fontWeight: 600, color: '#0f172a' }}>
                        {webhookLabels[t.webhook_topic] || t.webhook_topic}
                      </span>
                      <span style={{ fontSize: '0.7rem', color: '#64748b', fontStyle: 'italic' }}>
                        {t.template_name}
                      </span>
                    </div>
                    <div style={{ fontSize: '0.8rem', color: '#475569', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                      {t.template_body}
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
