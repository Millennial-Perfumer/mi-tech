import { useContext } from 'react'
import { PeriodFilterContext } from './periodFilterContext'

export function usePeriodFilter() {
  const value = useContext(PeriodFilterContext)
  if (!value) throw new Error('usePeriodFilter must be used inside PeriodFilterProvider')
  return value
}
