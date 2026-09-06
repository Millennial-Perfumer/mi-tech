import { useCallback, useEffect, useState } from 'react'
import { ChevronLeft, ChevronRight, CircleAlert, Search, SlidersHorizontal } from 'lucide-react'
import { API_BASE } from '../../lib/api'

type CustomersPageProps = {
  token: string
  onUnauthorized: () => void
}

type Customer = {
  id: number
  phone_number: string
  first_name?: string
  last_name?: string
  email?: string
  city?: string
  state?: string
  total_orders?: number
  total_spent?: number
  updated_at?: string
  source_id?: string
}

const pageSize = 20

function displayName(customer: Customer) {
  const name = `${customer.first_name || ''} ${customer.last_name || ''}`.trim()
  return name || 'Unnamed customer'
}

function formatMoney(value: number | undefined) {
  return `₹${Number(value || 0).toLocaleString('en-IN', { maximumFractionDigits: 0 })}`
}

function formatDate(value: string | undefined) {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return date.toLocaleDateString('en-IN', { day: 'numeric', month: 'short', year: 'numeric' })
}

export function CustomersPage({ token, onUnauthorized }: CustomersPageProps) {
  const [customers, setCustomers] = useState<Customer[]>([])
  const [search, setSearch] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const [sourceFilter, setSourceFilter] = useState('')
  const [minSpent, setMinSpent] = useState('')
  const [minOrders, setMinOrders] = useState('')
  const [location, setLocation] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedSearch(search.trim()), 250)
    return () => window.clearTimeout(timer)
  }, [search])

  useEffect(() => {
    setPage(1)
  }, [debouncedSearch, location, minOrders, minSpent, sourceFilter])

  const fetchCustomers = useCallback(async () => {
    setIsLoading(true)
    setError('')

    const query = new URLSearchParams({
      page: String(page),
      pageSize: String(pageSize),
      sortBy: 'updated_at',
      sortOrder: 'DESC',
    })
    if (debouncedSearch) query.set('search', debouncedSearch)
    if (sourceFilter) query.set('source_id', sourceFilter)
    if (minSpent) query.set('min_spent', minSpent)
    if (minOrders) query.set('min_orders', minOrders)
    if (location) query.set('city', location)

    try {
      const response = await fetch(`${API_BASE}/api/customers?${query.toString()}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (response.status === 401) {
        onUnauthorized()
        throw new Error('Your session has expired. Please sign in again.')
      }
      if (!response.ok) throw new Error(`Customers request failed with status ${response.status}`)
      const data = await response.json() as { success?: boolean; customers?: Customer[]; total?: number; message?: string }
      if (!data.success) throw new Error(data.message || 'Customers were not returned')
      setCustomers(data.customers || [])
      setTotal(data.total || 0)
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to load customers')
      setCustomers([])
      setTotal(0)
    } finally {
      setIsLoading(false)
    }
  }, [debouncedSearch, location, minOrders, minSpent, onUnauthorized, page, sourceFilter, token])

  useEffect(() => {
    void fetchCustomers()
  }, [fetchCustomers])

  const totalPages = Math.max(Math.ceil(total / pageSize), 1)

  const clearFilters = () => {
    setSearch('')
    setSourceFilter('')
    setMinSpent('')
    setMinOrders('')
    setLocation('')
  }

  return (
    <section className="workspace-page customers-page" aria-labelledby="customers-heading">
      <header className="workspace-page-header">
        <div>
          <p className="eyebrow">Operations / Customers</p>
          <h2 id="customers-heading">Know the people behind every order.</h2>
          <p>Find customers quickly, understand lifetime value, and keep the directory useful for the whole team.</p>
        </div>
        <span className="page-period-note">Customer totals are lifetime values</span>
      </header>

      <div className="customers-toolbar" aria-label="Customer filters">
        <label className="orders-search">
          <Search size={16} aria-hidden="true" />
          <span className="sr-only">Search customers</span>
          <input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search name, phone, or email" />
        </label>
        <label className="compact-select">
          <span className="sr-only">Customer source</span>
          <select value={sourceFilter} onChange={(event) => setSourceFilter(event.target.value)}>
            <option value="">All sources</option>
            <option value="shopify">Shopify</option>
            <option value="amazon">Amazon</option>
            <option value="pos">POS</option>
            <option value="b2b">B2B</option>
          </select>
        </label>
        <label className="customer-number-filter">
          <span className="sr-only">Minimum spend</span>
          <input type="number" min="0" value={minSpent} onChange={(event) => setMinSpent(event.target.value)} placeholder="Min spend" />
        </label>
        <label className="customer-number-filter">
          <span className="sr-only">Minimum orders</span>
          <input type="number" min="0" value={minOrders} onChange={(event) => setMinOrders(event.target.value)} placeholder="Min orders" />
        </label>
        <label className="customer-location-filter">
          <span className="sr-only">City</span>
          <input value={location} onChange={(event) => setLocation(event.target.value)} placeholder="City" />
        </label>
        <button className="clear-customer-filters" type="button" onClick={clearFilters}><SlidersHorizontal size={15} aria-hidden="true" /> Clear</button>
      </div>

      {error && (
        <div className="dashboard-error" role="alert">
          <CircleAlert size={18} aria-hidden="true" />
          <span>{error}</span>
          <button type="button" onClick={() => void fetchCustomers()}>Try again</button>
        </div>
      )}

      <div className="customers-card">
        <div className="orders-card-heading">
          <div>
            <p className="eyebrow">Customer directory</p>
            <h3>{isLoading ? 'Loading customers…' : `${total.toLocaleString('en-IN')} customers found`}</h3>
          </div>
          <span className="orders-card-meta">Recently active first</span>
        </div>

        <div className="orders-table-wrap">
          <table className="orders-table customers-table">
            <thead>
              <tr><th>Customer</th><th>Contact</th><th>Location</th><th>Orders</th><th>Lifetime spend</th><th>Last activity</th><th>Source</th></tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr><td colSpan={7} className="table-state">Loading customers…</td></tr>
              ) : customers.length === 0 ? (
                <tr><td colSpan={7} className="table-state">No customers match these filters.</td></tr>
              ) : customers.map((customer) => (
                <tr key={customer.id}>
                  <td><strong>{displayName(customer)}</strong></td>
                  <td><span className="customer-contact">{customer.phone_number}<small>{customer.email || 'No email'}</small></span></td>
                  <td>{customer.city || customer.state ? `${customer.city || ''}${customer.city && customer.state ? ', ' : ''}${customer.state || ''}` : '—'}</td>
                  <td>{(customer.total_orders || 0).toLocaleString('en-IN')}</td>
                  <td><strong>{formatMoney(customer.total_spent)}</strong></td>
                  <td className="table-muted">{formatDate(customer.updated_at)}</td>
                  <td><span className="channel-label">{customer.source_id || 'manual'}</span></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <footer className="orders-pagination">
          <span>{total ? `${((page - 1) * pageSize) + 1}–${Math.min(page * pageSize, total)} of ${total.toLocaleString('en-IN')}` : '0 customers'}</span>
          <div>
            <button type="button" aria-label="Previous page" disabled={page <= 1 || isLoading} onClick={() => setPage((current) => Math.max(current - 1, 1))}><ChevronLeft size={16} aria-hidden="true" /></button>
            <span>Page {page} of {totalPages}</span>
            <button type="button" aria-label="Next page" disabled={page >= totalPages || isLoading} onClick={() => setPage((current) => Math.min(current + 1, totalPages))}><ChevronRight size={16} aria-hidden="true" /></button>
          </div>
        </footer>
      </div>
    </section>
  )
}
