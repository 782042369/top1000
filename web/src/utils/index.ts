import type { GridReadyEvent } from 'ag-grid-community'

import type { DataType, ResDataType } from '@/types'

export * from './config'

const SIZE_UNITS = {
  KB: 1,
  MB: 1024,
  GB: 1024 * 1024,
  TB: 1024 * 1024 * 1024,
} as const

const SIZE_PATTERN = /([\d.]+)\s*(KB|MB|GB|TB)/i

export function convertSizeToKb(sizeStr: string): number {
  const match = sizeStr.match(SIZE_PATTERN)
  if (!match) {
    return 0
  }
  const value = Number.parseFloat(match[1])
  const unit = match[2].toUpperCase() as keyof typeof SIZE_UNITS
  return value * SIZE_UNITS[unit]
}

export async function fetchData(event: GridReadyEvent<DataType>): Promise<void> {
  try {
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
