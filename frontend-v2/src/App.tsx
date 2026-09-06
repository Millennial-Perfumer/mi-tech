import { useState } from 'react'
import {
  BarChart3,
  Boxes,
  ChevronRight,
  CircleUserRound,
  FileText,
  LayoutDashboard,
  Menu,
  MessageSquare,
  MoreHorizontal,
  Search,
  Settings,
  ShoppingBag,
  Ticket,
  Users,
  X,
} from 'lucide-react'

type ViewId =
  | 'dashboard'
  | 'orders'
  | 'customers'
  | 'reports'
  | 'inventory'
  | 'communication'
  | 'tickets'
  | 'marketing'
  | 'settings'

type NavigationItem = {
  id: ViewId
  label: string
  icon: typeof LayoutDashboard
}

const navigationGroups: { label: string; items: NavigationItem[] }[] = [
  {
    label: 'Operations',
    items: [
      { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
      { id: 'orders', label: 'Orders', icon: ShoppingBag },
      { id: 'customers', label: 'Customers', icon: Users },
      { id: 'reports', label: 'GST reports', icon: FileText },
      { id: 'inventory', label: 'Inventory', icon: Boxes },
    ],
  },
  {
    label: 'Engagement',
    items: [
      { id: 'communication', label: 'Communication', icon: MessageSquare },
      { id: 'tickets', label: 'Support tickets', icon: Ticket },
      { id: 'marketing', label: 'Marketing', icon: BarChart3 },
    ],
  },
]

const pageTitles: Record<ViewId, string> = {
  dashboard: 'Overview',
  orders: 'Orders',
  customers: 'Customers',
  reports: 'GST reports',
  inventory: 'Inventory',
  communication: 'Communication',
  tickets: 'Support tickets',
  marketing: 'Marketing',
  settings: 'Settings',
}

const metrics = [
  { label: 'Revenue', value: '₹0', detail: 'Connect data to begin' },
  { label: 'Orders', value: '0', detail: 'No orders loaded' },
  { label: 'GST collected', value: '₹0', detail: 'No reports loaded' },
  { label: 'Open tickets', value: '0', detail: 'No tickets loaded' },
]

function App() {
  const [activeView, setActiveView] = useState<ViewId>('dashboard')
  const [isMobileNavOpen, setIsMobileNavOpen] = useState(false)

  const navigate = (view: ViewId) => {
    setActiveView(view)
    setIsMobileNavOpen(false)
  }

  return (
    <div className="app-shell">
      <a className="skip-link" href="#main-content">
        Skip to content
      </a>

      {isMobileNavOpen && (
        <button
          className="navigation-scrim"
          type="button"
          aria-label="Close navigation"
          onClick={() => setIsMobileNavOpen(false)}
        />
      )}

      <aside
        className={`sidebar ${isMobileNavOpen ? 'sidebar-open' : ''}`}
        aria-label="Primary navigation"
      >
        <div className="sidebar-header">
          <div className="brand-mark" aria-hidden="true">
            M
          </div>
          <div className="brand-copy">
            <span className="brand-name">Mi Tech</span>
            <span className="brand-caption">Operations workspace</span>
          </div>
          <button
            className="icon-button mobile-close"
            type="button"
            aria-label="Close navigation"
            onClick={() => setIsMobileNavOpen(false)}
          >
            <X size={20} strokeWidth={1.8} />
          </button>
        </div>

        <nav className="sidebar-navigation">
          {navigationGroups.map((group) => (
            <div className="navigation-group" key={group.label}>
              <p className="navigation-label">{group.label}</p>
              {group.items.map((item) => {
                const Icon = item.icon
                const isActive = item.id === activeView

                return (
                  <button
                    className={`navigation-item ${isActive ? 'navigation-item-active' : ''}`}
                    type="button"
                    key={item.id}
                    aria-current={isActive ? 'page' : undefined}
                    onClick={() => navigate(item.id)}
                  >
                    <Icon size={18} strokeWidth={1.8} aria-hidden="true" />
                    <span>{item.label}</span>
                  </button>
                )
              })}
            </div>
          ))}
        </nav>

        <div className="sidebar-footer">
          <button
            className={`navigation-item ${activeView === 'settings' ? 'navigation-item-active' : ''}`}
            type="button"
            aria-current={activeView === 'settings' ? 'page' : undefined}
            onClick={() => navigate('settings')}
          >
            <Settings size={18} strokeWidth={1.8} aria-hidden="true" />
            <span>Settings</span>
          </button>
          <div className="user-summary">
            <span className="user-avatar" aria-hidden="true">
              <CircleUserRound size={20} strokeWidth={1.7} />
            </span>
            <span>
              <strong>Workspace user</strong>
              <small>Role-aware access</small>
            </span>
          </div>
        </div>
      </aside>

      <div className="app-content">
        <header className="topbar">
          <button
            className="icon-button mobile-menu"
            type="button"
            aria-label="Open navigation"
            aria-expanded={isMobileNavOpen}
            onClick={() => setIsMobileNavOpen(true)}
          >
            <Menu size={21} strokeWidth={1.8} />
          </button>
          <div className="topbar-heading">
            <span className="eyebrow">Workspace</span>
            <h1>{pageTitles[activeView]}</h1>
          </div>
          <div className="topbar-actions">
            <button className="icon-button" type="button" aria-label="Search">
              <Search size={19} strokeWidth={1.8} />
            </button>
            <button className="profile-button" type="button">
              <span className="profile-dot" aria-hidden="true" />
              <span className="profile-label">Admin</span>
              <ChevronRight size={16} strokeWidth={1.8} aria-hidden="true" />
            </button>
          </div>
        </header>

        <main className="main-content" id="main-content" tabIndex={-1}>
          <section className="intro-block" aria-labelledby="page-heading">
            <div>
              <p className="eyebrow">Mi Tech / {pageTitles[activeView]}</p>
              <h2 id="page-heading">
                A clearer, calmer way to run your business.
              </h2>
              <p className="intro-copy">
                The v2 foundation is ready for feature-by-feature migration.
              </p>
            </div>
            <button className="primary-button" type="button">
              Connect data
              <ChevronRight size={17} strokeWidth={1.9} aria-hidden="true" />
            </button>
          </section>

          {activeView === 'dashboard' ? (
            <>
              <section className="metrics-grid" aria-label="Business overview">
                {metrics.map((metric) => (
                  <article className="metric-card" key={metric.label}>
                    <span className="metric-label">{metric.label}</span>
                    <strong className="metric-value">{metric.value}</strong>
                    <span className="metric-detail">{metric.detail}</span>
                  </article>
                ))}
              </section>

              <section className="empty-panel" aria-labelledby="migration-heading">
                <div className="empty-panel-icon" aria-hidden="true">
                  <LayoutDashboard size={22} strokeWidth={1.7} />
                </div>
                <div>
                  <p className="eyebrow">Foundation milestone</p>
                  <h2 id="migration-heading">Ready for the first feature slice</h2>
                  <p>
                    Shared navigation, responsive layout, warm monochrome tokens,
                    and keyboard-visible focus states are now in place.
                  </p>
                </div>
              </section>
            </>
          ) : (
            <section className="empty-panel" aria-labelledby="view-heading">
              <div className="empty-panel-icon" aria-hidden="true">
                <MoreHorizontal size={22} strokeWidth={1.7} />
              </div>
              <div>
                <p className="eyebrow">Migration queue</p>
                <h2 id="view-heading">{pageTitles[activeView]} is next</h2>
                <p>
                  This route is reserved for the existing {pageTitles[activeView].toLowerCase()}{' '}
                  feature set and will be migrated without changing its API behavior.
                </p>
              </div>
            </section>
          )}
        </main>

        <nav className="mobile-tab-bar" aria-label="Mobile primary navigation">
          {[
            navigationGroups[0].items[0],
            navigationGroups[0].items[1],
            navigationGroups[0].items[2],
            { id: 'settings' as ViewId, label: 'More', icon: MoreHorizontal },
          ].map((item) => {
            const Icon = item.icon
            const isActive = item.id === activeView

            return (
              <button
                className={`mobile-tab ${isActive ? 'mobile-tab-active' : ''}`}
                type="button"
                key={item.id}
                aria-current={isActive ? 'page' : undefined}
                onClick={() => navigate(item.id)}
              >
                <Icon size={18} strokeWidth={1.8} aria-hidden="true" />
                <span>{item.label}</span>
              </button>
            )
          })}
        </nav>
      </div>
    </div>
  )
}

export default App
