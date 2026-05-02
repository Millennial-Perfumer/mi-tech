import React, { useState, useEffect, useRef } from 'react';
import { API_BASE } from './api';
import { useToast } from './ToastContext';

interface ManufacturingProduct {
  inventory_item_id: number;
  quantity_produced: number;
  inventory_item?: { title: string; mi_sku: string };
}

interface ManufacturingOil {
  oil_inventory_id: number;
  quantity_grams: number;
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

  const [formData, setFormData] = useState({
    notes: '',
    additions: [{ 
      oil_id: '', 
      oil_grams: '', 
      product_id: '', 
      product_qty: '' 
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
      additions: [...formData.additions, { oil_id: '', oil_grams: '', product_id: '', product_qty: '' }]
    });
  };

  const removeAddition = (index: number) => {
    setFormData({
      ...formData,
      additions: formData.additions.filter((_, i) => i !== index)
    });
  };

  const updateAddition = (index: number, field: string, value: string) => {
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Transform additions into the format expected by the backend
    const body = {
      notes: formData.notes,
      oils: formData.additions
        .filter(a => a.oil_id && a.oil_grams)
        .map(a => ({
          oil_inventory_id: parseInt(a.oil_id),
          quantity_grams: parseFloat(a.oil_grams)
        })),
      products: formData.additions
        .filter(a => a.product_id && a.product_qty)
        .map(a => ({
          inventory_item_id: parseInt(a.product_id),
          quantity_produced: parseInt(a.product_qty)
        }))
    };

    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/inventory/manufacturing`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      });
      if (resp.ok) {
        toastSuccess('Manufacturing record created');
        setShowAddModal(false);
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
            setFormData({ notes: '', additions: [{ oil_id: '', oil_grams: '', product_id: '', product_qty: '' }] }); 
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
                    <div style={{ fontWeight: 600, color: '#1e293b' }}>{new Date(r.manufacturing_date).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })}</div>
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
                    <button className="toolbar-btn" 
                      style={{ 
                        color: '#ef4444', 
                        background: 'rgba(239, 68, 68, 0.05)', 
                        padding: '0.5rem 1rem', 
                        borderRadius: '10px',
                        fontSize: '0.8rem'
                      }} 
                      onClick={async () => {
                        if (window.confirm('Delete this record? Inventory will NOT be automatically reverted.')) {
                          await fetchWithAuth(`${API_BASE}/api/inventory/manufacturing?id=${r.id}`, { method: 'DELETE' });
                          fetchData();
                        }
                    }}>Delete</button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {showAddModal && (
        <div className="modal-overlay" style={{ backdropFilter: 'blur(12px)', backgroundColor: 'rgba(15, 23, 42, 0.6)' }} onClick={() => setShowAddModal(false)}>
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
              position: 'relative'
            }}>
            
            <button onClick={() => setShowAddModal(false)} style={{ position: 'absolute', top: '2rem', right: '2rem', color: '#94a3b8', background: 'none', border: 'none', cursor: 'pointer' }}>
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
                <h2 style={{ margin: 0, fontSize: '1.85rem', fontWeight: 900, color: '#1e293b', letterSpacing: '-0.02em' }}>Production Log</h2>
                <p style={{ margin: '0.4rem 0 0 0', color: '#64748b', fontSize: '0.95rem', fontWeight: 500 }}>Select oils to auto-fill products. Inventory will be synced instantly.</p>
              </div>
            </div>

            <form onSubmit={handleSubmit}>
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
                
                <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                  {formData.additions.map((a, index) => (
                    <div key={index} style={{ 
                      display: 'grid', 
                      gridTemplateColumns: '2.5fr 120px 2.5fr 100px 40px', 
                      gap: '1.25rem', 
                      alignItems: 'center', 
                      background: 'rgba(248, 250, 252, 0.5)', 
                      padding: '1.25rem', 
                      borderRadius: '20px', 
                      border: '1px solid #e2e8f0',
                      transition: 'all 0.2s ease'
                    }}>
                      <div className="input-group" style={{ marginBottom: 0 }}>
                        <label style={{ fontSize: '0.65rem', fontWeight: 700, color: '#64748b', marginBottom: '0.5rem', display: 'block', textTransform: 'uppercase' }}>Fragrance Oil</label>
                        <SearchableSelect 
                          options={oils} 
                          value={a.oil_id} 
                          onChange={(val) => updateAddition(index, 'oil_id', val)} 
                          placeholder="Search Oil..." 
                        />
                      </div>
                      <div className="input-group" style={{ marginBottom: 0 }}>
                        <label style={{ fontSize: '0.65rem', fontWeight: 700, color: '#64748b', marginBottom: '0.5rem', display: 'block', textTransform: 'uppercase' }}>Grams</label>
                        <input required type="number" step="0.1" placeholder="0.0" 
                          style={{ borderRadius: '12px', border: '1px solid #cbd5e1', padding: '0.75rem' }}
                          value={a.oil_grams} onChange={e => updateAddition(index, 'oil_grams', e.target.value)} />
                      </div>
                      <div className="input-group" style={{ marginBottom: 0 }}>
                        <label style={{ fontSize: '0.65rem', fontWeight: 700, color: '#64748b', marginBottom: '0.5rem', display: 'block', textTransform: 'uppercase' }}>Produced Product</label>
                        <select required 
                          style={{ borderRadius: '12px', border: '1px solid #cbd5e1', padding: '0.75rem', background: a.product_id ? '#f1f5f9' : 'white' }}
                          value={a.product_id} onChange={e => updateAddition(index, 'product_id', e.target.value)}>
                          <option value="">Select Product...</option>
                          {Array.isArray(products) && products.map(p => <option key={p.id} value={p.id}>{p.mi_sku} - {p.title}</option>)}
                        </select>
                      </div>
                      <div className="input-group" style={{ marginBottom: 0 }}>
                        <label style={{ fontSize: '0.65rem', fontWeight: 700, color: '#64748b', marginBottom: '0.5rem', display: 'block', textTransform: 'uppercase' }}>Quantity</label>
                        <input required type="number" placeholder="0" 
                          style={{ borderRadius: '12px', border: '1px solid #cbd5e1', padding: '0.75rem' }}
                          value={a.product_qty} onChange={e => updateAddition(index, 'product_qty', e.target.value)} />
                      </div>
                      <div style={{ display: 'flex', justifyContent: 'center' }}>
                        {formData.additions.length > 1 && (
                          <button type="button" onClick={() => removeAddition(index)} 
                            style={{ 
                              color: '#94a3b8', 
                              marginTop: '1.25rem', 
                              background: 'none', 
                              border: 'none', 
                              cursor: 'pointer',
                              padding: '8px',
                              borderRadius: '8px'
                            }}
                            onMouseOver={e => e.currentTarget.style.color = '#ef4444'}
                            onMouseOut={e => e.currentTarget.style.color = '#94a3b8'}
                          >
                            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                          </button>
                        )}
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

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 2fr', gap: '1.5rem', marginTop: '3.5rem' }}>
                <button type="button" className="btn-secondary" 
                  style={{ height: '56px', borderRadius: '16px', fontWeight: 700, fontSize: '1rem', border: '2px solid #e2e8f0' }}
                  onClick={() => setShowAddModal(false)}>
                  Dismiss
                </button>
                <button type="submit" className="btn-primary" 
                  style={{ 
                    height: '56px', 
                    borderRadius: '16px', 
                    fontWeight: 800, 
                    fontSize: '1rem', 
                    background: 'linear-gradient(135deg, var(--accent-color), #34d399)',
                    boxShadow: '0 15px 25px -5px rgba(16, 185, 129, 0.4)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    border: 'none'
                  }}>
                  Execute Production Cycle
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
