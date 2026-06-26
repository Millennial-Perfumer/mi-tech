import { useState, useEffect } from 'react';
import { API_BASE } from './api';
import { AbandonedCartsDashboard } from './AbandonedCartsDashboard';

interface LineItem {
  title: string;
  variant_title?: string;
  sku?: string;
  quantity: number;
  price: number | string;
}

interface AbandonedCheckout {
  id: number;
  store_id: string;
  checkout_id: string;
  checkout_token: string;
  cart_token: string;
  email: string;
  phone: string;
  customer_name: string;
  checkout_url: string;
  line_items: any;
  total_price: number;
  currency: string;
  completed: boolean;
  completed_at?: string;
  order_id?: string;
  recovery_status: string;
  recovery_attempts: number;
  recovery_message_sent_at?: string;
  last_error?: string;
  marketing_consent: boolean;
  sms_consent: boolean;
  city?: string;
  province?: string;
  country?: string;
  zip?: string;
  abandoned_at: string;
  created_at: string;
  updated_at: string;
}

interface AbandonedCartsProps {
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
  startDate?: string;
  endDate?: string;
}

export function AbandonedCarts({ fetchWithAuth, startDate, endDate }: AbandonedCartsProps) {
  const [activeSubTab, setActiveSubTab] = useState<'list' | 'analytics'>('list');
  const [checkouts, setCheckouts] = useState<AbandonedCheckout[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [page, setPage] = useState(1);
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState(''); // '', 'COMPLETED', 'ABANDONED', 'PENDING', 'SENT', 'FAILED'
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [sendingId, setSendingId] = useState<number | null>(null);
  const [toastMsg, setToastMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [selectedCart, setSelectedCart] = useState<AbandonedCheckout | null>(null);

  const limit = 15;

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchQuery);
      setPage(1);
    }, 500);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  const fetchCheckouts = async (silent = false) => {
    if (!silent) {
      if (checkouts.length > 0) setIsRefreshing(true);
      else setIsLoading(true);
    }
    try {
      const statusParam = statusFilter ? `&status=${statusFilter}` : '';
      const dateParam = (startDate && endDate) ? `&start_date=${startDate}&end_date=${endDate}` : '';
      const url = `${API_BASE}/api/abandoned-checkouts?page=${page}&limit=${limit}&search=${encodeURIComponent(debouncedSearch)}${statusParam}${dateParam}`;
      const resp = await fetchWithAuth(url);
      if (resp.ok) {
        const data = await resp.json();
        setCheckouts(data.checkouts || []);
        setTotalCount(data.total_count || 0);
      }
    } catch (err) {
      console.error('Failed to fetch abandoned checkouts:', err);
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  };

  useEffect(() => {
    fetchCheckouts();
  }, [page, debouncedSearch, statusFilter, startDate, endDate]);

  const handleSendRecovery = async (id: number) => {
    setSendingId(id);
    setToastMsg(null);
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/abandoned-checkouts/recover?id=${id}`, {
        method: 'POST',
      });
      const data = await resp.json();
      if (resp.ok && data.success) {
        setToastMsg({ type: 'success', text: 'WhatsApp recovery message dispatched successfully!' });
        fetchCheckouts(true);
        // Update selectedCart view if open
        if (selectedCart && selectedCart.id === id) {
          setSelectedCart((prev: AbandonedCheckout | null) => prev ? { ...prev, recovery_status: 'SENT', recovery_attempts: prev.recovery_attempts + 1 } : null);
        }
      } else {
        setToastMsg({ type: 'error', text: data.message || 'Failed to dispatch recovery message.' });
      }
    } catch (err: any) {
      setToastMsg({ type: 'error', text: err.message || 'Network error sending recovery message.' });
    } finally {
      setSendingId(null);
      setTimeout(() => setToastMsg(null), 5000);
    }
  };

  const handleDeleteCheckout = async (id: number) => {
    const confirmed = window.confirm("Are you sure you want to delete this abandoned checkout record?");
    if (!confirmed) return;
    
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/abandoned-checkouts?id=${id}`, {
        method: 'DELETE',
      });
      const data = await resp.json();
      if (resp.ok && data.success) {
        setToastMsg({ type: 'success', text: 'Abandoned checkout deleted successfully!' });
        setCheckouts(prev => prev.filter(c => c.id !== id));
        setTotalCount(prev => prev - 1);
        if (selectedCart && selectedCart.id === id) {
          setSelectedCart(null);
        }
      } else {
        setToastMsg({ type: 'error', text: data.message || 'Failed to delete checkout.' });
      }
    } catch (err: any) {
      setToastMsg({ type: 'error', text: err.message || 'Network error deleting checkout.' });
    } finally {
      setTimeout(() => setToastMsg(null), 5000);
    }
  };

  const [updatingStatus, setUpdatingStatus] = useState(false);

  const handleStatusChange = async (id: number, val: string) => {
    setUpdatingStatus(true);
    try {
      const completed = val === 'RECOVERED';
      const recoveryStatus = completed ? 'SENT' : val;

      const resp = await fetchWithAuth(`${API_BASE}/api/abandoned-checkouts/status`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id, recovery_status: recoveryStatus, completed }),
      });
      const data = await resp.json();
      if (resp.ok && data.success) {
        setToastMsg({ type: 'success', text: 'Status updated successfully!' });
        
        // Update local list
        setCheckouts(prev => prev.map(c => c.id === id ? { ...c, recovery_status: recoveryStatus, completed } : c));
        
        // Update selected detail
        if (selectedCart && selectedCart.id === id) {
          setSelectedCart((prev: any) => prev ? { ...prev, recovery_status: recoveryStatus, completed } : null);
        }
      } else {
        setToastMsg({ type: 'error', text: data.message || 'Failed to update status.' });
      }
    } catch (err: any) {
      setToastMsg({ type: 'error', text: err.message || 'Network error updating status.' });
    } finally {
      setUpdatingStatus(false);
      setTimeout(() => setToastMsg(null), 5000);
    }
  };

  const getRecoveryBadge = (status: string, completed: boolean) => {
    if (completed) {
      return <span className="badge-pill badge-pill-success"><span className="dot" style={{ backgroundColor: '#10b981' }}></span>RECOVERED</span>;
    }
    switch (status.toUpperCase()) {
      case 'PENDING':
        return <span className="badge-pill badge-pill-gray"><span className="dot"></span>PENDING</span>;
      case 'PROCESSING':
        return <span className="badge-pill badge-pill-info"><span className="dot" style={{ backgroundColor: '#38bdf8' }}></span>PROCESSING</span>;
      case 'SENT':
        return <span className="badge-pill badge-pill-success"><span className="dot" style={{ backgroundColor: '#10b981' }}></span>SENT</span>;
      case 'FAILED':
        return <span className="badge-pill badge-pill-danger"><span className="dot"></span>FAILED</span>;
      case 'CANCELLED':
        return <span className="badge-pill badge-pill-gray"><span className="dot" style={{ backgroundColor: '#6b7280' }}></span>CANCELLED</span>;
      default:
        return <span className="badge-pill badge-pill-gray"><span className="dot"></span>{status}</span>;
    }
  };

  const parseLineItems = (raw: any): LineItem[] => {
    if (!raw) return [];
    try {
      if (typeof raw === 'string') {
        return JSON.parse(raw);
      }
      return raw;
    } catch {
      return [];
    }
  };

  const formatPhoneNumber = (num: string) => {
    if (!num) return '-';
    const digits = num.replace(/\D/g, '');
    let normalized = digits;
    if (digits.length === 10) normalized = '91' + digits;
    return `+${normalized.slice(0, 2)} ${normalized.slice(2, 7)} ${normalized.slice(7)}`;
  };

  const totalPages = Math.ceil(totalCount / limit);

  return (
    <div className="automation-page">
      {toastMsg && (
        <div className={`toast-message ${toastMsg.type}`} style={{
          position: 'fixed',
          top: '20px',
          right: '20px',
          backgroundColor: toastMsg.type === 'success' ? 'rgba(16, 185, 129, 0.95)' : 'rgba(239, 68, 68, 0.95)',
          color: 'white',
          padding: '0.85rem 1.5rem',
          borderRadius: '8px',
          zIndex: 9999,
          fontWeight: 600,
          boxShadow: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
          backdropFilter: 'blur(10px)',
          border: '1px solid rgba(255,255,255,0.1)',
          animation: 'slideIn 0.3s ease-out'
        }}>
          {toastMsg.text}
        </div>
      )}
      {isRefreshing && (
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--accent-color)', fontSize: '0.85rem', fontWeight: 600, marginBottom: '1rem' }}>
          <div className="dot-flashing"></div>
          Refreshing Carts...
        </div>
      )}

      {/* Modern Sub-Tab Switcher */}
      <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '2rem', background: 'var(--bg-input)', padding: '4px', borderRadius: '10px', width: 'fit-content' }}>
        <button
          onClick={() => setActiveSubTab('list')}
          style={{
            background: activeSubTab === 'list' ? 'var(--bg-card)' : 'transparent',
            color: activeSubTab === 'list' ? 'var(--text-primary)' : 'var(--text-secondary)',
            border: 'none',
            padding: '0.5rem 1.25rem',
            borderRadius: '8px',
            fontWeight: 700,
            cursor: 'pointer',
            fontSize: '0.85rem',
            boxShadow: activeSubTab === 'list' ? 'var(--shadow-sm)' : 'none',
            transition: 'all 0.2s'
          }}
        >
          Recovery List
        </button>
        <button
          onClick={() => setActiveSubTab('analytics')}
          style={{
            background: activeSubTab === 'analytics' ? 'var(--bg-card)' : 'transparent',
            color: activeSubTab === 'analytics' ? 'var(--text-primary)' : 'var(--text-secondary)',
            border: 'none',
            padding: '0.5rem 1.25rem',
            borderRadius: '8px',
            fontWeight: 700,
            cursor: 'pointer',
            fontSize: '0.85rem',
            boxShadow: activeSubTab === 'analytics' ? 'var(--shadow-sm)' : 'none',
            transition: 'all 0.2s'
          }}
        >
          Analytics Dashboard
        </button>
      </div>

      {activeSubTab === 'analytics' ? (
        <AbandonedCartsDashboard
          fetchWithAuth={fetchWithAuth}
          startDate={startDate}
          endDate={endDate}
          onSendRecovery={handleSendRecovery}
          sendingId={sendingId}
        />
      ) : (
        <>
          <div className="filter-bar">
            <div className="search-wrapper">
              <input
                type="text"
                className="search-input"
                style={{ paddingLeft: '3.5rem' }}
                placeholder="Search by customer name, phone, email, checkout ID..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
              <svg className="search-icon" aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <circle cx="11" cy="11" r="8"></circle>
                <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
              </svg>
              {searchQuery && (
                <button
                  className="clear-search"
                  onClick={() => { setSearchQuery(''); setPage(1); }}
                >
                  <svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                    <line x1="18" y1="6" x2="6" y2="18"></line>
                    <line x1="6" y1="6" x2="18" y2="18"></line>
                  </svg>
                </button>
              )}
            </div>

            <div className="select-wrapper">
              <select
                className="custom-select"
                value={statusFilter}
                onChange={(e) => {
                  setStatusFilter(e.target.value);
                  setPage(1);
                }}
              >
                <option value="">All Recovery States</option>
                <option value="ABANDONED">Abandoned (Unrecovered)</option>
                <option value="COMPLETED">Recovered (Order Completed)</option>
                <option value="PENDING">Pending (Scheduled)</option>
                <option value="SENT">Recovery Sent</option>
                <option value="FAILED">Recovery Failed</option>
              </select>
              <div style={{ position: 'absolute', right: '1rem', top: '50%', transform: 'translateY(-50%)', pointerEvents: 'none', color: 'var(--text-tertiary)' }}>
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
                  <path d="m6 9 6 6 6-6" />
                </svg>
              </div>
            </div>

            <button className="refresh-btn" onClick={() => fetchCheckouts()}>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <path d="M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8"></path>
                <path d="M21 3v5h-5"></path>
                <path d="M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16"></path>
                <path d="M3 21v-5h5"></path>
              </svg>
              Refresh
            </button>
          </div>

          <div className="table-container shadow-sm" style={{ overflowX: 'auto', width: '100%', borderRadius: '12px' }}>
            <table className="logs-table" style={{ width: '100%', borderCollapse: 'collapse', minWidth: '1000px' }}>
              <thead>
                <tr>
                  <th style={{ padding: '1rem 1.25rem' }}>Abandoned Date</th>
                  <th style={{ padding: '1rem 1.25rem' }}>Customer</th>
                  <th style={{ padding: '1rem 1.25rem' }}>Cart Details</th>
                  <th style={{ padding: '1rem 1.25rem' }}>Total Value</th>
                  <th style={{ padding: '1rem 1.25rem' }}>Recovery Status</th>
                  <th style={{ padding: '1rem 1.25rem', textAlign: 'center' }}>Attempts</th>
                  <th style={{ padding: '1rem 1.25rem', textAlign: 'right' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {isLoading ? (
                  <tr>
                    <td colSpan={7} style={{ textAlign: 'center', padding: '3rem' }}>
                      <div className="dot-flashing" style={{ margin: '0 auto' }}></div>
                    </td>
                  </tr>
                ) : checkouts.length === 0 ? (
                  <tr>
                    <td colSpan={7} style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-tertiary)' }}>No abandoned checkouts found.</td>
                  </tr>
                ) : (
                  checkouts.map((c) => {
                    let itemsCount = 0;
                    let itemsSummary = '';
                    if (c.line_items) {
                      try {
                        const parsed = typeof c.line_items === 'string' ? JSON.parse(c.line_items) : c.line_items;
                        if (Array.isArray(parsed)) {
                          itemsCount = parsed.reduce((acc: number, item: any) => acc + (item.quantity || 0), 0);
                          itemsSummary = parsed.map((item: any) => `${item.title} (x${item.quantity})`).join(', ');
                        }
                      } catch (e) {
                        console.error('Failed to parse line items', e);
                      }
                    }

                    return (
                      <tr
                        key={c.id}
                        className="log-row"
                        onClick={() => setSelectedCart(c)}
                        style={{ cursor: 'pointer', transition: 'background-color 0.2s' }}
                      >
                        <td style={{ verticalAlign: 'middle' }}>
                          <div style={{ display: 'flex', flexDirection: 'column' }}>
                            <span style={{ fontWeight: 600, color: 'var(--text-primary)' }}>
                              {new Date(c.abandoned_at).toLocaleDateString('en-IN', { day: 'numeric', month: 'short', year: 'numeric' })}
                            </span>
                            <span style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', marginTop: '2px' }}>
                              {new Date(c.abandoned_at).toLocaleTimeString('en-IN', { hour: '2-digit', minute: '2-digit' })}
                            </span>
                          </div>
                        </td>
                        <td style={{ verticalAlign: 'middle' }}>
                          <div style={{ display: 'flex', flexDirection: 'column', gap: '3px' }}>
                            <span style={{ fontWeight: 700, color: 'var(--text-primary)' }}>{c.customer_name || 'Anonymous'}</span>
                            {c.phone && <span style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>{formatPhoneNumber(c.phone)}</span>}
                            {c.email && <span style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>{c.email}</span>}
                          </div>
                        </td>
                        <td style={{ verticalAlign: 'middle', maxWidth: '300px' }}>
                          <div style={{ display: 'flex', flexDirection: 'column', gap: '2px' }}>
                            <span style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{itemsCount} {itemsCount === 1 ? 'item' : 'items'}</span>
                            <span
                              style={{
                                fontSize: '0.75rem',
                                color: 'var(--text-tertiary)',
                                overflow: 'hidden',
                                textOverflow: 'ellipsis',
                                whiteSpace: 'nowrap',
                              }}
                              title={itemsSummary}
                            >
                              {itemsSummary || 'No details'}
                            </span>
                          </div>
                        </td>
                        <td style={{ fontWeight: 700, fontSize: '1.05rem', verticalAlign: 'middle', color: 'var(--text-primary)' }}>
                          {c.currency} {parseFloat(c.total_price as any).toLocaleString('en-IN', { minimumFractionDigits: 2 })}
                        </td>
                        <td style={{ verticalAlign: 'middle' }}>
                          <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                            {getRecoveryBadge(c.recovery_status, c.completed)}
                            {c.completed_at && c.order_id && (
                              <span style={{ fontSize: '0.75rem', color: 'var(--status-active)', fontWeight: 600 }}>
                                Order: #{c.order_id}
                              </span>
                            )}
                          </div>
                        </td>
                        <td style={{ fontWeight: 600, textAlign: 'center', verticalAlign: 'middle', color: 'var(--text-primary)' }}>{c.recovery_attempts}</td>
                        <td style={{ textAlign: 'right', verticalAlign: 'middle', whiteSpace: 'nowrap' }}>
                          <div style={{ display: 'inline-flex', justifyContent: 'flex-end', gap: '8px', whiteSpace: 'nowrap', verticalAlign: 'middle' }}>
                            {c.checkout_url && (
                              <a
                                href={c.checkout_url}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="page-btn"
                                onClick={(e) => e.stopPropagation()}
                                title="Open Checkout URL"
                                style={{ 
                                  display: 'inline-flex', 
                                  alignItems: 'center', 
                                  justifyContent: 'center',
                                  width: '36px',
                                  height: '36px', 
                                  padding: '0',
                                  borderRadius: '8px',
                                  backgroundColor: 'var(--bg-input)',
                                  border: '1px solid var(--border-color)',
                                  color: 'var(--text-primary)',
                                  transition: 'all 0.2s ease'
                                }}
                              >
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path><polyline points="15 3 21 3 21 9"></polyline><line x1="10" y1="14" x2="21" y2="3"></line></svg>
                              </a>
                            )}
                            <button
                              className="btn-primary"
                              style={{
                                display: 'inline-flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                width: '36px',
                                height: '36px',
                                padding: '0',
                                borderRadius: '8px',
                                backgroundColor: c.completed ? 'var(--bg-hover)' : 'var(--accent-color)',
                                color: c.completed ? 'var(--text-tertiary)' : 'white',
                                cursor: c.completed || sendingId === c.id ? 'not-allowed' : 'pointer',
                                transition: 'all 0.2s ease',
                                border: 'none'
                              }}
                              disabled={c.completed || sendingId === c.id}
                              title={c.completed ? "Already recovered" : "Send WhatsApp recovery message"}
                              onClick={(e) => {
                                e.stopPropagation();
                                handleSendRecovery(c.id);
                              }}
                            >
                              {sendingId === c.id ? (
                                <div className="dot-flashing" style={{ width: '4px', height: '4px' }}></div>
                              ) : (
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><line x1="22" y1="2" x2="11" y2="13"></line><polygon points="22 2 15 22 11 13 2 9 22 2"></polygon></svg>
                              )}
                            </button>
                            <button
                              className="btn-danger"
                              style={{
                                display: 'inline-flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                width: '36px',
                                height: '36px',
                                padding: '0',
                                borderRadius: '8px',
                                backgroundColor: 'var(--status-danger-bg)',
                                color: 'var(--status-danger)',
                                cursor: 'pointer',
                                transition: 'all 0.2s ease',
                                border: 'none'
                              }}
                              title="Delete abandoned checkout"
                              onClick={(e) => {
                                e.stopPropagation();
                                handleDeleteCheckout(c.id);
                              }}
                            >
                              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>
                            </button>
                          </div>
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>

          {totalCount > 0 && (
            <div className="pagination-container">
              <div className="pagination-info">
                Showing carts <strong>{(page - 1) * limit + 1}</strong> – <strong>{Math.min(page * limit, totalCount)}</strong> of <strong>{totalCount}</strong>
              </div>
              <div className="pagination-controls">
                <button
                  className="page-btn"
                  onClick={() => setPage((p: number) => Math.max(1, p - 1))}
                  disabled={page === 1}
                >
                  Previous
                </button>
                <span className="current-page">
                  Page {page} of {totalPages || 1}
                </span>
                <button
                  className="page-btn"
                  onClick={() => setPage((p: number) => Math.min(totalPages, p + 1))}
                  disabled={page >= totalPages}
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </>
      )}

      {/* Checkout Details Modal */}
      {selectedCart && (
        <div className="modal-overlay" onClick={() => setSelectedCart(null)}>
          <div className="premium-modal wide" onClick={e => e.stopPropagation()} style={{ maxWidth: '650px', padding: '2rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '1rem' }}>
              <div>
                <h2 style={{ margin: 0, fontSize: '1.4rem' }}>Checkout Details</h2>
                <span style={{ fontSize: '0.8rem', color: 'var(--text-tertiary)' }}>ID: {selectedCart.checkout_id || 'N/A'}</span>
              </div>
              <button
                onClick={() => setSelectedCart(null)}
                style={{
                  background: 'none',
                  border: 'none',
                  color: 'var(--text-secondary)',
                  cursor: 'pointer',
                  padding: '4px'
                }}
              >
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                  <line x1="18" y1="6" x2="6" y2="18"></line>
                  <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg>
              </button>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.5rem', marginBottom: '1.5rem' }}>
              {/* Customer details */}
              <div style={{ background: 'var(--bg-input)', padding: '1rem', borderRadius: '12px', border: '1px solid var(--border-color)' }}>
                <h3 style={{ fontSize: '0.9rem', color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '0.75rem' }}>Customer Details</h3>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', fontSize: '0.9rem' }}>
                  <div><strong>Name:</strong> {selectedCart.customer_name || 'Anonymous'}</div>
                  <div><strong>Phone:</strong> {formatPhoneNumber(selectedCart.phone)}</div>
                  {selectedCart.email && <div><strong>Email:</strong> {selectedCart.email}</div>}
                  {(selectedCart.city || selectedCart.province || selectedCart.country || selectedCart.zip) && (
                    <div>
                      <strong>Location:</strong> {[selectedCart.city, selectedCart.province, selectedCart.country, selectedCart.zip].filter(Boolean).join(', ')}
                    </div>
                  )}
                  <div>
                    <strong>Marketing:</strong> {selectedCart.marketing_consent ? (
                      <span style={{ color: 'var(--status-active)', fontWeight: 600 }}>Subscribed</span>
                    ) : (
                      <span style={{ color: 'var(--text-tertiary)' }}>Not Subscribed</span>
                    )}
                  </div>
                  <div>
                    <strong>SMS Consent:</strong> {selectedCart.sms_consent ? (
                      <span style={{ color: 'var(--status-active)', fontWeight: 600 }}>Accepted</span>
                    ) : (
                      <span style={{ color: 'var(--text-tertiary)' }}>Not Accepted</span>
                    )}
                  </div>
                </div>
              </div>

              {/* Status details */}
              <div style={{ background: 'var(--bg-input)', padding: '1rem', borderRadius: '12px', border: '1px solid var(--border-color)' }}>
                <h3 style={{ fontSize: '0.9rem', color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '0.75rem' }}>Status & Times</h3>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', fontSize: '0.9rem' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap', marginBottom: '8px' }}>
                    <strong>State:</strong>
                    {(() => {
                      const getSelectColors = (status: string, completed: boolean) => {
                        if (completed) {
                          return { bg: 'var(--status-active-bg)', color: 'var(--status-active)' };
                        }
                        switch (status.toUpperCase()) {
                          case 'PENDING':
                            return { bg: 'var(--bg-input)', color: 'var(--text-secondary)' };
                          case 'PROCESSING':
                            return { bg: 'var(--status-warning-bg)', color: 'var(--status-warning)' };
                          case 'SENT':
                            return { bg: 'var(--status-active-bg)', color: 'var(--accent-color)' };
                          case 'FAILED':
                            return { bg: 'var(--status-danger-bg)', color: 'var(--status-danger)' };
                          case 'CANCELLED':
                            return { bg: 'var(--bg-input)', color: 'var(--text-secondary)' };
                          default:
                            return { bg: 'var(--bg-input)', color: 'var(--text-secondary)' };
                        }
                      };

                      const colors = getSelectColors(selectedCart.recovery_status, selectedCart.completed);
                      const currentState = selectedCart.completed ? 'RECOVERED' : selectedCart.recovery_status;
                      
                      return (
                        <div style={{ position: 'relative', display: 'inline-flex', alignItems: 'center' }}>
                          <select
                            value={currentState}
                            disabled={updatingStatus}
                            onChange={(e) => handleStatusChange(selectedCart.id, e.target.value)}
                            style={{
                              padding: '0.4rem 2rem 0.4rem 0.75rem',
                              fontSize: '0.75rem',
                              fontWeight: 800,
                              borderRadius: '8px',
                              border: '1px solid rgba(0,0,0,0.05)',
                              backgroundColor: colors.bg,
                              color: colors.color,
                              cursor: 'pointer',
                              appearance: 'none',
                              WebkitAppearance: 'none',
                              textTransform: 'uppercase',
                              transition: 'all 0.2s ease',
                              boxShadow: 'var(--shadow-sm)'
                            }}
                          >
                            <option value="PENDING" style={{ backgroundColor: 'var(--surface-color)', color: 'var(--text-primary)' }}>PENDING</option>
                            <option value="PROCESSING" style={{ backgroundColor: 'var(--surface-color)', color: 'var(--text-primary)' }}>PROCESSING</option>
                            <option value="SENT" style={{ backgroundColor: 'var(--surface-color)', color: 'var(--text-primary)' }}>SENT</option>
                            <option value="FAILED" style={{ backgroundColor: 'var(--surface-color)', color: 'var(--text-primary)' }}>FAILED</option>
                            <option value="CANCELLED" style={{ backgroundColor: 'var(--surface-color)', color: 'var(--text-primary)' }}>CANCELLED</option>
                            <option value="RECOVERED" style={{ backgroundColor: 'var(--surface-color)', color: 'var(--text-primary)' }}>RECOVERED</option>
                          </select>
                          <div style={{
                            position: 'absolute',
                            right: '0.6rem',
                            pointerEvents: 'none',
                            color: colors.color,
                            display: 'flex',
                            alignItems: 'center'
                          }}>
                            <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="4" strokeLinecap="round" strokeLinejoin="round">
                              <path d="m6 9 6 6 6-6" />
                            </svg>
                          </div>
                        </div>
                      );
                    })()}
                  </div>
                  <div><strong>Attempts:</strong> {selectedCart.recovery_attempts}</div>
                  <div><strong>Abandoned:</strong> {new Date(selectedCart.abandoned_at).toLocaleString()}</div>
                  {selectedCart.completed_at && <div><strong>Recovered At:</strong> {new Date(selectedCart.completed_at).toLocaleString()}</div>}
                  {selectedCart.order_id && <div><strong>Order ID:</strong> #{selectedCart.order_id}</div>}
                </div>
              </div>
            </div>

            {/* Cart Line Items */}
            <div style={{ marginBottom: '1.5rem' }}>
              <h3 style={{ fontSize: '0.9rem', color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '0.75rem' }}>Line Items</h3>
              <div style={{ border: '1px solid var(--border-color)', borderRadius: '12px', overflowY: 'auto', maxHeight: '240px' }}>
                <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                  <thead>
                    <tr style={{ background: 'var(--bg-input)' }}>
                      <th style={{ padding: '0.75rem 1rem', fontSize: '0.8rem' }}>Product</th>
                      <th style={{ padding: '0.75rem 1rem', fontSize: '0.8rem', textAlign: 'center' }}>Qty</th>
                      <th style={{ padding: '0.75rem 1rem', fontSize: '0.8rem', textAlign: 'right' }}>Price</th>
                    </tr>
                  </thead>
                  <tbody>
                    {parseLineItems(selectedCart.line_items).map((item, idx) => (
                      <tr key={idx} style={{ borderBottom: idx < parseLineItems(selectedCart.line_items).length - 1 ? '1px solid var(--border-color)' : 'none' }}>
                        <td style={{ padding: '0.75rem 1rem', fontSize: '0.85rem' }}>
                          <div style={{ fontWeight: 600 }}>{item.title}</div>
                          {item.variant_title && <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>{item.variant_title}</div>}
                          {item.sku && <div style={{ fontSize: '0.7rem', color: 'var(--text-tertiary)' }}>SKU: {item.sku}</div>}
                        </td>
                        <td style={{ padding: '0.75rem 1rem', fontSize: '0.85rem', textAlign: 'center' }}>{item.quantity}</td>
                        <td style={{ padding: '0.75rem 1rem', fontSize: '0.85rem', textAlign: 'right', fontWeight: 600 }}>
                          {selectedCart.currency} {((typeof item.price === 'string' ? parseFloat(item.price) : item.price) * item.quantity).toFixed(2)}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>

            {selectedCart.last_error && (
              <div style={{
                background: 'var(--status-danger-bg)',
                border: '1px solid rgba(239, 68, 68, 0.2)',
                borderRadius: '8px',
                padding: '0.75rem 1rem',
                color: 'var(--status-danger)',
                fontSize: '0.85rem',
                fontWeight: 500,
                marginBottom: '1.5rem'
              }}>
                <strong>Last Error:</strong> {selectedCart.last_error}
              </div>
            )}

            {/* Actions in modal */}
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '1rem', borderTop: '1px solid var(--border-color)', paddingTop: '1.25rem' }}>
              <button
                className="btn-secondary"
                onClick={() => setSelectedCart(null)}
                style={{ height: '38px', borderRadius: '8px', padding: '0 1.25rem' }}
              >
                Close
              </button>
              {selectedCart.checkout_url && (
                <a
                  href={selectedCart.checkout_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="page-btn"
                  style={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    gap: '6px',
                    textDecoration: 'none',
                    height: '38px',
                    padding: '0 1.25rem',
                    fontSize: '0.85rem',
                    borderRadius: '8px',
                    backgroundColor: 'var(--bg-input)',
                    border: '1px solid var(--border-color)',
                    color: 'var(--text-primary)',
                    fontWeight: 600,
                    transition: 'all 0.2s ease',
                    whiteSpace: 'nowrap'
                  }}
                >
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path><polyline points="15 3 21 3 21 9"></polyline><line x1="10" y1="14" x2="21" y2="3"></line></svg>
                  Checkout URL
                </a>
              )}
              <button
                className="btn-danger"
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '6px',
                  height: '38px',
                  padding: '0 1.25rem',
                  fontSize: '0.85rem',
                  borderRadius: '8px',
                  fontWeight: 600,
                  backgroundColor: 'var(--status-danger-bg)',
                  color: 'var(--status-danger)',
                  cursor: 'pointer',
                  transition: 'all 0.2s ease',
                  border: 'none',
                  whiteSpace: 'nowrap'
                }}
                onClick={(e) => {
                  e.stopPropagation();
                  handleDeleteCheckout(selectedCart.id);
                }}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>
                <span>Delete</span>
              </button>
              <button
                className="btn-primary"
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '6px',
                  height: '38px',
                  padding: '0 1.25rem',
                  fontSize: '0.85rem',
                  borderRadius: '8px',
                  fontWeight: 600,
                  backgroundColor: selectedCart.completed ? 'var(--bg-hover)' : 'var(--accent-color)',
                  color: selectedCart.completed ? 'var(--text-tertiary)' : 'white',
                  cursor: selectedCart.completed || sendingId === selectedCart.id ? 'not-allowed' : 'pointer',
                  transition: 'all 0.2s ease',
                  border: 'none',
                  whiteSpace: 'nowrap'
                }}
                disabled={selectedCart.completed || sendingId === selectedCart.id}
                onClick={(e) => {
                  e.stopPropagation();
                  handleSendRecovery(selectedCart.id);
                }}
              >
                {sendingId === selectedCart.id ? (
                  <>
                    <div className="dot-flashing" style={{ width: '4px', height: '4px' }}></div>
                    <span>Sending...</span>
                  </>
                ) : (
                  <>
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><line x1="22" y1="2" x2="11" y2="13"></line><polygon points="22 2 15 22 11 13 2 9 22 2"></polygon></svg>
                    <span>Recover (WhatsApp)</span>
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
