import { useState } from 'react'
import { 
  Home, 
  Package, 
  FileText, 
  Zap, 
  Users,
  Settings as LucideSettings
} from 'lucide-react'
import { useAuthStore } from './store/useAuthStore'

import { Login } from './pages/Login'
import { Dashboard } from './pages/Dashboard'
import { Orders } from './pages/Orders'
import { Customers } from './pages/Customers'
import { Marketing } from './pages/Marketing'
import { Automation } from './pages/Automation'
import { Settings } from './pages/Settings'

function App() {
  const [activeTab, setActiveTab] = useState('home')
  const { token } = useAuthStore()

  if (!token) {
    return <Login />
  }

  const renderContent = () => {
    switch (activeTab) {
      case 'home':
        return <Dashboard />
      case 'orders':
        return <Orders />
      case 'crm':
        return <Customers />
      case 'marketing':
        return <Marketing />
      case 'auto':
        return <Automation />
      case 'more':
        return <Settings />
      default:
        return (
          <div style={{ textAlign: 'center', marginTop: '50%' }}>
            <h1>{activeTab.toUpperCase()}</h1>
            <p>Coming Soon</p>
          </div>
        )
    }
  }

  return (
    <div className="mobile-shell">
      <main className="main-scroll">
        {renderContent()}
      </main>

      <nav className="bottom-nav glass-panel">
        {[
          { id: 'home', icon: Home, label: 'Home' },
          { id: 'orders', icon: Package, label: 'Orders' },
          { id: 'crm', icon: Users, label: 'CRM' },
          { id: 'marketing', icon: FileText, label: 'Meta' },
          { id: 'auto', icon: Zap, label: 'Auto' },
          { id: 'more', icon: LucideSettings, label: 'More' }
        ].map((item) => (
          <button 
            key={item.id}
            onClick={() => setActiveTab(item.id)}
            style={{ 
              flexDirection: 'column', 
              color: activeTab === item.id ? 'var(--accent-color)' : 'var(--text-tertiary)',
              gap: '4px',
              background: 'transparent',
              border: 'none',
              outline: 'none',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center'
            }}
          >
            <item.icon size={20} />
            <span style={{ fontSize: '10px', fontWeight: 600 }}>{item.label}</span>
          </button>
        ))}
      </nav>
    </div>
  )
}

export default App
