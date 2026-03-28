import React, { useState, useEffect } from 'react';
import { API_BASE, fetchWithAuth, getTodayIST } from '../api';
import { MobileCard } from '../components/MobileCard';
import { BottomSheet } from '../components/BottomSheet';
import './Dashboard.css';

interface DashboardMetrics {
  total_revenue: number;
  total_invoices: number;
  total_gst_collected: number;
  cgst_collected: number;
  sgst_collected: number;
  igst_collected: number;
  total_orders: number;
  cancelled_orders: number;
  fulfilled_orders: number;
  unfulfilled_orders: number;
}

export const Dashboard: React.FC = () => {
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedMetric, setSelectedMetric] = useState<string | null>(null);

  const fetchMetrics = async () => {
    setIsLoading(true);
    try {
      const today = getTodayIST();
      // Start of month to today
      const firstDay = new Date(new Date().getFullYear(), new Date().getMonth(), 1).toISOString().split('T')[0];
      const res = await fetchWithAuth(`${API_BASE}/api/dashboard/metrics?start_date=${firstDay}T00:00:00Z&end_date=${today}T23:59:59Z`);
      const data = await res.json();
      if (data.success) {
        setMetrics(data.metrics);
      }
    } catch (err) {
      console.error('Error fetching metrics:', err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchMetrics();
  }, []);

  const metricCards = [
    { id: 'revenue', label: 'Total Revenue', value: `₹${metrics?.total_revenue?.toLocaleString('en-IN') || 0}`, icon: '💰', color: '#6366f1' },
    { id: 'gst', label: 'GST Collected', value: `₹${metrics?.total_gst_collected?.toLocaleString('en-IN') || 0}`, icon: '📊', color: '#10b981' },
    { id: 'orders', label: 'Total Orders', value: metrics?.total_orders || 0, icon: '📦', color: '#f59e0b' },
    { id: 'fulfilled', label: 'Fulfilled', value: metrics?.fulfilled_orders || 0, icon: '✅', color: '#10b981' },
    { id: 'pending', label: 'Unfulfilled', value: metrics?.unfulfilled_orders || 0, icon: '⏳', color: '#f59e0b' },
    { id: 'cancelled', label: 'Cancelled', value: metrics?.cancelled_orders || 0, icon: '✕', color: '#ef4444' },
  ];

  return (
    <div className="dashboard-container">
      <h2 className="section-title">Business Metrics</h2>
      <p className="section-subtitle">Snapshot of your performance this month</p>

      <div className="metrics-feed">
        {isLoading ? (
          Array(6).fill(0).map((_, i) => (
            <div key={i} className="skeleton-card" />
          ))
        ) : (
          metricCards.map(metric => (
            <MobileCard
              key={metric.id}
              onClick={() => setSelectedMetric(metric.id)}
              className="metric-feed-card"
            >
              <div className="metric-icon-circle" style={{ backgroundColor: metric.color + '20', color: metric.color }}>
                {metric.icon}
              </div>
              <div className="metric-content">
                <span className="metric-label">{metric.label}</span>
                <span className="metric-value">{metric.value}</span>
              </div>
              <div className="metric-arrow">→</div>
            </MobileCard>
          ))
        )}
      </div>

      <BottomSheet
        isOpen={!!selectedMetric}
        onClose={() => setSelectedMetric(null)}
        title={metricCards.find(m => m.id === selectedMetric)?.label}
      >
        <div className="metric-drilldown">
          {selectedMetric === 'gst' && (
            <div className="gst-breakdown-list">
              <div className="breakdown-item">
                <span>CGST (Central)</span>
                <strong>₹{metrics?.cgst_collected?.toLocaleString('en-IN') || 0}</strong>
              </div>
              <div className="breakdown-item">
                <span>SGST (State)</span>
                <strong>₹{metrics?.sgst_collected?.toLocaleString('en-IN') || 0}</strong>
              </div>
              <div className="breakdown-item">
                <span>IGST (Integrated)</span>
                <strong>₹{metrics?.igst_collected?.toLocaleString('en-IN') || 0}</strong>
              </div>
            </div>
          )}
          {selectedMetric === 'revenue' && (
            <div className="drilldown-stat">
              <p>Total processed invoices: <strong>{metrics?.total_invoices || 0}</strong></p>
              <p>Average invoice value: <strong>₹{metrics?.total_invoices ? Math.round(metrics.total_revenue / metrics.total_invoices).toLocaleString('en-IN') : 0}</strong></p>
            </div>
          )}
          {selectedMetric === 'orders' && (
            <div className="drilldown-stat">
              <p>Fulfillment rate: <strong>{metrics?.total_orders ? Math.round((metrics.fulfilled_orders / metrics.total_orders) * 100) : 0}%</strong></p>
              <p>Cancellation rate: <strong>{metrics?.total_orders ? Math.round((metrics.cancelled_orders / metrics.total_orders) * 100) : 0}%</strong></p>
            </div>
          )}
          <button className="primary-btn full-width" style={{marginTop: '2rem'}} onClick={() => setSelectedMetric(null)}>
            Done
          </button>
        </div>
      </BottomSheet>
    </div>
  );
};
