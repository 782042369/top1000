import type { GridOptions } from 'ag-grid-community'

import type { DataType } from './types'

import { convertSizeToKb } from './utils'
import { operationRender } from './utils/operationRender'

export const columnDefs: GridOptions<DataType>['columnDefs'] = [
  {
    headerName: '序号',
    valueGetter: params => params.data!.id,
    width: 70,
    pinned: 'left',
    lockPosition: true,
    suppressSizeToFit: true,
  },
  {
    headerName: '名字',
    field: 'siteName',
    filter: true,
    pinned: 'left',
    minWidth: 120,
    width: 150,
  },
  {
    headerName: '资源ID',
    field: 'siteid',
    minWidth: 100,
    width: 100,
  },
  {
    headerName: '重复度',
    field: 'duplication',
    sortable: true,
    comparator: (valueA, valueB) => {
      const numA = Number.parseFloat(valueA) || 0
      const numB = Number.parseFloat(valueB) || 0
      return numA - numB
    },
    minWidth: 100,
    width: 100,
  },
  {
    headerName: '文件大小',
    field: 'size',
    sortable: true,
    comparator: (valueA, valueB) => convertSizeToKb(valueA) - convertSizeToKb(valueB),
    minWidth: 120,
    width: 120,
  },
  {
    headerName: '操作',
    cellRenderer: operationRender,
    pinned: 'right',
    minWidth: 180,
    width: 180,
    flex: 0,
    lockPosition: true,
    suppressSizeToFit: true,
  },
]

export const defaultColDef: GridOptions<DataType>['defaultColDef'] = {
  flex: 1,
  sortable: false,
  resizable: true,
}

export const performanceConfig = {
  rowBuffer: 10,
  enableCellTextSelection: true,
  domLayout: 'normal' as const,
}

export const interactionConfig = {
  suppressRowClickSelection: true,
  suppressDragLeaveHidesColumns: true,
}
