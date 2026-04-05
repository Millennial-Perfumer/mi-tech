import React, { useEffect, useState } from 'react';
import { useCustomerStore } from '../store/useCustomerStore';
import { Search, MessageSquare, Mail, Phone } from 'lucide-react';

export const Customers: React.FC = () => {
  const { customers, isLoading, fetchCustomers } = useCustomerStore();
  const [searchTerm, setSearchTerm] = useState('');

  useEffect(() => {
    fetchCustomers();
  }, []);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    fetchCustomers(searchTerm);
  };

  const openWhatsApp = (phone: string) => {
    const cleanPhone = phone.replace(/\D/g, '');
    window.open(`https://wa.me/${cleanPhone}`, '_blank');
  };

  return (
    <div>
      <header style={{ marginBottom: '2rem' }}>
        <p style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--accent-color)', textTransform: 'uppercase', letterSpacing: '0.15em' }}>CRM</p>
        <h1>Customers</h1>
      </header>

      <form onSubmit={handleSearch} style={{ position: 'relative', marginBottom: '1.5rem' }}>
        <input 
          type="text" 
          placeholder="Search by name, email or phone..."
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
          <p style={{ textAlign: 'center', color: 'var(--text-tertiary)' }}>Loading customers...</p>
        ) : customers.length === 0 ? (
          <p style={{ textAlign: 'center', color: 'var(--text-tertiary)' }}>No customers found.</p>
        ) : customers.map((customer) => (
          <div key={customer.id} className="glass-card" style={{ padding: '1.25rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1rem' }}>
              <div>
                <h3 style={{ fontSize: '1.1rem', color: '#fff', marginBottom: '0.25rem' }}>
                  {customer.first_name || ''} {customer.last_name || ''}
                </h3>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-tertiary)', fontSize: '0.85rem' }}>
                  <Mail size={14} />
                  <span>{customer.email || 'No email'}</span>
                </div>
              </div>
              <div style={{ textAlign: 'right' }}>
                <p style={{ fontWeight: 700, color: 'var(--accent-color)' }}>₹{parseFloat(customer.total_spent || '0').toLocaleString()}</p>
                <p style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>{customer.orders_count} orders</p>
              </div>
            </div>
            
            <div style={{ display: 'flex', gap: '0.75rem' }}>
              <button 
                onClick={() => openWhatsApp(customer.phone)}
                style={{ 
                  flex: 1,
                  background: 'rgba(37, 211, 102, 0.1)', 
                  color: '#25D366',
                  border: '1px solid rgba(37, 211, 102, 0.2)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '0.5rem',
                  borderRadius: '12px'
                }}
              >
                <MessageSquare size={18} />
                WhatsApp
              </button>
              <button 
                className="glass-panel"
                style={{ 
                  flex: 1, 
                  gap: '0.5rem',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  borderRadius: '12px'
                }}
                onClick={() => window.open(`tel:${customer.phone}`)}
              >
                <Phone size={18} />
                Call
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
