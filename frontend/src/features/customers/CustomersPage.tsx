import { FormEvent, useCallback, useEffect, useRef, useState } from 'react'
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
  total_orders?: number
  total_spent?: number
  address1?: string
  address2?: string
  country?: string
  zip_code?: string
  external_id?: string
  created_at?: string
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
  actor_type?: string
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
    first_name: customer.first_name || '',
    last_name: customer.last_name || '',
    phone_number: customer.phone_number || '',
    email: customer.email || '',
    address1: customer.address1 || '',
    address2: customer.address2 || '',
    city: customer.city || '',
    state: customer.state || '',
    country: customer.country || '',
    zip_code: customer.zip_code || '',
  }
}

function eventLabel(eventType: string) {
  return eventType
    .replace(/[_-]+/g, ' ')
    .replace(/\b\w/g, (letter) => letter.toUpperCase())
}

function eventChangedFields(event: CustomerEvent) {
  if (!event.diff_data || typeof event.diff_data !== 'object') return ''
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

  useEffect(() => {
    if (!selectedCustomer) return

    let isCurrent = true
    setIsLoadingHistory(true)
    setCustomerHistory([])

    void apiJson<{ customer_events?: CustomerEvent[] }>(token, onUnauthorized, `/api/customers/history?id=${selectedCustomer.id}&limit=8`)
      .then((data) => {
        if (isCurrent) setCustomerHistory(data.customer_events || [])
      })
      .catch(() => {
        if (isCurrent) setCustomerHistory([])
      })
      .finally(() => {
        if (isCurrent) setIsLoadingHistory(false)
      })

    return () => {
      isCurrent = false
    }
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
        const detailedCustomer = data.customer
        if (!detailedCustomer || customerRequestRef.current !== requestId) return
        setSelectedCustomer(detailedCustomer)
        setCustomerForm(customerFormFrom(detailedCustomer))
      })
      .catch((caughtError) => {
        if (customerRequestRef.current === requestId) setCustomerError(caughtError instanceof Error ? caughtError.message : 'Unable to load the latest customer details')
      })
      .finally(() => {
        if (customerRequestRef.current === requestId) setIsLoadingCustomer(false)
      })
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

  const updateCustomerField = (field: keyof CustomerForm, value: string) => {
    setCustomerForm((current) => current ? { ...current, [field]: value } : current)
  }

  const saveCustomer = async (event: FormEvent) => {
    event.preventDefault()
    if (!selectedCustomer || !customerForm) return

    setIsSavingCustomer(true)
    setCustomerError('')
    setCustomerNotice('')

    try {
      await apiRequest(token, onUnauthorized, `/api/customers/${selectedCustomer.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(customerForm),
      })

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
  