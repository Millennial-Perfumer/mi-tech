import { API_BASE } from './api'

export async function apiRequest(token: string, onUnauthorized: () => void, path: string, options: RequestInit = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      Authorization: `Bearer ${token}`,
      ...(options.headers || {}),
    },
  })
  if (response.status === 401) {
    onUnauthorized()
    throw new Error('Your session has expired. Please sign in again.')
  }
  if (!response.ok) throw new Error(`Request failed with status ${response.status}`)
  return response
}

export async function apiJson<T>(token: string, onUnauthorized: () => void, path: string, options?: RequestInit) {
  const response = await apiRequest(token, onUnauthorized, path, options)
  return response.json() as Promise<T>
}

export function arrayFrom(value: unknown, key?: string) {
  if (Array.isArray(value)) return value as Record<string, unknown>[]
  if (key && value && typeof value === 'object') {
    const nested = (value as Record<string, unknown>)[key]
    if (Array.isArray(nested)) return nested as Record<string, unknown>[]
  }
  return [] as Record<string, unknown>[]
}

export function numberValue(value: unknown) {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : 0
}

export function textValue(value: unknown, fallback = '—') {
  return typeof value === 'string' && value ? value : fallback
}

export function formatDate(value: unknown) {
  if (!value) return '—'
  const date = new Date(String(value))
  return Number.isNaN(date.getTime()) ? '—' : date.toLocaleDateString('en-IN', { day: 'numeric', month: 'short', year: 'numeric' })
}

export function formatMoney(value: unknown) {
  return `₹${numberValue(value).toLocaleString('en-IN', { maximumFractionDigits: 0 })}`
}
