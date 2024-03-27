/*
 * @Author: 杨宏旋
 * @Date: 2020-07-01 14:41:22
 * @LastEditors: yanghongxuan
 * @LastEditTime: 2024-03-27 17:38:40
 * @Description:
 */
import http from 'http';
import Koa from 'koa';
import compress from 'koa-compress';
import KoaStatic from 'koa-static';
import path from 'path';
import { normalizePort, onError } from './utils';
import scheduleJob from './utils/scheduleJob';

const app = new Koa();
// GZIP
const options = { threshold: 1024 };
app.use(compress(options));
// cors
app.use(async (ctx, next) => {
  ctx.set('Access-Control-Allow-Origin', '*');
  ctx.set(
    'Access-Control-Allow-Headers',
    'Content-Type,Content-Length, Authorization, Accept,X-Requested-With',
  );
  ctx.set('Access-Control-Allow-Methods', 'PUT,POST,GET,DELETE,OPTIONS');
  if (ctx.method === 'OPTIONS') {
    ctx.body = 200;
  } else {
    await next();
  }
});
// 定时任务
scheduleJob();
// logger
app.use(async (ctx, next) => {
  const start = new Date().getTime();
  await next();
  const ms = new Date().getTime() - start;
  console.info({
    请求方式: ctx.method,
    请求地址: ctx.url,
    请求时间: ms + 'ms',
  });
});

// error-handling
app.on('error', (err, ctx) => {
  console.error('server error', err, ctx);
});
const publicPath = path.join(__dirname, '../static');
app.use(
  KoaStatic(publicPath, {
    setHeaders: (res, path) => {
      if (path.endsWith('top1000.json') || path.includes('.html')) {
        // 设置 Cache-Control 头为 no-cache，使得 top1000.json 文件每次都从服务器读取
        res.setHeader('Cache-Control', 'no-cache');
      }
    },
    // 缓存时间365天
    maxage: 1000 * 60 * 60 * 365 * 24,
    hidden: true,
    gzip: true,
  }),
);

const port = normalizePort(process.env.PORT || '7066');
/**
 * Create HTTP server.
 */

const server = http.createServer(app.callback());

/**
 * Listen on provided port, on all network interfaces.
 */

server.listen(port);
server.on('error', err => onError(err, port.toString()));
server.on('listening', () => {
  const addr = server.address();
  const bind = typeof addr === 'string' ? `pipe ${addr}` : `port ${addr?.port}`;
  console.log(`Listening on ${bind}`);
});
