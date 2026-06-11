import React, { useState, useEffect } from 'react';
import { API_BASE } from './api';

interface OrderSummary {
  id: number;
  order_number: string;
  customer_name: string;
  created_at: string;
  total_price: string;
  financial_status: string;
  fulfillment_status: string;
  status: string;
  source_id: string;
}

interface DashboardOrdersModalProps {
  isOpen: boolean;
  onClose: () => void;
  metricLabel: string;
  startDate: string;
  endDate: string;
  selectedChannels: string[];
  token: string | null;
  onViewOrderDetails: (id: number) => void;
}

export const DashboardOrdersModal: React.FC<DashboardOrdersModalProps> = ({
  isOpen,
  onClose,
  metricLabel,
  startDate,
  endDate,
  selectedChannels,
  token,
  onViewOrderDetails
}) => {
  const [orders, setOrders] = useState<OrderSummary[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [totalCount, setTotalCount] = useState(0);
  const limit = 10;

  const fetchWithAuth = async (url: string, options: RequestInit = {}) => {
    const headers = {
      ...options.headers,
      'Authorization': `Bearer ${token}`
    };
    return fetch(url, { ...options, headers });
  };

  useEffect(() => {
    if (!isOpen) return;

    const fetchOrders = async () => {
      setIsLoading(true);
      try {
        let url = `${API_BASE}/api/orders?start_date=${startDate}&end_date=${endDate}&page=${page}&limit=${limit}&sort_by=created_at&sort_order=DESC`;
        
        if (selectedChannels.length > 0) {
          url += `&source=${selectedChannels.join(',')}`;
        }

        // Apply filters based on clicked metric card
        if (metricLabel.toLowerCase() === 'fulfilled') {
          url += '&fulfillment_status=fulfilled';
        } else if (metricLabel.toLowerCase() === 'unfulfilled') {
          url += '&fulfillment_status=unfulfilled';
        } else if (metricLabel.toLowerCase() === 'cancelled') {
          url += '&status=cancelled';
        }

        const response = await fetchWithAuth(url);
        const data = await response.json();
        if (data.success) {
          setOrders(data.orders || []);
          setTotalCount(data.total_count || 0);
        }
      } catch (err) {
        console.error('Failed to fetch dashboard orders:', err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchOrders();
  }, [isOpen, metricLabel, startDate, endDate, selectedChannels, page]);

  if (!isOpen) return null;

  const totalPages = Math.ceil(totalCount / limit);

  return (
    <div className="modal-overlay" onClick={onClose} style={{ zIndex: 1050, backdropFilter: 'blur(8px)', backgroundColor: 'rgba(0,0,0,0.4)' }}>
      <div className="premium-modal" onClick={e => e.stopPropagation()} style={{ 
        maxWidth: '850px', 
        width: '95%', 
        borderRadius: '24px',
        padding: '2rem',
        boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25)',
        position: 'relative'
      }}>
        {/* Close Button */}
        <button
          type="button"
          onClick={onClose}
          style={{ position: 'absolute', top: '1.5rem', right: '1.5rem', color: 'var(--text-tertiary)', background: 'none', border: 'none', cursor: 'pointer', padding: '0.5rem', borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
          className="hover-bg"
          aria-label="Close modal"
          title="Close modal"
        >
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
        </button>

        {/* Modal Header */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', marginBottom: '2rem' }}>
          <div style={{ 
            width: '48px', 
            height: '48px', 
            borderRadius: '14px', 
            background: 'linear-gradient(135deg, var(--accent-color), #4f46e5)', 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'center',
            color: 'white',
            boxShadow: '0 8px 16px rgba(99, 102, 241, 0.2)'
          }}>
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/>
            </svg>
          </div>
          <div>
            <h2 style={{ margin: 0, fontSize: '1.5rem', fontWeight: 800, letterSpacing: '-0.02em' }}>{metricLabel} Orders</h2>
            <p style={{ margin: '2px 0 0 0', color: 'var(--text-secondary)', fontSize: '0.85rem' }}>Viewing list of orders matching filter.</p>
          </div>
        </div>

        {/* Table Content */}
        <div className="table-container glass-card-premium" style={{ maxHeight: '450px', overflowY: 'auto', marginBottom: '1.5rem' }}>
          <table className="premium-table" style={{ width: '100%' }}>
            <thead>
              <tr>
                <th style={{ paddingLeft: '1.5rem' }}>Order Number</th>
                <th>Customer</th>
                <th>Date</th>
                <th>Amount</th>
                <th>Status</th>
                <th style={{ paddingRight: '1.5rem', textAlign: 'right' }}>Action</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr>
                  <td colSpan={6} style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-tertiary)' }}>Loading orders...</td>
                </tr>
              ) : orders.length === 0 ? (
                <tr>
                  <td colSpan={6} style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-tertiary)' }}>No orders found.</td>
                </tr>
              ) : (
                orders.map(o => (
                  <tr key={o.id} className="hover-row">
                    <td style={{ paddingLeft: '1.5rem', fontWeight: 700, color: 'var(--accent-color)' }}>
                      {o.order_number}
                    </td>
                    <td>{o.customer_name || 'N/A'}</td>
                    <td>{o.created_at && !o.created_at.startsWith('0001-01-01') && !isNaN(Date.parse(o.created_at)) ? new Date(o.created_at).toLocaleDateString() : 'N/A'}</td>
                    <td style={{ fontWeight: 800 }}>₹{o.total_price}</td>
                    <td>
                      <span className={`badge-pill badge-pill-${o.status === 'CANCELLED' || o.fulfillment_status === 'cancelled' ? 'danger' : (o.fulfillment_status === 'fulfilled' ? 'success' : 'warning')}`}>
                        <span className="dot"></span> {(o.status === 'CANCELLED' ? 'Cancelled' : o.fulfillment_status || 'Unfulfilled').toUpperCase()}
                      </span>
                    </td>
                    <td style={{ paddingRight: '1.5rem', textAlign: 'right' }}>
                      <button
                        className="btn-primary"
                        onClick={() => {
                          onClose();
                          onViewOrderDetails(o.id);
                        }}
                        style={{ padding: '0.35rem 0.8rem', fontSize: '0.75rem', borderRadius: '8px', cursor: 'pointer' }}
                      >
                        Details
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {totalPages > 1 && (
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '1rem' }}>
            <span style={{ fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
              Showing {orders.length} of {totalCount} orders
            </span>
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <button 
                className="btn-secondary" 
                disabled={page === 1 || isLoading}
                onClick={() => setPage(p => p - 1)}
                style={{ padding: '0.4rem 0.8rem', fontSize: '0.85rem' }}
              >
                Previous
              </button>
              <button 
                className="btn-secondary" 
                disabled={page === totalPages || isLoading}
                onClick={() => setPage(p => p + 1)}
                style={{ padding: '0.4rem 0.8rem', fontSize: '0.85rem' }}
              >
                Next
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
