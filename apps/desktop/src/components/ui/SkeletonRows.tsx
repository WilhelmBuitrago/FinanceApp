/* Skeleton rows — structural loading placeholder matching the real table
   layout (header bar + N rows). First load only; refetches keep old data. */

export const SkeletonRows = ({ rows = 5, className = '' }: { rows?: number; className?: string }) => (
  <div className={`overflow-hidden rounded-tk border border-line bg-surface ${className}`} aria-hidden="true">
    <div className="flex items-center gap-4 bg-raised px-3 py-2">
      <div className="h-2.5 w-20 animate-pulse rounded-tk bg-interactive" />
      <div className="h-2.5 flex-1 animate-pulse rounded-tk bg-interactive" />
      <div className="h-2.5 w-16 animate-pulse rounded-tk bg-interactive" />
    </div>
    {Array.from({ length: rows }).map((_, i) => (
      <div key={i} className="flex items-center gap-4 border-t border-line/60 px-3 py-2.5">
        <div className="h-3 w-14 animate-pulse rounded-tk bg-raised" />
        <div className="h-3 flex-1 animate-pulse rounded-tk bg-raised" />
        <div className="h-3 w-16 animate-pulse rounded-tk bg-raised" />
      </div>
    ))}
  </div>
);
