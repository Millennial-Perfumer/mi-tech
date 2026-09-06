import { FormEvent, useCallback, useEffect, useMemo, useState } from 'react'
import { CircleAlert, Edit3, MessageSquare, Plus, RefreshCw, Trash2, X } from 'lucide-react'
import { apiJson, apiRequest, arrayFrom, formatDate, numberValue, textValue } from '../../lib/http'
import { usePeriodFilter } from '../../lib/usePeriodFilter'

type Props = { token: string; onUnauthorized: () => void }
type Tab = 'activity' | 'templates' | 'triggers' | 'events'
type Modal = 'template' | 'trigger' | 'event'
type Row = Record<string, unknown>
type AutomationForm = { template_name: string; language: string; category: string; body: string; footer: string; header_type: string; variable_mappings: string; webhook_topic: string; template_id: string; name: string; topic: string; description: string }

const emptyForm: AutomationForm = { template_name: '', language: 'en', category: 'MARKETING', body: '', footer: '', header_type: 'none', variable_mappings: '{}', webhook_topic: '', template_id: '', name: '', topic: '', description: '' }

function jsonText(value: unknown) {
  if (typeof value === 'string') return value
  try { return JSON.stringify(value || {}, null, 2) } catch { return '{}' }
}

export function AutomationPage({ token, onUnauthorized }: Props) {
  const { startDate, endDate } = usePeriodFilter()
  const [tab, setTab] = useState<Tab>('activity')
  const [metrics, setMetrics] = useState<Row>({})
  const [rows, setRows] = useState<Row[]>([])
  const [search, setSearch] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [isWorking, setIsWorking] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [modal, setModal] = useState<Modal | null>(null)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [form, setForm] = useState<AutomationForm>(emptyForm)

  const requestPath = tab === 'activity' ? `/api/automation/whatsapp/messages?start_date=${startDate}&end_date=${endDate}` : `/api/automation/whatsapp/${tab}`
  const load = useCallback(async () => {
    setIsLoading(true)
    setError('')
    try {
      const [metricData, listData] = await Promise.all([
        apiJson<Row>(token, onUnauthorized, `/api/automation/whatsapp/metrics?start_date=${startDate}&end_date=${endDate}`),
        apiJson<unknown>(token, onUnauthorized, requestPath),
      ])
      setMetrics(metricData)
      setRows(arrayFrom(listData, tab === 'activity' ? 'messages' : tab))
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to load automation')
    } finally {
      setIsLoading(false)
    }
  }, [endDate, onUnauthorized, requestPath, startDate, tab, token])

  useEffect(() => { void load() }, [load])

  const visibleRows = useMemo(() => {
    const query = search.trim().toLowerCase()
    return rows.filter((row) => !query || Object.values(row).join(' ').toLowerCase().includes(query))
  }, [rows, search])

  const sync = async (path: string, label: string) => {
    setIsWorking(true)
    setError('')
    try { await apiRequest(token, onUnauthorized, path, { method: 'POST' }); setNotice(label); await load() } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Automation action failed') } finally { setIsWorking(false) }
  }

  const openCreate = (nextModal: Modal) => { setEditingId(null); setForm(emptyForm); setModal(nextModal) }
  const openTemplateEdit = (row: Row) => { setEditingId(textValue(row.id, '')); setForm({ ...emptyForm, template_name: textValue(row.template_name || row.name, ''), language: textValue(row.language, 'en'), category: textValue(row.category, 'MARKETING'), body: textValue(row.body, ''), footer: textValue(row.footer, ''), variable_mappings: jsonText(row.variable_mappings) }); setModal('template') }

  const toggleTrigger = async (row: Row) => {
    setIsWorking(true)
    try { await apiRequest(token, onUnauthorized, '/api/automation/whatsapp/triggers', { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ id: numberValue(row.id), enabled: !row.enabled }) }); await load() } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to update trigger') } finally { setIsWorking(false) }
  }

  const deleteRow = async (row: Row) => {
    const id = textValue(row.id, '')
    if (!id || !window.confirm(`Delete this ${tab === 'templates' ? 'template' : tab === 'triggers' ? 'trigger' : 'event'}?`)) return
    setIsWorking(true)
    try { await apiRequest(token, onUnauthorized, `/api/automation/whatsapp/${tab}?id=${id}`, { method: 'DELETE' }); setNotice('Automation record deleted'); await load() } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to delete automation record') } finally { setIsWorking(false) }
  }

  const submitModal = async (event: FormEvent) => {
    event.preventDefault()
    if (!modal) return
    setIsWorking(true)
    setError('')
    try {
      let path = ''
      let method = 'POST'
      let body: Record<string, unknown>
      if (modal === 'template') {
        path = '/api/automation/whatsapp/templates'
        if (editingId) {
          method = 'PUT'
          let mappings: unknown = {}
          try { mappings = JSON.parse(form.variable_mappings || '{}') } catch { throw new Error('Variable mappings must be valid JSON') }
          body = { id: Number(editingId), variable_mappings: mappings }
        } else {
          body = { name: form.template_name, language: form.language, category: form.category, body: form.body, footer: form.footer || undefined, header: form.header_type === 'none' ? undefined : { type: form.header_type } }
        }
      } else if (modal === 'trigger') {
        path = '/api/automation/whatsapp/triggers'
        body = { webhook_topic: form.webhook_topic, template_id: Number(form.template_id) }
      } else {
        path = '/api/automation/whatsapp/events'
        body = { name: form.name, topic: form.topic, description: form.description }
      }
      await apiRequest(token, onUnauthorized, path, { method, headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
      setNotice(`${modal === 'template' ? 'Template' : modal === 'trigger' ? 'Trigger' : 'Event'} ${editingId ? 'updated' : 'added'}`)
      setModal(null)
      setEditingId(null)
      setForm(emptyForm)
      await load()
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to save automation rule')
    } finally {
      setIsWorking(false)
    }
  }

  const modalTitle = editingId ? 'Edit template mapping' : modal === 'template' ? 'Add WhatsApp template' : modal === 'trigger' ? 'Add trigger' : 'Add event'

  return <section className="workspace-page automation-page" aria-labelledby="automation-heading"><header className="workspace-page-header"><div><p className="eyebrow">Engagement / Automation</p><h2 id="automation-heading">Keep useful messages moving.</h2><p>Monitor WhatsApp delivery, keep templates in sync, and make triggers understandable to the team.</p></div><div className="support-header-actions"><button className="secondary-button" type="button" onClick={() => void load()} disabled={isLoading}><RefreshCw size={15} className={isLoading ? 'spin' : undefined} aria-hidden="true" /> Refresh</button>{tab === 'templates' && <><button className="secondary-button" type="button" onClick={() => void sync('/api/automation/whatsapp/templates/sync', 'Template statuses synced')} disabled={isWorking}><RefreshCw size={15} aria-hidden="true" /> Sync status</button><button className="primary-button" type="button" onClick={() => openCreate('template')}><Plus size={15} aria-hidden="true" /> Add template</button></>}{tab === 'triggers' && <button className="primary-button" type="button" onClick={() => openCreate('trigger')}><Plus size={15} aria-hidden="true" /> Add trigger</button>}{tab === 'events' && <button className="primary-button" type="button" onClick={() => openCreate('event')}><Plus size={15} aria-hidden="true" /> Add event</button>}</div></header>{error && <div className="dashboard-error" role="alert"><CircleAlert size={18} aria-hidden="true" /><span>{error}</span><button type="button" onClick={() => void load()}>Try again</button></div>}{notice && <div className="inventory-notice" role="status">{notice}</div>}<div className="dashboard-metrics-grid automation-metrics">{[['sent', 'Sent'], ['delivered', 'Delivered'], ['read', 'Read'], ['read_rate', 'Read rate'], ['triggered', 'Triggered'], ['failed', 'Failed']].map(([key, label]) => <article className="dashboard-metric-card" key={key}><div className="dashboard-metric-heading"><span className="metric-label">{label}</span><span className="dashboard-metric-icon"><MessageSquare size={14} aria-hidden="true" /></span></div><strong>{key === 'read_rate' ? `${numberValue(metrics[key]).toFixed(1)}%` : numberValue(metrics[key]).toLocaleString('en-IN')}</strong><small className="metric-detail">Selected period</small></article>)}</div><section className="reports-table-card"><div className="b2b-tabs" role="tablist" aria-label="Automation sections">{(['activity', 'templates', 'triggers', 'events'] as Tab[]).map((item) => <button key={item} className={`filter-chip ${tab === item ? 'filter-chip-active' : ''}`} type="button" role="tab" aria-selected={tab === item} onClick={() => { setTab(item); setSearch('') }}>{item === 'activity' ? 'Activity' : item[0].toUpperCase() + item.slice(1)}</button>)}</div><div className="b2b-toolbar"><label className="orders-search"><MessageSquare size={16} aria-hidden="true" /><span className="sr-only">Search automation</span><input value={search} onChange={(event) => setSearch(event.target.value)} placeholder={`Search ${tab}`} /></label></div><div className="orders-table-wrap"><table className="orders-table"><caption className="sr-only">WhatsApp automation {tab}</caption><thead>{tab === 'activity' ? <tr><th>Sent</th><th>Customer</th><th>Template</th><th>Recipient</th><th>Status</th></tr> : tab === 'templates' ? <tr><th>Template</th><th>Category</th><th>Language</th><th>Status</th><th>Actions</th></tr> : tab === 'triggers' ? <tr><th>Topic</th><th>Template</th><th>Enabled</th><th>Created</th><th>Actions</th></tr> : <tr><th>Name</th><th>Topic</th><th>Description</th><th>Created</th><th>Actions</th></tr>}</thead><tbody>{isLoading ? <tr><td colSpan={5} className="table-state">Loading automation…</td></tr> : visibleRows.length === 0 ? <tr><td colSpan={5} className="table-state">No {tab} found.</td></tr> : visibleRows.map((row, index) => { const id = textValue(row.id, String(index)); if (tab === 'activity') return <tr key={id}><td>{formatDate(row.sent_at || row.created_at)}</td><td><strong>{textValue(row.customer_name, '—')}</strong><small className="table-subtext">{textValue(row.order_number || row.order_id, '')}</small></td><td className="mono-text">{textValue(row.template_name)}</td><td>{textValue(row.phone_number)}</td><td><span className={`status-pill status-pill-${textValue(row.status, 'queued') === 'failed' ? 'danger' : 'success'}`}>{textValue(row.status, 'queued')}</span></td></tr>; if (tab === 'templates') return <tr key={id}><td><strong>{textValue(row.template_name || row.name, 'Unnamed template')}</strong><small className="table-subtext">{textValue(row.body, 'No body configured').slice(0, 110)}</small></td><td>{textValue(row.category)}</td><td>{textValue(row.language)}</td><td><span className="status-pill status-pill-neutral">{textValue(row.status, 'unknown')}</span></td><td><div className="table-action-group"><button className="table-link-button" type="button" onClick={() => openTemplateEdit(row)}><Edit3 size={13} aria-hidden="true" /> Edit</button><button className="table-link-button danger-link" type="button" disabled={isWorking} onClick={() => void deleteRow(row)}><Trash2 size={13} aria-hidden="true" /> Delete</button></div></td></tr>; if (tab === 'triggers') return <tr key={id}><td className="mono-text">{textValue(row.webhook_topic || row.topic)}</td><td>{textValue(row.template_name || row.template_id)}</td><td><label className="toggle-control"><input type="checkbox" checked={Boolean(row.enabled)} onChange={() => void toggleTrigger(row)} /><span>{row.enabled ? 'On' : 'Off'}</span></label></td><td>{formatDate(row.created_at)}</td><td><button className="table-link-button danger-link" type="button" disabled={isWorking} onClick={() => void deleteRow(row)}><Trash2 size={13} aria-hidden="true" /> Delete</button></td></tr>; return <tr key={id}><td><strong>{textValue(row.name, 'Unnamed event')}</strong></td><td className="mono-text">{textValue(row.topic)}</td><td>{textValue(row.description)}</td><td>{formatDate(row.created_at)}</td><td><button className="table-link-button danger-link" type="button" disabled={isWorking} onClick={() => void deleteRow(row)}><Trash2 size={13} aria-hidden="true" /> Delete</button></td></tr> })}</tbody></table></div></section>{modal && <div className="modal-scrim" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) setModal(null) }}><form className="modal-card automation-modal" role="dialog" aria-modal="true" aria-labelledby="automation-modal-heading" onSubmit={submitModal}><div className="modal-heading"><div><p className="eyebrow">Automation setup</p><h2 id="automation-modal-heading">{modalTitle}</h2></div><button className="icon-button" type="button" aria-label="Close" onClick={() => setModal(null)}><X size={19} aria-hidden="true" /></button></div>{modal === 'template' ? editingId ? <><p className="form-help">This keeps the template content intact and updates only its variable mapping.</p><label className="form-field"><span>Variable mappings JSON</span><textarea required rows={8} value={form.variable_mappings} onChange={(event) => setForm({ ...form, variable_mappings: event.target.value })} /></label></> : <><div className="form-grid-two"><label className="form-field"><span>Template name</span><input required value={form.template_name} onChange={(event) => setForm({ ...form, template_name: event.target.value })} /></label><label className="form-field"><span>Language</span><input required value={form.language} onChange={(event) => setForm({ ...form, language: event.target.value })} placeholder="en" /></label><label className="form-field"><span>Category</span><select value={form.category} onChange={(event) => setForm({ ...form, category: event.target.value })}><option>MARKETING</option><option>UTILITY</option><option>AUTHENTICATION</option></select></label><label className="form-field"><span>Header</span><select value={form.header_type} onChange={(event) => setForm({ ...form, header_type: event.target.value })}><option value="none">No header</option><option value="text">Text</option><option value="image">Image</option><option value="video">Video</option></select></label></div><label className="form-field"><span>Body</span><textarea required rows={7} value={form.body} onChange={(event) => setForm({ ...form, body: event.target.value })} placeholder="Hello {{1}}, your order is ready." /></label><label className="form-field"><span>Footer <small>(optional)</small></span><input value={form.footer} onChange={(event) => setForm({ ...form, footer: event.target.value })} /></label></> : modal === 'trigger' ? <div className="form-grid-two"><label className="form-field"><span>Webhook topic</span><input required value={form.webhook_topic} onChange={(event) => setForm({ ...form, webhook_topic: event.target.value })} placeholder="orders/create" /></label><label className="form-field"><span>Template ID</span><input required type="number" min="1" value={form.template_id} onChange={(event) => setForm({ ...form, template_id: event.target.value })} /></label></div> : <><label className="form-field"><span>Event name</span><input required value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} /></label><label className="form-field"><span>Topic</span><input required value={form.topic} onChange={(event) => setForm({ ...form, topic: event.target.value })} placeholder="order.fulfilled" /></label><label className="form-field"><span>Description</span><textarea rows={3} value={form.description} onChange={(event) => setForm({ ...form, description: event.target.value })} /></label></>}<div className="modal-actions"><button className="secondary-button" type="button" onClick={() => setModal(null)}>Cancel</button><button className="primary-button" type="submit" disabled={isWorking}><Plus size={14} aria-hidden="true" /> {isWorking ? 'Saving…' : editingId ? 'Save mapping' : 'Save'}</button></div></form></div>}</section>
}
