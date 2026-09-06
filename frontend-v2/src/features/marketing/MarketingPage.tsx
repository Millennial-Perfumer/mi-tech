import { useCallback, useEffect, useMemo, useState } from 'react'
import { ArrowLeft, ChevronRight, CircleAlert, ExternalLink, RefreshCw, Target, TrendingUp } from 'lucide-react'
import { API_BASE } from '../../lib/api'
import { usePeriodFilter } from '../../lib/usePeriodFilter'

type MarketingPageProps = { token: string; onUnauthorized: () => void }
type Level = 'campaigns' | 'adsets' | 'ads'
type RecordValue = Record<string, unknown>

function numberValue(value: unknown) { const parsed = Number(value); return Number.isFinite(parsed) ? parsed : 0 }
function textValue(value: unknown, fallback = '—') { return typeof value === 'string' && value ? value : fallback }
function formatMoney(value: unknown) { return `₹${numberValue(value).toLocaleString('en-IN', { maximumFractionDigits: 0 })}` }
function formatNumber(value: unknown) { return numberValue(value).toLocaleString('en-IN', { maximumFractionDigits: 0 }) }
function statusTone(status: string) { return status.toLowerCase() === 'active' ? 'success' : 'neutral' }

function insightFor(insights: RecordValue[], id: string, level: Level) {
  const key = level === 'campaigns' ? 'campaign_id' : level === 'adsets' ? 'adset_id' : 'ad_id'
  return insights.find((insight) => String(insight[key] || '') === id)
}

function aggregateInsights(insights: RecordValue[]) {
  const spend = insights.reduce((sum, insight) => sum + numberValue(insight.spend), 0)
  const purchaseValue = insights.reduce((sum, insight) => sum + numberValue(insight.purchase_value), 0)
  const conversions = insights.reduce((sum, insight) => sum + numberValue(insight.conversions), 0)
  const reach = insights.reduce((sum, insight) => sum + numberValue(insight.reach), 0)
  return { spend, purchaseValue, conversions, reach, roas: spend ? purchaseValue / spend : 0 }
}

export function MarketingPage({ token, onUnauthorized }: MarketingPageProps) {
  const { startDate, endDate } = usePeriodFilter()
  const [level, setLevel] = useState<Level>('campaigns')
  const [campaigns, setCampaigns] = useState<RecordValue[]>([])
  const [adsets, setAdsets] = useState<RecordValue[]>([])
  const [ads, setAds] = useState<RecordValue[]>([])
  const [insights, setInsights] = useState<RecordValue[]>([])
  const [summary, setSummary] = useState<RecordValue | null>(null)
  const [activeAccountId, setActiveAccountId] = useState('')
  const [accountName, setAccountName] = useState('Meta ad account')
  const [campaignContext, setCampaignContext] = useState<{ id: string; name: string } | null>(null)
  const [adsetContext, setAdsetContext] = useState<{ id: string; name: string } | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState('')

  const requestJson = useCallback(async (path: string) => {
    const response = await fetch(`${API_BASE}${path}`, { headers: { Authorization: `Bearer ${token}` } })
    if (response.status === 401) { onUnauthorized(); throw new Error('Your session has expired. Please sign in again.') }
    if (!response.ok) throw new Error(`Marketing request failed with status ${response.status}`)
    return response.json() as Promise<RecordValue>
  }, [onUnauthorized, token])

  const loadOverview = useCallback(async () => {
    setIsLoading(true)
    setError('')
    try {
      const data = await requestJson(`/api/marketing/meta/overview?start_date=${startDate}&end_date=${endDate}`)
      if (!data.success) throw new Error(textValue(data.message, 'Meta overview was not returned'))
      const overviewInsights = (Array.isArray(data.insights) ? data.insights : []) as RecordValue[]
      const accounts = (Array.isArray(data.accounts) ? data.accounts : []) as RecordValue[]
      const configuredId = textValue(data.active_id, '')
      setInsights(overviewInsights)
      setSummary(Array.isArray(data.summary) && data.summary.length ? data.summary[0] as RecordValue : null)
      setActiveAccountId(configuredId)
      const selectedAccount = accounts.find((account) => textValue(account.id, '') === configuredId)
      setAccountName(textValue(selectedAccount?.name, configuredId ? 'Meta ad account' : 'No Meta ad account configured'))
      if (configuredId) {
        const campaignData = await requestJson(`/api/marketing/meta/campaigns?ad_account_id=${encodeURIComponent(configuredId)}`)
        setCampaigns((Array.isArray(campaignData.campaigns) ? campaignData.campaigns : []) as RecordValue[])
      } else setCampaigns([])
      setAdsets([]); setAds([]); setCampaignContext(null); setAdsetContext(null); setLevel('campaigns')
    } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to load Meta marketing') } finally { setIsLoading(false) }
  }, [endDate, requestJson, startDate])

  useEffect(() => { void loadOverview() }, [loadOverview])

  const loadAdsets = async (campaign: RecordValue) => {
    const id = textValue(campaign.id, '')
    if (!id) return
    setIsLoading(true); setError('')
    try {
      const data = await requestJson(`/api/marketing/meta/adsets?campaign_id=${encodeURIComponent(id)}&start_date=${startDate}&end_date=${endDate}`)
      if (!data.success) throw new Error(textValue(data.message, 'Ad sets were not returned'))
      setAdsets((Array.isArray(data.adsets) ? data.adsets : []) as RecordValue[])
      setInsights((Array.isArray(data.insights) ? data.insights : []) as RecordValue[])
      setCampaignContext({ id, name: textValue(campaign.name, 'Campaign') }); setAdsetContext(null); setLevel('adsets')
    } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to load ad sets') } finally { setIsLoading(false) }
  }

  const loadAds = async (adset: RecordValue) => {
    const id = textValue(adset.id, '')
    if (!id) return
    setIsLoading(true); setError('')
    try {
      const data = await requestJson(`/api/marketing/meta/ads?adset_id=${encodeURIComponent(id)}&start_date=${startDate}&end_date=${endDate}`)
      if (!data.success) throw new Error(textValue(data.message, 'Ads were not returned'))
      setAds((Array.isArray(data.ads) ? data.ads : []) as RecordValue[])
      setInsights((Array.isArray(data.insights) ? data.insights : []) as RecordValue[])
      setAdsetContext({ id, name: textValue(adset.name, 'Ad set') }); setLevel('ads')
    } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to load ads') } finally { setIsLoading(false) }
  }

  const backToCampaigns = () => { void loadOverview() }
  const backToAdsets = async () => {
    if (campaignContext) {
      const campaign = campaigns.find((item) => textValue(item.id, '') === campaignContext.id)
      if (campaign) await loadAdsets(campaign)
    }
  }

  const metrics = useMemo(() => aggregateInsights(summary ? [summary] : insights), [insights, summary])
  const rows = level === 'campaigns' ? campaigns : level === 'adsets' ? adsets : ads

  return (
    <section className="workspace-page marketing-page" aria-labelledby="marketing-heading">
      <header className="workspace-page-header"><div><p className="eyebrow">Engagement / Marketing</p><h2 id="marketing-heading">Make paid growth easier to read.</h2><p>Move from account performance to campaigns, ad sets, and creatives with the same period filter used across the workspace.</p></div><button className="secondary-button" type="button" onClick={() => void (level === 'campaigns' ? loadOverview() : level === 'adsets' && campaignContext ? loadAdsets({ id: campaignContext.id, name: campaignContext.name }) : adsetContext ? loadAds({ id: adsetContext.id, name: adsetContext.name }) : loadOverview())} disabled={isLoading}><RefreshCw size={15} className={isLoading ? 'spin' : undefined} aria-hidden="true" /> Refresh</button></header>
      {error && <div className="dashboard-error" role="alert"><CircleAlert size={18} aria-hidden="true" /><span>{error}</span><button type="button" onClick={() => void loadOverview()}>Try again</button></div>}
      <div className="marketing-account-bar"><div className="marketing-account-title"><span className="marketing-icon"><Target size={17} aria-hidden="true" /></span><span><strong>{accountName}</strong><small>{activeAccountId ? `Account ${activeAccountId}` : 'Connect Meta in Settings to see campaign data'}</small></span></div><span className="page-period-note">{startDate} → {endDate}</span></div>
      <div className="dashboard-metrics-grid marketing-metrics"><article className="dashboard-metric-card"><div className="dashboard-metric-heading"><span className="metric-label">Spend</span><span className="dashboard-metric-icon">₹</span></div><strong>{formatMoney(metrics.spend)}</strong><small className="metric-detail">Selected period</small></article><article className="dashboard-metric-card"><div className="dashboard-metric-heading"><span className="metric-label">ROAS</span><span className="dashboard-metric-icon"><TrendingUp size={14} aria-hidden="true" /></span></div><strong>{metrics.roas.toFixed(2)}x</strong><small className="metric-detail">Return on spend</small></article><article className="dashboard-metric-card"><div className="dashboard-metric-heading"><span className="metric-label">Conversions</span><span className="dashboard-metric-icon">↗</span></div><strong>{formatNumber(metrics.conversions)}</strong><small className="metric-detail">Tracked results</small></article><article className="dashboard-metric-card"><div className="dashboard-metric-heading"><span className="metric-label">Reach</span><span className="dashboard-metric-icon">◌</span></div><strong>{formatNumber(metrics.reach)}</strong><small className="metric-detail">People reached</small></article></div>
      <div className="marketing-breadcrumbs"><button type="button" onClick={backToCampaigns} className={level === 'campaigns' ? 'marketing-crumb-active' : ''}>All campaigns</button>{campaignContext && <><ChevronRight size={14} aria-hidden="true" /><button type="button" onClick={backToAdsets} className={level === 'adsets' ? 'marketing-crumb-active' : ''}>{campaignContext.name}</button></>}{adsetContext && <><ChevronRight size={14} aria-hidden="true" /><span className="marketing-crumb-active">{adsetContext.name}</span></>}</div>
      <section className="reports-table-card marketing-table-card"><div className="reports-table-heading"><div><p className="eyebrow">{level === 'campaigns' ? 'Campaigns' : level === 'adsets' ? 'Ad sets' : 'Ads'}</p><h3>{rows.length} {level}</h3></div><span className="orders-card-meta">Click a row to drill down</span></div><div className="orders-table-wrap"><table className="orders-table marketing-table"><caption className="sr-only">Meta marketing performance</caption><thead><tr><th>Name</th><th>Status</th><th>{level === 'campaigns' ? 'Objective' : level === 'adsets' ? 'Budget' : 'ROAS'}</th><th>Spend</th><th>Reach</th><th>Action</th></tr></thead><tbody>{isLoading ? <tr><td className="table-state" colSpan={6}>Loading Meta data…</td></tr> : rows.length === 0 ? <tr><td className="table-state" colSpan={6}>No {level} available for this account or period.</td></tr> : rows.map((row) => { const id = textValue(row.id, ''); const insight = insightFor(insights, id, level); const status = textValue(row.effective_status, textValue(row.status)); const name = textValue(row.name, 'Untitled'); const click = () => level === 'campaigns' ? void loadAdsets(row) : level === 'adsets' ? void loadAds(row) : undefined; return <tr key={id} className={level !== 'ads' ? 'marketing-clickable-row' : undefined} onClick={click}><td><div className="marketing-name-cell"><strong>{name}</strong><small>{id}</small></div></td><td><span className={`status-pill status-pill-${statusTone(status)}`}>{status}</span></td><td>{level === 'campaigns' ? textValue(row.objective, '—').replace(/_/g, ' ') : level === 'adsets' ? formatMoney(numberValue(row.daily_budget || row.lifetime_budget) / 100) : `${numberValue(insight?.purchase_roas_val).toFixed(2)}x`}</td><td className="table-money">{formatMoney(insight?.spend)}</td><td className="table-money">{formatNumber(insight?.reach)}</td><td>{level === 'ads' && activeAccountId ? <a className="table-link-button" href={`https://adsmanager.facebook.com/adsmanager/manage/ads?act=${activeAccountId}&selected_ad_ids=${id}`} target="_blank" rel="noreferrer" onClick={(event) => event.stopPropagation()}><ExternalLink size={14} aria-hidden="true" /> Open</a> : <button className="table-link-button" type="button" onClick={(event) => { event.stopPropagation(); click() }}>{level === 'campaigns' ? 'Ad sets' : 'Ads'} <ChevronRight size={14} aria-hidden="true" /></button>}</td></tr> })}</tbody></table></div></section>
      {level !== 'campaigns' && <button className="secondary-button marketing-back-button" type="button" onClick={() => level === 'ads' ? void backToAdsets() : backToCampaigns}><ArrowLeft size={15} aria-hidden="true" /> Back</button>}
    </section>
  )
}
