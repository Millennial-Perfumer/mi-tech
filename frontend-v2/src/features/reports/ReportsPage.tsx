import { useCallback, useEffect, useMemo, useState } from 'react'
import { Check, CircleAlert, Download, FileJson, RefreshCw, Table2 } from 'lucide-react'
import { API_BASE, dateToBoundary } from '../../lib/api'
import { usePeriodFilter } from '../../lib/usePeriodFilter'

type ReportsPageProps = {
  token: string
  onUnauthorized: () => void
}

type ReportTab = 'dashboard' | 'state' | 'hsn' | 'documents' | 'gstr1'
type TableValue = string | number

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

type StateReport = Record<string, TableValue> & {
  state: string
  orders: number
  taxable_value: number
  igst: number
  cgst: number
  sgst: number
  total_gst: number
  revenue: number
}

type HSNReport = Record<string, TableValue> & {
  hsn_code: string
  product_count: number
  qty_sold: number
  taxable_value: number
  igst: number
  cgst: number
  sgst: number
  total_gst: number
  revenue: number
}

type DocumentReport = Record<string, TableValue> & {
  document_type: string
  from_serial: string
  to_serial: string
  total_issued: number
  cancelled: number
  net_issued: number
}

type ReportResponse = {
  success?: boolean
  summary?: unknown
  data?: unknown
  message?: string
}

const sourceOptions = [
  { id: 'shopify', label: 'Shopify' },
  { id: 'amazon', label: 'Amazon' },
  { id: 'pos', label: 'POS' },
  { id: 'b2b', label: 'B2B' },
]

const reportTabs: Array<{ id: ReportTab; label: string; description: string }> = [
  { id: 'dashboard', label: 'Dashboard', description: 'Tax and order overview' },
  { id: 'state', label: 'B2C State-wise', description: 'Place-of-supply breakdown' },
  { id: 'hsn', label: 'HSN Summary', description: 'Product tax classification' },
  { id: 'documents', label: 'Documents Issued', description: 'Invoice and note sequences' },
  { id: 'gstr1', label: 'GSTR-1 Export', description: 'Offline utility JSON' },
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

const allSources = sourceOptions.map((source) => source.id)

function formatCurrency(value: number | undefined) {
  return `₹${Number(value || 0).toLocaleString('en-IN', { maximumFractionDigits: 2 })}`
}

function formatNumber(value: number | undefined) {
  return Number(value || 0).toLocaleString('en-IN', { maximumFractionDigits: 2 })
}

function csvValue(value: TableValue | null | undefined) {
  return `"${String(value ?? '').replaceAll('"', '""')}"`
}

function downloadCsv(filename: string, headers: string[], rows: TableValue[][]) {
  const content = [headers, ...rows].map((row) => row.map(csvValue).join(',')).join('\n')
  const url = URL.createObjectURL(new Blob([content], { type: 'text/csv;charset=utf-8' }))
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  link.remove()
  window.setTimeout(() => URL.revokeObjectURL(url), 0)
}

function sortRows<T extends Record<string, TableValue>>(rows: T[], field: string, direction: 'asc' | 'desc') {
  if (!field) return rows

  return [...rows].sort((left, right) => {
    const leftValue = left[field]
    const rightValue = right[field]
    if (leftValue === undefined || rightValue === undefined) return 0

    const comparison = typeof leftValue === 'number' && typeof rightValue === 'number'
      ? leftValue - rightValue
      : String(leftValue).localeCompare(String(rightValue), undefined, { numeric: true, sensitivity: 'base' })

    return direction === 'asc' ? comparison : -comparison
  })
}

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
  const [activeTab, setActiveTab] = useState<ReportTab>('dashboard')
  const [summary, setSummary] = useState<ReportSummary>(emptySummary)
  const [stateData, setStateData] = useState<StateReport[]>([])
  const [hsnData, setHsnData] = useState<HSNReport[]>([])
  const [documentsData, setDocumentsData] = useState<DocumentReport[]>([])
  const [selectedSources, setSelectedSources] = useState<string[]>(allSources)
  const [isLoading, setIsLoading] = useState(true)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [isExporting, setIsExporting] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [gstin, setGstin] = useState('33AUSPR1909H1ZC')
  const [sortField, setSortField] = useState('')
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc')

  const queryString = useMemo(() => {
    const query = new URLSearchParams()
    if (startDate) query.set('start_date', dateToBoundary(startDate))
    if (endDate) query.set('end_date', dateToBoundary(endDate, true))
    if (selectedSources.length) query.set('source_ids', selectedSources.join(','))
    return query.toString()
  }, [endDate, selectedSources, startDate])

  const fetchReports = useCallback(async (silent = false) => {
    if (activeTab === 'gstr1') {
      setIsLoading(false)
      setIsRefreshing(false)
      return
    }

    if (silent) setIsRefreshing(true)
    else setIsLoading(true)
    setError('')
    setNotice('')

    const fetchJson = async (path: string) => {
      const response = await fetch(`${API_BASE}${path}${queryString ? `?${queryString}` : ''}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (response.status === 401) {
        onUnauthorized()
        throw new Error('Your session has expired. Please sign in again.')
      }
      if (!response.ok) throw new Error(`Reports request failed with status ${response.status}`)
      const data = await response.json() as ReportResponse
      if (!data.success) throw new Error(data.message || 'Report data was not returned')
      return data
    }

    try {
      if (activeTab === 'dashboard') {
        const result = await fetchJson('/api/reports/summary')
        setSummary((result.summary as ReportSummary) || emptySummary)
      } else if (activeTab === 'state') {
        const result = await fetchJson('/api/reports/state-wise')
        setStateData((result.data as StateReport[]) || [])
      } else if (activeTab === 'hsn') {
        const result = await fetchJson('/api/reports/hsn-wise')
        setHsnData((result.data as HSNReport[]) || [])
      } else if (activeTab === 'documents') {
        const result = await fetchJson('/api/reports/documents-issued')
        setDocumentsData((result.data as DocumentReport[]) || [])
      }
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to load GST reports')
    } finally {
      setIsLoading(false)
      setIsRefreshing(false)
    }
  }, [activeTab, onUnauthorized, queryString, token])

  useEffect(() => {
    setSortField('')
    setSortDirection('desc')
    void fetchReports()
  }, [activeTab, fetchReports])

  const toggleSource = (sourceId: string) => {
    setSelectedSources((current) => current.includes(sourceId)
      ? current.filter((id) => id !== sourceId)
      : [...current, sourceId])
  }

  const handleSort = (field: string) => {
    if (sortField === field) {
      setSortDirection((current) => current === 'asc' ? 'desc' : 'asc')
      return
    }
    setSortField(field)
    setSortDirection('desc')
  }

  const sortedStateData = useMemo(() => sortRows(stateData, sortField, sortDirection), [sortDirection, sortField, stateData])
  const sortedHsnData = useMemo(() => sortRows(hsnData, sortField, sortDirection), [hsnData, sortDirection, sortField])

  const renderSortButton = (label: string, field: string) => (
    <button className="report-sort-button" type="button" onClick={() => handleSort(field)}>
      <span>{label}</span>
      <span aria-hidden="true">{sortField === field ? (sortDirection === 'asc' ? '↑' : '↓') : '↕'}</span>
    </button>
  )

  const exportCurrentReport = () => {
    if (activeTab === 'dashboard') {
      downloadCsv('gst_summary.csv', ['Metric', 'Value'], [
        ['Total revenue', summary.total_revenue],
        ['Taxable value', summary.total_taxable_value],
        ['GST collected', summary.total_gst_collected],
        ['IGST', summary.total_igst],
        ['CGST', summary.total_cgst],
        ['SGST', summary.total_sgst],
        ['Total orders', summary.total_orders],
        ['Cancelled orders', summary.cancelled_orders],
        ['Fulfilled orders', summary.fulfilled_orders],
        ['Unfulfilled orders', summary.unfulfilled_orders],
        ['Paid orders', summary.paid_orders],
        ['Invoices generated', summary.invoices_generated],
      ])
    } else if (activeTab === 'state') {
      downloadCsv('b2c_state_wise.csv', ['State', 'Orders', 'Taxable value', 'IGST', 'CGST', 'SGST', 'Total GST', 'Revenue'], sortedStateData.map((row) => [row.state, row.orders, row.taxable_value, row.igst, row.cgst, row.sgst, row.total_gst, row.revenue]))
    } else if (activeTab === 'hsn') {
      downloadCsv('hsn_summary.csv', ['HSN code', 'Products', 'Quantity sold', 'Taxable value', 'IGST', 'CGST', 'SGST', 'Total GST', 'Revenue'], sortedHsnData.map((row) => [row.hsn_code, row.product_count, row.qty_sold, row.taxable_value, row.igst, row.cgst, row.sgst, row.total_gst, row.revenue]))
    } else if (activeTab === 'documents') {
      downloadCsv('documents_issued.csv', ['Document type', 'From serial', 'To serial', 'Total issued', 'Cancelled', 'Net issued'], documentsData.map((row) => [row.document_type, row.from_serial, row.to_serial, row.total_issued, row.cancelled, row.net_issued]))
    }
  }

  const handleExportGstr1 = async () => {
    const normalizedGstin = gstin.trim().toUpperCase()
    if (!normalizedGstin) {
      setError('Enter the filing GSTIN before generating the GSTR-1 file.')
      return
    }

    setIsExporting(true)
    setError('')
    setNotice('')
    const query = new URLSearchParams(queryString)
    query.set('gstin', normalizedGstin)

    try {
      const response = await fetch(`${API_BASE}/api/reports/gstr1-json?${query.toString()}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (response.status === 401) {
        onUnauthorized()
        throw new Error('Your session has expired. Please sign in again.')
      }
      if (!response.ok) {
        const message = await response.text()
        throw new Error(message || `GSTR-1 export failed with status ${response.status}`)
      }

      const blob = await response.blob()
      const url = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `GSTR1_${normalizedGstin}.json`
      document.body.appendChild(link)
      link.click()
      link.remove()
      window.setTimeout(() => URL.revokeObjectURL(url), 0)
      setNotice('GSTR-1 JSON downloaded for the selected period and channels.')
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to generate the GSTR-1 file')
    } finally {
      setIsExporting(false)
    }
  }

  const activeTabDetails = reportTabs.find((tab) => tab.id === activeTab) || reportTabs[0]
  const canExport = activeTab === 'dashboard' || activeTab === 'state' || activeTab === 'hsn' || activeTab === 'documents'

  return (
    <section className="workspace-page reports-page" aria-labelledby="reports-heading">
      <header className="workspace-page-header">
        <div>
          <p className="eyebrow">Operations / GST</p>
          <h2 id="reports-heading">GST reports</h2>
          <p>Review tax collection, document sequences, and GST-ready exports for the selected period.</p>
        </div>
        <div className="support-header-actions">
          {canExport && <button className="secondary-button" type="button" onClick={exportCurrentReport} disabled={isLoading}><Download size={15} aria-hidden="true" /> Export CSV</button>}
          <button className="secondary-button" type="button" onClick={() => void fetchReports(true)} disabled={isRefreshing || activeTab === 'gstr1'}>
            <RefreshCw size={15} className={isRefreshing ? 'spin' : undefined} aria-hidden="true" /> Refresh
          </button>
        </div>
      </header>

      <nav className="report-tabs" aria-label="GST report sections" role="tablist">
        {reportTabs.map((tab) => (
          <button
            className={`report-tab ${activeTab === tab.id ? 'report-tab-active' : ''}`}
            key={tab.id}
            type="button"
            role="tab"
            aria-selected={activeTab === tab.id}
            onClick={() => setActiveTab(tab.id)}
          >
            <span>{tab.label}</span>
            <small>{tab.description}</small>
          </button>
        ))}
      </nav>

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
        <button className="report-source-clear" type="button" onClick={() => setSelectedSources(selectedSources.length ? [] : allSources)}>
          {selectedSources.length ? 'Clear' : 'Select all'}
        </button>
      </div>

      <div className="report-tab-context">
        <div>
          <p className="eyebrow">{activeTabDetails.label}</p>
          <h3>{activeTabDetails.description}</h3>
        </div>
        <span className="orders-card-meta">Date range is shared across the workspace</span>
      </div>

      {error && (
        <div className="dashboard-error" role="alert">
          <CircleAlert size={18} aria-hidden="true" />
          <span>{error}</span>
          <button type="button" onClick={() => void fetchReports()}>Try again</button>
        </div>
      )}
      {notice && <div className="inventory-notice" role="status"><Check size={16} aria-hidden="true" /> {notice}</div>}

      {activeTab === 'dashboard' && (
        <>
          <div className="report-metrics-grid">
            <ReportMetric label="Revenue" value={isLoading ? '—' : formatCurrency(summary.total_revenue)} detail={`${formatCurrency(summary.total_taxable_value)} taxable`} />
            <ReportMetric label="GST collected" value={isLoading ? '—' : formatCurrency(summary.total_gst_collected)} detail={`IGST ${formatCurrency(summary.total_igst)}`} />
            <ReportMetric label="Orders" value={isLoading ? '—' : formatNumber(summary.total_orders)} detail={`${formatNumber(summary.fulfilled_orders)} fulfilled`} />
            <ReportMetric label="Invoices generated" value={isLoading ? '—' : formatNumber(summary.invoices_generated)} detail={`${formatNumber(summary.cancelled_orders)} cancelled`} />
          </div>

          <div className="report-summary-grid">
            <section className="reports-table-card report-tax-card">
              <div className="reports-table-heading">
                <div><p className="eyebrow">Tax split</p><h3>GST collected</h3></div>
              </div>
              <div className="report-tax-split">
                <div><span>IGST</span><strong>{isLoading ? '—' : formatCurrency(summary.total_igst)}</strong></div>
                <div><span>CGST</span><strong>{isLoading ? '—' : formatCurrency(summary.total_cgst)}</strong></div>
                <div><span>SGST</span><strong>{isLoading ? '—' : formatCurrency(summary.total_sgst)}</strong></div>
              </div>
            </section>
            <section className="reports-table-card">
              <div className="reports-table-heading">
                <div><p className="eyebrow">Operations</p><h3>Order status</h3></div>
              </div>
              <div className="orders-table-wrap">
                <table className="orders-table report-operations-table">
                  <caption className="sr-only">GST report order status</caption>
                  <tbody>
                    {[
                      ['Total orders', summary.total_orders],
                      ['Fulfilled orders', summary.fulfilled_orders],
                      ['Unfulfilled orders', summary.unfulfilled_orders],
                      ['Paid orders', summary.paid_orders],
                      ['Cancelled orders', summary.cancelled_orders],
                    ].map(([label, value]) => <tr key={String(label)}><td>{label}</td><td><strong>{isLoading ? '—' : formatNumber(Number(value))}</strong></td></tr>)}
                  </tbody>
                </table>
              </div>
            </section>
          </div>
        </>
      )}

      {activeTab === 'state' && (
        <section className="reports-table-card">
          <div className="reports-table-heading">
            <div><p className="eyebrow">B2C State-wise</p><h3>State-wise GST summary</h3></div>
            <span className="orders-card-meta">{isLoading ? 'Loading…' : `${stateData.length} states`}</span>
          </div>
          <div className="orders-table-wrap">
            <table className="orders-table reports-table reports-table-wide">
              <caption className="sr-only">B2C state-wise GST summary</caption>
              <thead><tr>
                <th aria-sort={sortField === 'state' ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>{renderSortButton('State', 'state')}</th>
                <th aria-sort={sortField === 'orders' ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>{renderSortButton('Orders', 'orders')}</th>
                <th aria-sort={sortField === 'taxable_value' ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>{renderSortButton('Taxable value', 'taxable_value')}</th>
                <th aria-sort={sortField === 'igst' ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>{renderSortButton('IGST', 'igst')}</th>
                <th aria-sort={sortField === 'cgst' ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>{renderSortButton('CGST', 'cgst')}</th>
                <th aria-sort={sortField === 'sgst' ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>{renderSortButton('SGST', 'sgst')}</th>
                <th aria-sort={sortField === 'total_gst' ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>{renderSortButton('Total GST', 'total_gst')}</th>
                <th aria-sort={sortField === 'revenue' ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>{renderSortButton('Revenue', 'revenue')}</th>
              </tr></thead>
              <tbody>
                {isLoading ? <tr><td colSpan={8} className="table-state">Loading state-wise report…</td></tr> : sortedStateData.length === 0 ? <tr><td colSpan={8} className="table-state">No state-wise data for this period.</td></tr> : sortedStateData.map((row) => (
                  <tr key={row.state}><td><strong>{row.state || 'Unknown'}</strong></td><td>{formatNumber(row.orders)}</td><td className="table-money">{formatCurrency(row.taxable_value)}</td><td className="table-money">{formatCurrency(row.igst)}</td><td className="table-money">{formatCurrency(row.cgst)}</td><td className="table-money">{formatCurrency(row.sgst)}</td><td className="table-money">{formatCurrency(row.total_gst)}</td><td className="table-money"><strong>{formatCurrency(row.revenue)}</strong></td></tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      )}

      {activeTab === 'hsn' && (
        <section className="reports-table-card">
          <div className="reports-table-heading">
            <div><p className="eyebrow">HSN Summary</p><h3>Outward supplies by HSN</h3></div>
            <span className="orders-card-meta">{isLoading ? 'Loading…' : `${hsnData.length} HSN codes`}</span>
          </div>
          <div className="orders-table-wrap">
            <table className="orders-table reports-table reports-table-wide">
              <caption className="sr-only">HSN-wise GST summary</caption>
              <thead><tr>
                <th aria-sort={sortField === 'hsn_code' ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>{renderSortButton('HSN code', 'hsn_code')}</th>
                <th aria-sort={sortField === 'product_count' ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>{renderSortButton('Products', 'product_count')}</th>
                <th aria-sort={sortField === 'qty_sold' ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>{renderSortButton('Quantity sold', 'qty_sold')}</th>
                <th aria-sort={sortField === 'taxable_value' ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>{renderSortButton('Taxable value', 'taxable_value')}</th>
                <th aria-sort={sortField === 'igst' ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>{renderSortButton('IGST', 'igst')}</th>
                <th aria-sort={sortField === 'cgst' ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>{renderSortButton('CGST', 'cgst')}</th>
                <th aria-sort={sortField === 'sgst' ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>{renderSortButton('SGST', 'sgst')}</th>
                <th aria-sort={sortField === 'total_gst' ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>{renderSortButton('Total GST', 'total_gst')}</th>
                <th aria-sort={sortField === 'revenue' ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>{renderSortButton('Revenue', 'revenue')}</th>
              </tr></thead>
              <tbody>
                {isLoading ? <tr><td colSpan={9} className="table-state">Loading HSN report…</td></tr> : sortedHsnData.length === 0 ? <tr><td colSpan={9} className="table-state">No HSN data for this period.</td></tr> : sortedHsnData.map((row) => (
                  <tr key={row.hsn_code}><td><strong className="mono-text">{row.hsn_code || 'Unknown'}</strong></td><td>{formatNumber(row.product_count)}</td><td>{formatNumber(row.qty_sold)}</td><td className="table-money">{formatCurrency(row.taxable_value)}</td><td className="table-money">{formatCurrency(row.igst)}</td><td className="table-money">{formatCurrency(row.cgst)}</td><td className="table-money">{formatCurrency(row.sgst)}</td><td className="table-money">{formatCurrency(row.total_gst)}</td><td className="table-money"><strong>{formatCurrency(row.revenue)}</strong></td></tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      )}

      {activeTab === 'documents' && (
        <section className="reports-table-card">
          <div className="reports-table-heading">
            <div><p className="eyebrow">Documents Issued / Table 13</p><h3>Invoice and note sequences</h3></div>
            <span className="orders-card-meta">{isLoading ? 'Loading…' : `${documentsData.length} document types`}</span>
          </div>
          <div className="orders-table-wrap">
            <table className="orders-table reports-table">
              <caption className="sr-only">GSTR-1 documents issued report</caption>
              <thead><tr><th>Document type</th><th>From serial</th><th>To serial</th><th>Total issued</th><th>Cancelled</th><th>Net issued</th></tr></thead>
              <tbody>
                {isLoading ? <tr><td colSpan={6} className="table-state">Loading documents report…</td></tr> : documentsData.length === 0 ? <tr><td colSpan={6} className="table-state">No documents issued in this period.</td></tr> : documentsData.map((row) => (
                  <tr key={row.document_type}><td><strong>{row.document_type}</strong></td><td className="mono-text">{row.from_serial || '—'}</td><td className="mono-text">{row.to_serial || '—'}</td><td>{formatNumber(row.total_issued)}</td><td>{formatNumber(row.cancelled)}</td><td><strong>{formatNumber(row.net_issued)}</strong></td></tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      )}

      {activeTab === 'gstr1' && (
        <div className="gstr1-grid">
          <section className="reports-table-card gstr1-card">
            <div className="reports-table-heading">
              <div><p className="eyebrow">Offline utility</p><h3>Generate GSTR-1 JSON</h3></div>
              <span className="report-icon-badge"><FileJson size={18} aria-hidden="true" /></span>
            </div>
            <form className="gstr1-form" onSubmit={(event) => { event.preventDefault(); void handleExportGstr1() }}>
              <label className="form-field"><span>Filing GSTIN</span><input value={gstin} onChange={(event) => setGstin(event.target.value.toUpperCase())} maxLength={15} placeholder="15-character GSTIN" autoComplete="off" /></label>
              <p className="form-help">The file uses the selected date range and channel filters. Change the GSTIN only when filing for another registration.</p>
              <button className="primary-button" type="submit" disabled={isExporting}><Download size={15} aria-hidden="true" /> {isExporting ? 'Generating…' : 'Download GSTR-1 JSON'}</button>
            </form>
          </section>
          <section className="reports-table-card gstr1-card">
            <div className="reports-table-heading"><div><p className="eyebrow">Included sections</p><h3>What this export contains</h3></div><span className="report-icon-badge"><Table2 size={18} aria-hidden="true" /></span></div>
            <ul className="gstr1-section-list">
              <li><span>4</span><div><strong>B2B supplies</strong><small>Registered customer invoices and rate-wise tax lines.</small></div></li>
              <li><span>7</span><div><strong>B2CS sales</strong><small>Consolidated B2C sales by place of supply and rate.</small></div></li>
              <li><span>9</span><div><strong>Credit and debit notes</strong><small>Registered customer adjustments included in CDNR.</small></div></li>
              <li><span>12</span><div><strong>HSN summary</strong><small>Quantity, taxable value, and tax split by HSN code.</small></div></li>
              <li><span>13</span><div><strong>Documents issued</strong><small>Invoice and note serial ranges with cancellations.</small></div></li>
            </ul>
          </section>
        </div>
      )}
    </section>
  )
}
