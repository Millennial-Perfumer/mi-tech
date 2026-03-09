import { useState, useEffect } from 'react';
import { CustomDatePicker } from './CustomDatePicker';
import { ColumnSelector } from './ColumnSelector';
import type { ColumnOption } from './ColumnSelector';
import { GSTReports } from './GSTReports';
import { WhatsAppAutomation } from './WhatsAppAutomation';
import './App.css';

interface Order {
  id: string;
  order_number: string;
  total_price: string;
  created_at: string;
  customer_name: string;
  customer_city: string;
  customer_state: string;
  customer_country: string;
  financial_status: string;
  fulfillment_status: string;
  status: string;
  shopify_order_id: string;
}

interface WebhookStatus {
  topic: string;
  status: string;
  last_received: string;
}

interface DashboardMetrics {
  total_revenue: number;
  total_invoices: number;
  total_gst_collected: number;
  cgst_collected: number;
  sgst_collected: number;
  igst_collected: number;
  total_orders: number;
  cancelled_orders: number;
  fulfilled_orders: number;
  unfulfilled_orders: number;
}

const AVAILABLE_COLUMNS: (ColumnOption & { isDefault: boolean })[] = [
  { id: 'order_id', label: 'Order ID', category: 'General', isDefault: true },
  { id: 'customer_name', label: 'Customer Name', category: 'Customer', isDefault: true },
  { id: 'city', label: 'City', category: 'Location', isDefault: false },
  { id: 'state', label: 'State', category: 'Location', isDefault: true },
  { id: 'country', label: 'Country', category: 'Location', isDefault: false },
  { id: 'date', label: 'Date', category: 'Date', isDefault: true },
  { id: 'time', label: 'Time', category: 'Date', isDefault: false },
  { id: 'amount', label: 'Amount', category: 'General', isDefault: true },
  { id: 'financial_status', label: 'Payment', category: 'Status', isDefault: true },
  { id: 'fulfillment_status', label: 'Fulfillment', category: 'Status', isDefault: true },
  { id: 'source', label: 'Source', category: 'General', isDefault: true },
  { id: 'gst_invoice', label: 'GST Invoice', category: 'General', isDefault: true },
];

const DEFAULT_VISIBLE_COLUMNS = AVAILABLE_COLUMNS.filter(c => c.isDefault).map(c => c.id);

function App() {
  const [activeTab, setActiveTab] = useState('dashboard');
  const [orders, setOrders] = useState<Order[]>([]);
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null);
  const [isSyncing, setIsSyncing] = useState(false);
  const [isResetting, setIsResetting] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [totalCount, setTotalCount] = useState(0);
  const [webhookStatus, setWebhookStatus] = useState<WebhookStatus | null>(null);
  const limit = 25;

  // Column Selector State
  const [visibleColumns, setVisibleColumns] = useState<string[]>(() => {
    const saved = localStorage.getItem('shopifyAppVisibleColumns');
    if (saved) {
      try {
        return JSON.parse(saved);
      } catch {
        return DEFAULT_VISIBLE_COLUMNS;
      }
    }
    return DEFAULT_VISIBLE_COLUMNS;
  });

  useEffect(() => {
    localStorage.setItem('shopifyAppVisibleColumns', JSON.stringify(visibleColumns));
  }, [visibleColumns]);

  // Default to Month-to-Date (MTD)
  const defaultStartDate = new Date(new Date().getFullYear(), new Date().getMonth(), 1).toISOString().split('T')[0];
  const defaultEndDate = new Date().toISOString().split('T')[0];
  const [startDate, setStartDate] = useState(defaultStartDate);
  const [endDate, setEndDate] = useState(defaultEndDate);

  const fetchDashboardData = async () => {
    setIsLoading(true);
    try {
      let startObj = '';
      let endObj = '';
      
      if (startDate) {
        const [y, m, d] = startDate.split('-').map(Number);
        startObj = new Date(y, m - 1, d, 0, 0, 0, 0).toISOString();
      }
      
      if (endDate) {
        const [y, m, d] = endDate.split('-').map(Number);
        endObj = new Date(y, m - 1, d, 23, 59, 59, 999).toISOString();
      }
      
      const metricsRes = await fetch(`http://localhost:8080/api/dashboard/metrics?start_date=${startObj}&end_date=${endObj}`);
      const metricsData = await metricsRes.json();
      
      if (metricsData.success) {
        setMetrics(metricsData.metrics);
      }

      const ordersRes = await fetch(`http://localhost:8080/api/orders?start_date=${startObj}&end_date=${endObj}&page=${page}&limit=${limit}`);
      const ordersData = await ordersRes.json();
      if (ordersData.success) {
        setOrders(ordersData.orders);
        setTotalCount(ordersData.total_count);
      }

      const webhookRes = await fetch('http://localhost:8080/api/webhook/status');
      const webhookData = await webhookRes.json();
      setWebhookStatus(webhookData);
    } catch (error) {
      console.error('Error fetching data:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const syncShopify = async () => {
    setIsSyncing(true);
    try {
      const response = await fetch('http://localhost:8080/api/shopify/sync', {
        method: 'POST',
      });
      const data = await response.json();
      if (data.success) {
        alert(`Successfully synced ${data.count} orders!`);
        fetchDashboardData();
      } else {
        alert('Failed to sync orders.');
      }
    } catch (error) {
      console.error('Error syncing orders:', error);
      alert('Error occurred while syncing.');
    } finally {
      setIsSyncing(false);
    }
  };

  const resetShopify = async () => {
    if (!window.confirm("Are you sure you want to delete all historical synced data and force a full re-sync from March 2026? This cannot be undone.")) {
      return;
    }
    setIsResetting(true);
    try {
      const response = await fetch('http://localhost:8080/api/shopify/reset', {
        method: 'POST',
      });
      const data = await response.json();
      if (data.success) {
        alert(`Successfully wiped data and re-synced ${data.count} orders!`);
        fetchDashboardData();
      } else {
        alert('Failed to reset orders.');
      }
    } catch (error) {
      console.error('Error resetting orders:', error);
      alert('Error occurred while resetting.');
    } finally {
      setIsResetting(false);
    }
  };


  useEffect(() => {
    fetchDashboardData();
  }, [startDate, endDate, page]);

  return (
    <div className="app-container">
      <aside className="sidebar">
        <div className="sidebar-brand">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{color: 'var(--accent-color)'}}>
            <path d="M22 12h-4l-3 9L9 3l-3 9H2"></path>
          </svg>
          GST Invoice
        </div>
        
        <nav className="sidebar-nav">
          <a href="#" className={`nav-item ${activeTab === 'dashboard' ? 'active' : ''}`} onClick={() => setActiveTab('dashboard')}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="3" width="7" height="7"></rect><rect x="14" y="3" width="7" height="7"></rect><rect x="14" y="14" width="7" height="7"></rect><rect x="3" y="14" width="7" height="7"></rect></svg>
            Dashboard
          </a>
          <a href="#" className={`nav-item ${activeTab === 'reports' ? 'active' : ''}`} onClick={() => setActiveTab('reports')}>
             <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line><polyline points="10 9 9 9 8 9"></polyline></svg>
            GST Reports
          </a>
          <a href="#" className={`nav-item ${activeTab === 'invoices' ? 'active' : ''}`} onClick={() => setActiveTab('invoices')}>
             <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
            Invoices
          </a>
          <a href="#" className={`nav-item ${activeTab === 'automation' ? 'active' : ''}`} onClick={() => setActiveTab('automation')}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>
            Automation
          </a>
          <a href="#" className={`nav-item ${activeTab === 'shopify' ? 'active' : ''}`} onClick={() => setActiveTab('shopify')}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="9" cy="21" r="1"></circle><circle cx="20" cy="21" r="1"></circle><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"></path></svg>
            Orders
          </a>
          <a href="#" className={`nav-item ${activeTab === 'settings' ? 'active' : ''}`} onClick={() => setActiveTab('settings')}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>
            Settings
          </a>
        </nav>
      </aside>

      <main className="main-content">
        <header className="page-header">
          <div>
            <h1 className="page-title">{activeTab === 'dashboard' ? 'Overview' : activeTab === 'shopify' ? 'Orders' : activeTab === 'reports' ? 'GST Reports' : activeTab === 'automation' ? 'Automation Hub' : activeTab === 'invoices' ? 'Invoices' : 'Settings'}</h1>
            <p className="page-subtitle">
              {activeTab === 'dashboard' ? "Welcome back. Here's what's happening today." : activeTab === 'reports' ? "Review your GST collection and generate filing reports." : activeTab === 'automation' ? "Manage templates, triggers, and track WhatsApp communication." : activeTab === 'shopify' ? "Real-time orders synced via Shopify Webhooks." : "Manage your store data and preferences."}
            </p>
          </div>
          {activeTab !== 'automation' && (
            <div style={{display: 'flex', gap: '1rem'}}>
              <button className="btn-secondary">Export Data</button>
              <button 
                className="btn-secondary" 
                style={{display: 'flex', alignItems: 'center', gap: '0.5rem', opacity: isResetting ? 0.7 : 1, backgroundColor: '#ef4444', color: 'white', borderColor: '#ef4444'}}
                onClick={resetShopify}
                disabled={isResetting || isSyncing}
              >
                {isResetting ? 'Resetting...' : 'Reset & Resync'}
              </button>
              <button 
                className="btn-primary" 
                title="Manually fetch orders from Shopify in case webhook delivery fails."
                style={{display: 'flex', alignItems: 'center', gap: '0.5rem', opacity: isSyncing ? 0.7 : 1}}
                onClick={syncShopify}
                disabled={isSyncing || isResetting}
              >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={isSyncing ? 'spin' : ''}><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
                {isSyncing ? 'Syncing...' : 'Manual Sync'}
              </button>
            </div>
          )}
        </header>

        {activeTab !== 'automation' && (
          <div style={{ marginBottom: '2rem', display: 'flex', alignItems: 'center' }}>
            <CustomDatePicker 
              startDate={startDate} 
              endDate={endDate} 
              onDateChange={(start, end) => {
                setPage(1);
                setStartDate(start);
                setEndDate(end);
              }} 
            />
          </div>
        )}

        {activeTab === 'dashboard' && (
          <section className="dashboard-grid">
              <div className="card">
                <h3 className="card-title">Total Revenue</h3>
                <div className="card-value">₹{metrics?.total_revenue?.toLocaleString('en-IN') || '0'}</div>
              </div>
              <div className="card">
                <h3 className="card-title">Total Invoices</h3>
                <div className="card-value">{metrics?.total_invoices?.toLocaleString('en-IN') || '0'}</div>
              </div>
              <div className="card">
                <h3 className="card-title">Total GST Collected</h3>
                <div className="card-value">₹{metrics?.total_gst_collected?.toLocaleString('en-IN', { maximumFractionDigits: 2 }) || '0'}</div>
              </div>
              <div className="card" style={{gridColumn: 'span 3'}}>
                <div style={{ display: 'flex', justifyContent: 'space-between', gap: '1rem' }}>
                  <div style={{ flex: 1 }}>
                    <h3 className="card-title" style={{ fontSize: '0.75rem', color: 'var(--accent-color)' }}>CGST Collected</h3>
                    <div style={{ fontSize: '1.5rem', fontWeight: 700 }}>₹{metrics?.cgst_collected?.toLocaleString('en-IN', { maximumFractionDigits: 2 }) || '0'}</div>
                  </div>
                  <div style={{ flex: 1, borderLeft: '1px solid var(--border-color)', paddingLeft: '1.5rem' }}>
                    <h3 className="card-title" style={{ fontSize: '0.75rem', color: 'var(--accent-color)' }}>SGST Collected</h3>
                    <div style={{ fontSize: '1.5rem', fontWeight: 700 }}>₹{metrics?.sgst_collected?.toLocaleString('en-IN', { maximumFractionDigits: 2 }) || '0'}</div>
                  </div>
                  <div style={{ flex: 1, borderLeft: '1px solid var(--border-color)', paddingLeft: '1.5rem' }}>
                    <h3 className="card-title" style={{ fontSize: '0.75rem', color: 'var(--accent-color)' }}>IGST Collected</h3>
                    <div style={{ fontSize: '1.5rem', fontWeight: 700 }}>₹{metrics?.igst_collected?.toLocaleString('en-IN', { maximumFractionDigits: 2 }) || '0'}</div>
                  </div>
                </div>
              </div>
              <div className="card" style={{ borderColor: 'var(--border-color)' }}>
                <h3 className="card-title">Total Orders</h3>
                <div className="card-value">{metrics?.total_orders?.toLocaleString() || '0'}</div>
              </div>
              <div className="card" style={{ borderColor: '#fee2e2' }}>
                <h3 className="card-title" style={{ color: '#991b1b' }}>Cancelled Orders</h3>
                <div className="card-value" style={{ color: '#991b1b' }}>{metrics?.cancelled_orders?.toLocaleString() || '0'}</div>
              </div>
              <div className="card" style={{ borderColor: '#dcfce7' }}>
                <h3 className="card-title" style={{ color: '#166534' }}>Fulfilled Orders</h3>
                <div className="card-value" style={{ color: '#166534' }}>{metrics?.fulfilled_orders?.toLocaleString() || '0'}</div>
              </div>
              <div className="card" style={{ borderColor: '#fef9c3' }}>
                <h3 className="card-title" style={{ color: '#854d0e' }}>Unfulfilled Orders</h3>
                <div className="card-value" style={{ color: '#854d0e' }}>{metrics?.unfulfilled_orders?.toLocaleString() || '0'}</div>
              </div>
          </section>
        )}

        {activeTab === 'shopify' && (
          <section className="table-container">
            <div style={{ display: 'flex', gap: '1rem', marginBottom: '1.5rem', padding: '1rem', backgroundColor: '#f8fafc', borderRadius: '8px', border: '1px solid var(--border-color)' }}>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.025em', marginBottom: '0.25rem' }}>Webhook Status</div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <div style={{ width: '8px', height: '8px', borderRadius: '50%', backgroundColor: webhookStatus?.status === 'active' ? '#10b981' : '#f43f5e' }}></div>
                  <span style={{ fontWeight: 600, fontSize: '0.875rem' }}>{webhookStatus?.status === 'active' ? 'Active' : 'Inactive'}</span>
                </div>
              </div>
              <div style={{ flex: 1, borderLeft: '1px solid var(--border-color)', paddingLeft: '1rem' }}>
                <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.025em', marginBottom: '0.25rem' }}>Last Event</div>
                <div style={{ fontWeight: 600, fontSize: '0.875rem' }}>{webhookStatus?.topic || 'None'}</div>
              </div>
              <div style={{ flex: 1, borderLeft: '1px solid var(--border-color)', paddingLeft: '1rem' }}>
                <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.025em', marginBottom: '0.25rem' }}>Last Received</div>
                <div style={{ fontWeight: 600, fontSize: '0.875rem' }}>
                  {webhookStatus?.last_received ? new Date(webhookStatus.last_received).toLocaleTimeString() : 'N/A'}
                </div>
              </div>
            </div>

            <div className="table-header">
              <h3 style={{fontSize: '1rem'}}>Stored Orders</h3>
              <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                <ColumnSelector columns={AVAILABLE_COLUMNS} visibleColumns={visibleColumns} onChange={setVisibleColumns} />
              </div>
            </div>
            <div style={{overflowX: 'auto'}}>
            <table>
              <thead>
                <tr>
                  {visibleColumns.includes('order_id') && <th>Order ID</th>}
                  {visibleColumns.includes('customer_name') && <th>Customer</th>}
                  {visibleColumns.includes('city') && <th>City</th>}
                  {visibleColumns.includes('state') && <th>State</th>}
                  {visibleColumns.includes('country') && <th>Country</th>}
                  {visibleColumns.includes('date') && <th>Date</th>}
                  {visibleColumns.includes('time') && <th>Time</th>}
                  {visibleColumns.includes('amount') && <th>Amount</th>}
                  {visibleColumns.includes('financial_status') && <th>Payment</th>}
                  {visibleColumns.includes('fulfillment_status') && <th>Fulfillment</th>}
                  {visibleColumns.includes('source') && <th>Source</th>}
                  {visibleColumns.includes('gst_invoice') && <th>GST Invoice</th>}
                </tr>
              </thead>
              <tbody>
                {isLoading ? (
                  <tr>
                    <td colSpan={visibleColumns.length} style={{ textAlign: 'center', padding: '2rem' }}>Loading orders...</td>
                  </tr>
                ) : orders.length === 0 ? (
                  <tr>
                    <td colSpan={visibleColumns.length} style={{ textAlign: 'center', padding: '2rem' }}>No orders found. Click Sync Shopify to fetch.</td>
                  </tr>
                ) : (
                  orders.map((order) => (
                    <tr key={order.id}>
                      {visibleColumns.includes('order_id') && <td><a href="#">{order.order_number}</a></td>}
                      {visibleColumns.includes('customer_name') && <td>{order.customer_name}</td>}
                      {visibleColumns.includes('city') && <td>{order.customer_city || 'N/A'}</td>}
                      {visibleColumns.includes('state') && <td>{order.customer_state || 'N/A'}</td>}
                      {visibleColumns.includes('country') && <td>{order.customer_country || 'N/A'}</td>}
                      {visibleColumns.includes('date') && <td>{new Date(order.created_at).toLocaleDateString()}</td>}
                      {visibleColumns.includes('time') && <td>{new Date(order.created_at).toLocaleTimeString()}</td>}
                      {visibleColumns.includes('amount') && <td>₹{order.total_price}</td>}
                      {visibleColumns.includes('financial_status') && (
                        <td>
                          <span className={`badge ${order.financial_status === 'paid' ? 'badge-success' : 'badge-warning'}`}>
                            {order.financial_status?.toUpperCase() || 'UNKNOWN'}
                          </span>
                        </td>
                      )}
                      {visibleColumns.includes('fulfillment_status') && (
                        <td>
                          <span className={`badge ${order.fulfillment_status === 'fulfilled' ? 'badge-success' : 'badge-warning'}`}>
                            {order.fulfillment_status?.toUpperCase() || 'UNFULFILLED'}
                          </span>
                        </td>
                      )}
                      {visibleColumns.includes('source') && (
                        <td>
                          <span className="badge" style={{ backgroundColor: '#f1f5f9', color: '#64748b' }}>Webhook</span>
                        </td>
                      )}
                      {visibleColumns.includes('gst_invoice') && (
                        <td>
                          <a 
                            href={`http://localhost:8080/api/orders/invoice?id=${order.id}`} 
                            target="_blank" 
                            rel="noreferrer" 
                            download={`invoice-${order.order_number}.pdf`}
                            className="btn-primary" 
                            style={{fontSize: '0.8rem', padding: '0.4rem 0.8rem', display: 'inline-block'}}
                          >
                            Download
                          </a>
                        </td>
                      )}
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
          
          {/* Pagination Controls */}
          {activeTab === 'shopify' && orders.length > 0 && (
            <div style={{ padding: '1.5rem', borderTop: '1px solid var(--border-color)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
                {Math.min((page - 1) * limit + 1, totalCount)}–{Math.min(page * limit, totalCount)} of {totalCount} orders
              </div>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <button 
                  className="btn-secondary" 
                  onClick={() => setPage(prev => Math.max(prev - 1, 1))}
                  disabled={page === 1 || isLoading}
                  style={{ padding: '0.4rem 1rem', fontSize: '0.875rem' }}
                >
                  Previous
                </button>
                <button 
                  className="btn-secondary" 
                  onClick={() => setPage(prev => prev + 1)}
                  disabled={page * limit >= totalCount || isLoading}
                  style={{ padding: '0.4rem 1rem', fontSize: '0.875rem' }}
                >
                  Next
                </button>
              </div>
            </div>
          )}
          </section>
        )}

        {activeTab === 'reports' && (
          <GSTReports startDate={startDate} endDate={endDate} />
        )}

        {activeTab === 'automation' && (
          <WhatsAppAutomation />
        )}

        {activeTab === 'invoices' && (
          <div style={{ padding: '2rem', textAlign: 'center' }}>
            <h3>Invoices Manager</h3>
            <p>Select orders from the Orders tab to download or manage invoices.</p>
          </div>
        )}
      </main>
    </div>
  );
}

export default App;
