import { useState } from 'react';
import { API_BASE } from './api';
import type { Template } from './AutomationTemplates';

interface TemplateMapperProps {
  template: Template;
  onBack: () => void;
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
}

// Available mapping options based on entity.Order and entity.Customer
const availableVariables = [
  { value: '', label: '-- Not Mapped --' },
  { value: 'customer_name', label: 'Customer Name' },
  { value: 'customer_email', label: 'Customer Email' },
  { value: 'customer_phone', label: 'Customer Phone' },
  { value: 'customer_city', label: 'Customer City' },
  { value: 'customer_state', label: 'Customer State' },
  { value: 'customer_country', label: 'Customer Country' },
  { value: 'customer_zip', label: 'Customer Zip Code' },
  { value: 'customer_address1', label: 'Customer Address 1' },
  { value: 'customer_address2', label: 'Customer Address 2' },
  { value: 'customer_total_orders', label: 'Total Orders' },
  { value: 'customer_total_spent', label: 'Total Spent' },
  { value: 'order_id', label: 'Order Number (e.g. 1001)' },
  { value: 'internal_order_id', label: 'Internal DB Order ID' },
  { value: 'order_total', label: 'Order Total' },
  { value: 'tracking_link', label: 'Tracking Link' },
  { value: 'tracking_number', label: 'Tracking Number' },
  { value: 'shipping_company', label: 'Shipping Company' },
];

export function TemplateMapper({ template, onBack, fetchWithAuth }: TemplateMapperProps) {
  const [mappings, setMappings] = useState<Record<string, string>>(template.variable_mappings || {});
  const [isSaving, setIsSaving] = useState(false);
  
  // Parse required parameters from body
  const bodyVarCount = countRequiredParams(template.body);
  
  let headerType = 'NONE';
  let headerTextCount = 0;
  let headerText = '';
  if (template.header) {
    let headerObj = typeof template.header === 'string' ? JSON.parse(template.header) : template.header;
    headerType = headerObj.type?.toUpperCase() || 'NONE';
    if (headerType === 'TEXT' && headerObj.text) {
      headerText = headerObj.text;
      headerTextCount = countRequiredParams(headerObj.text);
    }
  }

  const isHeaderDynamic = (headerType === 'TEXT' && headerTextCount > 0) || 
                          (['DOCUMENT', 'IMAGE', 'VIDEO'].includes(headerType));

  // Buttons Parse
  let buttons: any[] = [];
  if (template.buttons) {
    buttons = typeof template.buttons === 'string' ? JSON.parse(template.buttons) : template.buttons;
  }

  const hasDynamicButtons = buttons.some(b => b.type === 'visit_website' && (b.url || '').includes('{{1}}'));
  const hasVariables = isHeaderDynamic || bodyVarCount > 0 || hasDynamicButtons;
  const isDirty = JSON.stringify(mappings) !== JSON.stringify(template.variable_mappings || {});

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSaving(true);
    try {
      // Instead of relying on a dedicated endpoint, we update the template through the standard PUT endpoint
      // We need to pass the whole template + mappings
      const updatedTemplate = {
        ...template,
        variable_mappings: mappings
      };

      const resp = await fetchWithAuth(`${API_BASE}/api/automation/whatsapp/templates`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updatedTemplate)
      });

      if (resp.ok) {
        alert('Variable mappings saved successfully!');
        onBack();
      } else {
        const err = await resp.text();
        alert(`Failed to save mappings: ${err}`);
      }
    } catch (err) {
      console.error(err);
      alert('Network error while saving mappings.');
    } finally {
      setIsSaving(false);
    }
  };

  const handleMappingChange = (key: string, value: string) => {
    setMappings(prev => ({
      ...prev,
      [key]: value
    }));
  };

  return (
    <div className="template-mapper" style={{ padding: '2rem', backgroundColor: 'var(--bg-color)', minHeight: '100vh' }}>
      <div className="section-header" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '2rem' }}>
        <div>
          <h3 style={{ fontSize: '0.9rem', color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '1.25rem', display: 'flex', alignItems: 'center', gap: '8px' }}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
            Variables: {template.template_name}
          </h3>
          <p style={{ color: 'var(--text-secondary)', fontSize: '1rem' }}>Map your local system data to Meta's dynamic parameters.</p>
        </div>
        <button className="btn-secondary" onClick={onBack} style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontWeight: 600 }}>
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="19" y1="12" x2="5" y2="12"></line><polyline points="12 19 5 12 12 5"></polyline></svg>
          Back to List
        </button>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 340px', gap: '2.5rem', alignItems: 'start' }}>
        
        {/* Mapping Form */}
        <div style={{ backgroundColor: 'var(--surface-color)', padding: '1.5rem', borderRadius: '12px', border: '1px solid var(--border-color)' }}>
          <form onSubmit={handleSave}>
            
            {/* Header Mapping Section */}
            {isHeaderDynamic && (
              <div style={{ marginBottom: '2rem' }}>
                <h3 style={{ fontSize: '1.1rem', fontWeight: 600, borderBottom: '1px solid var(--border-color)', paddingBottom: '0.5rem', marginBottom: '1rem', color: 'var(--text-primary)' }}>Header ({headerType})</h3>
                
                {headerType === 'TEXT' && headerTextCount > 0 && Array.from({ length: headerTextCount }).map((_, i) => (
                  <div key={`header_${i+1}`} className="form-group" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <label style={{ margin: 0 }}>Header Param {`{{${i+1}}}`}</label>
                    <select 
                      value={mappings[`header_text_0_{{${i+1}}}`] || ''}
                      onChange={(e) => handleMappingChange(`header_text_0_{{${i+1}}}`, e.target.value)}
                      style={{ width: '250px' }}
                    >
                      {availableVariables.map(v => <option key={v.value} value={v.value}>{v.label}</option>)}
                    </select>
                  </div>
                ))}

                {(headerType === 'DOCUMENT' || headerType === 'IMAGE' || headerType === 'VIDEO') && (
                  <div className="form-group" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <div>
                      <label style={{ margin: 0 }}>Dynamic Allocation</label>
                      <p style={{ margin: 0, fontSize: '0.75rem', color: 'var(--text-secondary)' }}>Provide a static Media ID, or select Dynamic Invoice to auto-generate.</p>
                    </div>
                    {/* The handle maps to either a Media ID string or predefined handle like 'Dynamic Invoice' */}
                    {headerType === 'DOCUMENT' ? (
                      <select 
                        value={mappings['header_handle'] || ''}
                        onChange={(e) => handleMappingChange('header_handle', e.target.value)}
                        style={{ width: '250px' }}
                      >
                        <option value="">-- No Source --</option>
                        <option value="Dynamic Invoice">Auto-Generate PDF Invoice</option>
                        {/* We could allow users to type a raw ID if needed using a switch, but dropdown is cleaner for now */}
                      </select>
                    ) : (
                      <input 
                        type="text" 
                        placeholder="Enter Meta Media ID"
                        value={mappings['header_handle'] || ''}
                        onChange={(e) => handleMappingChange('header_handle', e.target.value)}
                        style={{ width: '250px' }}
                      />
                    )}
                  </div>
                )}
              </div>
            )}

            {/* Body Mapping Section */}
            {bodyVarCount > 0 && (
              <div style={{ marginBottom: '2rem' }}>
                <h3 style={{ fontSize: '1.1rem', fontWeight: 600, borderBottom: '1px solid var(--border-color)', paddingBottom: '0.5rem', marginBottom: '1rem', color: 'var(--text-primary)' }}>Body Parameters</h3>
                {Array.from({ length: bodyVarCount }).map((_, i) => (
                  <div key={`body_${i+1}`} className="form-group" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
                    <label style={{ margin: 0, fontWeight: 600, color: '#0ea5e9' }}>{`{{${i+1}}}`}</label>
                    <select 
                      value={mappings[`body_text_0_{{${i+1}}}`] || ''}
                      onChange={(e) => handleMappingChange(`body_text_0_{{${i+1}}}`, e.target.value)}
                      style={{ width: '250px' }}
                      required
                    >
                      {availableVariables.map(v => <option key={v.value} value={v.value}>{v.label}</option>)}
                    </select>
                  </div>
                ))}
              </div>
            )}

            {/* Buttons Mapping Section */}
            {buttons.length > 0 && buttons.some(b => b.type === 'visit_website' && (b.url || '').includes('{{1}}')) && (
              <div style={{ marginBottom: '2rem' }}>
                <h3 style={{ fontSize: '1.1rem', fontWeight: 600, borderBottom: '1px solid var(--border-color)', paddingBottom: '0.5rem', marginBottom: '1rem', color: 'var(--text-primary)' }}>Dynamic Buttons</h3>
                {buttons.map((btn, i) => {
                  if (btn.type === 'visit_website' && (btn.url || '').includes('{{1}}')) {
                    return (
                      <div key={`btn_${i}`} className="form-group" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <div>
                          <label style={{ margin: 0 }}>URL Loop Variable (Button {i+1})</label>
                          <p style={{ margin: 0, fontSize: '0.75rem', color: 'var(--text-secondary)' }}>{btn.url}</p>
                        </div>
                        <select 
                          value={mappings[`button_url_${i}_{{1}}`] || ''}
                          onChange={(e) => handleMappingChange(`button_url_${i}_{{1}}`, e.target.value)}
                          style={{ width: '250px' }}
                          required
                        >
                          {availableVariables.map(v => <option key={v.value} value={v.value}>{v.label}</option>)}
                        </select>
                      </div>
                    );
                  }
                  return null;
                })}
              </div>
            )}

            {!hasVariables && (
              <div style={{ padding: '4rem 2rem', textAlign: 'center', borderRadius: '16px', background: 'var(--accent-subtle)', border: '1px dashed var(--border-color)' }}>
                <div style={{ fontSize: '3rem', marginBottom: '1rem' }}>✨</div>
                <h3 style={{ fontSize: '1.25rem', fontWeight: 700, color: 'var(--accent-color)', marginBottom: '0.5rem' }}>Ready to Send!</h3>
                <p style={{ color: 'var(--text-secondary)', fontSize: '0.95rem', maxWidth: '300px', margin: '0 auto' }}>
                  This template contains only static content and doesn't require any variable mapping.
                </p>
              </div>
            )}

            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '1rem', marginTop: '2rem', paddingTop: '1.5rem', borderTop: '1px solid var(--border-color)' }}>
              <button type="button" className="btn-secondary" onClick={onBack} disabled={isSaving}>
                {hasVariables ? 'Cancel' : 'Back'}
              </button>
              {hasVariables && (
                <button type="submit" className="btn-primary" disabled={isSaving || !isDirty}>
                  {isSaving ? 'Saving Mappings...' : 'Save Configuration'}
                </button>
              )}
            </div>
          </form>
        </div>

        {/* Read-Only Preview */}
        <div style={{ position: 'sticky', top: '2rem' }}>
          <div style={{ 
            backgroundColor: 'var(--bg-color)', 
            borderRadius: '32px', 
            padding: '12px', 
            boxShadow: 'var(--shadow-lg)',
            border: '4px solid var(--border-color)'
          }}>
            {/* Status Bar */}
            <div style={{ display: 'flex', justifyContent: 'space-between', padding: '4px 20px 8px', color: 'white', fontSize: '0.7rem', fontWeight: 600 }}>
              <span>12:45</span>
              <div style={{ display: 'flex', gap: '4px' }}>
                <svg width="12" height="12" viewBox="0 0 24 24" fill="white"><path d="M12 21l-12-18h24z"/></svg>
                <svg width="12" height="12" viewBox="0 0 24 24" fill="white"><path d="M13 3h-2v10h2v-10zm4.846 1.55l-1.414 1.414c1.111 1.111 1.8 2.644 1.8 4.336 0 3.397-2.753 6.15-6.15 6.15s-6.15-2.753-6.15-6.15c0-1.692.689-3.225 1.8-4.336l-1.414-1.414c-1.478 1.478-2.386 3.518-2.386 5.75 0 4.562 3.688 8.25 8.25 8.25s8.25-3.688 8.25-8.25c0-2.232-.908-4.272-2.386-5.75z"/></svg>
              </div>
            </div>
            
            <div style={{ 
              backgroundColor: '#0b141a', 
            borderRadius: '24px', 
            padding: '1rem', 
            backgroundImage: 'url("https://user-images.githubusercontent.com/15075759/28719144-86dc0f70-73b1-11e7-911d-60d70fcded21.png")', 
            backgroundSize: 'cover',
            opacity: 0.95,
            minHeight: '400px'
          }}>
            <div style={{ position: 'relative', maxWidth: '90%' }}>
              <div style={{ backgroundColor: '#005c4b', borderRadius: '8px 8px 8px 0', overflow: 'hidden', boxShadow: '0 1px 1px rgba(0,0,0,0.15)' }}>
                  
                  {headerType !== 'NONE' && headerType !== 'TEXT' && (
                    <div style={{ backgroundColor: 'var(--bg-input)', height: '140px', display: 'flex', alignItems: 'center', justifyContent: 'center', overflow: 'hidden', color: 'var(--text-secondary)', fontWeight: 600 }}>
                      [{headerType} ATTACHMENT]
                    </div>
                  )}
                  
                  <div style={{ padding: '8px 12px' }}>
                    {headerType === 'TEXT' && (
                      <div style={{ fontWeight: 700, fontSize: '0.9rem', marginBottom: '4px', color: 'var(--text-tertiary)' }}>
                        {headerText.replace(/\{\{(\d+)\}\}/g, (match, id) => {
                          const mappedVar = mappings[`header_text_0_{{${id}}}`];
                          return mappedVar ? `[${availableVariables.find(v => v.value === mappedVar)?.label || mappedVar}]` : match;
                        })}
                      </div>
                    )}
                    <div style={{ whiteSpace: 'pre-wrap', fontSize: '0.85rem', color: 'var(--text-tertiary)', lineHeight: '1.4' }}>
                      {template.body.replace(/\{\{(\d+)\}\}/g, (match, id) => {
                        const mappedVar = mappings[`body_text_0_{{${id}}}`];
                        return mappedVar ? `[${availableVariables.find(v => v.value === mappedVar)?.label || mappedVar}]` : match;
                      })}
                    </div>
                    {template.footer && <div style={{ fontSize: '0.72rem', color: 'rgba(233, 237, 239, 0.6)', marginTop: '4px' }}>{template.footer}</div>}
                    <div style={{ textAlign: 'right', fontSize: '0.65rem', color: 'rgba(233, 237, 239, 0.4)', marginTop: '2px' }}>12:45 PM</div>
                  </div>

                  {buttons.length > 0 && (
                    <div style={{ borderTop: '1px solid #f0f2f5' }}>
                      {buttons.map((b, idx) => (
                        <div key={idx} style={{ padding: '8px', textAlign: 'center', color: '#34b7f1', fontSize: '0.85rem', fontWeight: 600, borderBottom: idx < buttons.length - 1 ? '1px solid #f0f2f5' : 'none', cursor: 'pointer' }}>
                          {b.type === 'visit_website' && (b.url || '').includes('{{1}}') ? (
                            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '6px' }}>
                              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path><polyline points="15 3 21 3 21 9"></polyline><line x1="10" y1="14" x2="21" y2="3"></line></svg>
                              {b.text}
                            </div>
                          ) : (
                            b.text || 'Button'
                          )}
                        </div>
                      ))}
                    </div>
                  )}
                </div>
                {/* SVG Bubble Tail */}
                <svg style={{ position: 'absolute', left: '-8px', bottom: '0', color: '#005c4b' }} width="8" height="13" viewBox="0 0 8 13">
                  <path fill="currentColor" d="M1.533 3.568L8 12.193V1H2.812C1.042 1 .474 2.156 1.533 3.568z"></path>
                </svg>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

// Helper func to get variable count from text
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
