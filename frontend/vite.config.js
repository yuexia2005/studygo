import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 3000,
    proxy: {
      '/api': 'http://localhost:8085',
      '/login': 'http://localhost:8085',
      '/register': 'http://localhost:8085',
      '/health': 'http://localhost:8085',
      '/uploads': 'http://localhost:8085',
    },
  },
})
