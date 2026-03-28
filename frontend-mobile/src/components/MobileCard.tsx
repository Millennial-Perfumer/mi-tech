import React from 'react';
import './MobileCard.css';

interface MobileCardProps {
  children: React.ReactNode;
  onClick?: () => void;
  className?: string;
  variant?: 'elevated' | 'flat';
}

export const MobileCard: React.FC<MobileCardProps> = ({
  children,
  onClick,
  className = '',
  variant = 'elevated'
}) => {
  return (
    <div
      className={`mobile-card ${variant} ${onClick ? 'clickable' : ''} ${className}`}
      onClick={onClick}
    >
      {children}
    </div>
  );
};
