import { useCallback, useEffect, useState } from 'react'
import { ChevronLeft, ChevronRight, CircleAlert, X } from 'lucide-react'
import { dateToBoundary } from '../../lib/api'
import { apiJson, formatDate, formatMoney, textValue } from '../../lib/http'
import { OrderDetailsModal } from './OrderModals'

type StateOrdersModalProps = {
  state: string
  startDate: string
  endDate: string
  selectedChannels: string[]
  token: string
  onUnauthorized: () => void
  onClose: () => void
}

type OrderSummary = {
  id: string | number
  order_number?: string
  customer_name?: string
  customer_phone?: string
  created_at?: string
  total_price?: string | number
  financial_status?: string
  fulfillment_status?: string
  status?: string
  source_id?: string
}

type OrdersResponse = {
  success?: boolean
  orders?: OrderSummary[]
  total_count?: number
  message?: string
}

const pageSize = 10

function orderStatus(order: OrderSummary) {
  const status = textValue(order.status, '').toLowerCase()
  const fulfillment = textValue(order.fulfillment_status, '').toLowerCase()
  if (status === 'cancelled' || status === 'canceled' || fulfillment === 'cancelled' || fulfillment === 'canceled') {
    return { label: 'Cancelled', tone: 'danger' }
  }
  if (fulfillment === 'fulfilled') return { label: 'Fulfilled', tone: 'success' }
  return { label: textValue(order.fulfillment_status || order.status, 'Unfulfilled'), tone: 'neutral' }
}

export function StateOrdersModal({ state, startDate, endDate, selectedChannels, token, onUnauthorized, onClose }: StateOrdersModalProps) {
  const [orders, setOrders] = useState<OrderSummary[]>([])
  const [page, setPage] = useState(1)
  const [totalCount, setTotalCount] = useState(0)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState('')
  const [selectedOrderId, setSelectedOrderId] = useState<string | number | null>(null)

  const fetchOrders = useCallback(async () => {
    setIsLoading(true)
    setError('')

    const query = new URLSearchParams({
      page: String(page),
      limit: String(pageSize),
      sort_by: 'created_at',
      sort_order: 'DESC',
      state,
      exclude_cancelled: 'true',
    })
    if (startDate) query.set('start_date', dateToBoundary(startDate))
    if (endDate) query.set('end_date', dateToBoundary(endDate, true))
    if (selectedChannels.length) query.set('source', selectedChannels.join(','))

    try {
      const data = await apiJson<OrdersResponse>(token, onUnauthorized, `/api/orders?${query.toString()}`)
      if (!data.success) throw new Error(data.message || 'Orders were not returned')
      setOrders(data.orders || [])
      setTotalCount(data.total_count || 0)
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to load state orders')
      setOrders([])
      setTotalCount(0)
    } finally {
      setIsLoading(false)
    }
  }, [endDate, onUnauthorized, page, selectedChannels, startDate, state, token])

  useEffect(() => {
    setPage(1)
  }, [endDate, selectedChannels, startDate, state])

  useEffect(() => {
    void fetchOrders()
  }, [fetchOrders])

  const totalPages = Math.max(Math.ceil(totalCount / pageSize), 1)

  if (selectedOrderId !== null) {
    return (
      <OrderDetailsModal
        token={token}
        onUnauthorized={onUnauthorized}
        orderId={selectedOrderId}
        onClose={() => setSelectedOrderId(null)}
        onChanged={() => void fetchOrders()}
      />
    )
  }

  return (
    <div className="modal-scrim" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose() }}>
      <section className="modal-card state-orders-modal" role="dialog" aria-modal="true" aria-labelledby="state-orders-heading">
        <div className="modal-heading">
          <div>
            <p className="eyebrow">State drill-down</p>
            <h2 id="state-orders-heading">{state} orders</h2>
            <p className="modal-subtitle">Orders matching this state, the selected period, and active channels.</p>
          </div>
          <button className="icon-button" type="button" aria-label="Close state orders" onClick={onClose}>
            <X size={19} aria-hidden="true" />
          </button>
        </div>

        {error && (
          <div className="dashboard-error" role="alert">
            <CircleAlert size={16} aria-hidden="true" />
            <span>{error}</span>
            <button type="button" onClick={() => void fetchOrders()}>Try again</button>
          </div>
        )}

        <div className="orders-table-wrap state-orders-table-wrap">
          <table className="orders-table">
            <caption className="sr-only">Orders from {state}</caption>
            <thead>
              <tr>
                <th>Order</th>
                <th>Customer</th>
                <th>Date</th>
                <th>Channel</th>
                <th>Amount</th>
                <th>Status</th>
                <th>Action</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr><td colSpan={7} className="table-state">Loading orders…</td></tr>
              ) : orders.length === 0 ? (
                <tr><td colSpan={7} className="table-state">No orders found for {state}.</td></tr>
              ) : orders.map((order) => {
                const status = orderStatus(order)
                return (
                  <tr key={String(order.id)}>
                    <td className="mono-text"><strong>{textValue(order.order_number, `#${order.id}`)}</strong></td>
                    <td>{textValue(order.customer_name, 'Guest customer')}</td>
                    <td className="table-muted">{formatDate(order.created_at)}</td>
                    <td><span className="channel-label">{textValue(order.source_id)}</span></td>
                    <td className="table-money"><strong>{formatMoney(order.total_price)}</strong></td>
                    <td><span className={`status-pill status-pill-${status.tone}`}>{status.label}</span></td>
                    <td><button className="table-link-button" type="button" onClick={() => setSelectedOrderId(order.id)}>Details</button></td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>

        <footer className="orders-pagination">
          <span>{totalCount ? `${((page - 1) * pageSize) + 1}–${Math.min(page * pageSize, totalCount)} of ${totalCount.toLocaleString('en-IN')}` : '0 orders'}</span>
          <div>
            <button type="button" aria-label="Previous page" disabled={page <= 1 || isLoading} onClick={() => setPage((current) => Math.max(current - 1, 1))}>
              <ChevronLeft size={16} aria-hidden="true" />
            </button>
            <span>Page {page} of {totalPages}</span>
            <button type="button" aria-label="Next page" disabled={page >= totalPages || isLoading} onClick={() => setPage((current) => Math.min(current + 1, totalPages))}>
              <ChevronRight size={16} aria-hidden="true" />
            </button>
          </div>
        </footer>
      </section>
    </div>
  )
}
