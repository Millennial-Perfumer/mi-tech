import React, { useState, useEffect } from 'react';
import { useToast } from './ToastContext';

interface CustomerFeedback {
  id: number;
  order_id: number;
  order_number: string;
  customer_name: string;
  rating: number;
  comment: string;
  message?: string;
  customer_phone?: string;
  created_at: string;
}

interface FeedbackProps {
  API_BASE: string;
  token: string | null;
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
}

const Feedback: React.FC<FeedbackProps> = ({ API_BASE, token, fetchWithAuth }) => {
  const [feedbacks, setFeedbacks] = useState<CustomerFeedback[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isScanning, setIsScanning] = useState(false);
  const [scanResults, setScanResults] = useState<any[]>([]);
  const [selectedIds, setSelectedIds] = useState<number[]>([]);
  const [isSending, setIsSending] = useState(false);
  const [isScanModalOpen, setIsScanModalOpen] = useState(false);
  const { error, success } = useToast();

  const fetchFeedbacks = async () => {
    if (!token) return;
    setIsLoading(true);
    try {
      const response = await fetchWithAuth(`${API_BASE}/api/orders/feedback`);
      const data = await response.json();
      if (data.success) {
        setFeedbacks(data.feedback || []);
      } else {
        error(data.message || 'Failed to fetch feedback');
      }
    } catch (err) {
      console.error('Failed to fetch feedback:', err);
      error('Network error fetching feedback');
    } finally {
      setIsLoading(false);
    }
  };

  const handleScan = async () => {
    setIsScanning(true);
    try {
      const response = await fetchWithAuth(`${API_BASE}/api/feedback/scan`);
      const data = await response.json();
      if (response.ok && data.success) {
        const orders = data.orders || [];
        setScanResults(orders);
        setSelectedIds(orders.map((o: any) => o.id));
        if (orders.length > 0) {
          setIsScanModalOpen(true);
        } else {
          success('No new orders for feedback found.');
        }
      } else {
        error(`Failed to scan for orders: ${data.message || response.statusText} (${response.status})`);
      }
    } catch (err) {
      console.error('Scan Error:', err);
      error('Network error during order scan');
    } finally {
      setIsScanning(false);
    }
  };

  const toggleSelect = (id: number) => {
    setSelectedIds(prev => 
      prev.includes(id) ? prev.filter(i => i !== id) : [...prev, id]
    );
  };

  const handleBulkSend = async () => {
    if (selectedIds.length === 0) return;
    
    setIsSending(true);
    try {
      const response = await fetchWithAuth(`${API_BASE}/api/feedback/bulk-send`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ order_ids: selectedIds })
      });
      const data = await response.json();
      if (data.success) {
        success(`Successfully sent ${data.sent} feedback requests!`);
        setScanResults([]);
        setSelectedIds([]);
        setIsScanModalOpen(false);
        fetchFeedbacks();
      } else {
        error(data.message || 'Failed to send bulk feedback');
      }
    } catch (err) {
      error('Network error sending bulk feedback');
    } finally {
      setIsSending(false);
    }
  };

  useEffect(() => {
    fetchFeedbacks();
  }, [token]);

  const renderStars = (rating: number) => {
    return Array.from({ length: 5 }, (_, i) => (
      <svg
        key={i}
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill={i < rating ? "#f59e0b" : "none"}
        stroke={i < rating ? "#f59e0b" : "currentColor"}
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        style={{ marginRight: '2px' }}
      >
        <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
      </svg>
    ));
  };

  return (
    <div className="feedback-container">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2.5rem' }}>
        <div>
          <h2 style={{ fontSize: '1.25rem', fontWeight: 700, margin: 0 }}>Customer Sentiment Analysis</h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', marginTop: '4px' }}>Monitor ratings and reviews across your delivered orders</p>
        </div>
        <div style={{ display: 'flex', gap: '1rem' }}>
          <button 
            onClick={handleScan}
            disabled={isScanning}
            style={{
              background: 'var(--accent-color)',
              color: 'white',
              border: 'none',
              borderRadius: '12px',
              padding: '0.75rem 1.25rem',
              fontWeight: 600,
              fontSize: '0.875rem',
              cursor: isScanning ? 'not-allowed' : 'pointer',
              opacity: isScanning ? 0.7 : 1,
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
              boxShadow: 'var(--shadow-sm)'
            }}
          >
            {isScanning ? 'Scanning...' : (
              <>
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
                Scan for Orders
              </>
            )}
          </button>
          <div style={{ 
            background: 'var(--surface-color)', 
            padding: '0.75rem 1rem', 
            borderRadius: '12px', 
            border: '1px solid var(--border-color)',
            display: 'flex',
            alignItems: 'center',
            gap: '0.75rem',
            boxShadow: 'var(--shadow-xs)'
          }}>
            <span style={{ fontSize: '0.7rem', color: 'var(--text-secondary)', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.05em' }}>Average Rating</span>
            <span style={{ fontWeight: 800, color: 'var(--accent-color)', fontSize: '1.25rem', lineHeight: 1 }}>
              {feedbacks.length > 0 
                ? (feedbacks.reduce((acc, f) => acc + f.rating, 0) / feedbacks.length).toFixed(1) 
                : 'N/A'}
            </span>
          </div>
        </div>
      </div>

      {/* --- SCAN RESULTS MODAL --- */}
      {isScanModalOpen && (
        <div className="modal-overlay">
          <div className="premium-modal wide" style={{ padding: 0 }}>
            <div style={{ padding: '1.5rem 2rem', borderBottom: '1px solid var(--border-color)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <h2 style={{ fontSize: '1.25rem', margin: 0 }}>Orders for Feedback</h2>
                <p style={{ margin: '4px 0 0', fontSize: '0.85rem' }}>Select orders to send WhatsApp feedback requests</p>
              </div>
              <div style={{ display: 'flex', gap: '1rem' }}>
                <button 
                  onClick={() => setIsScanModalOpen(false)}
                  className="btn-secondary"
                  style={{ border: 'none', background: 'transparent', color: 'var(--text-secondary)', fontSize: '0.875rem' }}
                >
                  Cancel
                </button>
                <button 
                  onClick={handleBulkSend}
                  disabled={isSending || selectedIds.length === 0}
                  className="btn-primary"
                  style={{
                    padding: '0.6rem 1.25rem',
                    fontSize: '0.9rem',
                    borderRadius: '12px'
                  }}
                >
                  {isSending ? 'Sending...' : `Send to ${selectedIds.length} Selected`}
                </button>
              </div>
            </div>
            
            <div style={{ maxHeight: '60vh', overflowY: 'auto' }}>
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead style={{ position: 'sticky', top: 0, background: 'var(--surface-color)', zIndex: 1, borderBottom: '1px solid var(--border-color)' }}>
                  <tr style={{ textAlign: 'left' }}>
                    <th style={{ padding: '1rem 2rem', width: '40px' }}>
                      <input 
                        type="checkbox" 
                        checked={selectedIds.length === scanResults.length}
                        onChange={(e) => setSelectedIds(e.target.checked ? scanResults.map(c => c.id) : [])}
                      />
                    </th>
                    <th style={{ padding: '1rem 2rem', fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>Customer / Order</th>
                    <th style={{ padding: '1rem 2rem', fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>Delivered</th>
                    <th style={{ padding: '1rem 2rem', fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>Preview Link</th>
                  </tr>
                </thead>
                <tbody>
                  {scanResults.map((item) => (
                    <tr key={item.id} style={{ borderBottom: '1px solid var(--border-color)', transition: 'background 0.2s' }} className="hover-row">
                      <td style={{ padding: '1rem 2rem' }}>
                        <input 
                          type="checkbox" 
                          checked={selectedIds.includes(item.id)}
                          onChange={() => toggleSelect(item.id)}
                        />
                      </td>
                      <td style={{ padding: '1rem 2rem' }}>
                        <div style={{ fontWeight: 600, fontSize: '0.95rem', color: 'var(--text-primary)' }}>{item.customer_name}</div>
                        <div style={{ fontSize: '0.75rem', color: 'var(--accent-color)', fontWeight: 600 }}>{item.order_number}</div>
                      </td>
                      <td style={{ padding: '1rem 2rem', fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
                        {new Date(item.delivered_at).toLocaleString([], { dateStyle: 'short', timeStyle: 'short' })}
                      </td>
                      <td style={{ padding: '1rem 2rem' }}>
                        <a href={item.feedback_url} target="_blank" rel="noopener noreferrer" style={{ fontSize: '0.75rem', color: 'var(--accent-color)', textDecoration: 'none', fontWeight: 600, border: '1px solid var(--accent-subtle)', padding: '4px 8px', borderRadius: '6px', background: 'var(--accent-subtle)' }}>
                          Test Link
                        </a>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      <div style={{ background: 'var(--surface-color)', borderRadius: '20px', border: '1px solid var(--border-color)', overflow: 'hidden', boxShadow: 'var(--shadow-sm)' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ textAlign: 'left', borderBottom: '1px solid var(--border-color)', background: 'rgba(255,255,255,0.02)' }}>
              <th style={{ padding: '1.25rem 1.5rem', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>Customer</th>
              <th style={{ padding: '1.25rem 1.5rem', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>Order</th>
              <th style={{ padding: '1.25rem 1.5rem', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>Rating</th>
              <th style={{ padding: '1.25rem 1.5rem', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>Message</th>
              <th style={{ padding: '1.25rem 1.5rem', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>Date</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr><td colSpan={5} style={{ textAlign: 'center', padding: '5rem', color: 'var(--text-tertiary)' }}>
                <div className="dot-flashing" style={{ margin: '0 auto 1.5rem' }}></div>
                Analyzing customer sentiment...
              </td></tr>
            ) : feedbacks.length === 0 ? (
              <tr><td colSpan={5} style={{ textAlign: 'center', padding: '6rem 2rem', color: 'var(--text-secondary)' }}>
                <div style={{ opacity: 0.3, marginBottom: '1.5rem' }}>
                  <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg>
                </div>
                <div style={{ fontWeight: 600, fontSize: '1.1rem', color: 'var(--text-primary)' }}>No feedback collected yet</div>
                <p style={{ fontSize: '0.85rem', marginTop: '0.5rem' }}>Send feedback requests to start gathering customer insights.</p>
              </td></tr>
            ) : (
              feedbacks.map((item, idx) => (
                <tr key={item.id} style={{ borderBottom: idx === feedbacks.length - 1 ? 'none' : '1px solid var(--border-color)', transition: 'background 0.2s' }} className="hover-row">
                  <td style={{ padding: '1.5rem' }}>
                    <div style={{ fontWeight: 700, fontSize: '0.95rem', color: 'var(--text-primary)' }}>{item.customer_name}</div>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', marginTop: '2px' }}>{item.customer_phone}</div>
                  </td>
                  <td style={{ padding: '1.5rem' }}>
                    <div style={{ fontSize: '0.875rem', color: 'var(--accent-color)', fontWeight: 700, background: 'var(--accent-subtle)', display: 'inline-block', padding: '2px 8px', borderRadius: '6px' }}>
                      {item.order_number}
                    </div>
                  </td>
                  <td style={{ padding: '1.5rem' }}>
                    <div style={{ display: 'flex', gap: '2px' }}>{renderStars(item.rating)}</div>
                    <div style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-tertiary)', marginTop: '4px', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                      {item.rating === 5 ? 'Exceptional' : item.rating === 4 ? 'Great' : item.rating === 3 ? 'Average' : item.rating === 2 ? 'Poor' : 'Critical'}
                    </div>
                  </td>
                  <td style={{ padding: '1.5rem' }}>
                    <div style={{ 
                      fontSize: '0.9rem', 
                      maxWidth: '400px', 
                      lineHeight: '1.6', 
                      color: item.message ? 'var(--text-primary)' : 'var(--text-tertiary)',
                      fontStyle: item.message ? 'normal' : 'italic',
                      fontWeight: 500
                    }}>
                      {item.message || "Customer didn't leave a secondary comment."}
                    </div>
                  </td>
                  <td style={{ padding: '1.5rem', fontSize: '0.8rem', color: 'var(--text-secondary)', fontWeight: 600 }}>
                    {new Date(item.created_at).toLocaleDateString([], { day: '2-digit', month: 'short', year: 'numeric' })}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default Feedback;
