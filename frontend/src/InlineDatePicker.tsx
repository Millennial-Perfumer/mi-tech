import React, { useState, useEffect } from 'react';
import { DateRange } from 'react-date-range';
import type { RangeKeyDict } from 'react-date-range';
import {
  format,
  startOfMonth,
  endOfMonth,
  subMonths,
  subDays,
  startOfToday,
  endOfToday,
  isBefore,
  parseISO
} from 'date-fns';
import 'react-date-range/dist/styles.css';
import 'react-date-range/dist/theme/default.css';
import './CustomDatePicker.css';

interface InlineDatePickerProps {
  startDate: string;
  endDate: string;
  onChange: (start: string, end: string) => void;
  minDate?: string;
}

const PRESETS = [
  { label: 'Today', getValue: () => [startOfToday(), endOfToday()] },
  { label: 'Yesterday', getValue: () => [subDays(startOfToday(), 1), subDays(endOfToday(), 1)] },
  { label: 'Last 7 days', getValue: () => [subDays(startOfToday(), 6), endOfToday()] },
  { label: 'Last 30 days', getValue: () => [subDays(startOfToday(), 29), endOfToday()] },
  { label: 'Last 90 days', getValue: () => [subDays(startOfToday(), 89), endOfToday()] },
  { label: 'Month to date', getValue: () => [startOfMonth(startOfToday()), endOfToday()] },
  { label: 'Last month', getValue: () => [startOfMonth(subMonths(startOfToday(), 1)), endOfMonth(subMonths(startOfToday(), 1))] },
  { label: 'Last 12 months', getValue: () => [subMonths(startOfToday(), 12), endOfToday()] },
];

export const InlineDatePicker: React.FC<InlineDatePickerProps> = ({
  startDate,
  endDate,
  onChange,
  minDate = '2026-01-01'
}) => {
  const parsedStart = startDate ? parseISO(startDate) : startOfMonth(startOfToday());
  const parsedEnd = endDate ? parseISO(endDate) : endOfToday();
  const parsedMinDate = parseISO(minDate);

  const [activePreset, setActivePreset] = useState<string>('Custom');

  // Sync preset label if props match a preset
  useEffect(() => {
    console.log('InlineDatePicker: props updated', { startDate, endDate });
    const startStr = startDate;
    const endStr = endDate;
    
    // Find matching preset by comparing formatted strings
    const matchingPreset = PRESETS.find(p => {
      const [pStart, pEnd] = p.getValue();
      return format(pStart, 'yyyy-MM-dd') === startStr && format(pEnd, 'yyyy-MM-dd') === endStr;
    });

    setActivePreset(matchingPreset ? matchingPreset.label : 'Custom');
  }, [startDate, endDate]);

  const handlePresetClick = (preset: typeof PRESETS[0]) => {
    let [start, end] = preset.getValue();
    if (isBefore(start, parsedMinDate)) {
      start = parsedMinDate;
    }
    onChange(format(start, 'yyyy-MM-dd'), format(end, 'yyyy-MM-dd'));
  };

  const handleCalendarChange = (ranges: RangeKeyDict) => {
    const s = ranges.selection.startDate || new Date();
    const e = ranges.selection.endDate || new Date();
    onChange(format(s, 'yyyy-MM-dd'), format(e, 'yyyy-MM-dd'));
  };

  const selectionRange = {
    startDate: parsedStart,
    endDate: parsedEnd,
    key: 'selection',
  };

  return (
    <div className="datepicker-content" style={{ border: 'none', background: 'transparent', width: '100%', display: 'flex' }}>
      {/* Left Presets Sidebar */}
      <div className="datepicker-sidebar" style={{ background: 'transparent', width: '180px', flexShrink: 0 }}>
        {PRESETS.map((preset) => (
          <button
            key={preset.label}
            className={`preset-btn ${activePreset === preset.label ? 'active' : ''}`}
            onClick={() => handlePresetClick(preset)}
          >
            {preset.label}
          </button>
        ))}
      </div>

      {/* Right Calendar Section */}
      <div className="datepicker-calendar-section" style={{ padding: '0', flex: 1, display: 'flex', justifyContent: 'center', minWidth: '0' }}>
        <div className="calendar-wrapper inline-calendar-wrapper" style={{ minWidth: '0' }}>
          <DateRange
            ranges={[selectionRange]}
            onChange={handleCalendarChange}
            months={2}
            direction="horizontal"
            minDate={parsedMinDate}
            moveRangeOnFirstSelection={false}
            showMonthAndYearPickers={false}
            showPreview={false}
            rangeColors={['#10b981']}
          />
        </div>
      </div>
    </div>
  );
};
