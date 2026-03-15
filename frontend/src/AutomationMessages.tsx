import { useState, useEffect } from 'react';

interface Message {
  id: number;
  order_id: string;
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
}

export function AutomationMessages({ fetchWithAuth, startDate, endDate }: AutomationMessagesProps) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const fetchMessages = async (silent = false) => {
    if (!silent) setIsLoading(true);
    try {
      const resp = await fetchWithAuth(`http://localhost:8080/api/automation/whatsapp/messages?start_date=${startDate}&end_date=${endDate}`);
      const data = await resp.json();
      setMessages(data || []);
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
  }, [startDate, endDate]);

  const getStatusBadge = (status: string) => {
    switch (status.toLowerCase()) {
      case 'read': return <span className="badge" style={{ backgroundColor: '#0ea5e9', color: 'white' }}>READ</span>;
      case 'delivered': return <span className="badge badge-success">DELIVERED</span>;
      case 'sent': return <span className="badge" style={{ backgroundColor: '#64748b', color: 'white' }}>SENT</span>;
      case 'failed': return <span className="badge badge-danger">FAILED</span>;
      default: return <span className="badge">{status.toUpperCase()}</span>;
    }
  };

  return (
    <div className="automation-page">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <h2 style={{ fontSize: '1.25rem', fontWeight: 600 }}>Message Delivery Logs</h2>
        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <span style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>Auto-refreshing every 10s</span>
          <button className="btn-secondary" onClick={() => fetchMessages()}>Refresh Now</button>
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
            {isLoading ? (
              <tr><td colSpan={7} style={{ textAlign: 'center', padding: '2rem' }}>Loading logs...</td></tr>
            ) : messages.length === 0 ? (
              <tr><td colSpan={7} style={{ textAlign: 'center', padding: '2rem' }}>No messages sent yet.</td></tr>
            ) : (
              messages.map(m => (
                <tr key={m.id}>
                  <td style={{ fontSize: '0.85rem' }}>
                    {new Date(m.sent_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                  </td>
                  <td>{m.order_id ? `#${m.order_id}` : 'Bulk/Test'}</td>
                  <td>{m.phone_number.replace(/(\d{2})(\d{4})(\d{4})/, '+$1 xxxx $3')}</td>
                  <td><code>{m.template_name || 'Deleted Template'}</code></td>
                  <td>{getStatusBadge(m.status)}</td>
                  <td style={{ fontSize: '0.85rem' }}>
                    {m.delivered_at ? new Date(m.delivered_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : '-'}
                  </td>
                  <td style={{ fontSize: '0.85rem' }}>
                    {m.read_at ? new Date(m.read_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : '-'}
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
