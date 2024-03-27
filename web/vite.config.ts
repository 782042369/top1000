/*
 * @Author: yanghongxuan
 * @Date: 2024-02-04 16:54:18
 * @Description:
 * @LastEditTime: 2024-03-27 17:35:05
 * @LastEditors: yanghongxuan
 */
import react from '@vitejs/plugin-react-swc';
import { resolve } from 'node:path';
import { defineConfig } from 'vite';
// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
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
            );
          }
        },
      },
    },
    emptyOutDir: false,
    outDir: resolve(__dirname, '../service/static'),
  },
});
