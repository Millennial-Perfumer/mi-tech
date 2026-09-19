import { useEffect, useRef } from 'react'
import { API_BASE } from './api'
import { apiJson } from './http'

type RealtimeEvent = { type: string; at: string }
type Listener = (event: RealtimeEvent) => void
const listeners = new Set<Listener>()

function publish(event: RealtimeEvent) {
  for (const listener of listeners) listener(event)
}

export function useRealtimeConnection(token: string | null, onUnauthorized: () => void) {
  const unauthorizedRef = useRef(onUnauthorized)
  unauthorizedRef.current = onUnauthorized

  useEffect(() => {
    if (!token) return
    let active = true
    let socket: WebSocket | null = null
    let retryTimer: number | undefined
    let attempt = 0
    let connecting = false

    const connect = async () => {
      if (!active || connecting || socket?.readyState === WebSocket.OPEN || socket?.readyState === WebSocket.CONNECTING) return
      connecting = true
      try {
        const { ticket } = await apiJson<{ ticket: string }>(token, () => unauthorizedRef.current(), '/api/realtime/ticket', { method: 'POST' })
        if (!active) return
        const base = new URL(API_BASE || window.location.origin, window.location.origin)
        base.protocol = base.protocol === 'https:' ? 'wss:' : 'ws:'
        base.pathname = '/api/realtime/ws'
        base.search = new URLSearchParams({ ticket }).toString()
        socket = new WebSocket(base.toString())
        socket.onopen = () => {
          attempt = 0
          publish({ type: 'realtime.connected', at: new Date().toISOString() })
        }
        socket.onmessage = (message) => {
          try {
            const event: unknown = JSON.parse(String(message.data))
            if (event && typeof event === 'object' && 'type' in event && typeof event.type === 'string') {
              publish(event as RealtimeEvent)
            }
          } catch {
            // Ignore malformed events; the next valid event can still refresh data.
          }
        }
        socket.onclose = () => {
          if (active) scheduleRetry()
        }
        socket.onerror = () => socket?.close()
      } catch {
        if (active) scheduleRetry()
      } finally {
        connecting = false
      }
    }

    const scheduleRetry = () => {
      if (retryTimer !== undefined) window.clearTimeout(retryTimer)
      retryTimer = window.setTimeout(() => void connect(), Math.min(30000, 1000 * 2 ** Math.min(attempt++, 5)))
    }
    const wake = () => {
      if (document.visibilityState === 'visible' && socket?.readyState !== WebSocket.OPEN) {
        if (retryTimer !== undefined) window.clearTimeout(retryTimer)
        void connect()
      }
    }
    document.addEventListener('visibilitychange', wake)
    window.addEventListener('online', wake)
    void connect()
    return () => {
      active = false
      if (retryTimer !== undefined) window.clearTimeout(retryTimer)
      document.removeEventListener('visibilitychange', wake)
      window.removeEventListener('online', wake)
      socket?.close()
    }
  }, [token])
}

export function useRealtimeRefresh(types: readonly string[], refresh: () => void) {
  const refreshRef = useRef(refresh)
  refreshRef.current = refresh
  const key = types.join('|')

  useEffect(() => {
    let timer: number | undefined
    const wanted = new Set(key.split('|'))
    const listener: Listener = (event) => {
      if (!wanted.has(event.type) && event.type !== 'realtime.connected') return
      if (timer !== undefined) window.clearTimeout(timer)
      timer = window.setTimeout(() => refreshRef.current(), 200)
    }
    listeners.add(listener)
    return () => {
      listeners.delete(listener)
      if (timer !== undefined) window.clearTimeout(timer)
    }
  }, [key])
}
