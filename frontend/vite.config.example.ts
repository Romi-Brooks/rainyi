import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// 修改为你的后端地址
const baseUrl = 'http://localhost:8080'

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
