import React, { useState, useEffect, useMemo, useRef } from 'react';
import { API_BASE } from './api';
import { useToast } from './ToastContext';

interface InventoryItem {
  id: number;
  mi_sku: string;
  title: string;
  description: string;
  specification: string;
  current_stock: number;
  mappings?: InventoryMapping[];
}

interface InventoryMapping {
  id: number;
  platform: string;
  external_sku: string;
  external_variant_id: string;
}

interface InventoryLog {
  id: number;
  inventory_item_id: number;
  delta: number;
  reason: string;
  platform: string;
  external_order_id: string;
  created_at: string;
}

// Helper to parse Shopify Rich Text JSON into HTML
const parseShopifyRichText = (input: string) => {
  if (!input) return '';
  // Check if it's potentially a JSON string
  if (!input.trim().startsWith('{"type":"root"')) return input;

  try {
    const data = JSON.parse(input);
    const renderNode = (node: any): string => {
      if (!node) return '';
      if (node.type === 'root' || node.type === 'paragraph') {
        const content = node.children?.map(renderNode).join('') || '';
        return node.type === 'paragraph' ? `<p style="margin-bottom: 0.75rem;">${content}</p>` : content;
      }
      if (node.type === 'text') {
        let text = node.value || '';
        if (node.bold) text = `<strong>${text}</strong>`;
        if (node.italic) text = `<em>${text}</em>`;
        return text;
      }
      return '';
    };
    return renderNode(data);
  } catch (e) {
    return input;
  }
};

// Helper to extract SKU by platform
const getSKUForPlatform = (mappings: InventoryMapping[] | undefined, platform: string) => {
  if (!mappings) return '—';
  const mapping = mappings.find(m => m.platform.toLowerCase() === platform.toLowerCase());
  return mapping ? mapping.external_sku : '—';
};

export const Products: React.FC<{ token: string | null, userRole?: string, appConfigs?: any }> = ({ token, userRole, appConfigs }) => {
  const { success: toastSuccess, error: toastError } = useToast();
  const [items, setItems] = useState<InventoryItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showAddModal, setShowAddModal] = useState(false);
  const [selectedProduct, setSelectedProduct] = useState<InventoryItem | null>(null);
  const [productLogs, setProductLogs] = useState<InventoryLog[]>([]);
  const [showLogsModal, setShowLogsModal] = useState(false);
  const [amazonSKUInput, setAmazonSKUInput] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const searchInputRef = useRef<HTMLInputElement>(null);
  const [sortBy, setSortBy] = useState<string>('mi-sku-asc');
  const [isSyncing, setIsSyncing] = useState(false);
  const [showSyncModal, setShowSyncModal] = useState(false);
  const [syncMode, setSyncMode] = useState<'shopify' | 'amazon'>('shopify');
  const [syncStartDate, setSyncStartDate] = useState(new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Kolkata' }));
  const [syncEndDate, setSyncEndDate] = useState(new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Kolkata' }));
  const [stagedProducts, setStagedProducts] = useState<InventoryItem[]>([]);
  const [selectedStagedIds, setSelectedStagedIds] = useState<Set<string>>(new Set());
  const [isSaving, setIsSaving] = useState(false);

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

  const handleUpdateStock = async (id: number, newVal: number) => {
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/inventory/stock?id=${id}&val=${newVal}`, {
        method: 'PUT'
      });
      if (resp.ok) {
        toastSuccess('Stock updated successfully');
        fetchInventory();
      } else {
        toastError('Failed to update stock');
      }
    } catch (err) {
      toastError('Error updating stock');
    }
  };

  const EditableStockPill = ({ id, initialStock }: { id: number, initialStock: number }) => {
    const [isEditing, setIsEditing] = useState(false);
    const [val, setVal] = useState(initialStock);

    const handleBlur = async () => {
      if (val === initialStock) {
        setIsEditing(false);
        return;
      }
      await handleUpdateStock(id, val);
      setIsEditing(false);
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
      if (e.key === 'Enter') handleBlur();
      if (e.key === 'Escape') {
        setVal(initialStock);
        setIsEditing(false);
      }
    };

    if (isEditing) {
      return (
        <input 
          type="number"
          autoFocus
          value={val}
          onChange={(e) => setVal(parseInt(e.target.value) || 0)}
          onBlur={handleBlur}
          onKeyDown={handleKeyDown}
          style={{
            width: '70px',
            padding: '4px 8px',
            borderRadius: '8px',
            border: '1px solid var(--accent-color)',
            background: 'var(--surface-color-secondary)',
            color: 'var(--text-primary)',
            fontWeight: 700,
            fontSize: '0.9rem',
            outline: 'none',
            boxShadow: '0 0 10px rgba(var(--accent-color-rgb), 0.2)'
          }}
        />
      );
    }

    return (
      <div 
        onClick={(e) => { e.stopPropagation(); setIsEditing(true); }}
        className="editable-stock-trigger"
        style={{ 
          display: 'flex', 
          alignItems: 'center', 
          gap: '0.75rem', 
          cursor: 'pointer',
          padding: '4px 8px',
          borderRadius: '8px',
          transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)'
        }}
      >
        <div style={{ 
          width: '8px', 
          height: '8px', 
          borderRadius: '50%', 
          background: initialStock > 10 ? 'var(--status-active)' : initialStock > 0 ? 'var(--status-warning)' : 'var(--status-danger)',
          boxShadow: `0 0 10px ${initialStock > 10 ? 'var(--status-active)' : initialStock > 0 ? 'var(--status-warning)' : 'var(--status-danger)'}`
        }} />
        <span style={{ 
          fontWeight: 700, 
          fontSize: '0.9rem',
          color: initialStock > 10 ? 'var(--status-active)' : initialStock > 0 ? 'var(--status-warning)' : 'var(--status-danger)'
        }}>
          {initialStock} units
        </span>
        <span style={{ fontSize: '0.7rem', opacity: 0.3, marginLeft: '4px' }}>✎</span>
      </div>
    );
  };

  const EditableSKU = ({ itemID, currentSKU, platform }: { itemID: number, currentSKU: string, platform: string }) => {
    const [isEditing, setIsEditing] = useState(false);
    const [val, setVal] = useState(currentSKU === '—' ? '' : currentSKU);

    const handleSave = async () => {
      if (val === (currentSKU === '—' ? '' : currentSKU)) {
        setIsEditing(false);
        return;
      }

      try {
        const resp = await fetchWithAuth(`${API_BASE}/api/inventory/map`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            internal_item_id: itemID,
            platform: platform,
            external_sku: val,
            variant_id: 'default'
          })
        });
        if (resp.ok) {
          toastSuccess(`${platform.toUpperCase()} SKU updated`);
          fetchInventory();
        } else {
          toastError(`Failed to update ${platform} SKU`);
          setVal(currentSKU === '—' ? '' : currentSKU);
        }
      } catch (err) {
        toastError('Error updating SKU');
      }
      setIsEditing(false);
    };

    if (isEditing) {
      return (
        <input 
          type="text"
          autoFocus
          placeholder="Enter SKU..."
          value={val}
          onChange={(e) => setVal(e.target.value)}
          onBlur={handleSave}
          onKeyDown={(e) => {
            if (e.key === 'Enter') handleSave();
            if (e.key === 'Escape') {
              setVal(currentSKU === '—' ? '' : currentSKU);
              setIsEditing(false);
            }
          }}
          onClick={(e) => e.stopPropagation()}
          style={{
            width: '120px',
            padding: '4px 8px',
            borderRadius: '6px',
            border: '1px solid var(--accent-color)',
            background: 'var(--surface-color-secondary)',
            color: 'var(--text-primary)',
            fontSize: '0.8rem',
            outline: 'none'
          }}
        />
      );
    }

    return (
      <div 
        onClick={(e) => { e.stopPropagation(); setIsEditing(true); }}
        className="editable-sku-trigger"
        style={{ 
          cursor: 'pointer',
          padding: '2px 6px',
          borderRadius: '4px',
          display: 'inline-flex',
          alignItems: 'center',
          gap: '4px',
          transition: 'all 0.2s',
          border: '1px solid transparent'
        }}
      >
        <code style={{ fontSize: '0.8rem', color: currentSKU === '—' ? 'var(--text-tertiary)' : 'var(--text-secondary)' }}>
          {currentSKU}
        </code>
        <span style={{ fontSize: '0.6rem', opacity: 0.2 }}>✎</span>
      </div>
    );
  };

  const fetchInventory = async () => {
    setIsLoading(true);
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/inventory`);
      const data = await resp.json();
      setItems(data);
      if (selectedProduct) {
        const updated = data.find((item: InventoryItem) => item.id === selectedProduct.id);
        if (updated) {
          setSelectedProduct(updated);
        }
      }
    } catch (err) {
      toastError('Failed to fetch inventory');
    } finally {
      setIsLoading(false);
    }
  };

  const handleSetSelected = (item: InventoryItem) => {
    setSelectedProduct(item);
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
    if (isSaving) return;
    setIsSaving(true);
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
    } finally {
      setIsSaving(false);
    }
  };


  const fetchLogs = async (id: number) => {
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/inventory/logs?id=${id}`);
      const data = await resp.json();
      setProductLogs(data || []);
      setShowLogsModal(true);
    } catch (err) {
      toastError('Failed to fetch movement logs');
    }
  };

  const handleAddAmazonMapping = async (itemID: number) => {
    if (!amazonSKUInput.trim()) return;
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/inventory/map`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          internal_item_id: itemID,
          platform: 'amazon',
          external_sku: amazonSKUInput,
          variant_id: 'default' // Amazon doesn't use variant ID in our mapping logic yet
        })
      });
      if (resp.ok) {
        toastSuccess('Amazon SKU mapped successfully');
        setAmazonSKUInput('');
        fetchInventory();
      } else {
        toastError('Failed to map Amazon SKU');
      }
    } catch (err) {
      toastError('Failed to map Amazon SKU');
    }
  };

  const handleDeleteMapping = async (mappingID: number) => {
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/inventory/map?id=${mappingID}`, {
        method: 'DELETE'
      });
      if (resp.ok) {
        toastSuccess('SKU mapping removed successfully');
        fetchInventory();
      } else {
        toastError('Failed to delete mapping');
      }
    } catch (err) {
      toastError('Failed to delete mapping');
    }
  };


  const handleSyncShopify = async () => {
    setSyncMode('shopify');
    setShowSyncModal(true);
    setStagedProducts([]);
    setIsSyncing(true);
    
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/inventory/sync-shopify`, { method: 'POST' });
      if (resp.ok) {
        const data = await resp.json();
        setStagedProducts(data);
      } else {
        toastError('Failed to fetch products from Shopify');
      }
    } catch (err) {
      toastError('Network error fetching Shopify products');
    } finally {
      setIsSyncing(false);
    }
  };

  const handleSyncAmazon = () => {
    setSyncMode('amazon');
    setShowSyncModal(true);
  };

  const handleStartSync = async () => {
    if (syncMode === 'shopify') {
      // Handle bulk import of selected staged products
      if (selectedStagedIds.size === 0) {
        toastError('No products selected for import');
        return;
      }
      
      setIsSyncing(true);
      const toImport = stagedProducts.filter(p => {
        const variantId = p.mappings?.[0]?.external_variant_id;
        return variantId && selectedStagedIds.has(variantId);
      });

      try {
        const resp = await fetchWithAuth(`${API_BASE}/api/inventory/bulk`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(toImport)
        });

        if (resp.ok) {
          toastSuccess('Products imported successfully');
          setShowSyncModal(false);
          fetchInventory();
        } else {
          toastError('Failed to import products');
        }
      } catch (err) {
        toastError('Error during bulk import');
      } finally {
        setIsSyncing(false);
      }
      return;
    }

    // Amazon sync (Background)
    setIsSyncing(true);
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/inventory/amazon/sync`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          start_date: syncStartDate,
          end_date: syncEndDate
        })
      });

      if (resp.ok) {
        toastSuccess('Amazon sync triggered');
        setShowSyncModal(false);
      } else {
        toastError('Failed to trigger Amazon sync');
      }
    } catch (err) {
      toastError('Error triggering Amazon sync');
    } finally {
      setIsSyncing(false);
    }
  };

  const toggleStagedSelection = (variantId: string) => {
    const next = new Set(selectedStagedIds);
    if (next.has(variantId)) next.delete(variantId);
    else next.add(variantId);
    setSelectedStagedIds(next);
  };

  const toggleAllStaged = () => {
    if (selectedStagedIds.size === stagedProducts.length) {
      setSelectedStagedIds(new Set());
    } else {
      setSelectedStagedIds(new Set(stagedProducts.map(p => p.mappings?.[0]?.external_variant_id!).filter(Boolean)));
    }
  };

  useEffect(() => {
    fetchInventory();
  }, []);

  const sortedAndFilteredItems = useMemo(() => {
    let result = [...items];

    // Filter
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      result = result.filter(item => 
        item.title.toLowerCase().includes(q) || 
        item.mi_sku.toLowerCase().includes(q) ||
        getSKUForPlatform(item.mappings, 'shopify').toLowerCase().includes(q) ||
        getSKUForPlatform(item.mappings, 'amazon').toLowerCase().includes(q)
      );
    }

    // Sort
    result.sort((a, b) => {
      if (sortBy === 'name-asc') {
        return a.title.localeCompare(b.title);
      } else if (sortBy === 'stock-desc') {
        return b.current_stock - a.current_stock;
      } else if (sortBy === 'stock-asc') {
        return a.current_stock - b.current_stock;
      } else if (sortBy === 'mi-sku-asc') {
        return a.mi_sku.localeCompare(b.mi_sku, undefined, { numeric: true });
      } else if (sortBy === 'mi-sku-desc') {
        return b.mi_sku.localeCompare(a.mi_sku, undefined, { numeric: true });
      }
      return 0;
    });

    return result;
  }, [items, searchQuery, sortBy]);

  return (
    <div className="tab-content staggered-fade-in">
      <div className="section-header" style={{ marginBottom: '2rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h2 style={{ fontSize: '1.5rem', fontWeight: 700, margin: 0 }}>Warehouse Authority</h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>Manage physical products and SKU mappings.</p>
        </div>
        <div style={{ display: 'flex', gap: '1rem' }}>
          {appConfigs?.show_sync_button === 'true' && userRole === 'admin' && (
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <button className="btn-secondary" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }} onClick={handleSyncShopify}>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M21 2v6h-6"></path><path d="M3 12a9 9 0 0 1 15-6.7L21 8"></path><path d="M3 22v-6h6"></path><path d="M21 12a9 9 0 0 1-15 6.7L3 16"></path>
                </svg>
                Sync Shopify
              </button>
              <button className="btn-secondary" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }} onClick={handleSyncAmazon}>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M21 2v6h-6"></path><path d="M3 12a9 9 0 0 1 15-6.7L21 8"></path><path d="M3 22v-6h6"></path><path d="M21 12a9 9 0 0 1-15 6.7L3 16"></path>
                </svg>
                Sync Amazon
              </button>
            </div>
          )}
          <button className="btn-primary" onClick={() => { setShowAddModal(true); fetchNextSKU(); }}>
            + Add New Product
          </button>
        </div>
      </div>

      <div style={{ 
        display: 'flex', 
        gap: '1rem', 
        marginBottom: '1.5rem', 
        alignItems: 'center',
        padding: '1rem',
        background: 'var(--glass-bg)',
        backdropFilter: 'var(--glass-blur)',
        borderRadius: 'var(--radius-lg)',
        border: '1px solid var(--border-color)',
        boxShadow: 'var(--shadow-sm)'
      }}>
        <div style={{ position: 'relative', flex: 1 }}>
          <input 
            type="text" 
            ref={searchInputRef}
            placeholder="Search by name, MI SKU, or platform SKU..." 
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            style={{ 
              paddingLeft: '2.5rem',
              paddingRight: searchQuery ? '2.5rem' : '1rem',
              height: '42px',
              fontSize: '0.85rem',
              border: '1px solid var(--border-color)',
              background: 'var(--bg-input)',
              width: '100%'
            }}
          />
          {searchQuery && (
            <button
              type="button"
              onClick={() => {
                setSearchQuery('');
                searchInputRef.current?.focus();
              }}
              aria-label="Clear search"
              title="Clear search"
              style={{
                position: 'absolute',
                right: '10px',
                top: '50%',
                transform: 'translateY(-50%)',
                background: 'transparent',
                border: 'none',
                color: 'var(--text-tertiary)',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                padding: '4px',
                borderRadius: '50%',
                transition: 'all 0.2s'
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.color = 'var(--text-primary)';
                e.currentTarget.style.background = 'var(--bg-hover)';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.color = 'var(--text-tertiary)';
                e.currentTarget.style.background = 'transparent';
              }}
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
            </button>
          )}
          <svg 
            width="16" 
            height="16" 
            viewBox="0 0 24 24" 
            fill="none" 
            stroke="var(--text-tertiary)" 
            strokeWidth="2" 
            strokeLinecap="round" 
            strokeLinejoin="round" 
            style={{ position: 'absolute', left: '1rem', top: '50%', transform: 'translateY(-50%)' }}
          >
            <circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/>
          </svg>
        </div>

        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
          <span style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', whiteSpace: 'nowrap' }}>Sort By:</span>
          <select 
            value={sortBy} 
            onChange={(e) => setSortBy(e.target.value as any)}
            style={{ 
              height: '42px', 
              fontSize: '0.85rem', 
              paddingRight: '2rem',
              border: '1px solid var(--border-color)',
              background: 'var(--bg-input)',
              minWidth: '160px'
            }}
          >
            <option value="mi-sku-asc">MI SKU (Low to High)</option>
            <option value="mi-sku-desc">MI SKU (High to Low)</option>
            <option value="name-asc">Alphabetical (A-Z)</option>
            <option value="stock-desc">Stock: High to Low</option>
            <option value="stock-asc">Stock: Low to High</option>
          </select>
        </div>
      </div>

      <div className="table-container glass-card-premium" style={{ border: 'none', boxShadow: 'var(--shadow-lg)' }}>
        <table className="premium-table">
          <thead>
            <tr>
              <th style={{ paddingLeft: '2rem' }}>Product</th>
              <th 
                style={{ cursor: 'pointer' }} 
                onClick={() => setSortBy(sortBy === 'mi-sku-asc' ? 'mi-sku-desc' : 'mi-sku-asc')}
              >
                MI SKU 
                {sortBy.startsWith('mi-sku') && (
                  <span style={{ marginLeft: '4px', fontSize: '0.7rem' }}>
                    {sortBy === 'mi-sku-asc' ? '↑' : '↓'}
                  </span>
                )}
              </th>
              <th>Shopify SKU</th>
              <th>Amazon SKU</th>
              <th style={{ paddingRight: '2rem' }}>Inventory Status</th>
            </tr>
          </thead>
          <tbody>
            {items.length === 0 && !isLoading ? (
              <tr>
                <td colSpan={5} style={{ textAlign: 'center', padding: '5rem' }}>
                  <div style={{ color: 'var(--text-tertiary)' }}>
                    <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" style={{ marginBottom: '1rem', opacity: 0.5 }}>
                      <path d="m7.5 4.27 9 5.15"/><path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z"/><path d="m3.3 7 8.7 5 8.7-5"/><path d="M12 22V12"/>
                    </svg>
                    <p>No products found in warehouse.</p>
                  </div>
                </td>
              </tr>
            ) : (
              sortedAndFilteredItems.map(item => (
                <tr key={item.id} className="hover-row" style={{ cursor: 'pointer' }} onClick={() => handleSetSelected(item)}>
                  <td style={{ paddingLeft: '2rem' }}>
                    <div style={{ fontWeight: 700, color: 'var(--text-primary)', maxWidth: '300px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {item.title}
                    </div>
                  </td>
                   <td>
                    <span className="badge-pill badge-pill-gray" style={{ fontSize: '0.7rem', fontWeight: 700 }}>
                      {item.mi_sku}
                    </span>
                  </td>
                  <td>
                    <code style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                      {getSKUForPlatform(item.mappings, 'shopify')}
                    </code>
                  </td>
                  <td>
                    <EditableSKU 
                      itemID={item.id} 
                      currentSKU={getSKUForPlatform(item.mappings, 'amazon')} 
                      platform="amazon" 
                    />
                  </td>
                  <td style={{ paddingRight: '2rem' }}>
                    <EditableStockPill id={item.id} initialStock={item.current_stock} />
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
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
                    autoFocus
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
                <button type="button" className="btn-secondary" onClick={() => setShowAddModal(false)} disabled={isSaving}>Cancel</button>
                <button type="submit" className="btn-primary" disabled={isSaving}>
                  {isSaving ? 'Creating...' : 'Create Product'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Product Details Modal */}
      {/* Product Details Modal */}
      {selectedProduct && (
        <div className="modal-overlay" onClick={() => setSelectedProduct(null)}>
          <div className="premium-modal" onClick={e => e.stopPropagation()} style={{ 
            maxWidth: '500px', 
            width: '95%',
            maxHeight: '85vh', 
            overflowY: 'auto',
            padding: '1.5rem',
            position: 'relative'
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem', position: 'sticky', top: 0, background: 'var(--surface-color)', zIndex: 10, paddingBottom: '0.5rem' }}>
              <h3 style={{ margin: 0, fontSize: '1.1rem' }}>Product Specifications</h3>
              <button className="toolbar-btn" onClick={() => setSelectedProduct(null)} style={{ fontSize: '1.2rem', padding: '0.5rem' }}>✕</button>
            </div>
            
            <div style={{ display: 'grid', gap: '1.25rem' }}>
              <div>
                <label style={{ fontSize: '0.65rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Full Title</label>
                <div style={{ fontSize: '1rem', fontWeight: 700, color: 'var(--text-primary)', marginTop: '0.4rem', lineHeight: 1.3 }}>
                  {selectedProduct.title}
                </div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div>
                  <label style={{ fontSize: '0.65rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Internal SKU</label>
                  <div style={{ marginTop: '0.4rem' }}>
                    <span className="badge-pill badge-pill-gray" style={{ fontSize: '0.8rem', fontWeight: 700, padding: '4px 10px' }}>
                      {selectedProduct.mi_sku}
                    </span>
                  </div>
                </div>
                <div>
                  <label style={{ fontSize: '0.65rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Shopify SKU</label>
                  <div style={{ marginTop: '0.4rem' }}>
                    <code style={{ fontSize: '0.9rem', color: 'var(--text-secondary)' }}>
                      {getSKUForPlatform(selectedProduct.mappings, 'shopify')}
                    </code>
                  </div>
                </div>
              </div>

              <div>
                <label style={{ fontSize: '0.65rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Marketing Description</label>
                <div 
                  className="spec-value desc-scroll" 
                  style={{ 
                    fontSize: '0.875rem', 
                    color: 'var(--text-secondary)', 
                    marginTop: '0.4rem', 
                    lineHeight: 1.5,
                    padding: '1rem',
                    background: 'var(--bg-input)',
                    borderRadius: '10px',
                    border: '1px solid var(--border-color)',
                    maxHeight: '160px',
                    overflowY: 'auto'
                  }}
                  dangerouslySetInnerHTML={{ __html: parseShopifyRichText(selectedProduct.description) || '<i>No marketing description available.</i>' }}
                />
              </div>

              {selectedProduct.specification && (
                <div>
                  <label style={{ fontSize: '0.65rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Technical Specifications</label>
                  <div 
                    className="spec-value desc-scroll" 
                    style={{ 
                      fontSize: '0.825rem', 
                      color: 'var(--text-secondary)', 
                      marginTop: '0.4rem', 
                      lineHeight: 1.4,
                      padding: '1rem',
                      background: 'var(--bg-input)',
                      borderRadius: '10px',
                      border: '1px solid var(--border-color)',
                      maxHeight: '160px',
                      overflowY: 'auto',
                      fontFamily: 'monospace'
                    }}
                    dangerouslySetInnerHTML={{ __html: parseShopifyRichText(selectedProduct.specification) }}
                  />
                </div>
              )}

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div className="metric-card-adaptive" style={{ padding: '1rem' }}>
                  <div style={{ fontSize: '0.6rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>Warehouse Location</div>
                  <div style={{ fontSize: '1rem', fontWeight: 700, marginTop: '2px' }}>Main Cluster</div>
                </div>
                <div className="metric-card-adaptive" style={{ padding: '1rem' }} onClick={() => fetchLogs(selectedProduct.id)}>
                  <div style={{ fontSize: '0.6rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase', display: 'flex', justifyContent: 'space-between' }}>
                    <span>Global Status</span>
                    <span style={{ color: 'var(--accent-color)', cursor: 'pointer' }}>View History →</span>
                  </div>
                  <div style={{ fontSize: '1rem', fontWeight: 700, marginTop: '2px', color: selectedProduct.current_stock > 0 ? 'var(--status-active)' : 'var(--status-danger)' }}>
                    {selectedProduct.current_stock > 0 ? 'Active Supply' : 'Out of Stock'}
                  </div>
                </div>
              </div>

              {/* Amazon Mapping Section */}
              <div style={{ 
                padding: '1.25rem', 
                background: 'var(--surface-color-secondary)', 
                borderRadius: '16px',
                border: '1px solid var(--border-color)',
                display: 'grid',
                gap: '1rem'
              }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <label style={{ fontSize: '0.7rem', fontWeight: 800, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', margin: 0 }}>
                    Amazon India Mappings
                  </label>
                  <span className="badge-pill badge-pill-gray" style={{ fontSize: '0.65rem', padding: '2px 8px' }}>
                    {(selectedProduct.mappings?.filter(m => m.platform === 'amazon') || []).length} Active
                  </span>
                </div>

                {(selectedProduct.mappings?.filter(m => m.platform === 'amazon') || []).length > 0 && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                    {(selectedProduct.mappings?.filter(m => m.platform === 'amazon') || []).map(mapping => (
                      <div 
                        key={mapping.id}
                        className="glass-card-premium"
                        style={{ 
                          display: 'flex', 
                          justifyContent: 'space-between', 
                          alignItems: 'center', 
                          padding: '0.6rem 0.8rem', 
                          borderRadius: '10px',
                          border: '1px solid var(--border-color)',
                          background: 'var(--bg-input)'
                        }}
                      >
                        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                          <span style={{ fontSize: '0.9rem', color: 'var(--accent-color)' }}>📦</span>
                          <code style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-primary)' }}>
                            {mapping.external_sku}
                          </code>
                        </div>
                        <button 
                          className="toolbar-btn" 
                          onClick={() => handleDeleteMapping(mapping.id)}
                          style={{ 
                            padding: '4px 8px', 
                            fontSize: '0.75rem', 
                            color: 'var(--status-danger)',
                            background: 'rgba(239, 68, 68, 0.1)',
                            borderRadius: '6px',
                            border: '1px solid rgba(239, 68, 68, 0.2)',
                            cursor: 'pointer',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '4px'
                          }}
                          onMouseEnter={(e) => {
                            e.currentTarget.style.background = 'var(--status-danger)';
                            e.currentTarget.style.color = '#fff';
                          }}
                          onMouseLeave={(e) => {
                            e.currentTarget.style.background = 'rgba(239, 68, 68, 0.1)';
                            e.currentTarget.style.color = 'var(--status-danger)';
                          }}
                        >
                          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                            <polyline points="3 6 5 6 21 6"></polyline>
                            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                          </svg>
                          Delete
                        </button>
                      </div>
                    ))}
                  </div>
                )}

                <div style={{ display: 'flex', gap: '0.5rem', marginTop: '0.25rem' }}>
                  <div style={{ position: 'relative', flex: 1 }}>
                    <input 
                      type="text" 
                      placeholder="Add new Amazon SKU..."
                      value={amazonSKUInput}
                      onChange={(e) => setAmazonSKUInput(e.target.value)}
                      style={{ 
                        height: '38px',
                        fontSize: '0.85rem',
                        border: '1px solid var(--border-color)',
                        background: 'var(--bg-input)',
                        paddingLeft: '1rem',
                        borderRadius: '8px'
                      }}
                    />
                  </div>
                  <button 
                    className="btn-primary" 
                    style={{ height: '38px', padding: '0 1rem', fontSize: '0.8rem', borderRadius: '8px' }}
                    onClick={() => handleAddAmazonMapping(selectedProduct.id)}
                  >
                    Add Mapping
                  </button>
                </div>
                
                <p style={{ fontSize: '0.7rem', color: 'var(--text-tertiary)', margin: 0 }}>
                  Linking Amazon SKUs enables real-time stock sync for this product. You can map multiple SKUs if the same product is sold under different listings.
                </p>
              </div>
            </div>

            <div style={{ marginTop: '1.75rem' }}>
              <button className="btn-primary" style={{ width: '100%' }} onClick={() => setSelectedProduct(null)}>Close Overview</button>
            </div>
          </div>
        </div>
      )}

      {/* Inventory Logs Modal */}
      {showLogsModal && selectedProduct && (
        <div className="modal-overlay" onClick={() => setShowLogsModal(false)}>
          <div className="premium-modal" onClick={e => e.stopPropagation()} style={{ maxWidth: '700px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
              <div>
                <h2 style={{ margin: 0 }}>Stock Movement History</h2>
                <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', marginTop: '4px' }}>
                  {selectedProduct.title} ({selectedProduct.mi_sku})
                </p>
              </div>
              <button className="toolbar-btn" onClick={() => setShowLogsModal(false)}>✕</button>
            </div>

            <div style={{ maxHeight: '450px', overflowY: 'auto', borderRadius: '12px', border: '1px solid var(--border-color)' }}>
              <table style={{ margin: 0 }}>
                <thead style={{ position: 'sticky', top: 0, background: 'var(--surface-color)', zIndex: 10 }}>
                  <tr>
                    <th>Date</th>
                    <th>Change</th>
                    <th>Reason</th>
                    <th>Platform</th>
                    <th>Order/Ref</th>
                  </tr>
                </thead>
                <tbody>
                  {productLogs.length === 0 ? (
                    <tr>
                      <td colSpan={5} style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-tertiary)' }}>
                        No movement history recorded yet.
                      </td>
                    </tr>
                  ) : (
                    productLogs.map(log => (
                      <tr key={log.id}>
                        <td style={{ fontSize: '0.75rem' }}>{new Date(log.created_at).toLocaleString()}</td>
                        <td>
                          <span style={{ 
                            color: log.delta > 0 ? 'var(--status-active)' : 'var(--status-danger)',
                            fontWeight: 700
                          }}>
                            {log.delta > 0 ? `+${log.delta}` : log.delta}
                          </span>
                        </td>
                        <td>
                          <span className="badge-pill badge-pill-gray" style={{ textTransform: 'capitalize' }}>
                            {log.reason.replace('_', ' ')}
                          </span>
                        </td>
                        <td>
                          <span style={{ fontSize: '0.8rem', textTransform: 'uppercase' }}>{log.platform}</span>
                        </td>
                        <td>
                          <code style={{ fontSize: '0.7rem' }}>{log.external_order_id || '—'}</code>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>

            <div className="modal-actions" style={{ marginTop: '1.5rem' }}>
              <button className="btn-primary" onClick={() => setShowLogsModal(false)}>Close Activity Log</button>
            </div>
          </div>
        </div>
      )}

      {/* Sync Modal */}
      {showSyncModal && (
        <div className="modal-overlay" onClick={() => setShowSyncModal(false)}>
          <div className="premium-modal wide" onClick={e => e.stopPropagation()} style={{ maxWidth: syncMode === 'shopify' ? '900px' : '500px' }}>
            <div className="modal-header-icon" style={{ background: syncMode === 'shopify' ? 'linear-gradient(135deg, #10b981, #059669)' : 'linear-gradient(135deg, #6366f1, #4f46e5)', marginBottom: '1.5rem', width: '50px', height: '50px', borderRadius: '12px', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'white' }}>
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                {syncMode === 'shopify' ? (
                  <path d="M21 2v6h-6M3 12a9 9 0 0 1 15-6.7L21 8M3 22v-6h6M21 12a9 9 0 0 1-15 6.7L3 16" />
                ) : (
                  <polyline points="23 4 23 10 17 10"></polyline>
                )}
                <path d={syncMode === 'shopify' ? "" : "M20.49 15a9 9 0 1 1-2.12-9.36L23 10"} />
              </svg>
            </div>
            
            <h2 style={{ marginBottom: '0.5rem' }}>{syncMode === 'shopify' ? 'Import from Shopify' : 'Amazon Order Sync'}</h2>
            <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', marginBottom: '1.5rem' }}>
              {syncMode === 'shopify' 
                ? 'Select the products you wish to import from your Shopify store into the local warehouse.' 
                : 'Poll Amazon India for recent orders to update local inventory levels.'}
            </p>
            
            {syncMode === 'shopify' ? (
              <div className="staged-products-container">
                {isSyncing && stagedProducts.length === 0 ? (
                  <div style={{ textAlign: 'center', padding: '3rem' }}>
                    <div className="spinner" style={{ marginBottom: '1rem' }} />
                    <p>Fetching variants from Shopify...</p>
                  </div>
                ) : (
                  <div className="staged-table-wrapper" style={{ maxHeight: '400px', overflowY: 'auto', border: '1px solid var(--border-color)', borderRadius: '8px' }}>
                    <table className="inventory-table">
                      <thead style={{ position: 'sticky', top: 0, zIndex: 1, background: 'var(--surface-color)' }}>
                        <tr>
                          <th style={{ width: '40px' }}>
                            <input type="checkbox" checked={selectedStagedIds.size === stagedProducts.length && stagedProducts.length > 0} onChange={toggleAllStaged} />
                          </th>
                          <th>Product / Variant</th>
                          <th>Shopify SKU</th>
                          <th>Current Stock</th>
                          <th>Status</th>
                        </tr>
                      </thead>
                      <tbody>
                        {stagedProducts.map((p, idx) => {
                          const variantId = p.mappings?.[0]?.external_variant_id;
                          const sku = p.mappings?.[0]?.external_sku;
                          const isAlreadyMapped = items.some(item => item.mappings?.some(m => m.external_sku === sku));
                          
                          return (
                            <tr key={idx} style={{ opacity: isAlreadyMapped ? 0.6 : 1 }}>
                              <td>
                                <input 
                                  type="checkbox" 
                                  disabled={isAlreadyMapped}
                                  checked={selectedStagedIds.has(variantId!)} 
                                  onChange={() => toggleStagedSelection(variantId!)} 
                                />
                              </td>
                              <td>
                                <div style={{ fontWeight: 600 }}>{p.title}</div>
                                <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>{variantId}</div>
                              </td>
                              <td><code style={{ fontSize: '0.75rem' }}>{sku}</code></td>
                              <td>{p.current_stock}</td>
                              <td>
                                {isAlreadyMapped ? (
                                  <span className="badge-pill badge-pill-success">Mapped</span>
                                ) : (
                                  <span className="badge-pill badge-pill-gray">New</span>
                                )}
                              </td>
                            </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            ) : (
              <div style={{ display: 'grid', gap: '1.25rem', marginBottom: '2rem' }}>
                <div className="input-group">
                  <label>From Date</label>
                  <input 
                    type="date" 
                    value={syncStartDate}
                    onChange={(e) => setSyncStartDate(e.target.value)}
                    style={{ background: 'var(--bg-input)', border: '1px solid var(--border-color)', color: 'var(--text-primary)' }}
                  />
                </div>
                <div className="input-group">
                  <label>To Date</label>
                  <input 
                    type="date" 
                    value={syncEndDate}
                    onChange={(e) => setSyncEndDate(e.target.value)}
                    style={{ background: 'var(--bg-input)', border: '1px solid var(--border-color)', color: 'var(--text-primary)' }}
                  />
                </div>
              </div>
            )}
            
            <div className="modal-actions" style={{ marginTop: '2rem' }}>
              <button className="btn-secondary" onClick={() => setShowSyncModal(false)}>Cancel</button>
              <button 
                className="btn-primary" 
                onClick={handleStartSync}
                disabled={isSyncing || (syncMode === 'shopify' && selectedStagedIds.size === 0)}
                style={{ minWidth: '160px' }}
              >
                {isSyncing ? 'Processing...' : (syncMode === 'shopify' ? `Import ${selectedStagedIds.size} Products` : 'Start Synchronization')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Products;
