import { FormEvent, useCallback, useEffect, useRef, useState } from 'react'
import { useRealtimeRefresh } from '../../lib/realtime'
import { ChevronLeft, ChevronRight, CircleAlert, Clock3, Mail, MapPin, Pencil, Phone, Save, Search, SlidersHorizontal, UserRound, X } from 'lucide-react'
import { API_BASE } from '../../lib/api'
import { apiJson, apiRequest } from '../../lib/http'

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
  address1?: string
  address2?: string
  country?: string
  zip_code?: string
  external_id?: string
  created_at?: string
  total_orders?: number
  total_spent?: number
  updated_at?: string
  source_id?: string
}

type CustomerForm = {
  first_name: string
  last_name: string
  phone_number: string
  email: string
  address1: string
  address2: string
  city: string
  state: string
  country: string
  zip_code: string
}

type CustomerEvent = {
  id: number
  event_type: string
  source?: string
  occurred_at?: string
  diff_data?: Record<string, unknown> | null
}

type CustomerSortField = 'first_name' | 'phone_number' | 'city' | 'total_orders' | 'total_spent' | 'updated_at' | 'source_id'
type SortOrder = 'ASC' | 'DESC'

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

function sortIndicator(field: CustomerSortField, sortField: CustomerSortField, sortOrder: SortOrder) {
  if (field !== sortField) return '↕'
  return sortOrder === 'ASC' ? '↑' : '↓'
}

function customerFormFrom(customer: Customer): CustomerForm {
  return {
    first_name: customer.first_name || '', last_name: customer.last_name || '', phone_number: customer.phone_number || '', email: customer.email || '',
    address1: customer.address1 || '', address2: customer.address2 || '', city: customer.city || '', state: customer.state || '', country: customer.country || '', zip_code: customer.zip_code || '',
  }
}

function eventLabel(value: string) {
  return value.replace(/[_-]+/g, ' ').replace(/\b\w/g, (letter) => letter.toUpperCase())
}

function eventChangedFields(event: CustomerEvent) {
  if (!event.diff_data) return ''
  const fields = Object.keys(event.diff_data).filter((field) => field !== 'updated_at')
  return fields.length ? `Changed ${fields.join(', ')}` : ''
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
  const [sortField, setSortField] = useState<CustomerSortField>('updated_at')
  const [sortOrder, setSortOrder] = useState<SortOrder>('DESC')
  const [selectedCustomer, setSelectedCustomer] = useState<Customer | null>(null)
  const [customerForm, setCustomerForm] = useState<CustomerForm | null>(null)
  const [isEditingCustomer, setIsEditingCustomer] = useState(false)
  const [isSavingCustomer, setIsSavingCustomer] = useState(false)
  const [customerError, setCustomerError] = useState('')
  const [customerNotice, setCustomerNotice] = useState('')
  const [customerHistory, setCustomerHistory] = useState<CustomerEvent[]>([])
  const [isLoadingHistory, setIsLoadingHistory] = useState(false)
  const [isLoadingCustomer, setIsLoadingCustomer] = useState(false)
  const customerRequestRef = useRef(0)

  const handleSort = (field: CustomerSortField) => {
    setPage(1)
    if (field === sortField) {
      setSortOrder((current) => current === 'ASC' ? 'DESC' : 'ASC')
      return
    }
    setSortField(field)
    setSortOrder('ASC')
  }

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
      sortBy: sortField,
      sortOrder: sortOrder,
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
  }, [debouncedSearch, location, minOrders, minSpent, onUnauthorized, page, sortField, sortOrder, sourceFilter, token])

  useEffect(() => {
    void fetchCustomers()
  }, [fetchCustomers])
  useRealtimeRefresh(['customers.changed', 'orders.changed'], () => {
    void fetchCustomers()
    if (selectedCustomer && !isEditingCustomer && !isSavingCustomer) {
      const customerId = selectedCustomer.id
      void apiJson<{ customer?: Customer }>(token, onUnauthorized, `/api/customers/${customerId}`)
        .then((data) => {
          if (data.customer) {
            setSelectedCustomer((current) => current?.id === customerId ? data.customer! : current)
            setCustomerForm((current) => current ? customerFormFrom(data.customer!) : current)
          }
        })
        .catch(() => {})
    }
  })

  useEffect(() => {
    if (!selectedCustomer) return
    let active = true
    setIsLoadingHistory(true)
    void apiJson<{ customer_events?: CustomerEvent[] }>(token, onUnauthorized, `/api/customers/history?id=${selectedCustomer.id}&limit=8`)
      .then((data) => { if (active) setCustomerHistory(data.customer_events || []) })
      .catch(() => { if (active) setCustomerHistory([]) })
      .finally(() => { if (active) setIsLoadingHistory(false) })
    return () => { active = false }
  }, [onUnauthorized, selectedCustomer, token])

  const totalPages = Math.max(Math.ceil(total / pageSize), 1)

  const clearFilters = () => {
    setSearch('')
    setSourceFilter('')
    setMinSpent('')
    setMinOrders('')
    setLocation('')
  }

  const openCustomer = (customer: Customer) => {
    setSelectedCustomer(customer)
    setCustomerForm(customerFormFrom(customer))
    setIsEditingCustomer(false)
    setCustomerError('')
    setCustomerNotice('')
    setIsLoadingCustomer(true)
    const requestId = customerRequestRef.current + 1
    customerRequestRef.current = requestId
    void apiJson<{ customer?: Customer }>(token, onUnauthorized, `/api/customers/${customer.id}`)
      .then((data) => {
        if (data.customer && customerRequestRef.current === requestId) {
          setSelectedCustomer(data.customer)
          setCustomerForm(customerFormFrom(data.customer))
        }
      })
      .catch((caughtError) => { if (customerRequestRef.current === requestId) setCustomerError(caughtError instanceof Error ? caughtError.message : 'Unable to load the latest customer details') })
      .finally(() => { if (customerRequestRef.current === requestId) setIsLoadingCustomer(false) })
  }

  const closeCustomer = () => {
    if (isSavingCustomer) return
    customerRequestRef.current += 1
    setSelectedCustomer(null)
    setCustomerForm(null)
    setIsEditingCustomer(false)
    setCustomerError('')
    setCustomerNotice('')
    setCustomerHistory([])
  }

  const updateCustomerField = (field: keyof CustomerForm, value: string) => setCustomerForm((current) => current ? { ...current, [field]: value } : current)

  const saveCustomer = async (event: FormEvent) => {
    event.preventDefault()
    if (!selectedCustomer || !customerForm) return
    setIsSavingCustomer(true)
    setCustomerError('')
    setCustomerNotice('')
    try {
      await apiRequest(token, onUnauthorized, `/api/customers/${selectedCustomer.id}`, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(customerForm) })
      const updatedCustomer = { ...selectedCustomer, ...customerForm, updated_at: new Date().toISOString() }
      setSelectedCustomer(updatedCustomer)
      setCustomerForm(customerFormFrom(updatedCustomer))
      setCustomers((current) => current.map((customer) => customer.id === updatedCustomer.id ? updatedCustomer : customer))
      setIsEditingCustomer(false)
      setCustomerNotice('Customer details saved')
      void fetchCustomers()
    } catch (caughtError) {
      setCustomerError(caughtError instanceof Error ? caughtError.message : 'Unable to save customer details')
    } finally {
      setIsSavingCustomer(false)
    }
  }

  return (
    <section className="workspace-page customers-page" aria-labelledby="customers-heading">
      <header className="workspace-page-header">
        <div>
          <p className="eyebrow">Operations / Customers</p>
          <h2 id="customers-heading">Customers</h2>
          <p>Manage customer records, order history, and lifetime value.</p>
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
          <span className="orders-card-meta">{sortField === 'updated_at' && sortOrder === 'DESC' ? 'Recently active first' : 'Click a column to sort'}</span>
        </div>

        <div className="orders-table-wrap">
          <table className="orders-table customers-table">
            <thead>
              <tr>
                {([
                  ['Customer', 'first_name'],
                  ['Contact', 'phone_number'],
                  ['Location', 'city'],
                  ['Orders', 'total_orders'],
                  ['Lifetime spend', 'total_spent'],
                  ['Last activity', 'updated_at'],
                  ['Source', 'source_id'],
                ] as [string, CustomerSortField][]).map(([label, field]) => (
                  <th key={field} aria-sort={sortField === field ? (sortOrder === 'ASC' ? 'ascending' : 'descending') : 'none'}>
                    <button className="sortable-table-button" type="button" onClick={() => handleSort(field)} aria-label={`Sort by ${label}`}>
                      <span>{label}</span><span className="sortable-table-indicator" aria-hidden="true">{sortIndicator(field, sortField, sortOrder)}</span>
                    </button>
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr><td colSpan={7} className="table-state">Loading customers…</td></tr>
              ) : customers.length === 0 ? (
                <tr><td colSpan={7} className="table-state">No customers match these filters.</td></tr>
              ) : customers.map((customer) => (
                <tr key={customer.id} className="customer-row" tabIndex={0} role="button" aria-label={`Open ${displayName(customer)} customer profile`} onClick={() => openCustomer(customer)} onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') { event.preventDefault(); openCustomer(customer) } }}>
                  <td><button className="customer-name-button" type="button" onClick={(event) => { event.stopPropagation(); openCustomer(customer) }}>{displayName(customer)}</button></td>
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

      {selectedCustomer && customerForm && <div className="modal-scrim" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) closeCustomer() }}>
        <div className="modal-card customer-detail-modal" role="dialog" aria-modal="true" aria-labelledby="customer-detail-heading">
          <div className="customer-detail-heading"><div className="modal-heading-copy"><p className="eyebrow">Customer profile</p><h2 id="customer-detail-heading">{displayName(selectedCustomer)}</h2><p className="customer-detail-subtitle">{selectedCustomer.phone_number || 'No phone number'} · {selectedCustomer.source_id || 'manual'}</p></div><div className="customer-detail-header-actions">{!isEditingCustomer && <button className="secondary-button" type="button" disabled={isLoadingCustomer} onClick={() => { setIsEditingCustomer(true); setCustomerNotice(''); setCustomerError('') }}><Pencil size={14} aria-hidden="true" /> Edit details</button>}<button className="icon-button" type="button" aria-label="Close customer profile" onClick={closeCustomer}><X size={19} aria-hidden="true" /></button></div></div>
          {customerError && <div className="dashboard-error" role="alert"><CircleAlert size={16} aria-hidden="true" /><span>{customerError}</span></div>}
          {customerNotice && <div className="customer-detail-notice" role="status">{customerNotice}</div>}
          {isLoadingCustomer && <p className="customer-detail-loading" role="status">Loading the latest customer details…</p>}
          {isEditingCustomer ? <form className="customer-detail-form" onSubmit={saveCustomer}>
            <div className="customer-detail-section-heading"><div><p className="eyebrow">Editable details</p><h3>Keep this profile up to date</h3></div><span>Changes are saved to the customer record.</span></div>
            <div className="form-grid-two"><label className="form-field"><span>First name</span><input value={customerForm.first_name} onChange={(event) => updateCustomerField('first_name', event.target.value)} /></label><label className="form-field"><span>Last name</span><input value={customerForm.last_name} onChange={(event) => updateCustomerField('last_name', event.target.value)} /></label><label className="form-field"><span>Phone number</span><input required inputMode="tel" value={customerForm.phone_number} onChange={(event) => updateCustomerField('phone_number', event.target.value)} /></label><label className="form-field"><span>Email</span><input type="email" value={customerForm.email} onChange={(event) => updateCustomerField('email', event.target.value)} /></label><label className="form-field"><span>City</span><input value={customerForm.city} onChange={(event) => updateCustomerField('city', event.target.value)} /></label><label className="form-field"><span>State</span><input value={customerForm.state} onChange={(event) => updateCustomerField('state', event.target.value)} /></label><label className="form-field"><span>Country</span><input value={customerForm.country} onChange={(event) => updateCustomerField('country', event.target.value)} /></label><label className="form-field"><span>PIN code</span><input inputMode="numeric" value={customerForm.zip_code} onChange={(event) => updateCustomerField('zip_code', event.target.value)} /></label></div>
            <label className="form-field"><span>Address line 1</span><input value={customerForm.address1} onChange={(event) => updateCustomerField('address1', event.target.value)} /></label><label className="form-field"><span>Address line 2</span><input value={customerForm.address2} onChange={(event) => updateCustomerField('address2', event.target.value)} /></label>
            <div className="modal-actions"><button className="secondary-button" type="button" onClick={() => { setCustomerForm(customerFormFrom(selectedCustomer)); setIsEditingCustomer(false); setCustomerError('') }} disabled={isSavingCustomer}>Cancel</button><button className="primary-button" type="submit" disabled={isSavingCustomer}><Save size={14} aria-hidden="true" /> {isSavingCustomer ? 'Saving…' : 'Save details'}</button></div>
          </form> : <>
            <div className="customer-detail-summary"><div><span className="metric-label">Orders</span><strong>{(selectedCustomer.total_orders || 0).toLocaleString('en-IN')}</strong><small>Lifetime orders</small></div><div><span className="metric-label">Lifetime spend</span><strong>{formatMoney(selectedCustomer.total_spent)}</strong><small>Total value</small></div><div><span className="metric-label">Source</span><strong>{selectedCustomer.source_id || 'Manual'}</strong><small>Customer record</small></div></div>
            <div className="customer-detail-columns"><section className="customer-detail-section"><div className="customer-detail-section-heading"><div><p className="eyebrow">Contact</p><h3>How to reach them</h3></div><Phone size={18} aria-hidden="true" /></div><div className="customer-detail-fields"><div><span><UserRound size={14} aria-hidden="true" /> Name</span><strong>{displayName(selectedCustomer)}</strong></div><div><span><Phone size={14} aria-hidden="true" /> Phone</span><strong>{selectedCustomer.phone_number || 'Not provided'}</strong></div><div><span><Mail size={14} aria-hidden="true" /> Email</span><strong>{selectedCustomer.email || 'Not provided'}</strong></div></div></section><section className="customer-detail-section"><div className="customer-detail-section-heading"><div><p className="eyebrow">Address</p><h3>Delivery details</h3></div><MapPin size={18} aria-hidden="true" /></div><div className="customer-detail-fields"><div><span>Address</span><strong>{[selectedCustomer.address1, selectedCustomer.address2].filter(Boolean).join(', ') || 'Not provided'}</strong></div><div><span>Location</span><strong>{[selectedCustomer.city, selectedCustomer.state].filter(Boolean).join(', ') || 'Not provided'}</strong></div><div><span>Country / PIN</span><strong>{[selectedCustomer.country, selectedCustomer.zip_code].filter(Boolean).join(' · ') || 'Not provided'}</strong></div></div></section></div>
            <section className="customer-detail-section customer-detail-record-section"><div className="customer-detail-section-heading"><div><p className="eyebrow">Record details</p><h3>Customer metadata</h3></div></div><div className="customer-detail-fields customer-detail-fields-inline"><div><span>Customer ID</span><strong>#{selectedCustomer.id}</strong></div><div><span>Added</span><strong>{formatDate(selectedCustomer.created_at)}</strong></div><div><span>Last activity</span><strong>{formatDate(selectedCustomer.updated_at)}</strong></div><div><span>External ID</span><strong>{selectedCustomer.external_id || 'Not linked'}</strong></div></div></section>
            <section className="customer-detail-section customer-detail-history"><div className="customer-detail-section-heading"><div><p className="eyebrow">Activity</p><h3>Recent customer history</h3></div><Clock3 size={18} aria-hidden="true" /></div>{isLoadingHistory ? <p className="customer-detail-empty">Loading activity…</p> : customerHistory.length === 0 ? <p className="customer-detail-empty">No recorded activity for this customer yet.</p> : <div className="customer-history-list">{customerHistory.map((event) => <div className="customer-history-item" key={event.id}><span className="customer-history-dot" aria-hidden="true" /><div><strong>{eventLabel(event.event_type)}</strong><p>{eventChangedFields(event) || `${event.source || 'System'} update`}</p></div><time dateTime={event.occurred_at}>{formatDate(event.occurred_at)}</time></div>)}</div>}</section>
          </>}
        </div>
      </div>}
    </section>
  )
}
