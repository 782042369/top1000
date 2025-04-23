/*
 * @Author: yanghongxuan
 * @Date: 2025-04-23 15:00:35
 * @Description:
 * @LastEditTime: 2025-04-23 17:24:42
 * @LastEditors: yanghongxuan
 */
import type { BuildEnvironmentOptions, Plugin } from 'vite'

import { init } from 'es-module-lexer'
import assert from 'node:assert'
import path from 'node:path'

import { staticImportedScan } from './staticImportScan'
import { nodeName, normalizePath } from './utils'

type SingleArrayType<T> = T extends (infer U)[] ? U : T

type ManualChunksOption = SingleArrayType<BuildEnvironmentOptions['rollupOptions']['output']>['manualChunks']
function wrapCustomSplitConfig(manualChunks: ManualChunksOption): ManualChunksOption {
  assert(typeof manualChunks === 'function')
  return (
    moduleId,
    { getModuleIds, getModuleInfo }
  ) => {
    return manualChunks(moduleId, { getModuleIds, getModuleInfo })
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
        } else {
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
