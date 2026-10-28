import pkg from '../../package.json';

/* The one source of truth for the version the UI shows. Reading it from
   package.json keeps the sidebar label and the component version in
   sync; a stale hardcoded string is what drifted before. */

export const APP_VERSION: string = pkg.version;
