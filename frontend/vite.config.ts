import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  test: {
    globals: true,
    environment: 'happy-dom',
    setupFiles: './src/test/setup.ts',
    css: false,
  },
  base: '/revisitr/',
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 5921,
    allowedHosts: true,
    proxy: {
      '/api': {
        target: 'http://localhost:9721',
        changeOrigin: true,
      },
      '/revisitr/storage': {
        target: 'http://localhost:9721',
        changeOrigin: true,
      },
    },
  },
})
