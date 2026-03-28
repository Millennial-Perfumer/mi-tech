import { useState, useEffect, useRef } from 'react';
import { API_BASE } from './api';
import { useToast } from './ToastContext';
import { useConfirm } from './ConfirmContext';
import { ColumnSelector } from './ColumnSelector';
import type { ColumnOption } from './ColumnSelector';
import { TemplateMapper } from './TemplateMapper';

const availableColumns: ColumnOption[] = [
  { id: 'template_name', label: 'Template Name', category: 'Essential' },
  { id: 'category', label: 'Category', category: 'Essential' },
  { id: 'language', label: 'Language', category: 'Essential' },
  { id: 'status', label: 'Status', category: 'Status' },
  { id: 'created_at', label: 'Created Time', category: 'Status' },
  { id: 'sent_count', label: 'Sent', category: 'Metrics' },
  { id: 'delivered_count', label: 'Delivered', category: 'Metrics' },
  { id: 'read_count', label: 'Read', category: 'Metrics' },
];

export interface Template {
  id: number;
  template_name: string;
  category: string;
  language: string;
  body: string;
  header?: any;
  footer?: string;
  buttons?: any;
  status: string;
  created_at: string;
  sent_count: number;
  delivered_count: number;
  read_count: number;
  variable_mappings?: any;
}

interface AutomationTemplatesProps {
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
  userRole?: string;
}

interface MultiSelectFilterProps {
  label: string;
  options: string[];
  selectedOptions: string[];
  onChange: (options: string[]) => void;
  icon?: React.ReactNode;
}

function MultiSelectFilter({ label, options, selectedOptions, onChange, icon }: MultiSelectFilterProps) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const toggleOption = (option: string) => {
    if (selectedOptions.includes(option)) {
      if (selectedOptions.length > 1) {
        onChange(selectedOptions.filter(o => o !== option));
      }
    } else {
      onChange([...selectedOptions, option]);
    }
  };

  const selectAll = () => {
    if (selectedOptions.length === options.length) {
      if (options.length > 0) onChange([options[0]]); 
    } else {
      onChange([...options]);
    }
  };

  return (
    <div className="column-selector" ref={dropdownRef} style={{ position: 'relative' }}>
      <button 
        className="btn-secondary" 
        onClick={() => setIsOpen(!isOpen)} 
        style={{ 
          display: 'flex', 
          alignItems: 'center', 
          gap: '0.6rem', 
          padding: '0.5rem 1rem', 
          fontSize: '0.875rem', 
          fontWeight: 600,
          height: '42px',
          borderRadius: '10px',
          border: '1px solid var(--border-color)',
          backgroundColor: isOpen ? 'var(--bg-hover)' : 'var(--surface-color)',
          transition: 'all 0.2s ease',
          boxShadow: 'var(--shadow-sm)',
          whiteSpace: 'nowrap'
        }}
      >
        {icon || (
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"></polygon>
          </svg>
        )}
        {selectedOptions.length === options.length ? `All ${label}s` : `${selectedOptions.length} ${label}${selectedOptions.length > 1 ? (label.endsWith('s') ? 'es' : 's') : ''}`}
      </button>

      {isOpen && (
        <div className="column-selector-dropdown" style={{ right: 0, minWidth: '180px', zIndex: 100 }}>
          <div className="column-category">
            <div className="column-category-title" style={{ color: 'var(--text-primary)' }}>{label}</div>
            <label className="column-option" style={{ borderBottom: '1px solid var(--border-color)', marginBottom: '8px', paddingBottom: '8px' }}>
              <input
                type="checkbox"
                checked={selectedOptions.length === options.length}
                onChange={selectAll}
              />
              <strong>Select All</strong>
            </label>
            <div style={{ maxHeight: '200px', overflowY: 'auto' }}>
              {options.map(o => (
                <label key={o} className="column-option">
                  <input
                    type="checkbox"
                    checked={selectedOptions.includes(o)}
                    onChange={() => toggleOption(o)}
                  />
                  {o.charAt(0) + o.slice(1).toLowerCase().replace(/_/g, ' ')}
                </label>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export function AutomationTemplates({ fetchWithAuth, userRole = 'read' }: AutomationTemplatesProps) {
  const { success: toastSuccess, error: toastError } = useToast();
  const { confirm: customConfirm } = useConfirm();
  const [templates, setTemplates] = useState<Template[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSyncing, setIsSyncing] = useState(false);
  const [selectedTemplate, setSelectedTemplate] = useState<Template | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedStatuses, setSelectedStatuses] = useState<string[]>(['APPROVED', 'PENDING', 'REJECTED']);
  const [selectedCategories, setSelectedCategories] = useState<string[]>([]);
  const [selectedLanguages, setSelectedLanguages] = useState<string[]>([]);
  
  const [showFetchModal, setShowFetchModal] = useState(false);
  const [fetchTemplateName, setFetchTemplateName] = useState('');
  const [isFetchingFromMeta, setIsFetchingFromMeta] = useState(false);
  
  const [visibleColumns, setVisibleColumns] = useState<string[]>(() => {
    const saved = localStorage.getItem('gstAutomationVisibleCols');
    return saved ? JSON.parse(saved) : ['template_name', 'category', 'language', 'status', 'created_at'];
  });

  useEffect(() => {
    localStorage.setItem('gstAutomationVisibleCols', JSON.stringify(visibleColumns));
  }, [visibleColumns]);

  const fetchTemplates = async (silent = false) => {
    if (!silent) setIsLoading(true);
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/automation/whatsapp/templates`);
      const data = await resp.json();
      const loadedTemplates = Array.isArray(data) ? data : [];
      setTemplates(loadedTemplates);
      
      // Initialize filters if empty
      if (selectedCategories.length === 0) {
        const cats = Array.from(new Set(loadedTemplates.map(t => t.category.toUpperCase())));
        setSelectedCategories(cats);
      }
      if (selectedLanguages.length === 0) {
        const langs = Array.from(new Set(loadedTemplates.map(t => t.language)));
        setSelectedLanguages(langs);
      }
    } catch (err) {
      console.error('Failed to fetch templates:', err);
      setTemplates([]);
    } finally {
      setIsLoading(false);
    }
  };

  const handleFetchFromMeta = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!fetchTemplateName) return;

    setIsFetchingFromMeta(true);
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/automation/whatsapp/templates/sync-single?name=${fetchTemplateName}`, { method: 'POST' });
      if (resp.ok) {
        toastSuccess('Template successfully imported from Meta!');
        fetchTemplates(true);
        setShowFetchModal(false);
        setFetchTemplateName('');
      } else {
        const err = await resp.text();
        toastError(`Failed to import template: ${err}`);
      }
    } catch(err) {
      console.error('Failed to single sync', err);
      toastError('Network error while importing template.');
    } finally {
      setIsFetchingFromMeta(false);
    }
  };

  const handleSyncAll = async () => {
    const confirmed = await customConfirm({
      title: 'Full Template Synchronization',
      message: 'This will fetch all templates from your Meta Business Account. Depending on the number of templates, this may take a few seconds. Do you want to continue?',
      confirmLabel: 'Sync All'
    });

    if (!confirmed) return;
    
    setIsSyncing(true);
    try {
      const resp = await fetchWithAuth(`${API_BASE}/api/automation/whatsapp/templates/sync-all`, {
        method: 'POST'
      });
      if (resp.ok) {
        toastSuccess('Successfully synced all templates from Meta.');
        fetchTemplates(true);
      } else {
        const errText = await resp.text();
        toastError(`Sync failed: ${errText}`);
      }
    } catch (err) {
      console.error('Error syncing:', err);
      toastError('Network error while syncing templates.');
    } finally {
      setIsSyncing(false);
    }
  };

  useEffect(() => {
    if (!selectedTemplate) {
      fetchTemplates();
    }
  }, [selectedTemplate]);

  if (selectedTemplate) {
    return (
      <TemplateMapper 
        template={selectedTemplate} 
        onBack={() => setSelectedTemplate(null)}
        fetchWithAuth={fetchWithAuth}
      />
    );
  }

  const getStatusBadgeStyle = (status: string) => {
    switch (status.toUpperCase()) {
      case 'APPROVED': return { backgroundColor: '#dcfce7', color: '#166534', border: '1px solid #bbf7d0' };
      case 'PENDING': return { backgroundColor: '#fef9c3', color: '#854d0e', border: '1px solid #fef08a' };
      case 'REJECTED': return { backgroundColor: '#fee2e2', color: '#991b1b', border: '1px solid #fecaca' };
      case 'ARCHIVED': return { backgroundColor: '#f1f5f9', color: '#475569', border: '1px solid #e2e8f0' };
      default: return { backgroundColor: '#f1f5f9', color: '#475569' };
    }
  };

  const checkMappingStatus = (t: Template) => {
    const bodyVarCount = countRequiredParams(t.body);
    let headerTextCount = 0;
    let headerType = 'NONE';
    if (t.header) {
      const headerObj = typeof t.header === 'string' ? JSON.parse(t.header) : t.header;
      headerType = headerObj.type?.toUpperCase() || 'NONE';
      if (headerType === 'TEXT' && headerObj.text) {
        headerTextCount = countRequiredParams(headerObj.text);
      }
    }
    const isHeaderDynamic = (headerType === 'TEXT' && headerTextCount > 0) || ['DOCUMENT', 'IMAGE', 'VIDEO'].includes(headerType);
    
    let buttons: any[] = [];
    if (t.buttons) {
      buttons = typeof t.buttons === 'string' ? JSON.parse(t.buttons) : t.buttons;
    }
    const hasDynamicButtons = buttons.some(b => b.type === 'visit_website' && (b.url || '').includes('{{1}}'));
    
    const needsMapping = bodyVarCount > 0 || isHeaderDynamic || hasDynamicButtons;
    if (!needsMapping) return 'NOT_NEEDED';

    // If it needs mapping, check if mappings are present and non-empty
    const mappings = t.variable_mappings || {};
    
    // Very basic check: are there any mappings? Or do we need more advanced validation?
    // For now, if it needs mapping but has 0 or few mappings, it's a warning.
    if (Object.keys(mappings).length === 0) return 'REQUIRED';
    
    // Check if body params are all mapped
    for (let i = 1; i <= bodyVarCount; i++) {
        if (!mappings[`body_text_0_{{${i}}}`]) return 'INCOMPLETE';
    }
    
    return 'COMPLETE';
  };

  return (
    <div className="automation-section" style={{ padding: '2rem', backgroundColor: 'var(--bg-color)', minHeight: '100vh' }}>
      {/* Header Area */}
      <div className="section-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '2.5rem' }}>
        <div>
          <h2 style={{ fontSize: '2rem', fontWeight: 850, color: 'var(--text-primary)', letterSpacing: '-0.025em', marginBottom: '0.5rem' }}>
            WhatsApp Templates
          </h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '1.05rem', maxWidth: '600px', lineHeight: '1.5' }}>
            Centralized hub for managing your Meta WhatsApp templates. Sync directly from Meta and map variables to your local data.
          </p>
        </div>
        {userRole === 'admin' && (
          <div style={{ display: 'flex', gap: '0.75rem' }}>
            <button 
              className="btn-secondary" 
              onClick={() => setShowFetchModal(true)}
              style={{ fontWeight: 600, display: 'flex', alignItems: 'center', gap: '0.5rem', height: '42px', boxShadow: '0 1px 2px rgba(0,0,0,0.05)' }}
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
              Import by Name
            </button>
            <button 
              className="btn-primary" 
              onClick={handleSyncAll}
              disabled={isSyncing}
              style={{ display: 'flex', alignItems: 'center', gap: '0.6rem', height: '42px', fontWeight: 600, backgroundImage: 'linear-gradient(to bottom, #00C3F2, #00A8E0)', boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)' }}
            >
              {isSyncing ? (
                <>
                  <span className="spinner" style={{ width: '16px', height: '16px', border: '2.5px solid rgba(255,255,255,0.3)', borderTopColor: '#fff', borderRadius: '50%', animation: 'spin 1s linear infinite' }}></span>
                  Synchronizing...
                </>
              ) : (
                <>
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.3"/>
                  </svg>
                  Sync All from Meta
                </>
              )}
            </button>
          </div>
        )}
      </div>

      {/* Unified Toolbar */}
      <div style={{ 
        display: 'flex', 
        justifyContent: 'space-between', 
        alignItems: 'center', 
        marginBottom: '1.5rem', 
        backgroundColor: 'var(--surface-color)',
        padding: '0.75rem 1rem', 
        borderRadius: '16px', 
        boxShadow: 'var(--shadow-sm)',
        border: '1px solid var(--border-color)'
      }}>
        <div style={{ position: 'relative', width: '350px' }}>
          <svg 
            style={{ position: 'absolute', left: '14px', top: '50%', transform: 'translateY(-50%)', color: '#94a3b8' }}
            width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
          >
            <circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line>
          </svg>
          <input 
            type="text"
            placeholder="Search within filtered results..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            style={{ 
              paddingLeft: '42px', 
              borderRadius: '10px', 
              border: '1px solid var(--border-color)',
              fontSize: '0.95rem',
              width: '100%',
              height: '42px',
              backgroundColor: 'var(--bg-input)',
              color: 'var(--text-primary)'
            }}
          />
        </div>
        
        <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center', flexWrap: 'nowrap' }}>
          <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
            <MultiSelectFilter 
              label="Status"
              options={['APPROVED', 'PENDING', 'REJECTED', 'ARCHIVED']}
              selectedOptions={selectedStatuses}
              onChange={setSelectedStatuses}
            />
            <MultiSelectFilter 
              label="Type"
              options={Array.from(new Set(templates.map(t => t.category.toUpperCase())))}
              selectedOptions={selectedCategories}
              onChange={setSelectedCategories}
              icon={<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="3" width="7" height="7"></rect><rect x="14" y="3" width="7" height="7"></rect><rect x="14" y="14" width="7" height="7"></rect><rect x="3" y="14" width="7" height="7"></rect></svg>}
            />
            <MultiSelectFilter 
              label="Language"
              options={Array.from(new Set(templates.map(t => t.language))).sort()}
              selectedOptions={selectedLanguages}
              onChange={setSelectedLanguages}
              icon={<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M2 12h20"></path><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"></path><path d="M2 12a15.3 15.3 0 0 1 10-4 15.3 15.3 0 0 1 10 4"></path><path d="M2 12a15.3 15.3 0 0 0 10 4 15.3 15.3 0 0 0 10-4"></path></svg>}
            />
          </div>
          <ColumnSelector 
            columns={availableColumns}
            visibleColumns={visibleColumns}
            onChange={setVisibleColumns}
          />
        </div>
      </div>

      <div className="table-container" style={{ 
        backgroundColor: 'var(--surface-color)',
        borderRadius: '16px', 
        border: '1px solid var(--border-color)',
        boxShadow: 'var(--shadow-sm)',
        overflow: 'hidden'
      }}>
        <table style={{ minWidth: '1000px', borderCollapse: 'collapse' }}>
          <thead style={{ backgroundColor: 'var(--bg-input)', borderBottom: '1px solid var(--border-color)' }}>
            <tr style={{ textAlign: 'left' }}>
              {visibleColumns.includes('template_name') && <th style={{ padding: '1rem', fontSize: '0.75rem', fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Template Name</th>}
              {visibleColumns.includes('category') && <th style={{ padding: '1rem', fontSize: '0.75rem', fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Category</th>}
              {visibleColumns.includes('language') && <th style={{ padding: '1rem', fontSize: '0.75rem', fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Language</th>}
              {visibleColumns.includes('sent_count') && <th style={{ padding: '1rem', fontSize: '0.75rem', fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', textAlign: 'center' }}>Sent</th>}
              {visibleColumns.includes('delivered_count') && <th style={{ padding: '1rem', fontSize: '0.75rem', fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', textAlign: 'center' }}>Deliv.</th>}
              {visibleColumns.includes('read_count') && <th style={{ padding: '1rem', fontSize: '0.75rem', fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', textAlign: 'center' }}>Read</th>}
              {visibleColumns.includes('status') && <th style={{ padding: '1rem', fontSize: '0.75rem', fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Status</th>}
              {visibleColumns.includes('created_at') && <th style={{ padding: '1rem', fontSize: '0.75rem', fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Created</th>}
              <th style={{ padding: '1rem', fontSize: '0.75rem', fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr><td colSpan={10} style={{ textAlign: 'center', padding: '2rem' }}>Loading templates...</td></tr>
            ) : templates.filter(t => 
                t.template_name.toLowerCase().includes(searchTerm.toLowerCase()) && 
                selectedStatuses.includes((t.status || 'PENDING').toUpperCase()) &&
                selectedCategories.includes(t.category.toUpperCase()) &&
                selectedLanguages.includes(t.language)
              ).length === 0 ? (
              <tr><td colSpan={10} style={{ textAlign: 'center', padding: '2rem' }}>No templates matching your criteria found.</td></tr>
            ) : (
              templates
                .filter(t => 
                  t.template_name.toLowerCase().includes(searchTerm.toLowerCase()) && 
                  selectedStatuses.includes((t.status || 'PENDING').toUpperCase()) &&
                  selectedCategories.includes(t.category.toUpperCase()) &&
                  selectedLanguages.includes(t.language)
                )
                .sort((a, b) => {
                  const order: Record<string, number> = { 'APPROVED': 1, 'PENDING': 2, 'REJECTED': 3, 'ARCHIVED': 4 };
                  const valA = order[(a.status || 'PENDING').toUpperCase()] || 5;
                  const valB = order[(b.status || 'PENDING').toUpperCase()] || 5;
                  if (valA !== valB) return valA - valB;
                  return a.template_name.localeCompare(b.template_name);
                })
                .map(t => (
                  <tr key={t.id}>
                  {visibleColumns.includes('template_name') && <td><strong style={{ color: 'var(--text-primary)' }}>{t.template_name}</strong></td>}
                  {visibleColumns.includes('category') && <td><span className="badge" style={{ backgroundColor: 'var(--bg-input)', color: 'var(--text-secondary)' }}>{t.category}</span></td>}
                  {visibleColumns.includes('language') && <td>{t.language}</td>}
                  {visibleColumns.includes('sent_count') && <td style={{ textAlign: 'center' }}><strong>{t.sent_count || 0}</strong></td>}
                  {visibleColumns.includes('delivered_count') && <td style={{ textAlign: 'center' }}><span style={{ color: t.delivered_count > 0 ? '#00a884' : 'inherit' }}>{t.delivered_count || 0}</span></td>}
                  {visibleColumns.includes('read_count') && <td style={{ textAlign: 'center' }}><span style={{ color: t.read_count > 0 ? '#34b7f1' : 'inherit' }}>{t.read_count || 0}</span></td>}
                  {visibleColumns.includes('status') && (
                    <td>
                      <span className="badge" style={{ 
                        ...getStatusBadgeStyle(t.status || 'PENDING'), 
                        borderRadius: '9999px', 
                        padding: '0.2rem 0.75rem', 
                        fontSize: '0.75rem', 
                        fontWeight: 600,
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: '4px'
                      }}>
                        <span style={{ width: '6px', height: '6px', borderRadius: '50%', backgroundColor: 'currentColor' }}></span>
                        {(t.status || 'PENDING').toUpperCase()}
                      </span>
                    </td>
                  )}
                  {visibleColumns.includes('created_at') && <td style={{ fontSize: '0.85rem' }}>{new Date(t.created_at).toLocaleDateString()}</td>}
                  <td>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      <button 
                        className="btn-secondary"
                        style={{ padding: '0.5rem', display: 'flex', justifyContent: 'center', alignItems: 'center', borderRadius: '10px' }}
                        onClick={() => setSelectedTemplate(t)}
                        title={t.variable_mappings && Object.keys(t.variable_mappings).length > 0 ? 'Edit Mapping' : 'Map Variables'}
                      >
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                          <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"></path>
                          <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"></path>
                        </svg>
                      </button>
                      {['REQUIRED', 'INCOMPLETE'].includes(checkMappingStatus(t)) && t.status !== 'ARCHIVED' && (
                        <div title="Mapping Required" style={{ color: '#f59e0b', display: 'flex' }}>
                          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                            <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
                            <line x1="12" y1="9" x2="12" y2="13"></line>
                            <line x1="12" y1="17" x2="12.01" y2="17"></line>
                          </svg>
                        </div>
                      )}
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {showFetchModal && (
        <div className="modal-overlay">
          <div className="modal-content" style={{ maxWidth: '450px' }}>
            <div className="modal-header">
              <h3 className="modal-title">Fetch Template from Meta</h3>
            </div>
            <form onSubmit={handleFetchFromMeta}>
              <div className="form-group" style={{ marginBottom: '1.5rem' }}>
                <label>Template Name</label>
                <input 
                  type="text" 
                  value={fetchTemplateName} 
                  onChange={e => setFetchTemplateName(e.target.value.toLowerCase().trim())}
                  placeholder="e.g. shipping_update_v1"
                  required
                  autoFocus
                />
                <p style={{ fontSize: '0.8rem', color: '#64748b', marginTop: '0.5rem' }}>
                  Enter the exact name of the template as it appears in your Meta WhatsApp Manager.
                </p>
              </div>
              <div className="modal-footer" style={{ borderTop: 'none', padding: 0, display: 'flex', gap: '1rem', justifyContent: 'flex-end' }}>
                <button type="button" className="btn-secondary" onClick={() => setShowFetchModal(false)} disabled={isFetchingFromMeta}>Cancel</button>
                <button type="submit" className="btn-primary" disabled={isFetchingFromMeta || !fetchTemplateName}>
                  {isFetchingFromMeta ? 'Fetching...' : 'Continue'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

function countRequiredParams(body: string) {
  if (!body) return 0;
  const re = /\{\{(\d+)\}\}/g;
  let max = 0;
  let match;
  while ((match = re.exec(body)) !== null) {
    const n = parseInt(match[1]);
    if (n > max) { max = n; }
  }
  return max;
}
