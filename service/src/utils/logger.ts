/*
 * @Author: yanghongxuan
 * @Date: 2025-03-14 15:20:20
 * @Description:
 * @LastEditTime: 2025-03-14 15:29:20
 * @LastEditors: yanghongxuan
 */
import winston from 'winston';

const logger = winston.createLogger({
  level: 'info',
  format: winston.format.combine(
    winston.format.timestamp({
      format: 'YYYY-MM-DD HH:mm:ss',
    }),
    winston.format.printf(({ timestamp, level, message }) => {
      const messageStr = message?.toString() ?? message;
      return `[${timestamp}] ${level.toUpperCase()}: ${messageStr}`;
    }),
  ),
  transports: [
    new winston.transports.Console({
      handleExceptions: true,
      handleRejections: true,
    }),
  ],
});

export default logger;
