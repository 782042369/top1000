// utils/logger.ts
import { inspect } from 'node:util'
import winston from 'winston'

const { combine, timestamp, colorize, printf } = winston.format

// 智能对象格式化函数
function smartStringify(data: unknown) {
  if (data instanceof Error) {
    return `${data.message}\n${data.stack}`
  }
  if (typeof data === 'object') {
    return inspect(data, {
      depth: null,
      colors: true,
      compact: false,
      breakLength: Infinity,
    })
  }
  return data
}

const logFormat = printf(({ level, message, timestamp, ...meta }) => {
  const ts = (timestamp as string).slice(0, 19).replace('T', ' ')
  let logMessage = `[${ts}] ${level}:`

  // 处理多参数日志 (logger.info('msg', context))
  if (meta[Symbol.for('splat') as unknown as string]) {
    // @ts-ignore
    const args = [message, ...meta[Symbol.for('splat') as unknown as string]]
    logMessage += args.map(smartStringify).join(' ')
  } else {
    logMessage += smartStringify(message)
  }

  // 处理元数据
  if (Object.keys(meta).length > 0) {
    logMessage += `\n${smartStringify(meta)}`
  }

  return logMessage
})

const logger = winston.createLogger({
  level: 'debug',
  format: combine(timestamp(), colorize(), logFormat),
  transports: [new winston.transports.Console()],
  exceptionHandlers: [new winston.transports.Console()],
  rejectionHandlers: [new winston.transports.Console()],
})

export default logger
