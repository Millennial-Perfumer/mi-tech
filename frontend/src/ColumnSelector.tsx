import React, { useState, useRef, useEffect } from 'react';
import './ColumnSelector.css';

export interface ColumnOption {
  id: string;
  label: string;
  category: string;
}

interface ColumnSelectorProps {
  columns: ColumnOption[];
  visibleColumns: string[];
  onChange: (visibleColumns: string[]) => void;
  onReset?: () => void;
}

export const ColumnSelector: React.FC<ColumnSelectorProps> = ({ columns, visibleColumns, onChange, onReset }) => {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      document.addEventListener('keydown', handleKeyDown);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen]);

  const toggleColumn = (id: string) => {
    if (visibleColumns.includes(id)) {
      // Don't let users uncheck all columns (prevent empty state)
      if (visibleColumns.length > 1) {
        onChange(visibleColumns.filter(c => c !== id));
      }
    } else {
      onChange([...visibleColumns, id]);
    }
  };

  const categories = Array.from(new Set(columns.map(c => c.category)));

  return (
    <div className="column-selector" ref={dropdownRef}>
      <button
        className="btn-secondary"
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
        aria-haspopup="true"
        aria-controls="column-selector-dropdown"
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: '0.5rem',
          padding: '0.4rem 0.8rem',
          fontSize: '0.85rem',
        }}
      >
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
          <line x1="9" y1="3" x2="9" y2="21"></line>
          <line x1="15" y1="3" x2="15" y2="21"></line>
        </svg>
        Customize Columns
      </button>

      {isOpen && (
        <div className="column-selector-dropdown" id="column-selector-dropdown" role="menu">
          {onReset && (
            <button
              className="column-reset-btn"
              onClick={() => { onReset(); setIsOpen(false); }}
              role="menuitem"
            >
              Reset to default
            </button>
          )}
          {categories.map((category) => (
            <div key={category} className="column-category" role="group" aria-label={category}>
              <div className="column-category-title" aria-hidden="true">
                {category}
              </div>
              {columns
                .filter((c) => c.category === category)
                .map((col) => (
                  <label
                    key={col.id}
                    className="column-option"
                    role="menuitemcheckbox"
                    aria-checked={visibleColumns.includes(col.id)}
                  >
                    <input
                      type="checkbox"
                      checked={visibleColumns.includes(col.id)}
                      onChange={() => toggleColumn(col.id)}
                      aria-label={`Toggle ${col.label} column`}
                    />
                    {col.label}
                  </label>
                ))}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
