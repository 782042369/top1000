import type { ICellRendererParams } from 'ag-grid-community'

import type { DataType } from '../types'

import { ptUrlConfig } from './config'

// ptlsp 站点特殊 ID 映射
const PTLSP_ID_MAP: Record<string, string> = {
  649: '297203',
  8667: '353903',
  8765: '288867',
}

const PTLSP_DEFAULT_ID = '297203'

/**
 * 操作列渲染器
 * 生成查看详情和下载种子链接
 */
export function operationRender(params: ICellRendererParams): string | null {
  const data = params.data as DataType
  const { siteName, siteid } = data

  // 处理 ptlsp 站点的特殊 ID 映射
  const actualSiteId = siteName === 'ptlsp'
    ? (PTLSP_ID_MAP[siteid] || PTLSP_DEFAULT_ID)
    : siteid

  const urlConfig = ptUrlConfig[siteName === 'ptlsp' ? 'audiences' : siteName]
  if (!urlConfig) {
    return null
  }

  const detailsUrl = urlConfig.details(actualSiteId)
  const downloadUrl = urlConfig.download(actualSiteId)

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
