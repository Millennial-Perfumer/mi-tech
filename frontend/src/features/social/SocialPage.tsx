import { ChangeEvent, FormEvent, useCallback, useEffect, useMemo, useState } from 'react'
import { BarChart3, CircleAlert, ExternalLink, Eye, FileUp, Plus, RefreshCw, Send, X } from 'lucide-react'
import { apiJson, apiRequest, arrayFrom, formatDate, numberValue, textValue } from '../../lib/http'
import { usePeriodFilter } from '../../lib/usePeriodFilter'

type Props = { token: string; onUnauthorized: () => void; initialTab?: 'overview' | 'queue' }
type Platform = 'instagram' | 'facebook' | 'threads'
type Row = Record<string, unknown>
type ComposerMode = 'queue' | 'publish'
type Composer = { caption: string; hashtags: string; post_type: string; target_platforms: string; media_urls: string; platform: Platform; page_id: string; ig_id: string; threads_id: string }

const emptyComposer: Composer = { caption: '', hashtags: '', post_type: 'SINGLE_PHOTO', target_platforms: 'instagram', media_urls: '', platform: 'instagram', page_id: '', ig_id: '', threads_id: '' }

function metricValue(overview: Row, key: string) { return numberValue(overview[key] || (overview.account as Row | undefined)?.[key]) }

function listValue(value: unknown) {
  if (Array.isArray(value)) return value.map(String).join(', ')
  if (typeof value === 'string') {
    try { const parsed = JSON.parse(value) as unknown; if (Array.isArray(parsed)) return parsed.map(String).join(', ') } catch { return value }
    return value
  }
  return '—'
}

export function SocialPage({ token, onUnauthorized, initialTab = 'overview' }: Props) {
  const { startDate, endDate } = usePeriodFilter()
  const [tab, setTab] = useState(initialTab)
  const [platform, setPlatform] = useState<Platform>('instagram')
  const [overview, setOverview] = useState<Row>({})
  const [posts, setPosts] = useState<Row[]>([])
  const [queue, setQueue] = useState<Row[]>([])
  const [health, setHealth] = useState<Row>({})
  const [isLoading, setIsLoading] = useState(true)
  const [isWorking, setIsWorking] = useState(false)
  const [isInsightsLoading, setIsInsightsLoading] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [composerMode, setComposerMode] = useState<ComposerMode | null>(null)
  const [composer, setComposer] = useState<Composer>(emptyComposer)
  const [files, setFiles] = useState<File[]>([])
  const [selectedPost, setSelectedPost] = useState<Row>()
  const [insights, setInsights] = useState<Row>({})

  const loadOverview = useCallback(async () => {
    setIsLoading(true); setError('')
    try {
      const data = await apiJson<Row>(token, onUnauthorized, `/api/marketing/smm/overview?platform=${platform}&start_date=${startDate}&end_date=${endDate}`)
      if (data.success === false) throw new Error(textValue(data.error, 'Social overview was not returned'))
      const value = (data.overview && typeof data.overview === 'object' ? data.overview : data) as Row
      setOverview(value); setPosts(arrayFrom(value, 'posts'))
    } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to load social overview') } finally { setIsLoading(false) }
  }, [endDate, onUnauthorized, platform, startDate, token])

  const loadQueue = useCallback(async () => {
    try { const data = await apiJson<unknown>(token, onUnauthorized, '/api/marketing/smm/queue'); setQueue(arrayFrom(data, 'posts')) } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to load social queue') }
  }, [onUnauthorized, token])

  useEffect(() => { void loadOverview(); void loadQueue(); void apiJson<Row>(token, onUnauthorized, '/api/marketing/smm/health').then((data) => setHealth((data.health || data) as Row)).catch(() => undefined) }, [loadOverview, loadQueue, onUnauthorized, token])

  const sync = async () => { setIsWorking(true); try { await apiRequest(token, onUnauthorized, `/api/marketing/smm/sync?platform=${platform}`, { method: 'POST' }); setNotice(`${platform} data synced`); await loadOverview() } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to sync social data') } finally { setIsWorking(false) } }

  const openComposer = (mode: ComposerMode) => { setComposerMode(mode); setComposer({ ...emptyComposer, platform }); setFiles([]) }

  const createQueuePost = async (event: FormEvent) => {
    event.preventDefault(); setIsWorking(true); setError('')
    try {
      const targetPlatforms = composer.target_platforms.split(',').map((value) => value.trim()).filter(Boolean)
      if (targetPlatforms.length === 0) throw new Error('Select at least one destination platform')
      const postType = files.some((file) => file.type.startsWith('video/')) ? 'VIDEO' : files.length > 1 ? 'CAROUSEL' : composer.post_type
      if (files.length > 0) {
        const body = new FormData()
        body.append('caption', composer.caption)
        body.append('hashtags', composer.hashtags)
        body.append('post_type', postType)
        body.append('target_platforms', JSON.stringify(targetPlatforms))
        files.forEach((file) => body.append('files', file))
        await apiRequest(token, onUnauthorized, '/api/marketing/smm/queue', { method: 'POST', body })
      } else {
        await apiRequest(token, onUnauthorized, '/api/marketing/smm/queue', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ caption: composer.caption, hashtags: composer.hashtags.split(/[;,\s]+/).map((value) => value.replace(/^#/, '')).filter(Boolean).join(', '), post_type: postType, target_platforms: targetPlatforms, media_urls: composer.media_urls.split('\n').map((value) => value.trim()).filter(Boolean) }) })
      }
      setNotice('Post added to the publishing queue'); setComposerMode(null); setComposer(emptyComposer); setFiles([]); await loadQueue()
    } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to add post to queue') } finally { setIsWorking(false) }
  }

  const publishNow = async (event: FormEvent) => {
    event.preventDefault(); setIsWorking(true); setError('')
    try {
      if (composer.platform === 'threads' && composer.caption.length > 500) throw new Error('Threads posts must be 500 characters or fewer')
      const imageUrl = composer.media_urls.split('\n').map((value) => value.trim()).find(Boolean) || ''
      await apiRequest(token, onUnauthorized, '/api/marketing/smm/post', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ platform: composer.platform, page_id: composer.page_id, ig_id: composer.ig_id, threads_id: composer.threads_id, message: composer.caption, caption: composer.caption, text: composer.caption, image_url: imageUrl }) })
      setNotice(`${composer.platform} post published`); setComposerMode(null); setComposer(emptyComposer); await loadOverview()
    } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to publish post') } finally { setIsWorking(false) }
  }

  const handleFiles = (event: ChangeEvent<HTMLInputElement>) => { if (event.target.files) setFiles(Array.from(event.target.files)) }

  const openInsights = async (post: Row) => {
    const postId = textValue(post.post_id || post.id, '')
    if (!postId) return
    setSelectedPost(post); setInsights({}); setIsInsightsLoading(true)
    try { const data = await apiJson<Row>(token, onUnauthorized, `/api/marketing/smm/post/insights?id=${encodeURIComponent(postId)}&media_type=${encodeURIComponent(textValue(post.media_type || post.type, ''))}`); setInsights((data.insights || data) as Row) } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to load post insights') } finally { setIsInsightsLoading(false) }
  }

  const summary = useMemo(() => [['Followers', metricValue(overview, 'follower_count')], ['Reach', metricValue(overview, 'total_reach')], ['Views', metricValue(overview, 'total_views')], ['Engagement', metricValue(overview, 'total_engagement')]], [overview])

  return <section className="workspace-page social-page" aria-labelledby="social-heading"><header className="workspace-page-header"><div><p className="eyebrow">Engagement / Social</p><h2 id="social-heading">Social media</h2><p>Review channel performance and manage publishing.</p></div><div className="support-header-actions"><button className="secondary-button" type="button" onClick={() => void (tab === 'queue' ? loadQueue() : loadOverview())} disabled={isLoading}><RefreshCw size={15} className={isLoading ? 'spin' : undefined} aria-hidden="true" /> Refresh</button><button className="secondary-button" type="button" onClick={() => openComposer('publish')}><Send size={15} aria-hidden="true" /> Publish now</button><button className="primary-button" type="button" onClick={() => openComposer('queue')}><Plus size={15} aria-hidden="true" /> Queue post</button></div></header>{error && <div className="dashboard-error" role="alert"><CircleAlert size={18} aria-hidden="true" /><span>{error}</span><button type="button" onClick={() => { setError(''); void loadOverview() }}>Try again</button></div>}{notice && <div className="inventory-notice" role="status">{notice}</div>}<section className="social-control-bar"><div className="b2b-tabs" role="tablist" aria-label="Social sections"><button className={`filter-chip ${tab === 'overview' ? 'filter-chip-active' : ''}`} type="button" role="tab" aria-selected={tab === 'overview'} onClick={() => setTab('overview')}>Overview</button><button className={`filter-chip ${tab === 'queue' ? 'filter-chip-active' : ''}`} type="button" role="tab" aria-selected={tab === 'queue'} onClick={() => setTab('queue')}>Publishing queue</button></div>{tab === 'overview' && <div className="platform-tabs" role="tablist" aria-label="Social platforms">{(['instagram', 'facebook', 'threads'] as Platform[]).map((item) => <button key={item} className={`filter-chip ${platform === item ? 'filter-chip-active' : ''}`} type="button" role="tab" aria-selected={platform === item} onClick={() => setPlatform(item)}>{item}</button>)}</div>}</section>{tab === 'overview' ? <><div className="dashboard-metrics-grid social-metrics">{summary.map(([label, value]) => <article className="dashboard-metric-card" key={label}><div className="dashboard-metric-heading"><span className="metric-label">{label}</span><span className="dashboard-metric-icon"><BarChart3 size={14} aria-hidden="true" /></span></div><strong>{Number(value).toLocaleString('en-IN')}</strong><small className="metric-detail">{platform} · {startDate} to {endDate}</small></article>)}</div><div className="social-health-row"><span className={`status-pill status-pill-${textValue(health.status, 'unknown').toLowerCase() === 'healthy' ? 'success' : 'neutral'}`}>{textValue(health.status, 'Connection status unknown')}</span><span>{textValue(health.message, 'Channel health is checked from the connected account.')}</span><button className="secondary-button" type="button" onClick={() => void sync()} disabled={isWorking}><RefreshCw size={14} aria-hidden="true" /> Sync {platform}</button></div><section className="reports-table-card"><div className="reports-table-heading"><div><p className="eyebrow">Published content</p><h3>{posts.length} posts in the selected period</h3></div></div><div className="orders-table-wrap"><table className="orders-table"><caption className="sr-only">Published social posts</caption><thead><tr><th>Published</th><th>Content</th><th>Reach</th><th>Views</th><th>Engagement</th><th>Actions</th></tr></thead><tbody>{isLoading ? <tr><td className="table-state" colSpan={6}>Loading social content…</td></tr> : posts.length === 0 ? <tr><td className="table-state" colSpan={6}>No published content found for this period.</td></tr> : posts.map((post, index) => <tr key={textValue(post.id || post.post_id, String(index))}><td>{formatDate(post.published_at || post.timestamp)}</td><td><strong>{textValue(post.content || post.caption, 'Untitled post').slice(0, 90)}</strong></td><td className="table-money">{numberValue(post.reach).toLocaleString('en-IN')}</td><td className="table-money">{numberValue(post.views).toLocaleString('en-IN')}</td><td className="table-money">{numberValue(post.engagement || post.total_interactions).toLocaleString('en-IN')}</td><td><div className="table-action-group"><button className="table-link-button" type="button" onClick={() => void openInsights(post)}><Eye size={13} aria-hidden="true" /> Insights</button>{textValue(post.permalink, '') && <a className="table-link-button" href={textValue(post.permalink, '#')} target="_blank" rel="noreferrer">Open <ExternalLink size={13} aria-hidden="true" /></a>}</div></td></tr>)}</tbody></table></div></section></> : <section className="reports-table-card"><div className="reports-table-heading"><div><p className="eyebrow">Social queue</p><h3>{queue.length} queued posts</h3></div><button className="primary-button" type="button" onClick={() => openComposer('queue')}><Plus size={14} aria-hidden="true" /> Add post</button></div><div className="orders-table-wrap"><table className="orders-table"><caption className="sr-only">Queued social posts</caption><thead><tr><th>Caption</th><th>Platforms</th><th>Type</th><th>Status</th><th>Created</th></tr></thead><tbody>{queue.length === 0 ? <tr><td className="table-state" colSpan={5}>The publishing queue is empty.</td></tr> : queue.map((post, index) => <tr key={textValue(post.id, String(index))}><td><strong>{textValue(post.caption || post.content, 'Untitled post').slice(0, 100)}</strong><small className="table-subtext">{textValue(post.hashtags, '')}</small></td><td>{listValue(post.target_platforms)}</td><td>{textValue(post.post_type, '—')}</td><td><span className="status-pill status-pill-neutral">{textValue(post.status, 'queued')}</span></td><td>{formatDate(post.created_at)}</td></tr>)}</tbody></table></div></section>}{composerMode && <div className="modal-scrim" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) setComposerMode(null) }}><form className="modal-card social-composer-modal" role="dialog" aria-modal="true" aria-labelledby="social-composer-heading" onSubmit={composerMode === 'queue' ? createQueuePost : publishNow}><div className="modal-heading"><div><p className="eyebrow">Social / {composerMode === 'queue' ? 'Publishing queue' : 'Live publishing'}</p><h2 id="social-composer-heading">{composerMode === 'queue' ? 'Queue a social post' : 'Publish to a channel'}</h2></div><button className="icon-button" type="button" aria-label="Close" onClick={() => setComposerMode(null)}><X size={19} aria-hidden="true" /></button></div><label className="form-field"><span>Caption or message</span><textarea required rows={4} maxLength={composerMode === 'publish' && composer.platform === 'threads' ? 500 : undefined} value={composer.caption} onChange={(event) => setComposer({ ...composer, caption: event.target.value })} placeholder="Write a clear, short message" /><small className="form-help">{composer.caption.length} characters{composerMode === 'publish' && composer.platform === 'threads' ? ' / 500 for Threads' : ''}</small></label>{composerMode === 'queue' ? <><div className="form-grid-two"><label className="form-field"><span>Platforms</span><input value={composer.target_platforms} onChange={(event) => setComposer({ ...composer, target_platforms: event.target.value })} placeholder="instagram, facebook" /></label><label className="form-field"><span>Post type</span><select value={composer.post_type} onChange={(event) => setComposer({ ...composer, post_type: event.target.value })}><option value="SINGLE_PHOTO">Single photo</option><option value="CAROUSEL">Carousel</option><option value="VIDEO">Video</option></select></label></div><label className="form-field"><span>Hashtags</span><input value={composer.hashtags} onChange={(event) => setComposer({ ...composer, hashtags: event.target.value })} placeholder="#millennialperfumer #fragrance" /></label><label className="form-field"><span>Upload media <small>(optional)</small></span><input type="file" accept="image/*,video/*" multiple onChange={handleFiles} /><small className="form-help">{files.length > 0 ? `${files.length} file${files.length === 1 ? '' : 's'} selected` : 'Files are streamed to the queue without changing your local files.'}</small></label><label className="form-field"><span>Media URLs <small>(optional, one per line)</small></span><textarea rows={2} value={composer.media_urls} onChange={(event) => setComposer({ ...composer, media_urls: event.target.value })} /></label></> : <><label className="form-field"><span>Platform</span><select value={composer.platform} onChange={(event) => setComposer({ ...composer, platform: event.target.value as Platform })}><option value="instagram">Instagram</option><option value="facebook">Facebook</option><option value="threads">Threads</option></select></label><div className="form-grid-two"><label className="form-field"><span>Image URL <small>(required for Instagram)</small></span><input value={composer.media_urls} onChange={(event) => setComposer({ ...composer, media_urls: event.target.value })} placeholder="https://…" /></label><label className="form-field"><span>{composer.platform === 'facebook' ? 'Facebook page ID' : composer.platform === 'threads' ? 'Threads user ID' : 'Instagram user ID'}</span><input value={composer.platform === 'facebook' ? composer.page_id : composer.platform === 'threads' ? composer.threads_id : composer.ig_id} onChange={(event) => setComposer({ ...composer, [composer.platform === 'facebook' ? 'page_id' : composer.platform === 'threads' ? 'threads_id' : 'ig_id']: event.target.value })} placeholder="Configured account ID" /></label></div></>}<div className="modal-actions"><button className="secondary-button" type="button" onClick={() => setComposerMode(null)}>Cancel</button><button className="primary-button" type="submit" disabled={isWorking}>{composerMode === 'queue' ? <><FileUp size={14} aria-hidden="true" /> Add to queue</> : <><Send size={14} aria-hidden="true" /> Publish now</>}</button></div></form></div>}{selectedPost && <div className="modal-scrim" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) setSelectedPost(undefined) }}><section className="modal-card social-insights-modal" role="dialog" aria-modal="true" aria-labelledby="social-insights-heading"><div className="modal-heading"><div><p className="eyebrow">{platform} / post insights</p><h2 id="social-insights-heading">{textValue(selectedPost.content || selectedPost.caption, 'Untitled post').slice(0, 70)}</h2></div><button className="icon-button" type="button" aria-label="Close insights" onClick={() => setSelectedPost(undefined)}><X size={19} aria-hidden="true" /></button></div>{isInsightsLoading ? <div className="empty-panel"><RefreshCw className="spin" size={18} aria-hidden="true" /><p>Loading post insights…</p></div> : <div className="social-insight-grid">{Object.entries((insights.metrics && typeof insights.metrics === 'object' ? insights.metrics : insights)).map(([key, value]) => <div className="inventory-summary-card" key={key}><span className="metric-label">{key.replaceAll('_', ' ')}</span><strong>{numberValue(value).toLocaleString('en-IN')}</strong></div>)}</div>}</section></div>}</section>
}
