/*
 * @Author: yanghongxuan
 * @Date: 2025-02-10 15:29:50
 * @Description:
 * @LastEditTime: 2025-02-11 14:56:28
 * @LastEditors: yanghongxuan
 */
import siteData from './iyuuSites'

export const ptUrlConfig = siteData.reduce(
  (acc, cur) => {
    const base_url = cur.site === 'm-team' ? 'kp.m-team.cc' : cur.base_url
    const url = cur.is_https ? `https://${base_url}` : `http://${base_url}`
    acc[cur.site] = {
      details: (id: string) => `${url}/${cur.details_page.replace(`{}`, id)}`,
      download: (id: string) => {
        const download_page = cur.download_page
        if (download_page.includes('download.php')) {
          return `${url}/${download_page
            .replace(`{}`, id)
            .replace(`&passkey={passkey}`, '')
            .replace(`&downhash={downHash}`, '')}
            `
        }
      },
    }

    return acc
  },
  {} as Record<
    string,
    {
      details: (id: string) => string
      download: (id: string) => string | undefined
    }
  >,
)
