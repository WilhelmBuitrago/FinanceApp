/* Button classes — one primary/secondary pair for every screen. Three
   near-identical copies of the primary variant lived in the pages
   before the audit. Radii use the declared tokens, not magic values. */

export const primaryButton =
  'inline-flex h-8 items-center rounded-tk bg-accent-strong px-3 text-body font-medium text-white transition-colors duration-[120ms] hover:bg-accent-deep focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent disabled:opacity-40 disabled:hover:bg-accent-strong';

export const secondaryButton =
  'inline-flex h-8 items-center rounded-tk border border-line bg-raised px-3 text-body font-medium text-ink transition-colors duration-[120ms] hover:bg-interactive focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent';
