import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';
const path = require('path');

const fliptAddr = process.env.FLIPT_ADDRESS ?? 'http://localhost:8080';
const monacoPrefix = `monaco-editor/esm/vs`;

// https://vitejs.dev/config/
export default defineConfig({
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
          jsonWorker: [`${monacoPrefix}/language/json/json.worker`],
          typescriptWorker: [`${monacoPrefix}/language/typescript/ts.worker`],
          editorWorker: [`${monacoPrefix}/editor/editor.worker`]
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
      '/meta': fliptAddr
    },
    origin: 'http://localhost:5173'
  }
});
