import { FormEvent, useCallback, useEffect, useMemo, useState } from 'react'
import { Building2, ChevronDown, CircleAlert, CloudUpload, Copy, CreditCard, Eye, EyeOff, KeyRound, MessageCircle, Megaphone, RefreshCw, Save, Settings2, ShieldCheck, ShoppingBag, Store, Workflow, Wrench, type LucideIcon } from 'lucide-react'
import { apiJson, apiRequest, arrayFrom, formatDate, numberValue, textValue } from '../../lib/http'

type Props = { token: string; onUnauthorized: () => void }
type Row = Record<string, unknown>
type MachineKey = Row & { scopes?: string[]; permission_role?: string }
type AppConfig = Row & { key?: string; value?: string; is_secret?: boolean; label?: string; category?: string }
type CustomerScope = readonly [string, string]
type PermissionRole = 'read_only' | 'full_access' | 'custom'
type ServiceStatus = { label: 'Configured' | 'Needs attention' | 'Not configured'; tone: 'success' | 'warning' | 'neutral' }
type ServiceDefinition = { id: string; title: string; description: string; categories: string[]; requiredKeys?: string[]; isConfigured?: (fields: AppConfig[]) => boolean; icon: LucideIcon }

const machineScopeGroups: { id: string; label: string; description: string; scopes: CustomerScope[] }[] = [
  {
    id: 'read',
    label: 'Read access',
    description: 'View reports and operational data.',
    scopes: [
      ['orders:read', 'Orders'], ['customers:read', 'Customers'], ['metrics:read', 'Metrics'], ['gst:read', 'GST reports'],
      ['inventory:read', 'Inventory'], ['production:read', 'Production'], ['b2b:read', 'B2B billing'], ['communication:read', 'WhatsApp'],
      ['marketing:read', 'Marketing'], ['feedback:read', 'Feedback'], ['abandoned_checkout:read', 'Abandoned checkouts'],
      ['support:read', 'Support'], ['settings:read', 'Settings'], ['system:read', 'System'],
    ],
  },
  {
    id: 'write',
    label: 'Write access',
    description: 'Create or update records and run syncs.',
    scopes: [
      ['orders:write', 'Orders write'], ['customers:write', 'Customers write'], ['inventory:write', 'Inventory write'], ['production:write', 'Production write'],
      ['b2b:write', 'B2B write'], ['communication:write', 'WhatsApp write'], ['marketing:write', 'Marketing write'],
      ['feedback:write', 'Feedback write'], ['support:write', 'Support write'], ['settings:write', 'Settings write'],
    ],
  },
  {
    id: 'destructive',
    label: 'Destructive access',
    description: 'Delete, reset, or cancel data.',
    scopes: [
      ['orders:destructive', 'Orders delete'], ['customers:destructive', 'Customers delete'], ['inventory:destructive', 'Inventory delete'], ['production:destructive', 'Production delete'],
      ['b2b:destructive', 'B2B delete'], ['communication:destructive', 'WhatsApp delete'],
    ],
  },
]

const machineScopes: CustomerScope[] = machineScopeGroups.flatMap((group) => group.scopes)
const allScopeIds = machineScopes.map(([scope]) => scope)
const readOnlyScopeIds = machineScopeGroups.find((group) => group.id === 'read')?.scopes.map(([scope]) => scope) ?? []

const permissionRoleMeta: Record<PermissionRole, { label: string; description: string }> = {
  read_only: { label: 'Read only', description: 'View data without making changes.' },
  full_access: { label: 'Full access', description: 'Use every MCP tool, including destructive actions.' },
  custom: { label: 'Custom', description: 'Choose the exact scopes this client needs.' },
}

function permissionRoleForKey(key: MachineKey): PermissionRole {
  if (key.permission_role === 'full_access' || (Array.isArray(key.scopes) && allScopeIds.every((scope) => key.scopes?.includes(scope)))) return 'full_access'
  if (key.permission_role === 'read_only' || (Array.isArray(key.scopes) && key.scopes.length === readOnlyScopeIds.length && readOnlyScopeIds.every((scope) => key.scopes?.includes(scope)))) return 'read_only'
  return 'custom'
}

const serviceDefinitions: ServiceDefinition[] = [
  { id: 'shopify', title: 'Shopify', description: 'Orders, inventory, and customer synchronization.', categories: ['shopify'], requiredKeys: ['shopify_store_url', 'shopify_access_token'], icon: Store },
  { id: 'amazon', title: 'Amazon', description: 'Marketplace orders and seller account synchronization.', categories: ['amazon'], requiredKeys: ['amazon_lwa_client_id', 'amazon_lwa_client_secret', 'amazon_lwa_refresh_token'], icon: ShoppingBag },
  { id: 'whatsapp', title: 'WhatsApp', description: 'Customer messaging, invoices, and automation.', categories: ['whatsapp'], requiredKeys: ['whatsapp_phone_number_id', 'whatsapp_waba_id'], icon: MessageCircle },
  { id: 'meta', title: 'Meta services', description: 'Shared Meta credentials and paid marketing.', categories: ['meta_shared', 'marketing'], requiredKeys: ['meta_app_id', 'meta_system_user_token'], icon: Megaphone },
  {
    id: 'smm_queue',
    title: 'SMM Queue',
    description: 'Upload carousel media and post copy to the private Azure queue.',
    categories: ['smm_queue'],
    isConfigured: (fields) => {
      const valueFor = (key: string) => fields.find((field) => configKey(field) === key)?.value?.trim() ?? ''
      const accountName = valueFor('azure_storage_account_name')
      const connectionString = valueFor('azure_storage_connection_string')
      const sasToken = valueFor('azure_storage_sas_token')
      const container = valueFor('smm_queue_container')
      return container !== '' && (connectionString !== '' || (accountName !== '' && sasToken !== ''))
    },
    icon: CloudUpload,
  },
  { id: 'payments', title: 'Payments', description: 'Payment collection and webhook configuration.', categories: ['payment'], requiredKeys: ['razorpay_key_id', 'razorpay_key_secret'], icon: CreditCard },
  { id: 'automation', title: 'Feedback & automation', description: 'Feedback links, recovery, and scheduled automation.', categories: ['feedback', 'abandoned_cart'], icon: Workflow },
  { id: 'business', title: 'Business profile', description: 'Business identity, tax, and billing details.', categories: ['business', 'b2b'], icon: Building2 },
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

function getServiceStatus(fields: AppConfig[], requiredKeys: string[] = [], isConfigured?: (fields: AppConfig[]) => boolean): ServiceStatus {
  if (!fields.length || !fields.some(hasConfigValue)) return { label: 'Not configured', tone: 'neutral' }
  if (isConfigured) return isConfigured(fields) ? { label: 'Configured', tone: 'success' } : { label: 'Needs attention', tone: 'warning' }
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
  const status = getServiceStatus(fields, definition.requiredKeys, definition.isConfigured)
  const configuredCount = fields.filter(hasConfigValue).length
  const hasSecrets = fields.some(isSecretConfig)

  return <section className={`settings-service-card ${expanded ? 'settings-service-card-expanded' : ''}`}><div className="settings-service-card-header"><div className="settings-service-identity"><span className="settings-service-icon"><Icon size={18} aria-hidden="true" /></span><div><div className="settings-service-title-row"><h3>{definition.title}</h3><span className={`settings-service-status settings-service-status-${status.tone}`}>{status.label}</span></div><p>{definition.description}</p><small>{configuredCount} of {fields.length} settings configured</small></div></div><button className="secondary-button settings-service-toggle" type="button" onClick={onToggle}>{expanded ? 'Hide settings' : 'Configure'} <ChevronDown size={14} className={expanded ? 'settings-chevron-open' : undefined} aria-hidden="true" /></button></div>{expanded && <div className="settings-service-editor">{hasSecrets && !isRevealed && <div className="settings-service-note"><ShieldCheck size={15} aria-hidden="true" /> Secret values stay masked. Reveal them under Access &amp; security to edit.</div>}<div className="settings-service-fields">{fields.map((field, index) => <ConfigEditor key={configKey(field, index)} config={field} index={index} isRevealed={isRevealed} isWorking={isWorking} onChange={onChange} onSave={onSave} onRequestReveal={onRequestReveal} />)}</div></div>}</section>
}

function MachineKeysPanel({ token, onUnauthorized }: Props) {
  const [keys, setKeys] = useState<MachineKey[]>([])
  const [name, setName] = useState('')
  const [permissionRole, setPermissionRole] = useState<PermissionRole>('read_only')
  const [scopes, setScopes] = useState<string[]>(readOnlyScopeIds)
  const [showScopeDetails, setShowScopeDetails] = useState(false)
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

  const selectPermissionRole = (role: PermissionRole) => {
    setPermissionRole(role)
    setShowScopeDetails(role === 'custom')
    if (role === 'full_access') setScopes(allScopeIds)
    if (role === 'read_only') setScopes(readOnlyScopeIds)
  }
  const toggleScope = (scope: st