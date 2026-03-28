import { useState, useEffect, useCallback } from 'react';
import { API_BASE, fetchWithAuth } from '../api';
import { MobileCard } from '../components/MobileCard';
import { BottomSheet } from '../components/BottomSheet';
import './Orders.css';

interface Order {
  id: string | number;
  order_number: string;
  total_price: string;
  created_at: string;
  customer_name: string;
  financial_status: string;
  fulfillment_status: string;
  delivery_status: string;
  tracking_number: string;
}

export const Orders: React.FC = () => {
  const [orders, setOrders] = useState<Order[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isFilterOpen, setIsFilterOpen] = useState(false);
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [page, setPage] = useState(1);
  const [totalCount, setTotalCount] = useState(0);

  const fetchOrders = useCallback(async (reset = false) => {
    setIsLoading(true);
    try {
      const currentPage = reset ? 1 : page;
      const res = await fetchWithAuth(`${API_BASE}/api/orders?page=${currentPage}&limit=10&search=${search}&fulfillment_status=${statusFilter}`);
      const data = await res.json();
      if (data.success) {
        if (reset) {
          setOrders(data.orders || []);
        } else {
          setOrders(prev => [...prev, ...(data.orders || [])]);
        }
        setTotalCount(data.total_count || 0);
      }
    } catch (err) {
      console.error('Error fetching orders:', err);
    } finally {
      setIsLoading(false);
    }
  }, [page, search, statusFilter]);

  useEffect(() => {
    void fetchOrders(true);
  }, [fetchOrders]);

  const loadMore = () => {
    if (!isLoading && orders.length < totalCount) {
      setPage(prev => prev + 1);
    }
  };

  const handleStatusUpdate = async (id: string | number, newStatus: string) => {
    try {
      const res = await fetchWithAuth(`${API_BASE}/api/orders/status?id=${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status: newStatus })
      });
      const data = await res.json();
      if (data.success) {
        setOrders(prev => prev.map(o => o.id === id ? { ...o, fulfillment_status: newStatus } : o));
      }
    } catch (err) {
      console.error('Status update failed:', err);
    }
  };

  const copyTracking = (tracking: string) => {
    if (tracking) {
      navigator.clipboard.writeText(tracking);
      alert('Tracking copied!');
    }
  };

  return (
    <div className="orders-container">
      <div className="orders-header">
        <h2 className="section-title">Orders</h2>
        <button className="icon-btn-text" onClick={() => setIsFilterOpen(true)}>
          🔍 Filters
        </button>
      </div>

      <div className="orders-list">
        {orders.length === 0 && !isLoading ? (
          <div className="empty-state-container">
            <div className="empty-icon">📦</div>
            <p>No orders found matching your criteria.</p>
          </div>
        ) : (
          orders.map(order => (
            <OrderCard
              key={order.id}
              order={order}
              onCopyTracking={copyTracking}
              onUpdateStatus={handleStatusUpdate}
            />
          ))
        )}
        {isLoading && <div className="loader">Loading...</div>}
        {!isLoading && orders.length > 0 && orders.length < totalCount && (
          <button className="secondary-btn full-width" onClick={loadMore}>Load More</button>
        )}
      </div>

      <BottomSheet isOpen={isFilterOpen} onClose={() => setIsFilterOpen(false)} title="Filter Orders">
        <div className="filter-form">
          <div className="form-group">
            <label>Search</label>
            <input
              type="text"
              placeholder="Order ID, Name, Phone..."
              value={search}
              onChange={e => setSearch(e.target.value)}
            />
          </div>
          <div className="form-group">
            <label>Fulfillment Status</label>
            <select value={statusFilter} onChange={e => setStatusFilter(e.target.value)}>
              <option value="">All</option>
              <option value="fulfilled">Fulfilled</option>
              <option value="unfulfilled">Unfulfilled</option>
              <option value="cancelled">Cancelled</option>
            </select>
          </div>
          <button className="primary-btn full-width" onClick={() => setIsFilterOpen(false)}>
            Apply Filters
          </button>
        </div>
      </BottomSheet>
    </div>
  );
};

interface OrderCardProps {
  order: Order;
  onCopyTracking: (tracking: string) => void;
  onUpdateStatus: (id: string | number, status: string) => void;
}

const OrderCard: React.FC<OrderCardProps> = ({ order, onCopyTracking, onUpdateStatus }) => {
  const [startX, setStartX] = useState<number | null>(null);
  const [currentX, setCurrentX] = useState(0);

  const handleTouchStart = (e: React.TouchEvent) => {
    setStartX(e.touches[0].clientX);
  };

  const handleTouchMove = (e: React.TouchEvent) => {
    if (startX === null) return;
    const diff = e.touches[0].clientX - startX;
    // Limit swipe
    if (Math.abs(diff) < 100) {
      setCurrentX(diff);
    }
  };

  const handleTouchEnd = () => {
    if (currentX > 60) {
      // Swipe Right -> Copy Tracking
      onCopyTracking(order.tracking_number);
    } else if (currentX < -60) {
      // Swipe Left -> Quick Status Update (Toggle fulfilled/unfulfilled)
      const nextStatus = order.fulfillment_status === 'fulfilled' ? 'unfulfilled' : 'fulfilled';
      onUpdateStatus(order.id, nextStatus);
    }
    setCurrentX(0);
    setStartX(null);
  };

  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'fulfilled': return '#10b981';
      case 'unfulfilled': return '#f59e0b';
      case 'cancelled': return '#ef4444';
      default: return '#64748b';
    }
  };

  return (
    <div className="swipe-container">
      <div className="swipe-action left" style={{ opacity: currentX > 0 ? 1 : 0 }}>📋 Copy Tracking</div>
      <div className="swipe-action right" style={{ opacity: currentX < 0 ? 1 : 0 }}>🔄 Update Status</div>
      <div
        className="swipeable-card"
        style={{ transform: `translateX(${currentX}px)` }}
        onTouchStart={handleTouchStart}
        onTouchMove={handleTouchMove}
        onTouchEnd={handleTouchEnd}
      >
        <MobileCard>
          <div className="order-card-header">
            <span className="order-id">{order.order_number}</span>
            <span className="order-amount">₹{order.total_price}</span>
          </div>
          <div className="order-card-body">
            <span className="customer-name">{order.customer_name}</span>
            <div className="order-meta">
              <span className="order-date">{new Date(order.created_at).toLocaleDateString()}</span>
              <span className="status-pill" style={{ backgroundColor: getStatusColor(order.fulfillment_status) + '20', color: getStatusColor(order.fulfillment_status) }}>
                {order.fulfillment_status || 'Unfulfilled'}
              </span>
            </div>
          </div>
        </MobileCard>
      </div>
    </div>
  );
};
