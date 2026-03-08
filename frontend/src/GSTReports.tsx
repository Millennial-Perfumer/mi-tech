import React, { useState, useEffect } from 'react';

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

interface StateReport {
  state: string;
  orders: number;
  taxable_value: number;
  igst: number;
  cgst: number;
  sgst: number;
  total_gst: number;
  revenue: number;
}

interface HSNReport {
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

interface DocumentIssued {
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
}

export const GSTReports: React.FC<GSTReportsProps> = ({ startDate, endDate }) => {
  const [activeSubTab, setActiveSubTab] = useState<'summary' | 'state' | 'hsn' | 'documents'>('summary');
  const [summary, setSummary] = useState<GSTSummary | null>(null);
  const [stateData, setStateData] = useState<StateReport[]>([]);
  const [hsnData, setHsnData] = useState<HSNReport[]>([]);
  const [docsData, setDocsData] = useState<DocumentIssued[]>([]);
  const [opSummary, setOpSummary] = useState<OperationalSummary[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchReports = async () => {
      setLoading(true);
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

        const [sumRes, stateRes, hsnRes, docsRes] = await Promise.all([
          fetch(`http://localhost:8080/api/reports/summary?start_date=${startObj}&end_date=${endObj}`),
          fetch(`http://localhost:8080/api/reports/state-wise?start_date=${startObj}&end_date=${endObj}`),
          fetch(`http://localhost:8080/api/reports/hsn-wise?start_date=${startObj}&end_date=${endObj}`),
          fetch(`http://localhost:8080/api/reports/documents-issued?start_date=${startObj}&end_date=${endObj}`)
        ]);

        const sumData = await sumRes.json();
        const sData = await stateRes.json();
        const hData = await hsnRes.json();
        const dData = await docsRes.json();

        if (sumData.success) {
          setSummary(sumData.summary);
          setOpSummary([
            { metric: 'Total Orders', count: sumData.summary.total_orders },
            { metric: 'Cancelled Orders', count: sumData.summary.cancelled_orders },
            { metric: 'Fulfilled Orders', count: sumData.summary.fulfilled_orders },
            { metric: 'Unfulfilled Orders', count: sumData.summary.unfulfilled_orders },
            { metric: 'Paid Orders', count: sumData.summary.paid_orders },
            { metric: 'Invoices Generated', count: sumData.summary.invoices_generated }
          ]);
        }
        if (sData.success) setStateData(sData.data || []);
        if (hData.success) setHsnData(hData.data || []);
        if (dData.success) setDocsData(dData.data || []);
      } catch (err) {
        console.error('Failed to fetch reports:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchReports();
  }, [startDate, endDate]);

  const downloadCSV = (data: any[], filename: string) => {
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
      <div style={{ display: 'flex', gap: '0.5rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '0.5rem', overflowX: 'auto' }}>
        {[
          { id: 'summary', label: 'Dashboard' },
          { id: 'state', label: 'B2C State-wise' },
          { id: 'hsn', label: 'HSN Summary' },
          { id: 'documents', label: 'Documents Issued' }
        ].map((tab) => (
          <button 
            key={tab.id}
            onClick={() => setActiveSubTab(tab.id as any)}
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
              <div style={{ display: 'flex', justifyContent: 'space-between', gap: '1rem' }}>
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
                {opSummary.map((row, idx) => (
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
            <button className="btn-secondary" onClick={() => downloadCSV(stateData, 'state_gst_report.csv')}>Export CSV</button>
          </div>
          <div style={{ overflowX: 'auto' }}>
            <table>
              <thead>
                <tr>
                  <th>State</th>
                  <th>Orders</th>
                  <th>Taxable Value</th>
                  <th>IGST</th>
                  <th>CGST</th>
                  <th>SGST</th>
                  <th>Total GST</th>
                  <th>Revenue</th>
                </tr>
              </thead>
              <tbody>
                {stateData.length === 0 ? (
                   <tr>
                    <td colSpan={8} style={{ textAlign: 'center', padding: '2rem' }}>No state-wise data for this period.</td>
                  </tr>
                ) : (
                  stateData.map((row, idx) => (
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
            <button className="btn-secondary" onClick={() => downloadCSV(hsnData, 'hsn_summary_report.csv')}>Export CSV</button>
          </div>
          <div style={{ overflowX: 'auto' }}>
            <table>
              <thead>
                <tr>
                  <th>HSN Code</th>
                  <th>Qty</th>
                  <th>Taxable</th>
                  <th>IGST</th>
                  <th>CGST</th>
                  <th>SGST</th>
                  <th>Total GST</th>
                </tr>
              </thead>
              <tbody>
                {hsnData.length === 0 ? (
                   <tr>
                    <td colSpan={7} style={{ textAlign: 'center', padding: '2rem' }}>No HSN data for this period.</td>
                  </tr>
                ) : (
                  hsnData.map((row, idx) => (
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
            <button className="btn-secondary" onClick={() => downloadCSV(docsData, 'documents_issued_report.csv')}>Export CSV</button>
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
                {docsData.length === 0 ? (
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
