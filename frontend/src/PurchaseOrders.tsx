import React, { useState, useEffect } from 'react';
import { API_BASE } from './api';
import { useToast } from './ToastContext';
import { useConfirm } from './ConfirmContext';

interface PurchaseOrder {
  id: number;
  oil_inventory_id: number;
  supplier_id: number;
  quantity_grams: number;
  unit_price_per_kg: number;
  total_price: number;
  purchase_date: string;
  oil_inventory?: { name: string };
  supplier?: { name: string };
}

interface OilStock {
  id: number;
  name: string;
}

interface Supplier {
  id: number;
  name: string;
}

export const PurchaseOrders: React.FC<{ token: string | null }> = ({ token }) => {
  const { success: toastSuccess, error: toastError } = useToast();
  const { confirm } = useConfirm();
  const [pos, setPOs] = useState<PurchaseOrder[]>([]);
  const [oils, setOils] = useState<OilStock[]>([]);
  const [suppliers, setSuppliers] = useState<Supplier[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [showAddModal, setShowAddModal] = useState(false);
  const [editingRecordId, setEditingRecordId] = useState<number | null>(null);

  const [formData, setFormData] = useState({
    supplier_id: '',
    purchase_date: new Date().toISOString().split('T')[0],
    items: [{ oil_inventory_id: '', quantity_grams: '', unit_price_per_kg: '' }]
  });

  const fetchWithAuth = async (url: string, options: RequestInit = {}) => {
    const headers = { ...options.headers, 'Authorization': `Bearer ${token}` };
    return fetch(url, { ...options, headers });
  };

  const fetchData = async () => {
    setIsLoading(true);
    try {
      const [posRes, oilsRes, suppRes] = await Promise.all([
        fetchWithAuth(`${API_BASE}/api/inventory/po`),
        fetchWithAuth(`${API_BASE}/api/inventory/oil`),
        fetchWithAuth(`${API_BASE}/api/inventory/suppliers`)
      ]);

      if (posRes.ok) setPOs(await posRes.json());
      if (oilsRes.ok) setOils(await oilsRes.json());
      if (suppRes.ok) setSuppliers(await suppRes.json());
    } catch (err) {
      toastError('Failed to fetch data');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => { fetchData(); }, []);

  const handleAddItem = () => {
    setFormData({
      ...formData,
      items: [...formData.items, { oil_inventory_id: '', quantity_grams: '', unit_price_per_kg: '' }]
    });
  };

  const handleRemoveItem = (index: number) => {
    const newItems = [...formData.items];
    newItems.splice(index, 1);
    setFormData({ ...formData, items: newItems });
  };

  const handleUpdateItem = (index: number, field: string, value: string) => {
    const newItems = [...formData.items];
    newItems[index] = { ...newItems[index], [field]: value };
    setFormData({ ...formData, items: newItems });
  };

  const handleEdit = (po: PurchaseOrder) => {
    setEditingRecordId(po.id);
    let pDate = po.purchase_date.split('T')[0];
    if (pDate.startsWith('0001')) {
      pDate = new Date().toISOString().split('T')[0];
    }
    
    setFormData({
      supplier_id: po.supplier_id.toString(),
      purchase_date: pDate,
      items: [{
        oil_inventory_id: po.oil_inventory_id.toString(),
        quantity_grams: po.quantity_grams.toString(),
        unit_price_per_kg: po.unit_price_per_kg.toString()
      }]
    });
    setShowAddModal(true);
  };

  const handleCloseModal = () => {
    setShowAddModal(false);
    setEditingRecordId(null);
    setFormData({
      supplier_id: '',
      purchase_date: new Date().toISOString().split('T')[0],
      items: [{ oil_inventory_id: '', quantity_grams: '', unit_price_per_kg: '' }]
    });
  };

  const handleDelete = async (id: number) => {
    const confirmed = await confirm({
      title: 'Delete Purchase Order',
      message: 'Are you sure you want to delete this record? This will revert the oil stock levels.',
      confirmLabel: 'Delete',
      variant: 'danger'
    });
    if (!confirmed) return;
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/inventory/po?id=${id}`, { method: 'DELETE' });
      if (resp.ok) {
        toastSuccess('Record deleted');
        fetchData();
      }
    } catch (err) {
      toastError('Failed to delete record');
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (formData.items.length === 0 || isSaving) return;
    setIsSaving(true);

    try {
      if (editingRecordId) {
        // Single update
        const item = formData.items[0];
        const qty = parseFloat(item.quantity_grams);
        const price = parseFloat(item.unit_price_per_kg);
        const body = {
          id: editingRecordId,
          supplier_id: parseInt(formData.supplier_id),
          oil_inventory_id: parseInt(item.oil_inventory_id),
          quantity_grams: qty,
          unit_price_per_kg: price,
          total_price: (qty / 1000) * price,
          purchase_date: formData.purchase_date + 'T00:00:00Z'
        };

        const resp = await fetchWithAuth(`${API_BASE}/api/inventory/po`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(body)
        });
        if (!resp.ok) throw new Error();
      } else {
        // Bulk create
        const posToCreate = formData.items.map(item => {
          const qty = parseFloat(item.quantity_grams);
          const price = parseFloat(item.unit_price_per_kg);
          return {
            supplier_id: parseInt(formData.supplier_id),
            oil_inventory_id: parseInt(item.oil_inventory_id),
            quantity_grams: qty,
            unit_price_per_kg: price,
            total_price: (qty / 1000) * price,
            purchase_date: formData.purchase_date + 'T00:00:00Z'
          };
        });

        const resp = await fetchWithAuth(`${API_BASE}/api/inventory/po/bulk`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(posToCreate)
        });
        if (!resp.ok) throw new Error();
      }

      toastSuccess(editingRecordId ? 'Record updated' : 'Purchase orders recorded');
      handleCloseModal();
      fetchData();
    } catch (err) {
      toastError('Error saving purchase order');
    } finally {
      setIsSaving(false);
    }
  };

  const batchTotal = formData.items.reduce((acc, item) => {
    const qty = parseFloat(item.quantity_grams) || 0;
    const price = parseFloat(item.unit_price_per_kg) || 0;
    return acc + (qty / 1000) * price;
  }, 0);

  return (
    <div className="tab-content staggered-fade-in">
      <div className="section-header" style={{ marginBottom: '2rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h2 style={{ fontSize: '1.5rem', fontWeight: 700, margin: 0 }}>Purchase Orders</h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>History of raw material procurement.</p>
        </div>
        <button className="btn-primary" onClick={() => { 
          setEditingRecordId(null);
          setFormData({ 
            supplier_id: '', 
            purchase_date: new Date().toISOString().split('T')[0],
            items: [{ oil_inventory_id: '', quantity_grams: '', unit_price_per_kg: '' }] 
          }); 
          setShowAddModal(true); 
        }}>
          + New Purchase Order
        </button>
      </div>

      <div className="table-container glass-card-premium">
        <table className="premium-table">
          <thead>
            <tr>
              <th style={{ paddingLeft: '2rem' }}>Date</th>
              <th>Oil Name</th>
              <th>Supplier</th>
              <th>Quantity</th>
              <th>Price/kg</th>
              <th style={{ textAlign: 'right' }}>Total</th>
              <th style={{ paddingRight: '2rem', textAlign: 'center' }}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {pos.length === 0 && !isLoading ? (
              <tr>
                <td colSpan={7} style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-tertiary)' }}>No purchase orders found.</td>
              </tr>
            ) : (
              pos.map(po => {
                const isZeroDate = po.purchase_date.startsWith('0001');
                return (
                  <tr key={po.id} className="hover-row">
                    <td style={{ paddingLeft: '2rem', fontSize: '0.85rem' }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        {isZeroDate && <span title="Corrupted Date" style={{ color: '#ef4444' }}>⚠️</span>}
                        {isZeroDate ? 'N/A' : new Date(po.purchase_date).toLocaleDateString('en-GB')}
                      </div>
                    </td>
                    <td style={{ fontWeight: 700 }}>{po.oil_inventory?.name}</td>
                    <td><span className="badge-pill badge-pill-gray">{po.supplier?.name}</span></td>
                    <td style={{ fontWeight: 600 }}>{po.quantity_grams.toLocaleString()} g</td>
                    <td style={{ color: 'var(--text-secondary)' }}>₹{po.unit_price_per_kg.toLocaleString()}</td>
                    <td style={{ textAlign: 'right', fontWeight: 800, color: 'var(--accent-color)' }}>₹{po.total_price.toLocaleString()}</td>
                    <td style={{ paddingRight: '2rem', textAlign: 'center' }}>
                      <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'center' }}>
                        <button className="icon-btn" onClick={() => handleEdit(po)} title="Edit">
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                        </button>
                        <button className="icon-btn delete" onClick={() => handleDelete(po.id)} title="Delete">
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/></svg>
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      {showAddModal && (
        <div className="modal-overlay" style={{ backdropFilter: 'blur(8px)', backgroundColor: 'rgba(0,0,0,0.4)' }} onClick={handleCloseModal}>
          <div className="premium-modal" onClick={e => e.stopPropagation()} style={{ maxWidth: '800px', width: '95%', borderRadius: '24px', padding: '2.5rem' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '1.25rem', marginBottom: '2.5rem' }}>
              <div style={{ width: '56px', height: '56px', borderRadius: '16px', background: 'linear-gradient(135deg, #10b981, #34d399)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'white', boxShadow: '0 8px 16px rgba(16, 185, 129, 0.2)' }}>
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M20 7h-9m3 3H5m16 4h-9m3 3H5"/><path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/></svg>
              </div>
              <div>
                <h2 style={{ margin: 0, fontSize: '1.75rem', fontWeight: 800, color: 'var(--text-primary)' }}>{editingRecordId ? 'Edit Purchase' : 'Record Purchase'}</h2>
                <p style={{ margin: 0, color: 'var(--text-secondary)', fontSize: '0.9rem', opacity: 0.8 }}>{editingRecordId ? 'Correct mistakes in the purchase record.' : 'Batch upload multiple oil stocks from a supplier.'}</p>
              </div>
            </div>

            <form onSubmit={handleSubmit}>
              <div style={{ display: 'grid', gridTemplateColumns: '1.5fr 1fr', gap: '2rem', marginBottom: '2.5rem', background: 'var(--bg-color)', padding: '2rem', borderRadius: '20px', border: '1px solid var(--border-color)' }}>
                <div className="input-group">
                  <label style={{ fontSize: '0.75rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '0.75rem', display: 'block' }}>Supplier</label>
                  <select required value={formData.supplier_id} onChange={e => setFormData({ ...formData, supplier_id: e.target.value })} style={{ height: '52px', borderRadius: '12px', border: '1px solid var(--border-color)', backgroundColor: 'white', fontSize: '0.95rem', fontWeight: 500 }}>
                    <option value="">Select Supplier...</option>
                    {suppliers.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
                  </select>
                </div>

                <div className="input-group">
                  <label style={{ fontSize: '0.75rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '0.75rem', display: 'block' }}>Purchase Date</label>
                  <input required type="date" value={formData.purchase_date} onChange={e => setFormData({ ...formData, purchase_date: e.target.value })} style={{ height: '52px', borderRadius: '12px', border: '1px solid var(--border-color)', backgroundColor: 'white', padding: '0 1.25rem', fontSize: '0.95rem', width: '100%', outline: 'none' }} />
                </div>
              </div>

              <div style={{ marginBottom: '2rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
                  <h3 style={{ fontSize: '1.1rem', fontWeight: 800, margin: 0, color: 'var(--text-primary)', letterSpacing: '-0.02em' }}>Oil Items</h3>
                  {!editingRecordId && (
                    <button type="button" className="btn-secondary" onClick={handleAddItem} style={{ padding: '0.6rem 1.25rem', fontSize: '0.8rem', height: 'auto', borderRadius: '10px', border: '1px solid var(--border-color)', fontWeight: 700, backgroundColor: 'white', color: 'var(--text-secondary)' }}>
                      + Add Item
                    </button>
                  )}
                </div>

                <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem', maxHeight: '350px', overflowY: 'auto', paddingRight: '0.5rem', paddingLeft: '2px' }}>
                  {formData.items.map((item, idx) => (
                    <div key={idx} style={{ display: 'grid', gridTemplateColumns: '2fr 1fr 1.2fr auto', gap: '1.25rem', alignItems: 'end', background: 'white', padding: '1.25rem', borderRadius: '16px', border: '1px solid var(--border-color)', boxShadow: '0 2px 8px rgba(0,0,0,0.02)' }}>
                      <div className="input-group">
                        <label style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-tertiary)', marginBottom: '0.5rem', display: 'block' }}>Oil Name</label>
                        <select required value={item.oil_inventory_id} onChange={e => handleUpdateItem(idx, 'oil_inventory_id', e.target.value)} style={{ background: 'var(--bg-input)', border: 'none', height: '44px' }}>
                          <option value="">Select Oil...</option>
                          {oils.map(o => <option key={o.id} value={o.id}>{o.name}</option>)}
                        </select>
                      </div>
                      <div className="input-group">
                        <label style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-tertiary)', marginBottom: '0.5rem', display: 'block' }}>Grams</label>
                        <input required type="number" value={item.quantity_grams} onChange={e => handleUpdateItem(idx, 'quantity_grams', e.target.value)} placeholder="0" style={{ background: 'var(--bg-input)', border: 'none', height: '44px' }} />
                      </div>
                      <div className="input-group">
                        <label style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-tertiary)', marginBottom: '0.5rem', display: 'block' }}>Price / kg</label>
                        <div style={{ position: 'relative' }}>
                          <input required type="number" value={item.unit_price_per_kg} onChange={e => handleUpdateItem(idx, 'unit_price_per_kg', e.target.value)} placeholder="0.00" style={{ paddingLeft: '2rem', background: 'var(--bg-input)', border: 'none', height: '44px' }} />
                          <span style={{ position: 'absolute', left: '0.75rem', top: '50%', transform: 'translateY(-50%)', fontWeight: 800, fontSize: '0.9rem', color: 'var(--text-tertiary)' }}>₹</span>
                        </div>
                      </div>
                      <div style={{ height: '44px', display: 'flex', alignItems: 'center' }}>
                        {!editingRecordId && formData.items.length > 1 && (
                          <button type="button" className="icon-btn delete" onClick={() => handleRemoveItem(idx)} style={{ width: '36px', height: '36px' }}>
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M3 6h18m-2 0v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6m3 0V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"/></svg>
                          </button>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '2.5rem', alignItems: 'center', marginTop: '3rem', paddingTop: '2rem', borderTop: '1.5px solid var(--border-color)' }}>
                <div style={{ textAlign: 'right' }}>
                  <p style={{ margin: 0, fontSize: '0.75rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.1em' }}>Total Investment</p>
                  <p style={{ margin: 0, fontSize: '1.75rem', fontWeight: 900, color: 'var(--accent-color)', display: 'flex', alignItems: 'center', gap: '0.25rem', justifyContent: 'flex-end' }}>
                    <span style={{ fontSize: '1.25rem', opacity: 0.6 }}>₹</span>{batchTotal.toLocaleString()}
                  </p>
                </div>
                <div className="modal-actions" style={{ marginTop: 0, gap: '1.25rem' }}>
                  <button type="button" className="btn-secondary" onClick={handleCloseModal} disabled={isSaving} style={{ width: '130px', height: '54px', borderRadius: '16px', fontWeight: 700, fontSize: '1rem' }}>Cancel</button>
                  <button type="submit" className="btn-primary" disabled={isSaving} style={{ width: '220px', height: '54px', borderRadius: '16px', background: 'linear-gradient(135deg, #10b981, #059669)', fontWeight: 800, fontSize: '1rem', boxShadow: '0 10px 25px rgba(16, 185, 129, 0.25)', border: 'none', color: 'white', letterSpacing: '0.02em', cursor: isSaving ? 'not-allowed' : 'pointer', opacity: isSaving ? 0.7 : 1 }}>
                    {isSaving ? 'Recording...' : (editingRecordId ? 'Update Record' : 'Record Purchase')}
                  </button>
                </div>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
