import { FormEvent, useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Bot, CircleAlert, MessageCircle, RefreshCw, Search, Send, UserRound } from 'lucide-react'
import { API_BASE } from '../../lib/api'

type CommunicationPageProps = { token: string; onUnauthorized: () => void }

type Conversation = {
  id: number
  phone_number: string
  contact_name: string
  last_message: string
  last_message_at: string
  mode: 'auto' | 'human'
  priority?: string
}

type ChatMessage = {
  id: number
  text: string
  type: string
  direction: 'incoming' | 'outgoing'
  sender_role: string
  status: string
  sent_at: string
}

function formatMessageTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleTimeString('en-IN', { hour: 'numeric', minute: '2-digit' })
}

function formatConversationTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleDateString('en-IN', { day: 'numeric', month: 'short' })
}

function initials(name: string) {
  return name.split(' ').filter(Boolean).slice(0, 2).map((part) => part[0]).join('').toUpperCase() || '?'
}

export function CommunicationPage({ token, onUnauthorized }: CommunicationPageProps) {
  const [conversations, setConversations] = useState<Conversation[]>([])
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [search, setSearch] = useState('')
  const [messageText, setMessageText] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [isMessagesLoading, setIsMessagesLoading] = useState(false)
  const [isSending, setIsSending] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const request = useCallback(async (path: string, options: RequestInit = {}) => {
    const response = await fetch(`${API_BASE}${path}`, {
      ...options,
      headers: { Authorization: `Bearer ${token}`, ...(options.headers || {}) },
    })
    if (response.status === 401) {
      onUnauthorized()
      throw new Error('Your session has expired. Please sign in again.')
    }
    if (!response.ok) throw new Error(`Communication request failed with status ${response.status}`)
    return response
  }, [onUnauthorized, token])

  const loadConversations = useCallback(async () => {
    try {
      const response = await request('/api/automation/whatsapp/conversations')
      const data = await response.json() as Conversation[]
      if (!Array.isArray(data)) throw new Error('Conversations were not returned')
      setConversations(data)
      setSelectedId((current) => current ?? data[0]?.id ?? null)
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to load conversations')
    } finally {
      setIsLoading(false)
    }
  }, [request])

  const loadMessages = useCallback(async (conversationId: number, silent = false) => {
    if (!silent) setIsMessagesLoading(true)
    try {
      const response = await request(`/api/automation/whatsapp/chat?conversation_id=${conversationId}&limit=50&offset=0`)
      const data = await response.json() as ChatMessage[]
      setMessages(Array.isArray(data) ? data : [])
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to load messages')
    } finally {
      if (!silent) setIsMessagesLoading(false)
    }
  }, [request])

  useEffect(() => {
    void loadConversations()
    const interval = window.setInterval(() => void loadConversations(), 5000)
    return () => window.clearInterval(interval)
  }, [loadConversations])

  useEffect(() => {
    if (selectedId === null) {
      setMessages([])
      return undefined
    }
    void loadMessages(selectedId)
    const interval = window.setInterval(() => void loadMessages(selectedId, true), 5000)
    return () => window.clearInterval(interval)
  }, [loadMessages, selectedId])

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages.length, selectedId])

  const filteredConversations = useMemo(() => {
    const query = search.trim().toLowerCase()
    if (!query) return conversations
    return conversations.filter((conversation) => `${conversation.contact_name} ${conversation.phone_number} ${conversation.last_message}`.toLowerCase().includes(query))
  }, [conversations, search])
  const selectedConversation = conversations.find((conversation) => conversation.id === selectedId) || null

  const sendMessage = async (event: FormEvent) => {
    event.preventDefault()
    const text = messageText.trim()
    if (!selectedConversation || !text || isSending) return
    setIsSending(true)
    setError('')
    try {
      await request('/api/automation/whatsapp/send-message', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ phone_number: selectedConversation.phone_number, text }) })
      setMessageText('')
      setNotice('Message sent')
      await Promise.all([loadMessages(selectedConversation.id, true), loadConversations()])
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to send message')
    } finally {
      setIsSending(false)
    }
  }

  const toggleMode = async () => {
    if (!selectedConversation) return
    const nextMode = selectedConversation.mode === 'auto' ? 'human' : 'auto'
    try {
      await request('/api/automation/whatsapp/conversations/mode', { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ id: selectedConversation.id, mode: nextMode }) })
      setConversations((current) => current.map((conversation) => conversation.id === selectedConversation.id ? { ...conversation, mode: nextMode } : conversation))
      setNotice(nextMode === 'human' ? 'Human takeover enabled' : 'Automation resumed')
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to update conversation mode')
    }
  }

  return (
    <section className="workspace-page communication-page" aria-labelledby="communication-heading">
      <header className="workspace-page-header">
        <div><p className="eyebrow">Engagement / Communication</p><h2 id="communication-heading">Customer conversations, in one place.</h2><p>Move between WhatsApp conversations quickly, keep the context visible, and take over from automation when a person is needed.</p></div>
        <button className="secondary-button" type="button" onClick={() => void loadConversations()} disabled={isLoading}><RefreshCw size={15} className={isLoading ? 'spin' : undefined} aria-hidden="true" /> Refresh inbox</button>
      </header>
      {error && <div className="dashboard-error" role="alert"><CircleAlert size={18} aria-hidden="true" /><span>{error}</span><button type="button" onClick={() => { setError(''); void loadConversations() }}>Try again</button></div>}
      {notice && <div className="inventory-notice" role="status">{notice}</div>}

      <div className="inbox-shell">
        <aside className="conversation-list" aria-label="Conversations">
          <div className="conversation-list-heading"><div><p className="eyebrow">Inbox</p><strong>{conversations.length} conversations</strong></div><MessageCircle size={18} aria-hidden="true" /></div>
          <label className="orders-search"><Search size={16} aria-hidden="true" /><span className="sr-only">Search conversations</span><input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search name or number" /></label>
          <div className="conversation-items">
            {isLoading ? <p className="conversation-empty">Loading conversations…</p> : filteredConversations.length === 0 ? <p className="conversation-empty">No conversations found.</p> : filteredConversations.map((conversation) => (
              <button className={`conversation-item ${selectedId === conversation.id ? 'conversation-item-active' : ''}`} type="button" key={conversation.id} onClick={() => { setSelectedId(conversation.id); setNotice('') }}>
                <span className="conversation-avatar" aria-hidden="true">{initials(conversation.contact_name)}</span>
                <span className="conversation-item-copy"><strong>{conversation.contact_name || conversation.phone_number}</strong><small>{conversation.last_message || 'No messages yet'}</small></span>
                <span className="conversation-item-meta"><time>{formatConversationTime(conversation.last_message_at)}</time><i className={`mode-dot mode-dot-${conversation.mode}`} /></span>
              </button>
            ))}
          </div>
        </aside>

        <section className="chat-panel" aria-label={selectedConversation ? `Conversation with ${selectedConversation.contact_name || selectedConversation.phone_number}` : 'Conversation details'}>
          {selectedConversation ? <>
            <header className="chat-panel-heading"><div className="chat-contact"><span className="conversation-avatar" aria-hidden="true">{initials(selectedConversation.contact_name)}</span><div><h3>{selectedConversation.contact_name || 'Unknown contact'}</h3><p>{selectedConversation.phone_number}</p></div></div><button className={`mode-button mode-button-${selectedConversation.mode}`} type="button" onClick={() => void toggleMode()}><span className="mode-dot" aria-hidden="true" />{selectedConversation.mode === 'auto' ? 'Automation on' : 'Human takeover'}</button></header>
            <div className="messages-area" aria-live="polite">
              {isMessagesLoading ? <p className="conversation-empty">Loading messages…</p> : messages.length === 0 ? <div className="chat-empty"><MessageCircle size={24} aria-hidden="true" /><p>No messages in this conversation yet.</p></div> : messages.map((message) => <div className={`message-row message-row-${message.direction}`} key={message.id}><div className="message-bubble"><p>{message.text || `[${message.type} message]`}</p><time>{formatMessageTime(message.sent_at)}</time></div></div>)}
              <div ref={messagesEndRef} />
            </div>
            <form className="chat-composer" onSubmit={sendMessage}><label className="sr-only" htmlFor="message-input">Message</label><textarea id="message-input" value={messageText} onChange={(event) => setMessageText(event.target.value)} placeholder="Write a reply…" rows={1} onKeyDown={(event) => { if (event.key === 'Enter' && !event.shiftKey) { event.preventDefault(); event.currentTarget.form?.requestSubmit() } }} /><button className="primary-button" type="submit" disabled={isSending || !messageText.trim()}><Send size={16} aria-hidden="true" /> Send</button></form>
          </> : <div className="chat-empty chat-empty-large"><UserRound size={28} aria-hidden="true" /><h3>Choose a conversation</h3><p>Select a customer from the inbox to see the thread.</p></div>}
        </section>
      </div>
      <p className="communication-footnote"><Bot size={14} aria-hidden="true" /> Text replies are available in this migration slice. Media actions remain on the legacy communication screen until their v2 treatment is ready.</p>
    </section>
  )
}
