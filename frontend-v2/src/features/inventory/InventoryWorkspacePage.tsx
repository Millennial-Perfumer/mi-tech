import { useState } from 'react'
import { Boxes, FlaskConical, PackagePlus, Truck } from 'lucide-react'
import { InventoryOperations, InventorySection } from './InventoryOperations'
import { InventoryPage } from './InventoryPage'

type Props = { token: string; onUnauthorized: () => void }
type Section = 'products' | InventorySection

const sections: { id: Section; label: string; icon: typeof Boxes }[] = [
  { id: 'products', label: 'Products', icon: Boxes },
  { id: 'oils', label: 'Oil inventory', icon: FlaskConical },
  { id: 'suppliers', label: 'Suppliers', icon: Truck },
  { id: 'purchase-orders', label: 'Purchase orders', icon: PackagePlus },
  { id: 'manufacturing', label: 'Manufacturing', icon: PackagePlus },
]

export function InventoryWorkspacePage({ token, onUnauthorized }: Props) {
  const [section, setSection] = useState<Section>('products')
  return <section className="workspace-page inventory-workspace-page" aria-labelledby="inventory-workspace-heading"><header className="workspace-page-header"><div><p className="eyebrow">Operations / Inventory</p><h2 id="inventory-workspace-heading">Everything needed to keep stock moving.</h2><p>Products, raw materials, suppliers, purchasing, and production now live in one consistent workspace.</p></div></header><nav className="inventory-section-tabs" aria-label="Inventory sections">{sections.map((item) => { const Icon = item.icon; return <button key={item.id} className={`filter-chip ${section === item.id ? 'filter-chip-active' : ''}`} type="button" aria-current={section === item.id ? 'page' : undefined} onClick={() => setSection(item.id)}><Icon size={14} aria-hidden="true" /> {item.label}</button> })}</nav>{section === 'products' ? <InventoryPage token={token} onUnauthorized={onUnauthorized} embedded /> : <InventoryOperations token={token} onUnauthorized={onUnauthorized} section={section} />}</section>
}
