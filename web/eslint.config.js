/*
 * @Author: yanghongxuan
 * @Date: 2025-03-19 09:45:19
 * @Description:
 * @LastEditTime: 2025-04-08 14:43:34
 * @LastEditors: yanghongxuan
 */
import antfu from '@antfu/eslint-config'

export default antfu({
  react: true,
  typescript: true,
  stylistic: {
    indent: 2, // 缩进
    semi: false, // 语句分号
    quotes: 'single' // 单引号
  },
  rules: {
    'no-console': 'off', // 禁止使用 console
    'prefer-promise-reject-errors': 'off', // 允许 promise.reject()
    // 代码风格 相关规则
    'style/comma-dangle': 'off',
    'style/brace-style': 'off',
    'style/operator-linebreak': 'off',
    'antfu/consistent-list-newline': 'off',
    'antfu/if-newline': 'off',
    'perfectionist/sort-imports': [
      'error',
      {
        partitionByComment: true,
        type: 'natural',
        order: 'asc'
      }
    ],
    'import/no-duplicates': 'error',
    // jsdoc 相关规则
    'jsdoc/require-returns-description': 'off',
    'jsdoc/check-alignment': 'off',
    'jsdoc/check-param-names': 'off',
    'jsdoc/require-returns-check': 'off',
    // node 相关规则
    'node/prefer-global/process': 'off',
    // ts 相关规则
    'ts/ban-ts-comment': 'off',
    'ts/no-unsafe-function-type': 'off',
    'ts/explicit-function-return-type': 'off',
    // 正则 相关规则
    'regexp/no-unused-capturing-group': 'off',
    'regexp/no-super-linear-backtracking': 'off',
    'regexp/optimal-quantifier-concatenation': 'off'
  },
  ignores: [
    '**/node_modules/**',
    'pnpm-lock.yaml',
    '.stylelintrc.cjs'
  ],
  formatters: true,
  jsonc: false,
  yaml: false
})
