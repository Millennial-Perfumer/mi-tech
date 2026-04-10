import { API_BASE } from './api';
import { useState, useEffect, useMemo } from 'react';
import { CustomDatePicker } from './CustomDatePicker';
import { ColumnSelector } from './ColumnSelector';
import type { ColumnOption } from './ColumnSelector';
import { GSTReports } from './GSTReports';
import { WhatsAppAutomation } from './WhatsAppAutomation';
import fullLogo from './assets/full_logo.png';
import fullLogoDark from './assets/full_logo_dark_theme.png';
import halfLogo from './assets/half_logo.png';
import { Login } from './Login';
import { ManualWhatsAppModal } from './ManualWhatsAppModal';
import { SettingsTab } from './SettingsTab';
import { Customers } from './Customers';
import { Users } from './Users';
import { MarketingDashboard } from './MarketingDashboard';
import { SocialDashboard } from './SocialDashboard';
import { Planner } from './Planner';
import { Tickets } from './Tickets';
import { WhatsAppChat } from './WhatsAppChat';
import Feedback from './Feedback';
import OrderDetailsModal from './OrderDetailsModal';

import { useToast } from './ToastContext';
import { useConfirm } from './ConfirmContext';
import './App.css';


const getTodayIST = () => {
  return new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Kolkata' });
};

interface Order {
  id: string | number;
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
  customer_phone: string;
  feedback_status_id?: number;
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
  { id: 'whatsapp', label: 'WhatsApp', category: 'General', isDefault: true },
  { id: 'gst_invoice', label: 'GST Invoice', category: 'General', isDefault: true },
  { id: 'feedback_status', label: 'Feedback', category: 'Status', isDefault: false },
];

const DEFAULT_VISIBLE_COLUMNS = AVAILABLE_COLUMNS.filter(c => c.isDefault).map(c => c.id);

function App() {
  const { success: toastSuccess, error: toastError } = useToast();
  const { confirm } = useConfirm();

  const [token, setToken] = useState<string | null>(localStorage.getItem('token'));
  const [activeTab, setActiveTab] = useState<string>(() => {
    return localStorage.getItem('gstAppActiveTab') || 'dashboard';
  });
  const [theme, setTheme] = useState<'light' | 'dark'>(() => {
    return (localStorage.getItem('appTheme') as 'light' | 'dark') || 'light';
  });
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(() => localStorage.getItem('sidebarCollapsed') === 'true');
  const [isMobile, setIsMobile] = useState(() => window.matchMedia('(max-width: 768px)').matches);

  useEffect(() => {
    const mq = window.matchMedia('(max-width: 768px)');
    const handler = (e: MediaQueryListEvent) => setIsMobile(e.matches);
    mq.addEventListener('change', handler);
    return () => mq.removeEventListener('change', handler);
  }, []);

  // Apply theme to <html> element
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('appTheme', theme);
  }, [theme]);

  // Sidebar persistence
  useEffect(() => {
    localStorage.setItem('sidebarCollapsed', isSidebarCollapsed.toString());
  }, [isSidebarCollapsed]);

  const toggleSidebar = () => setIsSidebarCollapsed(!isSidebarCollapsed);
  const toggleTheme = () => setTheme(t => t === 'light' ? 'dark' : 'light');

  const userRole = useMemo(() => {
    if (!token) return 'read';
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      return payload?.role || 'read';
    } catch (err) {
      console.error('Error parsing token:', err);
      return 'read';
    }
  }, [token]);
  
  useEffect(() => {
    console.log('Current userRole:', userRole);
  }, [userRole]);

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
  const [appConfigs, setAppConfigs] = useState<Record<string, string>>({});
  const limit = 25;
  const [trackingOrder, setTrackingOrder] = useState<Order | null>(null);
  const [selectedOrderDetailsId, setSelectedOrderDetailsId] = useState<string | number | null>(null);
  const [editingStatusId, setEditingStatusId] = useState<string | number | null>(null);
  const [isUpdatingStatus, setIsUpdatingStatus] = useState(false);
  const [isMarkingAsDelivered, setIsMarkingAsDelivered] = useState<number | string | null>(null);
  const [whatsappOrder, setWhatsappOrder] = useState<Order | null>(null);
  // Sync Modal State
  const [showSyncModal, setShowSyncModal] = useState(false);
  const [syncStartDate, setSyncStartDate] = useState(getTodayIST());
  const [syncEndDate, setSyncEndDate] = useState(getTodayIST());

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

  const [refreshTrigger, setRefreshTrigger] = useState(0);

  const triggerRefresh = () => setRefreshTrigger(prev => prev + 1);

  useEffect(() => {
    localStorage.setItem('shopifyAppVisibleColumns', JSON.stringify(visibleColumns));
  }, [visibleColumns]);

  // Default to Year-to-Date (YTD) or January 1st as requested
  const defaultStartDate = new Date(new Date().getFullYear(), 0, 1).toISOString().split('T')[0];
  const defaultEndDate = getTodayIST();
  
  // Initialize with values from LocalStorage if available, otherwise defaults
  const [startDate, setStartDate] = useState(() => localStorage.getItem('socialSmmStartDate') || defaultStartDate);
  const [endDate, setEndDate] = useState(() => localStorage.getItem('socialSmmEndDate') || defaultEndDate);

  // Load saved date range from backend on startup
  useEffect(() => {
    if (!token) return;
    fetchWithAuth(`${API_BASE}/api/settings/date-range`)
      .then(res => res.json())
      .then(data => {
        if (data.success && data.start_date && data.end_date) {
          console.log('DEBUG: Loaded date range from backend:', data.start_date, data.end_date);
          setStartDate(data.start_date);
          setEndDate(data.end_date);
          localStorage.setItem('socialSmmStartDate', data.start_date);
          localStorage.setItem('socialSmmEndDate', data.end_date);
        }
      })
      .catch((err) => console.error('Failed to load date range:', err));
  }, [token]);

  const handleUpdateDateRange = (start: string, end: string) => {
    console.log('DEBUG: Updating global date range:', start, end);
    setPage(1);
    setStartDate(start);
    setEndDate(end);
    localStorage.setItem('socialSmmStartDate', start);
    localStorage.setItem('socialSmmEndDate', end);
    // Persist date range to backend
    fetchWithAuth(`${API_BASE}/api/settings/date-range`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ start_date: start, end_date: end }),
    }).catch(console.error);
  };

  // Close status popover when clicking elsewhere
  useEffect(() => {
    const handleOutsideClick = () => {
      if (editingStatusId) setEditingStatusId(null);
    };
    window.addEventListener('click', handleOutsideClick);
    return () => window.removeEventListener('click', handleOutsideClick);
  }, [editingStatusId]);

  // Debounced search effect
  const [debouncedSearch, setDebouncedSearch] = useState(search);
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
    }, 300);
    return () => clearTimeout(timer);
  }, [search]);

  const handleStatusUpdate = async (orderId: string | number, newStatus: string) => {
    setIsUpdatingStatus(true);
    try {
      const response = await fetchWithAuth(`${API_BASE}/api/orders/status?id=${orderId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status: newStatus }),
      });
      const data = await response.json();
      if (data.success) {
        toastSuccess('Status updated successfully');
        // Refresh data
        fetchDashboardData();
        setEditingStatusId(null);
      } else {
        toastError(data.message || 'Failed to update status');
      }
    } catch (error) {
      console.error('Error updating status:', error);
      toastError('Network error updating status');
    } finally {
      setIsUpdatingStatus(false);
    }
  };

  const handleMarkAsDelivered = async (orderId: string | number) => {
    setIsMarkingAsDelivered(orderId);
    try {
      const response = await fetchWithAuth(`${API_BASE}/api/orders/delivered?id=${orderId}`, {
        method: 'POST',
      });
      const data = await response.json();
      if (data.success) {
        toastSuccess('Order marked as delivered');
        fetchDashboardData();
      } else {
        toastError(data.message || 'Failed to update delivery status');
      }
    } catch (error) {
      console.error('Error marking as delivered:', error);
      toastError('Network error updating delivery status');
    } finally {
      setIsMarkingAsDelivered(null);
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
    } catch (err) {
      console.error('Failed to fetch app settings:', err);
    }
  };

  const fetchAppConfigs = async () => {
    if (!token) return;
    try {
      const response = await fetchWithAuth(`${API_BASE}/api/configs`);
      const data = await response.json();
      if (data.success && Array.isArray(data.configs)) {
        const configsMap: Record<string, string> = {};
        data.configs.forEach((cfg: any) => {
          configsMap[cfg.key] = cfg.value;
        });
        setAppConfigs(configsMap);
      }
    } catch (err) {
      console.error('Failed to fetch app configs:', err);
    }
  };


  // Dedicated effect for settings - only on mount or token change
  useEffect(() => {
    if (token) {
      fetchAppSettings();
      fetchAppConfigs();
    }
  }, [token]);

  const fetchDashboardData = async (silent = false, force = false) => {
    if (!token) return;

    // Only fetch dashboard/orders data if specifically on those tabs (unless forced)
    const isOnDashboard = activeTab === 'dashboard';
    const isOnOrders = activeTab === 'shopify';

    if (!force && !isOnDashboard && !isOnOrders) {
      return;
    }

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

      // Optimization: Conditional parallel fetching based on activeTab
      const fetchTasks: Promise<any>[] = [];

      // 1. Fetch Metrics (Dashboard only)
      if (force || isOnDashboard) {
        fetchTasks.push(
          fetchWithAuth(`${API_BASE}/api/dashboard/metrics?start_date=${startObj}&end_date=${endObj}`)
            .then(res => res.json())
            .then(data => { if (data.success) setMetrics(data.metrics); })
        );
      }

      // 2. Fetch Orders (Orders tab only)
      if (force || isOnOrders) {
        fetchTasks.push(
          fetchWithAuth(`${API_BASE}/api/orders?start_date=${startObj}&end_date=${endObj}&page=${page}&limit=${limit}&search=${debouncedSearch}&source=${sourceFilter}&financial_status=${paymentFilter}&fulfillment_status=${fulfillmentFilter}&sort_by=${sortBy}&sort_order=${sortOrder}`)
            .then(res => res.json())
            .then(data => {
              if (data.success) {
                setOrders(data.orders);
                setTotalCount(data.total_count);
              }
            })
        );

        // 3. Webhook Status (Orders tab only)
        fetchTasks.push(
          fetchWithAuth(`${API_BASE}/api/webhook/status`)
            .then(res => res.json())
            .then(data => setWebhookStatus(data))
        );
      }

      // Execute all required fetches in parallel
      await Promise.all(fetchTasks);

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
        toastSuccess(`Successfully synced ${data.count} orders!`);
        triggerRefresh();
        fetchDashboardData(false, true);
      } else {
        toastError(data.message || 'Failed to sync orders.');
      }
    } catch (error) {
      console.error('Error syncing orders:', error);
      toastError('Error occurred while syncing.');
    } finally {
      setIsSyncing(false);
    }
  };

  const resetShopify = async () => {
    const confirmed = await confirm({
      title: 'Full Database Reset',
      message: 'Are you sure you want to delete all historical synced data and force a full re-sync from January 2026? This cannot be undone.',
      variant: 'danger',
      confirmLabel: 'Reset Everything'
    });

    if (!confirmed) return;

    setIsResetting(true);
    try {
      const response = await fetchWithAuth(`${API_BASE}/api/shopify/reset`, {
        method: 'POST',
      });
      const data = await response.json();
      if (data.success) {
        toastSuccess(`Successfully wiped data and re-synced ${data.count} orders!`);
        fetchDashboardData();
      } else {
        toastError('Failed to reset orders.');
      }
    } catch (error) {
      console.error('Error resetting orders:', error);
      toastError('Error occurred while resetting.');
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
  }, [activeTab, startDate, endDate, page, debouncedSearch, sourceFilter, paymentFilter, fulfillmentFilter, sortBy, sortOrder]);

  const handleDownloadInvoice = async (orderId: string | number, orderNumber: string) => {
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
      toastError('Failed to download invoice. Please try again.');
    }
  };



  if (!token) {
    return <Login onLogin={handleLogin} />;
  }


  return (
    <div className="app-container">
      {showSyncModal && (
        <div className="modal-overlay">
          <div className="premium-modal wide">
            <div className="modal-header-icon" style={{ background: 'linear-gradient(135deg, #10b981, #059669)' }}>
              <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
            </div>
            
            <h2 style={{ marginBottom: '0.5rem' }}>Manual Synchronization</h2>
            
            <div className="step-content">
              <p style={{marginBottom: '2rem'}}>Select the date range you wish to synchronize from Shopify. Existing orders will be updated.</p>
              
              <div className="sync-date-selector" style={{ marginBottom: '2rem' }}>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(100px, 1fr))', gap: '8px', marginBottom: '1.5rem' }}>
                  {[
                    { label: 'Today', days: 0 },
                    { label: 'Yesterday', days: 1, exact: true },
                    { label: 'Last 7 days', days: 7 },
                    { label: 'Last 30 days', days: 30 },
                    { label: 'This Month', type: 'month', days: 0 }
                  ].map(preset => (
                    <button
                      key={preset.label}
                      className="btn-secondary"
                      style={{ 
                        padding: '0.5rem', 
                        fontSize: '0.85rem', 
                        background: 'var(--surface-color)',
                        borderColor: 'var(--border-color)',
                        color: 'var(--text-secondary)'
                      }}
                      onMouseOver={(e) => e.currentTarget.style.borderColor = 'var(--accent-color)'}
                      onMouseOut={(e) => e.currentTarget.style.borderColor = 'var(--border-color)'}
                      onClick={() => {
                        const today = new Date();
                        if (preset.type === 'month') {
                          const firstDay = new Date(today.getFullYear(), today.getMonth(), 1);
                          setSyncStartDate(firstDay.toISOString().split('T')[0]);
                          setSyncEndDate(today.toISOString().split('T')[0]);
                        } else if (preset.exact) {
                          const specificDate = new Date(today);
                          specificDate.setDate(today.getDate() - preset.days);
                          const dateStr = specificDate.toISOString().split('T')[0];
                          setSyncStartDate(dateStr);
                          setSyncEndDate(dateStr);
                        } else {
                          const pastDate = new Date(today);
                          pastDate.setDate(today.getDate() - preset.days);
                          setSyncStartDate(pastDate.toISOString().split('T')[0]);
                          setSyncEndDate(today.toISOString().split('T')[0]);
                        }
                      }}
                    >
                      {preset.label}
                    </button>
                  ))}
                </div>
                
                <div className="sync-date-row" style={{ display: 'flex', gap: '1rem', alignItems: 'center', background: 'var(--bg-input)', padding: '1.5rem', borderRadius: '16px', border: '1px solid var(--border-color)' }}>
                  <div style={{ flex: 1 }}>
                    <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-tertiary)', marginBottom: '0.5rem' }}>From Date</label>
                    <input 
                      type="date" 
                      value={syncStartDate}
                      max={syncEndDate}
                      onChange={(e) => setSyncStartDate(e.target.value)}
                      style={{ 
                        width: '100%', 
                        padding: '0.75rem', 
                        borderRadius: '8px', 
                        border: '2px solid transparent', 
                        boxShadow: '0 0 0 1px var(--border-color)',
                        outline: 'none', 
                        fontFamily: 'inherit',
                        fontSize: '0.95rem',
                        background: 'var(--surface-color)',
                        color: 'var(--text-primary)',
                        transition: 'all 0.2s'
                      }}
                      onFocus={(e) => e.currentTarget.style.boxShadow = '0 0 0 2px var(--accent-color)'}
                      onBlur={(e) => e.currentTarget.style.boxShadow = '0 0 0 1px var(--border-color)'}
                    />
                  </div>
                  <div className="sync-date-arrow" style={{ color: 'var(--text-tertiary)', fontSize: '1.5rem', alignSelf: 'flex-end', paddingBottom: '0.5rem' }}>→</div>
                  <div style={{ flex: 1 }}>
                    <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-tertiary)', marginBottom: '0.5rem' }}>To Date</label>
                    <input 
                      type="date" 
                      value={syncEndDate}
                      min={syncStartDate}
                      onChange={(e) => setSyncEndDate(e.target.value)}
                      style={{ 
                        width: '100%', 
                        padding: '0.75rem', 
                        borderRadius: '8px', 
                        border: '2px solid transparent', 
                        boxShadow: '0 0 0 1px var(--border-color)',
                        outline: 'none', 
                        fontFamily: 'inherit',
                        fontSize: '0.95rem',
                        background: 'var(--surface-color)',
                        color: 'var(--text-primary)',
                        transition: 'all 0.2s'
                      }}
                      onFocus={(e) => e.currentTarget.style.boxShadow = '0 0 0 2px var(--accent-color)'}
                      onBlur={(e) => e.currentTarget.style.boxShadow = '0 0 0 1px var(--border-color)'}
                    />
                  </div>
                </div>
              </div>

              <div className="info-banner" style={{ marginTop: '1.5rem' }}>
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{flexShrink: 0}}><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>
                <span>PII data (Customer Names, Emails, Phones) is preserved during synchronization.</span>
              </div>

              <div className="modal-actions" style={{marginTop: '2rem'}}>
                <button className="btn-secondary" onClick={() => setShowSyncModal(false)}>Cancel</button>
                <button 
                  className="btn-primary" 
                  onClick={syncShopify}
                  disabled={isSyncing}
                >
                  {isSyncing ? 'Syncing...' : 'Start Synchronization'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
      <aside className={`sidebar ${isSidebarCollapsed ? 'collapsed' : ''}`}>
        <div className="sidebar-brand" style={{ justifyContent: 'space-between', paddingLeft: '1rem', paddingRight: '0.5rem', marginBottom: '2.5rem' }}>
          <img 
            src={isSidebarCollapsed ? halfLogo : (theme === 'dark' ? fullLogoDark : fullLogo)} 
            alt="Mi Tech" 
            style={{ 
              width: isSidebarCollapsed ? '32px' : '140px', 
              height: 'auto', 
              objectFit: 'contain',
              transition: 'width 0.3s ease'
            }} 
          />
          <button 
            onClick={toggleSidebar}
            style={{ 
              color: 'var(--text-secondary)', 
              padding: '4px',
              borderRadius: '6px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              transition: 'all 0.2s',
              transform: isSidebarCollapsed ? 'rotate(180deg)' : 'none',
              backgroundColor: isSidebarCollapsed ? 'transparent' : 'var(--bg-hover)'
            }}
            title={isSidebarCollapsed ? "Expand sidebar" : "Collapse sidebar"}
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="15 18 9 12 15 6"></polyline></svg>
          </button>
        </div>
        <nav className="sidebar-nav">
          <a href="#" className={`nav-item nav-item-stagger ${activeTab === 'dashboard' ? 'active' : ''}`} onClick={() => setActiveTab('dashboard')} title={isSidebarCollapsed ? "Dashboard" : ""} style={{ animationDelay: '50ms' }}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="3" width="7" height="7"></rect><rect x="14" y="3" width="7" height="7"></rect><rect x="14" y="14" width="7" height="7"></rect><rect x="3" y="14" width="7" height="7"></rect></svg>
            <span>Dashboard</span>
          </a>
          <a href="#" className={`nav-item nav-item-stagger ${activeTab === 'reports' ? 'active' : ''}`} onClick={() => setActiveTab('reports')} title={isSidebarCollapsed ? "GST Reports" : ""} style={{ animationDelay: '100ms' }}>
             <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line><polyline points="10 9 9 9 8 9"></polyline></svg>
            <span>GST Reports</span>
          </a>
          <a href="#" className={`nav-item nav-item-stagger ${activeTab === 'automation' ? 'active' : ''}`} onClick={() => setActiveTab('automation')} title={isSidebarCollapsed ? "Automation" : ""} style={{ animationDelay: '150ms' }}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>
            <span>Automation</span>
          </a>
          <a href="#" className={`nav-item nav-item-stagger ${activeTab === 'communication' ? 'active' : ''}`} onClick={() => setActiveTab('communication')} title={isSidebarCollapsed ? "Communication" : ""} style={{ animationDelay: '160ms' }}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg>
            <span>Communication</span>
          </a>
          <a href="#" className={`nav-item nav-item-stagger ${activeTab === 'tickets' ? 'active' : ''}`} onClick={() => setActiveTab('tickets')} title={isSidebarCollapsed ? "Tickets" : ""} style={{ animationDelay: '170ms' }}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M2 9a3 3 0 0 1 0 6v2a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-2a3 3 0 0 1 0-6V7a2 2 0 0 0-2-2H4a2 2 0 0 0-2 2v2z"></path><line x1="13" y1="5" x2="13" y2="19"></line></svg>
            <span>Tickets</span>
          </a>
          <a href="#" className={`nav-item nav-item-stagger ${activeTab === 'shopify' ? 'active' : ''}`} onClick={() => setActiveTab('shopify')} title={isSidebarCollapsed ? "Orders" : ""} style={{ animationDelay: '200ms' }}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="9" cy="21" r="1"></circle><circle cx="20" cy="21" r="1"></circle><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"></path></svg>
            <span>Orders</span>
          </a>
          <a href="#" className={`nav-item nav-item-stagger ${activeTab === 'customers' ? 'active' : ''}`} onClick={() => setActiveTab('customers')} title={isSidebarCollapsed ? "Customers" : ""} style={{ animationDelay: '250ms' }}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>
            <span>Customers</span>
          </a>
          <a href="#" className={`nav-item nav-item-stagger ${activeTab === 'marketing' ? 'active' : ''}`} onClick={() => setActiveTab('marketing')} title={isSidebarCollapsed ? "Ads Intelligence" : ""} style={{ animationDelay: '275ms' }}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M11 5h2M11 9h2M11 13h2M11 17h2M7 9h10v10c0 1.1-.9 2-2 2H9c-1.1 0-2-.9-2-2V9zM18 5c1.1 0 2 .9 2 2v2H4V7c0-1.1.9-2 2-2h12z"/></svg>
            <span>Ads</span>
          </a>
          <a href="#" className={`nav-item nav-item-stagger ${activeTab === 'social' ? 'active' : ''}`} onClick={() => setActiveTab('social')} title={isSidebarCollapsed ? "Social Media" : ""} style={{ animationDelay: '285ms' }}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M16 8a6 6 0 0 1 6 6v7h-4v-7a2 2 0 0 0-2-2 2 2 0 0 0-2 2v7h-4v-7a6 6 0 0 1 6-6z"></path><rect x="2" y="9" width="4" height="12"></rect><circle cx="4" cy="4" r="2"></circle></svg>
            <span>Social Media</span>
          </a>
          <a href="#" className={`nav-item nav-item-stagger ${activeTab === 'settings' ? 'active' : ''}`} onClick={() => setActiveTab('settings')} title={isSidebarCollapsed ? "Settings" : ""} style={{ animationDelay: '300ms' }}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1-2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>
            <span>Settings</span>
          </a>
          {userRole === 'admin' && (
            <a href="#" className={`nav-item nav-item-stagger ${activeTab === 'users' ? 'active' : ''}`} onClick={() => setActiveTab('users')} title={isSidebarCollapsed ? "RBAC" : ""} style={{ animationDelay: '350ms' }}>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>
              <span>RBAC</span>
            </a>
          )}
          <a href="#" className={`nav-item nav-item-stagger ${activeTab === 'planner' ? 'active' : ''}`} onClick={() => setActiveTab('planner')} title={isSidebarCollapsed ? "Planner" : ""} style={{ animationDelay: '375ms' }}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="3" width="7" height="7"></rect><rect x="14" y="3" width="7" height="7"></rect><rect x="14" y="14" width="7" height="7"></rect><rect x="3" y="14" width="7" height="7"></rect></svg>
            <span>Planner</span>
          </a>
          <a href="#" className={`nav-item nav-item-stagger ${activeTab === 'feedback' ? 'active' : ''}`} onClick={() => setActiveTab('feedback')} title={isSidebarCollapsed ? "Feedback" : ""} style={{ animationDelay: '380ms' }}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path><path d="M9 10h6"></path><path d="M9 14h6"></path></svg>
            <span>Feedback</span>
          </a>
        </nav>

        <div className="sidebar-footer">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.25rem 0.25rem 0.5rem' }}>
            <div className="sidebar-user">
              <div className="sidebar-user-avatar">{userRole === 'admin' ? 'A' : 'U'}</div>
              <div className="sidebar-user-info">
                <div className="sidebar-user-name">{userRole === 'admin' ? 'Admin' : 'User'}</div>
                <div className="sidebar-user-role">{userRole}</div>
              </div>
            </div>
            <button
              className="theme-toggle"
              onClick={toggleTheme}
              title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
              aria-label="Toggle theme"
            >
              {theme === 'dark' ? (
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
              ) : (
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
              )}
            </button>
          </div>
          <button
            className="nav-item"
            onClick={handleLogout}
            style={{ color: '#ef4444', width: '100%', textAlign: 'left' }}
            title={isSidebarCollapsed ? "Sign Out" : ""}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path><polyline points="16 17 21 12 16 7"></polyline><line x1="21" y1="12" x2="9" y2="12"></line></svg>
            <span>Sign Out</span>
          </button>
        </div>
      </aside>

      {/* ---- MOBILE: Bottom Tab Bar ---- */}
      <nav className="bottom-tab-bar">
        <button className={`tab-btn nav-item-stagger ${activeTab === 'dashboard' ? 'active' : ''}`} onClick={() => setActiveTab('dashboard')} style={{ animationDelay: '50ms' }}>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>
          <span>Home</span>
        </button>
        <button className={`tab-btn nav-item-stagger ${activeTab === 'shopify' ? 'active' : ''}`} onClick={() => setActiveTab('shopify')} style={{ animationDelay: '100ms' }}>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/></svg>
          <span>Orders</span>
        </button>
        <button className={`tab-btn nav-item-stagger ${activeTab === 'reports' ? 'active' : ''}`} onClick={() => setActiveTab('reports')} style={{ animationDelay: '150ms' }}>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>
          <span>GST</span>
        </button>
        <button className={`tab-btn nav-item-stagger ${activeTab === 'automation' ? 'active' : ''}`} onClick={() => setActiveTab('automation')} style={{ animationDelay: '200ms' }}>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
          <span>Auto</span>
        </button>
        <button className={`tab-btn nav-item-stagger ${activeTab === 'customers' ? 'active' : ''}`} onClick={() => setActiveTab('customers')} style={{ animationDelay: '250ms' }}>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
          <span>People</span>
        </button>
        <button className={`tab-btn nav-item-stagger ${activeTab === 'settings' ? 'active' : ''}`} onClick={() => setActiveTab('settings')} style={{ animationDelay: '300ms' }}>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1-2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
          <span>More</span>
        </button>
        <button className={`tab-btn nav-item-stagger ${activeTab === 'planner' ? 'active' : ''}`} onClick={() => setActiveTab('planner')} style={{ animationDelay: '350ms' }}>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="3" width="7" height="7"></rect><rect x="14" y="3" width="7" height="7"></rect><rect x="14" y="14" width="7" height="7"></rect><rect x="3" y="14" width="7" height="7"></rect></svg>
          <span>Plan</span>
        </button>
      </nav>

      <main className="main-content">
        {/* Mobile-only top header bar */}
        {isMobile && (
          <div style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: '1rem',
            paddingBottom: '0.75rem',
            borderBottom: '1px solid var(--border-color)',
          }}>
            <img
              src={theme === 'dark' ? fullLogoDark : fullLogo}
              alt="Mi Tech"
              style={{ width: '100px', height: 'auto', objectFit: 'contain' }}
            />
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <button
                className="theme-toggle"
                onClick={toggleTheme}
                title="Toggle theme"
                style={{ width: '36px', height: '36px', borderRadius: '8px', background: 'var(--bg-input)', display: 'flex', alignItems: 'center', justifyContent: 'center', border: '1px solid var(--border-color)' }}
              >
                {theme === 'dark' ? (
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
                ) : (
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
                )}
              </button>
              <button
                onClick={handleLogout}
                title="Sign Out"
                style={{ width: '36px', height: '36px', borderRadius: '8px', background: 'var(--bg-input)', display: 'flex', alignItems: 'center', justifyContent: 'center', border: '1px solid var(--border-color)', color: '#ef4444' }}
              >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
              </button>
            </div>
          </div>
        )}

        <header className="page-header">
          <div>
            <h1 className="page-title">{activeTab === 'dashboard' ? 'Overview' : activeTab === 'shopify' ? 'Orders' : activeTab === 'reports' ? 'GST Reports' : activeTab === 'automation' ? 'Automation Engine' : activeTab === 'communication' ? 'Communication Hub' : activeTab === 'tickets' ? 'Support Tickets' : activeTab === 'customers' ? 'Customers' : activeTab === 'marketing' ? 'Ads Intelligence' : activeTab === 'social' ? 'Social Command Center' : activeTab === 'planner' ? 'Minimalist Planner' : activeTab === 'users' ? 'User Roles' : activeTab === 'feedback' ? 'Customer Sentiment' : 'Settings'}</h1>
            <p className="page-subtitle">
              {activeTab === 'dashboard' ? "Welcome back. Here's what's happening today." : activeTab === 'reports' ? "Review your GST collection and generate filing reports." : activeTab === 'automation' ? "Manage templates, triggers, and orchestration logic." : activeTab === 'communication' ? "Active customer conversations across WhatsApp and more." : activeTab === 'tickets' ? "Track and resolve customer concerns with formal ticketing." : activeTab === 'shopify' ? "Real-time orders synced via Shopify Webhooks." : activeTab === 'customers' ? "Manage your customer list and import historical data." : activeTab === 'marketing' ? "Scale your growth with Meta Ads and performance marketing." : activeTab === 'planner' ? "High-performance Kanban board with execution analytics." : activeTab === 'users' ? "Manage system access and roles across your team." : activeTab === 'settings' ? "Manage your store data and preferences." : ""}
            </p>
          </div>
          {activeTab !== 'automation' && activeTab === 'settings' && userRole === 'admin' && (
            <div style={{display: 'flex', gap: '1rem'}}>
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
            </div>
          )}
        </header>
        
        {activeTab !== 'automation' && activeTab !== 'settings' && activeTab !== 'customers' && activeTab !== 'users' && activeTab !== 'marketing' && activeTab !== 'planner' && activeTab !== 'communication' && activeTab !== 'tickets' && activeTab !== 'feedback' && (
          <div className="date-range-header-bar" style={{ 
            display: 'flex', 
            justifyContent: 'space-between', 
            alignItems: 'center', 
            marginBottom: '2rem',
            padding: '1.25rem 1.5rem',
            background: 'var(--surface-color)',
            borderRadius: '16px',
            boxShadow: 'var(--shadow-sm)',
            border: '1px solid var(--border-color)'
          }}>
            <div>
              <h1 style={{ margin: 0, fontSize: '1.5rem', fontWeight: 800, color: 'var(--text-primary)', letterSpacing: '-0.025em' }}>
                {activeTab === 'dashboard' ? 'Business Overview' : 
                 activeTab === 'reports' ? 'GST Reports' : 
                 activeTab === 'customers' ? 'Customer Directory' : 
                 activeTab === 'marketing' ? 'Social Marketing Pulse' : 
                 activeTab === 'social' ? 'Social Channel Pulse' :
                 'Shopify Orders'}
              </h1>
              <p style={{ margin: '4px 0 0 0', color: 'var(--text-secondary)', fontSize: '0.9rem', fontWeight: 500 }}>
                {activeTab === 'dashboard' ? 'Monitor your revenue and order metrics' : 
                 activeTab === 'reports' ? 'Generate and export GST-ready reports' : 
                 activeTab === 'customers' ? 'Manage and analyze your customer base' : 
                 activeTab === 'marketing' ? 'Analyze and manage social media engagement' : 
                 activeTab === 'social' ? 'Unified management for all your social channels' :
                 'Manage your Shopify store orders'}
              </p>
            </div>
            
            <div style={{ display: 'flex', alignItems: 'center', gap: '1.5rem' }}>
              <button className="btn-secondary" style={{ padding: '0.5rem 1rem', fontSize: '0.875rem' }}>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{marginRight: '8px'}}><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
                Export Data
              </button>

              {(activeTab === 'dashboard' || activeTab === 'reports' || activeTab === 'shopify') && (
                <>
                  <div style={{ width: '1px', height: '32px', backgroundColor: 'var(--border-color)' }}></div>
                  <CustomDatePicker 
                    startDate={startDate} 
                    endDate={endDate} 
                    onDateChange={handleUpdateDateRange} 
                  />
                </>
              )}
            </div>
          </div>
        )}

        {activeTab === 'dashboard' && metrics && (
          <section className="page-enter">
            {/* Hero Row: Revenue + GST */}
            <div className="metrics-hero-grid">
              <div className="metric-card metric-card-hero">
                <div className="metric-icon metric-icon-1">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>
                </div>
                <div className="metric-label">Total Revenue</div>
                <div className="metric-value">₹{metrics?.total_revenue?.toLocaleString('en-IN', { maximumFractionDigits: 0 }) || '0'}</div>
                <div className="metric-sub">
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
                  {metrics?.total_invoices?.toLocaleString('en-IN') || '0'} invoices
                </div>
              </div>
              <div className="metric-card metric-card-hero">
                <div className="metric-icon metric-icon-2">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>
                </div>
                <div className="metric-label">Total GST Collected</div>
                <div className="metric-value">₹{metrics?.total_gst_collected?.toLocaleString('en-IN', { maximumFractionDigits: 0 }) || '0'}</div>
                <div className="gst-breakdown">
                  <span className="gst-pill">CGST ₹{metrics?.cgst_collected?.toLocaleString('en-IN', { maximumFractionDigits: 0 }) || '0'}</span>
                  <span className="gst-pill">SGST ₹{metrics?.sgst_collected?.toLocaleString('en-IN', { maximumFractionDigits: 0 }) || '0'}</span>
                  <span className="gst-pill">IGST ₹{metrics?.igst_collected?.toLocaleString('en-IN', { maximumFractionDigits: 0 }) || '0'}</span>
                </div>
              </div>
            </div>

            {/* Order Metrics Grid */}
            <div className="metrics-grid">
              <div className="metric-card">
                <div className="metric-icon metric-icon-1">
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/></svg>
                </div>
                <div className="metric-label">Total Orders</div>
                <div className="metric-value" style={{ fontSize: '1.5rem', color: 'var(--text-primary)' }}>{metrics?.total_orders?.toLocaleString() || '0'}</div>
              </div>
              <div className="metric-card">
                <div className="metric-icon metric-icon-2">
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
                </div>
                <div className="metric-label">Fulfilled</div>
                <div className="metric-value" style={{ fontSize: '1.5rem', color: 'var(--status-active)' }}>{metrics?.fulfilled_orders?.toLocaleString() || '0'}</div>
              </div>
              <div className="metric-card">
                <div className="metric-icon metric-icon-3">
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
                </div>
                <div className="metric-label">Unfulfilled</div>
                <div className="metric-value" style={{ fontSize: '1.5rem', color: '#f59e0b' }}>{metrics?.unfulfilled_orders?.toLocaleString() || '0'}</div>
              </div>
              <div className="metric-card">
                <div className="metric-icon metric-icon-4">
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
                </div>
                <div className="metric-label">Cancelled</div>
                <div className="metric-value" style={{ fontSize: '1.5rem', color: '#ef4444' }}>{metrics?.cancelled_orders?.toLocaleString() || '0'}</div>
              </div>
              <div className="metric-card">
                <div className="metric-icon metric-icon-5">
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>
                </div>
                <div className="metric-label">Avg. Order Value</div>
                <div className="metric-value" style={{ fontSize: '1.5rem', color: 'var(--text-primary)' }}>₹{metrics?.total_orders && metrics.total_orders > 0 ? Math.round(metrics.total_revenue / metrics.total_orders).toLocaleString('en-IN') : '0'}</div>
              </div>
            </div>
          </section>
        )}

        {activeTab === 'dashboard' && isLoading && (
          <section className="metrics-hero-grid">
            {[1,2].map(i => (
              <div key={i} className="metric-card metric-card-hero" style={{ minHeight: 130 }}>
                <div style={{ width: 80, height: 12, borderRadius: 6, background: 'var(--border-color)', marginBottom: 8 }} />
                <div style={{ width: 140, height: 28, borderRadius: 6, background: 'var(--border-color)' }} />
              </div>
            ))}
          </section>
        )}

        {activeTab === 'shopify' && (
          <section className="table-container">


            <div className="webhook-status-bar" style={{ display: 'flex', gap: '1rem', marginBottom: '1.5rem', padding: '1rem', backgroundColor: 'var(--surface-color)', borderRadius: '8px', border: '1px solid var(--border-color)' }}>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.025em', marginBottom: '0.25rem' }}>Webhook Status</div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <div style={{ width: '8px', height: '8px', borderRadius: '50%', backgroundColor: webhookStatus?.status === 'active' ? 'var(--color-success)' : 'var(--color-danger)' }}></div>
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
                  {appConfigs?.show_sync_button === 'true' && userRole === 'admin' && (
                    <button 
                      className="btn-primary" 
                      title="Manually fetch orders from Shopify"
                      onClick={() => setShowSyncModal(true)}
                      style={{ 
                        display: 'flex', 
                        alignItems: 'center', 
                        gap: '0.5rem', 
                        padding: '0.5rem 1rem', 
                        fontSize: '0.85rem',
                        height: '42px',
                        borderRadius: '10px'
                      }}
                    >
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                        <path d="M21 2v6h-6"></path>
                        <path d="M3 12a9 9 0 0 1 15-6.7L21 8"></path>
                        <path d="M3 22v-6h6"></path>
                        <path d="M21 12a9 9 0 0 1-15 6.7L3 16"></path>
                      </svg>
                      Sync Shopify
                    </button>
                  )}
                  <ColumnSelector
                    columns={AVAILABLE_COLUMNS}
                    visibleColumns={visibleColumns}
                    onChange={setVisibleColumns}
                    onReset={() => setVisibleColumns(DEFAULT_VISIBLE_COLUMNS)}
                  />

                </div>
              </div>

              <div className="orders-filter-bar" style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap', alignItems: 'center', backgroundColor: 'var(--bg-input)', padding: '0.5rem', borderRadius: '8px', border: '1px solid var(--border-color)' }}>
                <div style={{ flex: 1, minWidth: '200px', position: 'relative' }}>
                  <svg style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)' }} width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
                  <input 
                    type="text" 
                    placeholder="Search orders or customers..." 
                    aria-label="Search orders or customers"
                    value={search}
                    onChange={(e) => { setSearch(e.target.value); setPage(1); }}
                    style={{ 
                      paddingLeft: '2.5rem', 
                      paddingRight: search ? '2.5rem' : '1rem',
                      fontSize: '0.875rem', 
                      background: 'transparent',
                      border: 'none',
                      color: 'var(--text-primary)',
                      width: '100%'
                    }}
                  />
                  {search && (
                    <button
                      onClick={() => { setSearch(''); setPage(1); }}
                      aria-label="Clear search"
                      title="Clear search"
                      style={{
                        position: 'absolute',
                        right: '12px',
                        top: '50%',
                        transform: 'translateY(-50%)',
                        color: 'var(--text-tertiary)',
                        padding: '4px',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        borderRadius: '50%',
                        transition: 'all 0.2s',
                        cursor: 'pointer',
                        border: 'none',
                        background: 'transparent'
                      }}
                      onMouseEnter={(e) => {
                        e.currentTarget.style.color = 'var(--text-primary)';
                        e.currentTarget.style.background = 'var(--bg-hover)';
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.color = 'var(--text-tertiary)';
                        e.currentTarget.style.background = 'transparent';
                      }}
                    >
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                        <line x1="18" y1="6" x2="6" y2="18"></line>
                        <line x1="6" y1="6" x2="18" y2="18"></line>
                      </svg>
                    </button>
                  )}
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
                    style={{ padding: '0.5rem 1rem', fontSize: '0.875rem', color: 'var(--text-secondary)' }}
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
                  {visibleColumns.includes('whatsapp') && <th>WhatsApp</th>}
                  {visibleColumns.includes('gst_invoice') && <th>GST Invoice</th>}
                  {visibleColumns.includes('feedback_status') && <th>Feedback</th>}
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
                  orders.map((order, idx) => (
                    <tr 
                      key={order.id} 
                      className="table-row-stagger" 
                      style={{ animationDelay: `${idx * 20}ms` }}

                    >

                      {visibleColumns.includes('order_id') && (
                        <td>
                          <a 
                            href="#" 
                            onClick={(e) => {
                              e.preventDefault();
                              setSelectedOrderDetailsId(order.id);
                            }}
                            style={{ fontWeight: 600, color: 'var(--accent-color)', textDecoration: 'none' }}
                          >
                            {order.order_number}
                          </a>
                        </td>
                      )}
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
                            className={`badge-pill badge-pill-${order.fulfillment_status?.toLowerCase() === 'fulfilled' ? 'gray' : (order.status?.toUpperCase() === 'CANCELLED' || order.fulfillment_status?.toLowerCase() === 'cancelled' ? 'danger' : 'yellow')}`}
                            style={{ cursor: isUpdatingStatus ? 'not-allowed' : 'pointer', opacity: isUpdatingStatus && editingStatusId === order.id ? 0.7 : 1 }}
                            onClick={(e) => {
                              if (isUpdatingStatus) return;
                              e.stopPropagation();
                              setEditingStatusId(editingStatusId === order.id ? null : order.id);
                            }}
                          >
                             <span className="dot"></span> {isUpdatingStatus && editingStatusId === order.id ? 'Updating...' : (order.status?.toUpperCase() === 'CANCELLED' || order.fulfillment_status?.toLowerCase() === 'cancelled' ? 'Cancelled' : (order.fulfillment_status?.charAt(0).toUpperCase() + order.fulfillment_status?.slice(1) || 'Unfulfilled'))}
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
                        <td>
                          {order.status?.toUpperCase() === 'CANCELLED' ? (
                            <span style={{color: 'var(--text-tertiary)', fontSize: '0.8rem'}}>—</span>
                          ) : order.delivery_status && order.delivery_status !== 'pending' && order.delivery_status !== 'fulfilled' ? (
                            <div 
                              className="delivery-status-collapsed"
                              title={`${order.delivery_status.charAt(0).toUpperCase() + order.delivery_status.slice(1).replace(/_/g, ' ')} ${order.tracking_number ? `- ${order.shipping_company || 'Standard'}: ${order.tracking_number}` : ''}`}
                              onClick={(e) => {
                                if (order.tracking_number) {
                                  e.stopPropagation();
                                  setTrackingOrder(order);
                                }
                              }}
                              style={{ cursor: order.tracking_number ? 'pointer' : 'default', display: 'inline-block' }}
                            >
                              <span className="badge-pill badge-pill-info">
                                <span className="dot"></span> {order.delivery_status?.replace(/_/g, ' ').split(' ').map((word: string) => word.charAt(0).toUpperCase() + word.slice(1)).join(' ')}
                              </span>
                            </div>
                          ) : (
                            <button 
                              className="btn-secondary" 
                              onClick={(e) => { e.stopPropagation(); handleMarkAsDelivered(order.id); }}
                              disabled={isMarkingAsDelivered === order.id}
                              style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem', height: '28px', borderRadius: '6px' }}
                            >
                              {isMarkingAsDelivered === order.id ? 'Updating...' : 'Mark Delivered'}
                            </button>
                          )}
                        </td>
                      )}
                      {visibleColumns.includes('source') && (
                        <td>
                          <span 
                            style={{ 
                              background: order.source_id === 'amazon' ? 'var(--status-warning-bg)' : 
                                          order.source_id === 'pos' ? 'var(--status-active-bg)' : 'var(--bg-input)', 
                              color: order.source_id === 'amazon' ? 'var(--status-warning)' : 
                                     order.source_id === 'pos' ? 'var(--status-active)' : 'var(--text-secondary)',
                              border: `1px solid ${order.source_id === 'amazon' ? 'var(--status-warning)' : 
                                                  order.source_id === 'pos' ? 'var(--status-active)' : 'var(--border-color)'}`,
                              padding: '0.25rem 0.75rem',
                              borderRadius: '9999px',
                              fontSize: '0.75rem',
                              fontWeight: 600,
                              textTransform: 'capitalize'
                            }}
                          >
                            {order.source_id || 'Shopify'}
                          </span>
                        </td>
                      )}
                      {visibleColumns.includes('whatsapp') && (
                        <td>
                          {order.customer_phone ? (
                            <button 
                              className="btn-icon-minimal" 
                              title="Send WhatsApp Message"
                              aria-label="Send WhatsApp Message"
                              onClick={(e) => {
                                e.stopPropagation();
                                setWhatsappOrder(order);
                              }}
                              style={{ 
                                background: 'var(--accent-color)', 
                                color: 'white',
                                borderRadius: '12px', 
                                width: '32px', 
                                height: '32px', 
                                display: 'flex', 
                                alignItems: 'center', 
                                justifyContent: 'center',
                                border: '1px solid var(--border-color)',
                                cursor: 'pointer',
                                transition: 'all 0.2s',
                              }}
                              onMouseEnter={(e) => {
                                e.currentTarget.style.borderColor = 'var(--accent-color)';
                                e.currentTarget.style.color = 'var(--accent-color)';
                                e.currentTarget.style.background = 'var(--accent-subtle)';
                              }}
                              onMouseLeave={(e) => {
                                e.currentTarget.style.borderColor = 'var(--border-color)';
                                e.currentTarget.style.color = 'var(--text-secondary)';
                                e.currentTarget.style.background = 'var(--surface-color)';
                              }}
                            >
                              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                <line x1="22" y1="2" x2="11" y2="13"></line>
                                <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
                              </svg>
                            </button>
                          ) : (
                            <span style={{ color: 'var(--text-tertiary)', fontSize: '0.8rem' }}>No phone</span>
                          )}
                        </td>
                      )}
                      {visibleColumns.includes('gst_invoice') && (
                        <td style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                          <button 
                            onClick={() => handleDownloadInvoice(order.id, order.order_number)}
                            className="btn-primary" 
                            style={{fontSize: '0.8rem', padding: '0.4rem 0.8rem', display: 'inline-block', cursor: 'pointer', border: 'none'}}
                          >
                            Invoice
                          </button>
                        </td>
                      )}
                      {visibleColumns.includes('feedback_status') && (
                        <td>
                          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                            {(() => {
                              const sentAt = order.feedback_sent_at ? new Date(order.feedback_sent_at) : null;
                              const statusId = order.feedback_status_id;
                              
                              if (statusId === 3) {
                                return (
                                  <span className="badge-pill badge-pill-success">
                                    <span className="dot"></span> Received
                                  </span>
                                );
                              }
                              
                              if (statusId === 2 && sentAt) {
                                const expiryMins = parseInt(appConfigs?.feedback_link_expiry_minutes || '2880');
                                const isExpired = (new Date().getTime() - sentAt.getTime()) > (expiryMins * 60 * 1000);
                                const surveyURL = `${appConfigs?.feedback_base_url || 'https://feedback-form.millennialperfumer.in'}?o=${order.id}&p=${order.customer_phone}`;
                                
                                if (isExpired) {
                                  return (
                                    <span className="badge-pill badge-pill-danger" title={`Expired after ${expiryMins} mins`}>
                                      <span className="dot"></span> Expired
                                    </span>
                                  );
                                }
                                
                                return (
                                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                    <span className="badge-pill badge-pill-info">
                                      <span className="dot"></span> Sent
                                    </span>
                                    <button 
                                      className="btn-icon-minimal" 
                                      title="Copy Feedback Link"
                                      onClick={(e) => {
                                        e.stopPropagation();
                                        navigator.clipboard.writeText(surveyURL);
                                        const btn = e.currentTarget;
                                        const original = btn.innerHTML;
                                        btn.innerHTML = '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#10b981" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>';
                                        setTimeout(() => btn.innerHTML = original, 2000);
                                      }}
                                      style={{ 
                                        width: '24px', 
                                        height: '24px', 
                                        padding: 0, 
                                        display: 'flex', 
                                        alignItems: 'center', 
                                        justifyContent: 'center',
                                        borderRadius: '6px',
                                        background: 'var(--bg-input)',
                                        border: '1px solid var(--border-color)',
                                        cursor: 'pointer'
                                      }}
                                    >
                                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                                        <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"></path>
                                        <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"></path>
                                      </svg>
                                    </button>
                                  </div>
                                );
                              }
                              
                              return (
                                <span className="badge-pill badge-pill-gray" style={{ opacity: 0.6 }}>
                                   Pending
                                </span>
                              );
                            })()}
                          </div>
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

        <div className="tab-content-fade" key={activeTab}>
          {activeTab === 'reports' && (
            <GSTReports 
              startDate={startDate} 
              endDate={endDate} 
              fetchWithAuth={fetchWithAuth} 
              refreshTrigger={refreshTrigger}
            />
          )}

          {activeTab === 'automation' && (
            <WhatsAppAutomation 
              fetchWithAuth={fetchWithAuth} 
              startDate={startDate} 
              endDate={endDate}
              onDateChange={handleUpdateDateRange}
              refreshTrigger={refreshTrigger}
              userRole={userRole}
            />
          )}

          {activeTab === 'communication' && (
            <WhatsAppChat fetchWithAuth={fetchWithAuth} />
          )}

          {activeTab === 'tickets' && (
            <Tickets fetchWithAuth={fetchWithAuth} />
          )}

          {activeTab === 'settings' && (
            <SettingsTab 
              fetchWithAuth={fetchWithAuth}
            />
          )}

          {activeTab === 'marketing' && (
            <MarketingDashboard 
              fetchWithAuth={fetchWithAuth}
              userRole={userRole}
              onNavigate={(tab) => setActiveTab(tab)}
            />
          )}

          {activeTab === 'social' && (
            <SocialDashboard 
              fetchWithAuth={fetchWithAuth}
              userRole={userRole}
              startDate={startDate}
              endDate={endDate}
              onUpdateDateRange={handleUpdateDateRange}
            />
          )}

          {activeTab === 'feedback' && (
            <Feedback API_BASE={API_BASE} token={token} fetchWithAuth={fetchWithAuth} />
          )}

          {activeTab === 'customers' && (
            <Customers 
              fetchWithAuth={fetchWithAuth} 
              showClearButton={appSettings?.show_clear_customers_button === 'true'} 
              bulkSuffix={appConfigs?.bulk_template_suffix || '_marketing'}
              userRole={userRole}
            />
          )}



          {activeTab === 'planner' && (
            <Planner fetchWithAuth={fetchWithAuth} />
          )}

          {activeTab === 'users' && userRole === 'admin' && (
            <Users 
              fetchWithAuth={fetchWithAuth} 
            />
          )}
        </div>
      </main>
      {/* WhatsApp Modal */}
      {whatsappOrder && (
        <ManualWhatsAppModal  
          isOpen={!!whatsappOrder} 
          onClose={() => setWhatsappOrder(null)}
          orderId={whatsappOrder.id}
          orderNumber={whatsappOrder.order_number}
          customerName={whatsappOrder.customer_name}
          token={token}
        />
      )}

      {/* Tracking Modal */}
      {trackingOrder && (
        <div className="modal-overlay" onClick={() => setTrackingOrder(null)}>
          <div className="premium-modal tracking-modal" onClick={e => e.stopPropagation()}>
            <div className="modal-header-icon" style={{ background: 'linear-gradient(135deg, #10b981, #06b6d4)' }}>
              <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <rect x="1" y="3" width="15" height="13"/><polygon points="16 8 20 8 23 11 23 16 16 16 16 8"/><circle cx="5.5" cy="18.5" r="2.5"/><circle cx="18.5" cy="18.5" r="2.5"/>
              </svg>
            </div>
            
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
              <div>
                <h2>Tracking Details</h2>
                <p>Order {trackingOrder.order_number}</p>
              </div>
              <span className="badge-pill badge-pill-info" style={{ fontSize: '0.8rem', padding: '6px 14px' }}>
                <span className="dot"></span> {trackingOrder.delivery_status?.replace(/_/g, ' ')}
              </span>
            </div>

            <div className="tracking-card">
              <div className="tracking-label">Carrier</div>
              <div className="tracking-value">{trackingOrder.shipping_company || 'Standard Tracking'}</div>
              
              <div style={{ margin: '1.5rem 0' }}>
                <div className="tracking-label">Tracking Number</div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                  <a 
                    href={trackingOrder.tracking_url || '#'} 
                    target="_blank" 
                    rel="noopener noreferrer"
                    className="tracking-modal-link"
                  >
                    {trackingOrder.tracking_number}
                  </a>
                  <button 
                    className="copy-btn-minimal"
                    title="Copy tracking number"
                    aria-label="Copy tracking number"
                    onClick={(e) => {
                      navigator.clipboard.writeText(trackingOrder.tracking_number);
                      const btn = e.currentTarget;
                      const original = btn.innerHTML;
                      btn.innerHTML = '<span style="color: #10b981; font-size: 0.75rem; font-weight: 700;">Copied!</span>';
                      setTimeout(() => btn.innerHTML = original, 2000);
                    }}
                  >
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                      <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/>
                    </svg>
                  </button>
                </div>
              </div>

              <div className="modal-actions" style={{ marginTop: '2rem' }}>
                <button className="btn-secondary" onClick={() => setTrackingOrder(null)}>Close</button>
                <a 
                  href={trackingOrder.tracking_url || '#'} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="btn-primary"
                  style={{ textDecoration: 'none', textAlign: 'center', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
                >
                  Track on Official Website
                </a>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Order Details Modal */}
      {selectedOrderDetailsId && (
        <OrderDetailsModal
          isOpen={!!selectedOrderDetailsId}
          onClose={() => setSelectedOrderDetailsId(null)}
          orderId={selectedOrderDetailsId}
          fetchWithAuth={fetchWithAuth}
          userRole={userRole}
          onOrderUpdated={() => fetchDashboardData(true)}
        />
      )}
    </div>
  );
}

export default App;
