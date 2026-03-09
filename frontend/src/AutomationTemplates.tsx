import { useState, useEffect, useRef } from 'react';

interface TemplateHeader {
  type: 'none' | 'IMAGE' | 'VIDEO' | 'DOCUMENT' | 'LOCATION';
  sample?: string;
  location?: { lat: string; lng: string; name: string; address: string };
}

interface TemplateButton {
  type: 'custom' | 'visit_website' | 'call_whatsapp' | 'call_phone' | 'flow' | 'copy_code';
  text: string;
  payload?: string;
  url?: string;
  phoneNumber?: string;
  flowID?: string;
  flowName?: string;
  offerCode?: string;
}

interface Template {
  id: number;
  template_name: string;
  category: string;
  language: string;
  body: string;
  header?: TemplateHeader;
  footer?: string;
  buttons?: TemplateButton[];
  status: string;
  created_at: string;
}

export function AutomationTemplates() {
  const [templates, setTemplates] = useState<Template[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showForm, setShowForm] = useState(false);
  const [editingTemplate, setEditingTemplate] = useState<Template | null>(null);
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  
  const [formData, setFormData] = useState({
    name: '',
    category: 'MARKETING',
    language: 'en_US',
    body: '',
    footer: '',
    header: { type: 'none', sample: '', location: { lat: '', lng: '', name: '', address: '' } } as TemplateHeader,
    buttons: [] as TemplateButton[],
    variableSamples: {} as Record<number, string>
  });

  const normalizeVariables = (text: string) => {
    const varRegex = /\{\{(\d+)?\}\}/g;
    let count = 0;
    const normalized = text.replace(varRegex, () => {
      count++;
      return `{{${count}}}`;
    });
    return { normalized, count };
  };

  const fetchTemplates = async () => {
    setIsLoading(true);
    try {
      const resp = await fetch('http://localhost:8080/api/automation/whatsapp/templates');
      const data = await resp.json();
      setTemplates(data || []);
    } catch (err) {
      console.error('Failed to fetch templates:', err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    const varRegex = /\{\{(\d+)\}\}/g;
    const matches = [...formData.body.matchAll(varRegex)];
    const currentVarIds = matches.map(m => parseInt(m[1]));
    
    const newSamples = { ...formData.variableSamples };
    let changed = false;

    // Remove samples for non-existent variables
    Object.keys(newSamples).forEach(id => {
      if (!currentVarIds.includes(parseInt(id))) {
        delete newSamples[parseInt(id)];
        changed = true;
      }
    });

    // Add missing samples
    currentVarIds.forEach(id => {
      if (!(id in newSamples)) {
        newSamples[id] = '';
        changed = true;
      }
    });

    if (changed) {
      setFormData(prev => ({ ...prev, variableSamples: newSamples }));
    }
  }, [formData.body]);

  useEffect(() => {
    fetchTemplates();
  }, []);

  const handleBodyChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    let value = e.target.value;
    
    // Auto-complete {{ -> {{n}}
    if (value.endsWith('{{') && value.length > (formData.body.length || 0)) {
      const { normalized } = normalizeVariables(value + '}');
      const newValue = normalized + '}';
      setFormData({ ...formData, body: newValue });
      
      // Focus cursor after }}
      setTimeout(() => {
        const textarea = e.target;
        textarea.setSelectionRange(newValue.length, newValue.length);
      }, 0);
      return;
    }

    const { normalized } = normalizeVariables(value);
    setFormData({ ...formData, body: normalized });
  };

  const applyFormatting = (prefix: string, suffix: string = prefix) => {
    const el = textareaRef.current;
    if (!el) return;

    const start = el.selectionStart;
    const end = el.selectionEnd;
    const text = el.value;
    const selectedText = text.substring(start, end);
    
    const before = text.substring(0, start);
    const after = text.substring(end);
    
    // If text is selected, wrap it. Else insert placeholder.
    const insertion = selectedText ? `${prefix}${selectedText}${suffix}` : `${prefix}${suffix}`;
    const newValue = before + insertion + after;
    
    const { normalized } = normalizeVariables(newValue);
    setFormData({ ...formData, body: normalized });

    // Refocus and set cursor
    setTimeout(() => {
      el.focus();
      const newCursorPos = selectedText ? start + insertion.length : start + prefix.length;
      el.setSelectionRange(newCursorPos, newCursorPos);
    }, 0);
  };

  const insertEmoji = (emoji: string) => {
    const el = textareaRef.current;
    if (!el) return;

    const start = el.selectionStart;
    const end = el.selectionEnd;
    const text = el.value;
    
    const before = text.substring(0, start);
    const after = text.substring(end);
    const newValue = before + emoji + after;
    
    setFormData({ ...formData, body: newValue });
    setShowEmojiPicker(false);

    setTimeout(() => {
      el.focus();
      el.setSelectionRange(start + emoji.length, start + emoji.length);
    }, 0);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);

    try {
      const payload = {
        id: editingTemplate?.id,
        name: formData.name,
        category: formData.category,
        language: formData.language,
        body: formData.body,
        header: formData.header,
        footer: formData.footer,
        buttons: formData.buttons,
        examples: Object.values(formData.variableSamples).join(', ')
      };

      const method = editingTemplate ? 'PUT' : 'POST';
      const resp = await fetch('http://localhost:8080/api/automation/whatsapp/templates', {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });

      if (resp.ok) {
        handleCancel();
        fetchTemplates();
      } else {
        const errorData = await resp.text();
        alert(`Failed to save template: ${errorData}`);
      }
    } catch (err) {
      console.error('Failed to save template:', err);
      alert('Operation failed. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this template? This will remove it from WhatsApp and the automation system.')) return;
    try {
      const resp = await fetch(`http://localhost:8080/api/automation/whatsapp/templates?id=${id}`, { method: 'DELETE' });
      if (resp.ok) {
        fetchTemplates();
      } else {
        const errorData = await resp.text();
        alert(`Deletion failed: ${errorData}`);
      }
    } catch (err) {
      console.error('Failed to delete template:', err);
      alert('Deletion failed. Please try again.');
    }
  };

  const handleEdit = (t: Template) => {
    setEditingTemplate(t);
    setFormData({
      name: t.template_name,
      category: t.category,
      language: t.language,
      body: t.body,
      footer: t.footer || '',
      header: t.header || { type: 'none', sample: '', location: { lat: '', lng: '', name: '', address: '' } },
      buttons: t.buttons || [],
      variableSamples: {}
    });
    setShowForm(true);
  };

  const handleCancel = () => {
    setShowForm(false);
    setEditingTemplate(null);
    setFormData({ 
      name: '', 
      category: 'MARKETING', 
      language: 'en_US', 
      body: '', 
      footer: '', 
      header: { type: 'none', sample: '', location: { lat: '', lng: '', name: '', address: '' } },
      buttons: [], 
      variableSamples: {} 
    });
  };

  const addButton = (type: TemplateButton['type']) => {
    if (formData.buttons.length >= 10) return;
    const newButton: TemplateButton = { type, text: '' };
    if (type === 'custom') newButton.payload = '';
    if (type === 'visit_website') newButton.url = '';
    if (type === 'call_phone') newButton.phoneNumber = '';
    if (type === 'flow') { newButton.flowID = ''; newButton.flowName = ''; }
    if (type === 'copy_code') newButton.offerCode = '';
    setFormData({ ...formData, buttons: [...formData.buttons, newButton] });
  };

  const updateButton = (index: number, updates: Partial<TemplateButton>) => {
    const newButtons = [...formData.buttons];
    newButtons[index] = { ...newButtons[index], ...updates };
    setFormData({ ...formData, buttons: newButtons });
  };

  const removeButton = (index: number) => {
    setFormData({ ...formData, buttons: formData.buttons.filter((_, i) => i !== index) });
  };

  const getStatusBadgeStyle = (status: string) => {
    const s = status.toLowerCase();
    if (s === 'approved') return { backgroundColor: '#dcfce7', color: '#166534' }; // Green
    if (s === 'rejected') return { backgroundColor: '#fee2e2', color: '#991b1b' }; // Red
    return { backgroundColor: '#fef9c3', color: '#854d0e' }; // Yellow for pending/other
  };

  return (
    <div className="automation-page">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <h2 style={{ fontSize: '1.25rem', fontWeight: 600 }}>WhatsApp Templates</h2>
        {!showForm && (
          <button className="btn-primary" onClick={() => setShowForm(true)}>
            Create New Template
          </button>
        )}
      </div>

      {showForm && (
        <div className="modal-overlay">
          <div className="modal-content" style={{ maxWidth: '1000px', display: 'grid', gridTemplateColumns: '1fr 340px', gap: '2rem' }}>
            <div>
              <div className="modal-header">
                <h3 className="modal-title">{editingTemplate ? 'Edit Template' : 'New Template'}</h3>
              </div>
              
              <form onSubmit={handleSubmit} style={{ maxHeight: '75vh', overflowY: 'auto', paddingRight: '1rem' }}>
                {/* Header Section */}
                <div className="form-group">
                  <label>Header Media (Optional)</label>
                  <select 
                    value={formData.header.type} 
                    onChange={e => setFormData({...formData, header: {...formData.header, type: e.target.value as TemplateHeader['type']}})}
                  >
                    <option value="none">None</option>
                    <option value="IMAGE">Image</option>
                    <option value="VIDEO">Video</option>
                    <option value="DOCUMENT">Document</option>
                    <option value="LOCATION">Location</option>
                  </select>
                  
                  {formData.header.type !== 'none' && formData.header.type !== 'LOCATION' && (
                    <div className="form-group" style={{ marginTop: '0.75rem' }}>
                      <label>Sample URL ({formData.header.type})</label>
                      <input 
                        type="text" 
                        placeholder={`https://example.com/file.${formData.header.type === 'IMAGE' ? 'jpg' : formData.header.type === 'VIDEO' ? 'mp4' : 'pdf'}`}
                        value={formData.header.sample}
                        onChange={e => setFormData({...formData, header: {...formData.header, sample: e.target.value}})}
                      />
                    </div>
                  )}

                  {formData.header.type === 'LOCATION' && (
                    <div className="form-row" style={{ marginTop: '0.75rem' }}>
                      <div className="form-group"><label>Lat</label><input type="text" value={formData.header.location?.lat} onChange={e => setFormData({...formData, header: {...formData.header, location: {...formData.header.location!, lat: e.target.value}}})} /></div>
                      <div className="form-group"><label>Lng</label><input type="text" value={formData.header.location?.lng} onChange={e => setFormData({...formData, header: {...formData.header, location: {...formData.header.location!, lng: e.target.value}}})} /></div>
                    </div>
                  )}
                </div>

                <div className="form-group">
                  <label>Template Name</label>
                  <input 
                    type="text" 
                    value={formData.name} 
                    onChange={e => setFormData({...formData, name: e.target.value.toLowerCase().replace(/\s/g, '_')})}
                    disabled={!!editingTemplate}
                    placeholder="e.g. promo_offer_01"
                    required
                  />
                </div>

                <div className="form-row">
                  <div className="form-group">
                    <label>Category</label>
                    <select value={formData.category} onChange={e => setFormData({...formData, category: e.target.value})} disabled={!!editingTemplate}>
                      <option value="MARKETING">Marketing</option>
                      <option value="UTILITY">Utility</option>
                    </select>
                  </div>
                  <div className="form-group">
                    <label>Language</label>
                    <select value={formData.language} onChange={e => setFormData({...formData, language: e.target.value})} disabled={!!editingTemplate}>
                      <option value="en_US">English (US)</option>
                      <option value="hi_IN">Hindi</option>
                    </select>
                  </div>
                </div>

                <div className="form-group">
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
                    <label style={{ marginBottom: 0 }}>Template Body</label>
                    <div style={{ position: 'relative' }}>
                      <button 
                        type="button" 
                        title="Emoji Rules"
                        style={{ background: 'none', border: 'none', cursor: 'help', color: '#64748b', display: 'flex', alignItems: 'center' }}
                        onClick={(e) => {
                          e.preventDefault();
                          alert("WhatsApp Guidelines:\n- Variables {{n}} must have surrounding text.\n- Variables cannot be at the very start or end.\n- Max 10 variables allowed.");
                        }}
                      >
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>
                      </button>
                    </div>
                  </div>

                  {/* Rich Toolbar */}
                  <div style={{ 
                    display: 'flex', 
                    alignItems: 'center', 
                    gap: '4px', 
                    padding: '6px', 
                    backgroundColor: '#f8fafc', 
                    border: '1px solid #e2e8f0', 
                    borderBottom: 'none',
                    borderTopLeftRadius: '8px', 
                    borderTopRightRadius: '8px' 
                  }}>
                    <button type="button" className="toolbar-btn" title="Emoji" onClick={() => setShowEmojiPicker(!showEmojiPicker)}>🙂</button>
                    <div style={{ width: '1px', height: '16px', backgroundColor: '#e2e8f0', margin: '0 4px' }}></div>
                    <button type="button" className="toolbar-btn" title="Bold" style={{ fontWeight: 700 }} onClick={() => applyFormatting('*')}>B</button>
                    <button type="button" className="toolbar-btn" title="Italic" style={{ fontStyle: 'italic' }} onClick={() => applyFormatting('_')}>I</button>
                    <button type="button" className="toolbar-btn" title="Strikethrough" style={{ textDecoration: 'line-through' }} onClick={() => applyFormatting('~')}>S</button>
                    <button type="button" className="toolbar-btn" title="Code" onClick={() => applyFormatting('```')}>{'</>'}</button>
                    <div style={{ width: '1px', height: '16px', backgroundColor: '#e2e8f0', margin: '0 4px' }}></div>
                    <button 
                      type="button" 
                      className="btn-secondary" 
                      style={{ padding: '0.2rem 0.6rem', fontSize: '0.75rem', marginLeft: 'auto' }}
                      onClick={() => {
                        const el = textareaRef.current;
                        const { normalized } = normalizeVariables(formData.body + '{{' + '}}');
                        setFormData({ ...formData, body: normalized });
                        if (el) setTimeout(() => { el.focus(); el.setSelectionRange(normalized.length, normalized.length); }, 0);
                      }}
                    >
                      + Add variable
                    </button>

                    {showEmojiPicker && (
                      <div style={{ 
                        position: 'absolute', 
                        top: '40px', 
                        left: '0', 
                        zIndex: 10, 
                        backgroundColor: 'white', 
                        boxShadow: '0 10px 15px -3px rgba(0,0,0,0.1)', 
                        border: '1px solid #e2e8f0', 
                        borderRadius: '8px', 
                        padding: '8px',
                        display: 'grid',
                        gridTemplateColumns: 'repeat(6, 1fr)',
                        gap: '4px'
                      }}>
                        {['😊','🚀','📦','✅','⚠️','🔥','💡','📱','🛒','💰','🎁','✨'].map(emoji => (
                          <button 
                            key={emoji} 
                            type="button" 
                            style={{ fontSize: '1.25rem', padding: '4px', background: 'none', border: 'none', cursor: 'pointer', borderRadius: '4px' }}
                            onMouseEnter={e => (e.currentTarget.style.backgroundColor = '#f1f5f9')}
                            onMouseLeave={e => (e.currentTarget.style.backgroundColor = 'transparent')}
                            onClick={() => insertEmoji(emoji)}
                          >
                            {emoji}
                          </button>
                        ))}
                      </div>
                    )}
                  </div>

                  <textarea 
                    ref={textareaRef}
                    value={formData.body} 
                    onChange={handleBodyChange}
                    placeholder="Enter template message... Use {{ }} for variables."
                    rows={4}
                    required
                    style={{ 
                      borderTopLeftRadius: 0, 
                      borderTopRightRadius: 0,
                      marginTop: '-1px'
                    }}
                  />
                </div>
                
                {/* Variable Samples Section */}
                {Object.keys(formData.variableSamples).length > 0 && (
                  <div style={{ marginBottom: '1.5rem', padding: '1rem', backgroundColor: '#f0f9ff', borderRadius: '8px', border: '1px solid #bae6fd' }}>
                    <label style={{ color: '#0369a1', fontWeight: 600, display: 'block', marginBottom: '0.75rem' }}>Variable Samples (Required)</label>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                      {Object.keys(formData.variableSamples).map(idStr => {
                        const id = parseInt(idStr);
                        return (
                          <div key={id} style={{ display: 'grid', gridTemplateColumns: '60px 1fr', alignItems: 'center', gap: '1rem' }}>
                            <span style={{ fontSize: '0.85rem', fontWeight: 700, color: '#0c4a6e' }}>{`{{${id}}}`}</span>
                            <input 
                              type="text" 
                              placeholder={`Sample value for {{${id}}}`}
                              value={formData.variableSamples[id]}
                              onChange={e => setFormData({
                                ...formData, 
                                variableSamples: { ...formData.variableSamples, [id]: e.target.value }
                              })}
                              required
                            />
                          </div>
                        );
                      })}
                    </div>
                  </div>
                )}
                
                <div className="form-group">
                  <label>Footer Text (Optional)</label>
                  <input 
                    type="text" 
                    value={formData.footer} 
                    onChange={e => setFormData({...formData, footer: e.target.value})}
                    placeholder="e.g. Best grocery deals on WhatsApp!"
                  />
                </div>

                {/* Buttons Section */}
                <div style={{ marginTop: '1.5rem', borderTop: '1px solid #e2e8f0', paddingTop: '1.5rem' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
                    <label style={{ fontSize: '1rem', fontWeight: 600 }}>Buttons (Optional)</label>
                    {formData.buttons.length < 10 && (
                      <div className="dropdown" style={{ position: 'relative' }}>
                        <select 
                          className="btn-secondary" 
                          style={{ padding: '0.4rem 2rem 0.4rem 1rem', width: 'auto' }}
                          onChange={e => {
                            if (e.target.value) {
                              addButton(e.target.value as TemplateButton['type']);
                              e.target.value = '';
                            }
                          }}
                        >
                          <option value="">+ Add Button</option>
                          <option value="custom">Custom (Quick Reply)</option>
                          <option value="visit_website">Visit Website</option>
                          <option value="call_phone">Call Phone</option>
                          <option value="flow">Flow</option>
                          <option value="copy_code">Copy Offer Code</option>
                        </select>
                      </div>
                    )}
                  </div>

                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                    {formData.buttons.map((btn, idx) => (
                      <div key={idx} style={{ padding: '1rem', backgroundColor: '#f8fafc', borderRadius: '8px', border: '1px solid #e2e8f0' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.75rem' }}>
                          <span style={{ fontSize: '0.75rem', fontWeight: 700, color: '#64748b', textTransform: 'uppercase' }}>{btn.type.replace('_', ' ')}</span>
                          <button type="button" onClick={() => removeButton(idx)} style={{ color: '#ef4444', background: 'none', border: 'none', cursor: 'pointer', fontSize: '1.2rem' }}>×</button>
                        </div>
                        <div className="form-row">
                          <div className="form-group"><label>Button Text</label><input type="text" value={btn.text} onChange={e => updateButton(idx, { text: e.target.value })} required /></div>
                          {btn.type === 'custom' && <div className="form-group"><label>Payload</label><input type="text" value={btn.payload} onChange={e => updateButton(idx, { payload: e.target.value })} /></div>}
                          {btn.type === 'visit_website' && <div className="form-group"><label>Website URL</label><input type="text" value={btn.url} onChange={e => updateButton(idx, { url: e.target.value })} required /></div>}
                          {btn.type === 'call_phone' && <div className="form-group"><label>Phone Number</label><input type="text" value={btn.phoneNumber} onChange={e => updateButton(idx, { phoneNumber: e.target.value })} required /></div>}
                          {btn.type === 'flow' && <div className="form-group"><label>Flow ID</label><input type="text" value={btn.flowID} onChange={e => updateButton(idx, { flowID: e.target.value })} required /></div>}
                          {btn.type === 'copy_code' && <div className="form-group"><label>Offer Code</label><input type="text" value={btn.offerCode} onChange={e => updateButton(idx, { offerCode: e.target.value })} required /></div>}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                <div className="modal-footer" style={{ borderTop: 'none', padding: '1.5rem 0 0' }}>
                  <button type="button" className="btn-secondary" onClick={handleCancel}>Cancel</button>
                  <button type="submit" className="btn-primary" disabled={isSubmitting}>
                    {isSubmitting ? 'Saving...' : editingTemplate ? 'Update Template' : 'Submit to Meta'}
                  </button>
                </div>
              </form>
            </div>

            {/* Preview Section */}
            <div style={{ borderLeft: '1px solid #e2e8f0', paddingLeft: '2rem' }}>
              <h4 style={{ fontSize: '0.9rem', fontWeight: 600, color: '#64748b', marginBottom: '1rem', textTransform: 'uppercase' }}>Template Preview</h4>
              <div style={{ backgroundColor: '#e5ddd5', borderRadius: '12px', padding: '1rem', backgroundImage: 'url("https://user-images.githubusercontent.com/15075759/28719144-86dc0f70-73b1-11e7-911d-60d70fcded21.png")', backgroundSize: 'cover' }}>
                <div style={{ backgroundColor: 'white', borderRadius: '10px', overflow: 'hidden', boxShadow: '0 1px 2px rgba(0,0,0,0.1)', maxWidth: '280px', position: 'relative' }}>
                  {formData.header.type !== 'none' && (
                    <div style={{ backgroundColor: '#f0f2f5', height: '140px', display: 'flex', alignItems: 'center', justifyContent: 'center', overflow: 'hidden' }}>
                      {formData.header.type === 'IMAGE' && formData.header.sample ? <img src={formData.header.sample} alt="preview" style={{ width: '100%', height: '100%', objectFit: 'cover' }} /> : 
                       formData.header.type === 'VIDEO' ? <div style={{ textAlign: 'center' }}><svg width="32" height="32" viewBox="0 0 24 24" fill="#64748b"><path d="M10 8l6 4-6 4V8z"/></svg></div> :
                       formData.header.type === 'DOCUMENT' ? <div style={{ textAlign: 'center' }}><svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#64748b" strokeWidth="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg></div> :
                       formData.header.type === 'LOCATION' ? <div style={{ fontSize: '0.6rem', textAlign: 'center' }}>Map View PIN</div> :
                       <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#94a3b8" strokeWidth="2"><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h7"/><line x1="16" y1="5" x2="22" y2="5"/><line x1="19" y1="2" x2="19" y2="8"/></svg>}
                    </div>
                  )}
                  <div style={{ padding: '8px 12px' }}>
                    <div style={{ whiteSpace: 'pre-wrap', fontSize: '0.85rem', color: '#111b21', lineHeight: '1.4' }}>
                      {formData.body.replace(/\{\{(\d+)\}\}/g, (match, id) => {
                        return formData.variableSamples[parseInt(id)] || match;
                      }) || 'Hello details...'}
                    </div>
                    <div style={{ fontSize: '0.75rem', color: '#667781', marginTop: '4px' }}>{formData.footer}</div>
                    <div style={{ textAlign: 'right', fontSize: '0.65rem', color: '#667781', marginTop: '4px' }}>{new Date().getHours()}:{new Date().getMinutes()}</div>
                  </div>
                  {formData.buttons.length > 0 && (
                    <div style={{ borderTop: '1px solid #f0f2f5' }}>
                      {formData.buttons.map((b, idx) => (
                        <div key={idx} style={{ padding: '10px', textAlign: 'center', color: '#00a884', fontSize: '0.85rem', fontWeight: 500, borderBottom: idx < formData.buttons.length - 1 ? '1px solid #f0f2f5' : 'none', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px' }}>
                          {b.type === 'visit_website' && <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>}
                          {b.type === 'call_phone' && <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"/></svg>}
                          {b.text || 'Button'}
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>
              <div className="alert-info" style={{ marginTop: '1.5rem', fontSize: '0.8rem' }}>
                <p><strong>Note:</strong> Buttons may appear as a list if more than 3 are added.</p>
              </div>
            </div>
          </div>
        </div>
      )}

      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>Template Name</th>
              <th>Category</th>
              <th>Language</th>
              <th>Status</th>
              <th>Created Time</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr><td colSpan={6} style={{ textAlign: 'center', padding: '2rem' }}>Loading templates...</td></tr>
            ) : templates.length === 0 ? (
              <tr><td colSpan={6} style={{ textAlign: 'center', padding: '2rem' }}>No templates found.</td></tr>
            ) : (
              templates.map(t => (
                <tr key={t.id}>
                  <td><strong>{t.template_name}</strong></td>
                  <td><span className="badge" style={{ backgroundColor: '#f1f5f9', color: '#64748b' }}>{t.category}</span></td>
                  <td>{t.language}</td>
                  <td>
                    <span className="badge" style={getStatusBadgeStyle(t.status || 'PENDING')}>
                      {(t.status || 'PENDING').toUpperCase()}
                    </span>
                  </td>
                  <td style={{ fontSize: '0.85rem' }}>{new Date(t.created_at).toLocaleDateString()}</td>
                  <td>
                    <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
                      <button 
                        onClick={() => handleEdit(t)} 
                        title="Edit Template"
                        style={{ background: 'none', border: 'none', cursor: 'pointer', padding: '4px', display: 'flex', alignItems: 'center', color: 'var(--text-secondary)' }}
                      >
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                          <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
                          <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
                        </svg>
                      </button>
                      <button 
                        onClick={() => handleDelete(t.id)} 
                        title="Delete Template"
                        style={{ background: 'none', border: 'none', cursor: 'pointer', padding: '4px', display: 'flex', alignItems: 'center', color: '#ef4444' }}
                      >
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                          <polyline points="3 6 5 6 21 6"></polyline>
                          <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                          <line x1="10" y1="11" x2="10" y2="17"></line>
                          <line x1="14" y1="11" x2="14" y2="17"></line>
                        </svg>
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
