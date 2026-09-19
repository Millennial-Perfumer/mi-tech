import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useRealtimeRefresh } from '../../lib/realtime'
import { ArrowDownRight, ArrowUpRight, Check, ChevronLeft, ChevronRight, CircleAlert, Edit3, History, Package, RefreshCw, Search, Trash2, X } from 'lucide-react'
import { API_BASE } from '../../lib/api'
import { apiJson, apiRequest, arrayFrom } from '../../lib/http'

type InventoryPageProps = {
  token: string
  onUnauthorized: () => void
  embedded?: boolean
}

type InventoryMapping = {
  id?: number
  platform: string
  external_sku: string
  external_variant_id?: string
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

type InventoryLog = {
  id: number
  inventory_item_id: number
  delta: number
  reason: string
  platform: string
  external_order_id?: string | null
  order_id?: number | null
  stock_before?: number | null
  stock_after?: number | null
  created_at: string
}

type InventoryLogPageResponse = {
  items?: InventoryLog[]
  page?: number
  limit?: number
  total?: number
}

const pageSize = 10
const recentLogLimit = 5
const historyPageSize = 10

type SyncMode = 'shopify' | 'amazon'

function stagedVariantId(item: InventoryItem) {
  return item.mappings?.find((mapping) => mapping.external_variant_id)?.external_variant_id || item.mappings?.[0]?.external_variant_id || ''
}

function stagedSku(item: InventoryItem) {
  return item.mappings?.find((mapping) => mapping.external_sku)?.external_sku || item.mappings?.[0]?.external_sku || '—'
}

function formatMoney(value: number | undefined) {
  return `₹${Number(value || 0).toLocaleString('en-IN', { maximumFractionDigits: 0 })}`
}

function getMappingRecord(item: InventoryItem, platform: string) {
  return item.mappings?.find((mapping) => mapping.platform.toLowerCase() === platform)
}

function getMapping(item: InventoryItem, platform: string) {
  return getMappingRecord(item, platform)?.external_sku || '—'
}

function stockLabel(stock: number) {
  if (stock <= 0) return 'Out of stock'
  if (stock <= 10) return 'Low stock'
  return 'In stock'
}

function formatMovementDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return 'Unknown date'
  return date.toLocaleString('en-IN', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function movementReason(reason: string) {
  return reason
    .replace(/[_-]+/g, ' ')
    .replace(/\b\w/g, (character) => character.toUpperCase())
}

function movementKind(delta: number) {
  if (delta > 0) return 'addition'
  if (delta < 0) return 'deduction'
  return 'correction'
}

function movementContext(log: InventoryLog) {
  if (log.external_order_id) return `${log.platform} · Order ${log.external_order_id}`
  if (log.platform) return log.platform
  return 'Internal'
}

function StockActivity({
  logs,
  totalCount,
  isLoading,
  error,
  accessDenied,
  onViewFullHistory,
  onRetry,
}: {
  logs: InventoryLog[]
  totalCount: number
  isLoading: boolean
  error: string
  accessDenied: boolean
  onViewFullHistory: () => void
  onRetry: () => void
}) {
  return (
    <section className="inventory-activity-section" aria-labelledby="inventory-activity-heading">
      <div className="inventory-activity-heading">
        <div>
          <p className="eyebrow">Stock ledger</p>
          <h3 id="inventory-activity-heading">Recent stock activity</h3>
          <p>Track deductions, additions, and corrections for this product.</p>
        </div>
        {totalCount > 0 && <button className="table-link-button" type="button" onClick={onViewFullHistory}>
          <History size={14} aria-hidden="true" />
          View full history
        </button>}
      </div>

      {accessDenied ? <div className="inventory-history-note" role="note"><CircleAlert size={16} aria-hidden="true" /><span>Stock history is available to administrators only.</span></div> : error ? <div className="inventory-history-error" role="alert"><CircleAlert size={16} aria-hidden="true" /><span>{error}</span><button type="button" onClick={onRetry}>Try again</button></div> : isLoading ? <p className="table-state inventory-activity-state">Loading stock activity…</p> : logs.length === 0 ? <p className="table-state inventory-activity-state">No stock movements recorded yet.</p> : <div className="inventory-activity-list" aria-label="Stock movement history">
        {logs.map((log) => {
          const kind = movementKind(log.delta)
          const deltaLabel = log.delta > 0 ? `+${log.delta}` : String(log.delta)
          const hasStockRange = log.stock_before !== null && log.stock_before !== undefined && log.stock_after !== null && log.stock_after !== undefined
          return <article className={`inventory-activity-row inventory-activity-${kind}`} key={log.id}>
            <span className="inventory-activity-icon" aria-hidden="true">{log.delta > 0 ? <ArrowUpRight size={16} /> : log.delta < 0 ? <ArrowDownRight size={16} /> : <History size={15} />}</span>
            <div className="inventory-activity-main">
              <div className="inventory-activity-title"><strong>{movementReason(log.reason)}</strong><span>{movementContext(log)}</span></div>
              <small>{formatMovementDate(log.created_at)}</small>
            </div>
            <div className="inventory-activity-change">
              <strong>{deltaLabel}</strong>
              {hasStockRange && <small>{log.stock_before} → {log.stock_after}</small>}
            </div>
          </article>
        })}
      </div>}
    </section>
  )
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
  const [manualSyncEnabled, setManualSyncEnabled] = useState(false)
  const [isLoadingSyncConfig, setIsLoadingSyncConfig] = useState(true)
  const [isSyncing, setIsSyncing] = useState(false)
  const [syncMode, setSyncMode] = useState<SyncMode | null>(null)
  const [stagedProducts, setStagedProducts] = useState<InventoryItem[]>([])
  const [selectedStagedIds, setSelectedStagedIds] = useState<Set<string>>(new Set())
  const [selectedProduct, setSelectedProduct] = useState<InventoryItem | null>(null)
  const [editMiSku, setEditMiSku] = useState('')
  const [editShopifySku, setEditShopifySku] = useState('')
  const [editAmazonSku, setEditAmazonSku] = useState('')
  const [isSavingProduct, setIsSavingProduct] = useState(false)
  const [isDeletingProduct, setIsDeletingProduct] = useState(false)
  const [productModalError, setProductModalError] = useState('')
  const [productLogs, setProductLogs] = useState<InventoryLog[]>([])
  const [isLoadingProductLogs, setIsLoadingProductLogs] = useState(false)
  const [productLogsError, setProductLogsError] = useState('')
  const [productLogsAccessDenied, setProductLogsAccessDenied] = useState(false)
  const [productLogsTotal, setProductLogsTotal] = useState(0)
  const [isHistoryModalOpen, setIsHistoryModalOpen] = useState(false)
  const [historyLogs, setHistoryLogs] = useState<InventoryLog[]>([])
  const [historyPage, setHistoryPage] = useState(1)
  const [historyTotal, setHistoryTotal] = useState(0)
  const [isLoadingHistory, setIsLoadingHistory] = useState(false)
  const [historyError, setHistoryError] = useState('')
  const [historyAccessDenied, setHistoryAccessDenied] = useState(false)
  const productLogsRequest = useRef(0)
  const historyRequest = useRef(0)

  const fetchSyncConfig = useCallback(async () => {
    setIsLoadingSyncConfig(true)
    try {
      const data = await apiJson<unknown>(token, onUnauthorized, '/api/configs')
      const configs = arrayFrom(data, 'configs')
      const manualSync = configs.find((config) => String(config.key || '') === 'show_sync_button')
      setManualSyncEnabled(String(manualSync?.value || '').trim().toLowerCase() === 'true')
    } catch {
      // Keep the controls hidden when the configuration cannot be read.
      setManualSyncEnabled(false)
    } finally {
      setIsLoadingSyncConfig(false)
    }
  }, [onUnauthorized, token])

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
  useRealtimeRefresh(['inventory.changed', 'orders.changed', 'b2b.changed'], () => void fetchInventory())

  useEffect(() => {
    void fetchSyncConfig()
  }, [fetchSyncConfig])

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

  const loadProductLogs = async (itemId: number) => {
    const requestId = productLogsRequest.current + 1
    productLogsRequest.current = requestId
    setIsLoadingProductLogs(true)
    setProductLogs([])
    setProductLogsTotal(0)
    setProductLogsError('')
    setProductLogsAccessDenied(false)

    try {
      const response = await fetch(`${API_BASE}/api/inventory/logs?id=${itemId}&page=1&limit=${recentLogLimit}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (response.status === 401) {
        onUnauthorized()
        throw new Error('Your session has expired. Please sign in again.')
      }
      if (response.status === 403) {
        if (requestId === productLogsRequest.current) setProductLogsAccessDenied(true)
        return
      }
      if (!response.ok) throw new Error(`Stock history request failed with status ${response.status}`)
      const data = await response.json() as InventoryLogPageResponse | InventoryLog[]
      const logs = Array.isArray(data) ? data : data.items || []
      const totalCount = Array.isArray(data) ? logs.length : data.total ?? logs.length
      if (requestId === productLogsRequest.current) {
        setProductLogs(logs)
        setProductLogsTotal(totalCount)
      }
    } catch (caughtError) {
      if (requestId === productLogsRequest.current) setProductLogsError(caughtError instanceof Error ? caughtError.message : 'Unable to load stock activity')
    } finally {
      if (requestId === productLogsRequest.current) setIsLoadingProductLogs(false)
    }
  }

  const loadHistoryPage = async (itemId: number, nextPage: number) => {
    const requestId = historyRequest.current + 1
    historyRequest.current = requestId
    setIsLoadingHistory(true)
    setHistoryError('')
    setHistoryAccessDenied(false)

    try {
      const response = await fetch(`${API_BASE}/api/inventory/logs?id=${itemId}&page=${nextPage}&limit=${historyPageSize}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (response.status === 401) {
        onUnauthorized()
        throw new Error('Your session has expired. Please sign in again.')
      }
      if (response.status === 403) {
        if (requestId === historyRequest.current) setHistoryAccessDenied(true)
        return
      }
      if (!response.ok) throw new Error(`Stock history request failed with status ${response.status}`)
      const data = await response.json() as InventoryLogPageResponse | InventoryLog[]
      const logs = Array.isArray(data) ? data : data.items || []
      const totalCount = Array.isArray(data) ? logs.length : data.total ?? logs.length
      if (requestId === historyRequest.current) {
        setHistoryLogs(logs)
        setHistoryPage(nextPage)
        setHistoryTotal(totalCount)
      }
    } catch (caughtError) {
      if (requestId === historyRequest.current) setHistoryError(caughtError instanceof Error ? caughtError.message : 'Unable to load stock history')
    } finally {
      if (requestId === historyRequest.current) setIsLoadingHistory(false)
    }
  }

  const openProductDetails = (item: InventoryItem) => {
    setSelectedProduct(item)
    setEditMiSku(item.mi_sku)
    setEditShopifySku(getMappingRecord(item, 'shopify')?.external_sku || '')
    setEditAmazonSku(getMappingRecord(item, 'amazon')?.external_sku || '')
    setProductModalError('')
    setIsHistoryModalOpen(false)
    void loadProductLogs(item.id)
  }

  const openHistoryModal = () => {
    if (!selectedProduct) return
    setHistoryLogs([])
    setHistoryPage(1)
    setHistoryTotal(0)
    setHistoryError('')
    setHistoryAccessDenied(false)
    setIsHistoryModalOpen(true)
    void loadHistoryPage(selectedProduct.id, 1)
  }

  const historyTotalPages = Math.max(Math.ceil(historyTotal / historyPageSize), 1)
  const historyStart = historyLogs.length ? (historyPage - 1) * historyPageSize + 1 : 0
  const historyEnd = historyLogs.length ? historyStart + historyLogs.length - 1 : 0

  const saveChannelSku = async (item: InventoryItem, platform: string, sku: string) => {
    const existing = getMappingRecord(item, platform)
    const nextSku = sku.trim()

    if (existing?.external_sku === nextSku) return

    if (existing?.id) {
      await apiRequest(token, onUnauthorized, `/api/inventory/map?id=${existing.id}`, { method: 'DELETE' })
    }

    if (!nextSku) return

    await apiRequest(token, onUnauthorized, '/api/inventory/map', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        internal_item_id: item.id,
        platform,
        external_sku: nextSku,
        variant_id: existing?.external_variant_id || '',
      }),
    })
  }

  const saveProductDetails = async () => {
    if (!selectedProduct) return
    const nextMiSku = editMiSku.trim()
    if (!nextMiSku) {
      setProductModalError('MI SKU is required')
      return
    }

    setIsSavingProduct(true)
    setProductModalError('')
    setError('')
    try {
      if (nextMiSku !== selectedProduct.mi_sku) {
        await apiRequest(token, onUnauthorized, '/api/inventory/item', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ id: selectedProduct.id, mi_sku: nextMiSku }),
        })
      }
      await saveChannelSku(selectedProduct, 'shopify', editShopifySku)
      await saveChannelSku(selectedProduct, 'amazon', editAmazonSku)
      setNotice(`${nextMiSku} updated`)
      setSelectedProduct(null)
      await fetchInventory()
    } catch (caughtError) {
      setProductModalError(caughtError instanceof Error ? caughtError.message : 'Unable to save product details')
    } finally {
      setIsSavingProduct(false)
    }
  }

  const deleteSelectedProduct = async () => {
    if (!selectedProduct || !window.confirm(`Delete ${selectedProduct.mi_sku}? This removes the local product and its SKU mappings.`)) return

    setIsDeletingProduct(true)
    setProductModalError('')
    setError('')
    try {
      await apiRequest(token, onUnauthorized, '/api/inventory/item/delete', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: selectedProduct.id }),
      })
      setNotice(`${selectedProduct.mi_sku} deleted`)
      setSelectedProduct(null)
      await fetchInventory()
    } catch (caughtError) {
      setProductModalError(caughtError instanceof Error ? caughtError.message : 'Unable to delete product')
    } finally {
      setIsDeletingProduct(false)
    }
  }

  const startShopifySync = async () => {
    setSyncMode('shopify')
    setStagedProducts([])
    setSelectedStagedIds(new Set())
    setIsSyncing(true)
    setError('')
    try {
      const data = await apiJson<unknown>(token, onUnauthorized, '/api/inventory/sync-shopify', { method: 'POST' })
      setStagedProducts(Array.isArray(data) ? data as InventoryItem[] : [])
    } catch (caughtError) {
      setSyncMode(null)
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to fetch products from Shopify')
    } finally {
      setIsSyncing(false)
    }
  }

  const startAmazonSync = async () => {
    setSyncMode('amazon')
    setIsSyncing(true)
    setError('')
    try {
      await apiRequest(token, onUnauthorized, '/api/inventory/amazon/sync', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({}),
      })
      setNotice('Amazon sync triggered. Inventory will update in the background.')
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to trigger Amazon sync')
    } finally {
      setIsSyncing(false)
      setSyncMode(null)
    }
  }

  const toggleStagedProduct = (variantId: string) => {
    setSelectedStagedIds((current) => {
      const next = new Set(current)
      if (next.has(variantId)) next.delete(variantId)
      else next.add(variantId)
      return next
    })
  }

  const toggleAllStagedProducts = () => {
    const selectableIds = stagedProducts
      .filter((product) => !items.some((item) => item.mappings?.some((mapping) => mapping.external_sku === stagedSku(product))))
      .map(stagedVariantId)
      .filter(Boolean)
    setSelectedStagedIds((current) => current.size === selectableIds.length ? new Set() : new Set(selectableIds))
  }

  const importSelectedShopifyProducts = async () => {
    const selected = stagedProducts.filter((product) => selectedStagedIds.has(stagedVariantId(product)))
    if (selected.length === 0) {
      setError('Select at least one Shopify product to import')
      return
    }

    setIsSyncing(true)
    setError('')
    try {
      await apiRequest(token, onUnauthorized, '/api/inventory/bulk', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(selected),
      })
      setNotice(`${selected.length} Shopify product${selected.length === 1 ? '' : 's'} imported`)
      setSyncMode(null)
      setStagedProducts([])
      setSelectedStagedIds(new Set())
      await fetchInventory()
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to import Shopify products')
    } finally {
      setIsSyncing(false)
    }
  }

  const syncActions = !isLoadingSyncConfig && manualSyncEnabled && (
    <div className="inventory-sync-actions" aria-label="Manual inventory synchronization">
      <button className="secondary-button" type="button" onClick={() => void startShopifySync()} disabled={isSyncing}>
        <RefreshCw size={15} className={isSyncing && syncMode === 'shopify' ? 'spin' : undefined} aria-hidden="true" />
        Sync Shopify
      </button>
      <button className="secondary-button" type="button" onClick={() => void startAmazonSync()} disabled={isSyncing}>
        <RefreshCw size={15} className={isSyncing && syncMode === 'amazon' ? 'spin' : undefined} aria-hidden="true" />
        Sync Amazon
      </button>
    </div>
  )

  return (
    <section className="workspace-page inventory-page" aria-labelledby="inventory-heading">
      {!embedded && <header className="workspace-page-header">
        <div>
          <p className="eyebrow">Operations / Inventory</p>
          <h2 id="inventory-heading">Products</h2>
          <p>Review products, stock levels, and inventory details.</p>
        </div>
        <div className="inventory-page-actions">
          {syncActions}
          <button className="secondary-button" type="button" onClick={() => void fetchInventory()} disabled={isLoading}>
            <RefreshCw size={15} className={isLoading ? 'spin' : undefined} aria-hidden="true" />
            Refresh catalogue
          </button>
        </div>
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
        {embedded && syncActions}
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
                  <td><button className="inventory-product-link" type="button" onClick={() => openProductDetails(item)} aria-label={`Open details for ${item.title || item.mi_sku}`}><span className="inventory-product-cell"><span className="inventory-product-icon"><Package size={16} aria-hidden="true" /></span><span><strong>{item.title || 'Untitled product'}</strong><small>{item.mi_sku}</small></span></span></button></td>
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

      {syncMode === 'shopify' && (
        <div className="modal-scrim" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget && !isSyncing) setSyncMode(null) }}>
          <section className="modal-card inventory-sync-modal" role="dialog" aria-modal="true" aria-labelledby="shopify-sync-heading" onClick={(event) => event.stopPropagation()}>
            <div className="modal-heading">
              <div>
                <p className="eyebrow">Inventory synchronization</p>
                <h2 id="shopify-sync-heading">Import from Shopify</h2>
              </div>
              <button className="icon-button" type="button" aria-label="Close Shopify sync" onClick={() => { if (!isSyncing) setSyncMode(null) }} disabled={isSyncing}>
                <X size={19} aria-hidden="true" />
              </button>
            </div>
            <p>Select the Shopify products to add to the local warehouse.</p>
            {isSyncing && stagedProducts.length === 0 ? <p className="table-state">Fetching products from Shopify…</p> : stagedProducts.length === 0 ? <p className="table-state">No Shopify products available to import.</p> : (
              <div className="orders-table-wrap inventory-sync-table-wrap">
                <table className="orders-table">
                  <thead><tr><th><input type="checkbox" aria-label="Select all Shopify products" checked={selectedStagedIds.size > 0 && selectedStagedIds.size === stagedProducts.filter((product) => !items.some((item) => item.mappings?.some((mapping) => mapping.external_sku === stagedSku(product)))).length} onChange={toggleAllStagedProducts} /></th><th>Product / variant</th><th>Shopify SKU</th><th>Status</th></tr></thead>
                  <tbody>{stagedProducts.map((product) => {
                    const variantId = stagedVariantId(product)
                    const sku = stagedSku(product)
                    const alreadyMapped = items.some((item) => item.mappings?.some((mapping) => mapping.external_sku === sku))
                    return <tr key={variantId || `${product.id}-${sku}`}><td><input type="checkbox" aria-label={`Select ${product.title || sku}`} disabled={alreadyMapped || !variantId} checked={selectedStagedIds.has(variantId)} onChange={() => toggleStagedProduct(variantId)} /></td><td><strong>{product.title || 'Untitled product'}</strong><small>{variantId || 'No variant ID'}</small></td><td>{sku}</td><td>{alreadyMapped ? 'Already imported' : 'Ready to import'}</td></tr>
                  })}</tbody>
                </table>
              </div>
            )}
            <div className="modal-actions">
              <button className="secondary-button" type="button" onClick={() => setSyncMode(null)} disabled={isSyncing}>Cancel</button>
              <button className="primary-button" type="button" onClick={() => void importSelectedShopifyProducts()} disabled={isSyncing || stagedProducts.length === 0}>{isSyncing ? 'Importing…' : `Import ${selectedStagedIds.size || ''} product${selectedStagedIds.size === 1 ? '' : 's'}`}</button>
            </div>
          </section>
        </div>
      )}

      {selectedProduct && !isHistoryModalOpen && (
        <div className="modal-scrim" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget && !isSavingProduct && !isDeletingProduct) setSelectedProduct(null) }}>
          <section className="modal-card inventory-product-modal" role="dialog" aria-modal="true" aria-labelledby="inventory-product-heading" onClick={(event) => event.stopPropagation()}>
            <div className="modal-heading">
              <div>
                <p className="eyebrow">Product specifications</p>
                <h2 id="inventory-product-heading">{selectedProduct.title || 'Untitled product'}</h2>
              </div>
              <button className="icon-button" type="button" aria-label="Close product details" onClick={() => { if (!isSavingProduct && !isDeletingProduct) setSelectedProduct(null) }} disabled={isSavingProduct || isDeletingProduct}>
                <X size={19} aria-hidden="true" />
              </button>
            </div>
            <p className="inventory-product-modal-copy">Update the warehouse SKU and channel mappings for this product.</p>
            <div className="inventory-product-detail-grid">
              <label className="form-field"><span>MI SKU</span><input value={editMiSku} onChange={(event) => setEditMiSku(event.target.value)} disabled={isSavingProduct || isDeletingProduct} /></label>
              <label className="form-field"><span>Shopify SKU</span><input value={editShopifySku} onChange={(event) => setEditShopifySku(event.target.value)} placeholder="Not mapped" disabled={isSavingProduct || isDeletingProduct} /></label>
              <label className="form-field"><span>Amazon SKU</span><input value={editAmazonSku} onChange={(event) => setEditAmazonSku(event.target.value)} placeholder="Not mapped" disabled={isSavingProduct || isDeletingProduct} /></label>
            </div>
            <StockActivity
              logs={productLogs}
              totalCount={productLogsTotal}
              isLoading={isLoadingProductLogs}
              error={productLogsError}
              accessDenied={productLogsAccessDenied}
              onViewFullHistory={openHistoryModal}
              onRetry={() => { if (selectedProduct) void loadProductLogs(selectedProduct.id) }}
            />
            {productModalError && <p className="modal-form-error" role="alert"><CircleAlert size={15} aria-hidden="true" />{productModalError}</p>}
            <div className="modal-actions">
              <button className="secondary-button danger-link" type="button" onClick={() => void deleteSelectedProduct()} disabled={isSavingProduct || isDeletingProduct}><Trash2 size={15} aria-hidden="true" />{isDeletingProduct ? 'Deleting…' : 'Delete product'}</button>
              <button className="secondary-button" type="button" onClick={() => setSelectedProduct(null)} disabled={isSavingProduct || isDeletingProduct}>Cancel</button>
              <button className="primary-button" type="button" onClick={() => void saveProductDetails()} disabled={isSavingProduct || isDeletingProduct}>{isSavingProduct ? 'Saving…' : 'Save changes'}</button>
            </div>
          </section>
        </div>
      )}

      {selectedProduct && isHistoryModalOpen && (
        <div className="modal-scrim" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget && !isLoadingHistory) setIsHistoryModalOpen(false) }}>
          <section className="modal-card inventory-history-modal" role="dialog" aria-modal="true" aria-labelledby="inventory-history-heading" onClick={(event) => event.stopPropagation()}>
            <div className="modal-heading">
              <div>
                <p className="eyebrow">Stock ledger</p>
                <h2 id="inventory-history-heading">Inventory history</h2>
                <p className="inventory-history-product">{selectedProduct.title || selectedProduct.mi_sku} · {selectedProduct.mi_sku}</p>
              </div>
              <button className="icon-button" type="button" aria-label="Close inventory history" onClick={() => setIsHistoryModalOpen(false)} disabled={isLoadingHistory}>
                <X size={19} aria-hidden="true" />
              </button>
            </div>

            {historyAccessDenied ? <div className="inventory-history-note" role="note"><CircleAlert size={16} aria-hidden="true" /><span>Stock history is available to administrators only.</span></div> : historyError ? <div className="inventory-history-error" role="alert"><CircleAlert size={16} aria-hidden="true" /><span>{historyError}</span><button type="button" onClick={() => void loadHistoryPage(selectedProduct.id, historyPage)}>Try again</button></div> : isLoadingHistory && historyLogs.length === 0 ? <p className="table-state inventory-activity-state">Loading stock history…</p> : historyLogs.length === 0 ? <p className="table-state inventory-activity-state">No stock movements recorded yet.</p> : (
              <>
                <div className="orders-table-wrap inventory-history-table-wrap">
                  <table className="orders-table inventory-history-table">
                    <caption className="sr-only">Inventory movement history for {selectedProduct.mi_sku}</caption>
                    <thead><tr><th>Date</th><th>Change</th><th>Reason</th><th>Channel / reference</th><th>Stock</th></tr></thead>
                    <tbody>{historyLogs.map((log) => {
                      const kind = movementKind(log.delta)
                      const deltaLabel = log.delta > 0 ? `+${log.delta}` : String(log.delta)
                      const hasStockRange = log.stock_before !== null && log.stock_before !== undefined && log.stock_after !== null && log.stock_after !== undefined
                      return <tr key={log.id}>
                        <td className="inventory-history-date" data-label="Date">{formatMovementDate(log.created_at)}</td>
                        <td data-label="Change"><span className={`inventory-history-change inventory-history-change-${kind}`}>{deltaLabel}</span></td>
                        <td data-label="Reason"><strong>{movementReason(log.reason)}</strong></td>
                        <td className="inventory-history-reference" data-label="Channel / reference">{movementContext(log)}</td>
                        <td className="inventory-history-stock" data-label="Stock">{hasStockRange ? `${log.stock_before} → ${log.stock_after}` : '—'}</td>
                      </tr>
                    })}</tbody>
                  </table>
                </div>
                <div className="orders-pagination inventory-history-pagination">
                  <span>Showing {historyStart}–{historyEnd} of {historyTotal}</span>
                  <div><button type="button" aria-label="Previous history page" disabled={historyPage <= 1 || isLoadingHistory} onClick={() => void loadHistoryPage(selectedProduct.id, historyPage - 1)}><ChevronLeft size={16} aria-hidden="true" /></button><span className="inventory-history-page-number">Page {historyPage} of {historyTotalPages}</span><button type="button" aria-label="Next history page" disabled={historyPage >= historyTotalPages || isLoadingHistory} onClick={() => void loadHistoryPage(selectedProduct.id, historyPage + 1)}><ChevronRight size={16} aria-hidden="true" /></button></div>
                </div>
              </>
            )}
            <div className="modal-actions">
              <button className="secondary-button" type="button" onClick={() => setIsHistoryModalOpen(false)} disabled={isLoadingHistory}>Back to product details</button>
            </div>
          </section>
        </div>
      )}
    </section>
  )
}
