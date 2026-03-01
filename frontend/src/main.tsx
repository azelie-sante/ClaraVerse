import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import './index.css'; // Tailwind first
import './styles/design-tokens.css'; // Design tokens second (won't override Tailwind utilities)
import '@xyflow/react/dist/style.css'; // React Flow styles (global for proper rendering)
import App from './App.tsx';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>
);
