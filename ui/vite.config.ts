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
    manifest: true,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules')) {
            if (id.includes('@codemirror')) {
              return 'vendor-codemirror';
            }
            if (id.includes('chart.js') || id.includes('react-chartjs-2')) {
              return 'vendor-charts';
            }
            if (id.includes('@mui')) {
              return 'vendor-mui';
            }
            if (id.includes('@fortawesome')) {
              return 'vendor-fontawesome';
            }
            if (id.includes('@dnd-kit')) {
              return 'vendor-dndkit';
            }
            if (id.includes('date-fns') || id.includes('chartjs-adapter')) {
              return 'vendor-date';
            }
            if (
              id.includes('react') ||
              id.includes('react-dom') ||
              id.includes('react-router')
            ) {
              return 'vendor-react';
            }
            if (id.includes('@reduxjs/toolkit') || id.includes('react-redux')) {
              return 'vendor-redux';
            }
            if (id.includes('lucide-react')) {
              return 'vendor-icons';
            }
            if (
              id.includes('formik') ||
              id.includes('yup') ||
              id.includes('zod')
            ) {
              return 'vendor-forms';
            }
          }
        }
      }
    }
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
