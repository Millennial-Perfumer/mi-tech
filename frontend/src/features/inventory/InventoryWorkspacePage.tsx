import { useState } from 'react'
import { Boxes, FlaskConical, PackagePlus, Truck } from 'lucide-react'
import { InventoryOperations, InventorySection } from './InventoryOperations'
import { InventoryPage } from './InventoryPage'

type Props = { token: string; onUnauthorized: () => void }
type Section = 'products' | InventorySection

const sections: { id: Section; label: string; detail: string; icon: typeof Boxes }[] = [
  { id: 'products', label: 'Products', detail: 'Finished catalogue', icon: Boxes },
  { id: 'oils', label: 'Oil inventory', detail: 'Raw materials', icon: FlaskConical },
  { id: 'suppliers', label: 'Suppliers', detail: 'Vendor records', icon: Truck },
  { id: 'purchase-orders', label: 'Purchase orders', detail: 'Incoming stock', icon: PackagePlus },
  { id: 'manufacturing', label: 'Manufacturing', detail: 'Production runs', icon: PackagePlus },
]

export function InventoryWorkspacePage({ token, onUnauthorized }: Props) {
  const [section, setSection] = useState<Section>('products')
  return <section className="workspace-page inventory-workspace-page" aria-labelledby="inventory-workspace-heading"><header className="workspace-page-header"><div><p className="eyebrow">Operations / Inventory</p><h2 id="inventory-workspace-heading">Inventory</h2><p>Manage products, oils, suppliers, purchase orders, and manufacturing.</p></div></header><nav className="report-tabs inventory-section-tabs" aria-label="Inventory sections">{sections.map((item) => { const Icon = item.icon; return <button key={item.id} className={`report-tab inventory-section-tab ${section === item.id ? 'report-tab-active' : ''}`} type="button" aria-current={section === item.id ? 'page' : undefined} onClick={() => setSection(item.id)}><span className="inventory-tab-label"><Icon size={14} aria-hidden="true" /> {item.label}</span><small>{item.detail}</small></button> })}</nav>{section === 'products' ? <InventoryPage token={token} onUnauthorized={onUnauthorized} embedded /> : <InventoryOperations token={token} onUnauthorized={onUnauthorized} section={section} />}</section>
}
