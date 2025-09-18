import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import path from 'path';
import { defineConfig } from 'vite';

const fliptAddr = process.env.FLIPT_ADDRESS ?? 'http://localhost:8080';

// https://vitejs.dev/config/
export default defineConfig({
  base: '',
  plugins: [react(), tailwindcss()],
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
      '/auth/v1': fliptAddr,
      '/evaluate/v1': fliptAddr,
      '/internal/v1': fliptAddr,
      '/internal/v2': fliptAddr,
      '/meta': fliptAddr,
      '/api/v1': fliptAddr,
      '/api/v2': fliptAddr
    },
    origin: 'http://localhost:5173'
  }
});
