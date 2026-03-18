import { API_BASE } from './api';
import { useState, useEffect } from 'react';
import { CustomDatePicker } from './CustomDatePicker';

interface Message {
  id: number;
  order_id: string | number;
  order_number: string;
  customer_name: string;
  phone_number: string;
  template_name: string;
  status: string;
  sent_at: string;
  delivered_at: string | null;
  read_at: string | null;
  error_message: string;
}

interface AutomationMessagesProps {
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
  startDate: string;
  endDate: string;
  onDateChange: (start: string, end: string) => void;
}

export function AutomationMessages({ fetchWithAuth, startDate, endDate, onDateChange }: AutomationMessagesProps) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [totalCount, setTotalCount] = useState(0);
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const limit = 25;

  // Debounce search query
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchQuery);
      setPage(1); // Reset to first page on new search
    }, 500);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  const fetchMessages = async (silent = false) => {
    if (!silent) setIsLoading(true);
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/automation/whatsapp/messages?start_date=${startDate}&end_date=${endDate}&page=${page}&limit=${limit}&search=${encodeURIComponent(debouncedSearch)}`);
      const data = await resp.json();
      setMessages(data.messages || []);
      setTotalCount(data.total_count || 0);
    } catch (err) {
      console.error('Failed to fetch messages:', err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchMessages();
    // Auto-refresh every 10 seconds for real-time status updates
    const interval = setInterval(() => fetchMessages(true), 10000);
    return () => clearInterval(interval);
  }, [startDate, endDate, page, debouncedSearch]);

  const getStatusBadge = (status: string) => {
    switch (status.toLowerCase()) {
      case 'read': return <span className="badge" style={{ backgroundColor: '#0ea5e9', color: 'white' }}>READ</span>;
      case 'delivered': return <span className="badge badge-success">DELIVERED</span>;
      case 'sent': return <span className="badge" style={{ backgroundColor: '#64748b', color: 'white' }}>SENT</span>;
      case 'failed': return <span className="badge badge-danger">FAILED</span>;
      default: return <span className="badge">{status.toUpperCase()}</span>;
    }
  };

  const totalPages = Math.ceil(totalCount / limit);

  const formatPhoneNumber = (num: string) => {
    if (!num) return '';
    // Remove any non-digit characters
    const digits = num.replace(/\D/g, '');
    let normalized = digits;
    
    if (digits.length === 10) {
      normalized = '91' + digits;
    } else if (!digits.startsWith('91')) {
      normalized = '91' + digits;
    }
    
    // Return formatted as +91 XXXXXXXXXX
    return `+${normalized.slice(0, 2)} ${normalized.slice(2)}`;
  };

  return (
    <div className="automation-page">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <h2 style={{ fontSize: '1.25rem', fontWeight: 600 }}>Message Delivery Logs</h2>
        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <div className="search-input-wrapper" style={{ position: 'relative' }}>
            <input
              type="text"
              placeholder="Search order"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              style={{
                padding: '0.4rem 0.8rem 0.4rem 2rem',
                fontSize: '0.85rem',
                border: '1px solid var(--border-color)',
                borderRadius: '6px',
                width: '240px',
                height: '42px'
              }}
            />
            <svg 
              width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
              style={{ position: 'absolute', left: '0.75rem', top: '50%', transform: 'translateY(-50%)', color: '#64748b' }}
            >
              <circle cx="11" cy="11" r="8"></circle>
              <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
            </svg>
            {searchQuery && (
              <button
                onClick={() => setSearchQuery('')}
                style={{
                  position: 'absolute', right: '0.5rem', top: '50%', transform: 'translateY(-50%)',
                  background: 'none', border: 'none', cursor: 'pointer', color: '#64748b', fontSize: '1.2rem',
                  display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '4px'
                }}
              >
                &times;
              </button>
            )}
          </div>
          <span style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>Auto-refreshing every 10s</span>
          <button className="btn-secondary" onClick={() => fetchMessages()} style={{ height: '42px' }}>Refresh Now</button>
          <CustomDatePicker 
            startDate={startDate}
            endDate={endDate}
            onDateChange={onDateChange}
          />
        </div>
      </div>

      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>Time</th>
              <th>Order ID</th>
              <th>Customer Phone</th>
              <th>Template</th>
              <th>Status</th>
              <th>Delivered</th>
              <th>Read</th>
            </tr>
          </thead>
          <tbody>
            {isLoading && messages.length === 0 ? (
              <tr><td colSpan={7} style={{ textAlign: 'center', padding: '2rem' }}>Loading logs...</td></tr>
            ) : messages.length === 0 ? (
              <tr><td colSpan={7} style={{ textAlign: 'center', padding: '2rem' }}>No messages sent yet.</td></tr>
            ) : (
              messages.map(m => (
                <tr key={m.id}>
                  <td style={{ fontSize: '0.85rem' }}>
                    {new Date(m.sent_at).toLocaleString([], { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' })}
                  </td>
                  <td>
                    <div style={{ fontWeight: 600 }}>
                      {m.order_number ? 
                        (String(m.order_number).startsWith('#') ? m.order_number : `#${m.order_number}`) : 
                        (m.order_id ? (String(m.order_id).startsWith('#') ? m.order_id : `#${m.order_id}`) : 'Bulk/Test')}
                    </div>
                    {m.customer_name && <div style={{ fontSize: '0.75rem', color: '#64748b' }}>{m.customer_name}</div>}
                  </td>
                  <td>{formatPhoneNumber(m.phone_number)}</td>
                  <td><code>{m.template_name || 'Deleted Template'}</code></td>
                  <td>{getStatusBadge(m.status)}</td>
                  <td style={{ fontSize: '0.85rem' }}>
                    {m.delivered_at ? new Date(m.delivered_at).toLocaleString([], { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' }) : '-'}
                  </td>
                  <td style={{ fontSize: '0.85rem' }}>
                    {m.read_at ? new Date(m.read_at).toLocaleString([], { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' }) : '-'}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {totalCount > 0 && (
        <div style={{ 
          display: 'flex', 
          justifyContent: 'space-between', 
          alignItems: 'center', 
          marginTop: '1.5rem',
          padding: '1rem 0',
          borderTop: '1px solid #f1f5f9'
        }}>
          <div style={{ fontSize: '0.85rem', color: '#64748b' }}>
            Showing <strong>{(page - 1) * limit + 1}</strong> to <strong>{Math.min(page * limit, totalCount)}</strong> of <strong>{totalCount}</strong> logs
          </div>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <button 
              className="btn-secondary" 
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
              style={{ padding: '0.4rem 0.8rem', fontSize: '0.85rem' }}
            >
              Previous
            </button>
            <div style={{ display: 'flex', alignItems: 'center', padding: '0 0.75rem', fontSize: '0.85rem', color: '#64748b' }}>
              Page {page} of {totalPages || 1}
            </div>
            <button 
              className="btn-secondary" 
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
              style={{ padding: '0.4rem 0.8rem', fontSize: '0.85rem' }}
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
