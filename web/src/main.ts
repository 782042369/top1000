/*
 * @Author: yanghongxuan
 * @Date: 2024-02-04 16:54:18
 * @LastEditors: yanghongxuan
 * @LastEditTime: 2024-02-04 17:11:06
 * @Description:
 */
import type { GridOptions } from 'ag-grid-community'

import { AG_GRID_LOCALE_CN } from '@ag-grid-community/locale'

import './index.css'
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

import type {
  DataType,
} from './types'

import { convertSizeToKb, fetchData } from './utils'
import { operationRender } from './utils/operationRender'

LicenseManager.setLicenseKey(
  '[v3][RELEASE][0102]_NDg2Njc4MzY3MDgzNw==16d78ca762fb5d2ff740aed081e2af7b',
)
// https://www.ag-grid.com/vue-data-grid/modules/
ModuleRegistry.registerModules([
  ClientSideRowModelModule,
  // ag-grid-community
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

const gridOptions: GridOptions<DataType> = {
  localeText: AG_GRID_LOCALE_CN,
  defaultColDef: {
    flex: 1,
    sortable: false,
  },
  onGridReady: fetchData,
  getRowId: params => `${params.data.id}`,
  columnDefs: [
    {
      headerName: '名字',
      field: 'siteName',
      filter: true,
    },
    {
      headerName: '资源ID',
      field: 'siteid',
    },
    {
      headerName: '重复度',
      field: 'duplication',
      sortable: true,
    },
    {
      headerName: '文件大小',
      field: 'size',
      sortable: true,
      comparator: (valueA, valueB) => {
        return convertSizeToKb(valueA) - convertSizeToKb(valueB)
      },
    },
    {
      headerName: '操作',
      cellRenderer: operationRender,
    },
  ],
}

createGrid(
  document.querySelector<HTMLElement>('#root')!,
  gridOptions,
)
