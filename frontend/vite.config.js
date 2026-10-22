import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// In dev, Vite forwards /api/* to the Go backend and strips the "/api" prefix.
// In Docker and Kubernetes, Nginx / the Ingress does the same job.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },
  test: {
    environment: 'jsdom',
    setupFiles: './src/setupTests.js',
  },
})
