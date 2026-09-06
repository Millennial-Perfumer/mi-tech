import { useCallback, useEffect, useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import { Check, CircleAlert, RefreshCw } from 'lucide-react'
import { API_BASE, dateToBoundary } from '../../lib/api'
import { usePeriodFilter } from '../../lib/usePeriodFilter'

type DashboardPageProps = {
  token: string
  onUnauthorized: () => void
}

type DashboardMetrics = {
  total_revenue: number
  total_invoices: number
  total_gst_collected: number
  cgst_collected: number
  sgst_collected: number
  igst_collected: number
  total_orders: number
  cancelled_orders: number
  fulfilled_orders: number
  unfulfilled_orders: number
  total_discount: number
  discount_percent: number
  channel_breakdown?: ChannelMetric[]
  payment_breakdown?: PaymentBreakdown
}

type ChannelMetric = {
  source_id: string
  revenue: number
  orders: number
  aov: number
}

type PaymentBreakdown = {
  paid: number
  pending: number
  partial: number
  cancelled: number
}

type TopProduct = {
  sku: string
  title: string
  quantity: number
  revenue: number
}

type RevenuePoint = {
  date: string
  revenue: number
  orders: number
}

type RegionMetric = {
  state: string
  orders: number
  revenue: number
}

const channelOptions = [
  { id: 'shopify', label: 'Shopify' },
  { id: 'amazon', label: 'Amazon' },
  { id: 'pos', label: 'POS' },
  { id: 'b2b', label: 'B2B' },
]

const fallbackMetrics: DashboardMetrics = {
  total_revenue: 0,
  total_invoices: 0,
  total_gst_collected: 0,
  cgst_collected: 0,
  sgst_collected: 0,
  igst_collected: 0,
  total_orders: 0,
  cancelled_orders: 0,
  fulfilled_orders: 0,
  unfulfilled_orders: 0,
  total_discount: 0,
  discount_percent: 0,
  channel_breakdown: [],
  payment_breakdown: { paid: 0, pending: 0, partial: 0, cancelled: 0 },
}

const currency = (value: number) => `₹${Math.round(value || 0).toLocaleString('en-IN')}`
const number = (value: number) => Math.round(value || 0).toLocaleString('en-IN')

function DashboardMetric({ label, value, detail, icon }: { label: string; value: string; detail: string; icon: ReactNode }) {
  return (
    <article className="dashboard-metric-card">
      <div className="dashboard-metric-heading">
        <span className="metric-label">{label}</span>
        <span className="dashboard-metric-icon" aria-hidden="true">{icon}</span>
      </div>
      <strong>{value}</strong>
      <span className="metric-detail">{detail}</span>
    </article>
  )
}

function DashboardEmpty({ label }: { label: string }) {
  return <p className="dashboard-empty">{label}</p>
}

export function DashboardPage({ token, onUnauthorized }: DashboardPageProps) {
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null)
  const [topProducts, setTopProducts] = useState<TopProduct[]>([])
  const [revenueTrend, setRevenueTrend] = useState<RevenuePoint[]>([])
  const [regions, setRegions] = useState<RegionMetric[]>([])
  const [selectedChannels, setSelectedChannels] = useState<string[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [error, setError] = useState('')
  const { startDate, endDate } = usePeriodFilter()

  const fetchDashboardData = useCallback(async (silent = false) => {
    if (!silent) setIsLoading(true)
    setIsRefreshing(true)
    setError('')

    const query = new URLSearchParams()
    if (startDate) query.set('start_date', dateToBoundary(startDate))
    if (endDate) query.set('end_date', dateToBoundary(endDate, true))
    if (selectedChannels.length) query.set('source_ids', selectedChannels.join(','))
    const channelQuery = query.toString()

    const fetchJson = async (path: string) => {
      const response = await fetch(`${API_BASE}${path}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (response.status === 401) {
        onUnauthorized()
        throw new Error('Your session has expired. Please sign in again.')
      }
      if (!response.ok) throw new Error(`Request failed with status ${response.status}`)
      return response.json() as Promise<Record<string, unknown>>
    }

    try {
      const [metricsResult, productsResult, trendResult, regionsResult] = await Promise.allSettled([
        fetchJson(`/api/dashboard/metrics?${channelQuery}`),
        fetchJson(`/api/dashboard/top-products?${channelQuery}&limit=5`),
        fetchJson(`/api/dashboard/revenue-trend?${channelQuery}`),
        fetchJson(`/api/dashboard/geo-distribution?${channelQuery}&limit=5`),
      ])

      if (metricsResult.status === 'rejected') throw metricsResult.reason
      if (metricsResult.value.success) {
        setMetrics((metricsResult.value.metrics as DashboardMetrics) || fallbackMetrics)
      } else {
        throw new Error('Dashboard metrics were not returned')
      }

      if (productsResult.status === 'fulfilled' && productsResult.value.success) {
        setTopProducts((productsResult.value.products as TopProduct[]) || [])
      }
      if (trendResult.status === 'fulfilled' && trendResult.value.success) {
        setRevenueTrend((trendResult.value.trend as RevenuePoint[]) || [])
      }
      if (regionsResult.status === 'fulfilled' && regionsResult.value.success) {
        setRegions((regionsResult.value.distribution as RegionMetric[]) || [])
      }
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to load dashboard data')
    } finally {
      setIsLoading(false)
      setIsRefreshing(false)
    }
  }, [endDate, onUnauthorized, selectedChannels, startDate, token])

  useEffect(() => {
    void fetchDashboardData()
  }, [fetchDashboardData])

  const safeMetrics = metrics || fallbackMetrics
  const payment = safeMetrics.payment_breakdown || fallbackMetrics.payment_breakdown!
  const maxRevenue = useMemo(() => Math.max(...revenueTrend.map((point) => point.revenue), 1), [revenueTrend])
  const maxChannelRevenue = useMemo(() => Math.max(...(safeMetrics.channel_breakdown || []).map((channel) => channel.revenue), 1), [safeMetrics.channel_breakdown])
  const totalPayments = Math.max(safeMetrics.total_orders, 1)

  return (
    <section className="dashboard-page" aria-labelledby="dashboard-heading">
      <header className="dashboard-header">
        <div>
          <p className="eyebrow">Mi Tech / Overview</p>
          <h2 id="dashboard-heading">Overview</h2>
          <p className="dashboard-subtitle">Revenue, orders, GST, and operational health for the selected period.</p>
        </div>
        <div className="dashboard-actions">
          <button className="secondary-button" type="button" onClick={() => void fetchDashboardData(true)} disabled={isRefreshing}>
            <RefreshCw size={16} className={isRefreshing ? 'spin' : ''} aria-hidden="true" />
            Refresh
          </button>
        </div>
      </header>

      <div className="channel-filter" aria-label="Filter dashboard by channel">
        <span className="filter-label">Channels</span>
        {channelOptions.map((channel) => {
          const isSelected = selectedChannels.includes(channel.id)
          return (
            <button
              className={`channel-filter-button ${isSelected ? 'channel-filter-button-active' : ''}`}
              type="button"
              key={channel.id}
              aria-pressed={isSelected}
              onClick={() => setSelectedChannels((current) => isSelected ? current.filter((id) => id !== channel.id) : [...current, channel.id])}
            >
              {isSelected && <Check size={14} aria-hidden="true" />}
              {channel.label}
            </button>
          )
        })}
        {selectedChannels.length > 0 && <button className="clear-filter" type="button" onClick={() => setSelectedChannels([])}>Clear</button>}
      </div>

      {error && (
        <div className="dashboard-error" role="alert">
          <CircleAlert size={18} aria-hidden="true" />
          <span>{error}</span>
          <button type="button" onClick={() => void fetchDashboardData()}>Try again</button>
        </div>
      )}

      <div className="dashboard-metrics-grid" aria-label="Business overview">
        <DashboardMetric label="Revenue" value={isLoading ? '—' : currency(safeMetrics.total_revenue)} detail={`${number(safeMetrics.total_invoices)} invoices`} icon={<span>₹</span>} />
        <DashboardMetric label="GST collected" value={isLoading ? '—' : currency(safeMetrics.total_gst_collected)} detail={`CGST ${currency(safeMetrics.cgst_collected)} · SGST ${currency(safeMetrics.sgst_collected)}`} icon={<span>税</span>} />
        <DashboardMetric label="Orders" value={isLoading ? '—' : number(safeMetrics.total_orders)} detail={`${number(safeMetrics.fulfilled_orders)} fulfilled`} icon={<span>↗</span>} />
        <DashboardMetric label="Unfulfilled" value={isLoading ? '—' : number(safeMetrics.unfulfilled_orders)} detail={`${number(safeMetrics.cancelled_orders)} cancelled`} icon={<span>!</span>} />
        <DashboardMetric label="Average order" value={isLoading ? '—' : currency(safeMetrics.total_orders ? safeMetrics.total_revenue / safeMetrics.total_orders : 0)} detail="Across all channels" icon={<span>⌁</span>} />
        <DashboardMetric label="Discounts" value={isLoading ? '—' : currency(safeMetrics.total_discount)} detail={`${(safeMetrics.discount_percent || 0).toFixed(1)}% of gross revenue`} icon={<span>−</span>} />
      </div>

      <div className="dashboard-analytics-grid">
        <article className="dashboard-card dashboard-trend-card">
          <div className="dashboard-card-heading">
            <div>
              <p className="eyebrow">Performance</p>
              <h3>Revenue trend</h3>
            </div>
            <strong>{currency(revenueTrend.reduce((total, point) => total + point.revenue, 0))}</strong>
          </div>
          {revenueTrend.length > 1 ? (
            <div className="revenue-chart" aria-label="Revenue trend chart">
              {revenueTrend.map((point) => (
                <div className="revenue-bar-wrap" key={point.date}>
                  <div className="revenue-bar" style={{ height: `${Math.max((point.revenue / maxRevenue) * 100, 3)}%` }} title={`${point.date}: ${currency(point.revenue)}`} />
                  <span>{point.date.slice(5)}</span>
                </div>
              ))}
            </div>
          ) : <DashboardEmpty label={isLoading ? 'Loading trend…' : 'Not enough data for a trend yet.'} />}
        </article>

        <article className="dashboard-card">
          <div className="dashboard-card-heading">
            <div><p className="eyebrow">Mix</p><h3>Revenue by channel</h3></div>
          </div>
          <div className="dashboard-list">
            {(safeMetrics.channel_breakdown || []).map((channel) => (
              <div className="channel-row" key={channel.source_id}>
                <div className="channel-row-top"><span>{channel.source_id}</span><strong>{currency(channel.revenue)}</strong></div>
                <div className="progress-track"><span style={{ width: `${(channel.revenue / maxChannelRevenue) * 100}%` }} /></div>
                <small>{number(channel.orders)} orders · AOV {currency(channel.aov)}</small>
              </div>
            ))}
            {!safeMetrics.channel_breakdown?.length && <DashboardEmpty label="No channel data for this range." />}
          </div>
        </article>

        <article className="dashboard-card">
          <div className="dashboard-card-heading"><div><p className="eyebrow">Products</p><h3>Top performers</h3></div></div>
          <div className="dashboard-list">
            {topProducts.map((product, index) => (
              <div className="product-row" key={product.sku}>
                <span className="rank-badge">{index + 1}</span>
                <div><strong>{product.title}</strong><small>{product.sku} · {number(product.quantity)} sold</small></div>
                <span>{currency(product.revenue)}</span>
              </div>
            ))}
            {!topProducts.length && <DashboardEmpty label="No product data for this range." />}
          </div>
        </article>

        <article className="dashboard-card">
          <div className="dashboard-card-heading"><div><p className="eyebrow">Geography</p><h3>Top regions</h3></div></div>
          <div className="dashboard-list">
            {regions.map((region) => (
              <div className="region-row" key={region.state}><span>{region.state}</span><div><strong>{number(region.orders)} orders</strong><small>{currency(region.revenue)}</small></div></div>
            ))}
            {!regions.length && <DashboardEmpty label="No regional data for this range." />}
          </div>
        </article>

        <article className="dashboard-card dashboard-payment-card">
          <div className="dashboard-card-heading"><div><p className="eyebrow">Collections</p><h3>Payment health</h3></div></div>
          <div className="payment-track" aria-label="Payment collection health">
            <span className="payment-paid" style={{ width: `${(payment.paid / totalPayments) * 100}%` }} />
            <span className="payment-pending" style={{ width: `${(payment.pending / totalPayments) * 100}%` }} />
            <span className="payment-partial" style={{ width: `${(payment.partial / totalPayments) * 100}%` }} />
          </div>
          <div className="payment-legend"><span><i className="payment-paid" /> Paid {number(payment.paid)}</span><span><i className="payment-pending" /> Pending {number(payment.pending)}</span><span><i className="payment-partial" /> Partial {number(payment.partial)}</span></div>
          <div className="discount-callout"><span>Discount leakage</span><strong>{currency(safeMetrics.total_discount)}</strong><small>{(safeMetrics.discount_percent || 0).toFixed(1)}% of gross revenue</small></div>
        </article>
      </div>
    </section>
  )
}
