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
import { CsvExportModule, ExcelExportModule, LicenseManager } from 'ag-grid-enterprise'

import './index.css'
import type { DataType } from './types'

import { columnDefs, defaultColDef, interactionConfig, performanceConfig } from './gridConfig'
import { fetchData } from './utils'
import { loadSitesConfig } from './utils/config'

// 排除列的表头名称
const EXCLUDED_COLUMN = '操作'

// 设置 AG Grid Enterprise License
LicenseManager.setLicenseKey(
  '[v3][RELEASE][0102]_NDg2Njc4MzY3MDgzNw==16d78ca762fb5d2ff740aed081e2af7b',
)

// Grid API 引用
let gridApi: any = null

// 生成导出文件名（带日期）
function getExportFileName(extension: string): string {
  const date = new Date().toISOString().slice(0, 10)
  return `top1000-${date}.${extension}`
}

// 判断列是否应该导出（排除操作列）
function shouldExportColumn(params: any): boolean {
  const colDef = params.column.getColDef()
  return colDef.headerName !== EXCLUDED_COLUMN
}

// 初始化应用（预加载站点配置）
async function initApp() {
  try {
    // 预加载站点配置
    await loadSitesConfig()

    // 初始化表格
    initGrid()
  }
  catch (error) {
    console.error('❌ 应用初始化失败:', error)
    // 显示错误信息给用户
    const rootElement = document.querySelector<HTMLElement>('#root')
    if (rootElement) {
      rootElement.innerHTML = `
        <div style="padding: 20px; color: #fff;">
          <h2>❌ 应用加载失败</h2>
          <p>无法加载站点配置，请检查网络连接或联系管理员。</p>
        </div>
      `
    }
  }
}

// 初始化表格
function initGrid(): void {
  ModuleRegistry.registerModules([
    ClientSideRowModelModule,
    TextFilterModule,
    LocaleModule,
    CsvExportModule,
    ExcelExportModule,
  ])

  const gridOptions: GridOptions<DataType> = {
    theme: themeAlpine,
    localeText: AG_GRID_LOCALE_CN,
    defaultColDef,
    onGridReady: (params) => {
      gridApi = params.api
      fetchData(params)
    },
    getRowId: params => `${params.data.id}`,
    columnDefs,
    ...performanceConfig,
    ...interactionConfig,
  }

  const rootElement = document.querySelector<HTMLElement>('#root')
  if (!rootElement) {
    console.error('❌ 未找到根元素 #root')
    return
  }

  createGrid(rootElement, gridOptions)
  setupExportButtons()
}

// 设置导出按钮事件
function setupExportButtons(): void {
  const exportCsvBtn = document.querySelector<HTMLElement>('#exportCsv')
  const exportExcelBtn = document.querySelector<HTMLElement>('#exportExcel')

  exportCsvBtn?.addEventListener('click', () => {
    if (gridApi) {
      gridApi.exportDataAsCsv({
        fileName: getExportFileName('csv'),
        shouldExportColumn,
      })
    }
  })

  exportExcelBtn?.addEventListener('click', () => {
    if (gridApi) {
      gridApi.exportDataAsExcel({
        fileName: getExportFileName('xlsx'),
        shouldExportColumn,
      })
    }
  })
}

// 启动应用
initApp()
