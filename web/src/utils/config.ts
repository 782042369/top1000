// 站点配置类型定义
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

// 缓存站点配置（默认空，启动时加载）
export const ptUrlConfig: Record<string, UrlConfig> = {}

/**
 * 从后端API加载站点配置
 * 在应用启动时调用一次，后续直接使用缓存
 */
export async function loadSitesConfig(): Promise<void> {
  try {
    const response = await fetch('/sites.json')
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const json = await response.json()

    // IYUU API 返回格式：{ ret: 200, data: [...], msg: "..." }
    const siteData: SiteConfig[] = json.data?.sites || []

    // 生成 ptUrlConfig
    siteData.reduce(
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
      ptUrlConfig,
    )
  }
  catch (error) {
    console.error('❌ 加载站点配置失败:', error)
    throw error
  }
}
