import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  optimizeDeps: {
    include: [
      '@atlaskit/pragmatic-drag-and-drop/element/adapter',
      '@atlaskit/pragmatic-drag-and-drop/element/set-custom-native-drag-preview',
      '@atlaskit/pragmatic-drag-and-drop/element/pointer-outside-of-preview',
    ],
  },
  server: {
    proxy: {
      '/api': {
        // Dev-only proxy target. Deliberately NOT named VITE_*: Vite injects
        // VITE_-prefixed vars into the client bundle, which would make the
        // app call the backend on an absolute cross-origin URL and fail CORS.
        target: process.env.VITE_DEV_PROXY_TARGET || 'http://localhost:3001',
        changeOrigin: true,
      },
    },
  },
});
