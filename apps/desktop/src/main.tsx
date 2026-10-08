import React from 'react';
import { createRoot } from 'react-dom/client';
import App from './app/App';

// Fonts (self-hosted, offline-capable — desktop app requirement)
import '@fontsource/ibm-plex-sans/400.css';
import '@fontsource/ibm-plex-sans/500.css';
import '@fontsource/ibm-plex-sans/600.css';
import '@fontsource/ibm-plex-mono/400.css';
import '@fontsource/ibm-plex-mono/500.css';

import './index.css';

const container = document.getElementById('root');
if (container) {
  createRoot(container).render(<App />);
}
