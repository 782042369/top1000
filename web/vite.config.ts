import { splitChunks } from '@xiaowaibuzheng/rolldown-vite-split-chunks'
import { defineConfig } from 'vite'
import { createHtmlPlugin } from 'vite-plugin-html'

const API_TARGET = 'http://127.0.0.1:7066'

export default defineConfig({
  plugins: [
    splitChunks(),
    createHtmlPlugin({ minify: true }),
  ],
  build: {
    rollupOptions: {
      output: {
        chunkFileNames: 'js/[name]-[hash].js',
        entryFileNames: 'js/[name]-[hash].js',
        assetFileNames: '[ext]/[name]-[hash].[ext]',
      },
    },
    emptyOutDir: true,
    chunkSizeWarningLimit: 1000,
  },
  server: {
    open: true,
    host: '0.0.0.0',
    proxy: {
      '/sites.json': { target: API_TARGET, changeOrigin: true },
      '/top1000.json': { target: API_TARGET, changeOrigin: true },
    },
  },
})
