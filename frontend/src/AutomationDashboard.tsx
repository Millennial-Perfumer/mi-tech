import { API_BASE } from './api';
import { useState, useEffect } from 'react';
import { useToast } from './ToastContext';
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
  refreshTrigger?: number;
}

export function AutomationDashboard({ fetchWithAuth, startDate, endDate, onDateChange, refreshTrigger }: AutomationDashboardProps) {
  const { success: toastSuccess, error: toastError } = useToast();
  const [metrics, setMetrics] = useState<Metrics | null>(null);
  const [activities, setActivities] = useState<Activity[]>([]);
  const [isLoading, setIsLoading] = useState(!metrics);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isSyncing, setIsSyncing] = useState(false);

  useEffect(() => {
    const fetchData = async (silent = false) => {
      if (!silent) {
        if (metrics) setIsRefreshing(true);
        else setIsLoading(true);
      }
      try {
        const queryParams = `?start_date=${startDate}&end_date=${endDate}`;
        const [metricsResp, messagesResp] = await Promise.all([
          fetchWithAuth(`${API_BASE}/api/automation/whatsapp/metrics${queryParams}`),
          fetchWithAuth(`${API_BASE}/api/automation/whatsapp/messages${queryParams}`)
        ]);
        
        const mData = await metricsResp.json();
        const aData = await messagesResp.json();
        
        setMetrics(mData);
        setActivities((aData.messages || []).slice(0, 10)); // Top 10 recent
      } catch (err) {
        console.error('Failed to fetch dashboard data:', err);
      } finally {
        setIsLoading(false);
        setIsRefreshing(false);
      }
    };

    fetchData();
    const interval = setInterval(() => {
      if (document.visibilityState === 'visible') {
        fetchData(true);
      }
    }, 15000); // 15 seconds for dashboard metrics
    return () => clearInterval(interval);
  }, [startDate, endDate, refreshTrigger]);

  const handleSync = async () => {
    setIsSyncing(true);
    try {
      const queryParams = `?start_date=${startDate}&end_date=${endDate}`;
      const resp = await fetchWithAuth(`${API_BASE}/api/automation/whatsapp/sync-metrics${queryParams}`);
      if (resp.ok) {
        const mData = await resp.json();
        setMetrics(mData);
        toastSuccess('Metrics synced successfully from Meta');
      } else {
        console.error('Failed to sync metrics from Meta');
        toastError('Failed to sync metrics from Meta');
      }
    } catch (err) {
      console.error('Error during metrics sync:', err);
      toastError('Network error while syncing metrics');
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
        {isRefreshing && (
          <div style={{ marginRight: 'auto', display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--accent-color)', fontSize: '0.85rem', fontWeight: 600 }}>
             <div className="dot-flashing"></div>
             Updating...
          </div>
        )}
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
      <div className="metrics-grid">
        <MetricCard title="Messages Sent" value={metrics?.sent || 0} iconClass="metric-icon-1" icon={<svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="22" y1="2" x2="11" y2="13"></line><polygon points="22 2 15 22 11 13 2 9 22 2"></polygon></svg>} />
        <MetricCard title="Messages Delivered" value={metrics?.delivered || 0} iconClass="metric-icon-2" icon={<svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>} />
        <MetricCard title="Messages Read" value={metrics?.read || 0} iconClass="metric-icon-3" icon={<svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path><circle cx="12" cy="12" r="3"></circle></svg>} />
        <MetricCard title="Read Rate" value={`${(metrics?.read_rate || 0).toFixed(1)}%`} iconClass="metric-icon-1" icon={<svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="19" y1="5" x2="5" y2="19"></line><circle cx="6.5" cy="6.5" r="2.5"></circle><circle cx="17.5" cy="17.5" r="2.5"></circle></svg>} />
        <MetricCard title="Triggers Executed" value={metrics?.triggered || 0} iconClass="metric-icon-5" icon={<svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon></svg>} />
        <MetricCard title="Failed Messages" value={metrics?.failed || 0} color="#ef4444" iconClass="metric-icon-4" icon={<svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>} />
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
                      {a.customer_name && <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>{a.customer_name}</div>}
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

function MetricCard({ title, value, color, icon, iconClass }: { title: string, value: string | number, color?: string, icon?: React.ReactNode, iconClass?: string }) {
  return (
    <div className="metric-card">
      {icon && <div className={`metric-icon ${iconClass || 'metric-icon-1'}`}>{icon}</div>}
      <div className="metric-label">{title}</div>
      <div className="metric-value" style={{ color: color || 'var(--text-primary)' }}>{value}</div>
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
