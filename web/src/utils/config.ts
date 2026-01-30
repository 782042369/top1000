interface SiteConfig {
  id: number
  site: string
  nickname: string
  base_url: string
  download_page: string
  details_page: string
  is_https: number
  cookie_required: number
}

interface UrlConfig {
  details: (id: string) => string
  download: (id: string) => string | undefined
}

export const ptUrlConfig: Record<string, UrlConfig> = {}

export async function loadSitesConfig(): Promise<void> {
  try {
    const response = await fetch('/sites.json')
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`)
    }

    const json = await response.json()
    const siteData: SiteConfig[] = json.data?.sites || []

    siteData.forEach((site) => {
      const baseUrl = site.site === 'm-team' ? 'kp.m-team.cc' : site.base_url
      const protocol = site.is_https >= 1 ? 'https' : 'http'
      const url = `${protocol}://${baseUrl}`

      ptUrlConfig[site.site] = {
        details: (id: string) => `${url}/${site.details_page.replace('{}', id)}`,
        download: (id: string) => {
          const downloadPage = site.download_page
          if (downloadPage.includes('download.php')) {
            return `${url}/${downloadPage
              .replace('{}', id)
              .replace('&passkey={passkey}', '')
              .replace('&downhash={downHash}', '')}`
          }
          return undefined
        },
      }
    })
  }
  catch (error) {
    console.error('加载站点配置失败:', error)
    throw error
  }
}
