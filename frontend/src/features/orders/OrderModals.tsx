import { FormEvent, useCallback, useEffect, useMemo, useState } from 'react'
import { CircleAlert, Download, MessageSquare, Pencil, Plus, RefreshCw, Save, Send, Trash2, X } from 'lucide-react'
import { apiJson, apiRequest, arrayFrom, formatDate, formatMoney, numberValue, textValue } from '../../lib/http'

type Props = { token: string; onUnauthorized: () => void }
type LineItem = { mi_sku: string; title: string; quantity: string; price: string; discount: string }
type Order = Record<string, unknown>
type CustomerRecord = { id: number; first_name?: string; last_name?: string; phone_number?: string; email?: string; address1?: string; address2?: string; city?: string; state?: string; country?: string; zip_code?: string }
type CustomerEditForm = { first_name: string; last_name: string; phone_number: string; email: string; address1: string; address2: string; city: string; state: string; country: string; zip_code: string }

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

function customerFormFromOrder(order: Order): CustomerEditForm {
  const fullName = textValue(order.customer_name, '')
  const [firstName = '', ...lastNameParts] = fullName.split(' ')
  return {
    first_name: textValue(order.customer_first_name, firstName),
    last_name: textValue(order.customer_last_name, lastNameParts.join(' ')),
    phone_number: textValue(order.customer_phone, ''),
    email: textValue(order.customer_email, ''),
    address1: textValue(order.customer_address1, ''),
    address2: textValue(order.customer_address2, ''),
    city: textValue(order.customer_city, ''),
    state: textValue(order.customer_state, ''),
    country: textValue(order.customer_country, ''),
    zip_code: textValue(order.customer_zip, ''),
  }
}

function customerFormFromRecord(customer: CustomerRecord): CustomerEditForm {
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

export function ConvertOrderToB2BModal({ token, onUnauthorized, orderId, onClose, onSuccess }: Props & { orderId: string | number; onClose: () => void; onSuccess: () => void }) {
  const [form, setForm] = useState<ConvertForm>({ orderNumber: '', invoiceDate: dateInput(), paymentDate: dateInput(), paymentStatus: 'UNPAID', paymentMethod: 'Bank transfer', customerId: '', customerName: '', customerGstin: '', customerEmail: '', customerPhone: '', customerState: 'Tamil Nadu', customerStateCode: '33', customerAddress: '', customerShippingAddress: '', discountPercent: '0', transportationCharge: '0', items: [] })
  const [customers, setCustomers] = useState<Record<string, unknown>[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isWorking, setIsWorking] = useState(false)
  const [error, setError] = useState('')

  useEffect((