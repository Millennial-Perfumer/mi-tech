import React, { useEffect } from 'react';
import { useMetricStore } from '../store/useMetricStore';
import { TrendingUp, Package, CheckCircle, AlertCircle, RefreshCw } from 'lucide-react';

export const Dashboard: React.FC = () => {
  const { metrics, isLoading, fetchMetrics } = useMetricStore();

  useEffect(() => {
    // Default to a 30-day view or current month
    const end = new Date();
    const start = new Date();
    start.setDate(end.getDate() - 30);
    
    fetchMetrics(start.toISOString().split('T')[0], end.toISOString().split('T')[0]);
  }, []);

  const StatCard = ({ title, value, icon: Icon, color }: any) => (
    <div className="glass-card" style={{ padding: '1.25rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '1rem' }}>
        <p style={{ fontSize: '0.75rem', fontWeight: 600, textTransform: 'uppercase', color: 'var(--text-tertiary)' }}>{title}</p>
        <div style={{ color }}>
          <Icon size={18} />
        </div>
      </div>
      <h2 style={{ fontSize: '1.5rem', color: '#fff' }}>{value}</h2>
    </div>
  );

  return (
    <div>
      <header style={{ marginBottom: '2.5rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <p style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--accent-color)', textTransform: 'uppercase', letterSpacing: '0.15em' }}>Overview</p>
          <h1>Console</h1>
        </div>
        <button 
          onClick={() => {
            const end = new Date();
            const start = new Date();
            start.setDate(end.getDate() - 30);
            fetchMetrics(start.toISOString().split('T')[0], end.toISOString().split('T')[0]);
          }}
          className={isLoading ? 'animate-spin' : ''}
          style={{ color: 'var(--text-tertiary)' }}
        >
          <RefreshCw size={20} />
        </button>
      </header>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1.5rem' }}>
        <StatCard 
          title="Revenue" 
          value={`₹${(metrics?.total_revenue || 0).toLocaleString()}`} 
          icon={TrendingUp} 
          color="var(--accent-color)" 
        />
        <StatCard 
          title="GST Collected" 
          value={`₹${(metrics?.total_gst_collected || 0).toLocaleString()}`} 
          icon={CheckCircle} 
          color="var(--secondary-color)" 
        />
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1.5rem' }}>
        <StatCard 
          title="Total Orders" 
          value={metrics?.total_orders || 0} 
          icon={Package} 
          color="#8b5cf6" 
        />
        <StatCard 
          title="Unfulfilled" 
          value={metrics?.unfulfilled_orders || 0} 
          icon={AlertCircle} 
          color="#ef4444" 
        />
      </div>

      <div className="glass-card" style={{ marginTop: '1rem' }}>
        <h3 style={{ marginBottom: '1rem' }}>Fulfillment Health</h3>
        <div style={{ height: '8px', background: 'rgba(255,255,255,0.05)', borderRadius: '4px', overflow: 'hidden', display: 'flex' }}>
          <div style={{ 
            width: `${((metrics?.fulfilled_orders || 0) / (metrics?.total_orders || 1)) * 100}%`, 
            background: 'var(--accent-color)' 
          }} />
        </div>
        <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '0.75rem', fontSize: '0.8rem' }}>
          <span>{metrics?.fulfilled_orders || 0} Fulfilled</span>
          <span>{metrics?.total_orders || 0} Total</span>
        </div>
      </div>
    </div>
  );
};
