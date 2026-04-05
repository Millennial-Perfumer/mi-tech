import React, { useEffect, useState } from 'react';
import { useOrderStore } from '../store/useOrderStore';
import { Search, ChevronRight } from 'lucide-react';

export const Orders: React.FC = () => {
  const { orders, isLoading, fetchOrders } = useOrderStore();
  const [searchTerm, setSearchTerm] = useState('');

  useEffect(() => {
    fetchOrders();
  }, []);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    fetchOrders(searchTerm);
  };

  const StatusBadge = ({ status }: { status: string; type: 'financial' | 'fulfillment' }) => {
    const isReady = status === 'paid' || status === 'fulfilled';
    
    return (
      <span style={{ 
        fontSize: '0.65rem', 
        fontWeight: 700, 
        textTransform: 'uppercase', 
        padding: '4px 8px', 
        borderRadius: '6px',
        background: isReady ? 'rgba(16, 185, 129, 0.1)' : 'rgba(245, 158, 11, 0.1)',
        color: isReady ? 'var(--accent-color)' : '#f59e0b',
        border: `1px solid ${isReady ? 'rgba(16, 185, 129, 0.2)' : 'rgba(245, 158, 11, 0.2)'}`
      }}>
        {status}
      </span>
    );
  };

  return (
    <div>
      <header style={{ marginBottom: '2rem' }}>
        <p style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--accent-color)', textTransform: 'uppercase', letterSpacing: '0.15em' }}>Inventory</p>
        <h1>Orders</h1>
      </header>

      <form onSubmit={handleSearch} style={{ position: 'relative', marginBottom: '1.5rem' }}>
        <input 
          type="text" 
          placeholder="Search by ID or Customer..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          style={{
            width: '100%',
            background: 'var(--bg-input)',
            border: '1px solid var(--glass-border)',
            padding: '1rem 1rem 1rem 3rem',
            borderRadius: '14px',
            color: '#fff',
            fontSize: '0.9rem'
          }}
        />
        <Search size={18} style={{ position: 'absolute', left: '1rem', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)' }} />
      </form>

      <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
        {isLoading ? (
          <p style={{ textAlign: 'center', color: 'var(--text-tertiary)' }}>Loading orders...</p>
        ) : orders.length === 0 ? (
          <p style={{ textAlign: 'center', color: 'var(--text-tertiary)' }}>No orders found.</p>
        ) : orders.map((order) => (
          <div key={order.id} className="glass-card" style={{ padding: '1rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '0.75rem' }}>
              <div>
                <h3 style={{ fontSize: '1rem', color: '#fff' }}>#{order.order_number}</h3>
                <p style={{ fontSize: '0.85rem' }}>{order.customer_name}</p>
              </div>
              <p style={{ fontWeight: 700, color: '#fff' }}>₹{parseFloat(order.total_price).toLocaleString()}</p>
            </div>
            
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <StatusBadge status={order.financial_status} type="financial" />
                <StatusBadge status={order.fulfillment_status} type="fulfillment" />
              </div>
              <ChevronRight size={18} style={{ color: 'var(--text-tertiary)' }} />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
