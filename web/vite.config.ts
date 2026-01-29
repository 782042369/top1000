import { splitChunks } from '@xiaowaibuzheng/rolldown-vite-split-chunks'
import { defineConfig } from 'vite'
import { createHtmlPlugin } from 'vite-plugin-html'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    splitChunks(),
    createHtmlPlugin({
      minify: true,
    }),
  ],
  build: {
    rollupOptions: {
      output: {
        chunkFileNames: 'js/[name]-[hash].js', // 引入文件名的名称
        entryFileNames: 'js/[name]-[hash].js', // 包的入口文件名称
        assetFileNames: '[ext]/[name]-[hash].[ext]', // 资源文件像 字体，图片等
      },
    },
    emptyOutDir: true,
    chunkSizeWarningLimit: 1000,
  },
  server: {
    open: true,
    host: '0.0.0.0',
    proxy: {
      '/sites.json': {
        target: 'http://127.0.0.1:7066',
        changeOrigin: true,
      },
      '/top1000.json': {
        target: 'http://127.0.0.1:7066',
        changeOrigin: true,
      },
    },
  },
})
