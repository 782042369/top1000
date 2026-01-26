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

import { fetchData } from './utils'
import { columnDefs, defaultColDef, interactionConfig, performanceConfig } from './gridConfig'
import { loadSitesConfig } from './utils/config'

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
function initGrid() {
  // 注册 AG Grid 模块
  ModuleRegistry.registerModules([
    ClientSideRowModelModule,
    TextFilterModule,
    LocaleModule,
  ])

  // 表格配置（小项目优化版）
  const gridOptions: GridOptions<DataType> = {
    theme: themeAlpine,
    localeText: AG_GRID_LOCALE_CN,

    // 默认列配置
    defaultColDef,

    // 初始化时加载数据
    onGridReady: fetchData,

    // 行ID用数据里的id字段
    getRowId: params => `${params.data.id}`,

    // 列定义（从配置文件导入）
    columnDefs,

    // 性能优化（小项目简化配置）
    ...performanceConfig,

    // 交互优化
    ...interactionConfig,
  }

  // 初始化表格
  const rootElement = document.querySelector<HTMLElement>('#root')
  if (rootElement) {
    createGrid(rootElement, gridOptions)
  }
  else {
    console.error('❌ 未找到根元素 #root')
  }
}

// 启动应用
initApp()

