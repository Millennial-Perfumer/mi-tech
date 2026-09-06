import { FormEvent, useCallback, useEffect, useMemo, useState } from 'react'
import { CircleAlert, Clock3, Plus, RefreshCw, Search, Ticket as TicketIcon, X } from 'lucide-react'
import { API_BASE } from '../../lib/api'

type SupportPageProps = { token: string; onUnauthorized: () => void }

type Ticket = {
  id: number
  ticket_id: string
  title: string
  description: string
  status: 'open' | 'in-progress' | 'resolved' | 'closed'
  priority: 'low' | 'medium' | 'high' | 'urgent'
  created_at: string
  updated_at: string
}

const tabs: { id: Ticket['status']; label: string }[] = [
  { id: 'open', label: 'Open' },
  { id: 'in-progress', label: 'In progress' },
  { id: 'resolved', label: 'Resolved' },
  { id: 'closed', label: 'Closed' },
]

function formatDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return date.toLocaleDateString('en-IN', { day: 'numeric', month: 'short', year: 'numeric' })
}

export function SupportPage({ token, onUnauthorized }: SupportPageProps) {
  const [tickets, setTickets] = useState<Ticket[]>([])
  const [activeTab, setActiveTab] = useState<Ticket['status']>('open')
  const [search, setSearch] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [isCreateOpen, setIsCreateOpen] = useState(false)
  const [newTicket, setNewTicket] = useState({ title: '', description: '', priority: 'medium' as Ticket['priority'] })

  const request = useCallback(async (path: string, options: RequestInit = {}) => {
    const response = await fetch(`${API_BASE}${path}`, { ...options, headers: { Authorization: `Bearer ${token}`, ...(options.headers || {}) } })
    if (response.status === 401) { onUnauthorized(); throw new Error('Your session has expired. Please sign in again.') }
    if (!response.ok) throw new Error(`Support request failed with status ${response.status}`)
    return response
  }, [onUnauthorized, token])

  const loadTickets = useCallback(async () => {
    setIsLoading(true)
    try {
      const response = await request('/api/support/tickets')
      const data = await response.json() as { success?: boolean; tickets?: Ticket[]; message?: string }
      if (!data.success) throw new Error(data.message || 'Tickets were not returned')
      setTickets(data.tickets || [])
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to load support tickets')
    } finally {
      setIsLoading(false)
    }
  }, [request])

  useEffect(() => { void loadTickets() }, [loadTickets])

  const filteredTickets = useMemo(() => {
    const query = search.trim().toLowerCase()
    return tickets.filter((ticket) => ticket.status === activeTab && (!query || `${ticket.ticket_id} ${ticket.title} ${ticket.description}`.toLowerCase().includes(query))).sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
  }, [activeTab, search, tickets])

  const updateStatus = async (ticket: Ticket, status: Ticket['status']) => {
    setIsSaving(true)
    try {
      await request(`/api/support/tickets/${ticket.id}/status`, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ status }) })
      setNotice(`${ticket.ticket_id} moved to ${status.replace('-', ' ')}`)
      await loadTickets()
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to update ticket')
    } finally { setIsSaving(false) }
  }

  const createTicket = async (event: FormEvent) => {
    event.preventDefault()
    if (!newTicket.title.trim() || isSaving) return
    setIsSaving(true)
    try {
      await request('/api/support/tickets', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(newTicket) })
      setIsCreateOpen(false)
      setNewTicket({ title: '', description: '', priority: 'medium' })
      setActiveTab('open')
      setNotice('Ticket raised successfully')
      await loadTickets()
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to create ticket')
    } finally { setIsSaving(false) }
  }

  return (
    <section className="workspace-page support-page" aria-labelledby="support-heading">
      <header className="workspace-page-header"><div><p className="eyebrow">Engagement / Support</p><h2 id="support-heading">Keep every issue moving.</h2><p>Prioritise customer concerns, give the team a clear next action, and close the loop without losing the original context.</p></div><div className="support-header-actions"><button className="secondary-button" type="button" onClick={() => void loadTickets()} disabled={isLoading}><RefreshCw size={15} className={isLoading ? 'spin' : undefined} aria-hidden="true" /> Refresh</button><button className="primary-button" type="button" onClick={() => { setError(''); setIsCreateOpen(true) }}><Plus size={16} aria-hidden="true" /> New ticket</button></div></header>
      {error && <div className="dashboard-error" role="alert"><CircleAlert size={18} aria-hidden="true" /><span>{error}</span><button type="button" onClick={() => { setError(''); void loadTickets() }}>Try again</button></div>}
      {notice && <div className="inventory-notice" role="status">{notice}</div>}
      <div className="support-toolbar"><div className="support-tabs" role="tablist" aria-label="Ticket status"><span className="sr-only">Ticket status</span>{tabs.map((tab) => <button type="button" role="tab" aria-selected={activeTab === tab.id} className={`support-tab ${activeTab === tab.id ? 'support-tab-active' : ''}`} key={tab.id} onClick={() => setActiveTab(tab.id)}>{tab.label}<span>{tickets.filter((ticket) => ticket.status === tab.id).length}</span></button>)}</div><label className="orders-search"><Search size={16} aria-hidden="true" /><span className="sr-only">Search tickets</span><input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search ticket ID or issue" /></label></div>
      <div className="support-list" aria-live="polite">{isLoading ? <div className="empty-panel"><TicketIcon size={20} aria-hidden="true" /><p>Loading tickets…</p></div> : filteredTickets.length === 0 ? <div className="empty-panel"><TicketIcon size={20} aria-hidden="true" /><div><h2>No {activeTab.replace('-', ' ')} tickets</h2><p>There is nothing matching this view right now.</p></div></div> : filteredTickets.map((ticket) => <article className="support-ticket-card" key={ticket.id}><div className="support-ticket-main"><div className="support-ticket-meta"><span className="support-ticket-id">{ticket.ticket_id}</span><span className={`priority-pill priority-${ticket.priority}`}>{ticket.priority}</span></div><h3>{ticket.title}</h3><p>{ticket.description || 'No description provided.'}</p><div className="support-ticket-foot"><span><Clock3 size={14} aria-hidden="true" /> Created {formatDate(ticket.created_at)}</span><span>Updated {formatDate(ticket.updated_at)}</span></div></div><div className="support-ticket-controls"><label className="compact-select"><span className="sr-only">Status for {ticket.ticket_id}</span><select value={ticket.status} disabled={isSaving} onChange={(event) => void updateStatus(ticket, event.target.value as Ticket['status'])}>{tabs.map((tab) => <option key={tab.id} value={tab.id}>{tab.label}</option>)}</select></label><span className="support-ticket-source">WhatsApp</span></div></article>)}</div>
      {isCreateOpen && <div className="modal-scrim" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) setIsCreateOpen(false) }}><form className="modal-card" role="dialog" aria-modal="true" aria-labelledby="new-ticket-heading" onSubmit={createTicket}><div className="modal-heading"><div><p className="eyebrow">Support</p><h2 id="new-ticket-heading">Raise a ticket</h2></div><button className="icon-button" type="button" aria-label="Close new ticket" onClick={() => setIsCreateOpen(false)}><X size={19} aria-hidden="true" /></button></div><label className="form-field"><span>Title</span><input value={newTicket.title} onChange={(event) => setNewTicket((current) => ({ ...current, title: event.target.value }))} placeholder="What needs attention?" required /></label><label className="form-field"><span>Description</span><textarea value={newTicket.description} onChange={(event) => setNewTicket((current) => ({ ...current, description: event.target.value }))} placeholder="Add the context your team will need" rows={5} /></label><label className="form-field"><span>Priority</span><select value={newTicket.priority} onChange={(event) => setNewTicket((current) => ({ ...current, priority: event.target.value as Ticket['priority'] }))}><option value="low">Low</option><option value="medium">Medium</option><option value="high">High</option><option value="urgent">Urgent</option></select></label><div className="modal-actions"><button className="secondary-button" type="button" onClick={() => setIsCreateOpen(false)}>Cancel</button><button className="primary-button" type="submit" disabled={isSaving || !newTicket.title.trim()}>{isSaving ? 'Saving…' : 'Create ticket'}</button></div></form></div>}
    </section>
  )
}
