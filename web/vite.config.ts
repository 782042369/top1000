/*
 * @Author: yanghongxuan
 * @Date: 2024-02-04 16:54:18
 * @Description:
 * @LastEditTime: 2025-04-21 10:01:35
 * @LastEditors: yanghongxuan
 */
import { antdResolver } from '@bit-ocean/auto-import'
import react from '@vitejs/plugin-react-swc'
import { resolve } from 'node:path'
import AutoImport from 'unplugin-auto-import/vite'
import { defineConfig } from 'vite'
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
  ],
  build: {
    rollupOptions: {
      // 自定义底层的 Rollup 打包配d置
      output: {
        chunkFileNames: 'js/[name]-[hash].js', // 引入文件名的名称
        entryFileNames: 'js/[name]-[hash].js', // 包的入口文件名称
        assetFileNames: '[ext]/[name]-[hash].[ext]', // 资源文件像 字体，图片等
        experimentalMinChunkSize: 5 * 1024, // 生成的chunk最小体积，小于这个值的chunk会被合并到一个文件中
        manualChunks(id) {
          if (id.includes('node_modules')) {
            return (
              id
                .toString()
                .match(/\/node_modules\/(?!.pnpm)(?<moduleName>[^\\/]*)\//)
                ?.groups!.moduleName ?? 'vender'
            )
          }
          if (id.includes('vite')) {
            return 'vite'
          }
          if (id.includes('/utils/')) {
            return 'utils'
          }
          console.log('%c [ id ]-36', 'font-size:13px; background:#41bf39; color:#85ff7d;', id)
        },
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
