import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://119.13.104.225:8080',
        changeOrigin: true,
        ws: true,
      },
      '/static': {
        target: 'http://119.13.104.225:8080',
        changeOrigin: true,
      },
      '/storage': {
        target: 'http://119.13.104.225:8080',
        changeOrigin: true,
      },
    },
  },
})
