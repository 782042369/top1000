/*
 * @Author: yanghongxuan
 * @Date: 2024-02-04 16:54:18
 * @Description:
 * @LastEditTime: 2025-04-23 17:13:04
 * @LastEditors: yanghongxuan
 */
import { antdResolver } from '@bit-ocean/auto-import'
import react from '@vitejs/plugin-react-swc'
import { resolve } from 'node:path'
import AutoImport from 'unplugin-auto-import/vite'
import { defineConfig } from 'vite'

import splitChunk from './plugins/split-chunk'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    react(),
    AutoImport({
      imports: ['react'],
      dts: '@types/auto-imports.d.ts',
      include: [
        /\.[tj]sx?$/, // .ts, .tsx, .js, .jsx
        /\.md$/, // .md
      ],
      resolvers: [antdResolver()],
    }),
    splitChunk()
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
    outDir: resolve(__dirname, '../service/dist'),
  },
  server: {
    proxy: {
      '/top1000': {
        target: 'http://top1000.939593.xyz',
        changeOrigin: true,
      },
    },
    open: true,
    host: '0.0.0.0',
  },
})
