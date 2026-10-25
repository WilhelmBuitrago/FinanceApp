import React from 'react';

/* FormField — label always visible above the control, optional explanatory
   note below (never a placeholder pretending to be a label). */

interface FormFieldProps {
  label: string;
  htmlFor?: string;
  note?: string;
  error?: string;
  children: React.ReactNode;
}

export function FormField({ label, htmlFor, note, error, children }: FormFieldProps) {
  return (
    <div className="space-y-1.5">
      <label htmlFor={htmlFor} className="block text-body font-medium text-ink">
        {label}
      </label>
      {children}
      {note && <p className="text-meta text-ink-3">ⓘ {note}</p>}
      {error && <p className="text-meta text-neg">{error}</p>}
    </div>
  );
}

export const inputClass =
  'h-8 w-full rounded-tk border border-line bg-bg px-3 text-body text-ink placeholder:text-ink-3 focus:border-accent focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent';

export const selectClass =
  'h-8 w-full rounded-tk border border-line bg-bg px-2 text-body text-ink focus:border-accent focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent';

/* Toolbar selects size to their content instead of stretching; the
   transactions toolbar carried its own copy of this string. */

export const selectClassAuto =
  'h-8 rounded-tk border border-line bg-bg px-2 text-body text-ink focus:border-accent focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent';
