import React from 'react';

type ImportStatusProps = {
  status: 'UPLOAD' | 'PROCESSING' | 'VALIDATING' | 'IMPORTING' | 'SUCCESS' | 'ERROR';
};

export default function ImportStatus({ status }: ImportStatusProps) {
  // TODO: Replace with proper SVG icons per status
  const renderIcon = () => {
    switch (status) {
      case 'UPLOAD':
        return <svg viewBox="0 0 24 24"><text x="12" y="16" textAnchor="middle">⬆️</text></svg>;
      case 'PROCESSING':
        return <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" strokeWidth="2" /></svg>;
      case 'VALIDATING':
        return <svg viewBox="0 0 24 24"><text x="12" y="16" textAnchor="middle">🔍</text></svg>;
      case 'IMPORTING':
        return <svg viewBox="0 0 24 24"><text x="12" y="16" textAnchor="middle">📥</text></svg>;
      case 'SUCCESS':
        return <svg viewBox="0 0 24 24"><text x="12" y="16" textAnchor="middle">✅</text></svg>;
      case 'ERROR':
        return <svg viewBox="0 0 24 24"><text x="12" y="16" textAnchor="middle">❌</text></svg>;
      default:
        return null;
    }
  };

  return (
    <div className="import-status">
      {renderIcon()}
      <span>{status}</span>
    </div>
  );
}
