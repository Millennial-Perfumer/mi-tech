import { useEffect, useRef, useState } from 'react'
import { CalendarDays, ChevronDown, ChevronLeft, ChevronRight } from 'lucide-react'
import { getTodayIST } from '../lib/api'

type DateRangePickerProps = {
  startDate: string
  endDate: string
  onChange: (startDate: string, endDate: string) => void
}

type DateField = 'start' | 'end'
type PeriodPresetId = 'today' | 'yesterday' | 'last-7-days' | 'last-30-days' | 'this-month' | 'last-month' | 'this-quarter' | 'this-financial-year' | 'all-time' | 'custom'
type DateRange = [string, string]

type PeriodPreset = {
  id: PeriodPresetId
  label: string
  description: string
}

const weekdays = ['S', 'M', 'T', 'W', 'T', 'F', 'S']

const periodPresets: PeriodPreset[] = [
  { id: 'today', label: 'Today', description: 'Just today' },
  { id: 'yesterday', label: 'Yesterday', description: 'The previous day' },
  { id: 'last-7-days', label: 'Last 7 days', description: 'Including today' },
  { id: 'last-30-days', label: 'Last 30 days', description: 'Including today' },
  { id: 'this-month', label: 'This month', description: 'From the 1st to today' },
  { id: 'last-month', label: 'Last month', description: 'The previous calendar month' },
  { id: 'this-quarter', label: 'This quarter', description: 'The current calendar quarter' },
  { id: 'this-financial-year', label: 'This financial year', description: 'April to March' },
  { id: 'all-time', label: 'All time', description: 'No date filter' },
  { id: 'custom', label: 'Custom range', description: 'Choose exact dates' },
]

function parseDate(value: string) {
  const [year, month, day] = value.split('-').map(Number)
  return new Date(year, month - 1, day)
}

function toDateValue(date: Date) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function addDays(date: Date, amount: number) {
  return toDateValue(new Date(date.getFullYear(), date.getMonth(), date.getDate() + amount))
}

function formatFieldDate(value: string) {
  if (!value) return 'Any date'
  const [year, month, day] = value.split('-')
  return `${day}/${month}/${year}`
}

function formatReadableDate(value: string) {
  if (!value) return 'No date filter'
  return parseDate(value).toLocaleDateString('en-IN', { day: 'numeric', month: 'short', year: 'numeric' })
}

function formatRange(startDate: string, endDate: string) {
  if (!startDate && !endDate) return 'No date filter'
  if (startDate === endDate) return formatReadableDate(startDate)
  return `${formatReadableDate(startDate)} – ${formatReadableDate(endDate)}`
}

function monthLabel(date: Date) {
  return date.toLocaleDateString('en-US', { month: 'long', year: 'numeric' })
}

function getPresetRange(preset: PeriodPresetId, todayValue: string): DateRange {
  const today = parseDate(todayValue)

  switch (preset) {
    case 'today':
      return [todayValue, todayValue]
    case 'yesterday': {
      const yesterday = addDays(today, -1)
      return [yesterday, yesterday]
    }
    case 'last-7-days':
      return [addDays(today, -6), todayValue]
    case 'last-30-days':
      return [addDays(today, -29), todayValue]
    case 'this-month':
      return [toDateValue(new Date(today.getFullYear(), today.getMonth(), 1)), todayValue]
    case 'last-month': {
      const lastMonth = new Date(today.getFullYear(), today.getMonth() - 1, 1)
      return [toDateValue(lastMonth), toDateValue(new Date(today.getFullYear(), today.getMonth(), 0))]
    }
    case 'this-quarter': {
      const quarterStartMonth = Math.floor(today.getMonth() / 3) * 3
      return [toDateValue(new Date(today.getFullYear(), quarterStartMonth, 1)), todayValue]
    }
    case 'this-financial-year': {
      const financialYearStart = today.getMonth() >= 3 ? today.getFullYear() : today.getFullYear() - 1
      return [toDateValue(new Date(financialYearStart, 3, 1)), todayValue]
    }
    case 'all-time':
      return ['', '']
    case 'custom':
      return ['', '']
  }
}

function getMatchingPreset(startDate: string, endDate: string): PeriodPresetId {
  const today = getTodayIST()
  const match = periodPresets.find((preset) => {
    if (preset.id === 'custom') return false
    const [presetStart, presetEnd] = getPresetRange(preset.id, today)
    return presetStart === startDate && presetEnd === endDate
  })
  return match?.id || 'custom'
}

function isBetween(value: string, start: string, end: string) {
  return Boolean(start && end && value > start && value < end)
}

function getCalendarDays(month: Date) {
  const firstDay = new Date(month.getFullYear(), month.getMonth(), 1)
  return Array.from({ length: 42 }, (_, index) => (
    new Date(month.getFullYear(), month.getMonth(), index - firstDay.getDay() + 1)
  ))
}

export function DateRangePicker({ startDate, endDate, onChange }: DateRangePickerProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [activePreset, setActivePreset] = useState<PeriodPresetId>(() => getMatchingPreset(startDate, endDate))
  const [activeField, setActiveField] = useState<DateField>('start')
  const [draftStart, setDraftStart] = useState(startDate)
  const [draftEnd, setDraftEnd] = useState(endDate)
  const [visibleMonth, setVisibleMonth] = useState(() => parseDate(startDate || endDate || getTodayIST()))
  const pickerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!isOpen) return

    const handleOutsidePointer = (event: PointerEvent) => {
      if (pickerRef.current && !pickerRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setIsOpen(false)
    }

    document.addEventListener('pointerdown', handleOutsidePointer)
    document.addEventListener('keydown', handleEscape)
    return () => {
      document.removeEventListener('pointerdown', handleOutsidePointer)
      document.removeEventListener('keydown', handleEscape)
    }
  }, [isOpen])

  const openPicker = () => {
    const currentPreset = getMatchingPreset(startDate, endDate)
    setActivePreset(currentPreset)
    setDraftStart(startDate)
    setDraftEnd(endDate)
    setVisibleMonth(parseDate(startDate || endDate || getTodayIST()))
    setActiveField('start')
    setIsOpen(true)
  }

  const selectPreset = (preset: PeriodPresetId) => {
    setActivePreset(preset)

    if (preset === 'custom') {
      setDraftStart(startDate)
      setDraftEnd(endDate)
      setActiveField('start')
      setVisibleMonth(parseDate(startDate || endDate || getTodayIST()))
      return
    }

    const [nextStartDate, nextEndDate] = getPresetRange(preset, getTodayIST())
    onChange(nextStartDate, nextEndDate)
    setIsOpen(false)
  }

  const selectDate = (date: Date) => {
    const value = toDateValue(date)
    setActivePreset('custom')
    setVisibleMonth(new Date(date.getFullYear(), date.getMonth(), 1))

    if (activeField === 'start') {
      setDraftStart(value)
      if (draftEnd && value > draftEnd) setDraftEnd('')
      setActiveField('end')
      return
    }

    if (!draftStart || value < draftStart) {
      setDraftStart(value)
      setDraftEnd('')
      setActiveField('end')
      return
    }

    setDraftEnd(value)
  }

  const applyCustomRange = () => {
    if (!draftStart || !draftEnd) return
    onChange(draftStart, draftEnd)
    setIsOpen(false)
  }

  const today = getTodayIST()
  const calendarMonths = [visibleMonth, new Date(visibleMonth.getFullYear(), visibleMonth.getMonth() + 1, 1)]
  const currentPreset = periodPresets.find((preset) => preset.id === getMatchingPreset(startDate, endDate)) || periodPresets[periodPresets.length - 1]
  const customRangeReady = Boolean(draftStart && draftEnd && draftStart <= draftEnd)

  return (
    <div className="date-range-picker" ref={pickerRef}>
      <button
        className="date-range-trigger"
        type="button"
        aria-haspopup="dialog"
        aria-expanded={isOpen}
        onClick={openPicker}
      >
        <CalendarDays size={16} aria-hidden="true" />
        <span className="date-range-trigger-copy">
          <strong>{currentPreset.label}</strong>
          <span>{formatRange(startDate, endDate)}</span>
        </span>
        <ChevronDown size={15} aria-hidden="true" />
      </button>

      {isOpen && (
        <div className="calendar-popover" role="dialog" aria-modal="true" aria-label="Choose reporting period">
          <div className="period-picker-heading">
            <div>
              <span className="period-picker-kicker">Filter by</span>
              <strong>Time period</strong>
            </div>
            <span className="period-picker-current">{activePreset === 'custom' ? formatRange(draftStart, draftEnd) : periodPresets.find((preset) => preset.id === activePreset)?.label}</span>
          </div>

          <div className="period-picker-layout">
            <nav className="period-presets" aria-label="Quick time periods">
              {periodPresets.map((preset) => (
                <button
                  className={`period-preset ${activePreset === preset.id ? 'period-preset-active' : ''}`}
                  type="button"
                  key={preset.id}
                  aria-pressed={activePreset === preset.id}
                  onClick={() => selectPreset(preset.id)}
                >
                  <span>{preset.label}</span>
                  <small>{preset.description}</small>
                </button>
              ))}
            </nav>

            {activePreset === 'custom' ? (
              <div className="custom-period-panel">
                <div className="calendar-selection">
                  <button
                    className={`calendar-selection-field ${activeField === 'start' ? 'calendar-selection-field-active' : ''}`}
                    type="button"
                    aria-pressed={activeField === 'start'}
                    onClick={() => setActiveField('start')}
                  >
                    <span>From</span>
                    <strong>{formatFieldDate(draftStart)}</strong>
                  </button>
                  <span className="calendar-selection-divider" aria-hidden="true">—</span>
                  <button
                    className={`calendar-selection-field ${activeField === 'end' ? 'calendar-selection-field-active' : ''}`}
                    type="button"
                    aria-pressed={activeField === 'end'}
                    onClick={() => setActiveField('end')}
                  >
                    <span>To</span>
                    <strong>{formatFieldDate(draftEnd)}</strong>
                  </button>
                </div>

                <div className="calendar-header">
                  <button className="calendar-nav-button" type="button" aria-label="Previous month" onClick={() => setVisibleMonth(new Date(visibleMonth.getFullYear(), visibleMonth.getMonth() - 1, 1))}>
                    <ChevronLeft size={17} aria-hidden="true" />
                  </button>
                  <div className="calendar-month-labels">
                    {calendarMonths.map((month) => <strong key={month.toISOString()}>{monthLabel(month)}</strong>)}
                  </div>
                  <button className="calendar-nav-button" type="button" aria-label="Next month" onClick={() => setVisibleMonth(new Date(visibleMonth.getFullYear(), visibleMonth.getMonth() + 1, 1))}>
                    <ChevronRight size={17} aria-hidden="true" />
                  </button>
                </div>

                <div className="calendar-months">
                  {calendarMonths.map((month) => (
                    <div className="calendar-month" key={month.toISOString()}>
                      <div className="calendar-weekdays" aria-hidden="true">
                        {weekdays.map((weekday, index) => <span key={`${weekday}-${index}`}>{weekday}</span>)}
                      </div>
                      <div className="calendar-grid" role="grid" aria-label={monthLabel(month)}>
                        {getCalendarDays(month).map((date) => {
                          const value = toDateValue(date)
                          const isCurrentMonth = date.getMonth() === month.getMonth()
                          const isStart = value === draftStart
                          const isEnd = value === draftEnd
                          const isInRange = isBetween(value, draftStart, draftEnd)
                          const isToday = value === today

                          return (
                            <div key={value} role="gridcell" aria-selected={isStart || isEnd}>
                              <button
                                className={`calendar-day ${!isCurrentMonth ? 'calendar-day-outside' : ''} ${isStart ? 'calendar-day-start' : ''} ${isEnd ? 'calendar-day-end' : ''} ${isInRange ? 'calendar-day-in-range' : ''}`}
                                type="button"
                                aria-label={date.toLocaleDateString('en-US', { dateStyle: 'full' })}
                                aria-current={isToday ? 'date' : undefined}
                                onClick={() => selectDate(date)}
                              >
                                {date.getDate()}
                              </button>
                            </div>
                          )
                        })}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ) : (
              <div className="period-preset-summary">
                <span className="period-preset-summary-icon"><CalendarDays size={18} aria-hidden="true" /></span>
                <strong>{periodPresets.find((preset) => preset.id === activePreset)?.label}</strong>
                <span>{formatRange(...getPresetRange(activePreset, today))}</span>
                <p>Choose a quick period, or select Custom range for exact dates.</p>
              </div>
            )}
          </div>

          {activePreset === 'custom' && (
            <div className="calendar-footer">
              <button className="calendar-cancel-button" type="button" onClick={() => setIsOpen(false)}>Cancel</button>
              <button className="calendar-apply-button" type="button" disabled={!customRangeReady} onClick={applyCustomRange}>Apply</button>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
