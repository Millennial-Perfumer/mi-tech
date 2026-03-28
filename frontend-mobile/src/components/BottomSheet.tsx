import React, { useEffect } from 'react';
import './BottomSheet.css';

interface BottomSheetProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  children: React.ReactNode;
}

export const BottomSheet: React.FC<BottomSheetProps> = ({ isOpen, onClose, title, children }) => {
  useEffect(() => {
    if (isOpen) {
      document.body.classList.add('scroll-lock');
    } else {
      document.body.classList.remove('scroll-lock');
    }
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div className="bottom-sheet-overlay" onClick={onClose}>
      <div className="bottom-sheet-content" onClick={e => e.stopPropagation()}>
        <div className="bottom-sheet-handle" />
        {title && (
          <div className="bottom-sheet-header">
            <h3>{title}</h3>
            <button className="close-btn" onClick={onClose}>✕</button>
          </div>
        )}
        <div className="bottom-sheet-body">
          {children}
        </div>
      </div>
    </div>
  );
};
