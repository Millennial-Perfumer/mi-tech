import { useCallback, useEffect, useMemo, useState } from 'react'
import { Check, Clipboard, CircleAlert, ExternalLink, MessageSquare, RefreshCw, Search, Send, Star, X } from 'lucide-react'
import { API_BASE } from '../../lib/api'

type FeedbackPageProps = { token: string; onUnauthorized: () => void }

type Feedback = {
  id: number
  order_id: number
  order_number: string
  customer_name: string
  customer_phone?: string
  rating: number
  message: string
  admin_comment?: string
  judgeme_posted?: boolean
  google_review_requested?: boolean
  created_at: string
}

type ScanCandidate = {
  id: number
  order_number: string
  customer_name: string
  customer_phone: string
  delivered_at: string
  feedback_url: string
}

function formatDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return date.toLocaleDateString('en-IN', { day: 'numeric', month: 'short', year: 'numeric' })
}

function Stars({ rating }: { rating: number }) {
  return <span className="feedback-stars" aria-label={`${rating} out of 5 stars`}>{Array.from({ length: 5 }, (_, index) => <Star key={index} size={15} fill={index < rating ? 'currentColor' : 'none'} aria-hidden="true" />)}</span>
}

export function FeedbackPage({ token, onUnauthorized }: FeedbackPageProps) {
  const [feedbacks, setFeedbacks] = useState<Feedback[]>([])
  const [search, setSearch] = useState('')
  const [ratingFilter, setRatingFilter] = useState('')
  const [selectedFeedback, setSelectedFeedback] = useState<Feedback | null>(null)
  const [adminComment, setAdminComment] = useState('')
  const [judgeMeEmail, setJudgeMeEmail] = useState('')
  const [scanResults, setScanResults] = useState<ScanCandidate[]>([])
  const [selectedScanIds, setSelectedScanIds] = useState<number[]>([])
  const [isScanOpen, setIsScanOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const [isWorking, setIsWorking] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')

  const request = useCallback(async (path: string, options: RequestInit = {}) => {
    const response = await fetch(`${API_BASE}${path}`, { ...options, headers: { Authorization: `Bearer ${token}`, ...(options.headers || {}) } })
    if (response.status === 401) { onUnauthorized(); throw new Error('Your session has expired. Please sign in again.') }
    if (!response.ok) throw new Error(`Feedback request failed with status ${response.status}`)
    return response
  }, [onUnauthorized, token])

  const loadFeedback = useCallback(async () => {
    setIsLoading(true)
    setError('')
    try {
      const response = await request('/api/orders/feedback')
      const data = await response.json() as { success?: boolean; feedback?: Feedback[]; message?: string }
      if (!data.success) throw new Error(data.message || 'Feedback was not returned')
      setFeedbacks(data.feedback || [])
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to load feedback')
    } finally { setIsLoading(false) }
  }, [request])

  useEffect(() => { void loadFeedback() }, [loadFeedback])

  const filteredFeedback = useMemo(() => {
    const query = search.trim().toLowerCase()
    return feedbacks.filter((feedback) => (!ratingFilter || String(feedback.rating) === ratingFilter) && (!query || `${feedback.customer_name} ${feedback.order_number} ${feedback.message}`.toLowerCase().includes(query))).sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
  }, [feedbacks, ratingFilter, search])

  const averageRating = feedbacks.length ? (feedbacks.reduce((sum, feedback) => sum + feedback.rating, 0) / feedbacks.length).toFixed(1) : '—'
  const positiveCount = feedbacks.filter((feedback) => feedback.rating >= 4).length

  const openFeedback = (feedback: Feedback) => {
    setSelectedFeedback(feedback)
    setAdminComment(feedback.admin_comment || '')
    setJudgeMeEmail('')
    setNotice('')
  }

  const saveComment = async () => {
    if (!selectedFeedback) return
    setIsWorking(true)
    try {
      await request(`/api/orders/feedback/comment?id=${selectedFeedback.id}`, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ admin_comment: adminComment }) })
      setNotice('Internal note saved')
      setFeedbacks((current) => current.map((feedback) => feedback.id === selectedFeedback.id ? { ...feedback, admin_comment: adminComment } : feedback))
      setSelectedFeedback((current) => current ? { ...current, admin_comment: adminComment } : current)
    } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to save note') } finally { setIsWorking(false) }
  }

  const postJudgeMe = async () => {
    if (!selectedFeedback) return
    setIsWorking(true)
    try {
      const response = await request(`/api/orders/feedback/post-judgeme?id=${selectedFeedback.id}`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ email: judgeMeEmail }) })
      const data = await response.json() as { message?: string }
      setNotice(data.message || 'Review sent to Judge.me')
      setFeedbacks((current) => current.map((feedback) => feedback.id === selectedFeedback.id ? { ...feedback, judgeme_posted: true } : feedback))
      setSelectedFeedback((current) => current ? { ...current, judgeme_posted: true } : current)
    } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to post to Judge.me') } finally { setIsWorking(false) }
  }

  const requestGoogleReview = async () => {
    if (!selectedFeedback) return
    setIsWorking(true)
    try {
      await request(`/api/orders/feedback/request-google-review?id=${selectedFeedback.id}`, { method: 'POST' })
      setNotice('Google review request recorded')
      setFeedbacks((current) => current.map((feedback) => feedback.id === selectedFeedback.id ? { ...feedback, google_review_requested: true } : feedback))
      setSelectedFeedback((current) => current ? { ...current, google_review_requested: true } : current)
    } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to record Google review request') } finally { setIsWorking(false) }
  }

  const copyText = async (text: string) => {
    if (!text) return
    try { await navigator.clipboard.writeText(text); setNotice('Review text copied') } catch { setError('Copy failed. Select the review text manually.') }
  }

  const scanForOrders = async () => {
    setIsWorking(true)
    setError('')
    try {
      const response = await request('/api/feedback/scan')
      const data = await response.json() as { success?: boolean; orders?: ScanCandidate[]; message?: string }
      if (!data.success) throw new Error(data.message || 'Feedback scan failed')
      const candidates = data.orders || []
      setScanResults(candidates)
      setSelectedScanIds(candidates.map((candidate) => candidate.id))
      setIsScanOpen(true)
      if (!candidates.length) setNotice('No new delivered orders are ready for feedback.')
    } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to scan for orders') } finally { setIsWorking(false) }
  }

  const sendFeedbackRequests = async () => {
    if (!selectedScanIds.length) return
    setIsWorking(true)
    try {
      const response = await request('/api/feedback/bulk-send', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ order_ids: selectedScanIds }) })
      const data = await response.json() as { sent?: number; failed?: number }
      setNotice(`${data.sent || 0} feedback requests sent${data.failed ? `, ${data.failed} failed` : ''}`)
      setIsScanOpen(false)
      setSelectedScanIds([])
      await loadFeedback()
    } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to send feedback requests') } finally { setIsWorking(false) }
  }

  return (
    <section className="workspace-page feedback-page" aria-labelledby="feedback-heading">
      <header className="workspace-page-header"><div><p className="eyebrow">Engagement / Feedback</p><h2 id="feedback-heading">Customer feedback</h2><p>Review feedback and follow up with customers.</p></div><div className="support-header-actions"><button className="secondary-button" type="button" onClick={() => void loadFeedback()} disabled={isLoading}><RefreshCw size={15} className={isLoading ? 'spin' : undefined} aria-hidden="true" /> Refresh</button><button className="primary-button" type="button" onClick={() => void scanForOrders()} disabled={isWorking}><Send size={15} aria-hidden="true" /> Scan delivered orders</button></div></header>
      {error && <div className="dashboard-error" role="alert"><CircleAlert size={18} aria-hidden="true" /><span>{error}</span><button type="button" onClick={() => { setError(''); void loadFeedback() }}>Try again</button></div>}
      {notice && <div className="inventory-notice" role="status">{notice}</div>}
      <div className="feedback-metrics-grid"><article className="report-metric-card"><span className="metric-label">Reviews</span><strong>{feedbacks.length}</strong><small>Collected responses</small></article><article className="report-metric-card"><span className="metric-label">Average rating</span><strong>{averageRating}<span className="metric-unit"> / 5</span></strong><small>Across all responses</small></article><article className="report-metric-card"><span className="metric-label">Positive</span><strong>{positiveCount}</strong><small>Rated 4 or 5 stars</small></article></div>
      <div className="feedback-toolbar"><label className="orders-search"><Search size={16} aria-hidden="true" /><span className="sr-only">Search feedback</span><input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search customer, order, or review" /></label><label className="compact-select"><span className="sr-only">Filter by rating</span><select value={ratingFilter} onChange={(event) => setRatingFilter(event.target.value)}><option value="">All ratings</option><option value="5">5 stars</option><option value="4">4 stars</option><option value="3">3 stars</option><option value="2">2 stars</option><option value="1">1 star</option></select></label></div>
      <div className="feedback-list" aria-live="polite">{isLoading ? <div className="empty-panel"><MessageSquare size={20} aria-hidden="true" /><p>Loading customer feedback…</p></div> : filteredFeedback.length === 0 ? <div className="empty-panel"><MessageSquare size={20} aria-hidden="true" /><div><h2>No feedback found</h2><p>Try another search or scan delivered orders for new responses.</p></div></div> : filteredFeedback.map((feedback) => <article className="feedback-card" key={feedback.id}><div className="feedback-card-heading"><div><span className="feedback-order">Order {feedback.order_number}</span><h3>{feedback.customer_name || 'Anonymous customer'}</h3></div><time>{formatDate(feedback.created_at)}</time></div><div className="feedback-card-rating"><Stars rating={feedback.rating} /><span>{feedback.rating}/5</span></div><p className="feedback-message">{feedback.message || 'No written comment.'}</p><div className="feedback-card-footer"><span>{feedback.judgeme_posted ? <><Check size={14} aria-hidden="true" /> Judge.me posted</> : 'Not posted to Judge.me'}</span><button className="text-button feedback-open-button" type="button" onClick={() => openFeedback(feedback)}>Open review <ExternalLink size={13} aria-hidden="true" /></button></div></article>)}</div>

      {selectedFeedback && <div className="modal-scrim" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) setSelectedFeedback(null) }}><div className="modal-card feedback-modal" role="dialog" aria-modal="true" aria-labelledby="feedback-detail-heading"><div className="modal-heading"><div><p className="eyebrow">Order {selectedFeedback.order_number}</p><h2 id="feedback-detail-heading">{selectedFeedback.customer_name || 'Anonymous customer'}</h2></div><button className="icon-button" type="button" aria-label="Close review" onClick={() => setSelectedFeedback(null)}><X size={19} aria-hidden="true" /></button></div><div className="feedback-detail-rating"><Stars rating={selectedFeedback.rating} /><strong>{selectedFeedback.rating}/5</strong><span>{formatDate(selectedFeedback.created_at)}</span></div><blockquote>{selectedFeedback.message || 'No written comment.'}</blockquote><div className="feedback-action-row"><button className="secondary-button" type="button" onClick={() => void copyText(selectedFeedback.message)}><Clipboard size={14} aria-hidden="true" /> Copy review</button>{selectedFeedback.customer_phone && <a className="secondary-button" href={`https://api.whatsapp.com/send?phone=${selectedFeedback.customer_phone.replace(/\D/g, '')}&text=${encodeURIComponent(selectedFeedback.message)}`} target="_blank" rel="noreferrer"><MessageSquare size={14} aria-hidden="true" /> Open WhatsApp</a>}</div><label className="form-field">Internal note<textarea rows={3} value={adminComment} onChange={(event) => setAdminComment(event.target.value)} placeholder="Add context for your team" /></label><button className="secondary-button" type="button" onClick={() => void saveComment()} disabled={isWorking}>Save note</button><div className="feedback-integrations"><div><span className="metric-label">Judge.me</span><p>{selectedFeedback.judgeme_posted ? 'Already posted' : 'Post this response as a product review.'}</p>{!selectedFeedback.judgeme_posted && <div className="feedback-email-row"><input aria-label="Judge.me email" type="email" value={judgeMeEmail} onChange={(event) => setJudgeMeEmail(event.target.value)} placeholder="Reviewer email (optional)" /><button className="primary-button" type="button" onClick={() => void postJudgeMe()} disabled={isWorking}><Send size={14} aria-hidden="true" /> Post</button></div>}</div><div><span className="metric-label">Google review</span><p>{selectedFeedback.google_review_requested ? 'Request recorded' : 'Record a Google review follow-up.'}</p>{!selectedFeedback.google_review_requested && <button className="secondary-button" type="button" onClick={() => void requestGoogleReview()} disabled={isWorking}>Record request</button>}</div></div></div></div>}
      {isScanOpen && <div className="modal-scrim" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) setIsScanOpen(false) }}><div className="modal-card scan-modal" role="dialog" aria-modal="true" aria-labelledby="scan-heading"><div className="modal-heading"><div><p className="eyebrow">Feedback requests</p><h2 id="scan-heading">Select delivered orders</h2><p className="modal-supporting-copy">Only orders with a customer phone number are included.</p></div><button className="icon-button" type="button" aria-label="Close order scan" onClick={() => setIsScanOpen(false)}><X size={19} aria-hidden="true" /></button></div><label className="scan-select-all"><input type="checkbox" checked={selectedScanIds.length === scanResults.length && scanResults.length > 0} onChange={(event) => setSelectedScanIds(event.target.checked ? scanResults.map((candidate) => candidate.id) : [])} /> Select all ({scanResults.length})</label><div className="scan-candidate-list">{scanResults.map((candidate) => <label className="scan-candidate" key={candidate.id}><input type="checkbox" checked={selectedScanIds.includes(candidate.id)} onChange={() => setSelectedScanIds((current) => current.includes(candidate.id) ? current.filter((id) => id !== candidate.id) : [...current, candidate.id])} /><span><strong>{candidate.customer_name || 'Unnamed customer'}</strong><small>Order {candidate.order_number} · Delivered {formatDate(candidate.delivered_at)}</small></span></label>)}</div><div className="modal-actions"><button className="secondary-button" type="button" onClick={() => setIsScanOpen(false)}>Cancel</button><button className="primary-button" type="button" onClick={() => void sendFeedbackRequests()} disabled={isWorking || !selectedScanIds.length}><Send size={14} aria-hidden="true" /> Send to {selectedScanIds.length}</button></div></div></div>}
    </section>
  )
}
