import React, { useState, useEffect, useRef } from 'react';
import { API_BASE } from './api';
import { useToast } from './ToastContext';

interface ManufacturingProduct {
  inventory_item_id: number;
  quantity_produced: number;
  add_stock: boolean;
  inventory_item?: { title: string; mi_sku: string };
}

interface ManufacturingOil {
  oil_inventory_id: number;
  quantity_grams: number;
  deduct_inventory: boolean;
  oil_inventory?: { name: string };
}

interface ManufacturingRecord {
  id: number;
  manufacturing_date: string;
  notes: string;
  oils: ManufacturingOil[];
  products: ManufacturingProduct[];
}

interface OilStock {
  id: number;
  name: string;
  inventory_item_id?: number;
}

interface Product {
  id: number;
  title: string;
  mi_sku: string;
}

interface SearchableSelectProps {
  options: { id: number; name: string; inventory_item_id?: number }[];
  value: string;
  onChange: (id: string) => void;
  placeholder: string;
}

const SearchableSelect: React.FC<SearchableSelectProps> = ({ options, value, onChange, placeholder }) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const safeOptions = Array.isArray(options) ? options : [];
  const selectedOption = safeOptions.find(o => o.id.toString() === value);
  const displayValue = isOpen ? searchTerm : (selectedOption?.name || '');

  const filteredOptions = safeOptions.filter(o => 
    o.name && o.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <div ref={dropdownRef} style={{ position: 'relative', width: '100%' }}>
      <input
        type="text"
        placeholder={placeholder}
        style={{ 
          width: '100%',
          borderRadius: '12px', 
          border: '1px solid #cbd5e1', 
          padding: '0.75rem',
          background: 'white',
          fontSize: '0.9rem',
          cursor: 'pointer'
        }}
        value={displayValue}
        onFocus={() => { setIsOpen(true); setSearchTerm(''); }}
        onChange={(e) => setSearchTerm(e.target.value)}
      />
      {isOpen && (
        <div style={{
          position: 'absolute',
          top: '110%',
          left: 0,
          right: 0,
          background: 'white',
          borderRadius: '16px',
          border: '1px solid #e2e8f0',
          boxShadow: '0 10px 25px -5px rgba(0,0,0,0.1)',
          zIndex: 1000,
          maxHeight: '300px',
          overflowY: 'auto',
          padding: '0.5rem'
        }}>
          {filteredOptions.length > 0 ? filteredOptions.map(option => (
            <div
              key={option.id}
              style={{
                padding: '0.75rem 1rem',
                borderRadius: '10px',
                cursor: 'pointer',
                fontSize: '0.85rem',
                fontWeight: 500,
                color: '#334155',
                background: value === option.id.toString() ? '#f8fafc' : 'transparent'
              }}
              onMouseDown={() => {
                onChange(option.id.toString());
                setIsOpen(false);
              }}
              onMouseEnter={(e) => e.currentTarget.style.background = '#f1f5f9'}
              onMouseLeave={(e) => e.currentTarget.style.background = value === option.id.toString() ? '#f8fafc' : 'transparent'}
            >
              {option.name}
            </div>
          )) : (
            <div style={{ padding: '1rem', textAlign: 'center', fontSize: '0.8rem', color: '#94a3b8' }}>No results found</div>
          )}
        </div>
      )}
    </div>
  );
};

export const Manufacturing: React.FC<{ token: string | null }> = ({ token }) => {
  const { success: toastSuccess, error: toastError } = useToast();
  const [records, setRecords] = useState<ManufacturingRecord[]>([]);
  const [oils, setOils] = useState<OilStock[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showAddModal, setShowAddModal] = useState(false);
  const [editingRecordId, setEditingRecordId] = useState<number | null>(null);

  const [formData, setFormData] = useState({
    notes: '',
    manufacturing_date: '',
    additions: [{ 
      oil_id: '', 
      oil_grams: '', 
      product_id: '', 
      product_qty: '',
      deduct_inventory: true,
      add_stock: true
    }]
  });

  const fetchWithAuth = async (url: string, options: RequestInit = {}) => {
    const headers = { ...options.headers, 'Authorization': `Bearer ${token}` };
    return fetch(url, { ...options, headers });
  };

  const fetchData = async () => {
    setIsLoading(true);
    try {
      const [mfgRes, oilsRes, prodRes] = await Promise.all([
        fetchWithAuth(`${API_BASE}/api/inventory/manufacturing`),
        fetchWithAuth(`${API_BASE}/api/inventory/oil`),
        fetchWithAuth(`${API_BASE}/api/inventory`)
      ]);

      if (mfgRes.ok) setRecords(await mfgRes.json());
      if (oilsRes.ok) setOils(await oilsRes.json());
      if (prodRes.ok) setProducts(await prodRes.json());
    } catch (err) {
      toastError('Failed to fetch data');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => { fetchData(); }, []);

  const addAddition = () => {
    setFormData({
      ...formData,
      additions: [...formData.additions, { oil_id: '', oil_grams: '', product_id: '', product_qty: '', deduct_inventory: true, add_stock: true }]
    });
  };

  const removeAddition = (index: number) => {
    setFormData({
      ...formData,
      additions: formData.additions.filter((_, i) => i !== index)
    });
  };

  const updateAddition = (index: number, field: string, value: any) => {
    const newAdditions = formData.additions.map((item, i) => {
      if (i !== index) return item;
      const newItem = { ...item, [field]: value };
      
      // Auto-select product if oil is selected
      if (field === 'oil_id' && value) {
        const selectedOil = oils.find(o => o.id === parseInt(value));
        if (selectedOil?.inventory_item_id) {
          newItem.product_id = selectedOil.inventory_item_id.toString();
        }
      }
      return newItem;
    });

    setFormData({ ...formData, additions: newAdditions });
  };

  const handleEdit = (r: ManufacturingRecord) => {
    setEditingRecordId(r.id);
    const mDate = (r.manufacturing_date && !r.manufacturing_date.startsWith('0001')) 
      ? r.manufacturing_date 
      : new Date().toISOString();
    
    setFormData({
      notes: r.notes || '',
      manufacturing_date: mDate,
      additions: r.oils.map((o, index) => ({
        oil_id: o.oil_inventory_id.toString(),
        oil_grams: o.quantity_grams.toString(),
        product_id: r.products[index]?.inventory_item_id.toString() || '',
        product_qty: r.products[index]?.quantity_produced.toString() || '',
        deduct_inventory: o.deduct_inventory,
        add_stock: r.products[index]?.add_stock ?? true
      }))
    });
    setShowAddModal(true);
  };

  const handleCloseModal = () => {
    setShowAddModal(false);
    setEditingRecordId(null);
    setFormData({ 
      notes: '', 
      manufacturing_date: new Date().toISOString(), 
      additions: [{ 
        oil_id: '', 
        oil_grams: '', 
        product_id: '', 
        product_qty: '', 
        deduct_inventory: true, 
        add_stock: true 
      }] 
    });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    const body = {
      id: editingRecordId,
      notes: formData.notes,
      manufacturing_date: formData.manufacturing_date ? new Date(formData.manufacturing_date).toISOString() : new Date().toISOString(),
      oils: formData.additions
        .filter(a => a.oil_id && a.oil_grams)
        .map(a => ({
          oil_inventory_id: parseInt(a.oil_id),
          quantity_grams: parseFloat(a.oil_grams),
          deduct_inventory: a.deduct_inventory
        })),
      products: formData.additions
        .filter(a => a.product_id && a.product_qty)
        .map(a => ({
          inventory_item_id: parseInt(a.product_id),
          quantity_produced: parseInt(a.product_qty),
          add_stock: a.add_stock
        }))
    };

    try {
      const method = editingRecordId ? 'PUT' : 'POST';
      const resp = await fetchWithAuth(`${API_BASE}/api/inventory/manufacturing`, {
        method: method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      });
      if (resp.ok) {
        toastSuccess(editingRecordId ? 'Manufacturing record updated' : 'Manufacturing record created');
        setShowAddModal(false);
        setEditingRecordId(null);
        setFormData({ notes: '', manufacturing_date: '', additions: [{ oil_id: '', oil_grams: '', product_id: '', product_qty: '', deduct_inventory: true, add_stock: true }] });
        fetchData();
      }
    } catch (err) {
      toastError('Error saving record');
    }
  };

  return (
    <div className="tab-content staggered-fade-in">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2.5rem' }}>
        <div>
          <h1 style={{ margin: 0, fontSize: '2.25rem', fontWeight: 900, background: 'linear-gradient(135deg, #1e293b, #475569)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent', letterSpacing: '-0.02em' }}>Manufacturing Hub</h1>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.95rem', marginTop: '0.4rem', fontWeight: 500 }}>Orchestrate production cycles and monitor oil-to-product efficiency.</p>
        </div>
        <button className="btn-primary" 
          style={{ 
            background: 'linear-gradient(135deg, var(--accent-color), #34d399)', 
            border: 'none', 
            boxShadow: '0 10px 15px -3px rgba(16, 185, 129, 0.3)',
            padding: '0.8rem 1.8rem',
            borderRadius: '14px',
            fontWeight: 700,
            display: 'flex',
            alignItems: 'center',
            gap: '8px'
          }}
          onClick={() => { 
            setEditingRecordId(null);
            setFormData({ notes: '', manufacturing_date: new Date().toISOString(), additions: [{ oil_id: '', oil_grams: '', product_id: '', product_qty: '', deduct_inventory: true, add_stock: true }] }); 
            setShowAddModal(true); 
          }}>
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
          Log Production
        </button>
      </div>

      <div className="table-container glass-card-premium" style={{ borderRadius: '24px', overflow: 'hidden', border: '1px solid rgba(255,255,255,0.6)', boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.05)' }}>
        <table className="premium-table">
          <thead style={{ background: 'rgba(248, 250, 252, 0.8)' }}>
            <tr>
              <th style={{ padding: '1.25rem 2rem', fontSize: '0.75rem', textTransform: 'uppercase', letterSpacing: '0.05em', color: '#64748b' }}>Date</th>
              <th style={{ padding: '1.25rem', fontSize: '0.75rem', textTransform: 'uppercase', letterSpacing: '0.05em', color: '#64748b' }}>Fragrance(s) Used</th>
              <th style={{ padding: '1.25rem', fontSize: '0.75rem', textTransform: 'uppercase', letterSpacing: '0.05em', color: '#64748b' }}>Oil Quantity</th>
              <th style={{ padding: '1.25rem', fontSize: '0.75rem', textTransform: 'uppercase', letterSpacing: '0.05em', color: '#64748b' }}>Products Produced</th>
              <th style={{ padding: '1.25rem 2rem', fontSize: '0.75rem', textTransform: 'uppercase', letterSpacing: '0.05em', color: '#64748b', textAlign: 'right' }}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {records.length === 0 && !isLoading ? (
              <tr>
                <td colSpan={5} style={{ textAlign: 'center', padding: '5rem', color: 'var(--text-tertiary)' }}>
                  <div style={{ opacity: 0.5, marginBottom: '1rem' }}>
                    <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path><polyline points="3.27 6.96 12 12.01 20.73 6.96"></polyline><line x1="12" y1="22.08" x2="12" y2="12"></line></svg>
                  </div>
                  No production records found.
                </td>
              </tr>
            ) : (
              records.map(r => (
                <tr key={r.id} className="hover-row" style={{ borderBottom: '1px solid #f1f5f9' }}>
                   <td style={{ padding: '1.5rem 2rem' }}>
                    <div style={{ fontWeight: 600, color: '#1e293b' }}>{new Date(r.manufacturing_date).toLocaleDateString('en-GB')}</div>
                    <div style={{ fontSize: '0.7rem', color: '#94a3b8', marginTop: '0.2rem' }}>Batch #{r.id.toString().padStart(4, '0')}</div>
                   </td>
                   <td style={{ padding: '1.5rem 1rem' }}>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                      {r.oils.map((o, i) => (
                        <div key={i} style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontWeight: 700, color: '#334155' }}>
                          <div style={{ width: '6px', height: '6px', borderRadius: '50%', background: '#f59e0b' }}></div>
                          {o.oil_inventory?.name}
                        </div>
                      ))}
                    </div>
                  </td>
                  <td style={{ padding: '1.5rem 1rem' }}>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                      {r.oils.map((o, i) => (
                        <span key={i} style={{ 
                          background: 'rgba(245, 158, 11, 0.1)', 
                          color: '#d97706', 
                          padding: '0.3rem 0.8rem', 
                          borderRadius: '8px', 
                          fontSize: '0.8rem', 
                          fontWeight: 700,
                          width: 'fit-content',
                          border: '1px solid rgba(245, 158, 11, 0.2)'
                        }}>
                          {o.quantity_grams.toLocaleString()} g
                        </span>
                      ))}
                    </div>
                  </td>
                  <td style={{ padding: '1.5rem 1rem' }}>
                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem' }}>
                      {r.products.map((p, i) => (
                        <div key={i} style={{ 
                          display: 'flex', 
                          alignItems: 'center', 
                          gap: '0.4rem', 
                          background: '#f8fafc', 
                          padding: '0.3rem 0.7rem', 
                          borderRadius: '8px', 
                          border: '1px solid #e2e8f0',
                          fontSize: '0.75rem',
                          color: '#475569'
                        }}>
                          <span style={{ fontWeight: 800, color: '#1e293b' }}>{p.quantity_produced}x</span>
                          <span style={{ color: '#94a3b8' }}>|</span>
                          <span>{p.inventory_item?.mi_sku}</span>
                        </div>
                      ))}
                    </div>
                  </td>
                   <td style={{ padding: '1.5rem 2rem', textAlign: 'right' }}>
                    <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
                      <button 
                        style={{ 
                          color: '#3b82f6', 
                          background: 'rgba(59, 130, 246, 0.05)', 
                          padding: '0.6rem', 
                          borderRadius: '12px',
                          border: 'none',
                          cursor: 'pointer',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          transition: 'all 0.2s ease'
                        }}
                        onMouseOver={e => e.currentTarget.style.background = 'rgba(59, 130, 246, 0.1)'}
                        onMouseOut={e => e.currentTarget.style.background = 'rgba(59, 130, 246, 0.05)'}
                        onClick={() => handleEdit(r)}
                      >
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
                      </button>
                      <button 
                        style={{ 
                          color: '#ef4444', 
                          background: 'rgba(239, 68, 68, 0.05)', 
                          padding: '0.6rem', 
                          borderRadius: '12px',
                          border: 'none',
                          cursor: 'pointer',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          transition: 'all 0.2s ease'
                        }}
                        onMouseOver={e => e.currentTarget.style.background = 'rgba(239, 68, 68, 0.1)'}
                        onMouseOut={e => e.currentTarget.style.background = 'rgba(239, 68, 68, 0.05)'}
                        onClick={async () => {
                          if (window.confirm('Are you sure you want to delete this production record? Inventory levels across all platforms and oil stocks will be automatically reverted.')) {
                            try {
                              const resp = await fetchWithAuth(`${API_BASE}/api/inventory/manufacturing?id=${r.id}`, { method: 'DELETE' });
                              if (resp.ok) {
                                toastSuccess('Record deleted and inventory reverted');
                                fetchData();
                              } else {
                                toastError('Failed to delete record');
                              }
                            } catch (err) {
                              toastError('Error during deletion');
                            }
                          }
                        }}
                      >
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>
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
        <div className="modal-overlay" style={{ backdropFilter: 'blur(12px)', backgroundColor: 'rgba(15, 23, 42, 0.6)', zIndex: 1000 }} onClick={handleCloseModal}>
          <div className="premium-modal" onClick={e => e.stopPropagation()} 
            style={{ 
              maxWidth: '960px', 
              width: '95%', 
              borderRadius: '32px', 
              padding: '3rem', 
              maxHeight: '90vh', 
              overflowY: 'auto',
              border: '1px solid rgba(255,255,255,0.2)',
              boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.5)',
              position: 'relative',
              background: '#ffffff'
            }}>
            
            <button onClick={handleCloseModal} style={{ position: 'absolute', top: '2rem', right: '2rem', color: '#94a3b8', background: 'none', border: 'none', cursor: 'pointer', padding: '0.5rem', borderRadius: '50%' }}>
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
            </button>

            <div style={{ display: 'flex', alignItems: 'center', gap: '1.5rem', marginBottom: '3rem' }}>
              <div style={{ 
                width: '64px', 
                height: '64px', 
                borderRadius: '20px', 
                background: 'linear-gradient(135deg, var(--accent-color), #34d399)', 
                display: 'flex', 
                alignItems: 'center', 
                justifyContent: 'center', 
                color: 'white',
                boxShadow: '0 10px 20px -5px rgba(16, 185, 129, 0.4)'
              }}>
                <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/></svg>
              </div>
              <div>
                <h2 style={{ margin: 0, fontSize: '1.85rem', fontWeight: 900, color: '#1e293b', letterSpacing: '-0.02em' }}>
                  {editingRecordId ? 'Edit Record' : 'Production Log'}
                </h2>
                <p style={{ margin: '0.4rem 0 0 0', color: '#64748b', fontSize: '0.95rem', fontWeight: 500 }}>
                  {editingRecordId ? 'Update manufacturing details. Note: Composition changes won\'t adjust stock.' : 'Select oils to auto-fill products. Inventory will be synced instantly.'}
                </p>
              </div>
            </div>

            <form onSubmit={handleSubmit}>
              <div style={{ marginBottom: '2.5rem', display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '2rem' }}>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                  <label style={{ fontSize: '0.7rem', fontWeight: 800, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.1em' }}>Production Date</label>
                  <input 
                    type="date"
                    value={formData.manufacturing_date ? formData.manufacturing_date.split('T')[0] : new Date().toISOString().split('T')[0]}
                    onChange={e => setFormData({ ...formData, manufacturing_date: new Date(e.target.value).toISOString() })}
                    style={{
                      borderRadius: '16px',
                      padding: '0.85rem 1.25rem',
                      border: '1px solid #e2e8f0',
                      fontSize: '1rem',
                      fontWeight: 600,
                      color: '#1e293b',
                      background: '#f8fafc',
                      width: '100%'
                    }}
                  />
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                   <label style={{ fontSize: '0.7rem', fontWeight: 800, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.1em' }}>Batch Identifier</label>
                   <div style={{ padding: '0.85rem 1.25rem', background: '#f1f5f9', borderRadius: '16px', color: '#64748b', fontWeight: 700, fontSize: '1rem' }}>
                     {editingRecordId ? `MAN-BATCH-${editingRecordId.toString().padStart(4, '0')}` : 'AUTO-GENERATED'}
                   </div>
                </div>
              </div>

              <div style={{ marginBottom: '2.5rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', padding: '0 0.5rem' }}>
                  <label style={{ fontSize: '0.8rem', fontWeight: 800, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.1em' }}>Batch Composition</label>
                  <button type="button" onClick={addAddition} 
                    style={{ 
                      color: 'white', 
                      background: '#1e293b',
                      padding: '0.5rem 1.25rem',
                      borderRadius: '10px',
                      fontWeight: 700, 
                      fontSize: '0.8rem', 
                      display: 'flex', 
                      alignItems: 'center', 
                      gap: '0.5rem',
                      boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)'
                    }}>
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
                    Add Row
                  </button>
                </div>
                
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
                    {formData.additions.map((a, index) => (
                      <div key={index} style={{ 
                        background: 'rgba(248, 250, 252, 0.6)', 
                        padding: '1.5rem', 
                        borderRadius: '20px', 
                        border: '1px solid #e2e8f0',
                        transition: 'all 0.25s ease',
                        position: 'relative'
                      }}>
                        {/* Row number badge */}
                        <div style={{ 
                          position: 'absolute', top: '-8px', left: '1.25rem', 
                          background: '#1e293b', color: 'white', 
                          fontSize: '0.6rem', fontWeight: 800, 
                          padding: '2px 10px', borderRadius: '6px', 
                          letterSpacing: '0.05em' 
                        }}>
                          ITEM {index + 1}
                        </div>

                        {/* Delete button */}
                        {formData.additions.length > 1 && (
                          <button type="button" onClick={() => removeAddition(index)} 
                            style={{ 
                              position: 'absolute', top: '0.75rem', right: '0.75rem',
                              color: '#cbd5e1', background: 'none', border: 'none', 
                              cursor: 'pointer', padding: '4px', borderRadius: '8px',
                              transition: 'color 0.2s ease'
                            }}
                            onMouseOver={e => e.currentTarget.style.color = '#ef4444'}
                            onMouseOut={e => e.currentTarget.style.color = '#cbd5e1'}
                          >
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                          </button>
                        )}

                        {/* --- RAW MATERIAL SECTION --- */}
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.75rem' }}>
                          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#f59e0b" strokeWidth="2.5" strokeLinecap="round"><circle cx="12" cy="12" r="10"/><path d="M12 6v6l4 2"/></svg>
                          <span style={{ fontSize: '0.65rem', fontWeight: 800, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.08em' }}>Raw Material</span>
                        </div>
                        <div style={{ display: 'grid', gridTemplateColumns: '1fr 120px', gap: '1rem', alignItems: 'end' }}>
                          <div className="input-group" style={{ marginBottom: 0 }}>
                            <label style={{ fontSize: '0.6rem', fontWeight: 800, color: '#64748b', marginBottom: '0.4rem', display: 'block', textTransform: 'uppercase', letterSpacing: '0.04em' }}>Fragrance Oil</label>
                            <SearchableSelect 
                              options={oils} 
                              value={a.oil_id} 
                              onChange={(val) => updateAddition(index, 'oil_id', val)} 
                              placeholder="Search Oil..." 
                            />
                          </div>
                          <div className="input-group" style={{ marginBottom: 0 }}>
                            <label style={{ fontSize: '0.6rem', fontWeight: 800, color: '#64748b', marginBottom: '0.4rem', display: 'block', textTransform: 'uppercase', letterSpacing: '0.04em' }}>Grams Used</label>
                            <input required type="number" step="0.1" placeholder="0.0" 
                              style={{ borderRadius: '12px', border: '1px solid #cbd5e1', padding: '0.75rem', width: '100%' }}
                              value={a.oil_grams} onChange={e => updateAddition(index, 'oil_grams', e.target.value)} />
                          </div>
                        </div>

                        {/* Deduct toggle - inline pill */}
                        <div 
                          onClick={() => updateAddition(index, 'deduct_inventory', !a.deduct_inventory)}
                          style={{ 
                            display: 'inline-flex', alignItems: 'center', gap: '8px', 
                            marginTop: '0.75rem', padding: '6px 14px 6px 8px', 
                            borderRadius: '20px', cursor: 'pointer',
                            background: a.deduct_inventory ? 'rgba(16, 185, 129, 0.08)' : 'rgba(148, 163, 184, 0.08)',
                            border: `1px solid ${a.deduct_inventory ? 'rgba(16, 185, 129, 0.25)' : 'rgba(148, 163, 184, 0.2)'}`,
                            transition: 'all 0.3s ease'
                          }}
                        >
                          <div style={{ 
                            width: '36px', height: '20px', borderRadius: '10px', position: 'relative',
                            background: a.deduct_inventory ? 'var(--accent-color)' : '#cbd5e1',
                            transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                            flexShrink: 0
                          }}>
                            <div style={{ 
                              position: 'absolute', top: '2px', 
                              left: a.deduct_inventory ? '18px' : '2px', 
                              width: '16px', height: '16px', borderRadius: '8px', 
                              background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.12)',
                              transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)'
                            }} />
                          </div>
                          <span style={{ 
                            fontSize: '0.7rem', fontWeight: 700, 
                            color: a.deduct_inventory ? '#059669' : '#94a3b8',
                            transition: 'color 0.3s ease', userSelect: 'none'
                          }}>
                            {a.deduct_inventory ? 'Deduct Oil Inventory' : 'Oil Not Deducted'}
                          </span>
                        </div>

                        {/* --- VISUAL DIVIDER --- */}
                        <div style={{ 
                          borderTop: '1px dashed #e2e8f0', 
                          margin: '1.25rem 0', 
                          position: 'relative' 
                        }}>
                          <span style={{ 
                            position: 'absolute', top: '-8px', left: '50%', transform: 'translateX(-50%)',
                            background: 'rgba(248, 250, 252, 0.6)', padding: '0 12px', 
                            fontSize: '0.55rem', fontWeight: 800, color: '#cbd5e1', 
                            textTransform: 'uppercase', letterSpacing: '0.15em'
                          }}>
                            produces
                          </span>
                        </div>

                        {/* --- OUTPUT SECTION --- */}
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.75rem' }}>
                          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#3b82f6" strokeWidth="2.5" strokeLinecap="round"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>
                          <span style={{ fontSize: '0.65rem', fontWeight: 800, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.08em' }}>Output</span>
                        </div>
                        <div style={{ display: 'grid', gridTemplateColumns: '1fr 120px', gap: '1rem', alignItems: 'end' }}>
                          <div className="input-group" style={{ marginBottom: 0 }}>
                            <label style={{ fontSize: '0.6rem', fontWeight: 800, color: '#64748b', marginBottom: '0.4rem', display: 'block', textTransform: 'uppercase', letterSpacing: '0.04em' }}>Produced Product</label>
                            <select required 
                              style={{ borderRadius: '12px', border: '1px solid #cbd5e1', padding: '0.75rem', background: a.product_id ? '#f1f5f9' : 'white', width: '100%' }}
                              value={a.product_id} onChange={e => updateAddition(index, 'product_id', e.target.value)}>
                              <option value="">Select Product...</option>
                              {Array.isArray(products) && products.map(p => <option key={p.id} value={p.id}>{p.mi_sku} - {p.title}</option>)}
                            </select>
                          </div>
                          <div className="input-group" style={{ marginBottom: 0 }}>
                            <label style={{ fontSize: '0.6rem', fontWeight: 800, color: '#64748b', marginBottom: '0.4rem', display: 'block', textTransform: 'uppercase', letterSpacing: '0.04em' }}>Qty Made</label>
                            <input required type="number" placeholder="0" 
                              style={{ borderRadius: '12px', border: '1px solid #cbd5e1', padding: '0.75rem', width: '100%' }}
                              value={a.product_qty} onChange={e => updateAddition(index, 'product_qty', e.target.value)} />
                          </div>
                        </div>

                        {/* Add Stock toggle - inline pill */}
                        <div 
                          onClick={() => updateAddition(index, 'add_stock', !a.add_stock)}
                          style={{ 
                            display: 'inline-flex', alignItems: 'center', gap: '8px', 
                            marginTop: '0.75rem', padding: '6px 14px 6px 8px', 
                            borderRadius: '20px', cursor: 'pointer',
                            background: a.add_stock ? 'rgba(59, 130, 246, 0.08)' : 'rgba(148, 163, 184, 0.08)',
                            border: `1px solid ${a.add_stock ? 'rgba(59, 130, 246, 0.25)' : 'rgba(148, 163, 184, 0.2)'}`,
                            transition: 'all 0.3s ease'
                          }}
                        >
                          <div style={{ 
                            width: '36px', height: '20px', borderRadius: '10px', position: 'relative',
                            background: a.add_stock ? '#3b82f6' : '#cbd5e1',
                            transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                            flexShrink: 0
                          }}>
                            <div style={{ 
                              position: 'absolute', top: '2px', 
                              left: a.add_stock ? '18px' : '2px', 
                              width: '16px', height: '16px', borderRadius: '8px', 
                              background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.12)',
                              transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)'
                            }} />
                          </div>
                          <span style={{ 
                            fontSize: '0.7rem', fontWeight: 700, 
                            color: a.add_stock ? '#2563eb' : '#94a3b8',
                            transition: 'color 0.3s ease', userSelect: 'none'
                          }}>
                            {a.add_stock ? 'Add to Product Stock' : 'Stock Not Updated'}
                          </span>
                        </div>
                      </div>
                    ))}
                  </div>
              </div>

              <div className="input-group" style={{ background: '#f8fafc', padding: '1.5rem', borderRadius: '20px', border: '1px solid #e2e8f0' }}>
                <label style={{ fontSize: '0.8rem', fontWeight: 800, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.1em', marginBottom: '1rem', display: 'block' }}>Production Notes</label>
                <textarea 
                  style={{ borderRadius: '14px', border: '1px solid #cbd5e1', padding: '1rem', background: 'white' }}
                  value={formData.notes} onChange={e => setFormData({ ...formData, notes: e.target.value })} placeholder="Record any observations about this production batch (e.g. ambient temperature, specific blend ratios)..." rows={3} />
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '1.25rem', marginTop: '1rem' }}>
                <button type="button" onClick={handleCloseModal} className="btn-secondary" style={{ padding: '0.8rem 2.5rem', borderRadius: '14px', fontWeight: 700 }}>Dismiss</button>
                <button type="submit" className="btn-primary" 
                  style={{ 
                    padding: '0.8rem 2.5rem', 
                    borderRadius: '14px', 
                    fontWeight: 700, 
                    background: 'linear-gradient(135deg, #10b981, #059669)',
                    border: 'none',
                    boxShadow: '0 10px 20px -5px rgba(16, 185, 129, 0.3)',
                    color: 'white',
                    cursor: 'pointer'
                  }}>
                  {editingRecordId ? 'Update Production Record' : 'Execute Production Cycle'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
