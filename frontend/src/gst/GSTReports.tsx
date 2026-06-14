import { API_BASE } from '../api';
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
  businessGstin?: string;
}


export const GSTReports: React.FC<GSTReportsProps> = ({ startDate, endDate, fetchWithAuth, refreshTrigger, businessGstin }) => {
  const [activeSubTab, setActiveSubTab] = useState<'summary' | 'state' | 'hsn' | 'documents' | 'gstr1'>('summary');
  const [summary, setSummary] = useState<GSTSummary | null>(null);
  const [stateData, setStateData] = useState<StateReport[] | null>(null);
  const [hsnData, setHsnData] = useState<HSNReport[] | null>(null);
  const [docsData, setDocsData] = useState<DocumentIssued[] | null>(null);
  const [opSummary, setOpSummary] = useState<OperationalSummary[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);

  const [gstin, setGstin] = useState(businessGstin || '33AUSPR1909H1ZC');
  const [isExporting, setIsExporting] = useState(false);

  const [selectedMonth, setSelectedMonth] = useState(() => {
    if (endDate) {
      const parts = endDate.split('-');
      if (parts.length >= 2) {
        return `${parts[0]}-${parts[1]}`;
      }
    }
    const today = new Date();
    const mm = today.getMonth() + 1;
    return `${today.getFullYear()}-${mm < 10 ? '0' + mm : mm}`;
  });

  useEffect(() => {
    if (endDate) {
      const parts = endDate.split('-');
      if (parts.length >= 2) {
        setSelectedMonth(`${parts[0]}-${parts[1]}`);
      }
    }
  }, [endDate]);

  const monthOptions = useMemo(() => {
    const options = [];
    const date = new Date();
    const monthNames = [
      'January', 'February', 'March', 'April', 'May', 'June',
      'July', 'August', 'September', 'October', 'November', 'December'
    ];
    
    for (let i = 0; i < 24; i++) {
      const year = date.getFullYear();
      const monthNum = date.getMonth() + 1; // 1-indexed
      const monthStr = monthNum < 10 ? `0${monthNum}` : `${monthNum}`;
      const value = `${year}-${monthStr}`;
      const label = `${monthNames[date.getMonth()]} ${year} (${monthStr}${year})`;
      options.push({ value, label });
      
      // Move to previous month
      date.setMonth(date.getMonth() - 1);
    }
    return options;
  }, []);


  useEffect(() => {
    if (businessGstin) {
      setGstin(businessGstin);
    }
  }, [businessGstin]);

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

  const handleExportGSTR1JSON = async () => {
    if (!gstin.trim()) {
      alert('Please enter a valid GSTIN.');
      return;
    }
    
    setIsExporting(true);
    try {
      const [yearStr, monthStr] = selectedMonth.split('-');
      const year = parseInt(yearStr, 10);
      const month = parseInt(monthStr, 10);

      const startObj = new Date(year, month - 1, 1, 0, 0, 0, 0).toISOString();
      const endObj = new Date(year, month, 0, 23, 59, 59, 999).toISOString();

      const response = await fetchWithAuth(
        `${API_BASE}/api/reports/gstr1-json?start_date=${startObj}&end_date=${endObj}&gstin=${gstin.trim()}`
      );

      if (!response.ok) {
        const errMsg = await response.text();
        throw new Error(errMsg || `HTTP ${response.status}`);
      }

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      
      const disposition = response.headers.get('content-disposition');
      let filename = `GSTR1_${gstin.trim()}.json`;
      if (disposition && disposition.indexOf('attachment') !== -1) {
        const filenameRegex = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/;
        const matches = filenameRegex.exec(disposition);
        if (matches != null && matches[1]) { 
          filename = matches[1].replace(/['"]/g, '');
        }
      }
      
      link.setAttribute('download', filename);
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (err: any) {
      console.error('Failed to export GSTR-1 JSON:', err);
      alert(`Failed to export GSTR-1 JSON: ${err.message || err}`);
    } finally {
      setIsExporting(false);
    }
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
          { id: 'documents', label: 'Documents Issued' },
          { id: 'gstr1', label: 'GSTR-1 Export' }
        ].map((tab) => (
          <button 
            key={tab.id}
            onClick={() => setActiveSubTab(tab.id as 'summary' | 'state' | 'hsn' | 'documents' | 'gstr1')}
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
        <div style={{ marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: '0.75rem', paddingLeft: '1rem' }}>
          {isRefreshing && (
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--accent-color)', fontSize: '0.8rem', fontWeight: 600, marginRight: '0.5rem' }}>
              <div className="dot-flashing"></div>
              Updating...
            </div>
          )}
        </div>
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

      {activeSubTab === 'gstr1' && (
        <div className="tab-content-fade" style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: '2rem', marginTop: '1rem' }}>
          {/* Left Column: Generation Controls */}
          <div className="glass-card-premium" style={{ padding: '2.5rem', display: 'flex', flexDirection: 'column', gap: '2rem', minHeight: '380px', justifyContent: 'space-between' }}>
            <div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '0.75rem' }}>
                <div style={{
                  background: 'var(--card-gradient-1)',
                  borderRadius: '12px',
                  width: '40px',
                  height: '40px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  color: 'white',
                  boxShadow: '0 4px 12px rgba(99, 102, 241, 0.2)'
                }}>
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                    <polyline points="14 2 14 8 20 8"></polyline>
                    <line x1="16" y1="13" x2="8" y2="13"></line>
                    <line x1="16" y1="17" x2="8" y2="17"></line>
                    <polyline points="10 9 9 9 8 9"></polyline>
                  </svg>
                </div>
                <h3 style={{ fontSize: '1.25rem', fontWeight: 700, margin: 0 }}>GSTR-1 JSON Compiler</h3>
              </div>
              <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', lineHeight: '1.5', margin: '0 0 1.5rem 0' }}>
                Compile and download the official GSTR-1 offline utility schema. You can upload this directly on the GST portal to auto-fill sales data.
              </p>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', background: 'rgba(255, 255, 255, 0.02)', border: '1px solid var(--border-color)', borderRadius: '12px', padding: '1.25rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px solid var(--border-color)', paddingBottom: '0.75rem' }}>
                  <span style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', fontWeight: 500 }}>Filing GSTIN</span>
                  <span style={{
                    fontFamily: 'monospace',
                    fontSize: '0.95rem',
                    fontWeight: 700,
                    color: 'var(--text-primary)',
                    background: 'var(--bg-input)',
                    padding: '0.25rem 0.75rem',
                    borderRadius: '6px',
                    letterSpacing: '0.05em'
                  }}>
                    {gstin || 'Not Configured'}
                  </span>
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', marginTop: '0.25rem' }}>
                  <label style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', fontWeight: 500 }}>Select Return Filing Month</label>
                  <select
                    value={selectedMonth}
                    onChange={(e) => setSelectedMonth(e.target.value)}
                    style={{
                      background: 'var(--bg-input)',
                      border: '1px solid var(--border-color)',
                      borderRadius: '8px',
                      color: 'var(--text-primary)',
                      fontWeight: 600,
                      padding: '0.6rem 0.75rem',
                      fontSize: '0.9rem',
                      cursor: 'pointer',
                      outline: 'none',
                      width: '100%',
                      fontFamily: 'inherit'
                    }}
                  >
                    {monthOptions.map((opt) => (
                      <option key={opt.value} value={opt.value} style={{ background: 'var(--surface-color)', color: 'var(--text-primary)' }}>
                        {opt.label}
                      </option>
                    ))}
                  </select>
                </div>
              </div>
            </div>

            <button
              onClick={handleExportGSTR1JSON}
              disabled={isExporting}
              style={{
                width: '100%',
                background: 'linear-gradient(135deg, #6366f1 0%, #4f46e5 100%)',
                border: 'none',
                color: 'white',
                borderRadius: '12px',
                padding: '1rem',
                fontSize: '1rem',
                fontWeight: 600,
                boxShadow: '0 4px 15px rgba(99, 102, 241, 0.3)',
                cursor: 'pointer',
                transition: 'all 0.2s ease',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '0.75rem'
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.transform = 'translateY(-2px)';
                e.currentTarget.style.boxShadow = '0 6px 20px rgba(99, 102, 241, 0.4)';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.transform = 'translateY(0)';
                e.currentTarget.style.boxShadow = '0 4px 15px rgba(99, 102, 241, 0.3)';
              }}
            >
              {isExporting ? (
                <>
                  <svg className="animate-spin" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" style={{ animation: 'spin 1s linear infinite' }}>
                    <circle cx="12" cy="12" r="10" stroke="currentColor" strokeOpacity="0.25"></circle>
                    <path d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" fill="currentColor"></path>
                  </svg>
                  Compiling Data...
                </>
              ) : (
                <>
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                    <polyline points="7 10 12 15 17 10"></polyline>
                    <line x1="12" y1="15" x2="12" y2="3"></line>
                  </svg>
                  Generate GSTR-1 JSON
                </>
              )}
            </button>
          </div>

          {/* Right Column: Compiled Sections Summary */}
          <div className="glass-card-premium" style={{ padding: '2.5rem', display: 'flex', flexDirection: 'column', gap: '1.5rem', minHeight: '380px' }}>
            <h3 style={{ fontSize: '1.15rem', fontWeight: 700, margin: 0, display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <span style={{ color: 'var(--accent-color)' }}>✓</span> Included GSTR-1 Sections
            </h3>
            
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
              <div style={{ display: 'flex', gap: '0.75rem' }}>
                <div style={{
                  color: 'var(--accent-color)',
                  background: 'var(--accent-subtle)',
                  borderRadius: '50%',
                  width: '24px',
                  height: '24px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontWeight: 700,
                  fontSize: '0.8rem',
                  flexShrink: 0
                }}>7</div>
                <div>
                  <h4 style={{ fontSize: '0.95rem', fontWeight: 600, margin: '0 0 0.25rem 0' }}>B2CS (B2C Small) Sales</h4>
                  <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', margin: 0, lineHeight: '1.4' }}>
                    Aggregates all state-wise retail sales made to unregistered buyers. These are divided by rate and place of supply.
                  </p>
                </div>
              </div>

              <div style={{ display: 'flex', gap: '0.75rem' }}>
                <div style={{
                  color: 'var(--accent-color)',
                  background: 'var(--accent-subtle)',
                  borderRadius: '50%',
                  width: '24px',
                  height: '24px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontWeight: 700,
                  fontSize: '0.75rem',
                  flexShrink: 0
                }}>12</div>
                <div>
                  <h4 style={{ fontSize: '0.95rem', fontWeight: 600, margin: '0 0 0.25rem 0' }}>HSN Summary of Outward Supplies</h4>
                  <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', margin: 0, lineHeight: '1.4' }}>
                    Lists quantities, taxable values, and tax bifurcation group-by-product HSN / SAC code.
                  </p>
                </div>
              </div>

              <div style={{ display: 'flex', gap: '0.75rem' }}>
                <div style={{
                  color: 'var(--accent-color)',
                  background: 'var(--accent-subtle)',
                  borderRadius: '50%',
                  width: '24px',
                  height: '24px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontWeight: 700,
                  fontSize: '0.75rem',
                  flexShrink: 0
                }}>13</div>
                <div>
                  <h4 style={{ fontSize: '0.95rem', fontWeight: 600, margin: '0 0 0.25rem 0' }}>Documents Issued</h4>
                  <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', margin: 0, lineHeight: '1.4' }}>
                    Declares sequential ranges of serial numbers issued for tax invoices and credit notes, specifying count and cancellations.
                  </p>
                </div>
              </div>
            </div>

            <div style={{
              marginTop: 'auto',
              padding: '1rem',
              background: 'var(--bg-hover)',
              borderRadius: '8px',
              borderLeft: '4px solid var(--accent-color)',
              fontSize: '0.8rem',
              color: 'var(--text-secondary)',
              lineHeight: '1.4'
            }}>
              <strong>Filing Instruction:</strong> Login to the <a href="https://www.gst.gov.in" target="_blank" rel="noopener noreferrer" style={{ color: 'var(--accent-color)', textDecoration: 'underline' }}>GST Portal</a>, navigate to <strong>Returns Dashboard &gt; GSTR-1 &gt; Prepare Offline</strong>, and upload this generated JSON file.
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
