import type { GridOptions } from 'ag-grid-community'

import { AG_GRID_LOCALE_CN } from '@ag-grid-community/locale'
import {
  ClientSideRowModelModule,
  createGrid,
  LocaleModule,
  ModuleRegistry,
  TextFilterModule,
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
  localeText: AG_GRID_LOCALE_CN,
  defaultColDef: {
    flex: 1,
    sortable: false,
  },
  onGridReady: fetchData,
  getRowId: params => `${params.data.id}`,
  columnDefs: [
    { headerName: '名字', field: 'siteName', filter: true },
    { headerName: '资源ID', field: 'siteid' },
    { headerName: '重复度', field: 'duplication', sortable: true },
    {
      headerName: '文件大小',
      field: 'size',
      sortable: true,
      comparator: (valueA, valueB) => convertSizeToKb(valueA) - convertSizeToKb(valueB),
    },
    { headerName: '操作', cellRenderer: operationRender },
  ],
}

// 初始化表格
const rootElement = document.querySelector<HTMLElement>('#root')
if (rootElement) {
  createGrid(rootElement, gridOptions)
}
