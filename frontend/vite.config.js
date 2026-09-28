import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5310,
    strictPort: true,
    host: '127.0.0.1',
    proxy: {
      // 前端 /api 请求代理到 Go 后端，避免跨域
      '/api': {
        target: 'http://127.0.0.1:8088',
        changeOrigin: true,
      },
    },
  },
})
