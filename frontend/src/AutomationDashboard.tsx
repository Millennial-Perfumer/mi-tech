import { useState, useEffect } from 'react';
import { CustomDatePicker } from './CustomDatePicker';

interface Metrics {
  sent: number;
  delivered: number;
  read: number;
  read_rate: number;
  triggered: number;
  failed: number;
}

interface Activity {
  id: number;
  sent_at: string;
  template_name: string;
  phone_number: string;
  status: string;
  order_id: string | number;
  order_number: string;
  customer_name: string;
}

interface AutomationDashboardProps {
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
  startDate: string;
  endDate: string;
  onDateChange: (start: string, end: string) => void;
}

export function AutomationDashboard({ fetchWithAuth, startDate, endDate, onDateChange }: AutomationDashboardProps) {
  const [metrics, setMetrics] = useState<Metrics | null>(null);
  const [activities, setActivities] = useState<Activity[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSyncing, setIsSyncing] = useState(false);

  useEffect(() => {
    const fetchData = async (silent = false) => {
      if (!silent) setIsLoading(true);
      try {
        const queryParams = `?start_date=${startDate}&end_date=${endDate}`;
        const [metricsResp, messagesResp] = await Promise.all([
          fetchWithAuth(`http://localhost:8080/api/automation/whatsapp/metrics${queryParams}`),
          fetchWithAuth(`http://localhost:8080/api/automation/whatsapp/messages${queryParams}`)
        ]);
        
        const mData = await metricsResp.json();
        const aData = await messagesResp.json();
        
        setMetrics(mData);
        setActivities((aData.messages || []).slice(0, 10)); // Top 10 recent
      } catch (err) {
        console.error('Failed to fetch dashboard data:', err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
    const interval = setInterval(() => {
      if (document.visibilityState === 'visible') {
        fetchData(true);
      }
    }, 15000); // 15 seconds for dashboard metrics
    return () => clearInterval(interval);
  }, [startDate, endDate]);

  const handleSync = async () => {
    setIsSyncing(true);
    try {
      const queryParams = `?start_date=${startDate}&end_date=${endDate}`;
      const resp = await fetchWithAuth(`http://localhost:8080/api/automation/whatsapp/sync-metrics${queryParams}`);
      if (resp.ok) {
        const mData = await resp.json();
        setMetrics(mData);
      } else {
        console.error('Failed to sync metrics from Meta');
      }
    } catch (err) {
      console.error('Error during metrics sync:', err);
    } finally {
      setIsSyncing(false);
    }
  };

  if (isLoading) return <div style={{ padding: '2rem', textAlign: 'center' }}>Loading dashboard...</div>;

  return (
    <div className="automation-dashboard">
      <div style={{ 
        display: 'flex', 
        justifyContent: 'flex-end', 
        alignItems: 'center',
        gap: '1rem',
        marginBottom: '2rem' 
      }}>
        <button 
          className="btn-secondary" 
          onClick={handleSync}
          disabled={isSyncing}
          style={{ 
            display: 'flex', 
            alignItems: 'center', 
            gap: '0.5rem',
            padding: '0.5rem 1rem',
            fontSize: '0.85rem',
            height: '42px', // Match date picker height
            opacity: isSyncing ? 0.7 : 1
          }}
        >
          <svg 
            width="16" 
            height="16" 
            viewBox="0 0 24 24" 
            fill="none" 
            stroke="currentColor" 
            strokeWidth="2.5" 
            strokeLinecap="round" 
            strokeLinejoin="round"
            style={{ animation: isSyncing ? 'spin 1.5s linear infinite' : 'none' }}
          >
            <path d="M21 2v6h-6"></path>
            <path d="M3 12a9 9 0 0 1 15-6.7L21 8"></path>
            <path d="M3 22v-6h6"></path>
            <path d="M21 12a9 9 0 0 1-15 6.7L3 16"></path>
          </svg>
          {isSyncing ? 'Syncing...' : 'Sync Metrics'}
        </button>

        <CustomDatePicker 
          startDate={startDate}
          endDate={endDate}
          onDateChange={onDateChange}
        />
      </div>

      <style>{`
        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
      `}</style>
      {/* KPI Grid */}
      <div className="metrics-grid" style={{ 
        display: 'grid', 
        gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', 
        gap: '1.5rem',
        marginBottom: '2rem'
      }}>
        <MetricCard title="Messages Sent" value={metrics?.sent || 0} />
        <MetricCard title="Messages Delivered" value={metrics?.delivered || 0} />
        <MetricCard title="Messages Read" value={metrics?.read || 0} />
        <MetricCard title="Read Rate" value={`${(metrics?.read_rate || 0).toFixed(1)}%`} />
        <MetricCard title="Triggers Executed" value={metrics?.triggered || 0} />
        <MetricCard title="Failed Messages" value={metrics?.failed || 0} color="#ef4444" />
      </div>

      {/* Recent Activity */}
      <div className="card">
        <h3 style={{ fontSize: '1.1rem', fontWeight: 600, marginBottom: '1.25rem' }}>Recent Automation Activity</h3>
        <div className="table-container">
          <table>
            <thead>
              <tr>
                <th>Time</th>
                <th>Order ID</th>
                <th>Template</th>
                <th>Recipient</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {activities.length === 0 ? (
                <tr><td colSpan={5} style={{ textAlign: 'center', padding: '2rem' }}>No recent activity.</td></tr>
              ) : (
                activities.map(a => (
                  <tr key={a.id}>
                    <td>{new Date(a.sent_at).toLocaleString([], { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' })}</td>
                    <td>
                      <div style={{ fontWeight: 600 }}>
                        {a.order_number ? 
                          (String(a.order_number).startsWith('#') ? a.order_number : `#${a.order_number}`) : 
                          (a.order_id ? (String(a.order_id).startsWith('#') ? a.order_id : `#${a.order_id}`) : '-')}
                      </div>
                      {a.customer_name && <div style={{ fontSize: '0.75rem', color: '#64748b' }}>{a.customer_name}</div>}
                    </td>
                    <td><code>{a.template_name}</code></td>
                    <td>{formatPhoneNumber(a.phone_number)}</td>
                    <td>
                      <span className={`badge ${
                        a.status === 'read' ? 'badge-success' : 
                        a.status === 'delivered' ? 'badge-info' : 
                        a.status === 'failed' ? 'badge-danger' : 'badge-warning'
                      }`}>
                        {a.status.toUpperCase()}
                      </span>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

function MetricCard({ title, value, color }: { title: string, value: string | number, color?: string }) {
  return (
    <div className="card" style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
      <span style={{ fontSize: '0.875rem', color: 'var(--text-secondary)' }}>{title}</span>
      <span style={{ fontSize: '1.5rem', fontWeight: 700, color: color || 'var(--text-primary)' }}>{value}</span>
    </div>
  );
}

function formatPhoneNumber(phone: string): string {
  if (!phone) return '-';
  // Remove non-digit characters
  const clean = phone.replace(/\D/g, '');
  
  // Handle 10-digit Indian numbers
  if (clean.length === 10) {
    return `+91 ${clean}`;
  }
  
  // Handle 12-digit numbers starting with 91
  if (clean.length === 12 && clean.startsWith('91')) {
    return `+${clean}`;
  }

  return phone.startsWith('+') ? phone : `+${phone}`;
}
