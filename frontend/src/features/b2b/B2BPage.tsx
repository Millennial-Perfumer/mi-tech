import { FormEvent, useCallback, useEffect, useMemo, useState } from 'react'
import { Building2, CircleAlert, FilePlus2, Pencil, RefreshCw, Search, Trash2, X } from 'lucide-react'
import { apiJson, apiRequest, arrayFrom, formatDate, formatMoney, numberValue, textValue } from '../../lib/http'
import { B2BDocumentKind, B2BDocumentMode, B2BDocumentModal } from './B2BDocumentModal'

type Props = { token: string; onUnauthorized: () => void }
type Tab = B2BDocumentKind | 'customers'
type Row = Record<string, unknown>
type CustomerForm = { legal_name: string; trade_name: string; gstin: string; email: string; phone: string; billing_address: string; shipping_address: string; state: string; state_code: string; pan: string }

const tabs: { id: Tab; label: string }[] = [
  { id: 'invoices', label: 'Invoices' },
  { id: 'proformas', label: 'Proformas' },
  { id: 'customers', label: 'Customers' },
  { id: 'credit-notes', label: 'Credit notes' },
  { id: 'debit-notes', label: 'Debit notes' },
]

const emptyCustomer: CustomerForm = { legal_name: '', trade_name: '', gstin: '', email: '', phone: '', billing_address: '', shipping_address: '', state: '', state_code: '', pan: '' }

function statusTone(status: string) {
  const normalized = status.toLowerCase()
  return normalized.includes('cancel') || normalized.includes('reject') ? 'danger' : normalized.includes('paid') || normalized.includes('issue') || normalized.includes('accept') ? 'success' : 'neutral'
}

function documentLabel(tab: Tab) {
  return tab === 'credit-notes' ? 'credit notes' : tab === 'debit-notes' ? 'debit notes' : tab
}

function customerName(row: Row) {
  return textValue(row.trade_name || row.legal_name, 'Unnamed business')
}

export function B2BPage({ token, onUnauthorized }: Props) {
  const [tab, setTab] = useState<Tab>('invoices')
  const [rows, setRows] = useState<Row[]>([])
  const [customers, setCustomers] = useState<Row[]>([])
  const [invoices, setInvoices] = useState<Row[]>([])
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [isWorking, setIsWorking] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [documentModal, setDocumentModal] = useState<{ kind: B2BDocumentKind; mode: B2BDocumentMode; record?: Row }>()
  const [isCustomerModalOpen, setIsCustomerModalOpen] = useState(false)
  const [editingCustomerId, setEditingCustomerId] = useState<string | null>(null)
  const [customerForm, setCustomerForm] = useState<CustomerForm>(emptyCustomer)

  const load = useCallback(async () => {
    setIsLoading(true)
    setError('')
    try {
      const [list, customerList, invoiceList] = await Promise.all([
        apiJson<unknown>(token, onUnauthorized, tab === 'customers' ? '/api/b2b/customers' : `/api/b2b/${tab}`),
        apiJson<unknown>(token, onUnauthorized, '/api/b2b/customers'),
        apiJson<unknown>(token, onUnauthorized, '/api/b2b/invoices'),
      ])
      setRows(tab === 'customers' ? arrayFrom(list, 'customers') : arrayFrom(list, tab === 'credit-notes' ? 'credit_notes' : tab === 'debit-notes' ? 'debit_notes' : tab))
      setCustomers(arrayFrom(customerList, 'customers'))
      setInvoices(arrayFrom(invoiceList, 'invoices'))
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to load B2B data')
    } finally {
      setIsLoading(false)
    }
  }, [onUnauthorized, tab, token])

  useEffect(() => { void load() }, [load])

  const filteredRows = useMemo(() => {
    const query = search.trim().toLowerCase()
    return rows.filter((row) => {
      const haystack = Object.values(row).join(' ').toLowerCase()
      return (!query || haystack.includes(query)) && (!status || textValue(row.status, '').toLowerCase() === status.toLowerCase())
    })
  }, [rows, search, status])

  const openCreate = () => {
    if (tab === 'customers') {
      setEditingCustomerId(null)
      setCustomerForm(emptyCustomer)
      setIsCustomerModalOpen(true)
      return
    }
    setDocumentModal({ kind: tab, mode: 'create' })
  }

  const openEditCustomer = (row: Row) => {
    setEditingCustomerId(textValue(row.id, ''))
    setCustomerForm({
      legal_name: textValue(row.legal_name, ''),
      trade_name: textValue(row.trade_name, ''),
      gstin: textValue(row.gstin, ''),
      email: textValue(row.email, ''),
      phone: textValue(row.phone, ''),
      billing_address: textValue(row.billing_address, ''),
      shipping_address: textValue(row.shipping_address, ''),
      state: textValue(row.state, ''),
      state_code: textValue(row.state_code, ''),
      pan: textValue(row.pan, ''),
    })
    setIsCustomerModalOpen(true)
  }

  const saveCustomer = async (event: FormEvent) => {
    event.preventDefault()
    setIsWorking(true)
    setError('')
    try {
      await apiRequest(token, onUnauthorized, '/api/b2b/customers', { method: editingCustomerId ? 'PUT' : 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ ...customerForm, id: editingCustomerId ? Number(editingCustomerId) : undefined }) })
      setNotice(`Business customer ${editingCustomerId ? 'updated' : 'added'}`)
      setIsCustomerModalOpen(false)
      setEditingCustomerId(null)
      setCustomerForm(emptyCustomer)
      await load()
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to save customer')
    } finally {
      setIsWorking(false)
    }
  }

  const deleteCustomer = async (row: Row) => {
    const id = textValue(row.id, '')
    if (!id || !window.confirm(`Delete ${customerName(row)}? Documents already created will keep their customer snapshot.`)) return
    setIsWorking(true)
    try {
      await apiRequest(token, onUnauthorized, `/api/b2b/customers?id=${id}`, { method: 'DELETE' })
      setNotice('Business customer deleted')
      await load()
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to delete customer')
    } finally {
      setIsWorking(false)
    }
  }

  const performDocumentAction = async (action: string, payload?: Record<string, string>) => {
    const modal = documentModal
    if (!modal?.record) return
    const id = textValue(modal.record.id, '')
    if (!id) return
    if (action === 'edit') {
      setDocumentModal({ ...modal, mode: 'edit' })
      return
    }
    if (action === 'delete') {
      if (!window.confirm(`Delete this ${modal.kind === 'credit-notes' ? 'credit note' : modal.kind === 'debit-notes' ? 'debit note' : modal.kind === 'proformas' ? 'proforma' : 'invoice'}?`)) return
      setIsWorking(true)
      try {
        await apiRequest(token, onUnauthorized, `/api/b2b/${modal.kind}?id=${id}`, { method: 'DELETE' })
        setNotice('Document deleted')
        setDocumentModal(undefined)
        await load()
      } catch (caughtError) {
        setError(caughtError instanceof Error ? caughtError.message : 'Unable to delete document')
      } finally {
        setIsWorking(false)
      }
      return
    }
    const endpointByAction: Record<string, string> = {
      issue: 'issue',
      cancel: 'cancel',
      accept: 'accept',
      reject: 'reject',
      'deduct-inventory': 'deduct-inventory',
      'revert-inventory': 'revert-inventory',
      convert: 'convert',
    }
    if (action === 'payment') {
      setIsWorking(true)
      try {
        await apiRequest(token, onUnauthorized, '/api/b2b/invoices/payment', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ id: Number(id), paid_amount: numberValue(payload?.paid_amount), payment_method: payload?.payment_method || 'Other' }) })
        setNotice('Payment recorded')
        setDocumentModal(undefined)
        await load()
      } catch (caughtError) {
        setError(caughtError instanceof Error ? caughtError.message : 'Unable to record payment')
      } finally {
        setIsWorking(false)
      }
      return
    }
    const endpoint = endpointByAction[action]
    if (!endpoint) return
    if (['cancel', 'reject', 'deduct-inventory', 'revert-inventory'].includes(action) && !window.confirm(`Confirm ${action.replaceAll('-', ' ')} for this document?`)) return
    setIsWorking(true)
    try {
      await apiRequest(token, onUnauthorized, `/api/b2b/${modal.kind}/${endpoint}?id=${id}`, { method: 'POST' })
      const actionLabel: Record<string, string> = { issue: 'issued', cancel: 'cancelled', accept: 'accepted', reject: 'rejected', 'deduct-inventory': 'deducted from inventory', 'revert-inventory': 'reverted in inventory' }
      setNotice(action === 'convert' ? 'Proforma converted to invoice' : `Document ${actionLabel[action] || action.replaceAll('-', ' ')}.`)
      setDocumentModal(undefined)
      await load()
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Document action failed')
    } finally {
      setIsWorking(false)
    }
  }

  const title = tab === 'customers' ? 'Customers' : tab === 'credit-notes' ? 'Credit notes' : tab === 'debit-notes' ? 'Debit notes' : tab[0].toUpperCase() + tab.slice(1)
  const isCustomerTab = tab === 'customers'

  return <section className="workspace-page b2b-page" aria-labelledby="b2b-heading">
    <header className="workspace-page-header"><div><p className="eyebrow">Operations / B2B billing</p><h2 id="b2b-heading">B2B billing</h2><p>Manage business customers, invoices, proformas, and adjustments.</p></div><div className="support-header-actions"><button className="secondary-button" type="button" onClick={() => void load()} disabled={isLoading}><RefreshCw size={15} className={isLoading ? 'spin' : undefined} aria-hidden="true" /> Refresh</button><button className="primary-button" type="button" onClick={openCreate}><FilePlus2 size={15} aria-hidden="true" /> {isCustomerTab ? 'Add customer' : `New ${tab === 'credit-notes' ? 'credit note' : tab === 'debit-notes' ? 'debit note' : tab.slice(0, -1)}`}</button></div></header>
    {error && <div className="dashboard-error" role="alert"><CircleAlert size={18} aria-hidden="true" /><span>{error}</span><button type="button" onClick={() => void load()}>Try again</button></div>}
    {notice && <div className="inventory-notice" role="status">{notice}</div>}
    <div className="b2b-summary-grid"><article className="report-metric-card"><span className="metric-label">Invoices</span><strong>{invoices.length}</strong><small>Drafts and issued documents</small></article><article className="report-metric-card"><span className="metric-label">Customers</span><strong>{customers.length}</strong><small>Business accounts</small></article><article className="report-metric-card"><span className="metric-label">Open value</span><strong>{formatMoney(invoices.reduce((sum, row) => sum + (textValue(row.payment_status, '').toLowerCase() === 'paid' ? 0 : numberValue(row.balance_amount || row.total_price)), 0))}</strong><small>Outstanding invoice balance</small></article></div>
    <section className="reports-table-card"><div className="b2b-tabs" role="tablist" aria-label="B2B sections">{tabs.map((item) => <button key={item.id} className={`filter-chip ${tab === item.id ? 'filter-chip-active' : ''}`} type="button" role="tab" aria-selected={tab === item.id} onClick={() => { setTab(item.id); setSearch(''); setStatus('') }}>{item.label}</button>)}</div><div className="b2b-toolbar"><label className="orders-search"><Search size={16} aria-hidden="true" /><span className="sr-only">Search {title}</span><input value={search} onChange={(event) => setSearch(event.target.value)} placeholder={`Search ${title.toLowerCase()}`} /></label>{!isCustomerTab && <label className="compact-select"><span className="sr-only">Filter status</span><select value={status} onChange={(event) => setStatus(event.target.value)}><option value="">All statuses</option><option value="DRAFT">Draft</option><option value="SENT">Sent</option><option value="ISSUED">Issued</option><option value="ACCEPTED">Accepted</option><option value="PAID">Paid</option><option value="CANCELLED">Cancelled</option></select></label>}</div><div className="orders-table-wrap"><table className="orders-table"><caption className="sr-only">{title}</caption><thead><tr>{isCustomerTab ? <><th>Business</th><th>GSTIN</th><th>Contact</th><th>State</th><th>Added</th><th>Actions</th></> : <><th>Number</th><th>Customer</th><th>Date</th><th>Status</th><th>Total</th><th>Actions</th></>}</tr></thead><tbody>{isLoading ? <tr><td className="table-state" colSpan={6}>Loading {documentLabel(tab)}…</td></tr> : filteredRows.length === 0 ? <tr><td className="table-state" colSpan={6}>No {documentLabel(tab)} found.</td></tr> : filteredRows.map((row, index) => { const id = textValue(row.id, String(index)); const rowStatus = textValue(row.status, 'DRAFT'); const number = textValue(row.invoice_number || row.proforma_number || row.credit_note_number || row.debit_note_number || row.number, `Document ${id}`); const customerRecord = row.customer && typeof row.customer === 'object' ? row.customer as Row : undefined; return isCustomerTab ? <tr key={id}><td><strong>{customerName(row)}</strong><small className="table-subtext">{textValue(row.legal_name, '')}</small></td><td className="mono-text">{textValue(row.gstin)}</td><td><span>{textValue(row.email, 'No email')}</span><small className="table-subtext">{textValue(row.phone, '')}</small></td><td>{textValue(row.state, '—')}</td><td>{formatDate(row.created_at)}</td><td><div className="table-action-group"><button className="table-link-button" type="button" onClick={() => openEditCustomer(row)}><Pencil size={13} aria-hidden="true" /> Edit</button><button className="table-link-button danger-link" type="button" disabled={isWorking} onClick={() => void deleteCustomer(row)}><Trash2 size={13} aria-hidden="true" /> Delete</button></div></td></tr> : <tr key={id}><td className="mono-text">{number}</td><td>{textValue(row.customer_name || customerRecord?.name, '—')}<small className="table-subtext">{textValue(row.customer_gstin, '')}</small></td><td>{formatDate(row.invoice_date || row.note_date || row.created_at)}</td><td><span className={`status-pill status-pill-${statusTone(rowStatus)}`}>{rowStatus}</span></td><td className="table-money">{formatMoney(row.total_price || row.total_amount || row.amount)}</td><td><div className="table-action-group"><button className="table-link-button" type="button" onClick={() => setDocumentModal({ kind: tab, mode: 'view', record: row })}>View</button>{rowStatus.toLowerCase() === 'draft' && <button className="table-link-button" type="button" disabled={isWorking} onClick={() => setDocumentModal({ kind: tab, mode: 'view', record: row })}>Issue</button>}</div></td></tr> })}</tbody></table></div></section>
    {isCustomerModalOpen && <div className="modal-scrim" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) setIsCustomerModalOpen(false) }}><form className="modal-card" role="dialog" aria-modal="true" aria-labelledby="customer-modal-heading" onSubmit={saveCustomer}><div className="modal-heading"><div><p className="eyebrow">B2B directory</p><h2 id="customer-modal-heading">{editingCustomerId ? 'Edit business customer' : 'Add business customer'}</h2></div><button className="icon-button" type="button" aria-label="Close" onClick={() => setIsCustomerModalOpen(false)}><X size={19} aria-hidden="true" /></button></div><div className="form-grid-two"><label className="form-field"><span>Legal name</span><input required value={customerForm.legal_name} onChange={(event) => setCustomerForm({ ...customerForm, legal_name: event.target.value })} /></label><label className="form-field"><span>Trade name</span><input value={customerForm.trade_name} onChange={(event) => setCustomerForm({ ...customerForm, trade_name: event.target.value })} /></label><label className="form-field"><span>GSTIN</span><input required value={customerForm.gstin} onChange={(event) => setCustomerForm({ ...customerForm, gstin: event.target.value.toUpperCase() })} /></label><label className="form-field"><span>PAN</span><input value={customerForm.pan} onChange={(event) => setCustomerForm({ ...customerForm, pan: event.target.value.toUpperCase() })} /></label><label className="form-field"><span>Email</span><input type="email" value={customerForm.email} onChange={(event) => setCustomerForm({ ...customerForm, email: event.target.value })} /></label><label className="form-field"><span>Phone</span><input value={customerForm.phone} onChange={(event) => setCustomerForm({ ...customerForm, phone: event.target.value })} /></label><label className="form-field"><span>State</span><input value={customerForm.state} onChange={(event) => setCustomerForm({ ...customerForm, state: event.target.value })} /></label><label className="form-field"><span>State code</span><input value={customerForm.state_code} onChange={(event) => setCustomerForm({ ...customerForm, state_code: event.target.value })} /></label></div><label className="form-field"><span>Billing address</span><textarea required rows={3} value={customerForm.billing_address} onChange={(event) => setCustomerForm({ ...customerForm, billing_address: event.target.value })} /></label><label className="form-field"><span>Shipping address</span><textarea rows={2} value={customerForm.shipping_address} onChange={(event) => setCustomerForm({ ...customerForm, shipping_address: event.target.value })} /></label><div className="modal-actions"><button className="secondary-button" type="button" onClick={() => setIsCustomerModalOpen(false)}>Cancel</button><button className="primary-button" type="submit" disabled={isWorking}><Building2 size={14} aria-hidden="true" /> {isWorking ? 'Saving…' : 'Save customer'}</button></div></form></div>}
    {documentModal && <B2BDocumentModal {...documentModal} customers={customers} invoices={invoices} token={token} onUnauthorized={onUnauthorized} onClose={() => setDocumentModal(undefined)} onSaved={() => { setNotice(`${documentModal.kind === 'credit-notes' ? 'Credit note' : documentModal.kind === 'debit-notes' ? 'Debit note' : documentModal.kind === 'proformas' ? 'Proforma' : 'Invoice'} saved`); setDocumentModal(undefined); void load() }} onAction={(action, payload) => void performDocumentAction(action, payload)} />}
  </section>
}
