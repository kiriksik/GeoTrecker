import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import fs from 'fs'

export default defineConfig({
  plugins: [vue()],
  server: {
    host: '192.168.31.242',
    https: {
      key: fs.readFileSync('./192.168.31.242-key.pem'),
      cert: fs.readFileSync('./192.168.31.242.pem'),
    },
    proxy: {
      '/api': {
        target: 'http://192.168.31.242:8080',
        changeOrigin: true,
        secure: false,
      },
      '/ws': {
        target: 'ws://192.168.31.242:8080',
        ws: true,
      },
    },
  },
})

