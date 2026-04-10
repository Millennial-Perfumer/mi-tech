import { API_BASE } from './api';
import { useState, useEffect } from 'react';
import { CustomDatePicker } from './CustomDatePicker';
import './AutomationMessages.css';

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
  refreshTrigger?: number;
}

export function AutomationMessages({ fetchWithAuth, startDate, endDate, onDateChange, refreshTrigger }: AutomationMessagesProps) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(!messages.length);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [page, setPage] = useState(1);
  const [totalCount, setTotalCount] = useState(0);
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [activeTemplates, setActiveTemplates] = useState<string[]>([]);
  const [selectedTemplate, setSelectedTemplate] = useState('');
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
    if (!silent) {
      if (messages.length > 0) setIsRefreshing(true);
      else setIsLoading(true);
    }
    try {
      const templateParam = selectedTemplate ? `&template_name=${encodeURIComponent(selectedTemplate)}` : '';
      const resp = await fetchWithAuth(`${API_BASE}/api/automation/whatsapp/messages?start_date=${startDate}&end_date=${endDate}&page=${page}&limit=${limit}&search=${encodeURIComponent(debouncedSearch)}${templateParam}`);
      const data = await resp.json();
      setMessages(data.messages || []);
      setTotalCount(data.total_count || 0);
      if (data.active_templates) {
        setActiveTemplates(data.active_templates);
      }
    } catch (err) {
      console.error('Failed to fetch messages:', err);
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  };

  useEffect(() => {
    fetchMessages();
    // Auto-refresh every 10 seconds for real-time status updates
    const interval = setInterval(() => fetchMessages(true), 10000);
    return () => clearInterval(interval);
  }, [startDate, endDate, page, debouncedSearch, selectedTemplate, refreshTrigger]);

  const getStatusBadge = (status: string) => {
    switch (status.toLowerCase()) {
      case 'read': 
        return <span className="badge-pill badge-pill-info"><span className="dot" style={{ backgroundColor: '#10b981' }}></span>READ</span>;
      case 'delivered': 
        return <span className="badge-pill badge-pill-success"><span className="dot"></span>DELIVERED</span>;
      case 'sent': 
        return <span className="badge-pill badge-pill-gray"><span className="dot"></span>SENT</span>;
      case 'failed': 
        return <span className="badge-pill badge-pill-danger"><span className="dot"></span>FAILED</span>;
      case 'accepted':
        return <span className="badge-pill badge-pill-info"><span className="dot" style={{ backgroundColor: '#38bdf8' }}></span>ACCEPTED</span>;
      default: 
        return <span className="badge-pill badge-pill-gray"><span className="dot"></span>{status.toUpperCase()}</span>;
    }
  };

  const totalPages = Math.ceil(totalCount / limit);

  const formatPhoneNumber = (num: string) => {
    if (!num) return '';
    const digits = num.replace(/\D/g, '');
    let normalized = digits;
    if (digits.length === 10) normalized = '91' + digits;
    else if (!digits.startsWith('91')) normalized = '91' + digits;
    return `+${normalized.slice(0, 2)} ${normalized.slice(2)}`;
  };

  return (
    <div className="automation-page">
      <div className="messages-header">
        <h2 className="messages-title">Message Delivery Logs</h2>
        {isRefreshing && (
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--accent-color)', fontSize: '0.85rem', fontWeight: 600 }}>
            <div className="dot-flashing"></div>
            Updating Logs...
          </div>
        )}
      </div>

      <div className="filter-bar">
        <div className="search-wrapper">
          <input
            type="text"
            className="search-input"
            style={{ paddingLeft: '3.5rem' }}
            placeholder="Search by Order ID or Customer Name..."
            aria-label="Search message logs"
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
              aria-label="Clear search"
              title="Clear search"
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
            value={selectedTemplate}
            onChange={(e) => {
              setSelectedTemplate(e.target.value);
              setPage(1);
            }}
          >
            <option value="">All Templates</option>
            {activeTemplates.map(name => (
              <option key={name} value={name}>{name}</option>
            ))}
          </select>
          <div style={{ position: 'absolute', right: '1rem', top: '50%', transform: 'translateY(-50%)', pointerEvents: 'none', color: 'var(--text-tertiary)' }}>
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
              <path d="m6 9 6 6 6-6"/>
            </svg>
          </div>
        </div>

        <CustomDatePicker 
          startDate={startDate}
          endDate={endDate}
          onDateChange={onDateChange}
        />

        <div className="refresh-group">
          <div className="auto-refresh-tag">
            <div className="status-dot"></div>
            Live
          </div>
          <button className="refresh-btn" onClick={() => fetchMessages()}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <path d="M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8"></path>
              <path d="M21 3v5h-5"></path>
              <path d="M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16"></path>
              <path d="M3 21v-5h5"></path>
            </svg>
            Refresh
          </button>
        </div>
      </div>

      <div className="table-container shadow-sm">
        <table className="logs-table">
          <thead>
            <tr>
              <th>Time</th>
              <th>Order Details</th>
              <th>Destination</th>
              <th>Template</th>
              <th>Delivery Status</th>
              <th>Delivered At</th>
              <th>Read At</th>
            </tr>
          </thead>
          <tbody>
            {isLoading && messages.length === 0 ? (
              <tr><td colSpan={7} style={{ textAlign: 'center', padding: '4rem' }}>
                <div className="dot-flashing" style={{ margin: '0 auto 1rem' }}></div>
                Loading message logs...
              </td></tr>
            ) : messages.length === 0 ? (
              <tr>
                <td colSpan={7} style={{ textAlign: 'center', padding: '5rem 2rem', color: 'var(--text-secondary)' }}>
                  <div style={{ marginBottom: '1.25rem', opacity: 0.5 }}>
                    <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                      <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
                    </svg>
                  </div>
                  <div style={{ fontWeight: 600, fontSize: '1.1rem', color: 'var(--text-primary)', marginBottom: '0.5rem' }}>
                    No message logs found
                  </div>
                  <p style={{ maxWidth: '320px', margin: '0 auto', fontSize: '0.9rem', lineHeight: 1.5 }}>
                    No message logs match your current criteria. Try adjusting your filters or search query.
                  </p>
                </td>
              </tr>
            ) : (
              messages.map(m => (
                <tr key={m.id}>
                  <td className="time-cell">
                    {new Date(m.sent_at).toLocaleString([], { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' })}
                  </td>
                  <td>
                    <div className="customer-info">
                      <span className={`order-badge ${!m.order_id && m.template_name?.includes('verification') ? 'system-badge' : ''}`}>
                        {m.order_number ? 
                          (String(m.order_number).startsWith('#') ? m.order_number : `#${m.order_number}`) : 
                          (m.order_id ? (String(m.order_id).startsWith('#') ? m.order_id : `#${m.order_id}`) : 
                           (m.template_name?.includes('verification') ? 'System/Auth' : 'Bulk/Test'))}
                      </span>
                      {m.customer_name && <span style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', marginTop: '4px' }}>{m.customer_name}</span>}
                    </div>
                  </td>
                  <td style={{ fontWeight: 500 }}>{formatPhoneNumber(m.phone_number)}</td>
                  <td><span className="template-code">{m.template_name || 'Deleted Template'}</span></td>
                  <td>{getStatusBadge(m.status)}</td>
                  <td className="time-cell">
                    {m.delivered_at ? new Date(m.delivered_at).toLocaleString([], { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' }) : '-'}
                  </td>
                  <td className="time-cell">
                    {m.read_at ? new Date(m.read_at).toLocaleString([], { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' }) : '-'}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {totalCount > 0 && (
        <div className="pagination-container">
          <div className="pagination-info">
            Showing logs <strong>{(page - 1) * limit + 1}</strong> – <strong>{Math.min(page * limit, totalCount)}</strong> of <strong>{totalCount}</strong>
          </div>
          <div className="pagination-controls">
            <button 
              className="page-btn" 
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
            >
              Previous
            </button>
            <span className="current-page">
              Page {page} of {totalPages || 1}
            </span>
            <button 
              className="page-btn" 
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
