import { useCallback, useEffect, useMemo, useState } from 'react'
import { Check, ChevronLeft, ChevronRight, CircleAlert, Edit3, Package, RefreshCw, Search, X } from 'lucide-react'
import { API_BASE } from '../../lib/api'

type InventoryPageProps = {
  token: string
  onUnauthorized: () => void
  embedded?: boolean
}

type InventoryMapping = {
  platform: string
  external_sku: string
}

type InventoryItem = {
  id: number
  mi_sku: string
  title: string
  current_stock: number
  price?: number
  mappings?: InventoryMapping[]
}

type InventoryResponse = {
  items?: InventoryItem[]
  total?: number
  message?: string
}

const pageSize = 10

function formatMoney(value: number | undefined) {
  return `₹${Number(value || 0).toLocaleString('en-IN', { maximumFractionDigits: 0 })}`
}

function getMapping(item: InventoryItem, platform: string) {
  return item.mappings?.find((mapping) => mapping.platform.toLowerCase() === platform)?.external_sku || '—'
}

function stockLabel(stock: number) {
  if (stock <= 0) return 'Out of stock'
  if (stock <= 10) return 'Low stock'
  return 'In stock'
}

function StockEditor({
  item,
  onSave,
  onCancel,
}: {
  item: InventoryItem
  onSave: (value: number) => void
  onCancel: () => void
}) {
  const [value, setValue] = useState(String(item.current_stock))

  const save = () => {
    const nextValue = Number.parseInt(value, 10)
    if (Number.isNaN(nextValue) || nextValue < 0) return
    onSave(nextValue)
  }

  return (
    <div className="inventory-stock-editor">
      <label className="sr-only" htmlFor={`stock-${item.id}`}>Stock units for {item.title}</label>
      <input
        id={`stock-${item.id}`}
        type="number"
        min="0"
        inputMode="numeric"
        value={value}
        onChange={(event) => setValue(event.target.value)}
        onKeyDown={(event) => {
          if (event.key === 'Enter') save()
          if (event.key === 'Escape') onCancel()
        }}
        autoFocus
      />
      <button type="button" className="table-action-button" aria-label="Save stock" onClick={save}>
        <Check size={14} aria-hidden="true" />
      </button>
      <button type="button" className="table-action-button" aria-label="Cancel stock edit" onClick={onCancel}>
        <X size={14} aria-hidden="true" />
      </button>
    </div>
  )
}

export function InventoryPage({ token, onUnauthorized, embedded = false }: InventoryPageProps) {
  const [items, setItems] = useState<InventoryItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [sort, setSort] = useState('mi-sku-asc')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState('')
  const [editingId, setEditingId] = useState<number | null>(null)
  const [isSavingStock, setIsSavingStock] = useState(false)
  const [notice, setNotice] = useState('')

  const fetchInventory = useCallback(async () => {
    setIsLoading(true)
    setError('')
    const query = new URLSearchParams({ page: String(page), limit: String(pageSize), sort })
    if (search.trim()) query.set('search', search.trim())

    try {
      const response = await fetch(`${API_BASE}/api/inventory?${query.toString()}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (response.status === 401) {
        onUnauthorized()
        throw new Error('Your session has expired. Please sign in again.')
      }
      if (!response.ok) throw new Error(`Inventory request failed with status ${response.status}`)
      const data = await response.json() as InventoryResponse
      setItems(data.items || [])
      setTotal(data.total || 0)
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to load inventory')
      setItems([])
      setTotal(0)
    } finally {
      setIsLoading(false)
    }
  }, [onUnauthorized, page, search, sort, token])

  useEffect(() => {
    void fetchInventory()
  }, [fetchInventory])

  const totalPages = Math.max(Math.ceil(total / pageSize), 1)
  const lowStockCount = useMemo(() => items.filter((item) => item.current_stock > 0 && item.current_stock <= 10).length, [items])
  const outOfStockCount = useMemo(() => items.filter((item) => item.current_stock <= 0).length, [items])

  const updateStock = async (item: InventoryItem, value: number) => {
    setIsSavingStock(true)
    setNotice('')
    try {
      const response = await fetch(`${API_BASE}/api/inventory/stock?id=${item.id}&val=${value}`, {
        method: 'PUT',
        headers: { Authorization: `Bearer ${token}` },
      })
      if (response.status === 401) {
        onUnauthorized()
        throw new Error('Your session has expired. Please sign in again.')
      }
      if (!response.ok) throw new Error('Stock could not be updated')
      setEditingId(null)
      setNotice(`${item.mi_sku} stock updated`)
      await fetchInventory()
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to update stock')
    } finally {
      setIsSavingStock(false)
    }
  }

  return (
    <section className="workspace-page inventory-page" aria-labelledby="inventory-heading">
      {!embedded && <header className="workspace-page-header">
        <div>
          <p className="eyebrow">Operations / Inventory</p>
          <h2 id="inventory-heading">Know what is ready to ship.</h2>
          <p>Keep the warehouse catalogue clear, spot stock risk early, and update counts without opening a separate detail screen.</p>
        </div>
        <button className="secondary-button" type="button" onClick={() => void fetchInventory()} disabled={isLoading}>
          <RefreshCw size={15} className={isLoading ? 'spin' : undefined} aria-hidden="true" />
          Refresh catalogue
        </button>
      </header>}

      <div className="inventory-summary-grid" aria-label="Inventory summary">
        <div className="inventory-summary-card"><span className="metric-label">Catalogue</span><strong>{total.toLocaleString('en-IN')}</strong><small>Total products</small></div>
        <div className="inventory-summary-card"><span className="metric-label">Low stock</span><strong>{lowStockCount}</strong><small>On this page</small></div>
        <div className="inventory-summary-card"><span className="metric-label">Unavailable</span><strong>{outOfStockCount}</strong><small>On this page</small></div>
      </div>

      <div className="inventory-toolbar" aria-label="Inventory filters">
        <label className="orders-search">
          <Search size={16} aria-hidden="true" />
          <span className="sr-only">Search inventory</span>
          <input value={search} onChange={(event) => { setSearch(event.target.value); setPage(1) }} placeholder="Search SKU, product, or channel SKU" />
        </label>
        <label className="compact-select">
          <span className="sr-only">Sort inventory</span>
          <select value={sort} onChange={(event) => { setSort(event.target.value); setPage(1) }}>
            <option value="mi-sku-asc">SKU: A to Z</option>
            <option value="mi-sku-desc">SKU: Z to A</option>
            <option value="name-asc">Product: A to Z</option>
            <option value="stock-asc">Stock: low to high</option>
            <option value="stock-desc">Stock: high to low</option>
          </select>
        </label>
        <span className="filter-count">{total.toLocaleString('en-IN')} products</span>
      </div>

      {error && <div className="dashboard-error" role="alert"><CircleAlert size={18} aria-hidden="true" /><span>{error}</span><button type="button" onClick={() => void fetchInventory()}>Try again</button></div>}
      {notice && <div className="inventory-notice" role="status">{notice}</div>}

      <div className="orders-card inventory-card">
        <div className="orders-card-heading">
          <div><p className="eyebrow">Warehouse authority</p><h3>Products</h3></div>
          <span className="orders-card-meta">Page {page} of {totalPages}</span>
        </div>
        <div className="orders-table-wrap">
          <table className="orders-table inventory-table">
            <caption className="sr-only">Inventory products and current stock</caption>
            <thead><tr><th>Product</th><th>Stock</th><th>Price</th><th>Shopify SKU</th><th>Amazon SKU</th></tr></thead>
            <tbody>
              {isLoading ? <tr><td className="table-state" colSpan={5}>Loading inventory…</td></tr> : items.length === 0 ? <tr><td className="table-state" colSpan={5}>No products match this search.</td></tr> : items.map((item) => (
                <tr key={item.id}>
                  <td><div className="inventory-product-cell"><span className="inventory-product-icon"><Package size={16} aria-hidden="true" /></span><span><strong>{item.title || 'Untitled product'}</strong><small>{item.mi_sku}</small></span></div></td>
                  <td>
                    {editingId === item.id ? <StockEditor item={item} onSave={(value) => { if (!isSavingStock) void updateStock(item, value) }} onCancel={() => setEditingId(null)} /> : (
                      <button className={`inventory-stock-button inventory-stock-${item.current_stock <= 0 ? 'empty' : item.current_stock <= 10 ? 'low' : 'ready'}`} type="button" onClick={() => setEditingId(item.id)} disabled={isSavingStock}>
                        <span className="inventory-stock-dot" aria-hidden="true" />{item.current_stock.toLocaleString('en-IN')}<small>{stockLabel(item.current_stock)}</small><Edit3 size={13} aria-hidden="true" />
                      </button>
                    )}
                  </td>
                  <td className="table-money">{formatMoney(item.price)}</td>
                  <td className="table-muted">{getMapping(item, 'shopify')}</td>
                  <td className="table-muted">{getMapping(item, 'amazon')}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div className="orders-pagination">
          <span>Showing {items.length ? (page - 1) * pageSize + 1 : 0}–{Math.min(page * pageSize, total)} of {total}</span>
          <div><button type="button" aria-label="Previous inventory page" disabled={page <= 1 || isLoading} onClick={() => setPage((current) => current - 1)}><ChevronLeft size={16} aria-hidden="true" /></button><button type="button" aria-label="Next inventory page" disabled={page >= totalPages || isLoading} onClick={() => setPage((current) => current + 1)}><ChevronRight size={16} aria-hidden="true" /></button></div>
        </div>
      </div>
    </section>
  )
}
