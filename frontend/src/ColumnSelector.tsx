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
}

export const ColumnSelector: React.FC<ColumnSelectorProps> = ({ columns, visibleColumns, onChange }) => {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

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
      <button className="btn-secondary" onClick={() => setIsOpen(!isOpen)} style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', padding: '0.4rem 0.8rem', fontSize: '0.85rem' }}>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect><line x1="9" y1="3" x2="9" y2="21"></line><line x1="15" y1="3" x2="15" y2="21"></line></svg>
        Customize Columns
      </button>

      {isOpen && (
        <div className="column-selector-dropdown">
          {categories.map(category => (
            <div key={category} className="column-category">
              <div className="column-category-title">{category}</div>
              {columns.filter(c => c.category === category).map(col => (
                <label key={col.id} className="column-option">
                  <input
                    type="checkbox"
                    checked={visibleColumns.includes(col.id)}
                    onChange={() => toggleColumn(col.id)}
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
