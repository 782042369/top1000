import type { GridOptions } from 'ag-grid-community'

import { AG_GRID_LOCALE_CN } from '@ag-grid-community/locale'
import {
  ClientSideRowModelModule,
  createGrid,
  LocaleModule,
  ModuleRegistry,
  TextFilterModule,
  themeAlpine,
} from 'ag-grid-community'

import './index.css'
import type { DataType } from './types'

import { convertSizeToKb, fetchData } from './utils'
import { operationRender } from './utils/operationRender'

// 注册 AG Grid 模块（仅导入实际使用的模块，减少体积）
ModuleRegistry.registerModules([
  ClientSideRowModelModule, // 客户端行模型（必需）
  TextFilterModule, // 文本过滤（名字字段使用）
  LocaleModule, // 中文本地化（必需）
])

// 表格配置
const gridOptions: GridOptions<DataType> = {
  theme: themeAlpine,
  localeText: AG_GRID_LOCALE_CN,
  defaultColDef: {
    flex: 1,
    sortable: false,
    resizable: true, // 允许调整列宽
  },
  onGridReady: fetchData,
  getRowId: params => `${params.data.id}`,
  columnDefs: [
    {
      headerName: '名字',
      field: 'siteName',
      filter: true,
      pinned: 'left', // 固定左侧列
      minWidth: 120,
    },
    {
      headerName: '资源ID',
      field: 'siteid',
      minWidth: 100,
    },
    {
      headerName: '重复度',
      field: 'duplication',
      sortable: true,
      comparator: (valueA, valueB) => {
        // 字符串比较（带百分号）
        const numA = Number.parseFloat(valueA) || 0
        const numB = Number.parseFloat(valueB) || 0
        return numA - numB
      },
      minWidth: 100,
    },
    {
      headerName: '文件大小',
      field: 'size',
      sortable: true,
      comparator: (valueA, valueB) => convertSizeToKb(valueA) - convertSizeToKb(valueB),
      minWidth: 120,
    },
    {
      headerName: '操作',
      cellRenderer: operationRender,
      pinned: 'right', // 固定右侧列
      minWidth: 180,
      flex: 0, // 不允许自动伸缩
    },
  ],

  // 性能优化配置
  rowBuffer: 10, // 缓冲区行数（虚拟滚动）
  enableCellTextSelection: true, // 允许选择文本
  suppressDragLeaveHidesColumns: true, // 拖拽离开不隐藏列

  // 交互优化
  suppressRowClickSelection: true, // 禁用点击选择行
}

// 初始化表格
const rootElement = document.querySelector<HTMLElement>('#root')
if (rootElement) {
  createGrid(rootElement, gridOptions)
}
else {
  console.error('❌ 未找到根元素 #root')
}
