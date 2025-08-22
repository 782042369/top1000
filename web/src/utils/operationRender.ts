import type { ICellRendererParams } from 'ag-grid-community'

import type { DataType } from '../types'

import { ptUrlConfig } from './config'

export function operationRender(params: ICellRendererParams) {
  const { siteName } = params.data as DataType
  let { siteid } = params.data as DataType
  const getUrl = ptUrlConfig[siteName === 'ptlsp' ? 'audiences' : siteName]
  if (!getUrl) {
    return null
  }
  if (siteName === 'ptlsp') {
    siteid = {
      649: '297203',
      8667: '353903',
      8765: '288867',
      default: '297203',
    }[siteid] as string
  }
  const downloadUrl = getUrl.download(siteid)
  const link = `<div>
        <a
          href="${getUrl.details(siteid)}"
          target="_blank"
          rel="noreferrer"
        >
          ${downloadUrl ? `查看详情` : `查看详情(下载到详情页面)`}
        </a>
        ${downloadUrl
      ? (
        `<a
                style="margin-left:10px"
                href="${getUrl.download(siteid)}"
                target="_blank"
                rel="noreferrer"
              >
                下载种子
              </a>`
      )
      : null}
      </div>`
  return link
}
