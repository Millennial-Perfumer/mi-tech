import { FormEvent, useCallback, useEffect, useMemo, useState } from 'react'
import { CircleAlert, Download, MessageSquare, Plus, RefreshCw, Send, Trash2, X } from 'lucide-react'
import { apiJson, apiRequest, arrayFrom, formatDate, formatMoney, numberValue, textValue } from '../../lib/http'

type Props = { token: string; onUnauthorized: () => void }
type LineItem = { mi_sku: string; title: string; quantity: string; price: string; discount: string }
type Order = Record<string, unknown>

const webhookLabels: Record<string, string> = {
  'orders/create': 'Order placed',
  'orders/assigned': 'Order assigned',
  'orders/fulfilled': 'Order dispatched',
  'orders/out_for_delivery': 'Order out for delivery',
  'orders/delivered': 'Order delivered',
  'orders/updated': 'Order updated',
  'orders/cancelled': 'Order cancelled',
  'orders/paid': 'Order paid',
}

const emptyLineItem = (): LineItem => ({ mi_sku: '', title: '', quantity: '1', price: '', discount: '0' })

export function OrderCreateModal({ token, onUnauthorized, onClose, onSuccess }: Props & { onClose: () => void; onSuccess: () => void }) {
  const [form, setForm] = useState({ customer_name: '', customer_phone: '', customer_email: '', customer_address1: '', customer_city: '', customer_state: 'Tamil Nadu', customer_zip: '', financial_status: 'paid', fulfillment_status: 'fulfilled' })
  const [items, setItems] = useState<LineItem[]>([emptyLineItem()])
  const [isWorking, setIsWorking] = useState(false)
  const [error, setError] = useState('')
  const subtotal = items.reduce((sum, item) => sum + Math.max(0, Number(item.quantity) || 0) * (Number(item.price) || 0) - (Number(item.discount) || 0), 0)

  const submit = async (event: FormEvent) => {
    event.preventDefault(); setIsWorking(true); setError('')
    if (!form.customer_name.trim() || !form.customer_phone.trim() || items.some((item) => !item.mi_sku.trim() || !item.title.trim() || Number(item.quantity) <= 0 || Number(item.price) < 0)) { setError('Add customer name, phone, and complete every line item.'); setIsWorking(false); return }
    try { await apiRequest(token, onUnauthorized, '/api/orders', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ terminal_code: 'POS1', ...form, total_price: Number(subtotal.toFixed(2)), total_discount: Number(items.reduce((sum, item) => sum + (Number(item.discount) || 0), 0).toFixed(2)), line_items: items.map((item) => ({ ...item, quantity: Number(item.quantity), price: Number(item.price), discount: Number(item.discount) || 0 })) }) }); onSuccess(); onClose() } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to create order') } finally { setIsWorking(false) }
  }

  return <div className="modal-scrim" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose() }}><form className="modal-card order-create-modal" role="dialog" aria-modal="true" aria-labelledby="create-order-heading" onSubmit={submit}><div className="modal-heading"><div><p className="eyebrow">Orders / POS</p><h2 id="create-order-heading">Create manual order</h2></div><button className="icon-button" type="button" aria-label="Close" onClick={onClose}><X size={19} aria-hidden="true" /></button></div>{error && <div className="dashboard-error" role="alert"><CircleAlert size={16} aria-hidden="true" /><span>{error}</span></div>}<div className="form-grid-two"><label className="form-field"><span>Customer name</span><input required value={form.customer_name} onChange={(event) => setForm({ ...form, customer_name: event.target.value })} /></label><label className="form-field"><span>Phone</span><input required inputMode="tel" value={form.customer_phone} onChange={(event) => setForm({ ...form, customer_phone: event.target.value })} /></label><label className="form-field"><span>Email</span><input type="email" value={form.customer_email} onChange={(event) => setForm({ ...form, customer_email: event.target.value })} /></label><label className="form-field"><span>City</span><input value={form.customer_city} onChange={(event) => setForm({ ...form, customer_city: event.target.value })} /></label><label className="form-field"><span>State</span><input value={form.customer_state} onChange={(event) => setForm({ ...form, customer_state: event.target.value })} /></label><label className="form-field"><span>PIN code</span><input inputMode="numeric" value={form.customer_zip} onChange={(event) => setForm({ ...form, customer_zip: event.target.value })} /></label></div><label className="form-field"><span>Address</span><textarea rows={2} value={form.customer_address1} onChange={(event) => setForm({ ...form, customer_address1: event.target.value })} /></label><div className="order-line-items"><div className="inline-section-heading"><div><p className="eyebrow">Line items</p><strong>{items.length} item{items.length === 1 ? '' : 's'}</strong></div><button className="secondary-button" type="button" onClick={() => setItems([...items, emptyLineItem()])}><Plus size={14} aria-hidden="true" /> Add item</button></div>{items.map((item, index) => <div className="order-line-item" key={index}><div className="form-grid-two"><label className="form-field"><span>MI SKU</span><input required value={item.mi_sku} onChange={(event) => setItems(items.map((current, itemIndex) => itemIndex === index ? { ...current, mi_sku: event.target.value } : current))} /></label><label className="form-field"><span>Product title</span><input required value={item.title} onChange={(event) => setItems(items.map((current, itemIndex) => itemIndex === index ? { ...current, title: event.target.value } : current))} /></label><label className="form-field"><span>Quantity</span><input required type="number" min="1" value={item.quantity} onChange={(event) => setItems(items.map((current, itemIndex) => itemIndex === index ? { ...current, quantity: event.target.value } : current))} /></label><label className="form-field"><span>Unit price</span><input required type="number" min="0" step="0.01" value={item.price} onChange={(event) => setItems(items.map((current, itemIndex) => itemIndex === index ? { ...current, price: event.target.value } : current))} /></label><label className="form-field"><span>Discount</span><input type="number" min="0" step="0.01" value={item.discount} onChange={(event) => setItems(items.map((current, itemIndex) => itemIndex === index ? { ...current, discount: event.target.value } : current))} /></label></div>{items.length > 1 && <button className="text-button line-item-remove" type="button" onClick={() => setItems(items.filter((_, itemIndex) => itemIndex !== index))}><Trash2 size={14} aria-hidden="true" /> Remove item</button>}</div>)}</div><div className="order-create-total"><span>Order total</span><strong>{formatMoney(subtotal)}</strong></div><div className="form-grid-two"><label className="form-field"><span>Payment</span><select value={form.financial_status} onChange={(event) => setForm({ ...form, financial_status: event.target.value })}><option value="paid">Paid</option><option value="pending">Pending</option><option value="partially_paid">Partially paid</option></select></label><label className="form-field"><span>Fulfilment</span><select value={form.fulfillment_status} onChange={(event) => setForm({ ...form, fulfillment_status: event.target.value })}><option value="fulfilled">Fulfilled</option><option value="unfulfilled">Unfulfilled</option></select></label></div><div className="modal-actions"><button className="secondary-button" type="button" onClick={onClose}>Cancel</button><button className="primary-button" type="submit" disabled={isWorking}><Plus size={14} aria-hidden="true" /> {isWorking ? 'Creating…' : 'Create order'}</button></div></form></div>
}

type ConvertForm = {
  orderNumber: string
  invoiceDate: string
  paymentDate: string
  paymentStatus: 'PAID' | 'UNPAID'
  paymentMethod: string
  customerId: string
  customerName: string
  customerGstin: string
  customerEmail: string
  customerPhone: string
  customerState: string
  customerStateCode: string
  customerAddress: string
  customerShippingAddress: string
  discountPercent: string
  transportationCharge: string
  items: Array<{ item_details: string; sku: string; hsn_code: string; quantity: string; rate: string }>
}

const gstStates: Record<string, string> = {
  '01': 'Jammu and Kashmir', '02': 'Himachal Pradesh', '03': 'Punjab', '04': 'Chandigarh', '05': 'Uttarakhand', '06': 'Haryana',
  '07': 'Delhi', '08': 'Rajasthan', '09': 'Uttar Pradesh', '10': 'Bihar', '11': 'Sikkim', '12': 'Arunachal Pradesh',
  '13': 'Nagaland', '14': 'Manipur', '15': 'Mizoram', '16': 'Tripura', '17': 'Meghalaya', '18': 'Assam', '19': 'West Bengal',
  '20': 'Jharkhand', '21': 'Odisha', '22': 'Chhattisgarh', '23': 'Madhya Pradesh', '24': 'Gujarat', '26': 'Dadra and Nagar Haveli and Daman and Diu',
  '27': 'Maharashtra', '29': 'Karnataka', '30': 'Goa', '31': 'Lakshadweep', '32': 'Kerala', '33': 'Tamil Nadu', '34': 'Puducherry',
  '35': 'Andaman and Nicobar Islands', '36': 'Telangana', '37': 'Andhra Pradesh', '38': 'Ladakh',
}

function dateInput(value?: unknown) {
  const date = value ? new Date(String(value)) : new Date()
  return Number.isNaN(date.getTime()) ? new Date().toISOString().slice(0, 10) : date.toISOString().slice(0, 10)
}

function convertOrderItem(value: unknown) {
  const item = value && typeof value === 'object' ? value as Record<string, unknown> : {}
  return {
    item_details: textValue(item.title || item.name, 'Fragrance product'),
    sku: textValue(item.sku || item.mi_sku, ''),
    hsn_code: textValue(item.hsn_code, '33029019'),
    quantity: String(Math.max(1, numberValue(item.quantity) || 1)),
    rate: String(Math.max(0, numberValue(item.price || item.rate))),
  }
}

export function ConvertOrderToB2BModal({ token, onUnauthorized, orderId, onClose, onSuccess }: Props & { orderId: string | number; onClose: () => void; onSuccess: () => void }) {
  const [form, setForm] = useState<ConvertForm>({ orderNumber: '', invoiceDate: dateInput(), paymentDate: dateInput(), paymentStatus: 'UNPAID', paymentMethod: 'Bank transfer', customerId: '', customerName: '', customerGstin: '', customerEmail: '', customerPhone: '', customerState: 'Tamil Nadu', customerStateCode: '33', customerAddress: '', customerShippingAddress: '', discountPercent: '0', transportationCharge: '0', items: [] })
  const [customers, setCustomers] = useState<Record<string, unknown>[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isWorking, setIsWorking] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    let active = true
    const load = async () => {
      setIsLoading(true)
      setError('')
      try {
        const [orderResult, customerResult] = await Promise.allSettled([
          apiJson<unknown>(token, onUnauthorized, `/api/orders?id=${orderId}`),
          apiJson<unknown>(token, onUnauthorized, '/api/b2b/customers'),
        ])
        if (!active) return
        if (orderResult.status === 'rejected') throw orderResult.reason
        const orderData = orderResult.value
        const customerData = customerResult.status === 'fulfilled' ? customerResult.value : []
        const order = orderData && typeof orderData === 'object' && 'order' in orderData ? (orderData as { order: Record<string, unknown> }).order : orderData as Record<string, unknown>
        const sourceItems = order && Array.isArray(order.line_items) ? order.line_items : []
        const customerName = textValue(order?.customer_name, [textValue(order?.customer_first_name, ''), textValue(order?.customer_last_name, '')].filter(Boolean).join(' ') || 'B2B customer')
        const address = [order?.customer_address1, order?.customer_address2, order?.customer_city, order?.customer_state, order?.customer_zip, order?.customer_country].map((value) => textValue(value, '')).filter(Boolean).join(', ')
        const state = textValue(order?.customer_state, 'Tamil Nadu')
        const stateEntry = Object.entries(gstStates).find(([, name]) => name.toLowerCase() === state.toLowerCase() || state.toLowerCase().includes(name.toLowerCase()))
        const items = sourceItems.length > 0 ? sourceItems.map(convertOrderItem) : [{ item_details: `${textValue(order?.source_id, 'Retail')} order ${textValue(order?.order_number, String(orderId))}`, sku: '', hsn_code: '33029019', quantity: '1', rate: String(numberValue(order?.total_price)) }]
        const financialStatus = textValue(order?.financial_status, '').toLowerCase()
        setForm({
          orderNumber: textValue(order?.order_number, String(orderId)),
          invoiceDate: dateInput(order?.created_at),
          paymentDate: dateInput(order?.created_at),
          paymentStatus: financialStatus === 'paid' ? 'PAID' : 'UNPAID',
          paymentMethod: 'Bank transfer',
          customerId: '',
          customerName,
          customerGstin: '',
          customerEmail: textValue(order?.customer_email, ''),
          customerPhone: textValue(order?.customer_phone, ''),
          customerState: state,
          customerStateCode: stateEntry?.[0] || '33',
          customerAddress: address,
          customerShippingAddress: address,
          discountPercent: '0',
          transportationCharge: '0',
          items,
        })
        setCustomers(arrayFrom(customerData, 'customers'))
      } catch (caughtError) {
        if (active) setError(caughtError instanceof Error ? caughtError.message : 'Unable to load order conversion details')
      } finally {
        if (active) setIsLoading(false)
      }
    }
    void load()
    return () => { active = false }
  }, [onUnauthorized, orderId, token])

  const totals = useMemo(() => {
    const subtotal = form.items.reduce((sum, item) => sum + numberValue(item.quantity) * numberValue(item.rate), 0)
    const discount = subtotal * Math.min(100, Math.max(0, numberValue(form.discountPercent))) / 100
    const taxable = Math.max(0, subtotal - discount)
    const tax = taxable * 18 / 100
    const sameState = form.customerStateCode === '33'
    return { subtotal, discount, taxable, cgst: sameState ? tax / 2 : 0, sgst: sameState ? tax / 2 : 0, igst: sameState ? 0 : tax, total: taxable + tax + numberValue(form.transportationCharge) }
  }, [form.customerStateCode, form.discountPercent, form.items, form.transportationCharge])

  const setField = <K extends keyof ConvertForm>(field: K, value: ConvertForm[K]) => setForm((current) => ({ ...current, [field]: value }))
  const updateGstin = (value: string) => {
    const customerGstin = value.toUpperCase().replace(/\s/g, '').slice(0, 15)
    const stateCode = customerGstin.slice(0, 2)
    setForm((current) => ({ ...current, customerGstin, customerStateCode: gstStates[stateCode] ? stateCode : current.customerStateCode, customerState: gstStates[stateCode] || current.customerState }))
  }
  const selectCustomer = (id: string) => {
    const customer = customers.find((item) => String(item.id) === id)
    if (!customer) {
      setField('customerId', '')
      return
    }
    setForm((current) => ({ ...current, customerId: id, customerName: textValue(customer.legal_name || customer.trade_name, current.customerName), customerGstin: textValue(customer.gstin, current.customerGstin), customerEmail: textValue(customer.email, current.customerEmail), customerPhone: textValue(customer.phone, current.customerPhone), customerState: textValue(customer.state, current.customerState), customerStateCode: textValue(customer.state_code, current.customerStateCode), customerAddress: textValue(customer.billing_address, current.customerAddress), customerShippingAddress: textValue(customer.shipping_address || customer.billing_address, current.customerShippingAddress) }))
  }
  const updateItem = (index: number, field: keyof ConvertForm['items'][number], value: string) => setForm((current) => ({ ...current, items: current.items.map((item, itemIndex) => itemIndex === index ? { ...item, [field]: value } : item) }))

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    if (!form.customerName.trim()) { setError('Add the legal or business name before issuing the invoice.'); return }
    if (form.customerGstin.trim().length !== 15) { setError('Enter a valid 15-character customer GSTIN before issuing the invoice.'); return }
    if (form.items.some((item) => !item.item_details.trim() || numberValue(item.quantity) <= 0 || numberValue(item.rate) < 0)) { setError('Complete every line item before issuing the invoice.'); return }
    setIsWorking(true)
    setError('')
    try {
      const response = await apiRequest(token, onUnauthorized, '/api/b2b/invoices', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({
        origin_order_id: String(orderId), order_number: `ORD-${form.orderNumber}`, invoice_date: `${form.invoiceDate}T00:00:00Z`, customer_id: form.customerId ? Number(form.customerId) : null,
        customer_gstin: form.customerGstin.trim(), customer_name: form.customerName.trim(), customer_email: form.customerEmail.trim(), customer_phone: form.customerPhone.trim(), customer_state: form.customerState, customer_state_code: form.customerStateCode,
        customer_address: form.customerAddress, customer_shipping_address: form.customerShippingAddress || form.customerAddress, discount_percent: numberValue(form.discountPercent), transportation_charge: numberValue(form.transportationCharge),
        payment_status: form.paymentStatus, paid_amount: form.paymentStatus === 'PAID' ? totals.total : 0, payment_date: form.paymentStatus === 'PAID' ? `${form.paymentDate}T00:00:00Z` : null, payment_method: form.paymentMethod,
        customer_notes: `Converted from order ${form.orderNumber}.`, items: form.items.map((item) => ({ item_details: item.item_details.trim(), sku: item.sku.trim(), hsn_code: item.hsn_code.trim() || '33029019', quantity: numberValue(item.quantity), rate: numberValue(item.rate), amount: numberValue(item.quantity) * numberValue(item.rate) })),
      }) })
      const saved = await response.json() as Record<string, unknown>
      const savedId = saved.id == null ? '' : String(saved.id)
      if (!savedId) throw new Error('Invoice was created without an identifier, so it could not be issued.')
      await apiRequest(token, onUnauthorized, `/api/b2b/invoices/issue?id=${savedId}`, { method: 'POST' })
      onSuccess()
      onClose()
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to convert order to a B2B invoice')
    } finally {
      setIsWorking(false)
    }
  }

  return <div className="modal-scrim" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget && !isWorking) onClose() }}><form className="modal-card convert-b2b-modal" role="dialog" aria-modal="true" aria-labelledby="convert-b2b-heading" onSubmit={submit}>
    <div className="modal-heading"><div><p className="eyebrow">Orders / B2B billing</p><h2 id="convert-b2b-heading">Convert to B2B invoice</h2><p className="modal-subtitle">Turn this retail order into a GST invoice, with the order details already filled in.</p></div><button className="icon-button" type="button" aria-label="Close conversion dialog" onClick={onClose} disabled={isWorking}><X size={19} aria-hidden="true" /></button></div>
    {error && <div className="dashboard-error" role="alert"><CircleAlert size={16} aria-hidden="true" /><span>{error}</span></div>}
    {isLoading ? <div className="convert-b2b-loading"><span className="loading-spinner" aria-hidden="true" /><p>Loading order and customer details…</p></div> : <>
      <div className="convert-b2b-summary"><div><span className="metric-label">Order</span><strong>{form.orderNumber || `#${orderId}`}</strong><small>Source order</small></div><div><span className="metric-label">Items</span><strong>{form.items.length}</strong><small>Line items to invoice</small></div><div><span className="metric-label">Preview total</span><strong>{formatMoney(totals.total)}</strong><small>GST included</small></div></div>
      <div className="convert-b2b-section"><div className="convert-b2b-section-heading"><div><p className="eyebrow">Customer & GST</p><h3>Who should receive this invoice?</h3></div><span className="form-help">Customer GSTIN is required</span></div>{customers.length > 0 && <label className="form-field"><span>Use an existing B2B customer <small>(optional)</small></span><select value={form.customerId} onChange={(event) => selectCustomer(event.target.value)}><option value="">Enter customer details manually</option>{customers.map((customer) => <option key={String(customer.id)} value={String(customer.id)}>{textValue(customer.legal_name || customer.trade_name, 'Unnamed business')} · {textValue(customer.gstin, 'No GSTIN')}</option>)}</select></label>}<div className="form-grid-two"><label className="form-field"><span>Business / legal name</span><input required value={form.customerName} onChange={(event) => setField('customerName', event.target.value)} /></label><label className="form-field"><span>GSTIN</span><input required maxLength={15} autoCapitalize="characters" value={form.customerGstin} onChange={(event) => updateGstin(event.target.value)} placeholder="33AAAAA0000A1Z5" /></label><label className="form-field"><span>Email</span><input type="email" value={form.customerEmail} onChange={(event) => setField('customerEmail', event.target.value)} /></label><label className="form-field"><span>Phone</span><input inputMode="tel" value={form.customerPhone} onChange={(event) => setField('customerPhone', event.target.value)} /></label><label className="form-field"><span>State</span><input value={form.customerState} onChange={(event) => setField('customerState', event.target.value)} /></label><label className="form-field"><span>GST state code</span><input inputMode="numeric" maxLength={2} value={form.customerStateCode} onChange={(event) => setField('customerStateCode', event.target.value.slice(0, 2))} /></label></div><label className="form-field"><span>Billing address</span><textarea rows={2} value={form.customerAddress} onChange={(event) => setField('customerAddress', event.target.value)} /></label></div>
      <div className="convert-b2b-section"><div className="convert-b2b-section-heading"><div><p className="eyebrow">Invoice setup</p><h3>Review the document</h3></div></div><div className="form-grid-two"><label className="form-field"><span>Invoice date</span><input required type="date" value={form.invoiceDate} onChange={(event) => setField('invoiceDate', event.target.value)} /></label><label className="form-field"><span>Payment status</span><select value={form.paymentStatus} onChange={(event) => setField('paymentStatus', event.target.value as ConvertForm['paymentStatus'])}><option value="PAID">Paid</option><option value="UNPAID">Unpaid</option></select></label><label className="form-field"><span>Payment date <small>(if paid)</small></span><input type="date" value={form.paymentDate} disabled={form.paymentStatus !== 'PAID'} onChange={(event) => setField('paymentDate', event.target.value)} /></label><label className="form-field"><span>Payment method</span><select value={form.paymentMethod} onChange={(event) => setField('paymentMethod', event.target.value)}><option>Bank transfer</option><option>UPI</option><option>Cash</option><option>Cheque</option><option>Other</option></select></label></div></div>
      <div className="convert-b2b-section"><div className="convert-b2b-section-heading"><div><p className="eyebrow">Line items</p><h3>What is being billed?</h3></div></div><div className="convert-b2b-items">{form.items.map((item, index) => <div className="convert-b2b-item" key={`${item.sku}-${index}`}><div className="convert-b2b-item-heading"><strong>Item {index + 1}</strong><span>{formatMoney(numberValue(item.quantity) * numberValue(item.rate))}</span></div><label className="form-field"><span>Description</span><input required value={item.item_details} onChange={(event) => updateItem(index, 'item_details', event.target.value)} /></label><div className="form-grid-three"><label className="form-field"><span>SKU</span><input value={item.sku} onChange={(event) => updateItem(index, 'sku', event.target.value)} /></label><label className="form-field"><span>Quantity</span><input required type="number" min="0.01" step="0.01" value={item.quantity} onChange={(event) => updateItem(index, 'quantity', event.target.value)} /></label><label className="form-field"><span>Rate</span><input required type="number" min="0" step="0.01" value={item.rate} onChange={(event) => updateItem(index, 'rate', event.target.value)} /></label></div></div>)}</div><div className="form-grid-two convert-b2b-charges"><label className="form-field"><span>Discount %</span><input type="number" min="0" max="100" step="0.01" value={form.discountPercent} onChange={(event) => setField('discountPercent', event.target.value)} /></label><label className="form-field"><span>Transport charge</span><input type="number" min="0" step="0.01" value={form.transportationCharge} onChange={(event) => setField('transportationCharge', event.target.value)} /></label></div><div className="convert-b2b-total"><span>Subtotal {formatMoney(totals.subtotal)} · Discount {formatMoney(totals.discount)}</span><strong>Total {formatMoney(totals.total)}</strong><small>{form.customerStateCode === '33' ? `CGST ${formatMoney(totals.cgst)} + SGST ${formatMoney(totals.sgst)}` : `IGST ${formatMoney(totals.igst)}`}</small></div></div>
      <div className="modal-actions"><button className="secondary-button" type="button" onClick={onClose} disabled={isWorking}>Cancel</button><button className="primary-button" type="submit" disabled={isWorking}>{isWorking ? 'Creating & issuing…' : 'Create & issue invoice'}</button></div>
    </>}
  </form></div>
}

export function ManualWhatsAppModal({ token, onUnauthorized, orderId, orderNumber, customerName, customerPhone, onClose, onSent }: Props & { orderId: string | number; orderNumber: string; customerName: string; customerPhone: string; onClose: () => void; onSent?: () => void }) {
  const [triggers, setTriggers] = useState<Record<string, unknown>[]>([])
  const [selectedTemplateId, setSelectedTemplateId] = useState<number | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isSending, setIsSending] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    let active = true
    const loadTriggers = async () => {
      setIsLoading(true)
      setError('')
      try {
        const data = await apiJson<unknown>(token, onUnauthorized, '/api/automation/whatsapp/triggers')
        const available = arrayFrom(data).filter((trigger) => numberValue(trigger.template_id) > 0 && textValue(trigger.template_status, '').toUpperCase() !== 'ARCHIVED')
        if (active) setTriggers(available)
      } catch (caughtError) {
        if (active) setError(caughtError instanceof Error ? caughtError.message : 'Unable to load message templates')
      } finally {
        if (active) setIsLoading(false)
      }
    }
    void loadTriggers()
    return () => { active = false }
  }, [onUnauthorized, token])

  const sendMessage = async () => {
    if (!selectedTemplateId || !customerPhone) return
    setIsSending(true)
    setError('')
    try {
      await apiRequest(token, onUnauthorized, '/api/automation/whatsapp/send-manual', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ order_id: String(orderId), template_id: selectedTemplateId }),
      })
      onSent?.()
      onClose()
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to send WhatsApp message')
    } finally {
      setIsSending(false)
    }
  }

  return <div className="modal-scrim" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget && !isSending) onClose() }}><section className="modal-card manual-message-modal" role="dialog" aria-modal="true" aria-labelledby="manual-message-heading">
    <div className="modal-heading"><div><p className="eyebrow">Orders / WhatsApp</p><h2 id="manual-message-heading">Send message</h2></div><button className="icon-button" type="button" aria-label="Close message dialog" onClick={onClose} disabled={isSending}><X size={19} aria-hidden="true" /></button></div>
    <div className="manual-message-recipient"><span className="metric-label">Sending to</span><strong>{customerName || 'Guest customer'}</strong><span>Order {orderNumber || `#${orderId}`} · {customerPhone || 'No phone number'}</span></div>
    {error && <div className="dashboard-error" role="alert"><CircleAlert size={16} aria-hidden="true" /><span>{error}</span></div>}
    <div className="manual-message-heading"><div><span className="metric-label">Message template</span><p>Select a configured WhatsApp event to send for this order.</p></div></div>
    <div className="manual-message-options" role="listbox" aria-label="Select WhatsApp message template" aria-busy={isLoading}>
      {isLoading ? <div className="manual-message-empty"><RefreshCw className="spin" size={17} aria-hidden="true" /> Loading templates…</div> : triggers.length === 0 ? <div className="manual-message-empty">No active WhatsApp templates are configured.</div> : triggers.map((trigger, index) => {
        const templateId = numberValue(trigger.template_id)
        const topic = textValue(trigger.webhook_topic, `Template ${index + 1}`)
        return <button className={`manual-message-option ${selectedTemplateId === templateId ? 'manual-message-option-selected' : ''}`} key={`${templateId}-${topic}`} type="button" role="option" aria-selected={selectedTemplateId === templateId} onClick={() => setSelectedTemplateId(templateId)}><span><strong>{webhookLabels[topic] || topic}</strong><small>{textValue(trigger.template_body, 'Configured WhatsApp message')}</small></span><em>{textValue(trigger.template_name, '')}</em></button>
      })}
    </div>
    <div className="modal-actions"><button className="secondary-button" type="button" onClick={onClose} disabled={isSending}>Cancel</button><button className="primary-button" type="button" onClick={() => void sendMessage()} disabled={isSending || !selectedTemplateId || !customerPhone}><Send size={14} aria-hidden="true" /> {isSending ? 'Sending…' : 'Send message'}</button></div>
  </section></div>
}

export function OrderDetailsModal({ token, onUnauthorized, orderId, onClose, onChanged, onConvertToB2B }: Props & { orderId: string | number; onClose: () => void; onChanged: () => void; onConvertToB2B?: () => void }) {
  const [order, setOrder] = useState<Order | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isWorking, setIsWorking] = useState(false)
  const [error, setError] = useState('')
  const [isMessageOpen, setIsMessageOpen] = useState(false)
  const load = useCallback(async () => { setIsLoading(true); try { const data = await apiJson<unknown>(token, onUnauthorized, `/api/orders?id=${orderId}`); setOrder((data && typeof data === 'object' && 'order' in data ? (data as { order: Order }).order : data) as Order) } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to load order details') } finally { setIsLoading(false) } }, [onUnauthorized, orderId, token])
  useEffect(() => { void load() }, [load])
  const update = async (path: string, body?: unknown, label = 'Order updated') => { setIsWorking(true); try { await apiRequest(token, onUnauthorized, path, body ? { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) } : { method: 'PUT' }); setOrder((current) => current ? { ...current, ...(body as Order || {}) } : current); onChanged() } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : `${label} failed`) } finally { setIsWorking(false) } }
  const downloadInvoice = async () => { setIsWorking(true); try { const response = await apiRequest(token, onUnauthorized, `/api/orders/invoice?id=${orderId}`); const blob = await response.blob(); const link = document.createElement('a'); link.href = URL.createObjectURL(blob); link.download = `order-${orderId}-invoice.pdf`; link.click(); URL.revokeObjectURL(link.href) } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to download invoice') } finally { setIsWorking(false) } }
  const isB2BOrder = textValue(order?.source_id, '').toLowerCase() === 'b2b'
  const customerPhone = textValue(order?.customer_phone, '')
  return <><div className="modal-scrim" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose() }}><div className="modal-card order-details-modal" role="dialog" aria-modal="true" aria-labelledby="order-details-heading"><div className="modal-heading"><div><p className="eyebrow">Order detail</p><h2 id="order-details-heading">{textValue(order?.order_number, `Order #${orderId}`)}</h2></div><button className="icon-button" type="button" aria-label="Close" onClick={onClose}><X size={19} aria-hidden="true" /></button></div>{error && <div className="dashboard-error" role="alert"><CircleAlert size={16} aria-hidden="true" /><span>{error}</span><button type="button" onClick={() => void load()}>Try again</button></div>}{isLoading ? <p className="table-state">Loading order details…</p> : order && <><div className="order-detail-grid"><div><span className="metric-label">Customer</span><strong>{textValue(order.customer_name, 'Guest customer')}</strong><small>{customerPhone}</small><small>{textValue(order.customer_email, '')}</small></div><div><span className="metric-label">Placed</span><strong>{formatDate(order.created_at)}</strong><small>{textValue(order.source_id, 'Unknown channel')}</small></div><div><span className="metric-label">Total</span><strong>{formatMoney(order.total_price)}</strong><small>{textValue(order.financial_status, 'Payment pending')}</small></div><div><span className="metric-label">Fulfilment</span><strong>{textValue(order.fulfillment_status || order.status, 'Unfulfilled')}</strong><small>{textValue(order.delivery_status, '')}</small></div></div><div className="order-detail-actions">{customerPhone && <button className="primary-button" type="button" onClick={() => setIsMessageOpen(true)} disabled={isWorking}><MessageSquare size={14} aria-hidden="true" /> Send message</button>}<button className="secondary-button" type="button" onClick={() => void downloadInvoice()} disabled={isWorking}><Download size={14} aria-hidden="true" /> Invoice</button><button className="secondary-button" type="button" disabled={isWorking} onClick={() => void update(`/api/orders/payment-status?id=${orderId}`, { financial_status: 'paid' }, 'Payment update')}>Mark paid</button><button className="secondary-button" type="button" disabled={isWorking} onClick={() => void update(`/api/orders/delivered?id=${orderId}`, undefined, 'Delivery update')}>Mark delivered</button>{onConvertToB2B && !isB2BOrder && <button className="primary-button" type="button" disabled={isWorking} onClick={onConvertToB2B}>Convert to B2B</button>}</div><div className="order-status-editor"><label className="form-field"><span>Order status</span><select value={textValue(order.status, 'open')} onChange={(event) => void update(`/api/orders/status?id=${orderId}`, { status: event.target.value }, 'Status update')} disabled={isWorking}><option value="open">Open</option><option value="processing">Processing</option><option value="completed">Completed</option><option value="cancelled">Cancelled</option></select></label></div><div className="order-detail-section"><p className="eyebrow">Customer address</p><p>{[order.customer_address1, order.customer_city, order.customer_state, order.customer_zip].map((value) => textValue(value, '')).filter(Boolean).join(', ') || 'No address recorded.'}</p></div></>}</div></div>{isMessageOpen && order && <ManualWhatsAppModal token={token} onUnauthorized={onUnauthorized} orderId={orderId} orderNumber={textValue(order.order_number, `#${orderId}`)} customerName={textValue(order.customer_name, 'Guest customer')} customerPhone={customerPhone} onClose={() => setIsMessageOpen(false)} onSent={onChanged} />}</>
}
