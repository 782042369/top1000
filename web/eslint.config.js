/*
 * @Author: yanghongxuan
 * @Date: 2025-03-19 09:45:19
 * @Description:
 * @LastEditTime: 2025-04-21 09:41:39
 * @LastEditors: yanghongxuan
 */
import antfu from '@antfu/eslint-config'

export default antfu({
  typescript: true,
  stylistic: {
    indent: 2, // 缩进
    semi: false, // 语句分号
    quotes: 'single', // 单引号
  },
  rules: {
    // 代码风格 相关规则
    'perfectionist/sort-imports': [
      'error',
      {
        partitionByComment: true,
        type: 'natural',
        order: 'asc',
      },
    ],
    'node/prefer-global/process': 'off',
  },
  ignores: [
    '**/node_modules/**',
    'pnpm-lock.yaml',
    '**/*.md',
  ],
  formatters: true,

})
