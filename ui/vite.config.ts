import react from '@vitejs/plugin-react';
import path from 'path';
import { defineConfig } from 'vite';

const fliptAddr = process.env.FLIPT_ADDRESS ?? 'http://localhost:8080';

// https://vitejs.dev/config/
export default defineConfig({
  base: '',
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
    host: true,
    proxy: {
      '/api/v1': fliptAddr,
      '/auth/v1': fliptAddr,
      '/evaluate/v1': fliptAddr,
      '/internal/v1': fliptAddr,
      '/meta': fliptAddr
    },
    origin: 'http://localhost:5173'
  }
});
