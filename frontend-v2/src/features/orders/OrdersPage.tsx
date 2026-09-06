import { useCallback, useEffect, useState } from 'react'
import { ChevronLeft, ChevronRight, CircleAlert, Plus, Search, SlidersHorizontal } from 'lucide-react'
import { API_BASE, dateToBoundary } from '../../lib/api'
import { usePeriodFilter } from '../../lib/usePeriodFilter'
import { ConvertOrderToB2BModal, OrderCreateModal, OrderDetailsModal } from './OrderModals'

type OrdersPageProps = {
  token: string
  onUnauthorized: () => void
}

type Order = {
  id: string | number
  order_number?: string
  customer_name?: string
  created_at?: string
  total_price?: string | number
  financial_status?: string
  fulfillment_status?: string
  status?: string
  source_id?: string
}

const pageSize = 20

function formatMoney(value: string | number | undefined) {
  return `₹${Number(value || 0).toLocaleString('en-IN', { maximumFractionDigits: 0 })}`
}

function formatDate(value: string | undefined) {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return date.toLocaleDateString('en-IN', { day: 'numeric', month: 'short', year: 'numeric' })
}

function formatStatus(value: string | undefined, fallback: string) {
  const status = value || fallback
  return status.replace(/_/g, ' ').toLowerCase().replace(/\b\w/g, (letter) => letter.toUpperCase())
}

function statusTone(value: string | undefined) {
  const normalized = (value || '').toLowerCase()
  if (normalized.includes('paid') || normalized.includes('fulfilled') || normalized === 'completed') return 'success'
  if (normalized.includes('cancel') || normalized.includes('failed')) return 'danger'
  return 'neutral'
}

export function OrdersPage({ token, onUnauthorized }: OrdersPageProps) {
  const { startDate, endDate } = usePeriodFilter()
  const [orders, setOrders] = useState<Order[]>([])
  const [search, setSearch] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const [sourceFilter, setSourceFilter] = useState('')
  const [paymentFilter, setPaymentFilter] = useState('')
  const [fulfillmentFilter, setFulfillmentFilter] = useState('')
  const [page, setPage] = useState(1)
  const [totalCount, setTotalCount] = useState(0)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [isCreateOpen, setIsCreateOpen] = useState(false)
  const [selectedOrderId, setSelectedOrderId] = useState<string | number | null>(null)
  const [convertOrderId, setConvertOrderId] = useState<string | number | null>(null)

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedSearch(search.trim()), 250)
    return () => window.clearTimeout(timer)
  }, [search])

  useEffect(() => {
    setPage(1)
  }, [debouncedSearch, endDate, fulfillmentFilter, paymentFilter, sourceFilter, startDate])

  const fetchOrders = useCallback(async () => {
    setIsLoading(true)
    setError('')

    const query = new URLSearchParams({
      page: String(page),
      limit: String(pageSize),
      sort_by: 'created_at',
      sort_order: 'DESC',
    })
    if (startDate) query.set('start_date', dateToBoundary(startDate))
    if (endDate) query.set('end_date', dateToBoundary(endDate, true))
    if (debouncedSearch) query.set('search', debouncedSearch)
    if (sourceFilter) query.set('source', sourceFilter)
    if (paymentFilter) query.set('financial_status', paymentFilter)
    if (fulfillmentFilter) query.set('fulfillment_status', fulfillmentFilter)

    try {
      const response = await fetch(`${API_BASE}/api/orders?${query.toString()}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (response.status === 401) {
        onUnauthorized()
        throw new Error('Your session has expired. Please sign in again.')
      }
      if (!response.ok) throw new Error(`Orders request failed with status ${response.status}`)
      const data = await response.json() as { success?: boolean; orders?: Order[]; total_count?: number; message?: string }
      if (!data.success) throw new Error(data.message || 'Orders were not returned')
      setOrders(data.orders || [])
      setTotalCount(data.total_count || 0)
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to load orders')
      setOrders([])
      setTotalCount(0)
    } finally {
      setIsLoading(false)
    }
  }, [debouncedSearch, endDate, fulfillmentFilter, onUnauthorized, page, paymentFilter, sourceFilter, startDate, token])

  useEffect(() => {
    void fetchOrders()
  }, [fetchOrders])

  const totalPages = Math.max(Math.ceil(totalCount / pageSize), 1)

  return (
    <section className="workspace-page orders-page" aria-labelledby="orders-heading">
      <header className="workspace-page-header">
        <div>
          <p className="eyebrow">Operations / Orders</p>
          <h2 id="orders-heading">Orders, without the noise.</h2>
          <p>Search every connected channel, inspect order health, and keep fulfilment moving.</p>
        </div>
          <div className="support-header-actions"><span className="page-period-note">Period filter is shared across the workspace</span><button className="primary-button" type="button" onClick={() => setIsCreateOpen(true)}><Plus size={15} aria-hidden="true" /> New order</button></div>
      </header>

      <div className="orders-toolbar" aria-label="Order filters">
        <label className="orders-search">
          <Search size={16} aria-hidden="true" />
          <span className="sr-only">Search orders</span>
          <input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search order or customer" />
        </label>
        <label className="compact-select">
          <span className="sr-only">Channel</span>
          <select value={sourceFilter} onChange={(event) => setSourceFilter(event.target.value)}>
            <option value="">All channels</option>
            <option value="shopify">Shopify</option>
            <option value="amazon">Amazon</option>
            <option value="pos">POS</option>
            <option value="b2b">B2B</option>
          </select>
        </label>
        <label className="compact-select">
          <span className="sr-only">Payment status</span>
          <select value={paymentFilter} onChange={(event) => setPaymentFilter(event.target.value)}>
            <option value="">Any payment</option>
            <option value="paid">Paid</option>
            <option value="pending">Pending</option>
            <option value="partially_paid">Partially paid</option>
          </select>
        </label>
        <label className="compact-select">
          <span className="sr-only">Fulfilment status</span>
          <select value={fulfillmentFilter} onChange={(event) => setFulfillmentFilter(event.target.value)}>
            <option value="">Any fulfilment</option>
            <option value="fulfilled">Fulfilled</option>
            <option value="unfulfilled">Unfulfilled</option>
          </select>
        </label>
        <span className="filter-count"><SlidersHorizontal size={15} aria-hidden="true" /> {totalCount.toLocaleString('en-IN')} orders</span>
      </div>

      {error && (
        <div className="dashboard-error" role="alert">
          <CircleAlert size={18} aria-hidden="true" />
          <span>{error}</span>
          <button type="button" onClick={() => void fetchOrders()}>Try again</button>
        </div>
      )}
      {notice && <div className="inventory-notice" role="status">{notice}</div>}

      <div className="orders-card">
        <div className="orders-card-heading">
          <div>
            <p className="eyebrow">Order register</p>
            <h3>{isLoading ? 'Loading orders…' : `${totalCount.toLocaleString('en-IN')} orders found`}</h3>
          </div>
          <span className="orders-card-meta">Newest first</span>
        </div>

        <div className="orders-table-wrap">
          <table className="orders-table">
            <thead>
              <tr>
                <th>Order</th>
                <th>Customer</th>
                <th>Date</th>
                <th>Channel</th>
                <th>Amount</th>
                <th>Payment</th>
                <th>Fulfilment</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr><td colSpan={8} className="table-state">Loading orders…</td></tr>
              ) : orders.length === 0 ? (
                <tr><td colSpan={8} className="table-state">No orders match this period and filter set.</td></tr>
              ) : orders.map((order) => (
                <tr key={order.id}>
                  <td><strong>{order.order_number || `#${order.id}`}</strong></td>
                  <td>{order.customer_name || 'Guest customer'}</td>
                  <td className="table-muted">{formatDate(order.created_at)}</td>
                  <td><span className="channel-label">{order.source_id || '—'}</span></td>
                  <td><strong>{formatMoney(order.total_price)}</strong></td>
                  <td><span className={`status-pill status-pill-${statusTone(order.financial_status)}`}>{formatStatus(order.financial_status, 'Pending')}</span></td>
                  <td><span className={`status-pill status-pill-${statusTone(order.fulfillment_status || order.status)}`}>{formatStatus(order.fulfillment_status, 'Unfulfilled')}</span></td>
                  <td><button className="table-link-button" type="button" onClick={() => setSelectedOrderId(order.id)}>View</button></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <footer className="orders-pagination">
          <span>{totalCount ? `${((page - 1) * pageSize) + 1}–${Math.min(page * pageSize, totalCount)} of ${totalCount.toLocaleString('en-IN')}` : '0 orders'}</span>
          <div>
            <button type="button" aria-label="Previous page" disabled={page <= 1 || isLoading} onClick={() => setPage((current) => Math.max(current - 1, 1))}><ChevronLeft size={16} aria-hidden="true" /></button>
            <span>Page {page} of {totalPages}</span>
            <button type="button" aria-label="Next page" disabled={page >= totalPages || isLoading} onClick={() => setPage((current) => Math.min(current + 1, totalPages))}><ChevronRight size={16} aria-hidden="true" /></button>
          </div>
        </footer>
      </div>
      {isCreateOpen && <OrderCreateModal token={token} onUnauthorized={onUnauthorized} onClose={() => setIsCreateOpen(false)} onSuccess={() => void fetchOrders()} />}
      {selectedOrderId !== null && <OrderDetailsModal token={token} onUnauthorized={onUnauthorized} orderId={selectedOrderId} onClose={() => setSelectedOrderId(null)} onChanged={() => void fetchOrders()} onConvertToB2B={() => { setSelectedOrderId(null); setConvertOrderId(selectedOrderId) }} />}
      {convertOrderId !== null && <ConvertOrderToB2BModal token={token} onUnauthorized={onUnauthorized} orderId={convertOrderId} onClose={() => setConvertOrderId(null)} onSuccess={() => { setNotice('B2B invoice created and issued.'); void fetchOrders() }} />}
    </section>
  )
}
