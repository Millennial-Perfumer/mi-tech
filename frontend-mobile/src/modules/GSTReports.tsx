import React, { useState, useEffect, useCallback } from 'react';
import { API_BASE, fetchWithAuth, getTodayIST } from '../api';
import { MobileCard } from '../components/MobileCard';
import { BottomSheet } from '../components/BottomSheet';
import './GSTReports.css';

interface GSTReport {
  month: string;
  total_revenue: number;
  total_gst: number;
  cgst: number;
  sgst: number;
  igst: number;
  order_count: number;
}

export const GSTReports: React.FC = () => {
  const [reports, setReports] = useState<GSTReport[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isDatePickerOpen, setIsDatePickerOpen] = useState(false);
  const [dateRange, setDateRange] = useState({
    start: new Date(new Date().getFullYear(), 0, 1).toISOString().split('T')[0],
    end: getTodayIST()
  });

  const fetchGSTData = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await fetchWithAuth(`${API_BASE}/api/reports/gst?start_date=${dateRange.start}T00:00:00Z&end_date=${dateRange.end}T23:59:59Z`);
      const data = await res.json();
      if (data.success) {
        setReports(Array.isArray(data.reports) ? data.reports : [data.report]);
      }
    } catch (err) {
      console.error('Error fetching GST data:', err);
    } finally {
      setIsLoading(false);
    }
  }, [dateRange]);

  useEffect(() => {
    void fetchGSTData();
  }, [fetchGSTData]);

  const handleExport = async (type: 'pdf' | 'excel') => {
    try {
      const res = await fetchWithAuth(`${API_BASE}/api/reports/export?type=${type}&start_date=${dateRange.start}&end_date=${dateRange.end}`);
      if (res.ok) {
        const blob = await res.blob();
        const url = window.URL.createObjectURL(blob);

        /* eslint-disable @typescript-eslint/no-explicit-any */
        const navigatorAny = navigator as any;
        if (navigatorAny.share) {
          const file = new File([blob], `gst_report_${dateRange.start}_${dateRange.end}.${type === 'pdf' ? 'pdf' : 'xlsx'}`, { type: blob.type });
          try {
            await navigatorAny.share({
              files: [file],
              title: 'GST Report',
              text: 'Here is your GST report snapshot.'
            });
          } catch (err) {
            console.error('Share failed:', err);
            const a = document.createElement('a');
            a.href = url;
            a.download = file.name;
            a.click();
          }
        } else {
          const a = document.createElement('a');
          a.href = url;
          a.download = `gst_report_${dateRange.start}_${dateRange.end}.${type === 'pdf' ? 'pdf' : 'xlsx'}`;
          a.click();
        }
      }
    } catch (err) {
      console.error('Export failed:', err);
    }
  };

  return (
    <div className="gst-container">
      <div className="gst-header">
        <h2 className="section-title">GST Engine</h2>
        <button className="date-pill-btn" onClick={() => setIsDatePickerOpen(true)}>
          🗓️ {dateRange.start} - {dateRange.end}
        </button>
      </div>

      <div className="reports-snapshots">
        {isLoading ? (
          <div className="loader">Calculating snapshots...</div>
        ) : reports.length === 0 ? (
          <div className="empty-state">No data for this period.</div>
        ) : (
          reports.map((report, idx) => (
            <MobileCard key={idx} className="snapshot-card">
              <div className="snapshot-header">
                <span className="snapshot-month">{report.month || 'Selected Period'}</span>
                <span className="snapshot-orders">{report.order_count} Invoices</span>
              </div>

              <div className="snapshot-main">
                <div className="snapshot-stat">
                  <label>Total GST</label>
                  <strong>₹{report.total_gst?.toLocaleString('en-IN')}</strong>
                </div>
                <div className="snapshot-stat secondary">
                  <label>Revenue</label>
                  <strong>₹{report.total_revenue?.toLocaleString('en-IN')}</strong>
                </div>
              </div>

              <div className="snapshot-breakdown">
                <div className="b-item"><span>CGST</span> <span>₹{report.cgst?.toLocaleString('en-IN')}</span></div>
                <div className="b-item"><span>SGST</span> <span>₹{report.sgst?.toLocaleString('en-IN')}</span></div>
                <div className="b-item"><span>IGST</span> <span>₹{report.igst?.toLocaleString('en-IN')}</span></div>
              </div>

              <div className="snapshot-actions">
                <button className="export-btn pdf" onClick={() => handleExport('pdf')}>📄 PDF</button>
                <button className="export-btn excel" onClick={() => handleExport('excel')}>📊 Excel</button>
              </div>
            </MobileCard>
          ))
        )}
      </div>

      <BottomSheet isOpen={isDatePickerOpen} onClose={() => setIsDatePickerOpen(false)} title="Select Date Range">
        <div className="date-picker-form">
          <div className="form-group">
            <label>From</label>
            <input
              type="date"
              value={dateRange.start}
              onChange={e => setDateRange(prev => ({ ...prev, start: e.target.value }))}
            />
          </div>
          <div className="form-group">
            <label>To</label>
            <input
              type="date"
              value={dateRange.end}
              onChange={e => setDateRange(prev => ({ ...prev, end: e.target.value }))}
            />
          </div>
          <button className="primary-btn full-width" onClick={() => setIsDatePickerOpen(false)}>
            Update Snapshots
          </button>
        </div>
      </BottomSheet>
    </div>
  );
};
