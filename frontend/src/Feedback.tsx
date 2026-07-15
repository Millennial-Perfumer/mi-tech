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
  admin_comment?: string;
  created_at: string;
}

interface FeedbackProps {
  API_BASE: string;
  token: string | null;
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
  onNavigate?: (tab: string) => void;
  initialSelectedOrderId?: number | null;
  clearInitialSelectedOrderId?: () => void;
}

const Feedback: React.FC<FeedbackProps> = ({ 
  API_BASE, 
  token, 
  fetchWithAuth, 
  onNavigate,
  initialSelectedOrderId,
  clearInitialSelectedOrderId
}) => {
  const [feedbacks, setFeedbacks] = useState<CustomerFeedback[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isScanning, setIsScanning] = useState(false);
  const [scanResults, setScanResults] = useState<any[]>([]);
  const [selectedIds, setSelectedIds] = useState<number[]>([]);
  const [isSending, setIsSending] = useState(false);
  const [isScanModalOpen, setIsScanModalOpen] = useState(false);
  const [configStatus, setConfigStatus] = useState<{ is_configured: boolean, missing_items: string[] } | null>(null);
  const [isCheckingConfig, setIsCheckingConfig] = useState(true);
  const [sortConfig, setSortConfig] = useState<{ key: string, direction: 'asc' | 'desc' } | null>({ key: 'delivered_at', direction: 'desc' });
  const { error, success } = useToast();

  // Admin comment dialog states
  const [selectedFeedback, setSelectedFeedback] = useState<CustomerFeedback | null>(null);
  const [isAdminModalOpen, setIsAdminModalOpen] = useState(false);
  const [adminCommentText, setAdminCommentText] = useState('');
  const [isSavingComment, setIsSavingComment] = useState(false);

  const handleOpenCommentModal = (feedback: CustomerFeedback) => {
    setSelectedFeedback(feedback);
    setAdminCommentText(feedback.admin_comment || '');
    setIsAdminModalOpen(true);
  };

  const handleSaveComment = async () => {
    if (!selectedFeedback) return;
    setIsSavingComment(true);
    try {
      const response = await fetchWithAuth(`${API_BASE}/api/orders/feedback/comment?id=${selectedFeedback.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ admin_comment: adminCommentText })
      });
      const data = await response.json();
      if (data.success) {
        success(data.message || 'Comment updated successfully');
        setIsAdminModalOpen(false);
        setSelectedFeedback(null);
        fetchFeedbacks();
      } else {
        error(data.message || 'Failed to update comment');
      }
    } catch (err) {
      console.error('Failed to update comment:', err);
      error('Network error updating comment');
    } finally {
      setIsSavingComment(false);
    }
  };

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

  const fetchConfigStatus = async () => {
    if (!token) return;
    setIsCheckingConfig(true);
    try {
      const response = await fetchWithAuth(`${API_BASE}/api/feedback/config-status`);
      const data = await response.json();
      if (data.success) {
        setConfigStatus({
          is_configured: data.is_configured,
          missing_items: data.missing_items
        });
      }
    } catch (err) {
      console.error('Failed to fetch config status:', err);
    } finally {
      setIsCheckingConfig(false);
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
        // Only auto-select orders that have a customer phone number
        setSelectedIds(orders.filter((o: any) => o.customer_phone).map((o: any) => o.id));
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

  const handleSort = (key: string) => {
    let direction: 'asc' | 'desc' = 'asc';
    if (sortConfig && sortConfig.key === key && sortConfig.direction === 'asc') {
      direction = 'desc';
    }
    setSortConfig({ key, direction });
  };

  const sortedScanResults = React.useMemo(() => {
    if (!sortConfig) return scanResults;
    return [...scanResults].sort((a, b) => {
      const aVal = a[sortConfig.key];
      const bVal = b[sortConfig.key];
      
      if (aVal < bVal) return sortConfig.direction === 'asc' ? -1 : 1;
      if (aVal > bVal) return sortConfig.direction === 'asc' ? 1 : -1;
      return 0;
    });
  }, [scanResults, sortConfig]);

  useEffect(() => {
    fetchFeedbacks();
    fetchConfigStatus();
  }, [token]);

  // Auto-open modal if initialSelectedOrderId is passed
  useEffect(() => {
    if (initialSelectedOrderId && feedbacks.length > 0) {
      const match = feedbacks.find(f => f.order_id === initialSelectedOrderId);
      if (match) {
        handleOpenCommentModal(match);
      }
      if (clearInitialSelectedOrderId) {
        clearInitialSelectedOrderId();
      }
    }
  }, [initialSelectedOrderId, feedbacks, clearInitialSelectedOrderId]);

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
                    <th style={{ padding: '1rem 2rem', fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>
                      <button onClick={() => handleSort('customer_name')} style={{ background: 'none', border: 'none', color: 'inherit', font: 'inherit', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '4px' }}>
                        Customer {sortConfig?.key === 'customer_name' && (sortConfig.direction === 'asc' ? '↑' : '↓')}
                      </button>
                    </th>
                    <th style={{ padding: '1rem 2rem', fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>
                      <button onClick={() => handleSort('order_number')} style={{ background: 'none', border: 'none', color: 'inherit', font: 'inherit', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '4px' }}>
                        Order # {sortConfig?.key === 'order_number' && (sortConfig.direction === 'asc' ? '↑' : '↓')}
                      </button>
                    </th>
                    <th style={{ padding: '1rem 2rem', fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>
                      <button onClick={() => handleSort('delivered_at')} style={{ background: 'none', border: 'none', color: 'inherit', font: 'inherit', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '4px' }}>
                        Delivered {sortConfig?.key === 'delivered_at' && (sortConfig.direction === 'asc' ? '↑' : '↓')}
                      </button>
                    </th>
                    <th style={{ padding: '1rem 2rem', fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>Action</th>
                  </tr>
                </thead>
                <tbody>
                  {sortedScanResults.map((item) => (
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
                      </td>
                      <td style={{ padding: '1rem 2rem' }} title={!item.customer_phone ? "No customer phone number found. WhatsApp feedback cannot be sent." : ""}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                          <div style={{ fontSize: '0.875rem', color: 'var(--accent-color)', fontWeight: 700, background: 'var(--accent-subtle)', display: 'inline-block', padding: '2px 8px', borderRadius: '6px' }}>
                            {item.order_number}
                          </div>
                          {!item.customer_phone && (
                            <div style={{ color: '#ef4444', display: 'flex', alignItems: 'center', gap: '4px', fontSize: '0.75rem', fontWeight: 600 }}>
                              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                                <circle cx="12" cy="12" r="10"></circle>
                                <line x1="12" y1="8" x2="12" y2="12"></line>
                                <line x1="12" y1="16" x2="12.01" y2="16"></line>
                              </svg>
                              <span>(No Phone)</span>
                            </div>
                          )}
                        </div>
                      </td>
                      <td style={{ padding: '1rem 2rem', fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
                        <div style={{ fontWeight: 500 }}>{new Date(item.delivered_at).toLocaleDateString([], { day: '2-digit', month: 'short' })}</div>
                        <div style={{ fontSize: '0.7rem', opacity: 0.6 }}>{new Date(item.delivered_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</div>
                      </td>
                      <td style={{ padding: '1rem 2rem' }}>
                        <a href={item.feedback_url} target="_blank" rel="noopener noreferrer" style={{ fontSize: '0.75rem', color: 'var(--accent-color)', textDecoration: 'none', fontWeight: 600, border: '1px solid var(--accent-subtle)', padding: '4px 8px', borderRadius: '6px', background: 'var(--accent-subtle)' }}>
                          Preview
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

      {/* --- ADMIN COMMENT MODAL --- */}
      {isAdminModalOpen && selectedFeedback && (
        <div className="modal-overlay" onClick={() => setIsAdminModalOpen(false)}>
          <div className="premium-modal" style={{ maxWidth: '500px' }} onClick={e => e.stopPropagation()}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1.5rem' }}>
              <div>
                <h2 style={{ fontSize: '1.5rem', fontWeight: 800, margin: 0 }}>Feedback Admin Note</h2>
                <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', marginTop: '4px', marginBottom: 0 }}>
                  Add an internal note or action taken for this customer review
                </p>
              </div>
              <button 
                onClick={() => setIsAdminModalOpen(false)}
                style={{ background: 'none', border: 'none', color: 'var(--text-tertiary)', cursor: 'pointer', padding: '4px' }}
              >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
              </button>
            </div>

            <div style={{ background: 'rgba(255,255,255,0.02)', padding: '1rem', borderRadius: '12px', border: '1px solid var(--border-color)', marginBottom: '1.5rem' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem', fontSize: '0.85rem' }}>
                <span style={{ color: 'var(--text-secondary)', fontWeight: 600 }}>Customer:</span>
                <span style={{ color: 'var(--text-primary)', fontWeight: 700 }}>{selectedFeedback.customer_name}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem', fontSize: '0.85rem' }}>
                <span style={{ color: 'var(--text-secondary)', fontWeight: 600 }}>Order:</span>
                <span style={{ color: 'var(--accent-color)', fontWeight: 700 }}>{selectedFeedback.order_number}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem', fontSize: '0.85rem' }}>
                <span style={{ color: 'var(--text-secondary)', fontWeight: 600 }}>Rating:</span>
                <span style={{ display: 'flex', gap: '2px', alignItems: 'center' }}>
                  {renderStars(selectedFeedback.rating)}
                </span>
              </div>
              <div style={{ borderTop: '1px solid var(--border-color)', marginTop: '0.75rem', paddingTop: '0.75rem' }}>
                <span style={{ color: 'var(--text-secondary)', fontSize: '0.8rem', fontWeight: 700, textTransform: 'uppercase', display: 'block', marginBottom: '4px' }}>Customer Comment:</span>
                <p style={{ color: 'var(--text-primary)', fontSize: '0.9rem', fontStyle: selectedFeedback.message ? 'normal' : 'italic', margin: 0, lineHeight: '1.5' }}>
                  {selectedFeedback.message || "No comment left."}
                </p>
              </div>
            </div>

            <div className="sync-form-group" style={{ marginBottom: '1.5rem' }}>
              <label style={{ fontSize: '0.8rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', display: 'block', marginBottom: '0.6rem' }}>Admin Comment / Note</label>
              <textarea
                value={adminCommentText}
                onChange={e => setAdminCommentText(e.target.value)}
                placeholder="Add an internal follow-up comment, resolution notes, or private remarks..."
                rows={4}
                style={{
                  width: '100%',
                  background: 'var(--bg-input)',
                  border: '2px solid var(--border-color)',
                  borderRadius: '12px',
                  padding: '0.75rem 1rem',
                  color: 'var(--text-primary)',
                  fontFamily: 'inherit',
                  fontSize: '0.95rem',
                  lineHeight: '1.5',
                  resize: 'vertical',
                  outline: 'none',
                  boxSizing: 'border-box'
                }}
              />
            </div>

            <div className="modal-actions" style={{ display: 'flex', gap: '1rem' }}>
              <button 
                onClick={() => setIsAdminModalOpen(false)} 
                className="btn-secondary"
                style={{
                  flex: 1,
                  padding: '0.75rem',
                  borderRadius: '12px',
                  fontWeight: 600,
                  cursor: 'pointer'
                }}
              >
                Cancel
              </button>
              <button 
                onClick={handleSaveComment} 
                disabled={isSavingComment}
                className="btn-primary"
                style={{
                  flex: 1,
                  padding: '0.75rem',
                  borderRadius: '12px',
                  fontWeight: 600,
                  cursor: isSavingComment ? 'not-allowed' : 'pointer'
                }}
              >
                {isSavingComment ? 'Saving...' : 'Save Comment'}
              </button>
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
              <th style={{ padding: '1.25rem 1.5rem', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>Admin Comment</th>
              <th style={{ padding: '1.25rem 1.5rem', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>Date</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr><td colSpan={6} style={{ textAlign: 'center', padding: '5rem', color: 'var(--text-tertiary)' }}>
                <div className="dot-flashing" style={{ margin: '0 auto 1.5rem' }}></div>
                Analyzing customer sentiment...
              </td></tr>
            ) : feedbacks.length === 0 ? (
              <tr><td colSpan={6} style={{ textAlign: 'center', padding: '6rem 2rem', color: 'var(--text-secondary)' }}>
                <div style={{ opacity: 0.3, marginBottom: '1.5rem' }}>
                  <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg>
                </div>
                <div style={{ fontWeight: 600, fontSize: '1.1rem', color: 'var(--text-primary)' }}>No feedback collected yet</div>
                <p style={{ fontSize: '0.85rem', marginTop: '0.5rem' }}>Send feedback requests to start gathering customer insights.</p>
              </td></tr>
            ) : (
              feedbacks.map((item, idx) => (
                <tr 
                  key={item.id} 
                  style={{ borderBottom: idx === feedbacks.length - 1 ? 'none' : '1px solid var(--border-color)', transition: 'background 0.2s', cursor: 'pointer' }} 
                  className="hover-row"
                  onClick={() => handleOpenCommentModal(item)}
                >
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
                      maxWidth: '300px', 
                      lineHeight: '1.6', 
                      color: item.message ? 'var(--text-primary)' : 'var(--text-tertiary)',
                      fontStyle: item.message ? 'normal' : 'italic',
                      fontWeight: 500,
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap'
                    }}>
                      {item.message || "Customer didn't leave a secondary comment."}
                    </div>
                  </td>
                  <td style={{ padding: '1.5rem' }}>
                    {item.admin_comment ? (
                      <div style={{ 
                        fontSize: '0.9rem', 
                        maxWidth: '250px', 
                        color: 'var(--text-primary)',
                        fontWeight: 600,
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap'
                      }} title={item.admin_comment}>
                        {item.admin_comment}
                      </div>
                    ) : (
                      <span style={{ fontSize: '0.85rem', color: 'var(--text-tertiary)', fontStyle: 'italic' }}>—</span>
                    )}
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

      {/* --- SETUP WIZARD / ONBOARDING --- */}
      {!isCheckingConfig && configStatus && !configStatus.is_configured && (
        <div className="modal-overlay" style={{ backdropFilter: 'blur(8px)', backgroundColor: 'rgba(0,0,0,0.6)', zIndex: 1000 }}>
          <div className="premium-modal" style={{ maxWidth: '480px', border: '1px solid rgba(255,255,255,0.1)' }}>
            <div className="modal-header-icon" style={{ background: 'linear-gradient(135deg, #f59e0b, #d97706)', marginBottom: '1.5rem' }}>
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"></path></svg>
            </div>
            
            <h2 style={{ fontSize: '1.25rem', fontWeight: 800, textAlign: 'center', marginBottom: '0.5rem' }}>Setup Required</h2>
            <p style={{ color: 'var(--text-secondary)', textAlign: 'center', marginBottom: '1.5rem', fontSize: '0.85rem', lineHeight: 1.5 }}>
              The feedback system is almost ready. Follow these steps to authorize and map your communication.
            </p>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', marginBottom: '2rem' }}>
              {[
                { 
                  id: 'template', 
                  title: 'WhatsApp Template', 
                  desc: 'Ensure your template is approved in Meta and added to your Automation Templates.',
                  isFixed: !configStatus.missing_items.includes('template_name') && !configStatus.missing_items.includes('template_not_found')
                },
                { 
                  id: 'mapping', 
                  title: 'Variable Mapping', 
                  desc: 'Ensure your template has defined variable mappings in the Automation tab.',
                  isFixed: !configStatus.missing_items.includes('mapping_missing') && !configStatus.missing_items.includes('template_name') && !configStatus.missing_items.includes('template_not_found')
                },
                { 
                  id: 'settings', 
                  title: 'Configure Feedback', 
                  desc: 'Go to Settings -> Feedback and pick your template and survey URL.',
                  isFixed: !configStatus.missing_items.includes('template_name') && !configStatus.missing_items.includes('template_not_found') && !configStatus.missing_items.includes('base_url')
                }
              ].map((step, i) => (
                <div key={step.id} style={{ display: 'flex', gap: '1rem', opacity: step.isFixed ? 0.6 : 1, background: 'rgba(255,255,255,0.03)', padding: '1rem', borderRadius: '12px', border: '1px solid var(--border-color)' }}>
                  <div style={{ 
                    width: '24px', 
                    height: '24px', 
                    borderRadius: '50%', 
                    background: step.isFixed ? '#22c55e' : 'var(--bg-hover)',
                    color: step.isFixed ? 'white' : 'var(--text-tertiary)',
                    display: 'flex', 
                    alignItems: 'center', 
                    justifyContent: 'center',
                    fontSize: '0.75rem',
                    fontWeight: 800,
                    flexShrink: 0
                  }}>
                    {step.isFixed ? '✓' : i + 1}
                  </div>
                  <div>
                    <h3 style={{ fontSize: '0.9rem', margin: 0, fontWeight: 700, color: step.isFixed ? '#22c55e' : 'var(--text-primary)' }}>{step.title}</h3>
                    <p style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', margin: '4px 0 0', lineHeight: 1.4 }}>{step.desc}</p>
                  </div>
                </div>
              ))}
            </div>

            <div style={{ display: 'flex', gap: '1rem' }}>
              <button 
                onClick={() => onNavigate ? onNavigate('settings') : window.location.hash = '#settings'} 
                className="btn-primary" 
                style={{ flex: 1, padding: '0.75rem', fontSize: '0.85rem' }}
              >
                Go to Settings
              </button>
              <button 
                onClick={() => fetchConfigStatus()} 
                className="btn-secondary" 
                style={{ flex: 1, padding: '0.75rem', fontSize: '0.85rem' }}
              >
                Refresh
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Feedback;
