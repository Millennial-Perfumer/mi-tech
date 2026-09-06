import { useCallback, useMemo, useState, type PropsWithChildren } from 'react'
import { getDefaultStartDate, getTodayIST } from './api'
import { PeriodFilterContext } from './periodFilterContext'

export function PeriodFilterProvider({ children }: PropsWithChildren) {
  const [startDate, setStartDate] = useState(() => localStorage.getItem('socialSmmStartDate') || getDefaultStartDate())
  const [endDate, setEndDate] = useState(() => localStorage.getItem('socialSmmEndDate') || getTodayIST())

  const setDateRange = useCallback((nextStartDate: string, nextEndDate: string) => {
    setStartDate(nextStartDate)
    setEndDate(nextEndDate)
    if (nextStartDate) localStorage.setItem('socialSmmStartDate', nextStartDate)
    else localStorage.removeItem('socialSmmStartDate')
    if (nextEndDate) localStorage.setItem('socialSmmEndDate', nextEndDate)
    else localStorage.removeItem('socialSmmEndDate')
  }, [])

  const value = useMemo(() => ({ startDate, endDate, setDateRange }), [endDate, setDateRange, startDate])

  return <PeriodFilterContext.Provider value={value}>{children}</PeriodFilterContext.Provider>
}
