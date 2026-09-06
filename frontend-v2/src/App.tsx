import { useState } from 'react'
import fullLogo from './assets/mi-black-full.png'
import { AuthScreen } from './components/AuthScreen'
import { DateRangePicker } from './components/DateRangePicker'
import { DashboardPage } from './features/dashboard/DashboardPage'
import { OrdersPage } from './features/orders/OrdersPage'
import { ReportsPage } from './features/reports/ReportsPage'
import { CustomersPage } from './features/customers/CustomersPage'
import { InventoryWorkspacePage } from './features/inventory/InventoryWorkspacePage'
import { CommunicationPage } from './features/communication/CommunicationPage'
import { SupportPage } from './features/support/SupportPage'
import { FeedbackPage } from './features/feedback/FeedbackPage'
import { AbandonedCartsPage } from './features/abandoned/AbandonedCartsPage'
import { MarketingPage } from './features/marketing/MarketingPage'
import { B2BPage } from './features/b2b/B2BPage'
import { AutomationPage } from './features/automation/AutomationPage'
import { SocialPage } from './features/social/SocialPage'
import { PlannerPage } from './features/planner/PlannerPage'
import { AIPage } from './features/ai/AIPage'
import { UsersPage } from './features/users/UsersPage'
import { SettingsPage } from './features/settings/SettingsPage'
import { JudgeMePage } from './features/judgeme/JudgeMePage'
import { TableSortEnhancer } from './components/TableSortEnhancer'
import { usePeriodFilter } from './lib/usePeriodFilter'
import {
  BarChart3,
  Boxes,
  CheckCircle2,
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
  Sparkles,
  Star,
  Ticket,
  Users,
  X,
} from 'lucide-react'

type ViewId =
  | 'dashboard'
  | 'shopify'
  | 'customers'
  | 'reports'
  | 'b2b'
  | 'inventory'
  | 'communication'
  | 'tickets'
  | 'feedback'
  | 'automation'
  | 'abandoned-carts'
  | 'marketing'
  | 'social'
  | 'social-queue'
  | 'judgeme'
  | 'planner'
  | 'ai-analysis'
  | 'users'
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
      { id: 'shopify', label: 'Orders', icon: ShoppingBag },
      { id: 'customers', label: 'Customers', icon: Users },
      { id: 'reports', label: 'GST reports', icon: FileText },
      { id: 'b2b', label: 'B2B billing', icon: FileText },
      { id: 'inventory', label: 'Inventory', icon: Boxes },
    ],
  },
  {
    label: 'Customer care',
    items: [
      { id: 'communication', label: 'Communication', icon: MessageSquare },
      { id: 'tickets', label: 'Support tickets', icon: Ticket },
      { id: 'feedback', label: 'Customer feedback', icon: MessageSquare },
      { id: 'abandoned-carts', label: 'Abandoned carts', icon: ShoppingBag },
    ],
  },
  {
    label: 'Growth',
    items: [
      { id: 'marketing', label: 'Marketing', icon: BarChart3 },
      { id: 'social', label: 'Social media', icon: BarChart3 },
      { id: 'social-queue', label: 'Auto queue', icon: Boxes },
      { id: 'judgeme', label: 'Judge.me reviews', icon: Star },
    ],
  },
  {
    label: 'Automation & planning',
    items: [
      { id: 'automation', label: 'Automation', icon: CheckCircle2 },
      { id: 'planner', label: 'Planner', icon: LayoutDashboard },
      { id: 'ai-analysis', label: 'AI analysis', icon: Sparkles },
    ],
  },
]

const pageTitles: Record<ViewId, string> = {
  dashboard: 'Overview',
  shopify: 'Orders',
  customers: 'Customers',
  reports: 'GST reports',
  b2b: 'B2B billing',
  inventory: 'Inventory',
  communication: 'Communication',
  tickets: 'Support tickets',
  feedback: 'Customer feedback',
  automation: 'Automation',
  'abandoned-carts': 'Abandoned carts',
  marketing: 'Marketing',
  social: 'Social media',
  'social-queue': 'Auto queue',
  judgeme: 'Judge.me reviews',
  planner: 'Planner',
  'ai-analysis': 'AI analysis',
  users: 'User roles',
  settings: 'Settings',
}

const pageDescriptions: Record<ViewId, string> = {
  dashboard: 'Monitor revenue, orders, GST, and operational health.',
  shopify: 'Review, search, and action orders from every connected channel.',
  customers: 'Manage your customer directory and engagement history.',
  reports: 'Generate and export GST-ready reports.',
  b2b: 'Create GST-compliant B2B invoices and manage business customers.',
  inventory: 'Manage products, stock, suppliers, and manufacturing.',
  communication: 'Customer conversations and WhatsApp activity.',
  tickets: 'Track, assign, and resolve customer concerns.',
  feedback: 'Review customer sentiment and follow-up opportunities.',
  automation: 'Manage templates, triggers, and messaging automation.',
  'abandoned-carts': 'Review abandoned checkouts and recovery messages.',
  marketing: 'Review paid marketing performance.',
  social: 'Manage social channels and publishing activity.',
  'social-queue': 'Manage scheduled social content and queues.',
  judgeme: 'Review and manage Judge.me feedback.',
  planner: 'Organize work across boards, sprints, and tasks.',
  'ai-analysis': 'Analyze business data with AI assistance.',
  users: 'Manage workspace access and roles.',
  settings: 'Manage integrations, store data, and workspace preferences.',
}

const periodFilteredViews: ViewId[] = ['dashboard', 'shopify', 'reports', 'abandoned-carts', 'marketing', 'automation', 'social']

function App() {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('token'))
  const [activeView, setActiveView] = useState<ViewId>(() => {
    const savedView = localStorage.getItem('gstAppActiveTab') as ViewId | null
    return savedView && savedView in pageTitles ? savedView : 'dashboard'
  })
  const [isMobileNavOpen, setIsMobileNavOpen] = useState(false)
  const { startDate, endDate, setDateRange } = usePeriodFilter()

  const handleLogin = (newToken: string) => {
    localStorage.setItem('token', newToken)
    setToken(newToken)
  }

  const handleLogout = () => {
    localStorage.removeItem('token')
    setToken(null)
  }

  const navigate = (view: ViewId) => {
    setActiveView(view)
    localStorage.setItem('gstAppActiveTab', view)
    setIsMobileNavOpen(false)
  }

  const focusPageSearch = () => {
    const searchField = document.querySelector<HTMLInputElement>('#main-content input[placeholder*="Search"], #main-content textarea[placeholder*="Ask"]')
    searchField?.focus()
  }

  if (!token) return <AuthScreen onLogin={handleLogin} />

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
          <div className="brand-lockup">
            <img className="brand-logo-full" src={fullLogo} alt="Millennial Perfumer" />
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
          <div className="sidebar-footer-navigation">
            <p className="navigation-label">Administration</p>
            <button
              className={`navigation-item ${activeView === 'users' ? 'navigation-item-active' : ''}`}
              type="button"
              aria-current={activeView === 'users' ? 'page' : undefined}
              onClick={() => navigate('users')}
            >
              <Users size={18} strokeWidth={1.8} aria-hidden="true" />
              <span>User roles</span>
            </button>
            <button
              className={`navigation-item ${activeView === 'settings' ? 'navigation-item-active' : ''}`}
              type="button"
              aria-current={activeView === 'settings' ? 'page' : undefined}
              onClick={() => navigate('settings')}
            >
              <Settings size={18} strokeWidth={1.8} aria-hidden="true" />
              <span>Settings</span>
            </button>
          </div>
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
          <span className="mobile-brand-lockup" aria-label="Millennial Perfumer">
            <img className="mobile-brand-logo-full" src={fullLogo} alt="Millennial Perfumer" />
          </span>
          {periodFilteredViews.includes(activeView) && (
            <div className="topbar-period-filter">
              <DateRangePicker startDate={startDate} endDate={endDate} onChange={setDateRange} />
            </div>
          )}
          <div className="topbar-actions">
            <button className="icon-button" type="button" aria-label="Focus page search" onClick={focusPageSearch}>
              <Search size={19} strokeWidth={1.8} />
            </button>
            <button className="profile-button" type="button" aria-label="Sign out" onClick={handleLogout}>
              <span className="profile-dot" aria-hidden="true" />
              <span className="profile-label">Admin</span>
              <ChevronRight size={16} strokeWidth={1.8} aria-hidden="true" />
            </button>
          </div>
        </header>

        <main className="main-content" id="main-content" tabIndex={-1}>
          <TableSortEnhancer>
          {activeView === 'dashboard' ? (
            <DashboardPage token={token} onUnauthorized={handleLogout} />
          ) : activeView === 'shopify' ? (
            <OrdersPage token={token} onUnauthorized={handleLogout} />
          ) : activeView === 'reports' ? (
            <ReportsPage token={token} onUnauthorized={handleLogout} />
          ) : activeView === 'customers' ? (
            <CustomersPage token={token} onUnauthorized={handleLogout} />
          ) : activeView === 'inventory' ? (
            <InventoryWorkspacePage token={token} onUnauthorized={handleLogout} />
          ) : activeView === 'communication' ? (
            <CommunicationPage token={token} onUnauthorized={handleLogout} />
          ) : activeView === 'tickets' ? (
            <SupportPage token={token} onUnauthorized={handleLogout} />
          ) : activeView === 'feedback' ? (
            <FeedbackPage token={token} onUnauthorized={handleLogout} />
          ) : activeView === 'abandoned-carts' ? (
            <AbandonedCartsPage token={token} onUnauthorized={handleLogout} />
          ) : activeView === 'marketing' ? (
            <MarketingPage token={token} onUnauthorized={handleLogout} />
          ) : activeView === 'b2b' ? (
            <B2BPage token={token} onUnauthorized={handleLogout} />
          ) : activeView === 'automation' ? (
            <AutomationPage token={token} onUnauthorized={handleLogout} />
          ) : activeView === 'social' ? (
            <SocialPage token={token} onUnauthorized={handleLogout} />
          ) : activeView === 'social-queue' ? (
            <SocialPage token={token} onUnauthorized={handleLogout} initialTab="queue" />
          ) : activeView === 'judgeme' ? (
            <JudgeMePage token={token} onUnauthorized={handleLogout} />
          ) : activeView === 'planner' ? (
            <PlannerPage token={token} onUnauthorized={handleLogout} />
          ) : activeView === 'ai-analysis' ? (
            <AIPage token={token} onUnauthorized={handleLogout} />
          ) : activeView === 'users' ? (
            <UsersPage token={token} onUnauthorized={handleLogout} />
          ) : activeView === 'settings' ? (
            <SettingsPage token={token} onUnauthorized={handleLogout} />
          ) : (
            <section className="migration-panel" aria-labelledby="view-heading">
              <div className="migration-panel-icon" aria-hidden="true">
                <MoreHorizontal size={22} strokeWidth={1.7} />
              </div>
              <div>
                <p className="eyebrow">Migration queue</p>
                <h2 id="view-heading">{pageTitles[activeView]}</h2>
                <p>{pageDescriptions[activeView]}</p>
                <span className="migration-status">Next v2 feature slice</span>
              </div>
            </section>
          )}
          </TableSortEnhancer>
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
                aria-expanded={item.id === 'settings' ? isMobileNavOpen : undefined}
                aria-label={item.id === 'settings' ? 'Open all workspace sections' : item.label}
                onClick={() => item.id === 'settings' ? setIsMobileNavOpen(true) : navigate(item.id)}
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
