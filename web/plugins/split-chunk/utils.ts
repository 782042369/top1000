/*
 * @Author: yanghongxuan
 * @Date: 2025-04-23 15:00:35
 * @Description:
 * @LastEditTime: 2025-04-23 16:21:23
 * @LastEditors: yanghongxuan
 */
import crypto from 'node:crypto'
import os from 'node:os'
import path from 'node:path'

export function slash(p: string): string {
  return p.replace(/\\/g, '/')
}
export const isWindows = os.platform() === 'win32'

export function normalizePath(id: string): string {
  let key = path.posix.normalize(isWindows ? slash(id) : id)
  if (key.charCodeAt(0) === 0) {
    key = key.substring(1)
  }
  if (/\.vue\?vue/.test(key)) {
    key = key.split(/\?vue/)[0]
  }
  return crypto
    .createHash('sha1')
    .update(key)
    .digest('base64')
    // replace `+=/` that may be escaped in the url
    // https://github.com/umijs/umi/issues/9845
    .replace(/\//g, '')
    .replace(/\+/g, '-')
    .replace(/=/g, '')
}

export const nodeName = (name: string) => name.toString().match(/\/node_modules\/(?!.pnpm)(?<moduleName>[^\\/]*)\//)?.groups?.moduleName
