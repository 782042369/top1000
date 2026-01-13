import type { GridReadyEvent } from 'ag-grid-community'

import type { DataType, ResDataType } from '@/types'

export * from './config'

// 文件大小单位转换系数（KB为基准）
const SIZE_UNITS = {
  KB: 1,
  MB: 1024,
  GB: 1024 * 1024,
  TB: 1024 * 1024 * 1024,
} as const

// 文件大小正则
const SIZE_PATTERN = /([\d.]+)\s*(KB|MB|GB|TB)/i

/**
 * 文件大小转换为 KB
 * @param sizeStr 文件大小字符串，如 "1.5GB"
 * @returns KB 数值
 */
export function convertSizeToKb(sizeStr: string): number {
  const match = sizeStr.match(SIZE_PATTERN)
  if (!match) {
    return 0
  }
  const value = Number.parseFloat(match[1])
  const unit = match[2].toUpperCase() as keyof typeof SIZE_UNITS
  return value * SIZE_UNITS[unit]
}

/**
 * 获取 Top1000 数据
 */
export async function fetchData(event: GridReadyEvent<DataType>): Promise<void> {
  try {
    // 使用相对路径，因为前端由 Go 后端提供
    const response = await fetch('/top1000.json')
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`)
    }
    const json: ResDataType = await response.json()
    event.api?.setGridOption('rowData', json.items)
  }
  catch (error) {
    console.error('加载数据失败:', error)
    event.api?.setGridOption('rowData', [])
  }
}
