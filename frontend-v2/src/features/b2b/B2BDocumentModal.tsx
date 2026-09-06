import { FormEvent, useEffect, useMemo, useState } from 'react'
import { FilePlus2, Minus, Plus, Save, X } from 'lucide-react'
import { apiRequest, formatDate, formatMoney, numberValue, textValue } from '../../lib/http'

export type B2BDocumentKind = 'invoices' | 'proformas' | 'credit-notes' | 'debit-notes'
export type B2BDocumentMode = 'create' | 'edit' | 'view'
type Row = Record<string, unknown>

type Props = {
  kind: B2BDocumentKind
  mode: B2BDocumentMode
  record?: Row
  customers: Row[]
  invoices: Row[]
  token: string
  onUnauthorized: () => void
  onClose: () => void
  onSaved: () => void
  onAction: (action: string, payload?: Record<string, string>) => void
}

type LineItem = {
  id?: number
  product_id: string
  item_details: string
  sku: string
  hsn_code: string
  quantity: string
  rate: string
}

type FormState = {
  id?: number
  customer_id: string
  customer_name: string
  customer_gstin: string
  customer_email: string
  customer_phone: string
  customer_state: string
  customer_state_code: string
  customer_address: string
  customer_shipping_address: string
  invoice_id: string
  invoice_date: string
  due_date: string
  note_date: string
  valid_until: string
  order_number: string
  terms: string
  reason: string
  discount_percent: string
  transportation_charge: string
  tds_tcs_type: string
  tds_tcs_rate: string
  paid_amount: string
  payment_method: string
  items: LineItem[]
}

const today = () => new Date().toISOString().slice(0, 10)

function dateInput(value: unknown, fallback = today()) {
  if (!value) return fallback
  const parsed = new Date(String(value))
  return Number.isNaN(parsed.getTime()) ? fallback : parsed.toISOString().slice(0, 10)
}

function lineItem(value?: Row): LineItem {
  return {
    id: typeof value?.id === 'number' ? value.id : undefined,
    product_id: textValue(value?.product_id, ''),
    item_details: textValue(value?.item_details, ''),
    sku: textValue(value?.sku, ''),
    hsn_code: textValue(value?.hsn_code, ''),
    quantity: value ? String(numberValue(value.quantity)) : '1',
    rate: value ? String(numberValue(value.rate)) : '0',
  }
}

function kindLabel(kind: B2BDocumentKind) {
  return kind === 'invoices' ? 'invoice' : kind === 'proformas' ? 'proforma' : kind === 'credit-notes' ? 'credit note' : 'debit note'
}

function collectionLabel(kind: B2BDocumentKind) {
  return kindLabel(kind).replace(/^./, (letter) => letter.toUpperCase())
}

function createForm(kind: B2BDocumentKind, record?: Row): FormState {
  const items = Array.isArray(record?.items) ? record.items.map((item) => lineItem(item as Row)) : [lineItem()]
  return {
    id: typeof record?.id === 'number' ? record.id : undefined,
    customer_id: textValue(record?.customer_id, ''),
    customer_name: textValue(record?.customer_name, ''),
    customer_gstin: textValue(record?.customer_gstin, ''),
    customer_email: textValue(record?.customer_email, ''),
    customer_phone: textValue(record?.customer_phone, ''),
    customer_state: textValue(record?.customer_state, ''),
    customer_state_code: textValue(record?.customer_state_code, ''),
    customer_address: textValue(record?.customer_address, ''),
    customer_shipping_address: textValue(record?.customer_shipping_address, ''),
    invoice_id: textValue(record?.invoice_id, ''),
    invoice_date: dateInput(record?.invoice_date),
    due_date: dateInput(record?.due_date, ''),
    note_date: dateInput(record?.note_date),
    valid_until: dateInput(record?.valid_until, ''),
    order_number: textValue(record?.order_number, ''),
    terms: textValue(record?.terms, ''),
    reason: textValue(record?.reason, ''),
    discount_percent: String(numberValue(record?.discount_percent)),
    transportation_charge: String(numberValue(record?.transportation_charge)),
    tds_tcs_type: textValue(record?.tds_tcs_type, 'NONE'),
    tds_tcs_rate: String(numberValue(record?.tds_tcs_rate)),
    paid_amount: String(numberValue(record?.paid_amount)),
    payment_method: textValue(record?.payment_method, 'Bank transfer'),
    items: items.length > 0 ? items : [lineItem()],
  }
}

function customerName(customer: Row) {
  return textValue(customer.trade_name || customer.legal_name, 'Unnamed business')
}

export function B2BDocumentModal({ kind, mode, record, customers, invoices, token, onUnauthorized, onClose, onSaved, onAction }: Props) {
  const [form, setForm] = useState<FormState>(() => createForm(kind, record))
  const [isWorking, setIsWorking] = useState(false)
  const [error, setError] = useState('')
  const isReadOnly = mode === 'view'
  const label = collectionLabel(kind)

  useEffect(() => {
    setForm(createForm(kind, record))
    setError('')
  }, [kind, record])

  const totals = useMemo(() => {
    const subtotal = form.items.reduce((sum, item) => sum + numberValue(item.quantity) * numberValue(item.rate), 0)
    const discount = subtotal * numberValue(form.discount_percent) / 100
    const taxable = subtotal - discount
    const tax = taxable * 18 / 100
    const transport = kind === 'invoices' ? numberValue(form.transportation_charge) : 0
    const tdsTcs = form.tds_tcs_type === 'NONE' ? 0 : taxable * numberValue(form.tds_tcs_rate) / 100
    const total = taxable + tax + transport + (form.tds_tcs_type === 'TCS' ? tdsTcs : form.tds_tcs_type === 'TDS' ? -tdsTcs : 0)
    return { subtotal, discount, tax, total }
  }, [form.discount_percent, form.items, form.tds_tcs_rate, form.tds_tcs_type, form.transportation_charge, kind])

  const updateField = (field: keyof FormState, value: string) => setForm((current) => ({ ...current, [field]: value }))

  const selectCustomer = (id: string) => {
    const customer = customers.find((item) => String(item.id) === id)
    if (!customer) {
      updateField('customer_id', '')
      return
    }
    setForm((current) => ({
      ...current,
      customer_id: id,
      customer_name: customerName(customer),
      customer_gstin: textValue(customer.gstin, ''),
      customer_email: textValue(customer.email, ''),
      customer_phone: textValue(customer.phone, ''),
      customer_state: textValue(customer.state, ''),
      customer_state_code: textValue(customer.state_code, ''),
      customer_address: textValue(customer.billing_address, ''),
      customer_shipping_address: textValue(customer.shipping_address, textValue(customer.billing_address, '')),
    }))
  }

  const selectInvoice = (id: string) => {
    const invoice = invoices.find((item) => String(item.id) === id)
    if (!invoice) {
      updateField('invoice_id', '')
      return
    }
    const invoiceItems = Array.isArray(invoice.items) ? invoice.items.map((item) => lineItem(item as Row)) : form.items
    setForm((current) => ({
      ...current,
      invoice_id: id,
      customer_id: textValue(invoice.customer_id, current.customer_id),
      customer_name: textValue(invoice.customer_name, current.customer_name),
      customer_gstin: textValue(invoice.customer_gstin, current.customer_gstin),
      customer_email: textValue(invoice.customer_email, current.customer_email),
      customer_phone: textValue(invoice.customer_phone, current.customer_phone),
      customer_state: textValue(invoice.customer_state, current.customer_state),
      customer_state_code: textValue(invoice.customer_state_code, current.customer_state_code),
      customer_address: textValue(invoice.customer_address, current.customer_address),
      items: invoiceItems,
    }))
  }

  const updateItem = (index: number, field: keyof LineItem, value: string) => setForm((current) => ({
    ...current,
    items: current.items.map((item, itemIndex) => itemIndex === index ? { ...item, [field]: value } : item),
  }))

  const save = async (event: FormEvent, issueAfterSave = false) => {
    event.preventDefault()
    setIsWorking(true)
    setError('')
    try {
      const payload: Record<string, unknown> = {
        ...form,
        customer_id: form.customer_id ? Number(form.customer_id) : null,
        invoice_id: form.invoice_id ? Number(form.invoice_id) : null,
        items: form.items.map((item) => ({
          ...item,
          product_id: item.product_id ? Number(item.product_id) : null,
          quantity: numberValue(item.quantity),
          rate: numberValue(item.rate),
          amount: numberValue(item.quantity) * numberValue(item.rate),
        })),
        discount_percent: numberValue(form.discount_percent),
        transportation_charge: numberValue(form.transportation_charge),
        tds_tcs_rate: numberValue(form.tds_tcs_rate),
        paid_amount: numberValue(form.paid_amount),
      }
      delete payload.id
      if (form.id) payload.id = form.id
      if (kind === 'invoices') {
        payload.invoice_date = `${form.invoice_date}T00:00:00Z`
        if (form.due_date) payload.due_date = `${form.due_date}T00:00:00Z`
      } else if (kind === 'proformas') {
        payload.note_date = `${form.note_date}T00:00:00Z`
        if (form.valid_until) payload.valid_until = `${form.valid_until}T00:00:00Z`
      } else {
        payload.note_date = `${form.note_date}T00:00:00Z`
      }
      const path = `/api/b2b/${kind}`
      const response = await apiRequest(token, onUnauthorized, path, { method: form.id ? 'PUT' : 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) })
      const saved = await response.json() as Row
      if (issueAfterSave) {
        const savedId = textValue(saved.id, form.id ? String(form.id) : '')
        await apiRequest(token, onUnauthorized, `/api/b2b/${kind}/issue?id=${savedId}`, { method: 'POST' })
      }
      onSaved()
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : `Unable to save ${kindLabel(kind)}`)
    } finally {
      setIsWorking(false)
    }
  }

  const title = mode === 'create' ? `New ${kindLabel(kind)}` : mode === 'edit' ? `Edit ${kindLabel(kind)}` : `${label} details`
  const status = textValue(record?.status, 'DRAFT').toUpperCase()
  const canIssue = mode === 'view' && status === 'DRAFT'

  return <div className="modal-scrim" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose() }}>
    <section className={`modal-card b2b-document-modal ${isReadOnly ? 'b2b-document-view' : ''}`} role="dialog" aria-modal="true" aria-labelledby="b2b-document-modal-heading">
      <div className="modal-heading"><div><p className="eyebrow">B2B billing / {kindLabel(kind)}</p><h2 id="b2b-document-modal-heading">{title}</h2>{isReadOnly && <span className={`status-pill status-pill-${status.toLowerCase() === 'cancelled' ? 'danger' : status === 'ISSUED' || status === 'PAID' || status === 'ACCEPTED' ? 'success' : 'neutral'}`}>{status}</span>}</div><button className="icon-button" type="button" aria-label="Close" onClick={onClose}><X size={19} aria-hidden="true" /></button></div>
      {error && <div className="dashboard-error" role="alert"><span>{error}</span></div>}
      {isReadOnly ? <>
        <div className="b2b-document-summary"><div><span className="metric-label">Customer</span><strong>{form.customer_name || '—'}</strong><small>{form.customer_gstin || 'GSTIN not supplied'}</small></div><div><span className="metric-label">Document date</span><strong>{formatDate(kind === 'invoices' ? form.invoice_date : form.note_date)}</strong><small>{record?.due_date ? `Due ${formatDate(record.due_date)}` : record?.valid_until ? `Valid until ${formatDate(record.valid_until)}` : 'No secondary date'}</small></div><div><span className="metric-label">Total</span><strong>{formatMoney(record?.total_price || totals.total)}</strong><small>{kind === 'invoices' ? `${textValue(record?.payment_status, 'UNPAID')} · ${formatMoney(record?.balance_amount)}` : 'Tax-inclusive total'}</small></div></div>
        <div className="b2b-document-address"><div><span className="metric-label">Bill to</span><p>{form.customer_address || 'No billing address'}</p></div><div><span className="metric-label">Reference</span><p className="mono-text">{textValue(record?.invoice_number || record?.proforma_number || record?.credit_note_number || record?.debit_note_number, `Draft ${textValue(record?.id)}`)}</p></div></div>
        <div className="b2b-items-table"><div className="b2b-items-table-header"><span>Item</span><span>Qty</span><span>Rate</span><span>Amount</span></div>{form.items.map((item, index) => <div className="b2b-items-table-row" key={`${item.id || 'item'}-${index}`}><span><strong>{item.item_details || 'Unnamed item'}</strong><small>{item.sku || item.hsn_code || 'No SKU / HSN'}</small></span><span>{item.quantity}</span><span>{formatMoney(item.rate)}</span><span>{formatMoney(numberValue(item.quantity) * numberValue(item.rate))}</span></div>)}</div>
        {kind === 'invoices' && status === 'ISSUED' && <div className="b2b-payment-box"><div><span className="metric-label">Record payment</span><p>Update the amount collected without changing the issued document.</p></div><div className="form-grid-two"><label className="form-field"><span>Paid amount</span><input type="number" min="0" step="0.01" value={form.paid_amount} onChange={(event) => updateField('paid_amount', event.target.value)} /></label><label className="form-field"><span>Payment method</span><select value={form.payment_method} onChange={(event) => updateField('payment_method', event.target.value)}><option>Bank transfer</option><option>UPI</option><option>Cash</option><option>Cheque</option><option>Other</option></select></label></div><button className="secondary-button" type="button" disabled={isWorking} onClick={() => onAction('payment', { paid_amount: form.paid_amount, payment_method: form.payment_method })}><Save size={14} aria-hidden="true" /> Save payment</button></div>}
        <div className="modal-actions b2b-document-actions"><div className="table-action-group">{status === 'DRAFT' && <button className="table-link-button" type="button" onClick={() => onAction('edit')}>Edit</button>}{status === 'DRAFT' && <button className="table-link-button danger-link" type="button" onClick={() => onAction('delete')}>Delete</button>}{canIssue && <button className="primary-button" type="button" disabled={isWorking} onClick={() => onAction('issue')}>Issue</button>}{kind === 'invoices' && status === 'ISSUED' && <>{record?.inventory_deducted ? <button className="table-link-button" type="button" onClick={() => onAction('revert-inventory')}>Revert stock</button> : <button className="table-link-button" type="button" onClick={() => onAction('deduct-inventory')}>Deduct stock</button>}<button className="table-link-button danger-link" type="button" onClick={() => onAction('cancel')}>Cancel invoice</button></>}{kind === 'proformas' && status === 'SENT' && <><button className="table-link-button" type="button" onClick={() => onAction('accept')}>Accept</button><button className="table-link-button" type="button" onClick={() => onAction('reject')}>Reject</button></>}{kind === 'proformas' && status === 'ACCEPTED' && <button className="primary-button" type="button" onClick={() => onAction('convert')}>Convert to invoice</button>}{kind === 'proformas' && ['SENT', 'ACCEPTED'].includes(status) && <button className="table-link-button danger-link" type="button" onClick={() => onAction('cancel')}>Cancel</button>}{['credit-notes', 'debit-notes'].includes(kind) && status === 'ISSUED' && <button className="table-link-button danger-link" type="button" onClick={() => onAction('cancel')}>Cancel</button>}</div><button className="secondary-button" type="button" onClick={onClose}>Close</button></div>
      </> : <form onSubmit={(event) => void save(event)}>
        <div className="form-grid-two"><label className="form-field"><span>{kind === 'invoices' ? 'Invoice date' : 'Note date'}</span><input required type="date" value={kind === 'invoices' ? form.invoice_date : form.note_date} onChange={(event) => updateField(kind === 'invoices' ? 'invoice_date' : 'note_date', event.target.value)} /></label>{kind === 'invoices' ? <label className="form-field"><span>Due date</span><input type="date" value={form.due_date} onChange={(event) => updateField('due_date', event.target.value)} /></label> : kind === 'proformas' ? <label className="form-field"><span>Valid until</span><input type="date" value={form.valid_until} onChange={(event) => updateField('valid_until', event.target.value)} /></label> : <label className="form-field"><span>Linked invoice</span><select value={form.invoice_id} onChange={(event) => selectInvoice(event.target.value)}><option value="">Select invoice (optional)</option>{invoices.map((invoice) => <option key={String(invoice.id)} value={String(invoice.id)}>{textValue(invoice.invoice_number, `Invoice ${invoice.id}`)} · {textValue(invoice.customer_name, 'Customer')}</option>)}</select></label>}</div>
        {kind === 'invoices' && <div className="form-grid-two"><label className="form-field"><span>Order reference</span><input value={form.order_number} onChange={(event) => updateField('order_number', event.target.value)} placeholder="Optional order or PO number" /></label><label className="form-field"><span>Payment terms</span><input value={form.terms} onChange={(event) => updateField('terms', event.target.value)} placeholder="e.g. Net 30" /></label></div>}{kind === 'proformas' && <label className="form-field"><span>Customer</span><select required value={form.customer_id} onChange={(event) => selectCustomer(event.target.value)}><option value="">Select business customer</option>{customers.map((customer) => <option key={String(customer.id)} value={String(customer.id)}>{customerName(customer)} · {textValue(customer.gstin)}</option>)}</select></label>}{kind === 'invoices' && <label className="form-field"><span>Customer</span><select required value={form.customer_id} onChange={(event) => selectCustomer(event.target.value)}><option value="">Select business customer</option>{customers.map((customer) => <option key={String(customer.id)} value={String(customer.id)}>{customerName(customer)} · {textValue(customer.gstin)}</option>)}</select></label>}{['credit-notes', 'debit-notes'].includes(kind) && <label className="form-field"><span>Customer</span><select required value={form.customer_id} onChange={(event) => selectCustomer(event.target.value)}><option value="">Select business customer</option>{customers.map((customer) => <option key={String(customer.id)} value={String(customer.id)}>{customerName(customer)} · {textValue(customer.gstin)}</option>)}</select></label>}
        {['credit-notes', 'debit-notes'].includes(kind) && <label className="form-field"><span>Reason</span><input required value={form.reason} onChange={(event) => updateField('reason', event.target.value)} placeholder="Why is this adjustment being raised?" /></label>}
        <div className="b2b-form-section"><div className="b2b-form-section-heading"><div><p className="eyebrow">Line items</p><h3>What is this document for?</h3></div><button className="secondary-button" type="button" onClick={() => setForm((current) => ({ ...current, items: [...current.items, lineItem()] }))}><Plus size={14} aria-hidden="true" /> Add item</button></div>{form.items.map((item, index) => <div className="b2b-edit-item" key={`${item.id || 'new'}-${index}`}><div className="b2b-edit-item-heading"><strong>Item {index + 1}</strong>{form.items.length > 1 && <button className="icon-button" type="button" aria-label={`Remove item ${index + 1}`} onClick={() => setForm((current) => ({ ...current, items: current.items.filter((_, itemIndex) => itemIndex !== index) }))}><Minus size={15} aria-hidden="true" /></button>}</div><label className="form-field"><span>Description</span><input required value={item.item_details} onChange={(event) => updateItem(index, 'item_details', event.target.value)} placeholder="Product or service description" /></label><div className="form-grid-three"><label className="form-field"><span>SKU</span><input value={item.sku} onChange={(event) => updateItem(index, 'sku', event.target.value)} /></label><label className="form-field"><span>Quantity</span><input required type="number" min="0.01" step="0.01" value={item.quantity} onChange={(event) => updateItem(index, 'quantity', event.target.value)} /></label><label className="form-field"><span>Rate</span><input required type="number" min="0" step="0.01" value={item.rate} onChange={(event) => updateItem(index, 'rate', event.target.value)} /></label></div></div>)}</div>
        <div className="form-grid-three"><label className="form-field"><span>Discount %</span><input type="number" min="0" max="100" step="0.01" value={form.discount_percent} onChange={(event) => updateField('discount_percent', event.target.value)} /></label>{kind === 'invoices' && <label className="form-field"><span>Transport charge</span><input type="number" min="0" step="0.01" value={form.transportation_charge} onChange={(event) => updateField('transportation_charge', event.target.value)} /></label>}<label className="form-field"><span>Tax adjustment</span><select value={form.tds_tcs_type} onChange={(event) => updateField('tds_tcs_type', event.target.value)}><option value="NONE">No TDS / TCS</option><option value="TDS">TDS</option><option value="TCS">TCS</option></select></label>{form.tds_tcs_type !== 'NONE' && <label className="form-field"><span>TDS / TCS %</span><input type="number" min="0" step="0.01" value={form.tds_tcs_rate} onChange={(event) => updateField('tds_tcs_rate', event.target.value)} /></label>}</div>
        <div className="b2b-total-preview"><span>Estimated total</span><strong>{formatMoney(totals.total)}</strong><small>Backend recalculates GST at 18% when saved.</small></div>
        <div className="modal-actions"><button className="secondary-button" type="button" onClick={onClose}>Cancel</button><button className="secondary-button" type="submit" disabled={isWorking}><Save size={14} aria-hidden="true" /> {isWorking ? 'Saving…' : 'Save draft'}</button><button className="primary-button" type="button" disabled={isWorking} onClick={(event) => void save(event as unknown as FormEvent, true)}><FilePlus2 size={14} aria-hidden="true" /> Save & issue</button></div>
      </form>}
    </section>
  </div>
}
