import { Link } from 'react-router-dom';

/* Section eyebrow (mono 11px uppercase) with optional contextual action
   link. Used across pages for the section rhythm. */

interface SectionHeaderProps {
  title: string;
  action?: { to: string; label: string };
}

export function SectionHeader({ title, action }: SectionHeaderProps) {
  return (
    <div className="mb-3 flex items-baseline justify-between">
      <h2 className="font-mono text-meta font-medium uppercase tracking-[0.08em] text-ink-3">{title}</h2>
      {action && (
        <Link to={action.to} className="text-meta font-medium text-accent transition-colors hover:underline">
          {action.label}
        </Link>
      )}
    </div>
  );
}
