import { useState, useEffect } from 'react';
import { API_BASE } from './api';

interface Customer {
    id: number;
    phone_number: string;
    first_name: string;
    last_name: string;
    email: string;
    address1: string;
    city: string;
    state: string;
    total_orders: number;
    total_spent: number;
    updated_at: string;
    source_id: string;
}

interface Source {
    id: string;
    name: string;
}

interface CustomersProps {
    fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
    showClearButton?: boolean;
}

type ColumnKey = 'name' | 'phone' | 'email' | 'location' | 'orders' | 'spent' | 'activity' | 'source';

interface ColumnDef {
    key: ColumnKey;
    label: string;
}

const ALL_COLUMNS: ColumnDef[] = [
    { key: 'name', label: 'Name' },
    { key: 'phone', label: 'Phone' },
    { key: 'email', label: 'Email' },
    { key: 'location', label: 'Location' },
    { key: 'orders', label: 'Orders' },
    { key: 'spent', label: 'Total Spent' },
    { key: 'activity', label: 'Last Activity' },
    { key: 'source', label: 'Source' },
];

export function Customers({ fetchWithAuth, showClearButton = false }: CustomersProps) {
    const [file, setFile] = useState<File | null>(null);
    const [customers, setCustomers] = useState<Customer[]>([]);
    const [total, setTotal] = useState(0);
    const [page, setPage] = useState(1);
    const [search, setSearch] = useState('');
    const [isLoading, setIsLoading] = useState(true);
    const [isImporting, setIsImporting] = useState(false);
    const [showImportModal, setShowImportModal] = useState(false);
    const [showColumnPicker, setShowColumnPicker] = useState(false);
    const [visibleColumns, setVisibleColumns] = useState<ColumnKey[]>(() => {
        const saved = localStorage.getItem('customer_columns');
        return saved ? JSON.parse(saved) : ['name', 'phone', 'location', 'orders', 'spent', 'activity'];
    });
    const [sortBy, setSortBy] = useState<ColumnKey>('activity');
    const [sortOrder, setSortOrder] = useState<'ASC' | 'DESC'>('DESC');
    const [sources, setSources] = useState<Source[]>([]);
    const [selectedSource, setSelectedSource] = useState<string>('');
    const [isDeleting, setIsDeleting] = useState(false);
    const limit = 25;

    useEffect(() => {
        localStorage.setItem('customer_columns', JSON.stringify(visibleColumns));
    }, [visibleColumns]);

    const toggleColumn = (key: ColumnKey) => {
        setVisibleColumns(prev => 
            prev.includes(key) ? prev.filter(k => k !== key) : [...prev, key]
        );
    };

    const handleSort = (key: ColumnKey) => {
        if (sortBy === key) {
            setSortOrder(prev => prev === 'ASC' ? 'DESC' : 'ASC');
        } else {
            setSortBy(key);
            setSortOrder('DESC');
        }
        setPage(1);
    };

    const fetchCustomers = async () => {
        setIsLoading(true);
        try {
            const response = await fetchWithAuth(`${API_BASE}/api/customers?page=${page}&pageSize=${limit}&search=${search}&sortBy=${sortBy}&sortOrder=${sortOrder}`);
            const data = await response.json();
            setCustomers(data.customers || []);
            setTotal(data.total || 0);
        } catch (error) {
            console.error('Error fetching customers:', error);
        } finally {
            setIsLoading(false);
        }
    };

    const fetchSources = async () => {
        try {
            const response = await fetchWithAuth(`${API_BASE}/api/sources`);
            if (!response.ok) {
                return;
            }
            const data = await response.json();
            if (data.success) {
                const fetchedSources = data.sources || [];
                setSources(fetchedSources);
                if (fetchedSources.length > 0 && !selectedSource) {
                    setSelectedSource(fetchedSources[0].id);
                }
            } else {
                console.error('API Error fetching sources:', data.message);
            }
        } catch (error) {
            console.error('Error fetching sources:', error);
        }
    };

    useEffect(() => {
        fetchSources();
    }, []);

    useEffect(() => {
        const timer = setTimeout(() => {
            fetchCustomers();
        }, 300);
        return () => clearTimeout(timer);
    }, [page, search, sortBy, sortOrder]);

    const handleImport = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        const formData = new FormData(e.currentTarget);
        const file = formData.get('file') as File;
        if (!file) return;

        setIsImporting(true);
        try {
            const response = await fetchWithAuth(`${API_BASE}/api/customers/import`, {
                method: 'POST',
                body: formData,
            });
            if (response.ok) {
                alert('Import successful!');
                setFile(null);
                setShowImportModal(false);
                fetchCustomers();
            } else {
                const errorData = await response.json();
                alert('Import failed: ' + (errorData.message || 'Unknown error'));
            }
        } catch (error) {
            console.error('Error importing customers:', error);
            alert('Error during import.');
        } finally {
            setIsImporting(false);
        }
    };

    const handleDeleteAll = async () => {
        if (!window.confirm('Are you absolutely sure? This will permanently delete ALL customers from the database.')) {
            return;
        }

        setIsDeleting(true);
        try {
            const response = await fetchWithAuth(`${API_BASE}/api/customers`, {
                method: 'DELETE',
            });
            if (response.ok) {
                alert('All customers cleared successfully.');
                fetchCustomers();
            } else {
                alert('Failed to clear customers.');
            }
        } catch (error) {
            console.error('Error deleting customers:', error);
            alert('Error during deletion.');
        } finally {
            setIsDeleting(false);
        }
    };

    return (
        <div className="tab-pane active">
            <div style={{ 
                display: 'flex', 
                justifyContent: 'space-between', 
                alignItems: 'center', 
                marginBottom: '2rem',
                padding: '1.25rem 1.5rem',
                background: 'white',
                borderRadius: '16px',
                boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.05), 0 2px 4px -1px rgba(0, 0, 0, 0.03)',
                border: '1px solid #f1f5f9'
            }}>
                <div>
                    <h1 style={{ margin: 0, fontSize: '1.5rem', fontWeight: 800, color: '#0f172a', letterSpacing: '-0.025em' }}>Customer Directory</h1>
                    <p style={{ margin: '4px 0 0 0', color: '#64748b', fontSize: '0.9rem', fontWeight: 500 }}>Manage and analyze your customer base</p>
                </div>
                <div style={{ display: 'flex', gap: '1rem' }}>
                    <button className="btn-secondary" style={{ padding: '0.5rem 1rem', fontSize: '0.875rem', display: 'flex', alignItems: 'center', gap: '8px' }}>
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
                        Export Data
                    </button>
                    <button className="btn-secondary" onClick={handleDeleteAll} disabled={isDeleting} style={{ padding: '0.5rem 1rem', fontSize: '0.875rem', display: 'flex', alignItems: 'center', gap: '8px', color: '#dc2626', borderColor: '#fee2e2', background: '#fff' }}>
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>
                        {isDeleting ? 'Clearing...' : 'Clear All'}
                    </button>
                    <button className="btn-primary" onClick={() => setShowImportModal(true)} style={{ padding: '0.5rem 1rem', fontSize: '0.875rem', display: 'flex', alignItems: 'center', gap: '8px' }}>
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
                        Import Customers
                    </button>
                </div>
            </div>

            <div className="stats-grid" style={{ marginBottom: '2rem' }}>
                <div className="stat-card">
                    <div className="stat-icon-wrapper primary">
                        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>
                    </div>
                    <div className="stat-content">
                        <div className="stat-label">Total Customers</div>
                        <div className="stat-value">{total || 0}</div>
                    </div>
                </div>
            </div>

            <div className="card">
                <div className="table-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <div style={{ flex: 1, position: 'relative', marginRight: '2rem' }}>
                        <svg style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: '#94a3b8' }} width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
                        <input 
                            type="text" 
                            placeholder="Search by name, phone or email..." 
                            value={search}
                            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
                            style={{ paddingLeft: '2.5rem', width: '100%' }}
                        />
                    </div>
                    
                    <div style={{ display: 'flex', gap: '1rem', position: 'relative' }}>
                        {showClearButton && (
                        <button 
                            className="btn-secondary" 
                            style={{ 
                                display: 'flex', 
                                alignItems: 'center', 
                                gap: '0.5rem', 
                                backgroundColor: '#fef2f2', 
                                color: '#ef4444', 
                                borderColor: '#fee2e2',
                                opacity: isDeleting ? 0.7 : 1
                            }}
                            onClick={handleDeleteAll}
                            disabled={isDeleting}
                        >
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M3 6h18"></path><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>
                            {isDeleting ? 'Clearing...' : 'Clear All Customers'}
                        </button>
                    )}
                        <button className="btn-secondary" onClick={() => setShowColumnPicker(!showColumnPicker)}>
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{marginRight: '8px'}}><line x1="4" y1="21" x2="4" y2="14"></line><line x1="4" y1="10" x2="4" y2="3"></line><line x1="12" y1="21" x2="12" y2="12"></line><line x1="12" y1="8" x2="12" y2="3"></line><line x1="20" y1="21" x2="20" y2="16"></line><line x1="20" y1="12" x2="20" y2="3"></line><line x1="1" y1="14" x2="7" y2="14"></line><line x1="9" y1="8" x2="15" y2="8"></line><line x1="17" y1="16" x2="23" y2="16"></line></svg>
                            Columns
                        </button>
                        
                        {showColumnPicker && (
                            <div className="premium-card" style={{ 
                                position: 'absolute', 
                                top: '100%', 
                                right: 0, 
                                zIndex: 100, 
                                marginTop: '8px', 
                                padding: '16px',
                                minWidth: '220px',
                                background: '#ffffff',
                                borderRadius: '12px',
                                boxShadow: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
                                border: '1px solid #e2e8f0'
                            }}>
                                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
                                    <div style={{ fontWeight: 600, fontSize: '0.9rem', color: '#1e293b' }}>Table Columns</div>
                                    <button 
                                        onClick={() => setVisibleColumns(['name', 'phone', 'location', 'orders', 'spent', 'activity'])}
                                        style={{ background: 'none', border: 'none', color: '#3b82f6', fontSize: '0.75rem', cursor: 'pointer', padding: 0 }}
                                    >
                                        Reset
                                    </button>
                                </div>
                                <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                                    {ALL_COLUMNS.map(col => (
                                        <label key={col.key} className="column-toggle-item" style={{ 
                                            display: 'flex', 
                                            alignItems: 'center', 
                                            gap: '10px', 
                                            padding: '6px 8px', 
                                            borderRadius: '6px',
                                            cursor: 'pointer',
                                            transition: 'background 0.2s'
                                        }}>
                                            <input 
                                                type="checkbox" 
                                                checked={visibleColumns.includes(col.key)} 
                                                onChange={() => toggleColumn(col.key)}
                                                style={{ cursor: 'pointer' }}
                                            />
                                            <span style={{ fontSize: '0.875rem', color: '#475569' }}>{col.label}</span>
                                        </label>
                                    ))}
                                </div>
                            </div>
                        )}
                    </div>
                </div>

                <div style={{ overflowX: 'auto' }}>
                    <table>
                        <thead>
                            <tr>
                                {ALL_COLUMNS.filter(c => visibleColumns.includes(c.key)).map(col => (
                                    <th 
                                        key={col.key} 
                                        onClick={() => handleSort(col.key)}
                                        style={{ cursor: 'pointer', userSelect: 'none' }}
                                    >
                                        <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                                            {col.label}
                                            {sortBy === col.key && (
                                                <span style={{ fontSize: '0.8rem' }}>
                                                    {sortOrder === 'ASC' ? '↑' : '↓'}
                                                </span>
                                            )}
                                        </div>
                                    </th>
                                ))}
                            </tr>
                        </thead>
                        <tbody>
                            {isLoading ? (
                                <tr><td colSpan={visibleColumns.length} style={{ textAlign: 'center', padding: '2rem' }}>Loading customers...</td></tr>
                            ) : customers.length === 0 ? (
                                <tr><td colSpan={visibleColumns.length} style={{ textAlign: 'center', padding: '2rem' }}>No customers found.</td></tr>
                            ) : (
                                customers.map((c) => (
                                    <tr key={c.id}>
                                        {visibleColumns.includes('name') && <td>{c.first_name || ''} {c.last_name || ''}</td>}
                                        {visibleColumns.includes('phone') && <td>{c.phone_number}</td>}
                                        {visibleColumns.includes('email') && <td>{c.email || 'N/A'}</td>}
                                        {visibleColumns.includes('location') && <td>{c.city ? `${c.city}, ${c.state || ''}` : 'N/A'}</td>}
                                        {visibleColumns.includes('orders') && <td>{c.total_orders || 0}</td>}
                                        {visibleColumns.includes('spent') && <td>₹{(c.total_spent || 0).toFixed(2)}</td>}
                                        {visibleColumns.includes('activity') && <td>{c.updated_at ? new Date(c.updated_at).toLocaleDateString() : 'N/A'}</td>}
                                        {visibleColumns.includes('source') && (
                                            <td>
                                                <span className={`badge ${c.source_id === 'shopify' ? 'badge-primary' : 'badge-secondary'}`} style={{ fontSize: '0.7rem' }}>
                                                    {c.source_id || 'manual'}
                                                </span>
                                            </td>
                                        )}
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>

                {customers.length > 0 && (
                    <div style={{ padding: '1rem', borderTop: '1px solid #e2e8f0', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <div style={{ color: '#64748b', fontSize: '0.875rem' }}>
                            Showing {(page-1)*limit + 1} to {Math.min(page*limit, total)} of {total} customers
                        </div>
                        <div style={{ display: 'flex', gap: '0.5rem' }}>
                            <button className="btn-secondary" disabled={page === 1} onClick={() => setPage(p => p - 1)}>Previous</button>
                            <button className="btn-secondary" disabled={page * limit >= total} onClick={() => setPage(p => p + 1)}>Next</button>
                        </div>
                    </div>
                )}
            </div>

            {showImportModal && (
                <div className="modal-overlay" onClick={() => !isImporting && setShowImportModal(false)}>
                    <div className="premium-modal" onClick={e => e.stopPropagation()}>
                        <h2 style={{ marginBottom: '1rem' }}>Import Customers</h2>
                        <form onSubmit={handleImport}>
                            <p style={{ color: '#64748b', marginBottom: '1.5rem', fontSize: '0.9rem' }}>
                                Upload your Shopify customer export CSV. Merging will be handled automatically by phone number.
                            </p>
                            <div style={{ display: 'grid', gap: '1.5rem', marginBottom: '2.5rem' }}>
                                <div style={{ margin: 0 }}>
                                    <label style={{ display: 'block', fontWeight: 600, marginBottom: '8px', fontSize: '0.9rem', color: '#1e293b' }}>Origin Source</label>
                                    <div style={{ position: 'relative' }}>
                                        <select 
                                            name="source_id" 
                                            value={selectedSource || ''} 
                                            onChange={(e) => setSelectedSource(e.target.value)}
                                            required
                                            style={{ 
                                                width: '100%', 
                                                padding: '0.875rem', 
                                                borderRadius: '12px', 
                                                border: '1px solid #e2e8f0', 
                                                background: '#f8fafc',
                                                fontSize: '0.95rem',
                                                fontWeight: 500,
                                                color: '#0f172a',
                                                cursor: 'pointer',
                                                appearance: 'none',
                                                transition: 'all 0.2s',
                                                boxShadow: '0 1px 2px rgba(0,0,0,0.05)'
                                            }}
                                        >
                                            <option value="" disabled>Select a source...</option>
                                            {sources.map(s => (
                                                <option key={s.id} value={s.id}>{s.name}</option>
                                            ))}
                                        </select>
                                        <div style={{ position: 'absolute', right: '12px', top: '50%', transform: 'translateY(-50%)', pointerEvents: 'none', color: '#64748b' }}>
                                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="6 9 12 15 18 9"></polyline></svg>
                                        </div>
                                    </div>
                                </div>

                                <div 
                                    style={{ 
                                        border: file ? '2px solid #10b981' : '2px dashed #e2e8f0', 
                                        borderRadius: '16px', 
                                        padding: '2.5rem 1.5rem', 
                                        textAlign: 'center',
                                        position: 'relative',
                                        background: file ? '#f0fdf4' : '#fafafa',
                                        transition: 'all 0.3s ease'
                                    }}
                                >
                                    <input 
                                        type="file" 
                                        name="file"
                                        accept=".csv"
                                        onChange={(e) => setFile(e.target.files?.[0] || null)}
                                        style={{ position: 'absolute', top: 0, left: 0, width: '100%', height: '100%', opacity: 0, cursor: 'pointer', zIndex: 10 }}
                                    />
                                    
                                    {file ? (
                                        <div style={{ animation: 'slideIn 0.3s ease-out' }}>
                                            <div style={{ 
                                                width: '48px', 
                                                height: '48px', 
                                                background: '#10b981', 
                                                borderRadius: '12px', 
                                                display: 'flex', 
                                                alignItems: 'center', 
                                                justifyContent: 'center',
                                                color: 'white',
                                                margin: '0 auto 1rem'
                                            }}>
                                                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><path d="M16 13l-4 4-2-2"></path></svg>
                                            </div>
                                            <p style={{ fontWeight: 700, color: '#065f46', marginBottom: '4px' }}>File Ready</p>
                                            <p style={{ fontSize: '0.875rem', color: '#059669', fontFamily: 'monospace' }}>{file.name}</p>
                                            <button 
                                                onClick={(e) => { e.stopPropagation(); setFile(null); }}
                                                style={{ marginTop: '1rem', background: 'none', border: 'none', color: '#ef4444', fontSize: '0.8rem', fontWeight: 600, cursor: 'pointer', zIndex: 20, position: 'relative' }}
                                            >
                                                Remove File
                                            </button>
                                        </div>
                                    ) : (
                                        <>
                                            <div style={{ color: '#94a3b8', marginBottom: '1rem' }}>
                                                <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="17 8 12 3 7 8"></polyline><line x1="12" y1="3" x2="12" y2="15"></line></svg>
                                            </div>
                                            <p style={{ fontWeight: 600, color: '#334155', marginBottom: '4px' }}>Click to upload or drag & drop</p>
                                            <p style={{ fontSize: '0.85rem', color: '#64748b' }}>Shopify Customer Export CSV only</p>
                                        </>
                                    )}
                                </div>
                            </div>
                            <div className="modal-actions">
                                <button type="button" className="btn-secondary" onClick={() => setShowImportModal(false)} disabled={isImporting}>Cancel</button>
                                <button type="submit" className="btn-primary" disabled={isImporting || !selectedSource || !file}>
                                    {isImporting ? 'Importing...' : 'Start Import'}
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
}
