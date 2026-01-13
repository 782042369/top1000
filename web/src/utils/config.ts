import siteData from './iyuuSites'

export const ptUrlConfig = siteData.reduce(
  (acc, cur) => {
    const baseUrl = cur.site === 'm-team' ? 'kp.m-team.cc' : cur.base_url
    const protocol = cur.is_https >= 1 ? 'https' : 'http'
    const url = `${protocol}://${baseUrl}`

    acc[cur.site] = {
      details: (id: string) => `${url}/${cur.details_page.replace(`{}`, id)}`,
      download: (id: string) => {
        const downloadPage = cur.download_page
        if (downloadPage.includes('download.php')) {
          return `${url}/${downloadPage
            .replace(`{}`, id)
            .replace(`&passkey={passkey}`, '')
            .replace(`&downhash={downHash}`, '')}`
        }
        return undefined
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
