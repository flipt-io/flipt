import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';
const path = require('path');

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '~': path.resolve(__dirname, 'src')
    }
  },
  build: {
    manifest: true
  },
  envPrefix: 'FLIPT_',
  server: {
    proxy: {
      '/api/v1': 'http://localhost:8080',
      '/auth/v1': 'http://localhost:8080',
      '/meta': 'http://localhost:8080'
    },
    origin: 'http://localhost:5173'
  }
});
