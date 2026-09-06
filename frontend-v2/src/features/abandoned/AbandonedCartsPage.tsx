import { useCallback, useEffect, useMemo, useState } from 'react'
import { Check, CircleAlert, Clock3, ExternalLink, RefreshCw, Search, ShoppingBag, Trash2, X } from 'lucide-react'
import { API_BASE } from '../../lib/api'
import { usePeriodFilter } from '../../lib/usePeriodFilter'

type AbandonedCartsPageProps = { token: string; onUnauthorized: () => void }

type Checkout = {
  id: number
  checkout_id: string
  email: string
  phone: string
  customer_name: string
  checkout_url: string
  line_items: unknown
  total_price: number
  currency: string
  completed: boolean
  order_id?: string
  recovery_status: string
  recovery_attempts: number
  marketing_consent: boolean
  abandoned_at: string
  last_error?: string
}

type Analytics = {
  totalAbandonedRevenue: number
  recoveredRevenue: number
  pendingRevenue: number
  abandonedCartCount: number
  recoveredCartCount: number
  recoveryRate: number
  whatsappStats?: { sent: number; delivered: number; read: number; clicked: number; failed: number }
  revenueTimeline?: { date: string; abandonedAmount: number; recoveredAmount: number }[]
  topLostCarts?: { customer_name: string; total_price: number; currency: string; recovery_status: string; attempts: number }[]
}

const statusOptions = ['PENDING', 'PROCESSING', 'SENT', 'FAILED', 'CANCELLED', 'RECOVERED']

function money(value: number, currency = 'INR') {
  return new Intl.NumberFormat('en-IN', { style: 'currency', currency: currency === 'INR' ? 'INR' : currency }).format(value || 0)
}

function dateTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return date.toLocaleString('en-IN', { day: 'numeric', month: 'short', year: 'numeric', hour: 'numeric', minute: '2-digit' })
}

function parseLineItems(raw: unknown) {
  if (!raw) return [] as { title?: string; quantity?: number; price?: number | string }[]
  try { return (typeof raw === 'string' ? JSON.parse(raw) : raw) as { title?: string; quantity?: number; price?: number | string }[] } catch { return [] }
}

export function AbandonedCartsPage({ token, onUnauthorized }: AbandonedCartsPageProps) {
  const { startDate, endDate } = usePeriodFilter()
  const [activeTab, setActiveTab] = useState<'list' | 'analytics'>('list')
  const [checkouts, setCheckouts] = useState<Checkout[]>([])
  const [analytics, setAnalytics] = useState<Analytics | null>(null)
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState('')
  const [selected, setSelected] = useState<Checkout | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isWorking, setIsWorking] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')

  const request = useCallback(async (path: string, options: RequestInit = {}) => {
    const response = await fetch(`${API_BASE}${path}`, { ...options, headers: { Authorization: `Bearer ${token}`, ...(options.headers || {}) } })
    if (response.status === 401) { onUnauthorized(); throw new Error('Your session has expired. Please sign in again.') }
    if (!response.ok) throw new Error(`Abandoned carts request failed with status ${response.status}`)
    return response
  }, [onUnauthorized, token])

  const loadList = useCallback(async () => {
    setIsLoading(true)
    const query = new URLSearchParams({ page: String(page), limit: '15', search, start_date: startDate, end_date: endDate })
    if (status) query.set('status', status)
    try {
      const response = await request(`/api/abandoned-checkouts?${query.toString()}`)
      const data = await response.json() as { checkouts?: Checkout[]; total_count?: number }
      setCheckouts(data.checkouts || [])
      setTotal(data.total_count || 0)
    } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to load abandoned carts') } finally { setIsLoading(false) }
  }, [endDate, page, request, search, startDate, status])

  const loadAnalytics = useCallback(async () => {
    try {
      const query = new URLSearchParams({ start_date: startDate, end_date: endDate })
      const response = await request(`/api/abandoned-checkouts/analytics?${query.toString()}`)
      const data = await response.json() as { analytics?: Analytics }
      setAnalytics(data.analytics || null)
    } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to load recovery analytics') }
  }, [endDate, request, startDate])

  useEffect(() => { if (activeTab === 'list') void loadList(); else void loadAnalytics() }, [activeTab, loadAnalytics, loadList])
  useEffect(() => { setPage(1) }, [endDate, search, startDate, status])

  const totalPages = Math.max(Math.ceil(total / 15), 1)
  const visibleStatuses = useMemo(() => checkouts.filter((checkout) => checkout.completed ? status === 'RECOVERED' || !status : true), [checkouts, status])

  const recover = async (checkout: Checkout) => {
    setIsWorking(true)
    try { await request(`/api/abandoned-checkouts/recover?id=${checkout.id}`, { method: 'POST' }); setNotice('Recovery message dispatched'); await loadList() } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to dispatch recovery') } finally { setIsWorking(false) }
  }

  const updateStatus = async (checkout: Checkout, nextStatus: string) => {
    setIsWorking(true)
    try { await request('/api/abandoned-checkouts/status', { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ id: checkout.id, recovery_status: nextStatus === 'RECOVERED' ? 'SENT' : nextStatus, completed: nextStatus === 'RECOVERED' }) }); setNotice('Checkout status updated'); await loadList() } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to update checkout status') } finally { setIsWorking(false) }
  }

  const deleteCheckout = async (checkout: Checkout) => {
    if (!window.confirm(`Delete checkout ${checkout.checkout_id}?`)) return
    setIsWorking(true)
    try { await request(`/api/abandoned-checkouts?id=${checkout.id}`, { method: 'DELETE' }); setSelected(null); setNotice('Checkout record deleted'); await loadList() } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to delete checkout') } finally { setIsWorking(false) }
  }

  return (
    <section className="workspace-page abandoned-page" aria-labelledby="abandoned-heading">
      <header className="workspace-page-header"><div><p className="eyebrow">Engagement / Recovery</p><h2 id="abandoned-heading">Abandoned carts</h2><p>Review abandoned checkouts and send recovery messages.</p></div><button className="secondary-button" type="button" onClick={() => activeTab === 'list' ? void loadList() : void loadAnalytics()} disabled={isLoading}><RefreshCw size={15} className={isLoading ? 'spin' : undefined} aria-hidden="true" /> Refresh</button></header>
      {error && <div className="dashboard-error" role="alert"><CircleAlert size={18} aria-hidden="true" /><span>{error}</span><button type="button" onClick={() => { setError(''); void loadList() }}>Try again</button></div>}
      {notice && <div className="inventory-notice" role="status">{notice}</div>}
      <div className="abandoned-tabs" role="tablist" aria-label="Abandoned checkout views"><button className={activeTab === 'list' ? 'abandoned-tab-active' : ''} type="button" role="tab" aria-selected={activeTab === 'list'} onClick={() => setActiveTab('list')}><ShoppingBag size={15} aria-hidden="true" /> Checkout list</button><button className={activeTab === 'analytics' ? 'abandoned-tab-active' : ''} type="button" role="tab" aria-selected={activeTab === 'analytics'} onClick={() => setActiveTab('analytics')}><span aria-hidden="true">↗</span> Recovery analytics</button></div>
      {activeTab === 'list' ? <><div className="abandoned-toolbar"><label className="orders-search"><Search size={16} aria-hidden="true" /><span className="sr-only">Search abandoned checkouts</span><input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search customer, phone, or checkout" /></label><label className="compact-select"><span className="sr-only">Recovery status</span><select value={status} onChange={(event) => setStatus(event.target.value)}><option value="">All statuses</option>{statusOptions.map((option) => <option key={option} value={option}>{option.replace('_', ' ')}</option>)}</select></label><span className="filter-count">{total.toLocaleString('en-IN')} checkouts</span></div><div className="abandoned-list">{isLoading ? <div className="empty-panel"><ShoppingBag size={20} aria-hidden="true" /><p>Loading abandoned checkouts…</p></div> : visibleStatuses.length === 0 ? <div className="empty-panel"><ShoppingBag size={20} aria-hidden="true" /><div><h2>No checkouts found</h2><p>Try another status or search term.</p></div></div> : visibleStatuses.map((checkout) => <article className="abandoned-card" key={checkout.id}><div className="abandoned-card-main"><div className="abandoned-card-heading"><div><span className="checkout-id">{checkout.checkout_id}</span><h3>{checkout.customer_name || 'Unknown customer'}</h3></div><strong>{money(checkout.total_price, checkout.currency)}</strong></div><p>{checkout.email || checkout.phone || 'No contact details'} · Abandoned {dateTime(checkout.abandoned_at)}</p><div className="abandoned-card-foot"><span className={`status-pill status-pill-${checkout.completed ? 'success' : checkout.recovery_status === 'FAILED' ? 'danger' : 'neutral'}`}>{checkout.completed ? 'Recovered' : checkout.recovery_status}</span><span><Clock3 size={13} aria-hidden="true" /> {checkout.recovery_attempts} recovery attempts</span>{checkout.marketing_consent && <span>Marketing consent</span>}</div></div><div className="abandoned-card-actions"><label className="compact-select"><span className="sr-only">Status for {checkout.checkout_id}</span><select value={checkout.completed ? 'RECOVERED' : checkout.recovery_status} disabled={isWorking} onChange={(event) => void updateStatus(checkout, event.target.value)}>{statusOptions.map((option) => <option key={option} value={option}>{option.replace('_', ' ')}</option>)}</select></label><button className="secondary-button" type="button" onClick={() => setSelected(checkout)}>Details</button>{!checkout.completed && <button className="primary-button" type="button" onClick={() => void recover(checkout)} disabled={isWorking || !checkout.phone}><SendIcon /> Recover</button>}</div></article>)}</div><div className="orders-pagination"><span>Showing {visibleStatuses.length ? (page - 1) * 15 + 1 : 0}–{Math.min(page * 15, total)} of {total}</span><div><button type="button" aria-label="Previous checkout page" disabled={page <= 1 || isLoading} onClick={() => setPage((current) => current - 1)}>‹</button><button type="button" aria-label="Next checkout page" disabled={page >= totalPages || isLoading} onClick={() => setPage((current) => current + 1)}>›</button></div></div></> : <AnalyticsPanel analytics={analytics} />}
      {selected && <div className="modal-scrim" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) setSelected(null) }}><div className="modal-card abandoned-modal" role="dialog" aria-modal="true" aria-labelledby="checkout-detail-heading"><div className="modal-heading"><div><p className="eyebrow">Checkout details</p><h2 id="checkout-detail-heading">{selected.customer_name || 'Unknown customer'}</h2></div><button className="icon-button" type="button" aria-label="Close checkout details" onClick={() => setSelected(null)}><X size={19} aria-hidden="true" /></button></div><div className="checkout-detail-grid"><span>Checkout ID<strong>{selected.checkout_id}</strong></span><span>Contact<strong>{selected.email || selected.phone || '—'}</strong></span><span>Abandoned<strong>{dateTime(selected.abandoned_at)}</strong></span><span>Recovery attempts<strong>{selected.recovery_attempts}</strong></span></div>{selected.checkout_url && <a className="secondary-button" href={selected.checkout_url} target="_blank" rel="noreferrer"><ExternalLink size={14} aria-hidden="true" /> Open checkout link</a>}<h3 className="detail-section-heading">Line items</h3><div className="checkout-line-items">{parseLineItems(selected.line_items).length ? parseLineItems(selected.line_items).map((item, index) => <div key={`${item.title}-${index}`}><span>{item.title || 'Product'} <small>× {item.quantity || 1}</small></span><strong>{money(Number(item.price || 0) * Number(item.quantity || 1), selected.currency)}</strong></div>) : <p>No line item details available.</p>}</div>{selected.last_error && <div className="dashboard-error"><CircleAlert size={16} aria-hidden="true" />{selected.last_error}</div>}<div className="modal-actions"><button className="secondary-button" type="button" onClick={() => void deleteCheckout(selected)} disabled={isWorking}><Trash2 size={14} aria-hidden="true" /> Delete record</button><button className="secondary-button" type="button" onClick={() => setSelected(null)}>Close</button></div></div></div>}
    </section>
  )
}

function SendIcon() { return <span aria-hidden="true">↗</span> }

function AnalyticsPanel({ analytics }: { analytics: Analytics | null }) {
  if (!analytics) return <div className="empty-panel"><ShoppingBag size={20} aria-hidden="true" /><p>Loading recovery analytics…</p></div>
  const stats = [{ label: 'Abandoned value', value: money(analytics.totalAbandonedRevenue), detail: `${analytics.abandonedCartCount} checkouts` }, { label: 'Recovered value', value: money(analytics.recoveredRevenue), detail: `${analytics.recoveredCartCount} recovered` }, { label: 'Pending value', value: money(analytics.pendingRevenue), detail: 'Still at risk' }, { label: 'Recovery rate', value: `${analytics.recoveryRate.toFixed(1)}%`, detail: 'Checkout to order' }]
  return <div className="recovery-analytics"><div className="report-metrics-grid">{stats.map((stat) => <article className="report-metric-card" key={stat.label}><span className="metric-label">{stat.label}</span><strong>{stat.value}</strong><small>{stat.detail}</small></article>)}</div><div className="analytics-grid"><section className="reports-table-card"><div className="reports-table-heading"><div><p className="eyebrow">Recovery timeline</p><h3>Value by day</h3></div></div><div className="orders-table-wrap"><table className="orders-table"><thead><tr><th>Date</th><th>Abandoned</th><th>Recovered</th></tr></thead><tbody>{analytics.revenueTimeline?.length ? analytics.revenueTimeline.map((row) => <tr key={row.date}><td>{row.date}</td><td className="table-money">{money(row.abandonedAmount)}</td><td className="table-money">{money(row.recoveredAmount)}</td></tr>) : <tr><td className="table-state" colSpan={3}>No timeline data for this period.</td></tr>}</tbody></table></div></section><section className="reports-table-card"><div className="reports-table-heading"><div><p className="eyebrow">Largest opportunities</p><h3>Top lost checkouts</h3></div></div><div className="lost-cart-list">{analytics.topLostCarts?.length ? analytics.topLostCarts.map((cart, index) => <div className="lost-cart-row" key={`${cart.customer_name}-${index}`}><span><strong>{cart.customer_name || 'Unknown customer'}</strong><small>{cart.recovery_status} · {cart.attempts} attempts</small></span><strong>{money(cart.total_price, cart.currency)}</strong></div>) : <p className="conversation-empty">No lost checkout data for this period.</p>}</div></section></div><p className="communication-footnote"><Check size={14} aria-hidden="true" /> WhatsApp funnel metrics are supplied by the existing recovery service.</p></div>
}
