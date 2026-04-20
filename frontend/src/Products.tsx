import React, { useState, useEffect } from 'react';
import { API_BASE } from './api';
import { useToast } from './ToastContext';

interface InventoryItem {
  id: number;
  mi_sku: string;
  title: string;
  description: string;
  current_stock: number;
  mappings?: InventoryMapping[];
}

interface InventoryMapping {
  id: number;
  platform: string;
  external_sku: string;
  external_variant_id: string;
}

interface StagedProduct {
  title: string;
  description: string;
  mappings: {
    platform: string;
    external_sku: string;
    external_variant_id: string;
  }[];
}

export const Products: React.FC<{ token: string | null }> = ({ token }) => {
  const { success: toastSuccess, error: toastError } = useToast();
  const [items, setItems] = useState<InventoryItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [showAddModal, setShowAddModal] = useState(false);
  const [showSyncWizard, setShowSyncWizard] = useState(false);
  const [stagedProducts, setStagedProducts] = useState<StagedProduct[]>([]);
  const [isSyncing, setIsSyncing] = useState(false);

  // New Product Form
  const [formData, setFormData] = useState({
    mi_sku: '',
    title: '',
    description: '',
    current_stock: 0
  });

  const fetchWithAuth = async (url: string, options: RequestInit = {}) => {
    const headers = {
      ...options.headers,
      'Authorization': `Bearer ${token}`
    };
    return fetch(url, { ...options, headers });
  };

  const fetchInventory = async (search: string = '') => {
    setIsLoading(true);
    try {
      const url = search ? `${API_BASE}/api/inventory?search=${encodeURIComponent(search)}` : `${API_BASE}/api/inventory`;
      const resp = await fetchWithAuth(url);
      const data = await resp.json();
      setItems(data || []);
    } catch (err) {
      toastError('Failed to fetch inventory');
    } finally {
      setIsLoading(false);
    }
  };

  const fetchNextSKU = async () => {
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/inventory/next-sku`);
      const data = await resp.json();
      setFormData(prev => ({ ...prev, mi_sku: data.next_sku }));
    } catch (err) {
      console.error('Failed to suggest SKU');
    }
  };

  const handleCreateItem = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/inventory`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData)
      });
      if (resp.ok) {
        toastSuccess('Product created successfully');
        setShowAddModal(false);
        fetchInventory();
      } else {
        toastError('Failed to create product');
      }
    } catch (err) {
      toastError('Error creating product');
    }
  };

  const handleSyncShopify = async () => {
    setIsSyncing(true);
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/inventory/sync-shopify`, { method: 'POST' });
      const data = await resp.json();
      setStagedProducts(data);
      setShowSyncWizard(true);
    } catch (err) {
      toastError('Failed to sync Shopify products');
    } finally {
      setIsSyncing(false);
    }
  };

  const handleMapStaged = async (staged: StagedProduct) => {
    // 1. Create the internal product first (Auto SKU)
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/inventory`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          title: staged.title,
          description: staged.description,
          current_stock: 0 // Default to zero for sync
        })
      });
      const newItem = await resp.json();

      // 2. Map the external variant
      const m = staged.mappings[0];
      await fetchWithAuth(`${API_BASE}/api/inventory/map`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          internal_item_id: newItem.id,
          platform: m.platform,
          external_sku: m.external_sku,
          variant_id: m.external_variant_id
        })
      });

      toastSuccess(`Mapped ${staged.title} to ${newItem.mi_sku}`);
      setStagedProducts(prev => prev.filter(p => p.mappings[0].external_sku !== m.external_sku));
      fetchInventory();
    } catch (err) {
      toastError('Failed to map product');
    }
  };

  useEffect(() => {
    if (searchQuery === '') {
      fetchInventory('');
      return;
    }
    const timer = setTimeout(() => {
      fetchInventory(searchQuery);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  return (
    <div className="tab-content staggered-fade-in">
      <StyleTag />
      <div className="section-header" style={{ marginBottom: '2rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h2 style={{ fontSize: '1.5rem', fontWeight: 700, margin: 0 }}>Warehouse Authority</h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>Manage physical products and SKU mappings.</p>
        </div>
        <div style={{ display: 'flex', gap: '1rem' }}>
          <button className="btn-secondary" onClick={handleSyncShopify} disabled={isSyncing}>
            {isSyncing ? 'Scanning Shopify...' : 'Sync From Shopify'}
          </button>
          <button className="btn-primary" onClick={() => { setShowAddModal(true); fetchNextSKU(); }}>
            + Add New Product
          </button>
        </div>
      </div>

      <div className="search-container" style={{ marginBottom: '2rem' }}>
        <div className="search-box-premium">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
          <input
            type="text"
            placeholder="Search by SKU or title..."
            aria-label="Search products"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            style={{ paddingRight: searchQuery ? '2.5rem' : '1rem' }}
          />
          {searchQuery && (
            <button
              onClick={() => setSearchQuery('')}
              aria-label="Clear search"
              title="Clear search"
              className="clear-search-btn"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
            </button>
          )}
        </div>
      </div>

      <div className="inventory-grid" style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))', gap: '1.5rem' }}>
        {items.length === 0 && !isLoading && (
          <div style={{ gridColumn: '1/-1', textAlign: 'center', padding: '4rem', background: 'var(--surface-color)', borderRadius: '16px', border: '1px dashed var(--border-color)' }}>
            <p style={{ color: 'var(--text-tertiary)' }}>
              {searchQuery ? `No products found matching "${searchQuery}"` : "No products in your warehouse yet."}
            </p>
            {searchQuery && (
              <button
                className="btn-secondary"
                style={{ marginTop: '1rem' }}
                onClick={() => setSearchQuery('')}
              >
                Clear Search
              </button>
            )}
          </div>
        )}

        {items.map(item => (
          <div key={item.id} className="premium-card hover-lift" style={{ padding: '1.5rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1rem' }}>
              <div>
                <span className="badge-pill badge-pill-gray" style={{ marginBottom: '0.5rem' }}>{item.mi_sku}</span>
                <h3 style={{ margin: 0, fontSize: '1.125rem' }}>{item.title}</h3>
              </div>
              <div style={{ textAlign: 'right' }}>
                <div style={{ fontSize: '1.25rem', fontWeight: 700, color: item.current_stock > 10 ? 'var(--status-active)' : 'var(--status-warning)' }}>
                  {item.current_stock}
                </div>
                <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>Units</div>
              </div>
            </div>
            
            <p style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', marginBottom: '1.5rem', display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', overflow: 'hidden' }}>
              {item.description || 'No description provided.'}
            </p>

            <div className="mapping-section" style={{ borderTop: '1px solid var(--border-color)', paddingTop: '1rem' }}>
              <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase', marginBottom: '0.5rem' }}>Linked Channels</div>
              {item.mappings && item.mappings.length > 0 ? (
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem' }}>
                  {item.mappings.map(m => (
                    <span key={m.id} className="badge-pill" style={{ fontSize: '0.7rem', background: 'var(--bg-input)' }}>
                      {m.platform}: {m.external_sku}
                    </span>
                  ))}
                </div>
              ) : (
                <div style={{ fontSize: '0.75rem', fontStyle: 'italic', color: 'var(--text-tertiary)' }}>Not mapped to Shopify/Amazon</div>
              )}
            </div>
          </div>
        ))}
      </div>

      {/* Add Product Modal */}
      {showAddModal && (
        <div className="modal-overlay">
          <div className="premium-modal">
            <h2>Add Internal Product</h2>
            <form onSubmit={handleCreateItem}>
              <div style={{ display: 'grid', gap: '1rem', marginTop: '1.5rem' }}>
                <div className="input-group">
                  <label>Internal SKU (Auto-generated)</label>
                  <input 
                    type="text" 
                    value={formData.mi_sku} 
                    onChange={e => setFormData({...formData, mi_sku: e.target.value})}
                  />
                </div>
                <div className="input-group">
                  <label>Product Title</label>
                  <input 
                    type="text" 
                    required
                    value={formData.title} 
                    onChange={e => setFormData({...formData, title: e.target.value})}
                  />
                </div>
                <div className="input-group">
                  <label>Description</label>
                  <textarea 
                    rows={4}
                    value={formData.description} 
                    onChange={e => setFormData({...formData, description: e.target.value})}
                    style={{ background: 'var(--bg-input)', border: '1px solid var(--border-color)', borderRadius: '8px', color: 'var(--text-primary)', padding: '0.5rem' }}
                  />
                </div>
                <div className="input-group">
                  <label>Initial Stock Level</label>
                  <input 
                    type="number" 
                    value={formData.current_stock} 
                    onChange={e => setFormData({...formData, current_stock: parseInt(e.target.value)})}
                  />
                </div>
              </div>
              <div className="modal-actions" style={{ marginTop: '2rem' }}>
                <button type="button" className="btn-secondary" onClick={() => setShowAddModal(false)}>Cancel</button>
                <button type="submit" className="btn-primary">Create Product</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Sync Wizard Modal */}
      {showSyncWizard && (
        <div className="modal-overlay">
          <div className="premium-modal wide" style={{ maxWidth: '900px' }}>
            <h2>Shopify Sync Wizard</h2>
            <p>We found {stagedProducts.length} variants on Shopify. Map them to your internal warehouse.</p>
            
            <div style={{ maxHeight: '400px', overflowY: 'auto', marginTop: '1.5rem', border: '1px solid var(--border-color)', borderRadius: '12px' }}>
              <table style={{ margin: 0 }}>
                <thead style={{ position: 'sticky', top: 0, background: 'var(--surface-color)', zIndex: 10 }}>
                  <tr>
                    <th>Shopify Product / Variant</th>
                    <th>Shopify SKU</th>
                    <th>Action</th>
                  </tr>
                </thead>
                <tbody>
                  {stagedProducts.map((p, idx) => (
                    <tr key={idx}>
                      <td>
                        <div style={{ fontWeight: 600 }}>{p.title}</div>
                        <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', maxWidth: '300px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                          {p.description}
                        </div>
                      </td>
                      <td><code>{p.mappings[0].external_sku}</code></td>
                      <td>
                        <button className="btn-secondary" style={{ fontSize: '0.75rem', padding: '0.25rem 0.75rem' }} onClick={() => handleMapStaged(p)}>
                          Add as Internal
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            
            <div className="modal-actions" style={{ marginTop: '2rem' }}>
              <button className="btn-secondary" onClick={() => setShowSyncWizard(false)}>Close</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

const styles = `
  .search-box-premium {
    flex: 1;
    background: var(--surface-color);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    padding: 0 1rem;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    height: 48px;
    box-shadow: var(--shadow-sm);
    position: relative;
  }
  .search-box-premium input {
    background: none;
    border: none;
    width: 100%;
    color: var(--text-primary);
    font-weight: 500;
    outline: none;
  }
  .search-box-premium:focus-within {
    border-color: var(--accent-color);
    box-shadow: 0 0 0 3px var(--accent-subtle);
  }
  .clear-search-btn {
    position: absolute;
    right: 12px;
    top: 50%;
    transform: translateY(-50%);
    color: var(--text-tertiary);
    padding: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    transition: all 0.2s;
    cursor: pointer;
    border: none;
    background: transparent;
  }
  .clear-search-btn:hover {
    color: var(--text-primary);
    background: var(--bg-hover);
  }
`;

const StyleTag = () => <style>{styles}</style>;

export default Products;
