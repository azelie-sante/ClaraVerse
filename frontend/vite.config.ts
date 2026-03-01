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
        // Uses VITE_API_BASE_URL from .env, defaults to localhost for local dev
        target: process.env.VITE_API_BASE_URL || 'http://localhost:3001',
        changeOrigin: true,
      },
    },
  },
});
