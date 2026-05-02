import React, { useState, useEffect } from 'react';
import { API_BASE } from './api';
import { useToast } from './ToastContext';

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
  const [pos, setPOs] = useState<PurchaseOrder[]>([]);
  const [oils, setOils] = useState<OilStock[]>([]);
  const [suppliers, setSuppliers] = useState<Supplier[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showAddModal, setShowAddModal] = useState(false);

  const [formData, setFormData] = useState({
    oil_inventory_id: '',
    supplier_id: '',
    quantity_grams: '',
    unit_price_per_kg: ''
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const qty = parseFloat(formData.quantity_grams);
    const price = parseFloat(formData.unit_price_per_kg);
    const body = {
      oil_inventory_id: parseInt(formData.oil_inventory_id),
      supplier_id: parseInt(formData.supplier_id),
      quantity_grams: qty,
      unit_price_per_kg: price,
      total_price: (qty / 1000) * price
    };

    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/inventory/po`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      });
      if (resp.ok) {
        toastSuccess('Purchase order recorded');
        setShowAddModal(false);
        fetchData();
      }
    } catch (err) {
      toastError('Error saving purchase order');
    }
  };

  return (
    <div className="tab-content staggered-fade-in">
      <div className="section-header" style={{ marginBottom: '2rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h2 style={{ fontSize: '1.5rem', fontWeight: 700, margin: 0 }}>Purchase Orders</h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>History of raw material procurement.</p>
        </div>
        <button className="btn-primary" onClick={() => { 
          setFormData({ oil_inventory_id: '', supplier_id: '', quantity_grams: '', unit_price_per_kg: '' }); 
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
              <th style={{ paddingRight: '2rem', textAlign: 'right' }}>Total</th>
            </tr>
          </thead>
          <tbody>
            {pos.length === 0 && !isLoading ? (
              <tr>
                <td colSpan={6} style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-tertiary)' }}>No purchase orders found.</td>
              </tr>
            ) : (
              pos.map(po => (
                <tr key={po.id} className="hover-row">
                  <td style={{ paddingLeft: '2rem', fontSize: '0.85rem' }}>{new Date(po.purchase_date).toLocaleDateString()}</td>
                  <td style={{ fontWeight: 700 }}>{po.oil_inventory?.name}</td>
                  <td><span className="badge-pill badge-pill-gray">{po.supplier?.name}</span></td>
                  <td style={{ fontWeight: 600 }}>{po.quantity_grams.toLocaleString()} g</td>
                  <td style={{ color: 'var(--text-secondary)' }}>₹{po.unit_price_per_kg.toLocaleString()}</td>
                  <td style={{ paddingRight: '2rem', textAlign: 'right', fontWeight: 800, color: 'var(--accent-color)' }}>₹{po.total_price.toLocaleString()}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {showAddModal && (
        <div className="modal-overlay" style={{ backdropFilter: 'blur(8px)', backgroundColor: 'rgba(0,0,0,0.4)' }} onClick={() => setShowAddModal(false)}>
          <div className="premium-modal" onClick={e => e.stopPropagation()} style={{ maxWidth: '550px', width: '95%', borderRadius: '24px', padding: '2rem' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', marginBottom: '2rem' }}>
              <div style={{ width: '48px', height: '48px', borderRadius: '14px', background: 'linear-gradient(135deg, #3b82f6, #2563eb)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'white' }}>
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M20 7h-9m3 3H5m16 4h-9m3 3H5"/><path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/></svg>
              </div>
              <div>
                <h2 style={{ margin: 0, fontSize: '1.5rem', fontWeight: 800 }}>Record Purchase</h2>
                <p style={{ margin: 0, color: 'var(--text-secondary)', fontSize: '0.85rem' }}>Updates Oil Stock and latest purchase price.</p>
              </div>
            </div>

            <form onSubmit={handleSubmit}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.25rem' }}>
                <div className="input-group" style={{ gridColumn: 'span 2' }}>
                  <label style={{ fontSize: '0.7rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', marginBottom: '0.6rem', display: 'block' }}>Oil to Buy</label>
                  <select required value={formData.oil_inventory_id} onChange={e => setFormData({ ...formData, oil_inventory_id: e.target.value })}>
                    <option value="">Select Oil...</option>
                    {oils.map(o => <option key={o.id} value={o.id}>{o.name}</option>)}
                  </select>
                </div>

                <div className="input-group" style={{ gridColumn: 'span 2' }}>
                  <label style={{ fontSize: '0.7rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', marginBottom: '0.6rem', display: 'block' }}>Supplier (Mandatory)</label>
                  <select required value={formData.supplier_id} onChange={e => setFormData({ ...formData, supplier_id: e.target.value })}>
                    <option value="">Select Supplier...</option>
                    {suppliers.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
                  </select>
                </div>

                <div className="input-group">
                  <label style={{ fontSize: '0.7rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', marginBottom: '0.6rem', display: 'block' }}>Quantity (Grams)</label>
                  <input required type="number" value={formData.quantity_grams} onChange={e => setFormData({ ...formData, quantity_grams: e.target.value })} placeholder="e.g. 1000" />
                </div>

                <div className="input-group">
                  <label style={{ fontSize: '0.7rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', marginBottom: '0.6rem', display: 'block' }}>Price per kg</label>
                  <div style={{ position: 'relative' }}>
                    <input required type="number" value={formData.unit_price_per_kg} onChange={e => setFormData({ ...formData, unit_price_per_kg: e.target.value })} placeholder="0.00" style={{ paddingLeft: '2rem' }} />
                    <span style={{ position: 'absolute', left: '0.75rem', top: '50%', transform: 'translateY(-50%)', fontWeight: 700 }}>₹</span>
                  </div>
                </div>
              </div>

              <div className="modal-actions" style={{ marginTop: '2.5rem', gap: '1rem' }}>
                <button type="button" className="btn-secondary" onClick={() => setShowAddModal(false)} style={{ flex: 1, height: '48px', borderRadius: '12px' }}>Cancel</button>
                <button type="submit" className="btn-primary" style={{ flex: 1, height: '48px', borderRadius: '12px', background: 'linear-gradient(135deg, #3b82f6, #2563eb)' }}>Record Purchase</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
