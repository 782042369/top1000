/*
 * @Author: yanghongxuan
 * @Date: 2023-07-10 11:12:24
 * @LastEditors: yanghongxuan
 * @LastEditTime: 2024-02-04 16:19:26
 * @Description:
 */
/**
 * Normalize a port into a number, string, or false.
 */


export function normalizePort(val: string) {
  const portNum = parseInt(val, 10);

  if (isNaN(portNum)) {
    // named pipe
    return val;
  }

  if (portNum >= 0) {
    // port number
    return portNum;
  }

  return false;
}

/**
 * Event listener for HTTP server "error" event.
 */

export function onError(error: any, port: string) {
  if (error.syscall !== "listen") {
    throw error;
  }

  const bind = typeof port === "string" ? `Pipe ${port}` : `Port ${port}`;

  // handle specific listen errors with friendly messages
  switch (error.code) {
    case "EACCES":
        console.error(`${bind} requires elevated privileges`);
      process.exit(1);
      break;
    case "EADDRINUSE":
        console.error(`${bind} is already in use`);
      process.exit(1);
      break;
    default:
      throw error;
  }
}
