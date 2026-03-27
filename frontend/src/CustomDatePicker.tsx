import React, { useState, useEffect, useRef } from 'react';
import { createPortal } from 'react-dom';
import { DateRangePicker } from 'react-date-range';
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
  parseISO,
  isValid
} from 'date-fns';
import 'react-date-range/dist/styles.css'; // main style file
import 'react-date-range/dist/theme/default.css'; // theme css file
import './CustomDatePicker.css'; // Custom Shopify-like styling

interface CustomDatePickerProps {
  startDate: string;
  endDate: string;
  onDateChange: (start: string, end: string) => void;
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

export const CustomDatePicker: React.FC<CustomDatePickerProps> = ({
  startDate,
  endDate,
  onDateChange,
  minDate = '2026-01-01'
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Parse initial dates safely with isValid checks
  const safeParseISO = (dateStr: string | undefined, fallback: Date) => {
    if (!dateStr) return fallback;
    const raw = parseISO(dateStr);
    return isValid(raw) ? raw : fallback;
  };

  const parsedStart = safeParseISO(startDate, startOfMonth(startOfToday()));
  const parsedEnd = safeParseISO(endDate, endOfToday());
  const parsedMinDate = safeParseISO(minDate, new Date('2026-01-01'));

  const [localSelection, setLocalSelection] = useState({
    startDate: parsedStart,
    endDate: parsedEnd,
    key: 'selection',
  });

  const [activePreset, setActivePreset] = useState<string>('Custom');

  // Sync if props change - ONLY when the props actually change from the parent
  // But we want to avoid resetting the user's "working" selection if they are picking
  useEffect(() => {
    if (!isOpen) { // Only sync when the dropdown is closed to avoid jumping while picking
      const start = startDate ? parseISO(startDate) : startOfMonth(startOfToday());
      const end = endDate ? parseISO(endDate) : endOfToday();
      
      setLocalSelection({
        startDate: start,
        endDate: end,
        key: 'selection'
      });

      const startStr = format(start, 'yyyy-MM-dd');
      const endStr = format(end, 'yyyy-MM-dd');
      
      const matchingPreset = PRESETS.find(p => {
        const [pStart, pEnd] = p.getValue();
        return format(pStart, 'yyyy-MM-dd') === startStr && format(pEnd, 'yyyy-MM-dd') === endStr;
      });

      setActivePreset(matchingPreset ? matchingPreset.label : 'Custom');
    }
  }, [startDate, endDate, isOpen]);



  const handleApply = () => {
    onDateChange(
      format(localSelection.startDate, 'yyyy-MM-dd'),
      format(localSelection.endDate || localSelection.startDate, 'yyyy-MM-dd')
    );
    setIsOpen(false);
  };

  const handleCancel = () => {
    setLocalSelection({
      startDate: parsedStart,
      endDate: parsedEnd,
      key: 'selection'
    });
    setIsOpen(false);
  };

  const handlePresetClick = (preset: typeof PRESETS[0]) => {
    let [start, end] = preset.getValue();
    
    // Enforce min constraint
    if (isBefore(start, parsedMinDate)) {
      start = parsedMinDate;
    }
    
    setLocalSelection({ startDate: start, endDate: end, key: 'selection' });
    setActivePreset(preset.label);
  };

  const handleCalendarChange = (ranges: RangeKeyDict) => {
    const s = ranges.selection.startDate || new Date();
    const e = ranges.selection.endDate || new Date();
    setLocalSelection({ startDate: s, endDate: e, key: 'selection' });
    setActivePreset('Custom'); // Unselect preset on manual selection
  };

  // Display text for the closed button
  const displayRange = format(parsedStart, 'MMM d, yyyy') + (parsedStart.getTime() !== parsedEnd.getTime() ? ` → ${format(parsedEnd, 'MMM d, yyyy')}` : '');

  return (
    <div className="custom-datepicker-wrapper" ref={dropdownRef}>
      {/* Trigger Button */}
      <button 
        className="datepicker-trigger-btn"
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
        style={{ minWidth: 'auto', padding: '0.4rem 0.8rem' }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ color: 'var(--text-tertiary)' }}>
            <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
            <line x1="16" y1="2" x2="16" y2="6"></line>
            <line x1="8" y1="2" x2="8" y2="6"></line>
            <line x1="3" y1="10" x2="21" y2="10"></line>
          </svg>
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-start', gap: '2px' }}>
            <span style={{ fontWeight: 600, fontSize: '0.85rem', color: 'var(--text-primary)', lineHeight: 1 }}>
              {activePreset}
            </span>
            <span style={{ color: 'var(--text-secondary)', fontSize: '0.75rem', fontWeight: 400, whiteSpace: 'nowrap' }}>
              {displayRange}
            </span>
          </div>
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="var(--text-tertiary)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ marginLeft: '4px' }}>
            <polyline points="6 9 12 15 18 9"></polyline>
          </svg>
        </div>
      </button>

      {/* Modal Overlay and Content - Rendering via Portal to avoid overflow:hidden from parent modals */}
      {isOpen && createPortal(
        <div 
          className="datepicker-modal-overlay" 
          onMouseDown={(e) => {
            if (e.target === e.currentTarget) {
              handleCancel();
            }
          }}
        >
          <div className="datepicker-modal-container">
            <div className="datepicker-modal-header">
              <h3>Select Date Range</h3>
              <button className="modal-close-btn" onClick={handleCancel}>&times;</button>
            </div>
            
            <div className="datepicker-content">
              {/* Left Presets Sidebar */}
              <div className="datepicker-sidebar">
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
              <div className="datepicker-calendar-section">
                <div className="datepicker-inputs-row">
                  <input type="text" value={format(localSelection.startDate, 'MMM d, yyyy')} readOnly className="date-display-input" />
                  <span style={{color: 'var(--text-secondary)'}}>→</span>
                  <input type="text" value={format(localSelection.endDate || localSelection.startDate, 'MMM d, yyyy')} readOnly className="date-display-input" />
                </div>

                <div className="calendar-wrapper">
                  <DateRangePicker
                    ranges={[localSelection]}
                    onChange={handleCalendarChange}
                    months={2}
                    direction="horizontal"
                    minDate={parsedMinDate} // Strict constraint
                    moveRangeOnFirstSelection={false}
                    staticRanges={[]}
                    inputRanges={[]}
                    showMonthAndYearPickers={false} // Shopify style usually hides year drop downs
                    showPreview={false} // clean look
                  />
                </div>
              </div>
            </div>

            {/* Footer Action Bar */}
            <div className="datepicker-footer">
              <button className="btn-secondary" onClick={handleCancel}>Cancel</button>
              <button className="btn-primary" onClick={handleApply}>Apply</button>
            </div>
          </div>
        </div>,
        document.body
      )}
    </div>
  );
};
