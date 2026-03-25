import { useState, useEffect } from 'react';
import { API_BASE } from './api';
import { useToast } from './ToastContext';
import { useConfirm } from './ConfirmContext';
import { ColumnSelector } from './ColumnSelector';
import type { ColumnOption } from './ColumnSelector';

interface Customer {
    id: number;
    phone_number: string;
    first_name?: string;
    last_name?: string;
    email?: string;
    address1?: string;
    address2?: string;
    city?: string;
    state?: string;
    country?: string;
    zip_code?: string;
    total_orders: number;
    total_spent: number;
    created_at: string;
    updated_at: string;
    source_id: string;
    external_id?: string;
}

interface Source {
    id: string;
    name: string;
}

interface WhatsAppTemplate {
    id: number;
    template_name: string;
    language: string;
    category: string;
    status: string;
    body: string;
}

interface CustomersProps {
    fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
    showClearButton?: boolean;
    bulkSuffix?: string;
    userRole?: string;
}

type ColumnKey = 'name' | 'phone' | 'email' | 'location' | 'orders' | 'spent' | 'activity' | 'source';

const CUSTOMER_COLUMN_OPTIONS: ColumnOption[] = [
    { id: 'name', label: 'Name', category: 'Identity' },
    { id: 'phone', label: 'Phone', category: 'Identity' },
    { id: 'email', label: 'Email', category: 'Identity' },
    { id: 'location', label: 'Location', category: 'Location' },
    { id: 'orders', label: 'Orders', category: 'Engagement' },
    { id: 'spent', label: 'Total Spent', category: 'Engagement' },
    { id: 'activity', label: 'Last Activity', category: 'Engagement' },
    { id: 'source', label: 'Source', category: 'System' },
];

const DEFAULT_CUSTOMER_COLUMNS: ColumnKey[] = ['name', 'phone', 'location', 'orders', 'spent', 'activity'];

export function Customers({ fetchWithAuth, showClearButton = false, bulkSuffix = '_marketing', userRole = 'read' }: CustomersProps) {
    const { success: toastSuccess, error: toastError } = useToast();
    const { confirm } = useConfirm();
    const [file, setFile] = useState<File | null>(null);
    const [customers, setCustomers] = useState<Customer[]>([]);
    const [total, setTotal] = useState(0);
    const [page, setPage] = useState(1);
    const [search, setSearch] = useState('');
    const [isLoading, setIsLoading] = useState(true);
    const [isImporting, setIsImporting] = useState(false);
    const [showImportModal, setShowImportModal] = useState(false);
    const [selectedCustomer, setSelectedCustomer] = useState<Customer | null>(null);
    const [selectedCustomerIDs, setSelectedCustomerIDs] = useState<Set<number>>(new Set());
    const [showBulkModal, setShowBulkModal] = useState(false);
    const [visibleColumns, setVisibleColumns] = useState<ColumnKey[]>(() => {
        const saved = localStorage.getItem('customer_columns');
        return saved ? JSON.parse(saved) : DEFAULT_CUSTOMER_COLUMNS;
    });
    const [sortBy, setSortBy] = useState<ColumnKey>('activity');
    const [sortOrder, setSortOrder] = useState<'ASC' | 'DESC'>('DESC');
    const [sources, setSources] = useState<Source[]>([]);
    const [selectedSource, setSelectedSource] = useState<string>('');
    const [isDeleting, setIsDeleting] = useState(false);
    const [showFilters, setShowFilters] = useState(false);
    const [filterSource, setFilterSource] = useState('');
    const [minSpent, setMinSpent] = useState('');
    const [maxSpent, setMaxSpent] = useState('');
    const [minOrders, setMinOrders] = useState('');
    const [city, setCity] = useState('');
    const [state, setState] = useState('');
    const [showAddModal, setShowAddModal] = useState(false);
    const [isEditMode, setIsEditMode] = useState(false);
    const [isSaving, setIsSaving] = useState(false);
    const [syncToShopify, setSyncToShopify] = useState(false);
    const [customerForm, setCustomerForm] = useState<Partial<Customer>>({});
    const limit = 25;

    useEffect(() => {
        localStorage.setItem('customer_columns', JSON.stringify(visibleColumns));
    }, [visibleColumns]);

    const handleSort = (key: ColumnKey) => {
        if (sortBy === key) {
            setSortOrder(prev => prev === 'ASC' ? 'DESC' : 'ASC');
        } else {
            setSortBy(key);
            setSortOrder('DESC');
        }
        setPage(1);
    };

    const handleSaveCustomer = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsSaving(true);
        try {
            const url = isEditMode ? `${API_BASE}/api/customers/${customerForm.id}/` : `${API_BASE}/api/customers`;
            const method = isEditMode ? 'PUT' : 'POST';
            
            const response = await fetchWithAuth(url, {
                method,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    ...customerForm,
                    sync_to_shopify: syncToShopify
                })
            });

            if (!response.ok) throw new Error('Failed to save customer');

            await fetchCustomers();
            setShowAddModal(false);
            setSelectedCustomer(null);
            setIsEditMode(false);
            setCustomerForm({});
            toastSuccess(isEditMode ? 'Customer updated successfully' : 'Customer created successfully');
        } catch (err: any) {
            console.error('Save error:', err);
            toastError(err.message || 'An error occurred while saving');
        } finally {
            setIsSaving(false);
        }
    };

    const editExistingCustomer = (c: Customer) => {
        setCustomerForm(c);
        setIsEditMode(true);
        setShowAddModal(true);
        setSelectedCustomer(null);
    };

    const handleDeleteCustomer = async (id: number) => {
        const confirmed = await confirm({
            title: 'Delete Customer',
            message: 'Are you sure you want to delete this customer? This will also remove them from Shopify if linked.',
            variant: 'danger',
            confirmLabel: 'Delete'
        });

        if (!confirmed) return;

        try {
            const response = await fetchWithAuth(`${API_BASE}/api/customers/${id}/`, {
                method: 'DELETE',
            });

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || 'Failed to delete customer');
            }

            toastSuccess('Customer deleted successfully');
            setSelectedCustomer(null);
            fetchCustomers();
        } catch (err: any) {
            console.error('Delete error:', err);
            toastError(err.message || 'An error occurred while deleting');
        }
    };

    const fetchCustomers = async () => {
        setIsLoading(true);
        try {
            let url = `${API_BASE}/api/customers?page=${page}&pageSize=${limit}&search=${search}&sortBy=${sortBy}&sortOrder=${sortOrder}`;
            if (filterSource) url += `&source_id=${filterSource}`;
            if (minSpent) url += `&min_spent=${minSpent}`;
            if (maxSpent) url += `&max_spent=${maxSpent}`;
            if (minOrders) url += `&min_orders=${minOrders}`;
            if (city) url += `&city=${encodeURIComponent(city)}`;
            if (state) url += `&state=${encodeURIComponent(state)}`;

            const response = await fetchWithAuth(url);
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
    }, [page, search, sortBy, sortOrder, filterSource, minSpent, maxSpent, minOrders, city, state]);

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
                toastSuccess('Import successful!');
                setFile(null);
                setShowImportModal(false);
                fetchCustomers();
            } else {
                const errorData = await response.json();
                toastError('Import failed: ' + (errorData.message || 'Unknown error'));
            }
        } catch (error) {
            console.error('Error importing customers:', error);
            toastError('Error during import.');
        } finally {
            setIsImporting(false);
        }
    };

    const handleBulkDelete = async () => {
        if (selectedCustomerIDs.size === 0) return;
        
        const confirmed = await confirm({
            title: 'Bulk Delete',
            message: `Are you sure you want to delete ${selectedCustomerIDs.size} selected customers? This will also remove them from Shopify if linked.`,
            variant: 'danger',
            confirmLabel: `Delete ${selectedCustomerIDs.size} Customers`
        });

        if (!confirmed) return;

        setIsDeleting(true);
        try {
            const response = await fetchWithAuth(`${API_BASE}/api/customers/bulk-delete`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ ids: Array.from(selectedCustomerIDs) }),
            });
            if (response.ok) {
                toastSuccess('Selected customers deleted successfully.');
                setSelectedCustomerIDs(new Set());
                fetchCustomers();
            } else {
                const errorText = await response.text();
                toastError('Failed to delete customers: ' + errorText);
            }
        } catch (error) {
            console.error('Error bulk deleting customers:', error);
            toastError('Error during bulk deletion.');
        } finally {
            setIsDeleting(false);
        }
    };

    const handleDeleteAll = async () => {
        const confirmed = await confirm({
            title: 'Clear All Customers',
            message: 'Are you absolutely sure? This will permanently delete ALL customers from the database. This action cannot be undone.',
            variant: 'danger',
            confirmLabel: 'Clear All'
        });

        if (!confirmed) return;

        setIsDeleting(true);
        try {
            const response = await fetchWithAuth(`${API_BASE}/api/customers`, {
                method: 'DELETE',
            });
            if (response.ok) {
                toastSuccess('All customers cleared successfully.');
                fetchCustomers();
            } else {
                toastError('Failed to clear customers.');
            }
        } catch (error) {
            console.error('Error deleting customers:', error);
            toastError('Error during deletion.');
        } finally {
            setIsDeleting(false);
        }
    };

    return (
        <div className="tab-pane active">
            <div style={{ 
                display: 'flex', 
                flexDirection: 'column',
                gap: '1.5rem',
                marginBottom: '2rem',
                padding: '2rem',
                background: 'linear-gradient(135deg, #ffffff 0%, #f8fafc 100%)',
                borderRadius: '20px',
                boxShadow: '0 10px 25px -5px rgba(0, 0, 0, 0.05), 0 8px 10px -6px rgba(0, 0, 0, 0.05)',
                border: '1px solid #f1f5f9',
                borderTop: '4px solid var(--accent-color)',
                position: 'relative',
                overflow: 'hidden'
            }}>
                {/* Decorative background element */}
                <div style={{ 
                    position: 'absolute', 
                    top: '-20px', 
                    right: '-20px', 
                    width: '120px', 
                    height: '120px', 
                    background: 'var(--accent-color)', 
                    opacity: 0.03, 
                    borderRadius: '50%',
                    pointerEvents: 'none'
                }}></div>

                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                    <div>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '0.5rem' }}>
                            <div style={{ 
                                padding: '0.6rem', 
                                background: 'rgba(59, 130, 246, 0.1)', 
                                borderRadius: '12px',
                                color: 'var(--accent-color)',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center'
                            }}>
                                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>
                            </div>
                            <h1 style={{ margin: 0, fontSize: '1.85rem', fontWeight: 800, color: '#0f172a', letterSpacing: '-0.03em' }}>
                                Customer Directory
                            </h1>
                        </div>
                        <p style={{ margin: 0, color: '#64748b', fontSize: '1rem', fontWeight: 500, paddingLeft: '3.2rem' }}>
                            Manage and analyze your customer base across all sources
                        </p>
                    </div>
                    
                    <div style={{ 
                        background: 'white', 
                        padding: '1rem 1.5rem', 
                        borderRadius: '16px', 
                        border: '1px solid #e2e8f0',
                        boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.02)',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '1rem',
                        minWidth: '180px'
                    }}>
                        <div style={{ 
                            width: '40px', 
                            height: '40px', 
                            borderRadius: '10px', 
                            background: '#f1f5f9', 
                            display: 'flex', 
                            alignItems: 'center', 
                            justifyContent: 'center',
                            color: '#64748b'
                        }}>
                            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="8.5" cy="7" r="4"></circle><polyline points="17 11 19 13 23 9"></polyline></svg>
                        </div>
                        <div>
                            <div style={{ fontSize: '0.7rem', fontWeight: 700, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Total Customers</div>
                            <div style={{ fontSize: '1.5rem', fontWeight: 800, color: '#0f172a', lineHeight: 1.2 }}>{total.toLocaleString()}</div>
                        </div>
                    </div>
                </div>

                {userRole === 'admin' && (
                    <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap', paddingTop: '0.5rem' }}>
                        <button 
                            className="btn-primary" 
                            onClick={() => { setCustomerForm({}); setIsEditMode(false); setShowAddModal(true); }} 
                            style={{ 
                                padding: '0.75rem 1.5rem', 
                                fontSize: '0.875rem', 
                                fontWeight: 600,
                                background: '#0f172a', 
                                display: 'flex', 
                                alignItems: 'center', 
                                gap: '10px', 
                                borderRadius: '12px',
                                transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
                                boxShadow: '0 4px 12px rgba(15, 23, 42, 0.15)'
                            }}
                        >
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
                            Add Customer
                        </button>
                        <button 
                            className="btn-secondary" 
                            onClick={() => setShowImportModal(true)} 
                            style={{ 
                                padding: '0.75rem 1.5rem', 
                                fontSize: '0.875rem', 
                                fontWeight: 600,
                                display: 'flex', 
                                alignItems: 'center', 
                                gap: '10px', 
                                borderRadius: '12px',
                                background: 'white',
                                border: '1px solid #e2e8f0',
                                transition: 'all 0.2s'
                            }}
                        >
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
                            Import
                        </button>
                        <button 
                            className="btn-secondary" 
                            style={{ 
                                padding: '0.75rem 1.5rem', 
                                fontSize: '0.875rem', 
                                fontWeight: 600,
                                display: 'flex', 
                                alignItems: 'center', 
                                gap: '10px', 
                                borderRadius: '12px',
                                background: 'white',
                                border: '1px solid #e2e8f0',
                                transition: 'all 0.2s'
                            }}
                        >
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
                            Export
                        </button>
                        
                        <div style={{ flex: 1 }}></div>

                        {selectedCustomerIDs.size > 0 && (
                            <button 
                                className="btn-danger-minimal" 
                                onClick={handleBulkDelete}
                                disabled={isDeleting}
                                style={{ 
                                    padding: '0.75rem 1.5rem', 
                                    fontSize: '0.875rem', 
                                    fontWeight: 600,
                                    display: 'flex', 
                                    alignItems: 'center', 
                                    gap: '10px', 
                                    color: '#dc2626', 
                                    background: '#fef2f2',
                                    border: '1px solid #fee2e2',
                                    borderRadius: '12px',
                                    transition: 'all 0.2s'
                                }}
                            >
                                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>
                                {isDeleting ? 'Deleting...' : `Clear Selected (${selectedCustomerIDs.size})`}
                            </button>
                        )}

                        {showClearButton && !isDeleting && selectedCustomerIDs.size === 0 && (
                            <button 
                                className="btn-secondary" 
                                onClick={handleDeleteAll} 
                                style={{ 
                                    padding: '0.75rem 1.5rem', 
                                    fontSize: '0.875rem', 
                                    fontWeight: 500,
                                    display: 'flex', 
                                    alignItems: 'center', 
                                    gap: '8px', 
                                    color: '#94a3b8', 
                                    border: '1px solid transparent',
                                    background: 'transparent',
                                    borderRadius: '12px',
                                    transition: 'all 0.2s'
                                }}
                                onMouseEnter={e => {
                                    e.currentTarget.style.color = '#dc2626';
                                    e.currentTarget.style.background = '#fef2f2';
                                    e.currentTarget.style.borderColor = '#fee2e2';
                                }}
                                onMouseLeave={e => {
                                    e.currentTarget.style.color = '#94a3b8';
                                    e.currentTarget.style.background = 'transparent';
                                    e.currentTarget.style.borderColor = 'transparent';
                                }}
                            >
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
                                Clear All
                            </button>
                        )}
                    </div>
                )}
            </div>

            <div className="card">
                <div className="table-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <div style={{ flex: 1, position: 'relative', marginRight: '2rem' }}>
                        <svg style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: '#94a3b8' }} width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
                        <input 
                            type="text" 
                            placeholder="Search (e.g. city:Mumbai spent>1000 or first_name='')" 
                            aria-label="Search customers"
                            value={search}
                            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
                            style={{ paddingLeft: '2.5rem', width: '100%' }}
                        />
                    </div>
                    <div style={{ display: 'flex', gap: '1rem', position: 'relative' }}>
                        <button 
                            className={`btn-secondary ${showFilters ? 'active' : ''}`} 
                            onClick={() => setShowFilters(!showFilters)}
                            style={{ 
                                display: 'flex', 
                                alignItems: 'center', 
                                gap: '0.5rem',
                                color: showFilters ? '#3b82f6' : 'inherit',
                                borderColor: showFilters ? '#3b82f6' : '#e2e8f0'
                            }}
                        >
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"></polygon></svg>
                            Filters
                        </button>

                        <ColumnSelector
                            columns={CUSTOMER_COLUMN_OPTIONS}
                            visibleColumns={visibleColumns}
                            onChange={(cols) => setVisibleColumns(cols as ColumnKey[])}
                            onReset={() => setVisibleColumns(DEFAULT_CUSTOMER_COLUMNS)}
                        />
                    </div>
                </div>

                {showFilters && (
                    <div style={{ padding: '1.5rem', borderBottom: '1px solid #f1f5f9', background: '#f8fafc', animation: 'slideDown 0.3s ease-out' }}>
                        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1.5rem' }}>
                            <div>
                                <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: '#64748b', marginBottom: '0.5rem' }}>Origin Source</label>
                                <select 
                                    className="form-input" 
                                    value={filterSource} 
                                    onChange={(e) => { setFilterSource(e.target.value); setPage(1); }}
                                    style={{ width: '100%', height: '40px', background: 'white' }}
                                >
                                    <option value="">All Sources</option>
                                    {sources.map(s => (
                                        <option key={s.id} value={s.id}>{s.name}</option>
                                    ))}
                                </select>
                            </div>
                            <div>
                                <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: '#64748b', marginBottom: '0.5rem' }}>Total Spent (Min)</label>
                                <input 
                                    type="number" 
                                    className="form-input" 
                                    placeholder="e.g. 1000" 
                                    value={minSpent} 
                                    onChange={(e) => { setMinSpent(e.target.value); setPage(1); }} 
                                    style={{ width: '100%', height: '40px', background: 'white' }}
                                />
                            </div>
                            <div>
                                <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: '#64748b', marginBottom: '0.5rem' }}>Min Orders</label>
                                <input 
                                    type="number" 
                                    className="form-input" 
                                    placeholder="e.g. 5" 
                                    value={minOrders} 
                                    onChange={(e) => { setMinOrders(e.target.value); setPage(1); }} 
                                    style={{ width: '100%', height: '40px', background: 'white' }}
                                />
                            </div>
                            <div>
                                <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: '#64748b', marginBottom: '0.5rem' }}>City / State</label>
                                <div style={{ display: 'flex', gap: '0.5rem' }}>
                                    <input 
                                        type="text" 
                                        className="form-input" 
                                        placeholder="City" 
                                        value={city} 
                                        onChange={(e) => { setCity(e.target.value); setPage(1); }} 
                                        style={{ width: '50%', height: '40px', background: 'white' }}
                                    />
                                    <input 
                                        type="text" 
                                        className="form-input" 
                                        placeholder="State" 
                                        value={state} 
                                        onChange={(e) => { setState(e.target.value); setPage(1); }} 
                                        style={{ width: '50%', height: '40px', background: 'white' }}
                                    />
                                </div>
                            </div>
                        </div>
                        <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: '1rem' }}>
                            <button 
                                className="btn-secondary" 
                                style={{ fontSize: '0.85rem', color: '#64748b' }}
                                onClick={() => {
                                    setFilterSource('');
                                    setMinSpent('');
                                    setMaxSpent('');
                                    setMinOrders('');
                                    setCity('');
                                    setState('');
                                    setSearch('');
                                    setPage(1);
                                }}
                            >
                                Clear All Filters
                            </button>
                        </div>
                    </div>
                )}

                <div style={{ overflowX: 'auto' }}>
                    <table>
                        <thead>
                            <tr>
                                <th style={{ width: '40px', padding: '0 1rem' }}>
                                    <input 
                                        type="checkbox" 
                                        checked={customers.length > 0 && customers.every(c => selectedCustomerIDs.has(c.id))}
                                        onChange={(e) => {
                                            const newSelection = new Set(selectedCustomerIDs);
                                            if (e.target.checked) {
                                                customers.forEach(c => newSelection.add(c.id));
                                            } else {
                                                customers.forEach(c => newSelection.delete(c.id));
                                            }
                                            setSelectedCustomerIDs(newSelection);
                                        }}
                                        style={{ cursor: 'pointer', width: '16px', height: '16px' }}
                                    />
                                </th>
                                {CUSTOMER_COLUMN_OPTIONS.filter(c => visibleColumns.includes(c.id as ColumnKey)).map(col => (
                                    <th 
                                        key={col.id}
                                        onClick={() => handleSort(col.id as ColumnKey)}
                                        style={{ cursor: 'pointer', userSelect: 'none' }}
                                    >
                                        <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                                            {col.label}
                                            {sortBy === col.id && (
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
                                    <tr 
                                        key={c.id} 
                                        onClick={() => setSelectedCustomer(c)}
                                        style={{ 
                                            cursor: 'pointer', 
                                            transition: 'background 0.2s',
                                            background: selectedCustomerIDs.has(c.id) ? '#f0f9ff' : 'transparent'
                                        }}
                                        onMouseEnter={(e) => !selectedCustomerIDs.has(c.id) && (e.currentTarget.style.background = '#f8fafc')}
                                        onMouseLeave={(e) => !selectedCustomerIDs.has(c.id) && (e.currentTarget.style.background = 'transparent')}
                                    >
                                        <td onClick={(e) => e.stopPropagation()} style={{ width: '40px', padding: '0 1rem' }}>
                                            <input 
                                                type="checkbox" 
                                                checked={selectedCustomerIDs.has(c.id)}
                                                onChange={(e) => {
                                                    const newSelection = new Set(selectedCustomerIDs);
                                                    if (e.target.checked) {
                                                        newSelection.add(c.id);
                                                    } else {
                                                        newSelection.delete(c.id);
                                                    }
                                                    setSelectedCustomerIDs(newSelection);
                                                }}
                                                style={{ cursor: 'pointer', width: '16px', height: '16px' }}
                                            />
                                        </td>
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
                    {/* ... existing import modal content ... */}
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
                                                type="button"
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

            {selectedCustomer && (
                <div className="modal-overlay" onClick={() => setSelectedCustomer(null)}>
                    <div className="premium-modal" onClick={e => e.stopPropagation()} style={{ maxWidth: '600px', width: '90%' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '1.5rem', marginBottom: '2rem' }}>
                            <div style={{ 
                                width: '64px', 
                                height: '64px', 
                                borderRadius: '50%', 
                                background: 'linear-gradient(135deg, #6366f1 0%, #a855f7 100%)',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                color: 'white',
                                fontSize: '1.5rem',
                                fontWeight: 600,
                                boxShadow: '0 4px 12px rgba(99, 102, 241, 0.3)'
                            }}>
                                {(selectedCustomer.first_name?.[0] || '') + (selectedCustomer.last_name?.[0] || '') || '?'}
                            </div>
                            <div>
                                <h2 style={{ margin: 0 }}>{selectedCustomer.first_name || 'Guest'} {selectedCustomer.last_name || 'Customer'}</h2>
                                <p style={{ color: '#64748b', margin: '4px 0 0' }}>ID: {selectedCustomer.id.toString().substring(0, 8)}{selectedCustomer.id.toString().length > 8 ? '...' : ''}</p>
                            </div>
                            <button 
                                onClick={(e) => { e.stopPropagation(); editExistingCustomer(selectedCustomer); }}
                                style={{ marginLeft: 'auto', background: '#f0f9ff', border: 'none', padding: '0.5rem 1rem', borderRadius: '8px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '6px', color: '#0369a1', fontWeight: 600, fontSize: '0.85rem' }}
                            >
                                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4L18.5 2.5z"></path></svg>
                                Edit
                            </button>
                            <button 
                                onClick={(e) => { e.stopPropagation(); handleDeleteCustomer(selectedCustomer.id); }}
                                style={{ background: '#fef2f2', border: 'none', padding: '0.5rem 1rem', borderRadius: '8px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '6px', color: '#b91c1c', fontWeight: 600, fontSize: '0.85rem' }}
                            >
                                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M3 6h18"></path><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
                                Delete
                            </button>
                            <button 
                                onClick={() => setSelectedCustomer(null)}
                                style={{ background: '#f1f5f9', border: 'none', width: '32px', height: '32px', borderRadius: '50%', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#64748b' }}
                                aria-label="Close customer details"
                            >
                                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                            </button>
                        </div>

                        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '2rem' }}>
                            <div style={{ background: '#f8fafc', padding: '1rem', borderRadius: '12px', border: '1px solid #f1f5f9' }}>
                                <p style={{ fontSize: '0.75rem', fontWeight: 600, color: '#64748b', textTransform: 'uppercase', marginBottom: '4px', letterSpacing: '0.025em' }}>Total Orders</p>
                                <p style={{ fontSize: '1.25rem', fontWeight: 700, margin: 0, color: '#1e293b' }}>{selectedCustomer.total_orders || 0}</p>
                            </div>
                            <div style={{ background: '#f8fafc', padding: '1rem', borderRadius: '12px', border: '1px solid #f1f5f9' }}>
                                <p style={{ fontSize: '0.75rem', fontWeight: 600, color: '#64748b', textTransform: 'uppercase', marginBottom: '4px', letterSpacing: '0.025em' }}>Total Spent</p>
                                <p style={{ fontSize: '1.25rem', fontWeight: 700, margin: 0, color: '#10b981' }}>₹{(selectedCustomer.total_spent || 0).toFixed(2)}</p>
                            </div>
                        </div>

                        <div style={{ display: 'grid', gap: '1.5rem' }}>
                            <section>
                                <h3 style={{ fontSize: '0.9rem', marginBottom: '1rem', color: '#1e293b', borderBottom: '1px solid #f1f5f9', paddingBottom: '0.5rem' }}>Contact Information</h3>
                                <div style={{ display: 'grid', gap: '0.75rem' }}>
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', fontSize: '0.95rem' }}>
                                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#64748b" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"></path></svg>
                                        <span>{selectedCustomer.phone_number}</span>
                                    </div>
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', fontSize: '0.95rem' }}>
                                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#64748b" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path><polyline points="22,6 12,13 2,6"></polyline></svg>
                                        <span>{selectedCustomer.email || 'No email provided'}</span>
                                    </div>
                                </div>
                            </section>

                            <section>
                                <h3 style={{ fontSize: '0.9rem', marginBottom: '1rem', color: '#1e293b', borderBottom: '1px solid #f1f5f9', paddingBottom: '0.5rem' }}>Address</h3>
                                <div style={{ color: '#475569', fontSize: '0.95rem', lineHeight: '1.5' }}>
                                    {selectedCustomer.address1 ? (
                                        <>
                                            <div>{selectedCustomer.address1}</div>
                                            {selectedCustomer.address2 && <div>{selectedCustomer.address2}</div>}
                                            <div>{selectedCustomer.city || ''}{selectedCustomer.city && selectedCustomer.state ? ', ' : ''}{selectedCustomer.state || ''} {selectedCustomer.zip_code || ''}</div>
                                            <div>{selectedCustomer.country || ''}</div>
                                        </>
                                    ) : (
                                        <span style={{ color: '#94a3b8', fontStyle: 'italic' }}>No address information available</span>
                                    )}
                                </div>
                            </section>

                            <section style={{ background: '#f8fafc', padding: '1.25rem', borderRadius: '12px', marginTop: '0.5rem' }}>
                                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.5rem' }}>
                                    <div>
                                        <p style={{ fontSize: '0.7rem', color: '#64748b', fontWeight: 600, textTransform: 'uppercase', marginBottom: '4px' }}>Source</p>
                                        <span className={`badge ${selectedCustomer.source_id === 'shopify' ? 'badge-primary' : 'badge-secondary'}`}>
                                            {selectedCustomer.source_id || 'manual'}
                                        </span>
                                    </div>
                                    <div>
                                        <p style={{ fontSize: '0.7rem', color: '#64748b', fontWeight: 600, textTransform: 'uppercase', marginBottom: '4px' }}>Member Since</p>
                                        <p style={{ fontSize: '0.9rem', margin: 0, fontWeight: 500 }}>{selectedCustomer.created_at ? new Date(selectedCustomer.created_at).toLocaleDateString() : 'N/A'}</p>
                                    </div>
                                </div>
                            </section>
                        </div>
                    </div>
                </div>
            )}

            {selectedCustomerIDs.size > 0 && (
                <div style={{ 
                    position: 'fixed', 
                    bottom: '2rem', 
                    left: '50%', 
                    transform: 'translateX(-50%)', 
                    background: '#1e293b', 
                    color: 'white', 
                    padding: '1rem 2rem', 
                    borderRadius: '99px', 
                    display: 'flex', 
                    alignItems: 'center', 
                    gap: '2rem', 
                    boxShadow: '0 10px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.1)',
                    zIndex: 1000,
                    animation: 'slideUp 0.3s cubic-bezier(0.34, 1.56, 0.64, 1)'
                }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                        <div style={{ background: '#38bdf8', color: '#0c4a6e', width: '24px', height: '24px', borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '0.75rem', fontWeight: 700 }}>
                            {selectedCustomerIDs.size}
                        </div>
                        <span style={{ fontWeight: 500, fontSize: '0.9rem' }}>Customers Selected</span>
                    </div>
                    <div style={{ height: '20px', width: '1px', background: '#334155' }}></div>
                    <div style={{ display: 'flex', gap: '1rem' }}>
                        <button 
                            className="btn-primary" 
                            style={{ padding: '0.5rem 1.25rem', fontSize: '0.85rem', background: '#38bdf8', color: '#0c4a6e', border: 'none' }}
                            onClick={() => setShowBulkModal(true)}
>
                            Send Bulk Message
                        </button>
                        <button 
                            style={{ background: 'transparent', border: 'none', color: '#94a3b8', fontSize: '0.85rem', cursor: 'pointer', fontWeight: 500 }}
                            onClick={() => setSelectedCustomerIDs(new Set())}
                        >
                            Clear Selection
                        </button>
                    </div>
                </div>
            )}

            <BulkTemplateModal
                isOpen={showBulkModal}
                onClose={() => setShowBulkModal(false)}
                customerIDs={Array.from(selectedCustomerIDs)}
                onSuccess={() => setSelectedCustomerIDs(new Set())}
                fetchWithAuth={fetchWithAuth}
                bulkSuffix={bulkSuffix}
            />

            {showAddModal && (
                <div className="modal-overlay" onClick={() => setShowAddModal(false)}>
                    <div className="premium-modal" onClick={e => e.stopPropagation()} style={{
                        maxWidth: '700px',
                        width: '95%',
                        maxHeight: '90vh',
                        display: 'flex',
                        flexDirection: 'column',
                        padding: 0,
                        overflow: 'hidden'
                    }}>
                        {/* Sticky Header */}
                        <div style={{
                            padding: '1.5rem',
                            borderBottom: '1px solid #e2e8f0',
                            display: 'flex',
                            justifyContent: 'space-between',
                            alignItems: 'center',
                            background: 'white',
                            zIndex: 10
                        }}>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                                <div style={{
                                    width: '40px',
                                    height: '40px',
                                    borderRadius: '10px',
                                    background: '#f1f5f9',
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center',
                                    color: '#0f172a'
                                }}>
                                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>
                                </div>
                                <div>
                                    <h2 style={{ margin: 0, fontSize: '1.25rem', color: '#0f172a' }}>{isEditMode ? 'Edit Customer' : 'Add New Customer'}</h2>
                                    <p style={{ margin: '2px 0 0', fontSize: '0.85rem', color: '#64748b' }}>
                                        {isEditMode ? 'Update existing customer details' : 'Create a new customer profile'}
                                    </p>
                                </div>
                            </div>
                            <button
                                onClick={() => setShowAddModal(false)}
                                style={{ background: 'none', border: 'none', cursor: 'pointer', color: '#94a3b8', padding: '4px', borderRadius: '6px' }}
                                className="hover-bg"
                                aria-label="Close modal"
                            >
                                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                            </button>
                        </div>

                        {/* Scrollable Content */}
                        <form onSubmit={handleSaveCustomer} style={{ display: 'flex', flexDirection: 'column', height: '100%', overflow: 'hidden' }}>
                            <div style={{ padding: '1.5rem', overflowY: 'auto', flex: 1 }}>

                                {/* Section 1: Basic Info */}
                                <div style={{ marginBottom: '2rem' }}>
                                    <h3 style={{ fontSize: '0.9rem', color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '1.25rem', display: 'flex', alignItems: 'center', gap: '8px' }}>
                                        <span style={{ width: '8px', height: '8px', borderRadius: '50%', background: '#3b82f6' }}></span>
                                        Contact Information
                                    </h3>

                                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.25rem', marginBottom: '1.25rem' }}>
                                        <div>
                                            <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: '#475569', marginBottom: '6px' }}>First Name</label>
                                            <input
                                                type="text"
                                                className="form-input"
                                                value={customerForm.first_name || ''}
                                                onChange={e => setCustomerForm({...customerForm, first_name: e.target.value})}
                                                placeholder="John"
                                                required
                                            />
                                        </div>
                                        <div>
                                            <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: '#475569', marginBottom: '6px' }}>Last Name</label>
                                            <input
                                                type="text"
                                                className="form-input"
                                                value={customerForm.last_name || ''}
                                                onChange={e => setCustomerForm({...customerForm, last_name: e.target.value})}
                                                placeholder="Doe"
                                            />
                                        </div>
                                    </div>

                                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.25rem', marginBottom: '1.25rem' }}>
                                        <div>
                                            <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: '#475569', marginBottom: '6px' }}>Phone Number</label>
                                            <div style={{ position: 'relative', display: 'flex', alignItems: 'center' }}>
                                                <span style={{
                                                    position: 'absolute',
                                                    left: '12px',
                                                    color: '#64748b',
                                                    fontWeight: 600,
                                                    fontSize: '0.9rem',
                                                    pointerEvents: 'none'
                                                }}>+91</span>
                                                <input
                                                    type="tel"
                                                    className="form-input"
                                                    style={{ paddingLeft: '45px' }}
                                                    value={customerForm.phone_number?.replace('+91', '') || ''}
                                                    onChange={e => {
                                                        const val = e.target.value.replace(/\D/g, '').substring(0, 10);
                                                        setCustomerForm({...customerForm, phone_number: '+91' + val});
                                                    }}
                                                    placeholder="9876543210"
                                                    required
                                                />
                                            </div>
                                        </div>
                                        <div>
                                            <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: '#475569', marginBottom: '6px' }}>Email Address</label>
                                            <input
                                                type="email"
                                                className="form-input"
                                                value={customerForm.email || ''}
                                                onChange={e => setCustomerForm({...customerForm, email: e.target.value})}
                                                placeholder="john@example.com"
                                            />
                                        </div>
                                    </div>

                                    <div>
                                        <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: '#475569', marginBottom: '6px' }}>Customer Source</label>
                                        <select
                                            className="form-input"
                                            value={customerForm.source_id || ''}
                                            onChange={e => {
                                                const newSource = e.target.value;
                                                setCustomerForm({...customerForm, source_id: newSource});
                                                if (newSource === 'shopify') {
                                                    setSyncToShopify(true);
                                                }
                                            }}
                                            required
                                        >
                                            <option value="">Select Source</option>
                                            {sources.map(s => (
                                                <option key={s.id} value={s.id}>{s.name}</option>
                                            ))}
                                        </select>
                                    </div>
                                </div>

                                {/* Section 2: Address */}
                                <div style={{ marginBottom: '1rem' }}>
                                    <h3 style={{ fontSize: '0.9rem', color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '1.25rem', display: 'flex', alignItems: 'center', gap: '8px' }}>
                                        <span style={{ width: '8px', height: '8px', borderRadius: '50%', background: '#10b981' }}></span>
                                        Shipping Address
                                    </h3>

                                    <div style={{ marginBottom: '1.25rem' }}>
                                        <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: '#475569', marginBottom: '6px' }}>Address Line 1</label>
                                        <input
                                            type="text"
                                            className="form-input"
                                            value={customerForm.address1 || ''}
                                            onChange={e => setCustomerForm({...customerForm, address1: e.target.value})}
                                            placeholder="Apartment, Street Name"
                                        />
                                    </div>

                                    <div style={{ marginBottom: '1.25rem' }}>
                                        <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: '#475569', marginBottom: '6px' }}>Address Line 2 (Optional)</label>
                                        <input
                                            type="text"
                                            className="form-input"
                                            value={customerForm.address2 || ''}
                                            onChange={e => setCustomerForm({...customerForm, address2: e.target.value})}
                                            placeholder="House No, Landmark, Area"
                                        />
                                    </div>

                                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.25rem', marginBottom: '1.25rem' }}>
                                        <div>
                                            <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: '#475569', marginBottom: '6px' }}>City</label>
                                            <input
                                                type="text"
                                                className="form-input"
                                                value={customerForm.city || ''}
                                                onChange={e => setCustomerForm({...customerForm, city: e.target.value})}
                                                placeholder="Mumbai"
                                            />
                                        </div>
                                        <div>
                                            <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: '#475569', marginBottom: '6px' }}>State</label>
                                            <input
                                                type="text"
                                                className="form-input"
                                                value={customerForm.state || ''}
                                                onChange={e => setCustomerForm({...customerForm, state: e.target.value})}
                                                placeholder="Maharashtra"
                                            />
                                        </div>
                                    </div>

                                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.25rem', marginBottom: '1.5rem' }}>
                                        <div>
                                            <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: '#475569', marginBottom: '6px' }}>Country</label>
                                            <input
                                                type="text"
                                                className="form-input"
                                                value={customerForm.country || 'India'}
                                                onChange={e => setCustomerForm({...customerForm, country: e.target.value})}
                                                placeholder="India"
                                            />
                                        </div>
                                        <div>
                                            <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: '#475569', marginBottom: '6px' }}>Zip / Postal Code</label>
                                            <input
                                                type="text"
                                                className="form-input"
                                                value={customerForm.zip_code || ''}
                                                onChange={e => setCustomerForm({...customerForm, zip_code: e.target.value})}
                                                placeholder="400001"
                                            />
                                        </div>
                                    </div>
                                </div>

                                <div style={{
                                    padding: '1.25rem',
                                    background: '#f8fafc',
                                    borderRadius: '12px',
                                    border: '1px solid #e2e8f0',
                                    marginBottom: '0.5rem'
                                }}>
                                    <label style={{ display: 'flex', alignItems: 'center', gap: '12px', cursor: 'pointer' }}>
                                        <div style={{
                                            position: 'relative',
                                            display: 'flex',
                                            alignItems: 'center',
                                            justifyContent: 'center'
                                        }}>
                                            <input
                                                type="checkbox"
                                                checked={syncToShopify}
                                                onChange={e => setSyncToShopify(e.target.checked)}
                                                style={{
                                                    width: '20px',
                                                    height: '20px',
                                                    cursor: 'pointer',
                                                    accentColor: '#0f172a'
                                                }}
                                            />
                                        </div>
                                        <div>
                                            <span style={{ fontWeight: 600, color: '#1e293b', fontSize: '0.95rem' }}>Sync with Shopify</span>
                                            <p style={{ margin: '2px 0 0', fontSize: '0.8rem', color: '#64748b' }}>
                                                {isEditMode
                                                    ? "Keep this customer updated on your Shopify store"
                                                    : "Automatically create this customer on your Shopify store"}
                                            </p>
                                        </div>
                                    </label>
                                </div>
                            </div>

                            {/* Sticky Footer */}
                            <div style={{
                                padding: '1.25rem 1.5rem',
                                borderTop: '1px solid #e2e8f0',
                                display: 'flex',
                                gap: '1rem',
                                justifyContent: 'flex-end',
                                background: '#f8fafc'
                            }}>
                                <button type="button" className="btn-secondary" onClick={() => setShowAddModal(false)} style={{ minWidth: '100px' }}>Cancel</button>
                                <button type="submit" className="btn-primary" disabled={isSaving} style={{ background: '#0f172a', minWidth: '140px' }}>
                                    {isSaving ? 'Saving...' : (isEditMode ? 'Update Customer' : 'Create Customer')}
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
}


const BulkTemplateModal: React.FC<{
    isOpen: boolean;
    onClose: () => void;
    customerIDs: (number | string)[];
    onSuccess: () => void;
    fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
    bulkSuffix?: string;
}> = ({ isOpen, onClose, customerIDs, onSuccess, fetchWithAuth, bulkSuffix = '_marketing' }) => {
    const [templates, setTemplates] = useState<WhatsAppTemplate[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [selectedTemplate, setSelectedTemplate] = useState<WhatsAppTemplate | null>(null);
    const [isSending, setIsSending] = useState(false);
    const [result, setResult] = useState<{ sent: number; total: number } | null>(null);

    useEffect(() => {
        if (isOpen) {
            fetchTemplates();
        } else {
            setResult(null);
            setSelectedTemplate(null);
        }
    }, [isOpen]);

    const fetchTemplates = async () => {
        setIsLoading(true);
        try {
            const response = await fetchWithAuth(`${API_BASE}/api/automation/whatsapp/templates`);
            const data = await response.json();
            // Filter only bulk templates
            const filteredTemplates = data.filter((t: WhatsAppTemplate) => {
                const name = (t.template_name || '').trim().toLowerCase();
                const suffix = (bulkSuffix || '').trim().toLowerCase();
                const status = (t.status || '').trim().toUpperCase();
                
                return name.endsWith(suffix) && status === 'APPROVED';
            });
            setTemplates(filteredTemplates);
        } catch (error) {
            console.error('Error fetching templates:', error);
        } finally {
            setIsLoading(false);
        }
    };

    const handleSend = async () => {
        if (!selectedTemplate) return;
        setIsSending(true);
        try {
            const response = await fetchWithAuth(`${API_BASE}/api/automation/whatsapp/send-bulk`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    customer_ids: customerIDs,
                    template_id: selectedTemplate.id
                })
            });
            const data = await response.json();
            setResult({ sent: data.sent, total: data.total });
            if (data.success) {
                setTimeout(() => {
                    onSuccess();
                    onClose();
                }, 2000);
            }
        } catch (error) {
            console.error('Error sending bulk messages:', error);
        } finally {
            setIsSending(false);
        }
    };

    if (!isOpen) return null;

    return (
        <div className="modal-overlay" onClick={onClose} style={{ zIndex: 1100 }}>
            <div className="modal-content" onClick={e => e.stopPropagation()} style={{ maxWidth: '600px', width: '90%' }}>
                <div className="modal-header">
                    <h2 style={{ fontSize: '1.25rem', fontWeight: 700, color: '#1e293b' }}>Send Bulk Messages</h2>
                    <button className="close-button" onClick={onClose} aria-label="Close modal">&times;</button>
                </div>

                <div className="modal-body" style={{ padding: '1.5rem' }}>
                    {result ? (
                        <div style={{ textAlign: 'center', padding: '2rem 1rem' }}>
                            <div style={{ fontSize: '3rem', marginBottom: '1rem' }}>✅</div>
                            <h3 style={{ fontSize: '1.25rem', fontWeight: 600, color: '#0f172a', marginBottom: '0.5rem' }}>Messages Sent!</h3>
                            <p style={{ color: '#64748b' }}>Successfully sent {result.sent} of {result.total} messages.</p>
                        </div>
                    ) : (
                        <>
                            <div style={{ background: '#f8fafc', padding: '1rem', borderRadius: '12px', marginBottom: '1.5rem', border: '1px solid #e2e8f0' }}>
                                <p style={{ fontSize: '0.9rem', color: '#64748b', margin: 0 }}>
                                    Targeting <span style={{ color: '#0f172a', fontWeight: 600 }}>{customerIDs.length}</span> selected customers.
                                </p>
                            </div>

                            <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: '#475569', marginBottom: '0.75rem' }}>
                                Select Bulk Template
                            </label>

                            {isLoading ? (
                                <div style={{ textAlign: 'center', padding: '2rem' }}>Loading templates...</div>
                            ) : templates.length === 0 ? (
                                <div style={{ textAlign: 'center', padding: '2rem', background: '#fef2f2', color: '#991b1b', borderRadius: '12px' }}>
                                    No templates found ending with <code style={{ background: '#fee2e2', padding: '2px 4px', borderRadius: '4px' }}>{bulkSuffix}</code> or none are approved yet.
                                </div>
                            ) : (
                                <div
                                    role="listbox"
                                    aria-label="Bulk message templates"
                                    style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', maxHeight: '400px', overflowY: 'auto', paddingRight: '0.5rem' }}
                                >
                                    {templates.map(t => (
                                        <button
                                            key={t.id}
                                            type="button"
                                            role="option"
                                            aria-selected={selectedTemplate?.id === t.id}
                                            onClick={() => setSelectedTemplate(t)}
                                            style={{
                                                padding: '1rem',
                                                borderRadius: '12px',
                                                border: '2px solid',
                                                borderColor: selectedTemplate?.id === t.id ? '#38bdf8' : '#e2e8f0',
                                                background: selectedTemplate?.id === t.id ? '#f0f9ff' : 'white',
                                                cursor: 'pointer',
                                                transition: 'all 0.2s',
                                                textAlign: 'left',
                                                width: '100%',
                                                fontFamily: 'inherit'
                                            }}
                                        >
                                            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
                                                <span style={{ fontWeight: 600, color: '#1e293b' }}>{t.template_name}</span>
                                                <span style={{ fontSize: '0.75rem', color: '#64748b', background: '#f1f5f9', padding: '2px 8px', borderRadius: '4px' }}>{t.language}</span>
                                            </div>
                                            <p style={{ fontSize: '0.85rem', color: '#64748b', margin: 0, display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', overflow: 'hidden' }}>
                                                {t.body}
                                            </p>
                                        </button>
                                    ))}
                                </div>
                            )}
                        </>
                    )}
                </div>

                <div className="modal-footer" style={{ borderTop: '1px solid #e2e8f0', padding: '1rem 1.5rem', display: 'flex', justifyContent: 'flex-end', gap: '1rem' }}>
                    {!result && (
                        <>
                            <button className="btn-secondary" onClick={onClose} disabled={isSending}>Cancel</button>
                                <button 
                                    className="btn-primary" 
                                    onClick={handleSend} 
                                    disabled={!selectedTemplate || isSending}
                                    style={{ background: '#38bdf8', color: '#0c4a6e', border: 'none' }}
                                >
                                    {isSending ? 'Sending...' : `Send to ${customerIDs.length} Customers`}
                                </button>
                        </>
                    )}
                </div>
            </div>
        </div>
    );
}
