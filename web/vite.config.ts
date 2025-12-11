/*
 * @Author: yanghongxuan
 * @Date: 2024-02-04 16:54:18
 * @Description:
 * @LastEditTime: 2025-04-23 17:13:04
 * @LastEditors: yanghongxuan
 */
import { splitChunks } from '@xiaowaibuzheng/rolldown-vite-split-chunks'
import { resolve } from 'node:path'
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
      // 自定义底层的 Rollup 打包配d置
      output: {
        chunkFileNames: 'js/[name]-[hash].js', // 引入文件名的名称
        entryFileNames: 'js/[name]-[hash].js', // 包的入口文件名称
        assetFileNames: '[ext]/[name]-[hash].[ext]', // 资源文件像 字体，图片等
      },
    },
    emptyOutDir: false,
    outDir: resolve(__dirname, '../web-dist'),
  },
  server: {
    open: true,
    host: '0.0.0.0',
  },
})
