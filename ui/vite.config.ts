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
    manifest: true,
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor-react': ['react', 'react-dom', 'react-router'],
          'vendor-redux': ['@reduxjs/toolkit', 'react-redux'],
          'vendor-codemirror': [
            '@codemirror/view',
            '@codemirror/state',
            '@codemirror/lang-json',
            '@codemirror/lint',
            '@codemirror/search'
          ],
          'vendor-charts': ['chart.js', 'react-chartjs-2'],
          'vendor-forms': ['formik', 'yup', 'zod'],
          'vendor-icons': ['lucide-react']
        }
      }
    }
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
