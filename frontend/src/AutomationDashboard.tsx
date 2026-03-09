import { useState, useEffect } from 'react';

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
  order_id: string;
}

export function AutomationDashboard() {
  const [metrics, setMetrics] = useState<Metrics | null>(null);
  const [activities, setActivities] = useState<Activity[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [metricsResp, messagesResp] = await Promise.all([
          fetch('http://localhost:8080/api/automation/whatsapp/metrics'),
          fetch('http://localhost:8080/api/automation/whatsapp/messages')
        ]);
        
        const mData = await metricsResp.json();
        const aData = await messagesResp.json();
        
        setMetrics(mData);
        setActivities(aData.slice(0, 10)); // Top 10 recent
      } catch (err) {
        console.error('Failed to fetch dashboard data:', err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, []);

  if (isLoading) return <div style={{ padding: '2rem', textAlign: 'center' }}>Loading dashboard...</div>;

  return (
    <div className="automation-dashboard">
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
                <th>Event/Order</th>
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
                    <td>{new Date(a.sent_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</td>
                    <td>{a.order_id ? `#${a.order_id}` : 'Webhook'}</td>
                    <td><code>{a.template_name}</code></td>
                    <td>{a.phone_number.replace(/(\d{2})(\d{4})(\d{4})/, '+$1 xxxx $3')}</td>
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
