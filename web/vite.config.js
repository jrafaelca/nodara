import path from 'node:path'
import { fileURLToPath } from 'node:url'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig } from 'vite'

const rootDir = path.dirname(fileURLToPath(import.meta.url))

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: { alias: { '@': path.resolve(rootDir, './src') } },
  server: {
    host: '0.0.0.0',
	proxy: {
	  '/api': { target: process.env.VITE_CORE_URL || 'http://nodara-core:8080', changeOrigin: true },
	  '/ws': { target: process.env.VITE_CORE_URL || 'ws://nodara-core:8080', changeOrigin: true, ws: true },
	},
    watch: {
      usePolling: true,
      interval: 100,
    },
  },
})
