import { useState, useEffect } from 'react';
import './App.css';
import { Dashboard } from './modules/Dashboard';
import { Orders } from './modules/Orders';
import { Customers } from './modules/Customers';
import { GSTReports } from './modules/GSTReports';
import { Settings } from './modules/Settings';

const App = () => {
  const [theme, setTheme] = useState<'light' | 'dark'>(() => {
    return (localStorage.getItem('mobileAppTheme') as 'light' | 'dark') || 'light';
  });
  const [isNavOpen, setIsNavOpen] = useState(false);
  const [activeTab, setActiveTab] = useState('dashboard');
  const [isOffline, setIsOffline] = useState(!navigator.onLine);

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('mobileAppTheme', theme);
  }, [theme]);

  useEffect(() => {
    if (isNavOpen) {
      document.body.classList.add('scroll-lock');
    } else {
      document.body.classList.remove('scroll-lock');
    }
  }, [isNavOpen]);

  useEffect(() => {
    const handleOnline = () => setIsOffline(false);
    const handleOffline = () => setIsOffline(true);

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);

  const toggleTheme = () => setTheme(t => t === 'light' ? 'dark' : 'light');

  const navItems = [
    { id: 'dashboard', label: 'Dashboard', icon: '🏠' },
    { id: 'orders', label: 'Orders', icon: '📦' },
    { id: 'customers', label: 'Customers', icon: '👥' },
    { id: 'gst', label: 'GST Reports', icon: '📄' },
    { id: 'settings', label: 'Settings', icon: '⚙️' },
  ];

  const handleNavClick = (id: string) => {
    setActiveTab(id);
    setIsNavOpen(false);
  };

  return (
    <div className="app-shell">
      {isOffline && (
        <div className="offline-banner">
          📶 Poor connection. Some data may be outdated.
        </div>
      )}

      <header className="sticky-header">
        <div className="header-content">
          <div className="logo">GST Admin</div>
          <div className="header-actions">
            <button className="icon-btn" onClick={toggleTheme} aria-label="Toggle theme">
              {theme === 'light' ? '🌙' : '☀️'}
            </button>
            <button className="icon-btn menu-trigger" onClick={() => setIsNavOpen(true)} aria-label="Open navigation">
              ☰
            </button>
          </div>
        </div>
      </header>

      <main className="main-content tab-content-fade" key={activeTab}>
        {activeTab === 'dashboard' && <Dashboard />}
        {activeTab === 'orders' && <Orders />}
        {activeTab === 'customers' && <Customers />}
        {activeTab === 'gst' && <GSTReports />}
        {activeTab === 'settings' && <Settings />}
      </main>

      {/* Navigation Overlay */}
      {isNavOpen && (
        <div className="nav-overlay glass-blur" onClick={() => setIsNavOpen(false)}>
          <div className="nav-container" onClick={e => e.stopPropagation()}>
            <div className="nav-header">
              <div className="logo">GST Admin</div>
              <button className="icon-btn" onClick={() => setIsNavOpen(false)}>✕</button>
            </div>
            <div className="nav-grid">
              {navItems.map((item, index) => (
                <button
                  key={item.id}
                  className={`nav-card nav-item-stagger ${activeTab === item.id ? 'active' : ''}`}
                  onClick={() => handleNavClick(item.id)}
                  style={{ animationDelay: `${index * 50}ms` }}
                >
                  <span className="nav-card-icon">{item.icon}</span>
                  <span className="nav-card-label">{item.label}</span>
                </button>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default App;
