import type { GridReadyEvent } from 'ag-grid-community'

import type { DataType, ResDataType } from '@/types'

export * from './config'
/* 文件大小转换为 KB */
export function convertSizeToKb(sizeStr: string) {
  const sizeUnits = {
    KB: 1,
    MB: 1024,
    GB: 1024 * 1024,
    TB: 1024 * 1024 * 1024,
  }
  const match = sizeStr.match(/([\d.]+)\s*(KB|MB|GB|TB)/i)
  return match
    ? Number.parseFloat(match[1])
    * sizeUnits[match[2].toUpperCase() as keyof typeof sizeUnits]
    : 0 // 优化：使用 keyof 来确保安全
}

export async function fetchData(event: GridReadyEvent<DataType, any>) {
  try {
    const response = await fetch('https://top1000.939593.xyz/top1000.json')
    const json: ResDataType = await response.json()
    event.api?.setGridOption('rowData', json.items)
  }
  catch (error) {
    console.error('Error:', error)
  }
}
