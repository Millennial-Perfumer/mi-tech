import React, { useState, useEffect } from 'react';
import { API_BASE } from './api';
import { useToast } from './ToastContext';

interface Supplier {
  id: number;
  name: string;
  contact_info: string;
}

export const Suppliers: React.FC<{ token: string | null }> = ({ token }) => {
  const { success: toastSuccess, error: toastError } = useToast();
  const [suppliers, setSuppliers] = useState<Supplier[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showAddModal, setShowAddModal] = useState(false);
  const [formData, setFormData] = useState({ name: '', contact_info: '' });
  const [editingId, setEditingId] = useState<number | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  const fetchWithAuth = async (url: string, options: RequestInit = {}) => {
    const headers = {
      ...options.headers,
      'Authorization': `Bearer ${token}`
    };
    return fetch(url, { ...options, headers });
  };

  const fetchSuppliers = async () => {
    setIsLoading(true);
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/inventory/suppliers`);
      if (resp.ok) {
        const data = await resp.json();
        setSuppliers(data || []);
      }
    } catch (err) {
      toastError('Failed to fetch suppliers');
    } finally {
      setIsLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const method = editingId ? 'PUT' : 'POST';
    const body = editingId ? { ...formData, id: editingId } : formData;

    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/inventory/suppliers`, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      });
      if (resp.ok) {
        toastSuccess(`Supplier ${editingId ? 'updated' : 'created'} successfully`);
        setShowAddModal(false);
        setEditingId(null);
        setFormData({ name: '', contact_info: '' });
        fetchSuppliers();
      }
    } catch (err) {
      toastError('Error saving supplier');
    }
  };

  const handleDelete = async (id: number) => {
    if (!window.confirm('Are you sure you want to delete this supplier?')) return;
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/inventory/suppliers?id=${id}`, {
        method: 'DELETE'
      });
      if (resp.ok) {
        toastSuccess('Supplier deleted');
        fetchSuppliers();
      }
    } catch (err) {
      toastError('Error deleting supplier');
    }
  };

  useEffect(() => {
    fetchSuppliers();
  }, []);

  const filteredSuppliers = suppliers.filter(s => 
    s.name.toLowerCase().includes(searchQuery.toLowerCase()) || 
    s.contact_info.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const getInitials = (name: string) => {
    return name.split(' ').map(n => n[0]).join('').toUpperCase().substring(0, 2);
  };

  return (
    <div className="tab-content staggered-fade-in">
      <div className="section-header" style={{ marginBottom: '2rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h2 style={{ fontSize: '1.5rem', fontWeight: 700, margin: 0 }}>Suppliers</h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>Manage raw material vendors.</p>
        </div>
        <button className="btn-primary" onClick={() => { setEditingId(null); setFormData({ name: '', contact_info: '' }); setShowAddModal(true); }}>
          + Add Supplier
        </button>
      </div>
      <div style={{ marginBottom: '1.5rem', display: 'flex', gap: '1rem', alignItems: 'center' }}>
        <div style={{ position: 'relative', flex: 1, maxWidth: '400px' }}>
          <input 
            type="text" 
            placeholder="Search suppliers by name or contact..." 
            className="search-input" 
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            style={{ width: '100%', paddingLeft: '2.75rem', height: '44px', borderRadius: '12px' }}
          />
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--text-tertiary)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ position: 'absolute', left: '1rem', top: '50%', transform: 'translateY(-50%)' }}>
            <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
        </div>
      </div>

      <div className="table-container glass-card-premium">
        <table className="premium-table">
          <thead>
            <tr>
              <th style={{ paddingLeft: '2rem' }}>Name</th>
              <th>Contact Info</th>
              <th style={{ paddingRight: '2rem', textAlign: 'right' }}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredSuppliers.length === 0 && !isLoading ? (
              <tr>
                <td colSpan={3} style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-tertiary)' }}>No suppliers found.</td>
              </tr>
            ) : (
              filteredSuppliers.map(s => (
                <tr key={s.id} className="hover-row">
                  <td style={{ paddingLeft: '2rem' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                      <div style={{ 
                        width: '40px', 
                        height: '40px', 
                        borderRadius: '10px', 
                        background: 'linear-gradient(135deg, #6366f1, #4f46e5)', 
                        display: 'flex', 
                        alignItems: 'center', 
                        justifyContent: 'center',
                        color: 'white',
                        fontWeight: 700,
                        fontSize: '0.85rem'
                      }}>
                        {getInitials(s.name)}
                      </div>
                      <span style={{ fontWeight: 700, fontSize: '0.95rem' }}>{s.name}</span>
                    </div>
                  </td>
                  <td>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-secondary)', fontSize: '0.9rem' }}>
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                        <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"/>
                      </svg>
                      {s.contact_info || '—'}
                    </div>
                  </td>
                  <td style={{ paddingRight: '2rem', textAlign: 'right' }}>
                    <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.5rem' }}>
                      <button className="toolbar-btn" style={{ background: 'var(--bg-input)', width: '36px', height: '36px', borderRadius: '10px' }} onClick={() => { setEditingId(s.id); setFormData({ name: s.name, contact_info: s.contact_info }); setShowAddModal(true); }}>
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                          <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
                        </svg>
                      </button>
                      <button className="toolbar-btn" style={{ background: 'rgba(239, 68, 68, 0.08)', color: '#ef4444', width: '36px', height: '36px', borderRadius: '10px' }} onClick={() => handleDelete(s.id)}>
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                          <polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/>
                        </svg>
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {showAddModal && (
        <div className="modal-overlay" style={{ backdropFilter: 'blur(8px)', backgroundColor: 'rgba(0,0,0,0.4)' }} onClick={() => setShowAddModal(false)}>
          <div className="premium-modal" onClick={e => e.stopPropagation()} style={{ 
            maxWidth: '500px', 
            width: '95%',
            borderRadius: '24px',
            padding: '2rem',
            boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25)'
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', marginBottom: '2rem' }}>
              <div style={{ 
                width: '48px', 
                height: '48px', 
                borderRadius: '14px', 
                background: 'linear-gradient(135deg, #6366f1, #4f46e5)', 
                display: 'flex', 
                alignItems: 'center', 
                justifyContent: 'center',
                color: 'white',
                boxShadow: '0 8px 16px rgba(99, 102, 241, 0.2)'
              }}>
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>
                </svg>
              </div>
              <div>
                <h2 style={{ margin: 0, fontSize: '1.5rem', fontWeight: 800, letterSpacing: '-0.02em' }}>{editingId ? 'Edit Supplier' : 'Add Supplier'}</h2>
                <p style={{ margin: '2px 0 0 0', color: 'var(--text-secondary)', fontSize: '0.85rem' }}>Manage your material vendor information.</p>
              </div>
            </div>

            <form onSubmit={handleSubmit}>
              <div style={{ display: 'grid', gap: '1.25rem' }}>
                <div className="input-group">
                  <label style={{ fontSize: '0.7rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '0.6rem', display: 'block' }}>Supplier Name</label>
                  <div style={{ position: 'relative' }}>
                    <input 
                      required 
                      type="text" 
                      placeholder="e.g. ScentCo Global"
                      value={formData.name} 
                      onChange={e => setFormData({ ...formData, name: e.target.value })} 
                      style={{ paddingLeft: '2.75rem', height: '48px', borderRadius: '12px', background: 'var(--bg-input)', border: '1px solid var(--border-color)' }}
                    />
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--text-tertiary)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ position: 'absolute', left: '1rem', top: '50%', transform: 'translateY(-50%)' }}>
                      <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/>
                    </svg>
                  </div>
                </div>

                <div className="input-group">
                  <label style={{ fontSize: '0.7rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '0.6rem', display: 'block' }}>Contact Info / Notes</label>
                  <div style={{ position: 'relative' }}>
                    <textarea 
                      rows={4} 
                      placeholder="Address, phone, or ordering notes..."
                      value={formData.contact_info} 
                      onChange={e => setFormData({ ...formData, contact_info: e.target.value })} 
                      style={{ padding: '0.75rem 1rem 0.75rem 2.75rem', borderRadius: '12px', background: 'var(--bg-input)', border: '1px solid var(--border-color)', width: '100%', minHeight: '100px', fontSize: '0.9rem', color: 'var(--text-primary)' }}
                    />
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--text-tertiary)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ position: 'absolute', left: '1rem', top: '1rem' }}>
                      <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
                    </svg>
                  </div>
                </div>
              </div>

              <div className="modal-actions" style={{ marginTop: '2.5rem', gap: '1rem' }}>
                <button 
                  type="button" 
                  className="btn-secondary" 
                  onClick={() => setShowAddModal(false)}
                  style={{ flex: 1, height: '48px', borderRadius: '12px', fontWeight: 700 }}
                >
                  Cancel
                </button>
                <button 
                  type="submit" 
                  className="btn-primary"
                  style={{ 
                    flex: 1, 
                    height: '48px', 
                    borderRadius: '12px', 
                    fontWeight: 700,
                    background: 'linear-gradient(135deg, #6366f1, #4f46e5)',
                    border: 'none',
                    boxShadow: '0 10px 15px -3px rgba(99, 102, 241, 0.3)'
                  }}
                >
                  {editingId ? 'Update Supplier' : 'Save Supplier'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
