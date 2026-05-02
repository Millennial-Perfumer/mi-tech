import React, { useState } from 'react';
import { Products } from './Products';
import { OilInventory } from './OilInventory';
import { Suppliers } from './Suppliers';
import { PurchaseOrders } from './PurchaseOrders';
import { Manufacturing } from './Manufacturing';

export const InventoryHub: React.FC<{ token: string | null, userRole?: string, appConfigs?: any }> = (props) => {
  const [activeSubTab, setActiveSubTab] = useState<'products' | 'oil' | 'suppliers' | 'po' | 'manufacturing'>('products');

  const tabs = [
    { id: 'products', label: 'Warehouse Authority' },
    { id: 'oil', label: 'Oil Stock' },
    { id: 'suppliers', label: 'Suppliers' },
    { id: 'po', label: 'Purchase Orders' },
    { id: 'manufacturing', label: 'Manufacturing' }
  ];

  return (
    <div className="inventory-hub-container">
      <div className="sub-tab-nav" style={{ 
        display: 'flex', 
        gap: '2.5rem', 
        marginBottom: '2rem', 
        borderBottom: '1px solid var(--border-color)',
        padding: '0 0.5rem'
      }}>
        {tabs.map(tab => (
          <button
            key={tab.id}
            onClick={() => setActiveSubTab(tab.id as any)}
            className={`sub-tab-btn ${activeSubTab === tab.id ? 'active' : ''}`}
            style={{
              padding: '0.75rem 0',
              border: 'none',
              borderBottom: activeSubTab === tab.id ? '2px solid var(--accent-color)' : '2px solid transparent',
              background: 'transparent',
              color: activeSubTab === tab.id ? 'var(--accent-color)' : 'var(--text-secondary)',
              fontWeight: 600,
              fontSize: '0.95rem',
              cursor: 'pointer',
              transition: 'all 0.2s',
              marginBottom: '-1px'
            }}
          >
            {tab.label}
          </button>
        ))}
      </div>

      <div className="inventory-sub-content">
        {activeSubTab === 'products' && <Products {...props} />}
        {activeSubTab === 'oil' && <OilInventory token={props.token} />}
        {activeSubTab === 'suppliers' && <Suppliers token={props.token} />}
        {activeSubTab === 'po' && <PurchaseOrders token={props.token} />}
        {activeSubTab === 'manufacturing' && <Manufacturing token={props.token} />}
      </div>
    </div>
  );
};
