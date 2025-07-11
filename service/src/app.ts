/*
 * @Author: 杨宏旋
 * @Date: 2020-07-01 14:41:22
 * @LastEditors: yanghongxuan
 * @LastEditTime: 2025-05-07 11:37:28
 * @Description: 完善了错误处理、安全头、SPA路由支持等功能
 */
import compress from '@fastify/compress'
import helmet from '@fastify/helmet'
import fastifyStatic from '@fastify/static'
import path from 'node:path'

import { server } from './core'
import { cacheHeader, checkExpired, scheduleJob } from './utils'

server.register(
  helmet,
  {
    contentSecurityPolicy: {
      directives: {
        'scriptSrc': ['\'self\'', 'https://log.939593.xyz'],
        'connect-src': ['\'self\'', 'https://log.939593.xyz'], // 添加允许的 API
      },
    },
  },
)

server.register(compress, {
  global: true,
  encodings: ['gzip'], // 仅启用 gzip
})
// 静态文件服务
const publicPath = path.join(__dirname, '../public')
const cacheHeaderText = cacheHeader({
  public: true,
  maxAge: '1year',
  immutable: true,
})
const noCacheHeaderText = cacheHeader({
  public: true,
  maxAge: '0ms',
})
server.register(fastifyStatic, {
  root: publicPath, // 静态文件根目录
  cacheControl: false, // 必须设置 才能动态
  setHeaders: (res, filePath) => {
    const ext = path.extname(filePath).toLowerCase()
    if (ext === '.json') {
      checkExpired()
    }
    res.setHeader('cache-control', ['.html', '.json'].includes(ext) ? noCacheHeaderText : cacheHeaderText)
  },
})

// 定时任务
scheduleJob()

server.listen({ port: 7066, host: '0.0.0.0' }).then(() => {
  server.log.info('server is running')
})
