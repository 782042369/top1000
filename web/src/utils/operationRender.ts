import type { ICellRendererParams } from 'ag-grid-community'

import type { DataType } from '../types'

import { ptUrlConfig } from './config'

/**
 * 操作列渲染器
 * 生成查看详情和下载种子链接
 */
export function operationRender(params: ICellRendererParams): string | null {
  const data = params.data as DataType
  const { siteName, siteid } = data

  const urlConfig = ptUrlConfig[siteName]
  if (!urlConfig) {
    return null
  }

  const detailsUrl = urlConfig.details(siteid)
  const downloadUrl = urlConfig.download(siteid)

  return renderLinks(detailsUrl, downloadUrl)
}

/**
 * 渲染链接 HTML
 */
function renderLinks(detailsUrl: string, downloadUrl?: string): string {
  const downloadLink = downloadUrl
    ? `<a href="${downloadUrl}" target="_blank" rel="noreferrer" style="margin-left:10px">下载种子</a>`
    : ''

  return `<div>
    <a href="${detailsUrl}" target="_blank" rel="noreferrer">查看详情</a>
    ${downloadLink}
  </div>`
}
