import React, { useState, useEffect, useMemo } from 'react';
import { API_BASE } from './api';
import { useToast } from './ToastContext';

interface OilStock {
  id: number;
  name: string;
  inventory_item_id?: number;
  purchase_price_per_kg: number | null;
  grams_left: number | null;
  supplier_id?: number | null;
  inventory_item?: { title: string; mi_sku: string };
  supplier?: { name: string };
}

interface Product {
  id: number;
  title: string;
  mi_sku: string;
}

interface Supplier {
  id: number;
  name: string;
}

const getInitials = (name: string) => {
  if (!name) return '??';
  const parts = name.split(/[\s-]+/).filter(p => p.length > 0);
  if (parts.length >= 2) {
    return (parts[0][0] + parts[1][0]).toUpperCase();
  }
  return parts[0].slice(0, 2).toUpperCase();
};

const SupplierAvatar: React.FC<{ name: string }> = ({ name }) => {
  const initials = getInitials(name);
  return (
    <div style={{ 
      width: '40px', 
      height: '40px', 
      borderRadius: '12px', 
      background: 'linear-gradient(135deg, #6366f1, #4f46e5)', 
      display: 'flex', 
      alignItems: 'center', 
      justifyContent: 'center', 
      color: 'white', 
      fontWeight: 800, 
      fontSize: '0.85rem',
      boxShadow: '0 4px 10px rgba(99, 102, 241, 0.2)',
      flexShrink: 0
    }}>
      {initials}
    </div>
  );
};

export const OilInventory: React.FC<{ token: string | null }> = ({ token }) => {
  const { success: toastSuccess, error: toastError } = useToast();
  const [oils, setOils] = useState<OilStock[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [suppliers, setSuppliers] = useState<Supplier[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showAddModal, setShowAddModal] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  const [formData, setFormData] = useState({
    name: '',
    inventory_item_id: '',
    purchase_price_per_kg: '' as string | number,
    grams_left: '' as string | number,
    supplier_id: ''
  });

  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
  const [sortConfig, setSortConfig] = useState<{ key: keyof OilStock | 'supplier.name' | 'inventory_item.title' | 'inventory_item.mi_sku'; direction: 'asc' | 'desc' } | null>({ key: 'name', direction: 'asc' });

  const fetchWithAuth = async (url: string, options: RequestInit = {}) => {
    const headers = {
      ...options.headers,
      'Authorization': `Bearer ${token}`
    };
    return fetch(url, { ...options, headers });
  };

  const fetchData = async () => {
    setIsLoading(true);
    try {
      const [oilsRes, prodRes, suppRes] = await Promise.all([
        fetchWithAuth(`${API_BASE}/api/inventory/oil`),
        fetchWithAuth(`${API_BASE}/api/inventory`),
        fetchWithAuth(`${API_BASE}/api/inventory/suppliers`)
      ]);

      if (oilsRes.ok) setOils(await oilsRes.json());
      if (prodRes.ok) setProducts(await prodRes.json());
      if (suppRes.ok) setSuppliers(await suppRes.json());
    } catch (err) {
      toastError('Failed to fetch inventory data');
    } finally {
      setIsLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const method = editingId ? 'PUT' : 'POST';
    const body = {
      ...formData,
      id: editingId,
      inventory_item_id: formData.inventory_item_id ? parseInt(formData.inventory_item_id) : null,
      supplier_id: formData.supplier_id ? parseInt(formData.supplier_id) : null,
      purchase_price_per_kg: formData.purchase_price_per_kg !== '' ? parseFloat(formData.purchase_price_per_kg.toString()) : null,
      grams_left: formData.grams_left !== '' ? parseFloat(formData.grams_left.toString()) : null
    };

    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/inventory/oil`, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      });
      if (resp.ok) {
        toastSuccess(`Oil ${editingId ? 'updated' : 'created'}`);
        setShowAddModal(false);
        fetchData();
      }
    } catch (err) {
      toastError('Error saving oil stock');
    }
  };

  const handleDelete = async (id: number) => {
    if (!window.confirm('Delete this oil stock record?')) return;
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/inventory/oil?id=${id}`, { method: 'DELETE' });
      if (resp.ok) {
        toastSuccess('Oil record deleted');
        fetchData();
      }
    } catch (err) {
      toastError('Error deleting record');
    }
  };

  const toggleSort = (key: any) => {
    setSortConfig(current => {
      if (current && current.key === key) {
        return { key, direction: current.direction === 'asc' ? 'desc' : 'asc' };
      }
      return { key, direction: 'asc' };
    });
  };

  const toggleSelectAll = () => {
    if (selectedIds.size === filteredOils.length) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(filteredOils.map(o => o.id)));
    }
  };

  const toggleSelectOne = (id: number) => {
    setSelectedIds(prev => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const handleBulkSupplierUpdate = async (supplierId: string) => {
    if (!supplierId || selectedIds.size === 0) return;
    if (!window.confirm(`Update supplier for ${selectedIds.size} selected oils?`)) return;

    setIsLoading(true);
    try {
      const updates = Array.from(selectedIds).map(id => {
        const oil = oils.find(o => o.id === id);
        if (!oil) return Promise.resolve();
        return fetchWithAuth(`${API_BASE}/api/inventory/oil`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ ...oil, supplier_id: parseInt(supplierId) })
        });
      });
      await Promise.all(updates);
      toastSuccess('Bulk update complete');
      setSelectedIds(new Set());
      fetchData();
    } catch (err) {
      toastError('Error during bulk update');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const filteredOils = useMemo(() => {
    let result = [...oils];
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      result = result.filter(o => 
        o.name.toLowerCase().includes(q) || 
        o.inventory_item?.title.toLowerCase().includes(q) ||
        o.supplier?.name.toLowerCase().includes(q)
      );
    }

    if (sortConfig) {
      result.sort((a, b) => {
        let aVal: any = a[sortConfig.key as keyof OilStock];
        let bVal: any = b[sortConfig.key as keyof OilStock];

        if (sortConfig.key === 'supplier.name') {
          aVal = a.supplier?.name || '';
          bVal = b.supplier?.name || '';
        } else if (sortConfig.key === 'inventory_item.title') {
          aVal = a.inventory_item?.title || '';
          bVal = b.inventory_item?.title || '';
        } else if (sortConfig.key === 'inventory_item.mi_sku') {
          aVal = a.inventory_item?.mi_sku || '';
          bVal = b.inventory_item?.mi_sku || '';
        }

        if (aVal === bVal) return 0;
        const comparison = (aVal || '') < (bVal || '') ? -1 : 1;
        return sortConfig.direction === 'asc' ? comparison : -comparison;
      });
    }
    return result;
  }, [oils, searchQuery, sortConfig]);

  return (
    <div className="tab-content staggered-fade-in">
      <div className="section-header" style={{ marginBottom: '2rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h2 style={{ fontSize: '1.5rem', fontWeight: 700, margin: 0 }}>Oil Inventory</h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>Track raw perfume oils and purchase history.</p>
        </div>
        <button className="btn-primary" onClick={() => { 
          setEditingId(null); 
          setFormData({ name: '', inventory_item_id: '', purchase_price_per_kg: '', grams_left: '', supplier_id: '' }); 
          setShowAddModal(true); 
        }}>
          + Add Oil Stock
        </button>
      </div>

      <div style={{ marginBottom: '1.5rem' }}>
        <input 
          type="text" 
          placeholder="Search by oil name, product, or supplier..." 
          className="search-input" 
          value={searchQuery}
          onChange={e => setSearchQuery(e.target.value)}
          style={{ width: '100%', maxWidth: '400px' }}
        />
      </div>

      <div className="table-container glass-card-premium">
        <table className="premium-table">
          <thead>
            <tr>
              <th style={{ width: '50px', paddingLeft: '2rem' }}>
                <input 
                  type="checkbox" 
                  checked={selectedIds.size === filteredOils.length && filteredOils.length > 0} 
                  onChange={toggleSelectAll} 
                />
              </th>
              <th onClick={() => toggleSort('name')} style={{ cursor: 'pointer' }}>
                Oil Name {sortConfig?.key === 'name' && (sortConfig.direction === 'asc' ? '↑' : '↓')}
              </th>
              <th onClick={() => toggleSort('inventory_item.mi_sku')} style={{ cursor: 'pointer' }}>
                SKU {sortConfig?.key === 'inventory_item.mi_sku' && (sortConfig.direction === 'asc' ? '↑' : '↓')}
              </th>
              <th onClick={() => toggleSort('inventory_item.title')} style={{ cursor: 'pointer' }}>
                Linked Product {sortConfig?.key === 'inventory_item.title' && (sortConfig.direction === 'asc' ? '↑' : '↓')}
              </th>
              <th onClick={() => toggleSort('supplier.name')} style={{ cursor: 'pointer' }}>
                Supplier {sortConfig?.key === 'supplier.name' && (sortConfig.direction === 'asc' ? '↑' : '↓')}
              </th>
              <th onClick={() => toggleSort('purchase_price_per_kg')} style={{ cursor: 'pointer' }}>
                Price/kg {sortConfig?.key === 'purchase_price_per_kg' && (sortConfig.direction === 'asc' ? '↑' : '↓')}
              </th>
              <th onClick={() => toggleSort('grams_left')} style={{ cursor: 'pointer' }}>
                Stock {sortConfig?.key === 'grams_left' && (sortConfig.direction === 'asc' ? '↑' : '↓')}
              </th>
              <th style={{ paddingRight: '2rem', textAlign: 'right' }}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredOils.length === 0 && !isLoading ? (
              <tr>
                <td colSpan={7} style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-tertiary)' }}>No oil inventory found.</td>
              </tr>
            ) : (
              filteredOils.map(o => (
                <tr key={o.id} className={`hover-row ${selectedIds.has(o.id) ? 'selected-row' : ''}`}>
                  <td style={{ paddingLeft: '2rem' }}>
                    <input 
                      type="checkbox" 
                      checked={selectedIds.has(o.id)} 
                      onChange={() => toggleSelectOne(o.id)} 
                    />
                  </td>
                  <td style={{ padding: '1.25rem 0' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                      <div style={{ 
                        width: '40px', height: '40px', borderRadius: '12px', 
                        background: 'rgba(16, 185, 129, 0.1)', color: 'var(--accent-color)',
                        display: 'flex', alignItems: 'center', justifyContent: 'center'
                      }}>
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M20.24 12.24a6 6 0 0 0-8.49-8.49L5 10.5V19h8.5z"/><line x1="16" y1="8" x2="2" y2="22"/><line x1="17.5" y1="15" x2="9" y2="15"/></svg>
                      </div>
                      <div>
                        <div style={{ fontWeight: 800, fontSize: '0.95rem', color: 'var(--text-primary)' }}>{o.name}</div>
                        <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', marginTop: '2px' }}>ID: #{o.id.toString().padStart(3, '0')}</div>
                      </div>
                    </div>
                  </td>
                  <td style={{ padding: '1.25rem 1rem' }}>
                    <div className="badge-pill" style={{ 
                      background: 'var(--bg-card)', 
                      border: '1px solid var(--border-color)',
                      color: 'var(--text-primary)',
                      fontWeight: 800,
                      fontSize: '0.75rem',
                      letterSpacing: '0.02em',
                      padding: '0.3rem 0.6rem'
                    }}>
                      {o.inventory_item?.mi_sku || '—'}
                    </div>
                  </td>
                  <td style={{ padding: '1.25rem 1rem' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-secondary)' }}>
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="m7.5 4.27 9 5.15"/><path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z"/><path d="m3.3 7 8.7 5 8.7-5"/><path d="M12 22V12"/></svg>
                      {o.inventory_item?.title || '—'}
                    </div>
                  </td>
                  <td style={{ padding: '1.25rem 1rem' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                      <SupplierAvatar name={o.supplier?.name || '??'} />
                      <div style={{ 
                        background: 'rgba(0,0,0,0.03)', 
                        border: '1px solid var(--border-color)', 
                        padding: '0.4rem 0.8rem',
                        borderRadius: '10px',
                        fontSize: '0.85rem',
                        fontWeight: 600,
                        color: 'var(--text-primary)',
                        maxWidth: '180px',
                        whiteSpace: 'nowrap',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis'
                      }}>
                        {o.supplier?.name || '—'}
                      </div>
                    </div>
                  </td>
                  <td style={{ padding: '1.25rem 1rem' }}>
                    <div style={{ fontWeight: 800, color: 'var(--accent-color)', fontSize: '1rem' }}>
                      {o.purchase_price_per_kg !== null ? `₹${o.purchase_price_per_kg.toLocaleString()}` : '—'}
                    </div>
                  </td>
                  <td style={{ padding: '1.25rem 1rem' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                      <div style={{ 
                        width: '10px', height: '10px', borderRadius: '50%', 
                        background: o.grams_left === null ? 'var(--text-tertiary)' : o.grams_left > 500 ? 'var(--status-active)' : o.grams_left > 100 ? 'var(--status-warning)' : 'var(--status-danger)',
                        boxShadow: `0 0 8px ${o.grams_left === null ? 'transparent' : o.grams_left > 500 ? 'var(--status-active)' : o.grams_left > 100 ? 'var(--status-warning)' : 'var(--status-danger)'}`
                      }} />
                      <span style={{ fontWeight: 800, fontSize: '0.95rem' }}>{o.grams_left !== null ? `${o.grams_left.toLocaleString()} g` : '—'}</span>
                    </div>
                  </td>
                  <td style={{ padding: '1.25rem 2rem', textAlign: 'right' }}>
                    <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
                      <button 
                        className="icon-btn-premium" 
                        title="Edit Oil Stock"
                        onClick={() => {
                          setEditingId(o.id);
                          setFormData({
                            name: o.name,
                            inventory_item_id: o.inventory_item_id?.toString() || '',
                            purchase_price_per_kg: o.purchase_price_per_kg !== null ? o.purchase_price_per_kg : '',
                            grams_left: o.grams_left !== null ? o.grams_left : '',
                            supplier_id: o.supplier_id?.toString() || ''
                          });
                          setShowAddModal(true);
                        }}
                        style={{ 
                          width: '36px', height: '36px', borderRadius: '10px', 
                          display: 'flex', alignItems: 'center', justifyContent: 'center',
                          background: 'rgba(99, 102, 241, 0.05)', color: 'var(--accent-indigo)',
                          border: '1px solid rgba(99, 102, 241, 0.1)', cursor: 'pointer', transition: 'all 0.2s'
                        }}
                      >
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                      </button>
                      <button 
                        className="icon-btn-premium delete" 
                        title="Delete Record"
                        onClick={() => handleDelete(o.id)}
                        style={{ 
                          width: '36px', height: '36px', borderRadius: '10px', 
                          display: 'flex', alignItems: 'center', justifyContent: 'center',
                          background: 'rgba(239, 68, 68, 0.05)', color: 'var(--status-danger)',
                          border: '1px solid rgba(239, 68, 68, 0.1)', cursor: 'pointer', transition: 'all 0.2s'
                        }}
                      >
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/></svg>
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {selectedIds.size > 0 && (
        <div className="bulk-actions-toolbar glass-card-premium animate-slide-up" style={{
          position: 'fixed',
          bottom: '2rem',
          left: '50%',
          transform: 'translateX(-50%)',
          padding: '1rem 2rem',
          display: 'flex',
          alignItems: 'center',
          gap: '2rem',
          zIndex: 1000,
          border: '1px solid var(--accent-color)',
          boxShadow: '0 20px 40px rgba(16, 185, 129, 0.2)',
          borderRadius: '20px',
          background: 'rgba(255, 255, 255, 0.95)',
          backdropFilter: 'blur(20px)'
        }}>
          <div style={{ color: 'var(--accent-color)', fontWeight: 800 }}>
            {selectedIds.size} Oils Selected
          </div>
          <div style={{ height: '24px', width: '1px', background: 'var(--border-color)' }} />
          <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
            <span style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-secondary)' }}>Bulk Update Supplier:</span>
            <select 
              className="search-input" 
              style={{ width: '200px', height: '36px', padding: '0 1rem' }}
              onChange={(e) => handleBulkSupplierUpdate(e.target.value)}
              value=""
            >
              <option value="">Choose Supplier...</option>
              {suppliers.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
            </select>
          </div>
          <button 
            className="toolbar-btn" 
            style={{ color: 'var(--status-danger)' }}
            onClick={() => {
              if (window.confirm(`Delete ${selectedIds.size} selected oils?`)) {
                Promise.all(Array.from(selectedIds).map(id => handleDelete(id))).then(() => setSelectedIds(new Set()));
              }
            }}
          >
            Bulk Delete
          </button>
          <button className="btn-secondary" style={{ padding: '0.4rem 1rem' }} onClick={() => setSelectedIds(new Set())}>Deselect All</button>
        </div>
      )}

      {showAddModal && (
        <div className="modal-overlay" style={{ backdropFilter: 'blur(8px)', backgroundColor: 'rgba(0,0,0,0.4)' }} onClick={() => setShowAddModal(false)}>
          <div className="premium-modal" onClick={e => e.stopPropagation()} style={{ 
            maxWidth: '550px', 
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
                background: 'linear-gradient(135deg, var(--accent-color), #059669)', 
                display: 'flex', 
                alignItems: 'center', 
                justifyContent: 'center',
                color: 'white',
                boxShadow: '0 8px 16px rgba(16, 185, 129, 0.2)'
              }}>
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/>
                </svg>
              </div>
              <div>
                <h2 style={{ margin: 0, fontSize: '1.5rem', fontWeight: 800, letterSpacing: '-0.02em' }}>{editingId ? 'Edit Oil Stock' : 'Add Oil Stock'}</h2>
                <p style={{ margin: '2px 0 0 0', color: 'var(--text-secondary)', fontSize: '0.85rem' }}>Update your raw material inventory levels.</p>
              </div>
            </div>

            <form onSubmit={handleSubmit}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.25rem' }}>
                <div className="input-group" style={{ gridColumn: 'span 2' }}>
                  <label style={{ fontSize: '0.7rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '0.6rem', display: 'block' }}>Oil Name</label>
                  <div style={{ position: 'relative' }}>
                    <input 
                      required 
                      type="text" 
                      placeholder="e.g. Ocean Drift Concentrate"
                      value={formData.name} 
                      onChange={e => setFormData({ ...formData, name: e.target.value })} 
                      style={{ paddingLeft: '2.75rem', height: '48px', borderRadius: '12px', background: 'var(--bg-input)', border: '1px solid var(--border-color)' }}
                    />
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--text-tertiary)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ position: 'absolute', left: '1rem', top: '50%', transform: 'translateY(-50%)' }}>
                      <path d="M20.24 12.24a6 6 0 0 0-8.49-8.49L5 10.5V19h8.5z"/><line x1="16" y1="8" x2="2" y2="22"/><line x1="17.5" y1="15" x2="9" y2="15"/>
                    </svg>
                  </div>
                </div>

                <div className="input-group">
                  <label style={{ fontSize: '0.7rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '0.6rem', display: 'block' }}>Linked Product</label>
                  <div style={{ position: 'relative' }}>
                    <select 
                      value={formData.inventory_item_id} 
                      onChange={e => setFormData({ ...formData, inventory_item_id: e.target.value })}
                      style={{ paddingLeft: '2.75rem', height: '48px', borderRadius: '12px', background: 'var(--bg-input)', border: '1px solid var(--border-color)' }}
                    >
                      <option value="">Select Product...</option>
                      {products.map(p => <option key={p.id} value={p.id}>{p.mi_sku} - {p.title}</option>)}
                    </select>
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--text-tertiary)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ position: 'absolute', left: '1rem', top: '50%', transform: 'translateY(-50%)', zIndex: 1 }}>
                      <path d="m7.5 4.27 9 5.15"/><path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z"/><path d="m3.3 7 8.7 5 8.7-5"/><path d="M12 22V12"/>
                    </svg>
                  </div>
                </div>

                <div className="input-group">
                  <label style={{ fontSize: '0.7rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '0.6rem', display: 'block' }}>Supplier</label>
                  <div style={{ position: 'relative' }}>
                    <select 
                      value={formData.supplier_id} 
                      onChange={e => setFormData({ ...formData, supplier_id: e.target.value })}
                      style={{ paddingLeft: '2.75rem', height: '48px', borderRadius: '12px', background: 'var(--bg-input)', border: '1px solid var(--border-color)' }}
                    >
                      <option value="">Select Supplier...</option>
                      {suppliers.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
                    </select>
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--text-tertiary)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ position: 'absolute', left: '1rem', top: '50%', transform: 'translateY(-50%)', zIndex: 1 }}>
                      <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>
                    </svg>
                  </div>
                </div>

                <div className="input-group">
                  <label style={{ fontSize: '0.7rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '0.6rem', display: 'block' }}>Purchase Price (per kg)</label>
                  <div style={{ position: 'relative' }}>
                    <input 
                      type="number" 
                      step="0.01" 
                      placeholder="0.00"
                      value={formData.purchase_price_per_kg} 
                      onChange={e => setFormData({ ...formData, purchase_price_per_kg: e.target.value })} 
                      style={{ paddingLeft: '2.75rem', height: '48px', borderRadius: '12px', background: 'var(--bg-input)', border: '1px solid var(--border-color)' }}
                    />
                    <div style={{ position: 'absolute', left: '1rem', top: '50%', transform: 'translateY(-50%)', fontWeight: 700, color: 'var(--text-tertiary)', fontSize: '0.9rem' }}>₹</div>
                  </div>
                </div>

                <div className="input-group">
                  <label style={{ fontSize: '0.7rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '0.6rem', display: 'block' }}>Grams Left</label>
                  <div style={{ position: 'relative' }}>
                    <input 
                      type="number" 
                      step="0.1" 
                      placeholder="0"
                      value={formData.grams_left} 
                      onChange={e => setFormData({ ...formData, grams_left: e.target.value })} 
                      style={{ paddingLeft: '2.75rem', height: '48px', borderRadius: '12px', background: 'var(--bg-input)', border: '1px solid var(--border-color)' }}
                    />
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--text-tertiary)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ position: 'absolute', left: '1rem', top: '50%', transform: 'translateY(-50%)' }}>
                      <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
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
                    background: 'linear-gradient(135deg, var(--accent-color), #059669)',
                    border: 'none',
                    boxShadow: '0 10px 15px -3px rgba(16, 185, 129, 0.3)'
                  }}
                >
                  {editingId ? 'Update Record' : 'Save Record'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
