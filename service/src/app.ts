/*
 * @Author: 杨宏旋
 * @Date: 2020-07-01 14:41:22
 * @LastEditors: yanghongxuan
 * @LastEditTime: 2025-03-14 15:18:04
 * @Description: 完善了错误处理、安全头、SPA路由支持等功能
 */
import http from 'http';
import Koa from 'koa';
import compress from 'koa-compress';
import helmet from 'koa-helmet';
import KoaStatic from 'koa-static';
import path from 'path';
import { logger, normalizePort, onError, scheduleJob } from './utils';

const app = new Koa();

// 安全头设置
app.use(helmet());

// GZIP 压缩
app.use(
  compress({
    threshold: 2048, // 超过 2KB 才压缩
    gzip: { flush: require('zlib').constants.Z_SYNC_FLUSH },
    deflate: false,
    br: false,
  }),
);

// 定时任务
scheduleJob();

// 请求日志中间件
app.use(async (ctx, next) => {
  const start = Date.now();
  try {
    await next();
    const ms = Date.now() - start;
    logger.info({
      method: ctx.method,
      url: ctx.url,
      status: ctx.status,
      responseTime: `${ms}ms`,
    });
  } catch (err: any) {
    const ms = Date.now() - start;
    logger.error({
      method: ctx.method,
      url: ctx.url,
      error: err.message,
      responseTime: `${ms}ms`,
    });
    throw err;
  }
});

// 错误处理中间件
app.use(async (ctx, next) => {
  try {
    await next();
  } catch (err: any) {
    ctx.status = err.statusCode || err.status || 500;
    ctx.body = {
      code: ctx.status,
      message: err.expose ? err.message : 'Internal Server Error',
    };
    ctx.app.emit('error', err, ctx);
  }
});

// 静态文件服务
const publicPath = path.join(__dirname, '../dist');

// SPA 路由回退
app.use(
  KoaStatic(publicPath, {
    setHeaders: (res, path) => {
      if (path.endsWith('top1000.json') || path.includes('.html')) {
        res.setHeader('Cache-Control', 'no-cache, max-age=0');
      } else {
        res.setHeader('Cache-Control', 'public, max-age=31536000, immutable');
      }
    },
    maxage: 1000 * 60 * 60 * 365 * 24,
    hidden: true,
    gzip: true,
    brotli: false,
  }),
);

// 服务器配置
const port = normalizePort(process.env.PORT || '7066');
const server = http.createServer(app.callback());

// 优雅关闭
const shutdown = () => {
  logger.info('Shutting down server...');
  server.close(() => {
    logger.info('Server closed');
    process.exit(0);
  });

  setTimeout(() => {
    logger.error('Force shutdown');
    process.exit(1);
  }, 5000);
};

process.on('SIGINT', shutdown);
process.on('SIGTERM', shutdown);

// 启动服务器
server.listen(port);
server.on('error', err => onError(err, port.toString()));
server.on('listening', () => {
  const addr = server.address();
  const bind = typeof addr === 'string' ? `pipe ${addr}` : `port ${addr?.port}`;
  logger.info(`Server started on ${bind}`);
  console.log(`Listening on ${bind}`);
});
