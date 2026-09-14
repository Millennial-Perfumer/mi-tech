import { createContext } from 'react'

export type PeriodFilterValue = {
  startDate: string
  endDate: string
  setDateRange: (startDate: string, endDate: string) => void
}

export const PeriodFilterContext = createContext<PeriodFilterValue | null>(null)
