/*
 * @Author: yanghongxuan
 * @Date: 2025-03-19 09:45:19
 * @Description:
 * @LastEditTime: 2025-04-21 09:41:33
 * @LastEditors: yanghongxuan
 */
import antfu from '@antfu/eslint-config'

export default antfu({
  node: true,
  typescript: true,
  stylistic: {
    indent: 2, // 缩进
    semi: false, // 语句分号
    quotes: 'single', // 单引号
  },
  rules: {
    'perfectionist/sort-imports': [
      'error',
      {
        partitionByComment: true,
        type: 'natural',
        order: 'asc',
      },
    ],
  },
  ignores: [
    '**/node_modules/**',
    'pnpm-lock.yaml',
    'dist',
  ],
  formatters: true,
  jsonc: false,
  yaml: false,
})
