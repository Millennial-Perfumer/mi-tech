import { FormEvent, useCallback, useEffect, useMemo, useState } from 'react'
import { Bot, Building2, ChevronDown, CircleAlert, Copy, CreditCard, Eye, EyeOff, KeyRound, MessageCircle, Megaphone, RefreshCw, Save, Settings2, ShieldCheck, ShoppingBag, Store, Workflow, Wrench, type LucideIcon } from 'lucide-react'
import { apiJson, apiRequest, arrayFrom, formatDate, numberValue, textValue } from '../../lib/http'

type Props = { token: string; onUnauthorized: () => void }
type Row = Record<string, unknown>
type MachineKey = Row & { scopes?: string[] }
type AppConfig = Row & { key?: string; value?: string; is_secret?: boolean; label?: string; category?: string }
type CustomerScope = readonly [string, string]
type ServiceStatus = { label: 'Configured' | 'Needs attention' | 'Not configured'; tone: 'success' | 'warning' | 'neutral' }
type ServiceDefinition = { id: string; title: string; description: string; categories: string[]; requiredKeys?: string[]; icon: LucideIcon }

const machineScopes: CustomerScope[] = [
  ['orders:read', 'Orders'], ['customers:read', 'Customers'], ['metrics:read', 'Metrics'], ['inventory:read', 'Inventory'], ['production:read', 'Production'], ['b2b:read', 'B2B billing'], ['communication:read', 'WhatsApp'], ['marketing:read', 'Marketing'], ['feedback:read', 'Feedback'], ['planner:read', 'Planner'], ['support:read', 'Support'], ['ai:read', 'AI'], ['settings:read', 'Settings'], ['system:read', 'System'],
  ['orders:write', 'Orders write'], ['customers:write', 'Customers write'], ['inventory:write', 'Inventory write'], ['production:write', 'Production write'], ['planner:write', 'Planner write'], ['b2b:write', 'B2B write'], ['communication:write', 'WhatsApp write'], ['marketing:write', 'Marketing write'], ['feedback:write', 'Feedback write'], ['support:write', 'Support write'], ['ai:write', 'AI write'], ['marketing:publish', 'Social publish'],
  ['orders:destructive', 'Orders delete'], ['customers:destructive', 'Customers delete'], ['inventory:destructive', 'Inventory delete'], ['production:destructive', 'Production delete'], ['planner:destructive', 'Planner delete'], ['b2b:destructive', 'B2B delete'], ['communication:destructive', 'WhatsApp delete'], ['ai:destructive', 'AI delete'],
]

const serviceDefinitions: ServiceDefinition[] = [
  { id: 'shopify', title: 'Shopify', description: 'Orders, inventory, and customer synchronization.', categories: ['shopify'], requiredKeys: ['shopify_store_url', 'shopify_access_token'], icon: Store },
  { id: 'amazon', title: 'Amazon', description: 'Marketplace orders and seller account synchronization.', categories: ['amazon'], requiredKeys: ['amazon_lwa_client_id', 'amazon_lwa_client_secret', 'amazon_lwa_refresh_token'], icon: ShoppingBag },
  { id: 'whatsapp', title: 'WhatsApp', description: 'Customer messaging, invoices, and automation.', categories: ['whatsapp'], requiredKeys: ['whatsapp_phone_number_id', 'whatsapp_waba_id'], icon: MessageCircle },
  { id: 'meta', title: 'Meta services', description: 'Shared Meta credentials, marketing, and social publishing.', categories: ['meta_shared', 'marketing', 'social_media'], requiredKeys: ['meta_app_id', 'meta_system_user_token'], icon: Megaphone },
  { id: 'ai', title: 'AI provider', description: 'Cloud and local model configuration.', categories: ['ai'], requiredKeys: ['ai_provider', 'ai_enabled'], icon: Bot },
  { id: 'payments', title: 'Payments', description: 'Payment collection and webhook configuration.', categories: ['payment'], requiredKeys: ['razorpay_key_id', 'razorpay_key_secret'], icon: CreditCard },
  { id: 'automation', title: 'Feedback & automation', description: 'Feedback links, recovery, and scheduled automation.', categories: ['feedback', 'auto_queue', 'abandoned_cart'], icon: Workflow },
  { id: 'business', title: 'Business profile', description: 'Business identity, tax, and billing details.', categories: ['business', 'b2b'], icon: Building2 },
  { id: 'planning', title: 'Planning', description: 'Planner and board defaults.', categories: ['kanban'], icon: Settings2 },
]

const advancedDefinition: ServiceDefinition = {
  id: 'advanced',
  title: 'Advanced configuration',
  description: 'System-level values that do not belong to a connected service.',
  categories: ['system'],
  icon: Wrench,
}

function stringValue(value: unknown, fallback = '') {
  if (value === null || value === undefined) return fallback
  return String(value)
}

function configKey(config: AppConfig, index = 0) {
  return stringValue(config.key, String(index))
}

function configLabel(config: AppConfig, index = 0) {
  const key = configKey(config, index)
  return stringValue(config.label, key.replace(/_/g, ' '))
}

function isSecretConfig(config: AppConfig) {
  return config.is_secret === true
}

function hasConfigValue(config: AppConfig) {
  return stringValue(config.value).trim() !== ''
}

function getServiceStatus(fields: AppConfig[], requiredKeys: string[] = []): ServiceStatus {
  if (!fields.length || !fields.some(hasConfigValue)) return { label: 'Not configured', tone: 'neutral' }
  const complete = requiredKeys.length > 0
    ? requiredKeys.every((key) => fields.some((field) => configKey(field) === key && hasConfigValue(field)))
    : fields.every(hasConfigValue)
  return complete ? { label: 'Configured', tone: 'success' } : { label: 'Needs attention', tone: 'warning' }
}

function ConfigEditor({ config, index, isRevealed, isWorking, onChange, onSave, onRequestReveal }: {
  config: AppConfig
  index: number
  isRevealed: boolean
  isWorking: boolean
  onChange: (key: string, value: string) => void
  onSave: (key: string, value: string) => void
  onRequestReveal: () => void
}) {
  const key = configKey(config, index)
  const current = stringValue(config.value)
  const secret = isSecretConfig(config)
  const masked = secret && !isRevealed

  return <form className="settings-row settings-service-row" onSubmit={(event) => { event.preventDefault(); if (masked) { onRequestReveal(); return } onSave(key, current) }}><label className="form-field"><span>{configLabel(config, index)}</span><input type={masked ? 'password' : 'text'} value={current} readOnly={masked} onChange={(event) => onChange(key, event.target.value)} aria-describedby={masked ? `${key}-masked-note` : undefined} /></label><button className="icon-button" type="submit" aria-label={masked ? `Reveal ${configLabel(config, index)}` : `Save ${configLabel(config, index)}`} disabled={isWorking || masked}><Save size={16} aria-hidden="true" /></button>{masked && <span className="settings-secret-note" id={`${key}-masked-note`}>Reveal to edit</span>}</form>
}

function ServiceCard({ definition, fields, expanded, isRevealed, isWorking, onToggle, onChange, onSave, onRequestReveal }: {
  definition: ServiceDefinition
  fields: AppConfig[]
  expanded: boolean
  isRevealed: boolean
  isWorking: boolean
  onToggle: () => void
  onChange: (key: string, value: string) => void
  onSave: (key: string, value: string) => void
  onRequestReveal: () => void
}) {
  const Icon = definition.icon
  const status = getServiceStatus(fields, definition.requiredKeys)
  const configuredCount = fields.filter(hasConfigValue).length
  const hasSecrets = fields.some(isSecretConfig)

  return <section className={`settings-service-card ${expanded ? 'settings-service-card-expanded' : ''}`}><div className="settings-service-card-header"><div className="settings-service-identity"><span className="settings-service-icon"><Icon size={18} aria-hidden="true" /></span><div><div className="settings-service-title-row"><h3>{definition.title}</h3><span className={`settings-service-status settings-service-status-${status.tone}`}>{status.label}</span></div><p>{definition.description}</p><small>{configuredCount} of {fields.length} settings configured</small></div></div><button className="secondary-button settings-service-toggle" type="button" onClick={onToggle}>{expanded ? 'Hide settings' : 'Configure'} <ChevronDown size={14} className={expanded ? 'settings-chevron-open' : undefined} aria-hidden="true" /></button></div>{expanded && <div className="settings-service-editor">{hasSecrets && !isRevealed && <div className="settings-service-note"><ShieldCheck size={15} aria-hidden="true" /> Secret values stay masked. Reveal them under Access &amp; security to edit.</div>}<div className="settings-service-fields">{fields.map((field, index) => <ConfigEditor key={configKey(field, index)} config={field} index={index} isRevealed={isRevealed} isWorking={isWorking} onChange={onChange} onSave={onSave} onRequestReveal={onRequestReveal} />)}</div></div>}</section>
}

function MachineKeysPanel({ token, onUnauthorized }: Props) {
  const [keys, setKeys] = useState<MachineKey[]>([])
  const [name, setName] = useState('')
  const [scopes, setScopes] = useState<string[]>(['orders:read'])
  const [rateLimit, setRateLimit] = useState('60')
  const [expiresAt, setExpiresAt] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [isWorking, setIsWorking] = useState(false)
  const [newPlaintext, setNewPlaintext] = useState('')
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')

  const loadKeys = useCallback(async () => {
    setIsLoading(true)
    try {
      const data = await apiJson<unknown>(token, onUnauthorized, '/api/mcp/keys')
      setKeys(arrayFrom(data, 'keys') as MachineKey[])
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to load machine keys')
    } finally {
      setIsLoading(false)
    }
  }, [onUnauthorized, token])

  useEffect(() => { void loadKeys() }, [loadKeys])

  const toggleScope = (scope: string) => setScopes((current) => current.includes(scope) ? current.filter((item) => item !== scope) : [...current, scope])
  const createKey = async (event: FormEvent) => { event.preventDefault(); if (!name.trim() || scopes.length === 0) { setError('Enter a name and choose at least one scope'); return }; setIsWorking(true); setError(''); try { const data = await apiJson<Row>(token, onUnauthorized, '/api/mcp/keys', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ name: name.trim(), scopes, rate_limit_per_min: Math.max(1, numberValue(rateLimit) || 60), ...(expiresAt ? { expires_at: `${expiresAt}T23:59:59Z` } : {}) }) }); setNewPlaintext(stringValue(data.plaintext)); setNotice('Machine key created. Save the plaintext now; it cannot be recovered later.'); setName(''); setScopes(['orders:read']); setRateLimit('60'); setExpiresAt(''); await loadKeys() } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to create machine key') } finally { setIsWorking(false) } }
  const revokeKey = async (key: MachineKey) => { if (!window.confirm(`Revoke ${textValue(key.name, 'this machine key')}? Connected clients stop working immediately.`)) return; setIsWorking(true); try { await apiRequest(token, onUnauthorized, `/api/mcp/keys/${key.id}`, { method: 'DELETE' }); setNotice('Machine key revoked'); await loadKeys() } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to revoke machine key') } finally { setIsWorking(false) } }
  const rotateKey = async (key: MachineKey) => { if (!window.confirm(`Rotate ${textValue(key.name, 'this machine key')}? The old key stops working immediately.`)) return; setIsWorking(true); try { const data = await apiJson<Row>(token, onUnauthorized, `/api/mcp/keys/${key.id}/rotate`, { method: 'POST' }); setNewPlaintext(stringValue(data.plaintext)); setNotice('Machine key rotated. Save the new plaintext now.'); await loadKeys() } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to rotate machine key') } finally { setIsWorking(false) } }
  const copyKey = async () => { if (!newPlaintext) return; try { await navigator.clipboard.writeText(newPlaintext); setNotice('Machine key copied') } catch { setError('Copy failed. Select the key manually.') } }

  return <section className="settings-card machine-keys-card"><div className="settings-card-heading"><div><p className="eyebrow">Developer access</p><h3><ShieldCheck size={17} aria-hidden="true" /> MCP machine keys</h3><p>Issue scoped keys for Codex or other MCP clients. Plaintext is shown once.</p></div><button className="icon-button" type="button" aria-label="Refresh machine keys" onClick={() => void loadKeys()} disabled={isLoading}><RefreshCw size={16} className={isLoading ? 'spin' : undefined} aria-hidden="true" /></button></div>{error && <div className="settings-inline-error" role="alert">{error}</div>}{notice && <div className="settings-inline-notice" role="status">{notice}</div>}{newPlaintext && <div className="machine-key-secret"><strong>Save this key now</strong><code>{newPlaintext}</code><div className="table-action-group"><button className="secondary-button" type="button" onClick={() => void copyKey()}><Copy size={14} aria-hidden="true" /> Copy</button><button className="table-link-button" type="button" onClick={() => setNewPlaintext('')}>Dismiss</button></div></div>}<form className="machine-key-form" onSubmit={createKey}><div className="form-grid-three"><label className="form-field"><span>Key name</span><input required value={name} onChange={(event) => setName(event.target.value)} placeholder="e.g. reporting client" /></label><label className="form-field"><span>Rate limit / min</span><input type="number" min="1" value={rateLimit} onChange={(event) => setRateLimit(event.target.value)} /></label><label className="form-field"><span>Expires on <small>(optional)</small></span><input type="date" value={expiresAt} onChange={(event) => setExpiresAt(event.target.value)} /></label></div><div className="machine-scope-picker"><div className="b2b-form-section-heading"><div><p className="eyebrow">Permissions</p><h4>{scopes.length} scopes selected</h4></div><button className="table-link-button" type="button" onClick={() => setScopes(['orders:read'])}>Reset</button></div><div className="machine-scope-list">{machineScopes.map(([scope, label]) => <label className="toggle-control" key={scope}><input type="checkbox" checked={scopes.includes(scope)} onChange={() => toggleScope(scope)} /><span>{label}</span><small>{scope}</small></label>)}</div></div><button className="primary-button" type="submit" disabled={isWorking}>{isWorking ? 'Creating…' : 'Generate machine key'}</button></form><div className="machine-key-list"><div className="b2b-form-section-heading"><div><p className="eyebrow">Issued keys</p><h4>{keys.length} keys</h4></div></div>{isLoading ? <p className="table-state">Loading machine keys…</p> : keys.length === 0 ? <p className="settings-muted">No machine keys have been issued.</p> : keys.map((key) => <div className="machine-key-row" key={String(key.id)}><div><strong>{textValue(key.name, 'Unnamed key')}</strong><small>{Array.isArray(key.scopes) ? key.scopes.join(', ') : 'No scopes'} · {textValue(key.revoked_at, '') ? 'Revoked' : textValue(key.expires_at, '') ? `Expires ${formatDate(key.expires_at)}` : 'No expiry'}</small></div><div className="table-action-group"><button className="table-link-button" type="button" onClick={() => void rotateKey(key)} disabled={isWorking}>Rotate</button>{!textValue(key.revoked_at, '') && <button className="table-link-button danger-link" type="button" onClick={() => void revokeKey(key)} disabled={isWorking}>Revoke</button>}</div></div>)}</div></section>
}

export function SettingsPage({ token, onUnauthorized }: Props) {
  const [settings, setSettings] = useState<Row>({})
  const [configs, setConfigs] = useState<AppConfig[]>([])
  const [dateRange, setDateRange] = useState({ start_date: '', end_date: '' })
  const [activeTab, setActiveTab] = useState<'workspace' | 'services' | 'access'>('workspace')
  const [selectedService, setSelectedService] = useState<string | null>(null)
  const [serviceSearch, setServiceSearch] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [isWorking, setIsWorking] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [revealPassword, setRevealPassword] = useState('')
  const [isRevealed, setIsRevealed] = useState(false)

  const load = useCallback(async () => {
    setIsLoading(true)
    setError('')
    try {
      const [settingsData, configsData, rangeData] = await Promise.all([apiJson<Row>(token, onUnauthorized, '/api/settings'), apiJson<unknown>(token, onUnauthorized, '/api/configs'), apiJson<Row>(token, onUnauthorized, '/api/settings/date-range')])
      setSettings((settingsData.settings || {}) as Row)
      setConfigs(arrayFrom(configsData, 'configs') as AppConfig[])
      setDateRange({ start_date: stringValue(rangeData.start_date), end_date: stringValue(rangeData.end_date) })
      setIsRevealed(false)
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to load settings')
    } finally {
      setIsLoading(false)
    }
  }, [onUnauthorized, token])

  useEffect(() => { void load() }, [load])

  const saveSetting = async (event: FormEvent, key: string) => { event.preventDefault(); setIsWorking(true); try { await apiRequest(token, onUnauthorized, '/api/settings', { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ key, value: stringValue(settings[key]) }) }); setNotice('Workspace setting saved') } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to save setting') } finally { setIsWorking(false) } }
  const saveConfig = async (key: string, value: string) => { setIsWorking(true); try { await apiRequest(token, onUnauthorized, '/api/configs', { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ key, value }) }); setNotice('Service setting saved') } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to save service setting') } finally { setIsWorking(false) } }
  const saveRange = async (event: FormEvent) => { event.preventDefault(); setIsWorking(true); try { await apiRequest(token, onUnauthorized, '/api/settings/date-range', { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(dateRange) }); setNotice('Default date range saved') } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to save date range') } finally { setIsWorking(false) } }
  const reveal = async (event: FormEvent) => { event.preventDefault(); setIsWorking(true); setError(''); try { const data = await apiJson<unknown>(token, onUnauthorized, '/api/configs/reveal', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ password: revealPassword }) }); setConfigs(arrayFrom(data, 'configs') as AppConfig[]); setIsRevealed(true); setNotice('Secret values revealed for this session') } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : 'Unable to reveal integration settings') } finally { setIsWorking(false) } }

  const serviceGroups = useMemo(() => serviceDefinitions.map((definition) => ({ definition, fields: configs.filter((config) => definition.categories.includes(stringValue(config.category))) })), [configs])
  const groupedKeys = useMemo(() => new Set(serviceGroups.flatMap(({ fields }) => fields.map((field, index) => configKey(field, index)))), [serviceGroups])
  const advancedFields = useMemo(() => configs.filter((config, index) => !groupedKeys.has(configKey(config, index))), [configs, groupedKeys])
  const searchTerm = serviceSearch.trim().toLowerCase()
  const matchesSearch = (definition: ServiceDefinition, fields: AppConfig[]) => !searchTerm || `${definition.title} ${definition.description} ${fields.map((field, index) => `${configLabel(field, index)} ${configKey(field, index)}`).join(' ')}`.toLowerCase().includes(searchTerm)
  const visibleServiceGroups = serviceGroups.filter(({ definition, fields }) => fields.length > 0 && matchesSearch(definition, fields))
  const showAdvanced = advancedFields.length > 0 && matchesSearch(advancedDefinition, advancedFields)
  const settingEntries = Object.entries(settings)

  const changeConfig = (key: string, value: string) => setConfigs((current) => current.map((config, index) => configKey(config, index) === key ? { ...config, value } : config))
  const requestReveal = () => { setActiveTab('access'); setNotice('Enter your password to reveal service secrets for this session.') }

  return <section className="workspace-page settings-page" aria-labelledby="settings-heading"><header className="workspace-page-header"><div><p className="eyebrow">Workspace / Settings</p><h2 id="settings-heading">Settings</h2><p>Manage workspace defaults, connected services, and access.</p></div><button className="secondary-button" type="button" onClick={() => void load()} disabled={isLoading}><RefreshCw size={15} className={isLoading ? 'spin' : undefined} aria-hidden="true" /> Refresh</button></header>{error && <div className="dashboard-error" role="alert"><CircleAlert size={18} aria-hidden="true" /><span>{error}</span><button type="button" onClick={() => void load()}>Try again</button></div>}{notice && <div className="inventory-notice" role="status">{notice}</div>}<div className="settings-tabs" role="tablist" aria-label="Settings sections"><button className={`filter-chip ${activeTab === 'workspace' ? 'filter-chip-active' : ''}`} type="button" role="tab" aria-selected={activeTab === 'workspace'} onClick={() => setActiveTab('workspace')}><Settings2 size={14} aria-hidden="true" /> Workspace defaults</button><button className={`filter-chip ${activeTab === 'services' ? 'filter-chip-active' : ''}`} type="button" role="tab" aria-selected={activeTab === 'services'} onClick={() => setActiveTab('services')}><Store size={14} aria-hidden="true" /> Connected services</button><button className={`filter-chip ${activeTab === 'access' ? 'filter-chip-active' : ''}`} type="button" role="tab" aria-selected={activeTab === 'access'} onClick={() => setActiveTab('access')}><KeyRound size={14} aria-hidden="true" /> Access &amp; security</button></div>{activeTab === 'workspace' ? <div className="settings-grid settings-workspace-grid"><section className="settings-card"><div className="settings-card-heading"><div><p className="eyebrow">Saved preferences</p><h3>Workspace defaults</h3><p>Defaults used across reports and operational screens.</p></div></div>{isLoading ? <p className="table-state">Loading settings…</p> : settingEntries.length === 0 ? <p className="settings-muted">No editable settings returned by the API.</p> : settingEntries.map(([key, rawValue]) => <form className="settings-row" key={key} onSubmit={(event) => void saveSetting(event, key)}><label className="form-field"><span>{key.replace(/_/g, ' ')}</span><input value={stringValue(settings[key], stringValue(rawValue))} onChange={(event) => setSettings({ ...settings, [key]: event.target.value })} /></label><button className="icon-button" type="submit" aria-label={`Save ${key}`} disabled={isWorking}><Save size={16} aria-hidden="true" /></button></form>)}</section><section className="settings-card"><div className="settings-card-heading"><div><p className="eyebrow">Reporting</p><h3>Default reporting period</h3><p>Set the period used when a screen has no local date selection.</p></div></div><form onSubmit={saveRange}><div className="form-grid-two"><label className="form-field"><span>Start date</span><input required type="date" value={dateRange.start_date} onChange={(event) => setDateRange({ ...dateRange, start_date: event.target.value })} /></label><label className="form-field"><span>End date</span><input required type="date" value={dateRange.end_date} onChange={(event) => setDateRange({ ...dateRange, end_date: event.target.value })} /></label></div><button className="primary-button" type="submit" disabled={isWorking}><Save size={14} aria-hidden="true" /> Save date range</button></form></section></div> : activeTab === 'services' ? <section className="settings-services-view"><div className="settings-services-toolbar"><div><p className="eyebrow">Connections</p><h3>Connected services</h3><p>Configure external systems without browsing raw configuration keys.</p></div><label className="orders-search settings-service-search"><span className="sr-only">Search services</span><input value={serviceSearch} onChange={(event) => setServiceSearch(event.target.value)} placeholder="Search services or settings" /></label></div>{visibleServiceGroups.length === 0 && !showAdvanced ? <div className="empty-panel"><Wrench size={20} aria-hidden="true" /><div><h2>No services found</h2><p>Try a different service or setting name.</p></div></div> : <div className="settings-services-grid">{visibleServiceGroups.map(({ definition, fields }) => <ServiceCard key={definition.id} definition={definition} fields={fields} expanded={selectedService === definition.id} isRevealed={isRevealed} isWorking={isWorking} onToggle={() => setSelectedService(selectedService === definition.id ? null : definition.id)} onChange={changeConfig} onSave={(key, value) => void saveConfig(key, value)} onRequestReveal={requestReveal} />)}{showAdvanced && <ServiceCard definition={advancedDefinition} fields={advancedFields} expanded={selectedService === advancedDefinition.id} isRevealed={isRevealed} isWorking={isWorking} onToggle={() => setSelectedService(selectedService === advancedDefinition.id ? null : advancedDefinition.id)} onChange={changeConfig} onSave={(key, value) => void saveConfig(key, value)} onRequestReveal={requestReveal} />}</div>}</section> : <><div className="settings-grid settings-access-grid"><section className="settings-card"><div className="settings-card-heading"><div><p className="eyebrow">Protected values</p><h3>Reveal service secrets</h3><p>Reveal values only when you need to verify or update a connection.</p></div></div><form onSubmit={reveal} className="reveal-form"><label className="form-field"><span>Admin password</span><div className="password-field"><input required type="password" value={revealPassword} onChange={(event) => setRevealPassword(event.target.value)} /><button className="icon-button" type="submit" aria-label="Reveal service secrets" disabled={isWorking}>{isRevealed ? <EyeOff size={16} aria-hidden="true" /> : <Eye size={16} aria-hidden="true" />}</button></div></label><button className="secondary-button" type="submit" disabled={isWorking}><Eye size={14} aria-hidden="true" /> {isRevealed ? 'Refresh revealed values' : 'Reveal for this session'}</button></form></section><section className="settings-card settings-security-card"><div className="settings-card-heading"><div><p className="eyebrow">Security policy</p><h3>Keep credentials protected</h3></div></div><ul className="settings-policy-list"><li>Secret values remain masked until explicitly revealed.</li><li>Revealed values are available only in this session.</li><li>Machine keys should use the smallest required scope.</li></ul></section></div><MachineKeysPanel token={token} onUnauthorized={onUnauthorized} /></>}</section>
}
