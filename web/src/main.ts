import type { GridApi, GridOptions } from 'ag-grid-community'

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

const EXCLUDED_COLUMN = '操作'
const ROOT_ID = '#root'

LicenseManager.setLicenseKey(
  '[v3][RELEASE][0102]_NDg2Njc4MzY3MDgzNw==16d78ca762fb5d2ff740aed081e2af7b',
)

let gridApi: GridApi<DataType> | null = null

function getExportFileName(extension: string): string {
  const date = new Date().toISOString().slice(0, 10)
  return `top1000-${date}.${extension}`
}

function shouldExportColumn(params: any): boolean {
  const colDef = params.column.getColDef()
  return colDef.headerName !== EXCLUDED_COLUMN
}
async function initApp() {
  try {
    await loadSitesConfig()
    initGrid()
  }
  catch (error) {
    console.error('应用初始化失败:', error)
    const rootElement = document.querySelector<HTMLElement>(ROOT_ID)
    if (rootElement) {
      rootElement.innerHTML = `
        <div style="padding: 20px; color: #fff;">
          <h2>应用加载失败</h2>
          <p>无法加载站点配置，请检查网络连接或联系管理员。</p>
        </div>
      `
    }
  }
}

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

  const rootElement = document.querySelector<HTMLElement>(ROOT_ID)
  if (!rootElement) {
    console.error('未找到根元素', ROOT_ID)
    return
  }

  createGrid(rootElement, gridOptions)
  setupExportButtons()
}

function setupExportButtons(): void {
  const exportCsvBtn = document.querySelector<HTMLElement>('#exportCsv')
  const exportExcelBtn = document.querySelector<HTMLElement>('#exportExcel')

  exportCsvBtn?.addEventListener('click', () => {
    gridApi?.exportDataAsCsv({ fileName: getExportFileName('csv'), shouldExportColumn })
  })

  exportExcelBtn?.addEventListener('click', () => {
    gridApi?.exportDataAsExcel({ fileName: getExportFileName('xlsx'), shouldExportColumn })
  })
}

initApp()
