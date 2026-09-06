import { useCallback, useEffect, useState } from 'react'
import { CircleAlert, Download, RefreshCw } from 'lucide-react'
import { API_BASE, dateToBoundary } from '../../lib/api'
import { usePeriodFilter } from '../../lib/usePeriodFilter'

type ReportsPageProps = {
  token: string
  onUnauthorized: () => void
}

type ReportSummary = {
  total_orders: number
  cancelled_orders: number
  invoices_generated: number
  total_revenue: number
  total_taxable_value: number
  total_gst_collected: number
  total_igst: number
  total_cgst: number
  total_sgst: number
  fulfilled_orders: number
  unfulfilled_orders: number
  paid_orders: number
}

type StateReport = {
  state: string
  orders: number
  taxable_value: number
  total_gst: number
  revenue: number
}

const sourceOptions = [
  { id: 'shopify', label: 'Shopify' },
  { id: 'amazon', label: 'Amazon' },
  { id: 'pos', label: 'POS' },
  { id: 'b2b', label: 'B2B' },
]

const emptySummary: ReportSummary = {
  total_orders: 0,
  cancelled_orders: 0,
  invoices_generated: 0,
  total_revenue: 0,
  total_taxable_value: 0,
  total_gst_collected: 0,
  total_igst: 0,
  total_cgst: 0,
  total_sgst: 0,
  fulfilled_orders: 0,
  unfulfilled_orders: 0,
  paid_orders: 0,
}

const formatCurrency = (value: number | undefined) => `₹${Math.round(value || 0).toLocaleString('en-IN')}`
const formatNumber = (value: number | undefined) => Math.round(value || 0).toLocaleString('en-IN')

function ReportMetric({ label, value, detail }: { label: string; value: string; detail: string }) {
  return (
    <article className="report-metric-card">
      <span className="metric-label">{label}</span>
      <strong>{value}</strong>
      <small>{detail}</small>
    </article>
  )
}

export function ReportsPage({ token, onUnauthorized }: ReportsPageProps) {
  const { startDate, endDate } = usePeriodFilter()
  const [summary, setSummary] = useState<ReportSummary>(emptySummary)
  const [stateData, setStateData] = useState<StateReport[]>([])
  const [selectedSources, setSelectedSources] = useState<string[]>(sourceOptions.map((source) => source.id))
  const [isLoading, setIsLoading] = useState(true)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [error, setError] = useState('')

  const fetchReports = useCallback(async (silent = false) => {
    if (silent) setIsRefreshing(true)
    else setIsLoading(true)
    setError('')

    const query = new URLSearchParams()
    if (startDate) query.set('start_date', dateToBoundary(startDate))
    if (endDate) query.set('end_date', dateToBoundary(endDate, true))
    if (selectedSources.length) query.set('source_ids', selectedSources.join(','))
    const queryString = query.toString()

    const fetchJson = async (path: string) => {
      const response = await fetch(`${API_BASE}${path}`, { headers: { Authorization: `Bearer ${token}` } })
      if (response.status === 401) {
        onUnauthorized()
        throw new Error('Your session has expired. Please sign in again.')
      }
      if (!response.ok) throw new Error(`Reports request failed with status ${response.status}`)
      return response.json() as Promise<Record<string, unknown>>
    }

    try {
      const [summaryResult, stateResult] = await Promise.all([
        fetchJson(`/api/reports/summary?${queryString}`),
        fetchJson(`/api/reports/state-wise?${queryString}`),
      ])
      if (!summaryResult.success) throw new Error('Report summary was not returned')
      setSummary((summaryResult.summary as ReportSummary) || emptySummary)
      if (stateResult.success) setStateData((stateResult.data as StateReport[]) || [])
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to load GST reports')
    } finally {
      setIsLoading(false)
      setIsRefreshing(false)
    }
  }, [endDate, onUnauthorized, selectedSources, startDate, token])

  useEffect(() => {
    void fetchReports()
  }, [fetchReports])

  const toggleSource = (sourceId: string) => {
    setSelectedSources((current) => current.includes(sourceId) ? current.filter((id) => id !== sourceId) : [...current, sourceId])
  }

  return (
    <section className="workspace-page reports-page" aria-labelledby="reports-heading">
      <header className="workspace-page-header">
        <div>
          <p className="eyebrow">Operations / GST</p>
          <h2 id="reports-heading">GST reporting, clearly laid out.</h2>
          <p>See tax collection, invoice activity, and state-wise performance for the shared workspace period.</p>
        </div>
        <button className="secondary-button" type="button" onClick={() => void fetchReports(true)} disabled={isRefreshing}>
          <RefreshCw size={16} className={isRefreshing ? 'spin' : ''} aria-hidden="true" />
          Refresh report
        </button>
      </header>

      <div className="report-source-filter" aria-label="Filter GST report by channel">
        <span className="filter-label">Channels</span>
        {sourceOptions.map((source) => {
          const isSelected = selectedSources.includes(source.id)
          return (
            <button className={`report-source-button ${isSelected ? 'report-source-button-active' : ''}`} type="button" key={source.id} aria-pressed={isSelected} onClick={() => toggleSource(source.id)}>
              <span className="report-source-dot" aria-hidden="true" />
              {source.label}
            </button>
          )
        })}
        <button className="report-source-clear" type="button" onClick={() => setSelectedSources([])}>Clear</button>
      </div>

      {error && (
        <div className="dashboard-error" role="alert">
          <CircleAlert size={18} aria-hidden="true" />
          <span>{error}</span>
          <button type="button" onClick={() => void fetchReports()}>Try again</button>
        </div>
      )}

      <div className="report-metrics-grid">
        <ReportMetric label="Revenue" value={isLoading ? '—' : formatCurrency(summary.total_revenue)} detail={`${formatCurrency(summary.total_taxable_value)} taxable`} />
        <ReportMetric label="GST collected" value={isLoading ? '—' : formatCurrency(summary.total_gst_collected)} detail={`CGST ${formatCurrency(summary.total_cgst)} · SGST ${formatCurrency(summary.total_sgst)}`} />
        <ReportMetric label="Orders" value={isLoading ? '—' : formatNumber(summary.total_orders)} detail={`${formatNumber(summary.fulfilled_orders)} fulfilled`} />
        <ReportMetric label="Invoices generated" value={isLoading ? '—' : formatNumber(summary.invoices_generated)} detail={`${formatNumber(summary.paid_orders)} paid orders`} />
      </div>

      <article className="reports-table-card">
        <div className="reports-table-heading">
          <div>
            <p className="eyebrow">Regional view</p>
            <h3>State-wise GST performance</h3>
          </div>
          <button className="text-button" type="button" disabled title="Export migration follows the legacy GST export flow">
            <Download size={15} aria-hidden="true" /> Export
          </button>
        </div>
        <div className="orders-table-wrap">
          <table className="orders-table reports-table">
            <thead>
              <tr><th>State</th><th>Orders</th><th>Taxable value</th><th>GST collected</th><th>Revenue</th></tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr><td colSpan={5} className="table-state">Loading report data…</td></tr>
              ) : stateData.length === 0 ? (
                <tr><td colSpan={5} className="table-state">No state-wise data for this period.</td></tr>
              ) : stateData.map((state) => (
                <tr key={state.state}>
                  <td><strong>{state.state || 'Unknown'}</strong></td>
                  <td>{formatNumber(state.orders)}</td>
                  <td>{formatCurrency(state.taxable_value)}</td>
                  <td>{formatCurrency(state.total_gst)}</td>
                  <td><strong>{formatCurrency(state.revenue)}</strong></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </article>
    </section>
  )
}
