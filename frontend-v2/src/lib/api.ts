// Local development always uses the Vite proxy so browser requests stay same-origin.
// Production builds should provide VITE_API_URL for the deployed API host.
export const API_BASE = import.meta.env.DEV ? '' : (import.meta.env.VITE_API_URL || 'http://localhost:8080')

export function dateToBoundary(date: string, endOfDay = false) {
  const [year, month, day] = date.split('-').map(Number)
  const value = new Date(year, month - 1, day, endOfDay ? 23 : 0, endOfDay ? 59 : 0, endOfDay ? 59 : 0, endOfDay ? 999 : 0)
  return value.toISOString()
}

export function getTodayIST() {
  return new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Kolkata' })
}

export function getDefaultStartDate() {
  const now = new Date()
  return `${now.getFullYear()}-01-01`
}
