import { API_BASE } from './api';
import { useState, useEffect, useRef } from 'react';
import { createPortal } from 'react-dom';
import { useToast } from './ToastContext';
import { useConfirm } from './ConfirmContext';

// Animation for collapsible sections
const SLIDE_IN_ANIMATION = `
  @keyframes slideIn {
    from { opacity: 0; transform: translateY(-8px); }
    to { opacity: 1; transform: translateY(0); }
  }
`;


interface AppConfig {
  key: string;
  value: string;
  is_secret: boolean;
  label: string;
  category: string;
  sort_order: number;
}

interface SettingsTabProps {
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
  userRole: string;
}

interface MachineApiKey {
  id: number;
  name: string;
  scopes: string[];
  rate_limit_per_min: number;
  expires_at: string | null;
  revoked_at: string | null;
  created_at: string;
}

const MCP_READ_SCOPES = [
  ['orders:read', 'Orders'],
  ['customers:read', 'Customers'],
  ['metrics:read', 'Dashboard metrics'],
  ['gst:read', 'GST reports'],
  ['inventory:read', 'Inventory'],
  ['production:read', 'Production'],
  ['b2b:read', 'B2B billing'],
  ['communication:read', 'WhatsApp automation'],
  ['marketing:read', 'Marketing'],
  ['feedback:read', 'Feedback'],
  ['abandoned_checkout:read', 'Abandoned checkouts'],
  ['planner:read', 'Planner'],
  ['support:read', 'Support tickets'],
  ['ai:read', 'AI conversations'],
  ['settings:read', 'Settings'],
  ['system:read', 'System docs']
] as const;

// Write access is intentionally opt-in. New keys still default to orders:read,
// while operators who need the newly deployed mutation tools can grant only
// the specific write/destructive scopes required by their client.
const MCP_WRITE_SCOPES = [
  ['marketing:publish', 'Publish to social queue'],
  ['orders:write', 'Orders write'],
  ['customers:write', 'Customers write'],
  ['inventory:write', 'Inventory write'],
  ['production:write', 'Production write'],
  ['planner:write', 'Planner write'],
  ['b2b:write', 'B2B write'],
  ['communication:write', 'WhatsApp write'],
  ['marketing:write', 'Marketing write'],
  ['feedback:write', 'Feedback write'],
  ['support:write', 'Support write'],
  ['settings:write', 'Settings write'],
  ['ai:write', 'AI write']
] as const;

const MCP_DESTRUCTIVE_SCOPES = [
  ['orders:destructive', 'Orders destructive'],
  ['customers:destructive', 'Customers destructive'],
  ['inventory:destructive', 'Inventory destructive'],
  ['production:destructive', 'Production destructive'],
  ['planner:destructive', 'Planner destructive'],
  ['b2b:destructive', 'B2B destructive'],
  ['communication:destructive', 'WhatsApp destructive'],
  ['ai:destructive', 'AI destructive']
] as const;

const MCP_SCOPES = [...MCP_READ_SCOPES, ...MCP_WRITE_SCOPES, ...MCP_DESTRUCTIVE_SCOPES] as const;

function MachineKeyPanel({ fetchWithAuth }: { fetchWithAuth: SettingsTabProps['fetchWithAuth'] }) {
  const { success: toastSuccess, error: toastError } = useToast();
  const { confirm } = useConfirm();
  const [keys, setKeys] = useState<MachineApiKey[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [name, setName] = useState('');
  const [scopes, setScopes] = useState<string[]>(['orders:read']);
  const [rateLimit, setRateLimit] = useState('60');
  const [expiresAt, setExpiresAt] = useState('');
  const [noExpiry, setNoExpiry] = useState(true);
  const [isCreating, setIsCreating] = useState(false);
  const [newPlaintext, setNewPlaintext] = useState<string | null>(null);

  const loadKeys = async () => {
    setIsLoading(true);
    try {
      const response = await fetchWithAuth(`${API_BASE}/api/mcp/keys`);
      const data = await response.json();
      if (!response.ok || !data.success) throw new Error(data.message || 'Failed to load machine keys');
      setKeys(data.keys || []);
    } catch (err) {
      console.error('Failed to load MCP keys:', err);
      toastError('Unable to load MCP keys. Please check your admin access.');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadKeys();
  }, []);

  const toggleScope = (scope: string) => {
    setScopes(current => current.includes(scope) ? current.filter(item => item !== scope) : [...current, scope]);
  };

  const handleCreate = async () => {
    if (!name.trim() || scopes.length === 0) {
      toastError('Enter a key name and select at least one scope.');
      return;
    }
    setIsCreating(true);
    try {
      const response = await fetchWithAuth(`${API_BASE}/api/mcp/keys`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: name.trim(),
          scopes,
          rate_limit_per_min: Math.max(1, Number(rateLimit) || 60),
          ...(!noExpiry && expiresAt ? { expires_at: `${expiresAt}T23:59:59Z` } : {})
        })
      });
      const data = await response.json();
      if (!response.ok || !data.success) throw new Error(data.message || 'Failed to create machine key');
      setNewPlaintext(data.plaintext);
      setName('');
      setScopes(['orders:read']);
      setRateLimit('60');
      setExpiresAt('');
      setNoExpiry(true);
      await loadKeys();
      toastSuccess('Machine key created. Copy it now; it cannot be recovered later.');
    } catch (err) {
      toastError(err instanceof Error ? err.message : 'Failed to create machine key');
    } finally {
      setIsCreating(false);
    }
  };

  const handleRevoke = async (key: MachineApiKey) => {
    if (!(await confirm({ title: 'Revoke machine key', message: `Revoke “${key.name}”? Clients using it will stop working immediately.`, variant: 'danger', confirmLabel: 'Revoke key' }))) return;
    try {
      const response = await fetchWithAuth(`${API_BASE}/api/mcp/keys/${key.id}`, { method: 'DELETE' });
      if (!response.ok) throw new Error('Failed to revoke machine key');
      toastSuccess('Machine key revoked');
      await loadKeys();
    } catch (err) {
      toastError(err instanceof Error ? err.message : 'Failed to revoke machine key');
    }
  };

  const handleRotate = async (key: MachineApiKey) => {
    if (!(await confirm({ title: 'Rotate machine key', message: `Rotate “${key.name}”? The current key will stop working immediately.`, variant: 'danger', confirmLabel: 'Rotate key' }))) return;
    try {
      const response = await fetchWithAuth(`${API_BASE}/api/mcp/keys/${key.id}/rotate`, { method: 'POST' });
      const data = await response.json();
      if (!response.ok || !data.success) throw new Error(data.message || 'Failed to rotate machine key');
      setNewPlaintext(data.plaintext);
      toastSuccess('Machine key rotated. Copy the new key now.');
      await loadKeys();
    } catch (err) {
      toastError(err instanceof Error ? err.message : 'Failed to rotate machine key');
    }
  };

  const copyKey = async () => {
    if (!newPlaintext) return;
    try {
      await navigator.clipboard.writeText(newPlaintext);
      toastSuccess('Machine key copied to clipboard');
    } catch {
      toastError('Copy failed. Select and copy the key manually.');
    }
  };

  const renderScopeButtons = (scopeList: readonly (readonly [string, string])[], activeColor: string) => scopeList.map(([scope, label]) => (
    <button key={scope} type="button" onClick={() => toggleScope(scope)} aria-pressed={scopes.includes(scope)} style={{ padding: '0.4rem 0.65rem', borderRadius: '999px', border: `1px solid ${scopes.includes(scope) ? activeColor : 'var(--border-color)'}`, background: scopes.includes(scope) ? 'var(--accent-subtle)' : 'transparent', color: scopes.includes(scope) ? activeColor : 'var(--text-secondary)', cursor: 'pointer', fontSize: '0.75rem' }}>{label}</button>
  ));

  return (
    <section className="card" style={{ padding: '2rem', marginTop: '1.5rem', overflow: 'hidden' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: '1rem', marginBottom: '1.5rem' }}>
        <div>
          <h2 style={{ fontSize: '1.25rem', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '0.7rem', marginBottom: '0.5rem' }}>
            <span style={{ width: '30px', height: '30px', display: 'inline-flex', alignItems: 'center', justifyContent: 'center', borderRadius: '9px', background: 'var(--accent-subtle)', color: 'var(--accent-color)' }}>
              <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M12 2v20M2 12h20" /><circle cx="12" cy="12" r="8" /></svg>
            </span>
            MCP Machine Keys
            <span style={{ fontSize: '0.62rem', letterSpacing: '0.08em', color: 'var(--status-active)', background: 'var(--status-active-bg)', border: '1px solid var(--status-active)', borderRadius: '999px', padding: '0.25rem 0.5rem' }}>READ-ONLY</span>
          </h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', margin: 0 }}>
            Create read-only keys for Codex, Claude, or other MCP clients. Keys are shown only once.
          </p>
        </div>
        <button className="toolbar-btn" title="Refresh machine keys" onClick={loadKeys} disabled={isLoading} style={{ color: 'var(--text-tertiary)' }}>
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M20 11a8.1 8.1 0 0 0-15.5-2M4 5v4h4" /><path d="M4 13a8.1 8.1 0 0 0 15.5 2M20 19v-4h-4" /></svg>
        </button>
      </div>

      {newPlaintext && (
        <div style={{ padding: '1rem', marginBottom: '1.5rem', borderRadius: '10px', border: '1px solid var(--status-warning)', background: 'var(--status-warning-bg)' }}>
          <div style={{ fontWeight: 700, color: 'var(--status-warning)', marginBottom: '0.4rem' }}>Save this machine key now</div>
          <div style={{ fontSize: '0.78rem', color: 'var(--text-secondary)', marginBottom: '0.75rem' }}>It will not be displayed again after you dismiss this message.</div>
          <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
            <code style={{ flex: 1, padding: '0.65rem 0.75rem', borderRadius: '6px', background: 'var(--bg-input)', overflow: 'auto', whiteSpace: 'nowrap' }}>{newPlaintext}</code>
            <button className="btn-primary" onClick={copyKey} style={{ padding: '0.6rem 0.9rem' }}>Copy</button>
            <button className="btn-secondary" onClick={() => setNewPlaintext(null)} style={{ padding: '0.6rem 0.9rem' }}>Dismiss</button>
          </div>
        </div>
      )}

      <div style={{ padding: '1rem', marginBottom: '1.25rem', borderRadius: '14px', border: '1px solid var(--border-color)', background: 'linear-gradient(135deg, var(--bg-input), var(--surface-color))' }}>
        <div style={{ fontSize: '0.72rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: '0.75rem' }}>Create access key</div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '0.75rem', alignItems: 'end' }}>
          <label style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem', fontSize: '0.72rem', fontWeight: 600, color: 'var(--text-secondary)' }}>
            Key name
            <input value={name} onChange={e => setName(e.target.value)} placeholder="e.g. Claude production" aria-label="Machine key name" style={{ padding: '0.7rem 0.8rem', borderRadius: '8px', border: '1px solid var(--border-color)', background: 'var(--bg-input)', color: 'var(--text-primary)' }} />
          </label>
          <label style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem', fontSize: '0.72rem', fontWeight: 600, color: 'var(--text-secondary)' }}>
            Rate limit / min
            <input value={rateLimit} onChange={e => setRateLimit(e.target.value)} type="number" min="1" aria-label="Rate limit per minute" style={{ padding: '0.7rem 0.8rem', borderRadius: '8px', border: '1px solid var(--border-color)', background: 'var(--bg-input)', color: 'var(--text-primary)' }} />
          </label>
          <label style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem', fontSize: '0.72rem', fontWeight: 600, color: 'var(--text-secondary)' }}>
            Expiration
            <input value={expiresAt} onChange={e => { setExpiresAt(e.target.value); setNoExpiry(false); }} type="date" aria-label="Expiration date" disabled={noExpiry} style={{ width: '100%', padding: '0.7rem 0.8rem', borderRadius: '8px', border: '1px solid var(--border-color)', background: noExpiry ? 'var(--bg-hover)' : 'var(--bg-input)', color: 'var(--text-primary)', opacity: noExpiry ? 0.6 : 1 }} />
          </label>
          <button className="btn-primary" onClick={handleCreate} disabled={isCreating} style={{ padding: '0.7rem 1rem', whiteSpace: 'nowrap' }}>{isCreating ? 'Creating…' : 'Generate key'}</button>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.45rem', marginTop: '0.7rem' }}>
          <label style={{ display: 'flex', alignItems: 'center', gap: '0.45rem', fontSize: '0.75rem', color: 'var(--text-secondary)', cursor: 'pointer' }}>
            <input type="checkbox" checked={noExpiry} onChange={e => { setNoExpiry(e.target.checked); if (e.target.checked) setExpiresAt(''); }} />
            No expiry
          </label>
          <span style={{ fontSize: '0.72rem', color: 'var(--text-tertiary)' }}>· The key remains active until revoked.</span>
        </div>
      </div>

      <div style={{ padding: '0.25rem 0 1.35rem', borderBottom: '1px solid var(--border-color)', marginBottom: '1.25rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: '1rem', marginBottom: '0.75rem' }}>
