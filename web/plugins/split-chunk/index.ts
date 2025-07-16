/*
 * @Author: yanghongxuan
 * @Date: 2025-04-23 15:00:35
 * @Description:
 * @LastEditTime: 2025-04-25 14:00:38
 * @LastEditors: yanghongxuan
 */
import type { Plugin } from 'vite'

import { init } from 'es-module-lexer'
import assert from 'node:assert'
import path from 'node:path'

import type { ManualChunksOption } from './type'

import { staticImportedScan } from './staticImportScan'
import { nodeName, normalizePath } from './utils'

function wrapCustomSplitConfig(manualChunks: ManualChunksOption): ManualChunksOption {
  assert(typeof manualChunks === 'function')
  return (
    moduleId,
    { getModuleInfo },
  ) => {
    return manualChunks(moduleId, { getModuleInfo })
  }
}
// eslint-disable-next-line node/prefer-global/process
const cwd = process.cwd()
function generateManualChunks(): ManualChunksOption {
  return wrapCustomSplitConfig(
    (id, { getModuleInfo }) => {
      if (id.includes('node_modules')) {
        if (staticImportedScan(id, getModuleInfo, new Map(), [])) {
          return `p-${nodeName(id) ?? 'vender'
            }`
        }
        else {
          return `p-${nodeName(id) ?? 'vender'
            }-async`
        }
      }
      if (!id.includes('node_modules')) {
        const extname = path.extname(id)
        return normalizePath(path.relative(cwd, id).replace(extname, ''))
      }
    },
  )
}

export default (
): Plugin => {
  return {
    name: 'vite-plugin-chunk-split',
    async config() {
      await init
      const manualChunks = generateManualChunks()
      return {
        build: {
          rollupOptions: {
            output: {
              manualChunks,
            },
          },
        },
      }
    },
  }
}
