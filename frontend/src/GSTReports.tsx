import { API_BASE } from './api';
import React, { useState, useEffect, useCallback, useMemo } from 'react';

interface GSTSummary {
  total_orders: number;
  cancelled_orders: number;
  invoices_generated: number;
  total_revenue: number;
  total_taxable_value: number;
  total_gst_collected: number;
  total_igst: number;
  total_cgst: number;
  total_sgst: number;
  fulfilled_orders: number;
  unfulfilled_orders: number;
  paid_orders: number;
}

interface StateReport extends Record<string, string | number | null> {
  state: string;
  orders: number;
  taxable_value: number;
  igst: number;
  cgst: number;
  sgst: number;
  total_gst: number;
  revenue: number;
}

interface HSNReport extends Record<string, string | number | null> {
  hsn_code: string;
  product_count: number;
  qty_sold: number;
  taxable_value: number;
  igst: number;
  cgst: number;
  sgst: number;
  total_gst: number;
  revenue: number;
}

interface DocumentIssued extends Record<string, string | number | null> {
  document_type: string;
  from_serial: string;
  to_serial: string;
  total_issued: number;
  cancelled: number;
  net_issued: number;
}

interface OperationalSummary {
  metric: string;
  count: number | string;
}

interface GSTReportsProps {
  startDate: string;
  endDate: string;
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
  refreshTrigger?: number;
}


export const GSTReports: React.FC<GSTReportsProps> = ({ startDate, endDate, fetchWithAuth, refreshTrigger }) => {
  const [activeSubTab, setActiveSubTab] = useState<'summary' | 'state' | 'hsn' | 'documents'>('summary');
  const [summary, setSummary] = useState<GSTSummary | null>(null);
  const [stateData, setStateData] = useState<StateReport[] | null>(null);
  const [hsnData, setHsnData] = useState<HSNReport[] | null>(null);
  const [docsData, setDocsData] = useState<DocumentIssued[] | null>(null);
  const [opSummary, setOpSummary] = useState<OperationalSummary[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);

  // Sorting State
  const [sortField, setSortField] = useState<string>('');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');

  /**
   * Performance Optimization: Memoized sorting.
   * By using useCallback for the sorting logic and useMemo for the sorted results,
   * we prevent expensive O(N log N) sorting operations from running on every component re-render.
   */
  const handleSort = (field: string) => {
    if (sortField === field) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortOrder('desc');
    }
  };

  const getSortedData = useCallback(<T extends Record<string, string | number | null>>(data: T[]): T[] => {
    if (!sortField) return data;
    return [...data].sort((a, b) => {
      const aVal = a[sortField] ?? null;
      const bVal = b[sortField] ?? null;
      
      if (aVal === null || bVal === null) return 0;

      if (typeof aVal === 'string' && typeof bVal === 'string') {
        return sortOrder === 'asc' 
          ? aVal.localeCompare(bVal) 
          : bVal.localeCompare(aVal);
      }
      
      return sortOrder === 'asc' 
        ? (aVal as number) - (bVal as number) 
        : (bVal as number) - (aVal as number);
    });
  }, [sortField, sortOrder]);

  const sortedStateData = useMemo(() => getSortedData(stateData || []), [stateData, getSortedData]);
  const sortedHsnData = useMemo(() => getSortedData(hsnData || []), [hsnData, getSortedData]);

  const renderSortArrow = (field: string) => {
    if (sortField !== field) return <span style={{ opacity: 0.2, marginLeft: '4px' }}>↕</span>;
    return <span style={{ marginLeft: '4px', color: 'var(--accent-color)' }}>{sortOrder === 'asc' ? '↑' : '↓'}</span>;
  };

  // Clear stale data when global filters change
  useEffect(() => {
    setSummary(null);
    setStateData(null);
    setHsnData(null);
    setDocsData(null);
    setOpSummary(null);
    setLoading(true);
  }, [startDate, endDate, refreshTrigger]);

  useEffect(() => {
    const fetchReports = async () => {
      /**
       * Performance Optimization: On-demand (lazy) loading.
       * We only fetch data for the active tab if it hasn't been loaded yet for the current date range.
       * This reduces redundant API calls and initial network payload.
       */
      const needsSummary = activeSubTab === 'summary' && !summary;
      const needsState = activeSubTab === 'state' && !stateData;
      const needsHSN = activeSubTab === 'hsn' && !hsnData;
      const needsDocs = activeSubTab === 'documents' && !docsData;

      if (!needsSummary && !needsState && !needsHSN && !needsDocs) {
        setLoading(false);
        return;
      }

      if (summary || stateData || hsnData || docsData) {
        setIsRefreshing(true);
      } else {
        setLoading(true);
      }

      try {
        let startObj = '';
        let endObj = '';
        
        if (startDate) {
          const [y, m, d] = startDate.split('-').map(Number);
          startObj = new Date(y, m - 1, d, 0, 0, 0, 0).toISOString();
        }
        
        if (endDate) {
          const [y, m, d] = endDate.split('-').map(Number);
          endObj = new Date(y, m - 1, d, 23, 59, 59, 999).toISOString();
        }

        const fetchTasks: Promise<unknown>[] = [];

        if (needsSummary) {
          fetchTasks.push(
            fetchWithAuth(`${API_BASE}/api/reports/summary?start_date=${startObj}&end_date=${endObj}`)
              .then(res => res.json())
              .then(data => {
                if (data.success) {
                  setSummary(data.summary);
                  setOpSummary([
                    { metric: 'Total Orders', count: data.summary.total_orders },
                    { metric: 'Cancelled Orders', count: data.summary.cancelled_orders },
                    { metric: 'Fulfilled Orders', count: data.summary.fulfilled_orders },
                    { metric: 'Unfulfilled Orders', count: data.summary.unfulfilled_orders },
                    { metric: 'Paid Orders', count: data.summary.paid_orders },
                    { metric: 'Invoices Generated', count: data.summary.invoices_generated }
                  ]);
                }
              })
          );
        }

        if (needsState) {
          fetchTasks.push(
            fetchWithAuth(`${API_BASE}/api/reports/state-wise?start_date=${startObj}&end_date=${endObj}`)
              .then(res => res.json())
              .then(data => {
                if (data.success) setStateData(data.data || []);
              })
          );
        }

        if (needsHSN) {
          fetchTasks.push(
            fetchWithAuth(`${API_BASE}/api/reports/hsn-wise?start_date=${startObj}&end_date=${endObj}`)
              .then(res => res.json())
              .then(data => {
                if (data.success) setHsnData(data.data || []);
              })
          );
        }

        if (needsDocs) {
          fetchTasks.push(
            fetchWithAuth(`${API_BASE}/api/reports/documents-issued?start_date=${startObj}&end_date=${endObj}`)
              .then(res => res.json())
              .then(data => {
                if (data.success) setDocsData(data.data || []);
              })
          );
        }

        await Promise.all(fetchTasks);
      } catch (err) {
        console.error('Failed to fetch reports:', err);
      } finally {
        setLoading(false);
        setIsRefreshing(false);
      }
    };

    fetchReports();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [startDate, endDate, fetchWithAuth, refreshTrigger, activeSubTab]);

  const downloadCSV = (data: Record<string, string | number | null>[], filename: string) => {
    if (data.length === 0) return;
    const headers = Object.keys(data[0]).join(',');
    const rows = data.map(obj => Object.values(obj).join(',')).join('\n');
    const csvContent = `data:text/csv;charset=utf-8,${headers}\n${rows}`;
    const encodedUri = encodeURI(csvContent);
    const link = document.createElement('a');
    link.setAttribute('href', encodedUri);
    link.setAttribute('download', filename);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  if (loading) return <div style={{ padding: '2rem', textAlign: 'center' }}>Loading Reports...</div>;

  return (
    <div className="gst-reports-container" style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
      
      {/* Sub-navigation */}
      <div className="sub-tabs-header" style={{ 
        display: 'flex', 
        gap: '0.5rem', 
        borderBottom: '1px solid var(--border-color)', 
        marginBottom: '1rem',
        paddingBottom: '0.5rem', 
        overflowX: 'auto', 
        WebkitOverflowScrolling: 'touch',
        alignItems: 'center' 
      }}>
        {[
          { id: 'summary', label: 'Dashboard' },
          { id: 'state', label: 'B2C State-wise' },
          { id: 'hsn', label: 'HSN Summary' },
          { id: 'documents', label: 'Documents Issued' }
        ].map((tab) => (
          <button 
            key={tab.id}
            onClick={() => setActiveSubTab(tab.id as 'summary' | 'state' | 'hsn' | 'documents')}
            style={{ 
              background: 'none', border: 'none', padding: '0.5rem 1.25rem', cursor: 'pointer',
              borderBottom: activeSubTab === tab.id ? '2px solid var(--accent-color)' : 'none',
              color: activeSubTab === tab.id ? 'var(--accent-color)' : 'var(--text-secondary)',
              fontWeight: activeSubTab === tab.id ? 600 : 400,
              fontSize: '0.9rem',
              whiteSpace: 'nowrap'
            }}
          >
            {tab.label}
          </button>
        ))}
        {isRefreshing && (
          <div style={{ marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--accent-color)', fontSize: '0.8rem', fontWeight: 600 }}>
            <div className="dot-flashing"></div>
            Updating...
          </div>
        )}
      </div>

      {activeSubTab === 'summary' && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
          {/* Summary Metrics Cards */}
          <section className="dashboard-grid">
            <div className="card">
              <h3 className="card-title">Total Revenue</h3>
              <div className="card-value">₹{summary?.total_revenue.toLocaleString('en-IN') || '0'}</div>
            </div>
            <div className="card">
              <h3 className="card-title">Taxable Value</h3>
              <div className="card-value">₹{summary?.total_taxable_value.toLocaleString('en-IN') || '0'}</div>
            </div>
            <div className="card">
              <h3 className="card-title">GST Collected</h3>
              <div className="card-value">₹{summary?.total_gst_collected.toLocaleString('en-IN', { maximumFractionDigits: 2 }) || '0'}</div>
            </div>
            <div className="card" style={{ gridColumn: 'span 3' }}>
              <div className="report-card-stats" style={{ display: 'flex', justifyContent: 'space-between', gap: '1rem' }}>
                <div style={{ flex: 1 }}>
                  <h3 className="card-title" style={{ fontSize: '0.75rem', color: 'var(--accent-color)' }}>IGST</h3>
                  <div style={{ fontSize: '1.5rem', fontWeight: 700 }}>₹{summary?.total_igst.toLocaleString('en-IN', { maximumFractionDigits: 2 }) || '0'}</div>
                </div>
                <div style={{ flex: 1, borderLeft: '1px solid var(--border-color)', paddingLeft: '1.5rem' }}>
                  <h3 className="card-title" style={{ fontSize: '0.75rem', color: 'var(--accent-color)' }}>CGST</h3>
                  <div style={{ fontSize: '1.5rem', fontWeight: 700 }}>₹{summary?.total_cgst.toLocaleString('en-IN', { maximumFractionDigits: 2 }) || '0'}</div>
                </div>
                <div style={{ flex: 1, borderLeft: '1px solid var(--border-color)', paddingLeft: '1.5rem' }}>
                  <h3 className="card-title" style={{ fontSize: '0.75rem', color: 'var(--accent-color)' }}>SGST</h3>
                  <div style={{ fontSize: '1.5rem', fontWeight: 700 }}>₹{summary?.total_sgst.toLocaleString('en-IN', { maximumFractionDigits: 2 }) || '0'}</div>
                </div>
              </div>
            </div>
          </section>
          
          {/* Operational Status Summary */}
          <section className="table-container" style={{ maxWidth: '400px' }}>
            <div className="table-header">
              <h3>Operational Summary</h3>
            </div>
            <table>
              <thead>
                <tr>
                  <th>Metric</th>
                  <th>Count</th>
                </tr>
              </thead>
              <tbody>
                {opSummary?.map((row, idx) => (
                  <tr key={idx}>
                    <td>{row.metric}</td>
                    <td>{row.count}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </section>
        </div>
      )}

      {activeSubTab === 'state' && (
        <section className="table-container">
          <div className="table-header">
            <h3>B2C State-wise GST Summary</h3>
            <button className="btn-secondary" onClick={() => downloadCSV(stateData || [], 'state_gst_report.csv')}>Export CSV</button>
          </div>
          <div style={{ overflowX: 'auto' }}>
            <table>
              <thead>
                <tr>
                  <th style={{ cursor: 'pointer' }} onClick={() => handleSort('state')}>State {renderSortArrow('state')}</th>
                  <th style={{ cursor: 'pointer' }} onClick={() => handleSort('orders')}>Orders {renderSortArrow('orders')}</th>
                  <th style={{ cursor: 'pointer' }} onClick={() => handleSort('taxable_value')}>Taxable Value {renderSortArrow('taxable_value')}</th>
                  <th style={{ cursor: 'pointer' }} onClick={() => handleSort('igst')}>IGST {renderSortArrow('igst')}</th>
                  <th style={{ cursor: 'pointer' }} onClick={() => handleSort('cgst')}>CGST {renderSortArrow('cgst')}</th>
                  <th style={{ cursor: 'pointer' }} onClick={() => handleSort('sgst')}>SGST {renderSortArrow('sgst')}</th>
                  <th style={{ cursor: 'pointer' }} onClick={() => handleSort('total_gst')}>Total GST {renderSortArrow('total_gst')}</th>
                  <th style={{ cursor: 'pointer' }} onClick={() => handleSort('revenue')}>Revenue {renderSortArrow('revenue')}</th>
                </tr>
              </thead>
              <tbody>
                {!sortedStateData || sortedStateData.length === 0 ? (
                   <tr>
                    <td colSpan={8} style={{ textAlign: 'center', padding: '2rem' }}>No state-wise data for this period.</td>
                  </tr>
                ) : (
                  sortedStateData.map((row, idx) => (
                    <tr key={idx}>
                      <td>{row.state || 'N/A'}</td>
                      <td>{row.orders}</td>
                      <td>₹{row.taxable_value.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</td>
                      <td>₹{row.igst.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</td>
                      <td>₹{row.cgst.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</td>
                      <td>₹{row.sgst.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</td>
                      <td>₹{row.total_gst.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</td>
                      <td>₹{row.revenue.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </section>
      )}

      {activeSubTab === 'hsn' && (
        <section className="table-container">
          <div className="table-header">
            <h3>HSN Summary</h3>
            <button className="btn-secondary" onClick={() => downloadCSV(hsnData || [], 'hsn_summary_report.csv')}>Export CSV</button>
          </div>
          <div style={{ overflowX: 'auto' }}>
            <table>
              <thead>
                <tr>
                  <th style={{ cursor: 'pointer' }} onClick={() => handleSort('hsn_code')}>HSN Code {renderSortArrow('hsn_code')}</th>
                  <th style={{ cursor: 'pointer' }} onClick={() => handleSort('qty_sold')}>Qty {renderSortArrow('qty_sold')}</th>
                  <th style={{ cursor: 'pointer' }} onClick={() => handleSort('taxable_value')}>Taxable {renderSortArrow('taxable_value')}</th>
                  <th style={{ cursor: 'pointer' }} onClick={() => handleSort('igst')}>IGST {renderSortArrow('igst')}</th>
                  <th style={{ cursor: 'pointer' }} onClick={() => handleSort('cgst')}>CGST {renderSortArrow('cgst')}</th>
                  <th style={{ cursor: 'pointer' }} onClick={() => handleSort('sgst')}>SGST {renderSortArrow('sgst')}</th>
                  <th style={{ cursor: 'pointer' }} onClick={() => handleSort('total_gst')}>Total GST {renderSortArrow('total_gst')}</th>
                </tr>
              </thead>
              <tbody>
                {!sortedHsnData || sortedHsnData.length === 0 ? (
                   <tr>
                    <td colSpan={7} style={{ textAlign: 'center', padding: '2rem' }}>No HSN data for this period.</td>
                  </tr>
                ) : (
                  sortedHsnData.map((row, idx) => (
                    <tr key={idx}>
                      <td>{row.hsn_code}</td>
                      <td>{row.qty_sold}</td>
                      <td>₹{row.taxable_value.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</td>
                      <td>₹{row.igst.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</td>
                      <td>₹{row.cgst.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</td>
                      <td>₹{row.sgst.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</td>
                      <td>₹{row.total_gst.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </section>
      )}

      {activeSubTab === 'documents' && (
        <section className="table-container">
          <div className="table-header">
            <h3>Documents Issued (GSTR-1 Table 13)</h3>
            <button className="btn-secondary" onClick={() => downloadCSV(docsData || [], 'documents_issued_report.csv')}>Export CSV</button>
          </div>
          <div style={{ overflowX: 'auto' }}>
            <table>
              <thead>
                <tr>
                  <th>Document Type</th>
                  <th>From Serial</th>
                  <th>To Serial</th>
                  <th>Total Issued</th>
                  <th>Cancelled</th>
                  <th>Net Issued</th>
                </tr>
              </thead>
              <tbody>
                {!docsData || docsData.length === 0 ? (
                   <tr>
                    <td colSpan={6} style={{ textAlign: 'center', padding: '2rem' }}>No documents issued in this period.</td>
                  </tr>
                ) : (
                  docsData.map((row, idx) => (
                    <tr key={idx}>
                      <td>{row.document_type}</td>
                      <td>{row.from_serial}</td>
                      <td>{row.to_serial}</td>
                      <td>{row.total_issued}</td>
                      <td>{row.cancelled}</td>
                      <td>{row.net_issued}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </section>
      )}
    </div>
  );
};
