import { useCallback, useMemo, useState, type PropsWithChildren } from 'react'
import { getDefaultStartDate, getTodayIST } from './api'
import { PeriodFilterContext } from './periodFilterContext'

export function PeriodFilterProvider({ children }: PropsWithChildren) {
  const [startDate, setStartDate] = useState(() => localStorage.getItem('periodFilterStartDate') || getDefaultStartDate())
  const [endDate, setEndDate] = useState(() => localStorage.getItem('periodFilterEndDate') || getTodayIST())

  const setDateRange = useCallback((nextStartDate: string, nextEndDate: string) => {
    setStartDate(nextStartDate)
    setEndDate(nextEndDate)
    if (nextStartDate) localStorage.setItem('periodFilterStartDate', nextStartDate)
    else localStorage.removeItem('periodFilterStartDate')
    if (nextEndDate) localStorage.setItem('periodFilterEndDate', nextEndDate)
    else localStorage.removeItem('periodFilterEndDate')
  }, [])

  const value = useMemo(() => ({ startDate, endDate, setDateRange }), [endDate, setDateRange, startDate])

  return <PeriodFilterContext.Provider value={value}>{children}</PeriodFilterContext.Provider>
}
