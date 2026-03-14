import { useState, useEffect } from 'react';
import { CustomDatePicker } from './CustomDatePicker';
import { ColumnSelector } from './ColumnSelector';
import type { ColumnOption } from './ColumnSelector';
import { GSTReports } from './GSTReports';
import { WhatsAppAutomation } from './WhatsAppAutomation';
import fullLogo from './assets/full_logo.png';
import { Login } from './Login';
import './App.css';

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';

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
  delivery_status: string;
  tracking_number: string;
  shipping_company: string;
  tracking_url: string;
  status: string;
  source_id: string;
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
  { id: 'delivery_status', label: 'Delivery', category: 'Status', isDefault: true },
  { id: 'source', label: 'Source', category: 'General', isDefault: true },
  { id: 'gst_invoice', label: 'GST Invoice', category: 'General', isDefault: true },
];

const DEFAULT_VISIBLE_COLUMNS = AVAILABLE_COLUMNS.filter(c => c.isDefault).map(c => c.id);

function App() {
  const [token, setToken] = useState<string | null>(localStorage.getItem('token'));
  const [activeTab, setActiveTab] = useState<string>(() => {
    return localStorage.getItem('gstAppActiveTab') || 'dashboard';
  });

  const handleLogin = (newToken: string) => {
    localStorage.setItem('token', newToken);
    setToken(newToken);
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    setToken(null);
  };

  const fetchWithAuth = async (url: string, options: RequestInit = {}) => {
    const headers = {
      ...options.headers,
      'Authorization': `Bearer ${token}`
    };
    const response = await fetch(url, { ...options, headers });
    if (response.status === 401) {
      handleLogout();
    }
    return response;
  };

  useEffect(() => {
    localStorage.setItem('gstAppActiveTab', activeTab);
  }, [activeTab]);
  const [orders, setOrders] = useState<Order[]>([]);
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null);
  const [isSyncing, setIsSyncing] = useState(false);
  const [isResetting, setIsResetting] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [totalCount, setTotalCount] = useState(0);
  const [webhookStatus, setWebhookStatus] = useState<WebhookStatus | null>(null);
  const [appSettings, setAppSettings] = useState<Record<string, string>>({});
  const limit = 25;
  const [openTrackingId, setOpenTrackingId] = useState<string | null>(null);
  const [editingStatusId, setEditingStatusId] = useState<string | null>(null);
  const [isUpdatingStatus, setIsUpdatingStatus] = useState(false);
  
  // Sync Modal State
  const [showSyncModal, setShowSyncModal] = useState(false);
  const [syncStartDate, setSyncStartDate] = useState(new Date().toISOString().split('T')[0]);
  const [syncEndDate, setSyncEndDate] = useState(new Date().toISOString().split('T')[0]);
  const [syncStep, setSyncStep] = useState<'date' | 'confirm'>('date');

  // Sorting and Filtering State
  const [search, setSearch] = useState('');
  const [sourceFilter, setSourceFilter] = useState('');
  const [paymentFilter, setPaymentFilter] = useState('');
  const [fulfillmentFilter, setFulfillmentFilter] = useState('');
  const [sortBy, setSortBy] = useState('created_at');
  const [sortOrder, setSortOrder] = useState<'ASC' | 'DESC'>('DESC');

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

  // Default to Year-to-Date (YTD) or January 1st as requested
  const defaultStartDate = new Date(new Date().getFullYear(), 0, 1).toISOString().split('T')[0];
  const defaultEndDate = new Date().toISOString().split('T')[0];
  const [startDate, setStartDate] = useState(defaultStartDate);
  const [endDate, setEndDate] = useState(defaultEndDate);

  // Load saved date range from backend on startup
  useEffect(() => {
    if (!token) return;
    fetchWithAuth('http://localhost:8080/api/settings/date-range')
      .then(res => res.json())
      .then(data => {
        if (data.success && data.start_date && data.end_date) {
          setStartDate(data.start_date);
          setEndDate(data.end_date);
        }
      })
      .catch(() => {});
  }, [token]);

  // Close tracking/status popover when clicking elsewhere
  useEffect(() => {
    const handleOutsideClick = () => {
      if (openTrackingId) setOpenTrackingId(null);
      if (editingStatusId) setEditingStatusId(null);
    };
    window.addEventListener('click', handleOutsideClick);
    return () => window.removeEventListener('click', handleOutsideClick);
  }, [openTrackingId, editingStatusId]);

  // Debounced search effect
  const [debouncedSearch, setDebouncedSearch] = useState(search);
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
    }, 300);
    return () => clearTimeout(timer);
  }, [search]);

  const handleStatusUpdate = async (orderId: string, newStatus: string) => {
    setIsUpdatingStatus(true);
    try {
      const response = await fetchWithAuth(`http://localhost:8080/api/orders/status?id=${orderId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status: newStatus }),
      });
      const data = await response.json();
      if (data.success) {
        // Refresh data
        fetchDashboardData();
        setEditingStatusId(null);
      } else {
        alert(data.message || 'Failed to update status');
      }
    } catch (error) {
      console.error('Error updating status:', error);
      alert('Network error updating status');
    } finally {
      setIsUpdatingStatus(false);
    }
  };

  const fetchAppSettings = async () => {
    if (!token) return;
    try {
      const response = await fetchWithAuth(`${API_BASE}/api/settings`);
      const data = await response.json();
      if (data.success) {
        setAppSettings(data.settings);
      }
    } catch (error) {
      console.error('Error fetching settings:', error);
    }
  };

  // Dedicated effect for settings - only on mount or token change
  useEffect(() => {
    if (token) {
      fetchAppSettings();
    }
  }, [token]);

  const fetchDashboardData = async (silent = false) => {
    if (!token) return;
    if (!silent) setIsLoading(true);
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
      
      const metricsRes = await fetchWithAuth(`${API_BASE}/api/dashboard/metrics?start_date=${startObj}&end_date=${endObj}`);
      const metricsData = await metricsRes.json();
      
      if (metricsData.success) {
        setMetrics(metricsData.metrics);
      }

      const ordersRes = await fetchWithAuth(`${API_BASE}/api/orders?start_date=${startObj}&end_date=${endObj}&page=${page}&limit=${limit}&search=${debouncedSearch}&source=${sourceFilter}&financial_status=${paymentFilter}&fulfillment_status=${fulfillmentFilter}&sort_by=${sortBy}&sort_order=${sortOrder}`);
      const ordersData = await ordersRes.json();
      if (ordersData.success) {
        setOrders(ordersData.orders);
        setTotalCount(ordersData.total_count);
      }

      const webhookRes = await fetchWithAuth(`${API_BASE}/api/webhook/status`);
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
    setShowSyncModal(false);
    try {
      const response = await fetchWithAuth(`${API_BASE}/api/shopify/sync`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          start_date: syncStartDate,
          end_date: syncEndDate
        })
      });
      const data = await response.json();
      if (data.success) {
        alert(`Successfully synced ${data.count} orders!`);
        fetchDashboardData();
      } else {
        alert(data.message || 'Failed to sync orders.');
      }
    } catch (error) {
      console.error('Error syncing orders:', error);
      alert('Error occurred while syncing.');
    } finally {
      setIsSyncing(false);
    }
  };

  const resetShopify = async () => {
    if (!window.confirm("Are you sure you want to delete all historical synced data and force a full re-sync from January 2026? This cannot be undone.")) {
      return;
    }
    setIsResetting(true);
    try {
      const response = await fetchWithAuth('http://localhost:8080/api/shopify/reset', {
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
    
    // Auto-refresh main data every 60 seconds
    const interval = setInterval(() => {
      // Only refresh if tab is active (browser might throttle intervals anyway, but good to be explicit)
      if (document.visibilityState === 'visible') {
        fetchDashboardData(true);
      }
    }, 60000);

    return () => clearInterval(interval);
  }, [startDate, endDate, page, search, sourceFilter, paymentFilter, fulfillmentFilter, sortBy, sortOrder]);

  const handleDownloadInvoice = async (orderId: string, orderNumber: string) => {
    try {
      const response = await fetchWithAuth(`${API_BASE}/api/orders/invoice?id=${orderId}`);
      if (!response.ok) {
        throw new Error('Failed to download invoice');
      }
      
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `invoice-${orderNumber}.pdf`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (error) {
      console.error('Error downloading invoice:', error);
      alert('Failed to download invoice. Please try again.');
    }
  };

  if (!token) {
    return <Login onLogin={handleLogin} />;
  }

  const minSyncDate = '2026-01-01';

  return (
    <div className="app-container">
      {showSyncModal && (
        <div className="modal-overlay">
          <div className="premium-modal">
            <div className="modal-header-icon">
              {syncStep === 'date' ? (
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
              ) : (
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z"></path><polyline points="12 6 12 12 16 14"></polyline></svg>
              )}
            </div>
            
            <h2>{syncStep === 'date' ? 'Manual Sync' : 'Ready to Sync?'}</h2>
            
            {syncStep === 'date' ? (
              <div className="step-content">
                <p>Select the date range you wish to synchronize from Shopify. Existing orders will be updated.</p>
                
                <div className="sync-form-group">
                  <label>Sync From</label>
                  <div className="sync-input-wrapper">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
                    <input 
                      type="date" 
                      className="premium-date-input" 
                      min={minSyncDate}
                      max={new Date().toISOString().split('T')[0]}
                      value={syncStartDate}
                      onChange={(e) => setSyncStartDate(e.target.value)}
                    />
                  </div>
                </div>

                <div className="sync-form-group">
                  <label>Sync To</label>
                  <div className="sync-input-wrapper">
                     <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
                    <input 
                      type="date" 
                      className="premium-date-input" 
                      min={syncStartDate}
                      max={new Date().toISOString().split('T')[0]}
                      value={syncEndDate}
                      onChange={(e) => setSyncEndDate(e.target.value)}
                    />
                  </div>
                </div>

                <div className="modal-actions">
                  <button className="btn-secondary" onClick={() => setShowSyncModal(false)}>Cancel</button>
                  <button className="btn-primary" onClick={() => setSyncStep('confirm')}>Next</button>
                </div>
              </div>
            ) : (
              <div className="step-content">
                <p>Please review your sync settings before proceeding.</p>
                
                <div className="confirm-box">
                  <div className="confirm-row">
                    <span className="confirm-label">Start Date</span>
                    <span className="confirm-value">{syncStartDate}</span>
                  </div>
                  <div className="confirm-row" style={{borderTop: '1px solid rgba(0,0,0,0.05)', marginTop: '4px', paddingTop: '4px'}}>
                    <span className="confirm-label">End Date</span>
                    <span className="confirm-value">{syncEndDate}</span>
                  </div>
                </div>

                <div className="info-banner">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{flexShrink: 0}}><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>
                  <span>Your existing PII data (Customer Names, Emails, Phones) is safe and will be preserved automatically.</span>
                </div>

                <div className="modal-actions" style={{marginTop: '2rem'}}>
                  <button className="btn-secondary" onClick={() => setSyncStep('date')}>Back</button>
                  <button className="btn-primary" onClick={syncShopify}>Start Sync</button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
      <aside className="sidebar">
        <div className="sidebar-brand" style={{ justifyContent: 'flex-start', paddingLeft: '1rem', marginBottom: '2rem' }}>
          <img src={fullLogo} alt="Mi Tech" style={{ width: '140px', height: 'auto', objectFit: 'contain' }} />
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
          <a href="#" className="nav-item" onClick={handleLogout} style={{ marginTop: 'auto', color: '#ef4444' }}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path><polyline points="16 17 21 12 16 7"></polyline><line x1="21" y1="12" x2="9" y2="12"></line></svg>
            Sign Out
          </a>
        </nav>
      </aside>

      <main className="main-content">
        <header className="page-header">
          <div>
            <h1 className="page-title">{activeTab === 'dashboard' ? 'Overview' : activeTab === 'shopify' ? 'Orders' : activeTab === 'reports' ? 'GST Reports' : activeTab === 'automation' ? 'Automation Hub' : 'Settings'}</h1>
            <p className="page-subtitle">
              {activeTab === 'dashboard' ? "Welcome back. Here's what's happening today." : activeTab === 'reports' ? "Review your GST collection and generate filing reports." : activeTab === 'automation' ? "Manage templates, triggers, and track WhatsApp communication." : activeTab === 'shopify' ? "Real-time orders synced via Shopify Webhooks." : "Manage your store data and preferences."}
            </p>
          </div>
          {activeTab !== 'automation' && (
            <div style={{display: 'flex', gap: '1rem'}}>
              <button className="btn-secondary">Export Data</button>
              {appSettings?.show_reset_button === 'true' && (
                <button 
                  className="btn-secondary" 
                  style={{display: 'flex', alignItems: 'center', gap: '0.5rem', opacity: isResetting ? 0.7 : 1, backgroundColor: '#ef4444', color: 'white', borderColor: '#ef4444'}}
                  onClick={resetShopify}
                  disabled={isResetting || isSyncing}
                >
                  {isResetting ? 'Resetting...' : 'Reset & Resync'}
                </button>
              )}
                 <button 
                className="btn-primary" 
                title="Manually fetch orders from Shopify in case webhook delivery fails."
                style={{display: 'flex', alignItems: 'center', gap: '0.5rem', opacity: isSyncing ? 0.7 : 1}}
                onClick={() => {
                  setSyncStartDate(new Date().toISOString().split('T')[0]);
                  setSyncEndDate(new Date().toISOString().split('T')[0]);
                  setSyncStep('date');
                  setShowSyncModal(true);
                }}
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
                // Persist date range to backend
                fetchWithAuth('http://localhost:8080/api/settings/date-range', {
                  method: 'PUT',
                  headers: { 'Content-Type': 'application/json' },
                  body: JSON.stringify({ start_date: start, end_date: end }),
                }).catch(console.error);
              }} 
            />
          </div>
        )}

        {activeTab === 'dashboard' && (
          <>
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
          </>
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

            <div className="table-header" style={{ flexDirection: 'column', alignItems: 'stretch', gap: '1.5rem' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <h3 style={{ fontSize: '1rem', margin: 0 }}>Stored Orders</h3>
                <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                  <ColumnSelector columns={AVAILABLE_COLUMNS} visibleColumns={visibleColumns} onChange={setVisibleColumns} />
                </div>
              </div>

              <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap', alignItems: 'center', backgroundColor: 'white', padding: '0.5rem', borderRadius: '8px' }}>
                <div style={{ flex: 1, minWidth: '200px', position: 'relative' }}>
                  <svg style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: '#94a3b8' }} width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
                  <input 
                    type="text" 
                    placeholder="Search orders or customers..." 
                    value={search}
                    onChange={(e) => { setSearch(e.target.value); setPage(1); }}
                    style={{ paddingLeft: '2.5rem', fontSize: '0.875rem' }}
                  />
                </div>
                
                <select 
                  value={sourceFilter} 
                  onChange={(e) => { setSourceFilter(e.target.value); setPage(1); }}
                  style={{ width: 'auto', fontSize: '0.875rem', padding: '0.5rem 2rem 0.5rem 1rem' }}
                >
                  <option value="">All Sources</option>
                  <option value="shopify">Shopify</option>
                  <option value="amazon">Amazon</option>
                  <option value="pos">POS</option>
                </select>

                <select 
                  value={paymentFilter} 
                  onChange={(e) => { setPaymentFilter(e.target.value); setPage(1); }}
                  style={{ width: 'auto', fontSize: '0.875rem', padding: '0.5rem 2rem 0.5rem 1rem' }}
                >
                  <option value="">Payment: All</option>
                  <option value="paid">Paid</option>
                  <option value="unpaid">Unpaid</option>
                </select>

                <select 
                  value={fulfillmentFilter} 
                  onChange={(e) => { setFulfillmentFilter(e.target.value); setPage(1); }}
                  style={{ width: 'auto', fontSize: '0.875rem', padding: '0.5rem 2rem 0.5rem 1rem' }}
                >
                  <option value="">Fulfillment: All</option>
                  <option value="fulfilled">Fulfilled</option>
                  <option value="unfulfilled">Unfulfilled</option>
                </select>

                {(search || sourceFilter || paymentFilter || fulfillmentFilter) && (
                  <button 
                    className="btn-secondary" 
                    onClick={() => { setSearch(''); setSourceFilter(''); setPaymentFilter(''); setFulfillmentFilter(''); setPage(1); }}
                    style={{ padding: '0.5rem 1rem', fontSize: '0.875rem', color: '#64748b' }}
                  >
                    Clear Filters
                  </button>
                )}
              </div>
            </div>
            <div style={{overflowX: 'auto'}}>
            <table>
              <thead>
                <tr>
                  {visibleColumns.includes('order_id') && (
                    <th onClick={() => { setSortBy('order_number'); setSortOrder(prev => prev === 'ASC' ? 'DESC' : 'ASC'); }} style={{ cursor: 'pointer' }}>
                      Order ID {sortBy === 'order_number' && (sortOrder === 'ASC' ? ' ↑' : ' ↓')}
                    </th>
                  )}
                  {visibleColumns.includes('customer_name') && (
                    <th onClick={() => { setSortBy('customer_name'); setSortOrder(prev => prev === 'ASC' ? 'DESC' : 'ASC'); }} style={{ cursor: 'pointer' }}>
                      Customer {sortBy === 'customer_name' && (sortOrder === 'ASC' ? ' ↑' : ' ↓')}
                    </th>
                  )}
                  {visibleColumns.includes('city') && <th>City</th>}
                  {visibleColumns.includes('state') && <th>State</th>}
                  {visibleColumns.includes('country') && <th>Country</th>}
                  {visibleColumns.includes('date') && (
                    <th onClick={() => { setSortBy('created_at'); setSortOrder(prev => prev === 'ASC' ? 'DESC' : 'ASC'); }} style={{ cursor: 'pointer' }}>
                      Date {sortBy === 'created_at' && (sortOrder === 'ASC' ? ' ↑' : ' ↓')}
                    </th>
                  )}
                  {visibleColumns.includes('time') && <th>Time</th>}
                  {visibleColumns.includes('amount') && (
                    <th onClick={() => { setSortBy('total_price'); setSortOrder(prev => prev === 'ASC' ? 'DESC' : 'ASC'); }} style={{ cursor: 'pointer' }}>
                      Amount {sortBy === 'total_price' && (sortOrder === 'ASC' ? ' ↑' : ' ↓')}
                    </th>
                  )}
                  {visibleColumns.includes('financial_status') && (
                    <th onClick={() => { setSortBy('financial_status'); setSortOrder(prev => prev === 'ASC' ? 'DESC' : 'ASC'); }} style={{ cursor: 'pointer' }}>
                      Payment {sortBy === 'financial_status' && (sortOrder === 'ASC' ? ' ↑' : ' ↓')}
                    </th>
                  )}
                  {visibleColumns.includes('fulfillment_status') && (
                    <th onClick={() => { setSortBy('fulfillment_status'); setSortOrder(prev => prev === 'ASC' ? 'DESC' : 'ASC'); }} style={{ cursor: 'pointer' }}>
                      Fulfillment {sortBy === 'fulfillment_status' && (sortOrder === 'ASC' ? ' ↑' : ' ↓')}
                    </th>
                  )}
                  {visibleColumns.includes('delivery_status') && <th>Delivery Status</th>}
                  {visibleColumns.includes('source') && (
                    <th onClick={() => { setSortBy('source_id'); setSortOrder(prev => prev === 'ASC' ? 'DESC' : 'ASC'); }} style={{ cursor: 'pointer' }}>
                      Source {sortBy === 'source_id' && (sortOrder === 'ASC' ? ' ↑' : ' ↓')}
                    </th>
                  )}
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
                          <span className={`badge-pill badge-pill-${order.financial_status === 'paid' ? 'success' : 'warning'}`}>
                            <span className="dot"></span> {order.financial_status?.charAt(0).toUpperCase() + order.financial_status?.slice(1) || 'Unknown'}
                          </span>
                        </td>
                      )}
                      {visibleColumns.includes('fulfillment_status') && (
                        <td style={{ position: 'relative' }}>
                          <span 
                            className={`badge-pill badge-pill-${order.fulfillment_status === 'fulfilled' ? 'gray' : (order.fulfillment_status === 'cancelled' || order.status === 'CANCELLED' ? 'danger' : 'yellow')}`}
                            style={{ cursor: isUpdatingStatus ? 'not-allowed' : 'pointer', opacity: isUpdatingStatus && editingStatusId === order.id ? 0.7 : 1 }}
                            onClick={(e) => {
                              if (isUpdatingStatus) return;
                              e.stopPropagation();
                              setEditingStatusId(editingStatusId === order.id ? null : order.id);
                            }}
                          >
                             <span className="dot"></span> {isUpdatingStatus && editingStatusId === order.id ? 'Updating...' : (order.fulfillment_status?.charAt(0).toUpperCase() + order.fulfillment_status?.slice(1) || 'Unfulfilled')}
                          </span>

                          {editingStatusId === order.id && (
                            <div className="status-popover" onClick={e => e.stopPropagation()}>
                              <div className="status-popover-header">Update Status</div>
                              <div 
                                className="status-option"
                                onClick={() => handleStatusUpdate(order.id, 'fulfilled')}
                              >
                                <span className="badge-pill badge-pill-gray"><span className="dot"></span> Fulfilled</span>
                              </div>
                              <div 
                                className="status-option"
                                onClick={() => handleStatusUpdate(order.id, 'unfulfilled')}
                              >
                                <span className="badge-pill badge-pill-yellow"><span className="dot"></span> Unfulfilled</span>
                              </div>
                              <div 
                                className="status-option"
                                onClick={() => handleStatusUpdate(order.id, 'cancelled')}
                              >
                                <span className="badge-pill badge-pill-danger"><span className="dot"></span> Cancelled</span>
                              </div>
                            </div>
                          )}
                        </td>
                      )}
                      {visibleColumns.includes('delivery_status') && (
                        <td style={{ position: 'relative' }}>
                          {order.delivery_status && order.delivery_status !== 'pending' && order.delivery_status !== 'fulfilled' ? (
                            <div className="delivery-status-container" style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                              <span
                                className="badge-pill badge-pill-info"
                                style={{ 
                                  cursor: order.tracking_number ? 'pointer' : 'default',
                                  display: 'inline-flex',
                                  alignItems: 'center',
                                  whiteSpace: 'nowrap'
                                }}
                                onClick={(e) => {
                                  e.stopPropagation();
                                  if (order.tracking_number) {
                                    setOpenTrackingId(openTrackingId === order.id ? null : order.id);
                                  }
                                }}
                              >
                                 <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{marginRight: '6px'}}><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path><circle cx="12" cy="12" r="3"></circle></svg>
                                 {order.delivery_status?.charAt(0).toUpperCase() + order.delivery_status?.slice(1)}
                              </span>

                              {openTrackingId === order.id && (
                                <div className="tracking-popover" onClick={e => e.stopPropagation()}>
                                  <div className="tracking-header">
                                    <span className="tracking-id">{order.order_number}</span>
                                    <span className="badge-pill badge-pill-info" style={{ fontSize: '0.7rem', padding: '2px 8px' }}>
                                      {order.delivery_status?.charAt(0).toUpperCase() + order.delivery_status?.slice(1)}
                                    </span>
                                  </div>
                                  
                                  <div className="tracking-info-row">
                                    <div className="tracking-icon-container">
                                      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="var(--accent-color)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                        <rect x="1" y="3" width="15" height="13"/><polygon points="16 8 20 8 23 11 23 16 16 16 16 8"/><circle cx="5.5" cy="18.5" r="2.5"/><circle cx="18.5" cy="18.5" r="2.5"/>
                                      </svg>
                                    </div>
                                    
                                    <div className="tracking-details">
                                      <a
                                        href={order.tracking_url || `https://www.delhivery.com/track/package/${order.tracking_number}`}
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        className="tracking-link"
                                        title="Open tracking page"
                                      >
                                        {order.tracking_number}
                                      </a>
                                      <div className="carrier-name">{order.shipping_company || 'Standard Tracking'}</div>
                                    </div>
                                    
                                    <button
                                      className="copy-button"
                                      title="Copy AWB Number"
                                      onClick={() => {
                                        navigator.clipboard.writeText(order.tracking_number);
                                        // Optional: add a tiny "Copied!" toast/feedback here if needed
                                      }}
                                    >
                                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/>
                                      </svg>
                                    </button>
                                  </div>
                                </div>
                              )}
                            </div>
                          ) : (
                            <span style={{color: '#94a3b8', fontSize: '0.8rem'}}>—</span>
                          )}
                        </td>
                      )}
                      {visibleColumns.includes('source') && (
                        <td>
                          <span 
                            className="badge" 
                            style={{ 
                              backgroundColor: order.source_id === 'amazon' ? '#fff7ed' : 
                                             order.source_id === 'pos' ? '#f0fdf4' : '#f1f5f9', 
                              color: order.source_id === 'amazon' ? '#c2410c' : 
                                     order.source_id === 'pos' ? '#166534' : '#64748b',
                              border: `1px solid ${order.source_id === 'amazon' ? '#fdba74' : 
                                                  order.source_id === 'pos' ? '#bbf7d0' : '#e2e8f0'}`
                            }}
                          >
                            {order.source_id?.charAt(0).toUpperCase() + order.source_id?.slice(1) || 'Shopify'}
                          </span>
                        </td>
                      )}
                      {visibleColumns.includes('gst_invoice') && (
                        <td>
                          <button 
                            onClick={() => handleDownloadInvoice(order.id, order.order_number)}
                            className="btn-primary" 
                            style={{fontSize: '0.8rem', padding: '0.4rem 0.8rem', display: 'inline-block', cursor: 'pointer', border: 'none'}}
                          >
                            Download
                          </button>
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
          <GSTReports startDate={startDate} endDate={endDate} fetchWithAuth={fetchWithAuth} />
        )}

        {activeTab === 'automation' && (
          <WhatsAppAutomation fetchWithAuth={fetchWithAuth} />
        )}

      </main>
    </div>
  );
}

export default App;
