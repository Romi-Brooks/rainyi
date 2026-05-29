import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// const baseUrl = 'http://localhost:8080'
const baseUrl = 'http://119.13.104.225:8080'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: baseUrl,
        changeOrigin: true,
        ws: true,
      },
      '/static': {
        target: baseUrl,
        changeOrigin: true,
      },
      '/storage': {
        target: baseUrl,
        changeOrigin: true,
      },
    },
  },
})
