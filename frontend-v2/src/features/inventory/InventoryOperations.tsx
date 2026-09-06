import { FormEvent, useCallback, useEffect, useMemo, useState } from 'react'
import { CircleAlert, Edit3, FlaskConical, PackagePlus, Plus, RefreshCw, Trash2, Truck, X } from 'lucide-react'
import { apiJson, apiRequest, arrayFrom, formatDate, formatMoney, numberValue, textValue } from '../../lib/http'

export type InventorySection = 'oils' | 'suppliers' | 'purchase-orders' | 'manufacturing'
type Props = { token: string; onUnauthorized: () => void; section: InventorySection }
type Row = Record<string, unknown>

type ManufacturingOilForm = {
  key: string
  oil_inventory_id: string
  quantity_grams: string
  deduct_inventory: boolean
}

type ManufacturingProductForm = {
  key: string
  inventory_item_id: string
  quantity_produced: string
  add_stock: boolean
}

type ManufacturingForm = {
  manufacturing_date: string
  notes: string
  oils: ManufacturingOilForm[]
  products: ManufacturingProductForm[]
}

const emptyOil = { name: '', inventory_item_id: '', supplier_id: '', purchase_price_per_kg: '', grams_left: '' }
const emptySupplier = { name: '', contact_info: '' }
const emptyPO = { supplier_id: '', oil_inventory_id: '', quantity_grams: '', unit_price_per_kg: '', purchase_date: new Date().toISOString().slice(0, 10) }
let rowSequence = 0

function nextRowKey() {
  rowSequence += 1
  return `manufacturing-row-${rowSequence}`
}

function newManufacturingOil(): ManufacturingOilForm {
  return { key: nextRowKey(), oil_inventory_id: '', quantity_grams: '', deduct_inventory: true }
}

function newManufacturingProduct(): ManufacturingProductForm {
  return { key: nextRowKey(), inventory_item_id: '', quantity_produced: '', add_stock: true }
}

function emptyManufacturing(): ManufacturingForm {
  return {
    manufacturing_date: new Date().toISOString().slice(0, 10),
    notes: '',
    oils: [newManufacturingOil()],
    products: [newManufacturingProduct()],
  }
}

function nestedName(value: unknown) {
  return value && typeof value === 'object' ? textValue((value as Row).name || (value as Row).title) : '—'
}

function inputValue(value: unknown, fallback = '') {
  if (value === null || value === undefined) return fallback
  return String(value)
}

function rowList(value: unknown) {
  return Array.isArray(value) ? value as Row[] : []
}

export function InventoryOperations({ token, onUnauthorized, section }: Props) {
  const [rows, setRows] = useState<Row[]>([])
  const [oils, setOils] = useState<Row[]>([])
  const [suppliers, setSuppliers] = useState<Row[]>([])
  const [products, setProducts] = useState<Row[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isWorking, setIsWorking] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [modal, setModal] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [oilForm, setOilForm] = useState(emptyOil)
  const [supplierForm, setSupplierForm] = useState(emptySupplier)
  const [poForm, setPOForm] = useState(emptyPO)
  const [manufacturingForm, setManufacturingForm] = useState<ManufacturingForm>(emptyManufacturing)

  const endpoint = section === 'oils' ? '/api/inventory/oil?limit=100' : section === 'suppliers' ? '/api/inventory/suppliers' : section === 'purchase-orders' ? '/api/inventory/po?limit=100' : '/api/inventory/manufacturing?limit=100'

  const load = useCallback(async () => {
    setIsLoading(true)
    setError('')
    try {
      const requests: Promise<unknown>[] = [apiJson<unknown>(token, onUnauthorized, endpoint)]
      if (section !== 'suppliers') requests.push(apiJson<unknown>(token, onUnauthorized, '/api/inventory/oil?limit=100'))
      if (section !== 'suppliers') requests.push(apiJson<unknown>(token, onUnauthorized, '/api/inventory/suppliers'))
      if (section === 'oils' || section === 'manufacturing') requests.push(apiJson<unknown>(token, onUnauthorized, '/api/inventory?limit=100'))
      const results = await Promise.all(requests)
      setRows(arrayFrom(results[0], section === 'oils' ? 'items' : section === 'purchase-orders' || section === 'manufacturing' ? 'items' : undefined))
      let resultIndex = 1
      if (section !== 'suppliers') {
        setOils(arrayFrom(results[resultIndex], 'items'))
        resultIndex += 1
      }
      if (section !== 'suppliers') {
        setSuppliers(arrayFrom(results[resultIndex], 'suppliers'))
        resultIndex += 1
      }
      if (section === 'oils' || section === 'manufacturing') setProducts(arrayFrom(results[resultIndex], 'items'))
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to load inventory operations')
    } finally {
      setIsLoading(false)
    }
  }, [endpoint, onUnauthorized, section, token])

  useEffect(() => {
    void load()
  }, [load])

  const title = section === 'oils' ? 'Oil inventory' : section === 'suppliers' ? 'Suppliers' : section === 'purchase-orders' ? 'Purchase orders' : 'Manufacturing'
  const icon = section === 'oils' ? <FlaskConical size={18} aria-hidden="true" /> : section === 'suppliers' ? <Truck size={18} aria-hidden="true" /> : <PackagePlus size={18} aria-hidden="true" />
  const summary = useMemo(() => section === 'oils'
    ? { primary: rows.length, primaryLabel: 'Oil records', secondary: rows.reduce((sum, row) => sum + numberValue(row.grams_left), 0), secondaryLabel: 'Grams available' }
    : section === 'suppliers'
      ? { primary: rows.length, primaryLabel: 'Suppliers', secondary: rows.reduce((sum, row) => sum + numberValue(row.oils_count), 0), secondaryLabel: 'Oil relationships' }
      : section === 'purchase-orders'
        ? { primary: rows.length, primaryLabel: 'Purchase records', secondary: rows.reduce((sum, row) => sum + numberValue(row.total_price), 0), secondaryLabel: 'Recorded value' }
        : { primary: rows.length, primaryLabel: 'Production runs', secondary: rows.reduce((sum, row) => sum + (Array.isArray(row.products) ? row.products.reduce((count, product) => count + numberValue((product as Row).quantity_produced), 0) : 0), 0), secondaryLabel: 'Units produced' }, [rows, section])

  const openCreate = () => {
    setEditingId(null)
    setOilForm(emptyOil)
    setSupplierForm(emptySupplier)
    setPOForm({ ...emptyPO, purchase_date: new Date().toISOString().slice(0, 10) })
    setManufacturingForm(emptyManufacturing())
    setError('')
    setModal(true)
  }

  const openEdit = (row: Row) => {
    setEditingId(inputValue(row.id))
    setError('')
    if (section === 'oils') {
      setOilForm({
        name: inputValue(row.name),
        inventory_item_id: inputValue(row.inventory_item_id),
        supplier_id: inputValue(row.supplier_id),
        purchase_price_per_kg: inputValue(row.purchase_price_per_kg),
        grams_left: inputValue(row.grams_left),
      })
    }
    if (section === 'suppliers') setSupplierForm({ name: inputValue(row.name), contact_info: inputValue(row.contact_info) })
    if (section === 'purchase-orders') {
      setPOForm({
        supplier_id: inputValue(row.supplier_id),
        oil_inventory_id: inputValue(row.oil_inventory_id),
        quantity_grams: inputValue(row.quantity_grams),
        unit_price_per_kg: inputValue(row.unit_price_per_kg),
        purchase_date: inputValue(row.purchase_date).slice(0, 10),
      })
    }
    if (section === 'manufacturing') {
      const oilRows = rowList(row.oils).map((oil) => ({
        key: nextRowKey(),
        oil_inventory_id: inputValue(oil.oil_inventory_id),
        quantity_grams: inputValue(oil.quantity_grams),
        deduct_inventory: oil.deduct_inventory !== false,
      }))
      const productRows = rowList(row.products).map((product) => ({
        key: nextRowKey(),
        inventory_item_id: inputValue(product.inventory_item_id),
        quantity_produced: inputValue(product.quantity_produced),
        add_stock: product.add_stock !== false,
      }))
      setManufacturingForm({
        manufacturing_date: inputValue(row.manufacturing_date).slice(0, 10),
        notes: inputValue(row.notes),
        oils: oilRows.length ? oilRows : [newManufacturingOil()],
        products: productRows.length ? productRows : [newManufacturingProduct()],
      })
    }
    setModal(true)
  }

  const updateManufacturingOil = (key: string, patch: Partial<Omit<ManufacturingOilForm, 'key'>>) => {
    setManufacturingForm((current) => ({ ...current, oils: current.oils.map((oil) => oil.key === key ? { ...oil, ...patch } : oil) }))
  }

  const updateManufacturingProduct = (key: string, patch: Partial<Omit<ManufacturingProductForm, 'key'>>) => {
    setManufacturingForm((current) => ({ ...current, products: current.products.map((product) => product.key === key ? { ...product, ...patch } : product) }))
  }

  const addManufacturingOil = () => setManufacturingForm((current) => ({ ...current, oils: [...current.oils, newManufacturingOil()] }))
  const addManufacturingProduct = () => setManufacturingForm((current) => ({ ...current, products: [...current.products, newManufacturingProduct()] }))
  const removeManufacturingOil = (key: string) => setManufacturingForm((current) => ({ ...current, oils: current.oils.filter((oil) => oil.key !== key) }))
  const removeManufacturingProduct = (key: string) => setManufacturingForm((current) => ({ ...current, products: current.products.filter((product) => product.key !== key) }))

  const save = async (event: FormEvent) => {
    event.preventDefault()
    setIsWorking(true)
    setError('')
    try {
      let path = ''
      const method = editingId ? 'PUT' : 'POST'
      let body: unknown
      if (section === 'oils') {
        path = '/api/inventory/oil'
        body = { ...oilForm, id: editingId ? Number(editingId) : undefined, inventory_item_id: oilForm.inventory_item_id ? Number(oilForm.inventory_item_id) : null, supplier_id: oilForm.supplier_id ? Number(oilForm.supplier_id) : null, purchase_price_per_kg: oilForm.purchase_price_per_kg ? Number(oilForm.purchase_price_per_kg) : null, grams_left: oilForm.grams_left ? Number(oilForm.grams_left) : null }
      } else if (section === 'suppliers') {
        path = '/api/inventory/suppliers'
        body = { ...supplierForm, id: editingId ? Number(editingId) : undefined }
      } else if (section === 'purchase-orders') {
        path = '/api/inventory/po'
        body = { id: editingId ? Number(editingId) : undefined, supplier_id: Number(poForm.supplier_id), oil_inventory_id: Number(poForm.oil_inventory_id), quantity_grams: Number(poForm.quantity_grams), unit_price_per_kg: Number(poForm.unit_price_per_kg), total_price: Number(((Number(poForm.quantity_grams) / 1000) * Number(poForm.unit_price_per_kg)).toFixed(2)), purchase_date: `${poForm.purchase_date}T00:00:00Z` }
      } else {
        const incompleteOil = manufacturingForm.oils.some((oil) => (oil.oil_inventory_id || oil.quantity_grams) && (!oil.oil_inventory_id || !oil.quantity_grams || Number(oil.quantity_grams) <= 0))
        const incompleteProduct = manufacturingForm.products.some((product) => (product.inventory_item_id || product.quantity_produced) && (!product.inventory_item_id || !product.quantity_produced || Number(product.quantity_produced) <= 0))
        const savedOils = manufacturingForm.oils.filter((oil) => oil.oil_inventory_id && Number(oil.quantity_grams) > 0)
        const savedProducts = manufacturingForm.products.filter((product) => product.inventory_item_id && Number(product.quantity_produced) > 0)
        if (incompleteOil || incompleteProduct) throw new Error('Complete or remove each partially filled material or product row.')
        if (!savedOils.length && !savedProducts.length) throw new Error('Add at least one oil or product row before saving.')
        path = '/api/inventory/manufacturing'
        body = {
          id: editingId ? Number(editingId) : undefined,
          notes: manufacturingForm.notes,
          manufacturing_date: `${manufacturingForm.manufacturing_date}T00:00:00Z`,
          oils: savedOils.map((oil) => ({ oil_inventory_id: Number(oil.oil_inventory_id), quantity_grams: Number(oil.quantity_grams), deduct_inventory: oil.deduct_inventory })),
          products: savedProducts.map((product) => ({ inventory_item_id: Number(product.inventory_item_id), quantity_produced: Number(product.quantity_produced), add_stock: product.add_stock })),
        }
      }
      await apiRequest(token, onUnauthorized, path, { method, headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
      setNotice(`${title} ${editingId ? 'updated' : 'created'}`)
      setModal(false)
      await load()
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : `Unable to save ${title.toLowerCase()}`)
    } finally {
      setIsWorking(false)
    }
  }

  const remove = async (row: Row) => {
    if (!window.confirm(`Delete this ${title.toLowerCase().replace(/s$/, '')}?`)) return
    const id = inputValue(row.id)
    setIsWorking(true)
    try {
      await apiRequest(token, onUnauthorized, `${section === 'oils' ? '/api/inventory/oil' : section === 'suppliers' ? '/api/inventory/suppliers' : section === 'purchase-orders' ? '/api/inventory/po' : '/api/inventory/manufacturing'}?id=${id}`, { method: 'DELETE' })
      setNotice('Record deleted')
      await load()
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to delete record')
    } finally {
      setIsWorking(false)
    }
  }

  const manufacturingFormContent = section === 'manufacturing' ? (
    <>
      <div className="manufacturing-record-fields form-grid-two">
        <label className="form-field"><span>Production date</span><input required type="date" value={manufacturingForm.manufacturing_date} onChange={(event) => setManufacturingForm((current) => ({ ...current, manufacturing_date: event.target.value }))} /></label>
        <label className="form-field"><span>Notes</span><input value={manufacturingForm.notes} onChange={(event) => setManufacturingForm((current) => ({ ...current, notes: event.target.value }))} placeholder="Optional batch notes" /></label>
      </div>
      <section className="manufacturing-form-section" aria-labelledby="manufacturing-oils-heading">
        <div className="manufacturing-section-heading"><div><p className="eyebrow">Raw materials</p><h3 id="manufacturing-oils-heading">Oils used</h3><p>Select each oil and quantity used in this batch.</p></div><button className="secondary-button" type="button" onClick={addManufacturingOil}><Plus size={14} aria-hidden="true" /> Add oil</button></div>
        <div className="manufacturing-row-list">{manufacturingForm.oils.length === 0 ? <div className="manufacturing-empty-row">No oils added. Use “Add oil” to add a material.</div> : manufacturingForm.oils.map((oil, index) => <article className="manufacturing-row-card" key={oil.key}><div className="manufacturing-row-heading"><span>Oil {index + 1}</span><button className="icon-button" type="button" aria-label={`Remove oil ${index + 1}`} onClick={() => removeManufacturingOil(oil.key)}><X size={16} aria-hidden="true" /></button></div><div className="manufacturing-row-fields form-grid-two"><label className="form-field"><span>Fragrance oil</span><select value={oil.oil_inventory_id} onChange={(event) => updateManufacturingOil(oil.key, { oil_inventory_id: event.target.value })}><option value="">Select oil</option>{oils.map((option) => <option key={String(option.id)} value={String(option.id)}>{textValue(option.name, 'Unnamed oil')}</option>)}</select></label><label className="form-field"><span>Grams used</span><input type="number" min="0" step="0.01" inputMode="decimal" value={oil.quantity_grams} onChange={(event) => updateManufacturingOil(oil.key, { quantity_grams: event.target.value })} placeholder="0" /></label></div><label className="toggle-control"><input type="checkbox" checked={oil.deduct_inventory} onChange={(event) => updateManufacturingOil(oil.key, { deduct_inventory: event.target.checked })} /> Deduct oil inventory</label></article>)}</div>
      </section>
      <section className="manufacturing-form-section" aria-labelledby="manufacturing-products-heading">
        <div className="manufacturing-section-heading"><div><p className="eyebrow">Finished goods</p><h3 id="manufacturing-products-heading">Products made</h3><p>Add each product and quantity produced in this batch.</p></div><button className="secondary-button" type="button" onClick={addManufacturingProduct}><Plus size={14} aria-hidden="true" /> Add product</button></div>
        <div className="manufacturing-row-list">{manufacturingForm.products.length === 0 ? <div className="manufacturing-empty-row">No products added. Use “Add product” to add a finished good.</div> : manufacturingForm.products.map((product, index) => <article className="manufacturing-row-card" key={product.key}><div className="manufacturing-row-heading"><span>Product {index + 1}</span><button className="icon-button" type="button" aria-label={`Remove product ${index + 1}`} onClick={() => removeManufacturingProduct(product.key)}><X size={16} aria-hidden="true" /></button></div><div className="manufacturing-row-fields form-grid-two"><label className="form-field"><span>Product</span><select value={product.inventory_item_id} onChange={(event) => updateManufacturingProduct(product.key, { inventory_item_id: event.target.value })}><option value="">Select product</option>{products.map((option) => <option key={String(option.id)} value={String(option.id)}>{textValue(option.mi_sku)} · {textValue(option.title)}</option>)}</select></label><label className="form-field"><span>Units produced</span><input type="number" min="0" step="1" inputMode="numeric" value={product.quantity_produced} onChange={(event) => updateManufacturingProduct(product.key, { quantity_produced: event.target.value })} placeholder="0" /></label></div><label className="toggle-control"><input type="checkbox" checked={product.add_stock} onChange={(event) => updateManufacturingProduct(product.key, { add_stock: event.target.checked })} /> Add to product stock</label></article>)}</div>
      </section>
    </>
  ) : null

  return <section className="inventory-operations"><div className="inventory-operation-actions"><button className="secondary-button" type="button" onClick={() => void load()} disabled={isLoading}><RefreshCw size={15} className={isLoading ? 'spin' : undefined} aria-hidden="true" /> Refresh</button><button className="primary-button" type="button" onClick={openCreate}><Plus size={15} aria-hidden="true" /> Add record</button></div>{error && !modal && <div className="dashboard-error" role="alert"><CircleAlert size={18} aria-hidden="true" /><span>{error}</span><button type="button" onClick={() => void load()}>Try again</button></div>}{notice && <div className="inventory-notice" role="status">{notice}</div>}<div className="inventory-summary-grid"><div className="inventory-summary-card"><span className="metric-label">{summary.primaryLabel}</span><strong>{summary.primary.toLocaleString('en-IN')}</strong><small>Visible records</small></div><div className="inventory-summary-card"><span className="metric-label">{summary.secondaryLabel}</span><strong>{section === 'purchase-orders' ? formatMoney(summary.secondary) : Number(summary.secondary).toLocaleString('en-IN')}</strong><small>Across this workspace</small></div></div><section className="orders-card"><div className="orders-card-heading"><div><p className="eyebrow">{title}</p><h3>{rows.length} records</h3></div><span className="inventory-operation-icon">{icon}</span></div><div className="orders-table-wrap"><table className={`orders-table ${section === 'oils' ? 'oil-inventory-table' : ''}`}><caption className="sr-only">{title}</caption><thead>{section === 'oils' ? <tr><th>Oil</th><th>Supplier</th><th>Stock</th><th>Purchase price</th><th>Actions</th></tr> : section === 'suppliers' ? <tr><th>Supplier</th><th>Contact</th><th>Oils</th><th>Added</th><th>Actions</th></tr> : section === 'purchase-orders' ? <tr><th>Purchase date</th><th>Supplier</th><th>Oil</th><th>Quantity</th><th>Total</th><th>Actions</th></tr> : <tr><th>Manufactured</th><th>Oils used</th><th>Products made</th><th>Notes</th><th>Actions</th></tr>}</thead><tbody>{isLoading ? <tr><td colSpan={6} className="table-state">Loading {title.toLowerCase()}…</td></tr> : rows.length === 0 ? <tr><td colSpan={6} className="table-state">No {title.toLowerCase()} records found.</td></tr> : rows.map((row, index) => { const id = inputValue(row.id, String(index)); const action = <div className="table-action-group"><button className="table-link-button" type="button" onClick={() => openEdit(row)}><Edit3 size={13} aria-hidden="true" /> Edit</button><button className="table-link-button" type="button" onClick={() => void remove(row)} disabled={isWorking}><Trash2 size={13} aria-hidden="true" /> Delete</button></div>; if (section === 'oils') { const oilName = textValue(row.name, 'Unnamed oil'); const linkedProduct = nestedName(row.inventory_item); return <tr key={id}><td className="oil-name-cell"><strong title={oilName}>{oilName}</strong><small className="table-subtext" title={linkedProduct}>{linkedProduct}</small></td><td>{nestedName(row.supplier)}</td><td className="table-money">{numberValue(row.grams_left).toLocaleString('en-IN')} g</td><td className="table-money">{formatMoney(row.purchase_price_per_kg)}/kg</td><td>{action}</td></tr> } if (section === 'suppliers') return <tr key={id}><td><strong>{textValue(row.name, 'Unnamed supplier')}</strong></td><td>{textValue(row.contact_info)}</td><td>{numberValue(row.oils_count)}</td><td>{formatDate(row.created_at)}</td><td>{action}</td></tr>; if (section === 'purchase-orders') return <tr key={id}><td>{formatDate(row.purchase_date)}</td><td>{nestedName(row.supplier)}</td><td>{nestedName(row.oil_inventory)}</td><td>{numberValue(row.quantity_grams).toLocaleString('en-IN')} g</td><td className="table-money">{formatMoney(row.total_price)}</td><td>{action}</td></tr>; return <tr key={id}><td>{formatDate(row.manufacturing_date)}</td><td>{Array.isArray(row.oils) ? row.oils.map((oil) => `${numberValue((oil as Row).quantity_grams)}g`).join(', ') || '—' : '—'}</td><td>{Array.isArray(row.products) ? row.products.map((product) => `${numberValue((product as Row).quantity_produced)}×`).join(', ') || '—' : '—'}</td><td>{textValue(row.notes)}</td><td>{action}</td></tr> })}</tbody></table></div></section>{modal && <div className="modal-scrim" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) setModal(false) }}><form className={`modal-card ${section === 'manufacturing' ? 'manufacturing-modal' : ''}`} role="dialog" aria-modal="true" aria-labelledby="inventory-operation-modal" onSubmit={save}><div className="modal-heading"><div><p className="eyebrow">Inventory / {title}</p><h2 id="inventory-operation-modal">{editingId ? 'Edit' : 'Add'} {title.toLowerCase()}</h2>{section === 'manufacturing' && <p className="modal-subtitle">Update the date, materials, and finished products in one record.</p>}</div><button className="icon-button" type="button" aria-label="Close" onClick={() => setModal(false)}><X size={19} aria-hidden="true" /></button></div>{section === 'oils' ? <><label className="form-field"><span>Oil name</span><input required value={oilForm.name} onChange={(event) => setOilForm({ ...oilForm, name: event.target.value })} /></label><div className="form-grid-two"><label className="form-field"><span>Product link</span><select value={oilForm.inventory_item_id} onChange={(event) => setOilForm({ ...oilForm, inventory_item_id: event.target.value })}><option value="">Unlinked</option>{products.map((product) => <option key={String(product.id)} value={String(product.id)}>{textValue(product.mi_sku)} · {textValue(product.title)}</option>)}</select></label><label className="form-field"><span>Supplier</span><select value={oilForm.supplier_id} onChange={(event) => setOilForm({ ...oilForm, supplier_id: event.target.value })}><option value="">Unassigned</option>{suppliers.map((supplier) => <option key={String(supplier.id)} value={String(supplier.id)}>{textValue(supplier.name)}</option>)}</select></label><label className="form-field"><span>Purchase price / kg</span><input type="number" min="0" step="0.01" value={oilForm.purchase_price_per_kg} onChange={(event) => setOilForm({ ...oilForm, purchase_price_per_kg: event.target.value })} /></label><label className="form-field"><span>Grams left</span><input type="number" min="0" step="0.01" value={oilForm.grams_left} onChange={(event) => setOilForm({ ...oilForm, grams_left: event.target.value })} /></label></div></> : section === 'suppliers' ? <><label className="form-field"><span>Supplier name</span><input required value={supplierForm.name} onChange={(event) => setSupplierForm({ ...supplierForm, name: event.target.value })} /></label><label className="form-field"><span>Contact information</span><textarea rows={3} value={supplierForm.contact_info} onChange={(event) => setSupplierForm({ ...supplierForm, contact_info: event.target.value })} /></label></> : section === 'purchase-orders' ? <><div className="form-grid-two"><label className="form-field"><span>Supplier</span><select required value={poForm.supplier_id} onChange={(event) => setPOForm({ ...poForm, supplier_id: event.target.value })}><option value="">Select supplier</option>{suppliers.map((supplier) => <option key={String(supplier.id)} value={String(supplier.id)}>{textValue(supplier.name)}</option>)}</select></label><label className="form-field"><span>Oil</span><select required value={poForm.oil_inventory_id} onChange={(event) => setPOForm({ ...poForm, oil_inventory_id: event.target.value })}><option value="">Select oil</option>{oils.map((oil) => <option key={String(oil.id)} value={String(oil.id)}>{textValue(oil.name)}</option>)}</select></label><label className="form-field"><span>Quantity in grams</span><input required type="number" min="0" step="0.01" value={poForm.quantity_grams} onChange={(event) => setPOForm({ ...poForm, quantity_grams: event.target.value })} /></label><label className="form-field"><span>Price per kg</span><input required type="number" min="0" step="0.01" value={poForm.unit_price_per_kg} onChange={(event) => setPOForm({ ...poForm, unit_price_per_kg: event.target.value })} /></label></div><label className="form-field"><span>Purchase date</span><input required type="date" value={poForm.purchase_date} onChange={(event) => setPOForm({ ...poForm, purchase_date: event.target.value })} /></label></> : manufacturingFormContent}<div className="modal-actions">{error && <p className="modal-form-error" role="alert"><CircleAlert size={15} aria-hidden="true" /> {error}</p>}<button className="secondary-button" type="button" onClick={() => setModal(false)}>Cancel</button><button className="primary-button" type="submit" disabled={isWorking}><Plus size={14} aria-hidden="true" /> {isWorking ? 'Saving…' : 'Save record'}</button></div></form></div>}</section>
}
