import type { GridOptions } from 'ag-grid-community'

import { AG_GRID_LOCALE_CN } from '@ag-grid-community/locale'
import {
  ClientSideRowModelModule,
  createGrid,
  CustomFilterModule,
  DateFilterModule,
  LocaleModule,
  ModuleRegistry,
  NumberFilterModule,
  TextFilterModule,
  ValidationModule,
} from 'ag-grid-community'
import { GroupFilterModule, LicenseManager, MultiFilterModule, SetFilterModule } from 'ag-grid-enterprise'

import './index.css'
import type { DataType } from './types'

import { convertSizeToKb, fetchData } from './utils'
import { operationRender } from './utils/operationRender'

// AG Grid 企业版许可证
LicenseManager.setLicenseKey(
  '[v3][RELEASE][0102]_NDg2Njc4MzY3MDgzNw==16d78ca762fb5d2ff740aed081e2af7b',
)

// 注册 AG Grid 模块
ModuleRegistry.registerModules([
  ClientSideRowModelModule,
  TextFilterModule,
  NumberFilterModule,
  DateFilterModule,
  SetFilterModule,
  MultiFilterModule,
  GroupFilterModule,
  CustomFilterModule,
  LocaleModule,
  ...(process.env.NODE_ENV !== 'production' ? [ValidationModule] : []),
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
